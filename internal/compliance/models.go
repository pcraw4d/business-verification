package compliance

import (
	"time"
)

// ComplianceRequirement represents a compliance requirement
type ComplianceRequirement struct {
	ID                   string                  `json:"id"`
	Framework            string                  `json:"framework"`      // "SOC2", "PCIDSS", "GDPR", etc.
	Category             string                  `json:"category"`       // "Security", "Privacy", "Financial", etc.
	RequirementID        string                  `json:"requirement_id"` // Framework-specific ID
	Title                string                  `json:"title"`
	Description          string                  `json:"description"`
	DetailedDescription  string                  `json:"detailed_description"`
	RiskLevel            ComplianceRiskLevel     `json:"risk_level"`
	Priority             CompliancePriority      `json:"priority"`
	Status               ComplianceStatus        `json:"status"`
	ImplementationStatus ImplementationStatus    `json:"implementation_status"`
	EvidenceRequired     bool                    `json:"evidence_required"`
	EvidenceDescription  string                  `json:"evidence_description"`
	Controls             []ComplianceControl     `json:"controls"`
	SubRequirements      []ComplianceRequirement `json:"sub_requirements,omitempty"`
	ParentRequirementID  *string                 `json:"parent_requirement_id,omitempty"`
	ApplicableBusinesses []string                `json:"applicable_businesses"` // Business types/categories
	GeographicScope      []string                `json:"geographic_scope"`      // Countries/regions
	IndustryScope        []string                `json:"industry_scope"`        // Industry codes
	EffectiveDate        time.Time               `json:"effective_date"`
	LastUpdated          time.Time               `json:"last_updated"`
	NextReviewDate       time.Time               `json:"next_review_date"`
	ReviewFrequency      string                  `json:"review_frequency"` // "monthly", "quarterly", "annually"
	ComplianceOfficer    string                  `json:"compliance_officer"`
	Tags                 []string                `json:"tags"`
	Metadata             map[string]interface{}  `json:"metadata,omitempty"`
}

// ComplianceRiskLevel represents the risk level of a compliance requirement
type ComplianceRiskLevel string

const (
	ComplianceRiskLevelLow      ComplianceRiskLevel = "low"
	ComplianceRiskLevelMedium   ComplianceRiskLevel = "medium"
	ComplianceRiskLevelHigh     ComplianceRiskLevel = "high"
	ComplianceRiskLevelCritical ComplianceRiskLevel = "critical"
)

// CompliancePriority represents the priority of a compliance requirement
type CompliancePriority string

const (
	CompliancePriorityLow      CompliancePriority = "low"
	CompliancePriorityMedium   CompliancePriority = "medium"
	CompliancePriorityHigh     CompliancePriority = "high"
	CompliancePriorityCritical CompliancePriority = "critical"
)

// ComplianceStatus represents the status of a compliance requirement
type ComplianceStatus string

const (
	ComplianceStatusNotStarted   ComplianceStatus = "not_started"
	ComplianceStatusInProgress   ComplianceStatus = "in_progress"
	ComplianceStatusImplemented  ComplianceStatus = "implemented"
	ComplianceStatusVerified     ComplianceStatus = "verified"
	ComplianceStatusNonCompliant ComplianceStatus = "non_compliant"
	ComplianceStatusExempt       ComplianceStatus = "exempt"
)

// ImplementationStatus represents the implementation status
type ImplementationStatus string

const (
	ImplementationStatusNotImplemented ImplementationStatus = "not_implemented"
	ImplementationStatusPlanned        ImplementationStatus = "planned"
	ImplementationStatusInProgress     ImplementationStatus = "in_progress"
	ImplementationStatusImplemented    ImplementationStatus = "implemented"
	ImplementationStatusTested         ImplementationStatus = "tested"
	ImplementationStatusDeployed       ImplementationStatus = "deployed"
)

