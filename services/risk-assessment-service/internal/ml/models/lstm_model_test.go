package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

func TestNewLSTMModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	assert.NotNil(t, model)
	assert.Equal(t, "test_model", model.name)
	assert.Equal(t, "1.0.0", model.version)
	assert.False(t, model.trained)
	assert.Equal(t, 12, model.sequenceLength)
	assert.Equal(t, []int{6, 9, 12}, model.predictionHorizons)
	assert.NotNil(t, model.featureExtractor)
	assert.NotNil(t, model.riskLevelEncoder)
	assert.NotNil(t, model.temporalBuilder)
	assert.NotNil(t, model.logger)
}

func TestLSTMModel_LoadModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	tests := []struct {
		name        string
		modelPath   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "non-existent model file",
			modelPath:   "./models/non_existent.onnx",
			expectError: true,
			errorMsg:    "failed to load ONNX model",
		},
		{
			name:        "empty model path",
			modelPath:   "",
			expectError: true,
			errorMsg:    "model path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := model.LoadModel(ctx, tt.modelPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLSTMModel_Predict(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	// Create a test business request
	business := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		Phone:             "+1-555-123-4567",
		Email:             "test@testcompany.com",
		Website:           "https://testcompany.com",
		PredictionHorizon: 6,
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	t.Run("successful prediction", func(t *testing.T) {
		ctx := context.Background()
		result, err := model.Predict(ctx, business)

		// Since we're using mock implementation, this should succeed
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, business.BusinessName, result.BusinessName)
		assert.Equal(t, business.BusinessAddress, result.BusinessAddress)
		assert.Equal(t, business.Industry, result.Industry)
		assert.Equal(t, business.Country, result.Country)
		assert.GreaterOrEqual(t, result.RiskScore, 0.0)
		assert.LessOrEqual(t, result.RiskScore, 1.0)
		assert.NotEmpty(t, result.RiskLevel)
		assert.GreaterOrEqual(t, result.ConfidenceScore, 0.0)
		assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
		assert.Equal(t, models.StatusCompleted, result.Status)
		assert.NotEmpty(t, result.ID)
		assert.NotZero(t, result.CreatedAt)
		assert.NotZero(t, result.UpdatedAt)
		assert.Contains(t, result.Metadata, "model_type")
		assert.Equal(t, "lstm", result.Metadata["model_type"])
		assert.Contains(t, result.Metadata, "model_version")
		assert.Equal(t, "1.0.0", result.Metadata["model_version"])
	})

	t.Run("nil business request", func(t *testing.T) {
		ctx := context.Background()
		result, err := model.Predict(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "business request cannot be nil")
	})

	t.Run("empty business name", func(t *testing.T) {
		ctx := context.Background()
		businessCopy := *business
		businessCopy.BusinessName = ""

		result, err := model.Predict(ctx, &businessCopy)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "business name is required")
	})
}

func TestLSTMModel_PredictFuture(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 6,
	}

	tests := []struct {
		name          string
		horizonMonths int
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "valid 6 month horizon",
			horizonMonths: 6,
			expectError:   false,
		},
		{
			name:          "valid 9 month horizon",
			horizonMonths: 9,
			expectError:   false,
		},
		{
			name:          "valid 12 month horizon",
			horizonMonths: 12,
			expectError:   false,
		},
		{
			name:          "invalid negative horizon",
			horizonMonths: -1,
			expectError:   true,
			errorMsg:      "horizon must be positive",
		},
		{
			name:          "invalid zero horizon",
			horizonMonths: 0,
			expectError:   true,
			errorMsg:      "horizon must be positive",
		},
		{
			name:          "invalid large horizon",
			horizonMonths: 25,
			expectError:   true,
			errorMsg:      "horizon exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := model.PredictFuture(ctx, business, tt.horizonMonths)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.horizonMonths, result.HorizonMonths)
				assert.GreaterOrEqual(t, result.PredictedScore, 0.0)
				assert.LessOrEqual(t, result.PredictedScore, 1.0)
				assert.NotEmpty(t, result.PredictedLevel)
				assert.GreaterOrEqual(t, result.ConfidenceScore, 0.0)
				assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
				assert.NotZero(t, result.PredictionDate)
				assert.NotZero(t, result.CreatedAt)
				assert.NotEmpty(t, result.RiskFactors)
				assert.NotEmpty(t, result.ScenarioAnalysis)
			}
		})
	}
}

func TestLSTMModel_GetModelInfo(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	info := model.GetModelInfo()

	assert.NotNil(t, info)
	assert.Equal(t, "test_model", info.Name)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "lstm", info.Type)
	assert.NotZero(t, info.TrainingDate)
	assert.GreaterOrEqual(t, info.Accuracy, 0.0)
	assert.LessOrEqual(t, info.Accuracy, 1.0)
	assert.NotEmpty(t, info.Features)
	assert.NotEmpty(t, info.Hyperparameters)
	assert.NotEmpty(t, info.Metadata)
}

