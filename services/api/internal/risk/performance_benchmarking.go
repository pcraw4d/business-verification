package risk

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PerformanceBenchmarking provides comprehensive performance benchmarking for the KYB platform
type PerformanceBenchmarking struct {
	logger           *zap.Logger
	config           *BenchmarkConfig
	benchmarks       map[string]*Benchmark
	results          *BenchmarkResults
	metricsCollector *MetricsCollector
	reportGenerator  *BenchmarkReportGenerator
}

// BenchmarkConfig contains configuration for performance benchmarking
type BenchmarkConfig struct {
	TestEnvironment      string            `json:"test_environment"`
	BenchmarkTimeout     time.Duration     `json:"benchmark_timeout"`
	ReportOutputPath     string            `json:"report_output_path"`
	LogLevel             string            `json:"log_level"`
	EnableProfiling      bool              `json:"enable_profiling"`
	ProfilingOutputPath  string            `json:"profiling_output_path"`
	ConcurrencyLevels    []int             `json:"concurrency_levels"`
	TestDataSizes        []int             `json:"test_data_sizes"`
	IterationCounts      []int             `json:"iteration_counts"`
	WarmupIterations     int               `json:"warmup_iterations"`
	EnvironmentVariables map[string]string `json:"environment_variables"`
	DatabaseConfig       *DatabaseConfig   `json:"database_config"`
	APIConfig            *APIConfig        `json:"api_config"`
	ResourceLimits       *ResourceLimits   `json:"resource_limits"`
}

// ResourceLimits contains resource limits for benchmarking
type ResourceLimits struct {
	MaxMemoryMB     int           `json:"max_memory_mb"`
	MaxCPUPercents  int           `json:"max_cpu_percents"`
	MaxGoroutines   int           `json:"max_goroutines"`
	MaxConnections  int           `json:"max_connections"`
	MaxFileHandles  int           `json:"max_file_handles"`
	TimeoutDuration time.Duration `json:"timeout_duration"`
}

// Benchmark represents a performance benchmark
type Benchmark struct {
	ID              string                                  `json:"id"`
	Name            string                                  `json:"name"`
	Description     string                                  `json:"description"`
	Category        string                                  `json:"category"`
	Priority        string                                  `json:"priority"`
	Function        func(*BenchmarkContext) BenchmarkResult `json:"-"`
	SetupFunction   func(*BenchmarkContext) error           `json:"-"`
	CleanupFunction func(*BenchmarkContext) error           `json:"-"`
	Parameters      map[string]interface{}                  `json:"parameters"`
	ExpectedMetrics *ExpectedMetrics                        `json:"expected_metrics"`
	Tags            []string                                `json:"tags"`
}

