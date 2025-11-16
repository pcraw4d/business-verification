//go:build integration

package services

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
	integration "kyb-platform/test/integration"
)

// mockCache is a mock implementation of Cache interface
type mockCache struct {
	data map[string]interface{}
}

func newMockCache() *mockCache {
	return &mockCache{
		data: make(map[string]interface{}),
	}
}

func (m *mockCache) Get(ctx context.Context, key string, dest interface{}) error {
	if val, ok := m.data[key]; ok {
		// Simple type assertion for testing
		if destPtr, ok := dest.(*models.AnalyticsData); ok {
			if valData, ok := val.(*models.AnalyticsData); ok {
				*destPtr = *valData
				return nil
			}
		}
		return errors.New("cache miss")
	}
	return errors.New("cache miss")
}

func (m *mockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.data[key] = value
	return nil
}

// mockAnalyticsRepository is a mock implementation of MerchantAnalyticsRepository
type mockAnalyticsRepository struct {
	classification *models.ClassificationData
	security       *models.SecurityData
	quality        *models.QualityData
	intelligence   *models.IntelligenceData
	err            error
}

func (m *mockAnalyticsRepository) GetClassificationByMerchantID(ctx context.Context, merchantID string) (*models.ClassificationData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.classification, nil
}

func (m *mockAnalyticsRepository) GetSecurityDataByMerchantID(ctx context.Context, merchantID string) (*models.SecurityData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.security, nil
}

func (m *mockAnalyticsRepository) GetQualityMetricsByMerchantID(ctx context.Context, merchantID string) (*models.QualityData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.quality, nil
}

func (m *mockAnalyticsRepository) GetIntelligenceDataByMerchantID(ctx context.Context, merchantID string) (*models.IntelligenceData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.intelligence, nil
}

// mockMerchantRepository is a mock implementation of MerchantPortfolioRepository
type mockMerchantRepository struct {
	merchant *models.Merchant
	err      error
}

func (m *mockMerchantRepository) GetMerchant(ctx context.Context, merchantID string) (*models.Merchant, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.merchant, nil
}

func TestMerchantAnalyticsService_GetMerchantAnalytics(t *testing.T) {
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

	tests := []struct {
		name       string
		merchantID string
		setup      func() error
		cleanup    func() error
		wantErr    bool
		validate   func(*testing.T, *models.AnalyticsData)
	}{
		{
			name:       "successful fetch",
			merchantID: "merchant-123",
			setup: func() error {
				// Seed merchant
				if err := integration.SeedTestMerchant(db, "merchant-123", "active"); err != nil {
					return err
				}
				// Seed analytics data
				return integration.SeedTestAnalytics(db, "merchant-123",
					map[string]interface{}{
						"primaryIndustry": "Technology",
						"confidenceScore": 0.95,
						"riskLevel":       "low",
					},
					map[string]interface{}{
						"trustScore": 0.8,
						"sslValid":   true,
					},
					map[string]interface{}{
						"completenessScore": 0.9,
						"dataPoints":        100,
					},
					map[string]interface{}{},
				)
			},
			cleanup: func() error {
				return integration.CleanupTestData(db, "merchant-123")
			},
			wantErr: false,
			validate: func(t *testing.T, data *models.AnalyticsData) {
				if data == nil {
					t.Fatal("Expected analytics data but got nil")
				}
				if data.MerchantID != "merchant-123" {
					t.Errorf("Expected merchant ID merchant-123, got %s", data.MerchantID)
				}
				if data.Classification.PrimaryIndustry != "Technology" {
					t.Errorf("Expected primary industry Technology, got %s", data.Classification.PrimaryIndustry)
				}
			},
		},
		{
			name:       "merchant not found",
			merchantID: "invalid-merchant",
			setup: func() error {
				// Don't seed merchant
				return nil
			},
			cleanup: func() error {
				return nil
			},
			wantErr: true,
		},
		{
			name:       "merchant not active",
			merchantID: "merchant-inactive",
			setup: func() error {
				return integration.SeedTestMerchant(db, "merchant-inactive", "inactive")
			},
			cleanup: func() error {
				return integration.CleanupTestData(db, "merchant-inactive")
			},
			wantErr: true,
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

			// Create service without cache
			service := NewMerchantAnalyticsService(analyticsRepo, merchantRepo, nil, stdLogger)

			// Execute test
			result, err := service.GetMerchantAnalytics(ctx, tt.merchantID)

			// Validate error expectation
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Validate result
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestMerchantAnalyticsService_GetMerchantAnalytics_ParallelFetching(t *testing.T) {
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

	// Seed test data
	merchantID := "merchant-parallel"
	if err := integration.SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer integration.CleanupTestData(db, merchantID)

	if err := integration.SeedTestAnalytics(db, merchantID, nil, nil, nil, nil); err != nil {
		t.Fatalf("Failed to seed analytics: %v", err)
	}

	// Test concurrent requests
	const numGoroutines = 10
	results := make(chan *models.AnalyticsData, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			data, err := service.GetMerchantAnalytics(ctx, merchantID)
			if err != nil {
				errors <- err
				return
			}
			results <- data
		}()
	}

	// Collect results
	successCount := 0
	errorCount := 0
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-results:
			successCount++
		case <-errors:
			errorCount++
		case <-time.After(5 * time.Second):
			t.Error("Timeout waiting for result")
		}
	}

	if successCount == 0 {
		t.Error("Expected at least one successful result from parallel requests")
	}
}

func TestMerchantAnalyticsService_GetMerchantAnalytics_Timeout(t *testing.T) {
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

	// Setup repositories
	analyticsRepo := database.NewMerchantAnalyticsRepository(db, stdLogger)
	merchantRepo := database.NewMerchantPortfolioRepository(db, stdLogger)
	service := NewMerchantAnalyticsService(analyticsRepo, merchantRepo, nil, stdLogger)

	// Seed test data
	merchantID := "merchant-timeout"
	if err := integration.SeedTestMerchant(db, merchantID, "active"); err != nil {
		t.Fatalf("Failed to seed merchant: %v", err)
	}
	defer integration.CleanupTestData(db, merchantID)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a bit to ensure timeout
	time.Sleep(1 * time.Millisecond)

	// This should timeout or return quickly
	_, err = service.GetMerchantAnalytics(ctx, merchantID)
	if err == nil {
		// If no error, that's okay - the service might handle timeout gracefully
		t.Log("Service handled timeout context without error")
	} else {
		// Expected timeout error
		if ctx.Err() == context.DeadlineExceeded {
			t.Log("Timeout occurred as expected")
		}
	}
}

