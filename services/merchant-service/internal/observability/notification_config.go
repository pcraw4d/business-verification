package observability

import (
	"fmt"
	"sync"
	"time"
)

// NotificationConfig holds configuration for notification channels
type NotificationConfig struct {
	// Email configuration
	Email *EmailNotificationConfig `json:"email,omitempty"`

	// Slack configuration
	Slack *SlackNotificationConfig `json:"slack,omitempty"`

	// Webhook configuration
	Webhook *WebhookNotificationConfig `json:"webhook,omitempty"`

	// General settings
	Enabled             bool          `json:"enabled"`
	DefaultChannels     []string      `json:"default_channels"`
	RetryAttempts       int           `json:"retry_attempts"`
	RetryInterval       time.Duration `json:"retry_interval"`
	Timeout             time.Duration `json:"timeout"`
	RateLimitPerMinute  int           `json:"rate_limit_per_minute"`
	SuppressionDuration time.Duration `json:"suppression_duration"`
}

// EmailNotificationConfig holds email notification configuration
type EmailNotificationConfig struct {
	Enabled  bool          `json:"enabled"`
	SMTPHost string        `json:"smtp_host"`
	SMTPPort int           `json:"smtp_port"`
	Username string        `json:"username"`
	Password string        `json:"password"`
	From     string        `json:"from"`
	To       []string      `json:"to"`
	CC       []string      `json:"cc,omitempty"`
	BCC      []string      `json:"bcc,omitempty"`
	Subject  string        `json:"subject"`
	Template string        `json:"template,omitempty"`
	UseTLS   bool          `json:"use_tls"`
	UseSSL   bool          `json:"use_ssl"`
	Timeout  time.Duration `json:"timeout"`
}

// SlackNotificationConfig holds Slack notification configuration
type SlackNotificationConfig struct {
	Enabled    bool          `json:"enabled"`
	WebhookURL string        `json:"webhook_url"`
	Channel    string        `json:"channel"`
	Username   string        `json:"username"`
	IconEmoji  string        `json:"icon_emoji,omitempty"`
	IconURL    string        `json:"icon_url,omitempty"`
	Template   string        `json:"template,omitempty"`
	Timeout    time.Duration `json:"timeout"`
}

// WebhookNotificationConfig holds webhook notification configuration
type WebhookNotificationConfig struct {
	Enabled        bool              `json:"enabled"`
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers,omitempty"`
	Timeout        time.Duration     `json:"timeout"`
	Template       string            `json:"template,omitempty"`
	RetryOnFailure bool              `json:"retry_on_failure"`
}

// DefaultNotificationConfig returns default notification configuration
func DefaultNotificationConfig() *NotificationConfig {
	return &NotificationConfig{
		Enabled:             true,
		DefaultChannels:     []string{"slack"},
		RetryAttempts:       3,
		RetryInterval:       30 * time.Second,
		Timeout:             10 * time.Second,
		RateLimitPerMinute:  60,
		SuppressionDuration: 5 * time.Minute,
		Email: &EmailNotificationConfig{
			Enabled:  false,
			SMTPHost: "localhost",
			SMTPPort: 587,
			From:     "alerts@kyb-platform.com",
			To:       []string{"admin@kyb-platform.com"},
			Subject:  "KYB Platform Alert",
			UseTLS:   true,
			Timeout:  10 * time.Second,
		},
		Slack: &SlackNotificationConfig{
			Enabled:   true,
			Channel:   "#alerts",
			Username:  "KYB Platform",
			IconEmoji: ":warning:",
			Timeout:   10 * time.Second,
		},
		Webhook: &WebhookNotificationConfig{
			Enabled:        false,
			Method:         "POST",
			Timeout:        10 * time.Second,
			RetryOnFailure: true,
		},
	}
}

// NotificationChannelFactory creates notification channels based on configuration
type NotificationChannelFactory struct {
	config *NotificationConfig
	logger *Logger
}

// NewNotificationChannelFactory creates a new notification channel factory
func NewNotificationChannelFactory(config *NotificationConfig, logger *Logger) *NotificationChannelFactory {
	return &NotificationChannelFactory{
		config: config,
		logger: logger,
	}
}

