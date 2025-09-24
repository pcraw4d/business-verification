package config

import (
	"fmt"
	"time"
)

// AlertingConfig holds configuration for the alerting system
type AlertingConfig struct {
	Enabled           bool              `json:"enabled" yaml:"enabled"`
	CheckInterval     time.Duration     `json:"check_interval" yaml:"check_interval"`
	DefaultThresholds AlertThresholds   `json:"default_thresholds" yaml:"default_thresholds"`
	Rules             []AlertRule       `json:"rules" yaml:"rules"`
	Notifiers         []NotifierConfig  `json:"notifiers" yaml:"notifiers"`
	SuppressionRules  []SuppressionRule `json:"suppression_rules" yaml:"suppression_rules"`
}

// AlertThresholds defines default thresholds for common metrics
type AlertThresholds struct {
	ErrorRate           float64       `json:"error_rate" yaml:"error_rate"`
	ResponseTime        time.Duration `json:"response_time" yaml:"response_time"`
	MemoryUsage         float64       `json:"memory_usage" yaml:"memory_usage"`
	CPUUsage            float64       `json:"cpu_usage" yaml:"cpu_usage"`
	DiskUsage           float64       `json:"disk_usage" yaml:"disk_usage"`
	ActiveUsers         int64         `json:"active_users" yaml:"active_users"`
	RequestRate         float64       `json:"request_rate" yaml:"request_rate"`
	DatabaseConnections int64         `json:"database_connections" yaml:"database_connections"`
}

// AlertRule defines a custom alert rule
type AlertRule struct {
	ID          string                 `json:"id" yaml:"id"`
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description" yaml:"description"`
	Severity    AlertSeverity          `json:"severity" yaml:"severity"`
	Condition   string                 `json:"condition" yaml:"condition"`
	Threshold   float64                `json:"threshold" yaml:"threshold"`
	Duration    time.Duration          `json:"duration" yaml:"duration"`
	Labels      map[string]string      `json:"labels" yaml:"labels"`
	Metadata    map[string]interface{} `json:"metadata" yaml:"metadata"`
	Enabled     bool                   `json:"enabled" yaml:"enabled"`
}

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// NotifierConfig defines configuration for alert notifiers
type NotifierConfig struct {
	Type       string                 `json:"type" yaml:"type"`
	Name       string                 `json:"name" yaml:"name"`
	Enabled    bool                   `json:"enabled" yaml:"enabled"`
	Config     map[string]interface{} `json:"config" yaml:"config"`
	Severities []AlertSeverity        `json:"severities" yaml:"severities"`
}

// SuppressionRule defines rules for suppressing alerts
type SuppressionRule struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Conditions  map[string]string `json:"conditions" yaml:"conditions"`
	Duration    time.Duration     `json:"duration" yaml:"duration"`
	Enabled     bool              `json:"enabled" yaml:"enabled"`
}

