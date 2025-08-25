package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestCorrelationMiddleware_GenerateIDs(t *testing.T) {
	logger := zap.NewNop()
	middleware := NewCorrelationMiddleware(logger)

	// Test correlation ID generation
	correlationID1 := middleware.generateCorrelationID()
	correlationID2 := middleware.generateCorrelationID()

	assert.NotEmpty(t, correlationID1)
	assert.NotEmpty(t, correlationID2)
	assert.NotEqual(t, correlationID1, correlationID2)
	assert.Len(t, correlationID1, 32) // 16 bytes = 32 hex characters

	// Test request ID generation
	requestID1 := middleware.generateRequestID()
	requestID2 := middleware.generateRequestID()

	assert.NotEmpty(t, requestID1)
	assert.NotEmpty(t, requestID2)
	assert.NotEqual(t, requestID1, requestID2)
	assert.Len(t, requestID1, 16) // 8 bytes = 16 hex characters
}

func TestCorrelationMiddleware_GetClientIP(t *testing.T) {
	logger := zap.NewNop()
	middleware := NewCorrelationMiddleware(logger)

	// Test with X-Forwarded-For header
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "192.168.1.100")
	ip1 := middleware.getClientIP(req1)
	assert.Equal(t, "192.168.1.100", ip1)

	// Test with X-Real-IP header
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Real-IP", "10.0.0.50")
	ip2 := middleware.getClientIP(req2)
	assert.Equal(t, "10.0.0.50", ip2)

	// Test with remote address
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "172.16.0.1:12345"
	ip3 := middleware.getClientIP(req3)
	assert.Equal(t, "172.16.0.1:12345", ip3)
}

func TestCorrelationMiddleware_Middleware(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	middleware := NewCorrelationMiddleware(logger)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correlation ID is in context
		correlationID := GetCorrelationID(r.Context())
		requestID := GetRequestID(r.Context())
		correlationCtx := GetCorrelationContext(r.Context())

		assert.NotEmpty(t, correlationID)
		assert.NotEmpty(t, requestID)
		assert.NotNil(t, correlationCtx)
		assert.Equal(t, correlationID, correlationCtx.CorrelationID)
		assert.Equal(t, requestID, correlationCtx.RequestID)
		assert.Equal(t, "GET", correlationCtx.Method)
		assert.Equal(t, "/test", correlationCtx.RequestPath)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "192.168.1.100")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Apply middleware
	middleware.Middleware(testHandler).ServeHTTP(rr, req)

	// Verify response headers
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("X-Correlation-ID"))
	assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
	assert.Equal(t, "test response", rr.Body.String())

	// Verify logs
	assert.Equal(t, 2, logs.Len()) // Request started and completed

	// Check request started log
	startLog := logs.All()[0]
	assert.Equal(t, "Request started", startLog.Message)
	assert.Equal(t, zap.InfoLevel, startLog.Level)

	contextMap := startLog.ContextMap()
	assert.Contains(t, contextMap, "correlation_id")
	assert.Contains(t, contextMap, "request_id")
	assert.Equal(t, "GET", contextMap["method"])
	assert.Equal(t, "/test", contextMap["path"])
	assert.Equal(t, "test-agent", contextMap["user_agent"])
	assert.Equal(t, "192.168.1.100", contextMap["client_ip"])

	// Check request completed log
	completeLog := logs.All()[1]
	assert.Equal(t, "Request completed", completeLog.Message)
	assert.Equal(t, zap.InfoLevel, completeLog.Level)

	contextMap = completeLog.ContextMap()
	assert.Contains(t, contextMap, "correlation_id")
	assert.Contains(t, contextMap, "request_id")
	assert.Equal(t, int64(200), contextMap["status_code"])
	assert.Contains(t, contextMap, "duration")
}

