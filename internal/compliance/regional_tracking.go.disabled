package compliance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// RegionalTrackingService provides regional compliance tracking for multiple frameworks
type RegionalTrackingService struct {
	logger        *observability.Logger
	statusSystem  *ComplianceStatusSystem
	mappingSystem *FrameworkMappingSystem
	mu            sync.RWMutex
	regionalData  map[string]map[string]*RegionalComplianceStatus // businessID -> framework -> status
}

// NewRegionalTrackingService creates a new regional tracking service
func NewRegionalTrackingService(logger *observability.Logger, statusSystem *ComplianceStatusSystem, mappingSystem *FrameworkMappingSystem) *RegionalTrackingService {
	return &RegionalTrackingService{
		logger:        logger,
		statusSystem:  statusSystem,
		mappingSystem: mappingSystem,
		regionalData:  make(map[string]map[string]*RegionalComplianceStatus),
	}
}

// InitializeRegionalTracking initializes regional compliance tracking for a business
func (s *RegionalTrackingService) InitializeRegionalTracking(ctx context.Context, businessID string, framework string, jurisdiction string, dataController bool, dataProcessor bool) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Initializing regional compliance tracking for business",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"jurisdiction", jurisdiction,
		"data_controller", dataController,
		"data_processor", dataProcessor,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate framework
	validFrameworks := []string{FrameworkCCPA, FrameworkLGPD, FrameworkPIPEDA, FrameworkPOPIA, FrameworkPDPA, FrameworkAPPI}
	validFramework := false
	for _, fw := range validFrameworks {
		if framework == fw {
			validFramework = true
			break
		}
	}
	if !validFramework {
		return fmt.Errorf("invalid framework: %s. Must be one of: %v", framework, validFrameworks)
	}

	// Check if regional tracking already exists for this business and framework
	if s.regionalData[businessID] == nil {
		s.regionalData[businessID] = make(map[string]*RegionalComplianceStatus)
	}
	if _, exists := s.regionalData[businessID][framework]; exists {
		return fmt.Errorf("regional tracking for framework %s already initialized for business %s", framework, businessID)
	}

	// Validate that at least one role is specified
	if !dataController && !dataProcessor {
		return fmt.Errorf("business must be either a data controller or data processor")
	}

	// Get framework definition
	var frameworkDef *RegionalFrameworkDefinition
	switch framework {
	case FrameworkCCPA:
		frameworkDef = NewCCPAFramework()
	case FrameworkLGPD:
		frameworkDef = NewLGPDFramework()
	case FrameworkPIPEDA:
		frameworkDef = NewPIPEDAFramework()
	case FrameworkPOPIA:
		frameworkDef = NewPOPIAFramework()
	case FrameworkPDPA:
		frameworkDef = NewPDPAFramework()
	case FrameworkAPPI:
		frameworkDef = NewAPPIFramework()
	default:
		return fmt.Errorf("framework %s not yet implemented", framework)
	}

	// Create new regional compliance status
	regionalStatus := &RegionalComplianceStatus{
		BusinessID:          businessID,
		Framework:           framework,
		Version:             frameworkDef.Version,
		Jurisdiction:        jurisdiction,
		DataController:      dataController,
		DataProcessor:       dataProcessor,
		OverallStatus:       ComplianceStatusNotStarted,
		ComplianceScore:     0.0,
		CategoryStatus:      make(map[string]RegionalCategoryStatus),
		RequirementsStatus:  make(map[string]RegionalRequirementStatus),
		LastAssessment:      time.Now(),
		NextAssessment:      time.Now().AddDate(1, 0, 0),
		AssessmentFrequency: "annually",
		ComplianceOfficer:   "system",
		Metadata:            make(map[string]interface{}),
	}

	// Initialize category status
	for _, category := range frameworkDef.Categories {
		regionalStatus.CategoryStatus[category.ID] = RegionalCategoryStatus{
			CategoryID:       category.ID,
			CategoryName:     category.Name,
			Status:           ComplianceStatusNotStarted,
			Score:            0.0,
			RequirementCount: 0,
			LastReviewed:     time.Now(),
			NextReview:       time.Now().AddDate(0, 6, 0),
			Reviewer:         "system",
		}
	}

	// Initialize requirements status
	for _, requirement := range frameworkDef.Requirements {
		regionalStatus.RequirementsStatus[requirement.RequirementID] = RegionalRequirementStatus{
			RequirementID:        requirement.RequirementID,
			FrameworkID:          framework,
			CategoryID:           requirement.Category,
			Title:                requirement.Title,
			Status:               ComplianceStatusNotStarted,
			ImplementationStatus: requirement.ImplementationStatus,
			ComplianceScore:      0.0,
			RiskLevel:            requirement.RiskLevel,
			Priority:             requirement.Priority,
			LastReviewed:         time.Now(),
			NextReview:           requirement.NextReviewDate,
			Reviewer:             "system",
			EvidenceCount:        0,
			ExceptionCount:       0,
			RemediationPlanCount: 0,
			Trend:                "stable",
			TrendStrength:        "none",
		}
	}

	// Store the regional status
	s.regionalData[businessID][framework] = regionalStatus

	// Register regional framework with mapping system if not already registered
	regionalRegulatoryFramework := frameworkDef.ConvertRegionalToRegulatoryFramework()
	err := s.mappingSystem.RegisterFramework(ctx, regionalRegulatoryFramework)
	if err != nil {
		s.logger.Warn("Failed to register regional framework with mapping system",
			"request_id", requestID,
			"business_id", businessID,
			"framework", framework,
			"error", err.Error(),
		)
		// Don't fail the initialization for this
	}

	s.logger.Info("Regional compliance tracking initialized successfully",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"jurisdiction", jurisdiction,
		"category_count", len(regionalStatus.CategoryStatus),
		"requirements_count", len(regionalStatus.RequirementsStatus),
	)

	return nil
}

