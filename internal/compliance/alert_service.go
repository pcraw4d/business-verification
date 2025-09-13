package compliance

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ComplianceAlertService provides comprehensive compliance alerting and notification capabilities
type ComplianceAlertService struct {
	logger           *observability.Logger
	frameworkService *ComplianceFrameworkService
	trackingService  *ComplianceTrackingService
	alerts           map[string]*ComplianceAlert
	alertRules       map[string]*AlertRule
	notifications    map[string]*Notification
	alertHistory     map[string]*AlertHistory
}

// ComplianceAlert represents a compliance alert
type ComplianceAlert struct {
	ID              string                 `json:"id"`
	BusinessID      string                 `json:"business_id"`
	FrameworkID     string                 `json:"framework_id"`
	AlertType       string                 `json:"alert_type"` // "deadline", "risk_threshold", "compliance_change", "gap_detected", "milestone_overdue"
	Severity        string                 `json:"severity"`   // "low", "medium", "high", "critical"
	Status          string                 `json:"status"`     // "active", "acknowledged", "resolved", "dismissed"
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Message         string                 `json:"message"`
	TriggeredBy     string                 `json:"triggered_by"` // "system", "user", "rule"
	TriggeredAt     time.Time              `json:"triggered_at"`
	AcknowledgedBy  string                 `json:"acknowledged_by,omitempty"`
	AcknowledgedAt  *time.Time             `json:"acknowledged_at,omitempty"`
	ResolvedBy      string                 `json:"resolved_by,omitempty"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty"`
	DueDate         *time.Time             `json:"due_date,omitempty"`
	RequirementID   string                 `json:"requirement_id,omitempty"`
	MilestoneID     string                 `json:"milestone_id,omitempty"`
	RiskScore       float64                `json:"risk_score,omitempty"`
	ComplianceScore float64                `json:"compliance_score,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// AlertRule defines rules for automatic alert generation
type AlertRule struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	AlertType   string           `json:"alert_type"`
	Severity    string           `json:"severity"`
	Conditions  []AlertCondition `json:"conditions"`
	Actions     []AlertAction    `json:"actions"`
	Enabled     bool             `json:"enabled"`
	BusinessID  string           `json:"business_id,omitempty"`  // Empty for global rules
	FrameworkID string           `json:"framework_id,omitempty"` // Empty for all frameworks
	CreatedBy   string           `json:"created_by"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// AlertCondition defines conditions for alert triggering
type AlertCondition struct {
	Field      string      `json:"field"`    // "compliance_score", "risk_score", "deadline", "requirement_status"
	Operator   string      `json:"operator"` // "eq", "ne", "gt", "gte", "lt", "lte", "contains", "in"
	Value      interface{} `json:"value"`
	Threshold  float64     `json:"threshold,omitempty"`
	TimeWindow string      `json:"time_window,omitempty"` // "1h", "24h", "7d", "30d"
}

// AlertAction defines actions to take when alert is triggered
type AlertAction struct {
	Type       string                 `json:"type"` // "email", "webhook", "slack", "create_ticket", "escalate"
	Config     map[string]interface{} `json:"config"`
	Delay      string                 `json:"delay,omitempty"` // "0s", "5m", "1h"
	RetryCount int                    `json:"retry_count,omitempty"`
	Enabled    bool                   `json:"enabled"`
}

// Notification represents a sent notification
type Notification struct {
	ID            string                 `json:"id"`
	AlertID       string                 `json:"alert_id"`
	Type          string                 `json:"type"` // "email", "webhook", "slack", "sms"
	Recipient     string                 `json:"recipient"`
	Subject       string                 `json:"subject,omitempty"`
	Message       string                 `json:"message"`
	Status        string                 `json:"status"` // "pending", "sent", "delivered", "failed", "bounced"
	SentAt        *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt   *time.Time             `json:"delivered_at,omitempty"`
	FailedAt      *time.Time             `json:"failed_at,omitempty"`
	FailureReason string                 `json:"failure_reason,omitempty"`
	RetryCount    int                    `json:"retry_count"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// AlertHistory tracks alert lifecycle events
type AlertHistory struct {
	ID               string                 `json:"id"`
	AlertID          string                 `json:"alert_id"`
	EventType        string                 `json:"event_type"` // "created", "acknowledged", "resolved", "dismissed", "escalated"
	EventDescription string                 `json:"event_description"`
	PerformedBy      string                 `json:"performed_by"`
	PerformedAt      time.Time              `json:"performed_at"`
	PreviousStatus   string                 `json:"previous_status,omitempty"`
	NewStatus        string                 `json:"new_status,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// AlertQuery represents query parameters for alert operations
type AlertQuery struct {
	BusinessID      string     `json:"business_id,omitempty"`
	FrameworkID     string     `json:"framework_id,omitempty"`
	AlertType       string     `json:"alert_type,omitempty"`
	Severity        string     `json:"severity,omitempty"`
	Status          string     `json:"status,omitempty"`
	TriggeredBy     string     `json:"triggered_by,omitempty"`
	IncludeResolved bool       `json:"include_resolved,omitempty"`
	StartDate       *time.Time `json:"start_date,omitempty"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	Limit           int        `json:"limit,omitempty"`
	Offset          int        `json:"offset,omitempty"`
}

// NotificationQuery represents query parameters for notification operations
type NotificationQuery struct {
	AlertID   string     `json:"alert_id,omitempty"`
	Type      string     `json:"type,omitempty"`
	Recipient string     `json:"recipient,omitempty"`
	Status    string     `json:"status,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// NewComplianceAlertService creates a new compliance alert service
func NewComplianceAlertService(logger *observability.Logger, frameworkService *ComplianceFrameworkService, trackingService *ComplianceTrackingService) *ComplianceAlertService {
	service := &ComplianceAlertService{
		logger:           logger,
		frameworkService: frameworkService,
		trackingService:  trackingService,
		alerts:           make(map[string]*ComplianceAlert),
		alertRules:       make(map[string]*AlertRule),
		notifications:    make(map[string]*Notification),
		alertHistory:     make(map[string]*AlertHistory),
	}

	// Load default alert rules
	service.loadDefaultAlertRules()

	return service
}

// CreateAlert creates a new compliance alert
func (cas *ComplianceAlertService) CreateAlert(ctx context.Context, alert *ComplianceAlert) error {
	cas.logger.Info("Creating compliance alert", map[string]interface{}{
		"business_id":  alert.BusinessID,
		"framework_id": alert.FrameworkID,
		"alert_type":   alert.AlertType,
		"severity":     alert.Severity,
	})

	// Set timestamps
	now := time.Now()
	alert.CreatedAt = now
	alert.UpdatedAt = now
	alert.TriggeredAt = now

	// Set default status if not provided
	if alert.Status == "" {
		alert.Status = "active"
	}

	// Generate alert ID if not provided
	if alert.ID == "" {
		alert.ID = cas.generateAlertID()
	}

	// Store alert
	cas.alerts[alert.ID] = alert

	// Create alert history entry
	cas.createAlertHistory(alert.ID, "created", "Alert created", alert.TriggeredBy, "active")

	// Process alert actions
	cas.processAlertActions(ctx, alert)

	cas.logger.Info("Created compliance alert", map[string]interface{}{
		"alert_id":     alert.ID,
		"business_id":  alert.BusinessID,
		"framework_id": alert.FrameworkID,
		"alert_type":   alert.AlertType,
		"severity":     alert.Severity,
		"status":       alert.Status,
	})

	return nil
}

// GetAlert retrieves a compliance alert by ID
func (cas *ComplianceAlertService) GetAlert(ctx context.Context, alertID string) (*ComplianceAlert, error) {
	cas.logger.Info("Retrieving compliance alert", map[string]interface{}{
		"alert_id": alertID,
	})

	alert, exists := cas.alerts[alertID]
	if !exists {
		return nil, fmt.Errorf("alert not found: %s", alertID)
	}

	cas.logger.Info("Retrieved compliance alert", map[string]interface{}{
		"alert_id":     alertID,
		"business_id":  alert.BusinessID,
		"framework_id": alert.FrameworkID,
		"alert_type":   alert.AlertType,
		"status":       alert.Status,
	})

	return alert, nil
}

// ListAlerts lists compliance alerts with optional filtering
func (cas *ComplianceAlertService) ListAlerts(ctx context.Context, query *AlertQuery) ([]*ComplianceAlert, error) {
	cas.logger.Info("Listing compliance alerts", map[string]interface{}{
		"query": query,
	})

	var alerts []*ComplianceAlert

	for _, alert := range cas.alerts {
		// Apply filters
		if query.BusinessID != "" && alert.BusinessID != query.BusinessID {
			continue
		}
		if query.FrameworkID != "" && alert.FrameworkID != query.FrameworkID {
			continue
		}
		if query.AlertType != "" && alert.AlertType != query.AlertType {
			continue
		}
		if query.Severity != "" && alert.Severity != query.Severity {
			continue
		}
		if query.Status != "" && alert.Status != query.Status {
			continue
		}
		if query.TriggeredBy != "" && alert.TriggeredBy != query.TriggeredBy {
			continue
		}
		if !query.IncludeResolved && alert.Status == "resolved" {
			continue
		}
		if query.StartDate != nil && alert.TriggeredAt.Before(*query.StartDate) {
			continue
		}
		if query.EndDate != nil && alert.TriggeredAt.After(*query.EndDate) {
			continue
		}

		alerts = append(alerts, alert)
	}

	// Sort by triggered date (newest first)
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].TriggeredAt.After(alerts[j].TriggeredAt)
	})

	// Apply pagination
	if query.Limit > 0 {
		start := query.Offset
		end := query.Offset + query.Limit
		if start >= len(alerts) {
			alerts = []*ComplianceAlert{}
		} else if end > len(alerts) {
			alerts = alerts[start:]
		} else {
			alerts = alerts[start:end]
		}
	}

	cas.logger.Info("Listed compliance alerts", map[string]interface{}{
		"count": len(alerts),
		"query": query,
	})

	return alerts, nil
}

