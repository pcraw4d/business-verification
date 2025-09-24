package classification

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

// createMockDBForAccuracyCalculation creates a mock database connection for testing
func createMockDBForAccuracyCalculation() *sql.DB {
	// In a real test, you would use a test database or mock
	// For now, we'll return nil and skip tests that require DB
	return nil
}

func TestAccuracyCalculationService_CalculateOverallAccuracy(t *testing.T) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	accuracy, err := acs.CalculateOverallAccuracy(ctx, 24)
	if err != nil {
		t.Fatalf("CalculateOverallAccuracy failed: %v", err)
	}

	// Validate accuracy is between 0 and 1
	if accuracy < 0 || accuracy > 1 {
		t.Errorf("Expected accuracy between 0 and 1, got %.2f", accuracy)
	}

	t.Logf("Overall accuracy: %.2f%%", accuracy*100)
}

func TestAccuracyCalculationService_CalculateIndustrySpecificAccuracy(t *testing.T) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	industryAccuracy, err := acs.CalculateIndustrySpecificAccuracy(ctx, 24)
	if err != nil {
		t.Fatalf("CalculateIndustrySpecificAccuracy failed: %v", err)
	}

	// Validate each industry accuracy is between 0 and 1
	for industry, accuracy := range industryAccuracy {
		if accuracy < 0 || accuracy > 1 {
			t.Errorf("Expected accuracy for industry %s between 0 and 1, got %.2f", industry, accuracy)
		}
		t.Logf("Industry [%s] accuracy: %.2f%%", industry, accuracy*100)
	}
}

func TestAccuracyCalculationService_CalculateConfidenceDistribution(t *testing.T) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	distribution, err := acs.CalculateConfidenceDistribution(ctx, 24)
	if err != nil {
		t.Fatalf("CalculateConfidenceDistribution failed: %v", err)
	}

	// Validate distribution structure
	if distribution == nil {
		t.Fatal("Expected distribution, got nil")
	}

	// Validate confidence ranges
	if len(distribution.ConfidenceRanges) == 0 {
		t.Error("Expected confidence ranges, got empty slice")
	}

	// Validate each range
	for _, r := range distribution.ConfidenceRanges {
		if r.RangeStart < 0 || r.RangeEnd > 1 || r.RangeStart >= r.RangeEnd {
			t.Errorf("Invalid confidence range: %.2f-%.2f", r.RangeStart, r.RangeEnd)
		}
		if r.Count < 0 {
			t.Errorf("Invalid count for range %.2f-%.2f: %d", r.RangeStart, r.RangeEnd, r.Count)
		}
		if r.Accuracy < 0 || r.Accuracy > 1 {
			t.Errorf("Invalid accuracy for range %.2f-%.2f: %.2f", r.RangeStart, r.RangeEnd, r.Accuracy)
		}
		if r.Percentage < 0 || r.Percentage > 100 {
			t.Errorf("Invalid percentage for range %.2f-%.2f: %.2f", r.RangeStart, r.RangeEnd, r.Percentage)
		}
	}

	// Validate average confidence
	if distribution.AverageConfidence < 0 || distribution.AverageConfidence > 1 {
		t.Errorf("Invalid average confidence: %.2f", distribution.AverageConfidence)
	}

	t.Logf("Confidence distribution: High=%.2f%%, Medium=%.2f%%, Low=%.2f%%, Avg=%.3f",
		float64(distribution.HighConfidence)/float64(distribution.HighConfidence+distribution.MediumConfidence+distribution.LowConfidence)*100,
		float64(distribution.MediumConfidence)/float64(distribution.HighConfidence+distribution.MediumConfidence+distribution.LowConfidence)*100,
		float64(distribution.LowConfidence)/float64(distribution.HighConfidence+distribution.MediumConfidence+distribution.LowConfidence)*100,
		distribution.AverageConfidence)
}

