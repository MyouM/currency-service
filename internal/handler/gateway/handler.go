package handler

import (
	"context"
	"currency-service/internal/auth"
	"currency-service/internal/config"
	kafkaCur "currency-service/internal/kafka"
	"currency-service/internal/logger"
	middleware "currency-service/internal/middleware/auth"
	"currency-service/internal/proto/currpb"
	redisCur "currency-service/internal/repository/redis"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type HandlerRelations struct {
	Cfg   *config.AppConfig
	Log   *zap.Logger
	Redis *redisCur.RedisClient
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func GatewayHandlersInit(
	router *http.ServeMux,
	conf *config.AppConfig,
	grpcClient currpb.CurrencyServiceClient) {

	rltns := HandlerRelations{
		Cfg:   conf,
		Log:   logger.GetLogger(),
		Redis: redisCur.GetRedisClient(),
	}
	router.HandleFunc(
		"GET /currency/one/{date}",
		middleware.Validate(
			rltns.Redis,
			rltns.getOneCurrencyRate(grpcClient)))
	router.HandleFunc(
		"GET /currency/period/{dates}",
		middleware.Validate(
			rltns.Redis,
			rltns.getIntervalCurrencyChanges(grpcClient)))
	router.HandleFunc("POST /registration", rltns.registrationHandler)
	router.HandleFunc("POST /login", rltns.loginHandler)
	router.HandleFunc("GET /livez", rltns.kuberLivez())
	router.HandleFunc("GET /readyz", rltns.kuberReadyz())
}

func (hr HandlerRelations) loginHandler(w http.ResponseWriter, req *http.Request) {
	var (
		user           User
		authResp       auth.AuthResponse
		kafkaKey       = "login-gateway"
		gwayRespWriter = kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{hr.Cfg.Kafka.BrokerHost},
			Topic:    kafkaCur.LoginRespTopic,
			Balancer: &kafka.LeastBytes{},
		})
		gwayReqReader = kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{hr.Cfg.Kafka.BrokerHost},
			Topic:    kafkaCur.LoginReqTopic,
			MinBytes: 1,
			MaxBytes: 10e6,
		})
	)
	defer gwayRespWriter.Close()
	defer gwayReqReader.Close()

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
	hr.Log.Info("Log in operation started")

	request := fmt.Sprintf(
		`{"id":"%s","login":"%s","password":"%s"}`,
		kafkaKey,
		user.Login,
		user.Password)
	err := gwayRespWriter.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(kafkaKey),
			Value: []byte(request),
		})
	if err != nil {
		hr.Log.Error("Kafka error:", zap.Error(err))
		http.Error(
			w,
			"Somethings gone wrong",
			http.StatusBadRequest)
		return
	}

	msg, err := gwayReqReader.ReadMessage(context.Background())
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
		hr.Log.Error("Auth error:", zap.String("", authResp.Error))
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
	hr.Log.Info("User logged in", zap.String("Login", user.Login))
	fmt.Fprintf(w, "Token: %s", authResp.Token)
}

func (hr HandlerRelations) registrationHandler(w http.ResponseWriter, req *http.Request) {
	var (
		user           User
		authResp       auth.AuthResponse
		kafkaKey       = "register-gateway"
		gwayRespWriter = kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{hr.Cfg.Kafka.BrokerHost},
			Topic:    kafkaCur.RegisterRespTopic,
			Balancer: &kafka.LeastBytes{},
		})
		gwayReqReader = kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{hr.Cfg.Kafka.BrokerHost},
			Topic:    kafkaCur.RegisterReqTopic,
			MinBytes: 1,
			MaxBytes: 10e6,
		})
	)
	defer gwayRespWriter.Close()
	defer gwayReqReader.Close()

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

	request := fmt.Sprintf(
		`{"id":"%s","login":"%s","password":"%s"}`,
		kafkaKey,
		user.Login,
		user.Password)
	err := gwayRespWriter.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(kafkaKey),
			Value: []byte(request),
		})
	if err != nil {
		hr.Log.Error("Kafka error:", zap.Error(err))
		http.Error(
			w,
			"Somethings gone wrong",
			http.StatusBadRequest)
		return
	}

	msg, err := gwayReqReader.ReadMessage(context.Background())
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
		hr.Log.Error("Auth error:", zap.String("", authResp.Error))
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

func (hr HandlerRelations) getIntervalCurrencyChanges(grpcClient currpb.CurrencyServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		dts := req.PathValue("dates")
		ctx, cancel := context.WithTimeout(req.Context(), 6*time.Second)
		defer cancel()

		dates := strings.Split(dts, "-")
		resp, err := grpcClient.GetIntervalCurrency(
			ctx,
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
			fmt.Fprintf(w, "%v, %v\n", rate.Date, rate.Rate)
		}

	}
}

func (hr HandlerRelations) getOneCurrencyRate(grpcClient currpb.CurrencyServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		date := req.PathValue("date")
		ctx, cancel := context.WithTimeout(req.Context(), 6*time.Second)
		defer cancel()

		resp, err := grpcClient.GetSpecificCurrency(
			ctx,
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

		fmt.Fprintf(w, "OK: %v\n", resp.GetCurrency())
	}
}
