<!-- 715e4fa9-6da8-499c-b826-badc59d82634 083fb47d-d0b1-4fa7-b6bb-3f8f9f83ecbb -->
# Web Scraping Legal Compliance Improvements

## Overview

This plan implements high-priority legal and compliance improvements to ensure our web scraping respects website restrictions, follows best practices, and maintains legal compliance. The improvements focus on robots.txt compliance, identifiable User-Agent strings, conservative rate limiting, proper HTTP status code handling, and documentation.

## Current State Analysis

### Existing Implementation

- **Robots.txt**: Simplified parser in `internal/classification/smart_website_crawler.go` (lines 822-861) that only checks for "disallow: /"
- **User-Agent**: Hardcoded browser-like strings in 5 locations without contact information
- **Rate Limiting**: 1-second minimum delay in `internal/classification/repository/supabase_repository.go` (line 224, 287)
- **Status Codes**: Generic error handling, no specific handling for 429/403/503
- **Documentation**: No scraping policy document exists

## Implementation Tasks

### Task 1: Add robots.txt Parser Library

**File**: `go.mod`

Add the robots.txt parsing library:

```go
github.com/temoto/robotstxt v1.1.2
```

**Action**: Run `go get github.com/temoto/robotstxt@v1.1.2`

### Task 2: Implement Proper robots.txt Parser

**File**: `internal/classification/smart_website_crawler.go`

Replace the simplified `checkRobotsTxt` function (lines 822-861) with a proper implementation:

- Import `github.com/temoto/robotstxt`
- Parse robots.txt using the library
- Check both specific User-Agent and wildcard (*) rules
- Respect `Crawl-Delay` directives
- Test specific paths, not just root disallow
- Return crawl delay value if specified

**Key Changes**:

- Use `robotstxt.FromBytes()` to parse robots.txt
- Use `FindGroup()` to find rules for our User-Agent
- Use `Test()` method to check if specific paths are allowed
- Extract and return `Crawl-Delay` value
- Handle missing robots.txt gracefully (allow crawling)

### Task 3: Create Centralized User-Agent Function

**File**: `internal/classification/user_agent.go` (new file)

Create a new file with:

- Function `GetUserAgent() string` that returns identifiable User-Agent
- Format: `Mozilla/5.0 (compatible; KYBPlatformBot/1.0; +https://your-domain.com/bot-info; Business Verification)`
- Include contact information URL
- Make it configurable via environment variable if needed

**User-Agent Format**:

```
Mozilla/5.0 (compatible; KYBPlatformBot/1.0; +https://kyb-platform.com/bot-info; Business Verification)
```

### Task 4: Replace Hardcoded User-Agent Strings

**Files to Update**:

1. `internal/classification/repository/supabase_repository.go` (lines 2514, 2729)
2. `internal/classification/smart_website_crawler.go` (line 715)
3. `internal/classification/enhanced_website_scraper.go` (line 173)
4. `internal/classification/multi_method_classifier.go` (line 778)

**Action**: Replace all hardcoded User-Agent strings with calls to the centralized function.

### Task 5: Increase Rate Limiting to Conservative Values

**File**: `internal/classification/repository/supabase_repository.go`

Update rate limiting configuration:

- Change `minDelay` from `1 * time.Second` to `3 * time.Second` (lines 224, 287)
- Make it configurable via environment variable `SCRAPING_RATE_LIMIT_DELAY` (default: 3s)
- Add validation to ensure minimum 2 seconds

**Configuration**:

- Default: 3 seconds between requests
- Minimum: 2 seconds (enforced)
- Maximum: 10 seconds (configurable)

### Task 6: Integrate robots.txt Crawl-Delay with Rate Limiting

**File**: `internal/classification/repository/supabase_repository.go`

Enhance `applyRateLimit` function to:

- Accept optional `crawlDelay` parameter from robots.txt
- Use robots.txt crawl delay if specified and greater than default
- Log when using robots.txt crawl delay

**Integration Point**: After robots.txt check in `SmartWebsiteCrawler`, pass crawl delay to rate limiter.

### Task 7: Add HTTP Status Code Handling

**Files to Update**:

1. `internal/classification/repository/supabase_repository.go`

   - `extractKeywordsFromHomepageWithRetry` (around line 2542)
   - `extractKeywordsFromWebsite` (around line 2746)

2. `internal/classification/smart_website_crawler.go`

   - `CrawlWebsite` and page analysis functions

**Implementation**:

- **429 (Too Many Requests)**: 
  - Read `Retry-After` header
  - Stop immediately, do not retry
  - Log with warning
  - Return error with retry-after information
- **403 (Forbidden)**:
  - Stop immediately, do not retry
  - Log as blocked
  - Return error indicating access forbidden
- **503 (Service Unavailable)**:
  - Implement exponential backoff
  - Retry up to 3 times with increasing delays
  - Log service unavailable status

**Code Pattern**:

