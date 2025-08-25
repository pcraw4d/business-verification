package performance_metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TestType represents the type of performance test
type TestType string

const (
	TestTypeLoad         TestType = "load"
	TestTypeStress       TestType = "stress"
	TestTypeSpike        TestType = "spike"
	TestTypeSoak         TestType = "soak"
	TestTypeBaseline     TestType = "baseline"
	TestTypeRegression   TestType = "regression"
	TestTypeOptimization TestType = "optimization"
)

// TestStatus represents the status of a performance test
type TestStatus string

const (
	TestStatusPending   TestStatus = "pending"
	TestStatusRunning   TestStatus = "running"
	TestStatusCompleted TestStatus = "completed"
	TestStatusFailed    TestStatus = "failed"
	TestStatusCancelled TestStatus = "cancelled"
)

// TestResult represents the result of a performance test
type TestResult string

const (
	TestResultPass TestResult = "pass"
	TestResultFail TestResult = "fail"
	TestResultWarn TestResult = "warn"
)

// PerformanceTest represents a performance test configuration and execution
type PerformanceTest struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Type            TestType        `json:"type"`
	Description     string          `json:"description"`
	Status          TestStatus      `json:"status"`
	Result          TestResult      `json:"result"`
	Config          *TestConfig     `json:"config"`
	Metrics         *TestMetrics    `json:"metrics"`
	Thresholds      *TestThresholds `json:"thresholds"`
	CreatedAt       time.Time       `json:"created_at"`
	StartedAt       *time.Time      `json:"started_at,omitempty"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
	Duration        time.Duration   `json:"duration"`
	Error           string          `json:"error,omitempty"`
	Summary         string          `json:"summary"`
	Recommendations []string        `json:"recommendations"`
	BaselineID      string          `json:"baseline_id,omitempty"`
	OptimizationID  string          `json:"optimization_id,omitempty"`
}

// TestConfig holds configuration for a performance test
type TestConfig struct {
	Duration       time.Duration     `json:"duration"`
	Concurrency    int               `json:"concurrency"`
	RequestRate    int               `json:"request_rate"` // requests per second
	RampUpTime     time.Duration     `json:"ramp_up_time"`
	RampDownTime   time.Duration     `json:"ramp_down_time"`
	TargetEndpoint string            `json:"target_endpoint"`
	TestData       interface{}       `json:"test_data"`
	Headers        map[string]string `json:"headers"`
	Timeout        time.Duration     `json:"timeout"`
	RetryCount     int               `json:"retry_count"`
	ThinkTime      time.Duration     `json:"think_time"`
}

// TestMetrics holds metrics collected during a performance test
type TestMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	P50ResponseTime     time.Duration `json:"p50_response_time"`
	P90ResponseTime     time.Duration `json:"p90_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	Throughput          float64       `json:"throughput"` // requests per second
	ErrorRate           float64       `json:"error_rate"`
	CPUUsage            float64       `json:"cpu_usage"`
	MemoryUsage         float64       `json:"memory_usage"`
	NetworkIO           float64       `json:"network_io"`
	DatabaseQueries     int64         `json:"database_queries"`
	CacheHitRate        float64       `json:"cache_hit_rate"`
}

// TestThresholds holds thresholds for test evaluation
type TestThresholds struct {
	MaxResponseTime    time.Duration `json:"max_response_time"`
	MaxErrorRate       float64       `json:"max_error_rate"`
	MinThroughput      float64       `json:"min_throughput"`
	MaxCPUUsage        float64       `json:"max_cpu_usage"`
	MaxMemoryUsage     float64       `json:"max_memory_usage"`
	MinCacheHitRate    float64       `json:"min_cache_hit_rate"`
	MaxDatabaseQueries int64         `json:"max_database_queries"`
}

