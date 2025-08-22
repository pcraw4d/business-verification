package compliance

import (
	"time"
)

// SOC2Framework defines the SOC 2 compliance framework
const (
	FrameworkSOC2 = "SOC2"

	// SOC 2 Trust Service Criteria
	SOC2CriteriaCC = "CC" // Common Criteria
	SOC2CriteriaA  = "A"  // Availability
	SOC2CriteriaC  = "C"  // Confidentiality
	SOC2CriteriaP  = "P"  // Privacy
	SOC2CriteriaI  = "I"  // Integrity

	// SOC 2 Report Types
	SOC2ReportType1 = "Type 1" // Point in time
	SOC2ReportType2 = "Type 2" // Period of time
)

// SOC2Requirement represents a SOC 2 requirement
type SOC2Requirement struct {
	ID                   string                 `json:"id"`
	Criteria             string                 `json:"criteria"`       // CC, A, C, P, I
	Category             string                 `json:"category"`       // Control Environment, Communication, Risk Assessment, etc.
	Principle            string                 `json:"principle"`      // COSO Principle number
	RequirementID        string                 `json:"requirement_id"` // e.g., CC1.1, CC2.1
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	DetailedDescription  string                 `json:"detailed_description"`
	RiskLevel            ComplianceRiskLevel    `json:"risk_level"`
	Priority             CompliancePriority     `json:"priority"`
	ImplementationStatus ImplementationStatus   `json:"implementation_status"`
	EvidenceRequired     bool                   `json:"evidence_required"`
	EvidenceDescription  string                 `json:"evidence_description"`
	TestingProcedures    []string               `json:"testing_procedures"`
	KeyControls          []string               `json:"key_controls"`
	SubRequirements      []SOC2Requirement      `json:"sub_requirements,omitempty"`
	ParentRequirementID  *string                `json:"parent_requirement_id,omitempty"`
	EffectiveDate        time.Time              `json:"effective_date"`
	LastUpdated          time.Time              `json:"last_updated"`
	NextReviewDate       time.Time              `json:"next_review_date"`
	ReviewFrequency      string                 `json:"review_frequency"`
	ComplianceOfficer    string                 `json:"compliance_officer"`
	Tags                 []string               `json:"tags"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// SOC2FrameworkDefinition represents the complete SOC 2 framework
type SOC2FrameworkDefinition struct {
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
	Requirements    []SOC2Requirement      `json:"requirements"`
	Criteria        []SOC2Criteria         `json:"criteria"`
	MappingRules    []FrameworkMapping     `json:"mapping_rules"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// SOC2Criteria represents a SOC 2 Trust Service Criteria
type SOC2Criteria struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Code          string                 `json:"code"` // CC, A, C, P, I
	Description   string                 `json:"description"`
	Requirements  []string               `json:"requirements"` // Requirement IDs
	Categories    []string               `json:"categories"`
	Principles    []string               `json:"principles"`
	RiskLevel     ComplianceRiskLevel    `json:"risk_level"`
	Priority      CompliancePriority     `json:"priority"`
	EffectiveDate time.Time              `json:"effective_date"`
	LastUpdated   time.Time              `json:"last_updated"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SOC2ComplianceStatus represents SOC 2 specific compliance status
type SOC2ComplianceStatus struct {
	BusinessID          string                           `json:"business_id"`
	Framework           string                           `json:"framework"`
	ReportType          string                           `json:"report_type"` // Type 1 or Type 2
	AssessmentPeriod    string                           `json:"assessment_period"`
	OverallStatus       ComplianceStatus                 `json:"overall_status"`
	ComplianceScore     float64                          `json:"compliance_score"`
	CriteriaStatus      map[string]CriteriaStatus        `json:"criteria_status"`
	RequirementsStatus  map[string]SOC2RequirementStatus `json:"requirements_status"`
	LastAssessment      time.Time                        `json:"last_assessment"`
	NextAssessment      time.Time                        `json:"next_assessment"`
	AssessmentFrequency string                           `json:"assessment_frequency"`
	ComplianceOfficer   string                           `json:"compliance_officer"`
	Auditor             string                           `json:"auditor,omitempty"`
	CertificationDate   *time.Time                       `json:"certification_date,omitempty"`
	CertificationExpiry *time.Time                       `json:"certification_expiry,omitempty"`
	CertificationBody   string                           `json:"certification_body,omitempty"`
	CertificationNumber string                           `json:"certification_number,omitempty"`
	Notes               string                           `json:"notes"`
	Metadata            map[string]interface{}           `json:"metadata,omitempty"`
}

// CriteriaStatus represents status for a specific SOC 2 criteria
type CriteriaStatus struct {
	CriteriaID        string           `json:"criteria_id"`
	CriteriaName      string           `json:"criteria_name"`
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

// SOC2RequirementStatus represents status for a specific SOC 2 requirement
type SOC2RequirementStatus struct {
	RequirementID        string               `json:"requirement_id"`
	CriteriaID           string               `json:"criteria_id"`
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

// NewSOC2Framework creates a new SOC 2 framework definition
func NewSOC2Framework() *SOC2FrameworkDefinition {
	return &SOC2FrameworkDefinition{
		ID:              FrameworkSOC2,
		Name:            "SOC 2 Trust Services Criteria",
		Version:         "2017",
		Description:     "AICPA Trust Services Criteria for Security, Availability, Processing Integrity, Confidentiality, and Privacy",
		Type:            FrameworkTypeSecurity,
		Jurisdiction:    "United States",
		GeographicScope: []string{"United States", "Canada"},
		IndustryScope:   []string{"Technology", "Financial Services", "Healthcare", "Retail"},
		EffectiveDate:   time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
		LastUpdated:     time.Now(),
		NextReviewDate:  time.Now().AddDate(1, 0, 0),
		Requirements:    getSOC2Requirements(),
		Criteria:        getSOC2Criteria(),
		Metadata:        make(map[string]interface{}),
	}
}

// getSOC2Criteria returns the SOC 2 Trust Service Criteria
func getSOC2Criteria() []SOC2Criteria {
	return []SOC2Criteria{
		{
			ID:          SOC2CriteriaCC,
			Name:        "Common Criteria",
			Code:        SOC2CriteriaCC,
			Description: "Security, availability, and processing integrity criteria that apply to all services",
			RiskLevel:   ComplianceRiskLevelHigh,
			Priority:    CompliancePriorityHigh,
			Categories:  []string{"Control Environment", "Communication", "Risk Assessment", "Monitoring Activities", "Control Activities", "Information and Communication", "Logical and Physical Access Controls", "System Operations", "Change Management", "Risk Mitigation"},
			Principles:  []string{"CC1", "CC2", "CC3", "CC4", "CC5", "CC6", "CC7", "CC8", "CC9"},
		},
		{
			ID:          SOC2CriteriaA,
			Name:        "Availability",
			Code:        SOC2CriteriaA,
			Description: "The system is available for operation and use as committed or agreed",
			RiskLevel:   ComplianceRiskLevelMedium,
			Priority:    CompliancePriorityMedium,
			Categories:  []string{"Availability", "System Operations", "Change Management"},
			Principles:  []string{"A1"},
		},
		{
			ID:          SOC2CriteriaC,
			Name:        "Confidentiality",
			Code:        SOC2CriteriaC,
			Description: "Information designated as confidential is protected as committed or agreed",
			RiskLevel:   ComplianceRiskLevelHigh,
			Priority:    CompliancePriorityHigh,
			Categories:  []string{"Confidentiality", "Logical and Physical Access Controls", "System Operations"},
			Principles:  []string{"C1"},
		},
		{
			ID:          SOC2CriteriaP,
			Name:        "Privacy",
			Code:        SOC2CriteriaP,
			Description: "Personal information is collected, used, retained, disclosed, and disposed of in conformity with the commitments in the entity's privacy notice",
			RiskLevel:   ComplianceRiskLevelHigh,
			Priority:    CompliancePriorityHigh,
			Categories:  []string{"Privacy", "Notice and Communication", "Choice and Consent", "Collection", "Use, Retention, and Disposal", "Access", "Disclosure to Third Parties", "Quality", "Monitoring and Enforcement"},
			Principles:  []string{"P1", "P2", "P3", "P4", "P5", "P6", "P7", "P8"},
		},
		{
			ID:          SOC2CriteriaI,
			Name:        "Processing Integrity",
			Code:        SOC2CriteriaI,
			Description: "System processing is complete, accurate, timely, and authorized",
			RiskLevel:   ComplianceRiskLevelMedium,
			Priority:    CompliancePriorityMedium,
			Categories:  []string{"Processing Integrity", "System Operations", "Change Management"},
			Principles:  []string{"PI1"},
		},
	}
}

// getSOC2Requirements returns the SOC 2 requirements
func getSOC2Requirements() []SOC2Requirement {
	return []SOC2Requirement{
		// Common Criteria (CC) Requirements
		{
			ID:                   "CC1.1",
			Criteria:             SOC2CriteriaCC,
			Category:             "Control Environment",
			Principle:            "CC1",
			RequirementID:        "CC1.1",
			Title:                "Commitment to Integrity and Ethical Values",
			Description:          "The entity demonstrates a commitment to integrity and ethical values",
			DetailedDescription:  "The entity demonstrates a commitment to integrity and ethical values through the development and use of entity standards, policies, and procedures that support the entity's culture and guide the conduct of the entity's personnel in performing their assigned functions and in relationships with customers, suppliers, and other parties.",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Entity standards, policies, and procedures; training materials; communication records; disciplinary actions",
			TestingProcedures:    []string{"Review entity standards and policies", "Interview management and personnel", "Review training records", "Review disciplinary actions"},
			KeyControls:          []string{"Code of Conduct", "Ethics Policy", "Training Program", "Disciplinary Procedures"},
			EffectiveDate:        time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"control-environment", "ethics", "integrity"},
		},
		{
			ID:                   "CC2.1",
			Criteria:             SOC2CriteriaCC,
			Category:             "Communication and Information",
			Principle:            "CC2",
			RequirementID:        "CC2.1",
			Title:                "Information Quality",
			Description:          "The entity obtains or generates and uses relevant, quality information to support the functioning of internal control",
			DetailedDescription:  "The entity obtains or generates and uses relevant, quality information to support the functioning of internal control. Information is quality information if it is appropriate, current, complete, accurate, accessible, and provided on a timely basis.",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Information quality standards; data validation procedures; information system controls",
			TestingProcedures:    []string{"Review information quality standards", "Test data validation procedures", "Review information system controls"},
			KeyControls:          []string{"Data Quality Standards", "Validation Procedures", "Information System Controls"},
			EffectiveDate:        time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"information-quality", "data-validation", "system-controls"},
		},
		{
			ID:                   "CC3.1",
			Criteria:             SOC2CriteriaCC,
			Category:             "Risk Assessment",
			Principle:            "CC3",
			RequirementID:        "CC3.1",
			Title:                "Risk Identification",
			Description:          "The entity identifies risks to the achievement of its objectives across the entity and analyzes risks as a basis for determining how the risks should be managed",
			DetailedDescription:  "The entity identifies risks to the achievement of its objectives across the entity and analyzes risks as a basis for determining how the risks should be managed. Risk identification includes the identification of risks from external sources, including those related to technology, economic, political, and social conditions, and natural catastrophes.",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Risk assessment procedures; risk registers; risk analysis reports",
			TestingProcedures:    []string{"Review risk assessment procedures", "Review risk registers", "Review risk analysis reports"},
			KeyControls:          []string{"Risk Assessment Procedures", "Risk Register", "Risk Analysis Reports"},
			EffectiveDate:        time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"risk-assessment", "risk-identification", "risk-analysis"},
		},
		// Add more CC requirements as needed...

		// Availability (A) Requirements
		{
			ID:                   "A1.1",
			Criteria:             SOC2CriteriaA,
			Category:             "Availability",
			Principle:            "A1",
			RequirementID:        "A1.1",
			Title:                "Availability",
			Description:          "The entity maintains, monitors, and evaluates current processing capacity and projected processing needs",
			DetailedDescription:  "The entity maintains, monitors, and evaluates current processing capacity and projected processing needs, including the following: a) Current and projected processing capacity and availability, b) Current and projected processing needs, c) Current and projected processing capacity and availability compared to current and projected processing needs, d) Current and projected processing capacity and availability compared to current and projected processing needs, and e) Current and projected processing capacity and availability compared to current and projected processing needs.",
			RiskLevel:            ComplianceRiskLevelMedium,
			Priority:             CompliancePriorityMedium,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Capacity planning procedures; monitoring reports; evaluation results",
			TestingProcedures:    []string{"Review capacity planning procedures", "Review monitoring reports", "Review evaluation results"},
			KeyControls:          []string{"Capacity Planning", "Monitoring Procedures", "Evaluation Procedures"},
			EffectiveDate:        time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"availability", "capacity-planning", "monitoring"},
		},

		// Confidentiality (C) Requirements
		{
			ID:                   "C1.1",
			Criteria:             SOC2CriteriaC,
			Category:             "Confidentiality",
			Principle:            "C1",
			RequirementID:        "C1.1",
			Title:                "Confidentiality",
			Description:          "The entity identifies and maintains confidential information to meet the entity's objectives related to confidentiality",
			DetailedDescription:  "The entity identifies and maintains confidential information to meet the entity's objectives related to confidentiality. Confidential information is information that is designated as confidential and that meets the entity's criteria for classification as confidential information.",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Confidentiality classification procedures; access controls; encryption procedures",
			TestingProcedures:    []string{"Review confidentiality classification procedures", "Test access controls", "Review encryption procedures"},
			KeyControls:          []string{"Classification Procedures", "Access Controls", "Encryption Procedures"},
			EffectiveDate:        time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"confidentiality", "classification", "access-controls"},
		},

		// Privacy (P) Requirements
		{
			ID:                   "P1.1",
			Criteria:             SOC2CriteriaP,
			Category:             "Notice and Communication",
			Principle:            "P1",
			RequirementID:        "P1.1",
			Title:                "Notice and Communication of Objectives",
			Description:          "The entity provides notice to data subjects about its privacy practices",
			DetailedDescription:  "The entity provides notice to data subjects about its privacy practices. The notice is provided to data subjects at or before the time personal information is collected or as soon as practicable thereafter.",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Privacy notices; communication procedures; consent mechanisms",
			TestingProcedures:    []string{"Review privacy notices", "Review communication procedures", "Test consent mechanisms"},
			KeyControls:          []string{"Privacy Notices", "Communication Procedures", "Consent Mechanisms"},
			EffectiveDate:        time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"privacy", "notice", "communication"},
		},

		// Processing Integrity (PI) Requirements
		{
			ID:                   "PI1.1",
			Criteria:             SOC2CriteriaI,
			Category:             "Processing Integrity",
			Principle:            "PI1",
			RequirementID:        "PI1.1",
			Title:                "Processing Integrity",
			Description:          "The entity implements policies and procedures to provide reasonable assurance that system processing is complete, accurate, timely, and authorized",
			DetailedDescription:  "The entity implements policies and procedures to provide reasonable assurance that system processing is complete, accurate, timely, and authorized. System processing includes the following: a) Input, b) Processing, c) Output, and d) Storage.",
			RiskLevel:            ComplianceRiskLevelMedium,
			Priority:             CompliancePriorityMedium,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Processing integrity policies; validation procedures; error handling procedures",
			TestingProcedures:    []string{"Review processing integrity policies", "Test validation procedures", "Review error handling procedures"},
			KeyControls:          []string{"Processing Policies", "Validation Procedures", "Error Handling"},
			EffectiveDate:        time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"processing-integrity", "validation", "error-handling"},
		},
	}
}

// ConvertSOC2ToRegulatoryFramework converts SOC2FrameworkDefinition to RegulatoryFramework
func (soc2 *SOC2FrameworkDefinition) ConvertSOC2ToRegulatoryFramework() *RegulatoryFramework {
	requirements := make([]ComplianceRequirement, len(soc2.Requirements))
	for i, req := range soc2.Requirements {
		requirements[i] = ComplianceRequirement{
			ID:                   req.ID,
			Framework:            soc2.ID,
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
			GeographicScope:      soc2.GeographicScope,
			IndustryScope:        soc2.IndustryScope,
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
		ID:              soc2.ID,
		Name:            soc2.Name,
		Version:         soc2.Version,
		Description:     soc2.Description,
		Type:            soc2.Type,
		Jurisdiction:    soc2.Jurisdiction,
		GeographicScope: soc2.GeographicScope,
		IndustryScope:   soc2.IndustryScope,
		EffectiveDate:   soc2.EffectiveDate,
		LastUpdated:     soc2.LastUpdated,
		NextReviewDate:  soc2.NextReviewDate,
		Requirements:    requirements,
		MappingRules:    soc2.MappingRules,
		Metadata:        soc2.Metadata,
	}
}
