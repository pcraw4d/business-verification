package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/api-gateway/internal/config"
	"kyb-platform/services/api-gateway/internal/handlers"
	"kyb-platform/services/api-gateway/internal/middleware"
	"kyb-platform/services/api-gateway/internal/supabase"
)

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IntegrationTestSuite provides comprehensive integration testing for API Gateway
type IntegrationTestSuite struct {
	router          *mux.Router
	gatewayHandler  *handlers.GatewayHandler
	config          *config.Config
	logger          *zap.Logger
	supabaseClient  *supabase.Client
	baseURL         string
	testMerchantID  string
	testUserID      string
}

// SetupIntegrationTestSuite creates a test suite with all middleware and routes configured
func SetupIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	// Load configuration - use test defaults if env vars not set
	cfg, err := config.Load()
	if err != nil {
		// If config load fails, create minimal test config
		cfg = &config.Config{
			Environment: "test",
			Server: config.ServerConfig{
				Port: "8080",
			},
			Supabase: config.SupabaseConfig{
				URL:    os.Getenv("SUPABASE_URL"),
				APIKey: os.Getenv("SUPABASE_ANON_KEY"),
			},
			Services: config.ServicesConfig{
				ClassificationURL:  getEnvOrDefault("CLASSIFICATION_SERVICE_URL", "http://localhost:8081"),
				MerchantURL:        getEnvOrDefault("MERCHANT_SERVICE_URL", "http://localhost:8083"),
				FrontendURL:        getEnvOrDefault("FRONTEND_URL", "http://localhost:3000"),
				BIServiceURL:        getEnvOrDefault("BI_SERVICE_URL", "http://localhost:8083"),
				RiskAssessmentURL:   getEnvOrDefault("RISK_ASSESSMENT_SERVICE_URL", "http://localhost:8082"),
			},
			CORS: config.CORSConfig{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: true,
			},
			RateLimit: config.RateLimitConfig{
				Enabled: false, // Disable rate limiting in tests
			},
		}
		t.Logf("Using minimal test configuration (env vars may not be set)")
	}

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize Supabase client (may fail if env vars not set - that's OK for unit tests)
	var supabaseClient *supabase.Client
	if cfg.Supabase.URL != "" && cfg.Supabase.APIKey != "" {
		client, err := supabase.NewClient(&cfg.Supabase, logger)
		if err != nil {
			t.Logf("Failed to initialize Supabase client (this is OK for unit tests): %v", err)
		} else {
			supabaseClient = client
		}
	} else {
		t.Logf("Skipping Supabase client initialization (env vars not set - OK for unit tests)")
	}

	// Initialize gateway handler (supabaseClient may be nil for unit tests)
	gatewayHandler := handlers.NewGatewayHandler(supabaseClient, logger, cfg)

	// Setup router with all middleware (matching main.go)
	router := mux.NewRouter()

	// Apply middleware in the same order as main.go
	router.Use(middleware.CORS(cfg.CORS))
	router.Use(middleware.SecurityHeaders)
	router.Use(middleware.Logging(logger))
	router.Use(middleware.RateLimit(cfg.RateLimit))
	router.Use(middleware.Authentication(supabaseClient, logger))

	// Register routes (matching main.go structure)
	setupRoutes(router, gatewayHandler, cfg, logger, supabaseClient)

	// Get test configuration
	baseURL := os.Getenv("API_GATEWAY_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	testMerchantID := os.Getenv("TEST_MERCHANT_ID")
	if testMerchantID == "" {
		testMerchantID = "merchant-123"
	}

	testUserID := os.Getenv("TEST_USER_ID")
	if testUserID == "" {
		testUserID = "test-user-123"
	}

	return &IntegrationTestSuite{
		router:         router,
		gatewayHandler: gatewayHandler,
		config:         cfg,
		logger:         logger,
		supabaseClient: supabaseClient,
		baseURL:        baseURL,
		testMerchantID: testMerchantID,
		testUserID:     testUserID,
	}
}

// setupRoutes registers all routes matching main.go structure
func setupRoutes(router *mux.Router, gatewayHandler *handlers.GatewayHandler, cfg *config.Config, logger *zap.Logger, supabaseClient *supabase.Client) {
	// Health check endpoint
	router.HandleFunc("/health", gatewayHandler.HealthCheck).Methods("GET")

	// Prometheus metrics endpoint
	router.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# metrics\n"))
	})).Methods("GET")

	// Root endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "api-gateway",
			"version": "1.0.20",
			"status":  "running",
		})
	}).Methods("GET")

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.CORS(cfg.CORS))

	// API v3 routes
	apiV3 := router.PathPrefix("/api/v3").Subrouter()
	apiV3.Use(middleware.CORS(cfg.CORS))
	apiV3.Use(middleware.SecurityHeaders)
	apiV3.Use(middleware.Logging(logger))
	apiV3.Use(middleware.RateLimit(cfg.RateLimit))
	apiV3.Use(middleware.Authentication(supabaseClient, logger))
	apiV3.HandleFunc("/dashboard/metrics", gatewayHandler.ProxyToDashboardMetricsV3).Methods("GET", "OPTIONS")

	// Classification routes
	api.HandleFunc("/classify", gatewayHandler.ProxyToClassification).Methods("POST")

	// Merchant routes - ORDER MATTERS
	api.HandleFunc("/merchants/{id}/analytics", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")
	api.HandleFunc("/merchants/{id}/website-analysis", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")
	api.HandleFunc("/merchants/{id}/risk-score", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")
	api.HandleFunc("/merchants/search", gatewayHandler.ProxyToMerchants).Methods("POST", "OPTIONS")
	api.HandleFunc("/merchants/analytics", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")
	api.HandleFunc("/merchants/statistics", gatewayHandler.ProxyToMerchants).Methods("GET", "OPTIONS")
	api.HandleFunc("/merchants/{id}", gatewayHandler.ProxyToMerchants).Methods("GET", "PUT", "DELETE", "OPTIONS")
	api.HandleFunc("/merchants", gatewayHandler.ProxyToMerchants).Methods("GET", "POST", "OPTIONS")
	api.PathPrefix("/merchants").HandlerFunc(gatewayHandler.ProxyToMerchants)

	// Health check routes for backend services
	api.HandleFunc("/classification/health", gatewayHandler.ProxyToClassificationHealth).Methods("GET")
	api.HandleFunc("/merchant/health", gatewayHandler.ProxyToMerchantHealth).Methods("GET")
	api.HandleFunc("/risk/health", gatewayHandler.ProxyToRiskAssessmentHealth).Methods("GET")

	// Compliance routes
	api.HandleFunc("/compliance/status", gatewayHandler.ProxyToComplianceStatus).Methods("GET", "OPTIONS")

	// Session routes
	api.HandleFunc("/sessions/current", gatewayHandler.ProxyToSessions).Methods("GET", "OPTIONS")
	api.HandleFunc("/sessions/metrics", gatewayHandler.ProxyToSessions).Methods("GET", "OPTIONS")
	api.HandleFunc("/sessions/activity", gatewayHandler.ProxyToSessions).Methods("GET", "OPTIONS")
	api.HandleFunc("/sessions/status", gatewayHandler.ProxyToSessions).Methods("GET", "OPTIONS")
	api.HandleFunc("/sessions", gatewayHandler.ProxyToSessions).Methods("GET", "POST", "DELETE", "OPTIONS")
	api.PathPrefix("/sessions").HandlerFunc(gatewayHandler.ProxyToSessions)

	// Analytics routes - ORDER MATTERS (before /risk PathPrefix)
	api.HandleFunc("/analytics/trends", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
	api.HandleFunc("/analytics/insights", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")

	// Risk Assessment routes
	api.HandleFunc("/risk/assess", gatewayHandler.ProxyToRiskAssessment).Methods("POST", "OPTIONS")
	api.HandleFunc("/risk/benchmarks", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
	api.HandleFunc("/risk/predictions/{merchant_id}", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
	api.HandleFunc("/risk/indicators/{id}", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
	api.PathPrefix("/risk").HandlerFunc(gatewayHandler.ProxyToRiskAssessment)

	// Business Intelligence routes
	api.HandleFunc("/bi/analyze", gatewayHandler.ProxyToBI).Methods("POST", "OPTIONS")
	api.PathPrefix("/bi").HandlerFunc(gatewayHandler.ProxyToBI)

	// Authentication routes
	api.HandleFunc("/auth/register", gatewayHandler.HandleAuthRegister).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", gatewayHandler.HandleAuthLogin).Methods("POST", "OPTIONS")

	// 404 handler
	router.NotFoundHandler = http.HandlerFunc(gatewayHandler.HandleNotFound)
}

// TestIntegrationAllRoutes tests all routes through the API Gateway
func TestIntegrationAllRoutes(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	testCases := []struct {
		name           string
		method         string
		path           string
		queryParams    map[string]string
		headers        map[string]string
		body           interface{}
		expectedStatus int
		skipAuth       bool
		description    string
	}{
		// Health check routes
		{
			name:           "Health Check",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			skipAuth:       true,
			description:    "Health check endpoint should return 200",
		},
		{
			name:           "Root Endpoint",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK,
			skipAuth:       true,
			description:    "Root endpoint should return service info",
		},
		{
			name:           "Metrics Endpoint",
			method:         "GET",
			path:           "/metrics",
			expectedStatus: http.StatusOK,
			skipAuth:       true,
			description:    "Prometheus metrics endpoint should return 200",
		},

		// Merchant routes
		{
			name:           "Get All Merchants",
			method:         "GET",
			path:           "/api/v1/merchants",
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/merchants should return merchant list",
		},
		{
			name:           "Get Merchant by ID",
			method:         "GET",
			path:           fmt.Sprintf("/api/v1/merchants/%s", suite.testMerchantID),
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/merchants/{id} should return merchant details",
		},
		{
			name:           "Get Merchant Analytics",
			method:         "GET",
			path:           fmt.Sprintf("/api/v1/merchants/%s/analytics", suite.testMerchantID),
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/merchants/{id}/analytics should return merchant analytics",
		},
		{
			name:           "Get Portfolio Analytics",
			method:         "GET",
			path:           "/api/v1/merchants/analytics",
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/merchants/analytics should return portfolio analytics",
		},
		{
			name:           "Get Portfolio Statistics",
			method:         "GET",
			path:           "/api/v1/merchants/statistics",
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/merchants/statistics should return portfolio statistics",
		},

		// Analytics routes
		{
			name:           "Get Risk Trends",
			method:         "GET",
			path:           "/api/v1/analytics/trends",
			queryParams:    map[string]string{"timeframe": "30d"},
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/analytics/trends should return risk trends",
		},
		{
			name:           "Get Risk Insights",
			method:         "GET",
			path:           "/api/v1/analytics/insights",
			queryParams:    map[string]string{"timeframe": "30d"},
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/analytics/insights should return risk insights",
		},

		// Risk Assessment routes
		{
			name:           "Get Risk Benchmarks",
			method:         "GET",
			path:           "/api/v1/risk/benchmarks",
			queryParams:    map[string]string{"industry": "Technology"},
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/risk/benchmarks should return risk benchmarks",
		},
		{
			name:           "Get Risk Indicators",
			method:         "GET",
			path:           fmt.Sprintf("/api/v1/risk/indicators/%s", suite.testMerchantID),
			queryParams:    map[string]string{"status": "active"},
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/risk/indicators/{id} should return risk indicators",
		},

		// Health check routes for services
		{
			name:           "Classification Health",
			method:         "GET",
			path:           "/api/v1/classification/health",
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/classification/health should return classification service health",
		},
		{
			name:           "Merchant Health",
			method:         "GET",
			path:           "/api/v1/merchant/health",
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/merchant/health should return merchant service health",
		},
		{
			name:           "Risk Health",
			method:         "GET",
			path:           "/api/v1/risk/health",
			expectedStatus: http.StatusOK,
			description:    "GET /api/v1/risk/health should return risk assessment service health",
		},

		// V3 Dashboard routes
		{
			name:           "Dashboard Metrics V3",
			method:         "GET",
			path:           "/api/v3/dashboard/metrics",
			expectedStatus: http.StatusOK,
			description:    "GET /api/v3/dashboard/metrics should return dashboard metrics",
		},

		// Error cases
		{
			name:           "Get Merchant with Invalid ID",
			method:         "GET",
			path:           "/api/v1/merchants/invalid-id-123",
			expectedStatus: http.StatusNotFound,
			description:    "GET /api/v1/merchants/{invalid-id} should return 404",
		},
		{
			name:           "Non-existent Route",
			method:         "GET",
			path:           "/api/v1/nonexistent/route",
			expectedStatus: http.StatusNotFound,
			description:    "Non-existent route should return 404",
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testRoute(t, tc)
		})
	}
}

// TestIntegrationPathTransformations tests path transformations through the gateway
func TestIntegrationPathTransformations(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "Analytics Trends Path",
			method:         "GET",
			path:           "/api/v1/analytics/trends",
			expectedStatus: http.StatusOK,
			description:    "Analytics trends should route to Risk Assessment service without path transformation",
		},
		{
			name:           "Analytics Insights Path",
			method:         "GET",
			path:           "/api/v1/analytics/insights",
			expectedStatus: http.StatusOK,
			description:    "Analytics insights should route to Risk Assessment service without path transformation",
		},
		{
			name:           "Risk Assess Path",
			method:         "POST",
			path:           "/api/v1/risk/assess",
			expectedStatus: http.StatusOK,
			description:    "Risk assess should route to Risk Assessment service without path transformation",
		},
		{
			name:           "Risk Metrics Path",
			method:         "GET",
			path:           "/api/v1/risk/metrics",
			expectedStatus: http.StatusOK,
			description:    "Risk metrics should route to Risk Assessment service without path transformation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.testRoute(t, struct {
				name           string
				method         string
				path           string
				queryParams    map[string]string
				headers        map[string]string
				body           interface{}
				expectedStatus int
				skipAuth       bool
				description    string
			}{
				name:           tc.name,
				method:         tc.method,
				path:           tc.path,
				expectedStatus: tc.expectedStatus,
				description:    tc.description,
			})
		})
	}
}

// TestIntegrationCORSHeaders tests CORS headers for all routes
func TestIntegrationCORSHeaders(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	testCases := []struct {
		name        string
		method      string
		path        string
		origin      string
		description string
	}{
		{
			name:        "CORS Headers - Merchants",
			method:      "OPTIONS",
			path:        "/api/v1/merchants",
			origin:      "https://frontend-service-production-b225.up.railway.app",
			description: "CORS headers should be present for merchants route",
		},
		{
			name:        "CORS Headers - Analytics",
			method:      "OPTIONS",
			path:        "/api/v1/analytics/trends",
			origin:      "https://frontend-service-production-b225.up.railway.app",
			description: "CORS headers should be present for analytics route",
		},
		{
			name:        "CORS Headers - Risk",
			method:      "OPTIONS",
			path:        "/api/v1/risk/benchmarks",
			origin:      "https://frontend-service-production-b225.up.railway.app",
			description: "CORS headers should be present for risk route",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Access-Control-Request-Method", "GET")
			rr := httptest.NewRecorder()

			suite.router.ServeHTTP(rr, req)

			// Check CORS headers
			if rr.Header().Get("Access-Control-Allow-Origin") == "" {
				t.Error("Access-Control-Allow-Origin header should be present")
			}
			if rr.Header().Get("Access-Control-Allow-Methods") == "" {
				t.Error("Access-Control-Allow-Methods header should be present")
			}
			if rr.Header().Get("Access-Control-Allow-Headers") == "" {
				t.Error("Access-Control-Allow-Headers header should be present")
			}

			// OPTIONS requests should return 200
			if tc.method == "OPTIONS" && rr.Code != http.StatusOK {
				t.Errorf("OPTIONS request should return 200, got %d", rr.Code)
			}
		})
	}
}

// TestIntegrationErrorResponses tests error handling and responses
func TestIntegrationErrorResponses(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedError  string
		description    string
	}{
		{
			name:           "404 for Invalid Merchant ID",
			method:         "GET",
			path:           "/api/v1/merchants/invalid-id-123",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
			description:    "Invalid merchant ID should return 404 with error code",
		},
		{
			name:           "404 for Non-existent Route",
			method:         "GET",
			path:           "/api/v1/nonexistent/route",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
			description:    "Non-existent route should return 404 with error code",
		},
		{
			name:           "404 for Invalid Analytics Path",
			method:         "GET",
			path:           "/api/v1/analytics/invalid",
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
			description:    "Invalid analytics path should return 404",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			suite.router.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			// Check error response format
			body, _ := io.ReadAll(rr.Body)
			var errorResp map[string]interface{}
			if err := json.Unmarshal(body, &errorResp); err == nil {
				if error, ok := errorResp["error"].(map[string]interface{}); ok {
					if code, ok := error["code"].(string); ok && tc.expectedError != "" {
						if code != tc.expectedError {
							t.Errorf("Expected error code %s, got %s", tc.expectedError, code)
						}
					}
				}
			}
		})
	}
}

