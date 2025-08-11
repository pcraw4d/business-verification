package security

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

func TestNewSecurityMonitor(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		AlertThresholds: map[SecuritySeverity]int{
			SeverityMedium: 5,
			SeverityLow:    10,
		},
		RetentionDays:  30,
		AutoResolution: false,
	}

	monitor := NewSecurityMonitor(logger, config)

	if monitor == nil {
		t.Fatal("Expected monitor to be created, got nil")
	}

	if monitor.logger != logger {
		t.Error("Expected logger to be set correctly")
	}

	if monitor.config.RetentionDays != 30 {
		t.Error("Expected retention days to be set correctly")
	}
}

func TestRecordEvent(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		AlertThresholds: map[SecuritySeverity]int{
			SeverityMedium: 5,
		},
		RetentionDays: 30,
	}

	monitor := NewSecurityMonitor(logger, config)

	event := SecurityEvent{
		EventType:   EventTypeAuthenticationFailure,
		Severity:    SeverityHigh,
		Source:      "test-source",
		Description: "Test authentication failure",
	}

	err := monitor.RecordEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	events, err := monitor.GetEvents(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	if events[0].EventType != EventTypeAuthenticationFailure {
		t.Error("Expected event type to match")
	}

	if events[0].Severity != SeverityHigh {
		t.Error("Expected severity to match")
	}
}

func TestRecordEventGeneratesAlert(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		AlertThresholds: map[SecuritySeverity]int{
			SeverityMedium: 5,
		},
		RetentionDays: 30,
	}

	monitor := NewSecurityMonitor(logger, config)

	// Record a critical event which should generate an alert
	event := SecurityEvent{
		EventType:   EventTypeAuthorizationFailure,
		Severity:    SeverityCritical,
		Source:      "test-source",
		Description: "Test authorization failure",
	}

	err := monitor.RecordEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	alerts, err := monitor.GetAlerts(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}

	if alerts[0].Severity != SeverityCritical {
		t.Error("Expected alert severity to match event severity")
	}

	if alerts[0].Status != AlertStatusOpen {
		t.Error("Expected alert status to be open")
	}
}

