package test

import (
	handler "currency-service/internal/handler/gateway"
	"currency-service/internal/proto/currpb"
	mock "currency-service/internal/proto/mock"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGrpcHandlers(t *testing.T) {
	var hr handler.HandlerRelations

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	grpcMock := mock.NewMockCurrencyServiceClient(ctrl)

	t.Run("GetOneCurrencyRate", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test/1234", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		mux.HandleFunc("/test/{date}", hr.GetOneCurrencyRate(grpcMock))

		grpcMock.
			EXPECT().
			GetSpecificCurrency(
				req.Context(),
				&currpb.ClientSpecRequest{
					Date: "1234",
				}).
			Return(&currpb.ClientSpecResponse{
				Currency: 1.0,
			},
				nil)

		mux.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()

		body, _ := io.ReadAll(res.Body)

		expect := fmt.Sprintf("OK: %f\n", 1.0)
		assert.Equal(t, expect, string(body))
	})

	t.Run("GetIntervalCurrencyChanges", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test/1234-4321", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		mux.HandleFunc("/test/{dates}", hr.GetIntervalCurrencyChanges(grpcMock))
		grpcMock.
			EXPECT().
			GetIntervalCurrency(
				req.Context(),
				&currpb.ClientIntervalRequest{
					DateBegin: "1234",
					DateEnd:   "4321",
				}).
			Return(&currpb.ClientIntervalResponse{
				Rates: []*currpb.CurrencyRates{{Date: "1234", Rate: 1.0},
					{Date: "4321", Rate: 2.0}},
			},
				nil)
		mux.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()

		body, _ := io.ReadAll(res.Body)
		expect := fmt.Sprintf("%s, %f\n%s, %f\n", "1234", 1.0, "4321", 2.0)
		assert.Equal(t, expect, string(body))
	})
}
