package jwt

import (
	"github.com/golang-jwt/jwt"
)

var (
	secret_key = []byte("Some_Secret_Text")
)

func MakeJWT(login, password string) (string, error) {
	claims := jwt.MapClaims{
		"login":    login,
		"password": password,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secret_key)
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func GetJWTSecret() []byte {
	return secret_key
}
