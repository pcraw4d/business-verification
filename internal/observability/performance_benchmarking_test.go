package observability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewPerformanceBenchmarkingSystem(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{
		BenchmarkInterval:       24 * time.Hour,
		BenchmarkTimeout:        30 * time.Minute,
		MaxConcurrentBenchmarks: 5,
		AutoGenerateReports:     true,
		ReportInterval:          7 * 24 * time.Hour,
		ReportRetention:         90 * 24 * time.Hour,
		MaxBenchmarkDuration:    1 * time.Hour,
		MinSampleSize:           100,
		MaxSampleSize:           10000,
	}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	assert.NotNil(t, pbs)
	assert.Equal(t, performanceMonitor, pbs.performanceMonitor)
	assert.Equal(t, regressionDetection, pbs.regressionDetection)
	assert.Equal(t, config, pbs.config)
	assert.NotNil(t, pbs.benchmarks)
	assert.NotNil(t, pbs.suites)
	assert.NotNil(t, pbs.comparisonEngine)
	assert.NotNil(t, pbs.benchmarkHistory)
}

func TestPerformanceBenchmarkingSystem_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{
		BenchmarkInterval: 100 * time.Millisecond, // Short interval for testing
		ReportInterval:    200 * time.Millisecond, // Short interval for testing
	}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the system
	err := pbs.Start(ctx)
	assert.NoError(t, err)

	// Wait a bit for goroutines to start
	time.Sleep(50 * time.Millisecond)

	// Stop the system
	err = pbs.Stop()
	assert.NoError(t, err)
}

func TestPerformanceBenchmarkingSystem_RunBenchmark(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	// Run a benchmark
	result, err := pbs.RunBenchmark("api_performance")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "api_performance", result.BenchmarkID)
	assert.NotZero(t, result.ExecutedAt)
	assert.NotZero(t, result.Duration)
	assert.NotNil(t, result.Performance)
	assert.NotNil(t, result.TestResults)
	assert.NotNil(t, result.ResourceUsage)
}

func TestPerformanceBenchmarkingSystem_RunSuite(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	// Run a suite
	results, err := pbs.RunSuite("comprehensive_performance")
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)
	assert.Equal(t, "comprehensive_performance", results[0].SuiteID)
}

func TestPerformanceBenchmarkingSystem_GetBenchmarks(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	benchmarks := pbs.GetBenchmarks()
	assert.NotEmpty(t, benchmarks)
	assert.Contains(t, benchmarks, "api_performance")

	apiBenchmark := benchmarks["api_performance"]
	assert.Equal(t, "API Performance Benchmark", apiBenchmark.Name)
	assert.Equal(t, "api", apiBenchmark.Category)
	assert.True(t, apiBenchmark.IsActive)
}

func TestPerformanceBenchmarkingSystem_GetBenchmark(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	// Get existing benchmark
	benchmark := pbs.GetBenchmark("api_performance")
	assert.NotNil(t, benchmark)
	assert.Equal(t, "API Performance Benchmark", benchmark.Name)

	// Get non-existent benchmark
	benchmark = pbs.GetBenchmark("non_existent")
	assert.Nil(t, benchmark)
}

func TestPerformanceBenchmarkingSystem_GetBenchmarkHistory(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	// Initially no history
	history := pbs.GetBenchmarkHistory()
	assert.Empty(t, history)

	// Run a benchmark to create history
	result, err := pbs.RunBenchmark("api_performance")
	assert.NoError(t, err)

	// Check that result was added to history
	history = pbs.GetBenchmarkHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, result.ID, history[0].ID)
}

func TestPerformanceBenchmarkingSystem_SimulateTestExecution(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	benchmark := pbs.GetBenchmark("api_performance")
	assert.NotNil(t, benchmark)

	scenario := BenchmarkScenario{
		Name:        "Normal Load",
		Description: "Normal operating conditions",
		Parameters:  map[string]string{"load": "normal"},
		Weight:      0.6,
	}

	testResult := TestResult{
		Name:       scenario.Name,
		Scenario:   scenario.Name,
		Status:     "running",
		Iterations: benchmark.Config.Iterations,
	}

	// Simulate test execution
	result := pbs.simulateTestExecution(benchmark, scenario, testResult)

	assert.NotZero(t, result.ResponseTime.Mean)
	assert.NotZero(t, result.Throughput)
	assert.Greater(t, result.SuccessRate, 0.0)
	assert.NotZero(t, result.CPUUsage)
	assert.NotZero(t, result.MemoryUsage)
	assert.Equal(t, "passed", result.Status)
}

