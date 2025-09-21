package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// UnifiedComplianceService provides comprehensive compliance tracking and management
// This service replaces the separate compliance_checks and compliance_records functionality
// with a unified approach using the compliance_tracking table
type UnifiedComplianceService struct {
	logger     *observability.Logger
	repository UnifiedComplianceRepository
	audit      AuditServiceInterface
}

// UnifiedComplianceRepository defines the interface for unified compliance data persistence
type UnifiedComplianceRepository interface {
	// SaveComplianceTracking saves a compliance tracking record
	SaveComplianceTracking(ctx context.Context, tracking *ComplianceTracking) error

	// GetComplianceTracking retrieves compliance tracking records with filtering
	GetComplianceTracking(ctx context.Context, filters *ComplianceTrackingFilters) ([]*ComplianceTracking, error)

	// GetComplianceTrackingByID retrieves a specific compliance tracking record by ID
	GetComplianceTrackingByID(ctx context.Context, id string) (*ComplianceTracking, error)

	// UpdateComplianceTracking updates an existing compliance tracking record
	UpdateComplianceTracking(ctx context.Context, tracking *ComplianceTracking) error

	// DeleteComplianceTracking deletes a compliance tracking record
	DeleteComplianceTracking(ctx context.Context, id string) error

	// GetMerchantComplianceSummary retrieves compliance summary for a merchant
	GetMerchantComplianceSummary(ctx context.Context, merchantID string) (*MerchantComplianceSummary, error)

	// GetComplianceAlerts retrieves compliance alerts for monitoring
	GetComplianceAlerts(ctx context.Context, filters *ComplianceAlertFilters) ([]*UnifiedComplianceAlert, error)

	// GetComplianceTrends retrieves compliance trends for reporting
	GetComplianceTrends(ctx context.Context, filters *ComplianceTrendFilters) ([]*UnifiedComplianceTrend, error)
}

// ComplianceTracking represents a unified compliance tracking record
type ComplianceTracking struct {
	ID                  string                 `json:"id" db:"id"`
	MerchantID          string                 `json:"merchant_id" db:"merchant_id"`
	ComplianceType      string                 `json:"compliance_type" db:"compliance_type"`
	ComplianceFramework string                 `json:"compliance_framework" db:"compliance_framework"`
	CheckType           string                 `json:"check_type" db:"check_type"`
	Status              ComplianceStatusType   `json:"status" db:"status"`
	Score               *float64               `json:"score" db:"score"`
	RiskLevel           string                 `json:"risk_level" db:"risk_level"`
	Requirements        map[string]interface{} `json:"requirements" db:"requirements"`
	CheckMethod         string                 `json:"check_method" db:"check_method"`
	Source              string                 `json:"source" db:"source"`
	RawData             map[string]interface{} `json:"raw_data" db:"raw_data"`
	Result              map[string]interface{} `json:"result" db:"result"`
	Findings            map[string]interface{} `json:"findings" db:"findings"`
	Recommendations     map[string]interface{} `json:"recommendations" db:"recommendations"`
	Evidence            map[string]interface{} `json:"evidence" db:"evidence"`
	CheckedBy           *string                `json:"checked_by" db:"checked_by"`
	CheckedAt           time.Time              `json:"checked_at" db:"checked_at"`
	ReviewedBy          *string                `json:"reviewed_by" db:"reviewed_by"`
	ReviewedAt          *time.Time             `json:"reviewed_at" db:"reviewed_at"`
	ApprovedBy          *string                `json:"approved_by" db:"approved_by"`
	ApprovedAt          *time.Time             `json:"approved_at" db:"approved_at"`
	DueDate             *time.Time             `json:"due_date" db:"due_date"`
	ExpiresAt           *time.Time             `json:"expires_at" db:"expires_at"`
	NextReviewDate      *time.Time             `json:"next_review_date" db:"next_review_date"`
	Priority            string                 `json:"priority" db:"priority"`
	AssignedTo          *string                `json:"assigned_to" db:"assigned_to"`
	Tags                []string               `json:"tags" db:"tags"`
	Notes               *string                `json:"notes" db:"notes"`
	Metadata            map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
}

