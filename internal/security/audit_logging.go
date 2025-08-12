package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/pkg/validators"
)

// AuditLoggingSystem provides comprehensive audit logging capabilities
type AuditLoggingSystem struct {
	logger         *observability.Logger
	config         AuditLoggingConfig
	fileLogger     *AuditFileLogger
	databaseLogger *AuditDatabaseLogger
	eventQueue     chan AuditEvent
	mutex          sync.RWMutex
	eventHandlers  map[EventType][]func(AuditEvent)
	metrics        *AuditMetrics
}

// AuditLoggingConfig defines configuration for audit logging
type AuditLoggingConfig struct {
	Enabled              bool          `json:"enabled"`
	LogLevel             string        `json:"log_level"`
	RetentionDays        int           `json:"retention_days"`
	MaxFileSize          int64         `json:"max_file_size_mb"`
	MaxFiles             int           `json:"max_files"`
	LogDirectory         string        `json:"log_directory"`
	DatabaseEnabled      bool          `json:"database_enabled"`
	FileEnabled          bool          `json:"file_enabled"`
	ConsoleEnabled       bool          `json:"console_enabled"`
	EventQueueSize       int           `json:"event_queue_size"`
	FlushInterval        time.Duration `json:"flush_interval"`
	CompressionEnabled   bool          `json:"compression_enabled"`
	EncryptionEnabled    bool          `json:"encryption_enabled"`
	EncryptionKey        string        `json:"encryption_key,omitempty"`
	ComplianceFrameworks []string      `json:"compliance_frameworks"`
}

// AuditEvent extends BaseEvent with audit-specific fields
type AuditEvent struct {
	BaseEvent
	SessionID      string   `json:"session_id,omitempty"`
	UserAgent      string   `json:"user_agent,omitempty"`
	Resource       string   `json:"resource,omitempty"`
	Action         string   `json:"action,omitempty"`
	Result         string   `json:"result,omitempty"`
	ComplianceTags []string `json:"compliance_tags,omitempty"`
	RiskScore      float64  `json:"risk_score,omitempty"`
	CorrelationID  string   `json:"correlation_id,omitempty"`
	RequestID      string   `json:"request_id,omitempty"`
}

// AuditFileLogger provides file-based audit logging
type AuditFileLogger struct {
	logger    *observability.Logger
	config    AuditLoggingConfig
	file      *os.File
	filePath  string
	mutex     sync.Mutex
	encoder   *json.Encoder
	fileSize  int64
	fileCount int
}

// AuditDatabaseLogger provides database-based audit logging
type AuditDatabaseLogger struct {
	logger *observability.Logger
	config AuditLoggingConfig
	// In a real implementation, this would have database connection
	// For now, we'll simulate database operations
}

// AuditMetrics represents audit logging metrics
type AuditMetrics struct {
	TotalEvents      int64                   `json:"total_events"`
	EventsByType     map[EventType]int64     `json:"events_by_type"`
	EventsByCategory map[EventCategory]int64 `json:"events_by_category"`
	EventsBySeverity map[Severity]int64      `json:"events_by_severity"`
	EventsByUser     map[string]int64        `json:"events_by_user"`
	EventsByResource map[string]int64        `json:"events_by_resource"`
	EventsByResult   map[string]int64        `json:"events_by_result"`
	LastEventTime    time.Time               `json:"last_event_time"`
	AverageEventRate float64                 `json:"average_event_rate"`
	ErrorCount       int64                   `json:"error_count"`
	LastUpdated      time.Time               `json:"last_updated"`
}

// NewAuditLoggingSystem creates a new audit logging system
func NewAuditLoggingSystem(logger *observability.Logger, config AuditLoggingConfig) *AuditLoggingSystem {
	als := &AuditLoggingSystem{
		logger:        logger,
		config:        config,
		eventQueue:    make(chan AuditEvent, config.EventQueueSize),
		eventHandlers: make(map[EventType][]func(AuditEvent)),
		metrics: &AuditMetrics{
			EventsByType:     make(map[EventType]int64),
			EventsByCategory: make(map[EventCategory]int64),
			EventsBySeverity: make(map[Severity]int64),
			EventsByUser:     make(map[string]int64),
			EventsByResource: make(map[string]int64),
			EventsByResult:   make(map[string]int64),
		},
	}

	// Initialize file logger if enabled
	if config.FileEnabled {
		als.fileLogger = NewAuditFileLogger(logger, config)
	}

	// Initialize database logger if enabled
	if config.DatabaseEnabled {
		als.databaseLogger = NewAuditDatabaseLogger(logger, config)
	}

	// Start event processing
	go als.eventProcessor()

	// Start metrics collection
	go als.metricsCollector()

	// Start cleanup routine
	go als.cleanupRoutine()

	return als
}

