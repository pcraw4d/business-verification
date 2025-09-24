package success_monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewSuccessRateBenchmarkManager(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.config)
	assert.Equal(t, 0.95, manager.config.TargetSuccessRate)
	assert.Equal(t, 0.95, manager.config.ConfidenceLevel)
	assert.Equal(t, 100, manager.config.MinSampleSize)
	assert.Equal(t, 10000, manager.config.MaxSampleSize)
	assert.Equal(t, 0.02, manager.config.ValidationThreshold)
}

func TestCreateBenchmarkSuite(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	suite := &BenchmarkSuite{
		Name:        "Test Suite",
		Description: "Test benchmark suite",
		Category:    "classification",
		TestCases:   []*BenchmarkTestCase{},
	}

	err := manager.CreateBenchmarkSuite(context.Background(), suite)
	require.NoError(t, err)

	assert.NotEmpty(t, suite.ID)
	assert.NotZero(t, suite.CreatedAt)
	assert.NotZero(t, suite.UpdatedAt)

	// Verify suite was stored
	manager.mu.RLock()
	_, exists := manager.benchmarks[suite.ID]
	manager.mu.RUnlock()
	assert.True(t, exists)
}

func TestExecuteBenchmark(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	// Create test cases
	testCases := []*BenchmarkTestCase{
		{
			ID:              "test1",
			Name:            "Easy Test",
			Description:     "Easy test case",
			Input:           "test input",
			ExpectedSuccess: true,
			Difficulty:      "easy",
			Weight:          1.0,
			CreatedAt:       time.Now(),
		},
		{
			ID:              "test2",
			Name:            "Medium Test",
			Description:     "Medium test case",
			Input:           "test input 2",
			ExpectedSuccess: true,
			Difficulty:      "medium",
			Weight:          1.0,
			CreatedAt:       time.Now(),
		},
		{
			ID:              "test3",
			Name:            "Hard Test",
			Description:     "Hard test case",
			Input:           "test input 3",
			ExpectedSuccess: true,
			Difficulty:      "hard",
			Weight:          1.0,
			CreatedAt:       time.Now(),
		},
	}

	suite := &BenchmarkSuite{
		Name:        "Test Suite",
		Description: "Test benchmark suite",
		Category:    "classification",
		TestCases:   testCases,
	}

	err := manager.CreateBenchmarkSuite(context.Background(), suite)
	require.NoError(t, err)

	result, err := manager.ExecuteBenchmark(context.Background(), suite.ID)
	require.NoError(t, err)

	assert.NotNil(t, result)
	assert.Equal(t, suite.ID, result.SuiteID)
	assert.NotZero(t, result.ExecutionTime)
	assert.NotZero(t, result.Duration)
	assert.NotNil(t, result.Metrics)
	assert.NotNil(t, result.Validation)
	assert.Len(t, result.TestResults, 3)

	// Verify metrics
	assert.GreaterOrEqual(t, result.Metrics.SuccessRate, 0.0)
	assert.LessOrEqual(t, result.Metrics.SuccessRate, 1.0)
	assert.NotNil(t, result.Metrics.ConfidenceInterval)
	assert.NotZero(t, result.Metrics.AverageResponseTime)
	assert.Greater(t, result.Metrics.ThroughputPerSec, 0.0)

	// Verify validation
	assert.NotNil(t, result.Validation)
	assert.GreaterOrEqual(t, result.Validation.ConfidenceScore, 0.0)
	assert.LessOrEqual(t, result.Validation.ConfidenceScore, 1.0)
}

