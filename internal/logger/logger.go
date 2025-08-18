package logger

import (
	"log"

	"go.uber.org/zap"
)

var logger = zap.NewNop()

func InitLogger() *zap.Logger {
	lgr, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("zap error: %v", err)
	}

	logger = lgr
	return lgr
}

func GetLogger() *zap.Logger {
	return logger
}
