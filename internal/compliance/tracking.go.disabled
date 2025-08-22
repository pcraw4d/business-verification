package compliance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// TrackingSystem provides comprehensive compliance tracking functionality
type TrackingSystem struct {
	logger       *observability.Logger
	mu           sync.RWMutex
	trackingData map[string]*ComplianceTracking    // businessID -> tracking
	requirements map[string]*ComplianceRequirement // requirementID -> requirement
	frameworks   map[string]*RegulatoryFramework   // frameworkID -> framework
	auditTrail   []ComplianceAuditTrail
}

// NewTrackingSystem creates a new compliance tracking system
func NewTrackingSystem(logger *observability.Logger) *TrackingSystem {
	return &TrackingSystem{
		logger:       logger,
		trackingData: make(map[string]*ComplianceTracking),
		requirements: make(map[string]*ComplianceRequirement),
		frameworks:   make(map[string]*RegulatoryFramework),
		auditTrail:   make([]ComplianceAuditTrail, 0),
	}
}

// InitializeBusinessTracking initializes compliance tracking for a business
func (s *TrackingSystem) InitializeBusinessTracking(ctx context.Context, businessID string, frameworks []string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Initializing compliance tracking for business",
		"request_id", requestID,
		"business_id", businessID,
		"frameworks", frameworks,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize tracking for each framework
	for _, framework := range frameworks {
		tracking := &ComplianceTracking{
			ID:                  fmt.Sprintf("tracking_%s_%s", businessID, framework),
			BusinessID:          businessID,
			Framework:           framework,
			OverallStatus:       ComplianceStatusNotStarted,
			ComplianceScore:     0.0,
			Requirements:        make([]RequirementTracking, 0),
			LastAssessment:      time.Now(),
			NextAssessment:      time.Now().Add(30 * 24 * time.Hour), // 30 days
			AssessmentFrequency: "monthly",
			ComplianceOfficer:   "system",
		}

		// Load framework requirements
		frameworkData, exists := s.frameworks[framework]
		if exists {
			// Initialize requirement tracking for each requirement
			for _, requirement := range frameworkData.Requirements {
				reqTracking := RequirementTracking{
					RequirementID:        requirement.RequirementID,
					Status:               ComplianceStatusNotStarted,
					ImplementationStatus: ImplementationStatusNotImplemented,
					ComplianceScore:      0.0,
					LastReviewed:         time.Now(),
					NextReview:           time.Now().Add(30 * 24 * time.Hour),
					Reviewer:             "system",
					Evidence:             make([]TrackingEvidence, 0),
					Controls:             make([]ControlTracking, 0),
					Exceptions:           make([]ComplianceException, 0),
				}

				// Initialize control tracking for each control
				for _, control := range requirement.Controls {
					controlTracking := ControlTracking{
						ControlID:            control.ID,
						Status:               ComplianceStatusNotStarted,
						ImplementationStatus: ImplementationStatusNotImplemented,
						Effectiveness:        ControlEffectivenessIneffective,
						TestResults:          make([]ControlTestResult, 0),
						Evidence:             make([]ControlEvidence, 0),
					}
					reqTracking.Controls = append(reqTracking.Controls, controlTracking)
				}

				tracking.Requirements = append(tracking.Requirements, reqTracking)
			}
		}

		s.trackingData[fmt.Sprintf("%s_%s", businessID, framework)] = tracking
	}

	s.logAuditTrail(ctx, businessID, "", nil, AuditActionCreate, "Initialized compliance tracking", "", "")

	s.logger.Info("Compliance tracking initialized successfully",
		"request_id", requestID,
		"business_id", businessID,
		"frameworks", frameworks,
	)

	return nil
}

