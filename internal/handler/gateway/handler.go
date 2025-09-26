package handler

import (
	"currency-service/internal/config"
	"currency-service/internal/kafka"
	"currency-service/internal/logger"
	middleware "currency-service/internal/middleware/auth"
	"currency-service/internal/proto/currpb"
	redisCur "currency-service/internal/repository/redis"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type HandlerRelations struct {
	Cfg   *config.AppConfig
	Log   *zap.Logger
	Redis redisCur.RedisFuncs
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
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
		middleware.Validate(
			rltns.Redis,
			rltns.GetOneCurrencyRate(grpcClient)))
	router.HandleFunc(
		"GET /currency/period/{dates}",
		middleware.Validate(
			rltns.Redis,
			rltns.GetIntervalCurrencyChanges(grpcClient)))
	router.HandleFunc(
		"POST /registration",
		middleware.KafkaInit(conf, rltns.RegistrationHandler))
	router.HandleFunc(
		"POST /login",
		middleware.KafkaInit(conf, rltns.LoginHandler))
	router.HandleFunc("GET /livez", rltns.KuberLivez())
	router.HandleFunc("GET /readyz", rltns.KuberReadyz())
}

func (hr *HandlerRelations) LoginHandler(
	producer kafka.ProducerFuncs,
	consumer kafka.ConsumerFuncs) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var (
			user     kafka.AuthRequest
			authResp kafka.AuthResponse
		)
		defer consumer.Close()
		defer producer.Close()

		dec := json.NewDecoder(req.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&user); err != nil {
			hr.Log.Error("Decode error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		value := fmt.Sprintf(
			`{"type":"%s","login":"%s","password":"%s"}`,
			"login",
			user.Login,
			user.Password)
		err := producer.Send(req.Context(), []byte(user.Login), []byte(value))
		if err != nil {
			hr.Log.Error("Kafka error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		msg, err := consumer.Listen(req.Context())
		if err != nil {
			hr.Log.Error("Kafka error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(msg.Value, &authResp); err != nil {
			hr.Log.Error("Unmarshal error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		if authResp.Error != "" {
			hr.Log.Warn("Auth error:", zap.String("", authResp.Error))
			http.Error(
				w,
				fmt.Sprintf("Error: %s", authResp.Error),
				http.StatusBadRequest)
			return
		}

		if err = hr.Redis.SetToken(authResp.Token); err != nil {
			hr.Log.Error("Redis error", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusInternalServerError)
			return
		}
		hr.Log.Info("User logged in", zap.String("Login", user.Login))
		fmt.Fprintf(w, "Token: %s", authResp.Token)
	}
}

func (hr *HandlerRelations) RegistrationHandler(
	producer kafka.ProducerFuncs,
	consumer kafka.ConsumerFuncs) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var (
			user     kafka.AuthRequest
			authResp kafka.AuthResponse
		)
		defer consumer.Close()
		defer producer.Close()

		dec := json.NewDecoder(req.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&user); err != nil {
			hr.Log.Error("Decode error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}
		hr.Log.Info("Registration operation started")

		value := fmt.Sprintf(
			`{"type":"%s","login":"%s","password":"%s"}`,
			"register",
			user.Login,
			user.Password)
		err := producer.Send(req.Context(), []byte(user.Login), []byte(value))
		if err != nil {
			hr.Log.Error("Kafka error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		msg, err := consumer.Listen(req.Context())
		if err != nil {
			hr.Log.Error("Kafka error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(msg.Value, &authResp); err != nil {
			hr.Log.Error("Unmarshal error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		if authResp.Error != "" {
			hr.Log.Warn("Auth error:", zap.String("", authResp.Error))
			http.Error(
				w,
				fmt.Sprintf("Error: %s", authResp.Error),
				http.StatusBadRequest)
			return
		}

		if err = hr.Redis.SetToken(authResp.Token); err != nil {
			hr.Log.Error("Redis error", zap.Error(err))
			http.Error(
				w,
				fmt.Sprint("Registration done, but token is not set"),
				http.StatusInternalServerError)
			return
		}
		hr.Log.Info("New registration", zap.String("Login", user.Login))
		fmt.Fprintf(w, "Token: %s", authResp.Token)
	}
}

func (hr *HandlerRelations) GetIntervalCurrencyChanges(grpcClient currpb.CurrencyServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		dts := req.PathValue("dates")

		dates := strings.Split(dts, "-")
		resp, err := grpcClient.GetIntervalCurrency(
			req.Context(),
			&currpb.ClientIntervalRequest{
				DateBegin: dates[0],
				DateEnd:   dates[1],
			})
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("gRPC error: %v", err),
				http.StatusBadGateway)
			return
		}
		currRates := resp.GetRates()
		for _, rate := range currRates {
			fmt.Fprintf(w, "%v, %f\n", rate.Date, rate.Rate)
		}

	}
}

func (hr *HandlerRelations) GetOneCurrencyRate(grpcClient currpb.CurrencyServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		date := req.PathValue("date")

		resp, err := grpcClient.GetSpecificCurrency(
			req.Context(),
			&currpb.ClientSpecRequest{
				Date: date,
			})
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("gRPC error: %v", err),
				http.StatusBadGateway)
			return
		}

		fmt.Fprintf(w, "OK: %f\n", resp.GetCurrency())
	}
}
