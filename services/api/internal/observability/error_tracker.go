package observability

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ErrorTracker handles error tracking and analysis
type ErrorTracker struct {
	logger      *Logger
	errors      map[string]*ErrorEvent
	mu          sync.RWMutex
	exporters   []ErrorExporter
	alertConfig *ErrorAlertConfig
}

// ErrorEvent represents a tracked error
type ErrorEvent struct {
	ID          string
	Message     string
	Stack       string
	Severity    ErrorSeverity
	Context     map[string]interface{}
	Timestamp   time.Time
	Count       int64
	FirstSeen   time.Time
	LastSeen    time.Time
	Resolved    bool
	Tags        map[string]string
	UserID      string
	SessionID   string
	RequestID   string
	Environment string
	Service     string
	Version     string
}

// ErrorSeverity represents the severity of an error
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "low"
	ErrorSeverityMedium   ErrorSeverity = "medium"
	ErrorSeverityHigh     ErrorSeverity = "high"
	ErrorSeverityCritical ErrorSeverity = "critical"
)

// ErrorExporter interface for exporting errors
type ErrorExporter interface {
	Export(error *ErrorEvent) error
	Name() string
}

// ErrorAlertConfig holds configuration for error alerting
type ErrorAlertConfig struct {
	Enabled           bool
	CriticalThreshold int64
	HighThreshold     int64
	MediumThreshold   int64
	TimeWindow        time.Duration
	AlertChannels     []string
}

// SentryExporter exports errors to Sentry
type SentryExporter struct {
	logger *Logger
	config map[string]interface{}
}

// NewSentryExporter creates a new Sentry exporter
func NewSentryExporter(logger *Logger, config map[string]interface{}) *SentryExporter {
	return &SentryExporter{
		logger: logger,
		config: config,
	}
}

// Export exports an error to Sentry
func (se *SentryExporter) Export(error *ErrorEvent) error {
	// In a real implementation, this would export to Sentry
	se.logger.Debug("Exporting error to Sentry", map[string]interface{}{
		"error_id": error.ID,
		"severity": error.Severity,
		"message":  error.Message,
		"count":    error.Count,
	})
	return nil
}

// Name returns the exporter name
func (se *SentryExporter) Name() string {
	return "sentry"
}

// LogExporter exports errors to logs
type LogExporter struct {
	logger *Logger
}

// NewLogExporter creates a new log exporter
func NewLogExporter(logger *Logger) *LogExporter {
	return &LogExporter{
		logger: logger,
	}
}

// Export exports an error to logs
func (le *LogExporter) Export(error *ErrorEvent) error {
	le.logger.Error("Error tracked", map[string]interface{}{
		"error_id":    error.ID,
		"message":     error.Message,
		"severity":    error.Severity,
		"count":       error.Count,
		"first_seen":  error.FirstSeen,
		"last_seen":   error.LastSeen,
		"resolved":    error.Resolved,
		"tags":        error.Tags,
		"user_id":     error.UserID,
		"session_id":  error.SessionID,
		"request_id":  error.RequestID,
		"environment": error.Environment,
		"service":     error.Service,
		"version":     error.Version,
		"context":     error.Context,
		"stack":       error.Stack,
	})
	return nil
}

// Name returns the exporter name
func (le *LogExporter) Name() string {
	return "log"
}

// NewErrorTracker creates a new error tracker
func NewErrorTracker(logger *Logger, alertConfig *ErrorAlertConfig) *ErrorTracker {
	return &ErrorTracker{
		logger:      logger,
		errors:      make(map[string]*ErrorEvent),
		exporters:   make([]ErrorExporter, 0),
		alertConfig: alertConfig,
	}
}

