<!-- Keyword Extraction and Classification Accuracy Improvement Plan -->

# Keyword Extraction and Classification Accuracy Improvement Plan

## Overview

This plan addresses critical issues in the keyword extraction pipeline that are causing poor classification accuracy. The current implementation only extracts 1 keyword ("grape") from "The Greene Grape" website, leading to incorrect industry classification ("Wineries" with no codes) instead of a more accurate classification with proper industry codes.

## Problem Analysis

### Current Issues

1. **Smart Crawler Keyword Extraction Not Implemented**

   - `extractPageKeywords()` method is a placeholder returning empty array
   - Even when pages are successfully fetched, no keywords are extracted from HTML content
   - Location: `internal/classification/smart_website_crawler.go:782-786`

2. **DNS Resolution Failure in Railway**

   - DNS lookup fails: `lookup www.thegreenegrape.com on [fd12::10]:53: no such host`
   - All page requests fail with `Status=0`, preventing content analysis
   - Custom DNS resolver exists but may not be properly configured for all HTTP clients

3. **URL Text Extraction Too Simplistic**

   - Only extracts single word "grape" from domain name "thegreenegrape.com"
   - Doesn't extract compound keywords like "green grape", "wine", "beverage", "retail"
   - Missing domain name parsing logic for multi-word domains

4. **Missing Industry-Specific Keywords**

   - No keywords extracted for wine/retail industry (wine, beverage, alcohol, retail, shop, store, etc.)
   - Industry detection relies on single keyword match instead of comprehensive keyword analysis

5. **No Fallback for Failed Page Analysis**

   - When all pages fail, system falls back to URL-only extraction
   - No intermediate fallbacks (e.g., try homepage only, try with different DNS, retry logic)

## Implementation Plan

### Phase 1: Implement Smart Crawler Keyword Extraction

**Priority**: Critical

**Estimated Time**: 4-6 hours

#### 1.1 Implement `extractPageKeywords()` Method

**File**: `internal/classification/smart_website_crawler.go`

**Implementation**:

- Extract text from HTML content (remove scripts, styles, tags)
- Use business keyword patterns from `extractBusinessKeywords()` in repository
- Extract keywords from:
  - Page title
  - Meta description
  - Heading tags (h1-h6)
  - Body text (weighted by position)
  - Structured data (JSON-LD, microdata)
  - Alt text from images
- Apply stop word filtering
- Extract 2-word and 3-word phrases
- Return top 30 keywords sorted by relevance

**Code Structure**:

```go
func (c *SmartWebsiteCrawler) extractPageKeywords(content string, pageType string) []string {
    // 1. Extract clean text from HTML
    textContent := c.extractTextFromHTML(content)

    // 2. Extract from structured elements (title, meta, headings)
    structuredKeywords := c.extractStructuredKeywords(content)

    // 3. Extract from body text using business patterns
    bodyKeywords := c.extractBusinessKeywordsFromText(textContent)

    // 4. Extract phrases (2-word, 3-word)
    phrases := c.extractPhrases(textContent, 2, 3)

    // 5. Combine, deduplicate, and rank
    allKeywords := c.combineAndRankKeywords(structuredKeywords, bodyKeywords, phrases, pageType)

    // 6. Return top 30
    return c.limitToTopKeywords(allKeywords, 30)
}
```

#### 1.2 Implement `extractIndustryIndicators()` Method

**Implementation**:

- Use industry-specific patterns from `content_relevance_analyzer.go`
- Extract industry signals from:
  - Page content
  - URL structure
  - Meta tags
  - Structured data
- Return industry indicators with confidence scores

#### 1.3 Implement Supporting Methods

- `extractTextFromHTML()` - Clean HTML to text
- `extractStructuredKeywords()` - Extract from title, meta, headings
- `extractBusinessKeywordsFromText()` - Use regex patterns
- `extractPhrases()` - Extract multi-word phrases
- `combineAndRankKeywords()` - Merge and rank keywords
- `limitToTopKeywords()` - Return top N keywords

