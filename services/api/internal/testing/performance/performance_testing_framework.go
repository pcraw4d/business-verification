package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// PerformanceTestConfig holds configuration for performance tests
type PerformanceTestConfig struct {
	BaseURL          string        `json:"base_url"`
	ConcurrentUsers  int           `json:"concurrent_users"`
	TestDuration     time.Duration `json:"test_duration"`
	RampUpDuration   time.Duration `json:"ramp_up_duration"`
	RequestTimeout   time.Duration `json:"request_timeout"`
	MaxMemoryMB      int64         `json:"max_memory_mb"`
	TargetResponseMs int           `json:"target_response_ms"`
	ErrorThreshold   float64       `json:"error_threshold"`
	ThroughputTarget int           `json:"throughput_target"`
}

// PerformanceMetrics tracks performance test results
type PerformanceMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	ThroughputPerSecond float64       `json:"throughput_per_second"`
	ErrorRate           float64       `json:"error_rate"`
	MemoryUsageMB       float64       `json:"memory_usage_mb"`
	CPUUsagePercent     float64       `json:"cpu_usage_percent"`
	TestDuration        time.Duration `json:"test_duration"`
	StartTime           time.Time     `json:"start_time"`
	EndTime             time.Time     `json:"end_time"`
}

// TestScenario represents a specific test scenario
type TestScenario struct {
	Name           string            `json:"name"`
	Method         string            `json:"method"`
	Endpoint       string            `json:"endpoint"`
	Headers        map[string]string `json:"headers"`
	Body           interface{}       `json:"body"`
	Weight         int               `json:"weight"` // Relative frequency of this scenario
	ExpectedStatus int               `json:"expected_status"`
}

// PerformanceTestFramework provides comprehensive performance testing capabilities
type PerformanceTestFramework struct {
	config    PerformanceTestConfig
	scenarios []TestScenario
	metrics   *PerformanceMetrics
	mu        sync.RWMutex
}

// NewPerformanceTestFramework creates a new performance testing framework
func NewPerformanceTestFramework(config PerformanceTestConfig) *PerformanceTestFramework {
	return &PerformanceTestFramework{
		config:    config,
		scenarios: make([]TestScenario, 0),
		metrics:   &PerformanceMetrics{},
	}
}

// AddScenario adds a test scenario to the framework
func (ptf *PerformanceTestFramework) AddScenario(scenario TestScenario) {
	ptf.mu.Lock()
	defer ptf.mu.Unlock()
	ptf.scenarios = append(ptf.scenarios, scenario)
}

// AddScenarios adds multiple test scenarios
func (ptf *PerformanceTestFramework) AddScenarios(scenarios []TestScenario) {
	ptf.mu.Lock()
	defer ptf.mu.Unlock()
	ptf.scenarios = append(ptf.scenarios, scenarios...)
}

// RunLoadTest executes a load test with the configured scenarios
func (ptf *PerformanceTestFramework) RunLoadTest(ctx context.Context) (*PerformanceMetrics, error) {
	log.Printf("Starting load test with %d concurrent users for %v",
		ptf.config.ConcurrentUsers, ptf.config.TestDuration)

	ptf.metrics.StartTime = time.Now()
	defer func() {
		ptf.metrics.EndTime = time.Now()
		ptf.metrics.TestDuration = ptf.metrics.EndTime.Sub(ptf.metrics.StartTime)
	}()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: ptf.config.RequestTimeout,
	}

	// Channel to collect results
	resultChan := make(chan *RequestResult, ptf.config.ConcurrentUsers*100)

	// Start monitoring goroutine
	monitorCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go ptf.monitorSystemResources(monitorCtx)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < ptf.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ptf.runWorker(ctx, client, resultChan, workerID)
		}(i)
	}

	// Wait for test duration
	time.Sleep(ptf.config.TestDuration)

	// Signal workers to stop
	cancel()

	// Wait for all workers to complete
	wg.Wait()
	close(resultChan)

	// Process results
	return ptf.processResults(resultChan)
}

// RequestResult represents the result of a single request
type RequestResult struct {
	Scenario     string
	Method       string
	Endpoint     string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
	Timestamp    time.Time
}

