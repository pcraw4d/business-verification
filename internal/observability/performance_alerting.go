package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PerformanceAlertingSystem provides advanced performance alerting and notification capabilities
type PerformanceAlertingSystem struct {
	// Core components
	alertingSystem     *AlertingSystem
	performanceMonitor *PerformanceMonitor
	automatedOptimizer *AutomatedOptimizer

	// Performance-specific alert rules
	performanceRules map[string]*PerformanceAlertRule
	ruleEngine       *PerformanceRuleEngine

	// Notification channels
	notificationChannels map[string]PerformanceNotificationChannel
	notificationQueue    chan *PerformanceAlertNotification

	// Alert management
	activeAlerts    map[string]*PerformanceAlert
	alertHistory    []*PerformanceAlert
	alertEscalation *AlertEscalationManager

	// Performance thresholds
	thresholds *PerformanceThresholds

	// Configuration
	config PerformanceAlertingConfig

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *zap.Logger

	// Control channels
	stopChannel chan struct{}
}

// PerformanceAlertingConfig holds configuration for performance alerting
type PerformanceAlertingConfig struct {
	// Alert evaluation settings
	EvaluationInterval    time.Duration `json:"evaluation_interval"`
	AlertTimeout          time.Duration `json:"alert_timeout"`
	MaxAlertsPerRule      int           `json:"max_alerts_per_rule"`
	AlertHistoryRetention time.Duration `json:"alert_history_retention"`

	// Performance thresholds
	ResponseTimeThresholds struct {
		Warning   time.Duration `json:"warning"`
		Critical  time.Duration `json:"critical"`
		Emergency time.Duration `json:"emergency"`
	} `json:"response_time_thresholds"`

	SuccessRateThresholds struct {
		Warning   float64 `json:"warning"`
		Critical  float64 `json:"critical"`
		Emergency float64 `json:"emergency"`
	} `json:"success_rate_thresholds"`

	ThroughputThresholds struct {
		Warning   float64 `json:"warning"`
		Critical  float64 `json:"critical"`
		Emergency float64 `json:"emergency"`
	} `json:"throughput_thresholds"`

	ResourceUtilizationThresholds struct {
		CPUWarning     float64 `json:"cpu_warning"`
		CPUCritical    float64 `json:"cpu_critical"`
		MemoryWarning  float64 `json:"memory_warning"`
		MemoryCritical float64 `json:"memory_critical"`
		DiskWarning    float64 `json:"disk_warning"`
		DiskCritical   float64 `json:"disk_critical"`
	} `json:"resource_utilization_thresholds"`

	// Notification settings
	NotificationTimeout time.Duration `json:"notification_timeout"`
	RetryAttempts       int           `json:"retry_attempts"`
	RetryDelay          time.Duration `json:"retry_delay"`

	// Escalation settings
	EscalationDelay     time.Duration `json:"escalation_delay"`
	MaxEscalationLevels int           `json:"max_escalation_levels"`

	// Alert grouping
	GroupByLabels      []string      `json:"group_by_labels"`
	GroupWaitTime      time.Duration `json:"group_wait_time"`
	GroupIntervalTime  time.Duration `json:"group_interval_time"`
	RepeatIntervalTime time.Duration `json:"repeat_interval_time"`

	// Integration settings
	EnablePrometheusIntegration bool `json:"enable_prometheus_integration"`
	EnableGrafanaIntegration    bool `json:"enable_grafana_integration"`
	EnableSlackIntegration      bool `json:"enable_slack_integration"`
	EnableEmailIntegration      bool `json:"enable_email_integration"`
	EnablePagerDutyIntegration  bool `json:"enable_pagerduty_integration"`
	EnableWebhookIntegration    bool `json:"enable_webhook_integration"`

	// Advanced features
	EnablePredictiveAlerting bool `json:"enable_predictive_alerting"`
	EnableAnomalyDetection   bool `json:"enable_anomaly_detection"`
	EnableTrendAnalysis      bool `json:"enable_trend_analysis"`
	EnableAutoRemediation    bool `json:"enable_auto_remediation"`
}