### Phase 2: Fix DNS Resolution and HTTP Client Configuration

**Priority**: Critical

**Estimated Time**: 2-3 hours

#### 2.1 Verify DNS Resolver in Smart Crawler

**File**: `internal/classification/smart_website_crawler.go`

**Current State**: DNS resolver is implemented in `NewSmartWebsiteCrawler()` but may not be working correctly

**Fixes**:

- Verify custom DNS resolver is being used
- Add fallback DNS servers (8.8.8.8, 1.1.1.1, 8.8.4.4)
- Add retry logic with exponential backoff
- Add DNS resolution timeout handling
- Log DNS resolution attempts for debugging

**Implementation**:

```go
// Enhanced DNS resolver with multiple fallback servers
dnsServers := []string{"8.8.8.8:53", "1.1.1.1:53", "8.8.4.4:53"}
dnsResolver := &net.Resolver{
    PreferGo: true,
    Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
        // Try each DNS server with retry logic
        for _, server := range dnsServers {
            d := net.Dialer{Timeout: 5 * time.Second}
            conn, err := d.DialContext(ctx, "udp4", server)
            if err == nil {
                return conn, nil
            }
        }
        return nil, fmt.Errorf("all DNS servers failed")
    },
}
```

#### 2.2 Add Retry Logic for Failed Requests

**Implementation**:

- Retry failed HTTP requests up to 3 times
- Exponential backoff between retries (1s, 2s, 4s)
- Different retry strategies for DNS vs HTTP errors
- Log retry attempts for observability

#### 2.3 Improve Error Handling

**Implementation**:

- Distinguish between DNS errors, network errors, and HTTP errors
- Provide specific error messages for each type
- Fallback strategies based on error type
- Continue processing other pages even if some fail

### Phase 3: Enhance URL Text Extraction

**Priority**: High

**Estimated Time**: 2-3 hours

#### 3.1 Improve Domain Name Parsing

**File**: `internal/classification/repository/supabase_repository.go`

**Current**: Only extracts single words from domain

**Enhancement**:

- Parse compound domain names (e.g., "thegreenegrape" → ["green", "grape", "green grape"])
- Extract meaningful subdomain keywords
- Use domain name analysis to infer industry
- Extract TLD-based hints (.shop, .store, .restaurant, etc.)

**Implementation**:

```go
func (r *SupabaseKeywordRepository) extractKeywordsFromURLEnhanced(websiteURL string) []ContextualKeyword {
    // 1. Parse domain name
    domain := extractDomain(websiteURL)
    domainParts := splitDomainName(domain) // ["the", "green", "grape"]

    // 2. Extract individual words (filter stop words)
    words := filterStopWords(domainParts)

    // 3. Extract 2-word phrases
    phrases := generatePhrases(domainParts, 2)

    // 4. Extract 3-word phrases for longer domains
    if len(domainParts) > 3 {
        phrases = append(phrases, generatePhrases(domainParts, 3)...)
    }

    // 5. Add TLD-based hints
    tldKeywords := extractTLDHints(websiteURL)

    // 6. Add industry inference from domain
    industryKeywords := inferIndustryFromDomain(domain)

    // 7. Combine and return
    return combineKeywords(words, phrases, tldKeywords, industryKeywords)
}
```

#### 3.2 Add Domain Name Industry Inference

**Implementation**:

- Create domain name → industry mapping patterns
- Examples:
  - "wine", "grape", "vineyard", "vintner" → Food & Beverage
  - "shop", "store", "market", "retail" → Retail
  - "tech", "software", "app", "digital" → Technology
- Use this to suggest additional keywords

#### 3.3 Enhance Phrase Extraction

**Implementation**:

- Extract meaningful 2-word phrases from domain
- Filter out common stop word combinations
- Prioritize business-relevant phrases
- Examples: "thegreenegrape" → ["green", "grape", "green grape", "wine", "retail"]

### Phase 4: Improve Keyword Extraction Patterns

**Priority**: High

**Estimated Time**: 3-4 hours

