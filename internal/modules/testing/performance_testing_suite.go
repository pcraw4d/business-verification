package testing

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PerformanceTest represents a single performance test
type PerformanceTest struct {
	ID          string
	Name        string
	Description string
	Function    PerformanceTestFunction
	Config      PerformanceTestConfig
	Tags        []string
	Components  []string
	Timeout     time.Duration
	Parallel    bool
	Skipped     bool
}

// PerformanceTestFunction defines the signature for performance test functions
type PerformanceTestFunction func(ctx *PerformanceContext) error

// PerformanceTestConfig defines configuration for performance tests
type PerformanceTestConfig struct {
	Iterations   int           // Number of iterations to run
	Concurrency  int           // Number of concurrent goroutines
	Duration     time.Duration // Test duration
	WarmupTime   time.Duration // Warmup period before measurements
	CooldownTime time.Duration // Cooldown period after measurements
	RequestRate  int           // Requests per second (for load tests)
	RampUpTime   time.Duration // Time to ramp up to full load
	RampDownTime time.Duration // Time to ramp down from full load
	Thresholds   PerformanceThresholds
	Metrics      []string // Metrics to collect
	Baseline     *PerformanceBaseline
	Regression   bool // Whether to check for performance regression
}

// PerformanceThresholds defines performance thresholds
type PerformanceThresholds struct {
	MaxResponseTime time.Duration
	MinThroughput   float64
	MaxErrorRate    float64
	MaxMemoryUsage  int64 // bytes
	MaxCPUUsage     float64
	MaxLatencyP95   time.Duration
	MaxLatencyP99   time.Duration
}

// PerformanceBaseline represents baseline performance metrics
type PerformanceBaseline struct {
	ResponseTime time.Duration
	Throughput   float64
	ErrorRate    float64
	MemoryUsage  int64
	CPUUsage     float64
	LatencyP95   time.Duration
	LatencyP99   time.Duration
	Timestamp    time.Time
	Version      string
}

// PerformanceContext provides context for performance tests
type PerformanceContext struct {
	context.Context
	Logger       *zap.Logger
	T            *PerformanceTest
	StartTime    time.Time
	EndTime      time.Time
	Metrics      *PerformanceMetrics
	CleanupFuncs []func()
	cancel       context.CancelFunc
	mu           sync.Mutex
}

// PerformanceMetrics collects performance measurements
type PerformanceMetrics struct {
	ResponseTimes []time.Duration
	Throughput    float64
	ErrorCount    int
	TotalRequests int
	MemoryUsage   []int64
	CPUUsage      []float64
	StartTime     time.Time
	EndTime       time.Time
	mu            sync.RWMutex
}

// PerformanceTestSuite represents a collection of performance tests
type PerformanceTestSuite struct {
	Name        string
	Description string
	Tests       []*PerformanceTest
	Setup       func(ctx *PerformanceContext) error
	Teardown    func(ctx *PerformanceContext) error
	BeforeEach  func(ctx *PerformanceContext) error
	AfterEach   func(ctx *PerformanceContext) error
	Parallel    bool
	Timeout     time.Duration
	Tags        []string
	Components  []string
	Config      PerformanceTestConfig
}

// PerformanceTestRunner manages performance test execution
type PerformanceTestRunner struct {
	Suites  []*PerformanceTestSuite
	Config  PerformanceTestConfig
	Logger  *zap.Logger
	Results []*PerformanceTestResult
	mu      sync.Mutex
}

// PerformanceTestResult represents the result of a performance test
type PerformanceTestResult struct {
	Test            *PerformanceTest
	Status          PerformanceTestStatus
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	Metrics         *PerformanceMetrics
	Baseline        *PerformanceBaseline
	Regression      bool
	Thresholds      PerformanceThresholds
	Passed          bool
	Error           error
	Recommendations []string
}

// PerformanceTestType defines different types of performance tests
type PerformanceTestType string

