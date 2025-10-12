package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/loadtesting"
)

func mainPerformanceTest() {
	// Command line flags
	var (
		baseURL         = flag.String("url", "http://localhost:8080", "Base URL of the risk assessment service")
		duration        = flag.Duration("duration", 5*time.Minute, "Test duration")
		concurrentUsers = flag.Int("users", 100, "Number of concurrent users")
		targetRPS       = flag.Float64("rps", 100, "Target requests per second")
		targetRPM       = flag.Float64("rpm", 5000, "Target requests per minute")
		testPattern     = flag.String("pattern", "constant", "Test pattern: constant, ramp, spike, sine")
		rampUpTime      = flag.Duration("ramp-up", 1*time.Minute, "Ramp up time")
		rampDownTime    = flag.Duration("ramp-down", 1*time.Minute, "Ramp down time")
		steadyStateTime = flag.Duration("steady-state", 3*time.Minute, "Steady state time")
		maxLatency      = flag.Duration("max-latency", 2*time.Second, "Maximum acceptable latency")
		maxErrorRate    = flag.Float64("max-error-rate", 0.01, "Maximum acceptable error rate (0.01 = 1%)")
		spikeMultiplier = flag.Float64("spike-multiplier", 5.0, "Spike multiplier for spike tests")
		sineAmplitude   = flag.Float64("sine-amplitude", 0.5, "Sine wave amplitude (0.5 = 50% variation)")
		sinePeriod      = flag.Duration("sine-period", 2*time.Minute, "Sine wave period")
		verbose         = flag.Bool("verbose", false, "Verbose output")
		outputFile      = flag.String("output", "", "Output file for results (JSON format)")
	)
	flag.Parse()

	// Initialize logger
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

	logger.Info("🚀 Starting Enhanced Risk Assessment Service Performance Test",
		zap.String("base_url", *baseURL),
		zap.Duration("duration", *duration),
		zap.Int("concurrent_users", *concurrentUsers),
		zap.Float64("target_rps", *targetRPS),
		zap.Float64("target_rpm", *targetRPM),
		zap.String("test_pattern", *testPattern))

	// Create enhanced load tester
	loadTester := loadtesting.NewEnhancedLoadTester(logger, *baseURL)

	// Create test configuration
	config := loadtesting.EnhancedLoadTestConfig{
		Duration:           *duration,
		ConcurrentUsers:    *concurrentUsers,
		TargetRPS:          *targetRPS,
		TargetRPM:          *targetRPM,
		RampUpTime:         *rampUpTime,
		RampDownTime:       *rampDownTime,
		SteadyStateTime:    *steadyStateTime,
		MaxLatency:         *maxLatency,
		MaxErrorRate:       *maxErrorRate,
		TestPattern:        *testPattern,
		SpikeMultiplier:    *spikeMultiplier,
		SineAmplitude:      *sineAmplitude,
		SinePeriod:         *sinePeriod,
		ConnectionPoolSize: 100,
		RequestTimeout:     30 * time.Second,
		KeepAliveTimeout:   90 * time.Second,
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("Received interrupt signal, stopping test...")
		cancel()
	}()

	// Print test configuration
	printTestConfiguration(config)

	// Run the performance test
	logger.Info("Starting performance test...")
	startTime := time.Now()

	metrics, err := loadTester.RunHighPerformanceLoadTest(ctx, config)
	if err != nil {
		logger.Fatal("Performance test failed", zap.Error(err))
	}

	testDuration := time.Since(startTime)
	logger.Info("Performance test completed", zap.Duration("test_duration", testDuration))

	// Print results
	printTestResults(metrics, config)

	// Check if targets were met
	checkTargets(metrics, config, logger)

	// Save results to file if requested
	if *outputFile != "" {
		saveResultsToFile(metrics, *outputFile, logger)
	}

	// Exit with appropriate code
	if metrics.IsTargetMet {
		logger.Info("✅ All performance targets met!")
		os.Exit(0)
	} else {
		logger.Warn("❌ Some performance targets were not met")
		os.Exit(1)
	}
}

