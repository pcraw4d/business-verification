# Root Cause Analysis

**Date**: 2025-11-18  
**Status**: Complete Analysis of All Issues  
**Source**: Code Review, Test Results, Handler Implementation Review

---

## Analysis Methodology

For each issue, we reviewed:
1. Code implementation in handlers
2. Route registration in main.go
3. Test results and error responses
4. Handler logic flow
5. Potential deployment issues

---

## Issue #1: Auth Login Route Returning 404

### Code Review

**Route Registration** (`services/api-gateway/cmd/main.go` line 183):
```go
api.HandleFunc("/auth/login", gatewayHandler.HandleAuthLogin).Methods("POST", "OPTIONS")
```

**Handler Implementation** (`services/api-gateway/internal/handlers/gateway.go` lines 760-853):
- Handler exists and is correctly implemented
- Validates request body
- Calls Supabase client
- Returns proper JSON responses

### Root Cause Analysis

**Most Likely Cause**: Code not deployed to Railway

**Evidence**:
1. Code exists and is correct locally
2. Route is registered correctly
3. Handler is implemented correctly
4. All login requests return 404 (not 500, which would indicate handler is called but errors)
5. 404 suggests route doesn't exist in deployed version

**Alternative Causes** (less likely):
1. Route being shadowed by PathPrefix - Unlikely, auth routes are registered after PathPrefix routes
2. Route registration order issue - Unlikely, auth routes are in correct subrouter
3. Deployment error - Possible, but would show in Railway logs

### Verification Steps

1. Check Railway deployment logs for latest commit
2. Verify code is deployed (check git commit hash in Railway)
3. Check Railway logs for route registration
4. Test route directly with curl to confirm 404

### Fix Strategy

1. **If code not deployed**: Trigger new deployment, verify deployment completes
2. **If route order issue**: Move auth routes before any PathPrefix (already correct)
3. **If route shadowing**: Check for conflicting routes (none found)

---

## Issue #2: UUID Validation Not Working

### Code Review

**Handler Implementation** (`services/api-gateway/internal/handlers/gateway.go` lines 534-557):
```go
} else if strings.HasPrefix(path, "/api/v1/risk/indicators/") {
    parts := strings.Split(path, "/")
    if len(parts) >= 6 {
        merchantID := parts[5] // /api/v1/risk/indicators/{id} - index 5 is the ID
        if isValidUUID(merchantID) {
            path = fmt.Sprintf("/api/v1/risk/predictions/%s", merchantID)
        } else {
            http.Error(w, "Invalid merchant ID format: expected UUID", http.StatusBadRequest)
            return
        }
    }
}
```

**UUID Validation Function** (lines 24-30):
```go
func isValidUUID(uuid string) bool {
    if uuid == "" {
        return false
    }
    return uuidPattern.MatchString(strings.ToLower(uuid))
}
```

**Route Registration** (`services/api-gateway/cmd/main.go` lines 167-172):
```go
api.HandleFunc("/risk/assess", gatewayHandler.ProxyToRiskAssessment).Methods("POST", "OPTIONS")
api.HandleFunc("/risk/benchmarks", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
api.HandleFunc("/risk/predictions/{merchant_id}", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
api.HandleFunc("/risk/indicators/{id}", gatewayHandler.ProxyToRiskAssessment).Methods("GET", "OPTIONS")
api.PathPrefix("/risk").HandlerFunc(gatewayHandler.ProxyToRiskAssessment)
```

### Root Cause Analysis

**Most Likely Cause**: PathPrefix route catching requests before specific route

**Evidence**:
1. Code logic is correct - validation exists and should work
2. Route `/risk/indicators/{id}` is registered before PathPrefix
3. But PathPrefix `/risk` might be matching first in some cases
4. Test shows invalid UUIDs return 200 with data, suggesting request reaches backend

**Alternative Causes**:
1. Route matching issue - PathPrefix might be evaluated differently
2. Handler not being called - Unlikely, since valid UUIDs work
3. Path parsing issue - Unlikely, code looks correct

### Verification Steps

1. Add logging to UUID validation to see if it's called
2. Check Railway logs for path transformation messages
3. Verify route registration order (already correct in code)
4. Test with various invalid UUID formats

### Fix Strategy

1. **Ensure route order is correct**: Verify specific route is registered before PathPrefix (already correct)
2. **Add logging**: Add debug logging to see if validation is reached
3. **Check route matching**: Verify gorilla/mux route matching behavior
4. **Test path parsing**: Verify `parts[5]` is correct index for UUID

**Potential Fix**: The issue might be that gorilla/mux's PathPrefix matches before specific routes in some cases. We may need to ensure the specific route handler is called, or move validation earlier in the handler.

---

## Issue #3: Register Endpoint Returning 500

### Code Review

