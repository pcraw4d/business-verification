//go:build integration

package integration

import (
	"context"
	"database/sql"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
)

// TestMerchantAnalyticsRepository_ErrorHandling tests error handling scenarios
func TestMerchantAnalyticsRepository_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewMerchantAnalyticsRepository(db, stdLogger)

	t.Run("GetClassificationByMerchantID with invalid merchant", func(t *testing.T) {
		result, err := repo.GetClassificationByMerchantID(ctx, "non-existent-merchant")
		if err == nil {
			t.Error("Expected error for non-existent merchant")
		}
		if result != nil {
			t.Error("Expected nil result for non-existent merchant")
		}
	})

	t.Run("GetSecurityDataByMerchantID with invalid merchant", func(t *testing.T) {
		result, err := repo.GetSecurityDataByMerchantID(ctx, "non-existent-merchant")
		if err == nil {
			t.Error("Expected error for non-existent merchant")
		}
		if result != nil {
			t.Error("Expected nil result for non-existent merchant")
		}
	})

	t.Run("GetQualityMetricsByMerchantID with invalid merchant", func(t *testing.T) {
		result, err := repo.GetQualityMetricsByMerchantID(ctx, "non-existent-merchant")
		if err == nil {
			t.Error("Expected error for non-existent merchant")
		}
		if result != nil {
			t.Error("Expected nil result for non-existent merchant")
		}
	})

	t.Run("GetIntelligenceDataByMerchantID with invalid merchant", func(t *testing.T) {
		result, err := repo.GetIntelligenceDataByMerchantID(ctx, "non-existent-merchant")
		if err == nil {
			t.Error("Expected error for non-existent merchant")
		}
		if result != nil {
			t.Error("Expected nil result for non-existent merchant")
		}
	})
}

// TestRiskIndicatorsRepository_ErrorHandling tests error handling scenarios
func TestRiskIndicatorsRepository_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewRiskIndicatorsRepository(db, stdLogger)

	t.Run("GetByMerchantID with invalid merchant", func(t *testing.T) {
		result, err := repo.GetByMerchantID(ctx, "non-existent-merchant", nil)
		if err != nil {
			t.Errorf("GetByMerchantID should return empty array, not error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected empty array for non-existent merchant, got %d items", len(result))
		}
	})

	t.Run("GetByMerchantID with invalid filters", func(t *testing.T) {
		merchantID := "merchant-filter-test"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		// Test with invalid severity
		filters := &database.RiskIndicatorFilters{
			Severity: "invalid-severity",
		}
		result, err := repo.GetByMerchantID(ctx, merchantID, filters)
		if err != nil {
			t.Errorf("GetByMerchantID should handle invalid filters gracefully: %v", err)
		}
		// Should return empty result or handle gracefully
		_ = result
	})
}

// TestRiskAssessmentRepository_ErrorHandling tests error handling scenarios
func TestRiskAssessmentRepository_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewRiskAssessmentRepository(db, stdLogger)

	t.Run("GetAssessmentByID with invalid ID", func(t *testing.T) {
		result, err := repo.GetAssessmentByID(ctx, "non-existent-assessment")
		if err == nil {
			t.Error("Expected error for non-existent assessment")
		}
		if result != nil {
			t.Error("Expected nil result for non-existent assessment")
		}
	})

	t.Run("GetAssessmentsByMerchantID with invalid merchant", func(t *testing.T) {
		result, err := repo.GetAssessmentsByMerchantID(ctx, "non-existent-merchant")
		if err != nil {
			t.Errorf("GetAssessmentsByMerchantID should return empty array, not error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected empty array for non-existent merchant, got %d items", len(result))
		}
	})

	t.Run("UpdateAssessmentStatus with invalid ID", func(t *testing.T) {
		err := repo.UpdateAssessmentStatus(ctx, "non-existent-assessment", models.AssessmentStatusCompleted, 100)
		if err == nil {
			t.Error("Expected error when updating non-existent assessment")
		}
	})

	t.Run("UpdateAssessmentResult with invalid ID", func(t *testing.T) {
		result := &models.RiskAssessmentResult{
			OverallScore: 0.7,
			RiskLevel:    "medium",
		}
		err := repo.UpdateAssessmentResult(ctx, "non-existent-assessment", result)
		if err == nil {
			t.Error("Expected error when updating non-existent assessment")
		}
	})
}

