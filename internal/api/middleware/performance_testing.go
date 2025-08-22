package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// PerformanceTestConfig configures automated performance testing
type PerformanceTestConfig struct {
	// Test Configuration
	TestEnabled     bool          `json:"test_enabled"`
	TestDirectory   string        `json:"test_directory"`
	TestTimeout     time.Duration `json:"test_timeout"`
	TestRetries     int           `json:"test_retries"`
	TestParallelism int           `json:"test_parallelism"`

	// Load Testing Configuration
	LoadTestEnabled  bool          `json:"load_test_enabled"`
	LoadTestDuration time.Duration `json:"load_test_duration"`
	LoadTestRPS      int           `json:"load_test_rps"`       // requests per second
	LoadTestUsers    int           `json:"load_test_users"`     // concurrent users
	LoadTestRampUp   time.Duration `json:"load_test_ramp_up"`   // ramp-up period
	LoadTestRampDown time.Duration `json:"load_test_ramp_down"` // ramp-down period

	// Stress Testing Configuration
	StressTestEnabled  bool          `json:"stress_test_enabled"`
	StressTestDuration time.Duration `json:"stress_test_duration"`
	StressTestMaxRPS   int           `json:"stress_test_max_rps"` // maximum RPS to test
	StressTestStep     int           `json:"stress_test_step"`    // RPS increment step
	StressTestTimeout  time.Duration `json:"stress_test_timeout"` // timeout per step

	// Benchmark Configuration
	BenchmarkEnabled    bool          `json:"benchmark_enabled"`
	BenchmarkIterations int           `json:"benchmark_iterations"`
	BenchmarkWarmup     time.Duration `json:"benchmark_warmup"`
	BenchmarkCooldown   time.Duration `json:"benchmark_cooldown"`

	// Failure Injection Configuration
	FailureInjectionEnabled bool          `json:"failure_injection_enabled"`
	FailureRate             float64       `json:"failure_rate"`     // percentage of failures to inject
	FailureTypes            []string      `json:"failure_types"`    // types of failures to inject
	FailureDuration         time.Duration `json:"failure_duration"` // duration of failure injection

	// Reporting Configuration
	ReportEnabled    bool               `json:"report_enabled"`
	ReportFormat     string             `json:"report_format"` // "json", "html", "csv"
	ReportDirectory  string             `json:"report_directory"`
	ReportRetention  int                `json:"report_retention"`  // number of reports to keep
	ReportThresholds map[string]float64 `json:"report_thresholds"` // performance thresholds

	// Monitoring Configuration
	MonitorEnabled  bool          `json:"monitor_enabled"`
	MonitorInterval time.Duration `json:"monitor_interval"`
	MonitorMetrics  []string      `json:"monitor_metrics"`
	MonitorAlerts   bool          `json:"monitor_alerts"`
}

// PerformanceTestManager manages automated performance testing
type PerformanceTestManager struct {
	config          *PerformanceTestConfig
	logger          *zap.Logger
	loadTester      *PerformanceLoadTester
	stressTester    *StressTester
	benchmarker     *Benchmarker
	failureInjector *FailureInjector
	reporter        *TestReporter
	monitor         *TestMonitor
	results         *TestResults
	mu              sync.RWMutex
	stopChan        chan struct{}
}

// PerformanceLoadTester performs load testing for performance testing
type PerformanceLoadTester struct {
	config *PerformanceTestConfig
	logger *zap.Logger
	client *http.Client
	stats  *LoadTestStats
	mu     sync.RWMutex
}

// StressTester performs stress testing
type StressTester struct {
	config *PerformanceTestConfig
	logger *zap.Logger
	client *http.Client
	stats  *StressTestStats
	mu     sync.RWMutex
}

// Benchmarker performs performance benchmarking
type Benchmarker struct {
	config *PerformanceTestConfig
	logger *zap.Logger
	stats  *BenchmarkStats
	mu     sync.RWMutex
}

// FailureInjector injects failures for testing resilience
type FailureInjector struct {
	config *PerformanceTestConfig
	logger *zap.Logger
	active bool
	mu     sync.RWMutex
}