func TestCalculateConfidenceInterval(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	tests := []struct {
		name            string
		successes       int
		total           int
		confidenceLevel float64
		expectedLower   float64
		expectedUpper   float64
	}{
		{
			name:            "perfect success rate",
			successes:       100,
			total:           100,
			confidenceLevel: 0.95,
			expectedLower:   0.95,
			expectedUpper:   1.0,
		},
		{
			name:            "high success rate",
			successes:       95,
			total:           100,
			confidenceLevel: 0.95,
			expectedLower:   0.88,
			expectedUpper:   0.98,
		},
		{
			name:            "medium success rate",
			successes:       80,
			total:           100,
			confidenceLevel: 0.95,
			expectedLower:   0.71,
			expectedUpper:   0.87,
		},
		{
			name:            "low success rate",
			successes:       50,
			total:           100,
			confidenceLevel: 0.95,
			expectedLower:   0.40,
			expectedUpper:   0.60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := manager.calculateConfidenceInterval(tt.successes, tt.total, tt.confidenceLevel)

			assert.NotNil(t, interval)
			assert.Equal(t, tt.confidenceLevel, interval.Level)
			assert.GreaterOrEqual(t, interval.LowerBound, 0.0)
			assert.LessOrEqual(t, interval.UpperBound, 1.0)
			assert.LessOrEqual(t, interval.LowerBound, interval.UpperBound)

			// Verify the interval contains the observed proportion
			observed := float64(tt.successes) / float64(tt.total)
			assert.GreaterOrEqual(t, observed, interval.LowerBound)
			assert.LessOrEqual(t, observed, interval.UpperBound)
		})
	}
}

func TestCalculateStatisticalSignificance(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	// Create test results with high success rate
	testResults := make([]*TestResult, 100)
	for i := 0; i < 100; i++ {
		testResults[i] = &TestResult{
			TestCaseID: fmt.Sprintf("test_%d", i),
			Success:    i < 95, // 95% success rate
		}
	}

	testCases := make([]*BenchmarkTestCase, 100)
	for i := 0; i < 100; i++ {
		testCases[i] = &BenchmarkTestCase{
			ID:         fmt.Sprintf("test_%d", i),
			Difficulty: "medium",
			Weight:     1.0,
		}
	}

	significance := manager.calculateStatisticalSignificance(testResults, testCases)

	assert.NotNil(t, significance)
	assert.Equal(t, "z_test", significance.TestType)
	assert.GreaterOrEqual(t, significance.PValue, 0.0)
	assert.LessOrEqual(t, significance.PValue, 1.0)
	assert.GreaterOrEqual(t, significance.Power, 0.0)
	assert.LessOrEqual(t, significance.Power, 1.0)

	// With 95% success rate vs 95% target, should not be significantly different
	assert.False(t, significance.IsSignificant)
}

func TestValidateBenchmarkResults(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	tests := []struct {
		name           string
		successRate    float64
		sampleSize     int
		expectedValid  bool
		expectedErrors int
	}{
		{
			name:           "valid results",
			successRate:    0.96,
			sampleSize:     200,
			expectedValid:  false, // Will fail due to confidence interval width
			expectedErrors: 1,
		},
		{
			name:           "low success rate",
			successRate:    0.85,
			sampleSize:     200,
			expectedValid:  false,
			expectedErrors: 2, // Success rate + confidence interval
		},
		{
			name:           "small sample size",
			successRate:    0.96,
			sampleSize:     50,
			expectedValid:  false,
			expectedErrors: 2, // Sample size + confidence interval
		},
		{
			name:           "low success rate and small sample",
			successRate:    0.85,
			sampleSize:     50,
			expectedValid:  false,
			expectedErrors: 3, // Sample size + success rate + confidence interval
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &BenchmarkMetrics{
				SuccessRate: tt.successRate,
				ConfidenceInterval: &ConfidenceInterval{
					LowerBound: tt.successRate - 0.05,
					UpperBound: tt.successRate + 0.05,
					Level:      0.95,
				},
				StatisticalSignificance: &StatisticalSignificance{
					IsSignificant: true,
					PValue:        0.01,
					TestType:      "z_test",
				},
			}

			testResults := make([]*TestResult, tt.sampleSize)
			for i := 0; i < tt.sampleSize; i++ {
				testResults[i] = &TestResult{
					TestCaseID: fmt.Sprintf("test_%d", i),
					Success:    i < int(float64(tt.sampleSize)*tt.successRate),
				}
			}

			validation := manager.validateBenchmarkResults(metrics, testResults)

			assert.NotNil(t, validation)
			assert.Equal(t, tt.expectedValid, validation.IsValid)
			assert.Len(t, validation.ValidationErrors, tt.expectedErrors)
			assert.GreaterOrEqual(t, validation.ConfidenceScore, 0.0)
			assert.LessOrEqual(t, validation.ConfidenceScore, 1.0)
		})
	}
}