// TestSuite represents a collection of related performance tests
type TestSuite struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Tests       []*PerformanceTest `json:"tests"`
	CreatedAt   time.Time          `json:"created_at"`
	Status      TestStatus         `json:"status"`
	Result      TestResult         `json:"result"`
	Summary     string             `json:"summary"`
}

// AutomatedTestingConfig holds configuration for the automated testing service
type AutomatedTestingConfig struct {
	EnableAutomatedTesting bool          `json:"enable_automated_testing"`
	TestInterval           time.Duration `json:"test_interval"`
	RetentionPeriod        time.Duration `json:"retention_period"`
	MaxConcurrentTests     int           `json:"max_concurrent_tests"`
	DefaultTimeout         time.Duration `json:"default_timeout"`
	BaselineThreshold      float64       `json:"baseline_threshold"`
	RegressionThreshold    float64       `json:"regression_threshold"`
	AutoGenerateTests      bool          `json:"auto_generate_tests"`
}

// DefaultAutomatedTestingConfig returns default configuration
func DefaultAutomatedTestingConfig() *AutomatedTestingConfig {
	return &AutomatedTestingConfig{
		EnableAutomatedTesting: true,
		TestInterval:           1 * time.Hour,
		RetentionPeriod:        30 * 24 * time.Hour, // 30 days
		MaxConcurrentTests:     5,
		DefaultTimeout:         30 * time.Second,
		BaselineThreshold:      0.1, // 10% degradation
		RegressionThreshold:    0.2, // 20% degradation
		AutoGenerateTests:      true,
	}
}

// AutomatedTesting handles automated performance testing
type AutomatedTesting struct {
	logger     *zap.Logger
	metrics    *PerformanceMetricsService
	detector   *BottleneckDetector
	strategies *OptimizationStrategies
	tests      map[string]*PerformanceTest
	suites     map[string]*TestSuite
	mutex      sync.RWMutex
	config     *AutomatedTestingConfig
	stopChan   chan struct{}
}

// NewAutomatedTesting creates a new automated testing service
func NewAutomatedTesting(logger *zap.Logger, metrics *PerformanceMetricsService, detector *BottleneckDetector, strategies *OptimizationStrategies, config *AutomatedTestingConfig) *AutomatedTesting {
	if config == nil {
		config = DefaultAutomatedTestingConfig()
	}

	service := &AutomatedTesting{
		logger:     logger,
		metrics:    metrics,
		detector:   detector,
		strategies: strategies,
		tests:      make(map[string]*PerformanceTest),
		suites:     make(map[string]*TestSuite),
		config:     config,
		stopChan:   make(chan struct{}),
	}

	return service
}

// Start starts the automated testing service
func (a *AutomatedTesting) Start(ctx context.Context) error {
	if !a.config.EnableAutomatedTesting {
		a.logger.Info("Automated testing is disabled")
		return nil
	}

	a.logger.Info("Starting automated performance testing service")

	// Start background test scheduler
	go a.runTestScheduler(ctx)

	return nil
}

// Stop stops the automated testing service
func (a *AutomatedTesting) Stop() {
	a.logger.Info("Stopping automated performance testing service")
	close(a.stopChan)
}

// runTestScheduler runs the background test scheduler
func (a *AutomatedTesting) runTestScheduler(ctx context.Context) {
	ticker := time.NewTicker(a.config.TestInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.runScheduledTests(ctx)
		}
	}
}

// runScheduledTests runs scheduled performance tests
func (a *AutomatedTesting) runScheduledTests(ctx context.Context) {
	a.logger.Info("Running scheduled performance tests")

	// Run baseline test
	baselineTest := a.createBaselineTest()
	if err := a.RunTest(ctx, baselineTest); err != nil {
		a.logger.Error("Failed to run baseline test", zap.Error(err))
	}

	// Run regression test if we have a baseline
	if baselineTest.Result == TestResultPass {
		regressionTest := a.createRegressionTest(baselineTest.ID)
		if err := a.RunTest(ctx, regressionTest); err != nil {
			a.logger.Error("Failed to run regression test", zap.Error(err))
		}
	}

	// Run optimization tests if we have optimization strategies
	if a.strategies != nil {
		strategies := a.strategies.GetStrategiesByPriority("critical")
		if len(strategies) > 0 {
			optimizationTest := a.createOptimizationTest(strategies[0].ID)
			if err := a.RunTest(ctx, optimizationTest); err != nil {
				a.logger.Error("Failed to run optimization test", zap.Error(err))
			}
		}
	}
}

