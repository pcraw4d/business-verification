package observability

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MonitoringSystem provides comprehensive application monitoring
type MonitoringSystem struct {
	logger *zap.Logger

	// Prometheus metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec

	// Business metrics
	classificationRequestsTotal  *prometheus.CounterVec
	classificationAccuracy       *prometheus.GaugeVec
	classificationDuration       *prometheus.HistogramVec
	riskAssessmentRequestsTotal  *prometheus.CounterVec
	riskAssessmentDuration       *prometheus.HistogramVec
	complianceCheckRequestsTotal *prometheus.CounterVec
	complianceCheckDuration      *prometheus.HistogramVec

	// System metrics
	systemMemoryUsage *prometheus.GaugeVec
	systemCPUUsage    *prometheus.GaugeVec
	systemGoroutines  *prometheus.GaugeVec
	systemHeapAlloc   *prometheus.GaugeVec
	systemHeapSys     *prometheus.GaugeVec

	// Database metrics
	databaseConnections   *prometheus.GaugeVec
	databaseQueryDuration *prometheus.HistogramVec
	databaseErrors        *prometheus.CounterVec

	// External API metrics
	externalAPICalls    *prometheus.CounterVec
	externalAPIDuration *prometheus.HistogramVec
	externalAPIErrors   *prometheus.CounterVec

	// Health check metrics
	healthCheckStatus   *prometheus.GaugeVec
	healthCheckDuration *prometheus.HistogramVec

	// Custom business metrics
	activeUsers          *prometheus.GaugeVec
	apiKeyUsage          *prometheus.CounterVec
	rateLimitHits        *prometheus.CounterVec
	authenticationEvents *prometheus.CounterVec
}

// NewMonitoringSystem creates a new monitoring system
func NewMonitoringSystem(logger *zap.Logger) *MonitoringSystem {
	ms := &MonitoringSystem{
		logger: logger,
	}

	ms.initializeMetrics()
	return ms
}

// initializeMetrics initializes all Prometheus metrics
func (ms *MonitoringSystem) initializeMetrics() {
	// HTTP metrics
	ms.httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyb_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code", "environment"},
	)

	ms.httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kyb_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "environment"},
	)

	ms.httpRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kyb_http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
		[]string{"method", "endpoint", "environment"},
	)

	// Business metrics
	ms.classificationRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyb_classification_requests_total",
			Help: "Total number of classification requests",
		},
		[]string{"method", "confidence_level", "environment"},
	)

	ms.classificationAccuracy = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kyb_classification_accuracy",
			Help: "Classification accuracy percentage",
		},
		[]string{"method", "environment"},
	)

	ms.classificationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kyb_classification_duration_seconds",
			Help:    "Classification request duration in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"method", "environment"},
	)

	ms.riskAssessmentRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyb_risk_assessment_requests_total",
			Help: "Total number of risk assessment requests",
		},
		[]string{"risk_level", "environment"},
	)

	ms.riskAssessmentDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kyb_risk_assessment_duration_seconds",
			Help:    "Risk assessment duration in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"environment"},
	)

	ms.complianceCheckRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyb_compliance_check_requests_total",
			Help: "Total number of compliance check requests",
		},
		[]string{"framework", "status", "environment"},
	)

	ms.complianceCheckDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kyb_compliance_check_duration_seconds",
			Help:    "Compliance check duration in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"framework", "environment"},
	)

	// System metrics
	ms.systemMemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kyb_system_memory_usage_bytes",
			Help: "System memory usage in bytes",
		},
		[]string{"type", "environment"},
	)

	ms.systemCPUUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kyb_system_cpu_usage_percent",
			Help: "System CPU usage percentage",
		},
		[]string{"environment"},
	)

	ms.systemGoroutines = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kyb_system_goroutines",
			Help: "Number of active goroutines",
		},
		[]string{"environment"},
	)

	ms.systemHeapAlloc = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kyb_system_heap_alloc_bytes",
			Help: "Heap memory allocated in bytes",
		},
		[]string{"environment"},
	)

	ms.systemHeapSys = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kyb_system_heap_sys_bytes",
			Help: "Heap memory system in bytes",
		},
		[]string{"environment"},
	)

	// Database metrics
	ms.databaseConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kyb_database_connections",
			Help: "Number of active database connections",
		},
		[]string{"database", "environment"},
	)

	ms.databaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kyb_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"database", "query_type", "environment"},
	)

	ms.databaseErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyb_database_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"database", "error_type", "environment"},
	)

	// External API metrics
	ms.externalAPICalls = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyb_external_api_calls_total",
			Help: "Total number of external API calls",
		},
		[]string{"provider", "endpoint", "status", "environment"},
	)

	ms.externalAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kyb_external_api_duration_seconds",
			Help:    "External API call duration in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0},
		},
		[]string{"provider", "endpoint", "environment"},
	)

	ms.externalAPIErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyb_external_api_errors_total",
			Help: "Total number of external API errors",
		},
		[]string{"provider", "endpoint", "error_type", "environment"},
	)

	// Health check metrics
	ms.healthCheckStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kyb_health_check_status",
			Help: "Health check status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"component", "environment"},
	)

	ms.healthCheckDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kyb_health_check_duration_seconds",
			Help:    "Health check duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
		},
		[]string{"component", "environment"},
	)

	// Custom business metrics
	ms.activeUsers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kyb_active_users",
			Help: "Number of active users",
		},
		[]string{"user_type", "environment"},
	)

	ms.apiKeyUsage = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyb_api_key_usage_total",
			Help: "Total number of API key usages",
		},
		[]string{"api_key_id", "endpoint", "environment"},
	)

	ms.rateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyb_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"api_key_id", "endpoint", "environment"},
	)

	ms.authenticationEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyb_authentication_events_total",
			Help: "Total number of authentication events",
		},
		[]string{"event_type", "status", "environment"},
	)
}

