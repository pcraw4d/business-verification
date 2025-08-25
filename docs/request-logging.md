# Request Logging Implementation

## Overview

The Enhanced Business Intelligence System implements a comprehensive request logging middleware that provides detailed request tracking, performance monitoring, and audit capabilities. The request logging middleware captures structured data about HTTP requests and responses, including timing, headers, bodies, and error conditions.

## Features

### Core Logging Features
- **Structured Logging**: JSON-formatted logs with consistent field structure
- **Request ID Generation**: Unique request IDs for correlation and tracing
- **Performance Timing**: Request duration tracking with slow request detection
- **Body Capture**: Configurable request and response body logging
- **Header Logging**: Complete header capture with sensitive data masking
- **Error Tracking**: Comprehensive error and panic logging
- **Path Filtering**: Include/exclude paths for selective logging
- **Remote Address Detection**: Support for proxy headers (X-Forwarded-For, X-Real-IP)

### Security Features
- **Sensitive Data Masking**: Automatic masking of sensitive headers and body fields
- **Configurable Masking**: Customizable lists of sensitive headers and fields
- **Body Size Limits**: Configurable limits to prevent excessive logging
- **Path Exclusion**: Exclude sensitive endpoints from logging

### Performance Features
- **Efficient Processing**: Minimal overhead with optimized logging
- **Configurable Levels**: Different log levels for different environments
- **Slow Request Detection**: Automatic detection and warning of slow requests
- **Performance Metrics**: Duration tracking in milliseconds

## Configuration

### Basic Configuration

```go
import (
    "time"
    "github.com/your-org/kyb-platform/internal/api/middleware"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Default configuration
config := &middleware.RequestLoggingConfig{
    LogLevel:             zapcore.InfoLevel,
    LogRequestBody:       false,
    MaxRequestBodySize:   1024, // 1KB
    LogResponseBody:      false,
    MaxResponseBodySize:  1024, // 1KB
    LogPerformance:       true,
    SlowRequestThreshold: 1 * time.Second,
    GenerateRequestID:    true,
    RequestIDHeader:      "X-Request-ID",
    MaskSensitiveHeaders: []string{"Authorization", "X-API-Key", "Cookie"},
    MaskSensitiveFields:  []string{"password", "token", "secret", "key"},
    IncludePaths:         []string{},
    ExcludePaths:         []string{"/health", "/metrics", "/favicon.ico"},
    CustomFields:         map[string]string{},
    LogErrors:            true,
    LogPanics:            true,
}
```

### Predefined Configurations

#### Development Configuration
```go
// Verbose logging for development
config := middleware.GetVerboseRequestLoggingConfig()
// Features:
// - Debug level logging
// - Request and response body logging (4KB limit)
// - Performance tracking with 500ms threshold
// - Request ID generation
// - Sensitive data masking
// - Error and panic logging
```

#### Production Configuration
```go
// Optimized for production
config := middleware.GetProductionRequestLoggingConfig()
// Features:
// - Info level logging
// - No body logging (performance focus)
// - Performance tracking with 2s threshold
// - Request ID generation
// - Enhanced sensitive data masking
// - Error and panic logging
```

#### Default Configuration
```go
// Balanced configuration
config := middleware.GetDefaultRequestLoggingConfig()
// Features:
// - Info level logging
// - No body logging by default
// - Performance tracking with 1s threshold
// - Request ID generation
// - Standard sensitive data masking
// - Error and panic logging
```

### Advanced Configuration

```go
config := &middleware.RequestLoggingConfig{
    // Logging Level
    LogLevel: zapcore.DebugLevel,
    
    // Request Body Logging
    LogRequestBody:       true,
    MaxRequestBodySize:   8192, // 8KB
    
    // Response Body Logging
    LogResponseBody:      true,
    MaxResponseBodySize:  8192, // 8KB
    
    // Performance Logging
    LogPerformance:       true,
    SlowRequestThreshold: 500 * time.Millisecond,
    
    // Request ID
    GenerateRequestID: true,
    RequestIDHeader:   "X-Request-ID",
    
    // Sensitive Data Masking
    MaskSensitiveHeaders: []string{
        "Authorization",
        "X-API-Key",
        "Cookie",
        "X-CSRF-Token",
        "X-Secret-Header",
    },
    MaskSensitiveFields: []string{
        "password",
        "token",
        "secret",
        "key",
        "api_key",
        "access_token",
        "refresh_token",
    },
    
    // Path Filtering
    IncludePaths: []string{"/api", "/admin"},
    ExcludePaths: []string{
        "/health",
        "/metrics",
        "/favicon.ico",
        "/robots.txt",
        "/api/health",
        "/admin/metrics",
    },
    
    // Custom Fields
    CustomFields: map[string]string{
        "service": "kyb-platform",
        "version": "1.0.0",
        "environment": "production",
    },
    
    // Error Logging
    LogErrors: true,
    LogPanics: true,
}
```

