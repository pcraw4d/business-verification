package models

import "time"

// ComplianceTracking represents a compliance tracking record
type ComplianceTracking struct {
	ID                  string                 `json:"id" db:"id"`
	MerchantID          string                 `json:"merchant_id" db:"merchant_id"`
	ComplianceType      string                 `json:"compliance_type" db:"compliance_type"`
	ComplianceFramework string                 `json:"compliance_framework" db:"compliance_framework"`
	CheckType           string                 `json:"check_type" db:"check_type"`
	Status              string                 `json:"status" db:"status"`
	Score               *float64               `json:"score" db:"score"`
	Requirements        map[string]interface{} `json:"requirements" db:"requirements"`
	CheckMethod         string                 `json:"check_method" db:"check_method"`
	Source              string                 `json:"source" db:"source"`
	RawData             map[string]interface{} `json:"raw_data" db:"raw_data"`
	Result              map[string]interface{} `json:"result" db:"result"`
	Findings            map[string]interface{} `json:"findings" db:"findings"`
	Recommendations     map[string]interface{} `json:"recommendations" db:"recommendations"`
	Evidence            map[string]interface{} `json:"evidence" db:"evidence"`
	CheckedBy           string                 `json:"checked_by" db:"checked_by"`
	CheckedAt           *time.Time             `json:"checked_at" db:"checked_at"`
	ReviewedBy          string                 `json:"reviewed_by" db:"reviewed_by"`
	ReviewedAt          *time.Time             `json:"reviewed_at" db:"reviewed_at"`
	ApprovedBy          string                 `json:"approved_by" db:"approved_by"`
	ApprovedAt          *time.Time             `json:"approved_at" db:"approved_at"`
	DueDate             *time.Time             `json:"due_date" db:"due_date"`
	ExpiresAt           *time.Time             `json:"expires_at" db:"expires_at"`
	NextReviewDate      *time.Time             `json:"next_review_date" db:"next_review_date"`
	Priority            string                 `json:"priority" db:"priority"`
	AssignedTo          string                 `json:"assigned_to" db:"assigned_to"`
	Tags                []string               `json:"tags" db:"tags"`
	LastChecked         *time.Time             `json:"last_checked" db:"last_checked"`
	NextCheck           *time.Time             `json:"next_check" db:"next_check"`
	ComplianceScore     *float64               `json:"compliance_score" db:"compliance_score"`
	RiskLevel           string                 `json:"risk_level" db:"risk_level"`
	Notes               *string                `json:"notes" db:"notes"`
	Metadata            map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy           string                 `json:"created_by" db:"created_by"`
	UpdatedBy           string                 `json:"updated_by" db:"updated_by"`
}

// ComplianceTrackingFilters represents filters for compliance tracking queries
type ComplianceTrackingFilters struct {
	MerchantID          *string    `json:"merchant_id,omitempty"`
	ComplianceType      *string    `json:"compliance_type,omitempty"`
	ComplianceFramework *string    `json:"compliance_framework,omitempty"`
	CheckType           *string    `json:"check_type,omitempty"`
	Status              *string    `json:"status,omitempty"`
	Priority            *string    `json:"priority,omitempty"`
	RiskLevel           *string    `json:"risk_level,omitempty"`
	LastCheckedFrom     *time.Time `json:"last_checked_from,omitempty"`
	LastCheckedTo       *time.Time `json:"last_checked_to,omitempty"`
	Limit               *int       `json:"limit,omitempty"`
	Offset              *int       `json:"offset,omitempty"`
	SortBy              *string    `json:"sort_by,omitempty"`
	SortOrder           *string    `json:"sort_order,omitempty"`
}

// MerchantComplianceSummary represents a summary of merchant compliance
type MerchantComplianceSummary struct {
	MerchantID             string     `json:"merchant_id"`
	TotalComplianceTypes   int        `json:"total_compliance_types"`
	CompliantTypes         int        `json:"compliant_types"`
	NonCompliantTypes      int        `json:"non_compliant_types"`
	PendingTypes           int        `json:"pending_types"`
	TotalChecks            int        `json:"total_checks"`
	CompletedChecks        int        `json:"completed_checks"`
	PendingChecks          int        `json:"pending_checks"`
	FailedChecks           int        `json:"failed_checks"`
	OverdueChecks          int        `json:"overdue_checks"`
	PastDueChecks          int        `json:"past_due_checks"`
	ComplianceTypesCovered []string   `json:"compliance_types_covered"`
	GeneratedAt            time.Time  `json:"generated_at"`
	ComplianceScore        float64    `json:"compliance_score"`
	AverageScore           float64    `json:"average_score"`
	LastCheckDate          *time.Time `json:"last_check_date"`
	NextReviewDate         *time.Time `json:"next_review_date"`
	OverallScore           float64    `json:"overall_score"`
	RiskLevel              string     `json:"risk_level"`
	LastUpdated            time.Time  `json:"last_updated"`
}

