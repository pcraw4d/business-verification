package industry

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// FintechModel implements industry-specific risk analysis for fintech companies
type FintechModel struct {
	logger *zap.Logger
}

// NewFintechModel creates a new fintech industry model
func NewFintechModel(logger *zap.Logger) *FintechModel {
	return &FintechModel{
		logger: logger,
	}
}

// GetIndustryType returns the industry type
func (fm *FintechModel) GetIndustryType() IndustryType {
	return IndustryFintech
}

// CalculateIndustryRisk calculates fintech-specific risk factors
func (fm *FintechModel) CalculateIndustryRisk(ctx context.Context, business *models.RiskAssessmentRequest) (*IndustryRiskResult, error) {
	fm.logger.Info("Calculating fintech industry risk", zap.String("business", business.BusinessName))

	// Calculate base industry risk score
	baseScore := fm.calculateBaseFintechRisk(business)

	// Generate industry-specific factors
	industryFactors := fm.generateFintechRiskFactors(business, baseScore)

	// Calculate compliance status
	complianceStatus := fm.assessFintechCompliance(business)

	// Generate recommendations
	recommendations := fm.generateFintechRecommendations(business, baseScore)

	// Generate regulatory factors
	regulatoryFactors := fm.generateFintechRegulatoryFactors()

	// Generate market factors
	marketFactors := fm.generateFintechMarketFactors()

	// Generate operational factors
	operationalFactors := fm.generateFintechOperationalFactors(business)

	// Calculate overall industry risk score
	industryRiskScore := fm.calculateOverallIndustryRisk(baseScore, industryFactors, complianceStatus)

	// Determine risk level
	riskLevel := fm.determineRiskLevel(industryRiskScore)

	// Calculate confidence score
	confidenceScore := fm.calculateConfidenceScore(business, industryFactors)

	result := &IndustryRiskResult{
		IndustryType:            IndustryFintech,
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

	fm.logger.Info("Fintech industry risk calculated",
		zap.Float64("risk_score", industryRiskScore),
		zap.String("risk_level", string(riskLevel)))

	return result, nil
}

// GetIndustrySpecificFactors returns fintech-specific risk factors
func (fm *FintechModel) GetIndustrySpecificFactors() []IndustryRiskFactor {
	return []IndustryRiskFactor{
		{
			FactorID:            "fintech_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			Description:         "Compliance with financial regulations and licensing requirements",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "fintech_cybersecurity",
			FactorName:          "Cybersecurity Risk",
			FactorCategory:      "operational",
			Description:         "Risk of cyber attacks and data breaches",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "fintech_anti_money_laundering",
			FactorName:          "Anti-Money Laundering",
			FactorCategory:      "compliance",
			Description:         "AML compliance and transaction monitoring",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "fintech_capital_requirements",
			FactorName:          "Capital Requirements",
			FactorCategory:      "financial",
			Description:         "Adequacy of capital reserves and liquidity",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "fintech_third_party_risk",
			FactorName:          "Third-Party Risk",
			FactorCategory:      "operational",
			Description:         "Risk from third-party service providers and partnerships",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "fintech_technology_risk",
			FactorName:          "Technology Risk",
			FactorCategory:      "operational",
			Description:         "Risk from technology failures and system outages",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "fintech_operational_resilience",
			FactorName:          "Operational Resilience",
			FactorCategory:      "operational",
			Description:         "Ability to maintain operations during disruptions",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "fintech_data_protection",
			FactorName:          "Data Protection",
			FactorCategory:      "compliance",
			Description:         "Compliance with data protection and privacy regulations",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
	}
}

// GetIndustryWeightings returns risk category weightings for fintech
func (fm *FintechModel) GetIndustryWeightings() map[string]float64 {
	return map[string]float64{
		"regulatory":    0.25, // High regulatory risk
		"compliance":    0.20, // High compliance requirements
		"operational":   0.20, // High operational risk
		"financial":     0.15, // Moderate financial risk
		"reputational":  0.10, // Moderate reputational risk
		"technology":    0.05, // Technology risk
		"geopolitical":  0.03, // Low geopolitical risk
		"environmental": 0.02, // Very low environmental risk
	}
}

// ValidateIndustryData validates fintech-specific business data
func (fm *FintechModel) ValidateIndustryData(business *models.RiskAssessmentRequest) []string {
	var errors []string

	if business == nil {
		errors = append(errors, "business information is required")
		return errors
	}

	// Check for required fintech-specific fields
	if business.BusinessName == "" {
		errors = append(errors, "business name is required")
	}

	// Validate business address
	if business.BusinessAddress == "" {
		errors = append(errors, "business address is required for regulatory compliance")
	}

	// Check for fintech-specific metadata
	if business.Metadata != nil {
		// Check for license information
		if _, hasLicense := business.Metadata["license_number"]; !hasLicense {
			errors = append(errors, "financial services license number is recommended")
		}

		// Check for regulatory body
		if _, hasRegulator := business.Metadata["regulatory_body"]; !hasRegulator {
			errors = append(errors, "regulatory body information is recommended")
		}
	}

	return errors
}

// GetIndustryComplianceRequirements returns fintech compliance requirements
func (fm *FintechModel) GetIndustryComplianceRequirements() []ComplianceRequirement {
	return []ComplianceRequirement{
		{
			RequirementID:   "fintech_licensing",
			RequirementName: "Financial Services License",
			RegulatoryBody:  "Financial Conduct Authority (FCA)",
			Jurisdiction:    "UK",
			Description:     "Required license for providing financial services",
			Required:        true,
			PenaltyAmount:   "Up to £50,000 per violation",
			ComplianceSteps: []string{
				"Obtain appropriate financial services license",
				"Maintain license in good standing",
				"Submit regular regulatory reports",
			},
			Documentation: []string{
				"License certificate",
				"Regulatory filings",
				"Compliance reports",
			},
		},
		{
			RequirementID:   "fintech_aml",
			RequirementName: "Anti-Money Laundering (AML)",
			RegulatoryBody:  "Financial Action Task Force (FATF)",
			Jurisdiction:    "Global",
			Description:     "AML compliance and customer due diligence",
			Required:        true,
			PenaltyAmount:   "Up to $1M per violation",
			ComplianceSteps: []string{
				"Implement AML policies and procedures",
				"Conduct customer due diligence",
				"Monitor transactions for suspicious activity",
				"Report suspicious transactions",
			},
			Documentation: []string{
				"AML policy document",
				"Customer due diligence records",
				"Suspicious activity reports",
			},
		},
		{
			RequirementID:   "fintech_data_protection",
			RequirementName: "Data Protection (GDPR)",
			RegulatoryBody:  "Information Commissioner's Office (ICO)",
			Jurisdiction:    "EU/UK",
			Description:     "Data protection and privacy compliance",
			Required:        true,
			PenaltyAmount:   "Up to €20M or 4% of annual turnover",
			ComplianceSteps: []string{
				"Implement data protection policies",
				"Conduct data protection impact assessments",
				"Appoint data protection officer if required",
				"Implement privacy by design",
			},
			Documentation: []string{
				"Data protection policy",
				"Privacy notices",
				"Data processing agreements",
			},
		},
		{
			RequirementID:   "fintech_cybersecurity",
			RequirementName: "Cybersecurity Framework",
			RegulatoryBody:  "National Institute of Standards and Technology (NIST)",
			Jurisdiction:    "US",
			Description:     "Cybersecurity risk management framework",
			Required:        false,
			PenaltyAmount:   "Regulatory sanctions and fines",
			ComplianceSteps: []string{
				"Implement cybersecurity framework",
				"Conduct regular security assessments",
				"Maintain incident response plan",
				"Provide security training",
			},
			Documentation: []string{
				"Cybersecurity policy",
				"Security assessment reports",
				"Incident response plan",
			},
		},
	}
}

// calculateBaseFintechRisk calculates the base risk score for fintech companies
func (fm *FintechModel) calculateBaseFintechRisk(business *models.RiskAssessmentRequest) float64 {
	baseScore := 0.4 // Base fintech risk is higher due to regulatory complexity

	// Adjust based on business characteristics
	if business.Metadata != nil {
		// Check for license status
		if licenseStatus, exists := business.Metadata["license_status"]; exists {
			if status, ok := licenseStatus.(string); ok {
				switch status {
				case "active", "valid":
					baseScore -= 0.1
				case "pending", "under_review":
					baseScore += 0.1
				case "suspended", "revoked":
					baseScore += 0.3
				}
			}
		}

		// Check for regulatory body
		if _, hasRegulator := business.Metadata["regulatory_body"]; hasRegulator {
			baseScore -= 0.05 // Having a regulator reduces risk
		}

		// Check for compliance certifications
		if certifications, exists := business.Metadata["certifications"]; exists {
			if certs, ok := certifications.([]string); ok {
				baseScore -= float64(len(certs)) * 0.02 // Each certification reduces risk
			}
		}
	}

	// Ensure score is within bounds
	return math.Max(0.0, math.Min(1.0, baseScore))
}

// generateFintechRiskFactors generates fintech-specific risk factors
func (fm *FintechModel) generateFintechRiskFactors(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRiskFactor {
	factors := []IndustryRiskFactor{
		{
			FactorID:            "fintech_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			RiskScore:           baseScore + 0.1,
			RiskLevel:           fm.calculateRiskLevel(baseScore + 0.1),
			Description:         "Compliance with financial regulations and licensing requirements",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Maintain active licenses, implement compliance monitoring, conduct regular audits",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "fintech_cybersecurity",
			FactorName:          "Cybersecurity Risk",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.15,
			RiskLevel:           fm.calculateRiskLevel(baseScore + 0.15),
			Description:         "Risk of cyber attacks and data breaches",
			Impact:              "high",
			Likelihood:          "high",
			MitigationAdvice:    "Implement multi-layered security, conduct penetration testing, maintain incident response plan",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "fintech_anti_money_laundering",
			FactorName:          "Anti-Money Laundering",
			FactorCategory:      "compliance",
			RiskScore:           baseScore + 0.08,
			RiskLevel:           fm.calculateRiskLevel(baseScore + 0.08),
			Description:         "AML compliance and transaction monitoring",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement robust AML systems, conduct customer due diligence, monitor transactions",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "fintech_capital_requirements",
			FactorName:          "Capital Requirements",
			FactorCategory:      "financial",
			RiskScore:           baseScore + 0.05,
			RiskLevel:           fm.calculateRiskLevel(baseScore + 0.05),
			Description:         "Adequacy of capital reserves and liquidity",
			Impact:              "high",
			Likelihood:          "low",
			MitigationAdvice:    "Maintain adequate capital reserves, implement liquidity management, stress test capital adequacy",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "fintech_third_party_risk",
			FactorName:          "Third-Party Risk",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.12,
			RiskLevel:           fm.calculateRiskLevel(baseScore + 0.12),
			Description:         "Risk from third-party service providers and partnerships",
			Impact:              "medium",
			Likelihood:          "medium",
			MitigationAdvice:    "Conduct due diligence on partners, implement vendor risk management, monitor third-party performance",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "fintech_technology_risk",
			FactorName:          "Technology Risk",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.10,
			RiskLevel:           fm.calculateRiskLevel(baseScore + 0.10),
			Description:         "Risk from technology failures and system outages",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement redundancy, conduct disaster recovery testing, maintain system monitoring",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "fintech_operational_resilience",
			FactorName:          "Operational Resilience",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.07,
			RiskLevel:           fm.calculateRiskLevel(baseScore + 0.07),
			Description:         "Ability to maintain operations during disruptions",
			Impact:              "high",
			Likelihood:          "low",
			MitigationAdvice:    "Develop business continuity plans, implement backup systems, conduct resilience testing",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "fintech_data_protection",
			FactorName:          "Data Protection",
			FactorCategory:      "compliance",
			RiskScore:           baseScore + 0.09,
			RiskLevel:           fm.calculateRiskLevel(baseScore + 0.09),
			Description:         "Compliance with data protection and privacy regulations",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement data protection policies, conduct privacy impact assessments, ensure data minimization",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
	}

	return factors
}

// assessFintechCompliance assesses compliance status for fintech requirements
func (fm *FintechModel) assessFintechCompliance(business *models.RiskAssessmentRequest) []ComplianceStatus {
	statuses := []ComplianceStatus{
		{
			RequirementID:   "fintech_licensing",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"License status not verified"},
			Recommendations: []string{"Verify financial services license status"},
		},
		{
			RequirementID:   "fintech_aml",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"AML compliance not verified"},
			Recommendations: []string{"Implement AML compliance program"},
		},
		{
			RequirementID:   "fintech_data_protection",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"Data protection compliance not verified"},
			Recommendations: []string{"Implement data protection framework"},
		},
		{
			RequirementID:   "fintech_cybersecurity",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"Cybersecurity framework not verified"},
			Recommendations: []string{"Implement cybersecurity framework"},
		},
	}

	// Adjust based on business metadata
	if business.Metadata != nil {
		if licenseStatus, exists := business.Metadata["license_status"]; exists {
			if status, ok := licenseStatus.(string); ok {
				for i := range statuses {
					if statuses[i].RequirementID == "fintech_licensing" {
						switch status {
						case "active", "valid":
							statuses[i].Status = "compliant"
							statuses[i].ComplianceScore = 0.9
							statuses[i].Issues = []string{}
							statuses[i].Recommendations = []string{"Maintain license compliance"}
						case "pending", "under_review":
							statuses[i].Status = "non_compliant"
							statuses[i].ComplianceScore = 0.3
							statuses[i].Issues = []string{"License under review"}
							statuses[i].Recommendations = []string{"Complete license review process"}
						case "suspended", "revoked":
							statuses[i].Status = "non_compliant"
							statuses[i].ComplianceScore = 0.1
							statuses[i].Issues = []string{"License suspended or revoked"}
							statuses[i].Recommendations = []string{"Address license issues immediately"}
						}
					}
				}
			}
		}
	}

	return statuses
}

