package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// PrometheusMetrics provides comprehensive metrics collection for the Risk Assessment Service
type PrometheusMetrics struct {
	// HTTP Metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestSize     *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec

	// Business Metrics
	riskAssessmentsTotal     *prometheus.CounterVec
	riskAssessmentDuration   *prometheus.HistogramVec
	riskScoreDistribution    *prometheus.HistogramVec
	complianceChecksTotal    *prometheus.CounterVec
	sanctionsScreeningsTotal *prometheus.CounterVec
	adverseMediaChecksTotal  *prometheus.CounterVec

	// System Metrics
	activeConnections   prometheus.Gauge
	databaseConnections *prometheus.GaugeVec
	cacheHitRate        *prometheus.GaugeVec
	externalAPICalls    *prometheus.CounterVec
	externalAPIDuration *prometheus.HistogramVec

	// Error Metrics
	errorsTotal      *prometheus.CounterVec
	errorRate        *prometheus.GaugeVec
	timeoutErrors    *prometheus.CounterVec
	validationErrors *prometheus.CounterVec

	// Performance Metrics
	responseTimeP50 *prometheus.GaugeVec
	responseTimeP95 *prometheus.GaugeVec
	responseTimeP99 *prometheus.GaugeVec
	throughput      *prometheus.GaugeVec

	// Tenant Metrics
	tenantRequests  *prometheus.CounterVec
	tenantDataUsage *prometheus.GaugeVec
	tenantErrorRate *prometheus.GaugeVec

	// Compliance Metrics
	auditEventsTotal     *prometheus.CounterVec
	complianceViolations *prometheus.CounterVec
	securityIncidents    *prometheus.CounterVec

	logger *zap.Logger
}

// NewPrometheusMetrics creates a new Prometheus metrics collector
func NewPrometheusMetrics(logger *zap.Logger) *PrometheusMetrics {
	return &PrometheusMetrics{
		// HTTP Metrics
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_assessment_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code", "tenant_id"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "risk_assessment_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "tenant_id"},
		),
		httpRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "risk_assessment_http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "endpoint", "tenant_id"},
		),
		httpResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "risk_assessment_http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "endpoint", "tenant_id"},
		),

		// Business Metrics
		riskAssessmentsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_assessment_total",
				Help: "Total number of risk assessments performed",
			},
			[]string{"assessment_type", "risk_level", "tenant_id", "country"},
		),
		riskAssessmentDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "risk_assessment_duration_seconds",
				Help:    "Risk assessment duration in seconds",
				Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 25, 50, 100},
			},
			[]string{"assessment_type", "tenant_id"},
		),
		riskScoreDistribution: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "risk_assessment_score_distribution",
				Help:    "Distribution of risk assessment scores",
				Buckets: []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
			},
			[]string{"assessment_type", "tenant_id"},
		),
		complianceChecksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "compliance_checks_total",
				Help: "Total number of compliance checks performed",
			},
			[]string{"regulation", "status", "tenant_id"},
		),
		sanctionsScreeningsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sanctions_screenings_total",
				Help: "Total number of sanctions screenings performed",
			},
			[]string{"list_type", "match_status", "tenant_id"},
		),
		adverseMediaChecksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "adverse_media_checks_total",
				Help: "Total number of adverse media checks performed",
			},
			[]string{"severity", "tenant_id"},
		),

		// System Metrics
		activeConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "risk_assessment_active_connections",
				Help: "Number of active connections",
			},
		),
		databaseConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "risk_assessment_database_connections",
				Help: "Number of database connections",
			},
			[]string{"state", "database"},
		),
		cacheHitRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "risk_assessment_cache_hit_rate",
				Help: "Cache hit rate percentage",
			},
			[]string{"cache_type"},
		),
		externalAPICalls: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_assessment_external_api_calls_total",
				Help: "Total number of external API calls",
			},
			[]string{"provider", "endpoint", "status", "tenant_id"},
		),
		externalAPIDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "risk_assessment_external_api_duration_seconds",
				Help:    "External API call duration in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
			},
			[]string{"provider", "endpoint", "tenant_id"},
		),

		// Error Metrics
		errorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_assessment_errors_total",
				Help: "Total number of errors",
			},
			[]string{"error_type", "component", "tenant_id"},
		),
		errorRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "risk_assessment_error_rate",
				Help: "Error rate percentage",
			},
			[]string{"component", "tenant_id"},
		),
		timeoutErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_assessment_timeout_errors_total",
				Help: "Total number of timeout errors",
			},
			[]string{"component", "tenant_id"},
		),
		validationErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_assessment_validation_errors_total",
				Help: "Total number of validation errors",
			},
			[]string{"validation_type", "tenant_id"},
		),

		// Performance Metrics
		responseTimeP50: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "risk_assessment_response_time_p50_seconds",
				Help: "50th percentile response time in seconds",
			},
			[]string{"endpoint", "tenant_id"},
		),
		responseTimeP95: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "risk_assessment_response_time_p95_seconds",
				Help: "95th percentile response time in seconds",
			},
			[]string{"endpoint", "tenant_id"},
		),
		responseTimeP99: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "risk_assessment_response_time_p99_seconds",
				Help: "99th percentile response time in seconds",
			},
			[]string{"endpoint", "tenant_id"},
		),
		throughput: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "risk_assessment_throughput_requests_per_second",
				Help: "Requests per second throughput",
			},
			[]string{"endpoint", "tenant_id"},
		),

		// Tenant Metrics
		tenantRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_assessment_tenant_requests_total",
				Help: "Total number of requests per tenant",
			},
			[]string{"tenant_id", "endpoint"},
		),
		tenantDataUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "risk_assessment_tenant_data_usage_bytes",
				Help: "Data usage per tenant in bytes",
			},
			[]string{"tenant_id", "data_type"},
		),
		tenantErrorRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "risk_assessment_tenant_error_rate",
				Help: "Error rate per tenant",
			},
			[]string{"tenant_id"},
		),

		// Compliance Metrics
		auditEventsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_assessment_audit_events_total",
				Help: "Total number of audit events",
			},
			[]string{"event_type", "tenant_id", "user_id"},
		),
		complianceViolations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_assessment_compliance_violations_total",
				Help: "Total number of compliance violations",
			},
			[]string{"violation_type", "regulation", "tenant_id"},
		),
		securityIncidents: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_assessment_security_incidents_total",
				Help: "Total number of security incidents",
			},
			[]string{"incident_type", "severity", "tenant_id"},
		),

		logger: logger,
	}
}

