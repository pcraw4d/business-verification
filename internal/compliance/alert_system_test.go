package compliance

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

func TestNewAlertSystem(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := &ComplianceStatusSystem{}
	checkEngine := &CheckEngine{}

	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	if alertSystem == nil {
		t.Fatal("Expected non-nil alert system")
	}

	if alertSystem.logger != logger {
		t.Error("Logger not set correctly")
	}

	if alertSystem.statusSystem != statusSystem {
		t.Error("StatusSystem not set correctly")
	}

	if alertSystem.checkEngine != checkEngine {
		t.Error("CheckEngine not set correctly")
	}

	if alertSystem.rules == nil {
		t.Error("Rules map not initialized")
	}

	if alertSystem.escalations == nil {
		t.Error("Escalations map not initialized")
	}

	if alertSystem.notifications == nil {
		t.Error("Notifications map not initialized")
	}
}

func TestRegisterAlertRule(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := &ComplianceStatusSystem{}
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")

	rule := &AlertRule{
		ID:          "test-rule-1",
		Name:        "Test Rule",
		Description: "A test alert rule",
		Enabled:     true,
		EntityType:  "overall",
		Severity:    "high",
		Conditions: []AlertCondition{
			{
				Type:     "score_below",
				Field:    "overall",
				Operator: "lt",
				Value:    80.0,
			},
		},
		Actions: []AlertAction{
			{
				Type: "create_alert",
			},
		},
	}

	err := alertSystem.RegisterAlertRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register alert rule: %v", err)
	}

	// Verify rule was registered
	registeredRule, err := alertSystem.GetAlertRule(ctx, "test-rule-1")
	if err != nil {
		t.Fatalf("Failed to get registered rule: %v", err)
	}

	if registeredRule.ID != "test-rule-1" {
		t.Errorf("Expected rule ID 'test-rule-1', got '%s'", registeredRule.ID)
	}

	if registeredRule.Name != "Test Rule" {
		t.Errorf("Expected rule name 'Test Rule', got '%s'", registeredRule.Name)
	}

	if !registeredRule.Enabled {
		t.Error("Expected rule to be enabled")
	}

	if registeredRule.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if registeredRule.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestUpdateAlertRule(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := &ComplianceStatusSystem{}
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")

	// Register initial rule
	rule := &AlertRule{
		ID:          "test-rule-2",
		Name:        "Original Name",
		Description: "Original description",
		Enabled:     true,
		EntityType:  "overall",
		Severity:    "medium",
	}

	err := alertSystem.RegisterAlertRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register alert rule: %v", err)
	}

	// Update rule
	updates := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated description",
		"enabled":     false,
		"severity":    "high",
	}

	err = alertSystem.UpdateAlertRule(ctx, "test-rule-2", updates)
	if err != nil {
		t.Fatalf("Failed to update alert rule: %v", err)
	}

	// Verify updates
	updatedRule, err := alertSystem.GetAlertRule(ctx, "test-rule-2")
	if err != nil {
		t.Fatalf("Failed to get updated rule: %v", err)
	}

	if updatedRule.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updatedRule.Name)
	}

	if updatedRule.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got '%s'", updatedRule.Description)
	}

	if updatedRule.Enabled {
		t.Error("Expected rule to be disabled")
	}

	if updatedRule.Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", updatedRule.Severity)
	}
}

func TestDeleteAlertRule(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := &ComplianceStatusSystem{}
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")

	// Register rule
	rule := &AlertRule{
		ID:   "test-rule-3",
		Name: "Test Rule",
	}

	err := alertSystem.RegisterAlertRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register alert rule: %v", err)
	}

	// Delete rule
	err = alertSystem.DeleteAlertRule(ctx, "test-rule-3")
	if err != nil {
		t.Fatalf("Failed to delete alert rule: %v", err)
	}

	// Verify rule was deleted
	_, err = alertSystem.GetAlertRule(ctx, "test-rule-3")
	if err == nil {
		t.Error("Expected error when getting deleted rule")
	}
}

func TestListAlertRules(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := &ComplianceStatusSystem{}
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")

	// Register multiple rules
	rules := []*AlertRule{
		{ID: "rule-1", Name: "Rule 1", Enabled: true},
		{ID: "rule-2", Name: "Rule 2", Enabled: true},
		{ID: "rule-3", Name: "Rule 3", Enabled: false},
	}

	for _, rule := range rules {
		err := alertSystem.RegisterAlertRule(ctx, rule)
		if err != nil {
			t.Fatalf("Failed to register rule %s: %v", rule.ID, err)
		}
	}

	// List rules
	listedRules, err := alertSystem.ListAlertRules(ctx)
	if err != nil {
		t.Fatalf("Failed to list alert rules: %v", err)
	}

	if len(listedRules) != 3 {
		t.Errorf("Expected 3 rules, got %d", len(listedRules))
	}

	// Verify all rules are present
	ruleIDs := make(map[string]bool)
	for _, rule := range listedRules {
		ruleIDs[rule.ID] = true
	}

	for _, expectedID := range []string{"rule-1", "rule-2", "rule-3"} {
		if !ruleIDs[expectedID] {
			t.Errorf("Expected rule %s not found in list", expectedID)
		}
	}
}

