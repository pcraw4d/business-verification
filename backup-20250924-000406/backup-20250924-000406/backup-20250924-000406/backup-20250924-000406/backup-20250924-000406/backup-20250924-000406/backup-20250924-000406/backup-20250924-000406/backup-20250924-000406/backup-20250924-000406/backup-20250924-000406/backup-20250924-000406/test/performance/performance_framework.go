package performance

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// PerformanceTestConfig holds configuration for performance tests
type PerformanceTestConfig struct {
	MaxMerchants      int
	ConcurrentUsers   int
	TestDuration      time.Duration
	BulkOperationSize int
	ResponseTimeLimit time.Duration
}

// DefaultPerformanceConfig returns default performance test configuration
func DefaultPerformanceConfig() *PerformanceTestConfig {
	return &PerformanceTestConfig{
		MaxMerchants:      5000,
		ConcurrentUsers:   20,
		TestDuration:      5 * time.Minute,
		BulkOperationSize: 1000,
		ResponseTimeLimit: 2 * time.Second,
	}
}

// PerformanceMetrics holds performance test results
type PerformanceMetrics struct {
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	AverageResponseTime time.Duration
	MaxResponseTime     time.Duration
	MinResponseTime     time.Duration
	RequestsPerSecond   float64
	ErrorRate           float64
}

// PerformanceTestRunner manages performance test execution
type PerformanceTestRunner struct {
	config    *PerformanceTestConfig
	metrics   *PerformanceMetrics
	startTime time.Time
	endTime   time.Time
}

// NewPerformanceTestRunner creates a new performance test runner
func NewPerformanceTestRunner(config *PerformanceTestConfig) *PerformanceTestRunner {
	return &PerformanceTestRunner{
		config:  config,
		metrics: &PerformanceMetrics{},
	}
}

// Start begins performance test execution
func (ptr *PerformanceTestRunner) Start() {
	ptr.startTime = time.Now()
	ptr.metrics = &PerformanceMetrics{
		MinResponseTime: time.Hour, // Initialize with high value
	}
}

// Stop ends performance test execution and calculates metrics
func (ptr *PerformanceTestRunner) Stop() {
	ptr.endTime = time.Now()
	ptr.calculateMetrics()
}

// RecordRequest records a single request's performance
func (ptr *PerformanceTestRunner) RecordRequest(responseTime time.Duration, success bool) {
	ptr.metrics.TotalRequests++

	if success {
		ptr.metrics.SuccessfulRequests++
	} else {
		ptr.metrics.FailedRequests++
	}

	// Update response time metrics
	if responseTime > ptr.metrics.MaxResponseTime {
		ptr.metrics.MaxResponseTime = responseTime
	}
	if responseTime < ptr.metrics.MinResponseTime {
		ptr.metrics.MinResponseTime = responseTime
	}
}

// calculateMetrics computes final performance metrics
func (ptr *PerformanceTestRunner) calculateMetrics() {
	if ptr.metrics.TotalRequests == 0 {
		return
	}

	// Calculate average response time
	totalDuration := ptr.endTime.Sub(ptr.startTime)
	ptr.metrics.AverageResponseTime = time.Duration(
		int64(ptr.metrics.MaxResponseTime+ptr.metrics.MinResponseTime) / 2,
	)

	// Calculate requests per second
	ptr.metrics.RequestsPerSecond = float64(ptr.metrics.TotalRequests) / totalDuration.Seconds()

	// Calculate error rate
	ptr.metrics.ErrorRate = float64(ptr.metrics.FailedRequests) / float64(ptr.metrics.TotalRequests) * 100
}

// GetMetrics returns current performance metrics
func (ptr *PerformanceTestRunner) GetMetrics() *PerformanceMetrics {
	return ptr.metrics
}

