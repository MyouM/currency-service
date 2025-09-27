package middleware

import (
	"currency-service/internal/config"
	"currency-service/internal/kafka"
	"currency-service/internal/repository/redis"
	"fmt"
	"net/http"
)

func Validate(rds redis.RedisFuncs, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		tokenStr := req.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(
				w,
				fmt.Sprint("Incorrect token"),
				http.StatusBadRequest)
			return
		}

		if err := rds.FindToken(tokenStr); err != nil {
			http.Error(
				w,
				fmt.Sprint("Incorrect token"),
				http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func KafkaInit(cfg *config.AppConfig, next func(kafka.ProducerFuncs, kafka.ConsumerFuncs) http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		consumer := kafka.NewConsumer(
			cfg.Kafka.ControllerHost,
			kafka.AuthGatewayTopic,
			kafka.GroupID)
		producer := kafka.NewProducer(
			cfg.Kafka.ControllerHost,
			kafka.GatewayAuthTopic)

		next(producer, consumer).ServeHTTP(w, req)
	})
}
