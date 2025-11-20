package test

import (
	"fmt"
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/api-gateway/internal/config"
	"kyb-platform/services/api-gateway/internal/handlers"
	"kyb-platform/services/api-gateway/internal/middleware"
	"kyb-platform/services/api-gateway/internal/supabase"
)

// PerformanceTestSuite provides performance testing for API Gateway
type PerformanceTestSuite struct {
	router         *mux.Router
	gatewayHandler *handlers.GatewayHandler
	config         *config.Config
	logger         *zap.Logger
	baseURL        string
}

// SetupPerformanceTestSuite creates a test suite for performance testing
func SetupPerformanceTestSuite(t *testing.T) *PerformanceTestSuite {
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
				Enabled: false, // Disable rate limiting in tests
			},
		}
	}

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize Supabase client (optional for performance tests)
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

	// Register routes (simplified for performance tests)
	setupPerformanceRoutes(router, gatewayHandler, cfg, logger, supabaseClient)

	baseURL := os.Getenv("API_GATEWAY_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &PerformanceTestSuite{
		router:         router,
		gatewayHandler: gatewayHandler,
		config:         cfg,
		logger:         logger,
		baseURL:        baseURL,
	}
}

// setupPerformanceRoutes registers essential routes for performance testing
func setupPerformanceRoutes(router *mux.Router, gatewayHandler *handlers.GatewayHandler, cfg *config.Config, logger *zap.Logger, supabaseClient *supabase.Client) {
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

// ResponseTimeStats holds statistics about response times
type ResponseTimeStats struct {
	Min       time.Duration
	Max       time.Duration
	Mean      time.Duration
	Median    time.Duration
	P50       time.Duration
	P95       time.Duration
	P99       time.Duration
	Total     time.Duration
	Count     int
	Successes int
	Failures  int
	Errors    []error
}

// CalculateStats calculates statistics from a slice of durations
func CalculateStats(durations []time.Duration, errors []error) ResponseTimeStats {
	if len(durations) == 0 {
		return ResponseTimeStats{}
	}

	// Sort durations for percentile calculations
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate sum
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}

	stats := ResponseTimeStats{
		Min:       sorted[0],
		Max:       sorted[len(sorted)-1],
		Mean:      sum / time.Duration(len(durations)),
		Median:    sorted[len(sorted)/2],
		Count:     len(durations),
		Successes: len(durations) - len(errors),
		Failures:  len(errors),
		Errors:    errors,
	}

	// Calculate percentiles
	if len(sorted) > 0 {
		stats.P50 = sorted[int(float64(len(sorted))*0.50)]
		stats.P95 = sorted[int(float64(len(sorted))*0.95)]
		stats.P99 = sorted[int(float64(len(sorted))*0.99)]
	}

	return stats
}

