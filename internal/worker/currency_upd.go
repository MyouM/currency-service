package worker

import (
	"currency-service/internal/config"
	"currency-service/internal/db"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RatesResponse struct {
	Date string             `json:"date"`
	Rub  map[string]float64 `json:"rub"`
}

func CurrencyWorker(cfg *config.AppConfig, DB *sql.DB) {
	timeChan := make(chan struct{})
	defer close(timeChan)
	go func() {
		timeChan <- struct{}{}
		for {
			time.Sleep(time.Hour * 24)
			timeChan <- struct{}{}
		}
	}()
	for {
		select {
		case <-timeChan:
			rateResponse, err := getBodyUrl(cfg)
			if err != nil {
				fmt.Println(err)
				break
			}

			currRate, ok := rateResponse.Rub[cfg.Currency.Target]
			if !ok {
				fmt.Println("Target currency is not in response")
				break
			}

			rateDate, err := time.Parse(time.DateOnly, rateResponse.Date)
			if err != nil {
				fmt.Println("Incorrect date format")
				break
			}

			err = db.AddWorkerInfo(DB,
				rateDate,
				cfg.Currency.Target,
				currRate)
			if err != nil {
				fmt.Println(err)
				break
			}
		}
	}
}

func getBodyUrl(cfg *config.AppConfig) (RatesResponse, error) {
	var rateResponse RatesResponse

	reqCurr, err := http.NewRequest(
		http.MethodGet,
		cfg.Currency.URL,
		nil)
	if err != nil {
		fmt.Println("failed to create request: ", err)
		return rateResponse, err
	}

	client := &http.Client{
		Timeout: time.Second * 6,
	}

	resp, err := client.Do(reqCurr)

	if err != nil {
		fmt.Println("failed to make API request: ", err)
		return rateResponse, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rateResponse, fmt.Errorf("failed to connect API")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read response body: ", err)
		return rateResponse, err
	}

	if err := json.Unmarshal(body, &rateResponse); err != nil {
		fmt.Println("failed to unmarshal response body: ", err)
		return rateResponse, err
	}

	return rateResponse, nil
}
