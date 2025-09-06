package risk

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.uber.org/zap"
)

// createTestLogger creates a logger for testing
func createTestLogger() *observability.Logger {
	zapLogger, _ := zap.NewDevelopment()
	return observability.NewLogger(zapLogger)
}

func TestNewAlertService(t *testing.T) {
	logger := createTestLogger()
	thresholdManager := CreateDefaultThresholds()

	alertService := NewAlertService(logger, thresholdManager)

	if alertService == nil {
		t.Fatal("Expected AlertService to be created")
	}

	if alertService.logger == nil {
		t.Error("Expected logger to be set")
	}

	if alertService.thresholdManager == nil {
		t.Error("Expected threshold manager to be set")
	}
}

func TestAlertService_GenerateAlerts(t *testing.T) {
	logger := createTestLogger()
	thresholdManager := CreateDefaultThresholds()
	alertService := NewAlertService(logger, thresholdManager)

	// Create a test assessment
	assessment := &RiskAssessment{
		ID:           "test_assessment",
		BusinessID:   "test_business",
		BusinessName: "Test Business",
		OverallScore: 85.0,
		OverallLevel: RiskLevelCritical,
		CategoryScores: map[RiskCategory]RiskScore{
			RiskCategoryFinancial: {
				FactorID:   "financial",
				FactorName: "Financial Risk",
				Category:   RiskCategoryFinancial,
				Score:      80.0,
				Level:      RiskLevelHigh,
				Confidence: 0.9,
			},
			RiskCategoryOperational: {
				FactorID:   "operational",
				FactorName: "Operational Risk",
				Category:   RiskCategoryOperational,
				Score:      75.0,
				Level:      RiskLevelHigh,
				Confidence: 0.8,
			},
		},
		FactorScores: []RiskScore{
			{
				FactorID:   "financial_stability",
				FactorName: "Financial Stability",
				Category:   RiskCategoryFinancial,
				Score:      85.0,
				Level:      RiskLevelCritical,
				Confidence: 0.9,
			},
			{
				FactorID:   "operational_efficiency",
				FactorName: "Operational Efficiency",
				Category:   RiskCategoryOperational,
				Score:      70.0,
				Level:      RiskLevelHigh,
				Confidence: 0.8,
			},
		},
		AssessedAt: time.Now(),
		ValidUntil: time.Now().Add(24 * time.Hour),
	}

	ctx := context.WithValue(context.Background(), "request_id", "test_request")

	alerts, err := alertService.GenerateAlerts(ctx, assessment)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(alerts) == 0 {
		t.Error("Expected alerts to be generated for high-risk assessment")
	}

	// Check for specific alert types
	foundOverallAlert := false
	foundFactorAlert := false

	for _, alert := range alerts {
		if alert.RiskFactor == "overall_risk" {
			foundOverallAlert = true
		}
		if alert.RiskFactor == "financial_stability" {
			foundFactorAlert = true
		}
	}

	if !foundOverallAlert {
		t.Error("Expected overall risk alert to be generated")
	}

	if !foundFactorAlert {
		t.Error("Expected factor-specific alert to be generated")
	}
}

