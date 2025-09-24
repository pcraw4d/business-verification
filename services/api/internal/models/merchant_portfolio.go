package models

import (
	"time"
)

// Address represents a business address
type Address struct {
	Street1     string `json:"street1" db:"street1"`
	Street2     string `json:"street2" db:"street2"`
	City        string `json:"city" db:"city"`
	State       string `json:"state" db:"state"`
	PostalCode  string `json:"postal_code" db:"postal_code"`
	Country     string `json:"country" db:"country"`
	CountryCode string `json:"country_code" db:"country_code"`
}

// ContactInfo represents business contact information
type ContactInfo struct {
	Phone          string `json:"phone" db:"phone"`
	Email          string `json:"email" db:"email"`
	Website        string `json:"website" db:"website"`
	PrimaryContact string `json:"primary_contact" db:"primary_contact"`
}

// Merchant represents a merchant in the portfolio system
type Merchant struct {
	ID                 string        `json:"id" db:"id"`
	Name               string        `json:"name" db:"name"`
	LegalName          string        `json:"legal_name" db:"legal_name"`
	RegistrationNumber string        `json:"registration_number" db:"registration_number"`
	TaxID              string        `json:"tax_id" db:"tax_id"`
	Industry           string        `json:"industry" db:"industry"`
	IndustryCode       string        `json:"industry_code" db:"industry_code"`
	BusinessType       string        `json:"business_type" db:"business_type"`
	FoundedDate        *time.Time    `json:"founded_date" db:"founded_date"`
	EmployeeCount      int           `json:"employee_count" db:"employee_count"`
	AnnualRevenue      *float64      `json:"annual_revenue" db:"annual_revenue"`
	Address            Address       `json:"address" db:"address"`
	ContactInfo        ContactInfo   `json:"contact_info" db:"contact_info"`
	PortfolioType      PortfolioType `json:"portfolio_type" db:"portfolio_type"`
	RiskLevel          RiskLevel     `json:"risk_level" db:"risk_level"`
	ComplianceStatus   string        `json:"compliance_status" db:"compliance_status"`
	Status             string        `json:"status" db:"status"`
	CreatedBy          string        `json:"created_by" db:"created_by"`
	CreatedAt          time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at" db:"updated_at"`
}

// PortfolioType represents the type of merchant in the portfolio
type PortfolioType string

const (
	PortfolioTypeOnboarded   PortfolioType = "onboarded"
	PortfolioTypeDeactivated PortfolioType = "deactivated"
	PortfolioTypeProspective PortfolioType = "prospective"
	PortfolioTypePending     PortfolioType = "pending"
)

// IsValid checks if the portfolio type is valid
func (pt PortfolioType) IsValid() bool {
	switch pt {
	case PortfolioTypeOnboarded, PortfolioTypeDeactivated, PortfolioTypeProspective, PortfolioTypePending:
		return true
	default:
		return false
	}
}

// String returns the string representation of the portfolio type
func (pt PortfolioType) String() string {
	return string(pt)
}

// RiskLevel represents the risk level of a merchant
type RiskLevel string

const (
	RiskLevelHigh   RiskLevel = "high"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelLow    RiskLevel = "low"
)

// IsValid checks if the risk level is valid
func (rl RiskLevel) IsValid() bool {
	switch rl {
	case RiskLevelHigh, RiskLevelMedium, RiskLevelLow:
		return true
	default:
		return false
	}
}

// String returns the string representation of the risk level
func (rl RiskLevel) String() string {
	return string(rl)
}

// GetNumericValue returns a numeric value for risk level comparison
func (rl RiskLevel) GetNumericValue() int {
	switch rl {
	case RiskLevelHigh:
		return 3
	case RiskLevelMedium:
		return 2
	case RiskLevelLow:
		return 1
	default:
		return 0
	}
}

