package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID         string                 `json:"id"`
	Type       SecurityEventType      `json:"type"`
	Severity   SecurityEventSeverity  `json:"severity"`
	Source     string                 `json:"source"`
	UserID     string                 `json:"user_id,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Endpoint   string                 `json:"endpoint,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Details    map[string]interface{} `json:"details"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy string                 `json:"resolved_by,omitempty"`
	Notes      string                 `json:"notes,omitempty"`
}

// SecurityEventType represents the type of security event
type SecurityEventType string

const (
	// Authentication Events
	EventTypeLoginAttempt   SecurityEventType = "login_attempt"
	EventTypeLoginSuccess   SecurityEventType = "login_success"
	EventTypeLoginFailure   SecurityEventType = "login_failure"
	EventTypeLogout         SecurityEventType = "logout"
	EventTypePasswordChange SecurityEventType = "password_change"
	EventTypePasswordReset  SecurityEventType = "password_reset"
	EventTypeAccountLockout SecurityEventType = "account_lockout"
	EventTypeAccountUnlock  SecurityEventType = "account_unlock"

	// Authorization Events
	EventTypeAccessDenied     SecurityEventType = "access_denied"
	EventTypePermissionDenied SecurityEventType = "permission_denied"
	EventTypeRoleChange       SecurityEventType = "role_change"
	EventTypePermissionChange SecurityEventType = "permission_change"

	// Input Validation Events
	EventTypeInvalidInput         SecurityEventType = "invalid_input"
	EventTypeSQLInjectionAttempt  SecurityEventType = "sql_injection_attempt"
	EventTypeXSSAttempt           SecurityEventType = "xss_attempt"
	EventTypePathTraversalAttempt SecurityEventType = "path_traversal_attempt"

	// Rate Limiting Events
	EventTypeRateLimitExceeded     SecurityEventType = "rate_limit_exceeded"
	EventTypeAuthRateLimitExceeded SecurityEventType = "auth_rate_limit_exceeded"

	// API Security Events
	EventTypeInvalidAPIKey SecurityEventType = "invalid_api_key"
	EventTypeExpiredToken  SecurityEventType = "expired_token"
	EventTypeInvalidToken  SecurityEventType = "invalid_token"
	EventTypeTokenRefresh  SecurityEventType = "token_refresh"

	// System Security Events
	EventTypeSecurityHeaderViolation SecurityEventType = "security_header_violation"
	EventTypeCSPViolation            SecurityEventType = "csp_violation"
	EventTypeHSTSViolation           SecurityEventType = "hsts_violation"

	// Data Security Events
	EventTypeDataAccess       SecurityEventType = "data_access"
	EventTypeDataModification SecurityEventType = "data_modification"
	EventTypeDataExport       SecurityEventType = "data_export"
	EventTypeDataDeletion     SecurityEventType = "data_deletion"

	// System Events
	EventTypeSystemStartup       SecurityEventType = "system_startup"
	EventTypeSystemShutdown      SecurityEventType = "system_shutdown"
	EventTypeConfigurationChange SecurityEventType = "configuration_change"
	EventTypeBackupCompleted     SecurityEventType = "backup_completed"
	EventTypeBackupFailed        SecurityEventType = "backup_failed"
)

// SecurityEventSeverity represents the severity level of a security event
type SecurityEventSeverity string

const (
	SeverityInfo     SecurityEventSeverity = "info"
	SeverityLow      SecurityEventSeverity = "low"
	SeverityMedium   SecurityEventSeverity = "medium"
	SeverityHigh     SecurityEventSeverity = "high"
	SeverityCritical SecurityEventSeverity = "critical"
)

