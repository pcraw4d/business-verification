package industry

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// GeneralModel implements industry-specific risk analysis for general businesses
type GeneralModel struct {
	logger *zap.Logger
}

// NewGeneralModel creates a new general industry model
func NewGeneralModel(logger *zap.Logger) *GeneralModel {
	return &GeneralModel{
		logger: logger,
	}
}

// GetIndustryType returns the industry type
func (gm *GeneralModel) GetIndustryType() IndustryType {
	return IndustryGeneral
}

// CalculateIndustryRisk calculates general business risk factors
func (gm *GeneralModel) CalculateIndustryRisk(ctx context.Context, business *models.RiskAssessmentRequest) (*IndustryRiskResult, error) {
	gm.logger.Info("Calculating general industry risk", zap.String("business", business.BusinessName))

	// Calculate base industry risk score
	baseScore := gm.calculateBaseGeneralRisk(business)

	// Generate industry-specific factors
	industryFactors := gm.generateGeneralRiskFactors(business, baseScore)

	// Calculate compliance status
	complianceStatus := gm.assessGeneralCompliance(business)

	// Generate recommendations
	recommendations := gm.generateGeneralRecommendations(business, baseScore)

	// Generate regulatory factors
	regulatoryFactors := gm.generateGeneralRegulatoryFactors()

	// Generate market factors
	marketFactors := gm.generateGeneralMarketFactors()

	// Generate operational factors
	operationalFactors := gm.generateGeneralOperationalFactors(business)

	// Calculate overall industry risk score
	industryRiskScore := gm.calculateOverallIndustryRisk(baseScore, industryFactors, complianceStatus)

	// Determine risk level
	riskLevel := gm.determineRiskLevel(industryRiskScore)

	// Calculate confidence score
	confidenceScore := gm.calculateConfidenceScore(business, industryFactors)

	result := &IndustryRiskResult{
		IndustryType:            IndustryGeneral,
		IndustryRiskScore:       industryRiskScore,
		IndustryRiskLevel:       riskLevel,
		IndustryFactors:         industryFactors,
		ComplianceStatus:        complianceStatus,
		IndustryRecommendations: recommendations,
		RegulatoryFactors:       regulatoryFactors,
		MarketFactors:           marketFactors,
		OperationalFactors:      operationalFactors,
		AnalysisTimestamp:       time.Now(),
		ConfidenceScore:         confidenceScore,
	}

	gm.logger.Info("General industry risk calculated",
		zap.Float64("risk_score", industryRiskScore),
		zap.String("risk_level", string(riskLevel)))

	return result, nil
}

