package compliance

import (
	"context"
	"testing"
	"time"

	"kyb-platform/internal/compliance"
	"kyb-platform/internal/observability"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestRegulatoryRequirementValidation tests the validation of regulatory requirements
func TestRegulatoryRequirementValidation(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create compliance framework service
	frameworkService := compliance.NewComplianceFrameworkService(logger)

	t.Run("Framework Validation", func(t *testing.T) {
		// Test SOC2 framework validation
		t.Log("Testing SOC2 framework validation...")
		soc2Framework, err := frameworkService.GetFramework(context.Background(), "SOC2")
		if err != nil {
			t.Fatalf("Failed to get SOC2 framework: %v", err)
		}

		// Validate SOC2 framework structure
		assert.Equal(t, "SOC2", soc2Framework.ID)
		assert.Equal(t, "SOC 2 Type II", soc2Framework.Name)
		assert.Equal(t, "security", soc2Framework.Category)
		assert.Equal(t, "active", soc2Framework.Status)
		assert.Equal(t, "AICPA", soc2Framework.Authority)
		assert.Contains(t, soc2Framework.Jurisdiction, "US")
		assert.Contains(t, soc2Framework.Jurisdiction, "Global")
		assert.Len(t, soc2Framework.Requirements, 2)
		t.Logf("✅ SOC2 framework validation passed: %s - %s", soc2Framework.ID, soc2Framework.Name)

		// Test GDPR framework validation
		t.Log("Testing GDPR framework validation...")
		gdprFramework, err := frameworkService.GetFramework(context.Background(), "GDPR")
		if err != nil {
			t.Fatalf("Failed to get GDPR framework: %v", err)
		}

		// Validate GDPR framework structure
		assert.Equal(t, "GDPR", gdprFramework.ID)
		assert.Equal(t, "General Data Protection Regulation", gdprFramework.Name)
		assert.Equal(t, "privacy", gdprFramework.Category)
		assert.Equal(t, "active", gdprFramework.Status)
		assert.Equal(t, "European Commission", gdprFramework.Authority)
		assert.Contains(t, gdprFramework.Jurisdiction, "EU")
		assert.Contains(t, gdprFramework.Jurisdiction, "EEA")
		assert.Contains(t, gdprFramework.Jurisdiction, "UK")
		assert.Len(t, gdprFramework.Requirements, 2)
		t.Logf("✅ GDPR framework validation passed: %s - %s", gdprFramework.ID, gdprFramework.Name)

		// Test PCI DSS framework validation
		t.Log("Testing PCI DSS framework validation...")
		pciFramework, err := frameworkService.GetFramework(context.Background(), "PCI_DSS")
		if err != nil {
			t.Fatalf("Failed to get PCI DSS framework: %v", err)
		}

		// Validate PCI DSS framework structure
		assert.Equal(t, "PCI_DSS", pciFramework.ID)
		assert.Equal(t, "Payment Card Industry Data Security Standard", pciFramework.Name)
		assert.Equal(t, "financial", pciFramework.Category)
		assert.Equal(t, "active", pciFramework.Status)
		assert.Equal(t, "PCI Security Standards Council", pciFramework.Authority)
		assert.Contains(t, pciFramework.Jurisdiction, "Global")
		assert.Contains(t, pciFramework.Scope, "financial")
		assert.Contains(t, pciFramework.Scope, "ecommerce")
		t.Logf("✅ PCI DSS framework validation passed: %s - %s", pciFramework.ID, pciFramework.Name)

		// Test HIPAA framework validation
		t.Log("Testing HIPAA framework validation...")
		hipaaFramework, err := frameworkService.GetFramework(context.Background(), "HIPAA")
		if err != nil {
			t.Fatalf("Failed to get HIPAA framework: %v", err)
		}

		// Validate HIPAA framework structure
		assert.Equal(t, "HIPAA", hipaaFramework.ID)
		assert.Equal(t, "Health Insurance Portability and Accountability Act", hipaaFramework.Name)
		assert.Equal(t, "privacy", hipaaFramework.Category)
		assert.Equal(t, "active", hipaaFramework.Status)
		assert.Equal(t, "HHS", hipaaFramework.Authority)
		assert.Contains(t, hipaaFramework.Jurisdiction, "US")
		assert.Contains(t, hipaaFramework.Scope, "healthcare")
		assert.Contains(t, hipaaFramework.Scope, "health_tech")
		t.Logf("✅ HIPAA framework validation passed: %s - %s", hipaaFramework.ID, hipaaFramework.Name)
	})

	t.Run("Requirement Validation", func(t *testing.T) {
		// Test SOC2 requirements validation
		t.Log("Testing SOC2 requirements validation...")

		soc2Requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), "SOC2")
		if err != nil {
			t.Fatalf("Failed to get SOC2 requirements: %v", err)
		}

		// Validate SOC2 requirements
		assert.Len(t, soc2Requirements, 2)

		// Find and validate CC6.1 requirement
		var cc6_1 *compliance.ComplianceRequirement
		for _, req := range soc2Requirements {
			if req.Code == "CC6.1" {
				cc6_1 = req
				break
			}
		}
		assert.NotNil(t, cc6_1, "CC6.1 requirement not found")
		assert.Equal(t, "SOC2_CC6_1", cc6_1.ID)
		assert.Equal(t, "SOC2", cc6_1.FrameworkID)
		assert.Equal(t, "Logical and Physical Access Controls", cc6_1.Name)
		assert.Equal(t, "access_control", cc6_1.Category)
		assert.Equal(t, "critical", cc6_1.Priority)
		assert.Equal(t, "technical", cc6_1.Type)
		assert.Equal(t, "hybrid", cc6_1.AssessmentMethod)
		assert.Equal(t, "continuous", cc6_1.Frequency)
		assert.Equal(t, "security_team", cc6_1.Owner)
		t.Logf("✅ SOC2 CC6.1 requirement validation passed: %s - %s", cc6_1.Code, cc6_1.Name)

		// Find and validate CC6.2 requirement
		var cc6_2 *compliance.ComplianceRequirement
		for _, req := range soc2Requirements {
			if req.Code == "CC6.2" {
				cc6_2 = req
				break
			}
		}
		assert.NotNil(t, cc6_2, "CC6.2 requirement not found")
		assert.Equal(t, "SOC2_CC6_2", cc6_2.ID)
		assert.Equal(t, "SOC2", cc6_2.FrameworkID)
		assert.Equal(t, "System Access", cc6_2.Name)
		assert.Equal(t, "access_control", cc6_2.Category)
		assert.Equal(t, "critical", cc6_2.Priority)
		assert.Equal(t, "administrative", cc6_2.Type)
		assert.Equal(t, "manual", cc6_2.AssessmentMethod)
		assert.Equal(t, "monthly", cc6_2.Frequency)
		assert.Equal(t, "security_team", cc6_2.Owner)
		t.Logf("✅ SOC2 CC6.2 requirement validation passed: %s - %s", cc6_2.Code, cc6_2.Name)

		// Test GDPR requirements validation
		t.Log("Testing GDPR requirements validation...")

		gdprRequirements, err := frameworkService.GetFrameworkRequirements(context.Background(), "GDPR")
		if err != nil {
			t.Fatalf("Failed to get GDPR requirements: %v", err)
		}

		// Validate GDPR requirements
		assert.Len(t, gdprRequirements, 2)

		// Find and validate GDPR_32 requirement
		var gdpr32 *compliance.ComplianceRequirement
		for _, req := range gdprRequirements {
			if req.Code == "GDPR_32" {
				gdpr32 = req
				break
			}
		}
		assert.NotNil(t, gdpr32, "GDPR_32 requirement not found")
		assert.Equal(t, "GDPR_32", gdpr32.ID)
		assert.Equal(t, "GDPR", gdpr32.FrameworkID)
		assert.Equal(t, "Security of Processing", gdpr32.Name)
		assert.Equal(t, "data_protection", gdpr32.Category)
		assert.Equal(t, "critical", gdpr32.Priority)
		assert.Equal(t, "technical", gdpr32.Type)
		assert.Equal(t, "hybrid", gdpr32.AssessmentMethod)
		assert.Equal(t, "continuous", gdpr32.Frequency)
		assert.Equal(t, "privacy_team", gdpr32.Owner)
		t.Logf("✅ GDPR_32 requirement validation passed: %s - %s", gdpr32.Code, gdpr32.Name)

		// Find and validate GDPR_25 requirement
		var gdpr25 *compliance.ComplianceRequirement
		for _, req := range gdprRequirements {
			if req.Code == "GDPR_25" {
				gdpr25 = req
				break
			}
		}
		assert.NotNil(t, gdpr25, "GDPR_25 requirement not found")
		assert.Equal(t, "GDPR_25", gdpr25.ID)
		assert.Equal(t, "GDPR", gdpr25.FrameworkID)
		assert.Equal(t, "Data Protection by Design and by Default", gdpr25.Name)
		assert.Equal(t, "data_protection", gdpr25.Category)
		assert.Equal(t, "high", gdpr25.Priority)
		assert.Equal(t, "technical", gdpr25.Type)
		assert.Equal(t, "hybrid", gdpr25.AssessmentMethod)
		assert.Equal(t, "quarterly", gdpr25.Frequency)
		assert.Equal(t, "privacy_team", gdpr25.Owner)
		t.Logf("✅ GDPR_25 requirement validation passed: %s - %s", gdpr25.Code, gdpr25.Name)
	})

	t.Run("Framework-Requirement Relationship Validation", func(t *testing.T) {
		// Test that frameworks have correct requirements
		t.Log("Testing framework-requirement relationship validation...")

		// Test SOC2 framework requirements
		soc2Framework, err := frameworkService.GetFramework(context.Background(), "SOC2")
		if err != nil {
			t.Fatalf("Failed to get SOC2 framework: %v", err)
		}

		// Validate SOC2 requirements
		assert.Len(t, soc2Framework.Requirements, 2)
		assert.Contains(t, soc2Framework.Requirements, "SOC2_CC6_1")
		assert.Contains(t, soc2Framework.Requirements, "SOC2_CC6_2")

		// Validate each requirement belongs to SOC2
		soc2Requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), "SOC2")
		if err != nil {
			t.Fatalf("Failed to get SOC2 requirements: %v", err)
		}
		for _, req := range soc2Requirements {
			assert.Equal(t, "SOC2", req.FrameworkID)
		}
		t.Logf("✅ SOC2 framework-requirement relationship validation passed: %d requirements", len(soc2Framework.Requirements))

		// Test GDPR framework requirements
		gdprFramework, err := frameworkService.GetFramework(context.Background(), "GDPR")
		if err != nil {
			t.Fatalf("Failed to get GDPR framework: %v", err)
		}

		// Validate GDPR requirements
		assert.Len(t, gdprFramework.Requirements, 2)
		assert.Contains(t, gdprFramework.Requirements, "GDPR_32")
		assert.Contains(t, gdprFramework.Requirements, "GDPR_25")

		// Validate each requirement belongs to GDPR
		gdprRequirements, err := frameworkService.GetFrameworkRequirements(context.Background(), "GDPR")
		if err != nil {
			t.Fatalf("Failed to get GDPR requirements: %v", err)
		}
		for _, req := range gdprRequirements {
			assert.Equal(t, "GDPR", req.FrameworkID)
		}
		t.Logf("✅ GDPR framework-requirement relationship validation passed: %d requirements", len(gdprFramework.Requirements))
	})
}

