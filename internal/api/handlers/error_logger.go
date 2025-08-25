package handlers

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrorLogger provides structured error logging with context and correlation
type ErrorLogger struct {
	logger *zap.Logger
}

// NewErrorLogger creates a new error logger
func NewErrorLogger(logger *zap.Logger) *ErrorLogger {
	return &ErrorLogger{
		logger: logger,
	}
}

// LogError logs an error with structured context
func (l *ErrorLogger) LogError(ctx context.Context, err error, fields ...zap.Field) {
	// Extract correlation ID from context
	correlationID := l.getCorrelationID(ctx)

	// Add correlation ID to fields
	fields = append(fields, zap.String("correlation_id", correlationID))

	// Add timestamp
	fields = append(fields, zap.Time("timestamp", time.Now()))

	// Log the error with structured fields
	l.logger.Error("API Error", append(fields, zap.Error(err))...)
}

// LogValidationError logs validation errors with detailed context
func (l *ErrorLogger) LogValidationError(ctx context.Context, err *ValidationError, requestData interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "validation_error"),
		zap.String("error_code", string(err.Code)),
		zap.String("field", err.Field),
		zap.String("constraint", err.Constraint),
		zap.Any("value", err.Value),
		zap.Any("request_data", requestData),
	}

	l.LogError(ctx, err, fields...)
}

// LogAuthenticationError logs authentication errors
func (l *ErrorLogger) LogAuthenticationError(ctx context.Context, err *AuthenticationError, requestInfo map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "authentication_error"),
		zap.String("error_code", string(err.Code)),
		zap.Any("request_info", requestInfo),
	}

	l.LogError(ctx, err, fields...)
}

// LogAuthorizationError logs authorization errors
func (l *ErrorLogger) LogAuthorizationError(ctx context.Context, err *AuthorizationError, userInfo map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "authorization_error"),
		zap.String("error_code", string(err.Code)),
		zap.Any("user_info", userInfo),
	}

	l.LogError(ctx, err, fields...)
}

// LogRateLimitError logs rate limit errors
func (l *ErrorLogger) LogRateLimitError(ctx context.Context, err *RateLimitError, clientInfo map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "rate_limit_error"),
		zap.String("error_code", string(err.Code)),
		zap.Int("retry_after", err.RetryAfter),
		zap.Any("client_info", clientInfo),
	}

	l.LogError(ctx, err, fields...)
}

// LogClassificationError logs classification errors
func (l *ErrorLogger) LogClassificationError(ctx context.Context, err *ClassificationError, classificationData interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "classification_error"),
		zap.String("error_code", string(err.Code)),
		zap.Any("classification_data", classificationData),
	}

	l.LogError(ctx, err, fields...)
}

// LogExternalServiceError logs external service errors
func (l *ErrorLogger) LogExternalServiceError(ctx context.Context, err *ExternalServiceError, serviceInfo map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "external_service_error"),
		zap.String("error_code", string(err.Code)),
		zap.String("service", err.Service),
		zap.Any("service_info", serviceInfo),
	}

	l.LogError(ctx, err, fields...)
}

// LogTimeoutError logs timeout errors
func (l *ErrorLogger) LogTimeoutError(ctx context.Context, err *TimeoutError, operationInfo map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "timeout_error"),
		zap.String("error_code", string(err.Code)),
		zap.String("operation", err.Operation),
		zap.Duration("timeout", err.Timeout),
		zap.Any("operation_info", operationInfo),
	}

	l.LogError(ctx, err, fields...)
}

// LogInternalError logs internal server errors
func (l *ErrorLogger) LogInternalError(ctx context.Context, err error, component string, additionalData map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "internal_error"),
		zap.String("component", component),
		zap.Any("additional_data", additionalData),
	}

	l.LogError(ctx, err, fields...)
}

// LogRequestError logs request processing errors
func (l *ErrorLogger) LogRequestError(ctx context.Context, err error, requestInfo map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "request_error"),
		zap.Any("request_info", requestInfo),
	}

	l.LogError(ctx, err, fields...)
}

// LogSecurityError logs security-related errors
func (l *ErrorLogger) LogSecurityError(ctx context.Context, err error, securityInfo map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "security_error"),
		zap.Any("security_info", securityInfo),
	}

	l.LogError(ctx, err, fields...)
}

// LogPerformanceError logs performance-related errors
func (l *ErrorLogger) LogPerformanceError(ctx context.Context, err error, performanceInfo map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "performance_error"),
		zap.Any("performance_info", performanceInfo),
	}

	l.LogError(ctx, err, fields...)
}

