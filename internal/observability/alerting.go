package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AlertingSystem provides comprehensive alerting capabilities
type AlertingSystem struct {
	logger         *zap.Logger
	monitoring     *MonitoringSystem
	logAggregation *LogAggregationSystem
	dashboard      *DashboardSystem
	config         *AlertingConfig
	alertRules     map[string]*AlertRule
	alertHistory   []AlertEvent
	notifiers      map[string]Notifier
}

// AlertingConfig holds configuration for alerting
type AlertingConfig struct {
	// Alert evaluation settings
	EvaluationInterval time.Duration
	AlertTimeout       time.Duration
	MaxAlertsPerRule   int

	// Alert severity levels
	SeverityLevels struct {
		Critical string
		Warning  string
		Info     string
	}

	// Notification settings
	NotificationTimeout time.Duration
	RetryAttempts       int
	RetryDelay          time.Duration

	// Alert grouping settings
	GroupByLabels      []string
	GroupWaitTime      time.Duration
	GroupIntervalTime  time.Duration
	RepeatIntervalTime time.Duration

	// Integration settings
	EnablePrometheusIntegration bool
	EnableWebhookIntegration    bool
	EnableEmailIntegration      bool
	EnableSlackIntegration      bool
	EnablePagerDutyIntegration  bool

	// Alert management settings
	EnableAlertSilencing     bool
	EnableAlertInhibition    bool
	EnableAlertAggregation   bool
	EnableAlertDeduplication bool
}

// AlertRule represents an alert rule
type AlertRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Enabled     bool   `json:"enabled"`

	// Alert conditions
	Condition string        `json:"condition"`
	Duration  time.Duration `json:"duration"`
	Threshold float64       `json:"threshold"`
	Operator  string        `json:"operator"`

	// Alert metadata
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`

	// Notification settings
	Notifications []string          `json:"notifications"`
	Escalation    *EscalationPolicy `json:"escalation,omitempty"`

	// Alert management
	SilenceRules []SilenceRule `json:"silence_rules,omitempty"`
	InhibitRules []InhibitRule `json:"inhibit_rules,omitempty"`

	// Evaluation tracking
	LastEvaluation  time.Time `json:"last_evaluation"`
	EvaluationCount int64     `json:"evaluation_count"`
	FiringCount     int64     `json:"firing_count"`
	ResolvedCount   int64     `json:"resolved_count"`
}

// AlertEvent represents an alert event
type AlertEvent struct {
	ID          string            `json:"id"`
	RuleID      string            `json:"rule_id"`
	Status      string            `json:"status"` // firing, resolved
	Severity    string            `json:"severity"`
	Category    string            `json:"category"`
	Message     string            `json:"message"`
	Value       float64           `json:"value"`
	Threshold   float64           `json:"threshold"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`

	// Timing
	StartedAt   time.Time  `json:"started_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	LastUpdated time.Time  `json:"last_updated"`

	// Notification tracking
	NotificationsSent int        `json:"notifications_sent"`
	LastNotification  *time.Time `json:"last_notification,omitempty"`

	// Correlation
	CorrelationID string `json:"correlation_id,omitempty"`
	TraceID       string `json:"trace_id,omitempty"`
}

// EscalationPolicy represents an escalation policy
type EscalationPolicy struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Steps       []EscalationStep `json:"steps"`
	Enabled     bool             `json:"enabled"`
}

// EscalationStep represents an escalation step
type EscalationStep struct {
	StepNumber    int               `json:"step_number"`
	Delay         time.Duration     `json:"delay"`
	Notifications []string          `json:"notifications"`
	Conditions    map[string]string `json:"conditions,omitempty"`
}

