package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/database"
)

// TestMerchantPortfolioHandlerWithRepository tests the handler with real repository
func TestMerchantPortfolioHandlerWithRepository(t *testing.T) {
	dbURL := getTestDatabaseURL(t)
	
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	
	// Setup test data
	setupTestDataForHandler(t, db, ctx)
	defer cleanupTestDataForHandler(t, db, ctx)
	
	// Create repository
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	repo := database.NewMerchantPortfolioRepository(db, logger)
	require.NotNil(t, repo, "Should be able to create repository")
	
	// Create handler with repository
	handler := handlers.NewMerchantPortfolioHandlerWithRepository(nil, repo, nil)
	require.NotNil(t, handler, "Handler should be created")
	
	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/merchants/analytics", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	
	// Call handler
	handler.GetMerchantAnalytics(w, req)
	
	// Verify response
	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")
	
	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err, "Should be able to decode response")
	
	// Verify response structure
	assert.Contains(t, response, "total_merchants", "Response should contain total_merchants")
	assert.Contains(t, response, "portfolio_distribution", "Response should contain portfolio_distribution")
	assert.Contains(t, response, "risk_distribution", "Response should contain risk_distribution")
	assert.Contains(t, response, "industry_distribution", "Response should contain industry_distribution")
	assert.Contains(t, response, "compliance_status", "Response should contain compliance_status")
	assert.Contains(t, response, "created_at", "Response should contain created_at")
	
	// Verify data matches what we inserted
	totalMerchants, ok := response["total_merchants"].(float64)
	require.True(t, ok, "total_merchants should be a number")
	assert.Greater(t, int(totalMerchants), 0, "Should have merchants")
	
	portfolioDist, ok := response["portfolio_distribution"].(map[string]interface{})
	require.True(t, ok, "portfolio_distribution should be a map")
	assert.Greater(t, len(portfolioDist), 0, "Should have portfolio distribution data")
	
	riskDist, ok := response["risk_distribution"].(map[string]interface{})
	require.True(t, ok, "risk_distribution should be a map")
	assert.Greater(t, len(riskDist), 0, "Should have risk distribution data")
	
	complianceDist, ok := response["compliance_status"].(map[string]interface{})
	require.True(t, ok, "compliance_status should be a map")
	assert.Greater(t, len(complianceDist), 0, "Should have compliance distribution data")
	
	t.Logf("✓ Handler returned analytics with %d total merchants", int(totalMerchants))
	t.Logf("✓ Portfolio distribution: %v", portfolioDist)
	t.Logf("✓ Risk distribution: %v", riskDist)
	t.Logf("✓ Compliance distribution: %v", complianceDist)
}

// setupTestDataForHandler creates test data for handler testing (reuses setup from repository_methods_test.go)
func setupTestDataForHandler(t *testing.T, db *sql.DB, ctx context.Context) {
	// Reuse the same setup function - we'll call it from repository_methods_test.go
	// For now, we'll create a simplified version
	setupTestData(t, db, ctx)
}

// cleanupTestDataForHandler cleans up test data
func cleanupTestDataForHandler(t *testing.T, db *sql.DB, ctx context.Context) {
	// Reuse cleanup
	cleanupTestData(t, db, ctx)
}