// UpdateAlertStatus updates the status of a compliance alert
func (cas *ComplianceAlertService) UpdateAlertStatus(ctx context.Context, alertID, newStatus, updatedBy string) error {
	cas.logger.Info("Updating alert status", map[string]interface{}{
		"alert_id":   alertID,
		"new_status": newStatus,
		"updated_by": updatedBy,
	})

	alert, exists := cas.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	previousStatus := alert.Status
	alert.Status = newStatus
	alert.UpdatedAt = time.Now()

	// Update specific fields based on status
	now := time.Now()
	switch newStatus {
	case "acknowledged":
		alert.AcknowledgedBy = updatedBy
		alert.AcknowledgedAt = &now
	case "resolved":
		alert.ResolvedBy = updatedBy
		alert.ResolvedAt = &now
	}

	// Store updated alert
	cas.alerts[alertID] = alert

	// Create alert history entry
	cas.createAlertHistory(alertID, "status_changed", fmt.Sprintf("Status changed from %s to %s", previousStatus, newStatus), updatedBy, newStatus)

	cas.logger.Info("Updated alert status", map[string]interface{}{
		"alert_id":        alertID,
		"previous_status": previousStatus,
		"new_status":      newStatus,
		"updated_by":      updatedBy,
	})

	return nil
}

