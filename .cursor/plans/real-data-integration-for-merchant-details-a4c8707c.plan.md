---
name: "Implementation Plan: Fix Classification Priority & Enhance Multi-Page Website Analysis"
overview: ""
todos:
  - id: 6fd9e87f-94d3-4bc1-8cba-3dce42bdf4d0
    content: Verify merchant_analytics table schema supports classification_data with status tracking
    status: pending
  - id: 6cd1db46-32c8-4f00-b901-899539ad7904
    content: Create ClassificationJob struct and Process() method that calls classification service and saves to Supabase
    status: pending
  - id: 853ab475-ecdb-40a6-baf9-1ac593a8982a
    content: Create job processor with worker pool for async classification job processing
    status: pending
  - id: 87249547-a287-449f-bf97-1eab133e416d
    content: Integrate classification job trigger into createMerchant() function (async, non-blocking)
    status: pending
  - id: 2d24a85f-5af1-41ee-a531-544465e71d39
    content: Update HandleMerchantSpecificAnalytics to read from merchant_analytics table instead of hardcoded values
    status: pending
  - id: 0eb7256e-a58f-4137-9215-2e177ccdf518
    content: Create GET /api/v1/merchants/{id}/analytics/status endpoint for processing status
    status: pending
  - id: 2bedb026-181b-4392-9cf3-991dcc8833cf
    content: Add website_analysis_data and website_analysis_status columns to merchant_analytics table
    status: pending
  - id: ed812263-6542-4942-884e-e0352a7e3890
    content: Create WebsiteAnalysisJob struct and Process() method that calls website analysis service
    status: pending
  - id: 0259f418-7a11-4b86-ad2b-4358babb49cc
    content: Add conditional website analysis job trigger in createMerchant() (only when website URL provided)
    status: pending
  - id: c77b5562-d9f1-4a91-9471-0f0207d490f5
    content: Update HandleMerchantWebsiteAnalysis to read from merchant_analytics table with status checking
    status: pending
  - id: b023775f-3f21-4ec8-a098-7d0c1ab65b2c
    content: Update HandleMerchantRiskScore to read from risk_assessments table instead of hardcoded mapping
    status: pending
  - id: e9c6f494-a655-4e95-9a4a-9fa54bb9e461
    content: Update HandleMerchantStatistics to query real data from Supabase merchants and risk_assessments tables
    status: pending
  - id: 23b6c871-2f74-4675-97b6-652bfcdc0d39
    content: Create AnalyticsStatusIndicator React component for showing processing status in UI
    status: pending
  - id: 9e2204a2-d009-4154-8879-e6399be4c839
    content: Add status indicators to BusinessAnalyticsTab component
    status: pending
  - id: 55406f6d-b7e5-401c-9f75-d770fa48c3bf
    content: Add getMerchantAnalyticsStatus() function to frontend API client
    status: pending
  - id: 404968ed-861f-483d-bf1e-bde059a3b75d
    content: Initialize job processor in main.go with worker pool and graceful shutdown
    status: pending
---

# Implementation Plan: Fix Classification Priority & Enhance Multi-Page Website Analysis

## Overview

This plan fixes the classification priority issue (website content should be primary, business name secondary) and enhances website analysis to analyze multiple pages with relevance-based weighting and structured data extraction. The implementation includes backend changes, Supabase schema updates, frontend enhancements, and comprehensive testing.

**Deployment Strategy**: Big bang deployment (all features enabled simultaneously in production). No staging environment - all testing must be comprehensive before deployment.

---

## Phase 1: Fix Priority Weights in Keyword Extraction

### 1.1 Reverse Priority in Keyword Extraction

**File**: `internal/classification/repository/supabase_repository.go`

**Location**: Lines 1557-1606 (`extractKeywords()` function)

**Changes**:

- Extract website content keywords FIRST (highest priority)
- Extract business name keywords LAST (only for high-confidence brand matches in MCC 3000-3831)
- Add brand matching check before extracting business name keywords

**Implementation Details**:

- Move website scraping to top of function
- Add `isHighConfidenceBrandMatch()` check for business name (only for MCC 3000-3831)
- Update logging to reflect new priority order

