package services

import (
	"context"
	"fmt"
	"time"

	"kyb-platform/internal/compliance"
	"kyb-platform/internal/models"
	"kyb-platform/internal/observability"

	"github.com/google/uuid"
)

// ComplianceSystem defines the interface for compliance audit operations
type ComplianceSystem interface {
	RecordAuditEvent(ctx context.Context, event *compliance.AuditEvent) error
	GetAuditEvents(ctx context.Context, filter *compliance.AuditFilter) ([]*compliance.AuditEvent, error)
	GetAuditTrail(ctx context.Context, entityType, entityID string) (*compliance.ComplianceAuditTrail, error)
	GenerateAuditReport(ctx context.Context, filter *compliance.AuditFilter) (*compliance.AuditSummary, error)
	GetAuditMetrics(ctx context.Context, filter *compliance.AuditFilter) (*compliance.AuditMetrics, error)
	UpdateAuditMetrics(ctx context.Context, metrics *compliance.AuditMetrics) error
}

// AuditService provides comprehensive audit logging and AML compliance tracking
type AuditService struct {
	logger     *observability.Logger
	compliance ComplianceSystem
	repository AuditRepository
}

// AuditRepository defines the interface for audit data persistence
type AuditRepository interface {
	// SaveAuditLog saves an audit log entry
	SaveAuditLog(ctx context.Context, auditLog *models.AuditLog) error

	// GetAuditLogs retrieves audit logs with filtering
	GetAuditLogs(ctx context.Context, filters *AuditLogFilters) ([]*models.AuditLog, error)

	// GetAuditLogByID retrieves a specific audit log by ID
	GetAuditLogByID(ctx context.Context, id string) (*models.AuditLog, error)

	// GetAuditTrail retrieves audit trail for a specific merchant
	GetAuditTrail(ctx context.Context, merchantID string, limit int, offset int) ([]*models.AuditLog, error)

	// SaveComplianceRecord saves a compliance record
	SaveComplianceRecord(ctx context.Context, record *ComplianceRecord) error

	// GetComplianceRecords retrieves compliance records with filtering
	GetComplianceRecords(ctx context.Context, filters *ComplianceFilters) ([]*ComplianceRecord, error)

	// GetComplianceStatus retrieves compliance status for a merchant
	GetComplianceStatus(ctx context.Context, merchantID string) (*ComplianceStatus, error)
}

