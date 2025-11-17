# Production Routing Fixes

**Date**: 2025-11-17  
**Status**: 🔧 **FIXES APPLIED - AWAITING DEPLOYMENT**

## Issues Identified

### Issue 1: Missing URL Scheme (https://)
**Error**: `Get "bi-service-production.up.railway.app/dashboard/kpis": unsupported protocol scheme ""`

**Root Cause**: The `proxyRequest` function was not validating that the target URL has a proper scheme (http:// or https://). If the `BIServiceURL` environment variable was empty or malformed, it would construct an invalid URL.

**Fix Applied**:
- Added URL validation in `proxyRequest` function
- Check if targetURL is empty and return error
- Automatically add `https://` prefix if scheme is missing
- Added logging for URL corrections

### Issue 2: Missing Middleware on v3 Routes
**Error**: Routes returning 404 or not matching correctly

**Root Cause**: The v3 API routes (`/api/v3/*`) were created as a separate subrouter but didn't have middleware applied, which could cause routing issues.

**Fix Applied**:
- Added full middleware chain to v3 routes:
  - CORS middleware
  - Security headers
  - Logging
  - Rate limiting
  - Authentication

### Issue 3: Missing BI Service URL in Logging
**Issue**: BI Service URL not logged during startup, making debugging difficult

**Fix Applied**:
- Added `bi_service_url` to configuration logging

## Code Changes

### `services/api-gateway/internal/handlers/gateway.go`
```go
// Added URL validation and scheme checking
func (h *GatewayHandler) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL, targetPath string) {
    // Validate targetURL is not empty
    if targetURL == "" {
        h.logger.Error("Target URL is empty", ...)
        gatewayerrors.WriteServiceUnavailable(w, r, "Backend service URL not configured")
        return
    }

    // Ensure targetURL has a scheme (https://)
    if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
        targetURL = "https://" + targetURL
        h.logger.Warn("Added https:// prefix to target URL", ...)
    }
    // ... rest of function
}
```

### `services/api-gateway/cmd/main.go`
```go
// Added middleware to v3 routes
apiV3 := router.PathPrefix("/api/v3").Subrouter()
apiV3.Use(middleware.CORS(cfg.CORS))
apiV3.Use(middleware.SecurityHeaders)
apiV3.Use(middleware.Logging(logger))
apiV3.Use(middleware.RateLimit(cfg.RateLimit))
apiV3.Use(middleware.Authentication(supabaseClient, logger))
apiV3.HandleFunc("/dashboard/metrics", gatewayHandler.ProxyToDashboardMetricsV3).Methods("GET", "OPTIONS")

// Added BI service URL to logging
logger.Info("🔧 Configuration loaded",
    ...
    zap.String("bi_service_url", cfg.Services.BIServiceURL),
    ...)
```

## Expected Results After Deployment

After Railway deploys the fixes:

1. **v3 Dashboard Metrics**: `/api/v3/dashboard/metrics` should return 200 OK with BI Service data
2. **v1 Dashboard Metrics**: `/api/v1/dashboard/metrics` should return 200 OK with Risk Assessment Service data
3. **Compliance Status**: `/api/v1/compliance/status` should return 200 OK or appropriate error
4. **Sessions**: `/api/v1/sessions` should return 200 OK with session data

## Next Steps

1. Wait for Railway deployment to complete
2. Re-run production tests
3. Verify all endpoints return expected responses
4. Check API Gateway logs for URL validation warnings/errors

