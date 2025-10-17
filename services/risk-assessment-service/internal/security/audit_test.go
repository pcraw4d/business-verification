package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	InfoLogs  []string
	WarnLogs  []string
	ErrorLogs []string
	DebugLogs []string
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.InfoLogs = append(m.InfoLogs, msg)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.WarnLogs = append(m.WarnLogs, msg)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.ErrorLogs = append(m.ErrorLogs, msg)
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.DebugLogs = append(m.DebugLogs, msg)
}

func TestNewAuditLogger(t *testing.T) {
	tests := []struct {
		name   string
		config *AuditConfig
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "custom config",
			config: &AuditConfig{
				LogLevel:         "DEBUG",
				RetentionPeriod:  24 * time.Hour,
				EncryptLogs:      true,
				IncludeSensitive: true,
				AsyncLogging:     false,
				BatchSize:        50,
				FlushInterval:    1 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			encryptor, err := NewEncryptionManager(nil)
			require.NoError(t, err)

			al := NewAuditLogger(mockLogger, encryptor, tt.config)
			assert.NotNil(t, al)
			assert.Equal(t, mockLogger, al.logger)
			assert.Equal(t, encryptor, al.encryptor)
		})
	}
}

func TestAuditLogger_LogRiskAssessment(t *testing.T) {
	mockLogger := &MockLogger{}
	encryptor, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	al := NewAuditLogger(mockLogger, encryptor, nil)

	tests := []struct {
		name  string
		event *AuditEvent
	}{
		{
			name: "complete event",
			event: &AuditEvent{
				UserID:    "user123",
				SessionID: "session456",
				Action:    "CREATE",
				Resource:  "risk_assessment",
				Result:    "SUCCESS",
				IPAddress: "192.168.1.1",
				UserAgent: "Mozilla/5.0",
				Details: map[string]interface{}{
					"assessment_id": "assess123",
					"risk_score":    0.75,
				},
			},
		},
		{
			name: "minimal event",
			event: &AuditEvent{
				Action: "READ",
			},
		},
		{
			name: "event with sensitive data",
			event: &AuditEvent{
				UserID:   "user123",
				Action:   "UPDATE",
				Resource: "user_profile",
				Details: map[string]interface{}{
					"password": "secret123",
					"ssn":      "123-45-6789",
					"name":     "John Doe",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := al.LogRiskAssessment(ctx, tt.event)
			require.NoError(t, err)

			// Wait a bit for async logging
			time.Sleep(10 * time.Millisecond)

			// Verify event was logged
			assert.NotEmpty(t, tt.event.ID)
			assert.False(t, tt.event.Timestamp.IsZero())
			assert.NotEmpty(t, tt.event.RiskLevel)
			assert.Contains(t, tt.event.Compliance, "SOX")
			assert.Contains(t, tt.event.Compliance, "GDPR")
			assert.Contains(t, tt.event.Compliance, "PCI-DSS")
		})
	}
}

func TestAuditLogger_LogAuthentication(t *testing.T) {
	mockLogger := &MockLogger{}
	al := NewAuditLogger(mockLogger, nil, nil)

	tests := []struct {
		name              string
		userID            string
		action            string
		result            string
		ipAddress         string
		userAgent         string
		expectedRiskLevel string
	}{
		{
			name:              "successful login",
			userID:            "user123",
			action:            "LOGIN",
			result:            "SUCCESS",
			ipAddress:         "192.168.1.1",
			userAgent:         "Mozilla/5.0",
			expectedRiskLevel: "LOW",
		},
		{
			name:              "failed login",
			userID:            "user123",
			action:            "LOGIN",
			result:            "FAILED",
			ipAddress:         "192.168.1.1",
			userAgent:         "Mozilla/5.0",
			expectedRiskLevel: "HIGH",
		},
		{
			name:              "password reset",
			userID:            "user123",
			action:            "PASSWORD_RESET",
			result:            "SUCCESS",
			ipAddress:         "192.168.1.1",
			userAgent:         "Mozilla/5.0",
			expectedRiskLevel: "MEDIUM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := al.LogAuthentication(ctx, tt.userID, tt.action, tt.result, tt.ipAddress, tt.userAgent)
			require.NoError(t, err)

			// Wait a bit for async logging
			time.Sleep(10 * time.Millisecond)

			// Verify authentication event was logged
			assert.True(t, len(mockLogger.InfoLogs) > 0 || len(mockLogger.WarnLogs) > 0 || len(mockLogger.ErrorLogs) > 0)
		})
	}
}

