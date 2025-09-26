package monitoring

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// StructuredLogger provides structured logging with context
type StructuredLogger struct {
	logger *zap.Logger
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(serviceName string) *StructuredLogger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	// Add service name to all logs
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
		"version": "4.0.0",
	}

	logger, err := config.Build()
	if err != nil {
		// Fallback to standard logger
		log.Printf("Failed to create structured logger: %v", err)
		return &StructuredLogger{
			logger: zap.NewNop(),
		}
	}

	return &StructuredLogger{logger: logger}
}

// LogRequest logs an HTTP request with context
func (sl *StructuredLogger) LogRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration, userID string) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("user_id", userID),
		zap.String("request_id", getRequestID(ctx)),
	}

	if statusCode >= 400 {
		sl.logger.Error("HTTP request completed with error", fields...)
	} else {
		sl.logger.Info("HTTP request completed", fields...)
	}
}

// LogError logs an error with context
func (sl *StructuredLogger) LogError(ctx context.Context, err error, message string, fields ...zap.Field) {
	allFields := append(fields,
		zap.Error(err),
		zap.String("request_id", getRequestID(ctx)),
		zap.String("error_type", getErrorType(err)),
	)

	sl.logger.Error(message, allFields...)
}

// LogInfo logs an info message with context
func (sl *StructuredLogger) LogInfo(ctx context.Context, message string, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("request_id", getRequestID(ctx)),
	)

	sl.logger.Info(message, allFields...)
}

// LogWarning logs a warning message with context
func (sl *StructuredLogger) LogWarning(ctx context.Context, message string, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("request_id", getRequestID(ctx)),
	)

	sl.logger.Warn(message, allFields...)
}

// LogPerformance logs performance metrics
func (sl *StructuredLogger) LogPerformance(ctx context.Context, operation string, duration time.Duration, success bool, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.Bool("success", success),
		zap.String("request_id", getRequestID(ctx)),
	)

	if success {
		sl.logger.Info("Performance metric", allFields...)
	} else {
		sl.logger.Warn("Performance issue detected", allFields...)
	}
}

// LogCache logs cache operations
func (sl *StructuredLogger) LogCache(ctx context.Context, operation string, key string, hit bool, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("operation", operation),
		zap.String("key", key),
		zap.Bool("hit", hit),
		zap.String("request_id", getRequestID(ctx)),
	)

	sl.logger.Info("Cache operation", allFields...)
}

// LogDatabase logs database operations
func (sl *StructuredLogger) LogDatabase(ctx context.Context, operation string, table string, duration time.Duration, success bool, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Duration("duration", duration),
		zap.Bool("success", success),
		zap.String("request_id", getRequestID(ctx)),
	)

	if success {
		sl.logger.Info("Database operation", allFields...)
	} else {
		sl.logger.Error("Database operation failed", allFields...)
	}
}

// LogBusinessEvent logs business events
func (sl *StructuredLogger) LogBusinessEvent(ctx context.Context, event string, businessID string, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("event", event),
		zap.String("business_id", businessID),
		zap.String("request_id", getRequestID(ctx)),
	)

	sl.logger.Info("Business event", allFields...)
}

// Close closes the logger
func (sl *StructuredLogger) Close() error {
	return sl.logger.Sync()
}

// Helper functions
func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return "unknown"
}

func getErrorType(err error) string {
	if err == nil {
		return "unknown"
	}
	return fmt.Sprintf("%T", err)
}

// RequestIDMiddleware adds request ID to context
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		ctx := context.WithValue(r.Context(), "request_id", requestID)

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d_%s", time.Now().UnixNano(), os.Getenv("HOSTNAME"))
}