// MerchantSession represents an active merchant session
type MerchantSession struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	MerchantID string    `json:"merchant_id" db:"merchant_id"`
	StartedAt  time.Time `json:"started_at" db:"started_at"`
	LastActive time.Time `json:"last_active" db:"last_active"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// IsExpired checks if the session has expired (24 hours)
func (ms *MerchantSession) IsExpired() bool {
	return time.Since(ms.LastActive) > 24*time.Hour
}

// UpdateLastActive updates the last active timestamp
func (ms *MerchantSession) UpdateLastActive() {
	ms.LastActive = time.Now()
	ms.UpdatedAt = time.Now()
}

// AuditLog represents an audit log entry for merchant operations
type AuditLog struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	MerchantID   string    `json:"merchant_id" db:"merchant_id"`
	Action       string    `json:"action" db:"action"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   string    `json:"resource_id" db:"resource_id"`
	Details      string    `json:"details" db:"details"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	RequestID    string    `json:"request_id" db:"request_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// MerchantSearchFilters represents filters for merchant search
type MerchantSearchFilters struct {
	PortfolioType    *PortfolioType `json:"portfolio_type,omitempty"`
	RiskLevel        *RiskLevel     `json:"risk_level,omitempty"`
	Industry         string         `json:"industry,omitempty"`
	Status           string         `json:"status,omitempty"`
	SearchQuery      string         `json:"search_query,omitempty"`
	CreatedAfter     *time.Time     `json:"created_after,omitempty"`
	CreatedBefore    *time.Time     `json:"created_before,omitempty"`
	EmployeeCountMin *int           `json:"employee_count_min,omitempty"`
	EmployeeCountMax *int           `json:"employee_count_max,omitempty"`
	RevenueMin       *float64       `json:"revenue_min,omitempty"`
	RevenueMax       *float64       `json:"revenue_max,omitempty"`
}

// IsEmpty checks if all filters are empty
func (f *MerchantSearchFilters) IsEmpty() bool {
	return f.PortfolioType == nil &&
		f.RiskLevel == nil &&
		f.Industry == "" &&
		f.Status == "" &&
		f.SearchQuery == "" &&
		f.CreatedAfter == nil &&
		f.CreatedBefore == nil &&
		f.EmployeeCountMin == nil &&
		f.EmployeeCountMax == nil &&
		f.RevenueMin == nil &&
		f.RevenueMax == nil
}

// MerchantListResult represents the result of a merchant list operation
type MerchantListResult struct {
	Merchants []*Merchant            `json:"merchants"`
	Total     int                    `json:"total"`
	Page      int                    `json:"page"`
	PageSize  int                    `json:"page_size"`
	HasMore   bool                   `json:"has_more"`
	Filters   *MerchantSearchFilters `json:"filters,omitempty"`
}

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	OperationID string                 `json:"operation_id"`
	Status      string                 `json:"status"`
	TotalItems  int                    `json:"total_items"`
	Processed   int                    `json:"processed"`
	Successful  int                    `json:"successful"`
	Failed      int                    `json:"failed"`
	Errors      []string               `json:"errors"`
	Results     []BulkOperationItem    `json:"results"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// BulkOperationItem represents a single item in a bulk operation
type BulkOperationItem struct {
	MerchantID string `json:"merchant_id"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

// BulkOperationStatus represents the status of a bulk operation
type BulkOperationStatus string

const (
	BulkOperationStatusPending    BulkOperationStatus = "pending"
	BulkOperationStatusProcessing BulkOperationStatus = "processing"
	BulkOperationStatusCompleted  BulkOperationStatus = "completed"
	BulkOperationStatusFailed     BulkOperationStatus = "failed"
	BulkOperationStatusCancelled  BulkOperationStatus = "cancelled"
)

