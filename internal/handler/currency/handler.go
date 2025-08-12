package handler

import (
	"context"
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

	rows, err := s.DB.Query(
		`SELECT date, currency_rate
		FROM exchange_rates
		WHERE date = $1`,
		date)
	if err != nil {
		fmt.Println("Database error: ", err)
		return &currpb.ClientSpecResponse{}, err
	}
	defer rows.Close()

	var (
		dt       time.Time
		currRate float64
	)
	if err := rows.Scan(&dt, &currRate); err != nil {
		fmt.Println("Database scan error: ", err)
		return &currpb.ClientSpecResponse{}, err
	}

	return &currpb.ClientSpecResponse{Currency: currRate}, err
}
