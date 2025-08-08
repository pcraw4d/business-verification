package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/risk"
)

func TestRiskHandler_AssessRiskHandler(t *testing.T) {
	// Create mock logger
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "text",
	})

	// Create mock risk service components
	categoryRegistry := risk.CreateDefaultRiskCategories()
	thresholdManager := risk.CreateDefaultThresholds()
	industryModelRegistry := risk.CreateDefaultIndustryModels()
	calculator := risk.NewRiskFactorCalculator(categoryRegistry)
	scoringAlgorithm := risk.NewWeightedScoringAlgorithm()
	predictionAlgorithm := risk.NewRiskPredictionAlgorithm()

	// Create risk history service
	riskHistoryService := risk.NewRiskHistoryService(logger, nil)

	// Create alert service
	alertService := risk.NewAlertService(logger, thresholdManager)

	// Create report service
	reportService := risk.NewReportService(logger, riskHistoryService, alertService)

	// Create risk service
	riskService := risk.NewRiskService(
		logger,
		calculator,
		scoringAlgorithm,
		predictionAlgorithm,
		thresholdManager,
		categoryRegistry,
		industryModelRegistry,
		riskHistoryService,
		alertService,
		reportService,
	)

	// Create risk handler
	handler := NewRiskHandler(logger, riskService, riskHistoryService)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Valid risk assessment request",
			requestBody: `{
				"business_id": "test-123",
				"business_name": "Test Company",
				"categories": ["financial", "operational"],
				"include_predictions": true
			}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response risk.RiskAssessmentResponse
				if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if response.Assessment == nil {
					t.Error("Expected assessment in response")
					return
				}

				if response.Assessment.BusinessID != "test-123" {
					t.Errorf("Expected business ID 'test-123', got '%s'", response.Assessment.BusinessID)
				}

				if response.Assessment.BusinessName != "Test Company" {
					t.Errorf("Expected business name 'Test Company', got '%s'", response.Assessment.BusinessName)
				}

				if response.Assessment.OverallScore < 0 || response.Assessment.OverallScore > 100 {
					t.Errorf("Expected overall score between 0 and 100, got %f", response.Assessment.OverallScore)
				}

				if len(response.Assessment.CategoryScores) == 0 {
					t.Error("Expected category scores in response")
				}

				if len(response.Assessment.FactorScores) == 0 {
					t.Error("Expected factor scores in response")
				}

				if len(response.Predictions) == 0 {
					t.Error("Expected predictions in response when include_predictions is true")
				}
			},
		},
		{
			name: "Missing business ID",
			requestBody: `{
				"business_name": "Test Company",
				"categories": ["financial"]
			}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if !bytes.Contains(rr.Body.Bytes(), []byte("business ID is required")) {
					t.Error("Expected error message about missing business ID")
				}
			},
		},
		{
			name: "Missing business name",
			requestBody: `{
				"business_id": "test-123",
				"categories": ["financial"]
			}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if !bytes.Contains(rr.Body.Bytes(), []byte("business name is required")) {
					t.Error("Expected error message about missing business name")
				}
			},
		},
		{
			name: "Invalid JSON",
			requestBody: `{
				"business_id": "test-123",
				"business_name": "Test Company",
				"categories": ["financial"
			}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if !bytes.Contains(rr.Body.Bytes(), []byte("Invalid request body")) {
					t.Error("Expected error message about invalid request body")
				}
			},
		},
		{
			name: "Empty categories",
			requestBody: `{
				"business_id": "test-123",
				"business_name": "Test Company",
				"categories": [],
				"include_predictions": false
			}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response risk.RiskAssessmentResponse
				if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if response.Assessment == nil {
					t.Error("Expected assessment in response")
					return
				}

				// Should still get an assessment even with empty categories
				if response.Assessment.OverallScore < 0 || response.Assessment.OverallScore > 100 {
					t.Errorf("Expected overall score between 0 and 100, got %f", response.Assessment.OverallScore)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("POST", "/v1/risk/assess", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}

			// Add request ID to context
			ctx := req.Context()
			ctx = context.WithValue(ctx, "request_id", "test-request-id")
			req = req.WithContext(ctx)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.AssessRiskHandler(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check response
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestRiskHandler_GetRiskCategoriesHandler(t *testing.T) {
	// Create mock logger
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "text",
	})

	// Create mock risk service components
	categoryRegistry := risk.CreateDefaultRiskCategories()
	thresholdManager := risk.CreateDefaultThresholds()
	industryModelRegistry := risk.CreateDefaultIndustryModels()
	calculator := risk.NewRiskFactorCalculator(categoryRegistry)
	scoringAlgorithm := risk.NewWeightedScoringAlgorithm()
	predictionAlgorithm := risk.NewRiskPredictionAlgorithm()

	// Create risk history service
	riskHistoryService := risk.NewRiskHistoryService(logger, nil)

	// Create alert service
	alertService := risk.NewAlertService(logger, thresholdManager)

	// Create report service
	reportService := risk.NewReportService(logger, riskHistoryService, alertService)

	// Create risk service
	riskService := risk.NewRiskService(
		logger,
		calculator,
		scoringAlgorithm,
		predictionAlgorithm,
		thresholdManager,
		categoryRegistry,
		industryModelRegistry,
		riskHistoryService,
		alertService,
		reportService,
	)

	// Create risk handler
	handler := NewRiskHandler(logger, riskService, riskHistoryService)

	// Create request
	req, err := http.NewRequest("GET", "/v1/risk/categories", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add request ID to context
	ctx := req.Context()
	ctx = context.WithValue(ctx, "request_id", "test-request-id")
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetRiskCategoriesHandler(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response
	var response map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
		return
	}

	if response["categories"] == nil {
		t.Error("Expected categories in response")
	}

	if response["total"] == nil {
		t.Error("Expected total count in response")
	}
}

func TestRiskHandler_GetRiskFactorsHandler(t *testing.T) {
	// Create mock logger
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "text",
	})

	// Create mock risk service components
	categoryRegistry := risk.CreateDefaultRiskCategories()
	thresholdManager := risk.CreateDefaultThresholds()
	industryModelRegistry := risk.CreateDefaultIndustryModels()
	calculator := risk.NewRiskFactorCalculator(categoryRegistry)
	scoringAlgorithm := risk.NewWeightedScoringAlgorithm()
	predictionAlgorithm := risk.NewRiskPredictionAlgorithm()

	// Create risk history service
	riskHistoryService := risk.NewRiskHistoryService(logger, nil)

	// Create alert service
	alertService := risk.NewAlertService(logger, thresholdManager)

	// Create report service
	reportService := risk.NewReportService(logger, riskHistoryService, alertService)

	// Create risk service
	riskService := risk.NewRiskService(
		logger,
		calculator,
		scoringAlgorithm,
		predictionAlgorithm,
		thresholdManager,
		categoryRegistry,
		industryModelRegistry,
		riskHistoryService,
		alertService,
		reportService,
	)

	// Create risk handler
	handler := NewRiskHandler(logger, riskService, riskHistoryService)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Get all factors",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if response["factors"] == nil {
					t.Error("Expected factors in response")
				}

				if response["total"] == nil {
					t.Error("Expected total count in response")
				}
			},
		},
		{
			name:           "Get factors by category",
			queryParams:    "?category=financial",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if response["factors"] == nil {
					t.Error("Expected factors in response")
				}

				if response["category"] != "financial" {
					t.Errorf("Expected category 'financial', got '%v'", response["category"])
				}
			},
		},
		{
			name:           "Invalid category",
			queryParams:    "?category=invalid",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if !bytes.Contains(rr.Body.Bytes(), []byte("Invalid risk category")) {
					t.Error("Expected error message about invalid risk category")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("GET", "/v1/risk/factors"+tt.queryParams, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Add request ID to context
			ctx := req.Context()
			ctx = context.WithValue(ctx, "request_id", "test-request-id")
			req = req.WithContext(ctx)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.GetRiskFactorsHandler(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check response
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestRiskHandler_GetRiskThresholdsHandler(t *testing.T) {
	// Create mock logger
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "text",
	})

	// Create mock risk service components
	categoryRegistry := risk.CreateDefaultRiskCategories()
	thresholdManager := risk.CreateDefaultThresholds()
	industryModelRegistry := risk.CreateDefaultIndustryModels()
	calculator := risk.NewRiskFactorCalculator(categoryRegistry)
	scoringAlgorithm := risk.NewWeightedScoringAlgorithm()
	predictionAlgorithm := risk.NewRiskPredictionAlgorithm()

	// Create risk history service
	riskHistoryService := risk.NewRiskHistoryService(logger, nil)

	// Create alert service
	alertService := risk.NewAlertService(logger, thresholdManager)

	// Create report service
	reportService := risk.NewReportService(logger, riskHistoryService, alertService)

	// Create risk service
	riskService := risk.NewRiskService(
		logger,
		calculator,
		scoringAlgorithm,
		predictionAlgorithm,
		thresholdManager,
		categoryRegistry,
		industryModelRegistry,
		riskHistoryService,
		alertService,
		reportService,
	)

	// Create risk handler
	handler := NewRiskHandler(logger, riskService, riskHistoryService)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Get all thresholds",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if response["thresholds"] == nil {
					t.Error("Expected thresholds in response")
				}

				if response["total"] == nil {
					t.Error("Expected total count in response")
				}
			},
		},
		{
			name:           "Get thresholds by category",
			queryParams:    "?category=financial",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				if err := json.NewDecoder(bytes.NewReader(rr.Body.Bytes())).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if response["thresholds"] == nil {
					t.Error("Expected thresholds in response")
				}

				if response["category"] != "financial" {
					t.Errorf("Expected category 'financial', got '%v'", response["category"])
				}
			},
		},
		{
			name:           "Invalid category",
			queryParams:    "?category=invalid",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if !bytes.Contains(rr.Body.Bytes(), []byte("Invalid risk category")) {
					t.Error("Expected error message about invalid risk category")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("GET", "/v1/risk/thresholds"+tt.queryParams, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Add request ID to context
			ctx := req.Context()
			ctx = context.WithValue(ctx, "request_id", "test-request-id")
			req = req.WithContext(ctx)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.GetRiskThresholdsHandler(rr, req)

			// Check status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check response
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}
