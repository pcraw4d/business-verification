package models

import (
	"fmt"
	"time"
)

// RiskAssessment represents a risk assessment request and response
type RiskAssessment struct {
	ID                string                 `json:"id" db:"id"`
	BusinessID        string                 `json:"business_id" db:"business_id"`
	BusinessName      string                 `json:"business_name" db:"business_name"`
	BusinessAddress   string                 `json:"business_address" db:"business_address"`
	Industry          string                 `json:"industry" db:"industry"`
	Country           string                 `json:"country" db:"country"`
	RiskScore         float64                `json:"risk_score" db:"risk_score"`
	RiskLevel         RiskLevel              `json:"risk_level" db:"risk_level"`
	RiskFactors       []RiskFactor           `json:"risk_factors" db:"risk_factors"`
	PredictionHorizon int                    `json:"prediction_horizon" db:"prediction_horizon"` // months
	ConfidenceScore   float64                `json:"confidence_score" db:"confidence_score"`
	Status            AssessmentStatus       `json:"status" db:"status"`
	ModelType         string                 `json:"model_type" db:"model_type"`           // "industry", "custom", "ensemble"
	CustomModelID     string                 `json:"custom_model_id" db:"custom_model_id"` // ID of custom model if used
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
}

// RiskLevel represents the risk level classification
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// ConvertScoreToRiskLevel converts a risk score to a risk level
func ConvertScoreToRiskLevel(score float64) RiskLevel {
	switch {
	case score < 0.3:
		return RiskLevelLow
	case score < 0.6:
		return RiskLevelMedium
	case score < 0.8:
		return RiskLevelHigh
	default:
		return RiskLevelCritical
	}
}

// AssessmentStatus represents the status of a risk assessment
type AssessmentStatus string

const (
	StatusPending   AssessmentStatus = "pending"
	StatusCompleted AssessmentStatus = "completed"
	StatusFailed    AssessmentStatus = "failed"
	StatusError     AssessmentStatus = "error"
)

// RiskFactor represents an individual risk factor
type RiskFactor struct {
	Category    RiskCategory `json:"category"`
	Subcategory string       `json:"subcategory"`
	Name        string       `json:"name"`
	Score       float64      `json:"score"`
	Weight      float64      `json:"weight"`
	Description string       `json:"description"`
	Source      string       `json:"source"`
	Confidence  float64      `json:"confidence"`
	Impact      string       `json:"impact,omitempty"`
	Mitigation  string       `json:"mitigation,omitempty"`
	LastUpdated *time.Time   `json:"last_updated,omitempty"`
}

// RiskCategory represents the category of risk
type RiskCategory string

const (
	RiskCategoryFinancial     RiskCategory = "financial"
	RiskCategoryOperational   RiskCategory = "operational"
	RiskCategoryCompliance    RiskCategory = "compliance"
	RiskCategoryReputational  RiskCategory = "reputational"
	RiskCategoryRegulatory    RiskCategory = "regulatory"
	RiskCategoryGeopolitical  RiskCategory = "geopolitical"
	RiskCategoryTechnology    RiskCategory = "technology"
	RiskCategoryEnvironmental RiskCategory = "environmental"
)

