package risk

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// AlertThresholdManager manages alert thresholds
type AlertThresholdManager struct {
	logger     *zap.Logger
	thresholds map[string][]AlertThreshold
}

// AlertThreshold represents a threshold for triggering alerts
type AlertThreshold struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Description          string                 `json:"description"`
	FactorID             string                 `json:"factor_id"`
	Category             RiskCategory           `json:"category"`
	BusinessID           string                 `json:"business_id,omitempty"` // Empty for global thresholds
	Operator             string                 `json:"operator"`              // >, >=, <, <=, ==, !=
	Value                float64                `json:"value"`
	RiskLevel            RiskLevel              `json:"risk_level"`
	Severity             AlertSeverity          `json:"severity"`
	Priority             AlertPriority          `json:"priority"`
	Enabled              bool                   `json:"enabled"`
	ExpirationHours      int                    `json:"expiration_hours"` // 0 = no expiration
	NotificationChannels []string               `json:"notification_channels"`
	Tags                 []string               `json:"tags"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// NewAlertThresholdManager creates a new threshold manager
func NewAlertThresholdManager(logger *zap.Logger) *AlertThresholdManager {
	manager := &AlertThresholdManager{
		logger:     logger,
		thresholds: make(map[string][]AlertThreshold),
	}

	// Initialize with default thresholds
	manager.initializeDefaultThresholds()

	return manager
}

// GetThresholds returns thresholds for a specific factor or category
func (atm *AlertThresholdManager) GetThresholds(factorID, businessID string) ([]AlertThreshold, error) {
	// Try business-specific thresholds first
	businessKey := fmt.Sprintf("%s_%s", businessID, factorID)
	if thresholds, exists := atm.thresholds[businessKey]; exists {
		return atm.filterEnabledThresholds(thresholds), nil
	}

	// Fall back to global thresholds
	if thresholds, exists := atm.thresholds[factorID]; exists {
		return atm.filterEnabledThresholds(thresholds), nil
	}

	// Return empty slice if no thresholds found
	return []AlertThreshold{}, nil
}

// filterEnabledThresholds filters out disabled thresholds
func (atm *AlertThresholdManager) filterEnabledThresholds(thresholds []AlertThreshold) []AlertThreshold {
	var enabled []AlertThreshold
	for _, threshold := range thresholds {
		if threshold.Enabled {
			enabled = append(enabled, threshold)
		}
	}
	return enabled
}

// AddThreshold adds a new threshold
func (atm *AlertThresholdManager) AddThreshold(threshold AlertThreshold) error {
	// Validate threshold
	if err := atm.validateThreshold(threshold); err != nil {
		return fmt.Errorf("invalid threshold: %w", err)
	}

	// Set timestamps
	now := time.Now()
	threshold.CreatedAt = now
	threshold.UpdatedAt = now

	// Determine key
	key := threshold.FactorID
	if threshold.BusinessID != "" {
		key = fmt.Sprintf("%s_%s", threshold.BusinessID, threshold.FactorID)
	}

	// Add threshold
	atm.thresholds[key] = append(atm.thresholds[key], threshold)

	atm.logger.Info("Added alert threshold",
		zap.String("threshold_id", threshold.ID),
		zap.String("factor_id", threshold.FactorID),
		zap.String("business_id", threshold.BusinessID))

	return nil
}

// UpdateThreshold updates an existing threshold
func (atm *AlertThresholdManager) UpdateThreshold(thresholdID string, updatedThreshold AlertThreshold) error {
	// Validate updated threshold
	if err := atm.validateThreshold(updatedThreshold); err != nil {
		return fmt.Errorf("invalid updated threshold: %w", err)
	}

	// Find and update threshold
	for key, thresholds := range atm.thresholds {
		for i, threshold := range thresholds {
			if threshold.ID == thresholdID {
				updatedThreshold.UpdatedAt = time.Now()
				atm.thresholds[key][i] = updatedThreshold

				atm.logger.Info("Updated alert threshold",
					zap.String("threshold_id", thresholdID))
				return nil
			}
		}
	}

	return fmt.Errorf("threshold not found: %s", thresholdID)
}

// DeleteThreshold deletes a threshold
func (atm *AlertThresholdManager) DeleteThreshold(thresholdID string) error {
	for key, thresholds := range atm.thresholds {
		for i, threshold := range thresholds {
			if threshold.ID == thresholdID {
				// Remove threshold
				atm.thresholds[key] = append(thresholds[:i], thresholds[i+1:]...)

				atm.logger.Info("Deleted alert threshold",
					zap.String("threshold_id", thresholdID))
				return nil
			}
		}
	}

	return fmt.Errorf("threshold not found: %s", thresholdID)
}

// validateThreshold validates a threshold
func (atm *AlertThresholdManager) validateThreshold(threshold AlertThreshold) error {
	if threshold.ID == "" {
		return fmt.Errorf("threshold ID is required")
	}

	if threshold.Name == "" {
		return fmt.Errorf("threshold name is required")
	}

	if threshold.FactorID == "" {
		return fmt.Errorf("factor ID is required")
	}

	if threshold.Operator == "" {
		return fmt.Errorf("operator is required")
	}

	// Validate operator
	validOperators := []string{">", ">=", "<", "<=", "==", "!="}
	valid := false
	for _, op := range validOperators {
		if threshold.Operator == op {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid operator: %s", threshold.Operator)
	}

	return nil
}

// initializeDefaultThresholds initializes default thresholds
func (atm *AlertThresholdManager) initializeDefaultThresholds() {
	// Overall risk thresholds
	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "overall_risk_critical",
		Name:                 "Overall Risk Critical",
		Description:          "Triggers when overall risk score reaches critical level",
		FactorID:             "overall_risk",
		Category:             RiskCategoryOperational,
		Operator:             ">=",
		Value:                75.0,
		RiskLevel:            RiskLevelCritical,
		Severity:             AlertSeverityCritical,
		Priority:             AlertPriorityUrgent,
		Enabled:              true,
		ExpirationHours:      0,
		NotificationChannels: []string{"email", "dashboard"},
		Tags:                 []string{"overall", "critical"},
	})

	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "overall_risk_high",
		Name:                 "Overall Risk High",
		Description:          "Triggers when overall risk score reaches high level",
		FactorID:             "overall_risk",
		Category:             RiskCategoryOperational,
		Operator:             ">=",
		Value:                60.0,
		RiskLevel:            RiskLevelHigh,
		Severity:             AlertSeverityHigh,
		Priority:             AlertPriorityHigh,
		Enabled:              true,
		ExpirationHours:      24,
		NotificationChannels: []string{"email", "dashboard"},
		Tags:                 []string{"overall", "high"},
	})

	// Financial risk thresholds
	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "cash_flow_critical",
		Name:                 "Cash Flow Critical",
		Description:          "Triggers when cash flow coverage ratio is critically low",
		FactorID:             "cash_flow_coverage",
		Category:             RiskCategoryFinancial,
		Operator:             "<=",
		Value:                1.0,
		RiskLevel:            RiskLevelCritical,
		Severity:             AlertSeverityCritical,
		Priority:             AlertPriorityUrgent,
		Enabled:              true,
		ExpirationHours:      0,
		NotificationChannels: []string{"email", "sms", "dashboard"},
		Tags:                 []string{"financial", "cash_flow", "critical"},
	})

	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "debt_ratio_high",
		Name:                 "Debt Ratio High",
		Description:          "Triggers when debt-to-equity ratio is high",
		FactorID:             "debt_to_equity_ratio",
		Category:             RiskCategoryFinancial,
		Operator:             ">=",
		Value:                2.0,
		RiskLevel:            RiskLevelHigh,
		Severity:             AlertSeverityHigh,
		Priority:             AlertPriorityHigh,
		Enabled:              true,
		ExpirationHours:      48,
		NotificationChannels: []string{"email", "dashboard"},
		Tags:                 []string{"financial", "debt", "high"},
	})

	// Operational risk thresholds
	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "turnover_high",
		Name:                 "Employee Turnover High",
		Description:          "Triggers when employee turnover rate is high",
		FactorID:             "employee_turnover_rate",
		Category:             RiskCategoryOperational,
		Operator:             ">=",
		Value:                0.25,
		RiskLevel:            RiskLevelHigh,
		Severity:             AlertSeverityHigh,
		Priority:             AlertPriorityHigh,
		Enabled:              true,
		ExpirationHours:      72,
		NotificationChannels: []string{"email", "dashboard"},
		Tags:                 []string{"operational", "turnover", "high"},
	})

	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "efficiency_low",
		Name:                 "Operational Efficiency Low",
		Description:          "Triggers when operational efficiency is low",
		FactorID:             "operational_efficiency",
		Category:             RiskCategoryOperational,
		Operator:             "<=",
		Value:                0.6,
		RiskLevel:            RiskLevelMedium,
		Severity:             AlertSeverityMedium,
		Priority:             AlertPriorityMedium,
		Enabled:              true,
		ExpirationHours:      96,
		NotificationChannels: []string{"email", "dashboard"},
		Tags:                 []string{"operational", "efficiency", "medium"},
	})

	// Regulatory risk thresholds
	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "compliance_low",
		Name:                 "Compliance Score Low",
		Description:          "Triggers when compliance score is low",
		FactorID:             "compliance_score",
		Category:             RiskCategoryRegulatory,
		Operator:             "<=",
		Value:                0.7,
		RiskLevel:            RiskLevelHigh,
		Severity:             AlertSeverityHigh,
		Priority:             AlertPriorityHigh,
		Enabled:              true,
		ExpirationHours:      0,
		NotificationChannels: []string{"email", "sms", "dashboard"},
		Tags:                 []string{"regulatory", "compliance", "high"},
	})

	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "license_expiring",
		Name:                 "License Expiring Soon",
		Description:          "Triggers when license is expiring within 90 days",
		FactorID:             "license_expiry_days",
		Category:             RiskCategoryRegulatory,
		Operator:             "<=",
		Value:                90.0,
		RiskLevel:            RiskLevelHigh,
		Severity:             AlertSeverityHigh,
		Priority:             AlertPriorityHigh,
		Enabled:              true,
		ExpirationHours:      0,
		NotificationChannels: []string{"email", "dashboard"},
		Tags:                 []string{"regulatory", "license", "expiry"},
	})

	// Cybersecurity risk thresholds
	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "security_score_low",
		Name:                 "Security Score Low",
		Description:          "Triggers when security score is low",
		FactorID:             "security_score",
		Category:             RiskCategoryCybersecurity,
		Operator:             "<=",
		Value:                0.6,
		RiskLevel:            RiskLevelHigh,
		Severity:             AlertSeverityHigh,
		Priority:             AlertPriorityHigh,
		Enabled:              true,
		ExpirationHours:      0,
		NotificationChannels: []string{"email", "sms", "dashboard"},
		Tags:                 []string{"cybersecurity", "security", "high"},
	})

	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "incident_response_slow",
		Name:                 "Incident Response Slow",
		Description:          "Triggers when incident response time is too slow",
		FactorID:             "incident_response_time",
		Category:             RiskCategoryCybersecurity,
		Operator:             ">=",
		Value:                24.0,
		RiskLevel:            RiskLevelHigh,
		Severity:             AlertSeverityHigh,
		Priority:             AlertPriorityHigh,
		Enabled:              true,
		ExpirationHours:      0,
		NotificationChannels: []string{"email", "sms", "dashboard"},
		Tags:                 []string{"cybersecurity", "incident", "response"},
	})

	// Reputational risk thresholds
	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "sentiment_negative",
		Name:                 "Negative Sentiment",
		Description:          "Triggers when sentiment score is negative",
		FactorID:             "sentiment_score",
		Category:             RiskCategoryReputational,
		Operator:             "<=",
		Value:                0.3,
		RiskLevel:            RiskLevelMedium,
		Severity:             AlertSeverityMedium,
		Priority:             AlertPriorityMedium,
		Enabled:              true,
		ExpirationHours:      48,
		NotificationChannels: []string{"email", "dashboard"},
		Tags:                 []string{"reputational", "sentiment", "medium"},
	})

	atm.addDefaultThreshold(AlertThreshold{
		ID:                   "negative_mentions_high",
		Name:                 "High Negative Mentions",
		Description:          "Triggers when negative mentions are high",
		FactorID:             "negative_mentions",
		Category:             RiskCategoryReputational,
		Operator:             ">=",
		Value:                10.0,
		RiskLevel:            RiskLevelHigh,
		Severity:             AlertSeverityHigh,
		Priority:             AlertPriorityHigh,
		Enabled:              true,
		ExpirationHours:      24,
		NotificationChannels: []string{"email", "dashboard"},
		Tags:                 []string{"reputational", "mentions", "high"},
	})
}

// addDefaultThreshold adds a default threshold
func (atm *AlertThresholdManager) addDefaultThreshold(threshold AlertThreshold) {
	now := time.Now()
	threshold.CreatedAt = now
	threshold.UpdatedAt = now

	key := threshold.FactorID
	atm.thresholds[key] = append(atm.thresholds[key], threshold)
}

// GetAllThresholds returns all thresholds
func (atm *AlertThresholdManager) GetAllThresholds() map[string][]AlertThreshold {
	return atm.thresholds
}

// GetThresholdsByCategory returns thresholds for a specific category
func (atm *AlertThresholdManager) GetThresholdsByCategory(category RiskCategory) []AlertThreshold {
	var categoryThresholds []AlertThreshold

	for _, thresholds := range atm.thresholds {
		for _, threshold := range thresholds {
			if threshold.Category == category {
				categoryThresholds = append(categoryThresholds, threshold)
			}
		}
	}

	return categoryThresholds
}

// EnableThreshold enables a threshold
func (atm *AlertThresholdManager) EnableThreshold(thresholdID string) error {
	return atm.setThresholdEnabled(thresholdID, true)
}

// DisableThreshold disables a threshold
func (atm *AlertThresholdManager) DisableThreshold(thresholdID string) error {
	return atm.setThresholdEnabled(thresholdID, false)
}

// setThresholdEnabled sets the enabled status of a threshold
func (atm *AlertThresholdManager) setThresholdEnabled(thresholdID string, enabled bool) error {
	for key, thresholds := range atm.thresholds {
		for i, threshold := range thresholds {
			if threshold.ID == thresholdID {
				atm.thresholds[key][i].Enabled = enabled
				atm.thresholds[key][i].UpdatedAt = time.Now()

				status := "disabled"
				if enabled {
					status = "enabled"
				}

				atm.logger.Info("Threshold status changed",
					zap.String("threshold_id", thresholdID),
					zap.String("status", status))
				return nil
			}
		}
	}

	return fmt.Errorf("threshold not found: %s", thresholdID)
}

// GetThreshold returns a specific threshold by ID
func (atm *AlertThresholdManager) GetThreshold(thresholdID string) (*AlertThreshold, error) {
	for _, thresholds := range atm.thresholds {
		for _, threshold := range thresholds {
			if threshold.ID == thresholdID {
				return &threshold, nil
			}
		}
	}

	return nil, fmt.Errorf("threshold not found: %s", thresholdID)
}

// CreateBusinessSpecificThreshold creates a business-specific threshold
func (atm *AlertThresholdManager) CreateBusinessSpecificThreshold(businessID string, baseThreshold AlertThreshold) error {
	// Create a copy of the base threshold
	businessThreshold := baseThreshold
	businessThreshold.ID = fmt.Sprintf("%s_%s", businessID, baseThreshold.ID)
	businessThreshold.BusinessID = businessID
	businessThreshold.CreatedAt = time.Now()
	businessThreshold.UpdatedAt = time.Now()

	return atm.AddThreshold(businessThreshold)
}

// GetBusinessThresholds returns all thresholds for a specific business
func (atm *AlertThresholdManager) GetBusinessThresholds(businessID string) []AlertThreshold {
	var businessThresholds []AlertThreshold

	for key, thresholds := range atm.thresholds {
		// Check if this is a business-specific threshold
		if len(key) > len(businessID)+1 && key[:len(businessID)+1] == businessID+"_" {
			businessThresholds = append(businessThresholds, thresholds...)
		}
	}

	return businessThresholds
}