## Usage

### Basic Usage

```go
package main

import (
    "net/http"
    "github.com/your-org/kyb-platform/internal/api/middleware"
    "go.uber.org/zap"
)

func main() {
    logger := zap.NewProduction()
    
    // Create request logging configuration
    config := middleware.GetDefaultRequestLoggingConfig()
    
    // Create request logging middleware
    requestLoggingMiddleware := middleware.NewRequestLoggingMiddleware(config, logger)
    
    // Create your main handler
    mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, logged API!"))
    })
    
    // Apply request logging middleware
    handler := requestLoggingMiddleware.Middleware(mainHandler)
    
    // Start server
    http.ListenAndServe(":8080", handler)
}
```

### Integration with Existing Middleware

```go
func setupMiddleware(router *http.ServeMux, logger *zap.Logger) {
    // Request logging middleware (should be first)
    requestLoggingConfig := middleware.GetDefaultRequestLoggingConfig()
    requestLoggingMiddleware := middleware.NewRequestLoggingMiddleware(requestLoggingConfig, logger)
    
    // CORS middleware
    corsConfig := middleware.GetDefaultCORSConfig()
    corsMiddleware := middleware.NewCORSMiddleware(corsConfig, logger)
    
    // Security headers middleware
    securityConfig := &middleware.SecurityHeadersConfig{
        CSPEnabled:     true,
        HSTSEnabled:    true,
        FrameOptions:   "DENY",
        XSSProtection:  "1; mode=block",
    }
    securityMiddleware := middleware.NewSecurityHeadersMiddleware(securityConfig, logger)
    
    // Rate limiting middleware
    rateLimitConfig := &middleware.RateLimitConfig{
        RequestsPerMinute: 100,
        BurstSize:         20,
    }
    rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimitConfig, logger)
    
    // Apply middleware chain
    handler := requestLoggingMiddleware.Middleware(
        corsMiddleware.Middleware(
            securityMiddleware.Middleware(
                rateLimitMiddleware.Middleware(router),
            ),
        ),
    )
    
    return handler
}
```

### Accessing Request ID in Handlers

```go
func (h *Handler) ProcessRequest(w http.ResponseWriter, r *http.Request) {
    // Get request ID from context
    requestID := r.Context().Value("request_id").(string)
    
    // Use request ID in your business logic
    h.logger.Info("Processing request",
        zap.String("request_id", requestID),
        zap.String("user_id", getUserID(r)),
    )
    
    // Process request...
    
    // Log completion
    h.logger.Info("Request completed",
        zap.String("request_id", requestID),
        zap.String("status", "success"),
    )
}
```

## Log Output Format

### Standard Request Log

```json
{
  "level": "info",
  "ts": 1640995200.123456,
  "msg": "HTTP request",
  "request_id": "a1b2c3d4e5f678901234567890123456",
  "method": "POST",
  "path": "/api/verify",
  "query": "version=v1",
  "remote_addr": "192.168.1.100",
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
  "status_code": 200,
  "duration": 0.125,
  "duration_ms": 125.0,
  "content_length": 1024,
  "headers": {
    "Accept": "application/json",
    "Authorization": "[MASKED]",
    "Content-Type": "application/json",
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
  },
  "request_body": "{\"name\": \"test\", \"password\": \"[MASKED]\"}",
  "response_body": "{\"status\": \"success\", \"id\": \"ver_123\"}",
  "service": "kyb-platform",
  "environment": "production"
}
```

### Error Request Log

