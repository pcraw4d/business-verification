# SmartWebsiteCrawler 403 Error Fix

## Problem

REI's website was returning 403 (Forbidden) errors for all scraping attempts, even after implementing bot evasion techniques (session management, header randomization, human-like delays).

## Root Causes Identified

1. **Too many concurrent requests** - 3 concurrent requests were triggering rate limiting
2. **Delays too short** - 2-second delays were insufficient for REI's bot detection
3. **No initial warm-up** - Starting immediately without establishing a session first
4. **Homepage not prioritized** - Not ensuring homepage is visited first to establish cookies

## Solutions Implemented

### 1. Sequential Processing (No Concurrency)
- **Before**: 3 concurrent requests using goroutines and semaphore
- **After**: Sequential processing (one request at a time)
- **Impact**: Eliminates concurrent request patterns that trigger bot detection

### 2. Increased Delays
- **Before**: 2-second base delay with Weibull distribution
- **After**: 5-8 second base delays (increasing for later requests)
- **Impact**: More realistic human-like timing patterns

### 3. Initial Warm-Up Delay
- **Added**: 3-second delay before starting crawl
- **Impact**: Simulates user arriving at website and waiting before navigating

### 4. Homepage First Strategy
- **Added**: Ensures homepage is always visited first in the prioritized list
- **Impact**: Establishes session and cookies before visiting other pages

### 5. Stop on 403
- **Added**: Immediately stops crawling if 403 is received
- **Impact**: Prevents further blocks and wasted requests

## Code Changes

### `analyzePages` Function
- Removed goroutines and concurrent processing
- Changed to sequential loop with delays between requests
- Added 403 detection to stop immediately

### `CrawlWebsite` Function
- Added homepage-first prioritization
- Added initial warm-up delay before starting crawl

## Expected Results

1. **Better success rate** - Sequential requests with longer delays should reduce 403 errors
2. **Session establishment** - Homepage-first strategy ensures cookies are set
3. **More realistic patterns** - Longer delays and warm-up simulate human behavior

## Limitations

REI may still block requests due to:
- **User-Agent detection** - "KYBPlatformBot" is clearly a bot identifier (required for legal compliance)
- **JavaScript requirements** - Some sites require JS execution for access
- **IP reputation** - Repeated requests from same IP may be flagged
- **Advanced bot detection** - TLS fingerprinting, browser fingerprinting, etc.

## Next Steps (If Still Getting 403s)

1. **Enable proxy rotation** - Set `SCRAPING_USE_PROXIES=true` and configure proxy list
2. **Increase delays further** - Consider 10-15 second delays for very strict sites
3. **Reduce page count** - Limit to 5-10 pages instead of 20
4. **Consider browser automation** - For sites requiring JavaScript (heavy solution)

## Configuration

All delays and behavior can be configured via environment variables:
- `SCRAPING_HUMAN_LIKE_TIMING_ENABLED` (default: true)
- `SCRAPING_SESSION_MANAGEMENT_ENABLED` (default: true)
- `SCRAPING_USE_PROXIES` (default: false)

## Testing

Run tests to verify:
```bash
go test ./internal/classification -run TestSmartWebsiteCrawler -v
```

All tests should pass, verifying:
- Sequential processing works correctly
- Delays are applied between requests
- Session management maintains cookies
- 403 detection stops crawling

