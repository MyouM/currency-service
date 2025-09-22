package middleware

import (
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