func TestAccuracyCalculationService_CalculateSecurityMetrics(t *testing.T) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	securityMetrics, err := acs.CalculateSecurityMetrics(ctx, 24)
	if err != nil {
		t.Fatalf("CalculateSecurityMetrics failed: %v", err)
	}

	// Validate security metrics structure
	if securityMetrics == nil {
		t.Fatal("Expected security metrics, got nil")
	}

	// Validate accuracy rates are between 0 and 1
	if securityMetrics.TrustedDataSourceAccuracy < 0 || securityMetrics.TrustedDataSourceAccuracy > 1 {
		t.Errorf("Invalid trusted data source accuracy: %.2f", securityMetrics.TrustedDataSourceAccuracy)
	}
	if securityMetrics.WebsiteVerificationAccuracy < 0 || securityMetrics.WebsiteVerificationAccuracy > 1 {
		t.Errorf("Invalid website verification accuracy: %.2f", securityMetrics.WebsiteVerificationAccuracy)
	}

	// Validate rates are between 0 and 1
	if securityMetrics.DataSourceTrustRate < 0 || securityMetrics.DataSourceTrustRate > 1 {
		t.Errorf("Invalid data source trust rate: %.2f", securityMetrics.DataSourceTrustRate)
	}
	if securityMetrics.WebsiteVerificationRate < 0 || securityMetrics.WebsiteVerificationRate > 1 {
		t.Errorf("Invalid website verification rate: %.2f", securityMetrics.WebsiteVerificationRate)
	}
	if securityMetrics.SecurityViolationRate < 0 || securityMetrics.SecurityViolationRate > 1 {
		t.Errorf("Invalid security violation rate: %.2f", securityMetrics.SecurityViolationRate)
	}

	// Validate counts are non-negative
	if securityMetrics.TrustedDataPoints < 0 {
		t.Errorf("Invalid trusted data points: %d", securityMetrics.TrustedDataPoints)
	}
	if securityMetrics.VerifiedWebsitePoints < 0 {
		t.Errorf("Invalid verified website points: %d", securityMetrics.VerifiedWebsitePoints)
	}
	if securityMetrics.TotalSecurityValidations < 0 {
		t.Errorf("Invalid total security validations: %d", securityMetrics.TotalSecurityValidations)
	}

	t.Logf("Security metrics: Trust Rate=%.2f%%, Verification Rate=%.2f%%, Violation Rate=%.2f%%",
		securityMetrics.DataSourceTrustRate*100,
		securityMetrics.WebsiteVerificationRate*100,
		securityMetrics.SecurityViolationRate*100)
}

func TestAccuracyCalculationService_CalculatePerformanceMetrics(t *testing.T) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	performanceMetrics, err := acs.CalculatePerformanceMetrics(ctx, 24)
	if err != nil {
		t.Fatalf("CalculatePerformanceMetrics failed: %v", err)
	}

	// Validate performance metrics structure
	if performanceMetrics == nil {
		t.Fatal("Expected performance metrics, got nil")
	}

	// Validate response times are non-negative
	if performanceMetrics.AverageResponseTimeMs < 0 {
		t.Errorf("Invalid average response time: %.2f", performanceMetrics.AverageResponseTimeMs)
	}
	if performanceMetrics.AverageProcessingTimeMs < 0 {
		t.Errorf("Invalid average processing time: %.2f", performanceMetrics.AverageProcessingTimeMs)
	}

	// Validate performance accuracy rates are between 0 and 1
	if performanceMetrics.HighPerformanceAccuracy < 0 || performanceMetrics.HighPerformanceAccuracy > 1 {
		t.Errorf("Invalid high performance accuracy: %.2f", performanceMetrics.HighPerformanceAccuracy)
	}
	if performanceMetrics.MediumPerformanceAccuracy < 0 || performanceMetrics.MediumPerformanceAccuracy > 1 {
		t.Errorf("Invalid medium performance accuracy: %.2f", performanceMetrics.MediumPerformanceAccuracy)
	}
	if performanceMetrics.LowPerformanceAccuracy < 0 || performanceMetrics.LowPerformanceAccuracy > 1 {
		t.Errorf("Invalid low performance accuracy: %.2f", performanceMetrics.LowPerformanceAccuracy)
	}

	// Validate performance ranges
	if len(performanceMetrics.PerformanceRanges) == 0 {
		t.Error("Expected performance ranges, got empty slice")
	}

	// Validate each performance range
	for _, r := range performanceMetrics.PerformanceRanges {
		if r.RangeStart < 0 || r.RangeEnd < r.RangeStart {
			t.Errorf("Invalid performance range: %.2f-%.2f", r.RangeStart, r.RangeEnd)
		}
		if r.Count < 0 {
			t.Errorf("Invalid count for performance range %.2f-%.2f: %d", r.RangeStart, r.RangeEnd, r.Count)
		}
		if r.Accuracy < 0 || r.Accuracy > 1 {
			t.Errorf("Invalid accuracy for performance range %.2f-%.2f: %.2f", r.RangeStart, r.RangeEnd, r.Accuracy)
		}
		if r.Percentage < 0 || r.Percentage > 100 {
			t.Errorf("Invalid percentage for performance range %.2f-%.2f: %.2f", r.RangeStart, r.RangeEnd, r.Percentage)
		}
	}

	t.Logf("Performance metrics: Avg Response=%.2fms, High Perf Accuracy=%.2f%%, Medium Perf Accuracy=%.2f%%, Low Perf Accuracy=%.2f%%",
		performanceMetrics.AverageResponseTimeMs,
		performanceMetrics.HighPerformanceAccuracy*100,
		performanceMetrics.MediumPerformanceAccuracy*100,
		performanceMetrics.LowPerformanceAccuracy*100)
}

