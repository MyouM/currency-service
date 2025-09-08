package postgres_auth

import (
	"currency-service/internal/repository/postgres"
)

func LogIn(reqLogin, reqPassword string) (bool, error) {
	var (
		DB              = postgres.DB
		login, password string
	)

	rows, err := DB.Query(
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

func IsLoginExist(reqLogin string) (bool, error) {
	var (
		DB    = postgres.DB
		login string
	)

	rows, err := DB.Query(
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
	if login == reqLogin {
		return false, nil
	}
	return true, nil
}

func AddUser(reqLogin, reqPassword string) error {
	DB := postgres.DB
	_, err := DB.Exec(
		`INSERT INTO authentication
		 (login, password) 
	    	 VALUES ($1, $2)`,
		reqLogin,
		reqPassword)
	return err
}