// TestPerformanceAPIResponseTimes tests API response times for all endpoints
func TestPerformanceAPIResponseTimes(t *testing.T) {
	suite := SetupPerformanceTestSuite(t)

	// Test endpoints with their expected max response times
	testCases := []struct {
		name           string
		method         string
		path           string
		queryParams    map[string]string
		maxP95         time.Duration // Maximum p95 response time
		iterations     int           // Number of requests to make
		concurrent     int           // Number of concurrent requests
		description    string
	}{
		{
			name:        "Health Check",
			method:      "GET",
			path:        "/health",
			maxP95:      100 * time.Millisecond,
			iterations:  100,
			concurrent:  10,
			description: "Health check should be very fast (< 100ms p95)",
		},
		{
			name:        "Get All Merchants",
			method:      "GET",
			path:        "/api/v1/merchants",
			maxP95:      500 * time.Millisecond,
			iterations:  50,
			concurrent:  5,
			description: "Get all merchants should be fast (< 500ms p95)",
		},
		{
			name:        "Get Merchant by ID",
			method:      "GET",
			path:        "/api/v1/merchants/merchant-123",
			maxP95:      500 * time.Millisecond,
			iterations:  50,
			concurrent:  5,
			description: "Get merchant by ID should be fast (< 500ms p95)",
		},
		{
			name:        "Get Portfolio Analytics",
			method:      "GET",
			path:        "/api/v1/merchants/analytics",
			maxP95:      500 * time.Millisecond,
			iterations:  30,
			concurrent:  3,
			description: "Get portfolio analytics should be fast (< 500ms p95)",
		},
		{
			name:        "Get Portfolio Statistics",
			method:      "GET",
			path:        "/api/v1/merchants/statistics",
			maxP95:      500 * time.Millisecond,
			iterations:  30,
			concurrent:  3,
			description: "Get portfolio statistics should be fast (< 500ms p95)",
		},
		{
			name:        "Get Risk Trends",
			method:      "GET",
			path:        "/api/v1/analytics/trends",
			queryParams: map[string]string{"timeframe": "30d"},
			maxP95:      500 * time.Millisecond,
			iterations:  30,
			concurrent:  3,
			description: "Get risk trends should be fast (< 500ms p95)",
		},
		{
			name:        "Get Risk Insights",
			method:      "GET",
			path:        "/api/v1/analytics/insights",
			queryParams: map[string]string{"timeframe": "30d"},
			maxP95:      500 * time.Millisecond,
			iterations:  30,
			concurrent:  3,
			description: "Get risk insights should be fast (< 500ms p95)",
		},
		{
			name:        "Get Risk Benchmarks",
			method:      "GET",
			path:        "/api/v1/risk/benchmarks",
			queryParams: map[string]string{"industry": "Technology"},
			maxP95:      500 * time.Millisecond,
			iterations:  30,
			concurrent:  3,
			description: "Get risk benchmarks should be fast (< 500ms p95)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stats := suite.measureResponseTimes(t, tc.method, tc.path, tc.queryParams, tc.iterations, tc.concurrent)

			// Report statistics
			t.Logf("Response Time Statistics for %s:", tc.name)
			t.Logf("  Min: %v", stats.Min)
			t.Logf("  Max: %v", stats.Max)
			t.Logf("  Mean: %v", stats.Mean)
			t.Logf("  Median: %v", stats.Median)
			t.Logf("  P50: %v", stats.P50)
			t.Logf("  P95: %v", stats.P95)
			t.Logf("  P99: %v", stats.P99)
			t.Logf("  Successes: %d", stats.Successes)
			t.Logf("  Failures: %d", stats.Failures)

			// Verify p95 is within acceptable range
			if stats.P95 > tc.maxP95 {
				t.Errorf("P95 response time (%v) exceeds maximum (%v) for %s", stats.P95, tc.maxP95, tc.name)
			} else {
				t.Logf("✅ P95 response time (%v) is within acceptable range (%v)", stats.P95, tc.maxP95)
			}

			// Verify success rate
			successRate := float64(stats.Successes) / float64(stats.Count) * 100
			if successRate < 95.0 {
				t.Errorf("Success rate (%.2f%%) is below 95%% for %s", successRate, tc.name)
			} else {
				t.Logf("✅ Success rate: %.2f%%", successRate)
			}
		})
	}
}

// TestPerformanceConcurrentRequests tests API Gateway under concurrent load
func TestPerformanceConcurrentRequests(t *testing.T) {
	suite := SetupPerformanceTestSuite(t)

	concurrencyLevels := []int{1, 5, 10, 20, 50}
	iterationsPerLevel := 100

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			stats := suite.measureResponseTimes(t, "GET", "/api/v1/merchants", nil, iterationsPerLevel, concurrency)

			t.Logf("Concurrency Level: %d", concurrency)
			t.Logf("  Mean Response Time: %v", stats.Mean)
			t.Logf("  P95 Response Time: %v", stats.P95)
			t.Logf("  Success Rate: %.2f%%", float64(stats.Successes)/float64(stats.Count)*100)

			// Verify p95 is still acceptable under load
			if stats.P95 > 1000*time.Millisecond {
				t.Errorf("P95 response time (%v) exceeds 1s under concurrency %d", stats.P95, concurrency)
			}
		})
	}
}

// TestPerformanceCachingEffectiveness tests caching effectiveness
func TestPerformanceCachingEffectiveness(t *testing.T) {
	suite := SetupPerformanceTestSuite(t)

	// Test endpoint that should benefit from caching
	endpoint := "/api/v1/merchants/statistics"
	iterations := 20

	// First request (cache miss)
	firstStats := suite.measureResponseTimes(t, "GET", endpoint, nil, 1, 1)
	firstRequestTime := firstStats.Mean

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Subsequent requests (cache hit)
	subsequentStats := suite.measureResponseTimes(t, "GET", endpoint, nil, iterations-1, 1)
	subsequentRequestTime := subsequentStats.Mean

	// Calculate cache effectiveness
	cacheSpeedup := float64(firstRequestTime) / float64(subsequentRequestTime)
	cacheImprovement := (1.0 - float64(subsequentRequestTime)/float64(firstRequestTime)) * 100

	t.Logf("Cache Effectiveness Test:")
	t.Logf("  First Request (Cache Miss): %v", firstRequestTime)
	t.Logf("  Subsequent Requests (Cache Hit): %v", subsequentRequestTime)
	t.Logf("  Cache Speedup: %.2fx", cacheSpeedup)
	t.Logf("  Cache Improvement: %.2f%%", cacheImprovement)

	// Verify caching provides improvement (if caching is enabled)
	if cacheSpeedup > 1.1 {
		t.Logf("✅ Caching is effective (%.2fx speedup)", cacheSpeedup)
	} else {
		t.Logf("⚠️  Caching may not be enabled or effective (speedup: %.2fx)", cacheSpeedup)
	}
}

