package compliance

import (
	"time"
)

// PCIDSSFramework defines the PCI DSS compliance framework
const (
	FrameworkPCIDSS = "PCIDSS"

	// PCI DSS Versions
	PCIDSSVersion4 = "4.0"
	PCIDSSVersion3 = "3.2.1"

	// PCI DSS Requirements Categories
	PCIDSSBuildMaintain = "Build and Maintain a Secure Network and Systems"
	PCIDSSProtectData   = "Protect Account Data"
	PCIDSSMaintainVuln  = "Maintain a Vulnerability Management Program"
	PCIDSSAccessControl = "Implement Strong Access Control Measures"
	PCIDSSMonitorTest   = "Regularly Monitor and Test Networks"
	PCIDSSPolicyInfo    = "Maintain an Information Security Policy"
)

// PCIDSSRequirement represents a PCI DSS requirement
type PCIDSSRequirement struct {
	ID                   string                 `json:"id"`
	RequirementID        string                 `json:"requirement_id"` // e.g., 1.1, 2.1, 3.1
	Category             string                 `json:"category"`       // Build and Maintain, Protect Data, etc.
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	DetailedDescription  string                 `json:"detailed_description"`
	TestingProcedures    []string               `json:"testing_procedures"`
	Guidance             string                 `json:"guidance"`
	RiskLevel            ComplianceRiskLevel    `json:"risk_level"`
	Priority             CompliancePriority     `json:"priority"`
	ImplementationStatus ImplementationStatus   `json:"implementation_status"`
	EvidenceRequired     bool                   `json:"evidence_required"`
	EvidenceDescription  string                 `json:"evidence_description"`
	KeyControls          []string               `json:"key_controls"`
	SubRequirements      []PCIDSSRequirement    `json:"sub_requirements,omitempty"`
	ParentRequirementID  *string                `json:"parent_requirement_id,omitempty"`
	EffectiveDate        time.Time              `json:"effective_date"`
	LastUpdated          time.Time              `json:"last_updated"`
	NextReviewDate       time.Time              `json:"next_review_date"`
	ReviewFrequency      string                 `json:"review_frequency"`
	ComplianceOfficer    string                 `json:"compliance_officer"`
	Tags                 []string               `json:"tags"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// PCIDSSFrameworkDefinition represents the complete PCI DSS framework
type PCIDSSFrameworkDefinition struct {
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
	Requirements    []PCIDSSRequirement    `json:"requirements"`
	Categories      []PCIDSSCategory       `json:"categories"`
	MappingRules    []FrameworkMapping     `json:"mapping_rules"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PCIDSSCategory represents a PCI DSS requirement category
type PCIDSSCategory struct {
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

// PCIDSSComplianceStatus represents PCI DSS specific compliance status
type PCIDSSComplianceStatus struct {
	BusinessID          string                             `json:"business_id"`
	Framework           string                             `json:"framework"`
	Version             string                             `json:"version"`
	MerchantLevel       string                             `json:"merchant_level"` // Level 1, 2, 3, 4
	ServiceProvider     bool                               `json:"service_provider"`
	OverallStatus       ComplianceStatus                   `json:"overall_status"`
	ComplianceScore     float64                            `json:"compliance_score"`
	CategoryStatus      map[string]CategoryStatus          `json:"category_status"`
	RequirementsStatus  map[string]PCIDSSRequirementStatus `json:"requirements_status"`
	LastAssessment      time.Time                          `json:"last_assessment"`
	NextAssessment      time.Time                          `json:"next_assessment"`
	AssessmentFrequency string                             `json:"assessment_frequency"`
	ComplianceOfficer   string                             `json:"compliance_officer"`
	QSA                 string                             `json:"qsa,omitempty"`
	CertificationDate   *time.Time                         `json:"certification_date,omitempty"`
	CertificationExpiry *time.Time                         `json:"certification_expiry,omitempty"`
	CertificationBody   string                             `json:"certification_body,omitempty"`
	CertificationNumber string                             `json:"certification_number,omitempty"`
	Notes               string                             `json:"notes"`
	Metadata            map[string]interface{}             `json:"metadata,omitempty"`
}

// CategoryStatus represents status for a specific PCI DSS category
type CategoryStatus struct {
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

// PCIDSSRequirementStatus represents status for a specific PCI DSS requirement
type PCIDSSRequirementStatus struct {
	RequirementID        string               `json:"requirement_id"`
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

// NewPCIDSSFramework creates a new PCI DSS framework definition
func NewPCIDSSFramework() *PCIDSSFrameworkDefinition {
	return &PCIDSSFrameworkDefinition{
		ID:              FrameworkPCIDSS,
		Name:            "Payment Card Industry Data Security Standard",
		Version:         PCIDSSVersion4,
		Description:     "PCI DSS is a set of security standards designed to ensure that all companies that process, store, or transmit credit card information maintain a secure environment",
		Type:            FrameworkTypeSecurity,
		Jurisdiction:    "International",
		GeographicScope: []string{"Global"},
		IndustryScope:   []string{"Financial Services", "Retail", "E-commerce", "Technology"},
		EffectiveDate:   time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
		LastUpdated:     time.Now(),
		NextReviewDate:  time.Now().AddDate(1, 0, 0),
		Requirements:    getPCIDSSRequirements(),
		Categories:      getPCIDSSCategories(),
		Metadata:        make(map[string]interface{}),
	}
}

// getPCIDSSCategories returns the PCI DSS requirement categories
func getPCIDSSCategories() []PCIDSSCategory {
	return []PCIDSSCategory{
		{
			ID:           PCIDSSBuildMaintain,
			Name:         "Build and Maintain a Secure Network and Systems",
			Code:         "1-2",
			Description:  "Requirements 1-2: Network security and system hardening",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"1.1", "1.2", "1.3", "1.4", "1.5", "1.6", "2.1", "2.2", "2.3", "2.4", "2.5", "2.6", "2.7", "2.8"},
		},
		{
			ID:           PCIDSSProtectData,
			Name:         "Protect Account Data",
			Code:         "3-4",
			Description:  "Requirements 3-4: Data protection and encryption",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"3.1", "3.2", "3.3", "3.4", "3.5", "3.6", "3.7", "3.8", "4.1", "4.2", "4.3"},
		},
		{
			ID:           PCIDSSMaintainVuln,
			Name:         "Maintain a Vulnerability Management Program",
			Code:         "5-6",
			Description:  "Requirements 5-6: Vulnerability management and security",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"5.1", "5.2", "5.3", "5.4", "6.1", "6.2", "6.3", "6.4", "6.5", "6.6", "6.7", "6.8"},
		},
		{
			ID:           PCIDSSAccessControl,
			Name:         "Implement Strong Access Control Measures",
			Code:         "7-9",
			Description:  "Requirements 7-9: Access control and physical security",
			RiskLevel:    ComplianceRiskLevelHigh,
			Priority:     CompliancePriorityHigh,
			Requirements: []string{"7.1", "7.2", "8.1", "8.2", "8.3", "8.4", "8.5", "8.6", "8.7", "8.8", "9.1", "9.2", "9.3", "9.4", "9.5", "9.6", "9.7", "9.8", "9.9", "9.10"},
		},
		{
			ID:           PCIDSSMonitorTest,
			Name:         "Regularly Monitor and Test Networks",
			Code:         "10-11",
			Description:  "Requirements 10-11: Monitoring and testing",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"10.1", "10.2", "10.3", "10.4", "10.5", "10.6", "10.7", "10.8", "11.1", "11.2", "11.3", "11.4", "11.5", "11.6"},
		},
		{
			ID:           PCIDSSPolicyInfo,
			Name:         "Maintain an Information Security Policy",
			Code:         "12",
			Description:  "Requirement 12: Security policy and procedures",
			RiskLevel:    ComplianceRiskLevelMedium,
			Priority:     CompliancePriorityMedium,
			Requirements: []string{"12.1", "12.2", "12.3", "12.4", "12.5", "12.6", "12.7", "12.8", "12.9", "12.10", "12.11"},
		},
	}
}