// runWorker runs a single worker that executes requests
func (ptf *PerformanceTestFramework) runWorker(ctx context.Context, client *http.Client, resultChan chan<- *RequestResult, workerID int) {
	ticker := time.NewTicker(100 * time.Millisecond) // 10 requests per second per worker
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			scenario := ptf.selectRandomScenario()
			if scenario == nil {
				continue
			}

			result := ptf.executeRequest(client, scenario)
			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// selectRandomScenario selects a random scenario based on weights
func (ptf *PerformanceTestFramework) selectRandomScenario() *TestScenario {
	ptf.mu.RLock()
	defer ptf.mu.RUnlock()

	if len(ptf.scenarios) == 0 {
		return nil
	}

	// Calculate total weight
	totalWeight := 0
	for _, scenario := range ptf.scenarios {
		totalWeight += scenario.Weight
	}

	if totalWeight == 0 {
		return &ptf.scenarios[rand.Intn(len(ptf.scenarios))]
	}

	// Select based on weight
	random := rand.Intn(totalWeight)
	currentWeight := 0
	for _, scenario := range ptf.scenarios {
		currentWeight += scenario.Weight
		if random < currentWeight {
			return &scenario
		}
	}

	return &ptf.scenarios[len(ptf.scenarios)-1]
}

// executeRequest executes a single HTTP request
func (ptf *PerformanceTestFramework) executeRequest(client *http.Client, scenario *TestScenario) *RequestResult {
	start := time.Now()

	// Build request
	url := ptf.config.BaseURL + scenario.Endpoint
	var req *http.Request
	var err error

	if scenario.Body != nil {
		jsonBody, _ := json.Marshal(scenario.Body)
		req, err = http.NewRequest(scenario.Method, url, strings.NewReader(string(jsonBody)))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	} else {
		req, err = http.NewRequest(scenario.Method, url, nil)
	}

	if err != nil {
		return &RequestResult{
			Scenario:     scenario.Name,
			Method:       scenario.Method,
			Endpoint:     scenario.Endpoint,
			StatusCode:   0,
			ResponseTime: time.Since(start),
			Error:        err,
			Timestamp:    time.Now(),
		}
	}

	// Add headers
	for key, value := range scenario.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := client.Do(req)
	responseTime := time.Since(start)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
		resp.Body.Close()
	}

	result := &RequestResult{
		Scenario:     scenario.Name,
		Method:       scenario.Method,
		Endpoint:     scenario.Endpoint,
		StatusCode:   statusCode,
		ResponseTime: responseTime,
		Error:        err,
		Timestamp:    time.Now(),
	}

	// Note: Prometheus metrics removed for simplicity

	return result
}

// monitorSystemResources monitors system resource usage
func (ptf *PerformanceTestFramework) monitorSystemResources(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Monitor memory usage
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			memoryMB := float64(m.Alloc) / 1024 / 1024

			// Update metrics
			ptf.mu.Lock()
			ptf.metrics.MemoryUsageMB = memoryMB
			ptf.mu.Unlock()
		}
	}
}

// processResults processes all request results and calculates metrics
func (ptf *PerformanceTestFramework) processResults(resultChan <-chan *RequestResult) (*PerformanceMetrics, error) {
	var responseTimes []time.Duration
	var totalRequests, successfulRequests, failedRequests int64

	for result := range resultChan {
		totalRequests++
		responseTimes = append(responseTimes, result.ResponseTime)

		if result.Error != nil || result.StatusCode >= 400 {
			failedRequests++
		} else {
			successfulRequests++
		}
	}

	// Calculate response time statistics
	if len(responseTimes) > 0 {
		ptf.calculateResponseTimeStats(responseTimes)
	}

	// Update metrics
	ptf.mu.Lock()
	ptf.metrics.TotalRequests = totalRequests
	ptf.metrics.SuccessfulRequests = successfulRequests
	ptf.metrics.FailedRequests = failedRequests
	ptf.metrics.ErrorRate = float64(failedRequests) / float64(totalRequests) * 100
	ptf.metrics.ThroughputPerSecond = float64(totalRequests) / ptf.metrics.TestDuration.Seconds()
	ptf.mu.Unlock()

	// Note: Prometheus metrics removed for simplicity

	return ptf.metrics, nil
}

// calculateResponseTimeStats calculates response time statistics
func (ptf *PerformanceTestFramework) calculateResponseTimeStats(responseTimes []time.Duration) {
	ptf.mu.Lock()
	defer ptf.mu.Unlock()

	// Sort response times for percentile calculations
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	// Calculate statistics
	var total time.Duration
	for _, rt := range responseTimes {
		total += rt
	}

	ptf.metrics.AverageResponseTime = total / time.Duration(len(responseTimes))
	ptf.metrics.MinResponseTime = responseTimes[0]
	ptf.metrics.MaxResponseTime = responseTimes[len(responseTimes)-1]

	// Calculate percentiles
	ptf.metrics.P95ResponseTime = ptf.calculatePercentile(responseTimes, 95)
	ptf.metrics.P99ResponseTime = ptf.calculatePercentile(responseTimes, 99)
}

