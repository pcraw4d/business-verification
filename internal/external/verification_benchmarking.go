package external

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// VerificationBenchmarkManager manages verification accuracy benchmarking
type VerificationBenchmarkManager struct {
	config     *BenchmarkConfig
	logger     *zap.Logger
	benchmarks map[string]*BenchmarkSuite
	results    []*BenchmarkResult
	mu         sync.RWMutex
	startTime  time.Time
}

// BenchmarkConfig holds configuration for benchmarking
type BenchmarkConfig struct {
	EnableBenchmarking       bool          `json:"enable_benchmarking"`
	EnableAccuracyTracking   bool          `json:"enable_accuracy_tracking"`
	EnablePerformanceMetrics bool          `json:"enable_performance_metrics"`
	EnableTrendAnalysis      bool          `json:"enable_trend_analysis"`
	BenchmarkInterval        time.Duration `json:"benchmark_interval"`
	MaxBenchmarkHistory      int           `json:"max_benchmark_history"`
	AccuracyThreshold        float64       `json:"accuracy_threshold"`    // Target accuracy threshold
	PerformanceThreshold     time.Duration `json:"performance_threshold"` // Target performance threshold
	MinSampleSize            int           `json:"min_sample_size"`       // Minimum samples for valid benchmark
	ConfidenceLevel          float64       `json:"confidence_level"`      // Statistical confidence level
}

// BenchmarkSuite represents a collection of benchmarks for a specific area
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

