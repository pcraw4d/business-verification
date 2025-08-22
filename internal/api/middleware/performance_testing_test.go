package middleware

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestPerformanceTestManager(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultPerformanceTestConfig()

	// Use temporary directory for testing
	tempDir, err := os.MkdirTemp("", "performance_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.TestDirectory = tempDir
	config.ReportDirectory = filepath.Join(tempDir, "reports")
	config.MonitorEnabled = false // Disable monitoring for tests

	manager := NewPerformanceTestManager(config, logger)
	defer manager.Shutdown()

	t.Run("default configuration", func(t *testing.T) {
		if manager.config == nil {
			t.Error("expected config to be set")
		}
		if manager.loadTester == nil {
			t.Error("expected load tester to be initialized")
		}
		if manager.stressTester == nil {
			t.Error("expected stress tester to be initialized")
		}
		if manager.benchmarker == nil {
			t.Error("expected benchmarker to be initialized")
		}
		if manager.reporter == nil {
			t.Error("expected reporter to be initialized")
		}
		if manager.monitor == nil {
			t.Error("expected monitor to be initialized")
		}
	})

	t.Run("get results", func(t *testing.T) {
		results := manager.GetResults()
		if results == nil {
			t.Error("expected results to be returned")
		}
		if results.GeneratedAt.IsZero() {
			t.Error("expected generated at time to be set")
		}
	})

	t.Run("shutdown", func(t *testing.T) {
		err := manager.Shutdown()
		if err != nil {
			t.Errorf("expected no error during shutdown, got %v", err)
		}
	})
}

func TestLoadTester(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultPerformanceTestConfig()

	// Use temporary directory for testing
	tempDir, err := os.MkdirTemp("", "load_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.TestDirectory = tempDir
	config.LoadTestDuration = 1 * time.Second // Short duration for testing
	config.LoadTestRPS = 10                   // Low RPS for testing
	config.LoadTestUsers = 2                  // Few users for testing

	loadTester := NewPerformanceLoadTester(config, logger)

	t.Run("create test scenarios", func(t *testing.T) {
		scenarios := loadTester.createTestScenarios()
		if len(scenarios) == 0 {
			t.Error("expected test scenarios to be created")
		}

		// Check for expected scenarios
		scenarioTypes := make(map[string]bool)
		for _, scenario := range scenarios {
			scenarioTypes[scenario.ID] = true
		}

		expectedScenarios := []string{"classification_request", "verification_request", "health_check"}
		for _, expected := range expectedScenarios {
			if !scenarioTypes[expected] {
				t.Errorf("expected scenario %s not found", expected)
			}
		}
	})

	t.Run("execute request", func(t *testing.T) {
		request := &TestRequest{
			ID:        "test_request",
			Method:    "GET",
			URL:       "http://localhost:8080/health",
			Headers:   map[string]string{},
			Body:      []byte{},
			StartTime: time.Now(),
		}

		response := loadTester.executeRequest(request)
		if response == nil {
			t.Error("expected response to be returned")
		}
		if response.RequestID != request.ID {
			t.Errorf("expected request ID %s, got %s", request.ID, response.RequestID)
		}
	})

	t.Run("calculate percentiles", func(t *testing.T) {
		responseTimes := []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			300 * time.Millisecond,
			400 * time.Millisecond,
			500 * time.Millisecond,
		}

		loadTester.calculatePercentiles(responseTimes)

		if loadTester.stats.P50ResponseTime != 300*time.Millisecond {
			t.Errorf("expected P50 %v, got %v", 300*time.Millisecond, loadTester.stats.P50ResponseTime)
		}
		if loadTester.stats.P95ResponseTime != 400*time.Millisecond {
			t.Errorf("expected P95 %v, got %v", 400*time.Millisecond, loadTester.stats.P95ResponseTime)
		}
		if loadTester.stats.P99ResponseTime != 400*time.Millisecond {
			t.Errorf("expected P99 %v, got %v", 400*time.Millisecond, loadTester.stats.P99ResponseTime)
		}
		if loadTester.stats.AverageResponseTime != 300*time.Millisecond {
			t.Errorf("expected average %v, got %v", 300*time.Millisecond, loadTester.stats.AverageResponseTime)
		}
	})

	t.Run("calculate final stats", func(t *testing.T) {
		loadTester.stats.TotalRequests = 100
		loadTester.stats.FailedRequests = 5
		loadTester.stats.Duration = 10 * time.Second

		loadTester.calculateFinalStats()

		expectedErrorRate := 0.05
		if loadTester.stats.ErrorRate != expectedErrorRate {
			t.Errorf("expected error rate %f, got %f", expectedErrorRate, loadTester.stats.ErrorRate)
		}

		expectedRPS := 10.0
		if loadTester.stats.RequestsPerSecond != expectedRPS {
			t.Errorf("expected RPS %f, got %f", expectedRPS, loadTester.stats.RequestsPerSecond)
		}
	})
}

func TestStressTester(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultPerformanceTestConfig()

	// Use temporary directory for testing
	tempDir, err := os.MkdirTemp("", "stress_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.TestDirectory = tempDir
	config.StressTestDuration = 5 * time.Second // Short duration for testing
	config.StressTestMaxRPS = 100               // Low max RPS for testing
	config.StressTestStep = 25                  // Small step for testing

	stressTester := NewStressTester(config, logger)

	t.Run("test RPS level", func(t *testing.T) {
		ctx := context.Background()
		errorRate, avgResponseTime := stressTester.testRPSLevel(ctx, 50)

		// These are mock values, so we just check they're reasonable
		if errorRate < 0 || errorRate > 1 {
			t.Errorf("error rate should be between 0 and 1, got %f", errorRate)
		}
		if avgResponseTime <= 0 {
			t.Errorf("average response time should be positive, got %v", avgResponseTime)
		}
	})
}

func TestBenchmarker(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultPerformanceTestConfig()

	// Use temporary directory for testing
	tempDir, err := os.MkdirTemp("", "benchmark_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.TestDirectory = tempDir
	config.BenchmarkIterations = 10 // Small number for testing
	config.BenchmarkWarmup = 10 * time.Millisecond
	config.BenchmarkCooldown = 10 * time.Millisecond

	benchmarker := NewBenchmarker(config, logger)

	t.Run("execute operation", func(t *testing.T) {
		operations := []string{"classification", "verification", "data_extraction", "risk_assessment"}

		for _, operation := range operations {
			start := time.Now()
			benchmarker.executeOperation(operation)
			duration := time.Since(start)

			// Check that operation took reasonable time (mock implementations)
			if duration < 5*time.Millisecond {
				t.Errorf("operation %s took too little time: %v", operation, duration)
			}
			if duration > 50*time.Millisecond {
				t.Errorf("operation %s took too much time: %v", operation, duration)
			}
		}
	})

	t.Run("calculate benchmark stats", func(t *testing.T) {
		stats := &BenchmarkStats{
			OperationName: "test_operation",
			Iterations:    5,
			TotalTime:     1 * time.Second,
		}

		responseTimes := []time.Duration{
			100 * time.Millisecond,
			150 * time.Millisecond,
			200 * time.Millisecond,
			250 * time.Millisecond,
			300 * time.Millisecond,
		}

		benchmarker.calculateBenchmarkStats(stats, responseTimes)

		if stats.P50Time != 200*time.Millisecond {
			t.Errorf("expected P50 %v, got %v", 200*time.Millisecond, stats.P50Time)
		}
		if stats.P95Time != 250*time.Millisecond {
			t.Errorf("expected P95 %v, got %v", 250*time.Millisecond, stats.P95Time)
		}
		if stats.P99Time != 250*time.Millisecond {
			t.Errorf("expected P99 %v, got %v", 250*time.Millisecond, stats.P99Time)
		}
		if stats.AverageTime != 200*time.Millisecond {
			t.Errorf("expected average %v, got %v", 200*time.Millisecond, stats.AverageTime)
		}
		if stats.OperationsPerSecond != 5.0 {
			t.Errorf("expected ops/sec %f, got %f", 5.0, stats.OperationsPerSecond)
		}
	})
}

func TestTestReporter(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultPerformanceTestConfig()

	// Use temporary directory for testing
	tempDir, err := os.MkdirTemp("", "reporter_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config.ReportDirectory = tempDir
	config.ReportRetention = 3

	reporter := NewTestReporter(config, logger)

	t.Run("generate report", func(t *testing.T) {
		results := &TestResults{
			OverallScore:    85.5,
			Recommendations: []string{"Optimize response time", "Reduce error rate"},
			GeneratedAt:     time.Now(),
		}

		err := reporter.GenerateReport(results)
		if err != nil {
			t.Errorf("expected no error generating report, got %v", err)
		}

		// Check that report file was created
		files, err := os.ReadDir(tempDir)
		if err != nil {
			t.Errorf("failed to read report directory: %v", err)
		}

		reportFound := false
		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
				reportFound = true
				break
			}
		}

		if !reportFound {
			t.Error("expected report file to be created")
		}
	})

	t.Run("cleanup old reports", func(t *testing.T) {
		// Create some test report files
		for i := 0; i < 5; i++ {
			filename := filepath.Join(tempDir, fmt.Sprintf("old_report_%d.json", i))
			if err := os.WriteFile(filename, []byte("{}"), 0644); err != nil {
				t.Errorf("failed to create test report file: %v", err)
			}
		}

		// Run cleanup
		reporter.cleanupOldReports()

		// Check that only retention number of files remain
		files, err := os.ReadDir(tempDir)
		if err != nil {
			t.Errorf("failed to read report directory: %v", err)
		}

		jsonFileCount := 0
		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
				jsonFileCount++
			}
		}

		if jsonFileCount > config.ReportRetention {
			t.Errorf("expected at most %d report files, got %d", config.ReportRetention, jsonFileCount)
		}
	})
}

