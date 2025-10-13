package audit

import (
	"time"
)

// AuditEvent represents a single audit event in the system
type AuditEvent struct {
	ID           string                 `json:"id" db:"id"`
	TenantID     string                 `json:"tenant_id" db:"tenant_id"`
	UserID       string                 `json:"user_id" db:"user_id"`
	SessionID    string                 `json:"session_id" db:"session_id"`
	Action       string                 `json:"action" db:"action"`
	Resource     string                 `json:"resource" db:"resource"`
	ResourceID   string                 `json:"resource_id" db:"resource_id"`
	Method       string                 `json:"method" db:"method"`
	Endpoint     string                 `json:"endpoint" db:"endpoint"`
	IPAddress    string                 `json:"ip_address" db:"ip_address"`
	UserAgent    string                 `json:"user_agent" db:"user_agent"`
	RequestID    string                 `json:"request_id" db:"request_id"`
	Status       int                    `json:"status" db:"status"`
	Duration     int64                  `json:"duration" db:"duration"` // milliseconds
	RequestSize  int64                  `json:"request_size" db:"request_size"`
	ResponseSize int64                  `json:"response_size" db:"response_size"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	Hash         string                 `json:"hash" db:"hash"` // Cryptographic hash for integrity
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// AuditLog represents the immutable audit log entry
type AuditLog struct {
	ID        string    `json:"id" db:"id"`
	EventID   string    `json:"event_id" db:"event_id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	EventData []byte    `json:"event_data" db:"event_data"` // JSON serialized AuditEvent
	Hash      string    `json:"hash" db:"hash"`             // SHA-256 hash of event data
	PrevHash  string    `json:"prev_hash" db:"prev_hash"`   // Hash of previous log entry (blockchain-like)
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AuditQuery represents parameters for querying audit logs
type AuditQuery struct {
	TenantID   string                 `json:"tenant_id"`
	UserID     string                 `json:"user_id"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id"`
	StartDate  *time.Time             `json:"start_date"`
	EndDate    *time.Time             `json:"end_date"`
	IPAddress  string                 `json:"ip_address"`
	Status     *int                   `json:"status"`
	Metadata   map[string]interface{} `json:"metadata"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
	SortBy     string                 `json:"sort_by"`
	SortOrder  string                 `json:"sort_order"`
}

// AuditStats represents audit statistics
type AuditStats struct {
	TotalEvents     int64            `json:"total_events"`
	EventsByAction  map[string]int64 `json:"events_by_action"`
	EventsByUser    map[string]int64 `json:"events_by_user"`
	EventsByStatus  map[int]int64    `json:"events_by_status"`
	EventsByDay     map[string]int64 `json:"events_by_day"`
	AverageDuration float64          `json:"average_duration"`
	ErrorRate       float64          `json:"error_rate"`
	TopEndpoints    []EndpointStats  `json:"top_endpoints"`
	TopUsers        []UserStats      `json:"top_users"`
}

// EndpointStats represents statistics for an endpoint
type EndpointStats struct {
	Endpoint     string  `json:"endpoint"`
	Method       string  `json:"method"`
	RequestCount int64   `json:"request_count"`
	AverageTime  float64 `json:"average_time"`
	ErrorCount   int64   `json:"error_count"`
	ErrorRate    float64 `json:"error_rate"`
}

// UserStats represents statistics for a user
type UserStats struct {
	UserID       string    `json:"user_id"`
	RequestCount int64     `json:"request_count"`
	LastActivity time.Time `json:"last_activity"`
	ErrorCount   int64     `json:"error_count"`
	ErrorRate    float64   `json:"error_rate"`
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	ReportType  string                 `json:"report_type" db:"report_type"`
	ReportName  string                 `json:"report_name" db:"report_name"`
	Period      string                 `json:"period" db:"period"`
	StartDate   time.Time              `json:"start_date" db:"start_date"`
	EndDate     time.Time              `json:"end_date" db:"end_date"`
	Status      string                 `json:"status" db:"status"`
	Data        map[string]interface{} `json:"data" db:"data"`
	GeneratedBy string                 `json:"generated_by" db:"generated_by"`
	GeneratedAt time.Time              `json:"generated_at" db:"generated_at"`
	ExpiresAt   *time.Time             `json:"expires_at" db:"expires_at"`
	Hash        string                 `json:"hash" db:"hash"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// ReportTemplate represents a compliance report template
type ReportTemplate struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Type        string                 `json:"type" db:"type"`
	Description string                 `json:"description" db:"description"`
	Template    map[string]interface{} `json:"template" db:"template"`
	Parameters  []ReportParameter      `json:"parameters" db:"parameters"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// ReportParameter represents a parameter for report generation
type ReportParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Options     []string    `json:"options,omitempty"`
}

// AuditRetentionPolicy represents data retention policies
type AuditRetentionPolicy struct {
	ID            string    `json:"id" db:"id"`
	TenantID      string    `json:"tenant_id" db:"tenant_id"`
	PolicyName    string    `json:"policy_name" db:"policy_name"`
	Description   string    `json:"description" db:"description"`
	RetentionDays int       `json:"retention_days" db:"retention_days"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// AuditExport represents an audit log export
type AuditExport struct {
	ID          string     `json:"id" db:"id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	ExportType  string     `json:"export_type" db:"export_type"` // json, csv, pdf
	Query       AuditQuery `json:"query" db:"query"`
	Status      string     `json:"status" db:"status"` // pending, processing, completed, failed
	FilePath    string     `json:"file_path" db:"file_path"`
	FileSize    int64      `json:"file_size" db:"file_size"`
	RecordCount int64      `json:"record_count" db:"record_count"`
	RequestedBy string     `json:"requested_by" db:"requested_by"`
	RequestedAt time.Time  `json:"requested_at" db:"requested_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	ExpiresAt   time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// AuditAlert represents an audit alert
type AuditAlert struct {
	ID          string                 `json:"id" db:"id"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	AlertType   string                 `json:"alert_type" db:"alert_type"`
	Severity    string                 `json:"severity" db:"severity"`
	Title       string                 `json:"title" db:"title"`
	Description string                 `json:"description" db:"description"`
	EventID     string                 `json:"event_id" db:"event_id"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	IsResolved  bool                   `json:"is_resolved" db:"is_resolved"`
	ResolvedBy  string                 `json:"resolved_by" db:"resolved_by"`
	ResolvedAt  *time.Time             `json:"resolved_at" db:"resolved_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// AuditConfig represents audit system configuration
type AuditConfig struct {
	Enabled           bool            `json:"enabled"`
	LogLevel          string          `json:"log_level"`
	RetentionDays     int             `json:"retention_days"`
	BatchSize         int             `json:"batch_size"`
	FlushInterval     time.Duration   `json:"flush_interval"`
	EnableHashing     bool            `json:"enable_hashing"`
	EnableCompression bool            `json:"enable_compression"`
	MaxFileSize       int64           `json:"max_file_size"`
	AlertThresholds   AlertThresholds `json:"alert_thresholds"`
}

// AlertThresholds represents alert thresholds
type AlertThresholds struct {
	FailedLogins int     `json:"failed_logins"`
	DataAccess   int     `json:"data_access"`
	AdminActions int     `json:"admin_actions"`
	ErrorRate    float64 `json:"error_rate"`
	ResponseTime int64   `json:"response_time"`
}