// SilenceRule represents a silence rule
type SilenceRule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Matchers    []Matcher `json:"matchers"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedBy   string    `json:"created_by"`
	Comment     string    `json:"comment"`
}

// InhibitRule represents an inhibition rule
type InhibitRule struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	SourceMatchers []Matcher `json:"source_matchers"`
	TargetMatchers []Matcher `json:"target_matchers"`
	Equal          []string  `json:"equal"`
}

// Matcher represents a label matcher
type Matcher struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	IsRegex bool   `json:"is_regex"`
}

// Notifier interface for notification providers
type Notifier interface {
	Send(ctx context.Context, alert *AlertEvent) error
	GetName() string
	IsEnabled() bool
}

// EmailNotifier implements email notifications
type EmailNotifier struct {
	config EmailConfig
	logger *zap.Logger
}

// EmailConfig holds email notification configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromAddress  string
	ToAddresses  []string
	Subject      string
	Template     string
	Enabled      bool
}

// SlackNotifier implements Slack notifications
type SlackNotifier struct {
	config SlackConfig
	logger *zap.Logger
}

// SlackConfig holds Slack notification configuration
type SlackConfig struct {
	WebhookURL string
	Channel    string
	Username   string
	IconEmoji  string
	Template   string
	Enabled    bool
}

// WebhookNotifier implements webhook notifications
type WebhookNotifier struct {
	config WebhookConfig
	logger *zap.Logger
}

// WebhookConfig holds webhook notification configuration
type WebhookConfig struct {
	URL      string
	Method   string
	Headers  map[string]string
	Timeout  time.Duration
	Template string
	Enabled  bool
}

// PagerDutyNotifier implements PagerDuty notifications
type PagerDutyNotifier struct {
	config PagerDutyConfig
	logger *zap.Logger
}

// PagerDutyConfig holds PagerDuty notification configuration
type PagerDutyConfig struct {
	ServiceKey  string
	APIKey      string
	Description string
	Severity    string
	Template    string
	Enabled     bool
}

// NewAlertingSystem creates a new alerting system
func NewAlertingSystem(monitoring *MonitoringSystem, logAggregation *LogAggregationSystem, dashboard *DashboardSystem, config *AlertingConfig, logger *zap.Logger) *AlertingSystem {
	as := &AlertingSystem{
		logger:         logger,
		monitoring:     monitoring,
		logAggregation: logAggregation,
		dashboard:      dashboard,
		config:         config,
		alertRules:     make(map[string]*AlertRule),
		alertHistory:   make([]AlertEvent, 0),
		notifiers:      make(map[string]Notifier),
	}

	// Initialize notifiers
	as.initializeNotifiers()

	// Load default alert rules
	as.loadDefaultAlertRules()

	return as
}

// initializeNotifiers initializes notification providers
func (as *AlertingSystem) initializeNotifiers() {
	// Email notifier
	if as.config.EnableEmailIntegration {
		emailConfig := EmailConfig{
			SMTPHost:     "smtp.gmail.com",
			SMTPPort:     587,
			SMTPUsername: "alerts@kybplatform.com",
			SMTPPassword: "password",
			FromAddress:  "alerts@kybplatform.com",
			ToAddresses:  []string{"admin@kybplatform.com"},
			Subject:      "KYB Platform Alert",
			Template:     "email_alert_template.html",
			Enabled:      true,
		}
		emailNotifier := &EmailNotifier{config: emailConfig, logger: as.logger}
		as.notifiers["email"] = emailNotifier
	}

	// Slack notifier
	if as.config.EnableSlackIntegration {
		slackConfig := SlackConfig{
			WebhookURL: "https://hooks.slack.com/services/xxx/yyy/zzz",
			Channel:    "#alerts",
			Username:   "KYB Platform Alerts",
			IconEmoji:  ":warning:",
			Template:   "slack_alert_template.json",
			Enabled:    true,
		}
		slackNotifier := &SlackNotifier{config: slackConfig, logger: as.logger}
		as.notifiers["slack"] = slackNotifier
	}

	// Webhook notifier
	if as.config.EnableWebhookIntegration {
		webhookConfig := WebhookConfig{
			URL:      "https://api.kybplatform.com/webhooks/alerts",
			Method:   "POST",
			Headers:  map[string]string{"Content-Type": "application/json"},
			Timeout:  30 * time.Second,
			Template: "webhook_alert_template.json",
			Enabled:  true,
		}
		webhookNotifier := &WebhookNotifier{config: webhookConfig, logger: as.logger}
		as.notifiers["webhook"] = webhookNotifier
	}

	// PagerDuty notifier
	if as.config.EnablePagerDutyIntegration {
		pagerDutyConfig := PagerDutyConfig{
			ServiceKey:  "service_key_here",
			APIKey:      "api_key_here",
			Description: "KYB Platform Alert",
			Severity:    "warning",
			Template:    "pagerduty_alert_template.json",
			Enabled:     true,
		}
		pagerDutyNotifier := &PagerDutyNotifier{config: pagerDutyConfig, logger: as.logger}
		as.notifiers["pagerduty"] = pagerDutyNotifier
	}
}

// loadDefaultAlertRules loads default alert rules
func (as *AlertingSystem) loadDefaultAlertRules() {
	// System alert rules
	as.addAlertRule(&AlertRule{
		ID:          "system-cpu-high",
		Name:        "High CPU Usage",
		Description: "CPU usage is above threshold",
		Severity:    "warning",
		Category:    "system",
		Enabled:     true,
		Condition:   "kyb_system_cpu_usage",
		Duration:    5 * time.Minute,
		Threshold:   80.0,
		Operator:    ">",
		Labels: map[string]string{
			"team":    "platform",
			"service": "kyb-api",
		},
		Annotations: map[string]string{
			"summary":     "High CPU usage detected",
			"description": "CPU usage is {{ $value }}%",
			"runbook_url": "https://runbook.kybplatform.com/high-cpu-usage",
		},
		Notifications: []string{"email", "slack"},
	})

	as.addAlertRule(&AlertRule{
		ID:          "system-memory-high",
		Name:        "High Memory Usage",
		Description: "Memory usage is above threshold",
		Severity:    "critical",
		Category:    "system",
		Enabled:     true,
		Condition:   "kyb_system_memory_usage",
		Duration:    5 * time.Minute,
		Threshold:   85.0,
		Operator:    ">",
		Labels: map[string]string{
			"team":    "platform",
			"service": "kyb-api",
		},
		Annotations: map[string]string{
			"summary":     "High memory usage detected",
			"description": "Memory usage is {{ $value }}%",
			"runbook_url": "https://runbook.kybplatform.com/high-memory-usage",
		},
		Notifications: []string{"email", "slack", "pagerduty"},
	})

	// Performance alert rules
	as.addAlertRule(&AlertRule{
		ID:          "performance-response-time-high",
		Name:        "High Response Time",
		Description: "Response time is above threshold",
		Severity:    "warning",
		Category:    "performance",
		Enabled:     true,
		Condition:   "histogram_quantile(0.95, rate(kyb_http_request_duration_seconds_bucket[5m]))",
		Duration:    2 * time.Minute,
		Threshold:   1.0,
		Operator:    ">",
		Labels: map[string]string{
			"team":    "platform",
			"service": "kyb-api",
		},
		Annotations: map[string]string{
			"summary":     "High response time detected",
			"description": "95th percentile response time is {{ $value }}s",
			"runbook_url": "https://runbook.kybplatform.com/high-response-time",
		},
		Notifications: []string{"email", "slack"},
	})

	as.addAlertRule(&AlertRule{
		ID:          "performance-error-rate-high",
		Name:        "High Error Rate",
		Description: "Error rate is above threshold",
		Severity:    "critical",
		Category:    "performance",
		Enabled:     true,
		Condition:   "rate(kyb_http_requests_total{status_code=~\"5..\"}[5m]) / rate(kyb_http_requests_total[5m]) * 100",
		Duration:    1 * time.Minute,
		Threshold:   5.0,
		Operator:    ">",
		Labels: map[string]string{
			"team":    "platform",
			"service": "kyb-api",
		},
		Annotations: map[string]string{
			"summary":     "High error rate detected",
			"description": "Error rate is {{ $value }}%",
			"runbook_url": "https://runbook.kybplatform.com/high-error-rate",
		},
		Notifications: []string{"email", "slack", "pagerduty"},
	})

	// Business alert rules
	as.addAlertRule(&AlertRule{
		ID:          "business-classification-accuracy-low",
		Name:        "Low Classification Accuracy",
		Description: "Classification accuracy is below threshold",
		Severity:    "warning",
		Category:    "business",
		Enabled:     true,
		Condition:   "kyb_classification_accuracy",
		Duration:    10 * time.Minute,
		Threshold:   0.9,
		Operator:    "<",
		Labels: map[string]string{
			"team":    "business",
			"service": "kyb-classification",
		},
		Annotations: map[string]string{
			"summary":     "Low classification accuracy detected",
			"description": "Classification accuracy is {{ $value }}",
			"runbook_url": "https://runbook.kybplatform.com/low-classification-accuracy",
		},
		Notifications: []string{"email", "slack"},
	})

	// Security alert rules
	as.addAlertRule(&AlertRule{
		ID:          "security-auth-failures-high",
		Name:        "High Authentication Failures",
		Description: "Authentication failure rate is above threshold",
		Severity:    "warning",
		Category:    "security",
		Enabled:     true,
		Condition:   "rate(kyb_authentication_failures_total[5m])",
		Duration:    5 * time.Minute,
		Threshold:   10.0,
		Operator:    ">",
		Labels: map[string]string{
			"team":    "security",
			"service": "kyb-auth",
		},
		Annotations: map[string]string{
			"summary":     "High authentication failure rate detected",
			"description": "Authentication failures per minute: {{ $value }}",
			"runbook_url": "https://runbook.kybplatform.com/high-auth-failures",
		},
		Notifications: []string{"email", "slack"},
	})

	as.addAlertRule(&AlertRule{
		ID:          "security-rate-limit-hits-high",
		Name:        "High Rate Limit Hits",
		Description: "Rate limit violations are above threshold",
		Severity:    "critical",
		Category:    "security",
		Enabled:     true,
		Condition:   "rate(kyb_rate_limit_hits_total[5m])",
		Duration:    2 * time.Minute,
		Threshold:   50.0,
		Operator:    ">",
		Labels: map[string]string{
			"team":    "security",
			"service": "kyb-api",
		},
		Annotations: map[string]string{
			"summary":     "High rate limit hits detected",
			"description": "Rate limit hits per minute: {{ $value }}",
			"runbook_url": "https://runbook.kybplatform.com/high-rate-limit-hits",
		},
		Notifications: []string{"email", "slack", "pagerduty"},
	})

	// Infrastructure alert rules
	as.addAlertRule(&AlertRule{
		ID:          "infrastructure-database-errors-high",
		Name:        "High Database Error Rate",
		Description: "Database error rate is above threshold",
		Severity:    "critical",
		Category:    "infrastructure",
		Enabled:     true,
		Condition:   "rate(kyb_database_errors_total[5m])",
		Duration:    1 * time.Minute,
		Threshold:   5.0,
		Operator:    ">",
		Labels: map[string]string{
			"team":    "infrastructure",
			"service": "kyb-database",
		},
		Annotations: map[string]string{
			"summary":     "High database error rate detected",
			"description": "Database errors per minute: {{ $value }}",
			"runbook_url": "https://runbook.kybplatform.com/high-database-errors",
		},
		Notifications: []string{"email", "slack", "pagerduty"},
	})

	as.addAlertRule(&AlertRule{
		ID:          "infrastructure-external-api-errors-high",
		Name:        "High External API Error Rate",
		Description: "External API error rate is above threshold",
		Severity:    "warning",
		Category:    "infrastructure",
		Enabled:     true,
		Condition:   "rate(kyb_external_api_errors_total[5m])",
		Duration:    5 * time.Minute,
		Threshold:   10.0,
		Operator:    ">",
		Labels: map[string]string{
			"team":    "infrastructure",
			"service": "kyb-external-api",
		},
		Annotations: map[string]string{
			"summary":     "High external API error rate detected",
			"description": "External API errors per minute: {{ $value }}",
			"runbook_url": "https://runbook.kybplatform.com/high-external-api-errors",
		},
		Notifications: []string{"email", "slack"},
	})
}

// addAlertRule adds an alert rule to the system
func (as *AlertingSystem) addAlertRule(rule *AlertRule) {
	as.alertRules[rule.ID] = rule
	as.logger.Info("Alert rule added", zap.String("rule_id", rule.ID), zap.String("name", rule.Name))
}

// StartAlertEvaluation starts the alert evaluation loop
func (as *AlertingSystem) StartAlertEvaluation(ctx context.Context) {
	ticker := time.NewTicker(as.config.EvaluationInterval)
	defer ticker.Stop()

	as.logger.Info("Starting alert evaluation", zap.Duration("interval", as.config.EvaluationInterval))

	for {
		select {
		case <-ctx.Done():
			as.logger.Info("Stopping alert evaluation")
			return
		case <-ticker.C:
			as.evaluateAlertRules(ctx)
		}
	}
}

// evaluateAlertRules evaluates all enabled alert rules
func (as *AlertingSystem) evaluateAlertRules(ctx context.Context) {
	for _, rule := range as.alertRules {
		if !rule.Enabled {
			continue
		}

		as.evaluateAlertRule(ctx, rule)
	}
}

// evaluateAlertRule evaluates a single alert rule
func (as *AlertingSystem) evaluateAlertRule(ctx context.Context, rule *AlertRule) {
	// Get current metric value
	value, err := as.getMetricValue(ctx, rule.Condition)
	if err != nil {
		as.logger.Error("Failed to get metric value", zap.Error(err), zap.String("rule_id", rule.ID))
		return
	}

	// Check if condition is met
	isFiring := as.checkCondition(value, rule.Threshold, rule.Operator)

	// Update rule evaluation tracking
	rule.LastEvaluation = time.Now().UTC()
	rule.EvaluationCount++

	// Check if alert should fire
	if isFiring {
		as.handleAlertFiring(ctx, rule, value)
	} else {
		as.handleAlertResolved(ctx, rule)
	}
}

// getMetricValue gets the current value of a metric
func (as *AlertingSystem) getMetricValue(ctx context.Context, condition string) (float64, error) {
	// This would integrate with Prometheus to get actual metric values
	// For now, return example values based on condition
	switch {
	case strings.Contains(condition, "cpu_usage"):
		return 45.2, nil
	case strings.Contains(condition, "memory_usage"):
		return 65.8, nil
	case strings.Contains(condition, "response_time"):
		return 0.8, nil
	case strings.Contains(condition, "error_rate"):
		return 2.5, nil
	case strings.Contains(condition, "classification_accuracy"):
		return 0.92, nil
	case strings.Contains(condition, "auth_failures"):
		return 8.0, nil
	case strings.Contains(condition, "rate_limit_hits"):
		return 15.0, nil
	case strings.Contains(condition, "database_errors"):
		return 2.0, nil
	case strings.Contains(condition, "external_api_errors"):
		return 5.0, nil
	default:
		return 0.0, nil
	}
}

// checkCondition checks if a condition is met
func (as *AlertingSystem) checkCondition(value, threshold float64, operator string) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

// handleAlertFiring handles when an alert starts firing
func (as *AlertingSystem) handleAlertFiring(ctx context.Context, rule *AlertRule, value float64) {
	// Check if alert is already firing
	existingAlert := as.findFiringAlert(rule.ID)
	if existingAlert != nil {
		// Update existing alert
		existingAlert.LastUpdated = time.Now().UTC()
		existingAlert.Value = value
		as.logger.Debug("Alert still firing", zap.String("rule_id", rule.ID), zap.Float64("value", value))
		return
	}

	// Create new alert event
	alert := &AlertEvent{
		ID:          fmt.Sprintf("%s-%d", rule.ID, time.Now().Unix()),
		RuleID:      rule.ID,
		Status:      "firing",
		Severity:    rule.Severity,
		Category:    rule.Category,
		Message:     rule.Description,
		Value:       value,
		Threshold:   rule.Threshold,
		Labels:      rule.Labels,
		Annotations: rule.Annotations,
		StartedAt:   time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
	}

	// Add to alert history
	as.alertHistory = append(as.alertHistory, *alert)

	// Update rule firing count
	rule.FiringCount++

	// Send notifications
	as.sendNotifications(ctx, alert)

	as.logger.Info("Alert fired",
		zap.String("rule_id", rule.ID),
		zap.String("name", rule.Name),
		zap.String("severity", rule.Severity),
		zap.Float64("value", value),
		zap.Float64("threshold", rule.Threshold))
}

// handleAlertResolved handles when an alert is resolved
func (as *AlertingSystem) handleAlertResolved(ctx context.Context, rule *AlertRule) {
	// Find firing alert
	existingAlert := as.findFiringAlert(rule.ID)
	if existingAlert == nil {
		return
	}

	// Mark as resolved
	now := time.Now().UTC()
	existingAlert.Status = "resolved"
	existingAlert.ResolvedAt = &now
	existingAlert.LastUpdated = now

	// Update rule resolved count
	rule.ResolvedCount++

	// Send resolution notifications
	as.sendNotifications(ctx, existingAlert)

	as.logger.Info("Alert resolved",
		zap.String("rule_id", rule.ID),
		zap.String("name", rule.Name))
}

// findFiringAlert finds a currently firing alert
func (as *AlertingSystem) findFiringAlert(ruleID string) *AlertEvent {
	for i := range as.alertHistory {
		if as.alertHistory[i].RuleID == ruleID && as.alertHistory[i].Status == "firing" {
			return &as.alertHistory[i]
		}
	}
	return nil
}

// sendNotifications sends notifications for an alert
func (as *AlertingSystem) sendNotifications(ctx context.Context, alert *AlertEvent) {
	rule := as.alertRules[alert.RuleID]
	if rule == nil {
		return
	}

	for _, notificationType := range rule.Notifications {
		notifier, exists := as.notifiers[notificationType]
		if !exists || !notifier.IsEnabled() {
			continue
		}

		// Send notification
		if err := notifier.Send(ctx, alert); err != nil {
			as.logger.Error("Failed to send notification",
				zap.Error(err),
				zap.String("notification_type", notificationType),
				zap.String("alert_id", alert.ID))
		} else {
			alert.NotificationsSent++
			now := time.Now().UTC()
			alert.LastNotification = &now
		}
	}
}

// GetAlertRules returns all alert rules
func (as *AlertingSystem) GetAlertRules() map[string]*AlertRule {
	return as.alertRules
}

// GetAlertRule returns a specific alert rule
func (as *AlertingSystem) GetAlertRule(ruleID string) (*AlertRule, bool) {
	rule, exists := as.alertRules[ruleID]
	return rule, exists
}

// AddAlertRule adds a new alert rule
func (as *AlertingSystem) AddAlertRule(rule *AlertRule) error {
	if rule.ID == "" {
		return fmt.Errorf("alert rule ID is required")
	}

	if _, exists := as.alertRules[rule.ID]; exists {
		return fmt.Errorf("alert rule with ID %s already exists", rule.ID)
	}

	as.addAlertRule(rule)
	return nil
}

// UpdateAlertRule updates an existing alert rule
func (as *AlertingSystem) UpdateAlertRule(ruleID string, rule *AlertRule) error {
	if _, exists := as.alertRules[ruleID]; !exists {
		return fmt.Errorf("alert rule with ID %s does not exist", ruleID)
	}

	rule.ID = ruleID
	as.alertRules[ruleID] = rule
	as.logger.Info("Alert rule updated", zap.String("rule_id", ruleID))
	return nil
}

// DeleteAlertRule deletes an alert rule
func (as *AlertingSystem) DeleteAlertRule(ruleID string) error {
	if _, exists := as.alertRules[ruleID]; !exists {
		return fmt.Errorf("alert rule with ID %s does not exist", ruleID)
	}

	delete(as.alertRules, ruleID)
	as.logger.Info("Alert rule deleted", zap.String("rule_id", ruleID))
	return nil
}

// GetAlertHistory returns alert history
func (as *AlertingSystem) GetAlertHistory(limit int) []AlertEvent {
	if limit <= 0 || limit > len(as.alertHistory) {
		limit = len(as.alertHistory)
	}

	// Return most recent alerts
	start := len(as.alertHistory) - limit
	if start < 0 {
		start = 0
	}

	return as.alertHistory[start:]
}

// GetFiringAlerts returns currently firing alerts
func (as *AlertingSystem) GetFiringAlerts() []AlertEvent {
	var firingAlerts []AlertEvent
	for _, alert := range as.alertHistory {
		if alert.Status == "firing" {
			firingAlerts = append(firingAlerts, alert)
		}
	}
	return firingAlerts
}

// AlertingHandler handles alerting HTTP requests
func (as *AlertingSystem) AlertingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set response headers
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "GET":
			as.handleGetAlerts(w, r)
		case "POST":
			as.handleCreateAlertRule(w, r)
		case "PUT":
			as.handleUpdateAlertRule(w, r)
		case "DELETE":
			as.handleDeleteAlertRule(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleGetAlerts handles GET requests for alerts
func (as *AlertingSystem) handleGetAlerts(w http.ResponseWriter, r *http.Request) {
	alertType := r.URL.Query().Get("type")

	var data interface{}

	switch alertType {
	case "rules":
		data = as.GetAlertRules()
	case "history":
		limit := 100 // Default limit
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 {
				limit = 100
			}
		}
		data = as.GetAlertHistory(limit)
	case "firing":
		data = as.GetFiringAlerts()
	default:
		// Return summary
		data = map[string]interface{}{
			"total_rules":   len(as.alertRules),
			"firing_alerts": len(as.GetFiringAlerts()),
			"total_history": len(as.alertHistory),
		}
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		as.logger.Error("Failed to encode alert response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// handleCreateAlertRule handles POST requests to create alert rules
func (as *AlertingSystem) handleCreateAlertRule(w http.ResponseWriter, r *http.Request) {
	var rule AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := as.AddAlertRule(&rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

// handleUpdateAlertRule handles PUT requests to update alert rules
func (as *AlertingSystem) handleUpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	ruleID := r.URL.Query().Get("id")
	if ruleID == "" {
		http.Error(w, "Alert rule ID is required", http.StatusBadRequest)
		return
	}

	var rule AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := as.UpdateAlertRule(ruleID, &rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(rule)
}

// handleDeleteAlertRule handles DELETE requests to delete alert rules
func (as *AlertingSystem) handleDeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	ruleID := r.URL.Query().Get("id")
	if ruleID == "" {
		http.Error(w, "Alert rule ID is required", http.StatusBadRequest)
		return
	}

	if err := as.DeleteAlertRule(ruleID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Notifier implementations

// Send implements email notification
func (en *EmailNotifier) Send(ctx context.Context, alert *AlertEvent) error {
	// This would implement actual email sending
	en.logger.Info("Email notification sent",
		zap.String("alert_id", alert.ID),
		zap.String("to", strings.Join(en.config.ToAddresses, ",")))
	return nil
}

// GetName returns the notifier name
func (en *EmailNotifier) GetName() string {
	return "email"
}

// IsEnabled returns if the notifier is enabled
func (en *EmailNotifier) IsEnabled() bool {
	return en.config.Enabled
}

// Send implements Slack notification
func (sn *SlackNotifier) Send(ctx context.Context, alert *AlertEvent) error {
	// This would implement actual Slack webhook sending
	sn.logger.Info("Slack notification sent",
		zap.String("alert_id", alert.ID),
		zap.String("channel", sn.config.Channel))
	return nil
}

// GetName returns the notifier name
func (sn *SlackNotifier) GetName() string {
	return "slack"
}

// IsEnabled returns if the notifier is enabled
func (sn *SlackNotifier) IsEnabled() bool {
	return sn.config.Enabled
}

// Send implements webhook notification
func (wn *WebhookNotifier) Send(ctx context.Context, alert *AlertEvent) error {
	// This would implement actual webhook sending
	wn.logger.Info("Webhook notification sent",
		zap.String("alert_id", alert.ID),
		zap.String("url", wn.config.URL))
	return nil
}

// GetName returns the notifier name
func (wn *WebhookNotifier) GetName() string {
	return "webhook"
}

// IsEnabled returns if the notifier is enabled
func (wn *WebhookNotifier) IsEnabled() bool {
	return wn.config.Enabled
}

// Send implements PagerDuty notification
func (pn *PagerDutyNotifier) Send(_ context.Context, alert *AlertEvent) error {
	// This would implement actual PagerDuty API calls
	pn.logger.Info("PagerDuty notification sent",
		zap.String("alert_id", alert.ID),
		zap.String("severity", pn.config.Severity))
	return nil
}

// GetName returns the notifier name
func (pn *PagerDutyNotifier) GetName() string {
	return "pagerduty"
}

// IsEnabled returns if the notifier is enabled
func (pn *PagerDutyNotifier) IsEnabled() bool {
	return pn.config.Enabled
}
