package compliance

import (
	"time"
)

// RegionalFramework defines the regional compliance frameworks
const (
	// Regional Framework Types
	FrameworkCCPA   = "CCPA"
	FrameworkLGPD   = "LGPD"
	FrameworkPIPEDA = "PIPEDA"
	FrameworkPOPIA  = "POPIA"
	FrameworkPDPA   = "PDPA"
	FrameworkAPPI   = "APPI"

	// CCPA Versions
	CCPAVersion2020 = "2020"
	CCPAVersion2023 = "2023"

	// LGPD Versions
	LGPDVersion2020 = "2020"
	LGPDVersion2021 = "2021"

	// PIPEDA Versions
	PIPEDAVersion2000 = "2000"
	PIPEDAVersion2015 = "2015"

	// POPIA Versions
	POPIAVersion2021 = "2021"

	// PDPA Versions
	PDPAVersion2012 = "2012"
	PDPAVersion2021 = "2021"

	// APPI Versions
	APPIVersion2003 = "2003"
	APPIVersion2022 = "2022"

	// CCPA Categories
	CCPACategoryConsumerRights      = "Consumer Rights"
	CCPACategoryBusinessObligations = "Business Obligations"
	CCPACategoryDataTransparency    = "Data Transparency"
	CCPACategoryEnforcement         = "Enforcement"

	// LGPD Categories
	LGPDCategoryLegalBasis          = "Legal Basis for Processing"
	LGPDCategoryDataSubjectRights   = "Data Subject Rights"
	LGPDCategoryBusinessObligations = "Business Obligations"
	LGPDCategoryDataProtection      = "Data Protection"

	// PIPEDA Categories
	PIPEDACategoryConsent               = "Consent"
	PIPEDACategoryLimitingCollection    = "Limiting Collection"
	PIPEDACategoryLimitingUse           = "Limiting Use, Disclosure, and Retention"
	PIPEDACategoryAccuracy              = "Accuracy"
	PIPEDACategorySafeguards            = "Safeguards"
	PIPEDACategoryOpenness              = "Openness"
	PIPEDACategoryIndividualAccess      = "Individual Access"
	PIPEDACategoryChallengingCompliance = "Challenging Compliance"
)