### 1.2 Create Brand Matcher

**New File**: `internal/classification/repository/brand_matcher.go`

**Purpose**: Check if business name matches known hotel brands for MCC codes 3000-3831 (Hotels, Motels, and Resorts)

**Implementation**:

- Create `BrandMatcher` struct with hardcoded hotel brands map
- Implement `isHighConfidenceBrandMatch()` method that only matches brands in MCC range 3000-3831
- Hardcoded known hotel brands (initial list):
  - Hilton (Hilton Hotels, Hilton Garden Inn, DoubleTree, Hampton, Embassy Suites, etc.)
  - Marriott (Marriott Hotels, Courtyard, Residence Inn, Fairfield, SpringHill, etc.)
  - Hyatt (Hyatt Hotels, Hyatt Regency, Grand Hyatt, Park Hyatt, etc.)
  - IHG (InterContinental, Holiday Inn, Crowne Plaza, Kimpton, etc.)
  - Accor (Novotel, Ibis, Sofitel, Mercure, etc.)
  - Wyndham (Wyndham Hotels, Ramada, Days Inn, Super 8, etc.)
- Normalize business names (remove Inc, LLC, Corp, Hotels, Hotel, Resorts, Resort suffixes)
- Only extract business name keywords if brand match AND classification would be in MCC 3000-3831 range
- Store brand match confidence and MCC code in metadata

### 1.3 Update Enhanced Scoring Algorithm Weights

**File**: `internal/classification/repository/enhanced_scoring_algorithm.go`

**Location**: Lines 60-91 (`DefaultEnhancedScoringConfig()`)

**Changes**:

- `WebsiteContentWeight: 2.0` (100% boost - highest priority)
- `BusinessNameWeight: 0.5` (50% reduction - low priority, only for brand matches in MCC 3000-3831)
- `WebsiteURLWeight: 1.0` (baseline - fallback when scraping fails)

### 1.4 Update Context Multiplier Logic

**File**: `internal/classification/repository/enhanced_scoring_algorithm.go`

**Location**: Lines 796-807 (`getContextMultiplier()`)

**Changes**: Update switch statement with specific multiplier values:

- `website_content`: 2.0 (100% boost - highest priority)
- `business_name`: 0.5 (50% reduction - only for brand matches in MCC 3000-3831)
- `website_url`: 1.0 (baseline - fallback when scraping fails)
- `description`: 1.0 (no change)

---

## Phase 2: Enhance Multi-Page Website Analysis

### 2.1 Enhance Page Priority Calculation

**File**: `internal/classification/smart_website_crawler.go`

**Location**: Lines 411-461 (`calculatePagePriority()`)

**Changes**:

- Add "sale", "sales" to highest priority patterns (weight: 90-100)
- Increase weights for industry-revealing pages (about, products, services, sale)
- Add structured data presence as priority boost

**Priority Weights**:

- Highest (90-100): about, products, services, sale, sales
- High (70-80): contact, team, careers, locations
- Medium (50-60): blog, news, case-studies, portfolio
- Low (30-40): support, help, faq, privacy, terms

### 2.2 Enhance Page Relevance Scoring

**File**: `internal/classification/smart_website_crawler.go`

**Location**: Lines 716-730 (`calculateRelevanceScore()`)

**Changes**:

- Increase base scores for industry-revealing pages (about, services, products, sale: 0.95)
- Add 10% boost if structured data is present
- Add 5% boost for high content quality (>0.7)
- Reduce score by 20% for low content length (<500 chars)

### 2.3 Integrate Structured Data Extraction

**File**: `internal/classification/repository/supabase_repository.go`

**Location**: `extractKeywordsFromWebsite()` function (lines 1608-1717)

**Changes**:

- Import and use existing `StructuredDataExtractor` from `internal/classification/structured_data_extractor.go`
- Call `ExtractStructuredData()` method on HTML content
- Extract keywords from:
  - `SchemaOrgData` (Organization, LocalBusiness types)
  - `ProductInfo` (product names, categories)
  - `ServiceInfo` (service names, categories)
  - `BusinessInfo` (industry, business type)
