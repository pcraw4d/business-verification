# Error Analysis - Classification Service - December 24, 2024

## Executive Summary

Analysis of the most recent classification attempt (`req_1766605568208420262`) and overall error patterns across all services reveals **5 critical issues** causing performance degradation and classification failures:

1. **JSON Unmarshaling Failures** (24 errors) - hrequests service returning HTML instead of JSON
2. **Extreme Performance Degradation** (74+ seconds) - Request taking 2.5x longer than timeout threshold
3. **Scraping Strategy Failures** - All scraping strategies failing for certain websites
4. **HTTP Request Timeouts** - Requests timing out at 60 seconds (2-minute timeout configured)
5. **Context Cancellation** (12 errors) - Premature context cancellation causing cascading failures

## Most Recent Request Analysis

**Request ID**: `req_1766605568208420262`  
**Timestamp**: 2025-12-24T19:46:33Z to 2025-12-24T19:47:23Z  
**Duration**: **74.47 seconds** (exceeded 60s timeout threshold)  
**Status**: Completed but with severe performance issues

### Timeline

1. **19:46:33** - Request started
2. **19:47:08** - HTTP request context cancelled (59.9s elapsed, timeout at 60s)
3. **19:47:23** - Classification completed (74.47s total)
4. **19:47:23** - Performance warnings triggered

### Errors and Warnings

- **2 Errors**: Very slow classification operation (74.45s), Very slow request detected
- **4 Warnings**: HTTP request context cancelled, Slow classification operation, Slow stages detected, Bottleneck detected

## Root Cause Analysis

### 1. JSON Unmarshaling Failures (CRITICAL)

**Frequency**: 24 errors in last 2000 log entries  
**Error Message**: `Time.UnmarshalJSON: input is not a JSON string`  
**Location**: `internal/external/hrequests_client.go:149`

#### Root Cause
The hrequests scraping service is returning **HTML content wrapped in JSON** instead of the expected structured JSON response. The response contains:
- HTML content in the `content` field
- Invalid time format in `scraped_at` field (likely `null` or HTML string instead of ISO 8601)

#### Example Error
```json
{
  "level": "error",
  "msg": "❌ [Hrequests] Failed to unmarshal response",
  "url": "https://rei.com/",
  "error": "Time.UnmarshalJSON: input is not a JSON string",
  "response_body": "{\"content\":{\"about_text\":\"REI: A Life Outdoors...\"}}"
}
```

#### Impact
- **Scraping failures**: All hrequests scraping attempts fail for affected websites
- **Cascading failures**: When hrequests fails, system falls back to other strategies, increasing latency
- **Classification accuracy**: Missing website content leads to lower confidence scores

#### Code Location
```go
// internal/external/hrequests_client.go:147-154
var hrequestsResp HrequestsScrapeResponse
if err := json.Unmarshal(body, &hrequestsResp); err != nil {
    c.logger.Error("❌ [Hrequests] Failed to unmarshal response",
        zap.String("url", url),
        zap.Error(err),
        zap.String("response_body", string(body)))
    return nil, fmt.Errorf("failed to unmarshal response: %w", err)
}
```

#### Expected vs Actual Response

**Expected**:
```json
{
  "success": true,
  "content": {
    "scraped_at": "2025-12-24T19:46:33Z",
    "plain_text": "...",
    "quality_score": 0.95
  }
}
```

**Actual** (from logs):
```json
{
  "content": {
    "about_text": "REI: A Life Outdoors is a Life Well Lived | REI Co-op...",
    // HTML content, not structured JSON
    // scraped_at is likely null or invalid format
  }
}
```

### 2. Extreme Performance Degradation (CRITICAL)

**Frequency**: 4 performance errors, 12 performance warnings  
**Duration**: 74.47 seconds (exceeds 60s timeout threshold by 24%)  
**Bottleneck**: Classification operation taking 74.45 seconds

#### Root Cause
Multiple factors contributing to extreme latency:

1. **Scraping Strategy Failures**: All scraping strategies failing, causing multiple retries
2. **JSON Unmarshaling Retries**: Each failed unmarshal triggers retry, adding latency
3. **Context Cancellation**: HTTP context cancelled at 60s, but processing continues until 74s
4. **No Early Termination**: System doesn't fail fast when all strategies fail

#### Impact
- **User Experience**: Requests taking 2.5x longer than acceptable threshold
- **Resource Exhaustion**: Long-running requests consume worker pool capacity
- **Timeout Mismatch**: HTTP timeout (60s) doesn't match actual processing time (74s+)

#### Performance Breakdown
- **Classification Operation**: 74.45 seconds (99.7% of total time)
- **Other Operations**: <0.1 seconds
- **Bottleneck**: Classification stage (100% of bottleneck)

### 3. Scraping Strategy Failures (HIGH)

**Frequency**: 25 scraping-related errors  
**Error Message**: `❌ [Phase1] All scraping strategies failed`

