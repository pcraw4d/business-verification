package observability

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestClassificationAlertManager(t *testing.T) {
	// Create test logger
	zapLogger := zap.NewNop()
	logger := NewLogger(zapLogger)

	// Create base alert manager
	baseConfig := &AlertConfig{
		Enabled:              true,
		EvaluationInterval:   30 * time.Second,
		NotificationTimeout:  10 * time.Second,
		MaxRetries:           3,
		RetryInterval:        30 * time.Second,
		SuppressionEnabled:   true,
		SuppressionDuration:  5 * time.Minute,
		DeduplicationEnabled: true,
		EscalationEnabled:    true,
		Environment:          "test",
		ServiceName:          "kyb-platform-test",
		Version:              "1.0.0-test",
	}

	baseAlertManager := NewAlertManager(logger, baseConfig)

	// Create classification alert manager
	classificationConfig := &ClassificationAlertConfig{
		Enabled:              true,
		EvaluationInterval:   30 * time.Second,
		NotificationTimeout:  10 * time.Second,
		MaxRetries:           3,
		RetryInterval:        30 * time.Second,
		SuppressionEnabled:   true,
		SuppressionDuration:  5 * time.Minute,
		DeduplicationEnabled: true,
		EscalationEnabled:    true,
		Environment:          "test",
		ServiceName:          "kyb-platform-test",
		Version:              "1.0.0-test",
	}

	cam := NewClassificationAlertManager(baseAlertManager, zapLogger, classificationConfig)

	// Test starting the alert manager
	if err := cam.Start(); err != nil {
		t.Fatalf("Failed to start classification alert manager: %v", err)
	}
	defer cam.Stop()

	// Test adding a custom alert rule
	customRule := &ClassificationAlertRule{
		ID:          "test_accuracy_rule",
		Name:        "Test Accuracy Rule",
		Description: "Test rule for accuracy monitoring",
		Category:    AlertCategoryAccuracy,
		MetricType:  MetricTypeOverallAccuracy,
		Query:       "kyb_test_accuracy",
		Condition:   "lt",
		Threshold:   0.90,
		Severity:    AlertSeverityWarning,
		Duration:    1 * time.Minute,
		Labels: map[string]string{
			"test": "true",
		},
		Annotations: map[string]string{
			"summary": "Test accuracy alert",
		},
		NotificationChannels: []string{"slack"},
		EscalationPolicy:     "default",
		Enabled:              true,
	}

	if err := cam.AddClassificationAlertRule(customRule); err != nil {
		t.Fatalf("Failed to add custom alert rule: %v", err)
	}

	// Test retrieving the rule
	retrievedRule, err := cam.GetClassificationAlertRule("test_accuracy_rule")
	if err != nil {
		t.Fatalf("Failed to retrieve alert rule: %v", err)
	}

	if retrievedRule.Name != customRule.Name {
		t.Errorf("Expected rule name %s, got %s", customRule.Name, retrievedRule.Name)
	}

	// Test listing all rules
	rules := cam.ListClassificationAlertRules()
	if len(rules) == 0 {
		t.Error("Expected at least one alert rule")
	}

	// Test getting alert summary
	summary := cam.GetClassificationAlertSummary()
	if summary == nil {
		t.Error("Expected alert summary to be non-nil")
	}

	// Test getting alerts by category
	accuracyAlerts := cam.GetClassificationAlertsByCategory(AlertCategoryAccuracy)
	if accuracyAlerts == nil {
		t.Error("Expected accuracy alerts to be non-nil")
	}

	// Test getting alerts by severity
	criticalAlerts := cam.GetClassificationAlertsBySeverity(AlertSeverityCritical)
	if criticalAlerts == nil {
		t.Error("Expected critical alerts to be non-nil")
	}
}

