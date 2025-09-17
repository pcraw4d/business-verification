package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AdvancedAlertingIntegration provides integration between classification monitoring and alerting
type AdvancedAlertingIntegration struct {
	classificationAlertManager *ClassificationAlertManager
	baseAlertManager           *AlertManager
	notificationConfig         *NotificationConfig
	templateManager            *NotificationTemplateManager
	rateLimiter                *RateLimiter
	suppressor                 *NotificationSuppressor
	logger                     *zap.Logger
	mu                         sync.RWMutex
	started                    bool
	ctx                        context.Context
	cancel                     context.CancelFunc
}

// AdvancedAlertingConfig holds configuration for advanced alerting integration
type AdvancedAlertingConfig struct {
	// Alerting settings
	Enabled             bool          `json:"enabled"`
	EvaluationInterval  time.Duration `json:"evaluation_interval"`
	NotificationTimeout time.Duration `json:"notification_timeout"`
	MaxRetries          int           `json:"max_retries"`
	RetryInterval       time.Duration `json:"retry_interval"`

	// Rate limiting
	RateLimitPerMinute  int           `json:"rate_limit_per_minute"`
	SuppressionDuration time.Duration `json:"suppression_duration"`

	// Integration settings
	IntegrateWithMLMonitoring       bool `json:"integrate_with_ml_monitoring"`
	IntegrateWithEnsembleMonitoring bool `json:"integrate_with_ensemble_monitoring"`
	IntegrateWithSecurityMonitoring bool `json:"integrate_with_security_monitoring"`
	IntegrateWithAccuracyTracking   bool `json:"integrate_with_accuracy_tracking"`

	// Notification settings
	NotificationConfig *NotificationConfig `json:"notification_config"`

	// Environment settings
	Environment string `json:"environment"`
	ServiceName string `json:"service_name"`
	Version     string `json:"version"`
}

// NewAdvancedAlertingIntegration creates a new advanced alerting integration
func NewAdvancedAlertingIntegration(
	baseAlertManager *AlertManager,
	logger *zap.Logger,
	config *AdvancedAlertingConfig,
) *AdvancedAlertingIntegration {
	ctx, cancel := context.WithCancel(context.Background())

	// Create classification alert manager
	classificationConfig := &ClassificationAlertConfig{
		Enabled:              config.Enabled,
		EvaluationInterval:   config.EvaluationInterval,
		NotificationTimeout:  config.NotificationTimeout,
		MaxRetries:           config.MaxRetries,
		RetryInterval:        config.RetryInterval,
		SuppressionEnabled:   true,
		SuppressionDuration:  config.SuppressionDuration,
		DeduplicationEnabled: true,
		EscalationEnabled:    true,
		Environment:          config.Environment,
		ServiceName:          config.ServiceName,
		Version:              config.Version,
	}

	classificationAlertManager := NewClassificationAlertManager(baseAlertManager, logger, classificationConfig)

	// Create template manager
	templateManager := NewNotificationTemplateManager()
	templateManager.LoadDefaultTemplates()

	// Create rate limiter
	rateLimiter := NewRateLimiter(config.RateLimitPerMinute, time.Minute)

	// Create suppressor
	suppressor := NewNotificationSuppressor(config.SuppressionDuration)

	return &AdvancedAlertingIntegration{
		classificationAlertManager: classificationAlertManager,
		baseAlertManager:           baseAlertManager,
		notificationConfig:         config.NotificationConfig,
		templateManager:            templateManager,
		rateLimiter:                rateLimiter,
		suppressor:                 suppressor,
		logger:                     logger,
		ctx:                        ctx,
		cancel:                     cancel,
	}
}

