package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics wraps Prometheus metrics with additional functionality
type Metrics struct {
	config *config.ObservabilityConfig

	// HTTP metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec

	// Database metrics
	dbOperationsTotal   *prometheus.CounterVec
	dbOperationDuration *prometheus.HistogramVec
	dbConnectionsActive *prometheus.GaugeVec

	// Business metrics
	businessClassificationsTotal *prometheus.CounterVec
	classificationDuration       *prometheus.HistogramVec
	riskAssessmentsTotal         *prometheus.CounterVec
	complianceChecksTotal        *prometheus.CounterVec

	// External service metrics
	externalServiceCallsTotal   *prometheus.CounterVec
	externalServiceCallDuration *prometheus.HistogramVec

	// System metrics
	goroutinesActive *prometheus.GaugeVec
	memoryUsage      *prometheus.GaugeVec
	cpuUsage         *prometheus.GaugeVec

	// Custom metrics
	customMetrics map[string]prometheus.Collector
}

// NewMetrics creates a new metrics collector
func NewMetrics(cfg *config.ObservabilityConfig) (*Metrics, error) {
	if !cfg.MetricsEnabled {
		return &Metrics{
			config:        cfg,
			customMetrics: make(map[string]prometheus.Collector),
		}, nil
	}

	metrics := &Metrics{
		config:        cfg,
		customMetrics: make(map[string]prometheus.Collector),
	}

	// Initialize HTTP metrics
	metrics.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	metrics.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	metrics.httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
		[]string{"method", "path"},
	)

	// Initialize database metrics
	metrics.dbOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "table", "status"},
	)

	metrics.dbOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_operation_duration_seconds",
			Help:    "Database operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	metrics.dbConnectionsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Current number of active database connections",
		},
		[]string{"database"},
	)

	// Initialize business metrics
	metrics.businessClassificationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "business_classifications_total",
			Help: "Total number of business classifications",
		},
		[]string{"status", "confidence_level"},
	)

	metrics.classificationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "classification_duration_seconds",
			Help:    "Duration of classification operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"}, // single|batch
	)

	metrics.riskAssessmentsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "risk_assessments_total",
			Help: "Total number of risk assessments",
		},
		[]string{"status", "risk_level"},
	)

	metrics.complianceChecksTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "compliance_checks_total",
			Help: "Total number of compliance checks",
		},
		[]string{"status", "compliance_type"},
	)

	// Initialize external service metrics
	metrics.externalServiceCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "external_service_calls_total",
			Help: "Total number of external service calls",
		},
		[]string{"service", "endpoint", "status_code"},
	)

	metrics.externalServiceCallDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "external_service_call_duration_seconds",
			Help:    "External service call duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "endpoint"},
	)

	// Initialize system metrics
	metrics.goroutinesActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "goroutines_active",
			Help: "Current number of active goroutines",
		},
		[]string{"component"},
	)

	metrics.memoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
		[]string{"type"},
	)

	metrics.cpuUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cpu_usage_percent",
			Help: "Current CPU usage percentage",
		},
		[]string{"type"},
	)

	// Register all metrics
	collectors := []prometheus.Collector{
		metrics.httpRequestsTotal,
		metrics.httpRequestDuration,
		metrics.httpRequestsInFlight,
		metrics.dbOperationsTotal,
		metrics.dbOperationDuration,
		metrics.dbConnectionsActive,
		metrics.businessClassificationsTotal,
		metrics.classificationDuration,
		metrics.riskAssessmentsTotal,
		metrics.complianceChecksTotal,
		metrics.externalServiceCallsTotal,
		metrics.externalServiceCallDuration,
		metrics.goroutinesActive,
		metrics.memoryUsage,
		metrics.cpuUsage,
	}

	for _, collector := range collectors {
		if err := prometheus.Register(collector); err != nil {
			return nil, fmt.Errorf("failed to register metric: %w", err)
		}
	}

	return metrics, nil
}

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	if !m.config.MetricsEnabled {
		return
	}

	m.httpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", statusCode)).Inc()
	m.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordHTTPRequestStart records the start of an HTTP request
func (m *Metrics) RecordHTTPRequestStart(method, path string) {
	if !m.config.MetricsEnabled {
		return
	}

	m.httpRequestsInFlight.WithLabelValues(method, path).Inc()
}

// RecordHTTPRequestEnd records the end of an HTTP request
func (m *Metrics) RecordHTTPRequestEnd(method, path string) {
	if !m.config.MetricsEnabled {
		return
	}

	m.httpRequestsInFlight.WithLabelValues(method, path).Dec()
}

