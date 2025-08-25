package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestErrorLogger_LogError(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &ValidationError{
		Code:    ErrorCodeMissingRequiredField,
		Message: "Test error",
		Field:   "test_field",
	}

	errorLogger.LogError(ctx, err, zap.String("test_key", "test_value"))

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	assert.Equal(t, "API Error", logEntry.Message)
	assert.Equal(t, zap.ErrorLevel, logEntry.Level)
	assert.Contains(t, logEntry.ContextMap(), "correlation_id")
	assert.Contains(t, logEntry.ContextMap(), "test_key")
	assert.Contains(t, logEntry.ContextMap(), "timestamp")
}

func TestErrorLogger_LogValidationError(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &ValidationError{
		Code:       ErrorCodeMissingRequiredField,
		Message:    "Test validation error",
		Field:      "business_name",
		Constraint: "required",
		Value:      "",
	}

	requestData := map[string]interface{}{
		"business_name": "",
		"website_url":   "https://example.com",
	}

	errorLogger.LogValidationError(ctx, err, requestData)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	assert.Equal(t, "API Error", logEntry.Message)
	assert.Equal(t, zap.ErrorLevel, logEntry.Level)

	contextMap := logEntry.ContextMap()
	assert.Equal(t, "validation_error", contextMap["error_type"])
	assert.Equal(t, string(ErrorCodeMissingRequiredField), contextMap["error_code"])
	assert.Equal(t, "business_name", contextMap["field"])
	assert.Equal(t, "required", contextMap["constraint"])
	assert.Equal(t, "", contextMap["value"])
	assert.NotNil(t, contextMap["request_data"])
}

func TestErrorLogger_LogAuthenticationError(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &AuthenticationError{
		Code:    ErrorCodeInvalidToken,
		Message: "Invalid API key",
	}

	requestInfo := map[string]interface{}{
		"api_key": "invalid-key",
		"ip":      "192.168.1.1",
	}

	errorLogger.LogAuthenticationError(ctx, err, requestInfo)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	contextMap := logEntry.ContextMap()
	assert.Equal(t, "authentication_error", contextMap["error_type"])
	assert.Equal(t, string(ErrorCodeInvalidToken), contextMap["error_code"])
	assert.NotNil(t, contextMap["request_info"])
}

func TestErrorLogger_LogAuthorizationError(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &AuthorizationError{
		Code:    ErrorCodeInsufficientPermissions,
		Message: "Insufficient permissions",
	}

	userInfo := map[string]interface{}{
		"user_id": "user-123",
		"role":    "user",
	}

	errorLogger.LogAuthorizationError(ctx, err, userInfo)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	contextMap := logEntry.ContextMap()
	assert.Equal(t, "authorization_error", contextMap["error_type"])
	assert.Equal(t, string(ErrorCodeInsufficientPermissions), contextMap["error_code"])
	assert.NotNil(t, contextMap["user_info"])
}

func TestErrorLogger_LogRateLimitError(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &RateLimitError{
		Code:       ErrorCodeRateLimitExceeded,
		Message:    "Rate limit exceeded",
		RetryAfter: 60,
	}

	clientInfo := map[string]interface{}{
		"client_id": "client-123",
		"ip":        "192.168.1.1",
	}

	errorLogger.LogRateLimitError(ctx, err, clientInfo)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	contextMap := logEntry.ContextMap()
	assert.Equal(t, "rate_limit_error", contextMap["error_type"])
	assert.Equal(t, string(ErrorCodeRateLimitExceeded), contextMap["error_code"])
	assert.Equal(t, int64(60), contextMap["retry_after"])
	assert.NotNil(t, contextMap["client_info"])
}

func TestErrorLogger_LogClassificationError(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &ClassificationError{
		Code:    ErrorCodeClassificationFailed,
		Message: "Classification failed",
	}

	classificationData := map[string]interface{}{
		"business_name": "Test Business",
		"confidence":    0.5,
	}

	errorLogger.LogClassificationError(ctx, err, classificationData)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	contextMap := logEntry.ContextMap()
	assert.Equal(t, "classification_error", contextMap["error_type"])
	assert.Equal(t, string(ErrorCodeClassificationFailed), contextMap["error_code"])
	assert.NotNil(t, contextMap["classification_data"])
}

func TestErrorLogger_LogExternalServiceError(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &ExternalServiceError{
		Code:    ErrorCodeExternalAPIFailed,
		Service: "government_api",
		Message: "External service unavailable",
	}

	serviceInfo := map[string]interface{}{
		"service_url": "https://api.government.gov",
		"retry_count": 3,
	}

	errorLogger.LogExternalServiceError(ctx, err, serviceInfo)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	contextMap := logEntry.ContextMap()
	assert.Equal(t, "external_service_error", contextMap["error_type"])
	assert.Equal(t, string(ErrorCodeExternalAPIFailed), contextMap["error_code"])
	assert.Equal(t, "government_api", contextMap["service"])
	assert.NotNil(t, contextMap["service_info"])
}

func TestErrorLogger_LogTimeoutError(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &TimeoutError{
		Code:      ErrorCodeTimeoutError,
		Message:   "Request timeout",
		Operation: "business_classification",
		Timeout:   30 * time.Second,
	}

	operationInfo := map[string]interface{}{
		"business_name": "Test Business",
		"algorithm":     "ml_classifier",
	}

	errorLogger.LogTimeoutError(ctx, err, operationInfo)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	contextMap := logEntry.ContextMap()
	assert.Equal(t, "timeout_error", contextMap["error_type"])
	assert.Equal(t, string(ErrorCodeTimeoutError), contextMap["error_code"])
	assert.Equal(t, "business_classification", contextMap["operation"])
	assert.NotNil(t, contextMap["timeout"])
	assert.NotNil(t, contextMap["operation_info"])
}