// RecordHTTPRequest records HTTP request metrics
func (pm *PrometheusMetrics) RecordHTTPRequest(method, endpoint, statusCode, tenantID string, duration time.Duration, requestSize, responseSize int64) {
	pm.httpRequestsTotal.WithLabelValues(method, endpoint, statusCode, tenantID).Inc()
	pm.httpRequestDuration.WithLabelValues(method, endpoint, tenantID).Observe(duration.Seconds())
	pm.httpRequestSize.WithLabelValues(method, endpoint, tenantID).Observe(float64(requestSize))
	pm.httpResponseSize.WithLabelValues(method, endpoint, tenantID).Observe(float64(responseSize))
	pm.tenantRequests.WithLabelValues(tenantID, endpoint).Inc()
}

// RecordRiskAssessment records risk assessment metrics
func (pm *PrometheusMetrics) RecordRiskAssessment(assessmentType, riskLevel, tenantID, country string, duration time.Duration, score float64) {
	pm.riskAssessmentsTotal.WithLabelValues(assessmentType, riskLevel, tenantID, country).Inc()
	pm.riskAssessmentDuration.WithLabelValues(assessmentType, tenantID).Observe(duration.Seconds())
	pm.riskScoreDistribution.WithLabelValues(assessmentType, tenantID).Observe(score)
}

// RecordComplianceCheck records compliance check metrics
func (pm *PrometheusMetrics) RecordComplianceCheck(regulation, status, tenantID string) {
	pm.complianceChecksTotal.WithLabelValues(regulation, status, tenantID).Inc()
}

// RecordSanctionsScreening records sanctions screening metrics
func (pm *PrometheusMetrics) RecordSanctionsScreening(listType, matchStatus, tenantID string) {
	pm.sanctionsScreeningsTotal.WithLabelValues(listType, matchStatus, tenantID).Inc()
}