// PerformanceAlertRule represents a performance-specific alert rule
type PerformanceAlertRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Enabled     bool   `json:"enabled"`

	// Performance conditions
	MetricType string        `json:"metric_type"` // response_time, success_rate, throughput, cpu, memory, disk
	Condition  string        `json:"condition"`   // threshold, trend, anomaly, prediction
	Duration   time.Duration `json:"duration"`
	Threshold  float64       `json:"threshold"`
	Operator   string        `json:"operator"` // gt, lt, eq, ne, gte, lte

	// Advanced conditions
	TrendWindow       time.Duration `json:"trend_window,omitempty"`
	TrendDirection    string        `json:"trend_direction,omitempty"` // increasing, decreasing
	AnomalyScore      float64       `json:"anomaly_score,omitempty"`
	PredictionHorizon time.Duration `json:"prediction_horizon,omitempty"`

	// Alert metadata
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`

	// Notification settings
	Notifications []string          `json:"notifications"`
	Escalation    *EscalationPolicy `json:"escalation,omitempty"`

	// Remediation settings
	AutoRemediation *AutoRemediationConfig `json:"auto_remediation,omitempty"`

	// Evaluation tracking
	LastEvaluation  time.Time `json:"last_evaluation"`
	EvaluationCount int64     `json:"evaluation_count"`
	FiringCount     int64     `json:"firing_count"`
	ResolvedCount   int64     `json:"resolved_count"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	ID       string `json:"id"`
	RuleID   string `json:"rule_id"`
	RuleName string `json:"rule_name"`
	Severity string `json:"severity"`
	Category string `json:"category"`
	Status   string `json:"status"` // firing, resolved, acknowledged

	// Performance data
	MetricType    string  `json:"metric_type"`
	CurrentValue  float64 `json:"current_value"`
	Threshold     float64 `json:"threshold"`
	BaselineValue float64 `json:"baseline_value,omitempty"`
	TrendValue    float64 `json:"trend_value,omitempty"`
	AnomalyScore  float64 `json:"anomaly_score,omitempty"`

	// Timing
	FiredAt     time.Time  `json:"fired_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	LastUpdated time.Time  `json:"last_updated"`

	// Notification tracking
	NotificationsSent int        `json:"notifications_sent"`
	LastNotification  *time.Time `json:"last_notification,omitempty"`

	// Escalation tracking
	EscalationLevel int        `json:"escalation_level"`
	EscalatedAt     *time.Time `json:"escalated_at,omitempty"`

	// Remediation tracking
	RemediationAttempted bool       `json:"remediation_attempted"`
	RemediationSuccess   bool       `json:"remediation_success,omitempty"`
	RemediatedAt         *time.Time `json:"remediated_at,omitempty"`

	// Context
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// PerformanceAlertNotification represents a performance alert notification
type PerformanceAlertNotification struct {
	Alert      *PerformanceAlert `json:"alert"`
	Channel    string            `json:"channel"`
	Message    string            `json:"message"`
	Priority   string            `json:"priority"`
	RetryCount int               `json:"retry_count"`
	MaxRetries int               `json:"max_retries"`
	CreatedAt  time.Time         `json:"created_at"`
	SentAt     *time.Time        `json:"sent_at,omitempty"`
	Error      *string           `json:"error,omitempty"`
}

// PerformanceNotificationChannel defines a notification channel interface
type PerformanceNotificationChannel interface {
	Name() string
	IsEnabled() bool
	Send(ctx context.Context, notification *PerformanceAlertNotification) error
	GetConfig() map[string]interface{}
}

// PerformanceRuleEngine evaluates performance alert rules
type PerformanceRuleEngine struct {
	rules   map[string]*PerformanceAlertRule
	metrics *PerformanceMetrics
	config  PerformanceAlertingConfig
	logger  *zap.Logger
}

// AlertEscalationManager manages alert escalation
type AlertEscalationManager struct {
	escalationPolicies map[string]*EscalationPolicy
	activeEscalations  map[string]*EscalationEvent
	config             PerformanceAlertingConfig
	logger             *zap.Logger
	mu                 sync.RWMutex
}

// EscalationPolicy defines an escalation policy
type EscalationPolicy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// Escalation levels
	Levels []EscalationLevel `json:"levels"`

	// Settings
	MaxEscalations  int           `json:"max_escalations"`
	EscalationDelay time.Duration `json:"escalation_delay"`
}

// EscalationLevel represents an escalation level
type EscalationLevel struct {
	Level         int           `json:"level"`
	Delay         time.Duration `json:"delay"`
	Notifications []string      `json:"notifications"`
	Recipients    []string      `json:"recipients"`
}

// EscalationEvent represents an escalation event
type EscalationEvent struct {
	ID                string     `json:"id"`
	AlertID           string     `json:"alert_id"`
	PolicyID          string     `json:"policy_id"`
	Level             int        `json:"level"`
	Status            string     `json:"status"` // active, completed, cancelled
	StartedAt         time.Time  `json:"started_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	NotificationsSent int        `json:"notifications_sent"`
}

