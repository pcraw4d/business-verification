package industry

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// HealthcareModel implements industry-specific risk analysis for healthcare companies
type HealthcareModel struct {
	logger *zap.Logger
}

// NewHealthcareModel creates a new healthcare industry model
func NewHealthcareModel(logger *zap.Logger) *HealthcareModel {
	return &HealthcareModel{
		logger: logger,
	}
}

// GetIndustryType returns the industry type
func (hm *HealthcareModel) GetIndustryType() IndustryType {
	return IndustryHealthcare
}

// CalculateIndustryRisk calculates healthcare-specific risk factors
func (hm *HealthcareModel) CalculateIndustryRisk(ctx context.Context, business *models.RiskAssessmentRequest) (*IndustryRiskResult, error) {
	hm.logger.Info("Calculating healthcare industry risk", zap.String("business", business.BusinessName))

	// Calculate base industry risk score
	baseScore := hm.calculateBaseHealthcareRisk(business)

	// Generate industry-specific factors
	industryFactors := hm.generateHealthcareRiskFactors(business, baseScore)

	// Calculate compliance status
	complianceStatus := hm.assessHealthcareCompliance(business)

	// Generate recommendations
	recommendations := hm.generateHealthcareRecommendations(business, baseScore)

	// Generate regulatory factors
	regulatoryFactors := hm.generateHealthcareRegulatoryFactors()

	// Generate market factors
	marketFactors := hm.generateHealthcareMarketFactors()

	// Generate operational factors
	operationalFactors := hm.generateHealthcareOperationalFactors(business)

	// Calculate overall industry risk score
	industryRiskScore := hm.calculateOverallIndustryRisk(baseScore, industryFactors, complianceStatus)

	// Determine risk level
	riskLevel := hm.determineRiskLevel(industryRiskScore)

	// Calculate confidence score
	confidenceScore := hm.calculateConfidenceScore(business, industryFactors)

	result := &IndustryRiskResult{
		IndustryType:            IndustryHealthcare,
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

	hm.logger.Info("Healthcare industry risk calculated",
		zap.Float64("risk_score", industryRiskScore),
		zap.String("risk_level", string(riskLevel)))

	return result, nil
}

// GetIndustrySpecificFactors returns healthcare-specific risk factors
func (hm *HealthcareModel) GetIndustrySpecificFactors() []IndustryRiskFactor {
	return []IndustryRiskFactor{
		{
			FactorID:            "healthcare_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			Description:         "Compliance with healthcare regulations and licensing",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "healthcare_patient_safety",
			FactorName:          "Patient Safety",
			FactorCategory:      "operational",
			Description:         "Risk to patient safety and care quality",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "healthcare_data_privacy",
			FactorName:          "Data Privacy (HIPAA)",
			FactorCategory:      "compliance",
			Description:         "Compliance with healthcare data privacy regulations",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "healthcare_medical_liability",
			FactorName:          "Medical Liability",
			FactorCategory:      "financial",
			Description:         "Risk of medical malpractice and liability claims",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "healthcare_quality_assurance",
			FactorName:          "Quality Assurance",
			FactorCategory:      "operational",
			Description:         "Quality of care and service delivery",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "healthcare_workforce_shortage",
			FactorName:          "Workforce Shortage",
			FactorCategory:      "operational",
			Description:         "Shortage of qualified healthcare professionals",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "healthcare_technology_adoption",
			FactorName:          "Technology Adoption",
			FactorCategory:      "technology",
			Description:         "Adoption and integration of healthcare technology",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "healthcare_insurance_reimbursement",
			FactorName:          "Insurance Reimbursement",
			FactorCategory:      "financial",
			Description:         "Risk from insurance reimbursement changes",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
	}
}

// GetIndustryWeightings returns risk category weightings for healthcare
func (hm *HealthcareModel) GetIndustryWeightings() map[string]float64 {
	return map[string]float64{
		"regulatory":    0.20, // High regulatory risk
		"compliance":    0.20, // High compliance requirements
		"operational":   0.25, // Very high operational risk
		"financial":     0.15, // Moderate financial risk
		"reputational":  0.10, // Moderate reputational risk
		"technology":    0.05, // Technology risk
		"geopolitical":  0.03, // Low geopolitical risk
		"environmental": 0.02, // Very low environmental risk
	}
}

// ValidateIndustryData validates healthcare-specific business data
func (hm *HealthcareModel) ValidateIndustryData(business *models.RiskAssessmentRequest) []string {
	var errors []string

	if business == nil {
		errors = append(errors, "business information is required")
		return errors
	}

	// Check for required healthcare-specific fields
	if business.BusinessName == "" {
		errors = append(errors, "business name is required")
	}

	// Validate business address
	if business.BusinessAddress == "" {
		errors = append(errors, "business address is required for regulatory compliance")
	}

	// Check for healthcare-specific metadata
	if business.Metadata != nil {
		// Check for license information
		if _, hasLicense := business.Metadata["medical_license"]; !hasLicense {
			errors = append(errors, "medical license information is recommended")
		}

		// Check for accreditation
		if _, hasAccreditation := business.Metadata["accreditation"]; !hasAccreditation {
			errors = append(errors, "healthcare accreditation information is recommended")
		}
	}

	return errors
}

// GetIndustryComplianceRequirements returns healthcare compliance requirements
func (hm *HealthcareModel) GetIndustryComplianceRequirements() []ComplianceRequirement {
	return []ComplianceRequirement{
		{
			RequirementID:   "healthcare_hipaa",
			RequirementName: "HIPAA Compliance",
			RegulatoryBody:  "Department of Health and Human Services (HHS)",
			Jurisdiction:    "US",
			Description:     "Health Insurance Portability and Accountability Act compliance",
			Required:        true,
			PenaltyAmount:   "Up to $1.5M per violation",
			ComplianceSteps: []string{
				"Implement HIPAA policies and procedures",
				"Conduct risk assessments",
				"Train workforce on HIPAA requirements",
				"Implement technical safeguards",
			},
			Documentation: []string{
				"HIPAA policies and procedures",
				"Risk assessment reports",
				"Training records",
				"Business associate agreements",
			},
		},
		{
			RequirementID:   "healthcare_medical_licensing",
			RequirementName: "Medical Licensing",
			RegulatoryBody:  "State Medical Boards",
			Jurisdiction:    "US States",
			Description:     "Medical practice licensing and credentialing",
			Required:        true,
			PenaltyAmount:   "License suspension or revocation",
			ComplianceSteps: []string{
				"Obtain appropriate medical licenses",
				"Maintain continuing education requirements",
				"Submit regular license renewals",
				"Report any disciplinary actions",
			},
			Documentation: []string{
				"Medical license certificates",
				"Continuing education records",
				"License renewal applications",
			},
		},
		{
			RequirementID:   "healthcare_jcaho",
			RequirementName: "Joint Commission Accreditation",
			RegulatoryBody:  "Joint Commission",
			Jurisdiction:    "US",
			Description:     "Healthcare facility accreditation",
			Required:        false,
			PenaltyAmount:   "Loss of accreditation",
			ComplianceSteps: []string{
				"Prepare for accreditation survey",
				"Implement quality improvement programs",
				"Maintain patient safety standards",
				"Conduct regular self-assessments",
			},
			Documentation: []string{
				"Accreditation survey reports",
				"Quality improvement plans",
				"Patient safety reports",
			},
		},
		{
			RequirementID:   "healthcare_fda",
			RequirementName: "FDA Compliance",
			RegulatoryBody:  "Food and Drug Administration",
			Jurisdiction:    "US",
			Description:     "FDA regulations for medical devices and drugs",
			Required:        false,
			PenaltyAmount:   "Product recalls and fines",
			ComplianceSteps: []string{
				"Register with FDA if applicable",
				"Comply with quality system regulations",
				"Report adverse events",
				"Maintain device/drug records",
			},
			Documentation: []string{
				"FDA registration certificates",
				"Quality system documentation",
				"Adverse event reports",
			},
		},
	}
}

// calculateBaseHealthcareRisk calculates the base risk score for healthcare companies
func (hm *HealthcareModel) calculateBaseHealthcareRisk(business *models.RiskAssessmentRequest) float64 {
	baseScore := 0.35 // Base healthcare risk is moderate due to regulatory complexity

	// Adjust based on business characteristics
	if business.Metadata != nil {
		// Check for accreditation status
		if accreditation, exists := business.Metadata["accreditation"]; exists {
			if acc, ok := accreditation.(string); ok {
				switch acc {
				case "joint_commission", "cms", "aaaasf":
					baseScore -= 0.1 // Accredited facilities have lower risk
				case "pending":
					baseScore += 0.05
				case "none":
					baseScore += 0.1
				}
			}
		}

		// Check for medical license status
		if licenseStatus, exists := business.Metadata["license_status"]; exists {
			if status, ok := licenseStatus.(string); ok {
				switch status {
				case "active", "valid":
					baseScore -= 0.05
				case "pending", "under_review":
					baseScore += 0.05
				case "suspended", "revoked":
					baseScore += 0.2
				}
			}
		}

		// Check for quality certifications
		if certifications, exists := business.Metadata["quality_certifications"]; exists {
			if certs, ok := certifications.([]string); ok {
				baseScore -= float64(len(certs)) * 0.02 // Each certification reduces risk
			}
		}
	}

	// Ensure score is within bounds
	return math.Max(0.0, math.Min(1.0, baseScore))
}

// generateHealthcareRiskFactors generates healthcare-specific risk factors
func (hm *HealthcareModel) generateHealthcareRiskFactors(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRiskFactor {
	factors := []IndustryRiskFactor{
		{
			FactorID:            "healthcare_regulatory_compliance",
			FactorName:          "Regulatory Compliance",
			FactorCategory:      "regulatory",
			RiskScore:           baseScore + 0.08,
			RiskLevel:           hm.calculateRiskLevel(baseScore + 0.08),
			Description:         "Compliance with healthcare regulations and licensing",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Maintain active licenses, implement compliance monitoring, conduct regular audits",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "healthcare_patient_safety",
			FactorName:          "Patient Safety",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.12,
			RiskLevel:           hm.calculateRiskLevel(baseScore + 0.12),
			Description:         "Risk to patient safety and care quality",
			Impact:              "very_high",
			Likelihood:          "low",
			MitigationAdvice:    "Implement patient safety protocols, conduct safety training, monitor outcomes",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "healthcare_data_privacy",
			FactorName:          "Data Privacy (HIPAA)",
			FactorCategory:      "compliance",
			RiskScore:           baseScore + 0.10,
			RiskLevel:           hm.calculateRiskLevel(baseScore + 0.10),
			Description:         "Compliance with healthcare data privacy regulations",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Implement HIPAA safeguards, conduct privacy training, monitor data access",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "healthcare_medical_liability",
			FactorName:          "Medical Liability",
			FactorCategory:      "financial",
			RiskScore:           baseScore + 0.15,
			RiskLevel:           hm.calculateRiskLevel(baseScore + 0.15),
			Description:         "Risk of medical malpractice and liability claims",
			Impact:              "high",
			Likelihood:          "medium",
			MitigationAdvice:    "Maintain malpractice insurance, implement risk management, document care",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "healthcare_quality_assurance",
			FactorName:          "Quality Assurance",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.07,
			RiskLevel:           hm.calculateRiskLevel(baseScore + 0.07),
			Description:         "Quality of care and service delivery",
			Impact:              "high",
			Likelihood:          "low",
			MitigationAdvice:    "Implement quality improvement programs, monitor outcomes, conduct peer review",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
		{
			FactorID:            "healthcare_workforce_shortage",
			FactorName:          "Workforce Shortage",
			FactorCategory:      "operational",
			RiskScore:           baseScore + 0.13,
			RiskLevel:           hm.calculateRiskLevel(baseScore + 0.13),
			Description:         "Shortage of qualified healthcare professionals",
			Impact:              "medium",
			Likelihood:          "high",
			MitigationAdvice:    "Develop recruitment strategies, offer competitive benefits, invest in training",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "healthcare_technology_adoption",
			FactorName:          "Technology Adoption",
			FactorCategory:      "technology",
			RiskScore:           baseScore + 0.05,
			RiskLevel:           hm.calculateRiskLevel(baseScore + 0.05),
			Description:         "Adoption and integration of healthcare technology",
			Impact:              "medium",
			Likelihood:          "medium",
			MitigationAdvice:    "Invest in healthcare IT, train staff on new systems, ensure interoperability",
			IndustrySpecific:    true,
			RegulatoryRelevance: false,
		},
		{
			FactorID:            "healthcare_insurance_reimbursement",
			FactorName:          "Insurance Reimbursement",
			FactorCategory:      "financial",
			RiskScore:           baseScore + 0.09,
			RiskLevel:           hm.calculateRiskLevel(baseScore + 0.09),
			Description:         "Risk from insurance reimbursement changes",
			Impact:              "medium",
			Likelihood:          "medium",
			MitigationAdvice:    "Diversify payer mix, negotiate contracts, monitor reimbursement trends",
			IndustrySpecific:    true,
			RegulatoryRelevance: true,
		},
	}

	return factors
}

// assessHealthcareCompliance assesses compliance status for healthcare requirements
func (hm *HealthcareModel) assessHealthcareCompliance(business *models.RiskAssessmentRequest) []ComplianceStatus {
	statuses := []ComplianceStatus{
		{
			RequirementID:   "healthcare_hipaa",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"HIPAA compliance not verified"},
			Recommendations: []string{"Implement HIPAA compliance program"},
		},
		{
			RequirementID:   "healthcare_medical_licensing",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"Medical licensing not verified"},
			Recommendations: []string{"Verify medical license status"},
		},
		{
			RequirementID:   "healthcare_jcaho",
			Status:          "unknown",
			LastChecked:     time.Now(),
			ComplianceScore: 0.5,
			Issues:          []string{"Accreditation status not verified"},
			Recommendations: []string{"Pursue healthcare accreditation"},
		},
		{
			RequirementID:   "healthcare_fda",
			Status:          "not_applicable",
			LastChecked:     time.Now(),
			ComplianceScore: 0.8,
			Issues:          []string{},
			Recommendations: []string{"Monitor FDA requirements if applicable"},
		},
	}

	// Adjust based on business metadata
	if business.Metadata != nil {
		if accreditation, exists := business.Metadata["accreditation"]; exists {
			if acc, ok := accreditation.(string); ok {
				for i := range statuses {
					if statuses[i].RequirementID == "healthcare_jcaho" {
						switch acc {
						case "joint_commission", "cms", "aaaasf":
							statuses[i].Status = "compliant"
							statuses[i].ComplianceScore = 0.9
							statuses[i].Issues = []string{}
							statuses[i].Recommendations = []string{"Maintain accreditation compliance"}
						case "pending":
							statuses[i].Status = "non_compliant"
							statuses[i].ComplianceScore = 0.3
							statuses[i].Issues = []string{"Accreditation pending"}
							statuses[i].Recommendations = []string{"Complete accreditation process"}
						case "none":
							statuses[i].Status = "non_compliant"
							statuses[i].ComplianceScore = 0.2
							statuses[i].Issues = []string{"No accreditation"}
							statuses[i].Recommendations = []string{"Pursue healthcare accreditation"}
						}
					}
				}
			}
		}
	}

	return statuses
}

// generateHealthcareRecommendations generates healthcare-specific recommendations
func (hm *HealthcareModel) generateHealthcareRecommendations(business *models.RiskAssessmentRequest, baseScore float64) []IndustryRecommendation {
	recommendations := []IndustryRecommendation{
		{
			RecommendationID:   "healthcare_patient_safety",
			Category:           "operational",
			Priority:           "high",
			Title:              "Enhance Patient Safety",
			Description:        "Implement comprehensive patient safety program",
			ActionItems:        []string{"Develop safety protocols", "Implement incident reporting", "Conduct safety training"},
			ExpectedBenefit:    "Reduced patient safety incidents",
			ImplementationCost: "Medium",
			Timeline:           "3-6 months",
		},
		{
			RecommendationID:   "healthcare_hipaa_compliance",
			Category:           "compliance",
			Priority:           "high",
			Title:              "Strengthen HIPAA Compliance",
			Description:        "Implement robust HIPAA compliance program",
			ActionItems:        []string{"Conduct HIPAA risk assessment", "Implement safeguards", "Train workforce"},
			ExpectedBenefit:    "HIPAA compliance and reduced penalties",
			ImplementationCost: "Medium",
			Timeline:           "3-6 months",
		},
		{
			RecommendationID:   "healthcare_quality_improvement",
			Category:           "operational",
			Priority:           "medium",
			Title:              "Implement Quality Improvement",
			Description:        "Develop quality improvement program",
			ActionItems:        []string{"Establish quality metrics", "Implement monitoring", "Conduct peer review"},
			ExpectedBenefit:    "Improved care quality and outcomes",
			ImplementationCost: "Medium",
			Timeline:           "6-12 months",
		},
		{
			RecommendationID:   "healthcare_workforce_development",
			Category:           "operational",
			Priority:           "medium",
			Title:              "Develop Workforce",
			Description:        "Address workforce shortage and development",
			ActionItems:        []string{"Recruitment strategies", "Training programs", "Retention initiatives"},
			ExpectedBenefit:    "Improved workforce stability",
			ImplementationCost: "High",
			Timeline:           "12-18 months",
		},
	}

	return recommendations
}

// generateHealthcareRegulatoryFactors generates healthcare regulatory factors
func (hm *HealthcareModel) generateHealthcareRegulatoryFactors() []RegulatoryFactor {
	return []RegulatoryFactor{
		{
			FactorID:       "healthcare_hipaa",
			RegulationName: "Health Insurance Portability and Accountability Act (HIPAA)",
			RegulatoryBody: "Department of Health and Human Services",
			Jurisdiction:   "US",
			RiskImpact:     0.4,
			ComplianceCost: "High",
			PenaltyRisk:    "Very High",
			Description:    "Healthcare data privacy and security regulation",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "healthcare_medicare",
			RegulationName: "Medicare Conditions of Participation",
			RegulatoryBody: "Centers for Medicare & Medicaid Services",
			Jurisdiction:   "US",
			RiskImpact:     0.3,
			ComplianceCost: "Medium",
			PenaltyRisk:    "High",
			Description:    "Medicare participation requirements",
			LastUpdated:    time.Now(),
		},
		{
			FactorID:       "healthcare_stark_law",
			RegulationName: "Stark Law (Physician Self-Referral)",
			RegulatoryBody: "Centers for Medicare & Medicaid Services",
			Jurisdiction:   "US",
			RiskImpact:     0.25,
			ComplianceCost: "Medium",
			PenaltyRisk:    "High",
			Description:    "Physician self-referral prohibition",
			LastUpdated:    time.Now(),
		},
	}
}

// generateHealthcareMarketFactors generates healthcare market factors
func (hm *HealthcareModel) generateHealthcareMarketFactors() []MarketFactor {
	return []MarketFactor{
		{
			FactorID:       "healthcare_aging_population",
			FactorName:     "Aging Population",
			MarketTrend:    "growing",
			ImpactScore:    0.8,
			TimeHorizon:    "long-term",
			Description:    "Increasing demand for healthcare services",
			KeyDrivers:     []string{"Demographic trends", "Chronic disease prevalence", "Life expectancy"},
			RiskMitigation: []string{"Expand geriatric services", "Focus on chronic care management"},
		},
		{
			FactorID:       "healthcare_technology_disruption",
			FactorName:     "Technology Disruption",
			MarketTrend:    "volatile",
			ImpactScore:    0.6,
			TimeHorizon:    "medium-term",
			Description:    "Technology disruption in healthcare delivery",
			KeyDrivers:     []string{"Telemedicine", "AI/ML", "Digital health", "Wearables"},
			RiskMitigation: []string{"Invest in technology", "Partner with tech companies"},
		},
		{
			FactorID:       "healthcare_cost_pressure",
			FactorName:     "Cost Pressure",
			MarketTrend:    "stable",
			ImpactScore:    0.7,
			TimeHorizon:    "medium-term",
			Description:    "Pressure to reduce healthcare costs",
			KeyDrivers:     []string{"Insurance pressure", "Government regulation", "Consumer demand"},
			RiskMitigation: []string{"Improve efficiency", "Focus on value-based care"},
		},
	}
}

// generateHealthcareOperationalFactors generates healthcare operational factors
func (hm *HealthcareModel) generateHealthcareOperationalFactors(business *models.RiskAssessmentRequest) []OperationalFactor {
	return []OperationalFactor{
		{
			FactorID:            "healthcare_staffing",
			FactorName:          "Staffing Levels",
			OperationalArea:     "human_resources",
			RiskScore:           0.3,
			Criticality:         "critical",
			Description:         "Adequacy of healthcare staff",
			ControlMeasures:     []string{"Recruitment", "Retention", "Training"},
			MonitoringFrequency: "monthly",
		},
		{
			FactorID:            "healthcare_equipment",
			FactorName:          "Medical Equipment",
			OperationalArea:     "facilities",
			RiskScore:           0.25,
			Criticality:         "high",
			Description:         "Availability and maintenance of medical equipment",
			ControlMeasures:     []string{"Preventive maintenance", "Backup equipment", "Training"},
			MonitoringFrequency: "weekly",
		},
		{
			FactorID:            "healthcare_supply_chain",
			FactorName:          "Supply Chain",
			OperationalArea:     "procurement",
			RiskScore:           0.2,
			Criticality:         "medium",
			Description:         "Reliability of medical supply chain",
			ControlMeasures:     []string{"Multiple suppliers", "Inventory management", "Quality control"},
			MonitoringFrequency: "weekly",
		},
	}
}

// calculateOverallIndustryRisk calculates the overall industry risk score
func (hm *HealthcareModel) calculateOverallIndustryRisk(baseScore float64, factors []IndustryRiskFactor, compliance []ComplianceStatus) float64 {
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
func (hm *HealthcareModel) determineRiskLevel(score float64) models.RiskLevel {
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
func (hm *HealthcareModel) calculateRiskLevel(score float64) string {
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
func (hm *HealthcareModel) calculateConfidenceScore(business *models.RiskAssessmentRequest, factors []IndustryRiskFactor) float64 {
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