// Start starts the advanced alerting integration
func (aai *AdvancedAlertingIntegration) Start() error {
	aai.mu.Lock()
	defer aai.mu.Unlock()

	if aai.started {
		return fmt.Errorf("advanced alerting integration already started")
	}

	aai.logger.Info("Starting advanced alerting integration",
		zap.String("service_name", aai.classificationAlertManager.baseAlertManager.config.ServiceName),
		zap.String("version", aai.classificationAlertManager.baseAlertManager.config.Version),
		zap.String("environment", aai.classificationAlertManager.baseAlertManager.config.Environment),
	)

	// Start base alert manager if not already started
	if err := aai.baseAlertManager.Start(); err != nil {
		return fmt.Errorf("failed to start base alert manager: %w", err)
	}

	// Start classification alert manager
	if err := aai.classificationAlertManager.Start(); err != nil {
		return fmt.Errorf("failed to start classification alert manager: %w", err)
	}

	// Configure notification channels
	if err := aai.configureNotificationChannels(); err != nil {
		return fmt.Errorf("failed to configure notification channels: %w", err)
	}

	// Start background cleanup
	go aai.startBackgroundCleanup()

	aai.started = true
	aai.logger.Info("Advanced alerting integration started successfully")
	return nil
}

// Stop stops the advanced alerting integration
func (aai *AdvancedAlertingIntegration) Stop() error {
	aai.mu.Lock()
	defer aai.mu.Unlock()

	if !aai.started {
		return fmt.Errorf("advanced alerting integration not started")
	}

	aai.logger.Info("Stopping advanced alerting integration")

	// Stop classification alert manager
	if err := aai.classificationAlertManager.Stop(); err != nil {
		aai.logger.Error("Failed to stop classification alert manager",
			zap.Error(err),
		)
	}

	// Stop base alert manager
	if err := aai.baseAlertManager.Stop(); err != nil {
		aai.logger.Error("Failed to stop base alert manager",
			zap.Error(err),
		)
	}

	aai.cancel()
	aai.started = false

	aai.logger.Info("Advanced alerting integration stopped successfully")
	return nil
}

// configureNotificationChannels configures notification channels
func (aai *AdvancedAlertingIntegration) configureNotificationChannels() error {
	if aai.notificationConfig == nil {
		aai.notificationConfig = DefaultNotificationConfig()
	}

	// Create notification channel factory
	factory := NewNotificationChannelFactory(aai.notificationConfig, &Logger{})

	// Create channels
	channels := factory.CreateNotificationChannels()

	// Add channels to base alert manager
	for name, channel := range channels {
		aai.baseAlertManager.AddNotificationChannel(name, channel)
		aai.logger.Info("Notification channel configured",
			zap.String("channel_name", name),
			zap.String("channel_type", channel.Type()),
		)
	}

	return nil
}

// startBackgroundCleanup starts background cleanup processes
func (aai *AdvancedAlertingIntegration) startBackgroundCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-aai.ctx.Done():
			return
		case <-ticker.C:
			// Cleanup expired suppressions
			aai.suppressor.CleanupExpiredSuppressions()
		}
	}
}

// TriggerAccuracyAlert triggers an accuracy-related alert
func (aai *AdvancedAlertingIntegration) TriggerAccuracyAlert(
	alertType ClassificationMetricType,
	value float64,
	threshold float64,
	labels map[string]string,
) error {
	// Check rate limiting
	if !aai.rateLimiter.Allow("accuracy_alerts") {
		aai.logger.Warn("Accuracy alert rate limited",
			zap.String("alert_type", string(alertType)),
			zap.Float64("value", value),
		)
		return nil
	}

	// Check suppression
	alertKey := fmt.Sprintf("accuracy_%s_%.2f", alertType, value)
	if aai.suppressor.IsSuppressed(alertKey) {
		aai.logger.Debug("Accuracy alert suppressed",
			zap.String("alert_key", alertKey),
		)
		return nil
	}

	// Create alert rule if it doesn't exist
	ruleID := fmt.Sprintf("accuracy_%s", alertType)
	rule, err := aai.classificationAlertManager.GetClassificationAlertRule(ruleID)
	if err != nil {
		// Create new rule
		rule = &ClassificationAlertRule{
			ID:          ruleID,
			Name:        fmt.Sprintf("Accuracy Alert - %s", alertType),
			Description: fmt.Sprintf("Alert for %s accuracy issues", alertType),
			Category:    AlertCategoryAccuracy,
			MetricType:  alertType,
			Query:       fmt.Sprintf("kyb_%s", alertType),
			Condition:   "lt",
			Threshold:   threshold,
			Severity:    AlertSeverityWarning,
			Duration:    2 * time.Minute,
			Labels:      labels,
			Annotations: map[string]string{
				"summary":     fmt.Sprintf("Accuracy alert for %s", alertType),
				"description": fmt.Sprintf("%s accuracy is %.2f (threshold: %.2f)", alertType, value, threshold),
			},
			NotificationChannels: []string{"slack"},
			EscalationPolicy:     "default",
			Enabled:              true,
		}

		if err := aai.classificationAlertManager.AddClassificationAlertRule(rule); err != nil {
			return fmt.Errorf("failed to add accuracy alert rule: %w", err)
		}
	}

	// Trigger the alert
	aai.classificationAlertManager.triggerClassificationAlert(rule, value)

	// Suppress for a short time to avoid spam
	aai.suppressor.Suppress(alertKey)

	return nil
}