// UpdateRequirementStatus updates the status of a specific requirement
func (s *TrackingSystem) UpdateRequirementStatus(ctx context.Context, businessID, framework, requirementID string, status ComplianceStatus, implementationStatus ImplementationStatus, score float64, notes string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating requirement status",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"requirement_id", requirementID,
		"status", status,
		"implementation_status", implementationStatus,
		"score", score,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	trackingKey := fmt.Sprintf("%s_%s", businessID, framework)
	tracking, exists := s.trackingData[trackingKey]
	if !exists {
		return fmt.Errorf("compliance tracking not found for business %s and framework %s", businessID, framework)
	}

	// Find and update the requirement
	for i, req := range tracking.Requirements {
		if req.RequirementID == requirementID {
			oldStatus := req.Status
			oldScore := req.ComplianceScore

			tracking.Requirements[i].Status = status
			tracking.Requirements[i].ImplementationStatus = implementationStatus
			tracking.Requirements[i].ComplianceScore = score
			tracking.Requirements[i].LastReviewed = time.Now()
			tracking.Requirements[i].NextReview = time.Now().Add(30 * 24 * time.Hour)
			tracking.Requirements[i].Notes = notes

			// Update overall compliance score
			s.updateOverallComplianceScore(tracking)

			s.logAuditTrail(ctx, businessID, framework, &requirementID, AuditActionUpdate,
				fmt.Sprintf("Updated requirement status from %s to %s", oldStatus, status),
				fmt.Sprintf("%.2f", oldScore), fmt.Sprintf("%.2f", score))

			s.logger.Info("Requirement status updated successfully",
				"request_id", requestID,
				"business_id", businessID,
				"requirement_id", requirementID,
				"old_status", oldStatus,
				"new_status", status,
				"old_score", oldScore,
				"new_score", score,
			)

			return nil
		}
	}

	return fmt.Errorf("requirement %s not found in tracking", requirementID)
}

// UpdateControlStatus updates the status of a specific control
func (s *TrackingSystem) UpdateControlStatus(ctx context.Context, businessID, framework, requirementID, controlID string, status ComplianceStatus, implementationStatus ImplementationStatus, effectiveness ControlEffectiveness, notes string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating control status",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"requirement_id", requirementID,
		"control_id", controlID,
		"status", status,
		"effectiveness", effectiveness,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	trackingKey := fmt.Sprintf("%s_%s", businessID, framework)
	tracking, exists := s.trackingData[trackingKey]
	if !exists {
		return fmt.Errorf("compliance tracking not found for business %s and framework %s", businessID, framework)
	}

	// Find and update the control
	for i, req := range tracking.Requirements {
		if req.RequirementID == requirementID {
			for j, control := range req.Controls {
				if control.ControlID == controlID {
					oldStatus := control.Status
					oldEffectiveness := control.Effectiveness

					tracking.Requirements[i].Controls[j].Status = status
					tracking.Requirements[i].Controls[j].ImplementationStatus = implementationStatus
					tracking.Requirements[i].Controls[j].Effectiveness = effectiveness
					tracking.Requirements[i].Controls[j].Notes = notes

					// Update requirement score based on control effectiveness
					s.updateRequirementScore(tracking, i)

					s.logAuditTrail(ctx, businessID, framework, &requirementID, AuditActionUpdate,
						fmt.Sprintf("Updated control status from %s to %s", oldStatus, status),
						string(oldEffectiveness), string(effectiveness))

					s.logger.Info("Control status updated successfully",
						"request_id", requestID,
						"business_id", businessID,
						"control_id", controlID,
						"old_status", oldStatus,
						"new_status", status,
						"old_effectiveness", oldEffectiveness,
						"new_effectiveness", effectiveness,
					)

					return nil
				}
			}
		}
	}

	return fmt.Errorf("control %s not found in requirement %s", controlID, requirementID)
}

// AddControlTestResult adds a test result for a control
func (s *TrackingSystem) AddControlTestResult(ctx context.Context, businessID, framework, requirementID, controlID string, testResult ControlTestResult) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Adding control test result",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"requirement_id", requirementID,
		"control_id", controlID,
		"test_type", testResult.TestType,
		"result", testResult.Result,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	trackingKey := fmt.Sprintf("%s_%s", businessID, framework)
	tracking, exists := s.trackingData[trackingKey]
	if !exists {
		return fmt.Errorf("compliance tracking not found for business %s and framework %s", businessID, framework)
	}

	// Find and add test result to the control
	for i, req := range tracking.Requirements {
		if req.RequirementID == requirementID {
			for j, control := range req.Controls {
				if control.ControlID == controlID {
					tracking.Requirements[i].Controls[j].TestResults = append(tracking.Requirements[i].Controls[j].TestResults, testResult)
					tracking.Requirements[i].Controls[j].LastTested = &testResult.TestDate
					tracking.Requirements[i].Controls[j].NextTestDate = s.calculateNextTestDate(testResult.TestDate, "monthly")

					// Update control effectiveness based on test results
					s.updateControlEffectiveness(tracking, i, j)

					s.logAuditTrail(ctx, businessID, framework, &requirementID, AuditActionUpdate,
						fmt.Sprintf("Added test result for control: %s - %s", testResult.TestType, testResult.Result),
						"", "")

					s.logger.Info("Control test result added successfully",
						"request_id", requestID,
						"business_id", businessID,
						"control_id", controlID,
						"test_type", testResult.TestType,
						"result", testResult.Result,
					)

					return nil
				}
			}
		}
	}

	return fmt.Errorf("control %s not found in requirement %s", controlID, requirementID)
}

