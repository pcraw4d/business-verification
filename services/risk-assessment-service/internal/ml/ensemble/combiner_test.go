package ensemble

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

func TestNewEnsembleCombiner(t *testing.T) {
	combiner := NewEnsembleCombiner()

	assert.NotNil(t, combiner)
	assert.NotNil(t, combiner.logger)
}

func TestEnsembleCombiner_SetLogger(t *testing.T) {
	combiner := NewEnsembleCombiner()
	logger := zap.NewNop()

	combiner.SetLogger(logger)

	assert.Equal(t, logger, combiner.logger)
}

func TestEnsembleCombiner_CombinePredictions(t *testing.T) {
	combiner := NewEnsembleCombiner()

	// Create test predictions
	xgbPrediction := &models.RiskAssessment{
		ID:              "xgb_assessment",
		BusinessID:      "test_business",
		BusinessName:    "Test Company",
		BusinessAddress: "123 Test St",
		Industry:        "technology",
		Country:         "US",
		RiskScore:       0.3,
		RiskLevel:       models.RiskLevelLow,
		RiskFactors: []models.RiskFactor{
			{
				Category:    models.RiskCategoryFinancial,
				Name:        "financial_health",
				Score:       0.8,
				Weight:      0.3,
				Description: "Good financial health",
				Source:      "xgboost",
				Confidence:  0.9,
			},
		},
		PredictionHorizon: 6,
		ConfidenceScore:   0.8,
		Status:            models.StatusCompleted,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Metadata: map[string]interface{}{
			"model_type": "xgboost",
		},
	}

	lstmPrediction := &models.RiskAssessment{
		ID:              "lstm_assessment",
		BusinessID:      "test_business",
		BusinessName:    "Test Company",
		BusinessAddress: "123 Test St",
		Industry:        "technology",
		Country:         "US",
		RiskScore:       0.4,
		RiskLevel:       models.RiskLevelMedium,
		RiskFactors: []models.RiskFactor{
			{
				Category:    models.RiskCategoryCompliance,
				Name:        "compliance_score",
				Score:       0.6,
				Weight:      0.2,
				Description: "Moderate compliance",
				Source:      "lstm",
				Confidence:  0.7,
			},
		},
		PredictionHorizon: 6,
		ConfidenceScore:   0.7,
		Status:            models.StatusCompleted,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Metadata: map[string]interface{}{
			"model_type": "lstm",
		},
	}

	tests := []struct {
		name           string
		xgbPrediction  *models.RiskAssessment
		lstmPrediction *models.RiskAssessment
		horizonMonths  int
		expectError    bool
		errorMsg       string
	}{
		{
			name:           "successful combination - short term",
			xgbPrediction:  xgbPrediction,
			lstmPrediction: lstmPrediction,
			horizonMonths:  3,
			expectError:    false,
		},
		{
			name:           "successful combination - medium term",
			xgbPrediction:  xgbPrediction,
			lstmPrediction: lstmPrediction,
			horizonMonths:  4,
			expectError:    false,
		},
		{
			name:           "successful combination - long term",
			xgbPrediction:  xgbPrediction,
			lstmPrediction: lstmPrediction,
			horizonMonths:  8,
			expectError:    false,
		},
		{
			name:           "only XGBoost prediction",
			xgbPrediction:  xgbPrediction,
			lstmPrediction: nil,
			horizonMonths:  3,
			expectError:    false,
		},
		{
			name:           "only LSTM prediction",
			xgbPrediction:  nil,
			lstmPrediction: lstmPrediction,
			horizonMonths:  8,
			expectError:    false,
		},
		{
			name:           "no predictions",
			xgbPrediction:  nil,
			lstmPrediction: nil,
			horizonMonths:  3,
			expectError:    true,
			errorMsg:       "no predictions to combine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := combiner.CombinePredictions(tt.xgbPrediction, tt.lstmPrediction, tt.horizonMonths)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "test_business", result.BusinessID)
				assert.Equal(t, "Test Company", result.BusinessName)
				assert.Equal(t, "123 Test St", result.BusinessAddress)
				assert.Equal(t, "technology", result.Industry)
				assert.Equal(t, "US", result.Country)
				assert.GreaterOrEqual(t, result.RiskScore, 0.0)
				assert.LessOrEqual(t, result.RiskScore, 1.0)
				assert.NotEmpty(t, result.RiskLevel)
				assert.GreaterOrEqual(t, result.ConfidenceScore, 0.0)
				assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
				assert.Equal(t, tt.horizonMonths, result.PredictionHorizon)
				assert.Equal(t, models.StatusCompleted, result.Status)
				assert.NotEmpty(t, result.ID)
				assert.NotZero(t, result.CreatedAt)
				assert.NotZero(t, result.UpdatedAt)
				assert.Contains(t, result.Metadata, "model_type")
				assert.Equal(t, "ensemble", result.Metadata["model_type"])
			}
		})
	}
}

