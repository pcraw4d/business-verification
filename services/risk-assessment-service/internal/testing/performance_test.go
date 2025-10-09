package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/config"
	"kyb-platform/services/risk-assessment-service/internal/engine"
	"kyb-platform/services/risk-assessment-service/internal/handlers"
	"kyb-platform/services/risk-assessment-service/internal/ml/service"
	"kyb-platform/services/risk-assessment-service/internal/models"
	"kyb-platform/services/risk-assessment-service/internal/supabase"
)

// PerformanceTestSuite provides comprehensive performance testing
type PerformanceTestSuite struct {
	server     *httptest.Server
	riskEngine *engine.RiskEngine
	logger     *zap.Logger
}

// NewPerformanceTestSuite creates a new performance test suite
func NewPerformanceTestSuite() *PerformanceTestSuite {
	logger, _ := zap.NewDevelopment()

	// Initialize ML service
	mlService := service.NewMLService(logger)
	mlService.InitializeModels(context.Background())

	// Initialize risk engine with performance-optimized config
	riskEngineConfig := &engine.Config{
		MaxConcurrentRequests: 200,
		RequestTimeout:        500 * time.Millisecond, // Sub-1-second target
		CacheTTL:              5 * time.Minute,
		CircuitBreakerConfig: engine.CircuitBreakerConfig{
			FailureThreshold: 10,
			RecoveryTimeout:  30 * time.Second,
			HalfOpenMaxCalls: 5,
		},
		EnableMetrics: true,
		EnableCaching: true,
	}
	riskEngine := engine.NewRiskEngine(mlService, logger, riskEngineConfig)

	// Create mock Supabase client
	supabaseClient := &supabase.Client{}

	// Create config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:         "8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}

	// Initialize handlers
	riskAssessmentHandler := handlers.NewRiskAssessmentHandler(supabaseClient, mlService, riskEngine, logger, cfg)
	metricsHandler := handlers.NewMetricsHandler(riskEngine, logger)

	// Create test server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/assess", riskAssessmentHandler.HandleRiskAssessment)
	mux.HandleFunc("/api/v1/assess/batch", riskAssessmentHandler.HandleBatchRiskAssessment)
	mux.HandleFunc("/api/v1/metrics", metricsHandler.HandleMetrics)
	mux.HandleFunc("/api/v1/health", metricsHandler.HandleHealth)
	mux.HandleFunc("/api/v1/performance", metricsHandler.HandlePerformance)

	server := httptest.NewServer(mux)

	return &PerformanceTestSuite{
		server:     server,
		riskEngine: riskEngine,
		logger:     logger,
	}
}

// Close closes the test suite
func (pts *PerformanceTestSuite) Close() {
	pts.server.Close()
	pts.riskEngine.Shutdown(context.Background())
}

// TestSingleRequestPerformance tests single request performance
func (pts *PerformanceTestSuite) TestSingleRequestPerformance(t *testing.T) {
	request := models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	requestBody, _ := json.Marshal(request)

	// Test multiple requests to get average performance
	iterations := 100
	var totalDuration time.Duration
	var successCount int

	for i := 0; i < iterations; i++ {
		start := time.Now()

		resp, err := http.Post(pts.server.URL+"/api/v1/assess", "application/json", bytes.NewBuffer(requestBody))
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Request failed: %v", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			successCount++
			totalDuration += duration
		}

		resp.Body.Close()
	}

	avgDuration := totalDuration / time.Duration(successCount)
	successRate := float64(successCount) / float64(iterations)

	t.Logf("Single Request Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Success Count: %d", successCount)
	t.Logf("  Success Rate: %.2f%%", successRate*100)
	t.Logf("  Average Duration: %v", avgDuration)
	t.Logf("  Average Duration (ms): %.2f", float64(avgDuration.Nanoseconds())/1e6)

	// Assert sub-1-second response time
	if avgDuration > 1*time.Second {
		t.Errorf("Average response time %v exceeds 1 second target", avgDuration)
	}

	// Assert high success rate
	if successRate < 0.95 {
		t.Errorf("Success rate %.2f%% below 95%% target", successRate*100)
	}
}

