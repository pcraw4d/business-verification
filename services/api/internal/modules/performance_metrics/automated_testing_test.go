package performance_metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewAutomatedTesting(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	config := DefaultAutomatedTestingConfig()

	testing := NewAutomatedTesting(logger, metrics, detector, strategies, config)

	assert.NotNil(t, testing)
	assert.Equal(t, logger, testing.logger)
	assert.Equal(t, metrics, testing.metrics)
	assert.Equal(t, detector, testing.detector)
	assert.Equal(t, strategies, testing.strategies)
	assert.Equal(t, config, testing.config)
	assert.NotNil(t, testing.tests)
	assert.NotNil(t, testing.suites)
	assert.NotNil(t, testing.stopChan)
}

func TestDefaultAutomatedTestingConfig(t *testing.T) {
	config := DefaultAutomatedTestingConfig()

	assert.True(t, config.EnableAutomatedTesting)
	assert.Equal(t, 1*time.Hour, config.TestInterval)
	assert.Equal(t, 30*24*time.Hour, config.RetentionPeriod)
	assert.Equal(t, 5, config.MaxConcurrentTests)
	assert.Equal(t, 30*time.Second, config.DefaultTimeout)
	assert.Equal(t, 0.1, config.BaselineThreshold)
	assert.Equal(t, 0.2, config.RegressionThreshold)
	assert.True(t, config.AutoGenerateTests)
}

func TestAutomatedTesting_Start_Stop(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	ctx := context.Background()

	// Test start
	err := testing.Start(ctx)
	require.NoError(t, err)

	// Test stop
	testing.Stop()
}

func TestAutomatedTesting_RunLoadTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	test := &PerformanceTest{
		ID:   "load_test",
		Name: "Load Test",
		Type: TestTypeLoad,
		Config: &TestConfig{
			Duration:       1 * time.Second, // Short duration for testing
			Concurrency:    5,
			RequestRate:    10,
			RampUpTime:     100 * time.Millisecond,
			RampDownTime:   100 * time.Millisecond,
			TargetEndpoint: "/api/test",
			Timeout:        5 * time.Second,
			RetryCount:     3,
			ThinkTime:      50 * time.Millisecond,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    1 * time.Second,
			MaxErrorRate:       0.1,
			MinThroughput:      5.0,
			MaxCPUUsage:        90.0,
			MaxMemoryUsage:     95.0,
			MinCacheHitRate:    0.5,
			MaxDatabaseQueries: 100,
		},
	}

	testMetrics := &TestMetrics{}
	err := testing.runLoadTest(context.Background(), test, testMetrics)

	require.NoError(t, err)
	assert.Greater(t, testMetrics.TotalRequests, int64(0))
	assert.Greater(t, testMetrics.SuccessfulRequests, int64(0))
	assert.Greater(t, testMetrics.AverageResponseTime, time.Duration(0))
	assert.Greater(t, testMetrics.Throughput, 0.0)
	assert.LessOrEqual(t, testMetrics.ErrorRate, 0.1)
}

func TestAutomatedTesting_RunStressTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	test := &PerformanceTest{
		ID:   "stress_test",
		Name: "Stress Test",
		Type: TestTypeStress,
		Config: &TestConfig{
			Duration:       1 * time.Second, // Short duration for testing
			Concurrency:    10,
			RequestRate:    20,
			RampUpTime:     100 * time.Millisecond,
			RampDownTime:   100 * time.Millisecond,
			TargetEndpoint: "/api/test",
			Timeout:        5 * time.Second,
			RetryCount:     3,
			ThinkTime:      25 * time.Millisecond,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    2 * time.Second,
			MaxErrorRate:       0.2,
			MinThroughput:      10.0,
			MaxCPUUsage:        95.0,
			MaxMemoryUsage:     98.0,
			MinCacheHitRate:    0.3,
			MaxDatabaseQueries: 200,
		},
	}

	testMetrics := &TestMetrics{}
	err := testing.runStressTest(context.Background(), test, testMetrics)

	require.NoError(t, err)
	assert.Greater(t, testMetrics.TotalRequests, int64(0))
	assert.Greater(t, testMetrics.SuccessfulRequests, int64(0))
	assert.Greater(t, testMetrics.AverageResponseTime, time.Duration(0))
	assert.Greater(t, testMetrics.Throughput, 0.0)
	assert.Greater(t, testMetrics.CPUUsage, 80.0)    // Should be high under stress
	assert.Greater(t, testMetrics.MemoryUsage, 90.0) // Should be high under stress
}

