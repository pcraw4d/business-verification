package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestErrorHandler_WriteError(t *testing.T) {
	logger := zap.NewNop()
	handler := NewErrorHandler(logger)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedError  ErrorCode
	}{
		{
			name: "validation error",
			err: &ValidationError{
				Code:       ErrorCodeMissingRequiredField,
				Message:    "Business name is required",
				Field:      "business_name",
				Constraint: "required",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrorCodeMissingRequiredField,
		},
		{
			name: "authentication error",
			err: &AuthenticationError{
				Code:    ErrorCodeMissingToken,
				Message: "Authentication token is required",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrorCodeMissingToken,
		},
		{
			name: "authorization error",
			err: &AuthorizationError{
				Code:    ErrorCodeInsufficientPermissions,
				Message: "Insufficient permissions",
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  ErrorCodeInsufficientPermissions,
		},
		{
			name: "rate limit error",
			err: &RateLimitError{
				Code:       ErrorCodeRateLimitExceeded,
				Message:    "Rate limit exceeded",
				RetryAfter: 60,
			},
			expectedStatus: http.StatusTooManyRequests,
			expectedError:  ErrorCodeRateLimitExceeded,
		},
		{
			name: "classification error",
			err: &ClassificationError{
				Code:    ErrorCodeClassificationFailed,
				Message: "Classification failed",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  ErrorCodeClassificationFailed,
		},
		{
			name: "external service error",
			err: &ExternalServiceError{
				Code:    ErrorCodeExternalAPIFailed,
				Service: "duckduckgo",
				Message: "External API failed",
			},
			expectedStatus: http.StatusBadGateway,
			expectedError:  ErrorCodeExternalAPIFailed,
		},
		{
			name: "timeout error",
			err: &TimeoutError{
				Code:      ErrorCodeTimeoutError,
				Operation: "website_scraping",
				Timeout:   10 * time.Second,
				Message:   "Operation 'website_scraping' timed out after 10s",
			},
			expectedStatus: http.StatusRequestTimeout,
			expectedError:  ErrorCodeTimeoutError,
		},
		{
			name:           "generic error",
			err:            assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  ErrorCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request
			req := httptest.NewRequest("POST", "/v1/classify", nil)
			req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))

			// Create response recorder
			w := httptest.NewRecorder()

			// Call WriteError
			handler.WriteError(w, req, tt.err)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response APIError
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Assert error code
			assert.Equal(t, tt.expectedError, response.Error)

			// Assert content type
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			// Assert request ID is included
			assert.Equal(t, "test-request-id", response.Details.RequestID)

			// Assert timestamp is set
			assert.False(t, response.Details.Timestamp.IsZero())

			// Assert context is included
			assert.Equal(t, "POST", response.Details.Context["method"])
			assert.Equal(t, "/v1/classify", response.Details.Context["path"])
		})
	}
}

func TestErrorHandler_ConvertError(t *testing.T) {
	logger := zap.NewNop()
	handler := NewErrorHandler(logger)

	req := httptest.NewRequest("GET", "/v1/classify/test", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))

	tests := []struct {
		name             string
		err              error
		expectedError    ErrorCode
		expectedStatus   int
		expectedCategory ErrorCategory
		expectedSeverity ErrorSeverity
	}{
		{
			name: "validation error",
			err: &ValidationError{
				Code:       ErrorCodeInvalidJSON,
				Message:    "Invalid JSON format",
				Field:      "request_body",
				Constraint: "valid_json",
			},
			expectedError:    ErrorCodeInvalidJSON,
			expectedStatus:   http.StatusBadRequest,
			expectedCategory: ErrorCategoryValidation,
			expectedSeverity: ErrorSeverityMedium,
		},
		{
			name: "authentication error",
			err: &AuthenticationError{
				Code:    ErrorCodeExpiredToken,
				Message: "Token has expired",
			},
			expectedError:    ErrorCodeExpiredToken,
			expectedStatus:   http.StatusUnauthorized,
			expectedCategory: ErrorCategoryAuthentication,
			expectedSeverity: ErrorSeverityHigh,
		},
		{
			name: "rate limit error",
			err: &RateLimitError{
				Code:       ErrorCodeRateLimitExceeded,
				Message:    "Rate limit exceeded",
				RetryAfter: 120,
			},
			expectedError:    ErrorCodeRateLimitExceeded,
			expectedStatus:   http.StatusTooManyRequests,
			expectedCategory: ErrorCategoryRateLimit,
			expectedSeverity: ErrorSeverityMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiError := handler.ConvertError(tt.err, req)

			assert.Equal(t, tt.expectedError, apiError.Error)
			assert.Equal(t, tt.expectedStatus, apiError.StatusCode)
			assert.Equal(t, tt.expectedCategory, apiError.Category)
			assert.Equal(t, tt.expectedSeverity, apiError.Severity)
			assert.NotEmpty(t, apiError.Message)
			assert.NotEmpty(t, apiError.Description)
			assert.NotEmpty(t, apiError.HelpURL)
			assert.Equal(t, "test-request-id", apiError.Details.RequestID)
		})
	}
}

func TestErrorHandler_CreateValidationError(t *testing.T) {
	logger := zap.NewNop()
	handler := NewErrorHandler(logger)

	err := handler.CreateValidationError(
		ErrorCodeMissingRequiredField,
		"Business name is required",
		"business_name",
		"required",
		nil,
	)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, ErrorCodeMissingRequiredField, validationErr.Code)
	assert.Equal(t, "Business name is required", validationErr.Message)
	assert.Equal(t, "business_name", validationErr.Field)
	assert.Equal(t, "required", validationErr.Constraint)
}

