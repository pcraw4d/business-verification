package external

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// AuditTrailManager manages verification audit trails and history
type AuditTrailManager struct {
	config *AuditTrailConfig
}

// AuditTrailConfig configures audit trail behavior
type AuditTrailConfig struct {
	MaxHistoryDays     int    `json:"max_history_days"`
	EnableDiskStorage  bool   `json:"enable_disk_storage"`
	EnableDetailedLogs bool   `json:"enable_detailed_logs"`
	RetentionPolicy    string `json:"retention_policy"` // "time_based", "count_based", "permanent"
	MaxEntries         int    `json:"max_entries"`
}

// DefaultAuditTrailConfig returns default configuration
func DefaultAuditTrailConfig() *AuditTrailConfig {
	return &AuditTrailConfig{
		MaxHistoryDays:     90,
		EnableDiskStorage:  true,
		EnableDetailedLogs: true,
		RetentionPolicy:    "time_based",
		MaxEntries:         10000,
	}
}

// NewAuditTrailManager creates a new audit trail manager
func NewAuditTrailManager(config *AuditTrailConfig) *AuditTrailManager {
	if config == nil {
		config = DefaultAuditTrailConfig()
	}
	return &AuditTrailManager{
		config: config,
	}
}

// VerificationHistory represents the complete history of a verification
type VerificationHistory struct {
	VerificationID string                 `json:"verification_id"`
	BusinessName   string                 `json:"business_name"`
	WebsiteURL     string                 `json:"website_url"`
	InitiatedAt    time.Time              `json:"initiated_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	Duration       time.Duration          `json:"duration"`
	FinalStatus    string                 `json:"final_status"`
	FinalScore     float64                `json:"final_score"`
	Events         []AuditEvent           `json:"events"`
	Milestones     []HistoryMilestone     `json:"milestones"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// HistoryMilestone represents key milestones in verification process
type HistoryMilestone struct {
	MilestoneID  string                 `json:"milestone_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Timestamp    time.Time              `json:"timestamp"`
	Status       string                 `json:"status"` // "started", "completed", "failed", "skipped"
	Duration     time.Duration          `json:"duration"`
	Details      map[string]interface{} `json:"details"`
	Dependencies []string               `json:"dependencies,omitempty"`
	CriticalPath bool                   `json:"critical_path"`
}

// AuditQuery defines parameters for querying audit trails
type AuditQuery struct {
	VerificationID string     `json:"verification_id,omitempty"`
	BusinessName   string     `json:"business_name,omitempty"`
	EventType      string     `json:"event_type,omitempty"`
	Severity       string     `json:"severity,omitempty"`
	UserID         string     `json:"user_id,omitempty"`
	StartTime      *time.Time `json:"start_time,omitempty"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
}

// CreateVerificationHistory creates a comprehensive verification history from events
func (m *AuditTrailManager) CreateVerificationHistory(
	ctx context.Context,
	verificationID, businessName, websiteURL string,
	events []AuditEvent,
	metadata map[string]interface{},
) (*VerificationHistory, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("no events provided for verification history")
	}

	// Sort events by timestamp
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	history := &VerificationHistory{
		VerificationID: verificationID,
		BusinessName:   businessName,
		WebsiteURL:     websiteURL,
		InitiatedAt:    events[0].Timestamp,
		Events:         events,
		Metadata:       metadata,
	}

	// Find completion time and final status
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		if event.EventType == "report_generated" || event.EventType == "verification_completed" {
			history.CompletedAt = &event.Timestamp
			history.Duration = event.Timestamp.Sub(history.InitiatedAt)
			break
		}
	}

	// Extract final status and score from events
	for _, event := range events {
		if event.EventType == "status_assigned" {
			if status, ok := event.Data["status"].(string); ok {
				history.FinalStatus = status
			}
			if score, ok := event.Data["overall_score"].(float64); ok {
				history.FinalScore = score
			}
		}
	}

	// Generate milestones
	history.Milestones = m.generateMilestones(events)

	return history, nil
}

