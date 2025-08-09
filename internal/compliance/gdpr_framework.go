package compliance

import (
	"time"
)

// GDPRFramework defines the GDPR compliance framework
const (
	FrameworkGDPR = "GDPR"
	
	// GDPR Versions
	GDPRVersion2018 = "2018"
	GDPRVersion2016 = "2016"
	
	// GDPR Principles
	GDPRPrincipleLawfulness = "Lawfulness, Fairness and Transparency"
	GDPRPrinciplePurpose    = "Purpose Limitation"
	GDPRPrincipleMinimization = "Data Minimization"
	GDPRPrincipleAccuracy   = "Accuracy"
	GDPRPrincipleStorage    = "Storage Limitation"
	GDPRPrincipleIntegrity  = "Integrity and Confidentiality"
	GDPRPrincipleAccountability = "Accountability"
	
	// GDPR Data Subject Rights
	GDPRRightAccess = "Right of Access"
	GDPRRightRectification = "Right to Rectification"
	GDPRRightErasure = "Right to Erasure"
	GDPRRightPortability = "Right to Data Portability"
	GDPRRightObjection = "Right to Object"
	GDPRRightRestriction = "Right to Restriction of Processing"
	GDPRRightAutomated = "Right to Automated Decision Making"
	GDPRRightCompensation = "Right to Compensation"
	
	// GDPR Legal Bases
	GDPRLegalBasisConsent = "Consent"
	GDPRLegalBasisContract = "Contract"
	GDPRLegalBasisLegalObligation = "Legal Obligation"
	GDPRLegalBasisVitalInterests = "Vital Interests"
	GDPRLegalBasisPublicTask = "Public Task"
	GDPRLegalBasisLegitimateInterests = "Legitimate Interests"
)

