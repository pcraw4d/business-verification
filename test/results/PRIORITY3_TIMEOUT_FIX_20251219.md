# Priority 3: Website Scraping Timeouts - Fix Implementation
## December 19, 2025

---

## Problem Summary

**Issue**: 29% of requests with website URLs are timing out (Target: <5%)

**Root Cause**:
- Middleware timeout: **30 seconds** (too short)
- Adaptive timeout: **86 seconds** (correctly calculated for website scraping)
- **Mismatch**: Middleware times out at 30s before adaptive timeout (86s) can be used

---

## Fixes Implemented

### Fix 1: Increase Middleware Timeout ✅

**File**: `services/classification-service/cmd/main.go` (line 275)

**Change**:
```go
// Before:
router.Use(timeoutMiddleware(30 * time.Second)) // Fixed 30s timeout

// After:
router.Use(timeoutMiddleware(120 * time.Second)) // Increased from 30s to 120s for website scraping support
```

**Rationale**:
- 120s matches worker pool timeout
- Handler already creates fresh context if <90s remaining
- Adaptive timeout calculation is correct (86s for website scraping)
- Provides safety margin for network delays

### Fix 2: Enhanced Timeout Monitoring ✅

**File**: `services/classification-service/cmd/main.go` (line 637)

**Changes**:
- Added timeout monitoring and logging
- Log slow requests (>30s) for monitoring
- Log timeout events with duration and timeout values
- Track request path and method for debugging

**Implementation**:
```go
// Log slow requests (>30s) for monitoring
if duration > 30*time.Second {
    log.Printf("⏱️ [TIMEOUT-MIDDLEWARE] Slow request completed: %s %s (duration: %v, timeout: %v)",
        r.Method, requestPath, duration, timeout)
}

// Log timeout events
log.Printf("❌ [TIMEOUT-MIDDLEWARE] Request timeout: %s %s (duration: %v, timeout: %v)",
    r.Method, requestPath, duration, timeout)
```

---

## Timeout Configuration Summary

| Layer | Before | After | Purpose |
|-------|--------|-------|---------|
| Middleware | 30s | **120s** ✅ | Request-level timeout |
| Handler Context | 90s+ | 90s+ | Handler processing |
| Adaptive Timeout | 86s | 86s | Website scraping requests |
| Worker Context | 120s | 120s | Worker processing |

---

## Expected Impact

### Before Fix
- Requests with website URLs: **Timeout at 30s** (middleware)
- Requests without URLs: Complete successfully (<30s)
- **Timeout Rate**: 29%

### After Fix
- Requests with website URLs: **Complete in 60-90s** (adaptive timeout)
- Requests without URLs: Complete in <30s
- **Expected Timeout Rate**: <5%

---

## Testing

### Test Script
Created: `test/scripts/test_website_timeout.sh`

**Test Cases**:
1. Microsoft Corporation (https://www.microsoft.com)
2. Apple Inc (https://www.apple.com)
3. Amazon (https://www.amazon.com)

**Expected Results**:
- All requests complete successfully within 120s
- No timeout errors (HTTP 408/504)
- Success rate: 100%

---

## Files Modified

1. `services/classification-service/cmd/main.go`
   - Increased middleware timeout from 30s to 120s
   - Added timeout monitoring and logging

2. `test/scripts/test_website_timeout.sh`
   - Created test script for website timeout verification

3. Documentation
   - `test/results/PRIORITY3_TIMEOUT_ANALYSIS_20251219.md`
   - `test/results/PRIORITY3_TIMEOUT_FIX_20251219.md` (this file)

---

## Next Steps

1. ✅ **Fix Implemented** (this document)
2. ⏳ **Test** with website URL requests
3. ⏳ **Deploy** to Railway
4. ⏳ **Verify** timeout rate improvement (<5%)
5. ⏳ **Monitor** timeout events in logs

---

## Monitoring

### Log Patterns to Watch

**Slow Requests** (>30s):
```
⏱️ [TIMEOUT-MIDDLEWARE] Slow request completed: POST /v1/classify (duration: 45s, timeout: 120s)
```

**Timeout Events**:
```
❌ [TIMEOUT-MIDDLEWARE] Request timeout: POST /v1/classify (duration: 120s, timeout: 120s)
```

**Adaptive Timeout Calculation**:
```
⏱️ [TIMEOUT] Calculated adaptive timeout
request_id: xxx
request_timeout: 86s
has_website_url: true
```

---

## Risk Assessment

**Low Risk**:
- Only increases timeout, doesn't change business logic
- Handler already handles context expiration gracefully
- Worker pool timeout already set to 120s

**Mitigation**:
- Monitor timeout events in logs
- Track timeout rate over time
- Alert if timeout rate increases unexpectedly

---

**Status**: ✅ **FIX IMPLEMENTED**