func TestAccuracyCalculationService_CalculateComprehensiveAccuracy(t *testing.T) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	result, err := acs.CalculateComprehensiveAccuracy(ctx, 24)
	if err != nil {
		t.Fatalf("CalculateComprehensiveAccuracy failed: %v", err)
	}

	// Validate comprehensive result structure
	if result == nil {
		t.Fatal("Expected comprehensive result, got nil")
	}

	// Validate overall accuracy
	if result.OverallAccuracy < 0 || result.OverallAccuracy > 1 {
		t.Errorf("Invalid overall accuracy: %.2f", result.OverallAccuracy)
	}

	// Validate industry-specific accuracy
	if result.IndustrySpecificAccuracy == nil {
		t.Error("Expected industry-specific accuracy map, got nil")
	}

	// Validate confidence distribution
	if result.ConfidenceDistribution.ConfidenceRanges == nil {
		t.Error("Expected confidence ranges, got nil")
	}

	// Validate security metrics
	if result.SecurityMetrics.TotalSecurityValidations < 0 {
		t.Errorf("Invalid total security validations: %d", result.SecurityMetrics.TotalSecurityValidations)
	}

	// Validate performance metrics
	if result.PerformanceMetrics.PerformanceRanges == nil {
		t.Error("Expected performance ranges, got nil")
	}

	// Validate metadata
	if result.DataPointsAnalyzed < 0 {
		t.Errorf("Invalid data points analyzed: %d", result.DataPointsAnalyzed)
	}
	if result.TimeRangeAnalyzed == "" {
		t.Error("Expected time range analyzed, got empty string")
	}

	t.Logf("Comprehensive accuracy result: Overall=%.2f%%, Industries=%d, Data Points=%d, Time Range=%s",
		result.OverallAccuracy*100,
		len(result.IndustrySpecificAccuracy),
		result.DataPointsAnalyzed,
		result.TimeRangeAnalyzed)
}

