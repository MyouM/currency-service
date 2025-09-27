package test

import (
	middleware "currency-service/internal/middleware/auth"
	mock_redis "currency-service/internal/repository/redis/mock"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func handlerTest(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := mock_redis.NewMockRedisFuncs(ctrl)
		mockRedis.EXPECT().FindToken("1234").Return(nil)

		req.Header.Add("Authorization", "1234")
		hndlr := middleware.Validate(mockRedis, handlerTest)
		hndlr.ServeHTTP(w, req)
		res1 := w.Result()
		defer res1.Body.Close()
		assert.Equal(t, http.StatusTeapot, res1.StatusCode)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := mock_redis.NewMockRedisFuncs(ctrl)
		mockRedis.EXPECT().FindToken("4321").Return(fmt.Errorf("error"))

		req.Header.Add("Authorization", "4321")
		hndlr := middleware.Validate(mockRedis, handlerTest)
		hndlr.ServeHTTP(w, req)
		res2 := w.Result()
		defer res2.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res2.StatusCode)
	})
}