func TestLSTMModel_SaveModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	tests := []struct {
		name        string
		modelPath   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty model path",
			modelPath:   "",
			expectError: true,
			errorMsg:    "model path cannot be empty",
		},
		{
			name:        "valid model path",
			modelPath:   "./models/test_save.onnx",
			expectError: true, // Will fail because model is not trained
			errorMsg:    "model is not trained",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := model.SaveModel(ctx, tt.modelPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLSTMModel_ValidateModel(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	// Create test data
	testData := []*models.RiskAssessment{
		{
			ID:              "test_1",
			BusinessID:      "biz_1",
			BusinessName:    "Test Company 1",
			RiskScore:       0.3,
			RiskLevel:       models.RiskLevelLow,
			ConfidenceScore: 0.8,
			Status:          models.StatusCompleted,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "test_2",
			BusinessID:      "biz_2",
			BusinessName:    "Test Company 2",
			RiskScore:       0.7,
			RiskLevel:       models.RiskLevelHigh,
			ConfidenceScore: 0.9,
			Status:          models.StatusCompleted,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	t.Run("successful validation", func(t *testing.T) {
		ctx := context.Background()
		result, err := model.ValidateModel(ctx, testData)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, result.Accuracy, 0.0)
		assert.LessOrEqual(t, result.Accuracy, 1.0)
		assert.GreaterOrEqual(t, result.Precision, 0.0)
		assert.LessOrEqual(t, result.Precision, 1.0)
		assert.GreaterOrEqual(t, result.Recall, 0.0)
		assert.LessOrEqual(t, result.Recall, 1.0)
		assert.GreaterOrEqual(t, result.F1Score, 0.0)
		assert.LessOrEqual(t, result.F1Score, 1.0)
		assert.NotZero(t, result.ValidationDate)
	})

	t.Run("empty test data", func(t *testing.T) {
		ctx := context.Background()
		result, err := model.ValidateModel(ctx, []*models.RiskAssessment{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "test data cannot be empty")
	})

	t.Run("nil test data", func(t *testing.T) {
		ctx := context.Background()
		result, err := model.ValidateModel(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "test data cannot be nil")
	})
}

func TestLSTMModel_generateMockRiskScore(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	tests := []struct {
		name        string
		inputTensor []float64
		horizon     int
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "empty input tensor",
			inputTensor: []float64{},
			horizon:     6,
			expectedMin: 0.0,
			expectedMax: 1.0,
		},
		{
			name:        "single value input",
			inputTensor: []float64{0.5},
			horizon:     6,
			expectedMin: 0.0,
			expectedMax: 1.0,
		},
		{
			name:        "multiple values input",
			inputTensor: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
			horizon:     6,
			expectedMin: 0.0,
			expectedMax: 1.0,
		},
		{
			name:        "longer horizon",
			inputTensor: []float64{0.1, 0.2, 0.3},
			horizon:     12,
			expectedMin: 0.0,
			expectedMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert []float64 to []float32 for the method
			inputTensor32 := make([]float32, len(tt.inputTensor))
			for i, v := range tt.inputTensor {
				inputTensor32[i] = float32(v)
			}
			score := model.generateMockRiskScore(inputTensor32, tt.horizon)

			assert.GreaterOrEqual(t, score, tt.expectedMin)
			assert.LessOrEqual(t, score, tt.expectedMax)
		})
	}
}

func TestLSTMModel_generateMockConfidence(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	tests := []struct {
		name        string
		inputTensor []float64
		horizon     int
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "empty input tensor",
			inputTensor: []float64{},
			horizon:     6,
			expectedMin: 0.1,
			expectedMax: 1.0,
		},
		{
			name:        "single value input",
			inputTensor: []float64{0.5},
			horizon:     6,
			expectedMin: 0.1,
			expectedMax: 1.0,
		},
		{
			name:        "multiple values input",
			inputTensor: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
			horizon:     6,
			expectedMin: 0.1,
			expectedMax: 1.0,
		},
		{
			name:        "longer horizon",
			inputTensor: []float64{0.1, 0.2, 0.3},
			horizon:     12,
			expectedMin: 0.1,
			expectedMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert []float64 to []float32 for the method
			inputTensor32 := make([]float32, len(tt.inputTensor))
			for i, v := range tt.inputTensor {
				inputTensor32[i] = float32(v)
			}
			confidence := model.generateMockConfidence(inputTensor32, tt.horizon)

			assert.GreaterOrEqual(t, confidence, tt.expectedMin)
			assert.LessOrEqual(t, confidence, tt.expectedMax)
		})
	}
}

func TestLSTMModel_generateLSTMRiskFactors(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	tests := []struct {
		name        string
		riskScore   float64
		confidence  float64
		expectCount int
	}{
		{
			name:        "low risk score",
			riskScore:   0.2,
			confidence:  0.8,
			expectCount: 5,
		},
		{
			name:        "medium risk score",
			riskScore:   0.5,
			confidence:  0.7,
			expectCount: 5,
		},
		{
			name:        "high risk score",
			riskScore:   0.8,
			confidence:  0.9,
			expectCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factors := model.generateLSTMRiskFactors(tt.riskScore, tt.confidence)

			assert.Len(t, factors, tt.expectCount)

			for _, factor := range factors {
				assert.NotEmpty(t, factor.Category)
				assert.NotEmpty(t, factor.Name)
				assert.GreaterOrEqual(t, factor.Score, 0.0)
				assert.LessOrEqual(t, factor.Score, 1.0)
				assert.GreaterOrEqual(t, factor.Weight, 0.0)
				assert.LessOrEqual(t, factor.Weight, 1.0)
				assert.NotEmpty(t, factor.Description)
				assert.Equal(t, "lstm", factor.Source)
				assert.GreaterOrEqual(t, factor.Confidence, 0.0)
				assert.LessOrEqual(t, factor.Confidence, 1.0)
			}
		})
	}
}