// AuditLogFilters represents filters for audit log queries
type AuditLogFilters struct {
	UserID       string    `json:"user_id,omitempty"`
	MerchantID   string    `json:"merchant_id,omitempty"`
	Action       string    `json:"action,omitempty"`
	ResourceType string    `json:"resource_type,omitempty"`
	ResourceID   string    `json:"resource_id,omitempty"`
	StartDate    time.Time `json:"start_date,omitempty"`
	EndDate      time.Time `json:"end_date,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	RequestID    string    `json:"request_id,omitempty"`
	Limit        int       `json:"limit,omitempty"`
	Offset       int       `json:"offset,omitempty"`
}

// ComplianceRecord represents a compliance tracking record
type ComplianceRecord struct {
	ID             string                 `json:"id" db:"id"`
	MerchantID     string                 `json:"merchant_id" db:"merchant_id"`
	ComplianceType ComplianceType         `json:"compliance_type" db:"compliance_type"`
	Status         ComplianceStatusType   `json:"status" db:"status"`
	Requirement    string                 `json:"requirement" db:"requirement"`
	Description    string                 `json:"description" db:"description"`
	DueDate        *time.Time             `json:"due_date" db:"due_date"`
	CompletedDate  *time.Time             `json:"completed_date" db:"completed_date"`
	AssignedTo     string                 `json:"assigned_to" db:"assigned_to"`
	Priority       CompliancePriority     `json:"priority" db:"priority"`
	RiskLevel      models.RiskLevel       `json:"risk_level" db:"risk_level"`
	Evidence       []string               `json:"evidence" db:"evidence"`
	Notes          string                 `json:"notes" db:"notes"`
	LastReviewDate *time.Time             `json:"last_review_date" db:"last_review_date"`
	NextReviewDate *time.Time             `json:"next_review_date" db:"next_review_date"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedBy      string                 `json:"created_by" db:"created_by"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// ComplianceType represents the type of compliance requirement
type ComplianceType string

const (
	ComplianceTypeAML      ComplianceType = "aml"
	ComplianceTypeKYC      ComplianceType = "kyc"
	ComplianceTypeKYB      ComplianceType = "kyb"
	ComplianceTypeFATF     ComplianceType = "fatf"
	ComplianceTypeGDPR     ComplianceType = "gdpr"
	ComplianceTypeSOX      ComplianceType = "sox"
	ComplianceTypePCI      ComplianceType = "pci"
	ComplianceTypeISO27001 ComplianceType = "iso27001"
	ComplianceTypeSOC2     ComplianceType = "soc2"
	ComplianceTypeCustom   ComplianceType = "custom"
)

// IsValid checks if the compliance type is valid
func (ct ComplianceType) IsValid() bool {
	switch ct {
	case ComplianceTypeAML, ComplianceTypeKYC, ComplianceTypeKYB, ComplianceTypeFATF,
		ComplianceTypeGDPR, ComplianceTypeSOX, ComplianceTypePCI, ComplianceTypeISO27001,
		ComplianceTypeSOC2, ComplianceTypeCustom:
		return true
	default:
		return false
	}
}

// String returns the string representation of the compliance type
func (ct ComplianceType) String() string {
	return string(ct)
}

// ComplianceStatusType represents the status of a compliance requirement
type ComplianceStatusType string

const (
	ComplianceStatusPending    ComplianceStatusType = "pending"
	ComplianceStatusInProgress ComplianceStatusType = "in_progress"
	ComplianceStatusCompleted  ComplianceStatusType = "completed"
	ComplianceStatusOverdue    ComplianceStatusType = "overdue"
	ComplianceStatusFailed     ComplianceStatusType = "failed"
	ComplianceStatusWaived     ComplianceStatusType = "waived"
	ComplianceStatusExempt     ComplianceStatusType = "exempt"
)

// IsValid checks if the compliance status is valid
func (cst ComplianceStatusType) IsValid() bool {
	switch cst {
	case ComplianceStatusPending, ComplianceStatusInProgress, ComplianceStatusCompleted,
		ComplianceStatusOverdue, ComplianceStatusFailed, ComplianceStatusWaived, ComplianceStatusExempt:
		return true
	default:
		return false
	}
}

// String returns the string representation of the compliance status
func (cst ComplianceStatusType) String() string {
	return string(cst)
}

// CompliancePriority represents the priority of a compliance requirement
type CompliancePriority string

const (
	CompliancePriorityLow      CompliancePriority = "low"
	CompliancePriorityMedium   CompliancePriority = "medium"
	CompliancePriorityHigh     CompliancePriority = "high"
	CompliancePriorityCritical CompliancePriority = "critical"
)

// IsValid checks if the compliance priority is valid
func (cp CompliancePriority) IsValid() bool {
	switch cp {
	case CompliancePriorityLow, CompliancePriorityMedium, CompliancePriorityHigh, CompliancePriorityCritical:
		return true
	default:
		return false
	}
}

// String returns the string representation of the compliance priority
func (cp CompliancePriority) String() string {
	return string(cp)
}

// GetNumericValue returns a numeric value for priority comparison
func (cp CompliancePriority) GetNumericValue() int {
	switch cp {
	case CompliancePriorityCritical:
		return 4
	case CompliancePriorityHigh:
		return 3
	case CompliancePriorityMedium:
		return 2
	case CompliancePriorityLow:
		return 1
	default:
		return 0
	}
}

// ComplianceFilters represents filters for compliance record queries
type ComplianceFilters struct {
	MerchantID     string               `json:"merchant_id,omitempty"`
	ComplianceType ComplianceType       `json:"compliance_type,omitempty"`
	Status         ComplianceStatusType `json:"status,omitempty"`
	Priority       CompliancePriority   `json:"priority,omitempty"`
	RiskLevel      models.RiskLevel     `json:"risk_level,omitempty"`
	AssignedTo     string               `json:"assigned_to,omitempty"`
	DueDateAfter   *time.Time           `json:"due_date_after,omitempty"`
	DueDateBefore  *time.Time           `json:"due_date_before,omitempty"`
	Overdue        bool                 `json:"overdue,omitempty"`
	Limit          int                  `json:"limit,omitempty"`
	Offset         int                  `json:"offset,omitempty"`
}

// ComplianceStatus represents the overall compliance status for a merchant
type ComplianceStatus struct {
	MerchantID            string               `json:"merchant_id"`
	OverallStatus         ComplianceStatusType `json:"overall_status"`
	ComplianceScore       float64              `json:"compliance_score"`
	TotalRequirements     int                  `json:"total_requirements"`
	CompletedRequirements int                  `json:"completed_requirements"`
	OverdueRequirements   int                  `json:"overdue_requirements"`
	FailedRequirements    int                  `json:"failed_requirements"`
	RiskLevel             models.RiskLevel     `json:"risk_level"`
	LastAssessmentDate    time.Time            `json:"last_assessment_date"`
	NextAssessmentDate    time.Time            `json:"next_assessment_date"`
	Requirements          []*ComplianceRecord  `json:"requirements"`
	Trends                []*ComplianceTrend   `json:"trends"`
	Alerts                []*ComplianceAlert   `json:"alerts"`
	GeneratedAt           time.Time            `json:"generated_at"`
}

// ComplianceTrend represents compliance trend data
type ComplianceTrend struct {
	Date         time.Time            `json:"date"`
	Score        float64              `json:"score"`
	Status       ComplianceStatusType `json:"status"`
	Requirements int                  `json:"requirements"`
}

// ComplianceAlert represents a compliance alert
type ComplianceAlert struct {
	ID         string               `json:"id"`
	Type       string               `json:"type"`
	Title      string               `json:"title"`
	Message    string               `json:"message"`
	Priority   CompliancePriority   `json:"priority"`
	Status     ComplianceStatusType `json:"status"`
	CreatedAt  time.Time            `json:"created_at"`
	ResolvedAt *time.Time           `json:"resolved_at,omitempty"`
}

// FATFRecommendation represents FATF recommendation compliance tracking
type FATFRecommendation struct {
	ID             string                 `json:"id"`
	Recommendation string                 `json:"recommendation"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"`
	Priority       CompliancePriority     `json:"priority"`
	Status         ComplianceStatusType   `json:"status"`
	Implementation string                 `json:"implementation"`
	Evidence       []string               `json:"evidence"`
	LastReview     *time.Time             `json:"last_review"`
	NextReview     *time.Time             `json:"next_review"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// NewAuditService creates a new audit service
func NewAuditService(logger *observability.Logger, compliance ComplianceSystem, repository AuditRepository) *AuditService {
	return &AuditService{
		logger:     logger,
		compliance: compliance,
		repository: repository,
	}
}

// LogMerchantOperation logs a merchant operation for audit purposes
func (as *AuditService) LogMerchantOperation(ctx context.Context, req *LogMerchantOperationRequest) error {
	auditLog := &models.AuditLog{
		ID:           uuid.New().String(),
		UserID:       req.UserID,
		MerchantID:   req.MerchantID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Details:      req.Details,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		RequestID:    req.RequestID,
		CreatedAt:    time.Now(),
	}

	// Validate the audit log
	if err := auditLog.Validate(); err != nil {
		as.logger.Error("audit log validation failed", map[string]interface{}{
			"error":     err.Error(),
			"audit_log": auditLog,
		})
		return fmt.Errorf("audit log validation failed: %w", err)
	}

	// Save to repository
	if err := as.repository.SaveAuditLog(ctx, auditLog); err != nil {
		as.logger.Error("failed to save audit log", map[string]interface{}{
			"error":     err.Error(),
			"audit_log": auditLog,
		})
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	// Also log to compliance system
	complianceEvent := &compliance.AuditEvent{
		ID:            auditLog.ID,
		UserID:        auditLog.UserID,
		Action:        compliance.AuditAction(auditLog.Action),
		Resource:      auditLog.ResourceType,
		ResourceID:    auditLog.ResourceID,
		Details:       auditLog.Details,
		IPAddress:     auditLog.IPAddress,
		UserAgent:     auditLog.UserAgent,
		Timestamp:     auditLog.CreatedAt,
		BusinessID:    auditLog.MerchantID,
		EventType:     "merchant_operation",
		EventCategory: "audit",
		EntityType:    auditLog.ResourceType,
		EntityID:      auditLog.ResourceID,
		Description:   req.Description,
		UserName:      req.UserName,
		UserRole:      req.UserRole,
		UserEmail:     req.UserEmail,
		SessionID:     req.SessionID,
		RequestID:     auditLog.RequestID,
		Success:       true,
		Metadata:      req.Metadata,
	}

	if err := as.compliance.RecordAuditEvent(ctx, complianceEvent); err != nil {
		as.logger.Warn("failed to record compliance audit event", map[string]interface{}{
			"error":        err.Error(),
			"audit_log_id": auditLog.ID,
		})
		// Don't fail the operation if compliance logging fails
	}

	as.logger.Info("merchant operation logged", map[string]interface{}{
		"audit_log_id": auditLog.ID,
		"merchant_id":  auditLog.MerchantID,
		"action":       auditLog.Action,
		"user_id":      auditLog.UserID,
	})

	return nil
}

// LogMerchantOperationRequest represents a request to log a merchant operation
type LogMerchantOperationRequest struct {
	UserID       string                 `json:"user_id" validate:"required"`
	MerchantID   string                 `json:"merchant_id" validate:"required"`
	Action       string                 `json:"action" validate:"required"`
	ResourceType string                 `json:"resource_type" validate:"required"`
	ResourceID   string                 `json:"resource_id" validate:"required"`
	Details      string                 `json:"details" validate:"required"`
	Description  string                 `json:"description" validate:"required"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	RequestID    string                 `json:"request_id"`
	SessionID    string                 `json:"session_id"`
	UserName     string                 `json:"user_name"`
	UserRole     string                 `json:"user_role"`
	UserEmail    string                 `json:"user_email"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// GetAuditTrail retrieves the audit trail for a specific merchant
func (as *AuditService) GetAuditTrail(ctx context.Context, merchantID string, limit int, offset int) ([]*models.AuditLog, error) {
	if merchantID == "" {
		return nil, fmt.Errorf("merchant ID is required")
	}

	auditLogs, err := as.repository.GetAuditTrail(ctx, merchantID, limit, offset)
	if err != nil {
		as.logger.Error("failed to get audit trail", map[string]interface{}{
			"error":       err.Error(),
			"merchant_id": merchantID,
		})
		return nil, fmt.Errorf("failed to get audit trail: %w", err)
	}

	as.logger.Info("audit trail retrieved", map[string]interface{}{
		"merchant_id": merchantID,
		"count":       len(auditLogs),
		"limit":       limit,
		"offset":      offset,
	})

	return auditLogs, nil
}

// GetAuditLogs retrieves audit logs with filtering
func (as *AuditService) GetAuditLogs(ctx context.Context, filters *AuditLogFilters) ([]*models.AuditLog, error) {
	auditLogs, err := as.repository.GetAuditLogs(ctx, filters)
	if err != nil {
		as.logger.Error("failed to get audit logs", map[string]interface{}{
			"error":   err.Error(),
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	as.logger.Info("audit logs retrieved", map[string]interface{}{
		"count":   len(auditLogs),
		"filters": filters,
	})

	return auditLogs, nil
}

// CreateComplianceRecord creates a new compliance record
func (as *AuditService) CreateComplianceRecord(ctx context.Context, req *CreateComplianceRecordRequest) (*ComplianceRecord, error) {
	record := &ComplianceRecord{
		ID:             uuid.New().String(),
		MerchantID:     req.MerchantID,
		ComplianceType: req.ComplianceType,
		Status:         ComplianceStatusPending,
		Requirement:    req.Requirement,
		Description:    req.Description,
		DueDate:        req.DueDate,
		AssignedTo:     req.AssignedTo,
		Priority:       req.Priority,
		RiskLevel:      req.RiskLevel,
		Evidence:       req.Evidence,
		Notes:          req.Notes,
		Metadata:       req.Metadata,
		CreatedBy:      req.CreatedBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Validate the compliance record
	if err := as.validateComplianceRecord(record); err != nil {
		as.logger.Error("compliance record validation failed", map[string]interface{}{
			"error":  err.Error(),
			"record": record,
		})
		return nil, fmt.Errorf("compliance record validation failed: %w", err)
	}

	// Save to repository
	if err := as.repository.SaveComplianceRecord(ctx, record); err != nil {
		as.logger.Error("failed to save compliance record", map[string]interface{}{
			"error":  err.Error(),
			"record": record,
		})
		return nil, fmt.Errorf("failed to save compliance record: %w", err)
	}

	// Log the creation
	if err := as.LogMerchantOperation(ctx, &LogMerchantOperationRequest{
		UserID:       req.CreatedBy,
		MerchantID:   req.MerchantID,
		Action:       "create",
		ResourceType: "compliance_record",
		ResourceID:   record.ID,
		Details:      fmt.Sprintf("Created compliance record: %s", record.Requirement),
		Description:  fmt.Sprintf("Compliance record created for %s requirement", record.ComplianceType),
		Metadata: map[string]interface{}{
			"compliance_type": record.ComplianceType,
			"priority":        record.Priority,
			"risk_level":      record.RiskLevel,
		},
	}); err != nil {
		as.logger.Warn("failed to log compliance record creation", map[string]interface{}{
			"error":     err.Error(),
			"record_id": record.ID,
		})
	}

	as.logger.Info("compliance record created", map[string]interface{}{
		"record_id":       record.ID,
		"merchant_id":     record.MerchantID,
		"compliance_type": record.ComplianceType,
		"requirement":     record.Requirement,
	})

	return record, nil
}

// CreateComplianceRecordRequest represents a request to create a compliance record
type CreateComplianceRecordRequest struct {
	MerchantID     string                 `json:"merchant_id" validate:"required"`
	ComplianceType ComplianceType         `json:"compliance_type" validate:"required"`
	Requirement    string                 `json:"requirement" validate:"required"`
	Description    string                 `json:"description" validate:"required"`
	DueDate        *time.Time             `json:"due_date,omitempty"`
	AssignedTo     string                 `json:"assigned_to,omitempty"`
	Priority       CompliancePriority     `json:"priority" validate:"required"`
	RiskLevel      models.RiskLevel       `json:"risk_level" validate:"required"`
	Evidence       []string               `json:"evidence,omitempty"`
	Notes          string                 `json:"notes,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy      string                 `json:"created_by" validate:"required"`
}