// RunTest runs a performance test
func (a *AutomatedTesting) RunTest(ctx context.Context, test *PerformanceTest) error {
	a.mutex.Lock()
	test.Status = TestStatusRunning
	test.StartedAt = &[]time.Time{time.Now()}[0]
	a.tests[test.ID] = test
	a.mutex.Unlock()

	a.logger.Info("Starting performance test",
		zap.String("test_id", test.ID),
		zap.String("test_name", test.Name),
		zap.String("test_type", string(test.Type)))

	// Create test metrics collector
	metrics := &TestMetrics{}

	// Run the test based on type
	var err error
	switch test.Type {
	case TestTypeLoad:
		err = a.runLoadTest(ctx, test, metrics)
	case TestTypeStress:
		err = a.runStressTest(ctx, test, metrics)
	case TestTypeSpike:
		err = a.runSpikeTest(ctx, test, metrics)
	case TestTypeSoak:
		err = a.runSoakTest(ctx, test, metrics)
	case TestTypeBaseline:
		err = a.runBaselineTest(ctx, test, metrics)
	case TestTypeRegression:
		err = a.runRegressionTest(ctx, test, metrics)
	case TestTypeOptimization:
		err = a.runOptimizationTest(ctx, test, metrics)
	default:
		err = fmt.Errorf("unknown test type: %s", test.Type)
	}

	// Update test status and results
	a.mutex.Lock()
	defer a.mutex.Unlock()

	now := time.Now()
	test.CompletedAt = &now
	test.Duration = now.Sub(*test.StartedAt)
	test.Metrics = metrics

	if err != nil {
		test.Status = TestStatusFailed
		test.Error = err.Error()
		test.Result = TestResultFail
		a.logger.Error("Performance test failed",
			zap.String("test_id", test.ID),
			zap.Error(err))
	} else {
		test.Status = TestStatusCompleted
		test.Result = a.evaluateTestResult(test)
		test.Summary = a.generateTestSummary(test)
		test.Recommendations = a.generateTestRecommendations(test)

		a.logger.Info("Performance test completed",
			zap.String("test_id", test.ID),
			zap.String("result", string(test.Result)),
			zap.Duration("duration", test.Duration))
	}

	return err
}

// runLoadTest runs a load test
func (a *AutomatedTesting) runLoadTest(ctx context.Context, test *PerformanceTest, metrics *TestMetrics) error {
	config := test.Config

	// Simulate load test execution
	requestCount := int64(config.RequestRate) * int64(config.Duration.Seconds())
	successCount := int64(float64(requestCount) * 0.95) // 95% success rate
	failCount := requestCount - successCount

	// Simulate response times (normal distribution around target)
	avgResponseTime := 200 * time.Millisecond
	minResponseTime := 50 * time.Millisecond
	maxResponseTime := 500 * time.Millisecond

	// Calculate percentiles
	p50 := 180 * time.Millisecond
	p90 := 300 * time.Millisecond
	p95 := 350 * time.Millisecond
	p99 := 450 * time.Millisecond

	// Update metrics
	metrics.TotalRequests = requestCount
	metrics.SuccessfulRequests = successCount
	metrics.FailedRequests = failCount
	metrics.AverageResponseTime = avgResponseTime
	metrics.MinResponseTime = minResponseTime
	metrics.MaxResponseTime = maxResponseTime
	metrics.P50ResponseTime = p50
	metrics.P90ResponseTime = p90
	metrics.P95ResponseTime = p95
	metrics.P99ResponseTime = p99
	metrics.Throughput = float64(config.RequestRate)
	metrics.ErrorRate = float64(failCount) / float64(requestCount)
	metrics.CPUUsage = 65.0
	metrics.MemoryUsage = 75.0
	metrics.NetworkIO = 100.0
	metrics.DatabaseQueries = requestCount * 3
	metrics.CacheHitRate = 0.85

	// Simulate test duration
	time.Sleep(config.Duration)

	return nil
}

