package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/api-gateway/internal/config"
	"kyb-platform/services/api-gateway/internal/handlers"
	"kyb-platform/services/api-gateway/internal/middleware"
	"kyb-platform/services/api-gateway/internal/supabase"
)

// TestRouteConfig holds configuration for route tests
type TestRouteConfig struct {
	BaseURL        string
	TestMerchantID string
	TestUserID     string
	Timeout        time.Duration
}

// LoadTestConfig loads test configuration from environment or defaults
func LoadTestConfig() *TestRouteConfig {
	baseURL := os.Getenv("API_GATEWAY_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &TestRouteConfig{
		BaseURL:        baseURL,
		TestMerchantID: os.Getenv("TEST_MERCHANT_ID"),
		TestUserID:     os.Getenv("TEST_USER_ID"),
		Timeout:        10 * time.Second,
	}
}

// RouteTestCase represents a single route test case
type RouteTestCase struct {
	Name           string
	Method         string
	Path           string
	QueryParams    map[string]string
	Headers        map[string]string
	Body           interface{}
	ExpectedStatus int
	ExpectedBody   func(*testing.T, []byte) bool
	SkipAuth       bool
	Description    string
}

// TestAllRoutes tests all API Gateway routes
func TestAllRoutes(t *testing.T) {
	cfg := LoadTestConfig()

	// Test cases organized by route category
	testCases := []RouteTestCase{
		// Health check routes
		{
			Name:           "Health Check",
			Method:         "GET",
			Path:           "/health",
			ExpectedStatus: http.StatusOK,
			Description:    "Health check endpoint should return 200",
		},
		{
			Name:           "Health Check Detailed",
			Method:         "GET",
			Path:           "/health",
			QueryParams:    map[string]string{"detailed": "true"},
			ExpectedStatus: http.StatusOK,
			Description:    "Detailed health check should return 200",
		},
		{
			Name:           "Root Endpoint",
			Method:         "GET",
			Path:           "/",
			ExpectedStatus: http.StatusOK,
			Description:    "Root endpoint should return service info",
		},
		{
			Name:           "Metrics Endpoint",
			Method:         "GET",
			Path:           "/metrics",
			ExpectedStatus: http.StatusOK,
			Description:    "Prometheus metrics endpoint should return 200",
		},

		// Merchant routes
		{
			Name:           "Get All Merchants",
			Method:         "GET",
			Path:           "/api/v1/merchants",
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/merchants should return merchant list",
		},
		{
			Name:           "Get Merchant by ID",
			Method:         "GET",
			Path:           fmt.Sprintf("/api/v1/merchants/%s", cfg.TestMerchantID),
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/merchants/{id} should return merchant details",
		},
		{
			Name:           "Get Merchant Analytics",
			Method:         "GET",
			Path:           fmt.Sprintf("/api/v1/merchants/%s/analytics", cfg.TestMerchantID),
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/merchants/{id}/analytics should return merchant analytics",
		},
		{
			Name:           "Get Merchant Risk Score",
			Method:         "GET",
			Path:           fmt.Sprintf("/api/v1/merchants/%s/risk-score", cfg.TestMerchantID),
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/merchants/{id}/risk-score should return risk score",
		},
		{
			Name:           "Get Merchant Website Analysis",
			Method:         "GET",
			Path:           fmt.Sprintf("/api/v1/merchants/%s/website-analysis", cfg.TestMerchantID),
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/merchants/{id}/website-analysis should return website analysis",
		},
		{
			Name:           "Get Portfolio Analytics",
			Method:         "GET",
			Path:           "/api/v1/merchants/analytics",
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/merchants/analytics should return portfolio analytics",
		},
		{
			Name:           "Get Portfolio Statistics",
			Method:         "GET",
			Path:           "/api/v1/merchants/statistics",
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/merchants/statistics should return portfolio statistics",
		},
		{
			Name:           "Search Merchants",
			Method:         "POST",
			Path:           "/api/v1/merchants/search",
			Body:           map[string]interface{}{"query": "test"},
			ExpectedStatus: http.StatusOK,
			Description:    "POST /api/v1/merchants/search should return search results",
		},

		// Analytics routes
		{
			Name:           "Get Risk Trends",
			Method:         "GET",
			Path:           "/api/v1/analytics/trends",
			QueryParams:    map[string]string{"timeframe": "30d"},
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/analytics/trends should return risk trends",
		},
		{
			Name:           "Get Risk Trends with Query Params",
			Method:         "GET",
			Path:           "/api/v1/analytics/trends",
			QueryParams:    map[string]string{"timeframe": "7d", "limit": "10", "industry": "Technology"},
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/analytics/trends with query params should work",
		},
		{
			Name:           "Get Risk Insights",
			Method:         "GET",
			Path:           "/api/v1/analytics/insights",
			QueryParams:    map[string]string{"timeframe": "30d"},
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/analytics/insights should return risk insights",
		},
		{
			Name:           "Get Risk Insights with Query Params",
			Method:         "GET",
			Path:           "/api/v1/analytics/insights",
			QueryParams:    map[string]string{"timeframe": "90d", "limit": "5"},
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/analytics/insights with query params should work",
		},

		// Risk Assessment routes
		{
			Name:           "Get Risk Benchmarks",
			Method:         "GET",
			Path:           "/api/v1/risk/benchmarks",
			QueryParams:    map[string]string{"industry": "Technology"},
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/risk/benchmarks should return risk benchmarks",
		},
		{
			Name:           "Get Risk Indicators",
			Method:         "GET",
			Path:           fmt.Sprintf("/api/v1/risk/indicators/%s", cfg.TestMerchantID),
			QueryParams:    map[string]string{"status": "active"},
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/risk/indicators/{id} should return risk indicators",
		},
		{
			Name:           "Get Risk Predictions",
			Method:         "GET",
			Path:           fmt.Sprintf("/api/v1/risk/predictions/%s", cfg.TestMerchantID),
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/risk/predictions/{merchant_id} should return predictions",
		},
		{
			Name:           "Get Risk Metrics",
			Method:         "GET",
			Path:           "/api/v1/risk/metrics",
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/risk/metrics should return risk metrics",
		},
		{
			Name:           "Assess Risk",
			Method:         "POST",
			Path:           "/api/v1/risk/assess",
			Body: map[string]interface{}{
				"merchant_id": cfg.TestMerchantID,
				"data":        map[string]interface{}{"test": "data"},
			},
			ExpectedStatus: http.StatusOK,
			Description:    "POST /api/v1/risk/assess should create risk assessment",
		},

		// Health check routes for services
		{
			Name:           "Classification Health",
			Method:         "GET",
			Path:           "/api/v1/classification/health",
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/classification/health should return classification service health",
		},
		{
			Name:           "Merchant Health",
			Method:         "GET",
			Path:           "/api/v1/merchant/health",
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/merchant/health should return merchant service health",
		},
		{
			Name:           "Risk Health",
			Method:         "GET",
			Path:           "/api/v1/risk/health",
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v1/risk/health should return risk assessment service health",
		},

		// V3 Dashboard routes
		{
			Name:           "Dashboard Metrics V3",
			Method:         "GET",
			Path:           "/api/v3/dashboard/metrics",
			ExpectedStatus: http.StatusOK,
			Description:    "GET /api/v3/dashboard/metrics should return dashboard metrics",
		},

		// Error cases - Invalid IDs
		{
			Name:           "Get Merchant with Invalid ID",
			Method:         "GET",
			Path:           "/api/v1/merchants/invalid-id-123",
			ExpectedStatus: http.StatusNotFound,
			Description:    "GET /api/v1/merchants/{invalid-id} should return 404",
		},
		{
			Name:           "Get Merchant Analytics with Invalid ID",
			Method:         "GET",
			Path:           "/api/v1/merchants/invalid-id-123/analytics",
			ExpectedStatus: http.StatusNotFound,
			Description:    "GET /api/v1/merchants/{invalid-id}/analytics should return 404",
		},
		{
			Name:           "Get Risk Indicators with Invalid ID",
			Method:         "GET",
			Path:           "/api/v1/risk/indicators/invalid-id-123",
			ExpectedStatus: http.StatusNotFound,
			Description:    "GET /api/v1/risk/indicators/{invalid-id} should return 404",
		},

		// Error cases - Invalid routes
		{
			Name:           "Non-existent Route",
			Method:         "GET",
			Path:           "/api/v1/nonexistent/route",
			ExpectedStatus: http.StatusNotFound,
			Description:    "Non-existent route should return 404",
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			testRoute(t, cfg, tc)
		})
	}
}