func TestAdvancedAlertingIntegration(t *testing.T) {
	// Create test logger
	zapLogger := zap.NewNop()
	logger := NewLogger(zapLogger)

	// Create base alert manager
	baseConfig := &AlertConfig{
		Enabled:              true,
		EvaluationInterval:   30 * time.Second,
		NotificationTimeout:  10 * time.Second,
		MaxRetries:           3,
		RetryInterval:        30 * time.Second,
		SuppressionEnabled:   true,
		SuppressionDuration:  5 * time.Minute,
		DeduplicationEnabled: true,
		EscalationEnabled:    true,
		Environment:          "test",
		ServiceName:          "kyb-platform-test",
		Version:              "1.0.0-test",
	}

	baseAlertManager := NewAlertManager(logger, baseConfig)

	// Create advanced alerting integration
	integrationConfig := &AdvancedAlertingConfig{
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
		Environment:                     "test",
		ServiceName:                     "kyb-platform-test",
		Version:                         "1.0.0-test",
	}

	integration := NewAdvancedAlertingIntegration(baseAlertManager, zapLogger, integrationConfig)

	// Test starting the integration
	if err := integration.Start(); err != nil {
		t.Fatalf("Failed to start advanced alerting integration: %v", err)
	}
	defer integration.Stop()

	// Test triggering accuracy alert
	err := integration.TriggerAccuracyAlert(
		MetricTypeOverallAccuracy,
		0.85, // Below 95% threshold
		0.95,
		map[string]string{
			"test": "true",
		},
	)
	if err != nil {
		t.Errorf("Failed to trigger accuracy alert: %v", err)
	}

	// Test triggering ML model alert
	err = integration.TriggerMLModelAlert(
		MetricTypeBERTModelDrift,
		0.9, // Above 0.8 threshold
		0.8,
		"bert-base",
		map[string]string{
			"test": "true",
		},
	)
	if err != nil {
		t.Errorf("Failed to trigger ML model alert: %v", err)
	}

	// Test triggering ensemble alert
	err = integration.TriggerEnsembleAlert(
		MetricTypeEnsembleDisagreement,
		0.4, // Above 0.3 threshold
		0.3,
		map[string]string{
			"test": "true",
		},
	)
	if err != nil {
		t.Errorf("Failed to trigger ensemble alert: %v", err)
	}

	// Test triggering security alert
	err = integration.TriggerSecurityAlert(
		MetricTypeSecurityViolation,
		1.0, // Above 0 threshold
		0.0,
		map[string]string{
			"test": "true",
		},
	)
	if err != nil {
		t.Errorf("Failed to trigger security alert: %v", err)
	}

	// Test getting alert summary
	summary := integration.GetAlertSummary()
	if summary == nil {
		t.Error("Expected alert summary to be non-nil")
	}

	if summary.TotalAlerts < 0 {
		t.Error("Expected total alerts to be non-negative")
	}
}

func TestNotificationChannelFactory(t *testing.T) {
	// Create test logger
	zapLogger := zap.NewNop()
	logger := NewLogger(zapLogger)

	// Create notification config
	config := &NotificationConfig{
		Enabled:             true,
		DefaultChannels:     []string{"slack"},
		RetryAttempts:       3,
		RetryInterval:       30 * time.Second,
		Timeout:             10 * time.Second,
		RateLimitPerMinute:  60,
		SuppressionDuration: 5 * time.Minute,
		Email: &EmailNotificationConfig{
			Enabled:  true,
			SMTPHost: "localhost",
			SMTPPort: 587,
			From:     "test@kyb-platform.com",
			To:       []string{"admin@kyb-platform.com"},
			Subject:  "Test Alert",
			UseTLS:   true,
			Timeout:  10 * time.Second,
		},
		Slack: &SlackNotificationConfig{
			Enabled:   true,
			Channel:   "#test-alerts",
			Username:  "Test Bot",
			IconEmoji: ":warning:",
			Timeout:   10 * time.Second,
		},
		Webhook: &WebhookNotificationConfig{
			Enabled:        true,
			URL:            "https://test.example.com/webhook",
			Method:         "POST",
			Timeout:        10 * time.Second,
			RetryOnFailure: true,
		},
	}

	// Create factory
	factory := NewNotificationChannelFactory(config, logger)

	// Create channels
	channels := factory.CreateNotificationChannels()

	// Test that channels were created
	if len(channels) == 0 {
		t.Error("Expected at least one notification channel")
	}

	// Test email channel
	if emailChannel, exists := channels["email"]; exists {
		if emailChannel.Name() != "email" {
			t.Errorf("Expected email channel name 'email', got '%s'", emailChannel.Name())
		}
		if emailChannel.Type() != "email" {
			t.Errorf("Expected email channel type 'email', got '%s'", emailChannel.Type())
		}
		if !emailChannel.Enabled() {
			t.Error("Expected email channel to be enabled")
		}
	}

	// Test Slack channel
	if slackChannel, exists := channels["slack"]; exists {
		if slackChannel.Name() != "slack" {
			t.Errorf("Expected Slack channel name 'slack', got '%s'", slackChannel.Name())
		}
		if slackChannel.Type() != "slack" {
			t.Errorf("Expected Slack channel type 'slack', got '%s'", slackChannel.Type())
		}
		if !slackChannel.Enabled() {
			t.Error("Expected Slack channel to be enabled")
		}
	}

	// Test webhook channel
	if webhookChannel, exists := channels["webhook"]; exists {
		if webhookChannel.Name() != "webhook" {
			t.Errorf("Expected webhook channel name 'webhook', got '%s'", webhookChannel.Name())
		}
		if webhookChannel.Type() != "webhook" {
			t.Errorf("Expected webhook channel type 'webhook', got '%s'", webhookChannel.Type())
		}
		if !webhookChannel.Enabled() {
			t.Error("Expected webhook channel to be enabled")
		}
	}
}

