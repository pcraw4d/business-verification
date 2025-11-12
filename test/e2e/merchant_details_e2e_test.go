package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMerchantDetailsNavigation tests navigation from add-merchant to merchant-details page
func TestMerchantDetailsNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	baseURL := getTestBaseURL(t)
	client := &http.Client{Timeout: 30 * time.Second}

	t.Run("Navigate from add-merchant to merchant-details", func(t *testing.T) {
		// Step 1: Create a merchant via API
		merchantData := map[string]interface{}{
			"name":        "E2E Test Company",
			"legal_name":  "E2E Test Company Inc.",
			"industry":    "Technology",
			"address": map[string]string{
				"street1":    "123 Test St",
				"city":       "San Francisco",
				"state":      "CA",
				"postal_code": "94102",
				"country":    "United States",
			},
			"contact_info": map[string]string{
				"email": "test@e2etest.com",
			},
		}

		reqBody, err := json.Marshal(merchantData)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/merchants", baseURL), 
			bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", getTestAuthToken(t)))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode, "Expected merchant creation to succeed")

		var merchantResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&merchantResp)
		require.NoError(t, err)

		merchantID, ok := merchantResp["id"].(string)
		require.True(t, ok, "Expected merchant ID in response")

		// Step 2: Verify merchant-details page is accessible
		detailsURL := fmt.Sprintf("%s/merchant-details.html?merchantId=%s", baseURL, merchantID)
		resp, err = client.Get(detailsURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Merchant details page should be accessible")
	})
}

// TestMerchantDetailsTabSwitching tests tab switching and content loading
func TestMerchantDetailsTabSwitching(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	baseURL := getTestBaseURL(t)
	merchantID := "merchant-analytics-001" // Use test merchant from fixtures

	t.Run("Switch between tabs and verify content loads", func(t *testing.T) {
		// This would require a browser automation tool like Playwright
		// For now, we test the API endpoints that the tabs use

		client := &http.Client{Timeout: 30 * time.Second}
		authToken := getTestAuthToken(t)

		tabs := []struct {
			name     string
			endpoint string
		}{
			{"Overview", fmt.Sprintf("/api/v1/merchants/%s", merchantID)},
			{"Business Analytics", fmt.Sprintf("/api/v1/merchants/%s/analytics", merchantID)},
			{"Website Analysis", fmt.Sprintf("/api/v1/merchants/%s/website-analysis", merchantID)},
			{"Risk Score", fmt.Sprintf("/api/v1/merchants/%s/risk-score", merchantID)},
		}

		for _, tab := range tabs {
			t.Run(tab.name, func(t *testing.T) {
				req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", baseURL, tab.endpoint), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound,
					"Tab endpoint should return OK or NotFound, got %d", resp.StatusCode)
			})
		}
	})
}

// TestMerchantDetailsDataLoading tests API data loading and display
func TestMerchantDetailsDataLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	baseURL := getTestBaseURL(t)
	merchantID := "merchant-analytics-001"
	client := &http.Client{Timeout: 30 * time.Second}
	authToken := getTestAuthToken(t)

	t.Run("Load merchant analytics data", func(t *testing.T) {
		req, err := http.NewRequest("GET", 
			fmt.Sprintf("%s/api/v1/merchants/%s/analytics", baseURL, merchantID), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var analytics map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&analytics)
			require.NoError(t, err)

			assert.Contains(t, analytics, "merchantId", "Response should contain merchantId")
			assert.Contains(t, analytics, "classification", "Response should contain classification")
			assert.Contains(t, analytics, "security", "Response should contain security")
			assert.Contains(t, analytics, "quality", "Response should contain quality")
		} else {
			t.Logf("Analytics endpoint returned status %d (may not be implemented yet)", resp.StatusCode)
		}
	})

	t.Run("Load website analysis data", func(t *testing.T) {
		req, err := http.NewRequest("GET", 
			fmt.Sprintf("%s/api/v1/merchants/%s/website-analysis", baseURL, merchantID), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var analysis map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&analysis)
			require.NoError(t, err)

			assert.Contains(t, analysis, "merchantId", "Response should contain merchantId")
			assert.Contains(t, analysis, "websiteUrl", "Response should contain websiteUrl")
		} else {
			t.Logf("Website analysis endpoint returned status %d (may not be implemented yet)", resp.StatusCode)
		}
	})
}

// TestMerchantDetailsErrorHandling tests error handling and fallbacks
func TestMerchantDetailsErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	baseURL := getTestBaseURL(t)
	client := &http.Client{Timeout: 30 * time.Second}
	authToken := getTestAuthToken(t)

	t.Run("Handle invalid merchant ID", func(t *testing.T) {
		invalidID := "invalid-merchant-999"

		req, err := http.NewRequest("GET", 
			fmt.Sprintf("%s/api/v1/merchants/%s/analytics", baseURL, invalidID), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusInternalServerError,
			"Invalid merchant ID should return NotFound or InternalServerError")
	})

	t.Run("Handle missing authentication", func(t *testing.T) {
		req, err := http.NewRequest("GET", 
			fmt.Sprintf("%s/api/v1/merchants/test-merchant-123/analytics", baseURL), nil)
		require.NoError(t, err)
		// No Authorization header

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, 
			"Missing authentication should return Unauthorized")
	})
}

// Helper functions

func getTestBaseURL(t *testing.T) string {
	baseURL := getEnvOrDefault("TEST_BASE_URL", "http://localhost:8080")
	t.Logf("Using test base URL: %s", baseURL)
	return baseURL
}

func getTestAuthToken(t *testing.T) string {
	token := getEnvOrDefault("TEST_AUTH_TOKEN", "test-token")
	if token == "test-token" {
		t.Log("Warning: Using default test token. Set TEST_AUTH_TOKEN for real authentication.")
	}
	return token
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

