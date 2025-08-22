package error_monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DefaultAlertManager provides a default implementation of AlertManager
type DefaultAlertManager struct {
	logger        *zap.Logger
	mu            sync.RWMutex
	activeAlerts  map[string]Alert
	alertHandlers []AlertHandler
	config        AlertManagerConfig
}

// AlertManagerConfig contains configuration for the alert manager
type AlertManagerConfig struct {
	MaxActiveAlerts     int           `json:"max_active_alerts"`
	AlertRetentionTime  time.Duration `json:"alert_retention_time"`
	CooldownPeriod      time.Duration `json:"cooldown_period"`
	EnableEmailAlerts   bool          `json:"enable_email_alerts"`
	EnableSlackAlerts   bool          `json:"enable_slack_alerts"`
	EnableWebhookAlerts bool          `json:"enable_webhook_alerts"`
	EmailConfig         EmailConfig   `json:"email_config"`
	SlackConfig         SlackConfig   `json:"slack_config"`
	WebhookConfig       WebhookConfig `json:"webhook_config"`
}

// EmailConfig contains email alert configuration
type EmailConfig struct {
	SMTPHost     string   `json:"smtp_host"`
	SMTPPort     int      `json:"smtp_port"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	FromAddress  string   `json:"from_address"`
	ToAddresses  []string `json:"to_addresses"`
	Subject      string   `json:"subject"`
	TemplateFile string   `json:"template_file"`
}

// SlackConfig contains Slack alert configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconEmoji  string `json:"icon_emoji"`
}

// WebhookConfig contains webhook alert configuration
type WebhookConfig struct {
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	Timeout    time.Duration     `json:"timeout"`
	RetryCount int               `json:"retry_count"`
	RetryDelay time.Duration     `json:"retry_delay"`
}

// AlertHandler interface for handling different types of alerts
type AlertHandler interface {
	HandleAlert(ctx context.Context, alert Alert) error
	GetType() string
	IsEnabled() bool
}

// NewDefaultAlertManager creates a new default alert manager
func NewDefaultAlertManager(config AlertManagerConfig, logger *zap.Logger) *DefaultAlertManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config.MaxActiveAlerts == 0 {
		config.MaxActiveAlerts = 1000
	}

	if config.AlertRetentionTime == 0 {
		config.AlertRetentionTime = 24 * time.Hour
	}

	if config.CooldownPeriod == 0 {
		config.CooldownPeriod = 5 * time.Minute
	}

	manager := &DefaultAlertManager{
		logger:        logger,
		activeAlerts:  make(map[string]Alert),
		alertHandlers: make([]AlertHandler, 0),
		config:        config,
	}

	// Initialize alert handlers
	if config.EnableEmailAlerts {
		emailHandler := NewEmailAlertHandler(config.EmailConfig, logger)
		manager.AddAlertHandler(emailHandler)
	}

	if config.EnableSlackAlerts {
		slackHandler := NewSlackAlertHandler(config.SlackConfig, logger)
		manager.AddAlertHandler(slackHandler)
	}

	if config.EnableWebhookAlerts {
		webhookHandler := NewWebhookAlertHandler(config.WebhookConfig, logger)
		manager.AddAlertHandler(webhookHandler)
	}

	return manager
}

// SendAlert sends an alert through configured channels
func (am *DefaultAlertManager) SendAlert(ctx context.Context, alert Alert) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Check cooldown period
	if existingAlert, exists := am.activeAlerts[alert.ID]; exists {
		if time.Since(existingAlert.Timestamp) < am.config.CooldownPeriod {
			am.logger.Debug("Alert skipped due to cooldown period",
				zap.String("alert_id", alert.ID),
				zap.Duration("cooldown_remaining", am.config.CooldownPeriod-time.Since(existingAlert.Timestamp)))
			return nil
		}
	}

	// Check active alert limit
	if len(am.activeAlerts) >= am.config.MaxActiveAlerts {
		am.logger.Warn("Maximum active alerts reached, skipping alert",
			zap.String("alert_id", alert.ID),
			zap.Int("max_alerts", am.config.MaxActiveAlerts))
		return fmt.Errorf("maximum active alerts reached")
	}

	// Store alert
	am.activeAlerts[alert.ID] = alert

	// Send alert through all enabled handlers
	var errors []error
	for _, handler := range am.alertHandlers {
		if handler.IsEnabled() {
			if err := handler.HandleAlert(ctx, alert); err != nil {
				errors = append(errors, fmt.Errorf("handler %s failed: %w", handler.GetType(), err))
				am.logger.Error("Alert handler failed",
					zap.String("handler_type", handler.GetType()),
					zap.String("alert_id", alert.ID),
					zap.Error(err))
			}
		}
	}

	am.logger.Info("Alert sent",
		zap.String("alert_id", alert.ID),
		zap.String("level", alert.Level),
		zap.String("process", alert.ProcessName),
		zap.String("message", alert.Message))

	// Return combined errors if any
	if len(errors) > 0 {
		return fmt.Errorf("some alert handlers failed: %v", errors)
	}

	return nil
}

// ClearAlert clears an active alert
func (am *DefaultAlertManager) ClearAlert(ctx context.Context, alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if alert, exists := am.activeAlerts[alertID]; exists {
		alert.Resolved = true
		now := time.Now()
		alert.ResolvedAt = &now
		am.activeAlerts[alertID] = alert

		am.logger.Info("Alert cleared",
			zap.String("alert_id", alertID),
			zap.String("level", alert.Level),
			zap.String("process", alert.ProcessName))

		// Clean up old resolved alerts
		am.cleanupResolvedAlerts()
	}

	return nil
}

// GetActiveAlerts returns all active alerts
func (am *DefaultAlertManager) GetActiveAlerts(ctx context.Context) ([]Alert, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]Alert, 0, len(am.activeAlerts))
	for _, alert := range am.activeAlerts {
		if !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// AddAlertHandler adds an alert handler
func (am *DefaultAlertManager) AddAlertHandler(handler AlertHandler) {
	am.alertHandlers = append(am.alertHandlers, handler)
	am.logger.Info("Alert handler added", zap.String("type", handler.GetType()))
}

// RemoveAlertHandler removes an alert handler by type
func (am *DefaultAlertManager) RemoveAlertHandler(handlerType string) {
	for i, handler := range am.alertHandlers {
		if handler.GetType() == handlerType {
			am.alertHandlers = append(am.alertHandlers[:i], am.alertHandlers[i+1:]...)
			am.logger.Info("Alert handler removed", zap.String("type", handlerType))
			break
		}
	}
}

// cleanupResolvedAlerts removes old resolved alerts
func (am *DefaultAlertManager) cleanupResolvedAlerts() {
	cutoff := time.Now().Add(-am.config.AlertRetentionTime)

	for alertID, alert := range am.activeAlerts {
		if alert.Resolved && alert.ResolvedAt != nil && alert.ResolvedAt.Before(cutoff) {
			delete(am.activeAlerts, alertID)
		}
	}
}

// GetAlertStats returns alert statistics
func (am *DefaultAlertManager) GetAlertStats() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()

	stats := make(map[string]interface{})

	totalAlerts := len(am.activeAlerts)
	activeAlerts := 0
	resolvedAlerts := 0
	alertsByLevel := make(map[string]int)
	alertsByProcess := make(map[string]int)

	for _, alert := range am.activeAlerts {
		alertsByLevel[alert.Level]++
		alertsByProcess[alert.ProcessName]++

		if alert.Resolved {
			resolvedAlerts++
		} else {
			activeAlerts++
		}
	}

	stats["total_alerts"] = totalAlerts
	stats["active_alerts"] = activeAlerts
	stats["resolved_alerts"] = resolvedAlerts
	stats["alerts_by_level"] = alertsByLevel
	stats["alerts_by_process"] = alertsByProcess
	stats["max_active_alerts"] = am.config.MaxActiveAlerts
	stats["enabled_handlers"] = len(am.alertHandlers)

	return stats
}

// LogAlertHandler provides logging-based alert handling
type LogAlertHandler struct {
	logger  *zap.Logger
	enabled bool
}

// NewLogAlertHandler creates a new log alert handler
func NewLogAlertHandler(logger *zap.Logger) *LogAlertHandler {
	return &LogAlertHandler{
		logger:  logger,
		enabled: true,
	}
}

// HandleAlert handles an alert by logging it
func (lah *LogAlertHandler) HandleAlert(ctx context.Context, alert Alert) error {
	switch alert.Level {
	case "critical":
		lah.logger.Error("CRITICAL ALERT",
			zap.String("alert_id", alert.ID),
			zap.String("process", alert.ProcessName),
			zap.String("message", alert.Message),
			zap.Any("context", alert.Context))
	case "warning":
		lah.logger.Warn("WARNING ALERT",
			zap.String("alert_id", alert.ID),
			zap.String("process", alert.ProcessName),
			zap.String("message", alert.Message),
			zap.Any("context", alert.Context))
	default:
		lah.logger.Info("ALERT",
			zap.String("alert_id", alert.ID),
			zap.String("level", alert.Level),
			zap.String("process", alert.ProcessName),
			zap.String("message", alert.Message),
			zap.Any("context", alert.Context))
	}

	return nil
}

// GetType returns the handler type
func (lah *LogAlertHandler) GetType() string {
	return "log"
}

// IsEnabled returns whether the handler is enabled
func (lah *LogAlertHandler) IsEnabled() bool {
	return lah.enabled
}

// SetEnabled sets the enabled state
func (lah *LogAlertHandler) SetEnabled(enabled bool) {
	lah.enabled = enabled
}

// Placeholder implementations for other alert handlers
// These would be implemented with actual email, Slack, webhook logic

// EmailAlertHandler placeholder
type EmailAlertHandler struct {
	config  EmailConfig
	logger  *zap.Logger
	enabled bool
}

func NewEmailAlertHandler(config EmailConfig, logger *zap.Logger) *EmailAlertHandler {
	return &EmailAlertHandler{
		config:  config,
		logger:  logger,
		enabled: true,
	}
}

func (eah *EmailAlertHandler) HandleAlert(ctx context.Context, alert Alert) error {
	// TODO: Implement actual email sending logic
	eah.logger.Info("Email alert would be sent",
		zap.String("alert_id", alert.ID),
		zap.Strings("to_addresses", eah.config.ToAddresses))
	return nil
}

func (eah *EmailAlertHandler) GetType() string {
	return "email"
}

func (eah *EmailAlertHandler) IsEnabled() bool {
	return eah.enabled
}

// SlackAlertHandler placeholder
type SlackAlertHandler struct {
	config  SlackConfig
	logger  *zap.Logger
	enabled bool
}

func NewSlackAlertHandler(config SlackConfig, logger *zap.Logger) *SlackAlertHandler {
	return &SlackAlertHandler{
		config:  config,
		logger:  logger,
		enabled: true,
	}
}

func (sah *SlackAlertHandler) HandleAlert(ctx context.Context, alert Alert) error {
	// TODO: Implement actual Slack webhook logic
	sah.logger.Info("Slack alert would be sent",
		zap.String("alert_id", alert.ID),
		zap.String("channel", sah.config.Channel))
	return nil
}

func (sah *SlackAlertHandler) GetType() string {
	return "slack"
}

func (sah *SlackAlertHandler) IsEnabled() bool {
	return sah.enabled
}

// WebhookAlertHandler placeholder
type WebhookAlertHandler struct {
	config  WebhookConfig
	logger  *zap.Logger
	enabled bool
}

func NewWebhookAlertHandler(config WebhookConfig, logger *zap.Logger) *WebhookAlertHandler {
	return &WebhookAlertHandler{
		config:  config,
		logger:  logger,
		enabled: true,
	}
}

func (wah *WebhookAlertHandler) HandleAlert(ctx context.Context, alert Alert) error {
	// TODO: Implement actual webhook HTTP request logic
	wah.logger.Info("Webhook alert would be sent",
		zap.String("alert_id", alert.ID),
		zap.String("url", wah.config.URL))
	return nil
}

func (wah *WebhookAlertHandler) GetType() string {
	return "webhook"
}

func (wah *WebhookAlertHandler) IsEnabled() bool {
	return wah.enabled
}
