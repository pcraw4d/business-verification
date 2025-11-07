# Testing Guide - Risk API Endpoints

**Date**: January 2025  
**Status**: Ready for Testing

---

## Testing Options

### Option 1: Test Against Running Services (Recommended)

#### Prerequisites
1. API Gateway running on port 8080
2. Risk Assessment Service running and accessible through gateway

#### Start Services

**API Gateway**:
```bash
cd services/api-gateway
go run cmd/main.go
```

**Risk Assessment Service**:
```bash
cd services/risk-assessment-service
go run cmd/main.go
```

#### Run Tests
```bash
./scripts/test-risk-endpoints.sh
```

---

### Option 2: Code Verification Tests (No Services Required)

These tests verify the code structure and implementation without requiring running services.

#### Test 1: Verify Handler Implementation

```bash
# Check if handlers are implemented
grep -r "HandleRiskBenchmarks\|GetRiskBenchmarksHandler" services/risk-assessment-service/internal/handlers/ internal/api/handlers/

# Expected: Should find both implementations
```

#### Test 2: Verify Route Registration

```bash
# Check if routes are registered
grep -r "risk/benchmarks\|risk/predictions" services/risk-assessment-service/cmd/main.go

# Expected: Should find route registrations
```

#### Test 3: Verify Frontend Integration

```bash
# Check if frontend uses endpoints
grep -r "riskBenchmarks\|riskPredictions" web/shared/data-services/risk-data-service.js

# Expected: Should find endpoint usage
```

---

### Option 3: Unit Tests (If Available)

```bash
# Run Go tests for handlers
cd services/risk-assessment-service
go test ./internal/handlers/... -v

# Run tests for main platform handlers
cd internal/api/handlers
go test -v
```

---

## Manual Testing Checklist

### ✅ Code Structure Verification

- [x] Handlers implemented in both locations
- [x] Routes registered in service
- [x] Frontend integration complete
- [x] API config updated

### ⏳ Service Testing (Requires Running Services)

- [ ] API Gateway health check
- [ ] Benchmarks endpoint (200 response)
- [ ] Predictions endpoint (200 response)
- [ ] Error handling (400/404 responses)
- [ ] Frontend integration (browser testing)

---

## Quick Verification Commands

### Verify Code Implementation

```bash
# 1. Check handlers exist
echo "=== Checking Handlers ==="
grep -l "HandleRiskBenchmarks" services/risk-assessment-service/internal/handlers/risk_assessment.go internal/api/handlers/risk.go
echo ""

# 2. Check routes registered
echo "=== Checking Routes ==="
grep -A2 "risk/benchmarks\|risk/predictions" services/risk-assessment-service/cmd/main.go
echo ""

# 3. Check frontend integration
echo "=== Checking Frontend ==="
grep -l "riskBenchmarks\|loadIndustryBenchmarks" web/shared/data-services/risk-data-service.js
echo ""

# 4. Check API config
echo "=== Checking API Config ==="
grep "riskBenchmarks\|riskPredictions" web/js/api-config.js
```

---

## Expected Test Results

### Code Verification (Should Pass)

✅ **Handlers**: Both implementations found  
✅ **Routes**: Routes registered in service  
✅ **Frontend**: Endpoints configured and used  
✅ **API Config**: Endpoints defined  

### Service Testing (Requires Running Services)

When services are running:
- ✅ Benchmarks: `200 OK` with JSON response
- ✅ Predictions: `200 OK` with JSON response
- ✅ Error handling: `400 Bad Request` for invalid requests

---

## Next Steps

1. **Start Services**: Follow service startup instructions
2. **Run Tests**: Execute test script
3. **Verify Responses**: Check JSON structure
4. **Frontend Testing**: Test in browser

---

## Troubleshooting

### Services Not Starting

**Check**:
- Go version (1.22+)
- Dependencies installed (`go mod download`)
- Port availability
- Environment variables

### Tests Failing

**Check**:
- Service logs
- Network connectivity
- Endpoint URLs
- CORS configuration

---

## Status

✅ **Code Implementation**: Complete  
⏳ **Service Testing**: Requires running services  
✅ **Code Verification**: Ready to run