// GetRegionalStatus retrieves regional compliance status for a business and framework
func (s *RegionalTrackingService) GetRegionalStatus(ctx context.Context, businessID string, framework string) (*RegionalComplianceStatus, error) {
	requestID := ctx.Value("request_id").(string)

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.regionalData[businessID] == nil {
		return nil, fmt.Errorf("regional tracking not initialized for business %s", businessID)
	}

	regionalStatus, exists := s.regionalData[businessID][framework]
	if !exists {
		return nil, fmt.Errorf("regional tracking for framework %s not initialized for business %s", framework, businessID)
	}

	s.logger.Info("Retrieved regional compliance status",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"overall_status", regionalStatus.OverallStatus,
		"compliance_score", regionalStatus.ComplianceScore,
	)

	return regionalStatus, nil
}

// UpdateRegionalRequirementStatus updates the status of a specific regional requirement
func (s *RegionalTrackingService) UpdateRegionalRequirementStatus(ctx context.Context, businessID, framework, requirementID string, status ComplianceStatus, implementationStatus ImplementationStatus, score float64, reviewer string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating regional requirement status",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"requirement_id", requirementID,
		"status", status,
		"implementation_status", implementationStatus,
		"score", score,
		"reviewer", reviewer,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.regionalData[businessID] == nil {
		return fmt.Errorf("regional tracking not initialized for business %s", businessID)
	}

	regionalStatus, exists := s.regionalData[businessID][framework]
	if !exists {
		return fmt.Errorf("regional tracking for framework %s not initialized for business %s", framework, businessID)
	}

	reqStatus, exists := regionalStatus.RequirementsStatus[requirementID]
	if !exists {
		return fmt.Errorf("requirement %s not found in regional framework %s", requirementID, framework)
	}

	// Update requirement status
	reqStatus.Status = status
	reqStatus.ImplementationStatus = implementationStatus
	reqStatus.ComplianceScore = score
	reqStatus.LastReviewed = time.Now()
	reqStatus.Reviewer = reviewer

	regionalStatus.RequirementsStatus[requirementID] = reqStatus

	// Update category status
	s.updateRegionalCategoryStatus(regionalStatus, reqStatus.CategoryID)

	// Update overall status
	s.updateRegionalOverallStatus(regionalStatus)

	s.logger.Info("Regional requirement status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"requirement_id", requirementID,
		"new_status", status,
		"new_score", score,
	)

	return nil
}

// UpdateRegionalCategoryStatus updates the status of a specific regional category
func (s *RegionalTrackingService) UpdateRegionalCategoryStatus(ctx context.Context, businessID, framework, categoryID string, status ComplianceStatus, score float64, reviewer string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating regional category status",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"category_id", categoryID,
		"status", status,
		"score", score,
		"reviewer", reviewer,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.regionalData[businessID] == nil {
		return fmt.Errorf("regional tracking not initialized for business %s", businessID)
	}

	regionalStatus, exists := s.regionalData[businessID][framework]
	if !exists {
		return fmt.Errorf("regional tracking for framework %s not initialized for business %s", framework, businessID)
	}

	categoryStatus, exists := regionalStatus.CategoryStatus[categoryID]
	if !exists {
		return fmt.Errorf("category %s not found in regional framework %s", categoryID, framework)
	}

	// Update category status
	categoryStatus.Status = status
	categoryStatus.Score = score
	categoryStatus.LastReviewed = time.Now()
	categoryStatus.Reviewer = reviewer

	regionalStatus.CategoryStatus[categoryID] = categoryStatus

	// Update overall status
	s.updateRegionalOverallStatus(regionalStatus)

	s.logger.Info("Regional category status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"category_id", categoryID,
		"new_status", status,
		"new_score", score,
	)

	return nil
}

