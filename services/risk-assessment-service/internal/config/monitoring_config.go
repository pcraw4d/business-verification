package config

import (
	"fmt"
	"time"
)

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Prometheus PrometheusConfig `json:"prometheus"`
	Grafana    GrafanaConfig    `json:"grafana"`
	Alerting   AlertingConfig   `json:"alerting"`
	Retention  RetentionConfig  `json:"retention"`
}

// PrometheusConfig represents Prometheus configuration
type PrometheusConfig struct {
	Enabled     bool          `json:"enabled"`
	Port        int           `json:"port"`
	Path        string        `json:"path"`
	Timeout     time.Duration `json:"timeout"`
	MetricsPath string        `json:"metrics_path"`
}

// GrafanaConfig represents Grafana configuration
type GrafanaConfig struct {
	Enabled      bool          `json:"enabled"`
	BaseURL      string        `json:"base_url"`
	APIKey       string        `json:"api_key"`
	Username     string        `json:"username"`
	Password     string        `json:"password"`
	Timeout      time.Duration `json:"timeout"`
	DashboardUID string        `json:"dashboard_uid"`
	AutoCreate   bool          `json:"auto_create"`
}

// AlertingConfig represents alerting configuration
type AlertingConfig struct {
	Enabled  bool                    `json:"enabled"`
	Channels map[string]ChannelConfig `json:"channels"`
	Rules    []AlertRuleConfig       `json:"rules"`
}

// ChannelConfig represents alert channel configuration
type ChannelConfig struct {
	Enabled bool                   `json:"enabled"`
	Type    string                 `json:"type"`
	Config  map[string]interface{} `json:"config"`
}

// AlertRuleConfig represents alert rule configuration
type AlertRuleConfig struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Metric      string        `json:"metric"`
	Condition   string        `json:"condition"`
	Threshold   float64       `json:"threshold"`
	Severity    string        `json:"severity"`
	Duration    time.Duration `json:"duration"`
	Enabled     bool          `json:"enabled"`
	TenantID    string        `json:"tenant_id"`
}

// RetentionConfig represents data retention configuration
type RetentionConfig struct {
	Metrics time.Duration `json:"metrics"`
	Logs    time.Duration `json:"logs"`
	Alerts  time.Duration `json:"alerts"`
	Traces  time.Duration `json:"traces"`
}

// DefaultMonitoringConfig returns default monitoring configuration
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		Prometheus: PrometheusConfig{
			Enabled:     true,
			Port:        9090,
			Path:        "/metrics",
			Timeout:     30 * time.Second,
			MetricsPath: "/metrics",
		},
		Grafana: GrafanaConfig{
			Enabled:      true,
			BaseURL:      "http://localhost:3000",
			APIKey:       "",
			Username:     "admin",
			Password:     "admin",
			Timeout:      30 * time.Second,
			DashboardUID: "risk-assessment-overview",
			AutoCreate:   true,
		},
		Alerting: AlertingConfig{
			Enabled: true,
			Channels: map[string]ChannelConfig{
				"email": {
					Enabled: true,
					Type:    "email",
					Config: map[string]interface{}{
						"smtp_host":     "localhost",
						"smtp_port":     587,
						"username":      "alerts@company.com",
						"password":      "",
						"from_address":  "alerts@company.com",
						"to_addresses":  []string{"admin@company.com"},
						"use_tls":       true,
					},
				},
				"slack": {
					Enabled: false,
					Type:    "slack",
					Config: map[string]interface{}{
						"webhook_url": "",
						"channel":     "#alerts",
						"username":    "Risk Assessment Bot",
						"icon_emoji":  ":warning:",
					},
				},
				"webhook": {
					Enabled: false,
					Type:    "webhook",
					Config: map[string]interface{}{
						"url":     "",
						"method":  "POST",
						"headers": map[string]string{"Content-Type": "application/json"},
						"timeout": "30s",
					},
				},
			},
			Rules: []AlertRuleConfig{
				{
					ID:          "high_error_rate",
					Name:        "High Error Rate",
					Description: "Alert when error rate exceeds 5%",
					Metric:      "rate(risk_assessment_errors_total[5m])",
					Condition:   "greater_than",
					Threshold:   0.05,
					Severity:    "warning",
					Duration:    5 * time.Minute,
					Enabled:     true,
					TenantID:    "",
				},
				{
					ID:          "high_response_time",
					Name:        "High Response Time",
					Description: "Alert when response time exceeds 2 seconds",
					Metric:      "histogram_quantile(0.95, rate(risk_assessment_http_request_duration_seconds_bucket[5m]))",
					Condition:   "greater_than",
					Threshold:   2.0,
					Severity:    "warning",
					Duration:    5 * time.Minute,
					Enabled:     true,
					TenantID:    "",
				},
				{
					ID:          "low_throughput",
					Name:        "Low Throughput",
					Description: "Alert when throughput drops below 10 requests per second",
					Metric:      "rate(risk_assessment_http_requests_total[5m])",
					Condition:   "less_than",
					Threshold:   10.0,
					Severity:    "info",
					Duration:    10 * time.Minute,
					Enabled:     true,
					TenantID:    "",
				},
				{
					ID:          "high_database_connections",
					Name:        "High Database Connections",
					Description: "Alert when database connections exceed 80% of max",
					Metric:      "risk_assessment_database_connections",
					Condition:   "greater_than",
					Threshold:   80.0,
					Severity:    "critical",
					Duration:    2 * time.Minute,
					Enabled:     true,
					TenantID:    "",
				},
				{
					ID:          "compliance_violations",
					Name:        "Compliance Violations",
					Description: "Alert when compliance violations are detected",
					Metric:      "rate(risk_assessment_compliance_violations_total[5m])",
					Condition:   "greater_than",
					Threshold:   0.0,
					Severity:    "critical",
					Duration:    1 * time.Minute,
					Enabled:     true,
					TenantID:    "",
				},
				{
					ID:          "security_incidents",
					Name:        "Security Incidents",
					Description: "Alert when security incidents are detected",
					Metric:      "rate(risk_assessment_security_incidents_total[5m])",
					Condition:   "greater_than",
					Threshold:   0.0,
					Severity:    "emergency",
					Duration:    0 * time.Minute,
					Enabled:     true,
					TenantID:    "",
				},
			},
		},
		Retention: RetentionConfig{
			Metrics: 30 * 24 * time.Hour, // 30 days
			Logs:    7 * 24 * time.Hour,  // 7 days
			Alerts:  90 * 24 * time.Hour, // 90 days
			Traces:  7 * 24 * time.Hour,  // 7 days
		},
	}
}

// LoadMonitoringConfig loads monitoring configuration from environment variables
func LoadMonitoringConfig() *MonitoringConfig {
	config := DefaultMonitoringConfig()
	
	// Load from environment variables if available
	// This would typically use a configuration library like viper
	
	return config
}

// Validate validates the monitoring configuration
func (mc *MonitoringConfig) Validate() error {
	if mc.Prometheus.Enabled && mc.Prometheus.Port <= 0 {
		return fmt.Errorf("prometheus port must be positive when enabled")
	}
	
	if mc.Grafana.Enabled && mc.Grafana.BaseURL == "" {
		return fmt.Errorf("grafana base URL is required when enabled")
	}
	
	if mc.Alerting.Enabled {
		for _, rule := range mc.Alerting.Rules {
			if rule.Metric == "" {
				return fmt.Errorf("alert rule metric is required")
			}
			if rule.Threshold < 0 {
				return fmt.Errorf("alert rule threshold must be non-negative")
			}
		}
	}
	
	return nil
}