// UpdateComplianceRecord updates an existing compliance record
func (as *AuditService) UpdateComplianceRecord(ctx context.Context, recordID string, req *UpdateComplianceRecordRequest) (*ComplianceRecord, error) {
	// Get existing record
	records, err := as.repository.GetComplianceRecords(ctx, &ComplianceFilters{
		Limit:  1,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance record: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("compliance record not found: %s", recordID)
	}

	record := records[0]

	// Update fields
	if req.Status != "" {
		record.Status = req.Status
	}
	if req.Description != "" {
		record.Description = req.Description
	}
	if req.DueDate != nil {
		record.DueDate = req.DueDate
	}
	if req.AssignedTo != "" {
		record.AssignedTo = req.AssignedTo
	}
	if req.Priority != "" {
		record.Priority = req.Priority
	}
	if req.RiskLevel != "" {
		record.RiskLevel = req.RiskLevel
	}
	if req.Evidence != nil {
		record.Evidence = req.Evidence
	}
	if req.Notes != "" {
		record.Notes = req.Notes
	}
	if req.Metadata != nil {
		record.Metadata = req.Metadata
	}

	// Set completion date if status is completed
	if req.Status == ComplianceStatusCompleted && record.CompletedDate == nil {
		now := time.Now()
		record.CompletedDate = &now
	}

	record.UpdatedAt = time.Now()

	// Validate the updated record
	if err := as.validateComplianceRecord(record); err != nil {
		as.logger.Error("compliance record validation failed", map[string]interface{}{
			"error":  err.Error(),
			"record": record,
		})
		return nil, fmt.Errorf("compliance record validation failed: %w", err)
	}

	// Save to repository
	if err := as.repository.SaveComplianceRecord(ctx, record); err != nil {
		as.logger.Error("failed to save compliance record", map[string]interface{}{
			"error":  err.Error(),
			"record": record,
		})
		return nil, fmt.Errorf("failed to save compliance record: %w", err)
	}

	// Log the update
	if err := as.LogMerchantOperation(ctx, &LogMerchantOperationRequest{
		UserID:       req.UpdatedBy,
		MerchantID:   record.MerchantID,
		Action:       "update",
		ResourceType: "compliance_record",
		ResourceID:   record.ID,
		Details:      fmt.Sprintf("Updated compliance record: %s", record.Requirement),
		Description:  fmt.Sprintf("Compliance record updated for %s requirement", record.ComplianceType),
		Metadata: map[string]interface{}{
			"compliance_type": record.ComplianceType,
			"status":          record.Status,
			"priority":        record.Priority,
			"risk_level":      record.RiskLevel,
		},
	}); err != nil {
		as.logger.Warn("failed to log compliance record update", map[string]interface{}{
			"error":     err.Error(),
			"record_id": record.ID,
		})
	}

	as.logger.Info("compliance record updated", map[string]interface{}{
		"record_id":       record.ID,
		"merchant_id":     record.MerchantID,
		"compliance_type": record.ComplianceType,
		"status":          record.Status,
	})

	return record, nil
}

// UpdateComplianceRecordRequest represents a request to update a compliance record
type UpdateComplianceRecordRequest struct {
	Status      ComplianceStatusType   `json:"status,omitempty"`
	Description string                 `json:"description,omitempty"`
	DueDate     *time.Time             `json:"due_date,omitempty"`
	AssignedTo  string                 `json:"assigned_to,omitempty"`
	Priority    CompliancePriority     `json:"priority,omitempty"`
	RiskLevel   models.RiskLevel       `json:"risk_level,omitempty"`
	Evidence    []string               `json:"evidence,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	UpdatedBy   string                 `json:"updated_by" validate:"required"`
}

// GetComplianceStatus retrieves the compliance status for a merchant
func (as *AuditService) GetComplianceStatus(ctx context.Context, merchantID string) (*ComplianceStatus, error) {
	if merchantID == "" {
		return nil, fmt.Errorf("merchant ID is required")
	}

	status, err := as.repository.GetComplianceStatus(ctx, merchantID)
	if err != nil {
		as.logger.Error("failed to get compliance status", map[string]interface{}{
			"error":       err.Error(),
			"merchant_id": merchantID,
		})
		return nil, fmt.Errorf("failed to get compliance status: %w", err)
	}

	as.logger.Info("compliance status retrieved", map[string]interface{}{
		"merchant_id":      merchantID,
		"overall_status":   status.OverallStatus,
		"compliance_score": status.ComplianceScore,
	})

	return status, nil
}

// GetComplianceRecords retrieves compliance records with filtering
func (as *AuditService) GetComplianceRecords(ctx context.Context, filters *ComplianceFilters) ([]*ComplianceRecord, error) {
	records, err := as.repository.GetComplianceRecords(ctx, filters)
	if err != nil {
		as.logger.Error("failed to get compliance records", map[string]interface{}{
			"error":   err.Error(),
			"filters": filters,
		})
		return nil, fmt.Errorf("failed to get compliance records: %w", err)
	}

	as.logger.Info("compliance records retrieved", map[string]interface{}{
		"count":   len(records),
		"filters": filters,
	})

	return records, nil
}

// TrackFATFCompliance tracks FATF recommendation compliance
func (as *AuditService) TrackFATFCompliance(ctx context.Context, merchantID string, recommendation *FATFRecommendation) error {
	// Create compliance record for FATF recommendation
	req := &CreateComplianceRecordRequest{
		MerchantID:     merchantID,
		ComplianceType: ComplianceTypeFATF,
		Requirement:    recommendation.Recommendation,
		Description:    recommendation.Description,
		DueDate:        recommendation.NextReview,
		Priority:       recommendation.Priority,
		RiskLevel:      models.RiskLevelHigh, // FATF recommendations are typically high risk
		Evidence:       recommendation.Evidence,
		Notes:          recommendation.Implementation,
		Metadata: map[string]interface{}{
			"fatf_category": recommendation.Category,
			"fatf_id":       recommendation.ID,
		},
		CreatedBy: "system", // System-created for FATF tracking
	}

	_, err := as.CreateComplianceRecord(ctx, req)
	if err != nil {
		as.logger.Error("failed to create FATF compliance record", map[string]interface{}{
			"error":          err.Error(),
			"merchant_id":    merchantID,
			"recommendation": recommendation.Recommendation,
		})
		return fmt.Errorf("failed to create FATF compliance record: %w", err)
	}

	as.logger.Info("FATF compliance tracked", map[string]interface{}{
		"merchant_id":    merchantID,
		"recommendation": recommendation.Recommendation,
		"category":       recommendation.Category,
	})

	return nil
}

// validateComplianceRecord validates a compliance record
func (as *AuditService) validateComplianceRecord(record *ComplianceRecord) error {
	if record.MerchantID == "" {
		return fmt.Errorf("merchant ID is required")
	}

	if !record.ComplianceType.IsValid() {
		return fmt.Errorf("invalid compliance type: %s", record.ComplianceType)
	}

	if !record.Status.IsValid() {
		return fmt.Errorf("invalid compliance status: %s", record.Status)
	}

	if !record.Priority.IsValid() {
		return fmt.Errorf("invalid compliance priority: %s", record.Priority)
	}

	if !record.RiskLevel.IsValid() {
		return fmt.Errorf("invalid risk level: %s", record.RiskLevel)
	}

	if record.Requirement == "" {
		return fmt.Errorf("requirement is required")
	}

	if record.Description == "" {
		return fmt.Errorf("description is required")
	}

	return nil
}

// GenerateComplianceReport generates a compliance report for a merchant
func (as *AuditService) GenerateComplianceReport(ctx context.Context, merchantID string) (*ComplianceReport, error) {
	// Get compliance status
	status, err := as.GetComplianceStatus(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance status: %w", err)
	}

	// Get audit trail
	auditTrail, err := as.GetAuditTrail(ctx, merchantID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit trail: %w", err)
	}

	report := &ComplianceReport{
		MerchantID:       merchantID,
		GeneratedAt:      time.Now(),
		ComplianceStatus: status,
		AuditTrail:       auditTrail,
		Summary:          as.generateComplianceSummary(status),
		Recommendations:  as.generateComplianceRecommendations(status),
		RiskAssessment:   as.generateRiskAssessment(status),
	}

	as.logger.Info("compliance report generated", map[string]interface{}{
		"merchant_id":      merchantID,
		"compliance_score": status.ComplianceScore,
		"overall_status":   status.OverallStatus,
	})

	return report, nil
}

// ComplianceReport represents a comprehensive compliance report
type ComplianceReport struct {
	MerchantID       string                      `json:"merchant_id"`
	GeneratedAt      time.Time                   `json:"generated_at"`
	ComplianceStatus *ComplianceStatus           `json:"compliance_status"`
	AuditTrail       []*models.AuditLog          `json:"audit_trail"`
	Summary          *ComplianceSummary          `json:"summary"`
	Recommendations  []*ComplianceRecommendation `json:"recommendations"`
	RiskAssessment   *RiskAssessment             `json:"risk_assessment"`
}

// ComplianceSummary represents a summary of compliance status
type ComplianceSummary struct {
	TotalRequirements     int       `json:"total_requirements"`
	CompletedRequirements int       `json:"completed_requirements"`
	OverdueRequirements   int       `json:"overdue_requirements"`
	FailedRequirements    int       `json:"failed_requirements"`
	ComplianceScore       float64   `json:"compliance_score"`
	RiskLevel             string    `json:"risk_level"`
	LastAssessment        time.Time `json:"last_assessment"`
	NextAssessment        time.Time `json:"next_assessment"`
}

// ComplianceRecommendation represents a compliance recommendation
type ComplianceRecommendation struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Action      string `json:"action"`
}

// RiskAssessment represents a risk assessment
type RiskAssessment struct {
	OverallRisk    string            `json:"overall_risk"`
	RiskScore      float64           `json:"risk_score"`
	RiskFactors    []*RiskFactor     `json:"risk_factors"`
	Mitigations    []*RiskMitigation `json:"mitigations"`
	LastAssessment time.Time         `json:"last_assessment"`
}

// RiskFactor represents a risk factor
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	Impact      string  `json:"impact"`
}

// RiskMitigation represents a risk mitigation
type RiskMitigation struct {
	Mitigation    string `json:"mitigation"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	Effectiveness string `json:"effectiveness"`
}

// generateComplianceSummary generates a compliance summary
func (as *AuditService) generateComplianceSummary(status *ComplianceStatus) *ComplianceSummary {
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
func (as *AuditService) generateComplianceRecommendations(status *ComplianceStatus) []*ComplianceRecommendation {
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
func (as *AuditService) generateRiskAssessment(status *ComplianceStatus) *RiskAssessment {
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
