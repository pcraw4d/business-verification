package classification_optimization

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewAccuracyValidator(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	assert.NotNil(t, validator)
	assert.NotNil(t, validator.logger)
	assert.NotNil(t, validator.performanceTracker)
	assert.NotNil(t, validator.algorithmRegistry)
	assert.Equal(t, 0, len(validator.validationHistory))
	assert.Equal(t, 0, len(validator.activeValidations))
}

func TestAccuracyValidator_SetPerformanceTracker(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)
	tracker := NewPerformanceTracker(logger)

	validator.SetPerformanceTracker(tracker)
	assert.Equal(t, tracker, validator.performanceTracker)
}

func TestAccuracyValidator_SetAlgorithmRegistry(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)
	registry := NewAlgorithmRegistry(logger)

	validator.SetAlgorithmRegistry(registry)
	assert.Equal(t, registry, validator.algorithmRegistry)
}

func TestAccuracyValidator_ValidateAccuracy_Success(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	// Register a test algorithm
	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	validator.algorithmRegistry.RegisterAlgorithm(algorithm)

	// Create test cases (need at least 100 for default config)
	testCases := make([]*TestCase, 100)
	for i := 0; i < 100; i++ {
		testCases[i] = &TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		}
	}

	// Perform validation
	result, err := validator.ValidateAccuracy(context.Background(), "test-algorithm", testCases)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-algorithm", result.AlgorithmID)
	assert.Equal(t, ValidationTypeAccuracy, result.ValidationType)
	assert.Equal(t, ValidationStatusCompleted, result.Status)
	assert.Len(t, result.TestCases, 100)
	assert.NotNil(t, result.Metrics)
	assert.NotNil(t, result.Recommendations)
	assert.NotNil(t, result.CompletionTime)
}

func TestAccuracyValidator_ValidateAccuracy_InsufficientTestCases(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	// Create insufficient test cases
	testCases := []*TestCase{
		{
			ID:             "test-1",
			Input:          map[string]interface{}{"name": "Test Company"},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		},
	}

	// Perform validation
	result, err := validator.ValidateAccuracy(context.Background(), "test-algorithm", testCases)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient test cases")
	assert.NotNil(t, result)
	assert.Equal(t, ValidationStatusFailed, result.Status)
}

func TestAccuracyValidator_ValidateAccuracy_AlgorithmNotFound(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	// Create test cases
	testCases := make([]*TestCase, 100)
	for i := 0; i < 100; i++ {
		testCases[i] = &TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		}
	}

	// Perform validation with non-existent algorithm
	result, err := validator.ValidateAccuracy(context.Background(), "non-existent", testCases)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "algorithm not found")
	assert.NotNil(t, result)
	assert.Equal(t, ValidationStatusFailed, result.Status)
}

func TestAccuracyValidator_PerformCrossValidation_Success(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	// Register a test algorithm
	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	validator.algorithmRegistry.RegisterAlgorithm(algorithm)

	// Create test cases (enough for 5-fold cross validation)
	testCases := make([]*TestCase, 100)
	for i := 0; i < 100; i++ {
		testCases[i] = &TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		}
	}

	// Perform cross validation
	result, err := validator.PerformCrossValidation(context.Background(), "test-algorithm", testCases)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-algorithm", result.AlgorithmID)
	assert.Equal(t, ValidationTypeCrossValidation, result.ValidationType)
	assert.Equal(t, ValidationStatusCompleted, result.Status)
	assert.Len(t, result.TestCases, 100)
	assert.NotNil(t, result.Metrics)
	assert.NotNil(t, result.Recommendations)
	assert.NotNil(t, result.CompletionTime)
}

func TestAccuracyValidator_PerformCrossValidation_InsufficientTestCases(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	// Register a test algorithm
	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	validator.algorithmRegistry.RegisterAlgorithm(algorithm)

	// Create insufficient test cases for 5-fold cross validation
	testCases := make([]*TestCase, 3)
	for i := 0; i < 3; i++ {
		testCases[i] = &TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		}
	}

	// Perform cross validation
	result, err := validator.PerformCrossValidation(context.Background(), "test-algorithm", testCases)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient test cases for cross validation")
	assert.NotNil(t, result)
	assert.Equal(t, ValidationStatusFailed, result.Status)
}

