---
name: Route Testing and Remediation Plan
overview: ""
todos:
  - id: 084d7857-b8fb-4482-b793-5bfba2e385b0
    content: Review all modified files for syntax errors, route registration order, UUID validation, CORS config, auth login handler, 404 handler, and port configs
    status: pending
  - id: 7b37bade-603e-4628-b231-76b09829390e
    content: Build all services locally (API Gateway, Merchant, Service Discovery, Frontend) and verify no compilation errors
    status: pending
  - id: 870fb212-992d-420d-870a-27a3c2b42ea9
    content: Test health endpoints for all 9 services and verify all return 200 OK with proper status
    status: pending
  - id: 07e70402-6c0c-4504-a894-b1ed5f4d6259
    content: Verify Merchant Service uses port 8080 (not 8082) and Service Discovery uses port 8080 (not 8086) by checking Railway logs and health checks
    status: pending
  - id: 6c591b62-c601-430d-8258-f5a47681ac4d
    content: Test POST /api/v1/auth/register and POST /api/v1/auth/login with valid data, invalid credentials, and error scenarios
    status: pending
  - id: eb664419-bf0d-4e73-9e65-fcc0213197f7
    content: Test GET /api/v1/risk/indicators/{uuid} with valid UUID, invalid UUID, and edge cases to verify UUID validation and path transformation
    status: pending
  - id: 34f01ce7-42bb-4e51-9145-c92816717f26
    content: Test OPTIONS preflight requests and cross-origin requests to verify CORS uses specific frontend origin (not wildcard) with credentials
    status: pending
  - id: 4d8f75f2-c59f-4f55-9b91-69b7d38225eb
    content: Test merchant and risk route precedence to verify specific routes match before PathPrefix catch-all routes
    status: pending
  - id: ecf2b0d1-4afa-40b7-bdbc-8827b3c6366a
    content: Test all path transformations (risk assess, risk metrics, risk indicators, sessions, BI) to verify correct path forwarding to backend services
    status: pending
  - id: bf630ca9-b5be-4f3d-9726-5c451273b43f
    content: Test 404 handler with invalid routes to verify helpful error messages, suggestions, and logging. Test UUID validation error responses
    status: pending
  - id: c65ce29f-aa66-4d5a-95ec-9c25126161da
    content: Test frontend API calls in browser to verify auth endpoints use /api/v1/auth/* paths, API base URL configured correctly, and no CORS errors
    status: pending
  - id: 6c96a290-e8d4-445c-a82b-afd5c6f9286c
    content: Verify NEXT_PUBLIC_API_BASE_URL set for Frontend, CORS_ALLOWED_ORIGINS set correctly for API Gateway, and risk-assessment-service restart retries is 10
    status: pending
  - id: cb97c948-91e7-4858-a858-369bfa1f0222
    content: Test all routes from comprehensive analysis report including health, classification, merchant, risk, auth, session, BI, and compliance routes
    status: pending
  - id: 0ec90f94-d6a3-4aa9-825d-72794bcf9d0c
    content: Verify response times meet targets (health <1s, classification <5s, merchant <2s, risk <10s, auth <2s) and check logging for route matching and transformations
    status: pending
  - id: ee565c90-4f46-4a62-a7a8-c1211104b39a
    content: Test input validation (SQL injection, XSS, path traversal), authentication requirements, and verify invalid UUIDs are rejected
    status: pending
  - id: 1f792914-4337-4b2f-8326-dca5c3757b9d
    content: Verify all previously working routes still work, no new 404 errors introduced, path transformations still correct, and CORS still works
    status: pending
  - id: 2286131d-592a-4061-b451-e8e522b612db
    content: Test end-to-end user flows (registration→login→merchant operations) and analyze Railway logs for errors (UUID parsing, route not found, CORS, path transformation)
    status: pending
  - id: 9bee3a25-01f4-4654-9f13-de1678bbc180
    content: Document all test results, track issues found, create remediation plans, and generate final status report for all routes
    status: pending
---

# Route Testing and Remediation Plan

## Overview

This plan provides a systematic approach to test and verify all route fixes implemented in the API Routes Fixes Implementation Plan. It includes pre-deployment verification, post-deployment testing, route-by-route validation, error scenario testing, and remediation procedures.

## Phase 1: Pre-Deployment Verification

### 1.1 Code Review and Static Analysis

**Objective**: Verify all code changes are correct before deployment

**Tasks**:

- [ ] Review all modified files for syntax errors
- [ ] Verify route registration order matches guidelines
- [ ] Check UUID validation logic is correct
- [ ] Verify CORS configuration changes
- [ ] Confirm auth login handler implementation
- [ ] Validate 404 handler implementation
- [ ] Check port configuration changes

**Files to Review**:

- `services/merchant-service/Dockerfile` (port fix)
- `cmd/service-discovery/main.go` (port fix)
- `frontend/lib/api-config.ts` (auth path fix)
- `services/api-gateway/internal/handlers/gateway.go` (UUID validation, auth login, 404 handler)
- `services/api-gateway/internal/supabase/client.go` (SignInWithPassword)
- `services/api-gateway/cmd/main.go` (route registration, comments)
- `services/api-gateway/internal/config/config.go` (CORS fix)
- `services/risk-assessment-service/railway.json` (restart retries)
- `services/merchant-service/cmd/main.go` (route comments)

**Verification Criteria**:

- No linting errors
- Route registration follows guidelines
- Path transformations are correct
- Error handling is comprehensive

---

### 1.2 Local Build Verification

**Objective**: Ensure all services build successfully

**Tasks**:

- [ ] Build API Gateway service locally
- [ ] Build Merchant Service locally
- [ ] Build Service Discovery locally
- [ ] Build Frontend Service locally
- [ ] Verify no compilation errors
- [ ] Check Docker images build successfully

**Commands**:

```bash
# API Gateway
cd services/api-gateway && go build ./cmd/main.go

# Merchant Service
cd services/merchant-service && go build ./cmd/main.go

# Service Discovery
cd cmd/service-discovery && go build main.go

# Frontend (if applicable)
cd frontend && npm run build
```

---

## Phase 2: Post-Deployment Health Checks

### 2.1 Service Health Verification

**Objective**: Verify all services are healthy after deployment

**Services to Test**:

- [ ] API Gateway: `GET https://api-gateway-service-production-21fd.up.railway.app/health`
- [ ] Classification: `GET https://classification-service-production.up.railway.app/health`
- [ ] Merchant: `GET https://merchant-service-production.up.railway.app/health`
- [ ] Risk Assessment: `GET https://risk-assessment-service-production.up.railway.app/health`
- [ ] Frontend: `GET https://frontend-service-production-b225.up.railway.app/health`
- [ ] Pipeline: `GET https://pipeline-service-production.up.railway.app/health`
- [ ] Service Discovery: `GET https://service-discovery-production.up.railway.app/health`
- [ ] BI Service: `GET https://bi-service-production.up.railway.app/health`
- [ ] Monitoring: `GET https://monitoring-service-production.up.railway.app/health`

**Expected Results**:

- All services return 200 OK
- Health check JSON includes service status
- Supabase connection status (where applicable)

**Remediation**:

- If service unhealthy: Check Railway logs, verify environment variables, check service startup

---

### 2.2 Port Configuration Verification

**Objective**: Verify port fixes are applied correctly

**Tests**:

- [ ] Merchant Service: Verify service listens on port 8080 (not 8082)
- [ ] Service Discovery: Verify service listens on port 8080 (not 8086)
- [ ] Check Railway logs for port binding errors
- [ ] Verify health checks use correct ports

**Verification Method**:

- Check Railway service logs for port binding messages
- Verify health check endpoints respond
- Check service metrics for correct port usage

**Remediation**:

- If port mismatch: Verify Dockerfile and service config match
- Redeploy service if needed

---

## Phase 3: Critical Route Testing

### 3.1 Authentication Routes (Newly Implemented/Fixed)

**Objective**: Verify auth routes work correctly after fixes

**Registration Route**:

- [ ] `POST /api/v1/auth/register`
  - **Test**: Valid registration data
  - **Expected**: 201 Created, user created in Supabase
  - **Verify**: Response includes user info, no path mismatch errors
  - **Frontend**: Verify frontend calls `/api/v1/auth/register` (not `/v1/auth/register`)

**Login Route (Newly Implemented)**:

- [ ] `POST /api/v1/auth/login`
  - **Test**: Valid email and password
  - **Expected**: 200 OK, JWT token returned
  - **Verify**: 
    - Token is valid format
    - User info included
    - Supabase authentication works
  - **Frontend**: Verify frontend calls `/api/v1/auth/login` (not `/v1/auth/login`)

**Error Scenarios**:

- [ ] Invalid credentials: 401 Unauthorized
- [ ] Missing fields: 400 Bad Request
- [ ] Invalid email format: 400 Bad Request
- [ ] Non-existent user: 401 Unauthorized

**Remediation**:

- If registration fails: Check Supabase configuration, verify API keys
- If login fails: Verify SignInWithPassword implementation, check Supabase Auth API
- If path mismatch: Verify frontend api-config.ts changes deployed

---

### 3.2 Risk Indicators Route (UUID Validation Fix)

**Objective**: Verify UUID validation and path transformation work correctly

**Valid UUID Test**:

- [ ] `GET /api/v1/risk/indicators/{valid-uuid}`
  - **Test**: Valid UUID format (e.g., `550e8400-e29b-41d4-a716-446655440000`)
  - **Expected**: 200 OK, risk indicators data
  - **Verify**: 
    - Path transformed to `/api/v1/risk/predictions/{uuid}`
    - UUID validation passes
    - Correct data returned from Risk Assessment Service
    - No UUID parsing errors in logs

**Invalid UUID Tests**:

- [ ] `GET /api/v1/risk/indicators/invalid-id`
  - **Expected**: 400 Bad Request
  - **Verify**: 
    - Error message: "Invalid merchant ID format: expected UUID"
    - UUID validation catches invalid format
    - Logs show validation failure

- [ ] `GET /api/v1/risk/indicators/indicators` (edge case)
  - **Expected**: 400 Bad Request
  - **Verify**: UUID validation prevents "indicators" being used as ID

- [ ] `GET /api/v1/risk/indicators/` (missing ID)
  - **Expected**: 400 Bad Request
  - **Verify**: Error message indicates missing merchant ID

**Remediation**:

- If valid UUID fails: Check path transformation logic, verify Risk Assessment Service endpoint
- If invalid UUID not caught: Review UUID validation regex pattern
- If path parsing wrong: Verify parts[5] index is correct

---

### 3.3 CORS Configuration Testing

**Objective**: Verify CORS wildcard fix works correctly

**Preflight Requests**:

- [ ] `OPTIONS /api/v1/merchants`
  - **Expected**: 200 OK
  - **Verify**: 
    - CORS headers present
    - `Access-Control-Allow-Origin` is specific frontend URL (not `*`)
    - `Access-Control-Allow-Credentials: true`
    - No browser CORS errors

**Cross-Origin Requests**:

- [ ] Request from Frontend Origin
  - **Expected**: CORS headers allow request
  - **Verify**: 
    - `Access-Control-Allow-Origin` matches frontend URL
    - No CORS errors in browser console
    - Credentials work correctly

**Remediation**:

- If CORS errors: Verify `CORS_ALLOWED_ORIGINS` environment variable set correctly
- If wildcard still used: Check config.go default value
- If credentials fail: Verify specific origin is set (not wildcard)

---

## Phase 4: Route Precedence Testing

### 4.1 Merchant Route Precedence

**Objective**: Verify route registration order is correct

**Tests**:

- [ ] `GET /api/v1/merchants/{id}/analytics`
  - **Verify**: Matches before `/api/v1/merchants/{id}`
  - **Expected**: Analytics handler called, not base handler

- [ ] `GET /api/v1/merchants/search`
  - **Verify**: Matches before `/api/v1/merchants/{id}`
  - **Expected**: Search handler called, not base handler

- [ ] `GET /api/v1/merchants/analytics`
  - **Verify**: Matches before `/api/v1/merchants/{id}`
  - **Expected**: Analytics handler called, not base handler

**Remediation**:

- If wrong handler called: Verify route registration order in main.go
- Check that specific routes registered before PathPrefix

---

### 4.2 Risk Route Precedence

**Objective**: Verify risk routes with transformations match before PathPrefix

**Tests**:

- [ ] `POST /api/v1/risk/assess`
  - **Verify**: Matches before PathPrefix
  - **Expected**: Path transformed to `/api/v1/assess`

- [ ] `GET /api/v1/risk/indicators/{id}`
  - **Verify**: Matches before PathPrefix
  - **Expected**: UUID validated and path transformed

**Remediation**:

- If PathPrefix matches first: Reorder route registration
- Verify specific routes registered before `api.PathPrefix("/risk")`

---

## Phase 5: Path Transformation Testing

### 5.1 Risk Assessment Transformations

**Tests**:

- [ ] `/api/v1/risk/assess` → `/api/v1/assess`
  - **Verify**: Backend receives `/api/v1/assess`
  - **Verify**: Response returned correctly

- [ ] `/api/v1/risk/metrics` → `/api/v1/metrics`
  - **Verify**: Backend receives `/api/v1/metrics`
  - **Verify**: Metrics data returned

- [ ] `/api/v1/risk/indicators/{uuid}` → `/api/v1/risk/predictions/{uuid}`
  - **Verify**: UUID extracted correctly
  - **Verify**: Path transformed correctly
  - **Verify**: Backend receives correct path

**Remediation**:

- If transformation wrong: Check ProxyToRiskAssessment handler logic
- Add logging to verify path transformation
- Verify Risk Assessment Service receives correct path

---

### 5.2 Session Transformations

**Tests**:

- [ ] `/api/v1/sessions` → `/v1/sessions`
  - **Verify**: `/api` prefix removed
  - **Verify**: Frontend Service receives `/v1/sessions`

- [ ] `/api/v1/sessions/current` → `/v1/sessions/current`
  - **Verify**: Path transformed correctly
  - **Verify**: Query parameters preserved

**Remediation**:

- If path wrong: Check ProxyToSessions handler
- Verify path trimming logic

---

### 5.3 BI Transformations

**Tests**:

- [ ] `/api/v1/bi/analyze` → `/analyze`
  - **Verify**: Prefix removed correctly

- [ ] `/api/v3/dashboard/metrics` → `/dashboard/kpis`
  - **Verify**: Path transformed to BI Service endpoint

**Remediation**:

- If transformation fails: Check ProxyToBI and ProxyToDashboardMetricsV3 handlers

---

## Phase 6: Error Handling Testing

### 6.1 404 Not Found Handler

**Objective**: Verify improved 404 error handling

**Tests**:

- [ ] `GET /api/v1/nonexistent`
  - **Expected**: 404 Not Found
  - **Verify**: 
    - Helpful error message
    - Suggestions provided
    - Available endpoints listed
    - Request logged with context

- [ ] `POST /api/v1/invalid/route`
  - **Expected**: 404 Not Found
  - **Verify**: Error response includes path, method, timestamp

**Remediation**:

- If 404 handler not called: Verify NotFoundHandler registered
- If error message not helpful: Review HandleNotFound implementation

---

### 6.2 UUID Validation Errors

**Tests**:

- [ ] Invalid UUID format
  - **Verify**: 400 Bad Request returned
  - **Verify**: Error message is clear
  - **Verify**: Logs show validation failure

**Remediation**:

- If validation not working: Check isValidUUID function
- Verify UUID regex pattern is correct

---

## Phase 7: Frontend Integration Testing

### 7.1 Frontend API Configuration

**Objective**: Verify frontend uses correct API paths

**Tests**:

- [ ] Frontend loads without console errors
- [ ] API base URL configured correctly
- [ ] No "localhost" warnings in console
- [ ] Auth endpoints use `/api/v1/auth/*` paths

**Browser Testing**:

1. Open frontend in browser
2. Open DevTools → Console
3. Check for API configuration warnings
4. Verify API calls use correct base URL

**Remediation**:

- If localhost warning: Verify `NEXT_PUBLIC_API_BASE_URL` set in Railway
- If wrong paths: Check api-config.ts changes deployed

---

### 7.2 Frontend API Calls

**Tests**:

- [ ] Registration form submission
  - **Verify**: POST to `/api/v1/auth/register` (not `/v1/auth/register`)
  - **Verify**: Request succeeds

- [ ] Login form submission
  - **Verify**: POST to `/api/v1/auth/login` (not `/v1/auth/login`)
  - **Verify**: Token received and stored

- [ ] Merchant list load
  - **Verify**: GET to `/api/v1/merchants` succeeds
  - **Verify**: Data displayed correctly

**Remediation**:

- If path mismatch: Verify frontend build includes api-config.ts changes
- Redeploy frontend if needed

---

## Phase 8: Railway Configuration Verification

### 8.1 Environment Variables

**Objective**: Verify all required environment variables are set

**Critical Variables to Verify**:

- [ ] `NEXT_PUBLIC_API_BASE_URL` (Frontend Service)
  - **Value**: `https://api-gateway-service-production-21fd.up.railway.app`
  - **Verify**: Set in Railway dashboard

- [ ] `CORS_ALLOWED_ORIGINS` (API Gateway)
  - **Value**: `https://frontend-service-production-b225.up.railway.app` (or specific origin)
  - **Verify**: Not set to `*` if credentials enabled

- [ ] All `SUPABASE_*` variables (All services)
  - **Verify**: Set and valid

**Remediation**:

- If variable missing: Set in Railway dashboard
- If wrong value: Update in Railway dashboard
- Redeploy service after variable changes

---

### 8.2 Railway Configuration Files

**Objective**: Verify railway.json configurations are correct

**Files to Verify**:

- [ ] `services/risk-assessment-service/railway.json`
  - **Verify**: `restartPolicyMaxRetries` is 10 (not 3)

**Remediation**:

- If config wrong: Update railway.json and redeploy

---

## Phase 9: Comprehensive Route Testing

### 9.1 All API Gateway Routes

**Test each route from the comprehensive analysis report**:

**Health Routes**:

- [ ] `/health` - API Gateway
- [ ] `/api/v1/classification/health` - Proxy
- [ ] `/api/v1/merchant/health` - Proxy
- [ ] `/api/v1/risk/health` - Proxy

**Classification Routes**:

- [ ] `POST /api/v1/classify` - With valid data
- [ ] `POST /api/v1/classify` - With invalid data (400)

**Merchant Routes** (Test all from checklist):

- [ ] All CRUD operations
- [ ] All sub-routes
- [ ] Search and analytics

**Risk Assessment Routes**:

- [ ] All core routes
- [ ] All path transformations
- [ ] UUID validation

**Auth Routes**:

- [ ] Registration (fixed path)
- [ ] Login (newly implemented)

**Session Routes**:

- [ ] All session operations
- [ ] Path transformations

**BI Routes**:

- [ ] All BI endpoints
- [ ] Path transformations

**Compliance Routes**:

- [ ] Compliance status endpoint

---

### 9.2 Route Registration Order Verification

**Objective**: Verify routes are registered in correct order

**Method**: Review code and test route matching

**Verify**:

- [ ] Specific routes before PathPrefix in API Gateway
- [ ] Specific routes before base routes in Merchant Service
- [ ] Route comments explain order

**Remediation**:

- If order wrong: Reorder route registration
- Add/update comments explaining order

---

## Phase 10: Performance and Monitoring

### 10.1 Response Time Testing

**Objective**: Verify routes respond within acceptable timeframes

**Target Response Times**:

- [ ] Health checks: < 1s
- [ ] Classification: < 5s
- [ ] Merchant CRUD: < 2s
- [ ] Risk assessment: < 10s
- [ ] Authentication: < 2s

**Remediation**:

- If slow: Check database queries, external API calls
- Optimize slow routes

---

### 10.2 Logging Verification

**Objective**: Verify routes are logged correctly

**Checks**:

- [ ] 404 routes logged with context
- [ ] Path transformations logged
- [ ] UUID validation failures logged
- [ ] Error routes logged with details

**Remediation**:

- If not logged: Add logging to handlers
- Verify logger configuration

---

## Phase 11: Security Testing

### 11.1 Input Validation

**Tests**:

- [ ] SQL Injection attempts blocked
- [ ] XSS attempts sanitized
- [ ] Path traversal attempts blocked
- [ ] Invalid UUIDs rejected (already tested)

**Remediation**:

- If vulnerabilities found: Add input sanitization
- Review security headers

---

### 11.2 Authentication Testing

**Tests**:

- [ ] Protected routes require auth (if any)
- [ ] Public routes accessible without auth
- [ ] Token validation works

**Remediation**:

- If auth issues: Review authentication middleware
- Verify public endpoint list

---

## Phase 12: Regression Testing

### 12.1 Previously Working Routes

**Objective**: Verify fixes didn't break existing functionality

**Tests**:

- [ ] All routes that worked before still work
- [ ] No new 404 errors introduced
- [ ] Path transformations still correct
- [ ] Route precedence maintained
- [ ] CORS still works correctly

**Remediation**:

- If regression found: Revert problematic changes
- Fix issue and retest

---

## Phase 13: Production Verification

### 13.1 End-to-End User Flows

**Objective**: Test complete user journeys

**Flows to Test**:

- [ ] User registration → Login → Access merchant list
- [ ] Create merchant → View merchant → Update merchant
- [ ] Classify business → Assess risk → View compliance
- [ ] Search merchants → View analytics → Export data

**Remediation**:

- If flow breaks: Identify failing step
- Fix route or handler
- Retest flow

---

### 13.2 Railway Logs Analysis

**Objective**: Verify no errors in production logs

**Checks**:

- [ ] No UUID parsing errors
- [ ] No route not found errors (except expected 404s)
- [ ] No CORS errors
- [ ] No path transformation errors
- [ ] No authentication errors (except expected 401s)

**Remediation**:

- If errors found: Investigate root cause
- Apply fixes
- Monitor logs after fix

---

## Phase 14: Documentation and Reporting

### 14.1 Test Results Documentation

**Objective**: Document all test results

**Deliverables**:

- [ ] Test results spreadsheet or document
- [ ] List of passing tests
- [ ] List of failing tests with details
- [ ] Remediation actions taken
- [ ] Final status of all routes

---

### 14.2 Issue Tracking

**Objective**: Track and resolve all issues found

**Process**:

- [ ] Document each issue found
- [ ] Prioritize issues (Critical, High, Medium, Low)
- [ ] Create remediation plan for each issue
- [ ] Track resolution status
- [ ] Verify fixes with retesting

---

## Remediation Procedures

### Critical Issues (Service Down)

1. **Immediate Action**: Check Railway service status
2. **Investigation**: Review Railway logs for errors
3. **Fix**: Apply fix or revert change
4. **Verification**: Test fix immediately
5. **Documentation**: Document issue and resolution

### High Priority Issues (Route Not Working)

1. **Investigation**: Test route manually
2. **Analysis**: Check route registration, path transformation, handler
3. **Fix**: Apply fix based on analysis
4. **Testing**: Test fix thoroughly
5. **Documentation**: Update test results

### Medium Priority Issues (Route Works But Has Issues)

1. **Documentation**: Document issue
2. **Planning**: Create fix plan
3. **Implementation**: Apply fix
4. **Testing**: Verify fix
5. **Documentation**: Update documentation

---

## Success Criteria

**All phases complete when**:

- [ ] All services healthy
- [ ] All critical routes working
- [ ] All path transformations correct
- [ ] All error handling working
- [ ] Frontend integration working
- [ ] No critical errors in logs
- [ ] CORS working correctly
- [ ] Authentication working
- [ ] Performance acceptable
- [ ] Documentation complete

---

## Testing Schedule

**Recommended Order**:

1. Phase 1-2: Pre-deployment and health checks (Day 1)
2. Phase 3-4: Critical routes and precedence (Day 1-2)
3. Phase 5-6: Transformations and errors (Day 2)
4. Phase 7-8: Frontend and configuration (Day 2-3)
5. Phase 9-10: Comprehensive and performance (Day 3)
6. Phase 11-12: Security and regression (Day 3-4)
7. Phase 13-14: Production verification and docs (Day 4)

---

## Tools and Resources

**Testing Tools**:

- curl for API testing
- Browser DevTools for frontend testing
- Railway logs for error analysis
- Postman for complex requests (optional)

**Documentation**:

- Route Testing Checklist
- Route Registration Guidelines
- API Routes Comprehensive Analysis Report
- Railway Environment Variables Documentation

---

## Notes

- **Test in production environment after deployment** - All fixes must be deployed before testing
- **Document all test results** - Keep detailed records for future reference
- **Prioritize critical routes first** - Focus on business-critical functionality
- **Fix issues as they are found** - Don't accumulate issues
- **Retest after each fix** - Verify fixes don't break other functionality
- **Keep detailed logs of all testing activities** - Include timestamps, test data, results
- **Test both through API Gateway and directly** - Verify proxy and direct access both work
- **Test with real data when possible** - Use production-like data for realistic testing
- **Clear browser cache between CORS tests** - CORS headers may be cached
- **Monitor Railway logs during testing** - Watch for errors in real-time
- **Test edge cases** - Don't just test happy paths
- **Verify backward compatibility** - Ensure existing integrations still work
- **Test concurrent requests** - Verify no race conditions
- **Document all remediation steps** - Help future debugging

## Risk Mitigation

**If Critical Routes Fail**:

1. Immediately check Railway service status
2. Review recent deployments
3. Check environment variables
4. Review Railway logs
5. Consider rollback if necessary

**If Multiple Routes Fail**:

1. Check API Gateway health
2. Verify service URLs configured correctly
3. Check network connectivity
4. Review middleware configuration
5. Check for service-wide issues

**If Frontend Integration Fails**:

1. Verify NEXT_PUBLIC_API_BASE_URL is set
2. Check browser console for errors
3. Verify CORS configuration
4. Test API endpoints directly
5. Check frontend build includes latest changes

## Escalation Procedures

**Critical Issues** (Service Down, Data Loss):

- Immediate escalation
- Stop testing until resolved
- Document issue thoroughly
- Create remediation plan
- Get approval before applying fixes

**High Priority Issues** (Route Not Working):

- Document immediately
- Create remediation plan
- Fix within same testing session
- Retest after fix

**Medium Priority Issues** (Route Works But Has Issues):

- Document for later
- Continue testing other routes
- Address in follow-up session