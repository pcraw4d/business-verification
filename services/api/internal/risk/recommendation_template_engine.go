package risk

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// RecommendationTemplateEngine generates recommendations from templates
type RecommendationTemplateEngine struct {
	logger    *zap.Logger
	templates map[string]RecommendationTemplate
}

// RecommendationTemplate represents a template for generating recommendations
type RecommendationTemplate struct {
	ID                  string                  `json:"id"`
	Name                string                  `json:"name"`
	Category            RiskCategory            `json:"category"`
	TitleTemplate       string                  `json:"title_template"`
	DescriptionTemplate string                  `json:"description_template"`
	ActionItems         []ActionItemTemplate    `json:"action_items"`
	SuccessMetrics      []SuccessMetricTemplate `json:"success_metrics"`
	Resources           []ResourceTemplate      `json:"resources"`
	ComplianceNotes     []string                `json:"compliance_notes"`
	DefaultPriority     RiskLevel               `json:"default_priority"`
	DefaultImpact       string                  `json:"default_impact"`
	DefaultEffort       string                  `json:"default_effort"`
	DefaultTimeline     string                  `json:"default_timeline"`
	DefaultCost         string                  `json:"default_cost"`
	Metadata            map[string]interface{}  `json:"metadata,omitempty"`
}

// ActionItemTemplate represents a template for action items
type ActionItemTemplate struct {
	DescriptionTemplate string `json:"description_template"`
	AssigneeTemplate    string `json:"assignee_template,omitempty"`
	DueDateTemplate     string `json:"due_date_template,omitempty"`
	PriorityTemplate    string `json:"priority_template,omitempty"`
}

// SuccessMetricTemplate represents a template for success metrics
type SuccessMetricTemplate struct {
	NameTemplate        string `json:"name_template"`
	DescriptionTemplate string `json:"description_template"`
	TargetTemplate      string `json:"target_template"`
	UnitTemplate        string `json:"unit_template"`
	BaselineTemplate    string `json:"baseline_template"`
}

// ResourceTemplate represents a template for resources
type ResourceTemplate struct {
	TypeTemplate         string `json:"type_template"`
	NameTemplate         string `json:"name_template"`
	DescriptionTemplate  string `json:"description_template"`
	CostTemplate         string `json:"cost_template,omitempty"`
	AvailabilityTemplate string `json:"availability_template,omitempty"`
}

// NewRecommendationTemplateEngine creates a new template engine
func NewRecommendationTemplateEngine(logger *zap.Logger) *RecommendationTemplateEngine {
	engine := &RecommendationTemplateEngine{
		logger:    logger,
		templates: make(map[string]RecommendationTemplate),
	}

	// Initialize with default templates
	engine.initializeDefaultTemplates()

	return engine
}