// RecordHTTPRequest records HTTP request metrics
func (ms *MonitoringSystem) RecordHTTPRequest(method, endpoint, statusCode, environment string, duration time.Duration) {
	ms.httpRequestsTotal.WithLabelValues(method, endpoint, statusCode, environment).Inc()
	ms.httpRequestDuration.WithLabelValues(method, endpoint, environment).Observe(duration.Seconds())
}

// RecordHTTPRequestInFlight records in-flight HTTP requests
func (ms *MonitoringSystem) RecordHTTPRequestInFlight(method, endpoint, environment string, count int) {
	ms.httpRequestsInFlight.WithLabelValues(method, endpoint, environment).Set(float64(count))
}

// RecordClassificationRequest records classification request metrics
func (ms *MonitoringSystem) RecordClassificationRequest(method, confidenceLevel, environment string, duration time.Duration) {
	ms.classificationRequestsTotal.WithLabelValues(method, confidenceLevel, environment).Inc()
	ms.classificationDuration.WithLabelValues(method, environment).Observe(duration.Seconds())
}

// RecordClassificationAccuracy records classification accuracy
func (ms *MonitoringSystem) RecordClassificationAccuracy(method, environment string, accuracy float64) {
	ms.classificationAccuracy.WithLabelValues(method, environment).Set(accuracy)
}

// RecordRiskAssessment records risk assessment metrics
func (ms *MonitoringSystem) RecordRiskAssessment(riskLevel, environment string, duration time.Duration) {
	ms.riskAssessmentRequestsTotal.WithLabelValues(riskLevel, environment).Inc()
	ms.riskAssessmentDuration.WithLabelValues(environment).Observe(duration.Seconds())
}

// RecordComplianceCheck records compliance check metrics
func (ms *MonitoringSystem) RecordComplianceCheck(framework, status, environment string, duration time.Duration) {
	ms.complianceCheckRequestsTotal.WithLabelValues(framework, status, environment).Inc()
	ms.complianceCheckDuration.WithLabelValues(framework, environment).Observe(duration.Seconds())
}

// RecordSystemMetrics records system metrics
func (ms *MonitoringSystem) RecordSystemMetrics(environment string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	ms.systemMemoryUsage.WithLabelValues("alloc", environment).Set(float64(m.Alloc))
	ms.systemMemoryUsage.WithLabelValues("sys", environment).Set(float64(m.Sys))
	ms.systemGoroutines.WithLabelValues(environment).Set(float64(runtime.NumGoroutine()))
	ms.systemHeapAlloc.WithLabelValues(environment).Set(float64(m.HeapAlloc))
	ms.systemHeapSys.WithLabelValues(environment).Set(float64(m.HeapSys))
}

// RecordDatabaseMetrics records database metrics
func (ms *MonitoringSystem) RecordDatabaseMetrics(database, queryType, environment string, duration time.Duration) {
	ms.databaseQueryDuration.WithLabelValues(database, queryType, environment).Observe(duration.Seconds())
}

// RecordDatabaseError records database errors
func (ms *MonitoringSystem) RecordDatabaseError(database, errorType, environment string) {
	ms.databaseErrors.WithLabelValues(database, errorType, environment).Inc()
}

// RecordDatabaseConnections records database connection count
func (ms *MonitoringSystem) RecordDatabaseConnections(database, environment string, count int) {
	ms.databaseConnections.WithLabelValues(database, environment).Set(float64(count))
}

// RecordExternalAPICall records external API call metrics
func (ms *MonitoringSystem) RecordExternalAPICall(provider, endpoint, status, environment string, duration time.Duration) {
	ms.externalAPICalls.WithLabelValues(provider, endpoint, status, environment).Inc()
	ms.externalAPIDuration.WithLabelValues(provider, endpoint, environment).Observe(duration.Seconds())
}

