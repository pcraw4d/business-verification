# Root Cause Investigation - Comprehensive Analysis

**Date:** 2025-12-06  
**Status:** Investigation Complete - Ready for Remediation

---

## Issue 1: Context Deadline Too Short (CRITICAL)

### Investigation Findings

**Observed Behavior:**
- Requests enter `ClassifyBusiness` with only ~5 seconds remaining
- Adaptive timeout calculates 45s (15s Phase 1 + 10s multi-page + 5s Go + 10s ML + 5s overhead)
- Gap: ~40 seconds consumed before `ClassifyBusiness` is called

**Code Flow Analysis:**
```
HandleClassification (line 352)
  ↓ calculateAdaptiveTimeout (line 418) → Returns 45s
  ↓ context.WithTimeout(ctx, 45s) (line 463)
  ↓ processClassification (line 1022)
    ↓ generateEnhancedClassification (line 1034)
      ↓ [Operations before ClassifyBusiness - CONSUMING ~40s]
      ↓ ClassifyBusiness (line 1880) → Only ~5s remaining!
```

**Root Cause Identified:**
1. **`BuildKeywordIndex` is called synchronously** in `ClassifyBusinessByContextualKeywords` (line 2384)
   - This can take 10-30 seconds on first call
   - Happens BEFORE `extractKeywords` is called
   - Not accounted for in adaptive timeout calculation

2. **Index check happens in `ClassifyBusinessByContextualKeywords`** (line 2031)
   - Called AFTER `extractKeywords` but BEFORE classification
   - If index is empty, `BuildKeywordIndex` is called synchronously
   - This consumes significant time

3. **Adaptive timeout doesn't account for index building**
   - Current budget: 45s total
   - Index building: 10-30s (not accounted for)
   - Result: Only ~5s remains for actual operations

**Evidence:**
- Profiling logs show "Before extractKeywords - time remaining: 4.3s"
- But only 659ms elapsed since ClassifyBusiness entry
- This means ~40s was consumed BEFORE ClassifyBusiness was called

**Remediation Plan:**
1. Add `indexBuildingBudget` to adaptive timeout calculation (30s)
2. Or: Make index building asynchronous/non-blocking
3. Or: Pre-build index on service startup
4. Increase total adaptive timeout to 75s (45s current + 30s index building)

---

## Issue 2: Playwright Browser Pool Exhaustion (CRITICAL)

### Investigation Findings

**Observed Behavior:**
- Queue has 26,097 requests waiting
- Queue wait times: 4+ million milliseconds (~1+ hour)
- All requests fail with "Timeout waiting for available browser"
- Service status: unhealthy

**Code Analysis:**
```javascript
// Browser acquisition (line 140-184)
acquire(timeout = 5000) {
  // Tries to find available browser
  // Timeout: 5 seconds
}

// Browser release (line 186-192)
release(poolItem) {
  poolItem.inUse = false;
  this.inUse.delete(poolItem.browser);
}

// Request handling (line 487-629)
queue.add(async () => {
  try {
    poolItem = await browserPool.acquire(5000);
    // ... scrape logic ...
  } finally {
    browserPool.release(poolItem); // Line 626
  }
}, { throwOnTimeout: true });
```

**Root Cause Identified:**
1. **Queue timeout (25s) > Browser acquisition timeout (5s)**
   - If all browsers are in use, `acquire()` times out after 5s
   - But request stays in queue for 25s
   - When queue timeout fires, `throwOnTimeout: true` causes error
   - The `finally` block may not execute properly if queue.add() throws synchronously

2. **Browser release happens in `finally` block**
   - If `queue.add()` throws a timeout error, the finally block should still execute
   - BUT: If the error is thrown synchronously (not awaited), finally might not run
   - Need to verify p-queue behavior with `throwOnTimeout: true`

3. **Potential race condition in `acquire()`**
   - Simple lock (`acquireLock`) may not prevent all race conditions
   - Multiple requests could acquire the same browser if timing is wrong

4. **Browser pool size may be insufficient**
   - Pool size: 3 browsers (default)
   - Max concurrent: 8 requests
   - If browsers take >25s to complete, pool gets exhausted

