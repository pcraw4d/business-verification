package compliance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// GDPRTrackingService provides GDPR specific compliance tracking
type GDPRTrackingService struct {
	logger        *observability.Logger
	statusSystem  *ComplianceStatusSystem
	mappingSystem *FrameworkMappingSystem
	mu            sync.RWMutex
	gdprData      map[string]*GDPRComplianceStatus // businessID -> GDPR status
}

// NewGDPRTrackingService creates a new GDPR tracking service
func NewGDPRTrackingService(logger *observability.Logger, statusSystem *ComplianceStatusSystem, mappingSystem *FrameworkMappingSystem) *GDPRTrackingService {
	return &GDPRTrackingService{
		logger:        logger,
		statusSystem:  statusSystem,
		mappingSystem: mappingSystem,
		gdprData:      make(map[string]*GDPRComplianceStatus),
	}
}

// InitializeGDPRTracking initializes GDPR tracking for a business
func (s *GDPRTrackingService) InitializeGDPRTracking(ctx context.Context, businessID string, dataController bool, dataProcessor bool, dataProtectionOfficer string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Initializing GDPR tracking for business",
		"request_id", requestID,
		"business_id", businessID,
		"data_controller", dataController,
		"data_processor", dataProcessor,
		"data_protection_officer", dataProtectionOfficer,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if GDPR tracking already exists
	if _, exists := s.gdprData[businessID]; exists {
		return fmt.Errorf("GDPR tracking already initialized for business %s", businessID)
	}

	// Validate that at least one role is specified
	if !dataController && !dataProcessor {
		return fmt.Errorf("business must be either a data controller or data processor")
	}

	// Create new GDPR compliance status
	gdprStatus := &GDPRComplianceStatus{
		BusinessID:            businessID,
		Framework:             FrameworkGDPR,
		Version:               GDPRVersion2018,
		DataController:        dataController,
		DataProcessor:         dataProcessor,
		DataProtectionOfficer: dataProtectionOfficer,
		OverallStatus:         ComplianceStatusNotStarted,
		ComplianceScore:       0.0,
		PrincipleStatus:       make(map[string]PrincipleStatus),
		RightsStatus:          make(map[string]DataSubjectRightStatus),
		RequirementsStatus:    make(map[string]GDPRRequirementStatus),
		LastAssessment:        time.Now(),
		NextAssessment:        time.Now().AddDate(1, 0, 0),
		AssessmentFrequency:   "annually",
		ComplianceOfficer:     "system",
		Metadata:              make(map[string]interface{}),
	}

	// Initialize principle status
	gdprFramework := NewGDPRFramework()
	for _, principle := range gdprFramework.Principles {
		gdprStatus.PrincipleStatus[principle.ID] = PrincipleStatus{
			PrincipleID:      principle.ID,
			PrincipleName:    principle.Name,
			Status:           ComplianceStatusNotStarted,
			Score:            0.0,
			RequirementCount: 0,
			LastReviewed:     time.Now(),
			NextReview:       time.Now().AddDate(0, 6, 0),
			Reviewer:         "system",
		}
	}

	// Initialize data subject rights status
	for _, right := range gdprFramework.DataSubjectRights {
		gdprStatus.RightsStatus[right.ID] = DataSubjectRightStatus{
			RightID:          right.ID,
			RightName:        right.Name,
			Status:           ComplianceStatusNotStarted,
			Score:            0.0,
			RequirementCount: 0,
			LastReviewed:     time.Now(),
			NextReview:       time.Now().AddDate(0, 6, 0),
			Reviewer:         "system",
		}
	}

	// Initialize requirements status
	for _, requirement := range gdprFramework.Requirements {
		gdprStatus.RequirementsStatus[requirement.RequirementID] = GDPRRequirementStatus{
			RequirementID:        requirement.RequirementID,
			ArticleID:            requirement.Article,
			PrincipleID:          requirement.Principle,
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

	// Store the GDPR status
	s.gdprData[businessID] = gdprStatus

	// Register GDPR framework with mapping system if not already registered
	gdprRegulatoryFramework := gdprFramework.ConvertGDPRToRegulatoryFramework()
	err := s.mappingSystem.RegisterFramework(ctx, gdprRegulatoryFramework)
	if err != nil {
		s.logger.Warn("Failed to register GDPR framework with mapping system",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		// Don't fail the initialization for this
	}

	s.logger.Info("GDPR tracking initialized successfully",
		"request_id", requestID,
		"business_id", businessID,
		"data_controller", dataController,
		"data_processor", dataProcessor,
		"principle_count", len(gdprStatus.PrincipleStatus),
		"rights_count", len(gdprStatus.RightsStatus),
		"requirements_count", len(gdprStatus.RequirementsStatus),
	)

	return nil
}

// GetGDPRStatus retrieves GDPR compliance status for a business
func (s *GDPRTrackingService) GetGDPRStatus(ctx context.Context, businessID string) (*GDPRComplianceStatus, error) {
	requestID := ctx.Value("request_id").(string)

	s.mu.RLock()
	defer s.mu.RUnlock()

	gdprStatus, exists := s.gdprData[businessID]
	if !exists {
		return nil, fmt.Errorf("GDPR tracking not initialized for business %s", businessID)
	}

	s.logger.Info("Retrieved GDPR status",
		"request_id", requestID,
		"business_id", businessID,
		"overall_status", gdprStatus.OverallStatus,
		"compliance_score", gdprStatus.ComplianceScore,
	)

	return gdprStatus, nil
}

// UpdateGDPRRequirementStatus updates the status of a specific GDPR requirement
func (s *GDPRTrackingService) UpdateGDPRRequirementStatus(ctx context.Context, businessID, requirementID string, status ComplianceStatus, implementationStatus ImplementationStatus, score float64, reviewer string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating GDPR requirement status",
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

	gdprStatus, exists := s.gdprData[businessID]
	if !exists {
		return fmt.Errorf("GDPR tracking not initialized for business %s", businessID)
	}

	reqStatus, exists := gdprStatus.RequirementsStatus[requirementID]
	if !exists {
		return fmt.Errorf("requirement %s not found in GDPR framework", requirementID)
	}

	// Update requirement status
	reqStatus.Status = status
	reqStatus.ImplementationStatus = implementationStatus
	reqStatus.ComplianceScore = score
	reqStatus.LastReviewed = time.Now()
	reqStatus.Reviewer = reviewer

	gdprStatus.RequirementsStatus[requirementID] = reqStatus

	// Update principle status
	s.updatePrincipleStatus(gdprStatus, reqStatus.PrincipleID)

	// Update overall status
	s.updateOverallStatus(gdprStatus)

	s.logger.Info("GDPR requirement status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"requirement_id", requirementID,
		"new_status", status,
		"new_score", score,
	)

	return nil
}

// UpdateGDPRPrincipleStatus updates the status of a specific GDPR principle
func (s *GDPRTrackingService) UpdateGDPRPrincipleStatus(ctx context.Context, businessID, principleID string, status ComplianceStatus, score float64, reviewer string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating GDPR principle status",
		"request_id", requestID,
		"business_id", businessID,
		"principle_id", principleID,
		"status", status,
		"score", score,
		"reviewer", reviewer,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	gdprStatus, exists := s.gdprData[businessID]
	if !exists {
		return fmt.Errorf("GDPR tracking not initialized for business %s", businessID)
	}

	principleStatus, exists := gdprStatus.PrincipleStatus[principleID]
	if !exists {
		return fmt.Errorf("principle %s not found in GDPR framework", principleID)
	}

	// Update principle status
	principleStatus.Status = status
	principleStatus.Score = score
	principleStatus.LastReviewed = time.Now()
	principleStatus.Reviewer = reviewer

	gdprStatus.PrincipleStatus[principleID] = principleStatus

	// Update overall status
	s.updateOverallStatus(gdprStatus)

	s.logger.Info("GDPR principle status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"principle_id", principleID,
		"new_status", status,
		"new_score", score,
	)

	return nil
}

// UpdateGDPRDataSubjectRightStatus updates the status of a specific GDPR data subject right
func (s *GDPRTrackingService) UpdateGDPRDataSubjectRightStatus(ctx context.Context, businessID, rightID string, status ComplianceStatus, score float64, reviewer string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating GDPR data subject right status",
		"request_id", requestID,
		"business_id", businessID,
		"right_id", rightID,
		"status", status,
		"score", score,
		"reviewer", reviewer,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	gdprStatus, exists := s.gdprData[businessID]
	if !exists {
		return fmt.Errorf("GDPR tracking not initialized for business %s", businessID)
	}

	rightStatus, exists := gdprStatus.RightsStatus[rightID]
	if !exists {
		return fmt.Errorf("data subject right %s not found in GDPR framework", rightID)
	}

	// Update right status
	rightStatus.Status = status
	rightStatus.Score = score
	rightStatus.LastReviewed = time.Now()
	rightStatus.Reviewer = reviewer

	gdprStatus.RightsStatus[rightID] = rightStatus

	// Update overall status
	s.updateOverallStatus(gdprStatus)

	s.logger.Info("GDPR data subject right status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"right_id", rightID,
		"new_status", status,
		"new_score", score,
	)

	return nil
}

// AssessGDPRCompliance performs a comprehensive GDPR compliance assessment
func (s *GDPRTrackingService) AssessGDPRCompliance(ctx context.Context, businessID string) (*GDPRComplianceStatus, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Starting GDPR compliance assessment",
		"request_id", requestID,
		"business_id", businessID,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	gdprStatus, exists := s.gdprData[businessID]
	if !exists {
		return nil, fmt.Errorf("GDPR tracking not initialized for business %s", businessID)
	}

	// Perform assessment for each principle
	gdprFramework := NewGDPRFramework()
	for _, principle := range gdprFramework.Principles {
		principleScore := s.assessPrinciple(ctx, gdprStatus, principle.ID)
		principleStatus := gdprStatus.PrincipleStatus[principle.ID]
		principleStatus.Score = principleScore
		principleStatus.LastReviewed = time.Now()
		principleStatus.Reviewer = "system"

		// Determine principle status based on score
		if principleScore >= 90.0 {
			principleStatus.Status = ComplianceStatusVerified
		} else if principleScore >= 70.0 {
			principleStatus.Status = ComplianceStatusImplemented
		} else if principleScore >= 30.0 {
			principleStatus.Status = ComplianceStatusInProgress
		} else {
			principleStatus.Status = ComplianceStatusNotStarted
		}

		gdprStatus.PrincipleStatus[principle.ID] = principleStatus
	}

	// Perform assessment for each data subject right
	for _, right := range gdprFramework.DataSubjectRights {
		rightScore := s.assessDataSubjectRight(ctx, gdprStatus, right.ID)
		rightStatus := gdprStatus.RightsStatus[right.ID]
		rightStatus.Score = rightScore
		rightStatus.LastReviewed = time.Now()
		rightStatus.Reviewer = "system"

		// Determine right status based on score
		if rightScore >= 90.0 {
			rightStatus.Status = ComplianceStatusVerified
		} else if rightScore >= 70.0 {
			rightStatus.Status = ComplianceStatusImplemented
		} else if rightScore >= 30.0 {
			rightStatus.Status = ComplianceStatusInProgress
		} else {
			rightStatus.Status = ComplianceStatusNotStarted
		}

		gdprStatus.RightsStatus[right.ID] = rightStatus
	}

	// Update overall status
	s.updateOverallStatus(gdprStatus)

	// Update assessment timestamp
	gdprStatus.LastAssessment = time.Now()
	gdprStatus.NextAssessment = time.Now().AddDate(1, 0, 0)

	s.logger.Info("GDPR compliance assessment completed",
		"request_id", requestID,
		"business_id", businessID,
		"overall_status", gdprStatus.OverallStatus,
		"compliance_score", gdprStatus.ComplianceScore,
	)

	return gdprStatus, nil
}

// GetGDPRReport generates a GDPR compliance report
func (s *GDPRTrackingService) GetGDPRReport(ctx context.Context, businessID string, reportType string) (*ComplianceReport, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Generating GDPR compliance report",
		"request_id", requestID,
		"business_id", businessID,
		"report_type", reportType,
	)

	gdprStatus, err := s.GetGDPRStatus(ctx, businessID)
	if err != nil {
		return nil, err
	}

	// Convert GDPR status to compliance report
	report := &ComplianceReport{
		ID:               fmt.Sprintf("gdpr_report_%s_%s", businessID, time.Now().Format("20060102")),
		BusinessID:       businessID,
		Framework:        FrameworkGDPR,
		ReportType:       ReportTypeStatus,
		Title:            fmt.Sprintf("GDPR Compliance Report - %s", businessID),
		Description:      "GDPR compliance assessment report",
		GeneratedAt:      time.Now(),
		GeneratedBy:      "system",
		Period:           "annual",
		OverallStatus:    gdprStatus.OverallStatus,
		ComplianceScore:  gdprStatus.ComplianceScore,
		Requirements:     []RequirementReport{},
		Controls:         []ControlReport{},
		Exceptions:       []ExceptionReport{},
		RemediationPlans: []RemediationReport{},
		Recommendations:  []ComplianceRecommendation{},
		Metadata:         make(map[string]interface{}),
	}

	// Add GDPR specific metadata
	report.Metadata["gdpr_version"] = gdprStatus.Version
	report.Metadata["data_controller"] = gdprStatus.DataController
	report.Metadata["data_processor"] = gdprStatus.DataProcessor
	report.Metadata["data_protection_officer"] = gdprStatus.DataProtectionOfficer
	report.Metadata["principle_count"] = len(gdprStatus.PrincipleStatus)
	report.Metadata["rights_count"] = len(gdprStatus.RightsStatus)
	report.Metadata["requirements_count"] = len(gdprStatus.RequirementsStatus)

	// Convert requirements to report format
	for reqID, reqStatus := range gdprStatus.RequirementsStatus {
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

	s.logger.Info("GDPR compliance report generated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"report_id", report.ID,
		"requirements_count", len(report.Requirements),
	)

	return report, nil
}

// updatePrincipleStatus updates the status of a principle based on its requirements
func (s *GDPRTrackingService) updatePrincipleStatus(gdprStatus *GDPRComplianceStatus, principleID string) {
	principleStatus := gdprStatus.PrincipleStatus[principleID]

	var totalScore float64
	var requirementCount int
	var implementedCount int
	var verifiedCount int
	var nonCompliantCount int
	var exemptCount int

	// Calculate principle status based on requirements
	for _, reqStatus := range gdprStatus.RequirementsStatus {
		if reqStatus.PrincipleID == principleID {
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
		principleStatus.Score = totalScore / float64(requirementCount)
		principleStatus.RequirementCount = requirementCount
		principleStatus.ImplementedCount = implementedCount
		principleStatus.VerifiedCount = verifiedCount
		principleStatus.NonCompliantCount = nonCompliantCount
		principleStatus.ExemptCount = exemptCount

		// Determine principle status based on score
		if principleStatus.Score >= 90.0 {
			principleStatus.Status = ComplianceStatusVerified
		} else if principleStatus.Score >= 70.0 {
			principleStatus.Status = ComplianceStatusImplemented
		} else if principleStatus.Score >= 30.0 {
			principleStatus.Status = ComplianceStatusInProgress
		} else {
			principleStatus.Status = ComplianceStatusNotStarted
		}
	}

	gdprStatus.PrincipleStatus[principleID] = principleStatus
}

// updateOverallStatus updates the overall GDPR compliance status
func (s *GDPRTrackingService) updateOverallStatus(gdprStatus *GDPRComplianceStatus) {
	var totalScore float64
	var componentCount int

	// Calculate overall score based on principle and rights scores
	for _, principleStatus := range gdprStatus.PrincipleStatus {
		totalScore += principleStatus.Score
		componentCount++
	}

	for _, rightStatus := range gdprStatus.RightsStatus {
		totalScore += rightStatus.Score
		componentCount++
	}

	if componentCount > 0 {
		gdprStatus.ComplianceScore = totalScore / float64(componentCount)
	}

	// Determine overall status based on score
	if gdprStatus.ComplianceScore >= 90.0 {
		gdprStatus.OverallStatus = ComplianceStatusVerified
	} else if gdprStatus.ComplianceScore >= 70.0 {
		gdprStatus.OverallStatus = ComplianceStatusImplemented
	} else if gdprStatus.ComplianceScore >= 30.0 {
		gdprStatus.OverallStatus = ComplianceStatusInProgress
	} else {
		gdprStatus.OverallStatus = ComplianceStatusNotStarted
	}
}

// assessPrinciple performs assessment for a specific principle
func (s *GDPRTrackingService) assessPrinciple(ctx context.Context, gdprStatus *GDPRComplianceStatus, principleID string) float64 {
	var totalScore float64
	var requirementCount int

	// Calculate principle score based on requirements
	for _, reqStatus := range gdprStatus.RequirementsStatus {
		if reqStatus.PrincipleID == principleID {
			requirementCount++
			totalScore += reqStatus.ComplianceScore
		}
	}

	if requirementCount > 0 {
		return totalScore / float64(requirementCount)
	}

	return 0.0
}

// assessDataSubjectRight performs assessment for a specific data subject right
func (s *GDPRTrackingService) assessDataSubjectRight(ctx context.Context, gdprStatus *GDPRComplianceStatus, rightID string) float64 {
	var totalScore float64
	var requirementCount int

	// Calculate right score based on requirements
	for _, reqStatus := range gdprStatus.RequirementsStatus {
		// This is a simplified assessment - in a real implementation, you'd map requirements to rights
		// For now, we'll use a placeholder score
		requirementCount++
		totalScore += reqStatus.ComplianceScore
	}

	if requirementCount > 0 {
		return totalScore / float64(requirementCount)
	}

	return 0.0
}