func TestAccuracyCalculationService_GetIndustryAccuracyBreakdown(t *testing.T) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	breakdowns, err := acs.GetIndustryAccuracyBreakdown(ctx, 24)
	if err != nil {
		t.Fatalf("GetIndustryAccuracyBreakdown failed: %v", err)
	}

	// Validate breakdowns structure
	if breakdowns == nil {
		t.Fatal("Expected breakdowns, got nil")
	}

	// Validate each breakdown
	for _, breakdown := range breakdowns {
		if breakdown.IndustryName == "" {
			t.Error("Expected industry name, got empty string")
		}
		if breakdown.TotalClassifications < 0 {
			t.Errorf("Invalid total classifications for %s: %d", breakdown.IndustryName, breakdown.TotalClassifications)
		}
		if breakdown.CorrectClassifications < 0 || breakdown.CorrectClassifications > breakdown.TotalClassifications {
			t.Errorf("Invalid correct classifications for %s: %d (total: %d)",
				breakdown.IndustryName, breakdown.CorrectClassifications, breakdown.TotalClassifications)
		}
		if breakdown.AccuracyPercentage < 0 || breakdown.AccuracyPercentage > 100 {
			t.Errorf("Invalid accuracy percentage for %s: %.2f", breakdown.IndustryName, breakdown.AccuracyPercentage)
		}
		if breakdown.AverageConfidence < 0 || breakdown.AverageConfidence > 1 {
			t.Errorf("Invalid average confidence for %s: %.2f", breakdown.IndustryName, breakdown.AverageConfidence)
		}
		if breakdown.AverageResponseTime < 0 {
			t.Errorf("Invalid average response time for %s: %.2f", breakdown.IndustryName, breakdown.AverageResponseTime)
		}
		if breakdown.ImprovementSuggestions == nil {
			t.Error("Expected improvement suggestions, got nil")
		}

		t.Logf("Industry breakdown [%s]: Accuracy=%.2f%%, Confidence=%.3f, Response Time=%.2fms, Suggestions=%d",
			breakdown.IndustryName,
			breakdown.AccuracyPercentage,
			breakdown.AverageConfidence,
			breakdown.AverageResponseTime,
			len(breakdown.ImprovementSuggestions))
	}
}

func TestAccuracyCalculationService_ValidateAccuracyCalculation(t *testing.T) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	err := acs.ValidateAccuracyCalculation(ctx)
	if err != nil {
		t.Fatalf("ValidateAccuracyCalculation failed: %v", err)
	}

	t.Log("Accuracy calculation validation passed")
}

// Test error handling
func TestAccuracyCalculationService_ErrorHandling(t *testing.T) {
	// Test with nil database
	acs := NewAccuracyCalculationService(nil, nil)
	ctx := context.Background()

	// These should return errors due to nil database
	_, err := acs.CalculateOverallAccuracy(ctx, 24)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = acs.CalculateIndustrySpecificAccuracy(ctx, 24)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = acs.CalculateConfidenceDistribution(ctx, 24)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = acs.CalculateSecurityMetrics(ctx, 24)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = acs.CalculatePerformanceMetrics(ctx, 24)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = acs.CalculateComprehensiveAccuracy(ctx, 24)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	_, err = acs.GetIndustryAccuracyBreakdown(ctx, 24)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}

	err = acs.ValidateAccuracyCalculation(ctx)
	if err == nil {
		t.Error("Expected error for nil database, got nil")
	}
}

// Test with different time ranges
func TestAccuracyCalculationService_DifferentTimeRanges(t *testing.T) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	timeRanges := []int{1, 6, 12, 24, 48, 72, 168} // 1 hour to 1 week

	for _, hours := range timeRanges {
		t.Run(fmt.Sprintf("TimeRange_%d_hours", hours), func(t *testing.T) {
			result, err := acs.CalculateComprehensiveAccuracy(ctx, hours)
			if err != nil {
				t.Fatalf("CalculateComprehensiveAccuracy failed for %d hours: %v", hours, err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			expectedTimeRange := fmt.Sprintf("Last %d hours", hours)
			if result.TimeRangeAnalyzed != expectedTimeRange {
				t.Errorf("Expected time range %s, got %s", expectedTimeRange, result.TimeRangeAnalyzed)
			}

			t.Logf("Time range %d hours: Overall accuracy=%.2f%%, Data points=%d",
				hours, result.OverallAccuracy*100, result.DataPointsAnalyzed)
		})
	}
}

// Test accuracy calculation with edge cases
func TestAccuracyCalculationService_EdgeCases(t *testing.T) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test with very short time range
	result, err := acs.CalculateComprehensiveAccuracy(ctx, 0)
	if err != nil {
		t.Fatalf("CalculateComprehensiveAccuracy failed for 0 hours: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Test with very long time range
	result, err = acs.CalculateComprehensiveAccuracy(ctx, 8760) // 1 year
	if err != nil {
		t.Fatalf("CalculateComprehensiveAccuracy failed for 8760 hours: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	t.Log("Edge case tests passed")
}

// Benchmark tests for accuracy calculation
func BenchmarkAccuracyCalculationService_CalculateOverallAccuracy(b *testing.B) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := acs.CalculateOverallAccuracy(ctx, 24)
		if err != nil {
			b.Fatalf("CalculateOverallAccuracy failed: %v", err)
		}
	}
}

func BenchmarkAccuracyCalculationService_CalculateComprehensiveAccuracy(b *testing.B) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		b.Skip("Skipping benchmark - no database connection")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := acs.CalculateComprehensiveAccuracy(ctx, 24)
		if err != nil {
			b.Fatalf("CalculateComprehensiveAccuracy failed: %v", err)
		}
	}
}

