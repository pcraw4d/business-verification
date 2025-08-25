package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	// Client Errors (4xx)
	ErrorCategoryValidation     ErrorCategory = "validation_error"
	ErrorCategoryAuthentication ErrorCategory = "authentication_error"
	ErrorCategoryAuthorization  ErrorCategory = "authorization_error"
	ErrorCategoryNotFound       ErrorCategory = "not_found_error"
	ErrorCategoryConflict       ErrorCategory = "conflict_error"
	ErrorCategoryRateLimit      ErrorCategory = "rate_limit_error"
	ErrorCategoryRequestTimeout ErrorCategory = "request_timeout_error"

	// Server Errors (5xx)
	ErrorCategoryInternalServer     ErrorCategory = "internal_server_error"
	ErrorCategoryServiceUnavailable ErrorCategory = "service_unavailable_error"
	ErrorCategoryGatewayTimeout     ErrorCategory = "gateway_timeout_error"
	ErrorCategoryDatabaseError      ErrorCategory = "database_error"
	ErrorCategoryExternalService    ErrorCategory = "external_service_error"

	// Business Logic Errors
	ErrorCategoryClassification ErrorCategory = "classification_error"
	ErrorCategoryVerification   ErrorCategory = "verification_error"
	ErrorCategoryRiskAssessment ErrorCategory = "risk_assessment_error"
	ErrorCategoryDataExtraction ErrorCategory = "data_extraction_error"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "low"
	ErrorSeverityMedium   ErrorSeverity = "medium"
	ErrorSeverityHigh     ErrorSeverity = "high"
	ErrorSeverityCritical ErrorSeverity = "critical"
)

// ErrorCode represents a specific error code
type ErrorCode string

const (
	// Validation Errors
	ErrorCodeInvalidJSON          ErrorCode = "INVALID_JSON"
	ErrorCodeMissingRequiredField ErrorCode = "MISSING_REQUIRED_FIELD"
	ErrorCodeInvalidFieldFormat   ErrorCode = "INVALID_FIELD_FORMAT"
	ErrorCodeFieldTooLong         ErrorCode = "FIELD_TOO_LONG"
	ErrorCodeFieldTooShort        ErrorCode = "FIELD_TOO_SHORT"
	ErrorCodeInvalidURL           ErrorCode = "INVALID_URL"
	ErrorCodeInvalidEmail         ErrorCode = "INVALID_EMAIL"
	ErrorCodeInvalidPhone         ErrorCode = "INVALID_PHONE"

	// Authentication Errors
	ErrorCodeMissingToken       ErrorCode = "MISSING_TOKEN"
	ErrorCodeInvalidToken       ErrorCode = "INVALID_TOKEN"
	ErrorCodeExpiredToken       ErrorCode = "EXPIRED_TOKEN"
	ErrorCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"

	// Authorization Errors
	ErrorCodeInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"
	ErrorCodeResourceAccessDenied    ErrorCode = "RESOURCE_ACCESS_DENIED"

	// Rate Limiting Errors
	ErrorCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrorCodeTooManyRequests   ErrorCode = "TOO_MANY_REQUESTS"

	// Classification Errors
	ErrorCodeClassificationFailed ErrorCode = "CLASSIFICATION_FAILED"
	ErrorCodeInvalidBusinessData  ErrorCode = "INVALID_BUSINESS_DATA"
	ErrorCodeWebsiteUnreachable   ErrorCode = "WEBSITE_UNREACHABLE"
	ErrorCodeExternalAPIFailed    ErrorCode = "EXTERNAL_API_FAILED"

	// Server Errors
	ErrorCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrorCodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeTimeoutError       ErrorCode = "TIMEOUT_ERROR"
)