// RecordExternalAPIError records external API errors
func (ms *MonitoringSystem) RecordExternalAPIError(provider, endpoint, errorType, environment string) {
	ms.externalAPIErrors.WithLabelValues(provider, endpoint, errorType, environment).Inc()
}

// RecordHealthCheck records health check metrics
func (ms *MonitoringSystem) RecordHealthCheck(component, environment string, healthy bool, duration time.Duration) {
	status := 0.0
	if healthy {
		status = 1.0
	}

	ms.healthCheckStatus.WithLabelValues(component, environment).Set(status)
	ms.healthCheckDuration.WithLabelValues(component, environment).Observe(duration.Seconds())
}

// RecordActiveUsers records active user count
func (ms *MonitoringSystem) RecordActiveUsers(userType, environment string, count int) {
	ms.activeUsers.WithLabelValues(userType, environment).Set(float64(count))
}

// RecordAPIKeyUsage records API key usage
func (ms *MonitoringSystem) RecordAPIKeyUsage(apiKeyID, endpoint, environment string) {
	ms.apiKeyUsage.WithLabelValues(apiKeyID, endpoint, environment).Inc()
}

// RecordRateLimitHit records rate limit hits
func (ms *MonitoringSystem) RecordRateLimitHit(apiKeyID, endpoint, environment string) {
	ms.rateLimitHits.WithLabelValues(apiKeyID, endpoint, environment).Inc()
}

// RecordAuthenticationEvent records authentication events
func (ms *MonitoringSystem) RecordAuthenticationEvent(eventType, status, environment string) {
	ms.authenticationEvents.WithLabelValues(eventType, status, environment).Inc()
}

// StartMetricsCollection starts periodic metrics collection
func (ms *MonitoringSystem) StartMetricsCollection(ctx context.Context, environment string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ms.RecordSystemMetrics(environment)
		}
	}
}

// MetricsHandler returns the Prometheus metrics handler
func (ms *MonitoringSystem) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// HealthCheckHandler returns a health check handler with metrics
func (ms *MonitoringSystem) HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Perform health checks
		healthy := ms.performHealthChecks(r.Context())

		// Record metrics
		ms.RecordHealthCheck("overall", "production", healthy, time.Since(start))

		if healthy {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().UTC().Format(time.RFC3339) + `"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status": "unhealthy", "timestamp": "` + time.Now().UTC().Format(time.RFC3339) + `"}`))
		}
	}
}

// performHealthChecks performs all health checks
func (ms *MonitoringSystem) performHealthChecks(ctx context.Context) bool {
	// Database health check
	dbHealthy := ms.checkDatabaseHealth(ctx)
	ms.RecordHealthCheck("database", "production", dbHealthy, time.Duration(0))

	// External API health check
	apiHealthy := ms.checkExternalAPIHealth(ctx)
	ms.RecordHealthCheck("external_api", "production", apiHealthy, time.Duration(0))

	// System health check
	systemHealthy := ms.checkSystemHealth(ctx)
	ms.RecordHealthCheck("system", "production", systemHealthy, time.Duration(0))

	return dbHealthy && apiHealthy && systemHealthy
}

// checkDatabaseHealth checks database connectivity
func (ms *MonitoringSystem) checkDatabaseHealth(ctx context.Context) bool {
	// This would be implemented based on your database connection
	// For now, return true as a placeholder
	return true
}

// checkExternalAPIHealth checks external API connectivity
func (ms *MonitoringSystem) checkExternalAPIHealth(ctx context.Context) bool {
	// This would be implemented based on your external API dependencies
	// For now, return true as a placeholder
	return true
}

// checkSystemHealth checks system resources
func (ms *MonitoringSystem) checkSystemHealth(ctx context.Context) bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Check if memory usage is reasonable (less than 80% of system memory)
	memoryUsagePercent := float64(m.Sys) / float64(1<<30) * 100 // Assuming 1GB system memory
	if memoryUsagePercent > 80 {
		ms.logger.Warn("High memory usage detected", zap.Float64("usage_percent", memoryUsagePercent))
		return false
	}

	// Check if goroutine count is reasonable (less than 1000)
	if runtime.NumGoroutine() > 1000 {
		ms.logger.Warn("High goroutine count detected", zap.Int("goroutines", runtime.NumGoroutine()))
		return false
	}

	return true
}

// GetMetricsSummary returns a summary of current metrics
func (ms *MonitoringSystem) GetMetricsSummary(environment string) map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"environment": environment,
		"system": map[string]interface{}{
			"goroutines":   runtime.NumGoroutine(),
			"memory_alloc": m.Alloc,
			"memory_sys":   m.Sys,
			"heap_alloc":   m.HeapAlloc,
			"heap_sys":     m.HeapSys,
		},
		"version": "1.0.0",
		"uptime":  time.Since(time.Now()).String(), // This would be actual uptime
	}
}