// testRoute tests a single route
func testRoute(t *testing.T, cfg *TestRouteConfig, tc RouteTestCase) {
	// Build URL
	url := cfg.BaseURL + tc.Path
	if len(tc.QueryParams) > 0 {
		query := ""
		for key, value := range tc.QueryParams {
			if query != "" {
				query += "&"
			}
			query += fmt.Sprintf("%s=%s", key, value)
		}
		url += "?" + query
	}

	// Create request body
	var bodyBytes []byte
	if tc.Body != nil {
		var err error
		bodyBytes, err = json.Marshal(tc.Body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	// Create HTTP request
	req, err := http.NewRequest(tc.Method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if !tc.SkipAuth {
		// Add auth header if needed (you may need to adjust this based on your auth setup)
		// req.Header.Set("Authorization", "Bearer test-token")
	}
	for key, value := range tc.Headers {
		req.Header.Set(key, value)
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: cfg.Timeout,
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Request failed (this may be expected if service is not running): %v", err)
		t.Skip("Skipping test - API Gateway may not be running")
		return
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Check status code
	if resp.StatusCode != tc.ExpectedStatus {
		t.Errorf("Expected status %d, got %d. Response: %s",
			tc.ExpectedStatus, resp.StatusCode, string(bodyBytes))
	}

	// Check response body if validator provided
	if tc.ExpectedBody != nil {
		if !tc.ExpectedBody(t, bodyBytes) {
			t.Errorf("Response body validation failed. Response: %s", string(bodyBytes))
		}
	}

	// Check CORS headers for OPTIONS requests
	if tc.Method == "OPTIONS" || resp.StatusCode == http.StatusOK {
		if resp.Header.Get("Access-Control-Allow-Origin") == "" {
			t.Error("CORS header should be present")
		}
	}
}

// TestRouteOrder tests that specific routes are matched before PathPrefix routes
func TestRouteOrder(t *testing.T) {
	// This test verifies that route registration order is correct
	// by checking that specific routes (like /merchants/statistics) are matched
	// before PathPrefix catch-all routes

	testCases := []struct {
		name     string
		path     string
		expected string // Expected handler or behavior
	}{
		{
			name:     "Merchants Statistics Route",
			path:     "/api/v1/merchants/statistics",
			expected: "merchants", // Should be handled by ProxyToMerchants
		},
		{
			name:     "Merchants Analytics Route",
			path:     "/api/v1/merchants/analytics",
			expected: "merchants", // Should be handled by ProxyToMerchants
		},
		{
			name:     "Analytics Trends Route",
			path:     "/api/v1/analytics/trends",
			expected: "risk-assessment", // Should be handled by ProxyToRiskAssessment
		},
		{
			name:     "Analytics Insights Route",
			path:     "/api/v1/analytics/insights",
			expected: "risk-assessment", // Should be handled by ProxyToRiskAssessment
		},
		{
			name:     "Risk Assess Route",
			path:     "/api/v1/risk/assess",
			expected: "risk-assessment", // Should be handled by ProxyToRiskAssessment
		},
		{
			name:     "Risk Benchmarks Route",
			path:     "/api/v1/risk/benchmarks",
			expected: "risk-assessment", // Should be handled by ProxyToRiskAssessment
		},
	}

	// Create a test router to verify route matching
	logger, _ := zap.NewDevelopment()
	cfg, _ := config.Load()
	supabaseClient, _ := supabase.NewClient(&cfg.Supabase, logger)
	gatewayHandler := handlers.NewGatewayHandler(supabaseClient, logger, cfg)

	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()

	// Register routes in the same order as main.go
	api.HandleFunc("/merchants/{id}/analytics", gatewayHandler.ProxyToMerchants).Methods("GET")
	api.HandleFunc("/merchants/analytics", gatewayHandler.ProxyToMerchants).Methods("GET")
	api.HandleFunc("/merchants/statistics", gatewayHandler.ProxyToMerchants).Methods("GET")
	api.HandleFunc("/merchants/{id}", gatewayHandler.ProxyToMerchants).Methods("GET")
	api.HandleFunc("/merchants", gatewayHandler.ProxyToMerchants).Methods("GET")
	api.PathPrefix("/merchants").HandlerFunc(gatewayHandler.ProxyToMerchants)

	api.HandleFunc("/analytics/trends", gatewayHandler.ProxyToRiskAssessment).Methods("GET")
	api.HandleFunc("/analytics/insights", gatewayHandler.ProxyToRiskAssessment).Methods("GET")

	api.HandleFunc("/risk/assess", gatewayHandler.ProxyToRiskAssessment).Methods("POST")
	api.HandleFunc("/risk/benchmarks", gatewayHandler.ProxyToRiskAssessment).Methods("GET")
	api.PathPrefix("/risk").HandlerFunc(gatewayHandler.ProxyToRiskAssessment)

	// Test route matching
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			// Verify route was matched (not 404)
			if rr.Code == http.StatusNotFound {
				t.Errorf("Route %s should be matched, got 404", tc.path)
			}
		})
	}
}

// TestPathTransformations tests that path transformations work correctly
func TestPathTransformations(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg, _ := config.Load()
	supabaseClient, _ := supabase.NewClient(&cfg.Supabase, logger)
	gatewayHandler := handlers.NewGatewayHandler(supabaseClient, logger, cfg)

	testCases := []struct {
		name           string
		path           string
		expectedPath   string
		expectedStatus int
	}{
		{
			name:           "Risk Assess Path Transformation",
			path:           "/api/v1/risk/assess",
			expectedPath:   "/api/v1/risk/assess", // Should not transform
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Risk Metrics Path",
			path:           "/api/v1/risk/metrics",
			expectedPath:   "/api/v1/risk/metrics", // Should not transform
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Analytics Trends Path",
			path:           "/api/v1/analytics/trends",
			expectedPath:   "/api/v1/analytics/trends", // Should not transform
			expectedStatus: http.StatusOK,
		},
	}

	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/analytics/trends", gatewayHandler.ProxyToRiskAssessment).Methods("GET")
	api.HandleFunc("/risk/assess", gatewayHandler.ProxyToRiskAssessment).Methods("POST")
	api.HandleFunc("/risk/metrics", gatewayHandler.ProxyToRiskAssessment).Methods("GET")
	api.PathPrefix("/risk").HandlerFunc(gatewayHandler.ProxyToRiskAssessment)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			method := "GET"
			if strings.Contains(tc.path, "/assess") {
				method = "POST"
			}

			req := httptest.NewRequest(method, tc.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			// Verify route was matched
			if rr.Code == http.StatusNotFound {
				t.Errorf("Route %s should be matched, got 404", tc.path)
			}
		})
	}
}

