# Phase 5 Kick-Off Guide: Production Ready
## Week 9: Polish, Cache, Monitor, Ship

**Goal:** Transform your 90-95% accurate classification system into a production-ready, monitored, cached, and user-friendly platform.

---

## Phase 4 Success Validation

Before starting Phase 5, verify your Phase 4 results:

✅ **Checklist:**
- [ ] LLM service deployed and stable (<5s response)
- [ ] 3-layer routing working correctly
- [ ] Layer 3 triggers for 5-10% of requests
- [ ] Accuracy at 90-95% on test set
- [ ] All layers functioning with proper fallbacks

**If all checked:** You're ready for Phase 5! 🎉

**If issues remain:** Address them before proceeding. Phase 5 adds polish, not new features.

---

## Phase 5 Overview

### What We're Adding

**Current State (After Phase 4):**
- ✅ 90-95% accuracy
- ✅ 3-layer orchestration
- ✅ LLM reasoning for complex cases
- ❌ No caching (repeated classifications slow)
- ❌ No monitoring dashboard
- ❌ Explanations not shown in UI
- ❌ No performance metrics tracking

**Target State (After Phase 5):**
- ✅ 30-day classification cache
- ✅ Monitoring dashboard with metrics
- ✅ UI shows explanations and reasoning
- ✅ Performance tracking and analytics
- ✅ Production deployment checklist complete
- ✅ Documentation for team

### Implementation Timeline

**Week 9 (5-7 days):**
- Day 1: Add 30-day classification cache
- Day 2: Build monitoring dashboard
- Day 3: UI integration (show explanations)
- Day 4: Performance optimization
- Day 5-6: Testing and validation
- Day 7: Production deployment

---

## Day 1: Classification Cache

### Why Caching?

**Problem:**
```
User classifies "https://mcdonalds.com"
→ Scrape (2s) + Layer 1 (0.3s) = 2.3s

User classifies "https://mcdonalds.com" again
→ Scrape (2s) + Layer 1 (0.3s) = 2.3s  ❌ Wasteful!
```

**With Cache:**
```
First request: 2.3s → Cache result
Second request: 0.05s ✅ Instant from cache!
```

**Benefits:**
- 98% faster for repeated classifications
- Lower compute costs
- Better user experience
- Reduces load on scraper and LLM

### Implementation

**File:** `supabase-migrations/060_add_classification_cache.sql`

```sql
-- Migration: Add classification caching infrastructure

-- Step 1: Create classification_cache table
CREATE TABLE classification_cache (
    id BIGSERIAL PRIMARY KEY,
    
    -- Cache key (content hash)
    content_hash VARCHAR(64) NOT NULL UNIQUE,  -- SHA-256 hash of website content
    
    -- Input data (for debugging)
    business_name VARCHAR(255),
    website_url TEXT,
    
    -- Cached result
    classification_result JSONB NOT NULL,
    
    -- Metadata
    layer_used VARCHAR(20),  -- layer1_high_conf, layer2_better, layer3_high_conf
    confidence DECIMAL(5,4),
    processing_time_ms INTEGER,
    
    -- Cache management
    created_at TIMESTAMPTZ DEFAULT NOW(),
    accessed_at TIMESTAMPTZ DEFAULT NOW(),
    access_count INTEGER DEFAULT 1,
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '30 days'),
    
    -- Indexes
    CONSTRAINT valid_confidence CHECK (confidence >= 0 AND confidence <= 1)
);

-- Step 2: Create indexes for fast lookups
CREATE INDEX idx_cache_content_hash ON classification_cache(content_hash);
CREATE INDEX idx_cache_expires_at ON classification_cache(expires_at);
CREATE INDEX idx_cache_accessed_at ON classification_cache(accessed_at);
CREATE INDEX idx_cache_created_at ON classification_cache(created_at);

-- Step 3: Create function to get cached result
CREATE OR REPLACE FUNCTION get_cached_classification(
    p_content_hash VARCHAR(64)
)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
    cached_result JSONB;
BEGIN
    -- Get result if not expired
    SELECT classification_result INTO cached_result
    FROM classification_cache
    WHERE content_hash = p_content_hash
        AND expires_at > NOW()
    LIMIT 1;
    
    -- Update access stats if found
    IF cached_result IS NOT NULL THEN
        UPDATE classification_cache
        SET accessed_at = NOW(),
            access_count = access_count + 1
        WHERE content_hash = p_content_hash;
    END IF;
    
    RETURN cached_result;
END;
$$;

-- Step 4: Create function to set cached result
CREATE OR REPLACE FUNCTION set_cached_classification(
    p_content_hash VARCHAR(64),
    p_business_name VARCHAR(255),
    p_website_url TEXT,
    p_result JSONB,
    p_layer_used VARCHAR(20),
    p_confidence DECIMAL(5,4),
    p_processing_time_ms INTEGER
)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO classification_cache (
        content_hash,
        business_name,
        website_url,
        classification_result,
        layer_used,
        confidence,
        processing_time_ms
    )
    VALUES (
        p_content_hash,
        p_business_name,
        p_website_url,
        p_result,
        p_layer_used,
        p_confidence,
        p_processing_time_ms
    )
    ON CONFLICT (content_hash) DO UPDATE
    SET
        classification_result = p_result,
        layer_used = p_layer_used,
        confidence = p_confidence,
        processing_time_ms = p_processing_time_ms,
        accessed_at = NOW(),
        access_count = classification_cache.access_count + 1,
        expires_at = NOW() + INTERVAL '30 days';
END;
$$;

-- Step 5: Create cleanup function for expired entries
CREATE OR REPLACE FUNCTION cleanup_expired_cache()
RETURNS INTEGER
LANGUAGE plpgsql
AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM classification_cache
    WHERE expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count;
END;
$$;

-- Step 6: Grant permissions
GRANT SELECT, INSERT, UPDATE ON classification_cache TO authenticated;
GRANT USAGE, SELECT ON SEQUENCE classification_cache_id_seq TO authenticated;
GRANT EXECUTE ON FUNCTION get_cached_classification TO authenticated;
GRANT EXECUTE ON FUNCTION set_cached_classification TO authenticated;

-- Step 7: Add helpful comments
COMMENT ON TABLE classification_cache IS '30-day cache for classification results to avoid repeated processing';
COMMENT ON COLUMN classification_cache.content_hash IS 'SHA-256 hash of website content for cache key';
COMMENT ON COLUMN classification_cache.expires_at IS 'Cache entries expire after 30 days';
COMMENT ON FUNCTION get_cached_classification IS 'Retrieve cached classification if not expired';
COMMENT ON FUNCTION set_cached_classification IS 'Store or update cached classification';
```

