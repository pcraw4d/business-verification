package compliance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ComplianceStatusSystem provides comprehensive compliance status tracking
type ComplianceStatusSystem struct {
	logger        *observability.Logger
	mu            sync.RWMutex
	statusData    map[string]*BusinessComplianceStatus // businessID -> status
	statusHistory map[string][]StatusChange            // businessID -> history
	statusAlerts  map[string][]StatusAlert             // businessID -> alerts
	statusReports map[string]*StatusReport             // businessID -> report
	statusMetrics map[string]*StatusMetrics            // businessID -> metrics
}

// BusinessComplianceStatus represents the overall compliance status for a business
type BusinessComplianceStatus struct {
	ID                  string                       `json:"id"`
	BusinessID          string                       `json:"business_id"`
	OverallStatus       ComplianceStatus             `json:"overall_status"`
	OverallScore        float64                      `json:"overall_score"` // 0.0 to 100.0
	FrameworkStatuses   map[string]FrameworkStatus   `json:"framework_statuses"`
	RequirementStatuses map[string]RequirementStatus `json:"requirement_statuses"`
	ControlStatuses     map[string]ControlStatus     `json:"control_statuses"`
	LastAssessment      time.Time                    `json:"last_assessment"`
	NextAssessment      time.Time                    `json:"next_assessment"`
	AssessmentFrequency string                       `json:"assessment_frequency"`
	ComplianceOfficer   string                       `json:"compliance_officer"`
	RiskLevel           ComplianceRiskLevel          `json:"risk_level"`
	Trend               string                       `json:"trend"` // "improving", "stable", "declining"
	TrendStrength       float64                      `json:"trend_strength"`
	LastUpdated         time.Time                    `json:"last_updated"`
	Metadata            map[string]interface{}       `json:"metadata,omitempty"`
}

// FrameworkStatus represents the status of a specific framework
type FrameworkStatus struct {
	FrameworkID       string              `json:"framework_id"`
	FrameworkName     string              `json:"framework_name"`
	Status            ComplianceStatus    `json:"status"`
	Score             float64             `json:"score"`
	RequirementCount  int                 `json:"requirement_count"`
	ImplementedCount  int                 `json:"implemented_count"`
	VerifiedCount     int                 `json:"verified_count"`
	NonCompliantCount int                 `json:"non_compliant_count"`
	ExemptCount       int                 `json:"exempt_count"`
	LastAssessment    time.Time           `json:"last_assessment"`
	NextAssessment    time.Time           `json:"next_assessment"`
	RiskLevel         ComplianceRiskLevel `json:"risk_level"`
	Trend             string              `json:"trend"`
	TrendStrength     float64             `json:"trend_strength"`
	LastUpdated       time.Time           `json:"last_updated"`
}

// RequirementStatus represents the status of a specific requirement
type RequirementStatus struct {
	RequirementID        string               `json:"requirement_id"`
	FrameworkID          string               `json:"framework_id"`
	Title                string               `json:"title"`
	Status               ComplianceStatus     `json:"status"`
	ImplementationStatus ImplementationStatus `json:"implementation_status"`
	Score                float64              `json:"score"`
	RiskLevel            ComplianceRiskLevel  `json:"risk_level"`
	Priority             CompliancePriority   `json:"priority"`
	LastReviewed         time.Time            `json:"last_reviewed"`
	NextReview           time.Time            `json:"next_review"`
	Reviewer             string               `json:"reviewer"`
	EvidenceCount        int                  `json:"evidence_count"`
	ExceptionCount       int                  `json:"exception_count"`
	RemediationPlanCount int                  `json:"remediation_plan_count"`
	Trend                string               `json:"trend"`
	TrendStrength        float64              `json:"trend_strength"`
	LastUpdated          time.Time            `json:"last_updated"`
}

// ControlStatus represents the status of a specific control
type ControlStatus struct {
	ControlID            string               `json:"control_id"`
	RequirementID        string               `json:"requirement_id"`
	Title                string               `json:"title"`
	Status               ComplianceStatus     `json:"status"`
	ImplementationStatus ImplementationStatus `json:"implementation_status"`
	Effectiveness        ControlEffectiveness `json:"effectiveness"`
	Score                float64              `json:"score"`
	LastTested           *time.Time           `json:"last_tested,omitempty"`
	NextTestDate         *time.Time           `json:"next_test_date,omitempty"`
	TestResultCount      int                  `json:"test_result_count"`
	PassCount            int                  `json:"pass_count"`
	FailCount            int                  `json:"fail_count"`
	EvidenceCount        int                  `json:"evidence_count"`
	Trend                string               `json:"trend"`
	TrendStrength        float64              `json:"trend_strength"`
	LastUpdated          time.Time            `json:"last_updated"`
}

