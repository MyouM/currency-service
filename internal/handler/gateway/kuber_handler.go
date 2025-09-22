package handler

import (
	"currency-service/internal/repository/postgres"
	"fmt"
	"net/http"
)

func (hr HandlerRelations) kuberLivez() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "ok")
	}
}

func (hr HandlerRelations) kuberReadyz() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		errors := make([]error, 0, 3)
		if _, _, err := postgres.NewDatabaseConnection(hr.Cfg.Database); err != nil {
			errors = append(errors, err)
		}

		if err := hr.Redis.Ping(); err != nil {
			errors = append(errors, err)
		}
		resp, err := http.Get(hr.Cfg.Currency.URL)
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
