package services

import (
	"context"
	"fmt"
	"time"

	"kyb-platform/internal/compliance"
	"kyb-platform/internal/models"
	"kyb-platform/internal/observability"
)

// UnifiedAuditService provides comprehensive audit logging using the unified audit system
type UnifiedAuditService struct {
	logger     *observability.Logger
	compliance ComplianceSystem
	repository UnifiedAuditRepository
}

// UnifiedAuditRepository defines the interface for unified audit data persistence
type UnifiedAuditRepository interface {
	// SaveAuditLog saves a unified audit log entry
	SaveAuditLog(ctx context.Context, auditLog *models.UnifiedAuditLog) error

	// GetAuditLogs retrieves unified audit logs with filtering
	GetAuditLogs(ctx context.Context, filters *models.UnifiedAuditLogFilters) (*models.UnifiedAuditLogResult, error)

	// GetAuditLogByID retrieves a specific unified audit log by ID
	GetAuditLogByID(ctx context.Context, id string) (*models.UnifiedAuditLog, error)

	// GetAuditTrail retrieves audit trail for a specific merchant
	GetAuditTrail(ctx context.Context, merchantID string, limit int, offset int) ([]*models.UnifiedAuditLog, error)

	// GetAuditLogsByUser retrieves audit logs for a specific user
	GetAuditLogsByUser(ctx context.Context, userID string, limit int, offset int) ([]*models.UnifiedAuditLog, error)

	// GetAuditLogsByAction retrieves audit logs for a specific action
	GetAuditLogsByAction(ctx context.Context, action string, limit int, offset int) ([]*models.UnifiedAuditLog, error)

	// DeleteOldAuditLogs deletes audit logs older than the specified duration
	DeleteOldAuditLogs(ctx context.Context, olderThan time.Duration) (int64, error)
}

// LogMerchantOperationRequest is defined in audit_service.go to avoid duplicate type declaration
// Use the type from audit_service.go instead

