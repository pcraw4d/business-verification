package risk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPerformanceBenchmarkingCreation tests the creation of the performance benchmarking framework
func TestPerformanceBenchmarkingCreation(t *testing.T) {
	config := &BenchmarkConfig{
		TestEnvironment:      "test",
		BenchmarkTimeout:     5 * time.Minute,
		ReportOutputPath:     "./test-reports",
		LogLevel:             "info",
		EnableProfiling:      false,
		ProfilingOutputPath:  "./profiling",
		ConcurrencyLevels:    []int{1, 5, 10},
		TestDataSizes:        []int{10, 100, 1000},
		IterationCounts:      []int{5, 10, 20},
		WarmupIterations:     2,
		EnvironmentVariables: make(map[string]string),
		DatabaseConfig: &DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "kyb_test",
			Username: "test_user",
			Password: "test_password",
			SSLMode:  "disable",
		},
		APIConfig: &APIConfig{
			BaseURL:   "http://localhost:8080",
			Timeout:   30 * time.Second,
			Headers:   make(map[string]string),
			RateLimit: 100,
		},
		ResourceLimits: &ResourceLimits{
			MaxMemoryMB:     512,
			MaxCPUPercents:  70,
			MaxGoroutines:   500,
			MaxConnections:  50,
			MaxFileHandles:  500,
			TimeoutDuration: 10 * time.Minute,
		},
	}

	benchmarking := NewPerformanceBenchmarking(config)
	require.NotNil(t, benchmarking)
	assert.Equal(t, config, benchmarking.config)
	assert.NotNil(t, benchmarking.benchmarks)
	assert.NotNil(t, benchmarking.results)
	assert.NotNil(t, benchmarking.metricsCollector)
	assert.NotNil(t, benchmarking.reportGenerator)
}

// TestBenchmarkRunnerCreation tests the creation of the benchmark runner
func TestBenchmarkRunnerCreation(t *testing.T) {
	config := &BenchmarkConfig{
		TestEnvironment:      "test",
		BenchmarkTimeout:     5 * time.Minute,
		ReportOutputPath:     "./test-reports",
		LogLevel:             "info",
		EnableProfiling:      false,
		ProfilingOutputPath:  "./profiling",
		ConcurrencyLevels:    []int{1, 5, 10},
		TestDataSizes:        []int{10, 100, 1000},
		IterationCounts:      []int{5, 10, 20},
		WarmupIterations:     2,
		EnvironmentVariables: make(map[string]string),
		DatabaseConfig: &DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "kyb_test",
			Username: "test_user",
			Password: "test_password",
			SSLMode:  "disable",
		},
		APIConfig: &APIConfig{
			BaseURL:   "http://localhost:8080",
			Timeout:   30 * time.Second,
			Headers:   make(map[string]string),
			RateLimit: 100,
		},
		ResourceLimits: &ResourceLimits{
			MaxMemoryMB:     512,
			MaxCPUPercents:  70,
			MaxGoroutines:   500,
			MaxConnections:  50,
			MaxFileHandles:  500,
			TimeoutDuration: 10 * time.Minute,
		},
	}

	runner := NewBenchmarkRunner(config)
	require.NotNil(t, runner)
	assert.Equal(t, config, runner.config)
	assert.NotNil(t, runner.benchmarking)
}

// TestKYBBenchmarksCreation tests the creation of KYB benchmarks
func TestKYBBenchmarksCreation(t *testing.T) {
	benchmarks := CreateKYBBenchmarks()
	require.NotEmpty(t, benchmarks)

	// Verify we have the expected benchmarks
	benchmarkIDs := make(map[string]bool)
	for _, benchmark := range benchmarks {
		benchmarkIDs[benchmark.ID] = true
	}

	assert.True(t, benchmarkIDs["BV_BENCH_001"], "Business verification throughput benchmark should exist")
	assert.True(t, benchmarkIDs["BV_BENCH_002"], "Business verification latency benchmark should exist")
	assert.True(t, benchmarkIDs["RA_BENCH_001"], "Risk assessment throughput benchmark should exist")
	assert.True(t, benchmarkIDs["RA_BENCH_002"], "Risk assessment latency benchmark should exist")
	assert.True(t, benchmarkIDs["DE_BENCH_001"], "Data export throughput benchmark should exist")
	assert.True(t, benchmarkIDs["DE_BENCH_002"], "Data export memory benchmark should exist")
	assert.True(t, benchmarkIDs["DB_BENCH_001"], "Database query performance benchmark should exist")
	assert.True(t, benchmarkIDs["DB_BENCH_002"], "Database write performance benchmark should exist")
	assert.True(t, benchmarkIDs["API_BENCH_001"], "API response time benchmark should exist")
	assert.True(t, benchmarkIDs["API_BENCH_002"], "API concurrent users benchmark should exist")
	assert.True(t, benchmarkIDs["MEM_BENCH_001"], "Memory usage under load benchmark should exist")
	assert.True(t, benchmarkIDs["MEM_BENCH_002"], "Memory leak detection benchmark should exist")
	assert.True(t, benchmarkIDs["CPU_BENCH_001"], "CPU usage under load benchmark should exist")
	assert.True(t, benchmarkIDs["CPU_BENCH_002"], "CPU efficiency benchmark should exist")
}