func TestAutomatedTesting_RunSpikeTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	test := &PerformanceTest{
		ID:   "spike_test",
		Name: "Spike Test",
		Type: TestTypeSpike,
		Config: &TestConfig{
			Duration:       1 * time.Second, // Short duration for testing
			Concurrency:    20,
			RequestRate:    50,
			RampUpTime:     50 * time.Millisecond,
			RampDownTime:   50 * time.Millisecond,
			TargetEndpoint: "/api/test",
			Timeout:        5 * time.Second,
			RetryCount:     3,
			ThinkTime:      10 * time.Millisecond,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    5 * time.Second,
			MaxErrorRate:       0.4,
			MinThroughput:      20.0,
			MaxCPUUsage:        99.0,
			MaxMemoryUsage:     99.0,
			MinCacheHitRate:    0.1,
			MaxDatabaseQueries: 500,
		},
	}

	testMetrics := &TestMetrics{}
	err := testing.runSpikeTest(context.Background(), test, testMetrics)

	require.NoError(t, err)
	assert.Greater(t, testMetrics.TotalRequests, int64(0))
	assert.Greater(t, testMetrics.SuccessfulRequests, int64(0))
	assert.Greater(t, testMetrics.AverageResponseTime, time.Duration(0))
	assert.Greater(t, testMetrics.Throughput, 0.0)
	assert.Greater(t, testMetrics.CPUUsage, 95.0)    // Should be very high during spike
	assert.Greater(t, testMetrics.MemoryUsage, 95.0) // Should be very high during spike
}

func TestAutomatedTesting_RunSoakTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	test := &PerformanceTest{
		ID:   "soak_test",
		Name: "Soak Test",
		Type: TestTypeSoak,
		Config: &TestConfig{
			Duration:       1 * time.Second, // Short duration for testing
			Concurrency:    8,
			RequestRate:    15,
			RampUpTime:     200 * time.Millisecond,
			RampDownTime:   200 * time.Millisecond,
			TargetEndpoint: "/api/test",
			Timeout:        5 * time.Second,
			RetryCount:     3,
			ThinkTime:      75 * time.Millisecond,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    1 * time.Second,
			MaxErrorRate:       0.05,
			MinThroughput:      10.0,
			MaxCPUUsage:        85.0,
			MaxMemoryUsage:     90.0,
			MinCacheHitRate:    0.8,
			MaxDatabaseQueries: 150,
		},
	}

	testMetrics := &TestMetrics{}
	err := testing.runSoakTest(context.Background(), test, testMetrics)

	require.NoError(t, err)
	assert.Greater(t, testMetrics.TotalRequests, int64(0))
	assert.Greater(t, testMetrics.SuccessfulRequests, int64(0))
	assert.Greater(t, testMetrics.AverageResponseTime, time.Duration(0))
	assert.Greater(t, testMetrics.Throughput, 0.0)
	assert.LessOrEqual(t, testMetrics.ErrorRate, 0.10) // Should be low for soak test
	assert.Greater(t, testMetrics.CacheHitRate, 0.8)   // Should be high for soak test
}

