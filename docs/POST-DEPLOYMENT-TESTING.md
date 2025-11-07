# Post-Deployment Testing Guide

**Date**: January 2025  
**Purpose**: Test endpoints after Railway deployment completes

---

## Quick Test Commands

### 1. Health Check

```bash
# API Gateway
curl https://kyb-api-gateway-production.up.railway.app/health

# Expected: {"status":"healthy",...}
```

### 2. Benchmarks Endpoint

```bash
# Test with MCC
curl "https://kyb-api-gateway-production.up.railway.app/api/v1/risk/benchmarks?mcc=5411" | jq .

# Test with NAICS
curl "https://kyb-api-gateway-production.up.railway.app/api/v1/risk/benchmarks?naics=541110" | jq .

# Test error case
curl "https://kyb-api-gateway-production.up.railway.app/api/v1/risk/benchmarks"
# Expected: 400 Bad Request
```

### 3. Predictions Endpoint

```bash
# Test with default horizons
curl "https://kyb-api-gateway-production.up.railway.app/api/v1/risk/predictions/test-merchant-123" | jq .

# Test with all options
curl "https://kyb-api-gateway-production.up.railway.app/api/v1/risk/predictions/test-merchant-123?horizons=3,6,12&includeScenarios=true&includeConfidence=true" | jq .
```

---

## Automated Test Script

Update the test script to use production URL:

```bash
# Edit script
nano scripts/test-risk-endpoints.sh

# Change API_BASE_URL
API_BASE_URL="https://kyb-api-gateway-production.up.railway.app"

# Run tests
./scripts/test-risk-endpoints.sh
```

---

## Frontend Testing

1. Open: `https://[frontend-url]/merchant-details.html`
2. Navigate to Risk Indicators tab
3. Open DevTools → Network tab
4. Verify:
   - Benchmarks request succeeds
   - Predictions request succeeds
   - No CORS errors
   - UI components render

---

## Expected Results

### Benchmarks Response
```json
{
  "industry_code": "5411",
  "industry_type": "mcc",
  "benchmarks": {
    "industry": "5411",
    "benchmarks": {
      "average_score": 70.0,
      "median_score": 72.0,
      "percentile_75": 80.0,
      "percentile_90": 85.0
    }
  }
}
```

### Predictions Response
```json
{
  "merchant_id": "test-merchant-123",
  "predictions": [
    {
      "horizon_months": 3,
      "predicted_score": 70.0,
      "trend": "STABLE"
    }
  ]
}
```

---

## Troubleshooting

### If endpoints return 404:
- Check Railway logs
- Verify services deployed
- Check route registration

### If endpoints return 500:
- Check service logs
- Verify handler implementation
- Check dependencies

### If CORS errors:
- Check CORS configuration
- Verify allowed origins
- Check API Gateway settings

---

## Success Indicators

✅ Health checks return 200  
✅ Benchmarks endpoint works  
✅ Predictions endpoint works  
✅ Frontend loads data  
✅ No console errors  
✅ UI renders correctly

