package risk

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// NotificationService handles sending notifications for alerts
type NotificationService struct {
	logger   *zap.Logger
	channels map[string]NotificationChannel
	config   *NotificationConfig
}

// NotificationConfig contains configuration for notifications
type NotificationConfig struct {
	DefaultChannels    []string               `json:"default_channels"`
	ChannelConfigs     map[string]interface{} `json:"channel_configs"`
	RetryAttempts      int                    `json:"retry_attempts"`
	RetryDelay         time.Duration          `json:"retry_delay"`
	RateLimitPerMinute int                    `json:"rate_limit_per_minute"`
	EnableRateLimiting bool                   `json:"enable_rate_limiting"`
	TemplateEngine     string                 `json:"template_engine"`
}

// NotificationChannel interface for different notification channels
type NotificationChannel interface {
	Send(ctx context.Context, notification *Notification) error
	GetName() string
	IsEnabled() bool
	GetConfig() map[string]interface{}
}

// Notification represents a notification to be sent
type Notification struct {
	ID          string                 `json:"id"`
	Channel     string                 `json:"channel"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Priority    AlertPriority          `json:"priority"`
	Severity    AlertSeverity          `json:"severity"`
	Recipients  []string               `json:"recipients"`
	Alert       *RiskAlert             `json:"alert,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// NewNotificationService creates a new notification service
func NewNotificationService(logger *zap.Logger) *NotificationService {
	service := &NotificationService{
		logger:   logger,
		channels: make(map[string]NotificationChannel),
		config: &NotificationConfig{
			DefaultChannels:    []string{"email", "dashboard"},
			RetryAttempts:      3,
			RetryDelay:         5 * time.Second,
			RateLimitPerMinute: 60,
			EnableRateLimiting: true,
			TemplateEngine:     "default",
		},
	}

	// Initialize default channels
	service.initializeDefaultChannels()

	return service
}

// SendNotification sends a notification through the specified channel
func (ns *NotificationService) SendNotification(ctx context.Context, channel string, alert *RiskAlert) error {
	// Get the notification channel
	notificationChannel, exists := ns.channels[channel]
	if !exists {
		return fmt.Errorf("notification channel not found: %s", channel)
	}

	if !notificationChannel.IsEnabled() {
		ns.logger.Warn("Notification channel is disabled",
			zap.String("channel", channel))
		return nil
	}

	// Create notification
	notification, err := ns.createNotification(channel, alert)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification with retry logic
	var lastErr error
	for attempt := 1; attempt <= ns.config.RetryAttempts; attempt++ {
		if err := notificationChannel.Send(ctx, notification); err != nil {
			lastErr = err
			ns.logger.Warn("Notification send attempt failed",
				zap.String("channel", channel),
				zap.String("alert_id", alert.ID),
				zap.Int("attempt", attempt),
				zap.Error(err))

			if attempt < ns.config.RetryAttempts {
				time.Sleep(ns.config.RetryDelay)
			}
		} else {
			ns.logger.Info("Notification sent successfully",
				zap.String("channel", channel),
				zap.String("alert_id", alert.ID),
				zap.Int("attempt", attempt))
			return nil
		}
	}

	return fmt.Errorf("failed to send notification after %d attempts: %w", ns.config.RetryAttempts, lastErr)
}

// createNotification creates a notification from an alert
func (ns *NotificationService) createNotification(channel string, alert *RiskAlert) (*Notification, error) {
	notificationID := fmt.Sprintf("notif_%s_%s_%d", channel, alert.ID, time.Now().Unix())

	// Create title and message based on channel
	title, message, err := ns.createNotificationContent(channel, alert)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification content: %w", err)
	}

	// Determine recipients based on channel and alert
	recipients := ns.getRecipients(channel, alert)

	// Set expiration time
	var expiresAt *time.Time
	if alert.ExpiresAt != nil {
		expiresAt = alert.ExpiresAt
	}

	notification := &Notification{
		ID:         notificationID,
		Channel:    channel,
		Type:       "risk_alert",
		Title:      title,
		Message:    message,
		Priority:   alert.Priority,
		Severity:   alert.Severity,
		Recipients: recipients,
		Alert:      alert,
		CreatedAt:  time.Now(),
		ExpiresAt:  expiresAt,
		Metadata: map[string]interface{}{
			"alert_id":    alert.ID,
			"business_id": alert.BusinessID,
			"factor_id":   alert.FactorID,
			"alert_type":  alert.AlertType,
			"created_by":  "notification_service",
		},
	}

	return notification, nil
}