#### 4.1 Expand Business Keyword Patterns

**File**: `internal/classification/repository/supabase_repository.go`

**Current**: Basic patterns exist but may miss industry-specific terms

**Enhancement**:

- Add more comprehensive patterns for each industry
- Include synonyms and related terms
- Add NAICS-aligned industry keywords
- Include common business terms and phrases

**New Patterns to Add**:

```go
// Food & Beverage (expanded)
`\b(wine|wines|winery|vineyard|vintner|sommelier|tasting|cellar|bottle|vintage|grape|grapes|grapevine|oenology|wine shop|wine store|wine bar|wine merchant|wine retailer|alcohol|spirits|liquor|beer|brewery|distillery|beverage|beverages)\b`,

// Retail (expanded)
`\b(retail|retailer|retail store|retail shop|brick and mortar|brick-and-mortar|physical store|storefront|merchandise|inventory|point of sale|POS|checkout|cash register|sales floor|showroom|boutique|outlet|marketplace|vendor|seller|selling|commerce)\b`,

// E-commerce (new)
`\b(ecommerce|e-commerce|online store|online shop|digital storefront|web store|internet retailer|online marketplace|digital commerce|online sales|web sales|internet sales|online retail)\b`,
```

#### 4.2 Add Context-Aware Keyword Extraction

**Implementation**:

- Weight keywords by their position in page (title > headings > body)
- Weight keywords by page type (about > services > products > contact)
- Consider keyword frequency and co-occurrence
- Boost keywords that appear in multiple pages

#### 4.3 Implement Keyword Relevance Scoring

**Implementation**:

- Score keywords based on:
  - Position in page (title=1.0, h1=0.9, h2=0.8, body=0.6)
  - Page type relevance (about=1.0, services=0.95, products=0.9)
  - Frequency in content
  - Industry pattern match strength
- Filter keywords below relevance threshold (0.3)
- Rank and return top keywords

### Phase 5: Improve Fallback Mechanisms

**Priority**: Medium

**Estimated Time**: 2-3 hours

#### 5.1 Implement Multi-Level Fallback Chain

**Current**: Multi-page → Single-page → URL extraction

**Enhanced**:

1. Multi-page analysis (15 pages)
2. Single-page analysis (homepage only)
3. Homepage with retry (different DNS, longer timeout)
4. URL text extraction (enhanced)
5. Business name analysis (if brand match)
6. Default to "General Business" with low confidence

**Implementation**:

```go
func (r *SupabaseKeywordRepository) extractKeywords(businessName, websiteURL string) []ContextualKeyword {
    // Level 1: Multi-page analysis
    keywords := r.extractKeywordsFromMultiPageWebsite(ctx, websiteURL)
    if len(keywords) >= 5 {
        return keywords
    }

    // Level 2: Single-page analysis (homepage)
    keywords = r.extractKeywordsFromSinglePage(ctx, websiteURL)
    if len(keywords) >= 3 {
        return keywords
    }

    // Level 3: Homepage with enhanced retry
    keywords = r.extractKeywordsFromHomepageWithRetry(ctx, websiteURL)
    if len(keywords) >= 2 {
        return keywords
    }

    // Level 4: Enhanced URL text extraction
    keywords = r.extractKeywordsFromURLEnhanced(websiteURL)
    if len(keywords) >= 1 {
        return keywords
    }

    // Level 5: Business name (if brand match)
    if isBrandMatch(businessName) {
        keywords = r.extractKeywordsFromBusinessName(businessName)
    }

    return keywords
}
```

#### 5.2 Add Partial Success Handling

**Implementation**:

- If some pages succeed but not enough (e.g., 1-2 pages), use those keywords
- Lower confidence score for partial results
- Log partial success for monitoring

#### 5.3 Implement Graceful Degradation

**Implementation**:

- Continue processing even if some steps fail
- Use available keywords even if below ideal threshold
- Provide confidence scores that reflect data quality

### Phase 6: Enhance Structured Data Extraction

**Priority**: Medium

