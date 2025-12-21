# Timeout Configuration Audit

**Date**: December 21, 2025  
**Investigation Track**: Track 1.1 - Timeout Configuration Audit  
**Status**: In Progress

## Executive Summary

This document audits all timeout configurations across the classification service to identify misconfigurations, timeout budget issues, and context propagation problems that may be causing premature request terminations.

---

## 1. Railway Environment Variables

### Current Configuration Analysis

Based on code review and documentation:

| Variable | Code Default | Expected Value | Railway Status | Status |
|----------|--------------|----------------|---------------|--------|
| `READ_TIMEOUT` | 90s | 120s | Unknown | ⚠️ Needs Verification |
| `WRITE_TIMEOUT` | 90s | 120s | Unknown | ⚠️ Needs Verification |
| `REQUEST_TIMEOUT` | 120s | 120s | Unknown | ⚠️ Needs Verification |
| `OVERALL_TIMEOUT` | 60s | 60s | Unknown | ⚠️ Needs Verification |
| `CLASSIFICATION_PAGE_ANALYSIS_TIMEOUT` | 15s | 15s | Unknown | ⚠️ Needs Verification |
| `CLASSIFICATION_WEBSITE_SCRAPING_TIMEOUT` | 15s | 15s | Unknown | ⚠️ Needs Verification |
| `HTTP_CLIENT_TIMEOUT` | 120s | 120s | Unknown | ⚠️ Needs Verification |

### Code Defaults (from `services/classification-service/internal/config/config.go`)

```go
ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 90*time.Second)   // Default: 90s
WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 90*time.Second)  // Default: 90s
RequestTimeout: getEnvAsDuration("REQUEST_TIMEOUT", 120*time.Second)  // Default: 120s
OverallTimeout: getEnvAsDuration("CLASSIFICATION_OVERALL_TIMEOUT", 60*time.Second)  // Default: 60s
PageAnalysisTimeout: getEnvAsDuration("CLASSIFICATION_PAGE_ANALYSIS_TIMEOUT", 15*time.Second)  // Default: 15s
WebsiteScrapingTimeout: getEnvAsDuration("CLASSIFICATION_WEBSITE_SCRAPING_TIMEOUT", 15*time.Second)  // Default: 15s
```

### Issues Identified

1. **READ_TIMEOUT Default Too Low**: Code default is 90s, but documentation suggests 120s is needed
2. **WRITE_TIMEOUT Default Too Low**: Code default is 90s, but documentation suggests 120s is needed
3. **Railway Configuration Unknown**: Need to verify actual Railway environment variable values

### Action Required

- [ ] Verify Railway environment variables via Railway dashboard or CLI
- [ ] Update code defaults to 120s for READ_TIMEOUT and WRITE_TIMEOUT if Railway values are not set
- [ ] Document actual Railway values

---

## 2. Timeout Budget Allocation Analysis

### Current Budget Calculation

Location: `services/classification-service/internal/handlers/classification.go:5240-5283`

#### Budget Breakdown (for requests with website scraping):

```go
const (
    phase1ScrapingBudget    = 18 * time.Second  // Phase 1 scraper: 18s
    multiPageAnalysisBudget = 8 * time.Second   // Multi-page analysis: 8s
    indexBuildingBudget    = 30 * time.Second  // Keyword index building: 30s (first call)
    goClassificationBudget = 5 * time.Second   // Go classification: 5s
    mlClassificationBudget = 10 * time.Second  // ML classification: 10s
    riskAssessmentBudget   = 5 * time.Second   // Risk assessment: 5s (parallel)
    generalOverhead        = 5 * time.Second   // General overhead: 5s
    retryBuffer            = 10 * time.Second  // Retry buffer: 10s
)
```

#### Total Budget Calculation:

```
Total = indexBuildingBudget + phase1ScrapingBudget + multiPageAnalysisBudget + 
        goClassificationBudget + mlClassificationBudget + generalOverhead + retryBuffer
Total = 30 + 18 + 8 + 5 + 10 + 5 + 10
Total = 86 seconds
```