func TestEvaluateAlerts(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := NewComplianceStatusSystem(logger)
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")

	// Initialize business status first
	err := statusSystem.InitializeBusinessStatus(ctx, "test-business")
	if err != nil {
		t.Fatalf("Failed to initialize business status: %v", err)
	}

	// Register a test rule
	rule := &AlertRule{
		ID:         "test-eval-rule",
		Name:       "Test Evaluation Rule",
		Enabled:    true,
		EntityType: "overall",
		Severity:   "high",
		Conditions: []AlertCondition{
			{
				Type:     "score_below",
				Field:    "overall",
				Operator: "lt",
				Value:    80.0,
			},
		},
		Actions: []AlertAction{
			{
				Type: "create_alert",
			},
		},
	}

	err = alertSystem.RegisterAlertRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register alert rule: %v", err)
	}

	// Evaluate alerts
	evaluations, err := alertSystem.EvaluateAlerts(ctx, "test-business")
	if err != nil {
		t.Fatalf("Failed to evaluate alerts: %v", err)
	}

	if len(evaluations) != 1 {
		t.Errorf("Expected 1 evaluation, got %d", len(evaluations))
	}

	evaluation := evaluations[0]
	if evaluation.RuleID != "test-eval-rule" {
		t.Errorf("Expected rule ID 'test-eval-rule', got '%s'", evaluation.RuleID)
	}

	if evaluation.BusinessID != "test-business" {
		t.Errorf("Expected business ID 'test-business', got '%s'", evaluation.BusinessID)
	}

	if evaluation.EntityType != "overall" {
		t.Errorf("Expected entity type 'overall', got '%s'", evaluation.EntityType)
	}

	if evaluation.EvaluatedAt.IsZero() {
		t.Error("Expected EvaluatedAt to be set")
	}
}

func TestEvaluateOperator(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := &ComplianceStatusSystem{}
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	tests := []struct {
		name      string
		current   interface{}
		operator  string
		threshold interface{}
		expected  bool
	}{
		{"lt_true", 50.0, "lt", 80.0, true},
		{"lt_false", 90.0, "lt", 80.0, false},
		{"lte_true", 80.0, "lte", 80.0, true},
		{"lte_false", 90.0, "lte", 80.0, false},
		{"eq_true", 80.0, "eq", 80.0, true},
		{"eq_false", 90.0, "eq", 80.0, false},
		{"gte_true", 90.0, "gte", 80.0, true},
		{"gte_false", 70.0, "gte", 80.0, false},
		{"gt_true", 90.0, "gt", 80.0, true},
		{"gt_false", 80.0, "gt", 80.0, false},
		{"ne_true", 90.0, "ne", 80.0, true},
		{"ne_false", 80.0, "ne", 80.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := alertSystem.evaluateOperator(tt.current, tt.operator, tt.threshold)
			if result != tt.expected {
				t.Errorf("evaluateOperator(%v, %s, %v) = %v, want %v",
					tt.current, tt.operator, tt.threshold, result, tt.expected)
			}
		})
	}
}

