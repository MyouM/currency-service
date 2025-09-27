package test

import (
	handler "currency-service/internal/handler/gateway"
	mock_kafka "currency-service/internal/kafka/mock"
	"currency-service/internal/logger"
	mock_redis "currency-service/internal/repository/redis/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestKafkaHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prodMock := mock_kafka.NewMockProducerFuncs(ctrl)
	consMock := mock_kafka.NewMockConsumerFuncs(ctrl)
	mockRedis := mock_redis.NewMockRedisFuncs(ctrl)

	t.Run("LoginHandler", func(t *testing.T) {
		body := strings.NewReader(`{"login":"login","password":"1234"}`)
		req := httptest.NewRequest(http.MethodGet, "/test", body)
		w := httptest.NewRecorder()

		valueProd := `{"type":"login","login":"login","password":"1234"}`
		valueCons := `{"type":"login","token":"1234","error":""}`
		prodMock.
			EXPECT().
			Send(req.Context(), []byte("login"), []byte(valueProd)).
			Return(nil)
		prodMock.EXPECT().Close().Return(nil)
		consMock.
			EXPECT().
			Listen(req.Context()).
			Return(
				kafka.Message{Key: []byte("login"), Value: []byte(valueCons)},
				nil)
		consMock.EXPECT().Close().Return(nil)
		mockRedis.EXPECT().SetToken("1234").Return(nil)

		hr := &handler.HandlerRelations{
			Redis: mockRedis,
			Log:   logger.GetLogger()}
		hr.LoginHandler(prodMock, consMock).ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()

		resBody, _ := io.ReadAll(res.Body)
		assert.Equal(t, "Token: 1234", string(resBody))
	})
	t.Run("RegisterHandler", func(t *testing.T) {
		body := strings.NewReader(`{"login":"login","password":"1234"}`)
		req := httptest.NewRequest(http.MethodGet, "/test", body)
		w := httptest.NewRecorder()

		valueProd := `{"type":"register","login":"login","password":"1234"}`
		valueCons := `{"type":"register","token":"1234","error":""}`
		prodMock.
			EXPECT().
			Send(req.Context(), []byte("login"), []byte(valueProd)).
			Return(nil)
		prodMock.EXPECT().Close().Return(nil)
		consMock.
			EXPECT().
			Listen(req.Context()).
			Return(
				kafka.Message{Key: []byte("login"), Value: []byte(valueCons)},
				nil)
		consMock.EXPECT().Close().Return(nil)
		mockRedis.EXPECT().SetToken("1234").Return(nil)

		hr := &handler.HandlerRelations{
			Redis: mockRedis,
			Log:   logger.GetLogger()}
		hr.RegistrationHandler(prodMock, consMock).ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()

		resBody, _ := io.ReadAll(res.Body)
		assert.Equal(t, "Token: 1234", string(resBody))
	})

}
