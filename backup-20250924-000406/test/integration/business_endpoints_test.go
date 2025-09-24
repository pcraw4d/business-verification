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

// TestBusinessRelatedEndpoints tests all business-related API endpoints
func TestBusinessRelatedEndpoints(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// Setup test environment
	suite := setupBusinessEndpointTestSuite(t)
	defer suite.cleanup()

	// Test classification endpoints
	t.Run("ClassificationEndpoints", suite.testClassificationEndpoints)

	// Test merchant management endpoints
	t.Run("MerchantManagementEndpoints", suite.testMerchantManagementEndpoints)

	// Test risk assessment endpoints
	t.Run("RiskAssessmentEndpoints", suite.testRiskAssessmentEndpoints)

	// Test business analytics endpoints
	t.Run("BusinessAnalyticsEndpoints", suite.testBusinessAnalyticsEndpoints)
}

// BusinessEndpointTestSuite provides testing for business-related endpoints
type BusinessEndpointTestSuite struct {
	server  *httptest.Server
	mux     *http.ServeMux
	logger  *observability.Logger
	cleanup func()
}

// setupBusinessEndpointTestSuite sets up the business endpoint test suite
func setupBusinessEndpointTestSuite(t *testing.T) *BusinessEndpointTestSuite {
	// Setup logger
	logger := observability.NewLogger(&observability.Config{
		LogLevel:  "debug",
		LogFormat: "json",
	})

	// Create main mux
	mux := http.NewServeMux()

	// Setup mock services
	mockClassificationService := mocks.NewMockClassificationService()
	mockRiskService := mocks.NewMockRiskService()
	mockAnalyticsService := mocks.NewMockAnalyticsService()

	// Setup handlers
	classificationHandler := handlers.NewClassificationHandler(mockClassificationService, logger)
	riskHandler := handlers.NewRiskHandler(mockRiskService, logger)
	analyticsHandler := handlers.NewAnalyticsHandler(mockAnalyticsService, logger)

	// Register business-related routes
	registerBusinessRoutes(mux, classificationHandler, riskHandler, analyticsHandler)

	// Create test server
	server := httptest.NewServer(mux)

	return &BusinessEndpointTestSuite{
		server:  server,
		mux:     mux,
		logger:  logger,
		cleanup: func() { server.Close() },
	}
}

// registerBusinessRoutes registers all business-related API routes
func registerBusinessRoutes(
	mux *http.ServeMux,
	classificationHandler *handlers.ClassificationHandler,
	riskHandler *handlers.RiskHandler,
	analyticsHandler *handlers.AnalyticsHandler,
) {
	// Classification endpoints
	mux.HandleFunc("POST /v1/classify", classificationHandler.ClassifyBusiness)
	mux.HandleFunc("POST /v1/classify/batch", classificationHandler.ClassifyBusinesses)
	mux.HandleFunc("GET /v1/classify/{business_id}", classificationHandler.GetClassification)
	mux.HandleFunc("GET /v1/classify/history", classificationHandler.GetClassificationHistory)

	// Enhanced classification endpoints
	mux.HandleFunc("POST /v2/classify", classificationHandler.EnhancedClassifyBusiness)
	mux.HandleFunc("POST /v2/classify/batch", classificationHandler.EnhancedClassifyBusinesses)

	// Merchant management endpoints
	mux.HandleFunc("POST /v1/merchants", classificationHandler.CreateMerchant)
	mux.HandleFunc("GET /v1/merchants/{merchant_id}", classificationHandler.GetMerchant)
	mux.HandleFunc("PUT /v1/merchants/{merchant_id}", classificationHandler.UpdateMerchant)
	mux.HandleFunc("DELETE /v1/merchants/{merchant_id}", classificationHandler.DeleteMerchant)
	mux.HandleFunc("GET /v1/merchants", classificationHandler.ListMerchants)

	// Risk assessment endpoints
	mux.HandleFunc("POST /v1/risk/assess", riskHandler.AssessRisk)
	mux.HandleFunc("GET /v1/risk/{business_id}", riskHandler.GetRiskAssessment)
	mux.HandleFunc("GET /v1/risk/history/{business_id}", riskHandler.GetRiskHistory)
	mux.HandleFunc("POST /v1/risk/enhanced/assess", riskHandler.EnhancedRiskAssessment)
	mux.HandleFunc("GET /v1/risk/alerts", riskHandler.GetRiskAlerts)

	// Business analytics endpoints
	mux.HandleFunc("GET /v1/analytics/business/{business_id}", analyticsHandler.GetBusinessAnalytics)
	mux.HandleFunc("GET /v1/analytics/business/{business_id}/trends", analyticsHandler.GetBusinessTrends)
	mux.HandleFunc("GET /v1/analytics/business/{business_id}/comparison", analyticsHandler.GetBusinessComparison)
}

