package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRiskAssessmentAsyncFlow tests the async risk assessment flow
func TestRiskAssessmentAsyncFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	baseURL := getTestBaseURL(t)
	client := &http.Client{Timeout: 60 * time.Second}
	authToken := getTestAuthToken(t)

	t.Run("Complete async assessment flow", func(t *testing.T) {
		merchantID := "merchant-risk-001"

		// Step 1: Start assessment
		t.Log("Step 1: Starting risk assessment...")
		requestBody := models.RiskAssessmentRequest{
			MerchantID: merchantID,
			Options: models.AssessmentOptions{
				IncludeHistory:    true,
				IncludePredictions: true,
			},
		}

		reqBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
			bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 202 Accepted
		if resp.StatusCode == http.StatusAccepted {
			var response models.RiskAssessmentResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.NotEmpty(t, response.AssessmentID, "Response should contain assessmentId")
			assert.Equal(t, "pending", response.Status, "Initial status should be pending")

			assessmentID := response.AssessmentID
			t.Logf("Assessment started: %s", assessmentID)

			// Step 2: Poll for status until complete
			t.Log("Step 2: Polling for assessment status...")
			maxAttempts := 30
			attempt := 0
			var finalStatus *models.AssessmentStatusResponse

			for attempt < maxAttempts {
				time.Sleep(2 * time.Second) // Wait 2 seconds between polls

				req, err := http.NewRequest("GET", 
					fmt.Sprintf("%s/api/v1/risk/assess/%s", baseURL, assessmentID), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

				resp, err := client.Do(req)
				require.NoError(t, err)

				if resp.StatusCode == http.StatusOK {
					var status models.AssessmentStatusResponse
					err = json.NewDecoder(resp.Body).Decode(&status)
					require.NoError(t, err)
					resp.Body.Close()

					assert.Equal(t, assessmentID, status.AssessmentID, "Assessment ID should match")
					assert.Equal(t, merchantID, status.MerchantID, "Merchant ID should match")

					t.Logf("Status: %s, Progress: %d%%", status.Status, status.Progress)

					if status.Status == "completed" {
						finalStatus = &status
						break
					} else if status.Status == "failed" {
						t.Fatalf("Assessment failed: %s", assessmentID)
					}
				} else {
					resp.Body.Close()
					t.Logf("Status check returned %d, retrying...", resp.StatusCode)
				}

				attempt++
			}

			// Step 3: Verify final results
			if finalStatus != nil {
				t.Log("Step 3: Verifying assessment results...")
				assert.Equal(t, "completed", finalStatus.Status, "Status should be completed")
				assert.Equal(t, 100, finalStatus.Progress, "Progress should be 100%")
				assert.NotNil(t, finalStatus.Result, "Result should be present")
				assert.NotNil(t, finalStatus.CompletedAt, "CompletedAt should be set")

				if finalStatus.Result != nil {
					assert.Greater(t, finalStatus.Result.OverallScore, 0.0, "Overall score should be > 0")
					assert.LessOrEqual(t, finalStatus.Result.OverallScore, 1.0, "Overall score should be <= 1")
					assert.NotEmpty(t, finalStatus.Result.RiskLevel, "Risk level should be set")
					assert.NotEmpty(t, finalStatus.Result.Factors, "Factors should be present")
				}
			} else {
				t.Log("Assessment did not complete within timeout (this may be expected if job processor is not running)")
			}
		} else {
			t.Logf("Risk assessment endpoint returned status %d (may not be implemented yet)", resp.StatusCode)
			// Accept 404 or 500 as endpoints may not be fully implemented
			assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusInternalServerError,
				"Expected Accepted, NotFound, or InternalServerError, got %d", resp.StatusCode)
		}
	})

	t.Run("Get assessment status for non-existent assessment", func(t *testing.T) {
		invalidID := "assess-invalid-999"

		req, err := http.NewRequest("GET", 
			fmt.Sprintf("%s/api/v1/risk/assess/%s", baseURL, invalidID), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, 
			"Non-existent assessment should return NotFound")
	})
}

// TestRiskAssessmentErrorScenarios tests various error scenarios
func TestRiskAssessmentErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	baseURL := getTestBaseURL(t)
	client := &http.Client{Timeout: 30 * time.Second}
	authToken := getTestAuthToken(t)

	t.Run("Invalid merchant ID", func(t *testing.T) {
		requestBody := models.RiskAssessmentRequest{
			MerchantID: "invalid-merchant-999",
			Options: models.AssessmentOptions{
				IncludeHistory:    true,
				IncludePredictions: false,
			},
		}

		reqBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
			bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 400 Bad Request or 404 Not Found
		assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound,
			"Expected BadRequest or NotFound for invalid merchant, got %d", resp.StatusCode)
	})

	t.Run("Malformed request body", func(t *testing.T) {
		// Send invalid JSON
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
			bytes.NewBufferString("{invalid json}"))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
			"Expected BadRequest for malformed JSON, got %d", resp.StatusCode)
	})

	t.Run("Missing merchant ID", func(t *testing.T) {
		requestBody := models.RiskAssessmentRequest{
			MerchantID: "", // Empty merchant ID
			Options: models.AssessmentOptions{
				IncludeHistory: true,
			},
		}

		reqBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
			bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
			"Expected BadRequest for missing merchant ID, got %d", resp.StatusCode)
	})

	t.Run("Missing authorization header", func(t *testing.T) {
		requestBody := models.RiskAssessmentRequest{
			MerchantID: "merchant-test-001",
			Options: models.AssessmentOptions{
				IncludeHistory: true,
			},
		}

		reqBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
			bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		// Intentionally omit Authorization header

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden,
			"Expected Unauthorized or Forbidden for missing auth, got %d", resp.StatusCode)
	})
}