// TestBenchmarkStructure tests the structure of benchmarks
func TestBenchmarkStructure(t *testing.T) {
	benchmarks := CreateKYBBenchmarks()
	require.NotEmpty(t, benchmarks)

	for _, benchmark := range benchmarks {
		// Verify required fields
		assert.NotEmpty(t, benchmark.ID, "Benchmark ID should not be empty")
		assert.NotEmpty(t, benchmark.Name, "Benchmark name should not be empty")
		assert.NotEmpty(t, benchmark.Description, "Benchmark description should not be empty")
		assert.NotEmpty(t, benchmark.Category, "Benchmark category should not be empty")
		assert.NotEmpty(t, benchmark.Priority, "Benchmark priority should not be empty")
		assert.NotNil(t, benchmark.Function, "Benchmark function should not be nil")
		assert.NotNil(t, benchmark.Parameters, "Benchmark parameters should not be nil")
		assert.NotNil(t, benchmark.ExpectedMetrics, "Benchmark expected metrics should not be nil")
		assert.NotEmpty(t, benchmark.Tags, "Benchmark should have tags")
	}
}

// TestBenchmarkExecution tests the execution of a specific benchmark
func TestBenchmarkExecution(t *testing.T) {
	config := &BenchmarkConfig{
		TestEnvironment:      "test",
		BenchmarkTimeout:     2 * time.Minute,
		ReportOutputPath:     "./test-reports",
		LogLevel:             "info",
		EnableProfiling:      false,
		ProfilingOutputPath:  "./profiling",
		ConcurrencyLevels:    []int{1, 5},
		TestDataSizes:        []int{10, 100},
		IterationCounts:      []int{2, 5},
		WarmupIterations:     1,
		EnvironmentVariables: make(map[string]string),
		DatabaseConfig: &DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "kyb_test",
			Username: "test_user",
			Password: "test_password",
			SSLMode:  "disable",
		},
		APIConfig: &APIConfig{
			BaseURL:   "http://localhost:8080",
			Timeout:   30 * time.Second,
			Headers:   make(map[string]string),
			RateLimit: 100,
		},
		ResourceLimits: &ResourceLimits{
			MaxMemoryMB:     512,
			MaxCPUPercents:  70,
			MaxGoroutines:   500,
			MaxConnections:  50,
			MaxFileHandles:  500,
			TimeoutDuration: 2 * time.Minute,
		},
	}

	runner := NewBenchmarkRunner(config)
	require.NotNil(t, runner)

	// Test executing a specific benchmark
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	result, err := runner.RunSpecificBenchmark(ctx, "BV_BENCH_001")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "BV_BENCH_001", result.BenchmarkID)
	assert.True(t, result.Success)
	assert.True(t, result.Duration > 0)
	assert.NotNil(t, result.Metrics)
	assert.NotNil(t, result.ResourceUsage)
}

// TestBenchmarkSuiteExecution tests the execution of the complete benchmark suite
func TestBenchmarkSuiteExecution(t *testing.T) {
	config := &BenchmarkConfig{
		TestEnvironment:      "test",
		BenchmarkTimeout:     2 * time.Minute,
		ReportOutputPath:     "./test-reports",
		LogLevel:             "info",
		EnableProfiling:      false,
		ProfilingOutputPath:  "./profiling",
		ConcurrencyLevels:    []int{1, 2},
		TestDataSizes:        []int{10, 50},
		IterationCounts:      []int{1, 2},
		WarmupIterations:     1,
		EnvironmentVariables: make(map[string]string),
		DatabaseConfig: &DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "kyb_test",
			Username: "test_user",
			Password: "test_password",
			SSLMode:  "disable",
		},
		APIConfig: &APIConfig{
			BaseURL:   "http://localhost:8080",
			Timeout:   30 * time.Second,
			Headers:   make(map[string]string),
			RateLimit: 100,
		},
		ResourceLimits: &ResourceLimits{
			MaxMemoryMB:     512,
			MaxCPUPercents:  70,
			MaxGoroutines:   500,
			MaxConnections:  50,
			MaxFileHandles:  500,
			TimeoutDuration: 2 * time.Minute,
		},
	}

	runner := NewBenchmarkRunner(config)
	require.NotNil(t, runner)

	// Test running the complete benchmark suite
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	results, err := runner.RunBenchmarkSuite(ctx)
	require.NoError(t, err)
	require.NotNil(t, results)

	assert.Equal(t, "test", results.Environment)
	assert.True(t, results.TotalBenchmarks > 0)
	assert.True(t, results.TotalDuration > 0)
	assert.True(t, results.PassRate >= 0)
	assert.NotNil(t, results.Summary)
	assert.NotNil(t, results.BenchmarkResults)
}