**Handler Implementation** (`services/api-gateway/internal/handlers/gateway.go` lines 659-759):
- Handler exists and is correctly implemented
- Validates request body
- Calls Supabase client to create user
- Returns proper JSON responses

### Root Cause Analysis

**Most Likely Cause**: Supabase connection or configuration issue

**Evidence**:
1. Handler code is correct
2. Validation works (missing fields returns 400)
3. Invalid email returns 400
4. Valid registration returns 500 (server error)
5. Suggests Supabase call is failing

**Possible Causes**:
1. Supabase connection issue
2. Missing environment variables (SUPABASE_URL, SUPABASE_API_KEY)
3. Database schema issue (table doesn't exist or wrong schema)
4. Supabase API error
5. Invalid request data handling

### Verification Steps

1. Check Railway logs for error stack traces
2. Verify Supabase environment variables are set
3. Test Supabase connection
4. Check Supabase project status
5. Verify database tables exist

### Fix Strategy

1. **Check Railway logs**: Look for error details
2. **Verify Supabase config**: Check environment variables
3. **Test Supabase connection**: Verify client can connect
4. **Check database schema**: Verify tables exist and schema matches
5. **Add error logging**: Improve error messages to identify issue

---

## Issue #4: 404 Handler Returning Plain Text

### Code Review

**Handler Implementation** (`services/api-gateway/internal/handlers/gateway.go` lines 856-913):
- Handler sets Content-Type to application/json (line 909)
- Handler returns JSON structure (lines 880-911)
- Handler code is correct

**Route Registration** (`services/api-gateway/cmd/main.go` line 187):
```go
router.NotFoundHandler = http.HandlerFunc(gatewayHandler.HandleNotFound)
```

### Root Cause Analysis

**Most Likely Cause**: NotFoundHandler not being called by gorilla/mux

**Evidence**:
1. Handler code is correct
2. NotFoundHandler is set on main router
3. But response is plain text "404 page not found" (default Go handler)
4. Suggests gorilla/mux isn't calling NotFoundHandler

**Possible Causes**:
1. NotFoundHandler only works for routes that don't match any route pattern
2. Subrouter routes might not trigger NotFoundHandler
3. PathPrefix routes might prevent NotFoundHandler from being called
4. Handler registration order issue

### Verification Steps

1. Test with route that definitely doesn't match any pattern
2. Check if NotFoundHandler is called (add logging)
3. Verify gorilla/mux NotFoundHandler behavior
4. Test with different route patterns

### Fix Strategy

1. **Verify NotFoundHandler behavior**: Test if it's called for unmatched routes
2. **Add logging**: Add logging to handler to see if it's called
3. **Check gorilla/mux documentation**: Verify NotFoundHandler usage
4. **Alternative approach**: Use middleware to catch 404s, or custom 404 route

**Potential Fix**: gorilla/mux's NotFoundHandler might not work as expected with subrouters. We may need to use a catch-all route or middleware to handle 404s.

---

## Issue #5-8: Service Health and Backend Issues

### Root Cause Analysis

**Frontend Service 502**:
- Service accessible at root `/`
- `/health` endpoint returns 502
- **Root Cause**: Service may not have `/health` endpoint, or endpoint is broken
- **Action**: Verify if health endpoint is required

**Pipeline, BI, Monitoring Services 502**:
- All three services return 502 on `/health`
- **Root Cause**: Services not deployed, down, or health endpoint at different path
- **Action**: Check Railway dashboard for service status

**Session Routes 502**:
- Routes match correctly
- Path transformations work
- Backend (Frontend Service) returns 502
- **Root Cause**: Frontend Service unavailable or session endpoints not implemented
- **Action**: Check Frontend Service status and logs

**BI Service 500**:
- Route matches correctly
- Path transformation works
- Backend returns 500
- **Root Cause**: BI Service erroring on requests
- **Action**: Check BI Service logs for errors

---

## Summary of Root Causes

### Code Issues (Fixable)
1. **Auth Login 404**: Code not deployed (most likely)
2. **UUID Validation**: Route matching issue (PathPrefix vs specific route)
3. **404 Handler**: NotFoundHandler not being called by gorilla/mux

### Configuration Issues (Fixable)
1. **Register 500**: Supabase configuration or connection issue

### Service Issues (Investigation Required)
1. **Frontend Service**: Health endpoint missing or broken
2. **Pipeline/BI/Monitoring**: Services not deployed or down
3. **Session Routes**: Backend service unavailable
4. **BI Service**: Service erroring

---

## Recommended Fix Order

1. **First**: Fix Auth Login 404 (deploy code or fix route)
2. **Second**: Fix UUID Validation (verify route matching, add logging)
3. **Third**: Fix Register 500 (check Supabase config)
4. **Fourth**: Fix 404 Handler (verify NotFoundHandler behavior)
5. **Fifth**: Investigate service health issues (check Railway dashboard)

---

**Last Updated**: 2025-11-18  
**Status**: Ready for Remediation Planning

