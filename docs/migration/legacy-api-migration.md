# Legacy Enhanced API Server Migration Guide

## Overview

This guide helps developers migrate from the legacy `EnhancedServer` to the new modular API architecture. The new architecture provides better separation of concerns, improved maintainability, and enhanced scalability.

## What Changed

### Before (Legacy Architecture)
```go
// Monolithic server with all functionality in one file
server := NewEnhancedServer(port)
server.Start()
```

### After (Modular Architecture)
```go
// Modular approach with separate handlers and services
server := NewServer(config, logger, metrics, classificationSvc, ...)
server.Start()
```

## Migration Steps

### Step 1: Update Server Initialization

**Replace legacy server creation:**
```go
// OLD
server := NewEnhancedServer(port)

// NEW
server := NewServer(
    config,
    logger,
    metrics,
    classificationSvc,
    riskService,
    riskHistoryService,
    riskHandler,
    dashboardHandler,
    authService,
    authHandler,
    authMiddleware,
    adminService,
    adminHandler,
    complianceHandler,
    soc2Handler,
    pciHandler,
    gdprHandler,
    auditHandler,
    rateLimiter,
    authRateLimiter,
    ipBlocker,
    validator,
)
```

### Step 2: Update Route Handlers

**Replace legacy route handlers:**
```go
// OLD (in main-enhanced.go)
mux.HandleFunc("POST /api/v1/classify", handleClassification)
mux.HandleFunc("GET /api/v1/health", handleHealth)

// NEW (in main.go with proper handlers)
server.setupRoutes()
```

### Step 3: Update Classification Endpoint

**Replace legacy classification handler:**
```go
// OLD
func handleClassification(w http.ResponseWriter, r *http.Request) {
    // All classification logic in one function
}

// NEW
// Use the intelligent router with modular classification
result := router.Process(ctx, request)
```

### Step 4: Update Response Models

**Replace legacy response models:**
```go
// OLD
type ClassificationResponse struct {
    PrimaryClassification *IndustryClassification
    RawData              interface{}
}

// NEW
type EnhancedClassificationResponse struct {
    Classification *IndustryClassification
    ModuleUsed     string
    Confidence     float64
    Metadata       map[string]interface{}
}
```

## Handler Migration

### Classification Handler

**Legacy Handler:** `handleClassification()` in `main-enhanced.go`
**New Handler:** `internal/api/handlers/enhanced_classification.go`

```go
// OLD
func handleClassification(w http.ResponseWriter, r *http.Request) {
    // Monolithic classification logic
}

// NEW
type EnhancedClassificationHandler struct {
    router *routing.IntelligentRouter
    logger *observability.Logger
}

func (h *EnhancedClassificationHandler) HandleClassification(w http.ResponseWriter, r *http.Request) {
    // Modular classification using intelligent router
}
```

### Health Check Handler

**Legacy Handler:** `handleHealth()` in `main-enhanced.go`
**New Handler:** `internal/api/handlers/health.go`

```go
// OLD
func handleHealth(w http.ResponseWriter, r *http.Request) {
    // Simple health check
}

// NEW
type HealthHandler struct {
    services map[string]HealthChecker
}

func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
    // Comprehensive health check for all services
}
```

## Configuration Changes

### Legacy Configuration
```go
// All configuration in environment variables
port := os.Getenv("PORT")
```

### New Configuration
```go
// Structured configuration with validation
config := &config.Config{
    Server: config.ServerConfig{
        Port: 8080,
        Host: "0.0.0.0",
    },
    Database: config.DatabaseConfig{
        Host:     "localhost",
        Port:     5432,
        Name:     "kyb_platform",
        Username: "postgres",
    },
    // ... other configuration sections
}
```

## Error Handling

### Legacy Error Handling
```go
// Basic error handling
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

### New Error Handling
```go
// Structured error handling with logging
if err != nil {
    logger.Error("Classification failed", "error", err, "request_id", requestID)
    response := ErrorResponse{
        Error:   "Classification failed",
        Code:    "CLASSIFICATION_ERROR",
        Details: err.Error(),
    }
    respondWithJSON(w, http.StatusInternalServerError, response)
    return
}
```

## Testing Migration

### Legacy Testing
```go
func TestClassificationEndpoint(t *testing.T) {
    server := NewEnhancedServer("8080")
    // Test with embedded server
}
```

### New Testing
```go
func TestClassificationEndpoint(t *testing.T) {
    server := NewTestServer()
    defer server.Close()
    
    resp, err := http.Post(server.URL+"/api/v1/classify", "application/json", body)
    // Assertions
}
```

## Performance Considerations

### Benefits of New Architecture
- **Separation of Concerns**: Each handler has a single responsibility
- **Middleware Support**: Authentication, rate limiting, validation
- **Observability**: Comprehensive logging and metrics
- **Scalability**: Independent service scaling
- **Testability**: Isolated unit tests for each component

### Migration Checklist

- [ ] Update server initialization
- [ ] Replace route handlers
- [ ] Update classification endpoint
- [ ] Update response models
- [ ] Update configuration
- [ ] Update error handling
- [ ] Update tests
- [ ] Verify functionality
- [ ] Update documentation

## Troubleshooting

### Common Issues

1. **Handler Not Found Errors**
   ```
   handler not found: /api/v1/classify
   ```
   **Solution**: Ensure routes are properly registered in `setupRoutes()`

2. **Configuration Errors**
   ```
   invalid configuration: missing database host
   ```
   **Solution**: Check configuration validation and required fields

3. **Response Structure Errors**
   ```
   unknown field PrimaryClassification
   ```
   **Solution**: Update response handling to use new structure

### Debugging Tips

1. **Enable Debug Logging**
   ```go
   logger.SetLevel(logging.Debug)
   ```

2. **Check Service Health**
   ```go
   health := server.GetHealthStatus()
   for service, status := range health {
       log.Printf("Service %s: %s", service, status)
   }
   ```

3. **Monitor Performance**
   ```go
   metrics := server.GetMetrics()
   log.Printf("Request rate: %f req/s", metrics.RequestRate)
   ```

## Support

If you encounter issues during migration:

1. Check the [API Documentation](../api/api-v3-endpoints.md)
2. Review the [Handler Documentation](../api/handlers.md)
3. Use the troubleshooting section above
4. Contact the development team for assistance

## Conclusion

The new modular API architecture provides significant benefits in terms of maintainability, scalability, and performance. By following this guide, you can successfully migrate from the legacy enhanced server to the new modular approach while maintaining backward compatibility during the transition period.