// testClassificationEndpoints tests all classification-related endpoints
func (suite *BusinessEndpointTestSuite) testClassificationEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Single classification
		{
			name:   "POST /v1/classify - Single business classification",
			method: "POST",
			path:   "/v1/classify",
			body: map[string]interface{}{
				"business_name": "Acme Corporation",
				"address":       "123 Business St, New York, NY 10001",
				"website":       "https://acme.com",
				"description":   "A technology company specializing in software development",
			},
			expectedStatus: http.StatusOK,
			description:    "Should classify a single business successfully",
		},
		{
			name:   "POST /v1/classify - Minimal business data",
			method: "POST",
			path:   "/v1/classify",
			body: map[string]interface{}{
				"business_name": "Simple Business",
				"address":       "456 Simple St",
			},
			expectedStatus: http.StatusOK,
			description:    "Should classify business with minimal data",
		},
		{
			name:   "POST /v1/classify - Complex business data",
			method: "POST",
			path:   "/v1/classify",
			body: map[string]interface{}{
				"business_name":  "Complex Multi-Industry Corp",
				"address":        "789 Complex Blvd, Complex City, CC 54321",
				"website":        "https://complexcorp.com",
				"description":    "A diversified corporation operating in technology, healthcare, and financial services",
				"contact_email":  "info@complexcorp.com",
				"phone":          "+1-555-987-6543",
				"industry_codes": []string{"541511", "621111", "522110"},
			},
			expectedStatus: http.StatusOK,
			description:    "Should classify complex business with multiple industry codes",
		},

		// Batch classification
		{
			name:   "POST /v1/classify/batch - Batch business classification",
			method: "POST",
			path:   "/v1/classify/batch",
			body: map[string]interface{}{
				"businesses": []map[string]interface{}{
					{
						"business_name": "Tech Company A",
						"address":       "123 Tech St",
						"description":   "Software development company",
					},
					{
						"business_name": "Retail Company B",
						"address":       "456 Retail Ave",
						"description":   "E-commerce retail business",
					},
					{
						"business_name": "Healthcare Company C",
						"address":       "789 Health Blvd",
						"description":   "Medical services provider",
					},
				},
			},
			expectedStatus: http.StatusOK,
			description:    "Should classify multiple businesses in batch",
		},

		// Enhanced classification
		{
			name:   "POST /v2/classify - Enhanced classification with ML",
			method: "POST",
			path:   "/v2/classify",
			body: map[string]interface{}{
				"business_name": "AI Technology Corp",
				"description":   "Artificial intelligence and machine learning solutions provider",
				"website":       "https://aitech.com",
				"use_ml_models": true,
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform enhanced classification using ML models",
		},

		// Classification retrieval
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
			description:    "Should retrieve classification history with pagination",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testMerchantManagementEndpoints tests all merchant management endpoints
func (suite *BusinessEndpointTestSuite) testMerchantManagementEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Create merchant
		{
			name:   "POST /v1/merchants - Create new merchant",
			method: "POST",
			path:   "/v1/merchants",
			body: map[string]interface{}{
				"business_name": "New Merchant Corp",
				"address":       "789 Merchant St, Merchant City, MC 12345",
				"contact_email": "contact@merchant.com",
				"phone":         "+1-555-123-4567",
				"website":       "https://merchant.com",
				"business_type": "corporation",
				"industry":      "technology",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should create a new merchant successfully",
		},
		{
			name:   "POST /v1/merchants - Create merchant with minimal data",
			method: "POST",
			path:   "/v1/merchants",
			body: map[string]interface{}{
				"business_name": "Minimal Merchant",
				"address":       "123 Minimal St",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should create merchant with minimal required data",
		},

		// Get merchant
		{
			name:           "GET /v1/merchants/{merchant_id} - Get merchant",
			method:         "GET",
			path:           "/v1/merchants/test-merchant-123",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve merchant information",
		},
		{
			name:           "GET /v1/merchants - List merchants",
			method:         "GET",
			path:           "/v1/merchants?limit=10&offset=0",
			expectedStatus: http.StatusOK,
			description:    "Should list merchants with pagination",
		},

		// Update merchant
		{
			name:   "PUT /v1/merchants/{merchant_id} - Update merchant",
			method: "PUT",
			path:   "/v1/merchants/test-merchant-123",
			body: map[string]interface{}{
				"business_name": "Updated Merchant Corp",
				"address":       "789 Updated St, Updated City, UC 54321",
				"contact_email": "updated@merchant.com",
				"phone":         "+1-555-987-6543",
			},
			expectedStatus: http.StatusOK,
			description:    "Should update merchant information",
		},
		{
			name:   "PUT /v1/merchants/{merchant_id} - Partial update",
			method: "PUT",
			path:   "/v1/merchants/test-merchant-123",
			body: map[string]interface{}{
				"contact_email": "newemail@merchant.com",
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform partial update of merchant",
		},

		// Delete merchant
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

// testRiskAssessmentEndpoints tests all risk assessment endpoints
func (suite *BusinessEndpointTestSuite) testRiskAssessmentEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Basic risk assessment
		{
			name:   "POST /v1/risk/assess - Basic risk assessment",
			method: "POST",
			path:   "/v1/risk/assess",
			body: map[string]interface{}{
				"business_id":     "test-business-123",
				"assessment_type": "basic",
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform basic risk assessment",
		},
		{
			name:   "POST /v1/risk/assess - Comprehensive risk assessment",
			method: "POST",
			path:   "/v1/risk/assess",
			body: map[string]interface{}{
				"business_id":             "test-business-123",
				"assessment_type":         "comprehensive",
				"include_factors":         true,
				"include_recommendations": true,
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform comprehensive risk assessment",
		},

		// Enhanced risk assessment
		{
			name:   "POST /v1/risk/enhanced/assess - Enhanced risk assessment",
			method: "POST",
			path:   "/v1/risk/enhanced/assess",
			body: map[string]interface{}{
				"business_id":         "test-business-123",
				"use_ml_models":       true,
				"include_trends":      true,
				"include_correlation": true,
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform enhanced risk assessment with ML models",
		},

		// Risk retrieval
		{
			name:           "GET /v1/risk/{business_id} - Get risk assessment",
			method:         "GET",
			path:           "/v1/risk/test-business-123",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve risk assessment results",
		},
		{
			name:           "GET /v1/risk/history/{business_id} - Get risk history",
			method:         "GET",
			path:           "/v1/risk/history/test-business-123?limit=10",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve risk assessment history",
		},
		{
			name:           "GET /v1/risk/alerts - Get risk alerts",
			method:         "GET",
			path:           "/v1/risk/alerts?status=active",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve active risk alerts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testBusinessAnalyticsEndpoints tests all business analytics endpoints
func (suite *BusinessEndpointTestSuite) testBusinessAnalyticsEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Business analytics
		{
			name:           "GET /v1/analytics/business/{business_id} - Get business analytics",
			method:         "GET",
			path:           "/v1/analytics/business/test-business-123",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve comprehensive business analytics",
		},
		{
			name:           "GET /v1/analytics/business/{business_id}/trends - Get business trends",
			method:         "GET",
			path:           "/v1/analytics/business/test-business-123/trends?period=30d",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve business trend analysis",
		},
		{
			name:           "GET /v1/analytics/business/{business_id}/comparison - Get business comparison",
			method:         "GET",
			path:           "/v1/analytics/business/test-business-123/comparison?industry=technology",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve business comparison with industry peers",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testEndpoint is a helper function to test individual endpoints
func (suite *BusinessEndpointTestSuite) testEndpoint(t *testing.T, method, path string, body interface{}, expectedStatus int, description string) {
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
	suite.logger.Info("Business endpoint test completed", map[string]interface{}{
		"method":          method,
		"path":            path,
		"status":          resp.StatusCode,
		"expected_status": expectedStatus,
		"description":     description,
		"success":         resp.StatusCode == expectedStatus,
	})
}

// TestBusinessEndpointPerformance tests business endpoint performance
func TestBusinessEndpointPerformance(t *testing.T) {
	// Skip if not running performance tests
	if os.Getenv("PERFORMANCE_TESTS") != "true" {
		t.Skip("Skipping performance tests - set PERFORMANCE_TESTS=true to run")
	}

	suite := setupBusinessEndpointTestSuite(t)
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
			name:   "Batch classification performance",
			method: "POST",
			path:   "/v1/classify/batch",
			body: map[string]interface{}{
				"businesses": []map[string]interface{}{
					{"business_name": "Business A"},
					{"business_name": "Business B"},
					{"business_name": "Business C"},
				},
			},
			maxDuration: 5 * time.Second,
			description: "Batch classification should complete within 5 seconds",
		},
		{
			name:        "Risk assessment performance",
			method:      "POST",
			path:        "/v1/risk/assess",
			body:        map[string]interface{}{"business_id": "test-business-123"},
			maxDuration: 3 * time.Second,
			description: "Risk assessment should complete within 3 seconds",
		},
		{
			name:        "Merchant retrieval performance",
			method:      "GET",
			path:        "/v1/merchants/test-merchant-123",
			maxDuration: 500 * time.Millisecond,
			description: "Merchant retrieval should complete within 500ms",
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