const (
	BenchmarkTest   PerformanceTestType = "benchmark"
	LoadTest        PerformanceTestType = "load"
	StressTest      PerformanceTestType = "stress"
	SpikeTest       PerformanceTestType = "spike"
	SoakTest        PerformanceTestType = "soak"
	ScalabilityTest PerformanceTestType = "scalability"
	EnduranceTest   PerformanceTestType = "endurance"
)

// PerformanceTestStatus represents the status of a performance test
type PerformanceTestStatus string

const (
	PerfStatusPassed     PerformanceTestStatus = "passed"
	PerfStatusFailed     PerformanceTestStatus = "failed"
	PerfStatusSkipped    PerformanceTestStatus = "skipped"
	PerfStatusError      PerformanceTestStatus = "error"
	PerfStatusRegression PerformanceTestStatus = "regression"
)

// NewPerformanceTest creates a new performance test
func NewPerformanceTest(name string, function PerformanceTestFunction) *PerformanceTest {
	return &PerformanceTest{
		ID:         generatePerformanceTestID(),
		Name:       name,
		Function:   function,
		Config:     DefaultPerformanceTestConfig(),
		Tags:       []string{},
		Components: []string{},
		Timeout:    30 * time.Second,
		Parallel:   false,
		Skipped:    false,
	}
}

// DefaultPerformanceTestConfig returns default performance test configuration
func DefaultPerformanceTestConfig() PerformanceTestConfig {
	return PerformanceTestConfig{
		Iterations:   100,
		Concurrency:  1,
		Duration:     60 * time.Second,
		WarmupTime:   5 * time.Second,
		CooldownTime: 5 * time.Second,
		RequestRate:  100,
		RampUpTime:   10 * time.Second,
		RampDownTime: 10 * time.Second,
		Thresholds:   DefaultPerformanceThresholds(),
		Metrics:      []string{"response_time", "throughput", "error_rate", "memory", "cpu"},
		Baseline:     nil,
		Regression:   false,
	}
}

// DefaultPerformanceThresholds returns default performance thresholds
func DefaultPerformanceThresholds() PerformanceThresholds {
	return PerformanceThresholds{
		MaxResponseTime: 1 * time.Second,
		MinThroughput:   100.0,
		MaxErrorRate:    0.01,              // 1%
		MaxMemoryUsage:  100 * 1024 * 1024, // 100MB
		MaxCPUUsage:     80.0,              // 80%
		MaxLatencyP95:   500 * time.Millisecond,
		MaxLatencyP99:   1 * time.Second,
	}
}

// AddTag adds a tag to the performance test
func (pt *PerformanceTest) AddTag(tag string) *PerformanceTest {
	pt.Tags = append(pt.Tags, tag)
	return pt
}

// SetConfig sets the performance test configuration
func (pt *PerformanceTest) SetConfig(config PerformanceTestConfig) *PerformanceTest {
	pt.Config = config
	return pt
}

// SetTimeout sets the test timeout
func (pt *PerformanceTest) SetTimeout(timeout time.Duration) *PerformanceTest {
	pt.Timeout = timeout
	return pt
}

// SetParallel sets whether the test can run in parallel
func (pt *PerformanceTest) SetParallel(parallel bool) *PerformanceTest {
	pt.Parallel = parallel
	return pt
}

// Skip marks the test as skipped
func (pt *PerformanceTest) Skip() *PerformanceTest {
	pt.Skipped = true
	return pt
}

// AddComponent adds a component to the performance test
func (pt *PerformanceTest) AddComponent(component string) *PerformanceTest {
	pt.Components = append(pt.Components, component)
	return pt
}

// NewPerformanceTestSuite creates a new performance test suite
func NewPerformanceTestSuite(name string) *PerformanceTestSuite {
	return &PerformanceTestSuite{
		Name:        name,
		Description: "",
		Tests:       []*PerformanceTest{},
		Setup:       nil,
		Teardown:    nil,
		BeforeEach:  nil,
		AfterEach:   nil,
		Parallel:    false,
		Timeout:     5 * time.Minute,
		Tags:        []string{},
		Components:  []string{},
		Config:      DefaultPerformanceTestConfig(),
	}
}

