package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/ml/ensemble"
	mlmodels "kyb-platform/services/risk-assessment-service/internal/ml/models"
	"kyb-platform/services/risk-assessment-service/internal/ml/monitoring"
	"kyb-platform/services/risk-assessment-service/internal/ml/training"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// MLService provides machine learning capabilities for risk assessment
type MLService struct {
	modelManager     *mlmodels.ModelManager
	trainer          *training.ModelTrainer
	ensembleRouter   *ensemble.EnsembleRouter
	metricsCollector *monitoring.MetricsCollector
	logger           *zap.Logger
}

// NewMLService creates a new ML service
func NewMLService(logger *zap.Logger) *MLService {
	return &MLService{
		modelManager:     mlmodels.NewModelManagerWithLogger(logger),
		trainer:          training.NewModelTrainer(),
		ensembleRouter:   nil, // Will be initialized after models are loaded
		metricsCollector: monitoring.NewMetricsCollector(logger),
		logger:           logger,
	}
}

// InitializeModels initializes and loads the ML models
func (mls *MLService) InitializeModels(ctx context.Context) error {
	mls.logger.Info("Initializing ML models")

	// Initialize models using the model manager
	if err := mls.modelManager.InitializeModels(ctx); err != nil {
		return fmt.Errorf("failed to initialize models: %w", err)
	}

	// Initialize ensemble router if both models are available
	availableModels := mls.modelManager.ListModels()
	if len(availableModels) >= 2 {
		xgbModel, err := mls.modelManager.GetModel("xgboost")
		if err != nil {
			mls.logger.Warn("XGBoost model not available for ensemble", zap.Error(err))
		}

		lstmModel, err := mls.modelManager.GetModel("lstm")
		if err != nil {
			mls.logger.Warn("LSTM model not available for ensemble", zap.Error(err))
		}

		if xgbModel != nil && lstmModel != nil {
			mls.ensembleRouter = ensemble.NewEnsembleRouter(xgbModel, lstmModel, mls.logger)
			mls.logger.Info("Ensemble router initialized successfully")
		}
	}

	mls.logger.Info("ML models initialized successfully",
		zap.Strings("available_models", mls.modelManager.ListModels()),
		zap.Bool("ensemble_enabled", mls.ensembleRouter != nil))

	return nil
}