func TestAutomatedTesting_RunBaselineTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	test := &PerformanceTest{
		ID:   "baseline_test",
		Name: "Baseline Test",
		Type: TestTypeBaseline,
		Config: &TestConfig{
			Duration:       1 * time.Second, // Short duration for testing
			Concurrency:    5,
			RequestRate:    10,
			RampUpTime:     100 * time.Millisecond,
			RampDownTime:   100 * time.Millisecond,
			TargetEndpoint: "/api/test",
			Timeout:        5 * time.Second,
			RetryCount:     3,
			ThinkTime:      50 * time.Millisecond,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    1 * time.Second,
			MaxErrorRate:       0.1,
			MinThroughput:      5.0,
			MaxCPUUsage:        90.0,
			MaxMemoryUsage:     95.0,
			MinCacheHitRate:    0.5,
			MaxDatabaseQueries: 100,
		},
	}

	testMetrics := &TestMetrics{}
	err := testing.runBaselineTest(context.Background(), test, testMetrics)

	require.NoError(t, err)
	assert.Greater(t, testMetrics.TotalRequests, int64(0))
	assert.Greater(t, testMetrics.SuccessfulRequests, int64(0))
	assert.Greater(t, testMetrics.AverageResponseTime, time.Duration(0))
	assert.Greater(t, testMetrics.Throughput, 0.0)
}

func TestAutomatedTesting_RunRegressionTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	// Create baseline test first
	baselineTest := &PerformanceTest{
		ID:   "baseline_test",
		Name: "Baseline Test",
		Type: TestTypeBaseline,
		Config: &TestConfig{
			Duration:       1 * time.Second,
			Concurrency:    5,
			RequestRate:    10,
			RampUpTime:     100 * time.Millisecond,
			RampDownTime:   100 * time.Millisecond,
			TargetEndpoint: "/api/test",
			Timeout:        5 * time.Second,
			RetryCount:     3,
			ThinkTime:      50 * time.Millisecond,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    1 * time.Second,
			MaxErrorRate:       0.1,
			MinThroughput:      5.0,
			MaxCPUUsage:        90.0,
			MaxMemoryUsage:     95.0,
			MinCacheHitRate:    0.5,
			MaxDatabaseQueries: 100,
		},
	}

	// Run baseline test
	baselineMetrics := &TestMetrics{}
	err := testing.runBaselineTest(context.Background(), baselineTest, baselineMetrics)
	require.NoError(t, err)

	// Store baseline test
	testing.tests[baselineTest.ID] = baselineTest
	baselineTest.Metrics = baselineMetrics

	// Create regression test
	regressionTest := &PerformanceTest{
		ID:         "regression_test",
		Name:       "Regression Test",
		Type:       TestTypeRegression,
		BaselineID: baselineTest.ID,
		Config: &TestConfig{
			Duration:       1 * time.Second,
			Concurrency:    5,
			RequestRate:    10,
			RampUpTime:     100 * time.Millisecond,
			RampDownTime:   100 * time.Millisecond,
			TargetEndpoint: "/api/test",
			Timeout:        5 * time.Second,
			RetryCount:     3,
			ThinkTime:      50 * time.Millisecond,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    1 * time.Second,
			MaxErrorRate:       0.1,
			MinThroughput:      5.0,
			MaxCPUUsage:        90.0,
			MaxMemoryUsage:     95.0,
			MinCacheHitRate:    0.5,
			MaxDatabaseQueries: 100,
		},
	}

	regressionMetrics := &TestMetrics{}
	err = testing.runRegressionTest(context.Background(), regressionTest, regressionMetrics)

	require.NoError(t, err)
	assert.Greater(t, regressionMetrics.TotalRequests, int64(0))
	assert.Greater(t, regressionMetrics.SuccessfulRequests, int64(0))
	assert.Greater(t, regressionMetrics.AverageResponseTime, time.Duration(0))
	assert.Greater(t, regressionMetrics.Throughput, 0.0)
}