// AutoRemediationConfig defines auto-remediation configuration
type AutoRemediationConfig struct {
	Enabled           bool          `json:"enabled"`
	Actions           []string      `json:"actions"`
	MaxAttempts       int           `json:"max_attempts"`
	Timeout           time.Duration `json:"timeout"`
	RollbackOnFailure bool          `json:"rollback_on_failure"`
}

// PerformanceThresholds holds performance thresholds
type PerformanceThresholds struct {
	ResponseTime struct {
		Warning   time.Duration `json:"warning"`
		Critical  time.Duration `json:"critical"`
		Emergency time.Duration `json:"emergency"`
	} `json:"response_time"`

	SuccessRate struct {
		Warning   float64 `json:"warning"`
		Critical  float64 `json:"critical"`
		Emergency float64 `json:"emergency"`
	} `json:"success_rate"`

	Throughput struct {
		Warning   float64 `json:"warning"`
		Critical  float64 `json:"critical"`
		Emergency float64 `json:"emergency"`
	} `json:"throughput"`

	ResourceUtilization struct {
		CPU struct {
			Warning  float64 `json:"warning"`
			Critical float64 `json:"critical"`
		} `json:"cpu"`
		Memory struct {
			Warning  float64 `json:"warning"`
			Critical float64 `json:"critical"`
		} `json:"memory"`
		Disk struct {
			Warning  float64 `json:"warning"`
			Critical float64 `json:"critical"`
		} `json:"disk"`
	} `json:"resource_utilization"`
}

// NewPerformanceAlertingSystem creates a new performance alerting system
func NewPerformanceAlertingSystem(
	alertingSystem *AlertingSystem,
	performanceMonitor *PerformanceMonitor,
	automatedOptimizer *AutomatedOptimizer,
	config PerformanceAlertingConfig,
	logger *zap.Logger,
) *PerformanceAlertingSystem {
	// Set default values
	if config.EvaluationInterval == 0 {
		config.EvaluationInterval = 30 * time.Second
	}
	if config.AlertTimeout == 0 {
		config.AlertTimeout = 5 * time.Minute
	}
	if config.NotificationTimeout == 0 {
		config.NotificationTimeout = 30 * time.Second
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Minute
	}
	if config.EscalationDelay == 0 {
		config.EscalationDelay = 15 * time.Minute
	}
	if config.GroupWaitTime == 0 {
		config.GroupWaitTime = 30 * time.Second
	}
	if config.GroupIntervalTime == 0 {
		config.GroupIntervalTime = 5 * time.Minute
	}
	if config.RepeatIntervalTime == 0 {
		config.RepeatIntervalTime = 4 * time.Hour
	}

	pas := &PerformanceAlertingSystem{
		alertingSystem:       alertingSystem,
		performanceMonitor:   performanceMonitor,
		automatedOptimizer:   automatedOptimizer,
		performanceRules:     make(map[string]*PerformanceAlertRule),
		notificationChannels: make(map[string]PerformanceNotificationChannel),
		notificationQueue:    make(chan *PerformanceAlertNotification, 1000),
		activeAlerts:         make(map[string]*PerformanceAlert),
		alertHistory:         make([]*PerformanceAlert, 0),
		config:               config,
		logger:               logger,
		stopChannel:          make(chan struct{}),
	}

	// Initialize components
	pas.ruleEngine = NewPerformanceRuleEngine(pas.performanceRules, config, logger)
	pas.alertEscalation = NewAlertEscalationManager(config, logger)
	pas.thresholds = pas.initializeThresholds()

	// Initialize notification channels
	pas.initializeNotificationChannels()

	// Initialize default performance rules
	pas.initializeDefaultRules()

	return pas
}