// StatusChange represents a change in compliance status
type StatusChange struct {
	ID           string           `json:"id"`
	BusinessID   string           `json:"business_id"`
	EntityType   string           `json:"entity_type"` // "overall", "framework", "requirement", "control"
	EntityID     string           `json:"entity_id"`
	OldStatus    ComplianceStatus `json:"old_status"`
	NewStatus    ComplianceStatus `json:"new_status"`
	OldScore     float64          `json:"old_score"`
	NewScore     float64          `json:"new_score"`
	ChangeReason string           `json:"change_reason"`
	ChangedBy    string           `json:"changed_by"`
	ChangedAt    time.Time        `json:"changed_at"`
	Impact       string           `json:"impact"` // "low", "medium", "high", "critical"
	Notes        string           `json:"notes"`
}

// StatusAlert represents a compliance status alert
type StatusAlert struct {
	ID             string      `json:"id"`
	BusinessID     string      `json:"business_id"`
	AlertType      string      `json:"alert_type"` // "status_change", "score_decline", "deadline_missed", "risk_increase"
	Severity       string      `json:"severity"`   // "low", "medium", "high", "critical"
	EntityType     string      `json:"entity_type"`
	EntityID       string      `json:"entity_id"`
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	CurrentValue   interface{} `json:"current_value"`
	ThresholdValue interface{} `json:"threshold_value"`
	TriggeredAt    time.Time   `json:"triggered_at"`
	AcknowledgedAt *time.Time  `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string      `json:"acknowledged_by,omitempty"`
	ResolvedAt     *time.Time  `json:"resolved_at,omitempty"`
	ResolvedBy     string      `json:"resolved_by,omitempty"`
	Status         string      `json:"status"` // "active", "acknowledged", "resolved"
	Notes          string      `json:"notes"`
}

// StatusReport represents a comprehensive compliance status report
type StatusReport struct {
	ID                  string                 `json:"id"`
	BusinessID          string                 `json:"business_id"`
	ReportType          string                 `json:"report_type"` // "summary", "detailed", "framework", "requirement"
	GeneratedAt         time.Time              `json:"generated_at"`
	GeneratedBy         string                 `json:"generated_by"`
	Period              string                 `json:"period"`
	OverallStatus       ComplianceStatus       `json:"overall_status"`
	OverallScore        float64                `json:"overall_score"`
	FrameworkCount      int                    `json:"framework_count"`
	RequirementCount    int                    `json:"requirement_count"`
	ControlCount        int                    `json:"control_count"`
	AlertCount          int                    `json:"alert_count"`
	ChangeCount         int                    `json:"change_count"`
	RiskLevel           ComplianceRiskLevel    `json:"risk_level"`
	Trend               string                 `json:"trend"`
	TrendStrength       float64                `json:"trend_strength"`
	FrameworkStatuses   []FrameworkStatus      `json:"framework_statuses"`
	RequirementStatuses []RequirementStatus    `json:"requirement_statuses"`
	ControlStatuses     []ControlStatus        `json:"control_statuses"`
	StatusChanges       []StatusChange         `json:"status_changes"`
	Alerts              []StatusAlert          `json:"alerts"`
	Recommendations     []StatusRecommendation `json:"recommendations"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// StatusMetrics represents compliance status metrics
type StatusMetrics struct {
	BusinessID         string             `json:"business_id"`
	OverallScore       float64            `json:"overall_score"`
	FrameworkScores    map[string]float64 `json:"framework_scores"`
	RequirementScores  map[string]float64 `json:"requirement_scores"`
	ControlScores      map[string]float64 `json:"control_scores"`
	StatusDistribution map[string]int     `json:"status_distribution"`
	RiskDistribution   map[string]int     `json:"risk_distribution"`
	TrendData          []TrendPoint       `json:"trend_data"`
	AlertMetrics       AlertMetrics       `json:"alert_metrics"`
	ChangeMetrics      ChangeMetrics      `json:"change_metrics"`
	LastCalculated     time.Time          `json:"last_calculated"`
}

// TrendPoint represents a point in compliance trend data
type TrendPoint struct {
	Date        time.Time           `json:"date"`
	Score       float64             `json:"score"`
	Status      ComplianceStatus    `json:"status"`
	RiskLevel   ComplianceRiskLevel `json:"risk_level"`
	ChangeCount int                 `json:"change_count"`
	AlertCount  int                 `json:"alert_count"`
}

// AlertMetrics represents alert-related metrics
type AlertMetrics struct {
	TotalAlerts           int           `json:"total_alerts"`
	ActiveAlerts          int           `json:"active_alerts"`
	ResolvedAlerts        int           `json:"resolved_alerts"`
	CriticalAlerts        int           `json:"critical_alerts"`
	HighAlerts            int           `json:"high_alerts"`
	MediumAlerts          int           `json:"medium_alerts"`
	LowAlerts             int           `json:"low_alerts"`
	AverageResolutionTime time.Duration `json:"average_resolution_time"`
}

// ChangeMetrics represents change-related metrics
type ChangeMetrics struct {
	TotalChanges           int     `json:"total_changes"`
	Improvements           int     `json:"improvements"`
	Declines               int     `json:"declines"`
	StatusChanges          int     `json:"status_changes"`
	ScoreChanges           int     `json:"score_changes"`
	AverageChangeMagnitude float64 `json:"average_change_magnitude"`
}

// StatusRecommendation represents a recommendation for compliance status improvement
type StatusRecommendation struct {
	ID                  string    `json:"id"`
	Type                string    `json:"type"`     // "improvement", "remediation", "monitoring", "assessment"
	Priority            string    `json:"priority"` // "low", "medium", "high", "critical"
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	Action              string    `json:"action"`
	Impact              string    `json:"impact"`
	Effort              string    `json:"effort"`
	Timeline            string    `json:"timeline"`
	EntityType          string    `json:"entity_type"`
	EntityID            string    `json:"entity_id"`
	ExpectedImprovement float64   `json:"expected_improvement"`
	AssignedTo          string    `json:"assigned_to"`
	Status              string    `json:"status"` // "open", "in_progress", "completed", "rejected"
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// NewComplianceStatusSystem creates a new compliance status tracking system
func NewComplianceStatusSystem(logger *observability.Logger) *ComplianceStatusSystem {
	return &ComplianceStatusSystem{
		logger:        logger,
		statusData:    make(map[string]*BusinessComplianceStatus),
		statusHistory: make(map[string][]StatusChange),
		statusAlerts:  make(map[string][]StatusAlert),
		statusReports: make(map[string]*StatusReport),
		statusMetrics: make(map[string]*StatusMetrics),
	}
}

// InitializeBusinessStatus initializes compliance status tracking for a business
func (s *ComplianceStatusSystem) InitializeBusinessStatus(ctx context.Context, businessID string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Initializing compliance status tracking for business",
		"request_id", requestID,
		"business_id", businessID,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	status := &BusinessComplianceStatus{
		ID:                  fmt.Sprintf("status_%s", businessID),
		BusinessID:          businessID,
		OverallStatus:       ComplianceStatusNotStarted,
		OverallScore:        0.0,
		FrameworkStatuses:   make(map[string]FrameworkStatus),
		RequirementStatuses: make(map[string]RequirementStatus),
		ControlStatuses:     make(map[string]ControlStatus),
		LastAssessment:      time.Now(),
		NextAssessment:      time.Now().Add(30 * 24 * time.Hour), // 30 days
		AssessmentFrequency: "monthly",
		ComplianceOfficer:   "system",
		RiskLevel:           ComplianceRiskLevelLow,
		Trend:               "stable",
		TrendStrength:       0.0,
		LastUpdated:         time.Now(),
	}

	s.statusData[businessID] = status
	s.statusHistory[businessID] = make([]StatusChange, 0)
	s.statusAlerts[businessID] = make([]StatusAlert, 0)

	s.logger.Info("Compliance status tracking initialized successfully",
		"request_id", requestID,
		"business_id", businessID,
	)

	return nil
}

// UpdateFrameworkStatus updates the status of a specific framework
func (s *ComplianceStatusSystem) UpdateFrameworkStatus(ctx context.Context, businessID, frameworkID string, status ComplianceStatus, score float64, riskLevel ComplianceRiskLevel, trend string, trendStrength float64) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating framework status",
		"request_id", requestID,
		"business_id", businessID,
		"framework_id", frameworkID,
		"status", status,
		"score", score,
		"risk_level", riskLevel,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	businessStatus, exists := s.statusData[businessID]
	if !exists {
		return fmt.Errorf("compliance status not found for business %s", businessID)
	}

	// Get existing framework status
	frameworkStatus, exists := businessStatus.FrameworkStatuses[frameworkID]
	oldStatus := frameworkStatus.Status
	oldScore := frameworkStatus.Score

	// Update framework status
	frameworkStatus.FrameworkID = frameworkID
	frameworkStatus.Status = status
	frameworkStatus.Score = score
	frameworkStatus.RiskLevel = riskLevel
	frameworkStatus.Trend = trend
	frameworkStatus.TrendStrength = trendStrength
	frameworkStatus.LastAssessment = time.Now()
	frameworkStatus.NextAssessment = time.Now().Add(30 * 24 * time.Hour)
	frameworkStatus.LastUpdated = time.Now()

	businessStatus.FrameworkStatuses[frameworkID] = frameworkStatus

	// Update overall status
	s.updateOverallStatus(businessStatus)

	// Record status change
	s.recordStatusChange(ctx, businessID, "framework", frameworkID, oldStatus, status, oldScore, score, "Framework status updated")

	s.logger.Info("Framework status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"framework_id", frameworkID,
		"old_status", oldStatus,
		"new_status", status,
		"old_score", oldScore,
		"new_score", score,
	)

	return nil
}

// UpdateRequirementStatus updates the status of a specific requirement
func (s *ComplianceStatusSystem) UpdateRequirementStatus(ctx context.Context, businessID, frameworkID, requirementID string, status ComplianceStatus, implementationStatus ImplementationStatus, score float64, riskLevel ComplianceRiskLevel, priority CompliancePriority, trend string, trendStrength float64) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating requirement status",
		"request_id", requestID,
		"business_id", businessID,
		"framework_id", frameworkID,
		"requirement_id", requirementID,
		"status", status,
		"score", score,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	businessStatus, exists := s.statusData[businessID]
	if !exists {
		return fmt.Errorf("compliance status not found for business %s", businessID)
	}

	// Get existing requirement status
	requirementStatus, exists := businessStatus.RequirementStatuses[requirementID]
	oldStatus := requirementStatus.Status
	oldScore := requirementStatus.Score

	// Update requirement status
	requirementStatus.RequirementID = requirementID
	requirementStatus.FrameworkID = frameworkID
	requirementStatus.Status = status
	requirementStatus.ImplementationStatus = implementationStatus
	requirementStatus.Score = score
	requirementStatus.RiskLevel = riskLevel
	requirementStatus.Priority = priority
	requirementStatus.Trend = trend
	requirementStatus.TrendStrength = trendStrength
	requirementStatus.LastReviewed = time.Now()
	requirementStatus.NextReview = time.Now().Add(30 * 24 * time.Hour)
	requirementStatus.LastUpdated = time.Now()

	businessStatus.RequirementStatuses[requirementID] = requirementStatus

	// Update framework status
	s.updateFrameworkStatusFromRequirements(businessStatus, frameworkID)

	// Record status change
	s.recordStatusChange(ctx, businessID, "requirement", requirementID, oldStatus, status, oldScore, score, "Requirement status updated")

	s.logger.Info("Requirement status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"requirement_id", requirementID,
		"old_status", oldStatus,
		"new_status", status,
		"old_score", oldScore,
		"new_score", score,
	)

	return nil
}