- Weight structured data keywords 1.5x higher than text keywords (higher confidence)
- Combine structured data keywords with text keywords
- Increase keyword limit from 20 to 30 to account for structured data

### 2.4 Implement Multi-Page Analysis

**File**: `internal/classification/repository/supabase_repository.go`

**New Method**: `extractKeywordsFromMultiPageWebsite()`

**Purpose**: Analyze multiple pages with relevance-based weighting

**Performance Targets**:
- Overall timeout: 60 seconds (must complete in <60s for 95% of cases)
- Per-page timeout: 15 seconds
- Circuit breaker: Skip remaining pages if 3+ consecutive failures
- Fallback trigger: Fallback to single-page if < 3 pages successfully analyzed

**Implementation**:

- Reuse existing `SmartWebsiteCrawler` infrastructure (CrawlWebsite, discoverSiteStructure, analyzePages)
- Use `SmartWebsiteCrawler.CrawlWebsite()` to discover pages (sitemap, internal links, common patterns)
- Limit to top 15 pages by priority to avoid timeout
- Analyze each page concurrently (max 5 concurrent requests with semaphore)
- Extract keywords from each page (text + structured data)
- Weight keywords by page relevance score
- Normalize keyword scores by total relevance
- Return top 30 keywords by weighted score
- Log analysis method (multi_page, single_page, url_only) for monitoring

**Page Analysis Flow**:

1. Discover pages using SmartWebsiteCrawler.CrawlWebsite()
2. Sort by priority (highest first)
3. Analyze top 15 pages with 60s overall timeout
4. Extract keywords from each page (text + structured data)
5. Weight keywords by page relevance score
6. Aggregate and normalize scores
7. Return top 30 keywords
8. Fallback to single-page if < 3 pages successfully analyzed

### 2.5 Update extractKeywords() to Use Multi-Page Analysis

**File**: `internal/classification/repository/supabase_repository.go`

**Location**: Lines 1576-1603

**Changes**:

- Replace single-page scraping with `extractKeywordsFromMultiPageWebsite()`
- Add fallback chain: multi-page → single-page → URL text extraction
- Update logging to indicate multi-page vs single-page analysis
- Fallback to single-page if < 3 pages successfully analyzed
- Fallback to URL extraction if single-page fails

---

## Phase 3: Supabase Schema Updates

### 3.1 Add Page Analysis Metadata

**New Migration**: `supabase-migrations/020_add_page_analysis_metadata.sql`

**Purpose**: Store metadata about which pages were analyzed and their relevance scores

**Changes**:

```sql
-- Add page analysis metadata to classification_data JSONB
-- Structure:
-- {
--   "pageAnalysis": {
--     "pagesAnalyzed": [
--       {
--         "url": "https://example.com/about",
--         "pageType": "about",
--         "relevanceScore": 0.95,
--         "keywordsExtracted": 15,
--         "hasStructuredData": true
--       }
--     ],
--     "totalPagesAnalyzed": 12,
--     "analysisMethod": "multi_page",
--     "structuredDataFound": true
--   }
-- }
```

**Note**: This is metadata stored within existing `classification_data` JSONB column, no schema change needed.

### 3.2 Add Brand Match Indicator

**Update**: Store brand match information in classification_data metadata

**Structure**:

```json
{
  "metadata": {
    "brandMatch": {
      "isBrandMatch": true,
      "brandName": "Hilton",
      "confidence": 0.95,
      "mccRange": "3000-3831"
    },
    "dataSourcePriority": {
      "websiteContent": "primary",
      "businessName": "secondary",
      "websiteURL": "fallback"
    }
  }
}
```

---

## Phase 4: Frontend Enhancements

### 4.1 Add Classification Metadata Display

**File**: `frontend/components/merchant/BusinessAnalyticsTab.tsx`

**Location**: Classification card section (lines 280-312)

**Changes**:

- Add "Data Source" indicator showing primary source (website content vs business name)
- Add "Pages Analyzed" count if multi-page analysis was used
- Add "Structured Data" indicator if schema.org data was found
- Add "Brand Match" badge if business name matched known brand (MCC 3000-3831 only)