func TestRateLimiter(t *testing.T) {
	// Create rate limiter with low limits for testing
	rateLimiter := NewRateLimiter(2, time.Minute)

	// Test allowing requests within limit
	if !rateLimiter.Allow("test_key") {
		t.Error("Expected first request to be allowed")
	}

	if !rateLimiter.Allow("test_key") {
		t.Error("Expected second request to be allowed")
	}

	// Test rate limiting
	if rateLimiter.Allow("test_key") {
		t.Error("Expected third request to be rate limited")
	}

	// Test different keys
	if !rateLimiter.Allow("different_key") {
		t.Error("Expected request with different key to be allowed")
	}
}

func TestNotificationSuppressor(t *testing.T) {
	// Create suppressor with short duration for testing
	suppressor := NewNotificationSuppressor(100 * time.Millisecond)

	// Test suppression
	alertKey := "test_alert"
	if suppressor.IsSuppressed(alertKey) {
		t.Error("Expected alert to not be suppressed initially")
	}

	suppressor.Suppress(alertKey)

	if !suppressor.IsSuppressed(alertKey) {
		t.Error("Expected alert to be suppressed after suppression")
	}

	// Test suppression expiration
	time.Sleep(150 * time.Millisecond)

	if suppressor.IsSuppressed(alertKey) {
		t.Error("Expected alert to not be suppressed after expiration")
	}

	// Test clearing suppression
	suppressor.Suppress(alertKey)
	suppressor.ClearSuppression(alertKey)

	if suppressor.IsSuppressed(alertKey) {
		t.Error("Expected alert to not be suppressed after clearing")
	}
}

func TestNotificationTemplateManager(t *testing.T) {
	// Create template manager
	templateManager := NewNotificationTemplateManager()

	// Load default templates
	templateManager.LoadDefaultTemplates()

	// Test getting existing template
	template, err := templateManager.GetTemplate("email_alert")
	if err != nil {
		t.Errorf("Failed to get email template: %v", err)
	}

	if template == "" {
		t.Error("Expected email template to be non-empty")
	}

	// Test getting non-existent template
	_, err = templateManager.GetTemplate("non_existent")
	if err == nil {
		t.Error("Expected error when getting non-existent template")
	}

	// Test setting custom template
	customTemplate := "Custom template: {{ .Alert.Name }}"
	templateManager.SetTemplate("custom", customTemplate)

	retrievedTemplate, err := templateManager.GetTemplate("custom")
	if err != nil {
		t.Errorf("Failed to get custom template: %v", err)
	}

	if retrievedTemplate != customTemplate {
		t.Errorf("Expected custom template %s, got %s", customTemplate, retrievedTemplate)
	}
}

func TestAlertCategories(t *testing.T) {
	// Test alert category constants
	categories := []AlertCategory{
		AlertCategoryAccuracy,
		AlertCategoryMLModel,
		AlertCategoryEnsemble,
		AlertCategorySecurity,
		AlertCategoryPerformance,
		AlertCategoryDataQuality,
	}

	expectedCategories := []string{
		"accuracy",
		"ml_model",
		"ensemble",
		"security",
		"performance",
		"data_quality",
	}

	for i, category := range categories {
		if string(category) != expectedCategories[i] {
			t.Errorf("Expected category %s, got %s", expectedCategories[i], string(category))
		}
	}
}