// Start starts the performance alerting system
func (pas *PerformanceAlertingSystem) Start(ctx context.Context) error {
	pas.logger.Info("Starting performance alerting system")

	// Start rule evaluation
	go pas.evaluateRules(ctx)

	// Start notification processing
	go pas.processNotifications(ctx)

	// Start escalation management
	go pas.manageEscalations(ctx)

	// Start alert cleanup
	go pas.cleanupAlerts(ctx)

	pas.logger.Info("Performance alerting system started")
	return nil
}

// Stop stops the performance alerting system
func (pas *PerformanceAlertingSystem) Stop() error {
	pas.logger.Info("Stopping performance alerting system")

	close(pas.stopChannel)

	// Close notification queue
	close(pas.notificationQueue)

	pas.logger.Info("Performance alerting system stopped")
	return nil
}

// evaluateRules evaluates performance alert rules
func (pas *PerformanceAlertingSystem) evaluateRules(ctx context.Context) {
	ticker := time.NewTicker(pas.config.EvaluationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pas.stopChannel:
			return
		case <-ticker.C:
			pas.evaluateAllRules(ctx)
		}
	}
}

// evaluateAllRules evaluates all performance alert rules
func (pas *PerformanceAlertingSystem) evaluateAllRules(ctx context.Context) {
	pas.mu.RLock()
	rules := make([]*PerformanceAlertRule, 0, len(pas.performanceRules))
	for _, rule := range pas.performanceRules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	pas.mu.RUnlock()

	for _, rule := range rules {
		pas.evaluateRule(ctx, rule)
	}
}

// evaluateRule evaluates a single performance alert rule
func (pas *PerformanceAlertingSystem) evaluateRule(ctx context.Context, rule *PerformanceAlertRule) {
	// Get current performance metrics
	metrics := pas.performanceMonitor.GetCurrentMetrics()
	if metrics == nil {
		return
	}

	// Evaluate rule condition
	shouldFire, currentValue, baselineValue, trendValue, anomalyScore := pas.ruleEngine.EvaluateRule(rule, metrics)

	// Update rule evaluation tracking
	rule.LastEvaluation = time.Now().UTC()
	rule.EvaluationCount++

	if shouldFire {
		pas.handleAlertFiring(ctx, rule, currentValue, baselineValue, trendValue, anomalyScore)
	} else {
		pas.handleAlertResolution(ctx, rule)
	}
}

