package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total requests",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	RateLimitedRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_rate_limited_total",
			Help: "Requests rejected by rate limiting",
		},
		[]string{"path"},
	)

	BackendFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_backend_failures_total",
			Help: "Backend failures detected by the gateway",
		},
		[]string{"backend"},
	)

	BackendUp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gateway_backend_up",
			Help: "Backend health status (1=up, 0=down)",
		},
		[]string{"backend"},
	)

	CircuitBreakerTrips = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_circuit_breaker_trips_total",
			Help: "Number of times a backend circuit breaker opened",
		},
		[]string{"backend"},
	)

	BackendsRecovered = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_backends_recovered",
			Help: "Number of times a dead backend recovered",
		},
		[]string{"backend"},
	)

	ActiveRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gateway_active_requests",
			Help: "Current requests being processed",
		},
	)

	ResponseBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_response_bytes_total",
			Help: "Total response bytes sent",
		},
		[]string{"path"},
	)
)

func Init() {
	prometheus.MustRegister(
		RequestsTotal,
		RequestDuration,
		RateLimitedRequests,
		BackendFailures,
		BackendUp,
		CircuitBreakerTrips,
		BackendsRecovered,
		ActiveRequests,
		ResponseBytes,
	)
}
