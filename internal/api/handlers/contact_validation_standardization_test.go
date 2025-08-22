package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func TestContactValidationHandler_ValidatePhone(t *testing.T) {
	logger := zap.NewNop()
	config := external.GetDefaultContactValidationConfig()
	standardizer := external.NewContactValidationStandardizer(config, logger)
	handler := NewContactValidationHandler(standardizer, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid phone number",
			requestBody: ValidationRequest{
				Value: "+1234567890",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid phone number",
			requestBody: ValidationRequest{
				Value: "invalid-phone",
			},
			expectedStatus: http.StatusOK, // Validation returns result even for invalid numbers
		},
		{
			name: "empty phone number",
			requestBody: ValidationRequest{
				Value: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Phone number is required",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/validate/phone", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ValidationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response.Error, tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response ValidationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Result)
			}
		})
	}
}

func TestContactValidationHandler_ValidateEmail(t *testing.T) {
	logger := zap.NewNop()
	standardizer := external.NewContactValidationStandardizer(logger)
	handler := NewContactValidationHandler(standardizer, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid email address",
			requestBody: ValidationRequest{
				Value: "test@example.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid email address",
			requestBody: ValidationRequest{
				Value: "invalid-email",
			},
			expectedStatus: http.StatusOK, // Validation returns result even for invalid emails
		},
		{
			name: "empty email address",
			requestBody: ValidationRequest{
				Value: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email address is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/validate/email", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ValidationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response.Error, tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response ValidationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Result)
			}
		})
	}
}

func TestContactValidationHandler_ValidateAddress(t *testing.T) {
	logger := zap.NewNop()
	standardizer := external.NewContactValidationStandardizer(logger)
	handler := NewContactValidationHandler(standardizer, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid address",
			requestBody: ValidationRequest{
				Value: "123 Main St, Anytown, ST 12345",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty address",
			requestBody: ValidationRequest{
				Value: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Address is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/validate/address", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ValidationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response.Error, tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response ValidationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Result)
			}
		})
	}
}

func TestContactValidationHandler_ValidateBatch(t *testing.T) {
	logger := zap.NewNop()
	standardizer := external.NewContactValidationStandardizer(logger)
	handler := NewContactValidationHandler(standardizer, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid batch phone validation",
			requestBody: BatchValidationRequest{
				ContactType: "phone",
				Values:      []string{"+1234567890", "+0987654321"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid batch email validation",
			requestBody: BatchValidationRequest{
				ContactType: "email",
				Values:      []string{"test1@example.com", "test2@example.com"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid batch address validation",
			requestBody: BatchValidationRequest{
				ContactType: "address",
				Values:      []string{"123 Main St", "456 Oak Ave"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid contact type",
			requestBody: BatchValidationRequest{
				ContactType: "invalid",
				Values:      []string{"test"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid contact type",
		},
		{
			name: "empty contact type",
			requestBody: BatchValidationRequest{
				ContactType: "",
				Values:      []string{"test"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Contact type is required",
		},
		{
			name: "empty values",
			requestBody: BatchValidationRequest{
				ContactType: "phone",
				Values:      []string{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "At least one value is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/validate/batch", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response BatchValidationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response.Error, tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response BatchValidationResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Result)
			}
		})
	}
}

func TestContactValidationHandler_GetConfig(t *testing.T) {
	logger := zap.NewNop()
	standardizer := external.NewContactValidationStandardizer(logger)
	handler := NewContactValidationHandler(standardizer, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/config", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Config)
	assert.Equal(t, "Configuration retrieved successfully", response.Message)
}

func TestContactValidationHandler_UpdateConfig(t *testing.T) {
	logger := zap.NewNop()
	standardizer := external.NewContactValidationStandardizer(logger)
	handler := NewContactValidationHandler(standardizer, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid config update",
			requestBody: ConfigRequest{
				Config: &external.ContactValidationConfig{
					MinValidationConfidence: 0.8,
					MaxBatchSize:            100,
					ValidationTimeout:       30 * time.Second,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid confidence threshold",
			requestBody: ConfigRequest{
				Config: &external.ContactValidationConfig{
					MinValidationConfidence: 1.5, // Invalid: > 1.0
					MaxBatchSize:            100,
					ValidationTimeout:       30 * time.Second,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "min_validation_confidence must be between 0.0 and 1.0",
		},
		{
			name: "invalid batch size",
			requestBody: ConfigRequest{
				Config: &external.ContactValidationConfig{
					MinValidationConfidence: 0.8,
					MaxBatchSize:            0, // Invalid: <= 0
					ValidationTimeout:       30 * time.Second,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "max_batch_size must be greater than 0",
		},
		{
			name: "nil config",
			requestBody: ConfigRequest{
				Config: nil,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Configuration is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("PUT", "/config", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ConfigResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response.Error, tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response ConfigResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Config)
				assert.Equal(t, "Configuration updated successfully", response.Message)
			}
		})
	}
}

func TestContactValidationHandler_GetStatistics(t *testing.T) {
	logger := zap.NewNop()
	standardizer := external.NewContactValidationStandardizer(logger)
	handler := NewContactValidationHandler(standardizer, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response StatisticsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Stats)
	assert.Equal(t, "Statistics retrieved successfully", response.Message)

	// Check that key statistics are present
	stats := response.Stats
	assert.Contains(t, stats, "phone_validation_enabled")
	assert.Contains(t, stats, "email_validation_enabled")
	assert.Contains(t, stats, "address_validation_enabled")
	assert.Contains(t, stats, "min_confidence_threshold")
	assert.Contains(t, stats, "max_batch_size")
	assert.Contains(t, stats, "timestamp")
}

func TestContactValidationHandler_HealthCheck(t *testing.T) {
	logger := zap.NewNop()
	standardizer := external.NewContactValidationStandardizer(logger)
	handler := NewContactValidationHandler(standardizer, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Equal(t, "contact_validation", response["service"])
}

func TestContactValidationHandler_RouteRegistration(t *testing.T) {
	logger := zap.NewNop()
	standardizer := external.NewContactValidationStandardizer(logger)
	handler := NewContactValidationHandler(standardizer, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test that all routes are registered correctly
	testRoutes := []struct {
		method string
		path   string
	}{
		{"POST", "/validate/phone"},
		{"POST", "/validate/email"},
		{"POST", "/validate/address"},
		{"POST", "/validate/batch"},
		{"GET", "/config"},
		{"PUT", "/config"},
		{"GET", "/stats"},
		{"GET", "/health"},
	}

	for _, route := range testRoutes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should not get 404 (route not found)
			assert.NotEqual(t, http.StatusNotFound, w.Code)
		})
	}
}

func TestContactValidationHandler_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	standardizer := external.NewContactValidationStandardizer(logger)
	handler := NewContactValidationHandler(standardizer, logger)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "invalid JSON in phone validation",
			method:         "POST",
			path:           "/validate/phone",
			body:           "{invalid json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON in email validation",
			method:         "POST",
			path:           "/validate/email",
			body:           "{invalid json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON in address validation",
			method:         "POST",
			path:           "/validate/address",
			body:           "{invalid json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON in batch validation",
			method:         "POST",
			path:           "/validate/batch",
			body:           "{invalid json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON in config update",
			method:         "PUT",
			path:           "/config",
			body:           "{invalid json}",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
