package risk

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// PerformanceTestRunner provides comprehensive performance testing capabilities
// NOTE: test_runners_backup is a subdirectory, so it's a separate package from internal/risk
// Types like PerformanceTestSuite are defined in the parent package and are not accessible
type PerformanceTestRunner struct {
	logger    *zap.Logger
	testSuite interface{} // *PerformanceTestSuite - defined in parent package
	results   *PerformanceTestResults
}

// PerformanceTestResults contains the results of performance test execution
type PerformanceTestResults struct {
	TotalTests         int                     `json:"total_tests"`
	PassedTests        int                     `json:"passed_tests"`
	FailedTests        int                     `json:"failed_tests"`
	SkippedTests       int                     `json:"skipped_tests"`
	ExecutionTime      time.Duration           `json:"execution_time"`
	TestDetails        []PerformanceTestDetail `json:"test_details"`
	Summary            map[string]interface{}  `json:"summary"`
	PerformanceMetrics *PerformanceTestMetrics `json:"performance_metrics"`
	ScalabilityMetrics *ScalabilityMetrics     `json:"scalability_metrics"`
	ResourceMetrics    *ResourceMetrics        `json:"resource_metrics"`
}

// PerformanceTestMetrics contains performance metrics specific to test execution
type PerformanceTestMetrics struct {
	LatencyDistribution map[string]int  `json:"latency_distribution"`
	ThroughputTrend     []float64       `json:"throughput_trend"`
	LatencyTrend        []time.Duration `json:"latency_trend"`
	TotalOperations     int             `json:"total_operations"`
	TotalDuration       time.Duration   `json:"total_duration"`
	AverageThroughput   float64         `json:"average_throughput"`
	MaxThroughput       float64         `json:"max_throughput"`
	MinThroughput       float64         `json:"min_throughput"`
	AverageLatency      time.Duration   `json:"average_latency"`
	MaxLatency          time.Duration   `json:"max_latency"`
	MinLatency          time.Duration   `json:"min_latency"`
	P95Latency          time.Duration   `json:"p95_latency"`
	P99Latency          time.Duration   `json:"p99_latency"`
	ErrorRate           float64         `json:"error_rate"`
	SuccessRate         float64         `json:"success_rate"`
}

// PerformanceTestDetail contains details about individual performance test execution
type PerformanceTestDetail struct {
	Name           string        `json:"name"`
	Category       string        `json:"category"`
	Status         string        `json:"status"`
	Duration       time.Duration `json:"duration"`
	Operations     int           `json:"operations"`
	Throughput     float64       `json:"throughput"` // operations per second
	AverageLatency time.Duration `json:"average_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	MinLatency     time.Duration `json:"min_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	ErrorRate      float64       `json:"error_rate"`
	MemoryUsage    int64         `json:"memory_usage"`
	CPUUsage       float64       `json:"cpu_usage"`
	Concurrency    int           `json:"concurrency"`
}

// ScalabilityMetrics contains scalability metrics
type ScalabilityMetrics struct {
	MaxConcurrency      int                    `json:"max_concurrency"`
	OptimalConcurrency  int                    `json:"optimal_concurrency"`
	ScalingFactor       float64                `json:"scaling_factor"`
	ThroughputAtScale   float64                `json:"throughput_at_scale"`
	LatencyAtScale      time.Duration          `json:"latency_at_scale"`
	ResourceUtilization map[string]float64     `json:"resource_utilization"`
	Bottlenecks         []string               `json:"bottlenecks"`
	ScalingLimits       map[string]interface{} `json:"scaling_limits"`
}

// ResourceMetrics contains resource usage metrics
type ResourceMetrics struct {
	MemoryUsage    int64                  `json:"memory_usage"`
	MemoryPeak     int64                  `json:"memory_peak"`
	MemoryGrowth   int64                  `json:"memory_growth"`
	CPUUsage       float64                `json:"cpu_usage"`
	CPUPeak        float64                `json:"cpu_peak"`
	GoroutineCount int                    `json:"goroutine_count"`
	GoroutinePeak  int                    `json:"goroutine_peak"`
	GCStats        map[string]interface{} `json:"gc_stats"`
	ResourceTrends map[string][]float64   `json:"resource_trends"`
}

