package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// MetricsMiddleware provides HTTP middleware for collecting metrics
type MetricsMiddleware struct {
	metrics *PrometheusMetrics
	logger  *zap.Logger
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(metrics *PrometheusMetrics, logger *zap.Logger) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: metrics,
		logger:  logger,
	}
}

// HTTPMetricsMiddleware returns an HTTP middleware function for collecting metrics
func (mm *MetricsMiddleware) HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Extract tenant ID from request context or headers
		tenantID := mm.extractTenantID(r)
		
		// Create a response writer wrapper to capture response size and status
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200,
			size:          0,
		}
		
		// Process the request
		next.ServeHTTP(wrapped, r)
		
		// Calculate metrics
		duration := time.Since(start)
		statusCode := strconv.Itoa(wrapped.statusCode)
		
		// Record metrics
		mm.metrics.RecordHTTPRequest(
			r.Method,
			r.URL.Path,
			statusCode,
			tenantID,
			duration,
			r.ContentLength,
			wrapped.size,
		)
		
		// Log slow requests
		if duration > 5*time.Second {
			mm.logger.Warn("Slow request detected",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("tenant_id", tenantID),
				zap.Duration("duration", duration),
				zap.Int("status_code", wrapped.statusCode),
			)
		}
	})
}

// extractTenantID extracts tenant ID from request
func (mm *MetricsMiddleware) extractTenantID(r *http.Request) string {
	// Try to get tenant ID from various sources
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}
	
	if tenantID := r.Header.Get("X-API-Key"); tenantID != "" {
		// In a real implementation, you would validate the API key and extract tenant ID
		return "tenant_" + tenantID[:8] // Mock tenant ID from API key
	}
	
	// Try to extract from JWT token
	if auth := r.Header.Get("Authorization"); auth != "" {
		// In a real implementation, you would parse the JWT and extract tenant ID
		return "jwt_tenant" // Mock tenant ID from JWT
	}
	
	return "anonymous"
}

// responseWriter wraps http.ResponseWriter to capture response metrics
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += int64(size)
	return size, err
}

// BusinessMetricsCollector collects business-specific metrics
type BusinessMetricsCollector struct {
	metrics *PrometheusMetrics
	logger  *zap.Logger
}

// NewBusinessMetricsCollector creates a new business metrics collector
func NewBusinessMetricsCollector(metrics *PrometheusMetrics, logger *zap.Logger) *BusinessMetricsCollector {
	return &BusinessMetricsCollector{
		metrics: metrics,
		logger:  logger,
	}
}

// RecordRiskAssessment records risk assessment metrics
func (bmc *BusinessMetricsCollector) RecordRiskAssessment(assessmentType, riskLevel, tenantID, country string, duration time.Duration, score float64) {
	bmc.metrics.RecordRiskAssessment(assessmentType, riskLevel, tenantID, country, duration, score)
	
	bmc.logger.Info("Risk assessment recorded",
		zap.String("assessment_type", assessmentType),
		zap.String("risk_level", riskLevel),
		zap.String("tenant_id", tenantID),
		zap.String("country", country),
		zap.Duration("duration", duration),
		zap.Float64("score", score),
	)
}

// RecordComplianceCheck records compliance check metrics
func (bmc *BusinessMetricsCollector) RecordComplianceCheck(regulation, status, tenantID string) {
	bmc.metrics.RecordComplianceCheck(regulation, status, tenantID)
	
	bmc.logger.Info("Compliance check recorded",
		zap.String("regulation", regulation),
		zap.String("status", status),
		zap.String("tenant_id", tenantID),
	)
}

// RecordSanctionsScreening records sanctions screening metrics
func (bmc *BusinessMetricsCollector) RecordSanctionsScreening(listType, matchStatus, tenantID string) {
	bmc.metrics.RecordSanctionsScreening(listType, matchStatus, tenantID)
	
	bmc.logger.Info("Sanctions screening recorded",
		zap.String("list_type", listType),
		zap.String("match_status", matchStatus),
		zap.String("tenant_id", tenantID),
	)
}

// RecordAdverseMediaCheck records adverse media check metrics
func (bmc *BusinessMetricsCollector) RecordAdverseMediaCheck(severity, tenantID string) {
	bmc.metrics.RecordAdverseMediaCheck(severity, tenantID)
	
	bmc.logger.Info("Adverse media check recorded",
		zap.String("severity", severity),
		zap.String("tenant_id", tenantID),
	)
}

// RecordExternalAPICall records external API call metrics
func (bmc *BusinessMetricsCollector) RecordExternalAPICall(provider, endpoint, status, tenantID string, duration time.Duration) {
	bmc.metrics.RecordExternalAPICall(provider, endpoint, status, tenantID, duration)
	
	bmc.logger.Info("External API call recorded",
		zap.String("provider", provider),
		zap.String("endpoint", endpoint),
		zap.String("status", status),
		zap.String("tenant_id", tenantID),
		zap.Duration("duration", duration),
	)
}

// RecordError records error metrics
func (bmc *BusinessMetricsCollector) RecordError(errorType, component, tenantID string) {
	bmc.metrics.RecordError(errorType, component, tenantID)
	
	bmc.logger.Error("Error recorded",
		zap.String("error_type", errorType),
		zap.String("component", component),
		zap.String("tenant_id", tenantID),
	)
}

// RecordTimeoutError records timeout error metrics
func (bmc *BusinessMetricsCollector) RecordTimeoutError(component, tenantID string) {
	bmc.metrics.RecordTimeoutError(component, tenantID)
	
	bmc.logger.Error("Timeout error recorded",
		zap.String("component", component),
		zap.String("tenant_id", tenantID),
	)
}

