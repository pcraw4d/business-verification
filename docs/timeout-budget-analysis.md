# Timeout Budget Analysis

**Date**: December 21, 2025  
**Investigation Track**: Track 1.1 - Timeout Configuration Audit  
**Status**: Completed

## Executive Summary

Analysis of timeout budget allocation revealed a critical issue: the total timeout budget (86s) exceeded the OverallTimeout (60s) by 26 seconds, causing premature request timeouts. This has been fixed by:
1. Increasing OverallTimeout from 60s to 90s
2. Optimizing budget allocation from 86s to 71s

---

## Problem Identified

### Original Budget Calculation

**Location**: `services/classification-service/internal/handlers/classification.go:5240-5283`

**Original Budget Breakdown** (for requests with website scraping):
```
Index Building:        30s
Phase 1 Scraping:      18s
Multi-page Analysis:   8s
Go Classification:     5s
ML Classification:     10s
General Overhead:      5s
Retry Buffer:         10s
─────────────────────────
TOTAL:                86s
```

**OverallTimeout**: 60s

**Issue**: Budget (86s) exceeds OverallTimeout (60s) by **26 seconds (43% over)**

**Impact**: 
- All requests with website scraping would timeout before completing
- 95th percentile latency hitting exactly 60s (timeout limit)
- 67.1% error rate likely caused by premature timeouts

---

## Solution Implemented

### Fix 1: Increase OverallTimeout

**Change**: Increased `OverallTimeout` from 60s to 90s

**Location**: `services/classification-service/internal/config/config.go:118`

```go
// Before
OverallTimeout: getEnvAsDuration("CLASSIFICATION_OVERALL_TIMEOUT", 60*time.Second)

// After
OverallTimeout: getEnvAsDuration("CLASSIFICATION_OVERALL_TIMEOUT", 90*time.Second) // Increased from 60s to 90s to accommodate timeout budget (86s)
```

**Rationale**: Provides sufficient time for all operations while maintaining reasonable timeout limits

### Fix 2: Optimize Budget Allocation

**Change**: Optimized budget allocation from 86s to 71s

**Location**: `services/classification-service/internal/handlers/classification.go:5250-5260`

**Optimized Budget Breakdown**:
```
Index Building:        20s (reduced from 30s - assumes cache hit after first call)
Phase 1 Scraping:      18s (unchanged)
Multi-page Analysis:   8s (unchanged - optional, may be skipped)
Go Classification:     5s (unchanged)
ML Classification:     10s (unchanged - optional, may be skipped)
General Overhead:      5s (unchanged)
Retry Buffer:          5s (reduced from 10s - retries are typically fast)
─────────────────────────
TOTAL:                71s
```

**Changes Made**:
1. **Index Building**: Reduced from 30s to 20s
   - First call takes 10-30s, but is cached for 5 minutes
   - After first call, cache hits are <1ms
   - **Savings**: 10s

2. **Retry Buffer**: Reduced from 10s to 5s
   - Retries are typically fast (<2s each)
   - Exponential backoff caps at 5s
   - **Savings**: 5s

**Total Savings**: 15s (from 86s to 71s)

### Final Budget Status

| Metric | Before | After | Status |
|--------|--------|-------|--------|
| **Total Budget** | 86s | 71s | ✅ Optimized |
| **OverallTimeout** | 60s | 90s | ✅ Increased |
| **Buffer** | -26s (exceeded) | +19s (buffer) | ✅ Fixed |
| **Margin** | -43% | +27% | ✅ Healthy |

---

## Budget Breakdown Details

### For Requests WITH Website Scraping

| Operation | Budget | Actual Time | Notes |
|-----------|--------|-------------|-------|
| Index Building | 20s | 10-30s (first call) | Cached for 5min after first call |
| Phase 1 Scraping | 18s | 15s + 3s buffer | Aligned with WebsiteScrapingTimeout |
| Multi-page Analysis | 8s | Variable | Optional, may be skipped if time < 10s |
| Go Classification | 5s | <5s | Usually fast |
| ML Classification | 10s | Variable | Optional, may be skipped if time < 15s |
| Risk Assessment | 5s | Parallel | Doesn't add to total (runs in parallel) |
| General Overhead | 5s | Variable | Retries and network latency |
| Retry Buffer | 5s | Variable | For retry attempts |
| **TOTAL** | **71s** | **Variable** | **Within 90s limit** |

### For Requests WITHOUT Website Scraping

| Operation | Budget | Actual Time | Notes |
|-----------|--------|-------------|-------|
| Index Building | 20s | 10-30s (first call) | Cached for 5min after first call |
| Go Classification | 5s | <5s | Usually fast |
| ML Classification | 10s | Variable | Optional, may be skipped if time < 15s |
| General Overhead | 5s | Variable | Retries and network latency |
| Retry Buffer | 5s | Variable | For retry attempts |
| **TOTAL** | **45s** | **Variable** | **Well within 90s limit** |

---

## Optimization Opportunities (Future)

### 1. Pre-warm Index Building Cache

**Current**: Index building takes 10-30s on first call  
**Opportunity**: Pre-warm cache on service startup  
**Potential Savings**: 20-30s on first request only  
**Effort**: Medium  
**Priority**: Low (only affects first request)

### 2. Make Multi-page Analysis Truly Optional

**Current**: Budget allocated but may be skipped  
**Opportunity**: Skip if time remaining < 10s  
**Potential Savings**: 8s when skipped  
**Effort**: Low  
**Priority**: Medium

### 3. Make ML Classification Optional

**Current**: Budget allocated but may be skipped  
**Opportunity**: Skip if time remaining < 15s  
**Potential Savings**: 10s when skipped  
**Effort**: Low  
**Priority**: Medium

### 4. Dynamic Budget Allocation

**Current**: Fixed budget allocation  
**Opportunity**: Adjust budgets based on operation success rates  
**Potential Savings**: Variable  
**Effort**: High  
**Priority**: Low

---

## Expected Impact

### Before Fix

- **Timeout Rate**: High (all requests with scraping timing out)
- **95th Percentile Latency**: 60s (hitting timeout limit)
- **Error Rate**: 67.1% (likely due to timeouts)
- **Success Rate**: 32.9%

### After Fix

- **Timeout Rate**: Low (71s budget within 90s limit)
- **95th Percentile Latency**: Expected <90s (within timeout limit)
- **Error Rate**: Expected significant reduction
- **Success Rate**: Expected significant increase

---

## Validation Plan

1. **Deploy Fix**: Deploy updated code to Railway
2. **Run 50-Sample Test**: Validate timeout improvements
3. **Monitor Metrics**:
   - Timeout rate
   - 95th percentile latency
   - Error rate
   - Success rate
4. **Compare Results**: Before vs After metrics

---

## Code Changes Summary

### Files Modified

1. `services/classification-service/internal/config/config.go`
   - Increased OverallTimeout default from 60s to 90s
   - Increased ReadTimeout default from 90s to 120s
   - Increased WriteTimeout default from 90s to 120s

2. `services/classification-service/internal/handlers/classification.go`
   - Optimized timeout budget allocation
   - Reduced index building budget from 30s to 20s
   - Reduced retry buffer from 10s to 5s
   - Updated budget calculation comments

### Testing Required

- [ ] Unit tests for timeout budget calculation
- [ ] Integration tests for timeout behavior
- [ ] E2E tests to validate timeout improvements

---

**Document Status**: Analysis Complete, Fixes Implemented  
**Next Steps**: Deploy and validate with 50-sample E2E test

