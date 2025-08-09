package compliance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// PCIDSSTrackingService provides PCI DSS specific compliance tracking
type PCIDSSTrackingService struct {
	logger        *observability.Logger
	statusSystem  *ComplianceStatusSystem
	mappingSystem *FrameworkMappingSystem
	mu            sync.RWMutex
	pciData       map[string]*PCIDSSComplianceStatus // businessID -> PCI DSS status
}

// NewPCIDSSTrackingService creates a new PCI DSS tracking service
func NewPCIDSSTrackingService(logger *observability.Logger, statusSystem *ComplianceStatusSystem, mappingSystem *FrameworkMappingSystem) *PCIDSSTrackingService {
	return &PCIDSSTrackingService{
		logger:        logger,
		statusSystem:  statusSystem,
		mappingSystem: mappingSystem,
		pciData:       make(map[string]*PCIDSSComplianceStatus),
	}
}

// InitializePCIDSSTracking initializes PCI DSS tracking for a business
func (s *PCIDSSTrackingService) InitializePCIDSSTracking(ctx context.Context, businessID string, merchantLevel string, serviceProvider bool) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Initializing PCI DSS tracking for business",
		"request_id", requestID,
		"business_id", businessID,
		"merchant_level", merchantLevel,
		"service_provider", serviceProvider,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if PCI DSS tracking already exists
	if _, exists := s.pciData[businessID]; exists {
		return fmt.Errorf("PCI DSS tracking already initialized for business %s", businessID)
	}

	// Validate merchant level
	validLevels := []string{"Level 1", "Level 2", "Level 3", "Level 4"}
	validLevel := false
	for _, level := range validLevels {
		if merchantLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid merchant level: %s. Must be one of: %v", merchantLevel, validLevels)
	}

	// Create new PCI DSS compliance status
	pciStatus := &PCIDSSComplianceStatus{
		BusinessID:          businessID,
		Framework:           FrameworkPCIDSS,
		Version:             PCIDSSVersion4,
		MerchantLevel:       merchantLevel,
		ServiceProvider:     serviceProvider,
		OverallStatus:       ComplianceStatusNotStarted,
		ComplianceScore:     0.0,
		CategoryStatus:      make(map[string]CategoryStatus),
		RequirementsStatus:  make(map[string]PCIDSSRequirementStatus),
		LastAssessment:      time.Now(),
		NextAssessment:      time.Now().AddDate(1, 0, 0),
		AssessmentFrequency: "annually",
		ComplianceOfficer:   "system",
		Metadata:            make(map[string]interface{}),
	}

	// Initialize category status
	pciFramework := NewPCIDSSFramework()
	for _, category := range pciFramework.Categories {
		pciStatus.CategoryStatus[category.ID] = CategoryStatus{
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
	for _, requirement := range pciFramework.Requirements {
		pciStatus.RequirementsStatus[requirement.RequirementID] = PCIDSSRequirementStatus{
			RequirementID:        requirement.RequirementID,
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

	// Store the PCI DSS status
	s.pciData[businessID] = pciStatus

	// Register PCI DSS framework with mapping system if not already registered
	pciRegulatoryFramework := pciFramework.ConvertPCIDSSToRegulatoryFramework()
	err := s.mappingSystem.RegisterFramework(ctx, pciRegulatoryFramework)
	if err != nil {
		s.logger.Warn("Failed to register PCI DSS framework with mapping system",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		// Don't fail the initialization for this
	}

	s.logger.Info("PCI DSS tracking initialized successfully",
		"request_id", requestID,
		"business_id", businessID,
		"merchant_level", merchantLevel,
		"service_provider", serviceProvider,
		"category_count", len(pciStatus.CategoryStatus),
		"requirements_count", len(pciStatus.RequirementsStatus),
	)

	return nil
}

// GetPCIDSSStatus retrieves PCI DSS compliance status for a business
func (s *PCIDSSTrackingService) GetPCIDSSStatus(ctx context.Context, businessID string) (*PCIDSSComplianceStatus, error) {
	requestID := ctx.Value("request_id").(string)

	s.mu.RLock()
	defer s.mu.RUnlock()

	pciStatus, exists := s.pciData[businessID]
	if !exists {
		return nil, fmt.Errorf("PCI DSS tracking not initialized for business %s", businessID)
	}

	s.logger.Info("Retrieved PCI DSS status",
		"request_id", requestID,
		"business_id", businessID,
		"overall_status", pciStatus.OverallStatus,
		"compliance_score", pciStatus.ComplianceScore,
	)

	return pciStatus, nil
}

// UpdatePCIDSSRequirementStatus updates the status of a specific PCI DSS requirement
func (s *PCIDSSTrackingService) UpdatePCIDSSRequirementStatus(ctx context.Context, businessID, requirementID string, status ComplianceStatus, implementationStatus ImplementationStatus, score float64, reviewer string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating PCI DSS requirement status",
		"request_id", requestID,
		"business_id", businessID,
		"requirement_id", requirementID,
		"status", status,
		"implementation_status", implementationStatus,
		"score", score,
		"reviewer", reviewer,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	pciStatus, exists := s.pciData[businessID]
	if !exists {
		return fmt.Errorf("PCI DSS tracking not initialized for business %s", businessID)
	}

	reqStatus, exists := pciStatus.RequirementsStatus[requirementID]
	if !exists {
		return fmt.Errorf("requirement %s not found in PCI DSS framework", requirementID)
	}

	// Update requirement status
	reqStatus.Status = status
	reqStatus.ImplementationStatus = implementationStatus
	reqStatus.ComplianceScore = score
	reqStatus.LastReviewed = time.Now()
	reqStatus.Reviewer = reviewer

	pciStatus.RequirementsStatus[requirementID] = reqStatus

	// Update category status
	s.updateCategoryStatus(pciStatus, reqStatus.CategoryID)

	// Update overall status
	s.updateOverallStatus(pciStatus)

	s.logger.Info("PCI DSS requirement status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"requirement_id", requirementID,
		"new_status", status,
		"new_score", score,
	)

	return nil
}

// UpdatePCIDSSCategoryStatus updates the status of a specific PCI DSS category
func (s *PCIDSSTrackingService) UpdatePCIDSSCategoryStatus(ctx context.Context, businessID, categoryID string, status ComplianceStatus, score float64, reviewer string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating PCI DSS category status",
		"request_id", requestID,
		"business_id", businessID,
		"category_id", categoryID,
		"status", status,
		"score", score,
		"reviewer", reviewer,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	pciStatus, exists := s.pciData[businessID]
	if !exists {
		return fmt.Errorf("PCI DSS tracking not initialized for business %s", businessID)
	}

	categoryStatus, exists := pciStatus.CategoryStatus[categoryID]
	if !exists {
		return fmt.Errorf("category %s not found in PCI DSS framework", categoryID)
	}

	// Update category status
	categoryStatus.Status = status
	categoryStatus.Score = score
	categoryStatus.LastReviewed = time.Now()
	categoryStatus.Reviewer = reviewer

	pciStatus.CategoryStatus[categoryID] = categoryStatus

	// Update overall status
	s.updateOverallStatus(pciStatus)

	s.logger.Info("PCI DSS category status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"category_id", categoryID,
		"new_status", status,
		"new_score", score,
	)

	return nil
}

// AssessPCIDSSCompliance performs a comprehensive PCI DSS compliance assessment
func (s *PCIDSSTrackingService) AssessPCIDSSCompliance(ctx context.Context, businessID string) (*PCIDSSComplianceStatus, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Starting PCI DSS compliance assessment",
		"request_id", requestID,
		"business_id", businessID,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	pciStatus, exists := s.pciData[businessID]
	if !exists {
		return nil, fmt.Errorf("PCI DSS tracking not initialized for business %s", businessID)
	}

	// Perform assessment for each category
	pciFramework := NewPCIDSSFramework()
	for _, category := range pciFramework.Categories {
		categoryScore := s.assessCategory(ctx, pciStatus, category.ID)
		categoryStatus := pciStatus.CategoryStatus[category.ID]
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

		pciStatus.CategoryStatus[category.ID] = categoryStatus
	}

	// Update overall status
	s.updateOverallStatus(pciStatus)

	// Update assessment timestamp
	pciStatus.LastAssessment = time.Now()
	pciStatus.NextAssessment = time.Now().AddDate(1, 0, 0)

	s.logger.Info("PCI DSS compliance assessment completed",
		"request_id", requestID,
		"business_id", businessID,
		"overall_status", pciStatus.OverallStatus,
		"compliance_score", pciStatus.ComplianceScore,
	)

	return pciStatus, nil
}

// GetPCIDSSReport generates a PCI DSS compliance report
func (s *PCIDSSTrackingService) GetPCIDSSReport(ctx context.Context, businessID string, reportType string) (*ComplianceReport, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Generating PCI DSS compliance report",
		"request_id", requestID,
		"business_id", businessID,
		"report_type", reportType,
	)

	pciStatus, err := s.GetPCIDSSStatus(ctx, businessID)
	if err != nil {
		return nil, err
	}

	// Convert PCI DSS status to compliance report
	report := &ComplianceReport{
		ID:               fmt.Sprintf("pci_dss_report_%s_%s", businessID, time.Now().Format("20060102")),
		BusinessID:       businessID,
		Framework:        FrameworkPCIDSS,
		ReportType:       ReportTypeStatus,
		Title:            fmt.Sprintf("PCI DSS Compliance Report - %s", businessID),
		Description:      "PCI DSS compliance assessment report",
		GeneratedAt:      time.Now(),
		GeneratedBy:      "system",
		Period:           "annual",
		OverallStatus:    pciStatus.OverallStatus,
		ComplianceScore:  pciStatus.ComplianceScore,
		Requirements:     []RequirementReport{},
		Controls:         []ControlReport{},
		Exceptions:       []ExceptionReport{},
		RemediationPlans: []RemediationReport{},
		Recommendations:  []ComplianceRecommendation{},
		Metadata:         make(map[string]interface{}),
	}

	// Add PCI DSS specific metadata
	report.Metadata["pci_dss_version"] = pciStatus.Version
	report.Metadata["merchant_level"] = pciStatus.MerchantLevel
	report.Metadata["service_provider"] = pciStatus.ServiceProvider
	report.Metadata["category_count"] = len(pciStatus.CategoryStatus)
	report.Metadata["requirements_count"] = len(pciStatus.RequirementsStatus)

	// Convert requirements to report format
	for reqID, reqStatus := range pciStatus.RequirementsStatus {
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

	s.logger.Info("PCI DSS compliance report generated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"report_id", report.ID,
		"requirements_count", len(report.Requirements),
	)

	return report, nil
}

// updateCategoryStatus updates the status of a category based on its requirements
func (s *PCIDSSTrackingService) updateCategoryStatus(pciStatus *PCIDSSComplianceStatus, categoryID string) {
	categoryStatus := pciStatus.CategoryStatus[categoryID]

	var totalScore float64
	var requirementCount int
	var implementedCount int
	var verifiedCount int
	var nonCompliantCount int
	var exemptCount int

	// Calculate category status based on requirements
	for _, reqStatus := range pciStatus.RequirementsStatus {
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

	pciStatus.CategoryStatus[categoryID] = categoryStatus
}

// updateOverallStatus updates the overall PCI DSS compliance status
func (s *PCIDSSTrackingService) updateOverallStatus(pciStatus *PCIDSSComplianceStatus) {
	var totalScore float64
	var categoryCount int

	// Calculate overall score based on category scores
	for _, categoryStatus := range pciStatus.CategoryStatus {
		totalScore += categoryStatus.Score
		categoryCount++
	}

	if categoryCount > 0 {
		pciStatus.ComplianceScore = totalScore / float64(categoryCount)
	}

	// Determine overall status based on score
	if pciStatus.ComplianceScore >= 90.0 {
		pciStatus.OverallStatus = ComplianceStatusVerified
	} else if pciStatus.ComplianceScore >= 70.0 {
		pciStatus.OverallStatus = ComplianceStatusImplemented
	} else if pciStatus.ComplianceScore >= 30.0 {
		pciStatus.OverallStatus = ComplianceStatusInProgress
	} else {
		pciStatus.OverallStatus = ComplianceStatusNotStarted
	}
}

// assessCategory performs assessment for a specific category
func (s *PCIDSSTrackingService) assessCategory(ctx context.Context, pciStatus *PCIDSSComplianceStatus, categoryID string) float64 {
	var totalScore float64
	var requirementCount int

	// Calculate category score based on requirements
	for _, reqStatus := range pciStatus.RequirementsStatus {
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