// TrackError tracks an error
func (et *ErrorTracker) TrackError(err error, severity ErrorSeverity, context map[string]interface{}, tags map[string]string) {
	if err == nil {
		return
	}

	et.mu.Lock()
	defer et.mu.Unlock()

	errorID := et.generateErrorID(err)
	now := time.Now()

	// Check if this error already exists
	if existing, exists := et.errors[errorID]; exists {
		// Update existing error
		existing.Count++
		existing.LastSeen = now
		existing.Context = et.mergeContext(existing.Context, context)
		existing.Tags = et.mergeTags(existing.Tags, tags)

		et.logger.Debug("Error count incremented", map[string]interface{}{
			"error_id": errorID,
			"count":    existing.Count,
		})
	} else {
		// Create new error event
		errorEvent := &ErrorEvent{
			ID:          errorID,
			Message:     err.Error(),
			Stack:       et.getStackTrace(),
			Severity:    severity,
			Context:     context,
			Timestamp:   now,
			Count:       1,
			FirstSeen:   now,
			LastSeen:    now,
			Resolved:    false,
			Tags:        tags,
			Environment: "development", // Would come from config
			Service:     "kyb-platform",
			Version:     "1.0.0",
		}

		et.errors[errorID] = errorEvent

		et.logger.Info("New error tracked", map[string]interface{}{
			"error_id":   errorID,
			"message":    errorEvent.Message,
			"severity":   severity,
			"first_seen": now,
		})
	}

	// Export the error
	et.exportError(et.errors[errorID])
}

// TrackErrorWithContext tracks an error with additional context
func (et *ErrorTracker) TrackErrorWithContext(err error, severity ErrorSeverity, userID, sessionID, requestID string, context map[string]interface{}, tags map[string]string) {
	if err == nil {
		return
	}

	et.mu.Lock()
	defer et.mu.Unlock()

	errorID := et.generateErrorID(err)
	now := time.Now()

	// Check if this error already exists
	if existing, exists := et.errors[errorID]; exists {
		// Update existing error
		existing.Count++
		existing.LastSeen = now
		existing.Context = et.mergeContext(existing.Context, context)
		existing.Tags = et.mergeTags(existing.Tags, tags)

		// Update user context if provided
		if userID != "" {
			existing.UserID = userID
		}
		if sessionID != "" {
			existing.SessionID = sessionID
		}
		if requestID != "" {
			existing.RequestID = requestID
		}
	} else {
		// Create new error event
		errorEvent := &ErrorEvent{
			ID:          errorID,
			Message:     err.Error(),
			Stack:       et.getStackTrace(),
			Severity:    severity,
			Context:     context,
			Timestamp:   now,
			Count:       1,
			FirstSeen:   now,
			LastSeen:    now,
			Resolved:    false,
			Tags:        tags,
			UserID:      userID,
			SessionID:   sessionID,
			RequestID:   requestID,
			Environment: "development",
			Service:     "kyb-platform",
			Version:     "1.0.0",
		}

		et.errors[errorID] = errorEvent
	}

	// Export the error
	et.exportError(et.errors[errorID])
}

// GetError returns a specific error by ID
func (et *ErrorTracker) GetError(errorID string) (*ErrorEvent, bool) {
	et.mu.RLock()
	defer et.mu.RUnlock()

	error, exists := et.errors[errorID]
	if !exists {
		return nil, false
	}

	// Return a copy
	return &ErrorEvent{
		ID:          error.ID,
		Message:     error.Message,
		Stack:       error.Stack,
		Severity:    error.Severity,
		Context:     error.Context,
		Timestamp:   error.Timestamp,
		Count:       error.Count,
		FirstSeen:   error.FirstSeen,
		LastSeen:    error.LastSeen,
		Resolved:    error.Resolved,
		Tags:        error.Tags,
		UserID:      error.UserID,
		SessionID:   error.SessionID,
		RequestID:   error.RequestID,
		Environment: error.Environment,
		Service:     error.Service,
		Version:     error.Version,
	}, true
}