#### Root Cause
When hrequests fails due to JSON unmarshaling, the system attempts fallback strategies:
1. hrequests (fails - JSON unmarshal error)
2. Playwright (may fail or timeout)
3. Fallback strategies (may fail)

All strategies fail for certain websites (e.g., `https://rei.com/`), causing:
- Complete scraping failure
- Classification proceeds without website content
- Lower confidence scores
- Reduced accuracy

#### Impact
- **Classification Accuracy**: Missing website content reduces confidence
- **Latency**: Multiple strategy attempts increase processing time
- **Resource Usage**: Failed scraping attempts waste resources

### 4. HTTP Request Timeouts (HIGH)

**Frequency**: 2 timeout errors in recent logs  
**Timeout Configuration**: 2 minutes (120 seconds)  
**Actual Timeout**: 60 seconds (HTTP context cancellation)

#### Root Cause
**Mismatch between configured timeout and actual timeout**:
- **Configured**: `REQUEST_TIMEOUT=120s` (2 minutes)
- **Actual**: HTTP request context cancelled at 60 seconds
- **Processing continues**: Classification completes at 74 seconds despite HTTP timeout

#### Code Location
```go
// services/classification-service/cmd/main.go
// timeoutMiddleware sets 2-minute timeout, but HTTP context may cancel earlier
```

#### Impact
- **Client Disconnection**: Client receives timeout error at 60s, but server continues processing
- **Wasted Resources**: Processing continues for 14+ seconds after client disconnects
- **Inconsistent Behavior**: Response may be generated but never sent to client

### 5. Context Cancellation (MEDIUM)

**Frequency**: 12 context cancellation errors  
**Error Message**: `HTTP request context cancelled`

#### Root Cause
HTTP request context is cancelled when:
1. Client disconnects (timeout, network issue)
2. HTTP server timeout (60 seconds)
3. Context deadline exceeded

However, processing continues because:
- Worker context is separate from HTTP context
- Classification processing doesn't check HTTP context status
- Response is generated but can't be sent to disconnected client

#### Impact
- **Resource Waste**: Processing continues after client disconnects
- **Error Handling**: Errors logged but not properly handled
- **User Experience**: Client sees timeout, but server completes processing

## Error Pattern Distribution

From last 2000 log entries:
- **Other errors**: 202 (71.8%)
- **Scraping errors**: 25 (8.9%)
- **JSON unmarshal errors**: 24 (8.5%)
- **Context cancellation**: 12 (4.3%)
- **Panic errors**: 12 (4.3%)
- **Performance errors**: 4 (1.4%)
- **Timeout errors**: 2 (0.7%)

## Opportunities for Improvement

### 1. Fix JSON Unmarshaling (CRITICAL - Priority 1)

**Problem**: hrequests service returning HTML instead of structured JSON

**Solutions**:
1. **Fix hrequests service response format**
   - Ensure `scraped_at` is always ISO 8601 format or `null`
   - Validate response structure before returning
   - Add response schema validation

2. **Add defensive unmarshaling in Go client**
   - Handle `null` or invalid time formats gracefully
   - Use custom `time.Time` unmarshaler with fallback
   - Validate response structure before unmarshaling

3. **Improve error handling**
   - Log full response body for debugging
   - Retry with different strategy on unmarshal failure
   - Fallback to Playwright if hrequests fails

**Expected Impact**:
- Reduce scraping failures by 90%+
- Reduce classification latency by 30-50%
- Improve classification accuracy

**Implementation**:
```go
// internal/external/hrequests_client.go
// Add custom time unmarshaler
type FlexibleTime struct {
    time.Time
}

func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
    // Try standard ISO 8601
    if err := json.Unmarshal(b, &ft.Time); err == nil {
        return nil
    }
    // Try null
    if string(b) == "null" {
        ft.Time = time.Time{}
        return nil
    }
    // Fallback to current time
    ft.Time = time.Now()
    return nil
}

// Update ScrapedContent struct
type ScrapedContent struct {
    // ... other fields ...
    ScrapedAt FlexibleTime `json:"scraped_at"`
}
```

### 2. Implement Fast Failure (CRITICAL - Priority 1)

**Problem**: System continues processing after all strategies fail

**Solutions**:
1. **Early termination on strategy failure**
   - Fail fast when all scraping strategies fail
   - Return partial classification with available data
   - Don't retry indefinitely

2. **Timeout alignment**
   - Align HTTP timeout with processing timeout
   - Cancel processing when HTTP context is cancelled
   - Return error immediately on timeout

3. **Circuit breaker for scraping**
   - Track scraping failure rate
   - Open circuit breaker after threshold failures
   - Skip scraping when circuit is open

**Expected Impact**:
- Reduce latency by 50-70% for failed requests
- Improve resource utilization
- Better user experience with faster error responses

