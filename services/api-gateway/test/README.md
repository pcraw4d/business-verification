# API Gateway Route Testing

This directory contains comprehensive route testing for the API Gateway.

## Test Files

### `route_test.go`
Go unit tests for route verification. Tests:
- All API Gateway routes
- Route order (specific before PathPrefix)
- Path transformations
- CORS headers
- Query parameter preservation
- Error handling

**Run tests:**
```bash
cd services/api-gateway
go test ./test/route_test.go -v
```

**Run specific test:**
```bash
go test ./test/route_test.go -v -run TestAllRoutes
go test ./test/route_test.go -v -run TestRouteOrder
go test ./test/route_test.go -v -run TestPathTransformations
go test ./test/route_test.go -v -run TestCORSHeaders
go test ./test/route_test.go -v -run TestQueryParameterPreservation
```

## Test Script

### `scripts/test-routes.sh`
Bash script for testing routes against a running API Gateway instance.

**Usage:**
```bash
# Test against local API Gateway
./scripts/test-routes.sh

# Test against remote API Gateway
API_GATEWAY_URL=https://api-gateway-service-production-21fd.up.railway.app ./scripts/test-routes.sh

# Test with specific merchant ID
TEST_MERCHANT_ID=merchant-456 ./scripts/test-routes.sh
```

**Environment Variables:**
- `API_GATEWAY_URL`: API Gateway base URL (default: `http://localhost:8080`)
- `TEST_MERCHANT_ID`: Merchant ID to use in tests (default: `merchant-123`)

**What it tests:**
- Health check routes
- Merchant routes (with valid/invalid IDs)
- Analytics routes (with query parameters)
- Risk assessment routes
- Service health routes
- V3 dashboard routes
- Error cases (404s)
- CORS headers

**Output:**
- ✓ Passed tests (green)
- ✗ Failed tests (red)
- ⊘ Skipped tests (yellow - when API Gateway is not running)
- Summary with pass/fail counts

## Test Coverage

### Routes Tested

#### Merchant Routes
- `GET /api/v1/merchants` - Get all merchants
- `GET /api/v1/merchants/{id}` - Get merchant by ID
- `GET /api/v1/merchants/{id}/analytics` - Get merchant analytics
- `GET /api/v1/merchants/{id}/risk-score` - Get merchant risk score
- `GET /api/v1/merchants/{id}/website-analysis` - Get website analysis
- `GET /api/v1/merchants/analytics` - Get portfolio analytics
- `GET /api/v1/merchants/statistics` - Get portfolio statistics
- `POST /api/v1/merchants/search` - Search merchants

#### Analytics Routes
- `GET /api/v1/analytics/trends` - Get risk trends
- `GET /api/v1/analytics/trends?timeframe=30d&limit=10` - With query params
- `GET /api/v1/analytics/insights` - Get risk insights
- `GET /api/v1/analytics/insights?timeframe=90d&limit=5` - With query params

#### Risk Assessment Routes
- `GET /api/v1/risk/benchmarks?industry=Technology` - Get benchmarks
- `GET /api/v1/risk/indicators/{id}?status=active` - Get risk indicators
- `GET /api/v1/risk/predictions/{merchant_id}` - Get predictions
- `GET /api/v1/risk/metrics` - Get risk metrics
- `POST /api/v1/risk/assess` - Create risk assessment

#### Service Health Routes
- `GET /api/v1/classification/health` - Classification service health
- `GET /api/v1/merchant/health` - Merchant service health
- `GET /api/v1/risk/health` - Risk assessment service health

#### V3 Routes
- `GET /api/v3/dashboard/metrics` - Dashboard metrics

#### Error Cases
- Invalid merchant IDs (should return 404)
- Non-existent routes (should return 404)

#### CORS Testing
- OPTIONS requests for preflight
- CORS headers presence
- Allowed origins, methods, headers

## Route Order Verification

The tests verify that route registration order is correct:
1. Specific routes (e.g., `/merchants/statistics`) are registered before PathPrefix catch-all
2. Analytics routes are registered before `/risk` PathPrefix
3. Route matching works correctly

## Path Transformation Testing

Tests verify that path transformations work correctly:
- `/api/v1/risk/assess` → Risk Assessment service
- `/api/v1/analytics/trends` → Risk Assessment service (no transformation)
- Query parameters are preserved

## Continuous Integration

These tests can be integrated into CI/CD pipelines:
```yaml
# Example GitHub Actions
- name: Test API Gateway Routes
  run: |
    cd services/api-gateway
    go test ./test/route_test.go -v
```

## Troubleshooting

### Tests Skip (API Gateway Not Running)
If tests are skipped, ensure:
1. API Gateway is running: `cd services/api-gateway && go run cmd/main.go`
2. Correct URL is set: `export API_GATEWAY_URL=http://localhost:8080`
3. Backend services are accessible

### Route Returns 404
Check:
1. Route is registered in `cmd/main.go`
2. Route order is correct (specific before PathPrefix)
3. HTTP method matches (GET, POST, etc.)
4. Path matches exactly (case-sensitive)

### CORS Headers Missing
Verify:
1. CORS middleware is applied: `router.Use(middleware.CORS(cfg.CORS))`
2. CORS config is correct in environment variables
3. OPTIONS requests are handled

## Next Steps

After route testing is complete:
1. Run integration tests (`integration-testing-api-gateway`)
2. Run performance tests
3. Run security tests
4. Document route mappings

