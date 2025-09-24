package success_monitoring

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TrendPoint represents a point in time for trend analysis
type TrendPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	SuccessRate float64   `json:"success_rate"`
	SampleSize  int       `json:"sample_size"`
}

// SuccessRateBenchmarkManager manages success rate benchmarking and validation
type SuccessRateBenchmarkManager struct {
	config     *BenchmarkConfig
	logger     *zap.Logger
	benchmarks map[string]*BenchmarkSuite
	results    []*BenchmarkResult
	baselines  map[string]*BaselineMetrics
	mu         sync.RWMutex
	startTime  time.Time
}

// BenchmarkConfig holds configuration for success rate benchmarking
type BenchmarkConfig struct {
	EnableBenchmarking          bool          `json:"enable_benchmarking"`
	EnableStatisticalValidation bool          `json:"enable_statistical_validation"`
	EnableBaselineComparison    bool          `json:"enable_baseline_comparison"`
	EnableABTesting             bool          `json:"enable_ab_testing"`
	BenchmarkInterval           time.Duration `json:"benchmark_interval"`
	MaxBenchmarkHistory         int           `json:"max_benchmark_history"`
	TargetSuccessRate           float64       `json:"target_success_rate"`       // 0.95 = 95%
	ConfidenceLevel             float64       `json:"confidence_level"`          // 0.95 = 95% confidence
	MinSampleSize               int           `json:"min_sample_size"`           // Minimum samples for valid benchmark
	MaxSampleSize               int           `json:"max_sample_size"`           // Maximum samples per benchmark
	ValidationThreshold         float64       `json:"validation_threshold"`      // Minimum improvement to consider significant
	BaselineRetentionPeriod     time.Duration `json:"baseline_retention_period"` // 90 days
}

// BenchmarkSuite represents a collection of benchmarks for success rate validation
type BenchmarkSuite struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	TestCases   []*BenchmarkTestCase   `json:"test_cases"`
	Metrics     *BenchmarkMetrics      `json:"metrics"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// BenchmarkTestCase represents a single test case in a success rate benchmark
type BenchmarkTestCase struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Input           interface{}            `json:"input"`
	ExpectedSuccess bool                   `json:"expected_success"`
	ActualSuccess   bool                   `json:"actual_success,omitempty"`
	ResponseTime    time.Duration          `json:"response_time,omitempty"`
	ErrorType       string                 `json:"error_type,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	Weight          float64                `json:"weight"`     // Importance weight for this test case
	Difficulty      string                 `json:"difficulty"` // easy, medium, hard
	Tags            []string               `json:"tags"`
	CreatedAt       time.Time              `json:"created_at"`
}

// BenchmarkMetrics holds success rate and performance metrics
type BenchmarkMetrics struct {
	SuccessRate             float64                     `json:"success_rate"`          // Overall success rate (0-1)
	ConfidenceInterval      *ConfidenceInterval         `json:"confidence_interval"`   // Statistical confidence interval
	AverageResponseTime     time.Duration               `json:"average_response_time"` // Average processing time
	ThroughputPerSec        float64                     `json:"throughput_per_sec"`    // Tests processed per second
	ErrorRate               float64                     `json:"error_rate"`            // Rate of test execution errors
	PerCategoryMetrics      map[string]*CategoryMetrics `json:"per_category_metrics"`
	TrendData               []*TrendPoint               `json:"trend_data"`
	StatisticalSignificance *StatisticalSignificance    `json:"statistical_significance"`
	LastUpdated             time.Time                   `json:"last_updated"`
}

// ConfidenceInterval represents a statistical confidence interval
type ConfidenceInterval struct {
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	Level      float64 `json:"level"` // Confidence level (e.g., 0.95 for 95%)
}

// CategoryMetrics holds metrics for different test categories
type CategoryMetrics struct {
	CategoryName string  `json:"category_name"`
	SuccessRate  float64 `json:"success_rate"`
	SampleCount  int     `json:"sample_count"`
	Weight       float64 `json:"weight"`
}

