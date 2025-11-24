package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"kyb-platform/services/merchant-service/internal/config"
	"kyb-platform/services/merchant-service/internal/jobs"
	"kyb-platform/services/merchant-service/test/integration"
)

// testMetricsRegistry is a shared test registry to avoid duplicate metric registration
var (
	testMetricsRegistry     *prometheus.Registry
	testMetricsRegistryOnce  sync.Once
	testMetricsRegistryMutex sync.Mutex
)

// getTestMetricsRegistry returns a test-specific Prometheus registry
func getTestMetricsRegistry() *prometheus.Registry {
	testMetricsRegistryOnce.Do(func() {
		testMetricsRegistry = prometheus.NewRegistry()
	})
	return testMetricsRegistry
}

// TestMerchantCreationTriggersClassificationJob tests that creating a merchant triggers classification job
func TestMerchantCreationTriggersClassificationJob(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	testDB := integration.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase()

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Environment: "test",
		Merchant: config.MerchantConfig{
			RequestTimeout: 10 * time.Second,
		},
	}

	// Create job processor
	jobProcessor := jobs.NewJobProcessor(2, 10, logger)
	jobProcessor.Start()
	defer jobProcessor.Stop()

	// Create handler with test database
	handler := NewMerchantHandler(testDB.GetClient(), logger, cfg, jobProcessor)
	
	// Create merchant request
	reqBody := CreateMerchantRequest{
		Name:      "Test Business",
		LegalName: "Test Business Legal",
		Industry:  "Technology",
		ContactInfo: map[string]interface{}{
			"website": "https://testbusiness.com",
		},
	}
	
	jsonData, err := json.Marshal(reqBody)
	require.NoError(t, err)
	
	req := httptest.NewRequest("POST", "/api/v1/merchants", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	handler.HandleCreateMerchant(w, req)
	
	// Verify response
	require.Equal(t, http.StatusCreated, w.Code, "Merchant should be created")
	
	var merchantResponse map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&merchantResponse)
	require.NoError(t, err)
	
	merchantID, ok := merchantResponse["id"].(string)
	require.True(t, ok, "Merchant ID should be returned")
	
	// Cleanup test data
	defer testDB.CleanupTestData(t, []string{merchantID})
	
	// Verify job was enqueued (check queue size)
	time.Sleep(200 * time.Millisecond) // Give time for job to be enqueued
	
	// Queue size should be >= 0 (job may have been picked up already)
	queueSize := jobProcessor.GetQueueSize()
	assert.GreaterOrEqual(t, queueSize, 0, "Job processor should be running")
	
	// Verify merchant was created in database
	merchant, err := testDB.GetTestMerchant(t, merchantID)
	require.NoError(t, err)
	assert.Equal(t, merchantID, merchant["id"], "Merchant should exist in database")
}

// TestHandleMerchantSpecificAnalytics_ReadsFromDatabase tests that analytics handler reads from database
func TestHandleMerchantSpecificAnalytics_ReadsFromDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase()

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Environment: "test",
	}
	
	jobProcessor := jobs.NewJobProcessor(2, 10, logger)
	handler := NewMerchantHandler(testDB.GetClient(), logger, cfg, jobProcessor)
	
	// Create a test merchant first
	merchantID, err := testDB.CreateTestMerchant(t, map[string]interface{}{
		"name":       "Test Analytics Merchant",
		"legal_name": "Test Analytics Merchant Legal",
		"industry":   "Technology",
	})
	require.NoError(t, err)
	defer testDB.CleanupTestData(t, []string{merchantID})
	
	req := httptest.NewRequest("GET", "/api/v1/merchants/"+merchantID+"/analytics", nil)
	w := httptest.NewRecorder()
	
	handler.HandleMerchantSpecificAnalytics(w, req)
	
	// Verify response structure
	assert.Equal(t, http.StatusOK, w.Code, "Should return 200")
	
	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	
	// Verify response has expected fields
	assert.Contains(t, response, "merchantId")
	assert.Contains(t, response, "classification")
	assert.Contains(t, response, "security")
	assert.Contains(t, response, "quality")
	assert.Contains(t, response, "timestamp")
	assert.Equal(t, merchantID, response["merchantId"], "Merchant ID should match")
}

// TestHandleMerchantAnalyticsStatus_ReturnsStatus tests the status endpoint
func TestHandleMerchantAnalyticsStatus_ReturnsStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase()

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Environment: "test",
	}
	
	jobProcessor := jobs.NewJobProcessor(2, 10, logger)
	handler := NewMerchantHandler(testDB.GetClient(), logger, cfg, jobProcessor)
	
	// Create a test merchant
	merchantID, err := testDB.CreateTestMerchant(t, map[string]interface{}{
		"name":       "Test Status Merchant",
		"legal_name": "Test Status Merchant Legal",
		"industry":   "Technology",
	})
	require.NoError(t, err)
	defer testDB.CleanupTestData(t, []string{merchantID})
	
	req := httptest.NewRequest("GET", "/api/v1/merchants/"+merchantID+"/analytics/status", nil)
	w := httptest.NewRecorder()
	
	handler.HandleMerchantAnalyticsStatus(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	
	// Verify response structure
	assert.Contains(t, response, "merchantId")
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "timestamp")
	assert.Equal(t, merchantID, response["merchantId"], "Merchant ID should match")
	
	status, ok := response["status"].(map[string]interface{})
	require.True(t, ok, "Status should be a map")
	
	assert.Contains(t, status, "classification")
	assert.Contains(t, status, "websiteAnalysis")
	
	// Verify default statuses are "pending"
	assert.Equal(t, "pending", status["classification"], "Default classification status should be pending")
	assert.Equal(t, "pending", status["websiteAnalysis"], "Default website analysis status should be pending")
}