func TestTestMonitor(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultPerformanceTestConfig()

	config.MonitorInterval = 10 * time.Millisecond // Short interval for testing

	monitor := NewTestMonitor(config, logger)

	t.Run("collect metrics", func(t *testing.T) {
		monitor.collectMetrics()

		if monitor.stats.LastUpdated.IsZero() {
			t.Error("expected last updated time to be set")
		}

		if monitor.stats.ResourceUsage == nil {
			t.Error("expected resource usage to be collected")
		}

		// Check that memory usage is collected
		if memoryUsage, exists := monitor.stats.ResourceUsage["memory_usage_mb"]; !exists {
			t.Error("expected memory usage to be collected")
		} else if memoryUsage < 0 {
			t.Errorf("memory usage should be positive, got %f", memoryUsage)
		}
	})

	t.Run("start and stop monitor", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// Start monitor in goroutine
		go monitor.Start(ctx)

		// Wait for context cancellation
		<-ctx.Done()
	})
}

func TestPerformanceTestConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := DefaultPerformanceTestConfig()

		if !config.TestEnabled {
			t.Error("expected test to be enabled by default")
		}
		if config.TestTimeout != 30*time.Minute {
			t.Errorf("expected test timeout %v, got %v", 30*time.Minute, config.TestTimeout)
		}
		if config.TestParallelism <= 0 {
			t.Error("expected test parallelism to be positive")
		}
		if !config.LoadTestEnabled {
			t.Error("expected load test to be enabled by default")
		}
		if !config.StressTestEnabled {
			t.Error("expected stress test to be enabled by default")
		}
		if !config.BenchmarkEnabled {
			t.Error("expected benchmark to be enabled by default")
		}
		if !config.ReportEnabled {
			t.Error("expected report to be enabled by default")
		}
		if !config.MonitorEnabled {
			t.Error("expected monitor to be enabled by default")
		}
	})
}

