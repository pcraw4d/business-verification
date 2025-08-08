package risk

import (
	"fmt"
	"strings"
)

// IndustryModel represents an industry-specific risk model
type IndustryModel struct {
	IndustryCode   string                 `json:"industry_code"`
	IndustryName   string                 `json:"industry_name"`
	RiskFactors    []RiskFactor           `json:"risk_factors"`
	Thresholds     map[RiskLevel]float64  `json:"thresholds"`
	ScoringWeights map[string]float64     `json:"scoring_weights"`
	SpecialFactors map[string]interface{} `json:"special_factors"`
	ModelVersion   string                 `json:"model_version"`
	LastUpdated    string                 `json:"last_updated"`
}

// IndustryModelRegistry manages industry-specific risk models
type IndustryModelRegistry struct {
	models map[string]*IndustryModel
}

// NewIndustryModelRegistry creates a new industry model registry
func NewIndustryModelRegistry() *IndustryModelRegistry {
	return &IndustryModelRegistry{
		models: make(map[string]*IndustryModel),
	}
}

// RegisterModel registers an industry-specific risk model
func (r *IndustryModelRegistry) RegisterModel(model *IndustryModel) error {
	if model.IndustryCode == "" {
		return fmt.Errorf("industry code cannot be empty")
	}

	if len(model.RiskFactors) == 0 {
		return fmt.Errorf("model must have at least one risk factor")
	}

	r.models[model.IndustryCode] = model
	return nil
}

// GetModel retrieves an industry-specific risk model
func (r *IndustryModelRegistry) GetModel(industryCode string) (*IndustryModel, bool) {
	model, exists := r.models[industryCode]
	return model, exists
}

// GetModelByNAICS retrieves a model by NAICS code
func (r *IndustryModelRegistry) GetModelByNAICS(naicsCode string) (*IndustryModel, bool) {
	// Try exact match first
	if model, exists := r.models[naicsCode]; exists {
		return model, true
	}

	// Try partial match (e.g., 5415 for 541511)
	for code, model := range r.models {
		if strings.HasPrefix(naicsCode, code) {
			return model, true
		}
	}

	return nil, false
}

// ListModels returns all registered models
func (r *IndustryModelRegistry) ListModels() []*IndustryModel {
	models := make([]*IndustryModel, 0, len(r.models))
	for _, model := range r.models {
		models = append(models, model)
	}
	return models
}

// IndustrySpecificScoringAlgorithm implements industry-specific risk scoring
type IndustrySpecificScoringAlgorithm struct {
	registry *IndustryModelRegistry
	base     *WeightedScoringAlgorithm
}

// NewIndustrySpecificScoringAlgorithm creates a new industry-specific scoring algorithm
func NewIndustrySpecificScoringAlgorithm(registry *IndustryModelRegistry) *IndustrySpecificScoringAlgorithm {
	return &IndustrySpecificScoringAlgorithm{
		registry: registry,
		base:     NewWeightedScoringAlgorithm(),
	}
}

// CalculateScore calculates risk score using industry-specific models
func (i *IndustrySpecificScoringAlgorithm) CalculateScore(factors []RiskFactor, data map[string]interface{}) (float64, float64, error) {
	// Get industry code from data
	industryCode, exists := data["industry_code"].(string)
	if !exists {
		// Fall back to base algorithm if no industry code
		return i.base.CalculateScore(factors, data)
	}

	// Get industry-specific model
	model, exists := i.registry.GetModelByNAICS(industryCode)
	if !exists {
		// Fall back to base algorithm if no industry model
		return i.base.CalculateScore(factors, data)
	}

	// Use industry-specific factors and weights
	return i.calculateIndustryScore(model, data)
}

// calculateIndustryScore calculates score using industry-specific model
func (i *IndustrySpecificScoringAlgorithm) calculateIndustryScore(model *IndustryModel, data map[string]interface{}) (float64, float64, error) {
	var totalScore float64
	var totalWeight float64
	var confidence float64

	for _, factor := range model.RiskFactors {
		// Get industry-specific weight
		weight := factor.Weight
		if modelWeight, exists := model.ScoringWeights[factor.ID]; exists {
			weight = modelWeight
		}

		// Calculate factor score
		factorScore, factorConfidence, err := i.calculateIndustryFactorScore(factor, model, data)
		if err != nil {
			continue
		}

		// Apply weight
		weightedScore := factorScore * weight
		totalScore += weightedScore
		totalWeight += weight
		confidence += factorConfidence * weight
	}

	if totalWeight == 0 {
		return 0.0, 0.0, nil
	}

	finalScore := totalScore / totalWeight
	finalConfidence := confidence / totalWeight

	return finalScore, finalConfidence, nil
}

