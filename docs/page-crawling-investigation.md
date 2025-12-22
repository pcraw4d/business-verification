# Page Crawling Investigation - Track 5.2

## Executive Summary

Investigation of page crawling reveals **multi-page analysis is being skipped due to time constraints** and **crawling may be failing silently**. The crawling infrastructure exists but is not being utilized effectively.

**Status**: ⚠️ **HIGH** - Crawling disabled by time constraints, needs optimization

## Page Crawling Configuration

### Configuration Settings

**Location**: `services/classification-service/internal/config/config.go:116-129`

**Settings**:
- `MaxPagesToAnalyze`: 15 (default)
- `PageAnalysisTimeout`: 15s (default)
- `ConcurrentPages`: 5 (default)
- `MultiPageAnalysisEnabled`: `true` (default)
- `MaxConcurrentPages`: 3 (default)
- `CrawlDelayMs`: 500ms (default)
- `FastPathMaxPages`: 8 (default)

**Status**: ✅ Configuration looks correct

### Feature Flag

**Location**: `services/classification-service/internal/config/config.go:123`

**Setting**: `MultiPageAnalysisEnabled: getEnvAsBool("ENABLE_MULTI_PAGE_ANALYSIS", true)`

**Status**: ✅ Enabled by default

## Page Crawling Logic

### Crawling Implementation

**Location**: `internal/classification/smart_website_crawler.go`

**Key Functions**:
1. `CrawlWebsite()` - Regular crawl mode
2. `CrawlWebsiteFast()` - Fast-path mode for short timeouts
3. `discoverSiteStructure()` - Discovers site structure
4. `analyzePagesParallel()` - Analyzes pages in parallel

**Status**: ✅ Crawling logic implemented

### Multi-Page Analysis Skipping

**Location**: `services/classification-service/internal/handlers/classification.go:2310-2342`

**Skip Conditions**:
```go
if timeRemaining < 30*time.Second {
    skipMultiPageAnalysis = true
    skipMLClassification = true
} else if timeRemaining < 60*time.Second {
    skipMultiPageAnalysis = true
}
```

**Issue**: ⚠️ **HIGH** - Multi-page analysis is skipped when time remaining < 60 seconds

**Impact**: 
- If request timeout is 120s and processing takes >60s, multi-page analysis is skipped
- This explains why 0 pages are being crawled on average

**Status**: ⚠️ **IDENTIFIED** - Time constraints causing skipping

### Crawling Flow

**Location**: `internal/classification/repository/supabase_repository.go:5836-5876`

**Flow**:
1. Check timeout duration
2. If timeout <= 5s: Use fast-path mode
3. If timeout > 5s: Use regular crawl mode
4. Discover site structure
5. Analyze pages in parallel
6. Return pages analyzed

**Status**: ✅ Flow looks correct

## Root Cause Analysis

### Primary Issues

1. **Time Constraints Skipping Multi-Page Analysis** ⚠️ **HIGH**
   - Multi-page analysis skipped when time remaining < 60s
   - Most requests likely have < 60s remaining by the time crawling starts
   - **Impact**: 0 pages crawled
   - **Evidence**: Skip logic in `classification.go:2338-2341`

2. **Crawling Timeout Too Short** ⚠️ **MEDIUM**
   - Fast-path mode used when timeout <= 5s
   - Regular crawl may timeout before completing
   - **Impact**: Crawling fails or returns 0 pages
   - **Evidence**: Fast-path threshold is 5s

3. **Site Structure Discovery Failing** ⚠️ **MEDIUM**
   - `discoverSiteStructure()` may be failing silently
   - Falls back to homepage only if discovery fails
   - **Impact**: Only 1 page crawled (homepage)
   - **Evidence**: Fallback logic in `smart_website_crawler.go:390-395`

4. **Page Analysis Timing Out** ⚠️ **MEDIUM**
   - Page analysis may be timing out
   - Concurrent pages may be too high
   - **Impact**: Pages not analyzed, 0 pages returned
   - **Evidence**: Need to check timeout handling