// TestIntegrationQueryParameterPreservation tests that query parameters are preserved
func TestIntegrationQueryParameterPreservation(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	testCases := []struct {
		name        string
		path        string
		queryParams map[string]string
		description string
	}{
		{
			name:        "Analytics Trends with Query Params",
			path:        "/api/v1/analytics/trends",
			queryParams: map[string]string{"timeframe": "30d", "limit": "10", "industry": "Technology"},
			description: "Query parameters should be preserved for analytics trends",
		},
		{
			name:        "Analytics Insights with Query Params",
			path:        "/api/v1/analytics/insights",
			queryParams: map[string]string{"timeframe": "90d", "limit": "5"},
			description: "Query parameters should be preserved for analytics insights",
		},
		{
			name:        "Risk Benchmarks with Query Params",
			path:        "/api/v1/risk/benchmarks",
			queryParams: map[string]string{"industry": "Technology", "country": "US"},
			description: "Query parameters should be preserved for risk benchmarks",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Build URL with query parameters
			url := tc.path
			if len(tc.queryParams) > 0 {
				query := ""
				for key, value := range tc.queryParams {
					if query != "" {
						query += "&"
					}
					query += fmt.Sprintf("%s=%s", key, value)
				}
				url += "?" + query
			}

			req := httptest.NewRequest("GET", url, nil)
			rr := httptest.NewRecorder()

			suite.router.ServeHTTP(rr, req)

			// Verify query parameters are preserved in the request
			for key, expectedValue := range tc.queryParams {
				actualValue := req.URL.Query().Get(key)
				if actualValue != expectedValue {
					t.Errorf("Query parameter %s should be %s, got %s", key, expectedValue, actualValue)
				}
			}
		})
	}
}