// LogEvent logs an audit event
func (als *AuditLoggingSystem) LogEvent(ctx context.Context, event AuditEvent) error {
	if !als.config.Enabled {
		return nil
	}

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = observability.GenerateRequestID()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Set correlation ID from context if available
	if event.CorrelationID == "" {
		if correlationID := ctx.Value("correlation_id"); correlationID != nil {
			if id, ok := correlationID.(string); ok {
				event.CorrelationID = id
			}
		}
	}

	// Set request ID from context if available
	if event.RequestID == "" {
		if requestID := ctx.Value("request_id"); requestID != nil {
			if id, ok := requestID.(string); ok {
				event.RequestID = id
			}
		}
	}

	// Add compliance tags based on event type
	event.ComplianceTags = als.getComplianceTags(event)

	// Calculate risk score
	event.RiskScore = als.calculateRiskScore(event)

	// Send event to queue
	select {
	case als.eventQueue <- event:
		return nil
	default:
		// Queue is full, log error
		als.logger.Error("Audit event queue is full, dropping event",
			"event_id", event.ID,
			"event_type", event.EventType,
		)
		return fmt.Errorf("audit event queue is full")
	}
}

// LogSecurityEvent logs a security-related audit event
func (als *AuditLoggingSystem) LogSecurityEvent(ctx context.Context, eventType EventType, userID, resource, action, result string, details map[string]interface{}) error {
	event := AuditEvent{
		BaseEvent: BaseEvent{
			EventType:   eventType,
			Category:    CategorySecurity,
			Severity:    als.getSecurityEventSeverity(eventType),
			UserID:      userID,
			Details:     details,
			Description: als.getSecurityEventDescription(eventType, resource, action, result),
		},
		Resource: resource,
		Action:   action,
		Result:   result,
	}

	return als.LogEvent(ctx, event)
}

// LogAuthenticationEvent logs an authentication-related audit event
func (als *AuditLoggingSystem) LogAuthenticationEvent(ctx context.Context, eventType EventType, userID, result string, details map[string]interface{}) error {
	event := AuditEvent{
		BaseEvent: BaseEvent{
			EventType:   eventType,
			Category:    CategoryAuthentication,
			Severity:    als.getAuthenticationEventSeverity(eventType, result),
			UserID:      userID,
			Details:     details,
			Description: als.getAuthenticationEventDescription(eventType, userID, result),
		},
		Result: result,
	}

	return als.LogEvent(ctx, event)
}

// LogDataAccessEvent logs a data access-related audit event
func (als *AuditLoggingSystem) LogDataAccessEvent(ctx context.Context, eventType EventType, userID, resource, action, result string, details map[string]interface{}) error {
	event := AuditEvent{
		BaseEvent: BaseEvent{
			EventType:   eventType,
			Category:    CategoryDataAccess,
			Severity:    als.getDataAccessEventSeverity(eventType, action),
			UserID:      userID,
			Details:     details,
			Description: als.getDataAccessEventDescription(eventType, userID, resource, action, result),
		},
		Resource: resource,
		Action:   action,
		Result:   result,
	}

	return als.LogEvent(ctx, event)
}

// LogSystemEvent logs a system-related audit event
func (als *AuditLoggingSystem) LogSystemEvent(ctx context.Context, eventType EventType, description string, details map[string]interface{}) error {
	event := AuditEvent{
		BaseEvent: BaseEvent{
			EventType:   eventType,
			Category:    CategorySystem,
			Severity:    als.getSystemEventSeverity(eventType),
			Description: description,
			Details:     details,
		},
	}

	return als.LogEvent(ctx, event)
}