**Run Migration:**
```bash
psql $SUPABASE_DB_URL -f supabase-migrations/060_add_classification_cache.sql
```

### Integrate Cache in Go

**File:** `internal/classification/cache.go` (NEW)

```go
package classification

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log/slog"
    "time"
)

type ClassificationCache struct {
    repo Repository
}

func NewClassificationCache(repo Repository) *ClassificationCache {
    return &ClassificationCache{
        repo: repo,
    }
}

// Generate cache key from website content
func (c *ClassificationCache) GenerateCacheKey(content *ScrapedContent) string {
    // Create deterministic string from content
    contentStr := fmt.Sprintf(
        "%s|%s|%s|%s",
        content.Title,
        content.MetaDesc,
        content.AboutText,
        content.Domain,
    )
    
    // Hash it
    hash := sha256.Sum256([]byte(contentStr))
    return hex.EncodeToString(hash[:])
}

// Get cached result
func (c *ClassificationCache) Get(
    ctx context.Context,
    cacheKey string,
) (*IndustryDetectionResult, error) {
    
    slog.Debug("Checking cache", "key", cacheKey[:16]+"...")
    
    // Query database
    result, err := c.repo.GetCachedClassification(ctx, cacheKey)
    if err != nil {
        return nil, err
    }
    
    if result == nil {
        slog.Debug("Cache miss")
        return nil, nil
    }
    
    slog.Info("Cache hit", 
        "key", cacheKey[:16]+"...",
        "age_hours", time.Since(result.CreatedAt).Hours())
    
    return result, nil
}

// Set cached result
func (c *ClassificationCache) Set(
    ctx context.Context,
    cacheKey string,
    businessName string,
    websiteURL string,
    result *IndustryDetectionResult,
) error {
    
    slog.Debug("Caching result", "key", cacheKey[:16]+"...")
    
    err := c.repo.SetCachedClassification(
        ctx,
        cacheKey,
        businessName,
        websiteURL,
        result,
    )
    
    if err != nil {
        slog.Warn("Failed to cache result", "error", err)
        return err
    }
    
    slog.Debug("Result cached successfully")
    return nil
}

// Get cache statistics
func (c *ClassificationCache) GetStats(ctx context.Context) (*CacheStats, error) {
    stats, err := c.repo.GetCacheStats(ctx)
    if err != nil {
        return nil, err
    }
    
    return stats, nil
}

type CacheStats struct {
    TotalEntries     int
    HitRate          float64
    AvgAge           time.Duration
    ExpiringSoon     int  // Expiring in next 7 days
}
```

**File:** `internal/classification/repository/supabase_repository.go` (Add methods)

