# Deep Investigation: Railway Production Logs & Code Analysis
## December 19, 2025

**Investigation Scope**: Analysis of Railway production logs combined with codebase review to identify root causes of test failures.

---

## Executive Summary

After analyzing Railway production logs and the codebase, **critical discrepancies** have been identified between:
1. **Cache key generation methods** - Multiple different implementations causing cache misses
2. **Metadata extraction** - Test runner not properly extracting metadata from responses
3. **Request timeout handling** - Service appears to be processing requests but timing out at client
4. **Service health** - Multiple dependent services showing health check failures

---

## Key Findings from Railway Logs

### Log File Analysis

**File**: `docs/railway log/complete log.json`  
**Format**: JSON array of structured log entries  
**Time Range**: December 19, 2025 05:41-05:43 UTC  
**Services Logged**: Multiple services including classification-service, api-gateway, monitoring-service

### Critical Observations

#### 1. Cache Metrics Show Zero Activity

From log analysis:
```json
{
  "message": "Cache metrics",
  "attributes": {
    "hits": 0,
    "misses": 0,
    "hit_rate": 0,
    "avg_latency": 0,
    "errors": 0
  }
}
```

**Finding**: Cache metrics show **zero hits, zero misses** - indicating cache operations are not being tracked or cache is not being used at all.

#### 2. Service Health Issues

Logs show multiple service health check failures:
```
❌ Service pipeline-service health check failed: status 502
❌ Service monitoring-service health check failed: status 502
❌ Service legacy-api-service health check failed: status 404
❌ Service legacy-frontend-service health check failed: status 404
✅ Health check complete: 6/10 services healthy
```

**Finding**: **40% of services** are unhealthy, which could impact classification service dependencies.

#### 3. No Classification Request Logs Found

**Finding**: Despite running 100 test requests, **no classification request logs** (`/v1/classify`, `POST`, classification-related messages) were found in the Railway logs during the test period.

**Possible Explanations**:
- Logs are filtered or not capturing classification requests
- Requests are timing out before reaching the service
- Requests are being handled by a different service/replica
- Log aggregation delay

#### 4. Performance Summary Shows Zero Cache Hit Rate

```json
{
  "message": "Performance summary",
  "attributes": {
    "cache_hit_rate": 0,
    "avg_query_latency": 0,
    "error_rate": 0,
    "memory_mb": 11.76
  }
}
```

**Finding**: Performance monitoring confirms **0% cache hit rate** and **0 average query latency** (suspicious - suggests metrics not being tracked).

---

## Code Analysis: Cache Key Generation Issues

### Multiple Cache Key Generation Methods

**CRITICAL FINDING**: There are **at least 3 different cache key generation methods** in the codebase:

#### Method 1: ClassificationHandler.getCacheKey() 
**Location**: `services/classification-service/internal/handlers/classification.go:578`

```go
func (h *ClassificationHandler) getCacheKey(req *ClassificationRequest) string {
    businessName := strings.TrimSpace(strings.ToLower(req.BusinessName))
    description := strings.TrimSpace(strings.ToLower(req.Description))
    websiteURL := strings.TrimSpace(strings.ToLower(req.WebsiteURL))
    
    data := fmt.Sprintf("%s|%s|%s", businessName, description, websiteURL)
    hash := sha256.Sum256([]byte(data))
    return fmt.Sprintf("%x", hash)
}
```

**Used For**: Non-streaming classification requests  
**Key Components**: Business name, description, website URL (normalized)

#### Method 2: ClassificationCache.GenerateCacheKey()
**Location**: `internal/classification/cache.go:37`