// handleAlertFiring handles when an alert rule fires
func (pas *PerformanceAlertingSystem) handleAlertFiring(ctx context.Context, rule *PerformanceAlertRule, currentValue, baselineValue, trendValue, anomalyScore float64) {
	pas.mu.Lock()
	defer pas.mu.Unlock()

	// Check if alert is already firing
	alertID := fmt.Sprintf("%s-%d", rule.ID, time.Now().Unix())
	if existingAlert, exists := pas.activeAlerts[rule.ID]; exists && existingAlert.Status == "firing" {
		// Update existing alert
		existingAlert.CurrentValue = currentValue
		existingAlert.LastUpdated = time.Now().UTC()
		existingAlert.TrendValue = trendValue
		existingAlert.AnomalyScore = anomalyScore
		return
	}

	// Create new alert
	alert := &PerformanceAlert{
		ID:            alertID,
		RuleID:        rule.ID,
		RuleName:      rule.Name,
		Severity:      rule.Severity,
		Category:      rule.Category,
		Status:        "firing",
		MetricType:    rule.MetricType,
		CurrentValue:  currentValue,
		Threshold:     rule.Threshold,
		BaselineValue: baselineValue,
		TrendValue:    trendValue,
		AnomalyScore:  anomalyScore,
		FiredAt:       time.Now().UTC(),
		LastUpdated:   time.Now().UTC(),
		Labels:        rule.Labels,
		Annotations:   rule.Annotations,
	}

	// Add to active alerts
	pas.activeAlerts[rule.ID] = alert

	// Add to history
	pas.alertHistory = append(pas.alertHistory, alert)

	// Update rule firing count
	rule.FiringCount++

	// Send notifications
	pas.sendAlertNotifications(ctx, alert)

	// Start escalation if configured
	if rule.Escalation != nil {
		pas.alertEscalation.StartEscalation(alert, rule.Escalation)
	}

	// Attempt auto-remediation if configured
	if rule.AutoRemediation != nil && rule.AutoRemediation.Enabled {
		go pas.attemptAutoRemediation(ctx, alert, rule.AutoRemediation)
	}

	pas.logger.Info("Performance alert fired",
		zap.String("alert_id", alert.ID),
		zap.String("rule_id", rule.ID),
		zap.String("rule_name", rule.Name),
		zap.String("severity", rule.Severity),
		zap.Float64("current_value", currentValue),
		zap.Float64("threshold", rule.Threshold))
}

// handleAlertResolution handles when an alert rule resolves
func (pas *PerformanceAlertingSystem) handleAlertResolution(ctx context.Context, rule *PerformanceAlertRule) {
	pas.mu.Lock()
	defer pas.mu.Unlock()

	// Check if alert is currently firing
	alert, exists := pas.activeAlerts[rule.ID]
	if !exists || alert.Status != "firing" {
		return
	}

	// Mark as resolved
	now := time.Now().UTC()
	alert.Status = "resolved"
	alert.ResolvedAt = &now
	alert.LastUpdated = now

	// Remove from active alerts
	delete(pas.activeAlerts, rule.ID)

	// Update rule resolved count
	rule.ResolvedCount++

	// Send resolution notifications
	pas.sendAlertNotifications(ctx, alert)

	// Stop escalation
	pas.alertEscalation.StopEscalation(alert.ID)

	pas.logger.Info("Performance alert resolved",
		zap.String("alert_id", alert.ID),
		zap.String("rule_id", rule.ID),
		zap.String("rule_name", rule.Name))
}

// sendAlertNotifications sends notifications for an alert
func (pas *PerformanceAlertingSystem) sendAlertNotifications(ctx context.Context, alert *PerformanceAlert) {
	rule := pas.performanceRules[alert.RuleID]
	if rule == nil {
		return
	}

	for _, notificationType := range rule.Notifications {
		notification := &PerformanceAlertNotification{
			Alert:      alert,
			Channel:    notificationType,
			Message:    pas.formatAlertMessage(alert),
			Priority:   alert.Severity,
			RetryCount: 0,
			MaxRetries: pas.config.RetryAttempts,
			CreatedAt:  time.Now().UTC(),
		}

		// Send to notification queue
		select {
		case pas.notificationQueue <- notification:
		default:
			pas.logger.Warn("Notification queue full, dropping notification",
				zap.String("alert_id", alert.ID),
				zap.String("channel", notificationType))
		}
	}

	alert.NotificationsSent++
	now := time.Now().UTC()
	alert.LastNotification = &now
}

