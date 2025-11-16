//go:build integration

package integration

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/database"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/services"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestWeeks24Integration tests the new features added in Weeks 2-4
func TestWeeks24Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database (if available)
	db, err := setupTestDB(t)
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer cleanupTestDB(t, db)

	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
	
	// Create observability logger for routes
	zapLogger, _ := zap.NewDevelopment()
	obsLogger := observability.NewLogger(zapLogger)

	// Setup repositories
	merchantRepo := database.NewMerchantPortfolioRepository(db, stdLogger)
	analyticsRepo := database.NewMerchantAnalyticsRepository(db, stdLogger)
	riskAssessmentRepo := database.NewRiskAssessmentRepository(db, stdLogger)
	riskIndicatorsRepo := database.NewRiskIndicatorsRepository(db, stdLogger)

	// Setup services
	analyticsService := services.NewMerchantAnalyticsService(analyticsRepo, merchantRepo, nil, stdLogger)
	riskAssessmentService := services.NewRiskAssessmentService(riskAssessmentRepo, nil, stdLogger)
	riskIndicatorsService := services.NewRiskIndicatorsService(riskIndicatorsRepo, stdLogger)
	dataEnrichmentService := services.NewDataEnrichmentService(stdLogger)

	// Setup handlers
	analyticsHandler := handlers.NewMerchantAnalyticsHandler(analyticsService, stdLogger)
	asyncRiskHandler := handlers.NewAsyncRiskAssessmentHandler(riskAssessmentService, stdLogger)
	riskIndicatorsHandler := handlers.NewRiskIndicatorsHandler(riskIndicatorsService, stdLogger)
	dataEnrichmentHandler := handlers.NewDataEnrichmentHandler(dataEnrichmentService, stdLogger)

	// Setup routes
	mux := http.NewServeMux()
	merchantConfig := &routes.MerchantRouteConfig{
		MerchantPortfolioHandler: nil,
		MerchantAnalyticsHandler: analyticsHandler,
		AsyncRiskHandler:         asyncRiskHandler,
		DataEnrichmentHandler:     dataEnrichmentHandler,
		AuthMiddleware:            nil,
		RateLimiter:               nil,
		Logger:                    obsLogger,
	}
	routes.RegisterMerchantRoutes(mux, merchantConfig)

	asyncRiskConfig := &routes.AsyncRiskAssessmentRouteConfig{
		AsyncRiskHandler:     asyncRiskHandler,
		RiskIndicatorsHandler: riskIndicatorsHandler,
		AuthMiddleware:       nil,
		RateLimiter:          nil,
	}
	routes.RegisterRiskRoutesWithConfig(mux, nil, asyncRiskConfig)

	// Create test server
	server := httptest.NewServer(mux)
	defer server.Close()

	t.Run("GetMerchantAnalytics", func(t *testing.T) {
		// This would require a test merchant in the database
		// For now, we test that the endpoint exists and returns proper error for missing merchant
		req := httptest.NewRequest("GET", "/api/v1/merchants/test-merchant-123/analytics", nil)
		w := httptest.NewRecorder()

		analyticsHandler.GetMerchantAnalytics(w, req)

		// Should return error for non-existent merchant
		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
	})

	t.Run("GetWebsiteAnalysis", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/merchants/test-merchant-123/website-analysis", nil)
		w := httptest.NewRecorder()

		analyticsHandler.GetWebsiteAnalysis(w, req)

		// Should return error for non-existent merchant
		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
	})

	t.Run("GetRiskHistory", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/risk/history/test-merchant-123?limit=10&offset=0", nil)
		w := httptest.NewRecorder()

		asyncRiskHandler.GetRiskHistory(w, req)

		// Should handle request (may return empty array or error)
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})

	t.Run("GetRiskPredictions", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/risk/predictions/test-merchant-123?horizons=3,6,12", nil)
		w := httptest.NewRecorder()

		asyncRiskHandler.GetRiskPredictions(w, req)

		// Should handle request
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})

	t.Run("GetRiskIndicators", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/risk/indicators/test-merchant-123", nil)
		w := httptest.NewRecorder()

		riskIndicatorsHandler.GetRiskIndicators(w, req)

		// Should handle request
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})

	t.Run("GetEnrichmentSources", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/merchants/test-merchant-123/enrichment/sources", nil)
		w := httptest.NewRecorder()

		dataEnrichmentHandler.GetEnrichmentSources(w, req)

		// Should return sources (even if empty)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Enhanced tests with seeded data
	t.Run("GetMerchantAnalytics_WithSeededData", func(t *testing.T) {
		merchantID := "merchant-enhanced-1"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		if err := SeedTestAnalytics(db, merchantID,
			map[string]interface{}{
				"primaryIndustry": "Technology",
				"confidenceScore": 0.95,
			},
			map[string]interface{}{
				"trustScore": 0.8,
				"sslValid":   true,
			},
			map[string]interface{}{
				"completenessScore": 0.9,
			},
			nil,
		); err != nil {
			t.Fatalf("Failed to seed analytics: %v", err)
		}

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/merchants/%s/analytics", merchantID), nil)
		w := httptest.NewRecorder()

		analyticsHandler.GetMerchantAnalytics(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetRiskHistory_WithSeededData", func(t *testing.T) {
		merchantID := "merchant-enhanced-2"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		// Seed multiple assessments
		for i := 1; i <= 5; i++ {
			if err := SeedTestRiskAssessment(db, merchantID, fmt.Sprintf("assessment-enhanced-%d", i), "completed", nil); err != nil {
				t.Fatalf("Failed to seed assessment %d: %v", i, err)
			}
		}

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/history/%s?limit=10&offset=0", merchantID), nil)
		w := httptest.NewRecorder()

		asyncRiskHandler.GetRiskHistory(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetRiskIndicators_WithSeededData", func(t *testing.T) {
		merchantID := "merchant-enhanced-3"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		if err := SeedTestRiskIndicators(db, merchantID, 5, ""); err != nil {
			t.Fatalf("Failed to seed indicators: %v", err)
		}

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/indicators/%s", merchantID), nil)
		w := httptest.NewRecorder()

		riskIndicatorsHandler.GetRiskIndicators(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Error scenarios
	t.Run("GetMerchantAnalytics_InvalidMerchantID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/merchants/invalid-id-format/analytics", nil)
		w := httptest.NewRecorder()

		analyticsHandler.GetMerchantAnalytics(w, req)

		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusBadRequest)
	})

	t.Run("GetRiskHistory_EmptyResult", func(t *testing.T) {
		merchantID := "merchant-empty"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/history/%s?limit=10&offset=0", merchantID), nil)
		w := httptest.NewRecorder()

		asyncRiskHandler.GetRiskHistory(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Edge cases
	t.Run("GetRiskHistory_PaginationBoundaries", func(t *testing.T) {
		merchantID := "merchant-pagination"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		// Seed 15 assessments
		for i := 1; i <= 15; i++ {
			if err := SeedTestRiskAssessment(db, merchantID, fmt.Sprintf("assessment-pag-%d", i), "completed", nil); err != nil {
				t.Fatalf("Failed to seed assessment %d: %v", i, err)
			}
		}

		// Test first page
		req1 := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/history/%s?limit=10&offset=0", merchantID), nil)
		w1 := httptest.NewRecorder()
		asyncRiskHandler.GetRiskHistory(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Test last page
		req2 := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/history/%s?limit=10&offset=10", merchantID), nil)
		w2 := httptest.NewRecorder()
		asyncRiskHandler.GetRiskHistory(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		// Test beyond boundaries
		req3 := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/history/%s?limit=10&offset=100", merchantID), nil)
		w3 := httptest.NewRecorder()
		asyncRiskHandler.GetRiskHistory(w3, req3)
		assert.Equal(t, http.StatusOK, w3.Code)
	})

	// Concurrent operations
	t.Run("ConcurrentGetMerchantAnalytics", func(t *testing.T) {
		merchantID := "merchant-concurrent"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		if err := SeedTestAnalytics(db, merchantID, nil, nil, nil, nil); err != nil {
			t.Fatalf("Failed to seed analytics: %v", err)
		}

		const numRequests = 10
		results := make(chan int, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/merchants/%s/analytics", merchantID), nil)
				w := httptest.NewRecorder()
				analyticsHandler.GetMerchantAnalytics(w, req)
				results <- w.Code
			}()
		}

		successCount := 0
		for i := 0; i < numRequests; i++ {
			code := <-results
			if code == http.StatusOK {
				successCount++
			}
		}

		if successCount == 0 {
			t.Error("Expected at least one successful concurrent request")
		}
	})
}

// setupTestDB sets up a test database connection using the database_setup helper
func setupTestDB(t *testing.T) (*sql.DB, error) {
	testDB, err := SetupTestDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to setup test database: %w", err)
	}

	// Store the TestDatabase instance for cleanup
	t.Cleanup(func() {
		if err := testDB.CleanupTestDatabase(); err != nil {
			t.Logf("Error cleaning up test database: %v", err)
		}
	})

	return testDB.GetDB(), nil
}

// cleanupTestDB cleans up test database
func cleanupTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

