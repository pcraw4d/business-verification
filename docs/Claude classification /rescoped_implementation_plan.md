# Re-Scoped Implementation Plan: Industry Classification System
## From <5% Accuracy to Production-Ready Hybrid System

**Date:** January 27, 2025  
**Project:** KYB Platform - Industry Classification Module  
**Timeline:** 9 weeks (flexible, side project)  
**Deployment:** Railway + Supabase (preserved)

---

## Executive Summary

### Current State Assessment

Your system has **excellent architecture** but is failing at **execution fundamentals**:

✅ **What's Working Well:**
- Sophisticated Go-based multi-strategy classifier (4 parallel strategies)
- Solid database schema with trigram and full-text search
- Good caching infrastructure (Redis + in-memory)
- Railway + Supabase deployment (keep this)
- Clean UI design (preserve this)
- 60-70% of code is reusable

❌ **Critical Failures (<5% accuracy):**
1. **Website scraper fails/returns no output** → Garbage In = Garbage Out
2. **Only returns top 1 code** (should return top 3 with confidence scores)
3. **Generic classifications** ("General Business" instead of specific industry)
4. **ML service (DistilBART) not integrated** → missing accuracy boost
5. **No explanation/reasoning shown** → can't audit decisions
6. **Confidence too low** (50% shown in UI)

### Root Cause Analysis

```
❌ Scraper Fails (50% of time)
    ↓
❌ No/Poor Content Extracted
    ↓
❌ Multi-Strategy Classifier Has Nothing to Work With
    ↓  
❌ Falls Back to Generic Classification
    ↓
❌ Low Confidence, Wrong Industry, Only 1 Code
    ↓
❌ <5% Accuracy
```

**The Good News:** The problem isn't your architecture - it's execution at the scraping layer and conservative code generation. Both are fixable.

### Re-Scoped Strategy

Instead of building from scratch, we'll **fix the foundation first**, then **layer on enhancements**:

**Phase 1 (Weeks 1-2):** Fix scraping + content extraction (ROOT CAUSE)  
**Phase 2 (Weeks 3-4):** Enhance Layer 1 (return top 3, better confidence)  
**Phase 3 (Weeks 5-6):** Add Layer 2 (embeddings for edge cases)  
**Phase 4 (Weeks 7-8):** Add Layer 3 (LLM for complex cases)  
**Phase 5 (Week 9):** Add explanation UI + testing

**Expected Outcome:**
- 50% → 85%+ accuracy (Phase 1-2 alone)
- <100ms for 70% of requests (Layer 1)
- Top 3 codes with confidence scores
- Full auditability with explanations

---

## Phase 1: Fix the Foundation (Weeks 1-2)
### Goal: Scraping → 95% success rate, Rich content extraction

**This is the most critical phase.** If scraping fails, everything downstream fails.

### 1.1 Diagnose Current Scraper Issues

**Task:** Add comprehensive logging to understand failure modes

```go
// internal/external/website_scraper.go
// Add detailed logging around:

func (s *WebsiteScraper) Scrape(url string) (*ScrapedContent, error) {
    log.Info("Starting scrape", "url", url)
    
    // Log every step:
    // - DNS resolution
    // - HTTP response code
    // - Content length
    // - Parsing success/failure
    // - Errors encountered
    
    if err != nil {
        log.Error("Scrape failed", 
            "url", url, 
            "error", err,
            "stage", "http_fetch", // or "html_parse", "text_extract"
            "response_code", resp.StatusCode)
    }
    
    log.Info("Scrape completed", 
        "url", url, 
        "content_length", len(content),
        "word_count", wordCount,
        "has_title", hasTitle,
        "has_description", hasDescription)
}
```

**Files to Modify:**
- `internal/external/website_scraper.go`

**Acceptance Criteria:**
- [ ] Can identify why scrapes fail (bot detection? timeout? parsing error?)
- [ ] Logging shows content length and quality metrics
- [ ] Dashboard shows scrape success rate

### 1.2 Enhance Content Extraction

**Current Issue:** Even when scraping succeeds, content quality is poor.

**Improvements:**

```go
// internal/external/website_scraper.go

type ScrapedContent struct {
    // Existing
    RawHTML     string
    PlainText   string
    
    // NEW: Add structured extraction
    Title       string
    MetaDesc    string
    Headings    []string        // H1, H2, H3
    KeyPhrases  []string        // Extract important phrases
    NavMenu     []string        // Navigation items (business areas)
    ProductList []string        // Product/service listings
    AboutText   string          // Specifically extract "About" section
    WordCount   int
    Language    string
}

func (s *WebsiteScraper) extractStructuredContent(doc *html.Node) *ScrapedContent {
    content := &ScrapedContent{}
    
    // Extract title
    content.Title = extractTitle(doc)
    
    // Extract meta description
    content.MetaDesc = extractMetaDescription(doc)
    
    // Extract all headings (H1-H3) - these are high-signal
    content.Headings = extractHeadings(doc)
    
    // Extract navigation menu - indicates business areas
    content.NavMenu = extractNavigation(doc)
    
    // Find and prioritize "About" section
    content.AboutText = extractAboutSection(doc)
    
    // Extract product/service listings
    content.ProductList = extractProductsServices(doc)
    
    // Clean and combine all text
    content.PlainText = cleanAndCombineText(
        content.Title,
        content.MetaDesc,
        content.Headings,
        content.AboutText,
        content.ProductList,
    )
    
    return content
}
```

**Priority Weighting for Classification:**
```
Title/Meta Description: 30% weight (most important)
About Section: 25% weight
Headings: 20% weight
Navigation Menu: 15% weight
Product Listings: 10% weight
```

**Files to Modify:**
- `internal/external/website_scraper.go`

**Acceptance Criteria:**
- [ ] Extracts structured content (title, headings, about, etc.)
- [ ] Content quality score ≥ 0.7 for 90%+ of successful scrapes
- [ ] Average word count ≥ 200 words

### 1.3 Add Fallback Scraping Strategies

**Problem:** Simple HTTP client fails on JS-heavy sites, bot detection, etc.

**Solution:** Multi-tier fallback approach

```go
// internal/external/website_scraper.go

type ScraperStrategy interface {
    Scrape(url string) (*ScrapedContent, error)
    Name() string
}

type WebsiteScraper struct {
    strategies []ScraperStrategy // Try in order
}

// Strategy 1: Simple HTTP (fastest, works 60% of time)
type SimpleHTTPScraper struct {}

// Strategy 2: HTTP with browser headers (works 80% of time)
type BrowserHeadersScraper struct {}

// Strategy 3: Headless browser via external service (works 95% of time)
type PlaywrightScraper struct {
    serviceURL string // Separate Playwright service on Railway
}

func (s *WebsiteScraper) Scrape(url string) (*ScrapedContent, error) {
    var lastErr error
    
    for _, strategy := range s.strategies {
        log.Info("Trying scraper strategy", "strategy", strategy.Name(), "url", url)
        
        content, err := strategy.Scrape(url)
        if err == nil && isContentValid(content) {
            log.Info("Strategy succeeded", "strategy", strategy.Name())
            return content, nil
        }
        
        lastErr = err
        log.Warn("Strategy failed, trying next", 
            "strategy", strategy.Name(), 
            "error", err)
    }
    
    return nil, fmt.Errorf("all strategies failed: %w", lastErr)
}
```