// RegionalRequirement represents a regional compliance requirement
type RegionalRequirement struct {
	ID                   string                 `json:"id"`
	RequirementID        string                 `json:"requirement_id"` // e.g., CCPA-1798.100, LGPD-Art.5
	Framework            string                 `json:"framework"`      // CCPA, LGPD, PIPEDA, etc.
	Category             string                 `json:"category"`       // Consumer Rights, Business Obligations, etc.
	Section              string                 `json:"section"`        // Legal section reference
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	DetailedDescription  string                 `json:"detailed_description"`
	LegalBasis           []string               `json:"legal_basis"`
	DataSubjectRights    []string               `json:"data_subject_rights"`
	RiskLevel            ComplianceRiskLevel    `json:"risk_level"`
	Priority             CompliancePriority     `json:"priority"`
	ImplementationStatus ImplementationStatus   `json:"implementation_status"`
	EvidenceRequired     bool                   `json:"evidence_required"`
	EvidenceDescription  string                 `json:"evidence_description"`
	KeyControls          []string               `json:"key_controls"`
	SubRequirements      []RegionalRequirement  `json:"sub_requirements,omitempty"`
	ParentRequirementID  *string                `json:"parent_requirement_id,omitempty"`
	EffectiveDate        time.Time              `json:"effective_date"`
	LastUpdated          time.Time              `json:"last_updated"`
	NextReviewDate       time.Time              `json:"next_review_date"`
	ReviewFrequency      string                 `json:"review_frequency"`
	ComplianceOfficer    string                 `json:"compliance_officer"`
	Tags                 []string               `json:"tags"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// RegionalFrameworkDefinition represents a regional compliance framework
type RegionalFrameworkDefinition struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Description     string                 `json:"description"`
	Type            FrameworkType          `json:"type"`
	Jurisdiction    string                 `json:"jurisdiction"`
	GeographicScope []string               `json:"geographic_scope"`
	IndustryScope   []string               `json:"industry_scope"`
	EffectiveDate   time.Time              `json:"effective_date"`
	LastUpdated     time.Time              `json:"last_updated"`
	NextReviewDate  time.Time              `json:"next_review_date"`
	Requirements    []RegionalRequirement  `json:"requirements"`
	Categories      []RegionalCategory     `json:"categories"`
	MappingRules    []FrameworkMapping     `json:"mapping_rules"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// RegionalCategory represents a category within a regional framework
type RegionalCategory struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Code          string                 `json:"code"`
	Description   string                 `json:"description"`
	Requirements  []string               `json:"requirements"` // Requirement IDs
	RiskLevel     ComplianceRiskLevel    `json:"risk_level"`
	Priority      CompliancePriority     `json:"priority"`
	EffectiveDate time.Time              `json:"effective_date"`
	LastUpdated   time.Time              `json:"last_updated"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// RegionalComplianceStatus represents regional compliance status
type RegionalComplianceStatus struct {
	BusinessID          string                               `json:"business_id"`
	Framework           string                               `json:"framework"`
	Version             string                               `json:"version"`
	Jurisdiction        string                               `json:"jurisdiction"`
	DataController      bool                                 `json:"data_controller"`
	DataProcessor       bool                                 `json:"data_processor"`
	OverallStatus       ComplianceStatus                     `json:"overall_status"`
	ComplianceScore     float64                              `json:"compliance_score"`
	CategoryStatus      map[string]RegionalCategoryStatus    `json:"category_status"`
	RequirementsStatus  map[string]RegionalRequirementStatus `json:"requirements_status"`
	LastAssessment      time.Time                            `json:"last_assessment"`
	NextAssessment      time.Time                            `json:"next_assessment"`
	AssessmentFrequency string                               `json:"assessment_frequency"`
	ComplianceOfficer   string                               `json:"compliance_officer"`
	RegulatoryAuthority string                               `json:"regulatory_authority,omitempty"`
	CertificationDate   *time.Time                           `json:"certification_date,omitempty"`
	CertificationExpiry *time.Time                           `json:"certification_expiry,omitempty"`
	CertificationBody   string                               `json:"certification_body,omitempty"`
	CertificationNumber string                               `json:"certification_number,omitempty"`
	Notes               string                               `json:"notes"`
	Metadata            map[string]interface{}               `json:"metadata,omitempty"`
}

// RegionalCategoryStatus represents status for a specific regional category
type RegionalCategoryStatus struct {
	CategoryID        string           `json:"category_id"`
	CategoryName      string           `json:"category_name"`
	Status            ComplianceStatus `json:"status"`
	Score             float64          `json:"score"`
	RequirementCount  int              `json:"requirement_count"`
	ImplementedCount  int              `json:"implemented_count"`
	VerifiedCount     int              `json:"verified_count"`
	NonCompliantCount int              `json:"non_compliant_count"`
	ExemptCount       int              `json:"exempt_count"`
	LastReviewed      time.Time        `json:"last_reviewed"`
	NextReview        time.Time        `json:"next_review"`
	Reviewer          string           `json:"reviewer"`
	Notes             string           `json:"notes"`
}

// RegionalRequirementStatus represents status for a specific regional requirement
type RegionalRequirementStatus struct {
	RequirementID        string               `json:"requirement_id"`
	FrameworkID          string               `json:"framework_id"`
	CategoryID           string               `json:"category_id"`
	Title                string               `json:"title"`
	Status               ComplianceStatus     `json:"status"`
	ImplementationStatus ImplementationStatus `json:"implementation_status"`
	ComplianceScore      float64              `json:"compliance_score"`
	RiskLevel            ComplianceRiskLevel  `json:"risk_level"`
	Priority             CompliancePriority   `json:"priority"`
	LastReviewed         time.Time            `json:"last_reviewed"`
	NextReview           time.Time            `json:"next_review"`
	Reviewer             string               `json:"reviewer"`
	EvidenceCount        int                  `json:"evidence_count"`
	ExceptionCount       int                  `json:"exception_count"`
	RemediationPlanCount int                  `json:"remediation_plan_count"`
	Trend                string               `json:"trend"`
	TrendStrength        string               `json:"trend_strength"`
	Notes                string               `json:"notes"`
}

// NewCCPAFramework creates a new CCPA framework definition
func NewCCPAFramework() *RegionalFrameworkDefinition {
	return &RegionalFrameworkDefinition{
		ID:              FrameworkCCPA,
		Name:            "California Consumer Privacy Act",
		Version:         CCPAVersion2023,
		Description:     "The CCPA is a state statute intended to enhance privacy rights and consumer protection for residents of California",
		Type:            FrameworkTypePrivacy,
		Jurisdiction:    "California, United States",
		GeographicScope: []string{"California", "United States"},
		IndustryScope:   []string{"All Industries"},
		EffectiveDate:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		LastUpdated:     time.Now(),
		NextReviewDate:  time.Now().AddDate(1, 0, 0),
		Requirements:    getCCPARequirements(),
		Categories:      getCCPACategories(),
		Metadata:        make(map[string]interface{}),
	}
}

// NewLGPDFramework creates a new LGPD framework definition
func NewLGPDFramework() *RegionalFrameworkDefinition {
	return &RegionalFrameworkDefinition{
		ID:              FrameworkLGPD,
		Name:            "Lei Geral de Proteção de Dados",
		Version:         LGPDVersion2021,
		Description:     "The LGPD is Brazil's comprehensive data protection law that regulates the processing of personal data",
		Type:            FrameworkTypePrivacy,
		Jurisdiction:    "Brazil",
		GeographicScope: []string{"Brazil"},
		IndustryScope:   []string{"All Industries"},
		EffectiveDate:   time.Date(2021, 9, 18, 0, 0, 0, 0, time.UTC),
		LastUpdated:     time.Now(),
		NextReviewDate:  time.Now().AddDate(1, 0, 0),
		Requirements:    getLGPDRequirements(),
		Categories:      getLGPDCategories(),
		Metadata:        make(map[string]interface{}),
	}
}

// NewPIPEDAFramework creates a new PIPEDA framework definition
func NewPIPEDAFramework() *RegionalFrameworkDefinition {
	return &RegionalFrameworkDefinition{
		ID:              FrameworkPIPEDA,
		Name:            "Personal Information Protection and Electronic Documents Act",
		Version:         PIPEDAVersion2015,
		Description:     "PIPEDA is Canada's federal privacy law for private-sector organizations",
		Type:            FrameworkTypePrivacy,
		Jurisdiction:    "Canada",
		GeographicScope: []string{"Canada"},
		IndustryScope:   []string{"All Industries"},
		EffectiveDate:   time.Date(2015, 6, 18, 0, 0, 0, 0, time.UTC),
		LastUpdated:     time.Now(),
		NextReviewDate:  time.Now().AddDate(1, 0, 0),
		Requirements:    getPIPEDARequirements(),
		Categories:      getPIPEDACategories(),
		Metadata:        make(map[string]interface{}),
	}
}

// NewPOPIAFramework creates a new POPIA framework definition
func NewPOPIAFramework() *RegionalFrameworkDefinition {
	return &RegionalFrameworkDefinition{
		ID:              FrameworkPOPIA,
		Name:            "Protection of Personal Information Act",
		Version:         POPIAVersion2021,
		Description:     "POPIA is South Africa's comprehensive data protection law that regulates the processing of personal information",
		Type:            FrameworkTypePrivacy,
		Jurisdiction:    "South Africa",
		GeographicScope: []string{"South Africa"},
		IndustryScope:   []string{"All Industries"},
		EffectiveDate:   time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
		LastUpdated:     time.Now(),
		NextReviewDate:  time.Now().AddDate(1, 0, 0),
		Requirements:    getPOPIARequirements(),
		Categories:      getPOPIACategories(),
		Metadata:        make(map[string]interface{}),
	}
}

// NewPDPAFramework creates a new PDPA framework definition
func NewPDPAFramework() *RegionalFrameworkDefinition {
	return &RegionalFrameworkDefinition{
		ID:              FrameworkPDPA,
		Name:            "Personal Data Protection Act",
		Version:         PDPAVersion2021,
		Description:     "PDPA is Singapore's comprehensive data protection law that governs the collection, use, and disclosure of personal data",
		Type:            FrameworkTypePrivacy,
		Jurisdiction:    "Singapore",
		GeographicScope: []string{"Singapore"},
		IndustryScope:   []string{"All Industries"},
		EffectiveDate:   time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
		LastUpdated:     time.Now(),
		NextReviewDate:  time.Now().AddDate(1, 0, 0),
		Requirements:    getPDPARequirements(),
		Categories:      getPDPACategories(),
		Metadata:        make(map[string]interface{}),
	}
}

// NewAPPIFramework creates a new APPI framework definition
func NewAPPIFramework() *RegionalFrameworkDefinition {
	return &RegionalFrameworkDefinition{
		ID:              FrameworkAPPI,
		Name:            "Act on the Protection of Personal Information",
		Version:         APPIVersion2022,
		Description:     "APPI is Japan's comprehensive data protection law that regulates the handling of personal information",
		Type:            FrameworkTypePrivacy,
		Jurisdiction:    "Japan",
		GeographicScope: []string{"Japan"},
		IndustryScope:   []string{"All Industries"},
		EffectiveDate:   time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
		LastUpdated:     time.Now(),
		NextReviewDate:  time.Now().AddDate(1, 0, 0),
		Requirements:    getAPPIRequirements(),
		Categories:      getAPPICategories(),
		Metadata:        make(map[string]interface{}),
	}
}

// getCCPACategories returns the CCPA categories
func getCCPACategories() []RegionalCategory {
	return []RegionalCategory{
		{
			ID:           CCPACategoryConsumerRights,
			Name:         "Consumer Rights",
			Code:         "CCPA-Rights",
			Description:  "Consumer rights under CCPA including access, deletion, and opt-out",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"CCPA-1798.100", "CCPA-1798.105", "CCPA-1798.110", "CCPA-1798.115", "CCPA-1798.120", "CCPA-1798.125"},
		},
		{
			ID:           CCPACategoryBusinessObligations,
			Name:         "Business Obligations",
			Code:         "CCPA-Obligations",
			Description:  "Business obligations for data handling and consumer requests",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"CCPA-1798.130", "CCPA-1798.135", "CCPA-1798.140", "CCPA-1798.145"},
		},
		{
			ID:           CCPACategoryDataTransparency,
			Name:         "Data Transparency",
			Code:         "CCPA-Transparency",
			Description:  "Requirements for data transparency and disclosure",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"CCPA-1798.100", "CCPA-1798.130", "CCPA-1798.135"},
		},
		{
			ID:           CCPACategoryEnforcement,
			Name:         "Enforcement",
			Code:         "CCPA-Enforcement",
			Description:  "Enforcement mechanisms and penalties",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"CCPA-1798.150", "CCPA-1798.155", "CCPA-1798.160"},
		},
	}
}

// getLGPDCategories returns the LGPD categories
func getLGPDCategories() []RegionalCategory {
	return []RegionalCategory{
		{
			ID:           LGPDCategoryLegalBasis,
			Name:         "Legal Basis for Processing",
			Code:         "LGPD-Basis",
			Description:  "Legal basis for processing personal data under LGPD",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"LGPD-Art.7", "LGPD-Art.8", "LGPD-Art.9", "LGPD-Art.10", "LGPD-Art.11"},
		},
		{
			ID:           LGPDCategoryDataSubjectRights,
			Name:         "Data Subject Rights",
			Code:         "LGPD-Rights",
			Description:  "Data subject rights under LGPD",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"LGPD-Art.18", "LGPD-Art.19", "LGPD-Art.20", "LGPD-Art.21", "LGPD-Art.22"},
		},
		{
			ID:           LGPDCategoryBusinessObligations,
			Name:         "Business Obligations",
			Code:         "LGPD-Obligations",
			Description:  "Business obligations for data processing",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"LGPD-Art.6", "LGPD-Art.12", "LGPD-Art.13", "LGPD-Art.14", "LGPD-Art.15"},
		},
		{
			ID:           LGPDCategoryDataProtection,
			Name:         "Data Protection",
			Code:         "LGPD-Protection",
			Description:  "Data protection and security requirements",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"LGPD-Art.46", "LGPD-Art.47", "LGPD-Art.48", "LGPD-Art.49"},
		},
	}
}

// getPIPEDACategories returns the PIPEDA categories
func getPIPEDACategories() []RegionalCategory {
	return []RegionalCategory{
		{
			ID:           PIPEDACategoryConsent,
			Name:         "Consent",
			Code:         "PIPEDA-Consent",
			Description:  "Consent requirements for data collection and use",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"PIPEDA-Principle1", "PIPEDA-Principle2"},
		},
		{
			ID:           PIPEDACategoryLimitingCollection,
			Name:         "Limiting Collection",
			Code:         "PIPEDA-LimitingCollection",
			Description:  "Limiting collection of personal information",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"PIPEDA-Principle4", "PIPEDA-Principle5"},
		},
		{
			ID:           PIPEDACategoryLimitingUse,
			Name:         "Limiting Use, Disclosure, and Retention",
			Code:         "PIPEDA-LimitingUse",
			Description:  "Limiting use, disclosure, and retention of personal information",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"PIPEDA-Principle5", "PIPEDA-Principle6"},
		},
		{
			ID:           PIPEDACategoryAccuracy,
			Name:         "Accuracy",
			Code:         "PIPEDA-Accuracy",
			Description:  "Accuracy of personal information",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"PIPEDA-Principle6"},
		},
		{
			ID:           PIPEDACategorySafeguards,
			Name:         "Safeguards",
			Code:         "PIPEDA-Safeguards",
			Description:  "Security safeguards for personal information",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"PIPEDA-Principle7"},
		},
		{
			ID:           PIPEDACategoryOpenness,
			Name:         "Openness",
			Code:         "PIPEDA-Openness",
			Description:  "Openness about policies and practices",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"PIPEDA-Principle8"},
		},
		{
			ID:           PIPEDACategoryIndividualAccess,
			Name:         "Individual Access",
			Code:         "PIPEDA-IndividualAccess",
			Description:  "Individual access to personal information",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"PIPEDA-Principle9"},
		},
		{
			ID:           PIPEDACategoryChallengingCompliance,
			Name:         "Challenging Compliance",
			Code:         "PIPEDA-ChallengingCompliance",
			Description:  "Challenging compliance with PIPEDA principles",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"PIPEDA-Principle10"},
		},
	}
}

// getPOPIACategories returns the POPIA categories
func getPOPIACategories() []RegionalCategory {
	return []RegionalCategory{
		{
			ID:           "POPIA-ProcessingLimitation",
			Name:         "Processing Limitation",
			Code:         "POPIA-ProcessingLimitation",
			Description:  "Limitations on processing personal information",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"POPIA-Section9", "POPIA-Section10", "POPIA-Section11", "POPIA-Section12"},
		},
		{
			ID:           "POPIA-PurposeSpecification",
			Name:         "Purpose Specification",
			Code:         "POPIA-PurposeSpecification",
			Description:  "Specification of purpose for processing",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"POPIA-Section13", "POPIA-Section14", "POPIA-Section15"},
		},
		{
			ID:           "POPIA-InformationQuality",
			Name:         "Information Quality",
			Code:         "POPIA-InformationQuality",
			Description:  "Quality of personal information",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"POPIA-Section16", "POPIA-Section17"},
		},
		{
			ID:           "POPIA-Openness",
			Name:         "Openness",
			Code:         "POPIA-Openness",
			Description:  "Openness about processing operations",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"POPIA-Section18", "POPIA-Section19"},
		},
		{
			ID:           "POPIA-SecuritySafeguards",
			Name:         "Security Safeguards",
			Code:         "POPIA-SecuritySafeguards",
			Description:  "Security safeguards for personal information",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"POPIA-Section19", "POPIA-Section20", "POPIA-Section21"},
		},
		{
			ID:           "POPIA-DataSubjectParticipation",
			Name:         "Data Subject Participation",
			Code:         "POPIA-DataSubjectParticipation",
			Description:  "Data subject participation rights",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"POPIA-Section22", "POPIA-Section23", "POPIA-Section24", "POPIA-Section25"},
		},
	}
}

// getPDPACategories returns the PDPA categories
func getPDPACategories() []RegionalCategory {
	return []RegionalCategory{
		{
			ID:           "PDPA-Consent",
			Name:         "Consent",
			Code:         "PDPA-Consent",
			Description:  "Consent requirements for data collection and use",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"PDPA-Section13", "PDPA-Section14", "PDPA-Section15"},
		},
		{
			ID:           "PDPA-PurposeLimitation",
			Name:         "Purpose Limitation",
			Code:         "PDPA-PurposeLimitation",
			Description:  "Limitation of purpose for data processing",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"PDPA-Section25", "PDPA-Section26"},
		},
		{
			ID:           "PDPA-Notification",
			Name:         "Notification",
			Code:         "PDPA-Notification",
			Description:  "Notification requirements for data breaches",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"PDPA-Section26D", "PDPA-Section26E"},
		},
		{
			ID:           "PDPA-TransferLimitation",
			Name:         "Transfer Limitation",
			Code:         "PDPA-TransferLimitation",
			Description:  "Limitations on data transfers",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"PDPA-Section26", "PDPA-Section27"},
		},
		{
			ID:           "PDPA-AccessCorrection",
			Name:         "Access and Correction",
			Code:         "PDPA-AccessCorrection",
			Description:  "Access and correction rights",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"PDPA-Section21", "PDPA-Section22", "PDPA-Section23"},
		},
		{
			ID:           "PDPA-Accuracy",
			Name:         "Accuracy",
			Code:         "PDPA-Accuracy",
			Description:  "Accuracy of personal data",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"PDPA-Section23"},
		},
		{
			ID:           "PDPA-Protection",
			Name:         "Protection",
			Code:         "PDPA-Protection",
			Description:  "Protection of personal data",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"PDPA-Section24"},
		},
		{
			ID:           "PDPA-RetentionLimitation",
			Name:         "Retention Limitation",
			Code:         "PDPA-RetentionLimitation",
			Description:  "Limitation of data retention",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"PDPA-Section25"},
		},
	}
}

// getAPPICategories returns the APPI categories
func getAPPICategories() []RegionalCategory {
	return []RegionalCategory{
		{
			ID:           "APPI-PurposeSpecification",
			Name:         "Purpose Specification",
			Code:         "APPI-PurposeSpecification",
			Description:  "Specification of purpose for personal information handling",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"APPI-Article15", "APPI-Article16", "APPI-Article17"},
		},
		{
			ID:           "APPI-UseLimitation",
			Name:         "Use Limitation",
			Code:         "APPI-UseLimitation",
			Description:  "Limitation of use of personal information",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"APPI-Article16", "APPI-Article17", "APPI-Article18"},
		},
		{
			ID:           "APPI-AcquisitionLimitation",
			Name:         "Acquisition Limitation",
			Code:         "APPI-AcquisitionLimitation",
			Description:  "Limitation of acquisition of personal information",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"APPI-Article17", "APPI-Article18"},
		},
		{
			ID:           "APPI-SecurityControl",
			Name:         "Security Control",
			Code:         "APPI-SecurityControl",
			Description:  "Security control measures for personal information",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"APPI-Article20", "APPI-Article21", "APPI-Article22"},
		},
		{
			ID:           "APPI-Supervision",
			Name:         "Supervision",
			Code:         "APPI-Supervision",
			Description:  "Supervision of personal information handling",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"APPI-Article23", "APPI-Article24"},
		},
		{
			ID:           "APPI-IndividualRights",
			Name:         "Individual Rights",
			Code:         "APPI-IndividualRights",
			Description:  "Rights of individuals regarding their personal information",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"APPI-Article28", "APPI-Article29", "APPI-Article30"},
		},
		{
			ID:           "APPI-Remedies",
			Name:         "Remedies",
			Code:         "APPI-Remedies",
			Description:  "Remedies for violations of personal information protection",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"APPI-Article31", "APPI-Article32", "APPI-Article33"},
		},
	}
}

// getCCPARequirements returns the CCPA requirements
func getCCPARequirements() []RegionalRequirement {
	return []RegionalRequirement{
		{
			ID:                   "CCPA-1798.100",
			RequirementID:        "CCPA-1798.100",
			Framework:            FrameworkCCPA,
			Category:             CCPACategoryConsumerRights,
			Section:              "1798.100",
			Title:                "General Duties of Businesses that Collect Personal Information",
			Description:          "Businesses must inform consumers about the categories of personal information collected and the purposes for which it is used",
			DetailedDescription:  "Businesses that collect personal information must inform consumers at or before the point of collection about the categories of personal information to be collected and the purposes for which the categories of personal information shall be used",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Privacy notices, collection point disclosures, purpose documentation",
			KeyControls:          []string{"Privacy Notices", "Collection Point Disclosures", "Purpose Documentation", "Consumer Notification"},
			EffectiveDate:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"consumer-rights", "transparency", "collection-notice"},
		},
		{
			ID:                   "CCPA-1798.105",
			RequirementID:        "CCPA-1798.105",
			Framework:            FrameworkCCPA,
			Category:             CCPACategoryConsumerRights,
			Section:              "1798.105",
			Title:                "Right to Deletion",
			Description:          "Consumers have the right to request deletion of their personal information",
			DetailedDescription:  "A consumer shall have the right to request that a business delete any personal information about the consumer which the business has collected from the consumer",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Deletion request procedures, verification processes, deletion confirmation",
			KeyControls:          []string{"Deletion Request Procedures", "Verification Processes", "Deletion Confirmation", "Third-party Notification"},
			EffectiveDate:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"consumer-rights", "deletion", "right-to-be-forgotten"},
		},
		{
			ID:                   "CCPA-1798.110",
			RequirementID:        "CCPA-1798.110",
			Framework:            FrameworkCCPA,
			Category:             CCPACategoryConsumerRights,
			Section:              "1798.110",
			Title:                "Right to Know",
			Description:          "Consumers have the right to know what personal information is collected and how it is used",
			DetailedDescription:  "A consumer shall have the right to request that a business that collects personal information about the consumer disclose to the consumer the categories and specific pieces of personal information the business has collected",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Access request procedures, data inventory, response templates",
			KeyControls:          []string{"Access Request Procedures", "Data Inventory", "Response Templates", "Verification"},
			EffectiveDate:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"consumer-rights", "access", "transparency"},
		},
		{
			ID:                   "CCPA-1798.120",
			RequirementID:        "CCPA-1798.120",
			Framework:            FrameworkCCPA,
			Category:             CCPACategoryConsumerRights,
			Section:              "1798.120",
			Title:                "Right to Opt-Out",
			Description:          "Consumers have the right to opt-out of the sale or sharing of personal information",
			DetailedDescription:  "A consumer shall have the right, at any time, to direct a business that sells or shares personal information about the consumer to third parties not to sell or share the consumer's personal information",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Opt-out mechanisms, preference management, third-party compliance",
			KeyControls:          []string{"Opt-out Mechanisms", "Preference Management", "Third-party Compliance", "Do Not Sell"},
			EffectiveDate:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"consumer-rights", "opt-out", "do-not-sell"},
		},
	}
}

// getLGPDRequirements returns the LGPD requirements
func getLGPDRequirements() []RegionalRequirement {
	return []RegionalRequirement{
		{
			ID:                   "LGPD-Art.7",
			RequirementID:        "LGPD-Art.7",
			Framework:            FrameworkLGPD,
			Category:             LGPDCategoryLegalBasis,
			Section:              "Article 7",
			Title:                "Legal Basis for Processing",
			Description:          "Legal basis for processing personal data under LGPD",
			DetailedDescription:  "Personal data may only be processed on the basis of legal grounds or for legitimate purposes, including consent, compliance with legal obligations, and legitimate interests",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Legal basis documentation, consent records, legitimate interest assessments",
			KeyControls:          []string{"Legal Basis Documentation", "Consent Management", "Legitimate Interest Assessment", "Compliance Records"},
			EffectiveDate:        time.Date(2021, 9, 18, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"legal-basis", "consent", "legitimate-interests"},
		},
		{
			ID:                   "LGPD-Art.18",
			RequirementID:        "LGPD-Art.18",
			Framework:            FrameworkLGPD,
			Category:             LGPDCategoryDataSubjectRights,
			Section:              "Article 18",
			Title:                "Data Subject Rights",
			Description:          "Data subject rights under LGPD",
			DetailedDescription:  "Data subjects have the right to confirmation of the existence of processing, access to data, correction of incomplete or inaccurate data, and other rights",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Rights request procedures, response templates, verification processes",
			KeyControls:          []string{"Rights Request Procedures", "Response Templates", "Verification Processes", "Data Access"},
			EffectiveDate:        time.Date(2021, 9, 18, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"data-subject-rights", "access", "correction"},
		},
		{
			ID:                   "LGPD-Art.46",
			RequirementID:        "LGPD-Art.46",
			Framework:            FrameworkLGPD,
			Category:             LGPDCategoryDataProtection,
			Section:              "Article 46",
			Title:                "Security Measures",
			Description:          "Security measures for personal data processing",
			DetailedDescription:  "Processing agents shall adopt security, technical and administrative measures able to protect personal data from unauthorized access and accidental or unlawful situations",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Security measures documentation, risk assessments, incident response procedures",
			KeyControls:          []string{"Security Measures", "Risk Assessments", "Incident Response", "Access Controls"},
			EffectiveDate:        time.Date(2021, 9, 18, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"security", "data-protection", "technical-measures"},
		},
	}
}

// getPIPEDARequirements returns the PIPEDA requirements
func getPIPEDARequirements() []RegionalRequirement {
	return []RegionalRequirement{
		{
			ID:                   "PIPEDA-Principle1",
			RequirementID:        "PIPEDA-Principle1",
			Framework:            FrameworkPIPEDA,
			Category:             PIPEDACategoryConsent,
			Section:              "Principle 1",
			Title:                "Accountability",
			Description:          "Organizations are accountable for personal information under their control",
			DetailedDescription:  "An organization is responsible for personal information in its possession or custody, including information that has been transferred to a third party for processing",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Accountability policies, responsibility assignment, third-party oversight",
			KeyControls:          []string{"Accountability Policies", "Responsibility Assignment", "Third-party Oversight", "Compliance Monitoring"},
			EffectiveDate:        time.Date(2015, 6, 18, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"accountability", "responsibility", "oversight"},
		},
		{
			ID:                   "PIPEDA-Principle2",
			RequirementID:        "PIPEDA-Principle2",
			Framework:            FrameworkPIPEDA,
			Category:             PIPEDACategoryConsent,
			Section:              "Principle 2",
			Title:                "Identifying Purposes",
			Description:          "Organizations must identify the purposes for which personal information is collected",
			DetailedDescription:  "The purposes for which personal information is collected shall be identified by the organization at or before the time the information is collected",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Purpose documentation, collection notices, purpose identification",
			KeyControls:          []string{"Purpose Documentation", "Collection Notices", "Purpose Identification", "Consent Management"},
			EffectiveDate:        time.Date(2015, 6, 18, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"purpose-limitation", "consent", "transparency"},
		},
		{
			ID:                   "PIPEDA-Principle7",
			RequirementID:        "PIPEDA-Principle7",
			Framework:            FrameworkPIPEDA,
			Category:             PIPEDACategorySafeguards,
			Section:              "Principle 7",
			Title:                "Safeguards",
			Description:          "Personal information shall be protected by security safeguards",
			DetailedDescription:  "Personal information shall be protected by security safeguards appropriate to the sensitivity of the information",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Security safeguards documentation, risk assessments, protection measures",
			KeyControls:          []string{"Security Safeguards", "Risk Assessments", "Protection Measures", "Access Controls"},
			EffectiveDate:        time.Date(2015, 6, 18, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"security", "safeguards", "protection"},
		},
	}
}

// getPOPIARequirements returns the POPIA requirements
func getPOPIARequirements() []RegionalRequirement {
	return []RegionalRequirement{
		{
			ID:                   "POPIA-Section9",
			RequirementID:        "POPIA-Section9",
			Framework:            FrameworkPOPIA,
			Category:             "POPIA-ProcessingLimitation",
			Section:              "Section 9",
			Title:                "Processing Limitation",
			Description:          "Personal information may only be processed if it is adequate, relevant and not excessive",
			DetailedDescription:  "Personal information may only be processed if, given the purpose for which it is processed, it is adequate, relevant and not excessive",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Processing limitation policies, adequacy assessments, relevance documentation",
			KeyControls:          []string{"Processing Limitation Policies", "Adequacy Assessments", "Relevance Documentation", "Excessiveness Reviews"},
			EffectiveDate:        time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"processing-limitation", "adequacy", "relevance"},
		},
		{
			ID:                   "POPIA-Section13",
			RequirementID:        "POPIA-Section13",
			Framework:            FrameworkPOPIA,
			Category:             "POPIA-PurposeSpecification",
			Section:              "Section 13",
			Title:                "Purpose Specification",
			Description:          "Personal information must be collected for a specific, explicitly defined and lawful purpose",
			DetailedDescription:  "Personal information must be collected for a specific, explicitly defined and lawful purpose related to a function or activity of the responsible party",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Purpose specification documentation, lawful basis assessment, function mapping",
			KeyControls:          []string{"Purpose Specification", "Lawful Basis Assessment", "Function Mapping", "Purpose Documentation"},
			EffectiveDate:        time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"purpose-specification", "lawful-basis", "function-activity"},
		},
		{
			ID:                   "POPIA-Section19",
			RequirementID:        "POPIA-Section19",
			Framework:            FrameworkPOPIA,
			Category:             "POPIA-SecuritySafeguards",
			Section:              "Section 19",
			Title:                "Security Safeguards",
			Description:          "A responsible party must secure the integrity and confidentiality of personal information",
			DetailedDescription:  "A responsible party must secure the integrity and confidentiality of personal information in its possession or under its control by taking appropriate, reasonable technical and organisational measures",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Security measures documentation, technical controls, organizational measures",
			KeyControls:          []string{"Security Measures", "Technical Controls", "Organizational Measures", "Integrity Protection"},
			EffectiveDate:        time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"security", "integrity", "confidentiality"},
		},
	}
}

// getPDPARequirements returns the PDPA requirements
func getPDPARequirements() []RegionalRequirement {
	return []RegionalRequirement{
		{
			ID:                   "PDPA-Section13",
			RequirementID:        "PDPA-Section13",
			Framework:            FrameworkPDPA,
			Category:             "PDPA-Consent",
			Section:              "Section 13",
			Title:                "Consent Required",
			Description:          "An organization shall not collect, use or disclose personal data about an individual unless the individual gives, or is deemed to have given, consent",
			DetailedDescription:  "An organization shall not, on or after the appointed day, collect, use or disclose personal data about an individual unless the individual gives, or is deemed to have given, consent",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Consent management procedures, consent records, deemed consent documentation",
			KeyControls:          []string{"Consent Management", "Consent Records", "Deemed Consent Documentation", "Consent Validation"},
			EffectiveDate:        time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"consent", "collection", "use", "disclosure"},
		},
		{
			ID:                   "PDPA-Section24",
			RequirementID:        "PDPA-Section24",
			Framework:            FrameworkPDPA,
			Category:             "PDPA-Protection",
			Section:              "Section 24",
			Title:                "Protection of Personal Data",
			Description:          "An organization shall protect personal data in its possession or under its control by making reasonable security arrangements",
			DetailedDescription:  "An organization shall protect personal data in its possession or under its control by making reasonable security arrangements to prevent unauthorized access, collection, use, disclosure, copying, modification, disposal or similar risks",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Security arrangements documentation, access controls, risk assessments",
			KeyControls:          []string{"Security Arrangements", "Access Controls", "Risk Assessments", "Protection Measures"},
			EffectiveDate:        time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"protection", "security", "unauthorized-access"},
		},
		{
			ID:                   "PDPA-Section26D",
			RequirementID:        "PDPA-Section26D",
			Framework:            FrameworkPDPA,
			Category:             "PDPA-Notification",
			Section:              "Section 26D",
			Title:                "Data Breach Notification",
			Description:          "An organization shall, as soon as practicable, assess whether it is notifiable data breach",
			DetailedDescription:  "An organization shall, as soon as practicable, assess whether it is notifiable data breach and notify the Commission and affected individuals",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Breach assessment procedures, notification templates, incident response plans",
			KeyControls:          []string{"Breach Assessment", "Notification Procedures", "Incident Response", "Commission Notification"},
			EffectiveDate:        time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"data-breach", "notification", "incident-response"},
		},
	}
}

// getAPPIRequirements returns the APPI requirements
func getAPPIRequirements() []RegionalRequirement {
	return []RegionalRequirement{
		{
			ID:                   "APPI-Article15",
			RequirementID:        "APPI-Article15",
			Framework:            FrameworkAPPI,
			Category:             "APPI-PurposeSpecification",
			Section:              "Article 15",
			Title:                "Purpose Specification",
			Description:          "A personal information handling business operator shall specify the purpose of utilization as much as possible",
			DetailedDescription:  "A personal information handling business operator shall specify the purpose of utilization of personal information as much as possible",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Purpose specification documentation, utilization scope definition, purpose clarity",
			KeyControls:          []string{"Purpose Specification", "Utilization Scope", "Purpose Clarity", "Documentation"},
			EffectiveDate:        time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"purpose-specification", "utilization", "clarity"},
		},
		{
			ID:                   "APPI-Article20",
			RequirementID:        "APPI-Article20",
			Framework:            FrameworkAPPI,
			Category:             "APPI-SecurityControl",
			Section:              "Article 20",
			Title:                "Security Control Measures",
			Description:          "A personal information handling business operator shall take necessary and proper measures for the prevention of leakage, loss, or damage",
			DetailedDescription:  "A personal information handling business operator shall take necessary and proper measures for the prevention of leakage, loss, or damage, and for other security control of personal data",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Security control measures documentation, leakage prevention, loss prevention, damage prevention",
			KeyControls:          []string{"Security Control Measures", "Leakage Prevention", "Loss Prevention", "Damage Prevention"},
			EffectiveDate:        time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"security-control", "leakage-prevention", "data-protection"},
		},
		{
			ID:                   "APPI-Article28",
			RequirementID:        "APPI-Article28",
			Framework:            FrameworkAPPI,
			Category:             "APPI-IndividualRights",
			Section:              "Article 28",
			Title:                "Individual Rights",
			Description:          "An individual may request disclosure of personal data held by a personal information handling business operator",
			DetailedDescription:  "An individual may request disclosure of personal data held by a personal information handling business operator",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Disclosure request procedures, response templates, verification processes",
			KeyControls:          []string{"Disclosure Request Procedures", "Response Templates", "Verification Processes", "Data Access"},
			EffectiveDate:        time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"individual-rights", "disclosure", "data-access"},
		},
	}
}

// ConvertRegionalToRegulatoryFramework converts RegionalFrameworkDefinition to RegulatoryFramework
func (regional *RegionalFrameworkDefinition) ConvertRegionalToRegulatoryFramework() *RegulatoryFramework {
	requirements := make([]ComplianceRequirement, len(regional.Requirements))
	for i, req := range regional.Requirements {
		requirements[i] = ComplianceRequirement{
			ID:                   req.ID,
			Framework:            regional.ID,
			Category:             req.Category,
			RequirementID:        req.RequirementID,
			Title:                req.Title,
			Description:          req.Description,
			DetailedDescription:  req.DetailedDescription,
			RiskLevel:            req.RiskLevel,
			Priority:             req.Priority,
			Status:               ComplianceStatusNotStarted,
			ImplementationStatus: req.ImplementationStatus,
			EvidenceRequired:     req.EvidenceRequired,
			EvidenceDescription:  req.EvidenceDescription,
			Controls:             []ComplianceControl{},
			SubRequirements:      []ComplianceRequirement{},
			ParentRequirementID:  req.ParentRequirementID,
			ApplicableBusinesses: []string{},
			GeographicScope:      regional.GeographicScope,
			IndustryScope:        regional.IndustryScope,
			EffectiveDate:        req.EffectiveDate,
			LastUpdated:          req.LastUpdated,
			NextReviewDate:       req.NextReviewDate,
			ReviewFrequency:      req.ReviewFrequency,
			ComplianceOfficer:    req.ComplianceOfficer,
			Tags:                 req.Tags,
			Metadata:             req.Metadata,
		}
	}

	return &RegulatoryFramework{
		ID:              regional.ID,
		Name:            regional.Name,
		Version:         regional.Version,
		Description:     regional.Description,
		Type:            regional.Type,
		Jurisdiction:    regional.Jurisdiction,
		GeographicScope: regional.GeographicScope,
		IndustryScope:   regional.IndustryScope,
		EffectiveDate:   regional.EffectiveDate,
		LastUpdated:     regional.LastUpdated,
		NextReviewDate:  regional.NextReviewDate,
		Requirements:    requirements,
		MappingRules:    regional.MappingRules,
		Metadata:        regional.Metadata,
	}
}