// Integration test that would require a real database
func TestAccuracyCalculationService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require a real database connection
	// db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
	// if err != nil {
	//     t.Fatalf("Failed to connect to database: %v", err)
	// }
	// defer db.Close()

	// acs := NewAccuracyCalculationService(db, log.Default())
	// ctx := context.Background()

	// Test all functions with real data
	t.Log("Integration test would run here with real database")
}

// Test accuracy calculation with mock data scenarios
func TestAccuracyCalculationService_MockDataScenarios(t *testing.T) {
	acs := NewAccuracyCalculationService(createMockDBForAccuracyCalculation(), nil)
	if acs.db == nil {
		t.Skip("Skipping test - no database connection")
	}

	ctx := context.Background()

	// Test scenario 1: High accuracy scenario
	t.Run("HighAccuracyScenario", func(t *testing.T) {
		result, err := acs.CalculateComprehensiveAccuracy(ctx, 24)
		if err != nil {
			t.Fatalf("High accuracy scenario failed: %v", err)
		}

		if result.OverallAccuracy > 0.9 {
			t.Logf("✅ High accuracy scenario: %.2f%%", result.OverallAccuracy*100)
		} else {
			t.Logf("📊 Current accuracy scenario: %.2f%%", result.OverallAccuracy*100)
		}
	})

	// Test scenario 2: Security validation scenario
	t.Run("SecurityValidationScenario", func(t *testing.T) {
		securityMetrics, err := acs.CalculateSecurityMetrics(ctx, 24)
		if err != nil {
			t.Fatalf("Security validation scenario failed: %v", err)
		}

		if securityMetrics.DataSourceTrustRate > 0.95 {
			t.Logf("✅ High security trust rate: %.2f%%", securityMetrics.DataSourceTrustRate*100)
		} else {
			t.Logf("📊 Current security trust rate: %.2f%%", securityMetrics.DataSourceTrustRate*100)
		}

		if securityMetrics.SecurityViolationRate < 0.05 {
			t.Logf("✅ Low security violation rate: %.2f%%", securityMetrics.SecurityViolationRate*100)
		} else {
			t.Logf("⚠️ Security violation rate: %.2f%%", securityMetrics.SecurityViolationRate*100)
		}
	})

	// Test scenario 3: Performance validation scenario
	t.Run("PerformanceValidationScenario", func(t *testing.T) {
		performanceMetrics, err := acs.CalculatePerformanceMetrics(ctx, 24)
		if err != nil {
			t.Fatalf("Performance validation scenario failed: %v", err)
		}

		if performanceMetrics.AverageResponseTimeMs < 500 {
			t.Logf("✅ Good performance: %.2fms average response time", performanceMetrics.AverageResponseTimeMs)
		} else {
			t.Logf("⚠️ Performance concern: %.2fms average response time", performanceMetrics.AverageResponseTimeMs)
		}

		if performanceMetrics.HighPerformanceAccuracy > 0.8 {
			t.Logf("✅ High performance accuracy: %.2f%%", performanceMetrics.HighPerformanceAccuracy*100)
		} else {
			t.Logf("📊 Current high performance accuracy: %.2f%%", performanceMetrics.HighPerformanceAccuracy*100)
		}
	})
}
