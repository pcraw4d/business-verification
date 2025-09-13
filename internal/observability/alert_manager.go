package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AlertManager provides comprehensive alerting functionality
type AlertManager struct {
	logger               *Logger
	alerts               map[string]*Alert
	alertRules           map[string]*AlertRule
	notificationChannels map[string]NotificationChannel
	escalationPolicies   map[string]*EscalationPolicy
	config               *AlertConfig
	mu                   sync.RWMutex
	ctx                  context.Context
	cancel               context.CancelFunc
	started              bool
}

// Alert represents an active alert
type Alert struct {
	ID                string                 `json:"id"`
	RuleID            string                 `json:"rule_id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Severity          AlertSeverity          `json:"severity"`
	Status            AlertStatus            `json:"status"`
	State             AlertState             `json:"state"`
	Labels            map[string]string      `json:"labels"`
	Annotations       map[string]string      `json:"annotations"`
	Value             float64                `json:"value"`
	Threshold         float64                `json:"threshold"`
	Condition         string                 `json:"condition"`
	StartedAt         time.Time              `json:"started_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	ResolvedAt        *time.Time             `json:"resolved_at,omitempty"`
	LastNotifiedAt    *time.Time             `json:"last_notified_at,omitempty"`
	NotificationCount int                    `json:"notification_count"`
	EscalationLevel   int                    `json:"escalation_level"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// AlertRule represents an alert rule
type AlertRule struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	Query                string            `json:"query"`
	Condition            string            `json:"condition"`
	Threshold            float64           `json:"threshold"`
	Severity             AlertSeverity     `json:"severity"`
	Duration             time.Duration     `json:"duration"`
	Labels               map[string]string `json:"labels"`
	Annotations          map[string]string `json:"annotations"`
	NotificationChannels []string          `json:"notification_channels"`
	EscalationPolicy     string            `json:"escalation_policy"`
	Enabled              bool              `json:"enabled"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityDebug    AlertSeverity = "debug"
)

// AlertStatus represents alert status
type AlertStatus string

const (
	AlertStatusActive     AlertStatus = "active"
	AlertStatusResolved   AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
)

// AlertState represents alert state
type AlertState string

const (
	AlertStateFiring     AlertState = "firing"
	AlertStatePending    AlertState = "pending"
	AlertStateResolved   AlertState = "resolved"
	AlertStateSuppressed AlertState = "suppressed"
)

// EscalationPolicy represents an escalation policy
type EscalationPolicy struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Levels      []*EscalationLevel `json:"levels"`
	Enabled     bool               `json:"enabled"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// EscalationLevel represents an escalation level
type EscalationLevel struct {
	Level                int                    `json:"level"`
	Duration             time.Duration          `json:"duration"`
	NotificationChannels []string               `json:"notification_channels"`
	Actions              []string               `json:"actions"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// AlertConfig holds configuration for alert manager
type AlertConfig struct {
	Enabled              bool
	EvaluationInterval   time.Duration
	NotificationTimeout  time.Duration
	MaxRetries           int
	RetryInterval        time.Duration
	SuppressionEnabled   bool
	SuppressionDuration  time.Duration
	DeduplicationEnabled bool
	EscalationEnabled    bool
	Environment          string
	ServiceName          string
	Version              string
}

// NotificationChannel interface for sending notifications
type NotificationChannel interface {
	Send(alert *Alert) error
	Name() string
	Type() string
	Enabled() bool
}

// EmailNotificationChannel sends email notifications
type EmailNotificationChannel struct {
	logger  *Logger
	config  *EmailConfig
	enabled bool
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
	To       []string
	Subject  string
	Template string
}

// NewEmailNotificationChannel creates a new email notification channel
func NewEmailNotificationChannel(logger *Logger, config *EmailConfig) *EmailNotificationChannel {
	return &EmailNotificationChannel{
		logger:  logger,
		config:  config,
		enabled: true,
	}
}

