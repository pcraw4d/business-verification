package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// CorrelationMiddleware provides request correlation and tracking
type CorrelationMiddleware struct {
	logger *zap.Logger
}

// NewCorrelationMiddleware creates a new correlation middleware
func NewCorrelationMiddleware(logger *zap.Logger) *CorrelationMiddleware {
	return &CorrelationMiddleware{
		logger: logger,
	}
}

// CorrelationContext contains correlation information
type CorrelationContext struct {
	CorrelationID string
	RequestID     string
	StartTime     time.Time
	UserAgent     string
	ClientIP      string
	RequestPath   string
	Method        string
}

// generateCorrelationID generates a unique correlation ID
func (cm *CorrelationMiddleware) generateCorrelationID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateRequestID generates a unique request ID
func (cm *CorrelationMiddleware) generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getClientIP extracts the client IP from the request
func (cm *CorrelationMiddleware) getClientIP(r *http.Request) string {
	// Check for forwarded headers first
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return forwardedFor
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to remote address
	return r.RemoteAddr
}

// Middleware adds correlation tracking to requests
func (cm *CorrelationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Generate correlation and request IDs
		correlationID := cm.generateCorrelationID()
		requestID := cm.generateRequestID()

		// Create correlation context
		correlationCtx := &CorrelationContext{
			CorrelationID: correlationID,
			RequestID:     requestID,
			StartTime:     startTime,
			UserAgent:     r.UserAgent(),
			ClientIP:      cm.getClientIP(r),
			RequestPath:   r.URL.Path,
			Method:        r.Method,
		}

		// Add correlation ID to response headers
		w.Header().Set("X-Correlation-ID", correlationID)
		w.Header().Set("X-Request-ID", requestID)

		// Add correlation information to request context
		ctx := context.WithValue(r.Context(), "correlation_id", correlationID)
		ctx = context.WithValue(ctx, "request_id", requestID)
		ctx = context.WithValue(ctx, "correlation_context", correlationCtx)

		// Log request start
		cm.logger.Info("Request started",
			zap.String("correlation_id", correlationID),
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("user_agent", r.UserAgent()),
			zap.String("client_ip", correlationCtx.ClientIP),
			zap.String("query", r.URL.RawQuery),
		)

		// Create a custom response writer to capture status code
		responseWriter := &ResponseWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		// Process the request
		next.ServeHTTP(responseWriter, r.WithContext(ctx))

		// Calculate request duration
		duration := time.Since(startTime)

		// Log request completion
		cm.logger.Info("Request completed",
			zap.String("correlation_id", correlationID),
			zap.String("request_id", requestID),
			zap.Int("status_code", responseWriter.StatusCode),
			zap.Duration("duration", duration),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		// Log slow requests
		if duration > 5*time.Second {
			cm.logger.Warn("Slow request detected",
				zap.String("correlation_id", correlationID),
				zap.String("request_id", requestID),
				zap.Duration("duration", duration),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)
		}
	})
}

// ResponseWriter wraps http.ResponseWriter to capture status code
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// WriteHeader captures the status code
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		return correlationID
	}
	return "unknown"
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return "unknown"
}

// GetCorrelationContext extracts correlation context from context
func GetCorrelationContext(ctx context.Context) *CorrelationContext {
	if correlationCtx, ok := ctx.Value("correlation_context").(*CorrelationContext); ok {
		return correlationCtx
	}
	return nil
}

// LogRequestError logs an error with correlation information
func (cm *CorrelationMiddleware) LogRequestError(ctx context.Context, err error, additionalFields ...zap.Field) {
	correlationID := GetCorrelationID(ctx)
	requestID := GetRequestID(ctx)

	fields := []zap.Field{
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID),
		zap.Error(err),
	}

	// Add additional fields
	fields = append(fields, additionalFields...)

	cm.logger.Error("Request error", fields...)
}

// LogRequestWarning logs a warning with correlation information
func (cm *CorrelationMiddleware) LogRequestWarning(ctx context.Context, message string, additionalFields ...zap.Field) {
	correlationID := GetCorrelationID(ctx)
	requestID := GetRequestID(ctx)

	fields := []zap.Field{
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID),
		zap.String("message", message),
	}

	// Add additional fields
	fields = append(fields, additionalFields...)

	cm.logger.Warn("Request warning", fields...)
}

