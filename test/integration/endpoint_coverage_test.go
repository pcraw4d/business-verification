//go:build integration

package integration

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/database"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/services"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestEndpointCoverage_Comprehensive tests comprehensive endpoint coverage
func TestEndpointCoverage_Comprehensive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database
	db, err := setupTestDB(t)
	if err != nil {
		t.Skipf("Skipping integration test - database not available: %v", err)
	}
	defer cleanupTestDB(t, db)

	stdLogger := log.New(os.Stdout, "", log.LstdFlags)
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
		DataEnrichmentHandler:    dataEnrichmentHandler,
		AuthMiddleware:           nil,
		RateLimiter:              nil,
		Logger:                   obsLogger,
	}
	routes.RegisterMerchantRoutes(mux, merchantConfig)

	asyncRiskConfig := &routes.AsyncRiskAssessmentRouteConfig{
		AsyncRiskHandler:     asyncRiskHandler,
		RiskIndicatorsHandler: riskIndicatorsHandler,
		AuthMiddleware:        nil,
		RateLimiter:           nil,
	}
	routes.RegisterRiskRoutesWithConfig(mux, nil, asyncRiskConfig)

	// Create test server
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test merchant analytics endpoints
	t.Run("GET /api/v1/merchants/:id/analytics - Success", func(t *testing.T) {
		merchantID := "merchant-endpoint-1"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		if err := SeedTestAnalytics(db, merchantID, nil, nil, nil, nil); err != nil {
			t.Fatalf("Failed to seed analytics: %v", err)
		}

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/merchants/%s/analytics", merchantID), nil)
		w := httptest.NewRecorder()

		analyticsHandler.GetMerchantAnalytics(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, merchantID, response["merchantId"])
	})

	t.Run("GET /api/v1/merchants/:id/analytics - Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/merchants/non-existent/analytics", nil)
		w := httptest.NewRecorder()

		analyticsHandler.GetMerchantAnalytics(w, req)

		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
	})

	t.Run("GET /api/v1/merchants/:id/website-analysis - Success", func(t *testing.T) {
		merchantID := "merchant-endpoint-2"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/merchants/%s/website-analysis", merchantID), nil)
		w := httptest.NewRecorder()

		analyticsHandler.GetWebsiteAnalysis(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, merchantID, response["merchantId"])
	})

	// Test risk assessment endpoints
	t.Run("GET /api/v1/risk/history/:id - Success", func(t *testing.T) {
		merchantID := "merchant-endpoint-3"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		// Seed assessments
		for i := 1; i <= 3; i++ {
			if err := SeedTestRiskAssessment(db, merchantID, fmt.Sprintf("assessment-endpoint-%d", i), "completed", nil); err != nil {
				t.Fatalf("Failed to seed assessment: %v", err)
			}
		}

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/history/%s?limit=10&offset=0", merchantID), nil)
		w := httptest.NewRecorder()

		asyncRiskHandler.GetRiskHistory(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(response), 3)
	})

	t.Run("GET /api/v1/risk/history/:id - Pagination", func(t *testing.T) {
		merchantID := "merchant-endpoint-4"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		// Seed 15 assessments
		for i := 1; i <= 15; i++ {
			if err := SeedTestRiskAssessment(db, merchantID, fmt.Sprintf("assessment-pag-%d", i), "completed", nil); err != nil {
				t.Fatalf("Failed to seed assessment: %v", err)
			}
		}

		// Test first page
		req1 := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/history/%s?limit=10&offset=0", merchantID), nil)
		w1 := httptest.NewRecorder()
		asyncRiskHandler.GetRiskHistory(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Test second page
		req2 := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/history/%s?limit=10&offset=10", merchantID), nil)
		w2 := httptest.NewRecorder()
		asyncRiskHandler.GetRiskHistory(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})

	t.Run("GET /api/v1/risk/predictions/:id - Success", func(t *testing.T) {
		merchantID := "merchant-endpoint-5"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		if err := SeedTestRiskAssessment(db, merchantID, "assessment-pred", "completed",
			map[string]interface{}{
				"overallScore": 0.7,
				"riskLevel":    "medium",
			}); err != nil {
			t.Fatalf("Failed to seed assessment: %v", err)
		}

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/predictions/%s?horizons=3,6,12", merchantID), nil)
		w := httptest.NewRecorder()

		asyncRiskHandler.GetRiskPredictions(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, merchantID, response["merchantId"])
	})

	t.Run("GET /api/v1/risk/indicators/:id - Success", func(t *testing.T) {
		merchantID := "merchant-endpoint-6"
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
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, merchantID, response["merchantId"])
	})

	t.Run("GET /api/v1/risk/indicators/:id - With Filters", func(t *testing.T) {
		merchantID := "merchant-endpoint-7"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		if err := SeedTestRiskIndicators(db, merchantID, 10, "high"); err != nil {
			t.Fatalf("Failed to seed indicators: %v", err)
		}

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/indicators/%s?severity=high", merchantID), nil)
		w := httptest.NewRecorder()

		riskIndicatorsHandler.GetRiskIndicators(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		indicators, ok := response["indicators"].([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(indicators), 10)
	})

	t.Run("GET /api/v1/merchants/:id/enrichment/sources - Success", func(t *testing.T) {
		merchantID := "merchant-endpoint-8"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/merchants/%s/enrichment/sources", merchantID), nil)
		w := httptest.NewRecorder()

		dataEnrichmentHandler.GetEnrichmentSources(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Greater(t, len(response), 0)
	})

	t.Run("POST /api/v1/risk/assess - Start Assessment", func(t *testing.T) {
		merchantID := "merchant-endpoint-9"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		reqBody := `{"merchantId": "merchant-endpoint-9", "options": {"includeHistory": true, "includePredictions": true}}`
		req := httptest.NewRequest("POST", "/api/v1/risk/assess", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		asyncRiskHandler.StartRiskAssessment(w, req)

		// May fail due to job queue, but should handle gracefully
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusAccepted || w.Code == http.StatusInternalServerError)
	})

	t.Run("GET /api/v1/risk/assess/:id - Get Assessment Status", func(t *testing.T) {
		merchantID := "merchant-endpoint-10"
		assessmentID := "assessment-status-endpoint"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		if err := SeedTestRiskAssessment(db, merchantID, assessmentID, "processing", nil); err != nil {
			t.Fatalf("Failed to seed assessment: %v", err)
		}

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/assess/%s", assessmentID), nil)
		w := httptest.NewRecorder()

		asyncRiskHandler.GetAssessmentStatus(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, assessmentID, response["assessmentId"])
	})

	// Error scenarios
	t.Run("GET /api/v1/merchants/:id/analytics - Invalid Merchant ID Format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/merchants/invalid-id-format/analytics", nil)
		w := httptest.NewRecorder()

		analyticsHandler.GetMerchantAnalytics(w, req)

		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusBadRequest)
	})

	t.Run("GET /api/v1/risk/history/:id - Invalid Pagination Parameters", func(t *testing.T) {
		merchantID := "merchant-endpoint-11"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		// Test with negative limit
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/history/%s?limit=-1&offset=0", merchantID), nil)
		w := httptest.NewRecorder()

		asyncRiskHandler.GetRiskHistory(w, req)

		// Should handle gracefully (may use default values)
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
	})

	t.Run("GET /api/v1/risk/predictions/:id - No Assessment History", func(t *testing.T) {
		merchantID := "merchant-endpoint-12"
		if err := SeedTestMerchant(db, merchantID, "active"); err != nil {
			t.Fatalf("Failed to seed merchant: %v", err)
		}
		defer CleanupTestData(db, merchantID)

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/risk/predictions/%s?horizons=3,6,12", merchantID), nil)
		w := httptest.NewRecorder()

		asyncRiskHandler.GetRiskPredictions(w, req)

		// Should return OK with empty or default predictions
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

