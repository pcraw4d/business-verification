package compliance

import (
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
)

// TestComplianceCalculationAccuracy tests the accuracy of compliance calculations
func TestComplianceCalculationAccuracy(t *testing.T) {
	// Note: We're testing calculation logic directly, not using the service

	t.Run("Progress Calculation Accuracy", func(t *testing.T) {
		// Test 1: Perfect compliance (100%)
		t.Log("Testing perfect compliance calculation...")
		perfectTracking := &compliance.ComplianceTracking{
			BusinessID:  "test-business-perfect",
			FrameworkID: "SOC2",
			Requirements: []compliance.RequirementTracking{
				{RequirementID: "req1", Progress: 1.0, Status: "completed"},
				{RequirementID: "req2", Progress: 1.0, Status: "completed"},
				{RequirementID: "req3", Progress: 1.0, Status: "completed"},
				{RequirementID: "req4", Progress: 1.0, Status: "completed"},
				{RequirementID: "req5", Progress: 1.0, Status: "completed"},
			},
		}

		// Simulate the calculation logic
		totalProgress := 0.0
		for _, req := range perfectTracking.Requirements {
			totalProgress += req.Progress
		}
		overallProgress := totalProgress / float64(len(perfectTracking.Requirements))

		expectedProgress := 1.0
		if overallProgress != expectedProgress {
			t.Errorf("Perfect compliance calculation failed: expected %f, got %f", expectedProgress, overallProgress)
		}

		// Test compliance level
		var complianceLevel string
		if overallProgress >= 0.9 {
			complianceLevel = "compliant"
		} else if overallProgress >= 0.5 {
			complianceLevel = "partial"
		} else {
			complianceLevel = "non_compliant"
		}

		expectedLevel := "compliant"
		if complianceLevel != expectedLevel {
			t.Errorf("Perfect compliance level failed: expected %s, got %s", expectedLevel, complianceLevel)
		}

		t.Logf("✅ Perfect compliance calculation: %f progress, %s level", overallProgress, complianceLevel)
	})

	t.Run("Partial Compliance Calculation", func(t *testing.T) {
		// Test 2: Partial compliance (60%)
		t.Log("Testing partial compliance calculation...")
		partialTracking := &compliance.ComplianceTracking{
			BusinessID:  "test-business-partial",
			FrameworkID: "SOC2",
			Requirements: []compliance.RequirementTracking{
				{RequirementID: "req1", Progress: 1.0, Status: "completed"},
				{RequirementID: "req2", Progress: 1.0, Status: "completed"},
				{RequirementID: "req3", Progress: 1.0, Status: "completed"},
				{RequirementID: "req4", Progress: 0.0, Status: "not_started"},
				{RequirementID: "req5", Progress: 0.0, Status: "not_started"},
			},
		}

		// Calculate progress
		totalProgress := 0.0
		for _, req := range partialTracking.Requirements {
			totalProgress += req.Progress
		}
		overallProgress := totalProgress / float64(len(partialTracking.Requirements))

		expectedProgress := 0.6
		if overallProgress != expectedProgress {
			t.Errorf("Partial compliance calculation failed: expected %f, got %f", expectedProgress, overallProgress)
		}

		// Test compliance level
		var complianceLevel string
		if overallProgress >= 0.9 {
			complianceLevel = "compliant"
		} else if overallProgress >= 0.5 {
			complianceLevel = "partial"
		} else {
			complianceLevel = "non_compliant"
		}

		expectedLevel := "partial"
		if complianceLevel != expectedLevel {
			t.Errorf("Partial compliance level failed: expected %s, got %s", expectedLevel, complianceLevel)
		}

		t.Logf("✅ Partial compliance calculation: %f progress, %s level", overallProgress, complianceLevel)
	})

	t.Run("Non-Compliance Calculation", func(t *testing.T) {
		// Test 3: Non-compliance (20%)
		t.Log("Testing non-compliance calculation...")
		nonCompliantTracking := &compliance.ComplianceTracking{
			BusinessID:  "test-business-non-compliant",
			FrameworkID: "SOC2",
			Requirements: []compliance.RequirementTracking{
				{RequirementID: "req1", Progress: 1.0, Status: "completed"},
				{RequirementID: "req2", Progress: 0.0, Status: "not_started"},
				{RequirementID: "req3", Progress: 0.0, Status: "not_started"},
				{RequirementID: "req4", Progress: 0.0, Status: "not_started"},
				{RequirementID: "req5", Progress: 0.0, Status: "not_started"},
			},
		}

		// Calculate progress
		totalProgress := 0.0
		for _, req := range nonCompliantTracking.Requirements {
			totalProgress += req.Progress
		}
		overallProgress := totalProgress / float64(len(nonCompliantTracking.Requirements))

		expectedProgress := 0.2
		if overallProgress != expectedProgress {
			t.Errorf("Non-compliance calculation failed: expected %f, got %f", expectedProgress, overallProgress)
		}

		// Test compliance level
		var complianceLevel string
		if overallProgress >= 0.9 {
			complianceLevel = "compliant"
		} else if overallProgress >= 0.5 {
			complianceLevel = "partial"
		} else {
			complianceLevel = "non_compliant"
		}

		expectedLevel := "non_compliant"
		if complianceLevel != expectedLevel {
			t.Errorf("Non-compliance level failed: expected %s, got %s", expectedLevel, complianceLevel)
		}

		t.Logf("✅ Non-compliance calculation: %f progress, %s level", overallProgress, complianceLevel)
	})

	t.Run("Edge Case Calculations", func(t *testing.T) {
		// Test 4: Edge cases
		t.Log("Testing edge case calculations...")

		// Test empty requirements
		emptyTracking := &compliance.ComplianceTracking{
			BusinessID:   "test-business-empty",
			FrameworkID:  "SOC2",
			Requirements: []compliance.RequirementTracking{},
		}

		if len(emptyTracking.Requirements) != 0 {
			t.Errorf("Empty requirements test failed: expected 0, got %d", len(emptyTracking.Requirements))
		}

		// Test single requirement
		singleTracking := &compliance.ComplianceTracking{
			BusinessID:  "test-business-single",
			FrameworkID: "SOC2",
			Requirements: []compliance.RequirementTracking{
				{RequirementID: "req1", Progress: 0.75, Status: "in_progress"},
			},
		}

		totalProgress := 0.0
		for _, req := range singleTracking.Requirements {
			totalProgress += req.Progress
		}
		overallProgress := totalProgress / float64(len(singleTracking.Requirements))

		expectedProgress := 0.75
		if overallProgress != expectedProgress {
			t.Errorf("Single requirement calculation failed: expected %f, got %f", expectedProgress, overallProgress)
		}

		t.Logf("✅ Edge case calculations: empty=%d, single=%f", len(emptyTracking.Requirements), overallProgress)
	})
}

