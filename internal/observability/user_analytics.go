package observability

import (
	"fmt"
	"sync"
	"time"
)

// UserAnalytics handles user behavior analytics
type UserAnalytics struct {
	logger    *Logger
	events    map[string]*UserEvent
	mu        sync.RWMutex
	exporters []UserAnalyticsExporter
	config    *UserAnalyticsConfig
}

// UserEvent represents a user interaction event
type UserEvent struct {
	ID          string
	UserID      string
	SessionID   string
	EventType   string
	EventData   map[string]interface{}
	Timestamp   time.Time
	IPAddress   string
	UserAgent   string
	Page        string
	Duration    time.Duration
	Tags        map[string]string
	Environment string
	Service     string
	Version     string
}

// UserAnalyticsExporter interface for exporting user analytics
type UserAnalyticsExporter interface {
	Export(event *UserEvent) error
	Name() string
}

// UserAnalyticsConfig holds configuration for user analytics
type UserAnalyticsConfig struct {
	Enabled              bool
	TrackPageViews       bool
	TrackClicks          bool
	TrackFormSubmissions bool
	TrackAPIUsage        bool
	AnonymizeIP          bool
	RetentionDays        int
	BatchSize            int
	FlushInterval        time.Duration
}

// GoogleAnalyticsExporter exports events to Google Analytics
type GoogleAnalyticsExporter struct {
	logger *Logger
	config map[string]interface{}
}

// NewGoogleAnalyticsExporter creates a new Google Analytics exporter
func NewGoogleAnalyticsExporter(logger *Logger, config map[string]interface{}) *GoogleAnalyticsExporter {
	return &GoogleAnalyticsExporter{
		logger: logger,
		config: config,
	}
}

// Export exports an event to Google Analytics
func (gae *GoogleAnalyticsExporter) Export(event *UserEvent) error {
	// In a real implementation, this would export to Google Analytics
	gae.logger.Debug("Exporting event to Google Analytics", map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.EventType,
		"user_id":    event.UserID,
		"session_id": event.SessionID,
	})
	return nil
}

// Name returns the exporter name
func (gae *GoogleAnalyticsExporter) Name() string {
	return "google_analytics"
}

// MixpanelExporter exports events to Mixpanel
type MixpanelExporter struct {
	logger *Logger
	config map[string]interface{}
}

// NewMixpanelExporter creates a new Mixpanel exporter
func NewMixpanelExporter(logger *Logger, config map[string]interface{}) *MixpanelExporter {
	return &MixpanelExporter{
		logger: logger,
		config: config,
	}
}

// Export exports an event to Mixpanel
func (me *MixpanelExporter) Export(event *UserEvent) error {
	// In a real implementation, this would export to Mixpanel
	me.logger.Debug("Exporting event to Mixpanel", map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.EventType,
		"user_id":    event.UserID,
		"session_id": event.SessionID,
	})
	return nil
}

// Name returns the exporter name
func (me *MixpanelExporter) Name() string {
	return "mixpanel"
}

// LogExporter exports events to logs
type UserAnalyticsLogExporter struct {
	logger *Logger
}

// NewUserAnalyticsLogExporter creates a new log exporter for user analytics
func NewUserAnalyticsLogExporter(logger *Logger) *UserAnalyticsLogExporter {
	return &UserAnalyticsLogExporter{
		logger: logger,
	}
}

// Export exports an event to logs
func (le *UserAnalyticsLogExporter) Export(event *UserEvent) error {
	le.logger.Info("User event tracked", map[string]interface{}{
		"event_id":    event.ID,
		"user_id":     event.UserID,
		"session_id":  event.SessionID,
		"event_type":  event.EventType,
		"event_data":  event.EventData,
		"timestamp":   event.Timestamp,
		"ip_address":  event.IPAddress,
		"user_agent":  event.UserAgent,
		"page":        event.Page,
		"duration":    event.Duration.String(),
		"tags":        event.Tags,
		"environment": event.Environment,
		"service":     event.Service,
		"version":     event.Version,
	})
	return nil
}

// Name returns the exporter name
func (le *UserAnalyticsLogExporter) Name() string {
	return "log"
}

