package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/external"
)

func TestNewVerificationReasoningHandler(t *testing.T) {
	logger := zap.NewNop()
	generator := external.NewVerificationReasoningGenerator(nil)

	handler := NewVerificationReasoningHandler(generator, logger)
	require.NotNil(t, handler)
	assert.Equal(t, generator, handler.generator)
	assert.Equal(t, logger, handler.logger)
}

func TestVerificationReasoningHandler_GenerateReasoning(t *testing.T) {
	logger := zap.NewNop()
	generator := external.NewVerificationReasoningGenerator(nil)
	handler := NewVerificationReasoningHandler(generator, logger)

	tests := []struct {
		name           string
		request        GenerateReasoningRequest
		expectedStatus int
		checkResponse  func(*testing.T, *GenerateReasoningResponse)
	}{
		{
			name: "valid request",
			request: GenerateReasoningRequest{
				VerificationID: "test-123",
				BusinessName:   "Test Business",
				WebsiteURL:     "https://test.com",
				Result: &external.VerificationResult{
					ID:           "test-123",
					Status:       external.StatusPassed,
					OverallScore: 0.85,
					FieldResults: map[string]external.FieldResult{
						"business_name": {
							Status:     external.StatusPassed,
							Score:      0.9,
							Confidence: 0.8,
							Matched:    true,
						},
					},
				},
				Comparison: &external.ComparisonResult{
					OverallScore: 0.85,
					FieldResults: map[string]external.FieldComparison{
						"business_name": {
							Score:      0.9,
							Confidence: 0.8,
							Matched:    true,
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response *GenerateReasoningResponse) {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Reasoning)
				assert.Equal(t, "PASSED", response.Reasoning.Status)
				assert.Equal(t, 0.85, response.Reasoning.OverallScore)
				assert.Equal(t, "high", response.Reasoning.ConfidenceLevel)
				assert.Equal(t, "test-123", response.Reasoning.VerificationID)
				assert.Equal(t, "Test Business", response.Reasoning.BusinessName)
				assert.Equal(t, "https://test.com", response.Reasoning.WebsiteURL)
			},
		},
		{
			name: "missing verification_id",
			request: GenerateReasoningRequest{
				BusinessName: "Test Business",
				Result: &external.VerificationResult{
					Status:       external.StatusPassed,
					OverallScore: 0.85,
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *GenerateReasoningResponse) {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, "verification_id is required")
			},
		},
		{
			name: "missing business_name",
			request: GenerateReasoningRequest{
				VerificationID: "test-123",
				Result: &external.VerificationResult{
					Status:       external.StatusPassed,
					OverallScore: 0.85,
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *GenerateReasoningResponse) {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, "business_name is required")
			},
		},
		{
			name: "missing result",
			request: GenerateReasoningRequest{
				VerificationID: "test-123",
				BusinessName:   "Test Business",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *GenerateReasoningResponse) {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, "result is required")
			},
		},
		{
			name: "nil result",
			request: GenerateReasoningRequest{
				VerificationID: "test-123",
				BusinessName:   "Test Business",
				Result:         nil,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *GenerateReasoningResponse) {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, "result is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest("POST", "/generate-reasoning", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GenerateReasoning(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response GenerateReasoningResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response
			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}

			// Check timestamp
			assert.NotZero(t, response.Timestamp)
		})
	}
}