**New UI Elements**:

- Info badge: "Analyzed 12 pages" (if multi-page)
- Badge: "Structured Data Found" (if schema.org detected)
- Badge: "Brand Match: Hilton" (if brand matched - MCC 3000-3831 only)
- Indicator: "Primary Source: Website Content" vs "Business Name"

### 4.2 Update Classification Data Type

**File**: `frontend/types/merchant.ts`

**Location**: `ClassificationData` interface (lines 54-61)

**Changes**: Add optional metadata fields

```typescript
export interface ClassificationData {
  primaryIndustry: string;
  confidenceScore: number;
  riskLevel: string;
  mccCodes?: IndustryCode[];
  sicCodes?: IndustryCode[];
  naicsCodes?: IndustryCode[];
  // New fields
  metadata?: {
    pageAnalysis?: {
      pagesAnalyzed: number;
      analysisMethod: 'multi_page' | 'single_page' | 'url_only';
      structuredDataFound: boolean;
    };
    brandMatch?: {
      isBrandMatch: boolean;
      brandName?: string;
      confidence?: number;
      mccRange?: string;
    };
    dataSourcePriority?: {
      websiteContent: 'primary' | 'secondary' | 'none';
      businessName: 'primary' | 'secondary' | 'none';
    };
  };
}
```

### 4.3 Update API Validation Schema

**File**: `frontend/lib/api-validation.ts`

**Location**: `ClassificationDataSchema` (lines 56-63)

**Changes**: Add optional metadata validation

```typescript
export const ClassificationDataSchema = z.object({
  primaryIndustry: z.string(),
  confidenceScore: z.number(),
  riskLevel: z.string(),
  mccCodes: z.array(IndustryCodeSchema).optional(),
  sicCodes: z.array(IndustryCodeSchema).optional(),
  naicsCodes: z.array(IndustryCodeSchema).optional(),
  metadata: z.object({
    pageAnalysis: z.object({
      pagesAnalyzed: z.number().optional(),
      analysisMethod: z.enum(['multi_page', 'single_page', 'url_only']).optional(),
      structuredDataFound: z.boolean().optional(),
    }).optional(),
    brandMatch: z.object({
      isBrandMatch: z.boolean(),
      brandName: z.string().optional(),
      confidence: z.number().optional(),
      mccRange: z.string().optional(),
    }).optional(),
  }).optional(),
});
```

### 4.4 Add Classification Metadata Component

**New File**: `frontend/components/merchant/ClassificationMetadata.tsx`

**Purpose**: Display classification metadata (pages analyzed, data sources, brand match)

**Props**:

- `metadata`: Classification metadata object
- `compact`: Boolean for compact display

**Features**:

- Show pages analyzed count with tooltip
- Show structured data indicator
- Show brand match badge (MCC 3000-3831 only)
- Show data source priority indicator

---

## Phase 5: Backend API Updates

### 5.1 Update Classification Job to Store Metadata

**File**: `services/merchant-service/internal/jobs/classification_job.go`

**Location**: `saveResultToDB()` method (lines 553-616)

**Changes**:

- Store page analysis metadata in `classification_data` JSONB under `metadata.pageAnalysis`
- Store brand match information under `metadata.brandMatch` (only for MCC 3000-3831)
- Store data source priority information under `metadata.dataSourcePriority`
- Include analysis method (multi_page, single_page, url_only) in metadata

**Implementation Details**:

```go
// In saveResultToDB(), after line 576 where metadata is added:
if len(result.Metadata) > 0 {
    classificationData["metadata"] = result.Metadata
} else {
    // Initialize metadata if not present
    classificationData["metadata"] = make(map[string]interface{})
}

// Add page analysis metadata if available
if pageAnalysis, ok := result.Metadata["pageAnalysis"].(map[string]interface{}); ok {
    metadata := classificationData["metadata"].(map[string]interface{})
    metadata["pageAnalysis"] = pageAnalysis
}

// Add brand match metadata if available (only for MCC 3000-3831)
if brandMatch, ok := result.Metadata["brandMatch"].(map[string]interface{}); ok {
    metadata := classificationData["metadata"].(map[string]interface{})
    metadata["brandMatch"] = brandMatch
}

// Add data source priority metadata
metadata := classificationData["metadata"].(map[string]interface{})
metadata["dataSourcePriority"] = map[string]string{
    "websiteContent": "primary",
    "businessName":    "secondary", // or "none" if not used
    "websiteURL":     "fallback",  // or "none" if not used
}
```

