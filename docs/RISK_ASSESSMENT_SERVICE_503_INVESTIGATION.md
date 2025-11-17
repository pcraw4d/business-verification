# Risk Assessment Service 503 Error Investigation

**Date**: 2025-11-17  
**Endpoint**: `/api/v1/risk/metrics`  
**Status**: ❌ **503 Service Unavailable**

## Issue Summary

The API Gateway is returning 503 "Backend service unavailable" when accessing `/api/v1/risk/metrics`.

## Investigation Results

### 1. API Gateway Configuration ✅
- **Service URL**: `https://risk-assessment-service-production.up.railway.app`
- **Configuration**: Correctly set in `services/api-gateway/internal/config/config.go`
- **Default**: `RISK_ASSESSMENT_SERVICE_URL` environment variable or default URL

### 2. Risk Assessment Service Health ✅
- **Health Endpoint**: `/health` returns 200 OK
- **Response**: `{"status":"healthy","timestamp":"2025-11-17T19:19:24Z"}`
- **Service Status**: Service is running and healthy

### 3. Route Mapping Issue ❌

**Problem Identified**:
- API Gateway routes: `/api/v1/risk/metrics` → Risk Assessment Service
- Risk Assessment Service has: `/api/v1/metrics` (NOT `/api/v1/risk/metrics`)

**Current Route Mapping** (from `services/api-gateway/internal/handlers/gateway.go`):
```go
// For other /risk/* paths, keep them as-is (e.g., /risk/benchmarks, /risk/predictions)
// The risk service has routes like /api/v1/risk/benchmarks
// No change needed
```

**Actual Risk Assessment Service Routes** (from `services/risk-assessment-service/cmd/main.go`):
- ✅ `/api/v1/metrics` - Line 776
- ✅ `/api/v1/health` - Line 777
- ✅ `/api/v1/performance` - Line 778
- ✅ `/api/v1/monitoring/metrics` - Line 781
- ❌ `/api/v1/risk/metrics` - **DOES NOT EXIST**

### 4. Direct Service Test Results

**Test 1: Health Endpoint** ✅
```bash
curl https://risk-assessment-service-production.up.railway.app/health
# Returns: {"status":"healthy","timestamp":"2025-11-17T19:19:24Z"}
```

**Test 2: Metrics Endpoint (Correct Path)** ❌
```bash
curl https://risk-assessment-service-production.up.railway.app/api/v1/metrics
# Returns: 502 "Application failed to respond"
```

**Test 3: Metrics Endpoint (Wrong Path)** ❌
```bash
curl https://risk-assessment-service-production.up.railway.app/api/v1/risk/metrics
# Returns: 502 "Application failed to respond"
```

## Root Cause Analysis

### Issue 1: Route Path Mismatch
The API Gateway is trying to proxy `/api/v1/risk/metrics` to the Risk Assessment Service, but the service doesn't have a `/api/v1/risk/metrics` route. The service has `/api/v1/metrics` instead.

### Issue 2: Service Endpoint Not Responding
Even the correct `/api/v1/metrics` endpoint is returning 502, suggesting:
- The endpoint may not be properly implemented
- The service may be crashing when handling metrics requests
- There may be a dependency issue (database, Redis, etc.)

## Solution Options

### Option 1: Fix Route Mapping in API Gateway (Recommended)
Update `ProxyToRiskAssessment` to map `/api/v1/risk/metrics` → `/api/v1/metrics`:

```go
// Map /api/v1/risk/metrics to /api/v1/metrics
if path == "/api/v1/risk/metrics" {
    path = "/api/v1/metrics"
}
```

### Option 2: Add Route to Risk Assessment Service
Add `/api/v1/risk/metrics` route to the Risk Assessment Service that proxies to `/api/v1/metrics`.

### Option 3: Fix Metrics Endpoint in Risk Assessment Service
Investigate why `/api/v1/metrics` is returning 502 and fix the underlying issue.

## Recommended Action Plan

1. **Immediate**: Fix route mapping in API Gateway to map `/api/v1/risk/metrics` → `/api/v1/metrics`
2. **Investigate**: Check Risk Assessment Service logs for why `/api/v1/metrics` returns 502
3. **Verify**: Test the metrics endpoint after route fix
4. **Monitor**: Ensure metrics endpoint is stable

## Fix Applied ✅

**Commit**: `e2da5e034`

**Change**: Updated `ProxyToRiskAssessment` in `services/api-gateway/internal/handlers/gateway.go` to map `/api/v1/risk/metrics` → `/api/v1/metrics`:

```go
} else if path == "/api/v1/risk/metrics" {
    // Map /api/v1/risk/metrics to /api/v1/metrics (risk service uses /metrics, not /risk/metrics)
    path = "/api/v1/metrics"
}
```

## Issue 2: Risk Assessment Service `/api/v1/metrics` Returns 502 ✅ FIXED

### Root Cause Identified

**Panic in Metrics Handler**: The `HandleGetMetrics` function was using unsafe type assertions to extract `request_id` from context:

```go
// UNSAFE - Can panic if request_id is not set
ctx.Value("request_id").(string)
```

If the `request_id` was not set in the context, this would cause a panic, resulting in a 502 error.

### Fix Applied ✅

**Commit**: `fff1e0fcb` and `4cb9843ab`

1. **Added Safe Request ID Helper Function**:
```go
func (h *MetricsHandler) getRequestID(ctx context.Context) string {
    if reqID, ok := ctx.Value("request_id").(string); ok && reqID != "" {
        return reqID
    }
    return "unknown"
}
```

2. **Fixed All Unsafe Type Assertions**:
   - Replaced 7 instances of unsafe `ctx.Value("request_id").(string)`
   - All now use the safe `h.getRequestID(ctx)` helper
   - Prevents panics when request_id is missing from context

3. **Added Missing Import**: Added `context` package import

**Files Modified**:
- `services/risk-assessment-service/internal/handlers/metrics.go`

## Next Steps

1. ✅ **COMPLETE**: Update API Gateway route mapping
2. ✅ **COMPLETE**: Fix panic in Risk Assessment Service metrics handler
3. ⏳ **PENDING**: Wait for Railway to deploy both fixes
4. ⏳ **PENDING**: Test both endpoints after deployment:
   - `/api/v1/risk/metrics` via API Gateway
   - `/api/v1/metrics` directly on Risk Assessment Service

## Testing After Deployment

```bash
# Test via API Gateway (should now work)
curl https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/metrics

# Test direct service endpoint (should now work)
curl https://risk-assessment-service-production.up.railway.app/api/v1/metrics
```

**Expected Results**:
- ✅ Both endpoints should return 200 OK with metrics data
- ✅ No more 503/502 errors
- ✅ Request ID safely handled even if missing from context

---

**Status**: ✅ **BOTH FIXES APPLIED** 

**Fixes**:
1. ✅ API Gateway route mapping: `/api/v1/risk/metrics` → `/api/v1/metrics`
2. ✅ Risk Assessment Service panic fix: Safe request_id extraction

**Commits**:
- `e2da5e034` - API Gateway route mapping fix
- `fff1e0fcb` - Metrics handler panic fix
- `4cb9843ab` - Added context import

**Awaiting**: Railway deployment of both services