// RecordValidationError records validation error metrics
func (bmc *BusinessMetricsCollector) RecordValidationError(validationType, tenantID string) {
	bmc.metrics.RecordValidationError(validationType, tenantID)
	
	bmc.logger.Warn("Validation error recorded",
		zap.String("validation_type", validationType),
		zap.String("tenant_id", tenantID),
	)
}

// UpdatePerformanceMetrics updates performance metrics
func (bmc *BusinessMetricsCollector) UpdatePerformanceMetrics(endpoint, tenantID string, p50, p95, p99, throughput float64) {
	bmc.metrics.UpdatePerformanceMetrics(endpoint, tenantID, p50, p95, p99, throughput)
	
	bmc.logger.Info("Performance metrics updated",
		zap.String("endpoint", endpoint),
		zap.String("tenant_id", tenantID),
		zap.Float64("p50", p50),
		zap.Float64("p95", p95),
		zap.Float64("p99", p99),
		zap.Float64("throughput", throughput),
	)
}

// UpdateSystemMetrics updates system metrics
func (bmc *BusinessMetricsCollector) UpdateSystemMetrics(activeConnections int, dbConnections map[string]int, cacheHitRates map[string]float64) {
	bmc.metrics.UpdateSystemMetrics(activeConnections, dbConnections, cacheHitRates)
	
	bmc.logger.Info("System metrics updated",
		zap.Int("active_connections", activeConnections),
		zap.Any("db_connections", dbConnections),
		zap.Any("cache_hit_rates", cacheHitRates),
	)
}

// UpdateTenantMetrics updates tenant-specific metrics
func (bmc *BusinessMetricsCollector) UpdateTenantMetrics(tenantID string, dataUsage map[string]int64, errorRate float64) {
	bmc.metrics.UpdateTenantMetrics(tenantID, dataUsage, errorRate)
	
	bmc.logger.Info("Tenant metrics updated",
		zap.String("tenant_id", tenantID),
		zap.Any("data_usage", dataUsage),
		zap.Float64("error_rate", errorRate),
	)
}

// RecordAuditEvent records audit event metrics
func (bmc *BusinessMetricsCollector) RecordAuditEvent(eventType, tenantID, userID string) {
	bmc.metrics.RecordAuditEvent(eventType, tenantID, userID)
	
	bmc.logger.Info("Audit event recorded",
		zap.String("event_type", eventType),
		zap.String("tenant_id", tenantID),
		zap.String("user_id", userID),
	)
}

// RecordComplianceViolation records compliance violation metrics
func (bmc *BusinessMetricsCollector) RecordComplianceViolation(violationType, regulation, tenantID string) {
	bmc.metrics.RecordComplianceViolation(violationType, regulation, tenantID)
	
	bmc.logger.Warn("Compliance violation recorded",
		zap.String("violation_type", violationType),
		zap.String("regulation", regulation),
		zap.String("tenant_id", tenantID),
	)
}

// RecordSecurityIncident records security incident metrics
func (bmc *BusinessMetricsCollector) RecordSecurityIncident(incidentType, severity, tenantID string) {
	bmc.metrics.RecordSecurityIncident(incidentType, severity, tenantID)
	
	bmc.logger.Error("Security incident recorded",
		zap.String("incident_type", incidentType),
		zap.String("severity", severity),
		zap.String("tenant_id", tenantID),
	)
}

// PerformanceAnalyzer analyzes performance metrics and provides insights
type PerformanceAnalyzer struct {
	metrics *PrometheusMetrics
	logger  *zap.Logger
}

// NewPerformanceAnalyzer creates a new performance analyzer
func NewPerformanceAnalyzer(metrics *PrometheusMetrics, logger *zap.Logger) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		metrics: metrics,
		logger:  logger,
	}
}

// AnalyzePerformance analyzes performance metrics and returns insights
func (pa *PerformanceAnalyzer) AnalyzePerformance(endpoint, tenantID string) *PerformanceInsights {
	// Mock implementation - in a real implementation, this would analyze actual metrics
	insights := &PerformanceInsights{
		Endpoint:        endpoint,
		TenantID:        tenantID,
		AverageResponseTime: 0.5,
		P95ResponseTime:     1.2,
		P99ResponseTime:     2.1,
		Throughput:          100.0,
		ErrorRate:           0.01,
		Recommendations: []string{
			"Response time is within acceptable limits",
			"Consider implementing caching for frequently accessed data",
			"Monitor error rate closely",
		},
		HealthScore: 85,
		LastUpdated: time.Now(),
	}
	
	pa.logger.Info("Performance analysis completed",
		zap.String("endpoint", endpoint),
		zap.String("tenant_id", tenantID),
		zap.Int("health_score", insights.HealthScore),
	)
	
	return insights
}

// PerformanceInsights represents performance analysis insights
type PerformanceInsights struct {
	Endpoint            string    `json:"endpoint"`
	TenantID            string    `json:"tenant_id"`
	AverageResponseTime float64   `json:"average_response_time"`
	P95ResponseTime     float64   `json:"p95_response_time"`
	P99ResponseTime     float64   `json:"p99_response_time"`
	Throughput          float64   `json:"throughput"`
	ErrorRate           float64   `json:"error_rate"`
	Recommendations     []string  `json:"recommendations"`
	HealthScore         int       `json:"health_score"`
	LastUpdated         time.Time `json:"last_updated"`
}