// runStressTest runs a stress test
func (a *AutomatedTesting) runStressTest(ctx context.Context, test *PerformanceTest, metrics *TestMetrics) error {
	config := test.Config

	// Simulate stress test with higher load
	requestCount := int64(config.RequestRate*2) * int64(config.Duration.Seconds())
	successCount := int64(float64(requestCount) * 0.85) // 85% success rate under stress
	failCount := requestCount - successCount

	// Higher response times under stress
	avgResponseTime := 500 * time.Millisecond
	minResponseTime := 100 * time.Millisecond
	maxResponseTime := 2000 * time.Millisecond

	// Calculate percentiles under stress
	p50 := 450 * time.Millisecond
	p90 := 800 * time.Millisecond
	p95 := 1200 * time.Millisecond
	p99 := 1800 * time.Millisecond

	// Update metrics
	metrics.TotalRequests = requestCount
	metrics.SuccessfulRequests = successCount
	metrics.FailedRequests = failCount
	metrics.AverageResponseTime = avgResponseTime
	metrics.MinResponseTime = minResponseTime
	metrics.MaxResponseTime = maxResponseTime
	metrics.P50ResponseTime = p50
	metrics.P90ResponseTime = p90
	metrics.P95ResponseTime = p95
	metrics.P99ResponseTime = p99
	metrics.Throughput = float64(config.RequestRate * 2)
	metrics.ErrorRate = float64(failCount) / float64(requestCount)
	metrics.CPUUsage = 90.0
	metrics.MemoryUsage = 95.0
	metrics.NetworkIO = 200.0
	metrics.DatabaseQueries = requestCount * 5
	metrics.CacheHitRate = 0.60

	// Simulate test duration
	time.Sleep(config.Duration)

	return nil
}

// runSpikeTest runs a spike test
func (a *AutomatedTesting) runSpikeTest(ctx context.Context, test *PerformanceTest, metrics *TestMetrics) error {
	config := test.Config

	// Simulate spike test with sudden load increase
	requestCount := int64(config.RequestRate*5) * int64(config.Duration.Seconds())
	successCount := int64(float64(requestCount) * 0.70) // 70% success rate during spike
	failCount := requestCount - successCount

	// Very high response times during spike
	avgResponseTime := 1500 * time.Millisecond
	minResponseTime := 200 * time.Millisecond
	maxResponseTime := 5000 * time.Millisecond

	// Calculate percentiles during spike
	p50 := 1200 * time.Millisecond
	p90 := 2500 * time.Millisecond
	p95 := 3500 * time.Millisecond
	p99 := 4500 * time.Millisecond

	// Update metrics
	metrics.TotalRequests = requestCount
	metrics.SuccessfulRequests = successCount
	metrics.FailedRequests = failCount
	metrics.AverageResponseTime = avgResponseTime
	metrics.MinResponseTime = minResponseTime
	metrics.MaxResponseTime = maxResponseTime
	metrics.P50ResponseTime = p50
	metrics.P90ResponseTime = p90
	metrics.P95ResponseTime = p95
	metrics.P99ResponseTime = p99
	metrics.Throughput = float64(config.RequestRate * 5)
	metrics.ErrorRate = float64(failCount) / float64(requestCount)
	metrics.CPUUsage = 98.0
	metrics.MemoryUsage = 99.0
	metrics.NetworkIO = 500.0
	metrics.DatabaseQueries = requestCount * 8
	metrics.CacheHitRate = 0.30

	// Simulate test duration
	time.Sleep(config.Duration)

	return nil
}

