package external

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewVerificationBenchmarkManager(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	manager := NewVerificationBenchmarkManager(nil, logger)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.config)
	assert.True(t, manager.config.EnableBenchmarking)
	assert.Equal(t, 24*time.Hour, manager.config.BenchmarkInterval)
	assert.Equal(t, 0.90, manager.config.AccuracyThreshold)

	// Test with custom config
	customConfig := &BenchmarkConfig{
		EnableBenchmarking: false,
		BenchmarkInterval:  12 * time.Hour,
		AccuracyThreshold:  0.95,
		MinSampleSize:      100,
	}

	manager2 := NewVerificationBenchmarkManager(customConfig, logger)
	assert.NotNil(t, manager2)
	assert.False(t, manager2.config.EnableBenchmarking)
	assert.Equal(t, 12*time.Hour, manager2.config.BenchmarkInterval)
	assert.Equal(t, 0.95, manager2.config.AccuracyThreshold)
	assert.Equal(t, 100, manager2.config.MinSampleSize)
}

func TestDefaultBenchmarkConfig(t *testing.T) {
	config := DefaultBenchmarkConfig()

	assert.True(t, config.EnableBenchmarking)
	assert.True(t, config.EnableAccuracyTracking)
	assert.True(t, config.EnablePerformanceMetrics)
	assert.True(t, config.EnableTrendAnalysis)
	assert.Equal(t, 24*time.Hour, config.BenchmarkInterval)
	assert.Equal(t, 100, config.MaxBenchmarkHistory)
	assert.Equal(t, 0.90, config.AccuracyThreshold)
	assert.Equal(t, 5*time.Second, config.PerformanceThreshold)
	assert.Equal(t, 50, config.MinSampleSize)
	assert.Equal(t, 0.95, config.ConfidenceLevel)
}

func TestCreateBenchmarkSuite(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	// Test with valid suite
	suite := &BenchmarkSuite{
		Name:        "Test Suite",
		Description: "A test benchmark suite",
		Category:    "verification",
		TestCases: []*BenchmarkTestCase{
			{
				Name:           "Test Case 1",
				Description:    "First test case",
				Input:          "test input",
				ExpectedOutput: "test output",
				GroundTruth: &VerificationResult{
					Status:       StatusPassed,
					OverallScore: 0.9,
				},
				Weight:     1.0,
				Difficulty: "easy",
			},
		},
	}

	err := manager.CreateBenchmarkSuite(suite)
	assert.NoError(t, err)
	assert.NotEmpty(t, suite.ID)
	assert.False(t, suite.CreatedAt.IsZero())
	assert.False(t, suite.UpdatedAt.IsZero())
	assert.NotNil(t, suite.Metrics)

	// Test with nil suite
	err = manager.CreateBenchmarkSuite(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "benchmark suite cannot be nil")
}

func TestRunBenchmark(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	// Create test suite
	suite := &BenchmarkSuite{
		ID:          "test-suite-1",
		Name:        "Test Suite",
		Description: "A test benchmark suite",
		Category:    "verification",
		TestCases: []*BenchmarkTestCase{
			{
				ID:             "test-case-1",
				Name:           "Test Case 1",
				Description:    "First test case",
				Input:          "test input 1",
				ExpectedOutput: "test output 1",
				GroundTruth: &VerificationResult{
					Status:       StatusPassed,
					OverallScore: 0.9,
				},
				Weight:     1.0,
				Difficulty: "easy",
			},
			{
				ID:             "test-case-2",
				Name:           "Test Case 2",
				Description:    "Second test case",
				Input:          "test input 2",
				ExpectedOutput: "test output 2",
				GroundTruth: &VerificationResult{
					Status:       StatusFailed,
					OverallScore: 0.3,
				},
				Weight:     1.0,
				Difficulty: "medium",
			},
		},
	}

	err := manager.CreateBenchmarkSuite(suite)
	assert.NoError(t, err)

	// Run benchmark
	result, err := manager.RunBenchmark(context.Background(), suite.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, suite.ID, result.SuiteID)
	assert.Equal(t, suite.Name, result.SuiteName)
	assert.Equal(t, "completed", result.Status)
	assert.NotNil(t, result.Metrics)
	assert.Len(t, result.TestResults, 2)
	assert.Greater(t, result.Duration, time.Duration(0))

	// Check metrics
	assert.GreaterOrEqual(t, result.Metrics.Accuracy, 0.0)
	assert.LessOrEqual(t, result.Metrics.Accuracy, 1.0)
	assert.NotNil(t, result.Metrics.ConfusionMatrix)

	// Test with non-existent suite
	_, err = manager.RunBenchmark(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "benchmark suite not found")
}

