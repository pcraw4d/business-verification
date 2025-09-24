package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/test/mocks"
)

// TestMonitoringEndpoints tests all monitoring API endpoints
func TestMonitoringEndpoints(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// Setup test environment
	suite := setupMonitoringEndpointTestSuite(t)
	defer suite.cleanup()

	// Test health and status endpoints
	t.Run("HealthAndStatusEndpoints", suite.testHealthAndStatusEndpoints)

	// Test metrics and analytics endpoints
	t.Run("MetricsAndAnalyticsEndpoints", suite.testMetricsAndAnalyticsEndpoints)

	// Test monitoring and alerting endpoints
	t.Run("MonitoringAndAlertingEndpoints", suite.testMonitoringAndAlertingEndpoints)

	// Test compliance monitoring endpoints
	t.Run("ComplianceMonitoringEndpoints", suite.testComplianceMonitoringEndpoints)
}

// MonitoringEndpointTestSuite provides testing for monitoring endpoints
type MonitoringEndpointTestSuite struct {
	server  *httptest.Server
	mux     *http.ServeMux
	logger  *observability.Logger
	cleanup func()
}

// setupMonitoringEndpointTestSuite sets up the monitoring endpoint test suite
func setupMonitoringEndpointTestSuite(t *testing.T) *MonitoringEndpointTestSuite {
	// Setup logger
	logger := observability.NewLogger(&observability.Config{
		LogLevel:  "debug",
		LogFormat: "json",
	})

	// Create main mux
	mux := http.NewServeMux()

	// Setup mock services
	mockMonitoringService := mocks.NewMockMonitoringService()
	mockAnalyticsService := mocks.NewMockAnalyticsService()
	mockAlertingService := mocks.NewMockAlertingService()
	mockComplianceService := mocks.NewMockComplianceService()

	// Setup handlers
	monitoringHandler := handlers.NewMonitoringHandler(mockMonitoringService, logger)
	analyticsHandler := handlers.NewAnalyticsHandler(mockAnalyticsService, logger)
	alertingHandler := handlers.NewAlertingHandler(mockAlertingService, logger)
	complianceHandler := handlers.NewComplianceHandler(mockComplianceService, logger)

	// Register monitoring routes
	registerMonitoringRoutes(mux, monitoringHandler, analyticsHandler, alertingHandler, complianceHandler)

	// Create test server
	server := httptest.NewServer(mux)

	return &MonitoringEndpointTestSuite{
		server:  server,
		mux:     mux,
		logger:  logger,
		cleanup: func() { server.Close() },
	}
}

// registerMonitoringRoutes registers all monitoring API routes
func registerMonitoringRoutes(
	mux *http.ServeMux,
	monitoringHandler *handlers.MonitoringHandler,
	analyticsHandler *handlers.AnalyticsHandler,
	alertingHandler *handlers.AlertingHandler,
	complianceHandler *handlers.ComplianceHandler,
) {
	// Health and status endpoints
	mux.HandleFunc("GET /health", monitoringHandler.HealthCheck)
	mux.HandleFunc("GET /v1/status", monitoringHandler.GetStatus)
	mux.HandleFunc("GET /v1/status/detailed", monitoringHandler.GetDetailedStatus)
	mux.HandleFunc("GET /v1/status/components", monitoringHandler.GetComponentStatus)

	// Metrics endpoints
	mux.HandleFunc("GET /v1/metrics", monitoringHandler.GetMetrics)
	mux.HandleFunc("GET /v1/metrics/system", monitoringHandler.GetSystemMetrics)
	mux.HandleFunc("GET /v1/metrics/application", monitoringHandler.GetApplicationMetrics)
	mux.HandleFunc("GET /v1/metrics/database", monitoringHandler.GetDatabaseMetrics)

	// Analytics endpoints
	mux.HandleFunc("GET /v1/analytics/classification", analyticsHandler.GetClassificationAnalytics)
	mux.HandleFunc("GET /v1/analytics/performance", analyticsHandler.GetPerformanceAnalytics)
	mux.HandleFunc("GET /v1/analytics/usage", analyticsHandler.GetUsageAnalytics)
	mux.HandleFunc("GET /v1/analytics/trends", analyticsHandler.GetTrendAnalytics)

	// Monitoring and alerting endpoints
	mux.HandleFunc("GET /v1/monitoring/alerts", alertingHandler.GetActiveAlerts)
	mux.HandleFunc("GET /v1/monitoring/alerts/history", alertingHandler.GetAlertHistory)
	mux.HandleFunc("POST /v1/monitoring/alerts/{alert_id}/acknowledge", alertingHandler.AcknowledgeAlert)
	mux.HandleFunc("POST /v1/monitoring/alerts/{alert_id}/resolve", alertingHandler.ResolveAlert)
	mux.HandleFunc("GET /v1/monitoring/dashboards", monitoringHandler.GetDashboards)

	// Compliance monitoring endpoints
	mux.HandleFunc("POST /v1/compliance/check", complianceHandler.CheckCompliance)
	mux.HandleFunc("GET /v1/compliance/status/{business_id}", complianceHandler.GetComplianceStatus)
	mux.HandleFunc("GET /v1/compliance/reports", complianceHandler.GetComplianceReports)
	mux.HandleFunc("GET /v1/compliance/frameworks", complianceHandler.GetComplianceFrameworks)
}