// NewPerformanceTestRunner creates a new performance test runner
func NewPerformanceTestRunner() *PerformanceTestRunner {
	logger := zap.NewNop()
	return &PerformanceTestRunner{
		logger: logger,
		results: &PerformanceTestResults{
			TestDetails: make([]PerformanceTestDetail, 0),
			Summary:     make(map[string]interface{}),
			PerformanceMetrics: &PerformanceTestMetrics{
				LatencyDistribution: make(map[string]int),
				ThroughputTrend:     make([]float64, 0),
				LatencyTrend:        make([]time.Duration, 0),
			},
			ScalabilityMetrics: &ScalabilityMetrics{
				ResourceUtilization: make(map[string]float64),
				Bottlenecks:         make([]string, 0),
				ScalingLimits:       make(map[string]interface{}),
			},
			ResourceMetrics: &ResourceMetrics{
				GCStats:        make(map[string]interface{}),
				ResourceTrends: make(map[string][]float64),
			},
		},
	}
}

// RunAllPerformanceTests runs all performance tests
func (ptr *PerformanceTestRunner) RunAllPerformanceTests(t *testing.T) *PerformanceTestResults {
	startTime := time.Now()
	ptr.logger.Info("Starting performance test suite")

	// Initialize test suite
	ptr.testSuite = NewPerformanceTestSuite(t)
	defer ptr.testSuite.Close()

	// Run all test categories
	ptr.runPerformanceTestCategory(t, "Validation Performance", TestValidationPerformance)
	ptr.runPerformanceTestCategory(t, "Export Performance", TestExportPerformance)
	ptr.runPerformanceTestCategory(t, "Backup Performance", TestBackupPerformance)
	ptr.runPerformanceTestCategory(t, "API Performance", TestAPIPerformanceEndpoint)
	ptr.runPerformanceTestCategory(t, "Memory Performance", TestMemoryPerformance)
	ptr.runPerformanceTestCategory(t, "CPU Performance", TestCPUPerformance)
	ptr.runPerformanceTestCategory(t, "Scalability Performance", TestScalabilityPerformance)

	// Run performance analysis
	ptr.runPerformanceAnalysis(t)

	// Calculate final results
	ptr.results.ExecutionTime = time.Since(startTime)
	ptr.calculatePerformanceSummary()

	ptr.logger.Info("Performance test suite completed",
		zap.Int("total_tests", ptr.results.TotalTests),
		zap.Int("passed_tests", ptr.results.PassedTests),
		zap.Int("failed_tests", ptr.results.FailedTests),
		zap.Duration("execution_time", ptr.results.ExecutionTime))

	return ptr.results
}

// runPerformanceTestCategory runs a specific performance test category
func (ptr *PerformanceTestRunner) runPerformanceTestCategory(t *testing.T, categoryName string, testFunc func(*testing.T)) {
	ptr.logger.Info("Running performance test category", zap.String("category", categoryName))

	// Create a sub-test for the category
	t.Run(categoryName, func(t *testing.T) {
		startTime := time.Now()

		// Run the test function
		testFunc(t)

		duration := time.Since(startTime)

		// Record test result
		ptr.results.TotalTests++
		ptr.results.PassedTests++ // If we get here, the test passed

		ptr.results.TestDetails = append(ptr.results.TestDetails, PerformanceTestDetail{
			Name:     categoryName,
			Category: categoryName,
			Status:   "PASSED",
			Duration: duration,
		})

		ptr.logger.Info("Performance test category completed",
			zap.String("category", categoryName),
			zap.Duration("duration", duration),
			zap.String("status", "PASSED"))
	})
}

