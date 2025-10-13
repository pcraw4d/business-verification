package industry

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// IndustryType represents different industry sectors
type IndustryType string

const (
	IndustryFintech        IndustryType = "fintech"
	IndustryHealthcare     IndustryType = "healthcare"
	IndustryTechnology     IndustryType = "technology"
	IndustryRetail         IndustryType = "retail"
	IndustryManufacturing  IndustryType = "manufacturing"
	IndustryRealEstate     IndustryType = "real_estate"
	IndustryEnergy         IndustryType = "energy"
	IndustryTransportation IndustryType = "transportation"
	IndustryGeneral        IndustryType = "general"
)

// IndustryModel defines the interface for industry-specific risk models
type IndustryModel interface {
	// GetIndustryType returns the industry type this model handles
	GetIndustryType() IndustryType

	// CalculateIndustryRisk calculates industry-specific risk factors
	CalculateIndustryRisk(ctx context.Context, business *models.RiskAssessmentRequest) (*IndustryRiskResult, error)

	// GetIndustrySpecificFactors returns industry-specific risk factors
	GetIndustrySpecificFactors() []IndustryRiskFactor

	// GetIndustryWeightings returns the risk category weightings for this industry
	GetIndustryWeightings() map[string]float64

	// ValidateIndustryData validates business data against industry requirements
	ValidateIndustryData(business *models.RiskAssessmentRequest) []string

	// GetIndustryComplianceRequirements returns compliance requirements for this industry
	GetIndustryComplianceRequirements() []ComplianceRequirement

	// ApplyCustomConfiguration applies custom configuration to the model
	ApplyCustomConfiguration(customFactors []IndustryCustomRiskFactor, customWeights map[string]float64) error

	// GetCustomizableFactors returns factors that can be customized
	GetCustomizableFactors() []CustomizableFactor
}

