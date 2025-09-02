package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	grpcRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_request_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)
)

type Prometh struct {
	GrpcRequestTotal *prometheus.CounterVec
}

func InitPrometh() *Prometh {
	prometh := Prometh{
		GrpcRequestTotal: grpcRequestTotal,
	}

	prometheus.MustRegister(grpcRequestTotal)

	return &prometh
}
