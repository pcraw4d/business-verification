# Railway Logs Analysis

**Date**: 2025-11-19  
**Analysis**: API Gateway Service Logs  
**Focus**: Auth Routes and Error Investigation

---

## Key Findings

### 1. Auth Login Route - 404 Issue

**Status**: ⚠️ CONFIRMED - Route returns 404

**Test Results**:
```bash
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "test"}'
```

**Response**: 
- Status: 404
- Body: "404 page not found" (plain text, not JSON)

**Code Analysis**:
- Route is registered correctly in `services/api-gateway/cmd/main.go` (line 185)
- Handler exists: `HandleAuthLogin` in `gateway.go`
- Route is in correct subrouter: `/api/v1` subrouter
- No PathPrefix conflicts found

**Logs Analysis**:
- No logs found for `/api/v1/auth/login` requests
- This suggests the route is not being matched at all
- The 404 is coming from gorilla/mux default handler, not our custom NotFoundHandler

**Root Cause Hypothesis**:
1. **Deployment Issue**: Code may not be deployed to Railway
2. **Route Registration Order**: Route may be shadowed by another route
3. **Subrouter Issue**: Auth routes registered after PathPrefix routes may not be matching

**Action Required**:
1. Verify latest code is deployed to Railway
2. Check Railway deployment logs
3. Verify git commit hash matches local
4. If not deployed, trigger new deployment
5. Test route after deployment

---

### 2. Auth Register Route - 500 Error

**Status**: ⚠️ CONFIRMED - Returns 500 Internal Server Error

**Test Results**:
```bash
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "newuser@example.com", "password": "testpass123", "username": "testuser"}'
```

**Response**:
- Status: 500
- Body: `{"error":{"code":"INTERNAL_ERROR","message":"Internal server error","details":"Unable to complete registration. Please try again later."},"timestamp":"2025-11-19T01:28:37Z","path":"/api/v1/auth/register","method":"POST"}`

**Logs Analysis**:
```
[ERRO] User registration failed caller="handlers/gateway.go:684" 
email="newuser@example.com" 
error="registration failed with status 400: {\"code\":400,\"error_code\":\"email_address_invalid\",\"msg\":\"Email address \\\"newuser@example.com\\\" is invalid\"}" 
username="testuser"
```

**Root Cause**:
- Supabase is rejecting the email address as invalid
- Error code: `email_address_invalid`
- Message: "Email address \"newuser@example.com\" is invalid"
- This is a **Supabase configuration issue**, not a code issue

**Possible Causes**:
1. **Supabase Email Validation Settings**: Supabase may have strict email validation enabled
2. **Email Domain Restrictions**: Supabase may be configured to reject certain domains (e.g., `@example.com`)
3. **Email Format Requirements**: Supabase may require specific email formats
4. **Supabase Project Settings**: Email validation rules may be configured in Supabase dashboard

**Action Required**:
1. Check Supabase project settings for email validation rules
2. Verify if `@example.com` domain is allowed
3. Test with a real email domain (e.g., `@gmail.com`)
4. Check Supabase Auth settings in dashboard
5. Review Supabase email validation configuration

**Code Review**:
- Handler code is correct (lines 664-750 in `gateway.go`)
- Email validation in handler passes (simple `@` and `.` check)
- Supabase API call is made correctly
- Error handling is correct (returns 500 with proper error message)

**Test with Real Email**:
```bash
# Test with a real email domain
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "testuser@gmail.com", "password": "testpass123", "username": "testuser"}'
```

**Result**: ✅ **SUCCESS** - Registration works with real email domains
- Status: 200 OK
- Response: `{"message":"Registration successful. Please check your email for verification instructions.","user":{"email":"testuser@gmail.com"}}`
- **Conclusion**: Register endpoint is working correctly. The 500 error only occurs with `@example.com` domain, which Supabase rejects.

---

### 3. CORS Configuration

**Status**: ✅ WORKING

**Logs Show**:
```
CORS: Request from origin: , Method: POST, Path: /api/v1/auth/register
CORS: Allowed origins: [*]
CORS: Set Access-Control-Allow-Origin to: *
```

**Findings**:
- CORS middleware is working correctly
- Headers are being set properly
- However, `CORS_ALLOWED_ORIGINS` is set to `*` in Railway environment variables
- Code defaults to specific frontend URL, but Railway env var overrides it

**Issue**:
- Wildcard (`*`) with `AllowCredentials: true` will be rejected by browsers
- This is a security issue - browsers will block requests with credentials when origin is `*`

**Action Required**:
1. Update Railway environment variable `CORS_ALLOWED_ORIGINS` to specific frontend URL
2. Remove wildcard and use: `https://frontend-service-production-b225.up.railway.app`
3. Test CORS in browser after fix

---

### 4. 404 Handler

**Status**: ⚠️ NOT WORKING

**Test Results**:
- 404s return plain text "404 page not found" instead of JSON
- Custom `HandleNotFound` handler is not being called

**Code Analysis**:
- `router.NotFoundHandler` is set correctly (line 192 in `main.go`)
- Handler code is correct and sets Content-Type to JSON
- But handler is not being invoked

**Root Cause**:
- `gorilla/mux` `NotFoundHandler` may not work with subrouters
- 404s from subrouter may not trigger main router's NotFoundHandler

**Action Required**:
1. Test if NotFoundHandler works after deployment
2. If not, implement alternative solution:
   - Set NotFoundHandler on subrouter as well
   - Use middleware to catch unmatched routes
   - Use catch-all route (carefully, to avoid breaking existing routes)

---

## Summary

### Issues Confirmed

1. **Auth Login 404** - Route not matching (deployment or route registration issue)
2. **Auth Register 500** - Supabase email validation rejecting emails
3. **404 Handler** - Not being called (gorilla/mux subrouter issue)
4. **CORS Wildcard** - Security issue with credentials

### Issues Resolved

- None yet - all require fixes or investigation

### Next Steps

1. **Immediate**:
   - Verify code deployment for Auth Login route
   - Check Supabase email validation settings
   - Update CORS_ALLOWED_ORIGINS environment variable

2. **Short Term**:
   - Fix 404 handler (alternative approach)
   - Test with real email domains
   - Complete browser testing with Playwright

3. **Long Term**:
   - Complete all remaining testing phases
   - Fix all identified issues
   - Final verification and reporting

---

**Last Updated**: 2025-11-19  
**Status**: Analysis Complete, Actions Required