func TestAuditLogger_LogDataAccess(t *testing.T) {
	mockLogger := &MockLogger{}
	al := NewAuditLogger(mockLogger, nil, nil)

	tests := []struct {
		name              string
		userID            string
		resource          string
		resourceID        string
		action            string
		result            string
		expectedRiskLevel string
	}{
		{
			name:              "read operation",
			userID:            "user123",
			resource:          "risk_assessment",
			resourceID:        "assess123",
			action:            "READ",
			result:            "SUCCESS",
			expectedRiskLevel: "LOW",
		},
		{
			name:              "delete operation",
			userID:            "user123",
			resource:          "risk_assessment",
			resourceID:        "assess123",
			action:            "DELETE",
			result:            "SUCCESS",
			expectedRiskLevel: "HIGH",
		},
		{
			name:              "sensitive data access",
			userID:            "user123",
			resource:          "sensitive_data",
			resourceID:        "data123",
			action:            "READ",
			result:            "SUCCESS",
			expectedRiskLevel: "MEDIUM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := al.LogDataAccess(ctx, tt.userID, tt.resource, tt.resourceID, tt.action, tt.result)
			require.NoError(t, err)

			// Wait a bit for async logging
			time.Sleep(10 * time.Millisecond)

			// Verify data access event was logged
			assert.True(t, len(mockLogger.InfoLogs) > 0 || len(mockLogger.WarnLogs) > 0 || len(mockLogger.ErrorLogs) > 0)
		})
	}
}

func TestAuditLogger_LogConfigurationChange(t *testing.T) {
	mockLogger := &MockLogger{}
	al := NewAuditLogger(mockLogger, nil, nil)

	ctx := context.Background()
	details := map[string]interface{}{
		"config_type": "rate_limits",
		"old_value":   "100",
		"new_value":   "200",
	}

	err := al.LogConfigurationChange(ctx, "admin123", "rate_limits", "UPDATE", "SUCCESS", details)
	require.NoError(t, err)

	// Wait a bit for async logging
	time.Sleep(10 * time.Millisecond)

	// Verify configuration change was logged
	assert.True(t, len(mockLogger.InfoLogs) > 0 || len(mockLogger.WarnLogs) > 0 || len(mockLogger.ErrorLogs) > 0)
}

func TestAuditLogger_LogSecurityEvent(t *testing.T) {
	mockLogger := &MockLogger{}
	al := NewAuditLogger(mockLogger, nil, nil)

	tests := []struct {
		name              string
		eventType         string
		severity          string
		description       string
		details           map[string]interface{}
		expectedRiskLevel string
	}{
		{
			name:        "critical security event",
			eventType:   "INTRUSION_DETECTED",
			severity:    "CRITICAL",
			description: "Multiple failed login attempts detected",
			details: map[string]interface{}{
				"attempts": 10,
				"ip":       "192.168.1.100",
			},
			expectedRiskLevel: "CRITICAL",
		},
		{
			name:        "medium security event",
			eventType:   "SUSPICIOUS_ACTIVITY",
			severity:    "MEDIUM",
			description: "Unusual data access pattern detected",
			details: map[string]interface{}{
				"user_id": "user123",
				"pattern": "bulk_download",
			},
			expectedRiskLevel: "MEDIUM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := al.LogSecurityEvent(ctx, tt.eventType, tt.severity, tt.description, tt.details)
			require.NoError(t, err)

			// Wait a bit for async logging
			time.Sleep(10 * time.Millisecond)

			// Verify security event was logged
			assert.True(t, len(mockLogger.InfoLogs) > 0 || len(mockLogger.WarnLogs) > 0 || len(mockLogger.ErrorLogs) > 0)
		})
	}
}