```json
{
  "level": "error",
  "ts": 1640995200.123456,
  "msg": "HTTP request",
  "request_id": "a1b2c3d4e5f678901234567890123456",
  "method": "POST",
  "path": "/api/verify",
  "query": "",
  "remote_addr": "192.168.1.100",
  "user_agent": "curl/7.68.0",
  "status_code": 400,
  "duration": 0.045,
  "duration_ms": 45.0,
  "content_length": 256,
  "headers": {
    "Accept": "*/*",
    "Content-Type": "application/json"
  },
  "request_body": "{\"invalid\": \"data\"}",
  "service": "kyb-platform",
  "environment": "production"
}
```

### Slow Request Log

```json
{
  "level": "warn",
  "ts": 1640995200.123456,
  "msg": "HTTP request",
  "request_id": "a1b2c3d4e5f678901234567890123456",
  "method": "GET",
  "path": "/api/reports",
  "query": "date=2024-01-01",
  "remote_addr": "192.168.1.100",
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
  "status_code": 200,
  "duration": 2.5,
  "duration_ms": 2500.0,
  "content_length": 0,
  "headers": {
    "Accept": "application/json",
    "Authorization": "[MASKED]"
  },
  "service": "kyb-platform",
  "environment": "production"
}
```

### Panic Log

```json
{
  "level": "error",
  "ts": 1640995200.123456,
  "msg": "HTTP request panic",
  "request_id": "a1b2c3d4e5f678901234567890123456",
  "method": "POST",
  "path": "/api/crash",
  "remote_addr": "192.168.1.100",
  "panic": "runtime error: invalid memory address or nil pointer dereference",
  "duration": 0.001,
  "service": "kyb-platform",
  "environment": "production"
}
```

## Sensitive Data Masking

### Header Masking

The middleware automatically masks sensitive headers:

```go
config := &middleware.RequestLoggingConfig{
    MaskSensitiveHeaders: []string{
        "Authorization",
        "X-API-Key",
        "Cookie",
        "X-CSRF-Token",
        "X-Secret-Header",
    },
}
```

**Example:**
- Original: `Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
- Logged: `Authorization: [MASKED]`

### Body Field Masking

The middleware masks sensitive fields in request and response bodies:

```go
config := &middleware.RequestLoggingConfig{
    MaskSensitiveFields: []string{
        "password",
        "token",
        "secret",
        "key",
        "api_key",
        "access_token",
        "refresh_token",
    },
}
```

**Example:**
- Original: `{"username": "john", "password": "secret123", "token": "abc123"}`
- Logged: `{"username": "john", "password": "[MASKED]", "token": "[MASKED]"}`

## Path Filtering

### Include Paths

Only log requests to specific paths:

```go
config := &middleware.RequestLoggingConfig{
    IncludePaths: []string{
        "/api",
        "/admin",
        "/webhook",
    },
}
```

### Exclude Paths

Exclude specific paths from logging:

```go
config := &middleware.RequestLoggingConfig{
    ExcludePaths: []string{
        "/health",
        "/metrics",
        "/favicon.ico",
        "/robots.txt",
        "/api/health",
        "/admin/metrics",
    },
}
```

### Combined Filtering

```go
config := &middleware.RequestLoggingConfig{
    IncludePaths: []string{"/api", "/admin"},
    ExcludePaths: []string{"/api/health", "/admin/metrics"},
}
```

This configuration will:
- Log requests to `/api/users`, `/api/verify`, `/admin/dashboard`
- Skip logging for `/api/health`, `/admin/metrics`, `/other/path`

## Performance Monitoring

### Slow Request Detection

```go
config := &middleware.RequestLoggingConfig{
    LogPerformance:       true,
    SlowRequestThreshold: 1 * time.Second,
}
```

Requests taking longer than the threshold will be logged at `warn` level.

### Performance Metrics

The middleware provides several performance metrics:

- `duration`: Request duration as a time.Duration
- `duration_ms`: Request duration in milliseconds (float64)

### Custom Performance Thresholds

```go
// Development - more sensitive to slow requests
config := &middleware.RequestLoggingConfig{
    SlowRequestThreshold: 500 * time.Millisecond,
}