// ComplianceTrackingFilters represents filters for compliance tracking queries
type ComplianceTrackingFilters struct {
	MerchantID          string               `json:"merchant_id,omitempty"`
	ComplianceType      string               `json:"compliance_type,omitempty"`
	ComplianceFramework string               `json:"compliance_framework,omitempty"`
	CheckType           string               `json:"check_type,omitempty"`
	Status              ComplianceStatusType `json:"status,omitempty"`
	RiskLevel           string               `json:"risk_level,omitempty"`
	Priority            string               `json:"priority,omitempty"`
	CheckedBy           string               `json:"checked_by,omitempty"`
	AssignedTo          string               `json:"assigned_to,omitempty"`
	DueDateAfter        *time.Time           `json:"due_date_after,omitempty"`
	DueDateBefore       *time.Time           `json:"due_date_before,omitempty"`
	ExpiresAtAfter      *time.Time           `json:"expires_at_after,omitempty"`
	ExpiresAtBefore     *time.Time           `json:"expires_at_before,omitempty"`
	CreatedAtAfter      *time.Time           `json:"created_at_after,omitempty"`
	CreatedAtBefore     *time.Time           `json:"created_at_before,omitempty"`
	Tags                []string             `json:"tags,omitempty"`
	Overdue             bool                 `json:"overdue,omitempty"`
	ExpiringSoon        bool                 `json:"expiring_soon,omitempty"`
	Limit               int                  `json:"limit,omitempty"`
	Offset              int                  `json:"offset,omitempty"`
}

// MerchantComplianceSummary represents a comprehensive compliance summary for a merchant
type MerchantComplianceSummary struct {
	MerchantID             string                    `json:"merchant_id"`
	TotalChecks            int                       `json:"total_checks"`
	CompletedChecks        int                       `json:"completed_checks"`
	PendingChecks          int                       `json:"pending_checks"`
	FailedChecks           int                       `json:"failed_checks"`
	OverdueChecks          int                       `json:"overdue_checks"`
	PastDueChecks          int                       `json:"past_due_checks"`
	AverageScore           *float64                  `json:"average_score"`
	LastCheckDate          *time.Time                `json:"last_check_date"`
	NextReviewDate         *time.Time                `json:"next_review_date"`
	ComplianceTypesCovered int                       `json:"compliance_types_covered"`
	RiskLevel              string                    `json:"risk_level"`
	ComplianceScore        float64                   `json:"compliance_score"`
	Trends                 []*UnifiedComplianceTrend `json:"trends"`
	Alerts                 []*UnifiedComplianceAlert `json:"alerts"`
	GeneratedAt            time.Time                 `json:"generated_at"`
}

