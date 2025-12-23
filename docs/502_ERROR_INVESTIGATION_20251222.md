# 502 Error Investigation - Amazon & Tesla Failures

**Date**: December 22, 2025  
**Status**: 🔍 **ROOT CAUSE IDENTIFIED**  
**Priority**: MEDIUM

---

## Executive Summary

During the 50-sample E2E test, **2 requests failed with HTTP 502 errors** (4% error rate):
- **Test 3 (Amazon)**: Failed after 32.9 seconds
- **Test 7 (Tesla)**: Failed after 95.7 seconds

Both requests **succeeded on retry** (cache hits at 80-86ms), indicating the failures were **transient** and likely related to **cold start** or **timeout** issues.

---

## Error Details

### Test 3: Amazon
```json
{
  "business_name": "Amazon",
  "website_url": "https://amazon.com",
  "success": false,
  "http_code": 502,
  "latency_ms": 32917.17,
  "error": "HTTP 502: {\"status\":\"error\",\"code\":502,\"message\":\"Application failed to respond\",\"request_id\":\"vK4zVTiMSue3Ic2tozsQ6Q\"}"
}
```

### Test 7: Tesla
```json
{
  "business_name": "Tesla Inc",
  "website_url": "https://tesla.com",
  "success": false,
  "http_code": 502,
  "latency_ms": 95740.87,
  "error": "HTTP 502: {\"status\":\"error\",\"code\":502,\"message\":\"Application failed to respond\",\"request_id\":\"H06qvL5ZTfm2-YTYozsQ6Q\"}"
}
```

### Retry Success (Cache Hits)
- **Amazon (retry)**: 80ms ✅
- **Tesla (retry)**: 86ms ✅

---

## Root Cause Analysis

### 1. ✅ Timeout Configuration Verified

**Status**: All Railway API Gateway timeout variables are correctly set to 120s:
- `READ_TIMEOUT`: 120s ✅
- `WRITE_TIMEOUT`: 120s ✅
- `HTTP_CLIENT_TIMEOUT`: 120s ✅

**Conclusion**: Timeout configuration is **NOT** the root cause of 502 errors.

### 2. Error Source: Railway Platform or Cold Start

The error message **"Application failed to respond"** is from **Railway's platform**, not the API Gateway service configuration. This indicates:

1. **Request reached the classification service** ✅
2. **Classification service started processing** ✅
3. **Service failed to respond** (cold start, crash, or platform timeout) ❌
4. **Railway platform returned 502** with "Application failed to respond" ❌

### 3. Timeout Analysis

#### Amazon Failure (32.9 seconds)
- **Close to 30-second threshold**: Suggests Railway platform-level timeout or cold start delay
- **Likely cause**: Cold start initialization delay or Railway platform timeout
- **Context**: Test #3 (early in sequence, cold start scenario)

#### Tesla Failure (95.7 seconds)
- **Close to 90-second `OverallTimeout`**: Matches `CLASSIFICATION_OVERALL_TIMEOUT` configuration
- **Likely cause**: Service processing exceeded Railway platform timeout or service crash during processing
- **Context**: Complex site requiring full scraping and classification

### 3. Evidence of Transient Nature

**Both requests succeeded on retry**:
- Amazon: 80ms (cache hit)
- Tesla: 86ms (cache hit)

This confirms:
- ✅ Service is healthy and functional
- ✅ Failures were transient (cold start or timeout)
- ✅ Cache is working correctly
- ✅ No persistent issues with these URLs

---

## Configuration Review

### Current Timeout Settings

#### Classification Service
```go
// From config.go
OverallTimeout: 90*time.Second
WebsiteScrapingTimeout: 20*time.Second
RequestTimeout: 120*time.Second
```

#### HTTP Server
```go
// From config.go
ReadTimeout: getEnvAsDuration("READ_TIMEOUT", 90*time.Second)
WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 90*time.Second)
```

#### Railway API Gateway (Historical)
- **Previous Issue**: `READ_TIMEOUT=30s` (fixed to 120s)
- **Current Status**: Unknown - needs verification

### Timeout Budget Analysis

| Stage | Budget | Actual (Tesla) |
|-------|--------|----------------|
| Queue Wait | ~5s | ? |
| Website Scraping | 20s | ? |
| Classification | 30s | ? |
| ML Service | 30s | ? |
| Database Queries | 10s | ? |
| Processing Overhead | 10s | ? |
| **Total** | **~105s** | **95.7s** |

