# Playwright Scraper Service Connectivity Audit - Track 6.2

## Executive Summary

Investigation of Playwright scraper service connectivity reveals **service may not be configured or deployed** in production. The service is used as a fallback strategy (Strategy 3) in the multi-tier scraping approach, but previous analysis indicates significant issues with browser pool management and timeouts.

**Status**: ⚠️ **HIGH** - Service configuration and deployment status unclear

## Service Configuration

### Environment Variable

**Location**: `internal/external/website_scraper.go:73`

```go
playwrightServiceURL := os.Getenv("PLAYWRIGHT_SERVICE_URL")
```

**Expected Value** (from `railway.env.example`):
- Local/Development: `http://playwright-scraper:3000`
- Production: Not specified in `railway.json`

**Status**: ⚠️ **UNCLEAR** - Not configured in `railway.json`, may be set via Railway dashboard

### Service Initialization

**Location**: `internal/external/website_scraper.go:138-153`

**Initialization Logic**:
```go
if playwrightServiceURL != "" {
    playwrightClient := &http.Client{Timeout: 60 * time.Second}
    strategies = append(strategies, &PlaywrightScraper{
        serviceURL: playwrightServiceURL,
        client:     playwrightClient,
        logger:     logger,
    })
    logger.Info("✅ [Scraper] Playwright strategy enabled",
        zap.String("service_url", playwrightServiceURL))
} else {
    logger.Info("ℹ️ [Scraper] Playwright strategy disabled (PLAYWRIGHT_SERVICE_URL not set)")
}
```

**Status**: ✅ Conditional initialization - Service only enabled if URL is configured

### Service Role in Multi-Tier Scraping

**Location**: `internal/external/website_scraper.go:114-153`

**Strategy Order**:
1. **Strategy 0**: hrequests (if `HREQUESTS_SERVICE_URL` configured)
2. **Strategy 1**: SimpleHTTP (always enabled)
3. **Strategy 2**: BrowserHeaders (always enabled)
4. **Strategy 3**: Playwright (if `PLAYWRIGHT_SERVICE_URL` configured) - **Fallback**

**Status**: ✅ Playwright is used as fallback when other strategies fail

## Service Endpoints

### Health Check Endpoint

**Endpoint**: `GET /health`

**Expected Response**:
```json
{
  "status": "healthy",
  "message": "Service is operational"
}
```

**Usage**: Service health verification

### Scrape Endpoint

**Endpoint**: `POST /scrape`

**Request Body**:
```json
{
  "url": "https://example.com"
}
```

**Response**:
```json
{
  "html": "<html>...</html>",
  "success": true,
  "requestId": "...",
  "metrics": {
    "scrapeDurationMs": 1234,
    "totalDurationMs": 2345,
    "queueWaitTimeMs": 111
  }
}
```

**Error Response**:
```json
{
  "html": "",
  "error": "Error message",
  "success": false
}
```

**Status**: ✅ Endpoint documented in `services/playwright-scraper/README.md`

## Known Issues from Previous Analysis

### Issue 1: Browser Pool Exhaustion ⚠️ **CRITICAL**

**From**: `docs/comprehensive-test-analysis-latest.md`

**Problem**:
- Browsers are not being released properly after requests complete
- Browsers may be stuck in "in use" state
- Mutex or browser pool management has a bug preventing browser release

**Evidence**:
```json
{
  "error": "Timeout waiting for available browser",
  "queueWaitTimeMs": 3307736, // 55+ minutes!
  "totalDurationMs": 3312774
}
```

**Impact**:
- 100% of Playwright strategy attempts fail
- All requests that require Playwright (JavaScript-heavy sites) fail
- Service becomes completely unresponsive

**Status**: ⚠️ **UNRESOLVED** - Needs investigation

### Issue 2: Context Deadline Exceeded ⚠️ **HIGH**

**From**: `docs/comprehensive-test-analysis-latest.md`

**Problem**:
- Playwright HTTP client is timing out after ~20-22 seconds
- Error: "context deadline exceeded (Client.Timeout exceeded while awaiting headers)"
- Happens even when Playwright service is healthy

**Evidence**:
```
Post "http://playwright-scraper:3000/scrape": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
duration: 22.028141739s
```

**Root Cause**:
- HTTP client timeout (20s) is expiring before Playwright service can respond
- Playwright service is taking too long due to browser pool exhaustion
- Context deadline propagation issue

**Impact**:
- All Playwright strategy attempts fail with timeout
- Even if browser pool was fixed, requests would still timeout

**Status**: ⚠️ **UNRESOLVED** - Needs investigation

### Issue 3: HTTP 429 Rate Limiting ⚠️ **HIGH**

**From**: `docs/comprehensive-test-analysis-latest.md`

**Problem**:
- Many target websites are returning `429 Too Many Requests`
- Affects SimpleHTTP and BrowserHeaders strategies
- Rate limiting is expected, but hitting it too frequently

**Evidence**:
```
HTTP error: 429 429 Too Many Requests
```

**Affected Sites**:
- google.com
- Multiple other high-traffic sites

**Root Cause**:
- Too many concurrent requests to the same domain
- No rate limiting or backoff strategy in scraper
- User-Agent may be flagged as a bot

**Impact**:
- SimpleHTTP and BrowserHeaders strategies fail for rate-limited sites
- Forces fallback to Playwright, which is also failing
- Reduces overall success rate

**Status**: ⚠️ **UNRESOLVED** - Needs investigation

## Service Implementation

### Retry Logic

**Location**: `internal/external/website_scraper.go:997-1104`

