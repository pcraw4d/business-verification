package compliance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// AlertSystem provides comprehensive compliance alert management
type AlertSystem struct {
	logger        *observability.Logger
	statusSystem  *ComplianceStatusSystem
	checkEngine   *CheckEngine
	rules         map[string]*AlertRule
	escalations   map[string]*EscalationPolicy
	notifications map[string]*NotificationChannel
	mu            sync.RWMutex
}

// AlertRule defines when and how alerts should be generated
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	EntityType  string                 `json:"entity_type"` // "overall", "framework", "requirement", "control"
	Conditions  []AlertCondition       `json:"conditions"`
	Severity    string                 `json:"severity"` // "low", "medium", "high", "critical"
	Actions     []AlertAction          `json:"actions"`
	Suppression *AlertSuppression      `json:"suppression,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AlertCondition defines a condition that triggers an alert
type AlertCondition struct {
	Type     string         `json:"type"`               // "score_below", "score_decline", "status_change", "deadline_missed", "risk_increase"
	Field    string         `json:"field"`              // Field to monitor
	Operator string         `json:"operator"`           // "lt", "lte", "eq", "gte", "gt", "ne"
	Value    interface{}    `json:"value"`              // Threshold value
	Duration *time.Duration `json:"duration,omitempty"` // Duration for sustained conditions
	Window   *time.Duration `json:"window,omitempty"`   // Time window for evaluation
}

// AlertAction defines what happens when an alert is triggered
type AlertAction struct {
	Type       string                 `json:"type"`                  // "create_alert", "send_notification", "escalate", "webhook"
	Parameters map[string]interface{} `json:"parameters"`            // Action-specific parameters
	Delay      *time.Duration         `json:"delay,omitempty"`       // Delay before action
	RetryCount int                    `json:"retry_count"`           // Number of retries
	RetryDelay *time.Duration         `json:"retry_delay,omitempty"` // Delay between retries
}

// AlertSuppression defines how to suppress duplicate alerts
type AlertSuppression struct {
	Enabled   bool          `json:"enabled"`
	Window    time.Duration `json:"window"`     // Time window for suppression
	MaxAlerts int           `json:"max_alerts"` // Maximum alerts in window
	GroupBy   []string      `json:"group_by"`   // Fields to group by for suppression
}

// EscalationPolicy defines how alerts should be escalated
type EscalationPolicy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Levels      []EscalationLevel      `json:"levels"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// EscalationLevel defines a level in the escalation policy
type EscalationLevel struct {
	Level      int                    `json:"level"`
	Name       string                 `json:"name"`
	Delay      time.Duration          `json:"delay"`      // Delay before escalation
	Recipients []string               `json:"recipients"` // Recipients for this level
	Actions    []string               `json:"actions"`    // Actions to take
	Conditions map[string]interface{} `json:"conditions"` // Conditions for escalation
}

