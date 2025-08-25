package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation_error"
	// ErrorTypeAuthentication represents authentication errors
	ErrorTypeAuthentication ErrorType = "authentication_error"
	// ErrorTypeAuthorization represents authorization errors
	ErrorTypeAuthorization ErrorType = "authorization_error"
	// ErrorTypeNotFound represents not found errors
	ErrorTypeNotFound ErrorType = "not_found_error"
	// ErrorTypeConflict represents conflict errors
	ErrorTypeConflict ErrorType = "conflict_error"
	// ErrorTypeRateLimit represents rate limit errors
	ErrorTypeRateLimit ErrorType = "rate_limit_error"
	// ErrorTypeInternal represents internal server errors
	ErrorTypeInternal ErrorType = "internal_error"
	// ErrorTypeExternal represents external service errors
	ErrorTypeExternal ErrorType = "external_error"
	// ErrorTypeTimeout represents timeout errors
	ErrorTypeTimeout ErrorType = "timeout_error"
	// ErrorTypeUnavailable represents service unavailable errors
	ErrorTypeUnavailable ErrorType = "unavailable_error"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	// ErrorSeverityLow represents low severity errors
	ErrorSeverityLow ErrorSeverity = "low"
	// ErrorSeverityMedium represents medium severity errors
	ErrorSeverityMedium ErrorSeverity = "medium"
	// ErrorSeverityHigh represents high severity errors
	ErrorSeverityHigh ErrorSeverity = "high"
	// ErrorSeverityCritical represents critical severity errors
	ErrorSeverityCritical ErrorSeverity = "critical"
)

// APIError represents a standardized API error
type APIError struct {
	Type       ErrorType              `json:"type"`
	Severity   ErrorSeverity          `json:"severity"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Code       string                 `json:"code,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Path       string                 `json:"path,omitempty"`
	Method     string                 `json:"method,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	RemoteAddr string                 `json:"remote_addr,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// ErrorResponse represents the error response sent to clients
type ErrorResponse struct {
	Error   APIError `json:"error"`
	Success bool     `json:"success"`
}

// CustomError represents a custom error with additional context
type CustomError struct {
	Type       ErrorType
	Severity   ErrorSeverity
	Message    string
	Details    string
	Code       string
	StatusCode int
	Context    map[string]interface{}
	Err        error
}

func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

// ErrorHandler defines the interface for custom error handlers
type ErrorHandler interface {
	HandleError(ctx context.Context, err error, req *http.Request) *APIError
}

// ErrorHandlerFunc is a function type that implements ErrorHandler
type ErrorHandlerFunc func(ctx context.Context, err error, req *http.Request) *APIError

func (f ErrorHandlerFunc) HandleError(ctx context.Context, err error, req *http.Request) *APIError {
	return f(ctx, err, req)
}

// ErrorHandlingConfig holds configuration for error handling middleware
type ErrorHandlingConfig struct {
	// Error Logging
	LogErrors    bool          `json:"log_errors" yaml:"log_errors"`
	LogLevel     zapcore.Level `json:"log_level" yaml:"log_level"`
	IncludeStack bool          `json:"include_stack" yaml:"include_stack"`

	// Error Response
	IncludeDetails bool `json:"include_details" yaml:"include_details"`
	IncludeContext bool `json:"include_context" yaml:"include_context"`
	MaskInternal   bool `json:"mask_internal" yaml:"mask_internal"`

	// Recovery
	RecoverPanics bool `json:"recover_panics" yaml:"recover_panics"`

	// Custom Handlers
	CustomHandlers map[ErrorType]ErrorHandler `json:"-" yaml:"-"`

	// Error Metrics
	TrackMetrics bool `json:"track_metrics" yaml:"track_metrics"`

	// Error Codes
	ErrorCodes map[string]string `json:"error_codes" yaml:"error_codes"`

	// Default Error
	DefaultErrorType     ErrorType     `json:"default_error_type" yaml:"default_error_type"`
	DefaultErrorSeverity ErrorSeverity `json:"default_error_severity" yaml:"default_error_severity"`
	DefaultStatusCode    int           `json:"default_status_code" yaml:"default_status_code"`
}

// ErrorHandlingMiddleware provides comprehensive error handling
type ErrorHandlingMiddleware struct {
	config  *ErrorHandlingConfig
	logger  *zap.Logger
	metrics *ErrorMetrics
}

// ErrorMetrics tracks error statistics
type ErrorMetrics struct {
	TotalErrors      int64
	ErrorsByType     map[ErrorType]int64
	ErrorsBySeverity map[ErrorSeverity]int64
	LastError        time.Time
}

// NewErrorHandlingMiddleware creates a new ErrorHandlingMiddleware
func NewErrorHandlingMiddleware(config *ErrorHandlingConfig, logger *zap.Logger) *ErrorHandlingMiddleware {
	if config == nil {
		config = &ErrorHandlingConfig{
			LogErrors:            true,
			LogLevel:             zapcore.ErrorLevel,
			IncludeStack:         false,
			IncludeDetails:       true,
			IncludeContext:       false,
			MaskInternal:         true,
			RecoverPanics:        true,
			CustomHandlers:       make(map[ErrorType]ErrorHandler),
			TrackMetrics:         true,
			ErrorCodes:           make(map[string]string),
			DefaultErrorType:     ErrorTypeInternal,
			DefaultErrorSeverity: ErrorSeverityMedium,
			DefaultStatusCode:    500,
		}
	}

	// Initialize default error codes
	if len(config.ErrorCodes) == 0 {
		config.ErrorCodes = map[string]string{
			"validation_error":     "INVALID_INPUT",
			"authentication_error": "UNAUTHORIZED",
			"authorization_error":  "FORBIDDEN",
			"not_found_error":      "NOT_FOUND",
			"conflict_error":       "CONFLICT",
			"rate_limit_error":     "RATE_LIMITED",
			"internal_error":       "INTERNAL_ERROR",
			"external_error":       "EXTERNAL_ERROR",
			"timeout_error":        "TIMEOUT",
			"unavailable_error":    "SERVICE_UNAVAILABLE",
		}
	}

	return &ErrorHandlingMiddleware{
		config: config,
		logger: logger,
		metrics: &ErrorMetrics{
			ErrorsByType:     make(map[ErrorType]int64),
			ErrorsBySeverity: make(map[ErrorSeverity]int64),
		},
	}
}

// Middleware applies error handling to HTTP requests
func (m *ErrorHandlingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer to capture status codes
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Handle panics if enabled
		if m.config.RecoverPanics {
			defer func() {
				if err := recover(); err != nil {
					m.handlePanic(rw, r, err)
				}
			}()
		}

		// Process request
		next.ServeHTTP(rw, r)

		// Handle errors based on status code
		if rw.statusCode >= 400 {
			m.handleErrorResponse(rw, r, rw.statusCode, nil)
		}
	})
}

