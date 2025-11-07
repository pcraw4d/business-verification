# API Gateway Deployment - Risk Endpoints Routing Fix

**Date**: November 7, 2025  
**Version**: v1.0.19  
**Commit**: 3737e1d3b

## Changes Deployed

### 1. Fixed Risk Benchmarks/Predictions Routing

**Issue**: API Gateway was returning 404 for:
- `/api/v1/risk/benchmarks`
- `/api/v1/risk/predictions/{merchant_id}`

**Root Cause**: The `ProxyToRiskAssessment` function was stripping `/risk` from the path before forwarding to the Risk Assessment Service, but the service expects the full path including `/risk`.

**Fix Applied**:
- Updated `services/api-gateway/internal/handlers/gateway.go`:
  - Changed `ProxyToRiskAssessment` to preserve the full path including `/risk`
  - Removed path transformation that was stripping `/risk`
- Updated `services/api-gateway/cmd/main.go`:
  - Added explicit route handlers for `/risk/benchmarks` and `/risk/predictions/{merchant_id}`
  - Ensures proper routing before the catch-all `PathPrefix` handler

### 2. Files Modified

- `services/api-gateway/internal/handlers/gateway.go`
- `services/api-gateway/cmd/main.go`
- `services/api-gateway/Dockerfile` (force rebuild comment)

## Deployment Status

✅ **Changes Committed**: Commit `3737e1d3b`  
✅ **Changes Pushed**: Pushed to `main` branch  
🔄 **Railway Deployment**: Auto-deploying (typically takes 2-5 minutes)

## Verification Steps

Once Railway deployment completes, verify the fixes:

### 1. Check API Gateway Health

```bash
curl https://api-gateway-service-production-21fd.up.railway.app/health
```

Expected: `200 OK` with service status

### 2. Test Risk Benchmarks Endpoint (MCC)

```bash
curl "https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/benchmarks?mcc=5411"
```

Expected: `200 OK` with benchmarks data (not 404)

### 3. Test Risk Benchmarks Endpoint (NAICS)

```bash
curl "https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/benchmarks?naics=541110"
```

Expected: `200 OK` with benchmarks data (not 404)

### 4. Test Risk Predictions Endpoint

```bash
curl "https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/predictions/biz_thegreen_1762487805256?horizons=3,6,12"
```

Expected: `200 OK` with predictions data (not 404)

### 5. Run Full Test Suite

```bash
./scripts/run-all-tests.sh
```

Expected: All 3 previously failing tests should now pass

## Monitoring

### Railway Dashboard
- Check Railway dashboard for deployment status
- Monitor logs for any errors during startup
- Verify service health check passes

### Service Logs
Look for this log message on startup:
```
🚀 Starting KYB API Gateway Service v1.0.19 - Fixed risk benchmarks/predictions routing
```

### Expected Behavior

**Before Fix**:
- `/api/v1/risk/benchmarks` → 404
- `/api/v1/risk/predictions/{id}` → 404

**After Fix**:
- `/api/v1/risk/benchmarks` → 200 OK with benchmarks data
- `/api/v1/risk/predictions/{id}` → 200 OK with predictions data

## Rollback Plan

If issues occur, rollback to previous version:

1. Revert commit `3737e1d3b`:
   ```bash
   git revert 3737e1d3b
   git push origin main
   ```

2. Or manually restore previous routing logic in `gateway.go`

## Related Issues

- Fixes test failures in `comprehensive-api-test.sh`:
  - Risk Benchmarks (MCC) test
  - Risk Benchmarks (NAICS) test
  - Risk Predictions test

## Next Steps

1. ✅ Wait for Railway deployment to complete (2-5 minutes)
2. ✅ Verify endpoints using curl commands above
3. ✅ Run full test suite to confirm all tests pass
4. ✅ Update test results documentation if needed

## Support

If deployment fails or endpoints still return 404:
1. Check Railway deployment logs
2. Verify Risk Assessment Service is healthy
3. Check API Gateway service logs for routing errors
4. Verify environment variables are set correctly

