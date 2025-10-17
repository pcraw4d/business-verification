//go:build integration

package ml

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/company/kyb-platform/services/risk-assessment-service/internal/config"
	"github.com/company/kyb-platform/services/risk-assessment-service/internal/ml"
	"github.com/company/kyb-platform/services/risk-assessment-service/internal/models"
)

// TestMLModelValidation tests ML model validation with cross-validation
func TestMLModelValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping ML model validation test")
	}

	logger := zap.NewNop()
	cfg := &config.MLConfig{
		ModelPath:    "/tmp/test-models",
		ModelType:    "xgboost",
		BatchSize:    100,
		MaxWorkers:   4,
		CacheEnabled: true,
		CacheTTL:     300 * time.Second,
	}

	service, err := ml.NewService(cfg, logger)
	require.NoError(t, err)
	defer service.Close()

	tests := []struct {
		name              string
		modelType         string
		expectedAccuracy  float64
		expectedPrecision float64
		expectedRecall    float64
		expectedF1Score   float64
	}{
		{
			name:              "XGBoost model validation",
			modelType:         "xgboost",
			expectedAccuracy:  0.85,
			expectedPrecision: 0.80,
			expectedRecall:    0.82,
			expectedF1Score:   0.81,
		},
		{
			name:              "LSTM model validation",
			modelType:         "lstm",
			expectedAccuracy:  0.88,
			expectedPrecision: 0.85,
			expectedRecall:    0.87,
			expectedF1Score:   0.86,
		},
		{
			name:              "Ensemble model validation",
			modelType:         "ensemble",
			expectedAccuracy:  0.90,
			expectedPrecision: 0.88,
			expectedRecall:    0.89,
			expectedF1Score:   0.88,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate test data
			trainingData := generateTrainingData(1000)
			validationData := generateValidationData(200)

			// Train model
			trainingResult, err := service.TrainModel(context.Background(), &ml.TrainingData{
				Samples:   trainingData,
				ModelType: tt.modelType,
			})
			require.NoError(t, err)
			require.NotNil(t, trainingResult)

			// Validate model
			validationResult, err := service.ValidateModel(context.Background(), &ml.ValidationData{
				Samples:   validationData,
				ModelType: tt.modelType,
			})
			require.NoError(t, err)
			require.NotNil(t, validationResult)

			// Assert performance metrics
			assert.GreaterOrEqual(t, validationResult.Accuracy, tt.expectedAccuracy,
				"Model accuracy should meet minimum threshold")
			assert.GreaterOrEqual(t, validationResult.Precision, tt.expectedPrecision,
				"Model precision should meet minimum threshold")
			assert.GreaterOrEqual(t, validationResult.Recall, tt.expectedRecall,
				"Model recall should meet minimum threshold")
			assert.GreaterOrEqual(t, validationResult.F1Score, tt.expectedF1Score,
				"Model F1 score should meet minimum threshold")

			// Assert additional metrics
			assert.GreaterOrEqual(t, validationResult.AUC, 0.80,
				"Model AUC should be at least 0.80")
			assert.LessOrEqual(t, validationResult.MAE, 0.15,
				"Model MAE should be at most 0.15")
			assert.LessOrEqual(t, validationResult.RMSE, 0.20,
				"Model RMSE should be at most 0.20")
		})
	}
}

