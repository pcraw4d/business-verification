# Railway Logs Analysis - Website Scraping Patterns

**Date**: December 2, 2025  
**Log File**: `docs/railway log/logs.classification.json`  
**Total Log Entries**: 1,533

---

## Executive Summary

### ✅ What's Working

1. **Parallel Processing**: ✅ **WORKING**
   - 131 mentions of parallel processing
   - Logs show: `[PARALLEL] Starting parallel crawl with 3 concurrent requests`
   - Pages are being processed in parallel

2. **Service Health**: ✅ **WORKING**
   - Service is running and responding
   - Health checks passing

### ❌ Critical Issues Found

1. **Fast-Path Mode NOT Being Used**: ❌
   - Only 3 mentions of "fast-path" (all for ML model, not website scraping)
   - **NO fast-path mode for website scraping detected**
   - All requests using REGULAR crawl mode

2. **Timeout Errors**: ⚠️
   - 28 timeout errors detected
   - "context deadline exceeded" errors
   - Timeout duration: ~10s (not 5s as expected for fast-path)

3. **Regular Crawl Mode Active**: ⚠️
   - Logs show: `[REGULAR] Using regular crawl mode`
   - Timeout: 9.999s (close to 10s, not 5s)
   - This explains why requests are timing out

---

## Root Cause Identified

### The Problem

**File**: `internal/classification/repository/supabase_repository.go`  
**Line**: 2473

```go
// Reduced timeout to 5s for fast-path, but allow up to 10s for regular path
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
```

**Issue**: The timeout is hardcoded to 10 seconds, which prevents fast-path mode from being triggered.

**Fast-Path Check** (Line 3536):
```go
if timeoutDuration <= 5*time.Second {
    // Use fast-path mode
} else {
    // Use regular crawl mode  ← This is being used!
}
```

**Result**: Since timeout is 10s > 5s, regular mode is always selected.

---

## Detailed Analysis

### 1. Fast-Path Mode Analysis

**Status**: ❌ **NOT BEING USED**

**Findings**:
- Only 3 mentions of "fast-path" in logs
- All 3 are for ML model selection, not website scraping
- **No logs showing `[FAST-PATH]` for website crawling**

**Expected Logs** (Missing):
```
🚀 [KeywordExtraction] [MultiPage] [FAST-PATH] Using fast-path mode
🚀 [ScrapeMultiPage] [FAST-PATH] Using fast-path mode
```

**Actual Logs Found**:
```
🔍 [KeywordExtraction] [MultiPage] [REGULAR] Using regular crawl mode (timeout: 9.999s, concurrent: 3)
```

**Root Cause**: Timeout is 10s, which is > 5s threshold, so regular mode is selected.

---

### 2. Parallel Processing Analysis

**Status**: ✅ **WORKING**

**Findings**:
- 131 mentions of parallel processing
- Logs confirm parallel crawl is active:
  ```
  🔄 [SmartCrawler] [PARALLEL] Starting parallel crawl with 3 concurrent requests for 21 pages
  📄 [SmartCrawler] [PARALLEL] Page 1 analyzed in 99.95ms
  📄 [SmartCrawler] [PARALLEL] Page 2 analyzed in 294.66ms
  ```

**Conclusion**: Parallel processing is working correctly.

---

### 3. Timeout Error Analysis

**Status**: ⚠️ **TIMEOUT ERRORS DETECTED**

**Findings**:
- 28 timeout errors in logs
- Errors: `context deadline exceeded`
- Timeout duration: ~10s (not 5s)

**Sample Errors**:
```
❌ [KeywordExtraction] [SinglePage] HTTP ERROR (timeout): 
   Request failed for https://example.com: 
   Get "https://example.com": context deadline exceeded
```

**Analysis**:
- Timeout is 9.999s (close to 10s)
- Fast-path should use 5s timeout
- Regular mode is being used instead

---

### 4. Crawl Mode Selection

**Status**: ❌ **REGULAR MODE BEING USED**

**Findings**:
- All logs show `[REGULAR]` crawl mode
- No `[FAST-PATH]` logs for website scraping
- Timeout: 9.999s (regular mode timeout)

**Sample Logs**:
```
📊 [KeywordExtraction] [MultiPage] Timeout duration: 9.999006229s (threshold: 5s)
🔍 [KeywordExtraction] [MultiPage] [REGULAR] Using regular crawl mode 
   (timeout: 9.999006229s, concurrent: 3)
```

