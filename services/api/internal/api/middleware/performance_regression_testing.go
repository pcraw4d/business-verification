package middleware

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ResourceUtilization represents resource utilization metrics
type ResourceUtilization struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Disk    float64 `json:"disk"`
	Network float64 `json:"network"`
}

// RegressionPerformanceMetric represents a performance metric for regression testing
type RegressionPerformanceMetric struct {
	ResponseTime        time.Duration        `json:"response_time"`
	Throughput          float64              `json:"throughput"`
	ErrorRate           float64              `json:"error_rate"`
	ResourceUtilization *ResourceUtilization `json:"resource_utilization,omitempty"`
	Timestamp           time.Time            `json:"timestamp"`
}

// RegressionTestConfig defines configuration for performance regression testing
type RegressionTestConfig struct {
	// Baseline configuration
	BaselineWindow     time.Duration `json:"baseline_window"`      // Time window for baseline calculation
	BaselineMinSamples int           `json:"baseline_min_samples"` // Minimum samples required for baseline
	BaselinePercentile float64       `json:"baseline_percentile"`  // Percentile to use for baseline (e.g., 95th)

	// Regression detection thresholds
	ResponseTimeThreshold   float64 `json:"response_time_threshold"`  // Percentage increase to trigger regression
	ThroughputThreshold     float64 `json:"throughput_threshold"`     // Percentage decrease to trigger regression
	ErrorRateThreshold      float64 `json:"error_rate_threshold"`     // Percentage increase to trigger regression
	StatisticalSignificance float64 `json:"statistical_significance"` // P-value threshold for statistical significance
	MinimumSampleSize       int     `json:"minimum_sample_size"`      // Minimum samples for statistical testing

	// Alerting configuration
	AlertOnRegression     bool `json:"alert_on_regression"`      // Whether to send alerts on regression
	AlertOnImprovement    bool `json:"alert_on_improvement"`     // Whether to send alerts on improvement
	AlertOnBaselineUpdate bool `json:"alert_on_baseline_update"` // Whether to alert when baseline is updated

	// Storage configuration
	BaselineRetentionDays int `json:"baseline_retention_days"` // How long to keep baseline data
	MetricsRetentionDays  int `json:"metrics_retention_days"`  // How long to keep metrics data
}

