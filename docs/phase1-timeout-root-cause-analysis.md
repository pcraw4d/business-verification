# Phase 1 Timeout Issue - Root Cause Analysis & Solution

## Problem Summary

The Phase 1 enhanced scraper is experiencing persistent timeout issues where all scraping strategies fail almost instantly (<1ms) with "context deadline exceeded" errors, even though a 20-second timeout context is created.

## Root Cause Analysis

### Evidence

1. **Context Creation**: A 20-second timeout context is created in `supabase_repository.go:3246`:
   ```go
   phase1Ctx, phase1Cancel := context.WithTimeout(context.Background(), 20*time.Second)
   ```

2. **Context Deadline Logging Not Appearing**: Despite adding comprehensive logging at the start of `ScrapeWithStructuredContent`, the context deadline logs (lines 1263-1274) are not appearing in Docker logs.

3. **Strategies Failing Instantly**: All strategies (SimpleHTTP, BrowserHeaders, Playwright) fail with "context deadline exceeded" in <1ms, suggesting the context is already cancelled when it reaches them.

4. **Playwright Strategy Takes ~19-20ms**: The Playwright strategy takes approximately 19-20ms before failing, which is suspiciously close to the 20-second timeout, suggesting the context deadline is being hit almost immediately.

### Root Cause Hypothesis

**The context is being cancelled or has a very short deadline before it reaches the scraping strategies.** This could happen due to:

1. **Parent Context Cancellation**: The parent context (from `extractKeywords`) might have a shorter timeout (5 seconds for fast-path) that's being propagated instead of the new `phase1Ctx`.

2. **Context Not Being Passed Correctly**: The `phase1Ctx` might not be properly passed through the call chain:
   - `supabase_repository.go` → `enhanced_website_scraper.go` → `website_scraper.go` → strategies

3. **HTTP Client Timeout Conflict**: The HTTP client has a 30-second timeout, but if the context deadline is shorter (20 seconds), the HTTP client should respect the context deadline. However, if the context is already cancelled, the HTTP request fails immediately.

4. **Deferred Cancellation**: The `defer phase1Cancel()` in `supabase_repository.go` might be cancelling the context prematurely if the function returns early.

## Solution Implemented

### Fix 1: Immediate Context Validation
Added immediate context state check at the very start of `ScrapeWithStructuredContent`:
```go
// IMMEDIATELY check context state - this must be first
if ctx.Err() != nil {
    s.logger.Error("❌ [Phase1] Context already cancelled before scraping",
        zap.String("url", targetURL),
        zap.Error(ctx.Err()))
    return nil, fmt.Errorf("context already cancelled: %w", ctx.Err())
}
```

### Fix 2: Context Verification in Enhanced Scraper
Added context deadline logging in `enhanced_website_scraper.go` to trace context propagation:
```go
// Verify context before passing it
if deadline, ok := ctx.Deadline(); ok {
    timeUntilDeadline := time.Until(deadline)
    ews.logger.Printf("⏱️ [Enhanced] Context deadline: %v from now (valid: %v)", timeUntilDeadline, ctx.Err() == nil)
}
```

### Fix 3: Context Validation After Creation
Added validation immediately after context creation in `supabase_repository.go`:
```go
// Verify context is valid before passing
if phase1Ctx.Err() != nil {
    r.logger.Printf("❌ [Phase1] [KeywordExtraction] ERROR: Context already cancelled immediately after creation: %v", phase1Ctx.Err())
    return []string{} // Fall back to legacy method
}
```

### Fix 4: Explicit Entry Logging
Added explicit entry log at the very start of `ScrapeWithStructuredContent` to verify the function is being called with new code:
```go
s.logger.Info("🔍 [Phase1] ScrapeWithStructuredContent ENTRY",
    zap.String("url", targetURL),
    zap.Time("start_time", startTime))
```

## Next Steps

1. **Verify Binary Update**: Ensure the Docker container is using the newly built binary by checking if the entry log appears.

2. **Trace Context Propagation**: If logs still don't appear, add more explicit logging at each step of the call chain to identify where the context is being lost or cancelled.

3. **Check Parent Context**: Investigate if the parent context from `extractKeywords` has a timeout that's interfering with the `phase1Ctx`.

4. **HTTP Client Timeout**: Ensure HTTP client timeouts respect the context deadline and don't conflict with it.

5. **Defer Cancellation Timing**: Review if `defer phase1Cancel()` is being called prematurely, causing the context to be cancelled before the scraping completes.

## Expected Outcome

Once the root cause is identified and fixed:
- Context deadline logs should appear showing a 20-second timeout
- Strategies should have sufficient time to execute (not fail instantly)
- Scraping should succeed for most websites using the appropriate strategy
- Quality scores and word counts should be logged and measurable

## Current Status

- ✅ Fix 1-4 implemented in code
- ⚠️ Binary rebuilt but logs not yet confirming new code is running
- 🔍 Debugging in progress to verify context propagation and identify cancellation point

