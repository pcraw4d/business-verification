# Phase 2 Quick Reference
## 50-60% → 80-85% Accuracy in 2 Weeks

**Status:** Phase 1 ✅ Complete | Phase 2 ⏳ In Progress

---

## 📋 Week-by-Week Breakdown

### Week 3: Core Enhancements
**Day 1-2:** Return top 3 codes (not just 1)  
**Day 3-4:** Improve confidence calibration  
**Day 5:** Optimize database queries

### Week 4: Performance & Polish
**Day 1-2:** Implement fast path (<100ms)  
**Day 3-4:** Generate structured explanations  
**Day 4-5:** Fix "General Business" fallback  
**Day 5:** Test on full test set

---

## 🎯 What Gets Fixed This Phase

| Issue | Current State | Target State |
|-------|---------------|--------------|
| **Codes returned** | Only 1 per type | **Top 3 with confidence** |
| **Confidence** | 50-70% (too low) | **70-95% (calibrated)** |
| **Speed** | 200-500ms all requests | **<100ms for 70% of requests** |
| **Explanations** | Exists but not structured | **Ready for UI display** |
| **Generic fallback** | "General Business" common | **Specific industries preferred** |
| **Accuracy** | 50-60% | **80-85%** |

---

## 📁 Files You'll Modify

### Core Changes
```
internal/classification/
├── classifier.go                 [MAJOR] Return top 3 codes
├── confidence_calibrator.go      [NEW] Calibration logic
├── multi_strategy_classifier.go  [MAJOR] Add fast path
├── explanation_generator.go      [NEW] Generate explanations
├── service.go                    [MODIFY] Integrate new components
└── repository/
    └── supabase_repository.go    [ADD] New query methods
```

### Supporting Changes
```
services/classification-service/internal/handlers/
└── classification.go             [MODIFY] Update API response

supabase-migrations/
└── 040_optimize_queries.sql      [NEW] Performance optimization
```

---

## 🚀 Quick Start Checklist

**Before You Start:**
- [ ] Phase 1 validated (95%+ scrape success)
- [ ] Test set ready in Cursor/Supabase
- [ ] Create branch: `git checkout -b phase-2-layer1-enhancements`

**Day 1-2: Top 3 Codes**
- [ ] Update `CodeResult` struct (add `Source` field)
- [ ] Update `ClassificationCodes` to return arrays
- [ ] Modify `GenerateCodes()` function
- [ ] Implement `getMCCCandidates()` (3 strategies: direct, keyword, trigram)
- [ ] Implement `selectTopCodes()` (sort by confidence, return top 3)
- [ ] Add `enrichWithCrosswalks()` (boost with code relationships)
- [ ] Update API response structure
- [ ] **Test:** Returns 3 codes per type ✅

**Day 3-4: Confidence Calibration**
- [ ] Create `confidence_calibrator.go` file
- [ ] Implement `CalibrateConfidence()` (5 factors)
- [ ] Implement `CalculateCodeAgreement()` (crosswalk validation)
- [ ] Integrate into `DetectIndustry()` service method
- [ ] **Test:** Confidence scores 70-95% ✅

**Day 5: Database Optimization**
- [ ] Create migration `040_optimize_queries.sql`
- [ ] Add composite indexes
- [ ] Create materialized view for code search
- [ ] Run migration on Supabase
- [ ] **Test:** Query performance <50ms ✅

**Day 6-7: Fast Path**
- [ ] Add `tryFastPath()` method
- [ ] Implement `extractObviousKeywords()` (50+ obvious keywords)
- [ ] Add `GetIndustriesByKeyword()` repository method
- [ ] Integrate into main classification flow
- [ ] **Test:** Fast path handles 60-70%, <100ms ✅

**Day 8-9: Explanations**
- [ ] Create `explanation_generator.go` file
- [ ] Implement `GenerateExplanation()` method
- [ ] Add `generatePrimaryReason()` (6+ templates)
- [ ] Add `generateSupportingFactors()` (5-7 factor types)
- [ ] Integrate into service and API response
- [ ] **Test:** Explanations present and useful ✅

**Day 9-10: Generic Fallback Fix**
- [ ] Modify `selectBestIndustry()` method
- [ ] Add generic industry detection
- [ ] Prefer specific over generic (confidence adjustment)
- [ ] Higher confidence threshold for generic industries
- [ ] **Test:** "General Business" minimized ✅

