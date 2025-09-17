package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/integrations"
)

// MockFreeDataValidationService is a mock implementation of the validation service
type MockFreeDataValidationService struct {
	validationResults map[string]*integrations.ValidationResult
	stats             map[string]interface{}
}

// Ensure MockFreeDataValidationService implements the interface
var _ integrations.FreeDataValidationServiceInterface = (*MockFreeDataValidationService)(nil)

func NewMockFreeDataValidationService() *MockFreeDataValidationService {
	return &MockFreeDataValidationService{
		validationResults: make(map[string]*integrations.ValidationResult),
		stats: map[string]interface{}{
			"total_validations":     10,
			"valid_count":           8,
			"invalid_count":         2,
			"average_quality_score": 0.85,
			"cache_size":            10,
			"cost_per_validation":   0.0,
		},
	}
}

func (m *MockFreeDataValidationService) ValidateBusinessData(ctx context.Context, data integrations.BusinessDataForValidation) (*integrations.ValidationResult, error) {
	// Return a mock result based on business ID
	if result, exists := m.validationResults[data.BusinessID]; exists {
		return result, nil
	}

	// Default mock result
	return &integrations.ValidationResult{
		BusinessID:            data.BusinessID,
		IsValid:               true,
		QualityScore:          0.85,
		ConsistencyScore:      0.9,
		CompletenessScore:     0.8,
		AccuracyScore:         0.85,
		FreshnessScore:        1.0,
		CrossReferenceResults: make(map[string]interface{}),
		ValidationErrors:      []integrations.ValidationError{},
		ValidationWarnings:    []integrations.ValidationWarning{},
		DataSources:           []integrations.DataSourceInfo{},
		ValidatedAt:           time.Now(),
		ValidationTime:        100 * time.Millisecond,
		Cost:                  0.0,
	}, nil
}

func (m *MockFreeDataValidationService) GetValidationStats() map[string]interface{} {
	return m.stats
}

func TestFreeDataValidationHandler_ValidateBusinessData(t *testing.T) {
	mockService := NewMockFreeDataValidationService()
	logger := zap.NewNop()
	handler := NewFreeDataValidationHandler(mockService, logger)

	tests := []struct {
		name           string
		request        ValidateBusinessDataRequest
		expectedStatus int
		expectedValid  bool
		expectedScore  float64
	}{
		{
			name: "valid_business_data",
			request: ValidateBusinessDataRequest{
				BusinessID:         "test-001",
				Name:               "Test Company",
				Description:        "A test company",
				Address:            "123 Test St, Test City, TC 12345",
				Phone:              "+1-555-123-4567",
				Email:              "contact@test.com",
				Website:            "https://www.test.com",
				Industry:           "Technology",
				Country:            "US",
				RegistrationNumber: "1234567890",
				TaxID:              "12-3456789",
			},
			expectedStatus: http.StatusOK,
			expectedValid:  true,
			expectedScore:  0.85,
		},
		{
			name: "missing_business_id",
			request: ValidateBusinessDataRequest{
				Name: "Test Company",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid_business_data",
			request: ValidateBusinessDataRequest{
				BusinessID: "invalid-test",
			},
			expectedStatus: http.StatusOK,
			expectedValid:  true,
			expectedScore:  0.85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			requestBody, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/api/v3/validate/business-data", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.ValidateBusinessData(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Parse response if successful
			if tt.expectedStatus == http.StatusOK {
				var response ValidateBusinessDataResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Success != true {
					t.Errorf("Expected success=true, got %v", response.Success)
				}

				if response.IsValid != tt.expectedValid {
					t.Errorf("Expected IsValid=%v, got %v", tt.expectedValid, response.IsValid)
				}

				if response.QualityScore != tt.expectedScore {
					t.Errorf("Expected QualityScore=%v, got %v", tt.expectedScore, response.QualityScore)
				}

				if response.Cost != 0.0 {
					t.Errorf("Expected Cost=0.0, got %v", response.Cost)
				}
			}
		})
	}
}