// UnifiedComplianceAlert represents a compliance alert
type UnifiedComplianceAlert struct {
	ID                  string     `json:"id" db:"id"`
	MerchantID          string     `json:"merchant_id" db:"merchant_id"`
	AlertType           string     `json:"alert_type" db:"alert_type"`
	Severity            string     `json:"severity" db:"severity"`
	Title               string     `json:"title" db:"title"`
	Description         string     `json:"description" db:"description"`
	ComplianceType      string     `json:"compliance_type" db:"compliance_type"`
	ComplianceFramework string     `json:"compliance_framework" db:"compliance_framework"`
	Priority            string     `json:"priority" db:"priority"`
	RiskLevel           string     `json:"risk_level" db:"risk_level"`
	Status              string     `json:"status" db:"status"`
	DueDate             *time.Time `json:"due_date" db:"due_date"`
	ExpiresAt           *time.Time `json:"expires_at" db:"expires_at"`
	AssignedTo          string     `json:"assigned_to" db:"assigned_to"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// ComplianceAlertFilters represents filters for compliance alert queries
type ComplianceAlertFilters struct {
	MerchantID     *string `json:"merchant_id,omitempty"`
	AlertType      *string `json:"alert_type,omitempty"`
	Severity       *string `json:"severity,omitempty"`
	ComplianceType *string `json:"compliance_type,omitempty"`
	Status         *string `json:"status,omitempty"`
	Limit          *int    `json:"limit,omitempty"`
	Offset         *int    `json:"offset,omitempty"`
	SortBy         *string `json:"sort_by,omitempty"`
	SortOrder      *string `json:"sort_order,omitempty"`
}

// UnifiedComplianceTrend represents a compliance trend
type UnifiedComplianceTrend struct {
	ID              string    `json:"id" db:"id"`
	MerchantID      string    `json:"merchant_id" db:"merchant_id"`
	ComplianceType  string    `json:"compliance_type" db:"compliance_type"`
	Period          string    `json:"period" db:"period"`
	Score           float64   `json:"score" db:"score"`
	Trend           string    `json:"trend" db:"trend"`
	Date            time.Time `json:"date" db:"date"`
	TotalChecks     int       `json:"total_checks" db:"total_checks"`
	CompletedChecks int       `json:"completed_checks" db:"completed_checks"`
	FailedChecks    int       `json:"failed_checks" db:"failed_checks"`
	ComplianceScore float64   `json:"compliance_score" db:"compliance_score"`
	AverageScore    float64   `json:"average_score" db:"average_score"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// ComplianceTrendFilters represents filters for compliance trend queries
type ComplianceTrendFilters struct {
	MerchantID     *string `json:"merchant_id,omitempty"`
	ComplianceType *string `json:"compliance_type,omitempty"`
	Period         *string `json:"period,omitempty"`
	Limit          *int    `json:"limit,omitempty"`
	Offset         *int    `json:"offset,omitempty"`
	SortBy         *string `json:"sort_by,omitempty"`
	SortOrder      *string `json:"sort_order,omitempty"`
}

// CreateComplianceTrackingRequest represents a request to create compliance tracking
type CreateComplianceTrackingRequest struct {
	MerchantID      string                 `json:"merchant_id" validate:"required"`
	ComplianceType  string                 `json:"compliance_type" validate:"required"`
	Status          string                 `json:"status" validate:"required"`
	LastChecked     *time.Time             `json:"last_checked,omitempty"`
	NextCheck       *time.Time             `json:"next_check,omitempty"`
	ComplianceScore *float64               `json:"compliance_score,omitempty"`
	RiskLevel       string                 `json:"risk_level,omitempty"`
	Notes           *string                `json:"notes,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateComplianceTrackingRequest represents a request to update compliance tracking
type UpdateComplianceTrackingRequest struct {
	ID              string                 `json:"id" validate:"required"`
	Status          *string                `json:"status,omitempty"`
	LastChecked     *time.Time             `json:"last_checked,omitempty"`
	NextCheck       *time.Time             `json:"next_check,omitempty"`
	ComplianceScore *float64               `json:"compliance_score,omitempty"`
	RiskLevel       *string                `json:"risk_level,omitempty"`
	Notes           *string                `json:"notes,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}