// Send sends an email notification
func (enc *EmailNotificationChannel) Send(alert *Alert) error {
	if !enc.enabled {
		return fmt.Errorf("email notification channel is disabled")
	}

	enc.logger.Info("Sending email notification", map[string]interface{}{
		"alert_id":   alert.ID,
		"alert_name": alert.Name,
		"severity":   alert.Severity,
		"to":         enc.config.To,
	})

	// In a real implementation, this would send an actual email
	return nil
}

// Name returns the channel name
func (enc *EmailNotificationChannel) Name() string {
	return "email"
}

// Type returns the channel type
func (enc *EmailNotificationChannel) Type() string {
	return "email"
}

// Enabled returns whether the channel is enabled
func (enc *EmailNotificationChannel) Enabled() bool {
	return enc.enabled
}

// SlackNotificationChannel sends Slack notifications
type SlackNotificationChannel struct {
	logger  *Logger
	config  *SlackConfig
	enabled bool
}

// SlackConfig holds Slack configuration
type SlackConfig struct {
	WebhookURL string
	Channel    string
	Username   string
	IconEmoji  string
	Template   string
}

// NewSlackNotificationChannel creates a new Slack notification channel
func NewSlackNotificationChannel(logger *Logger, config *SlackConfig) *SlackNotificationChannel {
	return &SlackNotificationChannel{
		logger:  logger,
		config:  config,
		enabled: true,
	}
}

// Send sends a Slack notification
func (snc *SlackNotificationChannel) Send(alert *Alert) error {
	if !snc.enabled {
		return fmt.Errorf("slack notification channel is disabled")
	}

	snc.logger.Info("Sending Slack notification", map[string]interface{}{
		"alert_id":   alert.ID,
		"alert_name": alert.Name,
		"severity":   alert.Severity,
		"channel":    snc.config.Channel,
	})

	// In a real implementation, this would send an actual Slack message
	return nil
}

// Name returns the channel name
func (snc *SlackNotificationChannel) Name() string {
	return "slack"
}

// Type returns the channel type
func (snc *SlackNotificationChannel) Type() string {
	return "slack"
}

// Enabled returns whether the channel is enabled
func (snc *SlackNotificationChannel) Enabled() bool {
	return snc.enabled
}

// WebhookNotificationChannel sends webhook notifications
type WebhookNotificationChannel struct {
	logger  *Logger
	config  *WebhookConfig
	enabled bool
}

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	URL      string
	Method   string
	Headers  map[string]string
	Timeout  time.Duration
	Template string
}

// NewWebhookNotificationChannel creates a new webhook notification channel
func NewWebhookNotificationChannel(logger *Logger, config *WebhookConfig) *WebhookNotificationChannel {
	return &WebhookNotificationChannel{
		logger:  logger,
		config:  config,
		enabled: true,
	}
}

// Send sends a webhook notification
func (wnc *WebhookNotificationChannel) Send(alert *Alert) error {
	if !wnc.enabled {
		return fmt.Errorf("webhook notification channel is disabled")
	}

	wnc.logger.Info("Sending webhook notification", map[string]interface{}{
		"alert_id":   alert.ID,
		"alert_name": alert.Name,
		"severity":   alert.Severity,
		"url":        wnc.config.URL,
	})

	// In a real implementation, this would send an actual HTTP request
	return nil
}

// Name returns the channel name
func (wnc *WebhookNotificationChannel) Name() string {
	return "webhook"
}

// Type returns the channel type
func (wnc *WebhookNotificationChannel) Type() string {
	return "webhook"
}

// Enabled returns whether the channel is enabled
func (wnc *WebhookNotificationChannel) Enabled() bool {
	return wnc.enabled
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logger *Logger, config *AlertConfig) *AlertManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &AlertManager{
		logger:               logger,
		alerts:               make(map[string]*Alert),
		alertRules:           make(map[string]*AlertRule),
		notificationChannels: make(map[string]NotificationChannel),
		escalationPolicies:   make(map[string]*EscalationPolicy),
		config:               config,
		ctx:                  ctx,
		cancel:               cancel,
	}
}