// LogRequestInfo logs info with correlation information
func (cm *CorrelationMiddleware) LogRequestInfo(ctx context.Context, message string, additionalFields ...zap.Field) {
	correlationID := GetCorrelationID(ctx)
	requestID := GetRequestID(ctx)

	fields := []zap.Field{
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID),
		zap.String("message", message),
	}

	// Add additional fields
	fields = append(fields, additionalFields...)

	cm.logger.Info("Request info", fields...)
}

// LogRequestDebug logs debug information with correlation information
func (cm *CorrelationMiddleware) LogRequestDebug(ctx context.Context, message string, additionalFields ...zap.Field) {
	correlationID := GetCorrelationID(ctx)
	requestID := GetRequestID(ctx)

	fields := []zap.Field{
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID),
		zap.String("message", message),
	}

	// Add additional fields
	fields = append(fields, additionalFields...)

	cm.logger.Debug("Request debug", fields...)
}

// TrackExternalCall tracks external service calls
func (cm *CorrelationMiddleware) TrackExternalCall(ctx context.Context, serviceName, endpoint string, startTime time.Time, err error) {
	correlationID := GetCorrelationID(ctx)
	requestID := GetRequestID(ctx)
	duration := time.Since(startTime)

	fields := []zap.Field{
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID),
		zap.String("service", serviceName),
		zap.String("endpoint", endpoint),
		zap.Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		cm.logger.Error("External service call failed", fields...)
	} else {
		cm.logger.Info("External service call completed", fields...)
	}

	// Log slow external calls
	if duration > 2*time.Second {
		cm.logger.Warn("Slow external service call",
			zap.String("correlation_id", correlationID),
			zap.String("request_id", requestID),
			zap.String("service", serviceName),
			zap.String("endpoint", endpoint),
			zap.Duration("duration", duration),
		)
	}
}

// TrackDatabaseCall tracks database operations
func (cm *CorrelationMiddleware) TrackDatabaseCall(ctx context.Context, operation, table string, startTime time.Time, err error) {
	correlationID := GetCorrelationID(ctx)
	requestID := GetRequestID(ctx)
	duration := time.Since(startTime)

	fields := []zap.Field{
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID),
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		cm.logger.Error("Database operation failed", fields...)
	} else {
		cm.logger.Debug("Database operation completed", fields...)
	}

	// Log slow database operations
	if duration > 1*time.Second {
		cm.logger.Warn("Slow database operation",
			zap.String("correlation_id", correlationID),
			zap.String("request_id", requestID),
			zap.String("operation", operation),
			zap.String("table", table),
			zap.Duration("duration", duration),
		)
	}
}

// TrackCacheCall tracks cache operations
func (cm *CorrelationMiddleware) TrackCacheCall(ctx context.Context, operation, key string, startTime time.Time, err error, hit bool) {
	correlationID := GetCorrelationID(ctx)
	requestID := GetRequestID(ctx)
	duration := time.Since(startTime)

	fields := []zap.Field{
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID),
		zap.String("operation", operation),
		zap.String("key", key),
		zap.Duration("duration", duration),
		zap.Bool("cache_hit", hit),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		cm.logger.Error("Cache operation failed", fields...)
	} else {
		cm.logger.Debug("Cache operation completed", fields...)
	}
}

// TrackBusinessLogic tracks business logic operations
func (cm *CorrelationMiddleware) TrackBusinessLogic(ctx context.Context, operation string, startTime time.Time, err error, additionalFields ...zap.Field) {
	correlationID := GetCorrelationID(ctx)
	requestID := GetRequestID(ctx)
	duration := time.Since(startTime)

	fields := []zap.Field{
		zap.String("correlation_id", correlationID),
		zap.String("request_id", requestID),
		zap.String("operation", operation),
		zap.Duration("duration", duration),
	}

	// Add additional fields
	fields = append(fields, additionalFields...)

	if err != nil {
		fields = append(fields, zap.Error(err))
		cm.logger.Error("Business logic operation failed", fields...)
	} else {
		cm.logger.Debug("Business logic operation completed", fields...)
	}

	// Log slow business logic operations
	if duration > 3*time.Second {
		cm.logger.Warn("Slow business logic operation",
			zap.String("correlation_id", correlationID),
			zap.String("request_id", requestID),
			zap.String("operation", operation),
			zap.Duration("duration", duration),
		)
	}
}