// TestCrossValidation tests cross-validation functionality
func TestCrossValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cross-validation test")
	}

	logger := zap.NewNop()
	cfg := &config.MLConfig{
		ModelPath:    "/tmp/test-models",
		ModelType:    "xgboost",
		BatchSize:    100,
		MaxWorkers:   4,
		CacheEnabled: true,
		CacheTTL:     300 * time.Second,
	}

	service, err := ml.NewService(cfg, logger)
	require.NoError(t, err)
	defer service.Close()

	// Generate test data
	trainingData := generateTrainingData(2000)

	tests := []struct {
		name             string
		folds            int
		expectedAccuracy float64
		expectedStdDev   float64
	}{
		{
			name:             "5-fold cross-validation",
			folds:            5,
			expectedAccuracy: 0.85,
			expectedStdDev:   0.05,
		},
		{
			name:             "10-fold cross-validation",
			folds:            10,
			expectedAccuracy: 0.85,
			expectedStdDev:   0.03,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Perform cross-validation
			cvResult, err := service.CrossValidate(context.Background(), &ml.CrossValidationData{
				Samples:   trainingData,
				Folds:     tt.folds,
				ModelType: "xgboost",
			})
			require.NoError(t, err)
			require.NotNil(t, cvResult)

			// Assert cross-validation results
			assert.GreaterOrEqual(t, cvResult.MeanAccuracy, tt.expectedAccuracy,
				"Cross-validation mean accuracy should meet threshold")
			assert.LessOrEqual(t, cvResult.StdDevAccuracy, tt.expectedStdDev,
				"Cross-validation standard deviation should be within acceptable range")
			assert.Equal(t, tt.folds, len(cvResult.FoldResults),
				"Number of fold results should match number of folds")

			// Assert individual fold results
			for i, foldResult := range cvResult.FoldResults {
				assert.GreaterOrEqual(t, foldResult.Accuracy, 0.70,
					"Fold %d accuracy should be at least 0.70", i+1)
				assert.GreaterOrEqual(t, foldResult.Precision, 0.70,
					"Fold %d precision should be at least 0.70", i+1)
				assert.GreaterOrEqual(t, foldResult.Recall, 0.70,
					"Fold %d recall should be at least 0.70", i+1)
			}
		})
	}
}

// TestModelPerformanceComparison tests performance comparison between different models
func TestModelPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping model performance comparison test")
	}

	logger := zap.NewNop()
	cfg := &config.MLConfig{
		ModelPath:    "/tmp/test-models",
		ModelType:    "xgboost",
		BatchSize:    100,
		MaxWorkers:   4,
		CacheEnabled: true,
		CacheTTL:     300 * time.Second,
	}

	service, err := ml.NewService(cfg, logger)
	require.NoError(t, err)
	defer service.Close()

	// Generate test data
	trainingData := generateTrainingData(1500)
	validationData := generateValidationData(300)

	modelTypes := []string{"xgboost", "lstm", "ensemble"}
	results := make(map[string]*ml.ValidationResult)

	// Train and validate each model
	for _, modelType := range modelTypes {
		// Train model
		_, err := service.TrainModel(context.Background(), &ml.TrainingData{
			Samples:   trainingData,
			ModelType: modelType,
		})
		require.NoError(t, err)

		// Validate model
		validationResult, err := service.ValidateModel(context.Background(), &ml.ValidationData{
			Samples:   validationData,
			ModelType: modelType,
		})
		require.NoError(t, err)
		results[modelType] = validationResult
	}

	// Compare model performance
	t.Run("accuracy_comparison", func(t *testing.T) {
		// Ensemble should have highest accuracy
		assert.Greater(t, results["ensemble"].Accuracy, results["xgboost"].Accuracy,
			"Ensemble model should have higher accuracy than XGBoost")
		assert.Greater(t, results["ensemble"].Accuracy, results["lstm"].Accuracy,
			"Ensemble model should have higher accuracy than LSTM")
	})

	t.Run("precision_comparison", func(t *testing.T) {
		// All models should have reasonable precision
		for modelType, result := range results {
			assert.GreaterOrEqual(t, result.Precision, 0.75,
				"%s model precision should be at least 0.75", modelType)
		}
	})

	t.Run("recall_comparison", func(t *testing.T) {
		// All models should have reasonable recall
		for modelType, result := range results {
			assert.GreaterOrEqual(t, result.Recall, 0.75,
				"%s model recall should be at least 0.75", modelType)
		}
	})

	t.Run("f1_score_comparison", func(t *testing.T) {
		// All models should have reasonable F1 score
		for modelType, result := range results {
			assert.GreaterOrEqual(t, result.F1Score, 0.75,
				"%s model F1 score should be at least 0.75", modelType)
		}
	})
}

