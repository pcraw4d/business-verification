package test

import (
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/api-gateway/internal/config"
	"kyb-platform/services/api-gateway/internal/handlers"
	"kyb-platform/services/api-gateway/internal/middleware"
	"kyb-platform/services/api-gateway/internal/supabase"
)

// LoadTestSuite provides load testing for API Gateway
type LoadTestSuite struct {
	router         *mux.Router
	gatewayHandler *handlers.GatewayHandler
	config         *config.Config
	logger         *zap.Logger
	baseURL        string
}

// SetupLoadTestSuite creates a test suite for load testing
func SetupLoadTestSuite(t *testing.T) *LoadTestSuite {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// Use minimal test config if env vars not set
		cfg = &config.Config{
			Environment: "test",
			Server: config.ServerConfig{
				Port: "8080",
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
				Enabled: false, // Disable rate limiting in load tests
			},
		}
	}

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize Supabase client (optional for load tests)
	var supabaseClient *supabase.Client
	if cfg.Supabase.URL != "" && cfg.Supabase.APIKey != "" {
		client, err := supabase.NewClient(&cfg.Supabase, logger)
		if err == nil {
			supabaseClient = client
		}
	}

	// Initialize gateway handler
	gatewayHandler := handlers.NewGatewayHandler(supabaseClient, logger, cfg)

	// Setup router with middleware
	router := mux.NewRouter()
	router.Use(middleware.CORS(cfg.CORS))
	router.Use(middleware.SecurityHeaders)
	router.Use(middleware.Logging(logger))
	router.Use(middleware.RateLimit(cfg.RateLimit))
	router.Use(middleware.Authentication(supabaseClient, logger))

	// Register routes (simplified for load tests)
	setupLoadTestRoutes(router, gatewayHandler, cfg, logger, supabaseClient)

	baseURL := os.Getenv("API_GATEWAY_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &LoadTestSuite{
		router:         router,
		gatewayHandler: gatewayHandler,
		config:         cfg,
		logger:         logger,
		baseURL:        baseURL,
	}
}

// setupLoadTestRoutes registers essential routes for load testing
func setupLoadTestRoutes(router *mux.Router, gatewayHandler *handlers.GatewayHandler, cfg *config.Config, logger *zap.Logger, supabaseClient *supabase.Client) {
	// Health check
	router.HandleFunc("/health", gatewayHandler.HealthCheck).Methods("GET")

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.CORS(cfg.CORS))

	// Merchant routes
	api.HandleFunc("/merchants/{id}/analytics", gatewayHandler.ProxyToMerchants).Methods("GET")
	api.HandleFunc("/merchants/analytics", gatewayHandler.ProxyToMerchants).Methods("GET")
	api.HandleFunc("/merchants/statistics", gatewayHandler.ProxyToMerchants).Methods("GET")
	api.HandleFunc("/merchants/{id}", gatewayHandler.ProxyToMerchants).Methods("GET")
	api.HandleFunc("/merchants", gatewayHandler.ProxyToMerchants).Methods("GET")

	// Analytics routes
	api.HandleFunc("/analytics/trends", gatewayHandler.ProxyToRiskAssessment).Methods("GET")
	api.HandleFunc("/analytics/insights", gatewayHandler.ProxyToRiskAssessment).Methods("GET")

	// Risk Assessment routes
	api.HandleFunc("/risk/benchmarks", gatewayHandler.ProxyToRiskAssessment).Methods("GET")
	api.HandleFunc("/risk/indicators/{id}", gatewayHandler.ProxyToRiskAssessment).Methods("GET")
	api.HandleFunc("/risk/metrics", gatewayHandler.ProxyToRiskAssessment).Methods("GET")
}

// LoadTestResult holds results from a load test
type LoadTestResult struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalDuration      time.Duration
	MinResponseTime    time.Duration
	MaxResponseTime    time.Duration
	AvgResponseTime    time.Duration
	P50ResponseTime    time.Duration
	P95ResponseTime    time.Duration
	P99ResponseTime    time.Duration
	Throughput         float64 // Requests per second
	ErrorRate          float64 // Percentage
	ConcurrentUsers    int
	Endpoint           string
}