**Configuration**:
- **Max Retries**: 2 (3 total attempts)
- **Backoff**: Exponential (1s, 2s)
- **Retry Conditions**:
  - Network errors
  - Timeouts (but not context cancellation)
  - HTTP 5xx errors

**Status**: ✅ Retry logic implemented

### Error Handling

**Location**: `internal/external/website_scraper.go:1037-1072`

**Error Categories**:
1. **Network Errors**: Retryable
2. **HTTP 4xx Errors**: Not retryable (except 429)
3. **HTTP 5xx Errors**: Retryable
4. **Context Cancellation**: Not retryable
5. **Service Errors**: Not retryable

**Status**: ✅ Error handling implemented

### Timeout Configuration

**Location**: `internal/external/website_scraper.go:143`

**HTTP Client Timeout**: 60 seconds

**Context Timeout**: Respects context deadline (typically 20s)

**Status**: ⚠️ **POTENTIAL ISSUE** - Client timeout (60s) may exceed context deadline (20s)

## Investigation Steps

### Step 1: Check Service Configuration

**Check Environment Variable**:
```bash
# Check if PLAYWRIGHT_SERVICE_URL is set in Railway
# Via Railway dashboard or logs
```

**Expected**:
- Production: `https://playwright-scraper-production.up.railway.app` (or similar)
- Staging: `https://playwright-scraper-staging.up.railway.app` (or similar)

**Status**: ⏳ **PENDING** - Need to verify in Railway dashboard

### Step 2: Test Service Connectivity

**Test Script**: `scripts/test_playwright_service.go`

**Tests**:
1. Health check (`/health`)
2. Scrape simple website (`example.com`)
3. Scrape JavaScript-heavy website (`github.com`)
4. Scrape invalid URL (error handling)

**Usage**:
```bash
go run scripts/test_playwright_service.go https://playwright-scraper-production.up.railway.app
```

**Status**: ⏳ **PENDING** - Need to run tests

### Step 3: Review Service Errors

**Check Railway Logs**:
- Look for Playwright service errors
- Check for browser pool exhaustion
- Review timeout patterns
- Analyze error distribution

**Error Patterns to Look For**:
- "Timeout waiting for available browser"
- "context deadline exceeded"
- "playwright service error: 5xx"
- "playwright service URL is not configured"

**Status**: ⏳ **PENDING** - Need to analyze logs

### Step 4: Review Service Deployment

**Check Railway Deployment**:
- Verify Playwright service is deployed
- Check service health
- Review resource allocation
- Check service logs

**Status**: ⏳ **PENDING** - Need to verify deployment

## Root Cause Analysis

### Primary Issues

1. **Service Configuration** ⚠️ **HIGH**
   - `PLAYWRIGHT_SERVICE_URL` may not be configured in production
   - Service may not be deployed
   - **Impact**: Playwright strategy disabled, no fallback for JavaScript-heavy sites

2. **Browser Pool Exhaustion** ⚠️ **CRITICAL**
   - Browsers not being released properly
   - Service becomes unresponsive
   - **Impact**: 100% failure rate for Playwright strategy

3. **Timeout Mismatches** ⚠️ **HIGH**
   - HTTP client timeout (60s) vs context deadline (20s)
   - Service taking too long due to browser pool issues
   - **Impact**: All requests timeout before completion

4. **Rate Limiting** ⚠️ **MEDIUM**
   - HTTP 429 errors from target sites
   - Forces fallback to Playwright
   - **Impact**: Reduced success rate

## Recommendations

### Immediate Actions (High Priority)

1. **Verify Service Configuration**:
   - Check if `PLAYWRIGHT_SERVICE_URL` is set in Railway
   - Verify service is deployed
   - Test service connectivity

2. **Fix Browser Pool Management**:
   - Review browser pool implementation
   - Fix browser release logic
   - Add browser pool monitoring

3. **Align Timeout Configurations**:
   - Reduce HTTP client timeout to match context deadline
   - Or increase context deadline to allow more time
   - Ensure timeouts are consistent

### Medium Priority Actions

4. **Improve Error Handling**:
   - Better error categorization
   - More informative error messages
   - Better retry logic

5. **Add Rate Limiting**:
   - Implement rate limiting for target sites
   - Add backoff strategy
   - Rotate User-Agents

6. **Add Monitoring**:
   - Track browser pool usage
   - Monitor service health
   - Alert on failures

## Code Locations

- **Service Configuration**: `internal/external/website_scraper.go:70-162`
- **Playwright Scraper**: `internal/external/website_scraper.go:951-1104`
- **Service Implementation**: `services/playwright-scraper/index.js`
- **Service README**: `services/playwright-scraper/README.md`

## Next Steps

1. ✅ **Complete Track 6.2 Investigation** - This document
2. **Verify Service Configuration** - Check Railway dashboard
3. **Test Service Connectivity** - Use test script
4. **Review Service Errors** - Analyze Railway logs
5. **Fix Browser Pool Issues** - If service is deployed
6. **Align Timeout Configurations** - Fix timeout mismatches

## Expected Impact

After fixing issues:

1. **Scraping Success Rate**: 10.4% → ≥70% (with Playwright fallback)
2. **JavaScript-Heavy Sites**: 0% → >80% success rate
3. **Service Reliability**: Improved with browser pool fixes
4. **Error Rate**: Reduced with better timeout handling

## References

- Previous Analysis: `docs/comprehensive-test-analysis-latest.md`
- Service README: `services/playwright-scraper/README.md`
- Scraper Implementation: `internal/external/website_scraper.go`
- Test Script: `scripts/test_playwright_service.go`