**Estimated Time**: 3-4 hours

#### 6.1 Implement JSON-LD Extraction

**File**: `internal/classification/smart_website_crawler.go`

**Implementation**:

- Extract JSON-LD structured data from pages
- Parse Schema.org types (LocalBusiness, Restaurant, Store, etc.)
- Extract business information:
  - Business name
  - Description
  - Industry/industry code
  - Services/products
  - Address
  - Contact information
- Use structured data to boost keyword confidence

#### 6.2 Implement Microdata Extraction

**Implementation**:

- Extract HTML5 microdata attributes
- Parse itemscope, itemtype, itemprop
- Extract business-relevant properties
- Convert to keywords with high confidence

#### 6.3 Implement Open Graph and Twitter Card Extraction

**Implementation**:

- Extract og:title, og:description, og:type
- Extract twitter:title, twitter:description
- Use for keyword extraction and industry hints

### Phase 7: Improve Industry Detection Accuracy

**Priority**: High

**Estimated Time**: 4-5 hours

#### 7.1 Multi-Keyword Industry Matching

**Current**: Single keyword match determines industry

**Enhancement**:

- Require multiple keyword matches for high confidence
- Use keyword co-occurrence patterns
- Weight keywords by industry relevance
- Consider keyword frequency across pages

**Implementation**:

```go
func (s *IndustryDetectionService) classifyByKeywords(ctx context.Context, keywords []string) (*IndustryDetectionResult, error) {
    // 1. Get industry matches for each keyword
    industryScores := make(map[string]float64)

    for _, keyword := range keywords {
        industries := s.getIndustriesForKeyword(ctx, keyword)
        for industry, score := range industries {
            industryScores[industry] += score
        }
    }

    // 2. Normalize scores by keyword count
    for industry := range industryScores {
        industryScores[industry] /= float64(len(keywords))
    }

    // 3. Require minimum keyword matches (e.g., 3+ keywords)
    // 4. Require minimum confidence threshold (e.g., 0.6)
    // 5. Return top industry with confidence score
}
```

#### 7.2 Add Industry Confidence Thresholds

**Implementation**:

- Set minimum keyword count per industry (e.g., 3 keywords)
- Set minimum confidence threshold (e.g., 0.6)
- If below threshold, use "General Business" with lower confidence
- Log when thresholds aren't met for monitoring

#### 7.3 Implement Industry Co-Occurrence Analysis

**Implementation**:

- Analyze which industries commonly appear together
- Use co-occurrence to boost confidence
- Example: "wine" + "retail" + "shop" → Retail (Food & Beverage) with high confidence

### Phase 8: Add Comprehensive Logging and Observability

**Priority**: Medium

**Estimated Time**: 2-3 hours

#### 8.1 Add Detailed Keyword Extraction Logging

**Implementation**:

- Log each step of keyword extraction
- Log keywords found at each stage
- Log confidence scores and reasoning
- Log fallback triggers and reasons

**Log Format**:

```
[KeywordExtraction] Starting extraction for: https://example.com
[KeywordExtraction] Multi-page analysis: 15 pages discovered
[KeywordExtraction] Page 1/15: Extracted 12 keywords (wine, grape, retail, shop, ...)
[KeywordExtraction] Page 2/15: Extracted 8 keywords (beverage, alcohol, ...)
[KeywordExtraction] Total keywords extracted: 45 unique keywords
[KeywordExtraction] Top keywords: wine(0.95), retail(0.90), shop(0.85), grape(0.80), ...
```

#### 8.2 Add Performance Metrics

**Implementation**:

- Track keyword extraction time per method
- Track success rates for each fallback level
- Track average keywords extracted per page
- Track DNS resolution success rate
- Export metrics to observability system

#### 8.3 Add Error Tracking

**Implementation**:

- Track DNS resolution failures
- Track HTTP request failures
- Track keyword extraction failures
- Categorize errors for analysis
- Alert on high error rates

### Phase 9: Performance Optimizations

**Priority**: Low