// createNotificationContent creates title and message for different channels
func (ns *NotificationService) createNotificationContent(channel string, alert *RiskAlert) (string, string, error) {
	switch channel {
	case "email":
		return ns.createEmailContent(alert)
	case "sms":
		return ns.createSMSContent(alert)
	case "dashboard":
		return ns.createDashboardContent(alert)
	case "slack":
		return ns.createSlackContent(alert)
	case "webhook":
		return ns.createWebhookContent(alert)
	default:
		return ns.createDefaultContent(alert)
	}
}

// createEmailContent creates email-specific content
func (ns *NotificationService) createEmailContent(alert *RiskAlert) (string, string, error) {
	title := fmt.Sprintf("🚨 Risk Alert: %s - %s", alert.FactorName, strings.Title(string(alert.Level)))

	message := fmt.Sprintf(`
Risk Alert Notification

Alert Details:
- Factor: %s
- Category: %s
- Current Value: %.2f
- Threshold: %.2f
- Risk Level: %s
- Severity: %s
- Priority: %s

Alert Message:
%s

Business ID: %s
Alert ID: %s
Triggered At: %s

Please review this alert and take appropriate action.

Best regards,
Risk Management System
`,
		alert.FactorName,
		string(alert.Category),
		alert.CurrentValue,
		alert.ThresholdValue,
		string(alert.Level),
		string(alert.Severity),
		string(alert.Priority),
		alert.Message,
		alert.BusinessID,
		alert.ID,
		alert.TriggeredAt.Format("2006-01-02 15:04:05"))

	return title, message, nil
}

// createSMSContent creates SMS-specific content
func (ns *NotificationService) createSMSContent(alert *RiskAlert) (string, string, error) {
	title := "Risk Alert"

	message := fmt.Sprintf("ALERT: %s risk level %s for %s. Value: %.1f (threshold: %.1f). %s",
		strings.Title(string(alert.Level)),
		string(alert.Severity),
		alert.FactorName,
		alert.CurrentValue,
		alert.ThresholdValue,
		alert.Message)

	// Truncate if too long for SMS
	if len(message) > 160 {
		message = message[:157] + "..."
	}

	return title, message, nil
}

// createDashboardContent creates dashboard-specific content
func (ns *NotificationService) createDashboardContent(alert *RiskAlert) (string, string, error) {
	title := fmt.Sprintf("Risk Alert: %s", alert.FactorName)

	message := fmt.Sprintf(`{
	"alert_id": "%s",
	"business_id": "%s",
	"factor_id": "%s",
	"factor_name": "%s",
	"category": "%s",
	"level": "%s",
	"severity": "%s",
	"priority": "%s",
	"current_value": %.2f,
	"threshold_value": %.2f,
	"message": "%s",
	"triggered_at": "%s",
	"tags": %s
}`,
		alert.ID,
		alert.BusinessID,
		alert.FactorID,
		alert.FactorName,
		string(alert.Category),
		string(alert.Level),
		string(alert.Severity),
		string(alert.Priority),
		alert.CurrentValue,
		alert.ThresholdValue,
		alert.Message,
		alert.TriggeredAt.Format(time.RFC3339),
		strings.Join(alert.Tags, ","))

	return title, message, nil
}