func TestCorrelationMiddleware_ContextHelpers(t *testing.T) {
	// Test with correlation ID in context
	ctx1 := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	correlationID := GetCorrelationID(ctx1)
	assert.Equal(t, "test-correlation-123", correlationID)

	// Test without correlation ID in context
	ctx2 := context.Background()
	correlationID2 := GetCorrelationID(ctx2)
	assert.Equal(t, "unknown", correlationID2)

	// Test with request ID in context
	ctx3 := context.WithValue(context.Background(), "request_id", "test-request-456")
	requestID := GetRequestID(ctx3)
	assert.Equal(t, "test-request-456", requestID)

	// Test without request ID in context
	requestID2 := GetRequestID(ctx2)
	assert.Equal(t, "unknown", requestID2)

	// Test with correlation context in context
	correlationCtx := &CorrelationContext{
		CorrelationID: "test-correlation-789",
		RequestID:     "test-request-789",
		StartTime:     time.Now(),
		UserAgent:     "test-agent",
		ClientIP:      "192.168.1.1",
		RequestPath:   "/test",
		Method:        "GET",
	}
	ctx4 := context.WithValue(context.Background(), "correlation_context", correlationCtx)
	retrievedCtx := GetCorrelationContext(ctx4)
	assert.Equal(t, correlationCtx, retrievedCtx)

	// Test without correlation context in context
	retrievedCtx2 := GetCorrelationContext(ctx2)
	assert.Nil(t, retrievedCtx2)
}