func TestUpdateBaseline(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	category := "classification"
	successRate := 0.95
	sampleCount := 1000

	err := manager.UpdateBaseline(context.Background(), category, successRate, sampleCount)
	require.NoError(t, err)

	baseline := manager.GetBaselineMetrics(category)
	assert.NotNil(t, baseline)
	assert.Equal(t, category, baseline.ProcessName)
	assert.Equal(t, successRate, baseline.SuccessRate)
	assert.Equal(t, sampleCount, baseline.SampleCount)
	assert.NotNil(t, baseline.ConfidenceInterval)
	assert.Len(t, baseline.HistoricalData, 1)

	// Update again
	err = manager.UpdateBaseline(context.Background(), category, 0.97, 1200)
	require.NoError(t, err)

	baseline = manager.GetBaselineMetrics(category)
	assert.Equal(t, 0.97, baseline.SuccessRate)
	assert.Equal(t, 1200, baseline.SampleCount)
	assert.Len(t, baseline.HistoricalData, 1) // Only one historical point after update
}

func TestCompareWithBaseline(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	// Set up baseline
	category := "classification"
	err := manager.UpdateBaseline(context.Background(), category, 0.90, 1000)
	require.NoError(t, err)

	// Test improvement
	metrics := &BenchmarkMetrics{
		SuccessRate: 0.95,
	}

	comparison := manager.compareWithBaseline(category, metrics)
	assert.NotNil(t, comparison)
	assert.Equal(t, 0.90, comparison.BaselineSuccessRate)
	assert.Equal(t, 0.95, comparison.CurrentSuccessRate)
	assert.InDelta(t, 0.05, comparison.Improvement, 0.001)
	assert.True(t, comparison.IsSignificant)

	// Test regression
	metrics.SuccessRate = 0.85
	comparison = manager.compareWithBaseline(category, metrics)
	assert.InDelta(t, -0.05, comparison.Improvement, 0.001)
	assert.False(t, comparison.IsSignificant)

	// Test no baseline
	comparison = manager.compareWithBaseline("nonexistent", metrics)
	assert.Equal(t, 0.0, comparison.BaselineSuccessRate)
	assert.Equal(t, 0.85, comparison.CurrentSuccessRate)
	assert.Equal(t, 0.85, comparison.Improvement)
	assert.False(t, comparison.IsSignificant)
}

func TestGenerateBenchmarkReport(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	// Create and execute a benchmark suite
	testCases := []*BenchmarkTestCase{
		{
			ID:         "test1",
			Name:       "Test 1",
			Difficulty: "easy",
			Weight:     1.0,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "test2",
			Name:       "Test 2",
			Difficulty: "medium",
			Weight:     1.0,
			CreatedAt:  time.Now(),
		},
	}

	suite := &BenchmarkSuite{
		Name:        "Test Suite",
		Description: "Test benchmark suite",
		Category:    "classification",
		TestCases:   testCases,
	}

	err := manager.CreateBenchmarkSuite(context.Background(), suite)
	require.NoError(t, err)

	// Execute benchmark multiple times
	for i := 0; i < 3; i++ {
		_, err = manager.ExecuteBenchmark(context.Background(), suite.ID)
		require.NoError(t, err)
	}

	// Generate report
	report, err := manager.GenerateBenchmarkReport(context.Background(), suite.ID)
	require.NoError(t, err)

	assert.NotNil(t, report)
	assert.Equal(t, suite.ID, report.SuiteID)
	assert.NotZero(t, report.GeneratedAt)
	assert.Len(t, report.Results, 3)
	assert.NotNil(t, report.TrendAnalysis)
	assert.NotNil(t, report.Summary)
	assert.NotEmpty(t, report.Recommendations)

	// Verify summary
	assert.Equal(t, 3, report.Summary.TotalExecutions)
	assert.GreaterOrEqual(t, report.Summary.AverageSuccessRate, 0.0)
	assert.LessOrEqual(t, report.Summary.AverageSuccessRate, 1.0)
	assert.GreaterOrEqual(t, report.Summary.BestSuccessRate, report.Summary.WorstSuccessRate)
	assert.NotZero(t, report.Summary.AverageDuration)
	assert.Equal(t, 6, report.Summary.TotalTestCases) // 2 test cases * 3 executions
}

