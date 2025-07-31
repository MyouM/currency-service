package main

import (
	"currency-service/internal/config"
	"currency-service/internal/migrations"
	"log"
)

func main() {
	cfg, err := config.LoadConfig("./internal/config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	err = migrations.MigrateDB(cfg.Database)
	if err != nil {
		log.Fatalf("error migration: %v", err)
	}
}