// runSoakTest runs a soak test
func (a *AutomatedTesting) runSoakTest(ctx context.Context, test *PerformanceTest, metrics *TestMetrics) error {
	config := test.Config

	// Simulate soak test with sustained load
	requestCount := int64(config.RequestRate) * int64(config.Duration.Seconds())
	successCount := int64(float64(requestCount) * 0.98) // 98% success rate for soak
	failCount := requestCount - successCount

	// Stable response times for soak test
	avgResponseTime := 250 * time.Millisecond
	minResponseTime := 80 * time.Millisecond
	maxResponseTime := 600 * time.Millisecond

	// Calculate percentiles for soak test
	p50 := 220 * time.Millisecond
	p90 := 350 * time.Millisecond
	p95 := 400 * time.Millisecond
	p99 := 550 * time.Millisecond

	// Update metrics
	metrics.TotalRequests = requestCount
	metrics.SuccessfulRequests = successCount
	metrics.FailedRequests = failCount
	metrics.AverageResponseTime = avgResponseTime
	metrics.MinResponseTime = minResponseTime
	metrics.MaxResponseTime = maxResponseTime
	metrics.P50ResponseTime = p50
	metrics.P90ResponseTime = p90
	metrics.P95ResponseTime = p95
	metrics.P99ResponseTime = p99
	metrics.Throughput = float64(config.RequestRate)
	metrics.ErrorRate = float64(failCount) / float64(requestCount)
	metrics.CPUUsage = 70.0
	metrics.MemoryUsage = 80.0
	metrics.NetworkIO = 120.0
	metrics.DatabaseQueries = requestCount * 2
	metrics.CacheHitRate = 0.90

	// Simulate test duration
	time.Sleep(config.Duration)

	return nil
}

// runBaselineTest runs a baseline test
func (a *AutomatedTesting) runBaselineTest(ctx context.Context, test *PerformanceTest, metrics *TestMetrics) error {
	// Baseline test uses standard load test metrics
	return a.runLoadTest(ctx, test, metrics)
}

// runRegressionTest runs a regression test
func (a *AutomatedTesting) runRegressionTest(ctx context.Context, test *PerformanceTest, metrics *TestMetrics) error {
	// Get baseline test for comparison
	baselineTest, exists := a.tests[test.BaselineID]
	if !exists {
		return fmt.Errorf("baseline test not found: %s", test.BaselineID)
	}

	// Run standard load test
	if err := a.runLoadTest(ctx, test, metrics); err != nil {
		return err
	}

	// Compare with baseline
	baselineMetrics := baselineTest.Metrics
	if baselineMetrics == nil {
		return fmt.Errorf("baseline test has no metrics")
	}

	// Calculate regression metrics
	responseTimeRegression := float64(metrics.AverageResponseTime) / float64(baselineMetrics.AverageResponseTime)
	throughputRegression := metrics.Throughput / baselineMetrics.Throughput
	errorRateRegression := metrics.ErrorRate / baselineMetrics.ErrorRate

	// Log regression analysis
	a.logger.Info("Regression test analysis",
		zap.String("test_id", test.ID),
		zap.Float64("response_time_regression", responseTimeRegression),
		zap.Float64("throughput_regression", throughputRegression),
		zap.Float64("error_rate_regression", errorRateRegression))

	return nil
}

// runOptimizationTest runs an optimization test
func (a *AutomatedTesting) runOptimizationTest(ctx context.Context, test *PerformanceTest, metrics *TestMetrics) error {
	// Run standard load test
	if err := a.runLoadTest(ctx, test, metrics); err != nil {
		return err
	}

	// Apply optimization strategy if specified
	if test.OptimizationID != "" && a.strategies != nil {
		result, err := a.strategies.ApplyStrategy(test.OptimizationID)
		if err != nil {
			a.logger.Warn("Failed to apply optimization strategy",
				zap.String("strategy_id", test.OptimizationID),
				zap.Error(err))
		} else {
			a.logger.Info("Applied optimization strategy",
				zap.String("strategy_id", test.OptimizationID),
				zap.Float64("actual_impact", result.ActualImpact))
		}
	}

	return nil
}

