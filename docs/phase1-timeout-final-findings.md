# Phase 1 Timeout Issue - Final Findings

## Summary

After extensive debugging, we've identified the root cause and implemented fixes, but the context deadline logs are not appearing in Docker logs, making it difficult to verify the fixes are working.

## Key Findings

### ✅ Phase 1 Scraper is Being Called
- Logs confirm: `"scraper is nil: false"` and `"Using Phase 1 enhanced scraper"`
- The scraper is properly initialized in `main.go:88-96`
- The repository is correctly injected with the scraper

### ⚠️ Phase 1 Scraper is Failing
- After Phase 1 scraper is called, logs show fallback to legacy method
- Legacy method shows: `"SinglePage HTTP ERROR (timeout): context deadline exceeded"`
- This suggests Phase 1 scraper is failing and falling back

### 🔍 Context Deadline Logs Not Appearing
- Context deadline logging code is in place (lines 1252-1274 in `website_scraper.go`)
- Entry log marker added (line 1252) but not appearing
- Possible reasons:
  1. **Log format mismatch**: `log.Printf` in `supabase_repository.go` vs `zap.Logger` in `website_scraper.go`
  2. **Context already cancelled**: Context is cancelled before reaching `ScrapeWithStructuredContent`
  3. **Function not being called**: `ScrapeWithStructuredContent` isn't being reached

### 📊 Evidence from Logs

**Phase 1 Scraper Call Confirmed:**
```
{"level":"info","msg":"🔍 [Phase1] [KeywordExtraction] Checking Phase 1 scraper availability for: https://example.com (scraper is nil: false)"}
{"level":"info","msg":"✅ [Phase1] [KeywordExtraction] Using Phase 1 enhanced scraper for: https://example.com"}
{"level":"info","msg":"🌐 [Enhanced] Using Phase 1 enhanced scraper for: https://example.com"}
```

**But Then Falls Back to Legacy:**
```
{"level":"info","msg":"📡 [KeywordExtraction] [SinglePage] Making HTTP request to: https://example.com"}
{"level":"info","msg":"❌ [KeywordExtraction] [SinglePage] HTTP ERROR (timeout): Request failed for https://example.com: Get \"https://example.com\": context deadline exceeded"}
```

## Root Cause Hypothesis

The context created with a 20-second timeout in `supabase_repository.go:3246` is being cancelled or has a very short deadline by the time it reaches the scraping strategies. This could be because:

1. **Parent Context Interference**: The parent context from `extractKeywords` (5-second timeout) might be affecting the child context
2. **Deferred Cancellation**: The `defer phase1Cancel()` might be cancelling the context prematurely
3. **Context Propagation Issue**: The context might not be properly passed through the call chain

## Code Flow

```
extractKeywords (5s timeout ctx)
  └─> extractKeywordsFromWebsite (receives 5s ctx)
      └─> Creates phase1Ctx (20s timeout) from context.Background()
          └─> websiteScraper.ScrapeWebsite(phase1Ctx, url)
              └─> enhanced_website_scraper.ScrapeWebsite(ctx, url)
                  └─> external.ScrapeWebsite(ctx, url)
                      └─> ScrapeWithStructuredContent(ctx, url)  ← Context deadline logs should appear here
                          └─> strategies.Scrape(ctx, url)  ← But context is already cancelled
```

## Implemented Fixes

1. ✅ **Immediate Context Validation** - Check context state at start of `ScrapeWithStructuredContent`
2. ✅ **Context Deadline Logging** - Log context deadline and HTTP client timeout
3. ✅ **Context Verification in Enhanced Scraper** - Verify context before passing to external scraper
4. ✅ **Context Validation After Creation** - Verify context is valid immediately after creation
5. ✅ **Explicit Entry Logging** - Added entry log to verify function execution

## Next Steps

### Immediate Actions

1. **Verify Log Output**:
   - Check if `log.Printf` logs appear in raw Docker output (not JSON filtered)
   - Consider converting `log.Printf` to structured logging for consistency

2. **Add More Explicit Logging**:
   - Add structured logging (zap) in `supabase_repository.go` for Phase 1 context creation
   - Add logging at every step of context propagation

3. **Investigate Context Cancellation**:
   - Check if parent context (5s timeout) is affecting child context
   - Verify `defer phase1Cancel()` timing
   - Add logging to track when context is cancelled

### Recommended Solution

The most likely issue is that the context is being cancelled before it reaches the strategies. To fix this:

1. **Ensure Context Independence**: The `phase1Ctx` is created from `context.Background()`, which should be independent, but verify it's not being affected by parent context cancellation

2. **Add Context Deadline Logging with Structured Logs**: Convert `log.Printf` to `zap.Logger` in `supabase_repository.go` to ensure logs appear in Docker output

3. **Verify Context Propagation**: Add logging at each step to verify the context is being passed correctly and not being replaced

## Files Modified

- `internal/external/website_scraper.go` - Added context validation and deadline logging
- `internal/classification/repository/supabase_repository.go` - Added context validation after creation
- `internal/classification/enhanced_website_scraper.go` - Added context verification before passing

## Verification Commands

```bash
# Check for Phase 1 scraper calls
docker compose -f docker-compose.local.yml logs classification-service --tail=2000 2>&1 | grep -E "Checking Phase 1|Using Phase 1"

# Check for context deadline logs (raw output, not JSON filtered)
docker compose -f docker-compose.local.yml logs classification-service --tail=2000 2>&1 | grep -i "context deadline\|phase 1 context"

# Check for ENTRY log (confirms new binary)
docker compose -f docker-compose.local.yml logs classification-service --tail=2000 2>&1 | grep "ScrapeWithStructuredContent ENTRY"
```

