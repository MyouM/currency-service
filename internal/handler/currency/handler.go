package handler

import (
	"context"
	"currency-service/internal/db"
	"currency-service/internal/proto/currpb"
	"database/sql"
	"time"
)

type Server struct {
	currpb.UnimplementedCurrencyServiceServer
	DB *sql.DB
}

func (s *Server) GetSpecificCurrency(
	ctx context.Context,
	req *currpb.ClientSpecRequest) (*currpb.ClientSpecResponse, error) {

	strDate := req.GetDate()
	date, err := time.Parse(time.DateOnly, strDate)
	if err != nil {
		return &currpb.ClientSpecResponse{}, err
	}

	currRate, err := db.GetOneCurrencyRate(s.DB, date)
	if err != nil {
		return &currpb.ClientSpecResponse{}, err
	}

	return &currpb.ClientSpecResponse{Currency: currRate}, nil
}

func (s *Server) GetIntervalCurrency(
	ctx context.Context,
	req *currpb.ClientIntervalRequest) (*currpb.ClientIntervalResponse, error) {

	strBeginDate := req.GetDateBegin()
	strEndDate := req.GetDateEnd()

	dateFrom, err := time.Parse(time.DateOnly, strBeginDate)
	if err != nil {
		return &currpb.ClientIntervalResponse{}, err
	}

	dateTo, err := time.Parse(time.DateOnly, strEndDate)
	if err != nil {
		return &currpb.ClientIntervalResponse{}, err
	}

	currRates, err := db.GetCurrencyChanges(s.DB, dateFrom, dateTo)
	if err != nil {
		return &currpb.ClientIntervalResponse{}, err
	}
	return &currpb.ClientIntervalResponse{Rates: currRates}, nil
}
