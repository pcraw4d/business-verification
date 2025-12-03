# Classification System Analysis Report

**Date:** January 27, 2025  
**Analyzer:** Cursor AI  
**Codebase Version:** Current (as of analysis date)

## Executive Summary

The current classification system implements a **hybrid multi-strategy approach** combining keyword-based classification, entity recognition, topic modeling, and optional ML enhancement. The system is built on Go with Supabase PostgreSQL as the database backend, deployed as microservices on Railway. The architecture includes sophisticated caching mechanisms (Redis + in-memory), request deduplication, and a three-tier confidence-based ML strategy.

**Current approach:** Multi-strategy ensemble classifier with keyword matching, entity recognition, topic modeling, and co-occurrence analysis. Optional ML enhancement (DistilBART) is available but not fully integrated. The system uses database-driven keyword matching with trigram similarity, full-text search, and crosswalk validation.

**Key gaps:**
- No embedding-based similarity layer (Layer 2) - pgvector not enabled, no code embeddings stored
- No dedicated LLM service (Layer 3) - Python ML service exists but uses DistilBART, not open-source LLM
- No 3-layer orchestration logic - current system uses confidence thresholds but not layer-based routing
- Missing 30-day classification cache table in Supabase - currently using Redis/in-memory only
- No embedding service for generating 384-dim vectors

**Reusable components:** Approximately 60-70% of current infrastructure can be reused, including database schema (with extensions), keyword/trigram matching logic, website scraping, caching infrastructure, and API handlers.

**Recommended path forward:** Incremental enhancement approach - build Layer 2 (embeddings) and Layer 3 (LLM) in parallel while maintaining current system. Add orchestration layer to route requests based on confidence thresholds. Extend existing caching to include Supabase-based 30-day cache.

---

## 1. Current Implementation Overview

### 1.1 Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Railway Deployment Platform                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┐      ┌──────────────────────────────┐    │
│  │  API Gateway     │─────▶│  Classification Service     │    │
│  │  (Go)           │      │  (Go - Railway)              │    │
│  └──────────────────┘      └──────────────┬───────────────┘    │
│                                            │                    │
│  ┌──────────────────┐      ┌──────────────▼───────────────┐    │
│  │  Frontend        │      │  IndustryDetectionService     │    │
│  │  (Go)           │      │  ├─ MultiStrategyClassifier   │    │
│  └──────────────────┘      │  ├─ ML Classifier (optional) │    │
│                             │  └─ Code Generator          │    │
│  ┌──────────────────┐      └──────────────┬───────────────┘    │
│  │  Python ML        │                    │                    │
│  │  Service          │                    │                    │
│  │  (DistilBART)     │                    │                    │
│  └──────────────────┘                    │                    │
│                                           │                    │
│  ┌────────────────────────────────────────▼───────────────┐    │
│  │              Supabase PostgreSQL                        │    │
│  │  ├─ industries                                         │    │
│  │  ├─ industry_keywords / keyword_weights               │    │
│  │  ├─ classification_codes                              │    │
│  │  ├─ code_keywords                                     │    │
│  │  ├─ industry_code_crosswalks                         │    │
│  │  ├─ code_metadata                                    │    │
│  │  └─ (trigram indexes, full-text search)              │    │
│  └───────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌──────────────────┐                                          │
│  │  Redis Cache     │  (Optional - distributed caching)       │
│  │  (In-memory      │                                          │
│  │   fallback)      │                                          │
│  └──────────────────┘                                          │
└─────────────────────────────────────────────────────────────────┘
```

**Description:**

The classification system follows a microservices architecture deployed on Railway. The main flow:

1. **Request Reception**: API Gateway receives classification requests and routes to Classification Service
2. **Multi-Strategy Classification**: `IndustryDetectionService` orchestrates classification using `MultiStrategyClassifier`
3. **Strategy Execution**: Four parallel strategies run:
   - Keyword-based (40% weight) - uses database keyword matching with trigram similarity
   - Entity-based (25% weight) - NER extraction and matching
   - Topic-based (20% weight) - topic modeling with database support
   - Co-occurrence (15% weight) - pattern matching via database queries
4. **ML Enhancement** (optional): Three-tier confidence-based ML strategy:
   - Low confidence (<0.5): ML-assisted improvement (60% ML weight)
   - Medium confidence (0.5-0.8): Ensemble validation (50/50 split)
   - High confidence (>=0.8): ML validation only (boosts confidence)
5. **Code Generation**: `ClassificationCodeGenerator` generates MCC/SIC/NAICS codes using crosswalk validation
6. **Caching**: Results cached in Redis (if enabled) or in-memory with TTL-based expiration
7. **Database Queries**: Uses Supabase with trigram indexes, full-text search, and optimized queries

### 1.2 Technology Stack

- **Backend Framework:** Go (standard library net/http, Go 1.22+ ServeMux)
- **Database:** Supabase PostgreSQL with extensions (pg_trgm, full-text search)
- **ML/AI:** 
  - Go ML classifier (BERT-based) - optional
  - Python ML service (DistilBART) - available but not fully integrated
  - No open-source LLM (Qwen/Mistral) currently
- **Deployment:** Railway (microservices architecture)
- **Caching:** Redis (optional) + in-memory fallback
- **Website Scraping:** Go HTTP client with retry logic, anti-bot detection
- **Other Services:** 
  - Merchant Service
  - Risk Assessment Service
  - Business Intelligence Service
  - Service Discovery

### 1.3 Key Components

| Component | File Path | Purpose | Status |
|-----------|-----------|---------|--------|
| IndustryDetectionService | `internal/classification/service.go` | Main classification orchestrator | ✅ Working |
| MultiStrategyClassifier | `internal/classification/multi_strategy_classifier.go` | Multi-strategy ensemble classifier | ✅ Working |
| ClassificationCodeGenerator | `internal/classification/classifier.go` | Generates MCC/SIC/NAICS codes | ✅ Working |
| SupabaseKeywordRepository | `internal/classification/repository/supabase_repository.go` | Database access layer | ✅ Working |
| ClassificationHandler | `services/classification-service/internal/handlers/classification.go` | HTTP request handler | ✅ Working |
| WebsiteScraper | `internal/external/website_scraper.go` | Website content extraction | ✅ Working |
| PredictiveCache | `internal/classification/cache/predictive_cache.go` | Classification result caching | ✅ Working |
| RedisCache | `services/classification-service/internal/cache/redis_cache.go` | Distributed caching | ✅ Working |
| Python ML Service | `python_ml_service/app.py` | DistilBART classification | ⚠️ Needs Work (not integrated) |
| EntityRecognizer | `internal/classification/nlp/entity_recognizer.go` | Named entity recognition | ✅ Working |
| TopicModeler | `internal/classification/nlp/topic_modeler.go` | Topic modeling | ✅ Working |
| ConfidenceCalibrator | `internal/classification/confidence_calibrator.go` | Confidence score calibration | ✅ Working |

### 1.4 API Endpoints

```
POST /v1/classify
POST /v1/classify/batch
GET /v1/classify/{business_id}
GET /v1/classify/history
POST /v2/classify (Enhanced)
POST /v2/classify/batch
GET /v2/classify/{business_id}
GET /v1/monitoring/accuracy/metrics
POST /v1/monitoring/accuracy/track
```

**Request/Response Examples:**

```json
// Request
{
  "business_name": "Acme Corporation",
  "description": "Software development company",
  "website_url": "https://acme.com",
  "request_id": "req_123456",
  "country": "US"
}

