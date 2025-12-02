# Fix: Timeout Change to Enable Fast-Path Mode

**Date**: December 2, 2025  
**Issue**: Fast-path mode not being triggered for website scraping  
**Root Cause**: Timeout hardcoded to 10s (fast-path requires <= 5s)

---

## Changes Made

### File: `internal/classification/repository/supabase_repository.go`

#### Change 1: Line 2473 - Timeout from 10s to 5s

**Before**:
```go
// Reduced timeout to 5s for fast-path, but allow up to 10s for regular path
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

r.logger.Printf("📊 [KeywordExtraction] Level 1: Starting multi-page website analysis (max 15 pages, timeout: 10s)")
```

**After**:
```go
// Use 5s timeout to enable fast-path mode (fast-path threshold is 5s)
// Fast-path mode: 5s timeout, 8 pages max, 3 concurrent
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

r.logger.Printf("📊 [KeywordExtraction] Level 1: Starting multi-page website analysis (max 15 pages, timeout: 5s)")
```

#### Change 2: Line 3513 & 3527 - Updated comments and default timeout

**Before**:
```go
// Use the parent context timeout (already set to 10s in calling function)
// ...
timeoutDuration := 10 * time.Second // Default from calling function
```

**After**:
```go
// Use the parent context timeout (set to 5s in calling function to enable fast-path)
// ...
timeoutDuration := 5 * time.Second // Default from calling function (5s to enable fast-path)
```

---

## Expected Behavior After Fix

### Before Fix
```
Timeout: 10s
Mode: REGULAR
Max Pages: 15
Log: [REGULAR] Using regular crawl mode
Result: Requests timing out
```

### After Fix
```
Timeout: 5s
Mode: FAST-PATH
Max Pages: 8
Concurrent: 3
Log: [FAST-PATH] Using fast-path mode
Result: Requests complete in 2-4s
```

---

## Fast-Path Mode Logic

The fast-path mode is triggered when:
```go
if timeoutDuration <= 5*time.Second {
    // Use fast-path mode
    crawlResult, err = crawler.CrawlWebsiteFast(analysisCtx, websiteURL, timeoutDuration, 8, 3)
} else {
    // Use regular crawl mode
    crawlResult, err = crawler.CrawlWebsite(analysisCtx, websiteURL)
}
```

**With 5s timeout**: `5s <= 5s` → **TRUE** → Fast-path mode enabled ✅

---

## Verification Steps

1. ✅ **Code Compiles**: Verified with `go build`
2. ⏳ **Deploy to Railway**: Deploy the fix
3. ⏳ **Check Logs**: Look for `[FAST-PATH]` markers in Railway logs
4. ⏳ **Monitor Performance**: Verify requests complete in 2-4s instead of timing out

---

## Expected Log Messages After Fix

### Fast-Path Mode (Expected)
```
📊 [KeywordExtraction] Level 1: Starting multi-page website analysis (max 15 pages, timeout: 5s)
📊 [KeywordExtraction] [MultiPage] Timeout duration: 5s (threshold: 5s)
🚀 [KeywordExtraction] [MultiPage] [FAST-PATH] Using fast-path mode (timeout: 5s, max pages: 8, concurrent: 3)
🚀 [SmartCrawler] [FAST-PATH] Starting parallel analysis of 8 pages (concurrent: 3, delay: 500ms)
✅ [SmartCrawler] [FAST-PATH] Crawl completed in X.XXs - 8 pages analyzed
```

### Regular Mode (Should Not Appear)
```
🔍 [KeywordExtraction] [MultiPage] [REGULAR] Using regular crawl mode
```

---

## Performance Impact

### Before Fix
- **Timeout**: 10s
- **Mode**: Regular
- **Max Pages**: 15
- **Success Rate**: ~0% (all timing out at Railway gateway)

### After Fix
- **Timeout**: 5s
- **Mode**: Fast-path
- **Max Pages**: 8
- **Concurrent**: 3
- **Expected Success Rate**: >80%
- **Expected Response Time**: 2-4s

---

## Related Files

- **Fixed**: `internal/classification/repository/supabase_repository.go`
- **Analysis**: `docs/railway-logs-analysis-website-scraping.md`
- **Config**: `services/classification-service/internal/config/config.go` (WebsiteScrapingTimeout: 5s)

---

## Next Steps

1. ✅ **Fix Implemented** - Timeout changed from 10s to 5s
2. ⏳ **Deploy to Railway** - Deploy the updated code
3. ⏳ **Monitor Logs** - Check for `[FAST-PATH]` markers
4. ⏳ **Verify Performance** - Confirm requests complete successfully
5. ⏳ **Update Documentation** - Document the fix in deployment notes

---

## Testing

After deployment, test with:
```bash
curl -X POST https://your-classification-service.railway.app/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Company",
    "description": "A test business",
    "website": "https://example.com"
  }'
```

**Expected**: Response in 2-4s with `[FAST-PATH]` logs

---

## Conclusion

The fix changes the timeout from 10s to 5s, which enables fast-path mode for website scraping. This should:
- ✅ Enable fast-path mode (timeout <= 5s threshold)
- ✅ Reduce response time from timeout to 2-4s
- ✅ Improve success rate from ~0% to >80%
- ✅ Reduce Railway gateway timeouts