### 5.2 Update Classification Service Response Parsing

**File**: `services/merchant-service/internal/jobs/classification_job.go`

**Location**: `extractClassificationFromResponse()` method (line 406)

**Changes**:

- Extract page analysis metadata from classification service response
- Extract brand match information if present (only for MCC 3000-3831)
- Extract analysis method from response (multi_page, single_page, url_only)
- Store in result.Metadata for persistence

**Implementation Details**:

```go
// In extractClassificationFromResponse(), extract metadata:
if metadata, ok := response["metadata"].(map[string]interface{}); ok {
    result.Metadata["pageAnalysis"] = metadata["pageAnalysis"]
    result.Metadata["brandMatch"] = metadata["brandMatch"]
    result.Metadata["analysisMethod"] = metadata["analysisMethod"]
}
```

### 5.3 Update Classification Handler Response

**File**: `services/merchant-service/internal/handlers/merchant.go`

**Location**: `HandleMerchantSpecificAnalytics()` method (around line 1700)

**Changes**:

- Include metadata in response if present in classification_data
- Ensure backward compatibility (metadata is optional)
- Format metadata according to frontend TypeScript interface

**Implementation Details**:

```go
// In HandleMerchantSpecificAnalytics(), when building response:
classification := map[string]interface{}{
    "primaryIndustry": classificationData["primaryIndustry"],
    "confidenceScore": classificationData["confidenceScore"],
    "riskLevel":       classificationData["riskLevel"],
}

// Add industry codes...
if mccCodes, ok := classificationData["mccCodes"].([]interface{}); ok {
    classification["mccCodes"] = mccCodes
}
// ... similar for SIC and NAICS

// Add metadata if present (backward compatible)
if metadata, ok := classificationData["metadata"].(map[string]interface{}); ok {
    classification["metadata"] = metadata
}
```

### 5.4 Add Error Handling for Multi-Page Analysis

**File**: `internal/classification/repository/supabase_repository.go`

**Location**: `extractKeywordsFromMultiPageWebsite()` method

**Error Handling**:

- Handle timeout errors gracefully (fallback to single-page)
- Handle network errors (retry with exponential backoff)
- Handle robots.txt blocking (skip blocked pages)
- Handle invalid URLs (skip and continue)
- Log errors but continue processing other pages
- Set analysis method to "single_page" if multi-page fails
- Fallback to single-page if < 3 pages successfully analyzed

**Implementation Details**:

```go
func (r *SupabaseKeywordRepository) extractKeywordsFromMultiPageWebsite(ctx context.Context, websiteURL string) []string {
    // Create overall timeout context (60 seconds)
    overallCtx, overallCancel := context.WithTimeout(ctx, 60*time.Second)
    defer overallCancel()
    
    // Use SmartWebsiteCrawler to discover and analyze pages
    crawler := NewSmartWebsiteCrawler(r.logger)
    crawlResult, err := crawler.CrawlWebsite(overallCtx, websiteURL)
    if err != nil {
        r.logger.Printf("⚠️ Multi-page crawl failed: %v, falling back to single-page", err)
        return []string{} // Will trigger fallback
    }
    
    // Analyze pages with error handling
    successfulPages := 0
    consecutiveFailures := 0
    pageKeywords := make(map[string]float64) // keyword -> weighted score
    
    for _, pageAnalysis := range crawlResult.PagesAnalyzed {
        if consecutiveFailures >= 3 {
            r.logger.Printf("⚠️ Circuit breaker triggered: 3+ consecutive failures")
            break
        }
        
        if pageAnalysis.Error != "" {
            consecutiveFailures++
            r.logger.Printf("⚠️ Failed to analyze page %s: %s", pageAnalysis.URL, pageAnalysis.Error)
            continue
        }
        
        consecutiveFailures = 0
        successfulPages++
        
        // Extract keywords from page (text + structured data)
        keywords := r.extractKeywordsFromPageContent(pageAnalysis)
        
        // Weight keywords by page relevance score
        relevanceScore := pageAnalysis.RelevanceScore
        for _, keyword := range keywords {
            pageKeywords[keyword] += relevanceScore
        }
    }
    
    // If < 3 pages successfully analyzed, return empty (will trigger fallback)
    if successfulPages < 3 {
        r.logger.Printf("⚠️ Only %d pages successfully analyzed (< 3), will fallback to single-page", successfulPages)
        return []string{}
    }
    
    // Sort keywords by weighted score and return top 30
    // ... return keywords ...
}
```