**Implementation**:
```go
// internal/external/website_scraper.go
// Add fast failure logic
func (s *WebsiteScraper) ScrapeWithStructuredContent(ctx context.Context, url string) (*ScrapedContent, error) {
    strategies := []ScrapingStrategy{s.hrequestsClient, s.playwrightClient}
    
    for _, strategy := range strategies {
        // Check context before each attempt
        if ctx.Err() != nil {
            return nil, ctx.Err()
        }
        
        result, err := strategy.Scrape(ctx, url)
        if err == nil {
            return result, nil
        }
        
        // Don't retry on non-transient errors
        if !s.isTransientError(err) {
            return nil, fmt.Errorf("strategy failed with non-transient error: %w", err)
        }
    }
    
    // All strategies failed - fail fast
    return nil, fmt.Errorf("all scraping strategies failed")
}
```

### 3. Improve Timeout Handling (HIGH - Priority 2)

**Problem**: Mismatch between HTTP timeout and processing timeout

**Solutions**:
1. **Align timeouts**
   - Set HTTP timeout to match processing timeout (120s)
   - Use consistent timeout values across layers
   - Propagate timeout context correctly

2. **Check HTTP context during processing**
   - Periodically check `r.Context().Err()` during processing
   - Cancel processing if HTTP context is cancelled
   - Return error immediately on cancellation

3. **Timeout middleware improvements**
   - Log timeout events with request ID
   - Cancel worker processing on HTTP timeout
   - Return proper timeout error response

**Expected Impact**:
- Eliminate wasted processing after client disconnects
- Improve resource utilization
- Consistent timeout behavior

**Implementation**:
```go
// services/classification-service/cmd/main.go
// Improve timeout middleware
func timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()
            
            // Create response writer that tracks timeout
            tw := &timeoutResponseWriter{
                ResponseWriter: w,
                timedOut:       false,
            }
            
            // Monitor for timeout
            go func() {
                <-ctx.Done()
                if ctx.Err() == context.DeadlineExceeded {
                    tw.timedOut = true
                    // Cancel any ongoing processing
                    // This should propagate to worker context
                }
            }()
            
            next.ServeHTTP(tw, r.WithContext(ctx))
        })
    }
}
```

### 4. Add Response Validation (MEDIUM - Priority 3)

**Problem**: Invalid responses from external services cause unmarshaling failures

**Solutions**:
1. **Validate response structure**
   - Check response content-type before unmarshaling
   - Validate JSON structure before parsing
   - Handle malformed responses gracefully

2. **Add response sanitization**
   - Sanitize time fields before unmarshaling
   - Handle null values correctly
   - Strip invalid characters from JSON

3. **Improve error messages**
   - Include response snippet in error logs
   - Log response headers for debugging
   - Provide actionable error messages

**Expected Impact**:
- Better error diagnostics
- Reduced unmarshaling failures
- Faster debugging of issues

### 5. Implement Request Deduplication Timeout (MEDIUM - Priority 3)

**Problem**: Long-running requests block duplicate requests

**Solutions**:
1. **Timeout in-flight requests**
   - Set maximum wait time for duplicate requests
   - Fail fast if in-flight request is taking too long
   - Allow new request to proceed independently

2. **Request cancellation**
   - Cancel in-flight request if duplicate request times out
   - Clean up resources on cancellation
   - Log cancellation events

**Expected Impact**:
- Prevent duplicate requests from blocking
- Improve system responsiveness
- Better resource management

## Recommended Action Plan

### Phase 1: Critical Fixes (Week 1)
1. ✅ Fix JSON unmarshaling in hrequests client (defensive handling)
2. ✅ Implement fast failure for scraping strategies
3. ✅ Align HTTP and processing timeouts
4. ✅ Add response validation

### Phase 2: Performance Improvements (Week 2)
1. ✅ Implement circuit breaker for scraping
2. ✅ Improve timeout handling
3. ✅ Add request deduplication timeout
4. ✅ Optimize retry logic

### Phase 3: Monitoring and Observability (Week 3)
1. ✅ Add metrics for scraping failures
2. ✅ Add alerts for performance degradation
3. ✅ Improve error logging and diagnostics
4. ✅ Add tracing for slow requests

## Metrics to Track

1. **Scraping Success Rate**: Target >95%
2. **Average Classification Latency**: Target <30s
3. **P95 Classification Latency**: Target <60s
4. **JSON Unmarshaling Failure Rate**: Target <1%
5. **Timeout Rate**: Target <5%
6. **Context Cancellation Rate**: Target <2%

## Related Files

- `internal/external/hrequests_client.go` - JSON unmarshaling issue
- `internal/external/website_scraper.go` - Scraping strategy failures
- `services/classification-service/cmd/main.go` - Timeout middleware
- `services/classification-service/internal/handlers/classification.go` - Request processing

## Next Steps

1. **Immediate**: Fix JSON unmarshaling with defensive handling
2. **Short-term**: Implement fast failure and timeout alignment
3. **Long-term**: Add comprehensive monitoring and alerting

