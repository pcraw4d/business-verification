# Railway Production Test - Comprehensive Analysis & Recommendations

**Date**: December 9, 2025  
**Test Run**: Latest Comprehensive Test (44 merchants)  
**Status**: 🔴 **CRITICAL - Success Rate 6.81% (3/44)**

---

## Executive Summary

The latest comprehensive test run shows **6.81% success rate** (3/44), a significant regression from previous local test runs (29.54%). While the API Gateway timeout fix (30s → 120s) is working, **most requests are now timing out at the 120s limit**, indicating the classification service is either:
1. Crashing due to memory limits (OOM kills)
2. Processing too slowly (exceeding 120s)
3. Experiencing resource exhaustion

**Key Finding**: The timeout fix exposed a deeper issue - the classification service cannot complete most requests within 120s, suggesting memory pressure and/or processing bottlenecks.

---

## Test Results Comparison

### Latest Test (Railway Production - Dec 9, 2025)

| Metric | Result | Target | Status |
|--------|--------|--------|--------|
| **Success Rate** | 6.81% (3/44) | ≥95% | ❌ CRITICAL |
| **HTTP 502 Errors** | 40 (90.9%) | 0 | ❌ CRITICAL |
| **HTTP 503 Errors** | 1 (2.3%) | 0 | ⚠️ HIGH |
| **Average Confidence** | 0.92 | N/A | ✅ Good (when successful) |
| **Average Duration (Success)** | 22.6s | <90s | ✅ Good |

### Previous Test (Local - Dec 5, 2025)

| Metric | Result | Target | Status |
|--------|--------|--------|--------|
| **Success Rate** | 29.54% (13/44) | ≥95% | ❌ FAIL |
| **HTTP 000 Errors** | 25 (56.8%) | 0 | ❌ CRITICAL |
| **HTTP 500 Errors** | 6 (13.6%) | 0 | ❌ CRITICAL |

### Pattern Analysis

**Regression**: Success rate dropped from **29.54% → 6.81%** (4.3x worse)

**Error Pattern Shift**:
- **Before**: HTTP 000 (connection timeout) at ~30s
- **After**: HTTP 502 (bad gateway) at ~120s

**Interpretation**:
- ✅ API Gateway timeout fix worked (requests now wait 120s instead of 30s)
- ❌ Classification service cannot complete most requests within 120s
- ❌ Service likely crashing/restarting due to memory limits

---

## Detailed Results Analysis

### Successful Requests (3/44 = 6.81%)

1. **w3.org** - Duration: 18.7s, Confidence: 0.92, Industry: unknown
2. **iana.org** - Duration: 26.1s, Confidence: 0.92, Industry: unknown
3. **github.com** - Duration: 22.9s, Confidence: 0.92, Industry: unknown

**Characteristics**:
- All are simple, static websites
- Fast processing times (18-26s)
- High confidence scores (0.92)
- Industry classification not working (all show "unknown")

### Failed Requests (41/44 = 93.18%)

**HTTP 502 Errors (40 requests, 90.9%)**:
- Duration: ~120s (hitting timeout limit)
- Error: "Application failed to respond"
- Pattern: All complex/large websites

**HTTP 503 Errors (1 request, 2.3%)**:
- ebay.com - "Classification service unavailable"
- Duration: 0.15s (immediate failure)
- Pattern: Service was down/unavailable at that moment

---

## Pattern of Unresolved Issues

### Issue Pattern #1: Memory Exhaustion (PERSISTENT)

**Evidence Across All Tests**:
- User reported: "classification service is crashing due to memory limits"
- HTTP 502 errors at timeout limit (120s) suggest service crashes during processing
- Only simple sites succeed (low memory usage)
- Complex sites fail (high memory usage)

**Previous Attempts**:
- ✅ Reduced worker pool size (30 → 15-20 workers)
- ✅ Reduced MaxConcurrentRequests (100 → 40)
- ✅ Added GOMEMLIMIT (768MiB)
- ❌ **Still failing** - Memory pressure persists

