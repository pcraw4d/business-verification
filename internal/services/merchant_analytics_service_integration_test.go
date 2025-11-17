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

func TestMerchantAnalyticsService_GetWebsiteAnalysis(t *testing.T) {
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

	// Setup repositories
	analyticsRepo := database.NewMerchantAnalyticsRepository(db, stdLogger)
	merchantRepo := database.NewMerchantPortfolioRepository(db, stdLogger)
	service := NewMerchantAnalyticsService(analyticsRepo, merchantRepo, nil, stdLogger)

	tests := []struct {
		name       string
		merchantID string
		setup      func() error
		cleanup    func() error
		wantErr    bool
		validate   func(*testing.T, *models.WebsiteAnalysisData)
	}{
		{
			name:       "successful fetch with website",
			merchantID: "merchant-website-1",
			setup: func() error {
				// Seed merchant with website
				if err := integration.SeedTestMerchant(db, "merchant-website-1", "active"); err != nil {
					return err
				}
				// Seed security data
				return integration.SeedTestAnalytics(db, "merchant-website-1", nil,
					map[string]interface{}{
						"trustScore": 0.8,
						"sslValid":   true,
					}, nil, nil)
			},
			cleanup: func() error {
				return integration.CleanupTestData(db, "merchant-website-1")
			},
			wantErr: false,
			validate: func(t *testing.T, data *models.WebsiteAnalysisData) {
				if data == nil {
					t.Fatal("Expected website analysis data but got nil")
				}
				if data.MerchantID != "merchant-website-1" {
					t.Errorf("Expected merchant ID merchant-website-1, got %s", data.MerchantID)
				}
				if data.WebsiteURL == "" {
					t.Error("Expected website URL to be set")
				}
			},
		},
		{
			name:       "merchant without website",
			merchantID: "merchant-no-website",
			setup: func() error {
				// Seed merchant without website (would need to update seed function)
				return integration.SeedTestMerchant(db, "merchant-no-website", "active")
			},
			cleanup: func() error {
				return integration.CleanupTestData(db, "merchant-no-website")
			},
			wantErr: true,
		},
		{
			name:       "merchant not found",
			merchantID: "invalid-merchant",
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

			result, err := service.GetWebsiteAnalysis(ctx, tt.merchantID)

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

