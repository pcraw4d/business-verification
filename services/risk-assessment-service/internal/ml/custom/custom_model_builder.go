package custom

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// CustomRiskModel represents a custom risk model for enterprise customers
type CustomRiskModel struct {
	ID              string                       `json:"id" db:"id"`
	TenantID        string                       `json:"tenant_id" db:"tenant_id"`
	Name            string                       `json:"name" db:"name"`
	Description     string                       `json:"description" db:"description"`
	BaseModel       string                       `json:"base_model" db:"base_model"`
	CustomFactors   []CustomRiskFactor           `json:"custom_factors" db:"custom_factors"`
	FactorWeights   map[string]float64           `json:"factor_weights" db:"factor_weights"`
	Thresholds      map[models.RiskLevel]float64 `json:"thresholds" db:"thresholds"`
	ValidationRules []ValidationRule             `json:"validation_rules" db:"validation_rules"`
	IsActive        bool                         `json:"is_active" db:"is_active"`
	Version         int                          `json:"version" db:"version"`
	CreatedAt       time.Time                    `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at" db:"updated_at"`
	CreatedBy       string                       `json:"created_by" db:"created_by"`
	UpdatedBy       string                       `json:"updated_by" db:"updated_by"`
	Metadata        map[string]interface{}       `json:"metadata" db:"metadata"`
}

// CustomRiskFactor represents a custom risk factor that can be added to a model
type CustomRiskFactor struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category"`
	Weight          float64                `json:"weight"`
	DataType        string                 `json:"data_type"` // "numeric", "boolean", "categorical", "text"
	ValidationRules []FactorValidationRule `json:"validation_rules"`
	ScoringFunction string                 `json:"scoring_function"` // "linear", "exponential", "logarithmic", "custom"
	ScoringParams   map[string]interface{} `json:"scoring_params"`
	IsRequired      bool                   `json:"is_required"`
	DefaultValue    interface{}            `json:"default_value"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ValidationRule represents a validation rule for the custom model
type ValidationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	RuleType    string                 `json:"rule_type"` // "threshold", "range", "pattern", "custom"
	Parameters  map[string]interface{} `json:"parameters"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
}

// FactorValidationRule represents a validation rule for a specific factor
type FactorValidationRule struct {
	ID           string                 `json:"id"`
	RuleType     string                 `json:"rule_type"` // "min", "max", "range", "pattern", "required"
	Parameters   map[string]interface{} `json:"parameters"`
	ErrorMessage string                 `json:"error_message"`
}

// CustomModelBuilder provides functionality to build and manage custom risk models
type CustomModelBuilder struct {
	Repository CustomModelRepository
	logger     *zap.Logger
}

// NewCustomModelBuilder creates a new custom model builder
func NewCustomModelBuilder(repository CustomModelRepository, logger *zap.Logger) *CustomModelBuilder {
	return &CustomModelBuilder{
		Repository: repository,
		logger:     logger,
	}
}

// CreateCustomModel creates a new custom risk model
func (cmb *CustomModelBuilder) CreateCustomModel(ctx context.Context, tenantID string, req *CreateCustomModelRequest) (*CustomRiskModel, error) {
	cmb.logger.Info("Creating custom risk model",
		zap.String("tenant_id", tenantID),
		zap.String("name", req.Name),
		zap.String("base_model", string(req.BaseModel)))

	// Validate the request
	if err := cmb.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create the custom model
	model := &CustomRiskModel{
		ID:              cmb.generateModelID(),
		TenantID:        tenantID,
		Name:            req.Name,
		Description:     req.Description,
		BaseModel:       req.BaseModel,
		CustomFactors:   req.CustomFactors,
		FactorWeights:   req.FactorWeights,
		Thresholds:      req.Thresholds,
		ValidationRules: req.ValidationRules,
		IsActive:        true,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		CreatedBy:       req.CreatedBy,
		UpdatedBy:       req.CreatedBy,
		Metadata:        req.Metadata,
	}

	// Validate the model configuration
	if err := cmb.validateModelConfiguration(model); err != nil {
		return nil, fmt.Errorf("model validation failed: %w", err)
	}

	// Save to repository
	if err := cmb.Repository.SaveCustomModel(ctx, model); err != nil {
		return nil, fmt.Errorf("failed to save custom model: %w", err)
	}

	cmb.logger.Info("Custom risk model created successfully",
		zap.String("model_id", model.ID),
		zap.String("tenant_id", tenantID))

	return model, nil
}

