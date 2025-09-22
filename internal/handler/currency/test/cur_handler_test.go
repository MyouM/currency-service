package test

import (
	"context"
	handler "currency-service/internal/handler/currency"
	"currency-service/internal/proto/currpb"
	mockDB "currency-service/internal/repository/postgres/mock"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetSpecificCurrency(t *testing.T) {
	req := &currpb.ClientSpecRequest{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPsql := mockDB.NewMockCurrencyPsqlFuncs(ctrl)
	srvr := handler.Server{
		Psql: mockPsql,
	}

	t.Run("value in db", func(t *testing.T) {
		var timeString = "2000-01-01"
		timePars, err := time.Parse("2006-01-02", timeString)
		if err != nil {
			fmt.Println("cannot parse time")
			return
		}
		mockPsql.EXPECT().GetOneCurrencyRate(timePars).Return(1.0, nil)
		req.Date = timeString
		resp, err := srvr.GetSpecificCurrency(context.Background(), req)
		assert.Equal(t, 1.0, resp.Currency)
		assert.NoError(t, err)
	})
}

func TestGetIntervalCurrency(t *testing.T) {
	req := &currpb.ClientIntervalRequest{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPsql := mockDB.NewMockCurrencyPsqlFuncs(ctrl)
	srvr := handler.Server{
		Psql: mockPsql,
	}

	t.Run("value in db", func(t *testing.T) {
		var (
			timeString1 = "2000-01-01"
			timeString2 = "2000-01-02"
		)
		expectRate := GetCurrencyChanges{
			{Rate: 1.0},
			{Rate: 2.0},
		}
		timePars1, err := time.Parse("2006-01-02", timeString1)
		if err != nil {
			fmt.Println("cannot parse time")
			return
		}
		timePars2, err := time.Parse("2006-01-02", timeString2)
		if err != nil {
			fmt.Println("cannot parse time")
			return
		}
		mockPsql.EXPECT().
			GetCurrencyChanges(timePars1, timePars2).
			Return(expectRate, nil)
		req.DateBegin = timeString1
		req.DateEnd = timeString2
		resp, err := srvr.GetIntervalCurrency(context.Background(), req)
		assert.Equal(t, expectRate[0].Rate, resp.Rates[0].Rate)
		assert.NoError(t, err)
	})
}
