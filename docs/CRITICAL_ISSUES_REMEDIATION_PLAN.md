# Critical Issues Remediation Plan

**Date**: 2025-11-18  
**Based On**: Postman Test Results  
**Priority**: HIGH - Immediate Action Required

---

## Executive Summary

Postman test results revealed **4 critical issues** that need immediate attention:

1. 🔴 **Auth Login Route Returning 404** - Authentication completely broken
2. 🔴 **UUID Validation Not Working** - Invalid UUIDs being accepted
3. 🟡 **Register Endpoint Returning 500** - Registration failing
4. 🟡 **404 Handler Returning Plain Text** - Poor error handling

**Test Results**: 10/23 tests passed (43.5% pass rate)

---

## Issue #1: Auth Login Route Not Found (404) 🔴

### Problem
All login requests to `/api/v1/auth/login` are returning 404 Not Found, even though:
- Route is registered in `services/api-gateway/cmd/main.go` line 183
- Handler exists in `services/api-gateway/internal/handlers/gateway.go`
- Code was verified locally

### Affected Endpoints
- `POST /api/v1/auth/login` - All variants (valid, invalid, missing fields)

### Root Cause Hypotheses
1. **Code not deployed to Railway** - Most likely
2. Route being shadowed by PathPrefix route
3. Route registration order issue
4. Deployment error preventing route registration

### Investigation Steps

1. **Verify Deployment**
   ```bash
   # Check Railway deployment logs
   # Verify latest commit is deployed
   # Check for deployment errors
   ```

2. **Check Route Registration**
   - Verify route is registered before any PathPrefix that might catch `/auth/*`
   - Check if there's a PathPrefix `/auth` or `/api/v1` that might shadow it
   - Verify route is in the correct subrouter

3. **Check Railway Logs**
   - Look for route registration logs
   - Check for errors during startup
   - Verify handler initialization

4. **Test Route Directly**
   - Use curl to test route
   - Check if route responds at all
   - Verify route path matches exactly

### Fix Steps

1. **If Code Not Deployed**:
   - Trigger new deployment
   - Verify deployment completes successfully
   - Check deployment logs

2. **If Route Order Issue**:
   - Move auth routes before any PathPrefix routes
   - Ensure auth routes are in correct subrouter
   - Test route registration order

3. **If Route Shadowing**:
   - Check for conflicting PathPrefix routes
   - Ensure specific routes registered before PathPrefix
   - Verify subrouter configuration

### Verification
- [ ] Route responds with 200/401/400 (not 404)
- [ ] Login with valid credentials works
- [ ] Login with invalid credentials returns 401
- [ ] Login with missing fields returns 400

### Priority: 🔴 CRITICAL
**Estimated Fix Time**: 30 minutes - 2 hours

---

## Issue #2: UUID Validation Not Working 🔴

### Problem
Invalid UUIDs in `/api/v1/risk/indicators/{id}` are returning 200 OK with data instead of 400 Bad Request.

**Test Cases Failing**:
- `GET /api/v1/risk/indicators/invalid-id` → Returns 200 (should be 400)
- `GET /api/v1/risk/indicators/indicators` → Returns 200 (should be 400)

**Test Case Passing**:
- `GET /api/v1/risk/indicators/{valid-uuid}` → Returns 200 ✅

### Root Cause Hypotheses
1. UUID validation logic not being reached
2. Path transformation happening before validation
3. Route matching different path than expected
4. Validation function not being called

### Investigation Steps

1. **Check Handler Implementation**
   - Review `ProxyToRiskAssessment` in `services/api-gateway/internal/handlers/gateway.go`
   - Verify UUID validation is called before path transformation
   - Check path parsing logic (`parts[5]`)

2. **Add Logging**
   ```go
   // Add logging to UUID validation
   h.logger.Info("UUID validation check",
       zap.String("path", path),
       zap.String("merchantID", merchantID),
       zap.Bool("isValid", isValidUUID(merchantID)))
   ```

3. **Check Route Matching**
   - Verify route `/api/v1/risk/indicators/{id}` is matching correctly
   - Check if PathPrefix route is catching it first
   - Verify route registration order

4. **Test Path Parsing**
   - Log path parts to verify parsing
   - Check if `parts[5]` is correct index
   - Verify UUID extraction logic

### Fix Steps

1. **If Validation Not Called**:
   - Ensure validation happens before path transformation
   - Move validation earlier in handler
   - Add early return for invalid UUIDs

2. **If Path Parsing Issue**:
   - Fix path parsing logic
   - Verify correct index for UUID
   - Test with various path formats

3. **If Route Matching Issue**:
   - Ensure specific route registered before PathPrefix
   - Check route registration order
   - Verify route pattern matches correctly

