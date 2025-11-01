package handler

import (
	"currency-service/internal/proto/currpb"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type CurrencyResponse struct {
	date string  `json:"date"`
	rate float64 `json:"rate"`
}

type IntervalResponse struct {
	CurRates []CurrencyResponse `json:"rates"`
}

func (hr *HandlerRelations) GetIntervalCurrencyChanges(grpcClient currpb.CurrencyServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		dts := req.PathValue("dates")

		dates := strings.Split(dts, "-")
		resp, err := grpcClient.GetIntervalCurrency(
			req.Context(),
			&currpb.ClientIntervalRequest{
				DateBegin: dates[0],
				DateEnd:   dates[1],
			})
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("gRPC error: %v", err),
				http.StatusBadGateway)
			return
		}
		currRates := resp.GetRates()

		w.Header().Set("Content-Type", "application/json")
		rates := IntervalResponse{
			CurRates: make([]CurrencyResponse, 0, 5),
		}
		for _, rate := range currRates {
			rates.CurRates = append(
				rates.CurRates,
				CurrencyResponse{
					date: rate.Date,
					rate: rate.Rate,
				})
		}
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(&rates); err != nil {
			http.Error(
				w,
				fmt.Sprintf("Something gona wrong"),
				http.StatusBadGateway)
			return
		}
	}
}

func (hr *HandlerRelations) GetOneCurrencyRate(grpcClient currpb.CurrencyServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		date := req.PathValue("date")

		resp, err := grpcClient.GetSpecificCurrency(
			req.Context(),
			&currpb.ClientSpecRequest{
				Date: date,
			})
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("gRPC error: %v", err),
				http.StatusBadGateway)
			return
		}

		cur := CurrencyResponse{
			date: date,
			rate: resp.GetCurrency(),
		}

		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(&cur); err != nil {
			http.Error(
				w,
				fmt.Sprintf("Something gona wrong"),
				http.StatusBadGateway)
			return
		}
	}
}