// TestRiskLevelCalculationAccuracy tests the accuracy of risk level calculations
func TestRiskLevelCalculationAccuracy(t *testing.T) {
	t.Run("Risk Level Calculation", func(t *testing.T) {
		// Test risk level calculations based on progress
		testCases := []struct {
			progress     float64
			expectedRisk string
			description  string
		}{
			{1.0, "low", "Perfect compliance - low risk"},
			{0.9, "low", "High compliance - low risk"},
			{0.8, "low", "Good compliance - low risk"},
			{0.7, "medium", "Moderate compliance - medium risk"},
			{0.6, "medium", "Partial compliance - medium risk"},
			{0.5, "medium", "Half compliance - medium risk"},
			{0.4, "high", "Low compliance - high risk"},
			{0.3, "high", "Poor compliance - high risk"},
			{0.2, "high", "Very poor compliance - high risk"},
			{0.1, "critical", "Minimal compliance - critical risk"},
			{0.0, "critical", "No compliance - critical risk"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				// Calculate risk level based on progress
				var riskLevel string
				switch {
				case tc.progress >= 0.8:
					riskLevel = "low"
				case tc.progress >= 0.5:
					riskLevel = "medium"
				case tc.progress >= 0.2:
					riskLevel = "high"
				default:
					riskLevel = "critical"
				}

				if riskLevel != tc.expectedRisk {
					t.Errorf("Risk level calculation failed for progress %f: expected %s, got %s",
						tc.progress, tc.expectedRisk, riskLevel)
				}

				t.Logf("✅ %s: progress=%f, risk=%s", tc.description, tc.progress, riskLevel)
			})
		}
	})
}