```go
func (r *SupabaseKeywordRepository) GetCachedClassification(
    ctx context.Context,
    contentHash string,
) (*IndustryDetectionResult, error) {
    
    var cached struct {
        ClassificationResult json.RawMessage `db:"classification_result"`
        CreatedAt            time.Time       `db:"created_at"`
    }
    
    query := `SELECT classification_result, created_at 
              FROM get_cached_classification($1)`
    
    err := r.db.GetContext(ctx, &cached, query, contentHash)
    if err == sql.ErrNoRows {
        return nil, nil  // Cache miss
    }
    if err != nil {
        return nil, err
    }
    
    // Parse JSON result
    var result IndustryDetectionResult
    if err := json.Unmarshal(cached.ClassificationResult, &result); err != nil {
        return nil, err
    }
    
    result.CachedAt = &cached.CreatedAt
    return &result, nil
}

func (r *SupabaseKeywordRepository) SetCachedClassification(
    ctx context.Context,
    contentHash string,
    businessName string,
    websiteURL string,
    result *IndustryDetectionResult,
) error {
    
    // Serialize result to JSON
    resultJSON, err := json.Marshal(result)
    if err != nil {
        return err
    }
    
    query := `SELECT set_cached_classification($1, $2, $3, $4, $5, $6, $7)`
    
    _, err = r.db.ExecContext(
        ctx,
        query,
        contentHash,
        businessName,
        websiteURL,
        resultJSON,
        result.Explanation.ProcessingPath,
        result.Classification.Confidence,
        result.ProcessingTimeMs,
    )
    
    return err
}

func (r *SupabaseKeywordRepository) GetCacheStats(ctx context.Context) (*CacheStats, error) {
    stats := &CacheStats{}
    
    // Total entries
    query := `SELECT COUNT(*) FROM classification_cache WHERE expires_at > NOW()`
    err := r.db.GetContext(ctx, &stats.TotalEntries, query)
    if err != nil {
        return nil, err
    }
    
    // Average age
    query = `SELECT EXTRACT(EPOCH FROM AVG(NOW() - created_at)) 
             FROM classification_cache WHERE expires_at > NOW()`
    var avgSeconds float64
    err = r.db.GetContext(ctx, &avgSeconds, query)
    if err == nil {
        stats.AvgAge = time.Duration(avgSeconds) * time.Second
    }
    
    // Expiring soon (next 7 days)
    query = `SELECT COUNT(*) FROM classification_cache 
             WHERE expires_at > NOW() AND expires_at < NOW() + INTERVAL '7 days'`
    err = r.db.GetContext(ctx, &stats.ExpiringSoon, query)
    
    return stats, nil
}
```

### Integrate Cache in Service

**File:** `internal/classification/service.go`

```go
type IndustryDetectionService struct {
    // Existing fields...
    cache *ClassificationCache // NEW
}

func NewIndustryDetectionService(
    // ... existing params ...
    repo Repository,
) *IndustryDetectionService {
    return &IndustryDetectionService{
        // ... existing initialization ...
        cache: NewClassificationCache(repo),
    }
}

func (s *IndustryDetectionService) DetectIndustry(
    ctx context.Context,
    businessName, description, websiteURL string,
) (*IndustryDetectionResult, error) {
    
    startTime := time.Now()
    
    slog.Info("Starting industry detection", "url", websiteURL)
    
    // Scrape website
    content, err := s.scraper.Scrape(websiteURL)
    if err != nil {
        return nil, fmt.Errorf("scraping failed: %w", err)
    }
    
    // Generate cache key
    cacheKey := s.cache.GenerateCacheKey(content)
    
    // Check cache first
    cachedResult, err := s.cache.Get(ctx, cacheKey)
    if err != nil {
        slog.Warn("Cache lookup error", "error", err)
        // Continue without cache
    } else if cachedResult != nil {
        slog.Info("Returning cached result",
            "age_hours", time.Since(*cachedResult.CachedAt).Hours(),
            "original_processing_ms", cachedResult.ProcessingTimeMs)
        
        // Update processing time to reflect cache hit
        cachedResult.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        cachedResult.FromCache = true
        
        return cachedResult, nil
    }
    
    // Cache miss - run classification
    slog.Info("Cache miss - running classification")
    
    // Run 3-layer orchestration (existing code)
    result := s.runClassification(ctx, content, businessName, description, websiteURL)
    
    result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
    
    // Cache the result (async, don't block response)
    go func() {
        cacheCtx := context.Background()
        if err := s.cache.Set(cacheCtx, cacheKey, businessName, websiteURL, result); err != nil {
            slog.Warn("Failed to cache result", "error", err)
        }
    }()
    
    return result, nil
}

// Extract existing classification logic to separate method
func (s *IndustryDetectionService) runClassification(
    ctx context.Context,
    content *ScrapedContent,
    businessName, description, websiteURL string,
) *IndustryDetectionResult {
    
    // Layer 1: Multi-Strategy
    layer1Result, _ := s.multiStrategyClassifier.ClassifyWithMultiStrategy(...)
    
    if layer1Result.Confidence >= 0.90 {
        return s.buildResult(layer1Result, "layer1_high_conf")
    }
    
    // Layer 2: Embeddings
    // ... existing Layer 2 logic ...
    
    // Layer 3: LLM
    // ... existing Layer 3 logic ...
    
    return result
}
```

