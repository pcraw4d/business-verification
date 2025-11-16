//go:build integration

package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
	integration "kyb-platform/test/integration"
)

// mockRiskAssessmentRepository is a mock implementation of RiskAssessmentRepository
type mockRiskAssessmentRepository struct {
	assessments []*models.RiskAssessment
	assessment  *models.RiskAssessment
	err         error
}

func (m *mockRiskAssessmentRepository) CreateAssessment(ctx context.Context, merchantID string, options models.AssessmentOptions) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "assessment-123", nil
}

func (m *mockRiskAssessmentRepository) GetAssessmentByID(ctx context.Context, assessmentID string) (*models.RiskAssessment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.assessment, nil
}

func (m *mockRiskAssessmentRepository) GetAssessmentsByMerchantID(ctx context.Context, merchantID string) ([]*models.RiskAssessment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.assessments, nil
}

func (m *mockRiskAssessmentRepository) UpdateAssessmentStatus(ctx context.Context, assessmentID string, status models.AssessmentStatus, progress int) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *mockRiskAssessmentRepository) UpdateAssessmentResult(ctx context.Context, assessmentID string, result *models.RiskAssessmentResult) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func TestRiskAssessmentService_GetRiskHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database
	testDB, err := integration.SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	// Setup repository
	repo := database.NewRiskAssessmentRepository(db, stdLogger)
	service := NewRiskAssessmentService(repo, nil, stdLogger)

	tests := []struct {
		name       string
		merchantID string
		limit      int
		offset     int
		setup      func() error
		cleanup    func() error
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "successful fetch with pagination",
			merchantID: "merchant-history-1",
			limit:      10,
			offset:     0,
			setup: func() error {
				// Seed merchant
				if err := integration.SeedTestMerchant(db, "merchant-history-1", "active"); err != nil {
					return err
				}
				// Seed multiple assessments
				for i := 1; i <= 2; i++ {
					if err := integration.SeedTestRiskAssessment(db, "merchant-history-1", 
						fmt.Sprintf("assessment-%d", i), "completed", 
						map[string]interface{}{
							"overallScore": 0.7,
							"riskLevel":    "medium",
						}); err != nil {
						return err
					}
				}
				return nil
			},
			cleanup: func() error {
				return integration.CleanupTestData(db, "merchant-history-1")
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:       "pagination with offset",
			merchantID: "merchant-history-2",
			limit:      10,
			offset:     5,
			setup: func() error {
				if err := integration.SeedTestMerchant(db, "merchant-history-2", "active"); err != nil {
					return err
				}
				// Seed 7 assessments
				for i := 1; i <= 7; i++ {
					if err := integration.SeedTestRiskAssessment(db, "merchant-history-2",
						fmt.Sprintf("assessment-%d", i), "completed", nil); err != nil {
						return err
					}
				}
				return nil
			},
			cleanup: func() error {
				return integration.CleanupTestData(db, "merchant-history-2")
			},
			wantErr:   false,
			wantCount: 2, // offset 5, limit 10, so should return 2 items (6th and 7th)
		},
		{
			name:       "empty history",
			merchantID: "merchant-history-3",
			limit:      10,
			offset:     0,
			setup: func() error {
				return integration.SeedTestMerchant(db, "merchant-history-3", "active")
			},
			cleanup: func() error {
				return integration.CleanupTestData(db, "merchant-history-3")
			},
			wantErr:   false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test data
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Cleanup after test
			if tt.cleanup != nil {
				defer func() {
					if err := tt.cleanup(); err != nil {
						t.Logf("Cleanup failed: %v", err)
					}
				}()
			}

			result, err := service.GetRiskHistory(ctx, tt.merchantID, tt.limit, tt.offset)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != tt.wantCount {
				t.Errorf("Expected %d assessments, got %d", tt.wantCount, len(result))
			}
		})
	}
}

func TestRiskAssessmentService_GetPredictions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database
	testDB, err := integration.SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	// Setup repository
	repo := database.NewRiskAssessmentRepository(db, stdLogger)
	service := NewRiskAssessmentService(repo, nil, stdLogger)

	// Seed test data
	merchantID := "merchant-predictions"
	if err := integration.SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer integration.CleanupTestData(db, merchantID)

	// Seed assessment with result
	if err := integration.SeedTestRiskAssessment(db, merchantID, "assessment-1", "completed",
		map[string]interface{}{
			"overallScore": 0.7,
			"riskLevel":    "medium",
		}); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	result, err := service.GetPredictions(ctx, merchantID, []int{3, 6, 12}, true, true)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	// Verify structure
	if result["merchantId"] != merchantID {
		t.Errorf("Expected merchantId %s, got %v", merchantID, result["merchantId"])
	}

	predictions, ok := result["predictions"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected predictions array")
	}

	if len(predictions) != 3 {
		t.Errorf("Expected 3 predictions, got %d", len(predictions))
	}
}