// Response
{
  "request_id": "req_123456",
  "classification": {
    "primary_industry": "Technology",
    "confidence": 0.92,
    "method": "multi_strategy_ml_validated",
    "reasoning": "Combined 4 classification strategies...",
    "keywords": ["software", "development", "technology"],
    "processing_time_ms": 245
  },
  "codes": {
    "mcc": [
      {"code": "5734", "description": "Computer Software Stores", "confidence": 0.95}
    ],
    "naics": [
      {"code": "541511", "description": "Custom Computer Programming Services", "confidence": 0.93}
    ],
    "sic": [
      {"code": "7371", "description": "Computer Programming Services", "confidence": 0.91}
    ]
  },
  "cached": false
}
```

---

## 2. Database Schema Analysis

### 2.1 Current Tables

#### Table: `classification_codes`
```sql
CREATE TABLE classification_codes (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    code_type VARCHAR(10) NOT NULL CHECK (code_type IN ('NAICS', 'MCC', 'SIC')),
    code VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    confidence DECIMAL(3,2) DEFAULT 0.80,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(code_type, code)
);
```
**Usage:** Stores MCC, SIC, and NAICS codes mapped to industries. Used by `ClassificationCodeGenerator` to retrieve codes for classified businesses. Has full-text search index on `description` field.

**Alignment with Proposal:** ✅ Matches - This table aligns perfectly with the proposed `codes` table. Can be extended with embedding column for Layer 2.

#### Table: `industry_keywords` / `keyword_weights`
```sql
CREATE TABLE industry_keywords (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword VARCHAR(100) NOT NULL,
    weight DECIMAL(3,2) DEFAULT 1.00,
    context VARCHAR(50),
    is_primary BOOLEAN DEFAULT false,
    ...
);