// processNotifications processes notifications from the queue
func (pas *PerformanceAlertingSystem) processNotifications(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-pas.stopChannel:
			return
		case notification := <-pas.notificationQueue:
			pas.sendNotification(ctx, notification)
		}
	}
}

// sendNotification sends a single notification
func (pas *PerformanceAlertingSystem) sendNotification(ctx context.Context, notification *PerformanceAlertNotification) {
	channel, exists := pas.notificationChannels[notification.Channel]
	if !exists || !channel.IsEnabled() {
		pas.logger.Warn("Notification channel not available or disabled",
			zap.String("channel", notification.Channel))
		return
	}

	// Send with timeout
	ctx, cancel := context.WithTimeout(ctx, pas.config.NotificationTimeout)
	defer cancel()

	if err := channel.Send(ctx, notification); err != nil {
		pas.logger.Error("Failed to send notification",
			zap.Error(err),
			zap.String("channel", notification.Channel),
			zap.String("alert_id", notification.Alert.ID))

		// Retry if within retry limit
		if notification.RetryCount < notification.MaxRetries {
			notification.RetryCount++
			time.Sleep(pas.config.RetryDelay)

			select {
			case pas.notificationQueue <- notification:
			default:
				pas.logger.Warn("Notification queue full, dropping retry",
					zap.String("alert_id", notification.Alert.ID),
					zap.String("channel", notification.Channel))
			}
		}
	} else {
		now := time.Now().UTC()
		notification.SentAt = &now
		pas.logger.Info("Notification sent successfully",
			zap.String("channel", notification.Channel),
			zap.String("alert_id", notification.Alert.ID))
	}
}

// formatAlertMessage formats an alert message
func (pas *PerformanceAlertingSystem) formatAlertMessage(alert *PerformanceAlert) string {
	rule := pas.performanceRules[alert.RuleID]
	if rule == nil {
		return fmt.Sprintf("Performance Alert: %s", alert.RuleName)
	}

	message := fmt.Sprintf("🚨 Performance Alert: %s\n", alert.RuleName)
	message += fmt.Sprintf("Severity: %s\n", alert.Severity)
	message += fmt.Sprintf("Category: %s\n", alert.Category)
	message += fmt.Sprintf("Metric: %s\n", alert.MetricType)
	message += fmt.Sprintf("Current Value: %.2f\n", alert.CurrentValue)
	message += fmt.Sprintf("Threshold: %.2f\n", alert.Threshold)

	if alert.BaselineValue > 0 {
		message += fmt.Sprintf("Baseline: %.2f\n", alert.BaselineValue)
	}

	if alert.TrendValue != 0 {
		message += fmt.Sprintf("Trend: %.2f\n", alert.TrendValue)
	}

	if alert.AnomalyScore > 0 {
		message += fmt.Sprintf("Anomaly Score: %.2f\n", alert.AnomalyScore)
	}

	message += fmt.Sprintf("Fired At: %s\n", alert.FiredAt.Format(time.RFC3339))

	return message
}