func TestVerificationReasoningHandler_GenerateReport(t *testing.T) {
	logger := zap.NewNop()
	generator := external.NewVerificationReasoningGenerator(nil)
	handler := NewVerificationReasoningHandler(generator, logger)

	tests := []struct {
		name           string
		request        GenerateReportRequest
		expectedStatus int
		checkResponse  func(*testing.T, *GenerateReportResponse)
	}{
		{
			name: "valid request",
			request: GenerateReportRequest{
				VerificationID: "test-123",
				BusinessName:   "Test Business",
				WebsiteURL:     "https://test.com",
				IncludeAudit:   true,
				Result: &external.VerificationResult{
					ID:           "test-123",
					Status:       external.StatusPassed,
					OverallScore: 0.85,
					FieldResults: map[string]external.FieldResult{
						"business_name": {
							Status:     external.StatusPassed,
							Score:      0.9,
							Confidence: 0.8,
							Matched:    true,
						},
					},
				},
				Comparison: &external.ComparisonResult{
					OverallScore: 0.85,
					FieldResults: map[string]external.FieldComparison{
						"business_name": {
							Score:      0.9,
							Confidence: 0.8,
							Matched:    true,
						},
					},
				},
				Metadata: map[string]interface{}{
					"source": "api",
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response *GenerateReportResponse) {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Report)
				assert.Equal(t, "test-123", response.Report.VerificationID)
				assert.Equal(t, "Test Business", response.Report.BusinessName)
				assert.Equal(t, "https://test.com", response.Report.WebsiteURL)
				assert.Equal(t, "PASSED", response.Report.Status)
				assert.Equal(t, 0.85, response.Report.OverallScore)
				assert.Equal(t, "high", response.Report.ConfidenceLevel)
				assert.NotNil(t, response.Report.Reasoning)
				assert.NotNil(t, response.Report.ComparisonDetails)
				assert.True(t, len(response.Report.AuditTrail) >= 5) // Should have multiple audit events
				// Check that the last event is report generation
				lastEvent := response.Report.AuditTrail[len(response.Report.AuditTrail)-1]
				assert.Equal(t, "report_generated", lastEvent.EventType)
				assert.Equal(t, "info", lastEvent.Severity)
				assert.Equal(t, "api", response.Report.Metadata["source"])
			},
		},
		{
			name: "missing verification_id",
			request: GenerateReportRequest{
				BusinessName: "Test Business",
				Result: &external.VerificationResult{
					Status:       external.StatusPassed,
					OverallScore: 0.85,
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *GenerateReportResponse) {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, "verification_id is required")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest("POST", "/generate-report", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GenerateReport(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response GenerateReportResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response
			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}

			// Check timestamp
			assert.NotZero(t, response.Timestamp)
		})
	}
}

func TestVerificationReasoningHandler_GetConfig(t *testing.T) {
	logger := zap.NewNop()
	generator := external.NewVerificationReasoningGenerator(nil)
	handler := NewVerificationReasoningHandler(generator, logger)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/config", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.GetConfig(w, req)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response UpdateConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check response
	assert.True(t, response.Success)
	assert.NotNil(t, response.Config)
	assert.True(t, response.Config.EnableDetailedExplanations)
	assert.True(t, response.Config.EnableRiskAnalysis)
	assert.True(t, response.Config.EnableRecommendations)
	assert.True(t, response.Config.EnableAuditTrail)
	assert.Equal(t, 0.6, response.Config.MinConfidenceThreshold)
	assert.Equal(t, 0.8, response.Config.MaxRiskProbability)
	assert.Equal(t, "en", response.Config.Language)
	assert.NotZero(t, response.Timestamp)
}

func TestVerificationReasoningHandler_UpdateConfig(t *testing.T) {
	logger := zap.NewNop()
	generator := external.NewVerificationReasoningGenerator(nil)
	handler := NewVerificationReasoningHandler(generator, logger)

	tests := []struct {
		name           string
		request        UpdateConfigRequest
		expectedStatus int
		checkResponse  func(*testing.T, *UpdateConfigResponse)
	}{
		{
			name: "valid config update",
			request: UpdateConfigRequest{
				EnableDetailedExplanations: false,
				EnableRiskAnalysis:         false,
				EnableRecommendations:      true,
				EnableAuditTrail:           true,
				MinConfidenceThreshold:     0.8,
				MaxRiskProbability:         0.9,
				Language:                   "es",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response *UpdateConfigResponse) {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Config)
				assert.False(t, response.Config.EnableDetailedExplanations)
				assert.False(t, response.Config.EnableRiskAnalysis)
				assert.True(t, response.Config.EnableRecommendations)
				assert.True(t, response.Config.EnableAuditTrail)
				assert.Equal(t, 0.8, response.Config.MinConfidenceThreshold)
				assert.Equal(t, 0.9, response.Config.MaxRiskProbability)
				assert.Equal(t, "es", response.Config.Language)
			},
		},
		{
			name: "invalid min confidence threshold",
			request: UpdateConfigRequest{
				MinConfidenceThreshold: 1.5, // Invalid: > 1
				MaxRiskProbability:     0.8,
				Language:               "en",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *UpdateConfigResponse) {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, "min_confidence_threshold must be between 0 and 1")
			},
		},
		{
			name: "invalid max risk probability",
			request: UpdateConfigRequest{
				MinConfidenceThreshold: 0.6,
				MaxRiskProbability:     -0.1, // Invalid: < 0
				Language:               "en",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *UpdateConfigResponse) {
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, "max_risk_probability must be between 0 and 1")
			},
		},
		{
			name: "empty language defaults to en",
			request: UpdateConfigRequest{
				MinConfidenceThreshold: 0.6,
				MaxRiskProbability:     0.8,
				Language:               "",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response *UpdateConfigResponse) {
				assert.True(t, response.Success)
				assert.NotNil(t, response.Config)
				assert.Equal(t, "en", response.Config.Language)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest("PUT", "/config", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.UpdateConfig(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response UpdateConfigResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response
			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}

			// Check timestamp
			assert.NotZero(t, response.Timestamp)
		})
	}
}