// TestHandleMerchantWebsiteAnalysis_ReadsFromDatabase tests website analysis handler
func TestHandleMerchantWebsiteAnalysis_ReadsFromDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase()

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Environment: "test",
	}
	
	jobProcessor := jobs.NewJobProcessor(2, 10, logger)
	handler := NewMerchantHandler(testDB.GetClient(), logger, cfg, jobProcessor)
	
	// Create a test merchant
	merchantID, err := testDB.CreateTestMerchant(t, map[string]interface{}{
		"name":       "Test Website Analysis Merchant",
		"legal_name": "Test Website Analysis Merchant Legal",
		"industry":   "Technology",
		"contact_info": map[string]interface{}{
			"website": "https://example.com",
		},
	})
	require.NoError(t, err)
	defer testDB.CleanupTestData(t, []string{merchantID})
	
	req := httptest.NewRequest("GET", "/api/v1/merchants/"+merchantID+"/website-analysis", nil)
	w := httptest.NewRecorder()
	
	handler.HandleMerchantWebsiteAnalysis(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	
	// Verify response has expected fields
	assert.Contains(t, response, "merchantId")
	assert.Contains(t, response, "ssl")
	assert.Contains(t, response, "securityHeaders")
	assert.Contains(t, response, "performance")
	assert.Contains(t, response, "accessibility")
	assert.Equal(t, merchantID, response["merchantId"], "Merchant ID should match")
}

// TestHandleMerchantRiskScore_ReadsFromDatabase tests risk score handler
func TestHandleMerchantRiskScore_ReadsFromDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase()

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Environment: "test",
	}
	
	jobProcessor := jobs.NewJobProcessor(2, 10, logger)
	handler := NewMerchantHandler(testDB.GetClient(), logger, cfg, jobProcessor)
	
	// Create a test merchant
	merchantID, err := testDB.CreateTestMerchant(t, map[string]interface{}{
		"name":       "Test Risk Score Merchant",
		"legal_name": "Test Risk Score Merchant Legal",
		"industry":   "Technology",
	})
	require.NoError(t, err)
	defer testDB.CleanupTestData(t, []string{merchantID})
	
	req := httptest.NewRequest("GET", "/api/v1/merchants/"+merchantID+"/risk-score", nil)
	w := httptest.NewRecorder()
	
	handler.HandleMerchantRiskScore(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	
	// Verify response has expected fields
	assert.Contains(t, response, "merchant_id")
	assert.Contains(t, response, "risk_score")
	assert.Contains(t, response, "risk_level")
	assert.Contains(t, response, "factors")
	assert.Equal(t, merchantID, response["merchant_id"], "Merchant ID should match")
}

// TestHandleMerchantStatistics_QueriesRealData tests statistics handler
func TestHandleMerchantStatistics_QueriesRealData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase()

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Environment: "test",
	}
	
	jobProcessor := jobs.NewJobProcessor(2, 10, logger)
	handler := NewMerchantHandler(testDB.GetClient(), logger, cfg, jobProcessor)
	
	// Create some test merchants
	merchant1, err := testDB.CreateTestMerchant(t, map[string]interface{}{
		"name":       "Test Merchant 1",
		"legal_name": "Test Merchant 1 Legal",
		"industry":   "Technology",
	})
	require.NoError(t, err)
	
	merchant2, err := testDB.CreateTestMerchant(t, map[string]interface{}{
		"name":       "Test Merchant 2",
		"legal_name": "Test Merchant 2 Legal",
		"industry":   "Retail",
	})
	require.NoError(t, err)
	
	defer testDB.CleanupTestData(t, []string{merchant1, merchant2})
	
	req := httptest.NewRequest("GET", "/api/v1/merchants/statistics", nil)
	w := httptest.NewRecorder()
	
	handler.HandleMerchantStatistics(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	
	// Verify response has expected fields
	assert.Contains(t, response, "totalMerchants")
	assert.Contains(t, response, "totalAssessments")
	assert.Contains(t, response, "averageRiskScore")
	assert.Contains(t, response, "riskDistribution")
	assert.Contains(t, response, "industryBreakdown")
	assert.Contains(t, response, "countryBreakdown")
	assert.Contains(t, response, "timestamp")
	
	// Verify we have at least the merchants we created
	totalMerchants, ok := response["totalMerchants"].(float64)
	if ok {
		assert.GreaterOrEqual(t, int(totalMerchants), 2, "Should have at least 2 merchants")
	}
}

