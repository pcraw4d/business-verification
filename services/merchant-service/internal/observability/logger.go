package observability

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Logger provides structured logging functionality
type Logger struct {
	zapLogger *zap.Logger
}

// ModuleLogger provides module-specific logging functionality
type ModuleLogger struct {
	*Logger
	moduleName string
}

// NewLogger creates a new logger instance
func NewLogger(zapLogger *zap.Logger) *Logger {
	return &Logger{
		zapLogger: zapLogger,
	}
}

// NewModuleLogger creates a new module logger instance
func NewModuleLogger(zapLogger *zap.Logger, moduleName string) *ModuleLogger {
	return &ModuleLogger{
		Logger:     NewLogger(zapLogger),
		moduleName: moduleName,
	}
}

// Module-specific logging methods
func (ml *ModuleLogger) LogModuleConfig(ctx context.Context, operation string, config map[string]interface{}) {
	ml.Info("Module configuration", map[string]interface{}{
		"module":    ml.moduleName,
		"operation": operation,
		"config":    config,
	})
}

func (ml *ModuleLogger) LogModuleHealth(ctx context.Context, healthy bool, status string, details map[string]interface{}) {
	ml.Info("Module health check", map[string]interface{}{
		"module":  ml.moduleName,
		"healthy": healthy,
		"status":  status,
		"details": details,
	})
}

func (ml *ModuleLogger) LogModuleStart(ctx context.Context, metadata map[string]interface{}) {
	ml.Info("Module started", map[string]interface{}{
		"module":   ml.moduleName,
		"metadata": metadata,
	})
}

func (ml *ModuleLogger) LogModuleStop(ctx context.Context, reason string) {
	ml.Info("Module stopped", map[string]interface{}{
		"module": ml.moduleName,
		"reason": reason,
	})
}

func (ml *ModuleLogger) LogModuleError(ctx context.Context, operation string, err error, metadata map[string]interface{}) {
	ml.Error("Module error", map[string]interface{}{
		"module":    ml.moduleName,
		"operation": operation,
		"error":     err.Error(),
		"metadata":  metadata,
	})
}

func (ml *ModuleLogger) LogModulePerformance(ctx context.Context, operation string, startTime, endTime time.Time, metrics map[string]interface{}) {
	duration := endTime.Sub(startTime)
	ml.Info("Module performance", map[string]interface{}{
		"module":     ml.moduleName,
		"operation":  operation,
		"duration":   duration.String(),
		"start_time": startTime,
		"end_time":   endTime,
		"metrics":    metrics,
	})
}

func (ml *ModuleLogger) LogModuleRequest(ctx context.Context, requestID string, statusCode int, duration time.Duration) {
	ml.Info("Module request", map[string]interface{}{
		"module":      ml.moduleName,
		"request_id":  requestID,
		"status_code": statusCode,
		"duration":    duration.String(),
	})
}

func (ml *ModuleLogger) LogModuleResponse(ctx context.Context, requestID string, statusCode int, duration time.Duration, metadata map[string]interface{}) {
	ml.Info("Module response", map[string]interface{}{
		"module":      ml.moduleName,
		"request_id":  requestID,
		"status_code": statusCode,
		"duration":    duration.String(),
		"metadata":    metadata,
	})
}

// GetZapLogger returns the underlying zap logger
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.zapLogger
}

// WithComponent creates a logger with component context
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		zapLogger: l.zapLogger.With(zap.String("component", component)),
	}
}

// Info logs an info level message
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	if l.zapLogger == nil {
		return
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	l.zapLogger.Info(msg, zapFields...)
}

// Warn logs a warning level message
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	if l.zapLogger == nil {
		return
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	l.zapLogger.Warn(msg, zapFields...)
}

// Error logs an error level message
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	if l.zapLogger == nil {
		return
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	l.zapLogger.Error(msg, zapFields...)
}

// Debug logs a debug level message
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	if l.zapLogger == nil {
		return
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	l.zapLogger.Debug(msg, zapFields...)
}

// Fatal logs a fatal level message
func (l *Logger) Fatal(msg string, fields map[string]interface{}) {
	if l.zapLogger == nil {
		return
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	l.zapLogger.Fatal(msg, zapFields...)
}

// Tracer returns a tracer instance
func (l *Logger) Tracer() Tracer {
	return NewTracer()
}

// LogBusinessEvent logs a business event
func (l *Logger) LogBusinessEvent(event string, fields map[string]interface{}) {
	zapFields := make([]zap.Field, 0, len(fields)+1)
	zapFields = append(zapFields, zap.String("event", event))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	l.zapLogger.Info("business_event", zapFields...)
}

// WithError creates a logger with error context
func (l *Logger) WithError(err error) *Logger {
	// In a real implementation, this would add error context to the logger
	return l
}

// LogAPIRequest logs an API request
func (l *Logger) LogAPIRequest(method, path string, statusCode int, duration time.Duration, fields map[string]interface{}) {
	zapFields := make([]zap.Field, 0, len(fields)+4)
	zapFields = append(zapFields, zap.String("method", method))
	zapFields = append(zapFields, zap.String("path", path))
	zapFields = append(zapFields, zap.Int("status_code", statusCode))
	zapFields = append(zapFields, zap.Duration("duration", duration))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	l.zapLogger.Info("api_request", zapFields...)
}
