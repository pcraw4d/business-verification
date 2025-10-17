package compliance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"kyb-platform/internal/compliance"
	"kyb-platform/internal/observability"
	"go.uber.org/zap"
)

// TestAlertSystemFunctionality tests the alert system functionality
func TestAlertSystemFunctionality(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)
	alertService := compliance.NewComplianceAlertService(logger, frameworkService, trackingService)

	ctx := context.Background()
	businessID := "alert-test-business"
	frameworkID := "SOC2"

	t.Run("Alert Creation and Retrieval", func(t *testing.T) {
		// Create a test alert
		alert := &compliance.ComplianceAlert{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			AlertType:   "compliance_change",
			Severity:    "high",
			Title:       "Test Compliance Alert",
			Description: "Test alert for compliance change",
			Message:     "Compliance score has changed significantly",
			TriggeredBy: "test",
		}

		// Create alert
		err := alertService.CreateAlert(ctx, alert)
		if err != nil {
			t.Fatalf("Failed to create alert: %v", err)
		}

		// Verify alert was created
		if alert.ID == "" {
			t.Error("Alert ID should be generated")
		}

		if alert.Status != "active" {
			t.Errorf("Expected alert status 'active', got '%s'", alert.Status)
		}

		if alert.TriggeredAt.IsZero() {
			t.Error("Alert triggered at timestamp should be set")
		}

		// Retrieve alert
		retrievedAlert, err := alertService.GetAlert(ctx, alert.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve alert: %v", err)
		}

		// Verify alert data
		if retrievedAlert.BusinessID != businessID {
			t.Errorf("Expected business ID %s, got %s", businessID, retrievedAlert.BusinessID)
		}

		if retrievedAlert.FrameworkID != frameworkID {
			t.Errorf("Expected framework ID %s, got %s", frameworkID, retrievedAlert.FrameworkID)
		}

		if retrievedAlert.AlertType != "compliance_change" {
			t.Errorf("Expected alert type 'compliance_change', got '%s'", retrievedAlert.AlertType)
		}

		if retrievedAlert.Severity != "high" {
			t.Errorf("Expected severity 'high', got '%s'", retrievedAlert.Severity)
		}

		t.Logf("✅ Alert Creation and Retrieval: Alert %s created and retrieved successfully", alert.ID)
	})

	t.Run("Alert Status Updates", func(t *testing.T) {
		// Create a test alert
		alert := &compliance.ComplianceAlert{
			BusinessID:  businessID + "-status",
			FrameworkID: frameworkID,
			AlertType:   "deadline",
			Severity:    "medium",
			Title:       "Test Status Alert",
			Description: "Test alert for status updates",
			TriggeredBy: "test",
		}

		err := alertService.CreateAlert(ctx, alert)
		if err != nil {
			t.Fatalf("Failed to create alert: %v", err)
		}

		// Test status updates
		statusUpdates := []struct {
			status    string
			updatedBy string
		}{
			{"acknowledged", "user1"},
			{"resolved", "user2"},
		}

		for _, update := range statusUpdates {
			err := alertService.UpdateAlertStatus(ctx, alert.ID, update.status, update.updatedBy)
			if err != nil {
				t.Fatalf("Failed to update alert status to %s: %v", update.status, err)
			}

			// Retrieve and verify status
			updatedAlert, err := alertService.GetAlert(ctx, alert.ID)
			if err != nil {
				t.Fatalf("Failed to retrieve updated alert: %v", err)
			}

			if updatedAlert.Status != update.status {
				t.Errorf("Expected status %s, got %s", update.status, updatedAlert.Status)
			}

			// Verify timestamps
			switch update.status {
			case "acknowledged":
				if updatedAlert.AcknowledgedBy != update.updatedBy {
					t.Errorf("Expected acknowledged by %s, got %s", update.updatedBy, updatedAlert.AcknowledgedBy)
				}
				if updatedAlert.AcknowledgedAt == nil {
					t.Error("Acknowledged at timestamp should be set")
				}
			case "resolved":
				if updatedAlert.ResolvedBy != update.updatedBy {
					t.Errorf("Expected resolved by %s, got %s", update.updatedBy, updatedAlert.ResolvedBy)
				}
				if updatedAlert.ResolvedAt == nil {
					t.Error("Resolved at timestamp should be set")
				}
			}

			t.Logf("✅ Alert Status Update: Status changed to %s by %s", update.status, update.updatedBy)
		}
	})

	t.Run("Alert Filtering and Listing", func(t *testing.T) {
		// Create multiple alerts with different properties
		alerts := []*compliance.ComplianceAlert{
			{
				BusinessID:  businessID + "-filter1",
				FrameworkID: frameworkID,
				AlertType:   "compliance_change",
				Severity:    "high",
				Title:       "High Severity Alert",
				TriggeredBy: "system",
			},
			{
				BusinessID:  businessID + "-filter2",
				FrameworkID: frameworkID,
				AlertType:   "deadline",
				Severity:    "medium",
				Title:       "Medium Severity Alert",
				TriggeredBy: "user",
			},
			{
				BusinessID:  businessID + "-filter3",
				FrameworkID: "GDPR",
				AlertType:   "risk_threshold",
				Severity:    "critical",
				Title:       "Critical Severity Alert",
				TriggeredBy: "system",
			},
		}

		// Create all alerts
		for _, alert := range alerts {
			err := alertService.CreateAlert(ctx, alert)
			if err != nil {
				t.Fatalf("Failed to create alert: %v", err)
			}
		}

		// Test filtering by severity
		query := &compliance.AlertQuery{
			Severity: "high",
		}
		highSeverityAlerts, err := alertService.ListAlerts(ctx, query)
		if err != nil {
			t.Fatalf("Failed to list high severity alerts: %v", err)
		}

		if len(highSeverityAlerts) < 1 {
			t.Errorf("Expected at least 1 high severity alert, got %d", len(highSeverityAlerts))
		}

		// Test filtering by alert type
		query = &compliance.AlertQuery{
			AlertType: "deadline",
		}
		deadlineAlerts, err := alertService.ListAlerts(ctx, query)
		if err != nil {
			t.Fatalf("Failed to list deadline alerts: %v", err)
		}

		if len(deadlineAlerts) < 1 {
			t.Errorf("Expected at least 1 deadline alert, got %d", len(deadlineAlerts))
		}

		// Test filtering by framework
		query = &compliance.AlertQuery{
			FrameworkID: "GDPR",
		}
		gdprAlerts, err := alertService.ListAlerts(ctx, query)
		if err != nil {
			t.Fatalf("Failed to list GDPR alerts: %v", err)
		}

		if len(gdprAlerts) < 1 {
			t.Errorf("Expected at least 1 GDPR alert, got %d", len(gdprAlerts))
		}

		// Test filtering by triggered by
		query = &compliance.AlertQuery{
			TriggeredBy: "system",
		}
		systemAlerts, err := alertService.ListAlerts(ctx, query)
		if err != nil {
			t.Fatalf("Failed to list system alerts: %v", err)
		}

		if len(systemAlerts) < 2 {
			t.Errorf("Expected at least 2 system alerts, got %d", len(systemAlerts))
		}

		t.Logf("✅ Alert Filtering: High=%d, Deadline=%d, GDPR=%d, System=%d",
			len(highSeverityAlerts), len(deadlineAlerts), len(gdprAlerts), len(systemAlerts))
	})
}