// CreateAlertRule creates a new alert rule
func (cas *ComplianceAlertService) CreateAlertRule(ctx context.Context, rule *AlertRule) error {
	cas.logger.Info("Creating alert rule", map[string]interface{}{
		"name":       rule.Name,
		"alert_type": rule.AlertType,
		"severity":   rule.Severity,
		"created_by": rule.CreatedBy,
	})

	// Set timestamps
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	// Generate rule ID if not provided
	if rule.ID == "" {
		rule.ID = cas.generateRuleID()
	}

	// Store rule
	cas.alertRules[rule.ID] = rule

	cas.logger.Info("Created alert rule", map[string]interface{}{
		"rule_id":    rule.ID,
		"name":       rule.Name,
		"alert_type": rule.AlertType,
		"severity":   rule.Severity,
		"enabled":    rule.Enabled,
	})

	return nil
}

// EvaluateAlertRules evaluates all applicable alert rules for a business/framework
func (cas *ComplianceAlertService) EvaluateAlertRules(ctx context.Context, businessID, frameworkID string) error {
	cas.logger.Info("Evaluating alert rules", map[string]interface{}{
		"business_id":  businessID,
		"framework_id": frameworkID,
	})

	// Get compliance tracking data
	tracking, err := cas.trackingService.GetComplianceTracking(ctx, businessID, frameworkID)
	if err != nil {
		return fmt.Errorf("failed to get compliance tracking: %w", err)
	}

	// Evaluate each rule
	for _, rule := range cas.alertRules {
		if !rule.Enabled {
			continue
		}

		// Check if rule applies to this business/framework
		if rule.BusinessID != "" && rule.BusinessID != businessID {
			continue
		}
		if rule.FrameworkID != "" && rule.FrameworkID != frameworkID {
			continue
		}

		// Evaluate rule conditions
		if cas.evaluateRuleConditions(rule, tracking) {
			// Create alert
			alert := &ComplianceAlert{
				BusinessID:      businessID,
				FrameworkID:     frameworkID,
				AlertType:       rule.AlertType,
				Severity:        rule.Severity,
				Title:           rule.Name,
				Description:     rule.Description,
				Message:         cas.generateAlertMessage(rule, tracking),
				TriggeredBy:     "rule",
				ComplianceScore: tracking.OverallProgress,
				RiskScore:       1.0 - tracking.OverallProgress,
			}

			err := cas.CreateAlert(ctx, alert)
			if err != nil {
				cas.logger.Error("Failed to create alert from rule", map[string]interface{}{
					"rule_id":      rule.ID,
					"business_id":  businessID,
					"framework_id": frameworkID,
					"error":        err.Error(),
				})
			}
		}
	}

	cas.logger.Info("Evaluated alert rules", map[string]interface{}{
		"business_id":     businessID,
		"framework_id":    frameworkID,
		"rules_evaluated": len(cas.alertRules),
	})

	return nil
}

