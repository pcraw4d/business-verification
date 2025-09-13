package compliance

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// ComplianceFrameworkService provides comprehensive compliance framework management
type ComplianceFrameworkService struct {
	logger       *observability.Logger
	frameworks   map[string]*ComplianceFramework
	requirements map[string]*ComplianceRequirement
	assessments  map[string]*ComplianceAssessment
}

// ComplianceFramework represents a compliance framework (SOC2, GDPR, PCI DSS, etc.)
type ComplianceFramework struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	Category      string                 `json:"category"` // "security", "privacy", "financial", "operational"
	Status        string                 `json:"status"`   // "active", "deprecated", "draft"
	EffectiveDate time.Time              `json:"effective_date"`
	ExpiryDate    *time.Time             `json:"expiry_date,omitempty"`
	Requirements  []string               `json:"requirements"`  // Requirement IDs
	Controls      []string               `json:"controls"`      // Control IDs
	Scope         []string               `json:"scope"`         // Applicable business types
	Jurisdiction  []string               `json:"jurisdiction"`  // Geographic applicability
	Authority     string                 `json:"authority"`     // Regulatory authority
	Documentation []string               `json:"documentation"` // Reference documents
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ComplianceRequirement represents a specific compliance requirement
type ComplianceRequirement struct {
	ID               string                 `json:"id"`
	FrameworkID      string                 `json:"framework_id"`
	Code             string                 `json:"code"` // e.g., "CC6.1", "GDPR_32"
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Category         string                 `json:"category"`          // "access_control", "data_protection", "monitoring"
	Priority         string                 `json:"priority"`          // "critical", "high", "medium", "low"
	Type             string                 `json:"type"`              // "technical", "administrative", "physical"
	Applicability    []string               `json:"applicability"`     // Business types this applies to
	EvidenceTypes    []string               `json:"evidence_types"`    // Required evidence types
	AssessmentMethod string                 `json:"assessment_method"` // "automated", "manual", "hybrid"
	Frequency        string                 `json:"frequency"`         // "continuous", "monthly", "quarterly", "annually"
	Owner            string                 `json:"owner"`             // Responsible team/role
	References       []string               `json:"references"`        // Reference documents
	RelatedControls  []string               `json:"related_controls"`  // Related control IDs
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// ComplianceAssessment represents a compliance assessment for a business
type ComplianceAssessment struct {
	ID              string                     `json:"id"`
	BusinessID      string                     `json:"business_id"`
	FrameworkID     string                     `json:"framework_id"`
	AssessmentType  string                     `json:"assessment_type"` // "initial", "periodic", "remediation", "audit"
	Status          string                     `json:"status"`          // "planned", "in_progress", "completed", "failed"
	StartDate       time.Time                  `json:"start_date"`
	EndDate         *time.Time                 `json:"end_date,omitempty"`
	Assessor        string                     `json:"assessor"`         // Person/team conducting assessment
	OverallScore    float64                    `json:"overall_score"`    // 0.0 to 1.0
	ComplianceLevel string                     `json:"compliance_level"` // "compliant", "partial", "non_compliant"
	Findings        []ComplianceFinding        `json:"findings"`
	Recommendations []ComplianceRecommendation `json:"recommendations"`
	Evidence        []ComplianceEvidence       `json:"evidence"`
	Metadata        map[string]interface{}     `json:"metadata,omitempty"`
	CreatedAt       time.Time                  `json:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at"`
}

// ComplianceFinding represents a specific compliance finding
type ComplianceFinding struct {
	ID            string                 `json:"id"`
	RequirementID string                 `json:"requirement_id"`
	Type          string                 `json:"type"`     // "gap", "violation", "observation", "strength"
	Severity      string                 `json:"severity"` // "critical", "high", "medium", "low"
	Status        string                 `json:"status"`   // "open", "in_progress", "resolved", "accepted_risk"
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Impact        string                 `json:"impact"`
	RootCause     string                 `json:"root_cause"`
	Remediation   string                 `json:"remediation"`
	DueDate       *time.Time             `json:"due_date,omitempty"`
	Owner         string                 `json:"owner"`
	Evidence      []string               `json:"evidence,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ComplianceRecommendation represents a compliance recommendation
type ComplianceRecommendation struct {
	ID             string                 `json:"id"`
	RequirementID  string                 `json:"requirement_id,omitempty"`
	FindingID      string                 `json:"finding_id,omitempty"`
	Priority       string                 `json:"priority"` // "critical", "high", "medium", "low"
	Category       string                 `json:"category"` // "technical", "process", "training", "documentation"
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Implementation string                 `json:"implementation"`
	Effort         string                 `json:"effort"`   // "low", "medium", "high"
	Timeline       string                 `json:"timeline"` // "immediate", "1_month", "3_months", "6_months"
	Owner          string                 `json:"owner"`
	Status         string                 `json:"status"` // "pending", "in_progress", "completed", "rejected"
	DueDate        *time.Time             `json:"due_date,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// ComplianceEvidence represents evidence for compliance
type ComplianceEvidence struct {
	ID               string                 `json:"id"`
	RequirementID    string                 `json:"requirement_id"`
	Type             string                 `json:"type"` // "document", "screenshot", "log", "test_result"
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Content          string                 `json:"content"` // Base64 encoded or reference
	Format           string                 `json:"format"`  // "pdf", "png", "txt", "json"
	Size             int64                  `json:"size"`
	Hash             string                 `json:"hash"` // Content hash for integrity
	ValidFrom        time.Time              `json:"valid_from"`
	ValidTo          *time.Time             `json:"valid_to,omitempty"`
	UploadedBy       string                 `json:"uploaded_by"`
	VerifiedBy       string                 `json:"verified_by,omitempty"`
	VerificationDate *time.Time             `json:"verification_date,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// FrameworkQuery represents query parameters for framework operations
type FrameworkQuery struct {
	Category        string   `json:"category,omitempty"`
	Status          string   `json:"status,omitempty"`
	Jurisdiction    []string `json:"jurisdiction,omitempty"`
	BusinessType    string   `json:"business_type,omitempty"`
	IncludeInactive bool     `json:"include_inactive,omitempty"`
	Limit           int      `json:"limit,omitempty"`
	Offset          int      `json:"offset,omitempty"`
}

// NewComplianceFrameworkService creates a new compliance framework service
func NewComplianceFrameworkService(logger *observability.Logger) *ComplianceFrameworkService {
	service := &ComplianceFrameworkService{
		logger:       logger,
		frameworks:   make(map[string]*ComplianceFramework),
		requirements: make(map[string]*ComplianceRequirement),
		assessments:  make(map[string]*ComplianceAssessment),
	}

	// Initialize with default frameworks
	service.initializeDefaultFrameworks()

	return service
}

// GetFrameworks retrieves compliance frameworks based on query criteria
func (cfs *ComplianceFrameworkService) GetFrameworks(ctx context.Context, query *FrameworkQuery) ([]*ComplianceFramework, error) {
	cfs.logger.Info("Retrieving compliance frameworks", map[string]interface{}{
		"query": query,
	})

	var frameworks []*ComplianceFramework

	for _, framework := range cfs.frameworks {
		// Apply filters
		if query.Category != "" && framework.Category != query.Category {
			continue
		}
		if query.Status != "" && framework.Status != query.Status {
			continue
		}
		if !query.IncludeInactive && framework.Status == "deprecated" {
			continue
		}
		if len(query.Jurisdiction) > 0 {
			jurisdictionMatch := false
			for _, j := range query.Jurisdiction {
				for _, f := range framework.Jurisdiction {
					if strings.EqualFold(j, f) {
						jurisdictionMatch = true
						break
					}
				}
				if jurisdictionMatch {
					break
				}
			}
			if !jurisdictionMatch {
				continue
			}
		}
		if query.BusinessType != "" {
			scopeMatch := false
			for _, s := range framework.Scope {
				if strings.EqualFold(s, query.BusinessType) || s == "all" {
					scopeMatch = true
					break
				}
			}
			if !scopeMatch {
				continue
			}
		}

		frameworks = append(frameworks, framework)
	}

	// Sort by name
	sort.Slice(frameworks, func(i, j int) bool {
		return frameworks[i].Name < frameworks[j].Name
	})

	// Apply pagination
	if query.Limit > 0 {
		start := query.Offset
		end := query.Offset + query.Limit
		if start >= len(frameworks) {
			frameworks = []*ComplianceFramework{}
		} else if end > len(frameworks) {
			frameworks = frameworks[start:]
		} else {
			frameworks = frameworks[start:end]
		}
	}

	cfs.logger.Info("Retrieved compliance frameworks", map[string]interface{}{
		"count": len(frameworks),
		"query": query,
	})

	return frameworks, nil
}

// GetFramework retrieves a specific compliance framework by ID
func (cfs *ComplianceFrameworkService) GetFramework(ctx context.Context, frameworkID string) (*ComplianceFramework, error) {
	cfs.logger.Info("Retrieving compliance framework", map[string]interface{}{
		"framework_id": frameworkID,
	})

	framework, exists := cfs.frameworks[frameworkID]
	if !exists {
		return nil, fmt.Errorf("framework not found: %s", frameworkID)
	}

	cfs.logger.Info("Retrieved compliance framework", map[string]interface{}{
		"framework_id": frameworkID,
		"name":         framework.Name,
		"category":     framework.Category,
	})

	return framework, nil
}

// GetFrameworkRequirements retrieves requirements for a specific framework
func (cfs *ComplianceFrameworkService) GetFrameworkRequirements(ctx context.Context, frameworkID string) ([]*ComplianceRequirement, error) {
	cfs.logger.Info("Retrieving framework requirements", map[string]interface{}{
		"framework_id": frameworkID,
	})

	framework, exists := cfs.frameworks[frameworkID]
	if !exists {
		return nil, fmt.Errorf("framework not found: %s", frameworkID)
	}

	var requirements []*ComplianceRequirement
	for _, reqID := range framework.Requirements {
		if req, exists := cfs.requirements[reqID]; exists {
			requirements = append(requirements, req)
		}
	}

	// Sort by code
	sort.Slice(requirements, func(i, j int) bool {
		return requirements[i].Code < requirements[j].Code
	})

	cfs.logger.Info("Retrieved framework requirements", map[string]interface{}{
		"framework_id": frameworkID,
		"count":        len(requirements),
	})

	return requirements, nil
}

// CreateAssessment creates a new compliance assessment
func (cfs *ComplianceFrameworkService) CreateAssessment(ctx context.Context, assessment *ComplianceAssessment) error {
	cfs.logger.Info("Creating compliance assessment", map[string]interface{}{
		"business_id":  assessment.BusinessID,
		"framework_id": assessment.FrameworkID,
		"type":         assessment.AssessmentType,
	})

	// Validate framework exists
	if _, exists := cfs.frameworks[assessment.FrameworkID]; !exists {
		return fmt.Errorf("framework not found: %s", assessment.FrameworkID)
	}

	// Set timestamps
	now := time.Now()
	assessment.CreatedAt = now
	assessment.UpdatedAt = now

	// Store assessment
	cfs.assessments[assessment.ID] = assessment

	cfs.logger.Info("Created compliance assessment", map[string]interface{}{
		"assessment_id": assessment.ID,
		"business_id":   assessment.BusinessID,
		"framework_id":  assessment.FrameworkID,
	})

	return nil
}

// GetAssessment retrieves a specific compliance assessment
func (cfs *ComplianceFrameworkService) GetAssessment(ctx context.Context, assessmentID string) (*ComplianceAssessment, error) {
	cfs.logger.Info("Retrieving compliance assessment", map[string]interface{}{
		"assessment_id": assessmentID,
	})

	assessment, exists := cfs.assessments[assessmentID]
	if !exists {
		return nil, fmt.Errorf("assessment not found: %s", assessmentID)
	}

	cfs.logger.Info("Retrieved compliance assessment", map[string]interface{}{
		"assessment_id": assessmentID,
		"business_id":   assessment.BusinessID,
		"framework_id":  assessment.FrameworkID,
		"status":        assessment.Status,
	})

	return assessment, nil
}

// UpdateAssessment updates an existing compliance assessment
func (cfs *ComplianceFrameworkService) UpdateAssessment(ctx context.Context, assessment *ComplianceAssessment) error {
	cfs.logger.Info("Updating compliance assessment", map[string]interface{}{
		"assessment_id": assessment.ID,
		"business_id":   assessment.BusinessID,
	})

	// Check if assessment exists
	if _, exists := cfs.assessments[assessment.ID]; !exists {
		return fmt.Errorf("assessment not found: %s", assessment.ID)
	}

	// Update timestamp
	assessment.UpdatedAt = time.Now()

	// Store updated assessment
	cfs.assessments[assessment.ID] = assessment

	cfs.logger.Info("Updated compliance assessment", map[string]interface{}{
		"assessment_id": assessment.ID,
		"business_id":   assessment.BusinessID,
		"status":        assessment.Status,
	})

	return nil
}

// GetBusinessAssessments retrieves all assessments for a business
func (cfs *ComplianceFrameworkService) GetBusinessAssessments(ctx context.Context, businessID string, frameworkID string) ([]*ComplianceAssessment, error) {
	cfs.logger.Info("Retrieving business assessments", map[string]interface{}{
		"business_id":  businessID,
		"framework_id": frameworkID,
	})

	var assessments []*ComplianceAssessment

	for _, assessment := range cfs.assessments {
		if assessment.BusinessID == businessID {
			if frameworkID == "" || assessment.FrameworkID == frameworkID {
				assessments = append(assessments, assessment)
			}
		}
	}

	// Sort by creation date (newest first)
	sort.Slice(assessments, func(i, j int) bool {
		return assessments[i].CreatedAt.After(assessments[j].CreatedAt)
	})

	cfs.logger.Info("Retrieved business assessments", map[string]interface{}{
		"business_id":  businessID,
		"framework_id": frameworkID,
		"count":        len(assessments),
	})

	return assessments, nil
}

// initializeDefaultFrameworks initializes the service with default compliance frameworks
func (cfs *ComplianceFrameworkService) initializeDefaultFrameworks() {
	now := time.Now()

	// SOC 2 Type II Framework
	soc2Framework := &ComplianceFramework{
		ID:            "SOC2",
		Name:          "SOC 2 Type II",
		Description:   "Service Organization Control 2 Type II - Security, Availability, Processing Integrity, Confidentiality, and Privacy",
		Version:       "2017",
		Category:      "security",
		Status:        "active",
		EffectiveDate: now.AddDate(-5, 0, 0),
		Scope:         []string{"all"},
		Jurisdiction:  []string{"US", "Global"},
		Authority:     "AICPA",
		Documentation: []string{"SOC 2 Trust Services Criteria"},
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// GDPR Framework
	gdprFramework := &ComplianceFramework{
		ID:            "GDPR",
		Name:          "General Data Protection Regulation",
		Description:   "EU General Data Protection Regulation - Data protection and privacy",
		Version:       "2018",
		Category:      "privacy",
		Status:        "active",
		EffectiveDate: now.AddDate(-6, 0, 0),
		Scope:         []string{"all"},
		Jurisdiction:  []string{"EU", "EEA", "UK"},
		Authority:     "European Commission",
		Documentation: []string{"GDPR Regulation (EU) 2016/679"},
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// PCI DSS Framework
	pciFramework := &ComplianceFramework{
		ID:            "PCI_DSS",
		Name:          "Payment Card Industry Data Security Standard",
		Description:   "PCI DSS - Security standards for organizations that handle credit card information",
		Version:       "4.0",
		Category:      "financial",
		Status:        "active",
		EffectiveDate: now.AddDate(-2, 0, 0),
		Scope:         []string{"financial", "ecommerce", "payment_processing"},
		Jurisdiction:  []string{"Global"},
		Authority:     "PCI Security Standards Council",
		Documentation: []string{"PCI DSS Requirements and Security Assessment Procedures"},
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// HIPAA Framework
	hipaaFramework := &ComplianceFramework{
		ID:            "HIPAA",
		Name:          "Health Insurance Portability and Accountability Act",
		Description:   "HIPAA - Healthcare data protection and privacy",
		Version:       "2013",
		Category:      "privacy",
		Status:        "active",
		EffectiveDate: now.AddDate(-10, 0, 0),
		Scope:         []string{"healthcare", "health_tech"},
		Jurisdiction:  []string{"US"},
		Authority:     "HHS",
		Documentation: []string{"HIPAA Privacy Rule", "HIPAA Security Rule"},
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Store frameworks
	cfs.frameworks["SOC2"] = soc2Framework
	cfs.frameworks["GDPR"] = gdprFramework
	cfs.frameworks["PCI_DSS"] = pciFramework
	cfs.frameworks["HIPAA"] = hipaaFramework

	// Initialize requirements for each framework
	cfs.initializeDefaultRequirements()

	cfs.logger.Info("Initialized default compliance frameworks", map[string]interface{}{
		"frameworks_count":   len(cfs.frameworks),
		"requirements_count": len(cfs.requirements),
	})
}

// initializeDefaultRequirements initializes default requirements for frameworks
func (cfs *ComplianceFrameworkService) initializeDefaultRequirements() {
	now := time.Now()

	// SOC 2 Requirements
	soc2Requirements := []*ComplianceRequirement{
		{
			ID:               "SOC2_CC6_1",
			FrameworkID:      "SOC2",
			Code:             "CC6.1",
			Name:             "Logical and Physical Access Controls",
			Description:      "The entity implements logical and physical access security measures to protect against threats from sources outside its system boundaries",
			Category:         "access_control",
			Priority:         "critical",
			Type:             "technical",
			Applicability:    []string{"all"},
			EvidenceTypes:    []string{"policy", "procedure", "test_result", "log"},
			AssessmentMethod: "hybrid",
			Frequency:        "continuous",
			Owner:            "security_team",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "SOC2_CC6_2",
			FrameworkID:      "SOC2",
			Code:             "CC6.2",
			Name:             "System Access",
			Description:      "Prior to issuing system credentials and granting system access, the entity registers and authorizes new internal and external users",
			Category:         "access_control",
			Priority:         "critical",
			Type:             "administrative",
			Applicability:    []string{"all"},
			EvidenceTypes:    []string{"policy", "procedure", "user_list", "approval"},
			AssessmentMethod: "manual",
			Frequency:        "monthly",
			Owner:            "security_team",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	// GDPR Requirements
	gdprRequirements := []*ComplianceRequirement{
		{
			ID:               "GDPR_32",
			FrameworkID:      "GDPR",
			Code:             "GDPR_32",
			Name:             "Security of Processing",
			Description:      "The controller and processor shall implement appropriate technical and organisational measures to ensure a level of security appropriate to the risk",
			Category:         "data_protection",
			Priority:         "critical",
			Type:             "technical",
			Applicability:    []string{"all"},
			EvidenceTypes:    []string{"policy", "procedure", "test_result", "audit"},
			AssessmentMethod: "hybrid",
			Frequency:        "continuous",
			Owner:            "privacy_team",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "GDPR_25",
			FrameworkID:      "GDPR",
			Code:             "GDPR_25",
			Name:             "Data Protection by Design and by Default",
			Description:      "The controller shall implement appropriate technical and organisational measures to ensure data protection by design and by default",
			Category:         "data_protection",
			Priority:         "high",
			Type:             "technical",
			Applicability:    []string{"all"},
			EvidenceTypes:    []string{"policy", "procedure", "design_document", "test_result"},
			AssessmentMethod: "hybrid",
			Frequency:        "quarterly",
			Owner:            "privacy_team",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	// Store requirements
	allRequirements := append(soc2Requirements, gdprRequirements...)
	for _, req := range allRequirements {
		cfs.requirements[req.ID] = req
	}

	// Update framework requirement lists
	cfs.frameworks["SOC2"].Requirements = []string{"SOC2_CC6_1", "SOC2_CC6_2"}
	cfs.frameworks["GDPR"].Requirements = []string{"GDPR_32", "GDPR_25"}
}
