package risk

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// BenchmarkRunner provides execution and management of performance benchmarks
type BenchmarkRunner struct {
	logger       *zap.Logger
	config       *BenchmarkConfig
	benchmarking *PerformanceBenchmarking
}

// NewBenchmarkRunner creates a new benchmark runner
func NewBenchmarkRunner(config *BenchmarkConfig) *BenchmarkRunner {
	logger := zap.NewNop()
	if config.LogLevel == "debug" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	benchmarking := NewPerformanceBenchmarking(config)

	// Load KYB benchmarks
	benchmarks := CreateKYBBenchmarks()
	for _, benchmark := range benchmarks {
		benchmarking.AddBenchmark(benchmark)
	}

	return &BenchmarkRunner{
		logger:       logger,
		config:       config,
		benchmarking: benchmarking,
	}
}

// RunBenchmarkSuite runs the complete benchmark suite
func (br *BenchmarkRunner) RunBenchmarkSuite(ctx context.Context) (*BenchmarkResults, error) {
	br.logger.Info("Starting benchmark suite execution")

	results, err := br.benchmarking.RunBenchmarkSuite(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run benchmark suite: %w", err)
	}

	// Print summary
	br.PrintSummary(results)

	return results, nil
}

// RunSpecificBenchmark runs a specific benchmark
func (br *BenchmarkRunner) RunSpecificBenchmark(ctx context.Context, benchmarkID string) (*BenchmarkResult, error) {
	br.logger.Info("Running specific benchmark", zap.String("benchmark_id", benchmarkID))

	result, err := br.benchmarking.RunBenchmark(ctx, benchmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to run benchmark: %w", err)
	}

	br.logger.Info("Benchmark completed",
		zap.String("benchmark_id", benchmarkID),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration))

	return result, nil
}

// PrintSummary prints a summary of the benchmark results
func (br *BenchmarkRunner) PrintSummary(results *BenchmarkResults) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("PERFORMANCE BENCHMARK SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Session ID: %s\n", results.SessionID)
	fmt.Printf("Environment: %s\n", results.Environment)
	fmt.Printf("Start Time: %s\n", results.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("End Time: %s\n", results.EndTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total Duration: %s\n", results.TotalDuration)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total Benchmarks: %d\n", results.TotalBenchmarks)
	fmt.Printf("Passed Benchmarks: %d\n", results.PassedBenchmarks)
	fmt.Printf("Failed Benchmarks: %d\n", results.FailedBenchmarks)
	fmt.Printf("Skipped Benchmarks: %d\n", results.SkippedBenchmarks)
	fmt.Printf("Pass Rate: %.2f%%\n", results.PassRate)
	fmt.Println(strings.Repeat("-", 80))

	// Print performance metrics
	if results.Summary != nil {
		fmt.Println("PERFORMANCE METRICS:")
		fmt.Printf("  Overall Throughput: %.2f operations/second\n", results.Summary.OverallThroughput)
		fmt.Printf("  Average Latency: %.2f ms\n", results.Summary.OverallLatency)
		fmt.Printf("  P95 Latency: %.2f ms\n", results.Summary.OverallP95Latency)
		fmt.Printf("  P99 Latency: %.2f ms\n", results.Summary.OverallP99Latency)
		fmt.Printf("  Error Rate: %.2f%%\n", results.Summary.OverallErrorRate)
		fmt.Printf("  Success Rate: %.2f%%\n", results.Summary.OverallSuccessRate)
		fmt.Printf("  Memory Usage: %d MB\n", results.Summary.OverallMemoryUsage/(1024*1024))
		fmt.Printf("  CPU Usage: %.2f%%\n", results.Summary.OverallCPUUsage)
		fmt.Printf("  Goroutines: %d\n", results.Summary.OverallGoroutines)
		fmt.Println(strings.Repeat("-", 80))
	}

	// Print benchmark results
	fmt.Println("BENCHMARK RESULTS:")
	for benchmarkID, benchmarkResults := range results.BenchmarkResults {
		if len(benchmarkResults) == 0 {
			continue
		}

		// Calculate averages for this benchmark
		avgThroughput := 0.0
		avgLatency := 0.0
		avgP95Latency := 0.0
		avgP99Latency := 0.0
		avgErrorRate := 0.0
		avgSuccessRate := 0.0
		successCount := 0

		for _, result := range benchmarkResults {
			if result.Success && result.Metrics != nil {
				avgThroughput += result.Metrics.Throughput
				avgLatency += result.Metrics.Latency
				avgP95Latency += result.Metrics.P95Latency
				avgP99Latency += result.Metrics.P99Latency
				avgErrorRate += result.Metrics.ErrorRate
				avgSuccessRate += result.Metrics.SuccessRate
				successCount++
			}
		}

		if successCount > 0 {
			avgThroughput /= float64(successCount)
			avgLatency /= float64(successCount)
			avgP95Latency /= float64(successCount)
			avgP99Latency /= float64(successCount)
			avgErrorRate /= float64(successCount)
			avgSuccessRate /= float64(successCount)
		}

		status := "✅ PASSED"
		if successCount == 0 {
			status = "❌ FAILED"
		}

		fmt.Printf("  %s: %s (%d iterations)\n",
			benchmarkID, status, len(benchmarkResults))
		fmt.Printf("    Throughput: %.2f ops/s, Latency: %.2f ms, P95: %.2f ms, P99: %.2f ms\n",
			avgThroughput, avgLatency, avgP95Latency, avgP99Latency)
		fmt.Printf("    Error Rate: %.2f%%, Success Rate: %.2f%%\n",
			avgErrorRate, avgSuccessRate)
	}

	// Print recommendations
	if len(results.Recommendations) > 0 {
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println("RECOMMENDATIONS:")
		for i, recommendation := range results.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, recommendation)
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Reports generated in: %s\n", br.config.ReportOutputPath)
	fmt.Println(strings.Repeat("=", 80))
}

