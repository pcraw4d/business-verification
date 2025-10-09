package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/loadtesting"
)

func main() {
	// Command line flags
	var (
		baseURL         = flag.String("url", "http://localhost:8080", "Base URL of the risk assessment service")
		duration        = flag.Duration("duration", 5*time.Minute, "Duration of the load test")
		concurrentUsers = flag.Int("users", 10, "Number of concurrent users")
		requestsPerUser = flag.Int("requests", 100, "Number of requests per user")
		rampUpTime      = flag.Duration("rampup", 30*time.Second, "Ramp-up time for concurrent users")
		targetRPS       = flag.Float64("rps", 16.67, "Target requests per second (1000 req/min = 16.67 RPS)")
		timeout         = flag.Duration("timeout", 30*time.Second, "Request timeout")
		testType        = flag.String("type", "load", "Type of test: load, stress, spike")
		outputFile      = flag.String("output", "", "Output file for test results (JSON format)")
		verbose         = flag.Bool("verbose", false, "Enable verbose logging")
	)

	flag.Parse()

	// Setup logging
	var logger *zap.Logger
	var err error
	if *verbose {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("🚀 Starting Load Test Tool",
		zap.String("url", *baseURL),
		zap.String("type", *testType),
		zap.Duration("duration", *duration),
		zap.Int("concurrent_users", *concurrentUsers),
		zap.Float64("target_rps", *targetRPS))

	// Create load tester
	loadTester := loadtesting.NewLoadTester(logger, *baseURL)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *duration+5*time.Minute)
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("🛑 Interrupt signal received, stopping load test...")
		cancel()
	}()

	// Load test configuration
	config := loadtesting.LoadTestConfig{
		Duration:        *duration,
		ConcurrentUsers: *concurrentUsers,
		RequestsPerUser: *requestsPerUser,
		RampUpTime:      *rampUpTime,
		TargetRPS:       *targetRPS,
		Timeout:         *timeout,
	}

	// Run the appropriate test
	var result *loadtesting.LoadTestResult
	switch *testType {
	case "load":
		result, err = runLoadTest(ctx, loadTester, config, logger)
	case "stress":
		result, err = runStressTest(ctx, loadTester, config, logger)
	case "spike":
		result, err = runSpikeTest(ctx, loadTester, config, logger)
	default:
		logger.Fatal("Invalid test type", zap.String("type", *testType))
	}

	if err != nil {
		logger.Fatal("Load test failed", zap.Error(err))
	}

	// Display results
	displayResults(result, logger)

	// Save results to file if specified
	if *outputFile != "" {
		if err := saveResults(result, *outputFile, logger); err != nil {
			logger.Error("Failed to save results", zap.Error(err))
		}
	}

	// Check if test passed
	if result.ErrorRate > 0.05 { // 5% error rate threshold
		logger.Error("❌ Load test failed - High error rate",
			zap.Float64("error_rate", result.ErrorRate))
		os.Exit(1)
	}

	if result.RequestsPerMinute < *targetRPS*60*0.8 { // 80% of target throughput
		logger.Error("❌ Load test failed - Low throughput",
			zap.Float64("actual_rpm", result.RequestsPerMinute),
			zap.Float64("target_rpm", *targetRPS*60))
		os.Exit(1)
	}

	logger.Info("✅ Load test passed successfully!")
}

// runLoadTest runs a standard load test
func runLoadTest(ctx context.Context, loadTester *loadtesting.LoadTester, config loadtesting.LoadTestConfig, logger *zap.Logger) (*loadtesting.LoadTestResult, error) {
	logger.Info("📊 Running Load Test",
		zap.Duration("duration", config.Duration),
		zap.Int("concurrent_users", config.ConcurrentUsers),
		zap.Float64("target_rps", config.TargetRPS))

	return loadTester.RunLoadTest(ctx, config)
}