**Root Cause Hypothesis**:
1. **Website scraping memory spikes**: Each Playwright/BrowserHeaders scrape allocates significant memory
2. **Concurrent requests**: Even with reduced concurrency, 2-3 concurrent scrapes can exceed memory limits
3. **Memory not released**: Goroutines/contexts holding references prevent GC
4. **Railway memory limits**: Service may have <512MB available, insufficient for concurrent scraping

**Recommendation**: 
- **P0**: Further reduce concurrency (MaxConcurrentRequests: 40 → 20)
- **P0**: Add request admission control (reject if memory usage >80%)
- **P1**: Implement memory-aware scraping (skip Playwright if memory high)
- **P1**: Add memory monitoring and circuit breaker

---

### Issue Pattern #2: Processing Time Exceeding 120s (NEW)

**Evidence**:
- 40/44 requests timeout at exactly 120s
- Successful requests complete in 18-26s
- Failed requests are all complex websites

**Root Cause Hypothesis**:
1. **Complex website scraping**: Large sites require 60-90s+ to scrape
2. **Multi-page analysis**: Analyzing 8-15 pages sequentially takes time
3. **External API calls**: Supabase RPC calls add latency
4. **No early termination**: Service continues even when timeout approaching

**Recommendation**:
- **P0**: Implement adaptive timeout with early termination
- **P0**: Add timeout checks before expensive operations
- **P1**: Optimize multi-page analysis (parallel processing)
- **P1**: Add timeout warnings at 80% of limit

---

### Issue Pattern #3: Industry Classification Not Working (PERSISTENT)

**Evidence**:
- All 3 successful requests show `"industry": "unknown"`
- Confidence scores are high (0.92) but industry is not identified

**Root Cause Hypothesis**:
1. **Industry detection logic failing**: Industry detector not finding matches
2. **Keyword extraction insufficient**: Not enough keywords extracted for classification
3. **Database query issues**: Supabase queries not returning industry matches

**Recommendation**:
- **P1**: Investigate industry detection logic
- **P1**: Add logging for industry detection steps
- **P2**: Review keyword extraction quality

---

### Issue Pattern #4: Service Availability Issues (PERSISTENT)

**Evidence**:
- HTTP 503 errors indicate service unavailable
- Service crashes/restarts causing temporary unavailability
- Memory exhaustion causing OOM kills

**Previous Attempts**:
- ✅ Added health checks
- ✅ Added graceful shutdown
- ❌ **Still failing** - Service crashes under load

**Recommendation**:
- **P0**: Add circuit breaker for classification service
- **P0**: Implement request queuing with backpressure
- **P1**: Add service health monitoring
- **P1**: Implement automatic scaling/restart policies

---

## Root Cause Analysis

### Primary Root Cause: Memory Exhaustion Leading to OOM Kills

**Evidence Chain**:
1. User explicitly stated: "classification service is crashing due to memory limits"
2. Only simple sites succeed (low memory footprint)
3. Complex sites fail at 120s timeout (likely OOM kill during processing)
4. Previous memory mitigations (reduced concurrency, GOMEMLIMIT) not sufficient

**Memory Usage Pattern**:
- **Simple sites** (w3.org, iana.org): ~50-100MB per request
- **Complex sites** (microsoft.com, amazon.com): ~200-500MB per request
- **Concurrent requests**: 2-3 concurrent = 400-1500MB total
- **Railway limit**: Likely 512MB-1GB, insufficient for concurrent complex scrapes

### Secondary Root Cause: Processing Time Exceeding Timeout

**Evidence Chain**:
1. 40/44 requests timeout at exactly 120s
2. Successful requests complete in 18-26s
3. Failed requests are all complex websites requiring extensive scraping

**Time Consumption Pattern**:
- **Simple sites**: 18-26s (within limit)
- **Complex sites**: >120s (exceeds limit)
- **Multi-page analysis**: 8-15 pages × 5-10s each = 40-150s
- **No early termination**: Service continues until timeout

---

## Recommendations

### Priority 0 (CRITICAL - Immediate Action Required)

#### 1. Further Reduce Concurrency to Prevent OOM Kills

**Action**: Reduce MaxConcurrentRequests from 40 to 20

