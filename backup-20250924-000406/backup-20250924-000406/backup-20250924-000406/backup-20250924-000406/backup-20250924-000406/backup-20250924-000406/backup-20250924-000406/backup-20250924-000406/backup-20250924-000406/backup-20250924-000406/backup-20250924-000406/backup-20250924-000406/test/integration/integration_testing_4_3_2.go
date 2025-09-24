package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestingSuite handles comprehensive integration testing for Task 4.3.2
type IntegrationTestingSuite struct {
	server                *httptest.Server
	mockDB                interface{} // Mock database interface
	classificationService interface{} // Classification service interface
	alertManager          interface{} // Alert manager interface
	notificationService   interface{} // Notification service interface
	webhookHandler        interface{} // Webhook handler interface
}

// NewIntegrationTestingSuite creates a new integration testing suite
func NewIntegrationTestingSuite(t *testing.T) *IntegrationTestingSuite {
	// Create test server with basic routes
	mux := http.NewServeMux()
	registerIntegrationTestRoutes(mux)
	server := httptest.NewServer(mux)

	return &IntegrationTestingSuite{
		server:                server,
		mockDB:                nil, // Will be initialized as needed
		classificationService: nil,
		alertManager:          nil,
		notificationService:   nil,
		webhookHandler:        nil,
	}
}

// registerIntegrationTestRoutes registers all routes for integration testing
func registerIntegrationTestRoutes(mux *http.ServeMux) {
	// Webhook endpoints
	mux.HandleFunc("POST /v1/webhooks", mockWebhookHandler)
	mux.HandleFunc("GET /v1/webhooks/{id}", mockWebhookHandler)
	mux.HandleFunc("PUT /v1/webhooks/{id}", mockWebhookHandler)
	mux.HandleFunc("DELETE /v1/webhooks/{id}", mockWebhookHandler)
	mux.HandleFunc("POST /v1/webhooks/{id}/test", mockWebhookHandler)

	// External service integration endpoints
	mux.HandleFunc("POST /v1/external/website-scrape", mockWebsiteScrapingHandler)
	mux.HandleFunc("POST /v1/external/business-data", mockBusinessDataHandler)
	mux.HandleFunc("POST /v1/external/ml-classification", mockMLClassificationHandler)

	// Notification endpoints
	mux.HandleFunc("POST /v1/notifications/email", mockEmailNotificationHandler)
	mux.HandleFunc("POST /v1/notifications/sms", mockSMSNotificationHandler)
	mux.HandleFunc("POST /v1/notifications/slack", mockSlackNotificationHandler)
	mux.HandleFunc("POST /v1/notifications/webhook", mockWebhookNotificationHandler)

	// Reporting endpoints
	mux.HandleFunc("GET /v1/reports/performance", mockPerformanceReportHandler)
	mux.HandleFunc("GET /v1/reports/compliance", mockComplianceReportHandler)
	mux.HandleFunc("GET /v1/reports/risk", mockRiskReportHandler)
}

// Mock handlers for testing
func mockWebhookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func mockWebsiteScrapingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func mockBusinessDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func mockMLClassificationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func mockEmailNotificationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func mockSMSNotificationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func mockSlackNotificationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func mockWebhookNotificationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func mockPerformanceReportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func mockComplianceReportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func mockRiskReportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// TestExternalServiceIntegrations tests all external service integrations
func (suite *IntegrationTestingSuite) TestExternalServiceIntegrations(t *testing.T) {
	t.Run("WebsiteScrapingIntegration", suite.testWebsiteScrapingIntegration)
	t.Run("BusinessDataAPIIntegration", suite.testBusinessDataAPIIntegration)
	t.Run("MLClassificationIntegration", suite.testMLClassificationIntegration)
}

