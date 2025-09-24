package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"kyb-platform/internal/config"
)

// EnhancedAlertingService provides enhanced alerting functionality with configuration support
type EnhancedAlertingService struct {
	logger    *zap.Logger
	config    *config.AlertingConfig
	alerts    map[string]*Alert
	mu        sync.RWMutex
	notifiers []AlertNotifier
	rules     map[string]*config.AlertRule
	stopCh    chan struct{}
	started   bool
}

// NewEnhancedAlertingService creates a new enhanced alerting service
func NewEnhancedAlertingService(logger *zap.Logger, alertingConfig *config.AlertingConfig) *EnhancedAlertingService {
	service := &EnhancedAlertingService{
		logger:    logger,
		config:    alertingConfig,
		alerts:    make(map[string]*Alert),
		notifiers: make([]AlertNotifier, 0),
		rules:     make(map[string]*config.AlertRule),
		stopCh:    make(chan struct{}),
	}

	// Initialize rules
	for i := range alertingConfig.Rules {
		rule := &alertingConfig.Rules[i]
		service.rules[rule.ID] = rule
	}

	// Initialize notifiers
	service.initializeNotifiers()

	return service
}

// Start starts the alerting service
func (a *EnhancedAlertingService) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.started {
		return fmt.Errorf("alerting service already started")
	}

	if !a.config.Enabled {
		a.logger.Info("Alerting service is disabled")
		return nil
	}

	a.logger.Info("Starting enhanced alerting service",
		zap.Duration("check_interval", a.config.CheckInterval),
		zap.Int("rules_count", len(a.rules)),
		zap.Int("notifiers_count", len(a.notifiers)))

	// Start the alert checking goroutine
	go a.alertCheckingLoop()

	a.started = true
	return nil
}

// Stop stops the alerting service
func (a *EnhancedAlertingService) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.started {
		return nil
	}

	a.logger.Info("Stopping enhanced alerting service")

	close(a.stopCh)
	a.started = false

	return nil
}

// CreateAlert creates a new alert
func (a *EnhancedAlertingService) CreateAlert(alert *Alert) error {
	if alert.ID == "" {
		alert.ID = fmt.Sprintf("alert-%d", time.Now().UnixNano())
	}

	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}

	if alert.Status == "" {
		alert.Status = string(AlertStatusActive)
	}

	// Check if alert should be suppressed
	if a.shouldSuppressAlert(alert) {
		a.logger.Debug("Alert suppressed",
			zap.String("alert_id", alert.ID),
			zap.String("title", alert.Title))
		return nil
	}

	a.mu.Lock()
	a.alerts[alert.ID] = alert
	a.mu.Unlock()

	a.logger.Info("Alert created",
		zap.String("alert_id", alert.ID),
		zap.String("title", alert.Title),
		zap.String("severity", alert.Severity))

	// Send notifications
	go a.sendNotifications(alert)

	return nil
}

// ResolveAlert resolves an alert
func (a *EnhancedAlertingService) ResolveAlert(alertID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	alert, exists := a.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	if alert.Status == string(AlertStatusResolved) {
		return fmt.Errorf("alert already resolved: %s", alertID)
	}

	now := time.Now()
	alert.Status = string(AlertStatusResolved)
	alert.ResolvedAt = &now

	a.logger.Info("Alert resolved",
		zap.String("alert_id", alertID),
		zap.String("title", alert.Title),
		zap.Duration("duration", now.Sub(alert.Timestamp)))

	return nil
}