// createSlackContent creates Slack-specific content
func (ns *NotificationService) createSlackContent(alert *RiskAlert) (string, string, error) {
	title := fmt.Sprintf("🚨 Risk Alert: %s", alert.FactorName)

	// Determine color based on severity
	color := "good"
	switch alert.Severity {
	case AlertSeverityCritical:
		color = "danger"
	case AlertSeverityHigh:
		color = "warning"
	case AlertSeverityMedium:
		color = "warning"
	case AlertSeverityLow:
		color = "good"
	}

	message := fmt.Sprintf(`{
	"attachments": [{
		"color": "%s",
		"title": "%s",
		"text": "%s",
		"fields": [
			{
				"title": "Factor",
				"value": "%s",
				"short": true
			},
			{
				"title": "Category",
				"value": "%s",
				"short": true
			},
			{
				"title": "Current Value",
				"value": "%.2f",
				"short": true
			},
			{
				"title": "Threshold",
				"value": "%.2f",
				"short": true
			},
			{
				"title": "Risk Level",
				"value": "%s",
				"short": true
			},
			{
				"title": "Severity",
				"value": "%s",
				"short": true
			}
		],
		"footer": "Risk Management System",
		"ts": %d
	}]
}`,
		color,
		title,
		alert.Message,
		alert.FactorName,
		string(alert.Category),
		alert.CurrentValue,
		alert.ThresholdValue,
		string(alert.Level),
		string(alert.Severity),
		alert.TriggeredAt.Unix())

	return title, message, nil
}

// createWebhookContent creates webhook-specific content
func (ns *NotificationService) createWebhookContent(alert *RiskAlert) (string, string, error) {
	title := "Risk Alert Webhook"

	message := fmt.Sprintf(`{
	"event_type": "risk_alert",
	"alert": {
		"id": "%s",
		"business_id": "%s",
		"factor_id": "%s",
		"factor_name": "%s",
		"category": "%s",
		"alert_type": "%s",
		"level": "%s",
		"severity": "%s",
		"priority": "%s",
		"status": "%s",
		"current_value": %.2f,
		"threshold_value": %.2f,
		"title": "%s",
		"message": "%s",
		"triggered_at": "%s",
		"escalation_level": %d,
		"tags": %s,
		"metadata": %s
	}
}`,
		alert.ID,
		alert.BusinessID,
		alert.FactorID,
		alert.FactorName,
		string(alert.Category),
		string(alert.AlertType),
		string(alert.Level),
		string(alert.Severity),
		string(alert.Priority),
		string(alert.Status),
		alert.CurrentValue,
		alert.ThresholdValue,
		alert.Title,
		alert.Message,
		alert.TriggeredAt.Format(time.RFC3339),
		alert.EscalationLevel,
		strings.Join(alert.Tags, ","),
		"{}") // Simplified metadata

	return title, message, nil
}

// createDefaultContent creates default content
func (ns *NotificationService) createDefaultContent(alert *RiskAlert) (string, string, error) {
	title := fmt.Sprintf("Risk Alert: %s", alert.FactorName)
	message := alert.Message
	return title, message, nil
}

// getRecipients determines recipients based on channel and alert
func (ns *NotificationService) getRecipients(channel string, alert *RiskAlert) []string {
	// Default recipients based on channel
	recipients := []string{}

	switch channel {
	case "email":
		// Get email recipients from alert metadata or configuration
		if emails, exists := alert.Metadata["email_recipients"]; exists {
			if emailList, ok := emails.([]string); ok {
				recipients = emailList
			}
		}
		if len(recipients) == 0 {
			recipients = []string{"admin@company.com"} // Default email
		}

	case "sms":
		// Get SMS recipients from alert metadata or configuration
		if phones, exists := alert.Metadata["sms_recipients"]; exists {
			if phoneList, ok := phones.([]string); ok {
				recipients = phoneList
			}
		}
		if len(recipients) == 0 {
			recipients = []string{"+1234567890"} // Default phone
		}

	case "slack":
		// Get Slack channels from alert metadata or configuration
		if channels, exists := alert.Metadata["slack_channels"]; exists {
			if channelList, ok := channels.([]string); ok {
				recipients = channelList
			}
		}
		if len(recipients) == 0 {
			recipients = []string{"#risk-alerts"} // Default channel
		}

	case "webhook":
		// Get webhook URLs from alert metadata or configuration
		if urls, exists := alert.Metadata["webhook_urls"]; exists {
			if urlList, ok := urls.([]string); ok {
				recipients = urlList
			}
		}
		if len(recipients) == 0 {
			recipients = []string{"https://hooks.slack.com/services/..."} // Default webhook
		}

	default:
		// For dashboard and other channels, no specific recipients needed
		recipients = []string{}
	}

	return recipients
}