**Implementation Options:**

**Option A (Quick Win):** Improve existing Go scraper
- Add proper User-Agent headers
- Add cookie handling
- Add retry with backoff
- Better HTML parsing

**Option B (Better Success Rate):** Add Playwright service
- Deploy small Node.js service with Playwright
- Only called when Strategy 1/2 fail (15-20% of requests)
- Handles JS rendering, bot detection bypass

**Recommendation:** Start with Option A (Week 1), add Option B if needed (Week 2)

**New Railway Service (Optional):**
```javascript
// playwright-scraper-service/index.js
const express = require('express');
const playwright = require('playwright');

app.post('/scrape', async (req, res) => {
    const { url } = req.body;
    
    const browser = await playwright.chromium.launch();
    const page = await browser.newPage();
    
    await page.goto(url, { waitUntil: 'networkidle' });
    const content = await page.content();
    
    await browser.close();
    
    res.json({ content });
});
```

**Files to Modify:**
- `internal/external/website_scraper.go` (add strategies)
- New service: `services/playwright-scraper/` (optional)

**Acceptance Criteria:**
- [ ] Scrape success rate ≥ 95%
- [ ] Average scrape time < 3 seconds
- [ ] Fallback to Playwright for <20% of requests
- [ ] No "no output" errors

### 1.4 Content Quality Validation

Add validation before sending content to classifier:

```go
func isContentValid(content *ScrapedContent) bool {
    // Minimum requirements
    if content.WordCount < 50 {
        return false // Not enough content
    }
    
    if content.Title == "" && content.MetaDesc == "" {
        return false // No basic metadata
    }
    
    // Check for error pages
    if containsErrorIndicators(content.PlainText) {
        return false // 404, Access Denied, etc.
    }
    
    // Quality score
    qualityScore := calculateContentQuality(content)
    return qualityScore >= 0.7
}

func calculateContentQuality(content *ScrapedContent) float64 {
    score := 0.0
    
    // Has title? +0.2
    if content.Title != "" { score += 0.2 }
    
    // Has meta description? +0.2
    if content.MetaDesc != "" { score += 0.2 }
    
    // Has headings? +0.2
    if len(content.Headings) > 0 { score += 0.2 }
    
    // Sufficient word count? +0.2
    if content.WordCount >= 200 { score += 0.2 }
    
    // Has about section? +0.2
    if content.AboutText != "" { score += 0.2 }
    
    return score
}
```

**Files to Modify:**
- `internal/external/website_scraper.go`

**Acceptance Criteria:**
- [ ] Content quality score calculated for every scrape
- [ ] Invalid content rejected before classification
- [ ] Quality metrics logged for monitoring

### Phase 1 Success Metrics

**Before Phase 1:**
- Scrape success rate: ~50%
- Content quality: Poor
- "No output" errors: Common
- Accuracy: <5%

**After Phase 1:**
- Scrape success rate: ≥95%
- Content quality score: ≥0.7 for 90%+ of scrapes
- Average word count: ≥200 words
- "No output" errors: <2%
- **Expected accuracy improvement: 50-60%** (just from better input)

---

## Phase 2: Enhance Layer 1 (Weeks 3-4)
### Goal: Multi-strategy classifier → 80%+ accuracy, top 3 codes, fast

Your multi-strategy classifier is well-designed but too conservative. Let's make it more decisive.

### 2.1 Return Top 3 Codes (Not Just Top 1)

**Current Issue:** Only returns 1 code per type (MCC/SIC/NAICS)

**Fix:** Return top 3 with confidence scores

```go
// internal/classification/classifier.go

type CodeResult struct {
    Code        string
    Description string
    Confidence  float64
    Source      string // "keyword_match", "crosswalk", "ml_prediction"
}

type ClassificationCodes struct {
    MCC   []CodeResult // Top 3
    SIC   []CodeResult // Top 3
    NAICS []CodeResult // Top 3
}

func (g *ClassificationCodeGenerator) GenerateCodes(
    ctx context.Context,
    industryName string,
    keywords []string,
    confidence float64,
) (*ClassificationCodes, error) {
    
    codes := &ClassificationCodes{}
    
    // For each code type, get top 5 candidates, return top 3
    mccCandidates := g.getMCCCandidates(ctx, industryName, keywords)
    codes.MCC = selectTopCodes(mccCandidates, 3)
    
    sicCandidates := g.getSICCandidates(ctx, industryName, keywords)
    codes.SIC = selectTopCodes(sicCandidates, 3)
    
    naicsCandidates := g.getNAICSCandidates(ctx, industryName, keywords)
    codes.NAICS = selectTopCodes(naicsCandidates, 3)
    
    // Use crosswalks to fill gaps
    codes = g.enrichWithCrosswalks(codes)
    
    return codes, nil
}

func selectTopCodes(candidates []CodeResult, limit int) []CodeResult {
    // Sort by confidence
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].Confidence > candidates[j].Confidence
    })
    
    // Return top N
    if len(candidates) > limit {
        return candidates[:limit]
    }
    return candidates
}
```

**Files to Modify:**
- `internal/classification/classifier.go`
- `internal/classification/service.go`
- `services/classification-service/internal/handlers/classification.go` (update response format)

**Acceptance Criteria:**
- [ ] Returns top 3 MCC, SIC, NAICS codes
- [ ] Each code has confidence score
- [ ] Response format updated in API

### 2.2 Improve Confidence Calibration

**Current Issue:** Confidence too low (50% shown in UI)

**Solution:** Better confidence scoring that reflects actual accuracy

```go
// internal/classification/confidence_calibrator.go

func (c *ConfidenceCalibrator) CalibrateConfidence(
    strategyResults map[string]float64,
    contentQuality float64,
    codeAgreement float64, // How much do MCC/SIC/NAICS agree?
) float64 {
    
    baseConfidence := calculateWeightedAverage(strategyResults)
    
    // Boost: High content quality
    if contentQuality > 0.8 {
        baseConfidence *= 1.1
    }
    
    // Boost: Strong agreement between strategy results
    variance := calculateVariance(strategyResults)
    if variance < 0.1 { // Low variance = high agreement
        baseConfidence *= 1.15
    }
    
    // Boost: MCC/SIC/NAICS codes align (crosswalks validate)
    if codeAgreement > 0.8 {
        baseConfidence *= 1.2
    }
    
    // Boost: Keyword match was strong
    if strategyResults["keyword"] > 0.85 {
        baseConfidence *= 1.1
    }
    
    // Cap at 0.95 (never claim 100% certainty)
    return math.Min(baseConfidence, 0.95)
}
```