func TestThresholdBasedAlerting(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		AlertThresholds: map[SecuritySeverity]int{
			SeverityMedium: 3,
		},
		RetentionDays: 30,
	}

	monitor := NewSecurityMonitor(logger, config)

	// Record 2 medium events (below threshold)
	for i := 0; i < 2; i++ {
		event := SecurityEvent{
			EventType:   EventTypeAuthenticationFailure,
			Severity:    SeverityMedium,
			Source:      "test-source",
			Description: "Test authentication failure",
		}
		err := monitor.RecordEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	alerts, err := monitor.GetAlerts(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(alerts) != 0 {
		t.Fatalf("Expected 0 alerts (below threshold), got %d", len(alerts))
	}

	// Record 1 more event to reach threshold
	event := SecurityEvent{
		EventType:   EventTypeAuthenticationFailure,
		Severity:    SeverityMedium,
		Source:      "test-source",
		Description: "Test authentication failure",
	}
	err := monitor.RecordEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	alerts, err = monitor.GetAlerts(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert (at threshold), got %d", len(alerts))
	}
}

func TestGetEventsWithFilters(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		RetentionDays: 30,
	}

	monitor := NewSecurityMonitor(logger, config)

	// Record different types of events
	events := []SecurityEvent{
		{
			EventType:   EventTypeAuthenticationFailure,
			Severity:    SeverityHigh,
			Source:      "source1",
			Description: "Auth failure 1",
		},
		{
			EventType:   EventTypeAuthorizationFailure,
			Severity:    SeverityMedium,
			Source:      "source2",
			Description: "Auth failure 2",
		},
		{
			EventType:   EventTypeAuthenticationFailure,
			Severity:    SeverityLow,
			Source:      "source1",
			Description: "Auth failure 3",
		},
	}

	for _, event := range events {
		err := monitor.RecordEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	// Test filtering by event type
	filters := map[string]interface{}{
		"event_type": EventTypeAuthenticationFailure,
	}
	filteredEvents, err := monitor.GetEvents(context.Background(), filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(filteredEvents) != 2 {
		t.Fatalf("Expected 2 authentication failure events, got %d", len(filteredEvents))
	}

	// Test filtering by severity
	filters = map[string]interface{}{
		"severity": SeverityHigh,
	}
	filteredEvents, err = monitor.GetEvents(context.Background(), filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(filteredEvents) != 1 {
		t.Fatalf("Expected 1 high severity event, got %d", len(filteredEvents))
	}

	// Test filtering by source
	filters = map[string]interface{}{
		"source": "source1",
	}
	filteredEvents, err = monitor.GetEvents(context.Background(), filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(filteredEvents) != 2 {
		t.Fatalf("Expected 2 events from source1, got %d", len(filteredEvents))
	}
}

func TestGetAlertsWithFilters(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		RetentionDays: 30,
	}

	monitor := NewSecurityMonitor(logger, config)

	// Record events that will generate alerts
	events := []SecurityEvent{
		{
			EventType:   EventTypeAuthenticationFailure,
			Severity:    SeverityCritical,
			Source:      "source1",
			Description: "Critical auth failure",
		},
		{
			EventType:   EventTypeAuthorizationFailure,
			Severity:    SeverityHigh,
			Source:      "source2",
			Description: "High auth failure",
		},
	}

	for _, event := range events {
		err := monitor.RecordEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	// Test filtering by severity
	filters := map[string]interface{}{
		"severity": SeverityCritical,
	}
	filteredAlerts, err := monitor.GetAlerts(context.Background(), filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(filteredAlerts) != 1 {
		t.Fatalf("Expected 1 critical alert, got %d", len(filteredAlerts))
	}

	// Test filtering by status
	filters = map[string]interface{}{
		"status": AlertStatusOpen,
	}
	filteredAlerts, err = monitor.GetAlerts(context.Background(), filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(filteredAlerts) != 2 {
		t.Fatalf("Expected 2 open alerts, got %d", len(filteredAlerts))
	}
}

func TestUpdateAlertStatus(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		RetentionDays: 30,
	}

	monitor := NewSecurityMonitor(logger, config)

	// Record an event that will generate an alert
	event := SecurityEvent{
		EventType:   EventTypeAuthenticationFailure,
		Severity:    SeverityCritical,
		Source:      "test-source",
		Description: "Test authentication failure",
	}

	err := monitor.RecordEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	alerts, err := monitor.GetAlerts(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}

	alertID := alerts[0].ID

	// Update alert status
	err = monitor.UpdateAlertStatus(context.Background(), alertID, AlertStatusResolved, "Issue resolved")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify status was updated
	updatedAlerts, err := monitor.GetAlerts(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(updatedAlerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(updatedAlerts))
	}

	if updatedAlerts[0].Status != AlertStatusResolved {
		t.Error("Expected alert status to be resolved")
	}

	if len(updatedAlerts[0].Notes) != 1 {
		t.Error("Expected alert to have 1 note")
	}

	if updatedAlerts[0].Notes[0] != "Issue resolved" {
		t.Error("Expected note to match")
	}
}

func TestGetMetrics(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		RetentionDays: 30,
	}

	monitor := NewSecurityMonitor(logger, config)

	// Record various events
	events := []SecurityEvent{
		{
			EventType:   EventTypeAuthenticationFailure,
			Severity:    SeverityHigh,
			Source:      "source1",
			Description: "Auth failure 1",
		},
		{
			EventType:   EventTypeAuthorizationFailure,
			Severity:    SeverityMedium,
			Source:      "source2",
			Description: "Auth failure 2",
		},
		{
			EventType:   EventTypeAuthenticationFailure,
			Severity:    SeverityLow,
			Source:      "source1",
			Description: "Auth failure 3",
		},
	}

	for _, event := range events {
		err := monitor.RecordEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	metrics, err := monitor.GetMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if metrics.TotalEvents != 3 {
		t.Errorf("Expected 3 total events, got %d", metrics.TotalEvents)
	}

	if metrics.EventsByType[EventTypeAuthenticationFailure] != 2 {
		t.Errorf("Expected 2 authentication failure events, got %d", metrics.EventsByType[EventTypeAuthenticationFailure])
	}

	if metrics.EventsByType[EventTypeAuthorizationFailure] != 1 {
		t.Errorf("Expected 1 authorization failure event, got %d", metrics.EventsByType[EventTypeAuthorizationFailure])
	}

	if metrics.EventsBySeverity[SeverityHigh] != 1 {
		t.Errorf("Expected 1 high severity event, got %d", metrics.EventsBySeverity[SeverityHigh])
	}

	if metrics.EventsBySeverity[SeverityMedium] != 1 {
		t.Errorf("Expected 1 medium severity event, got %d", metrics.EventsBySeverity[SeverityMedium])
	}

	if metrics.EventsBySeverity[SeverityLow] != 1 {
		t.Errorf("Expected 1 low severity event, got %d", metrics.EventsBySeverity[SeverityLow])
	}

	// Should have alerts for high severity events
	if metrics.OpenAlerts < 1 {
		t.Error("Expected at least 1 open alert for high severity events")
	}
}

func TestEventHandlerRegistration(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		RetentionDays: 30,
	}

	monitor := NewSecurityMonitor(logger, config)

	// Track if handler was called
	handlerCalled := false
	var capturedEvent SecurityEvent

	handler := func(event SecurityEvent) {
		handlerCalled = true
		capturedEvent = event
	}

	// Register handler
	monitor.RegisterEventHandler(EventTypeAuthenticationFailure, handler)

	// Record an event
	event := SecurityEvent{
		EventType:   EventTypeAuthenticationFailure,
		Severity:    SeverityMedium,
		Source:      "test-source",
		Description: "Test event",
	}

	err := monitor.RecordEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Wait a bit for handler to be called
	time.Sleep(100 * time.Millisecond)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if capturedEvent.EventType != EventTypeAuthenticationFailure {
		t.Error("Expected captured event to match")
	}
}

func TestAlertHandlerRegistration(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		RetentionDays: 30,
	}

	monitor := NewSecurityMonitor(logger, config)

	// Track if handler was called
	handlerCalled := false
	var capturedAlert SecurityAlert

	handler := func(alert SecurityAlert) {
		handlerCalled = true
		capturedAlert = alert
	}

	// Register handler
	monitor.RegisterAlertHandler(SeverityCritical, handler)

	// Record an event that will generate a critical alert
	event := SecurityEvent{
		EventType:   EventTypeAuthenticationFailure,
		Severity:    SeverityCritical,
		Source:      "test-source",
		Description: "Test event",
	}

	err := monitor.RecordEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Wait a bit for handler to be called
	time.Sleep(100 * time.Millisecond)

	if !handlerCalled {
		t.Error("Expected alert handler to be called")
	}

	if capturedAlert.Severity != SeverityCritical {
		t.Error("Expected captured alert severity to match")
	}
}

func TestExportEvents(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		RetentionDays: 30,
	}

	monitor := NewSecurityMonitor(logger, config)

	// Record an event
	event := SecurityEvent{
		EventType:   EventTypeAuthenticationFailure,
		Severity:    SeverityMedium,
		Source:      "test-source",
		Description: "Test event",
	}

	err := monitor.RecordEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Export events
	exported, err := monitor.ExportEvents(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(exported) == 0 {
		t.Error("Expected exported data to not be empty")
	}

	// Verify it's valid JSON
	if !isValidJSON(exported) {
		t.Error("Expected exported data to be valid JSON")
	}
}

func TestExportAlerts(t *testing.T) {
	logger := observability.NewLogger("test", "debug")
	config := SecurityMonitorConfig{
		RetentionDays: 30,
	}

	monitor := NewSecurityMonitor(logger, config)

	// Record an event that will generate an alert
	event := SecurityEvent{
		EventType:   EventTypeAuthenticationFailure,
		Severity:    SeverityCritical,
		Source:      "test-source",
		Description: "Test event",
	}

	err := monitor.RecordEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Export alerts
	exported, err := monitor.ExportAlerts(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(exported) == 0 {
		t.Error("Expected exported data to not be empty")
	}

	// Verify it's valid JSON
	if !isValidJSON(exported) {
		t.Error("Expected exported data to be valid JSON")
	}
}

// Helper function to check if data is valid JSON
func isValidJSON(data []byte) bool {
	var v interface{}
	return json.Unmarshal(data, &v) == nil
}