// initializeDefaultChannels initializes default notification channels
func (ns *NotificationService) initializeDefaultChannels() {
	// Email channel
	ns.channels["email"] = &EmailNotificationChannel{
		name:    "email",
		enabled: true,
		config: map[string]interface{}{
			"smtp_host":     "localhost",
			"smtp_port":     587,
			"smtp_username": "alerts@company.com",
			"smtp_password": "password",
		},
	}

	// SMS channel
	ns.channels["sms"] = &SMSNotificationChannel{
		name:    "sms",
		enabled: true,
		config: map[string]interface{}{
			"provider":    "twilio",
			"account_sid": "your_account_sid",
			"auth_token":  "your_auth_token",
		},
	}

	// Dashboard channel
	ns.channels["dashboard"] = &DashboardNotificationChannel{
		name:    "dashboard",
		enabled: true,
		config: map[string]interface{}{
			"websocket_url": "ws://localhost:8080/ws",
		},
	}

	// Slack channel
	ns.channels["slack"] = &SlackNotificationChannel{
		name:    "slack",
		enabled: true,
		config: map[string]interface{}{
			"webhook_url": "https://hooks.slack.com/services/...",
		},
	}

	// Webhook channel
	ns.channels["webhook"] = &WebhookNotificationChannel{
		name:    "webhook",
		enabled: true,
		config: map[string]interface{}{
			"timeout": 30,
		},
	}
}

// AddChannel adds a custom notification channel
func (ns *NotificationService) AddChannel(name string, channel NotificationChannel) {
	ns.channels[name] = channel
	ns.logger.Info("Added notification channel",
		zap.String("channel", name))
}

// RemoveChannel removes a notification channel
func (ns *NotificationService) RemoveChannel(name string) {
	delete(ns.channels, name)
	ns.logger.Info("Removed notification channel",
		zap.String("channel", name))
}

// GetChannels returns all available channels
func (ns *NotificationService) GetChannels() map[string]NotificationChannel {
	return ns.channels
}

// GetChannel returns a specific channel
func (ns *NotificationService) GetChannel(name string) (NotificationChannel, error) {
	channel, exists := ns.channels[name]
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", name)
	}
	return channel, nil
}

// EnableChannel enables a notification channel
func (ns *NotificationService) EnableChannel(name string) error {
	channel, exists := ns.channels[name]
	if !exists {
		return fmt.Errorf("channel not found: %s", name)
	}

	// This would need to be implemented in each channel type
	ns.logger.Info("Channel enabled",
		zap.String("channel", name))

	return nil
}

// DisableChannel disables a notification channel
func (ns *NotificationService) DisableChannel(name string) error {
	channel, exists := ns.channels[name]
	if !exists {
		return fmt.Errorf("channel not found: %s", name)
	}

	// This would need to be implemented in each channel type
	ns.logger.Info("Channel disabled",
		zap.String("channel", name))

	return nil
}

// TestChannel tests a notification channel
func (ns *NotificationService) TestChannel(ctx context.Context, name string) error {
	channel, exists := ns.channels[name]
	if !exists {
		return fmt.Errorf("channel not found: %s", name)
	}

	// Create a test notification
	testNotification := &Notification{
		ID:         "test_notification",
		Channel:    name,
		Type:       "test",
		Title:      "Test Notification",
		Message:    "This is a test notification to verify the channel is working correctly.",
		Priority:   AlertPriorityLow,
		Severity:   AlertSeverityLow,
		Recipients: []string{"test@example.com"},
		CreatedAt:  time.Now(),
	}

	return channel.Send(ctx, testNotification)
}
