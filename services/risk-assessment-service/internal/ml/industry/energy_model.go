package industry

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// EnergyModel implements industry-specific risk analysis for energy companies
type EnergyModel struct {
	logger *zap.Logger
}

// NewEnergyModel creates a new energy industry model
func NewEnergyModel(logger *zap.Logger) *EnergyModel {
	return &EnergyModel{
		logger: logger,
	}
}

// GetIndustryType returns the industry type
func (em *EnergyModel) GetIndustryType() IndustryType {
	return IndustryEnergy
}

// CalculateIndustryRisk calculates energy-specific risk factors
func (em *EnergyModel) CalculateIndustryRisk(ctx context.Context, business *models.RiskAssessmentRequest) (*IndustryRiskResult, error) {
	em.logger.Info("Calculating energy industry risk", zap.String("business", business.BusinessName))

	// Calculate base industry risk score
	baseScore := em.calculateBaseEnergyRisk(business)

	// Generate industry-specific factors
	industryFactors := em.generateEnergyRiskFactors(business, baseScore)

	// Calculate compliance status
	complianceStatus := em.assessEnergyCompliance(business)

	// Generate recommendations
	recommendations := em.generateEnergyRecommendations(business, baseScore)

	// Generate regulatory factors
	regulatoryFactors := em.generateEnergyRegulatoryFactors()

	// Generate market factors
	marketFactors := em.generateEnergyMarketFactors()

	// Generate operational factors
	operationalFactors := em.generateEnergyOperationalFactors(business)

	// Calculate overall industry risk score
	industryRiskScore := em.calculateOverallIndustryRisk(baseScore, industryFactors, complianceStatus)

	// Determine risk level
	riskLevel := em.determineRiskLevel(industryRiskScore)

	// Calculate confidence score
	confidenceScore := em.calculateConfidenceScore(business, industryFactors)

	result := &IndustryRiskResult{
		IndustryType:            IndustryEnergy,
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

	em.logger.Info("Energy industry risk calculated",
		zap.Float64("risk_score", industryRiskScore),
		zap.String("risk_level", string(riskLevel)))

	return result, nil
}

// GetIndustrySpecificFactors returns energy-specific risk factors
func (em *EnergyModel) GetIndustrySpecificFactors() []IndustryRiskFactor {
	return []IndustryRiskFactor{
		{
			FactorID:            "energy_commodity_prices",
			FactorName:          "Commodity Price Volatility",
			FactorCategory:      "market",
			Description:         "Risk from volatile energy commodity prices",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "energy_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			Description:         "Compliance with energy regulations",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "energy_environmental",
			FactorName:          "Environmental Risk",
			FactorCategory:      "environmental",
			Description:         "Risk from environmental regulations and climate change",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "energy_operational_safety",
			FactorName:          "Operational Safety",
			FactorCategory:      "operational",
			Description:         "Risk from operational accidents and safety violations",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "energy_technology_disruption",
			FactorName:          "Technology Disruption",
			FactorCategory:      "technology",
			Description:         "Risk from renewable energy and technology disruption",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "energy_infrastructure",
			FactorName:          "Infrastructure Risk",
			FactorCategory:      "operational",
			Description:         "Risk from infrastructure failures and maintenance",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "energy_geopolitical",
			FactorName:          "Geopolitical Risk",
			FactorCategory:      "geopolitical",
			Description:         "Risk from geopolitical events and trade disputes",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "energy_capital_intensity",
			FactorName:          "Capital Intensity",
			FactorCategory:      "financial",
			Description:         "Risk from high capital requirements and financing",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
	}
}

// GetIndustryWeightings returns risk category weightings for energy
func (em *EnergyModel) GetIndustryWeightings() map[string]float64 {
	return map[string]float64{
		"regulatory":    0.20, // High regulatory risk
		"compliance":    0.15, // High compliance requirements
		"operational":   0.25, // Very high operational risk
		"financial":     0.15, // Moderate financial risk
		"reputational":  0.10, // Moderate reputational risk
		"technology":    0.05, // Technology risk
		"geopolitical":  0.08, // Moderate geopolitical risk
		"environmental": 0.02, // Very low environmental risk
	}
}

// ValidateIndustryData validates energy-specific business data
func (em *EnergyModel) ValidateIndustryData(business *models.RiskAssessmentRequest) []string {
	var errors []string

	if business == nil {
		errors = append(errors, "business information is required")
		return errors
	}

	// Check for required energy-specific fields
	if business.BusinessName == "" {
		errors = append(errors, "business name is required")
	}

	return errors
}

// GetIndustryComplianceRequirements returns energy compliance requirements
func (em *EnergyModel) GetIndustryComplianceRequirements() []ComplianceRequirement {
	return []ComplianceRequirement{
		{
			RequirementID:   "energy_epa",
			RequirementName: "EPA Compliance",
			RegulatoryBody:  "Environmental Protection Agency",
			Jurisdiction:    "US",
			Description:     "Environmental protection compliance",
			Required:        true,
			PenaltyAmount:   "Up to $37,500 per violation",
			ComplianceSteps: []string{
				"Implement environmental programs",
				"Monitor emissions",
				"Maintain environmental records",
				"Conduct environmental audits",
			},
			Documentation: []string{
				"Environmental policies",
				"Emission reports",
				"Audit reports",
			},
		},
		{
			RequirementID:   "energy_osha",
			RequirementName: "OSHA Compliance",
			RegulatoryBody:  "Occupational Safety and Health Administration",
			Jurisdiction:    "US",
			Description:     "Workplace safety and health compliance",
			Required:        true,
			PenaltyAmount:   "Up to $136,532 per violation",
			ComplianceSteps: []string{
				"Implement safety programs",
				"Conduct safety training",
				"Maintain safety records",
				"Conduct safety inspections",
			},
			Documentation: []string{
				"Safety policies",
				"Training records",
				"Inspection reports",
			},
		},
		{
			RequirementID:   "energy_ferc",
			RequirementName: "FERC Compliance",
			RegulatoryBody:  "Federal Energy Regulatory Commission",
			Jurisdiction:    "US",
			Description:     "Energy market and transmission compliance",
			Required:        false,
			PenaltyAmount:   "Up to $1M per violation",
			ComplianceSteps: []string{
				"Comply with market rules",
				"Maintain transmission standards",
				"Report market activities",
				"Conduct compliance audits",
			},
			Documentation: []string{
				"Market compliance reports",
				"Transmission standards",
				"Audit reports",
			},
		},
	}
}

// calculateBaseEnergyRisk calculates the base risk score for energy companies
func (em *EnergyModel) calculateBaseEnergyRisk(business *models.RiskAssessmentRequest) float64 {
	baseScore := 0.45 // Base energy risk is high due to regulatory complexity

	// Adjust based on business characteristics
	if business.Metadata != nil {
		// Check for energy type
		if energyType, exists := business.Metadata["energy_type"]; exists {
			if eType, ok := energyType.(string); ok {
				switch eType {
				case "oil", "gas":
					baseScore += 0.05 // Oil and gas have higher risk
				case "renewable", "solar", "wind":
					baseScore -= 0.05 // Renewable energy has lower risk
				case "nuclear":
					baseScore += 0.1 // Nuclear has highest risk
				}
			}
		}
	}

	// Ensure score is within bounds
	return math.Max(0.0, math.Min(1.0, baseScore))
}

// generateEnergyRiskFactors generates energy-specific risk factors
func (em *EnergyModel) generateEnergyRiskFactors(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRiskFactor {
	factors := []IndustryRiskFactor{
		{
			FactorID:            "energy_commodity_prices",
			FactorName:          "Commodity Price Volatility",
			FactorCategory:      "market",
			RiskScore:           baseScore + 0.15,
			RiskLevel:           em.calculateRiskLevel(baseScore + 0.15),
			Description:         "Risk from volatile energy commodity prices",
			Impact:              "high",
			Likelihood:          "high",
			MitigationAdvice:    "Hedge commodity price risk, diversify energy sources, maintain flexible operations",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "energy_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			RiskScore:           baseScore + 0.12,
			RiskLevel:           em.calculateRiskLevel(baseScore + 0.12),
			Description:         "Compliance with energy regulations",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement compliance programs, conduct regular audits, maintain records",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "energy_environmental",
			FactorName:          "Environmental Risk",
			FactorCategory:      "environmental",
			RiskScore:           baseScore + 0.13,
			RiskLevel:           em.calculateRiskLevel(baseScore + 0.13),
			Description:         "Risk from environmental regulations and climate change",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement environmental programs, reduce emissions, invest in clean technology",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "energy_operational_safety",
			FactorName:          "Operational Safety",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.14,
			RiskLevel:           em.calculateRiskLevel(baseScore + 0.14),
			Description:         "Risk from operational accidents and safety violations",
			Impact:              "very_high",
			Likelihood:          "low",
			MitigationAdvice:    "Implement safety programs, conduct training, maintain safety records",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "energy_technology_disruption",
			FactorName:          "Technology Disruption",
			FactorCategory:      "technology",
			RiskScore:           baseScore + 0.08,
			RiskLevel:           em.calculateRiskLevel(baseScore + 0.08),
			Description:         "Risk from renewable energy and technology disruption",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Invest in renewable energy, adapt to technology changes, diversify portfolio",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "energy_infrastructure",
			FactorName:          "Infrastructure Risk",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.11,
			RiskLevel:           em.calculateRiskLevel(baseScore + 0.11),
			Description:         "Risk from infrastructure failures and maintenance",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement preventive maintenance, upgrade infrastructure, maintain redundancy",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "energy_geopolitical",
			FactorName:          "Geopolitical Risk",
			FactorCategory:      "geopolitical",
			RiskScore:           baseScore + 0.09,
			RiskLevel:           em.calculateRiskLevel(baseScore + 0.09),
			Description:         "Risk from geopolitical events and trade disputes",
			Impact:              "medium",
			Likelihood:          "medium",
			MitigationAdvice:    "Diversify supply sources, monitor geopolitical events, maintain contingency plans",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "energy_capital_intensity",
			FactorName:          "Capital Intensity",
			FactorCategory:      "financial",
			RiskScore:           baseScore + 0.07,
			RiskLevel:           em.calculateRiskLevel(baseScore + 0.07),
			Description:         "Risk from high capital requirements and financing",
			Impact:              "high",
			Likelihood:          "low",
			MitigationAdvice:    "Maintain strong credit, diversify financing sources, optimize capital structure",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
	}

	return factors
}

// assessEnergyCompliance assesses compliance status for energy requirements
func (em *EnergyModel) assessEnergyCompliance(business *models.RiskAssessmentRequest) []ComplianceStatus {
	statuses := []ComplianceStatus{
		{
			RequirementID:   "energy_epa",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"EPA compliance not verified"},
			Recommendations: []string{"Implement EPA compliance program"},
		},
		{
			RequirementID:   "energy_osha",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"OSHA compliance not verified"},
			Recommendations: []string{"Implement OSHA compliance program"},
		},
		{
			RequirementID:   "energy_ferc",
			Status:          "not_applicable",
			LastChecked:     time.Now(),
			ComplianceScore: 0.8,
			Issues:          []string{},
			Recommendations: []string{"Monitor FERC requirements if applicable"},
		},
	}

	return statuses
}

// generateEnergyRecommendations generates energy-specific recommendations
func (em *EnergyModel) generateEnergyRecommendations(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRecommendation {
	recommendations := []IndustryRecommendation{
		{
			RecommendationID:   "energy_renewable_transition",
			Category:           "technology",
			Priority:           "high",
			Title:              "Renewable Energy Transition",
			Description:        "Transition to renewable energy sources",
			ActionItems:        []string{"Invest in renewable energy", "Develop clean technology", "Reduce carbon footprint"},
			ExpectedBenefit:    "Reduced environmental risk and improved sustainability",
			ImplementationCost: "High",
			Timeline:           "12-24 months",
		},
		{
			RecommendationID:   "energy_safety_program",
			Category:           "operational",
			Priority:           "high",
			Title:              "Safety Program",
			Description:        "Implement comprehensive safety program",
			ActionItems:        []string{"Implement safety programs", "Conduct training", "Maintain records"},
			ExpectedBenefit:    "Reduced operational safety risk",
			ImplementationCost: "Medium",
			Timeline:           "6-12 months",
		},
		{
			RecommendationID:   "energy_compliance_program",
			Category:           "compliance",
			Priority:           "high",
			Title:              "Compliance Program",
			Description:        "Implement comprehensive compliance program",
			ActionItems:        []string{"Implement compliance programs", "Conduct audits", "Maintain records"},
			ExpectedBenefit:    "Reduced regulatory compliance risk",
			ImplementationCost: "Medium",
			Timeline:           "6-12 months",
		},
		{
			RecommendationID:   "energy_infrastructure_upgrade",
			Category:           "operational",
			Priority:           "medium",
			Title:              "Infrastructure Upgrade",
			Description:        "Upgrade energy infrastructure",
			ActionItems:        []string{"Assess infrastructure", "Implement upgrades", "Maintain systems"},
			ExpectedBenefit:    "Improved infrastructure reliability",
			ImplementationCost: "High",
			Timeline:           "12-24 months",
		},
	}

	return recommendations
}

// generateEnergyRegulatoryFactors generates energy regulatory factors
func (em *EnergyModel) generateEnergyRegulatoryFactors() []RegulatoryFactor {
	return []RegulatoryFactor{
		{
			FactorID:       "energy_epa",
			RegulationName: "Environmental Protection Regulations",
			RegulatoryBody: "Environmental Protection Agency",
			Jurisdiction:   "US",
			RiskImpact:     0.3,
			ComplianceCost: "High",
			PenaltyRisk:    "High",
			Description:    "Environmental protection and pollution control",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "energy_osha",
			RegulationName: "Occupational Safety and Health Act",
			RegulatoryBody: "Occupational Safety and Health Administration",
			Jurisdiction:   "US",
			RiskImpact:     0.25,
			ComplianceCost: "Medium",
			PenaltyRisk:    "High",
			Description:    "Workplace safety and health regulations",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "energy_ferc",
			RegulationName: "Federal Energy Regulatory Commission",
			RegulatoryBody: "Federal Energy Regulatory Commission",
			Jurisdiction:   "US",
			RiskImpact:     0.2,
			ComplianceCost: "Medium",
			PenaltyRisk:    "Medium",
			Description:    "Energy market and transmission regulations",
			LastUpdated:    time.Now(),
		},
	}
}

// generateEnergyMarketFactors generates energy market factors
func (em *EnergyModel) generateEnergyMarketFactors() []MarketFactor {
	return []MarketFactor{
		{
			FactorID:       "energy_commodity_prices",
			FactorName:     "Commodity Prices",
			MarketTrend:    "volatile",
			ImpactScore:    0.8,
			TimeHorizon:    "short-term",
			Description:    "Volatile energy commodity prices",
			KeyDrivers:     []string{"Supply and demand", "Geopolitical events", "Economic conditions"},
			RiskMitigation: []string{"Hedge price risk", "Diversify energy sources"},
		},
		{
			FactorID:       "energy_renewable_transition",
			FactorName:     "Renewable Energy Transition",
			MarketTrend:    "growing",
			ImpactScore:    0.7,
			TimeHorizon:    "long-term",
			Description:    "Transition to renewable energy sources",
			KeyDrivers:     []string{"Climate change", "Technology advancement", "Regulatory pressure"},
			RiskMitigation: []string{"Invest in renewables", "Adapt to changes"},
		},
		{
			FactorID:       "energy_demand_growth",
			FactorName:     "Energy Demand Growth",
			MarketTrend:    "growing",
			ImpactScore:    0.6,
			TimeHorizon:    "long-term",
			Description:    "Growing global energy demand",
			KeyDrivers:     []string{"Population growth", "Economic development", "Urbanization"},
			RiskMitigation: []string{"Invest in capacity", "Focus on efficiency"},
		},
	}
}

// generateEnergyOperationalFactors generates energy operational factors
func (em *EnergyModel) generateEnergyOperationalFactors(business *models.RiskAssessmentRequest) []OperationalFactor {
	return []OperationalFactor{
		{
			FactorID:            "energy_production_efficiency",
			FactorName:          "Production Efficiency",
			OperationalArea:     "production",
			RiskScore:           0.3,
			Criticality:         "critical",
			Description:         "Efficiency of energy production processes",
			ControlMeasures:     []string{"Process optimization", "Technology upgrades", "Maintenance"},
			MonitoringFrequency: "continuous",
		},
		{
			FactorID:            "energy_safety_management",
			FactorName:          "Safety Management",
			OperationalArea:     "safety",
			RiskScore:           0.25,
			Criticality:         "critical",
			Description:         "Management of operational safety",
			ControlMeasures:     []string{"Safety programs", "Training", "Inspections"},
			MonitoringFrequency: "daily",
		},
		{
			FactorID:            "energy_environmental_management",
			FactorName:          "Environmental Management",
			OperationalArea:     "environmental",
			RiskScore:           0.2,
			Criticality:         "high",
			Description:         "Management of environmental impact",
			ControlMeasures:     []string{"Environmental programs", "Monitoring", "Compliance"},
			MonitoringFrequency: "weekly",
		},
	}
}

// calculateOverallIndustryRisk calculates the overall industry risk score
func (em *EnergyModel) calculateOverallIndustryRisk(baseScore float64, factors []IndustryRiskFactor, compliance []ComplianceStatus) float64 {
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
func (em *EnergyModel) determineRiskLevel(score float64) models.RiskLevel {
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
func (em *EnergyModel) calculateRiskLevel(score float64) string {
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
func (em *EnergyModel) calculateConfidenceScore(business *models.RiskAssessmentRequest, factors []IndustryRiskFactor) float64 {
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