// TriggerMLModelAlert triggers an ML model-related alert
func (aai *AdvancedAlertingIntegration) TriggerMLModelAlert(
	alertType ClassificationMetricType,
	value float64,
	threshold float64,
	modelName string,
	labels map[string]string,
) error {
	// Check rate limiting
	if !aai.rateLimiter.Allow("ml_model_alerts") {
		aai.logger.Warn("ML model alert rate limited",
			zap.String("alert_type", string(alertType)),
			zap.String("model_name", modelName),
			zap.Float64("value", value),
		)
		return nil
	}

	// Check suppression
	alertKey := fmt.Sprintf("ml_model_%s_%s_%.2f", modelName, alertType, value)
	if aai.suppressor.IsSuppressed(alertKey) {
		aai.logger.Debug("ML model alert suppressed",
			zap.String("alert_key", alertKey),
		)
		return nil
	}

	// Create alert rule if it doesn't exist
	ruleID := fmt.Sprintf("ml_model_%s_%s", modelName, alertType)
	rule, err := aai.classificationAlertManager.GetClassificationAlertRule(ruleID)
	if err != nil {
		// Create new rule
		rule = &ClassificationAlertRule{
			ID:          ruleID,
			Name:        fmt.Sprintf("ML Model Alert - %s (%s)", alertType, modelName),
			Description: fmt.Sprintf("Alert for %s model %s issues", modelName, alertType),
			Category:    AlertCategoryMLModel,
			MetricType:  alertType,
			Query:       fmt.Sprintf("kyb_%s_%s", modelName, alertType),
			Condition:   "gt",
			Threshold:   threshold,
			Severity:    AlertSeverityCritical,
			Duration:    1 * time.Minute,
			Labels:      labels,
			Annotations: map[string]string{
				"summary":     fmt.Sprintf("ML model alert for %s (%s)", modelName, alertType),
				"description": fmt.Sprintf("%s model %s is %.2f (threshold: %.2f)", modelName, alertType, value, threshold),
			},
			NotificationChannels: []string{"email", "slack", "webhook"},
			EscalationPolicy:     "critical",
			Enabled:              true,
		}

		if err := aai.classificationAlertManager.AddClassificationAlertRule(rule); err != nil {
			return fmt.Errorf("failed to add ML model alert rule: %w", err)
		}
	}

	// Trigger the alert
	aai.classificationAlertManager.triggerClassificationAlert(rule, value)

	// Suppress for a short time to avoid spam
	aai.suppressor.Suppress(alertKey)

	return nil
}

