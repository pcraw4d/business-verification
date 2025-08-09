package compliance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ComplianceAuditSystem provides comprehensive compliance audit trail functionality
type ComplianceAuditSystem struct {
	logger       *observability.Logger
	mu           sync.RWMutex
	auditTrails  map[string][]ComplianceAuditTrail // businessID -> audit trails
	auditEvents  map[string][]AuditEvent           // businessID -> audit events
	auditReports map[string]*AuditReport           // businessID -> audit report
	auditMetrics map[string]*AuditMetrics          // businessID -> audit metrics
	auditFilters map[string]*AuditFilter           // businessID -> audit filter
}

// AuditEvent represents a compliance audit event
type AuditEvent struct {
	ID            string                 `json:"id"`
	BusinessID    string                 `json:"business_id"`
	EventType     string                 `json:"event_type"`     // "status_change", "requirement_update", "control_test", "evidence_added", "exception_created", "remediation_plan", "framework_mapping", "alert_triggered", "report_generated"
	EventCategory string                 `json:"event_category"` // "status", "requirement", "control", "evidence", "exception", "remediation", "mapping", "alert", "report"
	EntityType    string                 `json:"entity_type"`    // "overall", "framework", "requirement", "control", "evidence", "exception", "remediation", "mapping", "alert"
	EntityID      string                 `json:"entity_id"`
	Action        AuditAction            `json:"action"`
	Description   string                 `json:"description"`
	UserID        string                 `json:"user_id"`
	UserName      string                 `json:"user_name"`
	UserRole      string                 `json:"user_role"`
	UserEmail     string                 `json:"user_email"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
	SessionID     string                 `json:"session_id"`
	RequestID     string                 `json:"request_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Duration      time.Duration          `json:"duration,omitempty"`
	Success       bool                   `json:"success"`
	ErrorCode     string                 `json:"error_code,omitempty"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	OldValue      interface{}            `json:"old_value,omitempty"`
	NewValue      interface{}            `json:"new_value,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Severity      string                 `json:"severity"` // "low", "medium", "high", "critical"
	Impact        string                 `json:"impact"`   // "minimal", "low", "medium", "high", "critical"
	Tags          []string               `json:"tags"`
}