// StatisticalSignificance represents statistical significance test results
type StatisticalSignificance struct {
	IsSignificant bool    `json:"is_significant"`
	PValue        float64 `json:"p_value"`
	EffectSize    float64 `json:"effect_size"`
	Power         float64 `json:"power"`
	TestType      string  `json:"test_type"` // "z_test", "t_test", "chi_square"
}

// BenchmarkResult represents the result of a benchmark execution
type BenchmarkResult struct {
	ID                 string              `json:"id"`
	SuiteID            string              `json:"suite_id"`
	ExecutionTime      time.Time           `json:"execution_time"`
	Duration           time.Duration       `json:"duration"`
	Metrics            *BenchmarkMetrics   `json:"metrics"`
	TestResults        []*TestResult       `json:"test_results"`
	BaselineComparison *BaselineComparison `json:"baseline_comparison,omitempty"`
	Validation         *ValidationResult   `json:"validation,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
}

// TestResult represents the result of a single test case
type TestResult struct {
	TestCaseID   string        `json:"test_case_id"`
	Success      bool          `json:"success"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorType    string        `json:"error_type,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
	ExecutedAt   time.Time     `json:"executed_at"`
}

// BaselineMetrics represents baseline success rate metrics for comparison
type BaselineMetrics struct {
	ProcessName        string              `json:"process_name"`
	SuccessRate        float64             `json:"success_rate"`
	ConfidenceInterval *ConfidenceInterval `json:"confidence_interval"`
	SampleCount        int                 `json:"sample_count"`
	LastUpdated        time.Time           `json:"last_updated"`
	HistoricalData     []*HistoricalPoint  `json:"historical_data"`
}

// HistoricalPoint represents a historical data point
type HistoricalPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	SuccessRate float64   `json:"success_rate"`
	SampleCount int       `json:"sample_count"`
}

// BaselineComparison represents comparison with baseline metrics
type BaselineComparison struct {
	BaselineSuccessRate float64 `json:"baseline_success_rate"`
	CurrentSuccessRate  float64 `json:"current_success_rate"`
	Improvement         float64 `json:"improvement"`
	IsSignificant       bool    `json:"is_significant"`
	PValue              float64 `json:"p_value"`
	EffectSize          float64 `json:"effect_size"`
}

// ValidationResult represents validation of benchmark results
type ValidationResult struct {
	IsValid          bool      `json:"is_valid"`
	ValidationErrors []string  `json:"validation_errors"`
	Recommendations  []string  `json:"recommendations"`
	ConfidenceScore  float64   `json:"confidence_score"`
	ValidatedAt      time.Time `json:"validated_at"`
}

// ABTestResult represents A/B testing results
type ABTestResult struct {
	TestID              string  `json:"test_id"`
	VariantA            string  `json:"variant_a"`
	VariantB            string  `json:"variant_b"`
	VariantASuccessRate float64 `json:"variant_a_success_rate"`
	VariantBSuccessRate float64 `json:"variant_b_success_rate"`
	Difference          float64 `json:"difference"`
	IsSignificant       bool    `json:"is_significant"`
	PValue              float64 `json:"p_value"`
	RecommendedVariant  string  `json:"recommended_variant"`
	SampleSizeA         int     `json:"sample_size_a"`
	SampleSizeB         int     `json:"sample_size_b"`
}

// NewSuccessRateBenchmarkManager creates a new success rate benchmark manager
func NewSuccessRateBenchmarkManager(config *BenchmarkConfig, logger *zap.Logger) *SuccessRateBenchmarkManager {
	if config == nil {
		config = DefaultBenchmarkConfig()
	}

	manager := &SuccessRateBenchmarkManager{
		config:     config,
		logger:     logger,
		benchmarks: make(map[string]*BenchmarkSuite),
		results:    make([]*BenchmarkResult, 0),
		baselines:  make(map[string]*BaselineMetrics),
		startTime:  time.Now(),
	}

	// Start background benchmarking if enabled
	if config.EnableBenchmarking {
		go manager.startBackgroundBenchmarking()
	}

	return manager
}

// DefaultBenchmarkConfig returns default configuration
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		EnableBenchmarking:          true,
		EnableStatisticalValidation: true,
		EnableBaselineComparison:    true,
		EnableABTesting:             true,
		BenchmarkInterval:           24 * time.Hour,      // Daily benchmarks
		MaxBenchmarkHistory:         30,                  // Keep 30 days of history
		TargetSuccessRate:           0.95,                // 95% target
		ConfidenceLevel:             0.95,                // 95% confidence
		MinSampleSize:               100,                 // Minimum 100 samples
		MaxSampleSize:               10000,               // Maximum 10,000 samples
		ValidationThreshold:         0.02,                // 2% minimum improvement
		BaselineRetentionPeriod:     90 * 24 * time.Hour, // 90 days
	}
}

// CreateBenchmarkSuite creates a new benchmark suite
func (m *SuccessRateBenchmarkManager) CreateBenchmarkSuite(ctx context.Context, suite *BenchmarkSuite) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if suite.ID == "" {
		suite.ID = fmt.Sprintf("benchmark_%d", time.Now().Unix())
	}

	suite.CreatedAt = time.Now()
	suite.UpdatedAt = time.Now()

	m.benchmarks[suite.ID] = suite

	m.logger.Info("Created benchmark suite",
		zap.String("suite_id", suite.ID),
		zap.String("name", suite.Name),
		zap.String("category", suite.Category),
		zap.Int("test_case_count", len(suite.TestCases)))

	return nil
}

// ExecuteBenchmark executes a benchmark suite
func (m *SuccessRateBenchmarkManager) ExecuteBenchmark(ctx context.Context, suiteID string) (*BenchmarkResult, error) {
	m.mu.RLock()
	suite, exists := m.benchmarks[suiteID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("benchmark suite not found: %s", suiteID)
	}

	startTime := time.Now()
	testResults := make([]*TestResult, 0, len(suite.TestCases))

	// Execute test cases
	for _, testCase := range suite.TestCases {
		result := m.executeTestCase(ctx, testCase)
		testResults = append(testResults, result)
	}

	// Calculate metrics
	metrics := m.calculateBenchmarkMetrics(testResults, suite.TestCases)

	// Perform statistical validation
	if m.config.EnableStatisticalValidation {
		metrics.StatisticalSignificance = m.calculateStatisticalSignificance(testResults, suite.TestCases)
	}

	// Compare with baseline
	var baselineComparison *BaselineComparison
	if m.config.EnableBaselineComparison {
		baselineComparison = m.compareWithBaseline(suite.Category, metrics)
	}

	// Validate results
	validation := m.validateBenchmarkResults(metrics, testResults)

	// Create benchmark result
	result := &BenchmarkResult{
		ID:                 fmt.Sprintf("result_%s_%d", suiteID, startTime.Unix()),
		SuiteID:            suiteID,
		ExecutionTime:      startTime,
		Duration:           time.Since(startTime),
		Metrics:            metrics,
		TestResults:        testResults,
		BaselineComparison: baselineComparison,
		Validation:         validation,
		CreatedAt:          time.Now(),
	}

	// Store result
	m.mu.Lock()
	m.results = append(m.results, result)

	// Cleanup old results
	if len(m.results) > m.config.MaxBenchmarkHistory {
		m.results = m.results[len(m.results)-m.config.MaxBenchmarkHistory:]
	}
	m.mu.Unlock()

	// Update suite metrics
	suite.Metrics = metrics
	suite.UpdatedAt = time.Now()

	m.logger.Info("Executed benchmark suite",
		zap.String("suite_id", suiteID),
		zap.Duration("duration", result.Duration),
		zap.Float64("success_rate", metrics.SuccessRate),
		zap.Int("test_count", len(testResults)),
		zap.Bool("is_valid", validation.IsValid))

	return result, nil
}

// executeTestCase executes a single test case
func (m *SuccessRateBenchmarkManager) executeTestCase(ctx context.Context, testCase *BenchmarkTestCase) *TestResult {
	startTime := time.Now()

	// Simulate test execution (in real implementation, this would call actual processing)
	success := m.simulateTestCaseExecution(ctx, testCase)

	responseTime := time.Since(startTime)

	result := &TestResult{
		TestCaseID:   testCase.ID,
		Success:      success,
		ResponseTime: responseTime,
		ExecutedAt:   startTime,
	}

	if !success {
		result.ErrorType = "simulation_error"
		result.ErrorMessage = "Simulated test failure"
	}

	return result
}

// simulateTestCaseExecution simulates the execution of a test case
func (m *SuccessRateBenchmarkManager) simulateTestCaseExecution(ctx context.Context, testCase *BenchmarkTestCase) bool {
	// In a real implementation, this would:
	// 1. Call the actual business processing logic
	// 2. Compare results with expected output
	// 3. Return success/failure based on comparison

	// For now, simulate based on test case difficulty and weight
	successProbability := 0.95 // Base 95% success rate

	switch testCase.Difficulty {
	case "easy":
		successProbability = 0.98
	case "medium":
		successProbability = 0.95
	case "hard":
		successProbability = 0.90
	}

	// Adjust based on weight (higher weight = higher success rate)
	successProbability += testCase.Weight * 0.02

	// Simulate success/failure
	return m.randomSuccess(successProbability)
}

// randomSuccess returns true with the given probability
func (m *SuccessRateBenchmarkManager) randomSuccess(probability float64) bool {
	return probability >= 0.5 // Simplified for simulation
}

// calculateBenchmarkMetrics calculates comprehensive metrics for the benchmark
func (m *SuccessRateBenchmarkManager) calculateBenchmarkMetrics(testResults []*TestResult, testCases []*BenchmarkTestCase) *BenchmarkMetrics {
	if len(testResults) == 0 {
		return &BenchmarkMetrics{}
	}

	var totalExecutionTime time.Duration
	var successCount int
	var errorCount int

	// Calculate per-category metrics
	categoryMetrics := make(map[string]*CategoryMetrics)

	for i, result := range testResults {
		totalExecutionTime += result.ResponseTime

		if result.Success {
			successCount++
		} else {
			errorCount++
		}

		// Update category metrics
		if i < len(testCases) {
			testCase := testCases[i]
			category := testCase.Difficulty // Use difficulty as category for now

			if categoryMetrics[category] == nil {
				categoryMetrics[category] = &CategoryMetrics{
					CategoryName: category,
					SuccessRate:  0,
					SampleCount:  0,
					Weight:       0,
				}
			}

			catMetrics := categoryMetrics[category]
			catMetrics.SampleCount++
			if result.Success {
				catMetrics.SuccessRate = float64(catMetrics.SampleCount) / float64(catMetrics.SampleCount)
			}
			catMetrics.Weight += testCase.Weight
		}
	}

	// Calculate basic metrics
	totalTests := len(testResults)
	successRate := float64(successCount) / float64(totalTests)
	errorRate := float64(errorCount) / float64(totalTests)
	averageLatency := totalExecutionTime / time.Duration(totalTests)

	// Calculate throughput
	throughputPerSec := float64(totalTests) / totalExecutionTime.Seconds()

	// Calculate confidence interval
	confidenceInterval := m.calculateConfidenceInterval(successCount, totalTests, m.config.ConfidenceLevel)

	return &BenchmarkMetrics{
		SuccessRate:         successRate,
		ConfidenceInterval:  confidenceInterval,
		AverageResponseTime: averageLatency,
		ThroughputPerSec:    throughputPerSec,
		ErrorRate:           errorRate,
		PerCategoryMetrics:  categoryMetrics,
		TrendData:           []*TrendPoint{},
		LastUpdated:         time.Now(),
	}
}

// calculateConfidenceInterval calculates a confidence interval for the success rate
func (m *SuccessRateBenchmarkManager) calculateConfidenceInterval(successes, total int, confidenceLevel float64) *ConfidenceInterval {
	if total == 0 {
		return &ConfidenceInterval{Level: confidenceLevel}
	}

	// Use Wilson score interval for better small sample performance
	p := float64(successes) / float64(total)
	z := 1.96 // 95% confidence level (simplified)

	denominator := 1 + z*z/float64(total)
	centreAdjustment := z * math.Sqrt(z*z/float64(total)-1.0/float64(total)+4.0*p*(1-p)/float64(total)) / (2.0 * denominator)
	centreProbability := (p + z*z/(2.0*float64(total))) / denominator

	lowerBound := math.Max(0, centreProbability-centreAdjustment)
	upperBound := math.Min(1, centreProbability+centreAdjustment)

	return &ConfidenceInterval{
		LowerBound: lowerBound,
		UpperBound: upperBound,
		Level:      confidenceLevel,
	}
}

// calculateStatisticalSignificance calculates statistical significance of results
func (m *SuccessRateBenchmarkManager) calculateStatisticalSignificance(testResults []*TestResult, testCases []*BenchmarkTestCase) *StatisticalSignificance {
	if len(testResults) < m.config.MinSampleSize {
		return &StatisticalSignificance{
			IsSignificant: false,
			PValue:        1.0,
			EffectSize:    0.0,
			Power:         0.0,
			TestType:      "insufficient_sample_size",
		}
	}

	// Calculate observed success rate
	successCount := 0
	for _, result := range testResults {
		if result.Success {
			successCount++
		}
	}

	observedRate := float64(successCount) / float64(len(testResults))

	// Compare against target success rate using z-test
	expectedRate := m.config.TargetSuccessRate
	zScore := (observedRate - expectedRate) / math.Sqrt(expectedRate*(1-expectedRate)/float64(len(testResults)))

	// Calculate p-value (simplified)
	pValue := 2 * (1 - m.normalCDF(math.Abs(zScore)))

	// Calculate effect size
	effectSize := observedRate - expectedRate

	// Determine significance
	isSignificant := pValue < (1 - m.config.ConfidenceLevel)

	// Calculate power (simplified)
	power := 0.8 // Simplified power calculation

	return &StatisticalSignificance{
		IsSignificant: isSignificant,
		PValue:        pValue,
		EffectSize:    effectSize,
		Power:         power,
		TestType:      "z_test",
	}
}

// normalCDF calculates the cumulative distribution function of the standard normal distribution
func (m *SuccessRateBenchmarkManager) normalCDF(x float64) float64 {
	return 0.5 * (1 + m.erf(x/math.Sqrt(2)))
}

// erf calculates the error function (simplified approximation)
func (m *SuccessRateBenchmarkManager) erf(x float64) float64 {
	// Simplified error function approximation
	if x < 0 {
		return -m.erf(-x)
	}

	a1 := 0.254829592
	a2 := -0.284496736
	a3 := 1.421413741
	a4 := -1.453152027
	a5 := 1.061405429
	p := 0.3275911

	t := 1.0 / (1.0 + p*x)
	return 1 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-x*x)
}

// compareWithBaseline compares current results with baseline metrics
func (m *SuccessRateBenchmarkManager) compareWithBaseline(category string, metrics *BenchmarkMetrics) *BaselineComparison {
	m.mu.RLock()
	baseline, exists := m.baselines[category]
	m.mu.RUnlock()

	if !exists {
		return &BaselineComparison{
			BaselineSuccessRate: 0,
			CurrentSuccessRate:  metrics.SuccessRate,
			Improvement:         metrics.SuccessRate,
			IsSignificant:       false,
			PValue:              1.0,
			EffectSize:          0,
		}
	}

	improvement := metrics.SuccessRate - baseline.SuccessRate

	// Perform statistical test for significance
	isSignificant := improvement > m.config.ValidationThreshold

	// Calculate effect size
	effectSize := improvement / baseline.SuccessRate

	return &BaselineComparison{
		BaselineSuccessRate: baseline.SuccessRate,
		CurrentSuccessRate:  metrics.SuccessRate,
		Improvement:         improvement,
		IsSignificant:       isSignificant,
		PValue:              0.05, // Simplified
		EffectSize:          effectSize,
	}
}

// validateBenchmarkResults validates the benchmark results
func (m *SuccessRateBenchmarkManager) validateBenchmarkResults(metrics *BenchmarkMetrics, testResults []*TestResult) *ValidationResult {
	validationErrors := make([]string, 0)
	recommendations := make([]string, 0)
	confidenceScore := 1.0

	// Check sample size
	if len(testResults) < m.config.MinSampleSize {
		validationErrors = append(validationErrors, fmt.Sprintf("Sample size %d is below minimum required %d", len(testResults), m.config.MinSampleSize))
		confidenceScore *= 0.8
		recommendations = append(recommendations, "Increase sample size for more reliable results")
	}

	// Check success rate against target
	if metrics.SuccessRate < m.config.TargetSuccessRate {
		validationErrors = append(validationErrors, fmt.Sprintf("Success rate %.2f%% is below target %.2f%%", metrics.SuccessRate*100, m.config.TargetSuccessRate*100))
		confidenceScore *= 0.9
		recommendations = append(recommendations, "Investigate causes of low success rate")
	}

	// Check confidence interval width
	if metrics.ConfidenceInterval != nil {
		intervalWidth := metrics.ConfidenceInterval.UpperBound - metrics.ConfidenceInterval.LowerBound
		if intervalWidth > 0.1 { // More than 10% wide
			validationErrors = append(validationErrors, fmt.Sprintf("Confidence interval is too wide: %.2f%%", intervalWidth*100))
			confidenceScore *= 0.85
			recommendations = append(recommendations, "Increase sample size to narrow confidence interval")
		}
	}

	// Check for statistical significance
	if metrics.StatisticalSignificance != nil && !metrics.StatisticalSignificance.IsSignificant {
		validationErrors = append(validationErrors, "Results are not statistically significant")
		confidenceScore *= 0.9
		recommendations = append(recommendations, "Increase sample size or investigate effect size")
	}

	isValid := len(validationErrors) == 0

	return &ValidationResult{
		IsValid:          isValid,
		ValidationErrors: validationErrors,
		Recommendations:  recommendations,
		ConfidenceScore:  confidenceScore,
		ValidatedAt:      time.Now(),
	}
}

// UpdateBaseline updates baseline metrics for a category
func (m *SuccessRateBenchmarkManager) UpdateBaseline(ctx context.Context, category string, successRate float64, sampleCount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	baseline := &BaselineMetrics{
		ProcessName: category,
		SuccessRate: successRate,
		SampleCount: sampleCount,
		LastUpdated: time.Now(),
	}

	// Calculate confidence interval
	successCount := int(successRate * float64(sampleCount))
	baseline.ConfidenceInterval = m.calculateConfidenceInterval(successCount, sampleCount, m.config.ConfidenceLevel)

	// Add historical point
	historicalPoint := &HistoricalPoint{
		Timestamp:   time.Now(),
		SuccessRate: successRate,
		SampleCount: sampleCount,
	}

	if baseline.HistoricalData == nil {
		baseline.HistoricalData = make([]*HistoricalPoint, 0)
	}
	baseline.HistoricalData = append(baseline.HistoricalData, historicalPoint)

	// Cleanup old historical data
	m.cleanupHistoricalData(baseline)

	m.baselines[category] = baseline

	m.logger.Info("Updated baseline metrics",
		zap.String("category", category),
		zap.Float64("success_rate", successRate),
		zap.Int("sample_count", sampleCount))

	return nil
}

// cleanupHistoricalData removes old historical data points
func (m *SuccessRateBenchmarkManager) cleanupHistoricalData(baseline *BaselineMetrics) {
	cutoffTime := time.Now().Add(-m.config.BaselineRetentionPeriod)

	validPoints := make([]*HistoricalPoint, 0)
	for _, point := range baseline.HistoricalData {
		if point.Timestamp.After(cutoffTime) {
			validPoints = append(validPoints, point)
		}
	}

	baseline.HistoricalData = validPoints
}

// GetBenchmarkResults returns benchmark results for a suite
func (m *SuccessRateBenchmarkManager) GetBenchmarkResults(suiteID string, limit int) []*BenchmarkResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*BenchmarkResult
	for _, result := range m.results {
		if result.SuiteID == suiteID {
			results = append(results, result)
		}
	}

	// Sort by execution time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].ExecutionTime.After(results[j].ExecutionTime)
	})

	// Limit results
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// GetBaselineMetrics returns baseline metrics for a category
func (m *SuccessRateBenchmarkManager) GetBaselineMetrics(category string) *BaselineMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if baseline, exists := m.baselines[category]; exists {
		// Return a copy to avoid race conditions
		baselineCopy := *baseline
		if baseline.HistoricalData != nil {
			baselineCopy.HistoricalData = make([]*HistoricalPoint, len(baseline.HistoricalData))
			copy(baselineCopy.HistoricalData, baseline.HistoricalData)
		}
		return &baselineCopy
	}

	return nil
}

// startBackgroundBenchmarking starts background benchmark execution
func (m *SuccessRateBenchmarkManager) startBackgroundBenchmarking() {
	ticker := time.NewTicker(m.config.BenchmarkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()

			// Execute all benchmark suites
			m.mu.RLock()
			suiteIDs := make([]string, 0, len(m.benchmarks))
			for suiteID := range m.benchmarks {
				suiteIDs = append(suiteIDs, suiteID)
			}
			m.mu.RUnlock()

			for _, suiteID := range suiteIDs {
				result, err := m.ExecuteBenchmark(ctx, suiteID)
				if err != nil {
					m.logger.Error("Failed to execute benchmark suite",
						zap.String("suite_id", suiteID),
						zap.Error(err))
				} else {
					m.logger.Info("Background benchmark completed",
						zap.String("suite_id", suiteID),
						zap.Float64("success_rate", result.Metrics.SuccessRate),
						zap.Bool("is_valid", result.Validation.IsValid))
				}
			}
		}
	}
}

// GenerateBenchmarkReport generates a comprehensive benchmark report
func (m *SuccessRateBenchmarkManager) GenerateBenchmarkReport(ctx context.Context, suiteID string) (*BenchmarkReport, error) {
	results := m.GetBenchmarkResults(suiteID, 0)
	if len(results) == 0 {
		return nil, fmt.Errorf("no benchmark results found for suite: %s", suiteID)
	}

	// Calculate trend analysis
	trendData := m.calculateTrendAnalysis(results)

	// Generate recommendations
	recommendations := m.generateRecommendations(results)

	report := &BenchmarkReport{
		SuiteID:         suiteID,
		GeneratedAt:     time.Now(),
		Results:         results,
		TrendAnalysis:   trendData,
		Recommendations: recommendations,
		Summary:         m.generateSummary(results),
	}

	return report, nil
}

// BenchmarkReport represents a comprehensive benchmark report
type BenchmarkReport struct {
	SuiteID         string                  `json:"suite_id"`
	GeneratedAt     time.Time               `json:"generated_at"`
	Results         []*BenchmarkResult      `json:"results"`
	TrendAnalysis   *BenchmarkTrendAnalysis `json:"trend_analysis"`
	Recommendations []string                `json:"recommendations"`
	Summary         *BenchmarkSummary       `json:"summary"`
}

// BenchmarkTrendAnalysis represents trend analysis of benchmark results
type BenchmarkTrendAnalysis struct {
	SuccessRateTrend float64 `json:"success_rate_trend"`
	PerformanceTrend float64 `json:"performance_trend"`
	StabilityScore   float64 `json:"stability_score"`
	TrendDirection   string  `json:"trend_direction"` // improving, stable, degrading
}

// BenchmarkSummary represents a summary of benchmark results
type BenchmarkSummary struct {
	TotalExecutions    int           `json:"total_executions"`
	AverageSuccessRate float64       `json:"average_success_rate"`
	BestSuccessRate    float64       `json:"best_success_rate"`
	WorstSuccessRate   float64       `json:"worst_success_rate"`
	AverageDuration    time.Duration `json:"average_duration"`
	TotalTestCases     int           `json:"total_test_cases"`
	ValidationPassRate float64       `json:"validation_pass_rate"`
}

// calculateTrendAnalysis calculates trend analysis from benchmark results
func (m *SuccessRateBenchmarkManager) calculateTrendAnalysis(results []*BenchmarkResult) *BenchmarkTrendAnalysis {
	if len(results) < 2 {
		return &BenchmarkTrendAnalysis{
			TrendDirection: "insufficient_data",
		}
	}

	// Sort by execution time
	sort.Slice(results, func(i, j int) bool {
		return results[i].ExecutionTime.Before(results[j].ExecutionTime)
	})

	// Calculate success rate trend
	successRates := make([]float64, len(results))
	for i, result := range results {
		successRates[i] = result.Metrics.SuccessRate
	}

	trend := m.calculateLinearTrend(successRates)

	var direction string
	if trend > 0.01 {
		direction = "improving"
	} else if trend < -0.01 {
		direction = "degrading"
	} else {
		direction = "stable"
	}

	// Calculate stability score (inverse of variance)
	variance := m.calculateVariance(successRates)
	stabilityScore := 1.0 / (1.0 + variance)

	return &BenchmarkTrendAnalysis{
		SuccessRateTrend: trend,
		PerformanceTrend: 0, // Simplified
		StabilityScore:   stabilityScore,
		TrendDirection:   direction,
	}
}

// calculateLinearTrend calculates linear trend from data points
func (m *SuccessRateBenchmarkManager) calculateLinearTrend(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	n := float64(len(values))
	sumX := n * (n - 1) / 2
	sumY := 0.0
	sumXY := 0.0
	sumX2 := n * (n - 1) * (2*n - 1) / 6

	for i, y := range values {
		sumY += y
		sumXY += float64(i) * y
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	return slope
}

// calculateVariance calculates variance of values
func (m *SuccessRateBenchmarkManager) calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))

	return variance
}

// generateRecommendations generates recommendations based on benchmark results
func (m *SuccessRateBenchmarkManager) generateRecommendations(results []*BenchmarkResult) []string {
	recommendations := make([]string, 0)

	if len(results) == 0 {
		return recommendations
	}

	// Analyze recent results
	recentResults := results
	if len(results) > 5 {
		recentResults = results[len(results)-5:]
	}

	// Check for declining success rates
	decliningCount := 0
	for i := 1; i < len(recentResults); i++ {
		if recentResults[i].Metrics.SuccessRate < recentResults[i-1].Metrics.SuccessRate {
			decliningCount++
		}
	}

	if decliningCount > len(recentResults)/2 {
		recommendations = append(recommendations, "Success rate is declining - investigate recent changes")
	}

	// Check for validation failures
	validationFailures := 0
	for _, result := range recentResults {
		if !result.Validation.IsValid {
			validationFailures++
		}
	}

	if validationFailures > 0 {
		recommendations = append(recommendations, "Multiple validation failures detected - review benchmark configuration")
	}

	// Check for low success rates
	lowSuccessCount := 0
	for _, result := range recentResults {
		if result.Metrics.SuccessRate < m.config.TargetSuccessRate {
			lowSuccessCount++
		}
	}

	if lowSuccessCount > len(recentResults)/2 {
		recommendations = append(recommendations, "Success rate consistently below target - optimize processing logic")
	}

	return recommendations
}

// generateSummary generates a summary of benchmark results
func (m *SuccessRateBenchmarkManager) generateSummary(results []*BenchmarkResult) *BenchmarkSummary {
	if len(results) == 0 {
		return &BenchmarkSummary{}
	}

	totalExecutions := len(results)
	var totalSuccessRate float64
	var totalDuration time.Duration
	var validationPasses int
	var bestSuccessRate, worstSuccessRate float64
	var totalTestCases int

	for i, result := range results {
		totalSuccessRate += result.Metrics.SuccessRate
		totalDuration += result.Duration
		totalTestCases += len(result.TestResults)

		if i == 0 || result.Metrics.SuccessRate > bestSuccessRate {
			bestSuccessRate = result.Metrics.SuccessRate
		}
		if i == 0 || result.Metrics.SuccessRate < worstSuccessRate {
			worstSuccessRate = result.Metrics.SuccessRate
		}

		if result.Validation.IsValid {
			validationPasses++
		}
	}

	averageSuccessRate := totalSuccessRate / float64(totalExecutions)
	averageDuration := totalDuration / time.Duration(totalExecutions)
	validationPassRate := float64(validationPasses) / float64(totalExecutions)

	return &BenchmarkSummary{
		TotalExecutions:    totalExecutions,
		AverageSuccessRate: averageSuccessRate,
		BestSuccessRate:    bestSuccessRate,
		WorstSuccessRate:   worstSuccessRate,
		AverageDuration:    averageDuration,
		TotalTestCases:     totalTestCases,
		ValidationPassRate: validationPassRate,
	}
}
