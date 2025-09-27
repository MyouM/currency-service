package db_test

import (
	"context"
	"currency-service/internal/repository/postgres"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCurrnecyPostgres(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_USER":     "user",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(5 * time.Second),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("host=%s port=%s user=user password=password dbname=testdb sslmode=disable",
		host, port.Port())
	db, err := sql.Open("postgres", dsn)
	assert.NoError(t, err)
	defer db.Close()

	err = db.PingContext(ctx)
	assert.NoError(t, err)

	_, err = db.ExecContext(
		ctx,
		`CREATE TABLE exchange_rates (
	id SERIAL PRIMARY KEY,
	date DATE NOT NULL,
	base_currency VARCHAR(10) NOT NULL DEFAULT 'RUB',
	target_currency VARCHAR(10) NOT NULL,
	currency_rate REAL NOT NULL
	)`)
	assert.NoError(t, err)

	repo := postgres.CurrencyRepo{DB: db}

	timeString1 := "2000-01-01"
	timePars1, err := time.Parse("2006-01-02", timeString1)
	assert.NoError(t, err)
	timeString2 := "2000-01-02"
	timePars2, err := time.Parse("2006-01-02", timeString2)
	assert.NoError(t, err)

	err = repo.AddWorkerInfo(timePars1, "usd", 1.0)
	assert.NoError(t, err)

	err = repo.AddWorkerInfo(timePars2, "usd", 2.0)
	assert.NoError(t, err)

	resp1, err := repo.GetOneCurrencyRate(timePars1)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, resp1)

	resp2, err := repo.GetCurrencyChanges(timePars1, timePars2)

	assert.Equal(t, 1.0, resp2[0].Rate)
	assert.Equal(t, 2.0, resp2[1].Rate)
	assert.NoError(t, err)
}

func TestAuthPostgres(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_USER":     "user",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(5 * time.Second),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("host=%s port=%s user=user password=password dbname=testdb sslmode=disable",
		host, port.Port())
	db, err := sql.Open("postgres", dsn)
	assert.NoError(t, err)
	defer db.Close()

	err = db.PingContext(ctx)
	assert.NoError(t, err)

	_, err = db.ExecContext(
		ctx,
		`CREATE TABLE authentication (
		login VARCHAR(20) PRIMARY KEY,
		password VARCHAR(30) 
		)`)
	assert.NoError(t, err)

	repo := postgres.AuthRepo{DB: db}

	login := "user"
	password := "password"
	err = repo.AddUser(login, password)
	assert.NoError(t, err)

	check, err := repo.IsLoginExist(login)
	assert.Equal(t, true, check)
	assert.NoError(t, err)

	check, err = repo.IsLoginExist("nigol")
	assert.NoError(t, err)
	assert.Equal(t, false, check)

	check, err = repo.LogIn(login, password)
	assert.NoError(t, err)
	assert.Equal(t, true, check)

	check, err = repo.LogIn("nigol", password)
	assert.NoError(t, err)
	assert.Equal(t, false, check)
}
