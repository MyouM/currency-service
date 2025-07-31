package migrations

import (
	"currency-service/internal/config"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Currency_DB struct {
	ID             int       `json:"id" gorm:"primaryKey"`
	Date           time.Time `json:"date" gorm:"type:date;not null`
	Base_currency  string    `json:"base_currency" gorm:"type:varchar(10);default:rub"`
	Currency_rates string    `json:"currency_rates" gorm:"type:jsonb;not null"`
}

func (Currency_DB) TableName() string {
	return "currency"
}

func MigrateDB(cfg config.DatabaseConfig) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	db.AutoMigrate(&Currency_DB{})
	return nil
}