// NewUserAnalytics creates a new user analytics tracker
func NewUserAnalytics(logger *Logger, config *UserAnalyticsConfig) *UserAnalytics {
	return &UserAnalytics{
		logger:    logger,
		events:    make(map[string]*UserEvent),
		exporters: make([]UserAnalyticsExporter, 0),
		config:    config,
	}
}

// TrackEvent tracks a user event
func (ua *UserAnalytics) TrackEvent(userID, sessionID, eventType string, eventData map[string]interface{}, tags map[string]string) {
	if !ua.config.Enabled {
		return
	}

	ua.mu.Lock()
	defer ua.mu.Unlock()

	eventID := ua.generateEventID(userID, sessionID, eventType)
	now := time.Now()

	event := &UserEvent{
		ID:          eventID,
		UserID:      userID,
		SessionID:   sessionID,
		EventType:   eventType,
		EventData:   eventData,
		Timestamp:   now,
		Tags:        tags,
		Environment: "development",
		Service:     "kyb-platform",
		Version:     "1.0.0",
	}

	ua.events[eventID] = event

	ua.logger.Debug("User event tracked", map[string]interface{}{
		"event_id":   eventID,
		"user_id":    userID,
		"session_id": sessionID,
		"event_type": eventType,
	})

	// Export the event
	ua.exportEvent(event)
}

// TrackPageView tracks a page view event
func (ua *UserAnalytics) TrackPageView(userID, sessionID, page string, duration time.Duration, tags map[string]string) {
	if !ua.config.TrackPageViews {
		return
	}

	eventData := map[string]interface{}{
		"page":     page,
		"duration": duration.String(),
	}

	ua.TrackEvent(userID, sessionID, "page_view", eventData, tags)
}

// TrackClick tracks a click event
func (ua *UserAnalytics) TrackClick(userID, sessionID, element, page string, tags map[string]string) {
	if !ua.config.TrackClicks {
		return
	}

	eventData := map[string]interface{}{
		"element": element,
		"page":    page,
	}

	ua.TrackEvent(userID, sessionID, "click", eventData, tags)
}

// TrackFormSubmission tracks a form submission event
func (ua *UserAnalytics) TrackFormSubmission(userID, sessionID, formName string, success bool, tags map[string]string) {
	if !ua.config.TrackFormSubmissions {
		return
	}

	eventData := map[string]interface{}{
		"form_name": formName,
		"success":   success,
	}

	ua.TrackEvent(userID, sessionID, "form_submission", eventData, tags)
}

// TrackAPIUsage tracks API usage events
func (ua *UserAnalytics) TrackAPIUsage(userID, sessionID, endpoint, method string, statusCode int, duration time.Duration, tags map[string]string) {
	if !ua.config.TrackAPIUsage {
		return
	}

	eventData := map[string]interface{}{
		"endpoint":    endpoint,
		"method":      method,
		"status_code": statusCode,
		"duration":    duration.String(),
	}

	ua.TrackEvent(userID, sessionID, "api_usage", eventData, tags)
}

// TrackBusinessEvent tracks business-specific events
func (ua *UserAnalytics) TrackBusinessEvent(userID, sessionID, eventType string, businessData map[string]interface{}, tags map[string]string) {
	eventData := map[string]interface{}{
		"business_event": true,
		"data":           businessData,
	}

	ua.TrackEvent(userID, sessionID, eventType, eventData, tags)
}

// TrackMerchantEvent tracks merchant-specific events
func (ua *UserAnalytics) TrackMerchantEvent(userID, sessionID, merchantID, eventType string, eventData map[string]interface{}, tags map[string]string) {
	merchantEventData := map[string]interface{}{
		"merchant_id": merchantID,
		"data":        eventData,
	}

	ua.TrackEvent(userID, sessionID, eventType, merchantEventData, tags)
}