// TestBenchmarkConfigStructure tests the structure of benchmark configuration
func TestBenchmarkConfigStructure(t *testing.T) {
	config := &BenchmarkConfig{
		TestEnvironment:      "test",
		BenchmarkTimeout:     5 * time.Minute,
		ReportOutputPath:     "./test-reports",
		LogLevel:             "info",
		EnableProfiling:      false,
		ProfilingOutputPath:  "./profiling",
		ConcurrencyLevels:    []int{1, 5, 10},
		TestDataSizes:        []int{10, 100, 1000},
		IterationCounts:      []int{5, 10, 20},
		WarmupIterations:     2,
		EnvironmentVariables: make(map[string]string),
		DatabaseConfig: &DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "kyb_test",
			Username: "test_user",
			Password: "test_password",
			SSLMode:  "disable",
		},
		APIConfig: &APIConfig{
			BaseURL:   "http://localhost:8080",
			Timeout:   30 * time.Second,
			Headers:   make(map[string]string),
			RateLimit: 100,
		},
		ResourceLimits: &ResourceLimits{
			MaxMemoryMB:     512,
			MaxCPUPercents:  70,
			MaxGoroutines:   500,
			MaxConnections:  50,
			MaxFileHandles:  500,
			TimeoutDuration: 10 * time.Minute,
		},
	}

	// Verify configuration structure
	assert.Equal(t, "test", config.TestEnvironment)
	assert.Equal(t, 5*time.Minute, config.BenchmarkTimeout)
	assert.Equal(t, "./test-reports", config.ReportOutputPath)
	assert.Equal(t, "info", config.LogLevel)
	assert.False(t, config.EnableProfiling)
	assert.Equal(t, "./profiling", config.ProfilingOutputPath)
	assert.Len(t, config.ConcurrencyLevels, 3)
	assert.Len(t, config.TestDataSizes, 3)
	assert.Len(t, config.IterationCounts, 3)
	assert.Equal(t, 2, config.WarmupIterations)
	assert.NotNil(t, config.DatabaseConfig)
	assert.NotNil(t, config.APIConfig)
	assert.NotNil(t, config.ResourceLimits)
}

// TestResourceLimitsStructure tests the structure of resource limits
func TestResourceLimitsStructure(t *testing.T) {
	limits := &ResourceLimits{
		MaxMemoryMB:     1024,
		MaxCPUPercents:  80,
		MaxGoroutines:   1000,
		MaxConnections:  100,
		MaxFileHandles:  1000,
		TimeoutDuration: 30 * time.Minute,
	}

	// Verify resource limits structure
	assert.Equal(t, 1024, limits.MaxMemoryMB)
	assert.Equal(t, 80, limits.MaxCPUPercents)
	assert.Equal(t, 1000, limits.MaxGoroutines)
	assert.Equal(t, 100, limits.MaxConnections)
	assert.Equal(t, 1000, limits.MaxFileHandles)
	assert.Equal(t, 30*time.Minute, limits.TimeoutDuration)
}

// TestExpectedMetricsStructure tests the structure of expected metrics
func TestExpectedMetricsStructure(t *testing.T) {
	expected := &ExpectedMetrics{
		MinThroughput:  100.0,
		MaxLatency:     5000.0,
		MaxP95Latency:  3000.0,
		MaxP99Latency:  4000.0,
		MaxErrorRate:   1.0,
		MaxMemoryUsage: 100 * 1024 * 1024,
		MaxCPUUsage:    80.0,
		MaxGoroutines:  1000,
		MinSuccessRate: 99.0,
	}

	// Verify expected metrics structure
	assert.Equal(t, 100.0, expected.MinThroughput)
	assert.Equal(t, 5000.0, expected.MaxLatency)
	assert.Equal(t, 3000.0, expected.MaxP95Latency)
	assert.Equal(t, 4000.0, expected.MaxP99Latency)
	assert.Equal(t, 1.0, expected.MaxErrorRate)
	assert.Equal(t, uint64(100*1024*1024), expected.MaxMemoryUsage)
	assert.Equal(t, 80.0, expected.MaxCPUUsage)
	assert.Equal(t, 1000, expected.MaxGoroutines)
	assert.Equal(t, 99.0, expected.MinSuccessRate)
}