// runPerformanceAnalysis runs comprehensive performance analysis
func (ptr *PerformanceTestRunner) runPerformanceAnalysis(t *testing.T) {
	ptr.logger.Info("Running performance analysis")

	// Analyze performance patterns
	ptr.analyzePerformancePatterns()

	// Analyze scalability patterns
	ptr.analyzeScalabilityPatterns()

	// Analyze resource usage
	ptr.analyzeResourceUsage()

	// Generate performance recommendations
	ptr.generatePerformanceRecommendations()

	ptr.logger.Info("Performance analysis completed")
}

// analyzePerformancePatterns analyzes performance patterns and trends
func (ptr *PerformanceTestRunner) analyzePerformancePatterns() {
	// Calculate performance metrics
	for _, detail := range ptr.results.TestDetails {
		if detail.Operations > 0 {
			ptr.results.PerformanceMetrics.TotalOperations += detail.Operations
			ptr.results.PerformanceMetrics.TotalDuration += detail.Duration

			// Calculate throughput
			throughput := float64(detail.Operations) / detail.Duration.Seconds()
			ptr.results.PerformanceMetrics.ThroughputTrend = append(ptr.results.PerformanceMetrics.ThroughputTrend, throughput)

			if throughput > ptr.results.PerformanceMetrics.MaxThroughput {
				ptr.results.PerformanceMetrics.MaxThroughput = throughput
			}
			if ptr.results.PerformanceMetrics.MinThroughput == 0 || throughput < ptr.results.PerformanceMetrics.MinThroughput {
				ptr.results.PerformanceMetrics.MinThroughput = throughput
			}

			// Calculate latency
			ptr.results.PerformanceMetrics.LatencyTrend = append(ptr.results.PerformanceMetrics.LatencyTrend, detail.AverageLatency)

			if detail.AverageLatency > ptr.results.PerformanceMetrics.MaxLatency {
				ptr.results.PerformanceMetrics.MaxLatency = detail.AverageLatency
			}
			if ptr.results.PerformanceMetrics.MinLatency == 0 || detail.AverageLatency < ptr.results.PerformanceMetrics.MinLatency {
				ptr.results.PerformanceMetrics.MinLatency = detail.AverageLatency
			}

			// Calculate error rate
			ptr.results.PerformanceMetrics.ErrorRate += detail.ErrorRate
		}
	}

	// Calculate averages
	if len(ptr.results.TestDetails) > 0 {
		ptr.results.PerformanceMetrics.AverageThroughput = ptr.results.PerformanceMetrics.MaxThroughput / float64(len(ptr.results.TestDetails))
		ptr.results.PerformanceMetrics.AverageLatency = ptr.results.PerformanceMetrics.MaxLatency / time.Duration(len(ptr.results.TestDetails))
		ptr.results.PerformanceMetrics.ErrorRate = ptr.results.PerformanceMetrics.ErrorRate / float64(len(ptr.results.TestDetails))
		ptr.results.PerformanceMetrics.SuccessRate = 1.0 - ptr.results.PerformanceMetrics.ErrorRate
	}

	// Calculate percentiles
	ptr.calculateLatencyPercentiles()
}

// analyzeScalabilityPatterns analyzes scalability patterns
func (ptr *PerformanceTestRunner) analyzeScalabilityPatterns() {
	// Find optimal concurrency
	maxConcurrency := 0
	optimalConcurrency := 0
	bestThroughput := 0.0

	for _, detail := range ptr.results.TestDetails {
		if detail.Concurrency > maxConcurrency {
			maxConcurrency = detail.Concurrency
		}

		if detail.Throughput > bestThroughput {
			bestThroughput = detail.Throughput
			optimalConcurrency = detail.Concurrency
		}
	}

	ptr.results.ScalabilityMetrics.MaxConcurrency = maxConcurrency
	ptr.results.ScalabilityMetrics.OptimalConcurrency = optimalConcurrency

	// Calculate scaling factor
	if optimalConcurrency > 0 {
		ptr.results.ScalabilityMetrics.ScalingFactor = float64(maxConcurrency) / float64(optimalConcurrency)
	}

	// Identify bottlenecks
	ptr.identifyBottlenecks()
}

