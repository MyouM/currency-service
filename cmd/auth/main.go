package main

import (
	"context"
	"currency-service/internal/auth"
	"currency-service/internal/config"
	"currency-service/internal/logger"
	"currency-service/internal/migrations"
	"currency-service/internal/repository/postgres"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	//Получение данных из конфига
	cfg, err := config.LoadConfig("./internal/config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	//Миграция БД
	err = migrations.NewMigrations(cfg.Database, "auth")
	if err != nil {
		log.Fatalf("error migration: %v", err)
	}

	//Подключение к Postgres
	_, _, err = postgres.NewDatabaseConnection(cfg.Database)
	if err != nil {
		log.Fatalf("error init database connection: %v", err)
	}

	//Инициализация логгирования
	logger := logger.InitLogger()
	defer logger.Sync()

	//Контекст, реагирующий на сигналы системы
	sigShut, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	//Запуск сервиса авторизации
	logger.Info("Auth service start workinng.")
	auth.StartAuthService(sigShut)
}