**Update Result Structure:**

```go
type IndustryDetectionResult struct {
    // Existing fields...
    FromCache      bool       `json:"from_cache"`
    CachedAt       *time.Time `json:"cached_at,omitempty"`
}
```

**Test Cache:**

```bash
# First request (cache miss)
time curl -X POST http://localhost:8080/v1/classify \
  -d '{"website_url": "https://mcdonalds.com"}'
# Expected: 2-3s, from_cache: false

# Second request (cache hit)
time curl -X POST http://localhost:8080/v1/classify \
  -d '{"website_url": "https://mcdonalds.com"}'
# Expected: <100ms, from_cache: true ✅

# Check logs:
# "Cache hit age_hours=0.01"
```

---

## Day 2: Monitoring Dashboard

### Analytics Table

**File:** `supabase-migrations/061_add_analytics_tables.sql`

```sql
-- Migration: Add analytics and monitoring tables

-- Step 1: Create classification_metrics table
CREATE TABLE classification_metrics (
    id BIGSERIAL PRIMARY KEY,
    
    -- Request info
    request_id VARCHAR(36),
    business_name VARCHAR(255),
    website_url TEXT,
    
    -- Classification result
    primary_industry VARCHAR(255),
    confidence DECIMAL(5,4),
    layer_used VARCHAR(20),
    method VARCHAR(50),
    
    -- Performance
    total_time_ms INTEGER,
    scrape_time_ms INTEGER,
    layer1_time_ms INTEGER,
    layer2_time_ms INTEGER,
    layer3_time_ms INTEGER,
    
    -- Cache
    from_cache BOOLEAN DEFAULT FALSE,
    
    -- Codes
    mcc_codes JSONB,
    sic_codes JSONB,
    naics_codes JSONB,
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    user_agent TEXT,
    ip_address INET
);

-- Step 2: Create indexes
CREATE INDEX idx_metrics_created_at ON classification_metrics(created_at DESC);
CREATE INDEX idx_metrics_layer_used ON classification_metrics(layer_used);
CREATE INDEX idx_metrics_from_cache ON classification_metrics(from_cache);
CREATE INDEX idx_metrics_confidence ON classification_metrics(confidence);
CREATE INDEX idx_metrics_total_time ON classification_metrics(total_time_ms);

-- Step 3: Create materialized view for dashboard
CREATE MATERIALIZED VIEW classification_dashboard AS
SELECT
    DATE_TRUNC('day', created_at) as date,
    COUNT(*) as total_classifications,
    COUNT(*) FILTER (WHERE from_cache) as cache_hits,
    COUNT(*) FILTER (WHERE NOT from_cache) as cache_misses,
    ROUND(AVG(confidence), 4) as avg_confidence,
    ROUND(AVG(total_time_ms), 0) as avg_total_time_ms,
    COUNT(*) FILTER (WHERE layer_used LIKE 'layer1%') as layer1_count,
    COUNT(*) FILTER (WHERE layer_used LIKE 'layer2%') as layer2_count,
    COUNT(*) FILTER (WHERE layer_used LIKE 'layer3%') as layer3_count,
    COUNT(*) FILTER (WHERE confidence >= 0.90) as high_confidence_count,
    COUNT(*) FILTER (WHERE confidence < 0.70) as low_confidence_count
FROM classification_metrics
WHERE created_at >= NOW() - INTERVAL '90 days'
GROUP BY DATE_TRUNC('day', created_at)
ORDER BY date DESC;

-- Step 4: Create index on materialized view
CREATE UNIQUE INDEX ON classification_dashboard (date);

-- Step 5: Create function to refresh dashboard
CREATE OR REPLACE FUNCTION refresh_classification_dashboard()
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY classification_dashboard;
END;
$$;

-- Step 6: Create function to get dashboard summary
CREATE OR REPLACE FUNCTION get_dashboard_summary(days INTEGER DEFAULT 30)
RETURNS TABLE (
    metric VARCHAR,
    value NUMERIC,
    description TEXT
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT * FROM (
        VALUES
            ('total_classifications', 
             (SELECT COUNT(*)::NUMERIC FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Total classifications in period'),
            
            ('cache_hit_rate',
             (SELECT ROUND(COUNT(*) FILTER (WHERE from_cache)::NUMERIC / 
                     NULLIF(COUNT(*), 0) * 100, 2)
              FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Percentage of requests served from cache'),
            
            ('avg_confidence',
             (SELECT ROUND(AVG(confidence), 4)::NUMERIC FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Average confidence score'),
            
            ('avg_processing_time_ms',
             (SELECT ROUND(AVG(total_time_ms), 0)::NUMERIC FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL 
              AND NOT from_cache),
             'Average processing time (non-cached)'),
            
            ('layer1_percentage',
             (SELECT ROUND(COUNT(*) FILTER (WHERE layer_used LIKE 'layer1%')::NUMERIC / 
                     NULLIF(COUNT(*), 0) * 100, 2)
              FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Percentage using Layer 1'),
            
            ('layer3_percentage',
             (SELECT ROUND(COUNT(*) FILTER (WHERE layer_used LIKE 'layer3%')::NUMERIC / 
                     NULLIF(COUNT(*), 0) * 100, 2)
              FROM classification_metrics 
              WHERE created_at >= NOW() - (days || ' days')::INTERVAL),
             'Percentage using Layer 3 (LLM)')
    ) AS stats(metric, value, description);
END;
$$;

-- Step 7: Grant permissions
GRANT SELECT, INSERT ON classification_metrics TO authenticated;
GRANT USAGE, SELECT ON SEQUENCE classification_metrics_id_seq TO authenticated;
GRANT SELECT ON classification_dashboard TO authenticated;
GRANT EXECUTE ON FUNCTION get_dashboard_summary TO authenticated;

COMMENT ON TABLE classification_metrics IS 'Detailed metrics for every classification request';
COMMENT ON MATERIALIZED VIEW classification_dashboard IS 'Daily aggregated metrics for monitoring dashboard';
```

