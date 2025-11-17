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

## Next Steps

1. Update API Gateway route mapping
2. Test the fix
3. Investigate 502 error on `/api/v1/metrics` endpoint
4. Document the fix

---

**Status**: ⚠️ **INVESTIGATION COMPLETE** - Route mapping issue identified, fix required