---

## Phase 6: Testing Strategy

### 6.1 Unit Tests

**New File**: `internal/classification/repository/supabase_repository_priority_test.go`

**Test Cases**:

1. `TestExtractKeywords_PriorityOrder`: Verify website content extracted first, business name last
2. `TestExtractKeywords_BrandMatch`: Verify business name only extracted for brand matches in MCC 3000-3831
3. `TestExtractKeywords_NonBrandMatch`: Verify business name skipped for non-brand matches
4. `TestExtractKeywordsFromMultiPageWebsite`: Test multi-page analysis with relevance weighting
5. `TestExtractKeywords_StructuredData`: Test structured data extraction and integration
6. `TestCalculatePagePriority`: Test page priority calculation with new weights
7. `TestCalculateRelevanceScore`: Test relevance scoring with structured data boost
8. `TestMultiPageAnalysis_Timeout`: Test 60s overall timeout handling
9. `TestMultiPageAnalysis_Fallback`: Test fallback to single-page if < 3 pages succeed

**New File**: `internal/classification/repository/brand_matcher_test.go`

**Test Cases**:

1. `TestIsHighConfidenceBrandMatch_KnownBrands`: Test known hotel brand matching (MCC 3000-3831)
2. `TestIsHighConfidenceBrandMatch_UnknownBrands`: Test unknown brands return false
3. `TestIsHighConfidenceBrandMatch_Normalization`: Test name normalization (Inc, LLC, Corp removal)
4. `TestIsHighConfidenceBrandMatch_MCCRange`: Test brand matching only applies to MCC 3000-3831

### 6.2 Integration Tests

**New File**: `services/merchant-service/test/integration/classification_priority_test.go`

**Test Cases**:

1. `TestClassificationFlow_WebsiteFirstPriority`: Full flow with website-first priority
2. `TestClassificationFlow_MultiPageAnalysis`: Test multi-page analysis integration
3. `TestClassificationFlow_StructuredData`: Test structured data extraction in full flow
4. `TestClassificationFlow_BrandMatch`: Test brand matching in classification flow (MCC 3000-3831 only)
5. `TestClassificationFlow_FallbackChain`: Test fallback from multi-page → single-page → URL
6. `TestClassificationFlow_Performance`: Test multi-page analysis completes in <60s

### 6.3 Frontend Tests

**New File**: `frontend/__tests__/components/merchant/ClassificationMetadata.test.tsx`

**Test Cases**:

1. `TestClassificationMetadata_DisplaysPagesAnalyzed`: Test pages analyzed count display
2. `TestClassificationMetadata_DisplaysStructuredData`: Test structured data indicator
3. `TestClassificationMetadata_DisplaysBrandMatch`: Test brand match badge (MCC 3000-3831)
4. `TestClassificationMetadata_DisplaysDataSource`: Test data source priority indicator
5. `TestClassificationMetadata_CompactMode`: Test compact display mode
6. `TestClassificationMetadata_BackwardCompatibility`: Test graceful handling of missing metadata

**Update File**: `frontend/__tests__/components/merchant/BusinessAnalyticsTab.test.tsx`

**Test Cases**:

1. Update existing tests to include metadata
2. Test metadata display in classification card
3. Test backward compatibility (no metadata)

