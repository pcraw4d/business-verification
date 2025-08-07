package observability

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	config *config.ObservabilityConfig
}

// NewLogger creates a new logger with the given configuration
func NewLogger(cfg *config.ObservabilityConfig) *Logger {
	var handler slog.Handler

	// Create the appropriate handler based on log format
	switch cfg.LogFormat {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLogLevel(cfg.LogLevel),
		})
	case "text":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLogLevel(cfg.LogLevel),
		})
	default:
		// Default to JSON format
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLogLevel(cfg.LogLevel),
		})
	}

	logger := slog.New(handler)
	return &Logger{
		Logger: logger,
		config: cfg,
	}
}

// getLogLevel converts string log level to slog.Level
func getLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext adds context information to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract request ID from context if available
	if requestID := getRequestID(ctx); requestID != "" {
		return &Logger{
			Logger: l.Logger.With("request_id", requestID),
			config: l.config,
		}
	}
	return l
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	attrs := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		attrs = append(attrs, key, value)
	}
	return &Logger{
		Logger: l.Logger.With(attrs...),
		config: l.config,
	}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err.Error()),
		config: l.config,
	}
}

// WithUser adds user information to the logger
func (l *Logger) WithUser(userID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("user_id", userID),
		config: l.config,
	}
}

// WithComponent adds component information to the logger
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.Logger.With("component", component),
		config: l.config,
	}
}

// WithOperation adds operation information to the logger
func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{
		Logger: l.Logger.With("operation", operation),
		config: l.config,
	}
}

// WithDuration adds duration information to the logger
func (l *Logger) WithDuration(duration time.Duration) *Logger {
	return &Logger{
		Logger: l.Logger.With("duration_ms", duration.Milliseconds()),
		config: l.config,
	}
}

// WithRequest adds HTTP request information to the logger
func (l *Logger) WithRequest(method, path, userAgent string, statusCode int) *Logger {
	return &Logger{
		Logger: l.Logger.With(
			"method", method,
			"path", path,
			"user_agent", userAgent,
			"status_code", statusCode,
		),
		config: l.config,
	}
}

// WithDatabase adds database operation information to the logger
func (l *Logger) WithDatabase(operation, table string, rowsAffected int64) *Logger {
	return &Logger{
		Logger: l.Logger.With(
			"db_operation", operation,
			"db_table", table,
			"rows_affected", rowsAffected,
		),
		config: l.config,
	}
}

// WithExternalService adds external service information to the logger
func (l *Logger) WithExternalService(service, endpoint string, statusCode int) *Logger {
	return &Logger{
		Logger: l.Logger.With(
			"external_service", service,
			"endpoint", endpoint,
			"status_code", statusCode,
		),
		config: l.config,
	}
}

// LogAPIRequest logs an API request with timing information
func (l *Logger) LogAPIRequest(ctx context.Context, method, path, userAgent string, statusCode int, duration time.Duration) {
	l.WithContext(ctx).
		WithRequest(method, path, userAgent, statusCode).
		WithDuration(duration).
		Info("API request completed")
}

// LogDatabaseOperation logs a database operation
func (l *Logger) LogDatabaseOperation(ctx context.Context, operation, table string, rowsAffected int64, duration time.Duration, err error) {
	logger := l.WithContext(ctx).
		WithDatabase(operation, table, rowsAffected).
		WithDuration(duration)

	if err != nil {
		logger.WithError(err).Error("Database operation failed")
	} else {
		logger.Info("Database operation completed")
	}
}

// LogExternalServiceCall logs an external service call
func (l *Logger) LogExternalServiceCall(ctx context.Context, service, endpoint string, statusCode int, duration time.Duration, err error) {
	logger := l.WithContext(ctx).
		WithExternalService(service, endpoint, statusCode).
		WithDuration(duration)

	if err != nil {
		logger.WithError(err).Error("External service call failed")
	} else {
		logger.Info("External service call completed")
	}
}

// LogBusinessEvent logs a business event
func (l *Logger) LogBusinessEvent(ctx context.Context, eventType, eventID string, details map[string]interface{}) {
	logger := l.WithContext(ctx).
		WithFields(map[string]interface{}{
			"event_type": eventType,
			"event_id":   eventID,
		})

	if len(details) > 0 {
		logger = logger.WithFields(details)
	}

	logger.Info("Business event occurred")
}

// LogSecurityEvent logs a security event
func (l *Logger) LogSecurityEvent(ctx context.Context, eventType, userID, ipAddress string, details map[string]interface{}) {
	logger := l.WithContext(ctx).
		WithUser(userID).
		WithFields(map[string]interface{}{
			"event_type": eventType,
			"ip_address": ipAddress,
		})

	if len(details) > 0 {
		logger = logger.WithFields(details)
	}

	logger.Warn("Security event detected")
}

// LogPerformance logs performance metrics
func (l *Logger) LogPerformance(ctx context.Context, metric string, value float64, unit string) {
	l.WithContext(ctx).
		WithFields(map[string]interface{}{
			"metric": metric,
			"value":  value,
			"unit":   unit,
		}).
		Info("Performance metric recorded")
}

// LogStartup logs application startup information
func (l *Logger) LogStartup(version, commitHash, buildTime string) {
	l.WithFields(map[string]interface{}{
		"version":     version,
		"commit_hash": commitHash,
		"build_time":  buildTime,
		"pid":         os.Getpid(),
	}).Info("Application starting")
}

// LogShutdown logs application shutdown information
func (l *Logger) LogShutdown(reason string) {
	l.WithFields(map[string]interface{}{
		"reason": reason,
		"pid":    os.Getpid(),
	}).Info("Application shutting down")
}

// LogHealthCheck logs health check results
func (l *Logger) LogHealthCheck(component string, status string, details map[string]interface{}) {
	logger := l.WithComponent(component).
		WithFields(map[string]interface{}{
			"status": status,
		})

	if len(details) > 0 {
		logger = logger.WithFields(details)
	}

	switch status {
	case "healthy":
		logger.Info("Health check passed")
	case "unhealthy":
		logger.Error("Health check failed")
	case "degraded":
		logger.Warn("Health check degraded")
	default:
		logger.Info("Health check completed")
	}
}

// getRequestID extracts request ID from context
func getRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try to get request ID from context
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}

	return ""
}

// SetOutput sets the output destination for the logger
func (l *Logger) SetOutput(w io.Writer) {
	var handler slog.Handler

	switch l.config.LogFormat {
	case "json":
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: getLogLevel(l.config.LogLevel),
		})
	case "text":
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: getLogLevel(l.config.LogLevel),
		})
	default:
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: getLogLevel(l.config.LogLevel),
		})
	}

	l.Logger = slog.New(handler)
}

// String returns a string representation of the logger configuration
func (l *Logger) String() string {
	return fmt.Sprintf("Logger{level=%s, format=%s}", l.config.LogLevel, l.config.LogFormat)
}