// Start starts the alert manager
func (am *AlertManager) Start() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.started {
		return fmt.Errorf("alert manager already started")
	}

	am.logger.Info("Starting alert manager", map[string]interface{}{
		"service_name": am.config.ServiceName,
		"version":      am.config.Version,
		"environment":  am.config.Environment,
	})

	// Initialize default alert rules
	if err := am.initializeDefaultAlertRules(); err != nil {
		return fmt.Errorf("failed to initialize default alert rules: %w", err)
	}

	// Initialize default escalation policies
	if err := am.initializeDefaultEscalationPolicies(); err != nil {
		return fmt.Errorf("failed to initialize default escalation policies: %w", err)
	}

	// Start alert evaluation
	if am.config.Enabled {
		go am.startAlertEvaluation()
	}

	// Start escalation processing
	if am.config.EscalationEnabled {
		go am.startEscalationProcessing()
	}

	am.started = true
	am.logger.Info("Alert manager started successfully", map[string]interface{}{})
	return nil
}

// Stop stops the alert manager
func (am *AlertManager) Stop() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if !am.started {
		return fmt.Errorf("alert manager not started")
	}

	am.logger.Info("Stopping alert manager", map[string]interface{}{})

	am.cancel()
	am.started = false

	am.logger.Info("Alert manager stopped successfully", map[string]interface{}{})
	return nil
}

// AddAlertRule adds a new alert rule
func (am *AlertManager) AddAlertRule(rule *AlertRule) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if rule.ID == "" {
		return fmt.Errorf("alert rule ID cannot be empty")
	}

	if _, exists := am.alertRules[rule.ID]; exists {
		return fmt.Errorf("alert rule with ID %s already exists", rule.ID)
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	am.alertRules[rule.ID] = rule

	am.logger.Info("Alert rule added", map[string]interface{}{
		"rule_id":  rule.ID,
		"name":     rule.Name,
		"severity": rule.Severity,
	})

	return nil
}

// RemoveAlertRule removes an alert rule
func (am *AlertManager) RemoveAlertRule(ruleID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.alertRules[ruleID]; !exists {
		return fmt.Errorf("alert rule with ID %s not found", ruleID)
	}

	delete(am.alertRules, ruleID)

	am.logger.Info("Alert rule removed", map[string]interface{}{
		"rule_id": ruleID,
	})

	return nil
}

// GetAlertRule returns an alert rule
func (am *AlertManager) GetAlertRule(ruleID string) (*AlertRule, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	rule, exists := am.alertRules[ruleID]
	if !exists {
		return nil, fmt.Errorf("alert rule with ID %s not found", ruleID)
	}

	// Return a copy
	return &AlertRule{
		ID:                   rule.ID,
		Name:                 rule.Name,
		Description:          rule.Description,
		Query:                rule.Query,
		Condition:            rule.Condition,
		Threshold:            rule.Threshold,
		Severity:             rule.Severity,
		Duration:             rule.Duration,
		Labels:               rule.Labels,
		Annotations:          rule.Annotations,
		NotificationChannels: rule.NotificationChannels,
		EscalationPolicy:     rule.EscalationPolicy,
		Enabled:              rule.Enabled,
		CreatedAt:            rule.CreatedAt,
		UpdatedAt:            rule.UpdatedAt,
	}, nil
}

// ListAlertRules returns all alert rules
func (am *AlertManager) ListAlertRules() []*AlertRule {
	am.mu.RLock()
	defer am.mu.RUnlock()

	rules := make([]*AlertRule, 0, len(am.alertRules))
	for _, rule := range am.alertRules {
		rules = append(rules, &AlertRule{
			ID:                   rule.ID,
			Name:                 rule.Name,
			Description:          rule.Description,
			Query:                rule.Query,
			Condition:            rule.Condition,
			Threshold:            rule.Threshold,
			Severity:             rule.Severity,
			Duration:             rule.Duration,
			Labels:               rule.Labels,
			Annotations:          rule.Annotations,
			NotificationChannels: rule.NotificationChannels,
			EscalationPolicy:     rule.EscalationPolicy,
			Enabled:              rule.Enabled,
			CreatedAt:            rule.CreatedAt,
			UpdatedAt:            rule.UpdatedAt,
		})
	}

	return rules
}

