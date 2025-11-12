# Next Steps: Operational Checklist

This document outlines the remaining steps needed to make the merchant-details API endpoints fully operational.

## ✅ Completed

- [x] API testing environment (Postman/Insomnia collections)
- [x] OpenAPI 3.0 specification
- [x] Backend handlers, services, and repositories implemented
- [x] Route registration functions created
- [x] E2E and integration tests written
- [x] CI/CD pipeline enhancements
- [x] Bug fixes (async routes registration, error handling)

## 🔧 Required Next Steps

### 1. Run Database Migration

**Priority:** Critical  
**Duration:** 5 minutes

The async risk assessment feature requires a database migration to add necessary columns.

#### Steps:

1. **Connect to your database** (Supabase, PostgreSQL, etc.)

2. **Run the migration:**
   ```bash
   # Option 1: Using psql
   psql $DATABASE_URL -f internal/database/migrations/010_add_async_risk_assessment_columns.sql
   
   # Option 2: Using the migration tool
   go run cmd/migrate/main.go
   
   # Option 3: Via Supabase dashboard
   # Copy and paste the SQL from internal/database/migrations/010_add_async_risk_assessment_columns.sql
   ```

3. **Verify migration:**
   ```sql
   -- Check that columns exist
   SELECT column_name, data_type 
   FROM information_schema.columns 
   WHERE table_name = 'risk_assessments' 
   AND column_name IN ('merchant_id', 'status', 'options', 'result', 'progress', 'estimated_completion', 'completed_at');
   ```

**Migration File:** `internal/database/migrations/010_add_async_risk_assessment_columns.sql`

---

### 2. Register Routes in Main Application

**Priority:** Critical  
**Duration:** 15-30 minutes

The route registration functions exist, but they need to be called in your main server application.

#### Option A: If using a centralized route registration

Find where routes are registered (likely in `cmd/*/main.go` or a routes setup function) and add:

```go
import (
    "kyb-platform/internal/api/handlers"
    "kyb-platform/internal/api/middleware"
    "kyb-platform/internal/api/routes"
    "kyb-platform/internal/database"
    "kyb-platform/internal/services"
    "kyb-platform/internal/observability"
)

// In your server initialization:

// 1. Initialize database connection
db := initializeDatabase() // Your existing DB initialization
logger := log.Default()
obsLogger := observability.NewLogger(/* your config */)

// 2. Initialize repositories
merchantRepo := database.NewMerchantPortfolioRepository(db, logger)
analyticsRepo := database.NewMerchantAnalyticsRepository(db, logger)
riskAssessmentRepo := database.NewRiskAssessmentRepository(db, logger)

// 3. Initialize services
analyticsService := services.NewMerchantAnalyticsService(analyticsRepo, merchantRepo, logger)
riskAssessmentService := services.NewRiskAssessmentService(riskAssessmentRepo, logger)

// 4. Initialize handlers
merchantHandler := handlers.NewMerchantPortfolioHandler(/* existing params */)
analyticsHandler := handlers.NewMerchantAnalyticsHandler(analyticsService, logger)
riskHandler := handlers.NewRiskHandler(/* existing params */)
asyncRiskHandler := handlers.NewAsyncRiskAssessmentHandler(riskAssessmentService, logger)

// 5. Initialize middleware
authMiddleware := middleware.NewAuthMiddleware(/* your auth service */, logger)
rateLimiter := middleware.NewAPIRateLimiter(100, 1*time.Minute, logger)

// 6. Create mux
mux := http.NewServeMux()

// 7. Register merchant routes (includes analytics)
merchantConfig := &routes.MerchantRouteConfig{
    MerchantPortfolioHandler: merchantHandler,
    MerchantAnalyticsHandler:  analyticsHandler, // Add this
    AuthMiddleware:            authMiddleware,
    RateLimiter:              rateLimiter,
    Logger:                   obsLogger,
}
routes.RegisterMerchantRoutes(mux, merchantConfig)

// 8. Register risk routes (includes async routes)
asyncRiskConfig := &routes.AsyncRiskAssessmentRouteConfig{
    AsyncRiskHandler: asyncRiskHandler,
    AuthMiddleware:   authMiddleware,
    RateLimiter:      rateLimiter,
}
routes.RegisterRiskRoutesWithConfig(mux, riskHandler, asyncRiskConfig)

// 9. Start server with mux
http.ListenAndServe(":8080", mux)
```

#### Option B: If using individual route handlers

Add the routes manually to your existing route setup:

```go
// Merchant analytics routes
mux.Handle("GET /api/v1/merchants/{merchantId}/analytics",
    authMiddleware.RequireAuth(
        rateLimiter.Middleware(
            http.HandlerFunc(analyticsHandler.GetMerchantAnalytics),
        ),
    ),
)

mux.Handle("GET /api/v1/merchants/{merchantId}/website-analysis",
    authMiddleware.RequireAuth(
        rateLimiter.Middleware(
            http.HandlerFunc(analyticsHandler.GetWebsiteAnalysis),
        ),
    ),
)

// Async risk assessment routes
mux.Handle("POST /api/v1/risk/assess",
    authMiddleware.RequireAuth(
        rateLimiter.Middleware(
            http.HandlerFunc(asyncRiskHandler.AssessRisk),
        ),
    ),
)

mux.Handle("GET /api/v1/risk/assess/{assessmentId}",
    authMiddleware.RequireAuth(
        rateLimiter.Middleware(
            http.HandlerFunc(asyncRiskHandler.GetAssessmentStatus),
        ),
    ),
)
```