// UpdateControlStatus updates the status of a specific control
func (s *ComplianceStatusSystem) UpdateControlStatus(ctx context.Context, businessID, frameworkID, requirementID, controlID string, status ComplianceStatus, implementationStatus ImplementationStatus, effectiveness ControlEffectiveness, score float64, trend string, trendStrength float64) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating control status",
		"request_id", requestID,
		"business_id", businessID,
		"framework_id", frameworkID,
		"requirement_id", requirementID,
		"control_id", controlID,
		"status", status,
		"score", score,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	businessStatus, exists := s.statusData[businessID]
	if !exists {
		return fmt.Errorf("compliance status not found for business %s", businessID)
	}

	// Get existing control status
	controlStatus, exists := businessStatus.ControlStatuses[controlID]
	oldStatus := controlStatus.Status
	oldScore := controlStatus.Score

	// Update control status
	controlStatus.ControlID = controlID
	controlStatus.RequirementID = requirementID
	controlStatus.Status = status
	controlStatus.ImplementationStatus = implementationStatus
	controlStatus.Effectiveness = effectiveness
	controlStatus.Score = score
	controlStatus.Trend = trend
	controlStatus.TrendStrength = trendStrength
	controlStatus.LastUpdated = time.Now()

	businessStatus.ControlStatuses[controlID] = controlStatus

	// Update requirement status from controls
	s.updateRequirementStatusFromControls(businessStatus, requirementID)

	// Record status change
	s.recordStatusChange(ctx, businessID, "control", controlID, oldStatus, status, oldScore, score, "Control status updated")

	s.logger.Info("Control status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"control_id", controlID,
		"old_status", oldStatus,
		"new_status", status,
		"old_score", oldScore,
		"new_score", score,
	)

	return nil
}

