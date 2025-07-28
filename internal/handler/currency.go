package currency

import (
	"currency-service/internal/config"
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

func CurrencyHandlerInit(router *http.ServeMux, cfg *config.AppConfig) {
	//router.HandleFunc("POST /currency", AddCurrency())
	router.HandleFunc("GET /currency/{curr}", GetCurrency(cfg))
	//router.HandleFunc("DELETE /currency", DeleteCurrency(ctx, cfg))
}

func GetCurrency(cfg *config.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		reqCurr, err := http.NewRequest(
			http.MethodGet,
			cfg.URL.URL,
			nil)
		if err != nil {
			fmt.Println("failed to create request: ", err)
			return
		}

		client := &http.Client{
			Timeout: time.Second * 15,
		}

		resp, err := client.Do(reqCurr)

		if err != nil {
			fmt.Println("failed to make API request: ", err)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("failed to connect API: ", err)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("failed to read response body: ", err)
			return
		}

		var rateResponse RatesResponse
		if err := json.Unmarshal(body, &rateResponse); err != nil {
			fmt.Println("failed to unmarshal response body: ", err)
			return
		}

		requestCurr := req.PathValue("curr")

		exRate, ok := rateResponse.Rub[requestCurr]
		if !ok {
			fmt.Printf("%s don't have exchange rate to rub", requestCurr)
			return
		}

		fmt.Println(exRate)
	}
}
