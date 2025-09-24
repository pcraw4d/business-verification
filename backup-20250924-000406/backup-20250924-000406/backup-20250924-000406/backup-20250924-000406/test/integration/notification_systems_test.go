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

	"github.com/pcraw4d/business-verification/internal/risk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNotificationSystems tests all notification systems
func (suite *IntegrationTestingSuite) TestNotificationSystems(t *testing.T) {
	t.Run("EmailNotifications", suite.testEmailNotifications)
	t.Run("SMSNotifications", suite.testSMSNotifications)
	t.Run("SlackNotifications", suite.testSlackNotifications)
	t.Run("WebhookNotifications", suite.testWebhookNotifications)
	t.Run("NotificationChannels", suite.testNotificationChannels)
	t.Run("NotificationTemplates", suite.testNotificationTemplates)
}

// testEmailNotifications tests email notification functionality
func (suite *IntegrationTestingSuite) testEmailNotifications(t *testing.T) {
	tests := []struct {
		name           string
		request        EmailNotificationRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_email_notification",
			request: EmailNotificationRequest{
				To:      []string{"test@example.com"},
				Subject: "Test Email Notification",
				Body:    "This is a test email notification",
				Type:    "business_alert",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "email_with_template",
			request: EmailNotificationRequest{
				To:       []string{"test@example.com"},
				Template: "business_verification_complete",
				Data: map[string]interface{}{
					"business_name":       "Test Business",
					"verification_status": "approved",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid_email_address",
			request: EmailNotificationRequest{
				To:      []string{"invalid-email"},
				Subject: "Test Email",
				Body:    "Test body",
				Type:    "business_alert",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid email address",
		},
		{
			name: "missing_recipients",
			request: EmailNotificationRequest{
				To:      []string{},
				Subject: "Test Email",
				Body:    "Test body",
				Type:    "business_alert",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "recipients are required",
		},
		{
			name: "email_service_unavailable",
			request: EmailNotificationRequest{
				To:           []string{"test@example.com"},
				Subject:      "Test Email",
				Body:         "Test body",
				Type:         "business_alert",
				ForceFailure: true, // Simulate service failure
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", suite.server.URL+"/v1/notifications/email",
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
				var result EmailNotificationResponse
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result.MessageID)
				assert.Equal(t, "sent", result.Status)
				assert.NotEmpty(t, result.SentAt)
			}
		})
	}
}

// testSMSNotifications tests SMS notification functionality
func (suite *IntegrationTestingSuite) testSMSNotifications(t *testing.T) {
	tests := []struct {
		name           string
		request        SMSNotificationRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_sms_notification",
			request: SMSNotificationRequest{
				To:      []string{"+1234567890"},
				Message: "Test SMS notification",
				Type:    "risk_alert",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "sms_with_template",
			request: SMSNotificationRequest{
				To:       []string{"+1234567890"},
				Template: "verification_code",
				Data: map[string]interface{}{
					"code":       "123456",
					"expires_in": "5 minutes",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid_phone_number",
			request: SMSNotificationRequest{
				To:      []string{"invalid-phone"},
				Message: "Test SMS",
				Type:    "risk_alert",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid phone number",
		},
		{
			name: "message_too_long",
			request: SMSNotificationRequest{
				To:      []string{"+1234567890"},
				Message: string(make([]byte, 2000)), // Very long message
				Type:    "risk_alert",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "message too long",
		},
		{
			name: "sms_service_unavailable",
			request: SMSNotificationRequest{
				To:           []string{"+1234567890"},
				Message:      "Test SMS",
				Type:         "risk_alert",
				ForceFailure: true, // Simulate service failure
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", suite.server.URL+"/v1/notifications/sms",
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
				var result SMSNotificationResponse
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result.MessageID)
				assert.Equal(t, "sent", result.Status)
				assert.NotEmpty(t, result.SentAt)
			}
		})
	}
}

// testSlackNotifications tests Slack notification functionality
func (suite *IntegrationTestingSuite) testSlackNotifications(t *testing.T) {
	// Create a mock Slack webhook server
	mockSlackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Slack webhook format
		var slackPayload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&slackPayload)
		require.NoError(t, err)

		assert.NotEmpty(t, slackPayload["text"])
		assert.NotEmpty(t, slackPayload["channel"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer mockSlackServer.Close()

	tests := []struct {
		name           string
		request        SlackNotificationRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_slack_notification",
			request: SlackNotificationRequest{
				Channel: "#alerts",
				Text:    "Test Slack notification",
				Type:    "risk_alert",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "slack_with_attachments",
			request: SlackNotificationRequest{
				Channel: "#alerts",
				Text:    "Risk Alert",
				Type:    "risk_alert",
				Attachments: []SlackAttachment{
					{
						Color: "danger",
						Fields: []SlackField{
							{
								Title: "Business ID",
								Value: "biz_123",
								Short: true,
							},
							{
								Title: "Risk Level",
								Value: "High",
								Short: true,
							},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid_slack_channel",
			request: SlackNotificationRequest{
				Channel: "invalid-channel",
				Text:    "Test message",
				Type:    "risk_alert",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid channel format",
		},
		{
			name: "missing_slack_text",
			request: SlackNotificationRequest{
				Channel: "#alerts",
				Text:    "",
				Type:    "risk_alert",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "text is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", suite.server.URL+"/v1/notifications/slack",
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
				var result SlackNotificationResponse
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result.MessageID)
				assert.Equal(t, "sent", result.Status)
				assert.NotEmpty(t, result.SentAt)
			}
		})
	}
}

// testWebhookNotifications tests webhook notification functionality
func (suite *IntegrationTestingSuite) testWebhookNotifications(t *testing.T) {
	// Create a mock webhook server
	mockWebhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer mockWebhookServer.Close()

	tests := []struct {
		name           string
		request        WebhookNotificationRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_webhook_notification",
			request: WebhookNotificationRequest{
				URL:       mockWebhookServer.URL,
				EventType: "business.verified",
				Data: map[string]interface{}{
					"business_id": "biz_123",
					"status":      "verified",
				},
				Secret: "test_secret",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "webhook_with_custom_headers",
			request: WebhookNotificationRequest{
				URL:       mockWebhookServer.URL,
				EventType: "risk.alert",
				Data: map[string]interface{}{
					"alert_id":   "alert_123",
					"risk_level": "high",
				},
				Secret: "test_secret",
				Headers: map[string]string{
					"X-Custom-Header": "custom-value",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid_webhook_url",
			request: WebhookNotificationRequest{
				URL:       "invalid-url",
				EventType: "business.verified",
				Data: map[string]interface{}{
					"business_id": "biz_123",
				},
				Secret: "test_secret",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid URL",
		},
		{
			name: "webhook_timeout",
			request: WebhookNotificationRequest{
				URL:       "https://httpstat.us/200?sleep=5000", // 5 second delay
				EventType: "business.verified",
				Data: map[string]interface{}{
					"business_id": "biz_123",
				},
				Secret:  "test_secret",
				Timeout: 1, // 1 second timeout
			},
			expectedStatus: http.StatusRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", suite.server.URL+"/v1/notifications/webhook",
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
				var result WebhookNotificationResponse
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result.MessageID)
				assert.Equal(t, "sent", result.Status)
				assert.NotEmpty(t, result.SentAt)
			}
		})
	}
}

// testNotificationChannels tests notification channel management
func (suite *IntegrationTestingSuite) testNotificationChannels(t *testing.T) {
	// Test adding notification channels
	channels := []string{"email", "sms", "slack", "webhook"}

	for _, channel := range channels {
		t.Run(fmt.Sprintf("add_%s_channel", channel), func(t *testing.T) {
			// Test channel addition
			err := suite.notificationService.AddChannel(channel, &MockNotificationChannel{
				name:    channel,
				enabled: true,
			})
			require.NoError(t, err)

			// Verify channel was added
			addedChannel, err := suite.notificationService.GetChannel(channel)
			require.NoError(t, err)
			assert.Equal(t, channel, addedChannel.GetName())
		})
	}

	// Test channel enabling/disabling
	t.Run("channel_enable_disable", func(t *testing.T) {
		// Disable a channel
		err := suite.notificationService.DisableChannel("email")
		require.NoError(t, err)

		// Verify channel is disabled
		channel, err := suite.notificationService.GetChannel("email")
		require.NoError(t, err)
		assert.False(t, channel.IsEnabled())

		// Re-enable channel
		err = suite.notificationService.EnableChannel("email")
		require.NoError(t, err)

		// Verify channel is enabled
		channel, err = suite.notificationService.GetChannel("email")
		require.NoError(t, err)
		assert.True(t, channel.IsEnabled())
	})

	// Test channel removal
	t.Run("channel_removal", func(t *testing.T) {
		// Remove a channel
		suite.notificationService.RemoveChannel("webhook")

		// Verify channel was removed
		_, err := suite.notificationService.GetChannel("webhook")
		assert.Error(t, err)
	})
}

// testNotificationTemplates tests notification template functionality
func (suite *IntegrationTestingSuite) testNotificationTemplates(t *testing.T) {
	tests := []struct {
		name           string
		template       NotificationTemplateRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "create_email_template",
			template: NotificationTemplateRequest{
				Name:     "business_verification_complete",
				Type:     "email",
				Subject:  "Business Verification Complete",
				Body:     "Your business {{.business_name}} has been verified with status {{.status}}",
				Language: "en",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "create_sms_template",
			template: NotificationTemplateRequest{
				Name:     "verification_code",
				Type:     "sms",
				Message:  "Your verification code is {{.code}}. Expires in {{.expires_in}}",
				Language: "en",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "duplicate_template_name",
			template: NotificationTemplateRequest{
				Name:     "business_verification_complete", // Duplicate
				Type:     "email",
				Subject:  "Duplicate Template",
				Body:     "This is a duplicate",
				Language: "en",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "template already exists",
		},
		{
			name: "invalid_template_syntax",
			template: NotificationTemplateRequest{
				Name:     "invalid_template",
				Type:     "email",
				Subject:  "Invalid Template",
				Body:     "Invalid syntax {{.invalid_field", // Missing closing brace
				Language: "en",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid template syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.template)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", suite.server.URL+"/v1/notifications/templates",
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
				assert.NotEmpty(t, result["template_id"])
				assert.Equal(t, tt.template.Name, result["name"])
			}
		})
	}
}

// Mock notification channel for testing
type MockNotificationChannel struct {
	name    string
	enabled bool
}

func (m *MockNotificationChannel) GetName() string {
	return m.name
}

func (m *MockNotificationChannel) IsEnabled() bool {
	return m.enabled
}

func (m *MockNotificationChannel) Send(ctx context.Context, alert *risk.RiskAlert) error {
	return nil
}

// Request/Response types for notification testing
type EmailNotificationRequest struct {
	To           []string               `json:"to"`
	Subject      string                 `json:"subject,omitempty"`
	Body         string                 `json:"body,omitempty"`
	Template     string                 `json:"template,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Type         string                 `json:"type"`
	ForceFailure bool                   `json:"force_failure,omitempty"`
}

type EmailNotificationResponse struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
	SentAt    string `json:"sent_at"`
}

type SMSNotificationRequest struct {
	To           []string               `json:"to"`
	Message      string                 `json:"message,omitempty"`
	Template     string                 `json:"template,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Type         string                 `json:"type"`
	ForceFailure bool                   `json:"force_failure,omitempty"`
}

type SMSNotificationResponse struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
	SentAt    string `json:"sent_at"`
}

type SlackNotificationRequest struct {
	Channel     string            `json:"channel"`
	Text        string            `json:"text"`
	Type        string            `json:"type"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

type SlackAttachment struct {
	Color  string       `json:"color"`
	Fields []SlackField `json:"fields"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type SlackNotificationResponse struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
	SentAt    string `json:"sent_at"`
}

type WebhookNotificationRequest struct {
	URL       string                 `json:"url"`
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
	Secret    string                 `json:"secret"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Timeout   int                    `json:"timeout,omitempty"`
}

type WebhookNotificationResponse struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
	SentAt    string `json:"sent_at"`
}

type NotificationTemplateRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Subject  string `json:"subject,omitempty"`
	Body     string `json:"body,omitempty"`
	Message  string `json:"message,omitempty"`
	Language string `json:"language"`
}
