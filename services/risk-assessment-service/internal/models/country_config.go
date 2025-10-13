package models

import (
	"time"
)

// CountryConfig represents the configuration for a specific country
type CountryConfig struct {
	ID                 string                 `json:"id" db:"id"`
	Code               string                 `json:"code" db:"code"`
	Name               string                 `json:"name" db:"name"`
	Region             string                 `json:"region" db:"region"`
	Currency           string                 `json:"currency" db:"currency"`
	Language           string                 `json:"language" db:"language"`
	Timezone           string                 `json:"timezone" db:"timezone"`
	IsActive           bool                   `json:"is_active" db:"is_active"`
	RiskFactors        []AssessmentRiskFactor `json:"risk_factors" db:"risk_factors"`
	ComplianceRules    []ComplianceRule       `json:"compliance_rules" db:"compliance_rules"`
	ValidationRules    []ValidationRule       `json:"validation_rules" db:"validation_rules"`
	SanctionsLists     []string               `json:"sanctions_lists" db:"sanctions_lists"`
	RegulatoryBodies   []RegulatoryBody       `json:"regulatory_bodies" db:"regulatory_bodies"`
	DataResidencyRules DataResidencyRules     `json:"data_residency_rules" db:"data_residency_rules"`
	BusinessTypes      []BusinessType         `json:"business_types" db:"business_types"`
	DocumentTypes      []DocumentType         `json:"document_types" db:"document_types"`
	Metadata           map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
}

// CountryRiskFactor represents a country-specific risk factor
type CountryRiskFactor struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	Severity      string                 `json:"severity"`
	Weight        float64                `json:"weight"`
	IsActive      bool                   `json:"is_active"`
	LocalizedName map[string]string      `json:"localized_name"`
	LocalizedDesc map[string]string      `json:"localized_desc"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ComplianceRule represents a country-specific compliance rule
type ComplianceRule struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Type           string                 `json:"type"`
	Category       string                 `json:"category"`
	IsMandatory    bool                   `json:"is_mandatory"`
	EffectiveDate  time.Time              `json:"effective_date"`
	ExpiryDate     *time.Time             `json:"expiry_date,omitempty"`
	RegulatoryBody string                 `json:"regulatory_body"`
	Penalty        string                 `json:"penalty,omitempty"`
	Requirements   []string               `json:"requirements"`
	LocalizedName  map[string]string      `json:"localized_name"`
	LocalizedDesc  map[string]string      `json:"localized_desc"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ValidationRule represents a country-specific validation rule
type ValidationRule struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Field          string                 `json:"field"`
	Type           string                 `json:"type"`
	Pattern        string                 `json:"pattern,omitempty"`
	MinLength      int                    `json:"min_length,omitempty"`
	MaxLength      int                    `json:"max_length,omitempty"`
	Required       bool                   `json:"required"`
	ErrorMessage   string                 `json:"error_message"`
	LocalizedError map[string]string      `json:"localized_error"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// RegulatoryBody represents a regulatory body for a country
type RegulatoryBody struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Acronym          string                 `json:"acronym"`
	Type             string                 `json:"type"`
	Jurisdiction     string                 `json:"jurisdiction"`
	Website          string                 `json:"website,omitempty"`
	ContactInfo      ContactInfo            `json:"contact_info,omitempty"`
	Responsibilities []string               `json:"responsibilities"`
	LocalizedName    map[string]string      `json:"localized_name"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ContactInfo represents contact information
type ContactInfo struct {
	Address string `json:"address,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Email   string `json:"email,omitempty"`
	Website string `json:"website,omitempty"`
}

// DataResidencyRules represents data residency rules for a country
type DataResidencyRules struct {
	RequiresLocalStorage bool     `json:"requires_local_storage"`
	AllowedRegions       []string `json:"allowed_regions"`
	RestrictedRegions    []string `json:"restricted_regions"`
	CrossBorderTransfer  bool     `json:"cross_border_transfer"`
	TransferRequirements []string `json:"transfer_requirements"`
	RetentionPeriod      int      `json:"retention_period_days"`
	DeletionRequirements []string `json:"deletion_requirements"`
}

// BusinessType represents a business type for a country
type BusinessType struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Code          string                 `json:"code"`
	Category      string                 `json:"category"`
	RiskLevel     string                 `json:"risk_level"`
	Requirements  []string               `json:"requirements"`
	LocalizedName map[string]string      `json:"localized_name"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// DocumentType represents a document type for a country
type DocumentType struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Code           string                 `json:"code"`
	Category       string                 `json:"category"`
	Required       bool                   `json:"required"`
	ValidityPeriod int                    `json:"validity_period_days"`
	Format         string                 `json:"format"`
	LocalizedName  map[string]string      `json:"localized_name"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// CountryRiskAssessment represents a country-specific risk assessment
type CountryRiskAssessment struct {
	ID               string                  `json:"id" db:"id"`
	CountryCode      string                  `json:"country_code" db:"country_code"`
	BusinessID       string                  `json:"business_id" db:"business_id"`
	TenantID         string                  `json:"tenant_id" db:"tenant_id"`
	RiskScore        float64                 `json:"risk_score" db:"risk_score"`
	RiskLevel        string                  `json:"risk_level" db:"risk_level"`
	RiskFactors      []AssessmentRiskFactor  `json:"risk_factors" db:"risk_factors"`
	ComplianceStatus []ComplianceStatus      `json:"compliance_status" db:"compliance_status"`
	Recommendations  []CountryRecommendation `json:"recommendations" db:"recommendations"`
	AssessmentDate   time.Time               `json:"assessment_date" db:"assessment_date"`
	ExpiryDate       time.Time               `json:"expiry_date" db:"expiry_date"`
	IsActive         bool                    `json:"is_active" db:"is_active"`
	Metadata         map[string]interface{}  `json:"metadata" db:"metadata"`
	CreatedAt        time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at" db:"updated_at"`
}

