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

// TestSimpleComplianceIntegration tests basic compliance integration
func TestSimpleComplianceIntegration(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create compliance services
	frameworkService := compliance.NewComplianceFrameworkService(logger)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)

	t.Run("Service Integration", func(t *testing.T) {
		// Test service integration
		t.Log("Testing service integration...")

		businessID := "test-business-integration"
		frameworkID := "SOC2"

		// Test framework service
		framework, err := frameworkService.GetFramework(context.Background(), frameworkID)
		assert.NoError(t, err, "Framework service should work")
		assert.Equal(t, frameworkID, framework.ID, "Framework ID should match")

		// Test tracking service
		tracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "SOC2_CC6_1",
					Progress:      0.6,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
			},
		}

		err = trackingService.UpdateComplianceTracking(context.Background(), tracking)
		assert.NoError(t, err, "Tracking service should work")

		// Test retrieval
		retrievedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
		assert.NoError(t, err, "Tracking retrieval should work")
		assert.Equal(t, businessID, retrievedTracking.BusinessID, "Business ID should match")
		assert.Equal(t, frameworkID, retrievedTracking.FrameworkID, "Framework ID should match")

		t.Logf("✅ Service integration: Framework and tracking services integrated successfully")
	})

	t.Run("Multi-Framework Integration", func(t *testing.T) {
		// Test multi-framework integration
		t.Log("Testing multi-framework integration...")

		businessID := "test-business-multi"
		frameworks := []string{"SOC2", "GDPR"}

		// Test multiple frameworks
		for _, frameworkID := range frameworks {
			// Get framework
			framework, err := frameworkService.GetFramework(context.Background(), frameworkID)
			assert.NoError(t, err, "Framework %s should be accessible", frameworkID)
			assert.Equal(t, frameworkID, framework.ID, "Framework ID should match")

			// Create tracking
			tracking := &compliance.ComplianceTracking{
				BusinessID:  businessID,
				FrameworkID: frameworkID,
				Requirements: []compliance.RequirementTracking{
					{
						RequirementID: framework.Requirements[0],
						Progress:      0.5,
						Status:        "in_progress",
						LastAssessed:  time.Now(),
					},
				},
			}

			// Update tracking
			err = trackingService.UpdateComplianceTracking(context.Background(), tracking)
			assert.NoError(t, err, "Tracking update should work for framework %s", frameworkID)

			// Retrieve tracking
			retrievedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
			assert.NoError(t, err, "Tracking retrieval should work for framework %s", frameworkID)
			assert.Equal(t, businessID, retrievedTracking.BusinessID, "Business ID should match")
			assert.Equal(t, frameworkID, retrievedTracking.FrameworkID, "Framework ID should match")
		}

		t.Logf("✅ Multi-framework integration: %d frameworks integrated successfully", len(frameworks))
	})

	t.Run("Data Consistency Integration", func(t *testing.T) {
		// Test data consistency integration
		t.Log("Testing data consistency integration...")

		businessID := "test-business-consistency"
		frameworkID := "SOC2"

		// Create initial tracking
		tracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "SOC2_CC6_1",
					Progress:      0.3,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
				{
					RequirementID: "SOC2_CC6_2",
					Progress:      0.7,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
			},
		}

		// Update tracking
		err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
		assert.NoError(t, err, "Initial tracking update should work")

		// Retrieve and validate
		retrievedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
		assert.NoError(t, err, "Tracking retrieval should work")
		assert.Equal(t, 0.5, retrievedTracking.OverallProgress, "Overall progress should be 0.5")
		assert.Equal(t, "partial", retrievedTracking.ComplianceLevel, "Compliance level should be partial")
		assert.Equal(t, "medium", retrievedTracking.RiskLevel, "Risk level should be medium")

		// Update progress
		tracking.Requirements[0].Progress = 0.8
		tracking.Requirements[1].Progress = 0.9

		err = trackingService.UpdateComplianceTracking(context.Background(), tracking)
		assert.NoError(t, err, "Progress update should work")

		// Retrieve and validate updated data
		updatedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
		assert.NoError(t, err, "Updated tracking retrieval should work")
		assert.Equal(t, 0.5, updatedTracking.OverallProgress, "Updated overall progress should be 0.5")
		assert.Equal(t, "partial", updatedTracking.ComplianceLevel, "Updated compliance level should be partial")
		assert.Equal(t, "medium", updatedTracking.RiskLevel, "Updated risk level should be medium")

		t.Logf("✅ Data consistency integration: Data consistency validated successfully")
	})

	t.Run("Component Integration", func(t *testing.T) {
		// Test component integration
		t.Log("Testing component integration...")

		businessID := "test-business-component"
		frameworkID := "GDPR"

		// Test framework component
		framework, err := frameworkService.GetFramework(context.Background(), frameworkID)
		assert.NoError(t, err, "Framework component should work")
		assert.Equal(t, frameworkID, framework.ID, "Framework component should return correct data")

		// Test requirements component
		requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), frameworkID)
		assert.NoError(t, err, "Requirements component should work")
		assert.Len(t, requirements, 2, "Requirements component should return correct count")

		// Test tracking component
		tracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: requirements[0].ID,
					Progress:      0.6,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
			},
		}

		err = trackingService.UpdateComplianceTracking(context.Background(), tracking)
		assert.NoError(t, err, "Tracking component should work")

		// Test integration between components
		retrievedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
		assert.NoError(t, err, "Component integration should work")
		assert.Equal(t, businessID, retrievedTracking.BusinessID, "Component integration should maintain data integrity")
		assert.Equal(t, frameworkID, retrievedTracking.FrameworkID, "Component integration should maintain data integrity")

		t.Logf("✅ Component integration: All components integrated successfully")
	})

	t.Run("Cross-Component Data Flow", func(t *testing.T) {
		// Test cross-component data flow
		t.Log("Testing cross-component data flow...")

		businessID := "test-business-dataflow"
		frameworkID := "SOC2"

		// Get framework data
		_, err := frameworkService.GetFramework(context.Background(), frameworkID)
		assert.NoError(t, err, "Framework data should be accessible")

		// Get requirements data
		requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), frameworkID)
		assert.NoError(t, err, "Requirements data should be accessible")

		// Create tracking with framework and requirement data
		tracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: requirements[0].ID,
					Progress:      0.4,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
				{
					RequirementID: requirements[1].ID,
					Progress:      0.8,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
			},
		}

		// Update tracking
		err = trackingService.UpdateComplianceTracking(context.Background(), tracking)
		assert.NoError(t, err, "Cross-component data flow should work")

		// Retrieve and validate cross-component data
		retrievedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
		assert.NoError(t, err, "Cross-component data retrieval should work")

		// Use tolerance for floating-point comparison
		tolerance := 0.0001
		expectedProgress := 0.6
		if retrievedTracking.OverallProgress < expectedProgress-tolerance || retrievedTracking.OverallProgress > expectedProgress+tolerance {
			t.Errorf("Cross-component data should be consistent: expected %f, got %f", expectedProgress, retrievedTracking.OverallProgress)
		}
		assert.Equal(t, "partial", retrievedTracking.ComplianceLevel, "Cross-component data should be consistent")
		assert.Equal(t, "medium", retrievedTracking.RiskLevel, "Cross-component data should be consistent")

		t.Logf("✅ Cross-component data flow: Data flow validated successfully")
	})
}