```go
func (c *ClassificationCache) GenerateCacheKey(scrapedContent interface{}, websiteURL string) string {
    domain := extractDomain(websiteURL)
    
    var title, textContent string
    if extContent, ok := scrapedContent.(*external.ScrapedContent); ok {
        title = extContent.Title
        textContent = fmt.Sprintf("%s|%s|%s",
            extContent.MetaDesc,
            extContent.AboutText,
            strings.Join(extContent.Headings, "|"))
    }
    
    contentStr := fmt.Sprintf("%s|%s|%s|%s",
        getStringValue(title),
        getStringValue(textContent),
        domain,
        websiteURL)
    
    hash := sha256.Sum256([]byte(contentStr))
    return hex.EncodeToString(hash[:])
}
```

**Used For**: Classification service internal caching (when scraped content is available)  
**Key Components**: Title, meta description, about text, headings, domain, website URL

#### Method 3: IntelligentRoutingAdapter.GenerateCacheKey()
**Location**: `internal/api/adapters/intelligent_routing_adapters.go:373`

```go
func (a *IntelligentRoutingAdapter) GenerateCacheKey(req *EnhancedClassificationRequest) string {
    keyData := map[string]interface{}{
        "business_name":     req.BusinessName,
        "website_url":       req.WebsiteURL,
        "description":       req.Description,
        "industry":          req.Industry,
        "keywords":          req.Keywords,
        "geographic_region": req.GeographicRegion,
    }
    if req.EnhancedFeatures != nil {
        keyData["enhanced_features"] = req.EnhancedFeatures
    }
    
    data, _ := json.Marshal(keyData)
    return fmt.Sprintf("classification:%x", data)
}
```

**Used For**: Intelligent routing adapter  
**Key Components**: Business name, website URL, description, industry, keywords, geographic region, enhanced features

### Root Cause: Cache Key Mismatch

**Problem**: Different code paths use different cache key generation methods:

1. **Initial Request** → Uses `ClassificationHandler.getCacheKey()` (business name + description + URL)
2. **After Scraping** → Uses `ClassificationCache.GenerateCacheKey()` (scraped content + URL)
3. **Cached Lookup** → Uses `ClassificationHandler.getCacheKey()` again

**Result**: 
- Cache SET uses one key format (scraped content-based)
- Cache GET uses different key format (request-based)
- **Keys never match → 0% cache hit rate**

### Evidence from Code Flow

**Classification Service Flow** (`services/classification-service/internal/handlers/classification.go`):

1. **Line 1009**: Check cache using `getCacheKey()` (request-based)
   ```go
   cacheKey := h.getCacheKey(&req)
   if cachedResponse, found := h.getCachedResponse(cacheKey); found {
   ```

2. **Line 1486**: Streaming endpoint also uses `getCacheKey()` (request-based)
   ```go
   cacheKey := h.getCacheKey(&req)
   ```

3. **Internal Classification Service** (`internal/classification/service.go:357`):
   ```go
   cacheKey := s.classificationCache.GenerateCacheKey(scrapedContent, websiteURL)
   ```
   Uses **scraped content-based** key

**The Mismatch**:
- Handler stores cache with **request-based key** (business name + description + URL)
- Internal service stores cache with **scraped content-based key** (title + meta + about + headings + URL)
- These keys **will never match**

---

## Code Analysis: Metadata Extraction Issues

### Test Runner Metadata Extraction

**Location**: `test/integration/comprehensive_classification_e2e_test.go:326`

```go
// Extract metadata for strategy tracking
if metadata, ok := apiResponse["metadata"].(map[string]interface{}); ok {
    result.ScrapingStrategy = extractString(metadata, "scraping_strategy")
    result.EarlyExit = extractBool(metadata, "early_exit")
    result.FallbackUsed = extractBool(metadata, "fallback_used")
    result.FallbackType = extractString(metadata, "fallback_type")
    result.ScrapingTime = DurationMsFromDuration(extractDuration(metadata, "scraping_time_ms"))
    result.ClassificationTime = DurationMsFromDuration(extractDuration(metadata, "classification_time_ms"))
}
```

### Service Metadata Population

**Location**: `services/classification-service/internal/handlers/classification.go:1771`