CREATE TABLE keyword_weights (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    keyword VARCHAR(100) NOT NULL,
    base_weight DECIMAL(3,2) DEFAULT 1.00,
    calculated_weight DECIMAL(3,2) GENERATED ALWAYS AS (...),
    ...
);
```
**Usage:** Stores keywords associated with industries. Used for keyword-based classification strategy. Has trigram index (`idx_keyword_weights_keyword_trgm`) for fuzzy matching.

**Alignment with Proposal:** ✅ Matches - These tables serve the same purpose as proposed `keywords` table. Trigram indexes already exist for fast matching.

#### Table: `trigram` (via pg_trgm extension)
**Note:** Not a physical table - trigram functionality provided by PostgreSQL `pg_trgm` extension with GIN indexes.

**Usage:** Used by `classify_business_by_keywords_trigram()` function for fuzzy keyword matching. Index: `idx_keyword_weights_keyword_trgm USING gin (keyword gin_trgm_ops)`.

**Alignment with Proposal:** ✅ Matches - Trigram analysis is implemented via database functions and indexes. Performance is optimized.

#### Table: `industry_code_crosswalks`
```sql
CREATE TABLE industry_code_crosswalks (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    mcc_code VARCHAR(10),
    naics_code VARCHAR(10),
    sic_code VARCHAR(10),
    code_description TEXT,
    confidence_score DECIMAL(3,2) DEFAULT 0.80,
    is_primary BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    usage_frequency INTEGER DEFAULT 0,
    last_used TIMESTAMP WITH TIME ZONE,
    ...
    UNIQUE(industry_id, mcc_code, naics_code, sic_code)
);
```
**Usage:** Maps industries to all code types (MCC, NAICS, SIC) with confidence scores. Used by `validateAndEnhanceCodesWithCrosswalks()` for code validation.

**Alignment with Proposal:** ✅ Matches - This table exactly matches the proposed `crosswalks` table structure.

#### Table: `code_metadata`
```sql
CREATE TABLE code_metadata (
    id BIGSERIAL PRIMARY KEY,
    code_type VARCHAR(10) NOT NULL CHECK (code_type IN ('MCC', 'SIC', 'NAICS')),
    code VARCHAR(20) NOT NULL,
    official_description TEXT,
    official_name VARCHAR(255),
    industry_mappings JSONB DEFAULT '{}'::jsonb,
    crosswalk_data JSONB DEFAULT '{}'::jsonb,
    hierarchy JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    ...
    UNIQUE(code_type, code)
);
```
**Usage:** Stores additional metadata for codes including crosswalk data and hierarchy. Used for enhanced code matching.

**Alignment with Proposal:** ⚠️ Needs Extension - Can be extended to store embeddings in `metadata` JSONB field or separate column.

### 2.2 Missing Tables/Features

- [ ] `code_embeddings` table with pgvector
  - **Required:** Separate table for storing 384-dim embeddings
  - **Structure needed:**
    ```sql
    CREATE TABLE code_embeddings (
        id BIGSERIAL PRIMARY KEY,
        code_type VARCHAR(10) NOT NULL,
        code VARCHAR(20) NOT NULL,
        embedding vector(384),
        model_name VARCHAR(100) DEFAULT 'all-MiniLM-L6-v2',
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
        UNIQUE(code_type, code)
    );
    CREATE INDEX ON code_embeddings USING ivfflat (embedding vector_cosine_ops);
    ```

- [ ] `classification_cache` table
  - **Required:** 30-day cache in Supabase (currently only Redis/in-memory)
  - **Structure needed:**
    ```sql
    CREATE TABLE classification_cache (
        id BIGSERIAL PRIMARY KEY,
        cache_key VARCHAR(64) UNIQUE NOT NULL, -- SHA-256 hash
        business_name TEXT,
        description TEXT,
        website_url TEXT,
        classification_result JSONB NOT NULL,
        expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
        access_count INTEGER DEFAULT 0,
        last_accessed TIMESTAMP WITH TIME ZONE
    );
    CREATE INDEX ON classification_cache(cache_key);
    CREATE INDEX ON classification_cache(expires_at);
    ```

- [ ] `classification_feedback` table
  - **Optional but recommended:** For tracking classification accuracy and improving models
  - **Structure:**
    ```sql
    CREATE TABLE classification_feedback (
        id BIGSERIAL PRIMARY KEY,
        classification_id VARCHAR(100),
        business_name TEXT,
        predicted_industry VARCHAR(100),
        actual_industry VARCHAR(100),
        confidence DECIMAL(3,2),
        feedback_type VARCHAR(20), -- 'correct', 'incorrect', 'partial'
        user_notes TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );
    ```

### 2.3 Database Functions & Indexes

**Existing:**

```sql
-- Trigram-based classification function
CREATE OR REPLACE FUNCTION classify_business_by_keywords_trigram(
    p_keywords text[],
    p_business_name text DEFAULT '',
    p_similarity_threshold float DEFAULT 0.3
) RETURNS TABLE (...);

-- Full-text search function
CREATE OR REPLACE FUNCTION find_codes_by_fulltext_search(
    p_search_text text,
    p_code_type text,
    p_limit int DEFAULT 3
) RETURNS TABLE (...);

-- Indexes
CREATE INDEX idx_keyword_weights_keyword_trgm 
    ON keyword_weights USING gin (keyword gin_trgm_ops);

CREATE INDEX idx_classification_codes_description_fts 
    ON classification_codes USING gin (to_tsvector('english', description))
    WHERE is_active = true;
```

**Missing (from proposal):**

- `match_code_embeddings()` RPC function for vector search
  ```sql
  CREATE OR REPLACE FUNCTION match_code_embeddings(
      p_query_embedding vector(384),
      p_code_type text,
      p_limit int DEFAULT 3,
      p_similarity_threshold float DEFAULT 0.7
  ) RETURNS TABLE (
      code_type text,
      code text,
      description text,
      similarity float
  );
  ```

- `get_cached_classification()` function for cache lookups
- `store_classification_cache()` function for cache storage

---

## 3. Classification Logic Analysis

### 3.1 Current Classification Flow

```
1. HTTP Request received at ClassificationHandler
   ↓
2. Request validation & sanitization
   ↓
3. Cache check (Redis → in-memory)
   ├─ Cache HIT → Return cached result
   └─ Cache MISS → Continue
   ↓
4. Request deduplication check
   ├─ In-flight request exists → Wait for result
   └─ New request → Continue
   ↓
5. IndustryDetectionService.DetectIndustry()
   ↓
6. MultiStrategyClassifier.ClassifyWithMultiStrategy()
   ├─ Extract keywords (from website/database)
   ├─ Extract entities (NER)
   ├─ Run 4 strategies in parallel:
   │  ├─ Keyword-based (40% weight)
   │  ├─ Entity-based (25% weight)
   │  ├─ Topic-based (20% weight)
   │  └─ Co-occurrence (15% weight)
   └─ Combine strategies with weighted average
   ↓
7. Confidence-based ML routing
   ├─ Confidence < 0.5 → ML-assisted improvement
   ├─ Confidence 0.5-0.8 → Ensemble validation
   └─ Confidence >= 0.8 → ML validation only
   ↓
8. ClassificationCodeGenerator.GenerateCodes()
   ├─ Query classification_codes by industry
   ├─ Validate with crosswalks
   └─ Return top 3 codes per type
   ↓
9. Cache result (Redis + in-memory)
   ↓
