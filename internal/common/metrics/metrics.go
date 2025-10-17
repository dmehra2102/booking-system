package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	// HTTP metrics
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestsInFlight prometheus.Gauge

	// User metrics
	UsersTotal   *prometheus.CounterVec
	UsersDeleted *prometheus.CounterVec

	// Business metrics
	BookingsTotal   *prometheus.CounterVec
	BookingDuration *prometheus.HistogramVec

	// kafka metrics
	MessagesProduced *prometheus.CounterVec
	MessagesConsumed *prometheus.CounterVec
	MessageErrors    *prometheus.CounterVec

	// Database metrics
	DBConnections   prometheus.Gauge
	DBQueries       *prometheus.CounterVec
	DBQueryDuration *prometheus.HistogramVec
}

func New(serviceName string) *Metrics {
	return &Metrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "htt_request_duration_seconds",
				Help:      "Duration of HTTP requests in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		RequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "http_response_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
		),
		UsersTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "total_users_created",
				Help:      "Total number of users created",
			},
			[]string{"topic"},
		),
		UsersDeleted: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "total_users_deleted",
				Help:      "Total number of users deleted",
			},
			[]string{"topic"},
		),
		BookingsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "booking_total",
				Help:      "Total number of bookingd",
			},
			[]string{"status", "resource_type"},
		),
		BookingDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "booking_duration_seconds",
				Help:      "Duration of booking operations in seconds",
				Buckets:   []float64{0.1, 0.3, 0.5, 1, 3, 5, 10},
			},
			[]string{"operation"},
		),
		MessagesProduced: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "kafka_messages_produced_total",
				Help:      "Total number of Kafka messages produced",
			},
			[]string{"topic"},
		),
		MessagesConsumed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "kafka_messages_consumed_total",
				Help:      "Total number of Kafka messages consumed",
			},
			[]string{"topic"},
		),
		MessageErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "kafka_message_errors_total",
				Help:      "Total number of Kafka message errors",
			},
			[]string{"topic", "error_type"},
		),
		DBConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "database_connections",
				Help:      "Number of active database connection",
			},
		),
		DBQueries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "database_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "status"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "booking_system",
				Subsystem: serviceName,
				Name:      "database_query_duration_seconds",
				Help:      "Duration of database queries in seconds",
				Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5},
			},
			[]string{"operation"},
		),
	}
}

func (m *Metrics) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		m.RequestsInFlight.Inc()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		m.RequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(status),
		).Inc()

		m.RequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)

		m.RequestsInFlight.Dec()
	}
}

// Handler for Prometheus metrics endpoint
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}