func TestPerformanceBenchmarkingSystem_CalculateOverallPerformance(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	// Create test results
	testResults := []TestResult{
		{
			Name:     "Test 1",
			Scenario: "Normal",
			ResponseTime: struct {
				Min    time.Duration `json:"min"`
				Max    time.Duration `json:"max"`
				Mean   time.Duration `json:"mean"`
				P50    time.Duration `json:"p50"`
				P95    time.Duration `json:"p95"`
				P99    time.Duration `json:"p99"`
				StdDev time.Duration `json:"std_dev"`
			}{
				Mean: 250 * time.Millisecond,
			},
			Throughput:  1000.0,
			SuccessRate: 0.99,
		},
		{
			Name:     "Test 2",
			Scenario: "High Load",
			ResponseTime: struct {
				Min    time.Duration `json:"min"`
				Max    time.Duration `json:"max"`
				Mean   time.Duration `json:"mean"`
				P50    time.Duration `json:"p50"`
				P95    time.Duration `json:"p95"`
				P99    time.Duration `json:"p99"`
				StdDev time.Duration `json:"std_dev"`
			}{
				Mean: 300 * time.Millisecond,
			},
			Throughput:  800.0,
			SuccessRate: 0.98,
		},
	}

	// Calculate overall performance
	performance := pbs.calculateOverallPerformance(testResults)

	assert.NotNil(t, performance)
	assert.Equal(t, 275*time.Millisecond, performance.ResponseTime.P50) // Average of 250 and 300
	assert.Equal(t, 900.0, performance.Throughput.Target)               // Average of 1000 and 800
	assert.Equal(t, 0.985, performance.SuccessRate.Target)              // Average of 0.99 and 0.98
}

func TestPerformanceBenchmarkingSystem_DetermineBenchmarkStatus(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	benchmark := pbs.GetBenchmark("api_performance")
	assert.NotNil(t, benchmark)

	// Test with all passed tests
	result := &BenchmarkResult{
		TestResults: []TestResult{
			{Status: "passed"},
			{Status: "passed"},
			{Status: "passed"},
		},
	}

	status := pbs.determineBenchmarkStatus(benchmark, result)
	assert.Equal(t, "passed", status)

	// Test with some failed tests
	result = &BenchmarkResult{
		TestResults: []TestResult{
			{Status: "passed"},
			{Status: "failed"},
			{Status: "passed"},
		},
	}

	status = pbs.determineBenchmarkStatus(benchmark, result)
	assert.Equal(t, "partial", status)

	// Test with all failed tests
	result = &BenchmarkResult{
		TestResults: []TestResult{
			{Status: "failed"},
			{Status: "failed"},
			{Status: "failed"},
		},
	}

	status = pbs.determineBenchmarkStatus(benchmark, result)
	assert.Equal(t, "failed", status)
}

func TestBenchmarkComparisonEngine_Compare(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{
		ComparisonThresholds: struct {
			PerformanceGain   float64 `json:"performance_gain"`
			PerformanceLoss   float64 `json:"performance_loss"`
			SignificanceLevel float64 `json:"significance_level"`
		}{
			PerformanceGain: 5.0,
		},
	}

	engine := NewBenchmarkComparisonEngine(config, logger)

	// Create baseline result
	baseline := &BenchmarkResult{
		ID:         "baseline_1",
		ExecutedAt: time.Now().UTC().Add(-24 * time.Hour),
		Performance: &BenchmarkPerformance{
			ResponseTime: struct {
				P50 time.Duration `json:"p50"`
				P95 time.Duration `json:"p95"`
				P99 time.Duration `json:"p99"`
				Max time.Duration `json:"max"`
			}{
				P50: 250 * time.Millisecond,
			},
			Throughput: struct {
				Min    float64 `json:"min"`
				Target float64 `json:"target"`
				Max    float64 `json:"max"`
			}{
				Target: 1000.0,
			},
			SuccessRate: struct {
				Min     float64 `json:"min"`
				Target  float64 `json:"target"`
				Optimal float64 `json:"optimal"`
			}{
				Target: 0.99,
			},
		},
	}

	// Create current result with improvement
	current := &BenchmarkResult{
		ID:         "current_1",
		ExecutedAt: time.Now().UTC(),
		Performance: &BenchmarkPerformance{
			ResponseTime: struct {
				P50 time.Duration `json:"p50"`
				P95 time.Duration `json:"p95"`
				P99 time.Duration `json:"p99"`
				Max time.Duration `json:"max"`
			}{
				P50: 200 * time.Millisecond, // 20% improvement
			},
			Throughput: struct {
				Min    float64 `json:"min"`
				Target float64 `json:"target"`
				Max    float64 `json:"max"`
			}{
				Target: 1100.0, // 10% improvement
			},
			SuccessRate: struct {
				Min     float64 `json:"min"`
				Target  float64 `json:"target"`
				Optimal float64 `json:"optimal"`
			}{
				Target: 0.995, // 0.5% improvement
			},
		},
	}

	// Perform comparison
	comparison := engine.Compare(baseline, current)

	assert.NotNil(t, comparison)
	assert.Equal(t, baseline.ID, comparison.BaselineID)
	assert.Equal(t, current.ID, comparison.CurrentID)
	assert.True(t, comparison.ResponseTime.Improvement)
	assert.True(t, comparison.ResponseTime.Significant)
	assert.True(t, comparison.Throughput.Improvement)
	assert.True(t, comparison.Throughput.Significant)
	assert.True(t, comparison.SuccessRate.Improvement)
	assert.False(t, comparison.SuccessRate.Significant) // Below threshold
	assert.True(t, comparison.PerformanceGain)
	assert.Greater(t, comparison.OverallScore, 100.0) // Should be above 100 due to improvements
}