// AddEvidence adds evidence for a requirement or control
func (s *TrackingSystem) AddEvidence(ctx context.Context, businessID, framework, requirementID, controlID string, evidence TrackingEvidence) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Adding evidence",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"requirement_id", requirementID,
		"control_id", controlID,
		"evidence_type", evidence.Type,
		"title", evidence.Title,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	trackingKey := fmt.Sprintf("%s_%s", businessID, framework)
	tracking, exists := s.trackingData[trackingKey]
	if !exists {
		return fmt.Errorf("compliance tracking not found for business %s and framework %s", businessID, framework)
	}

	// Find and add evidence
	for i, req := range tracking.Requirements {
		if req.RequirementID == requirementID {
			if controlID == "" {
				// Add evidence to requirement
				tracking.Requirements[i].Evidence = append(tracking.Requirements[i].Evidence, evidence)
			} else {
				// Add evidence to specific control
				for j, control := range req.Controls {
					if control.ControlID == controlID {
						controlEvidence := ControlEvidence{
							ID:          evidence.ID,
							ControlID:   controlID,
							Type:        evidence.Type,
							Title:       evidence.Title,
							Description: evidence.Description,
							URL:         evidence.URL,
							UploadedAt:  evidence.UploadedAt,
							UploadedBy:  evidence.UploadedBy,
							ExpiresAt:   evidence.ExpiresAt,
						}
						tracking.Requirements[i].Controls[j].Evidence = append(tracking.Requirements[i].Controls[j].Evidence, controlEvidence)
						break
					}
				}
			}

			s.logAuditTrail(ctx, businessID, framework, &requirementID, AuditActionCreate,
				fmt.Sprintf("Added evidence: %s", evidence.Title),
				"", "")

			s.logger.Info("Evidence added successfully",
				"request_id", requestID,
				"business_id", businessID,
				"requirement_id", requirementID,
				"control_id", controlID,
				"evidence_type", evidence.Type,
				"title", evidence.Title,
			)

			return nil
		}
	}

	return fmt.Errorf("requirement %s not found in tracking", requirementID)
}

// CreateException creates a compliance exception
func (s *TrackingSystem) CreateException(ctx context.Context, businessID, framework, requirementID string, exception ComplianceException) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Creating compliance exception",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"requirement_id", requirementID,
		"exception_type", exception.Type,
		"reason", exception.Reason,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	trackingKey := fmt.Sprintf("%s_%s", businessID, framework)
	tracking, exists := s.trackingData[trackingKey]
	if !exists {
		return fmt.Errorf("compliance tracking not found for business %s and framework %s", businessID, framework)
	}

	// Find and add exception to the requirement
	for i, req := range tracking.Requirements {
		if req.RequirementID == requirementID {
			tracking.Requirements[i].Exceptions = append(tracking.Requirements[i].Exceptions, exception)

			s.logAuditTrail(ctx, businessID, framework, &requirementID, AuditActionException,
				fmt.Sprintf("Created exception: %s", exception.Reason),
				"", "")

			s.logger.Info("Compliance exception created successfully",
				"request_id", requestID,
				"business_id", businessID,
				"requirement_id", requirementID,
				"exception_type", exception.Type,
				"reason", exception.Reason,
			)

			return nil
		}
	}

	return fmt.Errorf("requirement %s not found in tracking", requirementID)
}

