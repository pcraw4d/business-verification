package services

import (
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestAlertingService_CreateAlert(t *testing.T) {
	logger := zap.NewNop()
	service := NewAlertingService(logger)

	alert := &Alert{
		Title:       "Test Alert",
		Description: "This is a test alert",
		Severity:    AlertSeverityWarning,
		Source:      "test",
		Labels:      map[string]string{"test": "true"},
	}

	err := service.CreateAlert(alert)
	if err != nil {
		t.Fatalf("Failed to create alert: %v", err)
	}

	if alert.ID == "" {
		t.Error("Alert ID should be generated")
	}

	if alert.Timestamp.IsZero() {
		t.Error("Alert timestamp should be set")
	}

	if alert.Status != AlertStatusActive {
		t.Errorf("Expected status 'active', got '%s'", alert.Status)
	}
}

func TestAlertingService_ResolveAlert(t *testing.T) {
	logger := zap.NewNop()
	service := NewAlertingService(logger)

	alert := &Alert{
		Title:       "Test Alert",
		Description: "This is a test alert",
		Severity:    AlertSeverityWarning,
		Source:      "test",
	}

	err := service.CreateAlert(alert)
	if err != nil {
		t.Fatalf("Failed to create alert: %v", err)
	}

	err = service.ResolveAlert(alert.ID)
	if err != nil {
		t.Fatalf("Failed to resolve alert: %v", err)
	}

	activeAlerts := service.GetActiveAlerts()
	for _, activeAlert := range activeAlerts {
		if activeAlert.ID == alert.ID {
			t.Error("Alert should not be in active alerts after resolution")
		}
	}
}

func TestAlertingService_ResolveNonExistentAlert(t *testing.T) {
	logger := zap.NewNop()
	service := NewAlertingService(logger)

	err := service.ResolveAlert("non-existent-id")
	if err == nil {
		t.Error("Expected error when resolving non-existent alert")
	}
}

func TestAlertingService_GetActiveAlerts(t *testing.T) {
	logger := zap.NewNop()
	service := NewAlertingService(logger)

	// Create multiple alerts
	alerts := []*Alert{
		{
			Title:       "Alert 1",
			Description: "First alert",
			Severity:    AlertSeverityInfo,
			Source:      "test",
		},
		{
			Title:       "Alert 2",
			Description: "Second alert",
			Severity:    AlertSeverityWarning,
			Source:      "test",
		},
		{
			Title:       "Alert 3",
			Description: "Third alert",
			Severity:    AlertSeverityCritical,
			Source:      "test",
		},
	}

	for _, alert := range alerts {
		err := service.CreateAlert(alert)
		if err != nil {
			t.Fatalf("Failed to create alert: %v", err)
		}
	}

	activeAlerts := service.GetActiveAlerts()
	if len(activeAlerts) != len(alerts) {
		t.Errorf("Expected %d active alerts, got %d", len(alerts), len(activeAlerts))
	}

	// Resolve one alert
	err := service.ResolveAlert(alerts[0].ID)
	if err != nil {
		t.Fatalf("Failed to resolve alert: %v", err)
	}

	activeAlerts = service.GetActiveAlerts()
	if len(activeAlerts) != len(alerts)-1 {
		t.Errorf("Expected %d active alerts after resolution, got %d", len(alerts)-1, len(activeAlerts))
	}
}

func TestAlertingService_GetAlertsBySeverity(t *testing.T) {
	logger := zap.NewNop()
	service := NewAlertingService(logger)

	// Create alerts with different severities
	alerts := []*Alert{
		{
			Title:       "Info Alert",
			Description: "Info level alert",
			Severity:    AlertSeverityInfo,
			Source:      "test",
		},
		{
			Title:       "Warning Alert",
			Description: "Warning level alert",
			Severity:    AlertSeverityWarning,
			Source:      "test",
		},
		{
			Title:       "Critical Alert",
			Description: "Critical level alert",
			Severity:    AlertSeverityCritical,
			Source:      "test",
		},
	}

	for _, alert := range alerts {
		err := service.CreateAlert(alert)
		if err != nil {
			t.Fatalf("Failed to create alert: %v", err)
		}
	}

	// Test filtering by severity
	warningAlerts := service.GetAlertsBySeverity(AlertSeverityWarning)
	if len(warningAlerts) != 1 {
		t.Errorf("Expected 1 warning alert, got %d", len(warningAlerts))
	}

	if warningAlerts[0].Severity != AlertSeverityWarning {
		t.Errorf("Expected warning severity, got %s", warningAlerts[0].Severity)
	}

	criticalAlerts := service.GetAlertsBySeverity(AlertSeverityCritical)
	if len(criticalAlerts) != 1 {
		t.Errorf("Expected 1 critical alert, got %d", len(criticalAlerts))
	}

	if criticalAlerts[0].Severity != AlertSeverityCritical {
		t.Errorf("Expected critical severity, got %s", criticalAlerts[0].Severity)
	}
}

func TestAlertingService_GetAlertHistory(t *testing.T) {
	logger := zap.NewNop()
	service := NewAlertingService(logger)

	// Create multiple alerts
	for i := 0; i < 5; i++ {
		alert := &Alert{
			Title:       fmt.Sprintf("Alert %d", i),
			Description: fmt.Sprintf("Alert number %d", i),
			Severity:    AlertSeverityInfo,
			Source:      "test",
		}
		err := service.CreateAlert(alert)
		if err != nil {
			t.Fatalf("Failed to create alert: %v", err)
		}
	}

	history := service.GetAlertHistory(10)
	if len(history) != 5 {
		t.Errorf("Expected 5 alerts in history, got %d", len(history))
	}

	// Test limit
	limitedHistory := service.GetAlertHistory(3)
	if len(limitedHistory) != 3 {
		t.Errorf("Expected 3 alerts in limited history, got %d", len(limitedHistory))
	}
}

func TestAlertingService_AddNotifier(t *testing.T) {
	logger := zap.NewNop()
	service := NewAlertingService(logger)

	notifier := NewMockAlertNotifier(logger)
	service.AddNotifier(notifier)

	// Create an alert to test notification
	alert := &Alert{
		Title:       "Test Alert",
		Description: "This is a test alert",
		Severity:    AlertSeverityWarning,
		Source:      "test",
	}

	err := service.CreateAlert(alert)
	if err != nil {
		t.Fatalf("Failed to create alert: %v", err)
	}

	// Give some time for notification to be sent
	time.Sleep(100 * time.Millisecond)
}

func TestAlertingService_CheckAlertRules(t *testing.T) {
	logger := zap.NewNop()
	service := NewAlertingService(logger)

	// Test with metrics that should trigger alerts
	metrics := map[string]float64{
		"high_error_rate":    6.0,    // Should trigger (threshold: 5.0)
		"high_response_time": 2500.0, // Should trigger (threshold: 2000.0)
		"high_memory_usage":  85.0,   // Should trigger (threshold: 80.0)
		"high_cpu_usage":     90.0,   // Should trigger (threshold: 80.0)
	}

	err := service.CheckAlertRules(nil, metrics)
	if err != nil {
		t.Fatalf("Failed to check alert rules: %v", err)
	}

	// Check that alerts were created
	activeAlerts := service.GetActiveAlerts()
	if len(activeAlerts) == 0 {
		t.Error("Expected alerts to be created from rules")
	}
}

func TestAlertingService_CheckAlertRules_NoTriggers(t *testing.T) {
	logger := zap.NewNop()
	service := NewAlertingService(logger)

	// Test with metrics that should not trigger alerts
	metrics := map[string]float64{
		"high_error_rate":    2.0,    // Should not trigger (threshold: 5.0)
		"high_response_time": 1000.0, // Should not trigger (threshold: 2000.0)
		"high_memory_usage":  50.0,   // Should not trigger (threshold: 80.0)
		"high_cpu_usage":     60.0,   // Should not trigger (threshold: 80.0)
	}

	err := service.CheckAlertRules(nil, metrics)
	if err != nil {
		t.Fatalf("Failed to check alert rules: %v", err)
	}

	// Check that no alerts were created
	activeAlerts := service.GetActiveAlerts()
	if len(activeAlerts) != 0 {
		t.Errorf("Expected no alerts to be created, got %d", len(activeAlerts))
	}
}

func TestMockAlertNotifier(t *testing.T) {
	logger := zap.NewNop()
	notifier := NewMockAlertNotifier(logger)

	alert := &Alert{
		ID:          "test-alert",
		Title:       "Test Alert",
		Description: "This is a test alert",
		Severity:    AlertSeverityInfo,
		Source:      "test",
	}

	err := notifier.Notify(alert)
	if err != nil {
		t.Fatalf("Mock notifier should not return error: %v", err)
	}

	if notifier.Name() != "mock" {
		t.Errorf("Expected notifier name 'mock', got '%s'", notifier.Name())
	}
}

func TestAlert_Structure(t *testing.T) {
	alert := &Alert{
		ID:          "test-alert",
		Title:       "Test Alert",
		Description: "This is a test alert",
		Severity:    AlertSeverityWarning,
		Status:      AlertStatusActive,
		Source:      "test",
		Timestamp:   time.Now(),
		Labels:      map[string]string{"test": "true"},
		Metadata:    map[string]interface{}{"count": 1},
	}

	// Test that all fields are properly set
	if alert.ID == "" {
		t.Error("ID should not be empty")
	}
	if alert.Title == "" {
		t.Error("Title should not be empty")
	}
	if alert.Description == "" {
		t.Error("Description should not be empty")
	}
	if alert.Severity == "" {
		t.Error("Severity should not be empty")
	}
	if alert.Status == "" {
		t.Error("Status should not be empty")
	}
	if alert.Source == "" {
		t.Error("Source should not be empty")
	}
	if alert.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if alert.Labels == nil {
		t.Error("Labels should not be nil")
	}
	if alert.Metadata == nil {
		t.Error("Metadata should not be nil")
	}
}

func BenchmarkAlertingService_CreateAlert(b *testing.B) {
	logger := zap.NewNop()
	service := NewAlertingService(logger)

	alert := &Alert{
		Title:       "Benchmark Alert",
		Description: "This is a benchmark alert",
		Severity:    AlertSeverityInfo,
		Source:      "benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateAlert(alert)
	}
}

func BenchmarkAlertingService_GetActiveAlerts(b *testing.B) {
	logger := zap.NewNop()
	service := NewAlertingService(logger)

	// Create some alerts first
	for i := 0; i < 100; i++ {
		alert := &Alert{
			Title:       fmt.Sprintf("Alert %d", i),
			Description: fmt.Sprintf("Alert number %d", i),
			Severity:    AlertSeverityInfo,
			Source:      "benchmark",
		}
		service.CreateAlert(alert)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetActiveAlerts()
	}
}
