package observability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewPerformanceAlertingSystem(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{
		EvaluationInterval: 30 * time.Second,
		AlertTimeout:       5 * time.Minute,
		ResponseTimeThresholds: struct {
			Warning   time.Duration `json:"warning"`
			Critical  time.Duration `json:"critical"`
			Emergency time.Duration `json:"emergency"`
		}{
			Warning:   500 * time.Millisecond,
			Critical:  1000 * time.Millisecond,
			Emergency: 2000 * time.Millisecond,
		},
		SuccessRateThresholds: struct {
			Warning   float64 `json:"warning"`
			Critical  float64 `json:"critical"`
			Emergency float64 `json:"emergency"`
		}{
			Warning:   0.95,
			Critical:  0.90,
			Emergency: 0.80,
		},
	}

	// Create mock components
	alertingSystem := &AlertingSystem{}
	performanceMonitor := &PerformanceMonitor{}
	automatedOptimizer := &AutomatedOptimizer{}

	pas := NewPerformanceAlertingSystem(alertingSystem, performanceMonitor, automatedOptimizer, config, logger)

	assert.NotNil(t, pas)
	assert.Equal(t, alertingSystem, pas.alertingSystem)
	assert.Equal(t, performanceMonitor, pas.performanceMonitor)
	assert.Equal(t, automatedOptimizer, pas.automatedOptimizer)
	assert.Equal(t, config, pas.config)
	assert.NotNil(t, pas.performanceRules)
	assert.NotNil(t, pas.notificationChannels)
	assert.NotNil(t, pas.notificationQueue)
	assert.NotNil(t, pas.activeAlerts)
	assert.NotNil(t, pas.alertHistory)
}

func TestPerformanceAlertingSystem_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{
		EvaluationInterval: 100 * time.Millisecond, // Short interval for testing
	}

	alertingSystem := &AlertingSystem{}
	performanceMonitor := &PerformanceMonitor{}
	automatedOptimizer := &AutomatedOptimizer{}

	pas := NewPerformanceAlertingSystem(alertingSystem, performanceMonitor, automatedOptimizer, config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the system
	err := pas.Start(ctx)
	assert.NoError(t, err)

	// Wait a bit for goroutines to start
	time.Sleep(50 * time.Millisecond)

	// Stop the system
	err = pas.Stop()
	assert.NoError(t, err)
}