func TestAutomatedTesting_RunOptimizationTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	test := &PerformanceTest{
		ID:             "optimization_test",
		Name:           "Optimization Test",
		Type:           TestTypeOptimization,
		OptimizationID: "test_strategy",
		Config: &TestConfig{
			Duration:       1 * time.Second, // Short duration for testing
			Concurrency:    5,
			RequestRate:    10,
			RampUpTime:     100 * time.Millisecond,
			RampDownTime:   100 * time.Millisecond,
			TargetEndpoint: "/api/test",
			Timeout:        5 * time.Second,
			RetryCount:     3,
			ThinkTime:      50 * time.Millisecond,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    1 * time.Second,
			MaxErrorRate:       0.1,
			MinThroughput:      5.0,
			MaxCPUUsage:        90.0,
			MaxMemoryUsage:     95.0,
			MinCacheHitRate:    0.5,
			MaxDatabaseQueries: 100,
		},
	}

	testMetrics := &TestMetrics{}
	err := testing.runOptimizationTest(context.Background(), test, testMetrics)

	require.NoError(t, err)
	assert.Greater(t, testMetrics.TotalRequests, int64(0))
	assert.Greater(t, testMetrics.SuccessfulRequests, int64(0))
	assert.Greater(t, testMetrics.AverageResponseTime, time.Duration(0))
	assert.Greater(t, testMetrics.Throughput, 0.0)
}

func TestAutomatedTesting_EvaluateTestResult(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	// Test with passing metrics
	test := &PerformanceTest{
		Metrics: &TestMetrics{
			AverageResponseTime: 200 * time.Millisecond,
			ErrorRate:           0.02,
			Throughput:          100.0,
			CPUUsage:            70.0,
			MemoryUsage:         75.0,
			CacheHitRate:        0.85,
			DatabaseQueries:     500,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    500 * time.Millisecond,
			MaxErrorRate:       0.05,
			MinThroughput:      80.0,
			MaxCPUUsage:        80.0,
			MaxMemoryUsage:     85.0,
			MinCacheHitRate:    0.70,
			MaxDatabaseQueries: 1000,
		},
	}

	result := testing.evaluateTestResult(test)
	assert.Equal(t, TestResultPass, result)

	// Test with failing response time
	test.Metrics.AverageResponseTime = 600 * time.Millisecond
	result = testing.evaluateTestResult(test)
	assert.Equal(t, TestResultFail, result)

	// Test with failing error rate
	test.Metrics.AverageResponseTime = 200 * time.Millisecond
	test.Metrics.ErrorRate = 0.08
	result = testing.evaluateTestResult(test)
	assert.Equal(t, TestResultFail, result)

	// Test with failing throughput
	test.Metrics.ErrorRate = 0.02
	test.Metrics.Throughput = 50.0
	result = testing.evaluateTestResult(test)
	assert.Equal(t, TestResultFail, result)

	// Test with warning CPU usage
	test.Metrics.Throughput = 100.0
	test.Metrics.CPUUsage = 85.0
	result = testing.evaluateTestResult(test)
	assert.Equal(t, TestResultWarn, result)

	// Test with warning memory usage
	test.Metrics.CPUUsage = 70.0
	test.Metrics.MemoryUsage = 90.0
	result = testing.evaluateTestResult(test)
	assert.Equal(t, TestResultWarn, result)

	// Test with warning cache hit rate
	test.Metrics.MemoryUsage = 75.0
	test.Metrics.CacheHitRate = 0.60
	result = testing.evaluateTestResult(test)
	assert.Equal(t, TestResultWarn, result)

	// Test with warning database queries
	test.Metrics.CacheHitRate = 0.85
	test.Metrics.DatabaseQueries = 1200
	result = testing.evaluateTestResult(test)
	assert.Equal(t, TestResultWarn, result)
}

func TestAutomatedTesting_GenerateTestSummary(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	// Test with metrics
	test := &PerformanceTest{
		Name: "Test Performance Test",
		Metrics: &TestMetrics{
			TotalRequests:       1000,
			SuccessfulRequests:  950,
			FailedRequests:      50,
			AverageResponseTime: 250 * time.Millisecond,
			Throughput:          100.0,
		},
	}

	summary := testing.generateTestSummary(test)
	assert.Contains(t, summary, "Test Performance Test completed")
	assert.Contains(t, summary, "1000 requests")
	assert.Contains(t, summary, "100.00% success rate")
	assert.Contains(t, summary, "250.00ms avg response time")
	assert.Contains(t, summary, "100.00 req/s throughput")

	// Test with nil metrics
	test.Metrics = nil
	summary = testing.generateTestSummary(test)
	assert.Equal(t, "Test completed with no metrics collected", summary)
}

