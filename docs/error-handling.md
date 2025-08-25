# Error Handling Middleware

## Overview

The Error Handling Middleware provides comprehensive error handling capabilities for the Enhanced Business Intelligence System. It centralizes error processing, provides standardized error responses, implements panic recovery, tracks error metrics, and integrates seamlessly with the existing logging and security infrastructure.

## Features

### Core Features
- **Centralized Error Handling**: Catch and process all errors consistently across the application
- **Custom Error Types**: Define specific error types for different scenarios (validation, authentication, etc.)
- **Standardized Error Responses**: Consistent JSON error response format with proper HTTP status codes
- **Panic Recovery**: Automatic panic recovery with detailed error logging
- **Error Logging**: Integration with structured logging for comprehensive error tracking
- **Error Metrics**: Track error statistics and trends for monitoring and alerting
- **Request ID Integration**: Correlate errors with specific requests for debugging

### Security Features
- **Sensitive Data Masking**: Automatically mask sensitive headers and data in error responses
- **Internal Error Masking**: Hide internal error details in production environments
- **Context Filtering**: Control what error context is exposed to clients
- **Remote Address Detection**: Proper handling of proxy headers for accurate client identification

### Customization Features
- **Configurable Error Codes**: Custom error codes for different error types
- **Custom Error Handlers**: Implement custom error handling logic for specific error types
- **Environment-Specific Configurations**: Predefined configurations for development, verbose, and production
- **Flexible Logging**: Configurable logging levels and detail inclusion

## Error Types

The middleware supports the following predefined error types:

| Error Type | Description | Default Code | Default Status |
|------------|-------------|--------------|----------------|
| `validation_error` | Input validation errors | `INVALID_INPUT` | 400 |
| `authentication_error` | Authentication failures | `UNAUTHORIZED` | 401 |
| `authorization_error` | Authorization failures | `FORBIDDEN` | 403 |
| `not_found_error` | Resource not found | `NOT_FOUND` | 404 |
| `conflict_error` | Resource conflicts | `CONFLICT` | 409 |
| `rate_limit_error` | Rate limit exceeded | `RATE_LIMITED` | 429 |
| `internal_error` | Internal server errors | `INTERNAL_ERROR` | 500 |
| `external_error` | External service errors | `EXTERNAL_ERROR` | 502 |
| `timeout_error` | Request timeouts | `TIMEOUT` | 408 |
| `unavailable_error` | Service unavailable | `SERVICE_UNAVAILABLE` | 503 |

## Error Severity Levels

| Severity | Description | Usage |
|----------|-------------|-------|
| `low` | Minor issues, non-critical | Validation errors, not found errors |
| `medium` | Moderate issues | Authentication, authorization, rate limiting |
| `high` | Serious issues | Internal errors, external service failures |
| `critical` | Critical issues | Panics, system failures |

## Configuration

### Basic Configuration

```go
config := &ErrorHandlingConfig{
    // Error Logging
    LogErrors:     true,
    LogLevel:      zapcore.ErrorLevel,
    IncludeStack:  false,
    
    // Error Response
    IncludeDetails: true,
    IncludeContext: false,
    MaskInternal:   true,
    
    // Recovery
    RecoverPanics: true,
    
    // Error Metrics
    TrackMetrics: true,
    
    // Default Error
    DefaultErrorType:     ErrorTypeInternal,
    DefaultErrorSeverity: ErrorSeverityMedium,
    DefaultStatusCode:    500,
}
```

### Predefined Configurations

#### Default Configuration
```go
config := GetDefaultErrorHandlingConfig()
```
- Logs errors at Error level
- Includes error details but not context
- Masks internal errors
- Tracks metrics
- Recovers from panics

#### Verbose Configuration
```go
config := GetVerboseErrorHandlingConfig()
```
- Logs errors at Debug level
- Includes stack traces and context
- Does not mask internal errors
- Suitable for development and debugging

#### Production Configuration
```go
config := GetProductionErrorHandlingConfig()
```
- Logs errors at Error level
- Excludes error details and context
- Masks internal errors
- Minimal information exposure

## Usage Examples

### Basic Usage

