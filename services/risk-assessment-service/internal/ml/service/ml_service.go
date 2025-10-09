package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	mlmodels "kyb-platform/services/risk-assessment-service/internal/ml/models"
	"kyb-platform/services/risk-assessment-service/internal/ml/training"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// MLService provides machine learning capabilities for risk assessment
type MLService struct {
	modelManager *mlmodels.ModelManager
	trainer      *training.ModelTrainer
	logger       *zap.Logger
}

// NewMLService creates a new ML service
func NewMLService(logger *zap.Logger) *MLService {
	return &MLService{
		modelManager: mlmodels.NewModelManager(),
		trainer:      training.NewModelTrainer(),
		logger:       logger,
	}
}

// InitializeModels initializes and loads the ML models
func (mls *MLService) InitializeModels(ctx context.Context) error {
	mls.logger.Info("Initializing ML models")

	// Create and register XGBoost model
	xgbModel := mlmodels.NewXGBoostModel("risk_prediction_xgb", "1.0.0")

	// Load the model (in a real implementation, this would load from storage)
	if err := xgbModel.LoadModel(ctx, "./models/xgb_model.json"); err != nil {
		mls.logger.Warn("Failed to load XGBoost model, using default", zap.Error(err))
	}

	// Register the model
	mls.modelManager.RegisterModel("xgboost", xgbModel)

	mls.logger.Info("ML models initialized successfully",
		zap.Strings("available_models", mls.modelManager.ListModels()))

	return nil
}

// PredictRisk performs risk prediction using the specified model
func (mls *MLService) PredictRisk(ctx context.Context, modelName string, business *models.RiskAssessmentRequest) (*models.RiskAssessment, error) {
	model, err := mls.modelManager.GetModel(modelName)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	startTime := time.Now()

	prediction, err := model.Predict(ctx, business)
	if err != nil {
		mls.logger.Error("Risk prediction failed",
			zap.String("model", modelName),
			zap.Error(err))
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	duration := time.Since(startTime)

	mls.logger.Info("Risk prediction completed",
		zap.String("model", modelName),
		zap.String("assessment_id", prediction.ID),
		zap.Float64("risk_score", prediction.RiskScore),
		zap.String("risk_level", string(prediction.RiskLevel)),
		zap.Duration("duration", duration))

	return prediction, nil
}

// PredictFutureRisk performs future risk prediction
func (mls *MLService) PredictFutureRisk(ctx context.Context, modelName string, business *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	model, err := mls.modelManager.GetModel(modelName)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	startTime := time.Now()

	prediction, err := model.PredictFuture(ctx, business, horizonMonths)
	if err != nil {
		mls.logger.Error("Future risk prediction failed",
			zap.String("model", modelName),
			zap.Int("horizon_months", horizonMonths),
			zap.Error(err))
		return nil, fmt.Errorf("future prediction failed: %w", err)
	}

	duration := time.Since(startTime)

	mls.logger.Info("Future risk prediction completed",
		zap.String("model", modelName),
		zap.String("business_id", prediction.BusinessID),
		zap.Int("horizon_months", horizonMonths),
		zap.Float64("predicted_score", prediction.PredictedScore),
		zap.String("predicted_level", string(prediction.PredictedLevel)),
		zap.Duration("duration", duration))

	return prediction, nil
}

// TrainModel trains a new model with the provided training data
func (mls *MLService) TrainModel(ctx context.Context, modelType string, trainingData []*models.RiskAssessment, config *training.TrainingConfig) (*training.TrainingResult, error) {
	mls.logger.Info("Starting model training",
		zap.String("model_type", modelType),
		zap.Int("training_data_size", len(trainingData)))

	startTime := time.Now()

	var result *training.TrainingResult
	var err error

	switch modelType {
	case "xgboost":
		result, err = mls.trainer.TrainXGBoostModel(ctx, trainingData, config)
	default:
		return nil, fmt.Errorf("unsupported model type: %s", modelType)
	}

	if err != nil {
		mls.logger.Error("Model training failed",
			zap.String("model_type", modelType),
			zap.Error(err))
		return nil, fmt.Errorf("training failed: %w", err)
	}

	totalTime := time.Since(startTime)

	mls.logger.Info("Model training completed",
		zap.String("model_type", modelType),
		zap.Float64("accuracy", result.ValidationResult.Accuracy),
		zap.Float64("precision", result.ValidationResult.Precision),
		zap.Float64("recall", result.ValidationResult.Recall),
		zap.Float64("f1_score", result.ValidationResult.F1Score),
		zap.Duration("training_time", result.TrainingTime),
		zap.Duration("total_time", totalTime))

	return result, nil
}

// ValidateModel validates a model's performance
func (mls *MLService) ValidateModel(ctx context.Context, modelName string, testData []*models.RiskAssessment) (*mlmodels.ValidationResult, error) {
	model, err := mls.modelManager.GetModel(modelName)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	mls.logger.Info("Starting model validation",
		zap.String("model", modelName),
		zap.Int("test_data_size", len(testData)))

	startTime := time.Now()

	result, err := model.ValidateModel(ctx, testData)
	if err != nil {
		mls.logger.Error("Model validation failed",
			zap.String("model", modelName),
			zap.Error(err))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	duration := time.Since(startTime)

	mls.logger.Info("Model validation completed",
		zap.String("model", modelName),
		zap.Float64("accuracy", result.Accuracy),
		zap.Float64("precision", result.Precision),
		zap.Float64("recall", result.Recall),
		zap.Float64("f1_score", result.F1Score),
		zap.Duration("duration", duration))

	return result, nil
}

// GetModelInfo returns information about a model
func (mls *MLService) GetModelInfo(modelName string) (*mlmodels.ModelInfo, error) {
	model, err := mls.modelManager.GetModel(modelName)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	return model.GetModelInfo(), nil
}

// ListModels returns a list of available models
func (mls *MLService) ListModels() []string {
	return mls.modelManager.ListModels()
}

// GetFeatureExtractor returns the feature extractor
func (mls *MLService) GetFeatureExtractor() *mlmodels.FeatureExtractor {
	return mls.modelManager.GetFeatureExtractor()
}

// GetRiskLevelEncoder returns the risk level encoder
func (mls *MLService) GetRiskLevelEncoder() *mlmodels.RiskLevelEncoder {
	return mls.modelManager.GetRiskLevelEncoder()
}

// GenerateMockTrainingData generates mock training data for testing
func (mls *MLService) GenerateMockTrainingData(count int) []*models.RiskAssessment {
	return mls.trainer.GenerateMockTrainingData(count)
}

// Health checks the health of the ML service
func (mls *MLService) Health(ctx context.Context) error {
	// Check if models are loaded
	models := mls.modelManager.ListModels()
	if len(models) == 0 {
		return fmt.Errorf("no models loaded")
	}

	// Check if models are accessible
	for _, modelName := range models {
		_, err := mls.modelManager.GetModel(modelName)
		if err != nil {
			return fmt.Errorf("model %s is not accessible: %w", modelName, err)
		}
	}

	return nil
}
