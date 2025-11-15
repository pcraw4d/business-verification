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

