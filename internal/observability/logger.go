package observability

import (
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
