# Async Risk Assessment Routes Integration Guide

This guide explains how to properly integrate the async risk assessment routes and merchant analytics routes into your application.

## Overview

The async risk assessment feature provides two endpoints:
- `POST /api/v1/risk/assess` - Start an asynchronous risk assessment
- `GET /api/v1/risk/assess/{assessmentId}` - Get the status of a risk assessment

The merchant analytics feature provides:
- `GET /api/v1/merchants/{merchantId}/analytics` - Get comprehensive analytics for a merchant
- `GET /api/v1/merchants/{merchantId}/website-analysis` - Get website analysis for a merchant

## Route Registration

### Option 1: Using RegisterRiskRoutesWithConfig (Recommended)

To register async risk assessment routes along with standard risk routes:

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "kyb-platform/internal/api/handlers"
    "kyb-platform/internal/api/middleware"
    "kyb-platform/internal/api/routes"
    "kyb-platform/internal/database"
    "kyb-platform/internal/services"
)

func main() {
    // Initialize dependencies
    db := initializeDatabase() // Your DB initialization
    logger := log.Default()
    
    // Create standard risk handler
    riskHandler := handlers.NewRiskHandler(/* ... */)
    
    // Create async risk assessment handler
    riskAssessmentRepo := database.NewRiskAssessmentRepository(db, logger)
    riskAssessmentService := services.NewRiskAssessmentService(riskAssessmentRepo, logger)
    asyncRiskHandler := handlers.NewAsyncRiskAssessmentHandler(riskAssessmentService, logger)
    
    // Create middleware
    authMiddleware := middleware.NewAuthMiddleware(/* ... */)
    rateLimiter := middleware.NewAPIRateLimiter(100, 1*time.Minute, logger)
    
    // Create mux
    mux := http.NewServeMux()
    
    // Register risk routes with async config
    asyncConfig := &routes.AsyncRiskAssessmentRouteConfig{
        AsyncRiskHandler: asyncRiskHandler,
        AuthMiddleware:   authMiddleware,
        RateLimiter:     rateLimiter,
    }
    
    routes.RegisterRiskRoutesWithConfig(mux, riskHandler, asyncConfig)
    
    // Start server
    http.ListenAndServe(":8080", mux)
}
```

### Option 2: Registering Async Routes Separately

If you prefer to register routes separately:

```go
// Register standard risk routes
routes.RegisterRiskRoutes(mux, riskHandler)

// Register async routes separately
asyncConfig := &routes.AsyncRiskAssessmentRouteConfig{
    AsyncRiskHandler: asyncRiskHandler,
    AuthMiddleware:   authMiddleware,
    RateLimiter:     rateLimiter,
}
routes.RegisterAsyncRiskAssessmentRoutes(mux, asyncConfig)
```

## Merchant Analytics Routes Integration

To register merchant analytics routes:

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "kyb-platform/internal/api/handlers"
    "kyb-platform/internal/api/middleware"
    "kyb-platform/internal/api/routes"
    "kyb-platform/internal/database"
    "kyb-platform/internal/observability"
    "kyb-platform/internal/services"
)

func main() {
    // Initialize dependencies
    db := initializeDatabase()
    logger := log.Default()
    obsLogger := observability.NewLogger(/* ... */)
    
    // Create merchant portfolio handler
    merchantHandler := handlers.NewMerchantPortfolioHandler(/* ... */)
    
    // Create merchant analytics handler
    merchantRepo := database.NewMerchantPortfolioRepository(db, logger)
    analyticsRepo := database.NewMerchantAnalyticsRepository(db, logger)
    analyticsService := services.NewMerchantAnalyticsService(analyticsRepo, merchantRepo, logger)
    analyticsHandler := handlers.NewMerchantAnalyticsHandler(analyticsService, logger)
    
    // Create middleware
    authMiddleware := middleware.NewAuthMiddleware(/* ... */)
    rateLimiter := middleware.NewAPIRateLimiter(100, 1*time.Minute, logger)
    
    // Create route config
    merchantConfig := &routes.MerchantRouteConfig{
        MerchantPortfolioHandler: merchantHandler,
        MerchantAnalyticsHandler:  analyticsHandler, // Add analytics handler
        AuthMiddleware:            authMiddleware,
        RateLimiter:              rateLimiter,
        Logger:                   obsLogger,
    }
    
    // Create mux
    mux := http.NewServeMux()
    
    // Register merchant routes (includes analytics routes)
    routes.RegisterMerchantRoutes(mux, merchantConfig)
    
    // Start server
    http.ListenAndServe(":8080", mux)
}
```

