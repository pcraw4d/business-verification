package ensemble

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	mlmodels "kyb-platform/services/risk-assessment-service/internal/ml/models"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// MockRiskModel is a mock implementation of the RiskModel interface for testing
type MockRiskModel struct {
	name        string
	version     string
	riskScore   float64
	confidence  float64
	shouldError bool
	errorMsg    string
}

func (m *MockRiskModel) Predict(ctx context.Context, business *models.RiskAssessmentRequest) (*models.RiskAssessment, error) {
	if m.shouldError {
		return nil, assert.AnError
	}

	return &models.RiskAssessment{
		ID:              "mock_assessment_" + m.name,
		BusinessID:      "mock_business",
		BusinessName:    business.BusinessName,
		BusinessAddress: business.BusinessAddress,
		Industry:        business.Industry,
		Country:         business.Country,
		RiskScore:       m.riskScore,
		RiskLevel:       models.RiskLevelMedium,
		RiskFactors: []models.RiskFactor{
			{
				Category:    models.RiskCategoryFinancial,
				Name:        "mock_factor_" + m.name,
				Score:       m.riskScore,
				Weight:      0.5,
				Description: "Mock risk factor from " + m.name,
				Source:      m.name,
				Confidence:  m.confidence,
			},
		},
		PredictionHorizon: business.PredictionHorizon,
		ConfidenceScore:   m.confidence,
		Status:            models.StatusCompleted,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Metadata: map[string]interface{}{
			"model_type": m.name,
			"mock":       true,
		},
	}, nil
}

func (m *MockRiskModel) PredictFuture(ctx context.Context, business *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	if m.shouldError {
		return nil, assert.AnError
	}

	return &models.RiskPrediction{
		BusinessID:      "mock_business",
		PredictionDate:  time.Now(),
		HorizonMonths:   horizonMonths,
		PredictedScore:  m.riskScore,
		PredictedLevel:  models.RiskLevelMedium,
		ConfidenceScore: m.confidence,
		RiskFactors: []models.RiskFactor{
			{
				Category:    models.RiskCategoryFinancial,
				Name:        "mock_future_factor_" + m.name,
				Score:       m.riskScore,
				Weight:      0.5,
				Description: "Mock future risk factor from " + m.name,
				Source:      m.name,
				Confidence:  m.confidence,
			},
		},
		ScenarioAnalysis: []models.ScenarioAnalysis{
			{
				ScenarioName: "mock_scenario_" + m.name,
				Description:  "Mock scenario from " + m.name,
				RiskScore:    m.riskScore,
				RiskLevel:    models.RiskLevelMedium,
				Probability:  0.5,
				Impact:       "medium",
			},
		},
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockRiskModel) GetModelInfo() *mlmodels.ModelInfo {
	return &mlmodels.ModelInfo{
		Name:         m.name,
		Version:      m.version,
		Type:         m.name,
		TrainingDate: time.Now(),
		Accuracy:     0.85,
		Precision:    0.82,
		Recall:       0.88,
		F1Score:      0.85,
		Features:     []string{"mock_feature"},
		Hyperparameters: map[string]interface{}{
			"learning_rate": 0.01,
		},
		Metadata: map[string]interface{}{
			"mock": true,
		},
	}
}

func (m *MockRiskModel) LoadModel(ctx context.Context, modelPath string) error {
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockRiskModel) SaveModel(ctx context.Context, modelPath string) error {
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockRiskModel) ValidateModel(ctx context.Context, testData []*models.RiskAssessment) (*mlmodels.ValidationResult, error) {
	if m.shouldError {
		return nil, assert.AnError
	}
	return &mlmodels.ValidationResult{
		Accuracy:       0.85,
		Precision:      0.82,
		Recall:         0.88,
		F1Score:        0.85,
		ValidationDate: time.Now(),
	}, nil
}

func TestNewEnsembleRouter(t *testing.T) {
	logger := zap.NewNop()
	xgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", riskScore: 0.3, confidence: 0.8}
	lstmModel := &MockRiskModel{name: "lstm", version: "1.0.0", riskScore: 0.4, confidence: 0.7}

	router := NewEnsembleRouter(xgbModel, lstmModel, logger)

	assert.NotNil(t, router)
	assert.Equal(t, xgbModel, router.xgboostModel)
	assert.Equal(t, lstmModel, router.lstmModel)
	assert.NotNil(t, router.combiner)
	assert.Equal(t, logger, router.logger)
}

func TestEnsembleRouter_Route(t *testing.T) {
	logger := zap.NewNop()
	xgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0"}
	lstmModel := &MockRiskModel{name: "lstm", version: "1.0.0"}
	router := NewEnsembleRouter(xgbModel, lstmModel, logger)

	tests := []struct {
		name          string
		horizonMonths int
		expectedModel string
	}{
		{
			name:          "short term - 1 month",
			horizonMonths: 1,
			expectedModel: "xgboost",
		},
		{
			name:          "short term - 3 months",
			horizonMonths: 3,
			expectedModel: "xgboost",
		},
		{
			name:          "medium term - 4 months",
			horizonMonths: 4,
			expectedModel: "ensemble",
		},
		{
			name:          "medium term - 5 months",
			horizonMonths: 5,
			expectedModel: "ensemble",
		},
		{
			name:          "long term - 6 months",
			horizonMonths: 6,
			expectedModel: "lstm",
		},
		{
			name:          "long term - 12 months",
			horizonMonths: 12,
			expectedModel: "lstm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.Route(tt.horizonMonths)
			assert.Equal(t, tt.expectedModel, result)
		})
	}
}