// RecordDatabaseOperation records a database operation
func (m *Metrics) RecordDatabaseOperation(operation, table string, status string, duration time.Duration) {
	if !m.config.MetricsEnabled {
		return
	}

	m.dbOperationsTotal.WithLabelValues(operation, table, status).Inc()
	m.dbOperationDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordDatabaseConnections records the number of active database connections
func (m *Metrics) RecordDatabaseConnections(database string, count int) {
	if !m.config.MetricsEnabled {
		return
	}

	m.dbConnectionsActive.WithLabelValues(database).Set(float64(count))
}

// RecordBusinessClassification records a business classification
func (m *Metrics) RecordBusinessClassification(status, confidenceLevel string) {
	if !m.config.MetricsEnabled {
		return
	}

	m.businessClassificationsTotal.WithLabelValues(status, confidenceLevel).Inc()
}

// RecordClassificationDuration records the duration of a classification operation
func (m *Metrics) RecordClassificationDuration(opType string, duration time.Duration) {
	if !m.config.MetricsEnabled {
		return
	}
	m.classificationDuration.WithLabelValues(opType).Observe(duration.Seconds())
}

// RecordRiskAssessment records a risk assessment
func (m *Metrics) RecordRiskAssessment(status, riskLevel string) {
	if !m.config.MetricsEnabled {
		return
	}

	m.riskAssessmentsTotal.WithLabelValues(status, riskLevel).Inc()
}

// RecordComplianceCheck records a compliance check
func (m *Metrics) RecordComplianceCheck(status, complianceType string) {
	if !m.config.MetricsEnabled {
		return
	}

	m.complianceChecksTotal.WithLabelValues(status, complianceType).Inc()
}

// RecordExternalServiceCall records an external service call
func (m *Metrics) RecordExternalServiceCall(service, endpoint string, statusCode int, duration time.Duration) {
	if !m.config.MetricsEnabled {
		return
	}

	m.externalServiceCallsTotal.WithLabelValues(service, endpoint, fmt.Sprintf("%d", statusCode)).Inc()
	m.externalServiceCallDuration.WithLabelValues(service, endpoint).Observe(duration.Seconds())
}

// RecordGoroutines records the number of active goroutines
func (m *Metrics) RecordGoroutines(component string, count int) {
	if !m.config.MetricsEnabled {
		return
	}

	m.goroutinesActive.WithLabelValues(component).Set(float64(count))
}

// RecordMemoryUsage records memory usage
func (m *Metrics) RecordMemoryUsage(memoryType string, bytes int64) {
	if !m.config.MetricsEnabled {
		return
	}

	m.memoryUsage.WithLabelValues(memoryType).Set(float64(bytes))
}

// RecordCPUUsage records CPU usage
func (m *Metrics) RecordCPUUsage(cpuType string, percentage float64) {
	if !m.config.MetricsEnabled {
		return
	}

	m.cpuUsage.WithLabelValues(cpuType).Set(percentage)
}

// AddCustomMetric adds a custom metric
func (m *Metrics) AddCustomMetric(name string, collector prometheus.Collector) error {
	if !m.config.MetricsEnabled {
		return nil
	}

	if err := prometheus.Register(collector); err != nil {
		return fmt.Errorf("failed to register custom metric %s: %w", name, err)
	}

	m.customMetrics[name] = collector
	return nil
}

// RemoveCustomMetric removes a custom metric
func (m *Metrics) RemoveCustomMetric(name string) error {
	if !m.config.MetricsEnabled {
		return nil
	}

	if collector, exists := m.customMetrics[name]; exists {
		if prometheus.Unregister(collector) {
			delete(m.customMetrics, name)
		}
	}

	return nil
}

// ServeHTTP serves the metrics endpoint
func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !m.config.MetricsEnabled {
		http.Error(w, "Metrics disabled", http.StatusServiceUnavailable)
		return
	}

	promhttp.Handler().ServeHTTP(w, r)
}

// StartMetricsServer starts the metrics server
func (m *Metrics) StartMetricsServer(ctx context.Context) error {
	if !m.config.MetricsEnabled {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(m.config.MetricsPath, m)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", m.config.MetricsPort),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error but don't return it since this is running in a goroutine
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}

// IsEnabled returns whether metrics are enabled
func (m *Metrics) IsEnabled() bool {
	return m.config.MetricsEnabled
}

// String returns a string representation of the metrics configuration
func (m *Metrics) String() string {
	return fmt.Sprintf("Metrics{enabled=%t, port=%d, path=%s}", m.config.MetricsEnabled, m.config.MetricsPort, m.config.MetricsPath)
}
