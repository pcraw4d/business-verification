package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewPerformanceTest(t *testing.T) {
	function := func(ctx *PerformanceContext) error { return nil }
	test := NewPerformanceTest("test", function)

	if test.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", test.Name)
	}

	if test.Function == nil {
		t.Error("Expected function to be set")
	}

	if test.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", test.Timeout)
	}

	if test.Parallel {
		t.Error("Expected parallel to be false by default")
	}

	if test.Skipped {
		t.Error("Expected skipped to be false by default")
	}
}

func TestPerformanceTest_AddTag(t *testing.T) {
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	test.AddTag("fast").AddTag("integration")

	if len(test.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(test.Tags))
	}

	if test.Tags[0] != "fast" {
		t.Errorf("Expected first tag 'fast', got '%s'", test.Tags[0])
	}

	if test.Tags[1] != "integration" {
		t.Errorf("Expected second tag 'integration', got '%s'", test.Tags[1])
	}
}

func TestPerformanceTest_SetConfig(t *testing.T) {
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	config := PerformanceTestConfig{
		Iterations:  500,
		Concurrency: 10,
		Duration:    2 * time.Minute,
	}
	test.SetConfig(config)

	if test.Config.Iterations != 500 {
		t.Errorf("Expected iterations 500, got %d", test.Config.Iterations)
	}

	if test.Config.Concurrency != 10 {
		t.Errorf("Expected concurrency 10, got %d", test.Config.Concurrency)
	}

	if test.Config.Duration != 2*time.Minute {
		t.Errorf("Expected duration 2m, got %v", test.Config.Duration)
	}
}

func TestPerformanceTest_SetTimeout(t *testing.T) {
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	test.SetTimeout(60 * time.Second)

	if test.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", test.Timeout)
	}
}

func TestPerformanceTest_SetParallel(t *testing.T) {
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	test.SetParallel(true)

	if !test.Parallel {
		t.Error("Expected parallel to be true")
	}
}

func TestPerformanceTest_Skip(t *testing.T) {
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	test.Skip()

	if !test.Skipped {
		t.Error("Expected skipped to be true")
	}
}

func TestPerformanceTest_AddComponent(t *testing.T) {
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	test.AddComponent("database").AddComponent("api")

	if len(test.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(test.Components))
	}

	if test.Components[0] != "database" {
		t.Errorf("Expected first component 'database', got '%s'", test.Components[0])
	}

	if test.Components[1] != "api" {
		t.Errorf("Expected second component 'api', got '%s'", test.Components[1])
	}
}

func TestDefaultPerformanceTestConfig(t *testing.T) {
	config := DefaultPerformanceTestConfig()

	if config.Iterations != 100 {
		t.Errorf("Expected iterations 100, got %d", config.Iterations)
	}

	if config.Concurrency != 1 {
		t.Errorf("Expected concurrency 1, got %d", config.Concurrency)
	}

	if config.Duration != 60*time.Second {
		t.Errorf("Expected duration 60s, got %v", config.Duration)
	}

	if config.WarmupTime != 5*time.Second {
		t.Errorf("Expected warmup time 5s, got %v", config.WarmupTime)
	}

	if config.CooldownTime != 5*time.Second {
		t.Errorf("Expected cooldown time 5s, got %v", config.CooldownTime)
	}

	if config.RequestRate != 100 {
		t.Errorf("Expected request rate 100, got %d", config.RequestRate)
	}

	if config.RampUpTime != 10*time.Second {
		t.Errorf("Expected ramp up time 10s, got %v", config.RampUpTime)
	}

	if config.RampDownTime != 10*time.Second {
		t.Errorf("Expected ramp down time 10s, got %v", config.RampDownTime)
	}

	if len(config.Metrics) != 5 {
		t.Errorf("Expected 5 metrics, got %d", len(config.Metrics))
	}

	if config.Baseline != nil {
		t.Error("Expected baseline to be nil by default")
	}

	if config.Regression {
		t.Error("Expected regression to be false by default")
	}
}