// LogBatchError logs batch processing errors
func (l *ErrorLogger) LogBatchError(ctx context.Context, err error, batchInfo map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_type", "batch_error"),
		zap.Any("batch_info", batchInfo),
	}

	l.LogError(ctx, err, fields...)
}

// LogErrorWithMetrics logs an error and records metrics
func (l *ErrorLogger) LogErrorWithMetrics(ctx context.Context, err error, errorType string, metrics map[string]interface{}) {
	fields := []zap.Field{
		zap.String("error_type", errorType),
		zap.Any("metrics", metrics),
	}

	l.LogError(ctx, err, fields...)
}

// LogErrorWithStack logs an error with stack trace
func (l *ErrorLogger) LogErrorWithStack(ctx context.Context, err error, fields ...zap.Field) {
	// Add stack trace field
	fields = append(fields, zap.String("stack_trace", fmt.Sprintf("%+v", err)))

	l.LogError(ctx, err, fields...)
}

// LogErrorWithContext logs an error with additional context
func (l *ErrorLogger) LogErrorWithContext(ctx context.Context, err error, contextData map[string]interface{}, fields ...zap.Field) {
	// Add context data
	fields = append(fields, zap.Any("context_data", contextData))

	l.LogError(ctx, err, fields...)
}

// LogErrorWithUser logs an error with user information
func (l *ErrorLogger) LogErrorWithUser(ctx context.Context, err error, userInfo map[string]interface{}, fields ...zap.Field) {
	// Add user information
	fields = append(fields, zap.Any("user_info", userInfo))

	l.LogError(ctx, err, fields...)
}

// LogErrorWithRequest logs an error with request information
func (l *ErrorLogger) LogErrorWithRequest(ctx context.Context, err error, requestInfo map[string]interface{}, fields ...zap.Field) {
	// Add request information
	fields = append(fields, zap.Any("request_info", requestInfo))

	l.LogError(ctx, err, fields...)
}

// LogErrorWithResponse logs an error with response information
func (l *ErrorLogger) LogErrorWithResponse(ctx context.Context, err error, responseInfo map[string]interface{}, fields ...zap.Field) {
	// Add response information
	fields = append(fields, zap.Any("response_info", responseInfo))

	l.LogError(ctx, err, fields...)
}

// LogErrorWithEnvironment logs an error with environment information
func (l *ErrorLogger) LogErrorWithEnvironment(ctx context.Context, err error, envInfo map[string]interface{}, fields ...zap.Field) {
	// Add environment information
	fields = append(fields, zap.Any("environment_info", envInfo))

	l.LogError(ctx, err, fields...)
}

// LogErrorWithSeverity logs an error with specific severity level
func (l *ErrorLogger) LogErrorWithSeverity(ctx context.Context, err error, severity zapcore.Level, fields ...zap.Field) {
	// Extract correlation ID from context
	correlationID := l.getCorrelationID(ctx)

	// Add correlation ID to fields
	fields = append(fields, zap.String("correlation_id", correlationID))

	// Add timestamp
	fields = append(fields, zap.Time("timestamp", time.Now()))

	// Log with specific severity level
	switch severity {
	case zapcore.DebugLevel:
		l.logger.Debug("API Error", append(fields, zap.Error(err))...)
	case zapcore.InfoLevel:
		l.logger.Info("API Error", append(fields, zap.Error(err))...)
	case zapcore.WarnLevel:
		l.logger.Warn("API Error", append(fields, zap.Error(err))...)
	case zapcore.ErrorLevel:
		l.logger.Error("API Error", append(fields, zap.Error(err))...)
	case zapcore.FatalLevel:
		l.logger.Fatal("API Error", append(fields, zap.Error(err))...)
	default:
		l.logger.Error("API Error", append(fields, zap.Error(err))...)
	}
}

// getCorrelationID extracts correlation ID from context
func (l *ErrorLogger) getCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		return correlationID
	}
	return "unknown"
}

// LogErrorSummary logs a summary of errors for monitoring
func (l *ErrorLogger) LogErrorSummary(ctx context.Context, summary map[string]interface{}) {
	fields := []zap.Field{
		zap.String("log_type", "error_summary"),
		zap.Any("summary", summary),
	}

	l.logger.Info("Error Summary", fields...)
}

// LogErrorTrend logs error trends for analysis
func (l *ErrorLogger) LogErrorTrend(ctx context.Context, trend map[string]interface{}) {
	fields := []zap.Field{
		zap.String("log_type", "error_trend"),
		zap.Any("trend", trend),
	}

	l.logger.Info("Error Trend", fields...)
}