```go
if resp.StatusCode == 429 {
    retryAfter := resp.Header.Get("Retry-After")
    r.logger.Printf("⚠️ Rate limited (429) for %s, Retry-After: %s", domain, retryAfter)
    return []string{}, fmt.Errorf("rate limited: retry after %s", retryAfter)
}
if resp.StatusCode == 403 {
    r.logger.Printf("🚫 Access forbidden (403) for %s", domain)
    return []string{}, fmt.Errorf("access forbidden")
}
if resp.StatusCode == 503 {
    // Handle with exponential backoff
}
```

### Task 8: Create Scraping Policy Documentation

**File**: `docs/SCRAPING_POLICY.md` (new file)

Create comprehensive documentation covering:

- **Compliance Measures**: robots.txt, rate limiting, User-Agent identification
- **Legal Basis**: Public data, business verification purpose, minimal extraction
- **Rate Limiting**: 3-second minimum delay, robots.txt crawl-delay respect
- **Error Handling**: How we handle 429/403/503
- **Contact Information**: How website owners can contact us
- **Data Usage**: Keywords only, no personal data, business verification purpose
- **Opt-out Mechanism**: How sites can request exclusion

### Task 9: Update robots.txt Check in Repository

**File**: `internal/classification/repository/supabase_repository.go`

If repository performs its own robots.txt checks, update to use the improved parser from `SmartWebsiteCrawler` or create a shared utility.

### Task 10: Add Configuration for Scraping Behavior

**File**: `internal/classification/config.go` or environment variables

Add configuration options:

- `SCRAPING_RATE_LIMIT_DELAY`: Default delay between requests (default: 3s)
- `SCRAPING_RESPECT_ROBOTS`: Whether to respect robots.txt (default: true)
- `SCRAPING_USER_AGENT_CONTACT_URL`: URL for bot information page

## Testing Requirements

### Unit Tests

1. **robots.txt Parser**: Test various robots.txt formats, crawl-delay extraction, path testing
2. **User-Agent**: Verify format includes contact information
3. **Rate Limiting**: Test 3-second delay, crawl-delay integration
4. **Status Code Handling**: Test 429/403/503 responses

### Integration Tests

1. Test with real websites that have robots.txt
2. Test rate limiting behavior
3. Test status code handling with mock responses

## Dependencies

- **New**: `github.com/temoto/robotstxt v1.1.2`
- **Existing**: All current dependencies remain

## Files to Create

1. `internal/classification/user_agent.go` - Centralized User-Agent function
2. `docs/SCRAPING_POLICY.md` - Scraping policy documentation

## Files to Modify

1. `go.mod` - Add robots.txt library
2. `internal/classification/smart_website_crawler.go` - Proper robots.txt parser
3. `internal/classification/repository/supabase_repository.go` - Rate limiting, status codes, User-Agent
4. `internal/classification/enhanced_website_scraper.go` - User-Agent
5. `internal/classification/multi_method_classifier.go` - User-Agent

## Success Criteria

1. ✅ Proper robots.txt parsing with path-specific checks
2. ✅ Identifiable User-Agent with contact information
3. ✅ Conservative rate limiting (3 seconds minimum)
4. ✅ Proper handling of 429/403/503 status codes
5. ✅ Comprehensive scraping policy documentation
6. ✅ All tests passing
7. ✅ No breaking changes to existing functionality

## Risk Mitigation

- **Backward Compatibility**: Maintain existing behavior when robots.txt is unavailable
- **Performance**: Rate limiting may slow down scraping, but improves compliance
- **Testing**: Test with various robots.txt formats to ensure compatibility
- **Documentation**: Clear policy helps with legal compliance

## Estimated Effort

- **Task 1-2**: 2-3 hours (robots.txt parser)
- **Task 3-4**: 1-2 hours (User-Agent centralization)
- **Task 5-6**: 1-2 hours (Rate limiting)
- **Task 7**: 2-3 hours (Status code handling)
- **Task 8**: 1-2 hours (Documentation)
- **Task 9-10**: 1 hour (Configuration)
- **Testing**: 2-3 hours

**Total**: ~10-15 hours

### To-dos

- [ ] Add github.com/temoto/robotstxt library to go.mod
- [ ] Replace simplified robots.txt parser in smart_website_crawler.go with proper implementation using robotstxt library
- [ ] Create centralized user_agent.go file with identifiable User-Agent function including contact information
- [ ] Replace all hardcoded User-Agent strings in 5 files with calls to centralized function
- [ ] Increase rate limiting from 1s to 3s minimum delay, make configurable via environment variable
- [ ] Integrate robots.txt Crawl-Delay directive with rate limiting system
- [ ] Add specific handling for 429 (stop immediately), 403 (stop immediately), 503 (exponential backoff) in all scraping functions
- [ ] Create comprehensive SCRAPING_POLICY.md documentation covering compliance measures, legal basis, and contact information
- [ ] Update any robots.txt checks in repository to use improved parser or shared utility
- [ ] Add configuration options for scraping behavior (rate limit delay, respect robots, contact URL)
- [ ] Write unit tests for robots.txt parser, User-Agent format, rate limiting, and status code handling
- [ ] Write integration tests with real websites and mock HTTP responses for status codes