// LogComplianceEvent logs a compliance-related audit event
func (als *AuditLoggingSystem) LogComplianceEvent(ctx context.Context, eventType EventType, description string, details map[string]interface{}) error {
	event := AuditEvent{
		BaseEvent: BaseEvent{
			EventType:   eventType,
			Category:    CategoryCompliance,
			Severity:    als.getComplianceEventSeverity(eventType),
			Description: description,
			Details:     details,
		},
	}

	return als.LogEvent(ctx, event)
}

// GetAuditEvents retrieves audit events with optional filtering
func (als *AuditLoggingSystem) GetAuditEvents(ctx context.Context, filters map[string]interface{}) ([]AuditEvent, error) {
	// In a real implementation, this would query the database
	// For now, we'll return an empty slice
	return []AuditEvent{}, nil
}

// GetAuditMetrics retrieves audit logging metrics
func (als *AuditLoggingSystem) GetAuditMetrics(ctx context.Context) (*AuditMetrics, error) {
	als.mutex.RLock()
	defer als.mutex.RUnlock()

	// Create a copy of metrics
	metrics := &AuditMetrics{
		TotalEvents:      als.metrics.TotalEvents,
		EventsByType:     make(map[EventType]int64),
		EventsByCategory: make(map[EventCategory]int64),
		EventsBySeverity: make(map[Severity]int64),
		EventsByUser:     make(map[string]int64),
		EventsByResource: make(map[string]int64),
		EventsByResult:   make(map[string]int64),
		LastEventTime:    als.metrics.LastEventTime,
		AverageEventRate: als.metrics.AverageEventRate,
		ErrorCount:       als.metrics.ErrorCount,
		LastUpdated:      als.metrics.LastUpdated,
	}

	// Copy maps
	for k, v := range als.metrics.EventsByType {
		metrics.EventsByType[k] = v
	}
	for k, v := range als.metrics.EventsByCategory {
		metrics.EventsByCategory[k] = v
	}
	for k, v := range als.metrics.EventsBySeverity {
		metrics.EventsBySeverity[k] = v
	}
	for k, v := range als.metrics.EventsByUser {
		metrics.EventsByUser[k] = v
	}
	for k, v := range als.metrics.EventsByResource {
		metrics.EventsByResource[k] = v
	}
	for k, v := range als.metrics.EventsByResult {
		metrics.EventsByResult[k] = v
	}

	return metrics, nil
}

// ExportAuditLogs exports audit logs to JSON format
func (als *AuditLoggingSystem) ExportAuditLogs(ctx context.Context, filters map[string]interface{}) ([]byte, error) {
	events, err := als.GetAuditEvents(ctx, filters)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(events, "", "  ")
}

// RegisterEventHandler registers a handler for specific event types
func (als *AuditLoggingSystem) RegisterEventHandler(eventType EventType, handler func(AuditEvent)) {
	als.mutex.Lock()
	defer als.mutex.Unlock()

	als.eventHandlers[eventType] = append(als.eventHandlers[eventType], handler)
}

// eventProcessor processes events from the queue
func (als *AuditLoggingSystem) eventProcessor() {
	for event := range als.eventQueue {
		// Update metrics
		als.updateMetrics(event)

		// Log to file if enabled
		if als.fileLogger != nil {
			if err := als.fileLogger.LogEvent(event); err != nil {
				als.logger.Error("Failed to log event to file",
					"error", err,
					"event_id", event.ID,
				)
			}
		}

		// Log to database if enabled
		if als.databaseLogger != nil {
			if err := als.databaseLogger.LogEvent(event); err != nil {
				als.logger.Error("Failed to log event to database",
					"error", err,
					"event_id", event.ID,
				)
			}
		}

		// Log to console if enabled
		if als.config.ConsoleEnabled {
			als.logger.Info("Audit event",
				"event_id", event.ID,
				"event_type", event.EventType,
				"category", event.Category,
				"severity", event.Severity,
				"user_id", event.UserID,
				"resource", event.Resource,
				"action", event.Action,
				"result", event.Result,
				"description", event.Description,
			)
		}

		// Call event handlers
		als.callEventHandlers(event)
	}
}