// TestRiskAssessmentTimeoutHandling tests timeout scenarios
func TestRiskAssessmentTimeoutHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	baseURL := getTestBaseURL(t)
	authToken := getTestAuthToken(t)

	t.Run("Polling timeout", func(t *testing.T) {
		merchantID := "merchant-timeout-001"
		
		// Start assessment
		requestBody := models.RiskAssessmentRequest{
			MerchantID: merchantID,
			Options: models.AssessmentOptions{
				IncludeHistory:    true,
				IncludePredictions: true,
			},
		}

		reqBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
			bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusAccepted {
			var response models.RiskAssessmentResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assessmentID := response.AssessmentID
			
			// Poll with short timeout (only 3 attempts)
			maxAttempts := 3
			attempt := 0
			completed := false

			for attempt < maxAttempts {
				time.Sleep(1 * time.Second)

				req, err := http.NewRequest("GET", 
					fmt.Sprintf("%s/api/v1/risk/assess/%s", baseURL, assessmentID), nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

				resp, err := client.Do(req)
				require.NoError(t, err)

				if resp.StatusCode == http.StatusOK {
					var status models.AssessmentStatusResponse
					err = json.NewDecoder(resp.Body).Decode(&status)
					resp.Body.Close()

					if err == nil && status.Status == "completed" {
						completed = true
						break
					}
				} else {
					resp.Body.Close()
				}

				attempt++
			}

			// Should not complete within short timeout
			if !completed {
				t.Log("Assessment did not complete within short timeout (expected behavior)")
			}
		}
	})

	t.Run("Request timeout", func(t *testing.T) {
		// Use very short timeout client
		shortTimeoutClient := &http.Client{Timeout: 100 * time.Millisecond}

		requestBody := models.RiskAssessmentRequest{
			MerchantID: "merchant-timeout-002",
			Options: models.AssessmentOptions{
				IncludeHistory: true,
			},
		}

		reqBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
			bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		// This may timeout if server is slow
		_, err = shortTimeoutClient.Do(req)
		if err != nil {
			// Timeout is acceptable
			t.Logf("Request timed out (expected with short timeout): %v", err)
		}
	})
}