func TestExecuteTestCase(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	testCase := &BenchmarkTestCase{
		ID:             "test-case-1",
		Name:           "Test Case 1",
		Description:    "First test case",
		Input:          "test input",
		ExpectedOutput: "test output",
		GroundTruth: &VerificationResult{
			Status:       StatusPassed,
			OverallScore: 0.9,
		},
		Weight:     1.0,
		Difficulty: "easy",
	}

	result := manager.executeTestCase(context.Background(), testCase)

	assert.NotNil(t, result)
	assert.Equal(t, testCase.ID, result.TestCaseID)
	assert.Equal(t, testCase.Name, result.TestCaseName)
	assert.Contains(t, []string{"passed", "failed", "error"}, result.Status)
	assert.Greater(t, result.ExecutionTime, time.Duration(0))
	assert.NotNil(t, result.ActualOutput)
	assert.GreaterOrEqual(t, result.ConfidenceScore, 0.0)
	assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
}

func TestCompareResults(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	// Test with matching results
	groundTruth := &VerificationResult{
		Status:       StatusPassed,
		OverallScore: 0.9,
	}
	actual := &VerificationResult{
		Status:       StatusPassed,
		OverallScore: 0.85, // Within tolerance
	}

	isCorrect := manager.compareResults(groundTruth, actual)
	assert.True(t, isCorrect)

	// Test with different status
	actual.Status = StatusFailed
	isCorrect = manager.compareResults(groundTruth, actual)
	assert.False(t, isCorrect)

	// Test with score outside tolerance
	actual.Status = StatusPassed
	actual.OverallScore = 0.5 // Outside 10% tolerance
	isCorrect = manager.compareResults(groundTruth, actual)
	assert.False(t, isCorrect)

	// Test with nil inputs
	isCorrect = manager.compareResults(nil, actual)
	assert.False(t, isCorrect)

	isCorrect = manager.compareResults(groundTruth, nil)
	assert.False(t, isCorrect)
}

func TestCalculateConfidenceScore(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	// Test with matching results
	groundTruth := &VerificationResult{
		Status:       StatusPassed,
		OverallScore: 0.9,
	}
	actual := &VerificationResult{
		Status:       StatusPassed,
		OverallScore: 0.9,
	}

	score := manager.calculateConfidenceScore(groundTruth, actual)
	assert.Equal(t, 1.0, score) // Perfect match

	// Test with different status
	actual.Status = StatusFailed
	score = manager.calculateConfidenceScore(groundTruth, actual)
	assert.LessOrEqual(t, score, 0.3) // Only confidence component

	// Test with nil inputs
	score = manager.calculateConfidenceScore(nil, actual)
	assert.Equal(t, 0.0, score)

	score = manager.calculateConfidenceScore(groundTruth, nil)
	assert.Equal(t, 0.0, score)
}