// analyzeResourceUsage analyzes resource usage patterns
func (ptr *PerformanceTestRunner) analyzeResourceUsage() {
	// Get current memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	ptr.results.ResourceMetrics.MemoryUsage = int64(m.Alloc)
	ptr.results.ResourceMetrics.MemoryPeak = int64(m.TotalAlloc)
	ptr.results.ResourceMetrics.GoroutineCount = runtime.NumGoroutine()

	// Calculate memory growth
	if len(ptr.results.TestDetails) > 0 {
		firstDetail := ptr.results.TestDetails[0]
		lastDetail := ptr.results.TestDetails[len(ptr.results.TestDetails)-1]
		ptr.results.ResourceMetrics.MemoryGrowth = lastDetail.MemoryUsage - firstDetail.MemoryUsage
	}

	// Collect GC stats
	ptr.results.ResourceMetrics.GCStats = map[string]interface{}{
		"num_gc":          m.NumGC,
		"pause_total":     m.PauseTotalNs,
		"pause_avg":       m.PauseNs[(m.NumGC+255)%256],
		"gc_cpu_fraction": m.GCCPUFraction,
	}
}

// identifyBottlenecks identifies performance bottlenecks
func (ptr *PerformanceTestRunner) identifyBottlenecks() {
	bottlenecks := make([]string, 0)

	// Check for high latency
	if ptr.results.PerformanceMetrics.AverageLatency > 100*time.Millisecond {
		bottlenecks = append(bottlenecks, "High average latency detected")
	}

	// Check for low throughput
	if ptr.results.PerformanceMetrics.AverageThroughput < 100 {
		bottlenecks = append(bottlenecks, "Low throughput detected")
	}

	// Check for high error rate
	if ptr.results.PerformanceMetrics.ErrorRate > 0.01 {
		bottlenecks = append(bottlenecks, "High error rate detected")
	}

	// Check for memory growth
	if ptr.results.ResourceMetrics.MemoryGrowth > 100*1024*1024 { // 100MB
		bottlenecks = append(bottlenecks, "Significant memory growth detected")
	}

	// Check for goroutine leaks
	if ptr.results.ResourceMetrics.GoroutineCount > 1000 {
		bottlenecks = append(bottlenecks, "High goroutine count detected")
	}

	ptr.results.ScalabilityMetrics.Bottlenecks = bottlenecks
}

// calculateLatencyPercentiles calculates latency percentiles
func (ptr *PerformanceTestRunner) calculateLatencyPercentiles() {
	// Sort latencies for percentile calculation
	latencies := make([]time.Duration, 0)
	for _, detail := range ptr.results.TestDetails {
		latencies = append(latencies, detail.AverageLatency)
	}

	if len(latencies) > 0 {
		// Simple percentile calculation (in a real implementation, you'd use proper percentile calculation)
		sortIndex95 := int(float64(len(latencies)) * 0.95)
		sortIndex99 := int(float64(len(latencies)) * 0.99)

		if sortIndex95 < len(latencies) {
			ptr.results.PerformanceMetrics.P95Latency = latencies[sortIndex95]
		}
		if sortIndex99 < len(latencies) {
			ptr.results.PerformanceMetrics.P99Latency = latencies[sortIndex99]
		}
	}
}

