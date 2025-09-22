package postgres

import (
	"currency-service/internal/config"
	"currency-service/internal/proto/currpb"
	"database/sql"
	"fmt"
	"time"
)

//go:generate mockgen -source=currency_psql.go -destination=mock/mock_currency_psql.go -package=test_postgres

type CurrencyRepo struct {
	DB *sql.DB
}

type CurrencyPsqlFuncs interface {
	GetCurrencyChanges(time.Time, time.Time) ([]*currpb.CurrencyRates, error)
	GetOneCurrencyRate(time.Time) (float64, error)
	AddWorkerInfo(time.Time, string, float64) error
}

func InitCurrencyRepo(cfg *config.DatabaseConfig) (CurrencyRepo, error) {
	db, _, err := NewDatabaseConnection(cfg)
	if err != nil {
		return CurrencyRepo{}, err
	}
	return CurrencyRepo{DB: db}, nil
}

func (repo CurrencyRepo) GetCurrencyChanges(dateFrom, dateTo time.Time) ([]*currpb.CurrencyRates, error) {
	currRates := make([]*currpb.CurrencyRates, 0, 2)
	rows, err := repo.DB.Query(
		`SELECT DISTINCT date, currency_rate
		FROM exchange_rates
		WHERE date BETWEEN $1 AND $2`,
		dateFrom,
		dateTo)
	if err != nil {
		fmt.Println("Database error: ", err)
		return currRates, err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		currRates = append(currRates, &currpb.CurrencyRates{})
		if err := rows.Scan(&currRates[i].Date, &currRates[i].Rate); err != nil {
			fmt.Println("Database scan error: ", err)
			return currRates, err
		}
		i++
	}
	if err = rows.Err(); err != nil {
		fmt.Println("Database scan error: ", err)
		return currRates, err
	}
	return currRates, nil
}

func (repo CurrencyRepo) GetOneCurrencyRate(date time.Time) (float64, error) {
	var (
		dt       time.Time
		currRate float64
	)
	rows, err := repo.DB.Query(
		`SELECT date, currency_rate
		FROM exchange_rates
		WHERE date = $1`,
		date)
	if err != nil {
		fmt.Println("Database error: ", err)
		return currRate, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&dt, &currRate); err != nil {
			fmt.Println("Database scan error: ", err)
			return currRate, err
		}
	}

	if err = rows.Err(); err != nil {
		fmt.Println("Database scan error: ", err)
		return currRate, err
	}
	return currRate, nil
}

func (repo CurrencyRepo) AddWorkerInfo(date time.Time, target string, rate float64) error {
	_, err := repo.DB.Exec(`INSERT INTO exchange_rates 
			 (date, target_currency, currency_rate) 
	            	 VALUES ($1, $2, $3)`,
		date,
		target,
		rate,
	)
	return err
}