**Day 10: Full Test Suite**
- [ ] Run classification on full test set
- [ ] Measure accuracy improvement
- [ ] Check performance distribution (p50, p90, p95)
- [ ] Validate all success criteria
- [ ] **Accuracy target: 80-85%** ✅

---

## ✅ Phase 2 Success Criteria

### Must Pass Before Phase 3:

**Functionality:**
- [ ] Returns exactly 3 codes per type (MCC/SIC/NAICS)
- [ ] Each code has confidence score (0.0-1.0)
- [ ] Crosswalk validation enriches codes
- [ ] Code agreement score calculated

**Confidence:**
- [ ] Confidence range: 60-95% (not all ~50%)
- [ ] High-quality cases: 85-95%
- [ ] Ambiguous cases: 65-80%
- [ ] Confidence correlates with accuracy

**Performance:**
- [ ] Fast path: 60-70% of requests
- [ ] Fast path latency: <100ms (p95)
- [ ] Full strategy: <500ms (p95)
- [ ] Database queries: <50ms average

**Explanations:**
- [ ] Primary reason exists
- [ ] 3-5 supporting factors
- [ ] Key terms listed
- [ ] Processing path indicated

**Quality:**
- [ ] "General Business" < 10% of results
- [ ] Specific industries preferred
- [ ] **Accuracy: 80-85% on test set** ⭐

---

## 📊 Expected Results

### Before Phase 2 (After Phase 1)
```
Scrape success: ✅ 95%
Codes returned: 1 per type
Confidence: 50-70%
Speed: 200-500ms (all)
Explanations: Not structured
Generic rate: 25%
Accuracy: 50-60%
```

### After Phase 2
```
Scrape success: ✅ 95% (maintained)
Codes returned: ✅ Top 3 per type
Confidence: ✅ 70-95%
Speed: ✅ <100ms (70%), <500ms (95%)
Explanations: ✅ Structured
Generic rate: ✅ <10%
Accuracy: ✅ 80-85%
```

**Performance Distribution:**
- **Fast path (60-70%):** <100ms, confidence 90-95%
- **Full strategy (30-40%):** 200-500ms, confidence 70-90%

---

## 🔧 Testing Commands

### Test Top 3 Codes
```bash
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"website_url": "https://mcdonalds.com"}' | jq '.codes'

# Expected:
# {
#   "mcc": [
#     {"code": "5814", "description": "...", "confidence": 0.92},
#     {"code": "5812", "description": "...", "confidence": 0.85},
#     {"code": "5813", "description": "...", "confidence": 0.73}
#   ],
#   "sic": [...], // 3 codes
#   "naics": [...] // 3 codes
# }
```

### Test Fast Path
```bash
curl -X POST http://localhost:8080/v1/classify \
  -d '{"business_name": "Pizza Restaurant", "website_url": "https://joespizza.com"}'

# Check logs for:
# INFO: Fast path succeeded duration_ms=45 method=fast_path_keyword

# Expected:
# - processing_time_ms < 100
# - explanation.processing_path = "fast_path"
```

### Test Explanation
```bash
curl -X POST http://localhost:8080/v1/classify \
  -d '{"website_url": "https://example.com"}' | jq '.explanation'

# Expected structure:
# {
#   "primary_reason": "...",
#   "supporting_factors": ["...", "...", "..."],
#   "key_terms_found": ["...", "..."],
#   "method_used": "multi_strategy",
#   "processing_path": "full_strategy"
# }
```

### Database Query Performance
```sql
-- Test keyword query speed
EXPLAIN ANALYZE
SELECT DISTINCT ck.code, cc.description, MAX(ck.weight)
FROM code_keywords ck
JOIN classification_codes cc ON cc.code = ck.code
WHERE ck.code_type = 'MCC' AND ck.keyword = ANY(ARRAY['restaurant', 'food'])
GROUP BY ck.code, cc.description
ORDER BY MAX(ck.weight) DESC LIMIT 10;

-- Should show: <50ms execution time, index usage
```

---

## 💡 Pro Tips

**Tip 1: Test Incrementally**
Don't wait until Day 10 to test. After each task, run a few test cases to verify the change works.

**Tip 2: Check Logs**
Use `slog` extensively. You should see:
- "Fast path succeeded" for obvious cases
- Strategy scores for full multi-strategy
- Confidence before/after calibration
- Code agreement scores

