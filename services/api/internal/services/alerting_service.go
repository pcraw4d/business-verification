package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AlertingService provides alerting functionality
type AlertingService struct {
	logger    *zap.Logger
	alerts    map[string]*Alert
	mu        sync.RWMutex
	notifiers []AlertNotifier
}

// Alert represents an alert
type Alert struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    AlertSeverity          `json:"severity"`
	Status      AlertStatus            `json:"status"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Labels      map[string]string      `json:"labels"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive     AlertStatus = "active"
	AlertStatusResolved   AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
)

// AlertNotifier interface for alert notification systems
type AlertNotifier interface {
	Notify(alert *Alert) error
	Name() string
}

// AlertRule represents a rule for generating alerts
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Severity    AlertSeverity          `json:"severity"`
	Condition   string                 `json:"condition"`
	Threshold   float64                `json:"threshold"`
	Duration    time.Duration          `json:"duration"`
	Labels      map[string]string      `json:"labels"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewAlertingService creates a new alerting service
func NewAlertingService(logger *zap.Logger) *AlertingService {
	return &AlertingService{
		logger:    logger,
		alerts:    make(map[string]*Alert),
		notifiers: make([]AlertNotifier, 0),
	}
}

// AddNotifier adds an alert notifier
func (a *AlertingService) AddNotifier(notifier AlertNotifier) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.notifiers = append(a.notifiers, notifier)
}

// CreateAlert creates a new alert
func (a *AlertingService) CreateAlert(alert *Alert) error {
	if alert.ID == "" {
		alert.ID = fmt.Sprintf("alert-%d", time.Now().UnixNano())
	}

	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}

	if alert.Status == "" {
		alert.Status = AlertStatusActive
	}

	a.mu.Lock()
	a.alerts[alert.ID] = alert
	a.mu.Unlock()

	a.logger.Info("Alert created",
		zap.String("alert_id", alert.ID),
		zap.String("title", alert.Title),
		zap.String("severity", string(alert.Severity)))

	// Send notifications
	go a.sendNotifications(alert)

	return nil
}

// ResolveAlert resolves an alert
func (a *AlertingService) ResolveAlert(alertID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	alert, exists := a.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	if alert.Status == AlertStatusResolved {
		return fmt.Errorf("alert already resolved: %s", alertID)
	}

	now := time.Now()
	alert.Status = AlertStatusResolved
	alert.ResolvedAt = &now

	a.logger.Info("Alert resolved",
		zap.String("alert_id", alertID),
		zap.String("title", alert.Title),
		zap.Duration("duration", now.Sub(alert.Timestamp)))

	return nil
}

// GetActiveAlerts returns all active alerts
func (a *AlertingService) GetActiveAlerts() []*Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var activeAlerts []*Alert
	for _, alert := range a.alerts {
		if alert.Status == AlertStatusActive {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// GetAlertsBySeverity returns alerts filtered by severity
func (a *AlertingService) GetAlertsBySeverity(severity AlertSeverity) []*Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var filteredAlerts []*Alert
	for _, alert := range a.alerts {
		if alert.Severity == severity {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	return filteredAlerts
}

// GetAlertHistory returns alert history
func (a *AlertingService) GetAlertHistory(limit int) []*Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var alerts []*Alert
	for _, alert := range a.alerts {
		alerts = append(alerts, alert)
	}

	// Sort by timestamp (newest first)
	// In a real implementation, you might want to use a more sophisticated sorting
	if len(alerts) > limit {
		alerts = alerts[:limit]
	}

	return alerts
}

// CheckAlertRules checks alert rules and creates alerts if conditions are met
func (a *AlertingService) CheckAlertRules(ctx context.Context, metrics map[string]float64) error {
	rules := a.getDefaultAlertRules()

	for _, rule := range rules {
		if a.evaluateRule(rule, metrics) {
			alert := &Alert{
				Title:       rule.Name,
				Description: rule.Description,
				Severity:    rule.Severity,
				Source:      "alerting_service",
				Labels:      rule.Labels,
				Metadata:    rule.Metadata,
			}

			if err := a.CreateAlert(alert); err != nil {
				a.logger.Error("Failed to create alert from rule",
					zap.String("rule_id", rule.ID),
					zap.Error(err))
			}
		}
	}

	return nil
}

// sendNotifications sends notifications for an alert
func (a *AlertingService) sendNotifications(alert *Alert) {
	for _, notifier := range a.notifiers {
		if err := notifier.Notify(alert); err != nil {
			a.logger.Error("Failed to send notification",
				zap.String("notifier", notifier.Name()),
				zap.String("alert_id", alert.ID),
				zap.Error(err))
		}
	}
}

// evaluateRule evaluates an alert rule against current metrics
func (a *AlertingService) evaluateRule(rule AlertRule, metrics map[string]float64) bool {
	// Simple rule evaluation - in real implementation, use a proper expression evaluator
	switch rule.Condition {
	case "greater_than":
		if value, exists := metrics[rule.ID]; exists {
			return value > rule.Threshold
		}
	case "less_than":
		if value, exists := metrics[rule.ID]; exists {
			return value < rule.Threshold
		}
	case "equals":
		if value, exists := metrics[rule.ID]; exists {
			return value == rule.Threshold
		}
	}

	return false
}

// getDefaultAlertRules returns default alert rules
func (a *AlertingService) getDefaultAlertRules() []AlertRule {
	return []AlertRule{
		{
			ID:          "high_error_rate",
			Name:        "High Error Rate",
			Description: "Error rate has exceeded the threshold",
			Severity:    AlertSeverityCritical,
			Condition:   "greater_than",
			Threshold:   5.0,
			Duration:    2 * time.Minute,
			Labels: map[string]string{
				"service": "api",
				"type":    "error_rate",
			},
		},
		{
			ID:          "high_response_time",
			Name:        "High Response Time",
			Description: "Response time has exceeded the threshold",
			Severity:    AlertSeverityWarning,
			Condition:   "greater_than",
			Threshold:   2000.0, // 2 seconds
			Duration:    5 * time.Minute,
			Labels: map[string]string{
				"service": "api",
				"type":    "response_time",
			},
		},
		{
			ID:          "high_memory_usage",
			Name:        "High Memory Usage",
			Description: "Memory usage has exceeded the threshold",
			Severity:    AlertSeverityWarning,
			Condition:   "greater_than",
			Threshold:   80.0, // 80%
			Duration:    5 * time.Minute,
			Labels: map[string]string{
				"service": "system",
				"type":    "memory",
			},
		},
		{
			ID:          "high_cpu_usage",
			Name:        "High CPU Usage",
			Description: "CPU usage has exceeded the threshold",
			Severity:    AlertSeverityWarning,
			Condition:   "greater_than",
			Threshold:   80.0, // 80%
			Duration:    5 * time.Minute,
			Labels: map[string]string{
				"service": "system",
				"type":    "cpu",
			},
		},
	}
}

// MockAlertNotifier is a mock implementation of AlertNotifier
type MockAlertNotifier struct {
	logger *zap.Logger
}

// NewMockAlertNotifier creates a new mock alert notifier
func NewMockAlertNotifier(logger *zap.Logger) *MockAlertNotifier {
	return &MockAlertNotifier{logger: logger}
}

// Notify sends a mock notification
func (m *MockAlertNotifier) Notify(alert *Alert) error {
	m.logger.Info("Mock alert notification sent",
		zap.String("alert_id", alert.ID),
		zap.String("title", alert.Title),
		zap.String("severity", string(alert.Severity)))
	return nil
}

// Name returns the notifier name
func (m *MockAlertNotifier) Name() string {
	return "mock"
}
