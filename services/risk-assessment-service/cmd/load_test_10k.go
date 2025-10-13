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

// LoadTestConfig represents the configuration for 10K concurrent users load testing
type LoadTestConfig struct {
	BaseURL         string        `json:"base_url"`
	Users           int           `json:"users"`
	Duration        time.Duration `json:"duration"`
	RampUpTime      time.Duration `json:"ramp_up_time"`
	RequestsPerUser int           `json:"requests_per_user"`
	TargetRPS       float64       `json:"target_rps"`
	Timeout         time.Duration `json:"timeout"`
	OutputFile      string        `json:"output_file"`
	Quiet           bool          `json:"quiet"`
}

// LoadTestResult represents the result of a load test
type LoadTestResult struct {
	Config              LoadTestConfig              `json:"config"`
	StartTime           time.Time                   `json:"start_time"`
	EndTime             time.Time                   `json:"end_time"`
	Duration            time.Duration               `json:"duration"`
	TotalRequests       int64                       `json:"total_requests"`
	SuccessfulRequests  int64                       `json:"successful_requests"`
	FailedRequests      int64                       `json:"failed_requests"`
	ErrorRate           float64                     `json:"error_rate"`
	AverageResponseTime time.Duration               `json:"average_response_time"`
	MinResponseTime     time.Duration               `json:"min_response_time"`
	MaxResponseTime     time.Duration               `json:"max_response_time"`
	P50ResponseTime     time.Duration               `json:"p50_response_time"`
	P95ResponseTime     time.Duration               `json:"p95_response_time"`
	P99ResponseTime     time.Duration               `json:"p99_response_time"`
	RequestsPerSecond   float64                     `json:"requests_per_second"`
	RequestsPerMinute   float64                     `json:"requests_per_minute"`
	PeakRPS             float64                     `json:"peak_rps"`
	ResponseTimes       []time.Duration             `json:"response_times"`
	Errors              []loadtesting.LoadTestError `json:"errors"`
	Targets             PerformanceTargets          `json:"targets"`
	Results             PerformanceResults          `json:"results"`
}

// PerformanceTargets represents the performance targets for 10K concurrent users
type PerformanceTargets struct {
	P95Latency    time.Duration `json:"p95_latency"`
	P99Latency    time.Duration `json:"p99_latency"`
	ErrorRate     float64       `json:"error_rate"`
	ThroughputRPM float64       `json:"throughput_rpm"`
}

// PerformanceResults represents whether targets were met
type PerformanceResults struct {
	P95LatencyMet  bool `json:"p95_latency_met"`
	P99LatencyMet  bool `json:"p99_latency_met"`
	ErrorRateMet   bool `json:"error_rate_met"`
	ThroughputMet  bool `json:"throughput_met"`
	OverallSuccess bool `json:"overall_success"`
}

func mainLoadTest10K() {
	// Parse command line flags
	var config LoadTestConfig

	flag.StringVar(&config.BaseURL, "url", "http://localhost:8080", "Base URL for the risk assessment service")
	flag.IntVar(&config.Users, "users", 10000, "Number of concurrent users")
	flag.DurationVar(&config.Duration, "duration", 30*time.Minute, "Test duration")
	flag.DurationVar(&config.RampUpTime, "ramp-up", 5*time.Minute, "Ramp-up time")
	flag.IntVar(&config.RequestsPerUser, "requests-per-user", 10, "Requests per user")
	flag.Float64Var(&config.TargetRPS, "target-rps", 2000, "Target requests per second")
	flag.DurationVar(&config.Timeout, "timeout", 30*time.Second, "Request timeout")
	flag.StringVar(&config.OutputFile, "output", "", "Output file for results (JSON)")
	flag.BoolVar(&config.Quiet, "quiet", false, "Quiet mode (minimal output)")

	flag.Parse()

	// Initialize logger
	var logger *zap.Logger
	var err error

	if config.Quiet {
		logger = zap.NewNop()
	} else {
		logger, err = zap.NewProduction()
		if err != nil {
			log.Fatal("Failed to initialize logger:", err)
		}
		defer logger.Sync()
	}

	// Create enhanced load tester
	loadTester := loadtesting.NewEnhancedLoadTester(logger, config.BaseURL)

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("Received interrupt signal, cancelling load test...")
		cancel()
	}()

	// Run the load test
	result, err := runLoadTest(ctx, loadTester, config, logger)
	if err != nil {
		logger.Error("Load test failed", zap.Error(err))
		os.Exit(1)
	}

	// Save results if output file specified
	if config.OutputFile != "" {
		if err := saveResults(result, config.OutputFile); err != nil {
			logger.Error("Failed to save results", zap.Error(err))
		} else {
			logger.Info("Results saved", zap.String("file", config.OutputFile))
		}
	}

	// Print summary
	printSummary(result, logger)
}