// updateMetrics updates audit metrics
func (als *AuditLoggingSystem) updateMetrics(event AuditEvent) {
	als.mutex.Lock()
	defer als.mutex.Unlock()

	als.metrics.TotalEvents++
	als.metrics.EventsByType[event.EventType]++
	als.metrics.EventsByCategory[event.Category]++
	als.metrics.EventsBySeverity[event.Severity]++

	if event.UserID != "" {
		als.metrics.EventsByUser[event.UserID]++
	}

	if event.Resource != "" {
		als.metrics.EventsByResource[event.Resource]++
	}

	if event.Result != "" {
		als.metrics.EventsByResult[event.Result]++
	}

	als.metrics.LastEventTime = event.Timestamp
	als.metrics.LastUpdated = time.Now()
}

// callEventHandlers calls registered event handlers
func (als *AuditLoggingSystem) callEventHandlers(event AuditEvent) {
	als.mutex.RLock()
	handlers := als.eventHandlers[event.EventType]
	als.mutex.RUnlock()

	for _, handler := range handlers {
		// Call handler in a goroutine to avoid blocking
		go func(h func(AuditEvent), e AuditEvent) {
			defer func() {
				if r := recover(); r != nil {
					als.logger.Error("Event handler panicked",
						"error", r,
						"event_id", e.ID,
					)
				}
			}()
			h(e)
		}(handler, event)
	}
}

// metricsCollector collects and updates metrics periodically
func (als *AuditLoggingSystem) metricsCollector() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		als.updateAverageEventRate()
	}
}

// updateAverageEventRate updates the average event rate
func (als *AuditLoggingSystem) updateAverageEventRate() {
	als.mutex.Lock()
	defer als.mutex.Unlock()

	// Calculate average event rate (events per minute)
	if als.metrics.TotalEvents > 0 {
		timeSinceFirst := time.Since(als.metrics.LastEventTime)
		if timeSinceFirst > 0 {
			als.metrics.AverageEventRate = float64(als.metrics.TotalEvents) / timeSinceFirst.Minutes()
		}
	}
}

// cleanupRoutine periodically cleans up old audit data
func (als *AuditLoggingSystem) cleanupRoutine() {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	defer ticker.Stop()

	for range ticker.C {
		als.cleanup()
	}
}

// cleanup removes old audit data based on retention policy
func (als *AuditLoggingSystem) cleanup() {
	// Clean up old files if file logging is enabled
	if als.fileLogger != nil {
		als.fileLogger.cleanup()
	}

	// Clean up old database records if database logging is enabled
	if als.databaseLogger != nil {
		als.databaseLogger.cleanup()
	}
}

// Helper methods for determining event properties
func (als *AuditLoggingSystem) getSecurityEventSeverity(eventType EventType) Severity {
	switch eventType {
	case EventTypeVulnerabilityDetected, EventTypeThreatDetected:
		return SeverityHigh
	case EventTypeIncidentReported:
		return SeverityMedium
	case EventTypeAlertGenerated:
		return SeverityLow
	default:
		return SeverityInfo
	}
}

func (als *AuditLoggingSystem) getAuthenticationEventSeverity(eventType EventType, result string) Severity {
	switch eventType {
	case EventTypeLoginFailed, EventTypeAccountLocked:
		return SeverityMedium
	case EventTypeLogin:
		if result == "success" {
			return SeverityInfo
		}
		return SeverityLow
	default:
		return SeverityInfo
	}
}

func (als *AuditLoggingSystem) getDataAccessEventSeverity(eventType EventType, action string) Severity {
	switch eventType {
	case EventTypeDataDelete:
		return SeverityHigh
	case EventTypeDataWrite, EventTypeDataExport:
		return SeverityMedium
	case EventTypeDataRead:
		return SeverityLow
	default:
		return SeverityInfo
	}
}

func (als *AuditLoggingSystem) getSystemEventSeverity(eventType EventType) Severity {
	switch eventType {
	case EventTypeSystemStop:
		return SeverityHigh
	case EventTypeConfigurationChange:
		return SeverityMedium
	default:
		return SeverityInfo
	}
}