**Run Migration:**
```bash
psql $SUPABASE_DB_URL -f supabase-migrations/061_add_analytics_tables.sql
```

### Log Metrics in Service

**File:** `internal/classification/service.go`

```go
func (s *IndustryDetectionService) DetectIndustry(
    ctx context.Context,
    businessName, description, websiteURL string,
) (*IndustryDetectionResult, error) {
    
    startTime := time.Now()
    requestID := generateRequestID()
    
    // ... existing classification logic ...
    
    result.RequestID = requestID
    result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
    
    // Log metrics (async, don't block response)
    go func() {
        metricsCtx := context.Background()
        if err := s.logMetrics(metricsCtx, result, businessName, websiteURL); err != nil {
            slog.Warn("Failed to log metrics", "error", err)
        }
    }()
    
    return result, nil
}

func (s *IndustryDetectionService) logMetrics(
    ctx context.Context,
    result *IndustryDetectionResult,
    businessName string,
    websiteURL string,
) error {
    
    metrics := ClassificationMetrics{
        RequestID:        result.RequestID,
        BusinessName:     businessName,
        WebsiteURL:       websiteURL,
        PrimaryIndustry:  result.Classification.PrimaryIndustry,
        Confidence:       result.Classification.Confidence,
        LayerUsed:        result.Explanation.ProcessingPath,
        Method:           result.Classification.Method,
        TotalTimeMs:      result.ProcessingTimeMs,
        FromCache:        result.FromCache,
        MCCCodes:         result.Codes.MCC,
        SICCodes:         result.Codes.SIC,
        NAICSCodes:       result.Codes.NAICS,
    }
    
    return s.repo.LogMetrics(ctx, metrics)
}
```

### Simple Dashboard API

**File:** `services/classification-service/internal/handlers/dashboard.go` (NEW)

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
)

type DashboardHandler struct {
    repo Repository
}

func NewDashboardHandler(repo Repository) *DashboardHandler {
    return &DashboardHandler{repo: repo}
}