// BenchmarkTestCase represents a single test case in a benchmark
type BenchmarkTestCase struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Input           interface{}            `json:"input"`
	ExpectedOutput  interface{}            `json:"expected_output"`
	ActualOutput    interface{}            `json:"actual_output,omitempty"`
	GroundTruth     *VerificationResult    `json:"ground_truth"`
	PredictedResult *VerificationResult    `json:"predicted_result,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	Weight          float64                `json:"weight"`     // Importance weight for this test case
	Difficulty      string                 `json:"difficulty"` // easy, medium, hard
	Tags            []string               `json:"tags"`
	CreatedAt       time.Time              `json:"created_at"`
}

// BenchmarkMetrics holds accuracy and performance metrics
type BenchmarkMetrics struct {
	Accuracy         float64                  `json:"accuracy"`           // Overall accuracy (0-1)
	Precision        float64                  `json:"precision"`          // True positives / (True positives + False positives)
	Recall           float64                  `json:"recall"`             // True positives / (True positives + False negatives)
	F1Score          float64                  `json:"f1_score"`           // Harmonic mean of precision and recall
	Specificity      float64                  `json:"specificity"`        // True negatives / (True negatives + False positives)
	AverageLatency   time.Duration            `json:"average_latency"`    // Average processing time
	ThroughputPerSec float64                  `json:"throughput_per_sec"` // Tests processed per second
	SuccessRate      float64                  `json:"success_rate"`       // Rate of successful test executions
	ErrorRate        float64                  `json:"error_rate"`         // Rate of test execution errors
	ConfusionMatrix  *ConfusionMatrix         `json:"confusion_matrix"`
	PerFieldMetrics  map[string]*FieldMetrics `json:"per_field_metrics"`
	TrendData        []*TrendPoint            `json:"trend_data"`
	LastUpdated      time.Time                `json:"last_updated"`
}

// ConfusionMatrix represents the confusion matrix for classification accuracy
type ConfusionMatrix struct {
	TruePositives  int `json:"true_positives"`
	TrueNegatives  int `json:"true_negatives"`
	FalsePositives int `json:"false_positives"`
	FalseNegatives int `json:"false_negatives"`
}

// FieldMetrics holds accuracy metrics for individual fields
type FieldMetrics struct {
	FieldName   string  `json:"field_name"`
	Accuracy    float64 `json:"accuracy"`
	Precision   float64 `json:"precision"`
	Recall      float64 `json:"recall"`
	F1Score     float64 `json:"f1_score"`
	SampleCount int     `json:"sample_count"`
	ErrorCount  int     `json:"error_count"`
	MeanError   float64 `json:"mean_error"` // For numerical fields
	StdError    float64 `json:"std_error"`  // Standard deviation of errors
}

// TrendPoint represents a point in time for trend analysis
type TrendPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Accuracy   float64   `json:"accuracy"`
	Precision  float64   `json:"precision"`
	Recall     float64   `json:"recall"`
	F1Score    float64   `json:"f1_score"`
	SampleSize int       `json:"sample_size"`
}

// BenchmarkResult represents the result of running a benchmark suite
type BenchmarkResult struct {
	ID           string            `json:"id"`
	SuiteID      string            `json:"suite_id"`
	SuiteName    string            `json:"suite_name"`
	ExecutedAt   time.Time         `json:"executed_at"`
	Duration     time.Duration     `json:"duration"`
	Status       string            `json:"status"` // "completed", "failed", "partial"
	Metrics      *BenchmarkMetrics `json:"metrics"`
	TestResults  []*TestResult     `json:"test_results"`
	Summary      string            `json:"summary"`
	ErrorMessage string            `json:"error_message,omitempty"`
}

// TestResult represents the result of a single test case
type TestResult struct {
	TestCaseID      string        `json:"test_case_id"`
	TestCaseName    string        `json:"test_case_name"`
	Status          string        `json:"status"` // "passed", "failed", "error"
	ExecutionTime   time.Duration `json:"execution_time"`
	IsCorrect       bool          `json:"is_correct"`
	ConfidenceScore float64       `json:"confidence_score"`
	ErrorMessage    string        `json:"error_message,omitempty"`
	ActualOutput    interface{}   `json:"actual_output"`
	ExpectedOutput  interface{}   `json:"expected_output"`
}

// BenchmarkComparison represents a comparison between two benchmarks
type BenchmarkComparison struct {
	BaselineID      string             `json:"baseline_id"`
	ComparisonID    string             `json:"comparison_id"`
	ComparedAt      time.Time          `json:"compared_at"`
	Improvements    map[string]float64 `json:"improvements"` // metric -> improvement percentage
	Regressions     map[string]float64 `json:"regressions"`  // metric -> regression percentage
	Summary         string             `json:"summary"`
	Significance    map[string]bool    `json:"significance"` // metric -> statistically significant
	Recommendations []string           `json:"recommendations"`
}

// NewVerificationBenchmarkManager creates a new benchmark manager
func NewVerificationBenchmarkManager(config *BenchmarkConfig, logger *zap.Logger) *VerificationBenchmarkManager {
	if config == nil {
		config = DefaultBenchmarkConfig()
	}

	return &VerificationBenchmarkManager{
		config:     config,
		logger:     logger,
		benchmarks: make(map[string]*BenchmarkSuite),
		results:    []*BenchmarkResult{},
		startTime:  time.Now(),
	}
}

// DefaultBenchmarkConfig returns default configuration
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		EnableBenchmarking:       true,
		EnableAccuracyTracking:   true,
		EnablePerformanceMetrics: true,
		EnableTrendAnalysis:      true,
		BenchmarkInterval:        24 * time.Hour,
		MaxBenchmarkHistory:      100,
		AccuracyThreshold:        0.90,
		PerformanceThreshold:     5 * time.Second,
		MinSampleSize:            50,
		ConfidenceLevel:          0.95,
	}
}

// CreateBenchmarkSuite creates a new benchmark suite
func (m *VerificationBenchmarkManager) CreateBenchmarkSuite(suite *BenchmarkSuite) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if suite == nil {
		return fmt.Errorf("benchmark suite cannot be nil")
	}

	if suite.ID == "" {
		suite.ID = generateBenchmarkID()
	}

	suite.CreatedAt = time.Now()
	suite.UpdatedAt = time.Now()

	// Initialize metrics
	if suite.Metrics == nil {
		suite.Metrics = &BenchmarkMetrics{
			PerFieldMetrics: make(map[string]*FieldMetrics),
			TrendData:       []*TrendPoint{},
			LastUpdated:     time.Now(),
		}
	}

	m.benchmarks[suite.ID] = suite

	m.logger.Info("Created benchmark suite",
		zap.String("suite_id", suite.ID),
		zap.String("suite_name", suite.Name),
		zap.Int("test_cases", len(suite.TestCases)))

	return nil
}

// RunBenchmark executes a benchmark suite and returns results
func (m *VerificationBenchmarkManager) RunBenchmark(ctx context.Context, suiteID string) (*BenchmarkResult, error) {
	m.mu.RLock()
	suite, exists := m.benchmarks[suiteID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("benchmark suite not found: %s", suiteID)
	}

	startTime := time.Now()
	result := &BenchmarkResult{
		ID:          generateResultID(),
		SuiteID:     suiteID,
		SuiteName:   suite.Name,
		ExecutedAt:  startTime,
		Status:      "running",
		TestResults: []*TestResult{},
	}

	m.logger.Info("Starting benchmark execution",
		zap.String("suite_id", suiteID),
		zap.String("suite_name", suite.Name),
		zap.Int("test_cases", len(suite.TestCases)))

	var testResults []*TestResult
	var totalExecutionTime time.Duration

	// Execute all test cases
	for _, testCase := range suite.TestCases {
		testResult := m.executeTestCase(ctx, testCase)
		testResults = append(testResults, testResult)
		totalExecutionTime += testResult.ExecutionTime
	}

	// Calculate metrics
	metrics := m.calculateBenchmarkMetrics(testResults, suite.TestCases)

	// Update result
	result.Duration = time.Since(startTime)
	result.Status = "completed"
	result.Metrics = metrics
	result.TestResults = testResults
	result.Summary = m.generateBenchmarkSummary(metrics, testResults)

	// Store result
	m.mu.Lock()
	m.results = append(m.results, result)
	if len(m.results) > m.config.MaxBenchmarkHistory {
		m.results = m.results[1:] // Remove oldest result
	}
	m.mu.Unlock()

	// Update suite metrics
	m.updateSuiteMetrics(suite, metrics)

	m.logger.Info("Completed benchmark execution",
		zap.String("suite_id", suiteID),
		zap.Duration("duration", result.Duration),
		zap.Float64("accuracy", metrics.Accuracy),
		zap.Float64("f1_score", metrics.F1Score))

	return result, nil
}

// executeTestCase executes a single test case
func (m *VerificationBenchmarkManager) executeTestCase(ctx context.Context, testCase *BenchmarkTestCase) *TestResult {
	startTime := time.Now()

	result := &TestResult{
		TestCaseID:     testCase.ID,
		TestCaseName:   testCase.Name,
		Status:         "running",
		ExpectedOutput: testCase.ExpectedOutput,
	}

	// Simulate verification execution
	// In a real implementation, this would call the actual verification system
	actualResult, err := m.simulateVerification(ctx, testCase.Input)

	executionTime := time.Since(startTime)
	result.ExecutionTime = executionTime

	if err != nil {
		result.Status = "error"
		result.ErrorMessage = err.Error()
		result.IsCorrect = false
		return result
	}

	result.ActualOutput = actualResult
	result.IsCorrect = m.compareResults(testCase.GroundTruth, actualResult)
	result.ConfidenceScore = m.calculateConfidenceScore(testCase.GroundTruth, actualResult)

	if result.IsCorrect {
		result.Status = "passed"
	} else {
		result.Status = "failed"
	}

	return result
}

// simulateVerification simulates a verification process
func (m *VerificationBenchmarkManager) simulateVerification(ctx context.Context, input interface{}) (*VerificationResult, error) {
	// This is a simulation - in real implementation, this would call the actual verification system
	// For now, we'll return a mock result based on the input

	// Simulate processing time
	time.Sleep(time.Millisecond * 100)

	return &VerificationResult{
		Status:          StatusPassed,
		OverallScore:    0.85,
		ConfidenceLevel: "high",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		FieldResults:    make(map[string]FieldResult),
		Metadata:        make(map[string]string),
	}, nil
}

// compareResults compares ground truth with actual verification result
func (m *VerificationBenchmarkManager) compareResults(groundTruth, actual *VerificationResult) bool {
	if groundTruth == nil || actual == nil {
		return false
	}

	// Compare status
	if groundTruth.Status != actual.Status {
		return false
	}

	// Compare overall score (within tolerance)
	scoreDiff := math.Abs(groundTruth.OverallScore - actual.OverallScore)
	if scoreDiff > 0.1 { // 10% tolerance
		return false
	}

	return true
}

// calculateConfidenceScore calculates confidence score for the test result
func (m *VerificationBenchmarkManager) calculateConfidenceScore(groundTruth, actual *VerificationResult) float64 {
	if groundTruth == nil || actual == nil {
		return 0.0
	}

	var score float64

	// Status match contributes 70%
	if groundTruth.Status == actual.Status {
		score += 0.7
	}

	// Overall score similarity contributes 30%
	scoreDiff := math.Abs(groundTruth.OverallScore - actual.OverallScore)
	scoreScore := math.Max(0, 1.0-scoreDiff)
	score += 0.3 * scoreScore

	return score
}

// calculateBenchmarkMetrics calculates comprehensive metrics for the benchmark
func (m *VerificationBenchmarkManager) calculateBenchmarkMetrics(testResults []*TestResult, testCases []*BenchmarkTestCase) *BenchmarkMetrics {
	if len(testResults) == 0 {
		return &BenchmarkMetrics{}
	}

	var totalExecutionTime time.Duration
	var correctCount int
	var errorCount int

	// Calculate confusion matrix
	confusionMatrix := &ConfusionMatrix{}

	for i, result := range testResults {
		totalExecutionTime += result.ExecutionTime

		if result.Status == "error" {
			errorCount++
			continue
		}

		if result.IsCorrect {
			correctCount++
		}

		// Update confusion matrix based on actual vs expected results
		if i < len(testCases) {
			m.updateConfusionMatrix(confusionMatrix, testCases[i], result)
		}
	}

	// Calculate basic metrics
	totalTests := len(testResults)
	accuracy := float64(correctCount) / float64(totalTests)
	errorRate := float64(errorCount) / float64(totalTests)
	successRate := 1.0 - errorRate
	averageLatency := totalExecutionTime / time.Duration(totalTests)

	// Calculate precision, recall, F1 score
	precision := m.calculatePrecision(confusionMatrix)
	recall := m.calculateRecall(confusionMatrix)
	f1Score := m.calculateF1Score(precision, recall)
	specificity := m.calculateSpecificity(confusionMatrix)

	// Calculate throughput
	throughputPerSec := float64(totalTests) / totalExecutionTime.Seconds()

	return &BenchmarkMetrics{
		Accuracy:         accuracy,
		Precision:        precision,
		Recall:           recall,
		F1Score:          f1Score,
		Specificity:      specificity,
		AverageLatency:   averageLatency,
		ThroughputPerSec: throughputPerSec,
		SuccessRate:      successRate,
		ErrorRate:        errorRate,
		ConfusionMatrix:  confusionMatrix,
		PerFieldMetrics:  make(map[string]*FieldMetrics),
		TrendData:        []*TrendPoint{},
		LastUpdated:      time.Now(),
	}
}

// updateConfusionMatrix updates the confusion matrix based on test results
func (m *VerificationBenchmarkManager) updateConfusionMatrix(matrix *ConfusionMatrix, testCase *BenchmarkTestCase, result *TestResult) {
	groundTruth := testCase.GroundTruth
	if groundTruth == nil {
		return
	}

	actualResult, ok := result.ActualOutput.(*VerificationResult)
	if !ok {
		return
	}

	// Determine if ground truth is positive (PASSED) or negative
	groundTruthPositive := groundTruth.Status == StatusPassed
	actualPositive := actualResult.Status == StatusPassed

	if groundTruthPositive && actualPositive {
		matrix.TruePositives++
	} else if !groundTruthPositive && !actualPositive {
		matrix.TrueNegatives++
	} else if !groundTruthPositive && actualPositive {
		matrix.FalsePositives++
	} else if groundTruthPositive && !actualPositive {
		matrix.FalseNegatives++
	}
}

// calculatePrecision calculates precision from confusion matrix
func (m *VerificationBenchmarkManager) calculatePrecision(matrix *ConfusionMatrix) float64 {
	denominator := matrix.TruePositives + matrix.FalsePositives
	if denominator == 0 {
		return 0.0
	}
	return float64(matrix.TruePositives) / float64(denominator)
}

// calculateRecall calculates recall from confusion matrix
func (m *VerificationBenchmarkManager) calculateRecall(matrix *ConfusionMatrix) float64 {
	denominator := matrix.TruePositives + matrix.FalseNegatives
	if denominator == 0 {
		return 0.0
	}
	return float64(matrix.TruePositives) / float64(denominator)
}

// calculateF1Score calculates F1 score from precision and recall
func (m *VerificationBenchmarkManager) calculateF1Score(precision, recall float64) float64 {
	if precision+recall == 0 {
		return 0.0
	}
	return 2 * (precision * recall) / (precision + recall)
}

// calculateSpecificity calculates specificity from confusion matrix
func (m *VerificationBenchmarkManager) calculateSpecificity(matrix *ConfusionMatrix) float64 {
	denominator := matrix.TrueNegatives + matrix.FalsePositives
	if denominator == 0 {
		return 0.0
	}
	return float64(matrix.TrueNegatives) / float64(denominator)
}

// generateBenchmarkSummary generates a human-readable summary
func (m *VerificationBenchmarkManager) generateBenchmarkSummary(metrics *BenchmarkMetrics, testResults []*TestResult) string {
	passed := 0
	failed := 0
	errors := 0

	for _, result := range testResults {
		switch result.Status {
		case "passed":
			passed++
		case "failed":
			failed++
		case "error":
			errors++
		}
	}

	return fmt.Sprintf("Benchmark completed: %d passed, %d failed, %d errors. Accuracy: %.2f%%, F1-Score: %.3f, Average latency: %v",
		passed, failed, errors, metrics.Accuracy*100, metrics.F1Score, metrics.AverageLatency)
}

// updateSuiteMetrics updates the metrics for a benchmark suite
func (m *VerificationBenchmarkManager) updateSuiteMetrics(suite *BenchmarkSuite, metrics *BenchmarkMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add trend point
	trendPoint := &TrendPoint{
		Timestamp:  time.Now(),
		Accuracy:   metrics.Accuracy,
		Precision:  metrics.Precision,
		Recall:     metrics.Recall,
		F1Score:    metrics.F1Score,
		SampleSize: len(suite.TestCases),
	}

	suite.Metrics.TrendData = append(suite.Metrics.TrendData, trendPoint)

	// Keep only recent trend data
	maxTrendPoints := 100
	if len(suite.Metrics.TrendData) > maxTrendPoints {
		suite.Metrics.TrendData = suite.Metrics.TrendData[len(suite.Metrics.TrendData)-maxTrendPoints:]
	}

	// Update current metrics
	suite.Metrics.Accuracy = metrics.Accuracy
	suite.Metrics.Precision = metrics.Precision
	suite.Metrics.Recall = metrics.Recall
	suite.Metrics.F1Score = metrics.F1Score
	suite.Metrics.Specificity = metrics.Specificity
	suite.Metrics.AverageLatency = metrics.AverageLatency
	suite.Metrics.ThroughputPerSec = metrics.ThroughputPerSec
	suite.Metrics.SuccessRate = metrics.SuccessRate
	suite.Metrics.ErrorRate = metrics.ErrorRate
	suite.Metrics.ConfusionMatrix = metrics.ConfusionMatrix
	suite.Metrics.LastUpdated = time.Now()

	suite.UpdatedAt = time.Now()
}

// GetBenchmarkSuite retrieves a benchmark suite by ID
func (m *VerificationBenchmarkManager) GetBenchmarkSuite(suiteID string) (*BenchmarkSuite, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	suite, exists := m.benchmarks[suiteID]
	if !exists {
		return nil, fmt.Errorf("benchmark suite not found: %s", suiteID)
	}

	return suite, nil
}

// ListBenchmarkSuites returns all benchmark suites
func (m *VerificationBenchmarkManager) ListBenchmarkSuites() []*BenchmarkSuite {
	m.mu.RLock()
	defer m.mu.RUnlock()

	suites := make([]*BenchmarkSuite, 0, len(m.benchmarks))
	for _, suite := range m.benchmarks {
		suites = append(suites, suite)
	}

	// Sort by name
	sort.Slice(suites, func(i, j int) bool {
		return suites[i].Name < suites[j].Name
	})

	return suites
}

// GetBenchmarkResults returns recent benchmark results
func (m *VerificationBenchmarkManager) GetBenchmarkResults(limit int) []*BenchmarkResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.results) {
		limit = len(m.results)
	}

	// Return most recent results
	start := len(m.results) - limit
	if start < 0 {
		start = 0
	}

	results := make([]*BenchmarkResult, limit)
	copy(results, m.results[start:])

	// Sort by execution time (most recent first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].ExecutedAt.After(results[j].ExecutedAt)
	})

	return results
}

// CompareBenchmarks compares two benchmark results
func (m *VerificationBenchmarkManager) CompareBenchmarks(baselineID, comparisonID string) (*BenchmarkComparison, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var baseline, comparison *BenchmarkResult

	// Find benchmark results
	for _, result := range m.results {
		if result.ID == baselineID {
			baseline = result
		}
		if result.ID == comparisonID {
			comparison = result
		}
	}

	if baseline == nil {
		return nil, fmt.Errorf("baseline benchmark not found: %s", baselineID)
	}
	if comparison == nil {
		return nil, fmt.Errorf("comparison benchmark not found: %s", comparisonID)
	}

	// Calculate improvements and regressions
	improvements := make(map[string]float64)
	regressions := make(map[string]float64)
	significance := make(map[string]bool)

	// Compare key metrics
	metrics := map[string][2]float64{
		"accuracy":   {baseline.Metrics.Accuracy, comparison.Metrics.Accuracy},
		"precision":  {baseline.Metrics.Precision, comparison.Metrics.Precision},
		"recall":     {baseline.Metrics.Recall, comparison.Metrics.Recall},
		"f1_score":   {baseline.Metrics.F1Score, comparison.Metrics.F1Score},
		"error_rate": {baseline.Metrics.ErrorRate, comparison.Metrics.ErrorRate},
	}

	for metric, values := range metrics {
		baselineValue := values[0]
		comparisonValue := values[1]

		if baselineValue > 0 {
			change := (comparisonValue - baselineValue) / baselineValue * 100

			if change > 0 {
				if metric == "error_rate" {
					regressions[metric] = change
				} else {
					improvements[metric] = change
				}
			} else if change < 0 {
				if metric == "error_rate" {
					improvements[metric] = -change
				} else {
					regressions[metric] = -change
				}
			}

			// Simple significance test (>5% change)
			significance[metric] = math.Abs(change) > 5.0
		}
	}

	// Generate summary
	summary := m.generateComparisonSummary(improvements, regressions)

	// Generate recommendations
	recommendations := m.generateComparisonRecommendations(improvements, regressions, significance)

	return &BenchmarkComparison{
		BaselineID:      baselineID,
		ComparisonID:    comparisonID,
		ComparedAt:      time.Now(),
		Improvements:    improvements,
		Regressions:     regressions,
		Summary:         summary,
		Significance:    significance,
		Recommendations: recommendations,
	}, nil
}

// generateComparisonSummary generates a summary of the comparison
func (m *VerificationBenchmarkManager) generateComparisonSummary(improvements, regressions map[string]float64) string {
	improvementCount := len(improvements)
	regressionCount := len(regressions)

	if improvementCount > regressionCount {
		return fmt.Sprintf("Overall improvement: %d metrics improved, %d regressed", improvementCount, regressionCount)
	} else if regressionCount > improvementCount {
		return fmt.Sprintf("Overall regression: %d metrics regressed, %d improved", regressionCount, improvementCount)
	} else {
		return fmt.Sprintf("Mixed results: %d metrics improved, %d regressed", improvementCount, regressionCount)
	}
}

// generateComparisonRecommendations generates recommendations based on comparison
func (m *VerificationBenchmarkManager) generateComparisonRecommendations(improvements, regressions map[string]float64, significance map[string]bool) []string {
	var recommendations []string

	// Check for significant regressions
	for metric, regression := range regressions {
		if significance[metric] && regression > 10 {
			recommendations = append(recommendations, fmt.Sprintf("Investigate significant %s regression (%.1f%%)", metric, regression))
		}
	}

	// Check for significant improvements
	for metric, improvement := range improvements {
		if significance[metric] && improvement > 15 {
			recommendations = append(recommendations, fmt.Sprintf("Analyze %s improvement (%.1f%%) for reproducibility", metric, improvement))
		}
	}

	// General recommendations
	if len(regressions) > len(improvements) {
		recommendations = append(recommendations, "Review recent changes that may have impacted accuracy")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Performance is stable - continue monitoring")
	}

	return recommendations
}

// GetConfig returns the current configuration
func (m *VerificationBenchmarkManager) GetConfig() *BenchmarkConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// UpdateConfig updates the configuration
func (m *VerificationBenchmarkManager) UpdateConfig(config *BenchmarkConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate configuration
	if config.AccuracyThreshold < 0 || config.AccuracyThreshold > 1 {
		return fmt.Errorf("accuracy threshold must be between 0 and 1")
	}

	if config.ConfidenceLevel < 0 || config.ConfidenceLevel > 1 {
		return fmt.Errorf("confidence level must be between 0 and 1")
	}

	if config.BenchmarkInterval < time.Minute {
		return fmt.Errorf("benchmark interval must be at least 1 minute")
	}

	m.config = config
	return nil
}

// Helper functions
func generateBenchmarkID() string {
	return fmt.Sprintf("benchmark_%d", time.Now().UnixNano())
}

func generateResultID() string {
	return fmt.Sprintf("result_%d", time.Now().UnixNano())
}