// runLoadTest executes the load test with the given configuration
func runLoadTest(ctx context.Context, loadTester *loadtesting.EnhancedLoadTester, config LoadTestConfig, logger *zap.Logger) (*LoadTestResult, error) {
	logger.Info("Starting 10K concurrent users load test",
		zap.String("url", config.BaseURL),
		zap.Int("users", config.Users),
		zap.Duration("duration", config.Duration),
		zap.Duration("ramp_up", config.RampUpTime),
		zap.Float64("target_rps", config.TargetRPS))

	// Create enhanced load test configuration
	loadTestConfig := loadtesting.EnhancedLoadTestConfig{
		Duration:           config.Duration,
		ConcurrentUsers:    config.Users,
		TargetRPS:          config.TargetRPS,
		TargetRPM:          10000, // 10K requests per minute
		RampUpTime:         config.RampUpTime,
		MaxLatency:         1 * time.Second,
		MaxErrorRate:       0.001, // 0.1%
		ConnectionPoolSize: 100,
		RequestTimeout:     config.Timeout,
		KeepAliveTimeout:   90 * time.Second,
		TestPattern:        "ramp",
		SpikeMultiplier:    1.5,
	}

	// Run the enhanced load test
	startTime := time.Now()
	metrics, err := loadTester.RunHighPerformanceLoadTest(ctx, loadTestConfig)
	if err != nil {
		return nil, fmt.Errorf("load test execution failed: %w", err)
	}
	endTime := time.Now()

	// Calculate performance results
	targets := PerformanceTargets{
		P95Latency:    1 * time.Second,
		P99Latency:    2 * time.Second,
		ErrorRate:     0.001, // 0.1%
		ThroughputRPM: 10000,
	}

	results := PerformanceResults{
		P95LatencyMet: metrics.P95ResponseTime <= targets.P95Latency,
		P99LatencyMet: metrics.P99ResponseTime <= targets.P99Latency,
		ErrorRateMet:  metrics.ErrorRate <= targets.ErrorRate,
		ThroughputMet: metrics.RequestsPerMinute >= targets.ThroughputRPM,
	}

	results.OverallSuccess = results.P95LatencyMet && results.P99LatencyMet && results.ErrorRateMet && results.ThroughputMet

	// Create result
	result := &LoadTestResult{
		Config:              config,
		StartTime:           startTime,
		EndTime:             endTime,
		Duration:            endTime.Sub(startTime),
		TotalRequests:       metrics.TotalRequests,
		SuccessfulRequests:  metrics.SuccessfulRequests,
		FailedRequests:      metrics.FailedRequests,
		ErrorRate:           metrics.ErrorRate,
		AverageResponseTime: metrics.P50ResponseTime, // Use P50 as average
		MinResponseTime:     metrics.MinResponseTime,
		MaxResponseTime:     metrics.MaxResponseTime,
		P50ResponseTime:     metrics.P50ResponseTime,
		P95ResponseTime:     metrics.P95ResponseTime,
		P99ResponseTime:     metrics.P99ResponseTime,
		RequestsPerSecond:   metrics.RequestsPerSecond,
		RequestsPerMinute:   metrics.RequestsPerMinute,
		PeakRPS:             metrics.PeakRPS,
		Targets:             targets,
		Results:             results,
	}

	return result, nil
}

// generateRiskAssessmentRequest generates a sample risk assessment request
func generateRiskAssessmentRequest() map[string]interface{} {
	return map[string]interface{}{
		"business_name":             "Test Company Inc",
		"business_address":          "123 Test Street, Test City, TC 12345",
		"industry":                  "Technology",
		"country":                   "US",
		"phone":                     "+1-555-123-4567",
		"email":                     "test@testcompany.com",
		"website":                   "https://testcompany.com",
		"prediction_horizon":        3,
		"model_type":                "auto",
		"include_temporal_analysis": true,
	}
}

// generateBatchJobRequest generates a sample batch job request
func generateBatchJobRequest() map[string]interface{} {
	requests := make([]map[string]interface{}, 10)
	for i := 0; i < 10; i++ {
		requests[i] = generateRiskAssessmentRequest()
	}

	return map[string]interface{}{
		"requests": requests,
		"job_type": "risk_assessment",
		"priority": 1,
	}
}

// saveResults saves the load test results to a JSON file
func saveResults(result *LoadTestResult, filename string) error {
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

	return nil
}

// printSummary prints a summary of the load test results
func printSummary(result *LoadTestResult, logger *zap.Logger) {
	logger.Info("Load Test Summary",
		zap.Duration("duration", result.Duration),
		zap.Int64("total_requests", result.TotalRequests),
		zap.Int64("successful_requests", result.SuccessfulRequests),
		zap.Int64("failed_requests", result.FailedRequests),
		zap.Float64("error_rate", result.ErrorRate),
		zap.Duration("avg_response_time", result.AverageResponseTime),
		zap.Duration("p95_response_time", result.P95ResponseTime),
		zap.Duration("p99_response_time", result.P99ResponseTime),
		zap.Float64("requests_per_second", result.RequestsPerSecond),
		zap.Float64("requests_per_minute", result.RequestsPerMinute),
		zap.Float64("peak_rps", result.PeakRPS))

	// Performance targets
	logger.Info("Performance Targets",
		zap.Duration("p95_latency_target", result.Targets.P95Latency),
		zap.Duration("p99_latency_target", result.Targets.P99Latency),
		zap.Float64("error_rate_target", result.Targets.ErrorRate),
		zap.Float64("throughput_target_rpm", result.Targets.ThroughputRPM))

	// Results
	logger.Info("Performance Results",
		zap.Bool("p95_latency_met", result.Results.P95LatencyMet),
		zap.Bool("p99_latency_met", result.Results.P99LatencyMet),
		zap.Bool("error_rate_met", result.Results.ErrorRateMet),
		zap.Bool("throughput_met", result.Results.ThroughputMet),
		zap.Bool("overall_success", result.Results.OverallSuccess))

	if result.Results.OverallSuccess {
		logger.Info("🎉 All performance targets met! Service is ready for 10K concurrent users.")
	} else {
		logger.Warn("⚠️  Some performance targets were not met. Consider optimization.")
	}
}
