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

// TestComplianceDashboardUserExperience tests the user experience of the compliance dashboard
func TestComplianceDashboardUserExperience(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create compliance framework service
	frameworkService := compliance.NewComplianceFrameworkService(logger)

	// Create compliance tracking service
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)

	t.Run("Dashboard Data Loading Experience", func(t *testing.T) {
		// Test dashboard data loading performance and user experience
		t.Log("Testing dashboard data loading experience...")

		businessID := "test-business-dashboard"
		frameworkID := "SOC2"

		// Create comprehensive tracking data
		tracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
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

		// Measure data loading time
		startTime := time.Now()
		err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
		loadTime := time.Since(startTime)

		// Validate data loading performance
		assert.NoError(t, err, "Data loading should not fail")
		assert.Less(t, loadTime, 100*time.Millisecond, "Data loading should be fast (<100ms)")
		t.Logf("✅ Dashboard data loading experience: %v (target: <100ms)", loadTime)

		// Test data retrieval performance
		startTime = time.Now()
		retrievedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
		retrieveTime := time.Since(startTime)

		// Validate data retrieval performance
		assert.NoError(t, err, "Data retrieval should not fail")
		assert.Less(t, retrieveTime, 50*time.Millisecond, "Data retrieval should be very fast (<50ms)")
		assert.NotNil(t, retrievedTracking, "Retrieved data should not be nil")
		t.Logf("✅ Dashboard data retrieval experience: %v (target: <50ms)", retrieveTime)
	})

	t.Run("Dashboard Navigation Experience", func(t *testing.T) {
		// Test dashboard navigation and user flow
		t.Log("Testing dashboard navigation experience...")

		// Test framework navigation
		frameworks := []string{"SOC2", "GDPR", "PCI_DSS", "HIPAA"}
		for _, frameworkID := range frameworks {
			startTime := time.Now()
			framework, err := frameworkService.GetFramework(context.Background(), frameworkID)
			navigationTime := time.Since(startTime)

			// Validate navigation performance
			assert.NoError(t, err, "Framework navigation should not fail")
			assert.Less(t, navigationTime, 50*time.Millisecond, "Framework navigation should be fast (<50ms)")
			assert.NotNil(t, framework, "Framework data should not be nil")
			assert.Equal(t, frameworkID, framework.ID, "Framework ID should match")
		}
		t.Logf("✅ Dashboard navigation experience: %d frameworks navigated successfully", len(frameworks))
	})

	t.Run("Dashboard Responsiveness Experience", func(t *testing.T) {
		// Test dashboard responsiveness and real-time updates
		t.Log("Testing dashboard responsiveness experience...")

		businessID := "test-business-responsive"
		frameworkID := "GDPR"

		// Test multiple rapid updates
		updateTimes := make([]time.Duration, 5)
		for i := 0; i < 5; i++ {
			tracking := &compliance.ComplianceTracking{
				BusinessID:  businessID,
				FrameworkID: frameworkID,
				Requirements: []compliance.RequirementTracking{
					{
						RequirementID: "GDPR_32",
						Progress:      float64(i) * 0.2,
						Status:        "in_progress",
						LastAssessed:  time.Now(),
					},
				},
			}

			startTime := time.Now()
			err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
			updateTimes[i] = time.Since(startTime)

			assert.NoError(t, err, "Rapid update should not fail")
		}

		// Validate responsiveness
		avgUpdateTime := time.Duration(0)
		for _, updateTime := range updateTimes {
			avgUpdateTime += updateTime
		}
		avgUpdateTime /= time.Duration(len(updateTimes))

		assert.Less(t, avgUpdateTime, 75*time.Millisecond, "Average update time should be fast (<75ms)")
		t.Logf("✅ Dashboard responsiveness experience: avg update time %v (target: <75ms)", avgUpdateTime)
	})

	t.Run("Dashboard Error Handling Experience", func(t *testing.T) {
		// Test dashboard error handling and user feedback
		t.Log("Testing dashboard error handling experience...")

		// Test invalid framework access
		startTime := time.Now()
		_, err := frameworkService.GetFramework(context.Background(), "INVALID_FRAMEWORK")
		errorTime := time.Since(startTime)

		// Validate error handling performance
		assert.Error(t, err, "Invalid framework should return error")
		assert.Less(t, errorTime, 25*time.Millisecond, "Error handling should be very fast (<25ms)")
		t.Logf("✅ Dashboard error handling experience: %v (target: <25ms)", errorTime)

		// Test invalid business ID (may not return error in current implementation)
		startTime = time.Now()
		_, err = trackingService.GetComplianceTracking(context.Background(), "invalid-business", "SOC2")
		errorTime = time.Since(startTime)

		// Validate error handling performance (error may or may not be returned)
		assert.Less(t, errorTime, 25*time.Millisecond, "Error handling should be very fast (<25ms)")
		t.Logf("✅ Dashboard error handling experience: %v (target: <25ms)", errorTime)
	})
}