// TestConcurrentRequestPerformance tests concurrent request performance
func (pts *PerformanceTestSuite) TestConcurrentRequestPerformance(t *testing.T) {
	request := models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	requestBody, _ := json.Marshal(request)

	// Test concurrent requests
	concurrency := 50
	requestsPerGoroutine := 10
	totalRequests := concurrency * requestsPerGoroutine

	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalDuration time.Duration
	var successCount int
	var maxDuration time.Duration

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerGoroutine; j++ {
				reqStart := time.Now()

				resp, err := http.Post(pts.server.URL+"/api/v1/assess", "application/json", bytes.NewBuffer(requestBody))
				reqDuration := time.Since(reqStart)

				mu.Lock()
				if err == nil && resp.StatusCode == http.StatusOK {
					successCount++
					totalDuration += reqDuration
					if reqDuration > maxDuration {
						maxDuration = reqDuration
					}
				}
				mu.Unlock()

				if resp != nil {
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()
	totalTime := time.Since(start)

	avgDuration := totalDuration / time.Duration(successCount)
	successRate := float64(successCount) / float64(totalRequests)
	requestsPerSecond := float64(totalRequests) / totalTime.Seconds()

	t.Logf("Concurrent Request Performance:")
	t.Logf("  Concurrency: %d", concurrency)
	t.Logf("  Total Requests: %d", totalRequests)
	t.Logf("  Success Count: %d", successCount)
	t.Logf("  Success Rate: %.2f%%", successRate*100)
	t.Logf("  Average Duration: %v", avgDuration)
	t.Logf("  Max Duration: %v", maxDuration)
	t.Logf("  Total Time: %v", totalTime)
	t.Logf("  Requests/Second: %.2f", requestsPerSecond)

	// Assert sub-1-second average response time
	if avgDuration > 1*time.Second {
		t.Errorf("Average response time %v exceeds 1 second target", avgDuration)
	}

	// Assert sub-1-second max response time
	if maxDuration > 1*time.Second {
		t.Errorf("Max response time %v exceeds 1 second target", maxDuration)
	}

	// Assert high success rate
	if successRate < 0.95 {
		t.Errorf("Success rate %.2f%% below 95%% target", successRate*100)
	}

	// Assert reasonable throughput
	if requestsPerSecond < 50 {
		t.Errorf("Throughput %.2f requests/second below 50 target", requestsPerSecond)
	}
}

// TestBatchRequestPerformance tests batch request performance
func (pts *PerformanceTestSuite) TestBatchRequestPerformance(t *testing.T) {
	// Create batch request
	var requests []models.RiskAssessmentRequest
	for i := 0; i < 10; i++ {
		requests = append(requests, models.RiskAssessmentRequest{
			BusinessName:      fmt.Sprintf("Test Company %d", i),
			BusinessAddress:   fmt.Sprintf("123 Test St %d, Test City, TC 12345", i),
			Industry:          "Technology",
			Country:           "US",
			PredictionHorizon: 3,
		})
	}

	batchRequest := struct {
		Requests []models.RiskAssessmentRequest `json:"requests"`
	}{
		Requests: requests,
	}

	requestBody, _ := json.Marshal(batchRequest)

	// Test batch requests
	iterations := 20
	var totalDuration time.Duration
	var successCount int

	for i := 0; i < iterations; i++ {
		start := time.Now()

		resp, err := http.Post(pts.server.URL+"/api/v1/assess/batch", "application/json", bytes.NewBuffer(requestBody))
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Batch request failed: %v", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			successCount++
			totalDuration += duration
		}

		resp.Body.Close()
	}

	avgDuration := totalDuration / time.Duration(successCount)
	successRate := float64(successCount) / float64(iterations)

	t.Logf("Batch Request Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Batch Size: %d", len(requests))
	t.Logf("  Success Count: %d", successCount)
	t.Logf("  Success Rate: %.2f%%", successRate*100)
	t.Logf("  Average Duration: %v", avgDuration)
	t.Logf("  Average Duration (ms): %.2f", float64(avgDuration.Nanoseconds())/1e6)

	// Assert sub-1-second response time for batch
	if avgDuration > 1*time.Second {
		t.Errorf("Average batch response time %v exceeds 1 second target", avgDuration)
	}

	// Assert high success rate
	if successRate < 0.95 {
		t.Errorf("Batch success rate %.2f%% below 95%% target", successRate*100)
	}
}

// TestCachePerformance tests cache performance
func (pts *PerformanceTestSuite) TestCachePerformance(t *testing.T) {
	request := models.RiskAssessmentRequest{
		BusinessName:      "Cache Test Company",
		BusinessAddress:   "123 Cache St, Cache City, CC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	requestBody, _ := json.Marshal(request)

	// First request (cache miss)
	start1 := time.Now()
	resp1, err1 := http.Post(pts.server.URL+"/api/v1/assess", "application/json", bytes.NewBuffer(requestBody))
	duration1 := time.Since(start1)
	if resp1 != nil {
		resp1.Body.Close()
	}

	// Second request (cache hit)
	start2 := time.Now()
	resp2, err2 := http.Post(pts.server.URL+"/api/v1/assess", "application/json", bytes.NewBuffer(requestBody))
	duration2 := time.Since(start2)
	if resp2 != nil {
		resp2.Body.Close()
	}

	if err1 != nil || err2 != nil {
		t.Errorf("Cache test requests failed: %v, %v", err1, err2)
		return
	}

	cacheSpeedup := float64(duration1.Nanoseconds()) / float64(duration2.Nanoseconds())

	t.Logf("Cache Performance:")
	t.Logf("  First Request (Cache Miss): %v", duration1)
	t.Logf("  Second Request (Cache Hit): %v", duration2)
	t.Logf("  Cache Speedup: %.2fx", cacheSpeedup)

	// Assert cache hit is faster
	if duration2 >= duration1 {
		t.Errorf("Cache hit duration %v not faster than cache miss %v", duration2, duration1)
	}

	// Assert significant speedup
	if cacheSpeedup < 2.0 {
		t.Errorf("Cache speedup %.2fx below 2x target", cacheSpeedup)
	}
}

// TestMetricsEndpoint tests metrics endpoint performance
func (pts *PerformanceTestSuite) TestMetricsEndpoint(t *testing.T) {
	// Make some requests to generate metrics
	request := models.RiskAssessmentRequest{
		BusinessName:      "Metrics Test Company",
		BusinessAddress:   "123 Metrics St, Metrics City, MC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	requestBody, _ := json.Marshal(request)

	// Generate some load
	for i := 0; i < 10; i++ {
		resp, _ := http.Post(pts.server.URL+"/api/v1/assess", "application/json", bytes.NewBuffer(requestBody))
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Test metrics endpoint
	start := time.Now()
	resp, err := http.Get(pts.server.URL + "/api/v1/metrics")
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Metrics request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Metrics request returned status %d", resp.StatusCode)
		return
	}

	t.Logf("Metrics Endpoint Performance:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Duration (ms): %.2f", float64(duration.Nanoseconds())/1e6)

	// Assert metrics endpoint is fast
	if duration > 100*time.Millisecond {
		t.Errorf("Metrics endpoint duration %v exceeds 100ms target", duration)
	}
}

// RunAllPerformanceTests runs all performance tests
func (pts *PerformanceTestSuite) RunAllPerformanceTests(t *testing.T) {
	t.Run("SingleRequest", pts.TestSingleRequestPerformance)
	t.Run("ConcurrentRequest", pts.TestConcurrentRequestPerformance)
	t.Run("BatchRequest", pts.TestBatchRequestPerformance)
	t.Run("Cache", pts.TestCachePerformance)
	t.Run("Metrics", pts.TestMetricsEndpoint)
}