// testWebsiteScrapingIntegration tests website scraping external service integration
func (suite *IntegrationTestingSuite) testWebsiteScrapingIntegration(t *testing.T) {
	tests := []struct {
		name           string
		request        WebsiteScrapingRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_website_scraping",
			request: WebsiteScrapingRequest{
				URL:     "https://example.com",
				Timeout: 30,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid_url",
			request: WebsiteScrapingRequest{
				URL:     "invalid-url",
				Timeout: 30,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid URL",
		},
		{
			name: "timeout_scenario",
			request: WebsiteScrapingRequest{
				URL:     "https://httpstat.us/200?sleep=5000",
				Timeout: 1,
			},
			expectedStatus: http.StatusRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", suite.server.URL+"/v1/external/website-scrape",
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			client := &http.Client{Timeout: 60 * time.Second}
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
				var result WebsiteScrapingResponse
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result.Content)
				assert.NotEmpty(t, result.Metadata)
			}
		})
	}
}

// testBusinessDataAPIIntegration tests business data API external service integration
func (suite *IntegrationTestingSuite) testBusinessDataAPIIntegration(t *testing.T) {
	tests := []struct {
		name           string
		request        BusinessDataRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_business_data_lookup",
			request: BusinessDataRequest{
				BusinessName: "Acme Corporation",
				Address:      "123 Main St, Anytown, ST 12345",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing_business_name",
			request: BusinessDataRequest{
				Address: "123 Main St, Anytown, ST 12345",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "business name is required",
		},
		{
			name: "invalid_business_data",
			request: BusinessDataRequest{
				BusinessName: "",
				Address:      "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid business data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", suite.server.URL+"/v1/external/business-data",
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

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
				var result BusinessDataResponse
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result.BusinessID)
				assert.NotEmpty(t, result.VerificationStatus)
			}
		})
	}
}

// testMLClassificationIntegration tests ML classification external service integration
func (suite *IntegrationTestingSuite) testMLClassificationIntegration(t *testing.T) {
	tests := []struct {
		name           string
		request        MLClassificationRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful_ml_classification",
			request: MLClassificationRequest{
				BusinessName: "Tech Startup Inc",
				Description:  "Software development and technology consulting",
				WebsiteURL:   "https://techstartup.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing_business_name",
			request: MLClassificationRequest{
				Description: "Software development and technology consulting",
				WebsiteURL:  "https://techstartup.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "business name is required",
		},
		{
			name: "ml_service_timeout",
			request: MLClassificationRequest{
				BusinessName: "Test Business",
				Description:  "Test description",
				WebsiteURL:   "https://test.com",
				Timeout:      1, // Very short timeout to test timeout handling
			},
			expectedStatus: http.StatusRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", suite.server.URL+"/v1/external/ml-classification",
				bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			client := &http.Client{Timeout: 60 * time.Second}
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
				var result MLClassificationResponse
				err = json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result.Classification)
				assert.Greater(t, result.ConfidenceScore, 0.0)
				assert.NotEmpty(t, result.IndustryCodes)
			}
		})
	}
}

// Request/Response types for integration testing
type WebsiteScrapingRequest struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout"`
}

type WebsiteScrapingResponse struct {
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Success  bool                   `json:"success"`
}

type BusinessDataRequest struct {
	BusinessName string `json:"business_name"`
	Address      string `json:"address"`
	Phone        string `json:"phone,omitempty"`
	Email        string `json:"email,omitempty"`
}

type BusinessDataResponse struct {
	BusinessID         string  `json:"business_id"`
	VerificationStatus string  `json:"verification_status"`
	RiskScore          float64 `json:"risk_score"`
	ComplianceStatus   string  `json:"compliance_status"`
}

type MLClassificationRequest struct {
	BusinessName string `json:"business_name"`
	Description  string `json:"description"`
	WebsiteURL   string `json:"website_url"`
	Timeout      int    `json:"timeout,omitempty"`
}

type MLClassificationResponse struct {
	Classification  string                 `json:"classification"`
	ConfidenceScore float64                `json:"confidence_score"`
	IndustryCodes   map[string]interface{} `json:"industry_codes"`
	RiskAssessment  map[string]interface{} `json:"risk_assessment"`
	ProcessingTime  int64                  `json:"processing_time_ms"`
}
