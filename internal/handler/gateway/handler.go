package handler

import (
	"context"
	"currency-service/internal/config"
	"currency-service/internal/proto/currpb"
	"fmt"
	"net/http"
	"time"
)

func GatewayHandlersInit(router *http.ServeMux,
	cfg *config.AppConfig,
	grpcClient currpb.CurrencyServiceClient) {
	router.HandleFunc("GET /currency/{date}", GetOneCurrencyRate(grpcClient))
}

func GetOneCurrencyRate(grpcClient currpb.CurrencyServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		date := req.PathValue("date")
		ctx, cancel := context.WithTimeout(req.Context(), 6*time.Second)
		defer cancel()

		resp, err := grpcClient.GetSpecificCurrency(ctx,
			&currpb.ClientSpecRequest{
				Date: date,
			})
		if err != nil {
			http.Error(w,
				fmt.Sprintf("gRPC error: %v", err),
				http.StatusBadGateway)
			return
		}

		fmt.Fprintf(w, "OK: %v\n", resp.GetCurrency())
	}
}