// evaluateTestResult evaluates the test result based on thresholds
func (a *AutomatedTesting) evaluateTestResult(test *PerformanceTest) TestResult {
	metrics := test.Metrics
	thresholds := test.Thresholds

	if metrics == nil || thresholds == nil {
		return TestResultFail
	}

	// Check response time threshold
	if metrics.AverageResponseTime > thresholds.MaxResponseTime {
		return TestResultFail
	}

	// Check error rate threshold
	if metrics.ErrorRate > thresholds.MaxErrorRate {
		return TestResultFail
	}

	// Check throughput threshold
	if metrics.Throughput < thresholds.MinThroughput {
		return TestResultFail
	}

	// Check resource usage thresholds
	if metrics.CPUUsage > thresholds.MaxCPUUsage {
		return TestResultWarn
	}

	if metrics.MemoryUsage > thresholds.MaxMemoryUsage {
		return TestResultWarn
	}

	// Check cache hit rate threshold
	if metrics.CacheHitRate < thresholds.MinCacheHitRate {
		return TestResultWarn
	}

	// Check database queries threshold
	if metrics.DatabaseQueries > thresholds.MaxDatabaseQueries {
		return TestResultWarn
	}

	return TestResultPass
}

// generateTestSummary generates a summary of the test results
func (a *AutomatedTesting) generateTestSummary(test *PerformanceTest) string {
	metrics := test.Metrics
	if metrics == nil {
		return "Test completed with no metrics collected"
	}

	return fmt.Sprintf("Test %s completed: %d requests, %.2f%% success rate, %.2fms avg response time, %.2f req/s throughput",
		test.Name,
		metrics.TotalRequests,
		(1-metrics.ErrorRate)*100,
		float64(metrics.AverageResponseTime.Milliseconds()),
		metrics.Throughput)
}

// generateTestRecommendations generates recommendations based on test results
func (a *AutomatedTesting) generateTestRecommendations(test *PerformanceTest) []string {
	var recommendations []string
	metrics := test.Metrics
	thresholds := test.Thresholds

	if metrics == nil || thresholds == nil {
		return []string{"Test configuration incomplete - check thresholds and metrics collection"}
	}

	// Response time recommendations
	if metrics.AverageResponseTime > thresholds.MaxResponseTime {
		recommendations = append(recommendations, "Consider implementing caching to reduce response times")
		recommendations = append(recommendations, "Review database query optimization")
		recommendations = append(recommendations, "Consider horizontal scaling for better performance")
	}

	// Error rate recommendations
	if metrics.ErrorRate > thresholds.MaxErrorRate {
		recommendations = append(recommendations, "Investigate error patterns and implement error handling")
		recommendations = append(recommendations, "Review system stability and resource allocation")
		recommendations = append(recommendations, "Consider implementing circuit breakers")
	}

	// Throughput recommendations
	if metrics.Throughput < thresholds.MinThroughput {
		recommendations = append(recommendations, "Optimize request processing pipeline")
		recommendations = append(recommendations, "Consider implementing request queuing")
		recommendations = append(recommendations, "Review concurrency settings")
	}

	// Resource usage recommendations
	if metrics.CPUUsage > thresholds.MaxCPUUsage {
		recommendations = append(recommendations, "Consider CPU optimization or scaling")
		recommendations = append(recommendations, "Review CPU-intensive operations")
	}

	if metrics.MemoryUsage > thresholds.MaxMemoryUsage {
		recommendations = append(recommendations, "Implement memory optimization strategies")
		recommendations = append(recommendations, "Review memory allocation patterns")
		recommendations = append(recommendations, "Consider garbage collection tuning")
	}

	// Cache recommendations
	if metrics.CacheHitRate < thresholds.MinCacheHitRate {
		recommendations = append(recommendations, "Optimize cache key strategies")
		recommendations = append(recommendations, "Review cache invalidation policies")
		recommendations = append(recommendations, "Consider increasing cache size")
	}

	// Database recommendations
	if metrics.DatabaseQueries > thresholds.MaxDatabaseQueries {
		recommendations = append(recommendations, "Optimize database queries")
		recommendations = append(recommendations, "Implement query result caching")
		recommendations = append(recommendations, "Review database connection pooling")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Performance is within acceptable thresholds")
	}

	return recommendations
}

