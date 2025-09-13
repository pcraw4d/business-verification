package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AlertingSystem manages cache alerts and notifications
type AlertingSystem struct {
	config           *AlertingConfig
	logger           *zap.Logger
	metricsCollector *MetricsCollector

	// Alert state
	activeAlerts map[string]*CacheMonitoringAlert
	alertHistory []*CacheMonitoringAlert
	mu           sync.RWMutex

	// Alert handlers
	handlers map[string][]AlertHandler

	// Cooldown management
	cooldowns map[string]time.Time
}

// AlertingConfig holds configuration for the alerting system
type AlertingConfig struct {
	EnableAlerts      bool                      `json:"enable_alerts"`
	AlertThresholds   map[string]AlertThreshold `json:"alert_thresholds"`
	CooldownPeriod    time.Duration             `json:"cooldown_period"`
	MaxHistoryEntries int                       `json:"max_history_entries"`
	EnableEscalation  bool                      `json:"enable_escalation"`
	EscalationDelay   time.Duration             `json:"escalation_delay"`
}

// AlertThreshold defines thresholds for different alert types
type AlertThreshold struct {
	Warning  float64 `json:"warning"`
	Critical float64 `json:"critical"`
	Enabled  bool    `json:"enabled"`
}

// CacheMonitoringAlert represents a cache monitoring alert
type CacheMonitoringAlert struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Severity     AlertSeverity          `json:"severity"`
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details"`
	Threshold    float64                `json:"threshold"`
	CurrentValue float64                `json:"current_value"`
	Timestamp    time.Time              `json:"timestamp"`
	ResolvedAt   *time.Time             `json:"resolved_at,omitempty"`
	Escalated    bool                   `json:"escalated"`
}

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertHandler defines the interface for alert handlers
type AlertHandler interface {
	HandleAlert(ctx context.Context, alert *CacheMonitoringAlert) error
	GetName() string
}

// LoggingAlertHandler logs alerts to the logger
type LoggingAlertHandler struct {
	logger *zap.Logger
}

// NewLoggingAlertHandler creates a new logging alert handler
func NewLoggingAlertHandler(logger *zap.Logger) *LoggingAlertHandler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &LoggingAlertHandler{
		logger: logger,
	}
}

// HandleAlert implements the AlertHandler interface
func (lah *LoggingAlertHandler) HandleAlert(ctx context.Context, alert *CacheMonitoringAlert) error {
	level := zap.InfoLevel
	switch alert.Severity {
	case AlertSeverityWarning:
		level = zap.WarnLevel
	case AlertSeverityCritical:
		level = zap.ErrorLevel
	}

	lah.logger.Log(level, "Cache alert triggered",
		zap.String("alert_id", alert.ID),
		zap.String("type", alert.Type),
		zap.String("severity", string(alert.Severity)),
		zap.String("message", alert.Message),
		zap.Float64("threshold", alert.Threshold),
		zap.Float64("current_value", alert.CurrentValue),
		zap.Time("timestamp", alert.Timestamp))

	return nil
}

// GetName returns the handler name
func (lah *LoggingAlertHandler) GetName() string {
	return "logging"
}

// WebhookAlertHandler sends alerts to a webhook endpoint
type WebhookAlertHandler struct {
	url     string
	timeout time.Duration
	logger  *zap.Logger
}

// NewWebhookAlertHandler creates a new webhook alert handler
func NewWebhookAlertHandler(url string, timeout time.Duration, logger *zap.Logger) *WebhookAlertHandler {
	if logger == nil {
		logger = zap.NewNop()
	}

	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &WebhookAlertHandler{
		url:     url,
		timeout: timeout,
		logger:  logger,
	}
}

// HandleAlert implements the AlertHandler interface
func (wah *WebhookAlertHandler) HandleAlert(ctx context.Context, alert *CacheMonitoringAlert) error {
	// In a real implementation, you would send an HTTP POST request to the webhook URL
	// For now, we'll just log the webhook call
	wah.logger.Info("Webhook alert handler called",
		zap.String("url", wah.url),
		zap.String("alert_id", alert.ID),
		zap.String("type", alert.Type),
		zap.String("severity", string(alert.Severity)))

	// TODO: Implement actual webhook HTTP call
	return nil
}