// TestReporter generates test reports
type TestReporter struct {
	config *PerformanceTestConfig
	logger *zap.Logger
	mu     sync.RWMutex
}

// TestMonitor monitors test execution
type TestMonitor struct {
	config *PerformanceTestConfig
	logger *zap.Logger
	stats  *TestMonitorStats
	mu     sync.RWMutex
}

// LoadTestStats tracks load test statistics
type LoadTestStats struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	P50ResponseTime     time.Duration `json:"p50_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	RequestsPerSecond   float64       `json:"requests_per_second"`
	ErrorRate           float64       `json:"error_rate"`
	StartTime           time.Time     `json:"start_time"`
	EndTime             time.Time     `json:"end_time"`
	Duration            time.Duration `json:"duration"`
}

// StressTestStats tracks stress test statistics
type StressTestStats struct {
	MaxRPS              int                `json:"max_rps"`
	BreakingPoint       int                `json:"breaking_point"`
	ResponseTimeAtMax   time.Duration      `json:"response_time_at_max"`
	ErrorRateAtMax      float64            `json:"error_rate_at_max"`
	ResourceUtilization map[string]float64 `json:"resource_utilization"`
	StartTime           time.Time          `json:"start_time"`
	EndTime             time.Time          `json:"end_time"`
	Duration            time.Duration      `json:"duration"`
}

// BenchmarkStats tracks benchmark statistics
type BenchmarkStats struct {
	OperationName       string        `json:"operation_name"`
	Iterations          int           `json:"iterations"`
	TotalTime           time.Duration `json:"total_time"`
	AverageTime         time.Duration `json:"average_time"`
	MinTime             time.Duration `json:"min_time"`
	MaxTime             time.Duration `json:"max_time"`
	P50Time             time.Duration `json:"p50_time"`
	P95Time             time.Duration `json:"p95_time"`
	P99Time             time.Duration `json:"p99_time"`
	OperationsPerSecond float64       `json:"operations_per_second"`
	MemoryUsage         uint64        `json:"memory_usage"`
	CPUUsage            float64       `json:"cpu_usage"`
	StartTime           time.Time     `json:"start_time"`
	EndTime             time.Time     `json:"end_time"`
}

// TestResults aggregates all test results
type TestResults struct {
	LoadTestResults         *LoadTestStats         `json:"load_test_results"`
	StressTestResults       *StressTestStats       `json:"stress_test_results"`
	BenchmarkResults        []*BenchmarkStats      `json:"benchmark_results"`
	FailureInjectionResults map[string]interface{} `json:"failure_injection_results"`
	OverallScore            float64                `json:"overall_score"`
	Recommendations         []string               `json:"recommendations"`
	GeneratedAt             time.Time              `json:"generated_at"`
}

// TestMonitorStats tracks monitoring statistics
type TestMonitorStats struct {
	ActiveTests         int32              `json:"active_tests"`
	CompletedTests      int32              `json:"completed_tests"`
	FailedTests         int32              `json:"failed_tests"`
	CurrentRPS          float64            `json:"current_rps"`
	CurrentResponseTime time.Duration      `json:"current_response_time"`
	CurrentErrorRate    float64            `json:"current_error_rate"`
	ResourceUsage       map[string]float64 `json:"resource_usage"`
	LastUpdated         time.Time          `json:"last_updated"`
}

// NewPerformanceTestManager creates a new performance test manager
func NewPerformanceTestManager(config *PerformanceTestConfig, logger *zap.Logger) *PerformanceTestManager {
	if config == nil {
		config = DefaultPerformanceTestConfig()
	}

	manager := &PerformanceTestManager{
		config:   config,
		logger:   logger,
		results:  &TestResults{GeneratedAt: time.Now()},
		stopChan: make(chan struct{}),
	}

	// Initialize components
	manager.loadTester = NewPerformanceLoadTester(config, logger)
	manager.stressTester = NewStressTester(config, logger)
	manager.benchmarker = NewBenchmarker(config, logger)
	manager.failureInjector = NewFailureInjector(config, logger)
	manager.reporter = NewTestReporter(config, logger)
	manager.monitor = NewTestMonitor(config, logger)

	return manager
}

// DefaultPerformanceTestConfig returns default performance test configuration
func DefaultPerformanceTestConfig() *PerformanceTestConfig {
	return &PerformanceTestConfig{
		// Test Configuration
		TestEnabled:     true,
		TestDirectory:   "/tmp/performance_tests",
		TestTimeout:     30 * time.Minute,
		TestRetries:     3,
		TestParallelism: runtime.NumCPU(),

		// Load Testing Configuration
		LoadTestEnabled:  true,
		LoadTestDuration: 5 * time.Minute,
		LoadTestRPS:      100,
		LoadTestUsers:    50,
		LoadTestRampUp:   30 * time.Second,
		LoadTestRampDown: 30 * time.Second,

		// Stress Testing Configuration
		StressTestEnabled:  true,
		StressTestDuration: 10 * time.Minute,
		StressTestMaxRPS:   1000,
		StressTestStep:     50,
		StressTestTimeout:  2 * time.Minute,

		// Benchmark Configuration
		BenchmarkEnabled:    true,
		BenchmarkIterations: 1000,
		BenchmarkWarmup:     5 * time.Second,
		BenchmarkCooldown:   5 * time.Second,

		// Failure Injection Configuration
		FailureInjectionEnabled: false,
		FailureRate:             0.1, // 10%
		FailureTypes:            []string{"timeout", "error", "slow"},
		FailureDuration:         30 * time.Second,

		// Reporting Configuration
		ReportEnabled:   true,
		ReportFormat:    "json",
		ReportDirectory: "/tmp/performance_reports",
		ReportRetention: 10,
		ReportThresholds: map[string]float64{
			"response_time_p95": 500,  // ms
			"error_rate":        0.05, // 5%
			"rps":               100,
		},

		// Monitoring Configuration
		MonitorEnabled:  true,
		MonitorInterval: 5 * time.Second,
		MonitorMetrics:  []string{"response_time", "error_rate", "rps", "memory", "cpu"},
		MonitorAlerts:   true,
	}
}

// RunAllTests runs all configured performance tests
func (ptm *PerformanceTestManager) RunAllTests(ctx context.Context) (*TestResults, error) {
	ptm.mu.Lock()
	defer ptm.mu.Unlock()

	ptm.logger.Info("starting comprehensive performance test suite")

	// Create test directory
	if err := os.MkdirAll(ptm.config.TestDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create test directory: %w", err)
	}

	// Start monitoring
	if ptm.config.MonitorEnabled {
		go ptm.monitor.Start(ctx)
	}

	var wg sync.WaitGroup
	var testErrors []error
	var mu sync.Mutex

	// Run load test
	if ptm.config.LoadTestEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := ptm.loadTester.RunLoadTest(ctx)
			if err != nil {
				mu.Lock()
				testErrors = append(testErrors, fmt.Errorf("load test failed: %w", err))
				mu.Unlock()
			} else {
				ptm.results.LoadTestResults = results
			}
		}()
	}

	// Run stress test
	if ptm.config.StressTestEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := ptm.stressTester.RunStressTest(ctx)
			if err != nil {
				mu.Lock()
				testErrors = append(testErrors, fmt.Errorf("stress test failed: %w", err))
				mu.Unlock()
			} else {
				ptm.results.StressTestResults = results
			}
		}()
	}

	// Run benchmarks
	if ptm.config.BenchmarkEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := ptm.benchmarker.RunBenchmarks(ctx)
			if err != nil {
				mu.Lock()
				testErrors = append(testErrors, fmt.Errorf("benchmarks failed: %w", err))
				mu.Unlock()
			} else {
				ptm.results.BenchmarkResults = results
			}
		}()
	}

	// Wait for all tests to complete
	wg.Wait()

	// Generate overall score and recommendations
	ptm.generateOverallScore()
	ptm.generateRecommendations()

	// Generate report
	if ptm.config.ReportEnabled {
		if err := ptm.reporter.GenerateReport(ptm.results); err != nil {
			ptm.logger.Error("failed to generate report", zap.Error(err))
		}
	}

	// Check for test errors
	if len(testErrors) > 0 {
		return ptm.results, fmt.Errorf("some tests failed: %v", testErrors)
	}

	ptm.logger.Info("performance test suite completed successfully")
	return ptm.results, nil
}

// RunLoadTest runs load testing
func (ptm *PerformanceTestManager) RunLoadTest(ctx context.Context) (*LoadTestStats, error) {
	return ptm.loadTester.RunLoadTest(ctx)
}

// RunStressTest runs stress testing
func (ptm *PerformanceTestManager) RunStressTest(ctx context.Context) (*StressTestStats, error) {
	return ptm.stressTester.RunStressTest(ctx)
}

// RunBenchmarks runs performance benchmarks
func (ptm *PerformanceTestManager) RunBenchmarks(ctx context.Context) ([]*BenchmarkStats, error) {
	return ptm.benchmarker.RunBenchmarks(ctx)
}

// GetResults returns current test results
func (ptm *PerformanceTestManager) GetResults() *TestResults {
	ptm.mu.RLock()
	defer ptm.mu.RUnlock()
	return ptm.results
}

// Shutdown gracefully shuts down the performance test manager
func (ptm *PerformanceTestManager) Shutdown() error {
	select {
	case <-ptm.stopChan:
		// Already shut down
		return nil
	default:
		close(ptm.stopChan)
	}
	return nil
}

// generateOverallScore calculates overall performance score
func (ptm *PerformanceTestManager) generateOverallScore() {
	score := 100.0

	// Deduct points based on load test results
	if ptm.results.LoadTestResults != nil {
		if ptm.results.LoadTestResults.ErrorRate > 0.05 {
			score -= 20
		}
		if ptm.results.LoadTestResults.P95ResponseTime > 500*time.Millisecond {
			score -= 15
		}
		if ptm.results.LoadTestResults.RequestsPerSecond < 100 {
			score -= 10
		}
	}

	// Deduct points based on stress test results
	if ptm.results.StressTestResults != nil {
		if ptm.results.StressTestResults.ErrorRateAtMax > 0.1 {
			score -= 15
		}
		if ptm.results.StressTestResults.MaxRPS < 500 {
			score -= 10
		}
	}

	// Deduct points based on benchmark results
	for _, benchmark := range ptm.results.BenchmarkResults {
		if benchmark.OperationsPerSecond < 1000 {
			score -= 5
		}
	}

	ptm.results.OverallScore = max(0, score)
}

// generateRecommendations generates performance recommendations
func (ptm *PerformanceTestManager) generateRecommendations() {
	var recommendations []string

	// Load test recommendations
	if ptm.results.LoadTestResults != nil {
		if ptm.results.LoadTestResults.ErrorRate > 0.05 {
			recommendations = append(recommendations, "Reduce error rate by improving error handling and retry logic")
		}
		if ptm.results.LoadTestResults.P95ResponseTime > 500*time.Millisecond {
			recommendations = append(recommendations, "Optimize response time by improving database queries and caching")
		}
		if ptm.results.LoadTestResults.RequestsPerSecond < 100 {
			recommendations = append(recommendations, "Increase throughput by optimizing resource utilization and scaling")
		}
	}

	// Stress test recommendations
	if ptm.results.StressTestResults != nil {
		if ptm.results.StressTestResults.ErrorRateAtMax > 0.1 {
			recommendations = append(recommendations, "Improve system resilience under high load")
		}
		if ptm.results.StressTestResults.MaxRPS < 500 {
			recommendations = append(recommendations, "Increase maximum throughput capacity")
		}
	}

	// Benchmark recommendations
	for _, benchmark := range ptm.results.BenchmarkResults {
		if benchmark.OperationsPerSecond < 1000 {
			recommendations = append(recommendations, fmt.Sprintf("Optimize %s operation performance", benchmark.OperationName))
		}
	}

	ptm.results.Recommendations = recommendations
}

// NewPerformanceLoadTester creates a new performance load tester
func NewPerformanceLoadTester(config *PerformanceTestConfig, logger *zap.Logger) *PerformanceLoadTester {
	return &PerformanceLoadTester{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		stats: &LoadTestStats{},
	}
}

// RunLoadTest performs load testing
func (lt *PerformanceLoadTester) RunLoadTest(ctx context.Context) (*LoadTestStats, error) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	lt.logger.Info("starting load test",
		zap.Duration("duration", lt.config.LoadTestDuration),
		zap.Int("rps", lt.config.LoadTestRPS),
		zap.Int("users", lt.config.LoadTestUsers))

	lt.stats.StartTime = time.Now()
	lt.stats.Duration = lt.config.LoadTestDuration

	// Create test scenarios
	scenarios := lt.createTestScenarios()

	// Run load test
	var wg sync.WaitGroup
	requestChan := make(chan *TestRequest, lt.config.LoadTestRPS*10)
	responseChan := make(chan *TestResponse, lt.config.LoadTestRPS*10)

	// Start request generator
	wg.Add(1)
	go func() {
		defer wg.Done()
		lt.generateRequests(ctx, requestChan, scenarios)
	}()

	// Start response processor
	wg.Add(1)
	go func() {
		defer wg.Done()
		lt.processResponses(ctx, responseChan)
	}()

	// Start workers
	for i := 0; i < lt.config.LoadTestUsers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			lt.worker(ctx, workerID, requestChan, responseChan)
		}(i)
	}

	// Wait for completion
	wg.Wait()
	close(requestChan)
	close(responseChan)

	lt.stats.EndTime = time.Now()
	lt.stats.Duration = lt.stats.EndTime.Sub(lt.stats.StartTime)

	// Calculate final statistics
	lt.calculateFinalStats()

	lt.logger.Info("load test completed",
		zap.Int64("total_requests", lt.stats.TotalRequests),
		zap.Int64("successful_requests", lt.stats.SuccessfulRequests),
		zap.Int64("failed_requests", lt.stats.FailedRequests),
		zap.Float64("error_rate", lt.stats.ErrorRate),
		zap.Duration("avg_response_time", lt.stats.AverageResponseTime))

	return lt.stats, nil
}

// TestRequest represents a test request
type TestRequest struct {
	ID        string
	Method    string
	URL       string
	Headers   map[string]string
	Body      []byte
	StartTime time.Time
}

// TestResponse represents a test response
type TestResponse struct {
	RequestID    string
	StatusCode   int
	ResponseTime time.Duration
	Error        error
	StartTime    time.Time
	EndTime      time.Time
}

// createTestScenarios creates test scenarios for load testing
func (lt *PerformanceLoadTester) createTestScenarios() []*TestRequest {
	scenarios := []*TestRequest{
		{
			ID:      "classification_request",
			Method:  "POST",
			URL:     "/api/v1/classify",
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    []byte(`{"text": "technology company specializing in software development"}`),
		},
		{
			ID:      "verification_request",
			Method:  "POST",
			URL:     "/api/v1/verify",
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    []byte(`{"business_name": "TechCorp", "website": "https://techcorp.com"}`),
		},
		{
			ID:      "health_check",
			Method:  "GET",
			URL:     "/health",
			Headers: map[string]string{},
			Body:    []byte{},
		},
	}
	return scenarios
}

// generateRequests generates test requests
func (lt *PerformanceLoadTester) generateRequests(ctx context.Context, requestChan chan<- *TestRequest, scenarios []*TestRequest) {
	ticker := time.NewTicker(time.Second / time.Duration(lt.config.LoadTestRPS))
	defer ticker.Stop()

	deadline := time.Now().Add(lt.config.LoadTestDuration)
	requestID := int64(0)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Select random scenario
			scenario := scenarios[rand.Intn(len(scenarios))]
			request := &TestRequest{
				ID:        fmt.Sprintf("req_%d", atomic.AddInt64(&requestID, 1)),
				Method:    scenario.Method,
				URL:       scenario.URL,
				Headers:   scenario.Headers,
				Body:      scenario.Body,
				StartTime: time.Now(),
			}
			requestChan <- request
		}
	}
}

// worker processes requests
func (lt *PerformanceLoadTester) worker(ctx context.Context, workerID int, requestChan <-chan *TestRequest, responseChan chan<- *TestResponse) {
	for request := range requestChan {
		select {
		case <-ctx.Done():
			return
		default:
		}

		response := lt.executeRequest(request)
		responseChan <- response
	}
}

// executeRequest executes a single test request
func (lt *PerformanceLoadTester) executeRequest(request *TestRequest) *TestResponse {
	startTime := time.Now()

	// Create HTTP request
	req, err := http.NewRequest(request.Method, request.URL, nil)
	if err != nil {
		return &TestResponse{
			RequestID:    request.ID,
			StatusCode:   0,
			ResponseTime: time.Since(startTime),
			Error:        err,
			StartTime:    startTime,
			EndTime:      time.Now(),
		}
	}

	// Add headers
	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := lt.client.Do(req)
	endTime := time.Now()
	responseTime := endTime.Sub(startTime)

	if err != nil {
		return &TestResponse{
			RequestID:    request.ID,
			StatusCode:   0,
			ResponseTime: responseTime,
			Error:        err,
			StartTime:    startTime,
			EndTime:      endTime,
		}
	}

	defer resp.Body.Close()

	return &TestResponse{
		RequestID:    request.ID,
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
		Error:        nil,
		StartTime:    startTime,
		EndTime:      endTime,
	}
}

// processResponses processes test responses
func (lt *PerformanceLoadTester) processResponses(ctx context.Context, responseChan <-chan *TestResponse) {
	var responseTimes []time.Duration

	for response := range responseChan {
		select {
		case <-ctx.Done():
			return
		default:
		}

		atomic.AddInt64(&lt.stats.TotalRequests, 1)

		if response.Error != nil || response.StatusCode >= 400 {
			atomic.AddInt64(&lt.stats.FailedRequests, 1)
		} else {
			atomic.AddInt64(&lt.stats.SuccessfulRequests, 1)
		}

		responseTimes = append(responseTimes, response.ResponseTime)

		// Update min/max response times
		lt.mu.Lock()
		if lt.stats.MinResponseTime == 0 || response.ResponseTime < lt.stats.MinResponseTime {
			lt.stats.MinResponseTime = response.ResponseTime
		}
		if response.ResponseTime > lt.stats.MaxResponseTime {
			lt.stats.MaxResponseTime = response.ResponseTime
		}
		lt.mu.Unlock()
	}

	// Calculate percentiles
	if len(responseTimes) > 0 {
		lt.calculatePercentiles(responseTimes)
	}
}

// calculatePercentiles calculates response time percentiles
func (lt *PerformanceLoadTester) calculatePercentiles(responseTimes []time.Duration) {
	// Sort response times
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	n := len(responseTimes)
	if n == 0 {
		return
	}

	// Calculate percentiles with proper indexing
	lt.stats.P50ResponseTime = responseTimes[(n-1)*50/100]
	lt.stats.P95ResponseTime = responseTimes[(n-1)*95/100]
	lt.stats.P99ResponseTime = responseTimes[(n-1)*99/100]

	// Calculate average
	total := time.Duration(0)
	for _, rt := range responseTimes {
		total += rt
	}
	lt.stats.AverageResponseTime = total / time.Duration(n)
}

// calculateFinalStats calculates final load test statistics
func (lt *PerformanceLoadTester) calculateFinalStats() {
	if lt.stats.TotalRequests > 0 {
		lt.stats.ErrorRate = float64(lt.stats.FailedRequests) / float64(lt.stats.TotalRequests)
		lt.stats.RequestsPerSecond = float64(lt.stats.TotalRequests) / lt.stats.Duration.Seconds()
	}
}

// NewStressTester creates a new stress tester
func NewStressTester(config *PerformanceTestConfig, logger *zap.Logger) *StressTester {
	return &StressTester{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		stats: &StressTestStats{},
	}
}

// RunStressTest performs stress testing
func (st *StressTester) RunStressTest(ctx context.Context) (*StressTestStats, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.logger.Info("starting stress test",
		zap.Duration("duration", st.config.StressTestDuration),
		zap.Int("max_rps", st.config.StressTestMaxRPS),
		zap.Int("step", st.config.StressTestStep))

	st.stats.StartTime = time.Now()
	st.stats.Duration = st.config.StressTestDuration

	// Test different RPS levels
	for rps := st.config.StressTestStep; rps <= st.config.StressTestMaxRPS; rps += st.config.StressTestStep {
		select {
		case <-ctx.Done():
			return st.stats, ctx.Err()
		default:
		}

		st.logger.Info("testing RPS level", zap.Int("rps", rps))

		// Run test at current RPS level
		errorRate, avgResponseTime := st.testRPSLevel(ctx, rps)

		// Check if breaking point reached
		if errorRate > 0.1 || avgResponseTime > 2*time.Second {
			st.stats.BreakingPoint = rps - st.config.StressTestStep
			st.stats.ResponseTimeAtMax = avgResponseTime
			st.stats.ErrorRateAtMax = errorRate
			break
		}

		st.stats.MaxRPS = rps
	}

	st.stats.EndTime = time.Now()
	st.stats.Duration = st.stats.EndTime.Sub(st.stats.StartTime)

	st.logger.Info("stress test completed",
		zap.Int("max_rps", st.stats.MaxRPS),
		zap.Int("breaking_point", st.stats.BreakingPoint),
		zap.Float64("error_rate_at_max", st.stats.ErrorRateAtMax))

	return st.stats, nil
}

// testRPSLevel tests a specific RPS level
func (st *StressTester) testRPSLevel(ctx context.Context, rps int) (float64, time.Duration) {
	// Implementation for testing specific RPS level
	// This would run a short load test at the specified RPS
	// and return error rate and average response time

	// For now, return mock values
	return 0.05, 500 * time.Millisecond
}

// NewBenchmarker creates a new benchmarker
func NewBenchmarker(config *PerformanceTestConfig, logger *zap.Logger) *Benchmarker {
	return &Benchmarker{
		config: config,
		logger: logger,
		stats:  &BenchmarkStats{},
	}
}

// RunBenchmarks runs performance benchmarks
func (bm *Benchmarker) RunBenchmarks(ctx context.Context) ([]*BenchmarkStats, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.logger.Info("starting benchmarks",
		zap.Int("iterations", bm.config.BenchmarkIterations))

	var results []*BenchmarkStats

	// Benchmark different operations
	operations := []string{"classification", "verification", "data_extraction", "risk_assessment"}

	for _, operation := range operations {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		bm.logger.Info("benchmarking operation", zap.String("operation", operation))

		result := bm.benchmarkOperation(ctx, operation)
		results = append(results, result)
	}

	bm.logger.Info("benchmarks completed", zap.Int("operations", len(results)))
	return results, nil
}

// benchmarkOperation benchmarks a specific operation
func (bm *Benchmarker) benchmarkOperation(ctx context.Context, operation string) *BenchmarkStats {
	stats := &BenchmarkStats{
		OperationName: operation,
		Iterations:    bm.config.BenchmarkIterations,
		StartTime:     time.Now(),
	}

	// Warmup
	time.Sleep(bm.config.BenchmarkWarmup)

	var responseTimes []time.Duration
	var m runtime.MemStats

	// Run iterations
	for i := 0; i < bm.config.BenchmarkIterations; i++ {
		select {
		case <-ctx.Done():
			return stats
		default:
		}

		start := time.Now()

		// Execute operation (mock implementation)
		bm.executeOperation(operation)

		responseTime := time.Since(start)
		responseTimes = append(responseTimes, responseTime)

		// Update min/max
		if stats.MinTime == 0 || responseTime < stats.MinTime {
			stats.MinTime = responseTime
		}
		if responseTime > stats.MaxTime {
			stats.MaxTime = responseTime
		}
	}

	// Cooldown
	time.Sleep(bm.config.BenchmarkCooldown)

	// Calculate statistics
	bm.calculateBenchmarkStats(stats, responseTimes)

	// Get memory stats
	runtime.ReadMemStats(&m)
	stats.MemoryUsage = m.Alloc

	stats.EndTime = time.Now()
	stats.TotalTime = stats.EndTime.Sub(stats.StartTime)

	return stats
}

// executeOperation executes a benchmark operation
func (bm *Benchmarker) executeOperation(operation string) {
	// Mock implementation - in real scenario, this would call actual operations
	switch operation {
	case "classification":
		time.Sleep(10 * time.Millisecond) // Simulate classification time
	case "verification":
		time.Sleep(15 * time.Millisecond) // Simulate verification time
	case "data_extraction":
		time.Sleep(20 * time.Millisecond) // Simulate data extraction time
	case "risk_assessment":
		time.Sleep(25 * time.Millisecond) // Simulate risk assessment time
	}
}

// calculateBenchmarkStats calculates benchmark statistics
func (bm *Benchmarker) calculateBenchmarkStats(stats *BenchmarkStats, responseTimes []time.Duration) {
	if len(responseTimes) == 0 {
		return
	}

	// Sort for percentiles
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	n := len(responseTimes)
	stats.P50Time = responseTimes[(n-1)*50/100]
	stats.P95Time = responseTimes[(n-1)*95/100]
	stats.P99Time = responseTimes[(n-1)*99/100]

	// Calculate average
	total := time.Duration(0)
	for _, rt := range responseTimes {
		total += rt
	}
	stats.AverageTime = total / time.Duration(n)

	// Calculate operations per second
	stats.OperationsPerSecond = float64(n) / stats.TotalTime.Seconds()
}

// NewFailureInjector creates a new failure injector
func NewFailureInjector(config *PerformanceTestConfig, logger *zap.Logger) *FailureInjector {
	return &FailureInjector{
		config: config,
		logger: logger,
	}
}

// NewTestReporter creates a new test reporter
func NewTestReporter(config *PerformanceTestConfig, logger *zap.Logger) *TestReporter {
	return &TestReporter{
		config: config,
		logger: logger,
	}
}

// GenerateReport generates a test report
func (tr *TestReporter) GenerateReport(results *TestResults) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	// Create report directory
	if err := os.MkdirAll(tr.config.ReportDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Generate report filename
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join(tr.config.ReportDirectory, fmt.Sprintf("performance_report_%s.json", timestamp))

	// Write report
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	tr.logger.Info("performance report generated", zap.String("filename", filename))

	// Clean up old reports
	tr.cleanupOldReports()

	return nil
}

// cleanupOldReports removes old reports based on retention policy
func (tr *TestReporter) cleanupOldReports() {
	files, err := os.ReadDir(tr.config.ReportDirectory)
	if err != nil {
		tr.logger.Error("failed to read report directory", zap.Error(err))
		return
	}

	// Sort files by modification time
	var reportFiles []os.FileInfo
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			if info, err := file.Info(); err == nil {
				reportFiles = append(reportFiles, info)
			}
		}
	}

	// Remove old files if we have too many
	if len(reportFiles) > tr.config.ReportRetention {
		// Sort by modification time (oldest first)
		sort.Slice(reportFiles, func(i, j int) bool {
			return reportFiles[i].ModTime().Before(reportFiles[j].ModTime())
		})

		// Remove oldest files
		for i := 0; i < len(reportFiles)-tr.config.ReportRetention; i++ {
			oldFile := filepath.Join(tr.config.ReportDirectory, reportFiles[i].Name())
			if err := os.Remove(oldFile); err != nil {
				tr.logger.Error("failed to remove old report", zap.String("file", oldFile), zap.Error(err))
			}
		}
	}
}

// NewTestMonitor creates a new test monitor
func NewTestMonitor(config *PerformanceTestConfig, logger *zap.Logger) *TestMonitor {
	return &TestMonitor{
		config: config,
		logger: logger,
		stats:  &TestMonitorStats{},
	}
}

// Start starts the test monitor
func (tm *TestMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(tm.config.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tm.collectMetrics()
		}
	}
}

// collectMetrics collects monitoring metrics
func (tm *TestMonitor) collectMetrics() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Update timestamp
	tm.stats.LastUpdated = time.Now()

	// Collect system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	tm.stats.ResourceUsage = map[string]float64{
		"memory_usage_mb":   float64(m.Alloc) / 1024 / 1024,
		"cpu_usage_percent": 0.0, // Would need external monitoring for CPU
	}

	// Log metrics if alerts enabled
	if tm.config.MonitorAlerts {
		if tm.stats.CurrentErrorRate > 0.1 {
			tm.logger.Warn("high error rate detected",
				zap.Float64("error_rate", tm.stats.CurrentErrorRate))
		}
		if tm.stats.CurrentResponseTime > 1*time.Second {
			tm.logger.Warn("high response time detected",
				zap.Duration("response_time", tm.stats.CurrentResponseTime))
		}
	}
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