// GetActiveAlerts returns all active alerts
func (a *EnhancedAlertingService) GetActiveAlerts() []*Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var activeAlerts []*Alert
	for _, alert := range a.alerts {
		if alert.Status == string(AlertStatusActive) {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// GetAlertsBySeverity returns alerts filtered by severity
func (a *EnhancedAlertingService) GetAlertsBySeverity(severity AlertSeverity) []*Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var filteredAlerts []*Alert
	for _, alert := range a.alerts {
		if alert.Severity == string(severity) {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	return filteredAlerts
}

// GetAlertHistory returns alert history
func (a *EnhancedAlertingService) GetAlertHistory(limit int) []*Alert {
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
func (a *EnhancedAlertingService) CheckAlertRules(ctx context.Context, metrics map[string]float64) error {
	for ruleID, rule := range a.rules {
		if !rule.Enabled {
			continue
		}

		if a.evaluateRule(rule, metrics) {
			alert := &Alert{
				Title:       rule.Name,
				Description: rule.Description,
				Severity:    string(rule.Severity),
				Source:      "alerting_service",
				Labels:      rule.Labels,
				Metadata:    rule.Metadata,
			}

			if err := a.CreateAlert(alert); err != nil {
				a.logger.Error("Failed to create alert from rule",
					zap.String("rule_id", ruleID),
					zap.Error(err))
			}
		}
	}

	return nil
}

// UpdateConfiguration updates the alerting configuration
func (a *EnhancedAlertingService) UpdateConfiguration(newConfig *config.AlertingConfig) error {
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.config = newConfig

	// Update rules
	a.rules = make(map[string]*config.AlertRule)
	for i := range newConfig.Rules {
		rule := &newConfig.Rules[i]
		a.rules[rule.ID] = rule
	}

	// Reinitialize notifiers
	a.initializeNotifiers()

	a.logger.Info("Alerting configuration updated",
		zap.Int("rules_count", len(a.rules)),
		zap.Int("notifiers_count", len(a.notifiers)))

	return nil
}

// GetConfiguration returns the current alerting configuration
func (a *EnhancedAlertingService) GetConfiguration() *config.AlertingConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config
}

// alertCheckingLoop runs the main alert checking loop
func (a *EnhancedAlertingService) alertCheckingLoop() {
	ticker := time.NewTicker(a.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// In a real implementation, you would collect metrics here
			// and call CheckAlertRules with the collected metrics
			a.logger.Debug("Alert checking tick")
		case <-a.stopCh:
			a.logger.Info("Alert checking loop stopped")
			return
		}
	}
}

// initializeNotifiers initializes alert notifiers based on configuration
func (a *EnhancedAlertingService) initializeNotifiers() {
	a.notifiers = make([]AlertNotifier, 0)

	for _, notifierConfig := range a.config.Notifiers {
		if !notifierConfig.Enabled {
			continue
		}

		var notifier AlertNotifier
		var err error

		switch notifierConfig.Type {
		case "email":
			notifier, err = a.createEmailNotifier(notifierConfig)
		case "slack":
			notifier, err = a.createSlackNotifier(notifierConfig)
		case "webhook":
			notifier, err = a.createWebhookNotifier(notifierConfig)
		case "mock":
			notifier = NewMockAlertNotifier(a.logger)
		default:
			a.logger.Warn("Unknown notifier type",
				zap.String("type", notifierConfig.Type),
				zap.String("name", notifierConfig.Name))
			continue
		}

		if err != nil {
			a.logger.Error("Failed to create notifier",
				zap.String("type", notifierConfig.Type),
				zap.String("name", notifierConfig.Name),
				zap.Error(err))
			continue
		}

		a.notifiers = append(a.notifiers, notifier)
		a.logger.Info("Notifier initialized",
			zap.String("type", notifierConfig.Type),
			zap.String("name", notifierConfig.Name))
	}
}

// createEmailNotifier creates an email notifier
func (a *EnhancedAlertingService) createEmailNotifier(config config.NotifierConfig) (AlertNotifier, error) {
	// In a real implementation, you would create an actual email notifier
	// For now, return a mock notifier
	return NewMockAlertNotifier(a.logger), nil
}

// createSlackNotifier creates a Slack notifier
func (a *EnhancedAlertingService) createSlackNotifier(config config.NotifierConfig) (AlertNotifier, error) {
	// In a real implementation, you would create an actual Slack notifier
	// For now, return a mock notifier
	return NewMockAlertNotifier(a.logger), nil
}

// createWebhookNotifier creates a webhook notifier
func (a *EnhancedAlertingService) createWebhookNotifier(config config.NotifierConfig) (AlertNotifier, error) {
	// In a real implementation, you would create an actual webhook notifier
	// For now, return a mock notifier
	return NewMockAlertNotifier(a.logger), nil
}

// shouldSuppressAlert checks if an alert should be suppressed based on suppression rules
func (a *EnhancedAlertingService) shouldSuppressAlert(alert *Alert) bool {
	for _, suppressionRule := range a.config.SuppressionRules {
		if !suppressionRule.Enabled {
			continue
		}

		if a.matchesSuppressionRule(alert, suppressionRule) {
			return true
		}
	}

	return false
}

// matchesSuppressionRule checks if an alert matches a suppression rule
func (a *EnhancedAlertingService) matchesSuppressionRule(alert *Alert, rule config.SuppressionRule) bool {
	// Simple implementation - in real implementation, you would have more sophisticated matching
	for key, value := range rule.Conditions {
		if alert.Labels[key] != value {
			return false
		}
	}

	return true
}

// evaluateRule evaluates an alert rule against current metrics
func (a *EnhancedAlertingService) evaluateRule(rule *config.AlertRule, metrics map[string]float64) bool {
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

// sendNotifications sends notifications for an alert
func (a *EnhancedAlertingService) sendNotifications(alert *Alert) {
	for _, notifier := range a.notifiers {
		if err := notifier.Notify(alert); err != nil {
			a.logger.Error("Failed to send notification",
				zap.String("notifier", notifier.Name()),
				zap.String("alert_id", alert.ID),
				zap.Error(err))
		}
	}
}
