package monitoring

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo      AlertSeverity = "info"
	AlertSeverityWarning   AlertSeverity = "warning"
	AlertSeverityCritical  AlertSeverity = "critical"
	AlertSeverityEmergency AlertSeverity = "emergency"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive     AlertStatus = "active"
	AlertStatusResolved   AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
)

// Alert represents a monitoring alert
type Alert struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    AlertSeverity          `json:"severity"`
	Status      AlertStatus            `json:"status"`
	Source      string                 `json:"source"`
	Metric      string                 `json:"metric"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	TenantID    string                 `json:"tenant_id"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AlertRule represents an alerting rule
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metric      string                 `json:"metric"`
	Condition   string                 `json:"condition"`
	Threshold   float64                `json:"threshold"`
	Severity    AlertSeverity          `json:"severity"`
	Duration    time.Duration          `json:"duration"`
	Enabled     bool                   `json:"enabled"`
	TenantID    string                 `json:"tenant_id"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AlertManager manages alerts and alerting rules
type AlertManager struct {
	rules    map[string]*AlertRule
	alerts   map[string]*Alert
	channels []AlertChannel
	logger   *zap.Logger
}

// AlertChannel represents a notification channel
type AlertChannel interface {
	SendAlert(ctx context.Context, alert *Alert) error
	GetName() string
	IsEnabled() bool
}

// EmailAlertChannel sends alerts via email
type EmailAlertChannel struct {
	enabled bool
	config  EmailConfig
	logger  *zap.Logger
}

// EmailConfig represents email configuration
type EmailConfig struct {
	SMTPHost    string   `json:"smtp_host"`
	SMTPPort    int      `json:"smtp_port"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	FromAddress string   `json:"from_address"`
	ToAddresses []string `json:"to_addresses"`
	UseTLS      bool     `json:"use_tls"`
}

// SlackAlertChannel sends alerts via Slack
type SlackAlertChannel struct {
	enabled bool
	config  SlackConfig
	logger  *zap.Logger
}

// SlackConfig represents Slack configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconEmoji  string `json:"icon_emoji"`
}

// WebhookAlertChannel sends alerts via webhook
type WebhookAlertChannel struct {
	enabled bool
	config  WebhookConfig
	logger  *zap.Logger
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Timeout time.Duration     `json:"timeout"`
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logger *zap.Logger) *AlertManager {
	return &AlertManager{
		rules:    make(map[string]*AlertRule),
		alerts:   make(map[string]*Alert),
		channels: make([]AlertChannel, 0),
		logger:   logger,
	}
}

// AddAlertChannel adds an alert channel
func (am *AlertManager) AddAlertChannel(channel AlertChannel) {
	am.channels = append(am.channels, channel)
	am.logger.Info("Added alert channel", zap.String("channel", channel.GetName()))
}

// AddAlertRule adds an alert rule
func (am *AlertManager) AddAlertRule(rule *AlertRule) {
	am.rules[rule.ID] = rule
	am.logger.Info("Added alert rule", zap.String("rule_id", rule.ID), zap.String("name", rule.Name))
}

// RemoveAlertRule removes an alert rule
func (am *AlertManager) RemoveAlertRule(ruleID string) {
	delete(am.rules, ruleID)
	am.logger.Info("Removed alert rule", zap.String("rule_id", ruleID))
}

// EvaluateMetric evaluates a metric against alert rules
func (am *AlertManager) EvaluateMetric(metric string, value float64, tenantID string) {
	for _, rule := range am.rules {
		if !rule.Enabled || rule.Metric != metric {
			continue
		}

		// Check if rule applies to this tenant
		if rule.TenantID != "" && rule.TenantID != tenantID {
			continue
		}

		// Evaluate condition
		shouldAlert := false
		switch rule.Condition {
		case "greater_than":
			shouldAlert = value > rule.Threshold
		case "less_than":
			shouldAlert = value < rule.Threshold
		case "equals":
			shouldAlert = value == rule.Threshold
		case "not_equals":
			shouldAlert = value != rule.Threshold
		}

		if shouldAlert {
			am.createAlert(rule, value, tenantID)
		} else {
			am.resolveAlert(rule.ID, tenantID)
		}
	}
}

// createAlert creates a new alert
func (am *AlertManager) createAlert(rule *AlertRule, value float64, tenantID string) {
	alertID := fmt.Sprintf("%s_%s_%s", rule.ID, tenantID, time.Now().Format("20060102150405"))

	// Check if alert already exists
	if existingAlert, exists := am.alerts[alertID]; exists && existingAlert.Status == AlertStatusActive {
		return
	}

	alert := &Alert{
		ID:          alertID,
		Title:       rule.Name,
		Description: rule.Description,
		Severity:    rule.Severity,
		Status:      AlertStatusActive,
		Source:      "prometheus",
		Metric:      rule.Metric,
		Value:       value,
		Threshold:   rule.Threshold,
		TenantID:    tenantID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"rule_id":   rule.ID,
			"condition": rule.Condition,
		},
	}

	am.alerts[alertID] = alert

	// Send alert to all channels
	am.sendAlert(alert)

	am.logger.Warn("Alert created",
		zap.String("alert_id", alertID),
		zap.String("rule_id", rule.ID),
		zap.String("severity", string(alert.Severity)),
		zap.String("tenant_id", tenantID),
		zap.Float64("value", value),
		zap.Float64("threshold", rule.Threshold),
	)
}