func TestEnsembleCombiner_CombineFuturePredictions(t *testing.T) {
	combiner := NewEnsembleCombiner()

	// Create test future predictions
	xgbPrediction := &models.RiskPrediction{
		BusinessID:      "test_business",
		PredictionDate:  time.Now(),
		HorizonMonths:   6,
		PredictedScore:  0.3,
		PredictedLevel:  models.RiskLevelLow,
		ConfidenceScore: 0.8,
		RiskFactors: []models.RiskFactor{
			{
				Category:    models.RiskCategoryFinancial,
				Name:        "financial_health",
				Score:       0.8,
				Weight:      0.3,
				Description: "Good financial health",
				Source:      "xgboost",
				Confidence:  0.9,
			},
		},
		ScenarioAnalysis: []models.ScenarioAnalysis{
			{
				ScenarioName: "optimistic",
				Description:  "Best case scenario",
				RiskScore:    0.2,
				RiskLevel:    models.RiskLevelLow,
				Probability:  0.3,
				Impact:       "low",
			},
		},
		CreatedAt: time.Now(),
	}

	lstmPrediction := &models.RiskPrediction{
		BusinessID:      "test_business",
		PredictionDate:  time.Now(),
		HorizonMonths:   6,
		PredictedScore:  0.4,
		PredictedLevel:  models.RiskLevelMedium,
		ConfidenceScore: 0.7,
		RiskFactors: []models.RiskFactor{
			{
				Category:    models.RiskCategoryCompliance,
				Name:        "compliance_score",
				Score:       0.6,
				Weight:      0.2,
				Description: "Moderate compliance",
				Source:      "lstm",
				Confidence:  0.7,
			},
		},
		ScenarioAnalysis: []models.ScenarioAnalysis{
			{
				ScenarioName: "pessimistic",
				Description:  "Worst case scenario",
				RiskScore:    0.6,
				RiskLevel:    models.RiskLevelHigh,
				Probability:  0.2,
				Impact:       "high",
			},
		},
		CreatedAt: time.Now(),
	}

	tests := []struct {
		name           string
		xgbPrediction  *models.RiskPrediction
		lstmPrediction *models.RiskPrediction
		horizonMonths  int
		expectError    bool
		errorMsg       string
	}{
		{
			name:           "successful combination",
			xgbPrediction:  xgbPrediction,
			lstmPrediction: lstmPrediction,
			horizonMonths:  6,
			expectError:    false,
		},
		{
			name:           "only XGBoost prediction",
			xgbPrediction:  xgbPrediction,
			lstmPrediction: nil,
			horizonMonths:  6,
			expectError:    false,
		},
		{
			name:           "only LSTM prediction",
			xgbPrediction:  nil,
			lstmPrediction: lstmPrediction,
			horizonMonths:  6,
			expectError:    false,
		},
		{
			name:           "no predictions",
			xgbPrediction:  nil,
			lstmPrediction: nil,
			horizonMonths:  6,
			expectError:    true,
			errorMsg:       "no future predictions to combine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := combiner.CombineFuturePredictions(tt.xgbPrediction, tt.lstmPrediction, tt.horizonMonths)

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

func TestEnsembleCombiner_CalculateModelAgreement(t *testing.T) {
	combiner := NewEnsembleCombiner()

	// Create test predictions with different agreement levels
	highAgreementXgb := &models.RiskAssessment{
		RiskScore:       0.3,
		RiskLevel:       models.RiskLevelLow,
		ConfidenceScore: 0.8,
	}

	highAgreementLstm := &models.RiskAssessment{
		RiskScore:       0.32,
		RiskLevel:       models.RiskLevelLow,
		ConfidenceScore: 0.7,
	}

	lowAgreementXgb := &models.RiskAssessment{
		RiskScore:       0.2,
		RiskLevel:       models.RiskLevelLow,
		ConfidenceScore: 0.9,
	}

	lowAgreementLstm := &models.RiskAssessment{
		RiskScore:       0.8,
		RiskLevel:       models.RiskLevelHigh,
		ConfidenceScore: 0.6,
	}

	tests := []struct {
		name           string
		xgbPrediction  *models.RiskAssessment
		lstmPrediction *models.RiskAssessment
		expectedMin    float64
		expectedMax    float64
	}{
		{
			name:           "high agreement",
			xgbPrediction:  highAgreementXgb,
			lstmPrediction: highAgreementLstm,
			expectedMin:    0.8,
			expectedMax:    1.0,
		},
		{
			name:           "low agreement",
			xgbPrediction:  lowAgreementXgb,
			lstmPrediction: lowAgreementLstm,
			expectedMin:    0.0,
			expectedMax:    0.5,
		},
		{
			name:           "nil predictions",
			xgbPrediction:  nil,
			lstmPrediction: nil,
			expectedMin:    0.0,
			expectedMax:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agreement := combiner.CalculateModelAgreement(tt.xgbPrediction, tt.lstmPrediction)

			assert.GreaterOrEqual(t, agreement, tt.expectedMin)
			assert.LessOrEqual(t, agreement, tt.expectedMax)
		})
	}
}

func TestEnsembleCombiner_CalculateEnsembleConfidence(t *testing.T) {
	combiner := NewEnsembleCombiner()

	// Create test predictions
	xgbPrediction := &models.RiskAssessment{
		ConfidenceScore: 0.8,
	}

	lstmPrediction := &models.RiskAssessment{
		ConfidenceScore: 0.7,
	}

	tests := []struct {
		name           string
		xgbPrediction  *models.RiskAssessment
		lstmPrediction *models.RiskAssessment
		agreement      float64
		expectedMin    float64
		expectedMax    float64
	}{
		{
			name:           "high agreement",
			xgbPrediction:  xgbPrediction,
			lstmPrediction: lstmPrediction,
			agreement:      0.9,
			expectedMin:    0.6,
			expectedMax:    1.0,
		},
		{
			name:           "low agreement",
			xgbPrediction:  xgbPrediction,
			lstmPrediction: lstmPrediction,
			agreement:      0.3,
			expectedMin:    0.3,
			expectedMax:    0.6,
		},
		{
			name:           "only XGBoost",
			xgbPrediction:  xgbPrediction,
			lstmPrediction: nil,
			agreement:      0.0,
			expectedMin:    0.1,
			expectedMax:    1.0,
		},
		{
			name:           "only LSTM",
			xgbPrediction:  nil,
			lstmPrediction: lstmPrediction,
			agreement:      0.0,
			expectedMin:    0.1,
			expectedMax:    1.0,
		},
		{
			name:           "no predictions",
			xgbPrediction:  nil,
			lstmPrediction: nil,
			agreement:      0.0,
			expectedMin:    0.0,
			expectedMax:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := combiner.CalculateEnsembleConfidence(tt.xgbPrediction, tt.lstmPrediction, tt.agreement)

			assert.GreaterOrEqual(t, confidence, tt.expectedMin)
			assert.LessOrEqual(t, confidence, tt.expectedMax)
		})
	}
}

func TestEnsembleCombiner_GetEnsembleMetrics(t *testing.T) {
	combiner := NewEnsembleCombiner()

	// Create test predictions
	xgbPrediction := &models.RiskAssessment{
		RiskScore:       0.3,
		RiskLevel:       models.RiskLevelLow,
		ConfidenceScore: 0.8,
	}

	lstmPrediction := &models.RiskAssessment{
		RiskScore:       0.4,
		RiskLevel:       models.RiskLevelMedium,
		ConfidenceScore: 0.7,
	}

	tests := []struct {
		name           string
		xgbPrediction  *models.RiskAssessment
		lstmPrediction *models.RiskAssessment
		expectedKeys   []string
	}{
		{
			name:           "both predictions",
			xgbPrediction:  xgbPrediction,
			lstmPrediction: lstmPrediction,
			expectedKeys: []string{
				"model_agreement",
				"ensemble_confidence",
				"xgb_risk_score",
				"lstm_risk_score",
				"risk_score_difference",
				"xgb_confidence",
				"lstm_confidence",
				"risk_level_agreement",
			},
		},
		{
			name:           "only XGBoost",
			xgbPrediction:  xgbPrediction,
			lstmPrediction: nil,
			expectedKeys: []string{
				"single_model",
				"risk_score",
				"confidence",
			},
		},
		{
			name:           "only LSTM",
			xgbPrediction:  nil,
			lstmPrediction: lstmPrediction,
			expectedKeys: []string{
				"single_model",
				"risk_score",
				"confidence",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := combiner.GetEnsembleMetrics(tt.xgbPrediction, tt.lstmPrediction)

			assert.NotNil(t, metrics)
			for _, key := range tt.expectedKeys {
				assert.Contains(t, metrics, key)
			}
		})
	}
}

func TestEnsembleCombiner_WeightCalculation(t *testing.T) {
	combiner := NewEnsembleCombiner()

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
			xgbWeight, lstmWeight := combiner.getHorizonWeights(tt.horizonMonths)

			assert.Equal(t, tt.expectedXgbWeight, xgbWeight)
			assert.Equal(t, tt.expectedLstmWeight, lstmWeight)
			assert.InDelta(t, 1.0, xgbWeight+lstmWeight, 0.001) // Weights should sum to 1.0
		})
	}
}

func TestEnsembleCombiner_RiskLevelDetermination(t *testing.T) {
	combiner := NewEnsembleCombiner()

	tests := []struct {
		name          string
		riskScore     float64
		expectedLevel models.RiskLevel
	}{
		{
			name:          "low risk",
			riskScore:     0.1,
			expectedLevel: models.RiskLevelLow,
		},
		{
			name:          "low-medium boundary",
			riskScore:     0.25,
			expectedLevel: models.RiskLevelLow,
		},
		{
			name:          "medium risk",
			riskScore:     0.4,
			expectedLevel: models.RiskLevelMedium,
		},
		{
			name:          "medium-high boundary",
			riskScore:     0.5,
			expectedLevel: models.RiskLevelMedium,
		},
		{
			name:          "high risk",
			riskScore:     0.7,
			expectedLevel: models.RiskLevelHigh,
		},
		{
			name:          "high-critical boundary",
			riskScore:     0.75,
			expectedLevel: models.RiskLevelHigh,
		},
		{
			name:          "critical risk",
			riskScore:     0.9,
			expectedLevel: models.RiskLevelCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := combiner.determineRiskLevel(tt.riskScore)
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}