**Files to Modify:**
- `internal/classification/confidence_calibrator.go`

**Acceptance Criteria:**
- [ ] Confidence scores reflect actual accuracy (validate with test set)
- [ ] High-quality classifications show 80-95% confidence
- [ ] Low-quality show 50-70% confidence appropriately

### 2.3 Optimize for Speed (<100ms for Layer 1)

**Current Performance:** Unknown, likely 200-500ms

**Target:** <100ms for high-confidence cases

**Optimizations:**

```go
// internal/classification/multi_strategy_classifier.go

func (c *MultiStrategyClassifier) ClassifyFastPath(
    ctx context.Context,
    content *ScrapedContent,
) (*ClassificationResult, bool) {
    
    // Fast path: Obvious keywords (e.g., "restaurant", "dentist")
    if obviousKeyword := c.detectObviousKeyword(content); obviousKeyword != "" {
        // Direct database lookup, single query
        result := c.repo.GetIndustryByKeyword(ctx, obviousKeyword)
        if result != nil {
            return &ClassificationResult{
                Industry:   result.Name,
                Confidence: 0.95,
                Method:     "fast_path_keyword",
                ProcessingTimeMs: 10, // Very fast
            }, true
        }
    }
    
    // Not a fast path case
    return nil, false
}

func (c *MultiStrategyClassifier) ClassifyWithMultiStrategy(
    ctx context.Context,
    businessName, description, websiteURL string,
) (*ClassificationResult, error) {
    
    content := c.scraper.Scrape(websiteURL)
    
    // Try fast path first
    if result, isFastPath := c.ClassifyFastPath(ctx, content); isFastPath {
        return result, nil
    }
    
    // Fall back to full multi-strategy (existing code)
    return c.classifyFullStrategy(ctx, content)
}
```

**Database Query Optimization:**

```sql
-- Add index for fast keyword lookups
CREATE INDEX idx_keyword_weights_fast_lookup 
ON keyword_weights (keyword, weight DESC)
WHERE weight > 0.8; -- Only high-confidence keywords

-- Optimize trigram queries
ANALYZE industry_keywords;
ANALYZE classification_codes;
```

**Files to Modify:**
- `internal/classification/multi_strategy_classifier.go`
- New migration: `supabase-migrations/040_optimize_fast_path.sql`

**Acceptance Criteria:**
- [ ] Fast path handles 60-70% of requests
- [ ] Fast path latency <100ms
- [ ] Full strategy latency <500ms
- [ ] Overall p95 latency <300ms

### 2.4 Add Reasoning/Explanation Generation

**Current Issue:** Explanation exists in code but not shown in UI

**Solution:** Generate structured, auditable explanations

```go
// internal/classification/service.go

type ClassificationExplanation struct {
    PrimaryReason   string   // "Strong keyword match for 'restaurant'"
    SupportingFactors []string // ["Menu items found", "Reservation system", "Multiple locations"]
    ConfidenceFactors map[string]float64 // {"keyword_match": 0.9, "entity_match": 0.85}
    KeyTermsFound   []string // ["menu", "reservations", "dining"]
    MethodUsed      string   // "multi_strategy"
}

func (s *IndustryDetectionService) generateExplanation(
    result *ClassificationResult,
    content *ScrapedContent,
) *ClassificationExplanation {
    
    exp := &ClassificationExplanation{
        MethodUsed: result.Method,
        ConfidenceFactors: result.StrategyScores,
        KeyTermsFound: result.Keywords,
    }
    
    // Determine primary reason
    highestStrategy := getHighestScoringStrategy(result.StrategyScores)
    exp.PrimaryReason = formatPrimaryReason(highestStrategy, result.Industry)
    
    // Add supporting factors
    exp.SupportingFactors = extractSupportingFactors(content, result)
    
    return exp
}

func formatPrimaryReason(strategy string, industry string) string {
    templates := map[string]string{
        "keyword": "Strong keyword match for '%s' industry",
        "entity": "Business entities match '%s' sector",
        "topic": "Content topic analysis indicates '%s'",
        "co_occurrence": "Pattern matching suggests '%s' industry",
    }
    
    return fmt.Sprintf(templates[strategy], industry)
}
```

**Files to Modify:**
- `internal/classification/service.go`
- `services/classification-service/internal/handlers/classification.go` (add to response)

**Acceptance Criteria:**
- [ ] Every classification includes explanation
- [ ] Explanation is human-readable
- [ ] Includes specific evidence (keywords found, etc.)
- [ ] Ready to display in UI

### 2.5 Fix "General Business" Problem

**Root Cause:** Classifier falls back to generic category when uncertain

**Solution:** Be more specific or admit uncertainty

```go
// internal/classification/service.go

func (s *IndustryDetectionService) selectIndustry(
    candidates []IndustryCandidate,
    minConfidence float64,
) (*IndustryCandidate, error) {
    
    // Sort by confidence
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].Confidence > candidates[j].Confidence
    })
    
    topCandidate := candidates[0]
    
    // Reject generic industries if confidence is low
    genericIndustries := []string{"General Business", "Other Services", "Miscellaneous"}
    if contains(genericIndustries, topCandidate.Name) && topCandidate.Confidence < 0.7 {
        // Try next candidate if it's more specific
        if len(candidates) > 1 && !contains(genericIndustries, candidates[1].Name) {
            return &candidates[1], nil
        }
    }
    
    // Require minimum confidence
    if topCandidate.Confidence < minConfidence {
        return nil, ErrInsufficientConfidence
    }
    
    return &topCandidate, nil
}
```

**Files to Modify:**
- `internal/classification/service.go`

**Acceptance Criteria:**
- [ ] "General Business" only used when truly appropriate
- [ ] Specific industries preferred even at slightly lower confidence
- [ ] Generic fallback requires 70%+ confidence

### Phase 2 Success Metrics

**Before Phase 2:**
- Returns: 1 code per type
- Confidence: ~50%
- Speed: Unknown
- Explanation: Exists but not used
- Generic classifications: Common
- Accuracy: 50-60% (after Phase 1)

**After Phase 2:**
- Returns: Top 3 codes per type with confidence
- Confidence: Calibrated 70-95%
- Speed: <100ms for 70% of requests
- Explanation: Generated and ready for UI
- Specific classifications: 90%+
- **Expected accuracy: 80-85%**

---

## Phase 3: Add Layer 2 - Embeddings (Weeks 5-6)
### Goal: Handle edge cases, improve accuracy to 85-90%

Layer 1 handles 70-80% of cases well. Layer 2 catches the remaining 20-30% that are ambiguous or novel.

### 3.1 Enable pgvector in Supabase