func TestBenchmarkComparisonEngine_CalculateOverallScore(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{}

	engine := NewBenchmarkComparisonEngine(config, logger)

	// Test with improvements
	comparison := &BenchmarkComparison{
		Improvements: []string{"response_time", "throughput"},
		Regressions:  []string{},
	}

	score := engine.calculateOverallScore(comparison)
	assert.Equal(t, 118.0, score) // 100 + 10 + 8

	// Test with regressions
	comparison = &BenchmarkComparison{
		Improvements: []string{},
		Regressions:  []string{"response_time", "success_rate"},
	}

	score = engine.calculateOverallScore(comparison)
	assert.Equal(t, 55.0, score) // 100 - 20 - 25

	// Test with mixed results
	comparison = &BenchmarkComparison{
		Improvements: []string{"response_time"},
		Regressions:  []string{"throughput"},
	}

	score = engine.calculateOverallScore(comparison)
	assert.Equal(t, 90.0, score) // 100 + 10 - 20

	// Test with no changes
	comparison = &BenchmarkComparison{
		Improvements: []string{},
		Regressions:  []string{},
	}

	score = engine.calculateOverallScore(comparison)
	assert.Equal(t, 100.0, score)
}

func TestPerformanceBenchmarkingSystem_StoreBenchmarkResult(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{
		ReportRetention: 1 * time.Hour, // Short retention for testing
	}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	// Create a benchmark result
	result := &BenchmarkResult{
		ID:          "test_result_1",
		BenchmarkID: "api_performance",
		ExecutedAt:  time.Now().UTC(),
		Status:      "passed",
	}

	// Store the result
	pbs.storeBenchmarkResult(result)

	// Check that result was stored
	history := pbs.GetBenchmarkHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, result.ID, history[0].ID)
}

func TestPerformanceBenchmarkingSystem_FindBaseline(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	// Initially no baseline
	baseline := pbs.findBaseline("api_performance")
	assert.Nil(t, baseline)

	// Add a failed result
	failedResult := &BenchmarkResult{
		ID:          "failed_result",
		BenchmarkID: "api_performance",
		ExecutedAt:  time.Now().UTC(),
		Status:      "failed",
	}
	pbs.storeBenchmarkResult(failedResult)

	// Should still not find baseline (failed result)
	baseline = pbs.findBaseline("api_performance")
	assert.Nil(t, baseline)

	// Add a passed result
	passedResult := &BenchmarkResult{
		ID:          "passed_result",
		BenchmarkID: "api_performance",
		ExecutedAt:  time.Now().UTC(),
		Status:      "passed",
	}
	pbs.storeBenchmarkResult(passedResult)

	// Should find the passed result as baseline
	baseline = pbs.findBaseline("api_performance")
	assert.NotNil(t, baseline)
	assert.Equal(t, "passed_result", baseline.ID)
}