// IsValid checks if the bulk operation status is valid
func (bos BulkOperationStatus) IsValid() bool {
	switch bos {
	case BulkOperationStatusPending, BulkOperationStatusProcessing, BulkOperationStatusCompleted, BulkOperationStatusFailed, BulkOperationStatusCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation of the bulk operation status
func (bos BulkOperationStatus) String() string {
	return string(bos)
}

// MerchantComparison represents a comparison between two merchants
type MerchantComparison struct {
	ID             string                 `json:"id" db:"id"`
	Merchant1ID    string                 `json:"merchant1_id" db:"merchant1_id"`
	Merchant2ID    string                 `json:"merchant2_id" db:"merchant2_id"`
	UserID         string                 `json:"user_id" db:"user_id"`
	ComparisonData map[string]interface{} `json:"comparison_data" db:"comparison_data"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// MerchantAnalytics represents analytics data for a merchant
type MerchantAnalytics struct {
	MerchantID        string                 `json:"merchant_id" db:"merchant_id"`
	RiskScore         float64                `json:"risk_score" db:"risk_score"`
	ComplianceScore   float64                `json:"compliance_score" db:"compliance_score"`
	TransactionVolume float64                `json:"transaction_volume" db:"transaction_volume"`
	LastActivity      *time.Time             `json:"last_activity" db:"last_activity"`
	Flags             []string               `json:"flags" db:"flags"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CalculatedAt      time.Time              `json:"calculated_at" db:"calculated_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// MerchantNotification represents a notification for a merchant
type MerchantNotification struct {
	ID         string     `json:"id" db:"id"`
	MerchantID string     `json:"merchant_id" db:"merchant_id"`
	UserID     string     `json:"user_id" db:"user_id"`
	Type       string     `json:"type" db:"type"`
	Title      string     `json:"title" db:"title"`
	Message    string     `json:"message" db:"message"`
	IsRead     bool       `json:"is_read" db:"is_read"`
	Priority   string     `json:"priority" db:"priority"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	ReadAt     *time.Time `json:"read_at" db:"read_at"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeRiskAlert     NotificationType = "risk_alert"
	NotificationTypeCompliance    NotificationType = "compliance"
	NotificationTypeStatusChange  NotificationType = "status_change"
	NotificationTypeBulkOperation NotificationType = "bulk_operation"
	NotificationTypeSystem        NotificationType = "system"
)

// IsValid checks if the notification type is valid
func (nt NotificationType) IsValid() bool {
	switch nt {
	case NotificationTypeRiskAlert, NotificationTypeCompliance, NotificationTypeStatusChange, NotificationTypeBulkOperation, NotificationTypeSystem:
		return true
	default:
		return false
	}
}

// String returns the string representation of the notification type
func (nt NotificationType) String() string {
	return string(nt)
}

// NotificationPriority represents the priority of a notification
type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityMedium   NotificationPriority = "medium"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityCritical NotificationPriority = "critical"
)

// IsValid checks if the notification priority is valid
func (np NotificationPriority) IsValid() bool {
	switch np {
	case NotificationPriorityLow, NotificationPriorityMedium, NotificationPriorityHigh, NotificationPriorityCritical:
		return true
	default:
		return false
	}
}

// String returns the string representation of the notification priority
func (np NotificationPriority) String() string {
	return string(np)
}

// GetNumericValue returns a numeric value for priority comparison
func (np NotificationPriority) GetNumericValue() int {
	switch np {
	case NotificationPriorityCritical:
		return 4
	case NotificationPriorityHigh:
		return 3
	case NotificationPriorityMedium:
		return 2
	case NotificationPriorityLow:
		return 1
	default:
		return 0
	}
}

// MerchantPortfolioSummary represents a summary of the merchant portfolio
type MerchantPortfolioSummary struct {
	TotalMerchants    int                 `json:"total_merchants"`
	OnboardedCount    int                 `json:"onboarded_count"`
	ProspectiveCount  int                 `json:"prospective_count"`
	PendingCount      int                 `json:"pending_count"`
	DeactivatedCount  int                 `json:"deactivated_count"`
	HighRiskCount     int                 `json:"high_risk_count"`
	MediumRiskCount   int                 `json:"medium_risk_count"`
	LowRiskCount      int                 `json:"low_risk_count"`
	IndustryBreakdown map[string]int      `json:"industry_breakdown"`
	ComplianceStatus  map[string]int      `json:"compliance_status"`
	RecentActivity    []*MerchantActivity `json:"recent_activity"`
	RiskTrends        []*RiskTrend        `json:"risk_trends"`
	GeneratedAt       time.Time           `json:"generated_at"`
}

// MerchantActivity represents recent activity for a merchant
type MerchantActivity struct {
	MerchantID string    `json:"merchant_id"`
	Action     string    `json:"action"`
	Details    string    `json:"details"`
	UserID     string    `json:"user_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// RiskTrend represents risk trend data
type RiskTrend struct {
	Date       time.Time `json:"date"`
	HighRisk   int       `json:"high_risk"`
	MediumRisk int       `json:"medium_risk"`
	LowRisk    int       `json:"low_risk"`
}

// Validation errors
var (
	ErrInvalidMerchantName         = "merchant name is required and cannot be empty"
	ErrInvalidPortfolioType        = "invalid portfolio type"
	ErrInvalidRiskLevel            = "invalid risk level"
	ErrInvalidNotificationType     = "invalid notification type"
	ErrInvalidNotificationPriority = "invalid notification priority"
	ErrInvalidBulkOperationStatus  = "invalid bulk operation status"
)

// Validate validates the merchant data
func (m *Merchant) Validate() error {
	if m.Name == "" {
		return &ValidationError{Field: "name", Message: ErrInvalidMerchantName}
	}

	if !m.PortfolioType.IsValid() {
		return &ValidationError{Field: "portfolio_type", Message: ErrInvalidPortfolioType}
	}

	if !m.RiskLevel.IsValid() {
		return &ValidationError{Field: "risk_level", Message: ErrInvalidRiskLevel}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	return ve.Message
}

// Validate validates the merchant session data
func (ms *MerchantSession) Validate() error {
	if ms.UserID == "" {
		return &ValidationError{Field: "user_id", Message: "user ID is required"}
	}

	if ms.MerchantID == "" {
		return &ValidationError{Field: "merchant_id", Message: "merchant ID is required"}
	}

	return nil
}

// Validate validates the audit log data
func (al *AuditLog) Validate() error {
	if al.UserID == "" {
		return &ValidationError{Field: "user_id", Message: "user ID is required"}
	}

	if al.Action == "" {
		return &ValidationError{Field: "action", Message: "action is required"}
	}

	if al.ResourceType == "" {
		return &ValidationError{Field: "resource_type", Message: "resource type is required"}
	}

	if al.ResourceID == "" {
		return &ValidationError{Field: "resource_id", Message: "resource ID is required"}
	}

	return nil
}

// Validate validates the merchant notification data
func (mn *MerchantNotification) Validate() error {
	if mn.MerchantID == "" {
		return &ValidationError{Field: "merchant_id", Message: "merchant ID is required"}
	}

	if mn.UserID == "" {
		return &ValidationError{Field: "user_id", Message: "user ID is required"}
	}

	if mn.Type == "" {
		return &ValidationError{Field: "type", Message: "notification type is required"}
	}

	if mn.Title == "" {
		return &ValidationError{Field: "title", Message: "notification title is required"}
	}

	if mn.Message == "" {
		return &ValidationError{Field: "message", Message: "notification message is required"}
	}

	return nil
}

// Validate validates the merchant comparison data
func (mc *MerchantComparison) Validate() error {
	if mc.Merchant1ID == "" {
		return &ValidationError{Field: "merchant1_id", Message: "first merchant ID is required"}
	}

	if mc.Merchant2ID == "" {
		return &ValidationError{Field: "merchant2_id", Message: "second merchant ID is required"}
	}

	if mc.UserID == "" {
		return &ValidationError{Field: "user_id", Message: "user ID is required"}
	}

	if mc.Merchant1ID == mc.Merchant2ID {
		return &ValidationError{Field: "merchants", Message: "cannot compare a merchant with itself"}
	}

	return nil
}

// Validate validates the merchant analytics data
func (ma *MerchantAnalytics) Validate() error {
	if ma.MerchantID == "" {
		return &ValidationError{Field: "merchant_id", Message: "merchant ID is required"}
	}

	if ma.RiskScore < 0 || ma.RiskScore > 1 {
		return &ValidationError{Field: "risk_score", Message: "risk score must be between 0 and 1"}
	}

	if ma.ComplianceScore < 0 || ma.ComplianceScore > 1 {
		return &ValidationError{Field: "compliance_score", Message: "compliance score must be between 0 and 1"}
	}

	return nil
}