```sql
-- supabase-migrations/050_enable_pgvector.sql

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create code embeddings table
CREATE TABLE code_embeddings (
    id BIGSERIAL PRIMARY KEY,
    code_type VARCHAR(10) NOT NULL, -- 'MCC', 'SIC', 'NAICS'
    code VARCHAR(10) NOT NULL,
    description TEXT NOT NULL,
    extended_description TEXT, -- Enriched with examples, context
    industry_context TEXT, -- Related industries
    embedding vector(384), -- all-MiniLM-L6-v2
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(code_type, code)
);

-- Create index for similarity search
CREATE INDEX idx_code_embeddings_vector ON code_embeddings 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Additional indexes
CREATE INDEX idx_code_embeddings_type ON code_embeddings(code_type);
CREATE INDEX idx_code_embeddings_code ON code_embeddings(code);
```

**Files to Create:**
- `supabase-migrations/050_enable_pgvector.sql`

### 3.2 Pre-compute Code Embeddings

```python
# scripts/precompute_embeddings.py
from sentence_transformers import SentenceTransformer
from supabase import create_client
import os

# Load model
model = SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')

# Connect to Supabase
supabase = create_client(
    os.getenv("SUPABASE_URL"),
    os.getenv("SUPABASE_SERVICE_KEY")
)

# Fetch all codes
codes = supabase.table('classification_codes').select('*').execute()

embeddings_to_insert = []

for code in codes.data:
    # Create rich description for embedding
    text = f"{code['description']}. "
    
    # Add industry context
    if code.get('industry'):
        text += f"Industry: {code['industry']}. "
    
    # Add examples if available
    if code.get('examples'):
        text += f"Examples: {code['examples']}. "
    
    # Add keywords
    keywords = supabase.table('code_keywords').select('keyword').eq(
        'code', code['code']
    ).execute()
    
    if keywords.data:
        kw_list = [k['keyword'] for k in keywords.data]
        text += f"Related terms: {', '.join(kw_list[:10])}"
    
    # Generate embedding
    embedding = model.encode(text)
    
    embeddings_to_insert.append({
        'code_type': code['code_type'],
        'code': code['code'],
        'description': code['description'],
        'extended_description': text,
        'embedding': embedding.tolist()
    })
    
    print(f"✓ {code['code_type']} {code['code']}")

# Batch insert
batch_size = 100
for i in range(0, len(embeddings_to_insert), batch_size):
    batch = embeddings_to_insert[i:i+batch_size]
    supabase.table('code_embeddings').insert(batch).execute()
    print(f"Inserted batch {i//batch_size + 1}")

print(f"✓ Total embeddings: {len(embeddings_to_insert)}")
```

**Files to Create:**
- `scripts/precompute_embeddings.py`
- `scripts/requirements.txt` (sentence-transformers, supabase)

**Run Once:**
```bash
cd scripts
python precompute_embeddings.py
```

### 3.3 Create Embedding Service (Python on Railway)

```python
# services/embedding-service/app.py
from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
from typing import List

app = FastAPI()

# Load model on startup
model = SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')

class EmbedRequest(BaseModel):
    text: str

class EmbedResponse(BaseModel):
    embedding: List[float]
    dimension: int

@app.post("/embed")
async def create_embedding(request: EmbedRequest):
    # Truncate if too long
    text = request.text[:5000]
    
    # Generate embedding
    embedding = model.encode(text)
    
    return EmbedResponse(
        embedding=embedding.tolist(),
        dimension=len(embedding)
    )

@app.get("/health")
async def health():
    return {"status": "ok", "model": "all-MiniLM-L6-v2"}
```

**Files to Create:**
- `services/embedding-service/app.py`
- `services/embedding-service/requirements.txt`
- `services/embedding-service/Dockerfile`

**Railway Service Config:**
```yaml
# railway.toml
[[services]]
name = "embedding-service"
source = "services/embedding-service"

[build]
builder = "DOCKERFILE"

[deploy]
healthcheckPath = "/health"
numReplicas = 1
restartPolicyType = "ON_FAILURE"
```

### 3.4 Add Vector Search RPC Function

```sql
-- supabase-migrations/051_add_vector_search_function.sql

CREATE OR REPLACE FUNCTION match_code_embeddings(
    query_embedding vector(384),
    code_type_filter text,
    match_threshold float,
    match_count int
)
RETURNS TABLE (
    code text,
    description text,
    extended_description text,
    similarity float
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT
        ce.code,
        ce.description,
        ce.extended_description,
        1 - (ce.embedding <=> query_embedding) as similarity
    FROM code_embeddings ce
    WHERE ce.code_type = code_type_filter
        AND 1 - (ce.embedding <=> query_embedding) > match_threshold
    ORDER BY ce.embedding <=> query_embedding
    LIMIT match_count;
END;
$$;
```

**Files to Create:**
- `supabase-migrations/051_add_vector_search_function.sql`

### 3.5 Integrate Embedding Layer into Go Service

```go
// internal/classification/embedding_classifier.go

package classification

import (
    "context"
    "encoding/json"
    "net/http"
)

type EmbeddingClassifier struct {
    embeddingServiceURL string
    supabaseClient      *supabase.Client
    httpClient          *http.Client
}

type EmbeddingResult struct {
    MCC   []CodeResult
    SIC   []CodeResult
    NAICS []CodeResult
    Confidence float64
    Method string
}

func (e *EmbeddingClassifier) ClassifyByEmbedding(
    ctx context.Context,
    content *ScrapedContent,
) (*EmbeddingResult, error) {
    
    // Step 1: Generate embedding for content
    text := combineContentForEmbedding(content)
    embedding, err := e.getEmbedding(ctx, text)
    if err != nil {
        return nil, err
    }
    
    // Step 2: Search for similar codes (each type)
    result := &EmbeddingResult{Method: "embedding"}
    
    mccMatches, err := e.searchSimilarCodes(ctx, embedding, "MCC", 0.7, 5)
    result.MCC = selectTopCodes(mccMatches, 3)
    
    sicMatches, err := e.searchSimilarCodes(ctx, embedding, "SIC", 0.7, 5)
    result.SIC = selectTopCodes(sicMatches, 3)
    
    naicsMatches, err := e.searchSimilarCodes(ctx, embedding, "NAICS", 0.7, 5)
    result.NAICS = selectTopCodes(naicsMatches, 3)
    
    // Step 3: Calculate overall confidence
    result.Confidence = calculateEmbeddingConfidence(mccMatches[0].Confidence)
    
    return result, nil
}

func (e *EmbeddingClassifier) getEmbedding(ctx context.Context, text string) ([]float64, error) {
    // Call Python embedding service
    reqBody, _ := json.Marshal(map[string]string{"text": text})
    
    resp, err := e.httpClient.Post(
        e.embeddingServiceURL + "/embed",
        "application/json",
        bytes.NewReader(reqBody),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Embedding []float64 `json:"embedding"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    
    return result.Embedding, nil
}