// TestAlertRuleSystem tests the alert rule system
func TestAlertRuleSystem(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)
	alertService := compliance.NewComplianceAlertService(logger, frameworkService, trackingService)

	ctx := context.Background()
	businessID := "rule-test-business"
	frameworkID := "SOC2"

	t.Run("Alert Rule Creation", func(t *testing.T) {
		// Create a test alert rule
		rule := &compliance.AlertRule{
			Name:        "Test Compliance Rule",
			Description: "Test rule for compliance score monitoring",
			AlertType:   "compliance_change",
			Severity:    "medium",
			Conditions: []compliance.AlertCondition{
				{
					Field:     "compliance_score",
					Operator:  "lt",
					Threshold: 0.5,
				},
			},
			Actions: []compliance.AlertAction{
				{
					Type:    "email",
					Config:  map[string]interface{}{"template": "low_compliance"},
					Enabled: true,
				},
			},
			Enabled:   true,
			CreatedBy: "test-user",
		}

		// Create rule
		err := alertService.CreateAlertRule(ctx, rule)
		if err != nil {
			t.Fatalf("Failed to create alert rule: %v", err)
		}

		// Verify rule was created
		if rule.ID == "" {
			t.Error("Rule ID should be generated")
		}

		if rule.CreatedAt.IsZero() {
			t.Error("Rule created at timestamp should be set")
		}

		if rule.UpdatedAt.IsZero() {
			t.Error("Rule updated at timestamp should be set")
		}

		t.Logf("✅ Alert Rule Creation: Rule %s created successfully", rule.ID)
	})

	t.Run("Alert Rule Evaluation", func(t *testing.T) {
		// Setup tracking with low compliance score to trigger rule
		tracking, err := trackingService.GetComplianceTracking(ctx, businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get tracking: %v", err)
		}

		// Set low compliance score
		tracking.OverallProgress = 0.3 // Below 0.5 threshold
		tracking.ComplianceLevel = "non_compliant"
		tracking.RiskLevel = "high"

		err = trackingService.UpdateComplianceTracking(ctx, tracking)
		if err != nil {
			t.Fatalf("Failed to update tracking: %v", err)
		}

		// Evaluate alert rules
		err = alertService.EvaluateAlertRules(ctx, businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to evaluate alert rules: %v", err)
		}

		// Check if alerts were created
		query := &compliance.AlertQuery{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
		}
		alerts, err := alertService.ListAlerts(ctx, query)
		if err != nil {
			t.Fatalf("Failed to list alerts: %v", err)
		}

		// Should have at least one alert from rule evaluation
		if len(alerts) == 0 {
			t.Error("Expected at least one alert from rule evaluation")
		}

		// Verify alert properties
		for _, alert := range alerts {
			if alert.TriggeredBy != "rule" {
				t.Errorf("Expected alert triggered by 'rule', got '%s'", alert.TriggeredBy)
			}

			if alert.BusinessID != businessID {
				t.Errorf("Expected business ID %s, got %s", businessID, alert.BusinessID)
			}

			if alert.FrameworkID != frameworkID {
				t.Errorf("Expected framework ID %s, got %s", frameworkID, alert.FrameworkID)
			}
		}

		t.Logf("✅ Alert Rule Evaluation: %d alerts created from rule evaluation", len(alerts))
	})

	t.Run("Default Alert Rules", func(t *testing.T) {
		// Test that default alert rules are loaded
		// This tests the loadDefaultAlertRules function

		// Create tracking with different scenarios to trigger default rules
		testCases := []struct {
			name            string
			complianceScore float64
			riskLevel       string
			expectedAlerts  int
		}{
			{
				name:            "Low compliance score",
				complianceScore: 0.3,
				riskLevel:       "high",
				expectedAlerts:  1, // Should trigger low compliance score rule
			},
			{
				name:            "Critical risk level",
				complianceScore: 0.1,
				riskLevel:       "critical",
				expectedAlerts:  2, // Should trigger both low compliance and critical risk rules
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup tracking
				tracking, err := trackingService.GetComplianceTracking(ctx, businessID+"-"+tc.name, frameworkID)
				if err != nil {
					t.Fatalf("Failed to get tracking: %v", err)
				}

				tracking.OverallProgress = tc.complianceScore
				tracking.RiskLevel = tc.riskLevel

				err = trackingService.UpdateComplianceTracking(ctx, tracking)
				if err != nil {
					t.Fatalf("Failed to update tracking: %v", err)
				}

				// Evaluate rules
				err = alertService.EvaluateAlertRules(ctx, businessID+"-"+tc.name, frameworkID)
				if err != nil {
					t.Fatalf("Failed to evaluate alert rules: %v", err)
				}

				// Check alerts
				query := &compliance.AlertQuery{
					BusinessID:  businessID + "-" + tc.name,
					FrameworkID: frameworkID,
				}
				alerts, err := alertService.ListAlerts(ctx, query)
				if err != nil {
					t.Fatalf("Failed to list alerts: %v", err)
				}

				if len(alerts) < tc.expectedAlerts {
					t.Errorf("Expected at least %d alerts, got %d", tc.expectedAlerts, len(alerts))
				}

				t.Logf("✅ %s: %d alerts created", tc.name, len(alerts))
			})
		}
	})
}