func TestLSTMModel_generateLSTMScenarioAnalysis(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	tests := []struct {
		name          string
		riskScore     float64
		horizonMonths int
		expectCount   int
	}{
		{
			name:          "6 month horizon",
			riskScore:     0.5,
			horizonMonths: 6,
			expectCount:   3,
		},
		{
			name:          "12 month horizon",
			riskScore:     0.7,
			horizonMonths: 12,
			expectCount:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenarios := model.generateLSTMScenarioAnalysis(tt.riskScore, tt.horizonMonths)

			assert.Len(t, scenarios, tt.expectCount)

			for _, scenario := range scenarios {
				assert.NotEmpty(t, scenario.ScenarioName)
				assert.NotEmpty(t, scenario.Description)
				assert.GreaterOrEqual(t, scenario.RiskScore, 0.0)
				assert.LessOrEqual(t, scenario.RiskScore, 1.0)
				assert.NotEmpty(t, scenario.RiskLevel)
				assert.GreaterOrEqual(t, scenario.Probability, 0.0)
				assert.LessOrEqual(t, scenario.Probability, 1.0)
				assert.NotEmpty(t, scenario.Impact)
			}
		})
	}
}

func TestLSTMModel_getClosestHorizonPrediction(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	// Create mock predictions
	predictions := map[string]interface{}{
		"horizon_6": map[string]float64{
			"risk_score": 0.3,
			"confidence": 0.8,
		},
		"horizon_9": map[string]float64{
			"risk_score": 0.4,
			"confidence": 0.7,
		},
		"horizon_12": map[string]float64{
			"risk_score": 0.5,
			"confidence": 0.6,
		},
	}

	tests := []struct {
		name            string
		targetHorizon   int
		expectedHorizon int
	}{
		{
			name:            "exact match 6 months",
			targetHorizon:   6,
			expectedHorizon: 6,
		},
		{
			name:            "exact match 9 months",
			targetHorizon:   9,
			expectedHorizon: 9,
		},
		{
			name:            "exact match 12 months",
			targetHorizon:   12,
			expectedHorizon: 12,
		},
		{
			name:            "closest to 6 months",
			targetHorizon:   7,
			expectedHorizon: 6,
		},
		{
			name:            "closest to 9 months",
			targetHorizon:   8,
			expectedHorizon: 9,
		},
		{
			name:            "closest to 12 months",
			targetHorizon:   11,
			expectedHorizon: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.getClosestHorizonPrediction(predictions, tt.targetHorizon)

			assert.NotNil(t, result)

			// Verify the result contains the expected horizon data
			expectedKey := fmt.Sprintf("horizon_%d", tt.expectedHorizon)
			expectedPrediction, exists := predictions[expectedKey]
			assert.True(t, exists)
			assert.Equal(t, expectedPrediction, result)
		})
	}
}

func TestLSTMModel_Concurrency(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 6,
	}

	// Test concurrent predictions
	numGoroutines := 10
	results := make(chan *models.RiskAssessment, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			ctx := context.Background()
			result, err := model.Predict(ctx, business)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}

	// Collect results
	var successfulResults []*models.RiskAssessment
	var errorResults []error

	for i := 0; i < numGoroutines; i++ {
		select {
		case result := <-results:
			successfulResults = append(successfulResults, result)
		case err := <-errors:
			errorResults = append(errorResults, err)
		}
	}

	// All predictions should succeed
	assert.Len(t, successfulResults, numGoroutines)
	assert.Len(t, errorResults, 0)

	// All results should be valid
	for _, result := range successfulResults {
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, result.RiskScore, 0.0)
		assert.LessOrEqual(t, result.RiskScore, 1.0)
		assert.NotEmpty(t, result.RiskLevel)
	}
}

func TestLSTMModel_Performance(t *testing.T) {
	logger := zap.NewNop()
	model := NewLSTMModel("test_model", "1.0.0", logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 6,
	}

	// Benchmark prediction performance
	numIterations := 100
	start := time.Now()

	for i := 0; i < numIterations; i++ {
		ctx := context.Background()
		result, err := model.Predict(ctx, business)
		require.NoError(t, err)
		require.NotNil(t, result)
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(numIterations)

	// Performance should be reasonable (less than 100ms per prediction)
	assert.Less(t, avgDuration, 100*time.Millisecond,
		"Average prediction time should be less than 100ms, got %v", avgDuration)

	t.Logf("Average prediction time: %v", avgDuration)
	t.Logf("Total time for %d predictions: %v", numIterations, duration)
}