**Files to modify:**
- Your main server file (e.g., `cmd/railway-server/main.go`, `cmd/frontend-service/main.go`, etc.)
- Look for `setupRoutes()` or similar function

---

### 3. Verify Dependencies

**Priority:** High  
**Duration:** 5 minutes

Ensure all required dependencies are available:

1. **Database tables exist:**
   - `merchant_classifications`
   - `merchant_security_data`
   - `merchant_quality_data`
   - `merchant_website_analysis`
   - `risk_assessments` (with new columns from migration)

2. **Check imports compile:**
   ```bash
   go build ./internal/api/handlers/merchant_analytics_handler.go
   go build ./internal/api/handlers/async_risk_assessment_handler.go
   ```

---

### 4. Test Endpoints

**Priority:** High  
**Duration:** 10-15 minutes

After registering routes, test the endpoints:

#### Test Merchant Analytics:

```bash
# Get merchant analytics
curl -X GET "http://localhost:8080/api/v1/merchants/{merchantId}/analytics" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"

# Get website analysis
curl -X GET "http://localhost:8080/api/v1/merchants/{merchantId}/website-analysis" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

#### Test Async Risk Assessment:

```bash
# Start assessment
curl -X POST "http://localhost:8080/api/v1/risk/assess" \
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

# Check status (use assessmentId from previous response)
curl -X GET "http://localhost:8080/api/v1/risk/assess/{assessmentId}" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

**Or use the Postman/Insomnia collections:**
- `tests/api/merchant-details/postman-collection.json`
- `tests/api/merchant-details/insomnia-collection.json`

---

### 5. Run Integration Tests

**Priority:** Medium  
**Duration:** 5 minutes

Verify the tests pass:

```bash
# Run E2E tests
go test -v -tags=e2e ./test/e2e/merchant_details_e2e_test.go ./test/e2e/merchant_analytics_api_test.go

# Run integration tests
go test -v -tags=integration ./test/integration/risk_assessment_integration_test.go
```

---

### 6. Update Documentation (Optional)

**Priority:** Low  
**Duration:** 10 minutes

If you have a main README or API documentation:

1. Add the new endpoints to your API documentation
2. Update any deployment guides
3. Add environment variable requirements if needed

---

## Quick Start Script

Here's a quick checklist script you can run:

```bash
#!/bin/bash

echo "🔍 Checking prerequisites..."

# Check if migration file exists
if [ -f "internal/database/migrations/010_add_async_risk_assessment_columns.sql" ]; then
    echo "✅ Migration file exists"
else
    echo "❌ Migration file not found"
    exit 1
fi

# Check if handlers compile
echo "🔨 Checking handlers compile..."
if go build ./internal/api/handlers/merchant_analytics_handler.go 2>/dev/null; then
    echo "✅ Merchant analytics handler compiles"
else
    echo "❌ Merchant analytics handler has compilation errors"
fi

if go build ./internal/api/handlers/async_risk_assessment_handler.go 2>/dev/null; then
    echo "✅ Async risk assessment handler compiles"
else
    echo "❌ Async risk assessment handler has compilation errors"
fi

# Check if routes compile
echo "🔨 Checking routes compile..."
if go build ./internal/api/routes/merchant_routes.go 2>/dev/null; then
    echo "✅ Merchant routes compile"
else
    echo "❌ Merchant routes have compilation errors"
fi

if go build ./internal/api/routes/risk_routes.go 2>/dev/null; then
    echo "✅ Risk routes compile"
else
    echo "❌ Risk routes have compilation errors"
fi

echo ""
echo "📋 Next steps:"
echo "1. Run database migration: psql \$DATABASE_URL -f internal/database/migrations/010_add_async_risk_assessment_columns.sql"
echo "2. Register routes in your main server application"
echo "3. Test endpoints using Postman/Insomnia collections"
echo "4. Run integration tests"
```

Save as `scripts/check-readiness.sh` and run: `chmod +x scripts/check-readiness.sh && ./scripts/check-readiness.sh`

---

## Troubleshooting

### Issue: Routes return 404

**Solution:**
- Verify routes are registered in your main server
- Check route paths match exactly (case-sensitive)
- Ensure middleware is properly configured

### Issue: Database errors

**Solution:**
- Verify migration has been run
- Check database connection string
- Ensure tables exist with correct schema

### Issue: Authentication errors

**Solution:**
- Verify `AuthMiddleware` is properly initialized
- Check token format and validity
- Ensure routes are wrapped with `RequireAuth`

### Issue: Handler not found errors

**Solution:**
- Verify handlers are initialized before route registration
- Check dependency injection is correct
- Ensure services and repositories are properly initialized

---

## Support

For detailed integration examples, see:
- `docs/async-routes-integration-guide.md` - Complete integration guide
- `api/openapi/merchant-details-api-spec.yaml` - API specification
- `tests/api/merchant-details/README.md` - Testing guide

---

**Last Updated:** January 2025  
**Status:** Ready for Implementation