// UpdateCustomModel updates an existing custom risk model
func (cmb *CustomModelBuilder) UpdateCustomModel(ctx context.Context, tenantID, modelID string, req *UpdateCustomModelRequest) (*CustomRiskModel, error) {
	cmb.logger.Info("Updating custom risk model",
		zap.String("tenant_id", tenantID),
		zap.String("model_id", modelID))

	// Get existing model
	existingModel, err := cmb.Repository.GetCustomModel(ctx, tenantID, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing model: %w", err)
	}

	if existingModel == nil {
		return nil, fmt.Errorf("custom model not found")
	}

	// Validate the update request
	if err := cmb.validateUpdateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update the model
	updatedModel := *existingModel
	updatedModel.Name = req.Name
	updatedModel.Description = req.Description
	updatedModel.CustomFactors = req.CustomFactors
	updatedModel.FactorWeights = req.FactorWeights
	updatedModel.Thresholds = req.Thresholds
	updatedModel.ValidationRules = req.ValidationRules
	updatedModel.UpdatedAt = time.Now()
	updatedModel.UpdatedBy = req.UpdatedBy
	updatedModel.Version++

	if req.Metadata != nil {
		updatedModel.Metadata = req.Metadata
	}

	// Validate the updated model configuration
	if err := cmb.validateModelConfiguration(&updatedModel); err != nil {
		return nil, fmt.Errorf("model validation failed: %w", err)
	}

	// Save to repository
	if err := cmb.Repository.SaveCustomModel(ctx, &updatedModel); err != nil {
		return nil, fmt.Errorf("failed to save updated model: %w", err)
	}

	cmb.logger.Info("Custom risk model updated successfully",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID),
		zap.Int("version", updatedModel.Version))

	return &updatedModel, nil
}

// ValidateCustomModel validates a custom model configuration
func (cmb *CustomModelBuilder) ValidateCustomModel(ctx context.Context, tenantID, modelID string) (*ModelValidationResult, error) {
	cmb.logger.Info("Validating custom risk model",
		zap.String("tenant_id", tenantID),
		zap.String("model_id", modelID))

	// Get the model
	model, err := cmb.Repository.GetCustomModel(ctx, tenantID, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	if model == nil {
		return nil, fmt.Errorf("custom model not found")
	}

	// Perform validation
	result := &ModelValidationResult{
		ModelID:     modelID,
		IsValid:     true,
		Errors:      []string{},
		Warnings:    []string{},
		ValidatedAt: time.Now(),
	}

	// Validate model configuration
	if err := cmb.validateModelConfiguration(model); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// Validate factor weights sum to 1.0
	totalWeight := 0.0
	for _, weight := range model.FactorWeights {
		totalWeight += weight
	}
	if totalWeight < 0.99 || totalWeight > 1.01 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Factor weights sum to %.3f, should be 1.0", totalWeight))
	}

	// Validate thresholds are in ascending order
	if err := cmb.validateThresholds(model.Thresholds); err != nil {
		result.Errors = append(result.Errors, err.Error())
		result.IsValid = false
	}

	// Validate custom factors
	for _, factor := range model.CustomFactors {
		if err := cmb.validateCustomFactor(factor); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Factor %s: %s", factor.Name, err.Error()))
			result.IsValid = false
		}
	}

	cmb.logger.Info("Custom model validation completed",
		zap.String("model_id", modelID),
		zap.Bool("is_valid", result.IsValid),
		zap.Int("error_count", len(result.Errors)),
		zap.Int("warning_count", len(result.Warnings)))

	return result, nil
}

