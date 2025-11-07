# API Gateway Proxy Path Fix

**Date**: January 2025  
**Issue**: API Gateway returning 404 for risk endpoints  
**Status**: ✅ **FIXED**

---

## Issue

**Problem**: API Gateway was returning 404 for `/api/v1/risk/benchmarks` and `/api/v1/risk/predictions`

**Root Cause**: The proxy was stripping `/api/v1/risk` from the path, sending only `/benchmarks` to the Risk Assessment Service. However, the Risk Assessment Service expects the full `/api/v1/risk/benchmarks` path.

---

## Fix Applied

### Before (Incorrect)
```go
// Extract the path after /api/v1/risk/
path := strings.TrimPrefix(r.URL.Path, "/api/v1/risk")
// Result: /api/v1/risk/benchmarks → /benchmarks ❌
```

### After (Correct)
```go
// Extract the path after /api/v1/
// The Risk Assessment Service expects /api/v1/risk/* paths
path := strings.TrimPrefix(r.URL.Path, "/api/v1")
// Result: /api/v1/risk/benchmarks → /risk/benchmarks

// Ensure path starts with /api/v1
if !strings.HasPrefix(path, "/api/v1") {
    path = "/api/v1" + path
}
// Result: /risk/benchmarks → /api/v1/risk/benchmarks ✅
```

---

## Path Flow

### Request Flow
1. **Client** → `/api/v1/risk/benchmarks?mcc=5411`
2. **API Gateway** receives: `/api/v1/risk/benchmarks`
3. **API Gateway** strips `/api/v1` → `/risk/benchmarks`
4. **API Gateway** ensures `/api/v1` prefix → `/api/v1/risk/benchmarks`
5. **API Gateway** proxies to Risk Service: `/api/v1/risk/benchmarks?mcc=5411`
6. **Risk Service** matches route: `/api/v1/risk/benchmarks` ✅

---

## Verification

**Direct to Risk Service**: ✅ Working
```bash
curl "https://risk-assessment-service-production.up.railway.app/api/v1/risk/benchmarks?mcc=5411"
# Returns: 200 OK with benchmarks data
```

**Through API Gateway**: ⏳ Will work after deployment
```bash
curl "https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/benchmarks?mcc=5411"
# Should return: 200 OK with benchmarks data
```

---

## Status

✅ **Fix Committed and Pushed**  
⏳ **Awaiting Railway Deployment**

Once Railway redeploys the API Gateway, all endpoints should work correctly through the gateway.

---

## Files Modified

- `services/api-gateway/internal/handlers/gateway.go` - Fixed proxy path logic