// GetErrorsBySeverity returns errors filtered by severity
func (et *ErrorTracker) GetErrorsBySeverity(severity ErrorSeverity) []*ErrorEvent {
	et.mu.RLock()
	defer et.mu.RUnlock()

	var filtered []*ErrorEvent
	for _, error := range et.errors {
		if error.Severity == severity {
			filtered = append(filtered, &ErrorEvent{
				ID:          error.ID,
				Message:     error.Message,
				Stack:       error.Stack,
				Severity:    error.Severity,
				Context:     error.Context,
				Timestamp:   error.Timestamp,
				Count:       error.Count,
				FirstSeen:   error.FirstSeen,
				LastSeen:    error.LastSeen,
				Resolved:    error.Resolved,
				Tags:        error.Tags,
				UserID:      error.UserID,
				SessionID:   error.SessionID,
				RequestID:   error.RequestID,
				Environment: error.Environment,
				Service:     error.Service,
				Version:     error.Version,
			})
		}
	}
	return filtered
}

// GetUnresolvedErrors returns all unresolved errors
func (et *ErrorTracker) GetUnresolvedErrors() []*ErrorEvent {
	et.mu.RLock()
	defer et.mu.RUnlock()

	var unresolved []*ErrorEvent
	for _, error := range et.errors {
		if !error.Resolved {
			unresolved = append(unresolved, &ErrorEvent{
				ID:          error.ID,
				Message:     error.Message,
				Stack:       error.Stack,
				Severity:    error.Severity,
				Context:     error.Context,
				Timestamp:   error.Timestamp,
				Count:       error.Count,
				FirstSeen:   error.FirstSeen,
				LastSeen:    error.LastSeen,
				Resolved:    error.Resolved,
				Tags:        error.Tags,
				UserID:      error.UserID,
				SessionID:   error.SessionID,
				RequestID:   error.RequestID,
				Environment: error.Environment,
				Service:     error.Service,
				Version:     error.Version,
			})
		}
	}
	return unresolved
}

// ResolveError marks an error as resolved
func (et *ErrorTracker) ResolveError(errorID string) error {
	et.mu.Lock()
	defer et.mu.Unlock()

	error, exists := et.errors[errorID]
	if !exists {
		return fmt.Errorf("error with ID %s not found", errorID)
	}

	error.Resolved = true
	et.logger.Info("Error resolved", map[string]interface{}{
		"error_id": errorID,
		"message":  error.Message,
		"count":    error.Count,
	})
	return nil
}

// GetSummary returns error summary statistics
func (et *ErrorTracker) GetSummary() map[string]interface{} {
	et.mu.RLock()
	defer et.mu.RUnlock()

	summary := map[string]interface{}{
		"total_errors":      len(et.errors),
		"unresolved":        0,
		"by_severity":       make(map[ErrorSeverity]int),
		"total_occurrences": 0,
		"recent_errors":     make([]*ErrorEvent, 0),
	}

	now := time.Now()
	recentThreshold := now.Add(-24 * time.Hour)

	for _, error := range et.errors {
		summary["by_severity"].(map[ErrorSeverity]int)[error.Severity]++
		summary["total_occurrences"] = summary["total_occurrences"].(int) + int(error.Count)

		if !error.Resolved {
			summary["unresolved"] = summary["unresolved"].(int) + 1
		}

		// Add recent errors (last 24 hours)
		if error.LastSeen.After(recentThreshold) {
			summary["recent_errors"] = append(summary["recent_errors"].([]*ErrorEvent), &ErrorEvent{
				ID:          error.ID,
				Message:     error.Message,
				Stack:       error.Stack,
				Severity:    error.Severity,
				Context:     error.Context,
				Timestamp:   error.Timestamp,
				Count:       error.Count,
				FirstSeen:   error.FirstSeen,
				LastSeen:    error.LastSeen,
				Resolved:    error.Resolved,
				Tags:        error.Tags,
				UserID:      error.UserID,
				SessionID:   error.SessionID,
				RequestID:   error.RequestID,
				Environment: error.Environment,
				Service:     error.Service,
				Version:     error.Version,
			})
		}
	}

	return summary
}