// RiskSubcategory represents subcategories within each risk category
type RiskSubcategory struct {
	Category    RiskCategory `json:"category"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Weight      float64      `json:"weight"`
	Factors     []string     `json:"factors"`
}

// GetRiskSubcategories returns all available risk subcategories
func GetRiskSubcategories() map[RiskCategory][]RiskSubcategory {
	return map[RiskCategory][]RiskSubcategory{
		RiskCategoryFinancial: {
			{
				Category:    RiskCategoryFinancial,
				Name:        "liquidity_risk",
				Description: "Risk of insufficient cash flow or inability to meet short-term obligations",
				Weight:      0.3,
				Factors:     []string{"cash_flow", "working_capital", "debt_service", "credit_rating"},
			},
			{
				Category:    RiskCategoryFinancial,
				Name:        "credit_risk",
				Description: "Risk of loss due to counterparty default or credit deterioration",
				Weight:      0.25,
				Factors:     []string{"credit_score", "payment_history", "debt_levels", "collateral"},
			},
			{
				Category:    RiskCategoryFinancial,
				Name:        "market_risk",
				Description: "Risk of loss due to adverse market movements",
				Weight:      0.2,
				Factors:     []string{"interest_rate", "currency_exposure", "commodity_prices", "equity_volatility"},
			},
			{
				Category:    RiskCategoryFinancial,
				Name:        "operational_financial_risk",
				Description: "Financial risks arising from operational activities",
				Weight:      0.15,
				Factors:     []string{"revenue_volatility", "cost_structure", "profit_margins", "financial_controls"},
			},
			{
				Category:    RiskCategoryFinancial,
				Name:        "regulatory_financial_risk",
				Description: "Financial risks from regulatory changes or compliance failures",
				Weight:      0.1,
				Factors:     []string{"tax_compliance", "financial_reporting", "capital_requirements", "regulatory_fines"},
			},
		},
		RiskCategoryOperational: {
			{
				Category:    RiskCategoryOperational,
				Name:        "process_risk",
				Description: "Risk from inadequate or failed internal processes",
				Weight:      0.25,
				Factors:     []string{"process_automation", "quality_controls", "workflow_efficiency", "error_rates"},
			},
			{
				Category:    RiskCategoryOperational,
				Name:        "people_risk",
				Description: "Risk from human factors including skills, knowledge, and behavior",
				Weight:      0.2,
				Factors:     []string{"employee_turnover", "skill_gaps", "training_levels", "safety_records"},
			},
			{
				Category:    RiskCategoryOperational,
				Name:        "system_risk",
				Description: "Risk from IT systems, infrastructure, and technology failures",
				Weight:      0.2,
				Factors:     []string{"system_uptime", "data_backup", "security_measures", "disaster_recovery"},
			},
			{
				Category:    RiskCategoryOperational,
				Name:        "supply_chain_risk",
				Description: "Risk from supply chain disruptions or vendor failures",
				Weight:      0.15,
				Factors:     []string{"vendor_dependency", "supply_diversity", "logistics_reliability", "inventory_levels"},
			},
			{
				Category:    RiskCategoryOperational,
				Name:        "business_continuity_risk",
				Description: "Risk from inability to maintain business operations during disruptions",
				Weight:      0.1,
				Factors:     []string{"continuity_planning", "recovery_time", "backup_systems", "crisis_management"},
			},
			{
				Category:    RiskCategoryOperational,
				Name:        "capacity_risk",
				Description: "Risk from insufficient capacity to meet demand or growth",
				Weight:      0.1,
				Factors:     []string{"production_capacity", "scalability", "resource_utilization", "growth_readiness"},
			},
		},
		RiskCategoryCompliance: {
			{
				Category:    RiskCategoryCompliance,
				Name:        "regulatory_compliance",
				Description: "Risk from failure to comply with applicable regulations",
				Weight:      0.3,
				Factors:     []string{"regulatory_licenses", "compliance_monitoring", "regulatory_changes", "audit_results"},
			},
			{
				Category:    RiskCategoryCompliance,
				Name:        "data_protection_compliance",
				Description: "Risk from data protection and privacy regulation violations",
				Weight:      0.25,
				Factors:     []string{"gdpr_compliance", "data_encryption", "privacy_policies", "data_breaches"},
			},
			{
				Category:    RiskCategoryCompliance,
				Name:        "industry_standards",
				Description: "Risk from failure to meet industry-specific standards",
				Weight:      0.2,
				Factors:     []string{"iso_certifications", "industry_certifications", "quality_standards", "best_practices"},
			},
			{
				Category:    RiskCategoryCompliance,
				Name:        "internal_policies",
				Description: "Risk from violation of internal policies and procedures",
				Weight:      0.15,
				Factors:     []string{"policy_adherence", "code_of_conduct", "internal_audits", "policy_updates"},
			},
			{
				Category:    RiskCategoryCompliance,
				Name:        "contractual_compliance",
				Description: "Risk from failure to meet contractual obligations",
				Weight:      0.1,
				Factors:     []string{"contract_performance", "sla_compliance", "delivery_metrics", "penalty_risk"},
			},
		},
		RiskCategoryReputational: {
			{
				Category:    RiskCategoryReputational,
				Name:        "brand_reputation",
				Description: "Risk to brand value and customer perception",
				Weight:      0.3,
				Factors:     []string{"brand_sentiment", "customer_satisfaction", "brand_awareness", "market_position"},
			},
			{
				Category:    RiskCategoryReputational,
				Name:        "media_risk",
				Description: "Risk from negative media coverage or public relations issues",
				Weight:      0.25,
				Factors:     []string{"media_sentiment", "public_relations", "crisis_management", "social_media"},
			},
			{
				Category:    RiskCategoryReputational,
				Name:        "stakeholder_relations",
				Description: "Risk from poor relationships with key stakeholders",
				Weight:      0.2,
				Factors:     []string{"investor_relations", "customer_relations", "supplier_relations", "community_relations"},
			},
			{
				Category:    RiskCategoryReputational,
				Name:        "ethical_conduct",
				Description: "Risk from unethical behavior or conduct violations",
				Weight:      0.15,
				Factors:     []string{"ethical_policies", "conduct_violations", "whistleblower_reports", "ethics_training"},
			},
			{
				Category:    RiskCategoryReputational,
				Name:        "social_responsibility",
				Description: "Risk from failure to meet social responsibility expectations",
				Weight:      0.1,
				Factors:     []string{"sustainability_practices", "community_investment", "environmental_impact", "social_impact"},
			},
		},
		RiskCategoryRegulatory: {
			{
				Category:    RiskCategoryRegulatory,
				Name:        "licensing_risk",
				Description: "Risk from loss or suspension of required licenses",
				Weight:      0.3,
				Factors:     []string{"license_status", "renewal_requirements", "regulatory_approvals", "permit_compliance"},
			},
			{
				Category:    RiskCategoryRegulatory,
				Name:        "regulatory_changes",
				Description: "Risk from changes in regulatory environment",
				Weight:      0.25,
				Factors:     []string{"regulatory_monitoring", "policy_changes", "compliance_costs", "adaptation_time"},
			},
			{
				Category:    RiskCategoryRegulatory,
				Name:        "enforcement_risk",
				Description: "Risk from regulatory enforcement actions",
				Weight:      0.2,
				Factors:     []string{"enforcement_history", "regulatory_scrutiny", "penalty_risk", "remedial_actions"},
			},
			{
				Category:    RiskCategoryRegulatory,
				Name:        "jurisdictional_risk",
				Description: "Risk from operating in multiple regulatory jurisdictions",
				Weight:      0.15,
				Factors:     []string{"jurisdiction_count", "regulatory_complexity", "cross_border_compliance", "local_requirements"},
			},
			{
				Category:    RiskCategoryRegulatory,
				Name:        "regulatory_relationships",
				Description: "Risk from poor relationships with regulatory authorities",
				Weight:      0.1,
				Factors:     []string{"regulator_communication", "inspection_results", "regulatory_guidance", "authority_relations"},
			},
		},
		RiskCategoryGeopolitical: {
			{
				Category:    RiskCategoryGeopolitical,
				Name:        "political_stability",
				Description: "Risk from political instability in operating countries",
				Weight:      0.3,
				Factors:     []string{"political_risk_index", "government_stability", "election_cycles", "policy_continuity"},
			},
			{
				Category:    RiskCategoryGeopolitical,
				Name:        "trade_risk",
				Description: "Risk from trade restrictions, tariffs, and trade wars",
				Weight:      0.25,
				Factors:     []string{"trade_agreements", "tariff_exposure", "trade_restrictions", "supply_chain_impact"},
			},
			{
				Category:    RiskCategoryGeopolitical,
				Name:        "currency_risk",
				Description: "Risk from currency fluctuations and exchange rate volatility",
				Weight:      0.2,
				Factors:     []string{"currency_exposure", "exchange_rate_volatility", "hedging_strategies", "currency_controls"},
			},
			{
				Category:    RiskCategoryGeopolitical,
				Name:        "sanctions_risk",
				Description: "Risk from economic sanctions and embargoes",
				Weight:      0.15,
				Factors:     []string{"sanctions_exposure", "embargo_risk", "compliance_monitoring", "sanctions_screening"},
			},
			{
				Category:    RiskCategoryGeopolitical,
				Name:        "sovereign_risk",
				Description: "Risk from sovereign default or government actions",
				Weight:      0.1,
				Factors:     []string{"sovereign_rating", "government_debt", "default_history", "political_interference"},
			},
		},
		RiskCategoryTechnology: {
			{
				Category:    RiskCategoryTechnology,
				Name:        "cybersecurity_risk",
				Description: "Risk from cyber attacks and security breaches",
				Weight:      0.3,
				Factors:     []string{"security_incidents", "vulnerability_management", "security_controls", "incident_response"},
			},
			{
				Category:    RiskCategoryTechnology,
				Name:        "technology_obsolescence",
				Description: "Risk from outdated or obsolete technology",
				Weight:      0.2,
				Factors:     []string{"technology_age", "upgrade_cycles", "vendor_support", "innovation_adoption"},
			},
			{
				Category:    RiskCategoryTechnology,
				Name:        "data_risk",
				Description: "Risk from data loss, corruption, or unauthorized access",
				Weight:      0.2,
				Factors:     []string{"data_backup", "data_integrity", "access_controls", "data_governance"},
			},
			{
				Category:    RiskCategoryTechnology,
				Name:        "system_reliability",
				Description: "Risk from system failures and downtime",
				Weight:      0.15,
				Factors:     []string{"system_uptime", "redundancy", "disaster_recovery", "maintenance_schedules"},
			},
			{
				Category:    RiskCategoryTechnology,
				Name:        "digital_transformation",
				Description: "Risk from digital transformation initiatives",
				Weight:      0.15,
				Factors:     []string{"transformation_progress", "change_management", "technology_adoption", "digital_readiness"},
			},
		},
		RiskCategoryEnvironmental: {
			{
				Category:    RiskCategoryEnvironmental,
				Name:        "climate_risk",
				Description: "Risk from climate change and extreme weather events",
				Weight:      0.3,
				Factors:     []string{"climate_exposure", "weather_events", "climate_adaptation", "carbon_footprint"},
			},
			{
				Category:    RiskCategoryEnvironmental,
				Name:        "environmental_compliance",
				Description: "Risk from environmental regulation violations",
				Weight:      0.25,
				Factors:     []string{"environmental_permits", "emissions_compliance", "waste_management", "environmental_audits"},
			},
			{
				Category:    RiskCategoryEnvironmental,
				Name:        "natural_disasters",
				Description: "Risk from natural disasters and environmental hazards",
				Weight:      0.2,
				Factors:     []string{"disaster_exposure", "geographic_risk", "disaster_preparedness", "recovery_planning"},
			},
			{
				Category:    RiskCategoryEnvironmental,
				Name:        "resource_scarcity",
				Description: "Risk from scarcity of natural resources",
				Weight:      0.15,
				Factors:     []string{"resource_dependency", "supply_availability", "resource_efficiency", "alternative_sources"},
			},
			{
				Category:    RiskCategoryEnvironmental,
				Name:        "sustainability_risk",
				Description: "Risk from failure to meet sustainability expectations",
				Weight:      0.1,
				Factors:     []string{"sustainability_goals", "esg_performance", "stakeholder_expectations", "sustainability_reporting"},
			},
		},
	}
}

// GenerateDetailedRiskFactors generates detailed risk factors with subcategories
func GenerateDetailedRiskFactors(business *RiskAssessmentRequest, baseScore float64) []RiskFactor {
	riskFactors := make([]RiskFactor, 0)
	subcategories := GetRiskSubcategories()
	now := time.Now()

	// Generate risk factors for each category and subcategory
	for category, subcats := range subcategories {
		for _, subcat := range subcats {
			// Calculate subcategory-specific risk score
			subcatScore := calculateSubcategoryRiskScore(business, category, subcat, baseScore)

			// Generate individual factors for this subcategory
			for _, factorName := range subcat.Factors {
				factorScore := calculateFactorScore(business, category, subcat.Name, factorName, subcatScore)

				riskFactor := RiskFactor{
					Category:    category,
					Subcategory: subcat.Name,
					Name:        factorName,
					Score:       factorScore,
					Weight:      subcat.Weight / float64(len(subcat.Factors)), // Distribute weight across factors
					Description: generateFactorDescription(category, subcat.Name, factorName, factorScore),
					Source:      "enhanced_risk_model",
					Confidence:  calculateFactorConfidence(business, factorName),
					Impact:      generateFactorImpact(factorScore),
					Mitigation:  generateFactorMitigation(category, subcat.Name, factorName, factorScore),
					LastUpdated: &now,
				}

				riskFactors = append(riskFactors, riskFactor)
			}
		}
	}

	return riskFactors
}

// calculateSubcategoryRiskScore calculates risk score for a specific subcategory
func calculateSubcategoryRiskScore(business *RiskAssessmentRequest, category RiskCategory, subcat RiskSubcategory, baseScore float64) float64 {
	// Base score from overall risk assessment
	subcatScore := baseScore

	// Adjust based on business characteristics
	switch category {
	case RiskCategoryFinancial:
		subcatScore = adjustFinancialRiskScore(business, subcat.Name, subcatScore)
	case RiskCategoryOperational:
		subcatScore = adjustOperationalRiskScore(business, subcat.Name, subcatScore)
	case RiskCategoryCompliance:
		subcatScore = adjustComplianceRiskScore(business, subcat.Name, subcatScore)
	case RiskCategoryReputational:
		subcatScore = adjustReputationalRiskScore(business, subcat.Name, subcatScore)
	case RiskCategoryRegulatory:
		subcatScore = adjustRegulatoryRiskScore(business, subcat.Name, subcatScore)
	case RiskCategoryGeopolitical:
		subcatScore = adjustGeopoliticalRiskScore(business, subcat.Name, subcatScore)
	case RiskCategoryTechnology:
		subcatScore = adjustTechnologyRiskScore(business, subcat.Name, subcatScore)
	case RiskCategoryEnvironmental:
		subcatScore = adjustEnvironmentalRiskScore(business, subcat.Name, subcatScore)
	}

	// Apply subcategory weight
	subcatScore *= subcat.Weight

	// Ensure score is between 0 and 1
	if subcatScore > 1.0 {
		subcatScore = 1.0
	} else if subcatScore < 0.0 {
		subcatScore = 0.0
	}

	return subcatScore
}

// calculateFactorScore calculates risk score for a specific factor
func calculateFactorScore(business *RiskAssessmentRequest, category RiskCategory, subcategory, factorName string, subcatScore float64) float64 {
	// Start with subcategory score
	factorScore := subcatScore

	// Add factor-specific adjustments
	switch factorName {
	// Financial factors
	case "cash_flow":
		factorScore += 0.1 // Assume positive cash flow reduces risk
	case "working_capital":
		factorScore += 0.05
	case "debt_service":
		factorScore += 0.15 // High debt service increases risk
	case "credit_rating":
		factorScore += 0.1

	// Operational factors
	case "process_automation":
		factorScore -= 0.1 // Automation reduces risk
	case "quality_controls":
		factorScore -= 0.05
	case "employee_turnover":
		factorScore += 0.1 // High turnover increases risk
	case "system_uptime":
		factorScore -= 0.1 // High uptime reduces risk

	// Compliance factors
	case "regulatory_licenses":
		factorScore -= 0.1 // Having licenses reduces risk
	case "gdpr_compliance":
		factorScore -= 0.05
	case "iso_certifications":
		factorScore -= 0.05

	// Technology factors
	case "security_incidents":
		factorScore += 0.2 // Security incidents increase risk
	case "vulnerability_management":
		factorScore -= 0.1
	case "data_backup":
		factorScore -= 0.05

	// Environmental factors
	case "climate_exposure":
		factorScore += 0.1
	case "environmental_permits":
		factorScore -= 0.05
	case "disaster_exposure":
		factorScore += 0.1
	}

	// Ensure score is between 0 and 1
	if factorScore > 1.0 {
		factorScore = 1.0
	} else if factorScore < 0.0 {
		factorScore = 0.0
	}

	return factorScore
}

// generateFactorDescription generates a description for a risk factor
func generateFactorDescription(category RiskCategory, subcategory, factorName string, score float64) string {
	riskLevel := "low"
	if score > 0.7 {
		riskLevel = "high"
	} else if score > 0.4 {
		riskLevel = "moderate"
	}

	return fmt.Sprintf("%s risk factor in %s subcategory shows %s risk level (score: %.2f)",
		factorName, subcategory, riskLevel, score)
}

// calculateFactorConfidence calculates confidence for a risk factor
func calculateFactorConfidence(business *RiskAssessmentRequest, factorName string) float64 {
	baseConfidence := 0.8

	// Adjust confidence based on data availability
	switch factorName {
	case "cash_flow", "working_capital", "debt_service":
		// Use business name length as proxy for business size
		if len(business.BusinessName) > 10 {
			baseConfidence += 0.1
		}
	case "employee_turnover", "training_levels":
		// Use industry as proxy for employee count
		if business.Industry == "technology" || business.Industry == "finance" {
			baseConfidence += 0.05
		}
	case "system_uptime", "security_incidents":
		if business.Website != "" {
			baseConfidence += 0.05
		}
	}

	// Adjust based on business completeness
	if business.BusinessAddress != "" {
		baseConfidence += 0.05
	}
	if business.Phone != "" {
		baseConfidence += 0.02
	}
	if business.Email != "" {
		baseConfidence += 0.02
	}

	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	return baseConfidence
}

// generateFactorImpact generates impact description for a risk factor
func generateFactorImpact(score float64) string {
	if score > 0.7 {
		return "High impact on overall risk assessment"
	} else if score > 0.4 {
		return "Moderate impact on overall risk assessment"
	} else {
		return "Low impact on overall risk assessment"
	}
}

// generateFactorMitigation generates mitigation recommendation for a risk factor
func generateFactorMitigation(category RiskCategory, subcategory, factorName string, score float64) string {
	if score < 0.3 {
		return "Continue current risk management practices"
	}

	switch category {
	case RiskCategoryFinancial:
		return "Implement additional financial controls and monitoring"
	case RiskCategoryOperational:
		return "Enhance operational processes and controls"
	case RiskCategoryCompliance:
		return "Strengthen compliance monitoring and training"
	case RiskCategoryReputational:
		return "Improve stakeholder communication and crisis management"
	case RiskCategoryRegulatory:
		return "Enhance regulatory monitoring and compliance procedures"
	case RiskCategoryGeopolitical:
		return "Develop geopolitical risk monitoring and mitigation strategies"
	case RiskCategoryTechnology:
		return "Strengthen cybersecurity and technology risk management"
	case RiskCategoryEnvironmental:
		return "Implement environmental risk management and sustainability practices"
	default:
		return "Review and enhance risk management practices"
	}
}

// Risk assessment adjustment functions for each category
func adjustFinancialRiskScore(business *RiskAssessmentRequest, subcategory string, baseScore float64) float64 {
	switch subcategory {
	case "liquidity_risk":
		// Use business name length as proxy for business size
		if len(business.BusinessName) > 15 {
			baseScore -= 0.1 // Larger businesses typically have better liquidity
		}
	case "credit_risk":
		// Use industry as proxy for business maturity
		if business.Industry == "finance" || business.Industry == "healthcare" {
			baseScore -= 0.05 // Established industries have lower credit risk
		}
	case "market_risk":
		if business.Industry == "technology" {
			baseScore += 0.1 // Technology companies have higher market risk
		}
	}
	return baseScore
}

func adjustOperationalRiskScore(business *RiskAssessmentRequest, subcategory string, baseScore float64) float64 {
	switch subcategory {
	case "process_risk":
		if business.Website != "" {
			baseScore -= 0.05 // Digital presence suggests better processes
		}
	case "people_risk":
		// Use industry as proxy for company size
		if business.Industry == "technology" || business.Industry == "finance" {
			baseScore += 0.05 // Larger industries typically have more people risk
		}
	case "system_risk":
		if business.Website != "" {
			baseScore += 0.1 // Digital presence increases system risk
		}
	}
	return baseScore
}

func adjustComplianceRiskScore(business *RiskAssessmentRequest, subcategory string, baseScore float64) float64 {
	switch subcategory {
	case "regulatory_compliance":
		if business.Industry == "finance" || business.Industry == "healthcare" {
			baseScore += 0.1 // Highly regulated industries
		}
	case "data_protection_compliance":
		if business.Website != "" || business.Email != "" {
			baseScore += 0.05 // Digital presence increases data protection risk
		}
	}
	return baseScore
}

func adjustReputationalRiskScore(business *RiskAssessmentRequest, subcategory string, baseScore float64) float64 {
	switch subcategory {
	case "brand_reputation":
		if business.Website != "" {
			baseScore -= 0.05 // Digital presence can improve reputation
		}
	case "media_risk":
		if business.Industry == "technology" {
			baseScore += 0.05 // Technology companies face more media scrutiny
		}
	}
	return baseScore
}

func adjustRegulatoryRiskScore(business *RiskAssessmentRequest, subcategory string, baseScore float64) float64 {
	switch subcategory {
	case "licensing_risk":
		if business.Industry == "finance" || business.Industry == "healthcare" {
			baseScore += 0.1 // Industries requiring licenses
		}
	case "regulatory_changes":
		if business.Country != "US" {
			baseScore += 0.05 // Non-US companies face more regulatory complexity
		}
	}
	return baseScore
}

func adjustGeopoliticalRiskScore(business *RiskAssessmentRequest, subcategory string, baseScore float64) float64 {
	switch subcategory {
	case "political_stability":
		if business.Country == "US" || business.Country == "CA" || business.Country == "GB" {
			baseScore -= 0.1 // Stable countries
		}
	case "trade_risk":
		if business.Country != "US" {
			baseScore += 0.05 // International companies face more trade risk
		}
	}
	return baseScore
}

func adjustTechnologyRiskScore(business *RiskAssessmentRequest, subcategory string, baseScore float64) float64 {
	switch subcategory {
	case "cybersecurity_risk":
		if business.Website != "" || business.Email != "" {
			baseScore += 0.1 // Digital presence increases cyber risk
		}
	case "technology_obsolescence":
		if business.Industry == "technology" {
			baseScore += 0.05 // Technology companies face obsolescence risk
		}
	}
	return baseScore
}

func adjustEnvironmentalRiskScore(business *RiskAssessmentRequest, subcategory string, baseScore float64) float64 {
	switch subcategory {
	case "climate_risk":
		// All businesses face climate risk
		baseScore += 0.05
	case "environmental_compliance":
		if business.Industry == "manufacturing" || business.Industry == "construction" {
			baseScore += 0.1 // Industries with environmental impact
		}
	}
	return baseScore
}

// RiskAssessmentRequest represents a request for risk assessment
type RiskAssessmentRequest struct {
	BusinessName            string                 `json:"business_name" validate:"required,min=1,max=255"`
	BusinessAddress         string                 `json:"business_address" validate:"required,min=10,max=500"`
	Industry                string                 `json:"industry" validate:"required,min=1,max=100"`
	Country                 string                 `json:"country" validate:"required,len=2"`
	Phone                   string                 `json:"phone,omitempty" validate:"omitempty,min=10,max=20"`
	Email                   string                 `json:"email,omitempty" validate:"omitempty,email"`
	Website                 string                 `json:"website,omitempty" validate:"omitempty,url"`
	PredictionHorizon       int                    `json:"prediction_horizon,omitempty" validate:"omitempty,min=1,max=24"`
	ModelType               string                 `json:"model_type,omitempty" validate:"omitempty,oneof=auto xgboost lstm ensemble custom"`
	CustomModelID           string                 `json:"custom_model_id,omitempty" validate:"omitempty,min=1,max=100"`
	IncludeTemporalAnalysis bool                   `json:"include_temporal_analysis,omitempty"`
	CustomBusinessData      map[string]interface{} `json:"custom_business_data,omitempty"` // Additional data for custom models
	Metadata                map[string]interface{} `json:"metadata,omitempty"`
}

// RiskAssessmentResponse represents the response from a risk assessment
type RiskAssessmentResponse struct {
	ID                string                 `json:"id"`
	BusinessID        string                 `json:"business_id"`
	RiskScore         float64                `json:"risk_score"`
	RiskLevel         RiskLevel              `json:"risk_level"`
	RiskFactors       []RiskFactor           `json:"risk_factors"`
	PredictionHorizon int                    `json:"prediction_horizon"`
	ConfidenceScore   float64                `json:"confidence_score"`
	Status            AssessmentStatus       `json:"status"`
	ModelType         string                 `json:"model_type"`                // "industry", "custom", "ensemble"
	CustomModelID     string                 `json:"custom_model_id,omitempty"` // ID of custom model if used
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// RiskPrediction represents a risk prediction for future time periods
type RiskPrediction struct {
	BusinessID       string             `json:"business_id"`
	PredictionDate   time.Time          `json:"prediction_date"`
	HorizonMonths    int                `json:"horizon_months"`
	PredictedScore   float64            `json:"predicted_score"`
	PredictedLevel   RiskLevel          `json:"predicted_level"`
	ConfidenceScore  float64            `json:"confidence_score"`
	RiskFactors      []RiskFactor       `json:"risk_factors"`
	ScenarioAnalysis []ScenarioAnalysis `json:"scenario_analysis,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
}