// GetNotifications retrieves notifications for an alert
func (cas *ComplianceAlertService) GetNotifications(ctx context.Context, query *NotificationQuery) ([]*Notification, error) {
	cas.logger.Info("Retrieving notifications", map[string]interface{}{
		"query": query,
	})

	var notifications []*Notification

	for _, notification := range cas.notifications {
		// Apply filters
		if query.AlertID != "" && notification.AlertID != query.AlertID {
			continue
		}
		if query.Type != "" && notification.Type != query.Type {
			continue
		}
		if query.Recipient != "" && notification.Recipient != query.Recipient {
			continue
		}
		if query.Status != "" && notification.Status != query.Status {
			continue
		}
		if query.StartDate != nil && notification.CreatedAt.Before(*query.StartDate) {
			continue
		}
		if query.EndDate != nil && notification.CreatedAt.After(*query.EndDate) {
			continue
		}

		notifications = append(notifications, notification)
	}

	// Sort by creation date (newest first)
	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].CreatedAt.After(notifications[j].CreatedAt)
	})

	// Apply pagination
	if query.Limit > 0 {
		start := query.Offset
		end := query.Offset + query.Limit
		if start >= len(notifications) {
			notifications = []*Notification{}
		} else if end > len(notifications) {
			notifications = notifications[start:]
		} else {
			notifications = notifications[start:end]
		}
	}

	cas.logger.Info("Retrieved notifications", map[string]interface{}{
		"count": len(notifications),
		"query": query,
	})

	return notifications, nil
}

