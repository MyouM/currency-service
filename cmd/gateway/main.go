package main

import (
	"context"
	"currency-service/internal/config"
	handler "currency-service/internal/handler/gateway"
	"currency-service/internal/logger"
	"currency-service/internal/proto/currpb"
	"currency-service/internal/repository/redis"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	//Получение данных из конфига
	cfg, err := config.LoadConfig("./internal/config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	//Подключение к gRPC серверу и инициализация клиента
	conn, err := grpc.Dial("localhost:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("error grpc dial: %v", err)
	}
	defer conn.Close()
	client := currpb.NewCurrencyServiceClient(conn)

	//Инициализация логгирования
	logger := logger.InitLogger()
	defer logger.Sync()

	//Подключение к Redis
	rds, err := redis.InitRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("error init redis: %v", err)
	}
	defer rds.Close()

	//Инициализация и запуск http сервера 
	router := http.NewServeMux()
	handler.GatewayHandlersInit(router, &cfg, client)
	server := http.Server{
		Addr:    ":8082",
		Handler: router,
	}
	go func() {
		logger.Info("Gateway server running on port 8082...")
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("Gateway server stopped: %v\n", err)
		}
	}()

	//GracefulShutdown
	sigShut, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-sigShut.Done()

	fmt.Println("Shutting down server...")
	if err := server.Shutdown(context.Background()); err != nil {
		fmt.Printf("Shuting down error: %v\n", err)
	}
	fmt.Println("Server stop working.")

}