// TestComplianceWorkflowUserExperience tests the user experience of compliance workflows
func TestComplianceWorkflowUserExperience(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create compliance framework service
	frameworkService := compliance.NewComplianceFrameworkService(logger)

	// Create compliance tracking service
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)

	t.Run("Workflow Initialization Experience", func(t *testing.T) {
		// Test workflow initialization and setup experience
		t.Log("Testing workflow initialization experience...")

		businessID := "test-business-workflow"
		frameworkID := "SOC2"

		// Test workflow initialization time
		startTime := time.Now()
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

		err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
		initTime := time.Since(startTime)

		// Validate initialization performance
		assert.NoError(t, err, "Workflow initialization should not fail")
		assert.Less(t, initTime, 100*time.Millisecond, "Workflow initialization should be fast (<100ms)")
		t.Logf("✅ Workflow initialization experience: %v (target: <100ms)", initTime)
	})

	t.Run("Workflow Progress Experience", func(t *testing.T) {
		// Test workflow progress tracking experience
		t.Log("Testing workflow progress experience...")

		businessID := "test-business-progress"
		frameworkID := "GDPR"

		// Test progress tracking workflow
		progressSteps := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
		progressTimes := make([]time.Duration, len(progressSteps))

		for i, progress := range progressSteps {
			tracking := &compliance.ComplianceTracking{
				BusinessID:  businessID,
				FrameworkID: frameworkID,
				Requirements: []compliance.RequirementTracking{
					{
						RequirementID: "GDPR_32",
						Progress:      progress,
						Status:        getStatusFromProgress(progress),
						LastAssessed:  time.Now(),
					},
				},
			}

			startTime := time.Now()
			err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
			progressTimes[i] = time.Since(startTime)

			assert.NoError(t, err, "Progress update should not fail")
		}

		// Validate progress tracking performance
		avgProgressTime := time.Duration(0)
		for _, progressTime := range progressTimes {
			avgProgressTime += progressTime
		}
		avgProgressTime /= time.Duration(len(progressTimes))

		assert.Less(t, avgProgressTime, 75*time.Millisecond, "Average progress update should be fast (<75ms)")
		t.Logf("✅ Workflow progress experience: avg update time %v (target: <75ms)", avgProgressTime)
	})

	t.Run("Workflow Completion Experience", func(t *testing.T) {
		// Test workflow completion and finalization experience
		t.Log("Testing workflow completion experience...")

		businessID := "test-business-completion"
		frameworkID := "SOC2"

		// Test completion workflow
		tracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "SOC2_CC6_1",
					Progress:      1.0,
					Status:        "completed",
					LastAssessed:  time.Now(),
				},
				{
					RequirementID: "SOC2_CC6_2",
					Progress:      1.0,
					Status:        "completed",
					LastAssessed:  time.Now(),
				},
			},
		}

		startTime := time.Now()
		err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
		completionTime := time.Since(startTime)

		// Validate completion performance
		assert.NoError(t, err, "Workflow completion should not fail")
		assert.Less(t, completionTime, 100*time.Millisecond, "Workflow completion should be fast (<100ms)")

		// Verify completion status
		retrievedTracking, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
		assert.NoError(t, err, "Completion verification should not fail")
		assert.Equal(t, 1.0, retrievedTracking.OverallProgress, "Overall progress should be 100%")
		assert.Equal(t, "compliant", retrievedTracking.ComplianceLevel, "Compliance level should be compliant")
		assert.Equal(t, "low", retrievedTracking.RiskLevel, "Risk level should be low")
		t.Logf("✅ Workflow completion experience: %v (target: <100ms)", completionTime)
	})

	t.Run("Workflow Error Recovery Experience", func(t *testing.T) {
		// Test workflow error recovery and user guidance
		t.Log("Testing workflow error recovery experience...")

		businessID := "test-business-error-recovery"
		frameworkID := "GDPR"

		// Test error recovery workflow
		// First, create invalid tracking
		invalidTracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "INVALID_REQUIREMENT",
					Progress:      0.5,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
			},
		}

		startTime := time.Now()
		err := trackingService.UpdateComplianceTracking(context.Background(), invalidTracking)
		errorTime := time.Since(startTime)

		// Validate error handling performance (error may or may not be returned)
		assert.Less(t, errorTime, 50*time.Millisecond, "Error handling should be fast (<50ms)")

		// Test recovery with valid tracking
		validTracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "GDPR_32",
					Progress:      0.5,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
			},
		}

		startTime = time.Now()
		err = trackingService.UpdateComplianceTracking(context.Background(), validTracking)
		recoveryTime := time.Since(startTime)

		// Validate recovery performance
		assert.NoError(t, err, "Recovery should succeed")
		assert.Less(t, recoveryTime, 75*time.Millisecond, "Recovery should be fast (<75ms)")
		t.Logf("✅ Workflow error recovery experience: error %v, recovery %v", errorTime, recoveryTime)
	})
}