// AuditReport represents a comprehensive compliance audit report
type AuditReport struct {
	ID              string                 `json:"id"`
	BusinessID      string                 `json:"business_id"`
	ReportType      string                 `json:"report_type"` // "summary", "detailed", "activity", "changes", "exceptions", "remediations"
	GeneratedAt     time.Time              `json:"generated_at"`
	GeneratedBy     string                 `json:"generated_by"`
	Period          string                 `json:"period"`
	StartDate       time.Time              `json:"start_date"`
	EndDate         time.Time              `json:"end_date"`
	TotalEvents     int                    `json:"total_events"`
	EventCounts     map[string]int         `json:"event_counts"`
	CategoryCounts  map[string]int         `json:"category_counts"`
	EntityCounts    map[string]int         `json:"entity_counts"`
	ActionCounts    map[string]int         `json:"action_counts"`
	UserCounts      map[string]int         `json:"user_count"`
	SeverityCounts  map[string]int         `json:"severity_counts"`
	ImpactCounts    map[string]int         `json:"impact_counts"`
	SuccessRate     float64                `json:"success_rate"`
	ErrorRate       float64                `json:"error_rate"`
	AverageDuration time.Duration          `json:"average_duration"`
	TotalDuration   time.Duration          `json:"total_duration"`
	Events          []AuditEvent           `json:"events"`
	Summary         AuditSummary           `json:"summary"`
	Trends          AuditTrends            `json:"trends"`
	Anomalies       []AuditAnomaly         `json:"anomalies"`
	Recommendations []AuditRecommendation  `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AuditSummary represents a summary of audit activities
type AuditSummary struct {
	TotalEvents          int           `json:"total_events"`
	SuccessfulEvents     int           `json:"successful_events"`
	FailedEvents         int           `json:"failed_events"`
	UniqueUsers          int           `json:"unique_users"`
	UniqueEntities       int           `json:"unique_entities"`
	MostActiveUser       string        `json:"most_active_user"`
	MostActiveEntity     string        `json:"most_active_entity"`
	MostCommonAction     string        `json:"most_common_action"`
	MostCommonEventType  string        `json:"most_common_event_type"`
	AverageEventsPerDay  float64       `json:"average_events_per_day"`
	PeakActivityTime     time.Time     `json:"peak_activity_time"`
	LowestActivityTime   time.Time     `json:"lowest_activity_time"`
	CriticalEvents       int           `json:"critical_events"`
	HighImpactEvents     int           `json:"high_impact_events"`
	AverageEventDuration time.Duration `json:"average_event_duration"`
	TotalEventDuration   time.Duration `json:"total_event_duration"`
}

// AuditTrends represents trends in audit activities
type AuditTrends struct {
	EventTrend          []AuditTrendPoint `json:"event_trend"`
	UserActivityTrend   []AuditTrendPoint `json:"user_activity_trend"`
	EntityActivityTrend []AuditTrendPoint `json:"entity_activity_trend"`
	ActionTrend         []AuditTrendPoint `json:"action_trend"`
	CategoryTrend       []AuditTrendPoint `json:"category_trend"`
	SeverityTrend       []AuditTrendPoint `json:"severity_trend"`
	ImpactTrend         []AuditTrendPoint `json:"impact_trend"`
	SuccessRateTrend    []AuditTrendPoint `json:"success_rate_trend"`
	DurationTrend       []AuditTrendPoint `json:"duration_trend"`
	PeakActivityPeriods []ActivityPeriod  `json:"peak_activity_periods"`
	LowActivityPeriods  []ActivityPeriod  `json:"low_activity_periods"`
}

// AuditTrendPoint represents a point in audit trend data
type AuditTrendPoint struct {
	Date          time.Time `json:"date"`
	Value         float64   `json:"value"`
	Count         int       `json:"count"`
	Percentage    float64   `json:"percentage"`
	Trend         string    `json:"trend"` // "increasing", "stable", "decreasing"
	TrendStrength float64   `json:"trend_strength"`
}

// ActivityPeriod represents a period of activity
type ActivityPeriod struct {
	StartTime            time.Time     `json:"start_time"`
	EndTime              time.Time     `json:"end_time"`
	Duration             time.Duration `json:"duration"`
	EventCount           int           `json:"event_count"`
	AverageEventsPerHour float64       `json:"average_events_per_hour"`
	PeakEventsPerHour    int           `json:"peak_events_per_hour"`
	UniqueUsers          int           `json:"unique_users"`
	MostActiveUser       string        `json:"most_active_user"`
	MostActiveEntity     string        `json:"most_active_entity"`
}

// AuditAnomaly represents an anomaly in audit activities
type AuditAnomaly struct {
	ID               string      `json:"id"`
	Type             string      `json:"type"`     // "unusual_activity", "failed_events", "critical_events", "user_anomaly", "entity_anomaly", "time_anomaly"
	Severity         string      `json:"severity"` // "low", "medium", "high", "critical"
	Description      string      `json:"description"`
	DetectedAt       time.Time   `json:"detected_at"`
	TimeRange        TimeRange   `json:"time_range"`
	EventCount       int         `json:"event_count"`
	AffectedUsers    []string    `json:"affected_users"`
	AffectedEntities []string    `json:"affected_entities"`
	Pattern          string      `json:"pattern"`
	ExpectedValue    interface{} `json:"expected_value"`
	ActualValue      interface{} `json:"actual_value"`
	Deviation        float64     `json:"deviation"`
	Confidence       float64     `json:"confidence"`
	Recommendation   string      `json:"recommendation"`
	Status           string      `json:"status"` // "open", "investigating", "resolved", "false_positive"
	InvestigatedBy   string      `json:"investigated_by,omitempty"`
	InvestigatedAt   *time.Time  `json:"investigated_at,omitempty"`
	Resolution       string      `json:"resolution,omitempty"`
	ResolvedAt       *time.Time  `json:"resolved_at,omitempty"`
	ResolvedBy       string      `json:"resolved_by,omitempty"`
}

// TimeRange represents a time range
type TimeRange struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}

// AuditRecommendation represents a recommendation based on audit analysis
type AuditRecommendation struct {
	ID               string    `json:"id"`
	Type             string    `json:"type"`     // "security", "compliance", "performance", "monitoring", "investigation"
	Priority         string    `json:"priority"` // "low", "medium", "high", "critical"
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Action           string    `json:"action"`
	Impact           string    `json:"impact"`
	Effort           string    `json:"effort"`
	Timeline         string    `json:"timeline"`
	Evidence         []string  `json:"evidence"`
	RelatedAnomalies []string  `json:"related_anomalies"`
	ExpectedOutcome  string    `json:"expected_outcome"`
	AssignedTo       string    `json:"assigned_to"`
	Status           string    `json:"status"` // "open", "in_progress", "completed", "rejected"
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AuditMetrics represents audit-related metrics
type AuditMetrics struct {
	BusinessID      string         `json:"business_id"`
	TotalEvents     int            `json:"total_events"`
	EventCounts     map[string]int `json:"event_counts"`
	CategoryCounts  map[string]int `json:"category_counts"`
	EntityCounts    map[string]int `json:"entity_counts"`
	ActionCounts    map[string]int `json:"action_counts"`
	UserCounts      map[string]int `json:"user_counts"`
	SeverityCounts  map[string]int `json:"severity_counts"`
	ImpactCounts    map[string]int `json:"impact_counts"`
	SuccessRate     float64        `json:"success_rate"`
	ErrorRate       float64        `json:"error_rate"`
	AverageDuration time.Duration  `json:"average_duration"`
	TotalDuration   time.Duration  `json:"total_duration"`
	TrendData       []TrendPoint   `json:"trend_data"`
	AnomalyCount    int            `json:"anomaly_count"`
	LastCalculated  time.Time      `json:"last_calculated"`
}

// AuditFilter represents audit filtering criteria
type AuditFilter struct {
	BusinessID      string        `json:"business_id"`
	EventTypes      []string      `json:"event_types"`
	EventCategories []string      `json:"event_categories"`
	EntityTypes     []string      `json:"entity_types"`
	EntityIDs       []string      `json:"entity_ids"`
	Actions         []AuditAction `json:"actions"`
	UserIDs         []string      `json:"user_ids"`
	UserRoles       []string      `json:"user_roles"`
	Severities      []string      `json:"severities"`
	Impacts         []string      `json:"impacts"`
	Success         *bool         `json:"success"`
	StartDate       *time.Time    `json:"start_date"`
	EndDate         *time.Time    `json:"end_date"`
	Tags            []string      `json:"tags"`
	Limit           int           `json:"limit"`
	Offset          int           `json:"offset"`
	SortBy          string        `json:"sort_by"`
	SortOrder       string        `json:"sort_order"`
}

// NewComplianceAuditSystem creates a new compliance audit system
func NewComplianceAuditSystem(logger *observability.Logger) *ComplianceAuditSystem {
	return &ComplianceAuditSystem{
		logger:       logger,
		auditTrails:  make(map[string][]ComplianceAuditTrail),
		auditEvents:  make(map[string][]AuditEvent),
		auditReports: make(map[string]*AuditReport),
		auditMetrics: make(map[string]*AuditMetrics),
		auditFilters: make(map[string]*AuditFilter),
	}
}

// RecordAuditEvent records a compliance audit event
func (s *ComplianceAuditSystem) RecordAuditEvent(ctx context.Context, event *AuditEvent) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Recording compliance audit event",
		"request_id", requestID,
		"business_id", event.BusinessID,
		"event_type", event.EventType,
		"event_category", event.EventCategory,
		"entity_type", event.EntityType,
		"entity_id", event.EntityID,
		"action", event.Action,
		"user_id", event.UserID,
		"severity", event.Severity,
		"impact", event.Impact,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate event ID if not provided
	if event.ID == "" {
		event.ID = fmt.Sprintf("audit_event_%s_%d", event.BusinessID, time.Now().UnixNano())
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Add to audit events
	s.auditEvents[event.BusinessID] = append(s.auditEvents[event.BusinessID], *event)

	// Create audit trail entry
	auditTrail := ComplianceAuditTrail{
		ID:            fmt.Sprintf("audit_trail_%s_%d", event.BusinessID, time.Now().UnixNano()),
		BusinessID:    event.BusinessID,
		Framework:     event.EntityType, // Use entity type as framework for now
		RequirementID: &event.EntityID,
		Action:        event.Action,
		Description:   event.Description,
		UserID:        event.UserID,
		UserName:      event.UserName,
		UserRole:      event.UserRole,
		Timestamp:     event.Timestamp,
		IPAddress:     event.IPAddress,
		UserAgent:     event.UserAgent,
		SessionID:     event.SessionID,
		RequestID:     event.RequestID,
		OldValue:      fmt.Sprintf("%v", event.OldValue),
		NewValue:      fmt.Sprintf("%v", event.NewValue),
	}

	s.auditTrails[event.BusinessID] = append(s.auditTrails[event.BusinessID], auditTrail)

	s.logger.Info("Compliance audit event recorded successfully",
		"request_id", requestID,
		"business_id", event.BusinessID,
		"event_id", event.ID,
		"event_type", event.EventType,
		"severity", event.Severity,
		"impact", event.Impact,
	)

	return nil
}

// GetAuditEvents gets audit events for a business with optional filtering
func (s *ComplianceAuditSystem) GetAuditEvents(ctx context.Context, businessID string, filter *AuditFilter) ([]AuditEvent, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting audit events",
		"request_id", requestID,
		"business_id", businessID,
		"filter", filter,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	events, exists := s.auditEvents[businessID]
	if !exists {
		return nil, fmt.Errorf("no audit events found for business %s", businessID)
	}

	if filter == nil {
		return events, nil
	}

	var filteredEvents []AuditEvent
	for _, event := range events {
		if s.matchesFilter(&event, filter) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	// Apply sorting
	if filter.SortBy != "" {
		s.sortEvents(filteredEvents, filter.SortBy, filter.SortOrder)
	}

	// Apply pagination
	if filter.Limit > 0 {
		start := filter.Offset
		end := start + filter.Limit
		if start >= len(filteredEvents) {
			return []AuditEvent{}, nil
		}
		if end > len(filteredEvents) {
			end = len(filteredEvents)
		}
		filteredEvents = filteredEvents[start:end]
	}

	return filteredEvents, nil
}

// GetAuditTrail gets the audit trail for a business
func (s *ComplianceAuditSystem) GetAuditTrail(ctx context.Context, businessID string, startDate, endDate time.Time) ([]ComplianceAuditTrail, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting audit trail",
		"request_id", requestID,
		"business_id", businessID,
		"start_date", startDate,
		"end_date", endDate,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	trails, exists := s.auditTrails[businessID]
	if !exists {
		return nil, fmt.Errorf("no audit trail found for business %s", businessID)
	}

	var filteredTrails []ComplianceAuditTrail
	for _, trail := range trails {
		if trail.Timestamp.After(startDate) && trail.Timestamp.Before(endDate) {
			filteredTrails = append(filteredTrails, trail)
		}
	}

	return filteredTrails, nil
}

// GenerateAuditReport generates a comprehensive audit report
func (s *ComplianceAuditSystem) GenerateAuditReport(ctx context.Context, businessID string, reportType string, startDate, endDate time.Time) (*AuditReport, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Generating audit report",
		"request_id", requestID,
		"business_id", businessID,
		"report_type", reportType,
		"start_date", startDate,
		"end_date", endDate,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	events, exists := s.auditEvents[businessID]
	if !exists {
		return nil, fmt.Errorf("no audit events found for business %s", businessID)
	}

	// Filter events by date range
	var filteredEvents []AuditEvent
	for _, event := range events {
		if event.Timestamp.After(startDate) && event.Timestamp.Before(endDate) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	report := &AuditReport{
		ID:          fmt.Sprintf("audit_report_%s_%d", businessID, time.Now().Unix()),
		BusinessID:  businessID,
		ReportType:  reportType,
		GeneratedAt: time.Now(),
		GeneratedBy: "system",
		Period:      fmt.Sprintf("%s_to_%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		StartDate:   startDate,
		EndDate:     endDate,
		TotalEvents: len(filteredEvents),
		Events:      filteredEvents,
	}

	// Calculate metrics
	s.calculateAuditMetrics(report, filteredEvents)

	// Generate summary
	report.Summary = s.generateAuditSummary(filteredEvents)

	// Generate trends
	report.Trends = s.generateAuditTrends(filteredEvents, startDate, endDate)

	// Detect anomalies
	report.Anomalies = s.detectAuditAnomalies(filteredEvents)

	// Generate recommendations
	report.Recommendations = s.generateAuditRecommendations(report)

	s.auditReports[businessID] = report

	s.logger.Info("Audit report generated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"report_type", reportType,
		"total_events", report.TotalEvents,
		"anomaly_count", len(report.Anomalies),
		"recommendation_count", len(report.Recommendations),
	)

	return report, nil
}

// GetAuditMetrics gets audit metrics for a business
func (s *ComplianceAuditSystem) GetAuditMetrics(ctx context.Context, businessID string) (*AuditMetrics, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting audit metrics",
		"request_id", requestID,
		"business_id", businessID,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics, exists := s.auditMetrics[businessID]
	if !exists {
		return nil, fmt.Errorf("no audit metrics found for business %s", businessID)
	}

	return metrics, nil
}

// UpdateAuditMetrics updates audit metrics for a business
func (s *ComplianceAuditSystem) UpdateAuditMetrics(ctx context.Context, businessID string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating audit metrics",
		"request_id", requestID,
		"business_id", businessID,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	events, exists := s.auditEvents[businessID]
	if !exists {
		return fmt.Errorf("no audit events found for business %s", businessID)
	}

	metrics := &AuditMetrics{
		BusinessID:     businessID,
		TotalEvents:    len(events),
		EventCounts:    make(map[string]int),
		CategoryCounts: make(map[string]int),
		EntityCounts:   make(map[string]int),
		ActionCounts:   make(map[string]int),
		UserCounts:     make(map[string]int),
		SeverityCounts: make(map[string]int),
		ImpactCounts:   make(map[string]int),
		LastCalculated: time.Now(),
	}

	successCount := 0
	totalDuration := time.Duration(0)

	for _, event := range events {
		// Count by event type
		metrics.EventCounts[event.EventType]++

		// Count by category
		metrics.CategoryCounts[event.EventCategory]++

		// Count by entity type
		metrics.EntityCounts[event.EntityType]++

		// Count by action
		metrics.ActionCounts[string(event.Action)]++

		// Count by user
		metrics.UserCounts[event.UserID]++

		// Count by severity
		metrics.SeverityCounts[event.Severity]++

		// Count by impact
		metrics.ImpactCounts[event.Impact]++

		// Calculate success rate
		if event.Success {
			successCount++
		}

		// Calculate total duration
		totalDuration += event.Duration
	}

	// Calculate rates
	if metrics.TotalEvents > 0 {
		metrics.SuccessRate = float64(successCount) / float64(metrics.TotalEvents) * 100.0
		metrics.ErrorRate = 100.0 - metrics.SuccessRate
		metrics.AverageDuration = totalDuration / time.Duration(metrics.TotalEvents)
		metrics.TotalDuration = totalDuration
	}

	s.auditMetrics[businessID] = metrics

	s.logger.Info("Audit metrics updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"total_events", metrics.TotalEvents,
		"success_rate", metrics.SuccessRate,
		"error_rate", metrics.ErrorRate,
	)

	return nil
}

// Helper methods
func (s *ComplianceAuditSystem) matchesFilter(event *AuditEvent, filter *AuditFilter) bool {
	// Event type filter
	if len(filter.EventTypes) > 0 {
		found := false
		for _, eventType := range filter.EventTypes {
			if event.EventType == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Event category filter
	if len(filter.EventCategories) > 0 {
		found := false
		for _, category := range filter.EventCategories {
			if event.EventCategory == category {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Entity type filter
	if len(filter.EntityTypes) > 0 {
		found := false
		for _, entityType := range filter.EntityTypes {
			if event.EntityType == entityType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Entity ID filter
	if len(filter.EntityIDs) > 0 {
		found := false
		for _, entityID := range filter.EntityIDs {
			if event.EntityID == entityID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Action filter
	if len(filter.Actions) > 0 {
		found := false
		for _, action := range filter.Actions {
			if event.Action == action {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// User ID filter
	if len(filter.UserIDs) > 0 {
		found := false
		for _, userID := range filter.UserIDs {
			if event.UserID == userID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// User role filter
	if len(filter.UserRoles) > 0 {
		found := false
		for _, role := range filter.UserRoles {
			if event.UserRole == role {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Severity filter
	if len(filter.Severities) > 0 {
		found := false
		for _, severity := range filter.Severities {
			if event.Severity == severity {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Impact filter
	if len(filter.Impacts) > 0 {
		found := false
		for _, impact := range filter.Impacts {
			if event.Impact == impact {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Success filter
	if filter.Success != nil {
		if event.Success != *filter.Success {
			return false
		}
	}

	// Date range filter
	if filter.StartDate != nil && event.Timestamp.Before(*filter.StartDate) {
		return false
	}
	if filter.EndDate != nil && event.Timestamp.After(*filter.EndDate) {
		return false
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		found := false
		for _, tag := range filter.Tags {
			for _, eventTag := range event.Tags {
				if eventTag == tag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (s *ComplianceAuditSystem) sortEvents(events []AuditEvent, sortBy, sortOrder string) {
	// Simple sorting implementation
	// In a real implementation, this would use a more sophisticated sorting algorithm
	switch sortBy {
	case "timestamp":
		if sortOrder == "desc" {
			// Sort by timestamp descending
			for i := 0; i < len(events)-1; i++ {
				for j := i + 1; j < len(events); j++ {
					if events[i].Timestamp.Before(events[j].Timestamp) {
						events[i], events[j] = events[j], events[i]
					}
				}
			}
		} else {
			// Sort by timestamp ascending
			for i := 0; i < len(events)-1; i++ {
				for j := i + 1; j < len(events); j++ {
					if events[i].Timestamp.After(events[j].Timestamp) {
						events[i], events[j] = events[j], events[i]
					}
				}
			}
		}
	}
}

func (s *ComplianceAuditSystem) calculateAuditMetrics(report *AuditReport, events []AuditEvent) {
	report.EventCounts = make(map[string]int)
	report.CategoryCounts = make(map[string]int)
	report.EntityCounts = make(map[string]int)
	report.ActionCounts = make(map[string]int)
	report.UserCounts = make(map[string]int)
	report.SeverityCounts = make(map[string]int)
	report.ImpactCounts = make(map[string]int)

	successCount := 0
	totalDuration := time.Duration(0)

	for _, event := range events {
		report.EventCounts[event.EventType]++
		report.CategoryCounts[event.EventCategory]++
		report.EntityCounts[event.EntityType]++
		report.ActionCounts[string(event.Action)]++
		report.UserCounts[event.UserID]++
		report.SeverityCounts[event.Severity]++
		report.ImpactCounts[event.Impact]++

		if event.Success {
			successCount++
		}

		totalDuration += event.Duration
	}

	if len(events) > 0 {
		report.SuccessRate = float64(successCount) / float64(len(events)) * 100.0
		report.ErrorRate = 100.0 - report.SuccessRate
		report.AverageDuration = totalDuration / time.Duration(len(events))
		report.TotalDuration = totalDuration
	}
}

func (s *ComplianceAuditSystem) generateAuditSummary(events []AuditEvent) AuditSummary {
	summary := AuditSummary{
		TotalEvents: len(events),
	}

	userCounts := make(map[string]int)
	entityCounts := make(map[string]int)
	actionCounts := make(map[string]int)
	eventTypeCounts := make(map[string]int)
	successCount := 0
	totalDuration := time.Duration(0)
	criticalEvents := 0
	highImpactEvents := 0

	for _, event := range events {
		userCounts[event.UserID]++
		entityCounts[event.EntityID]++
		actionCounts[string(event.Action)]++
		eventTypeCounts[event.EventType]++

		if event.Success {
			successCount++
		}

		totalDuration += event.Duration

		if event.Severity == "critical" {
			criticalEvents++
		}

		if event.Impact == "high" || event.Impact == "critical" {
			highImpactEvents++
		}
	}

	summary.SuccessfulEvents = successCount
	summary.FailedEvents = len(events) - successCount
	summary.UniqueUsers = len(userCounts)
	summary.UniqueEntities = len(entityCounts)
	summary.CriticalEvents = criticalEvents
	summary.HighImpactEvents = highImpactEvents
	summary.TotalEventDuration = totalDuration

	if len(events) > 0 {
		summary.AverageEventDuration = totalDuration / time.Duration(len(events))
		summary.AverageEventsPerDay = float64(len(events)) / 30.0 // Assuming 30 days
	}

	// Find most active user
	maxUserCount := 0
	for userID, count := range userCounts {
		if count > maxUserCount {
			maxUserCount = count
			summary.MostActiveUser = userID
		}
	}

	// Find most active entity
	maxEntityCount := 0
	for entityID, count := range entityCounts {
		if count > maxEntityCount {
			maxEntityCount = count
			summary.MostActiveEntity = entityID
		}
	}

	// Find most common action
	maxActionCount := 0
	for action, count := range actionCounts {
		if count > maxActionCount {
			maxActionCount = count
			summary.MostCommonAction = action
		}
	}

	// Find most common event type
	maxEventTypeCount := 0
	for eventType, count := range eventTypeCounts {
		if count > maxEventTypeCount {
			maxEventTypeCount = count
			summary.MostCommonEventType = eventType
		}
	}

	return summary
}

func (s *ComplianceAuditSystem) generateAuditTrends(events []AuditEvent, startDate, endDate time.Time) AuditTrends {
	trends := AuditTrends{}

	// Generate trend data for different metrics
	trends.EventTrend = s.generateTrendData(events, "event_count", startDate, endDate)
	trends.UserActivityTrend = s.generateTrendData(events, "user_activity", startDate, endDate)
	trends.EntityActivityTrend = s.generateTrendData(events, "entity_activity", startDate, endDate)
	trends.ActionTrend = s.generateTrendData(events, "action_count", startDate, endDate)
	trends.CategoryTrend = s.generateTrendData(events, "category_count", startDate, endDate)
	trends.SeverityTrend = s.generateTrendData(events, "severity_count", startDate, endDate)
	trends.ImpactTrend = s.generateTrendData(events, "impact_count", startDate, endDate)
	trends.SuccessRateTrend = s.generateTrendData(events, "success_rate", startDate, endDate)
	trends.DurationTrend = s.generateTrendData(events, "duration", startDate, endDate)

	// Generate activity periods
	trends.PeakActivityPeriods = s.generateActivityPeriods(events, "peak")
	trends.LowActivityPeriods = s.generateActivityPeriods(events, "low")

	return trends
}

func (s *ComplianceAuditSystem) generateTrendData(events []AuditEvent, metric string, startDate, endDate time.Time) []AuditTrendPoint {
	var trendPoints []AuditTrendPoint

	// Simple trend generation - in a real implementation, this would be more sophisticated
	duration := endDate.Sub(startDate)
	days := int(duration.Hours() / 24)

	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		count := 0
		value := 0.0

		for _, event := range events {
			if event.Timestamp.Year() == date.Year() &&
				event.Timestamp.YearDay() == date.YearDay() {
				count++
				switch metric {
				case "event_count":
					value = float64(count)
				case "user_activity":
					value = float64(count) // Simplified
				case "entity_activity":
					value = float64(count) // Simplified
				case "action_count":
					value = float64(count) // Simplified
				case "category_count":
					value = float64(count) // Simplified
				case "severity_count":
					value = float64(count) // Simplified
				case "impact_count":
					value = float64(count) // Simplified
				case "success_rate":
					value = 85.0 // Simplified
				case "duration":
					value = 100.0 // Simplified
				}
			}
		}

		trendPoint := AuditTrendPoint{
			Date:          date,
			Value:         value,
			Count:         count,
			Percentage:    float64(count) / float64(len(events)) * 100.0,
			Trend:         "stable",
			TrendStrength: 0.5,
		}

		trendPoints = append(trendPoints, trendPoint)
	}

	return trendPoints
}

func (s *ComplianceAuditSystem) generateActivityPeriods(events []AuditEvent, periodType string) []ActivityPeriod {
	var periods []ActivityPeriod

	// Simple activity period generation - in a real implementation, this would be more sophisticated
	if len(events) > 0 {
		period := ActivityPeriod{
			StartTime:            events[0].Timestamp,
			EndTime:              events[len(events)-1].Timestamp,
			Duration:             events[len(events)-1].Timestamp.Sub(events[0].Timestamp),
			EventCount:           len(events),
			AverageEventsPerHour: float64(len(events)) / 24.0, // Simplified
			PeakEventsPerHour:    len(events),                 // Simplified
			UniqueUsers:          1,                           // Simplified
			MostActiveUser:       events[0].UserID,
			MostActiveEntity:     events[0].EntityID,
		}

		periods = append(periods, period)
	}

	return periods
}

func (s *ComplianceAuditSystem) detectAuditAnomalies(events []AuditEvent) []AuditAnomaly {
	var anomalies []AuditAnomaly

	// Simple anomaly detection - in a real implementation, this would use more sophisticated algorithms
	if len(events) > 10 {
		// Detect unusual activity patterns
		anomaly := AuditAnomaly{
			ID:             fmt.Sprintf("anomaly_%d", time.Now().UnixNano()),
			Type:           "unusual_activity",
			Severity:       "medium",
			Description:    "Detected unusual activity pattern in audit events",
			DetectedAt:     time.Now(),
			EventCount:     len(events),
			Pattern:        "high_event_count",
			ExpectedValue:  5,
			ActualValue:    len(events),
			Deviation:      float64(len(events)) - 5.0,
			Confidence:     0.75,
			Recommendation: "Review recent audit events for unusual patterns",
			Status:         "open",
		}

		anomalies = append(anomalies, anomaly)
	}

	return anomalies
}

func (s *ComplianceAuditSystem) generateAuditRecommendations(report *AuditReport) []AuditRecommendation {
	var recommendations []AuditRecommendation

	// Generate recommendations based on audit analysis
	if report.SuccessRate < 90.0 {
		recommendations = append(recommendations, AuditRecommendation{
			ID:              fmt.Sprintf("rec_%s_low_success_rate", report.BusinessID),
			Type:            "performance",
			Priority:        "high",
			Title:           "Low Audit Success Rate",
			Description:     fmt.Sprintf("Audit success rate is %.1f%%, below the target of 90%%", report.SuccessRate),
			Action:          "Investigate failed audit events and improve error handling",
			Impact:          "High - Improves audit reliability",
			Effort:          "Medium - Requires investigation and fixes",
			Timeline:        "1-2 weeks",
			Evidence:        []string{"Low success rate in audit events"},
			ExpectedOutcome: "Improved audit success rate",
			Status:          "open",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		})
	}

	if len(report.Anomalies) > 0 {
		recommendations = append(recommendations, AuditRecommendation{
			ID:              fmt.Sprintf("rec_%s_anomalies_detected", report.BusinessID),
			Type:            "security",
			Priority:        "high",
			Title:           "Audit Anomalies Detected",
			Description:     fmt.Sprintf("Detected %d anomalies in audit events", len(report.Anomalies)),
			Action:          "Investigate detected anomalies and take appropriate action",
			Impact:          "High - Improves security posture",
			Effort:          "High - Requires investigation and remediation",
			Timeline:        "1-3 weeks",
			Evidence:        []string{"Multiple anomalies detected"},
			ExpectedOutcome: "Resolved anomalies and improved security",
			Status:          "open",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		})
	}

	return recommendations
}
