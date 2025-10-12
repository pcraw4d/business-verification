package industry

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// TechnologyModel implements industry-specific risk analysis for technology companies
type TechnologyModel struct {
	logger *zap.Logger
}

// NewTechnologyModel creates a new technology industry model
func NewTechnologyModel(logger *zap.Logger) *TechnologyModel {
	return &TechnologyModel{
		logger: logger,
	}
}

// GetIndustryType returns the industry type
func (tm *TechnologyModel) GetIndustryType() IndustryType {
	return IndustryTechnology
}

// CalculateIndustryRisk calculates technology-specific risk factors
func (tm *TechnologyModel) CalculateIndustryRisk(ctx context.Context, business *models.RiskAssessmentRequest) (*IndustryRiskResult, error) {
	tm.logger.Info("Calculating technology industry risk", zap.String("business", business.BusinessName))

	// Calculate base industry risk score
	baseScore := tm.calculateBaseTechnologyRisk(business)

	// Generate industry-specific factors
	industryFactors := tm.generateTechnologyRiskFactors(business, baseScore)

	// Calculate compliance status
	complianceStatus := tm.assessTechnologyCompliance(business)

	// Generate recommendations
	recommendations := tm.generateTechnologyRecommendations(business, baseScore)

	// Generate regulatory factors
	regulatoryFactors := tm.generateTechnologyRegulatoryFactors()

	// Generate market factors
	marketFactors := tm.generateTechnologyMarketFactors()

	// Generate operational factors
	operationalFactors := tm.generateTechnologyOperationalFactors(business)

	// Calculate overall industry risk score
	industryRiskScore := tm.calculateOverallIndustryRisk(baseScore, industryFactors, complianceStatus)

	// Determine risk level
	riskLevel := tm.determineRiskLevel(industryRiskScore)

	// Calculate confidence score
	confidenceScore := tm.calculateConfidenceScore(business, industryFactors)

	result := &IndustryRiskResult{
		IndustryType:            IndustryTechnology,
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

	tm.logger.Info("Technology industry risk calculated",
		zap.Float64("risk_score", industryRiskScore),
		zap.String("risk_level", string(riskLevel)))

	return result, nil
}

// GetIndustrySpecificFactors returns technology-specific risk factors
func (tm *TechnologyModel) GetIndustrySpecificFactors() []IndustryRiskFactor {
	return []IndustryRiskFactor{
		{
			FactorID:           "technology_cybersecurity",
			FactorName:         "Cybersecurity Risk",
			FactorCategory:     "operational",
			Description:        "Risk of cyber attacks and data breaches",
			IndustrySpecific:   true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:           "technology_intellectual_property",
			FactorName:         "Intellectual Property",
			FactorCategory:     "legal",
			Description:        "Protection and enforcement of intellectual property rights",
			IndustrySpecific:   true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:           "technology_data_privacy",
			FactorName:         "Data Privacy",
			FactorCategory:     "compliance",
			Description:        "Compliance with data protection regulations",
			IndustrySpecific:   true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:           "technology_rapid_innovation",
			FactorName:         "Rapid Innovation",
			FactorCategory:     "operational",
			Description:        "Risk from rapid technological change and disruption",
			IndustrySpecific:   true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:           "technology_talent_retention",
			FactorName:         "Talent Retention",
			FactorCategory:     "operational",
			Description:        "Risk of losing key technical talent",
			IndustrySpecific:   true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:           "technology_platform_dependency",
			FactorName:         "Platform Dependency",
			FactorCategory:     "operational",
			Description:        "Dependency on third-party platforms and services",
			IndustrySpecific:   true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:           "technology_scalability",
			FactorName:         "Scalability Risk",
			FactorCategory:     "operational",
			Description:        "Ability to scale technology infrastructure",
			IndustrySpecific:   true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:           "technology_regulatory_evolution",
			FactorName:         "Regulatory Evolution",
			FactorCategory:     "regulatory",
			Description:        "Risk from evolving technology regulations",
			IndustrySpecific:   true,
			RegulatoryRelevance: true,
		},
	}
}

// GetIndustryWeightings returns risk category weightings for technology
func (tm *TechnologyModel) GetIndustryWeightings() map[string]float64 {
	return map[string]float64{
		"regulatory":   0.15, // Moderate regulatory risk
		"compliance":   0.15, // Moderate compliance requirements
		"operational":  0.30, // Very high operational risk
		"financial":    0.10, // Low financial risk
		"reputational": 0.15, // Moderate reputational risk
		"technology":   0.10, // Technology risk
		"geopolitical": 0.03, // Low geopolitical risk
		"environmental": 0.02, // Very low environmental risk
	}
}

// ValidateIndustryData validates technology-specific business data
func (tm *TechnologyModel) ValidateIndustryData(business *models.RiskAssessmentRequest) []string {
	var errors []string

	if business == nil {
		errors = append(errors, "business information is required")
		return errors
	}

	// Check for required technology-specific fields
	if business.BusinessName == "" {
		errors = append(errors, "business name is required")
	}

	// Check for technology-specific metadata
	if business.Metadata != nil {
		// Check for technology stack information
		if _, hasTechStack := business.Metadata["technology_stack"]; !hasTechStack {
			errors = append(errors, "technology stack information is recommended")
		}

		// Check for security certifications
		if _, hasSecurityCert := business.Metadata["security_certifications"]; !hasSecurityCert {
			errors = append(errors, "security certifications information is recommended")
		}
	}

	return errors
}

// GetIndustryComplianceRequirements returns technology compliance requirements
func (tm *TechnologyModel) GetIndustryComplianceRequirements() []ComplianceRequirement {
	return []ComplianceRequirement{
		{
			RequirementID:   "technology_gdpr",
			RequirementName: "GDPR Compliance",
			RegulatoryBody:  "European Commission",
			Jurisdiction:    "EU",
			Description:     "General Data Protection Regulation compliance",
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
			RequirementID:   "technology_ccpa",
			RequirementName: "CCPA Compliance",
			RegulatoryBody:  "California Attorney General",
			Jurisdiction:    "California, US",
			Description:     "California Consumer Privacy Act compliance",
			Required:        false,
			PenaltyAmount:   "Up to $7,500 per violation",
			ComplianceSteps: []string{
				"Implement privacy policies",
				"Provide consumer rights",
				"Implement data deletion procedures",
				"Conduct privacy assessments",
			},
			Documentation: []string{
				"Privacy policy",
				"Consumer rights procedures",
				"Data deletion procedures",
			},
		},
		{
			RequirementID:   "technology_sox",
			RequirementName: "SOX Compliance",
			RegulatoryBody:  "Securities and Exchange Commission",
			Jurisdiction:    "US",
			Description:     "Sarbanes-Oxley Act compliance for public companies",
			Required:        false,
			PenaltyAmount:   "Criminal penalties and fines",
			ComplianceSteps: []string{
				"Implement internal controls",
				"Conduct financial audits",
				"Maintain audit trails",
				"Implement whistleblower procedures",
			},
			Documentation: []string{
				"Internal control documentation",
				"Audit reports",
				"Whistleblower procedures",
			},
		},
		{
			RequirementID:   "technology_iso27001",
			RequirementName: "ISO 27001 Certification",
			RegulatoryBody:  "International Organization for Standardization",
			Jurisdiction:    "Global",
			Description:     "Information security management system certification",
			Required:        false,
			PenaltyAmount:   "Loss of certification",
			ComplianceSteps: []string{
				"Implement ISMS",
				"Conduct risk assessments",
				"Implement security controls",
				"Conduct regular audits",
			},
			Documentation: []string{
				"ISMS documentation",
				"Risk assessment reports",
				"Security control documentation",
			},
		},
	}
}

// calculateBaseTechnologyRisk calculates the base risk score for technology companies
func (tm *TechnologyModel) calculateBaseTechnologyRisk(business *models.RiskAssessmentRequest) float64 {
	baseScore := 0.3 // Base technology risk is moderate

	// Adjust based on business characteristics
	if business.Metadata != nil {
		// Check for security certifications
		if certifications, exists := business.Metadata["security_certifications"]; exists {
			if certs, ok := certifications.([]string); ok {
				baseScore -= float64(len(certs)) * 0.03 // Each certification reduces risk
			}
		}

		// Check for technology maturity
		if maturity, exists := business.Metadata["technology_maturity"]; exists {
			if mat, ok := maturity.(string); ok {
				switch mat {
				case "enterprise", "mature":
					baseScore -= 0.05
				case "startup", "early_stage":
					baseScore += 0.1
				case "scale_up", "growth":
					baseScore += 0.05
				}
			}
		}

		// Check for funding status
		if funding, exists := business.Metadata["funding_status"]; exists {
			if fund, ok := funding.(string); ok {
				switch fund {
				case "profitable", "well_funded":
					baseScore -= 0.05
				case "bootstrapped", "self_funded":
					baseScore += 0.05
				case "pre_revenue", "early_stage":
					baseScore += 0.1
				}
			}
		}
	}

	// Ensure score is within bounds
	return math.Max(0.0, math.Min(1.0, baseScore))
}

// generateTechnologyRiskFactors generates technology-specific risk factors
func (tm *TechnologyModel) generateTechnologyRiskFactors(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRiskFactor {
	factors := []IndustryRiskFactor{
		{
			FactorID:          "technology_cybersecurity",
			FactorName:        "Cybersecurity Risk",
			FactorCategory:    "operational",
			RiskScore:         baseScore + 0.15,
			RiskLevel:         tm.calculateRiskLevel(baseScore + 0.15),
			Description:       "Risk of cyber attacks and data breaches",
			Impact:            "high",
			Likelihood:        "high",
			MitigationAdvice:  "Implement multi-layered security, conduct penetration testing, maintain incident response plan",
			IndustrySpecific:  true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:          "technology_intellectual_property",
			FactorName:        "Intellectual Property",
			FactorCategory:    "legal",
			RiskScore:         baseScore + 0.10,
			RiskLevel:         tm.calculateRiskLevel(baseScore + 0.10),
			Description:       "Protection and enforcement of intellectual property rights",
			Impact:            "high",
			Likelihood:        "medium",
			MitigationAdvice:  "File patents, implement IP protection policies, monitor for infringement",
			IndustrySpecific:  true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:          "technology_data_privacy",
			FactorName:        "Data Privacy",
			FactorCategory:    "compliance",
			RiskScore:         baseScore + 0.12,
			RiskLevel:         tm.calculateRiskLevel(baseScore + 0.12),
			Description:       "Compliance with data protection regulations",
			Impact:            "high",
			Likelihood:        "medium",
			MitigationAdvice:  "Implement privacy by design, conduct privacy impact assessments, maintain data governance",
			IndustrySpecific:  true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:          "technology_rapid_innovation",
			FactorName:        "Rapid Innovation",
			FactorCategory:    "operational",
			RiskScore:         baseScore + 0.08,
			RiskLevel:         tm.calculateRiskLevel(baseScore + 0.08),
			Description:       "Risk from rapid technological change and disruption",
			Impact:            "medium",
			Likelihood:        "high",
			MitigationAdvice:  "Invest in R&D, monitor technology trends, maintain agile development",
			IndustrySpecific:  true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:          "technology_talent_retention",
			FactorName:        "Talent Retention",
			FactorCategory:    "operational",
			RiskScore:         baseScore + 0.13,
			RiskLevel:         tm.calculateRiskLevel(baseScore + 0.13),
			Description:       "Risk of losing key technical talent",
			Impact:            "high",
			Likelihood:        "medium",
			MitigationAdvice:  "Offer competitive compensation, provide growth opportunities, maintain company culture",
			IndustrySpecific:  true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:          "technology_platform_dependency",
			FactorName:        "Platform Dependency",
			FactorCategory:    "operational",
			RiskScore:         baseScore + 0.07,
			RiskLevel:         tm.calculateRiskLevel(baseScore + 0.07),
			Description:       "Dependency on third-party platforms and services",
			Impact:            "medium",
			Likelihood:        "medium",
			MitigationAdvice:  "Diversify platform dependencies, implement fallback systems, negotiate SLAs",
			IndustrySpecific:  true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:          "technology_scalability",
			FactorName:        "Scalability Risk",
			FactorCategory:    "operational",
			RiskScore:         baseScore + 0.09,
			RiskLevel:         tm.calculateRiskLevel(baseScore + 0.09),
			Description:       "Ability to scale technology infrastructure",
			Impact:            "high",
			Likelihood:        "medium",
			MitigationAdvice:  "Design for scale, implement auto-scaling, monitor performance metrics",
			IndustrySpecific:  true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:          "technology_regulatory_evolution",
			FactorName:        "Regulatory Evolution",
			FactorCategory:    "regulatory",
			RiskScore:         baseScore + 0.06,
			RiskLevel:         tm.calculateRiskLevel(baseScore + 0.06),
			Description:       "Risk from evolving technology regulations",
			Impact:            "medium",
			Likelihood:        "medium",
			MitigationAdvice:  "Monitor regulatory changes, engage with policymakers, implement compliance frameworks",
			IndustrySpecific:  true,
			RegulatoryRelevance: true,
		},
	}

	return factors
}

// assessTechnologyCompliance assesses compliance status for technology requirements
func (tm *TechnologyModel) assessTechnologyCompliance(business *models.RiskAssessmentRequest) []ComplianceStatus {
	statuses := []ComplianceStatus{
		{
			RequirementID:   "technology_gdpr",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"GDPR compliance not verified"},
			Recommendations: []string{"Implement GDPR compliance program"},
		},
		{
			RequirementID:   "technology_ccpa",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"CCPA compliance not verified"},
			Recommendations: []string{"Implement CCPA compliance program"},
		},
		{
			RequirementID:   "technology_sox",
			Status:          "not_applicable",
			LastChecked:     time.Now(),
			ComplianceScore: 0.8,
			Issues:          []string{},
			Recommendations: []string{"Monitor SOX requirements if going public"},
		},
		{
			RequirementID:   "technology_iso27001",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"ISO 27001 certification not verified"},
			Recommendations: []string{"Consider ISO 27001 certification"},
		},
	}

	// Adjust based on business metadata
	if business.Metadata != nil {
		if certifications, exists := business.Metadata["security_certifications"]; exists {
			if certs, ok := certifications.([]string); ok {
				for _, cert := range certs {
					for i := range statuses {
						if statuses[i].RequirementID == "technology_iso27001" && cert == "iso27001" {
							statuses[i].Status = "compliant"
							statuses[i].ComplianceScore = 0.9
							statuses[i].Issues = []string{}
							statuses[i].Recommendations = []string{"Maintain ISO 27001 certification"}
						}
					}
				}
			}
		}
	}

	return statuses
}