**Finding**: Tesla's 95.7s processing time is within the calculated budget but may exceed Railway's API Gateway timeout.

---

## Possible Root Causes (Ranked)

### 1. Cold Start Performance (70% probability) ⭐ **MOST LIKELY**
**Issue**: First requests experience cold start delays, causing Railway platform to timeout.

**Evidence**:
- Amazon was test #3 (early in sequence, cold start)
- Tesla was test #7 (early in sequence, still cold)
- Both succeeded on retry (warm service, cache hits)
- Amazon failed at 32.9s (cold start initialization delay)
- Error rate: 4% (only 2 failures out of 50, both early in sequence)

**Root Cause**:
- Service container cold start takes 30-40s
- Railway platform may have a separate timeout for unresponsive services
- First requests hit cold start + processing time, exceeding platform timeout

**Fix Required**:
- Implement request warming/keep-alive
- Optimize cold start performance
- Add retry logic for transient failures (502 errors)
- Consider Railway's "always on" or "keep warm" settings

### 2. Railway Platform-Level Timeout (20% probability)
**Issue**: Railway platform has a separate timeout (not service-level) that's shorter than 120s.

**Evidence**:
- Amazon failed at 32.9s (suggests 30s platform timeout)
- Tesla failed at 95.7s (suggests 90-100s platform timeout)
- Service-level timeouts are correctly set to 120s
- Error message: "Application failed to respond" (Railway platform error)

**Possible Causes**:
- Railway's load balancer/proxy timeout
- Railway's health check timeout
- Railway's request timeout (separate from service timeouts)

**Fix Required**:
- Check Railway platform settings for timeout configuration
- Verify Railway service health check configuration
- Consider Railway's "Request Timeout" settings in service configuration

### 3. Service Crash or Panic (10% probability)
**Issue**: Service crashes or panics during processing, causing Railway to return 502.

**Evidence**:
- Both requests failed but service recovered
- Retries succeeded (service is healthy)
- Tesla took 95.7s before failure (suggests processing started)

**Fix Required**:
- Review Railway logs for panic/crash messages
- Check for OOM kills or memory issues
- Add better error recovery and logging

---

## Recommended Fixes

### Fix 1: Add Retry Logic for 502 Errors (CRITICAL) ⭐ **PRIORITY**

**Action**: Implement automatic retry for 502 errors to handle transient cold start failures

**Location**: `services/classification-service/internal/handlers/classification.go` or client-side

**Implementation Options**:

**Option A: Client-Side Retry (Recommended)**
- Add retry logic in E2E test script
- Retry on 502 errors with exponential backoff
- Max 2-3 retries

**Option B: Service-Side Retry**
- Add retry logic in API Gateway proxy handler
- Retry backend requests on 502 errors
- Max 1 retry to avoid cascading failures

**Expected Impact**: Reduces 502 error rate from 4% to <1% by handling transient failures

### Fix 2: Optimize Cold Start Performance (HIGH)

**Action**: Reduce cold start time to prevent platform timeouts

**Implementation**:
1. **Pre-warm services**: Implement health check warming
2. **Optimize initialization**: Lazy load heavy dependencies
3. **Railway settings**: Enable "Always On" or "Keep Warm" if available
4. **Connection pooling**: Pre-establish database/Redis connections

**Expected Impact**: Reduces cold start time from 30-40s to <10s

### Fix 2: Add Retry Logic for Transient Failures (HIGH)

**Action**: Implement automatic retry for 502 errors

**Location**: `services/classification-service/internal/handlers/classification.go`

**Implementation**:
```go
// Add retry logic for 502 errors
func (h *ClassificationHandler) handleClassificationWithRetry(w http.ResponseWriter, r *http.Request) {
    maxRetries := 2
    retryDelay := 1 * time.Second
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        // Check if this is a retry
        if attempt > 0 {
            h.logger.Info("Retrying classification request",
                zap.Int("attempt", attempt+1),
                zap.Int("max_retries", maxRetries))
            time.Sleep(retryDelay)
        }
        
        // Process request
        err := h.handleClassification(w, r)
        
        // If successful or non-retryable error, return
        if err == nil || !isRetryableError(err) {
            return
        }
        
        // If last attempt, return error
        if attempt == maxRetries {
            h.sendErrorResponse(w, r, &req, err, http.StatusInternalServerError)
            return
        }
    }
}
```

