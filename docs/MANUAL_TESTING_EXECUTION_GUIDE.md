# Manual Testing Execution Guide

**Date**: 2025-11-18  
**Status**: Ready for Execution  
**Prerequisites**: curl, Postman, or browser with DevTools

---

## Quick Start

Since automated testing tools are not available in this environment, use one of these methods:

### Option 1: Using curl (Recommended)

If you have curl installed, use the commands below.

### Option 2: Using Postman

Import the test collection (create from commands below).

### Option 3: Using Browser DevTools

Open browser DevTools → Network tab and test from frontend.

---

## Phase 3.1: Authentication Routes Testing

### Test 1: Valid Registration

```bash
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser_'$(date +%s)'@example.com",
    "password": "TestPassword123!",
    "username": "testuser",
    "first_name": "Test",
    "last_name": "User"
  }' \
  -v
```

**Expected**: 
- Status: 201 Created or 200 OK
- Response includes user info
- No errors

**Document Result**: Record status code and response

---

### Test 2: Missing Required Fields

```bash
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -v
```

**Expected**: 
- Status: 400 Bad Request
- Error message indicates missing password

**Document Result**: Record status code and error message

---

### Test 3: Invalid Email Format

```bash
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "not-an-email",
    "password": "TestPassword123!"
  }' \
  -v
```

**Expected**: 
- Status: 400 Bad Request
- Error message indicates invalid email format

**Document Result**: Record status code and error message

---

### Test 4: Valid Login

```bash
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "TestPassword123!"
  }' \
  -v
```

**Expected**: 
- Status: 200 OK
- Response includes `token` and `user` fields
- Token is valid JWT format

**Note**: Use email from Test 1 if user was created

**Document Result**: Record status code and verify token presence

---

### Test 5: Invalid Credentials

```bash
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "WrongPassword"
  }' \
  -v
```

**Expected**: 
- Status: 401 Unauthorized
- Error message doesn't reveal which field is wrong

**Document Result**: Record status code and error message

---

### Test 6: Missing Fields (Login)

```bash
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -v
```

**Expected**: 
- Status: 400 Bad Request
- Error message indicates missing password

**Document Result**: Record status code and error message

---

## Phase 3.2: UUID Validation Testing

### Test 1: Invalid UUID Format

```bash
curl -X GET "https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/indicators/invalid-id" \
  -v
```

**Expected**: 
- Status: 400 Bad Request
- Error message: "Invalid merchant ID format: expected UUID"

**Check Railway Logs**: Look for "Invalid merchant ID format in risk indicators endpoint"

**Document Result**: Record status code and error message

---

### Test 2: "indicators" as ID (Edge Case)

```bash
curl -X GET "https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/indicators/indicators" \
  -v
```

**Expected**: 
- Status: 400 Bad Request
- UUID validation catches this before path transformation

**Document Result**: Record status code

---

### Test 3: Valid UUID (If Available)

```bash
# Replace with actual merchant UUID from your database
VALID_UUID="550e8400-e29b-41d4-a716-446655440000"

curl -X GET "https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/indicators/${VALID_UUID}" \
  -v
```

**Expected**: 
- Status: 200 OK (or appropriate response from risk service)
- Request reaches risk service
- No UUID validation error

**Document Result**: Record status code

---

## Phase 3.3: CORS Configuration Testing

### Test 1: Preflight Request (OPTIONS)

```bash
curl -X OPTIONS https://api-gateway-service-production-21fd.up.railway.app/api/v1/auth/register \
  -H "Origin: https://frontend-service-production-b225.up.railway.app" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -v
```

**Expected Headers**:
- `Access-Control-Allow-Origin: https://frontend-service-production-b225.up.railway.app` (NOT `*`)
- `Access-Control-Allow-Credentials: true`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`
- Status: 200 OK

**Verify**: 
- ✅ Specific origin (not wildcard)
- ✅ Credentials allowed

**Document Result**: Record headers and status

---

### Test 2: Cross-Origin Request from Browser

1. Open browser: `https://frontend-service-production-b225.up.railway.app`
2. Open DevTools (F12) → Network tab
3. Try to register/login from frontend
4. Check Network tab:
   - Look for requests to `/api/v1/auth/register` or `/api/v1/auth/login`
   - Verify no CORS errors in console
   - Check response headers include CORS headers