func TestCalculateTrendAnalysis(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	// Create benchmark results with improving trend
	results := []*BenchmarkResult{
		{
			ExecutionTime: time.Now().Add(-2 * time.Hour),
			Metrics: &BenchmarkMetrics{
				SuccessRate: 0.90,
			},
		},
		{
			ExecutionTime: time.Now().Add(-1 * time.Hour),
			Metrics: &BenchmarkMetrics{
				SuccessRate: 0.92,
			},
		},
		{
			ExecutionTime: time.Now(),
			Metrics: &BenchmarkMetrics{
				SuccessRate: 0.95,
			},
		},
	}

	trendAnalysis := manager.calculateTrendAnalysis(results)
	assert.NotNil(t, trendAnalysis)
	assert.Greater(t, trendAnalysis.SuccessRateTrend, 0.0) // Should be improving
	assert.Equal(t, "improving", trendAnalysis.TrendDirection)
	assert.Greater(t, trendAnalysis.StabilityScore, 0.0)

	// Test with insufficient data
	trendAnalysis = manager.calculateTrendAnalysis([]*BenchmarkResult{})
	assert.Equal(t, "insufficient_data", trendAnalysis.TrendDirection)

	// Test with single result
	trendAnalysis = manager.calculateTrendAnalysis(results[:1])
	assert.Equal(t, "insufficient_data", trendAnalysis.TrendDirection)
}

func TestCalculateLinearTrend(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	// Test improving trend
	improving := []float64{0.90, 0.92, 0.95}
	trend := manager.calculateLinearTrend(improving)
	assert.Greater(t, trend, 0.0)

	// Test declining trend
	declining := []float64{0.95, 0.92, 0.90}
	trend = manager.calculateLinearTrend(declining)
	assert.Less(t, trend, 0.0)

	// Test stable trend
	stable := []float64{0.92, 0.92, 0.92}
	trend = manager.calculateLinearTrend(stable)
	assert.InDelta(t, 0.0, trend, 0.001)

	// Test insufficient data
	trend = manager.calculateLinearTrend([]float64{0.92})
	assert.Equal(t, 0.0, trend)
}

func TestCalculateVariance(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	// Test with constant values (zero variance)
	constant := []float64{0.92, 0.92, 0.92}
	variance := manager.calculateVariance(constant)
	assert.InDelta(t, 0.0, variance, 0.001)

	// Test with varying values
	varying := []float64{0.90, 0.92, 0.95}
	variance = manager.calculateVariance(varying)
	assert.Greater(t, variance, 0.0)

	// Test with empty slice
	variance = manager.calculateVariance([]float64{})
	assert.Equal(t, 0.0, variance)
}

func TestGenerateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	// Test with declining success rates
	decliningResults := []*BenchmarkResult{
		{
			ExecutionTime: time.Now().Add(-2 * time.Hour),
			Metrics: &BenchmarkMetrics{
				SuccessRate: 0.95,
			},
			Validation: &ValidationResult{
				IsValid: true,
			},
		},
		{
			ExecutionTime: time.Now().Add(-1 * time.Hour),
			Metrics: &BenchmarkMetrics{
				SuccessRate: 0.92,
			},
			Validation: &ValidationResult{
				IsValid: true,
			},
		},
		{
			ExecutionTime: time.Now(),
			Metrics: &BenchmarkMetrics{
				SuccessRate: 0.88,
			},
			Validation: &ValidationResult{
				IsValid: true,
			},
		},
	}

	recommendations := manager.generateRecommendations(decliningResults)
	assert.Contains(t, recommendations, "Success rate is declining - investigate recent changes")

	// Test with validation failures
	validationFailures := []*BenchmarkResult{
		{
			Metrics: &BenchmarkMetrics{
				SuccessRate: 0.95,
			},
			Validation: &ValidationResult{
				IsValid: false,
			},
		},
	}

	recommendations = manager.generateRecommendations(validationFailures)
	assert.Contains(t, recommendations, "Multiple validation failures detected - review benchmark configuration")

	// Test with low success rates
	lowSuccessResults := []*BenchmarkResult{
		{
			Metrics: &BenchmarkMetrics{
				SuccessRate: 0.85, // Below 95% target
			},
			Validation: &ValidationResult{
				IsValid: true,
			},
		},
	}

	recommendations = manager.generateRecommendations(lowSuccessResults)
	assert.Contains(t, recommendations, "Success rate consistently below target - optimize processing logic")
}