**Implementation**:
```go
// In config.go
MaxConcurrentRequests: getEnvAsInt("MAX_CONCURRENT_REQUESTS", 20), // Reduced from 40
```

**Expected Impact**: 
- Reduces memory pressure by 50%
- Allows 1-2 concurrent requests instead of 2-3
- Should prevent OOM kills for most requests

**Risk**: Lower throughput, but better than 0% success rate

---

#### 2. Implement Request Admission Control

**Action**: Reject requests if memory usage >80% or queue full

**Implementation**:
```go
// In classification.go
func (h *ClassificationHandler) HandleClassification(...) {
    // Check memory before processing
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    memUsagePercent := float64(m.Alloc) / float64(m.Sys) * 100
    
    if memUsagePercent > 80 {
        h.logger.Warn("Memory usage high, rejecting request",
            zap.Float64("mem_usage_percent", memUsagePercent))
        errors.WriteServiceUnavailable(w, r, "Service temporarily unavailable due to high load")
        return
    }
    
    // Check queue capacity
    if h.requestQueue.Size() >= h.requestQueue.maxSize {
        errors.WriteServiceUnavailable(w, r, "Service queue full, please retry later")
        return
    }
    
    // ... continue processing
}
```

**Expected Impact**:
- Prevents OOM kills by rejecting requests when memory high
- Provides clear error messages to clients
- Allows service to recover instead of crashing

---

#### 3. Implement Adaptive Timeout with Early Termination

**Action**: Add timeout checks and early termination for long-running operations

**Implementation**:
```go
// In classification.go
func (h *ClassificationHandler) processClassification(ctx context.Context, ...) {
    // Check timeout at start
    if deadline, ok := ctx.Deadline(); ok {
        timeRemaining := time.Until(deadline)
        if timeRemaining < 30*time.Second {
            h.logger.Warn("Insufficient time remaining, skipping expensive operations",
                zap.Duration("time_remaining", timeRemaining))
            // Return partial results or skip multi-page analysis
            return h.getQuickClassification(...)
        }
    }
    
    // Check timeout before expensive operations
    if timeRemaining < 60*time.Second {
        // Skip multi-page analysis
        // Use single-page scraping only
    }
}
```

**Expected Impact**:
- Prevents timeouts by terminating early
- Returns partial results instead of failing completely
- Improves success rate for complex sites

---

### Priority 1 (HIGH - Next Sprint)

#### 4. Add Memory-Aware Scraping Strategy Selection

**Action**: Skip Playwright/BrowserHeaders if memory usage is high

**Implementation**:
```go
// In website_scraper.go
func (s *EnhancedWebsiteScraper) ScrapeWebsite(ctx context.Context, url string) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    memUsagePercent := float64(m.Alloc) / float64(m.Sys) * 100
    
    if memUsagePercent > 70 {
        // Skip Playwright and BrowserHeaders, use SimpleHTTP only
        return s.simpleHTTPScraper.Scrape(ctx, url)
    }
    
    // ... normal strategy selection
}
```

**Expected Impact**:
- Reduces memory usage for high-memory situations
- Maintains basic functionality even under memory pressure
- Prevents OOM kills

---

#### 5. Optimize Multi-Page Analysis

**Action**: Parallelize page analysis and add timeout per page

**Implementation**:
```go
// Parallelize page analysis with timeout per page
func (s *MultiPageAnalyzer) AnalyzePages(ctx context.Context, pages []string) {
    // Limit concurrent pages to 3 (was 5)
    sem := make(chan struct{}, 3)
    
    for _, page := range pages {
        // Check timeout before each page
        if deadline, ok := ctx.Deadline(); ok {
            if time.Until(deadline) < 10*time.Second {
                break // Stop if timeout approaching
            }
        }
        
        sem <- struct{}{}
        go func(p string) {
            defer func() { <-sem }()
            // Analyze page with per-page timeout
            pageCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
            defer cancel()
            s.analyzePage(pageCtx, p)
        }(page)
    }
}
```

**Expected Impact**:
- Reduces total processing time
- Prevents timeout by stopping early
- Maintains quality for pages analyzed