// NotificationChannel defines how to send notifications
type NotificationChannel struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // "email", "webhook", "slack", "sms"
	Enabled    bool                   `json:"enabled"`
	Config     map[string]interface{} `json:"config"`     // Channel-specific configuration
	Recipients []string               `json:"recipients"` // Default recipients
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AlertEvaluation represents the evaluation of alert rules
type AlertEvaluation struct {
	RuleID      string                 `json:"rule_id"`
	BusinessID  string                 `json:"business_id"`
	EntityType  string                 `json:"entity_type"`
	EntityID    string                 `json:"entity_id"`
	Triggered   bool                   `json:"triggered"`
	Conditions  []ConditionResult      `json:"conditions"`
	Actions     []ActionResult         `json:"actions"`
	EvaluatedAt time.Time              `json:"evaluated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ConditionResult represents the result of evaluating a condition
type ConditionResult struct {
	Condition   AlertCondition `json:"condition"`
	Triggered   bool           `json:"triggered"`
	Value       interface{}    `json:"value"`
	Message     string         `json:"message"`
	EvaluatedAt time.Time      `json:"evaluated_at"`
}

// ActionResult represents the result of executing an action
type ActionResult struct {
	Action     AlertAction `json:"action"`
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	ExecutedAt time.Time   `json:"executed_at"`
	RetryCount int         `json:"retry_count"`
}

// AlertAnalytics represents analytics about alerts
type AlertAnalytics struct {
	BusinessID            string                 `json:"business_id"`
	Period                string                 `json:"period"`
	TotalAlerts           int                    `json:"total_alerts"`
	ActiveAlerts          int                    `json:"active_alerts"`
	ResolvedAlerts        int                    `json:"resolved_alerts"`
	AlertsBySeverity      map[string]int         `json:"alerts_by_severity"`
	AlertsByType          map[string]int         `json:"alerts_by_type"`
	AlertsByEntity        map[string]int         `json:"alerts_by_entity"`
	AverageResolutionTime time.Duration          `json:"average_resolution_time"`
	EscalationCount       int                    `json:"escalation_count"`
	SuppressionCount      int                    `json:"suppression_count"`
	Trends                []AlertTrend           `json:"trends"`
	GeneratedAt           time.Time              `json:"generated_at"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

// AlertTrend represents alert trends over time
type AlertTrend struct {
	Date           time.Time `json:"date"`
	TotalAlerts    int       `json:"total_alerts"`
	NewAlerts      int       `json:"new_alerts"`
	ResolvedAlerts int       `json:"resolved_alerts"`
	ActiveAlerts   int       `json:"active_alerts"`
}

// NewAlertSystem creates a new compliance alert system
func NewAlertSystem(logger *observability.Logger, statusSystem *ComplianceStatusSystem, checkEngine *CheckEngine) *AlertSystem {
	return &AlertSystem{
		logger:        logger,
		statusSystem:  statusSystem,
		checkEngine:   checkEngine,
		rules:         make(map[string]*AlertRule),
		escalations:   make(map[string]*EscalationPolicy),
		notifications: make(map[string]*NotificationChannel),
	}
}

// RegisterAlertRule registers a new alert rule
func (s *AlertSystem) RegisterAlertRule(ctx context.Context, rule *AlertRule) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Registering alert rule",
		"request_id", requestID,
		"rule_id", rule.ID,
		"rule_name", rule.Name,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	s.rules[rule.ID] = rule

	s.logger.Info("Alert rule registered successfully",
		"request_id", requestID,
		"rule_id", rule.ID,
	)

	return nil
}

// UpdateAlertRule updates an existing alert rule
func (s *AlertSystem) UpdateAlertRule(ctx context.Context, ruleID string, updates map[string]interface{}) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating alert rule",
		"request_id", requestID,
		"rule_id", ruleID,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	rule, exists := s.rules[ruleID]
	if !exists {
		return fmt.Errorf("alert rule %s not found", ruleID)
	}

	// Apply updates
	for field, value := range updates {
		switch field {
		case "name":
			if name, ok := value.(string); ok {
				rule.Name = name
			}
		case "description":
			if desc, ok := value.(string); ok {
				rule.Description = desc
			}
		case "enabled":
			if enabled, ok := value.(bool); ok {
				rule.Enabled = enabled
			}
		case "severity":
			if severity, ok := value.(string); ok {
				rule.Severity = severity
			}
		case "conditions":
			if conditions, ok := value.([]AlertCondition); ok {
				rule.Conditions = conditions
			}
		case "actions":
			if actions, ok := value.([]AlertAction); ok {
				rule.Actions = actions
			}
		}
	}

	rule.UpdatedAt = time.Now()

	s.logger.Info("Alert rule updated successfully",
		"request_id", requestID,
		"rule_id", ruleID,
	)

	return nil
}

// DeleteAlertRule deletes an alert rule
func (s *AlertSystem) DeleteAlertRule(ctx context.Context, ruleID string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Deleting alert rule",
		"request_id", requestID,
		"rule_id", ruleID,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rules[ruleID]; !exists {
		return fmt.Errorf("alert rule %s not found", ruleID)
	}

	delete(s.rules, ruleID)

	s.logger.Info("Alert rule deleted successfully",
		"request_id", requestID,
		"rule_id", ruleID,
	)

	return nil
}

// GetAlertRule gets an alert rule by ID
func (s *AlertSystem) GetAlertRule(ctx context.Context, ruleID string) (*AlertRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rule, exists := s.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("alert rule %s not found", ruleID)
	}

	return rule, nil
}

// ListAlertRules lists all alert rules
func (s *AlertSystem) ListAlertRules(ctx context.Context) ([]*AlertRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rules := make([]*AlertRule, 0, len(s.rules))
	for _, rule := range s.rules {
		rules = append(rules, rule)
	}

	return rules, nil
}