func TestAuditLogger_LogComplianceEvent(t *testing.T) {
	mockLogger := &MockLogger{}
	al := NewAuditLogger(mockLogger, nil, nil)

	ctx := context.Background()
	details := map[string]interface{}{
		"data_type": "personal_data",
		"action":    "export",
		"user_id":   "user123",
	}

	err := al.LogComplianceEvent(ctx, "GDPR", "DATA_EXPORT", "SUCCESS", details)
	require.NoError(t, err)

	// Wait a bit for async logging
	time.Sleep(10 * time.Millisecond)

	// Verify compliance event was logged
	assert.True(t, len(mockLogger.InfoLogs) > 0 || len(mockLogger.WarnLogs) > 0 || len(mockLogger.ErrorLogs) > 0)
}

func TestAuditLogger_QueryAuditLogs(t *testing.T) {
	mockLogger := &MockLogger{}
	al := NewAuditLogger(mockLogger, nil, nil)

	ctx := context.Background()
	filters := &AuditFilters{
		UserID:    "user123",
		Action:    "CREATE",
		StartTime: time.Now().Add(-24 * time.Hour),
		EndTime:   time.Now(),
		Limit:     100,
	}

	events, err := al.QueryAuditLogs(ctx, filters)
	require.NoError(t, err)
	assert.NotNil(t, events)
}

func TestAuditLogger_EncryptLogs(t *testing.T) {
	mockLogger := &MockLogger{}
	encryptor, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	config := &AuditConfig{
		EncryptLogs: true,
	}
	al := NewAuditLogger(mockLogger, encryptor, config)

	ctx := context.Background()
	event := &AuditEvent{
		UserID:   "user123",
		Action:   "CREATE",
		Resource: "sensitive_data",
		Details: map[string]interface{}{
			"data": "sensitive information",
		},
	}

	err = al.LogRiskAssessment(ctx, event)
	require.NoError(t, err)

	// Wait a bit for async logging
	time.Sleep(10 * time.Millisecond)

	// Verify event was encrypted
	assert.True(t, event.Encrypted)
	assert.NotNil(t, event.Details["encrypted_data"])
	assert.NotNil(t, event.Details["key_id"])
	assert.NotNil(t, event.Details["algorithm"])
	assert.NotNil(t, event.Details["iv"])
}

func TestAuditLogger_SanitizeEvent(t *testing.T) {
	mockLogger := &MockLogger{}
	config := &AuditConfig{
		IncludeSensitive: false,
	}
	al := NewAuditLogger(mockLogger, nil, config)

	event := &AuditEvent{
		UserID:   "user123",
		Action:   "UPDATE",
		Resource: "user_profile",
		Details: map[string]interface{}{
			"password":    "secret123",
			"ssn":         "123-45-6789",
			"credit_card": "4111-1111-1111-1111",
			"name":        "John Doe",
			"email":       "john@example.com",
		},
	}

	sanitized := al.sanitizeEvent(event)

	// Verify sensitive fields were removed
	assert.Nil(t, sanitized.Details["password"])
	assert.Nil(t, sanitized.Details["ssn"])
	assert.Nil(t, sanitized.Details["credit_card"])

	// Verify non-sensitive fields remain
	assert.Equal(t, "John Doe", sanitized.Details["name"])
	assert.Equal(t, "john@example.com", sanitized.Details["email"])
}

