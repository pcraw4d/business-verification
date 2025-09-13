package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pcraw4d/business-verification/internal/models"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// ComplianceService provides comprehensive compliance checking and regulatory requirement validation
type ComplianceService struct {
	logger     *observability.Logger
	repository ComplianceRepository
	audit      AuditServiceInterface
}

// ComplianceRepository defines the interface for compliance data persistence
type ComplianceRepository interface {
	// SaveComplianceRequirement saves a compliance requirement
	SaveComplianceRequirement(ctx context.Context, requirement *ComplianceRequirement) error

	// GetComplianceRequirements retrieves compliance requirements with filtering
	GetComplianceRequirements(ctx context.Context, filters *ComplianceRequirementFilters) ([]*ComplianceRequirement, error)

	// GetComplianceRequirementByID retrieves a specific compliance requirement by ID
	GetComplianceRequirementByID(ctx context.Context, id string) (*ComplianceRequirement, error)

	// UpdateComplianceRequirement updates an existing compliance requirement
	UpdateComplianceRequirement(ctx context.Context, requirement *ComplianceRequirement) error

	// DeleteComplianceRequirement deletes a compliance requirement
	DeleteComplianceRequirement(ctx context.Context, id string) error

	// GetMerchantComplianceStatus retrieves compliance status for a merchant
	GetMerchantComplianceStatus(ctx context.Context, merchantID string) (*MerchantComplianceStatus, error)

	// SaveComplianceAssessment saves a compliance assessment
	SaveComplianceAssessment(ctx context.Context, assessment *ComplianceAssessment) error

	// GetComplianceAssessments retrieves compliance assessments with filtering
	GetComplianceAssessments(ctx context.Context, filters *ComplianceAssessmentFilters) ([]*ComplianceAssessment, error)

	// GetComplianceReport retrieves a compliance report
	GetComplianceReport(ctx context.Context, reportID string) (*ComplianceReport, error)

	// SaveComplianceReport saves a compliance report
	SaveComplianceReport(ctx context.Context, report *ComplianceReport) error
}

// AuditServiceInterface defines the interface for audit operations
type AuditServiceInterface interface {
	LogMerchantOperation(ctx context.Context, req *LogMerchantOperationRequest) error
	GetComplianceStatus(ctx context.Context, merchantID string) (*ComplianceStatus, error)
	GetComplianceRecords(ctx context.Context, filters *ComplianceFilters) ([]*ComplianceRecord, error)
}