// Helper methods

// loadDefaultAlertRules loads default alert rules
func (cas *ComplianceAlertService) loadDefaultAlertRules() {
	rules := []*AlertRule{
		{
			ID:          "low_compliance_score",
			Name:        "Low Compliance Score Alert",
			Description: "Alert when compliance score drops below 50%",
			AlertType:   "compliance_change",
			Severity:    "high",
			Conditions: []AlertCondition{
				{
					Field:     "compliance_score",
					Operator:  "lt",
					Threshold: 0.5,
				},
			},
			Actions: []AlertAction{
				{
					Type:    "email",
					Config:  map[string]interface{}{"template": "low_compliance"},
					Enabled: true,
				},
			},
			Enabled:   true,
			CreatedBy: "system",
		},
		{
			ID:          "critical_risk_level",
			Name:        "Critical Risk Level Alert",
			Description: "Alert when risk level becomes critical",
			AlertType:   "risk_threshold",
			Severity:    "critical",
			Conditions: []AlertCondition{
				{
					Field:    "risk_level",
					Operator: "eq",
					Value:    "critical",
				},
			},
			Actions: []AlertAction{
				{
					Type:    "email",
					Config:  map[string]interface{}{"template": "critical_risk"},
					Enabled: true,
				},
				{
					Type:    "webhook",
					Config:  map[string]interface{}{"url": "https://alerts.company.com/webhook"},
					Enabled: true,
				},
			},
			Enabled:   true,
			CreatedBy: "system",
		},
		{
			ID:          "deadline_approaching",
			Name:        "Deadline Approaching Alert",
			Description: "Alert when compliance deadline is approaching",
			AlertType:   "deadline",
			Severity:    "medium",
			Conditions: []AlertCondition{
				{
					Field:      "deadline",
					Operator:   "lt",
					TimeWindow: "7d",
				},
			},
			Actions: []AlertAction{
				{
					Type:    "email",
					Config:  map[string]interface{}{"template": "deadline_approaching"},
					Enabled: true,
				},
			},
			Enabled:   true,
			CreatedBy: "system",
		},
	}

	for _, rule := range rules {
		rule.CreatedAt = time.Now()
		rule.UpdatedAt = time.Now()
		cas.alertRules[rule.ID] = rule
	}
}

// generateAlertID generates a unique alert ID
func (cas *ComplianceAlertService) generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// generateRuleID generates a unique rule ID
func (cas *ComplianceAlertService) generateRuleID() string {
	return fmt.Sprintf("rule_%d", time.Now().UnixNano())
}

// createAlertHistory creates an alert history entry
func (cas *ComplianceAlertService) createAlertHistory(alertID, eventType, description, performedBy, newStatus string) {
	history := &AlertHistory{
		ID:               fmt.Sprintf("history_%d", time.Now().UnixNano()),
		AlertID:          alertID,
		EventType:        eventType,
		EventDescription: description,
		PerformedBy:      performedBy,
		PerformedAt:      time.Now(),
		NewStatus:        newStatus,
	}
	cas.alertHistory[history.ID] = history
}

// processAlertActions processes alert actions
func (cas *ComplianceAlertService) processAlertActions(ctx context.Context, alert *ComplianceAlert) {
	// Find applicable rules
	for _, rule := range cas.alertRules {
		if rule.AlertType == alert.AlertType && rule.Enabled {
			for _, action := range rule.Actions {
				if action.Enabled {
					cas.executeAlertAction(ctx, alert, action)
				}
			}
		}
	}
}