// EvaluateAlerts evaluates all alert rules for a business
func (s *AlertSystem) EvaluateAlerts(ctx context.Context, businessID string) ([]AlertEvaluation, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Evaluating alerts for business",
		"request_id", requestID,
		"business_id", businessID,
	)

	s.mu.RLock()
	rules := make([]*AlertRule, 0, len(s.rules))
	for _, rule := range s.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	s.mu.RUnlock()

	evaluations := make([]AlertEvaluation, 0)

	for _, rule := range rules {
		evaluation, err := s.evaluateRule(ctx, businessID, rule)
		if err != nil {
			s.logger.Error("Failed to evaluate rule",
				"request_id", requestID,
				"rule_id", rule.ID,
				"error", err.Error(),
			)
			// Create a failed evaluation instead of skipping
			failedEvaluation := AlertEvaluation{
				RuleID:      rule.ID,
				BusinessID:  businessID,
				EntityType:  rule.EntityType,
				Triggered:   false,
				EvaluatedAt: time.Now(),
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			}
			evaluations = append(evaluations, failedEvaluation)
			continue
		}

		evaluations = append(evaluations, evaluation)

		// Execute actions if rule was triggered
		if evaluation.Triggered {
			s.executeAlertActions(ctx, businessID, rule, evaluation)
		}
	}

	s.logger.Info("Alert evaluation completed",
		"request_id", requestID,
		"business_id", businessID,
		"rules_evaluated", len(rules),
		"rules_triggered", len(evaluations),
	)

	return evaluations, nil
}

// evaluateRule evaluates a single alert rule
func (s *AlertSystem) evaluateRule(ctx context.Context, businessID string, rule *AlertRule) (AlertEvaluation, error) {
	evaluation := AlertEvaluation{
		RuleID:      rule.ID,
		BusinessID:  businessID,
		EntityType:  rule.EntityType,
		EvaluatedAt: time.Now(),
	}

	// Evaluate each condition
	allConditionsMet := true
	for _, condition := range rule.Conditions {
		result, err := s.evaluateCondition(ctx, businessID, condition)
		if err != nil {
			return evaluation, fmt.Errorf("failed to evaluate condition: %w", err)
		}

		evaluation.Conditions = append(evaluation.Conditions, result)
		if !result.Triggered {
			allConditionsMet = false
		}
	}

	evaluation.Triggered = allConditionsMet

	return evaluation, nil
}

// evaluateCondition evaluates a single alert condition
func (s *AlertSystem) evaluateCondition(ctx context.Context, businessID string, condition AlertCondition) (ConditionResult, error) {
	result := ConditionResult{
		Condition:   condition,
		EvaluatedAt: time.Now(),
	}

	// Get current value based on condition type
	var currentValue interface{}
	var err error

	switch condition.Type {
	case "score_below":
		currentValue, err = s.getCurrentScore(ctx, businessID, condition.Field)
	case "score_decline":
		currentValue, err = s.getScoreDecline(ctx, businessID, condition.Field, condition.Window)
	case "status_change":
		currentValue, err = s.getStatusChange(ctx, businessID, condition.Field, condition.Window)
	case "deadline_missed":
		currentValue, err = s.getDeadlineStatus(ctx, businessID, condition.Field)
	case "risk_increase":
		currentValue, err = s.getRiskIncrease(ctx, businessID, condition.Field, condition.Window)
	default:
		return result, fmt.Errorf("unsupported condition type: %s", condition.Type)
	}

	if err != nil {
		return result, fmt.Errorf("failed to get current value: %w", err)
	}

	result.Value = currentValue

	// Evaluate condition based on operator
	result.Triggered = s.evaluateOperator(currentValue, condition.Operator, condition.Value)
	result.Message = fmt.Sprintf("Condition %s: current=%v, threshold=%v, triggered=%v",
		condition.Type, currentValue, condition.Value, result.Triggered)

	return result, nil
}