// ComplianceRequirement represents a regulatory compliance requirement
type ComplianceRequirement struct {
	ID               string                 `json:"id" db:"id"`
	Regulation       string                 `json:"regulation" db:"regulation"`
	Requirement      string                 `json:"requirement" db:"requirement"`
	Description      string                 `json:"description" db:"description"`
	Category         ComplianceCategory     `json:"category" db:"category"`
	Priority         CompliancePriority     `json:"priority" db:"priority"`
	RiskLevel        models.RiskLevel       `json:"risk_level" db:"risk_level"`
	ApplicableTo     []string               `json:"applicable_to" db:"applicable_to"`
	EffectiveDate    time.Time              `json:"effective_date" db:"effective_date"`
	ExpiryDate       *time.Time             `json:"expiry_date" db:"expiry_date"`
	ReviewFrequency  string                 `json:"review_frequency" db:"review_frequency"`
	EvidenceRequired []string               `json:"evidence_required" db:"evidence_required"`
	ValidationRules  []string               `json:"validation_rules" db:"validation_rules"`
	Penalties        []string               `json:"penalties" db:"penalties"`
	Status           ComplianceStatusType   `json:"status" db:"status"`
	LastReviewDate   *time.Time             `json:"last_review_date" db:"last_review_date"`
	NextReviewDate   *time.Time             `json:"next_review_date" db:"next_review_date"`
	AssignedTo       string                 `json:"assigned_to" db:"assigned_to"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedBy        string                 `json:"created_by" db:"created_by"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// ComplianceCategory represents the category of compliance requirement
type ComplianceCategory string

const (
	ComplianceCategoryAML      ComplianceCategory = "aml"
	ComplianceCategoryKYC      ComplianceCategory = "kyc"
	ComplianceCategoryKYB      ComplianceCategory = "kyb"
	ComplianceCategoryFATF     ComplianceCategory = "fatf"
	ComplianceCategoryGDPR     ComplianceCategory = "gdpr"
	ComplianceCategorySOX      ComplianceCategory = "sox"
	ComplianceCategoryPCI      ComplianceCategory = "pci"
	ComplianceCategoryISO27001 ComplianceCategory = "iso27001"
	ComplianceCategorySOC2     ComplianceCategory = "soc2"
	ComplianceCategoryBSA      ComplianceCategory = "bsa"
	ComplianceCategoryOFAC     ComplianceCategory = "ofac"
	ComplianceCategoryCustom   ComplianceCategory = "custom"
)

// IsValid checks if the compliance category is valid
func (cc ComplianceCategory) IsValid() bool {
	switch cc {
	case ComplianceCategoryAML, ComplianceCategoryKYC, ComplianceCategoryKYB, ComplianceCategoryFATF,
		ComplianceCategoryGDPR, ComplianceCategorySOX, ComplianceCategoryPCI, ComplianceCategoryISO27001,
		ComplianceCategorySOC2, ComplianceCategoryBSA, ComplianceCategoryOFAC, ComplianceCategoryCustom:
		return true
	default:
		return false
	}
}

// String returns the string representation of the compliance category
func (cc ComplianceCategory) String() string {
	return string(cc)
}

// ComplianceRequirementFilters represents filters for compliance requirement queries
type ComplianceRequirementFilters struct {
	Regulation      string               `json:"regulation,omitempty"`
	Category        ComplianceCategory   `json:"category,omitempty"`
	Priority        CompliancePriority   `json:"priority,omitempty"`
	RiskLevel       models.RiskLevel     `json:"risk_level,omitempty"`
	Status          ComplianceStatusType `json:"status,omitempty"`
	AssignedTo      string               `json:"assigned_to,omitempty"`
	EffectiveAfter  *time.Time           `json:"effective_after,omitempty"`
	EffectiveBefore *time.Time           `json:"effective_before,omitempty"`
	Expired         bool                 `json:"expired,omitempty"`
	Limit           int                  `json:"limit,omitempty"`
	Offset          int                  `json:"offset,omitempty"`
}

// MerchantComplianceStatus represents the compliance status for a specific merchant
type MerchantComplianceStatus struct {
	MerchantID            string                    `json:"merchant_id"`
	OverallStatus         ComplianceStatusType      `json:"overall_status"`
	ComplianceScore       float64                   `json:"compliance_score"`
	TotalRequirements     int                       `json:"total_requirements"`
	CompletedRequirements int                       `json:"completed_requirements"`
	OverdueRequirements   int                       `json:"overdue_requirements"`
	FailedRequirements    int                       `json:"failed_requirements"`
	RiskLevel             models.RiskLevel          `json:"risk_level"`
	LastAssessmentDate    time.Time                 `json:"last_assessment_date"`
	NextAssessmentDate    time.Time                 `json:"next_assessment_date"`
	Requirements          []*MerchantComplianceItem `json:"requirements"`
	Trends                []*ComplianceTrend        `json:"trends"`
	Alerts                []*ComplianceAlert        `json:"alerts"`
	GeneratedAt           time.Time                 `json:"generated_at"`
}

// MerchantComplianceItem represents a compliance item for a specific merchant
type MerchantComplianceItem struct {
	RequirementID  string                 `json:"requirement_id"`
	Requirement    string                 `json:"requirement"`
	Category       ComplianceCategory     `json:"category"`
	Status         ComplianceStatusType   `json:"status"`
	Priority       CompliancePriority     `json:"priority"`
	RiskLevel      models.RiskLevel       `json:"risk_level"`
	DueDate        *time.Time             `json:"due_date"`
	CompletedDate  *time.Time             `json:"completed_date"`
	Evidence       []string               `json:"evidence"`
	Notes          string                 `json:"notes"`
	AssignedTo     string                 `json:"assigned_to"`
	LastReviewDate *time.Time             `json:"last_review_date"`
	NextReviewDate *time.Time             `json:"next_review_date"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ComplianceAssessment represents a compliance assessment
type ComplianceAssessment struct {
	ID              string                      `json:"id" db:"id"`
	MerchantID      string                      `json:"merchant_id" db:"merchant_id"`
	AssessmentType  string                      `json:"assessment_type" db:"assessment_type"`
	AssessorID      string                      `json:"assessor_id" db:"assessor_id"`
	Status          ComplianceStatusType        `json:"status" db:"status"`
	Score           float64                     `json:"score" db:"score"`
	Findings        []*ComplianceFinding        `json:"findings" db:"findings"`
	Recommendations []*ComplianceRecommendation `json:"recommendations" db:"recommendations"`
	StartDate       time.Time                   `json:"start_date" db:"start_date"`
	EndDate         *time.Time                  `json:"end_date" db:"end_date"`
	DueDate         *time.Time                  `json:"due_date" db:"due_date"`
	Metadata        map[string]interface{}      `json:"metadata" db:"metadata"`
	CreatedAt       time.Time                   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time                   `json:"updated_at" db:"updated_at"`
}

// ComplianceFinding represents a finding in a compliance assessment
type ComplianceFinding struct {
	ID             string                 `json:"id"`
	RequirementID  string                 `json:"requirement_id"`
	Finding        string                 `json:"finding"`
	Severity       CompliancePriority     `json:"severity"`
	Description    string                 `json:"description"`
	Evidence       []string               `json:"evidence"`
	Recommendation string                 `json:"recommendation"`
	Status         ComplianceStatusType   `json:"status"`
	AssignedTo     string                 `json:"assigned_to"`
	DueDate        *time.Time             `json:"due_date"`
	ResolvedDate   *time.Time             `json:"resolved_date"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ComplianceAssessmentFilters represents filters for compliance assessment queries
type ComplianceAssessmentFilters struct {
	MerchantID      string               `json:"merchant_id,omitempty"`
	AssessmentType  string               `json:"assessment_type,omitempty"`
	AssessorID      string               `json:"assessor_id,omitempty"`
	Status          ComplianceStatusType `json:"status,omitempty"`
	StartDateAfter  *time.Time           `json:"start_date_after,omitempty"`
	StartDateBefore *time.Time           `json:"start_date_before,omitempty"`
	Limit           int                  `json:"limit,omitempty"`
	Offset          int                  `json:"offset,omitempty"`
}

// NewComplianceService creates a new compliance service
func NewComplianceService(logger *observability.Logger, repository ComplianceRepository, audit AuditServiceInterface) *ComplianceService {
	return &ComplianceService{
		logger:     logger,
		repository: repository,
		audit:      audit,
	}
}

// CreateComplianceRequirement creates a new compliance requirement
func (cs *ComplianceService) CreateComplianceRequirement(ctx context.Context, req *CreateComplianceRequirementRequest) (*ComplianceRequirement, error) {
	requirement := &ComplianceRequirement{
		ID:               uuid.New().String(),
		Regulation:       req.Regulation,
		Requirement:      req.Requirement,
		Description:      req.Description,
		Category:         req.Category,
		Priority:         req.Priority,
		RiskLevel:        req.RiskLevel,
		ApplicableTo:     req.ApplicableTo,
		EffectiveDate:    req.EffectiveDate,
		ExpiryDate:       req.ExpiryDate,
		ReviewFrequency:  req.ReviewFrequency,
		EvidenceRequired: req.EvidenceRequired,
		ValidationRules:  req.ValidationRules,
		Penalties:        req.Penalties,
		Status:           ComplianceStatusPending,
		AssignedTo:       req.AssignedTo,
		Metadata:         req.Metadata,
		CreatedBy:        req.CreatedBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Calculate next review date
	if req.ReviewFrequency != "" {
		requirement.NextReviewDate = cs.calculateNextReviewDate(req.ReviewFrequency)
	}

	// Validate the compliance requirement
	if err := cs.validateComplianceRequirement(requirement); err != nil {
		cs.logger.Error("compliance requirement validation failed", map[string]interface{}{
			"error":       err.Error(),
			"requirement": requirement,
		})
		return nil, fmt.Errorf("compliance requirement validation failed: %w", err)
	}

	// Save to repository
	if err := cs.repository.SaveComplianceRequirement(ctx, requirement); err != nil {
		cs.logger.Error("failed to save compliance requirement", map[string]interface{}{
			"error":       err.Error(),
			"requirement": requirement,
		})
		return nil, fmt.Errorf("failed to save compliance requirement: %w", err)
	}

	// Log the creation
	if err := cs.audit.LogMerchantOperation(ctx, &LogMerchantOperationRequest{
		UserID:       req.CreatedBy,
		MerchantID:   "", // System-wide requirement
		Action:       "create",
		ResourceType: "compliance_requirement",
		ResourceID:   requirement.ID,
		Details:      fmt.Sprintf("Created compliance requirement: %s", requirement.Requirement),
		Description:  fmt.Sprintf("Compliance requirement created for %s regulation", requirement.Regulation),
		Metadata: map[string]interface{}{
			"regulation": requirement.Regulation,
			"category":   requirement.Category,
			"priority":   requirement.Priority,
			"risk_level": requirement.RiskLevel,
		},
	}); err != nil {
		cs.logger.Warn("failed to log compliance requirement creation", map[string]interface{}{
			"error":          err.Error(),
			"requirement_id": requirement.ID,
		})
	}

	cs.logger.Info("compliance requirement created", map[string]interface{}{
		"requirement_id": requirement.ID,
		"regulation":     requirement.Regulation,
		"requirement":    requirement.Requirement,
		"category":       requirement.Category,
	})

	return requirement, nil
}

// CreateComplianceRequirementRequest represents a request to create a compliance requirement
type CreateComplianceRequirementRequest struct {
	Regulation       string                 `json:"regulation" validate:"required"`
	Requirement      string                 `json:"requirement" validate:"required"`
	Description      string                 `json:"description" validate:"required"`
	Category         ComplianceCategory     `json:"category" validate:"required"`
	Priority         CompliancePriority     `json:"priority" validate:"required"`
	RiskLevel        models.RiskLevel       `json:"risk_level" validate:"required"`
	ApplicableTo     []string               `json:"applicable_to,omitempty"`
	EffectiveDate    time.Time              `json:"effective_date" validate:"required"`
	ExpiryDate       *time.Time             `json:"expiry_date,omitempty"`
	ReviewFrequency  string                 `json:"review_frequency,omitempty"`
	EvidenceRequired []string               `json:"evidence_required,omitempty"`
	ValidationRules  []string               `json:"validation_rules,omitempty"`
	Penalties        []string               `json:"penalties,omitempty"`
	AssignedTo       string                 `json:"assigned_to,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy        string                 `json:"created_by" validate:"required"`
}

// ValidateMerchantCompliance validates a merchant against compliance requirements
func (cs *ComplianceService) ValidateMerchantCompliance(ctx context.Context, merchantID string) (*MerchantComplianceStatus, error) {
	if merchantID == "" {
		return nil, fmt.Errorf("merchant ID is required")
	}

	// Get applicable compliance requirements
	requirements, err := cs.repository.GetComplianceRequirements(ctx, &ComplianceRequirementFilters{
		Status: ComplianceStatusPending, // Only active requirements
		Limit:  1000,                    // Get all applicable requirements
	})
	if err != nil {
		cs.logger.Error("failed to get compliance requirements", map[string]interface{}{
			"error":       err.Error(),
			"merchant_id": merchantID,
		})
		return nil, fmt.Errorf("failed to get compliance requirements: %w", err)
	}

	// Get existing compliance records for the merchant
	complianceRecords, err := cs.audit.GetComplianceRecords(ctx, &ComplianceFilters{
		MerchantID: merchantID,
		Limit:      1000,
	})
	if err != nil {
		cs.logger.Error("failed to get compliance records", map[string]interface{}{
			"error":       err.Error(),
			"merchant_id": merchantID,
		})
		return nil, fmt.Errorf("failed to get compliance records: %w", err)
	}

	// Build compliance status
	status := cs.buildMerchantComplianceStatus(merchantID, requirements, complianceRecords)

	// Save the status
	if err := cs.repository.SaveComplianceReport(ctx, &ComplianceReport{
		MerchantID:  merchantID,
		GeneratedAt: time.Now(),
		ComplianceStatus: &ComplianceStatus{
			MerchantID:            status.MerchantID,
			OverallStatus:         status.OverallStatus,
			ComplianceScore:       status.ComplianceScore,
			TotalRequirements:     status.TotalRequirements,
			CompletedRequirements: status.CompletedRequirements,
			OverdueRequirements:   status.OverdueRequirements,
			FailedRequirements:    status.FailedRequirements,
			RiskLevel:             status.RiskLevel,
			LastAssessmentDate:    status.LastAssessmentDate,
			NextAssessmentDate:    status.NextAssessmentDate,
			Requirements:          []*ComplianceRecord{},
			Trends:                status.Trends,
			Alerts:                status.Alerts,
			GeneratedAt:           status.GeneratedAt,
		},
		Summary:         cs.generateComplianceSummary(status),
		Recommendations: cs.generateComplianceRecommendations(status),
		RiskAssessment:  cs.generateRiskAssessment(status),
	}); err != nil {
		cs.logger.Warn("failed to save compliance status report", map[string]interface{}{
			"error":       err.Error(),
			"merchant_id": merchantID,
		})
	}

	cs.logger.Info("merchant compliance validated", map[string]interface{}{
		"merchant_id":        merchantID,
		"overall_status":     status.OverallStatus,
		"compliance_score":   status.ComplianceScore,
		"total_requirements": status.TotalRequirements,
	})

	return status, nil
}

// GenerateComplianceReport generates a comprehensive compliance report
func (cs *ComplianceService) GenerateComplianceReport(ctx context.Context, req *GenerateComplianceReportRequest) (*ComplianceReport, error) {
	// Validate request
	if req.MerchantID == "" {
		return nil, fmt.Errorf("merchant ID is required")
	}

	if req.ReportType == "" {
		return nil, fmt.Errorf("report type is required")
	}

	if req.GeneratedBy == "" {
		return nil, fmt.Errorf("generated by is required")
	}

	// Get compliance status
	status, err := cs.ValidateMerchantCompliance(ctx, req.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate merchant compliance: %w", err)
	}

	// Get compliance assessments
	_, err = cs.repository.GetComplianceAssessments(ctx, &ComplianceAssessmentFilters{
		MerchantID: req.MerchantID,
		Limit:      100,
	})
	if err != nil {
		cs.logger.Warn("failed to get compliance assessments", map[string]interface{}{
			"error":       err.Error(),
			"merchant_id": req.MerchantID,
		})
	}

	// Build comprehensive report
	report := &ComplianceReport{
		MerchantID:  req.MerchantID,
		GeneratedAt: time.Now(),
		ComplianceStatus: &ComplianceStatus{
			MerchantID:            status.MerchantID,
			OverallStatus:         status.OverallStatus,
			ComplianceScore:       status.ComplianceScore,
			TotalRequirements:     status.TotalRequirements,
			CompletedRequirements: status.CompletedRequirements,
			OverdueRequirements:   status.OverdueRequirements,
			FailedRequirements:    status.FailedRequirements,
			RiskLevel:             status.RiskLevel,
			LastAssessmentDate:    status.LastAssessmentDate,
			NextAssessmentDate:    status.NextAssessmentDate,
			Requirements:          []*ComplianceRecord{},
			Trends:                status.Trends,
			Alerts:                status.Alerts,
			GeneratedAt:           status.GeneratedAt,
		},
		Summary:         cs.generateComplianceSummary(status),
		Recommendations: cs.generateComplianceRecommendations(status),
		RiskAssessment:  cs.generateRiskAssessment(status),
	}

	// Save the report
	if err := cs.repository.SaveComplianceReport(ctx, report); err != nil {
		cs.logger.Error("failed to save compliance report", map[string]interface{}{
			"error":       err.Error(),
			"merchant_id": req.MerchantID,
		})
		return nil, fmt.Errorf("failed to save compliance report: %w", err)
	}

	// Log the report generation
	reportID := uuid.New().String()
	if err := cs.audit.LogMerchantOperation(ctx, &LogMerchantOperationRequest{
		UserID:       req.GeneratedBy,
		MerchantID:   req.MerchantID,
		Action:       "generate",
		ResourceType: "compliance_report",
		ResourceID:   reportID,
		Details:      fmt.Sprintf("Generated compliance report: %s", req.ReportType),
		Description:  fmt.Sprintf("Compliance report generated for merchant %s", req.MerchantID),
		Metadata: map[string]interface{}{
			"report_type":      req.ReportType,
			"compliance_score": status.ComplianceScore,
		},
	}); err != nil {
		cs.logger.Warn("failed to log compliance report generation", map[string]interface{}{
			"error":     err.Error(),
			"report_id": reportID,
		})
	}

	cs.logger.Info("compliance report generated", map[string]interface{}{
		"report_id":        reportID,
		"merchant_id":      req.MerchantID,
		"report_type":      req.ReportType,
		"compliance_score": status.ComplianceScore,
	})

	return report, nil
}

// GenerateComplianceReportRequest represents a request to generate a compliance report
type GenerateComplianceReportRequest struct {
	MerchantID  string                 `json:"merchant_id" validate:"required"`
	ReportType  string                 `json:"report_type" validate:"required"`
	GeneratedBy string                 `json:"generated_by" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// validateComplianceRequirement validates a compliance requirement
func (cs *ComplianceService) validateComplianceRequirement(requirement *ComplianceRequirement) error {
	if requirement.Regulation == "" {
		return fmt.Errorf("regulation is required")
	}

	if requirement.Requirement == "" {
		return fmt.Errorf("requirement is required")
	}

	if requirement.Description == "" {
		return fmt.Errorf("description is required")
	}

	if !requirement.Category.IsValid() {
		return fmt.Errorf("invalid compliance category: %s", requirement.Category)
	}

	if !requirement.Priority.IsValid() {
		return fmt.Errorf("invalid compliance priority: %s", requirement.Priority)
	}

	if !requirement.RiskLevel.IsValid() {
		return fmt.Errorf("invalid risk level: %s", requirement.RiskLevel)
	}

	if !requirement.Status.IsValid() {
		return fmt.Errorf("invalid compliance status: %s", requirement.Status)
	}

	return nil
}

// calculateNextReviewDate calculates the next review date based on frequency
func (cs *ComplianceService) calculateNextReviewDate(frequency string) *time.Time {
	now := time.Now()

	switch frequency {
	case "daily":
		nextDate := now.AddDate(0, 0, 1)
		return &nextDate
	case "weekly":
		nextDate := now.AddDate(0, 0, 7)
		return &nextDate
	case "monthly":
		nextDate := now.AddDate(0, 1, 0)
		return &nextDate
	case "quarterly":
		nextDate := now.AddDate(0, 3, 0)
		return &nextDate
	case "annually":
		nextDate := now.AddDate(1, 0, 0)
		return &nextDate
	default:
		nextDate := now.AddDate(0, 1, 0) // Default to monthly
		return &nextDate
	}
}

// buildMerchantComplianceStatus builds the compliance status for a merchant
func (cs *ComplianceService) buildMerchantComplianceStatus(merchantID string, requirements []*ComplianceRequirement, records []*ComplianceRecord) *MerchantComplianceStatus {
	status := &MerchantComplianceStatus{
		MerchantID:            merchantID,
		OverallStatus:         ComplianceStatusCompleted,
		ComplianceScore:       1.0,
		TotalRequirements:     len(requirements),
		CompletedRequirements: 0,
		OverdueRequirements:   0,
		FailedRequirements:    0,
		RiskLevel:             models.RiskLevelLow,
		LastAssessmentDate:    time.Now(),
		NextAssessmentDate:    time.Now().AddDate(0, 1, 0),
		Requirements:          []*MerchantComplianceItem{},
		Trends:                []*ComplianceTrend{},
		Alerts:                []*ComplianceAlert{},
		GeneratedAt:           time.Now(),
	}

	// Create a map of existing compliance records for quick lookup
	recordMap := make(map[string]*ComplianceRecord)
	for _, record := range records {
		recordMap[record.Requirement] = record
	}

	// Process each requirement
	for _, req := range requirements {
		item := &MerchantComplianceItem{
			RequirementID:  req.ID,
			Requirement:    req.Requirement,
			Category:       req.Category,
			Status:         ComplianceStatusPending,
			Priority:       req.Priority,
			RiskLevel:      req.RiskLevel,
			DueDate:        req.NextReviewDate,
			AssignedTo:     req.AssignedTo,
			LastReviewDate: req.LastReviewDate,
			NextReviewDate: req.NextReviewDate,
			Metadata:       req.Metadata,
		}

		// Check if there's an existing compliance record
		if record, exists := recordMap[req.Requirement]; exists {
			item.Status = record.Status
			item.CompletedDate = record.CompletedDate
			item.Evidence = record.Evidence
			item.Notes = record.Notes
		}

		// Update counters
		switch item.Status {
		case ComplianceStatusCompleted:
			status.CompletedRequirements++
		case ComplianceStatusOverdue:
			status.OverdueRequirements++
		case ComplianceStatusFailed:
			status.FailedRequirements++
		}

		status.Requirements = append(status.Requirements, item)
	}

	// Calculate compliance score
	if status.TotalRequirements > 0 {
		status.ComplianceScore = float64(status.CompletedRequirements) / float64(status.TotalRequirements)
	}

	// Determine overall status
	if status.OverdueRequirements > 0 || status.FailedRequirements > 0 {
		status.OverallStatus = ComplianceStatusFailed
	} else if status.CompletedRequirements == status.TotalRequirements {
		status.OverallStatus = ComplianceStatusCompleted
	} else {
		status.OverallStatus = ComplianceStatusInProgress
	}

	// Determine risk level based on compliance score
	if status.ComplianceScore < 0.5 {
		status.RiskLevel = models.RiskLevelHigh
	} else if status.ComplianceScore < 0.8 {
		status.RiskLevel = models.RiskLevelMedium
	} else {
		status.RiskLevel = models.RiskLevelLow
	}

	return status
}

// generateComplianceSummary generates a compliance summary
func (cs *ComplianceService) generateComplianceSummary(status *MerchantComplianceStatus) *ComplianceSummary {
	return &ComplianceSummary{
		TotalRequirements:     status.TotalRequirements,
		CompletedRequirements: status.CompletedRequirements,
		OverdueRequirements:   status.OverdueRequirements,
		FailedRequirements:    status.FailedRequirements,
		ComplianceScore:       status.ComplianceScore,
		RiskLevel:             string(status.RiskLevel),
		LastAssessment:        status.LastAssessmentDate,
		NextAssessment:        status.NextAssessmentDate,
	}
}

// generateComplianceRecommendations generates compliance recommendations
func (cs *ComplianceService) generateComplianceRecommendations(status *MerchantComplianceStatus) []*ComplianceRecommendation {
	var recommendations []*ComplianceRecommendation

	// Add recommendations based on compliance status
	if status.OverdueRequirements > 0 {
		recommendations = append(recommendations, &ComplianceRecommendation{
			Type:        "overdue",
			Title:       "Address Overdue Requirements",
			Description: fmt.Sprintf("There are %d overdue compliance requirements that need immediate attention.", status.OverdueRequirements),
			Priority:    "high",
			Action:      "Review and complete overdue requirements",
		})
	}

	if status.FailedRequirements > 0 {
		recommendations = append(recommendations, &ComplianceRecommendation{
			Type:        "failed",
			Title:       "Resolve Failed Requirements",
			Description: fmt.Sprintf("There are %d failed compliance requirements that need to be addressed.", status.FailedRequirements),
			Priority:    "critical",
			Action:      "Investigate and resolve failed requirements",
		})
	}

	if status.ComplianceScore < 0.8 {
		recommendations = append(recommendations, &ComplianceRecommendation{
			Type:        "score",
			Title:       "Improve Compliance Score",
			Description: fmt.Sprintf("The compliance score of %.2f is below the recommended threshold of 0.8.", status.ComplianceScore),
			Priority:    "medium",
			Action:      "Focus on completing pending requirements",
		})
	}

	return recommendations
}

// generateRiskAssessment generates a risk assessment
func (cs *ComplianceService) generateRiskAssessment(status *MerchantComplianceStatus) *RiskAssessment {
	riskScore := 1.0 - status.ComplianceScore // Higher compliance = lower risk

	var riskFactors []*RiskFactor
	var mitigations []*RiskMitigation

	// Add risk factors based on compliance status
	if status.OverdueRequirements > 0 {
		riskFactors = append(riskFactors, &RiskFactor{
			Factor:      "Overdue Requirements",
			Description: fmt.Sprintf("%d overdue compliance requirements", status.OverdueRequirements),
			Score:       0.8,
			Impact:      "high",
		})
	}

	if status.FailedRequirements > 0 {
		riskFactors = append(riskFactors, &RiskFactor{
			Factor:      "Failed Requirements",
			Description: fmt.Sprintf("%d failed compliance requirements", status.FailedRequirements),
			Score:       1.0,
			Impact:      "critical",
		})
	}

	// Add mitigations
	mitigations = append(mitigations, &RiskMitigation{
		Mitigation:    "Regular Compliance Monitoring",
		Description:   "Implement regular monitoring of compliance requirements",
		Status:        "recommended",
		Effectiveness: "high",
	})

	mitigations = append(mitigations, &RiskMitigation{
		Mitigation:    "Automated Alerts",
		Description:   "Set up automated alerts for overdue and failed requirements",
		Status:        "recommended",
		Effectiveness: "medium",
	})

	return &RiskAssessment{
		OverallRisk:    string(status.RiskLevel),
		RiskScore:      riskScore,
		RiskFactors:    riskFactors,
		Mitigations:    mitigations,
		LastAssessment: time.Now(),
	}
}