// TestCustomModel tests a custom model with sample data
func (cmb *CustomModelBuilder) TestCustomModel(ctx context.Context, tenantID, modelID string, testData *TestModelRequest) (*TestModelResult, error) {
	cmb.logger.Info("Testing custom risk model",
		zap.String("tenant_id", tenantID),
		zap.String("model_id", modelID))

	// Get the model
	model, err := cmb.Repository.GetCustomModel(ctx, tenantID, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	if model == nil {
		return nil, fmt.Errorf("custom model not found")
	}

	// Create test result
	result := &TestModelResult{
		ModelID:      modelID,
		TestData:     testData,
		RiskScore:    0.0,
		RiskLevel:    models.RiskLevelLow,
		FactorScores: make(map[string]float64),
		TestedAt:     time.Now(),
	}

	// Calculate risk score using the custom model
	riskScore, factorScores, err := cmb.calculateCustomRiskScore(model, testData.BusinessData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate risk score: %w", err)
	}

	result.RiskScore = riskScore
	result.RiskLevel = models.ConvertScoreToRiskLevel(riskScore)
	result.FactorScores = factorScores

	cmb.logger.Info("Custom model test completed",
		zap.String("model_id", modelID),
		zap.Float64("risk_score", riskScore),
		zap.String("risk_level", string(result.RiskLevel)))

	return result, nil
}

// validateCreateRequest validates a create custom model request
func (cmb *CustomModelBuilder) validateCreateRequest(req *CreateCustomModelRequest) error {
	if req.Name == "" {
		return fmt.Errorf("model name is required")
	}
	if req.BaseModel == "" {
		return fmt.Errorf("base model is required")
	}
	if req.CreatedBy == "" {
		return fmt.Errorf("created by is required")
	}
	return nil
}

// validateUpdateRequest validates an update custom model request
func (cmb *CustomModelBuilder) validateUpdateRequest(req *UpdateCustomModelRequest) error {
	if req.Name == "" {
		return fmt.Errorf("model name is required")
	}
	if req.UpdatedBy == "" {
		return fmt.Errorf("updated by is required")
	}
	return nil
}

// validateModelConfiguration validates the overall model configuration
func (cmb *CustomModelBuilder) validateModelConfiguration(model *CustomRiskModel) error {
	// Validate that at least one custom factor exists
	if len(model.CustomFactors) == 0 {
		return fmt.Errorf("at least one custom factor is required")
	}

	// Validate factor weights
	if len(model.FactorWeights) == 0 {
		return fmt.Errorf("factor weights are required")
	}

	// Validate thresholds
	if len(model.Thresholds) == 0 {
		return fmt.Errorf("risk level thresholds are required")
	}

	return nil
}

// validateThresholds validates that thresholds are in ascending order
func (cmb *CustomModelBuilder) validateThresholds(thresholds map[models.RiskLevel]float64) error {
	// Expected order: Low < Medium < High < Critical
	expectedOrder := []models.RiskLevel{
		models.RiskLevelLow,
		models.RiskLevelMedium,
		models.RiskLevelHigh,
		models.RiskLevelCritical,
	}

	for i := 1; i < len(expectedOrder); i++ {
		current := expectedOrder[i]
		previous := expectedOrder[i-1]

		if thresholds[current] <= thresholds[previous] {
			return fmt.Errorf("threshold for %s (%.3f) must be greater than %s (%.3f)",
				current, thresholds[current], previous, thresholds[previous])
		}
	}

	return nil
}

// validateCustomFactor validates a custom risk factor
func (cmb *CustomModelBuilder) validateCustomFactor(factor CustomRiskFactor) error {
	if factor.Name == "" {
		return fmt.Errorf("factor name is required")
	}
	if factor.Weight < 0 || factor.Weight > 1 {
		return fmt.Errorf("factor weight must be between 0 and 1")
	}
	if factor.DataType == "" {
		return fmt.Errorf("data type is required")
	}
	return nil
}

// calculateCustomRiskScore calculates the risk score using the custom model
func (cmb *CustomModelBuilder) calculateCustomRiskScore(model *CustomRiskModel, businessData map[string]interface{}) (float64, map[string]float64, error) {
	factorScores := make(map[string]float64)
	totalScore := 0.0

	// Calculate score for each custom factor
	for _, factor := range model.CustomFactors {
		score, err := cmb.calculateFactorScore(factor, businessData)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to calculate score for factor %s: %w", factor.Name, err)
		}

		factorScores[factor.ID] = score
		totalScore += score * factor.Weight
	}

	return totalScore, factorScores, nil
}