// runStressTest runs a stress test
func runStressTest(ctx context.Context, loadTester *loadtesting.LoadTester, config loadtesting.LoadTestConfig, logger *zap.Logger) (*loadtesting.LoadTestResult, error) {
	logger.Info("🔥 Running Stress Test",
		zap.Duration("duration", config.Duration),
		zap.Int("concurrent_users", config.ConcurrentUsers),
		zap.Float64("target_rps", config.TargetRPS))

	return loadTester.RunStressTest(ctx, config)
}

// runSpikeTest runs a spike test
func runSpikeTest(ctx context.Context, loadTester *loadtesting.LoadTester, config loadtesting.LoadTestConfig, logger *zap.Logger) (*loadtesting.LoadTestResult, error) {
	logger.Info("⚡ Running Spike Test",
		zap.Duration("duration", config.Duration),
		zap.Int("concurrent_users", config.ConcurrentUsers),
		zap.Float64("target_rps", config.TargetRPS))

	return loadTester.RunSpikeTest(ctx, config)
}

// displayResults displays test results in a formatted way
func displayResults(result *loadtesting.LoadTestResult, logger *zap.Logger) {
	logger.Info("📈 Load Test Results",
		zap.String("=", "====================================="))

	logger.Info("📊 Request Statistics",
		zap.Int64("total_requests", result.TotalRequests),
		zap.Int64("successful_requests", result.SuccessfulRequests),
		zap.Int64("failed_requests", result.FailedRequests),
		zap.Float64("error_rate_percent", result.ErrorRate*100))

	logger.Info("⏱️  Performance Metrics",
		zap.Duration("total_duration", result.TotalDuration),
		zap.Duration("average_response_time", result.AverageResponseTime),
		zap.Duration("min_response_time", result.MinResponseTime),
		zap.Duration("max_response_time", result.MaxResponseTime))

	logger.Info("🚀 Throughput Metrics",
		zap.Float64("requests_per_second", result.RequestsPerSecond),
		zap.Float64("requests_per_minute", result.RequestsPerMinute))

	// Display error summary
	if len(result.Errors) > 0 {
		logger.Info("❌ Error Summary",
			zap.Int("error_count", len(result.Errors)))

		// Show first 5 errors
		for i, err := range result.Errors {
			if i >= 5 {
				break
			}
			logger.Info("Error Details",
				zap.String("request_id", err.RequestID),
				zap.String("error", err.Error),
				zap.Duration("duration", err.Duration))
		}

		if len(result.Errors) > 5 {
			logger.Info("... and more errors", zap.Int("remaining", len(result.Errors)-5))
		}
	}

	// Performance assessment
	logger.Info("🎯 Performance Assessment")
	if result.ErrorRate < 0.01 {
		logger.Info("✅ Excellent error rate", zap.Float64("error_rate_percent", result.ErrorRate*100))
	} else if result.ErrorRate < 0.05 {
		logger.Info("⚠️  Acceptable error rate", zap.Float64("error_rate_percent", result.ErrorRate*100))
	} else {
		logger.Info("❌ High error rate", zap.Float64("error_rate_percent", result.ErrorRate*100))
	}

	if result.AverageResponseTime < 500*time.Millisecond {
		logger.Info("✅ Excellent response time", zap.Duration("avg_response_time", result.AverageResponseTime))
	} else if result.AverageResponseTime < 1*time.Second {
		logger.Info("⚠️  Acceptable response time", zap.Duration("avg_response_time", result.AverageResponseTime))
	} else {
		logger.Info("❌ Slow response time", zap.Duration("avg_response_time", result.AverageResponseTime))
	}

	if result.RequestsPerMinute >= 1000 {
		logger.Info("✅ Target throughput achieved", zap.Float64("rpm", result.RequestsPerMinute))
	} else {
		logger.Info("⚠️  Below target throughput", zap.Float64("rpm", result.RequestsPerMinute))
	}
}

// saveResults saves test results to a JSON file
func saveResults(result *loadtesting.LoadTestResult, filename string, logger *zap.Logger) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode results: %w", err)
	}

	logger.Info("💾 Results saved to file", zap.String("filename", filename))
	return nil
}
