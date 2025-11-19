# Route Testing Checklist

**Last Updated**: 2025-11-18  
**Status**: Active Testing Guide

## Overview

This document provides a comprehensive checklist for testing all routes in the KYB Platform services. Use this checklist to verify route functionality, path transformations, error handling, and integration.

---

## Pre-Testing Setup

### Environment Preparation

- [ ] All services deployed and healthy
- [ ] Environment variables configured correctly
- [ ] Database connections verified
- [ ] Redis cache accessible (if applicable)
- [ ] Supabase authentication configured
- [ ] API Gateway accessible
- [ ] Frontend service accessible

### Testing Tools

- [ ] Browser DevTools (Network tab)
- [ ] curl or Postman for API testing
- [ ] Railway logs access
- [ ] Health check endpoints verified

---

## Health Check Routes

### All Services

Test health endpoints for all services:

- [ ] API Gateway: `GET /health`
  - Expected: 200 OK, JSON with status
  - Verify: Supabase connection status

- [ ] Classification Service: `GET /health`
  - Expected: 200 OK
  - Verify: Service responding

- [ ] Merchant Service: `GET /health`
  - Expected: 200 OK
  - Verify: Service responding

- [ ] Risk Assessment Service: `GET /health`
  - Expected: 200 OK
  - Verify: Service responding

- [ ] Frontend Service: `GET /health`
  - Expected: 200 OK
  - Verify: Service responding

- [ ] Pipeline Service: `GET /health`
  - Expected: 200 OK
  - Verify: Service responding

- [ ] Service Discovery: `GET /health`
  - Expected: 200 OK
  - Verify: Service responding

- [ ] BI Service: `GET /health`
  - Expected: 200 OK
  - Verify: Service responding

- [ ] Monitoring Service: `GET /health`
  - Expected: 200 OK
  - Verify: Service responding

### Proxy Health Checks

- [ ] Classification Health Proxy: `GET /api/v1/classification/health`
  - Expected: 200 OK
  - Verify: Proxies correctly to Classification Service

- [ ] Merchant Health Proxy: `GET /api/v1/merchant/health`
  - Expected: 200 OK
  - Verify: Proxies correctly to Merchant Service

- [ ] Risk Health Proxy: `GET /api/v1/risk/health`
  - Expected: 200 OK
  - Verify: Proxies correctly to Risk Assessment Service

---

## Classification Routes

### Classification Endpoint

- [ ] `POST /api/v1/classify`
  - **Request Body**: Valid business data
  - **Expected**: 200 OK, classification results
  - **Verify**: 
    - Response includes MCC, SIC, NAICS codes
    - Confidence scores present
    - Caching works (second request faster)

- [ ] `POST /api/v1/classify` (Invalid Data)
  - **Request Body**: Missing required fields
  - **Expected**: 400 Bad Request
  - **Verify**: Error message is clear

- [ ] `POST /api/v1/classify` (CORS Preflight)
  - **Request**: OPTIONS /api/v1/classify
  - **Expected**: 200 OK with CORS headers
  - **Verify**: CORS headers present

---

## Merchant Routes

### Merchant CRUD Operations

- [ ] `GET /api/v1/merchants`
  - **Expected**: 200 OK, list of merchants
  - **Verify**: Pagination works (if implemented)

- [ ] `POST /api/v1/merchants`
  - **Request Body**: Valid merchant data
  - **Expected**: 201 Created or 200 OK
  - **Verify**: Merchant created in database

- [ ] `GET /api/v1/merchants/{id}`
  - **Expected**: 200 OK, merchant details
  - **Verify**: Correct merchant returned

- [ ] `PUT /api/v1/merchants/{id}`
  - **Request Body**: Updated merchant data
  - **Expected**: 200 OK
  - **Verify**: Changes persisted

- [ ] `DELETE /api/v1/merchants/{id}`
  - **Expected**: 200 OK or 204 No Content
  - **Verify**: Merchant deleted

### Merchant Sub-Routes

- [ ] `GET /api/v1/merchants/{id}/analytics`
  - **Expected**: 200 OK, analytics data
  - **Verify**: Route matches before /merchants/{id}