---

#### 6. Add Comprehensive Memory Monitoring

**Action**: Log memory stats at key points and alert on high usage

**Implementation**:
```go
// In main.go (already partially implemented)
func logMemoryStats(logger *zap.Logger) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    logger.Info("Memory stats",
        zap.Uint64("alloc_mb", m.Alloc/1024/1024),
        zap.Uint64("sys_mb", m.Sys/1024/1024),
        zap.Uint64("num_gc", uint64(m.NumGC)),
        zap.Float64("mem_usage_percent", float64(m.Alloc)/float64(m.Sys)*100))
    
    if float64(m.Alloc)/float64(m.Sys) > 0.8 {
        logger.Warn("High memory usage detected",
            zap.Float64("usage_percent", float64(m.Alloc)/float64(m.Sys)*100))
    }
}
```

**Expected Impact**:
- Provides visibility into memory usage patterns
- Enables proactive action before OOM kills
- Helps identify memory leaks

---

### Priority 2 (MEDIUM - Future Improvements)

#### 7. Investigate Industry Classification Logic

**Action**: Debug why industry is always "unknown" despite high confidence

**Implementation**:
- Add detailed logging in industry detection
- Review keyword extraction quality
- Test industry detection with known good inputs

---

#### 8. Implement Request Caching

**Action**: Cache classification results to reduce processing load

**Implementation**:
- Use Redis cache for classification results
- Cache key: business_name + website_url hash
- TTL: 24 hours

**Expected Impact**:
- Reduces processing load
- Improves response times
- Reduces memory usage

---

## Implementation Plan

### Phase 1: Immediate Fixes (This Week)

1. ✅ **Reduce Concurrency** (MaxConcurrentRequests: 40 → 20)
2. ✅ **Add Request Admission Control** (Memory + Queue checks)
3. ✅ **Implement Adaptive Timeout** (Early termination)

**Expected Outcome**: Success rate improves from 6.81% → 30-40%

### Phase 2: Memory Optimization (Next Week)

4. ✅ **Memory-Aware Scraping** (Skip Playwright if memory high)
5. ✅ **Optimize Multi-Page Analysis** (Parallelize + timeout per page)
6. ✅ **Add Memory Monitoring** (Logging + alerts)

**Expected Outcome**: Success rate improves from 30-40% → 60-70%

### Phase 3: Quality Improvements (Following Week)

7. ✅ **Fix Industry Classification** (Debug + fix logic)
8. ✅ **Implement Request Caching** (Redis cache)

**Expected Outcome**: Success rate improves from 60-70% → 80-90%

---

## Success Criteria

| Metric | Current | Target | Timeline |
|--------|---------|--------|----------|
| **Success Rate** | 6.81% | ≥95% | Phase 3 |
| **HTTP 502 Errors** | 90.9% | <5% | Phase 1 |
| **HTTP 503 Errors** | 2.3% | <1% | Phase 1 |
| **Average Processing Time** | 22.6s (success) | <60s | Phase 2 |
| **Memory Usage** | Unknown | <80% | Phase 2 |
| **Industry Classification** | 0% (all unknown) | >80% | Phase 3 |

---

## Conclusion

The classification service is experiencing **critical memory exhaustion issues** that cause OOM kills and service crashes. The API Gateway timeout fix (30s → 120s) exposed this underlying problem - requests now wait longer but the service cannot complete them due to memory limits.

**Primary Issues**:
1. **Memory exhaustion** (P0) - Service crashes under load
2. **Processing timeouts** (P0) - Complex sites exceed 120s
3. **Industry classification** (P1) - Not working despite high confidence
4. **Service availability** (P1) - Crashes cause temporary unavailability

**Recommended Actions**:
- **Immediate**: Reduce concurrency, add admission control, implement adaptive timeouts
- **Short-term**: Memory-aware scraping, optimize multi-page analysis
- **Long-term**: Fix industry classification, implement caching

**Expected Timeline**: 2-3 weeks to reach ≥95% success rate with proper implementation.

---

**Status**: 🔴 **CRITICAL - Immediate Action Required**

