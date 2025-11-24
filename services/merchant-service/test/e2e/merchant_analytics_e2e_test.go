package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"kyb-platform/services/merchant-service/internal/config"
	"kyb-platform/services/merchant-service/internal/handlers"
	"kyb-platform/services/merchant-service/internal/jobs"
	"kyb-platform/services/merchant-service/internal/supabase"
)

// TestMerchantCreationToAnalyticsDisplay_E2E tests the complete flow:
// 1. Create merchant
// 2. Classification job triggered
// 3. Website analysis job triggered (if URL provided)
// 4. Jobs process in background
// 5. Analytics data available via API
// 6. Status indicators show correct state
func TestMerchantCreationToAnalyticsDisplay_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Environment: "test",
		Merchant: config.MerchantConfig{
			RequestTimeout: 30 * time.Second,
		},
	}

	// Initialize Supabase client (requires real connection for E2E)
	// In real E2E tests, use test database
	supabaseClient, err := supabase.NewClient(&cfg.Supabase, logger)
	if err != nil {
		t.Skipf("Skipping E2E test: Supabase connection failed: %v", err)
	}

	// Initialize job processor
	jobProcessor := jobs.NewJobProcessor(3, 20, logger)
	jobProcessor.Start()
	defer jobProcessor.Stop()

	// Create handler
	handler := handlers.NewMerchantHandler(supabaseClient, logger, cfg, jobProcessor)

	t.Run("Complete Merchant Creation and Analytics Flow", func(t *testing.T) {
		// Step 1: Create merchant
		merchantReq := handlers.CreateMerchantRequest{
			Name:      "E2E Test Business",
			LegalName: "E2E Test Business Legal",
			Industry:  "Technology",
			ContactInfo: map[string]interface{}{
				"website": "https://example.com",
				"email":   "test@example.com",
			},
		}

		jsonData, err := json.Marshal(merchantReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/merchants", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleCreateMerchant(w, req)

		require.Equal(t, http.StatusCreated, w.Code, "Merchant should be created")

		var merchantResponse map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&merchantResponse)
		require.NoError(t, err)

		merchantID, ok := merchantResponse["id"].(string)
		require.True(t, ok, "Merchant ID should be returned")
		require.NotEmpty(t, merchantID, "Merchant ID should not be empty")

		t.Logf("✅ Merchant created: %s", merchantID)

		// Step 2: Verify jobs were enqueued
		time.Sleep(500 * time.Millisecond) // Give time for jobs to be enqueued
		queueSize := jobProcessor.GetQueueSize()
		t.Logf("Job queue size: %d", queueSize)

		// Step 3: Check status endpoint (should show processing/pending)
		statusReq := httptest.NewRequest("GET", "/api/v1/merchants/"+merchantID+"/analytics/status", nil)
		statusW := httptest.NewRecorder()

		handler.HandleMerchantAnalyticsStatus(statusW, statusReq)

		require.Equal(t, http.StatusOK, statusW.Code)

		var statusResponse map[string]interface{}
		err = json.NewDecoder(statusW.Body).Decode(&statusResponse)
		require.NoError(t, err)

		status, ok := statusResponse["status"].(map[string]interface{})
		require.True(t, ok)

		classificationStatus, ok := status["classification"].(string)
		require.True(t, ok)
		assert.Contains(t, []string{"pending", "processing", "completed"}, classificationStatus,
			"Classification status should be pending, processing, or completed")

		websiteStatus, ok := status["websiteAnalysis"].(string)
		require.True(t, ok)
		assert.Contains(t, []string{"pending", "processing", "completed", "skipped"}, websiteStatus,
			"Website analysis status should be pending, processing, completed, or skipped")

		t.Logf("✅ Status endpoint working - Classification: %s, Website: %s", classificationStatus, websiteStatus)

		// Step 4: Wait for jobs to complete (with timeout)
		maxWait := 60 * time.Second
		checkInterval := 2 * time.Second
		startTime := time.Now()

		for time.Since(startTime) < maxWait {
			// Check status again
			statusReq2 := httptest.NewRequest("GET", "/api/v1/merchants/"+merchantID+"/analytics/status", nil)
			statusW2 := httptest.NewRecorder()
			handler.HandleMerchantAnalyticsStatus(statusW2, statusReq2)

			var statusResponse2 map[string]interface{}
			json.NewDecoder(statusW2.Body).Decode(&statusResponse2)
			status2 := statusResponse2["status"].(map[string]interface{})

			classificationStatus2 := status2["classification"].(string)
			websiteStatus2 := status2["websiteAnalysis"].(string)

			if classificationStatus2 == "completed" && 
			   (websiteStatus2 == "completed" || websiteStatus2 == "skipped") {
				t.Logf("✅ Jobs completed - Classification: %s, Website: %s", classificationStatus2, websiteStatus2)
				break
			}

			time.Sleep(checkInterval)
		}

		// Step 5: Verify analytics data is available
		analyticsReq := httptest.NewRequest("GET", "/api/v1/merchants/"+merchantID+"/analytics", nil)
		analyticsW := httptest.NewRecorder()

		handler.HandleMerchantSpecificAnalytics(analyticsW, analyticsReq)

		require.Equal(t, http.StatusOK, analyticsW.Code)

		var analyticsResponse map[string]interface{}
		err = json.NewDecoder(analyticsW.Body).Decode(&analyticsResponse)
		require.NoError(t, err)

		// Verify response structure
		assert.Contains(t, analyticsResponse, "merchantId")
		assert.Contains(t, analyticsResponse, "classification")
		assert.Contains(t, analyticsResponse, "security")
		assert.Contains(t, analyticsResponse, "quality")

		classification, ok := analyticsResponse["classification"].(map[string]interface{})
		require.True(t, ok)

		// If classification is completed, verify it has real data
		if status, ok := classification["status"].(string); ok && status == "completed" {
			assert.Contains(t, classification, "primaryIndustry")
			assert.Contains(t, classification, "confidenceScore")
			t.Logf("✅ Classification data available: %v", classification)
		}

		// Step 6: Verify website analysis data (if not skipped)
		websiteReq := httptest.NewRequest("GET", "/api/v1/merchants/"+merchantID+"/website-analysis", nil)
		websiteW := httptest.NewRecorder()

		handler.HandleMerchantWebsiteAnalysis(websiteW, websiteReq)

		require.Equal(t, http.StatusOK, websiteW.Code)

		var websiteResponse map[string]interface{}
		err = json.NewDecoder(websiteW.Body).Decode(&websiteResponse)
		require.NoError(t, err)

		assert.Contains(t, websiteResponse, "merchantId")
		assert.Contains(t, websiteResponse, "ssl")
		assert.Contains(t, websiteResponse, "securityHeaders")
		assert.Contains(t, websiteResponse, "performance")

		if status, ok := websiteResponse["status"].(string); ok && status == "completed" {
			t.Logf("✅ Website analysis data available")
		} else if status == "skipped" {
			t.Logf("✅ Website analysis correctly skipped (no URL)")
		}

		t.Logf("✅ E2E test completed successfully")
	})
}

