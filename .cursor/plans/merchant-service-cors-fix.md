# Merchant Service CORS Fix

## Problem
The merchant-service was setting CORS headers (`Access-Control-Allow-Origin: *`) for all requests, including internal requests from the API Gateway. This caused duplicate CORS headers in the final response:
- `Access-Control-Allow-Origin: http://localhost:3000` (from API Gateway CORS middleware)
- `Access-Control-Allow-Origin: *` (from merchant-service CORS middleware)

## Root Cause
The merchant-service CORS middleware was setting CORS headers for ALL requests, including:
1. Direct external requests (with Origin header) - CORS needed
2. Internal requests from API Gateway (no Origin header) - CORS NOT needed (API Gateway handles it)

When the API Gateway proxies requests to merchant-service, it doesn't include an Origin header (internal request), so merchant-service was defaulting to `*`.

## Solution
Modified `services/merchant-service/cmd/main.go` CORS middleware to:
1. **Skip CORS headers for internal requests** (no Origin header)
   - These are requests from API Gateway
   - API Gateway will handle CORS, so merchant-service shouldn't set headers
2. **Set CORS headers only for external requests** (with Origin header)
   - These are direct requests to merchant-service
   - Only set headers when Origin is present

## Code Changes
```go
// Before: Always set CORS headers
if origin != "" {
    w.Header().Set("Access-Control-Allow-Origin", origin)
} else {
    w.Header().Set("Access-Control-Allow-Origin", "*") // ❌ This caused duplicates
}

// After: Skip CORS for internal requests
if origin == "" {
    // Internal request - skip CORS, let API Gateway handle it
    next.ServeHTTP(w, r)
    return
}
// External request - set CORS headers
w.Header().Set("Access-Control-Allow-Origin", origin)
```

## Testing
Once merchant-service is restarted with proper environment variables:
1. Test direct request to merchant-service (should have CORS headers)
2. Test request through API Gateway (should have single CORS header from API Gateway only)

## Status
- ✅ Code fix applied
- ⏳ Merchant-service needs to be restarted with proper Supabase environment variables
- ⏳ CORS test pending after merchant-service restart

## Next Steps
1. Ensure merchant-service has proper environment variables (SUPABASE_URL, SUPABASE_ANON_KEY, etc.)
2. Restart merchant-service
3. Test CORS through API Gateway
4. Verify single `Access-Control-Allow-Origin` header in response