// TestModelRobustness tests model robustness with different data distributions
func TestModelRobustness(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping model robustness test")
	}

	logger := zap.NewNop()
	cfg := &config.MLConfig{
		ModelPath:    "/tmp/test-models",
		ModelType:    "xgboost",
		BatchSize:    100,
		MaxWorkers:   4,
		CacheEnabled: true,
		CacheTTL:     300 * time.Second,
	}

	service, err := ml.NewService(cfg, logger)
	require.NoError(t, err)
	defer service.Close()

	// Generate different data distributions
	normalData := generateTrainingData(1000)
	skewedData := generateSkewedTrainingData(1000)
	imbalancedData := generateImbalancedTrainingData(1000)

	tests := []struct {
		name             string
		trainingData     []*ml.TrainingSample
		expectedAccuracy float64
	}{
		{
			name:             "normal distribution",
			trainingData:     normalData,
			expectedAccuracy: 0.85,
		},
		{
			name:             "skewed distribution",
			trainingData:     skewedData,
			expectedAccuracy: 0.80,
		},
		{
			name:             "imbalanced distribution",
			trainingData:     imbalancedData,
			expectedAccuracy: 0.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Train model
			trainingResult, err := service.TrainModel(context.Background(), &ml.TrainingData{
				Samples:   tt.trainingData,
				ModelType: "xgboost",
			})
			require.NoError(t, err)
			require.NotNil(t, trainingResult)

			// Validate model
			validationData := generateValidationData(200)
			validationResult, err := service.ValidateModel(context.Background(), &ml.ValidationData{
				Samples:   validationData,
				ModelType: "xgboost",
			})
			require.NoError(t, err)
			require.NotNil(t, validationResult)

			// Assert performance meets expectations
			assert.GreaterOrEqual(t, validationResult.Accuracy, tt.expectedAccuracy,
				"Model accuracy should meet threshold for %s", tt.name)
		})
	}
}

// TestModelInferencePerformance tests model inference performance
func TestModelInferencePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping model inference performance test")
	}

	logger := zap.NewNop()
	cfg := &config.MLConfig{
		ModelPath:    "/tmp/test-models",
		ModelType:    "xgboost",
		BatchSize:    100,
		MaxWorkers:   4,
		CacheEnabled: true,
		CacheTTL:     300 * time.Second,
	}

	service, err := ml.NewService(cfg, logger)
	require.NoError(t, err)
	defer service.Close()

	// Train model
	trainingData := generateTrainingData(1000)
	_, err = service.TrainModel(context.Background(), &ml.TrainingData{
		Samples:   trainingData,
		ModelType: "xgboost",
	})
	require.NoError(t, err)

	// Test single inference performance
	t.Run("single_inference", func(t *testing.T) {
		request := &models.RiskAssessmentRequest{
			BusinessName:      "Test Company Inc",
			BusinessAddress:   "123 Test Street, Test City, TC 12345",
			Industry:          "technology",
			Country:           "US",
			PredictionHorizon: 3,
			ModelType:         "xgboost",
		}

		start := time.Now()
		result, err := service.AssessRisk(context.Background(), request)
		duration := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, result)

		// Assert inference time is within acceptable limits
		assert.Less(t, duration, 200*time.Millisecond,
			"Single inference should complete within 200ms")
	})

	// Test batch inference performance
	t.Run("batch_inference", func(t *testing.T) {
		var requests []*models.RiskAssessmentRequest
		for i := 0; i < 100; i++ {
			requests = append(requests, &models.RiskAssessmentRequest{
				BusinessName:      fmt.Sprintf("Test Company %d", i),
				BusinessAddress:   fmt.Sprintf("%d Test Street, Test City, TC %05d", i, i),
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
				ModelType:         "xgboost",
			})
		}

		start := time.Now()
		result, err := service.AssessRiskBatch(context.Background(), requests)
		duration := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, result)

		// Assert batch inference time is within acceptable limits
		assert.Less(t, duration, 2*time.Second,
			"Batch inference of 100 requests should complete within 2 seconds")
	})
}

// Helper functions

// generateTrainingData generates training data for testing
func generateTrainingData(count int) []*ml.TrainingSample {
	var samples []*ml.TrainingSample

	for i := 0; i < count; i++ {
		// Generate random features
		features := map[string]float64{
			"business_age":     float64(i % 20),
			"industry_risk":    float64(i%10) / 10.0,
			"country_risk":     float64(i%5) / 5.0,
			"financial_health": float64(i%10) / 10.0,
			"operational_risk": float64(i%10) / 10.0,
			"compliance_score": float64(i%10) / 10.0,
		}

		// Generate target (risk score)
		target := calculateRiskScore(features)

		samples = append(samples, &ml.TrainingSample{
			Features: features,
			Target:   target,
		})
	}

	return samples
}