// evaluateOperator evaluates a condition using the specified operator
func (s *AlertSystem) evaluateOperator(current, operator, threshold interface{}) bool {
	switch operator {
	case "lt":
		return s.compareValues(current, threshold) < 0
	case "lte":
		return s.compareValues(current, threshold) <= 0
	case "eq":
		return s.compareValues(current, threshold) == 0
	case "gte":
		return s.compareValues(current, threshold) >= 0
	case "gt":
		return s.compareValues(current, threshold) > 0
	case "ne":
		return s.compareValues(current, threshold) != 0
	default:
		return false
	}
}

// compareValues compares two values for evaluation
func (s *AlertSystem) compareValues(a, b interface{}) int {
	// Convert to float64 for numeric comparison
	aFloat, aOk := s.toFloat64(a)
	bFloat, bOk := s.toFloat64(b)

	if aOk && bOk {
		if aFloat < bFloat {
			return -1
		} else if aFloat > bFloat {
			return 1
		}
		return 0
	}

	// String comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}

// toFloat64 converts a value to float64
func (s *AlertSystem) toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

// Helper methods for getting current values
func (s *AlertSystem) getCurrentScore(ctx context.Context, businessID, field string) (interface{}, error) {
	status, err := s.statusSystem.GetComplianceStatus(ctx, businessID)
	if err != nil {
		return nil, err
	}

	switch field {
	case "overall":
		return status.OverallScore, nil
	default:
		// Try to get framework score
		if frameworkStatus, exists := status.FrameworkStatuses[field]; exists {
			return frameworkStatus.Score, nil
		}
		return nil, fmt.Errorf("unknown score field: %s", field)
	}
}

func (s *AlertSystem) getScoreDecline(ctx context.Context, businessID, field string, window *time.Duration) (interface{}, error) {
	// This would compare current score with historical score
	// For now, return a placeholder
	return 0.0, nil
}

func (s *AlertSystem) getStatusChange(ctx context.Context, businessID, field string, window *time.Duration) (interface{}, error) {
	// This would check for status changes in the specified window
	// For now, return a placeholder
	return "stable", nil
}

func (s *AlertSystem) getDeadlineStatus(ctx context.Context, businessID, field string) (interface{}, error) {
	// This would check if deadlines are missed
	// For now, return a placeholder
	return false, nil
}

func (s *AlertSystem) getRiskIncrease(ctx context.Context, businessID, field string, window *time.Duration) (interface{}, error) {
	// This would check for risk level increases
	// For now, return a placeholder
	return "low", nil
}

// executeAlertActions executes the actions for a triggered alert
func (s *AlertSystem) executeAlertActions(ctx context.Context, businessID string, rule *AlertRule, evaluation AlertEvaluation) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Executing alert actions",
		"request_id", requestID,
		"business_id", businessID,
		"rule_id", rule.ID,
	)

	for _, action := range rule.Actions {
		result := s.executeAction(ctx, businessID, rule, action, evaluation)
		evaluation.Actions = append(evaluation.Actions, result)

		if !result.Success {
			s.logger.Error("Alert action failed",
				"request_id", requestID,
				"rule_id", rule.ID,
				"action_type", action.Type,
				"error", result.Message,
			)
		}
	}
}

// executeAction executes a single alert action
func (s *AlertSystem) executeAction(ctx context.Context, businessID string, rule *AlertRule, action AlertAction, evaluation AlertEvaluation) ActionResult {
	result := ActionResult{
		Action:     action,
		ExecutedAt: time.Now(),
	}

	switch action.Type {
	case "create_alert":
		result.Success = s.createAlertFromRule(ctx, businessID, rule, evaluation)
	case "send_notification":
		result.Success = s.sendNotification(ctx, businessID, rule, action, evaluation)
	case "escalate":
		result.Success = s.escalateAlert(ctx, businessID, rule, action, evaluation)
	case "webhook":
		result.Success = s.sendWebhook(ctx, businessID, rule, action, evaluation)
	default:
		result.Success = false
		result.Message = fmt.Sprintf("unknown action type: %s", action.Type)
	}

	return result
}

// createAlertFromRule creates an alert from a triggered rule
func (s *AlertSystem) createAlertFromRule(ctx context.Context, businessID string, rule *AlertRule, evaluation AlertEvaluation) bool {
	// Check suppression
	if rule.Suppression != nil && rule.Suppression.Enabled {
		if s.isAlertSuppressed(ctx, businessID, rule) {
			return true // Suppressed, but not an error
		}
	}

	// Create alert using the status system
	title := fmt.Sprintf("Alert: %s", rule.Name)
	description := fmt.Sprintf("Rule '%s' triggered for business %s", rule.Name, businessID)

	err := s.statusSystem.CreateStatusAlert(ctx, businessID, "rule_triggered", rule.Severity, rule.EntityType, "", title, description, nil, nil)
	return err == nil
}

