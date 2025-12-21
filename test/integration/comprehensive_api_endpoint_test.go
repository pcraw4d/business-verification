//go:build !comprehensive_test && !e2e_railway
// +build !comprehensive_test,!e2e_railway

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/observability"
	"kyb-platform/test/mocks"
)

// ComprehensiveAPITestSuite provides comprehensive testing for all API endpoints
type ComprehensiveAPITestSuite struct {
	server  *httptest.Server
	mux     *http.ServeMux
	logger  *observability.Logger
	cleanup func()
}

// TestComprehensiveAPIEndpoints tests all API endpoints as specified in subtask 4.2.1
func TestComprehensiveAPIEndpoints(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// Setup test suite
	suite := setupComprehensiveAPITestSuite(t)
	defer suite.cleanup()

	// Test all endpoint categories
	t.Run("BusinessRelatedEndpoints", suite.testBusinessRelatedEndpoints)
	t.Run("ClassificationEndpoints", suite.testClassificationEndpoints)
	t.Run("UserManagementEndpoints", suite.testUserManagementEndpoints)
	t.Run("MonitoringEndpoints", suite.testMonitoringEndpoints)
}

// setupComprehensiveAPITestSuite sets up the comprehensive API test suite
func setupComprehensiveAPITestSuite(t *testing.T) *ComprehensiveAPITestSuite {
	// Setup logger
	logger := observability.NewLogger(&observability.Config{
		LogLevel:  "debug",
		LogFormat: "json",
	})

	// Create main mux
	mux := http.NewServeMux()

	// Setup mock services
	mockClassificationService := mocks.NewMockClassificationService()
	mockAuthService := mocks.NewMockAuthService()
	mockRiskService := mocks.NewMockRiskService()
	mockMonitoringService := mocks.NewMockMonitoringService()

	// Setup handlers
	classificationHandler := handlers.NewClassificationHandler(mockClassificationService, logger)
	authHandler := handlers.NewAuthHandler(mockAuthService, logger)
	riskHandler := handlers.NewRiskHandler(mockRiskService, logger)
	monitoringHandler := handlers.NewMonitoringHandler(mockMonitoringService, logger)

	// Register all routes with comprehensive coverage
	registerAllAPIRoutes(mux, classificationHandler, authHandler, riskHandler, monitoringHandler, logger)

	// Create test server
	server := httptest.NewServer(mux)

	return &ComprehensiveAPITestSuite{
		server:  server,
		mux:     mux,
		logger:  logger,
		cleanup: func() { server.Close() },
	}
}

// registerAllAPIRoutes registers all API routes for comprehensive testing
func registerAllAPIRoutes(
	mux *http.ServeMux,
	classificationHandler *handlers.ClassificationHandler,
	authHandler *handlers.AuthHandler,
	riskHandler *handlers.RiskHandler,
	monitoringHandler *handlers.MonitoringHandler,
	logger *observability.Logger,
) {
	// Business-related endpoints
	mux.HandleFunc("POST /v1/classify", classificationHandler.ClassifyBusiness)
	mux.HandleFunc("POST /v1/classify/batch", classificationHandler.ClassifyBusinesses)
	mux.HandleFunc("GET /v1/classify/{business_id}", classificationHandler.GetClassification)
	mux.HandleFunc("GET /v1/classify/history", classificationHandler.GetClassificationHistory)
	mux.HandleFunc("POST /v1/merchants", classificationHandler.CreateMerchant)
	mux.HandleFunc("GET /v1/merchants/{merchant_id}", classificationHandler.GetMerchant)
	mux.HandleFunc("PUT /v1/merchants/{merchant_id}", classificationHandler.UpdateMerchant)
	mux.HandleFunc("DELETE /v1/merchants/{merchant_id}", classificationHandler.DeleteMerchant)

	// Risk assessment endpoints
	mux.HandleFunc("POST /v1/risk/assess", riskHandler.AssessRisk)
	mux.HandleFunc("GET /v1/risk/{business_id}", riskHandler.GetRiskAssessment)
	mux.HandleFunc("GET /v1/risk/history/{business_id}", riskHandler.GetRiskHistory)
	mux.HandleFunc("POST /v1/risk/enhanced/assess", riskHandler.EnhancedRiskAssessment)
	mux.HandleFunc("GET /v1/risk/alerts", riskHandler.GetRiskAlerts)

	// User management endpoints
	mux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /v1/auth/refresh", authHandler.RefreshToken)
	mux.HandleFunc("GET /v1/users/profile", authHandler.GetProfile)
	mux.HandleFunc("PUT /v1/users/profile", authHandler.UpdateProfile)
	mux.HandleFunc("GET /v1/users/api-keys", authHandler.GetAPIKeys)
	mux.HandleFunc("POST /v1/users/api-keys", authHandler.CreateAPIKey)
	mux.HandleFunc("DELETE /v1/users/api-keys/{key_id}", authHandler.DeleteAPIKey)

	// Monitoring endpoints
	mux.HandleFunc("GET /health", monitoringHandler.HealthCheck)
	mux.HandleFunc("GET /v1/status", monitoringHandler.GetStatus)
	mux.HandleFunc("GET /v1/metrics", monitoringHandler.GetMetrics)
	mux.HandleFunc("GET /v1/analytics/classification", monitoringHandler.GetClassificationAnalytics)
	mux.HandleFunc("GET /v1/analytics/performance", monitoringHandler.GetPerformanceAnalytics)
	mux.HandleFunc("GET /v1/monitoring/accuracy/metrics", monitoringHandler.GetAccuracyMetrics)
	mux.HandleFunc("GET /v1/monitoring/alerts", monitoringHandler.GetActiveAlerts)

	// Compliance endpoints
	mux.HandleFunc("POST /v1/compliance/check", monitoringHandler.CheckCompliance)
	mux.HandleFunc("GET /v1/compliance/status/{business_id}", monitoringHandler.GetComplianceStatus)
	mux.HandleFunc("GET /v1/compliance/reports", monitoringHandler.GetComplianceReports)
}