// handlePanic handles panic recovery
func (m *ErrorHandlingMiddleware) handlePanic(w http.ResponseWriter, r *http.Request, err interface{}) {
	// Create panic error
	panicErr := &CustomError{
		Type:       ErrorTypeInternal,
		Severity:   ErrorSeverityCritical,
		Message:    "Internal server error",
		Details:    fmt.Sprintf("Panic: %v", err),
		StatusCode: http.StatusInternalServerError,
		Context: map[string]interface{}{
			"panic_value": err,
			"stack_trace": string(debug.Stack()),
		},
	}

	m.handleErrorResponse(w, r, http.StatusInternalServerError, panicErr)
}

// handleErrorResponse handles error responses
func (m *ErrorHandlingMiddleware) handleErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	// Get request ID from context
	requestID := ""
	if ctx := r.Context(); ctx != nil {
		if id, ok := ctx.Value("request_id").(string); ok {
			requestID = id
		}
	}

	// Create API error
	apiError := m.createAPIError(err, r, requestID, statusCode)

	// Log error if enabled
	if m.config.LogErrors {
		m.logError(apiError, err)
	}

	// Track metrics if enabled
	if m.config.TrackMetrics {
		m.trackError(apiError)
	}

	// Create error response
	response := ErrorResponse{
		Error:   *apiError,
		Success: false,
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Error-Type", string(apiError.Type))
	w.Header().Set("X-Error-Code", apiError.Code)
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}

	// Write response
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// createAPIError creates an APIError from an error
func (m *ErrorHandlingMiddleware) createAPIError(err error, r *http.Request, requestID string, statusCode int) *APIError {
	// Handle custom errors
	if customErr, ok := err.(*CustomError); ok {
		return m.createAPIErrorFromCustom(customErr, r, requestID)
	}

	// Handle standard errors
	apiError := &APIError{
		Type:       m.config.DefaultErrorType,
		Severity:   m.config.DefaultErrorSeverity,
		Message:    m.getErrorMessage(err, statusCode),
		Details:    m.getErrorDetails(err),
		Code:       m.getErrorCode(m.config.DefaultErrorType),
		RequestID:  requestID,
		Timestamp:  time.Now(),
		Path:       r.URL.Path,
		Method:     r.Method,
		UserAgent:  r.UserAgent(),
		RemoteAddr: m.getRemoteAddr(r),
	}

	// Add context if enabled
	if m.config.IncludeContext {
		apiError.Context = m.getErrorContext(err, r)
	}

	// Mask internal details if enabled
	if m.config.MaskInternal && apiError.Type == ErrorTypeInternal {
		apiError.Details = "An internal error occurred"
		apiError.Context = nil
	}

	return apiError
}

