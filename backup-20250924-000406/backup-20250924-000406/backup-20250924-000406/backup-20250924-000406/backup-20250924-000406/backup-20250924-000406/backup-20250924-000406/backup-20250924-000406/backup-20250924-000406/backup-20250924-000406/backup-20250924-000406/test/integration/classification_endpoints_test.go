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

// TestClassificationEndpoints tests all classification-specific API endpoints
func TestClassificationEndpoints(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests - set INTEGRATION_TESTS=true to run")
	}

	// Setup test environment
	suite := setupClassificationEndpointTestSuite(t)
	defer suite.cleanup()

	// Test core classification endpoints
	t.Run("CoreClassificationEndpoints", suite.testCoreClassificationEndpoints)

	// Test enhanced classification endpoints
	t.Run("EnhancedClassificationEndpoints", suite.testEnhancedClassificationEndpoints)

	// Test classification monitoring endpoints
	t.Run("ClassificationMonitoringEndpoints", suite.testClassificationMonitoringEndpoints)

	// Test classification accuracy endpoints
	t.Run("ClassificationAccuracyEndpoints", suite.testClassificationAccuracyEndpoints)
}

// ClassificationEndpointTestSuite provides testing for classification endpoints
type ClassificationEndpointTestSuite struct {
	server  *httptest.Server
	mux     *http.ServeMux
	logger  *observability.Logger
	cleanup func()
}

// setupClassificationEndpointTestSuite sets up the classification endpoint test suite
func setupClassificationEndpointTestSuite(t *testing.T) *ClassificationEndpointTestSuite {
	// Setup logger
	logger := observability.NewLogger(&observability.Config{
		LogLevel:  "debug",
		LogFormat: "json",
	})

	// Create main mux
	mux := http.NewServeMux()

	// Setup mock services
	mockClassificationService := mocks.NewMockClassificationService()
	mockMonitoringService := mocks.NewMockMonitoringService()
	mockAccuracyService := mocks.NewMockAccuracyService()

	// Setup handlers
	classificationHandler := handlers.NewClassificationHandler(mockClassificationService, logger)
	monitoringHandler := handlers.NewMonitoringHandler(mockMonitoringService, logger)
	accuracyHandler := handlers.NewAccuracyHandler(mockAccuracyService, logger)

	// Register classification-related routes
	registerClassificationRoutes(mux, classificationHandler, monitoringHandler, accuracyHandler)

	// Create test server
	server := httptest.NewServer(mux)

	return &ClassificationEndpointTestSuite{
		server:  server,
		mux:     mux,
		logger:  logger,
		cleanup: func() { server.Close() },
	}
}

// registerClassificationRoutes registers all classification-related API routes
func registerClassificationRoutes(
	mux *http.ServeMux,
	classificationHandler *handlers.ClassificationHandler,
	monitoringHandler *handlers.MonitoringHandler,
	accuracyHandler *handlers.AccuracyHandler,
) {
	// Core classification endpoints
	mux.HandleFunc("POST /v1/classify", classificationHandler.ClassifyBusiness)
	mux.HandleFunc("POST /v1/classify/batch", classificationHandler.ClassifyBusinesses)
	mux.HandleFunc("GET /v1/classify/{business_id}", classificationHandler.GetClassification)
	mux.HandleFunc("GET /v1/classify/history", classificationHandler.GetClassificationHistory)

	// Enhanced classification endpoints (v2)
	mux.HandleFunc("POST /v2/classify", classificationHandler.EnhancedClassifyBusiness)
	mux.HandleFunc("POST /v2/classify/batch", classificationHandler.EnhancedClassifyBusinesses)
	mux.HandleFunc("GET /v2/classify/{business_id}", classificationHandler.GetEnhancedClassification)

	// Classification monitoring endpoints
	mux.HandleFunc("GET /v1/monitoring/accuracy/metrics", monitoringHandler.GetAccuracyMetrics)
	mux.HandleFunc("POST /v1/monitoring/accuracy/track", monitoringHandler.TrackClassification)
	mux.HandleFunc("GET /v1/monitoring/misclassifications", monitoringHandler.GetMisclassifications)
	mux.HandleFunc("GET /v1/monitoring/patterns", monitoringHandler.GetErrorPatterns)
	mux.HandleFunc("GET /v1/monitoring/statistics", monitoringHandler.GetErrorStatistics)

	// Classification accuracy endpoints
	mux.HandleFunc("GET /v1/accuracy/validation", accuracyHandler.GetValidationResults)
	mux.HandleFunc("POST /v1/accuracy/validate", accuracyHandler.ValidateClassification)
	mux.HandleFunc("GET /v1/accuracy/benchmarks", accuracyHandler.GetAccuracyBenchmarks)
}