func (e *EmbeddingClassifier) searchSimilarCodes(
    ctx context.Context,
    embedding []float64,
    codeType string,
    threshold float64,
    limit int,
) ([]CodeResult, error) {
    
    // Call Supabase RPC function
    results, err := e.supabaseClient.Rpc(
        "match_code_embeddings",
        map[string]interface{}{
            "query_embedding":   embedding,
            "code_type_filter": codeType,
            "match_threshold":  threshold,
            "match_count":      limit,
        },
    ).Execute()
    
    if err != nil {
        return nil, err
    }
    
    // Convert to CodeResult
    var codes []CodeResult
    for _, row := range results {
        codes = append(codes, CodeResult{
            Code:        row["code"].(string),
            Description: row["description"].(string),
            Confidence:  row["similarity"].(float64),
            Source:      "embedding",
        })
    }
    
    return codes, nil
}

func combineContentForEmbedding(content *ScrapedContent) string {
    // Prioritize high-signal content
    parts := []string{}
    
    if content.Title != "" {
        parts = append(parts, content.Title)
    }
    if content.MetaDesc != "" {
        parts = append(parts, content.MetaDesc)
    }
    if content.AboutText != "" {
        parts = append(parts, content.AboutText)
    }
    if len(content.Headings) > 0 {
        parts = append(parts, strings.Join(content.Headings, ". "))
    }
    
    // Combine with weights (title/meta more important)
    combined := strings.Join(parts, ". ")
    
    // Truncate to 5000 chars for embedding
    if len(combined) > 5000 {
        combined = combined[:5000]
    }
    
    return combined
}
```

**Files to Create:**
- `internal/classification/embedding_classifier.go`

**Files to Modify:**
- `internal/classification/service.go` (integrate embedding layer)

### 3.6 Add Layer Routing Logic

```go
// internal/classification/service.go

func (s *IndustryDetectionService) DetectIndustry(
    ctx context.Context,
    businessName, description, websiteURL string,
) (*IndustryDetectionResult, error) {
    
    // Scrape website
    content, err := s.scraper.Scrape(websiteURL)
    if err != nil {
        return nil, err
    }
    
    // Try Layer 1: Multi-strategy (fast path + full)
    layer1Result, err := s.multiStrategyClassifier.Classify(ctx, content)
    
    if layer1Result.Confidence >= 0.85 {
        // High confidence - use Layer 1 result
        return s.buildResult(layer1Result, "layer1"), nil
    }
    
    // Try Layer 2: Embedding-based
    layer2Result, err := s.embeddingClassifier.ClassifyByEmbedding(ctx, content)
    
    if layer2Result.Confidence >= 0.80 {
        // Good embedding match - use Layer 2
        return s.buildResult(layer2Result, "layer2"), nil
    }
    
    // Confidence still low - will add Layer 3 (LLM) in Phase 4
    // For now, use best available result with lower confidence
    if layer2Result.Confidence > layer1Result.Confidence {
        return s.buildResult(layer2Result, "layer2_low_conf"), nil
    }
    
    return s.buildResult(layer1Result, "layer1_low_conf"), nil
}
```

**Files to Modify:**
- `internal/classification/service.go`

### Phase 3 Success Metrics

**Before Phase 3:**
- Accuracy: 80-85%
- Handles: Common cases well
- Edge cases: Poor

**After Phase 3:**
- Accuracy: 85-90%
- Handles: Common cases + edge cases
- Embedding service: 95%+ uptime
- Vector search: <200ms

---

## Phase 4: Add Layer 3 - LLM (Weeks 7-8)
### Goal: Handle complex/ambiguous cases, generate rich explanations

For the remaining 10-15% of cases that are truly complex or novel, use an LLM.

### 4.1 Replace DistilBART with Qwen 7B

**Current:** DistilBART (not working well, not integrated)

**New:** Qwen/Qwen2.5-7B-Instruct (better reasoning, open source)

```python
# services/llm-service/app.py
from fastapi import FastAPI
from pydantic import BaseModel
from transformers import AutoTokenizer, AutoModelForCausalLM
import torch
import json
from typing import List, Dict

app = FastAPI()

# Load model on startup
MODEL_NAME = "Qwen/Qwen2.5-7B-Instruct"
tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
model = AutoModelForCausalLM.from_pretrained(
    MODEL_NAME,
    torch_dtype=torch.float16,
    device_map="auto"
)

class ClassifyRequest(BaseModel):
    content: Dict[str, str]  # title, description, about, etc.
    context: Dict  # hints from Layer 1/2

class ClassifyResponse(BaseModel):
    primary_industry: str
    mcc_codes: List[Dict]
    sic_codes: List[Dict]
    naics_codes: List[Dict]
    overall_confidence: float
    explanation: str
    evidence: Dict

@app.post("/classify")
async def classify_with_llm(request: ClassifyRequest):
    # Build prompt
    prompt = build_classification_prompt(request.content, request.context)
    
    # Generate
    inputs = tokenizer(prompt, return_tensors="pt").to(model.device)
    
    with torch.no_grad():
        outputs = model.generate(
            **inputs,
            max_new_tokens=1000,
            temperature=0.3,
            do_sample=True,
            top_p=0.9
        )
    
    response = tokenizer.decode(outputs[0], skip_special_tokens=True)
    
    # Parse JSON response
    classification = parse_llm_response(response)
    
    return ClassifyResponse(**classification)

def build_classification_prompt(content: Dict, context: Dict) -> str:
    prompt = f"""You are an expert business classifier analyzing a company's website.

Website Content:
Title: {content.get('title', '')}
Description: {content.get('meta_description', '')}
About: {content.get('about_text', '')[:500]}
Key Sections: {', '.join(content.get('headings', [])[:5])}

"""
    
    if context.get('layer1_suggestion'):
        prompt += f"\nInitial Analysis Suggestion: {context['layer1_suggestion']['industry']} (confidence: {context['layer1_suggestion']['confidence']})\n"
    
    if context.get('layer2_suggestion'):
        prompt += f"\nEmbedding Similarity Match: {context['layer2_suggestion']['top_match']} (similarity: {context['layer2_suggestion']['similarity']})\n"
    
    prompt += """
Based on this information, classify the business into the most specific industry possible.

Output a JSON object with this EXACT structure (no other text):
{
  "primary_industry": "Specific industry name",
  "mcc_codes": [
    {"code": "XXXX", "description": "Description", "confidence": 0.0-1.0}
  ],
  "sic_codes": [
    {"code": "XXXX", "description": "Description", "confidence": 0.0-1.0}
  ],
  "naics_codes": [
    {"code": "XXXXXX", "description": "Description", "confidence": 0.0-1.0}
  ],
  "overall_confidence": 0.0-1.0,
  "explanation": "Detailed reasoning for this classification, citing specific evidence from the website content",
  "evidence": {
    "key_indicators": ["list", "of", "specific", "indicators"],
    "business_model_signals": ["signals", "observed"]
  }
}

IMPORTANT: 
- Be specific with industry classification (avoid "General Business")
- Confidence should reflect actual certainty
- Explanation should cite specific evidence from the content
- Return 3 codes per type (MCC, SIC, NAICS) ranked by confidence

Respond ONLY with the JSON object. No other text.
"""
    
    return prompt