// ErrorDetails provides additional context about an error
type ErrorDetails struct {
	Field      string                 `json:"field,omitempty"`
	Constraint string                 `json:"constraint,omitempty"`
	Value      interface{}            `json:"value,omitempty"`
	Expected   interface{}            `json:"expected,omitempty"`
	Actual     interface{}            `json:"actual,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	RetryAfter *int                   `json:"retry_after,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// APIError represents a structured API error response
type APIError struct {
	Error       ErrorCode     `json:"error"`
	Category    ErrorCategory `json:"category"`
	Severity    ErrorSeverity `json:"severity"`
	Message     string        `json:"message"`
	Description string        `json:"description"`
	Details     ErrorDetails  `json:"details"`
	StatusCode  int           `json:"status_code"`
	RetryAfter  *int          `json:"retry_after,omitempty"`
	HelpURL     string        `json:"help_url,omitempty"`
}

// ErrorHandler provides comprehensive error handling for the API
type ErrorHandler struct {
	logger *zap.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *zap.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// WriteError writes a structured error response to the HTTP response writer
func (h *ErrorHandler) WriteError(w http.ResponseWriter, r *http.Request, err error) {
	apiError := h.ConvertError(err, r)

	// Log the error with context
	h.logError(r.Context(), apiError, err)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	if apiError.RetryAfter != nil {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", *apiError.RetryAfter))
	}
	w.WriteHeader(apiError.StatusCode)

	// Write error response
	if err := json.NewEncoder(w).Encode(apiError); err != nil {
		h.logger.Error("Failed to encode error response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ConvertError converts a generic error to a structured API error
func (h *ErrorHandler) ConvertError(err error, r *http.Request) *APIError {
	// Extract request ID from context
	requestID := ""
	if ctx := r.Context(); ctx != nil {
		if id, ok := ctx.Value("request_id").(string); ok {
			requestID = id
		}
	}

	// Default error details
	details := ErrorDetails{
		RequestID: requestID,
		Timestamp: time.Now().UTC(),
		Context: map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"user_agent": r.UserAgent(),
		},
	}

	// Convert specific error types to structured errors
	switch e := err.(type) {
	case *ValidationError:
		return &APIError{
			Error:       e.Code,
			Category:    ErrorCategoryValidation,
			Severity:    ErrorSeverityMedium,
			Message:     e.Message,
			Description: "The request data failed validation",
			Details:     details,
			StatusCode:  http.StatusBadRequest,
			HelpURL:     "https://docs.kyb-platform.com/api/errors/validation",
		}
	case *AuthenticationError:
		return &APIError{
			Error:       e.Code,
			Category:    ErrorCategoryAuthentication,
			Severity:    ErrorSeverityHigh,
			Message:     e.Message,
			Description: "Authentication failed",
			Details:     details,
			StatusCode:  http.StatusUnauthorized,
			HelpURL:     "https://docs.kyb-platform.com/api/errors/authentication",
		}
	case *AuthorizationError:
		return &APIError{
			Error:       e.Code,
			Category:    ErrorCategoryAuthorization,
			Severity:    ErrorSeverityHigh,
			Message:     e.Message,
			Description: "Authorization failed",
			Details:     details,
			StatusCode:  http.StatusForbidden,
			HelpURL:     "https://docs.kyb-platform.com/api/errors/authorization",
		}
	case *RateLimitError:
		return &APIError{
			Error:       e.Code,
			Category:    ErrorCategoryRateLimit,
			Severity:    ErrorSeverityMedium,
			Message:     e.Message,
			Description: "Rate limit exceeded",
			Details:     details,
			StatusCode:  http.StatusTooManyRequests,
			RetryAfter:  &e.RetryAfter,
			HelpURL:     "https://docs.kyb-platform.com/api/errors/rate-limiting",
		}
	case *ClassificationError:
		return &APIError{
			Error:       e.Code,
			Category:    ErrorCategoryClassification,
			Severity:    ErrorSeverityMedium,
			Message:     e.Message,
			Description: "Business classification failed",
			Details:     details,
			StatusCode:  http.StatusUnprocessableEntity,
			HelpURL:     "https://docs.kyb-platform.com/api/errors/classification",
		}
	case *ExternalServiceError:
		return &APIError{
			Error:       e.Code,
			Category:    ErrorCategoryExternalService,
			Severity:    ErrorSeverityHigh,
			Message:     e.Message,
			Description: "External service error",
			Details:     details,
			StatusCode:  http.StatusBadGateway,
			HelpURL:     "https://docs.kyb-platform.com/api/errors/external-services",
		}
	case *TimeoutError:
		return &APIError{
			Error:       e.Code,
			Category:    ErrorCategoryRequestTimeout,
			Severity:    ErrorSeverityMedium,
			Message:     e.Message,
			Description: "Request timeout",
			Details:     details,
			StatusCode:  http.StatusRequestTimeout,
			HelpURL:     "https://docs.kyb-platform.com/api/errors/timeouts",
		}
	default:
		// Generic error handling
		return &APIError{
			Error:       ErrorCodeInternalError,
			Category:    ErrorCategoryInternalServer,
			Severity:    ErrorSeverityCritical,
			Message:     "An unexpected error occurred",
			Description: "Please try again later or contact support if the problem persists",
			Details:     details,
			StatusCode:  http.StatusInternalServerError,
			HelpURL:     "https://docs.kyb-platform.com/api/errors/internal",
		}
	}
}

// logError logs the error with appropriate level and context
func (h *ErrorHandler) logError(ctx context.Context, apiError *APIError, originalErr error) {
	logger := h.logger.With(
		zap.String("error_code", string(apiError.Error)),
		zap.String("error_category", string(apiError.Category)),
		zap.String("error_severity", string(apiError.Severity)),
		zap.Int("status_code", apiError.StatusCode),
		zap.String("request_id", apiError.Details.RequestID),
		zap.String("path", apiError.Details.Context["path"].(string)),
		zap.String("method", apiError.Details.Context["method"].(string)),
	)

	switch apiError.Severity {
	case ErrorSeverityCritical:
		logger.Error("Critical error occurred", zap.Error(originalErr))
	case ErrorSeverityHigh:
		logger.Error("High severity error occurred", zap.Error(originalErr))
	case ErrorSeverityMedium:
		logger.Warn("Medium severity error occurred", zap.Error(originalErr))
	case ErrorSeverityLow:
		logger.Info("Low severity error occurred", zap.Error(originalErr))
	}
}

// CreateValidationError creates a validation error
func (h *ErrorHandler) CreateValidationError(code ErrorCode, message, field, constraint string, value interface{}) error {
	return &ValidationError{
		Code:       code,
		Message:    message,
		Field:      field,
		Constraint: constraint,
		Value:      value,
	}
}

// CreateAuthenticationError creates an authentication error
func (h *ErrorHandler) CreateAuthenticationError(code ErrorCode, message string) error {
	return &AuthenticationError{
		Code:    code,
		Message: message,
	}
}

// CreateAuthorizationError creates an authorization error
func (h *ErrorHandler) CreateAuthorizationError(code ErrorCode, message string) error {
	return &AuthorizationError{
		Code:    code,
		Message: message,
	}
}

// CreateRateLimitError creates a rate limit error
func (h *ErrorHandler) CreateRateLimitError(message string, retryAfter int) error {
	return &RateLimitError{
		Code:       ErrorCodeRateLimitExceeded,
		Message:    message,
		RetryAfter: retryAfter,
	}
}

// CreateClassificationError creates a classification error
func (h *ErrorHandler) CreateClassificationError(code ErrorCode, message string) error {
	return &ClassificationError{
		Code:    code,
		Message: message,
	}
}

// CreateExternalServiceError creates an external service error
func (h *ErrorHandler) CreateExternalServiceError(service, message string) error {
	return &ExternalServiceError{
		Code:    ErrorCodeExternalAPIFailed,
		Service: service,
		Message: message,
	}
}

// CreateTimeoutError creates a timeout error
func (h *ErrorHandler) CreateTimeoutError(operation string, timeout time.Duration) error {
	return &TimeoutError{
		Code:      ErrorCodeTimeoutError,
		Operation: operation,
		Timeout:   timeout,
		Message:   fmt.Sprintf("Operation '%s' timed out after %v", operation, timeout),
	}
}