// GetActiveAlerts returns all active alerts
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0)
	for _, alert := range am.alerts {
		if alert.Status == AlertStatusActive {
			alerts = append(alerts, &Alert{
				ID:                alert.ID,
				RuleID:            alert.RuleID,
				Name:              alert.Name,
				Description:       alert.Description,
				Severity:          alert.Severity,
				Status:            alert.Status,
				State:             alert.State,
				Labels:            alert.Labels,
				Annotations:       alert.Annotations,
				Value:             alert.Value,
				Threshold:         alert.Threshold,
				Condition:         alert.Condition,
				StartedAt:         alert.StartedAt,
				UpdatedAt:         alert.UpdatedAt,
				ResolvedAt:        alert.ResolvedAt,
				LastNotifiedAt:    alert.LastNotifiedAt,
				NotificationCount: alert.NotificationCount,
				EscalationLevel:   alert.EscalationLevel,
				Metadata:          alert.Metadata,
			})
		}
	}

	return alerts
}

// GetAlert returns a specific alert
func (am *AlertManager) GetAlert(alertID string) (*Alert, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return nil, fmt.Errorf("alert with ID %s not found", alertID)
	}

	// Return a copy
	return &Alert{
		ID:                alert.ID,
		RuleID:            alert.RuleID,
		Name:              alert.Name,
		Description:       alert.Description,
		Severity:          alert.Severity,
		Status:            alert.Status,
		State:             alert.State,
		Labels:            alert.Labels,
		Annotations:       alert.Annotations,
		Value:             alert.Value,
		Threshold:         alert.Threshold,
		Condition:         alert.Condition,
		StartedAt:         alert.StartedAt,
		UpdatedAt:         alert.UpdatedAt,
		ResolvedAt:        alert.ResolvedAt,
		LastNotifiedAt:    alert.LastNotifiedAt,
		NotificationCount: alert.NotificationCount,
		EscalationLevel:   alert.EscalationLevel,
		Metadata:          alert.Metadata,
	}, nil
}

// AddNotificationChannel adds a notification channel
func (am *AlertManager) AddNotificationChannel(name string, channel NotificationChannel) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.notificationChannels[name] = channel

	am.logger.Info("Notification channel added", map[string]interface{}{
		"name": name,
		"type": channel.Type(),
	})
}

// RemoveNotificationChannel removes a notification channel
func (am *AlertManager) RemoveNotificationChannel(name string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	delete(am.notificationChannels, name)

	am.logger.Info("Notification channel removed", map[string]interface{}{
		"name": name,
	})
}

// AddEscalationPolicy adds an escalation policy
func (am *AlertManager) AddEscalationPolicy(policy *EscalationPolicy) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if policy.ID == "" {
		return fmt.Errorf("escalation policy ID cannot be empty")
	}

	if _, exists := am.escalationPolicies[policy.ID]; exists {
		return fmt.Errorf("escalation policy with ID %s already exists", policy.ID)
	}

	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	am.escalationPolicies[policy.ID] = policy

	am.logger.Info("Escalation policy added", map[string]interface{}{
		"policy_id": policy.ID,
		"name":      policy.Name,
		"levels":    len(policy.Levels),
	})

	return nil
}