// GenerateRecommendation generates a recommendation from a rule and template
func (rte *RecommendationTemplateEngine) GenerateRecommendation(rule RecommendationRule, riskFactor RiskScore, businessContext map[string]interface{}) (RiskRecommendation, error) {
	// Get the template for the rule's first action
	if len(rule.Actions) == 0 {
		return RiskRecommendation{}, fmt.Errorf("rule has no actions")
	}

	action := rule.Actions[0]
	template, exists := rte.templates[action.Template]
	if !exists {
		return RiskRecommendation{}, fmt.Errorf("template not found: %s", action.Template)
	}

	// Generate recommendation ID
	recommendationID := fmt.Sprintf("rec_%s_%s_%d", rule.ID, riskFactor.FactorID, time.Now().Unix())

	// Create template context
	context := map[string]interface{}{
		"risk_factor":      riskFactor,
		"rule":             rule,
		"business_context": businessContext,
		"timestamp":        time.Now(),
	}

	// Generate recommendation
	recommendation := RiskRecommendation{
		ID:          recommendationID,
		Title:       rte.processTemplate(template.TitleTemplate, context),
		Description: rte.processTemplate(template.DescriptionTemplate, context),
		Category:    template.Category,
		RiskFactor:  riskFactor.FactorID,
		Priority:    action.Priority,
		Impact:      action.Impact,
		Effort:      action.Effort,
		Timeline:    action.Timeline,
		Cost:        action.Cost,
		Confidence:  rule.Confidence,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Generate action items
	for _, itemTemplate := range template.ActionItems {
		actionItem := ActionItem{
			ID:          fmt.Sprintf("action_%s_%d", recommendationID, time.Now().UnixNano()),
			Description: rte.processTemplate(itemTemplate.DescriptionTemplate, context),
			Assignee:    rte.processTemplate(itemTemplate.AssigneeTemplate, context),
			Priority:    rte.processTemplate(itemTemplate.PriorityTemplate, context),
			Status:      "pending",
		}
		recommendation.ActionItems = append(recommendation.ActionItems, actionItem)
	}

	// Generate success metrics
	for _, metricTemplate := range template.SuccessMetrics {
		metric := SuccessMetric{
			Name:        rte.processTemplate(metricTemplate.NameTemplate, context),
			Description: rte.processTemplate(metricTemplate.DescriptionTemplate, context),
			Unit:        rte.processTemplate(metricTemplate.UnitTemplate, context),
		}
		recommendation.SuccessMetrics = append(recommendation.SuccessMetrics, metric)
	}

	// Generate resources
	for _, resourceTemplate := range template.Resources {
		resource := Resource{
			Type:         rte.processTemplate(resourceTemplate.TypeTemplate, context),
			Name:         rte.processTemplate(resourceTemplate.NameTemplate, context),
			Description:  rte.processTemplate(resourceTemplate.DescriptionTemplate, context),
			Availability: rte.processTemplate(resourceTemplate.AvailabilityTemplate, context),
		}
		recommendation.Resources = append(recommendation.Resources, resource)
	}

	// Add compliance notes
	recommendation.ComplianceNotes = template.ComplianceNotes

	return recommendation, nil
}

// processTemplate processes a template string with context variables
func (rte *RecommendationTemplateEngine) processTemplate(template string, context map[string]interface{}) string {
	result := template

	// Replace common variables
	result = strings.ReplaceAll(result, "{{.risk_factor.name}}", rte.getStringValue(context, "risk_factor.name"))
	result = strings.ReplaceAll(result, "{{.risk_factor.score}}", rte.getStringValue(context, "risk_factor.score"))
	result = strings.ReplaceAll(result, "{{.risk_factor.level}}", rte.getStringValue(context, "risk_factor.level"))
	result = strings.ReplaceAll(result, "{{.risk_factor.category}}", rte.getStringValue(context, "risk_factor.category"))

	result = strings.ReplaceAll(result, "{{.rule.name}}", rte.getStringValue(context, "rule.name"))
	result = strings.ReplaceAll(result, "{{.rule.description}}", rte.getStringValue(context, "rule.description"))

	// Replace business context variables
	if businessContext, exists := context["business_context"]; exists {
		if contextMap, ok := businessContext.(map[string]interface{}); ok {
			for key, value := range contextMap {
				placeholder := fmt.Sprintf("{{.business_context.%s}}", key)
				result = strings.ReplaceAll(result, placeholder, rte.convertToString(value))
			}
		}
	}

	// Replace timestamp
	result = strings.ReplaceAll(result, "{{.timestamp}}", time.Now().Format("2006-01-02"))

	return result
}

// getStringValue gets a string value from nested context
func (rte *RecommendationTemplateEngine) getStringValue(context map[string]interface{}, path string) string {
	parts := strings.Split(path, ".")

	current := context
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part, return the value
			if value, exists := current[part]; exists {
				return rte.convertToString(value)
			}
		} else {
			// Navigate deeper
			if next, exists := current[part]; exists {
				if nextMap, ok := next.(map[string]interface{}); ok {
					current = nextMap
				} else {
					return ""
				}
			} else {
				return ""
			}
		}
	}

	return ""
}