func TestAutomatedTesting_GenerateTestRecommendations(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	// Test with all metrics within thresholds
	test := &PerformanceTest{
		Metrics: &TestMetrics{
			AverageResponseTime: 200 * time.Millisecond,
			ErrorRate:           0.02,
			Throughput:          100.0,
			CPUUsage:            70.0,
			MemoryUsage:         75.0,
			CacheHitRate:        0.85,
			DatabaseQueries:     500,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    500 * time.Millisecond,
			MaxErrorRate:       0.05,
			MinThroughput:      80.0,
			MaxCPUUsage:        80.0,
			MaxMemoryUsage:     85.0,
			MinCacheHitRate:    0.70,
			MaxDatabaseQueries: 1000,
		},
	}

	recommendations := testing.generateTestRecommendations(test)
	assert.Len(t, recommendations, 1)
	assert.Contains(t, recommendations[0], "Performance is within acceptable thresholds")

	// Test with response time issues
	test.Metrics.AverageResponseTime = 600 * time.Millisecond
	recommendations = testing.generateTestRecommendations(test)
	assert.Contains(t, recommendations[0], "Consider implementing caching to reduce response times")
	assert.Contains(t, recommendations[1], "Review database query optimization")
	assert.Contains(t, recommendations[2], "Consider horizontal scaling for better performance")

	// Test with error rate issues
	test.Metrics.AverageResponseTime = 200 * time.Millisecond
	test.Metrics.ErrorRate = 0.08
	recommendations = testing.generateTestRecommendations(test)
	assert.Contains(t, recommendations[0], "Investigate error patterns and implement error handling")
	assert.Contains(t, recommendations[1], "Review system stability and resource allocation")
	assert.Contains(t, recommendations[2], "Consider implementing circuit breakers")

	// Test with throughput issues
	test.Metrics.ErrorRate = 0.02
	test.Metrics.Throughput = 50.0
	recommendations = testing.generateTestRecommendations(test)
	assert.Contains(t, recommendations[0], "Optimize request processing pipeline")
	assert.Contains(t, recommendations[1], "Consider implementing request queuing")
	assert.Contains(t, recommendations[2], "Review concurrency settings")
}

func TestAutomatedTesting_CreateBaselineTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	baselineTest := testing.createBaselineTest()

	assert.NotEmpty(t, baselineTest.ID)
	assert.Contains(t, baselineTest.ID, "baseline_")
	assert.Equal(t, "Baseline Performance Test", baselineTest.Name)
	assert.Equal(t, TestTypeBaseline, baselineTest.Type)
	assert.Equal(t, TestStatusPending, baselineTest.Status)
	assert.NotNil(t, baselineTest.Config)
	assert.NotNil(t, baselineTest.Thresholds)
	assert.Equal(t, 5*time.Minute, baselineTest.Config.Duration)
	assert.Equal(t, 10, baselineTest.Config.Concurrency)
	assert.Equal(t, 100, baselineTest.Config.RequestRate)
	assert.Equal(t, 500*time.Millisecond, baselineTest.Thresholds.MaxResponseTime)
	assert.Equal(t, 0.05, baselineTest.Thresholds.MaxErrorRate)
	assert.Equal(t, 80.0, baselineTest.Thresholds.MinThroughput)
}

func TestAutomatedTesting_CreateRegressionTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	baselineID := "baseline_123"
	regressionTest := testing.createRegressionTest(baselineID)

	assert.NotEmpty(t, regressionTest.ID)
	assert.Contains(t, regressionTest.ID, "regression_")
	assert.Equal(t, "Regression Performance Test", regressionTest.Name)
	assert.Equal(t, TestTypeRegression, regressionTest.Type)
	assert.Equal(t, TestStatusPending, regressionTest.Status)
	assert.Equal(t, baselineID, regressionTest.BaselineID)
	assert.NotNil(t, regressionTest.Config)
	assert.NotNil(t, regressionTest.Thresholds)
}

