package main

import (
	"context"
	"currency-service/internal/config"
	handler "currency-service/internal/handler/gateway"
	"currency-service/internal/logger"
	"currency-service/internal/proto/currpb"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

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
	server := http.Server{
		Addr:    ":8082",
		Handler: router,
	}

	go func() {
		logger.Info("Server running on port 8082...")
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("Stopped listening: %v\n", err)
		}
	}()

	sigShut, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-sigShut.Done()

	fmt.Println("Shutting down server...")
	if err := server.Shutdown(context.Background()); err != nil {
		fmt.Printf("Shuting down error: %v\n", err)
	}
	fmt.Println("Server stop working.")

}
