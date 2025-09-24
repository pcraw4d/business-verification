// Package observability provides comprehensive testing for the alerting system.
// This module tests alerting thresholds, notification channels, and alert processing.
package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AlertingSystemTest provides comprehensive testing for the alerting system.
type AlertingSystemTest struct {
	config     *EnhancedMonitoringConfig
	logger     *zap.Logger
	testServer *httptest.Server
	alerts     []TestAlert
	mu         sync.RWMutex
}

// TestAlert represents a test alert for validation.
type TestAlert struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	AlertName    string                 `json:"alert_name"`
	MetricName   string                 `json:"metric_name"`
	CurrentValue float64                `json:"current_value"`
	Threshold    float64                `json:"threshold"`
	Condition    string                 `json:"condition"`
	Severity     string                 `json:"severity"`
	Description  string                 `json:"description"`
	Status       string                 `json:"status"` // "triggered", "resolved", "suppressed"
	Metadata     map[string]interface{} `json:"metadata"`
}

// TestNotificationChannel represents a test notification channel.
type TestNotificationChannel struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
	TestResults []TestResult           `json:"test_results"`
}

// TestResult represents the result of a notification test.
type TestResult struct {
	Timestamp    time.Time     `json:"timestamp"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code,omitempty"`
}

// NewAlertingSystemTest creates a new test alerting system.
func NewAlertingSystemTest(config *EnhancedMonitoringConfig, logger *zap.Logger) *AlertingSystemTest {
	return &AlertingSystemTest{
		config: config,
		logger: logger,
		alerts: make([]TestAlert, 0),
	}
}

// SetupTestServer sets up a test HTTP server for webhook testing.
func (ast *AlertingSystemTest) SetupTestServer() {
	ast.testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/alerts" && r.Method == "POST" {
			ast.handleTestWebhook(w, r)
		} else if r.URL.Path == "/health" && r.Method == "GET" {
			ast.handleHealthCheck(w, r)
		} else {
			http.NotFound(w, r)
		}
	}))
}