// CreateStatusAlert creates a compliance status alert
func (s *ComplianceStatusSystem) CreateStatusAlert(ctx context.Context, businessID, alertType, severity, entityType, entityID, title, description string, currentValue, thresholdValue interface{}) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Creating compliance status alert",
		"request_id", requestID,
		"business_id", businessID,
		"alert_type", alertType,
		"severity", severity,
		"entity_type", entityType,
		"entity_id", entityID,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	alert := StatusAlert{
		ID:             fmt.Sprintf("alert_%s_%d", businessID, time.Now().UnixNano()),
		BusinessID:     businessID,
		AlertType:      alertType,
		Severity:       severity,
		EntityType:     entityType,
		EntityID:       entityID,
		Title:          title,
		Description:    description,
		CurrentValue:   currentValue,
		ThresholdValue: thresholdValue,
		TriggeredAt:    time.Now(),
		Status:         "active",
	}

	s.statusAlerts[businessID] = append(s.statusAlerts[businessID], alert)

	s.logger.Info("Compliance status alert created successfully",
		"request_id", requestID,
		"business_id", businessID,
		"alert_id", alert.ID,
		"alert_type", alertType,
		"severity", severity,
	)

	return nil
}

// AcknowledgeAlert acknowledges a compliance status alert
func (s *ComplianceStatusSystem) AcknowledgeAlert(ctx context.Context, businessID, alertID, acknowledgedBy string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Acknowledging compliance status alert",
		"request_id", requestID,
		"business_id", businessID,
		"alert_id", alertID,
		"acknowledged_by", acknowledgedBy,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	alerts, exists := s.statusAlerts[businessID]
	if !exists {
		return fmt.Errorf("no alerts found for business %s", businessID)
	}

	for i, alert := range alerts {
		if alert.ID == alertID {
			now := time.Now()
			s.statusAlerts[businessID][i].AcknowledgedAt = &now
			s.statusAlerts[businessID][i].AcknowledgedBy = acknowledgedBy
			s.statusAlerts[businessID][i].Status = "acknowledged"

			s.logger.Info("Compliance status alert acknowledged successfully",
				"request_id", requestID,
				"business_id", businessID,
				"alert_id", alertID,
			)

			return nil
		}
	}

	return fmt.Errorf("alert %s not found", alertID)
}