// GetIndustrySpecificFactors returns general business risk factors
func (gm *GeneralModel) GetIndustrySpecificFactors() []IndustryRiskFactor {
	return []IndustryRiskFactor{
		{
			FactorID:            "general_business_continuity",
			FactorName:          "Business Continuity",
			FactorCategory:      "operational",
			Description:         "Risk from business continuity and operational disruptions",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_financial_management",
			FactorName:          "Financial Management",
			FactorCategory:      "financial",
			Description:         "Risk from financial management and cash flow",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_market_competition",
			FactorName:          "Market Competition",
			FactorCategory:      "market",
			Description:         "Risk from market competition and competitive pressure",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			Description:         "Risk from general regulatory compliance requirements",
			IndustrySpecific:    false,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "general_technology_adoption",
			FactorName:          "Technology Adoption",
			FactorCategory:      "technology",
			Description:         "Risk from technology adoption and digital transformation",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_workforce_management",
			FactorName:          "Workforce Management",
			FactorCategory:      "operational",
			Description:         "Risk from workforce management and human resources",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_supply_chain",
			FactorName:          "Supply Chain Risk",
			FactorCategory:      "operational",
			Description:         "Risk from supply chain disruptions and dependencies",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_reputational_risk",
			FactorName:          "Reputational Risk",
			FactorCategory:      "reputational",
			Description:         "Risk from reputational damage and brand management",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
	}
}

// GetIndustryWeightings returns risk category weightings for general businesses
func (gm *GeneralModel) GetIndustryWeightings() map[string]float64 {
	return map[string]float64{
		"regulatory":    0.10, // Low regulatory risk
		"compliance":    0.10, // Low compliance requirements
		"operational":   0.25, // High operational risk
		"financial":     0.20, // High financial risk
		"reputational":  0.15, // Moderate reputational risk
		"technology":    0.10, // Moderate technology risk
		"geopolitical":  0.05, // Low geopolitical risk
		"environmental": 0.05, // Low environmental risk
	}
}

// ValidateIndustryData validates general business data
func (gm *GeneralModel) ValidateIndustryData(business *models.RiskAssessmentRequest) []string {
	var errors []string

	if business == nil {
		errors = append(errors, "business information is required")
		return errors
	}

	// Check for required general business fields
	if business.BusinessName == "" {
		errors = append(errors, "business name is required")
	}

	return errors
}

// GetIndustryComplianceRequirements returns general business compliance requirements
func (gm *GeneralModel) GetIndustryComplianceRequirements() []ComplianceRequirement {
	return []ComplianceRequirement{
		{
			RequirementID:   "general_business_license",
			RequirementName: "Business License",
			RegulatoryBody:  "Local Government",
			Jurisdiction:    "Local",
			Description:     "General business licensing requirements",
			Required:        true,
			PenaltyAmount:   "Fines and business closure",
			ComplianceSteps: []string{
				"Obtain business license",
				"Renew license annually",
				"Comply with local regulations",
				"Maintain business records",
			},
			Documentation: []string{
				"Business license",
				"Renewal applications",
				"Business records",
			},
		},
		{
			RequirementID:   "general_tax_compliance",
			RequirementName: "Tax Compliance",
			RegulatoryBody:  "Internal Revenue Service",
			Jurisdiction:    "US",
			Description:     "Tax compliance and reporting requirements",
			Required:        true,
			PenaltyAmount:   "Penalties and interest",
			ComplianceSteps: []string{
				"File tax returns",
				"Pay taxes on time",
				"Maintain tax records",
				"Comply with tax regulations",
			},
			Documentation: []string{
				"Tax returns",
				"Tax payments",
				"Tax records",
			},
		},
		{
			RequirementID:   "general_employment_law",
			RequirementName: "Employment Law",
			RegulatoryBody:  "Department of Labor",
			Jurisdiction:    "US",
			Description:     "Employment and labor law compliance",
			Required:        true,
			PenaltyAmount:   "Back wages and penalties",
			ComplianceSteps: []string{
				"Comply with wage laws",
				"Maintain employment records",
				"Provide safe working conditions",
				"Implement anti-discrimination policies",
			},
			Documentation: []string{
				"Employment policies",
				"Wage records",
				"Safety records",
			},
		},
	}
}

// calculateBaseGeneralRisk calculates the base risk score for general businesses
func (gm *GeneralModel) calculateBaseGeneralRisk(business *models.RiskAssessmentRequest) float64 {
	baseScore := 0.3 // Base general business risk is moderate

	// Adjust based on business characteristics
	if business.Metadata != nil {
		// Check for business size
		if size, exists := business.Metadata["business_size"]; exists {
			if s, ok := size.(string); ok {
				switch s {
				case "small", "startup":
					baseScore += 0.1
				case "medium":
					baseScore += 0.05
				case "large", "enterprise":
					baseScore -= 0.05
				}
			}
		}

		// Check for business age
		if age, exists := business.Metadata["business_age"]; exists {
			if a, ok := age.(string); ok {
				switch a {
				case "startup", "new":
					baseScore += 0.1
				case "established", "mature":
					baseScore -= 0.05
				}
			}
		}
	}

	// Ensure score is within bounds
	return math.Max(0.0, math.Min(1.0, baseScore))
}

// generateGeneralRiskFactors generates general business risk factors
func (gm *GeneralModel) generateGeneralRiskFactors(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRiskFactor {
	factors := []IndustryRiskFactor{
		{
			FactorID:            "general_business_continuity",
			FactorName:          "Business Continuity",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.10,
			RiskLevel:           gm.calculateRiskLevel(baseScore + 0.10),
			Description:         "Risk from business continuity and operational disruptions",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Develop business continuity plans, implement backup systems, maintain emergency procedures",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_financial_management",
			FactorName:          "Financial Management",
			FactorCategory:      "financial",
			RiskScore:           baseScore + 0.12,
			RiskLevel:           gm.calculateRiskLevel(baseScore + 0.12),
			Description:         "Risk from financial management and cash flow",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement financial controls, maintain cash reserves, monitor cash flow",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_market_competition",
			FactorName:          "Market Competition",
			FactorCategory:      "market",
			RiskScore:           baseScore + 0.08,
			RiskLevel:           gm.calculateRiskLevel(baseScore + 0.08),
			Description:         "Risk from market competition and competitive pressure",
			Impact:              "medium",
			Likelihood:          "high",
			MitigationAdvice:    "Focus on differentiation, improve customer experience, monitor competitors",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			RiskScore:           baseScore + 0.06,
			RiskLevel:           gm.calculateRiskLevel(baseScore + 0.06),
			Description:         "Risk from general regulatory compliance requirements",
			Impact:              "medium",
			Likelihood:          "low",
			MitigationAdvice:    "Implement compliance programs, conduct regular audits, maintain records",
			IndustrySpecific:    false,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "general_technology_adoption",
			FactorName:          "Technology Adoption",
			FactorCategory:      "technology",
			RiskScore:           baseScore + 0.07,
			RiskLevel:           gm.calculateRiskLevel(baseScore + 0.07),
			Description:         "Risk from technology adoption and digital transformation",
			Impact:              "medium",
			Likelihood:          "medium",
			MitigationAdvice:    "Invest in technology, train staff, implement digital solutions",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_workforce_management",
			FactorName:          "Workforce Management",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.09,
			RiskLevel:           gm.calculateRiskLevel(baseScore + 0.09),
			Description:         "Risk from workforce management and human resources",
			Impact:              "medium",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement HR policies, provide training, maintain employee relations",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_supply_chain",
			FactorName:          "Supply Chain Risk",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.11,
			RiskLevel:           gm.calculateRiskLevel(baseScore + 0.11),
			Description:         "Risk from supply chain disruptions and dependencies",
			Impact:              "medium",
			Likelihood:          "medium",
			MitigationAdvice:    "Diversify suppliers, implement supply chain monitoring, maintain safety stock",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "general_reputational_risk",
			FactorName:          "Reputational Risk",
			FactorCategory:      "reputational",
			RiskScore:           baseScore + 0.05,
			RiskLevel:           gm.calculateRiskLevel(baseScore + 0.05),
			Description:         "Risk from reputational damage and brand management",
			Impact:              "high",
			Likelihood:          "low",
			MitigationAdvice:    "Implement brand management, monitor reputation, maintain customer relations",
			IndustrySpecific:    false,
			RegulatoryRelevance: false,
		},
	}

	return factors
}

// assessGeneralCompliance assesses compliance status for general business requirements
func (gm *GeneralModel) assessGeneralCompliance(business *models.RiskAssessmentRequest) []ComplianceStatus {
	statuses := []ComplianceStatus{
		{
			RequirementID:   "general_business_license",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"Business license compliance not verified"},
			Recommendations: []string{"Verify business license status"},
		},
		{
			RequirementID:   "general_tax_compliance",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"Tax compliance not verified"},
			Recommendations: []string{"Implement tax compliance program"},
		},
		{
			RequirementID:   "general_employment_law",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"Employment law compliance not verified"},
			Recommendations: []string{"Implement employment law compliance program"},
		},
	}

	return statuses
}