// CreateRemediationPlan creates a remediation plan for a requirement
func (s *TrackingSystem) CreateRemediationPlan(ctx context.Context, businessID, framework, requirementID string, plan RemediationPlan) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Creating remediation plan",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
		"requirement_id", requirementID,
		"plan_title", plan.Title,
		"priority", plan.Priority,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	trackingKey := fmt.Sprintf("%s_%s", businessID, framework)
	tracking, exists := s.trackingData[trackingKey]
	if !exists {
		return fmt.Errorf("compliance tracking not found for business %s and framework %s", businessID, framework)
	}

	// Find and add remediation plan to the requirement
	for i, req := range tracking.Requirements {
		if req.RequirementID == requirementID {
			tracking.Requirements[i].RemediationPlan = &plan

			s.logAuditTrail(ctx, businessID, framework, &requirementID, AuditActionRemediation,
				fmt.Sprintf("Created remediation plan: %s", plan.Title),
				"", "")

			s.logger.Info("Remediation plan created successfully",
				"request_id", requestID,
				"business_id", businessID,
				"requirement_id", requirementID,
				"plan_title", plan.Title,
				"priority", plan.Priority,
			)

			return nil
		}
	}

	return fmt.Errorf("requirement %s not found in tracking", requirementID)
}

// GetComplianceTracking gets compliance tracking for a business and framework
func (s *TrackingSystem) GetComplianceTracking(ctx context.Context, businessID, framework string) (*ComplianceTracking, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting compliance tracking",
		"request_id", requestID,
		"business_id", businessID,
		"framework", framework,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	trackingKey := fmt.Sprintf("%s_%s", businessID, framework)
	tracking, exists := s.trackingData[trackingKey]
	if !exists {
		return nil, fmt.Errorf("compliance tracking not found for business %s and framework %s", businessID, framework)
	}

	return tracking, nil
}

// GetBusinessComplianceSummary gets a summary of compliance for a business across all frameworks
func (s *TrackingSystem) GetBusinessComplianceSummary(ctx context.Context, businessID string) (map[string]*ComplianceTracking, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting business compliance summary",
		"request_id", requestID,
		"business_id", businessID,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	summary := make(map[string]*ComplianceTracking)
	for _, tracking := range s.trackingData {
		if tracking.BusinessID == businessID {
			framework := tracking.Framework
			summary[framework] = tracking
		}
	}

	return summary, nil
}

// GetComplianceAuditTrail gets the audit trail for compliance activities
func (s *TrackingSystem) GetComplianceAuditTrail(ctx context.Context, businessID string, startDate, endDate time.Time) ([]ComplianceAuditTrail, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting compliance audit trail",
		"request_id", requestID,
		"business_id", businessID,
		"start_date", startDate,
		"end_date", endDate,
	)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var filteredTrail []ComplianceAuditTrail
	for _, entry := range s.auditTrail {
		if entry.BusinessID == businessID &&
			entry.Timestamp.After(startDate) &&
			entry.Timestamp.Before(endDate) {
			filteredTrail = append(filteredTrail, entry)
		}
	}

	return filteredTrail, nil
}

// Helper methods
func (s *TrackingSystem) updateOverallComplianceScore(tracking *ComplianceTracking) {
	if len(tracking.Requirements) == 0 {
		tracking.ComplianceScore = 0.0
		return
	}

	totalScore := 0.0
	validRequirements := 0

	for _, req := range tracking.Requirements {
		if req.Status != ComplianceStatusExempt {
			totalScore += req.ComplianceScore
			validRequirements++
		}
	}

	if validRequirements > 0 {
		tracking.ComplianceScore = totalScore / float64(validRequirements)
	} else {
		tracking.ComplianceScore = 0.0
	}

	// Update overall status based on score
	if tracking.ComplianceScore >= 90.0 {
		tracking.OverallStatus = ComplianceStatusVerified
	} else if tracking.ComplianceScore >= 70.0 {
		tracking.OverallStatus = ComplianceStatusImplemented
	} else if tracking.ComplianceScore >= 30.0 {
		tracking.OverallStatus = ComplianceStatusInProgress
	} else {
		tracking.OverallStatus = ComplianceStatusNotStarted
	}
}

