# Railway Log Analysis - Classification Service Issues

**Date**: 2025-12-01  
**Analysis Period**: Logs from Railway deployment  
**Services Analyzed**: classification-service, python-ml-service, api-gateway

---

## Executive Summary

✅ **Python ML Service**: Working correctly - receiving and processing requests  
✅ **Service Communication**: Classification service successfully calling Python ML service  
❌ **Root Cause**: Website scraping timeouts causing overall classification failures  
⚠️ **Performance Issue**: Multi-page website analysis taking 1.5+ minutes

---

## Key Findings

### 1. Python ML Service Status ✅ WORKING

**Evidence from Logs**:
- Multiple successful requests: `POST /classify-enhanced HTTP/1.1" 200 OK`
- Service processing requests correctly
- Response times: 1-4 seconds per classification
- No errors in ML service itself

**Sample Log Entries**:
```
INFO: 100.64.0.6:42784 - "POST /classify-enhanced HTTP/1.1" 200 OK
INFO: 100.64.0.6:42790 - "POST /classify-enhanced HTTP/1.1" 200 OK
INFO: 100.64.0.6:42794 - "POST /classify-enhanced HTTP/1.1" 200 OK
```

**Warnings (Non-Critical)**:
- Summarization warnings about `cache_dir` parameter (cosmetic, doesn't affect functionality)
- Content quality warnings for minimal content (expected behavior)

### 2. Classification Service Communication ✅ WORKING

**Evidence from Logs**:
- Ensemble voting messages: "Using ensemble voting: Python ML + Go classification in parallel"
- Classification service is successfully calling Python ML service
- Circuit breaker is allowing requests through

**Sample Log Entry**:
```
Using ensemble voting: Python ML + Go classification in parallel
```

### 3. Website Scraping Timeouts ❌ PRIMARY ISSUE

**Critical Finding**: Website scraping is causing classification failures

**Evidence**:
1. **Single Page Timeouts**:
   ```
   ❌ [KeywordExtraction] [SinglePage] HTTP ERROR (timeout): 
   Request failed for https://www.acme.com: 
   Get "https://www.acme.com": context deadline exceeded
   ```

2. **Multi-Page Analysis Timeouts**:
   ```
   ⚠️ [PageAnalysis] Timeout error for https://www.acme.com/contact 
   (attempt 1/3): Get "https://www.acme.com/contact": 
   context deadline exceeded
   ```

3. **Total Extraction Time**:
   ```
   - Total extraction time: 1m36.345026603s
   ```

4. **Classification Failures After Timeouts**:
   ```
   Classification failed
   ```
   (Appears after long website scraping attempts)

### 4. TLS Handshake Failures ⚠️ SECONDARY ISSUE

**Evidence**:
- Multiple TLS handshake failures when retrying website requests
- Affects website scraping reliability

**Sample Log Entries**:
```
❌ [KeywordExtraction] [HomepageRetry] HTTP ERROR (unknown): 
Request failed (attempt 1, DNS 8.8.8.8:53): 
Get "https://www.acme.com": remote error: tls: handshake failure
```

**Impact**: 
- Causes website scraping to fail
- Forces fallback to keyword-only classification
- Reduces accuracy

### 5. Request Processing Flow

**Successful Flow** (when website scraping works):
1. Classification service receives request
2. Attempts website scraping (5s timeout per page)
3. Calls Python ML service (1-4s response time) ✅
4. Uses ensemble voting (Python ML + Go classifier)
5. Returns classification result

**Failed Flow** (when website scraping times out):
1. Classification service receives request
2. Website scraping attempts timeout (5s per page, multiple pages)
3. Total time exceeds request timeout (30s)
4. Request fails with "Classification failed"
5. Returns 502 error to client

---

## Root Cause Analysis

### Primary Root Cause: Website Scraping Timeout Accumulation

**Problem**:
- Each page has a 5-second timeout
- Multi-page analysis attempts 12+ pages
- Even with parallel processing, total time can exceed 30s request timeout
- When website scraping fails/times out, entire classification fails

**Timeline Example**:
```
15:30:02 - Start page analysis
15:30:02 - Timeout on /contact page (5s timeout)
15:30:02 - Timeout on main page (5s timeout)
15:30:23 - Multiple page timeouts
15:32:23 - More timeouts
15:32:59 - Total extraction time: 1m36s
15:32:59 - Classification failed (exceeded overall timeout)
```

### Secondary Issues

1. **TLS Handshake Failures**: Some websites reject connections
2. **Multi-Page Analysis**: Too many pages attempted (12+ pages)
3. **No Graceful Degradation**: When website scraping fails, entire request fails

---

## Impact Assessment

### On ML Service Utilization

- **Expected**: ML service should be used for all classifications
- **Actual**: ML service is being called, but requests fail before ML results are returned
- **Result**: 0% effective ML utilization (requests fail before completion)

### On Accuracy

- **Expected**: High accuracy with ML service
- **Actual**: Low accuracy (2.55%) because:
  1. Requests fail before ML results are used
  2. System falls back to keyword-only classification
  3. Website content not available for ML analysis

### On Performance

- **Expected**: < 10 seconds per request
- **Actual**: 8-96 seconds (many requests timing out)
- **Issue**: Website scraping taking too long

---

## Recommendations

### Priority 1: Fix Website Scraping Timeout Issue (CRITICAL)

**Problem**: Website scraping timeouts cause entire classification to fail

**Solutions**:

1. **Make Website Scraping Optional/Non-Blocking**
   - Don't fail entire request if website scraping times out
   - Use website content if available, but proceed without it
   - Allow ML service to work with just business name and description

2. **Reduce Multi-Page Analysis**
   - Limit to 3-5 pages instead of 12+
   - Prioritize homepage and about pages
   - Skip pages that timeout quickly

3. **Increase Website Scraping Timeout or Make It Configurable**
   - Current: 5 seconds per page
   - Consider: 3 seconds for faster failure
   - Or: Make timeout configurable per request type

4. **Implement Graceful Degradation**
   - If website scraping fails, continue with available data
   - Don't fail entire classification request
   - Log warning but proceed

### Priority 2: Optimize Request Processing

1. **Parallel Processing**
   - Already implemented but may need tuning
   - Consider reducing concurrent pages if causing issues

2. **Early Exit on Timeout**
   - If overall timeout approaching, skip remaining website scraping
   - Use available data immediately

3. **Cache Website Content**
   - Already implemented (24h TTL)
   - Verify cache is being used effectively

### Priority 3: Improve Error Handling

1. **Better Error Messages**
   - Distinguish between website scraping failures and ML service failures
   - Return partial results when possible

2. **Retry Logic**
   - For website scraping, not for entire request
   - Don't retry entire classification if website fails

### Priority 4: Monitoring and Observability

1. **Track Website Scraping Success Rate**
   - Monitor how often website scraping succeeds vs fails
   - Track average time per request

2. **Track ML Service Utilization**
   - Measure actual ML service usage (not just calls, but successful completions)
   - Track accuracy by classification method

---

## Immediate Actions

### Action 1: Make Website Scraping Non-Blocking

**File**: `internal/classification/multi_method_classifier.go`

**Change**: Modify website scraping to not block classification
- If website scraping fails/times out, continue with available data
- Don't fail entire request

### Action 2: Reduce Multi-Page Analysis

**File**: `services/classification-service/internal/config/config.go`

**Change**: Reduce `CLASSIFICATION_MAX_PAGES_TO_ANALYZE` from 15 to 5
- Focus on most important pages
- Reduce total processing time

### Action 3: Add Timeout Protection

**File**: `services/classification-service/internal/handlers/classification.go`

**Change**: Add early exit if overall timeout approaching
- Check remaining time before starting new operations
- Skip non-critical operations if time is short

---

## Expected Outcomes

After implementing fixes:

1. **Request Success Rate**: Should increase from ~0% to >80%
2. **ML Service Utilization**: Should increase from 0% to >80%
3. **Average Processing Time**: Should decrease from 8-96s to < 10s
4. **Accuracy**: Should improve significantly (ML service will be used)

---

## Log Evidence Summary

### Python ML Service (Working ✅)
- 20+ successful `/classify-enhanced` requests with 200 OK
- Response times: 1-4 seconds
- No critical errors

### Classification Service (Partially Working ⚠️)
- Successfully calling Python ML service
- Ensemble voting working
- Failing due to website scraping timeouts

### Website Scraping (Failing ❌)
- Multiple timeout errors
- TLS handshake failures
- Taking 1.5+ minutes in some cases
- Causing overall classification failures

---

## Conclusion

The Python ML service and classification service are working correctly and communicating properly. The root cause of classification failures is **website scraping timeouts** that cause the entire request to fail before ML results can be returned.

**Key Insight**: The ML service is being called successfully, but requests fail during website scraping, preventing ML results from being used.

**Solution**: Make website scraping non-blocking and implement graceful degradation so that classification can proceed even when website scraping fails.

