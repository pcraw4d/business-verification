# Quick Start Testing Guide

**Date**: January 2025  
**Purpose**: Quick guide to test the new risk API endpoints

---

## Prerequisites

1. **API Gateway Running**: `http://localhost:8080` (or production URL)
2. **Risk Assessment Service Running**: Should be accessible through gateway
3. **Test Merchant ID**: Any valid merchant ID for testing

---

## Quick Test Commands

### Test Benchmarks Endpoint

```bash
# Basic test with MCC code
curl "http://localhost:8080/api/v1/risk/benchmarks?mcc=5411"

# Test with NAICS code
curl "http://localhost:8080/api/v1/risk/benchmarks?naics=541110"

# Test with SIC code
curl "http://localhost:8080/api/v1/risk/benchmarks?sic=7372"

# Test error handling (should return 400)
curl "http://localhost:8080/api/v1/risk/benchmarks"
```

### Test Predictions Endpoint

```bash
# Basic test with default horizons
curl "http://localhost:8080/api/v1/risk/predictions/test-merchant-123"

# Test with custom horizons
curl "http://localhost:8080/api/v1/risk/predictions/test-merchant-123?horizons=3,6,12"

# Test with scenarios and confidence
curl "http://localhost:8080/api/v1/risk/predictions/test-merchant-123?horizons=3,6,12&includeScenarios=true&includeConfidence=true"

# Test with single horizon
curl "http://localhost:8080/api/v1/risk/predictions/test-merchant-123?horizons=6"
```

---

## Using the Test Script

```bash
# Run all tests
./scripts/test-risk-endpoints.sh

# Test against production
API_BASE_URL=https://kyb-api-gateway-production.up.railway.app ./scripts/test-risk-endpoints.sh

# Test with specific merchant
TEST_MERCHANT_ID=your-merchant-id ./scripts/test-risk-endpoints.sh
```

---

## Expected Responses

### Benchmarks Response

```json
{
  "industry_code": "5411",
  "industry_type": "mcc",
  "mcc": "5411",
  "naics": "",
  "sic": "",
  "benchmarks": {
    "industry": "5411",
    "benchmarks": {
      "average_score": 70.0,
      "median_score": 72.0,
      "percentile_75": 80.0,
      "percentile_90": 85.0
    },
    "last_updated": "2025-01-XX..."
  },
  "timestamp": "2025-01-XX..."
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
      "trend": "STABLE",
      "confidence": 0.75,
      "scenarios": {
        "optimistic": 65.0,
        "realistic": 70.0,
        "pessimistic": 75.0
      }
    },
    {
      "horizon_months": 6,
      "predicted_score": 72.0,
      "trend": "STABLE",
      "confidence": 0.75,
      "scenarios": {
        "optimistic": 67.0,
        "realistic": 72.0,
        "pessimistic": 77.0
      }
    },
    {
      "horizon_months": 12,
      "predicted_score": 75.0,
      "trend": "STABLE",
      "confidence": 0.75,
      "scenarios": {
        "optimistic": 70.0,
        "realistic": 75.0,
        "pessimistic": 80.0
      }
    }
  ],
  "generated_at": "2025-01-XX...",
  "data_points": 0
}
```

---

## Frontend Testing

### Test in Browser

1. **Open Developer Tools** (F12)
2. **Navigate to Network Tab**
3. **Open Merchant Details Page**
4. **Click on Risk Indicators Tab**
5. **Check Network Requests**:
   - Look for `/api/v1/risk/benchmarks`
   - Look for `/api/v1/risk/predictions/{merchantId}`

### Expected Frontend Behavior

- ✅ Benchmarks load automatically when tab opens
- ✅ Predictions load automatically when tab opens
- ✅ Charts update with real data
- ✅ Fallback to mock data if endpoints unavailable

---

## Troubleshooting

### Endpoint Returns 404

**Cause**: Route not registered or service not running

**Solution**:
1. Check if risk assessment service is running
2. Verify routes are registered in `cmd/main.go`
3. Check API gateway is proxying `/risk` paths

### Endpoint Returns 500

**Cause**: Handler error or service dependency issue

**Solution**:
1. Check service logs
2. Verify ML service is available (for predictions)
3. Check database connection (if using real data)

### Frontend Shows Mock Data

**Cause**: Endpoint unavailable or error occurred

**Solution**:
1. Check Network tab for failed requests
2. Verify API base URL is correct
3. Check CORS headers if cross-origin

---

## Next Steps

1. **Integration Testing**: Run full integration tests
2. **Performance Testing**: Test with load
3. **Database Integration**: Connect to real data sources
4. **Monitoring**: Set up logging and metrics

---

## Status

✅ **Endpoints Ready for Testing**

All endpoints are implemented and registered. Use the commands above to verify functionality.