// TestLoadAPIGatewayUnderLoad tests API Gateway under various load conditions
func TestLoadAPIGatewayUnderLoad(t *testing.T) {
	suite := SetupLoadTestSuite(t)

	// Test different load scenarios
	loadScenarios := []struct {
		name           string
		concurrent     int
		requestsPerUser int
		endpoint       string
		method         string
		description    string
	}{
		{
			name:           "Light Load - 10 Concurrent Users",
			concurrent:     10,
			requestsPerUser: 10,
			endpoint:       "/health",
			method:         "GET",
			description:    "Test with 10 concurrent users, 10 requests each",
		},
		{
			name:           "Medium Load - 50 Concurrent Users",
			concurrent:     50,
			requestsPerUser: 20,
			endpoint:       "/api/v1/merchants",
			method:         "GET",
			description:    "Test with 50 concurrent users, 20 requests each",
		},
		{
			name:           "Heavy Load - 100 Concurrent Users",
			concurrent:     100,
			requestsPerUser: 50,
			endpoint:       "/api/v1/merchants",
			method:         "GET",
			description:    "Test with 100 concurrent users, 50 requests each",
		},
		{
			name:           "Stress Test - 200 Concurrent Users",
			concurrent:     200,
			requestsPerUser: 100,
			endpoint:       "/api/v1/merchants",
			method:         "GET",
			description:    "Stress test with 200 concurrent users, 100 requests each",
		},
	}

	for _, scenario := range loadScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			result := suite.runLoadTest(t, scenario.endpoint, scenario.method, scenario.concurrent, scenario.requestsPerUser)

			// Report results
			t.Logf("Load Test Results for %s:", scenario.name)
			t.Logf("  Total Requests: %d", result.TotalRequests)
			t.Logf("  Successful: %d", result.SuccessfulRequests)
			t.Logf("  Failed: %d", result.FailedRequests)
			t.Logf("  Error Rate: %.2f%%", result.ErrorRate)
			t.Logf("  Throughput: %.2f req/s", result.Throughput)
			t.Logf("  Min Response Time: %v", result.MinResponseTime)
			t.Logf("  Max Response Time: %v", result.MaxResponseTime)
			t.Logf("  Avg Response Time: %v", result.AvgResponseTime)
			t.Logf("  P50 Response Time: %v", result.P50ResponseTime)
			t.Logf("  P95 Response Time: %v", result.P95ResponseTime)
			t.Logf("  P99 Response Time: %v", result.P99ResponseTime)

			// Verify error rate is acceptable (< 1%)
			if result.ErrorRate > 1.0 {
				t.Errorf("Error rate (%.2f%%) exceeds acceptable threshold (1%%)", result.ErrorRate)
			} else {
				t.Logf("✅ Error rate (%.2f%%) is within acceptable range", result.ErrorRate)
			}

			// Verify throughput is reasonable
			if result.Throughput < 10 {
				t.Logf("⚠️  Throughput (%.2f req/s) is low - may indicate bottleneck", result.Throughput)
			} else {
				t.Logf("✅ Throughput (%.2f req/s) is acceptable", result.Throughput)
			}

			// Verify P95 response time is reasonable (< 2s under load)
			if result.P95ResponseTime > 2*time.Second {
				t.Logf("⚠️  P95 response time (%v) exceeds 2s - may indicate bottleneck", result.P95ResponseTime)
			} else {
				t.Logf("✅ P95 response time (%v) is within acceptable range", result.P95ResponseTime)
			}
		})
	}
}