func TestPerformanceAlertingSystem_AddPerformanceRule(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}

	alertingSystem := &AlertingSystem{}
	performanceMonitor := &PerformanceMonitor{}
	automatedOptimizer := &AutomatedOptimizer{}

	pas := NewPerformanceAlertingSystem(alertingSystem, performanceMonitor, automatedOptimizer, config, logger)

	rule := &PerformanceAlertRule{
		ID:            "test_rule",
		Name:          "Test Rule",
		Description:   "Test performance alert rule",
		Severity:      "warning",
		Category:      "performance",
		Enabled:       true,
		MetricType:    "response_time",
		Condition:     "threshold",
		Duration:      5 * time.Minute,
		Threshold:     500.0,
		Operator:      "gt",
		Notifications: []string{"email", "slack"},
	}

	err := pas.AddPerformanceRule(rule)
	assert.NoError(t, err)

	// Try to add the same rule again
	err = pas.AddPerformanceRule(rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestPerformanceAlertingSystem_UpdatePerformanceRule(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}

	alertingSystem := &AlertingSystem{}
	performanceMonitor := &PerformanceMonitor{}
	automatedOptimizer := &AutomatedOptimizer{}

	pas := NewPerformanceAlertingSystem(alertingSystem, performanceMonitor, automatedOptimizer, config, logger)

	rule := &PerformanceAlertRule{
		ID:            "test_rule",
		Name:          "Test Rule",
		Description:   "Test performance alert rule",
		Severity:      "warning",
		Category:      "performance",
		Enabled:       true,
		MetricType:    "response_time",
		Condition:     "threshold",
		Duration:      5 * time.Minute,
		Threshold:     500.0,
		Operator:      "gt",
		Notifications: []string{"email", "slack"},
	}

	// Add the rule first
	err := pas.AddPerformanceRule(rule)
	assert.NoError(t, err)

	// Update the rule
	updatedRule := &PerformanceAlertRule{
		ID:            "test_rule",
		Name:          "Updated Test Rule",
		Description:   "Updated test performance alert rule",
		Severity:      "critical",
		Category:      "performance",
		Enabled:       true,
		MetricType:    "response_time",
		Condition:     "threshold",
		Duration:      5 * time.Minute,
		Threshold:     1000.0,
		Operator:      "gt",
		Notifications: []string{"email", "slack", "pagerduty"},
	}

	err = pas.UpdatePerformanceRule("test_rule", updatedRule)
	assert.NoError(t, err)

	// Try to update non-existent rule
	err = pas.UpdatePerformanceRule("non_existent", updatedRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestPerformanceAlertingSystem_DeletePerformanceRule(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}

	alertingSystem := &AlertingSystem{}
	performanceMonitor := &PerformanceMonitor{}
	automatedOptimizer := &AutomatedOptimizer{}

	pas := NewPerformanceAlertingSystem(alertingSystem, performanceMonitor, automatedOptimizer, config, logger)

	rule := &PerformanceAlertRule{
		ID:            "test_rule",
		Name:          "Test Rule",
		Description:   "Test performance alert rule",
		Severity:      "warning",
		Category:      "performance",
		Enabled:       true,
		MetricType:    "response_time",
		Condition:     "threshold",
		Duration:      5 * time.Minute,
		Threshold:     500.0,
		Operator:      "gt",
		Notifications: []string{"email", "slack"},
	}

	// Add the rule first
	err := pas.AddPerformanceRule(rule)
	assert.NoError(t, err)

	// Delete the rule
	err = pas.DeletePerformanceRule("test_rule")
	assert.NoError(t, err)

	// Try to delete non-existent rule
	err = pas.DeletePerformanceRule("non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestPerformanceAlertingSystem_GetActiveAlerts(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}

	alertingSystem := &AlertingSystem{}
	performanceMonitor := &PerformanceMonitor{}
	automatedOptimizer := &AutomatedOptimizer{}

	pas := NewPerformanceAlertingSystem(alertingSystem, performanceMonitor, automatedOptimizer, config, logger)

	// Initially no active alerts
	alerts := pas.GetActiveAlerts()
	assert.Empty(t, alerts)

	// Add a mock alert
	pas.mu.Lock()
	pas.activeAlerts["test_alert"] = &PerformanceAlert{
		ID:       "test_alert",
		RuleID:   "test_rule",
		RuleName: "Test Rule",
		Status:   "firing",
	}
	pas.mu.Unlock()

	alerts = pas.GetActiveAlerts()
	assert.Len(t, alerts, 1)
	assert.Equal(t, "test_alert", alerts[0].ID)
}

func TestPerformanceAlertingSystem_GetAlertHistory(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}

	alertingSystem := &AlertingSystem{}
	performanceMonitor := &PerformanceMonitor{}
	automatedOptimizer := &AutomatedOptimizer{}

	pas := NewPerformanceAlertingSystem(alertingSystem, performanceMonitor, automatedOptimizer, config, logger)

	// Initially no alert history
	history := pas.GetAlertHistory()
	assert.Empty(t, history)

	// Add a mock alert to history
	pas.mu.Lock()
	pas.alertHistory = append(pas.alertHistory, &PerformanceAlert{
		ID:       "test_alert",
		RuleID:   "test_rule",
		RuleName: "Test Rule",
		Status:   "resolved",
	})
	pas.mu.Unlock()

	history = pas.GetAlertHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, "test_alert", history[0].ID)
}

