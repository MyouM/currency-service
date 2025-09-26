package test

import (
	"currency-service/internal/auth"
	jwtCur "currency-service/internal/jwt"
	"currency-service/internal/kafka"
	mock_postgres "currency-service/internal/repository/postgres/mock"
	"encoding/json"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthChecks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPsql := mock_postgres.NewMockAuthPsqlFuncs(ctrl)
	t.Run("LoginService", func(t *testing.T) {
		var (
			resp = kafka.AuthResponse{}
			req  = kafka.AuthRequest{
				Type:     "login",
				Login:    "login",
				Password: "1234",
			}
		)
		mockPsql.EXPECT().LogIn(req.Login, req.Password).Return(true, nil)
		tool := &auth.AuthTools{
			Repo: mockPsql,
			Req:  req}
		msg, err := tool.LoginCheck()
		assert.NoError(t, err)

		err = json.Unmarshal(msg, &resp)
		assert.NoError(t, err)

		parsed, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
			return jwtCur.GetJWTSecret(), nil
		})
		assert.NoError(t, err)

		claims := parsed.Claims.(jwt.MapClaims)
		assert.Equal(t, req.Login, claims["login"])
	})
	t.Run("RegistrationService", func(t *testing.T) {
		var (
			resp = kafka.AuthResponse{}
			req  = kafka.AuthRequest{
				Type:     "login",
				Login:    "login",
				Password: "1234",
			}
		)
		mockPsql.EXPECT().AddUser(req.Login, req.Password).Return(nil)
		mockPsql.EXPECT().IsLoginExist(req.Login).Return(false, nil)
		tool := &auth.AuthTools{
			Repo: mockPsql,
			Req:  req}
		msg, err := tool.RegisterCheck()
		assert.NoError(t, err)

		err = json.Unmarshal(msg, &resp)
		assert.NoError(t, err)

		parsed, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
			return jwtCur.GetJWTSecret(), nil
		})
		assert.NoError(t, err)

		claims := parsed.Claims.(jwt.MapClaims)
		assert.Equal(t, req.Login, claims["login"])
	})
}