func TestPerformanceBenchmarkingSystem_PerformComparison(t *testing.T) {
	logger := zap.NewNop()
	config := BenchmarkingConfig{}

	performanceMonitor := &PerformanceMonitor{}
	regressionDetection := &RegressionDetectionSystem{}

	pbs := NewPerformanceBenchmarkingSystem(performanceMonitor, regressionDetection, config, logger)

	// Create a baseline result
	baseline := &BenchmarkResult{
		ID:          "baseline_1",
		BenchmarkID: "api_performance",
		ExecutedAt:  time.Now().UTC().Add(-24 * time.Hour),
		Status:      "passed",
		Performance: &BenchmarkPerformance{
			ResponseTime: struct {
				P50 time.Duration `json:"p50"`
				P95 time.Duration `json:"p95"`
				P99 time.Duration `json:"p99"`
				Max time.Duration `json:"max"`
			}{
				P50: 250 * time.Millisecond,
			},
			Throughput: struct {
				Min    float64 `json:"min"`
				Target float64 `json:"target"`
				Max    float64 `json:"max"`
			}{
				Target: 1000.0,
			},
			SuccessRate: struct {
				Min     float64 `json:"min"`
				Target  float64 `json:"target"`
				Optimal float64 `json:"optimal"`
			}{
				Target: 0.99,
			},
		},
	}

	// Store baseline
	pbs.storeBenchmarkResult(baseline)

	// Create current result
	current := &BenchmarkResult{
		ID:          "current_1",
		BenchmarkID: "api_performance",
		ExecutedAt:  time.Now().UTC(),
		Status:      "passed",
		Performance: &BenchmarkPerformance{
			ResponseTime: struct {
				P50 time.Duration `json:"p50"`
				P95 time.Duration `json:"p95"`
				P99 time.Duration `json:"p99"`
				Max time.Duration `json:"max"`
			}{
				P50: 300 * time.Millisecond, // 20% degradation
			},
			Throughput: struct {
				Min    float64 `json:"min"`
				Target float64 `json:"target"`
				Max    float64 `json:"max"`
			}{
				Target: 800.0, // 20% degradation
			},
			SuccessRate: struct {
				Min     float64 `json:"min"`
				Target  float64 `json:"target"`
				Optimal float64 `json:"optimal"`
			}{
				Target: 0.98, // 1% degradation
			},
		},
	}

	// Perform comparison
	pbs.performComparison(current)

	// Check that comparison was added
	assert.NotNil(t, current.Comparison)
	assert.Equal(t, baseline.ID, current.Comparison.BaselineID)
	assert.Equal(t, current.ID, current.Comparison.CurrentID)
	assert.False(t, current.Comparison.ResponseTime.Improvement)
	assert.False(t, current.Comparison.Throughput.Improvement)
	assert.False(t, current.Comparison.SuccessRate.Improvement)
}

func TestBenchmarkConfig_Validation(t *testing.T) {
	config := BenchmarkConfig{
		TestDuration:       5 * time.Minute,
		Concurrency:        100,
		RequestRate:        1000.0,
		DataSize:           1024,
		Iterations:         1000,
		TargetResponseTime: 500 * time.Millisecond,
		TargetThroughput:   1000.0,
		TargetSuccessRate:  0.99,
		TargetErrorRate:    0.01,
		MaxCPUUsage:        80.0,
		MaxMemoryUsage:     85.0,
		MaxDiskUsage:       90.0,
		Scenarios: []BenchmarkScenario{
			{
				Name:        "Normal Load",
				Description: "Normal operating conditions",
				Parameters:  map[string]string{"load": "normal"},
				Weight:      0.6,
			},
		},
	}

	// Validate configuration
	assert.Greater(t, config.TestDuration, 0)
	assert.Greater(t, config.Concurrency, 0)
	assert.Greater(t, config.RequestRate, 0)
	assert.Greater(t, config.DataSize, 0)
	assert.Greater(t, config.Iterations, 0)
	assert.Greater(t, config.TargetResponseTime, 0)
	assert.Greater(t, config.TargetThroughput, 0)
	assert.Greater(t, config.TargetSuccessRate, 0)
	assert.Less(t, config.TargetSuccessRate, 1.0)
	assert.Greater(t, config.TargetErrorRate, 0)
	assert.Less(t, config.TargetErrorRate, 1.0)
	assert.Greater(t, config.MaxCPUUsage, 0)
	assert.Less(t, config.MaxCPUUsage, 100.0)
	assert.Greater(t, config.MaxMemoryUsage, 0)
	assert.Less(t, config.MaxMemoryUsage, 100.0)
	assert.Greater(t, config.MaxDiskUsage, 0)
	assert.Less(t, config.MaxDiskUsage, 100.0)
	assert.NotEmpty(t, config.Scenarios)
}

func TestBenchmarkSuite_Validation(t *testing.T) {
	suite := &BenchmarkSuite{
		ID:                "test_suite",
		Name:              "Test Suite",
		Description:       "A test suite for validation",
		Version:           "1.0",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
		Benchmarks:        []string{"benchmark1", "benchmark2"},
		Order:             []string{"benchmark1", "benchmark2"},
		ParallelExecution: false,
		SuiteTimeout:      2 * time.Hour,
		RetryFailed:       true,
		MaxRetries:        3,
		Tags:              make(map[string]string),
		Environment:       "production",
		Category:          "comprehensive",
	}

	// Validate suite configuration
	assert.NotEmpty(t, suite.ID)
	assert.NotEmpty(t, suite.Name)
	assert.NotEmpty(t, suite.Description)
	assert.NotEmpty(t, suite.Version)
	assert.NotZero(t, suite.CreatedAt)
	assert.NotZero(t, suite.UpdatedAt)
	assert.NotEmpty(t, suite.Benchmarks)
	assert.NotEmpty(t, suite.Order)
	assert.Greater(t, suite.SuiteTimeout, 0)
	assert.Greater(t, suite.MaxRetries, 0)
	assert.NotEmpty(t, suite.Environment)
	assert.NotEmpty(t, suite.Category)
}