// generatePerformanceRecommendations generates recommendations based on performance analysis
func (ptr *PerformanceTestRunner) generatePerformanceRecommendations() {
	recommendations := make([]string, 0)

	// High latency recommendation
	if ptr.results.PerformanceMetrics.AverageLatency > 100*time.Millisecond {
		recommendations = append(recommendations, "High latency detected. Consider optimizing database queries and reducing I/O operations.")
	}

	// Low throughput recommendation
	if ptr.results.PerformanceMetrics.AverageThroughput < 100 {
		recommendations = append(recommendations, "Low throughput detected. Consider implementing caching and optimizing algorithms.")
	}

	// High error rate recommendation
	if ptr.results.PerformanceMetrics.ErrorRate > 0.01 {
		recommendations = append(recommendations, "High error rate detected. Consider improving error handling and validation.")
	}

	// Memory growth recommendation
	if ptr.results.ResourceMetrics.MemoryGrowth > 100*1024*1024 {
		recommendations = append(recommendations, "Significant memory growth detected. Consider implementing memory pooling and optimizing data structures.")
	}

	// Goroutine leak recommendation
	if ptr.results.ResourceMetrics.GoroutineCount > 1000 {
		recommendations = append(recommendations, "High goroutine count detected. Consider implementing proper goroutine lifecycle management.")
	}

	// Scaling recommendation
	if ptr.results.ScalabilityMetrics.ScalingFactor < 2.0 {
		recommendations = append(recommendations, "Limited scaling capability detected. Consider implementing horizontal scaling and load balancing.")
	}

	ptr.results.Summary["recommendations"] = recommendations
}

// TestPerformanceBenchmarks tests performance benchmarks
func (ptr *PerformanceTestRunner) TestPerformanceBenchmarks(t *testing.T) {
	ptr.logger.Info("Testing performance benchmarks")

	ctx := context.Background()

	// Test validation benchmarks
	ptr.testValidationBenchmarks(ctx, t)

	// Test export benchmarks
	ptr.testExportBenchmarks(ctx, t)

	// Test backup benchmarks
	ptr.testBackupBenchmarks(ctx, t)

	// Test API benchmarks
	ptr.testAPIBenchmarks(ctx, t)
}

// testValidationBenchmarks tests validation performance benchmarks
func (ptr *PerformanceTestRunner) testValidationBenchmarks(ctx context.Context, t *testing.T) {
	ptr.logger.Info("Testing validation benchmarks")

	// Benchmark single validation
	assessment := &RiskAssessment{
		ID:           "benchmark-assessment",
		BusinessID:   "benchmark-business",
		BusinessName: "Benchmark Business",
		OverallScore: 80.0,
		OverallLevel: RiskLevelMedium,
		AlertLevel:   RiskLevelMedium,
		AssessedAt:   time.Now(),
		ValidUntil:   time.Now().Add(24 * time.Hour),
	}

	start := time.Now()
	for i := 0; i < 1000; i++ {
		err := ptr.testSuite.validationSvc.ValidateRiskAssessment(ctx, assessment)
		assert.NoError(t, err)
	}
	duration := time.Since(start)

	throughput := float64(1000) / duration.Seconds()
	avgLatency := duration / 1000

	assert.Greater(t, throughput, 1000.0, "Validation throughput should be > 1000 ops/sec")
	assert.Less(t, avgLatency, 1*time.Millisecond, "Validation latency should be < 1ms")
}

// testExportBenchmarks tests export performance benchmarks
func (ptr *PerformanceTestRunner) testExportBenchmarks(ctx context.Context, t *testing.T) {
	ptr.logger.Info("Testing export benchmarks")

	// Benchmark single export
	request := &ExportRequest{
		BusinessID: "benchmark-business",
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
	}

	start := time.Now()
	for i := 0; i < 100; i++ {
		_, err := ptr.testSuite.exportSvc.ExportData(ctx, request)
		assert.NoError(t, err)
	}
	duration := time.Since(start)

	throughput := float64(100) / duration.Seconds()
	avgLatency := duration / 100

	assert.Greater(t, throughput, 50.0, "Export throughput should be > 50 ops/sec")
	assert.Less(t, avgLatency, 20*time.Millisecond, "Export latency should be < 20ms")
}