// Production - higher threshold
config := &middleware.RequestLoggingConfig{
    SlowRequestThreshold: 2 * time.Second,
}
```

## Request ID Generation

### Automatic Generation

```go
config := &middleware.RequestLoggingConfig{
    GenerateRequestID: true,
    RequestIDHeader:   "X-Request-ID",
}
```

The middleware will:
1. Check for existing `X-Request-ID` header
2. Generate a new 32-character hex ID if not present
3. Add the ID to the response headers
4. Include the ID in all log entries

### Custom Request ID Header

```go
config := &middleware.RequestLoggingConfig{
    GenerateRequestID: true,
    RequestIDHeader:   "X-Correlation-ID",
}
```

### Disable Request ID Generation

```go
config := &middleware.RequestLoggingConfig{
    GenerateRequestID: false,
}
```

## Environment-Specific Configurations

### Development Environment

```go
func getDevelopmentRequestLoggingConfig() *middleware.RequestLoggingConfig {
    return &middleware.RequestLoggingConfig{
        LogLevel:             zapcore.DebugLevel,
        LogRequestBody:       true,
        MaxRequestBodySize:   4096, // 4KB
        LogResponseBody:      true,
        MaxResponseBodySize:  4096, // 4KB
        LogPerformance:       true,
        SlowRequestThreshold: 500 * time.Millisecond,
        GenerateRequestID:    true,
        RequestIDHeader:      "X-Request-ID",
        MaskSensitiveHeaders: []string{"Authorization", "X-API-Key", "Cookie"},
        MaskSensitiveFields:  []string{"password", "token", "secret", "key"},
        IncludePaths:         []string{},
        ExcludePaths:         []string{"/health", "/metrics", "/favicon.ico"},
        CustomFields:         map[string]string{"environment": "development"},
        LogErrors:            true,
        LogPanics:            true,
    }
}
```

### Staging Environment

```go
func getStagingRequestLoggingConfig() *middleware.RequestLoggingConfig {
    return &middleware.RequestLoggingConfig{
        LogLevel:             zapcore.InfoLevel,
        LogRequestBody:       true,
        MaxRequestBodySize:   2048, // 2KB
        LogResponseBody:      false,
        MaxResponseBodySize:  0,
        LogPerformance:       true,
        SlowRequestThreshold: 1 * time.Second,
        GenerateRequestID:    true,
        RequestIDHeader:      "X-Request-ID",
        MaskSensitiveHeaders: []string{"Authorization", "X-API-Key", "Cookie", "X-CSRF-Token"},
        MaskSensitiveFields:  []string{"password", "token", "secret", "key", "api_key"},
        IncludePaths:         []string{},
        ExcludePaths:         []string{"/health", "/metrics", "/favicon.ico"},
        CustomFields:         map[string]string{"environment": "staging"},
        LogErrors:            true,
        LogPanics:            true,
    }
}
```

### Production Environment

```go
func getProductionRequestLoggingConfig() *middleware.RequestLoggingConfig {
    return &middleware.RequestLoggingConfig{
        LogLevel:             zapcore.InfoLevel,
        LogRequestBody:       false,
        MaxRequestBodySize:   0,
        LogResponseBody:      false,
        MaxResponseBodySize:  0,
        LogPerformance:       true,
        SlowRequestThreshold: 2 * time.Second,
        GenerateRequestID:    true,
        RequestIDHeader:      "X-Request-ID",
        MaskSensitiveHeaders: []string{
            "Authorization", "X-API-Key", "Cookie", "X-CSRF-Token",
            "X-Secret-Header", "X-Auth-Token",
        },
        MaskSensitiveFields: []string{
            "password", "token", "secret", "key", "api_key",
            "access_token", "refresh_token", "private_key",
        },
        IncludePaths: []string{},
        ExcludePaths: []string{
            "/health", "/metrics", "/favicon.ico", "/robots.txt",
            "/api/health", "/admin/metrics",
        },
        CustomFields: map[string]string{
            "environment": "production",
            "service": "kyb-platform",
            "version": "1.0.0",
        },
        LogErrors: true,
        LogPanics: true,
    }
}
```

## Testing

### Unit Testing

```go
func TestRequestLoggingMiddleware(t *testing.T) {
    // Create observer to capture logs
    core, obs := observer.New(zapcore.InfoLevel)
    logger := zap.New(core)
    
    config := &middleware.RequestLoggingConfig{
        LogLevel:             zapcore.InfoLevel,
        LogRequestBody:       true,
        MaxRequestBodySize:   1024,
        LogResponseBody:      true,
        MaxResponseBodySize:  1024,
        GenerateRequestID:    true,
        RequestIDHeader:      "X-Request-ID",
        MaskSensitiveHeaders: []string{"Authorization"},
        MaskSensitiveFields:  []string{"password"},
    }
    
    middleware := middleware.NewRequestLoggingMiddleware(config, logger)
    
    // Test request
    req := httptest.NewRequest("POST", "/api/test", strings.NewReader(`{"password": "secret123"}`))
    req.Header.Set("Authorization", "Bearer secret-token")
    req.Header.Set("Content-Type", "application/json")
    
    recorder := httptest.NewRecorder()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "success"}`))
    })
    
    middleware.Middleware(handler).ServeHTTP(recorder, req)
    
    // Assertions
    assert.Equal(t, http.StatusOK, recorder.Code)
    
    // Check request ID header
    requestID := recorder.Header().Get("X-Request-ID")
    assert.NotEmpty(t, requestID)
    assert.Len(t, requestID, 32)
    
    // Check logs
    logs := obs.FilterMessage("HTTP request").All()
    assert.Len(t, logs, 1)
    
    log := logs[0]
    logFields := log.ContextMap()
    
    assert.Equal(t, "POST", logFields["method"])
    assert.Equal(t, "/api/test", logFields["path"])
    assert.Equal(t, 200, logFields["status_code"])
    assert.Equal(t, requestID, logFields["request_id"])
    
    // Check sensitive data masking
    headers := logFields["headers"].(map[string]interface{})
    assert.Equal(t, "[MASKED]", headers["Authorization"])
    
    requestBody := logFields["request_body"].(string)
    assert.Contains(t, requestBody, "[MASKED]")
    assert.NotContains(t, requestBody, "secret123")
}
```

