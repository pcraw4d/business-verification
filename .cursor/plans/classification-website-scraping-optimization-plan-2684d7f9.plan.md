<!-- 2684d7f9-f5e6-4320-a7d3-de7da184fc51 9ecf6e8b-e874-4130-a1a2-040d8bb5a57b -->
# Classification Website Scraping Optimization Plan

## Problem Statement

Based on `docs/railway_log_analysis.md`, website scraping timeouts are causing classification failures:

- Multi-page analysis taking 1.5+ minutes
- Sequential processing with 2s delays between pages
- Up to 15 pages analyzed sequentially
- No early exit based on content quality
- Website scraping failures block entire classification request
- ML service is working but requests fail before ML results are returned

**Goal**: Complete website scraping in <5s while maintaining accuracy by:

1. Keeping discovery (robots.txt, sitemap parsing) to find critical pages
2. Using content-quality-based early exit (not strict page limits)
3. Implementing parallel processing
4. Making scraping non-blocking

## Implementation Plan

### Phase 1: Enhanced Discovery with Sitemap Prioritization

**File**: `internal/classification/smart_website_crawler.go`

**Changes**:

1. **Improve sitemap parsing** (currently returns empty):

   - Implement proper XML parsing for sitemap.xml
   - Extract URLs from sitemap entries
   - Prioritize sitemap URLs by page type (about, services, products, etc.)
   - Use sitemap URLs as primary source for critical pages

2. **Enhance page prioritization**:

   - After sitemap parsing, merge with homepage links and common patterns
   - Prioritize pages from sitemap that match high-priority patterns (about, services, products)
   - Ensure homepage is always first
   - Limit discovered pages to top 20 by priority (before analysis)

**Location**: Lines 469-500 (parseSitemap), 428-467 (discoverSiteStructure), 602-639 (prioritizePages)

### Phase 2: Content-Quality-Based Early Exit

**File**: `internal/classification/smart_website_crawler.go`

**Changes**:

1. **Add content sufficiency check function**:
   ```go
   func (c *SmartWebsiteCrawler) hasSufficientContent(analyses []PageAnalysis) bool {
       // Check multiple criteria:
       // 1. Total content length >= 500 characters (optimal threshold)
       // 2. At least 10 unique keywords extracted
       // 3. Average relevance score >= 0.7
       // 4. At least 2 successful pages with content
   }
   ```

2. **Modify analyzePages to exit early based on content quality**:

   - After each page analysis, check `hasSufficientContent()`
   - Exit early if sufficient content is gathered (regardless of page count)
   - Minimum 2 pages analyzed before early exit (to ensure quality)
   - Log early exit reason (content length, keywords, confidence)

**Location**: Lines 752-866 (analyzePages), add new function after line 920

**Content Quality Thresholds** (from `internal/classification/methods/ml_method.go`):

- Minimum: 50 characters
- Recommended: 100 characters  
- Optimal: 500 characters
- Use 500 characters as "sufficient" threshold for early exit

### Phase 3: Parallel Page Processing with Concurrency Control

**File**: `internal/classification/smart_website_crawler.go`

**Changes**:

1. **Add parallel analysis method**:
   ```go
   func (c *SmartWebsiteCrawler) analyzePagesParallel(ctx context.Context, pages []string, maxConcurrent int) []PageAnalysis
   ```


   - Use semaphore pattern to limit concurrent requests (default: 3)
   - Process pages in parallel with controlled concurrency
   - Collect results in order
   - Check for sufficient content after each batch completes
   - Early exit if sufficient content gathered

2. **Modify CrawlWebsite to use parallel processing**:

   - Replace sequential `analyzePages()` call with `analyzePagesParallel()`
   - Use configurable max concurrent pages (default: 3)
   - Maintain session management across parallel requests

**Location**: Lines 414 (analyzePages call), add new function after line 866

**Configuration**: Add `CLASSIFICATION_MAX_CONCURRENT_PAGES` env var (default: 3)

### Phase 4: Reduced Delays with Adaptive Timing

**File**: `internal/classification/smart_website_crawler.go`

**Changes**:

1. **Implement adaptive delay system**:

   - First page: No delay
   - Subsequent pages: 500ms minimum (reduced from 2s)
   - Respect robots.txt crawl delay if > 500ms
   - Skip delay if previous page failed quickly (< 1s) - indicates timeout
   - Skip delay if we have sufficient content and are in early exit mode

2. **Add fast-path delay configuration**:

   - Fast-path mode: 500ms delay
   - Regular mode: 2s delay (current)
   - Make delay configurable via `CLASSIFICATION_CRAWL_DELAY_MS` env var

**Location**: Lines 755-821 (delay logic in analyzePages)

### Phase 5: Fast-Path Mode with Time Constraints

**File**: `internal/classification/smart_website_crawler.go`

**Changes**:

1. **Add CrawlWebsiteFast method**:
   ```go
   func (c *SmartWebsiteCrawler) CrawlWebsiteFast(ctx context.Context, websiteURL string, maxTime time.Duration) (*CrawlResult, error)
   ```


   - Uses same discovery (robots.txt, sitemap) but with time constraints
   - Limits to top 5-8 pages by priority (from sitemap + homepage links)
   - Uses parallel processing (3 concurrent)
   - Uses reduced delays (500ms)
   - Early exit on sufficient content or time limit
   - Returns partial results if time expires

2. **Time-based early exit**:

   - Check remaining time before starting new page
   - Skip remaining pages if < 1s remaining
   - Use available content even if incomplete

**Location**: Add new method after line 426

### Phase 6: Make Website Scraping Non-Blocking

**File**: `internal/classification/multi_method_classifier.go`

**Changes**:

1. **Wrap website scraping in timeout context**:

   - Create 5s timeout context for website scraping
   - If timeout expires, return partial/empty content
   - Don't fail entire classification request
   - Log warning but continue with available data

2. **Graceful degradation**:

   - If website scraping fails/times out, proceed with business name + description
   - Allow ML service to work with available data
   - Return classification result even if website scraping incomplete

**Location**: Lines 1078-1135 (website content extraction), `internal/classification/methods/ml_method.go` lines 562-635

**File**: `internal/classification/repository/supabase_repository.go`

**Changes**:

1. **Make keyword extraction non-blocking**:

   - Wrap `extractKeywordsFromMultiPageWebsite()` in timeout
   - Return empty keywords if timeout expires
   - Continue with single-page or description-only fallback

**Location**: Lines 2470-2498 (Level 1 multi-page analysis)

### Phase 7: Configuration Updates

**File**: `services/classification-service/internal/config/config.go`

**Changes**:

1. **Add new configuration options**:
   ```go
   FastPathScrapingEnabled     bool          // Enable fast-path mode
   MaxConcurrentPages          int           // Max concurrent page requests
   CrawlDelayMs               int           // Delay between pages in ms
   FastPathMaxPages           int           // Max pages for fast path
   WebsiteScrapingTimeout     time.Duration // Overall timeout for scraping
   ```

2. **Environment variables**:

   - `ENABLE_FAST_PATH_SCRAPING=true` (default: true)
   - `CLASSIFICATION_MAX_CONCURRENT_PAGES=3` (default: 3)
   - `CLASSIFICATION_CRAWL_DELAY_MS=500` (default: 500)
   - `CLASSIFICATION_FAST_PATH_MAX_PAGES=8` (default: 8)
   - `CLASSIFICATION_WEBSITE_SCRAPING_TIMEOUT=5s` (default: 5s)

**Location**: Lines 47-99 (config struct and loading)

### Phase 8: Integration Points

**File**: `internal/classification/repository/supabase_repository.go`

**Changes**:

1. **Use fast-path mode in keyword extraction**:

   - Check if fast-path enabled
   - Use `CrawlWebsiteFast()` with 5s timeout
   - Fall back to regular crawl if fast-path fails and time allows

**Location**: Lines 3507-3524 (extractKeywordsFromMultiPageWebsite)

**File**: `internal/classification/methods/ml_method.go`

**Changes**:

1. **Use fast-path for website content extraction**:

   - Use fast-path crawler if enabled
   - Apply 5s timeout to website scraping
   - Continue with available content if timeout expires

**Location**: Lines 562-635 (extractWebsiteContent)

## Expected Outcomes

### Performance Improvements

- **Fast-path scraping**: 2-4s (down from 30-60s)
- **Regular scraping**: 8-12s (down from 60-90s)
- **Request success rate**: >80% (up from ~0%)
- **ML service utilization**: >80% (up from 0%)

### Accuracy Maintenance

- Content-quality-based early exit ensures sufficient information
- Sitemap prioritization finds most critical pages
- Minimum 2 pages analyzed before early exit
- 500+ character threshold ensures quality content

### Reliability Improvements

- Non-blocking website scraping prevents classification failures
- Graceful degradation allows ML service to work with partial data
- Timeout protection prevents request failures
- Parallel processing reduces total time

## Testing Strategy

1. **Unit tests**:

   - Test `hasSufficientContent()` with various scenarios
   - Test parallel page processing
   - Test early exit logic

2. **Integration tests**:

   - Test fast-path mode with real websites
   - Test timeout handling
   - Test graceful degradation

3. **Performance tests**:

   - Measure scraping time improvements
   - Verify <5s target for fast-path
   - Monitor success rates

## Risk Mitigation

1. **Bot detection**: Limit concurrent requests to 3, maintain delays
2. **Content quality**: Use 500+ character threshold, minimum 2 pages
3. **Accuracy**: Keep discovery, prioritize sitemap pages
4. **Reliability**: Non-blocking design, graceful degradation

## Implementation Order

1. Phase 1: Enhanced discovery (sitemap parsing)
2. Phase 2: Content-quality-based early exit
3. Phase 3: Parallel processing
4. Phase 4: Reduced delays
5. Phase 5: Fast-path mode
6. Phase 6: Non-blocking scraping
7. Phase 7: Configuration
8. Phase 8: Integration