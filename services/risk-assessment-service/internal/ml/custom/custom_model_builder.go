package custom

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/ml/industry"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// CustomRiskModel represents a custom risk model for enterprise customers
type CustomRiskModel struct {
	ID              string                       `json:"id"`
	TenantID        string                       `json:"tenant_id"`
	Name            string                       `json:"name"`
	Description     string                       `json:"description"`
	BaseModel       industry.IndustryType        `json:"base_model"`
	CustomFactors   []CustomRiskFactor           `json:"custom_factors"`
	FactorWeights   map[string]float64           `json:"factor_weights"`
	Thresholds      map[models.RiskLevel]float64 `json:"thresholds"`
	ValidationRules []ValidationRule             `json:"validation_rules"`
	IsActive        bool                         `json:"is_active"`
	Version         string                       `json:"version"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
	CreatedBy       string                       `json:"created_by"`
	Metadata        map[string]interface{}       `json:"metadata"`
}

// CustomRiskFactor represents a custom risk factor in the model
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
	Options         []FactorOption         `json:"options,omitempty"` // For categorical factors
	Metadata        map[string]interface{} `json:"metadata"`
}

// FactorOption represents an option for categorical factors
type FactorOption struct {
	Value       string  `json:"value"`
	Label       string  `json:"label"`
	Score       float64 `json:"score"`
	Description string  `json:"description"`
}

// ValidationRule represents a validation rule for the custom model
type ValidationRule struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	RuleType     string                 `json:"rule_type"` // "threshold", "range", "pattern", "custom"
	Parameters   map[string]interface{} `json:"parameters"`
	ErrorMessage string                 `json:"error_message"`
	IsActive     bool                   `json:"is_active"`
}

// FactorValidationRule represents a validation rule for a specific factor
type FactorValidationRule struct {
	RuleType     string                 `json:"rule_type"`
	Parameters   map[string]interface{} `json:"parameters"`
	ErrorMessage string                 `json:"error_message"`
}

// CustomModelBuilder provides functionality to build and configure custom risk models
type CustomModelBuilder struct {
	repository CustomModelRepository
	logger     *zap.Logger
}

// NewCustomModelBuilder creates a new custom model builder
func NewCustomModelBuilder(repository CustomModelRepository, logger *zap.Logger) *CustomModelBuilder {
	return &CustomModelBuilder{
		repository: repository,
		logger:     logger,
	}
}

// CreateCustomModel creates a new custom risk model
func (cmb *CustomModelBuilder) CreateCustomModel(ctx context.Context, request *CreateCustomModelRequest) (*CustomRiskModel, error) {
	cmb.logger.Info("Creating custom risk model",
		zap.String("tenant_id", request.TenantID),
		zap.String("name", request.Name),
		zap.String("base_model", string(request.BaseModel)))

	// Validate the request
	if err := cmb.validateCreateRequest(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create the custom model
	model := &CustomRiskModel{
		ID:              generateModelID(),
		TenantID:        request.TenantID,
		Name:            request.Name,
		Description:     request.Description,
		BaseModel:       request.BaseModel,
		CustomFactors:   request.CustomFactors,
		FactorWeights:   request.FactorWeights,
		Thresholds:      request.Thresholds,
		ValidationRules: request.ValidationRules,
		IsActive:        true,
		Version:         "1.0.0",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		CreatedBy:       request.CreatedBy,
		Metadata:        request.Metadata,
	}

	// Validate the model configuration
	if err := cmb.ValidateModelConfiguration(model); err != nil {
		return nil, fmt.Errorf("model configuration validation failed: %w", err)
	}

	cmb.logger.Info("Custom risk model created successfully",
		zap.String("model_id", model.ID),
		zap.String("tenant_id", model.TenantID))

	return model, nil
}

// UpdateCustomModel updates an existing custom risk model
func (cmb *CustomModelBuilder) UpdateCustomModel(ctx context.Context, modelID string, request *UpdateCustomModelRequest) (*CustomRiskModel, error) {
	cmb.logger.Info("Updating custom risk model",
		zap.String("model_id", modelID))

	// Validate the update request
	if err := cmb.validateUpdateRequest(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// This would typically load the existing model from storage
	// For now, we'll create a new model with the updated fields
	model := &CustomRiskModel{
		ID:              modelID,
		TenantID:        request.TenantID,
		Name:            request.Name,
		Description:     request.Description,
		BaseModel:       request.BaseModel,
		CustomFactors:   request.CustomFactors,
		FactorWeights:   request.FactorWeights,
		Thresholds:      request.Thresholds,
		ValidationRules: request.ValidationRules,
		IsActive:        request.IsActive,
		Version:         incrementVersion(request.Version),
		UpdatedAt:       time.Now(),
		CreatedBy:       request.CreatedBy,
		Metadata:        request.Metadata,
	}

	// Validate the updated model configuration
	if err := cmb.ValidateModelConfiguration(model); err != nil {
		return nil, fmt.Errorf("model configuration validation failed: %w", err)
	}

	cmb.logger.Info("Custom risk model updated successfully",
		zap.String("model_id", model.ID))

	return model, nil
}

// ValidateModelConfiguration validates the configuration of a custom model
func (cmb *CustomModelBuilder) ValidateModelConfiguration(model *CustomRiskModel) error {
	// Validate factor weights sum to 1.0
	totalWeight := 0.0
	for _, weight := range model.FactorWeights {
		totalWeight += weight
	}
	if totalWeight < 0.99 || totalWeight > 1.01 {
		return fmt.Errorf("factor weights must sum to 1.0, got %f", totalWeight)
	}

	// Validate thresholds are in ascending order
	thresholds := []float64{
		model.Thresholds[models.RiskLevelLow],
		model.Thresholds[models.RiskLevelMedium],
		model.Thresholds[models.RiskLevelHigh],
		model.Thresholds[models.RiskLevelCritical],
	}

	for i := 1; i < len(thresholds); i++ {
		if thresholds[i] <= thresholds[i-1] {
			return fmt.Errorf("risk level thresholds must be in ascending order")
		}
	}

	// Validate custom factors
	for _, factor := range model.CustomFactors {
		if err := cmb.validateCustomFactor(factor); err != nil {
			return fmt.Errorf("invalid custom factor %s: %w", factor.ID, err)
		}
	}

	// Validate validation rules
	for _, rule := range model.ValidationRules {
		if err := cmb.validateValidationRule(rule); err != nil {
			return fmt.Errorf("invalid validation rule %s: %w", rule.ID, err)
		}
	}

	return nil
}

// validateCreateRequest validates a create custom model request
func (cmb *CustomModelBuilder) validateCreateRequest(request *CreateCustomModelRequest) error {
	if request.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}
	if request.BaseModel == "" {
		return fmt.Errorf("base_model is required")
	}
	if request.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}
	return nil
}

// validateUpdateRequest validates an update custom model request
func (cmb *CustomModelBuilder) validateUpdateRequest(request *UpdateCustomModelRequest) error {
	if request.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}
	if request.BaseModel == "" {
		return fmt.Errorf("base_model is required")
	}
	return nil
}

// validateCustomFactor validates a custom risk factor
func (cmb *CustomModelBuilder) validateCustomFactor(factor CustomRiskFactor) error {
	if factor.ID == "" {
		return fmt.Errorf("factor ID is required")
	}
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

// validateValidationRule validates a validation rule
func (cmb *CustomModelBuilder) validateValidationRule(rule ValidationRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if rule.RuleType == "" {
		return fmt.Errorf("rule type is required")
	}
	return nil
}

// CreateCustomModelRequest represents a request to create a custom model
type CreateCustomModelRequest struct {
	TenantID        string                       `json:"tenant_id"`
	Name            string                       `json:"name"`
	Description     string                       `json:"description"`
	BaseModel       industry.IndustryType        `json:"base_model"`
	CustomFactors   []CustomRiskFactor           `json:"custom_factors"`
	FactorWeights   map[string]float64           `json:"factor_weights"`
	Thresholds      map[models.RiskLevel]float64 `json:"thresholds"`
	ValidationRules []ValidationRule             `json:"validation_rules"`
	CreatedBy       string                       `json:"created_by"`
	Metadata        map[string]interface{}       `json:"metadata"`
}

// UpdateCustomModelRequest represents a request to update a custom model
type UpdateCustomModelRequest struct {
	TenantID        string                       `json:"tenant_id"`
	Name            string                       `json:"name"`
	Description     string                       `json:"description"`
	BaseModel       industry.IndustryType        `json:"base_model"`
	CustomFactors   []CustomRiskFactor           `json:"custom_factors"`
	FactorWeights   map[string]float64           `json:"factor_weights"`
	Thresholds      map[models.RiskLevel]float64 `json:"thresholds"`
	ValidationRules []ValidationRule             `json:"validation_rules"`
	IsActive        bool                         `json:"is_active"`
	Version         string                       `json:"version"`
	CreatedBy       string                       `json:"created_by"`
	Metadata        map[string]interface{}       `json:"metadata"`
}

// CustomModelValidationResult represents the result of model validation
type CustomModelValidationResult struct {
	IsValid            bool                     `json:"is_valid"`
	Errors             []string                 `json:"errors"`
	Warnings           []string                 `json:"warnings"`
	ValidationDate     time.Time                `json:"validation_date"`
	ModelID            string                   `json:"model_id"`
	PerformanceMetrics *ModelPerformanceMetrics `json:"performance_metrics,omitempty"`
}

// ModelPerformanceMetrics represents performance metrics for a custom model
type ModelPerformanceMetrics struct {
	Accuracy        float64 `json:"accuracy"`
	Precision       float64 `json:"precision"`
	Recall          float64 `json:"recall"`
	F1Score         float64 `json:"f1_score"`
	ConfidenceScore float64 `json:"confidence_score"`
	TestDataSize    int     `json:"test_data_size"`
}

// Helper functions

func generateModelID() string {
	return fmt.Sprintf("custom_model_%d", time.Now().UnixNano())
}

func incrementVersion(version string) string {
	// Simple version increment - in production, use proper semantic versioning
	if version == "" {
		return "1.0.0"
	}
	return version + ".1"
}

// GetDefaultThresholds returns default risk level thresholds
func GetDefaultThresholds() map[models.RiskLevel]float64 {
	return map[models.RiskLevel]float64{
		models.RiskLevelLow:      0.25,
		models.RiskLevelMedium:   0.5,
		models.RiskLevelHigh:     0.75,
		models.RiskLevelCritical: 1.0,
	}
}

// GetDefaultFactorWeights returns default factor weights for common risk categories
func GetDefaultFactorWeights() map[string]float64 {
	return map[string]float64{
		"financial":    0.3,
		"operational":  0.25,
		"compliance":   0.2,
		"reputational": 0.15,
		"regulatory":   0.1,
	}
}

// GetDefaultCustomFactors returns default custom factors for a base model
func GetDefaultCustomFactors(baseModel industry.IndustryType) []CustomRiskFactor {
	factors := []CustomRiskFactor{
		{
			ID:              "annual_revenue",
			Name:            "Annual Revenue",
			Description:     "Annual revenue of the business",
			Category:        "financial",
			Weight:          0.3,
			DataType:        "numeric",
			IsRequired:      false,
			DefaultValue:    0.0,
			ScoringFunction: "linear",
			ScoringParams: map[string]interface{}{
				"min_value": 0,
				"max_value": 1000000000,
			},
		},
		{
			ID:              "employee_count",
			Name:            "Employee Count",
			Description:     "Number of employees",
			Category:        "operational",
			Weight:          0.2,
			DataType:        "numeric",
			IsRequired:      false,
			DefaultValue:    0,
			ScoringFunction: "logarithmic",
			ScoringParams: map[string]interface{}{
				"min_value": 0,
				"max_value": 10000,
			},
		},
		{
			ID:              "years_in_business",
			Name:            "Years in Business",
			Description:     "Number of years the business has been operating",
			Category:        "operational",
			Weight:          0.15,
			DataType:        "numeric",
			IsRequired:      false,
			DefaultValue:    0,
			ScoringFunction: "linear",
			ScoringParams: map[string]interface{}{
				"min_value": 0,
				"max_value": 100,
			},
		},
		{
			ID:              "has_website",
			Name:            "Has Website",
			Description:     "Whether the business has a website",
			Category:        "operational",
			Weight:          0.1,
			DataType:        "boolean",
			IsRequired:      false,
			DefaultValue:    false,
			ScoringFunction: "linear",
		},
		{
			ID:              "compliance_score",
			Name:            "Compliance Score",
			Description:     "Overall compliance score",
			Category:        "compliance",
			Weight:          0.25,
			DataType:        "numeric",
			IsRequired:      false,
			DefaultValue:    0.5,
			ScoringFunction: "linear",
			ScoringParams: map[string]interface{}{
				"min_value": 0,
				"max_value": 1,
			},
		},
	}

	// Add industry-specific factors based on base model
	switch baseModel {
	case industry.IndustryFintech:
		factors = append(factors, CustomRiskFactor{
			ID:              "regulatory_licenses",
			Name:            "Regulatory Licenses",
			Description:     "Number of regulatory licenses held",
			Category:        "compliance",
			Weight:          0.2,
			DataType:        "numeric",
			IsRequired:      true,
			DefaultValue:    0,
			ScoringFunction: "linear",
		})
	case industry.IndustryHealthcare:
		factors = append(factors, CustomRiskFactor{
			ID:              "medical_licenses",
			Name:            "Medical Licenses",
			Description:     "Number of medical licenses held",
			Category:        "compliance",
			Weight:          0.3,
			DataType:        "numeric",
			IsRequired:      true,
			DefaultValue:    0,
			ScoringFunction: "linear",
		})
	}

	return factors
}

// GetRepository returns the repository instance
func (cmb *CustomModelBuilder) GetRepository() CustomModelRepository {
	return cmb.repository
}