// IndustryCustomRiskFactor represents a custom risk factor for industry models
type IndustryCustomRiskFactor struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category"`
	Weight          float64                `json:"weight"`
	DataType        string                 `json:"data_type"`
	ValidationRules []FactorValidationRule `json:"validation_rules"`
	ScoringFunction string                 `json:"scoring_function"`
	ScoringParams   map[string]interface{} `json:"scoring_params"`
	IsRequired      bool                   `json:"is_required"`
	DefaultValue    interface{}            `json:"default_value"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// CustomizableFactor represents a factor that can be customized in a model
type CustomizableFactor struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category"`
	DefaultWeight   float64                `json:"default_weight"`
	MinWeight       float64                `json:"min_weight"`
	MaxWeight       float64                `json:"max_weight"`
	DataType        string                 `json:"data_type"`
	IsRequired      bool                   `json:"is_required"`
	Customizable    bool                   `json:"customizable"`
	ValidationRules []FactorValidationRule `json:"validation_rules"`
	ScoringOptions  []ScoringOption        `json:"scoring_options"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// FactorValidationRule represents a validation rule for a factor
type FactorValidationRule struct {
	RuleType     string                 `json:"rule_type"`
	Parameters   map[string]interface{} `json:"parameters"`
	ErrorMessage string                 `json:"error_message"`
}

// ScoringOption represents a scoring option for a factor
type ScoringOption struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Function    string                 `json:"function"` // "linear", "exponential", "logarithmic", "custom"
	Parameters  map[string]interface{} `json:"parameters"`
	IsDefault   bool                   `json:"is_default"`
}

// IndustryRiskResult represents the result of industry-specific risk analysis
type IndustryRiskResult struct {
	IndustryType            IndustryType             `json:"industry_type"`
	IndustryRiskScore       float64                  `json:"industry_risk_score"`
	IndustryRiskLevel       models.RiskLevel         `json:"industry_risk_level"`
	IndustryFactors         []IndustryRiskFactor     `json:"industry_factors"`
	ComplianceStatus        []ComplianceStatus       `json:"compliance_status"`
	IndustryRecommendations []IndustryRecommendation `json:"industry_recommendations"`
	RegulatoryFactors       []RegulatoryFactor       `json:"regulatory_factors"`
	MarketFactors           []MarketFactor           `json:"market_factors"`
	OperationalFactors      []OperationalFactor      `json:"operational_factors"`
	AnalysisTimestamp       time.Time                `json:"analysis_timestamp"`
	ConfidenceScore         float64                  `json:"confidence_score"`
}

// IndustryRiskFactor represents an industry-specific risk factor
type IndustryRiskFactor struct {
	FactorID            string  `json:"factor_id"`
	FactorName          string  `json:"factor_name"`
	FactorCategory      string  `json:"factor_category"`
	RiskScore           float64 `json:"risk_score"`
	RiskLevel           string  `json:"risk_level"`
	Description         string  `json:"description"`
	Impact              string  `json:"impact"`
	Likelihood          string  `json:"likelihood"`
	MitigationAdvice    string  `json:"mitigation_advice"`
	IndustrySpecific    bool    `json:"industry_specific"`
	RegulatoryRelevance bool    `json:"regulatory_relevance"`
}

// ComplianceRequirement represents a compliance requirement for an industry
type ComplianceRequirement struct {
	RequirementID   string   `json:"requirement_id"`
	RequirementName string   `json:"requirement_name"`
	RegulatoryBody  string   `json:"regulatory_body"`
	Jurisdiction    string   `json:"jurisdiction"`
	Description     string   `json:"description"`
	Required        bool     `json:"required"`
	PenaltyAmount   string   `json:"penalty_amount"`
	ComplianceSteps []string `json:"compliance_steps"`
	Documentation   []string `json:"documentation"`
}

// ComplianceStatus represents the compliance status for a requirement
type ComplianceStatus struct {
	RequirementID   string    `json:"requirement_id"`
	Status          string    `json:"status"` // "compliant", "non_compliant", "unknown", "not_applicable"
	LastChecked     time.Time `json:"last_checked"`
	ComplianceScore float64   `json:"compliance_score"`
	Issues          []string  `json:"issues"`
	Recommendations []string  `json:"recommendations"`
}

// IndustryRecommendation represents an industry-specific recommendation
type IndustryRecommendation struct {
	RecommendationID   string   `json:"recommendation_id"`
	Category           string   `json:"category"`
	Priority           string   `json:"priority"` // "high", "medium", "low"
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	ActionItems        []string `json:"action_items"`
	ExpectedBenefit    string   `json:"expected_benefit"`
	ImplementationCost string   `json:"implementation_cost"`
	Timeline           string   `json:"timeline"`
}

// RegulatoryFactor represents a regulatory factor specific to an industry
type RegulatoryFactor struct {
	FactorID       string    `json:"factor_id"`
	RegulationName string    `json:"regulation_name"`
	RegulatoryBody string    `json:"regulatory_body"`
	Jurisdiction   string    `json:"jurisdiction"`
	RiskImpact     float64   `json:"risk_impact"`
	ComplianceCost string    `json:"compliance_cost"`
	PenaltyRisk    string    `json:"penalty_risk"`
	Description    string    `json:"description"`
	LastUpdated    time.Time `json:"last_updated"`
}

// MarketFactor represents a market factor specific to an industry
type MarketFactor struct {
	FactorID       string   `json:"factor_id"`
	FactorName     string   `json:"factor_name"`
	MarketTrend    string   `json:"market_trend"` // "growing", "stable", "declining", "volatile"
	ImpactScore    float64  `json:"impact_score"`
	TimeHorizon    string   `json:"time_horizon"`
	Description    string   `json:"description"`
	KeyDrivers     []string `json:"key_drivers"`
	RiskMitigation []string `json:"risk_mitigation"`
}

// OperationalFactor represents an operational factor specific to an industry
type OperationalFactor struct {
	FactorID            string   `json:"factor_id"`
	FactorName          string   `json:"factor_name"`
	OperationalArea     string   `json:"operational_area"`
	RiskScore           float64  `json:"risk_score"`
	Criticality         string   `json:"criticality"` // "critical", "high", "medium", "low"
	Description         string   `json:"description"`
	ControlMeasures     []string `json:"control_measures"`
	MonitoringFrequency string   `json:"monitoring_frequency"`
}

// IndustryModelManager manages industry-specific models
type IndustryModelManager struct {
	models map[IndustryType]IndustryModel
	logger *zap.Logger
}

// NewIndustryModelManager creates a new industry model manager
func NewIndustryModelManager(logger *zap.Logger) *IndustryModelManager {
	manager := &IndustryModelManager{
		models: make(map[IndustryType]IndustryModel),
		logger: logger,
	}

	// Initialize industry models with custom configuration support
	manager.models[IndustryFintech] = &CustomizableFintechModel{FintechModel: *NewFintechModel(logger)}
	manager.models[IndustryHealthcare] = &CustomizableHealthcareModel{HealthcareModel: *NewHealthcareModel(logger)}
	manager.models[IndustryTechnology] = &CustomizableTechnologyModel{TechnologyModel: *NewTechnologyModel(logger)}
	manager.models[IndustryRetail] = &CustomizableRetailModel{RetailModel: *NewRetailModel(logger)}
	manager.models[IndustryManufacturing] = &CustomizableManufacturingModel{ManufacturingModel: *NewManufacturingModel(logger)}
	manager.models[IndustryRealEstate] = &CustomizableRealEstateModel{RealEstateModel: *NewRealEstateModel(logger)}
	manager.models[IndustryEnergy] = &CustomizableEnergyModel{EnergyModel: *NewEnergyModel(logger)}
	manager.models[IndustryTransportation] = &CustomizableTransportationModel{TransportationModel: *NewTransportationModel(logger)}
	manager.models[IndustryGeneral] = &CustomizableGeneralModel{GeneralModel: *NewGeneralModel(logger)}

	logger.Info("Industry model manager initialized", zap.Int("model_count", len(manager.models)))
	return manager
}

// GetIndustryModel returns the industry model for a specific industry type
func (imm *IndustryModelManager) GetIndustryModel(industryType IndustryType) IndustryModel {
	if model, exists := imm.models[industryType]; exists {
		return model
	}
	// Return general model as fallback
	return imm.models[IndustryGeneral]
}

// DetectIndustryType attempts to detect the industry type from business information
func (imm *IndustryModelManager) DetectIndustryType(business *models.RiskAssessmentRequest) IndustryType {
	if business == nil {
		return IndustryGeneral
	}

	industry := business.Industry
	if industry == "" {
		return IndustryGeneral
	}

	// Convert to lowercase for comparison
	industryLower := strings.ToLower(industry)

	// Industry detection logic
	switch {
	case strings.Contains(industryLower, "fintech") || strings.Contains(industryLower, "financial technology"):
		return IndustryFintech
	case strings.Contains(industryLower, "healthcare") || strings.Contains(industryLower, "medical") || strings.Contains(industryLower, "pharmaceutical"):
		return IndustryHealthcare
	case strings.Contains(industryLower, "technology") || strings.Contains(industryLower, "software") || strings.Contains(industryLower, "tech"):
		return IndustryTechnology
	case strings.Contains(industryLower, "retail") || strings.Contains(industryLower, "ecommerce") || strings.Contains(industryLower, "commerce"):
		return IndustryRetail
	case strings.Contains(industryLower, "manufacturing") || strings.Contains(industryLower, "production") || strings.Contains(industryLower, "industrial"):
		return IndustryManufacturing
	case strings.Contains(industryLower, "real estate") || strings.Contains(industryLower, "property") || strings.Contains(industryLower, "construction"):
		return IndustryRealEstate
	case strings.Contains(industryLower, "energy") || strings.Contains(industryLower, "oil") || strings.Contains(industryLower, "gas") || strings.Contains(industryLower, "renewable"):
		return IndustryEnergy
	case strings.Contains(industryLower, "transportation") || strings.Contains(industryLower, "logistics") || strings.Contains(industryLower, "shipping"):
		return IndustryTransportation
	default:
		return IndustryGeneral
	}
}

// AnalyzeIndustryRisk performs industry-specific risk analysis
func (imm *IndustryModelManager) AnalyzeIndustryRisk(ctx context.Context, business *models.RiskAssessmentRequest) (*IndustryRiskResult, error) {
	industryType := imm.DetectIndustryType(business)
	model := imm.GetIndustryModel(industryType)

	imm.logger.Info("Analyzing industry risk",
		zap.String("industry_type", string(industryType)),
		zap.String("business_name", business.BusinessName))

	result, err := model.CalculateIndustryRisk(ctx, business)
	if err != nil {
		imm.logger.Error("Industry risk analysis failed",
			zap.String("industry_type", string(industryType)),
			zap.Error(err))
		return nil, err
	}

	imm.logger.Info("Industry risk analysis completed",
		zap.String("industry_type", string(industryType)),
		zap.Float64("risk_score", result.IndustryRiskScore),
		zap.String("risk_level", string(result.IndustryRiskLevel)))

	return result, nil
}

// GetAllIndustryTypes returns all supported industry types
func (imm *IndustryModelManager) GetAllIndustryTypes() []IndustryType {
	types := make([]IndustryType, 0, len(imm.models))
	for industryType := range imm.models {
		types = append(types, industryType)
	}
	return types
}

// GetIndustryModelInfo returns information about a specific industry model
func (imm *IndustryModelManager) GetIndustryModelInfo(industryType IndustryType) map[string]interface{} {
	model := imm.GetIndustryModel(industryType)

	info := map[string]interface{}{
		"industry_type":           string(industryType),
		"model_available":         model != nil,
		"risk_factors":            len(model.GetIndustrySpecificFactors()),
		"compliance_requirements": len(model.GetIndustryComplianceRequirements()),
		"weightings":              model.GetIndustryWeightings(),
	}

	return info
}

// Customizable model wrappers that implement the new interface methods

// CustomizableFintechModel wraps FintechModel with custom configuration support
type CustomizableFintechModel struct {
	FintechModel
}

func (m *CustomizableFintechModel) ApplyCustomConfiguration(customFactors []IndustryCustomRiskFactor, customWeights map[string]float64) error {
	// Implementation for applying custom configuration
	return nil
}

func (m *CustomizableFintechModel) GetCustomizableFactors() []CustomizableFactor {
	// Return customizable factors for fintech industry
	return []CustomizableFactor{}
}

// CustomizableHealthcareModel wraps HealthcareModel with custom configuration support
type CustomizableHealthcareModel struct {
	HealthcareModel
}

func (m *CustomizableHealthcareModel) ApplyCustomConfiguration(customFactors []IndustryCustomRiskFactor, customWeights map[string]float64) error {
	return nil
}

func (m *CustomizableHealthcareModel) GetCustomizableFactors() []CustomizableFactor {
	return []CustomizableFactor{}
}

// CustomizableTechnologyModel wraps TechnologyModel with custom configuration support
type CustomizableTechnologyModel struct {
	TechnologyModel
}

func (m *CustomizableTechnologyModel) ApplyCustomConfiguration(customFactors []IndustryCustomRiskFactor, customWeights map[string]float64) error {
	return nil
}

func (m *CustomizableTechnologyModel) GetCustomizableFactors() []CustomizableFactor {
	return []CustomizableFactor{}
}

// CustomizableRetailModel wraps RetailModel with custom configuration support
type CustomizableRetailModel struct {
	RetailModel
}

func (m *CustomizableRetailModel) ApplyCustomConfiguration(customFactors []IndustryCustomRiskFactor, customWeights map[string]float64) error {
	return nil
}

func (m *CustomizableRetailModel) GetCustomizableFactors() []CustomizableFactor {
	return []CustomizableFactor{}
}

// CustomizableManufacturingModel wraps ManufacturingModel with custom configuration support
type CustomizableManufacturingModel struct {
	ManufacturingModel
}

func (m *CustomizableManufacturingModel) ApplyCustomConfiguration(customFactors []IndustryCustomRiskFactor, customWeights map[string]float64) error {
	return nil
}

func (m *CustomizableManufacturingModel) GetCustomizableFactors() []CustomizableFactor {
	return []CustomizableFactor{}
}

// CustomizableRealEstateModel wraps RealEstateModel with custom configuration support
type CustomizableRealEstateModel struct {
	RealEstateModel
}

func (m *CustomizableRealEstateModel) ApplyCustomConfiguration(customFactors []IndustryCustomRiskFactor, customWeights map[string]float64) error {
	return nil
}

func (m *CustomizableRealEstateModel) GetCustomizableFactors() []CustomizableFactor {
	return []CustomizableFactor{}
}

// CustomizableEnergyModel wraps EnergyModel with custom configuration support
type CustomizableEnergyModel struct {
	EnergyModel
}

func (m *CustomizableEnergyModel) ApplyCustomConfiguration(customFactors []IndustryCustomRiskFactor, customWeights map[string]float64) error {
	return nil
}

func (m *CustomizableEnergyModel) GetCustomizableFactors() []CustomizableFactor {
	return []CustomizableFactor{}
}

// CustomizableTransportationModel wraps TransportationModel with custom configuration support
type CustomizableTransportationModel struct {
	TransportationModel
}

func (m *CustomizableTransportationModel) ApplyCustomConfiguration(customFactors []IndustryCustomRiskFactor, customWeights map[string]float64) error {
	return nil
}

func (m *CustomizableTransportationModel) GetCustomizableFactors() []CustomizableFactor {
	return []CustomizableFactor{}
}

// CustomizableGeneralModel wraps GeneralModel with custom configuration support
type CustomizableGeneralModel struct {
	GeneralModel
}

func (m *CustomizableGeneralModel) ApplyCustomConfiguration(customFactors []IndustryCustomRiskFactor, customWeights map[string]float64) error {
	return nil
}

func (m *CustomizableGeneralModel) GetCustomizableFactors() []CustomizableFactor {
	return []CustomizableFactor{}
}