func (als *AuditLoggingSystem) getComplianceEventSeverity(eventType EventType) Severity {
	switch eventType {
	case EventTypeComplianceViolation:
		return SeverityHigh
	case EventTypeComplianceCheck:
		return SeverityLow
	default:
		return SeverityInfo
	}
}

func (als *AuditLoggingSystem) getSecurityEventDescription(eventType EventType, resource, action, result string) string {
	switch eventType {
	case EventTypeVulnerabilityDetected:
		return fmt.Sprintf("Vulnerability detected in %s", resource)
	case EventTypeThreatDetected:
		return fmt.Sprintf("Security threat detected: %s", action)
	case EventTypeIncidentReported:
		return fmt.Sprintf("Security incident reported: %s", action)
	case EventTypeAlertGenerated:
		return fmt.Sprintf("Security alert generated for %s", resource)
	default:
		return fmt.Sprintf("Security event: %s on %s - %s", action, resource, result)
	}
}

func (als *AuditLoggingSystem) getAuthenticationEventDescription(eventType EventType, userID, result string) string {
	switch eventType {
	case EventTypeLogin:
		return fmt.Sprintf("User %s login attempt - %s", userID, result)
	case EventTypeLogout:
		return fmt.Sprintf("User %s logged out", userID)
	case EventTypeLoginFailed:
		return fmt.Sprintf("Failed login attempt for user %s", userID)
	case EventTypeAccountLocked:
		return fmt.Sprintf("Account locked for user %s", userID)
	default:
		return fmt.Sprintf("Authentication event: %s for user %s - %s", eventType, userID, result)
	}
}

func (als *AuditLoggingSystem) getDataAccessEventDescription(eventType EventType, userID, resource, action, result string) string {
	return fmt.Sprintf("Data access: User %s performed %s on %s - %s", userID, action, resource, result)
}

func (als *AuditLoggingSystem) getComplianceTags(event AuditEvent) []string {
	var tags []string

	// Add framework-specific tags based on event type and category
	for _, framework := range als.config.ComplianceFrameworks {
		switch framework {
		case "SOC2":
			if event.Category == CategoryAuthentication || event.Category == CategoryAuthorization {
				tags = append(tags, "SOC2-CC6")
			}
			if event.Category == CategoryDataAccess {
				tags = append(tags, "SOC2-CC9")
			}
		case "PCI-DSS":
			if event.Category == CategoryAuthentication {
				tags = append(tags, "PCI-DSS-8")
			}
			if event.Category == CategoryDataAccess {
				tags = append(tags, "PCI-DSS-10")
			}
		case "GDPR":
			if event.Category == CategoryDataAccess {
				tags = append(tags, "GDPR-Art32")
			}
		}
	}

	return tags
}

func (als *AuditLoggingSystem) calculateRiskScore(event AuditEvent) float64 {
	score := 0.0

	// Base score by severity
	switch event.Severity {
	case SeverityCritical:
		score += 10.0
	case SeverityHigh:
		score += 7.0
	case SeverityMedium:
		score += 4.0
	case SeverityLow:
		score += 2.0
	case SeverityInfo:
		score += 0.5
	}

	// Additional score by event type
	switch event.EventType {
	case EventTypeLoginFailed, EventTypeAccountLocked:
		score += 3.0
	case EventTypeAccessDenied:
		score += 2.0
	case EventTypeDataDelete:
		score += 5.0
	case EventTypeVulnerabilityDetected:
		score += 8.0
	}

	// Cap at 10.0
	if score > 10.0 {
		score = 10.0
	}

	return score
}