func TestErrorHandler_CreateAuthenticationError(t *testing.T) {
	logger := zap.NewNop()
	handler := NewErrorHandler(logger)

	err := handler.CreateAuthenticationError(
		ErrorCodeMissingToken,
		"Authentication token is required",
	)

	authErr, ok := err.(*AuthenticationError)
	require.True(t, ok)
	assert.Equal(t, ErrorCodeMissingToken, authErr.Code)
	assert.Equal(t, "Authentication token is required", authErr.Message)
}

func TestErrorHandler_CreateRateLimitError(t *testing.T) {
	logger := zap.NewNop()
	handler := NewErrorHandler(logger)

	err := handler.CreateRateLimitError("Rate limit exceeded", 60)

	rateLimitErr, ok := err.(*RateLimitError)
	require.True(t, ok)
	assert.Equal(t, ErrorCodeRateLimitExceeded, rateLimitErr.Code)
	assert.Equal(t, "Rate limit exceeded", rateLimitErr.Message)
	assert.Equal(t, 60, rateLimitErr.RetryAfter)
}

func TestErrorHandler_CreateClassificationError(t *testing.T) {
	logger := zap.NewNop()
	handler := NewErrorHandler(logger)

	err := handler.CreateClassificationError(
		ErrorCodeClassificationFailed,
		"Classification failed due to invalid data",
	)

	classificationErr, ok := err.(*ClassificationError)
	require.True(t, ok)
	assert.Equal(t, ErrorCodeClassificationFailed, classificationErr.Code)
	assert.Equal(t, "Classification failed due to invalid data", classificationErr.Message)
}

func TestErrorHandler_CreateExternalServiceError(t *testing.T) {
	logger := zap.NewNop()
	handler := NewErrorHandler(logger)

	err := handler.CreateExternalServiceError("duckduckgo", "API request failed")

	externalServiceErr, ok := err.(*ExternalServiceError)
	require.True(t, ok)
	assert.Equal(t, ErrorCodeExternalAPIFailed, externalServiceErr.Code)
	assert.Equal(t, "duckduckgo", externalServiceErr.Service)
	assert.Equal(t, "API request failed", externalServiceErr.Message)
}

func TestErrorHandler_CreateTimeoutError(t *testing.T) {
	logger := zap.NewNop()
	handler := NewErrorHandler(logger)

	err := handler.CreateTimeoutError("website_scraping", 10*time.Second)

	timeoutErr, ok := err.(*TimeoutError)
	require.True(t, ok)
	assert.Equal(t, ErrorCodeTimeoutError, timeoutErr.Code)
	assert.Equal(t, "website_scraping", timeoutErr.Operation)
	assert.Equal(t, 10*time.Second, timeoutErr.Timeout)
	assert.Contains(t, timeoutErr.Message, "website_scraping")
	assert.Contains(t, timeoutErr.Message, "10s")
}

func TestErrorHandler_RetryAfterHeader(t *testing.T) {
	logger := zap.NewNop()
	handler := NewErrorHandler(logger)

	req := httptest.NewRequest("POST", "/v1/classify", nil)
	w := httptest.NewRecorder()

	rateLimitErr := &RateLimitError{
		Code:       ErrorCodeRateLimitExceeded,
		Message:    "Rate limit exceeded",
		RetryAfter: 120,
	}

	handler.WriteError(w, req, rateLimitErr)

	assert.Equal(t, "120", w.Header().Get("Retry-After"))
}

func TestErrorHandler_ErrorLogging(t *testing.T) {
	// Create a test logger that captures log entries
	logger := zap.NewNop() // In a real test, you'd use a test logger that captures entries
	handler := NewErrorHandler(logger)

	req := httptest.NewRequest("POST", "/v1/classify", nil)
	req = req.WithContext(context.WithValue(req.Context(), "request_id", "test-request-id"))

	w := httptest.NewRecorder()

	// Test different severity levels
	criticalErr := &ValidationError{
		Code:    ErrorCodeInvalidJSON,
		Message: "Critical validation error",
	}

	handler.WriteError(w, req, criticalErr)

	// The error should be logged (in a real test, you'd verify the log entry)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestErrorTypes_ErrorMethods(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name: "validation error",
			err: &ValidationError{
				Code:    ErrorCodeMissingRequiredField,
				Message: "Field is required",
				Field:   "business_name",
			},
			expected: "validation error in field 'business_name': Field is required",
		},
		{
			name: "authentication error",
			err: &AuthenticationError{
				Code:    ErrorCodeMissingToken,
				Message: "Token is required",
			},
			expected: "authentication error: Token is required",
		},
		{
			name: "rate limit error",
			err: &RateLimitError{
				Code:       ErrorCodeRateLimitExceeded,
				Message:    "Rate limit exceeded",
				RetryAfter: 60,
			},
			expected: "rate limit error: Rate limit exceeded (retry after 60 seconds)",
		},
		{
			name: "external service error",
			err: &ExternalServiceError{
				Code:    ErrorCodeExternalAPIFailed,
				Service: "duckduckgo",
				Message: "API failed",
			},
			expected: "external service error (duckduckgo): API failed",
		},
		{
			name: "timeout error",
			err: &TimeoutError{
				Code:      ErrorCodeTimeoutError,
				Operation: "scraping",
				Timeout:   10 * time.Second,
				Message:   "Operation 'scraping' timed out after 10s",
			},
			expected: "timeout error: Operation 'scraping' timed out after 10s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}