// GetName returns the handler name
func (wah *WebhookAlertHandler) GetName() string {
	return "webhook"
}

// NewAlertingSystem creates a new alerting system
func NewAlertingSystem(config *AlertingConfig, metricsCollector *MetricsCollector, logger *zap.Logger) *AlertingSystem {
	if config == nil {
		config = &AlertingConfig{
			EnableAlerts:      true,
			CooldownPeriod:    5 * time.Minute,
			MaxHistoryEntries: 1000,
			EnableEscalation:  true,
			EscalationDelay:   10 * time.Minute,
			AlertThresholds: map[string]AlertThreshold{
				"hit_rate_low": {
					Warning:  0.7,
					Critical: 0.5,
					Enabled:  true,
				},
				"memory_usage_high": {
					Warning:  100 * 1024 * 1024, // 100MB
					Critical: 500 * 1024 * 1024, // 500MB
					Enabled:  true,
				},
				"error_rate_high": {
					Warning:  0.05, // 5%
					Critical: 0.1,  // 10%
					Enabled:  true,
				},
				"cache_size_high": {
					Warning:  10000,
					Critical: 50000,
					Enabled:  true,
				},
			},
		}
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &AlertingSystem{
		config:           config,
		logger:           logger,
		metricsCollector: metricsCollector,
		activeAlerts:     make(map[string]*CacheMonitoringAlert),
		alertHistory:     make([]*CacheMonitoringAlert, 0),
		handlers:         make(map[string][]AlertHandler),
		cooldowns:        make(map[string]time.Time),
	}
}

// RegisterHandler registers an alert handler
func (as *AlertingSystem) RegisterHandler(alertType string, handler AlertHandler) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.handlers[alertType] = append(as.handlers[alertType], handler)

	as.logger.Debug("Registered alert handler",
		zap.String("alert_type", alertType),
		zap.String("handler", handler.GetName()))
}

// CheckAlerts checks for alert conditions based on current metrics
func (as *AlertingSystem) CheckAlerts(ctx context.Context) []*CacheMonitoringAlert {
	if !as.config.EnableAlerts {
		return nil
	}

	metrics := as.metricsCollector.GetCurrentMetrics()
	if metrics == nil {
		return nil
	}

	var newAlerts []*CacheMonitoringAlert

	// Check each threshold
	for alertType, threshold := range as.config.AlertThresholds {
		if !threshold.Enabled {
			continue
		}

		// Check if we're in cooldown for this alert type
		if as.isInCooldown(alertType) {
			continue
		}

		// Get current value for the metric
		currentValue, err := as.getMetricValue(metrics, alertType)
		if err != nil {
			as.logger.Error("Failed to get metric value",
				zap.String("alert_type", alertType),
				zap.Error(err))
			continue
		}

		// Check thresholds
		severity := as.checkThreshold(alertType, currentValue, threshold)
		if severity != "" {
			alert := as.createAlert(alertType, severity, currentValue, threshold)
			if alert != nil {
				newAlerts = append(newAlerts, alert)
			}
		}
	}

	// Process new alerts
	for _, alert := range newAlerts {
		as.processAlert(ctx, alert)
	}

	return newAlerts
}

// isInCooldown checks if an alert type is in cooldown period
func (as *AlertingSystem) isInCooldown(alertType string) bool {
	as.mu.RLock()
	defer as.mu.RUnlock()

	cooldownTime, exists := as.cooldowns[alertType]
	if !exists {
		return false
	}

	return time.Now().Before(cooldownTime)
}

// checkThreshold checks if a value exceeds the threshold
func (as *AlertingSystem) checkThreshold(alertType string, value float64, threshold AlertThreshold) AlertSeverity {
	// Determine if this is a "lower is worse" metric based on the alert type
	isLowerWorse := as.isLowerWorseMetric(alertType)

	if isLowerWorse {
		// For metrics where lower is worse (like hit rate), we check if value is below threshold
		switch {
		case value <= threshold.Critical:
			return AlertSeverityCritical
		case value <= threshold.Warning:
			return AlertSeverityWarning
		default:
			return ""
		}
	} else {
		// For metrics where higher is worse (like error rate, memory usage), we check if value is above threshold
		switch {
		case value >= threshold.Critical:
			return AlertSeverityCritical
		case value >= threshold.Warning:
			return AlertSeverityWarning
		default:
			return ""
		}
	}
}