// generateGeneralRecommendations generates general business recommendations
func (gm *GeneralModel) generateGeneralRecommendations(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRecommendation {
	recommendations := []IndustryRecommendation{
		{
			RecommendationID:   "general_business_continuity",
			Category:           "operational",
			Priority:           "high",
			Title:              "Business Continuity Planning",
			Description:        "Implement comprehensive business continuity planning",
			ActionItems:        []string{"Develop business continuity plans", "Implement backup systems", "Maintain emergency procedures"},
			ExpectedBenefit:    "Improved business resilience and continuity",
			ImplementationCost: "Medium",
			Timeline:           "6-12 months",
		},
		{
			RecommendationID:   "general_financial_management",
			Category:           "financial",
			Priority:           "high",
			Title:              "Financial Management",
			Description:        "Improve financial management and controls",
			ActionItems:        []string{"Implement financial controls", "Maintain cash reserves", "Monitor cash flow"},
			ExpectedBenefit:    "Improved financial stability and management",
			ImplementationCost: "Low",
			Timeline:           "3-6 months",
		},
		{
			RecommendationID:   "general_technology_upgrade",
			Category:           "technology",
			Priority:           "medium",
			Title:              "Technology Upgrade",
			Description:        "Upgrade technology and digital capabilities",
			ActionItems:        []string{"Invest in technology", "Train staff", "Implement digital solutions"},
			ExpectedBenefit:    "Improved efficiency and competitiveness",
			ImplementationCost: "Medium",
			Timeline:           "6-12 months",
		},
		{
			RecommendationID:   "general_compliance_program",
			Category:           "compliance",
			Priority:           "medium",
			Title:              "Compliance Program",
			Description:        "Implement comprehensive compliance program",
			ActionItems:        []string{"Implement compliance programs", "Conduct audits", "Maintain records"},
			ExpectedBenefit:    "Reduced compliance risk and penalties",
			ImplementationCost: "Low",
			Timeline:           "3-6 months",
		},
	}

	return recommendations
}

// generateGeneralRegulatoryFactors generates general business regulatory factors
func (gm *GeneralModel) generateGeneralRegulatoryFactors() []RegulatoryFactor {
	return []RegulatoryFactor{
		{
			FactorID:       "general_business_license",
			RegulationName: "Business Licensing",
			RegulatoryBody: "Local Government",
			Jurisdiction:   "Local",
			RiskImpact:     0.2,
			ComplianceCost: "Low",
			PenaltyRisk:    "Medium",
			Description:    "General business licensing requirements",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "general_tax_compliance",
			RegulationName: "Tax Compliance",
			RegulatoryBody: "Internal Revenue Service",
			Jurisdiction:   "US",
			RiskImpact:     0.3,
			ComplianceCost: "Medium",
			PenaltyRisk:    "High",
			Description:    "Tax compliance and reporting requirements",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "general_employment_law",
			RegulationName: "Employment Law",
			RegulatoryBody: "Department of Labor",
			Jurisdiction:   "US",
			RiskImpact:     0.25,
			ComplianceCost: "Medium",
			PenaltyRisk:    "Medium",
			Description:    "Employment and labor law compliance",
			LastUpdated:    time.Now(),
		},
	}
}

// generateGeneralMarketFactors generates general business market factors
func (gm *GeneralModel) generateGeneralMarketFactors() []MarketFactor {
	return []MarketFactor{
		{
			FactorID:       "general_economic_conditions",
			FactorName:     "Economic Conditions",
			MarketTrend:    "volatile",
			ImpactScore:    0.7,
			TimeHorizon:    "medium-term",
			Description:    "General economic conditions affecting business",
			KeyDrivers:     []string{"Economic growth", "Inflation", "Interest rates"},
			RiskMitigation: []string{"Monitor economic trends", "Maintain financial flexibility"},
		},
		{
			FactorID:       "general_technology_disruption",
			FactorName:     "Technology Disruption",
			MarketTrend:    "growing",
			ImpactScore:    0.6,
			TimeHorizon:    "long-term",
			Description:    "Technology disruption across industries",
			KeyDrivers:     []string{"Digital transformation", "Automation", "AI/ML"},
			RiskMitigation: []string{"Invest in technology", "Adapt to changes"},
		},
		{
			FactorID:       "general_consumer_behavior",
			FactorName:     "Consumer Behavior",
			MarketTrend:    "stable",
			ImpactScore:    0.5,
			TimeHorizon:    "medium-term",
			Description:    "Changing consumer behavior and preferences",
			KeyDrivers:     []string{"Digital adoption", "Sustainability", "Value consciousness"},
			RiskMitigation: []string{"Monitor trends", "Adapt offerings", "Focus on customer experience"},
		},
	}
}

// generateGeneralOperationalFactors generates general business operational factors
func (gm *GeneralModel) generateGeneralOperationalFactors(business *models.RiskAssessmentRequest) []OperationalFactor {
	return []OperationalFactor{
		{
			FactorID:            "general_operations_management",
			FactorName:          "Operations Management",
			OperationalArea:     "operations",
			RiskScore:           0.3,
			Criticality:         "critical",
			Description:         "Management of general business operations",
			ControlMeasures:     []string{"Process optimization", "Quality control", "Performance monitoring"},
			MonitoringFrequency: "daily",
		},
		{
			FactorID:            "general_customer_service",
			FactorName:          "Customer Service",
			OperationalArea:     "customer_service",
			RiskScore:           0.25,
			Criticality:         "high",
			Description:         "Quality of customer service and support",
			ControlMeasures:     []string{"Service standards", "Training", "Feedback systems"},
			MonitoringFrequency: "weekly",
		},
		{
			FactorID:            "general_human_resources",
			FactorName:          "Human Resources",
			OperationalArea:     "human_resources",
			RiskScore:           0.2,
			Criticality:         "medium",
			Description:         "Management of human resources and workforce",
			ControlMeasures:     []string{"HR policies", "Training", "Performance management"},
			MonitoringFrequency: "monthly",
		},
	}
}

// calculateOverallIndustryRisk calculates the overall industry risk score
func (gm *GeneralModel) calculateOverallIndustryRisk(baseScore float64, factors []IndustryRiskFactor, compliance []ComplianceStatus) float64 {
	// Start with base score
	totalScore := baseScore

	// Weight factors by their risk scores
	for _, factor := range factors {
		totalScore += factor.RiskScore * 0.1 // Weight each factor
	}

	// Adjust based on compliance status
	avgCompliance := 0.0
	for _, status := range compliance {
		avgCompliance += status.ComplianceScore
	}
	if len(compliance) > 0 {
		avgCompliance /= float64(len(compliance))
		totalScore += (1.0 - avgCompliance) * 0.2 // Poor compliance increases risk
	}

	// Ensure score is within bounds
	return math.Max(0.0, math.Min(1.0, totalScore))
}

// determineRiskLevel determines the risk level based on score
func (gm *GeneralModel) determineRiskLevel(score float64) models.RiskLevel {
	switch {
	case score < 0.3:
		return models.RiskLevelLow
	case score < 0.6:
		return models.RiskLevelMedium
	default:
		return models.RiskLevelHigh
	}
}

// calculateRiskLevel calculates risk level for individual factors
func (gm *GeneralModel) calculateRiskLevel(score float64) string {
	switch {
	case score < 0.3:
		return "low"
	case score < 0.6:
		return "medium"
	default:
		return "high"
	}
}

// calculateConfidenceScore calculates confidence in the analysis
func (gm *GeneralModel) calculateConfidenceScore(business *models.RiskAssessmentRequest, factors []IndustryRiskFactor) float64 {
	confidence := 0.7 // Base confidence

	// Increase confidence if we have more business information
	if business.Metadata != nil {
		confidence += 0.1
	}

	// Increase confidence based on number of factors analyzed
	confidence += math.Min(0.2, float64(len(factors))*0.02)

	// Ensure confidence is within bounds
	return math.Max(0.0, math.Min(1.0, confidence))
}