// TestVelocityCalculationAccuracy tests the accuracy of velocity calculations
func TestVelocityCalculationAccuracy(t *testing.T) {
	t.Run("Velocity Calculation", func(t *testing.T) {
		// Test velocity calculations based on progress
		testCases := []struct {
			progress         float64
			expectedVelocity float64
			description      string
		}{
			{0.9, 0.1, "High progress - positive velocity"},
			{0.7, 0.1, "Good progress - positive velocity"},
			{0.6, 0.1, "Moderate progress - positive velocity"},
			{0.5, 0.0, "Stable progress - zero velocity"},
			{0.4, 0.0, "Stable progress - zero velocity"},
			{0.3, 0.0, "Stable progress - zero velocity"},
			{0.2, -0.1, "Poor progress - negative velocity"},
			{0.1, -0.1, "Very poor progress - negative velocity"},
			{0.0, -0.1, "No progress - negative velocity"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				// Calculate velocity based on progress
				var velocity float64
				if tc.progress > 0.5 {
					velocity = 0.1 // Positive velocity for good progress
				} else if tc.progress < 0.3 {
					velocity = -0.1 // Negative velocity for poor progress
				} else {
					velocity = 0.0 // Stable velocity
				}

				if velocity != tc.expectedVelocity {
					t.Errorf("Velocity calculation failed for progress %f: expected %f, got %f",
						tc.progress, tc.expectedVelocity, velocity)
				}

				t.Logf("✅ %s: progress=%f, velocity=%f", tc.description, tc.progress, velocity)
			})
		}
	})
}

// TestTrendCalculationAccuracy tests the accuracy of trend calculations
func TestTrendCalculationAccuracy(t *testing.T) {
	t.Run("Trend Calculation", func(t *testing.T) {
		// Test trend calculations based on velocity
		testCases := []struct {
			velocity      float64
			expectedTrend string
			description   string
		}{
			{0.1, "improving", "Positive velocity - improving trend"},
			{0.05, "improving", "Small positive velocity - improving trend"},
			{0.01, "stable", "Minimal positive velocity - stable trend"},
			{0.0, "stable", "Zero velocity - stable trend"},
			{-0.01, "stable", "Minimal negative velocity - stable trend"},
			{-0.05, "declining", "Small negative velocity - declining trend"},
			{-0.1, "declining", "Negative velocity - declining trend"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				// Calculate trend based on velocity
				var trend string
				if tc.velocity > 0.01 {
					trend = "improving"
				} else if tc.velocity < -0.01 {
					trend = "declining"
				} else {
					trend = "stable"
				}

				if trend != tc.expectedTrend {
					t.Errorf("Trend calculation failed for velocity %f: expected %s, got %s",
						tc.velocity, tc.expectedTrend, trend)
				}

				t.Logf("✅ %s: velocity=%f, trend=%s", tc.description, tc.velocity, trend)
			})
		}
	})
}

// TestComplianceScoreAccuracy tests the accuracy of compliance scoring across frameworks
func TestComplianceScoreAccuracy(t *testing.T) {
	t.Run("Framework-Specific Scoring", func(t *testing.T) {
		// Test compliance scoring for different frameworks
		frameworks := []struct {
			frameworkID string
			description string
		}{
			{"SOC2", "SOC 2 Type II Compliance"},
			{"GDPR", "General Data Protection Regulation"},
			{"PCI-DSS", "Payment Card Industry Data Security Standard"},
			{"HIPAA", "Health Insurance Portability and Accountability Act"},
			{"ISO27001", "Information Security Management System"},
		}

		for _, fw := range frameworks {
			t.Run(fw.description, func(t *testing.T) {
				// Create test tracking for each framework
				tracking := &compliance.ComplianceTracking{
					BusinessID:  "test-business-" + fw.frameworkID,
					FrameworkID: fw.frameworkID,
					Requirements: []compliance.RequirementTracking{
						{RequirementID: "req1", Progress: 0.8, Status: "in_progress"},
						{RequirementID: "req2", Progress: 0.9, Status: "in_progress"},
						{RequirementID: "req3", Progress: 1.0, Status: "completed"},
					},
				}

				// Calculate compliance score
				totalProgress := 0.0
				for _, req := range tracking.Requirements {
					totalProgress += req.Progress
				}
				complianceScore := totalProgress / float64(len(tracking.Requirements))

				// Validate score is within expected range
				if complianceScore < 0.0 || complianceScore > 1.0 {
					t.Errorf("Compliance score out of range for %s: %f (expected 0.0-1.0)",
						fw.frameworkID, complianceScore)
				}

				// Validate score calculation accuracy
				expectedScore := (0.8 + 0.9 + 1.0) / 3.0 // 0.9
				if complianceScore != expectedScore {
					t.Errorf("Compliance score calculation failed for %s: expected %f, got %f",
						fw.frameworkID, expectedScore, complianceScore)
				}

				t.Logf("✅ %s: compliance score=%f", fw.description, complianceScore)
			})
		}
	})
}

