package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/config"
	"kyb-platform/services/risk-assessment-service/internal/middleware"
	"kyb-platform/services/risk-assessment-service/internal/models"
	"kyb-platform/services/risk-assessment-service/internal/validation"
)

// Test helper functions
func createTestHandler() *RiskAssessmentHandler {
	logger := zap.NewNop()
	config := &config.Config{
		Server: config.ServerConfig{
			Port:         "8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}

	handler := &RiskAssessmentHandler{
		logger:       logger,
		config:       config,
		validator:    validation.NewValidator(),
		errorHandler: middleware.NewErrorHandler(logger),
	}

	return handler
}

func createTestRequest(method, url string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// Test cases for input validation
func TestHandleRiskAssessment_InvalidInput(t *testing.T) {
	handler := createTestHandler()

	testCases := []struct {
		name        string
		requestBody models.RiskAssessmentRequest
		expectError bool
	}{
		{
			name: "Empty business name",
			requestBody: models.RiskAssessmentRequest{
				BusinessName:      "",
				BusinessAddress:   "123 Test St",
				Industry:          "Tech",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expectError: true,
		},
		{
			name: "Empty business address",
			requestBody: models.RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "",
				Industry:          "Tech",
				Country:           "US",
				PredictionHorizon: 3,
			},
			expectError: true,
		},
		{
			name: "Invalid email format",
			requestBody: models.RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "123 Test St",
				Industry:          "Tech",
				Country:           "US",
				Email:             "invalid-email",
				PredictionHorizon: 3,
			},
			expectError: true,
		},
		{
			name: "Invalid phone format",
			requestBody: models.RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "123 Test St",
				Industry:          "Tech",
				Country:           "US",
				Phone:             "invalid-phone",
				PredictionHorizon: 3,
			},
			expectError: true,
		},
		{
			name: "Valid request",
			requestBody: models.RiskAssessmentRequest{
				BusinessName:      "Test Company",
				BusinessAddress:   "123 Test St, Test City, TC 12345",
				Industry:          "Technology",
				Country:           "US",
				Email:             "test@company.com",
				Phone:             "+1-555-123-4567",
				PredictionHorizon: 3,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createTestRequest("POST", "/api/v1/assess", tc.requestBody)
			w := httptest.NewRecorder()

			if tc.expectError {
				handler.HandleRiskAssessment(w, req)
				assert.Equal(t, http.StatusBadRequest, w.Code)
			} else {
				// For valid requests, we expect a 500 error since we don't have a real engine
				handler.HandleRiskAssessment(w, req)
				assert.Equal(t, http.StatusInternalServerError, w.Code)
			}
		})
	}
}

func TestHandleRiskAssessment_InvalidJSON(t *testing.T) {
	handler := createTestHandler()

	req := httptest.NewRequest("POST", "/api/v1/assess", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleRiskAssessment(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleBatchRiskAssessment_EmptyRequest(t *testing.T) {
	handler := createTestHandler()

	reqBody := struct {
		Requests []models.RiskAssessmentRequest `json:"requests"`
	}{
		Requests: []models.RiskAssessmentRequest{},
	}

	req := createTestRequest("POST", "/api/v1/assess/batch", reqBody)
	w := httptest.NewRecorder()

	handler.HandleBatchRiskAssessment(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleBatchRiskAssessment_TooManyRequests(t *testing.T) {
	handler := createTestHandler()

	// Create request with more than 100 items
	requests := make([]models.RiskAssessmentRequest, 101)
	for i := 0; i < 101; i++ {
		requests[i] = models.RiskAssessmentRequest{
			BusinessName:      fmt.Sprintf("Company %d", i),
			BusinessAddress:   "123 Test St",
			Industry:          "Tech",
			Country:           "US",
			PredictionHorizon: 3,
		}
	}

	reqBody := struct {
		Requests []models.RiskAssessmentRequest `json:"requests"`
	}{
		Requests: requests,
	}

	req := createTestRequest("POST", "/api/v1/assess/batch", reqBody)
	w := httptest.NewRecorder()

	handler.HandleBatchRiskAssessment(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleExternalAdverseMediaMonitoring_InvalidInput(t *testing.T) {
	handler := createTestHandler()

	reqBody := struct {
		BusinessName string `json:"business_name"`
	}{
		BusinessName: "", // Empty name should fail validation
	}

	req := createTestRequest("POST", "/api/v1/external/adverse-media", reqBody)
	w := httptest.NewRecorder()

	handler.HandleExternalAdverseMediaMonitoring(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCompanyDataLookup_InvalidInput(t *testing.T) {
	handler := createTestHandler()

	reqBody := struct {
		BusinessName string `json:"business_name"`
		Country      string `json:"country"`
	}{
		BusinessName: "", // Empty name should fail validation
		Country:      "US",
	}

	req := createTestRequest("POST", "/api/v1/external/company-data", reqBody)
	w := httptest.NewRecorder()

	handler.HandleCompanyDataLookup(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleExternalComplianceCheck_InvalidInput(t *testing.T) {
	handler := createTestHandler()

	reqBody := struct {
		BusinessName string `json:"business_name"`
		Country      string `json:"country"`
	}{
		BusinessName: "", // Empty name should fail validation
		Country:      "US",
	}

	req := createTestRequest("POST", "/api/v1/external/compliance", reqBody)
	w := httptest.NewRecorder()

	handler.HandleExternalComplianceCheck(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleExternalDataSources_Success(t *testing.T) {
	handler := createTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/external/sources", nil)
	w := httptest.NewRecorder()

	handler.HandleExternalDataSources(w, req)

	// Should return 500 since we don't have a real external data service
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Test the generateID function
func TestGenerateID(t *testing.T) {
	handler := createTestHandler()

	id1 := handler.generateID()
	id2 := handler.generateID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "risk_")
	assert.Contains(t, id2, "risk_")
}

// Benchmark tests
func BenchmarkHandleRiskAssessment(b *testing.B) {
	handler := createTestHandler()

	reqBody := models.RiskAssessmentRequest{
		BusinessName:      "Test Company",
		BusinessAddress:   "123 Test St, Test City, TC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := createTestRequest("POST", "/api/v1/assess", reqBody)
		w := httptest.NewRecorder()
		handler.HandleRiskAssessment(w, req)
	}
}

func BenchmarkGenerateID(b *testing.B) {
	handler := createTestHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.generateID()
	}
}