// AddTest adds a test to the suite
func (pts *PerformanceTestSuite) AddTest(test *PerformanceTest) *PerformanceTestSuite {
	pts.Tests = append(pts.Tests, test)
	return pts
}

// CreateTest creates and adds a new test to the suite
func (pts *PerformanceTestSuite) CreateTest(name string, function PerformanceTestFunction) *PerformanceTest {
	test := NewPerformanceTest(name, function)
	pts.AddTest(test)
	return test
}

// SetSetup sets the suite setup function
func (pts *PerformanceTestSuite) SetSetup(setup func(ctx *PerformanceContext) error) *PerformanceTestSuite {
	pts.Setup = setup
	return pts
}

// SetTeardown sets the suite teardown function
func (pts *PerformanceTestSuite) SetTeardown(teardown func(ctx *PerformanceContext) error) *PerformanceTestSuite {
	pts.Teardown = teardown
	return pts
}

// SetBeforeEach sets the before each function
func (pts *PerformanceTestSuite) SetBeforeEach(beforeEach func(ctx *PerformanceContext) error) *PerformanceTestSuite {
	pts.BeforeEach = beforeEach
	return pts
}

// SetAfterEach sets the after each function
func (pts *PerformanceTestSuite) SetAfterEach(afterEach func(ctx *PerformanceContext) error) *PerformanceTestSuite {
	pts.AfterEach = afterEach
	return pts
}

// SetParallel sets whether tests can run in parallel
func (pts *PerformanceTestSuite) SetParallel(parallel bool) *PerformanceTestSuite {
	pts.Parallel = parallel
	return pts
}

// SetTimeout sets the suite timeout
func (pts *PerformanceTestSuite) SetTimeout(timeout time.Duration) *PerformanceTestSuite {
	pts.Timeout = timeout
	return pts
}

// AddTag adds a tag to the suite
func (pts *PerformanceTestSuite) AddTag(tag string) *PerformanceTestSuite {
	pts.Tags = append(pts.Tags, tag)
	return pts
}

// AddComponent adds a component to the suite
func (pts *PerformanceTestSuite) AddComponent(component string) *PerformanceTestSuite {
	pts.Components = append(pts.Components, component)
	return pts
}

// SetConfig sets the suite configuration
func (pts *PerformanceTestSuite) SetConfig(config PerformanceTestConfig) *PerformanceTestSuite {
	pts.Config = config
	return pts
}

// NewPerformanceContext creates a new performance context
func NewPerformanceContext(ctx context.Context, test *PerformanceTest, logger *zap.Logger) *PerformanceContext {
	ctx, cancel := context.WithTimeout(ctx, test.Timeout)
	return &PerformanceContext{
		Context:      ctx,
		Logger:       logger,
		T:            test,
		StartTime:    time.Now(),
		Metrics:      NewPerformanceMetrics(),
		CleanupFuncs: []func(){},
		cancel:       cancel,
		mu:           sync.Mutex{},
	}
}

// NewPerformanceMetrics creates new performance metrics
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		ResponseTimes: []time.Duration{},
		Throughput:    0.0,
		ErrorCount:    0,
		TotalRequests: 0,
		MemoryUsage:   []int64{},
		CPUUsage:      []float64{},
		StartTime:     time.Now(),
		EndTime:       time.Time{},
		mu:            sync.RWMutex{},
	}
}

// Cleanup performs cleanup operations
func (pc *PerformanceContext) Cleanup() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Execute cleanup functions in reverse order
	for i := len(pc.CleanupFuncs) - 1; i >= 0; i-- {
		if pc.CleanupFuncs[i] != nil {
			pc.CleanupFuncs[i]()
		}
	}

	if pc.cancel != nil {
		pc.cancel()
	}

	pc.EndTime = time.Now()
	pc.Metrics.EndTime = pc.EndTime
}

// AddCleanup adds a cleanup function
func (pc *PerformanceContext) AddCleanup(cleanup func()) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.CleanupFuncs = append(pc.CleanupFuncs, cleanup)
}