func TestDefaultPerformanceThresholds(t *testing.T) {
	thresholds := DefaultPerformanceThresholds()

	if thresholds.MaxResponseTime != 1*time.Second {
		t.Errorf("Expected max response time 1s, got %v", thresholds.MaxResponseTime)
	}

	if thresholds.MinThroughput != 100.0 {
		t.Errorf("Expected min throughput 100.0, got %f", thresholds.MinThroughput)
	}

	if thresholds.MaxErrorRate != 0.01 {
		t.Errorf("Expected max error rate 0.01, got %f", thresholds.MaxErrorRate)
	}

	if thresholds.MaxMemoryUsage != 100*1024*1024 {
		t.Errorf("Expected max memory usage 100MB, got %d", thresholds.MaxMemoryUsage)
	}

	if thresholds.MaxCPUUsage != 80.0 {
		t.Errorf("Expected max CPU usage 80.0, got %f", thresholds.MaxCPUUsage)
	}

	if thresholds.MaxLatencyP95 != 500*time.Millisecond {
		t.Errorf("Expected max latency P95 500ms, got %v", thresholds.MaxLatencyP95)
	}

	if thresholds.MaxLatencyP99 != 1*time.Second {
		t.Errorf("Expected max latency P99 1s, got %v", thresholds.MaxLatencyP99)
	}
}

func TestNewPerformanceTestSuite(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")

	if suite.Name != "test-suite" {
		t.Errorf("Expected name 'test-suite', got '%s'", suite.Name)
	}

	if len(suite.Tests) != 0 {
		t.Errorf("Expected 0 tests, got %d", len(suite.Tests))
	}

	if suite.Setup != nil {
		t.Error("Expected setup to be nil by default")
	}

	if suite.Teardown != nil {
		t.Error("Expected teardown to be nil by default")
	}

	if suite.BeforeEach != nil {
		t.Error("Expected beforeEach to be nil by default")
	}

	if suite.AfterEach != nil {
		t.Error("Expected afterEach to be nil by default")
	}

	if suite.Parallel {
		t.Error("Expected parallel to be false by default")
	}

	if suite.Timeout != 5*time.Minute {
		t.Errorf("Expected timeout 5m, got %v", suite.Timeout)
	}

	if len(suite.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(suite.Tags))
	}

	if len(suite.Components) != 0 {
		t.Errorf("Expected 0 components, got %d", len(suite.Components))
	}
}

func TestPerformanceTestSuite_AddTest(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	suite.AddTest(test)

	if len(suite.Tests) != 1 {
		t.Errorf("Expected 1 test, got %d", len(suite.Tests))
	}

	if suite.Tests[0] != test {
		t.Error("Expected test to be added to suite")
	}
}

func TestPerformanceTestSuite_CreateTest(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")
	test := suite.CreateTest("test", func(ctx *PerformanceContext) error { return nil })

	if test.Name != "test" {
		t.Errorf("Expected test name 'test', got '%s'", test.Name)
	}

	if len(suite.Tests) != 1 {
		t.Errorf("Expected 1 test, got %d", len(suite.Tests))
	}

	if suite.Tests[0] != test {
		t.Error("Expected test to be added to suite")
	}
}

func TestPerformanceTestSuite_SetSetup(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")
	setup := func(ctx *PerformanceContext) error { return nil }
	suite.SetSetup(setup)

	if suite.Setup == nil {
		t.Error("Expected setup to be set")
	}
}

func TestPerformanceTestSuite_SetTeardown(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")
	teardown := func(ctx *PerformanceContext) error { return nil }
	suite.SetTeardown(teardown)

	if suite.Teardown == nil {
		t.Error("Expected teardown to be set")
	}
}

func TestPerformanceTestSuite_SetBeforeEach(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")
	beforeEach := func(ctx *PerformanceContext) error { return nil }
	suite.SetBeforeEach(beforeEach)

	if suite.BeforeEach == nil {
		t.Error("Expected beforeEach to be set")
	}
}

func TestPerformanceTestSuite_SetAfterEach(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")
	afterEach := func(ctx *PerformanceContext) error { return nil }
	suite.SetAfterEach(afterEach)

	if suite.AfterEach == nil {
		t.Error("Expected afterEach to be set")
	}
}

func TestPerformanceTestSuite_SetParallel(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")
	suite.SetParallel(true)

	if !suite.Parallel {
		t.Error("Expected parallel to be true")
	}
}

func TestPerformanceTestSuite_SetTimeout(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")
	suite.SetTimeout(10 * time.Minute)

	if suite.Timeout != 10*time.Minute {
		t.Errorf("Expected timeout 10m, got %v", suite.Timeout)
	}
}

