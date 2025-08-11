package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// DashboardSystem provides comprehensive performance dashboards
type DashboardSystem struct {
	logger         *zap.Logger
	monitoring     *MonitoringSystem
	logAggregation *LogAggregationSystem
	config         *DashboardConfig
}

// DashboardConfig holds configuration for dashboards
type DashboardConfig struct {
	// Dashboard settings
	RefreshInterval time.Duration
	RetentionPeriod time.Duration
	MaxDataPoints   int

	// Performance thresholds
	ResponseTimeThresholds struct {
		Warning  time.Duration
		Critical time.Duration
	}
	ErrorRateThresholds struct {
		Warning  float64
		Critical float64
	}
	ThroughputThresholds struct {
		Warning  int
		Critical int
	}

	// Dashboard categories
	EnableSystemDashboard         bool
	EnablePerformanceDashboard    bool
	EnableBusinessDashboard       bool
	EnableSecurityDashboard       bool
	EnableInfrastructureDashboard bool

	// Export settings
	EnablePrometheusExport bool
	EnableJSONExport       bool
	EnableCSVExport        bool
}

// DashboardMetrics represents dashboard metrics
type DashboardMetrics struct {
	Timestamp time.Time `json:"timestamp"`

	// System metrics
	System struct {
		CPUUsage    float64 `json:"cpu_usage"`
		MemoryUsage float64 `json:"memory_usage"`
		Goroutines  int     `json:"goroutines"`
		HeapAlloc   uint64  `json:"heap_alloc"`
		HeapSys     uint64  `json:"heap_sys"`
		Uptime      string  `json:"uptime"`
	} `json:"system"`

	// Performance metrics
	Performance struct {
		RequestRate       float64 `json:"request_rate"`
		ResponseTimeP50   float64 `json:"response_time_p50"`
		ResponseTimeP95   float64 `json:"response_time_p95"`
		ResponseTimeP99   float64 `json:"response_time_p99"`
		ErrorRate         float64 `json:"error_rate"`
		Throughput        int     `json:"throughput"`
		ActiveConnections int     `json:"active_connections"`
	} `json:"performance"`

	// Business metrics
	Business struct {
		ClassificationRequests int     `json:"classification_requests"`
		ClassificationAccuracy float64 `json:"classification_accuracy"`
		RiskAssessments        int     `json:"risk_assessments"`
		ComplianceChecks       int     `json:"compliance_checks"`
		ActiveUsers            int     `json:"active_users"`
		APIKeyUsage            int     `json:"api_key_usage"`
	} `json:"business"`

	// Security metrics
	Security struct {
		AuthenticationAttempts int `json:"authentication_attempts"`
		AuthenticationFailures int `json:"authentication_failures"`
		RateLimitHits          int `json:"rate_limit_hits"`
		SecurityIncidents      int `json:"security_incidents"`
		APIKeyAbuse            int `json:"api_key_abuse"`
	} `json:"security"`

	// Infrastructure metrics
	Infrastructure struct {
		DatabaseConnections int     `json:"database_connections"`
		DatabaseQueryTime   float64 `json:"database_query_time"`
		DatabaseErrors      int     `json:"database_errors"`
		ExternalAPICalls    int     `json:"external_api_calls"`
		ExternalAPIErrors   int     `json:"external_api_errors"`
		ExternalAPILatency  float64 `json:"external_api_latency"`
		RedisConnections    int     `json:"redis_connections"`
		RedisMemoryUsage    float64 `json:"redis_memory_usage"`
	} `json:"infrastructure"`

	// Alerts
	Alerts []DashboardAlert `json:"alerts"`
}