// GetEvent returns a specific event by ID
func (ua *UserAnalytics) GetEvent(eventID string) (*UserEvent, bool) {
	ua.mu.RLock()
	defer ua.mu.RUnlock()

	event, exists := ua.events[eventID]
	if !exists {
		return nil, false
	}

	// Return a copy
	return &UserEvent{
		ID:          event.ID,
		UserID:      event.UserID,
		SessionID:   event.SessionID,
		EventType:   event.EventType,
		EventData:   event.EventData,
		Timestamp:   event.Timestamp,
		IPAddress:   event.IPAddress,
		UserAgent:   event.UserAgent,
		Page:        event.Page,
		Duration:    event.Duration,
		Tags:        event.Tags,
		Environment: event.Environment,
		Service:     event.Service,
		Version:     event.Version,
	}, true
}

// GetEventsByUser returns events for a specific user
func (ua *UserAnalytics) GetEventsByUser(userID string) []*UserEvent {
	ua.mu.RLock()
	defer ua.mu.RUnlock()

	var userEvents []*UserEvent
	for _, event := range ua.events {
		if event.UserID == userID {
			userEvents = append(userEvents, &UserEvent{
				ID:          event.ID,
				UserID:      event.UserID,
				SessionID:   event.SessionID,
				EventType:   event.EventType,
				EventData:   event.EventData,
				Timestamp:   event.Timestamp,
				IPAddress:   event.IPAddress,
				UserAgent:   event.UserAgent,
				Page:        event.Page,
				Duration:    event.Duration,
				Tags:        event.Tags,
				Environment: event.Environment,
				Service:     event.Service,
				Version:     event.Version,
			})
		}
	}
	return userEvents
}

// GetEventsBySession returns events for a specific session
func (ua *UserAnalytics) GetEventsBySession(sessionID string) []*UserEvent {
	ua.mu.RLock()
	defer ua.mu.RUnlock()

	var sessionEvents []*UserEvent
	for _, event := range ua.events {
		if event.SessionID == sessionID {
			sessionEvents = append(sessionEvents, &UserEvent{
				ID:          event.ID,
				UserID:      event.UserID,
				SessionID:   event.SessionID,
				EventType:   event.EventType,
				EventData:   event.EventData,
				Timestamp:   event.Timestamp,
				IPAddress:   event.IPAddress,
				UserAgent:   event.UserAgent,
				Page:        event.Page,
				Duration:    event.Duration,
				Tags:        event.Tags,
				Environment: event.Environment,
				Service:     event.Service,
				Version:     event.Version,
			})
		}
	}
	return sessionEvents
}

// GetEventsByType returns events of a specific type
func (ua *UserAnalytics) GetEventsByType(eventType string) []*UserEvent {
	ua.mu.RLock()
	defer ua.mu.RUnlock()

	var typeEvents []*UserEvent
	for _, event := range ua.events {
		if event.EventType == eventType {
			typeEvents = append(typeEvents, &UserEvent{
				ID:          event.ID,
				UserID:      event.UserID,
				SessionID:   event.SessionID,
				EventType:   event.EventType,
				EventData:   event.EventData,
				Timestamp:   event.Timestamp,
				IPAddress:   event.IPAddress,
				UserAgent:   event.UserAgent,
				Page:        event.Page,
				Duration:    event.Duration,
				Tags:        event.Tags,
				Environment: event.Environment,
				Service:     event.Service,
				Version:     event.Version,
			})
		}
	}
	return typeEvents
}