// calculateFactorScore calculates the score for a specific factor
func (cmb *CustomModelBuilder) calculateFactorScore(factor CustomRiskFactor, businessData map[string]interface{}) (float64, error) {
	// Get the value for this factor from business data
	value, exists := businessData[factor.ID]
	if !exists {
		if factor.IsRequired {
			return 0, fmt.Errorf("required factor %s not found in business data", factor.ID)
		}
		value = factor.DefaultValue
	}

	// Apply scoring function based on data type and scoring function
	switch factor.DataType {
	case "numeric":
		return cmb.calculateNumericScore(factor, value)
	case "boolean":
		return cmb.calculateBooleanScore(factor, value)
	case "categorical":
		return cmb.calculateCategoricalScore(factor, value)
	case "text":
		return cmb.calculateTextScore(factor, value)
	default:
		return 0, fmt.Errorf("unsupported data type: %s", factor.DataType)
	}
}

// calculateNumericScore calculates score for numeric data
func (cmb *CustomModelBuilder) calculateNumericScore(factor CustomRiskFactor, value interface{}) (float64, error) {
	numValue, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("expected numeric value, got %T", value)
	}

	// Apply scoring function
	switch factor.ScoringFunction {
	case "linear":
		return cmb.applyLinearScoring(factor, numValue)
	case "exponential":
		return cmb.applyExponentialScoring(factor, numValue)
	case "logarithmic":
		return cmb.applyLogarithmicScoring(factor, numValue)
	default:
		return 0, fmt.Errorf("unsupported scoring function: %s", factor.ScoringFunction)
	}
}

// calculateBooleanScore calculates score for boolean data
func (cmb *CustomModelBuilder) calculateBooleanScore(factor CustomRiskFactor, value interface{}) (float64, error) {
	boolValue, ok := value.(bool)
	if !ok {
		return 0, fmt.Errorf("expected boolean value, got %T", value)
	}

	// Get scoring parameters
	trueScore, ok := factor.ScoringParams["true_score"].(float64)
	if !ok {
		trueScore = 1.0
	}
	falseScore, ok := factor.ScoringParams["false_score"].(float64)
	if !ok {
		falseScore = 0.0
	}

	if boolValue {
		return trueScore, nil
	}
	return falseScore, nil
}

// calculateCategoricalScore calculates score for categorical data
func (cmb *CustomModelBuilder) calculateCategoricalScore(factor CustomRiskFactor, value interface{}) (float64, error) {
	strValue, ok := value.(string)
	if !ok {
		return 0, fmt.Errorf("expected string value, got %T", value)
	}

	// Get category scores from scoring parameters
	categoryScores, ok := factor.ScoringParams["category_scores"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("category_scores not found in scoring parameters")
	}

	score, exists := categoryScores[strValue]
	if !exists {
		// Use default score if category not found
		defaultScore, ok := factor.ScoringParams["default_score"].(float64)
		if !ok {
			defaultScore = 0.5
		}
		return defaultScore, nil
	}

	scoreFloat, ok := score.(float64)
	if !ok {
		return 0, fmt.Errorf("invalid score type for category %s", strValue)
	}

	return scoreFloat, nil
}

// calculateTextScore calculates score for text data
func (cmb *CustomModelBuilder) calculateTextScore(factor CustomRiskFactor, value interface{}) (float64, error) {
	strValue, ok := value.(string)
	if !ok {
		return 0, fmt.Errorf("expected string value, got %T", value)
	}

	// Simple text scoring based on length and keywords
	// This is a basic implementation - could be enhanced with NLP
	baseScore := 0.5

	// Adjust score based on text length
	length := len(strValue)
	if length > 100 {
		baseScore += 0.2
	} else if length < 10 {
		baseScore -= 0.2
	}

	// Adjust score based on keywords (if specified)
	keywords, ok := factor.ScoringParams["keywords"].(map[string]float64)
	if ok {
		for keyword, adjustment := range keywords {
			if contains(strValue, keyword) {
				baseScore += adjustment
			}
		}
	}

	// Ensure score is between 0 and 1
	if baseScore < 0 {
		baseScore = 0
	} else if baseScore > 1 {
		baseScore = 1
	}

	return baseScore, nil
}