def parse_llm_response(response: str) -> dict:
    try:
        # Extract JSON
        json_start = response.find('{')
        json_end = response.rfind('}') + 1
        json_str = response[json_start:json_end]
        
        classification = json.loads(json_str)
        
        # Validate
        required = ['primary_industry', 'mcc_codes', 'sic_codes', 
                   'naics_codes', 'overall_confidence', 'explanation']
        
        for field in required:
            if field not in classification:
                raise ValueError(f"Missing: {field}")
        
        return classification
        
    except Exception as e:
        # Fallback
        return {
            'primary_industry': 'Unknown',
            'mcc_codes': [],
            'sic_codes': [],
            'naics_codes': [],
            'overall_confidence': 0.0,
            'explanation': f"LLM parsing error: {str(e)}",
            'evidence': {}
        }

@app.get("/health")
async def health():
    return {"status": "ok", "model": MODEL_NAME}
```

**Files to Create:**
- `services/llm-service/app.py`
- `services/llm-service/requirements.txt`
- `services/llm-service/Dockerfile`

**Railway Config:**
```yaml
[[services]]
name = "llm-service"
source = "services/llm-service"

[build]
builder = "DOCKERFILE"

[deploy]
healthcheckPath = "/health"
numReplicas = 1
restartPolicyType = "ON_FAILURE"

[[services.resourceLimits]]
memory = 8192  # 8GB for 7B model
cpuCores = 2
```

### 4.2 Integrate LLM Layer into Go Service

```go
// internal/classification/llm_classifier.go

package classification

type LLMClassifier struct {
    llmServiceURL string
    httpClient    *http.Client
}

type LLMResult struct {
    PrimaryIndustry string
    MCC             []CodeResult
    SIC             []CodeResult
    NAICS           []CodeResult
    Confidence      float64
    Explanation     string
    Evidence        map[string]interface{}
    Method          string
}

func (l *LLMClassifier) ClassifyWithLLM(
    ctx context.Context,
    content *ScrapedContent,
    layer1Hint *ClassificationResult,
    layer2Hint *EmbeddingResult,
) (*LLMResult, error) {
    
    // Prepare request
    reqBody := map[string]interface{}{
        "content": map[string]interface{}{
            "title":            content.Title,
            "meta_description": content.MetaDesc,
            "about_text":       content.AboutText,
            "headings":         content.Headings,
        },
        "context": map[string]interface{}{
            "layer1_suggestion": map[string]interface{}{
                "industry":   layer1Hint.Industry,
                "confidence": layer1Hint.Confidence,
            },
            "layer2_suggestion": map[string]interface{}{
                "top_match":  layer2Hint.MCC[0].Description,
                "similarity": layer2Hint.MCC[0].Confidence,
            },
        },
    }
    
    reqBodyJSON, _ := json.Marshal(reqBody)
    
    // Call LLM service
    resp, err := l.httpClient.Post(
        l.llmServiceURL + "/classify",
        "application/json",
        bytes.NewReader(reqBodyJSON),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result LLMResult
    json.NewDecoder(resp.Body).Decode(&result)
    result.Method = "llm"
    
    return &result, nil
}
```

**Files to Create:**
- `internal/classification/llm_classifier.go`

### 4.3 Complete 3-Layer Orchestration

```go
// internal/classification/service.go

func (s *IndustryDetectionService) DetectIndustry(
    ctx context.Context,
    businessName, description, websiteURL string,
) (*IndustryDetectionResult, error) {
    
    startTime := time.Now()
    
    // Scrape website
    content, err := s.scraper.Scrape(websiteURL)
    if err != nil {
        return nil, fmt.Errorf("scraping failed: %w", err)
    }
    
    // Layer 1: Multi-strategy (fast + full)
    layer1Result, err := s.multiStrategyClassifier.Classify(ctx, content)
    
    if layer1Result.Confidence >= 0.90 {
        // Very high confidence - done
        result := s.buildResult(layer1Result, "layer1_fast")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    // Layer 2: Embedding-based
    layer2Result, err := s.embeddingClassifier.ClassifyByEmbedding(ctx, content)
    
    if layer2Result.Confidence >= 0.85 {
        // Good embedding match - done
        result := s.buildResult(layer2Result, "layer2")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    // Layer 3: LLM (for complex/ambiguous cases)
    llmResult, err := s.llmClassifier.ClassifyWithLLM(
        ctx,
        content,
        layer1Result,
        layer2Result,
    )
    
    if err != nil {
        // LLM failed - use best available
        if layer2Result.Confidence > layer1Result.Confidence {
            result := s.buildResult(layer2Result, "layer2_fallback")
            result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
            return result, nil
        }
        result := s.buildResult(layer1Result, "layer1_fallback")
        result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        return result, nil
    }
    
    // Use LLM result
    result := s.buildResultFromLLM(llmResult, "layer3_llm")
    result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
    
    return result, nil
}
```

**Files to Modify:**
- `internal/classification/service.go`

### Phase 4 Success Metrics

**Before Phase 4:**
- Accuracy: 85-90%
- Complex cases: Poor
- Explanations: Basic

**After Phase 4:**
- Accuracy: 90-95%
- Complex cases: Handled well
- Explanations: Rich, detailed, auditable
- LLM layer: Handles 10-15% of requests

---

## Phase 5: UI Integration & Testing (Week 9)
### Goal: Show explanations in UI, comprehensive testing

### 5.1 Add Explanation to Frontend

```go
// Update response structure to include explanation

type ClassificationResponse struct {
    RequestID   string                 `json:"request_id"`
    Classification ClassificationData  `json:"classification"`
    Codes        ClassificationCodes   `json:"codes"`
    Explanation  ClassificationExplanation `json:"explanation"` // NEW
}

type ClassificationExplanation struct {
    PrimaryReason     string              `json:"primary_reason"`
    SupportingFactors []string            `json:"supporting_factors"`
    KeyTermsFound     []string            `json:"key_terms_found"`
    MethodUsed        string              `json:"method_used"`
    LayerUsed         string              `json:"layer_used"` // "layer1", "layer2", "layer3"
}
```

**Frontend Update (wherever your UI code is):**
```javascript
// Display explanation in UI
<div class="explanation-section">
  <h3>Classification Reasoning</h3>
  <p class="primary-reason">{explanation.primary_reason}</p>
  
  <div class="supporting-factors">
    <h4>Supporting Evidence:</h4>
    <ul>
      {explanation.supporting_factors.map(factor => (
        <li>{factor}</li>
      ))}
    </ul>
  </div>
  
  <div class="key-terms">
    <strong>Key terms found:</strong> {explanation.key_terms_found.join(', ')}
  </div>
  
  <div class="method-badge">
    Method: {explanation.method_used} | Layer: {explanation.layer_used}
  </div>
</div>
```

**Files to Modify:**
- Frontend template files (wherever your UI renders classification results)
- `services/classification-service/internal/handlers/classification.go` (ensure explanation is included)

### 5.2 Comprehensive Testing

**Create Test Suite:**

```go
// internal/classification/service_test.go

func TestClassificationAccuracy(t *testing.T) {
    testCases := []struct {
        name           string
        websiteURL     string
        expectedIndustry string
        expectedMCC    string
    }{
        {
            name:           "Restaurant - Clear Case",
            websiteURL:     "https://mcdonalds.com",
            expectedIndustry: "Quick Service Restaurant",
            expectedMCC:    "5814",
        },
        {
            name:           "Software Company",
            websiteURL:     "https://github.com",
            expectedIndustry: "Software Development",
            expectedMCC:    "5734",
        },
        // Add 50-100 test cases
    }
    
    service := setupTestService()
    
    correctCount := 0
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := service.DetectIndustry(
                context.Background(),
                "",
                "",
                tc.websiteURL,
            )
            
            assert.NoError(t, err)
            
            if result.Classification.PrimaryIndustry == tc.expectedIndustry {
                correctCount++
            }
            
            // Also check if expected MCC is in top 3
            hasMCC := false
            for _, code := range result.Codes.MCC {
                if code.Code == tc.expectedMCC {
                    hasMCC = true
                    break
                }
            }
            
            assert.True(t, hasMCC, "Expected MCC %s in top 3", tc.expectedMCC)
        })
    }
    
    accuracy := float64(correctCount) / float64(len(testCases))
    t.Logf("Overall accuracy: %.2f%%", accuracy*100)
    
    assert.GreaterOrEqual(t, accuracy, 0.85, "Accuracy should be ≥85%")
}
```

**Files to Create:**
- `internal/classification/service_test.go`
- `test_data/classification_test_cases.json` (list of test URLs + expected results)

### 5.3 Performance Benchmarking

```go
// internal/classification/benchmark_test.go