// TestMerchantAnalyticsRepository_ConcurrentAccess tests concurrent access
func TestMerchantAnalyticsRepository_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewMerchantAnalyticsRepository(db, stdLogger)

	merchantID := "merchant-concurrent-analytics"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	if err := SeedTestAnalytics(db, merchantID, nil, nil, nil, nil); err != nil {
		t.Fatalf("Failed to seed analytics: %v", err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.GetClassificationByMerchantID(ctx, merchantID)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	errorCount := len(errors)
	if errorCount > 0 {
		t.Errorf("Expected 0 errors from concurrent access, got %d", errorCount)
	}
}

// TestRiskAssessmentRepository_ConcurrentAccess tests concurrent access
func TestRiskAssessmentRepository_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewRiskAssessmentRepository(db, stdLogger)

	merchantID := "merchant-concurrent-assessment"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	assessmentID := "assessment-concurrent"
	if err := SeedTestRiskAssessment(db, merchantID, assessmentID, "processing", nil); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.GetAssessmentByID(ctx, assessmentID)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	errorCount := len(errors)
	if errorCount > 0 {
		t.Errorf("Expected 0 errors from concurrent access, got %d", errorCount)
	}
}

// TestRiskAssessmentRepository_StatusTransitions tests status transition logic
func TestRiskAssessmentRepository_StatusTransitions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewRiskAssessmentRepository(db, stdLogger)

	merchantID := "merchant-status-transitions"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	assessmentID := "assessment-status-1"
	if err := SeedTestRiskAssessment(db, merchantID, assessmentID, "pending", nil); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	// Test status transitions: pending -> processing -> completed
	statuses := []models.AssessmentStatus{
		models.AssessmentStatusProcessing,
		models.AssessmentStatusCompleted,
	}
	progresses := []int{50, 100}

	for i, status := range statuses {
		err := repo.UpdateAssessmentStatus(ctx, assessmentID, status, progresses[i])
		if err != nil {
			t.Fatalf("Failed to update status to %s: %v", status, err)
		}

		assessment, err := repo.GetAssessmentByID(ctx, assessmentID)
		if err != nil {
			t.Fatalf("Failed to get assessment: %v", err)
		}

		if assessment.Status != status {
			t.Errorf("Expected status %s, got %s", status, assessment.Status)
		}

		if assessment.Progress != progresses[i] {
			t.Errorf("Expected progress %d, got %d", progresses[i], assessment.Progress)
		}
	}
}

// TestRiskAssessmentRepository_ResultUpdate tests result update functionality
func TestRiskAssessmentRepository_ResultUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewRiskAssessmentRepository(db, stdLogger)

	merchantID := "merchant-result-update"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	assessmentID := "assessment-result-1"
	if err := SeedTestRiskAssessment(db, merchantID, assessmentID, "processing", nil); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	// Update result multiple times
	results := []*models.RiskAssessmentResult{
		{
			OverallScore: 0.6,
			RiskLevel:    "medium",
			Factors: []models.RiskFactor{
				{Name: "Factor1", Score: 0.7, Weight: 0.5},
			},
		},
		{
			OverallScore: 0.8,
			RiskLevel:    "high",
			Factors: []models.RiskFactor{
				{Name: "Factor1", Score: 0.9, Weight: 0.5},
				{Name: "Factor2", Score: 0.7, Weight: 0.5},
			},
		},
	}

	for i, result := range results {
		err := repo.UpdateAssessmentResult(ctx, assessmentID, result)
		if err != nil {
			t.Fatalf("Failed to update result (iteration %d): %v", i, err)
		}

		assessment, err := repo.GetAssessmentByID(ctx, assessmentID)
		if err != nil {
			t.Fatalf("Failed to get assessment: %v", err)
		}

		if assessment.Result == nil {
			t.Fatal("Expected result to be set")
		}

		if assessment.Result.OverallScore != result.OverallScore {
			t.Errorf("Expected overall score %f, got %f", result.OverallScore, assessment.Result.OverallScore)
		}

		if len(assessment.Result.Factors) != len(result.Factors) {
			t.Errorf("Expected %d factors, got %d", len(result.Factors), len(assessment.Result.Factors))
		}
	}
}