### 6.4 E2E Tests

**New File**: `test/e2e/classification_priority.spec.ts`

**Test Cases**:

1. `test('classification prioritizes website content over business name')`: Create merchant, verify website content used
2. `test('multi-page analysis extracts more keywords')`: Verify multi-page analysis improves accuracy
3. `test('structured data improves classification')`: Verify structured data extraction works
4. `test('brand match uses business name for hotels')`: Test known hotel brand (e.g., Hilton) uses business name
5. `test('frontend displays classification metadata')`: Verify metadata displayed in UI
6. `test('multi-page analysis completes in <60s')`: Verify performance target

### 6.5 Performance Tests

**New File**: `test/performance/multi_page_analysis_test.go`

**Test Cases**:

1. `BenchmarkMultiPageAnalysis`: Benchmark multi-page analysis performance (target: <60s p95)
2. `TestMultiPageAnalysis_Timeout`: Test timeout handling for slow websites
3. `TestMultiPageAnalysis_ConcurrentLimit`: Test concurrent request limiting
4. `TestMultiPageAnalysis_MemoryUsage`: Test memory usage stays under 500MB

---

## Phase 7: Configuration & Deployment

### 7.1 Environment Variables

**New Variables**:

- `CLASSIFICATION_MAX_PAGES_TO_ANALYZE`: Maximum pages to analyze (default: 15)
- `CLASSIFICATION_PAGE_ANALYSIS_TIMEOUT`: Timeout per page (default: 15s)
- `CLASSIFICATION_OVERALL_TIMEOUT`: Overall timeout for multi-page analysis (default: 60s)
- `CLASSIFICATION_CONCURRENT_PAGES`: Max concurrent page requests (default: 5)
- `CLASSIFICATION_BRAND_MATCH_ENABLED`: Enable brand matching (default: true)
- `CLASSIFICATION_BRAND_MATCH_MCC_RANGE`: MCC range for brand matching (default: "3000-3831")

### 7.2 Feature Flags

**New Flags** (Environment Variables):

- `ENABLE_MULTI_PAGE_ANALYSIS`: Enable multi-page analysis (default: true)
- `ENABLE_STRUCTURED_DATA_EXTRACTION`: Enable structured data extraction (default: true)
- `ENABLE_BRAND_MATCHING`: Enable brand matching (default: true)

**Implementation**:
- Check flags at service initialization
- Use environment variables with sensible defaults
- Flags available for emergency rollback only (big bang deployment)
- No runtime flag updates (requires service restart)

### 7.3 Monitoring & Logging

**Prometheus Metrics** (Export to Prometheus):

**Counter Metrics**:
- `classification_pages_analyzed_total{method="multi_page|single_page|url_only"}`: Total pages analyzed by method
- `classification_structured_data_found_total{found="true|false"}`: Count of structured data found
- `classification_brand_matches_total{mcc_range="3000-3831|other|none"}`: Count of brand matches by MCC range

**Histogram Metrics**:
- `classification_analysis_duration_seconds{method="multi_page|single_page|url_only"}`: Analysis duration by method
- `classification_page_analysis_duration_seconds{page_type="about|services|products|other"}`: Per-page analysis duration

**Gauge Metrics**:
- `classification_concurrent_pages_analyzing`: Current concurrent page analyses
- `classification_memory_usage_bytes`: Memory usage per classification job

**Alert Thresholds**:
- `classification_analysis_duration_seconds{p95} > 60`: Warning (exceeds performance target)
- `classification_analysis_duration_seconds{p99} > 90`: Critical
- `classification_pages_analyzed_total{method="multi_page"} / classification_pages_analyzed_total * 100 < 50`: Warning (multi-page success rate < 50%)
- `classification_memory_usage_bytes > 500000000`: Warning (exceeds 500MB limit)

**Enhanced Logging**:

- Log page analysis metadata in classification job
- Log brand match results (with MCC range validation)
- Log structured data extraction results
- Log priority order used
- Log fallback chain triggers (multi-page → single-page → URL)

---

## Implementation Order