// testHealthAndStatusEndpoints tests health and status endpoints
func (suite *MonitoringEndpointTestSuite) testHealthAndStatusEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Basic health check
		{
			name:           "GET /health - Basic health check",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			description:    "Should return system health status",
		},
		{
			name:           "GET /health - Health check with details",
			method:         "GET",
			path:           "/health?detailed=true",
			expectedStatus: http.StatusOK,
			description:    "Should return detailed health status",
		},

		// System status
		{
			name:           "GET /v1/status - System status",
			method:         "GET",
			path:           "/v1/status",
			expectedStatus: http.StatusOK,
			description:    "Should return detailed system status",
		},
		{
			name:           "GET /v1/status - System status with filters",
			method:         "GET",
			path:           "/v1/status?include_metrics=true&include_components=true",
			expectedStatus: http.StatusOK,
			description:    "Should return system status with metrics and components",
		},

		// Detailed status
		{
			name:           "GET /v1/status/detailed - Detailed system status",
			method:         "GET",
			path:           "/v1/status/detailed",
			expectedStatus: http.StatusOK,
			description:    "Should return comprehensive system status",
		},
		{
			name:           "GET /v1/status/detailed - Detailed status with time range",
			method:         "GET",
			path:           "/v1/status/detailed?period=1h",
			expectedStatus: http.StatusOK,
			description:    "Should return detailed status for specific time period",
		},

		// Component status
		{
			name:           "GET /v1/status/components - Component status",
			method:         "GET",
			path:           "/v1/status/components",
			expectedStatus: http.StatusOK,
			description:    "Should return individual component status",
		},
		{
			name:           "GET /v1/status/components - Specific component status",
			method:         "GET",
			path:           "/v1/status/components?component=database",
			expectedStatus: http.StatusOK,
			description:    "Should return status for specific component",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testMetricsAndAnalyticsEndpoints tests metrics and analytics endpoints
func (suite *MonitoringEndpointTestSuite) testMetricsAndAnalyticsEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// System metrics
		{
			name:           "GET /v1/metrics - System metrics",
			method:         "GET",
			path:           "/v1/metrics",
			expectedStatus: http.StatusOK,
			description:    "Should return system performance metrics",
		},
		{
			name:           "GET /v1/metrics - System metrics with filters",
			method:         "GET",
			path:           "/v1/metrics?period=1h&include_details=true",
			expectedStatus: http.StatusOK,
			description:    "Should return filtered system metrics",
		},

		// Application metrics
		{
			name:           "GET /v1/metrics/application - Application metrics",
			method:         "GET",
			path:           "/v1/metrics/application",
			expectedStatus: http.StatusOK,
			description:    "Should return application performance metrics",
		},
		{
			name:           "GET /v1/metrics/application - Application metrics by service",
			method:         "GET",
			path:           "/v1/metrics/application?service=classification",
			expectedStatus: http.StatusOK,
			description:    "Should return metrics for specific service",
		},

		// Database metrics
		{
			name:           "GET /v1/metrics/database - Database metrics",
			method:         "GET",
			path:           "/v1/metrics/database",
			expectedStatus: http.StatusOK,
			description:    "Should return database performance metrics",
		},
		{
			name:           "GET /v1/metrics/database - Database metrics with details",
			method:         "GET",
			path:           "/v1/metrics/database?include_connections=true&include_queries=true",
			expectedStatus: http.StatusOK,
			description:    "Should return detailed database metrics",
		},

		// Classification analytics
		{
			name:           "GET /v1/analytics/classification - Classification analytics",
			method:         "GET",
			path:           "/v1/analytics/classification",
			expectedStatus: http.StatusOK,
			description:    "Should return classification analytics",
		},
		{
			name:           "GET /v1/analytics/classification - Classification analytics with filters",
			method:         "GET",
			path:           "/v1/analytics/classification?period=7d&model=bert",
			expectedStatus: http.StatusOK,
			description:    "Should return filtered classification analytics",
		},

		// Performance analytics
		{
			name:           "GET /v1/analytics/performance - Performance analytics",
			method:         "GET",
			path:           "/v1/analytics/performance",
			expectedStatus: http.StatusOK,
			description:    "Should return performance analytics",
		},
		{
			name:           "GET /v1/analytics/performance - Performance analytics by endpoint",
			method:         "GET",
			path:           "/v1/analytics/performance?endpoint=/v1/classify",
			expectedStatus: http.StatusOK,
			description:    "Should return performance analytics for specific endpoint",
		},

		// Usage analytics
		{
			name:           "GET /v1/analytics/usage - Usage analytics",
			method:         "GET",
			path:           "/v1/analytics/usage",
			expectedStatus: http.StatusOK,
			description:    "Should return usage analytics",
		},
		{
			name:           "GET /v1/analytics/usage - Usage analytics by user",
			method:         "GET",
			path:           "/v1/analytics/usage?user_id=test-user-123",
			expectedStatus: http.StatusOK,
			description:    "Should return usage analytics for specific user",
		},

		// Trend analytics
		{
			name:           "GET /v1/analytics/trends - Trend analytics",
			method:         "GET",
			path:           "/v1/analytics/trends",
			expectedStatus: http.StatusOK,
			description:    "Should return trend analytics",
		},
		{
			name:           "GET /v1/analytics/trends - Trend analytics with time range",
			method:         "GET",
			path:           "/v1/analytics/trends?period=30d&metric=accuracy",
			expectedStatus: http.StatusOK,
			description:    "Should return trend analytics for specific metric",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testMonitoringAndAlertingEndpoints tests monitoring and alerting endpoints
func (suite *MonitoringEndpointTestSuite) testMonitoringAndAlertingEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Active alerts
		{
			name:           "GET /v1/monitoring/alerts - Get active alerts",
			method:         "GET",
			path:           "/v1/monitoring/alerts",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve active monitoring alerts",
		},
		{
			name:           "GET /v1/monitoring/alerts - Get alerts with filters",
			method:         "GET",
			path:           "/v1/monitoring/alerts?severity=high&status=active",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve filtered active alerts",
		},

		// Alert history
		{
			name:           "GET /v1/monitoring/alerts/history - Get alert history",
			method:         "GET",
			path:           "/v1/monitoring/alerts/history",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve alert history",
		},
		{
			name:           "GET /v1/monitoring/alerts/history - Get alert history with filters",
			method:         "GET",
			path:           "/v1/monitoring/alerts/history?period=7d&severity=critical",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve filtered alert history",
		},

		// Alert acknowledgment
		{
			name:   "POST /v1/monitoring/alerts/{alert_id}/acknowledge - Acknowledge alert",
			method: "POST",
			path:   "/v1/monitoring/alerts/test-alert-123/acknowledge",
			body: map[string]interface{}{
				"acknowledged_by": "admin-user-456",
				"notes":           "Investigating the issue",
			},
			expectedStatus: http.StatusOK,
			description:    "Should acknowledge alert successfully",
		},

		// Alert resolution
		{
			name:   "POST /v1/monitoring/alerts/{alert_id}/resolve - Resolve alert",
			method: "POST",
			path:   "/v1/monitoring/alerts/test-alert-123/resolve",
			body: map[string]interface{}{
				"resolved_by": "admin-user-456",
				"resolution":  "Issue fixed by restarting service",
				"notes":       "Service restarted and monitoring shows normal operation",
			},
			expectedStatus: http.StatusOK,
			description:    "Should resolve alert successfully",
		},

		// Monitoring dashboards
		{
			name:           "GET /v1/monitoring/dashboards - Get monitoring dashboards",
			method:         "GET",
			path:           "/v1/monitoring/dashboards",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve monitoring dashboards",
		},
		{
			name:           "GET /v1/monitoring/dashboards - Get specific dashboard",
			method:         "GET",
			path:           "/v1/monitoring/dashboards?dashboard=system_overview",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve specific monitoring dashboard",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testComplianceMonitoringEndpoints tests compliance monitoring endpoints
func (suite *MonitoringEndpointTestSuite) testComplianceMonitoringEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Compliance checking
		{
			name:   "POST /v1/compliance/check - Check compliance",
			method: "POST",
			path:   "/v1/compliance/check",
			body: map[string]interface{}{
				"business_id": "test-business-123",
				"frameworks":  []string{"SOC2", "PCI_DSS"},
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform compliance check",
		},
		{
			name:   "POST /v1/compliance/check - Comprehensive compliance check",
			method: "POST",
			path:   "/v1/compliance/check",
			body: map[string]interface{}{
				"business_id":             "test-business-123",
				"frameworks":              []string{"SOC2", "PCI_DSS", "GDPR", "HIPAA"},
				"include_details":         true,
				"include_recommendations": true,
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform comprehensive compliance check",
		},

		// Compliance status
		{
			name:           "GET /v1/compliance/status/{business_id} - Get compliance status",
			method:         "GET",
			path:           "/v1/compliance/status/test-business-123",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve compliance status",
		},
		{
			name:           "GET /v1/compliance/status/{business_id} - Get compliance status with details",
			method:         "GET",
			path:           "/v1/compliance/status/test-business-123?include_details=true&include_history=true",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve detailed compliance status",
		},

		// Compliance reports
		{
			name:           "GET /v1/compliance/reports - Get compliance reports",
			method:         "GET",
			path:           "/v1/compliance/reports",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve compliance reports",
		},
		{
			name:           "GET /v1/compliance/reports - Get compliance reports with filters",
			method:         "GET",
			path:           "/v1/compliance/reports?framework=SOC2&period=30d",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve filtered compliance reports",
		},

		// Compliance frameworks
		{
			name:           "GET /v1/compliance/frameworks - Get compliance frameworks",
			method:         "GET",
			path:           "/v1/compliance/frameworks",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve available compliance frameworks",
		},
		{
			name:           "GET /v1/compliance/frameworks - Get specific framework",
			method:         "GET",
			path:           "/v1/compliance/frameworks?framework=SOC2",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve specific compliance framework",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testEndpoint is a helper function to test individual endpoints
func (suite *MonitoringEndpointTestSuite) testEndpoint(t *testing.T, method, path string, body interface{}, expectedStatus int, description string) {
	var reqBody []byte
	var err error

	// Prepare request body if provided
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	// Create request
	req, err := http.NewRequest(method, suite.server.URL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Validate response status
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d for %s %s: %s",
			expectedStatus, resp.StatusCode, method, path, description)
	}

	// Validate response headers
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
	}

	// Log successful test
	suite.logger.Info("Monitoring endpoint test completed", map[string]interface{}{
		"method":          method,
		"path":            path,
		"status":          resp.StatusCode,
		"expected_status": expectedStatus,
		"description":     description,
		"success":         resp.StatusCode == expectedStatus,
	})
}

// TestMonitoringEndpointPerformance tests monitoring endpoint performance
func TestMonitoringEndpointPerformance(t *testing.T) {
	// Skip if not running performance tests
	if os.Getenv("PERFORMANCE_TESTS") != "true" {
		t.Skip("Skipping performance tests - set PERFORMANCE_TESTS=true to run")
	}

	suite := setupMonitoringEndpointTestSuite(t)
	defer suite.cleanup()

	// Performance test cases
	performanceTests := []struct {
		name        string
		method      string
		path        string
		body        interface{}
		maxDuration time.Duration
		description string
	}{
		{
			name:        "Health check performance",
			method:      "GET",
			path:        "/health",
			maxDuration: 100 * time.Millisecond,
			description: "Health check should complete within 100ms",
		},
		{
			name:        "System status performance",
			method:      "GET",
			path:        "/v1/status",
			maxDuration: 500 * time.Millisecond,
			description: "System status should complete within 500ms",
		},
		{
			name:        "Metrics retrieval performance",
			method:      "GET",
			path:        "/v1/metrics",
			maxDuration: 1 * time.Second,
			description: "Metrics retrieval should complete within 1 second",
		},
		{
			name:        "Analytics performance",
			method:      "GET",
			path:        "/v1/analytics/classification",
			maxDuration: 2 * time.Second,
			description: "Analytics should complete within 2 seconds",
		},
		{
			name:        "Alert retrieval performance",
			method:      "GET",
			path:        "/v1/monitoring/alerts",
			maxDuration: 500 * time.Millisecond,
			description: "Alert retrieval should complete within 500ms",
		},
	}

	for _, pt := range performanceTests {
		t.Run(pt.name, func(t *testing.T) {
			start := time.Now()
			suite.testEndpoint(t, pt.method, pt.path, pt.body, http.StatusOK, pt.description)
			duration := time.Since(start)

			if duration > pt.maxDuration {
				t.Errorf("Performance test failed: %s took %v, expected < %v",
					pt.description, duration, pt.maxDuration)
			}

			suite.logger.Info("Performance test completed", map[string]interface{}{
				"test":         pt.name,
				"duration":     duration,
				"max_duration": pt.maxDuration,
				"passed":       duration <= pt.maxDuration,
			})
		})
	}
}