func TestPerformanceTestSuite_AddTag(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")
	suite.AddTag("fast").AddTag("integration")

	if len(suite.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(suite.Tags))
	}

	if suite.Tags[0] != "fast" {
		t.Errorf("Expected first tag 'fast', got '%s'", suite.Tags[0])
	}

	if suite.Tags[1] != "integration" {
		t.Errorf("Expected second tag 'integration', got '%s'", suite.Tags[1])
	}
}

func TestPerformanceTestSuite_AddComponent(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")
	suite.AddComponent("database").AddComponent("api")

	if len(suite.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(suite.Components))
	}

	if suite.Components[0] != "database" {
		t.Errorf("Expected first component 'database', got '%s'", suite.Components[0])
	}

	if suite.Components[1] != "api" {
		t.Errorf("Expected second component 'api', got '%s'", suite.Components[1])
	}
}

func TestPerformanceTestSuite_SetConfig(t *testing.T) {
	suite := NewPerformanceTestSuite("test-suite")
	config := PerformanceTestConfig{
		Iterations:  500,
		Concurrency: 10,
		Duration:    2 * time.Minute,
	}
	suite.SetConfig(config)

	if suite.Config.Iterations != 500 {
		t.Errorf("Expected iterations 500, got %d", suite.Config.Iterations)
	}

	if suite.Config.Concurrency != 10 {
		t.Errorf("Expected concurrency 10, got %d", suite.Config.Concurrency)
	}

	if suite.Config.Duration != 2*time.Minute {
		t.Errorf("Expected duration 2m, got %v", suite.Config.Duration)
	}
}

func TestNewPerformanceContext(t *testing.T) {
	logger := zap.NewNop()
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	ctx := context.Background()
	perfCtx := NewPerformanceContext(ctx, test, logger)

	if perfCtx.T != test {
		t.Error("Expected test to be set")
	}

	if perfCtx.Logger != logger {
		t.Error("Expected logger to be set")
	}

	if perfCtx.Metrics == nil {
		t.Error("Expected metrics to be initialized")
	}

	if len(perfCtx.CleanupFuncs) != 0 {
		t.Errorf("Expected 0 cleanup functions, got %d", len(perfCtx.CleanupFuncs))
	}

	if perfCtx.cancel == nil {
		t.Error("Expected cancel function to be set")
	}
}

func TestNewPerformanceMetrics(t *testing.T) {
	metrics := NewPerformanceMetrics()

	if len(metrics.ResponseTimes) != 0 {
		t.Errorf("Expected 0 response times, got %d", len(metrics.ResponseTimes))
	}

	if metrics.Throughput != 0.0 {
		t.Errorf("Expected throughput 0.0, got %f", metrics.Throughput)
	}

	if metrics.ErrorCount != 0 {
		t.Errorf("Expected error count 0, got %d", metrics.ErrorCount)
	}

	if metrics.TotalRequests != 0 {
		t.Errorf("Expected total requests 0, got %d", metrics.TotalRequests)
	}

	if len(metrics.MemoryUsage) != 0 {
		t.Errorf("Expected 0 memory usage entries, got %d", len(metrics.MemoryUsage))
	}

	if len(metrics.CPUUsage) != 0 {
		t.Errorf("Expected 0 CPU usage entries, got %d", len(metrics.CPUUsage))
	}

	if metrics.StartTime.IsZero() {
		t.Error("Expected start time to be set")
	}

	if !metrics.EndTime.IsZero() {
		t.Error("Expected end time to be zero")
	}
}

func TestPerformanceContext_Cleanup(t *testing.T) {
	logger := zap.NewNop()
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	ctx := context.Background()
	perfCtx := NewPerformanceContext(ctx, test, logger)

	cleanupCalled := false
	perfCtx.AddCleanup(func() {
		cleanupCalled = true
	})

	perfCtx.Cleanup()

	if !cleanupCalled {
		t.Error("Expected cleanup function to be called")
	}

	if perfCtx.EndTime.IsZero() {
		t.Error("Expected end time to be set")
	}

	if perfCtx.Metrics.EndTime.IsZero() {
		t.Error("Expected metrics end time to be set")
	}
}