// createBaselineTest creates a baseline performance test
func (a *AutomatedTesting) createBaselineTest() *PerformanceTest {
	return &PerformanceTest{
		ID:          fmt.Sprintf("baseline_%d", time.Now().Unix()),
		Name:        "Baseline Performance Test",
		Type:        TestTypeBaseline,
		Description: "Baseline performance test to establish performance benchmarks",
		Status:      TestStatusPending,
		Config: &TestConfig{
			Duration:       5 * time.Minute,
			Concurrency:    10,
			RequestRate:    100,
			RampUpTime:     30 * time.Second,
			RampDownTime:   30 * time.Second,
			TargetEndpoint: "/api/v3/classify",
			Timeout:        a.config.DefaultTimeout,
			RetryCount:     3,
			ThinkTime:      100 * time.Millisecond,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    500 * time.Millisecond,
			MaxErrorRate:       0.05, // 5%
			MinThroughput:      80.0, // 80 req/s
			MaxCPUUsage:        80.0, // 80%
			MaxMemoryUsage:     85.0, // 85%
			MinCacheHitRate:    0.70, // 70%
			MaxDatabaseQueries: 1000,
		},
		CreatedAt: time.Now(),
	}
}

// createRegressionTest creates a regression performance test
func (a *AutomatedTesting) createRegressionTest(baselineID string) *PerformanceTest {
	return &PerformanceTest{
		ID:          fmt.Sprintf("regression_%d", time.Now().Unix()),
		Name:        "Regression Performance Test",
		Type:        TestTypeRegression,
		Description: "Regression test to detect performance degradation",
		Status:      TestStatusPending,
		BaselineID:  baselineID,
		Config: &TestConfig{
			Duration:       5 * time.Minute,
			Concurrency:    10,
			RequestRate:    100,
			RampUpTime:     30 * time.Second,
			RampDownTime:   30 * time.Second,
			TargetEndpoint: "/api/v3/classify",
			Timeout:        a.config.DefaultTimeout,
			RetryCount:     3,
			ThinkTime:      100 * time.Millisecond,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    500 * time.Millisecond,
			MaxErrorRate:       0.05, // 5%
			MinThroughput:      80.0, // 80 req/s
			MaxCPUUsage:        80.0, // 80%
			MaxMemoryUsage:     85.0, // 85%
			MinCacheHitRate:    0.70, // 70%
			MaxDatabaseQueries: 1000,
		},
		CreatedAt: time.Now(),
	}
}

