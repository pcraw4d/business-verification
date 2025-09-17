package classification_monitoring

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AdvancedAlertManager manages advanced alerting for accuracy tracking
type AdvancedAlertManager struct {
	config       *AdvancedAccuracyConfig
	logger       *zap.Logger
	mu           sync.RWMutex
	alerts       map[string]*AdvancedAlert
	alertHistory []*AdvancedAlert
	cooldowns    map[string]time.Time
	activeCount  int
}

// AdvancedAlert represents an advanced alert
type AdvancedAlert struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	Severity       string                 `json:"severity"`
	DimensionName  string                 `json:"dimension_name"`
	DimensionValue string                 `json:"dimension_value"`
	CurrentValue   float64                `json:"current_value"`
	ThresholdValue float64                `json:"threshold_value"`
	Message        string                 `json:"message"`
	Timestamp      time.Time              `json:"timestamp"`
	Actions        []string               `json:"actions"`
	Status         string                 `json:"status"` // "active", "resolved", "acknowledged"
	Metadata       map[string]interface{} `json:"metadata"`
	ResolvedAt     *time.Time             `json:"resolved_at,omitempty"`
	AcknowledgedAt *time.Time             `json:"acknowledged_at,omitempty"`
}

// NewAdvancedAlertManager creates a new advanced alert manager
func NewAdvancedAlertManager(config *AdvancedAccuracyConfig, logger *zap.Logger) *AdvancedAlertManager {
	return &AdvancedAlertManager{
		config:       config,
		logger:       logger,
		alerts:       make(map[string]*AdvancedAlert),
		alertHistory: make([]*AdvancedAlert, 0),
		cooldowns:    make(map[string]time.Time),
	}
}

// CreateAlert creates a new alert
func (aam *AdvancedAlertManager) CreateAlert(alertType, severity, message, dimensionName string, currentValue, thresholdValue float64) {
	aam.mu.Lock()
	defer aam.mu.Unlock()

	// Check cooldown
	cooldownKey := fmt.Sprintf("%s_%s_%s", alertType, dimensionName, severity)
	if cooldown, exists := aam.cooldowns[cooldownKey]; exists {
		if time.Since(cooldown) < aam.config.AlertCooldownPeriod {
			aam.logger.Debug("Alert suppressed due to cooldown",
				zap.String("alert_type", alertType),
				zap.String("dimension", dimensionName),
				zap.Duration("remaining_cooldown", aam.config.AlertCooldownPeriod-time.Since(cooldown)))
			return
		}
	}

	// Check alert rate limit
	if aam.activeCount >= aam.config.MaxAlertsPerHour {
		aam.logger.Warn("Alert rate limit exceeded, suppressing new alerts",
			zap.Int("active_alerts", aam.activeCount),
			zap.Int("max_alerts_per_hour", aam.config.MaxAlertsPerHour))
		return
	}

	// Create alert
	alert := &AdvancedAlert{
		ID:             fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Type:           alertType,
		Severity:       severity,
		DimensionName:  dimensionName,
		DimensionValue: "all",
		CurrentValue:   currentValue,
		ThresholdValue: thresholdValue,
		Message:        message,
		Timestamp:      time.Now(),
		Actions:        aam.generateActions(alertType, severity),
		Status:         "active",
		Metadata:       make(map[string]interface{}),
	}

	// Add to active alerts
	aam.alerts[alert.ID] = alert
	aam.activeCount++

	// Add to history
	aam.alertHistory = append(aam.alertHistory, alert)

	// Set cooldown
	aam.cooldowns[cooldownKey] = time.Now()

	// Log alert
	aam.logAlert(alert)

	// Execute alert actions
	go aam.executeAlertActions(alert)
}

// GetActiveAlertCount returns the number of active alerts
func (aam *AdvancedAlertManager) GetActiveAlertCount() int {
	aam.mu.RLock()
	defer aam.mu.RUnlock()
	return aam.activeCount
}