func TestAutomatedTesting_CreateOptimizationTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	strategyID := "strategy_123"
	optimizationTest := testing.createOptimizationTest(strategyID)

	assert.NotEmpty(t, optimizationTest.ID)
	assert.Contains(t, optimizationTest.ID, "optimization_")
	assert.Equal(t, "Optimization Performance Test", optimizationTest.Name)
	assert.Equal(t, TestTypeOptimization, optimizationTest.Type)
	assert.Equal(t, TestStatusPending, optimizationTest.Status)
	assert.Equal(t, strategyID, optimizationTest.OptimizationID)
	assert.NotNil(t, optimizationTest.Config)
	assert.NotNil(t, optimizationTest.Thresholds)
}

func TestAutomatedTesting_GetTests(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	// Add some tests
	test1 := &PerformanceTest{ID: "test1", Name: "Test 1"}
	test2 := &PerformanceTest{ID: "test2", Name: "Test 2"}
	testing.tests["test1"] = test1
	testing.tests["test2"] = test2

	tests := testing.GetTests()
	assert.Len(t, tests, 2)
}

func TestAutomatedTesting_GetTestByID(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	// Add a test
	test := &PerformanceTest{ID: "test1", Name: "Test 1"}
	testing.tests["test1"] = test

	// Get existing test
	retrievedTest, exists := testing.GetTestByID("test1")
	assert.True(t, exists)
	assert.Equal(t, test, retrievedTest)

	// Get non-existing test
	_, exists = testing.GetTestByID("non_existent")
	assert.False(t, exists)
}

func TestAutomatedTesting_GetTestsByType(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	// Add tests with different types
	test1 := &PerformanceTest{ID: "test1", Type: TestTypeLoad}
	test2 := &PerformanceTest{ID: "test2", Type: TestTypeStress}
	test3 := &PerformanceTest{ID: "test3", Type: TestTypeLoad}
	testing.tests["test1"] = test1
	testing.tests["test2"] = test2
	testing.tests["test3"] = test3

	loadTests := testing.GetTestsByType(TestTypeLoad)
	assert.Len(t, loadTests, 2)

	stressTests := testing.GetTestsByType(TestTypeStress)
	assert.Len(t, stressTests, 1)

	baselineTests := testing.GetTestsByType(TestTypeBaseline)
	assert.Len(t, baselineTests, 0)
}

func TestAutomatedTesting_GetTestsByStatus(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	// Add tests with different statuses
	test1 := &PerformanceTest{ID: "test1", Status: TestStatusPending}
	test2 := &PerformanceTest{ID: "test2", Status: TestStatusCompleted}
	test3 := &PerformanceTest{ID: "test3", Status: TestStatusPending}
	testing.tests["test1"] = test1
	testing.tests["test2"] = test2
	testing.tests["test3"] = test3

	pendingTests := testing.GetTestsByStatus(TestStatusPending)
	assert.Len(t, pendingTests, 2)

	completedTests := testing.GetTestsByStatus(TestStatusCompleted)
	assert.Len(t, completedTests, 1)

	failedTests := testing.GetTestsByStatus(TestStatusFailed)
	assert.Len(t, failedTests, 0)
}

func TestAutomatedTesting_CreateTestSuite(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	tests := []*PerformanceTest{
		{ID: "test1", Name: "Test 1"},
		{ID: "test2", Name: "Test 2"},
	}

	suite := testing.CreateTestSuite("Test Suite", "A test suite", tests)

	assert.NotEmpty(t, suite.ID)
	assert.Contains(t, suite.ID, "suite_")
	assert.Equal(t, "Test Suite", suite.Name)
	assert.Equal(t, "A test suite", suite.Description)
	assert.Len(t, suite.Tests, 2)
	assert.Equal(t, TestStatusPending, suite.Status)

	// Verify suite is stored
	suites := testing.GetTestSuites()
	assert.Len(t, suites, 1)
	assert.Equal(t, suite, suites[0])
}