// createAPIErrorFromCustom creates an APIError from a CustomError
func (m *ErrorHandlingMiddleware) createAPIErrorFromCustom(customErr *CustomError, r *http.Request, requestID string) *APIError {
	apiError := &APIError{
		Type:       customErr.Type,
		Severity:   customErr.Severity,
		Message:    customErr.Message,
		Details:    customErr.Details,
		Code:       customErr.Code,
		RequestID:  requestID,
		Timestamp:  time.Now(),
		Path:       r.URL.Path,
		Method:     r.Method,
		UserAgent:  r.UserAgent(),
		RemoteAddr: m.getRemoteAddr(r),
	}

	// Add context if enabled
	if m.config.IncludeContext && customErr.Context != nil {
		apiError.Context = customErr.Context
	}

	// Mask internal details if enabled
	if m.config.MaskInternal && apiError.Type == ErrorTypeInternal {
		apiError.Details = "An internal error occurred"
		apiError.Context = nil
	}

	return apiError
}

// getErrorMessage gets a user-friendly error message
func (m *ErrorHandlingMiddleware) getErrorMessage(err error, statusCode int) string {
	if err != nil {
		// Check for custom error messages based on status code
		switch statusCode {
		case http.StatusBadRequest:
			return "Invalid request"
		case http.StatusUnauthorized:
			return "Authentication required"
		case http.StatusForbidden:
			return "Access denied"
		case http.StatusNotFound:
			return "Resource not found"
		case http.StatusConflict:
			return "Resource conflict"
		case http.StatusTooManyRequests:
			return "Rate limit exceeded"
		case http.StatusInternalServerError:
			return "Internal server error"
		case http.StatusServiceUnavailable:
			return "Service temporarily unavailable"
		default:
			return err.Error()
		}
	}

	return "An error occurred"
}