// isAlertSuppressed checks if an alert should be suppressed
func (s *AlertSystem) isAlertSuppressed(ctx context.Context, businessID string, rule *AlertRule) bool {
	// This would check if similar alerts were created recently
	// For now, return false (no suppression)
	return false
}

// sendNotification sends a notification
func (s *AlertSystem) sendNotification(ctx context.Context, businessID string, rule *AlertRule, action AlertAction, evaluation AlertEvaluation) bool {
	// This would send notifications via configured channels
	// For now, return true (success)
	return true
}

// escalateAlert escalates an alert
func (s *AlertSystem) escalateAlert(ctx context.Context, businessID string, rule *AlertRule, action AlertAction, evaluation AlertEvaluation) bool {
	// This would escalate the alert according to escalation policies
	// For now, return true (success)
	return true
}

// sendWebhook sends a webhook notification
func (s *AlertSystem) sendWebhook(ctx context.Context, businessID string, rule *AlertRule, action AlertAction, evaluation AlertEvaluation) bool {
	// This would send webhook notifications
	// For now, return true (success)
	return true
}

// GetAlertAnalytics gets analytics about alerts for a business
func (s *AlertSystem) GetAlertAnalytics(ctx context.Context, businessID string, period string) (*AlertAnalytics, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting alert analytics",
		"request_id", requestID,
		"business_id", businessID,
		"period", period,
	)

	// Get alerts from status system
	alerts, err := s.statusSystem.GetStatusAlerts(ctx, businessID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	analytics := &AlertAnalytics{
		BusinessID:       businessID,
		Period:           period,
		GeneratedAt:      time.Now(),
		AlertsBySeverity: make(map[string]int),
		AlertsByType:     make(map[string]int),
		AlertsByEntity:   make(map[string]int),
	}

	// Calculate analytics
	for _, alert := range alerts {
		analytics.TotalAlerts++

		if alert.Status == "active" {
			analytics.ActiveAlerts++
		} else if alert.Status == "resolved" {
			analytics.ResolvedAlerts++
		}

		analytics.AlertsBySeverity[alert.Severity]++
		analytics.AlertsByType[alert.AlertType]++
		analytics.AlertsByEntity[alert.EntityType]++
	}

	// Calculate average resolution time
	if analytics.ResolvedAlerts > 0 {
		totalResolutionTime := time.Duration(0)
		resolvedCount := 0

		for _, alert := range alerts {
			if alert.Status == "resolved" && alert.ResolvedAt != nil {
				resolutionTime := alert.ResolvedAt.Sub(alert.TriggeredAt)
				totalResolutionTime += resolutionTime
				resolvedCount++
			}
		}

		if resolvedCount > 0 {
			analytics.AverageResolutionTime = totalResolutionTime / time.Duration(resolvedCount)
		}
	}

	s.logger.Info("Alert analytics generated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"total_alerts", analytics.TotalAlerts,
		"active_alerts", analytics.ActiveAlerts,
	)

	return analytics, nil
}

// RegisterEscalationPolicy registers a new escalation policy
func (s *AlertSystem) RegisterEscalationPolicy(ctx context.Context, policy *EscalationPolicy) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Registering escalation policy",
		"request_id", requestID,
		"policy_id", policy.ID,
		"policy_name", policy.Name,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()
	s.escalations[policy.ID] = policy

	s.logger.Info("Escalation policy registered successfully",
		"request_id", requestID,
		"policy_id", policy.ID,
	)

	return nil
}

// RegisterNotificationChannel registers a new notification channel
func (s *AlertSystem) RegisterNotificationChannel(ctx context.Context, channel *NotificationChannel) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Registering notification channel",
		"request_id", requestID,
		"channel_id", channel.ID,
		"channel_name", channel.Name,
		"channel_type", channel.Type,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()
	s.notifications[channel.ID] = channel

	s.logger.Info("Notification channel registered successfully",
		"request_id", requestID,
		"channel_id", channel.ID,
	)

	return nil
}
