package observability

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
)

func TestNewLogger(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)
	if logger == nil {
		t.Fatal("Expected logger to be created")
	}

	if logger.config != cfg {
		t.Error("Expected logger config to match input config")
	}
}

func TestLoggerWithContext(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	// Test with context containing request ID
	ctx := context.WithValue(context.Background(), RequestIDKey, "test-request-id")
	loggerWithContext := logger.WithContext(ctx)

	if loggerWithContext == nil {
		t.Fatal("Expected logger with context to be created")
	}
}

func TestLoggerWithFields(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	loggerWithFields := logger.WithFields(fields)

	if loggerWithFields == nil {
		t.Fatal("Expected logger with fields to be created")
	}
}

func TestLoggerWithError(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	err := fmt.Errorf("test error")
	loggerWithError := logger.WithError(err)

	if loggerWithError == nil {
		t.Fatal("Expected logger with error to be created")
	}
}

func TestLoggerWithUser(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	userID := "user123"
	loggerWithUser := logger.WithUser(userID)

	if loggerWithUser == nil {
		t.Fatal("Expected logger with user to be created")
	}
}

func TestLoggerWithComponent(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	component := "auth"
	loggerWithComponent := logger.WithComponent(component)

	if loggerWithComponent == nil {
		t.Fatal("Expected logger with component to be created")
	}
}

func TestLoggerWithOperation(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	operation := "login"
	loggerWithOperation := logger.WithOperation(operation)

	if loggerWithOperation == nil {
		t.Fatal("Expected logger with operation to be created")
	}
}

func TestLoggerWithDuration(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	duration := 100 * time.Millisecond
	loggerWithDuration := logger.WithDuration(duration)

	if loggerWithDuration == nil {
		t.Fatal("Expected logger with duration to be created")
	}
}

func TestLoggerWithRequest(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	method := "GET"
	path := "/api/v1/users"
	userAgent := "test-agent"
	statusCode := 200

	loggerWithRequest := logger.WithRequest(method, path, userAgent, statusCode)

	if loggerWithRequest == nil {
		t.Fatal("Expected logger with request to be created")
	}
}

func TestLoggerWithDatabase(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	operation := "SELECT"
	table := "users"
	rowsAffected := int64(10)

	loggerWithDatabase := logger.WithDatabase(operation, table, rowsAffected)

	if loggerWithDatabase == nil {
		t.Fatal("Expected logger with database to be created")
	}
}

func TestLoggerWithExternalService(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	service := "payment-api"
	endpoint := "/payments"
	statusCode := 200

	loggerWithExternalService := logger.WithExternalService(service, endpoint, statusCode)

	if loggerWithExternalService == nil {
		t.Fatal("Expected logger with external service to be created")
	}
}

func TestLoggerLogAPIRequest(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	ctx := context.Background()
	method := "POST"
	path := "/api/v1/businesses"
	userAgent := "test-agent"
	statusCode := 201
	duration := 150 * time.Millisecond

	// This should not panic
	logger.LogAPIRequest(ctx, method, path, userAgent, statusCode, duration)
}

func TestLoggerLogDatabaseOperation(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	ctx := context.Background()
	operation := "INSERT"
	table := "businesses"
	rowsAffected := int64(1)
	duration := 50 * time.Millisecond

	// Test successful operation
	logger.LogDatabaseOperation(ctx, operation, table, rowsAffected, duration, nil)

	// Test failed operation
	err := fmt.Errorf("database connection failed")
	logger.LogDatabaseOperation(ctx, operation, table, rowsAffected, duration, err)
}

func TestLoggerLogExternalServiceCall(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	ctx := context.Background()
	service := "business-data-api"
	endpoint := "/businesses/search"
	statusCode := 200
	duration := 200 * time.Millisecond

	// Test successful call
	logger.LogExternalServiceCall(ctx, service, endpoint, statusCode, duration, nil)

	// Test failed call
	err := fmt.Errorf("service unavailable")
	logger.LogExternalServiceCall(ctx, service, endpoint, 503, duration, err)
}

func TestLoggerLogBusinessEvent(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	ctx := context.Background()
	eventType := "business_classified"
	eventID := "evt-123"
	details := map[string]interface{}{
		"business_name": "Test Corp",
		"confidence":    0.95,
	}

	logger.LogBusinessEvent(ctx, eventType, eventID, details)
}

func TestLoggerLogSecurityEvent(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	ctx := context.Background()
	eventType := "failed_login"
	userID := "user123"
	ipAddress := "192.168.1.1"
	details := map[string]interface{}{
		"attempts": 5,
		"reason":   "invalid_password",
	}

	logger.LogSecurityEvent(ctx, eventType, userID, ipAddress, details)
}

func TestLoggerLogPerformance(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	ctx := context.Background()
	metric := "response_time"
	value := 150.5
	unit := "ms"

	logger.LogPerformance(ctx, metric, value, unit)
}

func TestLoggerLogStartup(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	version := "1.0.0"
	commitHash := "abc123"
	buildTime := "2024-01-01T00:00:00Z"

	logger.LogStartup(version, commitHash, buildTime)
}

func TestLoggerLogShutdown(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	reason := "graceful shutdown"

	logger.LogShutdown(reason)
}

func TestLoggerLogHealthCheck(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	}

	logger := NewLogger(cfg)

	component := "database"
	status := "healthy"
	details := map[string]interface{}{
		"connection_count": 5,
		"response_time_ms": 10,
	}

	logger.LogHealthCheck(component, status, details)
}

func TestLoggerString(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "text",
	}

	logger := NewLogger(cfg)

	expected := "Logger{level=debug, format=text}"
	if logger.String() != expected {
		t.Errorf("Expected %s, got %s", expected, logger.String())
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo}, // default
		{"", slog.LevelInfo},        // default
	}

	for _, tt := range tests {
		result := getLogLevel(tt.input)
		if result != tt.expected {
			t.Errorf("getLogLevel(%s) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}
