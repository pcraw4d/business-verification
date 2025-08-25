package monitoring

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewSecurityMonitor(t *testing.T) {
	tests := []struct {
		name     string
		config   *SecurityMonitorConfig
		expected *SecurityMonitorConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			expected: &SecurityMonitorConfig{
				MaxEvents:      10000,
				EventRetention: 30 * 24 * time.Hour,
				AlertThresholds: map[SecurityEventSeverity]int{
					SeverityCritical: 1,
					SeverityHigh:     5,
					SeverityMedium:   10,
					SeverityLow:      50,
				},
				AlertCooldown:   5 * time.Minute,
				MetricsInterval: 1 * time.Minute,
				WebhookTimeout:  10 * time.Second,
			},
		},
		{
			name: "custom config",
			config: &SecurityMonitorConfig{
				MaxEvents:      5000,
				EventRetention: 7 * 24 * time.Hour,
				AlertCooldown:  1 * time.Minute,
				WebhookURL:     "https://example.com/webhook",
			},
			expected: &SecurityMonitorConfig{
				MaxEvents:      5000,
				EventRetention: 7 * 24 * time.Hour,
				AlertCooldown:  1 * time.Minute,
				WebhookURL:     "https://example.com/webhook",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			monitor := NewSecurityMonitor(tt.config, logger)

			assert.NotNil(t, monitor)
			assert.Equal(t, logger, monitor.logger)
			assert.NotNil(t, monitor.events)
			assert.NotNil(t, monitor.alerts)
			assert.NotNil(t, monitor.metrics)
			assert.NotNil(t, monitor.eventChan)
			assert.NotNil(t, monitor.alertChan)
			assert.NotNil(t, monitor.stopChan)

			// Clean up
			monitor.Stop()
		})
	}
}