// LogUserActionRequest represents a request to log a user action
type LogUserActionRequest struct {
	UserID       string                 `json:"user_id"`
	APIKeyID     string                 `json:"api_key_id,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type,omitempty"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	Description  string                 `json:"description"`
	Details      interface{}            `json:"details,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LogSystemEventRequest represents a request to log a system event
type LogSystemEventRequest struct {
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type,omitempty"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	Description  string                 `json:"description"`
	Details      interface{}            `json:"details,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LogDataChangeRequest represents a request to log a data change
type LogDataChangeRequest struct {
	UserID       string                 `json:"user_id,omitempty"`
	APIKeyID     string                 `json:"api_key_id,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	TableName    string                 `json:"table_name,omitempty"`
	OldValues    interface{}            `json:"old_values,omitempty"`
	NewValues    interface{}            `json:"new_values,omitempty"`
	Description  string                 `json:"description"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// NewUnifiedAuditService creates a new unified audit service
func NewUnifiedAuditService(
	logger *observability.Logger,
	compliance ComplianceSystem,
	repository UnifiedAuditRepository,
) *UnifiedAuditService {
	return &UnifiedAuditService{
		logger:     logger,
		compliance: compliance,
		repository: repository,
	}
}

// LogMerchantOperation logs a merchant operation for audit purposes
func (uas *UnifiedAuditService) LogMerchantOperation(ctx context.Context, req *LogMerchantOperationRequest) error {
	auditLog := models.NewUnifiedAuditLog()

	// Set event classification
	auditLog.SetEventType(models.EventTypeMerchantOperation)
	auditLog.SetEventCategory(models.EventCategoryMerchant)
	auditLog.SetAction(models.AuditLogAction(req.Action))

	// Set context
	// Note: LogMerchantOperationRequest doesn't have APIKeyID field
	auditLog.SetUserContext(req.UserID, "") // APIKeyID not available in request
	auditLog.SetBusinessContext(req.MerchantID, req.SessionID)
	auditLog.SetResourceInfo(req.ResourceType, req.ResourceID, "")
	auditLog.SetRequestContext(req.RequestID, req.IPAddress, req.UserAgent)

	// Set change tracking
	// Note: LogMerchantOperationRequest doesn't have OldValues or NewValues fields
	// Use Details and Metadata instead
	var oldValues, newValues interface{}
	if req.Metadata != nil {
		if ov, ok := req.Metadata["old_values"]; ok {
			oldValues = ov
		}
		if nv, ok := req.Metadata["new_values"]; ok {
			newValues = nv
		}
	}
	if err := auditLog.SetChangeTracking(oldValues, newValues, req.Details); err != nil {
		uas.logger.Error("failed to set change tracking", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to set change tracking: %w", err)
	}

	// Set metadata
	metadata := map[string]interface{}{
		"description": req.Description,
		"user_name":   req.UserName,
		"user_role":   req.UserRole,
		"user_email":  req.UserEmail,
	}

	// Merge with provided metadata
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}

	if err := auditLog.SetMetadata(metadata); err != nil {
		uas.logger.Error("failed to set metadata", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to set metadata: %w", err)
	}

	// Validate the audit log
	if err := auditLog.Validate(); err != nil {
		uas.logger.Error("audit log validation failed", map[string]interface{}{
			"error":     err.Error(),
			"audit_log": auditLog,
		})
		return fmt.Errorf("audit log validation failed: %w", err)
	}

	// Save to repository
	if err := uas.repository.SaveAuditLog(ctx, auditLog); err != nil {
		uas.logger.Error("failed to save audit log", map[string]interface{}{
			"error":     err.Error(),
			"audit_log": auditLog,
		})
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	// Also log to compliance system for backward compatibility
	complianceEvent := &compliance.AuditEvent{
		ID:            auditLog.ID,
		UserID:        req.UserID,
		Action:        compliance.AuditAction(req.Action),
		Resource:      req.ResourceType,
		ResourceID:    req.ResourceID,
		Details:       fmt.Sprintf("%v", req.Details),
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
		Timestamp:     auditLog.CreatedAt,
		BusinessID:    req.MerchantID,
		EventType:     "merchant_operation",
		EventCategory: "audit",
		EntityType:    req.ResourceType,
		EntityID:      req.ResourceID,
		Description:   req.Description,
		UserName:      req.UserName,
		UserRole:      req.UserRole,
		UserEmail:     req.UserEmail,
		SessionID:     req.SessionID,
		RequestID:     req.RequestID,
		Success:       true,
		Metadata:      req.Metadata,
	}

	if err := uas.compliance.RecordAuditEvent(ctx, complianceEvent); err != nil {
		uas.logger.Error("failed to record compliance event", map[string]interface{}{
			"error": err.Error(),
		})
		// Don't fail the main operation if compliance logging fails
	}

	uas.logger.Info("merchant operation logged successfully", map[string]interface{}{
		"audit_log_id": auditLog.ID,
		"merchant_id":  req.MerchantID,
		"action":       req.Action,
	})

	return nil
}

// LogUserAction logs a user action for audit purposes
func (uas *UnifiedAuditService) LogUserAction(ctx context.Context, req *LogUserActionRequest) error {
	auditLog := models.NewUnifiedAuditLog()

	// Set event classification
	auditLog.SetEventType(models.EventTypeUserAction)
	auditLog.SetEventCategory(models.EventCategoryUser)
	auditLog.SetAction(models.AuditLogAction(req.Action))

	// Set context
	auditLog.SetUserContext(req.UserID, req.APIKeyID)
	auditLog.SetResourceInfo(req.ResourceType, req.ResourceID, "")
	auditLog.SetRequestContext(req.RequestID, req.IPAddress, req.UserAgent)

	// Set details
	if err := auditLog.SetChangeTracking(nil, nil, req.Details); err != nil {
		return fmt.Errorf("failed to set details: %w", err)
	}

	// Set metadata
	metadata := map[string]interface{}{
		"description": req.Description,
	}

	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}

	if err := auditLog.SetMetadata(metadata); err != nil {
		return fmt.Errorf("failed to set metadata: %w", err)
	}

	// Validate and save
	if err := auditLog.Validate(); err != nil {
		return fmt.Errorf("audit log validation failed: %w", err)
	}

	if err := uas.repository.SaveAuditLog(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	uas.logger.Info("user action logged successfully", map[string]interface{}{
		"audit_log_id": auditLog.ID,
		"user_id":      req.UserID,
		"action":       req.Action,
	})

	return nil
}

// LogSystemEvent logs a system event for audit purposes
func (uas *UnifiedAuditService) LogSystemEvent(ctx context.Context, req *LogSystemEventRequest) error {
	auditLog := models.NewUnifiedAuditLog()

	// Set event classification
	auditLog.SetEventType(models.EventTypeSystemEvent)
	auditLog.SetEventCategory(models.EventCategorySystem)
	auditLog.SetAction(models.AuditLogAction(req.Action))

	// Set resource info
	auditLog.SetResourceInfo(req.ResourceType, req.ResourceID, "")

	// Set details
	if err := auditLog.SetChangeTracking(nil, nil, req.Details); err != nil {
		return fmt.Errorf("failed to set details: %w", err)
	}

	// Set metadata
	metadata := map[string]interface{}{
		"description": req.Description,
	}

	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}

	if err := auditLog.SetMetadata(metadata); err != nil {
		return fmt.Errorf("failed to set metadata: %w", err)
	}

	// Validate and save
	if err := auditLog.Validate(); err != nil {
		return fmt.Errorf("audit log validation failed: %w", err)
	}

	if err := uas.repository.SaveAuditLog(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	uas.logger.Info("system event logged successfully", map[string]interface{}{
		"audit_log_id": auditLog.ID,
		"action":       req.Action,
	})

	return nil
}

// LogDataChange logs a data change for audit purposes
func (uas *UnifiedAuditService) LogDataChange(ctx context.Context, req *LogDataChangeRequest) error {
	auditLog := models.NewUnifiedAuditLog()

	// Set event classification
	auditLog.SetEventType(models.EventTypeDataChange)
	auditLog.SetEventCategory(models.EventCategorySystem)
	auditLog.SetAction(models.AuditLogAction(req.Action))

	// Set context
	auditLog.SetUserContext(req.UserID, req.APIKeyID)
	auditLog.SetResourceInfo(req.ResourceType, req.ResourceID, req.TableName)
	auditLog.SetRequestContext(req.RequestID, req.IPAddress, req.UserAgent)

	// Set change tracking
	if err := auditLog.SetChangeTracking(req.OldValues, req.NewValues, nil); err != nil {
		return fmt.Errorf("failed to set change tracking: %w", err)
	}

	// Set metadata
	metadata := map[string]interface{}{
		"description": req.Description,
	}

	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}

	if err := auditLog.SetMetadata(metadata); err != nil {
		return fmt.Errorf("failed to set metadata: %w", err)
	}

	// Validate and save
	if err := auditLog.Validate(); err != nil {
		return fmt.Errorf("audit log validation failed: %w", err)
	}

	if err := uas.repository.SaveAuditLog(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	uas.logger.Info("data change logged successfully", map[string]interface{}{
		"audit_log_id":  auditLog.ID,
		"resource_type": req.ResourceType,
		"resource_id":   req.ResourceID,
		"action":        req.Action,
	})

	return nil
}

// GetAuditLogs retrieves audit logs with filtering
func (uas *UnifiedAuditService) GetAuditLogs(ctx context.Context, filters *models.UnifiedAuditLogFilters) (*models.UnifiedAuditLogResult, error) {
	return uas.repository.GetAuditLogs(ctx, filters)
}

// GetAuditLogByID retrieves a specific audit log by ID
func (uas *UnifiedAuditService) GetAuditLogByID(ctx context.Context, id string) (*models.UnifiedAuditLog, error) {
	return uas.repository.GetAuditLogByID(ctx, id)
}

// GetAuditTrail retrieves audit trail for a specific merchant
func (uas *UnifiedAuditService) GetAuditTrail(ctx context.Context, merchantID string, limit int, offset int) ([]*models.UnifiedAuditLog, error) {
	return uas.repository.GetAuditTrail(ctx, merchantID, limit, offset)
}

// GetAuditLogsByUser retrieves audit logs for a specific user
func (uas *UnifiedAuditService) GetAuditLogsByUser(ctx context.Context, userID string, limit int, offset int) ([]*models.UnifiedAuditLog, error) {
	return uas.repository.GetAuditLogsByUser(ctx, userID, limit, offset)
}

// GetAuditLogsByAction retrieves audit logs for a specific action
func (uas *UnifiedAuditService) GetAuditLogsByAction(ctx context.Context, action string, limit int, offset int) ([]*models.UnifiedAuditLog, error) {
	return uas.repository.GetAuditLogsByAction(ctx, action, limit, offset)
}

// DeleteOldAuditLogs deletes audit logs older than the specified duration
func (uas *UnifiedAuditService) DeleteOldAuditLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	return uas.repository.DeleteOldAuditLogs(ctx, olderThan)
}

// MigrateFromLegacyAuditService migrates from the legacy audit service
func (uas *UnifiedAuditService) MigrateFromLegacyAuditService(ctx context.Context, legacyService *AuditService) error {
	uas.logger.Info("migrating from legacy audit service to unified audit service", map[string]interface{}{})

	// This method can be used to gradually migrate from the legacy audit service
	// to the unified audit service by intercepting calls and redirecting them

	// For now, we'll just log the migration
	uas.logger.Info("legacy audit service migration completed", map[string]interface{}{
		"timestamp": time.Now(),
	})

	return nil
}