// TestRegulatoryRequirementTracking tests the tracking of regulatory requirements
func TestRegulatoryRequirementTracking(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create compliance framework service
	frameworkService := compliance.NewComplianceFrameworkService(logger)

	// Create compliance tracking service
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)

	t.Run("Requirement Progress Tracking", func(t *testing.T) {
		// Test tracking requirement progress
		t.Log("Testing requirement progress tracking...")

		businessID := "test-business-tracking"
		frameworkID := "SOC2"

		// Create initial tracking
		tracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "SOC2_CC6_1",
					Progress:      0.0,
					Status:        "not_started",
					LastAssessed:  time.Now(),
				},
				{
					RequirementID: "SOC2_CC6_2",
					Progress:      0.0,
					Status:        "not_started",
					LastAssessed:  time.Now(),
				},
			},
		}

		// Update tracking
		err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
		if err != nil {
			t.Fatalf("Failed to update compliance tracking: %v", err)
		}

		// Retrieve tracking
		retrievedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get compliance tracking: %v", err)
		}

		// Validate tracking
		assert.Equal(t, businessID, retrievedTracking.BusinessID)
		assert.Equal(t, frameworkID, retrievedTracking.FrameworkID)
		assert.Len(t, retrievedTracking.Requirements, 2)
		assert.Equal(t, 0.0, retrievedTracking.OverallProgress)
		assert.Equal(t, "non_compliant", retrievedTracking.ComplianceLevel)
		assert.Equal(t, "critical", retrievedTracking.RiskLevel)
		t.Logf("✅ Requirement progress tracking validation passed: %s - %s", businessID, frameworkID)

		// Update progress for one requirement
		tracking.Requirements[0].Progress = 0.5
		tracking.Requirements[0].Status = "in_progress"

		err = trackingService.UpdateComplianceTracking(context.Background(), tracking)
		if err != nil {
			t.Fatalf("Failed to update compliance tracking: %v", err)
		}

		// Retrieve updated tracking
		updatedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get updated compliance tracking: %v", err)
		}

		// Validate updated tracking
		assert.Equal(t, 0.25, updatedTracking.OverallProgress) // (0.5 + 0.0) / 2
		assert.Equal(t, "non_compliant", updatedTracking.ComplianceLevel)
		assert.Equal(t, "high", updatedTracking.RiskLevel)
		t.Logf("✅ Updated requirement progress tracking validation passed: progress=%f", updatedTracking.OverallProgress)
	})

	t.Run("Requirement Status Tracking", func(t *testing.T) {
		// Test tracking requirement status changes
		t.Log("Testing requirement status tracking...")

		businessID := "test-business-status"
		frameworkID := "GDPR"

		// Create tracking with different statuses
		tracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "GDPR_32",
					Progress:      1.0,
					Status:        "completed",
					LastAssessed:  time.Now(),
				},
				{
					RequirementID: "GDPR_25",
					Progress:      0.8,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
			},
		}

		// Update tracking
		err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
		if err != nil {
			t.Fatalf("Failed to update compliance tracking: %v", err)
		}

		// Retrieve tracking
		retrievedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get compliance tracking: %v", err)
		}

		// Validate tracking
		assert.Equal(t, businessID, retrievedTracking.BusinessID)
		assert.Equal(t, frameworkID, retrievedTracking.FrameworkID)
		assert.Len(t, retrievedTracking.Requirements, 2)
		assert.Equal(t, 0.9, retrievedTracking.OverallProgress) // (1.0 + 0.8) / 2
		assert.Equal(t, "compliant", retrievedTracking.ComplianceLevel)
		assert.Equal(t, "low", retrievedTracking.RiskLevel)

		// Validate individual requirement statuses
		assert.Equal(t, "completed", retrievedTracking.Requirements[0].Status)
		assert.Equal(t, "in_progress", retrievedTracking.Requirements[1].Status)
		t.Logf("✅ Requirement status tracking validation passed: %s - %s", businessID, frameworkID)
	})

	t.Run("Requirement Due Date Tracking", func(t *testing.T) {
		// Test tracking requirement due dates
		t.Log("Testing requirement due date tracking...")

		businessID := "test-business-due-dates"
		frameworkID := "SOC2"

		// Create tracking with due dates
		now := time.Now()
		dueDate1 := now.AddDate(0, 0, 30) // 30 days from now
		dueDate2 := now.AddDate(0, 0, 60) // 60 days from now

		tracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "SOC2_CC6_1",
					Progress:      0.3,
					Status:        "in_progress",
					NextDueDate:   &dueDate1,
					LastAssessed:  now,
				},
				{
					RequirementID: "SOC2_CC6_2",
					Progress:      0.7,
					Status:        "in_progress",
					NextDueDate:   &dueDate2,
					LastAssessed:  now,
				},
			},
		}

		// Update tracking
		err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
		if err != nil {
			t.Fatalf("Failed to update compliance tracking: %v", err)
		}

		// Retrieve tracking
		retrievedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
		if err != nil {
			t.Fatalf("Failed to get compliance tracking: %v", err)
		}

		// Validate tracking
		assert.Equal(t, businessID, retrievedTracking.BusinessID)
		assert.Equal(t, frameworkID, retrievedTracking.FrameworkID)
		assert.Len(t, retrievedTracking.Requirements, 2)
		assert.Equal(t, 0.5, retrievedTracking.OverallProgress) // (0.3 + 0.7) / 2
		assert.Equal(t, "partial", retrievedTracking.ComplianceLevel)
		assert.Equal(t, "medium", retrievedTracking.RiskLevel)

		// Validate due dates
		assert.NotNil(t, retrievedTracking.Requirements[0].NextDueDate)
		assert.NotNil(t, retrievedTracking.Requirements[1].NextDueDate)
		assert.True(t, retrievedTracking.Requirements[0].NextDueDate.Before(*retrievedTracking.Requirements[1].NextDueDate))
		t.Logf("✅ Requirement due date tracking validation passed: %s - %s", businessID, frameworkID)
	})
}