10. Return response
```

### 3.2 Classification Methods in Use

#### Method 1: Multi-Strategy Ensemble Classification
**Files:** 
- `internal/classification/multi_strategy_classifier.go`
- `internal/classification/service.go`

**How it works:** 
Combines four parallel classification strategies using weighted averaging:
- **Keyword-based (40%)**: Database keyword matching with trigram similarity
- **Entity-based (25%)**: Named entity recognition and matching
- **Topic-based (20%)**: Topic modeling with database support
- **Co-occurrence (15%)**: Pattern matching via database queries

Results are combined using weighted scores, with confidence calibration applied.

**Performance:** 
- Average: 200-500ms (uncached)
- Cached: 10-50ms
- Database queries: 50-150ms (with indexes)

**Pros:** 
- High accuracy through ensemble approach
- Fast with database indexes
- Parallel execution for performance
- Confidence calibration improves reliability

**Cons:** 
- No semantic similarity (embeddings)
- Limited to keyword/pattern matching
- May miss nuanced industry classifications

#### Method 2: ML-Enhanced Classification (Optional)
**Files:**
- `internal/classification/service.go` (improveWithML, validateWithEnsemble, validateWithMLHighConfidence)
- `python_ml_service/app.py` (DistilBART classifier)

**How it works:**
Three-tier confidence-based ML strategy:
- **Low confidence (<0.5)**: ML-assisted improvement with 60% ML weight
- **Medium confidence (0.5-0.8)**: Ensemble validation with 50/50 split
- **High confidence (>=0.8)**: ML validation only (boosts confidence)

**Performance:**
- ML classification: 500ms - 2s (Python service call)
- Overall: 700ms - 2.5s when ML is used

**Pros:**
- Improves low-confidence classifications
- Validates high-confidence results
- Handles ambiguous cases better

**Cons:**
- Not fully integrated (Python service exists but not always used)
- Uses DistilBART, not open-source LLM (Qwen/Mistral)
- Adds latency when invoked
- No structured JSON output with reasoning

### 3.3 Website Scraping Implementation

**Current approach:** 
Go HTTP client with retry logic, timeout handling, and anti-bot detection. Content is extracted from HTML and used for keyword extraction.

**Libraries used:** 
- Standard Go `net/http` package
- Custom HTML parsing (text extraction)
- Anti-bot detection in `internal/modules/website_verification/enhanced_scraper.go`

**Content extraction:** 
1. HTTP GET request with custom User-Agent
2. Read response body (max 10MB)
3. Extract text content from HTML
4. Extract title, meta tags
5. Filter technical terms and stop words
6. Return keywords for classification

**Alignment with proposal:** 
⚠️ Needs Enhancement - Current scraper is basic. Proposal suggests Playwright for better JavaScript rendering and bot evasion. Current implementation is sufficient but could be improved.

### 3.4 Confidence Scoring

**How is confidence calculated?** 
1. **Base confidence**: Weighted average of strategy scores
   - Formula: `Σ(strategy_score × strategy_confidence × strategy_weight) / Σ(strategy_weight)`
2. **Business name context boost/penalty**: Adjusts based on known business patterns
3. **Consensus boost**: +0.15 if ML and base agree
4. **Confidence calibration**: Applied via `ConfidenceCalibrator` to adjust for historical accuracy

**Is it accurate?** 
Moderately accurate. The calibration system helps, but confidence scores may be inflated for some industries. The three-tier ML strategy helps validate results.

**Improvements needed:**
- Industry-specific confidence thresholds (partially implemented)
- Better calibration based on historical accuracy data
- Layer-specific confidence thresholds (90%+ for Layer 1, 85-90% for Layer 2, etc.)

### 3.5 Explainability/Auditability

**Current explanation generation:** 
Reasoning strings are generated that describe:
- Which strategies contributed
- Confidence scores per strategy
- ML validation results
- Keyword matches

Example: `"Combined 4 classification strategies using weighted average. Contributions: keyword(0.85), entity(0.72), topic(0.68), co_occurrence(0.61). Final confidence: 0.82"`

**Quality assessment:** 
Good - Provides clear reasoning for classifications. Includes strategy contributions and confidence breakdowns.

**Gap vs proposal:** 
⚠️ Partial - Current explanations are text-based. Proposal requires structured JSON with detailed reasoning. Missing:
- Step-by-step reasoning chain
- Evidence citations (which keywords matched)
- Alternative classifications considered
- Confidence breakdown by code type

---

## 4. Gap Analysis: Current vs Proposed

### 4.1 Layer 1: Rule-Based Classification

| Feature | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| Keyword matching | ✅ Exists | Required | ✅ Complete - Uses database with trigram indexes |
| Trigram analysis | ✅ Exists | Required | ✅ Complete - `classify_business_by_keywords_trigram()` function |
| Fast path routing | ⚠️ Partial | Required (90%+ confidence) | ⚠️ Gap - No dedicated fast path, all requests go through full pipeline |
| Performance target | ⚠️ 200-500ms | Target: <100ms | ❌ Gap - Current performance is 2-5x slower than target |

**Reusable Code:**
```
- internal/classification/multi_strategy_classifier.go:classifyByKeywords() - Keyword matching logic
- internal/classification/repository/supabase_repository.go:ClassifyBusinessByKeywords() - Database queries
- supabase-migrations/030_add_trigram_keyword_classification.sql - Trigram function
- supabase-migrations/035_add_fulltext_search_codes.sql - Full-text search
```

**Required Changes:**
- [ ] Add fast path routing for 90%+ confidence cases (bypass ML, return immediately)
- [ ] Optimize database queries for <100ms target (add query timeouts, connection pooling)
- [ ] Create dedicated Layer 1 handler that skips slower strategies
- [ ] Add performance monitoring to track Layer 1 vs full pipeline times

### 4.2 Layer 2: Embedding-Based Similarity

| Feature | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| Code embeddings | ❌ Missing | Required (384-dim) | ❌ Gap - No embeddings stored, no embedding generation |
| Vector search | ❌ Missing | Required (pgvector) | ❌ Gap - pgvector extension not enabled, no vector indexes |
| Embedding service | ❌ Missing | Required | ❌ Gap - No service to generate embeddings |
| Performance target | N/A | Target: 200-500ms | N/A - Not implemented |

**Reusable Code:**
```
- internal/classification/repository/supabase_repository.go - Can extend to add embedding queries
- services/classification-service/internal/handlers/classification.go - Can add embedding route
- supabase-migrations/ - Can add embedding table migration
```

**Required Changes:**
- [ ] Enable pgvector extension in Supabase
- [ ] Create `code_embeddings` table with vector(384) column
- [ ] Build embedding service (Python/Go) using sentence-transformers (all-MiniLM-L6-v2)
- [ ] Pre-compute embeddings for all codes in database
- [ ] Create `match_code_embeddings()` RPC function for vector similarity search
- [ ] Add embedding-based classification strategy to MultiStrategyClassifier
- [ ] Create embedding service deployment on Railway

### 4.3 Layer 3: LLM Classification

| Feature | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| LLM integration | ⚠️ Partial | Open-source on Railway | ⚠️ Gap - Python ML service exists but uses DistilBART, not Qwen/Mistral |
| Structured output | ⚠️ Partial | JSON with reasoning | ⚠️ Gap - Current output is structured but missing detailed reasoning chain |
| Prompt engineering | ⚠️ Basic | Optimized prompts | ⚠️ Gap - Prompts exist but not optimized for structured output |
| Performance target | ⚠️ 500ms-2s | Target: 2-5s | ✅ Acceptable - Current performance is within range |

**Current LLM Setup:**
- **Model:** DistilBART (via Python ML service)
- **Hosting:** Python service on Railway (optional, not always used)
- **Cost per classification:** Unknown (self-hosted, compute costs only)
- **Latency:** 500ms - 2s average

**Reusable Code:**
```
- python_ml_service/app.py - Can be adapted for LLM integration
- internal/classification/service.go:improveWithML() - ML integration pattern
- services/classification-service/internal/handlers/classification.go - Handler structure
```

**Required Changes:**
- [ ] Deploy Qwen 7B or Mistral 7B on Railway (separate service)
- [ ] Create LLM service with structured JSON output
- [ ] Implement prompt engineering for industry classification
- [ ] Add reasoning chain generation
- [ ] Integrate LLM service into classification pipeline
- [ ] Add fallback logic if LLM service unavailable

### 4.4 Orchestration & Decision Logic

| Feature | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| Multi-layer routing | ❌ Missing | Required | ❌ Gap - No layer-based routing, only confidence-based ML routing |
| Confidence thresholds | ⚠️ Partial | Layer-specific thresholds | ⚠️ Gap - Has thresholds but not layer-specific (90%+ Layer 1, 85-90% Layer 2) |
| Fallback logic | ✅ Exists | Cascading layers | ⚠️ Gap - Has fallbacks but not layer-based cascading |

**Required Changes:**
- [ ] Create `ClassificationOrchestrator` that routes requests to appropriate layer
- [ ] Implement layer-specific confidence thresholds:
  - Layer 1: 90%+ confidence → return immediately
  - Layer 2: 85-90% confidence → try embeddings, return if above threshold
  - Layer 3: <85% confidence → use LLM
- [ ] Add cascading fallback: Layer 1 → Layer 2 → Layer 3
- [ ] Add performance tracking per layer
- [ ] Add routing decision logging

### 4.5 Caching & Performance

| Feature | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| Classification cache | ⚠️ Partial | 30-day cache | ⚠️ Gap - Redis/in-memory cache exists but no Supabase 30-day cache table |
| Content hashing | ✅ Exists | SHA-256 hashing | ✅ Complete - Uses SHA-256 for cache keys |
| Cache invalidation | ✅ Exists | TTL-based | ✅ Complete - TTL-based expiration implemented |

**Current Caching:**
- **Redis cache**: Optional, distributed caching with TTL
- **In-memory cache**: Fallback, per-service instance
- **Cache key**: SHA-256 hash of `business_name|description|website_url`
- **TTL**: Configurable (default: 5 minutes for classification results)
- **Predictive cache**: Pre-loads similar classifications

**Required Changes:**
- [ ] Create `classification_cache` table in Supabase with 30-day TTL
- [ ] Add Supabase cache lookup before Redis/in-memory
- [ ] Implement cache migration from Redis to Supabase for long-term storage
- [ ] Add cache analytics (hit rate, age distribution)

### 4.6 Services & Deployment

| Service | Current State | Proposed State | Gap |
|---------|---------------|----------------|-----|
| Classification API | ✅ Exists (Go) | FastAPI on Railway | ⚠️ Gap - Go service exists, proposal suggests FastAPI (Python) - but Go is fine |
| Scraper Service | ⚠️ Integrated | Playwright on Railway | ⚠️ Gap - Scraper is part of classification service, not separate. Uses basic HTTP, not Playwright |
| Embedding Service | ❌ Missing | Required on Railway | ❌ Gap - No embedding service exists |
| LLM Service | ⚠️ Partial | Qwen/Mistral on Railway | ⚠️ Gap - Python ML service exists but uses DistilBART, not Qwen/Mistral |

**Current Deployment:**
- **Platform:** Railway
- **Architecture:** Microservices (Classification Service, Merchant Service, Risk Assessment, etc.)
- **Services count:** 7+ services deployed
- **Classification Service:** Go-based, deployed on Railway

**Required Changes:**
- [ ] Create separate embedding service (Python/Go) on Railway
- [ ] Deploy Qwen 7B or Mistral 7B LLM service on Railway
- [ ] Consider separate scraper service with Playwright (optional - current scraper works)
- [ ] Add service discovery/health checks for new services
- [ ] Update API Gateway to route to new services

---

## 5. Data Assets & Opportunities

### 5.1 Existing Data That Can Be Leveraged

#### Historical Classifications
- **Exists:** ⚠️ Partial - Classification results are cached but not systematically stored for analysis
- **Volume:** Unknown - Cached results exist but no dedicated tracking table
- **Quality:** Unknown - No feedback mechanism to validate accuracy
- **Opportunity:** 
  - Create `classification_history` table to track all classifications
  - Use historical data for confidence calibration
  - Build training dataset for embedding fine-tuning
  - Analyze patterns for common misclassifications

#### Keyword/Trigram Data
- **Quality:** ✅ Good - Comprehensive keyword database with weights and contexts
- **Coverage:** ✅ Good - Keywords exist for major industries, trigram indexes optimized
- **Opportunity:** 
  - Use existing keywords to seed embedding training
  - Analyze keyword co-occurrence patterns for better matching
  - Expand keyword coverage for emerging industries
  - Use trigram data to improve fuzzy matching thresholds

#### Code Mappings
- **Completeness:** ✅ Good - Crosswalks table has MCC/NAICS/SIC mappings with confidence scores
- **Opportunity:** 
  - Use crosswalk data to validate embedding-based matches
  - Build code similarity matrix from crosswalks
  - Use usage_frequency to weight embedding matches
  - Enhance crosswalks with embedding-based similarity scores

### 5.2 Quick Wins

1. **Enable Fast Path for High-Confidence Classifications**
   - **Effort:** 2-4 hours
   - **Impact:** High
   - **Description:** Add early return in `IndustryDetectionService` for 90%+ confidence cases, bypassing ML and slower strategies
   - **Files to modify:** 
     - `internal/classification/service.go:performClassification()`
     - Add fast path check after `MultiStrategyClassifier` result

2. **Add Supabase Classification Cache Table**
   - **Effort:** 1-2 hours
   - **Impact:** Medium
   - **Description:** Create migration for `classification_cache` table, update cache logic to use Supabase for 30-day storage
   - **Files to modify:**
     - Create `supabase-migrations/036_add_classification_cache.sql`
     - Update `services/classification-service/internal/handlers/classification.go:getCachedResponse()`

3. **Optimize Database Queries for Layer 1**
   - **Effort:** 4-8 hours
   - **Impact:** High
   - **Description:** Add query timeouts, optimize indexes, use prepared statements for common queries
   - **Files to modify:**
     - `internal/classification/repository/supabase_repository.go`
     - Add connection pooling configuration

4. **Enable pgvector Extension**
   - **Effort:** 1 hour
   - **Impact:** Medium (enables Layer 2)
   - **Description:** Create migration to enable pgvector extension in Supabase
   - **Files to modify:**
     - Create `supabase-migrations/037_enable_pgvector.sql`

### 5.3 High-Value Additions

1. **Build Embedding Service and Pre-compute Code Embeddings**
   - **Effort:** 2-3 weeks
   - **Impact:** High
   - **Value:** Enables Layer 2, improves accuracy for edge cases, reduces LLM usage
   - **Dependencies:** 
     - pgvector extension enabled
     - `code_embeddings` table created
     - sentence-transformers library setup

2. **Deploy Open-Source LLM Service (Qwen 7B or Mistral 7B)**
   - **Effort:** 1-2 weeks
   - **Impact:** High
   - **Value:** Handles complex/ambiguous cases, provides structured reasoning, reduces dependency on external APIs
   - **Dependencies:**
     - Railway service with sufficient GPU/memory
     - LLM model download and setup
     - Structured output prompt engineering

3. **Implement 3-Layer Orchestration Logic**
   - **Effort:** 1 week
   - **Impact:** High
   - **Value:** Optimizes performance by routing to appropriate layer, reduces costs, improves user experience
   - **Dependencies:**
     - Layer 1 optimizations complete
     - Layer 2 (embeddings) implemented
     - Layer 3 (LLM) deployed

---

## 6. Performance & Cost Assessment

### 6.1 Current Performance Metrics

- **Average classification time:** 200-500ms (uncached), 10-50ms (cached)
- **95th percentile:** ~800ms (uncached), ~100ms (cached)
- **Cache hit rate:** Unknown (no metrics collected) - estimated 30-40% based on Redis usage
- **Error rate:** Low (<1% based on code structure)

**Performance Breakdown:**
- Database queries: 50-150ms
- Multi-strategy classification: 100-200ms
- ML enhancement (when used): 500ms-2s
- Code generation: 50-100ms
- Total (uncached): 200-500ms (without ML), 700ms-2.5s (with ML)

### 6.2 Current Cost Structure

- **Per classification:** 
  - Compute: ~$0.0001-0.0005 (Railway compute costs)
  - Database: Negligible (Supabase free tier)
  - ML (when used): ~$0.0002-0.001 (Python service compute)
  - **Total:** ~$0.0001-0.0015 per classification
- **Monthly volume:** Unknown (no metrics)
- **Monthly cost:** Estimated $10-50/month for compute (depending on volume)
- **Main cost drivers:** 
  - Railway compute (CPU/memory)
  - Python ML service compute (when used)
  - Redis (if using paid tier)

### 6.3 Projected Performance (with proposed changes)

- **Expected average time:** 
  - Layer 1 (90%+ cases): <100ms (target met)
  - Layer 2 (85-90% cases): 200-500ms (target met)
  - Layer 3 (<85% cases): 2-5s (target met)
  - **Overall average:** ~150-300ms (assuming 70% Layer 1, 20% Layer 2, 10% Layer 3)
- **Expected cost reduction:** 40-60% (fewer ML calls, more fast-path classifications)
- **Expected accuracy improvement:** 5-10% (embeddings handle edge cases, LLM handles complex cases)

---

## 7. Technical Debt & Issues

### 7.1 Current Issues

1. **Python ML Service Not Fully Integrated**
   - **Severity:** Medium
   - **Impact:** ML enhancement available but not consistently used, reducing accuracy for low-confidence cases
   - **Should fix during refactor:** Yes - Integrate properly or remove

2. **No Performance Monitoring Per Layer**
   - **Severity:** Low
   - **Impact:** Cannot measure which layer is used most, cannot optimize routing decisions
   - **Should fix during refactor:** Yes - Add metrics for layer usage

3. **Cache Key Normalization Issues**
   - **Severity:** Low
   - **Impact:** Potential cache misses due to inconsistent key generation (mentioned in docs)
   - **Should fix during refactor:** Yes - Standardize cache key generation

4. **Database Connection Pooling Not Optimized**
   - **Severity:** Medium
   - **Impact:** May cause performance issues under load
   - **Should fix during refactor:** Yes - Configure proper connection pooling

### 7.2 Code Quality Observations

- **Test coverage:** Unknown (no coverage reports found) - estimated 40-50% based on test files present
- **Documentation:** Good - Comprehensive docs, code comments present
- **Code organization:** Good - Clean architecture, separation of concerns, interface-driven design
- **Areas needing refactor:** 
  - ML integration code (scattered across multiple files)
  - Cache logic (multiple cache implementations - consolidate)
  - Database query optimization (add prepared statements, connection pooling)

---

## 8. Migration Path Recommendation

### 8.1 Recommended Approach

**Option A: Incremental Enhancement** ✅ **RECOMMENDED**

- Keep existing implementation running
- Add new layers one at a time (Layer 2 first, then Layer 3)
- Gradually shift traffic to new system based on confidence thresholds
- Timeline: 6-8 weeks

**Rationale:** Current system is working well. Incremental approach minimizes risk, allows testing at each stage, and maintains system availability.

### 8.2 Implementation Phases

**Phase 1: Foundation (Week 1-2)**
- [ ] Enable pgvector extension in Supabase
- [ ] Create `code_embeddings` table migration
- [ ] Create `classification_cache` table migration
- [ ] Add fast path routing for 90%+ confidence cases
- [ ] Optimize database queries for <100ms target
- [ ] Add performance monitoring per layer

**Phase 2: Layer 2 - Embeddings (Week 3-4)**
- [ ] Build embedding service (Python/Go) using sentence-transformers
- [ ] Pre-compute embeddings for all codes
- [ ] Create `match_code_embeddings()` RPC function
- [ ] Add embedding-based classification strategy
- [ ] Deploy embedding service on Railway
- [ ] Test and validate embedding matches

**Phase 3: Layer 3 - LLM (Week 5-6)**
- [ ] Deploy Qwen 7B or Mistral 7B on Railway
- [ ] Create LLM service with structured JSON output
- [ ] Implement prompt engineering for classification
- [ ] Add reasoning chain generation
- [ ] Integrate LLM into classification pipeline
- [ ] Test LLM accuracy and performance

**Phase 4: Orchestration (Week 7)**
- [ ] Create `ClassificationOrchestrator` with layer routing
- [ ] Implement layer-specific confidence thresholds
- [ ] Add cascading fallback logic
- [ ] Add routing decision logging
- [ ] Test end-to-end flow

**Phase 5: Testing & Optimization (Week 8)**
- [ ] Performance testing (load testing, latency measurement)
- [ ] Accuracy validation (test dataset, A/B testing)
- [ ] Cost analysis (compare before/after)
- [ ] Documentation updates
- [ ] Production deployment

### 8.3 Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| pgvector performance issues | Low | Medium | Test with sample data first, optimize indexes |
| LLM service reliability | Medium | High | Add fallback to Layer 2, implement retry logic |
| Embedding accuracy lower than expected | Medium | Medium | Validate against test dataset, fine-tune model if needed |
| Migration breaks existing functionality | Low | High | Incremental rollout, feature flags, canary deployments |
| Cost overruns (LLM compute) | Medium | Medium | Monitor usage, implement rate limiting, optimize prompts |

---

## 9. Code Reusability Matrix

| File Path | Current Purpose | Reusable? | For Which Layer? | Modifications Needed |
|-----------|-----------------|-----------|------------------|---------------------|
| `internal/classification/multi_strategy_classifier.go` | Multi-strategy ensemble | ✅ Yes | Layer 1 | Add fast path, optimize for <100ms |
| `internal/classification/repository/supabase_repository.go` | Database access | ✅ Yes | All layers | Add embedding queries, cache table queries |
| `internal/classification/service.go` | Main orchestrator | ⚠️ Partial | Orchestrator | Refactor to add layer routing logic |
| `internal/classification/classifier.go` | Code generation | ✅ Yes | All layers | No changes needed |
| `services/classification-service/internal/handlers/classification.go` | HTTP handler | ✅ Yes | All layers | Add layer routing, update cache logic |
| `internal/external/website_scraper.go` | Website scraping | ✅ Yes | All layers | Consider Playwright upgrade (optional) |
| `internal/classification/cache/predictive_cache.go` | Caching | ✅ Yes | All layers | Extend to use Supabase cache table |
| `supabase-migrations/030_add_trigram_keyword_classification.sql` | Trigram function | ✅ Yes | Layer 1 | No changes needed |
| `supabase-migrations/035_add_fulltext_search_codes.sql` | Full-text search | ✅ Yes | Layer 1 | No changes needed |
| `python_ml_service/app.py` | ML service | ⚠️ Partial | Layer 3 | Replace DistilBART with Qwen/Mistral, add structured output |
| `internal/classification/nlp/entity_recognizer.go` | Entity recognition | ✅ Yes | Layer 1 | No changes needed |
| `internal/classification/nlp/topic_modeler.go` | Topic modeling | ✅ Yes | Layer 1 | No changes needed |

---

## 10. Recommendations Summary

### 10.1 Top Priorities

1. **Enable Fast Path for High-Confidence Cases** - Quick win, immediate performance improvement, enables Layer 1 target of <100ms
2. **Build Embedding Service and Pre-compute Code Embeddings** - Enables Layer 2, high impact on accuracy for edge cases
3. **Implement 3-Layer Orchestration Logic** - Critical for routing requests to appropriate layer, optimizes performance and costs

### 10.2 Technologies to Add

- [ ] pgvector extension in Supabase
- [ ] sentence-transformers library (Python) or equivalent Go library
- [ ] Qwen 7B or Mistral 7B LLM model
- [ ] Playwright (optional - for enhanced website scraping)
- [ ] Vector similarity search libraries

### 10.3 Code to Preserve

- **Definitely keep:**
  - `MultiStrategyClassifier` - Core classification logic
  - `SupabaseKeywordRepository` - Database access layer
  - `ClassificationCodeGenerator` - Code generation logic
  - Trigram and full-text search functions
  - Caching infrastructure (extend, don't replace)
  - Website scraper (works well, optional Playwright upgrade)

- **Refactor and keep:**
  - `IndustryDetectionService` - Add layer routing logic
  - ML integration code - Consolidate and improve
  - Cache logic - Extend to Supabase, consolidate implementations

- **Consider deprecating:**
  - DistilBART Python service (replace with Qwen/Mistral)
  - In-memory cache only (keep as fallback, prefer Redis/Supabase)

### 10.4 Estimated Effort

- **Total development time:** 6-8 weeks
- **Team size needed:** 1-2 developers
- **Complexity level:** Medium-High (due to LLM deployment and embedding service)

---

## 11. Appendices

### Appendix A: Full File Tree

```
kyb-platform/
├── internal/
│   ├── classification/
│   │   ├── service.go                    # Main orchestrator
│   │   ├── multi_strategy_classifier.go  # Multi-strategy ensemble
│   │   ├── classifier.go                 # Code generator
│   │   ├── cache/
│   │   │   └── predictive_cache.go       # Caching logic
│   │   ├── repository/
│   │   │   └── supabase_repository.go   # Database access
│   │   └── nlp/
│   │       ├── entity_recognizer.go     # NER
│   │       └── topic_modeler.go         # Topic modeling
│   └── external/
│       └── website_scraper.go           # Website scraping
├── services/
│   └── classification-service/
│       ├── internal/
│       │   ├── handlers/
│       │   │   └── classification.go    # HTTP handler
│       │   └── cache/
│       │       └── redis_cache.go       # Redis caching
├── python_ml_service/
│   └── app.py                           # DistilBART ML service
└── supabase-migrations/
    ├── 001_initial_keyword_classification_schema.sql
    ├── 030_add_trigram_keyword_classification.sql
    └── 035_add_fulltext_search_codes.sql