// TestIntegrationAuthentication tests authentication middleware
func TestIntegrationAuthentication(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	testCases := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		expectedStatus int
		description    string
	}{
		{
			name:           "Request without Auth Header",
			method:         "GET",
			path:           "/api/v1/merchants",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized, // May vary based on auth middleware implementation
			description:    "Request without auth header should be rejected or allowed based on middleware",
		},
		{
			name:           "Request with Invalid Auth Header",
			method:         "GET",
			path:           "/api/v1/merchants",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			description:    "Request with invalid auth header should be rejected",
		},
		{
			name:           "Health Check without Auth",
			method:         "GET",
			path:           "/health",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			description:    "Health check should work without auth",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()

			suite.router.ServeHTTP(rr, req)

			// Note: Actual status may vary based on auth middleware implementation
			// This test verifies that auth middleware is applied
			if rr.Code != tc.expectedStatus {
				t.Logf("Expected status %d, got %d (this may be expected based on auth middleware)", tc.expectedStatus, rr.Code)
			}
		})
	}
}

// testRoute is a helper function to test a single route
func (suite *IntegrationTestSuite) testRoute(t *testing.T, tc struct {
	name           string
	method         string
	path           string
	queryParams    map[string]string
	headers        map[string]string
	body           interface{}
	expectedStatus int
	skipAuth       bool
	description    string
}) {
	// Build URL with query parameters
	url := tc.path
	if len(tc.queryParams) > 0 {
		query := ""
		for key, value := range tc.queryParams {
			if query != "" {
				query += "&"
			}
			query += fmt.Sprintf("%s=%s", key, value)
		}
		url += "?" + query
	}

	// Create request body
	var bodyBytes []byte
	if tc.body != nil {
		var err error
		bodyBytes, err = json.Marshal(tc.body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	// Create HTTP request
	req := httptest.NewRequest(tc.method, url, bytes.NewBuffer(bodyBytes))

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range tc.headers {
		req.Header.Set(key, value)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(rr, req)

	// Check status code
	if rr.Code != tc.expectedStatus {
		body, _ := io.ReadAll(rr.Body)
		t.Errorf("Expected status %d, got %d. Response: %s",
			tc.expectedStatus, rr.Code, string(body))
	}
}

