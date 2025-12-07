# Comprehensive Test Log Analysis

**Date:** 2025-01-06  
**Test Run:** After implementing fixes for context deadline, Playwright pool, and client timeout  
**Status:** ❌ Critical Issues Identified

---

## Executive Summary

The comprehensive test suite shows **25.00% success rate** (11/44), with **75.00% failures** (33/44). While this is a slight improvement from the previous 20.45%, the root causes remain unresolved. The primary issue is that **`BuildKeywordIndex` is consuming 1-3 minutes**, far exceeding the 30s budget allocated in the adaptive timeout calculation.

---

## Critical Findings

### 1. BuildKeywordIndex Duration (CRITICAL)

**Issue:** `ClassifyBusinessByContextualKeywords` is taking **1-3 minutes** instead of the expected 10-30 seconds.

**Evidence from Logs:**
```
classify_duration: 1m10.103359423s  (70 seconds)
classify_duration: 2m46.402755382s  (166 seconds!)
classify_duration: 35.156209339s
classify_duration: 32.573243687s
classify_duration: 1m23.901693449s  (83 seconds)
classify_duration: 3m12.026112383s  (192 seconds!)
```

**Impact:**
- The adaptive timeout allocates **30s** for `indexBuildingBudget`
- Actual duration is **60-180 seconds** (2-6x the budget)
- This consumes the **entire context deadline** before `extractKeywords()` even starts
- Requests enter `ClassifyBusiness` with **negative time remaining** (already expired)

**Root Cause:**
- `BuildKeywordIndex` is likely performing expensive database queries
- May be building the index synchronously for each request
- No caching or optimization of index building

**Evidence of Expired Contexts:**
```
⏱️ [PROFILING] ClassifyBusiness entry - time remaining: -17.298790037s
⏱️ [PROFILING] ClassifyBusiness entry - time remaining: -398.469474ms
⏱️ [PROFILING] ClassifyBusiness entry - time remaining: -1.641965649s
⏱️ [PROFILING] Before extractKeywords - time remaining: -5.856696113s
⏱️ [PROFILING] Before extractKeywords - time remaining: -24.128132382s
```

---

### 2. Playwright Browser Pool Exhaustion (HIGH)

**Issue:** Browsers are timing out during acquisition, despite the 25s timeout increase.

**Evidence from Logs:**
```
"error":"Timeout waiting for available browser"
"scrapeDurationMs":25824,"totalDurationMs":25884,"queueWaitTimeMs":0
"scrapeDurationMs":25715,"totalDurationMs":25729,"queueWaitTimeMs":10
"scrapeDurationMs":39418,"totalDurationMs":39497,"queueWaitTimeMs":61
"scrapeDurationMs":38835,"totalDurationMs":38844,"queueWaitTimeMs":1
"scrapeDurationMs":32499,"totalDurationMs":38688,"queueWaitTimeMs":6186
"scrapeDurationMs":36742,"totalDurationMs":56645,"queueWaitTimeMs":19794
"scrapeDurationMs":35908,"totalDurationMs":51843,"queueWaitTimeMs":15843
"scrapeDurationMs":70625,"totalDurationMs":102364,"queueWaitTimeMs":31726
```

**Observations:**
- Browsers ARE being released (good!): `"Browser released to pool after queue error"`
- Queue wait times are **0-31 seconds** (within the 25s timeout)
- Browser acquisition is still timing out after 25s
- Some requests wait **70+ seconds** total (queue + scrape)

**Possible Causes:**
1. **All 3 browsers are in use** and requests are queued
2. **Browser acquisition mutex is blocking** longer than expected
3. **Browsers are dead/crashed** and not being recovered quickly enough
4. **Queue is processing too slowly** (browsers taking 25s+ to complete scrapes)

**Browser Release Logs (Positive):**
```
"message":"Browser released from pool","browserId":0,"poolSize":3,"inUseCount":1
"message":"Browser released to pool after queue error","browserId":0,"error":"Promise timed out after 25000 milliseconds"
```

---

### 3. Context Deadline Exceeded Errors (HIGH)

**Issue:** All scraping strategies are failing with "context deadline exceeded" errors.

**Evidence from Logs:**
```
"error":"Get \"https://www.target.com\": context deadline exceeded"
"error":"Post \"http://playwright-scraper:3000/scrape\": context deadline exceeded"
"error":"context deadline exceeded (Client.Timeout exceeded while awaiting headers)"
"error":"context deadline exceeded (Client.Timeout or context cancellation while reading body)"
```

**Pattern:**
- SimpleHTTP: Failing with context deadline exceeded
- BrowserHeaders: Failing with context deadline exceeded
- Playwright: Failing with context deadline exceeded

**Root Cause:**
- Contexts are **already expired** when `extractKeywords()` is called
- `BuildKeywordIndex` consumed the entire 75s timeout
- No time remains for scraping operations

---

### 4. Strategy Success/Failure Analysis

**Successful Strategies:**
- SimpleHTTP: Some successes (quality_score: 0.8, word_count: 328)
- BrowserHeaders: Mostly failures (quality_score: 0.15, word_count: 12)
- Playwright: All failures (context deadline exceeded)

