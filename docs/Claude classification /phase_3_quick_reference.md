# Phase 3 Quick Reference
## 80-85% → 85-90% Accuracy in 2 Weeks

**Status:** Phase 2 ✅ Complete | Phase 3 ⏳ In Progress

---

## 📋 10-Day Execution Plan

### Week 5: Infrastructure (The Foundation)
**Day 1:** Enable pgvector in Supabase  
**Day 2:** Pre-compute code embeddings (one-time script)  
**Day 3:** Deploy embedding service to Railway  
**Day 4:** Add vector search RPC functions  
**Day 5:** Test infrastructure end-to-end

### Week 6: Integration (The Brains)
**Day 6-7:** Build Go embedding classifier  
**Day 8:** Add Layer 2 routing logic  
**Day 9:** Integration testing  
**Day 10:** Full test suite validation

---

## 🎯 What Phase 3 Adds

| Component | What It Does | Why It Matters |
|-----------|--------------|----------------|
| **pgvector** | Vector similarity search in PostgreSQL | Fast semantic matching |
| **Code embeddings** | 384-dim vectors for all codes | Pre-computed for speed |
| **Embedding service** | Python microservice | Generates embeddings |
| **Layer 2 routing** | Falls back when Layer 1 uncertain | Catches edge cases |

**The Magic:** Embeddings understand **meaning**, not just keywords.

```
Layer 1 (Keywords): "restaurant" → ✅ Restaurant
Layer 1 (Keywords): "cloud-native DevOps" → ❌ Confused

Layer 2 (Embeddings): "cloud-native DevOps" → ✅ Technology Consulting
```

---

## 📁 Files You'll Create

```
supabase-migrations/
└── 050_enable_pgvector.sql          [NEW] Enable pgvector + tables

scripts/
├── precompute_embeddings.py         [NEW] One-time: Generate embeddings
└── requirements.txt                  [NEW] Python dependencies

services/embedding-service/
├── app.py                            [NEW] FastAPI embedding service
├── requirements.txt                  [NEW] Service dependencies
├── Dockerfile                        [NEW] Container config
└── .dockerignore                     [NEW] Build exclusions

internal/classification/
├── embedding_classifier.go           [NEW] Layer 2 implementation
├── service.go                        [MODIFY] Add Layer 2 routing
└── repository/
    └── supabase_repository.go        [ADD] Vector search methods
```

---

## ✅ Daily Checklist

### Day 1: Enable pgvector
- [ ] Create `supabase-migrations/050_enable_pgvector.sql`
- [ ] Add `CREATE EXTENSION vector`
- [ ] Create `code_embeddings` table (384-dim vectors)
- [ ] Create IVFFlat index for similarity search
- [ ] Add `match_code_embeddings()` RPC function
- [ ] Run migration on Supabase
- [ ] Verify: `SELECT * FROM pg_extension WHERE extname = 'vector'` ✅

### Day 2: Pre-compute Embeddings
- [ ] Create `scripts/precompute_embeddings.py`
- [ ] Install: `pip install sentence-transformers supabase tqdm`
- [ ] Load model: `all-MiniLM-L6-v2` (384-dim)
- [ ] Fetch all codes from database (~1500 codes)
- [ ] Enrich with keywords and context
- [ ] Generate embeddings in batches (50 at a time)
- [ ] Insert into `code_embeddings` table
- [ ] Verify: `SELECT COUNT(*) FROM code_embeddings` → ~1500 ✅
- [ ] **Time required:** 15-30 minutes

### Day 3: Deploy Embedding Service
- [ ] Create `services/embedding-service/app.py` (FastAPI)
- [ ] Add endpoints: `/embed`, `/embed/batch`, `/health`
- [ ] Create Dockerfile
- [ ] Push to GitHub: `git push origin phase-3-embeddings`
- [ ] Deploy to Railway (auto-detects Dockerfile)
- [ ] Configure: 2GB memory, 1 CPU
- [ ] Test: `curl https://your-service.up.railway.app/health`
- [ ] Add env var: `EMBEDDING_SERVICE_URL=https://...`
- [ ] Verify: <100ms response time ✅

### Day 4: Vector Search Functions
- [ ] Already done in Day 1 migration!
- [ ] Test vector search:
  ```sql
  SELECT * FROM match_code_embeddings(
    (SELECT embedding FROM code_embeddings WHERE code = '5812' LIMIT 1),
    'MCC', 0.7, 5
  );
  ```
