package middleware

import (
	"currency-service/internal/metrics"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func GrpcMetrics(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrtr := &statusRecorder{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(wrtr, r)

		duration := time.Since(start).Seconds()

		metrics.GrpcRequestTotal.WithLabelValues(
			r.URL.Path, r.Method, http.StatusText(wrtr.statusCode)).Inc()
		metrics.GrpcRequestDuration.WithLabelValues(
			r.URL.Path, r.Method).Observe(duration)
	})
}
