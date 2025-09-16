package main

import (
	"context"
	"currency-service/internal/config"
	handler "currency-service/internal/handler/currency"
	"currency-service/internal/logger"
	"currency-service/internal/metrics"
	"currency-service/internal/migrations"
	"currency-service/internal/proto/currpb"
	"currency-service/internal/repository/postgres"
	"currency-service/internal/worker"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	//Получение данных из конфига
	cfg, err := config.LoadConfig("./internal/config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	//Подключение к Postgres
	repo, err := postgres.InitCurrencyRepo(cfg.Database)
	if err != nil {
		log.Fatalf("error init database connection: %v", err)
	}

	//Миграция БД
	err = migrations.NewMigrations(repo.DB, "currency")
	if err != nil {
		log.Fatalf("error migration: %v", err)
	}

	//Инициализация логгирования
	logger := logger.InitLogger()
	defer logger.Sync()

	//Инициализация Prometheus
	promets := metrics.InitPrometh()
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Info("Prometheus server runing on port 8080...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Error with Prometheus: %s", err)
		}
	}()

	//Занимаем порт 8081
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	//Инициализация gRPC сервера
	grpcServer := grpc.NewServer()
	currpb.RegisterCurrencyServiceServer(
		grpcServer,
		&handler.Server{
			Repo:    repo,
			Prometh: promets})

	//Контекст, реагирующий на сигналы системы
	sigShut, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Запуск воркера
	go worker.CurrencyWorker(&cfg, repo, sigShut)

	//Запуск gRPC сервера
	go func() {
		logger.Info("Currency server running on port 8081...")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Error on server: %v", err)
		}
	}()

	//GracefulShutdown
	<-sigShut.Done()
	grpcServer.GracefulStop()
	logger.Info("gRPC server stoped")
}