// isLowerWorseMetric determines if a metric type is one where lower values are worse
func (as *AlertingSystem) isLowerWorseMetric(alertType string) bool {
	lowerWorseMetrics := map[string]bool{
		"hit_rate_low":     true,
		"cache_hit_rate":   true,
		"performance_good": true,
	}

	return lowerWorseMetrics[alertType]
}

// getMetricValue extracts the metric value from aggregated metrics
func (as *AlertingSystem) getMetricValue(metrics *AggregatedMetrics, alertType string) (float64, error) {
	switch alertType {
	case "hit_rate_low":
		return metrics.AverageHitRate, nil
	case "error_rate_high":
		return metrics.AverageErrorRate, nil
	case "memory_usage_high":
		return float64(metrics.TotalMemoryUsage), nil
	case "cache_size_high":
		return float64(metrics.TotalSize), nil
	default:
		return 0, fmt.Errorf("unknown alert type: %s", alertType)
	}
}

// createAlert creates a new alert
func (as *AlertingSystem) createAlert(alertType string, severity AlertSeverity, currentValue float64, threshold AlertThreshold) *CacheMonitoringAlert {
	alertID := fmt.Sprintf("%s_%d", alertType, time.Now().Unix())

	// Check if we already have an active alert for this type
	as.mu.RLock()
	if existingAlert, exists := as.activeAlerts[alertType]; exists {
		as.mu.RUnlock()
		// Update existing alert if severity is higher
		if as.getSeverityLevel(severity) > as.getSeverityLevel(existingAlert.Severity) {
			existingAlert.Severity = severity
			existingAlert.CurrentValue = currentValue
			existingAlert.Timestamp = time.Now()
			return existingAlert
		}
		return nil // Don't create duplicate alert
	}
	as.mu.RUnlock()

	alert := &CacheMonitoringAlert{
		ID:           alertID,
		Type:         alertType,
		Severity:     severity,
		Message:      as.generateAlertMessage(alertType, severity, currentValue, threshold),
		Details:      make(map[string]interface{}),
		Threshold:    as.getThresholdValue(threshold, severity),
		CurrentValue: currentValue,
		Timestamp:    time.Now(),
	}

	// Add to active alerts
	as.mu.Lock()
	as.activeAlerts[alertType] = alert
	as.mu.Unlock()

	// Set cooldown
	as.setCooldown(alertType)

	return alert
}

// getSeverityLevel returns a numeric level for severity comparison
func (as *AlertingSystem) getSeverityLevel(severity AlertSeverity) int {
	switch severity {
	case AlertSeverityInfo:
		return 1
	case AlertSeverityWarning:
		return 2
	case AlertSeverityCritical:
		return 3
	default:
		return 0
	}
}

// generateAlertMessage generates a human-readable alert message
func (as *AlertingSystem) generateAlertMessage(alertType string, severity AlertSeverity, currentValue float64, threshold AlertThreshold) string {
	switch alertType {
	case "hit_rate_low":
		return fmt.Sprintf("Cache hit rate is %s: %.2f%% (threshold: %.2f%%)",
			string(severity), currentValue*100, threshold.Warning*100)
	case "error_rate_high":
		return fmt.Sprintf("Cache error rate is %s: %.2f%% (threshold: %.2f%%)",
			string(severity), currentValue*100, threshold.Warning*100)
	case "memory_usage_high":
		return fmt.Sprintf("Cache memory usage is %s: %.2f MB (threshold: %.2f MB)",
			string(severity), float64(currentValue)/(1024*1024), float64(threshold.Warning)/(1024*1024))
	case "cache_size_high":
		return fmt.Sprintf("Cache size is %s: %d entries (threshold: %d entries)",
			string(severity), int64(currentValue), int64(threshold.Warning))
	default:
		return fmt.Sprintf("Cache alert: %s is %s", alertType, string(severity))
	}
}

// getThresholdValue returns the threshold value for the given severity
func (as *AlertingSystem) getThresholdValue(threshold AlertThreshold, severity AlertSeverity) float64 {
	switch severity {
	case AlertSeverityCritical:
		return threshold.Critical
	case AlertSeverityWarning:
		return threshold.Warning
	default:
		return threshold.Warning
	}
}