// initializeDefaultAlertRules initializes default alert rules
func (am *AlertManager) initializeDefaultAlertRules() error {
	// High error rate alert
	highErrorRateRule := &AlertRule{
		ID:          "high_error_rate",
		Name:        "High Error Rate",
		Description: "Alert when error rate exceeds 5%",
		Query:       "rate(kyb_http_requests_total{status=~\"5..\"}[5m]) / rate(kyb_http_requests_total[5m]) * 100",
		Condition:   "gt",
		Threshold:   5.0,
		Severity:    AlertSeverityCritical,
		Duration:    2 * time.Minute,
		Labels: map[string]string{
			"service":     am.config.ServiceName,
			"environment": am.config.Environment,
		},
		Annotations: map[string]string{
			"summary":     "High error rate detected",
			"description": "Error rate is {{ $value }}% for the last 5 minutes",
		},
		NotificationChannels: []string{"email", "slack"},
		EscalationPolicy:     "default",
		Enabled:              true,
	}

	// High response time alert
	highResponseTimeRule := &AlertRule{
		ID:          "high_response_time",
		Name:        "High Response Time",
		Description: "Alert when 95th percentile response time exceeds 1 second",
		Query:       "histogram_quantile(0.95, rate(kyb_http_request_duration_seconds_bucket[5m]))",
		Condition:   "gt",
		Threshold:   1.0,
		Severity:    AlertSeverityWarning,
		Duration:    2 * time.Minute,
		Labels: map[string]string{
			"service":     am.config.ServiceName,
			"environment": am.config.Environment,
		},
		Annotations: map[string]string{
			"summary":     "High response time detected",
			"description": "95th percentile response time is {{ $value }} seconds",
		},
		NotificationChannels: []string{"slack"},
		EscalationPolicy:     "default",
		Enabled:              true,
	}

	// Service down alert
	serviceDownRule := &AlertRule{
		ID:          "service_down",
		Name:        "Service Down",
		Description: "Alert when service is not responding",
		Query:       "up{job=\"kyb-platform\"}",
		Condition:   "eq",
		Threshold:   0,
		Severity:    AlertSeverityCritical,
		Duration:    1 * time.Minute,
		Labels: map[string]string{
			"service":     am.config.ServiceName,
			"environment": am.config.Environment,
		},
		Annotations: map[string]string{
			"summary":     "Service is down",
			"description": "The KYB Platform service is not responding",
		},
		NotificationChannels: []string{"email", "slack", "webhook"},
		EscalationPolicy:     "critical",
		Enabled:              true,
	}

	// Add alert rules
	if err := am.AddAlertRule(highErrorRateRule); err != nil {
		return fmt.Errorf("failed to add high error rate rule: %w", err)
	}

	if err := am.AddAlertRule(highResponseTimeRule); err != nil {
		return fmt.Errorf("failed to add high response time rule: %w", err)
	}

	if err := am.AddAlertRule(serviceDownRule); err != nil {
		return fmt.Errorf("failed to add service down rule: %w", err)
	}

	return nil
}

// initializeDefaultEscalationPolicies initializes default escalation policies
func (am *AlertManager) initializeDefaultEscalationPolicies() error {
	// Default escalation policy
	defaultPolicy := &EscalationPolicy{
		ID:          "default",
		Name:        "Default Escalation Policy",
		Description: "Default escalation policy for non-critical alerts",
		Levels: []*EscalationLevel{
			{
				Level:                1,
				Duration:             5 * time.Minute,
				NotificationChannels: []string{"slack"},
				Actions:              []string{"notify"},
			},
			{
				Level:                2,
				Duration:             15 * time.Minute,
				NotificationChannels: []string{"email", "slack"},
				Actions:              []string{"notify", "escalate"},
			},
		},
		Enabled: true,
	}

	// Critical escalation policy
	criticalPolicy := &EscalationPolicy{
		ID:          "critical",
		Name:        "Critical Escalation Policy",
		Description: "Escalation policy for critical alerts",
		Levels: []*EscalationLevel{
			{
				Level:                1,
				Duration:             1 * time.Minute,
				NotificationChannels: []string{"email", "slack", "webhook"},
				Actions:              []string{"notify", "escalate"},
			},
			{
				Level:                2,
				Duration:             5 * time.Minute,
				NotificationChannels: []string{"email", "slack", "webhook"},
				Actions:              []string{"notify", "escalate", "page"},
			},
		},
		Enabled: true,
	}

	// Add escalation policies
	if err := am.AddEscalationPolicy(defaultPolicy); err != nil {
		return fmt.Errorf("failed to add default escalation policy: %w", err)
	}

	if err := am.AddEscalationPolicy(criticalPolicy); err != nil {
		return fmt.Errorf("failed to add critical escalation policy: %w", err)
	}

	return nil
}