// PrintMetrics prints performance test results
func (ptr *PerformanceTestRunner) PrintMetrics(testName string) {
	fmt.Printf("\n=== Performance Test Results: %s ===\n", testName)
	fmt.Printf("Total Requests: %d\n", ptr.metrics.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", ptr.metrics.SuccessfulRequests)
	fmt.Printf("Failed Requests: %d\n", ptr.metrics.FailedRequests)
	fmt.Printf("Error Rate: %.2f%%\n", ptr.metrics.ErrorRate)
	fmt.Printf("Average Response Time: %v\n", ptr.metrics.AverageResponseTime)
	fmt.Printf("Max Response Time: %v\n", ptr.metrics.MaxResponseTime)
	fmt.Printf("Min Response Time: %v\n", ptr.metrics.MinResponseTime)
	fmt.Printf("Requests Per Second: %.2f\n", ptr.metrics.RequestsPerSecond)
	fmt.Printf("Test Duration: %v\n", ptr.endTime.Sub(ptr.startTime))
	fmt.Println("=====================================\n")
}

// ConcurrentUserSimulator simulates concurrent user behavior
type ConcurrentUserSimulator struct {
	userID     int
	config     *PerformanceTestConfig
	runner     *PerformanceTestRunner
	ctx        context.Context
	cancel     context.CancelFunc
	operations chan func()
}

// NewConcurrentUserSimulator creates a new concurrent user simulator
func NewConcurrentUserSimulator(userID int, config *PerformanceTestConfig, runner *PerformanceTestRunner) *ConcurrentUserSimulator {
	ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)

	return &ConcurrentUserSimulator{
		userID:     userID,
		config:     config,
		runner:     runner,
		ctx:        ctx,
		cancel:     cancel,
		operations: make(chan func(), 100),
	}
}

// Start begins user simulation
func (cus *ConcurrentUserSimulator) Start() {
	go cus.runOperations()
}

// Stop stops user simulation
func (cus *ConcurrentUserSimulator) Stop() {
	cus.cancel()
	close(cus.operations)
}

// AddOperation adds an operation to the user's queue
func (cus *ConcurrentUserSimulator) AddOperation(operation func()) {
	select {
	case cus.operations <- operation:
	case <-cus.ctx.Done():
		return
	}
}

// runOperations executes queued operations
func (cus *ConcurrentUserSimulator) runOperations() {
	for {
		select {
		case operation := <-cus.operations:
			if operation != nil {
				start := time.Now()
				operation()
				responseTime := time.Since(start)
				cus.runner.RecordRequest(responseTime, true)
			}
		case <-cus.ctx.Done():
			return
		}
	}
}

// PerformanceTestSuite manages a collection of performance tests
type PerformanceTestSuite struct {
	tests  []PerformanceTest
	config *PerformanceTestConfig
}

// PerformanceTest defines a single performance test
type PerformanceTest struct {
	Name        string
	Description string
	TestFunc    func(t *testing.T, config *PerformanceTestConfig) error
}

// NewPerformanceTestSuite creates a new performance test suite
func NewPerformanceTestSuite(config *PerformanceTestConfig) *PerformanceTestSuite {
	return &PerformanceTestSuite{
		tests:  make([]PerformanceTest, 0),
		config: config,
	}
}

// AddTest adds a performance test to the suite
func (pts *PerformanceTestSuite) AddTest(test PerformanceTest) {
	pts.tests = append(pts.tests, test)
}

// RunAllTests executes all performance tests in the suite
func (pts *PerformanceTestSuite) RunAllTests(t *testing.T) {
	for _, test := range pts.tests {
		t.Run(test.Name, func(t *testing.T) {
			fmt.Printf("Running performance test: %s\n", test.Name)
			fmt.Printf("Description: %s\n", test.Description)

			err := test.TestFunc(t, pts.config)
			if err != nil {
				t.Errorf("Performance test %s failed: %v", test.Name, err)
			}
		})
	}
}