## Complete Integration Example

Here's a complete example showing both route sets:

```go
package main

import (
    "database/sql"
    "log"
    "net/http"
    "time"
    
    _ "github.com/lib/pq"
    
    "kyb-platform/internal/api/handlers"
    "kyb-platform/internal/api/middleware"
    "kyb-platform/internal/api/routes"
    "kyb-platform/internal/database"
    "kyb-platform/internal/observability"
    "kyb-platform/internal/services"
)

func main() {
    // Initialize database
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    logger := log.Default()
    obsLogger := observability.NewLogger(/* ... */)
    
    // Initialize repositories
    merchantRepo := database.NewMerchantPortfolioRepository(db, logger)
    analyticsRepo := database.NewMerchantAnalyticsRepository(db, logger)
    riskAssessmentRepo := database.NewRiskAssessmentRepository(db, logger)
    
    // Initialize services
    analyticsService := services.NewMerchantAnalyticsService(analyticsRepo, merchantRepo, logger)
    riskAssessmentService := services.NewRiskAssessmentService(riskAssessmentRepo, logger)
    
    // Initialize handlers
    merchantHandler := handlers.NewMerchantPortfolioHandler(/* ... */)
    analyticsHandler := handlers.NewMerchantAnalyticsHandler(analyticsService, logger)
    riskHandler := handlers.NewRiskHandler(/* ... */)
    asyncRiskHandler := handlers.NewAsyncRiskAssessmentHandler(riskAssessmentService, logger)
    
    // Initialize middleware
    authMiddleware := middleware.NewAuthMiddleware(/* ... */)
    rateLimiter := middleware.NewAPIRateLimiter(100, 1*time.Minute, logger)
    
    // Create mux
    mux := http.NewServeMux()
    
    // Register merchant routes
    merchantConfig := &routes.MerchantRouteConfig{
        MerchantPortfolioHandler: merchantHandler,
        MerchantAnalyticsHandler:  analyticsHandler,
        AuthMiddleware:            authMiddleware,
        RateLimiter:              rateLimiter,
        Logger:                   obsLogger,
    }
    routes.RegisterMerchantRoutes(mux, merchantConfig)
    
    // Register risk routes with async config
    asyncRiskConfig := &routes.AsyncRiskAssessmentRouteConfig{
        AsyncRiskHandler: asyncRiskHandler,
        AuthMiddleware:   authMiddleware,
        RateLimiter:      rateLimiter,
    }
    routes.RegisterRiskRoutesWithConfig(mux, riskHandler, asyncRiskConfig)
    
    // Start server
    log.Println("Server starting on :8080")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatal("Server failed:", err)
    }
}
```

## Error Handling

The handlers now properly handle wrapped errors:

- **Merchant Not Found**: Returns `404 Not Found` when a merchant doesn't exist
- **Assessment Not Found**: Returns `404 Not Found` when an assessment doesn't exist
- **Other Errors**: Returns `500 Internal Server Error` for unexpected errors

The error checking uses `errors.Is()` to properly unwrap and check for specific error types, even when errors are wrapped with `fmt.Errorf()`.

## Testing

To test the async routes:

```bash
# Start a risk assessment
curl -X POST http://localhost:8080/api/v1/risk/assess \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "merchantId": "merchant-123",
    "options": {
      "includeHistory": true,
      "includePredictions": true,
      "recalculate": false
    }
  }'

# Check assessment status
curl http://localhost:8080/api/v1/risk/assess/{assessmentId} \
  -H "Authorization: Bearer YOUR_TOKEN"
```

To test merchant analytics:

```bash
# Get merchant analytics
curl http://localhost:8080/api/v1/merchants/{merchantId}/analytics \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get website analysis
curl http://localhost:8080/api/v1/merchants/{merchantId}/website-analysis \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Notes

1. **Backward Compatibility**: The original `RegisterRiskRoutes()` function still works and will not register async routes unless you use `RegisterRiskRoutesWithConfig()`.

2. **Middleware**: Both async risk routes and merchant analytics routes require authentication and rate limiting middleware.

3. **Database Migrations**: Ensure you've run the migration `010_add_async_risk_assessment_columns.sql` before using async risk assessment features.

4. **Service Dependencies**: The async risk assessment service requires a background job processor. The service automatically starts a goroutine to process jobs from the queue.

