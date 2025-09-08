package handler

import (
	"context"
	"currency-service/internal/config"
	"currency-service/internal/logger"
	middleware "currency-service/internal/middleware/auth"
	"currency-service/internal/proto/currpb"
	"currency-service/internal/repository/redis"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func GatewayHandlersInit(
	router *http.ServeMux,
	cfg *config.AppConfig,
	grpcClient currpb.CurrencyServiceClient) {

	router.HandleFunc(
		"GET /currency/one/{date}",
		middleware.Validate(getOneCurrencyRate(grpcClient)))
	router.HandleFunc(
		"GET /currency/period/{dates}",
		middleware.Validate(getIntervalCurrencyChanges(grpcClient)))
	router.HandleFunc(
		"POST /registration",
		registrationHandler)
	router.HandleFunc(
		"POST /login",
		loginHandler)
}

func loginHandler(w http.ResponseWriter, req *http.Request) {
	var (
		user           User
		authResp       auth.AuthResponse
		kafkaKey       = "login-gateway"
		logger         = logger.GetLogger()
		gwayRespWriter = kafka.NewWriter(kafka.WriterConfig{
			Broker:   []string{auth.Host},
			Topic:    auth.LoginReqTopic,
			Balancer: &kafka.LeastBytes{},
		})
		gwayReqReader = kafka.NewReader(kafka.WriteConfig{
			Broker:   []string{auth.Host},
			Topic:    auth.LoginRespTopic,
			MinBytes: 1,
			MaxBytes: 10e6,
		})
	)
	defer gwayRespWriter.Close()
	defer gwayReqReader.Close()

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&user); err != nil {
		logger.Error("Decode error:", zap.Error(err))
		http.Error(
			w,
			"Somethings gone wrong",
			http.StatusBadRequest)
		return
	}

	request := fmt.Sprintf(
		`{"id":"%s","login":"%s","password":"%s"}`,
		kafkaKey,
		user.Login,
		user.Password)
	err = gwayRespWriter.WriteMessage(
		context.Background(),
		kafka.Message{
			Key:   []byte(kafkaKey),
			Value: []byte(request),
		})
	if err != nil {
		logger.Error("Kafka error:", zap.Error(err))
		http.Error(
			w,
			"Somethings gone wrong",
			http.StatusBadRequest)
		return
	}

	msg, err := gwayReqReader.ReadMessage(context.Background())
	if err != nil {
		logger.Error("Kafka error:", zap.Error(err))
		http.Error(
			w,
			"Somethings gone wrong",
			http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(msg.Value, &authResp); err != nil {
		logger.Error("Unmarshal error:", zap.Error(err))
		http.Error(
			w,
			"Somethings gone wrong",
			http.StatusBadRequest)
		return
	}

	if msg.Error != nil {
		logger.Error("Auth error:", zap.Error(msg.Error))
		http.Error(
			w,
			fmt.Sprintf("Error: %s", msg.Error),
			http.StatusBadRequest)
		return
	}

	if err = redis.SetToken(msg.Token); err != nil {
		logger.Error("Redis error", zap.Error(err))
		http.Error(
			w,
			fmt.Sprint("Registration done, but token is not set"),
			http.StatusInternalServerError)
		return
	}
	logger.Info("User logged in", zap.String("Login", user.Login))
	fmt.Fprintf(w, "Token: %s", tokenStr)
}

func registrationHandler(w http.ResponseWriter, req *http.Request) {
	var (
		user           User
		authResp       auth.AuthResponse
		kafkaKey       = "register-gateway"
		logger         = logger.GetLogger()
		gwayRespWriter = kafka.NewWriter(kafka.WriterConfig{
			Broker:   []string{auth.Host},
			Topic:    auth.RegisterReqTopic,
			Balancer: &kafka.LeastBytes{},
		})
		gwayReqReader = kafka.NewReader(kafka.WriteConfig{
			Broker:   []string{auth.Host},
			Topic:    auth.RegisterRespTopic,
			MinBytes: 1,
			MaxBytes: 10e6,
		})
	)
	defer gwayRespWriter.Close()
	defer gwayReqReader.Close()

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&user); err != nil {
		logger.Error("Decode error:", zap.Error(err))
		http.Error(
			w,
			"Somethings gone wrong",
			http.StatusBadRequest)
		return
	}

	request := fmt.Sprintf(
		`{"id":"%s","login":"%s","password":"%s"}`,
		kafkaKey,
		user.Login,
		user.Password)
	err = gwayRespWriter.WriteMessage(
		context.Background(),
		kafka.Message{
			Key:   []byte(kafkaKey),
			Value: []byte(request),
		})
	if err != nil {
		logger.Error("Kafka error:", zap.Error(err))
		http.Error(
			w,
			"Somethings gone wrong",
			http.StatusBadRequest)
		return
	}

	msg, err := gwayReqReader.ReadMessage(context.Background())
	if err != nil {
		logger.Error("Kafka error:", zap.Error(err))
		http.Error(
			w,
			"Somethings gone wrong",
			http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(msg.Value, &authResp); err != nil {
		logger.Error("Unmarshal error:", zap.Error(err))
		http.Error(
			w,
			"Somethings gone wrong",
			http.StatusBadRequest)
		return
	}

	if msg.Error != nil {
		logger.Error("Auth error:", zap.Error(msg.Error))
		http.Error(
			w,
			fmt.Sprintf("Error: %s", msg.Error),
			http.StatusBadRequest)
		return
	}

	if err = redis.SetToken(msg.Token); err != nil {
		logger.Error("Redis error", zap.Error(err))
		http.Error(
			w,
			fmt.Sprint("Registration done, but token is not set"),
			http.StatusInternalServerError)
		return
	}
	logger.Info("New registration", zap.String("Login", user.Login))
	fmt.Fprintf(w, "Token: %s", tokenStr)
}

func getIntervalCurrencyChanges(grpcClient currpb.CurrencyServiceClient) http.HandlerFunc {
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

func getOneCurrencyRate(grpcClient currpb.CurrencyServiceClient) http.HandlerFunc {
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
