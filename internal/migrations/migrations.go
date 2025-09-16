package migrations

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewMigrations(db *sql.DB, path string) error {
	file := fmt.Sprintf("file://./internal/migrations/%s", path)
	schema := fmt.Sprintf("schema_migrations_%s", path)

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: schema,
	})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(file, "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