- [ ] Verify: Returns similar codes ✅
- [ ] Test performance: <10ms execution ✅

### Day 5: Infrastructure Testing
- [ ] Test embedding service: Generate embedding for sample text
- [ ] Test vector search: Query with that embedding
- [ ] Test end-to-end: Text → Embedding → Vector search → Similar codes
- [ ] Measure latency: Embedding (50-100ms) + Vector search (5-10ms)
- [ ] Verify accuracy: Similar codes make sense ✅

### Day 6-7: Go Embedding Classifier
- [ ] Create `internal/classification/embedding_classifier.go`
- [ ] Implement `ClassifyByEmbedding()` method
- [ ] Add `prepareTextForEmbedding()` (prioritize title/meta/about)
- [ ] Add `getEmbedding()` (call embedding service)
- [ ] Add `searchSimilarCodes()` (call Supabase RPC)
- [ ] Add `calculateConfidence()` (based on similarity scores)
- [ ] Add repository method: `MatchCodeEmbeddings()`
- [ ] Test directly: Should return top 3 codes per type ✅

### Day 8: Layer 2 Routing
- [ ] Modify `internal/classification/service.go`
- [ ] Add embedding classifier to service struct
- [ ] Update `DetectIndustry()` with Layer 2 logic:
  ```go
  if layer1Confidence >= 0.90:
      return layer1Result  // High confidence, done
  else:
      layer2Result = embeddingClassifier.Classify()
      if layer2Confidence > layer1Confidence + 0.05:
          return layer2Result  // Embedding is better
      else:
          return layer1Result  // Layer 1 good enough
  ```
- [ ] Add logging to track layer usage
- [ ] Test: Verify Layer 2 triggers for low-confidence cases ✅

### Day 9: Integration Testing
- [ ] Test standard business (should use Layer 1):
  ```bash
  curl -X POST /classify -d '{"url": "https://mcdonalds.com"}'
  # Expected: Layer 1, confidence 0.92+
  ```
- [ ] Test edge case (should use Layer 2):
  ```bash
  curl -X POST /classify -d '{"description": "Cloud-native DevOps"}'
  # Expected: Layer 1 → Layer 2, confidence 0.85+
  ```
- [ ] Test ambiguous business (Layer 2 should help)
- [ ] Verify layer distribution: ~20-30% use Layer 2 ✅

### Day 10: Full Validation
- [ ] Run full test set (all test cases)
- [ ] Calculate accuracy: **Target 85-90%**
- [ ] Measure layer usage:
  - Layer 1 only: 70-80%
  - Layer 2: 20-30%
- [ ] Measure performance:
  - p50: <200ms
  - p90: <600ms
  - p95: <900ms
- [ ] **Accuracy ≥85%** ✅

---

## 🚀 Day 1 Quick Start

**Your immediate action:**

```bash
# 1. Create migration file
cat > supabase-migrations/050_enable_pgvector.sql << 'EOF'
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create code_embeddings table
CREATE TABLE code_embeddings (
    id BIGSERIAL PRIMARY KEY,
    code_type VARCHAR(10) NOT NULL,
    code VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    extended_description TEXT,
    embedding vector(384),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(code_type, code)
);

-- Create IVFFlat index
CREATE INDEX idx_code_embeddings_vector ON code_embeddings 
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Add similarity search function
CREATE OR REPLACE FUNCTION match_code_embeddings(
    query_embedding vector(384),
    code_type_filter text,
    match_threshold float DEFAULT 0.7,
    match_count int DEFAULT 5
)
RETURNS TABLE (code text, description text, similarity float)
LANGUAGE plpgsql AS $$
BEGIN
    RETURN QUERY
    SELECT ce.code, ce.description,
           1 - (ce.embedding <=> query_embedding) as similarity
    FROM code_embeddings ce
    WHERE ce.code_type = code_type_filter
        AND 1 - (ce.embedding <=> query_embedding) > match_threshold
    ORDER BY ce.embedding <=> query_embedding
    LIMIT match_count;
END;
$$;
EOF

# 2. Run migration
psql $SUPABASE_DB_URL -f supabase-migrations/050_enable_pgvector.sql

# 3. Verify
psql $SUPABASE_DB_URL -c "SELECT * FROM pg_extension WHERE extname = 'vector';"
```

**Expected:** pgvector extension enabled ✅

---

## 📊 Success Metrics

### Before Phase 3 (After Phase 2)
```
Accuracy: 80-85%
Layers: 1 (Multi-strategy)
Edge cases: Struggles
Novel terms: Poor
Latency p95: ~500ms
```