1. **Phase 1**: Fix priority weights (immediate impact)
2. **Phase 2**: Enhance multi-page analysis (accuracy improvement)
3. **Phase 3**: Supabase schema (metadata storage)
4. **Phase 4**: Frontend enhancements (user visibility)
5. **Phase 5**: Backend API updates (data flow)
6. **Phase 6**: Testing (quality assurance - comprehensive due to production-only environment)
7. **Phase 7**: Configuration & deployment (production readiness - big bang deployment)

## Deployment Strategy

**Big Bang Deployment**:
- All features enabled simultaneously in production
- No staging environment - all testing must be comprehensive before deployment
- Feature flags available for emergency rollback only
- Monitor closely for first 24 hours post-deployment

**Pre-Deployment Checklist**:
- [ ] All unit tests pass (>90% coverage)
- [ ] All integration tests pass
- [ ] Performance benchmarks validate <60s target (p95)
- [ ] Load tests completed successfully
- [ ] Code review approved
- [ ] Prometheus metrics verified and exported
- [ ] Rollback plan tested (feature flags work)
- [ ] Monitoring dashboards created
- [ ] Alert thresholds configured in Prometheus
- [ ] Backward compatibility verified (all metadata fields optional)

---

## Success Metrics

1. Website content keywords contribute 80%+ to final classification (vs current ~40%)
2. Business name keywords only used for brand matches in MCC 3000-3831 (<5% of cases)
3. Multi-page analysis increases keyword relevance by 30%+
4. Structured data extraction improves accuracy by 15%+
5. Industry-revealing pages (about, products, services) weighted 2x higher than other pages
6. Frontend displays classification metadata for transparency
7. All tests pass with >90% coverage
8. **Performance**: Multi-page analysis completes in <60s for 95% of cases (p95)
9. **Reliability**: <5% failure rate for multi-page analysis
10. **Backward Compatibility**: 100% of existing API clients continue working (all metadata optional)

---

## Rollback Plan (Emergency Only)

If critical issues arise in production:

1. **Immediate**: Disable multi-page analysis via `ENABLE_MULTI_PAGE_ANALYSIS=false` (requires service restart)
2. **If needed**: Revert context multipliers via code deployment (quick fix)
3. **If needed**: Disable brand matching via `ENABLE_BRAND_MATCHING=false`
4. **If needed**: Disable structured data extraction via `ENABLE_STRUCTURED_DATA_EXTRACTION=false`
5. Frontend gracefully handles missing metadata (all fields optional - no frontend rollback needed)

**Note**: Big bang deployment means all features go live together. Rollback should be rare but feature flags provide safety net.

---

## Files Summary

### New Files

1. `internal/classification/repository/brand_matcher.go`
2. `internal/classification/repository/supabase_repository_priority_test.go`
3. `internal/classification/repository/brand_matcher_test.go`
4. `services/merchant-service/test/integration/classification_priority_test.go`
5. `frontend/components/merchant/ClassificationMetadata.tsx`
6. `frontend/__tests__/components/merchant/ClassificationMetadata.test.tsx`
7. `test/e2e/classification_priority.spec.ts`
8. `test/performance/multi_page_analysis_test.go`

### Modified Files

1. `internal/classification/repository/supabase_repository.go` (extractKeywords, extractKeywordsFromWebsite, new extractKeywordsFromMultiPageWebsite)
2. `internal/classification/repository/enhanced_scoring_algorithm.go` (weights, context multiplier)
3. `internal/classification/smart_website_crawler.go` (page priority, relevance scoring)
4. `services/merchant-service/internal/jobs/classification_job.go` (metadata storage)
5. `services/merchant-service/internal/handlers/merchant.go` (response format)
6. `frontend/types/merchant.ts` (ClassificationData interface)
7. `frontend/lib/api-validation.ts` (validation schema)
8. `frontend/components/merchant/BusinessAnalyticsTab.tsx` (metadata display)
9. `frontend/__tests__/components/merchant/BusinessAnalyticsTab.test.tsx` (test updates)

### Supabase Migrations

1. `supabase-migrations/020_add_page_analysis_metadata.sql` (documentation only - metadata stored in JSONB)