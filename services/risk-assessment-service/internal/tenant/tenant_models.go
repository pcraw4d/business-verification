package tenant

import (
	"time"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID                 string                 `json:"id" db:"id"`
	Name               string                 `json:"name" db:"name"`
	Domain             string                 `json:"domain" db:"domain"`
	Status             TenantStatus           `json:"status" db:"status"`
	Plan               TenantPlan             `json:"plan" db:"plan"`
	Configuration      map[string]interface{} `json:"configuration" db:"configuration"`
	Quotas             TenantQuotas           `json:"quotas" db:"quotas"`
	Features           []string               `json:"features" db:"features"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
	LastActivityAt     *time.Time             `json:"last_activity_at" db:"last_activity_at"`
	SubscriptionEndsAt *time.Time             `json:"subscription_ends_at" db:"subscription_ends_at"`
	Metadata           map[string]interface{} `json:"metadata" db:"metadata"`
}

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusPending   TenantStatus = "pending"
	TenantStatusCancelled TenantStatus = "cancelled"
)

// TenantPlan represents the subscription plan of a tenant
type TenantPlan string

const (
	TenantPlanFree         TenantPlan = "free"
	TenantPlanBasic        TenantPlan = "basic"
	TenantPlanProfessional TenantPlan = "professional"
	TenantPlanEnterprise   TenantPlan = "enterprise"
)

// TenantQuotas represents the quotas for a tenant
type TenantQuotas struct {
	MaxAssessmentsPerMonth int64 `json:"max_assessments_per_month" db:"max_assessments_per_month"`
	MaxUsers               int   `json:"max_users" db:"max_users"`
	MaxAPIRequestsPerDay   int64 `json:"max_api_requests_per_day" db:"max_api_requests_per_day"`
	MaxDataRetentionDays   int   `json:"max_data_retention_days" db:"max_data_retention_days"`
	MaxConcurrentRequests  int   `json:"max_concurrent_requests" db:"max_concurrent_requests"`
	MaxFileUploadSize      int64 `json:"max_file_upload_size" db:"max_file_upload_size"`
	MaxAuditLogRetention   int   `json:"max_audit_log_retention" db:"max_audit_log_retention"`
}

// TenantUser represents a user within a tenant
type TenantUser struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	UserID      string                 `json:"user_id" db:"user_id"`
	Email       string                 `json:"email" db:"email"`
	Role        TenantUserRole         `json:"role" db:"role"`
	Permissions []string               `json:"permissions" db:"permissions"`
	Status      TenantUserStatus       `json:"status" db:"status"`
	LastLoginAt *time.Time             `json:"last_login_at" db:"last_login_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// TenantUserRole represents the role of a user within a tenant
type TenantUserRole string

const (
	TenantUserRoleOwner   TenantUserRole = "owner"
	TenantUserRoleAdmin   TenantUserRole = "admin"
	TenantUserRoleManager TenantUserRole = "manager"
	TenantUserRoleAnalyst TenantUserRole = "analyst"
	TenantUserRoleViewer  TenantUserRole = "viewer"
	TenantUserRoleAPI     TenantUserRole = "api"
)

// TenantUserStatus represents the status of a user within a tenant
type TenantUserStatus string

const (
	TenantUserStatusActive    TenantUserStatus = "active"
	TenantUserStatusInactive  TenantUserStatus = "inactive"
	TenantUserStatusPending   TenantUserStatus = "pending"
	TenantUserStatusSuspended TenantUserStatus = "suspended"
)

// TenantAPIKey represents an API key for a tenant
type TenantAPIKey struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	Name        string                 `json:"name" db:"name"`
	KeyHash     string                 `json:"key_hash" db:"key_hash"`
	Permissions []string               `json:"permissions" db:"permissions"`
	RateLimit   int                    `json:"rate_limit" db:"rate_limit"`
	Status      APIKeyStatus           `json:"status" db:"status"`
	LastUsedAt  *time.Time             `json:"last_used_at" db:"last_used_at"`
	ExpiresAt   *time.Time             `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// APIKeyStatus represents the status of an API key
type APIKeyStatus string

const (
	APIKeyStatusActive   APIKeyStatus = "active"
	APIKeyStatusInactive APIKeyStatus = "inactive"
	APIKeyStatusExpired  APIKeyStatus = "expired"
	APIKeyStatusRevoked  APIKeyStatus = "revoked"
)

// TenantConfiguration represents tenant-specific configuration
type TenantConfiguration struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	Category    string                 `json:"category" db:"category"`
	Key         string                 `json:"key" db:"key"`
	Value       interface{}            `json:"value" db:"value"`
	ValueType   string                 `json:"value_type" db:"value_type"`
	Description string                 `json:"description" db:"description"`
	IsEncrypted bool                   `json:"is_encrypted" db:"is_encrypted"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	UpdatedBy   string                 `json:"updated_by" db:"updated_by"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// TenantUsage represents usage statistics for a tenant
type TenantUsage struct {
	ID                     string    `json:"id" db:"id"`
	TenantID               string    `json:"tenant_id" db:"tenant_id"`
	Period                 string    `json:"period" db:"period"`
	AssessmentsCount       int64     `json:"assessments_count" db:"assessments_count"`
	APIRequestsCount       int64     `json:"api_requests_count" db:"api_requests_count"`
	UsersCount             int       `json:"users_count" db:"users_count"`
	DataStorageBytes       int64     `json:"data_storage_bytes" db:"data_storage_bytes"`
	AuditLogsCount         int64     `json:"audit_logs_count" db:"audit_logs_count"`
	ComplianceReportsCount int64     `json:"compliance_reports_count" db:"compliance_reports_count"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// TenantContext represents the tenant context for a request
type TenantContext struct {
	TenantID    string                 `json:"tenant_id"`
	UserID      string                 `json:"user_id"`
	UserRole    TenantUserRole         `json:"user_role"`
	Permissions []string               `json:"permissions"`
	APIKeyID    string                 `json:"api_key_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TenantIsolationConfig represents tenant isolation configuration
type TenantIsolationConfig struct {
	EnableRowLevelSecurity bool   `json:"enable_row_level_security"`
	EnableDataEncryption   bool   `json:"enable_data_encryption"`
	EnableAuditLogging     bool   `json:"enable_audit_logging"`
	EnableRateLimiting     bool   `json:"enable_rate_limiting"`
	EnableQuotaEnforcement bool   `json:"enable_quota_enforcement"`
	DataResidencyRegion    string `json:"data_residency_region"`
	ComplianceFramework    string `json:"compliance_framework"`
}

// TenantMetrics represents metrics for a tenant
type TenantMetrics struct {
	TenantID            string                 `json:"tenant_id"`
	ActiveUsers         int                    `json:"active_users"`
	TotalAssessments    int64                  `json:"total_assessments"`
	APIRequestsToday    int64                  `json:"api_requests_today"`
	StorageUsed         int64                  `json:"storage_used"`
	QuotaUtilization    map[string]float64     `json:"quota_utilization"`
	LastActivityAt      *time.Time             `json:"last_activity_at"`
	HealthScore         float64                `json:"health_score"`
	ComplianceScore     float64                `json:"compliance_score"`
	PerformanceMetrics  map[string]interface{} `json:"performance_metrics"`
	ErrorRate           float64                `json:"error_rate"`
	AverageResponseTime float64                `json:"average_response_time"`
}

// TenantEvent represents an event related to a tenant
type TenantEvent struct {
	ID        string                 `json:"id" db:"id"`
	TenantID  string                 `json:"tenant_id" db:"tenant_id"`
	EventType TenantEventType        `json:"event_type" db:"event_type"`
	EventData map[string]interface{} `json:"event_data" db:"event_data"`
	UserID    string                 `json:"user_id" db:"user_id"`
	IPAddress string                 `json:"ip_address" db:"ip_address"`
	UserAgent string                 `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// TenantEventType represents the type of tenant event
type TenantEventType string

const (
	TenantEventTypeCreated       TenantEventType = "tenant_created"
	TenantEventTypeUpdated       TenantEventType = "tenant_updated"
	TenantEventTypeSuspended     TenantEventType = "tenant_suspended"
	TenantEventTypeActivated     TenantEventType = "tenant_activated"
	TenantEventTypeUserAdded     TenantEventType = "user_added"
	TenantEventTypeUserRemoved   TenantEventType = "user_removed"
	TenantEventTypeAPIKeyCreated TenantEventType = "api_key_created"
	TenantEventTypeAPIKeyRevoked TenantEventType = "api_key_revoked"
	TenantEventTypeQuotaExceeded TenantEventType = "quota_exceeded"
	TenantEventTypePlanChanged   TenantEventType = "plan_changed"
	TenantEventTypeConfigUpdated TenantEventType = "config_updated"
)

// Default tenant quotas by plan
var DefaultQuotasByPlan = map[TenantPlan]TenantQuotas{
	TenantPlanFree: {
		MaxAssessmentsPerMonth: 100,
		MaxUsers:               3,
		MaxAPIRequestsPerDay:   1000,
		MaxDataRetentionDays:   30,
		MaxConcurrentRequests:  5,
		MaxFileUploadSize:      10 * 1024 * 1024, // 10MB
		MaxAuditLogRetention:   30,
	},
	TenantPlanBasic: {
		MaxAssessmentsPerMonth: 1000,
		MaxUsers:               10,
		MaxAPIRequestsPerDay:   10000,
		MaxDataRetentionDays:   90,
		MaxConcurrentRequests:  20,
		MaxFileUploadSize:      50 * 1024 * 1024, // 50MB
		MaxAuditLogRetention:   90,
	},
	TenantPlanProfessional: {
		MaxAssessmentsPerMonth: 10000,
		MaxUsers:               50,
		MaxAPIRequestsPerDay:   100000,
		MaxDataRetentionDays:   365,
		MaxConcurrentRequests:  100,
		MaxFileUploadSize:      200 * 1024 * 1024, // 200MB
		MaxAuditLogRetention:   365,
	},
	TenantPlanEnterprise: {
		MaxAssessmentsPerMonth: -1, // Unlimited
		MaxUsers:               -1, // Unlimited
		MaxAPIRequestsPerDay:   -1, // Unlimited
		MaxDataRetentionDays:   -1, // Unlimited
		MaxConcurrentRequests:  -1, // Unlimited
		MaxFileUploadSize:      -1, // Unlimited
		MaxAuditLogRetention:   -1, // Unlimited
	},
}

// Default permissions by role
var DefaultPermissionsByRole = map[TenantUserRole][]string{
	TenantUserRoleOwner: {
		"tenant:read", "tenant:write", "tenant:delete",
		"users:read", "users:write", "users:delete",
		"assessments:read", "assessments:write", "assessments:delete",
		"reports:read", "reports:write", "reports:delete",
		"api_keys:read", "api_keys:write", "api_keys:delete",
		"config:read", "config:write",
		"audit:read", "compliance:read", "compliance:write",
	},
	TenantUserRoleAdmin: {
		"tenant:read", "tenant:write",
		"users:read", "users:write", "users:delete",
		"assessments:read", "assessments:write", "assessments:delete",
		"reports:read", "reports:write", "reports:delete",
		"api_keys:read", "api_keys:write", "api_keys:delete",
		"config:read", "config:write",
		"audit:read", "compliance:read", "compliance:write",
	},
	TenantUserRoleManager: {
		"tenant:read",
		"users:read", "users:write",
		"assessments:read", "assessments:write",
		"reports:read", "reports:write",
		"api_keys:read", "api_keys:write",
		"config:read",
		"audit:read", "compliance:read",
	},
	TenantUserRoleAnalyst: {
		"assessments:read", "assessments:write",
		"reports:read", "reports:write",
		"audit:read", "compliance:read",
	},
	TenantUserRoleViewer: {
		"assessments:read",
		"reports:read",
		"audit:read",
	},
	TenantUserRoleAPI: {
		"assessments:read", "assessments:write",
		"reports:read",
	},
}