// applyLinearScoring applies linear scoring function
func (cmb *CustomModelBuilder) applyLinearScoring(factor CustomRiskFactor, value float64) (float64, error) {
	minVal, ok := factor.ScoringParams["min_value"].(float64)
	if !ok {
		minVal = 0
	}
	maxVal, ok := factor.ScoringParams["max_value"].(float64)
	if !ok {
		maxVal = 100
	}

	if maxVal == minVal {
		return 0.5, nil
	}

	// Normalize to 0-1 range
	normalized := (value - minVal) / (maxVal - minVal)
	if normalized < 0 {
		normalized = 0
	} else if normalized > 1 {
		normalized = 1
	}

	return normalized, nil
}

// applyExponentialScoring applies exponential scoring function
func (cmb *CustomModelBuilder) applyExponentialScoring(factor CustomRiskFactor, value float64) (float64, error) {
	base, ok := factor.ScoringParams["base"].(float64)
	if !ok {
		base = 2.0
	}
	exponent, ok := factor.ScoringParams["exponent"].(float64)
	if !ok {
		exponent = 1.0
	}

	score := math.Pow(base, value*exponent)

	// Normalize to 0-1 range
	maxScore := math.Pow(base, exponent)
	normalized := score / maxScore

	return normalized, nil
}

// applyLogarithmicScoring applies logarithmic scoring function
func (cmb *CustomModelBuilder) applyLogarithmicScoring(factor CustomRiskFactor, value float64) (float64, error) {
	base, ok := factor.ScoringParams["base"].(float64)
	if !ok {
		base = 10.0
	}

	if value <= 0 {
		return 0, nil
	}

	score := math.Log(value) / math.Log(base)

	// Normalize to 0-1 range (assuming max value of 1000)
	maxValue := 1000.0
	maxScore := math.Log(maxValue) / math.Log(base)
	normalized := score / maxScore

	if normalized < 0 {
		normalized = 0
	} else if normalized > 1 {
		normalized = 1
	}

	return normalized, nil
}

// generateModelID generates a unique model ID
func (cmb *CustomModelBuilder) generateModelID() string {
	return fmt.Sprintf("custom_model_%d", time.Now().UnixNano())
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// Request and Response Types

// CreateCustomModelRequest represents a request to create a custom model
type CreateCustomModelRequest struct {
	Name            string                       `json:"name"`
	Description     string                       `json:"description"`
	BaseModel       string                       `json:"base_model"`
	CustomFactors   []CustomRiskFactor           `json:"custom_factors"`
	FactorWeights   map[string]float64           `json:"factor_weights"`
	Thresholds      map[models.RiskLevel]float64 `json:"thresholds"`
	ValidationRules []ValidationRule             `json:"validation_rules"`
	CreatedBy       string                       `json:"created_by"`
	Metadata        map[string]interface{}       `json:"metadata"`
}

// UpdateCustomModelRequest represents a request to update a custom model
type UpdateCustomModelRequest struct {
	Name            string                       `json:"name"`
	Description     string                       `json:"description"`
	CustomFactors   []CustomRiskFactor           `json:"custom_factors"`
	FactorWeights   map[string]float64           `json:"factor_weights"`
	Thresholds      map[models.RiskLevel]float64 `json:"thresholds"`
	ValidationRules []ValidationRule             `json:"validation_rules"`
	UpdatedBy       string                       `json:"updated_by"`
	Metadata        map[string]interface{}       `json:"metadata"`
}

// ModelValidationResult represents the result of model validation
type ModelValidationResult struct {
	ModelID     string    `json:"model_id"`
	IsValid     bool      `json:"is_valid"`
	Errors      []string  `json:"errors"`
	Warnings    []string  `json:"warnings"`
	ValidatedAt time.Time `json:"validated_at"`
}

// TestModelRequest represents a request to test a custom model
type TestModelRequest struct {
	BusinessData map[string]interface{} `json:"business_data"`
}

// TestModelResult represents the result of testing a custom model
type TestModelResult struct {
	ModelID      string             `json:"model_id"`
	TestData     *TestModelRequest  `json:"test_data"`
	RiskScore    float64            `json:"risk_score"`
	RiskLevel    models.RiskLevel   `json:"risk_level"`
	FactorScores map[string]float64 `json:"factor_scores"`
	TestedAt     time.Time          `json:"tested_at"`
}