// RecordAdverseMediaCheck records adverse media check metrics
func (pm *PrometheusMetrics) RecordAdverseMediaCheck(severity, tenantID string) {
	pm.adverseMediaChecksTotal.WithLabelValues(severity, tenantID).Inc()
}

// RecordExternalAPICall records external API call metrics
func (pm *PrometheusMetrics) RecordExternalAPICall(provider, endpoint, status, tenantID string, duration time.Duration) {
	pm.externalAPICalls.WithLabelValues(provider, endpoint, status, tenantID).Inc()
	pm.externalAPIDuration.WithLabelValues(provider, endpoint, tenantID).Observe(duration.Seconds())
}

// RecordError records error metrics
func (pm *PrometheusMetrics) RecordError(errorType, component, tenantID string) {
	pm.errorsTotal.WithLabelValues(errorType, component, tenantID).Inc()
}

// RecordTimeoutError records timeout error metrics
func (pm *PrometheusMetrics) RecordTimeoutError(component, tenantID string) {
	pm.timeoutErrors.WithLabelValues(component, tenantID).Inc()
}

// RecordValidationError records validation error metrics
func (pm *PrometheusMetrics) RecordValidationError(validationType, tenantID string) {
	pm.validationErrors.WithLabelValues(validationType, tenantID).Inc()
}

// UpdatePerformanceMetrics updates performance metrics
func (pm *PrometheusMetrics) UpdatePerformanceMetrics(endpoint, tenantID string, p50, p95, p99, throughput float64) {
	pm.responseTimeP50.WithLabelValues(endpoint, tenantID).Set(p50)
	pm.responseTimeP95.WithLabelValues(endpoint, tenantID).Set(p95)
	pm.responseTimeP99.WithLabelValues(endpoint, tenantID).Set(p99)
	pm.throughput.WithLabelValues(endpoint, tenantID).Set(throughput)
}

// UpdateSystemMetrics updates system metrics
func (pm *PrometheusMetrics) UpdateSystemMetrics(activeConnections int, dbConnections map[string]int, cacheHitRates map[string]float64) {
	pm.activeConnections.Set(float64(activeConnections))

	for state, count := range dbConnections {
		pm.databaseConnections.WithLabelValues(state, "postgres").Set(float64(count))
	}

	for cacheType, hitRate := range cacheHitRates {
		pm.cacheHitRate.WithLabelValues(cacheType).Set(hitRate)
	}
}

// UpdateTenantMetrics updates tenant-specific metrics
func (pm *PrometheusMetrics) UpdateTenantMetrics(tenantID string, dataUsage map[string]int64, errorRate float64) {
	for dataType, usage := range dataUsage {
		pm.tenantDataUsage.WithLabelValues(tenantID, dataType).Set(float64(usage))
	}
	pm.tenantErrorRate.WithLabelValues(tenantID).Set(errorRate)
}

// RecordAuditEvent records audit event metrics
func (pm *PrometheusMetrics) RecordAuditEvent(eventType, tenantID, userID string) {
	pm.auditEventsTotal.WithLabelValues(eventType, tenantID, userID).Inc()
}

// RecordComplianceViolation records compliance violation metrics
func (pm *PrometheusMetrics) RecordComplianceViolation(violationType, regulation, tenantID string) {
	pm.complianceViolations.WithLabelValues(violationType, regulation, tenantID).Inc()
}

// RecordSecurityIncident records security incident metrics
func (pm *PrometheusMetrics) RecordSecurityIncident(incidentType, severity, tenantID string) {
	pm.securityIncidents.WithLabelValues(incidentType, severity, tenantID).Inc()
}

// StartMetricsServer starts the Prometheus metrics server
func (pm *PrometheusMetrics) StartMetricsServer(ctx context.Context, port int) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	pm.logger.Info("Starting Prometheus metrics server", zap.Int("port", port))

	go func() {
		<-ctx.Done()
		pm.logger.Info("Shutting down Prometheus metrics server")
		server.Shutdown(context.Background())
	}()

	return server.ListenAndServe()
}

// GetMetricsHandler returns the Prometheus metrics handler
func (pm *PrometheusMetrics) GetMetricsHandler() http.Handler {
	return promhttp.Handler()
}