// generateFintechRecommendations generates fintech-specific recommendations
func (fm *FintechModel) generateFintechRecommendations(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRecommendation {
	recommendations := []IndustryRecommendation{
		{
			RecommendationID:   "fintech_regulatory_compliance",
			Category:           "regulatory",
			Priority:           "high",
			Title:              "Strengthen Regulatory Compliance",
			Description:        "Implement comprehensive regulatory compliance framework",
			ActionItems:        []string{"Conduct compliance audit", "Implement compliance monitoring", "Train staff on regulations"},
			ExpectedBenefit:    "Reduced regulatory risk and penalties",
			ImplementationCost: "Medium",
			Timeline:           "3-6 months",
		},
		{
			RecommendationID:   "fintech_cybersecurity",
			Category:           "operational",
			Priority:           "high",
			Title:              "Enhance Cybersecurity",
			Description:        "Implement robust cybersecurity measures",
			ActionItems:        []string{"Conduct security assessment", "Implement multi-factor authentication", "Deploy intrusion detection"},
			ExpectedBenefit:    "Reduced cyber attack risk",
			ImplementationCost: "High",
			Timeline:           "6-12 months",
		},
		{
			RecommendationID:   "fintech_aml_program",
			Category:           "compliance",
			Priority:           "high",
			Title:              "Implement AML Program",
			Description:        "Develop comprehensive anti-money laundering program",
			ActionItems:        []string{"Develop AML policies", "Implement transaction monitoring", "Train staff on AML"},
			ExpectedBenefit:    "Compliance with AML regulations",
			ImplementationCost: "Medium",
			Timeline:           "3-6 months",
		},
		{
			RecommendationID:   "fintech_operational_resilience",
			Category:           "operational",
			Priority:           "medium",
			Title:              "Improve Operational Resilience",
			Description:        "Enhance business continuity and disaster recovery",
			ActionItems:        []string{"Develop business continuity plan", "Implement backup systems", "Conduct disaster recovery testing"},
			ExpectedBenefit:    "Improved operational stability",
			ImplementationCost: "Medium",
			Timeline:           "6-12 months",
		},
	}

	return recommendations
}

// generateFintechRegulatoryFactors generates fintech regulatory factors
func (fm *FintechModel) generateFintechRegulatoryFactors() []RegulatoryFactor {
	return []RegulatoryFactor{
		{
			FactorID:       "fintech_psd2",
			RegulationName: "Payment Services Directive 2 (PSD2)",
			RegulatoryBody: "European Banking Authority",
			Jurisdiction:   "EU",
			RiskImpact:     0.3,
			ComplianceCost: "Medium",
			PenaltyRisk:    "High",
			Description:    "Open banking and payment services regulation",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "fintech_gdpr",
			RegulationName: "General Data Protection Regulation (GDPR)",
			RegulatoryBody: "European Commission",
			Jurisdiction:   "EU",
			RiskImpact:     0.4,
			ComplianceCost: "High",
			PenaltyRisk:    "Very High",
			Description:    "Data protection and privacy regulation",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "fintech_fca",
			RegulationName: "FCA Regulatory Framework",
			RegulatoryBody: "Financial Conduct Authority",
			Jurisdiction:   "UK",
			RiskImpact:     0.35,
			ComplianceCost: "High",
			PenaltyRisk:    "High",
			Description:    "UK financial services regulation",
			LastUpdated:    time.Now(),
		},
	}
}

// generateFintechMarketFactors generates fintech market factors
func (fm *FintechModel) generateFintechMarketFactors() []MarketFactor {
	return []MarketFactor{
		{
			FactorID:       "fintech_digital_adoption",
			FactorName:     "Digital Banking Adoption",
			MarketTrend:    "growing",
			ImpactScore:    0.7,
			TimeHorizon:    "long-term",
			Description:    "Increasing adoption of digital banking services",
			KeyDrivers:     []string{"Consumer preference", "Technology advancement", "COVID-19 impact"},
			RiskMitigation: []string{"Invest in digital capabilities", "Focus on user experience"},
		},
		{
			FactorID:       "fintech_competition",
			FactorName:     "Market Competition",
			MarketTrend:    "volatile",
			ImpactScore:    0.6,
			TimeHorizon:    "medium-term",
			Description:    "Intense competition in fintech space",
			KeyDrivers:     []string{"New entrants", "Established players", "Technology disruption"},
			RiskMitigation: []string{"Differentiate offerings", "Focus on niche markets"},
		},
		{
			FactorID:       "fintech_regulatory_evolution",
			FactorName:     "Regulatory Evolution",
			MarketTrend:    "stable",
			ImpactScore:    0.5,
			TimeHorizon:    "long-term",
			Description:    "Evolving regulatory landscape",
			KeyDrivers:     []string{"Regulatory sandboxes", "New regulations", "International harmonization"},
			RiskMitigation: []string{"Monitor regulatory changes", "Engage with regulators"},
		},
	}
}

// generateFintechOperationalFactors generates fintech operational factors
func (fm *FintechModel) generateFintechOperationalFactors(business *models.RiskAssessmentRequest) []OperationalFactor {
	return []OperationalFactor{
		{
			FactorID:            "fintech_system_reliability",
			FactorName:          "System Reliability",
			OperationalArea:     "technology",
			RiskScore:           0.3,
			Criticality:         "critical",
			Description:         "Reliability of core financial systems",
			ControlMeasures:     []string{"Redundancy", "Monitoring", "Testing"},
			MonitoringFrequency: "continuous",
		},
		{
			FactorID:            "fintech_data_quality",
			FactorName:          "Data Quality",
			OperationalArea:     "data_management",
			RiskScore:           0.25,
			Criticality:         "high",
			Description:         "Quality and accuracy of financial data",
			ControlMeasures:     []string{"Data validation", "Quality checks", "Audit trails"},
			MonitoringFrequency: "daily",
		},
		{
			FactorID:            "fintech_customer_support",
			FactorName:          "Customer Support",
			OperationalArea:     "customer_service",
			RiskScore:           0.2,
			Criticality:         "medium",
			Description:         "Quality of customer support services",
			ControlMeasures:     []string{"Training", "SLA monitoring", "Feedback systems"},
			MonitoringFrequency: "weekly",
		},
	}
}

// calculateOverallIndustryRisk calculates the overall industry risk score
func (fm *FintechModel) calculateOverallIndustryRisk(baseScore float64, factors []IndustryRiskFactor, compliance []ComplianceStatus) float64 {
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
func (fm *FintechModel) determineRiskLevel(score float64) models.RiskLevel {
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
func (fm *FintechModel) calculateRiskLevel(score float64) string {
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
func (fm *FintechModel) calculateConfidenceScore(business *models.RiskAssessmentRequest, factors []IndustryRiskFactor) float64 {
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
