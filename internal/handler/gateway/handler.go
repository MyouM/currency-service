package handler

import (
	"context"
	"currency-service/internal/config"
	"currency-service/internal/db/redis"
	"currency-service/internal/logger"
	middleware "currency-service/internal/middleware/auth"
	"currency-service/internal/proto/currpb"
	"currency-service/internal/repository"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

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
	router.HandleFunc("POST /registration", registration)
}

func registration(w http.ResponseWriter, req *http.Request) {
	var user repository.User
	logger := logger.GetLogger()

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&user); err != nil {
		http.Error(
			w,
			fmt.Sprintf("Decode error: %v", err),
			http.StatusBadRequest)
		return
	}

	if user.Exist() {
		http.Error(
			w,
			fmt.Sprintf("User %v already exist", user.Login),
			http.StatusBadRequest)
		return
	}
	user.Add()

	claims := jwt.MapClaims{
		"login":    user.Login,
		"password": user.Password,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(repository.GetSecret())
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("JWT error: %v", err),
			http.StatusBadRequest)
		return
	}
	if err = redis.SetToken(tokenStr); err != nil {
		logger.Error("Redis error", zap.Error(err))
		http.Error(
			w,
			fmt.Sprint("Somethings wrong on server"),
			http.StatusInternalServerError)
		return
	}
	//repository.AddToken(tokenStr)
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