// testCoreClassificationEndpoints tests core classification endpoints
func (suite *ClassificationEndpointTestSuite) testCoreClassificationEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Single classification with various business types
		{
			name:   "POST /v1/classify - Technology company",
			method: "POST",
			path:   "/v1/classify",
			body: map[string]interface{}{
				"business_name": "TechCorp Solutions",
				"address":       "123 Technology Drive, Silicon Valley, CA 94000",
				"website":       "https://techcorp.com",
				"description":   "Software development and AI solutions provider",
			},
			expectedStatus: http.StatusOK,
			description:    "Should classify technology company correctly",
		},
		{
			name:   "POST /v1/classify - Retail business",
			method: "POST",
			path:   "/v1/classify",
			body: map[string]interface{}{
				"business_name": "RetailMart Inc",
				"address":       "456 Commerce Street, Retail City, RC 12345",
				"website":       "https://retailmart.com",
				"description":   "E-commerce retail business selling consumer goods",
			},
			expectedStatus: http.StatusOK,
			description:    "Should classify retail business correctly",
		},
		{
			name:   "POST /v1/classify - Healthcare provider",
			method: "POST",
			path:   "/v1/classify",
			body: map[string]interface{}{
				"business_name": "HealthCare Plus",
				"address":       "789 Medical Center, Health City, HC 67890",
				"website":       "https://healthcareplus.com",
				"description":   "Medical services and healthcare provider",
			},
			expectedStatus: http.StatusOK,
			description:    "Should classify healthcare provider correctly",
		},
		{
			name:   "POST /v1/classify - Financial services",
			method: "POST",
			path:   "/v1/classify",
			body: map[string]interface{}{
				"business_name": "FinanceFirst Bank",
				"address":       "321 Wall Street, Financial District, NY 10004",
				"website":       "https://financefirst.com",
				"description":   "Banking and financial services provider",
			},
			expectedStatus: http.StatusOK,
			description:    "Should classify financial services correctly",
		},

		// Batch classification
		{
			name:   "POST /v1/classify/batch - Mixed industry batch",
			method: "POST",
			path:   "/v1/classify/batch",
			body: map[string]interface{}{
				"businesses": []map[string]interface{}{
					{
						"business_name": "Software Solutions LLC",
						"address":       "123 Code Street",
						"description":   "Custom software development",
					},
					{
						"business_name": "Fashion Boutique",
						"address":       "456 Style Avenue",
						"description":   "Fashion retail store",
					},
					{
						"business_name": "Dental Clinic",
						"address":       "789 Health Boulevard",
						"description":   "Dental care services",
					},
					{
						"business_name": "Investment Firm",
						"address":       "321 Money Lane",
						"description":   "Investment advisory services",
					},
				},
			},
			expectedStatus: http.StatusOK,
			description:    "Should classify mixed industry batch correctly",
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
			path:           "/v1/classify/history?business_id=test-business-123&limit=10&offset=0",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve classification history with pagination",
		},
		{
			name:           "GET /v1/classify/history - Get all classification history",
			method:         "GET",
			path:           "/v1/classify/history?limit=50",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve all classification history",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testEnhancedClassificationEndpoints tests enhanced classification endpoints
func (suite *ClassificationEndpointTestSuite) testEnhancedClassificationEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Enhanced single classification
		{
			name:   "POST /v2/classify - Enhanced technology classification",
			method: "POST",
			path:   "/v2/classify",
			body: map[string]interface{}{
				"business_name":      "AI Innovations Corp",
				"description":        "Artificial intelligence and machine learning solutions for enterprise clients",
				"website":            "https://aiinnovations.com",
				"use_ml_models":      true,
				"include_confidence": true,
				"include_reasoning":  true,
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform enhanced classification with ML models",
		},
		{
			name:   "POST /v2/classify - Enhanced retail classification",
			method: "POST",
			path:   "/v2/classify",
			body: map[string]interface{}{
				"business_name":      "E-Commerce Giant",
				"description":        "Online marketplace connecting buyers and sellers globally",
				"website":            "https://ecommercegiant.com",
				"use_ml_models":      true,
				"include_confidence": true,
				"include_reasoning":  true,
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform enhanced retail classification with ML",
		},

		// Enhanced batch classification
		{
			name:   "POST /v2/classify/batch - Enhanced batch classification",
			method: "POST",
			path:   "/v2/classify/batch",
			body: map[string]interface{}{
				"businesses": []map[string]interface{}{
					{
						"business_name": "Blockchain Solutions",
						"description":   "Blockchain technology and cryptocurrency services",
						"use_ml_models": true,
					},
					{
						"business_name": "Green Energy Corp",
						"description":   "Renewable energy solutions and sustainability consulting",
						"use_ml_models": true,
					},
					{
						"business_name": "Biotech Research",
						"description":   "Biotechnology research and pharmaceutical development",
						"use_ml_models": true,
					},
				},
				"use_ml_models":      true,
				"include_confidence": true,
			},
			expectedStatus: http.StatusOK,
			description:    "Should perform enhanced batch classification with ML models",
		},

		// Enhanced classification retrieval
		{
			name:           "GET /v2/classify/{business_id} - Get enhanced classification",
			method:         "GET",
			path:           "/v2/classify/test-business-123",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve enhanced classification result",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testClassificationMonitoringEndpoints tests classification monitoring endpoints
func (suite *ClassificationEndpointTestSuite) testClassificationMonitoringEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Accuracy metrics
		{
			name:           "GET /v1/monitoring/accuracy/metrics - Get accuracy metrics",
			method:         "GET",
			path:           "/v1/monitoring/accuracy/metrics",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve classification accuracy metrics",
		},
		{
			name:           "GET /v1/monitoring/accuracy/metrics - Get accuracy metrics with filters",
			method:         "GET",
			path:           "/v1/monitoring/accuracy/metrics?period=30d&model=bert",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve filtered accuracy metrics",
		},

		// Classification tracking
		{
			name:   "POST /v1/monitoring/accuracy/track - Track classification",
			method: "POST",
			path:   "/v1/monitoring/accuracy/track",
			body: map[string]interface{}{
				"business_id":     "test-business-123",
				"predicted_codes": []string{"541511", "541512"},
				"actual_codes":    []string{"541511"},
				"confidence":      0.95,
				"model_version":   "v2.1.0",
			},
			expectedStatus: http.StatusOK,
			description:    "Should track classification accuracy",
		},
		{
			name:   "POST /v1/monitoring/accuracy/track - Track batch classification",
			method: "POST",
			path:   "/v1/monitoring/accuracy/track",
			body: map[string]interface{}{
				"batch_id": "batch-123",
				"classifications": []map[string]interface{}{
					{
						"business_id":     "business-1",
						"predicted_codes": []string{"541511"},
						"actual_codes":    []string{"541511"},
						"confidence":      0.98,
					},
					{
						"business_id":     "business-2",
						"predicted_codes": []string{"541512"},
						"actual_codes":    []string{"541511"},
						"confidence":      0.75,
					},
				},
			},
			expectedStatus: http.StatusOK,
			description:    "Should track batch classification accuracy",
		},

		// Misclassification analysis
		{
			name:           "GET /v1/monitoring/misclassifications - Get misclassifications",
			method:         "GET",
			path:           "/v1/monitoring/misclassifications",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve misclassification data",
		},
		{
			name:           "GET /v1/monitoring/misclassifications - Get misclassifications with filters",
			method:         "GET",
			path:           "/v1/monitoring/misclassifications?severity=high&period=7d",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve filtered misclassification data",
		},

		// Error pattern analysis
		{
			name:           "GET /v1/monitoring/patterns - Get error patterns",
			method:         "GET",
			path:           "/v1/monitoring/patterns",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve error patterns",
		},
		{
			name:           "GET /v1/monitoring/patterns - Get error patterns by type",
			method:         "GET",
			path:           "/v1/monitoring/patterns?type=industry_confusion",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve error patterns by type",
		},

		// Error statistics
		{
			name:           "GET /v1/monitoring/statistics - Get error statistics",
			method:         "GET",
			path:           "/v1/monitoring/statistics",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve error statistics",
		},
		{
			name:           "GET /v1/monitoring/statistics - Get error statistics with filters",
			method:         "GET",
			path:           "/v1/monitoring/statistics?period=30d&model=all",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve filtered error statistics",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testClassificationAccuracyEndpoints tests classification accuracy endpoints
func (suite *ClassificationEndpointTestSuite) testClassificationAccuracyEndpoints(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		description    string
	}{
		// Accuracy validation
		{
			name:           "GET /v1/accuracy/validation - Get validation results",
			method:         "GET",
			path:           "/v1/accuracy/validation",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve accuracy validation results",
		},
		{
			name:           "GET /v1/accuracy/validation - Get validation results with filters",
			method:         "GET",
			path:           "/v1/accuracy/validation?model=bert&period=7d",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve filtered validation results",
		},

		// Classification validation
		{
			name:   "POST /v1/accuracy/validate - Validate classification",
			method: "POST",
			path:   "/v1/accuracy/validate",
			body: map[string]interface{}{
				"business_id": "test-business-123",
				"predicted_classification": map[string]interface{}{
					"naics_code": "541511",
					"mcc_code":   "7372",
					"sic_code":   "7372",
					"confidence": 0.95,
				},
				"actual_classification": map[string]interface{}{
					"naics_code": "541511",
					"mcc_code":   "7372",
					"sic_code":   "7372",
				},
			},
			expectedStatus: http.StatusOK,
			description:    "Should validate classification accuracy",
		},
		{
			name:   "POST /v1/accuracy/validate - Validate batch classification",
			method: "POST",
			path:   "/v1/accuracy/validate",
			body: map[string]interface{}{
				"batch_id": "batch-123",
				"validations": []map[string]interface{}{
					{
						"business_id": "business-1",
						"predicted_classification": map[string]interface{}{
							"naics_code": "541511",
							"confidence": 0.98,
						},
						"actual_classification": map[string]interface{}{
							"naics_code": "541511",
						},
					},
					{
						"business_id": "business-2",
						"predicted_classification": map[string]interface{}{
							"naics_code": "541512",
							"confidence": 0.75,
						},
						"actual_classification": map[string]interface{}{
							"naics_code": "541511",
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			description:    "Should validate batch classification accuracy",
		},

		// Accuracy benchmarks
		{
			name:           "GET /v1/accuracy/benchmarks - Get accuracy benchmarks",
			method:         "GET",
			path:           "/v1/accuracy/benchmarks",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve accuracy benchmarks",
		},
		{
			name:           "GET /v1/accuracy/benchmarks - Get benchmarks by model",
			method:         "GET",
			path:           "/v1/accuracy/benchmarks?model=bert&industry=technology",
			expectedStatus: http.StatusOK,
			description:    "Should retrieve model-specific benchmarks",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testEndpoint(t, tc.method, tc.path, tc.body, tc.expectedStatus, tc.description)
		})
	}
}

// testEndpoint is a helper function to test individual endpoints
func (suite *ClassificationEndpointTestSuite) testEndpoint(t *testing.T, method, path string, body interface{}, expectedStatus int, description string) {
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
	suite.logger.Info("Classification endpoint test completed", map[string]interface{}{
		"method":          method,
		"path":            path,
		"status":          resp.StatusCode,
		"expected_status": expectedStatus,
		"description":     description,
		"success":         resp.StatusCode == expectedStatus,
	})
}

// TestClassificationEndpointPerformance tests classification endpoint performance
func TestClassificationEndpointPerformance(t *testing.T) {
	// Skip if not running performance tests
	if os.Getenv("PERFORMANCE_TESTS") != "true" {
		t.Skip("Skipping performance tests - set PERFORMANCE_TESTS=true to run")
	}

	suite := setupClassificationEndpointTestSuite(t)
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
			name:        "Core classification performance",
			method:      "POST",
			path:        "/v1/classify",
			body:        map[string]interface{}{"business_name": "Performance Test Business"},
			maxDuration: 2 * time.Second,
			description: "Core classification should complete within 2 seconds",
		},
		{
			name:        "Enhanced classification performance",
			method:      "POST",
			path:        "/v2/classify",
			body:        map[string]interface{}{"business_name": "Enhanced Test Business", "use_ml_models": true},
			maxDuration: 3 * time.Second,
			description: "Enhanced classification should complete within 3 seconds",
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
					{"business_name": "Business D"},
					{"business_name": "Business E"},
				},
			},
			maxDuration: 5 * time.Second,
			description: "Batch classification should complete within 5 seconds",
		},
		{
			name:        "Classification retrieval performance",
			method:      "GET",
			path:        "/v1/classify/test-business-123",
			maxDuration: 500 * time.Millisecond,
			description: "Classification retrieval should complete within 500ms",
		},
		{
			name:        "Accuracy metrics performance",
			method:      "GET",
			path:        "/v1/monitoring/accuracy/metrics",
			maxDuration: 1 * time.Second,
			description: "Accuracy metrics should complete within 1 second",
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
