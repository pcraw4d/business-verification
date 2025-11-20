# API Gateway Integration Tests

**Date:** 2025-01-27

## Overview

Comprehensive integration tests for the API Gateway that verify:
- All routes through the gateway
- Path transformations
- Error responses
- CORS headers
- Authentication/authorization
- Query parameter preservation

## Test Structure

### Test Files

1. **`integration_test.go`** - Main integration test suite
   - `TestIntegrationAllRoutes` - Tests all routes through the gateway
   - `TestIntegrationPathTransformations` - Tests path transformations
   - `TestIntegrationCORSHeaders` - Tests CORS headers
   - `TestIntegrationErrorResponses` - Tests error handling
   - `TestIntegrationQueryParameterPreservation` - Tests query parameter handling
   - `TestIntegrationAuthentication` - Tests authentication middleware

2. **`route_test.go`** - Unit tests for route configuration
   - `TestAllRoutes` - Route accessibility tests
   - `TestRouteOrder` - Route order verification
   - `TestPathTransformations` - Path transformation tests
   - `TestCORSHeaders` - CORS header tests
   - `TestQueryParameterPreservation` - Query parameter tests

## Running Tests

### Unit Tests (No Services Required)

```bash
cd services/api-gateway
go test ./test -v
```

These tests use `httptest` and don't require running services.

### Integration Tests (Against Live API Gateway)

```bash
# Set environment variables
export API_GATEWAY_URL=http://localhost:8080
export TEST_MERCHANT_ID=merchant-123
export TEST_USER_ID=test-user-123

# Run integration tests
cd services/api-gateway
go test ./test -v -run TestIntegration
```

### Specific Test Suites

```bash
# Test all routes
go test ./test -v -run TestIntegrationAllRoutes

# Test CORS headers
go test ./test -v -run TestIntegrationCORSHeaders

# Test error responses
go test ./test -v -run TestIntegrationErrorResponses

# Test path transformations
go test ./test -v -run TestIntegrationPathTransformations

# Test authentication
go test ./test -v -run TestIntegrationAuthentication
```

## Test Coverage

### Routes Tested

#### Health & Metrics
- ✅ `/health` - Health check
- ✅ `/` - Root endpoint
- ✅ `/metrics` - Prometheus metrics

#### Merchant Routes
- ✅ `GET /api/v1/merchants` - List merchants
- ✅ `GET /api/v1/merchants/{id}` - Get merchant by ID
- ✅ `GET /api/v1/merchants/{id}/analytics` - Merchant analytics
- ✅ `GET /api/v1/merchants/analytics` - Portfolio analytics
- ✅ `GET /api/v1/merchants/statistics` - Portfolio statistics

#### Analytics Routes
- ✅ `GET /api/v1/analytics/trends` - Risk trends
- ✅ `GET /api/v1/analytics/insights` - Risk insights

#### Risk Assessment Routes
- ✅ `GET /api/v1/risk/benchmarks` - Risk benchmarks
- ✅ `GET /api/v1/risk/indicators/{id}` - Risk indicators
- ✅ `POST /api/v1/risk/assess` - Risk assessment

#### Service Health
- ✅ `GET /api/v1/classification/health` - Classification service health
- ✅ `GET /api/v1/merchant/health` - Merchant service health
- ✅ `GET /api/v1/risk/health` - Risk assessment service health

#### V3 Routes
- ✅ `GET /api/v3/dashboard/metrics` - Dashboard metrics v3

### Features Tested

1. **Route Matching**
   - ✅ Specific routes match before PathPrefix
   - ✅ Route order is correct
   - ✅ PathPrefix catch-all works

2. **Path Transformations**
   - ✅ Analytics routes route to Risk Assessment service
   - ✅ Risk routes route to Risk Assessment service
   - ✅ Merchant routes route to Merchant service

3. **CORS Headers**
   - ✅ CORS headers present for all routes
   - ✅ OPTIONS requests handled correctly
   - ✅ Preflight requests return 200

4. **Error Handling**
   - ✅ 404 for invalid merchant IDs
   - ✅ 404 for non-existent routes
   - ✅ Proper error response format

5. **Query Parameters**
   - ✅ Query parameters preserved
   - ✅ Multiple query parameters work
   - ✅ Query parameters forwarded to backend

6. **Authentication**
   - ✅ Public endpoints accessible without auth
   - ✅ Protected endpoints require auth (when configured)
   - ✅ Invalid tokens rejected

## Test Configuration

### Environment Variables

```bash
# API Gateway URL (for live integration tests)
API_GATEWAY_URL=http://localhost:8080

# Test data
TEST_MERCHANT_ID=merchant-123
TEST_USER_ID=test-user-123

# Service URLs (optional, defaults to localhost)
CLASSIFICATION_SERVICE_URL=http://localhost:8081
MERCHANT_SERVICE_URL=http://localhost:8083
RISK_ASSESSMENT_SERVICE_URL=http://localhost:8082

# Supabase (optional for unit tests)
SUPABASE_URL=https://...
SUPABASE_ANON_KEY=...
```

### Test Modes

1. **Unit Test Mode** (Default)
   - Uses `httptest` - no services required
   - Tests route matching, middleware, handlers
   - Fast execution

2. **Integration Test Mode** (With `API_GATEWAY_URL` set)
   - Tests against live API Gateway
   - Requires services to be running
   - Tests full request/response flow

## Expected Results

### Unit Tests
- ✅ All route matching tests pass
- ✅ CORS header tests pass
- ✅ Error response tests pass
- ⚠️ Proxy tests may return 503 (service URLs not configured) - this is expected

### Integration Tests (With Services Running)
- ✅ All routes return expected status codes
- ✅ CORS headers present
- ✅ Error responses correct
- ✅ Query parameters preserved

## Troubleshooting

### Tests Fail with "Target URL is empty"
**Cause:** Service URLs not configured  
**Solution:** Set service URL environment variables or use unit test mode

### Tests Fail with "Connection refused"
**Cause:** Services not running  
**Solution:** Start services or use unit test mode (httptest)

### CORS Tests Fail
**Cause:** CORS middleware not applied  
**Solution:** Verify middleware is registered in `setupRoutes`

### Authentication Tests Fail
**Cause:** Auth middleware behavior changed  
**Solution:** Update test expectations to match middleware implementation

## Next Steps

1. ✅ Integration tests created
2. Run tests against live API Gateway
3. Document test results
4. Add performance tests
5. Add load tests