func TestFreeDataValidationHandler_GetValidationStats(t *testing.T) {
	mockService := NewMockFreeDataValidationService()
	logger := zap.NewNop()
	handler := NewFreeDataValidationHandler(mockService, logger)

	// Create HTTP request
	req := httptest.NewRequest(http.MethodGet, "/api/v3/validate/stats", nil)
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetValidationStats(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Parse response
	var response ValidationStatsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check response fields
	if response.Success != true {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.TotalValidations != 10 {
		t.Errorf("Expected TotalValidations=10, got %v", response.TotalValidations)
	}

	if response.ValidCount != 8 {
		t.Errorf("Expected ValidCount=8, got %v", response.ValidCount)
	}

	if response.InvalidCount != 2 {
		t.Errorf("Expected InvalidCount=2, got %v", response.InvalidCount)
	}

	if response.AverageQualityScore != 0.85 {
		t.Errorf("Expected AverageQualityScore=0.85, got %v", response.AverageQualityScore)
	}

	if response.CostPerValidation != 0.0 {
		t.Errorf("Expected CostPerValidation=0.0, got %v", response.CostPerValidation)
	}
}

func TestFreeDataValidationHandler_HealthCheck(t *testing.T) {
	mockService := NewMockFreeDataValidationService()
	logger := zap.NewNop()
	handler := NewFreeDataValidationHandler(mockService, logger)

	// Create HTTP request
	req := httptest.NewRequest(http.MethodGet, "/api/v3/validate/health", nil)
	rr := httptest.NewRecorder()

	// Call handler
	handler.HealthCheck(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check response fields
	if response["status"] != "healthy" {
		t.Errorf("Expected status=healthy, got %v", response["status"])
	}

	if response["service"] != "free_data_validation" {
		t.Errorf("Expected service=free_data_validation, got %v", response["service"])
	}
}

func TestFreeDataValidationHandler_GetValidationConfig(t *testing.T) {
	mockService := NewMockFreeDataValidationService()
	logger := zap.NewNop()
	handler := NewFreeDataValidationHandler(mockService, logger)

	// Create HTTP request
	req := httptest.NewRequest(http.MethodGet, "/api/v3/validate/config", nil)
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetValidationConfig(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check response fields
	if response["success"] != true {
		t.Errorf("Expected success=true, got %v", response["success"])
	}

	config, ok := response["config"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected config to be a map")
	}

	// Check some config values
	if config["min_quality_score"] != 0.7 {
		t.Errorf("Expected min_quality_score=0.7, got %v", config["min_quality_score"])
	}

	if config["enable_cross_reference"] != true {
		t.Errorf("Expected enable_cross_reference=true, got %v", config["enable_cross_reference"])
	}

	if config["max_validation_time"] != float64(30) {
		t.Errorf("Expected max_validation_time=30, got %v", config["max_validation_time"])
	}
}

func TestFreeDataValidationHandler_BatchValidateBusinessData(t *testing.T) {
	mockService := NewMockFreeDataValidationService()
	logger := zap.NewNop()
	handler := NewFreeDataValidationHandler(mockService, logger)

	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "valid_batch_request",
			request: map[string]interface{}{
				"businesses": []ValidateBusinessDataRequest{
					{
						BusinessID: "batch-001",
						Name:       "Batch Company 1",
						Country:    "US",
					},
					{
						BusinessID: "batch-002",
						Name:       "Batch Company 2",
						Country:    "UK",
					},
				},
				"max_concurrent": 2,
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "empty_batch_request",
			request: map[string]interface{}{
				"businesses": []ValidateBusinessDataRequest{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "batch_with_default_concurrency",
			request: map[string]interface{}{
				"businesses": []ValidateBusinessDataRequest{
					{
						BusinessID: "batch-003",
						Name:       "Batch Company 3",
						Country:    "US",
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			requestBody, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/api/v3/validate/business-data/batch", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.BatchValidateBusinessData(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Parse response if successful
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response["success"] != true {
					t.Errorf("Expected success=true, got %v", response["success"])
				}

				totalValidations, ok := response["total_validations"].(float64)
				if !ok {
					t.Fatal("Expected total_validations to be a number")
				}

				if int(totalValidations) != tt.expectedCount {
					t.Errorf("Expected total_validations=%d, got %v", tt.expectedCount, totalValidations)
				}

				if response["total_cost"] != float64(0.0) {
					t.Errorf("Expected total_cost=0.0, got %v", response["total_cost"])
				}
			}
		})
	}
}

func TestFreeDataValidationHandler_MethodNotAllowed(t *testing.T) {
	mockService := NewMockFreeDataValidationService()
	logger := zap.NewNop()
	handler := NewFreeDataValidationHandler(mockService, logger)

	tests := []struct {
		name     string
		method   string
		endpoint string
		handler  func(http.ResponseWriter, *http.Request)
	}{
		{
			name:     "validate_business_data_get",
			method:   http.MethodGet,
			endpoint: "/api/v3/validate/business-data",
			handler:  handler.ValidateBusinessData,
		},
		{
			name:     "get_validation_stats_post",
			method:   http.MethodPost,
			endpoint: "/api/v3/validate/stats",
			handler:  handler.GetValidationStats,
		},
		{
			name:     "health_check_post",
			method:   http.MethodPost,
			endpoint: "/api/v3/validate/health",
			handler:  handler.HealthCheck,
		},
		{
			name:     "get_validation_config_post",
			method:   http.MethodPost,
			endpoint: "/api/v3/validate/config",
			handler:  handler.GetValidationConfig,
		},
		{
			name:     "batch_validate_business_data_get",
			method:   http.MethodGet,
			endpoint: "/api/v3/validate/business-data/batch",
			handler:  handler.BatchValidateBusinessData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			rr := httptest.NewRecorder()

			tt.handler(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
			}
		})
	}
}

func TestFreeDataValidationHandler_RegisterRoutes(t *testing.T) {
	mockService := NewMockFreeDataValidationService()
	logger := zap.NewNop()
	handler := NewFreeDataValidationHandler(mockService, logger)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register routes
	handler.RegisterRoutes(mux)

	// Test that routes are registered by making requests
	testRoutes := []struct {
		method   string
		path     string
		expected int
	}{
		{http.MethodPost, "/api/v3/validate/business-data", http.StatusBadRequest}, // Missing body
		{http.MethodGet, "/api/v3/validate/stats", http.StatusOK},
		{http.MethodGet, "/api/v3/validate/config", http.StatusOK},
		{http.MethodGet, "/api/v3/validate/health", http.StatusOK},
		{http.MethodPost, "/api/v3/validate/business-data/batch", http.StatusBadRequest}, // Missing body
	}

	for _, route := range testRoutes {
		t.Run(route.method+"_"+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if rr.Code != route.expected {
				t.Errorf("Expected status %d for %s %s, got %d", route.expected, route.method, route.path, rr.Code)
			}
		})
	}
}

func TestFreeDataValidationHandler_InvalidJSON(t *testing.T) {
	mockService := NewMockFreeDataValidationService()
	logger := zap.NewNop()
	handler := NewFreeDataValidationHandler(mockService, logger)

	// Test with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v3/validate/business-data", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ValidateBusinessData(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestFreeDataValidationHandler_HelperFunctions(t *testing.T) {
	// Test conversion functions
	t.Run("convertValidationErrors", func(t *testing.T) {
		serviceErrors := []integrations.ValidationError{
			{
				Field:    "name",
				Message:  "Name is required",
				Severity: "high",
				Source:   "validation",
				Code:     "REQUIRED",
			},
		}

		handlerErrors := convertValidationErrors(serviceErrors)

		if len(handlerErrors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(handlerErrors))
		}

		if handlerErrors[0].Field != "name" {
			t.Errorf("Expected field=name, got %s", handlerErrors[0].Field)
		}
	})

	t.Run("convertValidationWarnings", func(t *testing.T) {
		serviceWarnings := []integrations.ValidationWarning{
			{
				Field:    "phone",
				Message:  "Phone format may be invalid",
				Severity: "medium",
				Source:   "validation",
				Code:     "FORMAT_WARNING",
			},
		}

		handlerWarnings := convertValidationWarnings(serviceWarnings)

		if len(handlerWarnings) != 1 {
			t.Errorf("Expected 1 warning, got %d", len(handlerWarnings))
		}

		if handlerWarnings[0].Field != "phone" {
			t.Errorf("Expected field=phone, got %s", handlerWarnings[0].Field)
		}
	})

	t.Run("convertDataSources", func(t *testing.T) {
		serviceSources := []integrations.DataSourceInfo{
			{
				Name:        "SEC EDGAR",
				Type:        "government_registry",
				IsFree:      true,
				Cost:        0.0,
				LastUpdated: time.Now(),
				Reliability: 0.95,
			},
		}

		handlerSources := convertDataSources(serviceSources)

		if len(handlerSources) != 1 {
			t.Errorf("Expected 1 source, got %d", len(handlerSources))
		}

		if handlerSources[0].Name != "SEC EDGAR" {
			t.Errorf("Expected name=SEC EDGAR, got %s", handlerSources[0].Name)
		}

		if !handlerSources[0].IsFree {
			t.Error("Expected IsFree=true")
		}
	})
}

// Benchmark tests
func BenchmarkFreeDataValidationHandler_ValidateBusinessData(b *testing.B) {
	mockService := NewMockFreeDataValidationService()
	logger := zap.NewNop()
	handler := NewFreeDataValidationHandler(mockService, logger)

	request := ValidateBusinessDataRequest{
		BusinessID:         "benchmark-test",
		Name:               "Benchmark Company",
		Description:        "A company for benchmarking",
		Address:            "123 Benchmark St, Benchmark City, BC 12345",
		Phone:              "+1-555-123-4567",
		Email:              "contact@benchmark.com",
		Website:            "https://www.benchmark.com",
		Industry:           "Technology",
		Country:            "US",
		RegistrationNumber: "1234567890",
		TaxID:              "12-3456789",
	}

	requestBody, _ := json.Marshal(request)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v3/validate/business-data", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ValidateBusinessData(rr, req)
	}
}