// AssessRegionalCompliance performs a comprehensive regional compliance assessment
func (s *RegionalTrackingService) AssessRegionalCompliance(ctx context.Context, businessID string, framework string) (*RegionalComplianceStatus, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Starting regional compliance assessment",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.regionalData[businessID] == nil {
		return nil, fmt.Errorf("regional tracking not initialized for business %s", businessID)
	}

	regionalStatus, exists := s.regionalData[businessID][framework]
	if !exists {
		return nil, fmt.Errorf("regional tracking for framework %s not initialized for business %s", framework, businessID)
	}

	// Get framework definition
	var frameworkDef *RegionalFrameworkDefinition
	switch framework {
	case FrameworkCCPA:
		frameworkDef = NewCCPAFramework()
	case FrameworkLGPD:
		frameworkDef = NewLGPDFramework()
	case FrameworkPIPEDA:
		frameworkDef = NewPIPEDAFramework()
	default:
		return nil, fmt.Errorf("framework %s not yet implemented", framework)
	}

	// Perform assessment for each category
	for _, category := range frameworkDef.Categories {
		categoryScore := s.assessRegionalCategory(ctx, regionalStatus, category.ID)
		categoryStatus := regionalStatus.CategoryStatus[category.ID]
		categoryStatus.Score = categoryScore
		categoryStatus.LastReviewed = time.Now()
		categoryStatus.Reviewer = "system"

		// Determine category status based on score
		if categoryScore >= 90.0 {
			categoryStatus.Status = ComplianceStatusVerified
		} else if categoryScore >= 70.0 {
			categoryStatus.Status = ComplianceStatusImplemented
		} else if categoryScore >= 30.0 {
			categoryStatus.Status = ComplianceStatusInProgress
		} else {
			categoryStatus.Status = ComplianceStatusNotStarted
		}

		regionalStatus.CategoryStatus[category.ID] = categoryStatus
	}

	// Update overall status
	s.updateRegionalOverallStatus(regionalStatus)

	// Update assessment timestamp
	regionalStatus.LastAssessment = time.Now()
	regionalStatus.NextAssessment = time.Now().AddDate(1, 0, 0)

	s.logger.Info("Regional compliance assessment completed",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"overall_status", regionalStatus.OverallStatus,
		"compliance_score", regionalStatus.ComplianceScore,
	)

	return regionalStatus, nil
}

// GetRegionalReport generates a regional compliance report
func (s *RegionalTrackingService) GetRegionalReport(ctx context.Context, businessID string, framework string, reportType string) (*ComplianceReport, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Generating regional compliance report",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"report_type", reportType,
	)

	regionalStatus, err := s.GetRegionalStatus(ctx, businessID, framework)
	if err != nil {
		return nil, err
	}

	// Convert regional status to compliance report
	report := &ComplianceReport{
		ID:               fmt.Sprintf("regional_report_%s_%s_%s", framework, businessID, time.Now().Format("20060102")),
		BusinessID:       businessID,
		Framework:        framework,
		ReportType:       ReportTypeStatus,
		Title:            fmt.Sprintf("%s Compliance Report - %s", framework, businessID),
		Description:      fmt.Sprintf("%s compliance assessment report", framework),
		GeneratedAt:      time.Now(),
		GeneratedBy:      "system",
		Period:           "annual",
		OverallStatus:    regionalStatus.OverallStatus,
		ComplianceScore:  regionalStatus.ComplianceScore,
		Requirements:     []RequirementReport{},
		Controls:         []ControlReport{},
		Exceptions:       []ExceptionReport{},
		RemediationPlans: []RemediationReport{},
		Recommendations:  []ComplianceRecommendation{},
		Metadata:         make(map[string]interface{}),
	}

	// Add regional specific metadata
	report.Metadata["framework"] = framework
	report.Metadata["version"] = regionalStatus.Version
	report.Metadata["jurisdiction"] = regionalStatus.Jurisdiction
	report.Metadata["data_controller"] = regionalStatus.DataController
	report.Metadata["data_processor"] = regionalStatus.DataProcessor
	report.Metadata["category_count"] = len(regionalStatus.CategoryStatus)
	report.Metadata["requirements_count"] = len(regionalStatus.RequirementsStatus)

	// Convert requirements to report format
	for reqID, reqStatus := range regionalStatus.RequirementsStatus {
		requirementReport := RequirementReport{
			RequirementID:        reqID,
			Title:                reqStatus.Title,
			Status:               reqStatus.Status,
			ImplementationStatus: reqStatus.ImplementationStatus,
			ComplianceScore:      reqStatus.ComplianceScore,
			RiskLevel:            reqStatus.RiskLevel,
			Priority:             reqStatus.Priority,
			LastReviewed:         reqStatus.LastReviewed,
			NextReview:           reqStatus.NextReview,
			Controls:             []ControlReport{},
			Exceptions:           []ExceptionReport{},
			RemediationPlans:     []RemediationReport{},
		}
		report.Requirements = append(report.Requirements, requirementReport)
	}

	s.logger.Info("Regional compliance report generated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"report_id", report.ID,
		"requirements_count", len(report.Requirements),
	)

	return report, nil
}