// generateMilestones creates key milestones from audit events
func (m *AuditTrailManager) generateMilestones(events []AuditEvent) []HistoryMilestone {
	var milestones []HistoryMilestone
	milestoneMap := make(map[string]*HistoryMilestone)

	for _, event := range events {
		switch event.EventType {
		case "verification_started":
			milestone := &HistoryMilestone{
				MilestoneID:  fmt.Sprintf("milestone_start_%d", event.Timestamp.Unix()),
				Name:         "Verification Initiated",
				Description:  "Business verification process started",
				Timestamp:    event.Timestamp,
				Status:       "completed",
				CriticalPath: true,
				Details: map[string]interface{}{
					"event_id": event.EventID,
				},
			}
			milestoneMap["start"] = milestone

		case "data_extracted":
			milestone := &HistoryMilestone{
				MilestoneID:  fmt.Sprintf("milestone_extract_%d", event.Timestamp.Unix()),
				Name:         "Data Extraction",
				Description:  "Business data extracted from website",
				Timestamp:    event.Timestamp,
				Status:       "completed",
				CriticalPath: true,
				Dependencies: []string{"start"},
				Details:      event.Data,
			}
			if startMilestone, exists := milestoneMap["start"]; exists {
				milestone.Duration = event.Timestamp.Sub(startMilestone.Timestamp)
			}
			milestoneMap["extraction"] = milestone

		case "confidence_calculated":
			milestone := &HistoryMilestone{
				MilestoneID:  fmt.Sprintf("milestone_confidence_%d", event.Timestamp.Unix()),
				Name:         "Confidence Scoring",
				Description:  "Verification confidence calculated",
				Timestamp:    event.Timestamp,
				Status:       "completed",
				CriticalPath: true,
				Dependencies: []string{"extraction"},
				Details:      event.Data,
			}
			if extractMilestone, exists := milestoneMap["extraction"]; exists {
				milestone.Duration = event.Timestamp.Sub(extractMilestone.Timestamp)
			}
			milestoneMap["confidence"] = milestone

		case "status_assigned":
			milestone := &HistoryMilestone{
				MilestoneID:  fmt.Sprintf("milestone_status_%d", event.Timestamp.Unix()),
				Name:         "Status Assignment",
				Description:  "Final verification status determined",
				Timestamp:    event.Timestamp,
				Status:       "completed",
				CriticalPath: true,
				Dependencies: []string{"confidence"},
				Details:      event.Data,
			}
			if confidenceMilestone, exists := milestoneMap["confidence"]; exists {
				milestone.Duration = event.Timestamp.Sub(confidenceMilestone.Timestamp)
			}
			milestoneMap["status"] = milestone

		case "report_generated":
			milestone := &HistoryMilestone{
				MilestoneID:  fmt.Sprintf("milestone_report_%d", event.Timestamp.Unix()),
				Name:         "Report Generation",
				Description:  "Comprehensive verification report generated",
				Timestamp:    event.Timestamp,
				Status:       "completed",
				CriticalPath: true,
				Dependencies: []string{"status"},
				Details:      event.Data,
			}
			if statusMilestone, exists := milestoneMap["status"]; exists {
				milestone.Duration = event.Timestamp.Sub(statusMilestone.Timestamp)
			}
			milestoneMap["report"] = milestone
		}
	}

	// Convert map to slice and sort by timestamp
	for _, milestone := range milestoneMap {
		milestones = append(milestones, *milestone)
	}

	sort.Slice(milestones, func(i, j int) bool {
		return milestones[i].Timestamp.Before(milestones[j].Timestamp)
	})

	return milestones
}

// QueryAuditTrail queries audit events based on the provided criteria
func (m *AuditTrailManager) QueryAuditTrail(
	ctx context.Context,
	query AuditQuery,
	events []AuditEvent,
) ([]AuditEvent, error) {
	var filtered []AuditEvent

	for _, event := range events {
		if m.matchesQuery(event, query) {
			filtered = append(filtered, event)
		}
	}

	// Sort by timestamp (most recent first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Apply pagination
	start := query.Offset
	if start < 0 {
		start = 0
	}
	if start >= len(filtered) {
		return []AuditEvent{}, nil
	}

	end := len(filtered)
	if query.Limit > 0 && start+query.Limit < len(filtered) {
		end = start + query.Limit
	}

	return filtered[start:end], nil
}