**Failed Strategy Details:**
```
⚠️ [Phase1] Strategy failed, trying next
  strategy: simple_http
  quality_score: 0.3
  word_count: 49
  meets_word_count: false
  meets_quality_threshold: false

⚠️ [Phase1] Strategy failed, trying next
  strategy: browser_headers
  quality_score: 0.15
  word_count: 12
  meets_word_count: false
  meets_quality_threshold: false

⚠️ [Phase1] Strategy failed, trying next
  strategy: playwright
  error: "Post \"http://playwright-scraper:3000/scrape\": context deadline exceeded"
```

---

### 5. Test Results Summary

**Success Rate:** 25.00% (11/44)
- **Successful:** 11 requests
- **Failed:** 33 requests

**Successful URLs:**
- https://www.w3.org
- https://www.iana.org
- https://www.microsoft.com
- https://www.apple.com
- https://www.starbucks.com
- https://www.coca-cola.com
- https://www.airbnb.com
- https://www.uber.com
- https://www.linkedin.com
- https://www.shopify.com
- https://www.github.com

**Failed URLs:**
- All other 33 URLs (primarily due to context deadline exceeded)

---

## Root Cause Analysis

### Primary Root Cause: BuildKeywordIndex Duration

**The Problem:**
1. `BuildKeywordIndex` is taking **60-180 seconds** (not 30s as estimated)
2. Adaptive timeout allocates **30s** for index building
3. Actual duration exceeds the **entire 75s timeout**
4. Context expires **before** `extractKeywords()` starts
5. All scraping strategies fail with "context deadline exceeded"

**Why This Happens:**
- `BuildKeywordIndex` likely performs expensive database queries
- May be building the index synchronously for each request
- No caching or optimization
- Index may be rebuilt on every request (or frequently)

**Evidence:**
- `ClassifyBusinessByContextualKeywords` duration: **1-3 minutes**
- Contexts entering `ClassifyBusiness` with **negative time remaining**
- All scraping strategies failing with context deadline exceeded

---

### Secondary Root Cause: Playwright Browser Pool

**The Problem:**
1. Browsers are being released (good!)
2. But browser acquisition is still timing out after 25s
3. Queue wait times are 0-31 seconds
4. Some requests wait 70+ seconds total

**Why This Happens:**
1. **All browsers in use:** 3 browsers may not be enough for concurrent load
2. **Slow browser operations:** Browsers taking 25s+ to complete scrapes
3. **Queue processing too slow:** Queue may not be processing requests fast enough
4. **Browser recovery delays:** Dead browsers may not be recovered quickly

**Evidence:**
- "Timeout waiting for available browser" errors
- Queue wait times: 0-31 seconds
- Total duration: 25-100+ seconds

---

## Recommendations

### Priority 1: Fix BuildKeywordIndex Duration (CRITICAL)

**Actions:**
1. **Profile `BuildKeywordIndex`** to identify slow operations
2. **Cache the keyword index** to avoid rebuilding on every request
3. **Optimize database queries** (add indexes, reduce query complexity)
4. **Build index asynchronously** or in background
5. **Increase `indexBuildingBudget`** in adaptive timeout to **120s** (temporary fix)

**Expected Impact:**
- Reduce index building time from 60-180s to <10s (with caching)
- Free up 50-170 seconds for scraping operations
- Increase success rate from 25% to 80%+

---

### Priority 2: Optimize Playwright Browser Pool (HIGH)

**Actions:**
1. **Increase browser pool size** from 3 to 5-8 browsers
2. **Reduce browser scrape timeout** from 15s to 10s (if possible)
3. **Optimize browser operations** (reduce page load time, faster selectors)
4. **Improve browser recovery** (faster detection and replacement)
5. **Monitor queue metrics** to identify bottlenecks

**Expected Impact:**
- Reduce browser acquisition timeouts
- Improve Playwright strategy success rate
- Reduce queue wait times

---

### Priority 3: Increase Adaptive Timeout (MEDIUM)

**Actions:**
1. **Temporarily increase `indexBuildingBudget`** to **120s** (until caching is implemented)
2. **Increase total timeout** from 75s to **150s** for website scraping scenarios
3. **Monitor timeout usage** to identify optimal values

**Expected Impact:**
- Provide sufficient time for index building
- Allow scraping operations to complete
- Increase success rate

---

## Next Steps

1. **Immediate:** Profile `BuildKeywordIndex` to identify slow operations
2. **Short-term:** Implement keyword index caching
3. **Short-term:** Optimize database queries in `BuildKeywordIndex`
4. **Medium-term:** Increase Playwright browser pool size
5. **Long-term:** Build index asynchronously or in background

---

## Metrics Comparison

| Metric | Previous | Current | Target | Status |
|--------|----------|---------|--------|--------|
| Success Rate | 20.45% | 25.00% | ≥95% | ❌ |
| HTTP 000 Errors | 56.8% | 75.00% | <5% | ❌ |
| HTTP 500 Errors | 13.6% | 0% | <2% | ✅ |
| Playwright Health | Unhealthy | Healthy | Healthy | ✅ |
| Context Deadline Issues | High | Critical | None | ❌ |

---

## Conclusion

The primary issue is **`BuildKeywordIndex` consuming 1-3 minutes**, which exhausts the entire context deadline before scraping operations can begin. This must be addressed immediately through caching and optimization. The Playwright browser pool is functioning but may need optimization for higher concurrency.

**Status:** ❌ Critical issues remain  
**Next Action:** Profile and optimize `BuildKeywordIndex`

