# API Gateway to Risk Assessment Service Connectivity Issue

**Date**: 2025-11-17  
**Status**: ⚠️ **INVESTIGATION IN PROGRESS**

## Issue Summary

The API Gateway cannot connect to the Risk Assessment Service, returning 503 "Backend service unavailable" for all Risk Assessment Service endpoints.

## Test Results

### ✅ Risk Assessment Service - Direct Access (WORKING)

All endpoints work when accessed directly:
- ✅ `/health` - 200 OK
- ✅ `/api/v1/health` - 200 OK  
- ✅ `/api/v1/metrics` - 200 OK (returns metrics data)

### ❌ API Gateway - Proxy Requests (FAILING)

All Risk Assessment Service endpoints fail via API Gateway:
- ❌ `/api/v1/risk/health` - 503 Service Unavailable
- ❌ `/api/v1/risk/metrics` - 503 Service Unavailable
- ❌ `/api/v1/risk/benchmarks` - 503 Service Unavailable

**Response Time**: ~0.1 seconds (not a timeout issue)

### ✅ Other Services - Via API Gateway (WORKING)

Other services work correctly via API Gateway:
- ✅ `/api/v1/merchants` - 200 OK
- ✅ `/api/v1/classification/health` - 200 OK

## Analysis

### Key Observations

1. **Service is Healthy**: Risk Assessment Service is running and responding correctly
2. **Direct Access Works**: Service is accessible from the internet
3. **Other Services Work**: API Gateway can reach Classification and Merchant services
4. **All Risk Endpoints Fail**: Not just metrics, but all Risk Assessment Service endpoints
5. **Fast Failure**: 503 returned in ~0.1 seconds (not a timeout)

### Possible Causes

1. **DNS Resolution Issue**: API Gateway container might not be able to resolve `risk-assessment-service-production.up.railway.app`
2. **Network Connectivity**: Railway internal network might have issues between API Gateway and Risk Assessment Service
3. **Environment Variable**: `RISK_ASSESSMENT_SERVICE_URL` might be incorrect or missing in Railway
4. **Service Discovery**: Railway service discovery might not be working for this specific service
5. **Firewall/Security**: Risk Assessment Service might be blocking requests from API Gateway (unlikely given CORS allows all)

## Configuration Check

### API Gateway Configuration

**File**: `services/api-gateway/internal/config/config.go`
```go
RiskAssessmentURL: getEnvAsString("RISK_ASSESSMENT_SERVICE_URL", 
    "https://risk-assessment-service-production.up.railway.app")
```

**Expected URL**: `https://risk-assessment-service-production.up.railway.app`

### Risk Assessment Service Configuration

**Service Name**: `risk-assessment-service`  
**Public URL**: `https://risk-assessment-service-production.up.railway.app`  
**Port**: `8080`  
**Health Check**: `/health`

## Enhanced Error Logging

**Commit**: Latest (pending)

Added enhanced error logging to `proxyRequest` function:
- Logs target URL, targetPath, and actual error
- Includes error details in response message
- Will help diagnose the exact connection failure

## Diagnostic Steps

### 1. Verify Environment Variable

**In Railway Dashboard**:
1. Go to API Gateway service
2. Check "Variables" tab
3. Verify `RISK_ASSESSMENT_SERVICE_URL` is set to:
   ```
   https://risk-assessment-service-production.up.railway.app
   ```

### 2. Check API Gateway Logs

**In Railway Dashboard**:
1. Go to API Gateway service
2. View "Logs" tab
3. Look for error messages when accessing `/api/v1/risk/metrics`
4. Check for DNS resolution errors, connection refused, or timeout errors

### 3. Test Network Connectivity

If you have shell access to API Gateway container:
```bash
# Test DNS resolution
nslookup risk-assessment-service-production.up.railway.app

# Test HTTP connection
curl -v https://risk-assessment-service-production.up.railway.app/health
```

### 4. Compare with Working Services

Compare Risk Assessment Service configuration with Classification Service (which works):
- Service URLs
- Railway service settings
- Network configuration
- Health check settings

## Potential Solutions

### Solution 1: Verify Environment Variable

Ensure `RISK_ASSESSMENT_SERVICE_URL` is correctly set in Railway API Gateway service variables.

### Solution 2: Use Railway Internal Service Discovery

If Railway supports internal service discovery, try using service name instead of public URL:
```
RISK_ASSESSMENT_SERVICE_URL=http://risk-assessment-service:8080
```

**Note**: This would require Railway internal networking, which may not be available.

### Solution 3: Check Railway Service Status

Verify in Railway dashboard:
- Risk Assessment Service is running
- Service is not in a failed/restarting state
- Network policies allow inter-service communication

### Solution 4: Manual Redeploy

Try manually redeploying both services:
1. Redeploy API Gateway service
2. Redeploy Risk Assessment Service
3. Test connectivity again

## Next Steps

1. ⏳ **PENDING**: Check Railway API Gateway logs for detailed error messages
2. ⏳ **PENDING**: Verify `RISK_ASSESSMENT_SERVICE_URL` environment variable
3. ⏳ **PENDING**: Compare Risk Assessment Service configuration with working services
4. ⏳ **PENDING**: Test after enhanced error logging is deployed

---

**Status**: ⚠️ **CONNECTIVITY ISSUE** - Service works directly but API Gateway cannot reach it

**Action Required**: Check Railway dashboard for API Gateway logs and environment variables

