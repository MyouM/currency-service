package main

import (
	"currency-service/internal/config"
	currency "currency-service/internal/handler"
	"fmt"
	"log"
	"net/http"
)

func main() {
	//configPath := flag.String("config", "./config", "path to the config file")
	//flag.Parse()

	//cfg, err := config.LoadConfig(*configPath)
	cfg, err := config.LoadConfig("./internal/config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	/*db, err := db.NewDatabaseConnection(cfg.Database)
	if err != nil {
		log.Fatalf("error init database connection: %v", err)
	}*/

	router := http.NewServeMux()

	/*repo, err := repository.NewCurrency(db)
	if err != nil {
		log.Fatalf("error creating repository: %v", err)
	}*/

	/*logger, err := zap.NewProduction(db)
	if err != nil {
		log.Fatalf("error init logger: %v", err)
	}*/

	currency.CurrencyHandlerInit(router, &cfg)

	server := http.Server{
		Addr:    ":8081",
		Handler: router,
	}
	fmt.Println("Server is listening on port 8081")
	server.ListenAndServe()

}
