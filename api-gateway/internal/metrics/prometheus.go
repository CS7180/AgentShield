package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	WebSocketConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "websocket_connections_active",
		Help: "Number of active WebSocket connections",
	})

	ScanCreationTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "scan_creation_total",
		Help: "Total number of scans created",
	})

	RateLimitExceededTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rate_limit_exceeded_total",
		Help: "Total number of rate limit exceeded events",
	}, []string{"endpoint"})
)
