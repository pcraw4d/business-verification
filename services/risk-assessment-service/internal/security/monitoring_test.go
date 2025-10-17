package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecurityMonitor(t *testing.T) {
	tests := []struct {
		name   string
		config *SecurityMonitoringConfig
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "custom config",
			config: &SecurityMonitoringConfig{
				AlertThresholds: map[string]int{
					"failed_login": 3,
					"brute_force":  5,
				},
				IncidentResponseTime: 10 * time.Minute,
				AutoResponseEnabled:  false,
				NotificationChannels: []string{"email"},
				RetentionPeriod:      7 * 24 * time.Hour,
				RealTimeMonitoring:   false,
				AnomalyDetection:     false,
				ThreatIntelligence:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			sm := NewSecurityMonitor(tt.config, mockLogger)
			assert.NotNil(t, sm)
			assert.NotNil(t, sm.config)
			assert.NotNil(t, sm.alerts)
		})
	}
}

func TestSecurityMonitor_CreateAlert(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	tests := []struct {
		name        string
		alertType   string
		severity    string
		title       string
		description string
		source      string
		metadata    map[string]interface{}
		expectError bool
	}{
		{
			name:        "valid alert",
			alertType:   "failed_login",
			severity:    "MEDIUM",
			title:       "Multiple failed login attempts",
			description: "User has failed to login 5 times in 10 minutes",
			source:      "authentication_service",
			metadata: map[string]interface{}{
				"user_id": "user123",
				"ip":      "192.168.1.1",
				"count":   5,
			},
			expectError: false,
		},
		{
			name:        "critical alert",
			alertType:   "data_breach",
			severity:    "CRITICAL",
			title:       "Potential data breach detected",
			description: "Unauthorized access to sensitive data detected",
			source:      "security_scanner",
			metadata: map[string]interface{}{
				"affected_records": 1000,
				"data_type":        "personal_data",
			},
			expectError: false,
		},
		{
			name:        "alert with minimal data",
			alertType:   "suspicious_activity",
			severity:    "LOW",
			title:       "Suspicious activity detected",
			description: "Unusual pattern detected",
			source:      "monitoring_system",
			metadata:    nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			alert, err := sm.CreateAlert(ctx, tt.alertType, tt.severity, tt.title, tt.description, tt.source, tt.metadata)

			require.NoError(t, err)
			assert.NotNil(t, alert)
			assert.Equal(t, tt.alertType, alert.Type)
			assert.Equal(t, tt.severity, alert.Severity)
			assert.Equal(t, tt.title, alert.Title)
			assert.Equal(t, tt.description, alert.Description)
			assert.Equal(t, tt.source, alert.Source)
			assert.Equal(t, "OPEN", alert.Status)
			assert.False(t, alert.Timestamp.IsZero())
			assert.False(t, alert.CreatedAt.IsZero())
			assert.False(t, alert.UpdatedAt.IsZero())
		})
	}
}

func TestSecurityMonitor_CreateIncident(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	tests := []struct {
		name        string
		alertID     string
		title       string
		description string
		severity    string
		expectError bool
	}{
		{
			name:        "valid incident",
			alertID:     "alert_failed_login_1234567890",
			title:       "Multiple failed login attempts",
			description: "User has failed to login multiple times",
			severity:    "MEDIUM",
			expectError: false,
		},
		{
			name:        "critical incident",
			alertID:     "alert_data_breach_1234567890",
			title:       "Data breach incident",
			description: "Unauthorized access to sensitive data",
			severity:    "CRITICAL",
			expectError: false,
		},
		{
			name:        "high severity incident",
			alertID:     "alert_brute_force_1234567890",
			title:       "Brute force attack",
			description: "Multiple failed login attempts from same IP",
			severity:    "HIGH",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			incident, err := sm.CreateIncident(ctx, tt.alertID, tt.title, tt.description, tt.severity)

			require.NoError(t, err)
			assert.NotNil(t, incident)
			assert.Equal(t, tt.alertID, incident.AlertID)
			assert.Equal(t, tt.title, incident.Title)
			assert.Equal(t, tt.description, incident.Description)
			assert.Equal(t, tt.severity, incident.Severity)
			assert.Equal(t, "OPEN", incident.Status)
			assert.GreaterOrEqual(t, incident.EscalationLevel, 1)
			assert.False(t, incident.CreatedAt.IsZero())
			assert.False(t, incident.UpdatedAt.IsZero())
		})
	}
}