func (h *DashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
    // Get days parameter (default 30)
    days := 30
    if d := r.URL.Query().Get("days"); d != "" {
        if parsed, err := strconv.Atoi(d); err == nil {
            days = parsed
        }
    }
    
    // Get summary from database
    summary, err := h.repo.GetDashboardSummary(r.Context(), days)
    if err != nil {
        http.Error(w, "Failed to get dashboard summary", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(summary)
}

func (h *DashboardHandler) GetTimeSeries(w http.ResponseWriter, r *http.Request) {
    days := 30
    if d := r.URL.Query().Get("days"); d != "" {
        if parsed, err := strconv.Atoi(d); err == nil {
            days = parsed
        }
    }
    
    timeSeries, err := h.repo.GetTimeSeriesData(r.Context(), days)
    if err != nil {
        http.Error(w, "Failed to get time series data", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(timeSeries)
}
```

**Add Routes:**

```go
// In main.go or routes.go
dashboardHandler := handlers.NewDashboardHandler(repo)
router.Get("/api/dashboard/summary", dashboardHandler.GetSummary)
router.Get("/api/dashboard/timeseries", dashboardHandler.GetTimeSeries)
```

**Test Dashboard API:**

```bash
# Get 30-day summary
curl http://localhost:8080/api/dashboard/summary

# Expected:
{
  "metrics": [
    {"metric": "total_classifications", "value": 1523, "description": "..."},
    {"metric": "cache_hit_rate", "value": 67.5, "description": "..."},
    {"metric": "avg_confidence", "value": 0.8923, "description": "..."},
    {"metric": "layer1_percentage", "value": 78.2, "description": "..."},
    {"metric": "layer3_percentage", "value": 6.8, "description": "..."}
  ]
}
```

---

## Day 3: UI Integration

### Display Explanations

**Current UI:** Shows only industry name and confidence

**Target UI:** Shows full explanation, codes, reasoning

**File:** `frontend/templates/classification_result.html` (or React component)

```html
<div class="classification-result">
    <!-- Primary Classification -->
    <div class="result-header">
        <h2>{{ classification.primary_industry }}</h2>
        <div class="confidence-badge" :class="confidenceClass">
            {{ (classification.confidence * 100).toFixed(1) }}% Confidence
        </div>
    </div>
    
    <!-- Explanation Section -->
    <div class="explanation-section">
        <h3>Why this classification?</h3>
        <p class="primary-reason">{{ explanation.primary_reason }}</p>
        
        <h4>Supporting Factors:</h4>
        <ul class="supporting-factors">
            <li v-for="factor in explanation.supporting_factors">
                {{ factor }}
            </li>
        </ul>
        
        <div class="processing-info">
            <span class="layer-badge">{{ explanation.processing_path }}</span>
            <span class="method">Method: {{ classification.method }}</span>
            <span class="time">{{ processing_time_ms }}ms</span>
            <span v-if="from_cache" class="cache-badge">Cached</span>
        </div>
    </div>
    
    <!-- Industry Codes -->
    <div class="codes-section">
        <h3>Industry Codes</h3>
        
        <div class="code-group">
            <h4>MCC (Merchant Category Codes)</h4>
            <div class="code-list">
                <div v-for="code in codes.mcc" class="code-item">
                    <span class="code-number">{{ code.code }}</span>
                    <span class="code-description">{{ code.description }}</span>
                    <span class="code-confidence">{{ (code.confidence * 100).toFixed(0) }}%</span>
                </div>
            </div>
        </div>
        
        <div class="code-group">
            <h4>SIC (Standard Industrial Classification)</h4>
            <div class="code-list">
                <div v-for="code in codes.sic" class="code-item">
                    <span class="code-number">{{ code.code }}</span>
                    <span class="code-description">{{ code.description }}</span>
                    <span class="code-confidence">{{ (code.confidence * 100).toFixed(0) }}%</span>
                </div>
            </div>
        </div>
        
        <div class="code-group">
            <h4>NAICS (North American Industry Classification)</h4>
            <div class="code-list">
                <div v-for="code in codes.naics" class="code-item">
                    <span class="code-number">{{ code.code }}</span>
                    <span class="code-description">{{ code.description }}</span>
                    <span class="code-confidence">{{ (code.confidence * 100).toFixed(0) }}%</span>
                </div>
            </div>
        </div>
    </div>
    
    <!-- Alternative Classifications (if present) -->
    <div v-if="explanation.alternative_classifications.length > 0" class="alternatives-section">
        <h3>Alternative Classifications</h3>
        <ul>
            <li v-for="alt in explanation.alternative_classifications">
                {{ alt }}
            </li>
        </ul>
    </div>
    
    <!-- Key Terms Found (Layer 1) -->
    <div v-if="explanation.key_terms_found.length > 0" class="keywords-section">
        <h3>Key Terms Identified</h3>
        <div class="keyword-tags">
            <span v-for="term in explanation.key_terms_found" class="keyword-tag">
                {{ term }}
            </span>
        </div>
    </div>
</div>

<style>
.classification-result {
    max-width: 1000px;
    margin: 0 auto;
    padding: 20px;
}

.result-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 30px;
    padding-bottom: 20px;
    border-bottom: 2px solid #e0e0e0;
}

.result-header h2 {
    font-size: 32px;
    color: #1a237e;
    margin: 0;
}

.confidence-badge {
    padding: 8px 16px;
    border-radius: 20px;
    font-weight: bold;
    font-size: 18px;
}

.confidence-badge.high { background: #4caf50; color: white; }
.confidence-badge.medium { background: #ff9800; color: white; }
.confidence-badge.low { background: #f44336; color: white; }

.explanation-section {
    background: #f5f5f5;
    padding: 20px;
    border-radius: 8px;
    margin-bottom: 30px;
}

.primary-reason {
    font-size: 18px;
    line-height: 1.6;
    color: #333;
    margin-bottom: 20px;
}

.supporting-factors {
    list-style: none;
    padding: 0;
}

.supporting-factors li {
    padding: 10px;
    margin-bottom: 8px;
    background: white;
    border-left: 4px solid #1a237e;
    border-radius: 4px;
}

.processing-info {
    display: flex;
    gap: 12px;
    margin-top: 20px;
    flex-wrap: wrap;
}

.layer-badge, .method, .time, .cache-badge {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 14px;
    background: #e3f2fd;
    color: #1976d2;
}

.cache-badge {
    background: #c8e6c9;
    color: #2e7d32;
}

.codes-section {
    margin-bottom: 30px;
}

.code-group {
    margin-bottom: 24px;
}

.code-group h4 {
    color: #666;
    font-size: 14px;
    text-transform: uppercase;
    margin-bottom: 12px;
}

.code-list {
    display: grid;
    gap: 8px;
}

.code-item {
    display: grid;
    grid-template-columns: 100px 1fr auto;
    align-items: center;
    padding: 12px;
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 4px;
}

.code-number {
    font-weight: bold;
    color: #1a237e;
}

.code-description {
    color: #555;
}

.code-confidence {
    color: #4caf50;
    font-weight: bold;
}

.alternatives-section {
    background: #fff3e0;
    padding: 20px;
    border-radius: 8px;
    margin-bottom: 30px;
}

.keyword-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
}

.keyword-tag {
    padding: 6px 12px;
    background: #e1f5fe;
    color: #0277bd;
    border-radius: 16px;
    font-size: 14px;
}
</style>

<script>
export default {
    computed: {
        confidenceClass() {
            const conf = this.classification.confidence;
            if (conf >= 0.85) return 'high';
            if (conf >= 0.70) return 'medium';
            return 'low';
        }
    }
}
</script>
```

---

## Day 4: Performance Optimization

### Add Request Timeout

**File:** `services/classification-service/main.go`

```go
func main() {
    router := chi.NewRouter()
    
    // Add timeout middleware
    router.Use(middleware.Timeout(30 * time.Second))
    
    // ... routes ...
}
```

### Add Rate Limiting

```go
import "golang.org/x/time/rate"

type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(rps int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Limit(rps), rps*2),
    }
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !rl.limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Use in main.go
rateLimiter := NewRateLimiter(100) // 100 requests per second
router.Use(rateLimiter.Middleware)
```

### Optimize Database Connections

```go
// In repository initialization
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Add Connection Pooling for HTTP Clients

```go
// For embedding and LLM services
httpClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

---

## Day 5-6: Testing & Validation

### Performance Testing Script

```bash
#!/bin/bash
# performance_test.sh

API_URL="http://localhost:8080/v1/classify"

echo "Running performance tests..."
echo "============================="
echo ""

# Test 1: Cache performance
echo "Test 1: Cache Performance"
echo "First request (cache miss):"
time curl -s -X POST $API_URL \
  -d '{"website_url": "https://mcdonalds.com"}' | jq '.processing_time_ms'

echo "Second request (cache hit):"
time curl -s -X POST $API_URL \
  -d '{"website_url": "https://mcdonalds.com"}' | jq '.processing_time_ms, .from_cache'

echo ""

# Test 2: Layer distribution
echo "Test 2: Layer Distribution (100 requests)"
for i in {1..100}; do
    curl -s -X POST $API_URL \
      -d "{\"business_name\": \"Test Business $i\"}" \
      | jq -r '.explanation.processing_path'
done | sort | uniq -c

echo ""

# Test 3: Confidence distribution
echo "Test 3: Confidence Distribution"
for i in {1..50}; do
    curl -s -X POST $API_URL \
      -d "{\"business_name\": \"Test Business $i\"}" \
      | jq -r '.classification.confidence'
done | awk '{sum+=$1; count++} END {print "Average:", sum/count}'

echo ""

# Test 4: Rate limiting
echo "Test 4: Rate Limiting (rapid requests)"
for i in {1..150}; do
    curl -s -X POST $API_URL \
      -d '{"business_name": "Test"}' \
      > /dev/null &
done
wait
echo "Check logs for rate limit messages"
```

### Accuracy Validation

```python
# validate_accuracy.py
import requests
import json
from typing import List, Dict

API_URL = "http://localhost:8080/v1/classify"

# Load test cases with expected results
test_cases = [
    {
        "url": "https://mcdonalds.com",
        "expected_industry": "Restaurants",
        "expected_mcc": "5814"
    },
    # Add 50-100 test cases
]

def validate_accuracy(test_cases: List[Dict]) -> Dict:
    results = {
        "total": len(test_cases),
        "correct": 0,
        "incorrect": 0,
        "errors": 0,
        "by_layer": {
            "layer1": {"total": 0, "correct": 0},
            "layer2": {"total": 0, "correct": 0},
            "layer3": {"total": 0, "correct": 0}
        },
        "confidence_dist": {
            "high": 0,  # >= 0.90
            "medium": 0,  # 0.70-0.90
            "low": 0  # < 0.70
        }
    }
    
    for test in test_cases:
        try:
            response = requests.post(API_URL, json={
                "website_url": test["url"]
            })
            
            if response.status_code != 200:
                results["errors"] += 1
                continue
            
            data = response.json()
            
            # Check accuracy
            primary_industry = data["classification"]["primary_industry"]
            if test["expected_industry"].lower() in primary_industry.lower():
                results["correct"] += 1
                
                # Track by layer
                layer = data["explanation"]["processing_path"]
                if "layer1" in layer:
                    results["by_layer"]["layer1"]["correct"] += 1
                elif "layer2" in layer:
                    results["by_layer"]["layer2"]["correct"] += 1
                elif "layer3" in layer:
                    results["by_layer"]["layer3"]["correct"] += 1
            else:
                results["incorrect"] += 1
            
            # Track layer usage
            layer = data["explanation"]["processing_path"]
            if "layer1" in layer:
                results["by_layer"]["layer1"]["total"] += 1
            elif "layer2" in layer:
                results["by_layer"]["layer2"]["total"] += 1
            elif "layer3" in layer:
                results["by_layer"]["layer3"]["total"] += 1
            
            # Track confidence
            conf = data["classification"]["confidence"]
            if conf >= 0.90:
                results["confidence_dist"]["high"] += 1
            elif conf >= 0.70:
                results["confidence_dist"]["medium"] += 1
            else:
                results["confidence_dist"]["low"] += 1
                
        except Exception as e:
            print(f"Error testing {test['url']}: {e}")
            results["errors"] += 1
    
    # Calculate percentages
    results["accuracy_pct"] = (results["correct"] / results["total"]) * 100
    
    for layer in results["by_layer"]:
        if results["by_layer"][layer]["total"] > 0:
            results["by_layer"][layer]["accuracy_pct"] = (
                results["by_layer"][layer]["correct"] / 
                results["by_layer"][layer]["total"]
            ) * 100
    
    return results

# Run validation
results = validate_accuracy(test_cases)

print(json.dumps(results, indent=2))
print(f"\nOverall Accuracy: {results['accuracy_pct']:.2f}%")
print(f"Target: 90-95% ✓" if results['accuracy_pct'] >= 90 else "Target: 90-95% ✗")
```

---

## Day 7: Production Deployment Checklist

### Pre-Deployment

- [ ] All tests passing (accuracy ≥90%)
- [ ] Performance benchmarks met (p95 <3s)
- [ ] Cache working (hit rate >60%)
- [ ] Monitoring dashboard functional
- [ ] UI showing explanations
- [ ] Error handling tested
- [ ] Rate limiting configured
- [ ] Database indexes optimized

### Security

- [ ] API authentication enabled
- [ ] CORS configured properly
- [ ] Sensitive data encrypted
- [ ] SQL injection prevented (parameterized queries)
- [ ] XSS prevention in UI
- [ ] HTTPS enforced
- [ ] Environment variables secured

### Documentation

- [ ] API documentation (Swagger/OpenAPI)
- [ ] Setup guide for new developers
- [ ] Architecture diagram
- [ ] Troubleshooting guide
- [ ] Deployment runbook

### Monitoring

- [ ] Error alerting configured
- [ ] Performance monitoring (Railway metrics)
- [ ] Database monitoring (Supabase dashboard)
- [ ] Log aggregation working
- [ ] Uptime monitoring

### Backup & Recovery

- [ ] Database backups configured (Supabase auto-backup)
- [ ] Disaster recovery plan documented
- [ ] Rollback procedure tested

---

## Phase 5 Success Criteria

- [ ] ✅ 30-day classification cache implemented
- [ ] ✅ Cache hit rate >60%
- [ ] ✅ Monitoring dashboard live
- [ ] ✅ Metrics tracked for all classifications
- [ ] ✅ UI shows full explanations and codes
- [ ] ✅ Performance optimized (p95 <2s with cache)
- [ ] ✅ Production deployment checklist complete
- [ ] ✅ Documentation finalized
- [ ] ✅ **System ready for production use**

---

## Final System Metrics

**Expected Production Performance:**
```
Accuracy: 90-95%
Cache hit rate: 60-70%
Latency (cached): <100ms
Latency (uncached, p95): <3000ms
Layer distribution:
  - Layer 1: 70-85%
  - Layer 2: 10-20%
  - Layer 3: 5-10%
Uptime: 99.9%
```

**Cost (Monthly):**
```
Classification Service: $7
Playwright: $5
Embedding Service: $15
LLM Service: $30
Supabase: $0-25
────────────────────────
Total: ~$60-80/month
Cost per classification: $0.006-0.008
```

---

## Congratulations! 🎉

You've built a production-ready, 3-layer AI classification system:

- ✅ 90-95% accuracy
- ✅ Intelligent caching
- ✅ Comprehensive monitoring
- ✅ Beautiful UI with explanations
- ✅ Cost-effective (80% cheaper than APIs)
- ✅ Scalable and performant

**Your system is ready to classify businesses at scale!**
