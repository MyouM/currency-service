package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	GrpcRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_request_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"path", "method", "status"},
	)
	GrpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Histogram of request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)
)

type Prometh struct {
	GrpcRequestTotal *prometheus.CounterVec
}

func InitPrometh() {
	prometheus.MustRegister(GrpcRequestTotal)
	prometheus.MustRegister(GrpcRequestDuration)
}