// calculateIndustryFactorScore calculates score for a specific factor using industry model
func (i *IndustrySpecificScoringAlgorithm) calculateIndustryFactorScore(factor RiskFactor, model *IndustryModel, data map[string]interface{}) (float64, float64, error) {
	// Get factor-specific data
	factorData, exists := data[factor.ID]
	if !exists {
		return 0.0, 0.0, nil
	}

	// Check for industry-specific special factors
	if specialFactor, exists := model.SpecialFactors[factor.ID]; exists {
		return i.calculateSpecialFactorScore(factor, specialFactor, factorData)
	}

	// Use base algorithm for standard factors
	return i.base.calculateFactorScore(factor, data)
}

// calculateSpecialFactorScore calculates score for industry-specific special factors
func (i *IndustrySpecificScoringAlgorithm) calculateSpecialFactorScore(factor RiskFactor, specialFactor interface{}, data interface{}) (float64, float64, error) {
	// Handle different types of special factors
	switch specialFactor.(type) {
	case map[string]interface{}:
		return i.calculateCustomFactorScore(factor, specialFactor.(map[string]interface{}), data)
	default:
		// Fall back to base calculation
		return i.base.calculateFactorScore(factor, map[string]interface{}{factor.ID: data})
	}
}

// calculateCustomFactorScore calculates score for custom industry factors
func (i *IndustrySpecificScoringAlgorithm) calculateCustomFactorScore(factor RiskFactor, specialConfig map[string]interface{}, data interface{}) (float64, float64, error) {
	// Implementation depends on the specific industry and factor
	// This is a placeholder for industry-specific logic
	return 50.0, 0.8, nil
}

// CalculateLevel determines risk level using industry-specific thresholds
func (i *IndustrySpecificScoringAlgorithm) CalculateLevel(score float64, thresholds map[RiskLevel]float64) RiskLevel {
	if len(thresholds) == 0 {
		// Use default thresholds
		thresholds = map[RiskLevel]float64{
			RiskLevelLow:      25.0,
			RiskLevelMedium:   50.0,
			RiskLevelHigh:     75.0,
			RiskLevelCritical: 90.0,
		}
	}

	return i.base.CalculateLevel(score, thresholds)
}

// CalculateConfidence calculates confidence using industry-specific factors
func (i *IndustrySpecificScoringAlgorithm) CalculateConfidence(factors []RiskFactor, data map[string]interface{}) float64 {
	// Get industry code from data
	industryCode, exists := data["industry_code"].(string)
	if !exists {
		return i.base.CalculateConfidence(factors, data)
	}

	// Get industry-specific model
	model, exists := i.registry.GetModelByNAICS(industryCode)
	if !exists {
		return i.base.CalculateConfidence(factors, data)
	}

	// Calculate confidence based on industry-specific factors
	return i.calculateIndustryConfidence(model, data)
}

// calculateIndustryConfidence calculates confidence for industry-specific model
func (i *IndustrySpecificScoringAlgorithm) calculateIndustryConfidence(model *IndustryModel, data map[string]interface{}) float64 {
	var totalConfidence float64
	var validFactors int

	for _, factor := range model.RiskFactors {
		if _, exists := data[factor.ID]; exists {
			totalConfidence += 0.9 // Higher confidence for industry-specific data
		} else {
			totalConfidence += 0.3 // Lower confidence for missing industry data
		}
		validFactors++
	}

	if validFactors == 0 {
		return 0.0
	}

	return totalConfidence / float64(validFactors)
}