// ComplianceControl represents a control for a compliance requirement
type ComplianceControl struct {
	ID                   string                 `json:"id"`
	RequirementID        string                 `json:"requirement_id"`
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	ControlType          ControlType            `json:"control_type"`
	ImplementationStatus ImplementationStatus   `json:"implementation_status"`
	Effectiveness        ControlEffectiveness   `json:"effectiveness"`
	TestingFrequency     string                 `json:"testing_frequency"`
	LastTested           *time.Time             `json:"last_tested,omitempty"`
	NextTestDate         *time.Time             `json:"next_test_date,omitempty"`
	TestResults          []ControlTestResult    `json:"test_results,omitempty"`
	Evidence             []ControlEvidence      `json:"evidence,omitempty"`
	ResponsibleParty     string                 `json:"responsible_party"`
	Automated            bool                   `json:"automated"`
	Frequency            string                 `json:"frequency"` // "continuous", "daily", "weekly", "monthly"
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// ControlType represents the type of control
type ControlType string

const (
	ControlTypePreventive   ControlType = "preventive"
	ControlTypeDetective    ControlType = "detective"
	ControlTypeCorrective   ControlType = "corrective"
	ControlTypeCompensating ControlType = "compensating"
)

// ControlEffectiveness represents the effectiveness of a control
type ControlEffectiveness string

const (
	ControlEffectivenessIneffective        ControlEffectiveness = "ineffective"
	ControlEffectivenessPartiallyEffective ControlEffectiveness = "partially_effective"
	ControlEffectivenessEffective          ControlEffectiveness = "effective"
	ControlEffectivenessHighlyEffective    ControlEffectiveness = "highly_effective"
)

// ControlTestResult represents a test result for a control
type ControlTestResult struct {
	ID          string     `json:"id"`
	ControlID   string     `json:"control_id"`
	TestDate    time.Time  `json:"test_date"`
	TestType    string     `json:"test_type"`
	Result      TestResult `json:"result"`
	Description string     `json:"description"`
	Evidence    string     `json:"evidence"`
	Tester      string     `json:"tester"`
	Notes       string     `json:"notes"`
}

// TestResult represents the result of a test
type TestResult string

const (
	TestResultPass          TestResult = "pass"
	TestResultFail          TestResult = "fail"
	TestResultPartial       TestResult = "partial"
	TestResultNotApplicable TestResult = "not_applicable"
)

// ControlEvidence represents evidence for a control
type ControlEvidence struct {
	ID          string     `json:"id"`
	ControlID   string     `json:"control_id"`
	Type        string     `json:"type"` // "document", "screenshot", "log", "report"
	Title       string     `json:"title"`
	Description string     `json:"description"`
	URL         string     `json:"url,omitempty"`
	FileSize    int64      `json:"file_size,omitempty"`
	UploadedAt  time.Time  `json:"uploaded_at"`
	UploadedBy  string     `json:"uploaded_by"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// ComplianceTracking represents compliance tracking for a business
type ComplianceTracking struct {
	ID                  string                 `json:"id"`
	BusinessID          string                 `json:"business_id"`
	Framework           string                 `json:"framework"`
	OverallStatus       ComplianceStatus       `json:"overall_status"`
	ComplianceScore     float64                `json:"compliance_score"` // 0.0 to 100.0
	Requirements        []RequirementTracking  `json:"requirements"`
	LastAssessment      time.Time              `json:"last_assessment"`
	NextAssessment      time.Time              `json:"next_assessment"`
	AssessmentFrequency string                 `json:"assessment_frequency"`
	ComplianceOfficer   string                 `json:"compliance_officer"`
	Auditor             string                 `json:"auditor,omitempty"`
	CertificationDate   *time.Time             `json:"certification_date,omitempty"`
	CertificationExpiry *time.Time             `json:"certification_expiry,omitempty"`
	CertificationBody   string                 `json:"certification_body,omitempty"`
	CertificationNumber string                 `json:"certification_number,omitempty"`
	Notes               string                 `json:"notes"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// RequirementTracking represents tracking for a specific requirement
type RequirementTracking struct {
	RequirementID        string                `json:"requirement_id"`
	Status               ComplianceStatus      `json:"status"`
	ImplementationStatus ImplementationStatus  `json:"implementation_status"`
	ComplianceScore      float64               `json:"compliance_score"`
	LastReviewed         time.Time             `json:"last_reviewed"`
	NextReview           time.Time             `json:"next_review"`
	Reviewer             string                `json:"reviewer"`
	Notes                string                `json:"notes"`
	Evidence             []TrackingEvidence    `json:"evidence"`
	Controls             []ControlTracking     `json:"controls"`
	Exceptions           []ComplianceException `json:"exceptions"`
	RemediationPlan      *RemediationPlan      `json:"remediation_plan,omitempty"`
}

// TrackingEvidence represents evidence for requirement tracking
type TrackingEvidence struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	URL         string     `json:"url,omitempty"`
	UploadedAt  time.Time  `json:"uploaded_at"`
	UploadedBy  string     `json:"uploaded_by"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// ControlTracking represents tracking for a specific control
type ControlTracking struct {
	ControlID            string               `json:"control_id"`
	Status               ComplianceStatus     `json:"status"`
	ImplementationStatus ImplementationStatus `json:"implementation_status"`
	Effectiveness        ControlEffectiveness `json:"effectiveness"`
	LastTested           *time.Time           `json:"last_tested,omitempty"`
	NextTestDate         *time.Time           `json:"next_test_date,omitempty"`
	TestResults          []ControlTestResult  `json:"test_results"`
	Evidence             []ControlEvidence    `json:"evidence"`
	Notes                string               `json:"notes"`
}

// ComplianceException represents an exception to a compliance requirement
type ComplianceException struct {
	ID             string          `json:"id"`
	RequirementID  string          `json:"requirement_id"`
	Type           ExceptionType   `json:"type"`
	Reason         string          `json:"reason"`
	Justification  string          `json:"justification"`
	RiskAssessment string          `json:"risk_assessment"`
	MitigationPlan string          `json:"mitigation_plan"`
	ApprovedBy     string          `json:"approved_by"`
	ApprovedAt     time.Time       `json:"approved_at"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty"`
	Status         ExceptionStatus `json:"status"`
	Notes          string          `json:"notes"`
}

// ExceptionType represents the type of exception
type ExceptionType string

const (
	ExceptionTypeTemporary ExceptionType = "temporary"
	ExceptionTypePermanent ExceptionType = "permanent"
	ExceptionTypePartial   ExceptionType = "partial"
)

// ExceptionStatus represents the status of an exception
type ExceptionStatus string

const (
	ExceptionStatusPending  ExceptionStatus = "pending"
	ExceptionStatusApproved ExceptionStatus = "approved"
	ExceptionStatusRejected ExceptionStatus = "rejected"
	ExceptionStatusExpired  ExceptionStatus = "expired"
)

// RemediationPlan represents a plan to remediate compliance issues
type RemediationPlan struct {
	ID                   string              `json:"id"`
	RequirementID        string              `json:"requirement_id"`
	Title                string              `json:"title"`
	Description          string              `json:"description"`
	Priority             CompliancePriority  `json:"priority"`
	Status               RemediationStatus   `json:"status"`
	TargetDate           time.Time           `json:"target_date"`
	ActualCompletionDate *time.Time          `json:"actual_completion_date,omitempty"`
	AssignedTo           string              `json:"assigned_to"`
	Budget               float64             `json:"budget,omitempty"`
	ActualCost           float64             `json:"actual_cost,omitempty"`
	Actions              []RemediationAction `json:"actions"`
	Progress             float64             `json:"progress"` // 0.0 to 100.0
	Notes                string              `json:"notes"`
}

// RemediationStatus represents the status of a remediation plan
type RemediationStatus string

const (
	RemediationStatusNotStarted RemediationStatus = "not_started"
	RemediationStatusInProgress RemediationStatus = "in_progress"
	RemediationStatusCompleted  RemediationStatus = "completed"
	RemediationStatusOnHold     RemediationStatus = "on_hold"
	RemediationStatusCancelled  RemediationStatus = "cancelled"
)

// RemediationAction represents an action in a remediation plan
type RemediationAction struct {
	ID                   string             `json:"id"`
	PlanID               string             `json:"plan_id"`
	Title                string             `json:"title"`
	Description          string             `json:"description"`
	Status               RemediationStatus  `json:"status"`
	Priority             CompliancePriority `json:"priority"`
	AssignedTo           string             `json:"assigned_to"`
	StartDate            time.Time          `json:"start_date"`
	TargetDate           time.Time          `json:"target_date"`
	ActualCompletionDate *time.Time         `json:"actual_completion_date,omitempty"`
	Progress             float64            `json:"progress"`
	Notes                string             `json:"notes"`
}

// RegulatoryFramework represents a regulatory framework
type RegulatoryFramework struct {
	ID              string                  `json:"id"`
	Name            string                  `json:"name"`
	Version         string                  `json:"version"`
	Description     string                  `json:"description"`
	Type            FrameworkType           `json:"type"`
	Jurisdiction    string                  `json:"jurisdiction"`
	GeographicScope []string                `json:"geographic_scope"`
	IndustryScope   []string                `json:"industry_scope"`
	EffectiveDate   time.Time               `json:"effective_date"`
	LastUpdated     time.Time               `json:"last_updated"`
	NextReviewDate  time.Time               `json:"next_review_date"`
	Requirements    []ComplianceRequirement `json:"requirements"`
	MappingRules    []FrameworkMapping      `json:"mapping_rules"`
	Metadata        map[string]interface{}  `json:"metadata,omitempty"`
}

// FrameworkType represents the type of regulatory framework
type FrameworkType string

const (
	FrameworkTypeSecurity    FrameworkType = "security"
	FrameworkTypePrivacy     FrameworkType = "privacy"
	FrameworkTypeFinancial   FrameworkType = "financial"
	FrameworkTypeOperational FrameworkType = "operational"
	FrameworkTypeIndustry    FrameworkType = "industry"
)

// FrameworkMapping represents a mapping between frameworks
type FrameworkMapping struct {
	ID                  string      `json:"id"`
	SourceFramework     string      `json:"source_framework"`
	SourceRequirementID string      `json:"source_requirement_id"`
	TargetFramework     string      `json:"target_framework"`
	TargetRequirementID string      `json:"target_requirement_id"`
	MappingType         MappingType `json:"mapping_type"`
	Confidence          float64     `json:"confidence"` // 0.0 to 1.0
	Notes               string      `json:"notes"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
}

// MappingType represents the type of framework mapping
type MappingType string

const (
	MappingTypeExact      MappingType = "exact"
	MappingTypePartial    MappingType = "partial"
	MappingTypeRelated    MappingType = "related"
	MappingTypeSuperseded MappingType = "superseded"
)

// ComplianceAuditTrail represents an audit trail entry
type ComplianceAuditTrail struct {
	ID            string                 `json:"id"`
	BusinessID    string                 `json:"business_id"`
	Framework     string                 `json:"framework"`
	RequirementID *string                `json:"requirement_id,omitempty"`
	ControlID     *string                `json:"control_id,omitempty"`
	Action        AuditAction            `json:"action"`
	Description   string                 `json:"description"`
	UserID        string                 `json:"user_id"`
	UserName      string                 `json:"user_name"`
	UserRole      string                 `json:"user_role"`
	Timestamp     time.Time              `json:"timestamp"`
	IPAddress     string                 `json:"ip_address"`
	UserAgent     string                 `json:"user_agent"`
	SessionID     string                 `json:"session_id"`
	RequestID     string                 `json:"request_id"`
	OldValue      interface{}            `json:"old_value,omitempty"`
	NewValue      interface{}            `json:"new_value,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AuditAction represents the type of audit action
type AuditAction string

const (
	AuditActionCreate      AuditAction = "create"
	AuditActionRead        AuditAction = "read"
	AuditActionUpdate      AuditAction = "update"
	AuditActionDelete      AuditAction = "delete"
	AuditActionLogin       AuditAction = "login"
	AuditActionLogout      AuditAction = "logout"
	AuditActionExport      AuditAction = "export"
	AuditActionImport      AuditAction = "import"
	AuditActionApprove     AuditAction = "approve"
	AuditActionReject      AuditAction = "reject"
	AuditActionException   AuditAction = "exception"
	AuditActionRemediation AuditAction = "remediation"
)

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	ID               string                     `json:"id"`
	BusinessID       string                     `json:"business_id"`
	Framework        string                     `json:"framework"`
	ReportType       ReportType                 `json:"report_type"`
	Title            string                     `json:"title"`
	Description      string                     `json:"description"`
	GeneratedAt      time.Time                  `json:"generated_at"`
	GeneratedBy      string                     `json:"generated_by"`
	Period           string                     `json:"period"`
	OverallStatus    ComplianceStatus           `json:"overall_status"`
	ComplianceScore  float64                    `json:"compliance_score"`
	Requirements     []RequirementReport        `json:"requirements"`
	Controls         []ControlReport            `json:"controls"`
	Exceptions       []ExceptionReport          `json:"exceptions"`
	RemediationPlans []RemediationReport        `json:"remediation_plans"`
	Recommendations  []ComplianceRecommendation `json:"recommendations"`
	Metadata         map[string]interface{}     `json:"metadata,omitempty"`
}

// ReportType represents the type of compliance report
type ReportType string

const (
	ReportTypeStatus      ReportType = "status"
	ReportTypeGap         ReportType = "gap"
	ReportTypeRemediation ReportType = "remediation"
	ReportTypeAudit       ReportType = "audit"
	ReportTypeExecutive   ReportType = "executive"
)

// RequirementReport represents a requirement in a compliance report
type RequirementReport struct {
	RequirementID        string               `json:"requirement_id"`
	Title                string               `json:"title"`
	Status               ComplianceStatus     `json:"status"`
	ImplementationStatus ImplementationStatus `json:"implementation_status"`
	ComplianceScore      float64              `json:"compliance_score"`
	RiskLevel            ComplianceRiskLevel  `json:"risk_level"`
	Priority             CompliancePriority   `json:"priority"`
	LastReviewed         time.Time            `json:"last_reviewed"`
	NextReview           time.Time            `json:"next_review"`
	Controls             []ControlReport      `json:"controls"`
	Exceptions           []ExceptionReport    `json:"exceptions"`
	RemediationPlans     []RemediationReport  `json:"remediation_plans"`
}

// ControlReport represents a control in a compliance report
type ControlReport struct {
	ControlID            string               `json:"control_id"`
	Title                string               `json:"title"`
	Status               ComplianceStatus     `json:"status"`
	ImplementationStatus ImplementationStatus `json:"implementation_status"`
	Effectiveness        ControlEffectiveness `json:"effectiveness"`
	LastTested           *time.Time           `json:"last_tested,omitempty"`
	NextTestDate         *time.Time           `json:"next_test_date,omitempty"`
	TestResults          []ControlTestResult  `json:"test_results"`
	Evidence             []ControlEvidence    `json:"evidence"`
}

// ExceptionReport represents an exception in a compliance report
type ExceptionReport struct {
	ExceptionID   string          `json:"exception_id"`
	RequirementID string          `json:"requirement_id"`
	Type          ExceptionType   `json:"type"`
	Reason        string          `json:"reason"`
	Status        ExceptionStatus `json:"status"`
	ApprovedBy    string          `json:"approved_by"`
	ApprovedAt    time.Time       `json:"approved_at"`
	ExpiresAt     *time.Time      `json:"expires_at,omitempty"`
}

// RemediationReport represents a remediation plan in a compliance report
type RemediationReport struct {
	PlanID     string             `json:"plan_id"`
	Title      string             `json:"title"`
	Status     RemediationStatus  `json:"status"`
	Priority   CompliancePriority `json:"priority"`
	TargetDate time.Time          `json:"target_date"`
	Progress   float64            `json:"progress"`
	AssignedTo string             `json:"assigned_to"`
}

// ComplianceRecommendation represents a compliance recommendation
type ComplianceRecommendation struct {
	ID            string               `json:"id"`
	Type          RecommendationType   `json:"type"`
	Priority      CompliancePriority   `json:"priority"`
	Title         string               `json:"title"`
	Description   string               `json:"description"`
	Action        string               `json:"action"`
	Timeline      string               `json:"timeline"`
	Impact        string               `json:"impact"`
	Effort        string               `json:"effort"`
	Cost          float64              `json:"cost,omitempty"`
	RequirementID *string              `json:"requirement_id,omitempty"`
	ControlID     *string              `json:"control_id,omitempty"`
	AssignedTo    string               `json:"assigned_to"`
	Status        RecommendationStatus `json:"status"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// RecommendationType represents the type of recommendation
type RecommendationType string

const (
	RecommendationTypeImplementation RecommendationType = "implementation"
	RecommendationTypeTesting        RecommendationType = "testing"
	RecommendationTypeDocumentation  RecommendationType = "documentation"
	RecommendationTypeTraining       RecommendationType = "training"
	RecommendationTypeProcess        RecommendationType = "process"
	RecommendationTypeTechnology     RecommendationType = "technology"
)

// RecommendationStatus represents the status of a recommendation
type RecommendationStatus string

const (
	RecommendationStatusOpen       RecommendationStatus = "open"
	RecommendationStatusInProgress RecommendationStatus = "in_progress"
	RecommendationStatusCompleted  RecommendationStatus = "completed"
	RecommendationStatusRejected   RecommendationStatus = "rejected"
	RecommendationStatusOnHold     RecommendationStatus = "on_hold"
)