// GetActiveAlerts returns all active alerts
func (aam *AdvancedAlertManager) GetActiveAlerts() []*AdvancedAlert {
	aam.mu.RLock()
	defer aam.mu.RUnlock()

	alerts := make([]*AdvancedAlert, 0, len(aam.alerts))
	for _, alert := range aam.alerts {
		if alert.Status == "active" {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// ResolveAlert resolves an alert
func (aam *AdvancedAlertManager) ResolveAlert(alertID string) error {
	aam.mu.Lock()
	defer aam.mu.Unlock()

	alert, exists := aam.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert %s not found", alertID)
	}

	if alert.Status != "active" {
		return fmt.Errorf("alert %s is not active", alertID)
	}

	now := time.Now()
	alert.Status = "resolved"
	alert.ResolvedAt = &now
	aam.activeCount--

	aam.logger.Info("Alert resolved",
		zap.String("alert_id", alertID),
		zap.String("type", alert.Type),
		zap.String("severity", alert.Severity),
		zap.Duration("duration", now.Sub(alert.Timestamp)))

	return nil
}

// AcknowledgeAlert acknowledges an alert
func (aam *AdvancedAlertManager) AcknowledgeAlert(alertID string) error {
	aam.mu.Lock()
	defer aam.mu.Unlock()

	alert, exists := aam.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert %s not found", alertID)
	}

	if alert.Status != "active" {
		return fmt.Errorf("alert %s is not active", alertID)
	}

	now := time.Now()
	alert.Status = "acknowledged"
	alert.AcknowledgedAt = &now

	aam.logger.Info("Alert acknowledged",
		zap.String("alert_id", alertID),
		zap.String("type", alert.Type),
		zap.String("severity", alert.Severity))

	return nil
}

// generateActions generates recommended actions for an alert
func (aam *AdvancedAlertManager) generateActions(alertType, severity string) []string {
	actions := make([]string, 0)

	switch alertType {
	case "accuracy_critical":
		actions = append(actions, "investigate_recent_changes", "check_data_quality", "review_classification_logic", "escalate_to_team")
	case "accuracy_warning":
		actions = append(actions, "monitor_trend", "analyze_patterns", "consider_optimization")
	case "target_achieved":
		actions = append(actions, "document_success", "analyze_success_factors", "consider_optimization")
	case "security_trust_low":
		actions = append(actions, "investigate_data_sources", "check_website_verification", "review_security_policies")
	case "performance_latency_high":
		actions = append(actions, "check_system_resources", "analyze_bottlenecks", "optimize_processing")
	default:
		actions = append(actions, "investigate_issue", "monitor_closely")
	}

	// Add severity-specific actions
	switch severity {
	case "critical":
		actions = append(actions, "immediate_attention_required", "escalate_to_management")
	case "high":
		actions = append(actions, "priority_investigation")
	case "medium":
		actions = append(actions, "schedule_investigation")
	case "low":
		actions = append(actions, "monitor_trend")
	}

	return actions
}

// logAlert logs an alert
func (aam *AdvancedAlertManager) logAlert(alert *AdvancedAlert) {
	logLevel := aam.getLogLevel(alert.Severity)

	logLevel(aam.logger, "Alert created",
		zap.String("alert_id", alert.ID),
		zap.String("type", alert.Type),
		zap.String("severity", alert.Severity),
		zap.String("dimension", alert.DimensionName),
		zap.Float64("current_value", alert.CurrentValue),
		zap.Float64("threshold_value", alert.ThresholdValue),
		zap.String("message", alert.Message),
		zap.Strings("actions", alert.Actions))
}

// getLogLevel returns the appropriate log level for alert severity
func (aam *AdvancedAlertManager) getLogLevel(severity string) func(*zap.Logger, string, ...zap.Field) {
	switch severity {
	case "critical":
		return (*zap.Logger).Error
	case "high":
		return (*zap.Logger).Warn
	case "medium":
		return (*zap.Logger).Info
	case "low":
		return (*zap.Logger).Info
	default:
		return (*zap.Logger).Info
	}
}

// executeAlertActions executes actions for an alert
func (aam *AdvancedAlertManager) executeAlertActions(alert *AdvancedAlert) {
	aam.logger.Info("Executing alert actions",
		zap.String("alert_id", alert.ID),
		zap.Strings("actions", alert.Actions))

	// In a real implementation, these would trigger actual actions
	// For now, we'll just log them
	for _, action := range alert.Actions {
		aam.logger.Debug("Executing alert action",
			zap.String("alert_id", alert.ID),
			zap.String("action", action))

		// Simulate action execution
		time.Sleep(100 * time.Millisecond)
	}
}

// CleanupOldAlerts cleans up old alerts
func (aam *AdvancedAlertManager) CleanupOldAlerts() {
	aam.mu.Lock()
	defer aam.mu.Unlock()

	cutoffTime := time.Now().Add(-24 * time.Hour) // Keep alerts for 24 hours

	// Clean up resolved alerts older than cutoff
	for alertID, alert := range aam.alerts {
		if alert.Status == "resolved" && alert.ResolvedAt != nil && alert.ResolvedAt.Before(cutoffTime) {
			delete(aam.alerts, alertID)
		}
	}

	// Clean up history
	var cleanedHistory []*AdvancedAlert
	for _, alert := range aam.alertHistory {
		if alert.Timestamp.After(cutoffTime) {
			cleanedHistory = append(cleanedHistory, alert)
		}
	}
	aam.alertHistory = cleanedHistory

	aam.logger.Debug("Alert cleanup completed",
		zap.Int("remaining_alerts", len(aam.alerts)),
		zap.Int("remaining_history", len(aam.alertHistory)))
}