// Predefined industry models
func CreateDefaultIndustryModels() *IndustryModelRegistry {
	registry := NewIndustryModelRegistry()

	// Financial Services Industry Model
	financialModel := &IndustryModel{
		IndustryCode: "52",
		IndustryName: "Finance and Insurance",
		RiskFactors: []RiskFactor{
			{
				ID:         "regulatory_compliance",
				Name:       "Regulatory Compliance",
				Category:   RiskCategoryRegulatory,
				Weight:     0.4,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 20, RiskLevelMedium: 40, RiskLevelHigh: 70, RiskLevelCritical: 85},
			},
			{
				ID:         "financial_stability",
				Name:       "Financial Stability",
				Category:   RiskCategoryFinancial,
				Weight:     0.3,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 25, RiskLevelMedium: 50, RiskLevelHigh: 75, RiskLevelCritical: 90},
			},
			{
				ID:         "cybersecurity",
				Name:       "Cybersecurity",
				Category:   RiskCategoryCybersecurity,
				Weight:     0.3,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 20, RiskLevelMedium: 45, RiskLevelHigh: 80, RiskLevelCritical: 95},
			},
		},
		Thresholds: map[RiskLevel]float64{
			RiskLevelLow:      30.0,
			RiskLevelMedium:   55.0,
			RiskLevelHigh:     80.0,
			RiskLevelCritical: 90.0,
		},
		ScoringWeights: map[string]float64{
			"regulatory_compliance": 0.4,
			"financial_stability":   0.3,
			"cybersecurity":         0.3,
		},
		SpecialFactors: map[string]interface{}{
			"regulatory_compliance": map[string]interface{}{
				"required_licenses": []string{"banking", "insurance", "securities"},
				"compliance_checks": []string{"SOX", "GLBA", "Dodd-Frank"},
			},
		},
		ModelVersion: "1.0",
		LastUpdated:  "2024-01-01",
	}
	registry.RegisterModel(financialModel)

	// Technology Industry Model
	techModel := &IndustryModel{
		IndustryCode: "54",
		IndustryName: "Professional, Scientific, and Technical Services",
		RiskFactors: []RiskFactor{
			{
				ID:         "cybersecurity",
				Name:       "Cybersecurity",
				Category:   RiskCategoryCybersecurity,
				Weight:     0.4,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 15, RiskLevelMedium: 40, RiskLevelHigh: 75, RiskLevelCritical: 90},
			},
			{
				ID:         "operational_efficiency",
				Name:       "Operational Efficiency",
				Category:   RiskCategoryOperational,
				Weight:     0.3,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 25, RiskLevelMedium: 50, RiskLevelHigh: 75, RiskLevelCritical: 90},
			},
			{
				ID:         "financial_stability",
				Name:       "Financial Stability",
				Category:   RiskCategoryFinancial,
				Weight:     0.3,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 20, RiskLevelMedium: 45, RiskLevelHigh: 70, RiskLevelCritical: 85},
			},
		},
		Thresholds: map[RiskLevel]float64{
			RiskLevelLow:      25.0,
			RiskLevelMedium:   50.0,
			RiskLevelHigh:     75.0,
			RiskLevelCritical: 90.0,
		},
		ScoringWeights: map[string]float64{
			"cybersecurity":          0.4,
			"operational_efficiency": 0.3,
			"financial_stability":    0.3,
		},
		SpecialFactors: map[string]interface{}{
			"cybersecurity": map[string]interface{}{
				"data_protection":     []string{"GDPR", "CCPA", "HIPAA"},
				"security_frameworks": []string{"ISO 27001", "SOC 2", "NIST"},
			},
		},
		ModelVersion: "1.0",
		LastUpdated:  "2024-01-01",
	}
	registry.RegisterModel(techModel)

	// Healthcare Industry Model
	healthcareModel := &IndustryModel{
		IndustryCode: "62",
		IndustryName: "Health Care and Social Assistance",
		RiskFactors: []RiskFactor{
			{
				ID:         "regulatory_compliance",
				Name:       "Regulatory Compliance",
				Category:   RiskCategoryRegulatory,
				Weight:     0.4,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 20, RiskLevelMedium: 45, RiskLevelHigh: 80, RiskLevelCritical: 95},
			},
			{
				ID:         "cybersecurity",
				Name:       "Cybersecurity",
				Category:   RiskCategoryCybersecurity,
				Weight:     0.3,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 15, RiskLevelMedium: 40, RiskLevelHigh: 75, RiskLevelCritical: 90},
			},
			{
				ID:         "operational_efficiency",
				Name:       "Operational Efficiency",
				Category:   RiskCategoryOperational,
				Weight:     0.3,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 25, RiskLevelMedium: 50, RiskLevelHigh: 75, RiskLevelCritical: 90},
			},
		},
		Thresholds: map[RiskLevel]float64{
			RiskLevelLow:      30.0,
			RiskLevelMedium:   55.0,
			RiskLevelHigh:     80.0,
			RiskLevelCritical: 95.0,
		},
		ScoringWeights: map[string]float64{
			"regulatory_compliance":  0.4,
			"cybersecurity":          0.3,
			"operational_efficiency": 0.3,
		},
		SpecialFactors: map[string]interface{}{
			"regulatory_compliance": map[string]interface{}{
				"healthcare_regulations": []string{"HIPAA", "HITECH", "FDA"},
				"licensing_requirements": []string{"medical_license", "facility_license"},
			},
		},
		ModelVersion: "1.0",
		LastUpdated:  "2024-01-01",
	}
	registry.RegisterModel(healthcareModel)

	// Manufacturing Industry Model
	manufacturingModel := &IndustryModel{
		IndustryCode: "31",
		IndustryName: "Manufacturing",
		RiskFactors: []RiskFactor{
			{
				ID:         "operational_efficiency",
				Name:       "Operational Efficiency",
				Category:   RiskCategoryOperational,
				Weight:     0.4,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 20, RiskLevelMedium: 45, RiskLevelHigh: 75, RiskLevelCritical: 90},
			},
			{
				ID:         "financial_stability",
				Name:       "Financial Stability",
				Category:   RiskCategoryFinancial,
				Weight:     0.3,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 25, RiskLevelMedium: 50, RiskLevelHigh: 75, RiskLevelCritical: 90},
			},
			{
				ID:         "regulatory_compliance",
				Name:       "Regulatory Compliance",
				Category:   RiskCategoryRegulatory,
				Weight:     0.3,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 20, RiskLevelMedium: 45, RiskLevelHigh: 75, RiskLevelCritical: 90},
			},
		},
		Thresholds: map[RiskLevel]float64{
			RiskLevelLow:      25.0,
			RiskLevelMedium:   50.0,
			RiskLevelHigh:     75.0,
			RiskLevelCritical: 90.0,
		},
		ScoringWeights: map[string]float64{
			"operational_efficiency": 0.4,
			"financial_stability":    0.3,
			"regulatory_compliance":  0.3,
		},
		SpecialFactors: map[string]interface{}{
			"operational_efficiency": map[string]interface{}{
				"quality_standards": []string{"ISO 9001", "Six Sigma", "Lean"},
				"safety_standards":  []string{"OSHA", "ISO 45001"},
			},
		},
		ModelVersion: "1.0",
		LastUpdated:  "2024-01-01",
	}
	registry.RegisterModel(manufacturingModel)

	// Retail Industry Model
	retailModel := &IndustryModel{
		IndustryCode: "44",
		IndustryName: "Retail Trade",
		RiskFactors: []RiskFactor{
			{
				ID:         "financial_stability",
				Name:       "Financial Stability",
				Category:   RiskCategoryFinancial,
				Weight:     0.4,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 20, RiskLevelMedium: 45, RiskLevelHigh: 75, RiskLevelCritical: 90},
			},
			{
				ID:         "operational_efficiency",
				Name:       "Operational Efficiency",
				Category:   RiskCategoryOperational,
				Weight:     0.3,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 25, RiskLevelMedium: 50, RiskLevelHigh: 75, RiskLevelCritical: 90},
			},
			{
				ID:         "reputational_risk",
				Name:       "Reputational Risk",
				Category:   RiskCategoryReputational,
				Weight:     0.3,
				Thresholds: map[RiskLevel]float64{RiskLevelLow: 20, RiskLevelMedium: 45, RiskLevelHigh: 75, RiskLevelCritical: 90},
			},
		},
		Thresholds: map[RiskLevel]float64{
			RiskLevelLow:      25.0,
			RiskLevelMedium:   50.0,
			RiskLevelHigh:     75.0,
			RiskLevelCritical: 90.0,
		},
		ScoringWeights: map[string]float64{
			"financial_stability":    0.4,
			"operational_efficiency": 0.3,
			"reputational_risk":      0.3,
		},
		SpecialFactors: map[string]interface{}{
			"operational_efficiency": map[string]interface{}{
				"inventory_management": []string{"JIT", "ABC_analysis", "cycle_counting"},
				"customer_service":     []string{"NPS", "customer_satisfaction", "return_rate"},
			},
		},
		ModelVersion: "1.0",
		LastUpdated:  "2024-01-01",
	}
	registry.RegisterModel(retailModel)

	return registry
}