// PredictRisk performs risk prediction using the specified model or ensemble
func (mls *MLService) PredictRisk(ctx context.Context, modelName string, business *models.RiskAssessmentRequest) (*models.RiskAssessment, error) {
	startTime := time.Now()
	horizonMonths := business.PredictionHorizon
	if horizonMonths == 0 {
		horizonMonths = 3 // Default horizon
	}

	// Handle ensemble prediction
	if modelName == "ensemble" || modelName == "auto" {
		if mls.ensembleRouter == nil {
			mls.metricsCollector.RecordInference(modelName, time.Since(startTime), horizonMonths, fmt.Errorf("ensemble router not available"))
			return nil, fmt.Errorf("ensemble router not available")
		}
		prediction, err := mls.PredictRiskWithEnsemble(ctx, business)
		mls.metricsCollector.RecordInference(modelName, time.Since(startTime), horizonMonths, err)
		return prediction, err
	}

	// Handle individual model prediction
	model, err := mls.modelManager.GetModel(modelName)
	if err != nil {
		mls.metricsCollector.RecordInference(modelName, time.Since(startTime), horizonMonths, err)
		return nil, fmt.Errorf("model not found: %w", err)
	}

	prediction, err := model.Predict(ctx, business)
	duration := time.Since(startTime)

	// Record metrics
	mls.metricsCollector.RecordInference(modelName, duration, horizonMonths, err)

	if err != nil {
		mls.logger.Error("Risk prediction failed",
			zap.String("model", modelName),
			zap.Error(err))
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	mls.logger.Info("Risk prediction completed",
		zap.String("model", modelName),
		zap.String("assessment_id", prediction.ID),
		zap.Float64("risk_score", prediction.RiskScore),
		zap.String("risk_level", string(prediction.RiskLevel)),
		zap.Duration("duration", duration))

	return prediction, nil
}

// PredictRiskWithEnsemble performs ensemble risk prediction with smart routing
func (mls *MLService) PredictRiskWithEnsemble(ctx context.Context, business *models.RiskAssessmentRequest) (*models.RiskAssessment, error) {
	if mls.ensembleRouter == nil {
		return nil, fmt.Errorf("ensemble router not available")
	}

	// Determine prediction horizon
	horizon := business.PredictionHorizon
	if horizon == 0 {
		horizon = 3 // Default to 3 months
	}

	// Route to appropriate model or ensemble
	modelType := mls.ensembleRouter.Route(horizon)

	mls.logger.Info("Ensemble routing decision",
		zap.Int("horizon_months", horizon),
		zap.String("selected_model", modelType))

	switch modelType {
	case "xgboost":
		return mls.PredictRisk(ctx, "xgboost", business)
	case "lstm":
		return mls.PredictRisk(ctx, "lstm", business)
	case "ensemble":
		return mls.ensembleRouter.PredictWithEnsemble(ctx, business)
	default:
		return nil, fmt.Errorf("unknown model type: %s", modelType)
	}
}

// PredictFutureRisk performs future risk prediction using the specified model or ensemble
func (mls *MLService) PredictFutureRisk(ctx context.Context, modelName string, business *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	startTime := time.Now()

	// Handle ensemble prediction
	if modelName == "ensemble" || modelName == "auto" {
		if mls.ensembleRouter == nil {
			mls.metricsCollector.RecordInference(modelName, time.Since(startTime), horizonMonths, fmt.Errorf("ensemble router not available"))
			return nil, fmt.Errorf("ensemble router not available")
		}
		prediction, err := mls.PredictFutureRiskWithEnsemble(ctx, business, horizonMonths)
		mls.metricsCollector.RecordInference(modelName, time.Since(startTime), horizonMonths, err)
		return prediction, err
	}

	// Handle individual model prediction
	model, err := mls.modelManager.GetModel(modelName)
	if err != nil {
		mls.metricsCollector.RecordInference(modelName, time.Since(startTime), horizonMonths, err)
		return nil, fmt.Errorf("model not found: %w", err)
	}

	prediction, err := model.PredictFuture(ctx, business, horizonMonths)
	duration := time.Since(startTime)

	// Record metrics
	mls.metricsCollector.RecordInference(modelName, duration, horizonMonths, err)

	if err != nil {
		mls.logger.Error("Future risk prediction failed",
			zap.String("model", modelName),
			zap.Int("horizon_months", horizonMonths),
			zap.Error(err))
		return nil, fmt.Errorf("future prediction failed: %w", err)
	}

	mls.logger.Info("Future risk prediction completed",
		zap.String("model", modelName),
		zap.String("business_id", prediction.BusinessID),
		zap.Int("horizon_months", horizonMonths),
		zap.Float64("predicted_score", prediction.PredictedScore),
		zap.String("predicted_level", string(prediction.PredictedLevel)),
		zap.Duration("duration", duration))

	return prediction, nil
}

// PredictFutureRiskWithEnsemble performs ensemble future risk prediction with smart routing
func (mls *MLService) PredictFutureRiskWithEnsemble(ctx context.Context, business *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	if mls.ensembleRouter == nil {
		return nil, fmt.Errorf("ensemble router not available")
	}

	// Route to appropriate model or ensemble
	modelType := mls.ensembleRouter.Route(horizonMonths)

	mls.logger.Info("Ensemble future routing decision",
		zap.Int("horizon_months", horizonMonths),
		zap.String("selected_model", modelType))

	switch modelType {
	case "xgboost":
		return mls.PredictFutureRisk(ctx, "xgboost", business, horizonMonths)
	case "lstm":
		return mls.PredictFutureRisk(ctx, "lstm", business, horizonMonths)
	case "ensemble":
		return mls.ensembleRouter.PredictFutureWithEnsemble(ctx, business, horizonMonths)
	default:
		return nil, fmt.Errorf("unknown model type: %s", modelType)
	}
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
	models := mls.modelManager.ListModels()
	if mls.ensembleRouter != nil {
		models = append(models, "ensemble", "auto")
	}
	return models
}

// GetEnsembleInfo returns information about the ensemble configuration
func (mls *MLService) GetEnsembleInfo() map[string]interface{} {
	info := make(map[string]interface{})

	if mls.ensembleRouter == nil {
		info["available"] = false
		info["reason"] = "ensemble router not initialized"
		return info
	}

	info["available"] = true
	info["routing_strategy"] = "horizon-based"
	info["supported_models"] = []string{"xgboost", "lstm", "ensemble"}
	info["routing_rules"] = map[string]string{
		"1-3_months":  "xgboost (80% weight)",
		"3-6_months":  "ensemble (50% xgb, 50% lstm)",
		"6-12_months": "lstm (80% weight)",
	}

	return info
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

// GetMetricsCollector returns the metrics collector
func (mls *MLService) GetMetricsCollector() *monitoring.MetricsCollector {
	return mls.metricsCollector
}