func printTestConfiguration(config loadtesting.EnhancedLoadTestConfig) {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("ENHANCED RISK ASSESSMENT SERVICE PERFORMANCE TEST")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Test Pattern:     %s\n", config.TestPattern)
	fmt.Printf("Duration:         %v\n", config.Duration)
	fmt.Printf("Concurrent Users: %d\n", config.ConcurrentUsers)
	fmt.Printf("Target RPS:       %.2f\n", config.TargetRPS)
	fmt.Printf("Target RPM:       %.2f\n", config.TargetRPM)
	fmt.Printf("Max Latency:      %v\n", config.MaxLatency)
	fmt.Printf("Max Error Rate:   %.2f%%\n", config.MaxErrorRate*100)

	if config.TestPattern == "ramp" {
		fmt.Printf("Ramp Up Time:     %v\n", config.RampUpTime)
		fmt.Printf("Steady State:     %v\n", config.SteadyStateTime)
		fmt.Printf("Ramp Down Time:   %v\n", config.RampDownTime)
	} else if config.TestPattern == "spike" {
		fmt.Printf("Spike Multiplier: %.2fx\n", config.SpikeMultiplier)
	} else if config.TestPattern == "sine" {
		fmt.Printf("Sine Amplitude:   %.2f\n", config.SineAmplitude)
		fmt.Printf("Sine Period:      %v\n", config.SinePeriod)
	}

	fmt.Println(strings.Repeat("=", 80))
}

func printTestResults(metrics *loadtesting.LoadTestMetrics, config loadtesting.EnhancedLoadTestConfig) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("PERFORMANCE TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	// Request metrics
	fmt.Printf("Total Requests:      %d\n", metrics.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", metrics.SuccessfulRequests)
	fmt.Printf("Failed Requests:     %d\n", metrics.FailedRequests)
	fmt.Printf("Error Rate:          %.2f%%\n", metrics.ErrorRate*100)

	// Timing metrics
	fmt.Printf("Total Duration:      %v\n", metrics.TotalDuration)
	fmt.Printf("Min Response Time:   %v\n", metrics.MinResponseTime)
	fmt.Printf("Max Response Time:   %v\n", metrics.MaxResponseTime)

	// Throughput metrics
	fmt.Printf("Requests/Second:     %.2f\n", metrics.RequestsPerSecond)
	fmt.Printf("Requests/Minute:     %.2f\n", metrics.RequestsPerMinute)
	fmt.Printf("Peak RPS:            %.2f\n", metrics.PeakRPS)

	// System metrics
	fmt.Printf("Memory Usage:        %d MB\n", metrics.MemoryUsage/(1024*1024))
	fmt.Printf("CPU Usage:           %.2f%%\n", metrics.CPUUsage)
	fmt.Printf("Goroutine Count:     %d\n", metrics.GoroutineCount)

	// Performance score
	fmt.Printf("Performance Score:   %.2f/100\n", metrics.PerformanceScore)
	fmt.Printf("Targets Met:         %t\n", metrics.IsTargetMet)

	fmt.Println(strings.Repeat("=", 80))
}

func checkTargets(metrics *loadtesting.LoadTestMetrics, config loadtesting.EnhancedLoadTestConfig, logger *zap.Logger) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TARGET VALIDATION")
	fmt.Println(strings.Repeat("=", 80))

	// Check throughput target
	throughputMet := metrics.RequestsPerMinute >= config.TargetRPM
	fmt.Printf("Throughput Target:   %.0f RPM → %.0f RPM [%s]\n",
		config.TargetRPM, metrics.RequestsPerMinute, getStatusIcon(throughputMet))

	// Check latency target
	latencyMet := metrics.MaxResponseTime <= config.MaxLatency
	fmt.Printf("Latency Target:      %v → %v [%s]\n",
		config.MaxLatency, metrics.MaxResponseTime, getStatusIcon(latencyMet))

	// Check error rate target
	errorRateMet := metrics.ErrorRate <= config.MaxErrorRate
	fmt.Printf("Error Rate Target:   %.2f%% → %.2f%% [%s]\n",
		config.MaxErrorRate*100, metrics.ErrorRate*100, getStatusIcon(errorRateMet))

	// Overall assessment
	allTargetsMet := throughputMet && latencyMet && errorRateMet
	fmt.Printf("\nOverall Result:      [%s]\n", getStatusIcon(allTargetsMet))

	if allTargetsMet {
		logger.Info("🎉 All performance targets achieved!")
	} else {
		logger.Warn("⚠️  Some performance targets were not met")
	}

	fmt.Println(strings.Repeat("=", 80))
}

func getStatusIcon(met bool) string {
	if met {
		return "✅ PASS"
	}
	return "❌ FAIL"
}

func saveResultsToFile(metrics *loadtesting.LoadTestMetrics, filename string, logger *zap.Logger) {
	// In a real implementation, this would save the metrics to a JSON file
	logger.Info("Saving results to file", zap.String("filename", filename))
	// TODO: Implement JSON serialization and file writing
}