// ScenarioAnalysis represents different risk scenarios
type ScenarioAnalysis struct {
	ScenarioName string    `json:"scenario_name"`
	Description  string    `json:"description"`
	RiskScore    float64   `json:"risk_score"`
	RiskLevel    RiskLevel `json:"risk_level"`
	Probability  float64   `json:"probability"`
	Impact       string    `json:"impact"`
}

// RiskScenario represents a risk scenario analysis
type RiskScenario struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Probability float64 `json:"probability"`
	Impact      string  `json:"impact"`
	RiskScore   float64 `json:"risk_score"`
	TimeHorizon int     `json:"time_horizon"`
	Mitigation  string  `json:"mitigation"`
}

// ExternalData represents data from external sources
type ExternalData struct {
	Source      string                 `json:"source"`
	SourceType  string                 `json:"source_type"`
	Data        map[string]interface{} `json:"data"`
	Confidence  float64                `json:"confidence"`
	LastUpdated time.Time              `json:"last_updated"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// ComplianceCheck represents a compliance check result
type ComplianceCheck struct {
	CheckType   string                 `json:"check_type"`
	Status      string                 `json:"status"`
	Description string                 `json:"description"`
	RiskLevel   RiskLevel              `json:"risk_level"`
	Details     map[string]interface{} `json:"details"`
	CheckedAt   time.Time              `json:"checked_at"`
	NextCheckAt *time.Time             `json:"next_check_at,omitempty"`
}

// SanctionsCheck represents a sanctions screening result
type SanctionsCheck struct {
	EntityName    string                 `json:"entity_name"`
	EntityType    string                 `json:"entity_type"`
	MatchType     string                 `json:"match_type"`
	MatchScore    float64                `json:"match_score"`
	SanctionsList string                 `json:"sanctions_list"`
	Details       map[string]interface{} `json:"details"`
	CheckedAt     time.Time              `json:"checked_at"`
}

// AdverseMedia represents adverse media monitoring results
type AdverseMedia struct {
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
	Sentiment   string    `json:"sentiment"`
	RiskLevel   RiskLevel `json:"risk_level"`
	Summary     string    `json:"summary"`
	Keywords    []string  `json:"keywords"`
}