// TestLoadDatabaseQueriesUnderLoad tests database queries under load
func TestLoadDatabaseQueriesUnderLoad(t *testing.T) {
	suite := SetupLoadTestSuite(t)

	// Test endpoints that hit the database
	databaseEndpoints := []struct {
		name     string
		endpoint string
		method   string
	}{
		{
			name:     "Get All Merchants",
			endpoint: "/api/v1/merchants",
			method:   "GET",
		},
		{
			name:     "Get Portfolio Statistics",
			endpoint: "/api/v1/merchants/statistics",
			method:   "GET",
		},
		{
			name:     "Get Portfolio Analytics",
			endpoint: "/api/v1/merchants/analytics",
			method:   "GET",
		},
		{
			name:     "Get Risk Trends",
			endpoint: "/api/v1/analytics/trends?timeframe=30d",
			method:   "GET",
		},
	}

	for _, endpoint := range databaseEndpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			// Run load test with 50 concurrent users
			result := suite.runLoadTest(t, endpoint.endpoint, endpoint.method, 50, 20)

			t.Logf("Database Load Test Results for %s:", endpoint.name)
			t.Logf("  Total Requests: %d", result.TotalRequests)
			t.Logf("  Successful: %d", result.SuccessfulRequests)
			t.Logf("  Failed: %d", result.FailedRequests)
			t.Logf("  Error Rate: %.2f%%", result.ErrorRate)
			t.Logf("  Throughput: %.2f req/s", result.Throughput)
			t.Logf("  P95 Response Time: %v", result.P95ResponseTime)

			// Verify error rate is acceptable
			if result.ErrorRate > 1.0 {
				t.Errorf("Error rate (%.2f%%) exceeds acceptable threshold (1%%)", result.ErrorRate)
			}

			// Check for potential database bottlenecks
			if result.P95ResponseTime > 1*time.Second {
				t.Logf("⚠️  WARNING: P95 response time (%v) > 1s - database may be a bottleneck", result.P95ResponseTime)
			}
		})
	}
}

// TestLoadIdentifyBottlenecks tests to identify bottlenecks under load
func TestLoadIdentifyBottlenecks(t *testing.T) {
	suite := SetupLoadTestSuite(t)

	// Test different endpoints to identify bottlenecks
	endpoints := []struct {
		name     string
		endpoint string
		method   string
	}{
		{name: "Health Check", endpoint: "/health", method: "GET"},
		{name: "Get Merchants", endpoint: "/api/v1/merchants", method: "GET"},
		{name: "Get Statistics", endpoint: "/api/v1/merchants/statistics", method: "GET"},
		{name: "Get Analytics", endpoint: "/api/v1/merchants/analytics", method: "GET"},
		{name: "Get Risk Trends", endpoint: "/api/v1/analytics/trends?timeframe=30d", method: "GET"},
	}

	// Run load test with increasing concurrency
	concurrencyLevels := []int{10, 25, 50, 100, 200}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			t.Logf("Bottleneck Analysis for %s:", endpoint.name)

			var previousThroughput float64
			var previousP95 time.Duration

			for _, concurrent := range concurrencyLevels {
				result := suite.runLoadTest(t, endpoint.endpoint, endpoint.method, concurrent, 10)

				t.Logf("  Concurrency %d:", concurrent)
				t.Logf("    Throughput: %.2f req/s", result.Throughput)
				t.Logf("    P95: %v", result.P95ResponseTime)
				t.Logf("    Error Rate: %.2f%%", result.ErrorRate)

				// Check for throughput degradation
				if previousThroughput > 0 {
					throughputChange := (result.Throughput - previousThroughput) / previousThroughput * 100
					if throughputChange < -20 {
						t.Logf("    ⚠️  Throughput decreased by %.2f%% - potential bottleneck at concurrency %d", -throughputChange, concurrent)
					}
				}

				// Check for response time degradation
				if previousP95 > 0 {
					responseTimeChange := float64(result.P95ResponseTime-previousP95) / float64(previousP95) * 100
					if responseTimeChange > 50 {
						t.Logf("    ⚠️  P95 response time increased by %.2f%% - potential bottleneck at concurrency %d", responseTimeChange, concurrent)
					}
				}

				previousThroughput = result.Throughput
				previousP95 = result.P95ResponseTime

				// Stop if error rate is too high
				if result.ErrorRate > 5.0 {
					t.Logf("    ⚠️  Error rate too high (%.2f%%) - stopping bottleneck analysis", result.ErrorRate)
					break
				}
			}
		})
	}
}

