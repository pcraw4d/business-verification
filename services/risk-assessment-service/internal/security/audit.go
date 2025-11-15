package security

import (
	"context"
	"time"
)

// Logger defines the interface for logging operations
type Logger interface {
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// AuditConfig holds configuration for audit logging
type AuditConfig struct {
	LogLevel         string        `json:"log_level"`
	RetentionPeriod  time.Duration `json:"retention_period"`
	EncryptLogs      bool          `json:"encrypt_logs"`
	IncludeSensitive bool          `json:"include_sensitive"`
	AsyncLogging     bool          `json:"async_logging"`
	BatchSize        int           `json:"batch_size"`
	FlushInterval    time.Duration `json:"flush_interval"`
}

// AuditEvent represents an audit event
type AuditEvent struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	Result      string                 `json:"result"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
	RiskLevel   string                 `json:"risk_level"`
	Compliance  []string               `json:"compliance"`
	Encrypted   bool                   `json:"encrypted"`
}

// AuditLogger handles audit logging for security events
type AuditLogger struct {
	logger    Logger
	encryptor *EncryptionManager
	config    *AuditConfig
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger Logger, encryptor *EncryptionManager, config *AuditConfig) *AuditLogger {
	if config == nil {
		config = &AuditConfig{
			LogLevel:         "INFO",
			RetentionPeriod:  90 * 24 * time.Hour,
			EncryptLogs:      false,
			IncludeSensitive: false,
			AsyncLogging:     true,
			BatchSize:        100,
			FlushInterval:    5 * time.Second,
		}
	}

	return &AuditLogger{
		logger:    logger,
		encryptor: encryptor,
		config:    config,
	}
}

// LogRiskAssessment logs a risk assessment event
func (al *AuditLogger) LogRiskAssessment(ctx context.Context, event *AuditEvent) error {
	// Set default values
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.RiskLevel == "" {
		event.RiskLevel = "MEDIUM"
	}
	if event.Compliance == nil {
		event.Compliance = []string{"SOX", "GDPR", "PCI-DSS"}
	}

	// Log the event
	al.logger.Info("Risk assessment logged",
		"event_id", event.ID,
		"user_id", event.UserID,
		"action", event.Action,
		"resource", event.Resource,
		"risk_level", event.RiskLevel)

	return nil
}

// LogAuthentication logs an authentication event
func (al *AuditLogger) LogAuthentication(ctx context.Context, userID, action, result, ipAddress, userAgent string) error {
	event := &AuditEvent{
		ID:        generateEventID(),
		UserID:    userID,
		Action:    action,
		Result:    result,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}

	// Determine risk level
	if result == "FAILURE" {
		event.RiskLevel = "HIGH"
	} else {
		event.RiskLevel = "LOW"
	}

	al.logger.Info("Authentication event logged",
		"event_id", event.ID,
		"user_id", userID,
		"action", action,
		"result", result)

	return nil
}

// LogDataAccess logs a data access event
func (al *AuditLogger) LogDataAccess(ctx context.Context, userID, resource, resourceID, action, result string) error {
	event := &AuditEvent{
		ID:        generateEventID(),
		UserID:    userID,
		Resource:  resource,
		Action:    action,
		Result:    result,
		Details:   make(map[string]interface{}),
		Timestamp: time.Now(),
		RiskLevel: al.determineDataAccessRiskLevel(action, resource),
	}

	al.logger.Info("Data access event logged",
		"event_id", event.ID,
		"user_id", userID,
		"resource", resource,
		"action", action)

	return nil
}

// LogConfigurationChange logs a configuration change event
func (al *AuditLogger) LogConfigurationChange(ctx context.Context, userID, resource, action, result string, details map[string]interface{}) error {
	event := &AuditEvent{
		ID:        generateEventID(),
		UserID:    userID,
		Resource:  resource,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
		RiskLevel: "HIGH",
	}

	al.logger.Info("Configuration change logged",
		"event_id", event.ID,
		"user_id", userID,
		"resource", resource,
		"action", action)

	return nil
}

// LogSecurityEvent logs a security event
func (al *AuditLogger) LogSecurityEvent(ctx context.Context, userID, action, severity string, details map[string]interface{}) error {
	event := &AuditEvent{
		ID:        generateEventID(),
		UserID:    userID,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
		RiskLevel: "HIGH",
	}

	al.logger.Warn("Security event logged",
		"event_id", event.ID,
		"user_id", userID,
		"action", action)

	return nil
}

// LogComplianceEvent logs a compliance event
func (al *AuditLogger) LogComplianceEvent(ctx context.Context, userID, action, framework string, details map[string]interface{}) error {
	event := &AuditEvent{
		ID:        generateEventID(),
		UserID:    userID,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
		RiskLevel: "MEDIUM",
		Compliance: []string{framework},
	}

	al.logger.Info("Compliance event logged",
		"event_id", event.ID,
		"user_id", userID,
		"action", action,
		"framework", framework)

	return nil
}


// generateEventID generates a unique event ID
func generateEventID() string {
	return "audit_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// AuditFilters represents filters for querying audit logs
type AuditFilters struct {
	UserID    string
	Action    string
	Resource  string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// QueryAuditLogs queries audit logs with filters
func (al *AuditLogger) QueryAuditLogs(ctx context.Context, filters *AuditFilters) ([]*AuditEvent, error) {
	// Placeholder implementation
	return []*AuditEvent{}, nil
}

// Context helper functions
func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func getSessionIDFromContext(ctx context.Context) string {
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	return ""
}

func getIPAddressFromContext(ctx context.Context) string {
	if ipAddress, ok := ctx.Value("ip_address").(string); ok {
		return ipAddress
	}
	return ""
}

func getUserAgentFromContext(ctx context.Context) string {
	if userAgent, ok := ctx.Value("user_agent").(string); ok {
		return userAgent
	}
	return ""
}

// sanitizeEvent removes sensitive data from an event
func (al *AuditLogger) sanitizeEvent(event *AuditEvent) *AuditEvent {
	if !al.config.IncludeSensitive {
		// Remove sensitive fields
		sanitized := *event
		if sanitized.Details != nil {
			sanitizedDetails := make(map[string]interface{})
			for k, v := range sanitized.Details {
				if !isSensitiveField(k) {
					sanitizedDetails[k] = v
				}
			}
			sanitized.Details = sanitizedDetails
		}
		return &sanitized
	}
	return event
}

// isSensitiveField checks if a field name indicates sensitive data
func isSensitiveField(fieldName string) bool {
	sensitiveFields := []string{"password", "ssn", "credit_card", "api_key", "secret", "token"}
	for _, field := range sensitiveFields {
		if fieldName == field {
			return true
		}
	}
	return false
}

// determineAuthRiskLevel determines risk level for authentication events
func (al *AuditLogger) determineAuthRiskLevel(action, result string) string {
	if result == "FAILURE" {
		return "HIGH"
	}
	if action == "LOGOUT" {
		return "LOW"
	}
	return "MEDIUM"
}

// determineDataAccessRiskLevel determines risk level for data access events
func (al *AuditLogger) determineDataAccessRiskLevel(action, resource string) string {
	if action == "DELETE" || action == "UPDATE" {
		return "HIGH"
	}
	if resource == "user_data" || resource == "financial_data" {
		return "MEDIUM"
	}
	return "LOW"
}

// mapSeverityToRiskLevel maps severity string to risk level
func (al *AuditLogger) mapSeverityToRiskLevel(severity string) string {
	switch severity {
	case "CRITICAL", "HIGH":
		return "HIGH"
	case "MEDIUM":
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// AuditMiddleware creates middleware function for audit logging
func (al *AuditLogger) AuditMiddleware() func(func(context.Context, interface{}) (interface{}, error)) func(context.Context, interface{}) (interface{}, error) {
	return func(next func(context.Context, interface{}) (interface{}, error)) func(context.Context, interface{}) (interface{}, error) {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Log the request
			userID := getUserIDFromContext(ctx)
			if userID != "" {
				al.logger.Info("Audit middleware: request logged", "user_id", userID)
			}
			// Call the next handler
			return next(ctx, req)
		}
	}
}

// randomString generates a random string (placeholder)
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