// TestRegulatoryRequirementIntegration tests the integration of regulatory requirements
func TestRegulatoryRequirementIntegration(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create compliance framework service
	frameworkService := compliance.NewComplianceFrameworkService(logger)

	// Create compliance tracking service
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)

	t.Run("Multi-Framework Requirement Integration", func(t *testing.T) {
		// Test integration across multiple frameworks
		t.Log("Testing multi-framework requirement integration...")

		businessID := "test-business-multi-framework"

		// Test SOC2 framework
		soc2Tracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: "SOC2",
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "SOC2_CC6_1",
					Progress:      0.8,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
				{
					RequirementID: "SOC2_CC6_2",
					Progress:      0.6,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
			},
		}

		err := trackingService.UpdateComplianceTracking(context.Background(), soc2Tracking)
		if err != nil {
			t.Fatalf("Failed to update SOC2 compliance tracking: %v", err)
		}

		// Test GDPR framework
		gdprTracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: "GDPR",
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "GDPR_32",
					Progress:      0.9,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
				{
					RequirementID: "GDPR_25",
					Progress:      0.7,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
			},
		}

		err = trackingService.UpdateComplianceTracking(context.Background(), gdprTracking)
		if err != nil {
			t.Fatalf("Failed to update GDPR compliance tracking: %v", err)
		}

		// Retrieve and validate SOC2 tracking
		retrievedSOC2, err := trackingService.GetComplianceTracking(context.Background(), businessID, "SOC2")
		if err != nil {
			t.Fatalf("Failed to get SOC2 compliance tracking: %v", err)
		}

		assert.Equal(t, businessID, retrievedSOC2.BusinessID)
		assert.Equal(t, "SOC2", retrievedSOC2.FrameworkID)
		assert.Equal(t, 0.7, retrievedSOC2.OverallProgress) // (0.8 + 0.6) / 2
		assert.Equal(t, "partial", retrievedSOC2.ComplianceLevel)
		assert.Equal(t, "medium", retrievedSOC2.RiskLevel)

		// Retrieve and validate GDPR tracking
		retrievedGDPR, err := trackingService.GetComplianceTracking(context.Background(), businessID, "GDPR")
		if err != nil {
			t.Fatalf("Failed to get GDPR compliance tracking: %v", err)
		}

		assert.Equal(t, businessID, retrievedGDPR.BusinessID)
		assert.Equal(t, "GDPR", retrievedGDPR.FrameworkID)
		assert.Equal(t, 0.8, retrievedGDPR.OverallProgress) // (0.9 + 0.7) / 2
		assert.Equal(t, "partial", retrievedGDPR.ComplianceLevel)
		assert.Equal(t, "low", retrievedGDPR.RiskLevel)

		t.Logf("✅ Multi-framework requirement integration validation passed: SOC2=%f, GDPR=%f",
			retrievedSOC2.OverallProgress, retrievedGDPR.OverallProgress)
	})

	t.Run("Requirement Cross-Reference Validation", func(t *testing.T) {
		// Test cross-reference validation between frameworks and requirements
		t.Log("Testing requirement cross-reference validation...")

		// Test that all framework requirements exist
		frameworks := []string{"SOC2", "GDPR", "PCI_DSS", "HIPAA"}
		for _, frameworkID := range frameworks {
			// Get framework requirements
			requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), frameworkID)
			if err != nil {
				t.Fatalf("Failed to get requirements for framework %s: %v", frameworkID, err)
			}

			// Validate each requirement belongs to the framework
			for _, req := range requirements {
				assert.Equal(t, frameworkID, req.FrameworkID)
			}
		}
		t.Logf("✅ Requirement cross-reference validation passed: %d frameworks", len(frameworks))
	})

	t.Run("Requirement Consistency Validation", func(t *testing.T) {
		// Test consistency validation across requirements
		t.Log("Testing requirement consistency validation...")

		// Test that all requirements have valid priorities
		validPriorities := []string{"critical", "high", "medium", "low"}
		frameworks := []string{"SOC2", "GDPR"}

		for _, frameworkID := range frameworks {
			requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), frameworkID)
			if err != nil {
				t.Fatalf("Failed to get requirements for framework %s: %v", frameworkID, err)
			}

			for _, req := range requirements {
				assert.Contains(t, validPriorities, req.Priority)
			}
		}
		t.Logf("✅ Requirement priority consistency validation passed: %d frameworks", len(frameworks))

		// Test that all requirements have valid types
		validTypes := []string{"technical", "administrative", "physical"}
		for _, frameworkID := range frameworks {
			requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), frameworkID)
			if err != nil {
				t.Fatalf("Failed to get requirements for framework %s: %v", frameworkID, err)
			}

			for _, req := range requirements {
				assert.Contains(t, validTypes, req.Type)
			}
		}
		t.Logf("✅ Requirement type consistency validation passed: %d frameworks", len(frameworks))

		// Test that all requirements have valid assessment methods
		validAssessmentMethods := []string{"automated", "manual", "hybrid"}
		for _, frameworkID := range frameworks {
			requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), frameworkID)
			if err != nil {
				t.Fatalf("Failed to get requirements for framework %s: %v", frameworkID, err)
			}

			for _, req := range requirements {
				assert.Contains(t, validAssessmentMethods, req.AssessmentMethod)
			}
		}
		t.Logf("✅ Requirement assessment method consistency validation passed: %d frameworks", len(frameworks))

		// Test that all requirements have valid frequencies
		validFrequencies := []string{"continuous", "monthly", "quarterly", "annually"}
		for _, frameworkID := range frameworks {
			requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), frameworkID)
			if err != nil {
				t.Fatalf("Failed to get requirements for framework %s: %v", frameworkID, err)
			}

			for _, req := range requirements {
				assert.Contains(t, validFrequencies, req.Frequency)
			}
		}
		t.Logf("✅ Requirement frequency consistency validation passed: %d frameworks", len(frameworks))
	})
}