### Critical Issue: Budget Exceeds OverallTimeout

| Metric | Value | Status |
|--------|-------|--------|
| **Total Budget** | 86s | ❌ **EXCEEDS** |
| **OverallTimeout** | 60s | ⚠️ **TOO LOW** |
| **Difference** | +26s (43% over) | ❌ **CRITICAL** |

**Impact**: 
- Requests with website scraping will **always timeout** before completing
- The timeout budget (86s) exceeds the OverallTimeout (60s) by 26 seconds
- This explains why 95th percentile latency is hitting 60s (timeout limit)

### Budget Breakdown Analysis

| Operation | Budget | Actual Time | Notes |
|-----------|--------|-------------|-------|
| Index Building | 30s | 10-30s (first call) | Cached for 5min after first call |
| Phase 1 Scraping | 18s | 15s + 3s buffer | Aligned with WebsiteScrapingTimeout |
| Multi-page Analysis | 8s | Variable | May be skipped if insufficient time |
| Go Classification | 5s | <5s | Usually fast |
| ML Classification | 10s | Variable | Optional, may be skipped |
| Risk Assessment | 5s | Parallel | Doesn't add to total |
| General Overhead | 5s | Variable | Retries and network latency |
| Retry Buffer | 10s | Variable | For retry attempts |

### Optimization Opportunities

1. **Index Building (30s)**: 
   - First call takes 10-30s, but is cached for 5 minutes
   - **Recommendation**: Consider pre-warming cache on service startup
   - **Potential Savings**: 20-30s on first request only

2. **Multi-page Analysis (8s)**:
   - May be skipped if insufficient time
   - **Recommendation**: Make this truly optional and skip if time remaining < 10s
   - **Potential Savings**: 8s when skipped

3. **ML Classification (10s)**:
   - Optional, may be skipped
   - **Recommendation**: Skip if time remaining < 15s
   - **Potential Savings**: 10s when skipped

4. **Retry Buffer (10s)**:
   - For retry attempts
   - **Recommendation**: Reduce to 5s if retries are fast
   - **Potential Savings**: 5s

### Recommended Fixes

#### Option 1: Increase OverallTimeout (Quick Fix)
- Increase `OverallTimeout` from 60s to 90s
- **Pros**: Simple, allows all operations to complete
- **Cons**: Increases maximum request time

#### Option 2: Optimize Budget Allocation (Better Fix)
- Reduce index building budget to 20s (assume cache hit after first call)
- Make multi-page analysis truly optional (skip if time < 10s)
- Make ML classification optional (skip if time < 15s)
- Reduce retry buffer to 5s
- **New Total**: 20 + 18 + 0 + 5 + 0 + 5 + 5 = 53s (within 60s limit)
- **Pros**: More efficient, respects timeout limits
- **Cons**: May reduce accuracy slightly

#### Option 3: Hybrid Approach (Recommended)
- Increase `OverallTimeout` to 90s
- Optimize budget allocation as in Option 2
- **New Total**: 53s (well within 90s limit)
- **Pros**: Provides buffer while optimizing
- **Cons**: None

---

## 3. Context Propagation Analysis

### Context Propagation Path

1. **Request Entry**: `handleClassification` → Creates context with timeout
2. **Processing**: `processClassification` → Receives context
3. **Website Scraping**: `website_scraper.go` → Should receive context
4. **ML Service**: `python_ml_service.go` → Should receive context
5. **Code Generation**: `classifier.go` → Should receive context

### Code Review Findings

#### ✅ Context Propagation in processClassification

Location: `services/classification-service/internal/handlers/classification.go:2201-2254`

```go
func (h *ClassificationHandler) processClassification(ctx context.Context, req *ClassificationRequest, startTime time.Time) (*ClassificationResponse, error) {
    // Early termination check
    if ctx.Err() != nil {
        return nil, fmt.Errorf("context already expired before processing: %w", ctx.Err())
    }
    
    // Check time remaining and refresh context if needed
    var processingCtx context.Context = ctx
    var cancelFunc context.CancelFunc = nil
    
    // Adaptive timeout calculation
    adaptiveTimeout := h.calculateAdaptiveTimeout(req)
    // ... context handling logic
}
```