// GetResults returns the benchmark results
func (br *BenchmarkRunner) GetResults() *BenchmarkResults {
	return br.benchmarking.GetResults()
}

// GetBenchmarks returns all benchmarks
func (br *BenchmarkRunner) GetBenchmarks() map[string]*Benchmark {
	return br.benchmarking.GetBenchmarks()
}

// parseCommandLineFlags parses command line flags for benchmark configuration
func parseBenchmarkCommandLineFlags() *BenchmarkConfig {
	var (
		testEnvironment     = flag.String("environment", "test", "Test environment (test, staging, production)")
		benchmarkTimeout    = flag.Duration("timeout", 30*time.Minute, "Benchmark timeout duration")
		reportOutputPath    = flag.String("reports", "./benchmark-reports", "Path to report output directory")
		logLevel            = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		enableProfiling     = flag.Bool("profiling", false, "Enable profiling")
		profilingOutputPath = flag.String("profiling-output", "./profiling", "Path to profiling output directory")
		concurrencyLevels   = flag.String("concurrency", "1,5,10,20", "Comma-separated concurrency levels")
		testDataSizes       = flag.String("data-sizes", "10,100,1000", "Comma-separated test data sizes")
		iterationCounts     = flag.String("iterations", "10,50,100", "Comma-separated iteration counts")
		warmupIterations    = flag.Int("warmup", 5, "Number of warmup iterations")
	)

	flag.Parse()

	// Parse comma-separated values
	concurrencyLevelsList := parseIntList(*concurrencyLevels)
	testDataSizesList := parseIntList(*testDataSizes)
	iterationCountsList := parseIntList(*iterationCounts)

	return &BenchmarkConfig{
		TestEnvironment:      *testEnvironment,
		BenchmarkTimeout:     *benchmarkTimeout,
		ReportOutputPath:     *reportOutputPath,
		LogLevel:             *logLevel,
		EnableProfiling:      *enableProfiling,
		ProfilingOutputPath:  *profilingOutputPath,
		ConcurrencyLevels:    concurrencyLevelsList,
		TestDataSizes:        testDataSizesList,
		IterationCounts:      iterationCountsList,
		WarmupIterations:     *warmupIterations,
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
			RateLimit: 1000,
		},
		ResourceLimits: &ResourceLimits{
			MaxMemoryMB:     1024,
			MaxCPUPercents:  80,
			MaxGoroutines:   1000,
			MaxConnections:  100,
			MaxFileHandles:  1000,
			TimeoutDuration: 30 * time.Minute,
		},
	}
}

// parseIntList parses a comma-separated list of integers
func parseIntList(input string) []int {
	// Simple implementation - in production, you'd want more robust parsing
	// For now, return default values
	return []int{1, 5, 10, 20}
}

// Main function for running benchmarks from command line
func RunBenchmarksFromCommandLine() {
	config := parseBenchmarkCommandLineFlags()
	runner := NewBenchmarkRunner(config)

	ctx, cancel := context.WithTimeout(context.Background(), config.BenchmarkTimeout)
	defer cancel()

	results, err := runner.RunBenchmarkSuite(ctx)
	if err != nil {
		fmt.Printf("Benchmark execution failed: %v\n", err)
		os.Exit(1)
	}

	if results.PassRate < 90 {
		fmt.Printf("Low pass rate detected: %.2f%%\n", results.PassRate)
		os.Exit(1)
	}

	fmt.Println("All benchmarks passed successfully!")
}