func TestClassificationMetricTypes(t *testing.T) {
	// Test classification metric type constants
	metricTypes := []ClassificationMetricType{
		MetricTypeOverallAccuracy,
		MetricTypeIndustryAccuracy,
		MetricTypeConfidenceScore,
		MetricTypeBERTModelDrift,
		MetricTypeBERTModelAccuracy,
		MetricTypeEnsembleDisagreement,
		MetricTypeWeightDistribution,
		MetricTypeSecurityViolation,
		MetricTypeDataSourceTrust,
		MetricTypeWebsiteVerification,
		MetricTypeProcessingLatency,
		MetricTypeErrorRate,
		MetricTypeThroughput,
	}

	expectedTypes := []string{
		"overall_accuracy",
		"industry_accuracy",
		"confidence_score",
		"bert_model_drift",
		"bert_model_accuracy",
		"ensemble_disagreement",
		"weight_distribution",
		"security_violation",
		"data_source_trust",
		"website_verification",
		"processing_latency",
		"error_rate",
		"throughput",
	}

	for i, metricType := range metricTypes {
		if string(metricType) != expectedTypes[i] {
			t.Errorf("Expected metric type %s, got %s", expectedTypes[i], string(metricType))
		}
	}
}

func TestAlertSeverityLevels(t *testing.T) {
	// Test alert severity constants
	severities := []AlertSeverity{
		AlertSeverityCritical,
		AlertSeverityWarning,
		AlertSeverityInfo,
		AlertSeverityDebug,
	}

	expectedSeverities := []string{
		"critical",
		"warning",
		"info",
		"debug",
	}

	for i, severity := range severities {
		if string(severity) != expectedSeverities[i] {
			t.Errorf("Expected severity %s, got %s", expectedSeverities[i], string(severity))
		}
	}
}

func TestAlertStatusValues(t *testing.T) {
	// Test alert status constants
	statuses := []AlertStatus{
		AlertStatusActive,
		AlertStatusResolved,
		AlertStatusSuppressed,
	}

	expectedStatuses := []string{
		"active",
		"resolved",
		"suppressed",
	}

	for i, status := range statuses {
		if string(status) != expectedStatuses[i] {
			t.Errorf("Expected status %s, got %s", expectedStatuses[i], string(status))
		}
	}
}

func TestAlertStateValues(t *testing.T) {
	// Test alert state constants
	states := []AlertState{
		AlertStateFiring,
		AlertStatePending,
		AlertStateResolved,
		AlertStateSuppressed,
	}

	expectedStates := []string{
		"firing",
		"pending",
		"resolved",
		"suppressed",
	}

	for i, state := range states {
		if string(state) != expectedStates[i] {
			t.Errorf("Expected state %s, got %s", expectedStates[i], string(state))
		}
	}
}

// Benchmark tests
func BenchmarkClassificationAlertManager(b *testing.B) {
	zapLogger := zap.NewNop()
	logger := NewLogger(zapLogger)
	baseConfig := &AlertConfig{
		Enabled:              true,
		EvaluationInterval:   30 * time.Second,
		NotificationTimeout:  10 * time.Second,
		MaxRetries:           3,
		RetryInterval:        30 * time.Second,
		SuppressionEnabled:   true,
		SuppressionDuration:  5 * time.Minute,
		DeduplicationEnabled: true,
		EscalationEnabled:    true,
		Environment:          "test",
		ServiceName:          "kyb-platform-test",
		Version:              "1.0.0-test",
	}

	baseAlertManager := NewAlertManager(logger, baseConfig)
	classificationConfig := &ClassificationAlertConfig{
		Enabled:              true,
		EvaluationInterval:   30 * time.Second,
		NotificationTimeout:  10 * time.Second,
		MaxRetries:           3,
		RetryInterval:        30 * time.Second,
		SuppressionEnabled:   true,
		SuppressionDuration:  5 * time.Minute,
		DeduplicationEnabled: true,
		EscalationEnabled:    true,
		Environment:          "test",
		ServiceName:          "kyb-platform-test",
		Version:              "1.0.0-test",
	}

	cam := NewClassificationAlertManager(baseAlertManager, zapLogger, classificationConfig)
	cam.Start()
	defer cam.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cam.GetClassificationAlertSummary()
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	rateLimiter := NewRateLimiter(1000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rateLimiter.Allow("test_key")
	}
}

func BenchmarkNotificationSuppressor(b *testing.B) {
	suppressor := NewNotificationSuppressor(time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		suppressor.IsSuppressed("test_alert")
	}
}