// GetSupportedFrameworks returns the list of supported regional frameworks
func (s *RegionalTrackingService) GetSupportedFrameworks(ctx context.Context) []string {
	return []string{
		FrameworkCCPA,
		FrameworkLGPD,
		FrameworkPIPEDA,
		FrameworkPOPIA,
		FrameworkPDPA,
		FrameworkAPPI,
	}
}

// updateRegionalCategoryStatus updates the status of a category based on its requirements
func (s *RegionalTrackingService) updateRegionalCategoryStatus(regionalStatus *RegionalComplianceStatus, categoryID string) {
	categoryStatus := regionalStatus.CategoryStatus[categoryID]

	var totalScore float64
	var requirementCount int
	var implementedCount int
	var verifiedCount int
	var nonCompliantCount int
	var exemptCount int

	// Calculate category status based on requirements
	for _, reqStatus := range regionalStatus.RequirementsStatus {
		if reqStatus.CategoryID == categoryID {
			requirementCount++
			totalScore += reqStatus.ComplianceScore

			switch reqStatus.Status {
			case ComplianceStatusVerified:
				verifiedCount++
			case ComplianceStatusImplemented:
				implementedCount++
			case ComplianceStatusNonCompliant:
				nonCompliantCount++
			case ComplianceStatusExempt:
				exemptCount++
			}
		}
	}

	if requirementCount > 0 {
		categoryStatus.Score = totalScore / float64(requirementCount)
		categoryStatus.RequirementCount = requirementCount
		categoryStatus.ImplementedCount = implementedCount
		categoryStatus.VerifiedCount = verifiedCount
		categoryStatus.NonCompliantCount = nonCompliantCount
		categoryStatus.ExemptCount = exemptCount

		// Determine category status based on score
		if categoryStatus.Score >= 90.0 {
			categoryStatus.Status = ComplianceStatusVerified
		} else if categoryStatus.Score >= 70.0 {
			categoryStatus.Status = ComplianceStatusImplemented
		} else if categoryStatus.Score >= 30.0 {
			categoryStatus.Status = ComplianceStatusInProgress
		} else {
			categoryStatus.Status = ComplianceStatusNotStarted
		}
	}

	regionalStatus.CategoryStatus[categoryID] = categoryStatus
}

// updateRegionalOverallStatus updates the overall regional compliance status
func (s *RegionalTrackingService) updateRegionalOverallStatus(regionalStatus *RegionalComplianceStatus) {
	var totalScore float64
	var categoryCount int

	// Calculate overall score based on category scores
	for _, categoryStatus := range regionalStatus.CategoryStatus {
		totalScore += categoryStatus.Score
		categoryCount++
	}

	if categoryCount > 0 {
		regionalStatus.ComplianceScore = totalScore / float64(categoryCount)
	}

	// Determine overall status based on score
	if regionalStatus.ComplianceScore >= 90.0 {
		regionalStatus.OverallStatus = ComplianceStatusVerified
	} else if regionalStatus.ComplianceScore >= 70.0 {
		regionalStatus.OverallStatus = ComplianceStatusImplemented
	} else if regionalStatus.ComplianceScore >= 30.0 {
		regionalStatus.OverallStatus = ComplianceStatusInProgress
	} else {
		regionalStatus.OverallStatus = ComplianceStatusNotStarted
	}
}

// assessRegionalCategory performs assessment for a specific category
func (s *RegionalTrackingService) assessRegionalCategory(ctx context.Context, regionalStatus *RegionalComplianceStatus, categoryID string) float64 {
	var totalScore float64
	var requirementCount int

	// Calculate category score based on requirements
	for _, reqStatus := range regionalStatus.RequirementsStatus {
		if reqStatus.CategoryID == categoryID {
			requirementCount++
			totalScore += reqStatus.ComplianceScore
		}
	}

	if requirementCount > 0 {
		return totalScore / float64(requirementCount)
	}

	return 0.0
}