func TestPerformanceContext_AddCleanup(t *testing.T) {
	logger := zap.NewNop()
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	ctx := context.Background()
	perfCtx := NewPerformanceContext(ctx, test, logger)

	perfCtx.AddCleanup(func() {})
	perfCtx.AddCleanup(func() {})

	if len(perfCtx.CleanupFuncs) != 2 {
		t.Errorf("Expected 2 cleanup functions, got %d", len(perfCtx.CleanupFuncs))
	}
}

func TestPerformanceContext_Log(t *testing.T) {
	logger := zap.NewNop()
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	ctx := context.Background()
	perfCtx := NewPerformanceContext(ctx, test, logger)

	// Should not panic
	perfCtx.Log("test message")
	perfCtx.Logf("test message %s", "formatted")
}

func TestPerformanceMetrics_RecordResponseTime(t *testing.T) {
	metrics := NewPerformanceMetrics()

	metrics.RecordResponseTime(100 * time.Millisecond)
	metrics.RecordResponseTime(200 * time.Millisecond)

	if len(metrics.ResponseTimes) != 2 {
		t.Errorf("Expected 2 response times, got %d", len(metrics.ResponseTimes))
	}

	if metrics.ResponseTimes[0] != 100*time.Millisecond {
		t.Errorf("Expected first response time 100ms, got %v", metrics.ResponseTimes[0])
	}

	if metrics.ResponseTimes[1] != 200*time.Millisecond {
		t.Errorf("Expected second response time 200ms, got %v", metrics.ResponseTimes[1])
	}

	if metrics.TotalRequests != 2 {
		t.Errorf("Expected total requests 2, got %d", metrics.TotalRequests)
	}
}

func TestPerformanceMetrics_RecordError(t *testing.T) {
	metrics := NewPerformanceMetrics()

	metrics.RecordError()
	metrics.RecordError()

	if metrics.ErrorCount != 2 {
		t.Errorf("Expected error count 2, got %d", metrics.ErrorCount)
	}

	if metrics.TotalRequests != 2 {
		t.Errorf("Expected total requests 2, got %d", metrics.TotalRequests)
	}
}

func TestPerformanceMetrics_RecordMemoryUsage(t *testing.T) {
	metrics := NewPerformanceMetrics()

	metrics.RecordMemoryUsage(1024 * 1024)     // 1MB
	metrics.RecordMemoryUsage(2 * 1024 * 1024) // 2MB

	if len(metrics.MemoryUsage) != 2 {
		t.Errorf("Expected 2 memory usage entries, got %d", len(metrics.MemoryUsage))
	}

	if metrics.MemoryUsage[0] != 1024*1024 {
		t.Errorf("Expected first memory usage 1MB, got %d", metrics.MemoryUsage[0])
	}

	if metrics.MemoryUsage[1] != 2*1024*1024 {
		t.Errorf("Expected second memory usage 2MB, got %d", metrics.MemoryUsage[1])
	}
}

func TestPerformanceMetrics_RecordCPUUsage(t *testing.T) {
	metrics := NewPerformanceMetrics()

	metrics.RecordCPUUsage(50.0)
	metrics.RecordCPUUsage(75.0)

	if len(metrics.CPUUsage) != 2 {
		t.Errorf("Expected 2 CPU usage entries, got %d", len(metrics.CPUUsage))
	}

	if metrics.CPUUsage[0] != 50.0 {
		t.Errorf("Expected first CPU usage 50.0, got %f", metrics.CPUUsage[0])
	}

	if metrics.CPUUsage[1] != 75.0 {
		t.Errorf("Expected second CPU usage 75.0, got %f", metrics.CPUUsage[1])
	}
}

func TestPerformanceMetrics_CalculateThroughput(t *testing.T) {
	metrics := NewPerformanceMetrics()
	metrics.StartTime = time.Now().Add(-1 * time.Second) // 1 second ago
	metrics.EndTime = time.Now()

	metrics.RecordResponseTime(100 * time.Millisecond)
	metrics.RecordResponseTime(200 * time.Millisecond)

	throughput := metrics.CalculateThroughput()

	if throughput <= 0 {
		t.Errorf("Expected positive throughput, got %f", throughput)
	}
}