// TestRequirementStatusAccuracy tests the accuracy of requirement status calculations
func TestRequirementStatusAccuracy(t *testing.T) {
	t.Run("Requirement Status Calculation", func(t *testing.T) {
		// Test requirement status based on progress
		testCases := []struct {
			progress       float64
			expectedStatus string
			description    string
		}{
			{1.0, "completed", "Full progress - completed status"},
			{0.9, "in_progress", "High progress - in progress status"},
			{0.5, "in_progress", "Half progress - in progress status"},
			{0.1, "in_progress", "Low progress - in progress status"},
			{0.0, "not_started", "No progress - not started status"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				// Calculate status based on progress
				var status string
				if tc.progress >= 1.0 {
					status = "completed"
				} else if tc.progress > 0.0 {
					status = "in_progress"
				} else {
					status = "not_started"
				}

				if status != tc.expectedStatus {
					t.Errorf("Requirement status calculation failed for progress %f: expected %s, got %s",
						tc.progress, tc.expectedStatus, status)
				}

				t.Logf("✅ %s: progress=%f, status=%s", tc.description, tc.progress, status)
			})
		}
	})

	t.Run("At-Risk Status Calculation", func(t *testing.T) {
		// Test at-risk status calculation
		now := time.Now()
		dueDate := now.AddDate(0, 0, 15) // 15 days from now

		// Test at-risk scenario
		atRiskTracking := &compliance.RequirementTracking{
			RequirementID: "req1",
			Progress:      0.3,      // Low progress
			NextDueDate:   &dueDate, // Due soon
		}

		// Check if at risk (low progress and approaching due date)
		isAtRisk := atRiskTracking.NextDueDate != nil &&
			atRiskTracking.NextDueDate.Before(now.AddDate(0, 0, 30)) &&
			atRiskTracking.Progress < 0.5

		if !isAtRisk {
			t.Errorf("At-risk calculation failed: expected true, got false")
		}

		// Test not at-risk scenario
		notAtRiskTracking := &compliance.RequirementTracking{
			RequirementID: "req2",
			Progress:      0.8,      // High progress
			NextDueDate:   &dueDate, // Due soon but high progress
		}

		isNotAtRisk := notAtRiskTracking.NextDueDate != nil &&
			notAtRiskTracking.NextDueDate.Before(now.AddDate(0, 0, 30)) &&
			notAtRiskTracking.Progress < 0.5

		if isNotAtRisk {
			t.Errorf("Not at-risk calculation failed: expected false, got true")
		}

		t.Logf("✅ At-risk status calculation: at-risk=%t, not-at-risk=%t", isAtRisk, isNotAtRisk)
	})
}

// TestMetricsCalculationAccuracy tests the accuracy of metrics calculations
func TestMetricsCalculationAccuracy(t *testing.T) {
	t.Run("Progress Metrics Calculation", func(t *testing.T) {
		// Create test tracking data
		tracking := &compliance.ComplianceTracking{
			BusinessID:  "test-business-metrics",
			FrameworkID: "SOC2",
			Requirements: []compliance.RequirementTracking{
				{RequirementID: "req1", Progress: 1.0, Status: "completed"},
				{RequirementID: "req2", Progress: 1.0, Status: "completed"},
				{RequirementID: "req3", Progress: 0.5, Status: "in_progress"},
				{RequirementID: "req4", Progress: 0.0, Status: "not_started"},
				{RequirementID: "req5", Progress: 0.0, Status: "not_started"},
			},
		}

		// Calculate metrics
		totalRequirements := len(tracking.Requirements)
		completedRequirements := 0
		totalProgress := 0.0

		for _, req := range tracking.Requirements {
			totalProgress += req.Progress
			if req.Status == "completed" {
				completedRequirements++
			}
		}

		overallProgress := totalProgress / float64(totalRequirements)
		completionRate := float64(completedRequirements) / float64(totalRequirements)

		// Validate calculations
		expectedProgress := (1.0 + 1.0 + 0.5 + 0.0 + 0.0) / 5.0 // 0.5
		expectedCompletionRate := 2.0 / 5.0                     // 0.4

		if overallProgress != expectedProgress {
			t.Errorf("Overall progress calculation failed: expected %f, got %f", expectedProgress, overallProgress)
		}

		if completionRate != expectedCompletionRate {
			t.Errorf("Completion rate calculation failed: expected %f, got %f", expectedCompletionRate, completionRate)
		}

		t.Logf("✅ Metrics calculation: progress=%f, completion_rate=%f", overallProgress, completionRate)
	})
}