```go
package main

import (
    "net/http"
    "go.uber.org/zap"
    "your-project/internal/api/middleware"
)

func main() {
    logger := zap.NewProduction()
    
    // Create error handling middleware
    errorMiddleware := middleware.NewErrorHandlingMiddleware(nil, logger)
    
    // Create your handlers
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Your handler logic here
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"success": true}`))
    })
    
    // Apply middleware
    finalHandler := errorMiddleware.Middleware(handler)
    
    // Start server
    http.ListenAndServe(":8080", finalHandler)
}
```

### Custom Error Creation

```go
// Create validation error
validationErr := middleware.CreateValidationError(
    "Invalid input data",
    "The provided email format is invalid",
)

// Create authentication error
authErr := middleware.CreateAuthenticationError(
    "Authentication failed",
    "Invalid API key provided",
)

// Create custom error
customErr := middleware.CreateCustomError(
    middleware.ErrorTypeExternal,
    middleware.ErrorSeverityHigh,
    "External service unavailable",
    "Database connection failed",
    "DB_CONNECTION_ERROR",
    http.StatusServiceUnavailable,
)
```

### Error Handler Integration

```go
// Create custom error handler
customHandler := middleware.ErrorHandlerFunc(func(ctx context.Context, err error, req *http.Request) *middleware.APIError {
    // Custom error handling logic
    if strings.Contains(err.Error(), "database") {
        return &middleware.APIError{
            Type:       middleware.ErrorTypeExternal,
            Severity:   middleware.ErrorSeverityHigh,
            Message:    "Database error",
            Details:    "Please try again later",
            Code:       "DB_ERROR",
            Timestamp:  time.Now(),
        }
    }
    return nil // Use default handling
})

// Configure middleware with custom handler
config := middleware.GetDefaultErrorHandlingConfig()
config.CustomHandlers[middleware.ErrorTypeExternal] = customHandler

errorMiddleware := middleware.NewErrorHandlingMiddleware(config, logger)
```

### Request ID Integration

```go
// In your handler, errors will automatically include request ID if available
func myHandler(w http.ResponseWriter, r *http.Request) {
    // Get request ID from context (set by request logging middleware)
    requestID := r.Context().Value("request_id")
    
    // Create error - request ID will be automatically included
    err := middleware.CreateValidationError("Invalid input", "Missing required field")
    
    // Error response will include the request ID for correlation
    http.Error(w, err.Error(), http.StatusBadRequest)
}
```

## Error Response Format

### Standard Error Response

```json
{
  "error": {
    "type": "validation_error",
    "severity": "low",
    "message": "Invalid input data",
    "details": "The provided email format is invalid",
    "code": "INVALID_INPUT",
    "request_id": "abc123def456",
    "timestamp": "2024-12-19T10:30:00Z",
    "path": "/api/users",
    "method": "POST",
    "user_agent": "Mozilla/5.0...",
    "remote_addr": "192.168.1.100",
    "context": {
      "url": "/api/users?validate=true",
      "headers": {
        "Content-Type": "application/json",
        "Authorization": "[MASKED]"
      }
    }
  },
  "success": false
}
```

### Response Headers

The middleware sets the following headers on error responses:

- `Content-Type: application/json`
- `X-Error-Type: validation_error`
- `X-Error-Code: INVALID_INPUT`
- `X-Request-ID: abc123def456` (if available)

## Integration with Other Middleware

### Request Logging Integration

The error handling middleware integrates seamlessly with the request logging middleware:

```go
// Create middleware stack
requestLoggingMiddleware := middleware.NewRequestLoggingMiddleware(
    middleware.GetDefaultRequestLoggingConfig(),
    logger,
)

errorHandlingMiddleware := middleware.NewErrorHandlingMiddleware(
    middleware.GetDefaultErrorHandlingConfig(),
    logger,
)

// Apply middleware in order
handler := errorHandlingMiddleware.Middleware(
    requestLoggingMiddleware.Middleware(
        yourHandler,
    ),
)
```

### Security Middleware Integration

```go
// Complete middleware stack
securityMiddleware := middleware.NewSecurityHeadersMiddleware(
    middleware.GetDefaultSecurityHeadersConfig(),
)

corsMiddleware := middleware.NewCORSMiddleware(
    middleware.GetDefaultCORSConfig(),
)

rateLimitMiddleware := middleware.NewRateLimitMiddleware(
    middleware.GetDefaultRateLimitConfig(),
    redisClient,
)

