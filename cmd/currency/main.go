package main

import (
	"context"
	"currency-service/internal/config"
	"currency-service/internal/db"
	handler "currency-service/internal/handler/currency"
	"currency-service/internal/logger"
	"currency-service/internal/proto/currpb"
	"currency-service/internal/worker"
	"log"
	"net"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.LoadConfig("./internal/config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	db, _, err := db.NewDatabaseConnection(cfg.Database)
	if err != nil {
		log.Fatalf("error init database connection: %v", err)
	}

	go worker.CurrencyWorker(&cfg, db)

	logger := logger.InitLogger()
	defer logger.Sync()

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	grpcServer := grpc.NewServer()
	currpb.RegisterCurrencyServiceServer(grpcServer, &handler.Server{DB: db})

	go func() {
		logger.Info("Server running on port 8081...")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Error on server: %v", err)
		}
	}()

	sigShut, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-sigShut.Done()
	grpcServer.GracefulStop()
	logger.Info("gRPC server stoped")
}
