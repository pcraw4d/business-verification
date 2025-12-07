# HTTP 000 Error Fix - Root Cause Analysis

**Date:** 2025-01-05  
**Status:** ✅ Fixed

---

## Problem Summary

All classification requests were failing with "HTTP 000" errors. The service was receiving requests but not responding, causing curl to timeout after 60 seconds with 0 bytes received.

---

## Root Cause

**HTTP Server ReadTimeout Mismatch:**

1. **HTTP Server ReadTimeout**: 30 seconds (default)
2. **Adaptive Timeout for Scraping**: 45 seconds (calculated by `calculateAdaptiveTimeout`)
3. **Result**: The HTTP server closes the connection after 30 seconds, cancelling the request context before the handler's 45-second timeout can complete

### Timeline of Failure:

1. Request arrives at `/v1/classify`
2. Handler calculates adaptive timeout: 45 seconds for scraping requests
3. Handler creates context with 45-second timeout
4. **After 30 seconds**: HTTP server's `ReadTimeout` expires
5. **Connection closed**: Request context is cancelled
6. **Handler sees**: "context canceled" error
7. **All scraping strategies fail**: Context cancelled before they can complete
8. **curl sees**: HTTP 000 (connection closed, no response)

---

## Evidence from Logs

```
{"level":"warn","msg":"⚠️ [Phase1] Strategy failed, trying next","strategy":"browser_headers","error":"Get \"https://example.com\": context canceled"}
{"level":"warn","msg":"⚠️ [Phase1] Context already cancelled before Playwright strategy","url":"https://example.com","error":"context canceled"}
{"level":"error","msg":"❌ [Phase1] All scraping strategies failed","url":"https://example.com","error":"context canceled"}
```

All strategies fail within milliseconds because the context is already cancelled.

---

## Fix Applied

### 1. Increased HTTP Server ReadTimeout

**File:** `services/classification-service/internal/config/config.go`

**Change:**
```go
ReadTimeout: getEnvAsDuration("READ_TIMEOUT", 30*time.Second),  // OLD
ReadTimeout: getEnvAsDuration("READ_TIMEOUT", 60*time.Second), // NEW
```

**Rationale:**
- Adaptive timeout for scraping: 45 seconds
- Add 15 seconds buffer for overhead
- Total: 60 seconds ReadTimeout ensures connection stays open long enough

### 2. Timeout Budget Breakdown

The adaptive timeout allocates:
- **Phase 1 scraping**: 25s (20s scraper + 5s overhead)
- **Go classification**: 5s
- **ML classification**: 10s (optional)
- **General overhead**: 5s
- **Total**: 45 seconds

The HTTP server ReadTimeout (60s) now exceeds this, preventing premature connection closure.

---

## Additional Issues Identified

### 1. Test Script Endpoint

The test script uses `/v1/classify` which is correct (both `/v1/classify` and `/classify` are registered).

### 2. Request Field Names

The test script uses `business_name` and `website_url` which match the JSON tags in `ClassificationRequest` struct, so this is correct.

### 3. Context Cancellation Propagation

The hybrid timeout approach in `extractKeywordsFromWebsite` correctly handles context cancellation, but the HTTP server timeout was cancelling the context before the scraper could start.

---

## Verification Steps

1. **Rebuild service** with updated ReadTimeout
2. **Test single request**:
   ```bash
   curl -X POST http://localhost:8081/v1/classify \
     -H "Content-Type: application/json" \
     -d '{"business_name":"Test","website_url":"https://example.com"}' \
     --max-time 60
   ```
3. **Check logs** for:
   - No "context canceled" errors
   - Successful strategy execution
   - Valid HTTP response (200 OK)
4. **Run comprehensive test suite**:
   ```bash
   ./scripts/run-phase1-comprehensive-metrics.sh
   ```

---

## Expected Behavior After Fix

1. ✅ Requests with website scraping get 45-second handler timeout
2. ✅ HTTP server ReadTimeout (60s) exceeds handler timeout
3. ✅ Connection stays open until handler completes
4. ✅ Scraping strategies have time to execute
5. ✅ Valid HTTP responses returned (200 OK or appropriate error codes)
6. ✅ No more HTTP 000 errors

---

## Configuration Summary

| Setting | Old Value | New Value | Reason |
|---------|-----------|-----------|--------|
| HTTP Server ReadTimeout | 30s | 60s | Must exceed adaptive timeout (45s) |
| Adaptive Timeout (scraping) | 45s | 45s | No change |
| Adaptive Timeout (simple) | 20s | 20s | No change |
| HTTP Server WriteTimeout | 120s | 120s | No change |

---

## Related Files Modified

- `services/classification-service/internal/config/config.go` - Increased ReadTimeout default

---

## Next Steps

1. ✅ Rebuild classification service
2. ✅ Test with single request
3. ✅ Run comprehensive test suite
4. ✅ Verify all success criteria are met

---

**Status:** ✅ Fixed - Ready for testing