// BenchmarkContext provides context for benchmark execution
type BenchmarkContext struct {
	ID               string                 `json:"id"`
	Iteration        int                    `json:"iteration"`
	ConcurrencyLevel int                    `json:"concurrency_level"`
	TestDataSize     int                    `json:"test_data_size"`
	Parameters       map[string]interface{} `json:"parameters"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	Duration         time.Duration          `json:"duration"`
	MemoryBefore     uint64                 `json:"memory_before"`
	MemoryAfter      uint64                 `json:"memory_after"`
	MemoryDelta      int64                  `json:"memory_delta"`
	GoroutinesBefore int                    `json:"goroutines_before"`
	GoroutinesAfter  int                    `json:"goroutines_after"`
	GoroutinesDelta  int                    `json:"goroutines_delta"`
	Context          context.Context        `json:"-"`
	Logger           *zap.Logger            `json:"-"`
}

// BenchmarkResult contains the result of a benchmark execution
type BenchmarkResult struct {
	BenchmarkID      string                 `json:"benchmark_id"`
	Iteration        int                    `json:"iteration"`
	ConcurrencyLevel int                    `json:"concurrency_level"`
	TestDataSize     int                    `json:"test_data_size"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	Duration         time.Duration          `json:"duration"`
	Success          bool                   `json:"success"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	Metrics          *PerformanceMetrics    `json:"metrics"`
	ResourceUsage    *ResourceUsage         `json:"resource_usage"`
	CustomMetrics    map[string]interface{} `json:"custom_metrics"`
}

// PerformanceMetrics contains performance metrics
type PerformanceMetrics struct {
	Throughput      float64 `json:"throughput"`       // Operations per second
	Latency         float64 `json:"latency"`          // Average latency in milliseconds
	P95Latency      float64 `json:"p95_latency"`      // 95th percentile latency
	P99Latency      float64 `json:"p99_latency"`      // 99th percentile latency
	MaxLatency      float64 `json:"max_latency"`      // Maximum latency
	MinLatency      float64 `json:"min_latency"`      // Minimum latency
	ErrorRate       float64 `json:"error_rate"`       // Error rate percentage
	SuccessRate     float64 `json:"success_rate"`     // Success rate percentage
	OperationsCount int     `json:"operations_count"` // Total operations performed
	SuccessfulOps   int     `json:"successful_ops"`   // Successful operations
	FailedOps       int     `json:"failed_ops"`       // Failed operations
	DataProcessed   int64   `json:"data_processed"`   // Data processed in bytes
	DataThroughput  float64 `json:"data_throughput"`  // Data throughput in MB/s
}

// ResourceUsage contains resource usage metrics
type ResourceUsage struct {
	MemoryUsage        uint64  `json:"memory_usage"`        // Memory usage in bytes
	MemoryPeak         uint64  `json:"memory_peak"`         // Peak memory usage
	MemoryDelta        int64   `json:"memory_delta"`        // Memory delta
	CPUUsage           float64 `json:"cpu_usage"`           // CPU usage percentage
	CPUTime            float64 `json:"cpu_time"`            // CPU time in seconds
	GoroutinesCount    int     `json:"goroutines_count"`    // Number of goroutines
	GoroutinesPeak     int     `json:"goroutines_peak"`     // Peak goroutines
	GoroutinesDelta    int     `json:"goroutines_delta"`    // Goroutines delta
	GCCollections      int     `json:"gc_collections"`      // GC collections
	GCPauseTime        float64 `json:"gc_pause_time"`       // GC pause time in milliseconds
	FileHandles        int     `json:"file_handles"`        // File handles count
	NetworkConnections int     `json:"network_connections"` // Network connections
}

// ExpectedMetrics contains expected performance metrics
type ExpectedMetrics struct {
	MinThroughput  float64 `json:"min_throughput"`
	MaxLatency     float64 `json:"max_latency"`
	MaxP95Latency  float64 `json:"max_p95_latency"`
	MaxP99Latency  float64 `json:"max_p99_latency"`
	MaxErrorRate   float64 `json:"max_error_rate"`
	MaxMemoryUsage uint64  `json:"max_memory_usage"`
	MaxCPUUsage    float64 `json:"max_cpu_usage"`
	MaxGoroutines  int     `json:"max_goroutines"`
	MinSuccessRate float64 `json:"min_success_rate"`
}

// BenchmarkResults contains the results of benchmark execution
type BenchmarkResults struct {
	SessionID         string                       `json:"session_id"`
	StartTime         time.Time                    `json:"start_time"`
	EndTime           time.Time                    `json:"end_time"`
	TotalDuration     time.Duration                `json:"total_duration"`
	Environment       string                       `json:"environment"`
	TotalBenchmarks   int                          `json:"total_benchmarks"`
	PassedBenchmarks  int                          `json:"passed_benchmarks"`
	FailedBenchmarks  int                          `json:"failed_benchmarks"`
	SkippedBenchmarks int                          `json:"skipped_benchmarks"`
	PassRate          float64                      `json:"pass_rate"`
	BenchmarkResults  map[string][]BenchmarkResult `json:"benchmark_results"`
	Summary           *BenchmarkSummary            `json:"summary"`
	Recommendations   []string                     `json:"recommendations"`
	Issues            []BenchmarkIssue             `json:"issues"`
}

// BenchmarkSummary contains a summary of benchmark results
type BenchmarkSummary struct {
	OverallThroughput   float64                    `json:"overall_throughput"`
	OverallLatency      float64                    `json:"overall_latency"`
	OverallP95Latency   float64                    `json:"overall_p95_latency"`
	OverallP99Latency   float64                    `json:"overall_p99_latency"`
	OverallErrorRate    float64                    `json:"overall_error_rate"`
	OverallSuccessRate  float64                    `json:"overall_success_rate"`
	OverallMemoryUsage  uint64                     `json:"overall_memory_usage"`
	OverallCPUUsage     float64                    `json:"overall_cpu_usage"`
	OverallGoroutines   int                        `json:"overall_goroutines"`
	CategoryMetrics     map[string]CategoryMetrics `json:"category_metrics"`
	PerformanceTrends   []PerformanceTrend         `json:"performance_trends"`
	ResourceUtilization *ResourceUtilization       `json:"resource_utilization"`
}

// CategoryMetrics contains metrics for a specific category
type CategoryMetrics struct {
	CategoryName   string  `json:"category_name"`
	BenchmarkCount int     `json:"benchmark_count"`
	AvgThroughput  float64 `json:"avg_throughput"`
	AvgLatency     float64 `json:"avg_latency"`
	AvgP95Latency  float64 `json:"avg_p95_latency"`
	AvgP99Latency  float64 `json:"avg_p99_latency"`
	AvgErrorRate   float64 `json:"avg_error_rate"`
	AvgSuccessRate float64 `json:"avg_success_rate"`
	AvgMemoryUsage uint64  `json:"avg_memory_usage"`
	AvgCPUUsage    float64 `json:"avg_cpu_usage"`
	AvgGoroutines  int     `json:"avg_goroutines"`
	PassRate       float64 `json:"pass_rate"`
}

// PerformanceTrend represents a performance trend
type PerformanceTrend struct {
	MetricName    string      `json:"metric_name"`
	Trend         string      `json:"trend"` // "improving", "degrading", "stable"
	ChangePercent float64     `json:"change_percent"`
	DataPoints    []float64   `json:"data_points"`
	Timestamps    []time.Time `json:"timestamps"`
}

// ResourceUtilization contains resource utilization metrics
type ResourceUtilization struct {
	MemoryUtilization     float64 `json:"memory_utilization"`
	CPUUtilization        float64 `json:"cpu_utilization"`
	GoroutineUtilization  float64 `json:"goroutine_utilization"`
	FileHandleUtilization float64 `json:"file_handle_utilization"`
	NetworkUtilization    float64 `json:"network_utilization"`
	DatabaseUtilization   float64 `json:"database_utilization"`
}

// BenchmarkIssue represents an issue found during benchmarking
type BenchmarkIssue struct {
	ID             string      `json:"id"`
	BenchmarkID    string      `json:"benchmark_id"`
	Severity       string      `json:"severity"` // Critical, High, Medium, Low
	Category       string      `json:"category"`
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	ExpectedValue  interface{} `json:"expected_value"`
	ActualValue    interface{} `json:"actual_value"`
	Impact         string      `json:"impact"`
	Recommendation string      `json:"recommendation"`
	DetectedAt     time.Time   `json:"detected_at"`
	Tags           []string    `json:"tags"`
}

// MetricsCollector collects and aggregates performance metrics
type MetricsCollector struct {
	logger    *zap.Logger
	metrics   map[string]*PerformanceMetrics
	mutex     sync.RWMutex
	startTime time.Time
}

// BenchmarkReportGenerator generates benchmark reports
type BenchmarkReportGenerator struct {
	logger *zap.Logger
	config *BenchmarkConfig
}

// NewPerformanceBenchmarking creates a new performance benchmarking instance
func NewPerformanceBenchmarking(config *BenchmarkConfig) *PerformanceBenchmarking {
	logger := zap.NewNop()
	if config.LogLevel == "debug" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	return &PerformanceBenchmarking{
		logger:           logger,
		config:           config,
		benchmarks:       make(map[string]*Benchmark),
		results:          &BenchmarkResults{},
		metricsCollector: NewMetricsCollector(logger),
		reportGenerator:  NewBenchmarkReportGenerator(logger, config),
	}
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		logger:    logger,
		metrics:   make(map[string]*PerformanceMetrics),
		startTime: time.Now(),
	}
}

// NewBenchmarkReportGenerator creates a new benchmark report generator
func NewBenchmarkReportGenerator(logger *zap.Logger, config *BenchmarkConfig) *BenchmarkReportGenerator {
	return &BenchmarkReportGenerator{
		logger: logger,
		config: config,
	}
}

// AddBenchmark adds a benchmark to the benchmarking suite
func (pb *PerformanceBenchmarking) AddBenchmark(benchmark *Benchmark) {
	pb.benchmarks[benchmark.ID] = benchmark
	pb.logger.Info("Added benchmark", zap.String("id", benchmark.ID), zap.String("name", benchmark.Name))
}

// RunBenchmark runs a specific benchmark
func (pb *PerformanceBenchmarking) RunBenchmark(ctx context.Context, benchmarkID string) (*BenchmarkResult, error) {
	benchmark, exists := pb.benchmarks[benchmarkID]
	if !exists {
		return nil, fmt.Errorf("benchmark with ID %s not found", benchmarkID)
	}

	pb.logger.Info("Running benchmark", zap.String("id", benchmarkID), zap.String("name", benchmark.Name))

	// Create benchmark context
	benchmarkCtx := &BenchmarkContext{
		ID:               benchmarkID,
		Parameters:       benchmark.Parameters,
		Context:          ctx,
		Logger:           pb.logger,
		MemoryBefore:     pb.getMemoryUsage(),
		GoroutinesBefore: runtime.NumGoroutine(),
		StartTime:        time.Now(),
	}

	// Run setup if provided
	if benchmark.SetupFunction != nil {
		if err := benchmark.SetupFunction(benchmarkCtx); err != nil {
			return nil, fmt.Errorf("benchmark setup failed: %w", err)
		}
	}

	// Execute benchmark
	result := benchmark.Function(benchmarkCtx)
	result.BenchmarkID = benchmarkID
	result.StartTime = benchmarkCtx.StartTime
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Update context
	benchmarkCtx.EndTime = result.EndTime
	benchmarkCtx.Duration = result.Duration
	benchmarkCtx.MemoryAfter = pb.getMemoryUsage()
	benchmarkCtx.MemoryDelta = int64(benchmarkCtx.MemoryAfter) - int64(benchmarkCtx.MemoryBefore)
	benchmarkCtx.GoroutinesAfter = runtime.NumGoroutine()
	benchmarkCtx.GoroutinesDelta = benchmarkCtx.GoroutinesAfter - benchmarkCtx.GoroutinesBefore

	// Update resource usage
	if result.ResourceUsage == nil {
		result.ResourceUsage = &ResourceUsage{}
	}
	result.ResourceUsage.MemoryUsage = benchmarkCtx.MemoryAfter
	result.ResourceUsage.MemoryDelta = benchmarkCtx.MemoryDelta
	result.ResourceUsage.GoroutinesCount = benchmarkCtx.GoroutinesAfter
	result.ResourceUsage.GoroutinesDelta = benchmarkCtx.GoroutinesDelta

	// Run cleanup if provided
	if benchmark.CleanupFunction != nil {
		if err := benchmark.CleanupFunction(benchmarkCtx); err != nil {
			pb.logger.Warn("Benchmark cleanup failed", zap.String("benchmark_id", benchmarkID), zap.Error(err))
		}
	}

	pb.logger.Info("Benchmark completed",
		zap.String("id", benchmarkID),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration))

	return &result, nil
}

// RunBenchmarkSuite runs all benchmarks in the suite
func (pb *PerformanceBenchmarking) RunBenchmarkSuite(ctx context.Context) (*BenchmarkResults, error) {
	pb.logger.Info("Starting benchmark suite execution")

	// Initialize results
	pb.results = &BenchmarkResults{
		SessionID:        fmt.Sprintf("benchmark_session_%d", time.Now().Unix()),
		StartTime:        time.Now(),
		Environment:      pb.config.TestEnvironment,
		BenchmarkResults: make(map[string][]BenchmarkResult),
		Summary:          &BenchmarkSummary{},
		Recommendations:  make([]string, 0),
		Issues:           make([]BenchmarkIssue, 0),
	}

	pb.results.TotalBenchmarks = len(pb.benchmarks)

	// Run each benchmark
	for benchmarkID, benchmark := range pb.benchmarks {
		select {
		case <-ctx.Done():
			pb.results.EndTime = time.Now()
			pb.results.TotalDuration = pb.results.EndTime.Sub(pb.results.StartTime)
			return pb.results, ctx.Err()
		default:
		}

		pb.logger.Info("Running benchmark", zap.String("id", benchmarkID), zap.String("name", benchmark.Name))

		// Run benchmark with different concurrency levels and data sizes
		benchmarkResults := make([]BenchmarkResult, 0)

		for _, concurrencyLevel := range pb.config.ConcurrencyLevels {
			for _, testDataSize := range pb.config.TestDataSizes {
				for _, iterationCount := range pb.config.IterationCounts {
					// Update benchmark parameters
					benchmark.Parameters["concurrency_level"] = concurrencyLevel
					benchmark.Parameters["test_data_size"] = testDataSize
					benchmark.Parameters["iteration_count"] = iterationCount

					// Run warmup iterations
					for i := 0; i < pb.config.WarmupIterations; i++ {
						_, err := pb.RunBenchmark(ctx, benchmarkID)
						if err != nil {
							pb.logger.Warn("Warmup iteration failed", zap.String("benchmark_id", benchmarkID), zap.Int("iteration", i), zap.Error(err))
						}
					}

					// Run actual benchmark iterations
					for i := 0; i < iterationCount; i++ {
						result, err := pb.RunBenchmark(ctx, benchmarkID)
						if err != nil {
							pb.logger.Error("Benchmark execution failed", zap.String("benchmark_id", benchmarkID), zap.Error(err))
							pb.results.FailedBenchmarks++
							continue
						}

						benchmarkResults = append(benchmarkResults, *result)

						// Check if benchmark passed
						if pb.isBenchmarkPassed(benchmark, result) {
							pb.results.PassedBenchmarks++
						} else {
							pb.results.FailedBenchmarks++
						}
					}
				}
			}
		}

		pb.results.BenchmarkResults[benchmarkID] = benchmarkResults
	}

	// Calculate pass rate
	if pb.results.TotalBenchmarks > 0 {
		pb.results.PassRate = float64(pb.results.PassedBenchmarks) / float64(pb.results.TotalBenchmarks) * 100
	}

	pb.results.EndTime = time.Now()
	pb.results.TotalDuration = pb.results.EndTime.Sub(pb.results.StartTime)

	// Generate summary
	pb.generateSummary()

	// Generate recommendations
	pb.generateRecommendations()

	// Generate reports
	if err := pb.generateReports(); err != nil {
		pb.logger.Error("Failed to generate reports", zap.Error(err))
	}

	pb.logger.Info("Benchmark suite execution completed",
		zap.Int("total_benchmarks", pb.results.TotalBenchmarks),
		zap.Int("passed_benchmarks", pb.results.PassedBenchmarks),
		zap.Int("failed_benchmarks", pb.results.FailedBenchmarks),
		zap.Float64("pass_rate", pb.results.PassRate))

	return pb.results, nil
}

// isBenchmarkPassed checks if a benchmark passed based on expected metrics
func (pb *PerformanceBenchmarking) isBenchmarkPassed(benchmark *Benchmark, result *BenchmarkResult) bool {
	if !result.Success {
		return false
	}

	if benchmark.ExpectedMetrics == nil {
		return true
	}

	expected := benchmark.ExpectedMetrics
	metrics := result.Metrics

	// Check throughput
	if expected.MinThroughput > 0 && metrics.Throughput < expected.MinThroughput {
		return false
	}

	// Check latency
	if expected.MaxLatency > 0 && metrics.Latency > expected.MaxLatency {
		return false
	}

	// Check P95 latency
	if expected.MaxP95Latency > 0 && metrics.P95Latency > expected.MaxP95Latency {
		return false
	}

	// Check P99 latency
	if expected.MaxP99Latency > 0 && metrics.P99Latency > expected.MaxP99Latency {
		return false
	}

	// Check error rate
	if expected.MaxErrorRate > 0 && metrics.ErrorRate > expected.MaxErrorRate {
		return false
	}

	// Check success rate
	if expected.MinSuccessRate > 0 && metrics.SuccessRate < expected.MinSuccessRate {
		return false
	}

	return true
}

// getMemoryUsage gets current memory usage
func (pb *PerformanceBenchmarking) getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// generateSummary generates a summary of benchmark results
func (pb *PerformanceBenchmarking) generateSummary() {
	summary := &BenchmarkSummary{
		CategoryMetrics:     make(map[string]CategoryMetrics),
		PerformanceTrends:   make([]PerformanceTrend, 0),
		ResourceUtilization: &ResourceUtilization{},
	}

	// Aggregate metrics across all benchmarks
	totalThroughput := 0.0
	totalLatency := 0.0
	totalP95Latency := 0.0
	totalP99Latency := 0.0
	totalErrorRate := 0.0
	totalSuccessRate := 0.0
	totalMemoryUsage := uint64(0)
	totalCPUUsage := 0.0
	totalGoroutines := 0
	benchmarkCount := 0

	categoryMetrics := make(map[string]*CategoryMetrics)

	for benchmarkID, results := range pb.results.BenchmarkResults {
		benchmark := pb.benchmarks[benchmarkID]
		category := benchmark.Category

		// Initialize category metrics if not exists
		if categoryMetrics[category] == nil {
			categoryMetrics[category] = &CategoryMetrics{
				CategoryName: category,
			}
		}

		// Aggregate results for this benchmark
		for _, result := range results {
			if result.Metrics != nil {
				totalThroughput += result.Metrics.Throughput
				totalLatency += result.Metrics.Latency
				totalP95Latency += result.Metrics.P95Latency
				totalP99Latency += result.Metrics.P99Latency
				totalErrorRate += result.Metrics.ErrorRate
				totalSuccessRate += result.Metrics.SuccessRate
				benchmarkCount++

				// Update category metrics
				catMetrics := categoryMetrics[category]
				catMetrics.BenchmarkCount++
				catMetrics.AvgThroughput += result.Metrics.Throughput
				catMetrics.AvgLatency += result.Metrics.Latency
				catMetrics.AvgP95Latency += result.Metrics.P95Latency
				catMetrics.AvgP99Latency += result.Metrics.P99Latency
				catMetrics.AvgErrorRate += result.Metrics.ErrorRate
				catMetrics.AvgSuccessRate += result.Metrics.SuccessRate
			}

			if result.ResourceUsage != nil {
				totalMemoryUsage += result.ResourceUsage.MemoryUsage
				totalCPUUsage += result.ResourceUsage.CPUUsage
				totalGoroutines += result.ResourceUsage.GoroutinesCount
			}
		}
	}

	// Calculate averages
	if benchmarkCount > 0 {
		summary.OverallThroughput = totalThroughput / float64(benchmarkCount)
		summary.OverallLatency = totalLatency / float64(benchmarkCount)
		summary.OverallP95Latency = totalP95Latency / float64(benchmarkCount)
		summary.OverallP99Latency = totalP99Latency / float64(benchmarkCount)
		summary.OverallErrorRate = totalErrorRate / float64(benchmarkCount)
		summary.OverallSuccessRate = totalSuccessRate / float64(benchmarkCount)
		summary.OverallMemoryUsage = totalMemoryUsage / uint64(benchmarkCount)
		summary.OverallCPUUsage = totalCPUUsage / float64(benchmarkCount)
		summary.OverallGoroutines = totalGoroutines / benchmarkCount
	}

	// Calculate category averages
	for category, catMetrics := range categoryMetrics {
		if catMetrics.BenchmarkCount > 0 {
			catMetrics.AvgThroughput /= float64(catMetrics.BenchmarkCount)
			catMetrics.AvgLatency /= float64(catMetrics.BenchmarkCount)
			catMetrics.AvgP95Latency /= float64(catMetrics.BenchmarkCount)
			catMetrics.AvgP99Latency /= float64(catMetrics.BenchmarkCount)
			catMetrics.AvgErrorRate /= float64(catMetrics.BenchmarkCount)
			catMetrics.AvgSuccessRate /= float64(catMetrics.BenchmarkCount)
		}
		summary.CategoryMetrics[category] = *catMetrics
	}

	pb.results.Summary = summary
}

// generateRecommendations generates recommendations based on benchmark results
func (pb *PerformanceBenchmarking) generateRecommendations() {
	recommendations := make([]string, 0)

	// Low pass rate recommendation
	if pb.results.PassRate < 90 {
		recommendations = append(recommendations, "Low pass rate detected. Review failed benchmarks and optimize performance.")
	}

	// High latency recommendation
	if pb.results.Summary.OverallLatency > 1000 {
		recommendations = append(recommendations, "High latency detected. Consider optimizing database queries and API responses.")
	}

	// High error rate recommendation
	if pb.results.Summary.OverallErrorRate > 5 {
		recommendations = append(recommendations, "High error rate detected. Review error handling and system stability.")
	}

	// High memory usage recommendation
	if pb.results.Summary.OverallMemoryUsage > 100*1024*1024 { // 100MB
		recommendations = append(recommendations, "High memory usage detected. Consider memory optimization and garbage collection tuning.")
	}

	// High CPU usage recommendation
	if pb.results.Summary.OverallCPUUsage > 80 {
		recommendations = append(recommendations, "High CPU usage detected. Consider CPU optimization and load balancing.")
	}

	pb.results.Recommendations = recommendations
}

// generateReports generates benchmark reports
func (pb *PerformanceBenchmarking) generateReports() error {
	pb.logger.Info("Generating benchmark reports")
	return pb.reportGenerator.GenerateReports(pb.results)
}

// GetResults returns the benchmark results
func (pb *PerformanceBenchmarking) GetResults() *BenchmarkResults {
	return pb.results
}

// GetBenchmarks returns all benchmarks
func (pb *PerformanceBenchmarking) GetBenchmarks() map[string]*Benchmark {
	return pb.benchmarks
}