// TestNotificationSystem tests the notification system
func TestNotificationSystem(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)
	alertService := compliance.NewComplianceAlertService(logger, frameworkService, trackingService)

	ctx := context.Background()
	businessID := "notification-test-business"
	frameworkID := "SOC2"

	t.Run("Notification Creation and Retrieval", func(t *testing.T) {
		// Create an alert first
		alert := &compliance.ComplianceAlert{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			AlertType:   "compliance_change",
			Severity:    "high",
			Title:       "Test Notification Alert",
			Description: "Test alert for notification system",
			TriggeredBy: "test",
		}

		err := alertService.CreateAlert(ctx, alert)
		if err != nil {
			t.Fatalf("Failed to create alert: %v", err)
		}

		// Check if notifications were created (alerts should trigger notifications)
		query := &compliance.NotificationQuery{
			AlertID: alert.ID,
		}
		notifications, err := alertService.GetNotifications(ctx, query)
		if err != nil {
			t.Fatalf("Failed to get notifications: %v", err)
		}

		// Should have at least one notification
		if len(notifications) == 0 {
			t.Error("Expected at least one notification for the alert")
		}

		// Verify notification properties
		for _, notification := range notifications {
			if notification.AlertID != alert.ID {
				t.Errorf("Expected alert ID %s, got %s", alert.ID, notification.AlertID)
			}

			if notification.Status != "sent" {
				t.Errorf("Expected notification status 'sent', got '%s'", notification.Status)
			}

			if notification.SentAt == nil {
				t.Error("Notification sent at timestamp should be set")
			}

			if notification.CreatedAt.IsZero() {
				t.Error("Notification created at timestamp should be set")
			}
		}

		t.Logf("✅ Notification Creation: %d notifications created for alert %s", len(notifications), alert.ID)
	})

	t.Run("Notification Filtering", func(t *testing.T) {
		// Create multiple alerts to generate notifications
		alerts := []*compliance.ComplianceAlert{
			{
				BusinessID:  businessID + "-notif1",
				FrameworkID: frameworkID,
				AlertType:   "compliance_change",
				Severity:    "high",
				Title:       "High Severity Alert",
				TriggeredBy: "test",
			},
			{
				BusinessID:  businessID + "-notif2",
				FrameworkID: frameworkID,
				AlertType:   "deadline",
				Severity:    "medium",
				Title:       "Medium Severity Alert",
				TriggeredBy: "test",
			},
		}

		// Create alerts
		for _, alert := range alerts {
			err := alertService.CreateAlert(ctx, alert)
			if err != nil {
				t.Fatalf("Failed to create alert: %v", err)
			}
		}

		// Test filtering by notification type
		query := &compliance.NotificationQuery{
			Type: "email",
		}
		emailNotifications, err := alertService.GetNotifications(ctx, query)
		if err != nil {
			t.Fatalf("Failed to get email notifications: %v", err)
		}

		// Should have email notifications
		if len(emailNotifications) == 0 {
			t.Error("Expected email notifications")
		}

		// Verify all notifications are email type
		for _, notification := range emailNotifications {
			if notification.Type != "email" {
				t.Errorf("Expected notification type 'email', got '%s'", notification.Type)
			}
		}

		// Test filtering by status
		query = &compliance.NotificationQuery{
			Status: "sent",
		}
		sentNotifications, err := alertService.GetNotifications(ctx, query)
		if err != nil {
			t.Fatalf("Failed to get sent notifications: %v", err)
		}

		// Should have sent notifications
		if len(sentNotifications) == 0 {
			t.Error("Expected sent notifications")
		}

		// Verify all notifications are sent
		for _, notification := range sentNotifications {
			if notification.Status != "sent" {
				t.Errorf("Expected notification status 'sent', got '%s'", notification.Status)
			}
		}

		t.Logf("✅ Notification Filtering: Email=%d, Sent=%d", len(emailNotifications), len(sentNotifications))
	})
}