// CreateNotificationChannels creates all configured notification channels
func (ncf *NotificationChannelFactory) CreateNotificationChannels() map[string]NotificationChannel {
	channels := make(map[string]NotificationChannel)

	// Create email channel
	if ncf.config.Email != nil && ncf.config.Email.Enabled {
		emailChannel := NewEmailNotificationChannel(ncf.logger, &EmailConfig{
			SMTPHost: ncf.config.Email.SMTPHost,
			SMTPPort: ncf.config.Email.SMTPPort,
			Username: ncf.config.Email.Username,
			Password: ncf.config.Email.Password,
			From:     ncf.config.Email.From,
			To:       ncf.config.Email.To,
			Subject:  ncf.config.Email.Subject,
			Template: ncf.config.Email.Template,
		})
		channels["email"] = emailChannel
	}

	// Create Slack channel
	if ncf.config.Slack != nil && ncf.config.Slack.Enabled {
		slackChannel := NewSlackNotificationChannel(ncf.logger, &SlackConfig{
			WebhookURL: ncf.config.Slack.WebhookURL,
			Channel:    ncf.config.Slack.Channel,
			Username:   ncf.config.Slack.Username,
			IconEmoji:  ncf.config.Slack.IconEmoji,
			Template:   ncf.config.Slack.Template,
		})
		channels["slack"] = slackChannel
	}

	// Create webhook channel
	if ncf.config.Webhook != nil && ncf.config.Webhook.Enabled {
		webhookChannel := NewWebhookNotificationChannel(ncf.logger, &WebhookConfig{
			URL:      ncf.config.Webhook.URL,
			Method:   ncf.config.Webhook.Method,
			Headers:  ncf.config.Webhook.Headers,
			Timeout:  ncf.config.Webhook.Timeout,
			Template: ncf.config.Webhook.Template,
		})
		channels["webhook"] = webhookChannel
	}

	return channels
}

// EnhancedEmailNotificationChannel provides enhanced email notifications
type EnhancedEmailNotificationChannel struct {
	*EmailNotificationChannel
	config *EmailNotificationConfig
}

// NewEnhancedEmailNotificationChannel creates an enhanced email notification channel
func NewEnhancedEmailNotificationChannel(logger *Logger, config *EmailNotificationConfig) *EnhancedEmailNotificationChannel {
	baseConfig := &EmailConfig{
		SMTPHost: config.SMTPHost,
		SMTPPort: config.SMTPPort,
		Username: config.Username,
		Password: config.Password,
		From:     config.From,
		To:       config.To,
		Subject:  config.Subject,
		Template: config.Template,
	}

	return &EnhancedEmailNotificationChannel{
		EmailNotificationChannel: NewEmailNotificationChannel(logger, baseConfig),
		config:                   config,
	}
}

// Send sends an enhanced email notification
func (eenc *EnhancedEmailNotificationChannel) Send(alert *Alert) error {
	if !eenc.config.Enabled {
		return fmt.Errorf("email notification channel is disabled")
	}

	// Use the base implementation
	return eenc.EmailNotificationChannel.Send(alert)
}

// EnhancedSlackNotificationChannel provides enhanced Slack notifications
type EnhancedSlackNotificationChannel struct {
	*SlackNotificationChannel
	config *SlackNotificationConfig
}

// NewEnhancedSlackNotificationChannel creates an enhanced Slack notification channel
func NewEnhancedSlackNotificationChannel(logger *Logger, config *SlackNotificationConfig) *EnhancedSlackNotificationChannel {
	baseConfig := &SlackConfig{
		WebhookURL: config.WebhookURL,
		Channel:    config.Channel,
		Username:   config.Username,
		IconEmoji:  config.IconEmoji,
		Template:   config.Template,
	}

	return &EnhancedSlackNotificationChannel{
		SlackNotificationChannel: NewSlackNotificationChannel(logger, baseConfig),
		config:                   config,
	}
}

// Send sends an enhanced Slack notification
func (esnc *EnhancedSlackNotificationChannel) Send(alert *Alert) error {
	if !esnc.config.Enabled {
		return fmt.Errorf("slack notification channel is disabled")
	}

	// Use the base implementation
	return esnc.SlackNotificationChannel.Send(alert)
}

// EnhancedWebhookNotificationChannel provides enhanced webhook notifications
type EnhancedWebhookNotificationChannel struct {
	*WebhookNotificationChannel
	config *WebhookNotificationConfig
}

// NewEnhancedWebhookNotificationChannel creates an enhanced webhook notification channel
func NewEnhancedWebhookNotificationChannel(logger *Logger, config *WebhookNotificationConfig) *EnhancedWebhookNotificationChannel {
	baseConfig := &WebhookConfig{
		URL:      config.URL,
		Method:   config.Method,
		Headers:  config.Headers,
		Timeout:  config.Timeout,
		Template: config.Template,
	}

	return &EnhancedWebhookNotificationChannel{
		WebhookNotificationChannel: NewWebhookNotificationChannel(logger, baseConfig),
		config:                     config,
	}
}

// Send sends an enhanced webhook notification
func (ewnc *EnhancedWebhookNotificationChannel) Send(alert *Alert) error {
	if !ewnc.config.Enabled {
		return fmt.Errorf("webhook notification channel is disabled")
	}

	// Use the base implementation
	return ewnc.WebhookNotificationChannel.Send(alert)
}

// NotificationTemplateManager manages notification templates
type NotificationTemplateManager struct {
	templates map[string]string
}

// NewNotificationTemplateManager creates a new template manager
func NewNotificationTemplateManager() *NotificationTemplateManager {
	return &NotificationTemplateManager{
		templates: make(map[string]string),
	}
}