func TestEnsembleRouter_PredictWithEnsemble(t *testing.T) {
	logger := zap.NewNop()
	xgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", riskScore: 0.3, confidence: 0.8}
	lstmModel := &MockRiskModel{name: "lstm", version: "1.0.0", riskScore: 0.4, confidence: 0.7}
	router := NewEnsembleRouter(xgbModel, lstmModel, logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 4, // Should trigger ensemble
	}

	t.Run("successful ensemble prediction", func(t *testing.T) {
		ctx := context.Background()
		result, err := router.PredictWithEnsemble(ctx, business)

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
		assert.Contains(t, result.Metadata, "model_type")
		assert.Equal(t, "ensemble", result.Metadata["model_type"])
	})

	t.Run("XGBoost model failure", func(t *testing.T) {
		// Create router with failing XGBoost model
		failingXgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", shouldError: true}
		routerWithFailingXgb := NewEnsembleRouter(failingXgbModel, lstmModel, logger)

		ctx := context.Background()
		result, err := routerWithFailingXgb.PredictWithEnsemble(ctx, business)

		assert.NoError(t, err) // Should fallback to LSTM
		assert.NotNil(t, result)
		assert.Equal(t, "lstm", result.Metadata["model_type"])
	})

	t.Run("LSTM model failure", func(t *testing.T) {
		// Create router with failing LSTM model
		failingLstmModel := &MockRiskModel{name: "lstm", version: "1.0.0", shouldError: true}
		routerWithFailingLstm := NewEnsembleRouter(xgbModel, failingLstmModel, logger)

		ctx := context.Background()
		result, err := routerWithFailingLstm.PredictWithEnsemble(ctx, business)

		assert.NoError(t, err) // Should fallback to XGBoost
		assert.NotNil(t, result)
		assert.Equal(t, "xgboost", result.Metadata["model_type"])
	})

	t.Run("both models failure", func(t *testing.T) {
		// Create router with both models failing
		failingXgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", shouldError: true}
		failingLstmModel := &MockRiskModel{name: "lstm", version: "1.0.0", shouldError: true}
		routerWithFailingModels := NewEnsembleRouter(failingXgbModel, failingLstmModel, logger)

		ctx := context.Background()
		result, err := routerWithFailingModels.PredictWithEnsemble(ctx, business)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "both model predictions failed")
	})
}

