package handler

import (
	"context"
	"currency-service/internal/db"
	"currency-service/internal/proto/currpb"
	"database/sql"
	"fmt"
	"time"
)

type Server struct {
	currpb.UnimplementedCurrencyServiceServer
	DB *sql.DB
}

func (s *Server) GetSpecificCurrency(ctx context.Context, req *currpb.ClientSpecRequest) (*currpb.ClientSpecResponse, error) {
	strDate := req.GetDate()
	date, err := time.Parse(time.DateOnly, strDate)
	if err != nil {
		fmt.Println("Wrong timestamp")
		return &currpb.ClientSpecResponse{}, err
	}

	currRate, err := db.GetOneCurrencyRate(s.DB, date)
	if err != nil {
		fmt.Println("Database error")
		return &currpb.ClientSpecResponse{}, err
	}

	return &currpb.ClientSpecResponse{Currency: currRate}, err
}