func TestRiskAssessmentService_ExplainAssessment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database
	testDB, err := integration.SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	// Setup repository
	repo := database.NewRiskAssessmentRepository(db, stdLogger)
	service := NewRiskAssessmentService(repo, nil, stdLogger)

	// Seed test data
	merchantID := "merchant-explain"
	assessmentID := "assessment-123"
	if err := integration.SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer integration.CleanupTestData(db, merchantID)

	// Seed assessment with result containing factors
	if err := integration.SeedTestRiskAssessment(db, merchantID, assessmentID, "completed",
		map[string]interface{}{
			"overallScore": 0.7,
			"riskLevel":    "medium",
			"factors": []map[string]interface{}{
				{"name": "Factor1", "score": 0.8, "weight": 0.5},
				{"name": "Factor2", "score": 0.6, "weight": 0.5},
			},
		}); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	result, err := service.ExplainAssessment(ctx, assessmentID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	if result["assessmentId"] != assessmentID {
		t.Errorf("Expected assessmentId %s, got %v", assessmentID, result["assessmentId"])
	}

	prediction, ok := result["prediction"].(float64)
	if !ok {
		t.Error("Expected prediction to be a float64")
	} else if prediction != 0.7 {
		t.Errorf("Expected prediction 0.7, got %v", prediction)
	}

	factors, ok := result["factors"].([]interface{})
	if !ok {
		t.Error("Expected factors to be an array")
	} else if len(factors) < 1 {
		t.Errorf("Expected at least 1 factor, got %d", len(factors))
	}
}

func TestRiskAssessmentService_GetRecommendations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database
	testDB, err := integration.SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	// Setup repository
	repo := database.NewRiskAssessmentRepository(db, stdLogger)
	service := NewRiskAssessmentService(repo, nil, stdLogger)

	// Seed test data
	merchantID := "merchant-recommendations"
	if err := integration.SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer integration.CleanupTestData(db, merchantID)

	// Seed assessment with high risk score
	if err := integration.SeedTestRiskAssessment(db, merchantID, "assessment-1", "completed",
		map[string]interface{}{
			"overallScore": 0.8, // High risk
			"riskLevel":    "high",
			"factors": []map[string]interface{}{
				{"name": "Factor1", "score": 0.7, "weight": 0.5},
			},
		}); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	result, err := service.GetRecommendations(ctx, merchantID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	if len(result) == 0 {
		t.Error("Expected at least one recommendation")
	}

	// Verify recommendation structure
	rec := result[0]
	if rec["merchantId"] != merchantID {
		t.Errorf("Expected merchantId %s in recommendation, got %v", merchantID, rec["merchantId"])
	}
}

func TestRiskAssessmentService_StartAssessment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database
	testDB, err := integration.SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	// Setup repository
	repo := database.NewRiskAssessmentRepository(db, stdLogger)
	// Service without job queue (will fail on enqueue, but we can test creation)
	service := NewRiskAssessmentService(repo, nil, stdLogger)

	// Seed merchant
	merchantID := "merchant-start"
	if err := integration.SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer integration.CleanupTestData(db, merchantID)

	// Start assessment (will fail on enqueue but assessment should be created)
	options := models.AssessmentOptions{
		IncludeHistorical: true,
		IncludePredictions: true,
	}
	assessmentID, err := service.StartAssessment(ctx, merchantID, options)

	// We expect an error because job queue is nil, but the assessment should be created
	if err == nil {
		t.Log("Assessment started successfully (job queue not available)")
	} else {
		// Check if assessment was created despite enqueue failure
		assessment, getErr := repo.GetAssessmentByID(ctx, assessmentID)
		if getErr == nil && assessment != nil {
			t.Logf("Assessment created successfully despite enqueue error: %s", assessmentID)
		} else {
			t.Errorf("Assessment not created: %v", getErr)
		}
	}
}

func TestRiskAssessmentService_GetAssessmentStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database
	testDB, err := integration.SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	// Setup repository
	repo := database.NewRiskAssessmentRepository(db, stdLogger)
	service := NewRiskAssessmentService(repo, nil, stdLogger)

	// Seed test data
	merchantID := "merchant-status"
	assessmentID := "assessment-status-1"
	if err := integration.SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer integration.CleanupTestData(db, merchantID)

	// Seed assessment with status
	if err := integration.SeedTestRiskAssessment(db, merchantID, assessmentID, "processing", nil); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	// Get status
	status, err := service.GetAssessmentStatus(ctx, assessmentID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if status == nil {
		t.Fatal("Expected status but got nil")
	}

	if status.AssessmentID != assessmentID {
		t.Errorf("Expected assessment ID %s, got %s", assessmentID, status.AssessmentID)
	}

	if status.Status != "processing" {
		t.Errorf("Expected status processing, got %s", status.Status)
	}
}

func TestRiskAssessmentService_ProcessAssessment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database
	testDB, err := integration.SetupTestDatabase()
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer testDB.CleanupTestDatabase()

	db := testDB.GetDB()
	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	// Setup repository
	repo := database.NewRiskAssessmentRepository(db, stdLogger)
	service := NewRiskAssessmentService(repo, nil, stdLogger)

	// Seed test data
	merchantID := "merchant-process"
	assessmentID := "assessment-process-1"
	if err := integration.SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer integration.CleanupTestData(db, merchantID)

	// Seed assessment
	if err := integration.SeedTestRiskAssessment(db, merchantID, assessmentID, "pending", nil); err != nil {
		t.Fatalf("Failed to seed assessment: %v", err)
	}

	// Process assessment (this is a no-op in the current implementation)
	err = service.ProcessAssessment(ctx, assessmentID)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify assessment still exists
	assessment, err := repo.GetAssessmentByID(ctx, assessmentID)
	if err != nil {
		t.Fatalf("Failed to get assessment: %v", err)
	}

	if assessment == nil {
		t.Fatal("Expected assessment but got nil")
	}
}

