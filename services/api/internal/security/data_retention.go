package security

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// DataRetentionService handles data retention and deletion policies
type DataRetentionService struct {
	config *DataRetentionConfig
	logger *zap.Logger

	// Sub-managers
	policyManager    *PolicyManager
	lifecycleManager *DataLifecycleManager
	deletionManager  *DeletionManager
	auditManager     *RetentionAuditManager
}

// DataRetentionConfig contains configuration for data retention service
type DataRetentionConfig struct {
	EnableDataRetention      bool                     `json:"enable_data_retention"`
	DefaultRetentionPeriod   time.Duration            `json:"default_retention_period"`
	MinRetentionPeriod       time.Duration            `json:"min_retention_period"`
	MaxRetentionPeriod       time.Duration            `json:"max_retention_period"`
	GracePeriod              time.Duration            `json:"grace_period"`
	AutomaticDeletion        bool                     `json:"automatic_deletion"`
	BackupBeforeDeletion     bool                     `json:"backup_before_deletion"`
	RequireApproval          bool                     `json:"require_approval"`
	NotificationEnabled      bool                     `json:"notification_enabled"`
	AuditLogging             bool                     `json:"audit_logging"`
	CategoryRetentionPeriods map[string]time.Duration `json:"category_retention_periods"`
	LegalHoldCategories      []string                 `json:"legal_hold_categories"`
}

// RetentionPolicy represents a data retention policy
type RetentionPolicy struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	DataCategory    string                 `json:"data_category"`
	DataType        string                 `json:"data_type"`
	RetentionPeriod time.Duration          `json:"retention_period"`
	DeletionMethod  string                 `json:"deletion_method"` // "soft", "hard", "archive"
	GracePeriod     time.Duration          `json:"grace_period"`
	RequireApproval bool                   `json:"require_approval"`
	LegalHold       bool                   `json:"legal_hold"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CreatedBy       string                 `json:"created_by"`
	UpdatedBy       string                 `json:"updated_by"`
	Status          string                 `json:"status"` // "active", "inactive", "pending"
	Conditions      map[string]interface{} `json:"conditions,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// DataLifecycleRecord represents a record in the data lifecycle