// startAlertEvaluation starts the alert evaluation process
func (am *AlertManager) startAlertEvaluation() {
	ticker := time.NewTicker(am.config.EvaluationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-am.ctx.Done():
			am.logger.Info("Alert evaluation stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			am.evaluateAlertRules()
		}
	}
}

// evaluateAlertRules evaluates all enabled alert rules
func (am *AlertManager) evaluateAlertRules() {
	am.mu.RLock()
	rules := make([]*AlertRule, 0, len(am.alertRules))
	for _, rule := range am.alertRules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	am.mu.RUnlock()

	for _, rule := range rules {
		am.evaluateAlertRule(rule)
	}
}

// evaluateAlertRule evaluates a specific alert rule
func (am *AlertManager) evaluateAlertRule(rule *AlertRule) {
	// In a real implementation, this would evaluate the query against metrics
	// For now, we'll simulate evaluation
	value := 0.0 // This would be the actual metric value

	// Check if the condition is met
	conditionMet := false
	switch rule.Condition {
	case "gt":
		conditionMet = value > rule.Threshold
	case "gte":
		conditionMet = value >= rule.Threshold
	case "lt":
		conditionMet = value < rule.Threshold
	case "lte":
		conditionMet = value <= rule.Threshold
	case "eq":
		conditionMet = value == rule.Threshold
	case "ne":
		conditionMet = value != rule.Threshold
	}

	if conditionMet {
		am.triggerAlert(rule, value)
	} else {
		am.resolveAlert(rule.ID)
	}
}

// triggerAlert triggers an alert
func (am *AlertManager) triggerAlert(rule *AlertRule, value float64) {
	am.mu.Lock()
	defer am.mu.Unlock()

	alertID := fmt.Sprintf("%s_%d", rule.ID, time.Now().Unix())

	// Check if alert already exists
	if _, exists := am.alerts[alertID]; exists {
		// Update existing alert
		alert := am.alerts[alertID]
		alert.Value = value
		alert.UpdatedAt = time.Now()
		alert.State = AlertStateFiring

		am.logger.Debug("Alert updated", map[string]interface{}{
			"alert_id": alertID,
			"rule_id":  rule.ID,
			"value":    value,
		})
	} else {
		// Create new alert
		alert := &Alert{
			ID:                alertID,
			RuleID:            rule.ID,
			Name:              rule.Name,
			Description:       rule.Description,
			Severity:          rule.Severity,
			Status:            AlertStatusActive,
			State:             AlertStateFiring,
			Labels:            rule.Labels,
			Annotations:       rule.Annotations,
			Value:             value,
			Threshold:         rule.Threshold,
			Condition:         rule.Condition,
			StartedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			NotificationCount: 0,
			EscalationLevel:   0,
			Metadata: map[string]interface{}{
				"rule_query":    rule.Query,
				"rule_duration": rule.Duration.String(),
			},
		}

		am.alerts[alertID] = alert

		am.logger.Info("Alert triggered", map[string]interface{}{
			"alert_id": alertID,
			"rule_id":  rule.ID,
			"name":     rule.Name,
			"severity": rule.Severity,
			"value":    value,
		})

		// Send initial notification
		go am.sendNotification(alert)
	}
}

// resolveAlert resolves an alert
func (am *AlertManager) resolveAlert(ruleID string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Find and resolve alerts for this rule
	for alertID, alert := range am.alerts {
		if alert.RuleID == ruleID && alert.Status == AlertStatusActive {
			alert.Status = AlertStatusResolved
			alert.State = AlertStateResolved
			alert.UpdatedAt = time.Now()
			now := time.Now()
			alert.ResolvedAt = &now

			am.logger.Info("Alert resolved", map[string]interface{}{
				"alert_id": alertID,
				"rule_id":  ruleID,
				"name":     alert.Name,
			})

			// Send resolution notification
			go am.sendNotification(alert)
		}
	}
}