// UnifiedComplianceAlert represents a compliance alert for monitoring
type UnifiedComplianceAlert struct {
	ID                  string     `json:"id"`
	MerchantID          string     `json:"merchant_id"`
	ComplianceType      string     `json:"compliance_type"`
	ComplianceFramework string     `json:"compliance_framework"`
	Status              string     `json:"status"`
	Priority            string     `json:"priority"`
	RiskLevel           string     `json:"risk_level"`
	AlertType           string     `json:"alert_type"`
	DueDate             *time.Time `json:"due_date"`
	ExpiresAt           *time.Time `json:"expires_at"`
	AssignedTo          *string    `json:"assigned_to"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// ComplianceAlertFilters represents filters for compliance alert queries
type ComplianceAlertFilters struct {
	MerchantID      string     `json:"merchant_id,omitempty"`
	ComplianceType  string     `json:"compliance_type,omitempty"`
	AlertType       string     `json:"alert_type,omitempty"`
	Priority        string     `json:"priority,omitempty"`
	RiskLevel       string     `json:"risk_level,omitempty"`
	AssignedTo      string     `json:"assigned_to,omitempty"`
	CreatedAtAfter  *time.Time `json:"created_at_after,omitempty"`
	CreatedAtBefore *time.Time `json:"created_at_before,omitempty"`
	Limit           int        `json:"limit,omitempty"`
	Offset          int        `json:"offset,omitempty"`
}

// UnifiedComplianceTrend represents compliance trends for reporting
type UnifiedComplianceTrend struct {
	Date            time.Time `json:"date"`
	MerchantID      string    `json:"merchant_id"`
	ComplianceType  string    `json:"compliance_type"`
	TotalChecks     int       `json:"total_checks"`
	CompletedChecks int       `json:"completed_checks"`
	FailedChecks    int       `json:"failed_checks"`
	AverageScore    *float64  `json:"average_score"`
	ComplianceScore float64   `json:"compliance_score"`
}

// ComplianceTrendFilters represents filters for compliance trend queries
type ComplianceTrendFilters struct {
	MerchantID     string    `json:"merchant_id,omitempty"`
	ComplianceType string    `json:"compliance_type,omitempty"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	GroupBy        string    `json:"group_by,omitempty"` // day, week, month
	Limit          int       `json:"limit,omitempty"`
	Offset         int       `json:"offset,omitempty"`
}

// NewUnifiedComplianceService creates a new unified compliance service
func NewUnifiedComplianceService(logger *observability.Logger, repository UnifiedComplianceRepository, audit AuditServiceInterface) *UnifiedComplianceService {
	return &UnifiedComplianceService{
		logger:     logger,
		repository: repository,
		audit:      audit,
	}
}