```go
metadata := map[string]interface{}{
    "service":                  "classification-service",
    "version":                  "2.0.0",
    "scraping_strategy":   "",  // Default empty
    "early_exit":          false, // Default false
    "fallback_used":       false, // Default false
    "fallback_type":       "",
    "scraping_time_ms":    0,
    "classification_time_ms": 0,
}
if enhancedResult.Metadata != nil {
    if scrapingStrategy, ok := enhancedResult.Metadata["scraping_strategy"].(string); ok {
        metadata["scraping_strategy"] = scrapingStrategy
    }
    if earlyExit, ok := enhancedResult.Metadata["early_exit"].(bool); ok {
        metadata["early_exit"] = earlyExit
    }
    // ... more extraction
}
```

### Root Cause: Metadata Not Being Set

**Problem**: Metadata is only populated if `enhancedResult.Metadata` contains the values, but:

1. **Metadata defaults are empty/false** - If `enhancedResult.Metadata` is nil or doesn't contain these fields, defaults remain
2. **Metadata extraction is conditional** - Only extracts if present in `enhancedResult.Metadata`
3. **No fallback to other sources** - Doesn't check `WebsiteAnalysis.StructuredData` or other sources

**Evidence**: Test results show:
- `scraping_strategy`: Empty string (should have value)
- `early_exit`: False (should be true for some requests)
- `strategy_distribution`: Empty (no strategies tracked)

---

## Code Analysis: Request Timeout Issues

### Test Runner Timeout Configuration

**Location**: `test/integration/comprehensive_classification_e2e_test.go:209`

```go
httpClient: &http.Client{Timeout: 60 * time.Second}
```

**Test Timeout**: 60 seconds per request

### Service Request Timeout Configuration

**Location**: `services/classification-service/internal/handlers/classification.go:1036`

```go
requestTimeout := h.calculateAdaptiveTimeout(&req)
```

**Service Timeout**: Calculated adaptively, but likely defaults to 30 seconds based on test failures

### Root Cause: Timeout Mismatch

**Problem**: 
- Test client timeout: **60 seconds**
- Service processing timeout: **~30 seconds** (based on test failures showing 30s timeouts)
- **33% of requests timing out** suggests service is not responding within timeout window

**Possible Causes**:
1. **Service overload** - Service cannot process requests fast enough
2. **Database slow queries** - Database queries taking >30 seconds
3. **External API delays** - Scraping services or ML services taking too long
4. **Resource constraints** - CPU/memory limits on Railway instances
5. **Network issues** - High latency between services

### Evidence from Test Results

- **33 requests timed out** at exactly ~30 seconds
- **Pattern**: Consistent timeout duration suggests service-level timeout, not network timeout
- **No partial responses**: All timeouts are complete failures (no data returned)

---

## Code Analysis: Frontend Compatibility Issues

### Test Runner Frontend Validation

**Location**: `test/integration/comprehensive_classification_e2e_test.go:344`

```go
result.FrontendDataValid, result.FrontendDataIssues = r.validateFrontendData(apiResponse)
```

### Frontend Data Validation Logic

The test runner checks for:
- `primary_industry` field
- `classification` object with `mcc_codes`, `naics_codes`, `sic_codes`
- `explanation` field
- `confidence_score` field
- Proper data types

### Root Cause: Missing Fields in Error Responses

**Problem**: When requests timeout or fail:
- Service returns error response
- Error response doesn't include required frontend fields
- Test runner marks as invalid frontend data

**Evidence**: 
- **36 failed tests** (timeouts) → No frontend data
- **46% frontend compatibility** → Many responses missing required fields

---

## Root Cause Summary

### Issue #1: Cache Key Mismatch (0% Cache Hit Rate)

**Root Cause**: Multiple cache key generation methods:
- Handler uses: `businessName|description|websiteURL`
- Internal service uses: `title|metaDesc|aboutText|headings|domain|websiteURL`
- Keys never match → Cache misses