func BenchmarkLayer1FastPath(b *testing.B) {
    service := setupTestService()
    content := loadTestContent("restaurant_obvious.html")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.multiStrategyClassifier.ClassifyFastPath(context.Background(), content)
    }
}

func BenchmarkLayer1FullStrategy(b *testing.B) {
    service := setupTestService()
    content := loadTestContent("software_company.html")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.multiStrategyClassifier.ClassifyWithMultiStrategy(context.Background(), "", "", "")
    }
}

func BenchmarkLayer2Embedding(b *testing.B) {
    service := setupTestService()
    content := loadTestContent("ambiguous_business.html")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.embeddingClassifier.ClassifyByEmbedding(context.Background(), content)
    }
}

// Run: go test -bench=. -benchmem
```

**Target Benchmarks:**
- Layer 1 Fast Path: <100ms (p95)
- Layer 1 Full: <500ms (p95)
- Layer 2 Embedding: <800ms (p95)
- Layer 3 LLM: <5s (p95)

### 5.4 Add Classification Cache Table

```sql
-- supabase-migrations/060_add_classification_cache.sql

CREATE TABLE classification_cache (
    id BIGSERIAL PRIMARY KEY,
    domain VARCHAR(255) NOT NULL,
    content_hash VARCHAR(64) NOT NULL, -- SHA-256 of content
    classification JSONB NOT NULL,
    confidence DECIMAL(3,2),
    method VARCHAR(20),
    layer VARCHAR(10), -- 'layer1', 'layer2', 'layer3'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    hit_count INTEGER DEFAULT 1,
    UNIQUE(domain, content_hash)
);

CREATE INDEX idx_cache_lookup ON classification_cache(domain, content_hash);
CREATE INDEX idx_cache_expiry ON classification_cache(expires_at);
CREATE INDEX idx_cache_layer ON classification_cache(layer);

-- Function to increment hit count
CREATE OR REPLACE FUNCTION increment_cache_hit(
    p_domain VARCHAR,
    p_content_hash VARCHAR
)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE classification_cache
    SET hit_count = hit_count + 1,
        expires_at = NOW() + INTERVAL '30 days' -- Reset TTL on hit
    WHERE domain = p_domain 
        AND content_hash = p_content_hash;
END;
$$;
```

**Integrate into Go service:**

```go
// internal/classification/repository/supabase_repository.go

func (r *SupabaseKeywordRepository) GetCachedClassification(
    ctx context.Context,
    domain string,
    contentHash string,
) (*ClassificationResult, error) {
    
    var result struct {
        Classification json.RawMessage `json:"classification"`
    }
    
    err := r.client.
        From("classification_cache").
        Select("classification").
        Eq("domain", domain).
        Eq("content_hash", contentHash).
        Gte("expires_at", time.Now()).
        Single().
        Execute(&result)
    
    if err != nil {
        return nil, err // Cache miss
    }
    
    // Increment hit count (async)
    go r.incrementCacheHit(domain, contentHash)
    
    var classification ClassificationResult
    json.Unmarshal(result.Classification, &classification)
    
    return &classification, nil
}

func (r *SupabaseKeywordRepository) CacheClassification(
    ctx context.Context,
    domain string,
    contentHash string,
    classification *ClassificationResult,
) error {
    
    expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
    
    _, err := r.client.
        From("classification_cache").
        Insert(map[string]interface{}{
            "domain":         domain,
            "content_hash":   contentHash,
            "classification": classification,
            "confidence":     classification.Confidence,
            "method":         classification.Method,
            "layer":          classification.Layer,
            "expires_at":     expiresAt,
        }).
        Execute()
    
    return err
}
```

**Files to Create:**
- `supabase-migrations/060_add_classification_cache.sql`

**Files to Modify:**
- `internal/classification/repository/supabase_repository.go`
- `internal/classification/service.go` (check cache before classifying)

### 5.5 Monitoring & Analytics

Add dashboard to track:
- Classification accuracy by layer
- Layer usage distribution (% Layer 1/2/3)
- Average latency by layer
- Cache hit rate
- Scraping success rate

```sql
-- supabase-migrations/061_add_analytics_tables.sql

CREATE TABLE classification_metrics (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    layer VARCHAR(10), -- 'layer1', 'layer2', 'layer3'
    method VARCHAR(30),
    confidence DECIMAL(3,2),
    processing_time_ms INTEGER,
    cache_hit BOOLEAN,
    scrape_success BOOLEAN,
    accuracy_validated BOOLEAN DEFAULT FALSE,
    was_correct BOOLEAN
);

CREATE INDEX idx_metrics_timestamp ON classification_metrics(timestamp);
CREATE INDEX idx_metrics_layer ON classification_metrics(layer);
```

**Simple Analytics Query:**
```sql
-- Layer usage over last 7 days
SELECT 
    layer,
    COUNT(*) as count,
    AVG(confidence) as avg_confidence,
    AVG(processing_time_ms) as avg_latency_ms,
    SUM(CASE WHEN cache_hit THEN 1 ELSE 0 END)::FLOAT / COUNT(*) as cache_hit_rate
