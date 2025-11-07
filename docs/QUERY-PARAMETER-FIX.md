# Query Parameter Fix

**Date**: January 2025  
**Issue**: Query parameters appearing in response values  
**Status**: ✅ **FIXED**

---

## Issue

**Problem**: Query parameters were appearing in response values instead of being parsed correctly.

**Example**:
```json
{
  "industry_code": "5411?mcc=5411",  // ❌ Should be "5411"
  "mcc": "5411?mcc=5411"              // ❌ Should be "5411"
}
```

**Root Cause**: Query parameters were being added to the path string in `ProxyToRiskAssessment`, and then `proxyRequest` was also adding them to the target URL. This caused the query string to be included as part of the path, which then got parsed incorrectly by the Risk Assessment Service.

---

## Fix Applied

### Before (Incorrect)
```go
// In ProxyToRiskAssessment
path := "/api/v1/risk/benchmarks"
if r.URL.RawQuery != "" {
    path += "?" + r.URL.RawQuery  // ❌ Adding query to path
}
h.proxyRequest(w, r, targetURL, path)

// In proxyRequest
target := targetURL + path
if r.URL.RawQuery != "" {
    target += "?" + r.URL.RawQuery  // ❌ Adding query again!
}
// Result: /api/v1/risk/benchmarks?mcc=5411?mcc=5411
```

### After (Correct)
```go
// In ProxyToRiskAssessment
path := "/api/v1/risk/benchmarks"
// ✅ Do NOT add query parameters here - proxyRequest handles them
h.proxyRequest(w, r, targetURL, path)

// In proxyRequest
target := targetURL + path
if r.URL.RawQuery != "" {
    target += "?" + r.URL.RawQuery  // ✅ Only add once
}
// Result: /api/v1/risk/benchmarks?mcc=5411
```

---

## Impact

- **Before**: Query parameters appeared in response values
- **After**: Query parameters are parsed correctly, response values are clean

---

## Status

✅ **Fix Committed and Pushed**  
⏳ **Awaiting Railway Deployment**

Once Railway redeploys the API Gateway, query parameters will be parsed correctly and response values will be clean.

---

## Files Modified

- `services/api-gateway/internal/handlers/gateway.go` - Removed duplicate query parameter handling

