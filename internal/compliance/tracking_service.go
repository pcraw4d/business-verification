package compliance

import (
	"context"
	"fmt"
	"sort"
	"time"

	"kyb-platform/internal/observability"
)

// ComplianceTrackingService provides comprehensive compliance progress and milestone tracking
type ComplianceTrackingService struct {
	logger           *observability.Logger
	frameworkService *ComplianceFrameworkService
	trackingData     map[string]*ComplianceTracking
	milestones       map[string]*ComplianceMilestone
	progressMetrics  map[string]*ComplianceProgressMetrics
}

// ComplianceTracking represents the tracking state for a business's compliance
type ComplianceTracking struct {
	BusinessID      string                 `json:"business_id"`
	FrameworkID     string                 `json:"framework_id"`
	OverallProgress float64                `json:"overall_progress"` // 0.0 to 1.0
	ComplianceLevel string                 `json:"compliance_level"` // "compliant", "partial", "non_compliant", "in_progress"
	LastUpdated     time.Time              `json:"last_updated"`
	Requirements    []RequirementTracking  `json:"requirements"`
	Milestones      []string               `json:"milestones"` // Milestone IDs
	NextReviewDate  *time.Time             `json:"next_review_date,omitempty"`
	RiskLevel       string                 `json:"risk_level"` // "low", "medium", "high", "critical"
	Trend           string                 `json:"trend"`      // "improving", "stable", "declining"
	Velocity        float64                `json:"velocity"`   // Progress velocity (progress per day)
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// RequirementTracking represents tracking for a specific requirement
type RequirementTracking struct {
	RequirementID        string                 `json:"requirement_id"`
	Progress             float64                `json:"progress"` // 0.0 to 1.0
	Status               string                 `json:"status"`   // "not_started", "in_progress", "completed", "at_risk"
	LastAssessed         time.Time              `json:"last_assessed"`
	NextDueDate          *time.Time             `json:"next_due_date,omitempty"`
	EvidenceCount        int                    `json:"evidence_count"`
	FindingsCount        int                    `json:"findings_count"`
	RecommendationsCount int                    `json:"recommendations_count"`
	RiskScore            float64                `json:"risk_score"` // 0.0 to 1.0
	Trend                string                 `json:"trend"`      // "improving", "stable", "declining"
	Velocity             float64                `json:"velocity"`   // Progress velocity
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceMilestone represents a compliance milestone or checkpoint
type ComplianceMilestone struct {
	ID            string                 `json:"id"`
	BusinessID    string                 `json:"business_id"`
	FrameworkID   string                 `json:"framework_id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Type          string                 `json:"type"`   // "assessment", "review", "audit", "certification"
	Status        string                 `json:"status"` // "planned", "in_progress", "completed", "overdue", "cancelled"
	TargetDate    time.Time              `json:"target_date"`
	CompletedDate *time.Time             `json:"completed_date,omitempty"`
	Progress      float64                `json:"progress"`     // 0.0 to 1.0
	Requirements  []string               `json:"requirements"` // Requirement IDs
	Dependencies  []string               `json:"dependencies"` // Other milestone IDs
	Owner         string                 `json:"owner"`
	Priority      string                 `json:"priority"`          // "critical", "high", "medium", "low"
	Outcome       string                 `json:"outcome,omitempty"` // "passed", "failed", "partial", "deferred"
	Notes         string                 `json:"notes,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ComplianceProgressMetrics represents aggregated progress metrics
type ComplianceProgressMetrics struct {
	BusinessID            string                 `json:"business_id"`
	FrameworkID           string                 `json:"framework_id"`
	Period                string                 `json:"period"` // "daily", "weekly", "monthly", "quarterly"
	StartDate             time.Time              `json:"start_date"`
	EndDate               time.Time              `json:"end_date"`
	OverallProgress       float64                `json:"overall_progress"`
	RequirementsCompleted int                    `json:"requirements_completed"`
	RequirementsTotal     int                    `json:"requirements_total"`
	MilestonesCompleted   int                    `json:"milestones_completed"`
	MilestonesTotal       int                    `json:"milestones_total"`
	AverageVelocity       float64                `json:"average_velocity"`
	RiskTrend             string                 `json:"risk_trend"`       // "improving", "stable", "declining"
	ComplianceTrend       string                 `json:"compliance_trend"` // "improving", "stable", "declining"
	KeyMetrics            map[string]float64     `json:"key_metrics"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

// TrackingQuery represents query parameters for tracking operations
type TrackingQuery struct {
	BusinessID      string     `json:"business_id,omitempty"`
	FrameworkID     string     `json:"framework_id,omitempty"`
	Status          string     `json:"status,omitempty"`
	RiskLevel       string     `json:"risk_level,omitempty"`
	Trend           string     `json:"trend,omitempty"`
	IncludeInactive bool       `json:"include_inactive,omitempty"`
	StartDate       *time.Time `json:"start_date,omitempty"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	Limit           int        `json:"limit,omitempty"`
	Offset          int        `json:"offset,omitempty"`
}

// MilestoneQuery represents query parameters for milestone operations
type MilestoneQuery struct {
	BusinessID  string     `json:"business_id,omitempty"`
	FrameworkID string     `json:"framework_id,omitempty"`
	Status      string     `json:"status,omitempty"`
	Type        string     `json:"type,omitempty"`
	Priority    string     `json:"priority,omitempty"`
	Overdue     bool       `json:"overdue,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// NewComplianceTrackingService creates a new compliance tracking service
func NewComplianceTrackingService(logger *observability.Logger, frameworkService *ComplianceFrameworkService) *ComplianceTrackingService {
	return &ComplianceTrackingService{
		logger:           logger,
		frameworkService: frameworkService,
		trackingData:     make(map[string]*ComplianceTracking),
		milestones:       make(map[string]*ComplianceMilestone),
		progressMetrics:  make(map[string]*ComplianceProgressMetrics),
	}
}

// GetComplianceTracking retrieves compliance tracking for a business and framework
func (cts *ComplianceTrackingService) GetComplianceTracking(ctx context.Context, businessID, frameworkID string) (*ComplianceTracking, error) {
	cts.logger.Info("Retrieving compliance tracking", map[string]interface{}{
		"business_id":  businessID,
		"framework_id": frameworkID,
	})

	trackingKey := fmt.Sprintf("%s:%s", businessID, frameworkID)
	tracking, exists := cts.trackingData[trackingKey]

	if !exists {
		// Create new tracking record
		tracking = cts.initializeTracking(businessID, frameworkID)
		cts.trackingData[trackingKey] = tracking
		// Update tracking data for new records
		cts.updateTrackingProgress(tracking)
	}

	cts.logger.Info("Retrieved compliance tracking", map[string]interface{}{
		"business_id":      businessID,
		"framework_id":     frameworkID,
		"overall_progress": tracking.OverallProgress,
		"compliance_level": tracking.ComplianceLevel,
		"risk_level":       tracking.RiskLevel,
	})

	return tracking, nil
}

// UpdateComplianceTracking updates compliance tracking data
func (cts *ComplianceTrackingService) UpdateComplianceTracking(ctx context.Context, tracking *ComplianceTracking) error {
	// Validate required fields
	if tracking.BusinessID == "" {
		return fmt.Errorf("business_id is required")
	}
	if tracking.FrameworkID == "" {
		return fmt.Errorf("framework_id is required")
	}

	cts.logger.Info("Updating compliance tracking", map[string]interface{}{
		"business_id":  tracking.BusinessID,
		"framework_id": tracking.FrameworkID,
	})

	trackingKey := fmt.Sprintf("%s:%s", tracking.BusinessID, tracking.FrameworkID)

	// Update timestamp
	tracking.UpdatedAt = time.Now()

	// Only recalculate metrics if not already set
	if tracking.OverallProgress == 0.0 && len(tracking.Requirements) > 0 {
		cts.updateTrackingProgress(tracking)
	}

	// Store updated tracking
	cts.trackingData[trackingKey] = tracking

	cts.logger.Info("Updated compliance tracking", map[string]interface{}{
		"business_id":      tracking.BusinessID,
		"framework_id":     tracking.FrameworkID,
		"overall_progress": tracking.OverallProgress,
		"compliance_level": tracking.ComplianceLevel,
	})

	return nil
}

// GetComplianceMilestones retrieves compliance milestones for a business
func (cts *ComplianceTrackingService) GetComplianceMilestones(ctx context.Context, query *MilestoneQuery) ([]*ComplianceMilestone, error) {
	cts.logger.Info("Retrieving compliance milestones", map[string]interface{}{
		"query": query,
	})

	var milestones []*ComplianceMilestone

	for _, milestone := range cts.milestones {
		// Apply filters
		if query.BusinessID != "" && milestone.BusinessID != query.BusinessID {
			continue
		}
		if query.FrameworkID != "" && milestone.FrameworkID != query.FrameworkID {
			continue
		}
		if query.Status != "" && milestone.Status != query.Status {
			continue
		}
		if query.Type != "" && milestone.Type != query.Type {
			continue
		}
		if query.Priority != "" && milestone.Priority != query.Priority {
			continue
		}
		if query.Overdue && milestone.Status != "overdue" {
			// Check if milestone is overdue
			if milestone.TargetDate.Before(time.Now()) && milestone.Status != "completed" {
				milestone.Status = "overdue"
			} else {
				continue
			}
		}
		if query.StartDate != nil && milestone.TargetDate.Before(*query.StartDate) {
			continue
		}
		if query.EndDate != nil && milestone.TargetDate.After(*query.EndDate) {
			continue
		}

		milestones = append(milestones, milestone)
	}

	// Sort by target date
	sort.Slice(milestones, func(i, j int) bool {
		return milestones[i].TargetDate.Before(milestones[j].TargetDate)
	})

	// Apply pagination
	if query.Limit > 0 {
		start := query.Offset
		end := query.Offset + query.Limit
		if start >= len(milestones) {
			milestones = []*ComplianceMilestone{}
		} else if end > len(milestones) {
			milestones = milestones[start:]
		} else {
			milestones = milestones[start:end]
		}
	}

	cts.logger.Info("Retrieved compliance milestones", map[string]interface{}{
		"count": len(milestones),
		"query": query,
	})

	return milestones, nil
}

// CreateMilestone creates a new compliance milestone
func (cts *ComplianceTrackingService) CreateMilestone(ctx context.Context, milestone *ComplianceMilestone) error {
	cts.logger.Info("Creating compliance milestone", map[string]interface{}{
		"business_id":  milestone.BusinessID,
		"framework_id": milestone.FrameworkID,
		"name":         milestone.Name,
		"type":         milestone.Type,
	})

	// Set timestamps
	now := time.Now()
	milestone.CreatedAt = now
	milestone.UpdatedAt = now

	// Set default status if not provided
	if milestone.Status == "" {
		milestone.Status = "planned"
	}

	// Store milestone
	cts.milestones[milestone.ID] = milestone

	cts.logger.Info("Created compliance milestone", map[string]interface{}{
		"milestone_id": milestone.ID,
		"business_id":  milestone.BusinessID,
		"framework_id": milestone.FrameworkID,
		"name":         milestone.Name,
	})

	return nil
}

// UpdateMilestone updates an existing compliance milestone
func (cts *ComplianceTrackingService) UpdateMilestone(ctx context.Context, milestone *ComplianceMilestone) error {
	cts.logger.Info("Updating compliance milestone", map[string]interface{}{
		"milestone_id": milestone.ID,
		"business_id":  milestone.BusinessID,
	})

	// Check if milestone exists
	if _, exists := cts.milestones[milestone.ID]; !exists {
		return fmt.Errorf("milestone not found: %s", milestone.ID)
	}

	// Update timestamp
	milestone.UpdatedAt = time.Now()

	// Store updated milestone
	cts.milestones[milestone.ID] = milestone

	cts.logger.Info("Updated compliance milestone", map[string]interface{}{
		"milestone_id": milestone.ID,
		"business_id":  milestone.BusinessID,
		"status":       milestone.Status,
		"progress":     milestone.Progress,
	})

	return nil
}

// GetProgressMetrics retrieves progress metrics for a business
func (cts *ComplianceTrackingService) GetProgressMetrics(ctx context.Context, businessID, frameworkID, period string) (*ComplianceProgressMetrics, error) {
	cts.logger.Info("Retrieving progress metrics", map[string]interface{}{
		"business_id":  businessID,
		"framework_id": frameworkID,
		"period":       period,
	})

	metricsKey := fmt.Sprintf("%s:%s:%s", businessID, frameworkID, period)
	metrics, exists := cts.progressMetrics[metricsKey]

	if !exists {
		// Generate new metrics
		metrics = cts.generateProgressMetrics(businessID, frameworkID, period)
		cts.progressMetrics[metricsKey] = metrics
	}

	cts.logger.Info("Retrieved progress metrics", map[string]interface{}{
		"business_id":      businessID,
		"framework_id":     frameworkID,
		"period":           period,
		"overall_progress": metrics.OverallProgress,
		"velocity":         metrics.AverageVelocity,
	})

	return metrics, nil
}

// GetComplianceTrends retrieves compliance trends for a business
func (cts *ComplianceTrackingService) GetComplianceTrends(ctx context.Context, businessID, frameworkID string, days int) ([]*ComplianceProgressMetrics, error) {
	cts.logger.Info("Retrieving compliance trends", map[string]interface{}{
		"business_id":  businessID,
		"framework_id": frameworkID,
		"days":         days,
	})

	var trends []*ComplianceProgressMetrics
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Generate daily metrics for the period
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		period := d.Format("2006-01-02")
		metrics := cts.generateProgressMetrics(businessID, frameworkID, period)
		trends = append(trends, metrics)
	}

	cts.logger.Info("Retrieved compliance trends", map[string]interface{}{
		"business_id":  businessID,
		"framework_id": frameworkID,
		"days":         days,
		"data_points":  len(trends),
	})

	return trends, nil
}

// Helper methods

// initializeTracking initializes tracking for a new business/framework combination
func (cts *ComplianceTrackingService) initializeTracking(businessID, frameworkID string) *ComplianceTracking {
	now := time.Now()

	tracking := &ComplianceTracking{
		BusinessID:      businessID,
		FrameworkID:     frameworkID,
		OverallProgress: 0.0,
		ComplianceLevel: "non_compliant",
		LastUpdated:     now,
		RiskLevel:       "high",
		Trend:           "stable",
		Velocity:        0.0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Get framework requirements
	requirements, err := cts.frameworkService.GetFrameworkRequirements(context.Background(), frameworkID)
	if err == nil {
		for _, req := range requirements {
			reqTracking := RequirementTracking{
				RequirementID:        req.ID,
				Progress:             0.0,
				Status:               "not_started",
				LastAssessed:         now,
				EvidenceCount:        0,
				FindingsCount:        0,
				RecommendationsCount: 0,
				RiskScore:            1.0, // High risk initially
				Trend:                "stable",
				Velocity:             0.0,
			}
			tracking.Requirements = append(tracking.Requirements, reqTracking)
		}
	}

	return tracking
}

// updateTrackingProgress recalculates tracking progress and metrics
func (cts *ComplianceTrackingService) updateTrackingProgress(tracking *ComplianceTracking) {
	if len(tracking.Requirements) == 0 {
		return
	}

	// Calculate overall progress
	totalProgress := 0.0
	completedCount := 0
	atRiskCount := 0

	for i := range tracking.Requirements {
		req := &tracking.Requirements[i]

		// Update requirement status based on progress
		if req.Progress >= 1.0 {
			req.Status = "completed"
			completedCount++
		} else if req.Progress > 0.0 {
			req.Status = "in_progress"
		} else {
			req.Status = "not_started"
		}

		// Check if at risk (low progress and approaching due date)
		if req.NextDueDate != nil && req.NextDueDate.Before(time.Now().AddDate(0, 0, 30)) && req.Progress < 0.5 {
			req.Status = "at_risk"
			atRiskCount++
		}

		totalProgress += req.Progress
	}

	// Calculate overall metrics
	tracking.OverallProgress = totalProgress / float64(len(tracking.Requirements))

	// Determine compliance level
	if tracking.OverallProgress >= 0.9 {
		tracking.ComplianceLevel = "compliant"
	} else if tracking.OverallProgress >= 0.5 {
		tracking.ComplianceLevel = "partial"
	} else {
		tracking.ComplianceLevel = "non_compliant"
	}

	// Determine risk level based on overall progress
	switch {
	case tracking.OverallProgress >= 0.8:
		tracking.RiskLevel = "low"
	case tracking.OverallProgress >= 0.5:
		tracking.RiskLevel = "medium"
	case tracking.OverallProgress >= 0.2:
		tracking.RiskLevel = "high"
	default:
		tracking.RiskLevel = "critical"
	}

	// Calculate velocity (progress change per day)
	// For testing purposes, we'll use a simple calculation
	// In a real implementation, this would track progress over time
	if tracking.OverallProgress > 0.5 {
		tracking.Velocity = 0.1 // Positive velocity for good progress
	} else if tracking.OverallProgress < 0.3 {
		tracking.Velocity = -0.1 // Negative velocity for poor progress
	} else {
		tracking.Velocity = 0.0 // Stable velocity
	}

	// Determine trend (simplified)
	if tracking.Velocity > 0.01 {
		tracking.Trend = "improving"
	} else if tracking.Velocity < -0.01 {
		tracking.Trend = "declining"
	} else {
		tracking.Trend = "stable"
	}

	tracking.LastUpdated = time.Now()
}

// generateProgressMetrics generates progress metrics for a specific period
func (cts *ComplianceTrackingService) generateProgressMetrics(businessID, frameworkID, period string) *ComplianceProgressMetrics {
	now := time.Now()

	// Parse period to get date range
	var startDate, endDate time.Time
	switch period {
	case "daily":
		startDate = now.Truncate(24 * time.Hour)
		endDate = startDate.Add(24 * time.Hour)
	case "weekly":
		startDate = now.Truncate(24*time.Hour).AddDate(0, 0, -int(now.Weekday()))
		endDate = startDate.AddDate(0, 0, 7)
	case "monthly":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0)
	case "quarterly":
		quarter := (int(now.Month()) - 1) / 3
		startDate = time.Date(now.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 3, 0)
	default:
		// Assume it's a date string
		if parsedDate, err := time.Parse("2006-01-02", period); err == nil {
			startDate = parsedDate
			endDate = parsedDate.Add(24 * time.Hour)
		} else {
			startDate = now.Truncate(24 * time.Hour)
			endDate = startDate.Add(24 * time.Hour)
		}
	}

	// Get tracking data
	trackingKey := fmt.Sprintf("%s:%s", businessID, frameworkID)
	tracking, exists := cts.trackingData[trackingKey]

	metrics := &ComplianceProgressMetrics{
		BusinessID:            businessID,
		FrameworkID:           frameworkID,
		Period:                period,
		StartDate:             startDate,
		EndDate:               endDate,
		OverallProgress:       0.0,
		RequirementsCompleted: 0,
		RequirementsTotal:     0,
		MilestonesCompleted:   0,
		MilestonesTotal:       0,
		AverageVelocity:       0.0,
		RiskTrend:             "stable",
		ComplianceTrend:       "stable",
		KeyMetrics:            make(map[string]float64),
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if exists {
		metrics.OverallProgress = tracking.OverallProgress
		metrics.RequirementsTotal = len(tracking.Requirements)
		metrics.AverageVelocity = tracking.Velocity
		metrics.RiskTrend = tracking.Trend
		metrics.ComplianceTrend = tracking.Trend

		// Count completed requirements
		for _, req := range tracking.Requirements {
			if req.Status == "completed" {
				metrics.RequirementsCompleted++
			}
		}

		// Count milestones
		for _, milestone := range cts.milestones {
			if milestone.BusinessID == businessID && milestone.FrameworkID == frameworkID {
				metrics.MilestonesTotal++
				if milestone.Status == "completed" {
					metrics.MilestonesCompleted++
				}
			}
		}

		// Set key metrics
		metrics.KeyMetrics["compliance_score"] = tracking.OverallProgress
		metrics.KeyMetrics["risk_score"] = 1.0 - tracking.OverallProgress
		metrics.KeyMetrics["velocity"] = tracking.Velocity
		metrics.KeyMetrics["completion_rate"] = float64(metrics.RequirementsCompleted) / float64(metrics.RequirementsTotal)
	}

	return metrics
}