### Integration Testing

```go
func TestRequestLoggingIntegration(t *testing.T) {
    // Setup test server with request logging middleware
    core, obs := observer.New(zapcore.InfoLevel)
    logger := zap.New(core)
    
    config := middleware.GetVerboseRequestLoggingConfig()
    requestLoggingMiddleware := middleware.NewRequestLoggingMiddleware(config, logger)
    
    router := http.NewServeMux()
    router.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("success"))
    })
    
    handler := requestLoggingMiddleware.Middleware(router)
    server := httptest.NewServer(handler)
    defer server.Close()
    
    // Test request
    req, _ := http.NewRequest("GET", server.URL+"/api/test", nil)
    req.Header.Set("User-Agent", "test-agent")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Assertions
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // Check request ID header
    requestID := resp.Header.Get("X-Request-ID")
    assert.NotEmpty(t, requestID)
    
    // Check logs
    logs := obs.FilterMessage("HTTP request").All()
    assert.Len(t, logs, 1)
    
    log := logs[0]
    logFields := log.ContextMap()
    
    assert.Equal(t, "GET", logFields["method"])
    assert.Equal(t, "/api/test", logFields["path"])
    assert.Equal(t, requestID, logFields["request_id"])
    assert.Equal(t, "test-agent", logFields["user_agent"])
}
```

## Troubleshooting

### Common Issues

1. **Missing Request ID**
   ```
   Request ID not found in context
   ```
   
   **Solution**: Ensure request logging middleware is applied before other middleware:
   ```go
   handler := requestLoggingMiddleware.Middleware(
       otherMiddleware.Middleware(router),
   )
   ```

2. **Sensitive Data Not Masked**
   ```
   Password appears in logs
   ```
   
   **Solution**: Add field to `MaskSensitiveFields`:
   ```go
   MaskSensitiveFields: []string{"password", "secret", "token"},
   ```

3. **Performance Issues**
   ```
   Logging is too slow
   ```
   
   **Solution**: Disable body logging in production:
   ```go
   LogRequestBody:  false,
   LogResponseBody: false,
   ```

4. **Too Much Logging**
   ```
   Too many log entries
   ```
   
   **Solution**: Use path filtering:
   ```go
   IncludePaths: []string{"/api"},
   ExcludePaths: []string{"/health", "/metrics"},
   ```

### Debug Mode

Enable debug logging for troubleshooting:

```go
config := &middleware.RequestLoggingConfig{
    LogLevel: zapcore.DebugLevel,
    // ... other config
}
```

### Performance Monitoring

Monitor request logging performance:

```go
// Check log volume
logger.Info("Request logging stats",
    zap.Int("requests_per_minute", requestsPerMinute),
    zap.Duration("average_duration", averageDuration),
    zap.Int("slow_requests", slowRequestCount),
)
```

## Best Practices

### Security Best Practices

1. **Always Mask Sensitive Data**
   ```go
   MaskSensitiveHeaders: []string{"Authorization", "X-API-Key", "Cookie"},
   MaskSensitiveFields:  []string{"password", "token", "secret"},
   ```

2. **Limit Body Logging in Production**
   ```go
   LogRequestBody:  false,
   LogResponseBody: false,
   ```

3. **Exclude Sensitive Endpoints**
   ```go
   ExcludePaths: []string{"/admin", "/internal", "/debug"},
   ```

### Performance Best Practices

1. **Use Appropriate Log Levels**
   ```go
   // Development
   LogLevel: zapcore.DebugLevel,
   
   // Production
   LogLevel: zapcore.InfoLevel,
   ```

2. **Set Reasonable Body Size Limits**
   ```go
   MaxRequestBodySize:  1024, // 1KB
   MaxResponseBodySize: 1024, // 1KB
   ```

3. **Configure Slow Request Thresholds**
   ```go
   // Development
   SlowRequestThreshold: 500 * time.Millisecond,
   
   // Production
   SlowRequestThreshold: 2 * time.Second,
   ```

### Monitoring Best Practices

1. **Use Request IDs for Correlation**
   ```go
   // Extract request ID from context
   requestID := r.Context().Value("request_id").(string)
   
   // Use in business logic logs
   logger.Info("Processing verification",
       zap.String("request_id", requestID),
       zap.String("user_id", userID),
   )
   ```

2. **Monitor Slow Requests**
   ```go
   // Set up alerts for slow requests
   if duration > 5*time.Second {
       // Send alert
   }
   ```

3. **Track Error Rates**
   ```go
   // Monitor 4xx and 5xx status codes
   if statusCode >= 400 {
       // Increment error counter
   }
   ```

## Integration with Observability

### Prometheus Metrics

```go
// Custom metrics for request logging
var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "Duration of HTTP requests",
        },
        []string{"method", "path", "status_code"},
    )
    
    requestTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status_code"},
    )
)
```

### Distributed Tracing

```go
// Extract trace context from request
func extractTraceContext(r *http.Request) (traceID, spanID string) {
    traceID = r.Header.Get("X-Trace-ID")
    spanID = r.Header.Get("X-Span-ID")
    return
}

// Add trace information to logs
logger.Info("HTTP request",
    zap.String("request_id", requestID),
    zap.String("trace_id", traceID),
    zap.String("span_id", spanID),
    // ... other fields
)
```

### Log Aggregation

```go
// Structured logging for log aggregation systems
logger.Info("HTTP request",
    zap.String("request_id", requestID),
    zap.String("method", r.Method),
    zap.String("path", r.URL.Path),
    zap.Int("status_code", statusCode),
    zap.Duration("duration", duration),
    zap.String("remote_addr", remoteAddr),
    zap.String("user_agent", r.UserAgent()),
    zap.Any("headers", headers),
    zap.String("service", "kyb-platform"),
    zap.String("environment", "production"),
)
```

## Conclusion

The request logging implementation provides comprehensive request tracking and monitoring capabilities for the Enhanced Business Intelligence System. With configurable logging levels, sensitive data masking, performance monitoring, and flexible path filtering, it supports various deployment scenarios while maintaining security and performance requirements.

Key benefits:
- **Complete Request Visibility**: Full request and response tracking
- **Security-First Design**: Automatic sensitive data masking
- **Performance Monitoring**: Built-in slow request detection
- **Flexible Configuration**: Environment-specific configurations
- **Request Correlation**: Unique request IDs for tracing
- **Error Tracking**: Comprehensive error and panic logging
- **Integration Ready**: Seamless integration with observability systems

For production deployments, always use appropriate log levels, disable body logging for performance, and ensure sensitive data is properly masked.