func TestVerificationReasoningHandler_GetHealth(t *testing.T) {
	logger := zap.NewNop()
	generator := external.NewVerificationReasoningGenerator(nil)
	handler := NewVerificationReasoningHandler(generator, logger)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/health", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.GetHealth(w, req)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var health map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &health)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, "healthy", health["status"])
	assert.Equal(t, "verification_reasoning", health["service"])
	assert.NotNil(t, health["timestamp"])
	assert.True(t, health["config"].(bool))
}

func TestVerificationReasoningHandler_RegisterRoutes(t *testing.T) {
	logger := zap.NewNop()
	generator := external.NewVerificationReasoningGenerator(nil)
	handler := NewVerificationReasoningHandler(generator, logger)

	// Create router
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test that routes are registered
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			t.Logf("Registered route: %s", pathTemplate)
		}
		return nil
	})

	// Test that we can create requests to the registered endpoints
	endpoints := []struct {
		path   string
		method string
	}{
		{"/generate-reasoning", "POST"},
		{"/generate-report", "POST"},
		{"/config", "GET"},
		{"/config", "PUT"},
		{"/health", "GET"},
	}

	for _, endpoint := range endpoints {
		req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
		var match mux.RouteMatch
		assert.True(t, router.Match(req, &match), "Route %s %s should be registered", endpoint.method, endpoint.path)
	}
}