// GetDefaultAlertingConfig returns the default alerting configuration
func GetDefaultAlertingConfig() *AlertingConfig {
	return &AlertingConfig{
		Enabled:       true,
		CheckInterval: 30 * time.Second,
		DefaultThresholds: AlertThresholds{
			ErrorRate:           5.0, // 5%
			ResponseTime:        2 * time.Second,
			MemoryUsage:         80.0,  // 80%
			CPUUsage:            80.0,  // 80%
			DiskUsage:           90.0,  // 90%
			ActiveUsers:         20,    // 20 users
			RequestRate:         100.0, // 100 req/s
			DatabaseConnections: 50,    // 50 connections
		},
		Rules: []AlertRule{
			{
				ID:          "high_error_rate",
				Name:        "High Error Rate",
				Description: "Error rate has exceeded the threshold",
				Severity:    AlertSeverityCritical,
				Condition:   "greater_than",
				Threshold:   5.0,
				Duration:    2 * time.Minute,
				Labels: map[string]string{
					"service": "api",
					"type":    "error_rate",
				},
				Enabled: true,
			},
			{
				ID:          "high_response_time",
				Name:        "High Response Time",
				Description: "Response time has exceeded the threshold",
				Severity:    AlertSeverityWarning,
				Condition:   "greater_than",
				Threshold:   2000.0, // 2 seconds in milliseconds
				Duration:    5 * time.Minute,
				Labels: map[string]string{
					"service": "api",
					"type":    "response_time",
				},
				Enabled: true,
			},
			{
				ID:          "high_memory_usage",
				Name:        "High Memory Usage",
				Description: "Memory usage has exceeded the threshold",
				Severity:    AlertSeverityWarning,
				Condition:   "greater_than",
				Threshold:   80.0, // 80%
				Duration:    5 * time.Minute,
				Labels: map[string]string{
					"service": "system",
					"type":    "memory",
				},
				Enabled: true,
			},
			{
				ID:          "high_cpu_usage",
				Name:        "High CPU Usage",
				Description: "CPU usage has exceeded the threshold",
				Severity:    AlertSeverityWarning,
				Condition:   "greater_than",
				Threshold:   80.0, // 80%
				Duration:    5 * time.Minute,
				Labels: map[string]string{
					"service": "system",
					"type":    "cpu",
				},
				Enabled: true,
			},
			{
				ID:          "high_disk_usage",
				Name:        "High Disk Usage",
				Description: "Disk usage has exceeded the threshold",
				Severity:    AlertSeverityWarning,
				Condition:   "greater_than",
				Threshold:   90.0, // 90%
				Duration:    10 * time.Minute,
				Labels: map[string]string{
					"service": "system",
					"type":    "disk",
				},
				Enabled: true,
			},
			{
				ID:          "high_active_users",
				Name:        "High Active Users",
				Description: "Number of active users has exceeded the threshold",
				Severity:    AlertSeverityInfo,
				Condition:   "greater_than",
				Threshold:   20.0, // 20 users
				Duration:    5 * time.Minute,
				Labels: map[string]string{
					"service": "api",
					"type":    "users",
				},
				Enabled: true,
			},
			{
				ID:          "high_request_rate",
				Name:        "High Request Rate",
				Description: "Request rate has exceeded the threshold",
				Severity:    AlertSeverityInfo,
				Condition:   "greater_than",
				Threshold:   100.0, // 100 req/s
				Duration:    5 * time.Minute,
				Labels: map[string]string{
					"service": "api",
					"type":    "request_rate",
				},
				Enabled: true,
			},
			{
				ID:          "high_database_connections",
				Name:        "High Database Connections",
				Description: "Number of database connections has exceeded the threshold",
				Severity:    AlertSeverityWarning,
				Condition:   "greater_than",
				Threshold:   50.0, // 50 connections
				Duration:    5 * time.Minute,
				Labels: map[string]string{
					"service": "database",
					"type":    "connections",
				},
				Enabled: true,
			},
			{
				ID:          "service_down",
				Name:        "Service Down",
				Description: "Service is not responding",
				Severity:    AlertSeverityCritical,
				Condition:   "equals",
				Threshold:   0.0, // 0 means down
				Duration:    1 * time.Minute,
				Labels: map[string]string{
					"service": "api",
					"type":    "availability",
				},
				Enabled: true,
			},
			{
				ID:          "database_connection_failed",
				Name:        "Database Connection Failed",
				Description: "Failed to connect to database",
				Severity:    AlertSeverityCritical,
				Condition:   "greater_than",
				Threshold:   0.0, // Any failures
				Duration:    30 * time.Second,
				Labels: map[string]string{
					"service": "database",
					"type":    "connectivity",
				},
				Enabled: true,
			},
		},
		Notifiers: []NotifierConfig{
			{
				Type:    "email",
				Name:    "admin-email",
				Enabled: true,
				Config: map[string]interface{}{
					"smtp_host":     "localhost",
					"smtp_port":     587,
					"smtp_username": "alerts@kyb-platform.com",
					"smtp_password": "password",
					"from":          "alerts@kyb-platform.com",
					"to":            []string{"admin@kyb-platform.com"},
				},
				Severities: []AlertSeverity{
					AlertSeverityCritical,
					AlertSeverityWarning,
				},
			},
			{
				Type:    "slack",
				Name:    "team-slack",
				Enabled: true,
				Config: map[string]interface{}{
					"webhook_url": "YOUR_SLACK_WEBHOOK_URL",
					"channel":     "#alerts",
					"username":    "KYB Platform Alerts",
				},
				Severities: []AlertSeverity{
					AlertSeverityCritical,
					AlertSeverityWarning,
				},
			},
			{
				Type:    "webhook",
				Name:    "custom-webhook",
				Enabled: false,
				Config: map[string]interface{}{
					"url":    "http://localhost:5001/alerts",
					"method": "POST",
					"headers": map[string]string{
						"Content-Type":  "application/json",
						"Authorization": "Bearer YOUR_TOKEN",
					},
				},
				Severities: []AlertSeverity{
					AlertSeverityCritical,
				},
			},
		},
		SuppressionRules: []SuppressionRule{
			{
				ID:          "maintenance_window",
				Name:        "Maintenance Window",
				Description: "Suppress alerts during scheduled maintenance",
				Conditions: map[string]string{
					"time": "02:00-04:00", // 2 AM to 4 AM
					"day":  "sunday",      // Sunday only
				},
				Duration: 2 * time.Hour,
				Enabled:  false, // Disabled by default
			},
			{
				ID:          "deployment_suppression",
				Name:        "Deployment Suppression",
				Description: "Suppress alerts during deployments",
				Conditions: map[string]string{
					"environment": "staging",
					"deployment":  "true",
				},
				Duration: 30 * time.Minute,
				Enabled:  true,
			},
		},
	}
}

// Validate validates the alerting configuration
func (c *AlertingConfig) Validate() error {
	if c.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be positive")
	}

	if c.DefaultThresholds.ErrorRate < 0 || c.DefaultThresholds.ErrorRate > 100 {
		return fmt.Errorf("error_rate threshold must be between 0 and 100")
	}

	if c.DefaultThresholds.MemoryUsage < 0 || c.DefaultThresholds.MemoryUsage > 100 {
		return fmt.Errorf("memory_usage threshold must be between 0 and 100")
	}

	if c.DefaultThresholds.CPUUsage < 0 || c.DefaultThresholds.CPUUsage > 100 {
		return fmt.Errorf("cpu_usage threshold must be between 0 and 100")
	}

	if c.DefaultThresholds.DiskUsage < 0 || c.DefaultThresholds.DiskUsage > 100 {
		return fmt.Errorf("disk_usage threshold must be between 0 and 100")
	}

	// Validate rules
	for i, rule := range c.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule %d: id is required", i)
		}
		if rule.Name == "" {
			return fmt.Errorf("rule %d: name is required", i)
		}
		if rule.Severity == "" {
			return fmt.Errorf("rule %d: severity is required", i)
		}
		if rule.Condition == "" {
			return fmt.Errorf("rule %d: condition is required", i)
		}
		if rule.Duration <= 0 {
			return fmt.Errorf("rule %d: duration must be positive", i)
		}
	}

	// Validate notifiers
	for i, notifier := range c.Notifiers {
		if notifier.Type == "" {
			return fmt.Errorf("notifier %d: type is required", i)
		}
		if notifier.Name == "" {
			return fmt.Errorf("notifier %d: name is required", i)
		}
	}

	return nil
}
