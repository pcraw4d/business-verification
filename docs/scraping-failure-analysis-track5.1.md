# Scraping Failure Analysis - Track 5.1

**Date**: December 21, 2025  
**Investigation Track**: Track 5.1 - Fix Scraping Success Rate  
**Status**: In Progress

## Executive Summary

The 50-sample validation test shows **0% scraping success rate**, down from 10.4% in baseline. Analysis reveals that:
- 19 requests (38%) have empty scraping_strategy (all strategies failed)
- 29 requests (58%) show "early_exit" (not actual scraping - classification optimization)
- 2 requests (4%) may have succeeded but need verification

---

## Problem Analysis

### Test Results Breakdown

| Category | Count | Percentage |
|----------|-------|------------|
| **Empty Strategy (All Failed)** | 19 | 38% |
| **Early Exit (No Scraping)** | 29 | 58% |
| **Unknown** | 2 | 4% |
| **Total** | 50 | 100% |

### Key Findings

1. **"Early Exit" is NOT a Scraping Strategy**
   - "Early exit" is a classification optimization that skips ML classification
   - It's set when classification confidence is high enough without needing ML
   - These requests don't actually attempt scraping
   - The test results incorrectly count these as "scraping attempts"

2. **All Scraping Strategies Failing for 38% of Requests**
   - 19 requests show empty `scraping_strategy` field
   - This indicates all strategies (hrequests, SimpleHTTP, BrowserHeaders, Playwright) failed
   - Processing times range from 26ms to 60s (timeouts)

3. **Content Validation May Be Too Strict**
   - Current requirements:
     - Minimum 50 words
     - Must have title OR meta description
     - Quality score >= 0.5
   - These requirements may be rejecting valid content

---

## Root Cause Analysis

### Issue 1: Content Validation Too Strict

**Location**: `internal/external/website_scraper.go:2324-2351`

**Current Requirements**:
```go
func isContentValid(content *ScrapedContent) bool {
    if content == nil {
        return false
    }
    // Minimum word count
    if content.WordCount < 50 {
        return false
    }
    // Must have basic metadata
    if content.Title == "" && content.MetaDesc == "" {
        return false
    }
    // Check for error pages
    if containsErrorIndicators(content.PlainText) {
        return false
    }
    // Quality score threshold
    if content.QualityScore < 0.5 {
        return false
    }
    return true
}
```

**Problem**: These requirements may be too strict for some websites:
- 50 words minimum may reject valid single-page sites
- Requiring title OR meta description may reject sites with only one
- Quality score 0.5 may reject valid but simple sites

**Fix**: Lower validation thresholds or make them configurable

---

### Issue 2: Scraping Strategies Not Being Attempted

**Location**: `internal/external/website_scraper.go:1914-2052`

**Problem**: Strategies may be failing silently or being skipped:
- Strategies are tried in order: hrequests → SimpleHTTP → BrowserHeaders → Playwright
- If all fail, `scraping_strategy` is empty
- No fallback mechanism if all strategies fail

**Possible Causes**:
1. **URL Validation Too Strict**: The hostname validation I added may be blocking valid URLs
2. **DNS Resolution Failures**: DNS lookups may be failing even with fallback servers
3. **Network Timeouts**: Requests may be timing out before strategies complete
4. **Content Validation Rejecting Valid Content**: Strategies succeed but content is rejected

---

### Issue 3: Early Exit Confusion

**Location**: `services/classification-service/internal/handlers/classification.go:3517-3548`

**Problem**: "Early exit" is being reported as a scraping strategy, but it's actually a classification optimization:
- Early exit happens when ML classification is skipped due to high confidence
- It's not related to scraping at all
- The test results incorrectly count these as "scraping attempts"

**Fix**: Separate scraping strategy from classification early exit in reporting

---

## Failed Scrapes Analysis

### Examples of Failed Scrapes

1. **Stripe** (`https://stripe.com`)
   - Processing time: 44,041ms
   - Strategy: Empty (all failed)
   - Likely cause: Timeout or content validation rejection

2. **Meta** (`https://www.meta.com`)
   - Processing time: 56,354ms
   - Strategy: Empty (all failed)
   - Likely cause: Timeout