FROM classification_metrics
WHERE timestamp > NOW() - INTERVAL '7 days'
GROUP BY layer;
```

### Phase 5 Success Metrics

**After Phase 5:**
- UI shows full explanation
- Comprehensive test coverage (50+ test cases)
- Accuracy validated: ≥90%
- Performance benchmarks met
- 30-day cache enabled
- Monitoring dashboard live

---

## Final Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Railway Platform                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐         ┌────────────────────────────┐   │
│  │  Frontend    │────────▶│  Classification API        │   │
│  │  (Go + UI)   │         │  (Go Service)              │   │
│  └──────────────┘         └────────┬───────────────────┘   │
│                                     │                        │
│                                     ▼                        │
│                          ┌──────────────────────┐           │
│                          │  Orchestrator        │           │
│                          │  (3-Layer Router)    │           │
│                          └──┬────────┬──────┬───┘           │
│                             │        │      │                │
│                   ┌─────────┘        │      └──────────┐    │
│                   ▼                  ▼                 ▼     │
│         ┌─────────────────┐  ┌─────────────┐  ┌────────────┐│
│         │  Layer 1        │  │  Layer 2    │  │  Layer 3   ││
│         │  Multi-Strategy │  │  Embedding  │  │  LLM       ││
│         │  (Go)           │  │  (Python)   │  │  (Python)  ││
│         │  <100ms         │  │  <800ms     │  │  <5s       ││
│         │  70% of reqs    │  │  20% of reqs│  │  10% reqs  ││
│         └────────┬────────┘  └──────┬──────┘  └─────┬──────┘│
│                  │                  │                │       │
│                  └──────────────────┴────────────────┘       │
│                                     │                        │
│  ┌──────────────────────────────────▼──────────────────┐    │
│  │           Supabase PostgreSQL                       │    │
│  │  ├─ industries, codes, keywords (existing)         │    │
│  │  ├─ code_embeddings (NEW - pgvector)              │    │
│  │  ├─ classification_cache (NEW - 30 day)           │    │
│  │  ├─ classification_metrics (NEW - analytics)      │    │
│  │  └─ trigram indexes, full-text search             │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌──────────────────┐                                       │
│  │  Redis Cache     │  (Short-term, <5 min TTL)            │
│  └──────────────────┘                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## Implementation Timeline Summary

| Phase | Duration | Focus | Expected Accuracy | Key Deliverable |
|-------|----------|-------|-------------------|-----------------|
| **Phase 1** | 2 weeks | Fix scraping | 50-60% | Reliable content extraction |
| **Phase 2** | 2 weeks | Enhance Layer 1 | 80-85% | Top 3 codes, better confidence |
| **Phase 3** | 2 weeks | Add Layer 2 | 85-90% | Embedding search working |
| **Phase 4** | 2 weeks | Add Layer 3 | 90-95% | LLM for complex cases |
| **Phase 5** | 1 week | UI + Testing | Validated 90%+ | Production-ready |
| **Total** | **9 weeks** | | **90-95%** | Full hybrid system |

---

## Week-by-Week Checklist

### Week 1-2: Phase 1 (Foundation)
- [ ] Add comprehensive logging to scraper
- [ ] Enhance content extraction (title, headings, about, etc.)
- [ ] Implement scraper fallback strategies
- [ ] Add content quality validation
- [ ] Test scraping on 50+ diverse websites
- [ ] Achieve 95%+ scrape success rate

### Week 3-4: Phase 2 (Layer 1)
- [ ] Return top 3 codes per type
- [ ] Improve confidence calibration
- [ ] Implement fast path (<100ms)
- [ ] Optimize database queries
- [ ] Generate structured explanations
- [ ] Fix "General Business" fallback
- [ ] Benchmark performance

### Week 5-6: Phase 3 (Layer 2)
- [ ] Enable pgvector in Supabase
- [ ] Pre-compute code embeddings
- [ ] Deploy embedding service (Python)
- [ ] Add vector search RPC function
- [ ] Integrate embedding layer in Go
- [ ] Add layer routing logic
- [ ] Test embedding accuracy

### Week 7-8: Phase 4 (Layer 3)
- [ ] Deploy LLM service (Qwen 7B)
- [ ] Build structured prompts
- [ ] Integrate LLM classifier in Go
- [ ] Complete 3-layer orchestration
- [ ] Test LLM on complex cases
- [ ] Optimize LLM performance

### Week 9: Phase 5 (UI & Testing)
- [ ] Add explanation to UI
- [ ] Create comprehensive test suite (50+ cases)
- [ ] Run performance benchmarks
- [ ] Add classification cache table
- [ ] Deploy monitoring dashboard
- [ ] Final accuracy validation

---

## Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Scraping still fails | Low | High | Multiple fallback strategies, Playwright option |
| LLM too slow (>5s) | Medium | Medium | Use smaller model (3B), optimize prompts, GPU on Railway |
| Embedding service crashes | Low | Medium | Auto-restart, fallback to Layer 1 |
| Accuracy still low | Low | High | Iterative testing, feedback loop, continuous improvement |
| Railway costs too high | Low | Medium | Optimize layer distribution, cache aggressively |

---

## Cost Optimization

**Current (Pure LLM - not integrated):**
- Unknown, likely $0.02-0.05 per classification if using API

**After Implementation:**
- Layer 1 (70% of requests): ~$0.0001 per classification (database queries only)
- Layer 2 (20% of requests): ~$0.001 per classification (embedding service)
- Layer 3 (10% of requests): ~$0.005 per classification (self-hosted LLM)
- **Average: ~$0.0007 per classification**

**Monthly cost (10,000 classifications):**
- ~$7/month for compute (Railway)
- ~$2/month for Supabase (free tier likely sufficient)
- **Total: ~$10/month**

---

## Success Criteria

**Phase 1 Complete:**
✅ Scrape success rate ≥95%
✅ Content quality score ≥0.7
✅ Zero "no output" errors

**Phase 2 Complete:**
✅ Returns top 3 codes per type
✅ Confidence calibrated (70-95%)
✅ Layer 1 p95 latency <300ms
✅ Accuracy 80-85%

**Phase 3 Complete:**
✅ Embedding service deployed
✅ Vector search <200ms
✅ Accuracy 85-90%

**Phase 4 Complete:**
✅ LLM service deployed
✅ 3-layer routing working
✅ Rich explanations generated
✅ Accuracy 90-95%

**Phase 5 Complete:**
✅ Explanation shown in UI
✅ Test suite passing (50+ cases)
✅ Monitoring dashboard live
✅ Production-ready

---

## Next Steps

1. **Review this plan** - Any questions or concerns?
2. **Start Phase 1** - Fix scraping immediately (this is your biggest bottleneck)
3. **Set up logging** - Visibility into what's failing
4. **Test on 20 diverse websites** - Understand failure modes
5. **Fix content extraction** - Get quality input for classifier

**Week 1 Priority:**
Just get scraping to 95%+ success with good content. Everything else depends on this.

Want me to create specific code files for Phase 1 to get you started immediately?