### After Phase 3
```
Accuracy: ✅ 85-90%
Layers: ✅ 2 (Multi-strategy + Embeddings)
Edge cases: ✅ Much improved
Novel terms: ✅ Good handling
Latency p95: ✅ <900ms
```

**Layer 2 Impact:**
- Triggers: 20-30% of requests
- Improves: ~10-15% of classifications
- Adds latency: 300-500ms (when used)
- Accuracy gain: +5-10 percentage points

---

## 🧪 Testing Commands

### Test Infrastructure

**Test embedding service:**
```bash
EMBED_URL="https://your-embedding-service.up.railway.app"

# Health check
curl $EMBED_URL/health

# Generate embedding
curl -X POST $EMBED_URL/embed \
  -H "Content-Type: application/json" \
  -d '{"text": "Italian restaurant serving pizza"}'

# Expected: 384-dim vector in <100ms
```

**Test vector search:**
```sql
-- Get an embedding from a known code
\x
SELECT * FROM match_code_embeddings(
    (SELECT embedding FROM code_embeddings 
     WHERE code = '5812' AND code_type = 'MCC' LIMIT 1),
    'MCC',
    0.7,
    5
);

-- Expected: Similar restaurant codes (5814, 5813, etc.)
```

### Test Classification

**Standard case (Layer 1):**
```bash
curl -X POST http://localhost:8080/v1/classify \
  -d '{"business_name": "Pizza Restaurant", "website_url": "https://joespizza.com"}'

# Expected:
# - method: "fast_path_keyword" or "multi_strategy"  
# - confidence: 0.90+
# - processing_time_ms: <150ms
```

**Edge case (Layer 2):**
```bash
curl -X POST http://localhost:8080/v1/classify \
  -d '{"business_name": "Cloud DevOps Consultancy", 
       "description": "Kubernetes orchestration and CI/CD automation"}'

# Expected:
# - method: "embedding_similarity"
# - confidence: 0.82-0.88
# - processing_time_ms: 400-800ms
# - explanation: "Semantic similarity analysis matched..."
```

---

## 💡 Key Concepts

### What Are Embeddings?

Embeddings are dense vector representations of text that capture semantic meaning.

```
"Restaurant" → [0.23, -0.45, 0.67, ..., 0.12]  (384 numbers)
"Dining establishment" → [0.25, -0.43, 0.69, ..., 0.11]  (similar!)

Cosine similarity: 0.92 (very similar)
```

**Why This Matters:**
- Understands synonyms
- Captures context
- Handles jargon
- Not limited to exact keywords

### Layer 1 vs Layer 2

**Layer 1 (Keywords/Trigrams):**
- Fast: <100-500ms
- Works for: 70-80% of cases
- Best at: Exact matches, common terms
- Struggles with: Novel terminology, context

**Layer 2 (Embeddings):**
- Slower: 400-800ms
- Works for: Edge cases (20-30%)
- Best at: Semantic understanding, jargon
- Catches: What Layer 1 misses

**Routing Logic:**
```
IF Layer 1 confidence ≥ 90%:
    ✅ Use Layer 1 (fast, confident)
ELSE IF Layer 1 confidence ≥ 80%:
    ✅ Use Layer 1 (good enough)
ELSE:
    🔄 Try Layer 2 (might be better)
    IF Layer 2 confidence > Layer 1 + 5%:
        ✅ Use Layer 2
    ELSE:
        ✅ Use Layer 1 (explainable)
```

### The pgvector Index

**IVFFlat** = Inverted File with Flat compression

- Approximate nearest neighbor search
- Trade-off: Speed vs. accuracy (we choose speed)
- `lists = 100` means 100 clusters
- Query time: O(√n) instead of O(n)

**Result:** Sub-10ms vector search on 1500 codes ⚡

---

## 🎓 Pro Tips

**Tip 1: Pre-compute Once**
Embedding generation is slow (50-100ms per code). Pre-computing all codes once means fast lookups forever. Only re-run if codes change.

**Tip 2: Enrich Descriptions**
Better input = better embeddings. We add keywords and context to code descriptions before generating embeddings.

```python
# Basic
"5812 - Eating Places"

# Enriched
"5812 - Eating Places. Industry: Food Service. Related terms: restaurant, dining, cafe, food. Examples: Restaurants, cafes, diners"
```

