package compliance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// SOC2TrackingService provides SOC 2 specific compliance tracking
type SOC2TrackingService struct {
	logger        *observability.Logger
	statusSystem  *ComplianceStatusSystem
	mappingSystem *FrameworkMappingSystem
	mu            sync.RWMutex
	soc2Data      map[string]*SOC2ComplianceStatus // businessID -> SOC2 status
}

// NewSOC2TrackingService creates a new SOC 2 tracking service
func NewSOC2TrackingService(logger *observability.Logger, statusSystem *ComplianceStatusSystem, mappingSystem *FrameworkMappingSystem) *SOC2TrackingService {
	return &SOC2TrackingService{
		logger:        logger,
		statusSystem:  statusSystem,
		mappingSystem: mappingSystem,
		soc2Data:      make(map[string]*SOC2ComplianceStatus),
	}
}

// InitializeSOC2Tracking initializes SOC 2 tracking for a business
func (s *SOC2TrackingService) InitializeSOC2Tracking(ctx context.Context, businessID string, reportType string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Initializing SOC 2 tracking for business",
		"request_id", requestID,
		"business_id", businessID,
		"report_type", reportType,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if SOC 2 tracking already exists
	if _, exists := s.soc2Data[businessID]; exists {
		return fmt.Errorf("SOC 2 tracking already initialized for business %s", businessID)
	}

	// Create new SOC 2 compliance status
	soc2Status := &SOC2ComplianceStatus{
		BusinessID:          businessID,
		Framework:           FrameworkSOC2,
		ReportType:          reportType,
		AssessmentPeriod:    "annual",
		OverallStatus:       ComplianceStatusNotStarted,
		ComplianceScore:     0.0,
		CriteriaStatus:      make(map[string]CriteriaStatus),
		RequirementsStatus:  make(map[string]SOC2RequirementStatus),
		LastAssessment:      time.Now(),
		NextAssessment:      time.Now().AddDate(1, 0, 0),
		AssessmentFrequency: "annually",
		ComplianceOfficer:   "system",
		Metadata:            make(map[string]interface{}),
	}

	// Initialize criteria status
	soc2Framework := NewSOC2Framework()
	for _, criteria := range soc2Framework.Criteria {
		soc2Status.CriteriaStatus[criteria.Code] = CriteriaStatus{
			CriteriaID:       criteria.Code,
			CriteriaName:     criteria.Name,
			Status:           ComplianceStatusNotStarted,
			Score:            0.0,
			RequirementCount: 0,
			LastReviewed:     time.Now(),
			NextReview:       time.Now().AddDate(0, 6, 0),
			Reviewer:         "system",
		}
	}

	// Initialize requirements status
	for _, requirement := range soc2Framework.Requirements {
		soc2Status.RequirementsStatus[requirement.RequirementID] = SOC2RequirementStatus{
			RequirementID:        requirement.RequirementID,
			CriteriaID:           requirement.Criteria,
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

	// Store the SOC 2 status
	s.soc2Data[businessID] = soc2Status

	// Register SOC 2 framework with mapping system if not already registered
	soc2RegulatoryFramework := soc2Framework.ConvertSOC2ToRegulatoryFramework()
	err := s.mappingSystem.RegisterFramework(ctx, soc2RegulatoryFramework)
	if err != nil {
		s.logger.Warn("Failed to register SOC 2 framework with mapping system",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		// Don't fail the initialization for this
	}

	s.logger.Info("SOC 2 tracking initialized successfully",
		"request_id", requestID,
		"business_id", businessID,
		"report_type", reportType,
		"criteria_count", len(soc2Status.CriteriaStatus),
		"requirements_count", len(soc2Status.RequirementsStatus),
	)

	return nil
}

// GetSOC2Status retrieves SOC 2 compliance status for a business
func (s *SOC2TrackingService) GetSOC2Status(ctx context.Context, businessID string) (*SOC2ComplianceStatus, error) {
	requestID := ctx.Value("request_id").(string)

	s.mu.RLock()
	defer s.mu.RUnlock()

	soc2Status, exists := s.soc2Data[businessID]
	if !exists {
		return nil, fmt.Errorf("SOC 2 tracking not initialized for business %s", businessID)
	}

	s.logger.Info("Retrieved SOC 2 status",
		"request_id", requestID,
		"business_id", businessID,
		"overall_status", soc2Status.OverallStatus,
		"compliance_score", soc2Status.ComplianceScore,
	)

	return soc2Status, nil
}

// UpdateSOC2RequirementStatus updates the status of a specific SOC 2 requirement
func (s *SOC2TrackingService) UpdateSOC2RequirementStatus(ctx context.Context, businessID, requirementID string, status ComplianceStatus, implementationStatus ImplementationStatus, score float64, reviewer string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating SOC 2 requirement status",
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

	soc2Status, exists := s.soc2Data[businessID]
	if !exists {
		return fmt.Errorf("SOC 2 tracking not initialized for business %s", businessID)
	}

	reqStatus, exists := soc2Status.RequirementsStatus[requirementID]
	if !exists {
		return fmt.Errorf("requirement %s not found in SOC 2 framework", requirementID)
	}

	// Update requirement status
	reqStatus.Status = status
	reqStatus.ImplementationStatus = implementationStatus
	reqStatus.ComplianceScore = score
	reqStatus.LastReviewed = time.Now()
	reqStatus.Reviewer = reviewer

	soc2Status.RequirementsStatus[requirementID] = reqStatus

	// Update criteria status
	s.updateCriteriaStatus(soc2Status, reqStatus.CriteriaID)

	// Update overall status
	s.updateOverallStatus(soc2Status)

	s.logger.Info("SOC 2 requirement status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"requirement_id", requirementID,
		"new_status", status,
		"new_score", score,
	)

	return nil
}

// UpdateSOC2CriteriaStatus updates the status of a specific SOC 2 criteria
func (s *SOC2TrackingService) UpdateSOC2CriteriaStatus(ctx context.Context, businessID, criteriaID string, status ComplianceStatus, score float64, reviewer string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating SOC 2 criteria status",
		"request_id", requestID,
		"business_id", businessID,
		"criteria_id", criteriaID,
		"status", status,
		"score", score,
		"reviewer", reviewer,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	soc2Status, exists := s.soc2Data[businessID]
	if !exists {
		return fmt.Errorf("SOC 2 tracking not initialized for business %s", businessID)
	}

	criteriaStatus, exists := soc2Status.CriteriaStatus[criteriaID]
	if !exists {
		return fmt.Errorf("criteria %s not found in SOC 2 framework", criteriaID)
	}

	// Update criteria status
	criteriaStatus.Status = status
	criteriaStatus.Score = score
	criteriaStatus.LastReviewed = time.Now()
	criteriaStatus.Reviewer = reviewer

	soc2Status.CriteriaStatus[criteriaID] = criteriaStatus

	// Update overall status
	s.updateOverallStatus(soc2Status)

	s.logger.Info("SOC 2 criteria status updated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"criteria_id", criteriaID,
		"new_status", status,
		"new_score", score,
	)

	return nil
}

// AssessSOC2Compliance performs a comprehensive SOC 2 compliance assessment
func (s *SOC2TrackingService) AssessSOC2Compliance(ctx context.Context, businessID string) (*SOC2ComplianceStatus, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Starting SOC 2 compliance assessment",
		"request_id", requestID,
		"business_id", businessID,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	soc2Status, exists := s.soc2Data[businessID]
	if !exists {
		return nil, fmt.Errorf("SOC 2 tracking not initialized for business %s", businessID)
	}

	// Perform assessment for each criteria
	soc2Framework := NewSOC2Framework()
	for _, criteria := range soc2Framework.Criteria {
		criteriaScore := s.assessCriteria(ctx, soc2Status, criteria.Code)
		criteriaStatus := soc2Status.CriteriaStatus[criteria.Code]
		criteriaStatus.Score = criteriaScore
		criteriaStatus.LastReviewed = time.Now()
		criteriaStatus.Reviewer = "system"

		// Determine criteria status based on score
		if criteriaScore >= 90.0 {
			criteriaStatus.Status = ComplianceStatusVerified
		} else if criteriaScore >= 70.0 {
			criteriaStatus.Status = ComplianceStatusImplemented
		} else if criteriaScore >= 30.0 {
			criteriaStatus.Status = ComplianceStatusInProgress
		} else {
			criteriaStatus.Status = ComplianceStatusNotStarted
		}

		soc2Status.CriteriaStatus[criteria.Code] = criteriaStatus
	}

	// Update overall status
	s.updateOverallStatus(soc2Status)

	// Update assessment timestamp
	soc2Status.LastAssessment = time.Now()
	soc2Status.NextAssessment = time.Now().AddDate(1, 0, 0)

	s.logger.Info("SOC 2 compliance assessment completed",
		"request_id", requestID,
		"business_id", businessID,
		"overall_status", soc2Status.OverallStatus,
		"compliance_score", soc2Status.ComplianceScore,
	)

	return soc2Status, nil
}

// GetSOC2Report generates a SOC 2 compliance report
func (s *SOC2TrackingService) GetSOC2Report(ctx context.Context, businessID string, reportType string) (*ComplianceReport, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Generating SOC 2 compliance report",
		"request_id", requestID,
		"business_id", businessID,
		"report_type", reportType,
	)

	soc2Status, err := s.GetSOC2Status(ctx, businessID)
	if err != nil {
		return nil, err
	}

	// Convert SOC 2 status to compliance report
	report := &ComplianceReport{
		ID:               fmt.Sprintf("soc2_report_%s_%s", businessID, time.Now().Format("20060102")),
		BusinessID:       businessID,
		Framework:        FrameworkSOC2,
		ReportType:       ReportTypeStatus,
		Title:            fmt.Sprintf("SOC 2 Compliance Report - %s", businessID),
		Description:      "SOC 2 Trust Services Criteria compliance assessment report",
		GeneratedAt:      time.Now(),
		GeneratedBy:      "system",
		Period:           "annual",
		OverallStatus:    soc2Status.OverallStatus,
		ComplianceScore:  soc2Status.ComplianceScore,
		Requirements:     []RequirementReport{},
		Controls:         []ControlReport{},
		Exceptions:       []ExceptionReport{},
		RemediationPlans: []RemediationReport{},
		Recommendations:  []ComplianceRecommendation{},
		Metadata:         make(map[string]interface{}),
	}

	// Add SOC 2 specific metadata
	report.Metadata["soc2_report_type"] = soc2Status.ReportType
	report.Metadata["soc2_assessment_period"] = soc2Status.AssessmentPeriod
	report.Metadata["soc2_criteria_count"] = len(soc2Status.CriteriaStatus)
	report.Metadata["soc2_requirements_count"] = len(soc2Status.RequirementsStatus)

	// Convert requirements to report format
	for reqID, reqStatus := range soc2Status.RequirementsStatus {
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

	s.logger.Info("SOC 2 compliance report generated successfully",
		"request_id", requestID,
		"business_id", businessID,
		"report_id", report.ID,
		"requirements_count", len(report.Requirements),
	)

	return report, nil
}

// updateCriteriaStatus updates the status of a criteria based on its requirements
func (s *SOC2TrackingService) updateCriteriaStatus(soc2Status *SOC2ComplianceStatus, criteriaID string) {
	criteriaStatus := soc2Status.CriteriaStatus[criteriaID]

	var totalScore float64
	var requirementCount int
	var implementedCount int
	var verifiedCount int
	var nonCompliantCount int
	var exemptCount int

	// Calculate criteria status based on requirements
	for _, reqStatus := range soc2Status.RequirementsStatus {
		if reqStatus.CriteriaID == criteriaID {
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
		criteriaStatus.Score = totalScore / float64(requirementCount)
		criteriaStatus.RequirementCount = requirementCount
		criteriaStatus.ImplementedCount = implementedCount
		criteriaStatus.VerifiedCount = verifiedCount
		criteriaStatus.NonCompliantCount = nonCompliantCount
		criteriaStatus.ExemptCount = exemptCount

		// Determine criteria status based on score
		if criteriaStatus.Score >= 90.0 {
			criteriaStatus.Status = ComplianceStatusVerified
		} else if criteriaStatus.Score >= 70.0 {
			criteriaStatus.Status = ComplianceStatusImplemented
		} else if criteriaStatus.Score >= 30.0 {
			criteriaStatus.Status = ComplianceStatusInProgress
		} else {
			criteriaStatus.Status = ComplianceStatusNotStarted
		}
	}

	soc2Status.CriteriaStatus[criteriaID] = criteriaStatus
}

// updateOverallStatus updates the overall SOC 2 compliance status
func (s *SOC2TrackingService) updateOverallStatus(soc2Status *SOC2ComplianceStatus) {
	var totalScore float64
	var criteriaCount int

	// Calculate overall score based on criteria scores
	for _, criteriaStatus := range soc2Status.CriteriaStatus {
		totalScore += criteriaStatus.Score
		criteriaCount++
	}

	if criteriaCount > 0 {
		soc2Status.ComplianceScore = totalScore / float64(criteriaCount)
	}

	// Determine overall status based on score
	if soc2Status.ComplianceScore >= 90.0 {
		soc2Status.OverallStatus = ComplianceStatusVerified
	} else if soc2Status.ComplianceScore >= 70.0 {
		soc2Status.OverallStatus = ComplianceStatusImplemented
	} else if soc2Status.ComplianceScore >= 30.0 {
		soc2Status.OverallStatus = ComplianceStatusInProgress
	} else {
		soc2Status.OverallStatus = ComplianceStatusNotStarted
	}
}

// assessCriteria performs assessment for a specific criteria
func (s *SOC2TrackingService) assessCriteria(ctx context.Context, soc2Status *SOC2ComplianceStatus, criteriaID string) float64 {
	var totalScore float64
	var requirementCount int

	// Calculate criteria score based on requirements
	for _, reqStatus := range soc2Status.RequirementsStatus {
		if reqStatus.CriteriaID == criteriaID {
			requirementCount++
			totalScore += reqStatus.ComplianceScore
		}
	}

	if requirementCount > 0 {
		return totalScore / float64(requirementCount)
	}

	return 0.0
}
