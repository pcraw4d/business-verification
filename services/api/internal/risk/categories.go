package risk

import (
	"time"
)

// RiskCategoryDefinition provides detailed information about a risk category
type RiskCategoryDefinition struct {
	Category      RiskCategory           `json:"category"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Subcategories []RiskSubcategory      `json:"subcategories"`
	Factors       []RiskFactorDefinition `json:"factors"`
	Weight        float64                `json:"weight"` // Overall weight in risk assessment
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// RiskSubcategory represents a specific area within a risk category
type RiskSubcategory struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Weight      float64                `json:"weight"`  // Weight within the category
	Factors     []string               `json:"factors"` // Factor IDs in this subcategory
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RiskFactorDefinition provides detailed information about a specific risk factor
type RiskFactorDefinition struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Category        RiskCategory           `json:"category"`
	Subcategory     string                 `json:"subcategory"`
	Weight          float64                `json:"weight"`           // Weight within the category
	CalculationType string                 `json:"calculation_type"` // "direct", "derived", "composite"
	DataSources     []string               `json:"data_sources"`
	Thresholds      map[RiskLevel]float64  `json:"thresholds"`
	Formula         string                 `json:"formula,omitempty"` // Calculation formula if applicable
	Unit            string                 `json:"unit,omitempty"`    // Unit of measurement
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// RiskCategoryRegistry manages risk category definitions
type RiskCategoryRegistry struct {
	categories map[RiskCategory]*RiskCategoryDefinition
	factors    map[string]*RiskFactorDefinition
}

// NewRiskCategoryRegistry creates a new risk category registry
func NewRiskCategoryRegistry() *RiskCategoryRegistry {
	return &RiskCategoryRegistry{
		categories: make(map[RiskCategory]*RiskCategoryDefinition),
		factors:    make(map[string]*RiskFactorDefinition),
	}
}

// RegisterCategory registers a risk category definition
func (r *RiskCategoryRegistry) RegisterCategory(definition *RiskCategoryDefinition) {
	r.categories[definition.Category] = definition

	// Register all factors in this category
	for _, factor := range definition.Factors {
		r.factors[factor.ID] = &factor
	}
}

// RegisterFactor registers a single risk factor definition
func (r *RiskCategoryRegistry) RegisterFactor(factor *RiskFactorDefinition) {
	r.factors[factor.ID] = factor
}

// GetCategory retrieves a risk category definition
func (r *RiskCategoryRegistry) GetCategory(category RiskCategory) (*RiskCategoryDefinition, bool) {
	definition, exists := r.categories[category]
	return definition, exists
}

// GetFactor retrieves a risk factor definition
func (r *RiskCategoryRegistry) GetFactor(factorID string) (*RiskFactorDefinition, bool) {
	factor, exists := r.factors[factorID]
	return factor, exists
}

// ListCategories returns all registered risk categories
func (r *RiskCategoryRegistry) ListCategories() []RiskCategory {
	categories := make([]RiskCategory, 0, len(r.categories))
	for category := range r.categories {
		categories = append(categories, category)
	}
	return categories
}

// ListFactors returns all registered risk factors
func (r *RiskCategoryRegistry) ListFactors() []*RiskFactorDefinition {
	factors := make([]*RiskFactorDefinition, 0, len(r.factors))
	for _, factor := range r.factors {
		factors = append(factors, factor)
	}
	return factors
}

// GetFactorsByCategory returns all factors for a specific category
func (r *RiskCategoryRegistry) GetFactorsByCategory(category RiskCategory) []*RiskFactorDefinition {
	var factors []*RiskFactorDefinition
	for _, factor := range r.factors {
		if factor.Category == category {
			factors = append(factors, factor)
		}
	}
	return factors
}

// GetFactorsBySubcategory returns all factors for a specific subcategory
func (r *RiskCategoryRegistry) GetFactorsBySubcategory(category RiskCategory, subcategory string) []*RiskFactorDefinition {
	var factors []*RiskFactorDefinition
	for _, factor := range r.factors {
		if factor.Category == category && factor.Subcategory == subcategory {
			factors = append(factors, factor)
		}
	}
	return factors
}

// CreateDefaultRiskCategories creates the default risk category definitions
func CreateDefaultRiskCategories() *RiskCategoryRegistry {
	registry := NewRiskCategoryRegistry()

	// Financial Risk Category
	financialCategory := &RiskCategoryDefinition{
		Category:    RiskCategoryFinancial,
		Name:        "Financial Risk",
		Description: "Risks related to financial stability, liquidity, creditworthiness, and financial performance of the business",
		Weight:      0.25,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Subcategories: []RiskSubcategory{
			{
				ID:          "financial_liquidity",
				Name:        "Liquidity Risk",
				Description: "Risk of insufficient cash flow and liquid assets to meet obligations",
				Weight:      0.3,
			},
			{
				ID:          "financial_credit",
				Name:        "Credit Risk",
				Description: "Risk of default on financial obligations and creditworthiness",
				Weight:      0.25,
			},
			{
				ID:          "financial_performance",
				Name:        "Performance Risk",
				Description: "Risk of poor financial performance and profitability",
				Weight:      0.25,
			},
			{
				ID:          "financial_structure",
				Name:        "Capital Structure Risk",
				Description: "Risk related to debt levels, leverage, and capital adequacy",
				Weight:      0.2,
			},
		},
		Factors: []RiskFactorDefinition{
			{
				ID:              "cash_flow_coverage",
				Name:            "Cash Flow Coverage Ratio",
				Description:     "Ratio of operating cash flow to total debt obligations",
				Category:        RiskCategoryFinancial,
				Subcategory:     "financial_liquidity",
				Weight:          0.4,
				CalculationType: "direct",
				DataSources:     []string{"financial_statements", "cash_flow_reports"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      2.0,
					RiskLevelMedium:   1.5,
					RiskLevelHigh:     1.0,
					RiskLevelCritical: 0.5,
				},
				Formula:   "Operating Cash Flow / Total Debt Obligations",
				Unit:      "ratio",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:              "debt_to_equity",
				Name:            "Debt-to-Equity Ratio",
				Description:     "Ratio of total debt to shareholders' equity",
				Category:        RiskCategoryFinancial,
				Subcategory:     "financial_structure",
				Weight:          0.3,
				CalculationType: "direct",
				DataSources:     []string{"balance_sheet", "financial_statements"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      0.5,
					RiskLevelMedium:   1.0,
					RiskLevelHigh:     2.0,
					RiskLevelCritical: 3.0,
				},
				Formula:   "Total Debt / Shareholders' Equity",
				Unit:      "ratio",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:              "credit_score",
				Name:            "Business Credit Score",
				Description:     "Credit score from business credit bureaus",
				Category:        RiskCategoryFinancial,
				Subcategory:     "financial_credit",
				Weight:          0.35,
				CalculationType: "direct",
				DataSources:     []string{"credit_bureaus", "dun_bradstreet", "equifax_business"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      80,
					RiskLevelMedium:   60,
					RiskLevelHigh:     40,
					RiskLevelCritical: 20,
				},
				Unit:      "score",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:              "profit_margin",
				Name:            "Profit Margin Trend",
				Description:     "Trend in profit margins over the last 12 months",
				Category:        RiskCategoryFinancial,
				Subcategory:     "financial_performance",
				Weight:          0.3,
				CalculationType: "derived",
				DataSources:     []string{"income_statements", "financial_reports"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      0.05,  // 5% improvement
					RiskLevelMedium:   0.0,   // stable
					RiskLevelHigh:     -0.05, // 5% decline
					RiskLevelCritical: -0.15, // 15% decline
				},
				Formula:   "(Current Margin - Previous Margin) / Previous Margin",
				Unit:      "percentage_change",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Operational Risk Category
	operationalCategory := &RiskCategoryDefinition{
		Category:    RiskCategoryOperational,
		Name:        "Operational Risk",
		Description: "Risks related to internal processes, systems, people, and operational efficiency",
		Weight:      0.2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Subcategories: []RiskSubcategory{
			{
				ID:          "operational_processes",
				Name:        "Process Risk",
				Description: "Risk of process failures and operational inefficiencies",
				Weight:      0.3,
			},
			{
				ID:          "operational_people",
				Name:        "People Risk",
				Description: "Risk related to human resources, skills, and management",
				Weight:      0.25,
			},
			{
				ID:          "operational_systems",
				Name:        "Systems Risk",
				Description: "Risk of system failures and technology issues",
				Weight:      0.25,
			},
			{
				ID:          "operational_supply",
				Name:        "Supply Chain Risk",
				Description: "Risk related to suppliers and supply chain disruptions",
				Weight:      0.2,
			},
		},
		Factors: []RiskFactorDefinition{
			{
				ID:              "employee_turnover",
				Name:            "Employee Turnover Rate",
				Description:     "Annual employee turnover rate as a percentage",
				Category:        RiskCategoryOperational,
				Subcategory:     "operational_people",
				Weight:          0.4,
				CalculationType: "direct",
				DataSources:     []string{"hr_records", "employment_data"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      0.05, // 5%
					RiskLevelMedium:   0.15, // 15%
					RiskLevelHigh:     0.25, // 25%
					RiskLevelCritical: 0.4,  // 40%
				},
				Formula:   "(Employees Left / Average Employees) * 100",
				Unit:      "percentage",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:              "system_uptime",
				Name:            "System Uptime",
				Description:     "Percentage of time systems are operational",
				Category:        RiskCategoryOperational,
				Subcategory:     "operational_systems",
				Weight:          0.35,
				CalculationType: "direct",
				DataSources:     []string{"system_monitoring", "it_reports"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      0.99, // 99%
					RiskLevelMedium:   0.95, // 95%
					RiskLevelHigh:     0.90, // 90%
					RiskLevelCritical: 0.85, // 85%
				},
				Unit:      "percentage",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:              "supplier_concentration",
				Name:            "Supplier Concentration",
				Description:     "Percentage of spend with top 3 suppliers",
				Category:        RiskCategoryOperational,
				Subcategory:     "operational_supply",
				Weight:          0.3,
				CalculationType: "direct",
				DataSources:     []string{"procurement_data", "supplier_reports"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      0.3, // 30%
					RiskLevelMedium:   0.5, // 50%
					RiskLevelHigh:     0.7, // 70%
					RiskLevelCritical: 0.9, // 90%
				},
				Formula:   "(Top 3 Supplier Spend / Total Spend) * 100",
				Unit:      "percentage",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Regulatory Risk Category
	regulatoryCategory := &RiskCategoryDefinition{
		Category:    RiskCategoryRegulatory,
		Name:        "Regulatory Risk",
		Description: "Risks related to compliance with laws, regulations, and industry standards",
		Weight:      0.2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Subcategories: []RiskSubcategory{
			{
				ID:          "regulatory_compliance",
				Name:        "Compliance Risk",
				Description: "Risk of non-compliance with applicable regulations",
				Weight:      0.4,
			},
			{
				ID:          "regulatory_changes",
				Name:        "Regulatory Change Risk",
				Description: "Risk from changes in regulations and laws",
				Weight:      0.3,
			},
			{
				ID:          "regulatory_enforcement",
				Name:        "Enforcement Risk",
				Description: "Risk of regulatory enforcement actions",
				Weight:      0.3,
			},
		},
		Factors: []RiskFactorDefinition{
			{
				ID:              "compliance_score",
				Name:            "Compliance Score",
				Description:     "Overall compliance score based on regulatory requirements",
				Category:        RiskCategoryRegulatory,
				Subcategory:     "regulatory_compliance",
				Weight:          0.4,
				CalculationType: "composite",
				DataSources:     []string{"compliance_audits", "regulatory_reports"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      90,
					RiskLevelMedium:   75,
					RiskLevelHigh:     60,
					RiskLevelCritical: 40,
				},
				Unit:      "score",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:              "regulatory_violations",
				Name:            "Regulatory Violations",
				Description:     "Number of regulatory violations in the last 24 months",
				Category:        RiskCategoryRegulatory,
				Subcategory:     "regulatory_enforcement",
				Weight:          0.35,
				CalculationType: "direct",
				DataSources:     []string{"regulatory_reports", "enforcement_actions"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      0,
					RiskLevelMedium:   1,
					RiskLevelHigh:     3,
					RiskLevelCritical: 5,
				},
				Unit:      "count",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:              "license_status",
				Name:            "License Status",
				Description:     "Status of required business licenses and permits",
				Category:        RiskCategoryRegulatory,
				Subcategory:     "regulatory_compliance",
				Weight:          0.25,
				CalculationType: "direct",
				DataSources:     []string{"license_databases", "regulatory_agencies"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      1.0, // All valid
					RiskLevelMedium:   0.8, // 80% valid
					RiskLevelHigh:     0.6, // 60% valid
					RiskLevelCritical: 0.4, // 40% valid
				},
				Formula:   "Valid Licenses / Total Required Licenses",
				Unit:      "ratio",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Reputational Risk Category
	reputationalCategory := &RiskCategoryDefinition{
		Category:    RiskCategoryReputational,
		Name:        "Reputational Risk",
		Description: "Risks related to brand reputation, public perception, and stakeholder trust",
		Weight:      0.15,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Subcategories: []RiskSubcategory{
			{
				ID:          "reputational_media",
				Name:        "Media Risk",
				Description: "Risk from negative media coverage and public relations issues",
				Weight:      0.3,
			},
			{
				ID:          "reputational_social",
				Name:        "Social Media Risk",
				Description: "Risk from social media sentiment and online reputation",
				Weight:      0.3,
			},
			{
				ID:          "reputational_customer",
				Name:        "Customer Satisfaction Risk",
				Description: "Risk from customer complaints and satisfaction issues",
				Weight:      0.25,
			},
			{
				ID:          "reputational_stakeholder",
				Name:        "Stakeholder Risk",
				Description: "Risk from stakeholder relationships and trust issues",
				Weight:      0.15,
			},
		},
		Factors: []RiskFactorDefinition{
			{
				ID:              "sentiment_score",
				Name:            "Online Sentiment Score",
				Description:     "Overall sentiment score from online mentions and reviews",
				Category:        RiskCategoryReputational,
				Subcategory:     "reputational_social",
				Weight:          0.35,
				CalculationType: "derived",
				DataSources:     []string{"social_media", "review_platforms", "news_monitoring"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      0.7, // 70% positive
					RiskLevelMedium:   0.5, // 50% positive
					RiskLevelHigh:     0.3, // 30% positive
					RiskLevelCritical: 0.1, // 10% positive
				},
				Unit:      "sentiment_ratio",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:              "customer_satisfaction",
				Name:            "Customer Satisfaction Score",
				Description:     "Customer satisfaction score from surveys and feedback",
				Category:        RiskCategoryReputational,
				Subcategory:     "reputational_customer",
				Weight:          0.3,
				CalculationType: "direct",
				DataSources:     []string{"customer_surveys", "feedback_systems"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      4.0, // 4.0/5.0
					RiskLevelMedium:   3.5, // 3.5/5.0
					RiskLevelHigh:     3.0, // 3.0/5.0
					RiskLevelCritical: 2.5, // 2.5/5.0
				},
				Unit:      "score",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:              "negative_mentions",
				Name:            "Negative Media Mentions",
				Description:     "Number of negative media mentions in the last 30 days",
				Category:        RiskCategoryReputational,
				Subcategory:     "reputational_media",
				Weight:          0.25,
				CalculationType: "direct",
				DataSources:     []string{"news_monitoring", "media_tracking"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      0,
					RiskLevelMedium:   5,
					RiskLevelHigh:     15,
					RiskLevelCritical: 30,
				},
				Unit:      "count",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Cybersecurity Risk Category
	cybersecurityCategory := &RiskCategoryDefinition{
		Category:    RiskCategoryCybersecurity,
		Name:        "Cybersecurity Risk",
		Description: "Risks related to information security, data protection, and cyber threats",
		Weight:      0.2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Subcategories: []RiskSubcategory{
			{
				ID:          "cybersecurity_technical",
				Name:        "Technical Security Risk",
				Description: "Risk from technical vulnerabilities and security gaps",
				Weight:      0.35,
			},
			{
				ID:          "cybersecurity_data",
				Name:        "Data Protection Risk",
				Description: "Risk related to data privacy and protection",
				Weight:      0.3,
			},
			{
				ID:          "cybersecurity_incident",
				Name:        "Incident Response Risk",
				Description: "Risk from security incidents and response capabilities",
				Weight:      0.2,
			},
			{
				ID:          "cybersecurity_compliance",
				Name:        "Security Compliance Risk",
				Description: "Risk from security compliance and governance",
				Weight:      0.15,
			},
		},
		Factors: []RiskFactorDefinition{
			{
				ID:              "security_score",
				Name:            "Security Posture Score",
				Description:     "Overall cybersecurity posture score",
				Category:        RiskCategoryCybersecurity,
				Subcategory:     "cybersecurity_technical",
				Weight:          0.4,
				CalculationType: "composite",
				DataSources:     []string{"security_assessments", "vulnerability_scans"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      85,
					RiskLevelMedium:   70,
					RiskLevelHigh:     55,
					RiskLevelCritical: 40,
				},
				Unit:      "score",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:              "data_breaches",
				Name:            "Data Breach Incidents",
				Description:     "Number of data breach incidents in the last 12 months",
				Category:        RiskCategoryCybersecurity,
				Subcategory:     "cybersecurity_data",
				Weight:          0.35,
				CalculationType: "direct",
				DataSources:     []string{"incident_reports", "breach_notifications"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      0,
					RiskLevelMedium:   1,
					RiskLevelHigh:     2,
					RiskLevelCritical: 3,
				},
				Unit:      "count",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:              "patch_compliance",
				Name:            "Security Patch Compliance",
				Description:     "Percentage of critical security patches applied within SLA",
				Category:        RiskCategoryCybersecurity,
				Subcategory:     "cybersecurity_technical",
				Weight:          0.25,
				CalculationType: "direct",
				DataSources:     []string{"patch_management", "vulnerability_management"},
				Thresholds: map[RiskLevel]float64{
					RiskLevelLow:      0.95, // 95%
					RiskLevelMedium:   0.85, // 85%
					RiskLevelHigh:     0.75, // 75%
					RiskLevelCritical: 0.65, // 65%
				},
				Unit:      "percentage",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// Register all categories
	registry.RegisterCategory(financialCategory)
	registry.RegisterCategory(operationalCategory)
	registry.RegisterCategory(regulatoryCategory)
	registry.RegisterCategory(reputationalCategory)
	registry.RegisterCategory(cybersecurityCategory)

	return registry
}