// Log logs a message
func (pc *PerformanceContext) Log(msg string, fields ...zap.Field) {
	if pc.Logger != nil {
		pc.Logger.Info(msg, fields...)
	}
}

// Logf logs a formatted message
func (pc *PerformanceContext) Logf(format string, args ...interface{}) {
	if pc.Logger != nil {
		pc.Logger.Sugar().Infof(format, args...)
	}
}

// RecordResponseTime records a response time measurement
func (pm *PerformanceMetrics) RecordResponseTime(duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.ResponseTimes = append(pm.ResponseTimes, duration)
	pm.TotalRequests++
}

// RecordError records an error
func (pm *PerformanceMetrics) RecordError() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.ErrorCount++
	pm.TotalRequests++
}

// RecordMemoryUsage records memory usage
func (pm *PerformanceMetrics) RecordMemoryUsage(usage int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.MemoryUsage = append(pm.MemoryUsage, usage)
}

// RecordCPUUsage records CPU usage
func (pm *PerformanceMetrics) RecordCPUUsage(usage float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.CPUUsage = append(pm.CPUUsage, usage)
}

// CalculateThroughput calculates throughput (requests per second)
func (pm *PerformanceMetrics) CalculateThroughput() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.EndTime.IsZero() {
		pm.EndTime = time.Now()
	}

	duration := pm.EndTime.Sub(pm.StartTime).Seconds()
	if duration <= 0 {
		return 0.0
	}

	return float64(pm.TotalRequests) / duration
}

// CalculateErrorRate calculates error rate
func (pm *PerformanceMetrics) CalculateErrorRate() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.TotalRequests == 0 {
		return 0.0
	}

	return float64(pm.ErrorCount) / float64(pm.TotalRequests)
}

// CalculateAverageResponseTime calculates average response time
func (pm *PerformanceMetrics) CalculateAverageResponseTime() time.Duration {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.ResponseTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, rt := range pm.ResponseTimes {
		total += rt
	}

	return total / time.Duration(len(pm.ResponseTimes))
}

// CalculatePercentileResponseTime calculates percentile response time
func (pm *PerformanceMetrics) CalculatePercentileResponseTime(percentile float64) time.Duration {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.ResponseTimes) == 0 {
		return 0
	}

	// Create a copy and sort
	times := make([]time.Duration, len(pm.ResponseTimes))
	copy(times, pm.ResponseTimes)

	// Sort response times
	for i := 0; i < len(times)-1; i++ {
		for j := i + 1; j < len(times); j++ {
			if times[i] > times[j] {
				times[i], times[j] = times[j], times[i]
			}
		}
	}

	index := int(math.Floor(percentile / 100.0 * float64(len(times))))
	if index >= len(times) {
		index = len(times) - 1
	}

	return times[index]
}

// CalculateMaxMemoryUsage calculates maximum memory usage
func (pm *PerformanceMetrics) CalculateMaxMemoryUsage() int64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.MemoryUsage) == 0 {
		return 0
	}

	max := pm.MemoryUsage[0]
	for _, usage := range pm.MemoryUsage {
		if usage > max {
			max = usage
		}
	}

	return max
}

// CalculateAverageCPUUsage calculates average CPU usage
func (pm *PerformanceMetrics) CalculateAverageCPUUsage() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.CPUUsage) == 0 {
		return 0.0
	}

	var total float64
	for _, usage := range pm.CPUUsage {
		total += usage
	}

	return total / float64(len(pm.CPUUsage))
}

// NewPerformanceTestRunner creates a new performance test runner
func NewPerformanceTestRunner(logger *zap.Logger) *PerformanceTestRunner {
	return &PerformanceTestRunner{
		Suites:  []*PerformanceTestSuite{},
		Config:  DefaultPerformanceTestConfig(),
		Logger:  logger,
		Results: []*PerformanceTestResult{},
		mu:      sync.Mutex{},
	}
}