// GetSummary returns user analytics summary
func (ua *UserAnalytics) GetSummary() map[string]interface{} {
	ua.mu.RLock()
	defer ua.mu.RUnlock()

	summary := map[string]interface{}{
		"total_events":    len(ua.events),
		"by_event_type":   make(map[string]int),
		"by_user":         make(map[string]int),
		"by_session":      make(map[string]int),
		"unique_users":    make(map[string]bool),
		"unique_sessions": make(map[string]bool),
		"recent_events":   make([]*UserEvent, 0),
	}

	now := time.Now()
	recentThreshold := now.Add(-24 * time.Hour)

	for _, event := range ua.events {
		// Count by event type
		summary["by_event_type"].(map[string]int)[event.EventType]++

		// Count by user
		summary["by_user"].(map[string]int)[event.UserID]++

		// Count by session
		summary["by_session"].(map[string]int)[event.SessionID]++

		// Track unique users and sessions
		summary["unique_users"].(map[string]bool)[event.UserID] = true
		summary["unique_sessions"].(map[string]bool)[event.SessionID] = true

		// Add recent events (last 24 hours)
		if event.Timestamp.After(recentThreshold) {
			summary["recent_events"] = append(summary["recent_events"].([]*UserEvent), &UserEvent{
				ID:          event.ID,
				UserID:      event.UserID,
				SessionID:   event.SessionID,
				EventType:   event.EventType,
				EventData:   event.EventData,
				Timestamp:   event.Timestamp,
				IPAddress:   event.IPAddress,
				UserAgent:   event.UserAgent,
				Page:        event.Page,
				Duration:    event.Duration,
				Tags:        event.Tags,
				Environment: event.Environment,
				Service:     event.Service,
				Version:     event.Version,
			})
		}
	}

	// Convert unique maps to counts
	summary["unique_user_count"] = len(summary["unique_users"].(map[string]bool))
	summary["unique_session_count"] = len(summary["unique_sessions"].(map[string]bool))
	delete(summary, "unique_users")
	delete(summary, "unique_sessions")

	return summary
}

// AddExporter adds a user analytics exporter
func (ua *UserAnalytics) AddExporter(exporter UserAnalyticsExporter) {
	ua.mu.Lock()
	defer ua.mu.Unlock()

	ua.exporters = append(ua.exporters, exporter)
	ua.logger.Info("User analytics exporter added", map[string]interface{}{
		"exporter": exporter.Name(),
	})
}

// FlushEvents flushes events to exporters
func (ua *UserAnalytics) FlushEvents() {
	ua.mu.RLock()
	events := make([]*UserEvent, 0, len(ua.events))
	for _, event := range ua.events {
		events = append(events, event)
	}
	ua.mu.RUnlock()

	for _, exporter := range ua.exporters {
		for _, event := range events {
			if err := exporter.Export(event); err != nil {
				ua.logger.Error("Failed to export user event", map[string]interface{}{
					"exporter": exporter.Name(),
					"event_id": event.ID,
					"error":    err.Error(),
				})
			}
		}
	}

	ua.logger.Debug("User events flushed", map[string]interface{}{
		"event_count": len(events),
		"exporters":   len(ua.exporters),
	})
}

// exportEvent exports an event using registered exporters
func (ua *UserAnalytics) exportEvent(event *UserEvent) {
	for _, exporter := range ua.exporters {
		if err := exporter.Export(event); err != nil {
			ua.logger.Error("Failed to export user event", map[string]interface{}{
				"exporter": exporter.Name(),
				"event_id": event.ID,
				"error":    err.Error(),
			})
		}
	}
}

// generateEventID generates a unique ID for an event
func (ua *UserAnalytics) generateEventID(userID, sessionID, eventType string) string {
	// Simple ID generation - in real implementation would use proper UUID
	return fmt.Sprintf("event_%s_%s_%s_%d", userID, sessionID, eventType, time.Now().UnixNano())
}

// ClearOldEvents removes events older than the retention period
func (ua *UserAnalytics) ClearOldEvents() {
	ua.mu.Lock()
	defer ua.mu.Unlock()

	now := time.Now()
	retentionThreshold := now.Add(-time.Duration(ua.config.RetentionDays) * 24 * time.Hour)
	count := 0

	for id, event := range ua.events {
		if event.Timestamp.Before(retentionThreshold) {
			delete(ua.events, id)
			count++
		}
	}

	if count > 0 {
		ua.logger.Info("Cleared old user events", map[string]interface{}{
			"count":          count,
			"retention_days": ua.config.RetentionDays,
		})
	}
}

// GetUserJourney returns the user journey for a specific user
func (ua *UserAnalytics) GetUserJourney(userID string) []*UserEvent {
	events := ua.GetEventsByUser(userID)

	// Sort by timestamp
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.After(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	return events
}

// GetSessionJourney returns the session journey for a specific session
func (ua *UserAnalytics) GetSessionJourney(sessionID string) []*UserEvent {
	events := ua.GetEventsBySession(sessionID)

	// Sort by timestamp
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.After(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	return events
}
