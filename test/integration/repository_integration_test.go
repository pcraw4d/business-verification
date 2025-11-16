//go:build integration

package integration

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
)

func TestMerchantAnalyticsRepository_GetClassificationByMerchantID(t *testing.T) {
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

	tests := []struct {
		name       string
		merchantID string
		setup      func() error
		cleanup    func() error
		wantErr    bool
		validate   func(*testing.T, *models.ClassificationData)
	}{
		{
			name:       "successful retrieval",
			merchantID: "merchant-class-1",
			setup: func() error {
				if err := SeedTestMerchant(db, "merchant-class-1", "active"); err != nil {
					return err
				}
				return SeedTestAnalytics(db, "merchant-class-1",
					map[string]interface{}{
						"primaryIndustry": "Technology",
						"confidenceScore": 0.95,
						"riskLevel":       "low",
					}, nil, nil, nil)
			},
			cleanup: func() error {
				return CleanupTestData(db, "merchant-class-1")
			},
			wantErr: false,
			validate: func(t *testing.T, data *models.ClassificationData) {
				if data == nil {
					t.Fatal("Expected classification data but got nil")
				}
				if data.PrimaryIndustry == "" {
					t.Error("Expected primary industry to be set")
				}
			},
		},
		{
			name:       "merchant not found",
			merchantID: "invalid-merchant",
			setup:      func() error { return nil },
			cleanup:    func() error { return nil },
			wantErr:    true,
		},
		{
			name:       "empty JSONB fields",
			merchantID: "merchant-class-2",
			setup: func() error {
				if err := SeedTestMerchant(db, "merchant-class-2", "active"); err != nil {
					return err
				}
				return SeedTestAnalytics(db, "merchant-class-2", nil, nil, nil, nil)
			},
			cleanup: func() error {
				return CleanupTestData(db, "merchant-class-2")
			},
			wantErr: false,
			validate: func(t *testing.T, data *models.ClassificationData) {
				if data == nil {
					t.Fatal("Expected classification data but got nil")
				}
				// Should return empty struct, not error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			if tt.cleanup != nil {
				defer func() {
					if err := tt.cleanup(); err != nil {
						t.Logf("Cleanup failed: %v", err)
					}
				}()
			}

			result, err := repo.GetClassificationByMerchantID(ctx, tt.merchantID)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestMerchantAnalyticsRepository_GetSecurityDataByMerchantID(t *testing.T) {
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

	merchantID := "merchant-security-1"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	result, err := repo.GetSecurityDataByMerchantID(ctx, merchantID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected security data but got nil")
	}

	if result.TrustScore <= 0 {
		t.Error("Expected trust score to be set")
	}
}

func TestMerchantAnalyticsRepository_GetQualityMetricsByMerchantID(t *testing.T) {
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

	merchantID := "merchant-quality-1"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	result, err := repo.GetQualityMetricsByMerchantID(ctx, merchantID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected quality data but got nil")
	}

	if result.CompletenessScore < 0 || result.CompletenessScore > 1 {
		t.Errorf("Expected completeness score between 0 and 1, got %f", result.CompletenessScore)
	}
}

func TestMerchantAnalyticsRepository_GetIntelligenceDataByMerchantID(t *testing.T) {
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

	merchantID := "merchant-intel-1"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	result, err := repo.GetIntelligenceDataByMerchantID(ctx, merchantID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected intelligence data but got nil")
	}
}

func TestRiskIndicatorsRepository_GetByMerchantID(t *testing.T) {
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

	tests := []struct {
		name       string
		merchantID string
		filters    *database.RiskIndicatorFilters
		setup      func() error
		cleanup    func() error
		wantCount  int
	}{
		{
			name:       "successful retrieval without filters",
			merchantID: "merchant-indicators-1",
			filters:    nil,
			setup: func() error {
				if err := SeedTestMerchant(db, "merchant-indicators-1", "active"); err != nil {
					return err
				}
				return SeedTestRiskIndicators(db, "merchant-indicators-1", 5, "")
			},
			cleanup: func() error {
				return CleanupTestData(db, "merchant-indicators-1")
			},
			wantCount: 5,
		},
		{
			name:       "filter by severity",
			merchantID: "merchant-indicators-2",
			filters:    &database.RiskIndicatorFilters{Severity: "high"},
			setup: func() error {
				if err := SeedTestMerchant(db, "merchant-indicators-2", "active"); err != nil {
					return err
				}
				return SeedTestRiskIndicators(db, "merchant-indicators-2", 10, "high")
			},
			cleanup: func() error {
				return CleanupTestData(db, "merchant-indicators-2")
			},
			wantCount: 10,
		},
		{
			name:       "filter by status",
			merchantID: "merchant-indicators-3",
			filters:    &database.RiskIndicatorFilters{Status: "active"},
			setup: func() error {
				if err := SeedTestMerchant(db, "merchant-indicators-3", "active"); err != nil {
					return err
				}
				return SeedTestRiskIndicators(db, "merchant-indicators-3", 5, "")
			},
			cleanup: func() error {
				return CleanupTestData(db, "merchant-indicators-3")
			},
			wantCount: 4, // Some will be resolved based on seed logic
		},
		{
			name:       "combined filters",
			merchantID: "merchant-indicators-4",
			filters:    &database.RiskIndicatorFilters{Severity: "medium", Status: "active"},
			setup: func() error {
				if err := SeedTestMerchant(db, "merchant-indicators-4", "active"); err != nil {
					return err
				}
				return SeedTestRiskIndicators(db, "merchant-indicators-4", 5, "medium")
			},
			cleanup: func() error {
				return CleanupTestData(db, "merchant-indicators-4")
			},
			wantCount: 4, // Some will be resolved
		},
		{
			name:       "empty result set",
			merchantID: "merchant-indicators-5",
			filters:    nil,
			setup: func() error {
				return SeedTestMerchant(db, "merchant-indicators-5", "active")
			},
			cleanup: func() error {
				return CleanupTestData(db, "merchant-indicators-5")
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			if tt.cleanup != nil {
				defer func() {
					if err := tt.cleanup(); err != nil {
						t.Logf("Cleanup failed: %v", err)
					}
				}()
			}

			result, err := repo.GetByMerchantID(ctx, tt.merchantID, tt.filters)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != tt.wantCount {
				t.Errorf("Expected %d indicators, got %d", tt.wantCount, len(result))
			}
		})
	}
}

func TestRiskAssessmentRepository_CreateAssessment(t *testing.T) {
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

	merchantID := "merchant-repo-1"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	assessment := &models.RiskAssessment{
		ID:         "assessment-repo-1",
		MerchantID: merchantID,
		Status:     models.AssessmentStatusPending,
		Progress:   0,
	}

	err = repo.CreateAssessment(ctx, assessment)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify it was created
	retrieved, err := repo.GetAssessmentByID(ctx, "assessment-repo-1")
	if err != nil {
		t.Fatalf("Failed to retrieve created assessment: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected assessment but got nil")
	}

	if retrieved.ID != "assessment-repo-1" {
		t.Errorf("Expected ID assessment-repo-1, got %s", retrieved.ID)
	}
}

func TestRiskAssessmentRepository_GetAssessmentByID(t *testing.T) {
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

	merchantID := "merchant-repo-2"
	assessmentID := "assessment-repo-2"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	if err := SeedTestRiskAssessment(db, merchantID, assessmentID, "completed", nil); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	result, err := repo.GetAssessmentByID(ctx, assessmentID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected assessment but got nil")
	}

	if result.ID != assessmentID {
		t.Errorf("Expected ID %s, got %s", assessmentID, result.ID)
	}
}

func TestRiskAssessmentRepository_GetAssessmentsByMerchantID(t *testing.T) {
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

	merchantID := "merchant-repo-3"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	// Seed multiple assessments
	for i := 1; i <= 3; i++ {
		assessmentID := "assessment-repo-3-" + string(rune('0'+i))
		if err := SeedTestRiskAssessment(db, merchantID, assessmentID, "completed", nil); err != nil {
			t.Fatalf("Failed to seed assessment %d: %v", i, err)
		}
	}

	result, err := repo.GetAssessmentsByMerchantID(ctx, merchantID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 assessments, got %d", len(result))
	}
}

func TestRiskAssessmentRepository_UpdateAssessmentStatus(t *testing.T) {
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

	merchantID := "merchant-repo-4"
	assessmentID := "assessment-repo-4"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	if err := SeedTestRiskAssessment(db, merchantID, assessmentID, "pending", nil); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	// Update status through lifecycle
	err = repo.UpdateAssessmentStatus(ctx, assessmentID, models.AssessmentStatusProcessing, 50)
	if err != nil {
		t.Fatalf("Failed to update status to processing: %v", err)
	}

	err = repo.UpdateAssessmentStatus(ctx, assessmentID, models.AssessmentStatusCompleted, 100)
	if err != nil {
		t.Fatalf("Failed to update status to completed: %v", err)
	}

	// Verify final status
	assessment, err := repo.GetAssessmentByID(ctx, assessmentID)
	if err != nil {
		t.Fatalf("Failed to retrieve assessment: %v", err)
	}

	if assessment.Status != models.AssessmentStatusCompleted {
		t.Errorf("Expected status completed, got %s", assessment.Status)
	}

	if assessment.Progress != 100 {
		t.Errorf("Expected progress 100, got %d", assessment.Progress)
	}
}

func TestRiskAssessmentRepository_UpdateAssessmentResult(t *testing.T) {
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

	merchantID := "merchant-repo-5"
	assessmentID := "assessment-repo-5"
	if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer CleanupTestData(db, merchantID)

	if err := SeedTestRiskAssessment(db, merchantID, assessmentID, "processing", nil); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	result := &models.RiskAssessmentResult{
		OverallScore: 0.75,
		RiskLevel:    "medium",
		Factors: []models.RiskFactor{
			{Name: "Factor1", Score: 0.8, Weight: 0.5},
			{Name: "Factor2", Score: 0.7, Weight: 0.5},
		},
	}

	err = repo.UpdateAssessmentResult(ctx, assessmentID, result)
	if err != nil {
		t.Fatalf("Failed to update result: %v", err)
	}

	// Verify result was saved
	assessment, err := repo.GetAssessmentByID(ctx, assessmentID)
	if err != nil {
		t.Fatalf("Failed to retrieve assessment: %v", err)
	}

	if assessment.Result == nil {
		t.Fatal("Expected result but got nil")
	}

	if assessment.Result.OverallScore != 0.75 {
		t.Errorf("Expected overall score 0.75, got %f", assessment.Result.OverallScore)
	}
}

