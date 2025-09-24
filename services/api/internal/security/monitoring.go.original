package security

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// SecurityEvent extends BaseEvent with security-specific fields
type SecurityEvent struct {
	BaseEvent
	Source     string     `json:"source"`
	Resolved   bool       `json:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// SecurityAlert represents a security alert that requires attention
type SecurityAlert struct {
	ID          string            `json:"id"`
	EventID     string            `json:"event_id"`
	Timestamp   time.Time         `json:"timestamp"`
	Severity    Severity          `json:"severity"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Status      AlertStatus       `json:"status"`
	AssignedTo  string            `json:"assigned_to,omitempty"`
	Notes       []string          `json:"notes,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SecurityMetrics represents security-related metrics
type SecurityMetrics struct {
	TotalEvents      int64                 `json:"total_events"`
	EventsByType     map[EventType]int64   `json:"events_by_type"`
	EventsBySeverity map[Severity]int64    `json:"events_by_severity"`
	OpenAlerts       int64                 `json:"open_alerts"`
	AlertsByStatus   map[AlertStatus]int64 `json:"alerts_by_status"`
	MTTD             time.Duration         `json:"mean_time_to_detect"`
	MTTR             time.Duration         `json:"mean_time_to_resolve"`
	LastUpdated      time.Time             `json:"last_updated"`
}

// SecurityMonitor provides security monitoring capabilities
type SecurityMonitor struct {
	logger        *observability.Logger
	events        []SecurityEvent
	alerts        []SecurityAlert
	eventHandlers map[EventType][]func(SecurityEvent)
	alertHandlers map[Severity][]func(SecurityAlert)
	mutex         sync.RWMutex
	config        SecurityMonitorConfig
}

// SecurityMonitorConfig defines configuration for security monitoring
type SecurityMonitorConfig struct {
	AlertThresholds      map[Severity]int `json:"alert_thresholds"`
	RetentionDays        int              `json:"retention_days"`
	AutoResolution       bool             `json:"auto_resolution"`
	NotificationChannels []string         `json:"notification_channels"`
}

// NewSecurityMonitor creates a new security monitor instance
func NewSecurityMonitor(logger *observability.Logger, config SecurityMonitorConfig) *SecurityMonitor {
	monitor := &SecurityMonitor{
		logger:        logger,
		events:        make([]SecurityEvent, 0),
		alerts:        make([]SecurityAlert, 0),
		eventHandlers: make(map[EventType][]func(SecurityEvent)),
		alertHandlers: make(map[Severity][]func(SecurityAlert)),
		config:        config,
	}

	// Start background cleanup
	go monitor.cleanupRoutine()

	return monitor
}

// RecordEvent records a security event
func (sm *SecurityMonitor) RecordEvent(ctx context.Context, event SecurityEvent) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = fmt.Sprintf("sec_%d", time.Now().UnixNano())
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Add event to the list
	sm.events = append(sm.events, event)

	// Check if we should create an alert
	if sm.shouldCreateAlert(event) {
		alert := sm.createAlertFromEvent(event)
		sm.alerts = append(sm.alerts, alert)

		// Call alert handlers
		if handlers, exists := sm.alertHandlers[event.Severity]; exists {
			for _, handler := range handlers {
				go handler(alert)
			}
		}

		sm.logger.Info("Security alert created",
			"alert_id", alert.ID,
			"event_id", event.ID,
			"severity", alert.Severity,
			"title", alert.Title,
		)
	}

	// Call event handlers
	if handlers, exists := sm.eventHandlers[event.EventType]; exists {
		for _, handler := range handlers {
			go handler(event)
		}
	}

	sm.logger.Info("Security event recorded",
		"event_id", event.ID,
		"type", event.EventType,
		"severity", event.Severity,
		"source", event.Source,
		"user_id", event.UserID,
	)

	return nil
}

// shouldCreateAlert determines if an alert should be created for an event
func (sm *SecurityMonitor) shouldCreateAlert(event SecurityEvent) bool {
	// Check if severity meets threshold
	if threshold, exists := sm.config.AlertThresholds[event.Severity]; exists {
		// Count recent events of this type and severity
		recentCount := sm.countRecentEvents(event.EventType, event.Severity, 1*time.Hour)
		return recentCount >= threshold
	}

	// Default: create alerts for high and critical events
	return event.Severity == SeverityHigh || event.Severity == SeverityCritical
}

// createAlertFromEvent creates a security alert from a security event
func (sm *SecurityMonitor) createAlertFromEvent(event SecurityEvent) SecurityAlert {
	alert := SecurityAlert{
		ID:          fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		EventID:     event.ID,
		Timestamp:   time.Now(),
		Severity:    event.Severity,
		Title:       fmt.Sprintf("Security Event: %s", event.EventType),
		Description: event.Description,
		Category:    string(event.Category),
		Status:      AlertStatusOpen,
		Metadata:    make(map[string]string),
	}

	// Copy relevant metadata
	if event.Metadata != nil {
		for k, v := range event.Metadata {
			if str, ok := v.(string); ok {
				alert.Metadata[k] = str
			}
		}
	}

	return alert
}

// countRecentEvents counts events of a specific type and severity within a time window
func (sm *SecurityMonitor) countRecentEvents(eventType EventType, severity Severity, window time.Duration) int {
	count := 0
	cutoff := time.Now().Add(-window)

	for _, event := range sm.events {
		if event.EventType == eventType && event.Severity == severity && event.Timestamp.After(cutoff) {
			count++
		}
	}

	return count
}

// GetEvents retrieves security events with optional filtering
func (sm *SecurityMonitor) GetEvents(ctx context.Context, filters map[string]interface{}) ([]SecurityEvent, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var filteredEvents []SecurityEvent
	for _, event := range sm.events {
		if sm.matchesFilters(event, filters) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

// GetAlerts retrieves security alerts with optional filtering
func (sm *SecurityMonitor) GetAlerts(ctx context.Context, filters map[string]interface{}) ([]SecurityAlert, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var filteredAlerts []SecurityAlert
	for _, alert := range sm.alerts {
		if sm.matchesAlertFilters(alert, filters) {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	return filteredAlerts, nil
}

// UpdateAlertStatus updates the status of a security alert
func (sm *SecurityMonitor) UpdateAlertStatus(ctx context.Context, alertID string, status AlertStatus, notes string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for i, alert := range sm.alerts {
		if alert.ID == alertID {
			sm.alerts[i].Status = status
			if notes != "" {
				sm.alerts[i].Notes = append(sm.alerts[i].Notes, notes)
			}

			sm.logger.Info("Alert status updated",
				"alert_id", alertID,
				"status", status,
				"notes", notes,
			)

			return nil
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// GetMetrics retrieves security monitoring metrics
func (sm *SecurityMonitor) GetMetrics(ctx context.Context) (*SecurityMetrics, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	metrics := &SecurityMetrics{
		TotalEvents:      int64(len(sm.events)),
		EventsByType:     make(map[EventType]int64),
		EventsBySeverity: make(map[Severity]int64),
		OpenAlerts:       0,
		AlertsByStatus:   make(map[AlertStatus]int64),
		LastUpdated:      time.Now(),
	}

	// Count events by type and severity
	for _, event := range sm.events {
		metrics.EventsByType[event.EventType]++
		metrics.EventsBySeverity[event.Severity]++
	}

	// Count alerts by status
	for _, alert := range sm.alerts {
		metrics.AlertsByStatus[alert.Status]++
		if alert.Status == AlertStatusOpen {
			metrics.OpenAlerts++
		}
	}

	// Calculate MTTD and MTTR
	metrics.MTTD = sm.calculateMTTD()
	metrics.MTTR = sm.calculateMTTR()

	return metrics, nil
}

// calculateMTTD calculates the mean time to detect
func (sm *SecurityMonitor) calculateMTTD() time.Duration {
	if len(sm.events) == 0 {
		return 0
	}

	var totalDetectionTime time.Duration
	detectionCount := 0

	for _, event := range sm.events {
		// For now, we'll use a simple calculation
		// In a real implementation, this would track when events were first detected
		if event.Severity == SeverityHigh || event.Severity == SeverityCritical {
			detectionCount++
		}
	}

	if detectionCount == 0 {
		return 0
	}

	// Simplified calculation - in reality, this would track actual detection times
	return totalDetectionTime / time.Duration(detectionCount)
}

// calculateMTTR calculates the mean time to resolve
func (sm *SecurityMonitor) calculateMTTR() time.Duration {
	if len(sm.alerts) == 0 {
		return 0
	}

	var totalResolutionTime time.Duration
	resolutionCount := 0

	for _, alert := range sm.alerts {
		if alert.Status == AlertStatusResolved || alert.Status == AlertStatusClosed {
			resolutionCount++
			// In a real implementation, this would calculate actual resolution time
		}
	}

	if resolutionCount == 0 {
		return 0
	}

	// Simplified calculation - in reality, this would track actual resolution times
	return totalResolutionTime / time.Duration(resolutionCount)
}

// RegisterEventHandler registers a handler for a specific event type
func (sm *SecurityMonitor) RegisterEventHandler(eventType EventType, handler func(SecurityEvent)) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.eventHandlers[eventType] = append(sm.eventHandlers[eventType], handler)
}

// RegisterAlertHandler registers a handler for a specific alert severity
func (sm *SecurityMonitor) RegisterAlertHandler(severity Severity, handler func(SecurityAlert)) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.alertHandlers[severity] = append(sm.alertHandlers[severity], handler)
}

// matchesFilters checks if an event matches the given filters
func (sm *SecurityMonitor) matchesFilters(event SecurityEvent, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "event_type":
			if eventType, ok := value.(EventType); ok && event.EventType != eventType {
				return false
			}
		case "severity":
			if severity, ok := value.(Severity); ok && event.Severity != severity {
				return false
			}
		case "user_id":
			if userID, ok := value.(string); ok && event.UserID != userID {
				return false
			}
		case "source":
			if source, ok := value.(string); ok && event.Source != source {
				return false
			}
		case "resolved":
			if resolved, ok := value.(bool); ok && event.Resolved != resolved {
				return false
			}
		}
	}
	return true
}

// matchesAlertFilters checks if an alert matches the given filters
func (sm *SecurityMonitor) matchesAlertFilters(alert SecurityAlert, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "status":
			if status, ok := value.(AlertStatus); ok && alert.Status != status {
				return false
			}
		case "severity":
			if severity, ok := value.(Severity); ok && alert.Severity != severity {
				return false
			}
		case "assigned_to":
			if assignedTo, ok := value.(string); ok && alert.AssignedTo != assignedTo {
				return false
			}
		}
	}
	return true
}

// cleanupRoutine runs background cleanup tasks
func (sm *SecurityMonitor) cleanupRoutine() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanup()
	}
}

// cleanup removes old events and alerts based on retention policy
func (sm *SecurityMonitor) cleanup() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.config.RetentionDays <= 0 {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -sm.config.RetentionDays)

	// Clean up old events
	var newEvents []SecurityEvent
	for _, event := range sm.events {
		if event.Timestamp.After(cutoff) {
			newEvents = append(newEvents, event)
		}
	}
	sm.events = newEvents

	// Clean up old alerts
	var newAlerts []SecurityAlert
	for _, alert := range sm.alerts {
		if alert.Timestamp.After(cutoff) {
			newAlerts = append(newAlerts, alert)
		}
	}
	sm.alerts = newAlerts

	log.Printf("Security monitor cleanup completed: %d events, %d alerts retained", len(sm.events), len(sm.alerts))
}

// ExportEvents exports security events to JSON
func (sm *SecurityMonitor) ExportEvents(ctx context.Context, filters map[string]interface{}) ([]byte, error) {
	events, err := sm.GetEvents(ctx, filters)
	if err != nil {
		return nil, err
	}

	return json.Marshal(events)
}

// ExportAlerts exports security alerts to JSON
func (sm *SecurityMonitor) ExportAlerts(ctx context.Context, filters map[string]interface{}) ([]byte, error) {
	alerts, err := sm.GetAlerts(ctx, filters)
	if err != nil {
		return nil, err
	}

	return json.Marshal(alerts)
}