// NewAuditFileLogger creates a new file-based audit logger
func NewAuditFileLogger(logger *observability.Logger, config AuditLoggingConfig) *AuditFileLogger {
	// Validate log directory
	validDir, err := validators.ValidateDirectoryPath(config.LogDirectory, "")
	if err != nil {
		logger.Error("Invalid log directory", "error", err)
		return nil
	}

	// Create log directory if it doesn't exist with secure permissions
	if err := os.MkdirAll(validDir, 0750); err != nil {
		logger.Error("Failed to create audit log directory", "error", err)
		return nil
	}

	// Generate initial file path
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	fileName := fmt.Sprintf("audit-%s.log", timestamp)
	filePath := filepath.Join(validDir, fileName)

	// Validate file path
	validFilePath, err := validators.ValidateFilePath(filePath, validDir)
	if err != nil {
		logger.Error("Invalid file path", "error", err)
		return nil
	}

	// Open log file with secure permissions (owner read/write only)
	file, err := os.OpenFile(validFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		logger.Error("Failed to open audit log file", "error", err)
		return nil
	}

	// Get file info for size tracking
	fileInfo, err := file.Stat()
	if err != nil {
		logger.Error("Failed to get file info", "error", err)
		file.Close()
		return nil
	}

	return &AuditFileLogger{
		logger:    logger,
		config:    config,
		file:      file,
		filePath:  validFilePath,
		encoder:   json.NewEncoder(file),
		fileSize:  fileInfo.Size(),
		fileCount: 1,
	}
}

// LogEvent logs an event to file
func (afl *AuditFileLogger) LogEvent(event AuditEvent) error {
	afl.mutex.Lock()
	defer afl.mutex.Unlock()

	// Check if we need to rotate the file
	if afl.shouldRotateFile() {
		if err := afl.rotateFile(); err != nil {
			return err
		}
	}

	// Encode and write event
	if err := afl.encoder.Encode(event); err != nil {
		return err
	}

	// Update file size
	afl.fileSize += int64(len(fmt.Sprintf("%+v", event)))

	return nil
}

// shouldRotateFile checks if the file should be rotated
func (afl *AuditFileLogger) shouldRotateFile() bool {
	return afl.fileSize > afl.config.MaxFileSize*1024*1024 // Convert MB to bytes
}

// rotateFile rotates the current log file
func (afl *AuditFileLogger) rotateFile() error {
	// Close current file
	if err := afl.file.Close(); err != nil {
		return err
	}

	// Validate log directory
	validDir, err := validators.ValidateDirectoryPath(afl.config.LogDirectory, "")
	if err != nil {
		return fmt.Errorf("invalid log directory: %w", err)
	}

	// Generate new file path
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	fileName := fmt.Sprintf("audit-%s.log", timestamp)
	filePath := filepath.Join(validDir, fileName)

	// Validate file path
	validFilePath, err := validators.ValidateFilePath(filePath, validDir)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Open new file with secure permissions (owner read/write only)
	file, err := os.OpenFile(validFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	afl.file = file
	afl.filePath = validFilePath
	afl.encoder = json.NewEncoder(file)
	afl.fileSize = 0
	afl.fileCount++

	// Clean up old files if we exceed max files
	if afl.fileCount > afl.config.MaxFiles {
		afl.cleanupOldFiles()
	}

	return nil
}

// cleanupOldFiles removes old log files
func (afl *AuditFileLogger) cleanupOldFiles() {
	// In a real implementation, this would remove old files
	// For now, we'll just log that cleanup is needed
	afl.logger.Info("Audit log file cleanup needed",
		"file_count", afl.fileCount,
		"max_files", afl.config.MaxFiles,
	)
}

// cleanup removes old audit files
func (afl *AuditFileLogger) cleanup() {
	// In a real implementation, this would remove files older than retention days
	// For now, we'll just log that cleanup is needed
	afl.logger.Info("Audit log cleanup routine executed")
}

// NewAuditDatabaseLogger creates a new database-based audit logger
func NewAuditDatabaseLogger(logger *observability.Logger, config AuditLoggingConfig) *AuditDatabaseLogger {
	return &AuditDatabaseLogger{
		logger: logger,
		config: config,
	}
}

// LogEvent logs an event to database
func (adl *AuditDatabaseLogger) LogEvent(event AuditEvent) error {
	// In a real implementation, this would insert the event into a database
	// For now, we'll just log that the event would be stored
	adl.logger.Debug("Audit event would be stored in database",
		"event_id", event.ID,
		"event_type", event.EventType,
	)
	return nil
}

// cleanup removes old audit records from database
func (adl *AuditDatabaseLogger) cleanup() {
	// In a real implementation, this would remove records older than retention days
	// For now, we'll just log that cleanup is needed
	adl.logger.Info("Database audit cleanup routine executed")
}