// TestPerformanceMetricsStructure tests the structure of performance metrics
func TestPerformanceMetricsStructure(t *testing.T) {
	metrics := &PerformanceMetrics{
		Throughput:      150.0,
		Latency:         2000.0,
		P95Latency:      3000.0,
		P99Latency:      4000.0,
		MaxLatency:      5000.0,
		MinLatency:      1000.0,
		ErrorRate:       0.5,
		SuccessRate:     99.5,
		OperationsCount: 1000,
		SuccessfulOps:   995,
		FailedOps:       5,
		DataProcessed:   1024 * 1024,
		DataThroughput:  10.0,
	}

	// Verify performance metrics structure
	assert.Equal(t, 150.0, metrics.Throughput)
	assert.Equal(t, 2000.0, metrics.Latency)
	assert.Equal(t, 3000.0, metrics.P95Latency)
	assert.Equal(t, 4000.0, metrics.P99Latency)
	assert.Equal(t, 5000.0, metrics.MaxLatency)
	assert.Equal(t, 1000.0, metrics.MinLatency)
	assert.Equal(t, 0.5, metrics.ErrorRate)
	assert.Equal(t, 99.5, metrics.SuccessRate)
	assert.Equal(t, 1000, metrics.OperationsCount)
	assert.Equal(t, 995, metrics.SuccessfulOps)
	assert.Equal(t, 5, metrics.FailedOps)
	assert.Equal(t, int64(1024*1024), metrics.DataProcessed)
	assert.Equal(t, 10.0, metrics.DataThroughput)
}

// TestResourceUsageStructure tests the structure of resource usage
func TestResourceUsageStructure(t *testing.T) {
	usage := &ResourceUsage{
		MemoryUsage:        200 * 1024 * 1024,
		MemoryPeak:         250 * 1024 * 1024,
		MemoryDelta:        50 * 1024 * 1024,
		CPUUsage:           65.0,
		CPUTime:            300.0,
		GoroutinesCount:    100,
		GoroutinesPeak:     150,
		GoroutinesDelta:    50,
		GCCollections:      10,
		GCPauseTime:        5.0,
		FileHandles:        50,
		NetworkConnections: 25,
	}

	// Verify resource usage structure
	assert.Equal(t, uint64(200*1024*1024), usage.MemoryUsage)
	assert.Equal(t, uint64(250*1024*1024), usage.MemoryPeak)
	assert.Equal(t, int64(50*1024*1024), usage.MemoryDelta)
	assert.Equal(t, 65.0, usage.CPUUsage)
	assert.Equal(t, 300.0, usage.CPUTime)
	assert.Equal(t, 100, usage.GoroutinesCount)
	assert.Equal(t, 150, usage.GoroutinesPeak)
	assert.Equal(t, 50, usage.GoroutinesDelta)
	assert.Equal(t, 10, usage.GCCollections)
	assert.Equal(t, 5.0, usage.GCPauseTime)
	assert.Equal(t, 50, usage.FileHandles)
	assert.Equal(t, 25, usage.NetworkConnections)
}

// TestBenchmarkResultStructure tests the structure of benchmark result
func TestBenchmarkResultStructure(t *testing.T) {
	result := &BenchmarkResult{
		BenchmarkID:      "TEST_BENCH_001",
		Iteration:        1,
		ConcurrencyLevel: 10,
		TestDataSize:     100,
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(5 * time.Second),
		Duration:         5 * time.Second,
		Success:          true,
		ErrorMessage:     "",
		Metrics: &PerformanceMetrics{
			Throughput:  100.0,
			Latency:     2000.0,
			ErrorRate:   0.5,
			SuccessRate: 99.5,
		},
		ResourceUsage: &ResourceUsage{
			MemoryUsage:     100 * 1024 * 1024,
			CPUUsage:        50.0,
			GoroutinesCount: 50,
		},
		CustomMetrics: map[string]interface{}{
			"custom_metric": "value",
		},
	}

	// Verify benchmark result structure
	assert.Equal(t, "TEST_BENCH_001", result.BenchmarkID)
	assert.Equal(t, 1, result.Iteration)
	assert.Equal(t, 10, result.ConcurrencyLevel)
	assert.Equal(t, 100, result.TestDataSize)
	assert.True(t, result.Success)
	assert.Empty(t, result.ErrorMessage)
	assert.NotNil(t, result.Metrics)
	assert.NotNil(t, result.ResourceUsage)
	assert.NotNil(t, result.CustomMetrics)
	assert.Equal(t, "value", result.CustomMetrics["custom_metric"])
}
