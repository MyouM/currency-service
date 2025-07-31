package handler

import (
	"currency-service/internal/config"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RatesResponse struct {
	Date time.Time          `json:"date"`
	Rub  map[string]float64 `json:"rub"`
}

func CurrencyHandlerInit(router *http.ServeMux, cfg *config.AppConfig, db *sql.DB) {
	//router.HandleFunc("POST /currency", AddCurrency())
	router.HandleFunc("GET /currency/{curr}", GetCurrency(db))
	//router.HandleFunc("DELETE /currency", DeleteCurrency(ctx, cfg))
}

func GetCurrency(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var (
			rateResponse RatesResponse
			jsonRub      []byte
		)

		requestCurr := req.PathValue("curr")
		query := `SELECT date, currency_rates
			  FROM currency
			  ORDER BY date DESC
			  LIMIT 1`
		row, err := db.Query(query)
		if err != nil {
			fmt.Println(err)
			return
		}
		row.Scan(&rateResponse.Date, &jsonRub)

		if err := json.Unmarshal(jsonRub, &rateResponse.Rub); err != nil {
			fmt.Println(err)
			return
		}

		exRate, ok := rateResponse.Rub[requestCurr]
		if !ok {
			fmt.Printf("%s don't have exchange rate to rub", requestCurr)
			return
		}

		fmt.Println(exRate)
	}
}

func GetBodyUrl(cfg *config.AppConfig) (RatesResponse, error) {
	var rateResponse RatesResponse
	reqCurr, err := http.NewRequest(
		http.MethodGet,
		cfg.URL.URL,
		nil)
	if err != nil {
		//fmt.Println("failed to create request: ", err)
		return rateResponse, err
	}

	client := &http.Client{
		Timeout: time.Second * 15,
	}

	resp, err := client.Do(reqCurr)

	if err != nil {
		//fmt.Println("failed to make API request: ", err)
		return rateResponse, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rateResponse, fmt.Errorf("failed to connect API")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		//fmt.Println("failed to read response body: ", err)
		return rateResponse, err
	}

	if err := json.Unmarshal(body, &rateResponse); err != nil {
		//fmt.Println("failed to unmarshal response body: ", err)
		return rateResponse, err
	}

	return rateResponse, nil
}