// AddSuite adds a test suite to the runner
func (ptr *PerformanceTestRunner) AddSuite(suite *PerformanceTestSuite) *PerformanceTestRunner {
	ptr.mu.Lock()
	defer ptr.mu.Unlock()
	ptr.Suites = append(ptr.Suites, suite)
	return ptr
}

// RunAllSuites runs all test suites
func (ptr *PerformanceTestRunner) RunAllSuites(ctx context.Context) error {
	ptr.mu.Lock()
	defer ptr.mu.Unlock()

	for _, suite := range ptr.Suites {
		if err := ptr.runSuite(ctx, suite); err != nil {
			return fmt.Errorf("suite %s failed: %w", suite.Name, err)
		}
	}

	return nil
}

// runSuite runs a single test suite
func (ptr *PerformanceTestRunner) runSuite(ctx context.Context, suite *PerformanceTestSuite) error {
	ptr.Logger.Info("Running performance test suite", zap.String("suite", suite.Name))

	// Create suite context
	suiteCtx, cancel := context.WithTimeout(ctx, suite.Timeout)
	defer cancel()

	// Run setup if provided
	if suite.Setup != nil {
		testCtx := NewPerformanceContext(suiteCtx, &PerformanceTest{Name: "setup"}, ptr.Logger)
		if err := suite.Setup(testCtx); err != nil {
			return fmt.Errorf("suite setup failed: %w", err)
		}
		testCtx.Cleanup()
	}

	// Run tests
	if suite.Parallel {
		if err := ptr.runTestsParallel(suiteCtx, suite); err != nil {
			return fmt.Errorf("parallel test execution failed: %w", err)
		}
	} else {
		if err := ptr.runTestsSequential(suiteCtx, suite); err != nil {
			return fmt.Errorf("sequential test execution failed: %w", err)
		}
	}

	// Run teardown if provided
	if suite.Teardown != nil {
		testCtx := NewPerformanceContext(suiteCtx, &PerformanceTest{Name: "teardown"}, ptr.Logger)
		if err := suite.Teardown(testCtx); err != nil {
			ptr.Logger.Error("Suite teardown failed", zap.Error(err))
		}
		testCtx.Cleanup()
	}

	return nil
}

// runTestsParallel runs tests in parallel
func (ptr *PerformanceTestRunner) runTestsParallel(ctx context.Context, suite *PerformanceTestSuite) error {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent tests

	for _, test := range suite.Tests {
		if test.Skipped {
			continue
		}

		wg.Add(1)
		go func(t *PerformanceTest) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := ptr.runSingleTest(ctx, suite, t); err != nil {
				ptr.Logger.Error("Test failed", zap.String("test", t.Name), zap.Error(err))
			}
		}(test)
	}

	wg.Wait()
	return nil
}

// runTestsSequential runs tests sequentially
func (ptr *PerformanceTestRunner) runTestsSequential(ctx context.Context, suite *PerformanceTestSuite) error {
	for _, test := range suite.Tests {
		if test.Skipped {
			continue
		}

		if err := ptr.runSingleTest(ctx, suite, test); err != nil {
			return fmt.Errorf("test %s failed: %w", test.Name, err)
		}
	}

	return nil
}

// runSingleTest runs a single test
func (ptr *PerformanceTestRunner) runSingleTest(ctx context.Context, suite *PerformanceTestSuite, test *PerformanceTest) error {
	ptr.Logger.Info("Running performance test", zap.String("test", test.Name))

	// Create test context
	testCtx := NewPerformanceContext(ctx, test, ptr.Logger)
	defer testCtx.Cleanup()

	// Run before each if provided
	if suite.BeforeEach != nil {
		if err := suite.BeforeEach(testCtx); err != nil {
			return fmt.Errorf("before each failed: %w", err)
		}
	}

	// Execute test function
	result := ptr.executeTestFunction(testCtx, test)

	// Run after each if provided
	if suite.AfterEach != nil {
		if err := suite.AfterEach(testCtx); err != nil {
			ptr.Logger.Error("After each failed", zap.Error(err))
		}
	}

	// Store result
	ptr.mu.Lock()
	ptr.Results = append(ptr.Results, result)
	ptr.mu.Unlock()

	return nil
}