// convertToString converts various types to string
func (rte *RecommendationTemplateEngine) convertToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.2f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// initializeDefaultTemplates initializes the engine with default templates
func (rte *RecommendationTemplateEngine) initializeDefaultTemplates() {
	// Financial templates
	rte.addTemplate(RecommendationTemplate{
		ID:                  "financial_cash_flow_template",
		Name:                "Cash Flow Improvement Template",
		Category:            RiskCategoryFinancial,
		TitleTemplate:       "Improve Cash Flow Management for {{.risk_factor.name}}",
		DescriptionTemplate: "Implement comprehensive cash flow monitoring and improvement strategies to address the {{.risk_factor.level}} risk in {{.risk_factor.name}} (Score: {{.risk_factor.score}}).",
		ActionItems: []ActionItemTemplate{
			{
				DescriptionTemplate: "Implement daily cash flow monitoring dashboard",
				AssigneeTemplate:    "Finance Team",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "high",
			},
			{
				DescriptionTemplate: "Establish cash flow forecasting process",
				AssigneeTemplate:    "CFO",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "medium",
			},
		},
		SuccessMetrics: []SuccessMetricTemplate{
			{
				NameTemplate:        "Cash Flow Coverage Ratio",
				DescriptionTemplate: "Monthly cash flow coverage ratio",
				TargetTemplate:      "1.5",
				UnitTemplate:        "ratio",
			},
		},
		DefaultPriority: RiskLevelHigh,
		DefaultImpact:   "high",
		DefaultEffort:   "medium",
		DefaultTimeline: "30-60 days",
		DefaultCost:     "low",
	})

	rte.addTemplate(RecommendationTemplate{
		ID:                  "financial_debt_reduction_template",
		Name:                "Debt Reduction Template",
		Category:            RiskCategoryFinancial,
		TitleTemplate:       "Reduce Debt Burden for {{.risk_factor.name}}",
		DescriptionTemplate: "Develop and implement debt reduction strategies to address the {{.risk_factor.level}} risk in {{.risk_factor.name}} (Score: {{.risk_factor.score}}).",
		ActionItems: []ActionItemTemplate{
			{
				DescriptionTemplate: "Conduct comprehensive debt analysis",
				AssigneeTemplate:    "Finance Team",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "critical",
			},
			{
				DescriptionTemplate: "Negotiate with creditors for better terms",
				AssigneeTemplate:    "CFO",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "high",
			},
		},
		SuccessMetrics: []SuccessMetricTemplate{
			{
				NameTemplate:        "Debt-to-Equity Ratio",
				DescriptionTemplate: "Monthly debt-to-equity ratio",
				TargetTemplate:      "1.5",
				UnitTemplate:        "ratio",
			},
		},
		DefaultPriority: RiskLevelCritical,
		DefaultImpact:   "high",
		DefaultEffort:   "high",
		DefaultTimeline: "90-180 days",
		DefaultCost:     "medium",
	})

	// Operational templates
	rte.addTemplate(RecommendationTemplate{
		ID:                  "operational_process_template",
		Name:                "Process Improvement Template",
		Category:            RiskCategoryOperational,
		TitleTemplate:       "Improve Operational Processes for {{.risk_factor.name}}",
		DescriptionTemplate: "Streamline and optimize operational processes to address the {{.risk_factor.level}} risk in {{.risk_factor.name}} (Score: {{.risk_factor.score}}).",
		ActionItems: []ActionItemTemplate{
			{
				DescriptionTemplate: "Map current operational processes",
				AssigneeTemplate:    "Operations Team",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "medium",
			},
			{
				DescriptionTemplate: "Identify process bottlenecks and inefficiencies",
				AssigneeTemplate:    "Process Improvement Team",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "medium",
			},
		},
		SuccessMetrics: []SuccessMetricTemplate{
			{
				NameTemplate:        "Process Efficiency Score",
				DescriptionTemplate: "Monthly process efficiency measurement",
				TargetTemplate:      "85",
				UnitTemplate:        "percentage",
			},
		},
		DefaultPriority: RiskLevelMedium,
		DefaultImpact:   "medium",
		DefaultEffort:   "medium",
		DefaultTimeline: "60-90 days",
		DefaultCost:     "low",
	})

	// Regulatory templates
	rte.addTemplate(RecommendationTemplate{
		ID:                  "regulatory_compliance_template",
		Name:                "Compliance Improvement Template",
		Category:            RiskCategoryRegulatory,
		TitleTemplate:       "Enhance Regulatory Compliance for {{.risk_factor.name}}",
		DescriptionTemplate: "Improve compliance with regulatory requirements to address the {{.risk_factor.level}} risk in {{.risk_factor.name}} (Score: {{.risk_factor.score}}).",
		ActionItems: []ActionItemTemplate{
			{
				DescriptionTemplate: "Conduct compliance gap analysis",
				AssigneeTemplate:    "Compliance Team",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "high",
			},
			{
				DescriptionTemplate: "Develop compliance improvement plan",
				AssigneeTemplate:    "Legal Team",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "high",
			},
		},
		SuccessMetrics: []SuccessMetricTemplate{
			{
				NameTemplate:        "Compliance Score",
				DescriptionTemplate: "Monthly compliance assessment score",
				TargetTemplate:      "95",
				UnitTemplate:        "percentage",
			},
		},
		ComplianceNotes: []string{
			"Ensure all regulatory requirements are met",
			"Document compliance procedures",
			"Schedule regular compliance audits",
		},
		DefaultPriority: RiskLevelHigh,
		DefaultImpact:   "high",
		DefaultEffort:   "high",
		DefaultTimeline: "120-180 days",
		DefaultCost:     "high",
	})

	// Cybersecurity templates
	rte.addTemplate(RecommendationTemplate{
		ID:                  "cybersecurity_audit_template",
		Name:                "Security Audit Template",
		Category:            RiskCategoryCybersecurity,
		TitleTemplate:       "Conduct Security Audit for {{.risk_factor.name}}",
		DescriptionTemplate: "Perform comprehensive security audit and assessment to address the {{.risk_factor.level}} risk in {{.risk_factor.name}} (Score: {{.risk_factor.score}}).",
		ActionItems: []ActionItemTemplate{
			{
				DescriptionTemplate: "Engage third-party security auditor",
				AssigneeTemplate:    "IT Security Team",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "high",
			},
			{
				DescriptionTemplate: "Review and update security policies",
				AssigneeTemplate:    "CISO",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "high",
			},
		},
		SuccessMetrics: []SuccessMetricTemplate{
			{
				NameTemplate:        "Security Score",
				DescriptionTemplate: "Monthly security assessment score",
				TargetTemplate:      "90",
				UnitTemplate:        "percentage",
			},
		},
		DefaultPriority: RiskLevelHigh,
		DefaultImpact:   "high",
		DefaultEffort:   "high",
		DefaultTimeline: "90-120 days",
		DefaultCost:     "high",
	})

	// Reputational templates
	rte.addTemplate(RecommendationTemplate{
		ID:                  "reputational_sentiment_template",
		Name:                "Sentiment Improvement Template",
		Category:            RiskCategoryReputational,
		TitleTemplate:       "Improve Public Sentiment for {{.risk_factor.name}}",
		DescriptionTemplate: "Address negative public sentiment and improve reputation to address the {{.risk_factor.level}} risk in {{.risk_factor.name}} (Score: {{.risk_factor.score}}).",
		ActionItems: []ActionItemTemplate{
			{
				DescriptionTemplate: "Monitor online sentiment and mentions",
				AssigneeTemplate:    "Marketing Team",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "medium",
			},
			{
				DescriptionTemplate: "Develop reputation management strategy",
				AssigneeTemplate:    "PR Team",
				DueDateTemplate:     "{{.timestamp}}",
				PriorityTemplate:    "medium",
			},
		},
		SuccessMetrics: []SuccessMetricTemplate{
			{
				NameTemplate:        "Sentiment Score",
				DescriptionTemplate: "Monthly sentiment analysis score",
				TargetTemplate:      "0.7",
				UnitTemplate:        "score",
			},
		},
		DefaultPriority: RiskLevelMedium,
		DefaultImpact:   "medium",
		DefaultEffort:   "medium",
		DefaultTimeline: "60-90 days",
		DefaultCost:     "medium",
	})
}