5. **Feature Flag Disabled** ⚠️ **LOW**
   - `ENABLE_MULTI_PAGE_ANALYSIS` may be disabled in Railway
   - **Impact**: Multi-page analysis never runs
   - **Evidence**: Need to verify in Railway

## Investigation Steps

### Step 1: Review Skip Logic

**Check**:
- When is `skipMultiPageAnalysis` set to true?
- What is the typical time remaining when crawling starts?
- Is the 60s threshold too restrictive?

**Status**: ✅ **COMPLETED** - Skip logic identified

### Step 2: Check Crawling Configuration

**Verify**:
- `ENABLE_MULTI_PAGE_ANALYSIS` is `true` in Railway
- Timeout values are appropriate
- Concurrent page limits are reasonable

**Status**: ⏳ **PENDING** - Need to verify in Railway

### Step 3: Test Crawling Manually

**Test**:
- Create test script to crawl known websites
- Verify site structure discovery works
- Verify pages are being crawled
- Verify content is being extracted

**Status**: ⏳ **PENDING** - Need to create test script

### Step 4: Analyze Crawling Failures

**Check**:
- Railway logs for crawling errors
- Timeout errors
- Site structure discovery failures
- Page analysis failures

**Status**: ⏳ **PENDING** - Need to analyze logs

## Recommendations

### Immediate Actions (High Priority)

1. **Adjust Time Threshold for Multi-Page Analysis**:
   - Reduce threshold from 60s to 30s (or remove threshold)
   - Allow multi-page analysis even with limited time
   - **Expected Impact**: More pages crawled

2. **Optimize Crawling Performance**:
   - Reduce page analysis timeout if needed
   - Optimize site structure discovery
   - **Expected Impact**: Faster crawling, more pages

3. **Verify Feature Flag**:
   - Check `ENABLE_MULTI_PAGE_ANALYSIS` in Railway
   - Ensure it's set to `true`
   - **Expected Impact**: Multi-page analysis enabled

### Medium Priority Actions

4. **Improve Site Structure Discovery**:
   - Add retry logic for discovery failures
   - Improve discovery algorithms
   - **Expected Impact**: More pages discovered

5. **Optimize Page Analysis**:
   - Reduce analysis time per page
   - Improve concurrent page handling
   - **Expected Impact**: More pages analyzed

6. **Add Crawling Metrics**:
   - Track pages discovered
   - Track pages analyzed
   - Track crawling failures
   - **Expected Impact**: Better visibility

### Low Priority Actions

7. **Review Crawling Strategy**:
   - Consider different crawling strategies
   - Optimize for different website types
   - **Expected Impact**: Better crawling success

## Code Locations

- **Skip Logic**: `services/classification-service/internal/handlers/classification.go:2310-2342`
- **Crawling Config**: `services/classification-service/internal/config/config.go:116-129`
- **Crawling Implementation**: `internal/classification/smart_website_crawler.go`
- **Multi-Page Analysis**: `internal/classification/repository/supabase_repository.go:5836-5876`

## Next Steps

1. ✅ **Complete Track 5.2 Investigation** - This document
2. **Adjust Time Threshold** - Reduce or remove 60s threshold
3. **Verify Feature Flag** - Check Railway settings
4. **Test Crawling** - Create test script
5. **Optimize Performance** - Improve crawling speed
6. **Add Metrics** - Track crawling success

## Expected Impact

After fixing issues:

1. **Pages Crawled**: 0 → >0 (target: average 3-5 pages)
2. **Scraping Success**: Improved with more pages
3. **Classification Accuracy**: Improved with more data
4. **Code Generation**: Improved with better content

## References

- Skip Logic: `services/classification-service/internal/handlers/classification.go:2310-2342`
- Crawling Config: `services/classification-service/internal/config/config.go:116-129`
- Crawling Implementation: `internal/classification/smart_website_crawler.go`
- Track 5.1: `docs/scraping-failure-analysis.md` (if exists)

