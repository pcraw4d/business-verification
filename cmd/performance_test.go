package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PerformanceTestConfig contains performance test configuration
type PerformanceTestConfig struct {
	BaseURL          string         `json:"base_url"`
	ConcurrentUsers  int            `json:"concurrent_users"`
	RequestsPerUser  int            `json:"requests_per_user"`
	TestDuration     time.Duration  `json:"test_duration"`
	RampUpTime       time.Duration  `json:"ramp_up_time"`
	TargetP95        time.Duration  `json:"target_p95"`
	TargetP99        time.Duration  `json:"target_p99"`
	TargetThroughput int            `json:"target_throughput"`
	EnableWarmup     bool           `json:"enable_warmup"`
	WarmupRequests   int            `json:"warmup_requests"`
	TestEndpoints    []TestEndpoint `json:"test_endpoints"`
}

// TestEndpoint represents an endpoint to test
type TestEndpoint struct {
	Path           string            `json:"path"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Body           interface{}       `json:"body"`
	Weight         int               `json:"weight"`
	ExpectedStatus int               `json:"expected_status"`
}

// PerformanceTestResult contains test results
type PerformanceTestResult struct {
	TotalRequests      int64           `json:"total_requests"`
	SuccessfulRequests int64           `json:"successful_requests"`
	FailedRequests     int64           `json:"failed_requests"`
	AverageTime        time.Duration   `json:"average_time"`
	P50Time            time.Duration   `json:"p50_time"`
	P95Time            time.Duration   `json:"p95_time"`
	P99Time            time.Duration   `json:"p99_time"`
	MaxTime            time.Duration   `json:"max_time"`
	MinTime            time.Duration   `json:"min_time"`
	Throughput         float64         `json:"throughput"`
	ErrorRate          float64         `json:"error_rate"`
	TestDuration       time.Duration   `json:"test_duration"`
	StartTime          time.Time       `json:"start_time"`
	EndTime            time.Time       `json:"end_time"`
	ResponseTimes      []time.Duration `json:"-"`
	Errors             []TestError     `json:"errors"`
}

// TestError represents a test error
type TestError struct {
	Endpoint   string        `json:"endpoint"`
	Method     string        `json:"method"`
	StatusCode int           `json:"status_code"`
	Error      string        `json:"error"`
	Duration   time.Duration `json:"duration"`
	Timestamp  time.Time     `json:"timestamp"`
}

// PerformanceTester performs performance testing
type PerformanceTester struct {
	config  *PerformanceTestConfig
	logger  *zap.Logger
	client  *http.Client
	results *PerformanceTestResult
	mu      sync.RWMutex
}

// NewPerformanceTester creates a new performance tester
func NewPerformanceTester(config *PerformanceTestConfig, logger *zap.Logger) *PerformanceTester {
	return &PerformanceTester{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		results: &PerformanceTestResult{
			ResponseTimes: make([]time.Duration, 0),
			Errors:        make([]TestError, 0),
		},
	}
}

// RunTest runs the performance test
func (pt *PerformanceTester) RunTest() (*PerformanceTestResult, error) {
	pt.logger.Info("Starting performance test",
		zap.String("base_url", pt.config.BaseURL),
		zap.Int("concurrent_users", pt.config.ConcurrentUsers),
		zap.Int("requests_per_user", pt.config.RequestsPerUser),
		zap.Duration("test_duration", pt.config.TestDuration))

	pt.results.StartTime = time.Now()

	// Warmup phase
	if pt.config.EnableWarmup {
		pt.logger.Info("Running warmup phase")
		if err := pt.runWarmup(); err != nil {
			pt.logger.Warn("Warmup failed", zap.Error(err))
		}
	}

	// Main test phase
	pt.logger.Info("Running main test phase")
	if err := pt.runMainTest(); err != nil {
		return nil, fmt.Errorf("main test failed: %w", err)
	}

	pt.results.EndTime = time.Now()
	pt.results.TestDuration = pt.results.EndTime.Sub(pt.results.StartTime)

	// Calculate final statistics
	pt.calculateStatistics()

	pt.logger.Info("Performance test completed",
		zap.Int64("total_requests", pt.results.TotalRequests),
		zap.Int64("successful_requests", pt.results.SuccessfulRequests),
		zap.Int64("failed_requests", pt.results.FailedRequests),
		zap.Duration("average_time", pt.results.AverageTime),
		zap.Duration("p95_time", pt.results.P95Time),
		zap.Duration("p99_time", pt.results.P99Time),
		zap.Float64("throughput", pt.results.Throughput),
		zap.Float64("error_rate", pt.results.ErrorRate))

	return pt.results, nil
}

// runWarmup runs the warmup phase
func (pt *PerformanceTester) runWarmup() error {
	for i := 0; i < pt.config.WarmupRequests; i++ {
		endpoint := pt.selectEndpoint()
		if err := pt.executeRequest(endpoint); err != nil {
			pt.logger.Debug("Warmup request failed", zap.Error(err))
		}
		time.Sleep(100 * time.Millisecond) // Small delay between warmup requests
	}
	return nil
}

// runMainTest runs the main test phase
func (pt *PerformanceTester) runMainTest() error {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, pt.config.ConcurrentUsers)

	// Calculate total requests
	totalRequests := pt.config.ConcurrentUsers * pt.config.RequestsPerUser
	pt.results.ResponseTimes = make([]time.Duration, 0, totalRequests)

	// Start test
	for i := 0; i < pt.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// Ramp up delay
			if pt.config.RampUpTime > 0 {
				delay := time.Duration(userID) * pt.config.RampUpTime / time.Duration(pt.config.ConcurrentUsers)
				time.Sleep(delay)
			}

			for j := 0; j < pt.config.RequestsPerUser; j++ {
				semaphore <- struct{}{} // Acquire semaphore

				endpoint := pt.selectEndpoint()
				pt.executeRequest(endpoint)

				<-semaphore // Release semaphore
			}
		}(i)
	}

	wg.Wait()
	return nil
}

// selectEndpoint selects an endpoint based on weights
func (pt *PerformanceTester) selectEndpoint() TestEndpoint {
	if len(pt.config.TestEndpoints) == 0 {
		// Default endpoint
		return TestEndpoint{
			Path:   "/health",
			Method: "GET",
			Weight: 1,
		}
	}

	// Simple weighted selection
	totalWeight := 0
	for _, endpoint := range pt.config.TestEndpoints {
		totalWeight += endpoint.Weight
	}

	// For simplicity, just return the first endpoint
	// In a real implementation, you'd implement proper weighted selection
	return pt.config.TestEndpoints[0]
}

// executeRequest executes a single request
func (pt *PerformanceTester) executeRequest(endpoint TestEndpoint) error {
	start := time.Now()

	// Build URL
	url := pt.config.BaseURL + endpoint.Path

	// Prepare request body
	var body io.Reader
	if endpoint.Body != nil {
		jsonBody, err := json.Marshal(endpoint.Body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequest(endpoint.Method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Performance-Tester/1.0")
	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := pt.client.Do(req)
	duration := time.Since(start)

	// Record response time
	pt.mu.Lock()
	pt.results.ResponseTimes = append(pt.results.ResponseTimes, duration)
	pt.results.TotalRequests++
	pt.mu.Unlock()

	// Check for errors
	if err != nil {
		pt.mu.Lock()
		pt.results.FailedRequests++
		pt.results.Errors = append(pt.results.Errors, TestError{
			Endpoint:   endpoint.Path,
			Method:     endpoint.Method,
			StatusCode: 0,
			Error:      err.Error(),
			Duration:   duration,
			Timestamp:  time.Now(),
		})
		pt.mu.Unlock()
		return err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		pt.mu.Lock()
		pt.results.FailedRequests++
		pt.results.Errors = append(pt.results.Errors, TestError{
			Endpoint:   endpoint.Path,
			Method:     endpoint.Method,
			StatusCode: resp.StatusCode,
			Error:      fmt.Sprintf("HTTP %d", resp.StatusCode),
			Duration:   duration,
			Timestamp:  time.Now(),
		})
		pt.mu.Unlock()
	} else {
		pt.mu.Lock()
		pt.results.SuccessfulRequests++
		pt.mu.Unlock()
	}

	return nil
}

// calculateStatistics calculates final statistics
func (pt *PerformanceTester) calculateStatistics() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if len(pt.results.ResponseTimes) == 0 {
		return
	}

	// Sort response times for percentile calculations
	times := pt.results.ResponseTimes
	for i := 0; i < len(times)-1; i++ {
		for j := 0; j < len(times)-i-1; j++ {
			if times[j] > times[j+1] {
				times[j], times[j+1] = times[j+1], times[j]
			}
		}
	}

	// Calculate statistics
	total := time.Duration(0)
	for _, t := range times {
		total += t
	}

	pt.results.AverageTime = total / time.Duration(len(times))
	pt.results.MinTime = times[0]
	pt.results.MaxTime = times[len(times)-1]
	pt.results.P50Time = pt.calculatePercentile(times, 0.50)
	pt.results.P95Time = pt.calculatePercentile(times, 0.95)
	pt.results.P99Time = pt.calculatePercentile(times, 0.99)

	// Calculate throughput (requests per second)
	if pt.results.TestDuration > 0 {
		pt.results.Throughput = float64(pt.results.TotalRequests) / pt.results.TestDuration.Seconds()
	}

	// Calculate error rate
	if pt.results.TotalRequests > 0 {
		pt.results.ErrorRate = float64(pt.results.FailedRequests) / float64(pt.results.TotalRequests) * 100
	}
}

// calculatePercentile calculates the nth percentile
func (pt *PerformanceTester) calculatePercentile(times []time.Duration, percentile float64) time.Duration {
	if len(times) == 0 {
		return 0
	}

	index := int(float64(len(times)-1) * percentile)
	if index >= len(times) {
		index = len(times) - 1
	}

	return times[index]
}

// ValidateTargets validates if performance targets are met
func (pt *PerformanceTester) ValidateTargets() map[string]bool {
	results := make(map[string]bool)

	results["p95_target"] = pt.results.P95Time <= pt.config.TargetP95
	results["p99_target"] = pt.results.P99Time <= pt.config.TargetP99
	results["throughput_target"] = pt.results.Throughput >= float64(pt.config.TargetThroughput)
	results["error_rate_target"] = pt.results.ErrorRate < 1.0 // Less than 1% error rate

	return results
}

// GenerateReport generates a performance test report
func (pt *PerformanceTester) GenerateReport() string {
	report := fmt.Sprintf("=== PERFORMANCE TEST REPORT ===\n")
	report += fmt.Sprintf("Test Duration: %v\n", pt.results.TestDuration)
	report += fmt.Sprintf("Start Time: %s\n", pt.results.StartTime.Format(time.RFC3339))
	report += fmt.Sprintf("End Time: %s\n", pt.results.EndTime.Format(time.RFC3339))
	report += fmt.Sprintf("\n=== REQUEST STATISTICS ===\n")
	report += fmt.Sprintf("Total Requests: %d\n", pt.results.TotalRequests)
	report += fmt.Sprintf("Successful Requests: %d\n", pt.results.SuccessfulRequests)
	report += fmt.Sprintf("Failed Requests: %d\n", pt.results.FailedRequests)
	report += fmt.Sprintf("Error Rate: %.2f%%\n", pt.results.ErrorRate)
	report += fmt.Sprintf("Throughput: %.2f req/sec\n", pt.results.Throughput)

	report += fmt.Sprintf("\n=== RESPONSE TIME STATISTICS ===\n")
	report += fmt.Sprintf("Average Time: %v\n", pt.results.AverageTime)
	report += fmt.Sprintf("P50 Time: %v\n", pt.results.P50Time)
	report += fmt.Sprintf("P95 Time: %v\n", pt.results.P95Time)
	report += fmt.Sprintf("P99 Time: %v\n", pt.results.P99Time)
	report += fmt.Sprintf("Min Time: %v\n", pt.results.MinTime)
	report += fmt.Sprintf("Max Time: %v\n", pt.results.MaxTime)

	report += fmt.Sprintf("\n=== TARGET VALIDATION ===\n")
	targets := pt.ValidateTargets()
	report += fmt.Sprintf("P95 Target (%v): %s\n", pt.config.TargetP95,
		func() string {
			if targets["p95_target"] {
				return "✅ PASS"
			}
			return "❌ FAIL"
		}())
	report += fmt.Sprintf("P99 Target (%v): %s\n", pt.config.TargetP99,
		func() string {
			if targets["p99_target"] {
				return "✅ PASS"
			}
			return "❌ FAIL"
		}())
	report += fmt.Sprintf("Throughput Target (%d req/sec): %s\n", pt.config.TargetThroughput,
		func() string {
			if targets["throughput_target"] {
				return "✅ PASS"
			}
			return "❌ FAIL"
		}())
	report += fmt.Sprintf("Error Rate Target (<1%%): %s\n",
		func() string {
			if targets["error_rate_target"] {
				return "✅ PASS"
			}
			return "❌ FAIL"
		}())

	report += fmt.Sprintf("\n=== ERRORS ===\n")
	if len(pt.results.Errors) == 0 {
		report += "No errors\n"
	} else {
		report += fmt.Sprintf("Total Errors: %d\n", len(pt.results.Errors))
		for i, err := range pt.results.Errors {
			if i >= 10 { // Limit to first 10 errors
				report += fmt.Sprintf("... and %d more errors\n", len(pt.results.Errors)-10)
				break
			}
			report += fmt.Sprintf("%d. %s %s - %s (Status: %d, Time: %v)\n",
				i+1, err.Method, err.Endpoint, err.Error, err.StatusCode, err.Duration)
		}
	}

	return report
}

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Default test configuration
	config := &PerformanceTestConfig{
		BaseURL:          "http://localhost:8080",
		ConcurrentUsers:  10,
		RequestsPerUser:  100,
		TestDuration:     5 * time.Minute,
		RampUpTime:       30 * time.Second,
		TargetP95:        1 * time.Second,
		TargetP99:        2 * time.Second,
		TargetThroughput: 100,
		EnableWarmup:     true,
		WarmupRequests:   50,
		TestEndpoints: []TestEndpoint{
			{
				Path:   "/health",
				Method: "GET",
				Weight: 1,
			},
			{
				Path:   "/risk-assessment",
				Method: "POST",
				Body: map[string]interface{}{
					"business_name": "Test Company",
					"address":       "123 Test St, Test City, TC 12345",
					"industry":      "Technology",
				},
				Weight: 3,
			},
		},
	}

	// Create and run performance test
	tester := NewPerformanceTester(config, logger)
	result, err := tester.RunTest()
	if err != nil {
		logger.Fatal("Performance test failed", zap.Error(err))
	}

	// Generate and print report
	report := tester.GenerateReport()
	fmt.Println(report)

	// Validate targets
	targets := tester.ValidateTargets()
	allPassed := true
	for name, passed := range targets {
		if !passed {
			allPassed = false
			logger.Warn("Target not met", zap.String("target", name), zap.Bool("passed", passed))
		}
	}

	if allPassed {
		logger.Info("All performance targets met! 🎉")
	} else {
		logger.Warn("Some performance targets not met")
	}
}