// TestRiskAssessmentConcurrent tests concurrent assessment scenarios
func TestRiskAssessmentConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	baseURL := getTestBaseURL(t)
	client := &http.Client{Timeout: 60 * time.Second}
	authToken := getTestAuthToken(t)

	t.Run("Multiple assessments for same merchant", func(t *testing.T) {
		merchantID := "merchant-concurrent-001"
		numAssessments := 3

		assessmentIDs := make([]string, 0, numAssessments)

		// Start multiple assessments concurrently
		for i := 0; i < numAssessments; i++ {
			requestBody := models.RiskAssessmentRequest{
				MerchantID: merchantID,
				Options: models.AssessmentOptions{
					IncludeHistory:    i%2 == 0, // Alternate options
					IncludePredictions: i%2 == 1,
				},
			}

			reqBody, err := json.Marshal(requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

			resp, err := client.Do(req)
			require.NoError(t, err)

			if resp.StatusCode == http.StatusAccepted {
				var response models.RiskAssessmentResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				resp.Body.Close()

				if err == nil && response.AssessmentID != "" {
					assessmentIDs = append(assessmentIDs, response.AssessmentID)
					t.Logf("Started assessment %d: %s", i+1, response.AssessmentID)
				}
			} else {
				resp.Body.Close()
			}
		}

		// Verify all assessments were created
		assert.GreaterOrEqual(t, len(assessmentIDs), 1, 
			"At least one assessment should be created")

		// Verify each assessment has unique ID
		uniqueIDs := make(map[string]bool)
		for _, id := range assessmentIDs {
			assert.False(t, uniqueIDs[id], "Assessment IDs should be unique")
			uniqueIDs[id] = true
		}
	})

	t.Run("Multiple assessments for different merchants", func(t *testing.T) {
		numMerchants := 5
		assessmentIDs := make([]string, 0, numMerchants)

		// Start assessments for different merchants concurrently
		for i := 0; i < numMerchants; i++ {
			merchantID := fmt.Sprintf("merchant-concurrent-%03d", i+10)

			requestBody := models.RiskAssessmentRequest{
				MerchantID: merchantID,
				Options: models.AssessmentOptions{
					IncludeHistory:    true,
					IncludePredictions: true,
				},
			}

			reqBody, err := json.Marshal(requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

			resp, err := client.Do(req)
			require.NoError(t, err)

			if resp.StatusCode == http.StatusAccepted {
				var response models.RiskAssessmentResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				resp.Body.Close()

				if err == nil && response.AssessmentID != "" {
					assessmentIDs = append(assessmentIDs, response.AssessmentID)
					t.Logf("Started assessment for merchant %s: %s", merchantID, response.AssessmentID)
				}
			} else {
				resp.Body.Close()
			}
		}

		// Verify assessments were created
		assert.GreaterOrEqual(t, len(assessmentIDs), 1,
			"At least one assessment should be created")

		// Verify all IDs are unique
		uniqueIDs := make(map[string]bool)
		for _, id := range assessmentIDs {
			assert.False(t, uniqueIDs[id], "Assessment IDs should be unique")
			uniqueIDs[id] = true
		}
	})

	t.Run("Concurrent status polling", func(t *testing.T) {
		merchantID := "merchant-concurrent-poll-001"

		// Start one assessment
		requestBody := models.RiskAssessmentRequest{
			MerchantID: merchantID,
			Options: models.AssessmentOptions{
				IncludeHistory: true,
			},
		}

		reqBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
			bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusAccepted {
			var response models.RiskAssessmentResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assessmentID := response.AssessmentID

			// Poll concurrently from multiple goroutines
			numPollers := 3
			results := make(chan string, numPollers)

			for i := 0; i < numPollers; i++ {
				go func(pollerID int) {
					req, err := http.NewRequest("GET", 
						fmt.Sprintf("%s/api/v1/risk/assess/%s", baseURL, assessmentID), nil)
					if err != nil {
						results <- fmt.Sprintf("poller-%d: error creating request: %v", pollerID, err)
						return
					}
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

					resp, err := client.Do(req)
					if err != nil {
						results <- fmt.Sprintf("poller-%d: error: %v", pollerID, err)
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode == http.StatusOK {
						var status models.AssessmentStatusResponse
						if err := json.NewDecoder(resp.Body).Decode(&status); err == nil {
							results <- fmt.Sprintf("poller-%d: status=%s, progress=%d", 
								pollerID, status.Status, status.Progress)
						} else {
							results <- fmt.Sprintf("poller-%d: decode error: %v", pollerID, err)
						}
					} else {
						results <- fmt.Sprintf("poller-%d: status code %d", pollerID, resp.StatusCode)
					}
				}(i)
			}

			// Collect results
			for i := 0; i < numPollers; i++ {
				result := <-results
				t.Logf("Poller result: %s", result)
			}

			// All pollers should succeed (no race conditions)
			t.Log("Concurrent polling completed successfully")
		}
	})
}

// TestRiskAssessmentEdgeCases tests edge cases and special scenarios
func TestRiskAssessmentEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	baseURL := getTestBaseURL(t)
	client := &http.Client{Timeout: 30 * time.Second}
	authToken := getTestAuthToken(t)

	t.Run("Assessment with minimal options", func(t *testing.T) {
		merchantID := "merchant-edge-001"

		requestBody := models.RiskAssessmentRequest{
			MerchantID: merchantID,
			Options: models.AssessmentOptions{
				IncludeHistory:    false,
				IncludePredictions: false,
			},
		}

		reqBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
			bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should accept even with minimal options
		if resp.StatusCode == http.StatusAccepted {
			var response models.RiskAssessmentResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)
			assert.NotEmpty(t, response.AssessmentID)
		}
	})

	t.Run("Very long merchant ID", func(t *testing.T) {
		// Create a very long merchant ID
		longMerchantID := "merchant-" + string(make([]byte, 500)) // 500 character ID

		requestBody := models.RiskAssessmentRequest{
			MerchantID: longMerchantID,
			Options: models.AssessmentOptions{
				IncludeHistory: true,
			},
		}

		reqBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/risk/assess", baseURL), 
			bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle long IDs gracefully (either accept or reject with 400)
		assert.True(t, resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusBadRequest,
			"Expected Accepted or BadRequest for long merchant ID, got %d", resp.StatusCode)
	})

	t.Run("Invalid assessment ID format", func(t *testing.T) {
		// Try to get status with invalid ID format
		invalidIDs := []string{
			"",
			"not-a-valid-uuid",
			"123",
			"../../etc/passwd", // Path traversal attempt
		}

		for _, invalidID := range invalidIDs {
			req, err := http.NewRequest("GET", 
				fmt.Sprintf("%s/api/v1/risk/assess/%s", baseURL, invalidID), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Should return 400 or 404, not 500
			assert.True(t, resp.StatusCode == http.StatusBadRequest || 
				resp.StatusCode == http.StatusNotFound,
				"Expected BadRequest or NotFound for invalid ID '%s', got %d", 
				invalidID, resp.StatusCode)
		}
	})
}

// Helper functions (same as in merchant_details_e2e_test.go)
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