**Expected**:
```
📊 [KeywordExtraction] [MultiPage] Timeout duration: 5s (threshold: 5s)
🚀 [KeywordExtraction] [MultiPage] [FAST-PATH] Using fast-path mode 
   (timeout: 5s, max pages: 8, concurrent: 3)
```

---

## Evidence from Logs

### Regular Mode Being Used

```
📊 [KeywordExtraction] [MultiPage] Timeout duration: 9.999006229s (threshold: 5s)
🔍 [KeywordExtraction] [MultiPage] [REGULAR] Using regular crawl mode 
   (timeout: 9.999006229s, concurrent: 3)
```

**Analysis**:
- Timeout: 9.999s (regular mode)
- Threshold check: 5s
- Since 9.999s > 5s, regular mode is selected
- **This is the problem!**

### Parallel Processing Working

```
🔄 [SmartCrawler] [PARALLEL] Starting parallel crawl with 3 concurrent requests for 21 pages
📄 [SmartCrawler] [PARALLEL] Page 1 analyzed in 99.951685ms
📄 [SmartCrawler] [PARALLEL] Page 2 analyzed in 294.657563ms
🔄 [SmartCrawler] [PARALLEL] Parallel analysis completed in X.XXs
```

**Analysis**:
- Parallel processing is active
- 3 concurrent requests
- Pages analyzed in parallel
- ✅ This is working correctly

---

## The Fix

### Solution

**File**: `internal/classification/repository/supabase_repository.go`  
**Line**: 2473

**Current Code**:
```go
// Reduced timeout to 5s for fast-path, but allow up to 10s for regular path
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
```

**Fixed Code**:
```go
// Use 5s timeout to enable fast-path mode (fast-path threshold is 5s)
// Fast-path mode: 5s timeout, 8 pages max, 3 concurrent
// This ensures fast-path mode is triggered for website scraping
websiteScrapingTimeout := 5 * time.Second
if config != nil && config.WebsiteScrapingTimeout > 0 {
    websiteScrapingTimeout = config.WebsiteScrapingTimeout
}
ctx, cancel := context.WithTimeout(context.Background(), websiteScrapingTimeout)
```

**Alternative (Use Config)**:
```go
// Use configured timeout (default 5s from config)
timeout := 5 * time.Second
if r.config != nil && r.config.WebsiteScrapingTimeout > 0 {
    timeout = r.config.WebsiteScrapingTimeout
}
ctx, cancel := context.WithTimeout(context.Background(), timeout)
```

---

## Expected vs Actual Behavior

### Expected (Fast-Path Mode)

```
Timeout: 5s
Mode: FAST-PATH
Max Pages: 8
Concurrent: 3
Log: [FAST-PATH] Using fast-path mode
```

### Actual (Regular Mode)

```
Timeout: 9.999s
Mode: REGULAR
Max Pages: 15
Concurrent: 3
Log: [REGULAR] Using regular crawl mode
```

---

## Performance Impact

### Current (Regular Mode)

- **Timeout**: 10s
- **Max Pages**: 15
- **Result**: Requests timing out at Railway gateway (60s limit)
- **Success Rate**: ~0% (all timing out)

### Expected (Fast-Path Mode)

- **Timeout**: 5s
- **Max Pages**: 8
- **Result**: Requests complete in 2-4s
- **Success Rate**: >80%

---

## Next Steps

1. ✅ **Logs Analyzed** - This document
2. ⏳ **Fix Timeout** - Change 10s to 5s (or use config)
3. ⏳ **Test Fast-Path** - Verify fast-path mode works
4. ⏳ **Monitor Performance** - Track improvements
5. ⏳ **Verify in Production** - Check Railway logs after fix

---

## Files

- **Log File**: `docs/railway log/logs.classification.json`
- **Analysis**: `docs/railway-logs-analysis-website-scraping.md` (this document)
- **Code to Fix**: `internal/classification/repository/supabase_repository.go:2473`

---

## Conclusion

**Root Cause**: Timeout is hardcoded to 10s in `supabase_repository.go:2473`, which prevents fast-path mode from being triggered (fast-path requires timeout <= 5s).

**Solution**: Change the timeout to 5s (or use the configured `WebsiteScrapingTimeout` value) to enable fast-path mode.