// setCooldown sets a cooldown period for an alert type
func (as *AlertingSystem) setCooldown(alertType string) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.cooldowns[alertType] = time.Now().Add(as.config.CooldownPeriod)
}

// processAlert processes a new alert
func (as *AlertingSystem) processAlert(ctx context.Context, alert *CacheMonitoringAlert) {
	as.logger.Info("Processing cache alert",
		zap.String("alert_id", alert.ID),
		zap.String("type", alert.Type),
		zap.String("severity", string(alert.Severity)))

	// Add to history
	as.addToHistory(alert)

	// Notify handlers
	as.notifyHandlers(ctx, alert)

	// Start escalation timer if enabled
	if as.config.EnableEscalation && alert.Severity == AlertSeverityCritical {
		go as.escalateAlert(ctx, alert)
	}
}

// addToHistory adds an alert to the history
func (as *AlertingSystem) addToHistory(alert *CacheMonitoringAlert) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.alertHistory = append(as.alertHistory, alert)

	// Trim history if it exceeds max entries
	if len(as.alertHistory) > as.config.MaxHistoryEntries {
		as.alertHistory = as.alertHistory[1:]
	}
}

// notifyHandlers notifies all registered handlers for the alert type
func (as *AlertingSystem) notifyHandlers(ctx context.Context, alert *CacheMonitoringAlert) {
	as.mu.RLock()
	handlers := as.handlers[alert.Type]
	as.mu.RUnlock()

	for _, handler := range handlers {
		go func(h AlertHandler) {
			if err := h.HandleAlert(ctx, alert); err != nil {
				as.logger.Error("Alert handler failed",
					zap.String("handler", h.GetName()),
					zap.String("alert_id", alert.ID),
					zap.Error(err))
			}
		}(handler)
	}
}

// escalateAlert escalates a critical alert after a delay
func (as *AlertingSystem) escalateAlert(ctx context.Context, alert *CacheMonitoringAlert) {
	time.Sleep(as.config.EscalationDelay)

	// Check if alert is still active and unresolved
	as.mu.RLock()
	activeAlert, exists := as.activeAlerts[alert.Type]
	as.mu.RUnlock()

	if exists && activeAlert.ID == alert.ID && activeAlert.ResolvedAt == nil {
		alert.Escalated = true
		alert.Severity = AlertSeverityCritical // Ensure it's still critical

		as.logger.Warn("Escalating critical cache alert",
			zap.String("alert_id", alert.ID),
			zap.String("type", alert.Type))

		// Notify handlers again
		as.notifyHandlers(ctx, alert)
	}
}

// ResolveAlert resolves an active alert
func (as *AlertingSystem) ResolveAlert(alertType string) {
	as.mu.Lock()
	defer as.mu.Unlock()

	if alert, exists := as.activeAlerts[alertType]; exists {
		now := time.Now()
		alert.ResolvedAt = &now

		as.logger.Info("Resolved cache alert",
			zap.String("alert_id", alert.ID),
			zap.String("type", alert.Type))

		// Remove from active alerts
		delete(as.activeAlerts, alertType)
	}
}

// GetActiveAlerts returns all currently active alerts
func (as *AlertingSystem) GetActiveAlerts() []*CacheMonitoringAlert {
	as.mu.RLock()
	defer as.mu.RUnlock()

	var alerts []*CacheMonitoringAlert
	for _, alert := range as.activeAlerts {
		alerts = append(alerts, alert)
	}

	return alerts
}

// GetAlertHistory returns the alert history
func (as *AlertingSystem) GetAlertHistory() []*CacheMonitoringAlert {
	as.mu.RLock()
	defer as.mu.RUnlock()

	// Return a copy of the history
	history := make([]*CacheMonitoringAlert, len(as.alertHistory))
	copy(history, as.alertHistory)

	return history
}

// GetConfig returns the alerting configuration
func (as *AlertingSystem) GetConfig() *AlertingConfig {
	return as.config
}

// UpdateConfig updates the alerting configuration
func (as *AlertingSystem) UpdateConfig(config *AlertingConfig) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.config = config
	as.logger.Info("Updated alerting system configuration",
		zap.Bool("alerts_enabled", config.EnableAlerts),
		zap.Duration("cooldown_period", config.CooldownPeriod))
}