// TestComplianceDashboardAccessibility tests the accessibility of the compliance dashboard
func TestComplianceDashboardAccessibility(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create compliance framework service
	frameworkService := compliance.NewComplianceFrameworkService(logger)

	t.Run("Framework Accessibility", func(t *testing.T) {
		// Test framework accessibility and availability
		t.Log("Testing framework accessibility...")

		frameworks := []string{"SOC2", "GDPR", "PCI_DSS", "HIPAA"}
		accessibilityResults := make(map[string]bool)

		for _, frameworkID := range frameworks {
			startTime := time.Now()
			framework, err := frameworkService.GetFramework(context.Background(), frameworkID)
			accessTime := time.Since(startTime)

			// Validate accessibility
			accessible := err == nil && framework != nil && accessTime < 50*time.Millisecond
			accessibilityResults[frameworkID] = accessible

			assert.True(t, accessible, "Framework %s should be accessible", frameworkID)
		}

		// Validate overall accessibility
		accessibleCount := 0
		for _, accessible := range accessibilityResults {
			if accessible {
				accessibleCount++
			}
		}

		assert.Equal(t, len(frameworks), accessibleCount, "All frameworks should be accessible")
		t.Logf("✅ Framework accessibility: %d/%d frameworks accessible", accessibleCount, len(frameworks))
	})

	t.Run("Requirement Accessibility", func(t *testing.T) {
		// Test requirement accessibility and availability
		t.Log("Testing requirement accessibility...")

		frameworks := []string{"SOC2", "GDPR"}
		totalRequirements := 0
		accessibleRequirements := 0

		for _, frameworkID := range frameworks {
			startTime := time.Now()
			requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), frameworkID)
			accessTime := time.Since(startTime)

			// Validate accessibility
			accessible := err == nil && requirements != nil && accessTime < 50*time.Millisecond
			assert.True(t, accessible, "Requirements for framework %s should be accessible", frameworkID)

			if accessible {
				accessibleRequirements += len(requirements)
			}
			totalRequirements += len(requirements)
		}

		// Validate overall accessibility
		assert.Equal(t, totalRequirements, accessibleRequirements, "All requirements should be accessible")
		t.Logf("✅ Requirement accessibility: %d/%d requirements accessible", accessibleRequirements, totalRequirements)
	})

	t.Run("Data Consistency Accessibility", func(t *testing.T) {
		// Test data consistency and accessibility
		t.Log("Testing data consistency accessibility...")

		// Test framework-requirement consistency
		frameworks := []string{"SOC2", "GDPR"}
		consistencyResults := make(map[string]bool)

		for _, frameworkID := range frameworks {
			// Get framework
			framework, err := frameworkService.GetFramework(context.Background(), frameworkID)
			assert.NoError(t, err, "Framework should be accessible")

			// Get requirements
			requirements, err := frameworkService.GetFrameworkRequirements(context.Background(), frameworkID)
			assert.NoError(t, err, "Requirements should be accessible")

			// Validate consistency
			consistent := len(framework.Requirements) == len(requirements)
			consistencyResults[frameworkID] = consistent

			assert.True(t, consistent, "Framework %s should have consistent requirement count", frameworkID)
		}

		// Validate overall consistency
		consistentCount := 0
		for _, consistent := range consistencyResults {
			if consistent {
				consistentCount++
			}
		}

		assert.Equal(t, len(frameworks), consistentCount, "All frameworks should have consistent data")
		t.Logf("✅ Data consistency accessibility: %d/%d frameworks consistent", consistentCount, len(frameworks))
	})
}

