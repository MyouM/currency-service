package db

import (
	"currency-service/internal/config"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func NewDatabaseConnection(cfg config.DatabaseConfig) (*sql.DB, string, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, "", fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, dsn, nil
}

func GetOneCurrencyRate(DB *sql.DB, date time.Time) (float64, error) {
	var (
		dt       time.Time
		currRate float64
	)
	rows, err := DB.Query(
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

func AddWorkerInfo(DB *sql.DB, date time.Time, target string, rate float64) error {
	_, err := DB.Exec(`INSERT INTO exchange_rates 
			 (date, target_currency, currency_rate) 
	            	 VALUES ($1, $2, $3)`,
		date,
		target,
		rate,
	)
	return err
}