// resolveAlert resolves an existing alert
func (am *AlertManager) resolveAlert(ruleID, tenantID string) {
	for alertID, alert := range am.alerts {
		if alert.Status == AlertStatusActive &&
			alert.Metadata["rule_id"] == ruleID &&
			alert.TenantID == tenantID {

			now := time.Now()
			alert.Status = AlertStatusResolved
			alert.UpdatedAt = now
			alert.ResolvedAt = &now

			am.logger.Info("Alert resolved",
				zap.String("alert_id", alertID),
				zap.String("rule_id", ruleID),
				zap.String("tenant_id", tenantID),
			)
		}
	}
}

// sendAlert sends an alert to all configured channels
func (am *AlertManager) sendAlert(alert *Alert) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, channel := range am.channels {
		if !channel.IsEnabled() {
			continue
		}

		go func(ch AlertChannel) {
			if err := ch.SendAlert(ctx, alert); err != nil {
				am.logger.Error("Failed to send alert",
					zap.String("channel", ch.GetName()),
					zap.String("alert_id", alert.ID),
					zap.Error(err),
				)
			}
		}(channel)
	}
}

// GetActiveAlerts returns all active alerts
func (am *AlertManager) GetActiveAlerts(tenantID string) []*Alert {
	var activeAlerts []*Alert
	for _, alert := range am.alerts {
		if alert.Status == AlertStatusActive {
			if tenantID == "" || alert.TenantID == tenantID {
				activeAlerts = append(activeAlerts, alert)
			}
		}
	}
	return activeAlerts
}

// GetAlertHistory returns alert history
func (am *AlertManager) GetAlertHistory(tenantID string, limit int) []*Alert {
	var history []*Alert

	// Sort alerts by creation time (newest first)
	for _, alert := range am.alerts {
		if tenantID == "" || alert.TenantID == tenantID {
			history = append(history, alert)
		}
	}

	// Sort by creation time
	for i := 0; i < len(history)-1; i++ {
		for j := i + 1; j < len(history); j++ {
			if history[i].CreatedAt.Before(history[j].CreatedAt) {
				history[i], history[j] = history[j], history[i]
			}
		}
	}

	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	return history
}

// SuppressAlert suppresses an alert
func (am *AlertManager) SuppressAlert(alertID string, duration time.Duration) error {
	alert, exists := am.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	alert.Status = AlertStatusSuppressed
	alert.UpdatedAt = time.Now()

	// Auto-resolve after suppression duration
	go func() {
		time.Sleep(duration)
		alert.Status = AlertStatusResolved
		now := time.Now()
		alert.UpdatedAt = now
		alert.ResolvedAt = &now
	}()

	am.logger.Info("Alert suppressed",
		zap.String("alert_id", alertID),
		zap.Duration("duration", duration),
	)

	return nil
}

// EmailAlertChannel implementation

// NewEmailAlertChannel creates a new email alert channel
func NewEmailAlertChannel(config EmailConfig, logger *zap.Logger) *EmailAlertChannel {
	return &EmailAlertChannel{
		enabled: true,
		config:  config,
		logger:  logger,
	}
}

// SendAlert sends an alert via email
func (eac *EmailAlertChannel) SendAlert(ctx context.Context, alert *Alert) error {
	// Mock implementation - in a real implementation, this would send actual emails
	eac.logger.Info("Sending email alert",
		zap.String("alert_id", alert.ID),
		zap.String("severity", string(alert.Severity)),
		zap.String("title", alert.Title),
	)
	return nil
}

// GetName returns the channel name
func (eac *EmailAlertChannel) GetName() string {
	return "email"
}

// IsEnabled returns whether the channel is enabled
func (eac *EmailAlertChannel) IsEnabled() bool {
	return eac.enabled
}

// SlackAlertChannel implementation

// NewSlackAlertChannel creates a new Slack alert channel
func NewSlackAlertChannel(config SlackConfig, logger *zap.Logger) *SlackAlertChannel {
	return &SlackAlertChannel{
		enabled: true,
		config:  config,
		logger:  logger,
	}
}

// SendAlert sends an alert via Slack
func (sac *SlackAlertChannel) SendAlert(ctx context.Context, alert *Alert) error {
	// Mock implementation - in a real implementation, this would send actual Slack messages
	sac.logger.Info("Sending Slack alert",
		zap.String("alert_id", alert.ID),
		zap.String("severity", string(alert.Severity)),
		zap.String("title", alert.Title),
		zap.String("channel", sac.config.Channel),
	)
	return nil
}

// GetName returns the channel name
func (sac *SlackAlertChannel) GetName() string {
	return "slack"
}

// IsEnabled returns whether the channel is enabled
func (sac *SlackAlertChannel) IsEnabled() bool {
	return sac.enabled
}

// WebhookAlertChannel implementation

// NewWebhookAlertChannel creates a new webhook alert channel
func NewWebhookAlertChannel(config WebhookConfig, logger *zap.Logger) *WebhookAlertChannel {
	return &WebhookAlertChannel{
		enabled: true,
		config:  config,
		logger:  logger,
	}
}

// SendAlert sends an alert via webhook
func (wac *WebhookAlertChannel) SendAlert(ctx context.Context, alert *Alert) error {
	// Mock implementation - in a real implementation, this would send actual webhook requests
	wac.logger.Info("Sending webhook alert",
		zap.String("alert_id", alert.ID),
		zap.String("severity", string(alert.Severity)),
		zap.String("title", alert.Title),
		zap.String("url", wac.config.URL),
	)
	return nil
}

// GetName returns the channel name
func (wac *WebhookAlertChannel) GetName() string {
	return "webhook"
}

// IsEnabled returns whether the channel is enabled
func (wac *WebhookAlertChannel) IsEnabled() bool {
	return wac.enabled
}
