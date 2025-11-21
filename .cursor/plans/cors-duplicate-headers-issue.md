# CORS Duplicate Headers Issue

## Problem
The API Gateway is returning duplicate `Access-Control-Allow-Origin` headers:
- `Access-Control-Allow-Origin: http://localhost:3000` (from CORS middleware)
- `Access-Control-Allow-Origin: *` (from upstream merchant-service)

## Root Cause
1. The CORS middleware sets the header to the requesting origin (`http://localhost:3000`)
2. The upstream merchant-service also sets CORS headers (`*`)
3. Even though we're trying to skip CORS headers when copying from upstream, they're still getting through

## Attempted Fixes
1. ✅ Removed duplicate CORS middleware from subrouters
2. ✅ Added explicit header deletion in CORS middleware
3. ✅ Added CORS header deletion in proxy function before copying headers
4. ✅ Added case-insensitive header matching
5. ✅ Added second deletion before writing response
6. ❌ Still seeing duplicate headers

## Next Steps
The issue appears to be that the upstream merchant-service is setting CORS headers that are somehow getting through despite our skip logic. We may need to:
1. Fix the merchant-service to not set CORS headers (let API Gateway handle it)
2. Or ensure the proxy function completely strips all CORS headers before the response is written
3. Or use a response writer wrapper that intercepts and removes duplicate headers

## Current Status
- Backend restarted with fixes
- CORS test still failing with duplicate headers
- Need to investigate why headers are getting through despite skip logic

