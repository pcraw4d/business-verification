package industry

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// TransportationModel implements industry-specific risk analysis for transportation companies
type TransportationModel struct {
	logger *zap.Logger
}

// NewTransportationModel creates a new transportation industry model
func NewTransportationModel(logger *zap.Logger) *TransportationModel {
	return &TransportationModel{
		logger: logger,
	}
}

// GetIndustryType returns the industry type
func (tm *TransportationModel) GetIndustryType() IndustryType {
	return IndustryTransportation
}

// CalculateIndustryRisk calculates transportation-specific risk factors
func (tm *TransportationModel) CalculateIndustryRisk(ctx context.Context, business *models.RiskAssessmentRequest) (*IndustryRiskResult, error) {
	tm.logger.Info("Calculating transportation industry risk", zap.String("business", business.BusinessName))

	// Calculate base industry risk score
	baseScore := tm.calculateBaseTransportationRisk(business)

	// Generate industry-specific factors
	industryFactors := tm.generateTransportationRiskFactors(business, baseScore)

	// Calculate compliance status
	complianceStatus := tm.assessTransportationCompliance(business)

	// Generate recommendations
	recommendations := tm.generateTransportationRecommendations(business, baseScore)

	// Generate regulatory factors
	regulatoryFactors := tm.generateTransportationRegulatoryFactors()

	// Generate market factors
	marketFactors := tm.generateTransportationMarketFactors()

	// Generate operational factors
	operationalFactors := tm.generateTransportationOperationalFactors(business)

	// Calculate overall industry risk score
	industryRiskScore := tm.calculateOverallIndustryRisk(baseScore, industryFactors, complianceStatus)

	// Determine risk level
	riskLevel := tm.determineRiskLevel(industryRiskScore)

	// Calculate confidence score
	confidenceScore := tm.calculateConfidenceScore(business, industryFactors)

	result := &IndustryRiskResult{
		IndustryType:            IndustryTransportation,
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

	tm.logger.Info("Transportation industry risk calculated",
		zap.Float64("risk_score", industryRiskScore),
		zap.String("risk_level", string(riskLevel)))

	return result, nil
}

// GetIndustrySpecificFactors returns transportation-specific risk factors
func (tm *TransportationModel) GetIndustrySpecificFactors() []IndustryRiskFactor {
	return []IndustryRiskFactor{
		{
			FactorID:            "transportation_safety",
			FactorName:          "Safety Risk",
			FactorCategory:      "operational",
			Description:         "Risk from transportation accidents and safety violations",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "transportation_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			Description:         "Compliance with transportation regulations",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "transportation_fuel_costs",
			FactorName:          "Fuel Cost Volatility",
			FactorCategory:      "financial",
			Description:         "Risk from volatile fuel costs",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "transportation_technology_disruption",
			FactorName:          "Technology Disruption",
			FactorCategory:      "technology",
			Description:         "Risk from autonomous vehicles and technology disruption",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "transportation_infrastructure",
			FactorName:          "Infrastructure Risk",
			FactorCategory:      "operational",
			Description:         "Risk from infrastructure failures and congestion",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "transportation_environmental",
			FactorName:          "Environmental Risk",
			FactorCategory:      "environmental",
			Description:         "Risk from environmental regulations and emissions",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "transportation_workforce",
			FactorName:          "Workforce Risk",
			FactorCategory:      "operational",
			Description:         "Risk from driver shortage and workforce issues",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "transportation_cybersecurity",
			FactorName:          "Cybersecurity Risk",
			FactorCategory:      "operational",
			Description:         "Risk from cyber attacks on transportation systems",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
	}
}

// GetIndustryWeightings returns risk category weightings for transportation
func (tm *TransportationModel) GetIndustryWeightings() map[string]float64 {
	return map[string]float64{
		"regulatory":    0.20, // High regulatory risk
		"compliance":    0.15, // High compliance requirements
		"operational":   0.30, // Very high operational risk
		"financial":     0.15, // Moderate financial risk
		"reputational":  0.10, // Moderate reputational risk
		"technology":    0.05, // Technology risk
		"geopolitical":  0.03, // Low geopolitical risk
		"environmental": 0.02, // Very low environmental risk
	}
}

// ValidateIndustryData validates transportation-specific business data
func (tm *TransportationModel) ValidateIndustryData(business *models.RiskAssessmentRequest) []string {
	var errors []string

	if business == nil {
		errors = append(errors, "business information is required")
		return errors
	}

	// Check for required transportation-specific fields
	if business.BusinessName == "" {
		errors = append(errors, "business name is required")
	}

	return errors
}

// GetIndustryComplianceRequirements returns transportation compliance requirements
func (tm *TransportationModel) GetIndustryComplianceRequirements() []ComplianceRequirement {
	return []ComplianceRequirement{
		{
			RequirementID:   "transportation_dot",
			RequirementName: "DOT Compliance",
			RegulatoryBody:  "Department of Transportation",
			Jurisdiction:    "US",
			Description:     "Department of Transportation compliance",
			Required:        true,
			PenaltyAmount:   "Up to $10,000 per violation",
			ComplianceSteps: []string{
				"Maintain DOT registration",
				"Comply with safety regulations",
				"Maintain driver records",
				"Conduct safety inspections",
			},
			Documentation: []string{
				"DOT registration",
				"Safety records",
				"Inspection reports",
			},
		},
		{
			RequirementID:   "transportation_fmcsa",
			RequirementName: "FMCSA Compliance",
			RegulatoryBody:  "Federal Motor Carrier Safety Administration",
			Jurisdiction:    "US",
			Description:     "Federal Motor Carrier Safety Administration compliance",
			Required:        true,
			PenaltyAmount:   "Up to $5,000 per violation",
			ComplianceSteps: []string{
				"Maintain FMCSA registration",
				"Comply with safety regulations",
				"Maintain driver records",
				"Conduct safety inspections",
			},
			Documentation: []string{
				"FMCSA registration",
				"Safety records",
				"Inspection reports",
			},
		},
		{
			RequirementID:   "transportation_tsa",
			RequirementName: "TSA Compliance",
			RegulatoryBody:  "Transportation Security Administration",
			Jurisdiction:    "US",
			Description:     "Transportation Security Administration compliance",
			Required:        false,
			PenaltyAmount:   "Up to $10,000 per violation",
			ComplianceSteps: []string{
				"Implement security programs",
				"Conduct security training",
				"Maintain security records",
				"Conduct security inspections",
			},
			Documentation: []string{
				"Security programs",
				"Training records",
				"Inspection reports",
			},
		},
	}
}

// calculateBaseTransportationRisk calculates the base risk score for transportation companies
func (tm *TransportationModel) calculateBaseTransportationRisk(business *models.RiskAssessmentRequest) float64 {
	baseScore := 0.4 // Base transportation risk is moderate-high

	// Adjust based on business characteristics
	if business.Metadata != nil {
		// Check for transportation type
		if transType, exists := business.Metadata["transportation_type"]; exists {
			if tType, ok := transType.(string); ok {
				switch tType {
				case "airline", "aviation":
					baseScore += 0.1 // Aviation has higher risk
				case "trucking", "logistics":
					baseScore += 0.05 // Trucking has moderate risk
				case "rail", "railroad":
					baseScore += 0.03 // Rail has lower risk
				}
			}
		}
	}

	// Ensure score is within bounds
	return math.Max(0.0, math.Min(1.0, baseScore))
}

// generateTransportationRiskFactors generates transportation-specific risk factors
func (tm *TransportationModel) generateTransportationRiskFactors(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRiskFactor {
	factors := []IndustryRiskFactor{
		{
			FactorID:            "transportation_safety",
			FactorName:          "Safety Risk",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.15,
			RiskLevel:           tm.calculateRiskLevel(baseScore + 0.15),
			Description:         "Risk from transportation accidents and safety violations",
			Impact:              "very_high",
			Likelihood:          "low",
			MitigationAdvice:    "Implement safety programs, conduct training, maintain safety records",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "transportation_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			RiskScore:           baseScore + 0.12,
			RiskLevel:           tm.calculateRiskLevel(baseScore + 0.12),
			Description:         "Compliance with transportation regulations",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement compliance programs, conduct regular audits, maintain records",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "transportation_fuel_costs",
			FactorName:          "Fuel Cost Volatility",
			FactorCategory:      "financial",
			RiskScore:           baseScore + 0.13,
			RiskLevel:           tm.calculateRiskLevel(baseScore + 0.13),
			Description:         "Risk from volatile fuel costs",
			Impact:              "medium",
			Likelihood:          "high",
			MitigationAdvice:    "Hedge fuel costs, improve fuel efficiency, use alternative fuels",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "transportation_technology_disruption",
			FactorName:          "Technology Disruption",
			FactorCategory:      "technology",
			RiskScore:           baseScore + 0.08,
			RiskLevel:           tm.calculateRiskLevel(baseScore + 0.08),
			Description:         "Risk from autonomous vehicles and technology disruption",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Invest in technology, adapt to changes, focus on human-AI collaboration",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "transportation_infrastructure",
			FactorName:          "Infrastructure Risk",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.11,
			RiskLevel:           tm.calculateRiskLevel(baseScore + 0.11),
			Description:         "Risk from infrastructure failures and congestion",
			Impact:              "medium",
			Likelihood:          "medium",
			MitigationAdvice:    "Monitor infrastructure, implement contingency plans, optimize routes",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "transportation_environmental",
			FactorName:          "Environmental Risk",
			FactorCategory:      "environmental",
			RiskScore:           baseScore + 0.09,
			RiskLevel:           tm.calculateRiskLevel(baseScore + 0.09),
			Description:         "Risk from environmental regulations and emissions",
			Impact:              "medium",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement environmental programs, reduce emissions, use clean technology",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "transportation_workforce",
			FactorName:          "Workforce Risk",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.14,
			RiskLevel:           tm.calculateRiskLevel(baseScore + 0.14),
			Description:         "Risk from driver shortage and workforce issues",
			Impact:              "high",
			Likelihood:          "high",
			MitigationAdvice:    "Improve driver retention, offer competitive benefits, invest in training",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "transportation_cybersecurity",
			FactorName:          "Cybersecurity Risk",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.10,
			RiskLevel:           tm.calculateRiskLevel(baseScore + 0.10),
			Description:         "Risk from cyber attacks on transportation systems",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement cybersecurity measures, conduct training, maintain incident response",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
	}

	return factors
}

// assessTransportationCompliance assesses compliance status for transportation requirements
func (tm *TransportationModel) assessTransportationCompliance(business *models.RiskAssessmentRequest) []ComplianceStatus {
	statuses := []ComplianceStatus{
		{
			RequirementID:   "transportation_dot",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"DOT compliance not verified"},
			Recommendations: []string{"Implement DOT compliance program"},
		},
		{
			RequirementID:   "transportation_fmcsa",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"FMCSA compliance not verified"},
			Recommendations: []string{"Implement FMCSA compliance program"},
		},
		{
			RequirementID:   "transportation_tsa",
			Status:          "not_applicable",
			LastChecked:     time.Now(),
			ComplianceScore: 0.8,
			Issues:          []string{},
			Recommendations: []string{"Monitor TSA requirements if applicable"},
		},
	}

	return statuses
}

// generateTransportationRecommendations generates transportation-specific recommendations
func (tm *TransportationModel) generateTransportationRecommendations(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRecommendation {
	recommendations := []IndustryRecommendation{
		{
			RecommendationID:   "transportation_safety_program",
			Category:           "operational",
			Priority:           "high",
			Title:              "Safety Program",
			Description:        "Implement comprehensive safety program",
			ActionItems:        []string{"Implement safety programs", "Conduct training", "Maintain records"},
			ExpectedBenefit:    "Reduced safety risk and accidents",
			ImplementationCost: "Medium",
			Timeline:           "6-12 months",
		},
		{
			RecommendationID:   "transportation_compliance_program",
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
			RecommendationID:   "transportation_technology_upgrade",
			Category:           "technology",
			Priority:           "medium",
			Title:              "Technology Upgrade",
			Description:        "Upgrade transportation technology",
			ActionItems:        []string{"Invest in technology", "Implement automation", "Train staff"},
			ExpectedBenefit:    "Improved efficiency and reduced costs",
			ImplementationCost: "High",
			Timeline:           "12-18 months",
		},
		{
			RecommendationID:   "transportation_workforce_development",
			Category:           "operational",
			Priority:           "medium",
			Title:              "Workforce Development",
			Description:        "Develop and retain transportation workforce",
			ActionItems:        []string{"Improve retention", "Offer benefits", "Invest in training"},
			ExpectedBenefit:    "Reduced workforce risk and improved performance",
			ImplementationCost: "Medium",
			Timeline:           "6-12 months",
		},
	}

	return recommendations
}

// generateTransportationRegulatoryFactors generates transportation regulatory factors
func (tm *TransportationModel) generateTransportationRegulatoryFactors() []RegulatoryFactor {
	return []RegulatoryFactor{
		{
			FactorID:       "transportation_dot",
			RegulationName: "Department of Transportation Regulations",
			RegulatoryBody: "Department of Transportation",
			Jurisdiction:   "US",
			RiskImpact:     0.3,
			ComplianceCost: "Medium",
			PenaltyRisk:    "High",
			Description:    "Transportation safety and regulation compliance",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "transportation_fmcsa",
			RegulationName: "Federal Motor Carrier Safety Administration",
			RegulatoryBody: "Federal Motor Carrier Safety Administration",
			Jurisdiction:   "US",
			RiskImpact:     0.25,
			ComplianceCost: "Medium",
			PenaltyRisk:    "High",
			Description:    "Motor carrier safety and regulation compliance",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "transportation_tsa",
			RegulationName: "Transportation Security Administration",
			RegulatoryBody: "Transportation Security Administration",
			Jurisdiction:   "US",
			RiskImpact:     0.2,
			ComplianceCost: "Medium",
			PenaltyRisk:    "Medium",
			Description:    "Transportation security and regulation compliance",
			LastUpdated:    time.Now(),
		},
	}
}

// generateTransportationMarketFactors generates transportation market factors
func (tm *TransportationModel) generateTransportationMarketFactors() []MarketFactor {
	return []MarketFactor{
		{
			FactorID:       "transportation_ecommerce_growth",
			FactorName:     "E-commerce Growth",
			MarketTrend:    "growing",
			ImpactScore:    0.7,
			TimeHorizon:    "long-term",
			Description:    "Growing e-commerce driving transportation demand",
			KeyDrivers:     []string{"E-commerce growth", "Consumer preference", "Technology advancement"},
			RiskMitigation: []string{"Invest in logistics", "Focus on last-mile delivery"},
		},
		{
			FactorID:       "transportation_autonomous_vehicles",
			FactorName:     "Autonomous Vehicles",
			MarketTrend:    "growing",
			ImpactScore:    0.6,
			TimeHorizon:    "medium-term",
			Description:    "Development of autonomous vehicles",
			KeyDrivers:     []string{"Technology advancement", "Safety improvements", "Cost reduction"},
			RiskMitigation: []string{"Invest in technology", "Adapt to changes"},
		},
		{
			FactorID:       "transportation_sustainability",
			FactorName:     "Sustainability",
			MarketTrend:    "growing",
			ImpactScore:    0.5,
			TimeHorizon:    "long-term",
			Description:    "Increasing focus on sustainable transportation",
			KeyDrivers:     []string{"Environmental concerns", "Regulatory pressure", "Consumer demand"},
			RiskMitigation: []string{"Invest in clean technology", "Reduce emissions"},
		},
	}
}

// generateTransportationOperationalFactors generates transportation operational factors
func (tm *TransportationModel) generateTransportationOperationalFactors(business *models.RiskAssessmentRequest) []OperationalFactor {
	return []OperationalFactor{
		{
			FactorID:            "transportation_fleet_management",
			FactorName:          "Fleet Management",
			OperationalArea:     "fleet",
			RiskScore:           0.3,
			Criticality:         "critical",
			Description:         "Management of transportation fleet",
			ControlMeasures:     []string{"Fleet maintenance", "Driver management", "Route optimization"},
			MonitoringFrequency: "daily",
		},
		{
			FactorID:            "transportation_safety_management",
			FactorName:          "Safety Management",
			OperationalArea:     "safety",
			RiskScore:           0.25,
			Criticality:         "critical",
			Description:         "Management of transportation safety",
			ControlMeasures:     []string{"Safety programs", "Training", "Inspections"},
			MonitoringFrequency: "daily",
		},
		{
			FactorID:            "transportation_logistics_management",
			FactorName:          "Logistics Management",
			OperationalArea:     "logistics",
			RiskScore:           0.2,
			Criticality:         "high",
			Description:         "Management of transportation logistics",
			ControlMeasures:     []string{"Route optimization", "Load planning", "Delivery tracking"},
			MonitoringFrequency: "continuous",
		},
	}
}

// calculateOverallIndustryRisk calculates the overall industry risk score
func (tm *TransportationModel) calculateOverallIndustryRisk(baseScore float64, factors []IndustryRiskFactor, compliance []ComplianceStatus) float64 {
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
func (tm *TransportationModel) determineRiskLevel(score float64) models.RiskLevel {
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
func (tm *TransportationModel) calculateRiskLevel(score float64) string {
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
func (tm *TransportationModel) calculateConfidenceScore(business *models.RiskAssessmentRequest, factors []IndustryRiskFactor) float64 {
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