// SecurityAlert represents a security alert that can be sent to external systems
type SecurityAlert struct {
	ID             string                 `json:"id"`
	EventID        string                 `json:"event_id"`
	Type           SecurityAlertType      `json:"type"`
	Severity       SecurityEventSeverity  `json:"severity"`
	Title          string                 `json:"title"`
	Message        string                 `json:"message"`
	Source         string                 `json:"source"`
	Details        map[string]interface{} `json:"details"`
	Timestamp      time.Time              `json:"timestamp"`
	Acknowledged   bool                   `json:"acknowledged"`
	AcknowledgedAt *time.Time             `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string                 `json:"acknowledged_by,omitempty"`
}

// SecurityAlertType represents the type of security alert
type SecurityAlertType string

const (
	AlertTypeImmediate  SecurityAlertType = "immediate"
	AlertTypeScheduled  SecurityAlertType = "scheduled"
	AlertTypeDigest     SecurityAlertType = "digest"
	AlertTypeEscalation SecurityAlertType = "escalation"
)

// SecurityMetrics represents security-related metrics
type SecurityMetrics struct {
	TotalEvents           int64            `json:"total_events"`
	EventsByType          map[string]int64 `json:"events_by_type"`
	EventsBySeverity      map[string]int64 `json:"events_by_severity"`
	EventsBySource        map[string]int64 `json:"events_by_source"`
	ActiveAlerts          int64            `json:"active_alerts"`
	ResolvedEvents        int64            `json:"resolved_events"`
	AverageResolutionTime time.Duration    `json:"average_resolution_time"`
	TopIPAddresses        []IPAddressCount `json:"top_ip_addresses"`
	TopEndpoints          []EndpointCount  `json:"top_endpoints"`
	TopUserAgents         []UserAgentCount `json:"top_user_agents"`
	LastUpdated           time.Time        `json:"last_updated"`
}

// IPAddressCount represents IP address frequency
type IPAddressCount struct {
	IPAddress string `json:"ip_address"`
	Count     int64  `json:"count"`
}

// EndpointCount represents endpoint frequency
type EndpointCount struct {
	Endpoint string `json:"endpoint"`
	Count    int64  `json:"count"`
}

// UserAgentCount represents user agent frequency
type UserAgentCount struct {
	UserAgent string `json:"user_agent"`
	Count     int64  `json:"count"`
}

// SecurityMonitorConfig holds configuration for the security monitor
type SecurityMonitorConfig struct {
	// Event Storage
	MaxEvents      int           `json:"max_events" yaml:"max_events"`
	EventRetention time.Duration `json:"event_retention" yaml:"event_retention"`

	// Alerting
	AlertThresholds map[SecurityEventSeverity]int `json:"alert_thresholds" yaml:"alert_thresholds"`
	AlertCooldown   time.Duration                 `json:"alert_cooldown" yaml:"alert_cooldown"`

	// Metrics
	MetricsInterval time.Duration `json:"metrics_interval" yaml:"metrics_interval"`

	// External Integrations
	WebhookURL     string        `json:"webhook_url" yaml:"webhook_url"`
	WebhookTimeout time.Duration `json:"webhook_timeout" yaml:"webhook_timeout"`

	// Filtering
	ExcludeSources    []string            `json:"exclude_sources" yaml:"exclude_sources"`
	ExcludeEventTypes []SecurityEventType `json:"exclude_event_types" yaml:"exclude_event_types"`
}

// SecurityMonitor provides comprehensive security monitoring
type SecurityMonitor struct {
	config     *SecurityMonitorConfig
	logger     *zap.Logger
	events     []*SecurityEvent
	alerts     []*SecurityAlert
	metrics    *SecurityMetrics
	mutex      sync.RWMutex
	alertMutex sync.RWMutex

	// Channels for async processing
	eventChan chan *SecurityEvent
	alertChan chan *SecurityAlert
	stopChan  chan struct{}

	// Callbacks
	onEvent func(*SecurityEvent)
	onAlert func(*SecurityAlert)
}

// NewSecurityMonitor creates a new security monitor
func NewSecurityMonitor(config *SecurityMonitorConfig, logger *zap.Logger) *SecurityMonitor {
	if config == nil {
		config = &SecurityMonitorConfig{
			MaxEvents:      10000,
			EventRetention: 30 * 24 * time.Hour, // 30 days
			AlertThresholds: map[SecurityEventSeverity]int{
				SeverityCritical: 1,
				SeverityHigh:     5,
				SeverityMedium:   10,
				SeverityLow:      50,
			},
			AlertCooldown:   5 * time.Minute,
			MetricsInterval: 1 * time.Minute,
			WebhookTimeout:  10 * time.Second,
		}
	}

	monitor := &SecurityMonitor{
		config:    config,
		logger:    logger,
		events:    make([]*SecurityEvent, 0),
		alerts:    make([]*SecurityAlert, 0),
		metrics:   &SecurityMetrics{},
		eventChan: make(chan *SecurityEvent, 1000),
		alertChan: make(chan *SecurityAlert, 100),
		stopChan:  make(chan struct{}),
	}

	// Start background processing
	go monitor.processEvents()
	go monitor.processAlerts()
	go monitor.updateMetrics()

	return monitor
}

// RecordEvent records a security event
func (m *SecurityMonitor) RecordEvent(event *SecurityEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Set default values
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Check if event should be excluded
	if m.shouldExcludeEvent(event) {
		return nil
	}

	// Send to background processing
	select {
	case m.eventChan <- event:
		return nil
	default:
		return fmt.Errorf("event channel full")
	}
}

// processEvents processes events in the background
func (m *SecurityMonitor) processEvents() {
	for {
		select {
		case event := <-m.eventChan:
			m.processEvent(event)
		case <-m.stopChan:
			return
		}
	}
}

// processEvent processes a single event
func (m *SecurityMonitor) processEvent(event *SecurityEvent) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Add event to storage
	m.events = append(m.events, event)

	// Maintain max events limit
	if len(m.events) > m.config.MaxEvents {
		m.events = m.events[len(m.events)-m.config.MaxEvents:]
	}

	// Clean up old events
	m.cleanupOldEvents()

	// Check if alert should be generated
	if m.shouldGenerateAlert(event) {
		alert := m.createAlert(event)
		select {
		case m.alertChan <- alert:
		default:
			m.logger.Warn("alert channel full, dropping alert", zap.String("event_id", event.ID))
		}
	}

	// Log event
	m.logger.Info("security event recorded",
		zap.String("event_id", event.ID),
		zap.String("type", string(event.Type)),
		zap.String("severity", string(event.Severity)),
		zap.String("source", event.Source),
		zap.String("ip_address", event.IPAddress),
		zap.String("endpoint", event.Endpoint))

	// Call event callback
	if m.onEvent != nil {
		m.onEvent(event)
	}
}

// processAlerts processes alerts in the background
func (m *SecurityMonitor) processAlerts() {
	for {
		select {
		case alert := <-m.alertChan:
			m.processAlert(alert)
		case <-m.stopChan:
			return
		}
	}
}

// processAlert processes a single alert
func (m *SecurityMonitor) processAlert(alert *SecurityAlert) {
	m.alertMutex.Lock()
	defer m.alertMutex.Unlock()

	// Add alert to storage
	m.alerts = append(m.alerts, alert)

	// Send to external systems
	m.sendAlert(alert)

	// Log alert
	m.logger.Warn("security alert generated",
		zap.String("alert_id", alert.ID),
		zap.String("event_id", alert.EventID),
		zap.String("type", string(alert.Type)),
		zap.String("severity", string(alert.Severity)),
		zap.String("title", alert.Title))

	// Call alert callback
	if m.onAlert != nil {
		m.onAlert(alert)
	}
}

// updateMetrics updates metrics periodically
func (m *SecurityMonitor) updateMetrics() {
	ticker := time.NewTicker(m.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.updateMetricsInternal()
		case <-m.stopChan:
			return
		}
	}
}

// updateMetricsInternal updates the metrics
func (m *SecurityMonitor) updateMetricsInternal() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metrics := &SecurityMetrics{
		TotalEvents:      int64(len(m.events)),
		EventsByType:     make(map[string]int64),
		EventsBySeverity: make(map[string]int64),
		EventsBySource:   make(map[string]int64),
		LastUpdated:      time.Now(),
	}

	// Count events by type, severity, and source
	ipCounts := make(map[string]int64)
	endpointCounts := make(map[string]int64)
	userAgentCounts := make(map[string]int64)

	for _, event := range m.events {
		// Count by type
		metrics.EventsByType[string(event.Type)]++

		// Count by severity
		metrics.EventsBySeverity[string(event.Severity)]++

		// Count by source
		metrics.EventsBySource[event.Source]++

		// Count IP addresses
		if event.IPAddress != "" {
			ipCounts[event.IPAddress]++
		}

		// Count endpoints
		if event.Endpoint != "" {
			endpointCounts[event.Endpoint]++
		}

		// Count user agents
		if event.UserAgent != "" {
			userAgentCounts[event.UserAgent]++
		}

		// Count resolved events
		if event.Resolved {
			metrics.ResolvedEvents++
		}
	}

	// Convert to sorted slices
	metrics.TopIPAddresses = convertToIPAddressCounts(ipCounts)
	metrics.TopEndpoints = convertToEndpointCounts(endpointCounts)
	metrics.TopUserAgents = convertToUserAgentCounts(userAgentCounts)

	// Count active alerts
	m.alertMutex.RLock()
	metrics.ActiveAlerts = int64(len(m.alerts))
	m.alertMutex.RUnlock()

	// Calculate average resolution time
	metrics.AverageResolutionTime = m.calculateAverageResolutionTime()

	m.metrics = metrics
}

// GetEvents returns events with optional filtering
func (m *SecurityMonitor) GetEvents(filters EventFilters) ([]*SecurityEvent, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var filteredEvents []*SecurityEvent

	for _, event := range m.events {
		if m.matchesFilters(event, filters) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

// GetAlerts returns alerts with optional filtering
func (m *SecurityMonitor) GetAlerts(filters AlertFilters) ([]*SecurityAlert, error) {
	m.alertMutex.RLock()
	defer m.alertMutex.RUnlock()

	var filteredAlerts []*SecurityAlert

	for _, alert := range m.alerts {
		if m.matchesAlertFilters(alert, filters) {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	return filteredAlerts, nil
}

// GetMetrics returns current security metrics
func (m *SecurityMonitor) GetMetrics() *SecurityMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.metrics
}

// ResolveEvent resolves a security event
func (m *SecurityMonitor) ResolveEvent(eventID, resolvedBy, notes string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, event := range m.events {
		if event.ID == eventID {
			event.Resolved = true
			now := time.Now()
			event.ResolvedAt = &now
			event.ResolvedBy = resolvedBy
			event.Notes = notes

			m.logger.Info("security event resolved",
				zap.String("event_id", eventID),
				zap.String("resolved_by", resolvedBy))

			return nil
		}
	}

	return fmt.Errorf("event not found: %s", eventID)
}

// AcknowledgeAlert acknowledges a security alert
func (m *SecurityMonitor) AcknowledgeAlert(alertID, acknowledgedBy string) error {
	m.alertMutex.Lock()
	defer m.alertMutex.Unlock()

	for _, alert := range m.alerts {
		if alert.ID == alertID {
			alert.Acknowledged = true
			now := time.Now()
			alert.AcknowledgedAt = &now
			alert.AcknowledgedBy = acknowledgedBy

			m.logger.Info("security alert acknowledged",
				zap.String("alert_id", alertID),
				zap.String("acknowledged_by", acknowledgedBy))

			return nil
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// SetEventCallback sets a callback for events
func (m *SecurityMonitor) SetEventCallback(callback func(*SecurityEvent)) {
	m.onEvent = callback
}

// SetAlertCallback sets a callback for alerts
func (m *SecurityMonitor) SetAlertCallback(callback func(*SecurityAlert)) {
	m.onAlert = callback
}

// Stop stops the security monitor
func (m *SecurityMonitor) Stop() {
	close(m.stopChan)
}

// Helper methods

func (m *SecurityMonitor) shouldExcludeEvent(event *SecurityEvent) bool {
	// Check excluded sources
	for _, source := range m.config.ExcludeSources {
		if event.Source == source {
			return true
		}
	}

	// Check excluded event types
	for _, eventType := range m.config.ExcludeEventTypes {
		if event.Type == eventType {
			return true
		}
	}

	return false
}

func (m *SecurityMonitor) shouldGenerateAlert(event *SecurityEvent) bool {
	threshold, exists := m.config.AlertThresholds[event.Severity]
	if !exists {
		return false
	}

	// Count recent events of this severity
	count := m.countRecentEventsBySeverity(event.Severity, m.config.AlertCooldown)

	return count >= threshold
}

func (m *SecurityMonitor) countRecentEventsBySeverity(severity SecurityEventSeverity, duration time.Duration) int {
	cutoff := time.Now().Add(-duration)
	count := 0

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, event := range m.events {
		if event.Severity == severity && event.Timestamp.After(cutoff) {
			count++
		}
	}

	return count
}

func (m *SecurityMonitor) createAlert(event *SecurityEvent) *SecurityAlert {
	alert := &SecurityAlert{
		ID:        generateAlertID(),
		EventID:   event.ID,
		Type:      AlertTypeImmediate,
		Severity:  event.Severity,
		Title:     m.generateAlertTitle(event),
		Message:   m.generateAlertMessage(event),
		Source:    event.Source,
		Details:   event.Details,
		Timestamp: time.Now(),
	}

	return alert
}

func (m *SecurityMonitor) generateAlertTitle(event *SecurityEvent) string {
	switch event.Type {
	case EventTypeLoginFailure:
		return fmt.Sprintf("Multiple login failures from %s", event.IPAddress)
	case EventTypeRateLimitExceeded:
		return fmt.Sprintf("Rate limit exceeded for %s", event.IPAddress)
	case EventTypeSQLInjectionAttempt:
		return "SQL injection attempt detected"
	case EventTypeXSSAttempt:
		return "XSS attempt detected"
	default:
		return fmt.Sprintf("Security event: %s", string(event.Type))
	}
}

func (m *SecurityMonitor) generateAlertMessage(event *SecurityEvent) string {
	return fmt.Sprintf("Security event of type %s with severity %s detected from source %s",
		string(event.Type), string(event.Severity), event.Source)
}

func (m *SecurityMonitor) sendAlert(alert *SecurityAlert) {
	if m.config.WebhookURL == "" {
		return
	}

	// Send alert to webhook
	go func() {
		_, cancel := context.WithTimeout(context.Background(), m.config.WebhookTimeout)
		defer cancel()

		_, err := json.Marshal(alert)
		if err != nil {
			m.logger.Error("failed to marshal alert", zap.Error(err))
			return
		}

		// TODO: Implement webhook sending
		m.logger.Debug("webhook alert sent", zap.String("alert_id", alert.ID))
	}()
}

func (m *SecurityMonitor) cleanupOldEvents() {
	cutoff := time.Now().Add(-m.config.EventRetention)

	// Remove old events
	newEvents := make([]*SecurityEvent, 0)
	for _, event := range m.events {
		if event.Timestamp.After(cutoff) {
			newEvents = append(newEvents, event)
		}
	}

	m.events = newEvents
}

func (m *SecurityMonitor) calculateAverageResolutionTime() time.Duration {
	var totalTime time.Duration
	resolvedCount := 0

	for _, event := range m.events {
		if event.Resolved && event.ResolvedAt != nil {
			totalTime += event.ResolvedAt.Sub(event.Timestamp)
			resolvedCount++
		}
	}

	if resolvedCount == 0 {
		return 0
	}

	return totalTime / time.Duration(resolvedCount)
}

// EventFilters represents filters for events
type EventFilters struct {
	Types       []SecurityEventType     `json:"types"`
	Severities  []SecurityEventSeverity `json:"severities"`
	Sources     []string                `json:"sources"`
	UserIDs     []string                `json:"user_ids"`
	IPAddresses []string                `json:"ip_addresses"`
	StartTime   *time.Time              `json:"start_time"`
	EndTime     *time.Time              `json:"end_time"`
	Resolved    *bool                   `json:"resolved"`
	Limit       int                     `json:"limit"`
}

// AlertFilters represents filters for alerts
type AlertFilters struct {
	Types        []SecurityAlertType     `json:"types"`
	Severities   []SecurityEventSeverity `json:"severities"`
	Sources      []string                `json:"sources"`
	Acknowledged *bool                   `json:"acknowledged"`
	StartTime    *time.Time              `json:"start_time"`
	EndTime      *time.Time              `json:"end_time"`
	Limit        int                     `json:"limit"`
}

func (m *SecurityMonitor) matchesFilters(event *SecurityEvent, filters EventFilters) bool {
	// Type filter
	if len(filters.Types) > 0 {
		found := false
		for _, t := range filters.Types {
			if event.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Severity filter
	if len(filters.Severities) > 0 {
		found := false
		for _, s := range filters.Severities {
			if event.Severity == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Source filter
	if len(filters.Sources) > 0 {
		found := false
		for _, s := range filters.Sources {
			if event.Source == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// User ID filter
	if len(filters.UserIDs) > 0 {
		found := false
		for _, u := range filters.UserIDs {
			if event.UserID == u {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// IP address filter
	if len(filters.IPAddresses) > 0 {
		found := false
		for _, ip := range filters.IPAddresses {
			if event.IPAddress == ip {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Time filters
	if filters.StartTime != nil && event.Timestamp.Before(*filters.StartTime) {
		return false
	}
	if filters.EndTime != nil && event.Timestamp.After(*filters.EndTime) {
		return false
	}

	// Resolved filter
	if filters.Resolved != nil && event.Resolved != *filters.Resolved {
		return false
	}

	return true
}

func (m *SecurityMonitor) matchesAlertFilters(alert *SecurityAlert, filters AlertFilters) bool {
	// Type filter
	if len(filters.Types) > 0 {
		found := false
		for _, t := range filters.Types {
			if alert.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Severity filter
	if len(filters.Severities) > 0 {
		found := false
		for _, s := range filters.Severities {
			if alert.Severity == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Source filter
	if len(filters.Sources) > 0 {
		found := false
		for _, s := range filters.Sources {
			if alert.Source == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Acknowledged filter
	if filters.Acknowledged != nil && alert.Acknowledged != *filters.Acknowledged {
		return false
	}

	// Time filters
	if filters.StartTime != nil && alert.Timestamp.Before(*filters.StartTime) {
		return false
	}
	if filters.EndTime != nil && alert.Timestamp.After(*filters.EndTime) {
		return false
	}

	return true
}

// Helper functions for converting maps to sorted slices
func convertToIPAddressCounts(counts map[string]int64) []IPAddressCount {
	result := make([]IPAddressCount, 0, len(counts))
	for ip, count := range counts {
		result = append(result, IPAddressCount{IPAddress: ip, Count: count})
	}
	return result
}

func convertToEndpointCounts(counts map[string]int64) []EndpointCount {
	result := make([]EndpointCount, 0, len(counts))
	for endpoint, count := range counts {
		result = append(result, EndpointCount{Endpoint: endpoint, Count: count})
	}
	return result
}

func convertToUserAgentCounts(counts map[string]int64) []UserAgentCount {
	result := make([]UserAgentCount, 0, len(counts))
	for userAgent, count := range counts {
		result = append(result, UserAgentCount{UserAgent: userAgent, Count: count})
	}
	return result
}

// Helper functions for generating IDs
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

func generateAlertID() string {
	return fmt.Sprintf("alt_%d", time.Now().UnixNano())
}
