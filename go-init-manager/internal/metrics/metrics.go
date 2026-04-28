package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "go_init_manager",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "go_init_manager",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	KafkaMessagesProducedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "go_init_manager",
			Name:      "kafka_messages_produced_total",
			Help:      "Total number of Kafka messages produced.",
		},
		[]string{"topic"},
	)

	KafkaMessagesConsumedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "go_init_manager",
			Name:      "kafka_messages_consumed_total",
			Help:      "Total number of Kafka messages consumed.",
		},
		[]string{"topic", "status"},
	)

	DBOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "go_init_manager",
			Name:      "db_operations_total",
			Help:      "Total number of database operations.",
		},
		[]string{"operation", "status"},
	)
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// HTTPMiddleware wraps an http.Handler and records request metrics.
func HTTPMiddleware(path string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		statusStr := strconv.Itoa(rec.status)

		HTTPRequestsTotal.WithLabelValues(r.Method, path, statusStr).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}