// ResolveAlert resolves a compliance status alert
func (s *ComplianceStatusSystem) ResolveAlert(ctx context.Context, businessID, alertID, resolvedBy string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Resolving compliance status alert",
		"request_id", requestID,
		"business_id", businessID,
		"alert_id", alertID,
		"resolved_by", resolvedBy,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	alerts, exists := s.statusAlerts[businessID]
	if !exists {
		return fmt.Errorf("no alerts found for business %s", businessID)
	}

	for i, alert := range alerts {
		if alert.ID == alertID {
			now := time.Now()
			s.statusAlerts[businessID][i].ResolvedAt = &now
			s.statusAlerts[businessID][i].ResolvedBy = resolvedBy
			s.statusAlerts[businessID][i].Status = "resolved"

			s.logger.Info("Compliance status alert resolved successfully",
				"request_id", requestID,
				"business_id", businessID,
				"alert_id", alertID,
			)

			return nil
		}
	}

	return fmt.Errorf("alert %s not found", alertID)
}

// GetComplianceStatus gets the compliance status for a business
func (s *ComplianceStatusSystem) GetComplianceStatus(ctx context.Context, businessID string) (*BusinessComplianceStatus, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting compliance status",
		"request_id", requestID,
		"business_id", businessID,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.statusData[businessID]
	if !exists {
		return nil, fmt.Errorf("compliance status not found for business %s", businessID)
	}

	return status, nil
}