// DashboardAlert represents a dashboard alert
type DashboardAlert struct {
	ID         string     `json:"id"`
	Severity   string     `json:"severity"`
	Category   string     `json:"category"`
	Message    string     `json:"message"`
	Value      float64    `json:"value"`
	Threshold  float64    `json:"threshold"`
	Timestamp  time.Time  `json:"timestamp"`
	Resolved   bool       `json:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// DashboardWidget represents a dashboard widget
type DashboardWidget struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Position    WidgetPosition         `json:"position"`
	Size        WidgetSize             `json:"size"`
	Config      map[string]interface{} `json:"config"`
	Data        interface{}            `json:"data"`
}

// WidgetPosition represents widget position
type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WidgetSize represents widget size
type WidgetSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// NewDashboardSystem creates a new dashboard system
func NewDashboardSystem(monitoring *MonitoringSystem, logAggregation *LogAggregationSystem, config *DashboardConfig, logger *zap.Logger) *DashboardSystem {
	return &DashboardSystem{
		logger:         logger,
		monitoring:     monitoring,
		logAggregation: logAggregation,
		config:         config,
	}
}

// GetDashboardMetrics retrieves comprehensive dashboard metrics
func (ds *DashboardSystem) GetDashboardMetrics(ctx context.Context) (*DashboardMetrics, error) {
	metrics := &DashboardMetrics{
		Timestamp: time.Now().UTC(),
	}

	// Collect system metrics
	if err := ds.collectSystemMetrics(ctx, metrics); err != nil {
		return nil, fmt.Errorf("failed to collect system metrics: %w", err)
	}

	// Collect performance metrics
	if err := ds.collectPerformanceMetrics(ctx, metrics); err != nil {
		return nil, fmt.Errorf("failed to collect performance metrics: %w", err)
	}

	// Collect business metrics
	if err := ds.collectBusinessMetrics(ctx, metrics); err != nil {
		return nil, fmt.Errorf("failed to collect business metrics: %w", err)
	}

	// Collect security metrics
	if err := ds.collectSecurityMetrics(ctx, metrics); err != nil {
		return nil, fmt.Errorf("failed to collect security metrics: %w", err)
	}

	// Collect infrastructure metrics
	if err := ds.collectInfrastructureMetrics(ctx, metrics); err != nil {
		return nil, fmt.Errorf("failed to collect infrastructure metrics: %w", err)
	}

	// Generate alerts
	metrics.Alerts = ds.generateAlerts(metrics)

	return metrics, nil
}

// collectSystemMetrics collects system-level metrics
func (ds *DashboardSystem) collectSystemMetrics(ctx context.Context, metrics *DashboardMetrics) error {
	// Get system metrics from monitoring system
	systemMetrics := ds.monitoring.GetMetricsSummary("production")

	// Extract system metrics from the summary
	if cpuUsage, ok := systemMetrics["cpu_usage"].(float64); ok {
		metrics.System.CPUUsage = cpuUsage
	}
	if memoryUsage, ok := systemMetrics["memory_usage"].(float64); ok {
		metrics.System.MemoryUsage = memoryUsage
	}
	if goroutines, ok := systemMetrics["goroutines"].(int); ok {
		metrics.System.Goroutines = goroutines
	}
	if heapAlloc, ok := systemMetrics["heap_alloc"].(uint64); ok {
		metrics.System.HeapAlloc = heapAlloc
	}
	if heapSys, ok := systemMetrics["heap_sys"].(uint64); ok {
		metrics.System.HeapSys = heapSys
	}
	if uptime, ok := systemMetrics["uptime"].(string); ok {
		metrics.System.Uptime = uptime
	}

	return nil
}

// collectPerformanceMetrics collects performance-related metrics
func (ds *DashboardSystem) collectPerformanceMetrics(ctx context.Context, metrics *DashboardMetrics) error {
	// Get performance metrics from monitoring system
	// These would be calculated from Prometheus metrics

	// Request rate (requests per second)
	metrics.Performance.RequestRate = ds.calculateRequestRate()

	// Response time percentiles
	metrics.Performance.ResponseTimeP50 = ds.calculateResponseTimePercentile(50)
	metrics.Performance.ResponseTimeP95 = ds.calculateResponseTimePercentile(95)
	metrics.Performance.ResponseTimeP99 = ds.calculateResponseTimePercentile(99)

	// Error rate
	metrics.Performance.ErrorRate = ds.calculateErrorRate()

	// Throughput
	metrics.Performance.Throughput = ds.calculateThroughput()

	// Active connections
	metrics.Performance.ActiveConnections = ds.calculateActiveConnections()

	return nil
}

// collectBusinessMetrics collects business-related metrics
func (ds *DashboardSystem) collectBusinessMetrics(ctx context.Context, metrics *DashboardMetrics) error {
	// Get business metrics from monitoring system
	// These would be calculated from business-specific metrics

	// Classification metrics
	metrics.Business.ClassificationRequests = ds.getClassificationRequests()
	metrics.Business.ClassificationAccuracy = ds.getClassificationAccuracy()

	// Risk assessment metrics
	metrics.Business.RiskAssessments = ds.getRiskAssessments()

	// Compliance check metrics
	metrics.Business.ComplianceChecks = ds.getComplianceChecks()

	// User activity metrics
	metrics.Business.ActiveUsers = ds.getActiveUsers()
	metrics.Business.APIKeyUsage = ds.getAPIKeyUsage()

	return nil
}

// collectSecurityMetrics collects security-related metrics
func (ds *DashboardSystem) collectSecurityMetrics(ctx context.Context, metrics *DashboardMetrics) error {
	// Get security metrics from monitoring system
	// These would be calculated from security-specific metrics

	// Authentication metrics
	metrics.Security.AuthenticationAttempts = ds.getAuthenticationAttempts()
	metrics.Security.AuthenticationFailures = ds.getAuthenticationFailures()

	// Rate limiting metrics
	metrics.Security.RateLimitHits = ds.getRateLimitHits()

	// Security incident metrics
	metrics.Security.SecurityIncidents = ds.getSecurityIncidents()
	metrics.Security.APIKeyAbuse = ds.getAPIKeyAbuse()

	return nil
}

// collectInfrastructureMetrics collects infrastructure-related metrics
func (ds *DashboardSystem) collectInfrastructureMetrics(ctx context.Context, metrics *DashboardMetrics) error {
	// Get infrastructure metrics from monitoring system
	// These would be calculated from infrastructure-specific metrics

	// Database metrics
	metrics.Infrastructure.DatabaseConnections = ds.getDatabaseConnections()
	metrics.Infrastructure.DatabaseQueryTime = ds.getDatabaseQueryTime()
	metrics.Infrastructure.DatabaseErrors = ds.getDatabaseErrors()

	// External API metrics
	metrics.Infrastructure.ExternalAPICalls = ds.getExternalAPICalls()
	metrics.Infrastructure.ExternalAPIErrors = ds.getExternalAPIErrors()
	metrics.Infrastructure.ExternalAPILatency = ds.getExternalAPILatency()

	// Redis metrics
	metrics.Infrastructure.RedisConnections = ds.getRedisConnections()
	metrics.Infrastructure.RedisMemoryUsage = ds.getRedisMemoryUsage()

	return nil
}

// generateAlerts generates alerts based on metrics and thresholds
func (ds *DashboardSystem) generateAlerts(metrics *DashboardMetrics) []DashboardAlert {
	var alerts []DashboardAlert

	// System alerts
	if metrics.System.CPUUsage > 80 {
		alerts = append(alerts, DashboardAlert{
			ID:        "system-cpu-high",
			Severity:  "warning",
			Category:  "system",
			Message:   "High CPU usage detected",
			Value:     metrics.System.CPUUsage,
			Threshold: 80,
			Timestamp: time.Now().UTC(),
		})
	}

	if metrics.System.MemoryUsage > 85 {
		alerts = append(alerts, DashboardAlert{
			ID:        "system-memory-high",
			Severity:  "critical",
			Category:  "system",
			Message:   "High memory usage detected",
			Value:     metrics.System.MemoryUsage,
			Threshold: 85,
			Timestamp: time.Now().UTC(),
		})
	}

	// Performance alerts
	if metrics.Performance.ResponseTimeP95 > ds.config.ResponseTimeThresholds.Warning.Seconds() {
		alerts = append(alerts, DashboardAlert{
			ID:        "performance-response-time-high",
			Severity:  "warning",
			Category:  "performance",
			Message:   "High response time detected",
			Value:     metrics.Performance.ResponseTimeP95,
			Threshold: ds.config.ResponseTimeThresholds.Warning.Seconds(),
			Timestamp: time.Now().UTC(),
		})
	}

	if metrics.Performance.ErrorRate > ds.config.ErrorRateThresholds.Warning {
		alerts = append(alerts, DashboardAlert{
			ID:        "performance-error-rate-high",
			Severity:  "critical",
			Category:  "performance",
			Message:   "High error rate detected",
			Value:     metrics.Performance.ErrorRate,
			Threshold: ds.config.ErrorRateThresholds.Warning,
			Timestamp: time.Now().UTC(),
		})
	}

	// Business alerts
	if metrics.Business.ClassificationAccuracy < 0.9 {
		alerts = append(alerts, DashboardAlert{
			ID:        "business-classification-accuracy-low",
			Severity:  "warning",
			Category:  "business",
			Message:   "Low classification accuracy detected",
			Value:     metrics.Business.ClassificationAccuracy,
			Threshold: 0.9,
			Timestamp: time.Now().UTC(),
		})
	}

	// Security alerts
	if metrics.Security.AuthenticationFailures > 10 {
		alerts = append(alerts, DashboardAlert{
			ID:        "security-auth-failures-high",
			Severity:  "warning",
			Category:  "security",
			Message:   "High authentication failure rate detected",
			Value:     float64(metrics.Security.AuthenticationFailures),
			Threshold: 10,
			Timestamp: time.Now().UTC(),
		})
	}

	if metrics.Security.RateLimitHits > 50 {
		alerts = append(alerts, DashboardAlert{
			ID:        "security-rate-limit-hits-high",
			Severity:  "critical",
			Category:  "security",
			Message:   "High rate limit hits detected",
			Value:     float64(metrics.Security.RateLimitHits),
			Threshold: 50,
			Timestamp: time.Now().UTC(),
		})
	}

	// Infrastructure alerts
	if metrics.Infrastructure.DatabaseErrors > 5 {
		alerts = append(alerts, DashboardAlert{
			ID:        "infrastructure-database-errors-high",
			Severity:  "critical",
			Category:  "infrastructure",
			Message:   "High database error rate detected",
			Value:     float64(metrics.Infrastructure.DatabaseErrors),
			Threshold: 5,
			Timestamp: time.Now().UTC(),
		})
	}

	if metrics.Infrastructure.ExternalAPIErrors > 10 {
		alerts = append(alerts, DashboardAlert{
			ID:        "infrastructure-external-api-errors-high",
			Severity:  "warning",
			Category:  "infrastructure",
			Message:   "High external API error rate detected",
			Value:     float64(metrics.Infrastructure.ExternalAPIErrors),
			Threshold: 10,
			Timestamp: time.Now().UTC(),
		})
	}

	return alerts
}

// GetSystemDashboard returns system dashboard widgets
func (ds *DashboardSystem) GetSystemDashboard(ctx context.Context) ([]DashboardWidget, error) {
	if !ds.config.EnableSystemDashboard {
		return nil, fmt.Errorf("system dashboard is disabled")
	}

	metrics, err := ds.GetDashboardMetrics(ctx)
	if err != nil {
		return nil, err
	}

	widgets := []DashboardWidget{
		{
			ID:          "cpu-usage",
			Type:        "gauge",
			Title:       "CPU Usage",
			Description: "Current CPU usage percentage",
			Position:    WidgetPosition{X: 0, Y: 0},
			Size:        WidgetSize{Width: 4, Height: 3},
			Config: map[string]interface{}{
				"min":  0,
				"max":  100,
				"unit": "%",
			},
			Data: metrics.System.CPUUsage,
		},
		{
			ID:          "memory-usage",
			Type:        "gauge",
			Title:       "Memory Usage",
			Description: "Current memory usage percentage",
			Position:    WidgetPosition{X: 4, Y: 0},
			Size:        WidgetSize{Width: 4, Height: 3},
			Config: map[string]interface{}{
				"min":  0,
				"max":  100,
				"unit": "%",
			},
			Data: metrics.System.MemoryUsage,
		},
		{
			ID:          "goroutines",
			Type:        "line",
			Title:       "Goroutines",
			Description: "Number of active goroutines over time",
			Position:    WidgetPosition{X: 0, Y: 3},
			Size:        WidgetSize{Width: 8, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
			},
			Data: metrics.System.Goroutines,
		},
		{
			ID:          "heap-usage",
			Type:        "line",
			Title:       "Heap Usage",
			Description: "Heap memory usage over time",
			Position:    WidgetPosition{X: 0, Y: 6},
			Size:        WidgetSize{Width: 8, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
			},
			Data: map[string]interface{}{
				"alloc": metrics.System.HeapAlloc,
				"sys":   metrics.System.HeapSys,
			},
		},
	}

	return widgets, nil
}

// GetPerformanceDashboard returns performance dashboard widgets
func (ds *DashboardSystem) GetPerformanceDashboard(ctx context.Context) ([]DashboardWidget, error) {
	if !ds.config.EnablePerformanceDashboard {
		return nil, fmt.Errorf("performance dashboard is disabled")
	}

	metrics, err := ds.GetDashboardMetrics(ctx)
	if err != nil {
		return nil, err
	}

	widgets := []DashboardWidget{
		{
			ID:          "request-rate",
			Type:        "line",
			Title:       "Request Rate",
			Description: "Requests per second over time",
			Position:    WidgetPosition{X: 0, Y: 0},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "req/s",
			},
			Data: metrics.Performance.RequestRate,
		},
		{
			ID:          "response-time",
			Type:        "line",
			Title:       "Response Time",
			Description: "Response time percentiles over time",
			Position:    WidgetPosition{X: 6, Y: 0},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "ms",
			},
			Data: map[string]interface{}{
				"p50": metrics.Performance.ResponseTimeP50,
				"p95": metrics.Performance.ResponseTimeP95,
				"p99": metrics.Performance.ResponseTimeP99,
			},
		},
		{
			ID:          "error-rate",
			Type:        "line",
			Title:       "Error Rate",
			Description: "Error rate over time",
			Position:    WidgetPosition{X: 0, Y: 3},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "%",
			},
			Data: metrics.Performance.ErrorRate,
		},
		{
			ID:          "throughput",
			Type:        "line",
			Title:       "Throughput",
			Description: "Total throughput over time",
			Position:    WidgetPosition{X: 6, Y: 3},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "req/min",
			},
			Data: metrics.Performance.Throughput,
		},
		{
			ID:          "active-connections",
			Type:        "gauge",
			Title:       "Active Connections",
			Description: "Number of active connections",
			Position:    WidgetPosition{X: 0, Y: 6},
			Size:        WidgetSize{Width: 4, Height: 3},
			Config: map[string]interface{}{
				"min": 0,
				"max": 1000,
			},
			Data: metrics.Performance.ActiveConnections,
		},
	}

	return widgets, nil
}

// GetBusinessDashboard returns business dashboard widgets
func (ds *DashboardSystem) GetBusinessDashboard(ctx context.Context) ([]DashboardWidget, error) {
	if !ds.config.EnableBusinessDashboard {
		return nil, fmt.Errorf("business dashboard is disabled")
	}

	metrics, err := ds.GetDashboardMetrics(ctx)
	if err != nil {
		return nil, err
	}

	widgets := []DashboardWidget{
		{
			ID:          "classification-requests",
			Type:        "line",
			Title:       "Classification Requests",
			Description: "Business classification requests over time",
			Position:    WidgetPosition{X: 0, Y: 0},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "requests",
			},
			Data: metrics.Business.ClassificationRequests,
		},
		{
			ID:          "classification-accuracy",
			Type:        "gauge",
			Title:       "Classification Accuracy",
			Description: "Business classification accuracy percentage",
			Position:    WidgetPosition{X: 6, Y: 0},
			Size:        WidgetSize{Width: 4, Height: 3},
			Config: map[string]interface{}{
				"min":  0,
				"max":  1,
				"unit": "%",
			},
			Data: metrics.Business.ClassificationAccuracy,
		},
		{
			ID:          "risk-assessments",
			Type:        "line",
			Title:       "Risk Assessments",
			Description: "Risk assessment requests over time",
			Position:    WidgetPosition{X: 0, Y: 3},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "assessments",
			},
			Data: metrics.Business.RiskAssessments,
		},
		{
			ID:          "compliance-checks",
			Type:        "line",
			Title:       "Compliance Checks",
			Description: "Compliance check requests over time",
			Position:    WidgetPosition{X: 6, Y: 3},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "checks",
			},
			Data: metrics.Business.ComplianceChecks,
		},
		{
			ID:          "active-users",
			Type:        "gauge",
			Title:       "Active Users",
			Description: "Number of active users",
			Position:    WidgetPosition{X: 0, Y: 6},
			Size:        WidgetSize{Width: 4, Height: 3},
			Config: map[string]interface{}{
				"min": 0,
				"max": 10000,
			},
			Data: metrics.Business.ActiveUsers,
		},
		{
			ID:          "api-key-usage",
			Type:        "line",
			Title:       "API Key Usage",
			Description: "API key usage over time",
			Position:    WidgetPosition{X: 4, Y: 6},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "requests",
			},
			Data: metrics.Business.APIKeyUsage,
		},
	}

	return widgets, nil
}

// GetSecurityDashboard returns security dashboard widgets
func (ds *DashboardSystem) GetSecurityDashboard(ctx context.Context) ([]DashboardWidget, error) {
	if !ds.config.EnableSecurityDashboard {
		return nil, fmt.Errorf("security dashboard is disabled")
	}

	metrics, err := ds.GetDashboardMetrics(ctx)
	if err != nil {
		return nil, err
	}

	widgets := []DashboardWidget{
		{
			ID:          "authentication-attempts",
			Type:        "line",
			Title:       "Authentication Attempts",
			Description: "Authentication attempts over time",
			Position:    WidgetPosition{X: 0, Y: 0},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "attempts",
			},
			Data: metrics.Security.AuthenticationAttempts,
		},
		{
			ID:          "authentication-failures",
			Type:        "line",
			Title:       "Authentication Failures",
			Description: "Authentication failures over time",
			Position:    WidgetPosition{X: 6, Y: 0},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "failures",
			},
			Data: metrics.Security.AuthenticationFailures,
		},
		{
			ID:          "rate-limit-hits",
			Type:        "line",
			Title:       "Rate Limit Hits",
			Description: "Rate limit violations over time",
			Position:    WidgetPosition{X: 0, Y: 3},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "hits",
			},
			Data: metrics.Security.RateLimitHits,
		},
		{
			ID:          "security-incidents",
			Type:        "line",
			Title:       "Security Incidents",
			Description: "Security incidents over time",
			Position:    WidgetPosition{X: 6, Y: 3},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "incidents",
			},
			Data: metrics.Security.SecurityIncidents,
		},
		{
			ID:          "api-key-abuse",
			Type:        "line",
			Title:       "API Key Abuse",
			Description: "API key abuse attempts over time",
			Position:    WidgetPosition{X: 0, Y: 6},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "attempts",
			},
			Data: metrics.Security.APIKeyAbuse,
		},
	}

	return widgets, nil
}

// GetInfrastructureDashboard returns infrastructure dashboard widgets
func (ds *DashboardSystem) GetInfrastructureDashboard(ctx context.Context) ([]DashboardWidget, error) {
	if !ds.config.EnableInfrastructureDashboard {
		return nil, fmt.Errorf("infrastructure dashboard is disabled")
	}

	metrics, err := ds.GetDashboardMetrics(ctx)
	if err != nil {
		return nil, err
	}

	widgets := []DashboardWidget{
		{
			ID:          "database-connections",
			Type:        "gauge",
			Title:       "Database Connections",
			Description: "Active database connections",
			Position:    WidgetPosition{X: 0, Y: 0},
			Size:        WidgetSize{Width: 4, Height: 3},
			Config: map[string]interface{}{
				"min": 0,
				"max": 100,
			},
			Data: metrics.Infrastructure.DatabaseConnections,
		},
		{
			ID:          "database-query-time",
			Type:        "line",
			Title:       "Database Query Time",
			Description: "Average database query time over time",
			Position:    WidgetPosition{X: 4, Y: 0},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "ms",
			},
			Data: metrics.Infrastructure.DatabaseQueryTime,
		},
		{
			ID:          "database-errors",
			Type:        "line",
			Title:       "Database Errors",
			Description: "Database errors over time",
			Position:    WidgetPosition{X: 0, Y: 3},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "errors",
			},
			Data: metrics.Infrastructure.DatabaseErrors,
		},
		{
			ID:          "external-api-calls",
			Type:        "line",
			Title:       "External API Calls",
			Description: "External API calls over time",
			Position:    WidgetPosition{X: 6, Y: 3},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "calls",
			},
			Data: metrics.Infrastructure.ExternalAPICalls,
		},
		{
			ID:          "external-api-errors",
			Type:        "line",
			Title:       "External API Errors",
			Description: "External API errors over time",
			Position:    WidgetPosition{X: 0, Y: 6},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "errors",
			},
			Data: metrics.Infrastructure.ExternalAPIErrors,
		},
		{
			ID:          "external-api-latency",
			Type:        "line",
			Title:       "External API Latency",
			Description: "External API latency over time",
			Position:    WidgetPosition{X: 6, Y: 6},
			Size:        WidgetSize{Width: 6, Height: 3},
			Config: map[string]interface{}{
				"timeRange": "1h",
				"unit":      "ms",
			},
			Data: metrics.Infrastructure.ExternalAPILatency,
		},
	}

	return widgets, nil
}

// DashboardHandler handles dashboard HTTP requests
func (ds *DashboardSystem) DashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")

		// Get dashboard type from query parameter
		dashboardType := r.URL.Query().Get("type")

		var data interface{}
		var err error

		switch dashboardType {
		case "metrics":
			data, err = ds.GetDashboardMetrics(ctx)
		case "system":
			data, err = ds.GetSystemDashboard(ctx)
		case "performance":
			data, err = ds.GetPerformanceDashboard(ctx)
		case "business":
			data, err = ds.GetBusinessDashboard(ctx)
		case "security":
			data, err = ds.GetSecurityDashboard(ctx)
		case "infrastructure":
			data, err = ds.GetInfrastructureDashboard(ctx)
		default:
			http.Error(w, "Invalid dashboard type", http.StatusBadRequest)
			return
		}

		if err != nil {
			ds.logger.Error("Failed to get dashboard data", zap.Error(err), zap.String("type", dashboardType))
			http.Error(w, "Failed to get dashboard data", http.StatusInternalServerError)
			return
		}

		// Encode response
		if err := json.NewEncoder(w).Encode(data); err != nil {
			ds.logger.Error("Failed to encode dashboard response", zap.Error(err))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// Metric calculation helper methods (these would integrate with Prometheus metrics)

func (ds *DashboardSystem) calculateRequestRate() float64 {
	// This would calculate from Prometheus metrics
	return 150.5 // Example value
}

func (ds *DashboardSystem) calculateResponseTimePercentile(percentile int) float64 {
	// This would calculate from Prometheus metrics
	switch percentile {
	case 50:
		return 45.2
	case 95:
		return 125.8
	case 99:
		return 245.3
	default:
		return 0
	}
}

func (ds *DashboardSystem) calculateErrorRate() float64 {
	// This would calculate from Prometheus metrics
	return 0.5 // Example value
}

func (ds *DashboardSystem) calculateThroughput() int {
	// This would calculate from Prometheus metrics
	return 9000 // Example value
}

func (ds *DashboardSystem) calculateActiveConnections() int {
	// This would calculate from Prometheus metrics
	return 45 // Example value
}

func (ds *DashboardSystem) getClassificationRequests() int {
	// This would get from business metrics
	return 1250 // Example value
}

func (ds *DashboardSystem) getClassificationAccuracy() float64 {
	// This would get from business metrics
	return 0.95 // Example value
}

func (ds *DashboardSystem) getRiskAssessments() int {
	// This would get from business metrics
	return 850 // Example value
}

func (ds *DashboardSystem) getComplianceChecks() int {
	// This would get from business metrics
	return 650 // Example value
}

func (ds *DashboardSystem) getActiveUsers() int {
	// This would get from business metrics
	return 125 // Example value
}

func (ds *DashboardSystem) getAPIKeyUsage() int {
	// This would get from business metrics
	return 3200 // Example value
}

func (ds *DashboardSystem) getAuthenticationAttempts() int {
	// This would get from security metrics
	return 450 // Example value
}

func (ds *DashboardSystem) getAuthenticationFailures() int {
	// This would get from security metrics
	return 12 // Example value
}

func (ds *DashboardSystem) getRateLimitHits() int {
	// This would get from security metrics
	return 8 // Example value
}

func (ds *DashboardSystem) getSecurityIncidents() int {
	// This would get from security metrics
	return 2 // Example value
}

func (ds *DashboardSystem) getAPIKeyAbuse() int {
	// This would get from security metrics
	return 3 // Example value
}

func (ds *DashboardSystem) getDatabaseConnections() int {
	// This would get from infrastructure metrics
	return 25 // Example value
}

func (ds *DashboardSystem) getDatabaseQueryTime() float64 {
	// This would get from infrastructure metrics
	return 15.3 // Example value
}

func (ds *DashboardSystem) getDatabaseErrors() int {
	// This would get from infrastructure metrics
	return 1 // Example value
}

func (ds *DashboardSystem) getExternalAPICalls() int {
	// This would get from infrastructure metrics
	return 1800 // Example value
}

func (ds *DashboardSystem) getExternalAPIErrors() int {
	// This would get from infrastructure metrics
	return 5 // Example value
}

func (ds *DashboardSystem) getExternalAPILatency() float64 {
	// This would get from infrastructure metrics
	return 85.2 // Example value
}

func (ds *DashboardSystem) getRedisConnections() int {
	// This would get from infrastructure metrics
	return 15 // Example value
}

func (ds *DashboardSystem) getRedisMemoryUsage() float64 {
	// This would get from infrastructure metrics
	return 45.8 // Example value
}