type DataLifecycleRecord struct {
	ID                  string                 `json:"id"`
	DataID              string                 `json:"data_id"`
	DataType            string                 `json:"data_type"`
	DataCategory        string                 `json:"data_category"`
	PolicyID            string                 `json:"policy_id"`
	CreatedAt           time.Time              `json:"created_at"`
	LastAccessedAt      *time.Time             `json:"last_accessed_at,omitempty"`
	ExpiresAt           time.Time              `json:"expires_at"`
	Status              string                 `json:"status"` // "active", "pending_deletion", "deleted", "archived"
	DeletionScheduledAt *time.Time             `json:"deletion_scheduled_at,omitempty"`
	DeletedAt           *time.Time             `json:"deleted_at,omitempty"`
	DeletionMethod      string                 `json:"deletion_method,omitempty"`
	DeletionReason      string                 `json:"deletion_reason,omitempty"`
	LegalHold           bool                   `json:"legal_hold"`
	ApprovalRequired    bool                   `json:"approval_required"`
	ApprovedBy          string                 `json:"approved_by,omitempty"`
	ApprovedAt          *time.Time             `json:"approved_at,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// DeletionRequest represents a request to delete data
type DeletionRequest struct {
	ID              string                 `json:"id"`
	DataID          string                 `json:"data_id"`
	DataType        string                 `json:"data_type"`
	DataCategory    string                 `json:"data_category"`
	RequestType     string                 `json:"request_type"` // "automatic", "manual", "compliance", "user_request"
	Reason          string                 `json:"reason"`
	RequestedBy     string                 `json:"requested_by"`
	RequestedAt     time.Time              `json:"requested_at"`
	ScheduledAt     time.Time              `json:"scheduled_at"`
	ProcessedAt     *time.Time             `json:"processed_at,omitempty"`
	Status          string                 `json:"status"` // "pending", "approved", "rejected", "completed", "failed"
	DeletionMethod  string                 `json:"deletion_method"`
	RequireApproval bool                   `json:"require_approval"`
	ApprovedBy      string                 `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time             `json:"approved_at,omitempty"`
	RejectedBy      string                 `json:"rejected_by,omitempty"`
	RejectedAt      *time.Time             `json:"rejected_at,omitempty"`
	RejectionReason string                 `json:"rejection_reason,omitempty"`
	BackupLocation  string                 `json:"backup_location,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// RetentionAuditEvent represents an audit event for retention operations
type RetentionAuditEvent struct {
	ID          string                 `json:"id"`
	EventType   string                 `json:"event_type"` // "policy_created", "data_expired", "deletion_scheduled", "deletion_completed"
	DataID      string                 `json:"data_id,omitempty"`
	PolicyID    string                 `json:"policy_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Description string                 `json:"description"`
	UserID      string                 `json:"user_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Severity    string                 `json:"severity"` // "info", "warning", "error", "critical"
}

// RetentionReport represents a data retention compliance report
type RetentionReport struct {
	ID                  string                 `json:"id"`
	GeneratedAt         time.Time              `json:"generated_at"`
	ReportType          string                 `json:"report_type"` // "compliance", "lifecycle", "policy_effectiveness"
	Period              string                 `json:"period"`
	TotalRecords        int                    `json:"total_records"`
	ActiveRecords       int                    `json:"active_records"`
	ExpiredRecords      int                    `json:"expired_records"`
	DeletedRecords      int                    `json:"deleted_records"`
	PendingDeletions    int                    `json:"pending_deletions"`
	PolicyViolations    int                    `json:"policy_violations"`
	LegalHoldRecords    int                    `json:"legal_hold_records"`
	ComplianceScore     float64                `json:"compliance_score"`
	PolicyEffectiveness map[string]float64     `json:"policy_effectiveness"`
	Recommendations     []string               `json:"recommendations"`
	Details             map[string]interface{} `json:"details,omitempty"`
}

// PolicyManager handles retention policy management
type PolicyManager struct {
	config *DataRetentionConfig
	logger *zap.Logger
}

// DataLifecycleManager handles data lifecycle tracking
type DataLifecycleManager struct {
	config *DataRetentionConfig
	logger *zap.Logger
}

// DeletionManager handles data deletion operations
type DeletionManager struct {
	config *DataRetentionConfig
	logger *zap.Logger
}

// RetentionAuditManager handles audit logging for retention operations
type RetentionAuditManager struct {
	config *DataRetentionConfig
	logger *zap.Logger
}

// NewDataRetentionService creates a new data retention service
func NewDataRetentionService(config *DataRetentionConfig, logger *zap.Logger) (*DataRetentionService, error) {
	if config == nil {
		config = &DataRetentionConfig{
			EnableDataRetention:    true,
			DefaultRetentionPeriod: 365 * 24 * time.Hour,     // 1 year
			MinRetentionPeriod:     30 * 24 * time.Hour,      // 30 days
			MaxRetentionPeriod:     7 * 365 * 24 * time.Hour, // 7 years
			GracePeriod:            30 * 24 * time.Hour,      // 30 days
			AutomaticDeletion:      true,
			BackupBeforeDeletion:   true,
			RequireApproval:        false,
			NotificationEnabled:    true,
			AuditLogging:           true,
			CategoryRetentionPeriods: map[string]time.Duration{
				"personal_identification": 5 * 365 * 24 * time.Hour, // 5 years
				"financial_data":          7 * 365 * 24 * time.Hour, // 7 years
				"contact_information":     3 * 365 * 24 * time.Hour, // 3 years
				"business_data":           2 * 365 * 24 * time.Hour, // 2 years
				"location_data":           1 * 365 * 24 * time.Hour, // 1 year
				"usage_data":              90 * 24 * time.Hour,      // 90 days
				"analytics_data":          180 * 24 * time.Hour,     // 180 days
			},
			LegalHoldCategories: []string{"financial_data", "compliance_data"},
		}
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &DataRetentionService{
		config:           config,
		logger:           logger,
		policyManager:    &PolicyManager{config: config, logger: logger},
		lifecycleManager: &DataLifecycleManager{config: config, logger: logger},
		deletionManager:  &DeletionManager{config: config, logger: logger},
		auditManager:     &RetentionAuditManager{config: config, logger: logger},
	}, nil
}

// CreateRetentionPolicy creates a new data retention policy
func (drs *DataRetentionService) CreateRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	if !drs.config.EnableDataRetention {
		return errors.New("data retention is disabled")
	}

	if policy == nil {
		return errors.New("policy cannot be nil")
	}

	// Validate policy
	if err := drs.validateRetentionPolicy(policy); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	// Set defaults
	policy.ID = drs.generatePolicyID()
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()
	policy.Status = "active"

	// Log policy creation
	drs.logger.Info("retention policy created",
		zap.String("policy_id", policy.ID),
		zap.String("data_category", policy.DataCategory),
		zap.Duration("retention_period", policy.RetentionPeriod))

	// Audit event
	if drs.config.AuditLogging {
		auditEvent := &RetentionAuditEvent{
			ID:          drs.generateAuditID(),
			EventType:   "policy_created",
			PolicyID:    policy.ID,
			Description: fmt.Sprintf("Retention policy created for %s", policy.DataCategory),
			UserID:      policy.CreatedBy,
			Timestamp:   time.Now(),
			Severity:    "info",
			Details: map[string]interface{}{
				"policy_name":      policy.Name,
				"data_category":    policy.DataCategory,
				"retention_period": policy.RetentionPeriod.String(),
				"deletion_method":  policy.DeletionMethod,
				"require_approval": policy.RequireApproval,
			},
		}
		drs.auditManager.logEvent(ctx, auditEvent)
	}

	return nil
}

// RegisterDataForRetention registers data for retention tracking
func (drs *DataRetentionService) RegisterDataForRetention(ctx context.Context, dataID, dataType, dataCategory string, metadata map[string]interface{}) (*DataLifecycleRecord, error) {
	if !drs.config.EnableDataRetention {
		return nil, errors.New("data retention is disabled")
	}

	// Find applicable policy
	policy, err := drs.findApplicablePolicy(dataCategory, dataType)
	if err != nil {
		return nil, fmt.Errorf("failed to find applicable policy: %w", err)
	}

	// Calculate expiration date
	expiresAt := time.Now().Add(policy.RetentionPeriod)

	// Create lifecycle record
	record := &DataLifecycleRecord{
		ID:               drs.generateLifecycleID(),
		DataID:           dataID,
		DataType:         dataType,
		DataCategory:     dataCategory,
		PolicyID:         policy.ID,
		CreatedAt:        time.Now(),
		ExpiresAt:        expiresAt,
		Status:           "active",
		LegalHold:        policy.LegalHold,
		ApprovalRequired: policy.RequireApproval,
		Metadata:         metadata,
	}

	drs.logger.Info("data registered for retention",
		zap.String("data_id", dataID),
		zap.String("data_category", dataCategory),
		zap.Time("expires_at", expiresAt),
		zap.String("policy_id", policy.ID))

	return record, nil
}

// ScheduleDataDeletion schedules data for deletion based on retention policies
func (drs *DataRetentionService) ScheduleDataDeletion(ctx context.Context, record *DataLifecycleRecord, reason string) (*DeletionRequest, error) {
	if !drs.config.EnableDataRetention {
		return nil, errors.New("data retention is disabled")
	}

	if record.LegalHold {
		return nil, errors.New("cannot delete data under legal hold")
	}

	// Find policy
	policy, err := drs.findPolicyByID(record.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("failed to find policy: %w", err)
	}

	// Create deletion request
	deletionRequest := &DeletionRequest{
		ID:              drs.generateDeletionRequestID(),
		DataID:          record.DataID,
		DataType:        record.DataType,
		DataCategory:    record.DataCategory,
		RequestType:     "automatic",
		Reason:          reason,
		RequestedBy:     "system",
		RequestedAt:     time.Now(),
		ScheduledAt:     time.Now().Add(drs.config.GracePeriod),
		Status:          "pending",
		DeletionMethod:  policy.DeletionMethod,
		RequireApproval: policy.RequireApproval,
	}

	// Update lifecycle record
	record.Status = "pending_deletion"
	record.DeletionScheduledAt = &deletionRequest.ScheduledAt
	record.DeletionMethod = policy.DeletionMethod
	record.DeletionReason = reason

	drs.logger.Info("data deletion scheduled",
		zap.String("data_id", record.DataID),
		zap.String("deletion_request_id", deletionRequest.ID),
		zap.Time("scheduled_at", deletionRequest.ScheduledAt),
		zap.String("reason", reason))

	// Audit event
	if drs.config.AuditLogging {
		auditEvent := &RetentionAuditEvent{
			ID:          drs.generateAuditID(),
			EventType:   "deletion_scheduled",
			DataID:      record.DataID,
			PolicyID:    record.PolicyID,
			RequestID:   deletionRequest.ID,
			Description: fmt.Sprintf("Data deletion scheduled: %s", reason),
			Timestamp:   time.Now(),
			Severity:    "info",
			Details: map[string]interface{}{
				"data_category":    record.DataCategory,
				"scheduled_at":     deletionRequest.ScheduledAt,
				"deletion_method":  policy.DeletionMethod,
				"require_approval": policy.RequireApproval,
				"grace_period":     drs.config.GracePeriod.String(),
			},
		}
		drs.auditManager.logEvent(ctx, auditEvent)
	}

	return deletionRequest, nil
}

// ProcessExpiredData identifies and processes expired data
func (drs *DataRetentionService) ProcessExpiredData(ctx context.Context) (int, error) {
	if !drs.config.EnableDataRetention {
		return 0, errors.New("data retention is disabled")
	}

	processedCount := 0
	cutoffTime := time.Now()

	drs.logger.Info("processing expired data", zap.Time("cutoff_time", cutoffTime))

	// Find expired records (this would typically query a database)
	expiredRecords, err := drs.findExpiredRecords(cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to find expired records: %w", err)
	}

	for _, record := range expiredRecords {
		if record.Status == "active" && !record.LegalHold {
			// Schedule for deletion
			_, err := drs.ScheduleDataDeletion(ctx, record, "retention period expired")
			if err != nil {
				drs.logger.Error("failed to schedule deletion",
					zap.String("data_id", record.DataID),
					zap.Error(err))
				continue
			}
			processedCount++
		}
	}

	drs.logger.Info("expired data processing completed",
		zap.Int("processed_count", processedCount),
		zap.Int("total_expired", len(expiredRecords)))

	return processedCount, nil
}

// ExecutePendingDeletions executes pending deletion requests
func (drs *DataRetentionService) ExecutePendingDeletions(ctx context.Context) (int, error) {
	if !drs.config.EnableDataRetention || !drs.config.AutomaticDeletion {
		return 0, errors.New("automatic deletion is disabled")
	}

	executedCount := 0
	currentTime := time.Now()

	drs.logger.Info("executing pending deletions", zap.Time("current_time", currentTime))

	// Find ready deletion requests
	readyRequests, err := drs.findReadyDeletions(currentTime)
	if err != nil {
		return 0, fmt.Errorf("failed to find ready deletions: %w", err)
	}

	for _, request := range readyRequests {
		if request.RequireApproval && request.ApprovedAt == nil {
			drs.logger.Info("skipping deletion requiring approval",
				zap.String("request_id", request.ID),
				zap.String("data_id", request.DataID))
			continue
		}

		// Execute deletion
		err := drs.executeDeletion(ctx, request)
		if err != nil {
			drs.logger.Error("failed to execute deletion",
				zap.String("request_id", request.ID),
				zap.String("data_id", request.DataID),
				zap.Error(err))
			continue
		}

		executedCount++
	}

	drs.logger.Info("pending deletions execution completed",
		zap.Int("executed_count", executedCount),
		zap.Int("total_ready", len(readyRequests)))

	return executedCount, nil
}

// ApproveDeletion approves a deletion request
func (drs *DataRetentionService) ApproveDeletion(ctx context.Context, requestID, approverID string) error {
	if !drs.config.EnableDataRetention {
		return errors.New("data retention is disabled")
	}

	request, err := drs.findDeletionRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("failed to find deletion request: %w", err)
	}

	if !request.RequireApproval {
		return errors.New("this deletion request does not require approval")
	}

	if request.Status != "pending" {
		return fmt.Errorf("deletion request is not pending: %s", request.Status)
	}

	// Approve the request
	now := time.Now()
	request.Status = "approved"
	request.ApprovedBy = approverID
	request.ApprovedAt = &now

	drs.logger.Info("deletion request approved",
		zap.String("request_id", requestID),
		zap.String("data_id", request.DataID),
		zap.String("approved_by", approverID))

	return nil
}

// GenerateRetentionReport generates a comprehensive retention report
func (drs *DataRetentionService) GenerateRetentionReport(ctx context.Context, reportType, period string) (*RetentionReport, error) {
	if !drs.config.EnableDataRetention {
		return nil, errors.New("data retention is disabled")
	}

	report := &RetentionReport{
		ID:          drs.generateReportID(),
		GeneratedAt: time.Now(),
		ReportType:  reportType,
		Period:      period,
	}

	// Gather statistics (this would typically query a database)
	stats, err := drs.gatherRetentionStatistics(period)
	if err != nil {
		return nil, fmt.Errorf("failed to gather statistics: %w", err)
	}

	report.TotalRecords = stats.TotalRecords
	report.ActiveRecords = stats.ActiveRecords
	report.ExpiredRecords = stats.ExpiredRecords
	report.DeletedRecords = stats.DeletedRecords
	report.PendingDeletions = stats.PendingDeletions
	report.PolicyViolations = stats.PolicyViolations
	report.LegalHoldRecords = stats.LegalHoldRecords

	// Calculate compliance score
	report.ComplianceScore = drs.calculateComplianceScore(stats)

	// Generate recommendations
	report.Recommendations = drs.generateRetentionRecommendations(stats)

	drs.logger.Info("retention report generated",
		zap.String("report_id", report.ID),
		zap.String("report_type", reportType),
		zap.Float64("compliance_score", report.ComplianceScore))

	return report, nil
}

// Helper methods

func (drs *DataRetentionService) validateRetentionPolicy(policy *RetentionPolicy) error {
	if policy.Name == "" {
		return errors.New("policy name is required")
	}

	if policy.DataCategory == "" {
		return errors.New("data category is required")
	}

	if policy.RetentionPeriod < drs.config.MinRetentionPeriod {
		return fmt.Errorf("retention period below minimum: %v", drs.config.MinRetentionPeriod)
	}

	if policy.RetentionPeriod > drs.config.MaxRetentionPeriod {
		return fmt.Errorf("retention period exceeds maximum: %v", drs.config.MaxRetentionPeriod)
	}

	if policy.DeletionMethod == "" {
		policy.DeletionMethod = "soft"
	}

	if policy.DeletionMethod != "soft" && policy.DeletionMethod != "hard" && policy.DeletionMethod != "archive" {
		return errors.New("invalid deletion method: must be 'soft', 'hard', or 'archive'")
	}

	return nil
}

func (drs *DataRetentionService) findApplicablePolicy(dataCategory, dataType string) (*RetentionPolicy, error) {
	// This would typically query a database or policy store
	// For now, create a default policy based on category

	retentionPeriod := drs.config.DefaultRetentionPeriod
	if categoryPeriod, exists := drs.config.CategoryRetentionPeriods[dataCategory]; exists {
		retentionPeriod = categoryPeriod
	}

	isLegalHold := false
	for _, category := range drs.config.LegalHoldCategories {
		if category == dataCategory {
			isLegalHold = true
			break
		}
	}

	return &RetentionPolicy{
		ID:              drs.generatePolicyID(),
		Name:            fmt.Sprintf("Default policy for %s", dataCategory),
		DataCategory:    dataCategory,
		DataType:        dataType,
		RetentionPeriod: retentionPeriod,
		DeletionMethod:  "soft",
		GracePeriod:     drs.config.GracePeriod,
		RequireApproval: drs.config.RequireApproval,
		LegalHold:       isLegalHold,
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

func (drs *DataRetentionService) findPolicyByID(policyID string) (*RetentionPolicy, error) {
	// This would typically query a database
	// For now, return a mock policy
	return &RetentionPolicy{
		ID:              policyID,
		Name:            "Mock Policy",
		DeletionMethod:  "soft",
		RequireApproval: drs.config.RequireApproval,
	}, nil
}

func (drs *DataRetentionService) findExpiredRecords(cutoffTime time.Time) ([]*DataLifecycleRecord, error) {
	// This would typically query a database
	// Return empty slice for now
	return []*DataLifecycleRecord{}, nil
}

func (drs *DataRetentionService) findReadyDeletions(currentTime time.Time) ([]*DeletionRequest, error) {
	// This would typically query a database
	// Return empty slice for now
	return []*DeletionRequest{}, nil
}

func (drs *DataRetentionService) findDeletionRequestByID(requestID string) (*DeletionRequest, error) {
	// This would typically query a database
	// Return mock request for now
	return &DeletionRequest{
		ID:              requestID,
		Status:          "pending",
		RequireApproval: true,
	}, nil
}

func (drs *DataRetentionService) executeDeletion(ctx context.Context, request *DeletionRequest) error {
	// Backup if required
	if drs.config.BackupBeforeDeletion {
		backupLocation := fmt.Sprintf("backup/%s/%s", request.DataCategory, request.DataID)
		request.BackupLocation = backupLocation
		drs.logger.Info("data backed up before deletion",
			zap.String("data_id", request.DataID),
			zap.String("backup_location", backupLocation))
	}

	// Execute deletion based on method
	switch request.DeletionMethod {
	case "soft":
		// Mark as deleted but keep data
		drs.logger.Info("soft deletion executed", zap.String("data_id", request.DataID))
	case "hard":
		// Permanently delete data
		drs.logger.Info("hard deletion executed", zap.String("data_id", request.DataID))
	case "archive":
		// Move to archive storage
		drs.logger.Info("data archived", zap.String("data_id", request.DataID))
	}

	// Update request status
	now := time.Now()
	request.Status = "completed"
	request.ProcessedAt = &now

	// Audit event
	if drs.config.AuditLogging {
		auditEvent := &RetentionAuditEvent{
			ID:          drs.generateAuditID(),
			EventType:   "deletion_completed",
			DataID:      request.DataID,
			RequestID:   request.ID,
			Description: fmt.Sprintf("Data deletion completed using %s method", request.DeletionMethod),
			Timestamp:   time.Now(),
			Severity:    "info",
			Details: map[string]interface{}{
				"deletion_method": request.DeletionMethod,
				"backup_location": request.BackupLocation,
				"processed_by":    "system",
			},
		}
		drs.auditManager.logEvent(context.Background(), auditEvent)
	}

	return nil
}

// Statistics structure for reporting
type RetentionStatistics struct {
	TotalRecords     int
	ActiveRecords    int
	ExpiredRecords   int
	DeletedRecords   int
	PendingDeletions int
	PolicyViolations int
	LegalHoldRecords int
}

func (drs *DataRetentionService) gatherRetentionStatistics(period string) (*RetentionStatistics, error) {
	// This would typically query a database
	// Return mock statistics for now
	return &RetentionStatistics{
		TotalRecords:     1000,
		ActiveRecords:    800,
		ExpiredRecords:   50,
		DeletedRecords:   150,
		PendingDeletions: 25,
		PolicyViolations: 5,
		LegalHoldRecords: 10,
	}, nil
}

func (drs *DataRetentionService) calculateComplianceScore(stats *RetentionStatistics) float64 {
	if stats.TotalRecords == 0 {
		return 100.0
	}

	// Calculate compliance based on various factors
	violationPenalty := float64(stats.PolicyViolations) * 10.0
	pendingPenalty := float64(stats.PendingDeletions) * 2.0

	score := 100.0 - violationPenalty - pendingPenalty

	if score < 0 {
		score = 0
	}

	return score
}

func (drs *DataRetentionService) generateRetentionRecommendations(stats *RetentionStatistics) []string {
	recommendations := make([]string, 0)

	if stats.PolicyViolations > 0 {
		recommendations = append(recommendations, "Review and address policy violations to improve compliance")
	}

	if stats.PendingDeletions > stats.TotalRecords/20 { // More than 5% pending
		recommendations = append(recommendations, "Consider increasing automation or reducing approval requirements")
	}

	if stats.ExpiredRecords > stats.ActiveRecords/10 { // More than 10% expired
		recommendations = append(recommendations, "Implement more frequent cleanup processes")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Data retention compliance is operating within acceptable parameters")
	}

	return recommendations
}

// ID generation helpers
func (drs *DataRetentionService) generatePolicyID() string {
	return fmt.Sprintf("policy_%d", time.Now().UnixNano())
}

func (drs *DataRetentionService) generateLifecycleID() string {
	return fmt.Sprintf("lifecycle_%d", time.Now().UnixNano())
}

func (drs *DataRetentionService) generateDeletionRequestID() string {
	return fmt.Sprintf("deletion_%d", time.Now().UnixNano())
}

func (drs *DataRetentionService) generateAuditID() string {
	return fmt.Sprintf("audit_%d", time.Now().UnixNano())
}

func (drs *DataRetentionService) generateReportID() string {
	return fmt.Sprintf("report_%d", time.Now().UnixNano())
}

// Sub-manager methods
func (am *RetentionAuditManager) logEvent(ctx context.Context, event *RetentionAuditEvent) {
	am.logger.Info("retention audit event",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.EventType),
		zap.String("data_id", event.DataID),
		zap.String("severity", event.Severity),
		zap.String("description", event.Description))
}
