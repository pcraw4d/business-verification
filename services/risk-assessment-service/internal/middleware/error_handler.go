package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
	Path      string      `json:"path"`
	Method    string      `json:"method"`
}

// ErrorDetail contains detailed error information
type ErrorDetail struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	Field      string            `json:"field,omitempty"`
	Validation []ValidationError `json:"validation,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ErrorHandler provides comprehensive error handling
type ErrorHandler struct {
	logger *zap.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *zap.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError handles errors and returns appropriate HTTP responses
func (eh *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	// Get request ID from context
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	// Determine error type and status code
	statusCode, errorDetail := eh.classifyError(err)

	// Create error response
	errorResponse := ErrorResponse{
		Error:     errorDetail,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
		Method:    r.Method,
	}

	// Log error
	eh.logError(r, err, statusCode, requestID)

	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)

	// Write response
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

// classifyError classifies an error and returns appropriate status code and details
func (eh *ErrorHandler) classifyError(err error) (int, ErrorDetail) {
	if err == nil {
		return http.StatusOK, ErrorDetail{
			Code:    "SUCCESS",
			Message: "Request completed successfully",
		}
	}

	// Check for specific error types
	switch {
	case isValidationError(err):
		return http.StatusBadRequest, ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: "Request validation failed",
			Details: err.Error(),
		}

	case isAuthenticationError(err):
		return http.StatusUnauthorized, ErrorDetail{
			Code:    "AUTHENTICATION_ERROR",
			Message: "Authentication required",
			Details: err.Error(),
		}

	case isAuthorizationError(err):
		return http.StatusForbidden, ErrorDetail{
			Code:    "AUTHORIZATION_ERROR",
			Message: "Insufficient permissions",
			Details: err.Error(),
		}

	case isNotFoundError(err):
		return http.StatusNotFound, ErrorDetail{
			Code:    "NOT_FOUND",
			Message: "Resource not found",
			Details: err.Error(),
		}

	case isConflictError(err):
		return http.StatusConflict, ErrorDetail{
			Code:    "CONFLICT",
			Message: "Resource conflict",
			Details: err.Error(),
		}

	case isRateLimitError(err):
		return http.StatusTooManyRequests, ErrorDetail{
			Code:    "RATE_LIMIT_EXCEEDED",
			Message: "Rate limit exceeded",
			Details: err.Error(),
		}

	case isServiceUnavailableError(err):
		return http.StatusServiceUnavailable, ErrorDetail{
			Code:    "SERVICE_UNAVAILABLE",
			Message: "Service temporarily unavailable",
			Details: err.Error(),
		}

	case isTimeoutError(err):
		return http.StatusRequestTimeout, ErrorDetail{
			Code:    "REQUEST_TIMEOUT",
			Message: "Request timeout",
			Details: err.Error(),
		}

	case isInternalError(err):
		return http.StatusInternalServerError, ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: "Internal server error",
			Details: "An unexpected error occurred",
		}

	default:
		return http.StatusInternalServerError, ErrorDetail{
			Code:    "UNKNOWN_ERROR",
			Message: "An unknown error occurred",
			Details: err.Error(),
		}
	}
}

// logError logs the error with appropriate level
func (eh *ErrorHandler) logError(r *http.Request, err error, statusCode int, requestID string) {
	// Get stack trace for internal errors
	var stackTrace string
	if statusCode >= 500 {
		stackTrace = getStackTrace()
	}

	// Log with appropriate level
	fields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.Int("status_code", statusCode),
		zap.Error(err),
	}

	if stackTrace != "" {
		fields = append(fields, zap.String("stack_trace", stackTrace))
	}

	switch {
	case statusCode >= 500:
		eh.logger.Error("Internal server error", fields...)
	case statusCode >= 400:
		eh.logger.Warn("Client error", fields...)
	default:
		eh.logger.Info("Request completed", fields...)
	}
}

// getStackTrace returns the current stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// Error type checking functions
func isValidationError(err error) bool {
	return strings.Contains(err.Error(), "validation") ||
		strings.Contains(err.Error(), "invalid") ||
		strings.Contains(err.Error(), "required") ||
		strings.Contains(err.Error(), "format")
}

func isAuthenticationError(err error) bool {
	return strings.Contains(err.Error(), "authentication") ||
		strings.Contains(err.Error(), "unauthorized") ||
		strings.Contains(err.Error(), "token") ||
		strings.Contains(err.Error(), "credentials")
}

func isAuthorizationError(err error) bool {
	return strings.Contains(err.Error(), "authorization") ||
		strings.Contains(err.Error(), "permission") ||
		strings.Contains(err.Error(), "forbidden") ||
		strings.Contains(err.Error(), "access denied")
}

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "not found") ||
		strings.Contains(err.Error(), "does not exist") ||
		strings.Contains(err.Error(), "no such")
}

func isConflictError(err error) bool {
	return strings.Contains(err.Error(), "conflict") ||
		strings.Contains(err.Error(), "already exists") ||
		strings.Contains(err.Error(), "duplicate")
}

func isRateLimitError(err error) bool {
	return strings.Contains(err.Error(), "rate limit") ||
		strings.Contains(err.Error(), "too many requests") ||
		strings.Contains(err.Error(), "quota exceeded")
}

func isServiceUnavailableError(err error) bool {
	return strings.Contains(err.Error(), "service unavailable") ||
		strings.Contains(err.Error(), "maintenance") ||
		strings.Contains(err.Error(), "temporarily unavailable")
}

func isTimeoutError(err error) bool {
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline exceeded") ||
		strings.Contains(err.Error(), "context canceled")
}

func isInternalError(err error) bool {
	return strings.Contains(err.Error(), "internal") ||
		strings.Contains(err.Error(), "database") ||
		strings.Contains(err.Error(), "connection") ||
		strings.Contains(err.Error(), "panic")
}

// Custom error types for better error handling
type ValidationErrorType struct {
	Field   string
	Message string
	Code    string
}

func (ve ValidationErrorType) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", ve.Field, ve.Message)
}

type AuthenticationError struct {
	Message string
}

func (ae AuthenticationError) Error() string {
	return ae.Message
}

type AuthorizationError struct {
	Message string
}

func (ae AuthorizationError) Error() string {
	return ae.Message
}

type NotFoundError struct {
	Resource string
}

func (nfe NotFoundError) Error() string {
	return fmt.Sprintf("resource '%s' not found", nfe.Resource)
}

type ConflictError struct {
	Resource string
	Message  string
}

func (ce ConflictError) Error() string {
	if ce.Message != "" {
		return ce.Message
	}
	return fmt.Sprintf("conflict with resource '%s'", ce.Resource)
}

type RateLimitError struct {
	Limit     int
	Remaining int
	ResetTime time.Time
}

func (rle RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %d requests per hour, %d remaining, resets at %s",
		rle.Limit, rle.Remaining, rle.ResetTime.Format(time.RFC3339))
}

type ServiceUnavailableError struct {
	Service string
	Message string
}

func (sue ServiceUnavailableError) Error() string {
	if sue.Message != "" {
		return sue.Message
	}
	return fmt.Sprintf("service '%s' is temporarily unavailable", sue.Service)
}

type TimeoutError struct {
	Operation string
	Timeout   time.Duration
}

func (te TimeoutError) Error() string {
	return fmt.Sprintf("operation '%s' timed out after %v", te.Operation, te.Timeout)
}

type InternalError struct {
	Operation string
	Message   string
}

func (ie InternalError) Error() string {
	if ie.Message != "" {
		return ie.Message
	}
	return fmt.Sprintf("internal error in operation '%s'", ie.Operation)
}