// TriggerEnsembleAlert triggers an ensemble-related alert
func (aai *AdvancedAlertingIntegration) TriggerEnsembleAlert(
	alertType ClassificationMetricType,
	value float64,
	threshold float64,
	labels map[string]string,
) error {
	// Check rate limiting
	if !aai.rateLimiter.Allow("ensemble_alerts") {
		aai.logger.Warn("Ensemble alert rate limited",
			zap.String("alert_type", string(alertType)),
			zap.Float64("value", value),
		)
		return nil
	}

	// Check suppression
	alertKey := fmt.Sprintf("ensemble_%s_%.2f", alertType, value)
	if aai.suppressor.IsSuppressed(alertKey) {
		aai.logger.Debug("Ensemble alert suppressed",
			zap.String("alert_key", alertKey),
		)
		return nil
	}

	// Create alert rule if it doesn't exist
	ruleID := fmt.Sprintf("ensemble_%s", alertType)
	rule, err := aai.classificationAlertManager.GetClassificationAlertRule(ruleID)
	if err != nil {
		// Create new rule
		rule = &ClassificationAlertRule{
			ID:          ruleID,
			Name:        fmt.Sprintf("Ensemble Alert - %s", alertType),
			Description: fmt.Sprintf("Alert for ensemble %s issues", alertType),
			Category:    AlertCategoryEnsemble,
			MetricType:  alertType,
			Query:       fmt.Sprintf("kyb_%s", alertType),
			Condition:   "gt",
			Threshold:   threshold,
			Severity:    AlertSeverityWarning,
			Duration:    5 * time.Minute,
			Labels:      labels,
			Annotations: map[string]string{
				"summary":     fmt.Sprintf("Ensemble alert for %s", alertType),
				"description": fmt.Sprintf("Ensemble %s is %.2f (threshold: %.2f)", alertType, value, threshold),
			},
			NotificationChannels: []string{"slack"},
			EscalationPolicy:     "default",
			Enabled:              true,
		}

		if err := aai.classificationAlertManager.AddClassificationAlertRule(rule); err != nil {
			return fmt.Errorf("failed to add ensemble alert rule: %w", err)
		}
	}

	// Trigger the alert
	aai.classificationAlertManager.triggerClassificationAlert(rule, value)

	// Suppress for a short time to avoid spam
	aai.suppressor.Suppress(alertKey)

	return nil
}

// TriggerSecurityAlert triggers a security-related alert
func (aai *AdvancedAlertingIntegration) TriggerSecurityAlert(
	alertType ClassificationMetricType,
	value float64,
	threshold float64,
	labels map[string]string,
) error {
	// Check rate limiting
	if !aai.rateLimiter.Allow("security_alerts") {
		aai.logger.Warn("Security alert rate limited",
			zap.String("alert_type", string(alertType)),
			zap.Float64("value", value),
		)
		return nil
	}

	// Check suppression
	alertKey := fmt.Sprintf("security_%s_%.2f", alertType, value)
	if aai.suppressor.IsSuppressed(alertKey) {
		aai.logger.Debug("Security alert suppressed",
			zap.String("alert_key", alertKey),
		)
		return nil
	}

	// Create alert rule if it doesn't exist
	ruleID := fmt.Sprintf("security_%s", alertType)
	rule, err := aai.classificationAlertManager.GetClassificationAlertRule(ruleID)
	if err != nil {
		// Create new rule
		rule = &ClassificationAlertRule{
			ID:          ruleID,
			Name:        fmt.Sprintf("Security Alert - %s", alertType),
			Description: fmt.Sprintf("Alert for security %s issues", alertType),
			Category:    AlertCategorySecurity,
			MetricType:  alertType,
			Query:       fmt.Sprintf("kyb_%s", alertType),
			Condition:   "gt",
			Threshold:   threshold,
			Severity:    AlertSeverityCritical,
			Duration:    0,
			Labels:      labels,
			Annotations: map[string]string{
				"summary":     fmt.Sprintf("Security alert for %s", alertType),
				"description": fmt.Sprintf("Security %s is %.2f (threshold: %.2f)", alertType, value, threshold),
			},
			NotificationChannels: []string{"email", "slack", "webhook"},
			EscalationPolicy:     "critical",
			Enabled:              true,
		}

		if err := aai.classificationAlertManager.AddClassificationAlertRule(rule); err != nil {
			return fmt.Errorf("failed to add security alert rule: %w", err)
		}
	}

	// Trigger the alert
	aai.classificationAlertManager.triggerClassificationAlert(rule, value)

	// Suppress for a short time to avoid spam
	aai.suppressor.Suppress(alertKey)

	return nil
}

