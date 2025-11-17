//go:build integration

package services

import (
	"context"
	"log"
	"os"
	"testing"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
	integration "kyb-platform/test/integration"
)

func TestRiskIndicatorsService_GetRiskIndicators(t *testing.T) {
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

	// Setup repository and service
	indicatorsRepo := database.NewRiskIndicatorsRepository(db, stdLogger)
	service := NewRiskIndicatorsService(indicatorsRepo, stdLogger)

	tests := []struct {
		name       string
		merchantID string
		filters    *database.RiskIndicatorFilters
		setup      func() error
		cleanup    func() error
		wantErr    bool
		validate   func(*testing.T, *models.RiskIndicatorsData)
	}{
		{
			name:       "successful fetch without filters",
			merchantID: "merchant-indicators-svc-1",
			filters:    nil,
			setup: func() error {
				if err := integration.SeedTestMerchant(db, "merchant-indicators-svc-1", "active"); err != nil {
					return err
				}
				return integration.SeedTestRiskIndicators(db, "merchant-indicators-svc-1", 5, "")
			},
			cleanup: func() error {
				return integration.CleanupTestData(db, "merchant-indicators-svc-1")
			},
			wantErr: false,
			validate: func(t *testing.T, data *models.RiskIndicatorsData) {
				if data == nil {
					t.Fatal("Expected risk indicators data but got nil")
				}
				if len(data.Indicators) != 5 {
					t.Errorf("Expected 5 indicators, got %d", len(data.Indicators))
				}
				if data.OverallScore < 0 || data.OverallScore > 1 {
					t.Errorf("Expected overall score between 0 and 1, got %f", data.OverallScore)
				}
			},
		},
		{
			name:       "filter by severity",
			merchantID: "merchant-indicators-svc-2",
			filters:    &database.RiskIndicatorFilters{Severity: "high"},
			setup: func() error {
				if err := integration.SeedTestMerchant(db, "merchant-indicators-svc-2", "active"); err != nil {
					return err
				}
				return integration.SeedTestRiskIndicators(db, "merchant-indicators-svc-2", 10, "high")
			},
			cleanup: func() error {
				return integration.CleanupTestData(db, "merchant-indicators-svc-2")
			},
			wantErr: false,
			validate: func(t *testing.T, data *models.RiskIndicatorsData) {
				if len(data.Indicators) != 10 {
					t.Errorf("Expected 10 indicators, got %d", len(data.Indicators))
				}
				// Verify all indicators have high severity
				for _, indicator := range data.Indicators {
					if indicator.Severity != "high" {
						t.Errorf("Expected all indicators to have high severity, got %s", indicator.Severity)
					}
				}
			},
		},
		{
			name:       "empty result set",
			merchantID: "merchant-indicators-svc-3",
			filters:    nil,
			setup: func() error {
				return integration.SeedTestMerchant(db, "merchant-indicators-svc-3", "active")
			},
			cleanup: func() error {
				return integration.CleanupTestData(db, "merchant-indicators-svc-3")
			},
			wantErr: false,
			validate: func(t *testing.T, data *models.RiskIndicatorsData) {
				if len(data.Indicators) != 0 {
					t.Errorf("Expected 0 indicators, got %d", len(data.Indicators))
				}
				if data.OverallScore != 0.0 {
					t.Errorf("Expected overall score 0.0 for empty indicators, got %f", data.OverallScore)
				}
			},
		},
		{
			name:       "merchant not found",
			merchantID: "invalid-merchant",
			filters:    nil,
			setup:      func() error { return nil },
			cleanup:    func() error { return nil },
			wantErr:    true,
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

			result, err := service.GetRiskIndicators(ctx, tt.merchantID, tt.filters)

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

func TestRiskIndicatorsService_CalculateOverallScore(t *testing.T) {
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

	// Setup repository and service
	indicatorsRepo := database.NewRiskIndicatorsRepository(db, stdLogger)
	service := NewRiskIndicatorsService(indicatorsRepo, stdLogger)

	merchantID := "merchant-score-test"
	if err := integration.SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer integration.CleanupTestData(db, merchantID)

	// Seed indicators with different severities
	if err := integration.SeedTestRiskIndicators(db, merchantID, 4, ""); err != nil {
		t.Fatalf("Failed to seed indicators: %v", err)
	}

	result, err := service.GetRiskIndicators(ctx, merchantID, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify score calculation
	if result.OverallScore <= 0 {
		t.Error("Expected overall score to be greater than 0")
	}

	if result.OverallScore > 1 {
		t.Errorf("Expected overall score to be <= 1, got %f", result.OverallScore)
	}
}