func TestPerformanceMetrics_CalculateErrorRate(t *testing.T) {
	metrics := NewPerformanceMetrics()

	// Record 2 successful requests and 1 error
	metrics.RecordResponseTime(100 * time.Millisecond)
	metrics.RecordResponseTime(200 * time.Millisecond)
	metrics.RecordError()

	errorRate := metrics.CalculateErrorRate()

	expectedErrorRate := 1.0 / 3.0 // 1 error out of 3 total requests
	if errorRate != expectedErrorRate {
		t.Errorf("Expected error rate %f, got %f", expectedErrorRate, errorRate)
	}
}

func TestPerformanceMetrics_CalculateAverageResponseTime(t *testing.T) {
	metrics := NewPerformanceMetrics()

	metrics.RecordResponseTime(100 * time.Millisecond)
	metrics.RecordResponseTime(200 * time.Millisecond)
	metrics.RecordResponseTime(300 * time.Millisecond)

	avgResponseTime := metrics.CalculateAverageResponseTime()
	expectedAvg := 200 * time.Millisecond // (100 + 200 + 300) / 3

	if avgResponseTime != expectedAvg {
		t.Errorf("Expected average response time %v, got %v", expectedAvg, avgResponseTime)
	}
}

func TestPerformanceMetrics_CalculatePercentileResponseTime(t *testing.T) {
	metrics := NewPerformanceMetrics()

	// Record response times: 100, 200, 300, 400, 500ms
	for i := 1; i <= 5; i++ {
		metrics.RecordResponseTime(time.Duration(i*100) * time.Millisecond)
	}

	p95 := metrics.CalculatePercentileResponseTime(95)
	expectedP95 := 500 * time.Millisecond // 5th value (index 4)

	if p95 != expectedP95 {
		t.Errorf("Expected P95 response time %v, got %v", expectedP95, p95)
	}

	p50 := metrics.CalculatePercentileResponseTime(50)
	expectedP50 := 300 * time.Millisecond // 3rd value (index 2)

	if p50 != expectedP50 {
		t.Errorf("Expected P50 response time %v, got %v", expectedP50, p50)
	}
}

func TestPerformanceMetrics_CalculateMaxMemoryUsage(t *testing.T) {
	metrics := NewPerformanceMetrics()

	metrics.RecordMemoryUsage(1024 * 1024)     // 1MB
	metrics.RecordMemoryUsage(2 * 1024 * 1024) // 2MB
	metrics.RecordMemoryUsage(512 * 1024)      // 512KB

	maxMemory := metrics.CalculateMaxMemoryUsage()
	expectedMax := int64(2 * 1024 * 1024) // 2MB

	if maxMemory != expectedMax {
		t.Errorf("Expected max memory usage %d, got %d", expectedMax, maxMemory)
	}
}

func TestPerformanceMetrics_CalculateAverageCPUUsage(t *testing.T) {
	metrics := NewPerformanceMetrics()

	metrics.RecordCPUUsage(50.0)
	metrics.RecordCPUUsage(75.0)
	metrics.RecordCPUUsage(25.0)

	avgCPU := metrics.CalculateAverageCPUUsage()
	expectedAvg := 50.0 // (50 + 75 + 25) / 3

	if avgCPU != expectedAvg {
		t.Errorf("Expected average CPU usage %f, got %f", expectedAvg, avgCPU)
	}
}

func TestNewPerformanceTestRunner(t *testing.T) {
	logger := zap.NewNop()
	runner := NewPerformanceTestRunner(logger)

	if len(runner.Suites) != 0 {
		t.Errorf("Expected 0 suites, got %d", len(runner.Suites))
	}

	if runner.Logger != logger {
		t.Error("Expected logger to be set")
	}

	if len(runner.Results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(runner.Results))
	}
}

func TestPerformanceTestRunner_AddSuite(t *testing.T) {
	logger := zap.NewNop()
	runner := NewPerformanceTestRunner(logger)
	suite := NewPerformanceTestSuite("test-suite")

	runner.AddSuite(suite)

	if len(runner.Suites) != 1 {
		t.Errorf("Expected 1 suite, got %d", len(runner.Suites))
	}

	if runner.Suites[0] != suite {
		t.Error("Expected suite to be added to runner")
	}
}