// GetStatusHistory gets the status change history for a business
func (s *ComplianceStatusSystem) GetStatusHistory(ctx context.Context, businessID string, startDate, endDate time.Time) ([]StatusChange, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting status history",
		"request_id", requestID,
		"business_id", businessID,
		"start_date", startDate,
		"end_date", endDate,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	history, exists := s.statusHistory[businessID]
	if !exists {
		return nil, fmt.Errorf("status history not found for business %s", businessID)
	}

	var filteredHistory []StatusChange
	for _, change := range history {
		if change.ChangedAt.After(startDate) && change.ChangedAt.Before(endDate) {
			filteredHistory = append(filteredHistory, change)
		}
	}

	return filteredHistory, nil
}

// GetStatusAlerts gets the status alerts for a business
func (s *ComplianceStatusSystem) GetStatusAlerts(ctx context.Context, businessID string, status string) ([]StatusAlert, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting status alerts",
		"request_id", requestID,
		"business_id", businessID,
		"status", status,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts, exists := s.statusAlerts[businessID]
	if !exists {
		return nil, fmt.Errorf("status alerts not found for business %s", businessID)
	}

	if status == "" {
		return alerts, nil
	}

	var filteredAlerts []StatusAlert
	for _, alert := range alerts {
		if alert.Status == status {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	return filteredAlerts, nil
}

// GenerateStatusReport generates a comprehensive compliance status report
func (s *ComplianceStatusSystem) GenerateStatusReport(ctx context.Context, businessID string, reportType string) (*StatusReport, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Generating compliance status report",
		"request_id", requestID,
		"business_id", businessID,
		"report_type", reportType,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.statusData[businessID]
	if !exists {
		return nil, fmt.Errorf("compliance status not found for business %s", businessID)
	}

	alerts, _ := s.GetStatusAlerts(ctx, businessID, "")
	history, _ := s.GetStatusHistory(ctx, businessID, time.Now().Add(-30*24*time.Hour), time.Now())

	report := &StatusReport{
		ID:               fmt.Sprintf("report_%s_%d", businessID, time.Now().Unix()),
		BusinessID:       businessID,
		ReportType:       reportType,
		GeneratedAt:      time.Now(),
		GeneratedBy:      "system",
		Period:           "30_days",
		OverallStatus:    status.OverallStatus,
		OverallScore:     status.OverallScore,
		FrameworkCount:   len(status.FrameworkStatuses),
		RequirementCount: len(status.RequirementStatuses),
		ControlCount:     len(status.ControlStatuses),
		AlertCount:       len(alerts),
		ChangeCount:      len(history),
		RiskLevel:        status.RiskLevel,
		Trend:            status.Trend,
		TrendStrength:    status.TrendStrength,
		StatusChanges:    history,
		Alerts:           alerts,
		Recommendations:  s.generateStatusRecommendations(status, alerts, history),
	}

	// Add framework statuses
	for _, frameworkStatus := range status.FrameworkStatuses {
		report.FrameworkStatuses = append(report.FrameworkStatuses, frameworkStatus)
	}

	// Add requirement statuses
	for _, requirementStatus := range status.RequirementStatuses {
		report.RequirementStatuses = append(report.RequirementStatuses, requirementStatus)
	}

	// Add control statuses
	for _, controlStatus := range status.ControlStatuses {
		report.ControlStatuses = append(report.ControlStatuses, controlStatus)
	}

	s.statusReports[businessID] = report

	s.logger.Info("Compliance status report generated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"report_type", reportType,
		"framework_count", report.FrameworkCount,
		"requirement_count", report.RequirementCount,
		"control_count", report.ControlCount,
	)

	return report, nil
}

// Helper methods
func (s *ComplianceStatusSystem) updateOverallStatus(status *BusinessComplianceStatus) {
	if len(status.FrameworkStatuses) == 0 {
		status.OverallScore = 0.0
		status.OverallStatus = ComplianceStatusNotStarted
		return
	}

	totalScore := 0.0
	validFrameworks := 0

	for _, frameworkStatus := range status.FrameworkStatuses {
		totalScore += frameworkStatus.Score
		validFrameworks++
	}

	if validFrameworks > 0 {
		status.OverallScore = totalScore / float64(validFrameworks)
	} else {
		status.OverallScore = 0.0
	}

	// Update overall status based on score
	if status.OverallScore >= 90.0 {
		status.OverallStatus = ComplianceStatusVerified
	} else if status.OverallScore >= 70.0 {
		status.OverallStatus = ComplianceStatusImplemented
	} else if status.OverallScore >= 30.0 {
		status.OverallStatus = ComplianceStatusInProgress
	} else {
		status.OverallStatus = ComplianceStatusNotStarted
	}

	// Update risk level based on score
	if status.OverallScore >= 80.0 {
		status.RiskLevel = ComplianceRiskLevelLow
	} else if status.OverallScore >= 60.0 {
		status.RiskLevel = ComplianceRiskLevelMedium
	} else if status.OverallScore >= 40.0 {
		status.RiskLevel = ComplianceRiskLevelHigh
	} else {
		status.RiskLevel = ComplianceRiskLevelCritical
	}

	status.LastUpdated = time.Now()
}

func (s *ComplianceStatusSystem) updateFrameworkStatusFromRequirements(status *BusinessComplianceStatus, frameworkID string) {
	var frameworkRequirements []RequirementStatus
	for _, reqStatus := range status.RequirementStatuses {
		if reqStatus.FrameworkID == frameworkID {
			frameworkRequirements = append(frameworkRequirements, reqStatus)
		}
	}

	if len(frameworkRequirements) == 0 {
		return
	}

	frameworkStatus := status.FrameworkStatuses[frameworkID]
	totalScore := 0.0
	implementedCount := 0
	verifiedCount := 0
	nonCompliantCount := 0
	exemptCount := 0

	for _, reqStatus := range frameworkRequirements {
		totalScore += reqStatus.Score
		switch reqStatus.Status {
		case ComplianceStatusImplemented:
			implementedCount++
		case ComplianceStatusVerified:
			verifiedCount++
		case ComplianceStatusNonCompliant:
			nonCompliantCount++
		case ComplianceStatusExempt:
			exemptCount++
		}
	}

	frameworkStatus.Score = totalScore / float64(len(frameworkRequirements))
	frameworkStatus.RequirementCount = len(frameworkRequirements)
	frameworkStatus.ImplementedCount = implementedCount
	frameworkStatus.VerifiedCount = verifiedCount
	frameworkStatus.NonCompliantCount = nonCompliantCount
	frameworkStatus.ExemptCount = exemptCount

	// Update framework status based on score
	if frameworkStatus.Score >= 90.0 {
		frameworkStatus.Status = ComplianceStatusVerified
	} else if frameworkStatus.Score >= 70.0 {
		frameworkStatus.Status = ComplianceStatusImplemented
	} else if frameworkStatus.Score >= 30.0 {
		frameworkStatus.Status = ComplianceStatusInProgress
	} else {
		frameworkStatus.Status = ComplianceStatusNotStarted
	}

	status.FrameworkStatuses[frameworkID] = frameworkStatus
}

func (s *ComplianceStatusSystem) updateRequirementStatusFromControls(status *BusinessComplianceStatus, requirementID string) {
	var requirementControls []ControlStatus
	for _, controlStatus := range status.ControlStatuses {
		if controlStatus.RequirementID == requirementID {
			requirementControls = append(requirementControls, controlStatus)
		}
	}

	if len(requirementControls) == 0 {
		return
	}

	requirementStatus := status.RequirementStatuses[requirementID]
	totalScore := 0.0
	passCount := 0
	failCount := 0

	for _, controlStatus := range requirementControls {
		totalScore += controlStatus.Score
		if controlStatus.Effectiveness == ControlEffectivenessEffective || controlStatus.Effectiveness == ControlEffectivenessHighlyEffective {
			passCount++
		} else {
			failCount++
		}
	}

	requirementStatus.Score = totalScore / float64(len(requirementControls))

	// Update requirement status based on score
	if requirementStatus.Score >= 90.0 {
		requirementStatus.Status = ComplianceStatusVerified
	} else if requirementStatus.Score >= 70.0 {
		requirementStatus.Status = ComplianceStatusImplemented
	} else if requirementStatus.Score >= 30.0 {
		requirementStatus.Status = ComplianceStatusInProgress
	} else {
		requirementStatus.Status = ComplianceStatusNotStarted
	}

	status.RequirementStatuses[requirementID] = requirementStatus
}

func (s *ComplianceStatusSystem) recordStatusChange(ctx context.Context, businessID, entityType, entityID string, oldStatus, newStatus ComplianceStatus, oldScore, newScore float64, reason string) {
	change := StatusChange{
		ID:           fmt.Sprintf("change_%s_%d", businessID, time.Now().UnixNano()),
		BusinessID:   businessID,
		EntityType:   entityType,
		EntityID:     entityID,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		OldScore:     oldScore,
		NewScore:     newScore,
		ChangeReason: reason,
		ChangedBy:    "system",
		ChangedAt:    time.Now(),
		Impact:       s.calculateChangeImpact(oldScore, newScore),
	}

	s.statusHistory[businessID] = append(s.statusHistory[businessID], change)
}

func (s *ComplianceStatusSystem) calculateChangeImpact(oldScore, newScore float64) string {
	change := newScore - oldScore
	if change >= 10.0 {
		return "high"
	} else if change >= 5.0 {
		return "medium"
	} else if change >= 1.0 {
		return "low"
	}
	return "minimal"
}

func (s *ComplianceStatusSystem) generateStatusRecommendations(status *BusinessComplianceStatus, alerts []StatusAlert, history []StatusChange) []StatusRecommendation {
	var recommendations []StatusRecommendation

	// Low overall score recommendation
	if status.OverallScore < 70.0 {
		recommendations = append(recommendations, StatusRecommendation{
			ID:                  fmt.Sprintf("rec_%s_low_score", status.BusinessID),
			Type:                "improvement",
			Priority:            "high",
			Title:               "Low Overall Compliance Score",
			Description:         fmt.Sprintf("Overall compliance score is %.1f%%, below the target of 70%%", status.OverallScore),
			Action:              "Focus on implementing high-priority requirements and controls",
			Impact:              "High - Improves overall compliance posture",
			Effort:              "High - Requires significant implementation effort",
			Timeline:            "3-6 months",
			EntityType:          "overall",
			EntityID:            status.BusinessID,
			ExpectedImprovement: 20.0,
			Status:              "open",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		})
	}

	// High alert count recommendation
	if len(alerts) > 5 {
		recommendations = append(recommendations, StatusRecommendation{
			ID:                  fmt.Sprintf("rec_%s_high_alerts", status.BusinessID),
			Type:                "monitoring",
			Priority:            "medium",
			Title:               "High Number of Active Alerts",
			Description:         fmt.Sprintf("There are %d active compliance alerts that require attention", len(alerts)),
			Action:              "Review and address all active alerts promptly",
			Impact:              "Medium - Reduces compliance risk",
			Effort:              "Medium - Requires alert review and remediation",
			Timeline:            "1-2 weeks",
			EntityType:          "overall",
			EntityID:            status.BusinessID,
			ExpectedImprovement: 5.0,
			Status:              "open",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		})
	}

	// Declining trend recommendation
	if status.Trend == "declining" && status.TrendStrength > 0.5 {
		recommendations = append(recommendations, StatusRecommendation{
			ID:                  fmt.Sprintf("rec_%s_declining_trend", status.BusinessID),
			Type:                "remediation",
			Priority:            "high",
			Title:               "Declining Compliance Trend",
			Description:         "Compliance score is declining, indicating potential issues",
			Action:              "Investigate root causes and implement corrective actions",
			Impact:              "High - Prevents further compliance deterioration",
			Effort:              "High - Requires investigation and remediation",
			Timeline:            "2-4 weeks",
			EntityType:          "overall",
			EntityID:            status.BusinessID,
			ExpectedImprovement: 15.0,
			Status:              "open",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		})
	}

	return recommendations
}