func TestAuditLogger_DetermineRiskLevels(t *testing.T) {
	mockLogger := &MockLogger{}
	al := NewAuditLogger(mockLogger, nil, nil)

	tests := []struct {
		name     string
		action   string
		result   string
		resource string
		expected string
	}{
		// Authentication risk levels
		{"failed login", "LOGIN", "FAILED", "", "HIGH"},
		{"successful login", "LOGIN", "SUCCESS", "", "LOW"},
		{"password reset", "PASSWORD_RESET", "SUCCESS", "", "MEDIUM"},
		{"account locked", "ACCOUNT_LOCKED", "SUCCESS", "", "MEDIUM"},

		// Data access risk levels
		{"delete operation", "DELETE", "SUCCESS", "data", "HIGH"},
		{"update operation", "UPDATE", "SUCCESS", "data", "MEDIUM"},
		{"read operation", "READ", "SUCCESS", "data", "LOW"},
		{"sensitive data read", "READ", "SUCCESS", "sensitive_data", "MEDIUM"},
		{"financial data read", "READ", "SUCCESS", "financial_data", "MEDIUM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var riskLevel string
			if tt.resource == "" {
				// Test authentication risk level
				riskLevel = al.determineAuthRiskLevel(tt.action, tt.result)
			} else {
				// Test data access risk level
				riskLevel = al.determineDataAccessRiskLevel(tt.action, tt.resource)
			}
			assert.Equal(t, tt.expected, riskLevel)
		})
	}
}

func TestAuditLogger_MapSeverityToRiskLevel(t *testing.T) {
	mockLogger := &MockLogger{}
	al := NewAuditLogger(mockLogger, nil, nil)

	tests := []struct {
		name     string
		severity string
		expected string
	}{
		{"critical severity", "CRITICAL", "CRITICAL"},
		{"high severity", "HIGH", "HIGH"},
		{"medium severity", "MEDIUM", "MEDIUM"},
		{"low severity", "LOW", "LOW"},
		{"unknown severity", "UNKNOWN", "MEDIUM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			riskLevel := al.mapSeverityToRiskLevel(tt.severity)
			assert.Equal(t, tt.expected, riskLevel)
		})
	}
}

func TestAuditLogger_AuditMiddleware(t *testing.T) {
	mockLogger := &MockLogger{}
	al := NewAuditLogger(mockLogger, nil, nil)

	// Create middleware
	middleware := al.AuditMiddleware()

	// Create a simple handler
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	// Wrap handler with middleware
	wrappedHandler := middleware(handler)

	// Create context with audit information
	ctx := context.WithValue(context.Background(), "user_id", "user123")
	ctx = context.WithValue(ctx, "session_id", "session456")
	ctx = context.WithValue(ctx, "ip_address", "192.168.1.1")
	ctx = context.WithValue(ctx, "user_agent", "Mozilla/5.0")

	// Execute wrapped handler
	resp, err := wrappedHandler(ctx, "test request")
	require.NoError(t, err)
	assert.Equal(t, "success", resp)

	// Wait a bit for async logging
	time.Sleep(10 * time.Millisecond)

	// Verify audit events were logged
	assert.True(t, len(mockLogger.InfoLogs) > 0 || len(mockLogger.WarnLogs) > 0 || len(mockLogger.ErrorLogs) > 0)
}

func TestAuditLogger_ContextHelpers(t *testing.T) {
	ctx := context.Background()

	// Test empty context
	assert.Empty(t, getUserIDFromContext(ctx))
	assert.Empty(t, getSessionIDFromContext(ctx))
	assert.Empty(t, getIPAddressFromContext(ctx))
	assert.Empty(t, getUserAgentFromContext(ctx))

	// Test context with values
	ctx = context.WithValue(ctx, "user_id", "user123")
	ctx = context.WithValue(ctx, "session_id", "session456")
	ctx = context.WithValue(ctx, "ip_address", "192.168.1.1")
	ctx = context.WithValue(ctx, "user_agent", "Mozilla/5.0")

	assert.Equal(t, "user123", getUserIDFromContext(ctx))
	assert.Equal(t, "session456", getSessionIDFromContext(ctx))
	assert.Equal(t, "192.168.1.1", getIPAddressFromContext(ctx))
	assert.Equal(t, "Mozilla/5.0", getUserAgentFromContext(ctx))
}