func TestPerformanceAlertingSystem_FormatAlertMessage(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}

	alertingSystem := &AlertingSystem{}
	performanceMonitor := &PerformanceMonitor{}
	automatedOptimizer := &AutomatedOptimizer{}

	pas := NewPerformanceAlertingSystem(alertingSystem, performanceMonitor, automatedOptimizer, config, logger)

	// Add a rule first
	rule := &PerformanceAlertRule{
		ID:            "test_rule",
		Name:          "High Response Time",
		Description:   "Alert when response time is high",
		Severity:      "warning",
		Category:      "performance",
		Enabled:       true,
		MetricType:    "response_time",
		Condition:     "threshold",
		Duration:      5 * time.Minute,
		Threshold:     500.0,
		Operator:      "gt",
		Notifications: []string{"email", "slack"},
	}
	pas.AddPerformanceRule(rule)

	alert := &PerformanceAlert{
		ID:            "test_alert",
		RuleID:        "test_rule",
		RuleName:      "High Response Time",
		Severity:      "warning",
		Category:      "performance",
		Status:        "firing",
		MetricType:    "response_time",
		CurrentValue:  750.0,
		Threshold:     500.0,
		BaselineValue: 200.0,
		TrendValue:    25.0,
		AnomalyScore:  0.8,
		FiredAt:       time.Now().UTC(),
	}

	message := pas.formatAlertMessage(alert)
	assert.Contains(t, message, "🚨 Performance Alert: High Response Time")
	assert.Contains(t, message, "Severity: warning")
	assert.Contains(t, message, "Category: performance")
	assert.Contains(t, message, "Metric: response_time")
	assert.Contains(t, message, "Current Value: 750.00")
	assert.Contains(t, message, "Threshold: 500.00")
	assert.Contains(t, message, "Baseline: 200.00")
	assert.Contains(t, message, "Trend: 25.00")
	assert.Contains(t, message, "Anomaly Score: 0.80")
}