func TestAlertService_GenerateFactorAlerts(t *testing.T) {
	logger := createTestLogger()
	thresholdManager := CreateDefaultThresholds()
	alertService := NewAlertService(logger, thresholdManager)

	assessment := &RiskAssessment{
		ID:           "test_assessment",
		BusinessID:   "test_business",
		BusinessName: "Test Business",
		FactorScores: []RiskScore{
			{
				FactorID:   "test_factor",
				FactorName: "Test Factor",
				Category:   RiskCategoryFinancial,
				Score:      85.0,
				Level:      RiskLevelCritical,
				Confidence: 0.9,
			},
		},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test_request")

	alerts, err := alertService.generateFactorAlerts(ctx, assessment)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(alerts) == 0 {
		t.Error("Expected factor alerts to be generated")
	}

	// Check that the alert has the correct properties
	alert := alerts[0]
	if alert.Level != RiskLevelCritical {
		t.Errorf("Expected alert level to be Critical, got: %s", alert.Level)
	}

	if alert.Score != 85.0 {
		t.Errorf("Expected alert score to be 85.0, got: %.1f", alert.Score)
	}
}

func TestAlertService_GenerateCategoryAlerts(t *testing.T) {
	logger := createTestLogger()
	thresholdManager := CreateDefaultThresholds()
	alertService := NewAlertService(logger, thresholdManager)

	assessment := &RiskAssessment{
		ID:           "test_assessment",
		BusinessID:   "test_business",
		BusinessName: "Test Business",
		CategoryScores: map[RiskCategory]RiskScore{
			RiskCategoryFinancial: {
				FactorID:   "financial",
				FactorName: "Financial Risk",
				Category:   RiskCategoryFinancial,
				Score:      80.0,
				Level:      RiskLevelHigh,
				Confidence: 0.9,
			},
		},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test_request")

	alerts, err := alertService.generateCategoryAlerts(ctx, assessment)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(alerts) == 0 {
		t.Error("Expected category alerts to be generated")
	}

	// Check that the alert has the correct properties
	alert := alerts[0]
	if alert.Level != RiskLevelHigh {
		t.Errorf("Expected alert level to be High, got: %s", alert.Level)
	}

	if alert.Score != 80.0 {
		t.Errorf("Expected alert score to be 80.0, got: %.1f", alert.Score)
	}
}

func TestAlertService_GenerateOverallAlerts(t *testing.T) {
	logger := createTestLogger()
	thresholdManager := CreateDefaultThresholds()
	alertService := NewAlertService(logger, thresholdManager)

	assessment := &RiskAssessment{
		ID:           "test_assessment",
		BusinessID:   "test_business",
		BusinessName: "Test Business",
		OverallScore: 85.0,
		OverallLevel: RiskLevelCritical,
		FactorScores: []RiskScore{
			{
				FactorID:   "factor1",
				FactorName: "Factor 1",
				Category:   RiskCategoryFinancial,
				Score:      75.0,
				Level:      RiskLevelHigh,
				Confidence: 0.9,
			},
			{
				FactorID:   "factor2",
				FactorName: "Factor 2",
				Category:   RiskCategoryOperational,
				Score:      80.0,
				Level:      RiskLevelHigh,
				Confidence: 0.8,
			},
			{
				FactorID:   "factor3",
				FactorName: "Factor 3",
				Category:   RiskCategoryRegulatory,
				Score:      85.0,
				Level:      RiskLevelCritical,
				Confidence: 0.9,
			},
		},
	}

	ctx := context.WithValue(context.Background(), "request_id", "test_request")

	alerts, err := alertService.generateOverallAlerts(ctx, assessment)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(alerts) == 0 {
		t.Error("Expected overall alerts to be generated")
	}

	// Check for overall risk alert
	foundOverallAlert := false
	for _, alert := range alerts {
		if alert.RiskFactor == "overall_risk" {
			foundOverallAlert = true
			if alert.Level != RiskLevelCritical {
				t.Errorf("Expected overall alert level to be Critical, got: %s", alert.Level)
			}
		}
	}

	if !foundOverallAlert {
		t.Error("Expected overall risk alert to be generated")
	}
}

func TestAlertService_GetThresholdForFactor(t *testing.T) {
	logger := createTestLogger()
	thresholdManager := CreateDefaultThresholds()
	alertService := NewAlertService(logger, thresholdManager)

	tests := []struct {
		factorID string
		category RiskCategory
		expected float64
	}{
		{"financial_factor", RiskCategoryFinancial, 70.0},
		{"operational_factor", RiskCategoryOperational, 65.0},
		{"regulatory_factor", RiskCategoryRegulatory, 80.0},
		{"reputational_factor", RiskCategoryReputational, 75.0},
		{"cybersecurity_factor", RiskCategoryCybersecurity, 85.0},
		{"unknown_factor", RiskCategoryOperational, 65.0}, // default for operational
	}

	for _, test := range tests {
		threshold := alertService.getThresholdForFactor(test.factorID, test.category)
		if threshold != test.expected {
			t.Errorf("For factor %s in category %s, expected threshold %.1f, got %.1f",
				test.factorID, test.category, test.expected, threshold)
		}
	}
}

func TestAlertService_GetThresholdForCategory(t *testing.T) {
	logger := createTestLogger()
	thresholdManager := CreateDefaultThresholds()
	alertService := NewAlertService(logger, thresholdManager)

	tests := []struct {
		category RiskCategory
		expected float64
	}{
		{RiskCategoryFinancial, 70.0},
		{RiskCategoryOperational, 65.0},
		{RiskCategoryRegulatory, 80.0},
		{RiskCategoryReputational, 75.0},
		{RiskCategoryCybersecurity, 85.0},
	}

	for _, test := range tests {
		threshold := alertService.getThresholdForCategory(test.category)
		if threshold != test.expected {
			t.Errorf("For category %s, expected threshold %.1f, got %.1f",
				test.category, test.expected, threshold)
		}
	}
}

func TestAlertService_CreateAlertRule(t *testing.T) {
	logger := createTestLogger()
	thresholdManager := CreateDefaultThresholds()
	alertService := NewAlertService(logger, thresholdManager)

	rule := &AlertRule{
		Name:        "Test Rule",
		Description: "Test alert rule",
		Category:    RiskCategoryFinancial,
		Condition:   AlertConditionGreaterThan,
		Threshold:   75.0,
		Level:       RiskLevelHigh,
		Message:     "Test alert message",
		Enabled:     true,
	}

	err := alertService.CreateAlertRule(rule)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if rule.ID == "" {
		t.Error("Expected rule ID to be set")
	}

	if rule.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if rule.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestAlertService_GetAlertRules(t *testing.T) {
	logger := createTestLogger()
	thresholdManager := CreateDefaultThresholds()
	alertService := NewAlertService(logger, thresholdManager)

	rules, err := alertService.GetAlertRules()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(rules) == 0 {
		t.Error("Expected default alert rules to be returned")
	}

	// Check that default rules have expected properties
	for _, rule := range rules {
		if rule.ID == "" {
			t.Error("Expected rule ID to be set")
		}

		if rule.Name == "" {
			t.Error("Expected rule name to be set")
		}

		if rule.Threshold <= 0 {
			t.Error("Expected rule threshold to be positive")
		}
	}
}

func TestAlertCondition_Constants(t *testing.T) {
	// Test that alert condition constants are properly defined
	conditions := []AlertCondition{
		AlertConditionGreaterThan,
		AlertConditionLessThan,
		AlertConditionEquals,
		AlertConditionNotEquals,
		AlertConditionIncreasesBy,
		AlertConditionDecreasesBy,
		AlertConditionCrossesAbove,
		AlertConditionCrossesBelow,
	}

	for _, condition := range conditions {
		if string(condition) == "" {
			t.Errorf("Alert condition constant should not be empty: %v", condition)
		}
	}
}

func TestAlertRule_Validation(t *testing.T) {
	// Test alert rule creation and validation
	rule := &AlertRule{
		ID:          "test_rule",
		Name:        "Test Rule",
		Description: "Test alert rule",
		Category:    RiskCategoryFinancial,
		Condition:   AlertConditionGreaterThan,
		Threshold:   75.0,
		Level:       RiskLevelHigh,
		Message:     "Test alert message",
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if rule.ID == "" {
		t.Error("Expected rule ID to be set")
	}

	if rule.Name == "" {
		t.Error("Expected rule name to be set")
	}

	if rule.Threshold <= 0 {
		t.Error("Expected rule threshold to be positive")
	}

	if string(rule.Category) == "" {
		t.Error("Expected rule category to be set")
	}

	if string(rule.Condition) == "" {
		t.Error("Expected rule condition to be set")
	}

	if string(rule.Level) == "" {
		t.Error("Expected rule level to be set")
	}
}

func TestAlertNotification_Validation(t *testing.T) {
	// Test alert notification creation and validation
	now := time.Now()
	notification := &AlertNotification{
		ID:         "test_notification",
		AlertID:    "test_alert",
		Type:       "email",
		Recipient:  "test@example.com",
		Message:    "Test notification message",
		Status:     "pending",
		SentAt:     &now,
		RetryCount: 0,
		CreatedAt:  now,
	}

	if notification.ID == "" {
		t.Error("Expected notification ID to be set")
	}

	if notification.AlertID == "" {
		t.Error("Expected alert ID to be set")
	}

	if notification.Type == "" {
		t.Error("Expected notification type to be set")
	}

	if notification.Recipient == "" {
		t.Error("Expected notification recipient to be set")
	}

	if notification.Message == "" {
		t.Error("Expected notification message to be set")
	}

	if notification.Status == "" {
		t.Error("Expected notification status to be set")
	}
}