func TestAutomatedTesting_RunTestSuite(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	// Create test suite with short-duration tests
	tests := []*PerformanceTest{
		{
			ID:   "test1",
			Name: "Test 1",
			Type: TestTypeLoad,
			Config: &TestConfig{
				Duration:       100 * time.Millisecond,
				Concurrency:    2,
				RequestRate:    5,
				RampUpTime:     10 * time.Millisecond,
				RampDownTime:   10 * time.Millisecond,
				TargetEndpoint: "/api/test",
				Timeout:        1 * time.Second,
				RetryCount:     1,
				ThinkTime:      10 * time.Millisecond,
			},
			Thresholds: &TestThresholds{
				MaxResponseTime:    1 * time.Second,
				MaxErrorRate:       0.1,
				MinThroughput:      1.0,
				MaxCPUUsage:        90.0,
				MaxMemoryUsage:     95.0,
				MinCacheHitRate:    0.5,
				MaxDatabaseQueries: 100,
			},
		},
		{
			ID:   "test2",
			Name: "Test 2",
			Type: TestTypeLoad,
			Config: &TestConfig{
				Duration:       100 * time.Millisecond,
				Concurrency:    2,
				RequestRate:    5,
				RampUpTime:     10 * time.Millisecond,
				RampDownTime:   10 * time.Millisecond,
				TargetEndpoint: "/api/test",
				Timeout:        1 * time.Second,
				RetryCount:     1,
				ThinkTime:      10 * time.Millisecond,
			},
			Thresholds: &TestThresholds{
				MaxResponseTime:    1 * time.Second,
				MaxErrorRate:       0.1,
				MinThroughput:      1.0,
				MaxCPUUsage:        90.0,
				MaxMemoryUsage:     95.0,
				MinCacheHitRate:    0.5,
				MaxDatabaseQueries: 100,
			},
		},
	}

	suite := testing.CreateTestSuite("Test Suite", "A test suite", tests)

	ctx := context.Background()
	err := testing.RunTestSuite(ctx, suite)

	require.NoError(t, err)
	assert.Equal(t, TestStatusCompleted, suite.Status)
	assert.Equal(t, TestResultPass, suite.Result)
	assert.Contains(t, suite.Summary, "All tests in suite completed successfully")
}

func TestAutomatedTesting_IntegrationTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)
	testing := NewAutomatedTesting(logger, metrics, detector, strategies, nil)

	ctx := context.Background()

	// Start the testing service
	err := testing.Start(ctx)
	require.NoError(t, err)

	// Create and run a baseline test
	baselineTest := testing.createBaselineTest()
	baselineTest.Config.Duration = 1 * time.Second // Longer duration for testing

	err = testing.RunTest(ctx, baselineTest)
	require.NoError(t, err)

	// Verify test was executed
	assert.Equal(t, TestStatusCompleted, baselineTest.Status)
	assert.NotNil(t, baselineTest.Metrics)
	assert.Greater(t, baselineTest.Metrics.TotalRequests, int64(0))
	assert.NotEmpty(t, baselineTest.Summary)
	assert.NotEmpty(t, baselineTest.Recommendations)

	// Create and run a regression test
	regressionTest := testing.createRegressionTest(baselineTest.ID)
	regressionTest.Config.Duration = 1 * time.Second // Longer duration for testing

	err = testing.RunTest(ctx, regressionTest)
	require.NoError(t, err)

	// Verify regression test was executed
	assert.Equal(t, TestStatusCompleted, regressionTest.Status)
	assert.NotNil(t, regressionTest.Metrics)
	assert.Greater(t, regressionTest.Metrics.TotalRequests, int64(0))

	// Verify tests are stored
	tests := testing.GetTests()
	assert.Len(t, tests, 2)

	// Stop the testing service
	testing.Stop()
}