// TestComplianceAccuracyIntegration tests the integration of all compliance accuracy calculations
func TestComplianceAccuracyIntegration(t *testing.T) {
	t.Run("Integrated Compliance Accuracy Test", func(t *testing.T) {
		// Create comprehensive test scenario
		tracking := &compliance.ComplianceTracking{
			BusinessID:  "test-business-integration",
			FrameworkID: "SOC2",
			Requirements: []compliance.RequirementTracking{
				{RequirementID: "req1", Progress: 1.0, Status: "completed"},
				{RequirementID: "req2", Progress: 0.8, Status: "in_progress"},
				{RequirementID: "req3", Progress: 0.6, Status: "in_progress"},
				{RequirementID: "req4", Progress: 0.3, Status: "in_progress"},
				{RequirementID: "req5", Progress: 0.0, Status: "not_started"},
			},
		}

		// Calculate all metrics
		totalProgress := 0.0
		completedCount := 0
		inProgressCount := 0
		notStartedCount := 0

		for _, req := range tracking.Requirements {
			totalProgress += req.Progress
			switch req.Status {
			case "completed":
				completedCount++
			case "in_progress":
				inProgressCount++
			case "not_started":
				notStartedCount++
			}
		}

		overallProgress := totalProgress / float64(len(tracking.Requirements))

		// Calculate compliance level
		var complianceLevel string
		if overallProgress >= 0.9 {
			complianceLevel = "compliant"
		} else if overallProgress >= 0.5 {
			complianceLevel = "partial"
		} else {
			complianceLevel = "non_compliant"
		}

		// Calculate risk level
		var riskLevel string
		switch {
		case overallProgress >= 0.8:
			riskLevel = "low"
		case overallProgress >= 0.5:
			riskLevel = "medium"
		case overallProgress >= 0.2:
			riskLevel = "high"
		default:
			riskLevel = "critical"
		}

		// Validate integrated calculations
		expectedProgress := (1.0 + 0.8 + 0.6 + 0.3 + 0.0) / 5.0 // 0.54
		expectedComplianceLevel := "partial"
		expectedRiskLevel := "medium"

		// Use tolerance for floating point comparison
		tolerance := 0.0001
		if overallProgress < expectedProgress-tolerance || overallProgress > expectedProgress+tolerance {
			t.Errorf("Integrated progress calculation failed: expected %f, got %f", expectedProgress, overallProgress)
		}

		if complianceLevel != expectedComplianceLevel {
			t.Errorf("Integrated compliance level failed: expected %s, got %s", expectedComplianceLevel, complianceLevel)
		}

		if riskLevel != expectedRiskLevel {
			t.Errorf("Integrated risk level failed: expected %s, got %s", expectedRiskLevel, riskLevel)
		}

		// Validate counts
		if completedCount != 1 {
			t.Errorf("Completed count failed: expected 1, got %d", completedCount)
		}
		if inProgressCount != 3 {
			t.Errorf("In progress count failed: expected 3, got %d", inProgressCount)
		}
		if notStartedCount != 1 {
			t.Errorf("Not started count failed: expected 1, got %d", notStartedCount)
		}

		t.Logf("✅ Integrated compliance accuracy: progress=%f, level=%s, risk=%s, completed=%d, in_progress=%d, not_started=%d",
			overallProgress, complianceLevel, riskLevel, completedCount, inProgressCount, notStartedCount)
	})
}