// IndustryRiskAssessment represents an industry-specific risk assessment
type IndustryRiskAssessment struct {
	BusinessID      string                 `json:"business_id"`
	IndustryCode    string                 `json:"industry_code"`
	IndustryName    string                 `json:"industry_name"`
	ModelVersion    string                 `json:"model_version"`
	OverallScore    float64                `json:"overall_score"`
	OverallLevel    RiskLevel              `json:"overall_level"`
	FactorScores    []RiskScore            `json:"factor_scores"`
	IndustryFactors []IndustryRiskFactor   `json:"industry_factors"`
	Recommendations []RiskRecommendation   `json:"recommendations"`
	AssessedAt      string                 `json:"assessed_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// IndustryRiskFactor represents an industry-specific risk factor
type IndustryRiskFactor struct {
	FactorID       string                 `json:"factor_id"`
	FactorName     string                 `json:"factor_name"`
	Category       RiskCategory           `json:"category"`
	Score          float64                `json:"score"`
	Level          RiskLevel              `json:"level"`
	Confidence     float64                `json:"confidence"`
	IndustryWeight float64                `json:"industry_weight"`
	SpecialFactors map[string]interface{} `json:"special_factors,omitempty"`
	Explanation    string                 `json:"explanation"`
	Evidence       []string               `json:"evidence"`
}