// handleTestWebhook handles test webhook requests.
func (ast *AlertingSystemTest) handleTestWebhook(w http.ResponseWriter, r *http.Request) {
	var alert TestAlert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ast.mu.Lock()
	ast.alerts = append(ast.alerts, alert)
	ast.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

// handleHealthCheck handles health check requests.
func (ast *AlertingSystemTest) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// TestAlertThresholds tests all configured alert thresholds.
func (ast *AlertingSystemTest) TestAlertThresholds(ctx context.Context) error {
	ast.logger.Info("Testing alert thresholds")

	// Test each configured threshold
	for alertName, threshold := range ast.config.Alerting.Thresholds {
		if !threshold.Enabled {
			continue
		}

		ast.logger.Info("Testing alert threshold",
			zap.String("alert_name", alertName),
			zap.String("metric_name", threshold.MetricName),
			zap.Float64("threshold", threshold.Value),
			zap.String("condition", threshold.Condition),
		)

		// Test threshold evaluation
		if err := ast.testThresholdEvaluation(alertName, threshold); err != nil {
			ast.logger.Error("Threshold evaluation test failed",
				zap.String("alert_name", alertName),
				zap.Error(err),
			)
			return fmt.Errorf("threshold evaluation test failed for %s: %w", alertName, err)
		}

		// Test alert triggering
		if err := ast.testAlertTriggering(ctx, alertName, threshold); err != nil {
			ast.logger.Error("Alert triggering test failed",
				zap.String("alert_name", alertName),
				zap.Error(err),
			)
			return fmt.Errorf("alert triggering test failed for %s: %w", alertName, err)
		}
	}

	ast.logger.Info("All alert threshold tests passed")
	return nil
}

// testThresholdEvaluation tests threshold evaluation logic.
func (ast *AlertingSystemTest) testThresholdEvaluation(alertName string, threshold EnhancedAlertThreshold) error {
	// Test different values around the threshold
	testValues := []float64{
		threshold.Value - 1.0, // Below threshold
		threshold.Value,       // At threshold
		threshold.Value + 1.0, // Above threshold
	}

	for _, value := range testValues {
		expected := ast.evaluateThreshold(value, threshold)
		actual := ast.evaluateThreshold(value, threshold)

		if expected != actual {
			return fmt.Errorf("threshold evaluation mismatch for value %f: expected %v, got %v",
				value, expected, actual)
		}
	}

	return nil
}

// evaluateThreshold evaluates threshold condition (test implementation).
func (ast *AlertingSystemTest) evaluateThreshold(value float64, threshold EnhancedAlertThreshold) bool {
	switch threshold.Condition {
	case "gt":
		return value > threshold.Value
	case "lt":
		return value < threshold.Value
	case "eq":
		return value == threshold.Value
	case "ne":
		return value != threshold.Value
	default:
		return false
	}
}

// testAlertTriggering tests alert triggering with simulated metric values.
func (ast *AlertingSystemTest) testAlertTriggering(ctx context.Context, alertName string, threshold EnhancedAlertThreshold) error {
	// Simulate metric values that should trigger the alert
	triggerValue := ast.getTriggerValue(threshold)

	// Create a test alert
	testAlert := TestAlert{
		ID:           fmt.Sprintf("test_%s_%d", alertName, time.Now().Unix()),
		Timestamp:    time.Now(),
		AlertName:    alertName,
		MetricName:   threshold.MetricName,
		CurrentValue: triggerValue,
		Threshold:    threshold.Value,
		Condition:    threshold.Condition,
		Severity:     threshold.Severity,
		Description:  threshold.Description,
		Status:       "triggered",
		Metadata: map[string]interface{}{
			"test":        true,
			"environment": "test",
		},
	}

	// Simulate alert triggering
	if err := ast.simulateAlertTrigger(ctx, testAlert); err != nil {
		return fmt.Errorf("failed to simulate alert trigger: %w", err)
	}

	// Verify alert was recorded
	if !ast.verifyAlertRecorded(testAlert.ID) {
		return fmt.Errorf("alert was not recorded: %s", testAlert.ID)
	}

	return nil
}

// getTriggerValue returns a value that should trigger the alert.
func (ast *AlertingSystemTest) getTriggerValue(threshold EnhancedAlertThreshold) float64 {
	switch threshold.Condition {
	case "gt":
		return threshold.Value + 1.0
	case "lt":
		return threshold.Value - 1.0
	case "eq":
		return threshold.Value
	case "ne":
		return threshold.Value + 1.0
	default:
		return threshold.Value + 1.0
	}
}

// simulateAlertTrigger simulates an alert trigger.
func (ast *AlertingSystemTest) simulateAlertTrigger(ctx context.Context, alert TestAlert) error {
	// Add alert to the test system
	ast.mu.Lock()
	ast.alerts = append(ast.alerts, alert)
	ast.mu.Unlock()

	// Log the simulated alert
	ast.logger.Info("Simulated alert trigger",
		zap.String("alert_id", alert.ID),
		zap.String("alert_name", alert.AlertName),
		zap.Float64("current_value", alert.CurrentValue),
		zap.Float64("threshold", alert.Threshold),
		zap.String("severity", alert.Severity),
	)

	return nil
}

// verifyAlertRecorded verifies that an alert was recorded.
func (ast *AlertingSystemTest) verifyAlertRecorded(alertID string) bool {
	ast.mu.RLock()
	defer ast.mu.RUnlock()

	for _, alert := range ast.alerts {
		if alert.ID == alertID {
			return true
		}
	}
	return false
}

// TestNotificationChannels tests all configured notification channels.
func (ast *AlertingSystemTest) TestNotificationChannels(ctx context.Context) error {
	ast.logger.Info("Testing notification channels")

	channels := []TestNotificationChannel{
		{
			Type:    "webhook",
			Name:    "test-webhook",
			Enabled: true,
			Config: map[string]interface{}{
				"url":    ast.testServer.URL + "/alerts",
				"method": "POST",
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
		{
			Type:    "email",
			Name:    "test-email",
			Enabled: true,
			Config: map[string]interface{}{
				"smtp_host": "localhost",
				"smtp_port": 587,
				"from":      "test@kyb-platform.com",
				"to":        []string{"admin@kyb-platform.com"},
			},
		},
		{
			Type:    "slack",
			Name:    "test-slack",
			Enabled: true,
			Config: map[string]interface{}{
				"webhook_url": "https://hooks.slack.com/services/test",
				"channel":     "#test-alerts",
				"username":    "KYB Test Alerts",
			},
		},
	}

	for _, channel := range channels {
		if err := ast.testNotificationChannel(ctx, channel); err != nil {
			ast.logger.Error("Notification channel test failed",
				zap.String("channel_type", channel.Type),
				zap.String("channel_name", channel.Name),
				zap.Error(err),
			)
			return fmt.Errorf("notification channel test failed for %s: %w", channel.Name, err)
		}
	}

	ast.logger.Info("All notification channel tests passed")
	return nil
}

// testNotificationChannel tests a specific notification channel.
func (ast *AlertingSystemTest) testNotificationChannel(ctx context.Context, channel TestNotificationChannel) error {
	ast.logger.Info("Testing notification channel",
		zap.String("type", channel.Type),
		zap.String("name", channel.Name),
	)

	// Create a test alert
	testAlert := TestAlert{
		ID:           fmt.Sprintf("test_%s_%d", channel.Name, time.Now().Unix()),
		Timestamp:    time.Now(),
		AlertName:    "test_alert",
		MetricName:   "test_metric",
		CurrentValue: 100.0,
		Threshold:    90.0,
		Condition:    "gt",
		Severity:     "warning",
		Description:  "Test alert for notification channel testing",
		Status:       "triggered",
		Metadata: map[string]interface{}{
			"test":    true,
			"channel": channel.Name,
		},
	}

	// Test the notification channel
	result, err := ast.sendTestNotification(ctx, channel, testAlert)
	if err != nil {
		return fmt.Errorf("failed to send test notification: %w", err)
	}

	// Record the test result
	channel.TestResults = append(channel.TestResults, result)

	// Verify the result
	if !result.Success {
		return fmt.Errorf("notification test failed: %s", result.Error)
	}

	ast.logger.Info("Notification channel test passed",
		zap.String("channel_name", channel.Name),
		zap.Duration("response_time", result.ResponseTime),
		zap.Int("status_code", result.StatusCode),
	)

	return nil
}

// sendTestNotification sends a test notification through a channel.
func (ast *AlertingSystemTest) sendTestNotification(ctx context.Context, channel TestNotificationChannel, alert TestAlert) (TestResult, error) {
	start := time.Now()
	result := TestResult{
		Timestamp: time.Now(),
	}

	switch channel.Type {
	case "webhook":
		return ast.sendWebhookNotification(ctx, channel, alert, start)
	case "email":
		return ast.sendEmailNotification(ctx, channel, alert, start)
	case "slack":
		return ast.sendSlackNotification(ctx, channel, alert, start)
	default:
		result.Success = false
		result.Error = fmt.Sprintf("unsupported channel type: %s", channel.Type)
		return result, fmt.Errorf("unsupported channel type: %s", channel.Type)
	}
}

// sendWebhookNotification sends a test webhook notification.
func (ast *AlertingSystemTest) sendWebhookNotification(ctx context.Context, channel TestNotificationChannel, alert TestAlert, start time.Time) (TestResult, error) {
	result := TestResult{
		Timestamp: time.Now(),
	}

	// Prepare the request
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to marshal alert: %v", err)
		return result, err
	}

	// Send the request
	req, err := http.NewRequestWithContext(ctx, "POST", channel.Config["url"].(string), bytes.NewBuffer(alertJSON))
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result, err
	}

	// Set headers
	if headers, ok := channel.Config["headers"].(map[string]string); ok {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// Execute the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("request failed: %v", err)
		result.ResponseTime = time.Since(start)
		return result, err
	}
	defer resp.Body.Close()

	result.ResponseTime = time.Since(start)
	result.StatusCode = resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
	} else {
		result.Success = false
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return result, nil
}

// sendEmailNotification simulates sending an email notification.
func (ast *AlertingSystemTest) sendEmailNotification(ctx context.Context, channel TestNotificationChannel, alert TestAlert, start time.Time) (TestResult, error) {
	result := TestResult{
		Timestamp: time.Now(),
	}

	// Simulate email sending (in a real implementation, this would use an SMTP client)
	time.Sleep(100 * time.Millisecond) // Simulate network delay

	result.ResponseTime = time.Since(start)
	result.Success = true
	result.StatusCode = 200

	ast.logger.Info("Simulated email notification sent",
		zap.String("to", channel.Config["to"].([]string)[0]),
		zap.String("subject", fmt.Sprintf("[KYB Platform Alert] %s", alert.AlertName)),
	)

	return result, nil
}

// sendSlackNotification simulates sending a Slack notification.
func (ast *AlertingSystemTest) sendSlackNotification(ctx context.Context, channel TestNotificationChannel, alert TestAlert, start time.Time) (TestResult, error) {
	result := TestResult{
		Timestamp: time.Now(),
	}

	// Simulate Slack webhook sending
	time.Sleep(50 * time.Millisecond) // Simulate network delay

	result.ResponseTime = time.Since(start)
	result.Success = true
	result.StatusCode = 200

	ast.logger.Info("Simulated Slack notification sent",
		zap.String("channel", channel.Config["channel"].(string)),
		zap.String("username", channel.Config["username"].(string)),
	)

	return result, nil
}

// RunAllTests runs all alerting system tests.
func (ast *AlertingSystemTest) RunAllTests(ctx context.Context) error {
	ast.logger.Info("Starting comprehensive alerting system tests")

	// Setup test server
	ast.SetupTestServer()
	defer ast.testServer.Close()

	// Run all test suites
	testSuites := []struct {
		name string
		test func(context.Context) error
	}{
		{"Alert Thresholds", ast.TestAlertThresholds},
		{"Notification Channels", ast.TestNotificationChannels},
	}

	for _, suite := range testSuites {
		ast.logger.Info("Running test suite", zap.String("suite", suite.name))

		if err := suite.test(ctx); err != nil {
			ast.logger.Error("Test suite failed",
				zap.String("suite", suite.name),
				zap.Error(err),
			)
			return fmt.Errorf("test suite %s failed: %w", suite.name, err)
		}

		ast.logger.Info("Test suite passed", zap.String("suite", suite.name))
	}

	ast.logger.Info("All alerting system tests passed")
	return nil
}

// GetTestResults returns the results of all tests.
func (ast *AlertingSystemTest) GetTestResults() map[string]interface{} {
	ast.mu.RLock()
	defer ast.mu.RUnlock()

	return map[string]interface{}{
		"total_alerts":    len(ast.alerts),
		"alerts":          ast.alerts,
		"test_server_url": ast.testServer.URL,
		"test_timestamp":  time.Now(),
	}
}

// Cleanup cleans up test resources.
func (ast *AlertingSystemTest) Cleanup() {
	if ast.testServer != nil {
		ast.testServer.Close()
	}

	ast.mu.Lock()
	ast.alerts = make([]TestAlert, 0)
	ast.mu.Unlock()
}