**Remediation Plan:**
1. **Fix browser release guarantee:**
   - Wrap entire queue.add() in try-finally to ensure release
   - Add explicit release in catch block for queue timeout errors

2. **Increase browser acquisition timeout:**
   - Match browser acquisition timeout to queue timeout (25s)
   - Or: Make acquisition timeout configurable

3. **Add browser pool monitoring:**
   - Log browser pool state (available/in-use)
   - Alert when pool is exhausted

4. **Consider increasing pool size:**
   - Current: 3 browsers
   - Recommended: 5-8 browsers for 8 concurrent requests

---

## Issue 3: HTTP 429 Rate Limiting (HIGH)

### Investigation Findings

**Observed Behavior:**
- Many sites returning `429 Too Many Requests`
- Affects SimpleHTTP and BrowserHeaders strategies
- Forces fallback to Playwright (which is also failing)

**Root Cause Identified:**
1. **No rate limiting per domain**
   - Multiple concurrent requests to same domain
   - No delay between requests
   - User-Agent may be flagged as bot

2. **No exponential backoff for 429 errors**
   - Immediate retry on 429
   - No delay or backoff strategy

**Remediation Plan:**
1. Implement per-domain rate limiting
2. Add exponential backoff for 429 errors
3. Rotate User-Agents
4. Add delays between requests to same domain

---

## Issue 4: Playwright Client Timeout (HIGH)

### Investigation Findings

**Observed Behavior:**
- Playwright HTTP client times out after ~20-22 seconds
- Error: "context deadline exceeded (Client.Timeout exceeded while awaiting headers)"
- Happens even when Playwright service is healthy

**Root Cause Identified:**
1. **HTTP client timeout (20s) < Queue timeout (25s)**
   - If request waits 20s in queue, client timeout fires
   - Service hasn't even started processing yet

2. **Context deadline propagation**
   - Context deadline from classification service: 20s
   - But queue wait can be 20s+
   - Client timeout fires before service responds

**Remediation Plan:**
1. Increase HTTP client timeout to account for queue wait
2. Make client timeout dynamic based on queue metrics
3. Or: Return 503 immediately if queue is full (fail fast)

---

## Issue 5: Negative Time Calculation (MEDIUM)

### Investigation Findings

**Observed Behavior:**
- Context deadline checks show negative time remaining
- Example: "-1h0m49.321845981s < 6s required"

**Root Cause Identified:**
1. **Context already expired when checked**
   - Context deadline is in the past
   - `time.Until(deadline)` returns negative value
   - Indicates context was created with deadline in past, or clock issue

2. **Possible causes:**
   - Context created with expired deadline
   - Clock/timezone mismatch
   - Context deadline set incorrectly

**Remediation Plan:**
1. Add defensive check: if `time.Until(deadline) < 0`, skip operation
2. Log warning when negative time detected
3. Investigate context creation to ensure deadlines are in future

---

## Priority Remediation Order

1. **CRITICAL: Fix Context Time Budget** (Issue 1)
   - Add index building budget to adaptive timeout
   - Increase total timeout to 75s

2. **CRITICAL: Fix Playwright Browser Pool** (Issue 2)
   - Ensure browser release in all error paths
   - Increase browser acquisition timeout
   - Add monitoring

3. **HIGH: Fix Playwright Client Timeout** (Issue 4)
   - Increase client timeout to account for queue wait
   - Or implement fail-fast for full queue

4. **HIGH: Implement Rate Limiting** (Issue 3)
   - Add per-domain rate limiting
   - Implement exponential backoff

5. **MEDIUM: Fix Negative Time Calculation** (Issue 5)
   - Add defensive checks
   - Log warnings

---

## Implementation Plan

### Phase 1: Critical Fixes (Issues 1 & 2)
1. Increase adaptive timeout to 75s (add 30s for index building)
2. Fix Playwright browser release guarantee
3. Increase browser acquisition timeout to 25s

### Phase 2: High Priority Fixes (Issues 3 & 4)
1. Implement rate limiting strategy
2. Fix Playwright client timeout

### Phase 3: Medium Priority (Issue 5)
1. Add defensive checks for negative time
2. Improve logging

---

**Next Steps:** Implement fixes in priority order, starting with Phase 1.