func TestVerificationReasoningHandler_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()
	generator := external.NewVerificationReasoningGenerator(nil)
	handler := NewVerificationReasoningHandler(generator, logger)

	tests := []struct {
		name           string
		endpoint       string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "invalid JSON in generate reasoning",
			endpoint:       "/generate-reasoning",
			method:         "POST",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON in generate report",
			endpoint:       "/generate-report",
			method:         "POST",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON in update config",
			endpoint:       "/config",
			method:         "PUT",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP request
			req := httptest.NewRequest(tt.method, tt.endpoint, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call appropriate handler method
			switch tt.endpoint {
			case "/generate-reasoning":
				handler.GenerateReasoning(w, req)
			case "/generate-report":
				handler.GenerateReport(w, req)
			case "/config":
				if tt.method == "PUT" {
					handler.UpdateConfig(w, req)
				} else {
					handler.GetConfig(w, req)
				}
			}

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestVerificationReasoningHandler_RequestStructs(t *testing.T) {
	// Test GenerateReasoningRequest
	req := GenerateReasoningRequest{
		VerificationID: "test-123",
		BusinessName:   "Test Business",
		WebsiteURL:     "https://test.com",
		Result: &external.VerificationResult{
			Status:       external.StatusPassed,
			OverallScore: 0.85,
		},
		Comparison: &external.ComparisonResult{
			OverallScore: 0.85,
		},
	}

	assert.Equal(t, "test-123", req.VerificationID)
	assert.Equal(t, "Test Business", req.BusinessName)
	assert.Equal(t, "https://test.com", req.WebsiteURL)
	assert.NotNil(t, req.Result)
	assert.NotNil(t, req.Comparison)

	// Test GenerateReportRequest
	reportReq := GenerateReportRequest{
		VerificationID: "test-123",
		BusinessName:   "Test Business",
		WebsiteURL:     "https://test.com",
		IncludeAudit:   true,
		Result: &external.VerificationResult{
			Status:       external.StatusPassed,
			OverallScore: 0.85,
		},
		Comparison: &external.ComparisonResult{
			OverallScore: 0.85,
		},
		Metadata: map[string]interface{}{
			"source": "test",
		},
	}

	assert.Equal(t, "test-123", reportReq.VerificationID)
	assert.Equal(t, "Test Business", reportReq.BusinessName)
	assert.Equal(t, "https://test.com", reportReq.WebsiteURL)
	assert.True(t, reportReq.IncludeAudit)
	assert.NotNil(t, reportReq.Result)
	assert.NotNil(t, reportReq.Comparison)
	assert.Equal(t, "test", reportReq.Metadata["source"])

	// Test UpdateConfigRequest
	configReq := UpdateConfigRequest{
		EnableDetailedExplanations: true,
		EnableRiskAnalysis:         true,
		EnableRecommendations:      true,
		EnableAuditTrail:           true,
		MinConfidenceThreshold:     0.6,
		MaxRiskProbability:         0.8,
		Language:                   "en",
	}

	assert.True(t, configReq.EnableDetailedExplanations)
	assert.True(t, configReq.EnableRiskAnalysis)
	assert.True(t, configReq.EnableRecommendations)
	assert.True(t, configReq.EnableAuditTrail)
	assert.Equal(t, 0.6, configReq.MinConfidenceThreshold)
	assert.Equal(t, 0.8, configReq.MaxRiskProbability)
	assert.Equal(t, "en", configReq.Language)
}

func TestVerificationReasoningHandler_ResponseStructs(t *testing.T) {
	// Test GenerateReasoningResponse
	reasoning := &external.VerificationReasoning{
		Status:          "PASSED",
		OverallScore:    0.85,
		ConfidenceLevel: "high",
		Explanation:     "Test explanation",
	}

	response := GenerateReasoningResponse{
		Success:   true,
		Reasoning: reasoning,
		Error:     "",
		Timestamp: time.Now(),
	}

	assert.True(t, response.Success)
	assert.Equal(t, reasoning, response.Reasoning)
	assert.Empty(t, response.Error)
	assert.NotZero(t, response.Timestamp)

	// Test GenerateReportResponse
	report := &external.VerificationReport{
		ReportID:       "report-123",
		VerificationID: "test-123",
		BusinessName:   "Test Business",
		Status:         "PASSED",
		OverallScore:   0.85,
	}

	reportResponse := GenerateReportResponse{
		Success:   true,
		Report:    report,
		Error:     "",
		Timestamp: time.Now(),
	}

	assert.True(t, reportResponse.Success)
	assert.Equal(t, report, reportResponse.Report)
	assert.Empty(t, reportResponse.Error)
	assert.NotZero(t, reportResponse.Timestamp)

	// Test UpdateConfigResponse
	config := &external.VerificationReasoningConfig{
		EnableDetailedExplanations: true,
		EnableRiskAnalysis:         true,
		EnableRecommendations:      true,
		EnableAuditTrail:           true,
		MinConfidenceThreshold:     0.6,
		MaxRiskProbability:         0.8,
		Language:                   "en",
	}

	configResponse := UpdateConfigResponse{
		Success:   true,
		Config:    config,
		Error:     "",
		Timestamp: time.Now(),
	}

	assert.True(t, configResponse.Success)
	assert.Equal(t, config, configResponse.Config)
	assert.Empty(t, configResponse.Error)
	assert.NotZero(t, configResponse.Timestamp)
}