func TestCorrelationMiddleware_LoggingMethods(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	middleware := NewCorrelationMiddleware(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	ctx = context.WithValue(ctx, "request_id", "test-request-456")

	// Test LogRequestError
	testErr := &ValidationError{
		Code:    ErrorCodeMissingRequiredField,
		Message: "Test error",
	}
	middleware.LogRequestError(ctx, testErr, zap.String("test_field", "test_value"))

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	assert.Equal(t, "Request error", logEntry.Message)
	assert.Equal(t, zap.ErrorLevel, logEntry.Level)

	contextMap := logEntry.ContextMap()
	assert.Equal(t, "test-correlation-123", contextMap["correlation_id"])
	assert.Equal(t, "test-request-456", contextMap["request_id"])
	assert.Equal(t, "test_value", contextMap["test_field"])

	// Clear logs
	logs.TakeAll()

	// Test LogRequestWarning
	middleware.LogRequestWarning(ctx, "Test warning", zap.String("warning_field", "warning_value"))

	assert.Equal(t, 1, logs.Len())
	logEntry = logs.All()[0]
	assert.Equal(t, "Request warning", logEntry.Message)
	assert.Equal(t, zap.WarnLevel, logEntry.Level)

	contextMap = logEntry.ContextMap()
	assert.Equal(t, "test-correlation-123", contextMap["correlation_id"])
	assert.Equal(t, "test-request-456", contextMap["request_id"])
	assert.Equal(t, "Test warning", contextMap["message"])
	assert.Equal(t, "warning_value", contextMap["warning_field"])

	// Clear logs
	logs.TakeAll()

	// Test LogRequestInfo
	middleware.LogRequestInfo(ctx, "Test info", zap.String("info_field", "info_value"))

	assert.Equal(t, 1, logs.Len())
	logEntry = logs.All()[0]
	assert.Equal(t, "Request info", logEntry.Message)
	assert.Equal(t, zap.InfoLevel, logEntry.Level)

	contextMap = logEntry.ContextMap()
	assert.Equal(t, "test-correlation-123", contextMap["correlation_id"])
	assert.Equal(t, "test-request-456", contextMap["request_id"])
	assert.Equal(t, "Test info", contextMap["message"])
	assert.Equal(t, "info_value", contextMap["info_field"])
}

func TestCorrelationMiddleware_TrackingMethods(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	middleware := NewCorrelationMiddleware(logger)

	ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")
	ctx = context.WithValue(ctx, "request_id", "test-request-456")

	// Test TrackExternalCall - success
	startTime := time.Now()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure duration > 0
	middleware.TrackExternalCall(ctx, "test-service", "/api/test", startTime, nil)

	assert.Equal(t, 1, logs.Len())
	logEntry := logs.All()[0]
	assert.Equal(t, "External service call completed", logEntry.Message)
	assert.Equal(t, zap.InfoLevel, logEntry.Level)

	contextMap := logEntry.ContextMap()
	assert.Equal(t, "test-correlation-123", contextMap["correlation_id"])
	assert.Equal(t, "test-request-456", contextMap["request_id"])
	assert.Equal(t, "test-service", contextMap["service"])
	assert.Equal(t, "/api/test", contextMap["endpoint"])
	assert.Contains(t, contextMap, "duration")

	// Clear logs
	logs.TakeAll()

	// Test TrackExternalCall - error
	testErr := &ExternalServiceError{
		Code:    ErrorCodeExternalAPIFailed,
		Service: "test-service",
		Message: "Service unavailable",
	}
	middleware.TrackExternalCall(ctx, "test-service", "/api/test", startTime, testErr)

	assert.Equal(t, 1, logs.Len())
	logEntry = logs.All()[0]
	assert.Equal(t, "External service call failed", logEntry.Message)
	assert.Equal(t, zap.ErrorLevel, logEntry.Level)

	contextMap = logEntry.ContextMap()
	assert.Equal(t, "test-correlation-123", contextMap["correlation_id"])
	assert.Equal(t, "test-request-456", contextMap["request_id"])
	assert.Equal(t, "test-service", contextMap["service"])
	assert.Equal(t, "/api/test", contextMap["endpoint"])
	assert.Contains(t, contextMap, "duration")

	// Clear logs
	logs.TakeAll()

	// Test TrackDatabaseCall - success
	middleware.TrackDatabaseCall(ctx, "SELECT", "businesses", startTime, nil)

	assert.Equal(t, 1, logs.Len())
	logEntry = logs.All()[0]
	assert.Equal(t, "Database operation completed", logEntry.Message)
	assert.Equal(t, zap.DebugLevel, logEntry.Level)

	contextMap = logEntry.ContextMap()
	assert.Equal(t, "test-correlation-123", contextMap["correlation_id"])
	assert.Equal(t, "test-request-456", contextMap["request_id"])
	assert.Equal(t, "SELECT", contextMap["operation"])
	assert.Equal(t, "businesses", contextMap["table"])
	assert.Contains(t, contextMap, "duration")

	// Clear logs
	logs.TakeAll()

	// Test TrackDatabaseCall - error
	dbErr := &DatabaseError{
		Code:      ErrorCodeDatabaseConnection,
		Operation: "SELECT",
		Message:   "Connection failed",
	}
	middleware.TrackDatabaseCall(ctx, "SELECT", "businesses", startTime, dbErr)

	assert.Equal(t, 1, logs.Len())
	logEntry = logs.All()[0]
	assert.Equal(t, "Database operation failed", logEntry.Message)
	assert.Equal(t, zap.ErrorLevel, logEntry.Level)

	contextMap = logEntry.ContextMap()
	assert.Equal(t, "test-correlation-123", contextMap["correlation_id"])
	assert.Equal(t, "test-request-456", contextMap["request_id"])
	assert.Equal(t, "SELECT", contextMap["operation"])
	assert.Equal(t, "businesses", contextMap["table"])
	assert.Contains(t, contextMap, "duration")

	// Clear logs
	logs.TakeAll()

	// Test TrackCacheCall - hit
	middleware.TrackCacheCall(ctx, "GET", "cache-key", startTime, nil, true)

	assert.Equal(t, 1, logs.Len())
	logEntry = logs.All()[0]
	assert.Equal(t, "Cache operation completed", logEntry.Message)
	assert.Equal(t, zap.DebugLevel, logEntry.Level)

	contextMap = logEntry.ContextMap()
	assert.Equal(t, "test-correlation-123", contextMap["correlation_id"])
	assert.Equal(t, "test-request-456", contextMap["request_id"])
	assert.Equal(t, "GET", contextMap["operation"])
	assert.Equal(t, "cache-key", contextMap["key"])
	assert.Equal(t, true, contextMap["cache_hit"])
	assert.Contains(t, contextMap, "duration")

	// Clear logs
	logs.TakeAll()

	// Test TrackCacheCall - miss
	middleware.TrackCacheCall(ctx, "GET", "cache-key", startTime, nil, false)

	assert.Equal(t, 1, logs.Len())
	logEntry = logs.All()[0]
	assert.Equal(t, "Cache operation completed", logEntry.Message)
	assert.Equal(t, zap.DebugLevel, logEntry.Level)

	contextMap = logEntry.ContextMap()
	assert.Equal(t, false, contextMap["cache_hit"])

	// Clear logs
	logs.TakeAll()

	// Test TrackBusinessLogic - success
	middleware.TrackBusinessLogic(ctx, "business_classification", startTime, nil, zap.String("business_name", "Test Corp"))

	assert.Equal(t, 1, logs.Len())
	logEntry = logs.All()[0]
	assert.Equal(t, "Business logic operation completed", logEntry.Message)
	assert.Equal(t, zap.DebugLevel, logEntry.Level)

	contextMap = logEntry.ContextMap()
	assert.Equal(t, "test-correlation-123", contextMap["correlation_id"])
	assert.Equal(t, "test-request-456", contextMap["request_id"])
	assert.Equal(t, "business_classification", contextMap["operation"])
	assert.Equal(t, "Test Corp", contextMap["business_name"])
	assert.Contains(t, contextMap, "duration")

	// Clear logs
	logs.TakeAll()

	// Test TrackBusinessLogic - error
	businessErr := &ClassificationError{
		Code:    ErrorCodeClassificationFailed,
		Message: "Classification failed",
	}
	middleware.TrackBusinessLogic(ctx, "business_classification", startTime, businessErr, zap.String("business_name", "Test Corp"))

	assert.Equal(t, 1, logs.Len())
	logEntry = logs.All()[0]
	assert.Equal(t, "Business logic operation failed", logEntry.Message)
	assert.Equal(t, zap.ErrorLevel, logEntry.Level)

	contextMap = logEntry.ContextMap()
	assert.Equal(t, "test-correlation-123", contextMap["correlation_id"])
	assert.Equal(t, "test-request-456", contextMap["request_id"])
	assert.Equal(t, "business_classification", contextMap["operation"])
	assert.Equal(t, "Test Corp", contextMap["business_name"])
	assert.Contains(t, contextMap, "duration")
}

func TestCorrelationMiddleware_SlowRequestDetection(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	middleware := NewCorrelationMiddleware(logger)

	// Create a test handler that takes time
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(6 * time.Second) // Longer than 5 second threshold
		w.WriteHeader(http.StatusOK)
	})

	// Create test request
	req := httptest.NewRequest("GET", "/slow-test", nil)
	rr := httptest.NewRecorder()

	// Apply middleware
	middleware.Middleware(testHandler).ServeHTTP(rr, req)

	// Verify slow request warning was logged
	assert.GreaterOrEqual(t, logs.Len(), 3) // Request started, slow request warning, request completed

	// Find slow request warning
	var slowRequestLog *observer.LoggedEntry
	for _, log := range logs.All() {
		if log.Message == "Slow request detected" {
			slowRequestLog = &log
			break
		}
	}

	assert.NotNil(t, slowRequestLog)
	assert.Equal(t, zap.WarnLevel, slowRequestLog.Level)

	contextMap := slowRequestLog.ContextMap()
	assert.Contains(t, contextMap, "correlation_id")
	assert.Contains(t, contextMap, "request_id")
	assert.Contains(t, contextMap, "duration")
	assert.Equal(t, "GET", contextMap["method"])
	assert.Equal(t, "/slow-test", contextMap["path"])
}

func TestResponseWriter(t *testing.T) {
	// Create a test response writer
	rr := httptest.NewRecorder()
	responseWriter := &ResponseWriter{
		ResponseWriter: rr,
		StatusCode:     http.StatusOK,
	}

	// Test WriteHeader
	responseWriter.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, responseWriter.StatusCode)
	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Test Write
	responseWriter.Write([]byte("test response"))
	assert.Equal(t, "test response", rr.Body.String())
}
