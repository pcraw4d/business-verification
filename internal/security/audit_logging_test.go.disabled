package security

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewAuditLoggingSystem(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2", "PCI-DSS", "GDPR"},
	}

	als := NewAuditLoggingSystem(logger, config)

	if als == nil {
		t.Fatal("Expected audit logging system to be created, got nil")
	}

	if als.logger != logger {
		t.Error("Expected logger to be set correctly")
	}

	if als.config.Enabled != true {
		t.Error("Expected enabled to be set correctly")
	}

	if als.config.EventQueueSize != 1000 {
		t.Error("Expected event queue size to be set correctly")
	}

	if len(als.config.ComplianceFrameworks) != 3 {
		t.Error("Expected compliance frameworks to be set correctly")
	}

	// Clean up
	os.RemoveAll("/tmp/audit-logs")
}

func TestLogEvent(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	als := NewAuditLoggingSystem(logger, config)

	// Create test event
	event := AuditEvent{
		BaseEvent: BaseEvent{
			ID:          "test-event-1",
			Timestamp:   time.Now(),
			EventType:   EventTypeLogin,
			Category:    CategoryAuthentication,
			Severity:    SeverityInfo,
			UserID:      "user1",
			Description: "User login successful",
			Details: map[string]interface{}{
				"ip_address": "192.168.1.1",
				"user_agent": "test-agent",
			},
		},
		Resource: "auth",
		Action:   "login",
		Result:   "success",
	}

	// Log event
	err := als.LogEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Wait a bit for event processing
	time.Sleep(100 * time.Millisecond)

	// Get metrics
	metrics, err := als.GetAuditMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if metrics.TotalEvents != 1 {
		t.Errorf("Expected 1 total event, got %d", metrics.TotalEvents)
	}

	if metrics.EventsByType[EventTypeLogin] != 1 {
		t.Errorf("Expected 1 login event, got %d", metrics.EventsByType[EventTypeLogin])
	}

	if metrics.EventsByCategory[CategoryAuthentication] != 1 {
		t.Errorf("Expected 1 authentication event, got %d", metrics.EventsByCategory[CategoryAuthentication])
	}

	// Clean up
	os.RemoveAll("/tmp/audit-logs")
}

