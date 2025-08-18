package worker

import (
	"currency-service/internal/config"
	"currency-service/internal/db"
	"currency-service/internal/logger"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type RatesResponse struct {
	Date string             `json:"date"`
	Rub  map[string]float64 `json:"rub"`
}

func CurrencyWorker(cfg *config.AppConfig, DB *sql.DB) {
	logger := logger.GetLogger()
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
				logger.Error("failed to get url", zap.Error(err))
				break
			}

			currRate, ok := rateResponse.Rub[cfg.Currency.Target]
			if !ok {
				logger.Warn("Target currency is not in response")
				break
			}

			rateDate, err := time.Parse(time.DateOnly, rateResponse.Date)
			if err != nil {
				logger.Warn("Incorrect date format")
				break
			}

			err = db.AddWorkerInfo(DB,
				rateDate,
				cfg.Currency.Target,
				currRate)
			if err != nil {
				logger.Error("Database error", zap.Error(err))
				break
			}
			logger.Info("Update currency")
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
		return rateResponse, err
	}

	client := &http.Client{
		Timeout: time.Second * 6,
	}

	resp, err := client.Do(reqCurr)

	if err != nil {
		return rateResponse, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rateResponse, fmt.Errorf("failed to connect API")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return rateResponse, err
	}

	if err := json.Unmarshal(body, &rateResponse); err != nil {
		return rateResponse, err
	}

	return rateResponse, nil
}