// testBusinessRelatedEndpoints tests all business-related API endpoints
func (suite *ComprehensiveAPITestSuite) testBusinessRelatedEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Classification endpoints
		{
			name:   "POST /v1/classify - Single business classification",
			method: "POST",
			path:   "/v1/classify",
			body: map[string]interface{}{
				"business_name": "Test Business Corp",
				"address":       "123 Test St, Test City, TC 12345",
				"website":       "https://testbusiness.com",
			},
			expectedStatus: http.StatusOK,
			description:    "Should classify a single business successfully",
		},
		{
			name:   "POST /v1/classify/batch - Batch business classification",
			method: "POST",
			path:   "/v1/classify/batch",
			body: map[string]interface{}{
				"businesses": []map[string]interface{}{
					{
						"business_name": "Business A",
						"address":       "123 Main St",
					},
					{
						"business_name": "Business B",
						"address":       "456 Oak Ave",
					},
				},
			},
			expectedStatus: http.StatusOK,
			description:    "Should classify multiple businesses in batch",
		},
		{
			name:           "GET /v1/classify/{business_id} - Get classification result",
			method:         "GET",
			path:           "/v1/classify/test-business-123",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve classification result for a business",
		},
		{
			name:           "GET /v1/classify/history - Get classification history",
			method:         "GET",
			path:           "/v1/classify/history?business_id=test-business-123&limit=10",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve classification history",
		},
		// Merchant management endpoints
		{
			name:   "POST /v1/merchants - Create merchant",
			method: "POST",
			path:   "/v1/merchants",
			body: map[string]interface{}{
				"business_name": "New Merchant Corp",
				"address":       "789 Merchant St",
				"contact_email": "contact@merchant.com",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should create a new merchant successfully",
		},
		{
			name:           "GET /v1/merchants/{merchant_id} - Get merchant",
			method:         "GET",
			path:           "/v1/merchants/test-merchant-123",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve merchant information",
		},
		{
			name:   "PUT /v1/merchants/{merchant_id} - Update merchant",
			method: "PUT",
			path:   "/v1/merchants/test-merchant-123",
			body: map[string]interface{}{
				"business_name": "Updated Merchant Corp",
				"address":       "789 Updated St",
			},
			expectedStatus: http.StatusOK,
			description:    "Should update merchant information",
		},
		{
			name:           "DELETE /v1/merchants/{merchant_id} - Delete merchant",
			method:         "DELETE",
			path:           "/v1/merchants/test-merchant-123",
			expectedStatus: http.StatusNoContent,
			description:    "Should delete merchant successfully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testClassificationEndpoints tests all classification-specific API endpoints
func (suite *ComprehensiveAPITestSuite) testClassificationEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Enhanced classification endpoints
		{
			name:   "POST /v2/classify - Enhanced classification",
			method: "POST",
			path:   "/v2/classify",
			body: map[string]interface{}{
				"business_name": "Enhanced Test Business",
				"description":   "A technology company specializing in AI solutions",
				"website":       "https://enhancedtest.com",
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform enhanced classification with ML models",
		},
		{
			name:   "POST /v2/classify/batch - Enhanced batch classification",
			method: "POST",
			path:   "/v2/classify/batch",
			body: map[string]interface{}{
				"businesses": []map[string]interface{}{
					{
						"business_name": "Tech Company A",
						"description":   "Software development company",
					},
					{
						"business_name": "Retail Company B",
						"description":   "E-commerce retail business",
					},
				},
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform enhanced batch classification",
		},
		// Classification monitoring endpoints
		{
			name:           "GET /v1/monitoring/accuracy/metrics - Get accuracy metrics",
			method:         "GET",
			path:           "/v1/monitoring/accuracy/metrics",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve classification accuracy metrics",
		},
		{
			name:   "POST /v1/monitoring/accuracy/track - Track classification",
			method: "POST",
			path:   "/v1/monitoring/accuracy/track",
			body: map[string]interface{}{
				"business_id":     "test-business-123",
				"predicted_codes": []string{"541511", "541512"},
				"actual_codes":    []string{"541511"},
				"confidence":      0.95,
			},
			expectedStatus: http.StatusOK,
			description:    "Should track classification accuracy",
		},
		{
			name:           "GET /v1/monitoring/misclassifications - Get misclassifications",
			method:         "GET",
			path:           "/v1/monitoring/misclassifications",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve misclassification data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testUserManagementEndpoints tests all user management API endpoints
func (suite *ComprehensiveAPITestSuite) testUserManagementEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Authentication endpoints
		{
			name:   "POST /v1/auth/register - User registration",
			method: "POST",
			path:   "/v1/auth/register",
			body: map[string]interface{}{
				"email":     "test@example.com",
				"password":  "securepassword123",
				"full_name": "Test User",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should register a new user successfully",
		},
		{
			name:   "POST /v1/auth/login - User login",
			method: "POST",
			path:   "/v1/auth/login",
			body: map[string]interface{}{
				"email":    "test@example.com",
				"password": "securepassword123",
			},
			expectedStatus: http.StatusOK,
			description:    "Should authenticate user and return tokens",
		},
		{
			name:   "POST /v1/auth/refresh - Token refresh",
			method: "POST",
			path:   "/v1/auth/refresh",
			body: map[string]interface{}{
				"refresh_token": "valid-refresh-token",
			},
			expectedStatus: http.StatusOK,
			description:    "Should refresh access token",
		},
		{
			name:   "POST /v1/auth/logout - User logout",
			method: "POST",
			path:   "/v1/auth/logout",
			body: map[string]interface{}{
				"access_token": "valid-access-token",
			},
			expectedStatus: http.StatusOK,
			description:    "Should logout user and invalidate tokens",
		},
		// Profile management endpoints
		{
			name:           "GET /v1/users/profile - Get user profile",
			method:         "GET",
			path:           "/v1/users/profile",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve user profile information",
		},
		{
			name:   "PUT /v1/users/profile - Update user profile",
			method: "PUT",
			path:   "/v1/users/profile",
			body: map[string]interface{}{
				"full_name": "Updated Test User",
				"company":   "Test Company Inc",
			},
			expectedStatus: http.StatusOK,
			description:    "Should update user profile information",
		},
		// API key management endpoints
		{
			name:           "GET /v1/users/api-keys - Get API keys",
			method:         "GET",
			path:           "/v1/users/api-keys",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve user's API keys",
		},
		{
			name:   "POST /v1/users/api-keys - Create API key",
			method: "POST",
			path:   "/v1/users/api-keys",
			body: map[string]interface{}{
				"name":        "Test API Key",
				"description": "API key for testing purposes",
				"permissions": []string{"classify", "read"},
			},
			expectedStatus: http.StatusCreated,
			description:    "Should create a new API key",
		},
		{
			name:           "DELETE /v1/users/api-keys/{key_id} - Delete API key",
			method:         "DELETE",
			path:           "/v1/users/api-keys/test-key-123",
			expectedStatus: http.StatusNoContent,
			description:    "Should delete API key successfully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testMonitoringEndpoints tests all monitoring API endpoints
func (suite *ComprehensiveAPITestSuite) testMonitoringEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Health and status endpoints
		{
			name:           "GET /health - Health check",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			description:    "Should return system health status",
		},
		{
			name:           "GET /v1/status - Detailed status",
			method:         "GET",
			path:           "/v1/status",
			expectedStatus: http.StatusOK,
			description:    "Should return detailed system status",
		},
		// Metrics endpoints
		{
			name:           "GET /v1/metrics - System metrics",
			method:         "GET",
			path:           "/v1/metrics",
			expectedStatus: http.StatusOK,
			description:    "Should return system performance metrics",
		},
		{
			name:           "GET /v1/analytics/classification - Classification analytics",
			method:         "GET",
			path:           "/v1/analytics/classification",
			expectedStatus: http.StatusOK,
			description:    "Should return classification analytics",
		},
		{
			name:           "GET /v1/analytics/performance - Performance analytics",
			method:         "GET",
			path:           "/v1/analytics/performance",
			expectedStatus: http.StatusOK,
			description:    "Should return performance analytics",
		},
		// Monitoring and alerting endpoints
		{
			name:           "GET /v1/monitoring/alerts - Get active alerts",
			method:         "GET",
			path:           "/v1/monitoring/alerts",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve active monitoring alerts",
		},
		{
			name:           "GET /v1/monitoring/alerts/history - Get alert history",
			method:         "GET",
			path:           "/v1/monitoring/alerts/history",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve alert history",
		},
		// Compliance monitoring endpoints
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
			name:           "GET /v1/compliance/status/{business_id} - Get compliance status",
			method:         "GET",
			path:           "/v1/compliance/status/test-business-123",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve compliance status",
		},
		{
			name:           "GET /v1/compliance/reports - Get compliance reports",
			method:         "GET",
			path:           "/v1/compliance/reports",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve compliance reports",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testEndpoint is a helper function to test individual endpoints
func (suite *ComprehensiveAPITestSuite) testEndpoint(t *testing.T, method, path string, body interface{}, expectedStatus int, description string) {
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

	// Add authentication header for protected endpoints
	if strings.Contains(path, "/users/") || strings.Contains(path, "/auth/") {
		req.Header.Set("Authorization", "Bearer test-token")
	}

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
	suite.logger.Info("API endpoint test completed", map[string]interface{}{
		"method":          method,
		"path":            path,
		"status":          resp.StatusCode,
		"expected_status": expectedStatus,
		"description":     description,
		"success":         resp.StatusCode == expectedStatus,
	})
}

// TestAPIEndpointPerformance tests API endpoint performance
func TestAPIEndpointPerformance(t *testing.T) {
	// Skip if not running performance tests
	if os.Getenv("PERFORMANCE_TESTS") != "true" {
		t.Skip("Skipping performance tests - set PERFORMANCE_TESTS=true to run")
	}

	suite := setupComprehensiveAPITestSuite(t)
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
			name:        "Classification endpoint performance",
			method:      "POST",
			path:        "/v1/classify",
			body:        map[string]interface{}{"business_name": "Performance Test Business"},
			maxDuration: 2 * time.Second,
			description: "Classification should complete within 2 seconds",
		},
		{
			name:        "Health check performance",
			method:      "GET",
			path:        "/health",
			maxDuration: 100 * time.Millisecond,
			description: "Health check should complete within 100ms",
		},
		{
			name:        "Metrics endpoint performance",
			method:      "GET",
			path:        "/v1/metrics",
			maxDuration: 500 * time.Millisecond,
			description: "Metrics should be retrieved within 500ms",
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

// TestAPIEndpointErrorHandling tests API endpoint error handling
func TestAPIEndpointErrorHandling(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	suite := setupComprehensiveAPITestSuite(t)
	defer suite.cleanup()

	// Error handling test cases
	errorTests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		{
			name:           "Invalid JSON body",
			method:         "POST",
			path:           "/v1/classify",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 for invalid JSON",
		},
		{
			name:           "Missing required fields",
			method:         "POST",
			path:           "/v1/classify",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			description:    "Should return 400 for missing required fields",
		},
		{
			name:           "Non-existent business ID",
			method:         "GET",
			path:           "/v1/classify/non-existent-id",
			expectedStatus: http.StatusNotFound,
			description:    "Should return 404 for non-existent business",
		},
		{
			name:           "Unauthorized access",
			method:         "GET",
			path:           "/v1/users/profile",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should return 401 for unauthorized access",
		},
	}

	for _, et := range errorTests {
		t.Run(et.name, func(t *testing.T) {
			suite.testEndpoint(t, et.method, et.path, et.body, et.expectedStatus, et.description)
		})
	}
}