// GetActiveAlerts returns all active alerts
func (pas *PerformanceAlertingSystem) GetActiveAlerts() []*PerformanceAlert {
	pas.mu.RLock()
	defer pas.mu.RUnlock()

	alerts := make([]*PerformanceAlert, 0, len(pas.activeAlerts))
	for _, alert := range pas.activeAlerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// GetAlertHistory returns alert history
func (pas *PerformanceAlertingSystem) GetAlertHistory() []*PerformanceAlert {
	pas.mu.RLock()
	defer pas.mu.RUnlock()

	history := make([]*PerformanceAlert, len(pas.alertHistory))
	copy(history, pas.alertHistory)
	return history
}

// AddPerformanceRule adds a new performance alert rule
func (pas *PerformanceAlertingSystem) AddPerformanceRule(rule *PerformanceAlertRule) error {
	pas.mu.Lock()
	defer pas.mu.Unlock()

	if rule.ID == "" {
		return fmt.Errorf("performance rule ID is required")
	}

	if _, exists := pas.performanceRules[rule.ID]; exists {
		return fmt.Errorf("performance rule with ID %s already exists", rule.ID)
	}

	pas.performanceRules[rule.ID] = rule
	pas.logger.Info("Performance rule added", zap.String("rule_id", rule.ID))
	return nil
}

// UpdatePerformanceRule updates an existing performance alert rule
func (pas *PerformanceAlertingSystem) UpdatePerformanceRule(ruleID string, rule *PerformanceAlertRule) error {
	pas.mu.Lock()
	defer pas.mu.Unlock()

	if _, exists := pas.performanceRules[ruleID]; !exists {
		return fmt.Errorf("performance rule with ID %s does not exist", ruleID)
	}

	rule.ID = ruleID
	pas.performanceRules[ruleID] = rule
	pas.logger.Info("Performance rule updated", zap.String("rule_id", ruleID))
	return nil
}

// DeletePerformanceRule deletes a performance alert rule
func (pas *PerformanceAlertingSystem) DeletePerformanceRule(ruleID string) error {
	pas.mu.Lock()
	defer pas.mu.Unlock()

	if _, exists := pas.performanceRules[ruleID]; !exists {
		return fmt.Errorf("performance rule with ID %s does not exist", ruleID)
	}

	delete(pas.performanceRules, ruleID)
	pas.logger.Info("Performance rule deleted", zap.String("rule_id", ruleID))
	return nil
}

// initializeThresholds initializes performance thresholds
func (pas *PerformanceAlertingSystem) initializeThresholds() *PerformanceThresholds {
	return &PerformanceThresholds{
		ResponseTime: pas.config.ResponseTimeThresholds,
		SuccessRate:  pas.config.SuccessRateThresholds,
		Throughput:   pas.config.ThroughputThresholds,
		ResourceUtilization: struct {
			CPU struct {
				Warning  float64 `json:"warning"`
				Critical float64 `json:"critical"`
			} `json:"cpu"`
			Memory struct {
				Warning  float64 `json:"warning"`
				Critical float64 `json:"critical"`
			} `json:"memory"`
			Disk struct {
				Warning  float64 `json:"warning"`
				Critical float64 `json:"critical"`
			} `json:"disk"`
		}{
			CPU:    pas.config.ResourceUtilizationThresholds.CPUWarning,
			Memory: pas.config.ResourceUtilizationThresholds.MemoryWarning,
			Disk:   pas.config.ResourceUtilizationThresholds.DiskWarning,
		},
	}
}

// initializeNotificationChannels initializes notification channels
func (pas *PerformanceAlertingSystem) initializeNotificationChannels() {
	// This would be implemented with actual notification channel providers
	// For now, we'll create placeholder channels
}

// initializeDefaultRules initializes default performance alert rules
func (pas *PerformanceAlertingSystem) initializeDefaultRules() {
	// Response time rules
	pas.AddPerformanceRule(&PerformanceAlertRule{
		ID:            "response_time_warning",
		Name:          "High Response Time Warning",
		Description:   "Alert when response time exceeds warning threshold",
		Severity:      "warning",
		Category:      "performance",
		Enabled:       true,
		MetricType:    "response_time",
		Condition:     "threshold",
		Duration:      5 * time.Minute,
		Threshold:     float64(pas.config.ResponseTimeThresholds.Warning.Milliseconds()),
		Operator:      "gt",
		Notifications: []string{"email", "slack"},
	})

	pas.AddPerformanceRule(&PerformanceAlertRule{
		ID:            "response_time_critical",
		Name:          "High Response Time Critical",
		Description:   "Alert when response time exceeds critical threshold",
		Severity:      "critical",
		Category:      "performance",
		Enabled:       true,
		MetricType:    "response_time",
		Condition:     "threshold",
		Duration:      2 * time.Minute,
		Threshold:     float64(pas.config.ResponseTimeThresholds.Critical.Milliseconds()),
		Operator:      "gt",
		Notifications: []string{"email", "slack", "pagerduty"},
	})

	// Success rate rules
	pas.AddPerformanceRule(&PerformanceAlertRule{
		ID:            "success_rate_warning",
		Name:          "Low Success Rate Warning",
		Description:   "Alert when success rate falls below warning threshold",
		Severity:      "warning",
		Category:      "performance",
		Enabled:       true,
		MetricType:    "success_rate",
		Condition:     "threshold",
		Duration:      5 * time.Minute,
		Threshold:     pas.config.SuccessRateThresholds.Warning,
		Operator:      "lt",
		Notifications: []string{"email", "slack"},
	})

	pas.AddPerformanceRule(&PerformanceAlertRule{
		ID:            "success_rate_critical",
		Name:          "Low Success Rate Critical",
		Description:   "Alert when success rate falls below critical threshold",
		Severity:      "critical",
		Category:      "performance",
		Enabled:       true,
		MetricType:    "success_rate",
		Condition:     "threshold",
		Duration:      2 * time.Minute,
		Threshold:     pas.config.SuccessRateThresholds.Critical,
		Operator:      "lt",
		Notifications: []string{"email", "slack", "pagerduty"},
	})

	// Resource utilization rules
	pas.AddPerformanceRule(&PerformanceAlertRule{
		ID:            "cpu_usage_warning",
		Name:          "High CPU Usage Warning",
		Description:   "Alert when CPU usage exceeds warning threshold",
		Severity:      "warning",
		Category:      "infrastructure",
		Enabled:       true,
		MetricType:    "cpu",
		Condition:     "threshold",
		Duration:      5 * time.Minute,
		Threshold:     pas.config.ResourceUtilizationThresholds.CPUWarning,
		Operator:      "gt",
		Notifications: []string{"email", "slack"},
	})

	pas.AddPerformanceRule(&PerformanceAlertRule{
		ID:            "memory_usage_warning",
		Name:          "High Memory Usage Warning",
		Description:   "Alert when memory usage exceeds warning threshold",
		Severity:      "warning",
		Category:      "infrastructure",
		Enabled:       true,
		MetricType:    "memory",
		Condition:     "threshold",
		Duration:      5 * time.Minute,
		Threshold:     pas.config.ResourceUtilizationThresholds.MemoryWarning,
		Operator:      "gt",
		Notifications: []string{"email", "slack"},
	})
}

// manageEscalations manages alert escalations
func (pas *PerformanceAlertingSystem) manageEscalations(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pas.stopChannel:
			return
		case <-ticker.C:
			pas.alertEscalation.ProcessEscalations(ctx)
		}
	}
}

