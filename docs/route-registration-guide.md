# Risk API Routes Registration Guide

**Date**: January 2025  
**Purpose**: Guide for registering new risk API endpoints

---

## Overview

The new risk API endpoints (`/api/v1/risk/benchmarks` and `/api/v1/risk/predictions/{merchantId}`) have been implemented in `internal/api/handlers/risk.go` and route registration is available in `internal/api/routes/risk_routes.go`.

---

## Route Registration

### Option 1: Using RegisterRiskRoutes (Recommended)

The `RegisterRiskRoutes` function in `internal/api/routes/risk_routes.go` registers all risk endpoints including the new ones:

```go
import (
    "kyb-platform/internal/api/handlers"
    "kyb-platform/internal/api/routes"
    "kyb-platform/internal/observability"
    "kyb-platform/internal/risk"
)

// Initialize services
logger := observability.NewLogger()
riskService := risk.NewRiskService(...)
riskHistoryService := risk.NewRiskHistoryService(...)

// Initialize handler
riskHandler := handlers.NewRiskHandler(logger, riskService, riskHistoryService)

// Register routes
mux := http.NewServeMux()
routes.RegisterRiskRoutes(mux, riskHandler)
```

### Option 2: Manual Registration

If you need to register routes manually:

```go
import (
    "kyb-platform/internal/api/handlers"
    "kyb-platform/internal/api/middleware"
)

// Initialize handler (same as above)
riskHandler := handlers.NewRiskHandler(logger, riskService, riskHistoryService)

// Register new endpoints
mux.HandleFunc("GET /v1/risk/benchmarks",
    middleware.RequestIDMiddleware(
        middleware.LoggingMiddleware(
            middleware.CORSMiddleware(
                riskHandler.GetRiskBenchmarksHandler))))

mux.HandleFunc("GET /v1/risk/predictions/{merchant_id}",
    middleware.RequestIDMiddleware(
        middleware.LoggingMiddleware(
            middleware.CORSMiddleware(
                riskHandler.GetRiskPredictionsHandler))))
```

---

## Endpoints Registered

The `RegisterRiskRoutes` function registers the following endpoints:

### Existing Endpoints
- `POST /v1/risk/assess` - Risk assessment
- `GET /v1/risk/history/{business_id}` - Risk history
- `GET /v1/risk/categories` - Risk categories
- `GET /v1/risk/factors` - Risk factors
- `GET /v1/risk/thresholds` - Risk thresholds
- `GET /v1/risk/industry-benchmarks/{industry}` - Industry benchmarks (legacy)

### New Endpoints
- **`GET /v1/risk/benchmarks`** - Industry benchmarks by MCC/NAICS/SIC
- **`GET /v1/risk/predictions/{merchant_id}`** - Risk predictions

---

## Integration Points

### API Gateway

The API Gateway (`services/api-gateway/cmd/main.go`) currently proxies `/risk` paths to the risk assessment service:

```go
api.PathPrefix("/risk").HandlerFunc(gatewayHandler.ProxyToRiskAssessment)
```

**Note**: The new endpoints are in the main platform handlers, not the risk assessment service. You may need to:

1. **Option A**: Register routes in the main platform server (if separate from API gateway)
2. **Option B**: Update API gateway to route specific paths to main platform
3. **Option C**: Add endpoints to risk assessment service (requires handler migration)

### Recommended Approach

Since the handlers are in `internal/api/handlers/risk.go` (main platform), the routes should be registered in the main platform server, not the risk assessment service.

If using the API gateway pattern:
- The gateway should route `/api/v1/risk/benchmarks` and `/api/v1/risk/predictions/*` to the main platform server
- Or these endpoints can be added directly to the gateway if it has access to the handlers

---

## Service Dependencies

The `RiskHandler` requires:

1. **RiskService** (`*risk.RiskService`)
   - Provides `GetIndustryBenchmarks()` method
   - Used by benchmarks endpoint

2. **RiskHistoryService** (`*risk.RiskHistoryService`)
   - Provides `GetRiskHistory()` method
   - Used by predictions endpoint

3. **Logger** (`*observability.Logger`)
   - For logging and observability

---

## Testing Routes

After registration, test the endpoints:

```bash
# Test benchmarks endpoint
curl "http://localhost:8080/api/v1/risk/benchmarks?mcc=5411"

# Test predictions endpoint
curl "http://localhost:8080/api/v1/risk/predictions/merchant-123?horizons=3,6,12&includeScenarios=true&includeConfidence=true"
```

---

## Middleware Stack

All routes use the following middleware stack:
1. **RequestIDMiddleware** - Adds request ID for tracing
2. **LoggingMiddleware** - Logs request/response
3. **CORSMiddleware** - Handles CORS headers

Additional middleware can be added as needed (auth, rate limiting, etc.).

---

## Next Steps

1. **Identify Main Server**: Determine which server initializes the main platform handlers
2. **Register Routes**: Call `RegisterRiskRoutes()` in server initialization
3. **Update API Gateway**: Configure routing if using gateway pattern
4. **Test Integration**: Verify endpoints are accessible
5. **Update Documentation**: Add endpoints to API documentation

---

## Files Reference

- **Handlers**: `internal/api/handlers/risk.go`
- **Routes**: `internal/api/routes/risk_routes.go`
- **API Gateway**: `services/api-gateway/cmd/main.go`
- **Risk Assessment Service**: `services/risk-assessment-service/cmd/main.go`

---

## Status

✅ **Handlers Implemented**  
✅ **Route Registration Function Created**  
⏳ **Routes Need to be Registered in Main Server** (pending server identification)