func TestEnsembleRouter_PredictFutureWithEnsemble(t *testing.T) {
	logger := zap.NewNop()
	xgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", riskScore: 0.3, confidence: 0.8}
	lstmModel := &MockRiskModel{name: "lstm", version: "1.0.0", riskScore: 0.4, confidence: 0.7}
	router := NewEnsembleRouter(xgbModel, lstmModel, logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 4,
	}

	t.Run("successful ensemble future prediction", func(t *testing.T) {
		ctx := context.Background()
		horizonMonths := 6
		result, err := router.PredictFutureWithEnsemble(ctx, business, horizonMonths)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, horizonMonths, result.HorizonMonths)
		assert.GreaterOrEqual(t, result.PredictedScore, 0.0)
		assert.LessOrEqual(t, result.PredictedScore, 1.0)
		assert.NotEmpty(t, result.PredictedLevel)
		assert.GreaterOrEqual(t, result.ConfidenceScore, 0.0)
		assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
		assert.NotZero(t, result.PredictionDate)
		assert.NotZero(t, result.CreatedAt)
		assert.NotEmpty(t, result.RiskFactors)
		assert.NotEmpty(t, result.ScenarioAnalysis)
	})

	t.Run("XGBoost model failure", func(t *testing.T) {
		failingXgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", shouldError: true}
		routerWithFailingXgb := NewEnsembleRouter(failingXgbModel, lstmModel, logger)

		ctx := context.Background()
		horizonMonths := 6
		result, err := routerWithFailingXgb.PredictFutureWithEnsemble(ctx, business, horizonMonths)

		assert.NoError(t, err) // Should fallback to LSTM
		assert.NotNil(t, result)
		assert.Equal(t, horizonMonths, result.HorizonMonths)
	})

	t.Run("both models failure", func(t *testing.T) {
		failingXgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", shouldError: true}
		failingLstmModel := &MockRiskModel{name: "lstm", version: "1.0.0", shouldError: true}
		routerWithFailingModels := NewEnsembleRouter(failingXgbModel, failingLstmModel, logger)

		ctx := context.Background()
		horizonMonths := 6
		result, err := routerWithFailingModels.PredictFutureWithEnsemble(ctx, business, horizonMonths)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "both model future predictions failed")
	})
}

func TestEnsembleRouter_GetModelWeights(t *testing.T) {
	logger := zap.NewNop()
	xgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0"}
	lstmModel := &MockRiskModel{name: "lstm", version: "1.0.0"}
	router := NewEnsembleRouter(xgbModel, lstmModel, logger)

	tests := []struct {
		name               string
		horizonMonths      int
		expectedXgbWeight  float64
		expectedLstmWeight float64
	}{
		{
			name:               "short term - 1 month",
			horizonMonths:      1,
			expectedXgbWeight:  0.8,
			expectedLstmWeight: 0.2,
		},
		{
			name:               "short term - 3 months",
			horizonMonths:      3,
			expectedXgbWeight:  0.8,
			expectedLstmWeight: 0.2,
		},
		{
			name:               "medium term - 4 months",
			horizonMonths:      4,
			expectedXgbWeight:  0.5,
			expectedLstmWeight: 0.5,
		},
		{
			name:               "medium term - 5 months",
			horizonMonths:      5,
			expectedXgbWeight:  0.5,
			expectedLstmWeight: 0.5,
		},
		{
			name:               "long term - 6 months",
			horizonMonths:      6,
			expectedXgbWeight:  0.2,
			expectedLstmWeight: 0.8,
		},
		{
			name:               "long term - 12 months",
			horizonMonths:      12,
			expectedXgbWeight:  0.2,
			expectedLstmWeight: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xgbWeight, lstmWeight := router.GetModelWeights(tt.horizonMonths)
			assert.Equal(t, tt.expectedXgbWeight, xgbWeight)
			assert.Equal(t, tt.expectedLstmWeight, lstmWeight)
		})
	}
}