func TestErrorLogger_LogInternalError(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &ClassificationError{
		Code:    ErrorCodeClassificationFailed,
		Message: "Internal classification error",
	}

	additionalData := map[string]interface{}{
		"component": "ml_classifier",
		"version":   "1.0.0",
	}

	errorLogger.LogInternalError(ctx, err, "classification_service", additionalData)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	contextMap := logEntry.ContextMap()
	assert.Equal(t, "internal_error", contextMap["error_type"])
	assert.Equal(t, "classification_service", contextMap["component"])
	assert.NotNil(t, contextMap["additional_data"])
}

func TestErrorLogger_LogErrorWithSeverity(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &ValidationError{
		Code:    ErrorCodeMissingRequiredField,
		Message: "Test error",
	}

	// Test different severity levels
	testCases := []struct {
		severity zapcore.Level
		expected int
	}{
		{zapcore.DebugLevel, 1},
		{zapcore.InfoLevel, 1},
		{zapcore.WarnLevel, 1},
		{zapcore.ErrorLevel, 1},
	}

	for _, tc := range testCases {
		logs.TakeAll() // Clear previous logs
		errorLogger.LogErrorWithSeverity(ctx, err, tc.severity)
		assert.Equal(t, tc.expected, logs.Len())
	}
}

func TestErrorLogger_LogErrorWithStack(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &ValidationError{
		Code:    ErrorCodeMissingRequiredField,
		Message: "Test error",
	}

	errorLogger.LogErrorWithStack(ctx, err)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	contextMap := logEntry.ContextMap()
	assert.Contains(t, contextMap, "stack_trace")
}

func TestErrorLogger_LogErrorWithContext(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	err := &ValidationError{
		Code:    ErrorCodeMissingRequiredField,
		Message: "Test error",
	}

	contextData := map[string]interface{}{
		"request_id": "req-123",
		"user_agent": "test-agent",
	}

	errorLogger.LogErrorWithContext(ctx, err, contextData)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	contextMap := logEntry.ContextMap()
	assert.NotNil(t, contextMap["context_data"])
}

func TestErrorLogger_LogErrorSummary(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	summary := map[string]interface{}{
		"total_errors":      100,
		"validation_errors": 50,
		"auth_errors":       30,
		"timeout_errors":    20,
	}

	errorLogger.LogErrorSummary(ctx, summary)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	assert.Equal(t, "Error Summary", logEntry.Message)
	assert.Equal(t, zap.InfoLevel, logEntry.Level)

	contextMap := logEntry.ContextMap()
	assert.Equal(t, "error_summary", contextMap["log_type"])
	assert.NotNil(t, contextMap["summary"])
}

func TestErrorLogger_LogErrorTrend(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	trend := map[string]interface{}{
		"period":          "1h",
		"error_rate":      0.05,
		"trend_direction": "increasing",
		"peak_error_time": "2023-01-01T10:00:00Z",
	}

	errorLogger.LogErrorTrend(ctx, trend)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	assert.Equal(t, "Error Trend", logEntry.Message)
	assert.Equal(t, zap.InfoLevel, logEntry.Level)

	contextMap := logEntry.ContextMap()
	assert.Equal(t, "error_trend", contextMap["log_type"])
	assert.NotNil(t, contextMap["trend"])
}

func TestErrorLogger_GetCorrelationID(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	// Test with correlation ID in context
	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	correlationID := errorLogger.getCorrelationID(ctx)
	assert.Equal(t, "test-correlation-123", correlationID)

	// Test without correlation ID in context
	ctx2 := context.Background()
	correlationID2 := errorLogger.getCorrelationID(ctx2)
	assert.Equal(t, "unknown", correlationID2)
}

func TestErrorLogger_Integration(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	errorLogger := NewErrorLogger(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "integration-test-123")

	// Test multiple error types in sequence
	validationErr := &ValidationError{
		Code:    ErrorCodeMissingRequiredField,
		Message: "Business name is required",
		Field:   "business_name",
	}

	authErr := &AuthenticationError{
		Code:    ErrorCodeInvalidToken,
		Message: "Invalid API key",
	}

	rateLimitErr := &RateLimitError{
		Code:       ErrorCodeRateLimitExceeded,
		Message:    "Rate limit exceeded",
		RetryAfter: 60,
	}

	// Log all errors
	errorLogger.LogValidationError(ctx, validationErr, nil)
	errorLogger.LogAuthenticationError(ctx, authErr, nil)
	errorLogger.LogRateLimitError(ctx, rateLimitErr, nil)

	// Verify all errors were logged
	assert.Equal(t, 3, logs.Len())

	// Verify each error has the correct type
	logEntries := logs.All()
	assert.Equal(t, "validation_error", logEntries[0].ContextMap()["error_type"])
	assert.Equal(t, "authentication_error", logEntries[1].ContextMap()["error_type"])
	assert.Equal(t, "rate_limit_error", logEntries[2].ContextMap()["error_type"])

	// Verify all have correlation ID
	for _, entry := range logEntries {
		assert.Equal(t, "integration-test-123", entry.ContextMap()["correlation_id"])
	}
}
