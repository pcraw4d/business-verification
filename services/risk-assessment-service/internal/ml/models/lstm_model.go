package models

import (
	"context"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// LSTMModel implements the RiskModel interface using the ONNX LSTM model
type LSTMModel struct {
	onnxModel *LSTMONNXModel
}

// NewLSTMModel creates a new LSTM model instance
func NewLSTMModel(name, version string, logger *zap.Logger) *LSTMModel {
	return &LSTMModel{
		onnxModel: NewLSTMONNXModel(name, version, logger),
	}
}

// LoadModel loads the LSTM model
func (lstm *LSTMModel) LoadModel(ctx context.Context, modelPath string) error {
	return lstm.onnxModel.LoadModel(ctx, modelPath)
}

// Predict performs risk assessment using LSTM model
func (lstm *LSTMModel) Predict(ctx context.Context, business *models.RiskAssessmentRequest) (*models.RiskAssessment, error) {
	return lstm.onnxModel.Predict(ctx, business)
}

// PredictFuture performs future risk prediction using LSTM model
func (lstm *LSTMModel) PredictFuture(ctx context.Context, business *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	// Use enhanced multi-step prediction for 6-12 month forecasts
	if horizonMonths >= 6 && horizonMonths <= 12 {
		return lstm.onnxModel.PredictMultiStep(ctx, business, horizonMonths)
	}

	// Fall back to standard prediction for shorter horizons
	return lstm.onnxModel.PredictFuture(ctx, business, horizonMonths)
}

// SaveModel saves the LSTM model
func (lstm *LSTMModel) SaveModel(ctx context.Context, modelPath string) error {
	return lstm.onnxModel.SaveModel(ctx, modelPath)
}

// ValidateModel validates the LSTM model
func (lstm *LSTMModel) ValidateModel(ctx context.Context, testData []*models.RiskAssessment) (*ValidationResult, error) {
	return lstm.onnxModel.ValidateModel(ctx, testData)
}

// GetModelInfo returns information about the LSTM model
func (lstm *LSTMModel) GetModelInfo() *ModelInfo {
	return lstm.onnxModel.GetModelInfo()
}