// CreateComplianceTracking creates a new compliance tracking record
func (ucs *UnifiedComplianceService) CreateComplianceTracking(ctx context.Context, req *CreateComplianceTrackingRequest) (*ComplianceTracking, error) {
	// Validate request
	if err := ucs.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create compliance tracking record
	tracking := &ComplianceTracking{
		ID:                  uuid.New().String(),
		MerchantID:          req.MerchantID,
		ComplianceType:      req.ComplianceType,
		ComplianceFramework: req.ComplianceFramework,
		CheckType:           req.CheckType,
		Status:              req.Status,
		Score:               req.Score,
		RiskLevel:           req.RiskLevel,
		Requirements:        req.Requirements,
		CheckMethod:         req.CheckMethod,
		Source:              req.Source,
		RawData:             req.RawData,
		Result:              req.Result,
		Findings:            req.Findings,
		Recommendations:     req.Recommendations,
		Evidence:            req.Evidence,
		CheckedBy:           req.CheckedBy,
		CheckedAt:           time.Now(),
		DueDate:             req.DueDate,
		ExpiresAt:           req.ExpiresAt,
		NextReviewDate:      req.NextReviewDate,
		Priority:            req.Priority,
		AssignedTo:          req.AssignedTo,
		Tags:                req.Tags,
		Notes:               req.Notes,
		Metadata:            req.Metadata,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Save to repository
	if err := ucs.repository.SaveComplianceTracking(ctx, tracking); err != nil {
		ucs.logger.Error("failed to save compliance tracking", map[string]interface{}{
			"error":           err.Error(),
			"merchant_id":     req.MerchantID,
			"compliance_type": req.ComplianceType,
		})
		return nil, fmt.Errorf("failed to save compliance tracking: %w", err)
	}

	// Log the creation
	var userID string
	if req.CheckedBy != nil {
		userID = *req.CheckedBy
	}
	if err := ucs.audit.LogMerchantOperation(ctx, &LogMerchantOperationRequest{
		UserID:       userID,
		MerchantID:   req.MerchantID,
		Action:       "create",
		ResourceType: "compliance_tracking",
		ResourceID:   tracking.ID,
		Details:      fmt.Sprintf("Created compliance tracking: %s", req.ComplianceType),
		Description:  fmt.Sprintf("Compliance tracking created for merchant %s", req.MerchantID),
		Metadata: map[string]interface{}{
			"compliance_type":      req.ComplianceType,
			"compliance_framework": req.ComplianceFramework,
			"check_type":           req.CheckType,
			"status":               req.Status,
			"priority":             req.Priority,
		},
	}); err != nil {
		ucs.logger.Warn("failed to log compliance tracking creation", map[string]interface{}{
			"error":       err.Error(),
			"tracking_id": tracking.ID,
		})
	}

	ucs.logger.Info("compliance tracking created", map[string]interface{}{
		"tracking_id":          tracking.ID,
		"merchant_id":          req.MerchantID,
		"compliance_type":      req.ComplianceType,
		"compliance_framework": req.ComplianceFramework,
		"status":               req.Status,
	})

	return tracking, nil
}

// CreateComplianceTrackingRequest represents a request to create compliance tracking
type CreateComplianceTrackingRequest struct {
	MerchantID          string                 `json:"merchant_id" validate:"required"`
	ComplianceType      string                 `json:"compliance_type" validate:"required"`
	ComplianceFramework string                 `json:"compliance_framework,omitempty"`
	CheckType           string                 `json:"check_type" validate:"required"`
	Status              ComplianceStatusType   `json:"status" validate:"required"`
	Score               *float64               `json:"score,omitempty"`
	RiskLevel           string                 `json:"risk_level,omitempty"`
	Requirements        map[string]interface{} `json:"requirements,omitempty"`
	CheckMethod         string                 `json:"check_method" validate:"required"`
	Source              string                 `json:"source" validate:"required"`
	RawData             map[string]interface{} `json:"raw_data,omitempty"`
	Result              map[string]interface{} `json:"result,omitempty"`
	Findings            map[string]interface{} `json:"findings,omitempty"`
	Recommendations     map[string]interface{} `json:"recommendations,omitempty"`
	Evidence            map[string]interface{} `json:"evidence,omitempty"`
	CheckedBy           *string                `json:"checked_by,omitempty"`
	DueDate             *time.Time             `json:"due_date,omitempty"`
	ExpiresAt           *time.Time             `json:"expires_at,omitempty"`
	NextReviewDate      *time.Time             `json:"next_review_date,omitempty"`
	Priority            string                 `json:"priority,omitempty"`
	AssignedTo          *string                `json:"assigned_to,omitempty"`
	Tags                []string               `json:"tags,omitempty"`
	Notes               *string                `json:"notes,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// GetComplianceTracking retrieves compliance tracking records
func (ucs *UnifiedComplianceService) GetComplianceTracking(ctx context.Context, filters *ComplianceTrackingFilters) ([]*ComplianceTracking, error) {
	records, err := ucs.repository.GetComplianceTracking(ctx, filters)
	if err != nil {
		ucs.logger.Error("failed to get compliance tracking", map[string]interface{}{
			"error":   err.Error(),
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to get compliance tracking: %w", err)
	}

	ucs.logger.Info("compliance tracking retrieved", map[string]interface{}{
		"count":   len(records),
		"filters": filters,
	})

	return records, nil
}

// GetMerchantComplianceSummary retrieves compliance summary for a merchant
func (ucs *UnifiedComplianceService) GetMerchantComplianceSummary(ctx context.Context, merchantID string) (*MerchantComplianceSummary, error) {
	if merchantID == "" {
		return nil, fmt.Errorf("merchant ID is required")
	}

	summary, err := ucs.repository.GetMerchantComplianceSummary(ctx, merchantID)
	if err != nil {
		ucs.logger.Error("failed to get merchant compliance summary", map[string]interface{}{
			"error":       err.Error(),
			"merchant_id": merchantID,
		})
		return nil, fmt.Errorf("failed to get merchant compliance summary: %w", err)
	}

	// Get compliance alerts
	alerts, err := ucs.repository.GetComplianceAlerts(ctx, &ComplianceAlertFilters{
		MerchantID: merchantID,
		Limit:      50,
	})
	if err != nil {
		ucs.logger.Warn("failed to get compliance alerts", map[string]interface{}{
			"error":       err.Error(),
			"merchant_id": merchantID,
		})
	} else {
		summary.Alerts = alerts
	}

	// Get compliance trends
	trends, err := ucs.repository.GetComplianceTrends(ctx, &ComplianceTrendFilters{
		MerchantID: merchantID,
		StartDate:  time.Now().AddDate(0, -3, 0), // Last 3 months
		EndDate:    time.Now(),
		GroupBy:    "week",
		Limit:      12,
	})
	if err != nil {
		ucs.logger.Warn("failed to get compliance trends", map[string]interface{}{
			"error":       err.Error(),
			"merchant_id": merchantID,
		})
	} else {
		summary.Trends = trends
	}

	ucs.logger.Info("merchant compliance summary retrieved", map[string]interface{}{
		"merchant_id":      merchantID,
		"total_checks":     summary.TotalChecks,
		"compliance_score": summary.ComplianceScore,
		"alerts_count":     len(summary.Alerts),
		"trends_count":     len(summary.Trends),
	})

	return summary, nil
}

// UpdateComplianceTracking updates an existing compliance tracking record
func (ucs *UnifiedComplianceService) UpdateComplianceTracking(ctx context.Context, req *UpdateComplianceTrackingRequest) (*ComplianceTracking, error) {
	// Get existing record
	existing, err := ucs.repository.GetComplianceTrackingByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing compliance tracking: %w", err)
	}

	// Update fields
	existing.Status = req.Status
	if req.Score != nil {
		existing.Score = req.Score
	}
	if req.RiskLevel != "" {
		existing.RiskLevel = req.RiskLevel
	}
	if req.Result != nil {
		existing.Result = req.Result
	}
	if req.Findings != nil {
		existing.Findings = req.Findings
	}
	if req.Recommendations != nil {
		existing.Recommendations = req.Recommendations
	}
	if req.Evidence != nil {
		existing.Evidence = req.Evidence
	}
	if req.ReviewedBy != nil {
		existing.ReviewedBy = req.ReviewedBy
		now := time.Now()
		existing.ReviewedAt = &now
	}
	if req.ApprovedBy != nil {
		existing.ApprovedBy = req.ApprovedBy
		now := time.Now()
		existing.ApprovedAt = &now
	}
	if req.DueDate != nil {
		existing.DueDate = req.DueDate
	}
	if req.ExpiresAt != nil {
		existing.ExpiresAt = req.ExpiresAt
	}
	if req.NextReviewDate != nil {
		existing.NextReviewDate = req.NextReviewDate
	}
	if req.Priority != "" {
		existing.Priority = req.Priority
	}
	if req.AssignedTo != nil {
		existing.AssignedTo = req.AssignedTo
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}
	if req.Metadata != nil {
		existing.Metadata = req.Metadata
	}
	existing.UpdatedAt = time.Now()

	// Save updated record
	if err := ucs.repository.UpdateComplianceTracking(ctx, existing); err != nil {
		ucs.logger.Error("failed to update compliance tracking", map[string]interface{}{
			"error":       err.Error(),
			"tracking_id": req.ID,
		})
		return nil, fmt.Errorf("failed to update compliance tracking: %w", err)
	}

	// Log the update
	if err := ucs.audit.LogMerchantOperation(ctx, &LogMerchantOperationRequest{
		UserID:       req.UpdatedBy,
		MerchantID:   existing.MerchantID,
		Action:       "update",
		ResourceType: "compliance_tracking",
		ResourceID:   existing.ID,
		Details:      fmt.Sprintf("Updated compliance tracking: %s", existing.ComplianceType),
		Description:  fmt.Sprintf("Compliance tracking updated for merchant %s", existing.MerchantID),
		Metadata: map[string]interface{}{
			"compliance_type": existing.ComplianceType,
			"status":          existing.Status,
			"priority":        existing.Priority,
		},
	}); err != nil {
		ucs.logger.Warn("failed to log compliance tracking update", map[string]interface{}{
			"error":       err.Error(),
			"tracking_id": existing.ID,
		})
	}

	ucs.logger.Info("compliance tracking updated", map[string]interface{}{
		"tracking_id":     existing.ID,
		"merchant_id":     existing.MerchantID,
		"compliance_type": existing.ComplianceType,
		"status":          existing.Status,
	})

	return existing, nil
}

// UpdateComplianceTrackingRequest represents a request to update compliance tracking
type UpdateComplianceTrackingRequest struct {
	ID              string                 `json:"id" validate:"required"`
	Status          ComplianceStatusType   `json:"status,omitempty"`
	Score           *float64               `json:"score,omitempty"`
	RiskLevel       string                 `json:"risk_level,omitempty"`
	Result          map[string]interface{} `json:"result,omitempty"`
	Findings        map[string]interface{} `json:"findings,omitempty"`
	Recommendations map[string]interface{} `json:"recommendations,omitempty"`
	Evidence        map[string]interface{} `json:"evidence,omitempty"`
	ReviewedBy      *string                `json:"reviewed_by,omitempty"`
	ApprovedBy      *string                `json:"approved_by,omitempty"`
	DueDate         *time.Time             `json:"due_date,omitempty"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	NextReviewDate  *time.Time             `json:"next_review_date,omitempty"`
	Priority        string                 `json:"priority,omitempty"`
	AssignedTo      *string                `json:"assigned_to,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	Notes           *string                `json:"notes,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	UpdatedBy       string                 `json:"updated_by" validate:"required"`
}

// GetComplianceAlerts retrieves compliance alerts
func (ucs *UnifiedComplianceService) GetComplianceAlerts(ctx context.Context, filters *ComplianceAlertFilters) ([]*UnifiedComplianceAlert, error) {
	alerts, err := ucs.repository.GetComplianceAlerts(ctx, filters)
	if err != nil {
		ucs.logger.Error("failed to get compliance alerts", map[string]interface{}{
			"error":   err.Error(),
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to get compliance alerts: %w", err)
	}

	ucs.logger.Info("compliance alerts retrieved", map[string]interface{}{
		"count":   len(alerts),
		"filters": filters,
	})

	return alerts, nil
}

// validateCreateRequest validates a create compliance tracking request
func (ucs *UnifiedComplianceService) validateCreateRequest(req *CreateComplianceTrackingRequest) error {
	if req.MerchantID == "" {
		return fmt.Errorf("merchant ID is required")
	}

	if req.ComplianceType == "" {
		return fmt.Errorf("compliance type is required")
	}

	if req.CheckType == "" {
		return fmt.Errorf("check type is required")
	}

	if req.CheckMethod == "" {
		return fmt.Errorf("check method is required")
	}

	if req.Source == "" {
		return fmt.Errorf("source is required")
	}

	if !req.Status.IsValid() {
		return fmt.Errorf("invalid compliance status: %s", req.Status)
	}

	// Validate score range if provided
	if req.Score != nil && (*req.Score < 0.0 || *req.Score > 1.0) {
		return fmt.Errorf("score must be between 0.0 and 1.0")
	}

	// Validate risk level if provided
	if req.RiskLevel != "" {
		validRiskLevels := []string{"low", "medium", "high", "critical"}
		valid := false
		for _, level := range validRiskLevels {
			if req.RiskLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid risk level: %s", req.RiskLevel)
		}
	}

	// Validate priority if provided
	if req.Priority != "" {
		validPriorities := []string{"low", "medium", "high", "critical"}
		valid := false
		for _, priority := range validPriorities {
			if req.Priority == priority {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid priority: %s", req.Priority)
		}
	}

	return nil
}