func TestSecurityMonitor_RecordEvent(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSecurityMonitor(nil, logger)
	defer monitor.Stop()

	tests := []struct {
		name        string
		event       *SecurityEvent
		expectError bool
	}{
		{
			name:        "nil event",
			event:       nil,
			expectError: true,
		},
		{
			name: "valid event",
			event: &SecurityEvent{
				Type:      EventTypeLoginAttempt,
				Severity:  SeverityInfo,
				Source:    "auth",
				UserID:    "user123",
				IPAddress: "192.168.1.1",
			},
			expectError: false,
		},
		{
			name: "excluded source",
			event: &SecurityEvent{
				Type:     EventTypeLoginAttempt,
				Severity: SeverityInfo,
				Source:   "excluded_source",
			},
			expectError: false, // Not an error, just excluded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := monitor.RecordEvent(tt.event)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityMonitor_GetEvents(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSecurityMonitor(nil, logger)
	defer monitor.Stop()

	// Record some test events
	events := []*SecurityEvent{
		{
			Type:      EventTypeLoginAttempt,
			Severity:  SeverityInfo,
			Source:    "auth",
			UserID:    "user1",
			IPAddress: "192.168.1.1",
		},
		{
			Type:      EventTypeLoginFailure,
			Severity:  SeverityMedium,
			Source:    "auth",
			UserID:    "user2",
			IPAddress: "192.168.1.2",
		},
		{
			Type:      EventTypeRateLimitExceeded,
			Severity:  SeverityHigh,
			Source:    "api",
			IPAddress: "192.168.1.3",
		},
	}

	for _, event := range events {
		err := monitor.RecordEvent(event)
		assert.NoError(t, err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name     string
		filters  EventFilters
		expected int
	}{
		{
			name:     "no filters",
			filters:  EventFilters{},
			expected: 3,
		},
		{
			name: "filter by type",
			filters: EventFilters{
				Types: []SecurityEventType{EventTypeLoginAttempt},
			},
			expected: 1,
		},
		{
			name: "filter by severity",
			filters: EventFilters{
				Severities: []SecurityEventSeverity{SeverityHigh},
			},
			expected: 1,
		},
		{
			name: "filter by source",
			filters: EventFilters{
				Sources: []string{"auth"},
			},
			expected: 2,
		},
		{
			name: "filter by user ID",
			filters: EventFilters{
				UserIDs: []string{"user1"},
			},
			expected: 1,
		},
		{
			name: "filter by IP address",
			filters: EventFilters{
				IPAddresses: []string{"192.168.1.1"},
			},
			expected: 1,
		},
		{
			name: "multiple filters",
			filters: EventFilters{
				Types:      []SecurityEventType{EventTypeLoginAttempt, EventTypeLoginFailure},
				Severities: []SecurityEventSeverity{SeverityInfo},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := monitor.GetEvents(tt.filters)
			assert.NoError(t, err)
			assert.Len(t, events, tt.expected)
		})
	}
}

func TestSecurityMonitor_GetAlerts(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMonitorConfig{
		AlertThresholds: map[SecurityEventSeverity]int{
			SeverityCritical: 1,
			SeverityHigh:     1,
			SeverityMedium:   1,
		},
		AlertCooldown: 1 * time.Second,
	}
	monitor := NewSecurityMonitor(config, logger)
	defer monitor.Stop()

	// Record events that should generate alerts
	events := []*SecurityEvent{
		{
			Type:      EventTypeSQLInjectionAttempt,
			Severity:  SeverityCritical,
			Source:    "api",
			IPAddress: "192.168.1.1",
		},
		{
			Type:      EventTypeRateLimitExceeded,
			Severity:  SeverityHigh,
			Source:    "api",
			IPAddress: "192.168.1.2",
		},
		{
			Type:      EventTypeLoginFailure,
			Severity:  SeverityMedium,
			Source:    "auth",
			IPAddress: "192.168.1.3",
		},
	}

	for _, event := range events {
		err := monitor.RecordEvent(event)
		assert.NoError(t, err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name     string
		filters  AlertFilters
		expected int
	}{
		{
			name:     "no filters",
			filters:  AlertFilters{},
			expected: 3,
		},
		{
			name: "filter by severity",
			filters: AlertFilters{
				Severities: []SecurityEventSeverity{SeverityCritical},
			},
			expected: 1,
		},
		{
			name: "filter by source",
			filters: AlertFilters{
				Sources: []string{"api"},
			},
			expected: 2,
		},
		{
			name: "filter by acknowledged",
			filters: AlertFilters{
				Acknowledged: &[]bool{false}[0],
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts, err := monitor.GetAlerts(tt.filters)
			assert.NoError(t, err)
			assert.Len(t, alerts, tt.expected)
		})
	}
}

func TestSecurityMonitor_GetMetrics(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSecurityMonitor(nil, logger)
	defer monitor.Stop()

	// Record some test events
	events := []*SecurityEvent{
		{
			Type:      EventTypeLoginAttempt,
			Severity:  SeverityInfo,
			Source:    "auth",
			UserID:    "user1",
			IPAddress: "192.168.1.1",
		},
		{
			Type:      EventTypeLoginFailure,
			Severity:  SeverityMedium,
			Source:    "auth",
			UserID:    "user2",
			IPAddress: "192.168.1.2",
		},
		{
			Type:      EventTypeRateLimitExceeded,
			Severity:  SeverityHigh,
			Source:    "api",
			IPAddress: "192.168.1.3",
		},
	}

	for _, event := range events {
		err := monitor.RecordEvent(event)
		assert.NoError(t, err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	metrics := monitor.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(3), metrics.TotalEvents)
	assert.Equal(t, int64(1), metrics.EventsByType["login_attempt"])
	assert.Equal(t, int64(1), metrics.EventsByType["login_failure"])
	assert.Equal(t, int64(1), metrics.EventsByType["rate_limit_exceeded"])
	assert.Equal(t, int64(1), metrics.EventsBySeverity["info"])
	assert.Equal(t, int64(1), metrics.EventsBySeverity["medium"])
	assert.Equal(t, int64(1), metrics.EventsBySeverity["high"])
	assert.Equal(t, int64(2), metrics.EventsBySource["auth"])
	assert.Equal(t, int64(1), metrics.EventsBySource["api"])
}

func TestSecurityMonitor_ResolveEvent(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSecurityMonitor(nil, logger)
	defer monitor.Stop()

	// Record a test event
	event := &SecurityEvent{
		Type:      EventTypeLoginFailure,
		Severity:  SeverityMedium,
		Source:    "auth",
		UserID:    "user1",
		IPAddress: "192.168.1.1",
	}

	err := monitor.RecordEvent(event)
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Get the event to get its ID
	events, err := monitor.GetEvents(EventFilters{})
	assert.NoError(t, err)
	assert.Len(t, events, 1)

	eventID := events[0].ID

	// Resolve the event
	err = monitor.ResolveEvent(eventID, "admin", "Investigated and resolved")
	assert.NoError(t, err)

	// Verify the event is resolved
	events, err = monitor.GetEvents(EventFilters{})
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.True(t, events[0].Resolved)
	assert.Equal(t, "admin", events[0].ResolvedBy)
	assert.Equal(t, "Investigated and resolved", events[0].Notes)
	assert.NotNil(t, events[0].ResolvedAt)
}

func TestSecurityMonitor_AcknowledgeAlert(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMonitorConfig{
		AlertThresholds: map[SecurityEventSeverity]int{
			SeverityCritical: 1,
		},
		AlertCooldown: 1 * time.Second,
	}
	monitor := NewSecurityMonitor(config, logger)
	defer monitor.Stop()

	// Record an event that should generate an alert
	event := &SecurityEvent{
		Type:      EventTypeSQLInjectionAttempt,
		Severity:  SeverityCritical,
		Source:    "api",
		IPAddress: "192.168.1.1",
	}

	err := monitor.RecordEvent(event)
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Get the alert to get its ID
	alerts, err := monitor.GetAlerts(AlertFilters{})
	assert.NoError(t, err)
	assert.Len(t, alerts, 1)

	alertID := alerts[0].ID

	// Acknowledge the alert
	err = monitor.AcknowledgeAlert(alertID, "admin")
	assert.NoError(t, err)

	// Verify the alert is acknowledged
	alerts, err = monitor.GetAlerts(AlertFilters{})
	assert.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.True(t, alerts[0].Acknowledged)
	assert.Equal(t, "admin", alerts[0].AcknowledgedBy)
	assert.NotNil(t, alerts[0].AcknowledgedAt)
}

func TestSecurityMonitor_EventCallbacks(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSecurityMonitor(nil, logger)
	defer monitor.Stop()

	var eventCallbackCalled bool
	var alertCallbackCalled bool

	// Set callbacks
	monitor.SetEventCallback(func(event *SecurityEvent) {
		eventCallbackCalled = true
		assert.Equal(t, EventTypeLoginAttempt, event.Type)
	})

	monitor.SetAlertCallback(func(alert *SecurityAlert) {
		alertCallbackCalled = true
		assert.Equal(t, EventTypeSQLInjectionAttempt, SecurityEventType(alert.EventID))
	})

	// Record an event
	event := &SecurityEvent{
		Type:     EventTypeLoginAttempt,
		Severity: SeverityInfo,
		Source:   "auth",
	}

	err := monitor.RecordEvent(event)
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify event callback was called
	assert.True(t, eventCallbackCalled)

	// Record an event that should generate an alert
	config := &SecurityMonitorConfig{
		AlertThresholds: map[SecurityEventSeverity]int{
			SeverityCritical: 1,
		},
		AlertCooldown: 1 * time.Second,
	}
	alertMonitor := NewSecurityMonitor(config, logger)
	defer alertMonitor.Stop()

	alertMonitor.SetAlertCallback(func(alert *SecurityAlert) {
		alertCallbackCalled = true
	})

	alertEvent := &SecurityEvent{
		Type:     EventTypeSQLInjectionAttempt,
		Severity: SeverityCritical,
		Source:   "api",
	}

	err = alertMonitor.RecordEvent(alertEvent)
	assert.NoError(t, err)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify alert callback was called
	assert.True(t, alertCallbackCalled)
}

func TestSecurityMonitor_ExcludedEvents(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMonitorConfig{
		ExcludeSources:    []string{"excluded_source"},
		ExcludeEventTypes: []SecurityEventType{EventTypeSystemStartup},
	}
	monitor := NewSecurityMonitor(config, logger)
	defer monitor.Stop()

	tests := []struct {
		name          string
		event         *SecurityEvent
		shouldExclude bool
	}{
		{
			name: "excluded source",
			event: &SecurityEvent{
				Type:     EventTypeLoginAttempt,
				Severity: SeverityInfo,
				Source:   "excluded_source",
			},
			shouldExclude: true,
		},
		{
			name: "excluded event type",
			event: &SecurityEvent{
				Type:     EventTypeSystemStartup,
				Severity: SeverityInfo,
				Source:   "system",
			},
			shouldExclude: true,
		},
		{
			name: "normal event",
			event: &SecurityEvent{
				Type:     EventTypeLoginAttempt,
				Severity: SeverityInfo,
				Source:   "auth",
			},
			shouldExclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := monitor.RecordEvent(tt.event)
			assert.NoError(t, err)

			// Wait for processing
			time.Sleep(100 * time.Millisecond)

			events, err := monitor.GetEvents(EventFilters{})
			assert.NoError(t, err)

			if tt.shouldExclude {
				assert.Len(t, events, 0)
			} else {
				assert.Len(t, events, 1)
			}
		})
	}
}

func TestSecurityMonitor_AlertGeneration(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMonitorConfig{
		AlertThresholds: map[SecurityEventSeverity]int{
			SeverityCritical: 2,
			SeverityHigh:     3,
			SeverityMedium:   5,
		},
		AlertCooldown: 1 * time.Second,
	}
	monitor := NewSecurityMonitor(config, logger)
	defer monitor.Stop()

	tests := []struct {
		name           string
		events         []*SecurityEvent
		expectedAlerts int
	}{
		{
			name: "critical threshold met",
			events: []*SecurityEvent{
				{Type: EventTypeSQLInjectionAttempt, Severity: SeverityCritical, Source: "api", IPAddress: "192.168.1.1"},
				{Type: EventTypeSQLInjectionAttempt, Severity: SeverityCritical, Source: "api", IPAddress: "192.168.1.2"},
			},
			expectedAlerts: 1,
		},
		{
			name: "high threshold met",
			events: []*SecurityEvent{
				{Type: EventTypeRateLimitExceeded, Severity: SeverityHigh, Source: "api", IPAddress: "192.168.1.1"},
				{Type: EventTypeRateLimitExceeded, Severity: SeverityHigh, Source: "api", IPAddress: "192.168.1.2"},
				{Type: EventTypeRateLimitExceeded, Severity: SeverityHigh, Source: "api", IPAddress: "192.168.1.3"},
			},
			expectedAlerts: 1,
		},
		{
			name: "threshold not met",
			events: []*SecurityEvent{
				{Type: EventTypeLoginFailure, Severity: SeverityMedium, Source: "auth", IPAddress: "192.168.1.1"},
				{Type: EventTypeLoginFailure, Severity: SeverityMedium, Source: "auth", IPAddress: "192.168.1.2"},
			},
			expectedAlerts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Record events
			for _, event := range tt.events {
				err := monitor.RecordEvent(event)
				assert.NoError(t, err)
			}

			// Wait for processing
			time.Sleep(100 * time.Millisecond)

			// Check alerts
			alerts, err := monitor.GetAlerts(AlertFilters{})
			assert.NoError(t, err)
			assert.Len(t, alerts, tt.expectedAlerts)
		})
	}
}

func TestSecurityMonitor_EventRetention(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMonitorConfig{
		MaxEvents:      5,
		EventRetention: 100 * time.Millisecond,
	}
	monitor := NewSecurityMonitor(config, logger)
	defer monitor.Stop()

	// Record more events than max
	for i := 0; i < 10; i++ {
		event := &SecurityEvent{
			Type:     EventTypeLoginAttempt,
			Severity: SeverityInfo,
			Source:   "auth",
			UserID:   fmt.Sprintf("user%d", i),
		}
		err := monitor.RecordEvent(event)
		assert.NoError(t, err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Check that only max events are kept
	events, err := monitor.GetEvents(EventFilters{})
	assert.NoError(t, err)
	assert.Len(t, events, 5)

	// Wait for retention period
	time.Sleep(200 * time.Millisecond)

	// Check that events are cleaned up
	events, err = monitor.GetEvents(EventFilters{})
	assert.NoError(t, err)
	assert.Len(t, events, 0)
}

func TestSecurityMonitor_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSecurityMonitor(nil, logger)
	defer monitor.Stop()

	// Test concurrent event recording
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			event := &SecurityEvent{
				Type:     EventTypeLoginAttempt,
				Severity: SeverityInfo,
				Source:   "auth",
				UserID:   fmt.Sprintf("user%d", id),
			}
			err := monitor.RecordEvent(event)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify all events were recorded
	events, err := monitor.GetEvents(EventFilters{})
	assert.NoError(t, err)
	assert.Len(t, events, 10)
}

func TestSecurityMonitor_Stop(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSecurityMonitor(nil, logger)

	// Record an event
	event := &SecurityEvent{
		Type:     EventTypeLoginAttempt,
		Severity: SeverityInfo,
		Source:   "auth",
	}

	err := monitor.RecordEvent(event)
	assert.NoError(t, err)

	// Stop the monitor
	monitor.Stop()

	// Try to record another event (should fail due to closed channels)
	event2 := &SecurityEvent{
		Type:     EventTypeLoginFailure,
		Severity: SeverityMedium,
		Source:   "auth",
	}

	err = monitor.RecordEvent(event2)
	assert.Error(t, err)
}

func TestSecurityEvent_Validation(t *testing.T) {
	tests := []struct {
		name        string
		event       *SecurityEvent
		expectValid bool
	}{
		{
			name:        "nil event",
			event:       nil,
			expectValid: false,
		},
		{
			name: "valid event",
			event: &SecurityEvent{
				Type:     EventTypeLoginAttempt,
				Severity: SeverityInfo,
				Source:   "auth",
			},
			expectValid: true,
		},
		{
			name: "missing type",
			event: &SecurityEvent{
				Severity: SeverityInfo,
				Source:   "auth",
			},
			expectValid: true, // Type will be set to empty string
		},
		{
			name: "missing severity",
			event: &SecurityEvent{
				Type:   EventTypeLoginAttempt,
				Source: "auth",
			},
			expectValid: true, // Severity will be set to empty string
		},
		{
			name: "missing source",
			event: &SecurityEvent{
				Type:     EventTypeLoginAttempt,
				Severity: SeverityInfo,
			},
			expectValid: true, // Source will be set to empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			monitor := NewSecurityMonitor(nil, logger)
			defer monitor.Stop()

			err := monitor.RecordEvent(tt.event)

			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