// Apply all middleware
finalHandler := errorHandlingMiddleware.Middleware(
    requestLoggingMiddleware.Middleware(
        securityMiddleware.Middleware(
            corsMiddleware.Middleware(
                rateLimitMiddleware.Middleware(
                    yourHandler,
                ),
            ),
        ),
    ),
)
```

## Error Metrics

The middleware tracks comprehensive error metrics:

```go
// Get error metrics
metrics := errorMiddleware.GetErrorMetrics()

fmt.Printf("Total Errors: %d\n", metrics.TotalErrors)
fmt.Printf("Errors by Type: %+v\n", metrics.ErrorsByType)
fmt.Printf("Errors by Severity: %+v\n", metrics.ErrorsBySeverity)
fmt.Printf("Last Error: %s\n", metrics.LastError)
```

### Metrics Integration

```go
// Export metrics to Prometheus
func exportErrorMetrics(middleware *middleware.ErrorHandlingMiddleware) {
    metrics := middleware.GetErrorMetrics()
    
    // Export to your metrics system
    prometheus.NewCounter(prometheus.CounterOpts{
        Name: "api_errors_total",
        Help: "Total number of API errors",
    }).Add(float64(metrics.TotalErrors))
}
```

## Testing

### Unit Testing

```go
func TestErrorHandling(t *testing.T) {
    // Create test logger
    core, obs := observer.New(zapcore.InfoLevel)
    logger := zap.New(core)
    
    // Create middleware
    middleware := middleware.NewErrorHandlingMiddleware(nil, logger)
    
    // Test handler that returns error
    handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
    }))
    
    // Make request
    req := httptest.NewRequest("GET", "/test", nil)
    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)
    
    // Assertions
    assert.Equal(t, http.StatusBadRequest, rr.Code)
    
    var response middleware.ErrorResponse
    json.Unmarshal(rr.Body.Bytes(), &response)
    
    assert.False(t, response.Success)
    assert.Equal(t, "Invalid request", response.Error.Message)
    assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}
```

### Integration Testing

```go
func TestErrorHandlingIntegration(t *testing.T) {
    // Test with real server
    server := httptest.NewServer(errorHandlingMiddleware.Middleware(
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            panic("test panic")
        }),
    ))
    defer server.Close()
    
    // Make request
    resp, err := http.Get(server.URL)
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Assertions
    assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
    
    var response middleware.ErrorResponse
    json.NewDecoder(resp.Body).Decode(&response)
    
    assert.False(t, response.Success)
    assert.Equal(t, middleware.ErrorTypeInternal, response.Error.Type)
    assert.Equal(t, middleware.ErrorSeverityCritical, response.Error.Severity)
}
```

## Best Practices

### Error Creation

1. **Use Appropriate Error Types**: Choose the most specific error type for your scenario
2. **Provide Clear Messages**: Write user-friendly error messages
3. **Include Relevant Details**: Add context that helps with debugging
4. **Set Correct Severity**: Use appropriate severity levels for monitoring

```go
// Good: Specific error type with clear message
err := middleware.CreateValidationError(
    "Invalid email format",
    "Email must be a valid format (e.g., user@example.com)",
)

// Good: Include context for debugging
customErr := middleware.CreateCustomError(
    middleware.ErrorTypeExternal,
    middleware.ErrorSeverityHigh,
    "Database connection failed",
    "Unable to connect to PostgreSQL database",
    "DB_CONNECTION_ERROR",
    http.StatusServiceUnavailable,
)
customErr.Context["database_host"] = "db.example.com"
customErr.Context["connection_timeout"] = "30s"
```

### Configuration

1. **Environment-Specific Settings**: Use appropriate configurations for each environment
2. **Security First**: Always mask internal errors in production
3. **Monitoring**: Enable metrics tracking for production monitoring
4. **Logging**: Configure appropriate log levels for your environment

```go
// Development
config := middleware.GetVerboseErrorHandlingConfig()

// Production
config := middleware.GetProductionErrorHandlingConfig()