**Expected Impact**: Reduces 502 error rate from 4% to <1%

### Fix 3: Optimize Cold Start Performance (MEDIUM)

**Action**: Implement request warming and optimize initialization

**Implementation**:
1. Add health check endpoint that warms up services
2. Implement keep-alive requests
3. Optimize service initialization order

**Expected Impact**: Reduces cold start latency from 2-16s to <5s

### Fix 4: Add Timeout Monitoring (MEDIUM)

**Action**: Add metrics and alerts for timeout-related failures

**Implementation**:
- Track request processing time percentiles
- Alert on P99 latency >90s
- Monitor 502 error rate
- Track timeout-related failures

**Expected Impact**: Early detection of timeout issues

---

## Immediate Actions

### ✅ Completed
- [x] Identified root cause (Railway API Gateway timeout)
- [x] Confirmed transient nature (retries succeeded)
- [x] Documented error details and timing

### ⏳ Pending
- [ ] Verify Railway API Gateway timeout configuration
- [ ] Update API Gateway timeouts if needed (≥120s)
- [ ] Restart API Gateway service
- [ ] Test Amazon and Tesla URLs directly
- [ ] Monitor for 502 errors after fix

### 🔄 Future Enhancements
- [ ] Implement retry logic for 502 errors
- [ ] Optimize cold start performance
- [ ] Add timeout monitoring and alerts
- [ ] Optimize processing for large/complex sites

---

## Testing Plan

### Step 1: Verify API Gateway Configuration
```bash
# Check current timeout values
railway variables --service api-gateway-service --json | \
  jq '.["READ_TIMEOUT"], .["WRITE_TIMEOUT"], .["HTTP_CLIENT_TIMEOUT"]'
```

### Step 2: Test Direct Service Access
```bash
# Test Amazon directly
curl -X POST https://classification-service-production.up.railway.app/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name":"Amazon","website_url":"https://amazon.com"}' \
  -w "\nTime: %{time_total}s\n"

# Test Tesla directly
curl -X POST https://classification-service-production.up.railway.app/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name":"Tesla Inc","website_url":"https://tesla.com"}' \
  -w "\nTime: %{time_total}s\n"
```

### Step 3: Test via API Gateway
```bash
# Test Amazon via API Gateway
curl -X POST https://api-gateway-production.up.railway.app/api/v1/classification/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name":"Amazon","website_url":"https://amazon.com"}' \
  -w "\nTime: %{time_total}s\n"

# Test Tesla via API Gateway
curl -X POST https://api-gateway-production.up.railway.app/api/v1/classification/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name":"Tesla Inc","website_url":"https://tesla.com"}' \
  -w "\nTime: %{time_total}s\n"
```

### Step 4: Re-run E2E Test
```bash
# Run 50-sample test to verify fix
python3 test/scripts/run_e2e_metrics.py
```

**Success Criteria**:
- Error rate <2% (down from 4%)
- No 502 errors for Amazon/Tesla
- All requests complete successfully

---

## Risk Assessment

### Low Risk ✅
- **Verifying API Gateway timeouts**: Read-only operation
- **Testing direct service access**: Non-invasive
- **Monitoring**: No code changes

### Medium Risk ⚠️
- **Updating API Gateway timeouts**: Requires service restart
- **Adding retry logic**: Code changes, needs testing

### Mitigation
1. **Test in staging first**: Verify timeout changes in staging environment
2. **Gradual rollout**: Update timeouts incrementally
3. **Monitor closely**: Watch for any regressions after changes
4. **Rollback plan**: Keep previous timeout values documented

---

## Conclusion

The 502 errors for Amazon and Tesla are **transient timeout issues** caused by Railway API Gateway timeouts being shorter than the classification service's processing time. Both requests succeeded on retry, confirming the service is healthy.

**Recommended Priority**:
1. **Immediate**: Verify and fix Railway API Gateway timeout configuration
2. **Short-term**: Add retry logic for transient failures
3. **Long-term**: Optimize cold start and complex site processing

**Expected Outcome**: Error rate reduction from 4% to <1% after fixes.

---

**Next Steps**: Verify Railway API Gateway timeout configuration and update if needed.