func TestTestResults(t *testing.T) {
	t.Run("test results creation", func(t *testing.T) {
		results := &TestResults{
			LoadTestResults: &LoadTestStats{
				TotalRequests:      1000,
				SuccessfulRequests: 950,
				FailedRequests:     50,
				ErrorRate:          0.05,
				RequestsPerSecond:  100.0,
				P95ResponseTime:    500 * time.Millisecond,
			},
			StressTestResults: &StressTestStats{
				MaxRPS:         500,
				BreakingPoint:  450,
				ErrorRateAtMax: 0.08,
			},
			BenchmarkResults: []*BenchmarkStats{
				{
					OperationName:       "classification",
					OperationsPerSecond: 1500.0,
				},
				{
					OperationName:       "verification",
					OperationsPerSecond: 1200.0,
				},
			},
			OverallScore:    85.5,
			Recommendations: []string{"Optimize response time", "Reduce error rate"},
			GeneratedAt:     time.Now(),
		}

		if results.LoadTestResults == nil {
			t.Error("expected load test results to be set")
		}
		if results.StressTestResults == nil {
			t.Error("expected stress test results to be set")
		}
		if len(results.BenchmarkResults) == 0 {
			t.Error("expected benchmark results to be set")
		}
		if results.OverallScore <= 0 {
			t.Error("expected overall score to be positive")
		}
		if len(results.Recommendations) == 0 {
			t.Error("expected recommendations to be set")
		}
		if results.GeneratedAt.IsZero() {
			t.Error("expected generated at time to be set")
		}
	})
}