// generateTechnologyRecommendations generates technology-specific recommendations
func (tm *TechnologyModel) generateTechnologyRecommendations(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRecommendation {
	recommendations := []IndustryRecommendation{
		{
			RecommendationID:    "technology_cybersecurity",
			Category:            "operational",
			Priority:            "high",
			Title:               "Strengthen Cybersecurity",
			Description:         "Implement comprehensive cybersecurity program",
			ActionItems:         []string{"Conduct security assessment", "Implement multi-factor authentication", "Deploy security monitoring"},
			ExpectedBenefit:     "Reduced cyber attack risk",
			ImplementationCost:  "High",
			Timeline:            "6-12 months",
		},
		{
			RecommendationID:    "technology_data_privacy",
			Category:            "compliance",
			Priority:            "high",
			Title:               "Implement Data Privacy Program",
			Description:         "Develop comprehensive data privacy compliance program",
			ActionItems:         []string{"Conduct privacy impact assessment", "Implement privacy by design", "Train staff on privacy"},
			ExpectedBenefit:     "GDPR/CCPA compliance and reduced penalties",
			ImplementationCost:  "Medium",
			Timeline:            "3-6 months",
		},
		{
			RecommendationID:    "technology_talent_retention",
			Category:            "operational",
			Priority:            "medium",
			Title:               "Improve Talent Retention",
			Description:         "Develop talent retention strategies",
			ActionItems:         []string{"Conduct employee surveys", "Implement retention programs", "Offer competitive benefits"},
			ExpectedBenefit:     "Reduced talent turnover",
			ImplementationCost:  "Medium",
			Timeline:            "6-12 months",
		},
		{
			RecommendationID:    "technology_scalability",
			Category:            "operational",
			Priority:            "medium",
			Title:               "Improve Scalability",
			Description:         "Enhance technology infrastructure scalability",
			ActionItems:         []string{"Implement auto-scaling", "Optimize database performance", "Implement monitoring"},
			ExpectedBenefit:     "Improved system performance and reliability",
			ImplementationCost:  "High",
			Timeline:            "6-12 months",
		},
	}

	return recommendations
}

// generateTechnologyRegulatoryFactors generates technology regulatory factors
func (tm *TechnologyModel) generateTechnologyRegulatoryFactors() []RegulatoryFactor {
	return []RegulatoryFactor{
		{
			FactorID:       "technology_gdpr",
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
			FactorID:       "technology_ccpa",
			RegulationName: "California Consumer Privacy Act (CCPA)",
			RegulatoryBody: "California Attorney General",
			Jurisdiction:   "California, US",
			RiskImpact:     0.3,
			ComplianceCost: "Medium",
			PenaltyRisk:    "High",
			Description:    "Consumer privacy rights regulation",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "technology_digital_services_act",
			RegulationName: "Digital Services Act (DSA)",
			RegulatoryBody: "European Commission",
			Jurisdiction:   "EU",
			RiskImpact:     0.25,
			ComplianceCost: "Medium",
			PenaltyRisk:    "High",
			Description:    "Digital services regulation",
			LastUpdated:    time.Now(),
		},
	}
}

// generateTechnologyMarketFactors generates technology market factors
func (tm *TechnologyModel) generateTechnologyMarketFactors() []MarketFactor {
	return []MarketFactor{
		{
			FactorID:       "technology_ai_disruption",
			FactorName:     "AI Disruption",
			MarketTrend:    "volatile",
			ImpactScore:    0.8,
			TimeHorizon:    "medium-term",
			Description:    "Artificial intelligence disruption across industries",
			KeyDrivers:     []string{"AI advancement", "Automation", "Machine learning"},
			RiskMitigation: []string{"Invest in AI capabilities", "Partner with AI companies"},
		},
		{
			FactorID:       "technology_cloud_adoption",
			FactorName:     "Cloud Adoption",
			MarketTrend:    "growing",
			ImpactScore:    0.6,
			TimeHorizon:    "long-term",
			Description:    "Increasing adoption of cloud computing",
			KeyDrivers:     []string{"Cost efficiency", "Scalability", "Remote work"},
			RiskMitigation: []string{"Develop cloud-native solutions", "Focus on cloud security"},
		},
		{
			FactorID:       "technology_competition",
			FactorName:     "Market Competition",
			MarketTrend:    "volatile",
			ImpactScore:    0.7,
			TimeHorizon:    "medium-term",
			Description:    "Intense competition in technology markets",
			KeyDrivers:     []string{"New entrants", "Established players", "Innovation"},
			RiskMitigation: []string{"Focus on differentiation", "Invest in innovation"},
		},
	}
}

// generateTechnologyOperationalFactors generates technology operational factors
func (tm *TechnologyModel) generateTechnologyOperationalFactors(business *models.RiskAssessmentRequest) []OperationalFactor {
	return []OperationalFactor{
		{
			FactorID:             "technology_system_reliability",
			FactorName:           "System Reliability",
			OperationalArea:      "infrastructure",
			RiskScore:            0.3,
			Criticality:          "critical",
			Description:          "Reliability of technology systems and infrastructure",
			ControlMeasures:      []string{"Redundancy", "Monitoring", "Testing"},
			MonitoringFrequency:  "continuous",
		},
		{
			FactorID:             "technology_data_quality",
			FactorName:           "Data Quality",
			OperationalArea:      "data_management",
			RiskScore:            0.25,
			Criticality:          "high",
			Description:          "Quality and accuracy of data",
			ControlMeasures:      []string{"Data validation", "Quality checks", "Audit trails"},
			MonitoringFrequency:  "daily",
		},
		{
			FactorID:             "technology_development_velocity",
			FactorName:           "Development Velocity",
			OperationalArea:      "software_development",
			RiskScore:            0.2,
			Criticality:          "medium",
			Description:          "Speed and efficiency of software development",
			ControlMeasures:      []string{"Agile methodologies", "Automation", "Team training"},
			MonitoringFrequency:  "weekly",
		},
	}
}

// calculateOverallIndustryRisk calculates the overall industry risk score
func (tm *TechnologyModel) calculateOverallIndustryRisk(baseScore float64, factors []IndustryRiskFactor, compliance []ComplianceStatus) float64 {
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
func (tm *TechnologyModel) determineRiskLevel(score float64) models.RiskLevel {
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
func (tm *TechnologyModel) calculateRiskLevel(score float64) string {
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
func (tm *TechnologyModel) calculateConfidenceScore(business *models.RiskAssessmentRequest, factors []IndustryRiskFactor) float64 {
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