func (s *TrackingSystem) updateRequirementScore(tracking *ComplianceTracking, reqIndex int) {
	req := &tracking.Requirements[reqIndex]
	if len(req.Controls) == 0 {
		req.ComplianceScore = 0.0
		return
	}

	totalScore := 0.0
	validControls := 0

	for _, control := range req.Controls {
		// Calculate control score based on effectiveness
		var controlScore float64
		switch control.Effectiveness {
		case ControlEffectivenessHighlyEffective:
			controlScore = 100.0
		case ControlEffectivenessEffective:
			controlScore = 80.0
		case ControlEffectivenessPartiallyEffective:
			controlScore = 50.0
		case ControlEffectivenessIneffective:
			controlScore = 0.0
		}

		totalScore += controlScore
		validControls++
	}

	if validControls > 0 {
		req.ComplianceScore = totalScore / float64(validControls)
	} else {
		req.ComplianceScore = 0.0
	}

	// Update requirement status based on score
	if req.ComplianceScore >= 90.0 {
		req.Status = ComplianceStatusVerified
	} else if req.ComplianceScore >= 70.0 {
		req.Status = ComplianceStatusImplemented
	} else if req.ComplianceScore >= 30.0 {
		req.Status = ComplianceStatusInProgress
	} else {
		req.Status = ComplianceStatusNotStarted
	}
}

func (s *TrackingSystem) updateControlEffectiveness(tracking *ComplianceTracking, reqIndex, controlIndex int) {
	control := &tracking.Requirements[reqIndex].Controls[controlIndex]

	if len(control.TestResults) == 0 {
		control.Effectiveness = ControlEffectivenessIneffective
		return
	}

	// Calculate effectiveness based on recent test results
	recentTests := 0
	passCount := 0
	lastMonth := time.Now().Add(-30 * 24 * time.Hour)

	for _, test := range control.TestResults {
		if test.TestDate.After(lastMonth) {
			recentTests++
			if test.Result == TestResultPass {
				passCount++
			}
		}
	}

	if recentTests == 0 {
		control.Effectiveness = ControlEffectivenessIneffective
		return
	}

	passRate := float64(passCount) / float64(recentTests)

	if passRate >= 0.9 {
		control.Effectiveness = ControlEffectivenessHighlyEffective
	} else if passRate >= 0.7 {
		control.Effectiveness = ControlEffectivenessEffective
	} else if passRate >= 0.5 {
		control.Effectiveness = ControlEffectivenessPartiallyEffective
	} else {
		control.Effectiveness = ControlEffectivenessIneffective
	}
}

func (s *TrackingSystem) calculateNextTestDate(lastTestDate time.Time, frequency string) *time.Time {
	var nextDate time.Time

	switch frequency {
	case "daily":
		nextDate = lastTestDate.Add(24 * time.Hour)
	case "weekly":
		nextDate = lastTestDate.Add(7 * 24 * time.Hour)
	case "monthly":
		nextDate = lastTestDate.Add(30 * 24 * time.Hour)
	case "quarterly":
		nextDate = lastTestDate.Add(90 * 24 * time.Hour)
	case "annually":
		nextDate = lastTestDate.Add(365 * 24 * time.Hour)
	default:
		nextDate = lastTestDate.Add(30 * 24 * time.Hour) // Default to monthly
	}

	return &nextDate
}

func (s *TrackingSystem) logAuditTrail(ctx context.Context, businessID, framework string, requirementID *string, action AuditAction, description, oldValue, newValue string) {
	requestID := ctx.Value("request_id").(string)

	auditEntry := ComplianceAuditTrail{
		ID:            fmt.Sprintf("audit_%d", time.Now().UnixNano()),
		BusinessID:    businessID,
		Framework:     framework,
		RequirementID: requirementID,
		Action:        action,
		Description:   description,
		UserID:        "system",
		UserName:      "System",
		UserRole:      "compliance_system",
		Timestamp:     time.Now(),
		IPAddress:     "127.0.0.1",
		UserAgent:     "compliance-system",
		SessionID:     requestID,
		RequestID:     requestID,
		OldValue:      oldValue,
		NewValue:      newValue,
	}

	s.auditTrail = append(s.auditTrail, auditEntry)
}
