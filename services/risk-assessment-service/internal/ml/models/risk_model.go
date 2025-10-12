package models

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// RiskModel defines the interface for risk prediction models
type RiskModel interface {
	// Predict performs risk prediction for a given business
	Predict(ctx context.Context, business *models.RiskAssessmentRequest) (*models.RiskAssessment, error)

	// PredictFuture performs future risk prediction for a given horizon
	PredictFuture(ctx context.Context, business *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error)

	// GetModelInfo returns information about the model
	GetModelInfo() *ModelInfo

	// LoadModel loads the model from storage
	LoadModel(ctx context.Context, modelPath string) error

	// SaveModel saves the model to storage
	SaveModel(ctx context.Context, modelPath string) error

	// ValidateModel validates the model performance
	ValidateModel(ctx context.Context, testData []*models.RiskAssessment) (*ValidationResult, error)
}

// ModelInfo contains information about a machine learning model
type ModelInfo struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Type            string                 `json:"type"`
	TrainingDate    time.Time              `json:"training_date"`
	Accuracy        float64                `json:"accuracy"`
	Precision       float64                `json:"precision"`
	Recall          float64                `json:"recall"`
	F1Score         float64                `json:"f1_score"`
	Features        []string               `json:"features"`
	Hyperparameters map[string]interface{} `json:"hyperparameters"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ValidationResult contains model validation results
type ValidationResult struct {
	Accuracy          float64                   `json:"accuracy"`
	Precision         float64                   `json:"precision"`
	Recall            float64                   `json:"recall"`
	F1Score           float64                   `json:"f1_score"`
	ConfusionMatrix   map[string]map[string]int `json:"confusion_matrix"`
	FeatureImportance map[string]float64        `json:"feature_importance"`
	ValidationDate    time.Time                 `json:"validation_date"`
	TestDataSize      int                       `json:"test_data_size"`
}

// FeatureExtractor extracts features from business data for ML models
type FeatureExtractor struct {
	// Industry mapping for categorical encoding
	IndustryMapping map[string]int `json:"industry_mapping"`

	// Country mapping for categorical encoding
	CountryMapping map[string]int `json:"country_mapping"`

	// Risk factor weights
	RiskFactorWeights map[string]float64 `json:"risk_factor_weights"`
}

// NewFeatureExtractor creates a new feature extractor
func NewFeatureExtractor() *FeatureExtractor {
	return &FeatureExtractor{
		IndustryMapping: map[string]int{
			"technology":     1,
			"financial":      2,
			"healthcare":     3,
			"manufacturing":  4,
			"retail":         5,
			"real_estate":    6,
			"energy":         7,
			"transportation": 8,
			"education":      9,
			"other":          10,
		},
		CountryMapping: map[string]int{
			"US": 1,
			"CA": 2,
			"GB": 3,
			"DE": 4,
			"FR": 5,
			"IT": 6,
			"ES": 7,
			"NL": 8,
			"AU": 9,
			"JP": 10,
		},
		RiskFactorWeights: map[string]float64{
			"financial":    0.3,
			"operational":  0.25,
			"compliance":   0.2,
			"reputational": 0.15,
			"regulatory":   0.1,
		},
	}
}

// ExtractFeatures extracts numerical features from business data
func (fe *FeatureExtractor) ExtractFeatures(business *models.RiskAssessmentRequest) ([]float64, error) {
	features := make([]float64, 0, 20)

	// Business name length (normalized)
	nameLength := float64(len(business.BusinessName)) / 100.0
	features = append(features, nameLength)

	// Business address length (normalized)
	addressLength := float64(len(business.BusinessAddress)) / 200.0
	features = append(features, addressLength)

	// Industry encoding
	industryCode, exists := fe.IndustryMapping[business.Industry]
	if !exists {
		industryCode = fe.IndustryMapping["other"]
	}
	features = append(features, float64(industryCode))

	// Country encoding
	countryCode, exists := fe.CountryMapping[business.Country]
	if !exists {
		countryCode = 0 // Unknown country
	}
	features = append(features, float64(countryCode))

	// Has phone number (binary)
	hasPhone := 0.0
	if business.Phone != "" {
		hasPhone = 1.0
	}
	features = append(features, hasPhone)

	// Has email (binary)
	hasEmail := 0.0
	if business.Email != "" {
		hasEmail = 1.0
	}
	features = append(features, hasEmail)

	// Has website (binary)
	hasWebsite := 0.0
	if business.Website != "" {
		hasWebsite = 1.0
	}
	features = append(features, hasWebsite)

	// Prediction horizon
	horizon := float64(business.PredictionHorizon)
	if horizon == 0 {
		horizon = 3.0 // Default to 3 months
	}
	features = append(features, horizon/12.0) // Normalize to years

	// Additional features from metadata
	if business.Metadata != nil {
		// Annual revenue (if available)
		if revenue, ok := business.Metadata["annual_revenue"].(float64); ok {
			features = append(features, revenue/1000000.0) // Normalize to millions
		} else {
			features = append(features, 0.0)
		}

		// Employee count (if available)
		if employees, ok := business.Metadata["employee_count"].(float64); ok {
			features = append(features, employees/100.0) // Normalize to hundreds
		} else {
			features = append(features, 0.0)
		}

		// Years in business (if available)
		if years, ok := business.Metadata["years_in_business"].(float64); ok {
			features = append(features, years/10.0) // Normalize to decades
		} else {
			features = append(features, 0.0)
		}
	} else {
		// Add zeros for missing metadata
		features = append(features, 0.0, 0.0, 0.0)
	}

	// Pad with zeros to ensure consistent feature vector length
	for len(features) < 20 {
		features = append(features, 0.0)
	}

	return features, nil
}

// GetFeatureNames returns the names of the features
func (fe *FeatureExtractor) GetFeatureNames() []string {
	return []string{
		"name_length",
		"address_length",
		"industry_code",
		"country_code",
		"has_phone",
		"has_email",
		"has_website",
		"prediction_horizon",
		"annual_revenue",
		"employee_count",
		"years_in_business",
		"feature_11",
		"feature_12",
		"feature_13",
		"feature_14",
		"feature_15",
		"feature_16",
		"feature_17",
		"feature_18",
		"feature_19",
	}
}

// RiskLevelEncoder encodes risk levels to numerical values
type RiskLevelEncoder struct {
	LevelMapping map[models.RiskLevel]float64
}

// NewRiskLevelEncoder creates a new risk level encoder
func NewRiskLevelEncoder() *RiskLevelEncoder {
	return &RiskLevelEncoder{
		LevelMapping: map[models.RiskLevel]float64{
			models.RiskLevelLow:      0.25,
			models.RiskLevelMedium:   0.5,
			models.RiskLevelHigh:     0.75,
			models.RiskLevelCritical: 1.0,
		},
	}
}

// EncodeRiskLevel encodes a risk level to a numerical value
func (rle *RiskLevelEncoder) EncodeRiskLevel(level models.RiskLevel) float64 {
	if value, exists := rle.LevelMapping[level]; exists {
		return value
	}
	return 0.5 // Default to medium risk
}

// DecodeRiskLevel decodes a numerical value to a risk level
func (rle *RiskLevelEncoder) DecodeRiskLevel(value float64) models.RiskLevel {
	switch {
	case value <= 0.25:
		return models.RiskLevelLow
	case value <= 0.5:
		return models.RiskLevelMedium
	case value <= 0.75:
		return models.RiskLevelHigh
	default:
		return models.RiskLevelCritical
	}
}

// ModelManager manages multiple risk prediction models
type ModelManager struct {
	models           map[string]RiskModel
	featureExtractor *FeatureExtractor
	riskLevelEncoder *RiskLevelEncoder
	logger           *zap.Logger
}

// NewModelManager creates a new model manager
func NewModelManager() *ModelManager {
	return &ModelManager{
		models:           make(map[string]RiskModel),
		featureExtractor: NewFeatureExtractor(),
		riskLevelEncoder: NewRiskLevelEncoder(),
		logger:           zap.NewNop(),
	}
}

// NewModelManagerWithLogger creates a new model manager with logger
func NewModelManagerWithLogger(logger *zap.Logger) *ModelManager {
	return &ModelManager{
		models:           make(map[string]RiskModel),
		featureExtractor: NewFeatureExtractor(),
		riskLevelEncoder: NewRiskLevelEncoder(),
		logger:           logger,
	}
}

// InitializeModels initializes and loads all available models
func (mm *ModelManager) InitializeModels(ctx context.Context) error {
	mm.logger.Info("Initializing risk prediction models")

	// Create and register XGBoost model
	xgbModel := NewXGBoostModel("risk_prediction_xgb", "1.0.0")

	// Load the XGBoost model (in a real implementation, this would load from storage)
	xgbModelPath := os.Getenv("XGBOOST_MODEL_PATH")
	if xgbModelPath == "" {
		xgbModelPath = "./models/xgb_model.json" // fallback
	}
	if err := xgbModel.LoadModel(ctx, xgbModelPath); err != nil {
		mm.logger.Warn("Failed to load XGBoost model, using default", zap.Error(err))
	}

	// Register the XGBoost model
	mm.RegisterModel("xgboost", xgbModel)
	mm.logger.Info("XGBoost model registered")

	// Create and register LSTM model
	lstmModel := NewLSTMModel("risk_prediction_lstm", "1.0.0", mm.logger)

	// Load the LSTM model (always succeeds with enhanced placeholder)
	lstmModelPath := os.Getenv("LSTM_MODEL_PATH")
	if lstmModelPath == "" {
		lstmModelPath = "./models/risk_lstm_v1.onnx" // fallback
	}
	if err := lstmModel.LoadModel(ctx, lstmModelPath); err != nil {
		mm.logger.Error("Failed to load LSTM model", zap.Error(err))
		// Still register LSTM with enhanced placeholder implementation
	}

	// Always register the LSTM model (enhanced placeholder is robust)
	mm.RegisterModel("lstm", lstmModel)
	mm.logger.Info("LSTM model registered with enhanced placeholder implementation")

	mm.logger.Info("Model initialization completed",
		zap.Strings("available_models", mm.ListModels()))

	return nil
}

// RegisterModel registers a model with the manager
func (mm *ModelManager) RegisterModel(name string, model RiskModel) {
	mm.models[name] = model
}

// GetModel retrieves a model by name
func (mm *ModelManager) GetModel(name string) (RiskModel, error) {
	model, exists := mm.models[name]
	if !exists {
		return nil, fmt.Errorf("model %s not found", name)
	}
	return model, nil
}

// GetFeatureExtractor returns the feature extractor
func (mm *ModelManager) GetFeatureExtractor() *FeatureExtractor {
	return mm.featureExtractor
}

// GetRiskLevelEncoder returns the risk level encoder
func (mm *ModelManager) GetRiskLevelEncoder() *RiskLevelEncoder {
	return mm.riskLevelEncoder
}

// ListModels returns a list of available models
func (mm *ModelManager) ListModels() []string {
	models := make([]string, 0, len(mm.models))
	for name := range mm.models {
		models = append(models, name)
	}
	return models
}