// calculatePercentile calculates the nth percentile of response times
func (ptf *PerformanceTestFramework) calculatePercentile(responseTimes []time.Duration, percentile int) time.Duration {
	if len(responseTimes) == 0 {
		return 0
	}

	index := int(float64(len(responseTimes)) * float64(percentile) / 100.0)
	if index >= len(responseTimes) {
		index = len(responseTimes) - 1
	}

	return responseTimes[index]
}

// GetMetrics returns current performance metrics
func (ptf *PerformanceTestFramework) GetMetrics() *PerformanceMetrics {
	ptf.mu.RLock()
	defer ptf.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := *ptf.metrics
	return &metrics
}

// ValidatePerformance checks if performance meets targets
func (ptf *PerformanceTestFramework) ValidatePerformance() []string {
	var issues []string

	ptf.mu.RLock()
	defer ptf.mu.RUnlock()

	// Check response time targets
	if ptf.metrics.AverageResponseTime > time.Duration(ptf.config.TargetResponseMs)*time.Millisecond {
		issues = append(issues, fmt.Sprintf("Average response time %v exceeds target %dms",
			ptf.metrics.AverageResponseTime, ptf.config.TargetResponseMs))
	}

	// Check error rate
	if ptf.metrics.ErrorRate > ptf.config.ErrorThreshold {
		issues = append(issues, fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%",
			ptf.metrics.ErrorRate, ptf.config.ErrorThreshold))
	}

	// Check throughput
	if ptf.metrics.ThroughputPerSecond < float64(ptf.config.ThroughputTarget) {
		issues = append(issues, fmt.Sprintf("Throughput %.2f req/s below target %d req/s",
			ptf.metrics.ThroughputPerSecond, ptf.config.ThroughputTarget))
	}

	// Check memory usage
	if ptf.metrics.MemoryUsageMB > float64(ptf.config.MaxMemoryMB) {
		issues = append(issues, fmt.Sprintf("Memory usage %.2fMB exceeds limit %dMB",
			ptf.metrics.MemoryUsageMB, ptf.config.MaxMemoryMB))
	}

	return issues
}

// GenerateReport generates a comprehensive performance test report
func (ptf *PerformanceTestFramework) GenerateReport() *PerformanceTestReport {
	ptf.mu.RLock()
	defer ptf.mu.RUnlock()

	report := &PerformanceTestReport{
		TestConfig:       ptf.config,
		Metrics:          *ptf.metrics,
		ValidationIssues: ptf.ValidatePerformance(),
		GeneratedAt:      time.Now(),
	}

	// Add recommendations
	report.Recommendations = ptf.generateRecommendations()

	return report
}

// PerformanceTestReport represents a comprehensive performance test report
type PerformanceTestReport struct {
	TestConfig       PerformanceTestConfig `json:"test_config"`
	Metrics          PerformanceMetrics    `json:"metrics"`
	ValidationIssues []string              `json:"validation_issues"`
	Recommendations  []string              `json:"recommendations"`
	GeneratedAt      time.Time             `json:"generated_at"`
}

// generateRecommendations generates performance improvement recommendations
func (ptf *PerformanceTestFramework) generateRecommendations() []string {
	var recommendations []string

	ptf.mu.RLock()
	defer ptf.mu.RUnlock()

	// Response time recommendations
	if ptf.metrics.AverageResponseTime > 200*time.Millisecond {
		recommendations = append(recommendations,
			"Consider implementing caching for frequently accessed data")
		recommendations = append(recommendations,
			"Review database query performance and add indexes if needed")
		recommendations = append(recommendations,
			"Consider implementing connection pooling for database connections")
	}

	// Memory recommendations
	if ptf.metrics.MemoryUsageMB > 500 {
		recommendations = append(recommendations,
			"Review memory usage patterns and implement memory optimization")
		recommendations = append(recommendations,
			"Consider implementing garbage collection tuning")
	}

	// Error rate recommendations
	if ptf.metrics.ErrorRate > 1.0 {
		recommendations = append(recommendations,
			"Investigate error patterns and implement better error handling")
		recommendations = append(recommendations,
			"Consider implementing circuit breakers for external dependencies")
	}

	// Throughput recommendations
	if ptf.metrics.ThroughputPerSecond < 100 {
		recommendations = append(recommendations,
			"Consider horizontal scaling to increase throughput")
		recommendations = append(recommendations,
			"Review application architecture for bottlenecks")
	}

	return recommendations
}