// executeTestFunction executes the test function and evaluates results
func (ptr *PerformanceTestRunner) executeTestFunction(ctx *PerformanceContext, test *PerformanceTest) *PerformanceTestResult {
	startTime := time.Now()

	// Execute test function
	err := test.Function(ctx)

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Calculate metrics
	ctx.Metrics.CalculateThroughput()

	// Evaluate results
	result := &PerformanceTestResult{
		Test:       test,
		StartTime:  startTime,
		EndTime:    endTime,
		Duration:   duration,
		Metrics:    ctx.Metrics,
		Baseline:   test.Config.Baseline,
		Thresholds: test.Config.Thresholds,
		Error:      err,
	}

	// Determine status and passed status
	status, passed, regression := ptr.evaluateTestResult(result)
	result.Status = status
	result.Passed = passed
	result.Regression = regression

	// Generate recommendations
	result.Recommendations = ptr.generateRecommendations(result)

	return result
}

// evaluateTestResult evaluates test results against thresholds and baseline
func (ptr *PerformanceTestRunner) evaluateTestResult(result *PerformanceTestResult) (PerformanceTestStatus, bool, bool) {
	if result.Error != nil {
		return PerfStatusError, false, false
	}

	metrics := result.Metrics
	thresholds := result.Thresholds

	// Check thresholds
	passed := true
	regression := false

	// Response time check
	avgResponseTime := metrics.CalculateAverageResponseTime()
	if avgResponseTime > thresholds.MaxResponseTime {
		passed = false
	}

	// Throughput check
	throughput := metrics.CalculateThroughput()
	if throughput < thresholds.MinThroughput {
		passed = false
	}

	// Error rate check
	errorRate := metrics.CalculateErrorRate()
	if errorRate > thresholds.MaxErrorRate {
		passed = false
	}

	// Memory usage check
	maxMemory := metrics.CalculateMaxMemoryUsage()
	if maxMemory > thresholds.MaxMemoryUsage {
		passed = false
	}

	// CPU usage check
	avgCPU := metrics.CalculateAverageCPUUsage()
	if avgCPU > thresholds.MaxCPUUsage {
		passed = false
	}

	// Latency percentile checks
	p95Latency := metrics.CalculatePercentileResponseTime(95)
	if p95Latency > thresholds.MaxLatencyP95 {
		passed = false
	}

	p99Latency := metrics.CalculatePercentileResponseTime(99)
	if p99Latency > thresholds.MaxLatencyP99 {
		passed = false
	}

	// Baseline comparison
	if result.Baseline != nil && result.Test.Config.Regression {
		regression = ptr.checkRegression(result)
		if regression {
			return PerfStatusRegression, false, true
		}
	}

	if passed {
		return PerfStatusPassed, true, false
	}

	return PerfStatusFailed, false, false
}

// checkRegression checks if performance has regressed compared to baseline
func (ptr *PerformanceTestRunner) checkRegression(result *PerformanceTestResult) bool {
	metrics := result.Metrics
	baseline := result.Baseline

	// Define regression thresholds (e.g., 10% degradation)
	regressionThreshold := 0.1

	// Check response time regression
	avgResponseTime := metrics.CalculateAverageResponseTime()
	if baseline.ResponseTime > 0 {
		degradation := float64(avgResponseTime-baseline.ResponseTime) / float64(baseline.ResponseTime)
		if degradation > regressionThreshold {
			return true
		}
	}

	// Check throughput regression
	throughput := metrics.CalculateThroughput()
	if baseline.Throughput > 0 {
		degradation := (baseline.Throughput - throughput) / baseline.Throughput
		if degradation > regressionThreshold {
			return true
		}
	}

	// Check error rate regression
	errorRate := metrics.CalculateErrorRate()
	if baseline.ErrorRate > 0 {
		increase := (errorRate - baseline.ErrorRate) / baseline.ErrorRate
		if increase > regressionThreshold {
			return true
		}
	}

	return false
}

