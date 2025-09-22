package handler

import (
	"context"
	"currency-service/internal/metrics"
	"currency-service/internal/proto/currpb"
	"currency-service/internal/repository/postgres"
	"time"
)

type Server struct {
	currpb.UnimplementedCurrencyServiceServer
	Prometh *metrics.Prometh
	Psql    postgres.CurrencyPsqlFuncs
}

func (s *Server) GetSpecificCurrency(
	ctx context.Context,
	req *currpb.ClientSpecRequest) (*currpb.ClientSpecResponse, error) {

	//s.Prometh.GrpcRequestTotal.WithLabelValues("GetRate").Inc()
	strDate := req.GetDate()
	date, err := time.Parse(time.DateOnly, strDate)
	if err != nil {
		return &currpb.ClientSpecResponse{}, err
	}

	currRate, err := s.Psql.GetOneCurrencyRate(date)
	if err != nil {
		return &currpb.ClientSpecResponse{}, err
	}

	return &currpb.ClientSpecResponse{Currency: currRate}, nil
}

func (s *Server) GetIntervalCurrency(
	ctx context.Context,
	req *currpb.ClientIntervalRequest) (*currpb.ClientIntervalResponse, error) {

	//s.Prometh.GrpcRequestTotal.WithLabelValues("GetRate").Inc()
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

	currRates, err := s.Psql.GetCurrencyChanges(dateFrom, dateTo)
	if err != nil {
		return &currpb.ClientIntervalResponse{}, err
	}
	return &currpb.ClientIntervalResponse{Rates: currRates}, nil
}