// AddExporter adds an error exporter
func (et *ErrorTracker) AddExporter(exporter ErrorExporter) {
	et.mu.Lock()
	defer et.mu.Unlock()

	et.exporters = append(et.exporters, exporter)
	et.logger.Info("Error exporter added", map[string]interface{}{
		"exporter": exporter.Name(),
	})
}

// ProcessAlerts processes error alerts based on thresholds
func (et *ErrorTracker) ProcessAlerts() {
	if !et.alertConfig.Enabled {
		return
	}

	et.mu.RLock()
	errors := make([]*ErrorEvent, 0, len(et.errors))
	for _, error := range et.errors {
		errors = append(errors, error)
	}
	et.mu.RUnlock()

	now := time.Now()
	timeWindow := now.Add(-et.alertConfig.TimeWindow)

	for _, error := range errors {
		if error.Resolved {
			continue
		}

		// Check if error occurred within the time window
		if error.LastSeen.Before(timeWindow) {
			continue
		}

		// Check thresholds
		var shouldAlert bool
		var alertLevel string

		switch error.Severity {
		case ErrorSeverityCritical:
			shouldAlert = error.Count >= et.alertConfig.CriticalThreshold
			alertLevel = "critical"
		case ErrorSeverityHigh:
			shouldAlert = error.Count >= et.alertConfig.HighThreshold
			alertLevel = "high"
		case ErrorSeverityMedium:
			shouldAlert = error.Count >= et.alertConfig.MediumThreshold
			alertLevel = "medium"
		}

		if shouldAlert {
			et.sendAlert(error, alertLevel)
		}
	}
}

// sendAlert sends an alert for an error
func (et *ErrorTracker) sendAlert(error *ErrorEvent, alertLevel string) {
	et.logger.Warn("Error alert triggered", map[string]interface{}{
		"error_id":    error.ID,
		"message":     error.Message,
		"severity":    error.Severity,
		"count":       error.Count,
		"alert_level": alertLevel,
		"channels":    et.alertConfig.AlertChannels,
	})

	// In a real implementation, this would send alerts via configured channels
}

// exportError exports an error using registered exporters
func (et *ErrorTracker) exportError(error *ErrorEvent) {
	for _, exporter := range et.exporters {
		if err := exporter.Export(error); err != nil {
			et.logger.Error("Failed to export error", map[string]interface{}{
				"exporter": exporter.Name(),
				"error_id": error.ID,
				"error":    err.Error(),
			})
		}
	}
}

// generateErrorID generates a unique ID for an error
func (et *ErrorTracker) generateErrorID(err error) string {
	// Simple hash of error message for grouping similar errors
	message := err.Error()
	// In a real implementation, this would use a proper hash function
	return fmt.Sprintf("error_%d", len(message))
}

// getStackTrace gets the current stack trace
func (et *ErrorTracker) getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// mergeContext merges two context maps
func (et *ErrorTracker) mergeContext(existing, new map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy existing context
	for k, v := range existing {
		merged[k] = v
	}

	// Add new context
	for k, v := range new {
		merged[k] = v
	}

	return merged
}

// mergeTags merges two tag maps
func (et *ErrorTracker) mergeTags(existing, new map[string]string) map[string]string {
	merged := make(map[string]string)

	// Copy existing tags
	for k, v := range existing {
		merged[k] = v
	}

	// Add new tags
	for k, v := range new {
		merged[k] = v
	}

	return merged
}

// ClearResolvedErrors removes resolved errors older than the specified duration
func (et *ErrorTracker) ClearResolvedErrors(olderThan time.Duration) {
	et.mu.Lock()
	defer et.mu.Unlock()

	now := time.Now()
	threshold := now.Add(-olderThan)
	count := 0

	for id, error := range et.errors {
		if error.Resolved && error.LastSeen.Before(threshold) {
			delete(et.errors, id)
			count++
		}
	}

	if count > 0 {
		et.logger.Info("Cleared resolved errors", map[string]interface{}{
			"count":      count,
			"older_than": olderThan.String(),
		})
	}
}
