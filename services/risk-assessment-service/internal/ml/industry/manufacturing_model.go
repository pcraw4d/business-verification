package industry

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// ManufacturingModel implements industry-specific risk analysis for manufacturing companies
type ManufacturingModel struct {
	logger *zap.Logger
}

// NewManufacturingModel creates a new manufacturing industry model
func NewManufacturingModel(logger *zap.Logger) *ManufacturingModel {
	return &ManufacturingModel{
		logger: logger,
	}
}

// GetIndustryType returns the industry type
func (mm *ManufacturingModel) GetIndustryType() IndustryType {
	return IndustryManufacturing
}

// CalculateIndustryRisk calculates manufacturing-specific risk factors
func (mm *ManufacturingModel) CalculateIndustryRisk(ctx context.Context, business *models.RiskAssessmentRequest) (*IndustryRiskResult, error) {
	mm.logger.Info("Calculating manufacturing industry risk", zap.String("business", business.BusinessName))

	// Calculate base industry risk score
	baseScore := mm.calculateBaseManufacturingRisk(business)

	// Generate industry-specific factors
	industryFactors := mm.generateManufacturingRiskFactors(business, baseScore)

	// Calculate compliance status
	complianceStatus := mm.assessManufacturingCompliance(business)

	// Generate recommendations
	recommendations := mm.generateManufacturingRecommendations(business, baseScore)

	// Generate regulatory factors
	regulatoryFactors := mm.generateManufacturingRegulatoryFactors()

	// Generate market factors
	marketFactors := mm.generateManufacturingMarketFactors()

	// Generate operational factors
	operationalFactors := mm.generateManufacturingOperationalFactors(business)

	// Calculate overall industry risk score
	industryRiskScore := mm.calculateOverallIndustryRisk(baseScore, industryFactors, complianceStatus)

	// Determine risk level
	riskLevel := mm.determineRiskLevel(industryRiskScore)

	// Calculate confidence score
	confidenceScore := mm.calculateConfidenceScore(business, industryFactors)

	result := &IndustryRiskResult{
		IndustryType:            IndustryManufacturing,
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

	mm.logger.Info("Manufacturing industry risk calculated",
		zap.Float64("risk_score", industryRiskScore),
		zap.String("risk_level", string(riskLevel)))

	return result, nil
}

// GetIndustrySpecificFactors returns manufacturing-specific risk factors
func (mm *ManufacturingModel) GetIndustrySpecificFactors() []IndustryRiskFactor {
	return []IndustryRiskFactor{
		{
			FactorID:            "manufacturing_supply_chain",
			FactorName:          "Supply Chain Risk",
			FactorCategory:      "operational",
			Description:         "Risk from supply chain disruptions and dependencies",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "manufacturing_quality_control",
			FactorName:          "Quality Control",
			FactorCategory:      "operational",
			Description:         "Risk from product quality issues and recalls",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "manufacturing_environmental",
			FactorName:          "Environmental Compliance",
			FactorCategory:      "compliance",
			Description:         "Compliance with environmental regulations",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "manufacturing_workplace_safety",
			FactorName:          "Workplace Safety",
			FactorCategory:      "operational",
			Description:         "Risk from workplace accidents and safety violations",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "manufacturing_technology_obsolescence",
			FactorName:          "Technology Obsolescence",
			FactorCategory:      "technology",
			Description:         "Risk from outdated manufacturing technology",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "manufacturing_capacity_utilization",
			FactorName:          "Capacity Utilization",
			FactorCategory:      "operational",
			Description:         "Risk from underutilized manufacturing capacity",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "manufacturing_energy_costs",
			FactorName:          "Energy Costs",
			FactorCategory:      "financial",
			Description:         "Risk from volatile energy costs",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "manufacturing_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			Description:         "Compliance with manufacturing regulations",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
	}
}

// GetIndustryWeightings returns risk category weightings for manufacturing
func (mm *ManufacturingModel) GetIndustryWeightings() map[string]float64 {
	return map[string]float64{
		"regulatory":    0.15, // Moderate regulatory risk
		"compliance":    0.20, // High compliance requirements
		"operational":   0.30, // Very high operational risk
		"financial":     0.15, // Moderate financial risk
		"reputational":  0.10, // Moderate reputational risk
		"technology":    0.05, // Technology risk
		"geopolitical":  0.03, // Low geopolitical risk
		"environmental": 0.02, // Very low environmental risk
	}
}

// ValidateIndustryData validates manufacturing-specific business data
func (mm *ManufacturingModel) ValidateIndustryData(business *models.RiskAssessmentRequest) []string {
	var errors []string

	if business == nil {
		errors = append(errors, "business information is required")
		return errors
	}

	// Check for required manufacturing-specific fields
	if business.BusinessName == "" {
		errors = append(errors, "business name is required")
	}

	return errors
}

// GetIndustryComplianceRequirements returns manufacturing compliance requirements
func (mm *ManufacturingModel) GetIndustryComplianceRequirements() []ComplianceRequirement {
	return []ComplianceRequirement{
		{
			RequirementID:   "manufacturing_osha",
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
			RequirementID:   "manufacturing_epa",
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
			RequirementID:   "manufacturing_fda",
			RequirementName: "FDA Compliance",
			RegulatoryBody:  "Food and Drug Administration",
			Jurisdiction:    "US",
			Description:     "FDA regulations for food and drug manufacturing",
			Required:        false,
			PenaltyAmount:   "Product recalls and fines",
			ComplianceSteps: []string{
				"Implement quality systems",
				"Conduct quality testing",
				"Maintain quality records",
				"Report quality issues",
			},
			Documentation: []string{
				"Quality system documentation",
				"Testing reports",
				"Quality records",
			},
		},
	}
}

// calculateBaseManufacturingRisk calculates the base risk score for manufacturing companies
func (mm *ManufacturingModel) calculateBaseManufacturingRisk(business *models.RiskAssessmentRequest) float64 {
	baseScore := 0.4 // Base manufacturing risk is moderate-high

	// Adjust based on business characteristics
	if business.Metadata != nil {
		// Check for manufacturing type
		if mfgType, exists := business.Metadata["manufacturing_type"]; exists {
			if mType, ok := mfgType.(string); ok {
				switch mType {
				case "automotive", "aerospace":
					baseScore += 0.05 // High-tech manufacturing has higher risk
				case "food", "pharmaceutical":
					baseScore += 0.1 // Regulated industries have higher risk
				case "textile", "furniture":
					baseScore -= 0.05 // Lower-tech manufacturing has lower risk
				}
			}
		}
	}

	// Ensure score is within bounds
	return math.Max(0.0, math.Min(1.0, baseScore))
}

// generateManufacturingRiskFactors generates manufacturing-specific risk factors
func (mm *ManufacturingModel) generateManufacturingRiskFactors(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRiskFactor {
	factors := []IndustryRiskFactor{
		{
			FactorID:            "manufacturing_supply_chain",
			FactorName:          "Supply Chain Risk",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.15,
			RiskLevel:           mm.calculateRiskLevel(baseScore + 0.15),
			Description:         "Risk from supply chain disruptions and dependencies",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Diversify suppliers, implement supply chain monitoring, maintain safety stock",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "manufacturing_quality_control",
			FactorName:          "Quality Control",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.12,
			RiskLevel:           mm.calculateRiskLevel(baseScore + 0.12),
			Description:         "Risk from product quality issues and recalls",
			Impact:              "high",
			Likelihood:          "low",
			MitigationAdvice:    "Implement quality systems, conduct regular testing, maintain quality records",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "manufacturing_environmental",
			FactorName:          "Environmental Compliance",
			FactorCategory:      "compliance",
			RiskScore:           baseScore + 0.10,
			RiskLevel:           mm.calculateRiskLevel(baseScore + 0.10),
			Description:         "Compliance with environmental regulations",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement environmental programs, monitor emissions, conduct audits",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "manufacturing_workplace_safety",
			FactorName:          "Workplace Safety",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.13,
			RiskLevel:           mm.calculateRiskLevel(baseScore + 0.13),
			Description:         "Risk from workplace accidents and safety violations",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement safety programs, conduct training, maintain safety records",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "manufacturing_technology_obsolescence",
			FactorName:          "Technology Obsolescence",
			FactorCategory:      "technology",
			RiskScore:           baseScore + 0.08,
			RiskLevel:           mm.calculateRiskLevel(baseScore + 0.08),
			Description:         "Risk from outdated manufacturing technology",
			Impact:              "medium",
			Likelihood:          "medium",
			MitigationAdvice:    "Invest in modern technology, implement automation, upgrade equipment",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "manufacturing_capacity_utilization",
			FactorName:          "Capacity Utilization",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.09,
			RiskLevel:           mm.calculateRiskLevel(baseScore + 0.09),
			Description:         "Risk from underutilized manufacturing capacity",
			Impact:              "medium",
			Likelihood:          "medium",
			MitigationAdvice:    "Optimize production planning, diversify products, improve efficiency",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "manufacturing_energy_costs",
			FactorName:          "Energy Costs",
			FactorCategory:      "financial",
			RiskScore:           baseScore + 0.07,
			RiskLevel:           mm.calculateRiskLevel(baseScore + 0.07),
			Description:         "Risk from volatile energy costs",
			Impact:              "medium",
			Likelihood:          "high",
			MitigationAdvice:    "Implement energy efficiency measures, use renewable energy, hedge energy costs",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "manufacturing_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			RiskScore:           baseScore + 0.11,
			RiskLevel:           mm.calculateRiskLevel(baseScore + 0.11),
			Description:         "Compliance with manufacturing regulations",
			Impact:              "high",
			Likelihood:          "low",
			MitigationAdvice:    "Implement compliance programs, conduct regular audits, maintain records",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
	}

	return factors
}

// assessManufacturingCompliance assesses compliance status for manufacturing requirements
func (mm *ManufacturingModel) assessManufacturingCompliance(business *models.RiskAssessmentRequest) []ComplianceStatus {
	statuses := []ComplianceStatus{
		{
			RequirementID:   "manufacturing_osha",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"OSHA compliance not verified"},
			Recommendations: []string{"Implement OSHA compliance program"},
		},
		{
			RequirementID:   "manufacturing_epa",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"EPA compliance not verified"},
			Recommendations: []string{"Implement EPA compliance program"},
		},
		{
			RequirementID:   "manufacturing_fda",
			Status:          "not_applicable",
			LastChecked:     time.Now(),
			ComplianceScore: 0.8,
			Issues:          []string{},
			Recommendations: []string{"Monitor FDA requirements if applicable"},
		},
	}

	return statuses
}

// generateManufacturingRecommendations generates manufacturing-specific recommendations
func (mm *ManufacturingModel) generateManufacturingRecommendations(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRecommendation {
	recommendations := []IndustryRecommendation{
		{
			RecommendationID:   "manufacturing_supply_chain",
			Category:           "operational",
			Priority:           "high",
			Title:              "Supply Chain Optimization",
			Description:        "Optimize supply chain and reduce dependencies",
			ActionItems:        []string{"Diversify suppliers", "Implement monitoring", "Maintain safety stock"},
			ExpectedBenefit:    "Reduced supply chain risk",
			ImplementationCost: "Medium",
			Timeline:           "6-12 months",
		},
		{
			RecommendationID:   "manufacturing_quality_systems",
			Category:           "operational",
			Priority:           "high",
			Title:              "Quality Systems",
			Description:        "Implement comprehensive quality management systems",
			ActionItems:        []string{"Implement quality systems", "Conduct testing", "Maintain records"},
			ExpectedBenefit:    "Improved product quality and reduced recalls",
			ImplementationCost: "Medium",
			Timeline:           "6-12 months",
		},
		{
			RecommendationID:   "manufacturing_automation",
			Category:           "technology",
			Priority:           "medium",
			Title:              "Manufacturing Automation",
			Description:        "Implement automation and modern manufacturing technology",
			ActionItems:        []string{"Assess automation opportunities", "Implement automation", "Train staff"},
			ExpectedBenefit:    "Improved efficiency and reduced costs",
			ImplementationCost: "High",
			Timeline:           "12-18 months",
		},
		{
			RecommendationID:   "manufacturing_safety",
			Category:           "operational",
			Priority:           "high",
			Title:              "Workplace Safety",
			Description:        "Enhance workplace safety programs",
			ActionItems:        []string{"Implement safety programs", "Conduct training", "Maintain records"},
			ExpectedBenefit:    "Reduced workplace accidents and OSHA violations",
			ImplementationCost: "Low",
			Timeline:           "3-6 months",
		},
	}

	return recommendations
}

// generateManufacturingRegulatoryFactors generates manufacturing regulatory factors
func (mm *ManufacturingModel) generateManufacturingRegulatoryFactors() []RegulatoryFactor {
	return []RegulatoryFactor{
		{
			FactorID:       "manufacturing_osha",
			RegulationName: "Occupational Safety and Health Act",
			RegulatoryBody: "Occupational Safety and Health Administration",
			Jurisdiction:   "US",
			RiskImpact:     0.3,
			ComplianceCost: "Medium",
			PenaltyRisk:    "High",
			Description:    "Workplace safety and health regulations",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "manufacturing_epa",
			RegulationName: "Environmental Protection Regulations",
			RegulatoryBody: "Environmental Protection Agency",
			Jurisdiction:   "US",
			RiskImpact:     0.25,
			ComplianceCost: "Medium",
			PenaltyRisk:    "Medium",
			Description:    "Environmental protection and pollution control",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "manufacturing_fda",
			RegulationName: "Food and Drug Administration Regulations",
			RegulatoryBody: "Food and Drug Administration",
			Jurisdiction:   "US",
			RiskImpact:     0.2,
			ComplianceCost: "High",
			PenaltyRisk:    "High",
			Description:    "Food and drug manufacturing regulations",
			LastUpdated:    time.Now(),
		},
	}
}

// generateManufacturingMarketFactors generates manufacturing market factors
func (mm *ManufacturingModel) generateManufacturingMarketFactors() []MarketFactor {
	return []MarketFactor{
		{
			FactorID:       "manufacturing_globalization",
			FactorName:     "Globalization",
			MarketTrend:    "stable",
			ImpactScore:    0.6,
			TimeHorizon:    "long-term",
			Description:    "Global manufacturing competition and supply chains",
			KeyDrivers:     []string{"Global competition", "Supply chain complexity", "Trade policies"},
			RiskMitigation: []string{"Diversify markets", "Optimize supply chain", "Focus on quality"},
		},
		{
			FactorID:       "manufacturing_automation",
			FactorName:     "Automation Trend",
			MarketTrend:    "growing",
			ImpactScore:    0.7,
			TimeHorizon:    "medium-term",
			Description:    "Increasing automation in manufacturing",
			KeyDrivers:     []string{"Technology advancement", "Cost pressure", "Labor shortage"},
			RiskMitigation: []string{"Invest in automation", "Train workforce", "Focus on high-value products"},
		},
		{
			FactorID:       "manufacturing_sustainability",
			FactorName:     "Sustainability",
			MarketTrend:    "growing",
			ImpactScore:    0.5,
			TimeHorizon:    "long-term",
			Description:    "Increasing focus on sustainable manufacturing",
			KeyDrivers:     []string{"Environmental concerns", "Regulatory pressure", "Consumer demand"},
			RiskMitigation: []string{"Implement green manufacturing", "Reduce waste", "Use renewable energy"},
		},
	}
}

// generateManufacturingOperationalFactors generates manufacturing operational factors
func (mm *ManufacturingModel) generateManufacturingOperationalFactors(business *models.RiskAssessmentRequest) []OperationalFactor {
	return []OperationalFactor{
		{
			FactorID:            "manufacturing_production_efficiency",
			FactorName:          "Production Efficiency",
			OperationalArea:     "production",
			RiskScore:           0.3,
			Criticality:         "critical",
			Description:         "Efficiency of production processes",
			ControlMeasures:     []string{"Process optimization", "Automation", "Quality control"},
			MonitoringFrequency: "daily",
		},
		{
			FactorID:            "manufacturing_equipment_maintenance",
			FactorName:          "Equipment Maintenance",
			OperationalArea:     "maintenance",
			RiskScore:           0.25,
			Criticality:         "high",
			Description:         "Maintenance of manufacturing equipment",
			ControlMeasures:     []string{"Preventive maintenance", "Predictive maintenance", "Spare parts inventory"},
			MonitoringFrequency: "weekly",
		},
		{
			FactorID:            "manufacturing_workforce_management",
			FactorName:          "Workforce Management",
			OperationalArea:     "human_resources",
			RiskScore:           0.2,
			Criticality:         "medium",
			Description:         "Management of manufacturing workforce",
			ControlMeasures:     []string{"Training programs", "Safety protocols", "Performance monitoring"},
			MonitoringFrequency: "monthly",
		},
	}
}

// calculateOverallIndustryRisk calculates the overall industry risk score
func (mm *ManufacturingModel) calculateOverallIndustryRisk(baseScore float64, factors []IndustryRiskFactor, compliance []ComplianceStatus) float64 {
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
func (mm *ManufacturingModel) determineRiskLevel(score float64) models.RiskLevel {
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
func (mm *ManufacturingModel) calculateRiskLevel(score float64) string {
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
func (mm *ManufacturingModel) calculateConfidenceScore(business *models.RiskAssessmentRequest, factors []IndustryRiskFactor) float64 {
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