func TestSecurityMonitor_UpdateIncident(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	tests := []struct {
		name        string
		incidentID  string
		status      string
		assignedTo  string
		metadata    map[string]interface{}
		expectError bool
	}{
		{
			name:       "valid update",
			incidentID: "incident_alert_1234567890",
			status:     "IN_PROGRESS",
			assignedTo: "security_analyst_1",
			metadata: map[string]interface{}{
				"priority": "high",
				"notes":    "Investigating the incident",
			},
			expectError: false,
		},
		{
			name:        "status update only",
			incidentID:  "incident_alert_1234567890",
			status:      "RESOLVED",
			assignedTo:  "",
			metadata:    nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := sm.UpdateIncident(ctx, tt.incidentID, tt.status, tt.assignedTo, tt.metadata)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityMonitor_ResolveIncident(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	tests := []struct {
		name           string
		incidentID     string
		rootCause      string
		remediation    []string
		lessonsLearned []string
		expectError    bool
	}{
		{
			name:       "valid resolution",
			incidentID: "incident_alert_1234567890",
			rootCause:  "Weak password policy allowed brute force attack",
			remediation: []string{
				"Implement stronger password requirements",
				"Enable account lockout after failed attempts",
				"Add rate limiting for login attempts",
			},
			lessonsLearned: []string{
				"Need to review password policies",
				"Should implement additional monitoring",
			},
			expectError: false,
		},
		{
			name:           "minimal resolution",
			incidentID:     "incident_alert_1234567890",
			rootCause:      "False positive",
			remediation:    []string{},
			lessonsLearned: []string{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := sm.ResolveIncident(ctx, tt.incidentID, tt.rootCause, tt.remediation, tt.lessonsLearned)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityMonitor_AddThreatIntelligence(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	tests := []struct {
		name        string
		threatType  string
		source      string
		indicator   string
		confidence  float64
		severity    string
		description string
		tags        []string
		expectError bool
	}{
		{
			name:        "valid threat intelligence",
			threatType:  "ip_address",
			source:      "threat_feed_1",
			indicator:   "192.168.1.100",
			confidence:  0.95,
			severity:    "HIGH",
			description: "Known malicious IP address",
			tags:        []string{"malware", "botnet", "c2"},
			expectError: false,
		},
		{
			name:        "domain threat",
			threatType:  "domain",
			source:      "threat_feed_2",
			indicator:   "malicious-domain.com",
			confidence:  0.85,
			severity:    "MEDIUM",
			description: "Suspicious domain used for phishing",
			tags:        []string{"phishing", "malware"},
			expectError: false,
		},
		{
			name:        "file hash threat",
			threatType:  "file_hash",
			source:      "threat_feed_3",
			indicator:   "a1b2c3d4e5f6...",
			confidence:  1.0,
			severity:    "CRITICAL",
			description: "Known malware file hash",
			tags:        []string{"malware", "ransomware"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			threat, err := sm.AddThreatIntelligence(ctx, tt.threatType, tt.source, tt.indicator, tt.confidence, tt.severity, tt.description, tt.tags)

			require.NoError(t, err)
			assert.NotNil(t, threat)
			assert.Equal(t, tt.threatType, threat.Type)
			assert.Equal(t, tt.source, threat.Source)
			assert.Equal(t, tt.indicator, threat.Indicator)
			assert.Equal(t, tt.confidence, threat.Confidence)
			assert.Equal(t, tt.severity, threat.Severity)
			assert.Equal(t, tt.description, threat.Description)
			assert.Equal(t, tt.tags, threat.Tags)
			assert.False(t, threat.CreatedAt.IsZero())
			assert.False(t, threat.UpdatedAt.IsZero())
		})
	}
}

func TestSecurityMonitor_GetSecurityMetrics(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	ctx := context.Background()
	metrics, err := sm.GetSecurityMetrics(ctx)

	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.False(t, metrics.Timestamp.IsZero())
	assert.GreaterOrEqual(t, metrics.TotalAlerts, 0)
	assert.GreaterOrEqual(t, metrics.CriticalAlerts, 0)
	assert.GreaterOrEqual(t, metrics.HighAlerts, 0)
	assert.GreaterOrEqual(t, metrics.MediumAlerts, 0)
	assert.GreaterOrEqual(t, metrics.LowAlerts, 0)
	assert.GreaterOrEqual(t, metrics.OpenIncidents, 0)
	assert.GreaterOrEqual(t, metrics.ResolvedIncidents, 0)
	assert.GreaterOrEqual(t, metrics.ThreatsBlocked, 0)
	assert.GreaterOrEqual(t, metrics.FalsePositives, 0)
}

func TestSecurityMonitor_GenerateSecurityReport(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	startDate := time.Now().Add(-30 * 24 * time.Hour)
	endDate := time.Now()

	ctx := context.Background()
	report, err := sm.GenerateSecurityReport(ctx, startDate, endDate)

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Contains(t, report, "generated_at")
	assert.Contains(t, report, "period")
	assert.Contains(t, report, "summary")
	assert.Contains(t, report, "threats")
	assert.Contains(t, report, "recommendations")

	// Check period
	period, ok := report["period"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, period, "start_date")
	assert.Contains(t, period, "end_date")

	// Check summary
	summary, ok := report["summary"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, summary, "total_alerts")
	assert.Contains(t, summary, "total_incidents")
	assert.Contains(t, summary, "resolved_incidents")

	// Check recommendations
	recommendations, ok := report["recommendations"].([]string)
	assert.True(t, ok)
	assert.NotEmpty(t, recommendations)
}

func TestSecurityMonitor_DetermineEscalationLevel(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	tests := []struct {
		name     string
		severity string
		expected int
	}{
		{"critical severity", "CRITICAL", 3},
		{"high severity", "HIGH", 2},
		{"medium severity", "MEDIUM", 1},
		{"low severity", "LOW", 1},
		{"unknown severity", "UNKNOWN", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := sm.determineEscalationLevel(tt.severity)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestSecurityMonitor_ShouldCreateIncident(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	tests := []struct {
		name     string
		alert    *SecurityAlert
		expected bool
	}{
		{
			name: "critical alert should create incident",
			alert: &SecurityAlert{
				Type:     "data_breach",
				Severity: "CRITICAL",
			},
			expected: true,
		},
		{
			name: "high severity alert should create incident",
			alert: &SecurityAlert{
				Type:     "brute_force",
				Severity: "HIGH",
			},
			expected: true,
		},
		{
			name: "medium severity alert should not create incident",
			alert: &SecurityAlert{
				Type:     "failed_login",
				Severity: "MEDIUM",
			},
			expected: false,
		},
		{
			name: "low severity alert should not create incident",
			alert: &SecurityAlert{
				Type:     "suspicious_activity",
				Severity: "LOW",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.shouldCreateIncident(tt.alert)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityMonitor_StartMonitoring(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := sm.StartMonitoring(ctx)
	assert.NoError(t, err)

	// Wait for context to timeout to allow goroutines to start
	time.Sleep(50 * time.Millisecond)
}

func TestSecurityMonitor_HandleAlert(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	alert := &SecurityAlert{
		ID:          "alert_test_1234567890",
		Type:        "data_breach",
		Severity:    "CRITICAL",
		Title:       "Test Alert",
		Description: "Test alert description",
		Source:      "test_source",
		Timestamp:   time.Now(),
		Status:      "OPEN",
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx := context.Background()
	sm.handleAlert(ctx, alert)

	// The function should complete without error
	// In a real implementation, this would create an incident and send notifications
}

func TestSecurityMonitor_AutoRespond(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	tests := []struct {
		name     string
		incident *SecurityIncident
	}{
		{
			name: "critical incident auto-response",
			incident: &SecurityIncident{
				ID:       "incident_test_1234567890",
				Severity: "CRITICAL",
			},
		},
		{
			name: "high severity incident auto-response",
			incident: &SecurityIncident{
				ID:       "incident_test_1234567890",
				Severity: "HIGH",
			},
		},
		{
			name: "medium severity incident auto-response",
			incident: &SecurityIncident{
				ID:       "incident_test_1234567890",
				Severity: "MEDIUM",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			sm.autoRespond(ctx, tt.incident)

			// The function should complete without error
			// In a real implementation, this would perform automatic response actions
		})
	}
}

func TestSecurityMonitor_PerformAnomalyDetection(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	ctx := context.Background()
	sm.performAnomalyDetection(ctx)

	// The function should complete without error
	// In a real implementation, this would perform anomaly detection
}

func TestSecurityMonitor_UpdateThreatData(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	ctx := context.Background()
	sm.updateThreatData(ctx)

	// The function should complete without error
	// In a real implementation, this would update threat intelligence data
}

func TestSecurityMonitor_SendNotifications(t *testing.T) {
	mockLogger := &MockLogger{}
	sm := NewSecurityMonitor(nil, mockLogger)

	incident := &SecurityIncident{
		ID:              "incident_test_1234567890",
		Severity:        "HIGH",
		EscalationLevel: 2,
	}

	ctx := context.Background()
	sm.sendNotifications(ctx, incident)

	// The function should complete without error
	// In a real implementation, this would send notifications via configured channels
}