func TestPerformanceRuleEngine_ValidateRule(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}
	rules := make(map[string]*PerformanceAlertRule)

	pre := NewPerformanceRuleEngine(rules, config, logger)

	// Valid rule
	validRule := &PerformanceAlertRule{
		ID:            "valid_rule",
		Name:          "Valid Rule",
		Description:   "A valid performance alert rule",
		Severity:      "warning",
		Category:      "performance",
		Enabled:       true,
		MetricType:    "response_time",
		Condition:     "threshold",
		Duration:      5 * time.Minute,
		Threshold:     500.0,
		Operator:      "gt",
		Notifications: []string{"email", "slack"},
	}

	err := pre.ValidateRule(validRule)
	assert.NoError(t, err)

	// Invalid rule - missing ID
	invalidRule := &PerformanceAlertRule{
		Name:          "Invalid Rule",
		Description:   "An invalid performance alert rule",
		Severity:      "warning",
		Category:      "performance",
		Enabled:       true,
		MetricType:    "response_time",
		Condition:     "threshold",
		Duration:      5 * time.Minute,
		Threshold:     500.0,
		Operator:      "gt",
		Notifications: []string{"email", "slack"},
	}

	err = pre.ValidateRule(invalidRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule ID is required")

	// Invalid rule - invalid metric type
	invalidRule.ID = "invalid_metric_rule"
	invalidRule.MetricType = "invalid_metric"
	err = pre.ValidateRule(invalidRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid metric type")

	// Invalid rule - invalid condition
	invalidRule.MetricType = "response_time"
	invalidRule.Condition = "invalid_condition"
	err = pre.ValidateRule(invalidRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid condition")

	// Invalid rule - invalid operator
	invalidRule.Condition = "threshold"
	invalidRule.Operator = "invalid_operator"
	err = pre.ValidateRule(invalidRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid operator")

	// Invalid rule - invalid severity
	invalidRule.Operator = "gt"
	invalidRule.Severity = "invalid_severity"
	err = pre.ValidateRule(invalidRule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid severity")
}

func TestPerformanceRuleEngine_GetMetricValue(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}
	rules := make(map[string]*PerformanceAlertRule)

	pre := NewPerformanceRuleEngine(rules, config, logger)

	metrics := &PerformanceMetrics{
		AverageResponseTime: 250 * time.Millisecond,
		P95ResponseTime:     500 * time.Millisecond,
		P99ResponseTime:     1000 * time.Millisecond,
		OverallSuccessRate:  0.98,
		RequestsPerSecond:   1000.0,
		CPUUsage:            75.0,
		MemoryUsage:         80.0,
		DiskUsage:           85.0,
		ErrorRate:           0.02,
		Availability:        0.999,
	}

	// Test response_time metric
	value := pre.getMetricValue("response_time", metrics)
	assert.Equal(t, 250.0, value)

	// Test success_rate metric
	value = pre.getMetricValue("success_rate", metrics)
	assert.Equal(t, 0.98, value)

	// Test throughput metric
	value = pre.getMetricValue("throughput", metrics)
	assert.Equal(t, 1000.0, value)

	// Test cpu metric
	value = pre.getMetricValue("cpu", metrics)
	assert.Equal(t, 75.0, value)

	// Test memory metric
	value = pre.getMetricValue("memory", metrics)
	assert.Equal(t, 80.0, value)

	// Test disk metric
	value = pre.getMetricValue("disk", metrics)
	assert.Equal(t, 85.0, value)

	// Test error_rate metric
	value = pre.getMetricValue("error_rate", metrics)
	assert.Equal(t, 0.02, value)

	// Test availability metric
	value = pre.getMetricValue("availability", metrics)
	assert.Equal(t, 0.999, value)

	// Test latency_p95 metric
	value = pre.getMetricValue("latency_p95", metrics)
	assert.Equal(t, 500.0, value)

	// Test latency_p99 metric
	value = pre.getMetricValue("latency_p99", metrics)
	assert.Equal(t, 1000.0, value)

	// Test invalid metric type
	value = pre.getMetricValue("invalid_metric", metrics)
	assert.Equal(t, -1.0, value)
}

func TestPerformanceRuleEngine_EvaluateThresholdCondition(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}
	rules := make(map[string]*PerformanceAlertRule)

	pre := NewPerformanceRuleEngine(rules, config, logger)

	rule := &PerformanceAlertRule{
		Threshold: 100.0,
	}

	// Test greater than
	rule.Operator = "gt"
	assert.True(t, pre.evaluateThresholdCondition(rule, 150.0))
	assert.False(t, pre.evaluateThresholdCondition(rule, 50.0))
	assert.False(t, pre.evaluateThresholdCondition(rule, 100.0))

	// Test greater than or equal
	rule.Operator = "gte"
	assert.True(t, pre.evaluateThresholdCondition(rule, 150.0))
	assert.False(t, pre.evaluateThresholdCondition(rule, 50.0))
	assert.True(t, pre.evaluateThresholdCondition(rule, 100.0))

	// Test less than
	rule.Operator = "lt"
	assert.False(t, pre.evaluateThresholdCondition(rule, 150.0))
	assert.True(t, pre.evaluateThresholdCondition(rule, 50.0))
	assert.False(t, pre.evaluateThresholdCondition(rule, 100.0))

	// Test less than or equal
	rule.Operator = "lte"
	assert.False(t, pre.evaluateThresholdCondition(rule, 150.0))
	assert.True(t, pre.evaluateThresholdCondition(rule, 50.0))
	assert.True(t, pre.evaluateThresholdCondition(rule, 100.0))

	// Test equal
	rule.Operator = "eq"
	assert.False(t, pre.evaluateThresholdCondition(rule, 150.0))
	assert.False(t, pre.evaluateThresholdCondition(rule, 50.0))
	assert.True(t, pre.evaluateThresholdCondition(rule, 100.0))

	// Test not equal
	rule.Operator = "ne"
	assert.True(t, pre.evaluateThresholdCondition(rule, 150.0))
	assert.True(t, pre.evaluateThresholdCondition(rule, 50.0))
	assert.False(t, pre.evaluateThresholdCondition(rule, 100.0))
}

func TestAlertEscalationManager_StartStopEscalation(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}

	aem := NewAlertEscalationManager(config, logger)

	alert := &PerformanceAlert{
		ID:       "test_alert",
		RuleID:   "test_rule",
		RuleName: "Test Rule",
		Status:   "firing",
	}

	policy := &EscalationPolicy{
		ID:              "test_policy",
		Name:            "Test Policy",
		Description:     "Test escalation policy",
		MaxEscalations:  2,
		EscalationDelay: 5 * time.Minute,
		Levels: []EscalationLevel{
			{
				Level:         1,
				Delay:         1 * time.Minute,
				Notifications: []string{"email", "slack"},
				Recipients:    []string{"test@example.com"},
			},
			{
				Level:         2,
				Delay:         5 * time.Minute,
				Notifications: []string{"email", "slack", "pagerduty"},
				Recipients:    []string{"test@example.com", "pagerduty"},
			},
		},
	}

	// Start escalation
	aem.StartEscalation(alert, policy)

	// Check that escalation was created
	escalation, exists := aem.GetEscalationForAlert(alert.ID)
	assert.True(t, exists)
	assert.Equal(t, "active", escalation.Status)
	assert.Equal(t, 1, escalation.Level)

	// Try to start escalation again (should be ignored)
	aem.StartEscalation(alert, policy)

	// Stop escalation
	aem.StopEscalation(alert.ID)

	// Check that escalation was stopped
	escalation, exists = aem.GetEscalationForAlert(alert.ID)
	assert.False(t, exists)
}

func TestAlertEscalationManager_AddEscalationPolicy(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}

	aem := NewAlertEscalationManager(config, logger)

	policy := &EscalationPolicy{
		ID:              "test_policy",
		Name:            "Test Policy",
		Description:     "Test escalation policy",
		MaxEscalations:  2,
		EscalationDelay: 5 * time.Minute,
		Levels: []EscalationLevel{
			{
				Level:         1,
				Delay:         1 * time.Minute,
				Notifications: []string{"email", "slack"},
				Recipients:    []string{"test@example.com"},
			},
			{
				Level:         2,
				Delay:         5 * time.Minute,
				Notifications: []string{"email", "slack", "pagerduty"},
				Recipients:    []string{"test@example.com", "pagerduty"},
			},
		},
	}

	// Add policy
	err := aem.AddEscalationPolicy(policy)
	assert.NoError(t, err)

	// Try to add the same policy again
	err = aem.AddEscalationPolicy(policy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Get policy
	retrievedPolicy, exists := aem.GetEscalationPolicy("test_policy")
	assert.True(t, exists)
	assert.Equal(t, policy.ID, retrievedPolicy.ID)
	assert.Equal(t, policy.Name, retrievedPolicy.Name)
}

func TestAlertEscalationManager_ValidateEscalationPolicy(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}

	aem := NewAlertEscalationManager(config, logger)

	// Valid policy
	validPolicy := &EscalationPolicy{
		ID:              "valid_policy",
		Name:            "Valid Policy",
		Description:     "A valid escalation policy",
		MaxEscalations:  2,
		EscalationDelay: 5 * time.Minute,
		Levels: []EscalationLevel{
			{
				Level:         1,
				Delay:         1 * time.Minute,
				Notifications: []string{"email", "slack"},
				Recipients:    []string{"test@example.com"},
			},
			{
				Level:         2,
				Delay:         5 * time.Minute,
				Notifications: []string{"email", "slack", "pagerduty"},
				Recipients:    []string{"test@example.com", "pagerduty"},
			},
		},
	}

	err := aem.validateEscalationPolicy(validPolicy)
	assert.NoError(t, err)

	// Invalid policy - missing ID
	invalidPolicy := &EscalationPolicy{
		Name:            "Invalid Policy",
		Description:     "An invalid escalation policy",
		MaxEscalations:  2,
		EscalationDelay: 5 * time.Minute,
		Levels: []EscalationLevel{
			{
				Level:         1,
				Delay:         1 * time.Minute,
				Notifications: []string{"email", "slack"},
				Recipients:    []string{"test@example.com"},
			},
		},
	}

	err = aem.validateEscalationPolicy(invalidPolicy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy ID is required")

	// Invalid policy - missing name
	invalidPolicy.ID = "invalid_policy"
	invalidPolicy.Name = ""
	err = aem.validateEscalationPolicy(invalidPolicy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy name is required")

	// Invalid policy - no levels
	invalidPolicy.Name = "Invalid Policy"
	invalidPolicy.Levels = []EscalationLevel{}
	err = aem.validateEscalationPolicy(invalidPolicy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have at least one escalation level")

	// Invalid policy - invalid notification type
	invalidPolicy.Levels = []EscalationLevel{
		{
			Level:         1,
			Delay:         1 * time.Minute,
			Notifications: []string{"invalid_notification"},
			Recipients:    []string{"test@example.com"},
		},
	}
	err = aem.validateEscalationPolicy(invalidPolicy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid notification type")

	// Invalid policy - non-sequential levels
	invalidPolicy.Levels = []EscalationLevel{
		{
			Level:         1,
			Delay:         1 * time.Minute,
			Notifications: []string{"email", "slack"},
			Recipients:    []string{"test@example.com"},
		},
		{
			Level:         3, // Should be 2
			Delay:         5 * time.Minute,
			Notifications: []string{"email", "slack", "pagerduty"},
			Recipients:    []string{"test@example.com", "pagerduty"},
		},
	}
	err = aem.validateEscalationPolicy(invalidPolicy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be sequential")
}

func TestAlertEscalationManager_GetEscalationStatistics(t *testing.T) {
	logger := zap.NewNop()
	config := PerformanceAlertingConfig{}

	aem := NewAlertEscalationManager(config, logger)

	// Add a policy
	policy := &EscalationPolicy{
		ID:              "test_policy",
		Name:            "Test Policy",
		Description:     "Test escalation policy",
		MaxEscalations:  2,
		EscalationDelay: 5 * time.Minute,
		Levels: []EscalationLevel{
			{
				Level:         1,
				Delay:         1 * time.Minute,
				Notifications: []string{"email", "slack"},
				Recipients:    []string{"test@example.com"},
			},
			{
				Level:         2,
				Delay:         5 * time.Minute,
				Notifications: []string{"email", "slack", "pagerduty"},
				Recipients:    []string{"test@example.com", "pagerduty"},
			},
		},
	}
	aem.AddEscalationPolicy(policy)

	// Add an active escalation
	alert := &PerformanceAlert{
		ID:       "test_alert",
		RuleID:   "test_rule",
		RuleName: "Test Rule",
		Status:   "firing",
	}
	aem.StartEscalation(alert, policy)

	// Get statistics
	stats := aem.GetEscalationStatistics()
	assert.Equal(t, 1, stats.TotalPolicies)
	assert.Equal(t, 1, stats.ActiveEscalations)
	assert.Len(t, stats.PolicyStats, 1)

	policyStats, exists := stats.PolicyStats["test_policy"]
	assert.True(t, exists)
	assert.Equal(t, "test_policy", policyStats.PolicyID)
	assert.Equal(t, "Test Policy", policyStats.PolicyName)
	assert.Equal(t, 2, policyStats.MaxLevels)
	assert.Equal(t, 1, policyStats.ActiveEscalations)
}