// GDPRRequirement represents a GDPR requirement
type GDPRRequirement struct {
	ID                   string                 `json:"id"`
	RequirementID        string                 `json:"requirement_id"` // e.g., Art.5, Art.6, Art.12
	Article              string                 `json:"article"`        // GDPR Article number
	Principle            string                 `json:"principle"`      // GDPR Principle
	Category             string                 `json:"category"`       // Data Protection, Rights, etc.
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
	SubRequirements      []GDPRRequirement      `json:"sub_requirements,omitempty"`
	ParentRequirementID  *string                `json:"parent_requirement_id,omitempty"`
	EffectiveDate        time.Time              `json:"effective_date"`
	LastUpdated          time.Time              `json:"last_updated"`
	NextReviewDate       time.Time              `json:"next_review_date"`
	ReviewFrequency      string                 `json:"review_frequency"`
	ComplianceOfficer    string                 `json:"compliance_officer"`
	Tags                 []string               `json:"tags"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// GDPRFrameworkDefinition represents the complete GDPR framework
type GDPRFrameworkDefinition struct {
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
	Requirements    []GDPRRequirement      `json:"requirements"`
	Principles      []GDPRPrinciple        `json:"principles"`
	DataSubjectRights []GDPRDataSubjectRight `json:"data_subject_rights"`
	MappingRules    []FrameworkMapping     `json:"mapping_rules"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// GDPRPrinciple represents a GDPR principle
type GDPRPrinciple struct {
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

// GDPRDataSubjectRight represents a GDPR data subject right
type GDPRDataSubjectRight struct {
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

// GDPRComplianceStatus represents GDPR specific compliance status
type GDPRComplianceStatus struct {
	BusinessID          string                           `json:"business_id"`
	Framework           string                           `json:"framework"`
	Version             string                           `json:"version"`
	DataController      bool                             `json:"data_controller"`
	DataProcessor       bool                             `json:"data_processor"`
	DataProtectionOfficer string                         `json:"data_protection_officer"`
	OverallStatus       ComplianceStatus                 `json:"overall_status"`
	ComplianceScore     float64                          `json:"compliance_score"`
	PrincipleStatus     map[string]PrincipleStatus       `json:"principle_status"`
	RightsStatus        map[string]DataSubjectRightStatus `json:"rights_status"`
	RequirementsStatus  map[string]GDPRRequirementStatus `json:"requirements_status"`
	LastAssessment      time.Time                        `json:"last_assessment"`
	NextAssessment      time.Time                        `json:"next_assessment"`
	AssessmentFrequency string                           `json:"assessment_frequency"`
	ComplianceOfficer   string                           `json:"compliance_officer"`
	SupervisoryAuthority string                          `json:"supervisory_authority,omitempty"`
	CertificationDate   *time.Time                       `json:"certification_date,omitempty"`
	CertificationExpiry *time.Time                       `json:"certification_expiry,omitempty"`
	CertificationBody   string                           `json:"certification_body,omitempty"`
	CertificationNumber string                           `json:"certification_number,omitempty"`
	Notes               string                           `json:"notes"`
	Metadata            map[string]interface{}           `json:"metadata,omitempty"`
}

// PrincipleStatus represents status for a specific GDPR principle
type PrincipleStatus struct {
	PrincipleID        string           `json:"principle_id"`
	PrincipleName      string           `json:"principle_name"`
	Status             ComplianceStatus `json:"status"`
	Score              float64          `json:"score"`
	RequirementCount   int              `json:"requirement_count"`
	ImplementedCount   int              `json:"implemented_count"`
	VerifiedCount      int              `json:"verified_count"`
	NonCompliantCount  int              `json:"non_compliant_count"`
	ExemptCount        int              `json:"exempt_count"`
	LastReviewed       time.Time        `json:"last_reviewed"`
	NextReview         time.Time        `json:"next_review"`
	Reviewer           string           `json:"reviewer"`
	Notes              string           `json:"notes"`
}

// DataSubjectRightStatus represents status for a specific GDPR data subject right
type DataSubjectRightStatus struct {
	RightID            string           `json:"right_id"`
	RightName          string           `json:"right_name"`
	Status             ComplianceStatus `json:"status"`
	Score              float64          `json:"score"`
	RequirementCount   int              `json:"requirement_count"`
	ImplementedCount   int              `json:"implemented_count"`
	VerifiedCount      int              `json:"verified_count"`
	NonCompliantCount  int              `json:"non_compliant_count"`
	ExemptCount        int              `json:"exempt_count"`
	LastReviewed       time.Time        `json:"last_reviewed"`
	NextReview         time.Time        `json:"next_review"`
	Reviewer           string           `json:"reviewer"`
	Notes              string           `json:"notes"`
}

// GDPRRequirementStatus represents status for a specific GDPR requirement
type GDPRRequirementStatus struct {
	RequirementID        string               `json:"requirement_id"`
	ArticleID            string               `json:"article_id"`
	PrincipleID          string               `json:"principle_id"`
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

// NewGDPRFramework creates a new GDPR framework definition
func NewGDPRFramework() *GDPRFrameworkDefinition {
	return &GDPRFrameworkDefinition{
		ID:              FrameworkGDPR,
		Name:            "General Data Protection Regulation",
		Version:         GDPRVersion2018,
		Description:     "The GDPR is a regulation in EU law on data protection and privacy in the European Union and the European Economic Area",
		Type:            FrameworkTypePrivacy,
		Jurisdiction:    "European Union",
		GeographicScope: []string{"European Union", "European Economic Area"},
		IndustryScope:   []string{"All Industries"},
		EffectiveDate:   time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
		LastUpdated:     time.Now(),
		NextReviewDate:  time.Now().AddDate(1, 0, 0),
		Requirements:    getGDPRRequirements(),
		Principles:      getGDPRPrinciples(),
		DataSubjectRights: getGDPRDataSubjectRights(),
		Metadata:        make(map[string]interface{}),
	}
}

// getGDPRPrinciples returns the GDPR principles
func getGDPRPrinciples() []GDPRPrinciple {
	return []GDPRPrinciple{
		{
			ID:          GDPRPrincipleLawfulness,
			Name:        "Lawfulness, Fairness and Transparency",
			Code:        "Art.5(1)(a)",
			Description: "Personal data shall be processed lawfully, fairly and in a transparent manner",
			RiskLevel:   ComplianceRiskLevelHigh,
			Priority:    CompliancePriorityHigh,
			Requirements: []string{"Art.5.1.a", "Art.6", "Art.7", "Art.8", "Art.9"},
		},
		{
			ID:          GDPRPrinciplePurpose,
			Name:        "Purpose Limitation",
			Code:        "Art.5(1)(b)",
			Description: "Personal data shall be collected for specified, explicit and legitimate purposes",
			RiskLevel:   ComplianceRiskLevelHigh,
			Priority:    CompliancePriorityHigh,
			Requirements: []string{"Art.5.1.b", "Art.6.4"},
		},
		{
			ID:          GDPRPrincipleMinimization,
			Name:        "Data Minimization",
			Code:        "Art.5(1)(c)",
			Description: "Personal data shall be adequate, relevant and limited to what is necessary",
			RiskLevel:   ComplianceRiskLevelHigh,
			Priority:    CompliancePriorityHigh,
			Requirements: []string{"Art.5.1.c", "Art.25"},
		},
		{
			ID:          GDPRPrincipleAccuracy,
			Name:        "Accuracy",
			Code:        "Art.5(1)(d)",
			Description: "Personal data shall be accurate and, where necessary, kept up to date",
			RiskLevel:   ComplianceRiskLevelMedium,
			Priority:    CompliancePriorityMedium,
			Requirements: []string{"Art.5.1.d", "Art.16"},
		},
		{
			ID:          GDPRPrincipleStorage,
			Name:        "Storage Limitation",
			Code:        "Art.5(1)(e)",
			Description: "Personal data shall be kept in a form which permits identification for no longer than necessary",
			RiskLevel:   ComplianceRiskLevelMedium,
			Priority:    CompliancePriorityMedium,
			Requirements: []string{"Art.5.1.e", "Art.17"},
		},
		{
			ID:          GDPRPrincipleIntegrity,
			Name:        "Integrity and Confidentiality",
			Code:        "Art.5(1)(f)",
			Description: "Personal data shall be processed in a manner that ensures appropriate security",
			RiskLevel:   ComplianceRiskLevelHigh,
			Priority:    CompliancePriorityHigh,
			Requirements: []string{"Art.5.1.f", "Art.32"},
		},
		{
			ID:          GDPRPrincipleAccountability,
			Name:        "Accountability",
			Code:        "Art.5(2)",
			Description: "The controller shall be responsible for, and be able to demonstrate compliance",
			RiskLevel:   ComplianceRiskLevelHigh,
			Priority:    CompliancePriorityHigh,
			Requirements: []string{"Art.5.2", "Art.24", "Art.25", "Art.30", "Art.33", "Art.34"},
		},
	}
}

// getGDPRDataSubjectRights returns the GDPR data subject rights
func getGDPRDataSubjectRights() []GDPRDataSubjectRight {
	return []GDPRDataSubjectRight{
		{
			ID:          GDPRRightAccess,
			Name:        "Right of Access",
			Code:        "Art.15",
			Description: "Data subjects have the right to obtain confirmation and access to their personal data",
			RiskLevel:   ComplianceRiskLevelHigh,
			Priority:    CompliancePriorityHigh,
			Requirements: []string{"Art.15", "Art.12"},
		},
		{
			ID:          GDPRRightRectification,
			Name:        "Right to Rectification",
			Code:        "Art.16",
			Description: "Data subjects have the right to have inaccurate personal data rectified",
			RiskLevel:   ComplianceRiskLevelMedium,
			Priority:    CompliancePriorityMedium,
			Requirements: []string{"Art.16", "Art.12"},
		},
		{
			ID:          GDPRRightErasure,
			Name:        "Right to Erasure",
			Code:        "Art.17",
			Description: "Data subjects have the right to have their personal data erased",
			RiskLevel:   ComplianceRiskLevelHigh,
			Priority:    CompliancePriorityHigh,
			Requirements: []string{"Art.17", "Art.12"},
		},
		{
			ID:          GDPRRightPortability,
			Name:        "Right to Data Portability",
			Code:        "Art.20",
			Description: "Data subjects have the right to receive their personal data in a structured format",
			RiskLevel:   ComplianceRiskLevelMedium,
			Priority:    CompliancePriorityMedium,
			Requirements: []string{"Art.20", "Art.12"},
		},
		{
			ID:          GDPRRightObjection,
			Name:        "Right to Object",
			Code:        "Art.21",
			Description: "Data subjects have the right to object to processing of their personal data",
			RiskLevel:   ComplianceRiskLevelMedium,
			Priority:    CompliancePriorityMedium,
			Requirements: []string{"Art.21", "Art.12"},
		},
		{
			ID:          GDPRRightRestriction,
			Name:        "Right to Restriction of Processing",
			Code:        "Art.18",
			Description: "Data subjects have the right to restrict processing of their personal data",
			RiskLevel:   ComplianceRiskLevelMedium,
			Priority:    CompliancePriorityMedium,
			Requirements: []string{"Art.18", "Art.12"},
		},
		{
			ID:          GDPRRightAutomated,
			Name:        "Right to Automated Decision Making",
			Code:        "Art.22",
			Description: "Data subjects have the right not to be subject to automated decision-making",
			RiskLevel:   ComplianceRiskLevelHigh,
			Priority:    CompliancePriorityHigh,
			Requirements: []string{"Art.22", "Art.12"},
		},
		{
			ID:          GDPRRightCompensation,
			Name:        "Right to Compensation",
			Code:        "Art.82",
			Description: "Data subjects have the right to compensation for material or non-material damage",
			RiskLevel:   ComplianceRiskLevelLow,
			Priority:    CompliancePriorityLow,
			Requirements: []string{"Art.82"},
		},
	}
}

// getGDPRRequirements returns the GDPR requirements
func getGDPRRequirements() []GDPRRequirement {
	return []GDPRRequirement{
		// Lawfulness, Fairness and Transparency
		{
			ID:                   "Art.5.1.a",
			RequirementID:        "Art.5.1.a",
			Article:              "Article 5(1)(a)",
			Principle:            GDPRPrincipleLawfulness,
			Category:             "Data Protection Principles",
			Title:                "Lawfulness, Fairness and Transparency",
			Description:          "Personal data shall be processed lawfully, fairly and in a transparent manner",
			DetailedDescription:  "Processing of personal data must have a legal basis, be fair to the data subject, and be transparent about how data is processed",
			LegalBasis:           []string{GDPRLegalBasisConsent, GDPRLegalBasisContract, GDPRLegalBasisLegalObligation, GDPRLegalBasisVitalInterests, GDPRLegalBasisPublicTask, GDPRLegalBasisLegitimateInterests},
			DataSubjectRights:    []string{GDPRRightAccess, GDPRRightObjection},
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Legal basis documentation, privacy notices, transparency measures",
			KeyControls:          []string{"Legal Basis Assessment", "Privacy Notices", "Transparency Measures", "Fair Processing"},
			EffectiveDate:        time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"lawfulness", "fairness", "transparency", "legal-basis"},
		},
		{
			ID:                   "Art.6",
			RequirementID:        "Art.6",
			Article:              "Article 6",
			Principle:            GDPRPrincipleLawfulness,
			Category:             "Legal Basis",
			Title:                "Lawfulness of Processing",
			Description:          "Processing shall be lawful only if and to the extent that at least one legal basis applies",
			DetailedDescription:  "Personal data processing must have a valid legal basis under Article 6(1) of the GDPR",
			LegalBasis:           []string{GDPRLegalBasisConsent, GDPRLegalBasisContract, GDPRLegalBasisLegalObligation, GDPRLegalBasisVitalInterests, GDPRLegalBasisPublicTask, GDPRLegalBasisLegitimateInterests},
			DataSubjectRights:    []string{GDPRRightAccess, GDPRRightObjection},
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Legal basis documentation, consent records, legitimate interest assessments",
			KeyControls:          []string{"Legal Basis Documentation", "Consent Management", "Legitimate Interest Assessment", "Contract Processing"},
			EffectiveDate:        time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"legal-basis", "lawfulness", "consent", "legitimate-interests"},
		},
		
		// Purpose Limitation
		{
			ID:                   "Art.5.1.b",
			RequirementID:        "Art.5.1.b",
			Article:              "Article 5(1)(b)",
			Principle:            GDPRPrinciplePurpose,
			Category:             "Data Protection Principles",
			Title:                "Purpose Limitation",
			Description:          "Personal data shall be collected for specified, explicit and legitimate purposes",
			DetailedDescription:  "Personal data must be collected for specific, clearly defined purposes and not processed in a manner incompatible with those purposes",
			LegalBasis:           []string{GDPRLegalBasisConsent, GDPRLegalBasisContract, GDPRLegalBasisLegitimateInterests},
			DataSubjectRights:    []string{GDPRRightAccess, GDPRRightObjection},
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Purpose documentation, privacy notices, processing records",
			KeyControls:          []string{"Purpose Documentation", "Privacy Notices", "Processing Records", "Purpose Limitation"},
			EffectiveDate:        time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"purpose-limitation", "data-collection", "processing-purposes"},
		},
		
		// Data Minimization
		{
			ID:                   "Art.5.1.c",
			RequirementID:        "Art.5.1.c",
			Article:              "Article 5(1)(c)",
			Principle:            GDPRPrincipleMinimization,
			Category:             "Data Protection Principles",
			Title:                "Data Minimization",
			Description:          "Personal data shall be adequate, relevant and limited to what is necessary",
			DetailedDescription:  "Only collect and process personal data that is necessary for the specified purposes",
			LegalBasis:           []string{GDPRLegalBasisConsent, GDPRLegalBasisContract, GDPRLegalBasisLegitimateInterests},
			DataSubjectRights:    []string{GDPRRightAccess, GDPRRightRectification},
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Data minimization assessments, processing records, data inventory",
			KeyControls:          []string{"Data Minimization Assessment", "Processing Records", "Data Inventory", "Collection Limits"},
			EffectiveDate:        time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"data-minimization", "adequacy", "relevance", "necessity"},
		},
		
		// Accuracy
		{
			ID:                   "Art.5.1.d",
			RequirementID:        "Art.5.1.d",
			Article:              "Article 5(1)(d)",
			Principle:            GDPRPrincipleAccuracy,
			Category:             "Data Protection Principles",
			Title:                "Accuracy",
			Description:          "Personal data shall be accurate and, where necessary, kept up to date",
			DetailedDescription:  "Personal data must be accurate and kept up to date, with reasonable steps taken to ensure accuracy",
			LegalBasis:           []string{GDPRLegalBasisConsent, GDPRLegalBasisContract, GDPRLegalBasisLegitimateInterests},
			DataSubjectRights:    []string{GDPRRightAccess, GDPRRightRectification},
			RiskLevel:            ComplianceRiskLevelMedium,
			Priority:             CompliancePriorityMedium,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Data accuracy procedures, update processes, rectification records",
			KeyControls:          []string{"Data Accuracy Procedures", "Update Processes", "Rectification Records", "Quality Controls"},
			EffectiveDate:        time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"accuracy", "data-quality", "updates", "rectification"},
		},
		
		// Storage Limitation
		{
			ID:                   "Art.5.1.e",
			RequirementID:        "Art.5.1.e",
			Article:              "Article 5(1)(e)",
			Principle:            GDPRPrincipleStorage,
			Category:             "Data Protection Principles",
			Title:                "Storage Limitation",
			Description:          "Personal data shall be kept in a form which permits identification for no longer than necessary",
			DetailedDescription:  "Personal data must not be kept longer than necessary for the specified purposes",
			LegalBasis:           []string{GDPRLegalBasisConsent, GDPRLegalBasisContract, GDPRLegalBasisLegalObligation},
			DataSubjectRights:    []string{GDPRRightAccess, GDPRRightErasure},
			RiskLevel:            ComplianceRiskLevelMedium,
			Priority:             CompliancePriorityMedium,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Retention policies, deletion procedures, storage records",
			KeyControls:          []string{"Retention Policies", "Deletion Procedures", "Storage Records", "Retention Limits"},
			EffectiveDate:        time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"storage-limitation", "retention", "deletion", "identification"},
		},
		
		// Integrity and Confidentiality
		{
			ID:                   "Art.5.1.f",
			RequirementID:        "Art.5.1.f",
			Article:              "Article 5(1)(f)",
			Principle:            GDPRPrincipleIntegrity,
			Category:             "Data Protection Principles",
			Title:                "Integrity and Confidentiality",
			Description:          "Personal data shall be processed in a manner that ensures appropriate security",
			DetailedDescription:  "Personal data must be processed with appropriate technical and organizational security measures",
			LegalBasis:           []string{GDPRLegalBasisConsent, GDPRLegalBasisContract, GDPRLegalBasisLegitimateInterests},
			DataSubjectRights:    []string{GDPRRightAccess, GDPRRightCompensation},
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Security measures documentation, risk assessments, incident response procedures",
			KeyControls:          []string{"Security Measures", "Risk Assessments", "Incident Response", "Access Controls"},
			EffectiveDate:        time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"integrity", "confidentiality", "security", "protection"},
		},
		
		// Accountability
		{
			ID:                   "Art.5.2",
			RequirementID:        "Art.5.2",
			Article:              "Article 5(2)",
			Principle:            GDPRPrincipleAccountability,
			Category:             "Data Protection Principles",
			Title:                "Accountability",
			Description:          "The controller shall be responsible for, and be able to demonstrate compliance",
			DetailedDescription:  "Data controllers must be able to demonstrate compliance with all GDPR principles",
			LegalBasis:           []string{GDPRLegalBasisConsent, GDPRLegalBasisContract, GDPRLegalBasisLegitimateInterests},
			DataSubjectRights:    []string{GDPRRightAccess, GDPRRightCompensation},
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Compliance documentation, audit trails, governance procedures",
			KeyControls:          []string{"Compliance Documentation", "Audit Trails", "Governance Procedures", "Demonstration"},
			EffectiveDate:        time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"accountability", "compliance", "demonstration", "governance"},
		},
		
		// Data Subject Rights
		{
			ID:                   "Art.15",
			RequirementID:        "Art.15",
			Article:              "Article 15",
			Principle:            GDPRPrincipleLawfulness,
			Category:             "Data Subject Rights",
			Title:                "Right of Access",
			Description:          "Data subjects have the right to obtain confirmation and access to their personal data",
			DetailedDescription:  "Data subjects must be able to access their personal data and obtain information about processing",
			LegalBasis:           []string{GDPRLegalBasisConsent, GDPRLegalBasisContract, GDPRLegalBasisLegitimateInterests},
			DataSubjectRights:    []string{GDPRRightAccess},
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Access request procedures, response templates, processing records",
			KeyControls:          []string{"Access Request Procedures", "Response Templates", "Processing Records", "Verification"},
			EffectiveDate:        time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"access-right", "data-subject-rights", "transparency"},
		},
		{
			ID:                   "Art.16",
			RequirementID:        "Art.16",
			Article:              "Article 16",
			Principle:            GDPRPrincipleAccuracy,
			Category:             "Data Subject Rights",
			Title:                "Right to Rectification",
			Description:          "Data subjects have the right to have inaccurate personal data rectified",
			DetailedDescription:  "Data subjects must be able to have inaccurate or incomplete personal data corrected",
			LegalBasis:           []string{GDPRLegalBasisConsent, GDPRLegalBasisContract, GDPRLegalBasisLegitimateInterests},
			DataSubjectRights:    []string{GDPRRightRectification},
			RiskLevel:            ComplianceRiskLevelMedium,
			Priority:             CompliancePriorityMedium,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Rectification procedures, verification processes, update records",
			KeyControls:          []string{"Rectification Procedures", "Verification Processes", "Update Records", "Accuracy Checks"},
			EffectiveDate:        time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"rectification", "accuracy", "correction", "data-subject-rights"},
		},
		{
			ID:                   "Art.17",
			RequirementID:        "Art.17",
			Article:              "Article 17",
			Principle:            GDPRPrincipleStorage,
			Category:             "Data Subject Rights",
			Title:                "Right to Erasure",
			Description:          "Data subjects have the right to have their personal data erased",
			DetailedDescription:  "Data subjects have the right to have their personal data deleted in certain circumstances",
			LegalBasis:           []string{GDPRLegalBasisConsent, GDPRLegalBasisContract, GDPRLegalBasisLegitimateInterests},
			DataSubjectRights:    []string{GDPRRightErasure},
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Erasure procedures, deletion verification, third-party notification",
			KeyControls:          []string{"Erasure Procedures", "Deletion Verification", "Third-party Notification", "Backup Deletion"},
			EffectiveDate:        time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"erasure", "deletion", "right-to-be-forgotten", "data-subject-rights"},
		},
	}
}

// ConvertGDPRToRegulatoryFramework converts GDPRFrameworkDefinition to RegulatoryFramework
func (gdpr *GDPRFrameworkDefinition) ConvertGDPRToRegulatoryFramework() *RegulatoryFramework {
	requirements := make([]ComplianceRequirement, len(gdpr.Requirements))
	for i, req := range gdpr.Requirements {
		requirements[i] = ComplianceRequirement{
			ID:                   req.ID,
			Framework:            gdpr.ID,
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
			GeographicScope:      gdpr.GeographicScope,
			IndustryScope:        gdpr.IndustryScope,
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
		ID:              gdpr.ID,
		Name:            gdpr.Name,
		Version:         gdpr.Version,
		Description:     gdpr.Description,
		Type:            gdpr.Type,
		Jurisdiction:    gdpr.Jurisdiction,
		GeographicScope: gdpr.GeographicScope,
		IndustryScope:   gdpr.IndustryScope,
		EffectiveDate:   gdpr.EffectiveDate,
		LastUpdated:     gdpr.LastUpdated,
		NextReviewDate:  gdpr.NextReviewDate,
		Requirements:    requirements,
		MappingRules:    gdpr.MappingRules,
		Metadata:        gdpr.Metadata,
	}
}