// sendNotification sends a notification for an alert
func (am *AlertManager) sendNotification(alert *Alert) {
	am.mu.RLock()
	rule, exists := am.alertRules[alert.RuleID]
	am.mu.RUnlock()

	if !exists {
		am.logger.Error("Alert rule not found for notification", map[string]interface{}{
			"alert_id": alert.ID,
			"rule_id":  alert.RuleID,
		})
		return
	}

	// Send to configured notification channels
	for _, channelName := range rule.NotificationChannels {
		am.mu.RLock()
		channel, exists := am.notificationChannels[channelName]
		am.mu.RUnlock()

		if !exists {
			am.logger.Warn("Notification channel not found", map[string]interface{}{
				"channel_name": channelName,
				"alert_id":     alert.ID,
			})
			continue
		}

		if !channel.Enabled() {
			am.logger.Debug("Notification channel disabled", map[string]interface{}{
				"channel_name": channelName,
				"alert_id":     alert.ID,
			})
			continue
		}

		if err := channel.Send(alert); err != nil {
			am.logger.Error("Failed to send notification", map[string]interface{}{
				"channel_name": channelName,
				"alert_id":     alert.ID,
				"error":        err.Error(),
			})
		} else {
			am.logger.Debug("Notification sent successfully", map[string]interface{}{
				"channel_name": channelName,
				"alert_id":     alert.ID,
			})

			// Update notification count and timestamp
			am.mu.Lock()
			alert.NotificationCount++
			now := time.Now()
			alert.LastNotifiedAt = &now
			am.mu.Unlock()
		}
	}
}

// startEscalationProcessing starts the escalation processing
func (am *AlertManager) startEscalationProcessing() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-am.ctx.Done():
			am.logger.Info("Escalation processing stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			am.processEscalations()
		}
	}
}

// processEscalations processes alert escalations
func (am *AlertManager) processEscalations() {
	am.mu.RLock()
	alerts := make([]*Alert, 0, len(am.alerts))
	for _, alert := range am.alerts {
		if alert.Status == AlertStatusActive {
			alerts = append(alerts, alert)
		}
	}
	am.mu.RUnlock()

	for _, alert := range alerts {
		am.processAlertEscalation(alert)
	}
}

// processAlertEscalation processes escalation for a specific alert
func (am *AlertManager) processAlertEscalation(alert *Alert) {
	am.mu.RLock()
	rule, ruleExists := am.alertRules[alert.RuleID]
	policy, policyExists := am.escalationPolicies[rule.EscalationPolicy]
	am.mu.RUnlock()

	if !ruleExists || !policyExists {
		return
	}

	// Check if escalation is needed
	timeSinceLastNotification := time.Since(alert.StartedAt)
	if alert.LastNotifiedAt != nil {
		timeSinceLastNotification = time.Since(*alert.LastNotifiedAt)
	}

	// Find the appropriate escalation level
	var currentLevel *EscalationLevel
	for _, level := range policy.Levels {
		if level.Level > alert.EscalationLevel {
			if timeSinceLastNotification >= level.Duration {
				currentLevel = level
				break
			}
		}
	}

	if currentLevel != nil {
		// Escalate the alert
		am.mu.Lock()
		alert.EscalationLevel = currentLevel.Level
		alert.UpdatedAt = time.Now()
		am.mu.Unlock()

		am.logger.Info("Alert escalated", map[string]interface{}{
			"alert_id":         alert.ID,
			"escalation_level": currentLevel.Level,
			"policy":           policy.Name,
		})

		// Send escalation notification
		go am.sendEscalationNotification(alert, currentLevel)
	}
}

// sendEscalationNotification sends an escalation notification
func (am *AlertManager) sendEscalationNotification(alert *Alert, level *EscalationLevel) {
	for _, channelName := range level.NotificationChannels {
		am.mu.RLock()
		channel, exists := am.notificationChannels[channelName]
		am.mu.RUnlock()

		if !exists || !channel.Enabled() {
			continue
		}

		if err := channel.Send(alert); err != nil {
			am.logger.Error("Failed to send escalation notification", map[string]interface{}{
				"channel_name": channelName,
				"alert_id":     alert.ID,
				"level":        level.Level,
				"error":        err.Error(),
			})
		}
	}
}