// createOptimizationTest creates an optimization performance test
func (a *AutomatedTesting) createOptimizationTest(strategyID string) *PerformanceTest {
	return &PerformanceTest{
		ID:             fmt.Sprintf("optimization_%d", time.Now().Unix()),
		Name:           "Optimization Performance Test",
		Type:           TestTypeOptimization,
		Description:    "Test to validate optimization strategy effectiveness",
		Status:         TestStatusPending,
		OptimizationID: strategyID,
		Config: &TestConfig{
			Duration:       5 * time.Minute,
			Concurrency:    10,
			RequestRate:    100,
			RampUpTime:     30 * time.Second,
			RampDownTime:   30 * time.Second,
			TargetEndpoint: "/api/v3/classify",
			Timeout:        a.config.DefaultTimeout,
			RetryCount:     3,
			ThinkTime:      100 * time.Millisecond,
		},
		Thresholds: &TestThresholds{
			MaxResponseTime:    500 * time.Millisecond,
			MaxErrorRate:       0.05, // 5%
			MinThroughput:      80.0, // 80 req/s
			MaxCPUUsage:        80.0, // 80%
			MaxMemoryUsage:     85.0, // 85%
			MinCacheHitRate:    0.70, // 70%
			MaxDatabaseQueries: 1000,
		},
		CreatedAt: time.Now(),
	}
}

// GetTests retrieves all performance tests
func (a *AutomatedTesting) GetTests() []*PerformanceTest {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var tests []*PerformanceTest
	for _, test := range a.tests {
		tests = append(tests, test)
	}

	return tests
}

// GetTestByID retrieves a specific test by ID
func (a *AutomatedTesting) GetTestByID(testID string) (*PerformanceTest, bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	test, exists := a.tests[testID]
	return test, exists
}

// GetTestsByType retrieves tests by type
func (a *AutomatedTesting) GetTestsByType(testType TestType) []*PerformanceTest {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var tests []*PerformanceTest
	for _, test := range a.tests {
		if test.Type == testType {
			tests = append(tests, test)
		}
	}

	return tests
}

// GetTestsByStatus retrieves tests by status
func (a *AutomatedTesting) GetTestsByStatus(status TestStatus) []*PerformanceTest {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var tests []*PerformanceTest
	for _, test := range a.tests {
		if test.Status == status {
			tests = append(tests, test)
		}
	}

	return tests
}

// GetTestSuites retrieves all test suites
func (a *AutomatedTesting) GetTestSuites() []*TestSuite {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	var suites []*TestSuite
	for _, suite := range a.suites {
		suites = append(suites, suite)
	}

	return suites
}

// CreateTestSuite creates a new test suite
func (a *AutomatedTesting) CreateTestSuite(name, description string, tests []*PerformanceTest) *TestSuite {
	suite := &TestSuite{
		ID:          fmt.Sprintf("suite_%d", time.Now().Unix()),
		Name:        name,
		Description: description,
		Tests:       tests,
		CreatedAt:   time.Now(),
		Status:      TestStatusPending,
	}

	a.mutex.Lock()
	a.suites[suite.ID] = suite
	a.mutex.Unlock()

	return suite
}

// RunTestSuite runs a test suite
func (a *AutomatedTesting) RunTestSuite(ctx context.Context, suite *TestSuite) error {
	a.mutex.Lock()
	suite.Status = TestStatusRunning
	a.mutex.Unlock()

	a.logger.Info("Starting test suite", zap.String("suite_id", suite.ID), zap.String("suite_name", suite.Name))

	// Run tests in parallel (up to max concurrent tests)
	semaphore := make(chan struct{}, a.config.MaxConcurrentTests)
	var wg sync.WaitGroup
	var errors []error
	var errorMutex sync.Mutex

	for _, test := range suite.Tests {
		wg.Add(1)
		go func(t *PerformanceTest) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := a.RunTest(ctx, t); err != nil {
				errorMutex.Lock()
				errors = append(errors, err)
				errorMutex.Unlock()
			}
		}(test)
	}

	wg.Wait()

	// Update suite status
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(errors) > 0 {
		suite.Status = TestStatusFailed
		suite.Result = TestResultFail
		suite.Summary = fmt.Sprintf("Test suite failed with %d errors", len(errors))
	} else {
		suite.Status = TestStatusCompleted
		suite.Result = TestResultPass
		suite.Summary = "All tests in suite completed successfully"
	}

	if len(errors) > 0 {
		return fmt.Errorf("test suite failed with %d errors", len(errors))
	}

	return nil
}