// addTemplate adds a template to the engine
func (rte *RecommendationTemplateEngine) addTemplate(template RecommendationTemplate) {
	rte.templates[template.ID] = template
}

// GetTemplate returns a template by ID
func (rte *RecommendationTemplateEngine) GetTemplate(templateID string) (*RecommendationTemplate, error) {
	template, exists := rte.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	return &template, nil
}

// AddCustomTemplate adds a custom template
func (rte *RecommendationTemplateEngine) AddCustomTemplate(template RecommendationTemplate) error {
	// Validate template
	if err := rte.validateTemplate(template); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	rte.addTemplate(template)
	rte.logger.Info("Added custom recommendation template",
		zap.String("template_id", template.ID),
		zap.String("category", string(template.Category)))

	return nil
}

// validateTemplate validates a recommendation template
func (rte *RecommendationTemplateEngine) validateTemplate(template RecommendationTemplate) error {
	if template.ID == "" {
		return fmt.Errorf("template ID is required")
	}

	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if template.Category == "" {
		return fmt.Errorf("template category is required")
	}

	if template.TitleTemplate == "" {
		return fmt.Errorf("title template is required")
	}

	if template.DescriptionTemplate == "" {
		return fmt.Errorf("description template is required")
	}

	return nil
}

// GetAllTemplates returns all templates
func (rte *RecommendationTemplateEngine) GetAllTemplates() map[string]RecommendationTemplate {
	return rte.templates
}

// GetTemplatesByCategory returns templates for a specific category
func (rte *RecommendationTemplateEngine) GetTemplatesByCategory(category RiskCategory) map[string]RecommendationTemplate {
	templates := make(map[string]RecommendationTemplate)

	for id, template := range rte.templates {
		if template.Category == category {
			templates[id] = template
		}
	}

	return templates
}