func TestCalculateBenchmarkMetrics(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	// Create test results
	testResults := []*TestResult{
		{
			TestCaseID:      "test-1",
			Status:          "passed",
			ExecutionTime:   100 * time.Millisecond,
			IsCorrect:       true,
			ConfidenceScore: 1.0,
			ActualOutput: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.9,
			},
		},
		{
			TestCaseID:      "test-2",
			Status:          "failed",
			ExecutionTime:   150 * time.Millisecond,
			IsCorrect:       false,
			ConfidenceScore: 0.3,
			ActualOutput: &VerificationResult{
				Status:       StatusFailed,
				OverallScore: 0.3,
			},
		},
		{
			TestCaseID:      "test-3",
			Status:          "passed",
			ExecutionTime:   120 * time.Millisecond,
			IsCorrect:       true,
			ConfidenceScore: 0.9,
			ActualOutput: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.85,
			},
		},
	}

	testCases := []*BenchmarkTestCase{
		{
			GroundTruth: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.9,
			},
		},
		{
			GroundTruth: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.8,
			},
		},
		{
			GroundTruth: &VerificationResult{
				Status:       StatusPassed,
				OverallScore: 0.85,
			},
		},
	}

	metrics := manager.calculateBenchmarkMetrics(testResults, testCases)

	assert.NotNil(t, metrics)
	assert.Equal(t, 2.0/3.0, metrics.Accuracy) // 2 correct out of 3
	assert.Equal(t, 0.0, metrics.ErrorRate)    // No errors
	assert.Equal(t, 1.0, metrics.SuccessRate)  // All executed successfully
	assert.Greater(t, metrics.AverageLatency, time.Duration(0))
	assert.Greater(t, metrics.ThroughputPerSec, 0.0)
	assert.NotNil(t, metrics.ConfusionMatrix)

	// Test with empty results
	emptyMetrics := manager.calculateBenchmarkMetrics([]*TestResult{}, []*BenchmarkTestCase{})
	assert.NotNil(t, emptyMetrics)
}

func TestUpdateConfusionMatrix(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	matrix := &ConfusionMatrix{}

	// Test True Positive
	testCase := &BenchmarkTestCase{
		GroundTruth: &VerificationResult{
			Status: StatusPassed,
		},
	}
	result := &TestResult{
		ActualOutput: &VerificationResult{
			Status: StatusPassed,
		},
	}

	manager.updateConfusionMatrix(matrix, testCase, result)
	assert.Equal(t, 1, matrix.TruePositives)

	// Test False Negative
	result.ActualOutput = &VerificationResult{
		Status: StatusFailed,
	}

	manager.updateConfusionMatrix(matrix, testCase, result)
	assert.Equal(t, 1, matrix.FalseNegatives)

	// Test True Negative
	testCase.GroundTruth.Status = StatusFailed
	manager.updateConfusionMatrix(matrix, testCase, result)
	assert.Equal(t, 1, matrix.TrueNegatives)

	// Test False Positive
	result.ActualOutput = &VerificationResult{
		Status: StatusPassed,
	}
	manager.updateConfusionMatrix(matrix, testCase, result)
	assert.Equal(t, 1, matrix.FalsePositives)
}

func TestCalculateMetricComponents(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	matrix := &ConfusionMatrix{
		TruePositives:  10,
		TrueNegatives:  5,
		FalsePositives: 2,
		FalseNegatives: 3,
	}

	// Test precision calculation
	precision := manager.calculatePrecision(matrix)
	expected := 10.0 / (10.0 + 2.0) // TP / (TP + FP)
	assert.InDelta(t, expected, precision, 0.001)

	// Test recall calculation
	recall := manager.calculateRecall(matrix)
	expected = 10.0 / (10.0 + 3.0) // TP / (TP + FN)
	assert.InDelta(t, expected, recall, 0.001)

	// Test F1 score calculation
	f1Score := manager.calculateF1Score(precision, recall)
	expected = 2 * (precision * recall) / (precision + recall)
	assert.InDelta(t, expected, f1Score, 0.001)

	// Test specificity calculation
	specificity := manager.calculateSpecificity(matrix)
	expected = 5.0 / (5.0 + 2.0) // TN / (TN + FP)
	assert.InDelta(t, expected, specificity, 0.001)

	// Test with zero denominators
	zeroMatrix := &ConfusionMatrix{}

	precision = manager.calculatePrecision(zeroMatrix)
	assert.Equal(t, 0.0, precision)

	recall = manager.calculateRecall(zeroMatrix)
	assert.Equal(t, 0.0, recall)

	f1Score = manager.calculateF1Score(0, 0)
	assert.Equal(t, 0.0, f1Score)
}