// LoadDefaultTemplates loads default notification templates
func (ntm *NotificationTemplateManager) LoadDefaultTemplates() {
	ntm.templates["email_alert"] = `
Subject: {{ .Alert.Name }} - {{ .Alert.Severity | upper }}

Alert Details:
- Name: {{ .Alert.Name }}
- Description: {{ .Alert.Description }}
- Severity: {{ .Alert.Severity }}
- Value: {{ .Alert.Value }}
- Threshold: {{ .Alert.Threshold }}
- Started At: {{ .Alert.StartedAt.Format "2006-01-02 15:04:05 UTC" }}

Labels:
{{ range $key, $value := .Alert.Labels }}
- {{ $key }}: {{ $value }}
{{ end }}

Annotations:
{{ range $key, $value := .Alert.Annotations }}
- {{ $key }}: {{ $value }}
{{ end }}

Please investigate this alert immediately.

Best regards,
KYB Platform Monitoring System
`

	ntm.templates["slack_alert"] = `
🚨 *{{ .Alert.Name }}* - *{{ .Alert.Severity | upper }}*

*Description:* {{ .Alert.Description }}
*Value:* {{ .Alert.Value }} (Threshold: {{ .Alert.Threshold }})
*Started:* {{ .Alert.StartedAt.Format "2006-01-02 15:04:05 UTC" }}

{{ if .Alert.Labels }}
*Labels:*
{{ range $key, $value := .Alert.Labels }}• {{ $key }}: {{ $value }}
{{ end }}
{{ end }}

{{ if .Alert.Annotations }}
*Annotations:*
{{ range $key, $value := .Alert.Annotations }}• {{ $key }}: {{ $value }}
{{ end }}
{{ end }}

Please investigate this alert immediately.
`

	ntm.templates["webhook_alert"] = `{
  "alert_id": "{{ .Alert.ID }}",
  "rule_id": "{{ .Alert.RuleID }}",
  "name": "{{ .Alert.Name }}",
  "description": "{{ .Alert.Description }}",
  "severity": "{{ .Alert.Severity }}",
  "status": "{{ .Alert.Status }}",
  "state": "{{ .Alert.State }}",
  "value": {{ .Alert.Value }},
  "threshold": {{ .Alert.Threshold }},
  "condition": "{{ .Alert.Condition }}",
  "started_at": "{{ .Alert.StartedAt.Format "2006-01-02T15:04:05Z07:00" }}",
  "updated_at": "{{ .Alert.UpdatedAt.Format "2006-01-02T15:04:05Z07:00" }}",
  "labels": {{ .Alert.Labels | toJson }},
  "annotations": {{ .Alert.Annotations | toJson }},
  "metadata": {{ .Alert.Metadata | toJson }}
}`
}

// GetTemplate returns a notification template
func (ntm *NotificationTemplateManager) GetTemplate(templateName string) (string, error) {
	template, exists := ntm.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template %s not found", templateName)
	}
	return template, nil
}

// SetTemplate sets a notification template
func (ntm *NotificationTemplateManager) SetTemplate(templateName, template string) {
	ntm.templates[templateName] = template
}

// RateLimiter provides rate limiting for notifications
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is allowed under rate limiting
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean up old requests
	if requests, exists := rl.requests[key]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[key] = validRequests
	}

	// Check if we're under the limit
	if len(rl.requests[key]) < rl.limit {
		rl.requests[key] = append(rl.requests[key], now)
		return true
	}

	return false
}

// NotificationSuppressor provides alert suppression functionality
type NotificationSuppressor struct {
	suppressed map[string]time.Time
	duration   time.Duration
	mu         sync.RWMutex
}

// NewNotificationSuppressor creates a new notification suppressor
func NewNotificationSuppressor(duration time.Duration) *NotificationSuppressor {
	return &NotificationSuppressor{
		suppressed: make(map[string]time.Time),
		duration:   duration,
	}
}

// IsSuppressed checks if an alert is currently suppressed
func (ns *NotificationSuppressor) IsSuppressed(alertKey string) bool {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	suppressTime, exists := ns.suppressed[alertKey]
	if !exists {
		return false
	}

	// Check if suppression has expired
	if time.Since(suppressTime) > ns.duration {
		return false
	}

	return true
}

// Suppress suppresses an alert for the configured duration
func (ns *NotificationSuppressor) Suppress(alertKey string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.suppressed[alertKey] = time.Now()
}

// ClearSuppression clears suppression for an alert
func (ns *NotificationSuppressor) ClearSuppression(alertKey string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	delete(ns.suppressed, alertKey)
}

// CleanupExpiredSuppressions removes expired suppressions
func (ns *NotificationSuppressor) CleanupExpiredSuppressions() {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	now := time.Now()
	for key, suppressTime := range ns.suppressed {
		if now.Sub(suppressTime) > ns.duration {
			delete(ns.suppressed, key)
		}
	}
}