- [ ] `GET /api/v1/merchants/{id}/website-analysis`
  - **Expected**: 200 OK, website analysis
  - **Verify**: Route matches correctly

- [ ] `GET /api/v1/merchants/{id}/risk-score`
  - **Expected**: 200 OK, risk score
  - **Verify**: Route matches correctly

### Merchant Search and Analytics

- [ ] `POST /api/v1/merchants/search`
  - **Request Body**: Search criteria
  - **Expected**: 200 OK, search results
  - **Verify**: Search works correctly

- [ ] `GET /api/v1/merchants/analytics`
  - **Expected**: 200 OK, general analytics
  - **Verify**: Route matches before /merchants/{id}

- [ ] `GET /api/v1/merchants/statistics`
  - **Expected**: 200 OK, statistics
  - **Verify**: Statistics returned

---

## Risk Assessment Routes

### Core Risk Assessment

- [ ] `POST /api/v1/risk/assess`
  - **Request Body**: Business data for risk assessment
  - **Expected**: 200 OK, risk assessment results
  - **Verify**: 
    - Path transformed to /api/v1/assess
    - Risk score calculated
    - Response includes risk factors

- [ ] `GET /api/v1/risk/benchmarks`
  - **Expected**: 200 OK or error if not implemented
  - **Verify**: Handles gracefully if not available

- [ ] `GET /api/v1/risk/predictions/{merchant_id}`
  - **Expected**: 200 OK, risk predictions
  - **Verify**: Valid UUID required

### Risk Indicators (Path Transformation)

- [ ] `GET /api/v1/risk/indicators/{valid-uuid}`
  - **Expected**: 200 OK, risk indicators
  - **Verify**: 
    - Path transformed to /api/v1/risk/predictions/{uuid}
    - UUID validation works
    - Correct data returned

- [ ] `GET /api/v1/risk/indicators/invalid-id`
  - **Expected**: 400 Bad Request
  - **Verify**: 
    - UUID validation catches invalid format
    - Error message is clear
    - Logs show validation failure

- [ ] `GET /api/v1/risk/indicators/` (missing ID)
  - **Expected**: 400 Bad Request
  - **Verify**: Handles missing ID gracefully

### Risk Metrics

- [ ] `GET /api/v1/risk/metrics`
  - **Expected**: 200 OK, metrics data
  - **Verify**: Path transformed to /api/v1/metrics

---

## Authentication Routes

### Registration

- [ ] `POST /api/v1/auth/register`
  - **Request Body**: Valid registration data
  - **Expected**: 201 Created
  - **Verify**: 
    - User created in Supabase
    - Response includes user info
    - Email validation works

- [ ] `POST /api/v1/auth/register` (Duplicate Email)
  - **Request Body**: Existing email
  - **Expected**: 409 Conflict
  - **Verify**: Error message indicates duplicate

- [ ] `POST /api/v1/auth/register` (Invalid Data)
  - **Request Body**: Missing required fields
  - **Expected**: 400 Bad Request
  - **Verify**: Validation errors clear

### Login

- [ ] `POST /api/v1/auth/login`
  - **Request Body**: Valid email and password
  - **Expected**: 200 OK
  - **Verify**: 
    - JWT token returned
    - User info included
    - Token is valid format

- [ ] `POST /api/v1/auth/login` (Invalid Credentials)
  - **Request Body**: Wrong password
  - **Expected**: 401 Unauthorized
  - **Verify**: Error message doesn't reveal which field is wrong

- [ ] `POST /api/v1/auth/login` (Non-existent User)
  - **Request Body**: Unknown email
  - **Expected**: 401 Unauthorized
  - **Verify**: Generic error message

---

## Session Routes

- [ ] `GET /api/v1/sessions`
  - **Expected**: 200 OK, session list
  - **Verify**: Path transformed correctly

- [ ] `POST /api/v1/sessions`
  - **Expected**: 201 Created
  - **Verify**: Session created

- [ ] `GET /api/v1/sessions/current`
  - **Expected**: 200 OK, current session
  - **Verify**: Route matches before /sessions PathPrefix

- [ ] `GET /api/v1/sessions/metrics`
  - **Expected**: 200 OK, session metrics
  - **Verify**: Route matches correctly

---