func TestCompareValues(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := &ComplianceStatusSystem{}
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected int
	}{
		{"float64_lt", 50.0, 80.0, -1},
		{"float64_eq", 80.0, 80.0, 0},
		{"float64_gt", 90.0, 80.0, 1},
		{"int_lt", 50, 80, -1},
		{"int_eq", 80, 80, 0},
		{"int_gt", 90, 80, 1},
		{"string_lt", "abc", "def", -1},
		{"string_eq", "abc", "abc", 0},
		{"string_gt", "def", "abc", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := alertSystem.compareValues(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareValues(%v, %v) = %d, want %d",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := &ComplianceStatusSystem{}
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	tests := []struct {
		name     string
		input    interface{}
		expected float64
		ok       bool
	}{
		{"float64", 50.0, 50.0, true},
		{"float32", float32(50.0), 50.0, true},
		{"int", 50, 50.0, true},
		{"int32", int32(50), 50.0, true},
		{"int64", int64(50), 50.0, true},
		{"string", "50", 0.0, false},
		{"bool", true, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := alertSystem.toFloat64(tt.input)
			if ok != tt.ok {
				t.Errorf("toFloat64(%v) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && result != tt.expected {
				t.Errorf("toFloat64(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetAlertAnalytics(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := NewComplianceStatusSystem(logger)
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")

	// Initialize business status first
	err := statusSystem.InitializeBusinessStatus(ctx, "test-business")
	if err != nil {
		t.Fatalf("Failed to initialize business status: %v", err)
	}

	// Test analytics generation
	analytics, err := alertSystem.GetAlertAnalytics(ctx, "test-business", "7d")
	if err != nil {
		t.Fatalf("Failed to get alert analytics: %v", err)
	}

	if analytics == nil {
		t.Fatal("Expected non-nil analytics")
	}

	if analytics.BusinessID != "test-business" {
		t.Errorf("Expected business ID 'test-business', got '%s'", analytics.BusinessID)
	}

	if analytics.Period != "7d" {
		t.Errorf("Expected period '7d', got '%s'", analytics.Period)
	}

	if analytics.GeneratedAt.IsZero() {
		t.Error("Expected GeneratedAt to be set")
	}

	if analytics.AlertsBySeverity == nil {
		t.Error("Expected AlertsBySeverity to be initialized")
	}

	if analytics.AlertsByType == nil {
		t.Error("Expected AlertsByType to be initialized")
	}

	if analytics.AlertsByEntity == nil {
		t.Error("Expected AlertsByEntity to be initialized")
	}
}

func TestRegisterEscalationPolicy(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := &ComplianceStatusSystem{}
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")

	policy := &EscalationPolicy{
		ID:          "test-policy-1",
		Name:        "Test Escalation Policy",
		Description: "A test escalation policy",
		Enabled:     true,
		Levels: []EscalationLevel{
			{
				Level:      1,
				Name:       "Level 1",
				Delay:      1 * time.Hour,
				Recipients: []string{"user1@example.com"},
				Actions:    []string{"email", "slack"},
			},
		},
	}

	err := alertSystem.RegisterEscalationPolicy(ctx, policy)
	if err != nil {
		t.Fatalf("Failed to register escalation policy: %v", err)
	}

	// Verify policy was registered (we'd need a getter method to verify)
	// For now, just check that no error was returned
}

func TestRegisterNotificationChannel(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := &ComplianceStatusSystem{}
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")

	channel := &NotificationChannel{
		ID:      "test-channel-1",
		Name:    "Test Email Channel",
		Type:    "email",
		Enabled: true,
		Config: map[string]interface{}{
			"smtp_server": "smtp.example.com",
			"smtp_port":   587,
		},
		Recipients: []string{"admin@example.com"},
	}

	err := alertSystem.RegisterNotificationChannel(ctx, channel)
	if err != nil {
		t.Fatalf("Failed to register notification channel: %v", err)
	}

	// Verify channel was registered (we'd need a getter method to verify)
	// For now, just check that no error was returned
}

func TestAlertRuleValidation(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{})
	statusSystem := NewComplianceStatusSystem(logger)
	checkEngine := &CheckEngine{}
	alertSystem := NewAlertSystem(logger, statusSystem, checkEngine)

	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")

	// Initialize business status first
	err := statusSystem.InitializeBusinessStatus(ctx, "test-business")
	if err != nil {
		t.Fatalf("Failed to initialize business status: %v", err)
	}

	// Test rule with invalid condition type
	rule := &AlertRule{
		ID:         "test-invalid-rule",
		Name:       "Test Invalid Rule",
		Enabled:    true,
		EntityType: "overall",
		Severity:   "high",
		Conditions: []AlertCondition{
			{
				Type:     "invalid_type",
				Field:    "overall",
				Operator: "lt",
				Value:    80.0,
			},
		},
		Actions: []AlertAction{
			{
				Type: "create_alert",
			},
		},
	}

	err = alertSystem.RegisterAlertRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register alert rule: %v", err)
	}

	// Evaluate alerts - should handle invalid condition gracefully
	evaluations, err := alertSystem.EvaluateAlerts(ctx, "test-business")
	if err != nil {
		t.Fatalf("Failed to evaluate alerts: %v", err)
	}

	// Should return evaluations, even if some rules fail
	if len(evaluations) != 1 {
		t.Errorf("Expected 1 evaluation, got %d", len(evaluations))
	}

	// The evaluation should exist but not be triggered due to invalid condition
	evaluation := evaluations[0]
	if evaluation.RuleID != "test-invalid-rule" {
		t.Errorf("Expected rule ID 'test-invalid-rule', got '%s'", evaluation.RuleID)
	}

	if evaluation.Triggered {
		t.Error("Expected evaluation to not be triggered due to invalid condition")
	}
}
