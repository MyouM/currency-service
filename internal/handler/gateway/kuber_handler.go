package handler

import (
	"fmt"
	"net/http"
)

func (hr HandlerRelations) KuberLivez() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "ok")
	}
}

func (hr HandlerRelations) KuberReadyz() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		errors := make([]error, 0, 2)

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
