package middleware

import (
	"currency-service/internal/repository"
	"fmt"
	"net/http"
	"strings"
)

func Validate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		tokenStr := req.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(
				w,
				fmt.Sprint("Incorrect token"),
				http.StatusBadRequest)
			return
		}
		parts := strings.SplitN(tokenStr, " ", 2)
		if !repository.FindToken(parts[1]) {
			http.Error(
				w,
				fmt.Sprint("Incorrect token"),
				http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, req)
	}
}