// TestRiskIndicatorsRepository_EdgeCases tests edge cases
func TestRiskIndicatorsRepository_EdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewRiskIndicatorsRepository(db, stdLogger)

	merchantID := "merchant-edge-cases"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	t.Run("GetByMerchantID with empty string filters", func(t *testing.T) {
		filters := &database.RiskIndicatorFilters{
			Severity: "",
			Status:   "",
		}
		result, err := repo.GetByMerchantID(ctx, merchantID, filters)
		if err != nil {
			t.Errorf("Expected no error with empty filters: %v", err)
		}
		_ = result
	})

	t.Run("GetByMerchantID with all filter combinations", func(t *testing.T) {
		if err := SeedTestRiskIndicators(db, merchantID, 10, ""); err != nil {
			t.Fatalf("Failed to seed indicators: %v", err)
		}

		severities := []string{"low", "medium", "high", "critical"}
		statuses := []string{"active", "resolved"}

		for _, severity := range severities {
			for _, status := range statuses {
				filters := &database.RiskIndicatorFilters{
					Severity: severity,
					Status:   status,
				}
				result, err := repo.GetByMerchantID(ctx, merchantID, filters)
				if err != nil {
					t.Errorf("Error with filters severity=%s, status=%s: %v", severity, status, err)
				}
				// Verify all returned indicators match filters
				for _, indicator := range result {
					if filters.Severity != "" && indicator.Severity != filters.Severity {
						t.Errorf("Indicator severity %s does not match filter %s", indicator.Severity, filters.Severity)
					}
					if filters.Status != "" && indicator.Status != filters.Status {
						t.Errorf("Indicator status %s does not match filter %s", indicator.Status, filters.Status)
					}
				}
			}
		}
	})
}

// TestMerchantAnalyticsRepository_EmptyData tests handling of empty/null JSONB fields
func TestMerchantAnalyticsRepository_EmptyData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	testDB, err := SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	repo := database.NewMerchantAnalyticsRepository(db, stdLogger)

	merchantID := "merchant-empty-data"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	// Seed analytics with empty/null JSONB fields
	if err := SeedTestAnalytics(db, merchantID, nil, nil, nil, nil); err != nil {
		t.Fatalf("Failed to seed analytics: %v", err)
	}

	t.Run("GetClassificationByMerchantID with empty data", func(t *testing.T) {
		result, err := repo.GetClassificationByMerchantID(ctx, merchantID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected classification data but got nil")
		}
	})

	t.Run("GetSecurityDataByMerchantID with empty data", func(t *testing.T) {
		result, err := repo.GetSecurityDataByMerchantID(ctx, merchantID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected security data but got nil")
		}
	})

	t.Run("GetQualityMetricsByMerchantID with empty data", func(t *testing.T) {
		result, err := repo.GetQualityMetricsByMerchantID(ctx, merchantID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected quality data but got nil")
		}
	})

	t.Run("GetIntelligenceDataByMerchantID with empty data", func(t *testing.T) {
		result, err := repo.GetIntelligenceDataByMerchantID(ctx, merchantID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected intelligence data but got nil")
		}
	})
}