// runLoadTest runs a load test with specified parameters
func (suite *LoadTestSuite) runLoadTest(t *testing.T, endpoint, method string, concurrentUsers, requestsPerUser int) LoadTestResult {
	var (
		totalRequests      int64
		successfulRequests int64
		failedRequests     int64
		responseTimes      []time.Duration
		responseTimesMu    sync.Mutex
		wg                 sync.WaitGroup
	)

	startTime := time.Now()

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, concurrentUsers)

	// Launch concurrent users
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(userID int) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Each user makes multiple requests
			for j := 0; j < requestsPerUser; j++ {
				requestStart := time.Now()

				// Create request
				req := httptest.NewRequest(method, endpoint, nil)
				req.Header.Set("Content-Type", "application/json")

				// Create response recorder
				rr := httptest.NewRecorder()

				// Execute request
				suite.router.ServeHTTP(rr, req)

				responseTime := time.Since(requestStart)

				// Record metrics
				atomic.AddInt64(&totalRequests, 1)

				if rr.Code >= 200 && rr.Code < 300 {
					atomic.AddInt64(&successfulRequests, 1)
				} else {
					atomic.AddInt64(&failedRequests, 1)
				}

				// Record response time
				responseTimesMu.Lock()
				responseTimes = append(responseTimes, responseTime)
				responseTimesMu.Unlock()
			}
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()

	totalDuration := time.Since(startTime)

	// Calculate statistics
	return calculateLoadTestStats(responseTimes, totalRequests, successfulRequests, failedRequests, totalDuration, concurrentUsers, endpoint)
}

// calculateLoadTestStats calculates statistics from load test results
func calculateLoadTestStats(responseTimes []time.Duration, totalRequests, successfulRequests, failedRequests int64, totalDuration time.Duration, concurrentUsers int, endpoint string) LoadTestResult {
	if len(responseTimes) == 0 {
		return LoadTestResult{
			Endpoint:        endpoint,
			ConcurrentUsers: concurrentUsers,
		}
	}

	// Sort response times for percentile calculations
	sorted := make([]time.Duration, len(responseTimes))
	copy(sorted, responseTimes)
	sortDurations(sorted)

	// Calculate sum
	var sum time.Duration
	for _, rt := range responseTimes {
		sum += rt
	}

	// Calculate percentiles
	var p50, p95, p99 time.Duration
	if len(sorted) > 0 {
		p50 = sorted[int(float64(len(sorted))*0.50)]
		p95 = sorted[int(float64(len(sorted))*0.95)]
		if len(sorted) > 1 {
			p99 = sorted[int(float64(len(sorted))*0.99)]
		} else {
			p99 = sorted[0]
		}
	}

	throughput := float64(totalRequests) / totalDuration.Seconds()
	errorRate := float64(failedRequests) / float64(totalRequests) * 100

	return LoadTestResult{
		TotalRequests:      totalRequests,
		SuccessfulRequests: successfulRequests,
		FailedRequests:     failedRequests,
		TotalDuration:      totalDuration,
		MinResponseTime:    sorted[0],
		MaxResponseTime:    sorted[len(sorted)-1],
		AvgResponseTime:    sum / time.Duration(len(responseTimes)),
		P50ResponseTime:    p50,
		P95ResponseTime:    p95,
		P99ResponseTime:    p99,
		Throughput:         throughput,
		ErrorRate:          errorRate,
		ConcurrentUsers:    concurrentUsers,
		Endpoint:           endpoint,
	}
}

// sortDurations sorts a slice of durations
func sortDurations(durations []time.Duration) {
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
}