// generateValidationData generates validation data for testing
func generateValidationData(count int) []*ml.ValidationSample {
	var samples []*ml.ValidationSample

	for i := 0; i < count; i++ {
		// Generate random features
		features := map[string]float64{
			"business_age":     float64(i % 20),
			"industry_risk":    float64(i%10) / 10.0,
			"country_risk":     float64(i%5) / 5.0,
			"financial_health": float64(i%10) / 10.0,
			"operational_risk": float64(i%10) / 10.0,
			"compliance_score": float64(i%10) / 10.0,
		}

		// Generate target (risk score)
		target := calculateRiskScore(features)

		samples = append(samples, &ml.ValidationSample{
			Features: features,
			Target:   target,
		})
	}

	return samples
}

// generateSkewedTrainingData generates skewed training data
func generateSkewedTrainingData(count int) []*ml.TrainingSample {
	var samples []*ml.TrainingSample

	for i := 0; i < count; i++ {
		// Generate skewed features (higher risk bias)
		features := map[string]float64{
			"business_age":     float64(i % 20),
			"industry_risk":    math.Min(1.0, float64(i%10)/5.0), // Skewed towards higher risk
			"country_risk":     math.Min(1.0, float64(i%5)/2.0),  // Skewed towards higher risk
			"financial_health": float64(i%10) / 10.0,
			"operational_risk": math.Min(1.0, float64(i%10)/5.0), // Skewed towards higher risk
			"compliance_score": float64(i%10) / 10.0,
		}

		// Generate target (risk score)
		target := calculateRiskScore(features)

		samples = append(samples, &ml.TrainingSample{
			Features: features,
			Target:   target,
		})
	}

	return samples
}

// generateImbalancedTrainingData generates imbalanced training data
func generateImbalancedTrainingData(count int) []*ml.TrainingSample {
	var samples []*ml.TrainingSample

	for i := 0; i < count; i++ {
		// Generate features with class imbalance (80% low risk, 20% high risk)
		var features map[string]float64
		if i < count*8/10 {
			// Low risk samples
			features = map[string]float64{
				"business_age":     float64(i % 20),
				"industry_risk":    float64(i%3) / 10.0, // Low risk
				"country_risk":     float64(i%2) / 10.0, // Low risk
				"financial_health": float64(i%10) / 10.0,
				"operational_risk": float64(i%3) / 10.0, // Low risk
				"compliance_score": float64(i%10) / 10.0,
			}
		} else {
			// High risk samples
			features = map[string]float64{
				"business_age":     float64(i % 20),
				"industry_risk":    float64(i%10) / 10.0,
				"country_risk":     float64(i%5) / 5.0,
				"financial_health": float64(i%10) / 10.0,
				"operational_risk": float64(i%10) / 10.0,
				"compliance_score": float64(i%10) / 10.0,
			}
		}

		// Generate target (risk score)
		target := calculateRiskScore(features)

		samples = append(samples, &ml.TrainingSample{
			Features: features,
			Target:   target,
		})
	}

	return samples
}

// calculateRiskScore calculates risk score based on features
func calculateRiskScore(features map[string]float64) float64 {
	// Simple risk score calculation for testing
	riskScore := 0.0

	// Weighted combination of features
	riskScore += features["industry_risk"] * 0.3
	riskScore += features["country_risk"] * 0.2
	riskScore += features["operational_risk"] * 0.25
	riskScore += features["compliance_score"] * 0.15
	riskScore += (1.0 - features["financial_health"]) * 0.1

	// Add some noise
	noise := (float64(len(features)) % 100) / 1000.0
	riskScore += noise

	// Ensure score is between 0 and 1
	if riskScore > 1.0 {
		riskScore = 1.0
	}
	if riskScore < 0.0 {
		riskScore = 0.0
	}

	return riskScore
}
