package compliance

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestRegulatoryAccuracy tests regulatory accuracy across all compliance frameworks
func TestRegulatoryAccuracy(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create compliance services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)

	t.Run("Framework Accuracy Validation", func(t *testing.T) {
		// Test framework accuracy validation
		t.Log("Testing framework accuracy validation...")

		// Test SOC2 framework accuracy
		soc2Framework, err := frameworkService.GetFramework(context.Background(), "SOC2")
		assert.NoError(t, err, "SOC2 framework should be accessible")
		assert.Equal(t, "SOC2", soc2Framework.ID, "SOC2 framework ID should be correct")
		assert.Equal(t, "SOC 2 Type II", soc2Framework.Name, "SOC2 framework name should be correct")
		assert.Equal(t, "security", soc2Framework.Category, "SOC2 framework category should be security")
		assert.Equal(t, "active", soc2Framework.Status, "SOC2 framework status should be active")
		assert.Equal(t, "AICPA", soc2Framework.Authority, "SOC2 framework authority should be AICPA")
		assert.Contains(t, soc2Framework.Jurisdiction, "US", "SOC2 should apply to US jurisdiction")
		assert.Contains(t, soc2Framework.Jurisdiction, "Global", "SOC2 should apply globally")

		// Test GDPR framework accuracy
		gdprFramework, err := frameworkService.GetFramework(context.Background(), "GDPR")
		assert.NoError(t, err, "GDPR framework should be accessible")
		assert.Equal(t, "GDPR", gdprFramework.ID, "GDPR framework ID should be correct")
		assert.Equal(t, "General Data Protection Regulation", gdprFramework.Name, "GDPR framework name should be correct")
		assert.Equal(t, "privacy", gdprFramework.Category, "GDPR framework category should be privacy")
		assert.Equal(t, "active", gdprFramework.Status, "GDPR framework status should be active")
		assert.Equal(t, "European Commission", gdprFramework.Authority, "GDPR framework authority should be European Commission")
		assert.Contains(t, gdprFramework.Jurisdiction, "EU", "GDPR should apply to EU jurisdiction")
		assert.Contains(t, gdprFramework.Jurisdiction, "EEA", "GDPR should apply to EEA jurisdiction")

		// Test PCI DSS framework accuracy
		pciFramework, err := frameworkService.GetFramework(context.Background(), "PCI_DSS")
		assert.NoError(t, err, "PCI DSS framework should be accessible")
		assert.Equal(t, "PCI_DSS", pciFramework.ID, "PCI DSS framework ID should be correct")
		assert.Equal(t, "Payment Card Industry Data Security Standard", pciFramework.Name, "PCI DSS framework name should be correct")
		assert.Equal(t, "financial", pciFramework.Category, "PCI DSS framework category should be financial")
		assert.Equal(t, "active", pciFramework.Status, "PCI DSS framework status should be active")
		assert.Equal(t, "PCI Security Standards Council", pciFramework.Authority, "PCI DSS framework authority should be PCI Security Standards Council")
		assert.Contains(t, pciFramework.Jurisdiction, "Global", "PCI DSS should apply globally")

		// Test HIPAA framework accuracy
		hipaaFramework, err := frameworkService.GetFramework(context.Background(), "HIPAA")
		assert.NoError(t, err, "HIPAA framework should be accessible")
		assert.Equal(t, "HIPAA", hipaaFramework.ID, "HIPAA framework ID should be correct")
		assert.Equal(t, "Health Insurance Portability and Accountability Act", hipaaFramework.Name, "HIPAA framework name should be correct")
		assert.Equal(t, "privacy", hipaaFramework.Category, "HIPAA framework category should be privacy")
		assert.Equal(t, "active", hipaaFramework.Status, "HIPAA framework status should be active")
		assert.Equal(t, "HHS", hipaaFramework.Authority, "HIPAA framework authority should be HHS")
		assert.Contains(t, hipaaFramework.Jurisdiction, "US", "HIPAA should apply to US jurisdiction")

		t.Logf("✅ Framework accuracy validation: All 4 frameworks validated with 100%% accuracy")
	})

	t.Run("Requirement Accuracy Validation", func(t *testing.T) {
		// Test requirement accuracy validation
		t.Log("Testing requirement accuracy validation...")

		// Test SOC2 requirements accuracy
		soc2Requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), "SOC2")
		assert.NoError(t, err, "SOC2 requirements should be accessible")
		assert.Len(t, soc2Requirements, 2, "SOC2 should have 2 requirements")

		// Validate SOC2_CC6_1 requirement
		soc2CC61 := soc2Requirements[0]
		assert.Equal(t, "SOC2_CC6_1", soc2CC61.ID, "SOC2_CC6_1 requirement ID should be correct")
		assert.Equal(t, "CC6.1", soc2CC61.Code, "SOC2_CC6_1 requirement code should be CC6.1")
		assert.Equal(t, "Logical and Physical Access Controls", soc2CC61.Name, "SOC2_CC6_1 requirement name should be correct")
		assert.Equal(t, "access_control", soc2CC61.Category, "SOC2_CC6_1 requirement category should be access_control")
		assert.Equal(t, "critical", soc2CC61.Priority, "SOC2_CC6_1 requirement priority should be critical")
		assert.Equal(t, "technical", soc2CC61.Type, "SOC2_CC6_1 requirement type should be technical")
		assert.Equal(t, "continuous", soc2CC61.Frequency, "SOC2_CC6_1 requirement frequency should be continuous")
		assert.Equal(t, "security_team", soc2CC61.Owner, "SOC2_CC6_1 requirement owner should be security_team")

		// Validate SOC2_CC6_2 requirement
		soc2CC62 := soc2Requirements[1]
		assert.Equal(t, "SOC2_CC6_2", soc2CC62.ID, "SOC2_CC6_2 requirement ID should be correct")
		assert.Equal(t, "CC6.2", soc2CC62.Code, "SOC2_CC6_2 requirement code should be CC6.2")
		assert.Equal(t, "System Access", soc2CC62.Name, "SOC2_CC6_2 requirement name should be correct")
		assert.Equal(t, "access_control", soc2CC62.Category, "SOC2_CC6_2 requirement category should be access_control")
		assert.Equal(t, "critical", soc2CC62.Priority, "SOC2_CC6_2 requirement priority should be critical")
		assert.Equal(t, "administrative", soc2CC62.Type, "SOC2_CC6_2 requirement type should be administrative")
		assert.Equal(t, "monthly", soc2CC62.Frequency, "SOC2_CC6_2 requirement frequency should be monthly")
		assert.Equal(t, "security_team", soc2CC62.Owner, "SOC2_CC6_2 requirement owner should be security_team")

		// Test GDPR requirements accuracy
		gdprRequirements, err := frameworkService.GetFrameworkRequirements(context.Background(), "GDPR")
		assert.NoError(t, err, "GDPR requirements should be accessible")
		assert.Len(t, gdprRequirements, 2, "GDPR should have 2 requirements")

		// Find GDPR_25 requirement (sorted by code, so GDPR_25 comes first)
		var gdpr25, gdpr32 *compliance.ComplianceRequirement
		for _, req := range gdprRequirements {
			if req.ID == "GDPR_25" {
				gdpr25 = req
			} else if req.ID == "GDPR_32" {
				gdpr32 = req
			}
		}

		// Validate GDPR_25 requirement
		assert.NotNil(t, gdpr25, "GDPR_25 requirement should exist")
		assert.Equal(t, "GDPR_25", gdpr25.ID, "GDPR_25 requirement ID should be correct")
		assert.Equal(t, "GDPR_25", gdpr25.Code, "GDPR_25 requirement code should be GDPR_25")
		assert.Equal(t, "Data Protection by Design and by Default", gdpr25.Name, "GDPR_25 requirement name should be correct")
		assert.Equal(t, "data_protection", gdpr25.Category, "GDPR_25 requirement category should be data_protection")
		assert.Equal(t, "high", gdpr25.Priority, "GDPR_25 requirement priority should be high")
		assert.Equal(t, "technical", gdpr25.Type, "GDPR_25 requirement type should be technical")
		assert.Equal(t, "quarterly", gdpr25.Frequency, "GDPR_25 requirement frequency should be quarterly")
		assert.Equal(t, "privacy_team", gdpr25.Owner, "GDPR_25 requirement owner should be privacy_team")

		// Validate GDPR_32 requirement
		assert.NotNil(t, gdpr32, "GDPR_32 requirement should exist")
		assert.Equal(t, "GDPR_32", gdpr32.ID, "GDPR_32 requirement ID should be correct")
		assert.Equal(t, "GDPR_32", gdpr32.Code, "GDPR_32 requirement code should be GDPR_32")
		assert.Equal(t, "Security of Processing", gdpr32.Name, "GDPR_32 requirement name should be correct")
		assert.Equal(t, "data_protection", gdpr32.Category, "GDPR_32 requirement category should be data_protection")
		assert.Equal(t, "critical", gdpr32.Priority, "GDPR_32 requirement priority should be critical")
		assert.Equal(t, "technical", gdpr32.Type, "GDPR_32 requirement type should be technical")
		assert.Equal(t, "continuous", gdpr32.Frequency, "GDPR_32 requirement frequency should be continuous")
		assert.Equal(t, "privacy_team", gdpr32.Owner, "GDPR_32 requirement owner should be privacy_team")

		t.Logf("✅ Requirement accuracy validation: All 4 requirements validated with 100%% accuracy")
	})

	t.Run("Compliance Calculation Accuracy", func(t *testing.T) {
		// Test compliance calculation accuracy
		t.Log("Testing compliance calculation accuracy...")

		businessID := "test-business-accuracy"
		frameworkID := "SOC2"

		// Test case 1: 50% compliance (1 of 2 requirements at 100%)
		tracking1 := &compliance.ComplianceTracking{
			BusinessID:  businessID + "-1",
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "SOC2_CC6_1",
					Progress:      1.0, // 100% complete
					Status:        "completed",
					LastAssessed:  time.Now(),
				},
				{
					RequirementID: "SOC2_CC6_2",
					Progress:      0.0, // 0% complete
					Status:        "not_started",
					LastAssessed:  time.Now(),
				},
			},
		}

		err := trackingService.UpdateComplianceTracking(context.Background(), tracking1)
		assert.NoError(t, err, "Tracking update should work")

		retrievedTracking1, err := trackingService.GetComplianceTracking(context.Background(), businessID+"-1", frameworkID)
		assert.NoError(t, err, "Tracking retrieval should work")
		assert.Equal(t, 0.5, retrievedTracking1.OverallProgress, "Overall progress should be 0.5 (50%)")
		assert.Equal(t, "partial", retrievedTracking1.ComplianceLevel, "Compliance level should be partial")
		assert.Equal(t, "medium", retrievedTracking1.RiskLevel, "Risk level should be medium")

		// Test case 2: 100% compliance (both requirements at 100%)
		tracking2 := &compliance.ComplianceTracking{
			BusinessID:  businessID + "-2",
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "SOC2_CC6_1",
					Progress:      1.0, // 100% complete
					Status:        "completed",
					LastAssessed:  time.Now(),
				},
				{
					RequirementID: "SOC2_CC6_2",
					Progress:      1.0, // 100% complete
					Status:        "completed",
					LastAssessed:  time.Now(),
				},
			},
		}

		err = trackingService.UpdateComplianceTracking(context.Background(), tracking2)
		assert.NoError(t, err, "Tracking update should work")

		retrievedTracking2, err := trackingService.GetComplianceTracking(context.Background(), businessID+"-2", frameworkID)
		assert.NoError(t, err, "Tracking retrieval should work")
		assert.Equal(t, 1.0, retrievedTracking2.OverallProgress, "Overall progress should be 1.0 (100%)")
		assert.Equal(t, "compliant", retrievedTracking2.ComplianceLevel, "Compliance level should be compliant")
		assert.Equal(t, "low", retrievedTracking2.RiskLevel, "Risk level should be low")

		// Test case 3: 0% compliance (both requirements at 0%)
		tracking3 := &compliance.ComplianceTracking{
			BusinessID:  businessID + "-3",
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "SOC2_CC6_1",
					Progress:      0.0, // 0% complete
					Status:        "not_started",
					LastAssessed:  time.Now(),
				},
				{
					RequirementID: "SOC2_CC6_2",
					Progress:      0.0, // 0% complete
					Status:        "not_started",
					LastAssessed:  time.Now(),
				},
			},
		}

		err = trackingService.UpdateComplianceTracking(context.Background(), tracking3)
		assert.NoError(t, err, "Tracking update should work")

		retrievedTracking3, err := trackingService.GetComplianceTracking(context.Background(), businessID+"-3", frameworkID)
		assert.NoError(t, err, "Tracking retrieval should work")
		assert.Equal(t, 0.0, retrievedTracking3.OverallProgress, "Overall progress should be 0.0 (0%)")
		assert.Equal(t, "non_compliant", retrievedTracking3.ComplianceLevel, "Compliance level should be non_compliant")
		assert.Equal(t, "critical", retrievedTracking3.RiskLevel, "Risk level should be critical")

		t.Logf("✅ Compliance calculation accuracy: All 3 test cases validated with 100%% accuracy")
	})

	t.Run("Multi-Framework Accuracy Validation", func(t *testing.T) {
		// Test multi-framework accuracy validation
		t.Log("Testing multi-framework accuracy validation...")

		businessID := "test-business-multi-accuracy"
		frameworks := []string{"SOC2", "GDPR"}

		for _, frameworkID := range frameworks {
			// Create tracking for each framework
			tracking := &compliance.ComplianceTracking{
				BusinessID:  businessID,
				FrameworkID: frameworkID,
				Requirements: []compliance.RequirementTracking{
					{
						RequirementID: frameworkID + "_REQ_1",
						Progress:      0.8,
						Status:        "in_progress",
						LastAssessed:  time.Now(),
					},
				},
			}

			// Update tracking
			err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
			assert.NoError(t, err, "Multi-framework tracking should work for %s", frameworkID)

			// Retrieve tracking
			retrievedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
			assert.NoError(t, err, "Multi-framework retrieval should work for %s", frameworkID)
			assert.Equal(t, businessID, retrievedTracking.BusinessID, "Business ID should match for %s", frameworkID)
			assert.Equal(t, frameworkID, retrievedTracking.FrameworkID, "Framework ID should match for %s", frameworkID)
			assert.Equal(t, 0.8, retrievedTracking.OverallProgress, "Overall progress should be 0.8 for %s", frameworkID)
			assert.Equal(t, "partial", retrievedTracking.ComplianceLevel, "Compliance level should be partial for %s", frameworkID)
		}

		t.Logf("✅ Multi-framework accuracy validation: %d frameworks validated with 100%% accuracy", len(frameworks))
	})

	t.Run("Regulatory Mapping Accuracy", func(t *testing.T) {
		// Test regulatory mapping accuracy
		t.Log("Testing regulatory mapping accuracy...")

		// Test framework-to-requirement mapping accuracy
		soc2Framework, err := frameworkService.GetFramework(context.Background(), "SOC2")
		assert.NoError(t, err, "SOC2 framework should be accessible")
		assert.Len(t, soc2Framework.Requirements, 2, "SOC2 should have 2 requirements")
		assert.Contains(t, soc2Framework.Requirements, "SOC2_CC6_1", "SOC2 should include CC6.1 requirement")
		assert.Contains(t, soc2Framework.Requirements, "SOC2_CC6_2", "SOC2 should include CC6.2 requirement")

		gdprFramework, err := frameworkService.GetFramework(context.Background(), "GDPR")
		assert.NoError(t, err, "GDPR framework should be accessible")
		assert.Len(t, gdprFramework.Requirements, 2, "GDPR should have 2 requirements")
		assert.Contains(t, gdprFramework.Requirements, "GDPR_32", "GDPR should include Article 32 requirement")
		assert.Contains(t, gdprFramework.Requirements, "GDPR_25", "GDPR should include Article 25 requirement")

		// Test requirement-to-framework mapping accuracy
		soc2Requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), "SOC2")
		assert.NoError(t, err, "SOC2 requirements should be accessible")
		for _, req := range soc2Requirements {
			assert.Equal(t, "SOC2", req.FrameworkID, "SOC2 requirement should have SOC2 framework ID")
		}

		gdprRequirements, err := frameworkService.GetFrameworkRequirements(context.Background(), "GDPR")
		assert.NoError(t, err, "GDPR requirements should be accessible")
		for _, req := range gdprRequirements {
			assert.Equal(t, "GDPR", req.FrameworkID, "GDPR requirement should have GDPR framework ID")
		}

		t.Logf("✅ Regulatory mapping accuracy: All framework-requirement mappings validated with 100%% accuracy")
	})

	t.Run("Jurisdiction and Scope Accuracy", func(t *testing.T) {
		// Test jurisdiction and scope accuracy
		t.Log("Testing jurisdiction and scope accuracy...")

		// Test SOC2 jurisdiction and scope
		soc2Framework, err := frameworkService.GetFramework(context.Background(), "SOC2")
		assert.NoError(t, err, "SOC2 framework should be accessible")
		assert.Contains(t, soc2Framework.Jurisdiction, "US", "SOC2 should apply to US")
		assert.Contains(t, soc2Framework.Jurisdiction, "Global", "SOC2 should apply globally")
		assert.Contains(t, soc2Framework.Scope, "all", "SOC2 should apply to all business types")

		// Test GDPR jurisdiction and scope
		gdprFramework, err := frameworkService.GetFramework(context.Background(), "GDPR")
		assert.NoError(t, err, "GDPR framework should be accessible")
		assert.Contains(t, gdprFramework.Jurisdiction, "EU", "GDPR should apply to EU")
		assert.Contains(t, gdprFramework.Jurisdiction, "EEA", "GDPR should apply to EEA")
		assert.Contains(t, gdprFramework.Jurisdiction, "UK", "GDPR should apply to UK")
		assert.Contains(t, gdprFramework.Scope, "all", "GDPR should apply to all business types")

		// Test PCI DSS jurisdiction and scope
		pciFramework, err := frameworkService.GetFramework(context.Background(), "PCI_DSS")
		assert.NoError(t, err, "PCI DSS framework should be accessible")
		assert.Contains(t, pciFramework.Jurisdiction, "Global", "PCI DSS should apply globally")
		assert.Contains(t, pciFramework.Scope, "financial", "PCI DSS should apply to financial businesses")
		assert.Contains(t, pciFramework.Scope, "ecommerce", "PCI DSS should apply to ecommerce businesses")
		assert.Contains(t, pciFramework.Scope, "payment_processing", "PCI DSS should apply to payment processing businesses")

		// Test HIPAA jurisdiction and scope
		hipaaFramework, err := frameworkService.GetFramework(context.Background(), "HIPAA")
		assert.NoError(t, err, "HIPAA framework should be accessible")
		assert.Contains(t, hipaaFramework.Jurisdiction, "US", "HIPAA should apply to US")
		assert.Contains(t, hipaaFramework.Scope, "healthcare", "HIPAA should apply to healthcare businesses")
		assert.Contains(t, hipaaFramework.Scope, "health_tech", "HIPAA should apply to health tech businesses")

		t.Logf("✅ Jurisdiction and scope accuracy: All 4 frameworks validated with 100%% accuracy")
	})

	t.Run("Authority and Documentation Accuracy", func(t *testing.T) {
		// Test authority and documentation accuracy
		t.Log("Testing authority and documentation accuracy...")

		// Test SOC2 authority and documentation
		soc2Framework, err := frameworkService.GetFramework(context.Background(), "SOC2")
		assert.NoError(t, err, "SOC2 framework should be accessible")
		assert.Equal(t, "AICPA", soc2Framework.Authority, "SOC2 authority should be AICPA")
		assert.Contains(t, soc2Framework.Documentation, "SOC 2 Trust Services Criteria", "SOC2 should reference Trust Services Criteria")

		// Test GDPR authority and documentation
		gdprFramework, err := frameworkService.GetFramework(context.Background(), "GDPR")
		assert.NoError(t, err, "GDPR framework should be accessible")
		assert.Equal(t, "European Commission", gdprFramework.Authority, "GDPR authority should be European Commission")
		assert.Contains(t, gdprFramework.Documentation, "GDPR Regulation (EU) 2016/679", "GDPR should reference Regulation (EU) 2016/679")

		// Test PCI DSS authority and documentation
		pciFramework, err := frameworkService.GetFramework(context.Background(), "PCI_DSS")
		assert.NoError(t, err, "PCI DSS framework should be accessible")
		assert.Equal(t, "PCI Security Standards Council", pciFramework.Authority, "PCI DSS authority should be PCI Security Standards Council")
		assert.Contains(t, pciFramework.Documentation, "PCI DSS Requirements and Security Assessment Procedures", "PCI DSS should reference Requirements and Security Assessment Procedures")

		// Test HIPAA authority and documentation
		hipaaFramework, err := frameworkService.GetFramework(context.Background(), "HIPAA")
		assert.NoError(t, err, "HIPAA framework should be accessible")
		assert.Equal(t, "HHS", hipaaFramework.Authority, "HIPAA authority should be HHS")
		assert.Contains(t, hipaaFramework.Documentation, "HIPAA Privacy Rule", "HIPAA should reference Privacy Rule")
		assert.Contains(t, hipaaFramework.Documentation, "HIPAA Security Rule", "HIPAA should reference Security Rule")

		t.Logf("✅ Authority and documentation accuracy: All 4 frameworks validated with 100%% accuracy")
	})
}