func TestPerformanceTestRunner_GenerateSummary(t *testing.T) {
	logger := zap.NewNop()
	runner := NewPerformanceTestRunner(logger)

	// Add some test results
	test1 := NewPerformanceTest("test1", func(ctx *PerformanceContext) error { return nil })
	test2 := NewPerformanceTest("test2", func(ctx *PerformanceContext) error { return nil })

	result1 := &PerformanceTestResult{
		Test:   test1,
		Status: PerfStatusPassed,
	}
	result2 := &PerformanceTestResult{
		Test:   test2,
		Status: PerfStatusFailed,
	}

	runner.Results = []*PerformanceTestResult{result1, result2}

	summary := runner.GenerateSummary()

	if summary.TotalTests != 2 {
		t.Errorf("Expected total tests 2, got %d", summary.TotalTests)
	}

	if summary.PassedTests != 1 {
		t.Errorf("Expected passed tests 1, got %d", summary.PassedTests)
	}

	if summary.FailedTests != 1 {
		t.Errorf("Expected failed tests 1, got %d", summary.FailedTests)
	}

	if summary.SkippedTests != 0 {
		t.Errorf("Expected skipped tests 0, got %d", summary.SkippedTests)
	}

	if summary.ErrorTests != 0 {
		t.Errorf("Expected error tests 0, got %d", summary.ErrorTests)
	}

	if summary.RegressionTests != 0 {
		t.Errorf("Expected regression tests 0, got %d", summary.RegressionTests)
	}
}

func TestPerformanceTestRunner_EvaluateTestResult(t *testing.T) {
	logger := zap.NewNop()
	runner := NewPerformanceTestRunner(logger)

	// Test with error
	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	result := &PerformanceTestResult{
		Test:       test,
		Error:      fmt.Errorf("test error"),
		Metrics:    NewPerformanceMetrics(),
		Thresholds: DefaultPerformanceThresholds(),
	}

	status, passed, regression := runner.evaluateTestResult(result)

	if status != PerfStatusError {
		t.Errorf("Expected status PerfStatusError, got %s", status)
	}

	if passed {
		t.Error("Expected passed to be false")
	}

	if regression {
		t.Error("Expected regression to be false")
	}

	// Test with passing metrics
	result.Error = nil
	result.Metrics.RecordResponseTime(50 * time.Millisecond) // Below threshold
	result.Metrics.RecordResponseTime(50 * time.Millisecond) // Below threshold

	status, passed, regression = runner.evaluateTestResult(result)

	if status != PerfStatusPassed {
		t.Errorf("Expected status PerfStatusPassed, got %s", status)
	}

	if !passed {
		t.Error("Expected passed to be true")
	}

	if regression {
		t.Error("Expected regression to be false")
	}
}

func TestPerformanceTestRunner_CheckRegression(t *testing.T) {
	logger := zap.NewNop()
	runner := NewPerformanceTestRunner(logger)

	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	metrics := NewPerformanceMetrics()
	metrics.RecordResponseTime(200 * time.Millisecond) // 100% increase from baseline

	baseline := &PerformanceBaseline{
		ResponseTime: 100 * time.Millisecond,
		Throughput:   100.0,
		ErrorRate:    0.01,
	}

	result := &PerformanceTestResult{
		Test:     test,
		Metrics:  metrics,
		Baseline: baseline,
	}

	regression := runner.checkRegression(result)

	if !regression {
		t.Error("Expected regression to be true")
	}
}

func TestPerformanceTestRunner_GenerateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	runner := NewPerformanceTestRunner(logger)

	test := NewPerformanceTest("test", func(ctx *PerformanceContext) error { return nil })
	metrics := NewPerformanceMetrics()
	metrics.RecordResponseTime(2 * time.Second) // Above threshold
	metrics.RecordResponseTime(2 * time.Second) // Above threshold

	thresholds := DefaultPerformanceThresholds()

	result := &PerformanceTestResult{
		Test:       test,
		Metrics:    metrics,
		Thresholds: thresholds,
	}

	recommendations := runner.generateRecommendations(result)

	if len(recommendations) == 0 {
		t.Error("Expected recommendations to be generated")
	}

	// Check that response time recommendation is included
	foundResponseTimeRec := false
	for _, rec := range recommendations {
		if contains(rec, "response time") {
			foundResponseTimeRec = true
			break
		}
	}

	if !foundResponseTimeRec {
		t.Error("Expected response time recommendation to be included")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
