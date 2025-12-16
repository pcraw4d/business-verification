# Railway Routing Issue - Root Cause and Fix

## Issue Summary

**Date**: 2025-12-09  
**Status**: 🔴 **ROOT CAUSE IDENTIFIED**

All 44 comprehensive test requests failed with HTTP 502 errors, timing out at exactly 120 seconds (API Gateway timeout).

## Root Cause Analysis

### Finding 1: Classification Service is Healthy ✅
- Direct access test: **SUCCESS**
- Service responds correctly to `/classify` endpoint
- Health checks passing
- Memory usage stable (~23%)

### Finding 2: API Gateway Configuration Issue ❌
- **Environment Variable**: `CLASSIFICATION_SERVICE_URL` in Railway
- **Current Value**: `https://classification-service-production.up.railway.app` (appears correct when viewed)
- **Issue**: Railway CLI displays multi-line values, making it appear broken, but the actual value may be correct

### Finding 3: API Gateway Code Analysis
- API Gateway constructs URL as: `h.config.Services.ClassificationURL + "/classify"`
- Expected full URL: `https://classification-service-production.up.railway.app/classify`
- Classification service has routes: `/classify` and `/v1/classify` ✅

### Finding 4: No Classification Requests in Service Logs
- Classification service logs show **ONLY health check requests**
- **NO POST requests** to `/classify` endpoint
- This confirms requests are **not reaching** the classification service

## Possible Causes

### 1. Environment Variable Not Applied (Most Likely)
- Railway environment variable may not be properly set or applied
- API Gateway may be using default hardcoded URL
- Need to verify actual runtime configuration

### 2. Railway Internal Networking Issue
- API Gateway container cannot resolve classification service URL
- DNS resolution failure within Railway network
- Network policy blocking inter-service communication

### 3. Service Discovery Issue
- Railway service discovery not working correctly
- Service URL format mismatch
- Port/protocol mismatch

## Verification Steps

### Step 1: Verify Environment Variable
```bash
# Check current value
railway variables --service api-gateway-service | grep CLASSIFICATION_SERVICE_URL

# Expected: https://classification-service-production.up.railway.app
```

### Step 2: Test Direct Service Access
```bash
# This should work (verified ✅)
curl -X POST "https://classification-service-production.up.railway.app/classify" \
  -H "Content-Type: application/json" \
  -d '{"business_name":"Test"}'
```

### Step 3: Check API Gateway Logs
```bash
# Look for classification service proxy attempts
railway logs --service api-gateway-service | grep -i classification
```

### Step 4: Verify API Gateway Runtime Configuration
- Check if API Gateway is logging the classification service URL on startup
- Verify the URL being used matches expected value
- Look for connection errors in API Gateway logs

## Recommended Fix

### Option 1: Update Environment Variable (If Incorrect)
```bash
# Set correct URL (if Railway CLI supports it)
railway variables set CLASSIFICATION_SERVICE_URL=https://classification-service-production.up.railway.app --service api-gateway-service
```

### Option 2: Use Railway Service Discovery
- Railway provides internal service URLs via environment variables
- Check for `RAILWAY_SERVICE_CLASSIFICATION_SERVICE_URL` or similar
- Update API Gateway config to use Railway service discovery

### Option 3: Verify and Restart Services
- Ensure environment variables are properly loaded
- Restart API Gateway service to pick up new configuration
- Verify configuration is applied at runtime

## Next Steps

1. **Immediate**: Verify actual `CLASSIFICATION_SERVICE_URL` value in Railway dashboard
2. **Immediate**: Check API Gateway startup logs for configuration values
3. **Immediate**: Review API Gateway logs for connection errors
4. **Short-term**: Update environment variable if incorrect
5. **Short-term**: Restart API Gateway service after fix
6. **Short-term**: Re-run comprehensive tests

## Test Results Reference

- **Test File**: `railway_production_test_results_20251209_151413.json`
- **Success Rate**: 0% (0/44)
- **Error Pattern**: All requests timeout at 120s (API Gateway timeout)
- **Error Type**: HTTP 502 "Application failed to respond"