// Custom production with specific settings
config := middleware.GetProductionErrorHandlingConfig()
config.LogLevel = zapcore.WarnLevel
config.ErrorCodes["custom_error"] = "CUSTOM_ERROR_CODE"
```

### Error Handling

1. **Don't Panic**: Use proper error handling instead of panics
2. **Log Appropriately**: Log errors with sufficient context for debugging
3. **Monitor Metrics**: Track error rates and patterns
4. **User Experience**: Provide helpful error messages to users

```go
// Good: Proper error handling
func myHandler(w http.ResponseWriter, r *http.Request) {
    data, err := processRequest(r)
    if err != nil {
        // Log error with context
        logger.Error("Failed to process request",
            zap.Error(err),
            zap.String("path", r.URL.Path),
            zap.String("method", r.Method),
        )
        
        // Return appropriate error
        http.Error(w, "Failed to process request", http.StatusBadRequest)
        return
    }
    
    // Process successful response
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(data)
}
```

## Troubleshooting

### Common Issues

1. **Errors Not Being Logged**
   - Check that `LogErrors` is set to `true`
   - Verify logger configuration
   - Ensure middleware is applied correctly

2. **Panic Recovery Not Working**
   - Verify `RecoverPanics` is set to `true`
   - Check that middleware is applied before handlers
   - Ensure proper error response is being sent

3. **Request ID Not Included**
   - Ensure request logging middleware is applied before error handling
   - Check that request ID is set in context
   - Verify middleware order

4. **Custom Error Handlers Not Working**
   - Check that custom handlers are registered correctly
   - Verify error type matching
   - Ensure handler returns proper APIError

### Debug Mode

Enable verbose configuration for debugging:

```go
config := middleware.GetVerboseErrorHandlingConfig()
config.LogLevel = zapcore.DebugLevel
config.IncludeStack = true
config.IncludeContext = true
config.MaskInternal = false

errorMiddleware := middleware.NewErrorHandlingMiddleware(config, logger)
```

### Performance Considerations

1. **Error Metrics**: Metrics tracking has minimal overhead
2. **Logging**: Use appropriate log levels to avoid performance impact
3. **Context Inclusion**: Only include necessary context to avoid large responses
4. **Custom Handlers**: Keep custom handlers lightweight

## Migration Guide

### From Basic Error Handling

If you're migrating from basic error handling:

1. **Replace http.Error calls**:
   ```go
   // Before
   http.Error(w, "Invalid input", http.StatusBadRequest)
   
   // After
   err := middleware.CreateValidationError("Invalid input", "Missing required field")
   // Error handling middleware will process this automatically
   ```

2. **Update error responses**:
   ```go
   // Before
   w.WriteHeader(http.StatusBadRequest)
   json.NewEncoder(w).Encode(map[string]string{"error": "Invalid input"})
   
   // After
   // Error handling middleware provides standardized responses
   ```

3. **Add middleware to your stack**:
   ```go
   // Add error handling middleware to your existing middleware chain
   handler := errorHandlingMiddleware.Middleware(yourExistingHandler)
   ```

### From Custom Error Handling

If you have existing custom error handling:

1. **Map existing error types** to the new error types
2. **Update error creation** to use the new helper functions
3. **Configure custom handlers** for specific error scenarios
4. **Update tests** to use the new error response format

## Security Considerations

1. **Information Disclosure**: Never expose sensitive information in error responses
2. **Internal Errors**: Always mask internal errors in production
3. **Stack Traces**: Only include stack traces in development environments
4. **Error Context**: Be careful about what context is included in error responses
5. **Logging**: Ensure sensitive data is not logged in error messages

## Monitoring and Alerting

### Error Rate Monitoring

```go
// Monitor error rates
func monitorErrorRates(middleware *middleware.ErrorHandlingMiddleware) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        metrics := middleware.GetErrorMetrics()
        
        // Alert on high error rates
        if metrics.TotalErrors > 100 {
            // Send alert
            sendAlert("High error rate detected", metrics)
        }
        
        // Alert on critical errors
        if metrics.ErrorsBySeverity[middleware.ErrorSeverityCritical] > 10 {
            // Send critical alert
            sendCriticalAlert("Critical errors detected", metrics)
        }
    }
}
```

### Error Pattern Analysis

```go
// Analyze error patterns
func analyzeErrorPatterns(middleware *middleware.ErrorHandlingMiddleware) {
    metrics := middleware.GetErrorMetrics()
    
    // Check for specific error patterns
    if metrics.ErrorsByType[middleware.ErrorTypeAuthentication] > 50 {
        // Potential security issue
        logSecurityAlert("High authentication failure rate")
    }
    
    if metrics.ErrorsByType[middleware.ErrorTypeExternal] > 20 {
        // External service issues
        logServiceAlert("External service errors detected")
    }
}
```

This comprehensive error handling middleware provides robust error management capabilities while maintaining security, performance, and ease of use. It integrates seamlessly with the existing middleware stack and provides the foundation for reliable error handling across the Enhanced Business Intelligence System.