// AssessmentRiskFactor represents a country-specific risk factor in an assessment
type AssessmentRiskFactor struct {
	ID           string                 `json:"id"`
	RiskFactorID string                 `json:"risk_factor_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Severity     string                 `json:"severity"`
	Impact       string                 `json:"impact"`
	Mitigation   string                 `json:"mitigation"`
	Score        float64                `json:"score"`
	Weight       float64                `json:"weight"`
	IsActive     bool                   `json:"is_active"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ComplianceStatus represents compliance status for a specific rule
type ComplianceStatus struct {
	ID             string                 `json:"id"`
	RuleID         string                 `json:"rule_id"`
	RuleName       string                 `json:"rule_name"`
	Status         string                 `json:"status"`
	ComplianceDate *time.Time             `json:"compliance_date,omitempty"`
	ExpiryDate     *time.Time             `json:"expiry_date,omitempty"`
	Notes          string                 `json:"notes,omitempty"`
	Evidence       []string               `json:"evidence,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// CountryRecommendation represents a country-specific recommendation
type CountryRecommendation struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Priority      string                 `json:"priority"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Action        string                 `json:"action"`
	Timeline      string                 `json:"timeline"`
	Status        string                 `json:"status"`
	AssignedTo    string                 `json:"assigned_to,omitempty"`
	DueDate       *time.Time             `json:"due_date,omitempty"`
	CompletedDate *time.Time             `json:"completed_date,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// CountryBusinessData represents business data specific to a country
type CountryBusinessData struct {
	ID                 string                 `json:"id" db:"id"`
	BusinessID         string                 `json:"business_id" db:"business_id"`
	CountryCode        string                 `json:"country_code" db:"country_code"`
	TenantID           string                 `json:"tenant_id" db:"tenant_id"`
	RegistrationNumber string                 `json:"registration_number" db:"registration_number"`
	TaxID              string                 `json:"tax_id" db:"tax_id"`
	BusinessType       string                 `json:"business_type" db:"business_type"`
	Industry           string                 `json:"industry" db:"industry"`
	Address            Address                `json:"address" db:"address"`
	ContactInfo        ContactInfo            `json:"contact_info" db:"contact_info"`
	Documents          []BusinessDocument     `json:"documents" db:"documents"`
	ComplianceData     map[string]interface{} `json:"compliance_data" db:"compliance_data"`
	ValidationStatus   string                 `json:"validation_status" db:"validation_status"`
	ValidationDate     *time.Time             `json:"validation_date,omitempty" db:"validation_date"`
	IsActive           bool                   `json:"is_active" db:"is_active"`
	Metadata           map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
}

// Address represents an address
type Address struct {
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

// BusinessDocument represents a business document
type BusinessDocument struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	FileURL        string                 `json:"file_url"`
	FileHash       string                 `json:"file_hash"`
	FileSize       int64                  `json:"file_size"`
	MimeType       string                 `json:"mime_type"`
	UploadDate     time.Time              `json:"upload_date"`
	ExpiryDate     *time.Time             `json:"expiry_date,omitempty"`
	IsValid        bool                   `json:"is_valid"`
	ValidationDate *time.Time             `json:"validation_date,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// CountryComplianceReport represents a country-specific compliance report
type CountryComplianceReport struct {
	ID                string                  `json:"id" db:"id"`
	CountryCode       string                  `json:"country_code" db:"country_code"`
	TenantID          string                  `json:"tenant_id" db:"tenant_id"`
	ReportType        string                  `json:"report_type" db:"report_type"`
	Period            string                  `json:"period" db:"period"`
	StartDate         time.Time               `json:"start_date" db:"start_date"`
	EndDate           time.Time               `json:"end_date" db:"end_date"`
	Status            string                  `json:"status" db:"status"`
	ComplianceScore   float64                 `json:"compliance_score" db:"compliance_score"`
	TotalRules        int                     `json:"total_rules" db:"total_rules"`
	CompliantRules    int                     `json:"compliant_rules" db:"compliant_rules"`
	NonCompliantRules int                     `json:"non_compliant_rules" db:"non_compliant_rules"`
	PendingRules      int                     `json:"pending_rules" db:"pending_rules"`
	Violations        []ComplianceViolation   `json:"violations" db:"violations"`
	Recommendations   []CountryRecommendation `json:"recommendations" db:"recommendations"`
	GeneratedAt       time.Time               `json:"generated_at" db:"generated_at"`
	GeneratedBy       string                  `json:"generated_by" db:"generated_by"`
	Metadata          map[string]interface{}  `json:"metadata" db:"metadata"`
	CreatedAt         time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at" db:"updated_at"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	ID           string                 `json:"id"`
	RuleID       string                 `json:"rule_id"`
	RuleName     string                 `json:"rule_name"`
	Severity     string                 `json:"severity"`
	Description  string                 `json:"description"`
	DetectedDate time.Time              `json:"detected_date"`
	ResolvedDate *time.Time             `json:"resolved_date,omitempty"`
	Status       string                 `json:"status"`
	Penalty      string                 `json:"penalty,omitempty"`
	Remediation  string                 `json:"remediation,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}
