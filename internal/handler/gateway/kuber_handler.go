package handler

import (
	"currency-service/internal/config"
	"currency-service/internal/repository/postgres"
	"currency-service/internal/repository/redis"
	"fmt"
	"net/http"
)

func kuberLivez(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "ok")
}

func kuberReadyz(cfg *config.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		errors := make([]error, 0, 3)
		if _, _, err := postgres.NewDatabaseConnection(cfg.Database); err != nil {
			errors = append(errors, err)
		}

		redis := redis.GetRedisClient()
		if err := redis.Ping(); err != nil {
			errors = append(errors, err)
		}
		resp, err := http.Get(cfg.Currency.URL)
		if err != nil || resp.StatusCode != http.StatusOK {
			errors = append(errors, err)
		}
		if len(errors) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			for i, err := range errors {
				_, _ = fmt.Fprintf(w, "%d: %v\n", i, err)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "ok")
	}
}