**Expected**: No CORS errors, requests succeed

**Document Result**: Record any CORS errors

---

## Phase 4: Route Precedence Testing

### Test Merchant Route Precedence

**Test 1: Specific Sub-route**
```bash
curl -X GET "https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants/test-id-123/analytics" \
  -v
```

**Expected**: 200 OK  
**Verify**: Route matches specific handler, not PathPrefix catch-all

---

**Test 2: General Endpoint**
```bash
curl -X GET "https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants/analytics" \
  -v
```

**Expected**: 200 OK  
**Verify**: Matches `/merchants/analytics`, not `/merchants/{id}` with id="analytics"

---

**Test 3: Base Route**
```bash
curl -X GET "https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants" \
  -v
```

**Expected**: 200 OK with merchant list  
**Verify**: Matches base `/merchants` route

---

## Phase 5: Path Transformation Testing

### Test Risk Assessment Transformations

**Test 1: Risk Assess**
```bash
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/assess \
  -H "Content-Type: application/json" \
  -d '{"merchant_id": "test-123"}' \
  -v
```

**Check Railway Logs**: Verify path transformed to `/api/v1/assess`

---

**Test 2: Risk Metrics**
```bash
curl -X GET https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/metrics \
  -v
```

**Check Railway Logs**: Verify path transformed to `/api/v1/metrics`

---

## Phase 6: Error Handling Testing

### Test 404 Handler

```bash
curl -X GET https://api-gateway-service-production-21fd.up.railway.app/api/v1/nonexistent-route \
  -v
```

**Expected**: 
- Status: 404 Not Found
- Response includes:
  - Error code: "NOT_FOUND"
  - Helpful message
  - Suggestions (if applicable)
  - Available endpoints list

**Check Railway Logs**: Verify route logged with context

**Document Result**: Record response structure

---

## Test Results Template

Update `docs/MANUAL_TEST_RESULTS.md` with results:

```markdown
### Test Name
- **Status**: PASS/FAIL/SKIP
- **Date**: 2025-11-18
- **Status Code**: XXX
- **Response**: [Brief summary]
- **Notes**: [Any issues or observations]
```

---

## Railway Logs Analysis

After running tests, check Railway logs for:

1. **UUID Validation Errors**:
   - Look for: "Invalid merchant ID format in risk indicators endpoint"
   - Should appear for invalid UUID tests

2. **Route Not Found Errors**:
   - Look for: "Route not found" with method, path, query
   - Should appear for 404 tests

3. **Path Transformations**:
   - Check logs for path transformation messages
   - Verify correct backend routing

4. **CORS Errors**:
   - Should NOT appear if CORS is configured correctly

---

## Next Steps After Testing

1. **Document All Results** in `docs/MANUAL_TEST_RESULTS.md`
2. **Create Issue List** for any failures
3. **Prioritize Fixes** (Critical, High, Medium, Low)
4. **Retest** after fixes
5. **Update Final Report** with test outcomes

---

## Troubleshooting

**Issue**: Cannot connect to API
- **Check**: Service is deployed and healthy
- **Verify**: URL is correct
- **Solution**: Check Railway dashboard for service status

**Issue**: CORS errors
- **Check**: `CORS_ALLOWED_ORIGINS` environment variable
- **Verify**: Frontend origin matches exactly
- **Solution**: Update environment variable in Railway

**Issue**: 404 on valid routes
- **Check**: Route registration order in code
- **Verify**: Specific routes before PathPrefix
- **Solution**: Review route registration in `main.go`

---

**Ready to Execute**: Copy commands above and run in your terminal with curl, or use Postman/browser DevTools.