// getPCIDSSRequirements returns the PCI DSS requirements
func getPCIDSSRequirements() []PCIDSSRequirement {
	return []PCIDSSRequirement{
		// Build and Maintain a Secure Network and Systems
		{
			ID:                   "1.1",
			RequirementID:        "1.1",
			Category:             PCIDSSBuildMaintain,
			Title:                "Install and maintain network security controls",
			Description:          "Install and maintain network security controls to protect cardholder data",
			DetailedDescription:  "Network security controls (NSCs) are implemented to protect the CDE from unauthorized access. NSCs are configured to restrict connections to untrusted networks and prevent unauthorized access to the CDE.",
			TestingProcedures:    []string{"Review network security controls configuration", "Verify firewall rules and access control lists", "Test network segmentation"},
			Guidance:             "Implement network security controls to protect the cardholder data environment from unauthorized access",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Network security controls configuration, firewall rules, access control lists, network segmentation documentation",
			KeyControls:          []string{"Firewalls", "Network Segmentation", "Access Control Lists", "Intrusion Detection/Prevention Systems"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"network-security", "firewalls", "segmentation"},
		},
		{
			ID:                   "1.2",
			RequirementID:        "1.2",
			Category:             PCIDSSBuildMaintain,
			Title:                "Configure network security controls",
			Description:          "Configure network security controls to protect cardholder data",
			DetailedDescription:  "Network security controls are configured to restrict connections to untrusted networks and prevent unauthorized access to the CDE. Default settings are changed and security parameters are configured.",
			TestingProcedures:    []string{"Review network security control configurations", "Verify default settings have been changed", "Test security parameter configurations"},
			Guidance:             "Configure network security controls with secure settings and change default configurations",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Network security control configurations, change management documentation, security parameter settings",
			KeyControls:          []string{"Configuration Management", "Security Parameters", "Default Settings", "Change Management"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"configuration", "security-parameters", "default-settings"},
		},
		{
			ID:                   "2.1",
			RequirementID:        "2.1",
			Category:             PCIDSSBuildMaintain,
			Title:                "Apply secure configurations to all system components",
			Description:          "Apply secure configurations to all system components to protect cardholder data",
			DetailedDescription:  "Secure configurations are applied to all system components to protect cardholder data. Default settings are changed and security parameters are configured according to vendor recommendations and industry best practices.",
			TestingProcedures:    []string{"Review system configurations", "Verify default settings have been changed", "Test security parameter configurations"},
			Guidance:             "Apply secure configurations to all system components and change default settings",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "System configuration documentation, change management records, security parameter settings",
			KeyControls:          []string{"System Hardening", "Configuration Management", "Security Baselines", "Change Management"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"system-hardening", "configuration", "security-baselines"},
		},

		// Protect Account Data
		{
			ID:                   "3.1",
			RequirementID:        "3.1",
			Category:             PCIDSSProtectData,
			Title:                "Define and implement data retention and disposal policies",
			Description:          "Define and implement data retention and disposal policies for cardholder data",
			DetailedDescription:  "Data retention and disposal policies are defined and implemented to ensure that cardholder data is retained only for the time period required by business, legal, and/or regulatory requirements, and is disposed of securely when no longer needed.",
			TestingProcedures:    []string{"Review data retention and disposal policies", "Verify policy implementation", "Test data disposal procedures"},
			Guidance:             "Define and implement policies for data retention and secure disposal",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Data retention and disposal policies, implementation documentation, disposal procedures",
			KeyControls:          []string{"Data Retention Policy", "Data Disposal Policy", "Secure Disposal Procedures", "Policy Implementation"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"data-retention", "data-disposal", "policies"},
		},
		{
			ID:                   "3.2",
			RequirementID:        "3.2",
			Category:             PCIDSSProtectData,
			Title:                "Protect stored cardholder data",
			Description:          "Protect stored cardholder data using strong cryptography",
			DetailedDescription:  "Stored cardholder data is protected using strong cryptography. Encryption keys are managed securely and cryptographic algorithms are industry-accepted and properly implemented.",
			TestingProcedures:    []string{"Review encryption implementation", "Verify key management procedures", "Test cryptographic algorithms"},
			Guidance:             "Protect stored cardholder data using strong cryptography and secure key management",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Encryption implementation documentation, key management procedures, cryptographic algorithm specifications",
			KeyControls:          []string{"Strong Cryptography", "Key Management", "Encryption Algorithms", "Data Protection"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"encryption", "key-management", "cryptography"},
		},

		// Maintain a Vulnerability Management Program
		{
			ID:                   "5.1",
			RequirementID:        "5.1",
			Category:             PCIDSSMaintainVuln,
			Title:                "Define and implement processes to protect against malware",
			Description:          "Define and implement processes to protect against malware",
			DetailedDescription:  "Processes are defined and implemented to protect against malware. Anti-malware solutions are deployed and maintained on all systems commonly affected by malware.",
			TestingProcedures:    []string{"Review anti-malware processes", "Verify anti-malware deployment", "Test malware protection"},
			Guidance:             "Define and implement processes to protect against malware on all systems",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Anti-malware processes, deployment documentation, protection testing results",
			KeyControls:          []string{"Anti-malware Software", "Malware Protection", "Process Documentation", "System Protection"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"malware-protection", "anti-malware", "system-protection"},
		},
		{
			ID:                   "6.1",
			RequirementID:        "6.1",
			Category:             PCIDSSMaintainVuln,
			Title:                "Define and implement a process to identify security vulnerabilities",
			Description:          "Define and implement a process to identify security vulnerabilities",
			DetailedDescription:  "A process is defined and implemented to identify security vulnerabilities using industry-recognized vulnerability sources. Vulnerabilities are identified and ranked according to risk.",
			TestingProcedures:    []string{"Review vulnerability identification process", "Verify vulnerability sources", "Test vulnerability ranking"},
			Guidance:             "Define and implement a process to identify and rank security vulnerabilities",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Vulnerability identification process, vulnerability sources, ranking methodology",
			KeyControls:          []string{"Vulnerability Management", "Vulnerability Sources", "Risk Ranking", "Process Documentation"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"vulnerability-management", "security-vulnerabilities", "risk-ranking"},
		},

		// Implement Strong Access Control Measures
		{
			ID:                   "7.1",
			RequirementID:        "7.1",
			Category:             PCIDSSAccessControl,
			Title:                "Define and implement roles and responsibilities for managing logical access",
			Description:          "Define and implement roles and responsibilities for managing logical access",
			DetailedDescription:  "Roles and responsibilities for managing logical access to system components and cardholder data are defined and implemented. Access is based on job function and least privilege principles.",
			TestingProcedures:    []string{"Review roles and responsibilities", "Verify access management procedures", "Test least privilege implementation"},
			Guidance:             "Define and implement roles and responsibilities for logical access management",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Roles and responsibilities documentation, access management procedures, least privilege implementation",
			KeyControls:          []string{"Role Definition", "Access Management", "Least Privilege", "Responsibility Assignment"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"access-control", "roles", "responsibilities", "least-privilege"},
		},
		{
			ID:                   "8.1",
			RequirementID:        "8.1",
			Category:             PCIDSSAccessControl,
			Title:                "Define and implement processes for user identification and authentication",
			Description:          "Define and implement processes for user identification and authentication",
			DetailedDescription:  "Processes are defined and implemented for user identification and authentication. Users are uniquely identified and authenticated before accessing system components and cardholder data.",
			TestingProcedures:    []string{"Review user identification processes", "Verify authentication procedures", "Test unique user identification"},
			Guidance:             "Define and implement processes for user identification and authentication",
			RiskLevel:            ComplianceRiskLevelHigh,
			Priority:             CompliancePriorityHigh,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "User identification processes, authentication procedures, unique user identification implementation",
			KeyControls:          []string{"User Identification", "Authentication", "Unique User IDs", "Access Control"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"user-identification", "authentication", "access-control"},
		},

		// Regularly Monitor and Test Networks
		{
			ID:                   "10.1",
			RequirementID:        "10.1",
			Category:             PCIDSSMonitorTest,
			Title:                "Implement audit logging",
			Description:          "Implement audit logging to link access to system components to each individual user",
			DetailedDescription:  "Audit logging is implemented to link access to system components to each individual user. Audit logs capture all access to system components and cardholder data.",
			TestingProcedures:    []string{"Review audit logging implementation", "Verify log capture procedures", "Test user access linking"},
			Guidance:             "Implement audit logging to link system access to individual users",
			RiskLevel:            ComplianceRiskLevelMedium,
			Priority:             CompliancePriorityMedium,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Audit logging implementation, log capture procedures, user access linking documentation",
			KeyControls:          []string{"Audit Logging", "Log Capture", "User Access Tracking", "System Monitoring"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"audit-logging", "system-monitoring", "user-tracking"},
		},
		{
			ID:                   "11.1",
			RequirementID:        "11.1",
			Category:             PCIDSSMonitorTest,
			Title:                "Implement processes to test for the presence of wireless access points",
			Description:          "Implement processes to test for the presence of wireless access points",
			DetailedDescription:  "Processes are implemented to test for the presence of wireless access points and detect unauthorized wireless access points. Wireless access points are identified and authorized.",
			TestingProcedures:    []string{"Review wireless testing processes", "Verify wireless access point detection", "Test unauthorized access point detection"},
			Guidance:             "Implement processes to test for and detect wireless access points",
			RiskLevel:            ComplianceRiskLevelMedium,
			Priority:             CompliancePriorityMedium,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Wireless testing processes, access point detection procedures, unauthorized access point detection",
			KeyControls:          []string{"Wireless Testing", "Access Point Detection", "Unauthorized Detection", "Network Monitoring"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"wireless-testing", "access-point-detection", "network-monitoring"},
		},

		// Maintain an Information Security Policy
		{
			ID:                   "12.1",
			RequirementID:        "12.1",
			Category:             PCIDSSPolicyInfo,
			Title:                "Establish, publish, maintain, and disseminate a security policy",
			Description:          "Establish, publish, maintain, and disseminate a security policy",
			DetailedDescription:  "A security policy is established, published, maintained, and disseminated to all personnel. The policy defines security expectations and responsibilities for all personnel.",
			TestingProcedures:    []string{"Review security policy", "Verify policy dissemination", "Test policy maintenance procedures"},
			Guidance:             "Establish, publish, maintain, and disseminate a comprehensive security policy",
			RiskLevel:            ComplianceRiskLevelMedium,
			Priority:             CompliancePriorityMedium,
			ImplementationStatus: ImplementationStatusNotImplemented,
			EvidenceRequired:     true,
			EvidenceDescription:  "Security policy document, policy dissemination procedures, policy maintenance documentation",
			KeyControls:          []string{"Security Policy", "Policy Dissemination", "Policy Maintenance", "Documentation"},
			EffectiveDate:        time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
			LastUpdated:          time.Now(),
			NextReviewDate:       time.Now().AddDate(0, 6, 0),
			ReviewFrequency:      "semi-annually",
			Tags:                 []string{"security-policy", "policy-management", "documentation"},
		},
	}
}

// ConvertPCIDSSToRegulatoryFramework converts PCIDSSFrameworkDefinition to RegulatoryFramework
func (pci *PCIDSSFrameworkDefinition) ConvertPCIDSSToRegulatoryFramework() *RegulatoryFramework {
	requirements := make([]ComplianceRequirement, len(pci.Requirements))
	for i, req := range pci.Requirements {
		requirements[i] = ComplianceRequirement{
			ID:                   req.ID,
			Framework:            pci.ID,
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
			GeographicScope:      pci.GeographicScope,
			IndustryScope:        pci.IndustryScope,
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
		ID:              pci.ID,
		Name:            pci.Name,
		Version:         pci.Version,
		Description:     pci.Description,
		Type:            pci.Type,
		Jurisdiction:    pci.Jurisdiction,
		GeographicScope: pci.GeographicScope,
		IndustryScope:   pci.IndustryScope,
		EffectiveDate:   pci.EffectiveDate,
		LastUpdated:     pci.LastUpdated,
		NextReviewDate:  pci.NextReviewDate,
		Requirements:    requirements,
		MappingRules:    pci.MappingRules,
		Metadata:        pci.Metadata,
	}
}