3. **Oracle** (`https://www.oracle.com`)
   - Processing time: 53,349ms
   - Strategy: Empty (all failed)
   - Likely cause: Timeout

4. **eBay** (`https://www.ebay.com`)
   - Processing time: 60,001ms (timeout)
   - Strategy: Empty (all failed)
   - Likely cause: Timeout

5. **Tesla** (`https://www.tesla.com`)
   - Processing time: 60,001ms (timeout)
   - Strategy: Empty (all failed)
   - Likely cause: Timeout

**Pattern**: Most failures are timeouts (60s), suggesting:
- Strategies are being attempted but timing out
- Timeout budget may be insufficient for some sites
- Network latency may be high

---

## Investigation Steps

### Step 1: Verify URL Validation

**Action**: Test URL validation with actual test URLs

**Expected**: URL validation should not block valid URLs like:
- `https://stripe.com`
- `https://www.meta.com`
- `https://www.oracle.com`

**Status**: ✅ URL validation regex is correct (tested with Python)

---

### Step 2: Check Content Validation Thresholds

**Action**: Review content validation requirements

**Current Thresholds**:
- Word count: 50 minimum
- Metadata: Title OR meta description required
- Quality score: 0.5 minimum

**Recommendation**: Lower thresholds or make configurable:
- Word count: 30 minimum (was 50)
- Metadata: Optional (remove requirement)
- Quality score: 0.3 minimum (was 0.5)

---

### Step 3: Investigate Strategy Failures

**Action**: Add detailed logging for strategy failures

**Questions to Answer**:
1. Which strategy is being attempted first?
2. What error is each strategy returning?
3. Is content being returned but rejected by validation?
4. Are strategies timing out or failing immediately?

---

### Step 4: Check Timeout Budget

**Action**: Review timeout budget allocation for scraping

**Current**: 
- WebsiteScrapingTimeout: 15s
- Per-strategy timeout: Calculated from context deadline
- Total timeout: 60s (OverallTimeout)

**Issue**: If multiple strategies are tried sequentially, total time may exceed 60s

**Fix**: Optimize strategy selection or reduce per-strategy timeout

---

## Recommended Fixes

### Fix 1: Lower Content Validation Thresholds

**File**: `internal/external/website_scraper.go`

**Changes**:
```go
// Lower word count requirement from 50 to 30
if content.WordCount < 30 {  // was 50
    return false
}

// Make metadata optional (remove requirement)
// Remove: if content.Title == "" && content.MetaDesc == "" { return false }

// Lower quality score threshold from 0.5 to 0.3
if content.QualityScore < 0.3 {  // was 0.5
    return false
}
```

**Expected Impact**: More content will pass validation, increasing scraping success rate

---

### Fix 2: Add Strategy Failure Logging

**File**: `internal/external/website_scraper.go`

**Changes**: Add detailed logging for each strategy failure:
- Strategy name
- Error message
- Content quality metrics (if content returned)
- Validation failure reason (if content rejected)

**Expected Impact**: Better visibility into why strategies are failing

---

### Fix 3: Optimize Strategy Selection

**File**: `internal/external/website_scraper.go`

**Changes**: 
- Skip strategies that are unlikely to succeed (e.g., Playwright for simple sites)
- Add strategy success rate tracking
- Prefer strategies with higher success rates

**Expected Impact**: Faster scraping with higher success rate

---

### Fix 4: Separate Scraping Strategy from Classification Early Exit

**File**: `services/classification-service/internal/handlers/classification.go`

**Changes**: 
- Don't set `scraping_strategy` to "early_exit"
- Use separate field for classification early exit
- Only set `scraping_strategy` when actual scraping occurs

**Expected Impact**: Accurate scraping success rate reporting

---

## Next Steps

1. [ ] Implement Fix 1: Lower content validation thresholds
2. [ ] Implement Fix 2: Add strategy failure logging
3. [ ] Implement Fix 3: Optimize strategy selection
4. [ ] Implement Fix 4: Separate scraping strategy from early exit
5. [ ] Test fixes with 50-sample validation test
6. [ ] Analyze results and iterate

---

**Document Status**: Analysis Complete  
**Next Action**: Implement Fix 1 (Lower Content Validation Thresholds)