func TestAccuracyValidator_GetValidationHistory(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	// Initially empty
	history := validator.GetValidationHistory()
	assert.Len(t, history, 0)

	// Register algorithm and perform validation
	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	validator.algorithmRegistry.RegisterAlgorithm(algorithm)

	testCases := make([]*TestCase, 100)
	for i := 0; i < 100; i++ {
		testCases[i] = &TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		}
	}

	_, err := validator.ValidateAccuracy(context.Background(), "test-algorithm", testCases)
	assert.NoError(t, err)

	// Check history
	history = validator.GetValidationHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, "test-algorithm", history[0].AlgorithmID)
	assert.Equal(t, ValidationStatusCompleted, history[0].Status)
}

func TestAccuracyValidator_GetActiveValidations(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	// Initially empty
	active := validator.GetActiveValidations()
	assert.Len(t, active, 0)

	// Register algorithm
	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	validator.algorithmRegistry.RegisterAlgorithm(algorithm)

	testCases := make([]*TestCase, 100)
	for i := 0; i < 100; i++ {
		testCases[i] = &TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		}
	}

	// Start validation (it will complete immediately in this test)
	_, err := validator.ValidateAccuracy(context.Background(), "test-algorithm", testCases)
	assert.NoError(t, err)

	// Check active validations (should be empty after completion)
	active = validator.GetActiveValidations()
	assert.Len(t, active, 0)
}

func TestAccuracyValidator_GetValidationSummary(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	// Initially empty summary
	summary := validator.GetValidationSummary()
	assert.Equal(t, 0, summary.TotalValidations)
	assert.Equal(t, 0, summary.ActiveValidations)
	assert.Equal(t, 0.0, summary.AverageAccuracy)
	assert.Equal(t, 0.0, summary.AverageF1Score)

	// Register algorithm and perform validation
	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmRegistry := NewAlgorithmRegistry(zap.NewNop())
	algorithmRegistry.RegisterAlgorithm(algorithm)
	validator.SetAlgorithmRegistry(algorithmRegistry)

	testCases := make([]*TestCase, 100)
	for i := 0; i < 100; i++ {
		testCases[i] = &TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		}
	}

	_, err := validator.ValidateAccuracy(context.Background(), "test-algorithm", testCases)
	assert.NoError(t, err)

	// Check summary
	summary = validator.GetValidationSummary()
	assert.Equal(t, 1, summary.TotalValidations)
	assert.Equal(t, 0, summary.ActiveValidations)
	assert.Equal(t, 1, summary.ValidationsByType["accuracy"])
	assert.Equal(t, 1, summary.ValidationsByStatus["completed"])
	assert.GreaterOrEqual(t, summary.AverageAccuracy, 0.0)
	assert.GreaterOrEqual(t, summary.AverageF1Score, 0.0)
}

func TestAccuracyValidator_executeTestCases(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}

	testCases := []*TestCase{
		{
			ID:             "test-1",
			Input:          map[string]interface{}{"name": "Microsoft Corporation"},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		},
		{
			ID:             "test-2",
			Input:          map[string]interface{}{"name": "Walmart Store"},
			ExpectedOutput: "retail",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		},
	}

	metrics, err := validator.executeTestCases(context.Background(), algorithm, testCases)

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 2, metrics.TotalTestCases)
	assert.GreaterOrEqual(t, metrics.PassedTestCases, 0)
	assert.GreaterOrEqual(t, metrics.FailedTestCases, 0)
	assert.GreaterOrEqual(t, metrics.Accuracy, 0.0)
	assert.LessOrEqual(t, metrics.Accuracy, 1.0)
	assert.GreaterOrEqual(t, metrics.AverageConfidence, 0.0)
	assert.LessOrEqual(t, metrics.AverageConfidence, 1.0)
}