// generateRecommendations generates performance improvement recommendations
func (ptr *PerformanceTestRunner) generateRecommendations(result *PerformanceTestResult) []string {
	var recommendations []string

	metrics := result.Metrics
	thresholds := result.Thresholds

	// Response time recommendations
	avgResponseTime := metrics.CalculateAverageResponseTime()
	if avgResponseTime > thresholds.MaxResponseTime {
		recommendations = append(recommendations,
			fmt.Sprintf("Average response time (%.2fms) exceeds threshold (%.2fms). Consider optimizing database queries, implementing caching, or scaling resources.",
				float64(avgResponseTime.Milliseconds()), float64(thresholds.MaxResponseTime.Milliseconds())))
	}

	// Throughput recommendations
	throughput := metrics.CalculateThroughput()
	if throughput < thresholds.MinThroughput {
		recommendations = append(recommendations,
			fmt.Sprintf("Throughput (%.2f req/s) below threshold (%.2f req/s). Consider increasing concurrency, optimizing algorithms, or scaling horizontally.",
				throughput, thresholds.MinThroughput))
	}

	// Error rate recommendations
	errorRate := metrics.CalculateErrorRate()
	if errorRate > thresholds.MaxErrorRate {
		recommendations = append(recommendations,
			fmt.Sprintf("Error rate (%.2f%%) exceeds threshold (%.2f%%). Review error handling, improve input validation, or check external dependencies.",
				errorRate*100, thresholds.MaxErrorRate*100))
	}

	// Memory usage recommendations
	maxMemory := metrics.CalculateMaxMemoryUsage()
	if maxMemory > thresholds.MaxMemoryUsage {
		recommendations = append(recommendations,
			fmt.Sprintf("Memory usage (%d MB) exceeds threshold (%d MB). Consider memory optimization, garbage collection tuning, or increasing memory limits.",
				maxMemory/(1024*1024), thresholds.MaxMemoryUsage/(1024*1024)))
	}

	// CPU usage recommendations
	avgCPU := metrics.CalculateAverageCPUUsage()
	if avgCPU > thresholds.MaxCPUUsage {
		recommendations = append(recommendations,
			fmt.Sprintf("CPU usage (%.1f%%) exceeds threshold (%.1f%%). Consider algorithm optimization, reducing computational complexity, or scaling CPU resources.",
				avgCPU, thresholds.MaxCPUUsage))
	}

	return recommendations
}

// GenerateSummary generates a summary of all test results
func (ptr *PerformanceTestRunner) GenerateSummary() *PerformanceTestSummary {
	ptr.mu.Lock()
	defer ptr.mu.Unlock()

	summary := &PerformanceTestSummary{
		TotalTests:      len(ptr.Results),
		PassedTests:     0,
		FailedTests:     0,
		SkippedTests:    0,
		ErrorTests:      0,
		RegressionTests: 0,
		TotalDuration:   0,
		Results:         ptr.Results,
		Recommendations: []string{},
	}

	for _, result := range ptr.Results {
		summary.TotalDuration += result.Duration

		switch result.Status {
		case PerfStatusPassed:
			summary.PassedTests++
		case PerfStatusFailed:
			summary.FailedTests++
		case PerfStatusSkipped:
			summary.SkippedTests++
		case PerfStatusError:
			summary.ErrorTests++
		case PerfStatusRegression:
			summary.RegressionTests++
		}

		// Collect recommendations
		summary.Recommendations = append(summary.Recommendations, result.Recommendations...)
	}

	return summary
}

// PerformanceTestSummary represents a summary of performance test results
type PerformanceTestSummary struct {
	TotalTests      int
	PassedTests     int
	FailedTests     int
	SkippedTests    int
	ErrorTests      int
	RegressionTests int
	TotalDuration   time.Duration
	Results         []*PerformanceTestResult
	Recommendations []string
}

// generatePerformanceTestID generates a unique performance test ID
func generatePerformanceTestID() string {
	return fmt.Sprintf("perf_%d", time.Now().UnixNano())
}