**Estimated Time**: 2-3 hours

#### 9.1 Optimize Keyword Extraction Performance

**Implementation**:

- Cache compiled regex patterns
- Use concurrent processing for multiple pages
- Limit content size for processing (e.g., first 50KB)
- Use streaming HTML parsing for large pages

#### 9.2 Optimize DNS Resolution

**Implementation**:

- Cache DNS resolutions (TTL-based)
- Use connection pooling
- Parallel DNS lookups for multiple domains
- Timeout optimization

#### 9.3 Add Request Rate Limiting

**Implementation**:

- Respect robots.txt crawl-delay
- Add configurable rate limits
- Implement exponential backoff
- Add jitter to avoid thundering herd

### Phase 10: Testing and Validation

**Priority**: Critical

**Estimated Time**: 4-6 hours

#### 10.1 Unit Tests

**Files to Test**:

- `internal/classification/smart_website_crawler.go`
  - `extractPageKeywords()`
  - `extractIndustryIndicators()`
  - `extractTextFromHTML()`
  - `extractStructuredKeywords()`
- `internal/classification/repository/supabase_repository.go`
  - `extractKeywordsFromURLEnhanced()`
  - `extractKeywordsFromMultiPageWebsite()`
  - `extractBusinessKeywords()`

**Test Cases**:

- Extract keywords from HTML with various structures
- Handle missing/empty content gracefully
- Extract keywords from structured data
- Extract keywords from domain names
- Handle DNS failures gracefully
- Test fallback chain

#### 10.2 Integration Tests

**Test Scenarios**:

- Full keyword extraction flow for real websites
- Multi-page analysis with successful pages
- Multi-page analysis with failed pages
- URL-only fallback scenario
- Business name fallback scenario
- End-to-end classification with improved keywords

#### 10.3 Performance Tests

**Test Scenarios**:

- Keyword extraction time for 15 pages
- DNS resolution time with retries
- Memory usage during keyword extraction
- Concurrent request handling

#### 10.4 Accuracy Tests

**Test Scenarios**:

- Test with known business websites
- Verify keyword extraction accuracy
- Verify industry classification accuracy
- Compare before/after improvements

## Implementation Order

### Week 1: Critical Fixes

1. **Day 1-2**: Phase 1 - Implement Smart Crawler Keyword Extraction
2. **Day 3**: Phase 2 - Fix DNS Resolution
3. **Day 4**: Phase 3 - Enhance URL Text Extraction
4. **Day 5**: Phase 10.1 - Unit Tests for Phases 1-3

### Week 2: Enhancements

1. **Day 1-2**: Phase 4 - Improve Keyword Extraction Patterns
2. **Day 3**: Phase 5 - Improve Fallback Mechanisms
3. **Day 4**: Phase 7 - Improve Industry Detection Accuracy
4. **Day 5**: Phase 10.2 - Integration Tests

### Week 3: Polish and Optimization

1. **Day 1-2**: Phase 6 - Enhance Structured Data Extraction
2. **Day 3**: Phase 8 - Add Comprehensive Logging
3. **Day 4**: Phase 9 - Performance Optimizations
4. **Day 5**: Phase 10.3-10.4 - Performance and Accuracy Tests

## Success Criteria

### Keyword Extraction

- ✅ Extract 10+ keywords from multi-page websites (currently 1)
- ✅ Extract 5+ keywords from single-page websites
- ✅ Extract 3+ keywords from URL-only fallback
- ✅ Keywords include industry-relevant terms

### Industry Classification

- ✅ Correctly classify businesses with 80%+ confidence
- ✅ Generate classification codes for detected industries
- ✅ Reduce "General Business" fallback rate by 50%

### Performance

- ✅ Keyword extraction completes in < 5 seconds for 15 pages
- ✅ DNS resolution succeeds 95%+ of the time
- ✅ Fallback chain completes in < 2 seconds

### Observability

- ✅ All keyword extraction steps are logged
- ✅ Performance metrics are tracked
- ✅ Error rates are monitored and alerted