func TestGetBenchmarkSuite(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	// Create test suite
	suite := &BenchmarkSuite{
		Name:        "Test Suite",
		Description: "A test benchmark suite",
		Category:    "verification",
	}

	err := manager.CreateBenchmarkSuite(suite)
	assert.NoError(t, err)

	// Test getting existing suite
	retrieved, err := manager.GetBenchmarkSuite(suite.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, suite.ID, retrieved.ID)
	assert.Equal(t, suite.Name, retrieved.Name)

	// Test getting non-existent suite
	_, err = manager.GetBenchmarkSuite("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "benchmark suite not found")
}

func TestListBenchmarkSuites(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	// Initially empty
	suites := manager.ListBenchmarkSuites()
	assert.Empty(t, suites)

	// Create test suites
	suite1 := &BenchmarkSuite{
		Name:        "Suite A",
		Description: "First test suite",
		Category:    "verification",
	}
	suite2 := &BenchmarkSuite{
		Name:        "Suite B",
		Description: "Second test suite",
		Category:    "performance",
	}

	err := manager.CreateBenchmarkSuite(suite1)
	assert.NoError(t, err)

	err = manager.CreateBenchmarkSuite(suite2)
	assert.NoError(t, err)

	// List all suites
	suites = manager.ListBenchmarkSuites()
	assert.Len(t, suites, 2)

	// Check sorting by name
	assert.Equal(t, "Suite A", suites[0].Name)
	assert.Equal(t, "Suite B", suites[1].Name)
}

func TestGetBenchmarkResults(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	// Initially empty
	results := manager.GetBenchmarkResults(10)
	assert.Empty(t, results)

	// Create and run benchmark to generate results
	suite := &BenchmarkSuite{
		Name:        "Test Suite",
		Description: "A test benchmark suite",
		Category:    "verification",
		TestCases: []*BenchmarkTestCase{
			{
				Name:           "Test Case 1",
				Description:    "First test case",
				Input:          "test input",
				ExpectedOutput: "test output",
				GroundTruth: &VerificationResult{
					Status:       StatusPassed,
					OverallScore: 0.9,
				},
				Weight:     1.0,
				Difficulty: "easy",
			},
		},
	}

	err := manager.CreateBenchmarkSuite(suite)
	assert.NoError(t, err)

	_, err = manager.RunBenchmark(context.Background(), suite.ID)
	assert.NoError(t, err)

	// Get results
	results = manager.GetBenchmarkResults(10)
	assert.Len(t, results, 1)

	// Test limit
	results = manager.GetBenchmarkResults(0)
	assert.Len(t, results, 1) // Should return all when limit is 0
}

func TestCompareBenchmarks(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	// Create and run benchmarks to generate results
	suite := &BenchmarkSuite{
		Name:        "Test Suite",
		Description: "A test benchmark suite",
		Category:    "verification",
		TestCases: []*BenchmarkTestCase{
			{
				Name:           "Test Case 1",
				Description:    "First test case",
				Input:          "test input",
				ExpectedOutput: "test output",
				GroundTruth: &VerificationResult{
					Status:       StatusPassed,
					OverallScore: 0.9,
				},
				Weight:     1.0,
				Difficulty: "easy",
			},
		},
	}

	err := manager.CreateBenchmarkSuite(suite)
	assert.NoError(t, err)

	result1, err := manager.RunBenchmark(context.Background(), suite.ID)
	assert.NoError(t, err)

	result2, err := manager.RunBenchmark(context.Background(), suite.ID)
	assert.NoError(t, err)

	// Compare benchmarks
	comparison, err := manager.CompareBenchmarks(result1.ID, result2.ID)
	assert.NoError(t, err)
	assert.NotNil(t, comparison)

	assert.Equal(t, result1.ID, comparison.BaselineID)
	assert.Equal(t, result2.ID, comparison.ComparisonID)
	assert.NotEmpty(t, comparison.Summary)
	assert.NotNil(t, comparison.Improvements)
	assert.NotNil(t, comparison.Regressions)
	assert.NotNil(t, comparison.Significance)
	assert.NotNil(t, comparison.Recommendations)

	// Test with non-existent benchmarks
	_, err = manager.CompareBenchmarks("non-existent", result2.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "baseline benchmark not found")

	_, err = manager.CompareBenchmarks(result1.ID, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "comparison benchmark not found")
}

func TestUpdateBenchmarkConfig(t *testing.T) {
	logger := zap.NewNop()
	manager := NewVerificationBenchmarkManager(nil, logger)

	// Test valid config update
	newConfig := &BenchmarkConfig{
		EnableBenchmarking: false,
		AccuracyThreshold:  0.95,
		BenchmarkInterval:  2 * time.Hour,
		MinSampleSize:      200,
		ConfidenceLevel:    0.99,
	}

	err := manager.UpdateConfig(newConfig)
	assert.NoError(t, err)

	updatedConfig := manager.GetConfig()
	assert.False(t, updatedConfig.EnableBenchmarking)
	assert.Equal(t, 0.95, updatedConfig.AccuracyThreshold)
	assert.Equal(t, 2*time.Hour, updatedConfig.BenchmarkInterval)
	assert.Equal(t, 200, updatedConfig.MinSampleSize)
	assert.Equal(t, 0.99, updatedConfig.ConfidenceLevel)

	// Test invalid config (nil)
	err = manager.UpdateConfig(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")

	// Test invalid config (accuracy threshold out of range)
	invalidConfig := &BenchmarkConfig{
		AccuracyThreshold: 1.5, // Invalid: > 1
	}

	err = manager.UpdateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accuracy threshold must be between 0 and 1")

	// Test invalid config (confidence level out of range)
	invalidConfig2 := &BenchmarkConfig{
		ConfidenceLevel: -0.1, // Invalid: < 0
	}

	err = manager.UpdateConfig(invalidConfig2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "confidence level must be between 0 and 1")

	// Test invalid config (benchmark interval too short)
	invalidConfig3 := &BenchmarkConfig{
		BenchmarkInterval: 30 * time.Second, // Invalid: < 1 minute
	}

	err = manager.UpdateConfig(invalidConfig3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "benchmark interval must be at least 1 minute")
}

func TestBenchmarkStructsValidation(t *testing.T) {
	// Test BenchmarkSuite struct
	suite := &BenchmarkSuite{
		ID:          "test-suite-1",
		Name:        "Test Suite",
		Description: "A test benchmark suite",
		Category:    "verification",
		TestCases:   []*BenchmarkTestCase{},
		Metrics:     &BenchmarkMetrics{},
		Config:      map[string]interface{}{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.Equal(t, "test-suite-1", suite.ID)
	assert.Equal(t, "Test Suite", suite.Name)
	assert.Equal(t, "verification", suite.Category)
	assert.NotNil(t, suite.TestCases)
	assert.NotNil(t, suite.Metrics)
	assert.NotNil(t, suite.Config)

	// Test BenchmarkTestCase struct
	testCase := &BenchmarkTestCase{
		ID:             "test-case-1",
		Name:           "Test Case",
		Description:    "A test case",
		Input:          "test input",
		ExpectedOutput: "test output",
		GroundTruth:    &VerificationResult{},
		Metadata:       map[string]interface{}{},
		Weight:         1.0,
		Difficulty:     "medium",
		Tags:           []string{"tag1", "tag2"},
		CreatedAt:      time.Now(),
	}

	assert.Equal(t, "test-case-1", testCase.ID)
	assert.Equal(t, "Test Case", testCase.Name)
	assert.Equal(t, 1.0, testCase.Weight)
	assert.Equal(t, "medium", testCase.Difficulty)
	assert.Len(t, testCase.Tags, 2)

	// Test ConfusionMatrix struct
	matrix := &ConfusionMatrix{
		TruePositives:  10,
		TrueNegatives:  8,
		FalsePositives: 2,
		FalseNegatives: 3,
	}

	assert.Equal(t, 10, matrix.TruePositives)
	assert.Equal(t, 8, matrix.TrueNegatives)
	assert.Equal(t, 2, matrix.FalsePositives)
	assert.Equal(t, 3, matrix.FalseNegatives)
}

func TestBenchmarkHelperFunctions(t *testing.T) {
	// Test benchmark ID generation
	id1 := generateBenchmarkID()
	id2 := generateBenchmarkID()

	assert.Contains(t, id1, "benchmark_")
	assert.Contains(t, id2, "benchmark_")
	assert.NotEqual(t, id1, id2)

	// Test result ID generation
	resultID1 := generateResultID()
	resultID2 := generateResultID()

	assert.Contains(t, resultID1, "result_")
	assert.Contains(t, resultID2, "result_")
	assert.NotEqual(t, resultID1, resultID2)

	// Test that IDs have reasonable length
	assert.Greater(t, len(id1), 15)
	assert.Greater(t, len(id2), 15)
	assert.Greater(t, len(resultID1), 10)
	assert.Greater(t, len(resultID2), 10)
}
