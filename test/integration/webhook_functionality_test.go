//go:build !comprehensive_test
// +build !comprehensive_test

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebhookFunctionality tests all webhook functionality
func (suite *IntegrationTestingSuite) TestWebhookFunctionality(t *testing.T) {
	t.Run("WebhookCreation", suite.testWebhookCreation)
	t.Run("WebhookEvents", suite.testWebhookEvents)
	t.Run("WebhookDelivery", suite.testWebhookDelivery)
	t.Run("WebhookErrorHandling", suite.testWebhookErrorHandling)
	t.Run("WebhookRetryMechanism", suite.testWebhookRetryMechanism)
}

// testWebhookCreation tests webhook creation functionality
func (suite *IntegrationTestingSuite) testWebhookCreation(t *testing.T) {
	tests := []struct {
		name           string
		webhook        Webhook
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_webhook_creation",
			webhook: Webhook{
				UserID:   "user_123",
				Name:     "Test Webhook",
				URL:      "https://example.com/webhook",
				Events:   []string{"business.created", "business.updated"},
				Secret:   "test_secret",
				IsActive: true,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid_webhook_url",
			webhook: Webhook{
				UserID:   "user_123",
				Name:     "Invalid Webhook",
				URL:      "invalid-url",
				Events:   []string{"business.created"},
				Secret:   "test_secret",
				IsActive: true,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid URL",
		},
		{
			name: "missing_webhook_events",
			webhook: Webhook{
				UserID:   "user_123",
				Name:     "No Events Webhook",
				URL:      "https://example.com/webhook",
				Events:   []string{},
				Secret:   "test_secret",
				IsActive: true,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "events are required",
		},
		{
			name: "duplicate_webhook_name",
			webhook: Webhook{
				UserID:   "user_123",
				Name:     "Duplicate Webhook", // This should conflict with existing
				URL:      "https://example.com/webhook2",
				Events:   []string{"business.created"},
				Secret:   "test_secret",
				IsActive: true,
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "webhook name already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.webhook)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", suite.server.URL+"/v1/webhooks",
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test_token")

			// Execute request
			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var errorResp map[string]string
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Contains(t, errorResp["error"], tt.expectedError)
			} else if resp.StatusCode == http.StatusCreated {
				var result map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result["id"])
				assert.Equal(t, tt.webhook.Name, result["name"])
				assert.Equal(t, tt.webhook.URL, result["url"])
			}
		})
	}
}

// testWebhookEvents tests webhook event functionality
func (suite *IntegrationTestingSuite) testWebhookEvents(t *testing.T) {
	// First create a webhook
	webhook := Webhook{
		UserID:   "user_123",
		Name:     "Event Test Webhook",
		URL:      "https://example.com/webhook",
		Events:   []string{"business.created", "business.updated", "risk.alert"},
		Secret:   "test_secret",
		IsActive: true,
	}

	webhookID := suite.createTestWebhook(t, webhook)

	tests := []struct {
		name           string
		event          WebhookEventRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_business_created_event",
			event: WebhookEventRequest{
				EventType: "business.created",
				Data: map[string]interface{}{
					"business_id": "biz_123",
					"name":        "Test Business",
					"status":      "active",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful_risk_alert_event",
			event: WebhookEventRequest{
				EventType: "risk.alert",
				Data: map[string]interface{}{
					"alert_id":    "alert_123",
					"risk_level":  "high",
					"business_id": "biz_123",
					"message":     "High risk detected",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unsupported_event_type",
			event: WebhookEventRequest{
				EventType: "unsupported.event",
				Data: map[string]interface{}{
					"test": "data",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "unsupported event type",
		},
		{
			name: "invalid_event_data",
			event: WebhookEventRequest{
				EventType: "business.created",
				Data:      nil,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "event data is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.event)
			require.NoError(t, err)

			req, err := http.NewRequest("POST",
				fmt.Sprintf("%s/v1/webhooks/%s/events", suite.server.URL, webhookID),
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test_token")

			// Execute request
			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var errorResp map[string]string
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Contains(t, errorResp["error"], tt.expectedError)
			} else if resp.StatusCode == http.StatusOK {
				var result map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result["event_id"])
				assert.Equal(t, tt.event.EventType, result["event_type"])
			}
		})
	}
}

// testWebhookDelivery tests webhook delivery functionality
func (suite *IntegrationTestingSuite) testWebhookDelivery(t *testing.T) {
	// Create a test webhook with a mock endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify webhook signature
		signature := r.Header.Get("X-Webhook-Signature")
		assert.NotEmpty(t, signature)

		// Verify content type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse and verify payload
		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)

		assert.NotEmpty(t, payload["event_type"])
		assert.NotEmpty(t, payload["data"])
		assert.NotEmpty(t, payload["timestamp"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "received"})
	}))
	defer mockServer.Close()

	// Create webhook pointing to mock server
	webhook := Webhook{
		UserID:   "user_123",
		Name:     "Delivery Test Webhook",
		URL:      mockServer.URL,
		Events:   []string{"business.created"},
		Secret:   "test_secret",
		IsActive: true,
	}

	webhookID := suite.createTestWebhook(t, webhook)

	// Test webhook delivery
	event := WebhookEventRequest{
		EventType: "business.created",
		Data: map[string]interface{}{
			"business_id": "biz_123",
			"name":        "Test Business",
		},
	}

	reqBody, err := json.Marshal(event)
	require.NoError(t, err)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/v1/webhooks/%s/events", suite.server.URL, webhookID),
		bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test_token")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify webhook was delivered (check webhook events)
	time.Sleep(100 * time.Millisecond) // Allow for async delivery

	events, err := suite.mockDB.GetWebhookEventsByWebhookID(context.Background(), webhookID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "delivered", events[0].Status)
}