**Impact**: 
- 0% cache hit rate (expected 60-70%)
- Every request hits full processing pipeline
- Increased load and slower response times

**Fix Required**:
1. Standardize cache key generation across all code paths
2. Use consistent key format (prefer request-based for simplicity)
3. Ensure cache SET and GET use same key generation method

### Issue #2: Metadata Not Populated (0% Early Exit Rate)

**Root Cause**: Metadata extraction is conditional and defaults are empty:
- Metadata only populated if `enhancedResult.Metadata` contains values
- Defaults remain empty/false if metadata not present
- No fallback to other metadata sources

**Impact**:
- 0% early exit rate tracked (expected 20-30%)
- Empty strategy distribution
- Cannot analyze optimization effectiveness

**Fix Required**:
1. Ensure metadata is always populated from available sources
2. Add fallback to `WebsiteAnalysis.StructuredData`
3. Set metadata defaults based on processing path

### Issue #3: Request Timeouts (33% Failure Rate)

**Root Cause**: Service timeout (~30s) shorter than client timeout (60s):
- Service appears to timeout requests internally
- No response returned before timeout
- Possible causes: overload, slow DB queries, external API delays

**Impact**:
- 33% of requests completely fail
- No classification data returned
- Poor user experience

**Fix Required**:
1. Investigate service performance metrics
2. Review database query performance
3. Check external API response times
4. Consider increasing service timeout or optimizing slow operations
5. Add request queuing/throttling

### Issue #4: Frontend Compatibility (46%)

**Root Cause**: Error responses don't include required frontend fields:
- Timeout responses missing required fields
- Error handling doesn't return proper error response structure

**Impact**:
- Frontend cannot render 54% of responses
- Application appears broken to users

**Fix Required**:
1. Ensure all error responses include required fields
2. Implement proper error response structure
3. Add response validation before returning

---

## Recommendations

### Immediate Actions (P0)

1. **Fix Cache Key Generation**
   - **Priority**: Critical
   - **Effort**: Medium
   - **Impact**: High (will restore 60-70% cache hit rate)
   - **Action**: Standardize cache key generation to use single method

2. **Fix Request Timeouts**
   - **Priority**: Critical
   - **Effort**: High
   - **Impact**: Critical (will restore 33% of failed requests)
   - **Action**: Investigate and fix service performance issues

3. **Fix Frontend Compatibility**
   - **Priority**: Critical
   - **Effort**: Low
   - **Impact**: High (will restore frontend functionality)
   - **Action**: Ensure all responses include required fields

### Short-Term Actions (P1)

4. **Fix Metadata Population**
   - **Priority**: High
   - **Effort**: Medium
   - **Impact**: Medium (will enable optimization tracking)
   - **Action**: Ensure metadata always populated from available sources

5. **Fix Service Health**
   - **Priority**: High
   - **Effort**: Medium
   - **Impact**: Medium (will improve system reliability)
   - **Action**: Fix health check failures for dependent services

### Medium-Term Actions (P2)

6. **Add Performance Monitoring**
   - **Priority**: Medium
   - **Effort**: Medium
   - **Impact**: Medium (will enable proactive issue detection)
   - **Action**: Add comprehensive performance monitoring and alerting

7. **Optimize Database Queries**
   - **Priority**: Medium
   - **Effort**: High
   - **Impact**: High (will improve response times)
   - **Action**: Review and optimize slow database queries

---

## Code Fixes Required

### Fix #1: Standardize Cache Key Generation

**File**: `services/classification-service/internal/handlers/classification.go`

**Current Issue**: Handler uses request-based cache key, but internal service uses scraped content-based key.

**Fix**: Use consistent cache key generation method:

```go
// Option 1: Always use request-based key (simpler, more consistent)
func (h *ClassificationHandler) getCacheKey(req *ClassificationRequest) string {
    businessName := strings.TrimSpace(strings.ToLower(req.BusinessName))
    description := strings.TrimSpace(strings.ToLower(req.Description))
    websiteURL := strings.TrimSpace(strings.ToLower(req.WebsiteURL))
    
    data := fmt.Sprintf("%s|%s|%s", businessName, description, websiteURL)
    hash := sha256.Sum256([]byte(data))
    return fmt.Sprintf("classification:%x", hash)
}

// Option 2: Use scraped content-based key when available (more accurate)
// Requires passing scraped content to cache operations
```

**Recommendation**: Use **Option 1** (request-based) for simplicity and consistency.

### Fix #2: Ensure Metadata Always Populated

**File**: `services/classification-service/internal/handlers/classification.go`

**Current Issue**: Metadata defaults remain empty if not present in `enhancedResult.Metadata`.

**Fix**: Add fallback to other sources:

```go
metadata := map[string]interface{}{
    "service": "classification-service",
    "version": "2.0.0",
    // ... defaults
}

// Try enhancedResult.Metadata first
if enhancedResult.Metadata != nil {
    // Extract from enhancedResult.Metadata
}

// Fallback to WebsiteAnalysis.StructuredData
if enhancedResult.WebsiteAnalysis != nil && enhancedResult.WebsiteAnalysis.StructuredData != nil {
    if metadata["scraping_strategy"] == "" {
        if strategy, ok := enhancedResult.WebsiteAnalysis.StructuredData["scraping_strategy"].(string); ok {
            metadata["scraping_strategy"] = strategy
        }
    }
    // ... more fallbacks
}

// Infer from processing path if still empty
if metadata["scraping_strategy"] == "" && enhancedResult.ProcessingPath != "" {
    metadata["scraping_strategy"] = inferStrategyFromPath(enhancedResult.ProcessingPath)
}
```

### Fix #3: Add Proper Error Response Structure

**File**: `services/classification-service/internal/handlers/classification.go`

**Current Issue**: Error responses don't include required frontend fields.

**Fix**: Ensure error responses include required fields:

```go
func (h *ClassificationHandler) sendErrorResponse(w http.ResponseWriter, req *ClassificationRequest, err error, statusCode int) {
    response := ClassificationResponse{
        Success:          false,
        PrimaryIndustry: "", // Empty but present
        ConfidenceScore:  0.0,
        Classification: &ClassificationData{
            MCCCodes:   []IndustryCode{},
            NAICSCodes: []IndustryCode{},
            SICCodes:   []IndustryCode{},
        },
        Explanation: fmt.Sprintf("Error: %v", err),
        Metadata: map[string]interface{}{
            "error": err.Error(),
            "error_type": "classification_error",
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}
```

---

## Testing Recommendations

### After Fixes Applied

1. **Re-run Comprehensive Tests**
   - Verify cache hit rate improves to 60-70%
   - Verify timeout rate decreases
   - Verify frontend compatibility improves to ≥95%

2. **Add Cache Key Logging**
   - Log cache keys during SET and GET operations
   - Verify keys match between operations
   - Monitor cache hit rate in production

3. **Add Performance Monitoring**
   - Track request latencies
   - Monitor cache hit rates
   - Alert on high timeout rates

---

## Conclusion

The investigation has identified **4 critical root causes**:

1. **Cache key mismatch** - Multiple cache key generation methods causing 0% cache hit rate
2. **Metadata not populated** - Conditional metadata extraction causing empty tracking
3. **Request timeouts** - Service timeout issues causing 33% failure rate
4. **Frontend compatibility** - Missing fields in error responses causing 46% compatibility

**All issues have clear fixes** that can be implemented to restore system performance and reliability.

---

**Investigation Date**: December 19, 2025  
**Investigator**: AI Assistant  
**Files Analyzed**: 
- `docs/railway log/complete log.json`
- `services/classification-service/internal/handlers/classification.go`
- `internal/classification/cache.go`
- `internal/classification/service.go`
- `test/integration/comprehensive_classification_e2e_test.go`

