//go:build integration

package services

import (
	"context"
	"log"
	"os"
	"testing"

	integration "kyb-platform/test/integration"
)

func TestDataEnrichmentService_GetEnrichmentSources(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	service := NewDataEnrichmentService(stdLogger)
	ctx := context.Background()

	sources, err := service.GetEnrichmentSources(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(sources) == 0 {
		t.Error("Expected at least one enrichment source")
	}

	// Verify expected sources
	expectedSources := []string{"thomson-reuters", "dun-bradstreet", "government-registry"}
	foundSources := make(map[string]bool)
	for _, source := range sources {
		foundSources[source.ID] = true
		if !source.Enabled {
			t.Errorf("Expected source %s to be enabled", source.ID)
		}
	}

	for _, expected := range expectedSources {
		if !foundSources[expected] {
			t.Errorf("Expected source %s not found", expected)
		}
	}
}

func TestDataEnrichmentService_TriggerEnrichment(t *testing.T) {
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
	service := NewDataEnrichmentService(stdLogger)
	ctx := context.Background()

	tests := []struct {
		name       string
		merchantID string
		source     string
		wantErr    bool
		validate   func(*testing.T, *EnrichmentJob)
	}{
		{
			name:       "successful trigger with valid source",
			merchantID: "merchant-enrich-1",
			source:     "thomson-reuters",
			wantErr:    false,
			validate: func(t *testing.T, job *EnrichmentJob) {
				if job == nil {
					t.Fatal("Expected enrichment job but got nil")
				}
				if job.JobID == "" {
					t.Error("Expected job ID to be set")
				}
				if job.MerchantID != "merchant-enrich-1" {
					t.Errorf("Expected merchant ID merchant-enrich-1, got %s", job.MerchantID)
				}
				if job.Source != "thomson-reuters" {
					t.Errorf("Expected source thomson-reuters, got %s", job.Source)
				}
				if job.Status != "pending" {
					t.Errorf("Expected status pending, got %s", job.Status)
				}
			},
		},
		{
			name:       "invalid source",
			merchantID: "merchant-enrich-2",
			source:     "invalid-source",
			wantErr:    true,
		},
		{
			name:       "valid source dun-bradstreet",
			merchantID: "merchant-enrich-3",
			source:     "dun-bradstreet",
			wantErr:    false,
			validate: func(t *testing.T, job *EnrichmentJob) {
				if job.Source != "dun-bradstreet" {
					t.Errorf("Expected source dun-bradstreet, got %s", job.Source)
				}
			},
		},
		{
			name:       "valid source government-registry",
			merchantID: "merchant-enrich-4",
			source:     "government-registry",
			wantErr:    false,
			validate: func(t *testing.T, job *EnrichmentJob) {
				if job.Source != "government-registry" {
					t.Errorf("Expected source government-registry, got %s", job.Source)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Seed merchant if needed
			if !tt.wantErr {
				if err := integration.SeedTestMerchant(db, tt.merchantID, "active"); err != nil {
					t.Fatalf("Failed to seed merchant: %v", err)
				}
				defer integration.CleanupTestData(db, tt.merchantID)
			}

			job, err := service.TriggerEnrichment(ctx, tt.merchantID, tt.source)

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
				tt.validate(t, job)
			}
		})
	}
}

