package observability

import (
	"time"

	"go.uber.org/zap"
)

// Logger provides structured logging functionality
type Logger struct {
	zapLogger *zap.Logger
}

// NewLogger creates a new logger instance
func NewLogger(zapLogger *zap.Logger) *Logger {
	return &Logger{
		zapLogger: zapLogger,
	}
}

// WithComponent creates a logger with component context
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		zapLogger: l.zapLogger.With(zap.String("component", component)),
	}
}

// Info logs an info level message
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	l.zapLogger.Info(msg, zapFields...)
}

// Warn logs a warning level message
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	l.zapLogger.Warn(msg, zapFields...)
}

// Error logs an error level message
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	l.zapLogger.Error(msg, zapFields...)
}

// Debug logs a debug level message
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	l.zapLogger.Debug(msg, zapFields...)
}

// Fatal logs a fatal level message
func (l *Logger) Fatal(msg string, fields map[string]interface{}) {
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