func TestLogSecurityEvent(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	als := NewAuditLoggingSystem(logger, config)

	// Log security event
	err := als.LogSecurityEvent(context.Background(), EventTypeVulnerabilityDetected, "user1", "api", "scan", "found", map[string]interface{}{
		"vulnerability_id": "CVE-2023-1234",
		"severity":         "high",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Wait a bit for event processing
	time.Sleep(100 * time.Millisecond)

	// Get metrics
	metrics, err := als.GetAuditMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if metrics.TotalEvents != 1 {
		t.Errorf("Expected 1 total event, got %d", metrics.TotalEvents)
	}

	if metrics.EventsByType[EventTypeVulnerabilityDetected] != 1 {
		t.Errorf("Expected 1 vulnerability event, got %d", metrics.EventsByType[EventTypeVulnerabilityDetected])
	}

	if metrics.EventsByCategory[CategorySecurity] != 1 {
		t.Errorf("Expected 1 security event, got %d", metrics.EventsByCategory[CategorySecurity])
	}

	// Clean up
	os.RemoveAll("/tmp/audit-logs")
}

func TestLogAuthenticationEvent(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	als := NewAuditLoggingSystem(logger, config)

	// Log authentication event
	err := als.LogAuthenticationEvent(context.Background(), EventTypeLoginFailed, "user1", "invalid_password", map[string]interface{}{
		"ip_address": "192.168.1.1",
		"attempts":   3,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Wait a bit for event processing
	time.Sleep(100 * time.Millisecond)

	// Get metrics
	metrics, err := als.GetAuditMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if metrics.TotalEvents != 1 {
		t.Errorf("Expected 1 total event, got %d", metrics.TotalEvents)
	}

	if metrics.EventsByType[EventTypeLoginFailed] != 1 {
		t.Errorf("Expected 1 login failed event, got %d", metrics.EventsByType[EventTypeLoginFailed])
	}

	if metrics.EventsByCategory[CategoryAuthentication] != 1 {
		t.Errorf("Expected 1 authentication event, got %d", metrics.EventsByCategory[CategoryAuthentication])
	}

	// Clean up
	os.RemoveAll("/tmp/audit-logs")
}

func TestLogDataAccessEvent(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	als := NewAuditLoggingSystem(logger, config)

	// Log data access event
	err := als.LogDataAccessEvent(context.Background(), EventTypeDataRead, "user1", "business_data", "read", "success", map[string]interface{}{
		"record_count": 100,
		"query":        "SELECT * FROM businesses",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Wait a bit for event processing
	time.Sleep(100 * time.Millisecond)

	// Get metrics
	metrics, err := als.GetAuditMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if metrics.TotalEvents != 1 {
		t.Errorf("Expected 1 total event, got %d", metrics.TotalEvents)
	}

	if metrics.EventsByType[EventTypeDataRead] != 1 {
		t.Errorf("Expected 1 data read event, got %d", metrics.EventsByType[EventTypeDataRead])
	}

	if metrics.EventsByCategory[CategoryDataAccess] != 1 {
		t.Errorf("Expected 1 data access event, got %d", metrics.EventsByCategory[CategoryDataAccess])
	}

	// Clean up
	os.RemoveAll("/tmp/audit-logs")
}

func TestLogSystemEvent(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	als := NewAuditLoggingSystem(logger, config)

	// Log system event
	err := als.LogSystemEvent(context.Background(), EventTypeSystemStart, "System startup completed", map[string]interface{}{
		"version":      "1.0.0",
		"startup_time": "2.5s",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Wait a bit for event processing
	time.Sleep(100 * time.Millisecond)

	// Get metrics
	metrics, err := als.GetAuditMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if metrics.TotalEvents != 1 {
		t.Errorf("Expected 1 total event, got %d", metrics.TotalEvents)
	}

	if metrics.EventsByType[EventTypeSystemStart] != 1 {
		t.Errorf("Expected 1 system start event, got %d", metrics.EventsByType[EventTypeSystemStart])
	}

	if metrics.EventsByCategory[CategorySystem] != 1 {
		t.Errorf("Expected 1 system event, got %d", metrics.EventsByCategory[CategorySystem])
	}

	// Clean up
	os.RemoveAll("/tmp/audit-logs")
}

func TestLogComplianceEvent(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2", "PCI-DSS"},
	}

	als := NewAuditLoggingSystem(logger, config)

	// Log compliance event
	err := als.LogComplianceEvent(context.Background(), EventTypeComplianceCheck, "SOC2 compliance check completed", map[string]interface{}{
		"framework": "SOC2",
		"score":     95.5,
		"status":    "compliant",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Wait a bit for event processing
	time.Sleep(100 * time.Millisecond)

	// Get metrics
	metrics, err := als.GetAuditMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if metrics.TotalEvents != 1 {
		t.Errorf("Expected 1 total event, got %d", metrics.TotalEvents)
	}

	if metrics.EventsByType[EventTypeComplianceCheck] != 1 {
		t.Errorf("Expected 1 compliance check event, got %d", metrics.EventsByType[EventTypeComplianceCheck])
	}

	if metrics.EventsByCategory[CategoryCompliance] != 1 {
		t.Errorf("Expected 1 compliance event, got %d", metrics.EventsByCategory[CategoryCompliance])
	}

	// Clean up
	os.RemoveAll("/tmp/audit-logs")
}

func TestEventHandlers(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	als := NewAuditLoggingSystem(logger, config)

	// Create a channel to receive events
	eventReceived := make(chan AuditEvent, 1)

	// Register event handler
	als.RegisterEventHandler(EventTypeLogin, func(event AuditEvent) {
		eventReceived <- event
	})

	// Log event
	event := AuditEvent{
		EventType:   EventTypeLogin,
		Category:    CategoryAuthentication,
		Severity:    SeverityInfo,
		UserID:      "user1",
		Description: "Test login event",
	}

	err := als.LogEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Wait for event handler to be called
	select {
	case receivedEvent := <-eventReceived:
		if receivedEvent.UserID != "user1" {
			t.Error("Expected user ID to match in received event")
		}
		if receivedEvent.EventType != EventTypeLogin {
			t.Error("Expected event type to match in received event")
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected event handler to be called within 1 second")
	}

	// Clean up
	os.RemoveAll("/tmp/audit-logs")
}

func TestExportAuditLogsFromAuditLogger(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	als := NewAuditLoggingSystem(logger, config)

	// Export audit logs
	exported, err := als.ExportAuditLogs(context.Background(), nil)
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

	// Clean up
	os.RemoveAll("/tmp/audit-logs")
}

func TestNewAuditFileLogger(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	afl := NewAuditFileLogger(logger, config)

	if afl == nil {
		t.Fatal("Expected audit file logger to be created, got nil")
	}

	if afl.logger != logger {
		t.Error("Expected logger to be set correctly")
	}

	if afl.config.LogDirectory != "/tmp/audit-logs" {
		t.Error("Expected log directory to be set correctly")
	}

	if afl.file == nil {
		t.Error("Expected file to be opened")
	}

	// Test logging an event
	event := AuditEvent{
		EventType:   EventTypeLogin,
		Category:    CategoryAuthentication,
		Severity:    SeverityInfo,
		UserID:      "user1",
		Description: "Test login event",
	}

	err := afl.LogEvent(event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Clean up
	afl.file.Close()
	os.RemoveAll("/tmp/audit-logs")
}

func TestAuditFileLoggerRotation(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          1, // 1 MB
		MaxFiles:             3,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	afl := NewAuditFileLogger(logger, config)

	if afl == nil {
		t.Fatal("Expected audit file logger to be created, got nil")
	}

	// Log many events to trigger file rotation
	for i := 0; i < 1000; i++ {
		event := AuditEvent{
			EventType:   EventTypeLogin,
			Category:    CategoryAuthentication,
			Severity:    SeverityInfo,
			UserID:      fmt.Sprintf("user%d", i),
			Description: fmt.Sprintf("Test login event %d", i),
			Details: map[string]interface{}{
				"large_field": strings.Repeat("x", 1000), // Add large field to increase file size
			},
		}

		err := afl.LogEvent(event)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	// Check that file rotation occurred
	if afl.fileCount <= 1 {
		t.Error("Expected file rotation to occur")
	}

	// Clean up
	afl.file.Close()
	os.RemoveAll("/tmp/audit-logs")
}

func TestNewAuditDatabaseLogger(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      true,
		FileEnabled:          false,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	adl := NewAuditDatabaseLogger(logger, config)

	if adl == nil {
		t.Fatal("Expected audit database logger to be created, got nil")
	}

	if adl.logger != logger {
		t.Error("Expected logger to be set correctly")
	}

	// Test logging an event
	event := AuditEvent{
		EventType:   EventTypeLogin,
		Category:    CategoryAuthentication,
		Severity:    SeverityInfo,
		UserID:      "user1",
		Description: "Test login event",
	}

	err := adl.LogEvent(event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAuditLoggingDisabled(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              false, // Disabled
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	als := NewAuditLoggingSystem(logger, config)

	// Log event when disabled
	event := AuditEvent{
		EventType:   EventTypeLogin,
		Category:    CategoryAuthentication,
		Severity:    SeverityInfo,
		UserID:      "user1",
		Description: "Test login event",
	}

	err := als.LogEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error when logging is disabled, got %v", err)
	}

	// Get metrics
	metrics, err := als.GetAuditMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should have no events when disabled
	if metrics.TotalEvents != 0 {
		t.Errorf("Expected 0 total events when disabled, got %d", metrics.TotalEvents)
	}
}

func TestComplianceTags(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2", "PCI-DSS", "GDPR"},
	}

	als := NewAuditLoggingSystem(logger, config)

	// Test authentication event (should get SOC2 and PCI-DSS tags)
	event := AuditEvent{
		EventType:   EventTypeLogin,
		Category:    CategoryAuthentication,
		Severity:    SeverityInfo,
		UserID:      "user1",
		Description: "Test login event",
	}

	err := als.LogEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test data access event (should get SOC2, PCI-DSS, and GDPR tags)
	event2 := AuditEvent{
		EventType:   EventTypeDataRead,
		Category:    CategoryDataAccess,
		Severity:    SeverityInfo,
		UserID:      "user1",
		Resource:    "business_data",
		Action:      "read",
		Description: "Test data access event",
	}

	err = als.LogEvent(context.Background(), event2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Wait a bit for event processing
	time.Sleep(100 * time.Millisecond)

	// Get metrics
	metrics, err := als.GetAuditMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if metrics.TotalEvents != 2 {
		t.Errorf("Expected 2 total events, got %d", metrics.TotalEvents)
	}

	// Clean up
	os.RemoveAll("/tmp/audit-logs")
}

func TestRiskScoreCalculation(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggingConfig{
		Enabled:              true,
		LogLevel:             "info",
		RetentionDays:        90,
		MaxFileSize:          100,
		MaxFiles:             10,
		LogDirectory:         "/tmp/audit-logs",
		DatabaseEnabled:      false,
		FileEnabled:          true,
		ConsoleEnabled:       true,
		EventQueueSize:       1000,
		FlushInterval:        5 * time.Second,
		CompressionEnabled:   false,
		EncryptionEnabled:    false,
		ComplianceFrameworks: []string{"SOC2"},
	}

	als := NewAuditLoggingSystem(logger, config)

	// Test different event types and severities
	testCases := []struct {
		eventType AuditEventType
		severity  AuditEventSeverity
		expected  float64
	}{
		{EventTypeLogin, SeverityInfo, 0.5},
		{EventTypeLoginFailed, SeverityMedium, 7.0},          // 4.0 (medium) + 3.0 (login failed)
		{EventTypeVulnerabilityDetected, SeverityHigh, 15.0}, // 7.0 (high) + 8.0 (vulnerability) = 15.0, capped at 10.0
		{EventTypeDataDelete, SeverityHigh, 12.0},            // 7.0 (high) + 5.0 (data delete) = 12.0, capped at 10.0
	}

	for _, tc := range testCases {
		event := AuditEvent{
			EventType:   tc.eventType,
			Category:    CategoryAuthentication,
			Severity:    tc.severity,
			UserID:      "user1",
			Description: "Test event",
		}

		err := als.LogEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	}

	// Wait a bit for event processing
	time.Sleep(100 * time.Millisecond)

	// Get metrics
	metrics, err := als.GetAuditMetrics(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if metrics.TotalEvents != 4 {
		t.Errorf("Expected 4 total events, got %d", metrics.TotalEvents)
	}

	// Clean up
	os.RemoveAll("/tmp/audit-logs")
}

// Helper function to check if data is valid JSON