func TestLoadTestStats(t *testing.T) {
	t.Run("load test stats creation", func(t *testing.T) {
		stats := &LoadTestStats{
			TotalRequests:       1000,
			SuccessfulRequests:  950,
			FailedRequests:      50,
			AverageResponseTime: 200 * time.Millisecond,
			MinResponseTime:     50 * time.Millisecond,
			MaxResponseTime:     1000 * time.Millisecond,
			P50ResponseTime:     180 * time.Millisecond,
			P95ResponseTime:     500 * time.Millisecond,
			P99ResponseTime:     800 * time.Millisecond,
			RequestsPerSecond:   100.0,
			ErrorRate:           0.05,
			StartTime:           time.Now(),
			EndTime:             time.Now().Add(10 * time.Second),
			Duration:            10 * time.Second,
		}

		if stats.TotalRequests != 1000 {
			t.Errorf("expected total requests 1000, got %d", stats.TotalRequests)
		}
		if stats.SuccessfulRequests != 950 {
			t.Errorf("expected successful requests 950, got %d", stats.SuccessfulRequests)
		}
		if stats.FailedRequests != 50 {
			t.Errorf("expected failed requests 50, got %d", stats.FailedRequests)
		}
		if stats.ErrorRate != 0.05 {
			t.Errorf("expected error rate 0.05, got %f", stats.ErrorRate)
		}
		if stats.RequestsPerSecond != 100.0 {
			t.Errorf("expected RPS 100.0, got %f", stats.RequestsPerSecond)
		}
		if stats.Duration != 10*time.Second {
			t.Errorf("expected duration 10s, got %v", stats.Duration)
		}
	})
}

func TestStressTestStats(t *testing.T) {
	t.Run("stress test stats creation", func(t *testing.T) {
		stats := &StressTestStats{
			MaxRPS:            500,
			BreakingPoint:     450,
			ResponseTimeAtMax: 800 * time.Millisecond,
			ErrorRateAtMax:    0.08,
			ResourceUtilization: map[string]float64{
				"cpu":    75.5,
				"memory": 60.2,
			},
			StartTime: time.Now(),
			EndTime:   time.Now().Add(5 * time.Minute),
			Duration:  5 * time.Minute,
		}

		if stats.MaxRPS != 500 {
			t.Errorf("expected max RPS 500, got %d", stats.MaxRPS)
		}
		if stats.BreakingPoint != 450 {
			t.Errorf("expected breaking point 450, got %d", stats.BreakingPoint)
		}
		if stats.ErrorRateAtMax != 0.08 {
			t.Errorf("expected error rate at max 0.08, got %f", stats.ErrorRateAtMax)
		}
		if len(stats.ResourceUtilization) == 0 {
			t.Error("expected resource utilization to be set")
		}
		if stats.Duration != 5*time.Minute {
			t.Errorf("expected duration 5m, got %v", stats.Duration)
		}
	})
}

func TestBenchmarkStats(t *testing.T) {
	t.Run("benchmark stats creation", func(t *testing.T) {
		stats := &BenchmarkStats{
			OperationName:       "classification",
			Iterations:          1000,
			TotalTime:           2 * time.Second,
			AverageTime:         2 * time.Millisecond,
			MinTime:             1 * time.Millisecond,
			MaxTime:             10 * time.Millisecond,
			P50Time:             2 * time.Millisecond,
			P95Time:             5 * time.Millisecond,
			P99Time:             8 * time.Millisecond,
			OperationsPerSecond: 500.0,
			MemoryUsage:         1024 * 1024, // 1MB
			CPUUsage:            25.5,
			StartTime:           time.Now(),
			EndTime:             time.Now().Add(2 * time.Second),
		}

		if stats.OperationName != "classification" {
			t.Errorf("expected operation name 'classification', got %s", stats.OperationName)
		}
		if stats.Iterations != 1000 {
			t.Errorf("expected iterations 1000, got %d", stats.Iterations)
		}
		if stats.OperationsPerSecond != 500.0 {
			t.Errorf("expected ops/sec 500.0, got %f", stats.OperationsPerSecond)
		}
		if stats.MemoryUsage != 1024*1024 {
			t.Errorf("expected memory usage 1MB, got %d", stats.MemoryUsage)
		}
		if stats.CPUUsage != 25.5 {
			t.Errorf("expected CPU usage 25.5, got %f", stats.CPUUsage)
		}
	})
}

func BenchmarkPerformanceTestManager_GetResults(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultPerformanceTestConfig()
	manager := NewPerformanceTestManager(config, logger)
	defer manager.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetResults()
	}
}

func BenchmarkLoadTester_CalculatePercentiles(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultPerformanceTestConfig()
	loadTester := NewPerformanceLoadTester(config, logger)

	// Create test data
	responseTimes := make([]time.Duration, 1000)
	for i := range responseTimes {
		responseTimes[i] = time.Duration(i) * time.Millisecond
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loadTester.calculatePercentiles(responseTimes)
	}
}

func BenchmarkBenchmarker_CalculateBenchmarkStats(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultPerformanceTestConfig()
	benchmarker := NewBenchmarker(config, logger)

	stats := &BenchmarkStats{
		OperationName: "test",
		Iterations:    1000,
		TotalTime:     1 * time.Second,
	}

	responseTimes := make([]time.Duration, 1000)
	for i := range responseTimes {
		responseTimes[i] = time.Duration(i) * time.Millisecond
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarker.calculateBenchmarkStats(stats, responseTimes)
	}
}