**Tip 3: Validate with Test Set**
After Day 5 and Day 10, run your full test set and calculate:
```bash
# Accuracy
correct / total

# Performance percentiles
sort processing_times | calculate p50, p90, p95

# Fast path percentage
count(method="fast_path") / total
```

**Tip 4: Database Performance**
If queries are slow:
```sql
-- Check index usage
SELECT * FROM pg_stat_user_indexes WHERE tablename = 'code_keywords';

-- Refresh statistics
ANALYZE code_keywords;
ANALYZE classification_codes;
```

**Tip 5: Explanation Quality**
Read 10-20 explanations manually. They should:
- Make sense to a human
- Cite specific evidence
- Explain the reasoning
- Not be too technical

---

## 🚨 Common Issues & Solutions

**Issue: Still returning 1 code**
```go
// Check this line returns array:
codes.MCC = g.selectTopCodes(mccCandidates, 3) // Should return []CodeResult with len=3
```

**Issue: Confidence not improving**
```go
// Check calibration factors are being applied:
slog.Info("Confidence calibration",
    "before", baseConfidence,
    "after", calibratedConfidence,
    "quality_boost", contentQuality > 0.8)
```

**Issue: Fast path not triggering**
```go
// Check keyword extraction:
obviousKeywords := c.extractObviousKeywords(content)
slog.Info("Fast path attempt", "keywords", obviousKeywords, "found", len(obviousKeywords) > 0)
```

**Issue: Database queries slow**
```sql
-- Verify indexes exist:
\d+ code_keywords  -- Should show indexes

-- If missing:
CREATE INDEX idx_code_keywords_composite 
ON code_keywords (code_type, keyword, weight DESC);
```

---

## 📈 Progress Tracking

Track these metrics daily:

| Day | Task | Metric | Target | Actual |
|-----|------|--------|--------|--------|
| 1-2 | Top 3 codes | Codes returned | 3 per type | ___ |
| 3-4 | Confidence | Avg confidence | 75-85% | ___ |
| 5 | DB optimization | Query time | <50ms | ___ |
| 6-7 | Fast path | Fast path % | 60-70% | ___ |
| 6-7 | Fast path | Fast path time | <100ms | ___ |
| 8-9 | Explanations | Has explanation | 100% | ___ |
| 9-10 | Generic fix | Generic rate | <10% | ___ |
| 10 | **Final** | **Accuracy** | **80-85%** | ___ |

---

## 🎯 Day 1 Action Plan

**Your immediate next steps:**

1. **Create branch**
```bash
git checkout -b phase-2-layer1-enhancements
```

2. **Open files**
```bash
# Main file to modify:
code internal/classification/classifier.go

# You'll modify the GenerateCodes() method
# Current: Returns 1 code per type
# Target: Returns top 3 codes per type
```

3. **Start with data structures**
```go
// Step 1: Update CodeResult struct (add Source field)
type CodeResult struct {
    Code        string  `json:"code"`
    Description string  `json:"description"`
    Confidence  float64 `json:"confidence"`
    Source      string  `json:"source"` // NEW
}

// Step 2: Update return type to arrays
type ClassificationCodes struct {
    MCC   []CodeResult `json:"mcc"`   // Was: CodeResult, Now: []CodeResult
    SIC   []CodeResult `json:"sic"`
    NAICS []CodeResult `json:"naics"`
}
```

4. **Follow Phase 2 guide Task 1**
The complete implementation is in the full Phase 2 guide.

---

## 📞 When to Check In

**Good checkpoints:**
- After Day 2 (top 3 codes working)
- After Day 5 (Week 3 complete)
- After Day 10 (Phase 2 complete)

**Red flags to address immediately:**
- Still returning 1 code after Day 2
- Confidence not improved after Day 4
- Fast path not triggering after Day 7
- Accuracy not 75%+ after Day 10

---

## 🎉 Phase 2 Complete!

**You'll know Phase 2 is done when:**
- ✅ Full test set shows 80-85% accuracy
- ✅ Returns 3 codes per type
- ✅ Confidence calibrated (70-95% range)
- ✅ Fast path handling 60-70% in <100ms
- ✅ Explanations present and useful
- ✅ "General Business" < 10% of results

**Then move to Phase 3:** Embedding-based similarity (85-90% accuracy)

Ready to start? Open `classifier.go` and let's return those top 3 codes! 💪