### Code Review Needed
```go
// In ProxyToRiskAssessment handler
// Verify this logic is correct:
if strings.HasPrefix(path, "/api/v1/risk/indicators/") {
    parts := strings.Split(path, "/")
    if len(parts) >= 6 {
        merchantID := parts[5] // Should be the UUID
        if isValidUUID(merchantID) {
            // Transform path
        } else {
            // Return 400 - THIS SHOULD BE HAPPENING
            http.Error(w, "Invalid merchant ID format: expected UUID", http.StatusBadRequest)
            return
        }
    }
}
```

### Verification
- [ ] Invalid UUID returns 400 Bad Request
- [ ] Edge case "indicators" returns 400
- [ ] Valid UUID returns 200 OK
- [ ] Error message mentions UUID format

### Priority: 🔴 CRITICAL
**Estimated Fix Time**: 1-2 hours

---

## Issue #3: Register Endpoint Internal Server Error (500) 🟡

### Problem
Valid registration request to `/api/v1/auth/register` returns 500 Internal Server Error.

### Root Cause Hypotheses
1. Supabase connection issue
2. Missing environment variables
3. Database schema issue
4. Invalid request data handling
5. Supabase API error

### Investigation Steps

1. **Check Railway Logs**
   - Look for error stack traces
   - Check Supabase connection errors
   - Verify environment variables

2. **Verify Supabase Configuration**
   - Check `SUPABASE_URL` environment variable
   - Check `SUPABASE_API_KEY` environment variable
   - Verify Supabase project is accessible

3. **Test Supabase Connection**
   - Test Supabase client initialization
   - Verify Supabase API is reachable
   - Check Supabase project status

4. **Check Request Data**
   - Verify request body structure
   - Check required fields
   - Validate data format

### Fix Steps

1. **If Supabase Connection Issue**:
   - Verify environment variables
   - Test Supabase connection
   - Check Supabase project status

2. **If Database Schema Issue**:
   - Verify database tables exist
   - Check table schema matches code
   - Verify permissions

3. **If Request Data Issue**:
   - Add request validation
   - Improve error handling
   - Add logging

### Verification
- [ ] Valid registration returns 200/201
- [ ] User is created in Supabase
- [ ] Response includes user info
- [ ] No 500 errors

### Priority: 🟡 MEDIUM
**Estimated Fix Time**: 1-3 hours

---

## Issue #4: 404 Handler Returns Plain Text Instead of JSON 🟡

### Problem
404 handler returns plain text "404 page not found" instead of JSON error structure.

### Root Cause Hypotheses
1. `HandleNotFound` handler not being called
2. Default Go 404 handler being used
3. Handler not setting Content-Type header
4. Handler implementation issue

### Investigation Steps

1. **Check Handler Registration**
   ```go
   // Verify this is set:
   router.NotFoundHandler = http.HandlerFunc(gatewayHandler.HandleNotFound)
   ```

2. **Check Handler Implementation**
   - Verify handler sets Content-Type to application/json
   - Check handler returns JSON structure
   - Verify handler is being called

3. **Test Handler Directly**
   - Call handler function directly
   - Verify JSON output
   - Check Content-Type header

### Fix Steps

1. **If Handler Not Called**:
   - Verify `router.NotFoundHandler` is set
   - Ensure it's set after all routes
   - Check router configuration

2. **If Content-Type Not Set**:
   - Add `w.Header().Set("Content-Type", "application/json")`
   - Ensure it's set before writing response
   - Verify header is sent

3. **If Handler Implementation Issue**:
   - Review handler code
   - Ensure JSON encoding
   - Test handler output

### Code Review Needed
```go
// In HandleNotFound handler
func (h *GatewayHandler) HandleNotFound(w http.ResponseWriter, r *http.Request) {
    // MUST set Content-Type BEFORE writing
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotFound)
    
    // Return JSON structure
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": map[string]interface{}{
            "code": "NOT_FOUND",
            "message": "Route not found",
            // ...
        },
    })
}
```

### Verification
- [ ] 404 responses return JSON
- [ ] Content-Type is application/json
- [ ] Error structure matches expected format
- [ ] Error code is "NOT_FOUND"

### Priority: 🟡 MEDIUM
**Estimated Fix Time**: 30 minutes - 1 hour

---

## Implementation Priority

1. **First**: Fix Auth Login Route (404) - Blocks authentication
2. **Second**: Fix UUID Validation - Security/validation issue
3. **Third**: Fix Register 500 Error - Blocks registration
4. **Fourth**: Fix 404 Handler - User experience

---

## Testing After Fixes

1. **Re-run Postman Collection**
   - All tests should pass
   - Verify fixes work correctly
   - Check for regressions

2. **Manual Testing**
   - Test authentication flow
   - Test UUID validation
   - Test error handling

3. **Monitor Railway Logs**
   - Watch for errors
   - Check response times
   - Verify route matching

---

## Success Criteria

- [ ] All 23 Postman tests pass
- [ ] Auth login route works (200/401/400, not 404)
- [ ] UUID validation works (400 for invalid, 200 for valid)
- [ ] Register endpoint works (200/201, not 500)
- [ ] 404 handler returns JSON (not plain text)

---

**Last Updated**: 2025-11-18  
**Status**: Ready for Implementation