## Business Intelligence Routes

- [ ] `POST /api/v1/bi/analyze`
  - **Request Body**: Analysis request
  - **Expected**: 200 OK, analysis results
  - **Verify**: Path transformed correctly

- [ ] `GET /api/v3/dashboard/metrics`
  - **Expected**: 200 OK, dashboard metrics
  - **Verify**: 
    - Path transformed to /dashboard/kpis
    - Proxies to BI Service correctly

---

## Compliance Routes

- [ ] `GET /api/v1/compliance/status`
  - **Expected**: 200 OK, compliance status
  - **Verify**: 
    - Path transformation works
    - Aggregate or business-specific status returned

---

## Error Scenarios

### 404 Not Found

- [ ] Invalid Route: `GET /api/v1/nonexistent`
  - **Expected**: 404 Not Found
  - **Verify**: 
    - Helpful error message
    - Suggestions provided
    - Available endpoints listed
    - Logged for debugging

- [ ] Invalid Method: `DELETE /api/v1/classify`
  - **Expected**: 405 Method Not Allowed
  - **Verify**: Error message indicates allowed methods

### 400 Bad Request

- [ ] Invalid JSON: `POST /api/v1/merchants` with malformed JSON
  - **Expected**: 400 Bad Request
  - **Verify**: Clear error message

- [ ] Missing Required Fields: `POST /api/v1/auth/register` without email
  - **Expected**: 400 Bad Request
  - **Verify**: Lists missing fields

### 500 Internal Server Error

- [ ] Database Connection Failure
  - **Expected**: 503 Service Unavailable or 500
  - **Verify**: Error doesn't expose internal details

---

## CORS Testing

### Preflight Requests

- [ ] `OPTIONS /api/v1/merchants`
  - **Expected**: 200 OK
  - **Verify**: 
    - CORS headers present
    - Access-Control-Allow-Origin set correctly
    - Access-Control-Allow-Methods includes requested method
    - Access-Control-Allow-Headers includes requested headers

### Cross-Origin Requests

- [ ] Request from Frontend Origin
  - **Expected**: CORS headers allow request
  - **Verify**: 
    - Access-Control-Allow-Origin matches frontend URL
    - Credentials allowed if needed
    - No CORS errors in browser console

- [ ] Request from Different Origin
  - **Expected**: CORS headers may block or allow
  - **Verify**: Behavior matches configuration

---

## Path Transformation Testing

### Risk Assessment Transformations

- [ ] `/api/v1/risk/assess` → `/api/v1/assess`
  - **Verify**: Backend receives correct path
  - **Verify**: Response returned correctly

- [ ] `/api/v1/risk/metrics` → `/api/v1/metrics`
  - **Verify**: Path transformed correctly

- [ ] `/api/v1/risk/indicators/{uuid}` → `/api/v1/risk/predictions/{uuid}`
  - **Verify**: UUID extracted correctly
  - **Verify**: Path transformed correctly
  - **Verify**: Invalid UUID rejected

### Session Transformations

- [ ] `/api/v1/sessions/*` → `/v1/sessions/*`
  - **Verify**: /api prefix removed
  - **Verify**: Query parameters preserved

### BI Transformations

- [ ] `/api/v1/bi/*` → `/*` (after /api/v1/bi)
  - **Verify**: Prefix removed correctly
  - **Verify**: Remaining path forwarded

---

## Route Precedence Testing

### Merchant Routes

- [ ] `/api/v1/merchants/{id}/analytics` matches before `/api/v1/merchants/{id}`
  - **Verify**: Analytics handler called, not base handler

- [ ] `/api/v1/merchants/search` matches before `/api/v1/merchants/{id}`
  - **Verify**: Search handler called, not base handler

- [ ] `/api/v1/merchants/analytics` matches before `/api/v1/merchants/{id}`
  - **Verify**: Analytics handler called, not base handler

### Risk Routes

- [ ] `/api/v1/risk/assess` matches before PathPrefix
  - **Verify**: Specific handler called with transformation

- [ ] `/api/v1/risk/indicators/{id}` matches before PathPrefix
  - **Verify**: Specific handler called with UUID validation

---

## Frontend Integration Testing

### API Calls from Frontend