// GetAlertSummary returns a comprehensive alert summary
func (aai *AdvancedAlertingIntegration) GetAlertSummary() *AdvancedAlertSummary {
	baseAlerts := aai.baseAlertManager.GetActiveAlerts()
	classificationSummary := aai.classificationAlertManager.GetClassificationAlertSummary()

	summary := &AdvancedAlertSummary{
		Timestamp:             time.Now(),
		TotalAlerts:           len(baseAlerts),
		CriticalAlerts:        classificationSummary.CriticalAlerts,
		WarningAlerts:         classificationSummary.WarningAlerts,
		InfoAlerts:            classificationSummary.InfoAlerts,
		ClassificationSummary: classificationSummary,
		AlertsByCategory:      classificationSummary.AlertsByCategory,
		AlertsBySeverity:      classificationSummary.AlertsBySeverity,
		AlertsByMetricType:    classificationSummary.AlertsByMetricType,
		RateLimitStatus:       aai.getRateLimitStatus(),
		SuppressionStatus:     aai.getSuppressionStatus(),
	}

	return summary
}

// getRateLimitStatus returns current rate limiting status
func (aai *AdvancedAlertingIntegration) getRateLimitStatus() map[string]interface{} {
	return map[string]interface{}{
		"accuracy_alerts": aai.rateLimiter.Allow("accuracy_alerts"),
		"ml_model_alerts": aai.rateLimiter.Allow("ml_model_alerts"),
		"ensemble_alerts": aai.rateLimiter.Allow("ensemble_alerts"),
		"security_alerts": aai.rateLimiter.Allow("security_alerts"),
	}
}

// getSuppressionStatus returns current suppression status
func (aai *AdvancedAlertingIntegration) getSuppressionStatus() map[string]bool {
	aai.suppressor.mu.RLock()
	defer aai.suppressor.mu.RUnlock()

	status := make(map[string]bool)
	for key := range aai.suppressor.suppressed {
		status[key] = true
	}
	return status
}

// AdvancedAlertSummary represents a comprehensive alert summary
type AdvancedAlertSummary struct {
	Timestamp             time.Time                        `json:"timestamp"`
	TotalAlerts           int                              `json:"total_alerts"`
	CriticalAlerts        int                              `json:"critical_alerts"`
	WarningAlerts         int                              `json:"warning_alerts"`
	InfoAlerts            int                              `json:"info_alerts"`
	ClassificationSummary *ClassificationAlertSummary      `json:"classification_summary"`
	AlertsByCategory      map[AlertCategory]int            `json:"alerts_by_category"`
	AlertsBySeverity      map[AlertSeverity]int            `json:"alerts_by_severity"`
	AlertsByMetricType    map[ClassificationMetricType]int `json:"alerts_by_metric_type"`
	RateLimitStatus       map[string]interface{}           `json:"rate_limit_status"`
	SuppressionStatus     map[string]bool                  `json:"suppression_status"`
}

// DefaultAdvancedAlertingConfig returns default configuration for advanced alerting
func DefaultAdvancedAlertingConfig() *AdvancedAlertingConfig {
	return &AdvancedAlertingConfig{
		Enabled:                         true,
		EvaluationInterval:              30 * time.Second,
		NotificationTimeout:             10 * time.Second,
		MaxRetries:                      3,
		RetryInterval:                   30 * time.Second,
		RateLimitPerMinute:              60,
		SuppressionDuration:             5 * time.Minute,
		IntegrateWithMLMonitoring:       true,
		IntegrateWithEnsembleMonitoring: true,
		IntegrateWithSecurityMonitoring: true,
		IntegrateWithAccuracyTracking:   true,
		NotificationConfig:              DefaultNotificationConfig(),
		Environment:                     "production",
		ServiceName:                     "kyb-platform",
		Version:                         "1.0.0",
	}
}