func TestGenerateSummary(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	results := []*BenchmarkResult{
		{
			Metrics: &BenchmarkMetrics{
				SuccessRate: 0.90,
			},
			Duration: 1 * time.Second,
			TestResults: []*TestResult{
				{TestCaseID: "test1"},
				{TestCaseID: "test2"},
			},
			Validation: &ValidationResult{
				IsValid: true,
			},
		},
		{
			Metrics: &BenchmarkMetrics{
				SuccessRate: 0.95,
			},
			Duration: 2 * time.Second,
			TestResults: []*TestResult{
				{TestCaseID: "test3"},
				{TestCaseID: "test4"},
			},
			Validation: &ValidationResult{
				IsValid: true,
			},
		},
		{
			Metrics: &BenchmarkMetrics{
				SuccessRate: 0.88,
			},
			Duration: 1500 * time.Millisecond,
			TestResults: []*TestResult{
				{TestCaseID: "test5"},
			},
			Validation: &ValidationResult{
				IsValid: false,
			},
		},
	}

	summary := manager.generateSummary(results)
	assert.NotNil(t, summary)
	assert.Equal(t, 3, summary.TotalExecutions)
	assert.InDelta(t, 0.91, summary.AverageSuccessRate, 0.01) // (0.90 + 0.95 + 0.88) / 3
	assert.Equal(t, 0.95, summary.BestSuccessRate)
	assert.Equal(t, 0.88, summary.WorstSuccessRate)
	assert.Equal(t, time.Duration(1500*time.Millisecond), summary.AverageDuration)
	assert.Equal(t, 5, summary.TotalTestCases)                // 2 + 2 + 1
	assert.InDelta(t, 0.67, summary.ValidationPassRate, 0.01) // 2 valid out of 3
}

func TestBenchmarkManagerIntegration(t *testing.T) {
	logger := zap.NewNop()
	manager := NewSuccessRateBenchmarkManager(nil, logger)

	// Create comprehensive test suite
	testCases := []*BenchmarkTestCase{
		{
			ID:         "easy_1",
			Name:       "Easy Test 1",
			Difficulty: "easy",
			Weight:     1.0,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "easy_2",
			Name:       "Easy Test 2",
			Difficulty: "easy",
			Weight:     1.0,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "medium_1",
			Name:       "Medium Test 1",
			Difficulty: "medium",
			Weight:     1.5,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "hard_1",
			Name:       "Hard Test 1",
			Difficulty: "hard",
			Weight:     2.0,
			CreatedAt:  time.Now(),
		},
	}

	suite := &BenchmarkSuite{
		Name:        "Integration Test Suite",
		Description: "Comprehensive test suite for integration testing",
		Category:    "classification",
		TestCases:   testCases,
	}

	// Create suite
	err := manager.CreateBenchmarkSuite(context.Background(), suite)
	require.NoError(t, err)

	// Set baseline
	err = manager.UpdateBaseline(context.Background(), "classification", 0.92, 1000)
	require.NoError(t, err)

	// Execute benchmark
	result, err := manager.ExecuteBenchmark(context.Background(), suite.ID)
	require.NoError(t, err)

	// Verify results
	assert.NotNil(t, result)
	assert.Equal(t, suite.ID, result.SuiteID)
	assert.Len(t, result.TestResults, 4)
	assert.NotNil(t, result.BaselineComparison)
	assert.NotNil(t, result.Validation)

	// Verify metrics
	assert.GreaterOrEqual(t, result.Metrics.SuccessRate, 0.0)
	assert.LessOrEqual(t, result.Metrics.SuccessRate, 1.0)
	assert.NotNil(t, result.Metrics.ConfidenceInterval)
	assert.Len(t, result.Metrics.PerCategoryMetrics, 3) // easy, medium, hard

	// Verify baseline comparison
	assert.Equal(t, 0.92, result.BaselineComparison.BaselineSuccessRate)
	assert.Equal(t, result.Metrics.SuccessRate, result.BaselineComparison.CurrentSuccessRate)

	// Generate report
	report, err := manager.GenerateBenchmarkReport(context.Background(), suite.ID)
	require.NoError(t, err)

	assert.NotNil(t, report)
	assert.Equal(t, suite.ID, report.SuiteID)
	assert.Len(t, report.Results, 1)
	assert.NotNil(t, report.TrendAnalysis)
	assert.NotNil(t, report.Summary)
	assert.Equal(t, 1, report.Summary.TotalExecutions)
	assert.Equal(t, 4, report.Summary.TotalTestCases)
}