func TestEnsembleRouter_GetModelInfo(t *testing.T) {
	logger := zap.NewNop()
	xgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0"}
	lstmModel := &MockRiskModel{name: "lstm", version: "1.0.0"}
	router := NewEnsembleRouter(xgbModel, lstmModel, logger)

	info := router.GetModelInfo()

	assert.NotNil(t, info)
	assert.Equal(t, "ensemble_router", info["type"])
	assert.Contains(t, info["available_models"], "xgboost")
	assert.Contains(t, info["available_models"], "lstm")
	assert.Contains(t, info, "routing_logic")
	assert.Contains(t, info, "ensemble_weights")
	assert.Contains(t, info, "xgboost_model")
	assert.Contains(t, info, "lstm_model")
}

func TestEnsembleRouter_Health(t *testing.T) {
	logger := zap.NewNop()

	t.Run("healthy models", func(t *testing.T) {
		xgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", riskScore: 0.3, confidence: 0.8}
		lstmModel := &MockRiskModel{name: "lstm", version: "1.0.0", riskScore: 0.4, confidence: 0.7}
		router := NewEnsembleRouter(xgbModel, lstmModel, logger)

		ctx := context.Background()
		err := router.Health(ctx)

		assert.NoError(t, err)
	})

	t.Run("failing XGBoost model", func(t *testing.T) {
		failingXgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", shouldError: true}
		lstmModel := &MockRiskModel{name: "lstm", version: "1.0.0", riskScore: 0.4, confidence: 0.7}
		router := NewEnsembleRouter(failingXgbModel, lstmModel, logger)

		ctx := context.Background()
		err := router.Health(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "XGBoost model health check failed")
	})

	t.Run("failing LSTM model", func(t *testing.T) {
		xgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", riskScore: 0.3, confidence: 0.8}
		failingLstmModel := &MockRiskModel{name: "lstm", version: "1.0.0", shouldError: true}
		router := NewEnsembleRouter(xgbModel, failingLstmModel, logger)

		ctx := context.Background()
		err := router.Health(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LSTM model health check failed")
	})
}

func TestEnsembleRouter_Concurrency(t *testing.T) {
	logger := zap.NewNop()
	xgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", riskScore: 0.3, confidence: 0.8}
	lstmModel := &MockRiskModel{name: "lstm", version: "1.0.0", riskScore: 0.4, confidence: 0.7}
	router := NewEnsembleRouter(xgbModel, lstmModel, logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 4,
	}

	// Test concurrent ensemble predictions
	numGoroutines := 10
	results := make(chan *models.RiskAssessment, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			ctx := context.Background()
			result, err := router.PredictWithEnsemble(ctx, business)
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
		assert.Equal(t, "ensemble", result.Metadata["model_type"])
	}
}

func TestEnsembleRouter_Performance(t *testing.T) {
	logger := zap.NewNop()
	xgbModel := &MockRiskModel{name: "xgboost", version: "1.0.0", riskScore: 0.3, confidence: 0.8}
	lstmModel := &MockRiskModel{name: "lstm", version: "1.0.0", riskScore: 0.4, confidence: 0.7}
	router := NewEnsembleRouter(xgbModel, lstmModel, logger)

	business := &models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 4,
	}

	// Benchmark ensemble prediction performance
	numIterations := 50
	start := time.Now()

	for i := 0; i < numIterations; i++ {
		ctx := context.Background()
		result, err := router.PredictWithEnsemble(ctx, business)
		require.NoError(t, err)
		require.NotNil(t, result)
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(numIterations)

	// Performance should be reasonable (less than 200ms per prediction)
	assert.Less(t, avgDuration, 200*time.Millisecond,
		"Average ensemble prediction time should be less than 200ms, got %v", avgDuration)

	t.Logf("Average ensemble prediction time: %v", avgDuration)
	t.Logf("Total time for %d predictions: %v", numIterations, duration)
}