// getErrorDetails gets error details
func (m *ErrorHandlingMiddleware) getErrorDetails(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

// getErrorCode gets the error code for an error type
func (m *ErrorHandlingMiddleware) getErrorCode(errorType ErrorType) string {
	if code, exists := m.config.ErrorCodes[string(errorType)]; exists {
		return code
	}
	return "UNKNOWN_ERROR"
}

// getErrorContext gets error context
func (m *ErrorHandlingMiddleware) getErrorContext(err error, r *http.Request) map[string]interface{} {
	context := make(map[string]interface{})

	// Add request context
	context["url"] = r.URL.String()
	context["headers"] = m.getSafeHeaders(r.Header)

	// Add error context if available
	if customErr, ok := err.(*CustomError); ok && customErr.Context != nil {
		for k, v := range customErr.Context {
			context[k] = v
		}
	}

	return context
}

// getSafeHeaders gets safe headers (excluding sensitive ones)
func (m *ErrorHandlingMiddleware) getSafeHeaders(headers http.Header) map[string]string {
	safe := make(map[string]string)
	sensitiveHeaders := []string{"authorization", "cookie", "x-api-key", "x-csrf-token"}

	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		isSensitive := false

		for _, sensitive := range sensitiveHeaders {
			if lowerKey == sensitive {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			safe[key] = "[MASKED]"
		} else {
			safe[key] = strings.Join(values, ", ")
		}
	}

	return safe
}

// getRemoteAddr gets the real remote address
func (m *ErrorHandlingMiddleware) getRemoteAddr(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		if commaIndex := strings.Index(forwardedFor, ","); commaIndex != -1 {
			return strings.TrimSpace(forwardedFor[:commaIndex])
		}
		return strings.TrimSpace(forwardedFor)
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}

// logError logs an error
func (m *ErrorHandlingMiddleware) logError(apiError *APIError, originalErr error) {
	fields := []zap.Field{
		zap.String("error_type", string(apiError.Type)),
		zap.String("error_severity", string(apiError.Severity)),
		zap.String("error_code", apiError.Code),
		zap.String("error_message", apiError.Message),
		zap.String("request_id", apiError.RequestID),
		zap.String("path", apiError.Path),
		zap.String("method", apiError.Method),
		zap.String("remote_addr", apiError.RemoteAddr),
		zap.String("user_agent", apiError.UserAgent),
	}

	if apiError.Details != "" {
		fields = append(fields, zap.String("error_details", apiError.Details))
	}

	if apiError.Context != nil {
		fields = append(fields, zap.Any("error_context", apiError.Context))
	}

	if originalErr != nil {
		fields = append(fields, zap.Error(originalErr))
	}

	switch m.config.LogLevel {
	case zapcore.DebugLevel:
		m.logger.Debug("API Error", fields...)
	case zapcore.InfoLevel:
		m.logger.Info("API Error", fields...)
	case zapcore.WarnLevel:
		m.logger.Warn("API Error", fields...)
	case zapcore.ErrorLevel:
		m.logger.Error("API Error", fields...)
	}
}

// trackError tracks error metrics
func (m *ErrorHandlingMiddleware) trackError(apiError *APIError) {
	m.metrics.TotalErrors++
	m.metrics.ErrorsByType[apiError.Type]++
	m.metrics.ErrorsBySeverity[apiError.Severity]++
	m.metrics.LastError = apiError.Timestamp
}

// GetErrorMetrics returns current error metrics
func (m *ErrorHandlingMiddleware) GetErrorMetrics() *ErrorMetrics {
	return m.metrics
}

// CreateCustomError creates a new CustomError
func CreateCustomError(errorType ErrorType, severity ErrorSeverity, message, details, code string, statusCode int) *CustomError {
	return &CustomError{
		Type:       errorType,
		Severity:   severity,
		Message:    message,
		Details:    details,
		Code:       code,
		StatusCode: statusCode,
		Context:    make(map[string]interface{}),
	}
}

// CreateValidationError creates a validation error
func CreateValidationError(message, details string) *CustomError {
	return CreateCustomError(
		ErrorTypeValidation,
		ErrorSeverityLow,
		message,
		details,
		"INVALID_INPUT",
		http.StatusBadRequest,
	)
}

// CreateAuthenticationError creates an authentication error
func CreateAuthenticationError(message, details string) *CustomError {
	return CreateCustomError(
		ErrorTypeAuthentication,
		ErrorSeverityMedium,
		message,
		details,
		"UNAUTHORIZED",
		http.StatusUnauthorized,
	)
}

// CreateAuthorizationError creates an authorization error
func CreateAuthorizationError(message, details string) *CustomError {
	return CreateCustomError(
		ErrorTypeAuthorization,
		ErrorSeverityMedium,
		message,
		details,
		"FORBIDDEN",
		http.StatusForbidden,
	)
}

// CreateNotFoundError creates a not found error
func CreateNotFoundError(message, details string) *CustomError {
	return CreateCustomError(
		ErrorTypeNotFound,
		ErrorSeverityLow,
		message,
		details,
		"NOT_FOUND",
		http.StatusNotFound,
	)
}

// CreateConflictError creates a conflict error
func CreateConflictError(message, details string) *CustomError {
	return CreateCustomError(
		ErrorTypeConflict,
		ErrorSeverityMedium,
		message,
		details,
		"CONFLICT",
		http.StatusConflict,
	)
}

// CreateRateLimitError creates a rate limit error
func CreateRateLimitError(message, details string) *CustomError {
	return CreateCustomError(
		ErrorTypeRateLimit,
		ErrorSeverityMedium,
		message,
		details,
		"RATE_LIMITED",
		http.StatusTooManyRequests,
	)
}

// CreateInternalError creates an internal error
func CreateInternalError(message, details string) *CustomError {
	return CreateCustomError(
		ErrorTypeInternal,
		ErrorSeverityHigh,
		message,
		details,
		"INTERNAL_ERROR",
		http.StatusInternalServerError,
	)
}

// CreateExternalError creates an external service error
func CreateExternalError(message, details string) *CustomError {
	return CreateCustomError(
		ErrorTypeExternal,
		ErrorSeverityMedium,
		message,
		details,
		"EXTERNAL_ERROR",
		http.StatusBadGateway,
	)
}

// CreateTimeoutError creates a timeout error
func CreateTimeoutError(message, details string) *CustomError {
	return CreateCustomError(
		ErrorTypeTimeout,
		ErrorSeverityMedium,
		message,
		details,
		"TIMEOUT",
		http.StatusRequestTimeout,
	)
}

// CreateUnavailableError creates a service unavailable error
func CreateUnavailableError(message, details string) *CustomError {
	return CreateCustomError(
		ErrorTypeUnavailable,
		ErrorSeverityHigh,
		message,
		details,
		"SERVICE_UNAVAILABLE",
		http.StatusServiceUnavailable,
	)
}

// GetDefaultErrorHandlingConfig returns a default error handling configuration
func GetDefaultErrorHandlingConfig() *ErrorHandlingConfig {
	return &ErrorHandlingConfig{
		LogErrors:      true,
		LogLevel:       zapcore.ErrorLevel,
		IncludeStack:   false,
		IncludeDetails: true,
		IncludeContext: false,
		MaskInternal:   true,
		RecoverPanics:  true,
		CustomHandlers: make(map[ErrorType]ErrorHandler),
		TrackMetrics:   true,
		ErrorCodes: map[string]string{
			"validation_error":     "INVALID_INPUT",
			"authentication_error": "UNAUTHORIZED",
			"authorization_error":  "FORBIDDEN",
			"not_found_error":      "NOT_FOUND",
			"conflict_error":       "CONFLICT",
			"rate_limit_error":     "RATE_LIMITED",
			"internal_error":       "INTERNAL_ERROR",
			"external_error":       "EXTERNAL_ERROR",
			"timeout_error":        "TIMEOUT",
			"unavailable_error":    "SERVICE_UNAVAILABLE",
		},
		DefaultErrorType:     ErrorTypeInternal,
		DefaultErrorSeverity: ErrorSeverityMedium,
		DefaultStatusCode:    500,
	}
}

// GetVerboseErrorHandlingConfig returns a verbose error handling configuration
func GetVerboseErrorHandlingConfig() *ErrorHandlingConfig {
	return &ErrorHandlingConfig{
		LogErrors:      true,
		LogLevel:       zapcore.DebugLevel,
		IncludeStack:   true,
		IncludeDetails: true,
		IncludeContext: true,
		MaskInternal:   false,
		RecoverPanics:  true,
		CustomHandlers: make(map[ErrorType]ErrorHandler),
		TrackMetrics:   true,
		ErrorCodes: map[string]string{
			"validation_error":     "INVALID_INPUT",
			"authentication_error": "UNAUTHORIZED",
			"authorization_error":  "FORBIDDEN",
			"not_found_error":      "NOT_FOUND",
			"conflict_error":       "CONFLICT",
			"rate_limit_error":     "RATE_LIMITED",
			"internal_error":       "INTERNAL_ERROR",
			"external_error":       "EXTERNAL_ERROR",
			"timeout_error":        "TIMEOUT",
			"unavailable_error":    "SERVICE_UNAVAILABLE",
		},
		DefaultErrorType:     ErrorTypeInternal,
		DefaultErrorSeverity: ErrorSeverityMedium,
		DefaultStatusCode:    500,
	}
}

// GetProductionErrorHandlingConfig returns a production error handling configuration
func GetProductionErrorHandlingConfig() *ErrorHandlingConfig {
	return &ErrorHandlingConfig{
		LogErrors:      true,
		LogLevel:       zapcore.ErrorLevel,
		IncludeStack:   false,
		IncludeDetails: false,
		IncludeContext: false,
		MaskInternal:   true,
		RecoverPanics:  true,
		CustomHandlers: make(map[ErrorType]ErrorHandler),
		TrackMetrics:   true,
		ErrorCodes: map[string]string{
			"validation_error":     "INVALID_INPUT",
			"authentication_error": "UNAUTHORIZED",
			"authorization_error":  "FORBIDDEN",
			"not_found_error":      "NOT_FOUND",
			"conflict_error":       "CONFLICT",
			"rate_limit_error":     "RATE_LIMITED",
			"internal_error":       "INTERNAL_ERROR",
			"external_error":       "EXTERNAL_ERROR",
			"timeout_error":        "TIMEOUT",
			"unavailable_error":    "SERVICE_UNAVAILABLE",
		},
		DefaultErrorType:     ErrorTypeInternal,
		DefaultErrorSeverity: ErrorSeverityMedium,
		DefaultStatusCode:    500,
	}
}