func TestAccuracyValidator_performRegressionAnalysis(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	// No previous validations
	currentMetrics := &ValidationMetrics{
		Accuracy: 0.85,
		F1Score:  0.83,
	}

	regression := validator.performRegressionAnalysis("test-algorithm", currentMetrics)
	assert.NotNil(t, regression)
	assert.Equal(t, 0.85, regression.CurrentAccuracy)
	assert.Equal(t, 0.0, regression.AccuracyChange)

	// Add a previous validation
	previousResult := &ValidationResult{
		ID:          "prev-val",
		AlgorithmID: "test-algorithm",
		Status:      ValidationStatusCompleted,
		Metrics: &ValidationMetrics{
			Accuracy: 0.80,
			F1Score:  0.78,
		},
	}

	validator.mu.Lock()
	validator.validationHistory = append(validator.validationHistory, previousResult)
	validator.mu.Unlock()

	// Perform regression analysis
	regression = validator.performRegressionAnalysis("test-algorithm", currentMetrics)
	assert.NotNil(t, regression)
	assert.Equal(t, 0.80, regression.PreviousAccuracy)
	assert.Equal(t, 0.85, regression.CurrentAccuracy)
	assert.InDelta(t, 0.05, regression.AccuracyChange, 0.0001)
	assert.True(t, regression.AccuracyImprovement)
	assert.False(t, regression.RegressionDetected)
}

func TestAccuracyValidator_generateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	// Test low accuracy recommendation
	lowAccuracyMetrics := &ValidationMetrics{
		Accuracy:              0.65,
		ConfidenceCorrelation: 0.6,
	}

	recommendations := validator.generateRecommendations(lowAccuracyMetrics, nil)
	assert.Len(t, recommendations, 2) // Low accuracy + low confidence correlation

	// Test regression recommendation
	regression := &RegressionAnalysis{
		PreviousAccuracy:   0.85,
		CurrentAccuracy:    0.75,
		AccuracyChange:     -0.10,
		RegressionDetected: true,
	}

	recommendations = validator.generateRecommendations(lowAccuracyMetrics, regression)
	assert.Len(t, recommendations, 3) // Low accuracy + low confidence correlation + regression
}

func TestAccuracyValidator_aggregateCrossValidationMetrics(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	metrics := []*ValidationMetrics{
		{
			TotalTestCases:        20,
			PassedTestCases:       16,
			FailedTestCases:       4,
			Accuracy:              0.8,
			F1Score:               0.78,
			AverageConfidence:     0.75,
			AverageProcessingTime: 100.0,
		},
		{
			TotalTestCases:        20,
			PassedTestCases:       18,
			FailedTestCases:       2,
			Accuracy:              0.9,
			F1Score:               0.88,
			AverageConfidence:     0.85,
			AverageProcessingTime: 95.0,
		},
	}

	aggregated := validator.aggregateCrossValidationMetrics(metrics)
	assert.NotNil(t, aggregated)
	assert.Equal(t, 40, aggregated.TotalTestCases)
	assert.Equal(t, 34, aggregated.PassedTestCases)
	assert.Equal(t, 6, aggregated.FailedTestCases)
	assert.InDelta(t, 0.85, aggregated.Accuracy, 0.0001)              // Average of 0.8 and 0.9
	assert.InDelta(t, 0.83, aggregated.F1Score, 0.0001)               // Average of 0.78 and 0.88
	assert.Equal(t, 0.8, aggregated.AverageConfidence)      // Average of 0.75 and 0.85
	assert.Equal(t, 97.5, aggregated.AverageProcessingTime) // Average of 100 and 95
}

func TestAccuracyValidator_mockClassification(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	algorithm := &ClassificationAlgorithm{
		ConfidenceThreshold: 0.7,
	}

	// Test long name -> technology
	result := validator.mockClassification(algorithm, map[string]interface{}{"name": "Microsoft Corporation"})
	assert.Equal(t, "technology", result)

	// Test medium name -> retail
	result = validator.mockClassification(algorithm, map[string]interface{}{"name": "Walmart"})
	assert.Equal(t, "retail", result)

	// Test short name -> other
	result = validator.mockClassification(algorithm, map[string]interface{}{"name": "ABC"})
	assert.Equal(t, "other", result)
}

