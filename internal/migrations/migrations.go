package migrations

import (
	"currency-service/internal/config"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewMigrations(db *config.DatabaseConfig, path string) error {
	pg := fmt.Sprint(
		"postgres://",
		db.User,
		":",
		db.Password,
		"@",
		db.Host,
		":",
		db.Port,
		"/",
		db.Name,
		"?sslmode=disable",
	)
	file := "file://./internal/migrations/" + path
	m, err := migrate.New(file, pg)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil {
		return err
	}
	return nil
}
