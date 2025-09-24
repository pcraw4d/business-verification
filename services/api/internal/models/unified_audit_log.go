package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UnifiedAuditLog represents a comprehensive audit log entry in the unified audit system
type UnifiedAuditLog struct {
	// Primary Key
	ID string `json:"id" db:"id"`

	// User and Authentication Context
	UserID   *string `json:"user_id,omitempty" db:"user_id"`
	APIKeyID *string `json:"api_key_id,omitempty" db:"api_key_id"`

	// Business Context
	MerchantID *string `json:"merchant_id,omitempty" db:"merchant_id"`
	SessionID  *string `json:"session_id,omitempty" db:"session_id"`

	// Event Classification
	EventType     string `json:"event_type" db:"event_type"`
	EventCategory string `json:"event_category" db:"event_category"`
	Action        string `json:"action" db:"action"`

	// Resource Information
	ResourceType *string `json:"resource_type,omitempty" db:"resource_type"`
	ResourceID   *string `json:"resource_id,omitempty" db:"resource_id"`
	TableName    *string `json:"table_name,omitempty" db:"table_name"`

	// Change Tracking
	OldValues *json.RawMessage `json:"old_values,omitempty" db:"old_values"`
	NewValues *json.RawMessage `json:"new_values,omitempty" db:"new_values"`
	Details   *json.RawMessage `json:"details,omitempty" db:"details"`

	// Request Context
	RequestID *string `json:"request_id,omitempty" db:"request_id"`
	IPAddress *string `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string `json:"user_agent,omitempty" db:"user_agent"`

	// Metadata and Timestamps
	Metadata  *json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}

// AuditLogEventType represents the type of audit event
type AuditLogEventType string

const (
	EventTypeUserAction        AuditLogEventType = "user_action"
	EventTypeSystemEvent       AuditLogEventType = "system_event"
	EventTypeAPICall           AuditLogEventType = "api_call"
	EventTypeDataChange        AuditLogEventType = "data_change"
	EventTypeSecurityEvent     AuditLogEventType = "security_event"
	EventTypeComplianceCheck   AuditLogEventType = "compliance_check"
	EventTypeBusinessOperation AuditLogEventType = "business_operation"
	EventTypeMerchantOperation AuditLogEventType = "merchant_operation"
	EventTypeClassification    AuditLogEventType = "classification"
	EventTypeRiskAssessment    AuditLogEventType = "risk_assessment"
	EventTypeVerification      AuditLogEventType = "verification"
	EventTypeAuthentication    AuditLogEventType = "authentication"
	EventTypeAuthorization     AuditLogEventType = "authorization"
)

// AuditLogEventCategory represents the category of audit event
type AuditLogEventCategory string

const (
	EventCategoryAudit      AuditLogEventCategory = "audit"
	EventCategoryCompliance AuditLogEventCategory = "compliance"
	EventCategorySecurity   AuditLogEventCategory = "security"
	EventCategoryBusiness   AuditLogEventCategory = "business"
	EventCategorySystem     AuditLogEventCategory = "system"
	EventCategoryUser       AuditLogEventCategory = "user"
	EventCategoryMerchant   AuditLogEventCategory = "merchant"
)

// AuditLogAction represents the action performed
type AuditLogAction string

const (
	ActionInsert   AuditLogAction = "INSERT"
	ActionUpdate   AuditLogAction = "UPDATE"
	ActionDelete   AuditLogAction = "DELETE"
	ActionCreate   AuditLogAction = "CREATE"
	ActionRead     AuditLogAction = "READ"
	ActionLogin    AuditLogAction = "LOGIN"
	ActionLogout   AuditLogAction = "LOGOUT"
	ActionAccess   AuditLogAction = "ACCESS"
	ActionExport   AuditLogAction = "EXPORT"
	ActionImport   AuditLogAction = "IMPORT"
	ActionVerify   AuditLogAction = "VERIFY"
	ActionApprove  AuditLogAction = "APPROVE"
	ActionReject   AuditLogAction = "REJECT"
	ActionClassify AuditLogAction = "CLASSIFY"
	ActionAssess   AuditLogAction = "ASSESS"
	ActionScan     AuditLogAction = "SCAN"
	ActionAnalyze  AuditLogAction = "ANALYZE"
)

// NewUnifiedAuditLog creates a new unified audit log entry with default values
func NewUnifiedAuditLog() *UnifiedAuditLog {
	return &UnifiedAuditLog{
		ID:            uuid.New().String(),
		EventCategory: string(EventCategoryAudit),
		CreatedAt:     time.Now(),
	}
}

// SetEventType sets the event type with validation
func (ual *UnifiedAuditLog) SetEventType(eventType AuditLogEventType) error {
	validTypes := []AuditLogEventType{
		EventTypeUserAction, EventTypeSystemEvent, EventTypeAPICall,
		EventTypeDataChange, EventTypeSecurityEvent, EventTypeComplianceCheck,
		EventTypeBusinessOperation, EventTypeMerchantOperation, EventTypeClassification,
		EventTypeRiskAssessment, EventTypeVerification, EventTypeAuthentication,
		EventTypeAuthorization,
	}

	for _, validType := range validTypes {
		if eventType == validType {
			ual.EventType = string(eventType)
			return nil
		}
	}

	return fmt.Errorf("invalid event type: %s", eventType)
}

// SetEventCategory sets the event category with validation
func (ual *UnifiedAuditLog) SetEventCategory(category AuditLogEventCategory) error {
	validCategories := []AuditLogEventCategory{
		EventCategoryAudit, EventCategoryCompliance, EventCategorySecurity,
		EventCategoryBusiness, EventCategorySystem, EventCategoryUser,
		EventCategoryMerchant,
	}

	for _, validCategory := range validCategories {
		if category == validCategory {
			ual.EventCategory = string(category)
			return nil
		}
	}

	return fmt.Errorf("invalid event category: %s", category)
}

// SetAction sets the action with validation
func (ual *UnifiedAuditLog) SetAction(action AuditLogAction) error {
	validActions := []AuditLogAction{
		ActionInsert, ActionUpdate, ActionDelete, ActionCreate, ActionRead,
		ActionLogin, ActionLogout, ActionAccess, ActionExport, ActionImport,
		ActionVerify, ActionApprove, ActionReject, ActionClassify, ActionAssess,
		ActionScan, ActionAnalyze,
	}

	for _, validAction := range validActions {
		if action == validAction {
			ual.Action = string(action)
			return nil
		}
	}

	return fmt.Errorf("invalid action: %s", action)
}

// SetUserContext sets the user context for the audit log
func (ual *UnifiedAuditLog) SetUserContext(userID, apiKeyID string) {
	if userID != "" {
		ual.UserID = &userID
	}
	if apiKeyID != "" {
		ual.APIKeyID = &apiKeyID
	}
}

// SetBusinessContext sets the business context for the audit log
func (ual *UnifiedAuditLog) SetBusinessContext(merchantID, sessionID string) {
	if merchantID != "" {
		ual.MerchantID = &merchantID
	}
	if sessionID != "" {
		ual.SessionID = &sessionID
	}
}

// SetResourceInfo sets the resource information for the audit log
func (ual *UnifiedAuditLog) SetResourceInfo(resourceType, resourceID, tableName string) {
	if resourceType != "" {
		ual.ResourceType = &resourceType
	}
	if resourceID != "" {
		ual.ResourceID = &resourceID
	}
	if tableName != "" {
		ual.TableName = &tableName
	}
}

// SetRequestContext sets the request context for the audit log
func (ual *UnifiedAuditLog) SetRequestContext(requestID, ipAddress, userAgent string) {
	if requestID != "" {
		ual.RequestID = &requestID
	}
	if ipAddress != "" {
		ual.IPAddress = &ipAddress
	}
	if userAgent != "" {
		ual.UserAgent = &userAgent
	}
}

// SetChangeTracking sets the change tracking data for the audit log
func (ual *UnifiedAuditLog) SetChangeTracking(oldValues, newValues, details interface{}) error {
	if oldValues != nil {
		oldData, err := json.Marshal(oldValues)
		if err != nil {
			return fmt.Errorf("failed to marshal old values: %w", err)
		}
		ual.OldValues = (*json.RawMessage)(&oldData)
	}

	if newValues != nil {
		newData, err := json.Marshal(newValues)
		if err != nil {
			return fmt.Errorf("failed to marshal new values: %w", err)
		}
		ual.NewValues = (*json.RawMessage)(&newData)
	}

	if details != nil {
		detailsData, err := json.Marshal(details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
		ual.Details = (*json.RawMessage)(&detailsData)
	}

	return nil
}

// SetMetadata sets the metadata for the audit log
func (ual *UnifiedAuditLog) SetMetadata(metadata interface{}) error {
	if metadata != nil {
		metadataData, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		ual.Metadata = (*json.RawMessage)(&metadataData)
	}
	return nil
}

// Validate validates the unified audit log entry
func (ual *UnifiedAuditLog) Validate() error {
	if ual.ID == "" {
		return fmt.Errorf("ID is required")
	}

	if ual.EventType == "" {
		return fmt.Errorf("event type is required")
	}

	if ual.EventCategory == "" {
		return fmt.Errorf("event category is required")
	}

	if ual.Action == "" {
		return fmt.Errorf("action is required")
	}

	// Validate event type
	if err := ual.SetEventType(AuditLogEventType(ual.EventType)); err != nil {
		return err
	}

	// Validate event category
	if err := ual.SetEventCategory(AuditLogEventCategory(ual.EventCategory)); err != nil {
		return err
	}

	// Validate action
	if err := ual.SetAction(AuditLogAction(ual.Action)); err != nil {
		return err
	}

	return nil
}

// ToLegacyAuditLog converts the unified audit log to the legacy AuditLog format
// for backward compatibility during migration
func (ual *UnifiedAuditLog) ToLegacyAuditLog() *AuditLog {
	legacy := &AuditLog{
		ID:           ual.ID,
		Action:       ual.Action,
		ResourceType: getStringValue(ual.ResourceType),
		ResourceID:   getStringValue(ual.ResourceID),
		IPAddress:    getStringValue(ual.IPAddress),
		UserAgent:    getStringValue(ual.UserAgent),
		RequestID:    getStringValue(ual.RequestID),
		CreatedAt:    ual.CreatedAt,
	}

	if ual.UserID != nil {
		legacy.UserID = *ual.UserID
	}

	if ual.Details != nil {
		legacy.Details = string(*ual.Details)
	}

	return legacy
}

// FromLegacyAuditLog creates a unified audit log from a legacy AuditLog
func FromLegacyAuditLog(legacy *AuditLog) *UnifiedAuditLog {
	ual := NewUnifiedAuditLog()
	ual.ID = legacy.ID
	ual.Action = legacy.Action
	ual.CreatedAt = legacy.CreatedAt

	if legacy.UserID != "" {
		ual.UserID = &legacy.UserID
	}

	if legacy.ResourceType != "" {
		ual.ResourceType = &legacy.ResourceType
	}

	if legacy.ResourceID != "" {
		ual.ResourceID = &legacy.ResourceID
	}

	if legacy.IPAddress != "" {
		ual.IPAddress = &legacy.IPAddress
	}

	if legacy.UserAgent != "" {
		ual.UserAgent = &legacy.UserAgent
	}

	if legacy.RequestID != "" {
		ual.RequestID = &legacy.RequestID
	}

	if legacy.Details != "" {
		detailsData := json.RawMessage(legacy.Details)
		ual.Details = &detailsData
	}

	// Set default values
	ual.EventType = string(EventTypeUserAction)
	ual.EventCategory = string(EventCategoryAudit)

	return ual
}

// getStringValue safely gets a string value from a pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// UnifiedAuditLogFilters represents filters for unified audit log queries
type UnifiedAuditLogFilters struct {
	UserID        *string    `json:"user_id,omitempty"`
	APIKeyID      *string    `json:"api_key_id,omitempty"`
	MerchantID    *string    `json:"merchant_id,omitempty"`
	SessionID     *string    `json:"session_id,omitempty"`
	EventType     *string    `json:"event_type,omitempty"`
	EventCategory *string    `json:"event_category,omitempty"`
	Action        *string    `json:"action,omitempty"`
	ResourceType  *string    `json:"resource_type,omitempty"`
	ResourceID    *string    `json:"resource_id,omitempty"`
	TableName     *string    `json:"table_name,omitempty"`
	RequestID     *string    `json:"request_id,omitempty"`
	IPAddress     *string    `json:"ip_address,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Limit         int        `json:"limit,omitempty"`
	Offset        int        `json:"offset,omitempty"`
}

// IsEmpty checks if all filters are empty
func (f *UnifiedAuditLogFilters) IsEmpty() bool {
	return f.UserID == nil &&
		f.APIKeyID == nil &&
		f.MerchantID == nil &&
		f.SessionID == nil &&
		f.EventType == nil &&
		f.EventCategory == nil &&
		f.Action == nil &&
		f.ResourceType == nil &&
		f.ResourceID == nil &&
		f.TableName == nil &&
		f.RequestID == nil &&
		f.IPAddress == nil &&
		f.StartDate == nil &&
		f.EndDate == nil
}

// UnifiedAuditLogResult represents the result of a unified audit log query
type UnifiedAuditLogResult struct {
	AuditLogs []*UnifiedAuditLog `json:"audit_logs"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
	HasMore   bool               `json:"has_more"`
}
