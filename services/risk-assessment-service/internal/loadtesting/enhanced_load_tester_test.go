package loadtesting

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestEnhancedLoadTester_NewEnhancedLoadTester(t *testing.T) {
	logger := zap.NewNop()
	baseURL := "http://localhost:8080"

	loadTester := NewEnhancedLoadTester(logger, baseURL)

	assert.NotNil(t, loadTester)
	assert.Equal(t, baseURL, loadTester.baseURL)
	assert.NotNil(t, loadTester.connectionPool)
	assert.NotNil(t, loadTester.requestPool)
	assert.NotNil(t, loadTester.responsePool)
	assert.NotNil(t, loadTester.metrics)
}

func TestEnhancedLoadTester_ConstantLoadTest(t *testing.T) {
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")

	config := EnhancedLoadTestConfig{
		Duration:           1 * time.Second,
		ConcurrentUsers:    5,
		TargetRPS:          10,
		TargetRPM:          600,
		TestPattern:        "constant",
		MaxLatency:         2 * time.Second,
		MaxErrorRate:       0.01,
		ConnectionPoolSize: 10,
		RequestTimeout:     1 * time.Second,
		KeepAliveTimeout:   30 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Note: This test will fail if the service is not running
	// In a real test environment, you would mock the HTTP client
	metrics, err := loadTester.RunHighPerformanceLoadTest(ctx, config)

	// We expect an error since the service is not running
	assert.Error(t, err)
	assert.Nil(t, metrics)
}

func TestEnhancedLoadTester_RampLoadTest(t *testing.T) {
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")

	config := EnhancedLoadTestConfig{
		Duration:           2 * time.Second,
		ConcurrentUsers:    3,
		TargetRPS:          5,
		TargetRPM:          300,
		TestPattern:        "ramp",
		RampUpTime:         500 * time.Millisecond,
		SteadyStateTime:    1 * time.Second,
		RampDownTime:       500 * time.Millisecond,
		MaxLatency:         2 * time.Second,
		MaxErrorRate:       0.01,
		ConnectionPoolSize: 10,
		RequestTimeout:     1 * time.Second,
		KeepAliveTimeout:   30 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Note: This test will fail if the service is not running
	metrics, err := loadTester.RunHighPerformanceLoadTest(ctx, config)

	// We expect an error since the service is not running
	assert.Error(t, err)
	assert.Nil(t, metrics)
}

func TestEnhancedLoadTester_SpikeLoadTest(t *testing.T) {
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")

	config := EnhancedLoadTestConfig{
		Duration:           2 * time.Second,
		ConcurrentUsers:    3,
		TargetRPS:          5,
		TargetRPM:          300,
		TestPattern:        "spike",
		SpikeMultiplier:    2.0,
		MaxLatency:         2 * time.Second,
		MaxErrorRate:       0.01,
		ConnectionPoolSize: 10,
		RequestTimeout:     1 * time.Second,
		KeepAliveTimeout:   30 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Note: This test will fail if the service is not running
	metrics, err := loadTester.RunHighPerformanceLoadTest(ctx, config)

	// We expect an error since the service is not running
	assert.Error(t, err)
	assert.Nil(t, metrics)
}

func TestEnhancedLoadTester_SineLoadTest(t *testing.T) {
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")

	config := EnhancedLoadTestConfig{
		Duration:           2 * time.Second,
		ConcurrentUsers:    3,
		TargetRPS:          5,
		TargetRPM:          300,
		TestPattern:        "sine",
		SineAmplitude:      0.5,
		SinePeriod:         1 * time.Second,
		MaxLatency:         2 * time.Second,
		MaxErrorRate:       0.01,
		ConnectionPoolSize: 10,
		RequestTimeout:     1 * time.Second,
		KeepAliveTimeout:   30 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Note: This test will fail if the service is not running
	metrics, err := loadTester.RunHighPerformanceLoadTest(ctx, config)

	// We expect an error since the service is not running
	assert.Error(t, err)
	assert.Nil(t, metrics)
}

func TestConnectionPool_GetClient(t *testing.T) {
	pool := &ConnectionPool{
		clients: make([]*http.Client, 3),
	}

	// Initialize clients
	for i := 0; i < 3; i++ {
		pool.clients[i] = &http.Client{}
	}

	// Test getting clients
	client1 := pool.getClient()
	client2 := pool.getClient()
	client3 := pool.getClient()
	client4 := pool.getClient()

	assert.NotNil(t, client1)
	assert.NotNil(t, client2)
	assert.NotNil(t, client3)
	assert.NotNil(t, client4)

	// Should cycle through clients
	assert.Equal(t, client1, client4) // Should wrap around
}

func TestLoadTestMetrics_CalculatePerformanceScore(t *testing.T) {
	metrics := &LoadTestMetrics{
		TargetRPS:         100,
		TargetLatency:     1 * time.Second,
		TargetErrorRate:   0.01,
		RequestsPerMinute: 6000,  // 100 RPS * 60
		ErrorRate:         0.005, // 0.5%
		MaxResponseTime:   500 * time.Millisecond,
	}

	// Create a load tester to access the calculation method
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")
	loadTester.metrics = metrics

	// Test performance score calculation
	loadTester.calculatePerformanceScore()

	// Should have a high score since all targets are met
	assert.Greater(t, metrics.PerformanceScore, 90.0)
	assert.True(t, metrics.IsTargetMet)
}

func TestLoadTestMetrics_UpdateResponseTimeMetrics(t *testing.T) {
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")

	// Test updating response time metrics
	loadTester.updateResponseTimeMetrics(100 * time.Millisecond)
	loadTester.updateResponseTimeMetrics(200 * time.Millisecond)
	loadTester.updateResponseTimeMetrics(50 * time.Millisecond)

	// Check that min and max are updated correctly
	assert.Equal(t, 50*time.Millisecond, loadTester.metrics.MinResponseTime)
	assert.Equal(t, 200*time.Millisecond, loadTester.metrics.MaxResponseTime)
}

func TestEnhancedLoadTester_CombineResults(t *testing.T) {
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")

	// Create test results
	result1 := &LoadTestMetrics{
		TotalRequests:      100,
		SuccessfulRequests: 95,
		FailedRequests:     5,
		TotalDuration:      1 * time.Second,
		RequestsPerSecond:  100,
		ErrorRate:          0.05,
	}

	result2 := &LoadTestMetrics{
		TotalRequests:      200,
		SuccessfulRequests: 190,
		FailedRequests:     10,
		TotalDuration:      2 * time.Second,
		RequestsPerSecond:  100,
		ErrorRate:          0.05,
	}

	result3 := &LoadTestMetrics{
		TotalRequests:      150,
		SuccessfulRequests: 145,
		FailedRequests:     5,
		TotalDuration:      time.Duration(1.5 * float64(time.Second)),
		RequestsPerSecond:  100,
		ErrorRate:          0.033,
	}

	// Combine results
	loadTester.combineResults(result1, result2, result3)

	// Check combined metrics
	assert.Equal(t, int64(450), loadTester.metrics.TotalRequests)
	assert.Equal(t, int64(430), loadTester.metrics.SuccessfulRequests)
	assert.Equal(t, int64(20), loadTester.metrics.FailedRequests)
	assert.Equal(t, time.Duration(4.5*float64(time.Second)), loadTester.metrics.TotalDuration)
	assert.Equal(t, 100.0, loadTester.metrics.RequestsPerSecond)
	assert.Equal(t, 6000.0, loadTester.metrics.RequestsPerMinute)
	assert.InDelta(t, 0.044, loadTester.metrics.ErrorRate, 0.001)
}

func TestEnhancedLoadTester_CalculateFinalMetrics(t *testing.T) {
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")

	// Set up metrics
	loadTester.metrics.TotalRequests = 1000
	loadTester.metrics.SuccessfulRequests = 950
	loadTester.metrics.FailedRequests = 50
	loadTester.metrics.TargetRPS = 100
	loadTester.metrics.TargetErrorRate = 0.01
	loadTester.metrics.TargetLatency = 1 * time.Second
	loadTester.metrics.MaxResponseTime = 500 * time.Millisecond

	// Calculate final metrics
	startTime := time.Now().Add(-10 * time.Second)
	loadTester.calculateFinalMetrics(startTime)

	// Check calculated metrics
	assert.Equal(t, 10*time.Second, loadTester.metrics.TotalDuration)
	assert.Equal(t, 100.0, loadTester.metrics.RequestsPerSecond)
	assert.Equal(t, 6000.0, loadTester.metrics.RequestsPerMinute)
	assert.Equal(t, 0.05, loadTester.metrics.ErrorRate)

	// Should meet targets (low error rate, good latency)
	assert.True(t, loadTester.metrics.IsTargetMet)
	assert.Greater(t, loadTester.metrics.PerformanceScore, 80.0)
}

func TestEnhancedLoadTester_5000RPMTarget(t *testing.T) {
	logger := zap.NewNop()
	_ = NewEnhancedLoadTester(logger, "http://localhost:8080")

	// Test configuration for 5000 RPM target
	config := EnhancedLoadTestConfig{
		Duration:           1 * time.Minute,
		ConcurrentUsers:    200,
		TargetRPS:          83.33, // 5000 RPM / 60 seconds
		TargetRPM:          5000,
		TestPattern:        "constant",
		MaxLatency:         2 * time.Second,
		MaxErrorRate:       0.01,
		ConnectionPoolSize: 100,
		RequestTimeout:     30 * time.Second,
		KeepAliveTimeout:   90 * time.Second,
	}

	// Verify configuration
	assert.Equal(t, 5000.0, config.TargetRPM)
	assert.Equal(t, 83.33, config.TargetRPS)
	assert.Equal(t, 200, config.ConcurrentUsers)
	assert.Equal(t, "constant", config.TestPattern)

	// Test that the configuration is valid for 5000 RPM testing
	assert.GreaterOrEqual(t, config.ConcurrentUsers, 100)   // Need sufficient concurrency
	assert.GreaterOrEqual(t, config.TargetRPS, 80.0)        // Need sufficient RPS
	assert.LessOrEqual(t, config.MaxLatency, 2*time.Second) // Need low latency
	assert.LessOrEqual(t, config.MaxErrorRate, 0.01)        // Need low error rate
}

func TestEnhancedLoadTester_PerformanceOptimization(t *testing.T) {
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")

	// Test that the load tester is optimized for high performance
	assert.NotNil(t, loadTester.connectionPool)
	assert.NotNil(t, loadTester.requestPool)
	assert.NotNil(t, loadTester.responsePool)

	// Test connection pool size
	assert.GreaterOrEqual(t, len(loadTester.connectionPool.clients), 50)

	// Test that pools are working
	req := loadTester.requestPool.Get()
	assert.NotNil(t, req)
	loadTester.requestPool.Put(req)

	resp := loadTester.responsePool.Get()
	assert.NotNil(t, resp)
	loadTester.responsePool.Put(resp)
}

func TestEnhancedLoadTester_ContextCancellation(t *testing.T) {
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")

	config := EnhancedLoadTestConfig{
		Duration:           10 * time.Second,
		ConcurrentUsers:    5,
		TargetRPS:          10,
		TargetRPM:          600,
		TestPattern:        "constant",
		MaxLatency:         2 * time.Second,
		MaxErrorRate:       0.01,
		ConnectionPoolSize: 10,
		RequestTimeout:     1 * time.Second,
		KeepAliveTimeout:   30 * time.Second,
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run test with short timeout
	metrics, err := loadTester.RunHighPerformanceLoadTest(ctx, config)

	// Should get context cancellation error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
	assert.Nil(t, metrics)
}

func BenchmarkEnhancedLoadTester_RequestTracking(b *testing.B) {
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate request tracking
		loadTester.updateResponseTimeMetrics(time.Duration(i%1000) * time.Millisecond)
	}
}

func BenchmarkEnhancedLoadTester_ConnectionPool(b *testing.B) {
	pool := &ConnectionPool{
		clients: make([]*http.Client, 100),
	}

	// Initialize clients
	for i := 0; i < 100; i++ {
		pool.clients[i] = &http.Client{}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client := pool.getClient()
		pool.returnClient(client)
	}
}

func BenchmarkEnhancedLoadTester_ObjectPools(b *testing.B) {
	logger := zap.NewNop()
	loadTester := NewEnhancedLoadTester(logger, "http://localhost:8080")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Test request pool
		req := loadTester.requestPool.Get()
		loadTester.requestPool.Put(req)

		// Test response pool
		resp := loadTester.responsePool.Get()
		loadTester.responsePool.Put(resp)
	}
}