// PerformanceBaseline represents a performance baseline
type PerformanceBaseline struct {
	ID          string    `json:"id"`
	Endpoint    string    `json:"endpoint"`
	Method      string    `json:"method"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	SampleCount int       `json:"sample_count"`

	// Response time metrics
	ResponseTime struct {
		P50  time.Duration `json:"p50"`
		P95  time.Duration `json:"p95"`
		P99  time.Duration `json:"p99"`
		Mean time.Duration `json:"mean"`
		Std  time.Duration `json:"std"`
	} `json:"response_time"`

	// Throughput metrics
	Throughput struct {
		RequestsPerSecond float64 `json:"requests_per_second"`
		Mean              float64 `json:"mean"`
		Std               float64 `json:"std"`
	} `json:"throughput"`

	// Error rate metrics
	ErrorRate struct {
		Percentage float64 `json:"percentage"`
		Mean       float64 `json:"mean"`
		Std        float64 `json:"std"`
	} `json:"error_rate"`

	// Resource utilization metrics
	ResourceUtilization struct {
		CPU     float64 `json:"cpu"`
		Memory  float64 `json:"memory"`
		Disk    float64 `json:"disk"`
		Network float64 `json:"network"`
	} `json:"resource_utilization"`

	// Metadata
	Environment string            `json:"environment"`
	Version     string            `json:"version"`
	Tags        map[string]string `json:"tags"`
}

// RegressionResult represents the result of a regression test
type RegressionResult struct {
	ID          string    `json:"id"`
	BaselineID  string    `json:"baseline_id"`
	Endpoint    string    `json:"endpoint"`
	Method      string    `json:"method"`
	TestedAt    time.Time `json:"tested_at"`
	SampleCount int       `json:"sample_count"`

	// Regression detection results
	ResponseTimeRegression        *RegressionMetric `json:"response_time_regression,omitempty"`
	ThroughputRegression          *RegressionMetric `json:"throughput_regression,omitempty"`
	ErrorRateRegression           *RegressionMetric `json:"error_rate_regression,omitempty"`
	ResourceUtilizationRegression *RegressionMetric `json:"resource_utilization_regression,omitempty"`

	// Statistical analysis
	StatisticalSignificance float64 `json:"statistical_significance"`
	ConfidenceInterval      struct {
		Lower float64 `json:"lower"`
		Upper float64 `json:"upper"`
	} `json:"confidence_interval"`

	// Overall assessment
	HasRegression bool    `json:"has_regression"`
	Severity      string  `json:"severity"` // "none", "low", "medium", "high", "critical"
	Score         float64 `json:"score"`    // 0-100, higher means worse performance

	// Recommendations
	Recommendations []string `json:"recommendations"`
}

// RegressionMetric represents a specific metric regression
type RegressionMetric struct {
	MetricName    string  `json:"metric_name"`
	BaselineValue float64 `json:"baseline_value"`
	CurrentValue  float64 `json:"current_value"`
	ChangePercent float64 `json:"change_percent"`
	IsRegression  bool    `json:"is_regression"`
	IsSignificant bool    `json:"is_significant"`
	PValue        float64 `json:"p_value"`
	EffectSize    float64 `json:"effect_size"`
	Severity      string  `json:"severity"`
}

// PerformanceRegressionTester manages performance regression testing
type PerformanceRegressionTester struct {
	config    *RegressionTestConfig
	logger    *zap.Logger
	baselines map[string]*PerformanceBaseline
	results   map[string]*RegressionResult
	mu        sync.RWMutex
	stopChan  chan struct{}
}

// NewPerformanceRegressionTester creates a new performance regression tester
func NewPerformanceRegressionTester(config *RegressionTestConfig, logger *zap.Logger) *PerformanceRegressionTester {
	if config == nil {
		config = &RegressionTestConfig{
			BaselineWindow:          24 * time.Hour,
			BaselineMinSamples:      100,
			BaselinePercentile:      95.0,
			ResponseTimeThreshold:   10.0, // 10% increase
			ThroughputThreshold:     5.0,  // 5% decrease
			ErrorRateThreshold:      2.0,  // 2% increase
			StatisticalSignificance: 0.05, // 5% significance level
			MinimumSampleSize:       30,
			AlertOnRegression:       true,
			AlertOnImprovement:      false,
			AlertOnBaselineUpdate:   false,
			BaselineRetentionDays:   30,
			MetricsRetentionDays:    7,
		}
	}

	return &PerformanceRegressionTester{
		config:    config,
		logger:    logger,
		baselines: make(map[string]*PerformanceBaseline),
		results:   make(map[string]*RegressionResult),
		stopChan:  make(chan struct{}),
	}
}

// CreateBaseline creates a new performance baseline from metrics
func (prt *PerformanceRegressionTester) CreateBaseline(ctx context.Context, endpoint, method string, metrics []RegressionPerformanceMetric) (*PerformanceBaseline, error) {
	if len(metrics) < prt.config.BaselineMinSamples {
		return nil, fmt.Errorf("insufficient samples for baseline: got %d, need %d", len(metrics), prt.config.BaselineMinSamples)
	}

	baseline := &PerformanceBaseline{
		ID:        generateID(),
		Endpoint:  endpoint,
		Method:    method,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tags:      make(map[string]string),
	}

	// Calculate response time statistics
	responseTimes := make([]time.Duration, len(metrics))
	for i, metric := range metrics {
		responseTimes[i] = metric.ResponseTime
	}
	prt.calculateResponseTimeStats(responseTimes, &baseline.ResponseTime)

	// Calculate throughput statistics
	throughputs := make([]float64, len(metrics))
	for i, metric := range metrics {
		throughputs[i] = metric.Throughput
	}
	prt.calculateThroughputStats(throughputs, &baseline.Throughput)

	// Calculate error rate statistics
	errorRates := make([]float64, len(metrics))
	for i, metric := range metrics {
		errorRates[i] = metric.ErrorRate
	}
	prt.calculateErrorRateStats(errorRates, &baseline.ErrorRate)

	// Calculate resource utilization averages
	prt.calculateResourceUtilizationStats(metrics, &baseline.ResourceUtilization)

	baseline.SampleCount = len(metrics)

	// Store baseline
	prt.mu.Lock()
	prt.baselines[baseline.ID] = baseline
	prt.mu.Unlock()

	prt.logger.Info("created performance baseline",
		zap.String("baseline_id", baseline.ID),
		zap.String("endpoint", endpoint),
		zap.String("method", method),
		zap.Int("sample_count", baseline.SampleCount),
	)

	return baseline, nil
}

// TestRegression tests for performance regression against a baseline
func (prt *PerformanceRegressionTester) TestRegression(ctx context.Context, baselineID string, currentMetrics []RegressionPerformanceMetric) (*RegressionResult, error) {
	prt.mu.RLock()
	baseline, exists := prt.baselines[baselineID]
	prt.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("baseline not found: %s", baselineID)
	}

	if len(currentMetrics) < prt.config.MinimumSampleSize {
		return nil, fmt.Errorf("insufficient samples for regression test: got %d, need %d", len(currentMetrics), prt.config.MinimumSampleSize)
	}

	result := &RegressionResult{
		ID:              generateID(),
		BaselineID:      baselineID,
		Endpoint:        baseline.Endpoint,
		Method:          baseline.Method,
		TestedAt:        time.Now(),
		SampleCount:     len(currentMetrics),
		Recommendations: []string{},
	}

	// Test response time regression
	responseTimes := make([]time.Duration, len(currentMetrics))
	for i, metric := range currentMetrics {
		responseTimes[i] = metric.ResponseTime
	}
	result.ResponseTimeRegression = prt.testResponseTimeRegression(responseTimes, &baseline.ResponseTime)

	// Test throughput regression
	throughputs := make([]float64, len(currentMetrics))
	for i, metric := range currentMetrics {
		throughputs[i] = metric.Throughput
	}
	result.ThroughputRegression = prt.testThroughputRegression(throughputs, &baseline.Throughput)

	// Test error rate regression
	errorRates := make([]float64, len(currentMetrics))
	for i, metric := range currentMetrics {
		errorRates[i] = metric.ErrorRate
	}
	result.ErrorRateRegression = prt.testErrorRateRegression(errorRates, &baseline.ErrorRate)

	// Test resource utilization regression
	result.ResourceUtilizationRegression = prt.testResourceUtilizationRegression(currentMetrics, &baseline.ResourceUtilization)

	// Calculate overall regression assessment
	prt.calculateOverallRegression(result)

	// Generate recommendations
	prt.generateRecommendations(result)

	// Store result
	prt.mu.Lock()
	prt.results[result.ID] = result
	prt.mu.Unlock()

	prt.logger.Info("completed regression test",
		zap.String("result_id", result.ID),
		zap.String("baseline_id", baselineID),
		zap.Bool("has_regression", result.HasRegression),
		zap.String("severity", result.Severity),
		zap.Float64("score", result.Score),
	)

	return result, nil
}

// GetBaseline retrieves a baseline by ID
func (prt *PerformanceRegressionTester) GetBaseline(baselineID string) (*PerformanceBaseline, error) {
	prt.mu.RLock()
	defer prt.mu.RUnlock()

	baseline, exists := prt.baselines[baselineID]
	if !exists {
		return nil, fmt.Errorf("baseline not found: %s", baselineID)
	}

	return baseline, nil
}

// GetResult retrieves a regression test result by ID
func (prt *PerformanceRegressionTester) GetResult(resultID string) (*RegressionResult, error) {
	prt.mu.RLock()
	defer prt.mu.RUnlock()

	result, exists := prt.results[resultID]
	if !exists {
		return nil, fmt.Errorf("result not found: %s", resultID)
	}

	return result, nil
}

// ListBaselines returns all baselines
func (prt *PerformanceRegressionTester) ListBaselines() []*PerformanceBaseline {
	prt.mu.RLock()
	defer prt.mu.RUnlock()

	baselines := make([]*PerformanceBaseline, 0, len(prt.baselines))
	for _, baseline := range prt.baselines {
		baselines = append(baselines, baseline)
	}

	// Sort by creation date (newest first)
	sort.Slice(baselines, func(i, j int) bool {
		return baselines[i].CreatedAt.After(baselines[j].CreatedAt)
	})

	return baselines
}

// ListResults returns all regression test results
func (prt *PerformanceRegressionTester) ListResults() []*RegressionResult {
	prt.mu.RLock()
	defer prt.mu.RUnlock()

	results := make([]*RegressionResult, 0, len(prt.results))
	for _, result := range prt.results {
		results = append(results, result)
	}

	// Sort by test date (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].TestedAt.After(results[j].TestedAt)
	})

	return results
}

// UpdateBaseline updates an existing baseline with new metrics
func (prt *PerformanceRegressionTester) UpdateBaseline(ctx context.Context, baselineID string, newMetrics []RegressionPerformanceMetric) (*PerformanceBaseline, error) {
	prt.mu.Lock()
	baseline, exists := prt.baselines[baselineID]
	if !exists {
		prt.mu.Unlock()
		return nil, fmt.Errorf("baseline not found: %s", baselineID)
	}

	// Combine existing and new metrics (simplified - in practice, you'd want more sophisticated merging)
	// For now, we'll just update with the new metrics
	prt.mu.Unlock()

	updatedBaseline, err := prt.CreateBaseline(ctx, baseline.Endpoint, baseline.Method, newMetrics)
	if err != nil {
		return nil, err
	}

	// Update the ID to match the original
	updatedBaseline.ID = baselineID
	updatedBaseline.CreatedAt = baseline.CreatedAt
	updatedBaseline.UpdatedAt = time.Now()

	prt.mu.Lock()
	prt.baselines[baselineID] = updatedBaseline
	prt.mu.Unlock()

	prt.logger.Info("updated performance baseline",
		zap.String("baseline_id", baselineID),
		zap.Int("new_sample_count", updatedBaseline.SampleCount),
	)

	return updatedBaseline, nil
}

// Cleanup removes old baselines and results
func (prt *PerformanceRegressionTester) Cleanup() error {
	now := time.Now()
	baselineCutoff := now.AddDate(0, 0, -prt.config.BaselineRetentionDays)
	metricsCutoff := now.AddDate(0, 0, -prt.config.MetricsRetentionDays)

	prt.mu.Lock()
	defer prt.mu.Unlock()

	// Clean up old baselines
	baselineCount := 0
	for id, baseline := range prt.baselines {
		if baseline.UpdatedAt.Before(baselineCutoff) {
			delete(prt.baselines, id)
			baselineCount++
		}
	}

	// Clean up old results
	resultCount := 0
	for id, result := range prt.results {
		if result.TestedAt.Before(metricsCutoff) {
			delete(prt.results, id)
			resultCount++
		}
	}

	prt.logger.Info("cleaned up old data",
		zap.Int("baselines_removed", baselineCount),
		zap.Int("results_removed", resultCount),
	)

	return nil
}

// Shutdown gracefully shuts down the regression tester
func (prt *PerformanceRegressionTester) Shutdown() error {
	select {
	case <-prt.stopChan:
		// Already shut down
		return nil
	default:
		close(prt.stopChan)
	}
	return nil
}

// calculateResponseTimeStats calculates response time statistics
func (prt *PerformanceRegressionTester) calculateResponseTimeStats(responseTimes []time.Duration, stats *struct {
	P50  time.Duration `json:"p50"`
	P95  time.Duration `json:"p95"`
	P99  time.Duration `json:"p99"`
	Mean time.Duration `json:"mean"`
	Std  time.Duration `json:"std"`
}) {
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	n := len(responseTimes)
	if n == 0 {
		return
	}

	// Calculate percentiles
	stats.P50 = responseTimes[(n-1)*50/100]
	stats.P95 = responseTimes[(n-1)*95/100]
	stats.P99 = responseTimes[(n-1)*99/100]

	// Calculate mean
	total := time.Duration(0)
	for _, rt := range responseTimes {
		total += rt
	}
	stats.Mean = total / time.Duration(n)

	// Calculate standard deviation
	variance := float64(0)
	for _, rt := range responseTimes {
		diff := float64(rt - stats.Mean)
		variance += diff * diff
	}
	variance /= float64(n)
	stats.Std = time.Duration(math.Sqrt(variance))
}

// calculateThroughputStats calculates throughput statistics
func (prt *PerformanceRegressionTester) calculateThroughputStats(throughputs []float64, stats *struct {
	RequestsPerSecond float64 `json:"requests_per_second"`
	Mean              float64 `json:"mean"`
	Std               float64 `json:"std"`
}) {
	sort.Float64s(throughputs)

	n := len(throughputs)
	if n == 0 {
		return
	}

	// Calculate mean
	sum := 0.0
	for _, t := range throughputs {
		sum += t
	}
	stats.Mean = sum / float64(n)
	stats.RequestsPerSecond = stats.Mean

	// Calculate standard deviation
	variance := 0.0
	for _, t := range throughputs {
		diff := t - stats.Mean
		variance += diff * diff
	}
	variance /= float64(n)
	stats.Std = math.Sqrt(variance)
}

// calculateErrorRateStats calculates error rate statistics
func (prt *PerformanceRegressionTester) calculateErrorRateStats(errorRates []float64, stats *struct {
	Percentage float64 `json:"percentage"`
	Mean       float64 `json:"mean"`
	Std        float64 `json:"std"`
}) {
	sort.Float64s(errorRates)

	n := len(errorRates)
	if n == 0 {
		return
	}

	// Calculate mean
	sum := 0.0
	for _, er := range errorRates {
		sum += er
	}
	stats.Mean = sum / float64(n)
	stats.Percentage = stats.Mean

	// Calculate standard deviation
	variance := 0.0
	for _, er := range errorRates {
		diff := er - stats.Mean
		variance += diff * diff
	}
	variance /= float64(n)
	stats.Std = math.Sqrt(variance)
}

// calculateResourceUtilizationStats calculates resource utilization statistics
func (prt *PerformanceRegressionTester) calculateResourceUtilizationStats(metrics []RegressionPerformanceMetric, stats *struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Disk    float64 `json:"disk"`
	Network float64 `json:"network"`
}) {
	if len(metrics) == 0 {
		return
	}

	cpuSum, memorySum, diskSum, networkSum := 0.0, 0.0, 0.0, 0.0
	count := 0

	for _, metric := range metrics {
		if metric.ResourceUtilization != nil {
			cpuSum += metric.ResourceUtilization.CPU
			memorySum += metric.ResourceUtilization.Memory
			diskSum += metric.ResourceUtilization.Disk
			networkSum += metric.ResourceUtilization.Network
			count++
		}
	}

	if count > 0 {
		stats.CPU = cpuSum / float64(count)
		stats.Memory = memorySum / float64(count)
		stats.Disk = diskSum / float64(count)
		stats.Network = networkSum / float64(count)
	}
}

// testResponseTimeRegression tests for response time regression
func (prt *PerformanceRegressionTester) testResponseTimeRegression(currentResponseTimes []time.Duration, baseline *struct {
	P50  time.Duration `json:"p50"`
	P95  time.Duration `json:"p95"`
	P99  time.Duration `json:"p99"`
	Mean time.Duration `json:"mean"`
	Std  time.Duration `json:"std"`
}) *RegressionMetric {
	// Calculate current P95 response time
	sort.Slice(currentResponseTimes, func(i, j int) bool {
		return currentResponseTimes[i] < currentResponseTimes[j]
	})

	n := len(currentResponseTimes)
	if n == 0 {
		return nil
	}

	currentP95 := currentResponseTimes[(n-1)*95/100]
	currentMean := time.Duration(0)
	for _, rt := range currentResponseTimes {
		currentMean += rt
	}
	currentMean /= time.Duration(n)

	// Calculate change percentage
	changePercent := float64(currentP95-baseline.P95) / float64(baseline.P95) * 100

	// Determine if it's a regression
	isRegression := changePercent > prt.config.ResponseTimeThreshold

	// Calculate statistical significance (simplified t-test)
	pValue := prt.calculateTTest(currentResponseTimes, baseline.Mean, baseline.Std)
	isSignificant := pValue < prt.config.StatisticalSignificance

	// Calculate effect size (Cohen's d)
	effectSize := prt.calculateEffectSize(currentResponseTimes, baseline.Mean, baseline.Std)

	// Determine severity
	severity := prt.determineSeverity(changePercent, isSignificant)

	return &RegressionMetric{
		MetricName:    "response_time_p95",
		BaselineValue: float64(baseline.P95),
		CurrentValue:  float64(currentP95),
		ChangePercent: changePercent,
		IsRegression:  isRegression,
		IsSignificant: isSignificant,
		PValue:        pValue,
		EffectSize:    effectSize,
		Severity:      severity,
	}
}

// testThroughputRegression tests for throughput regression
func (prt *PerformanceRegressionTester) testThroughputRegression(currentThroughputs []float64, baseline *struct {
	RequestsPerSecond float64 `json:"requests_per_second"`
	Mean              float64 `json:"mean"`
	Std               float64 `json:"std"`
}) *RegressionMetric {
	if len(currentThroughputs) == 0 {
		return nil
	}

	// Calculate current mean throughput
	currentMean := 0.0
	for _, t := range currentThroughputs {
		currentMean += t
	}
	currentMean /= float64(len(currentThroughputs))

	// Calculate change percentage (negative means regression)
	changePercent := (currentMean - baseline.Mean) / baseline.Mean * 100

	// Determine if it's a regression (throughput decrease)
	isRegression := changePercent < -prt.config.ThroughputThreshold

	// Calculate statistical significance
	pValue := prt.calculateTTestFloat(currentThroughputs, baseline.Mean, baseline.Std)
	isSignificant := pValue < prt.config.StatisticalSignificance

	// Calculate effect size
	effectSize := prt.calculateEffectSizeFloat(currentThroughputs, baseline.Mean, baseline.Std)

	// Determine severity (use absolute value for severity calculation)
	severity := prt.determineSeverity(math.Abs(changePercent), isSignificant)

	return &RegressionMetric{
		MetricName:    "throughput",
		BaselineValue: baseline.Mean,
		CurrentValue:  currentMean,
		ChangePercent: changePercent,
		IsRegression:  isRegression,
		IsSignificant: isSignificant,
		PValue:        pValue,
		EffectSize:    effectSize,
		Severity:      severity,
	}
}

// testErrorRateRegression tests for error rate regression
func (prt *PerformanceRegressionTester) testErrorRateRegression(currentErrorRates []float64, baseline *struct {
	Percentage float64 `json:"percentage"`
	Mean       float64 `json:"mean"`
	Std        float64 `json:"std"`
}) *RegressionMetric {
	if len(currentErrorRates) == 0 {
		return nil
	}

	// Calculate current mean error rate
	currentMean := 0.0
	for _, er := range currentErrorRates {
		currentMean += er
	}
	currentMean /= float64(len(currentErrorRates))

	// Calculate change percentage
	changePercent := (currentMean - baseline.Mean) / baseline.Mean * 100

	// Determine if it's a regression (error rate increase)
	isRegression := changePercent > prt.config.ErrorRateThreshold

	// Calculate statistical significance
	pValue := prt.calculateTTestFloat(currentErrorRates, baseline.Mean, baseline.Std)
	isSignificant := pValue < prt.config.StatisticalSignificance

	// Calculate effect size
	effectSize := prt.calculateEffectSizeFloat(currentErrorRates, baseline.Mean, baseline.Std)

	// Determine severity
	severity := prt.determineSeverity(changePercent, isSignificant)

	return &RegressionMetric{
		MetricName:    "error_rate",
		BaselineValue: baseline.Mean,
		CurrentValue:  currentMean,
		ChangePercent: changePercent,
		IsRegression:  isRegression,
		IsSignificant: isSignificant,
		PValue:        pValue,
		EffectSize:    effectSize,
		Severity:      severity,
	}
}

// testResourceUtilizationRegression tests for resource utilization regression
func (prt *PerformanceRegressionTester) testResourceUtilizationRegression(currentMetrics []RegressionPerformanceMetric, baseline *struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Disk    float64 `json:"disk"`
	Network float64 `json:"network"`
}) *RegressionMetric {
	if len(currentMetrics) == 0 {
		return nil
	}

	// Calculate current average CPU utilization
	cpuSum := 0.0
	count := 0
	for _, metric := range currentMetrics {
		if metric.ResourceUtilization != nil {
			cpuSum += metric.ResourceUtilization.CPU
			count++
		}
	}

	if count == 0 {
		return nil
	}

	currentCPU := cpuSum / float64(count)

	// Calculate change percentage
	changePercent := (currentCPU - baseline.CPU) / baseline.CPU * 100

	// Determine if it's a regression (CPU increase)
	isRegression := changePercent > 10.0 // 10% increase threshold

	// For resource utilization, we'll use a simplified significance test
	isSignificant := math.Abs(changePercent) > 5.0 // 5% threshold

	// Calculate effect size
	effectSize := math.Abs(changePercent) / 10.0 // Normalized effect size

	// Determine severity
	severity := prt.determineSeverity(changePercent, isSignificant)

	return &RegressionMetric{
		MetricName:    "cpu_utilization",
		BaselineValue: baseline.CPU,
		CurrentValue:  currentCPU,
		ChangePercent: changePercent,
		IsRegression:  isRegression,
		IsSignificant: isSignificant,
		PValue:        0.05, // Simplified
		EffectSize:    effectSize,
		Severity:      severity,
	}
}

// calculateOverallRegression calculates the overall regression assessment
func (prt *PerformanceRegressionTester) calculateOverallRegression(result *RegressionResult) {
	regressionCount := 0
	totalScore := 0.0
	maxScore := 0.0

	// Check each metric for regression
	metrics := []*RegressionMetric{
		result.ResponseTimeRegression,
		result.ThroughputRegression,
		result.ErrorRateRegression,
		result.ResourceUtilizationRegression,
	}

	for _, metric := range metrics {
		if metric != nil {
			maxScore += 25.0 // Each metric contributes up to 25 points
			if metric.IsRegression {
				regressionCount++
				// Calculate score based on severity and significance
				score := 25.0 * (math.Abs(metric.ChangePercent) / 100.0)
				if metric.IsSignificant {
					score *= 1.5 // Boost score for statistically significant regressions
				}
				totalScore += math.Min(score, 25.0)
			}
		}
	}

	result.HasRegression = regressionCount > 0
	result.Score = totalScore

	// Determine overall severity
	if result.Score >= 75.0 {
		result.Severity = "critical"
	} else if result.Score >= 50.0 {
		result.Severity = "high"
	} else if result.Score >= 25.0 {
		result.Severity = "medium"
	} else if result.Score >= 10.0 {
		result.Severity = "low"
	} else {
		result.Severity = "none"
	}
}

// generateRecommendations generates recommendations based on regression results
func (prt *PerformanceRegressionTester) generateRecommendations(result *RegressionResult) {
	if result.ResponseTimeRegression != nil && result.ResponseTimeRegression.IsRegression {
		if result.ResponseTimeRegression.ChangePercent > 50.0 {
			result.Recommendations = append(result.Recommendations, "Critical response time regression detected. Immediate investigation required.")
		} else if result.ResponseTimeRegression.ChangePercent > 20.0 {
			result.Recommendations = append(result.Recommendations, "Significant response time regression. Review recent code changes and database queries.")
		} else {
			result.Recommendations = append(result.Recommendations, "Minor response time regression. Monitor closely and investigate if trend continues.")
		}
	}

	if result.ThroughputRegression != nil && result.ThroughputRegression.IsRegression {
		result.Recommendations = append(result.Recommendations, "Throughput regression detected. Check for resource bottlenecks and optimization opportunities.")
	}

	if result.ErrorRateRegression != nil && result.ErrorRateRegression.IsRegression {
		result.Recommendations = append(result.Recommendations, "Error rate increase detected. Review error logs and investigate root causes.")
	}

	if result.ResourceUtilizationRegression != nil && result.ResourceUtilizationRegression.IsRegression {
		result.Recommendations = append(result.Recommendations, "Resource utilization increase detected. Consider scaling or optimization.")
	}

	if len(result.Recommendations) == 0 {
		result.Recommendations = append(result.Recommendations, "No significant regressions detected. Continue monitoring.")
	}
}

// calculateTTest calculates a simplified t-test p-value
func (prt *PerformanceRegressionTester) calculateTTest(samples []time.Duration, baselineMean, baselineStd time.Duration) float64 {
	if len(samples) < 2 {
		return 1.0
	}

	// Calculate sample mean
	sampleMean := time.Duration(0)
	for _, sample := range samples {
		sampleMean += sample
	}
	sampleMean /= time.Duration(len(samples))

	// Calculate sample standard deviation
	sampleVariance := 0.0
	for _, sample := range samples {
		diff := float64(sample - sampleMean)
		sampleVariance += diff * diff
	}
	sampleVariance /= float64(len(samples) - 1)
	_ = math.Sqrt(sampleVariance) // sampleStd calculated but not used in simplified t-test

	// Calculate t-statistic (simplified)
	tStat := math.Abs(float64(sampleMean-baselineMean)) / (math.Sqrt(sampleVariance) / math.Sqrt(float64(len(samples))))

	// Simplified p-value calculation (for large samples, t-distribution approaches normal)
	// This is a rough approximation
	if tStat > 3.0 {
		return 0.001
	} else if tStat > 2.0 {
		return 0.05
	} else if tStat > 1.5 {
		return 0.1
	}
	return 0.5
}

// calculateTTestFloat calculates a simplified t-test p-value for float64 samples
func (prt *PerformanceRegressionTester) calculateTTestFloat(samples []float64, baselineMean, baselineStd float64) float64 {
	if len(samples) < 2 {
		return 1.0
	}

	// Calculate sample mean
	sampleMean := 0.0
	for _, sample := range samples {
		sampleMean += sample
	}
	sampleMean /= float64(len(samples))

	// Calculate sample standard deviation
	sampleVariance := 0.0
	for _, sample := range samples {
		diff := sample - sampleMean
		sampleVariance += diff * diff
	}
	sampleVariance /= float64(len(samples) - 1)
	_ = math.Sqrt(sampleVariance) // sampleStd calculated but not used in simplified t-test

	// Calculate t-statistic (simplified)
	tStat := math.Abs(sampleMean-baselineMean) / (math.Sqrt(sampleVariance) / math.Sqrt(float64(len(samples))))

	// Simplified p-value calculation
	if tStat > 3.0 {
		return 0.001
	} else if tStat > 2.0 {
		return 0.05
	} else if tStat > 1.5 {
		return 0.1
	}
	return 0.5
}

// calculateEffectSize calculates Cohen's d effect size
func (prt *PerformanceRegressionTester) calculateEffectSize(samples []time.Duration, baselineMean, baselineStd time.Duration) float64 {
	if len(samples) == 0 {
		return 0.0
	}

	// Calculate sample mean
	sampleMean := time.Duration(0)
	for _, sample := range samples {
		sampleMean += sample
	}
	sampleMean /= time.Duration(len(samples))

	// Calculate pooled standard deviation
	sampleVariance := 0.0
	for _, sample := range samples {
		diff := float64(sample - sampleMean)
		sampleVariance += diff * diff
	}
	sampleVariance /= float64(len(samples) - 1)
	sampleStd := math.Sqrt(sampleVariance)

	// Pooled standard deviation
	pooledStd := math.Sqrt((float64(baselineStd*baselineStd) + sampleStd*sampleStd) / 2.0)

	// Cohen's d
	return math.Abs(float64(sampleMean-baselineMean)) / pooledStd
}

// calculateEffectSizeFloat calculates Cohen's d effect size for float64 samples
func (prt *PerformanceRegressionTester) calculateEffectSizeFloat(samples []float64, baselineMean, baselineStd float64) float64 {
	if len(samples) == 0 {
		return 0.0
	}

	// Calculate sample mean
	sampleMean := 0.0
	for _, sample := range samples {
		sampleMean += sample
	}
	sampleMean /= float64(len(samples))

	// Calculate sample standard deviation
	sampleVariance := 0.0
	for _, sample := range samples {
		diff := sample - sampleMean
		sampleVariance += diff * diff
	}
	sampleVariance /= float64(len(samples) - 1)
	sampleStd := math.Sqrt(sampleVariance)

	// Pooled standard deviation
	pooledStd := math.Sqrt((baselineStd*baselineStd + sampleStd*sampleStd) / 2.0)

	// Cohen's d
	return math.Abs(sampleMean-baselineMean) / pooledStd
}

// determineSeverity determines the severity level based on change percentage and significance
func (prt *PerformanceRegressionTester) determineSeverity(changePercent float64, isSignificant bool) string {
	absChange := math.Abs(changePercent)

	if absChange > 50.0 {
		return "critical"
	} else if absChange > 20.0 {
		return "high"
	} else if absChange > 10.0 {
		return "medium"
	} else if absChange > 5.0 && isSignificant {
		return "low"
	}
	return "none"
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("reg_%d", time.Now().UnixNano())
}