// matchesQuery checks if an event matches the query criteria
func (m *AuditTrailManager) matchesQuery(event AuditEvent, query AuditQuery) bool {
	if query.VerificationID != "" {
		if verificationID, ok := event.Data["verification_id"].(string); ok {
			if verificationID != query.VerificationID {
				return false
			}
		}
	}

	if query.EventType != "" && event.EventType != query.EventType {
		return false
	}

	if query.Severity != "" && event.Severity != query.Severity {
		return false
	}

	if query.UserID != "" && event.UserID != query.UserID {
		return false
	}

	if query.StartTime != nil && event.Timestamp.Before(*query.StartTime) {
		return false
	}

	if query.EndTime != nil && event.Timestamp.After(*query.EndTime) {
		return false
	}

	return true
}

// GenerateHistorySummary creates a summary of verification history
func (m *AuditTrailManager) GenerateHistorySummary(history *VerificationHistory) *HistorySummary {
	summary := &HistorySummary{
		VerificationID:   history.VerificationID,
		BusinessName:     history.BusinessName,
		TotalDuration:    history.Duration,
		EventCount:       len(history.Events),
		MilestoneCount:   len(history.Milestones),
		FinalStatus:      history.FinalStatus,
		FinalScore:       history.FinalScore,
		CompletionRate:   m.calculateCompletionRate(history),
		CriticalPathTime: m.calculateCriticalPathTime(history),
	}

	// Count events by type
	summary.EventsByType = make(map[string]int)
	for _, event := range history.Events {
		summary.EventsByType[event.EventType]++
	}

	// Count events by severity
	summary.EventsBySeverity = make(map[string]int)
	for _, event := range history.Events {
		summary.EventsBySeverity[event.Severity]++
	}

	return summary
}

// HistorySummary provides a summary view of verification history
type HistorySummary struct {
	VerificationID   string         `json:"verification_id"`
	BusinessName     string         `json:"business_name"`
	TotalDuration    time.Duration  `json:"total_duration"`
	EventCount       int            `json:"event_count"`
	MilestoneCount   int            `json:"milestone_count"`
	FinalStatus      string         `json:"final_status"`
	FinalScore       float64        `json:"final_score"`
	CompletionRate   float64        `json:"completion_rate"`
	CriticalPathTime time.Duration  `json:"critical_path_time"`
	EventsByType     map[string]int `json:"events_by_type"`
	EventsBySeverity map[string]int `json:"events_by_severity"`
}

// calculateCompletionRate calculates the completion rate based on milestones
func (m *AuditTrailManager) calculateCompletionRate(history *VerificationHistory) float64 {
	if len(history.Milestones) == 0 {
		return 0.0
	}

	completed := 0
	for _, milestone := range history.Milestones {
		if milestone.Status == "completed" {
			completed++
		}
	}

	return float64(completed) / float64(len(history.Milestones))
}

// calculateCriticalPathTime calculates total time for critical path milestones
func (m *AuditTrailManager) calculateCriticalPathTime(history *VerificationHistory) time.Duration {
	var totalTime time.Duration

	for _, milestone := range history.Milestones {
		if milestone.CriticalPath {
			totalTime += milestone.Duration
		}
	}

	return totalTime
}

// GetConfig returns the current audit trail configuration
func (m *AuditTrailManager) GetConfig() *AuditTrailConfig {
	return m.config
}

// UpdateConfig updates the audit trail configuration
func (m *AuditTrailManager) UpdateConfig(config *AuditTrailConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.MaxHistoryDays < 1 {
		return fmt.Errorf("max_history_days must be at least 1")
	}

	if config.MaxEntries < 1 {
		return fmt.Errorf("max_entries must be at least 1")
	}

	m.config = config
	return nil
}