## Risk Mitigation

### Risks

1. **DNS resolution still fails in Railway**

   - Mitigation: Multiple DNS servers, retry logic, fallback to URL extraction

2. **Keyword extraction too slow**

   - Mitigation: Limit content size, use concurrent processing, optimize regex

3. **Too many keywords extracted (noise)**

   - Mitigation: Relevance scoring, filtering, top N keywords only

4. **Breaking changes to existing functionality**

   - Mitigation: Comprehensive testing, gradual rollout, feature flags

## Rollout Strategy

### Phase 1: Internal Testing (Week 1)

- Deploy to staging environment
- Test with known business websites
- Monitor logs and metrics
- Fix critical issues

### Phase 2: Gradual Rollout (Week 2)

- Enable for 10% of requests (feature flag)
- Monitor accuracy and performance
- Gradually increase to 50%, then 100%
- Rollback if issues detected

### Phase 3: Production (Week 3)

- Full deployment
- Monitor for 1 week
- Collect accuracy metrics
- Iterate based on results

## Dependencies

### External

- None (all improvements are internal)

### Internal

- Existing keyword extraction infrastructure
- Smart crawler infrastructure
- Industry detection service
- Classification code generation

## Documentation Updates

### Code Documentation

- Document new keyword extraction methods
- Document fallback chain logic
- Document DNS resolution configuration
- Document performance considerations

### User Documentation

- Update API documentation if response format changes
- Document new metadata fields
- Document confidence score interpretation

### Operations Documentation

- Document DNS configuration requirements
- Document monitoring and alerting
- Document troubleshooting guide

## Future Enhancements (Out of Scope)

1. **Machine Learning for Keyword Extraction**

   - Train ML model on labeled business websites
   - Improve keyword relevance scoring
   - Better industry classification

2. **External Data Sources**

   - Integrate with business directories (Yelp, Google Business)
   - Use social media profiles for keywords
   - Use public records for industry verification

3. **Advanced NLP**

   - Named entity recognition for business information
   - Sentiment analysis for industry signals
   - Topic modeling for industry classification

4. **Caching and Optimization**

   - Cache website analysis results
   - Pre-compute keywords for common domains
   - CDN for static keyword patterns

5. **Word Segmentation Library for Compound Domain Names**

   - Integrate a word segmentation library (e.g., `github.com/kljensen/snowball` or custom dictionary-based approach)
   - Improve compound word splitting for domain names like "thegreenegrape" → ["green", "grape"]
   - Use dictionary-based word segmentation to identify word boundaries in compound domains
   - Support multiple languages for international domain names
   - Fallback to current heuristic approach if segmentation fails
   - Performance considerations: cache segmentation results for common domain patterns
   - Implementation location: `internal/classification/repository/supabase_repository.go:splitDomainName()`
   - Current limitation: Only splits on hyphens, underscores, and camelCase; compound words without separators remain unsplit
   - Expected improvement: Extract 2-3x more keywords from compound domain names

## Appendix

### A. Keyword Patterns Reference

See existing patterns in:

- `internal/classification/repository/supabase_repository.go:2449-2466`
- `internal/classification/multi_method_classifier.go:878-895`
- `internal/classification/enhanced_website_scraper.go:320-342`

### B. DNS Configuration Reference

Current DNS resolver implementation:

- `internal/classification/smart_website_crawler.go:130-176`
- `internal/classification/repository/supabase_repository.go:1962-2008`

### C. Testing Data

Test websites to use:

- Wine retailer: https://www.thegreenegrape.com
- Technology company: (to be identified)
- Retail store: (to be identified)
- Restaurant: (to be identified)

### D. Performance Benchmarks

Current performance:

- Multi-page analysis: ~110ms (but fails)
- Single-page analysis: ~200ms (but fails)
- URL extraction: < 1ms (but only 1 keyword)

Target performance:

- Multi-page analysis: < 5 seconds
- Single-page analysis: < 2 seconds
- URL extraction: < 10ms (with 3+ keywords)