func TestAccuracyValidator_calculateConfidence(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	algorithm := &ClassificationAlgorithm{
		ConfidenceThreshold: 0.7,
	}

	// Test long name (should increase confidence)
	confidence := validator.calculateConfidence(algorithm, map[string]interface{}{"name": "Microsoft Corporation"}, "technology")
	assert.Greater(t, confidence, 0.7)

	// Test short name (should use base confidence)
	confidence = validator.calculateConfidence(algorithm, map[string]interface{}{"name": "ABC"}, "other")
	assert.Equal(t, 0.7, confidence)
}

func TestAccuracyValidator_calculateConfidenceCorrelation(t *testing.T) {
	logger := zap.NewNop()
	validator := NewAccuracyValidator(nil, logger)

	testCases := []*TestCase{
		{IsCorrect: true},
		{IsCorrect: true},
		{IsCorrect: false},
		{IsCorrect: false},
	}

	confidences := []float64{0.9, 0.8, 0.3, 0.4}

	correlation := validator.calculateConfidenceCorrelation(testCases, confidences)
	assert.Greater(t, correlation, 0.0)
	assert.LessOrEqual(t, correlation, 1.0)
}

func TestValidationResult_Validation(t *testing.T) {
	// Test valid validation result
	result := &ValidationResult{
		ID:             "test-val",
		AlgorithmID:    "test-algorithm",
		ValidationType: ValidationTypeAccuracy,
		Status:         ValidationStatusCompleted,
		ValidationTime: time.Now(),
	}

	assert.NotEmpty(t, result.ID)
	assert.NotEmpty(t, result.AlgorithmID)
	assert.Equal(t, ValidationTypeAccuracy, result.ValidationType)
	assert.Equal(t, ValidationStatusCompleted, result.Status)
}

func TestValidationMetrics_Validation(t *testing.T) {
	// Test valid validation metrics
	metrics := &ValidationMetrics{
		TotalTestCases:        100,
		PassedTestCases:       85,
		FailedTestCases:       15,
		Accuracy:              0.85,
		Precision:             0.85,
		Recall:                0.85,
		F1Score:               0.85,
		AverageConfidence:     0.8,
		AverageProcessingTime: 50.0,
		ErrorRate:             0.15,
		ConfidenceCorrelation: 0.75,
	}

	assert.Equal(t, 100, metrics.TotalTestCases)
	assert.Equal(t, 85, metrics.PassedTestCases)
	assert.Equal(t, 15, metrics.FailedTestCases)
	assert.Equal(t, 0.85, metrics.Accuracy)
	assert.Equal(t, 0.85, metrics.Precision)
	assert.Equal(t, 0.85, metrics.Recall)
	assert.Equal(t, 0.85, metrics.F1Score)
	assert.Equal(t, 0.8, metrics.AverageConfidence)
	assert.Equal(t, 50.0, metrics.AverageProcessingTime)
	assert.Equal(t, 0.15, metrics.ErrorRate)
	assert.Equal(t, 0.75, metrics.ConfidenceCorrelation)
}

func TestTestCase_Validation(t *testing.T) {
	// Test valid test case
	testCase := &TestCase{
		ID:             "test-1",
		Input:          map[string]interface{}{"name": "Test Company"},
		ExpectedOutput: "technology",
		ActualOutput:   "technology",
		Confidence:     0.85,
		ProcessingTime: 50 * time.Millisecond,
		IsCorrect:      true,
		TestCaseType:   "standard",
		Difficulty:     "easy",
	}

	assert.NotEmpty(t, testCase.ID)
	assert.NotNil(t, testCase.Input)
	assert.NotEmpty(t, testCase.ExpectedOutput)
	assert.NotEmpty(t, testCase.ActualOutput)
	assert.Greater(t, testCase.Confidence, 0.0)
	assert.LessOrEqual(t, testCase.Confidence, 1.0)
	assert.True(t, testCase.IsCorrect)
	assert.NotEmpty(t, testCase.TestCaseType)
	assert.NotEmpty(t, testCase.Difficulty)
}
