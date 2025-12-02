package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics содержит все метрики приложения
type Metrics struct {
	// HTTP метрики
	HTTPRequestsTotal   prometheus.Counter
	HTTPRequestDuration prometheus.Histogram
	HTTPErrors          prometheus.Counter

	// gRPC метрики
	GRPCRequestsTotal   prometheus.Counter
	GRPCRequestDuration prometheus.Histogram
	GRPCErrors          prometheus.Counter

	// Task метрики
	TasksCreated   prometheus.Counter
	TasksUpdated   prometheus.Counter
	TasksDeleted   prometheus.Counter
	TasksProcessed prometheus.Counter

	// БД метрики
	DBConnections     prometheus.Gauge
	DBConnectionsPool prometheus.Gauge
	DBErrors          prometheus.Counter
	DBQueryDuration   prometheus.Histogram
}

// NewMetrics создает новый экземпляр метрик
func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP метрики
		HTTPRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		}),
		HTTPRequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		HTTPErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total HTTP errors",
		}),

		// gRPC метрики
		GRPCRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total gRPC requests",
		}),
		GRPCRequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		GRPCErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "grpc_errors_total",
			Help: "Total gRPC errors",
		}),

		// Task метрики
		TasksCreated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "tasks_created_total",
			Help: "Total tasks created",
		}),
		TasksUpdated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "tasks_updated_total",
			Help: "Total tasks updated",
		}),
		TasksDeleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "tasks_deleted_total",
			Help: "Total tasks deleted",
		}),
		TasksProcessed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "tasks_processed_total",
			Help: "Total tasks processed",
		}),

		// БД метрики
		DBConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Active database connections",
		}),
		DBConnectionsPool: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "db_connections_pool_total",
			Help: "Total database connections in pool",
		}),
		DBErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "Total database errors",
		}),
		DBQueryDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		}),
	}
}

// Global переменная для метрик (для удобства доступа)
var globalMetrics *Metrics

// GetMetrics возвращает глобальный экземпляр метрик
func GetMetrics() *Metrics {
	if globalMetrics == nil {
		globalMetrics = NewMetrics()
	}
	return globalMetrics
}
