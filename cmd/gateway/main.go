package main

import (
	"currency-service/internal/config"
	handler "currency-service/internal/handler/gateway"
	"currency-service/internal/logger"
	"currency-service/internal/proto/currpb"
	"log"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg, err := config.LoadConfig("./internal/config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	conn, err := grpc.Dial("localhost:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("error grpc dial: %v", err)
	}
	defer conn.Close()

	client := currpb.NewCurrencyServiceClient(conn)

	router := http.NewServeMux()

	logger := logger.InitLogger()
	defer logger.Sync()

	handler.GatewayHandlersInit(router, &cfg, client)

	logger.Info("Server running on port 8082...")
	server := http.Server{
		Addr:    ":8082",
		Handler: router,
	}
	server.ListenAndServe()

}