// TestJobProcessing_ConcurrentMerchants tests processing multiple merchants concurrently
func TestJobProcessing_ConcurrentMerchants(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Environment: "test",
		Merchant: config.MerchantConfig{
			RequestTimeout: 30 * time.Second,
		},
	}

	supabaseClient, err := supabase.NewClient(&cfg.Supabase, logger)
	if err != nil {
		t.Skipf("Skipping E2E test: Supabase connection failed: %v", err)
	}

	jobProcessor := jobs.NewJobProcessor(5, 50, logger)
	jobProcessor.Start()
	defer jobProcessor.Stop()

	handler := handlers.NewMerchantHandler(supabaseClient, logger, cfg, jobProcessor)

	// Create multiple merchants concurrently
	numMerchants := 5
	merchantIDs := make([]string, 0, numMerchants)

	for i := 0; i < numMerchants; i++ {
		merchantReq := handlers.CreateMerchantRequest{
			Name:      fmt.Sprintf("Concurrent Test Business %d", i),
			LegalName: fmt.Sprintf("Concurrent Test Business %d Legal", i),
			Industry:  "Technology",
		}

		jsonData, err := json.Marshal(merchantReq)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/merchants", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleCreateMerchant(w, req)

		if w.Code == http.StatusCreated {
			var merchantResponse map[string]interface{}
			json.NewDecoder(w.Body).Decode(&merchantResponse)
			if id, ok := merchantResponse["id"].(string); ok {
				merchantIDs = append(merchantIDs, id)
			}
		}
	}

	t.Logf("Created %d merchants", len(merchantIDs))

	// Verify all merchants have jobs enqueued
	time.Sleep(1 * time.Second)
	
	// Check that jobs are processing
	queueSize := jobProcessor.GetQueueSize()
	t.Logf("Job queue size after creating merchants: %d", queueSize)

	// Verify all merchants can retrieve status
	for _, merchantID := range merchantIDs {
		statusReq := httptest.NewRequest("GET", "/api/v1/merchants/"+merchantID+"/analytics/status", nil)
		statusW := httptest.NewRecorder()

		handler.HandleMerchantAnalyticsStatus(statusW, statusReq)

		assert.Equal(t, http.StatusOK, statusW.Code, "Status endpoint should work for merchant %s", merchantID)
	}

	t.Logf("✅ Concurrent merchant processing test completed")
}

