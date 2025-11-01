package handler

import (
	"currency-service/internal/config"
	"currency-service/internal/logger"
	midAuth "currency-service/internal/middleware/auth"
	midMetric "currency-service/internal/middleware/metrics"
	"currency-service/internal/proto/currpb"
	redisCur "currency-service/internal/repository/redis"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type HandlerRelations struct {
	Cfg   *config.AppConfig
	Log   *zap.Logger
	Redis redisCur.RedisFuncs
}

func GatewayHandlersInit(
	router *http.ServeMux,
	conf *config.AppConfig,
	grpcClient currpb.CurrencyServiceClient) {

	rltns := &HandlerRelations{
		Cfg:   conf,
		Log:   logger.GetLogger(),
		Redis: redisCur.GetRedisClient(),
	}
	router.HandleFunc(
		"GET /currency/one/{date}",
		midAuth.Validate(
			rltns.Redis,
			midMetric.GrpcMetrics(
				rltns.GetOneCurrencyRate(grpcClient))))
	router.HandleFunc(
		"GET /currency/period/{dates}",
		midAuth.Validate(
			rltns.Redis,
			midMetric.GrpcMetrics(
				rltns.GetIntervalCurrencyChanges(grpcClient))))
	router.HandleFunc(
		"POST /registration",
		midAuth.KafkaInit(conf, rltns.RegistrationHandler))
	router.HandleFunc(
		"POST /login",
		midAuth.KafkaInit(conf, rltns.LoginHandler))
	router.HandleFunc("GET /livez", rltns.KuberLivez())
	router.HandleFunc("GET /readyz", rltns.KuberReadyz())
	router.Handle("/metrics", promhttp.Handler())
}
