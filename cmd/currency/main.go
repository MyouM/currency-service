package main

import (
	"currency-service/internal/config"
	"currency-service/internal/db"
	"currency-service/internal/handler"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	//configPath := flag.String("config", "./config", "path to the config file")
	//flag.Parse()

	//cfg, err := config.LoadConfig(*configPath)
	cfg, err := config.LoadConfig("./internal/config/config.yaml")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	db, _, err := db.NewDatabaseConnection(cfg.Database)
	if err != nil {
		log.Fatalf("error init database connection: %v", err)
	}

	router := http.NewServeMux()

	/*repo, err := repository.NewCurrency(db)
	if err != nil {
		log.Fatalf("error creating repository: %v", err)
	}*/

	/*logger, err := zap.NewProduction(db)
	if err != nil {
		log.Fatalf("error init logger: %v", err)
	}*/

	timeChan := make(chan struct{})
	defer close(timeChan)
	go func() {
		timeChan <- struct{}{}
		for {
			time.Sleep(time.Hour * 24)
			timeChan <- struct{}{}
		}
	}()

	go func() {
		for {
			select {
			case <-timeChan:
				rateResponse, err := handler.GetBodyUrl(&cfg)
				if err != nil {
					fmt.Println(err)
					break
				}
				ratesJson, _ := json.Marshal(rateResponse.Rub)
				_, err = db.Exec(`INSERT INTO currency (date, currency_rates)
						VALUES ($1, $2)`, rateResponse.Date,
					ratesJson)
				if err != nil {
					fmt.Println(err)
					break
				}
			}
		}
	}()

	handler.CurrencyHandlerInit(router, &cfg, db)

	server := http.Server{
		Addr:    ":8081",
		Handler: router,
	}
	server.ListenAndServe()

}