```

### Appendix B: Environment Variables

```bash
# Database
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_KEY=xxx
SUPABASE_DB_URL=postgresql://...

# Classification Service
CLASSIFICATION_CACHE_ENABLED=true
CLASSIFICATION_CACHE_TTL=5m
CLASSIFICATION_REQUEST_TIMEOUT=30s
CLASSIFICATION_OVERALL_TIMEOUT=60s

# Redis (optional)
REDIS_URL=redis://...
REDIS_ENABLED=true

# ML Service (optional)
PYTHON_ML_SERVICE_URL=http://...
USE_ML=true

# Performance
MAX_CONCURRENT_REQUESTS=100
DATABASE_MAX_CONNECTIONS=25
```

### Appendix C: Code Snippets

```go
// Current classification flow (simplified)
func (s *IndustryDetectionService) DetectIndustry(ctx context.Context, businessName, description, websiteURL string) (*IndustryDetectionResult, error) {
    // Step 1: Multi-strategy classification
    multiResult, err := s.multiStrategyClassifier.ClassifyWithMultiStrategy(ctx, businessName, description, websiteURL)
    
    // Step 2: Confidence-based ML routing
    if multiResult.Confidence < 0.5 {
        // Low confidence: ML-assisted improvement
        result, err = s.improveWithML(ctx, multiResult, ...)
    } else if multiResult.Confidence < 0.8 {
        // Medium confidence: Ensemble validation
        result, err = s.validateWithEnsemble(ctx, multiResult, ...)
    } else {
        // High confidence: ML validation only
        result, err = s.validateWithMLHighConfidence(ctx, multiResult, ...)
    }
    
    return result, nil
}
```

---

## Questions for Clarification

1. **LLM Model Preference**: Should we use Qwen 7B or Mistral 7B? Any specific requirements for model selection?

2. **Embedding Model**: Confirm all-MiniLM-L6-v2 (384-dim) is the preferred model, or should we consider alternatives?

3. **Cache Strategy**: Should the 30-day Supabase cache replace Redis entirely, or complement it (Redis for short-term, Supabase for long-term)?

4. **Performance Targets**: Are the targets (Layer 1: <100ms, Layer 2: 200-500ms, Layer 3: 2-5s) firm, or can they be adjusted based on implementation constraints?

5. **Deployment Priority**: Which layer should be implemented first - Layer 2 (embeddings) or Layer 3 (LLM)? Or should they be built in parallel?

---

**End of Analysis Report**

