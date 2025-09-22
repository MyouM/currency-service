package postgres

import (
	"currency-service/internal/config"
	"database/sql"
)

//go:generate mockgen -source=auth_psql.go -destination=mock/mock_auth_psql.go -package=test_postgres

type AuthRepo struct {
	DB *sql.DB
}

type AuthPsqlFuncs interface {
	LogIn(string, string) (bool, error)
	IsLoginExist(string) (bool, error)
	AddUser(string, string) error
}

func InitAuthRepo(cfg *config.DatabaseConfig) (AuthRepo, error) {
	db, _, err := NewDatabaseConnection(cfg)
	if err != nil {
		return AuthRepo{}, err
	}
	return AuthRepo{DB: db}, nil
}

func (repo AuthRepo) LogIn(reqLogin, reqPassword string) (bool, error) {
	var (
		login, password string
	)

	rows, err := repo.DB.Query(
		`SELECT login, password
		FROM authentication
		WHERE login = $1 AND
		password = $2
		LIMIT 1`,
		reqLogin,
		reqPassword)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&login, &password); err != nil {
			return false, err
		}
	}

	if err = rows.Err(); err != nil {
		return false, err
	}
	if login != reqLogin {
		return false, nil
	}
	return true, nil
}

func (repo AuthRepo) IsLoginExist(reqLogin string) (bool, error) {
	var (
		login string
	)

	rows, err := repo.DB.Query(
		`SELECT login
		FROM authentication
		WHERE login = $1
		LIMIT 1`,
		reqLogin)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&login); err != nil {
			return false, err
		}
	}

	if err = rows.Err(); err != nil {
		return false, err
	}
	if login != reqLogin {
		return false, nil
	}
	return true, nil
}

func (repo AuthRepo) AddUser(reqLogin, reqPassword string) error {
	_, err := repo.DB.Exec(
		`INSERT INTO authentication
		 (login, password) 
	    	 VALUES ($1, $2)`,
		reqLogin,
		reqPassword)
	return err
}