// testWebhookErrorHandling tests webhook error handling
func (suite *IntegrationTestingSuite) testWebhookErrorHandling(t *testing.T) {
	// Create a mock server that returns errors
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
	}))
	defer mockServer.Close()

	// Create webhook pointing to error server
	webhook := Webhook{
		UserID:   "user_123",
		Name:     "Error Test Webhook",
		URL:      mockServer.URL,
		Events:   []string{"business.created"},
		Secret:   "test_secret",
		IsActive: true,
	}

	webhookID := suite.createTestWebhook(t, webhook)

	// Test webhook with error response
	event := WebhookEventRequest{
		EventType: "business.created",
		Data: map[string]interface{}{
			"business_id": "biz_123",
			"name":        "Test Business",
		},
	}

	reqBody, err := json.Marshal(event)
	require.NoError(t, err)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/v1/webhooks/%s/events", suite.server.URL, webhookID),
		bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test_token")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should still return OK (event was queued)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify webhook event shows error status
	time.Sleep(100 * time.Millisecond) // Allow for async processing

	events, err := suite.mockDB.GetWebhookEventsByWebhookID(context.Background(), webhookID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "failed", events[0].Status)
	assert.NotEmpty(t, events[0].ErrorMessage)
}

// testWebhookRetryMechanism tests webhook retry mechanism
func (suite *IntegrationTestingSuite) testWebhookRetryMechanism(t *testing.T) {
	retryCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		if retryCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "service unavailable"})
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "received"})
		}
	}))
	defer mockServer.Close()

	// Create webhook with retry configuration
	webhook := Webhook{
		UserID:   "user_123",
		Name:     "Retry Test Webhook",
		URL:      mockServer.URL,
		Events:   []string{"business.created"},
		Secret:   "test_secret",
		IsActive: true,
	}

	webhookID := suite.createTestWebhook(t, webhook)

	// Test webhook with retry scenario
	event := WebhookEventRequest{
		EventType: "business.created",
		Data: map[string]interface{}{
			"business_id": "biz_123",
			"name":        "Test Business",
		},
	}

	reqBody, err := json.Marshal(event)
	require.NoError(t, err)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/v1/webhooks/%s/events", suite.server.URL, webhookID),
		bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test_token")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Wait for retries to complete
	time.Sleep(500 * time.Millisecond)

	// Verify webhook was eventually delivered
	events, err := suite.mockDB.GetWebhookEventsByWebhookID(context.Background(), webhookID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "delivered", events[0].Status)
	assert.Equal(t, 3, retryCount) // Should have retried 3 times
}

// Helper methods
func (suite *IntegrationTestingSuite) createTestWebhook(t *testing.T, webhook Webhook) string {
	reqBody, err := json.Marshal(webhook)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", suite.server.URL+"/v1/webhooks",
		bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test_token")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result["id"].(string)
}

// Webhook represents a webhook configuration
type Webhook struct {
	UserID   string   `json:"user_id"`
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Events   []string `json:"events"`
	Secret   string   `json:"secret"`
	IsActive bool     `json:"is_active"`
}

// WebhookEventRequest represents a webhook event request
type WebhookEventRequest struct {
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
}