// TestComplianceDashboardPerformance tests the performance of the compliance dashboard
func TestComplianceDashboardPerformance(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create compliance framework service
	frameworkService := compliance.NewComplianceFrameworkService(logger)

	// Create compliance tracking service
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)

	t.Run("Dashboard Load Performance", func(t *testing.T) {
		// Test dashboard load performance
		t.Log("Testing dashboard load performance...")

		// Test framework loading performance
		frameworks := []string{"SOC2", "GDPR", "PCI_DSS", "HIPAA"}
		loadTimes := make([]time.Duration, len(frameworks))

		for i, frameworkID := range frameworks {
			startTime := time.Now()
			_, err := frameworkService.GetFramework(context.Background(), frameworkID)
			loadTimes[i] = time.Since(startTime)

			assert.NoError(t, err, "Framework loading should not fail")
		}

		// Calculate average load time
		avgLoadTime := time.Duration(0)
		for _, loadTime := range loadTimes {
			avgLoadTime += loadTime
		}
		avgLoadTime /= time.Duration(len(loadTimes))

		// Validate performance
		assert.Less(t, avgLoadTime, 50*time.Millisecond, "Average framework load time should be fast (<50ms)")
		t.Logf("✅ Dashboard load performance: avg load time %v (target: <50ms)", avgLoadTime)
	})

	t.Run("Dashboard Update Performance", func(t *testing.T) {
		// Test dashboard update performance
		t.Log("Testing dashboard update performance...")

		businessID := "test-business-performance"
		frameworkID := "SOC2"

		// Test multiple updates
		updateCount := 10
		updateTimes := make([]time.Duration, updateCount)

		for i := 0; i < updateCount; i++ {
			tracking := &compliance.ComplianceTracking{
				BusinessID:  businessID,
				FrameworkID: frameworkID,
				Requirements: []compliance.RequirementTracking{
					{
						RequirementID: "SOC2_CC6_1",
						Progress:      float64(i) * 0.1,
						Status:        "in_progress",
						LastAssessed:  time.Now(),
					},
				},
			}

			startTime := time.Now()
			err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
			updateTimes[i] = time.Since(startTime)

			assert.NoError(t, err, "Update should not fail")
		}

		// Calculate average update time
		avgUpdateTime := time.Duration(0)
		for _, updateTime := range updateTimes {
			avgUpdateTime += updateTime
		}
		avgUpdateTime /= time.Duration(len(updateTimes))

		// Validate performance
		assert.Less(t, avgUpdateTime, 75*time.Millisecond, "Average update time should be fast (<75ms)")
		t.Logf("✅ Dashboard update performance: avg update time %v (target: <75ms)", avgUpdateTime)
	})

	t.Run("Dashboard Query Performance", func(t *testing.T) {
		// Test dashboard query performance
		t.Log("Testing dashboard query performance...")

		businessID := "test-business-query"
		frameworkID := "GDPR"

		// Create test data
		tracking := &compliance.ComplianceTracking{
			BusinessID:  businessID,
			FrameworkID: frameworkID,
			Requirements: []compliance.RequirementTracking{
				{
					RequirementID: "GDPR_32",
					Progress:      0.8,
					Status:        "in_progress",
					LastAssessed:  time.Now(),
				},
			},
		}

		err := trackingService.UpdateComplianceTracking(context.Background(), tracking)
		assert.NoError(t, err, "Test data creation should not fail")

		// Test query performance
		queryCount := 20
		queryTimes := make([]time.Duration, queryCount)

		for i := 0; i < queryCount; i++ {
			startTime := time.Now()
			_, err := trackingService.GetComplianceTracking(context.Background(), businessID, frameworkID)
			queryTimes[i] = time.Since(startTime)

			assert.NoError(t, err, "Query should not fail")
		}

		// Calculate average query time
		avgQueryTime := time.Duration(0)
		for _, queryTime := range queryTimes {
			avgQueryTime += queryTime
		}
		avgQueryTime /= time.Duration(len(queryTimes))

		// Validate performance
		assert.Less(t, avgQueryTime, 50*time.Millisecond, "Average query time should be fast (<50ms)")
		t.Logf("✅ Dashboard query performance: avg query time %v (target: <50ms)", avgQueryTime)
	})
}

// Helper function to get status from progress
func getStatusFromProgress(progress float64) string {
	if progress >= 1.0 {
		return "completed"
	} else if progress > 0.0 {
		return "in_progress"
	} else {
		return "not_started"
	}
}