- [ ] Frontend loads successfully
  - **Verify**: No console errors
  - **Verify**: API base URL configured correctly

- [ ] Merchant list loads
  - **Verify**: API call to `/api/v1/merchants` succeeds
  - **Verify**: Data displayed correctly

- [ ] Add merchant form
  - **Verify**: POST to `/api/v1/merchants` succeeds
  - **Verify**: Success message displayed

- [ ] Merchant details page
  - **Verify**: GET to `/api/v1/merchants/{id}` succeeds
  - **Verify**: Data displayed correctly

- [ ] Registration form
  - **Verify**: POST to `/api/v1/auth/register` succeeds
  - **Verify**: Path matches (no /v1/auth/register)

- [ ] Login form
  - **Verify**: POST to `/api/v1/auth/login` succeeds
  - **Verify**: Token stored correctly

### CORS in Browser

- [ ] Check Network tab for CORS headers
  - **Verify**: Access-Control-Allow-Origin present
  - **Verify**: No CORS errors in console
  - **Verify**: Credentials sent if needed

---

## Performance Testing

### Response Times

- [ ] Health checks: < 1s
- [ ] Classification: < 5s
- [ ] Merchant CRUD: < 2s
- [ ] Risk assessment: < 10s
- [ ] Authentication: < 2s

### Caching

- [ ] Classification results cached
  - **Verify**: Second request faster
  - **Verify**: Same results returned

---

## Security Testing

### Authentication

- [ ] Protected routes require auth
  - **Verify**: 401 Unauthorized without token
  - **Verify**: 200 OK with valid token

### Input Validation

- [ ] SQL Injection attempts blocked
- [ ] XSS attempts sanitized
- [ ] Path traversal attempts blocked
- [ ] Invalid UUIDs rejected

---

## Logging Verification

### Route Matching Logs

- [ ] 404 routes logged
  - **Verify**: Method, path, query logged
  - **Verify**: User agent logged
  - **Verify**: Request ID present

### Path Transformation Logs

- [ ] Transformations logged
  - **Verify**: Original and transformed paths logged
  - **Verify**: UUID validation failures logged

### Error Logs

- [ ] Errors logged with context
  - **Verify**: Request ID present
  - **Verify**: Path and method logged
  - **Verify**: Error details included

---

## Regression Testing

After making route changes:

- [ ] All previously working routes still work
- [ ] No new 404 errors introduced
- [ ] Path transformations still correct
- [ ] Route precedence maintained
- [ ] CORS still works correctly

---

## Production Verification

### Railway Deployment

- [ ] All services deployed successfully
- [ ] Health checks passing
- [ ] No errors in Railway logs
- [ ] Environment variables set correctly
- [ ] Service URLs accessible

### End-to-End Flows

- [ ] User registration → Login → Access protected resource
- [ ] Create merchant → View merchant → Update merchant
- [ ] Classify business → Assess risk → View compliance

---

## Testing Tools and Commands

### curl Examples

```bash
# Health check
curl https://api-gateway-service-production-21fd.up.railway.app/health

# Classification
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Business", "description": "Test"}'

# Merchant list
curl https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants

# Risk assessment
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/assess \
  -H "Content-Type: application/json" \
  -d '{"business_id": "test-123"}'
```

### Browser Testing

1. Open browser DevTools → Network tab
2. Navigate to frontend
3. Perform actions (click buttons, submit forms)
4. Verify API calls in Network tab
5. Check for CORS errors
6. Verify response status codes

---

## Issue Reporting Template

When reporting route issues, include:

- Route path and method
- Request body (if applicable)
- Expected behavior
- Actual behavior
- Error messages
- Response status code
- Railway logs (if available)
- Browser console errors (if applicable)

---

## Related Documentation

- [Route Registration Guidelines](./ROUTE_REGISTRATION_GUIDELINES.md)
- [API Routes Comprehensive Analysis Report](../API_ROUTES_COMPREHENSIVE_ANALYSIS_REPORT.md)
- [Railway Environment Variables](./RAILWAY_ENVIRONMENT_VARIABLES.md)

---

**Document Version**: 1.0.0  
**Last Updated**: 2025-11-18  
**Next Review**: 2025-12-18