**Tip 3: Test Incrementally**
Don't wait until Day 10. Test after:
- Day 2 (embeddings generated)
- Day 3 (service deployed)
- Day 5 (infrastructure working)
- Day 7 (Go classifier working)

**Tip 4: Monitor Layer Usage**
Add logging to track:
- % of requests using Layer 2
- Layer 2 confidence vs Layer 1
- When Layer 2 improves classification

Expected: Layer 2 should trigger ~20-30% of the time.

**Tip 5: Similarity Threshold**
We use 0.70 (70% similarity). Lower = more matches but less accurate. Higher = fewer matches but more precise.

```sql
-- More permissive
match_threshold: 0.60  → ~10 matches

-- Standard
match_threshold: 0.70  → ~5 matches

-- Strict
match_threshold: 0.80  → ~2-3 matches
```

---

## ⚠️ Common Issues

**Issue: "extension 'vector' does not exist"**
```sql
-- Check PostgreSQL version
SELECT version();  -- Need 11+

-- Try as superuser
CREATE EXTENSION vector;

-- If fails, contact Supabase support
```

**Issue: Embedding service out of memory**
- Check Railway: Service needs 2GB RAM
- Model uses ~1.5GB when loaded
- Set memory limit: Railway Dashboard → Service → Settings → Memory: 2048MB

**Issue: Vector search returns no results**
```sql
-- Check if embeddings exist
SELECT COUNT(*) FROM code_embeddings;  -- Should be ~1500

-- Lower threshold
SELECT * FROM match_code_embeddings(..., 0.5, 10);  -- Try 0.5

-- Check embedding dimension matches
SELECT vector_dims(embedding) FROM code_embeddings LIMIT 1;  -- Should be 384
```

**Issue: Layer 2 never triggers**
```go
// Check threshold in service.go
if layer1Result.Confidence >= 0.90 {  // Should be 0.90, not higher

// Add debug logging
slog.Info("Layer routing decision",
    "layer1_confidence", layer1Result.Confidence,
    "threshold", 0.90,
    "will_use_layer2", layer1Result.Confidence < 0.90)
```

**Issue: Layer 2 too slow (>1s)**
- Check embedding service latency: Should be <100ms
- Check vector search query: Should use index, <10ms
- Problem usually: Embedding service not responding fast
- Solution: Verify Railway deployment, check logs

---

## 📈 Progress Tracker

| Day | Task | Status | Notes |
|-----|------|--------|-------|
| 1 | pgvector enabled | ⬜ | Migration run? |
| 2 | Embeddings computed | ⬜ | ~1500 codes? |
| 3 | Service deployed | ⬜ | Railway URL? |
| 4 | Vector search tested | ⬜ | <10ms? |
| 5 | Infrastructure validated | ⬜ | E2E working? |
| 6-7 | Go classifier built | ⬜ | Tests passing? |
| 8 | Layer 2 routing added | ⬜ | Triggers 20-30%? |
| 9 | Integration tested | ⬜ | Edge cases improved? |
| 10 | **Accuracy validated** | ⬜ | **≥85%?** ⭐ |

---

## 📞 Checkpoints

**Good times to validate:**
- After Day 2: "Embeddings computed successfully"
- After Day 3: "Embedding service responding"
- After Day 5: "Infrastructure end-to-end working"
- After Day 8: "Layer 2 routing integrated"
- After Day 10: "Accuracy at 85-90%"

**Red flags:**
- Day 2: <1000 embeddings (should have ~1500)
- Day 3: Embedding service >200ms (should be <100ms)
- Day 5: Vector search >50ms (should be <10ms)
- Day 8: Layer 2 never triggers (check routing logic)
- Day 10: Accuracy <83% (troubleshoot Layer 2)

---

## 🎉 Phase 3 Complete!

**You'll know it's done when:**
- ✅ pgvector enabled with 1500+ code embeddings
- ✅ Embedding service deployed and responding fast
- ✅ Layer 2 triggers for 20-30% of requests
- ✅ Edge cases and novel terminology improved
- ✅ **Accuracy at 85-90% on test set**

**Then you're ready for Phase 4:** LLM-based Layer 3 (90-95% accuracy)

---

## 🔥 Let's Build Layer 2!

You've got a solid 2-layer system with multi-strategy (Layer 1) working great. Now add semantic understanding (Layer 2) and watch accuracy climb to 85-90%!

Start with Day 1: Enable pgvector in Supabase. The complete SQL is in the Phase 3 Kick-Off Guide.

Ready to start? 💪