// measureResponseTimes measures response times for a given endpoint
func (suite *PerformanceTestSuite) measureResponseTimes(t *testing.T, method, path string, queryParams map[string]string, iterations, concurrent int) ResponseTimeStats {
	// Build URL
	url := path
	if len(queryParams) > 0 {
		query := ""
		for key, value := range queryParams {
			if query != "" {
				query += "&"
			}
			query += fmt.Sprintf("%s=%s", key, value)
		}
		url += "?" + query
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	durations := make([]time.Duration, 0, iterations)
	errors := make([]error, 0)

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, concurrent)

	// Make requests
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func() {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			start := time.Now()

			// Create request
			req := httptest.NewRequest(method, url, nil)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			suite.router.ServeHTTP(rr, req)

			duration := time.Since(start)

			// Record duration and status
			mu.Lock()
			durations = append(durations, duration)
			if rr.Code >= 400 {
				errors = append(errors, fmt.Errorf("HTTP %d", rr.Code))
			}
			mu.Unlock()
		}()
	}

	// Wait for all requests to complete
	wg.Wait()

	return CalculateStats(durations, errors)
}

// TestPerformanceSlowQueries identifies slow queries
func TestPerformanceSlowQueries(t *testing.T) {
	suite := SetupPerformanceTestSuite(t)

	// Test endpoints that might be slow
	testCases := []struct {
		name        string
		method      string
		path        string
		queryParams map[string]string
		description string
	}{
		{
			name:        "Portfolio Analytics",
			method:      "GET",
			path:        "/api/v1/merchants/analytics",
			description: "Portfolio analytics may be slow due to aggregation",
		},
		{
			name:        "Risk Trends with Long Timeframe",
			method:      "GET",
			path:        "/api/v1/analytics/trends",
			queryParams: map[string]string{"timeframe": "1y"},
			description: "Risk trends with long timeframe may be slow",
		},
		{
			name:        "Risk Insights with Long Timeframe",
			method:      "GET",
			path:        "/api/v1/analytics/insights",
			queryParams: map[string]string{"timeframe": "1y"},
			description: "Risk insights with long timeframe may be slow",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stats := suite.measureResponseTimes(t, tc.method, tc.path, tc.queryParams, 10, 1)

			t.Logf("Slow Query Analysis for %s:", tc.name)
			t.Logf("  Mean: %v", stats.Mean)
			t.Logf("  P95: %v", stats.P95)
			t.Logf("  Max: %v", stats.Max)

			// Identify slow queries (> 1 second)
			if stats.P95 > 1*time.Second {
				t.Logf("⚠️  WARNING: %s has P95 > 1s (%v) - consider optimization", tc.name, stats.P95)
			} else if stats.P95 > 500*time.Millisecond {
				t.Logf("⚠️  INFO: %s has P95 > 500ms (%v) - monitor for optimization", tc.name, stats.P95)
			} else {
				t.Logf("✅ %s performance is acceptable (P95: %v)", tc.name, stats.P95)
			}
		})
	}
}

// BenchmarkAPIEndpoints provides Go benchmarks for API endpoints
func BenchmarkAPIEndpoints(b *testing.B) {
	suite := SetupPerformanceTestSuite(&testing.T{})

	benchmarks := []struct {
		name        string
		method      string
		path        string
		queryParams map[string]string
	}{
		{
			name:   "HealthCheck",
			method: "GET",
			path:   "/health",
		},
		{
			name:   "GetAllMerchants",
			method: "GET",
			path:   "/api/v1/merchants",
		},
		{
			name:   "GetMerchantByID",
			method: "GET",
			path:   "/api/v1/merchants/merchant-123",
		},
		{
			name:   "GetPortfolioAnalytics",
			method: "GET",
			path:   "/api/v1/merchants/analytics",
		},
		{
			name:   "GetPortfolioStatistics",
			method: "GET",
			path:   "/api/v1/merchants/statistics",
		},
		{
			name:        "GetRiskTrends",
			method:      "GET",
			path:        "/api/v1/analytics/trends",
			queryParams: map[string]string{"timeframe": "30d"},
		},
		{
			name:        "GetRiskInsights",
			method:      "GET",
			path:        "/api/v1/analytics/insights",
			queryParams: map[string]string{"timeframe": "30d"},
		},
		{
			name:        "GetRiskBenchmarks",
			method:      "GET",
			path:        "/api/v1/risk/benchmarks",
			queryParams: map[string]string{"industry": "Technology"},
		},
	}

	for _, bm := range benchmarks {
		// Build URL
		url := bm.path
		if len(bm.queryParams) > 0 {
			query := ""
			for key, value := range bm.queryParams {
				if query != "" {
					query += "&"
				}
				query += fmt.Sprintf("%s=%s", key, value)
			}
			url += "?" + query
		}

		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(bm.method, url, nil)
				req.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()

				suite.router.ServeHTTP(rr, req)

				if rr.Code >= 400 {
					b.Errorf("Request failed with status %d", rr.Code)
				}
			}
		})
	}
}