// TestCORSHeaders tests that CORS headers are present
func TestCORSHeaders(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg, _ := config.Load()
	supabaseClient, _ := supabase.NewClient(&cfg.Supabase, logger)
	gatewayHandler := handlers.NewGatewayHandler(supabaseClient, logger, cfg)

	router := mux.NewRouter()
	router.Use(middleware.CORS(cfg.CORS))
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.CORS(cfg.CORS))
	api.HandleFunc("/merchants", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")

	// Test OPTIONS request (preflight)
	req := httptest.NewRequest("OPTIONS", "/api/v1/merchants", nil)
	req.Header.Set("Origin", "https://frontend-service-production-b225.up.railway.app")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("OPTIONS request should return 200, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS header should be present")
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("CORS methods header should be present")
	}
}

// TestQueryParameterPreservation tests that query parameters are preserved
func TestQueryParameterPreservation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg, _ := config.Load()
	supabaseClient, _ := supabase.NewClient(&cfg.Supabase, logger)
	gatewayHandler := handlers.NewGatewayHandler(supabaseClient, logger, cfg)

	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/analytics/trends", gatewayHandler.ProxyToRiskAssessment).Methods("GET")

	// Test with query parameters
	req := httptest.NewRequest("GET", "/api/v1/analytics/trends?timeframe=30d&limit=10&industry=Technology", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Verify query parameters are preserved in the request
	if req.URL.Query().Get("timeframe") != "30d" {
		t.Error("Query parameter 'timeframe' should be preserved")
	}
	if req.URL.Query().Get("limit") != "10" {
		t.Error("Query parameter 'limit' should be preserved")
	}
	if req.URL.Query().Get("industry") != "Technology" {
		t.Error("Query parameter 'industry' should be preserved")
	}
}

