package risk

import (
	"fmt"

	"go.uber.org/zap"
)

// RecommendationRuleEngine manages recommendation rules
type RecommendationRuleEngine struct {
	logger *zap.Logger
	rules  map[string][]RecommendationRule
}

// RecommendationRule represents a rule for generating recommendations
type RecommendationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    RiskCategory           `json:"category"`
	RiskLevels  []RiskLevel            `json:"risk_levels"`
	Conditions  []RuleCondition        `json:"conditions"`
	Actions     []RuleAction           `json:"actions"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RuleCondition represents a condition for a rule
type RuleCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Required bool        `json:"required"`
}

// RuleAction represents an action to take when a rule matches
type RuleAction struct {
	Type       string                 `json:"type"`
	Template   string                 `json:"template"`
	Parameters map[string]interface{} `json:"parameters"`
	Priority   RiskLevel              `json:"priority"`
	Impact     string                 `json:"impact"`
	Effort     string                 `json:"effort"`
	Timeline   string                 `json:"timeline"`
	Cost       string                 `json:"cost"`
}

// NewRecommendationRuleEngine creates a new rule engine
func NewRecommendationRuleEngine(logger *zap.Logger) *RecommendationRuleEngine {
	engine := &RecommendationRuleEngine{
		logger: logger,
		rules:  make(map[string][]RecommendationRule),
	}

	// Initialize with default rules
	engine.initializeDefaultRules()

	return engine
}

// GetApplicableRules returns rules applicable to a risk category and level
func (rre *RecommendationRuleEngine) GetApplicableRules(category RiskCategory, level RiskLevel) []RecommendationRule {
	var applicableRules []RecommendationRule

	// Get rules for the category
	categoryRules, exists := rre.rules[string(category)]
	if !exists {
		return applicableRules
	}

	// Filter by risk level and enabled status
	for _, rule := range categoryRules {
		if rule.Enabled && rre.isRiskLevelApplicable(rule, level) {
			applicableRules = append(applicableRules, rule)
		}
	}

	// Sort by priority (higher priority first)
	for i := 0; i < len(applicableRules)-1; i++ {
		for j := i + 1; j < len(applicableRules); j++ {
			if applicableRules[i].Priority < applicableRules[j].Priority {
				applicableRules[i], applicableRules[j] = applicableRules[j], applicableRules[i]
			}
		}
	}

	return applicableRules
}

// isRiskLevelApplicable checks if a rule applies to a specific risk level
func (rre *RecommendationRuleEngine) isRiskLevelApplicable(rule RecommendationRule, level RiskLevel) bool {
	// If no specific levels defined, apply to all
	if len(rule.RiskLevels) == 0 {
		return true
	}

	// Check if the level is in the rule's applicable levels
	for _, ruleLevel := range rule.RiskLevels {
		if ruleLevel == level {
			return true
		}
	}

	return false
}

// initializeDefaultRules initializes the engine with default recommendation rules
func (rre *RecommendationRuleEngine) initializeDefaultRules() {
	// Financial Risk Rules
	rre.addRule(RecommendationRule{
		ID:          "financial_cash_flow_improvement",
		Name:        "Improve Cash Flow Management",
		Description: "Implement cash flow monitoring and improvement strategies",
		Category:    RiskCategoryFinancial,
		RiskLevels:  []RiskLevel{RiskLevelHigh, RiskLevelCritical},
		Conditions: []RuleCondition{
			{Field: "cash_flow_coverage", Operator: "<", Value: 1.5, Required: true},
		},
		Actions: []RuleAction{
			{
				Type:     "recommendation",
				Template: "financial_cash_flow_template",
				Priority: RiskLevelHigh,
				Impact:   "high",
				Effort:   "medium",
				Timeline: "30-60 days",
				Cost:     "low",
			},
		},
		Priority:   100,
		Enabled:    true,
		Confidence: 0.9,
	})

	rre.addRule(RecommendationRule{
		ID:          "financial_debt_reduction",
		Name:        "Reduce Debt Burden",
		Description: "Develop debt reduction and management strategies",
		Category:    RiskCategoryFinancial,
		RiskLevels:  []RiskLevel{RiskLevelHigh, RiskLevelCritical},
		Conditions: []RuleCondition{
			{Field: "debt_to_equity_ratio", Operator: ">", Value: 2.0, Required: true},
		},
		Actions: []RuleAction{
			{
				Type:     "recommendation",
				Template: "financial_debt_reduction_template",
				Priority: RiskLevelCritical,
				Impact:   "high",
				Effort:   "high",
				Timeline: "90-180 days",
				Cost:     "medium",
			},
		},
		Priority:   90,
		Enabled:    true,
		Confidence: 0.85,
	})

	// Operational Risk Rules
	rre.addRule(RecommendationRule{
		ID:          "operational_process_improvement",
		Name:        "Improve Operational Processes",
		Description: "Streamline and optimize operational processes",
		Category:    RiskCategoryOperational,
		RiskLevels:  []RiskLevel{RiskLevelMedium, RiskLevelHigh},
		Conditions: []RuleCondition{
			{Field: "operational_efficiency", Operator: "<", Value: 0.7, Required: true},
		},
		Actions: []RuleAction{
			{
				Type:     "recommendation",
				Template: "operational_process_template",
				Priority: RiskLevelMedium,
				Impact:   "medium",
				Effort:   "medium",
				Timeline: "60-90 days",
				Cost:     "low",
			},
		},
		Priority:   80,
		Enabled:    true,
		Confidence: 0.8,
	})

	rre.addRule(RecommendationRule{
		ID:          "operational_workforce_stability",
		Name:        "Improve Workforce Stability",
		Description: "Address high turnover and workforce instability",
		Category:    RiskCategoryOperational,
		RiskLevels:  []RiskLevel{RiskLevelHigh, RiskLevelCritical},
		Conditions: []RuleCondition{
			{Field: "employee_turnover_rate", Operator: ">", Value: 0.2, Required: true},
		},
		Actions: []RuleAction{
			{
				Type:     "recommendation",
				Template: "operational_workforce_template",
				Priority: RiskLevelHigh,
				Impact:   "high",
				Effort:   "high",
				Timeline: "90-120 days",
				Cost:     "high",
			},
		},
		Priority:   85,
		Enabled:    true,
		Confidence: 0.85,
	})

	// Regulatory Risk Rules
	rre.addRule(RecommendationRule{
		ID:          "regulatory_compliance_improvement",
		Name:        "Enhance Regulatory Compliance",
		Description: "Improve compliance with regulatory requirements",
		Category:    RiskCategoryRegulatory,
		RiskLevels:  []RiskLevel{RiskLevelMedium, RiskLevelHigh, RiskLevelCritical},
		Conditions: []RuleCondition{
			{Field: "compliance_score", Operator: "<", Value: 0.8, Required: true},
		},
		Actions: []RuleAction{
			{
				Type:     "recommendation",
				Template: "regulatory_compliance_template",
				Priority: RiskLevelHigh,
				Impact:   "high",
				Effort:   "high",
				Timeline: "120-180 days",
				Cost:     "high",
			},
		},
		Priority:   95,
		Enabled:    true,
		Confidence: 0.9,
	})

	rre.addRule(RecommendationRule{
		ID:          "regulatory_license_renewal",
		Name:        "Ensure License Renewal",
		Description: "Address license renewal requirements and deadlines",
		Category:    RiskCategoryRegulatory,
		RiskLevels:  []RiskLevel{RiskLevelHigh, RiskLevelCritical},
		Conditions: []RuleCondition{
			{Field: "license_expiry_days", Operator: "<", Value: 90, Required: true},
		},
		Actions: []RuleAction{
			{
				Type:     "recommendation",
				Template: "regulatory_license_template",
				Priority: RiskLevelCritical,
				Impact:   "critical",
				Effort:   "medium",
				Timeline: "30-60 days",
				Cost:     "medium",
			},
		},
		Priority:   100,
		Enabled:    true,
		Confidence: 0.95,
	})

	// Reputational Risk Rules
	rre.addRule(RecommendationRule{
		ID:          "reputational_sentiment_improvement",
		Name:        "Improve Public Sentiment",
		Description: "Address negative public sentiment and improve reputation",
		Category:    RiskCategoryReputational,
		RiskLevels:  []RiskLevel{RiskLevelMedium, RiskLevelHigh},
		Conditions: []RuleCondition{
			{Field: "sentiment_score", Operator: "<", Value: 0.5, Required: true},
		},
		Actions: []RuleAction{
			{
				Type:     "recommendation",
				Template: "reputational_sentiment_template",
				Priority: RiskLevelMedium,
				Impact:   "medium",
				Effort:   "medium",
				Timeline: "60-90 days",
				Cost:     "medium",
			},
		},
		Priority:   75,
		Enabled:    true,
		Confidence: 0.8,
	})

	rre.addRule(RecommendationRule{
		ID:          "reputational_crisis_management",
		Name:        "Implement Crisis Management",
		Description: "Develop and implement crisis management procedures",
		Category:    RiskCategoryReputational,
		RiskLevels:  []RiskLevel{RiskLevelHigh, RiskLevelCritical},
		Conditions: []RuleCondition{
			{Field: "negative_mentions", Operator: ">", Value: 10, Required: true},
		},
		Actions: []RuleAction{
			{
				Type:     "recommendation",
				Template: "reputational_crisis_template",
				Priority: RiskLevelCritical,
				Impact:   "high",
				Effort:   "high",
				Timeline: "30-60 days",
				Cost:     "high",
			},
		},
		Priority:   90,
		Enabled:    true,
		Confidence: 0.85,
	})

	// Cybersecurity Risk Rules
	rre.addRule(RecommendationRule{
		ID:          "cybersecurity_security_audit",
		Name:        "Conduct Security Audit",
		Description: "Perform comprehensive security audit and assessment",
		Category:    RiskCategoryCybersecurity,
		RiskLevels:  []RiskLevel{RiskLevelMedium, RiskLevelHigh, RiskLevelCritical},
		Conditions: []RuleCondition{
			{Field: "security_score", Operator: "<", Value: 0.7, Required: true},
		},
		Actions: []RuleAction{
			{
				Type:     "recommendation",
				Template: "cybersecurity_audit_template",
				Priority: RiskLevelHigh,
				Impact:   "high",
				Effort:   "high",
				Timeline: "90-120 days",
				Cost:     "high",
			},
		},
		Priority:   90,
		Enabled:    true,
		Confidence: 0.9,
	})

	rre.addRule(RecommendationRule{
		ID:          "cybersecurity_incident_response",
		Name:        "Improve Incident Response",
		Description: "Enhance incident response capabilities and procedures",
		Category:    RiskCategoryCybersecurity,
		RiskLevels:  []RiskLevel{RiskLevelHigh, RiskLevelCritical},
		Conditions: []RuleCondition{
			{Field: "incident_response_time", Operator: ">", Value: 24, Required: true},
		},
		Actions: []RuleAction{
			{
				Type:     "recommendation",
				Template: "cybersecurity_incident_template",
				Priority: RiskLevelCritical,
				Impact:   "high",
				Effort:   "medium",
				Timeline: "60-90 days",
				Cost:     "medium",
			},
		},
		Priority:   95,
		Enabled:    true,
		Confidence: 0.9,
	})
}

// addRule adds a rule to the engine
func (rre *RecommendationRuleEngine) addRule(rule RecommendationRule) {
	category := string(rule.Category)
	if rre.rules[category] == nil {
		rre.rules[category] = []RecommendationRule{}
	}
	rre.rules[category] = append(rre.rules[category], rule)
}

// AddCustomRule adds a custom rule to the engine
func (rre *RecommendationRuleEngine) AddCustomRule(rule RecommendationRule) error {
	// Validate rule
	if err := rre.validateRule(rule); err != nil {
		return fmt.Errorf("invalid rule: %w", err)
	}

	rre.addRule(rule)
	rre.logger.Info("Added custom recommendation rule",
		zap.String("rule_id", rule.ID),
		zap.String("category", string(rule.Category)))

	return nil
}

// validateRule validates a recommendation rule
func (rre *RecommendationRuleEngine) validateRule(rule RecommendationRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}

	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if rule.Category == "" {
		return fmt.Errorf("rule category is required")
	}

	if len(rule.Actions) == 0 {
		return fmt.Errorf("rule must have at least one action")
	}

	if rule.Confidence < 0 || rule.Confidence > 1 {
		return fmt.Errorf("rule confidence must be between 0 and 1")
	}

	return nil
}

// GetRule returns a specific rule by ID
func (rre *RecommendationRuleEngine) GetRule(ruleID string) (*RecommendationRule, error) {
	for _, categoryRules := range rre.rules {
		for _, rule := range categoryRules {
			if rule.ID == ruleID {
				return &rule, nil
			}
		}
	}

	return nil, fmt.Errorf("rule not found: %s", ruleID)
}

// UpdateRule updates an existing rule
func (rre *RecommendationRuleEngine) UpdateRule(ruleID string, updatedRule RecommendationRule) error {
	// Validate updated rule
	if err := rre.validateRule(updatedRule); err != nil {
		return fmt.Errorf("invalid updated rule: %w", err)
	}

	// Find and update the rule
	for category, categoryRules := range rre.rules {
		for i, rule := range categoryRules {
			if rule.ID == ruleID {
				rre.rules[category][i] = updatedRule
				rre.logger.Info("Updated recommendation rule",
					zap.String("rule_id", ruleID),
					zap.String("category", category))
				return nil
			}
		}
	}

	return fmt.Errorf("rule not found: %s", ruleID)
}

// DeleteRule deletes a rule by ID
func (rre *RecommendationRuleEngine) DeleteRule(ruleID string) error {
	for category, categoryRules := range rre.rules {
		for i, rule := range categoryRules {
			if rule.ID == ruleID {
				// Remove the rule
				rre.rules[category] = append(categoryRules[:i], categoryRules[i+1:]...)
				rre.logger.Info("Deleted recommendation rule",
					zap.String("rule_id", ruleID),
					zap.String("category", category))
				return nil
			}
		}
	}

	return fmt.Errorf("rule not found: %s", ruleID)
}

// GetRulesByCategory returns all rules for a specific category
func (rre *RecommendationRuleEngine) GetRulesByCategory(category RiskCategory) []RecommendationRule {
	rules, exists := rre.rules[string(category)]
	if !exists {
		return []RecommendationRule{}
	}
	return rules
}

// GetAllRules returns all rules
func (rre *RecommendationRuleEngine) GetAllRules() map[string][]RecommendationRule {
	return rre.rules
}

// EnableRule enables a rule by ID
func (rre *RecommendationRuleEngine) EnableRule(ruleID string) error {
	return rre.setRuleEnabled(ruleID, true)
}

// DisableRule disables a rule by ID
func (rre *RecommendationRuleEngine) DisableRule(ruleID string) error {
	return rre.setRuleEnabled(ruleID, false)
}

// setRuleEnabled sets the enabled status of a rule
func (rre *RecommendationRuleEngine) setRuleEnabled(ruleID string, enabled bool) error {
	for category, categoryRules := range rre.rules {
		for i, rule := range categoryRules {
			if rule.ID == ruleID {
				rre.rules[category][i].Enabled = enabled
				status := "disabled"
				if enabled {
					status = "enabled"
				}
				rre.logger.Info("Rule status changed",
					zap.String("rule_id", ruleID),
					zap.String("status", status))
				return nil
			}
		}
	}

	return fmt.Errorf("rule not found: %s", ruleID)
}