// TestAlertSystemPerformance tests performance of the alert system
func TestAlertSystemPerformance(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)
	alertService := compliance.NewComplianceAlertService(logger, frameworkService, trackingService)

	ctx := context.Background()
	businessID := "perf-test-business"
	frameworkID := "SOC2"

	t.Run("Alert Creation Performance", func(t *testing.T) {
		// Test alert creation performance
		numAlerts := 100
		start := time.Now()

		for i := 0; i < numAlerts; i++ {
			alert := &compliance.ComplianceAlert{
				BusinessID:  businessID,
				FrameworkID: frameworkID,
				AlertType:   "compliance_change",
				Severity:    "medium",
				Title:       fmt.Sprintf("Performance Test Alert %d", i),
				TriggeredBy: "test",
			}

			err := alertService.CreateAlert(ctx, alert)
			if err != nil {
				t.Fatalf("Failed to create alert %d: %v", i, err)
			}
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(numAlerts)

		if avgDuration > 10*time.Millisecond {
			t.Errorf("Alert creation too slow: average %v per alert", avgDuration)
		}

		t.Logf("✅ Alert Creation Performance: %d alerts in %v (avg: %v per alert)",
			numAlerts, duration, avgDuration)
	})

	t.Run("Alert Listing Performance", func(t *testing.T) {
		// Test alert listing performance
		start := time.Now()

		query := &compliance.AlertQuery{
			BusinessID: businessID,
		}
		alerts, err := alertService.ListAlerts(ctx, query)
		if err != nil {
			t.Fatalf("Failed to list alerts: %v", err)
		}

		duration := time.Since(start)

		if duration > 100*time.Millisecond {
			t.Errorf("Alert listing too slow: %v", duration)
		}

		t.Logf("✅ Alert Listing Performance: %d alerts listed in %v", len(alerts), duration)
	})

	t.Run("Rule Evaluation Performance", func(t *testing.T) {
		// Test rule evaluation performance
		start := time.Now()

		err := alertService.EvaluateAlertRules(ctx, businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to evaluate alert rules: %v", err)
		}

		duration := time.Since(start)

		if duration > 200*time.Millisecond {
			t.Errorf("Rule evaluation too slow: %v", duration)
		}

		t.Logf("✅ Rule Evaluation Performance: Rules evaluated in %v", duration)
	})
}