// executeAlertAction executes a specific alert action
func (cas *ComplianceAlertService) executeAlertAction(ctx context.Context, alert *ComplianceAlert, action AlertAction) {
	// Create notification
	notification := &Notification{
		ID:        fmt.Sprintf("notif_%d", time.Now().UnixNano()),
		AlertID:   alert.ID,
		Type:      action.Type,
		Recipient: cas.getRecipientFromConfig(action.Config),
		Subject:   cas.generateNotificationSubject(alert),
		Message:   cas.generateNotificationMessage(alert),
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store notification
	cas.notifications[notification.ID] = notification

	// Execute action (mock implementation)
	cas.logger.Info("Executing alert action", map[string]interface{}{
		"alert_id":        alert.ID,
		"action_type":     action.Type,
		"notification_id": notification.ID,
	})

	// Simulate action execution
	now := time.Now()
	notification.Status = "sent"
	notification.SentAt = &now
	notification.UpdatedAt = now
	cas.notifications[notification.ID] = notification
}

// evaluateRuleConditions evaluates rule conditions against tracking data
func (cas *ComplianceAlertService) evaluateRuleConditions(rule *AlertRule, tracking *ComplianceTracking) bool {
	for _, condition := range rule.Conditions {
		if !cas.evaluateCondition(condition, tracking) {
			return false
		}
	}
	return true
}

// evaluateCondition evaluates a single condition
func (cas *ComplianceAlertService) evaluateCondition(condition AlertCondition, tracking *ComplianceTracking) bool {
	switch condition.Field {
	case "compliance_score":
		return cas.compareValues(tracking.OverallProgress, condition.Operator, condition.Threshold)
	case "risk_level":
		return cas.compareValues(tracking.RiskLevel, condition.Operator, condition.Value)
	case "compliance_level":
		return cas.compareValues(tracking.ComplianceLevel, condition.Operator, condition.Value)
	case "trend":
		return cas.compareValues(tracking.Trend, condition.Operator, condition.Value)
	default:
		return false
	}
}

// compareValues compares two values using the specified operator
func (cas *ComplianceAlertService) compareValues(actual interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "eq":
		return actual == expected
	case "ne":
		return actual != expected
	case "gt":
		if actualFloat, ok := actual.(float64); ok {
			if expectedFloat, ok := expected.(float64); ok {
				return actualFloat > expectedFloat
			}
		}
		return false
	case "gte":
		if actualFloat, ok := actual.(float64); ok {
			if expectedFloat, ok := expected.(float64); ok {
				return actualFloat >= expectedFloat
			}
		}
		return false
	case "lt":
		if actualFloat, ok := actual.(float64); ok {
			if expectedFloat, ok := expected.(float64); ok {
				return actualFloat < expectedFloat
			}
		}
		return false
	case "lte":
		if actualFloat, ok := actual.(float64); ok {
			if expectedFloat, ok := expected.(float64); ok {
				return actualFloat <= expectedFloat
			}
		}
		return false
	default:
		return false
	}
}

// generateAlertMessage generates an alert message
func (cas *ComplianceAlertService) generateAlertMessage(rule *AlertRule, tracking *ComplianceTracking) string {
	return fmt.Sprintf("Alert: %s. Current compliance score: %.1f%%, Risk level: %s",
		rule.Description, tracking.OverallProgress*100, tracking.RiskLevel)
}

// generateNotificationSubject generates a notification subject
func (cas *ComplianceAlertService) generateNotificationSubject(alert *ComplianceAlert) string {
	return fmt.Sprintf("[%s] %s - %s", alert.Severity, alert.AlertType, alert.Title)
}

// generateNotificationMessage generates a notification message
func (cas *ComplianceAlertService) generateNotificationMessage(alert *ComplianceAlert) string {
	return fmt.Sprintf("Compliance Alert: %s\n\n%s\n\nSeverity: %s\nBusiness ID: %s\nFramework ID: %s",
		alert.Title, alert.Description, alert.Severity, alert.BusinessID, alert.FrameworkID)
}

// getRecipientFromConfig extracts recipient from action config
func (cas *ComplianceAlertService) getRecipientFromConfig(config map[string]interface{}) string {
	if recipient, ok := config["recipient"].(string); ok {
		return recipient
	}
	return "admin@company.com" // Default recipient
}