// AssertPerformanceRequirements validates performance test results
func AssertPerformanceRequirements(t *testing.T, metrics *PerformanceMetrics, config *PerformanceTestConfig) {
	// Assert response time requirements
	assert.LessOrEqual(t, metrics.AverageResponseTime, config.ResponseTimeLimit,
		"Average response time should be within limit")

	// Assert error rate requirements (should be less than 5%)
	assert.Less(t, metrics.ErrorRate, 5.0,
		"Error rate should be less than 5%%")

	// Assert minimum throughput (at least 10 requests per second)
	assert.GreaterOrEqual(t, metrics.RequestsPerSecond, 10.0,
		"Should handle at least 10 requests per second")

	// Assert successful requests
	assert.Greater(t, metrics.SuccessfulRequests, int64(0),
		"Should have successful requests")
}

// BenchmarkHelper provides helper functions for performance testing
type BenchmarkHelper struct {
	config *PerformanceTestConfig
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(config *PerformanceTestConfig) *BenchmarkHelper {
	return &BenchmarkHelper{
		config: config,
	}
}

// MeasureOperation measures the performance of a single operation
func (bh *BenchmarkHelper) MeasureOperation(operation func() error) (time.Duration, error) {
	start := time.Now()
	err := operation()
	duration := time.Since(start)
	return duration, err
}

// MeasureConcurrentOperations measures concurrent operation performance
func (bh *BenchmarkHelper) MeasureConcurrentOperations(operation func() error, concurrency int) (*PerformanceMetrics, error) {
	runner := NewPerformanceTestRunner(bh.config)
	runner.Start()

	var wg sync.WaitGroup
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			err := operation()
			responseTime := time.Since(start)

			runner.RecordRequest(responseTime, err == nil)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	runner.Stop()

	// Check for errors
	select {
	case err := <-errors:
		return runner.GetMetrics(), err
	default:
		return runner.GetMetrics(), nil
	}
}

// GenerateTestData creates test data for performance testing
func (bh *BenchmarkHelper) GenerateTestData(merchantCount int) []map[string]interface{} {
	data := make([]map[string]interface{}, merchantCount)

	for i := 0; i < merchantCount; i++ {
		data[i] = map[string]interface{}{
			"id":             fmt.Sprintf("merchant_%d", i),
			"name":           fmt.Sprintf("Test Merchant %d", i),
			"email":          fmt.Sprintf("merchant%d@test.com", i),
			"phone":          fmt.Sprintf("+1-555-%04d", i),
			"address":        fmt.Sprintf("%d Test Street, Test City, TS %05d", i, i),
			"website":        fmt.Sprintf("https://merchant%d.com", i),
			"industry":       "Technology",
			"portfolio_type": "onboarded",
			"risk_level":     "medium",
			"created_at":     time.Now(),
			"updated_at":     time.Now(),
		}
	}

	return data
}

// Simulation helper functions
func simulateListMerchants(testData []map[string]interface{}) {
	time.Sleep(time.Duration(len(testData)/1000) * time.Millisecond)
}

func simulateSearchMerchants(testData []map[string]interface{}) {
	time.Sleep(time.Duration(len(testData)/2000) * time.Millisecond)
}

func simulateFilterMerchants(testData []map[string]interface{}) {
	time.Sleep(time.Duration(len(testData)/1500) * time.Millisecond)
}

func simulatePaginateMerchants(testData []map[string]interface{}) {
	time.Sleep(50 * time.Millisecond)
}

func simulateLoadMerchantDetails(merchant map[string]interface{}) {
	time.Sleep(100 * time.Millisecond)
}

func simulateSearchOperation(testData []map[string]interface{}, query string) {
	time.Sleep(time.Duration(len(testData)/3000) * time.Millisecond)
}

func simulateBulkUpdate(testData []map[string]interface{}, batchSize int) {
	time.Sleep(time.Duration(batchSize/10) * time.Millisecond)
}

func simulateBulkExport(testData []map[string]interface{}, batchSize int) {
	time.Sleep(time.Duration(batchSize/5) * time.Millisecond)
}

func simulateBulkStatusChange(testData []map[string]interface{}, batchSize int) {
	time.Sleep(time.Duration(batchSize/15) * time.Millisecond)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

// containsSubstring performs substring search
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