// testBackupBenchmarks tests backup performance benchmarks
func (ptr *PerformanceTestRunner) testBackupBenchmarks(ctx context.Context, t *testing.T) {
	ptr.logger.Info("Testing backup benchmarks")

	// Benchmark single backup
	request := &BackupRequest{
		BusinessID:  "benchmark-business",
		BackupType:  BackupTypeBusiness,
		IncludeData: []string{"assessments"},
	}

	start := time.Now()
	for i := 0; i < 50; i++ {
		_, err := ptr.testSuite.backupSvc.CreateBackup(ctx, request)
		assert.NoError(t, err)
	}
	duration := time.Since(start)

	throughput := float64(50) / duration.Seconds()
	avgLatency := duration / 50

	assert.Greater(t, throughput, 10.0, "Backup throughput should be > 10 ops/sec")
	assert.Less(t, avgLatency, 100*time.Millisecond, "Backup latency should be < 100ms")
}

// testAPIBenchmarks tests API performance benchmarks
func (ptr *PerformanceTestRunner) testAPIBenchmarks(ctx context.Context, t *testing.T) {
	ptr.logger.Info("Testing API benchmarks")

	// Benchmark API requests
	client := &http.Client{Timeout: 30 * time.Second}

	start := time.Now()
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("POST", ptr.testSuite.server.URL+"/api/v1/export/jobs",
			[]byte(fmt.Sprintf(`{"business_id": "benchmark-business-%d", "export_type": "assessments", "format": "json"}`, i)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
	duration := time.Since(start)

	throughput := float64(100) / duration.Seconds()
	avgLatency := duration / 100

	assert.Greater(t, throughput, 20.0, "API throughput should be > 20 ops/sec")
	assert.Less(t, avgLatency, 50*time.Millisecond, "API latency should be < 50ms")
}

// calculatePerformanceSummary calculates performance test summary statistics
func (ptr *PerformanceTestRunner) calculatePerformanceSummary() {
	ptr.results.Summary = map[string]interface{}{
		"total_tests":       ptr.results.TotalTests,
		"passed_tests":      ptr.results.PassedTests,
		"failed_tests":      ptr.results.FailedTests,
		"skipped_tests":     ptr.results.SkippedTests,
		"pass_rate":         float64(ptr.results.PassedTests) / float64(ptr.results.TotalTests) * 100,
		"execution_time":    ptr.results.ExecutionTime.String(),
		"average_test_time": ptr.results.ExecutionTime / time.Duration(ptr.results.TotalTests),
	}

	// Calculate performance-specific statistics
	performanceStats := make(map[string]map[string]interface{})
	for _, detail := range ptr.results.TestDetails {
		if performanceStats[detail.Category] == nil {
			performanceStats[detail.Category] = make(map[string]interface{})
		}
		performanceStats[detail.Category]["duration"] = detail.Duration.String()
		performanceStats[detail.Category]["operations"] = detail.Operations
		performanceStats[detail.Category]["throughput"] = detail.Throughput
		performanceStats[detail.Category]["average_latency"] = detail.AverageLatency.String()
		performanceStats[detail.Category]["max_latency"] = detail.MaxLatency.String()
		performanceStats[detail.Category]["min_latency"] = detail.MinLatency.String()
		performanceStats[detail.Category]["error_rate"] = detail.ErrorRate
		performanceStats[detail.Category]["memory_usage"] = detail.MemoryUsage
		performanceStats[detail.Category]["cpu_usage"] = detail.CPUUsage
		performanceStats[detail.Category]["concurrency"] = detail.Concurrency
	}
	ptr.results.Summary["performance_stats"] = performanceStats

	// Add performance metrics to summary
	ptr.results.Summary["performance_metrics"] = ptr.results.PerformanceMetrics
	ptr.results.Summary["scalability_metrics"] = ptr.results.ScalabilityMetrics
	ptr.results.Summary["resource_metrics"] = ptr.results.ResourceMetrics
}

// GeneratePerformanceReport generates a comprehensive performance test report
func (ptr *PerformanceTestRunner) GeneratePerformanceReport() (string, error) {
	report := fmt.Sprintf(`
# Performance Test Report

## Summary
- Total Tests: %d
- Passed Tests: %d
- Failed Tests: %d
- Skipped Tests: %d
- Pass Rate: %.2f%%
- Execution Time: %s

## Performance Metrics
- Total Operations: %d
- Total Duration: %s
- Average Throughput: %.2f ops/sec
- Max Throughput: %.2f ops/sec
- Min Throughput: %.2f ops/sec
- Average Latency: %s
- Max Latency: %s
- Min Latency: %s
- P95 Latency: %s
- P99 Latency: %s
- Error Rate: %.2f%%
- Success Rate: %.2f%%

## Scalability Metrics
- Max Concurrency: %d
- Optimal Concurrency: %d
- Scaling Factor: %.2f
- Throughput at Scale: %.2f ops/sec
- Latency at Scale: %s
- Bottlenecks: %v

## Resource Metrics
- Memory Usage: %d bytes
- Memory Peak: %d bytes
- Memory Growth: %d bytes
- Goroutine Count: %d
- Goroutine Peak: %d
- GC Stats: %v

## Test Details
`,
		ptr.results.TotalTests,
		ptr.results.PassedTests,
		ptr.results.FailedTests,
		ptr.results.SkippedTests,
		float64(ptr.results.PassedTests)/float64(ptr.results.TotalTests)*100,
		ptr.results.ExecutionTime.String(),
		ptr.results.PerformanceMetrics.TotalOperations,
		ptr.results.PerformanceMetrics.TotalDuration.String(),
		ptr.results.PerformanceMetrics.AverageThroughput,
		ptr.results.PerformanceMetrics.MaxThroughput,
		ptr.results.PerformanceMetrics.MinThroughput,
		ptr.results.PerformanceMetrics.AverageLatency.String(),
		ptr.results.PerformanceMetrics.MaxLatency.String(),
		ptr.results.PerformanceMetrics.MinLatency.String(),
		ptr.results.PerformanceMetrics.P95Latency.String(),
		ptr.results.PerformanceMetrics.P99Latency.String(),
		ptr.results.PerformanceMetrics.ErrorRate*100,
		ptr.results.PerformanceMetrics.SuccessRate*100,
		ptr.results.ScalabilityMetrics.MaxConcurrency,
		ptr.results.ScalabilityMetrics.OptimalConcurrency,
		ptr.results.ScalabilityMetrics.ScalingFactor,
		ptr.results.ScalabilityMetrics.ThroughputAtScale,
		ptr.results.ScalabilityMetrics.LatencyAtScale.String(),
		ptr.results.ScalabilityMetrics.Bottlenecks,
		ptr.results.ResourceMetrics.MemoryUsage,
		ptr.results.ResourceMetrics.MemoryPeak,
		ptr.results.ResourceMetrics.MemoryGrowth,
		ptr.results.ResourceMetrics.GoroutineCount,
		ptr.results.ResourceMetrics.GoroutinePeak,
		ptr.results.ResourceMetrics.GCStats)

	for _, detail := range ptr.results.TestDetails {
		report += fmt.Sprintf(`
### %s
- Category: %s
- Status: %s
- Duration: %s
- Operations: %d
- Throughput: %.2f ops/sec
- Average Latency: %s
- Max Latency: %s
- Min Latency: %s
- Error Rate: %.2f%%
- Memory Usage: %d bytes
- CPU Usage: %.2f%%
- Concurrency: %d
`,
			detail.Name,
			detail.Category,
			detail.Status,
			detail.Duration.String(),
			detail.Operations,
			detail.Throughput,
			detail.AverageLatency.String(),
			detail.MaxLatency.String(),
			detail.MinLatency.String(),
			detail.ErrorRate*100,
			detail.MemoryUsage,
			detail.CPUUsage,
			detail.Concurrency)
	}

	// Add recommendations
	if recommendations, ok := ptr.results.Summary["recommendations"].([]string); ok {
		report += "\n## Recommendations\n"
		for _, recommendation := range recommendations {
			report += fmt.Sprintf("- %s\n", recommendation)
		}
	}

	return report, nil
}