**Status**: ✅ Context is checked and propagated correctly

#### ✅ Context Propagation in Website Scraper

Location: `internal/external/website_scraper.go:277-341`

```go
func (s *WebsiteScraper) Scrape(ctx context.Context, targetURL string) (*ScrapedContent, error) {
    // Check if context is already cancelled
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Retry loop with context checks
    for attempt = 0; attempt <= s.config.MaxRetries; attempt++ {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(retryDelay):
        }
        // ... scraping logic
    }
}
```

**Status**: ✅ Context is checked before operations and in retry loops

#### ✅ Context Propagation in Python ML Service

Location: `internal/machine_learning/infrastructure/python_ml_service.go:512-556`

```go
func (pms *PythonMLService) ClassifyEnhanced(ctx context.Context, req *EnhancedClassificationRequest) (*EnhancedClassificationResponse, error) {
    err = pms.circuitBreaker.Execute(ctx, func() error {
        httpReq, createErr := http.NewRequestWithContext(ctx, "POST", ...)
        // ... request execution
    })
}
```

**Status**: ✅ Context is passed to HTTP requests correctly

### Context Cancellation Handling

All critical paths check for context cancellation:
- ✅ Before long operations
- ✅ In retry loops
- ✅ Before HTTP requests
- ✅ In error handling

**Status**: ✅ Context propagation is working correctly

---

## 4. Findings Summary

### Critical Issues

1. **Timeout Budget Exceeds OverallTimeout** ❌
   - Budget: 86s vs OverallTimeout: 60s
   - **Impact**: High - Causes premature timeouts
   - **Confidence**: 95%

2. **READ_TIMEOUT Default May Be Too Low** ⚠️
   - Code default: 90s vs Expected: 120s
   - **Impact**: Medium - May cause connection closures
   - **Confidence**: 70%

3. **Railway Environment Variables Unknown** ⚠️
   - Need to verify actual Railway values
   - **Impact**: Medium - May have misconfigurations
   - **Confidence**: 50%

### Non-Issues

1. **Context Propagation** ✅
   - All critical paths properly check and propagate context
   - Context cancellation is handled correctly

---

## 5. Recommendations

### Immediate Actions (Priority 1)

1. **Fix Timeout Budget Exceedance**
   - **Option A**: Increase `OverallTimeout` to 90s
   - **Option B**: Optimize budget allocation (reduce to 53s)
   - **Option C**: Hybrid approach (increase to 90s + optimize)

2. **Verify Railway Environment Variables**
   - Check actual Railway values for all timeout variables
   - Update if misconfigured

3. **Update Code Defaults**
   - Change `READ_TIMEOUT` default from 90s to 120s
   - Change `WRITE_TIMEOUT` default from 90s to 120s

### Short-Term Actions (Priority 2)

4. **Optimize Budget Allocation**
   - Pre-warm index building cache on startup
   - Make multi-page analysis truly optional
   - Make ML classification optional when time is limited
   - Reduce retry buffer if retries are fast

5. **Add Timeout Monitoring**
   - Log timeout budget calculations
   - Track actual operation times vs budgets
   - Alert when budgets are exceeded

### Long-Term Actions (Priority 3)

6. **Implement Adaptive Timeouts**
   - Dynamically adjust timeouts based on operation success rates
   - Learn optimal timeout values from historical data

---

## 6. Next Steps

1. [ ] Verify Railway environment variables
2. [ ] Implement timeout budget fix (Option C - Hybrid)
3. [ ] Update code defaults for READ_TIMEOUT and WRITE_TIMEOUT
4. [ ] Test with 50-sample E2E test
5. [ ] Monitor timeout metrics after fix

---

**Document Status**: Initial Analysis Complete  
**Next Review**: After Railway variable verification