// cleanupAlerts cleans up old alerts
func (pas *PerformanceAlertingSystem) cleanupAlerts(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pas.stopChannel:
			return
		case <-ticker.C:
			pas.cleanupOldAlerts()
		}
	}
}

// cleanupOldAlerts removes old alerts from history
func (pas *PerformanceAlertingSystem) cleanupOldAlerts() {
	pas.mu.Lock()
	defer pas.mu.Unlock()

	cutoff := time.Now().UTC().Add(-pas.config.AlertHistoryRetention)
	newHistory := make([]*PerformanceAlert, 0)

	for _, alert := range pas.alertHistory {
		if alert.FiredAt.After(cutoff) {
			newHistory = append(newHistory, alert)
		}
	}

	pas.alertHistory = newHistory
	pas.logger.Info("Cleaned up old alerts", zap.Int("removed", len(pas.alertHistory)-len(newHistory)))
}

// attemptAutoRemediation attempts automatic remediation for an alert
func (pas *PerformanceAlertingSystem) attemptAutoRemediation(ctx context.Context, alert *PerformanceAlert, config *AutoRemediationConfig) {
	// This would implement actual remediation logic
	// For now, we'll just log the attempt
	pas.logger.Info("Auto-remediation attempted",
		zap.String("alert_id", alert.ID),
		zap.Strings("actions", config.Actions))

	alert.RemediationAttempted = true
	// In a real implementation, this would execute the remediation actions
}
