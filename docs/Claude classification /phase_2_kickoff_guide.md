# Phase 2 Kick-Off Guide: Enhance Layer 1
## Weeks 3-4: From 50-60% to 80-85% Accuracy

**Goal:** Make your multi-strategy classifier more decisive and informative by returning top 3 codes, improving confidence, and adding fast path optimization.

---

## Phase 1 Success Validation

Before starting Phase 2, verify your Phase 1 results:

✅ **Checklist:**
- [ ] Scrape success rate ≥95%
- [ ] Content quality score ≥0.7 for 90%+ of scrapes
- [ ] Average word count ≥200
- [ ] Playwright service deployed and working
- [ ] Comprehensive logging in place
- [ ] No "no output" errors

**If all checked:** You're ready for Phase 2! 🎉

**If issues remain:** Address them before proceeding. Phase 2 builds on Phase 1's foundation.

---

## Phase 2 Overview

### What We're Fixing

**Current State (After Phase 1):**
- ✅ Good scraping (95%+ success)
- ✅ Quality content extraction
- ❌ Only returns 1 code per type (MCC/SIC/NAICS)
- ❌ Confidence too low (~50-70%)
- ❌ No fast path (all requests take 200-500ms)
- ❌ Generic classifications still occur
- ❌ Explanations exist but not structured for UI

**Target State (After Phase 2):**
- ✅ Returns top 3 codes per type with confidence
- ✅ Calibrated confidence (70-95% range)
- ✅ Fast path handles 60-70% of requests in <100ms
- ✅ Specific industry classifications
- ✅ Structured explanations ready for UI display
- ✅ 80-85% accuracy

### Implementation Timeline

**Week 3:**
- Day 1-2: Return top 3 codes (not just 1)
- Day 3-4: Improve confidence calibration
- Day 5: Optimize database queries

**Week 4:**
- Day 1-2: Implement fast path (<100ms)
- Day 3-4: Generate structured explanations
- Day 4-5: Fix "General Business" fallback
- Day 5: Test on full test set

---

## Week 3: Core Enhancements

### Task 1: Return Top 3 Codes Per Type (Day 1-2)

**Current Issue:** Your `ClassificationCodeGenerator` only returns the single best code for each type.

**File:** `internal/classification/classifier.go`

#### Step 1: Update Data Structures

```go
// Update CodeResult to include source
type CodeResult struct {
    Code        string  `json:"code"`
    Description string  `json:"description"`
    Confidence  float64 `json:"confidence"`
    Source      string  `json:"source"` // NEW: "keyword_match", "crosswalk", "trigram_match", "ml_prediction"
}

// Update to return arrays instead of single values
type ClassificationCodes struct {
    MCC   []CodeResult `json:"mcc"`   // Top 3
    SIC   []CodeResult `json:"sic"`   // Top 3
    NAICS []CodeResult `json:"naics"` // Top 3
}
```

#### Step 2: Modify Code Generation Logic

```go
// internal/classification/classifier.go

func (g *ClassificationCodeGenerator) GenerateCodes(
    ctx context.Context,
    industryName string,
    keywords []string,
    confidence float64,
) (*ClassificationCodes, error) {
    
    codes := &ClassificationCodes{
        MCC:   make([]CodeResult, 0, 3),
        SIC:   make([]CodeResult, 0, 3),
        NAICS: make([]CodeResult, 0, 3),
    }
    
    // Generate candidates for each code type
    mccCandidates := g.getMCCCandidates(ctx, industryName, keywords)
    sicCandidates := g.getSICCandidates(ctx, industryName, keywords)
    naicsCandidates := g.getNAICSCandidates(ctx, industryName, keywords)
    
    // Select top 3 from each
    codes.MCC = g.selectTopCodes(mccCandidates, 3)
    codes.SIC = g.selectTopCodes(sicCandidates, 3)
    codes.NAICS = g.selectTopCodes(naicsCandidates, 3)
    
    // Use crosswalks to enrich and validate
    codes = g.enrichWithCrosswalks(codes)
    
    // If any type has <3 codes, use crosswalks to fill gaps
    codes = g.fillGapsWithCrosswalks(codes)
    
    return codes, nil
}

func (g *ClassificationCodeGenerator) getMCCCandidates(
    ctx context.Context,
    industryName string,
    keywords []string,
) []CodeResult {
    
    candidates := make(map[string]*CodeResult) // Use map to deduplicate
    
    // Strategy 1: Direct industry lookup
    if industryCode := g.repo.GetCodeByIndustryName(ctx, "MCC", industryName); industryCode != nil {
        key := industryCode.Code
        if _, exists := candidates[key]; !exists {
            candidates[key] = &CodeResult{
                Code:        industryCode.Code,
                Description: industryCode.Description,
                Confidence:  0.90, // High confidence from direct match
                Source:      "industry_match",
            }
        }
    }
    
    // Strategy 2: Keyword matching
    keywordCodes := g.repo.GetCodesByKeywords(ctx, "MCC", keywords)
    for _, kc := range keywordCodes {
        key := kc.Code
        if existing, exists := candidates[key]; exists {
            // Boost confidence if found through multiple strategies
            existing.Confidence = math.Min(existing.Confidence + 0.1, 0.98)
        } else {
            candidates[key] = &CodeResult{
                Code:        kc.Code,
                Description: kc.Description,
                Confidence:  kc.Weight, // Use keyword weight as confidence
                Source:      "keyword_match",
            }
        }
    }
    
    // Strategy 3: Trigram similarity (fuzzy matching)
    trigramCodes := g.repo.GetCodesByTrigramSimilarity(ctx, "MCC", industryName, 0.3, 10)
    for _, tc := range trigramCodes {
        key := tc.Code
        if existing, exists := candidates[key]; exists {
            existing.Confidence = math.Min(existing.Confidence + 0.05, 0.98)
        } else {
            candidates[key] = &CodeResult{
                Code:        tc.Code,
                Description: tc.Description,
                Confidence:  tc.Similarity * 0.7, // Scale down trigram confidence
                Source:      "trigram_match",
            }
        }
    }
    
    // Convert map to slice
    result := make([]CodeResult, 0, len(candidates))
    for _, candidate := range candidates {
        result = append(result, *candidate)
    }
    
    return result
}

// Similar implementations for getSICCandidates and getNAICSCandidates
func (g *ClassificationCodeGenerator) getSICCandidates(ctx context.Context, industryName string, keywords []string) []CodeResult {
    // Same logic as getMCCCandidates but for SIC codes
    // ... (copy and modify getMCCCandidates logic)
}

func (g *ClassificationCodeGenerator) getNAICSCandidates(ctx context.Context, industryName string, keywords []string) []CodeResult {
    // Same logic as getMCCCandidates but for NAICS codes
    // ... (copy and modify getMCCCandidates logic)
}

func (g *ClassificationCodeGenerator) selectTopCodes(candidates []CodeResult, limit int) []CodeResult {
    if len(candidates) == 0 {
        return []CodeResult{}
    }
    
    // Sort by confidence (highest first)
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].Confidence > candidates[j].Confidence
    })
    
    // Return top N
    if len(candidates) > limit {
        return candidates[:limit]
    }
    
    return candidates
}

func (g *ClassificationCodeGenerator) enrichWithCrosswalks(codes *ClassificationCodes) *ClassificationCodes {
    // If we have a high-confidence MCC code, use crosswalks to find corresponding SIC/NAICS
    if len(codes.MCC) > 0 && codes.MCC[0].Confidence > 0.8 {
        // Get crosswalks from MCC to SIC
        sicCrosswalks := g.repo.GetCrosswalks("MCC", codes.MCC[0].Code, "SIC")
        for _, xwalk := range sicCrosswalks {
            // Check if this SIC code is already in our list
            found := false
            for i, existing := range codes.SIC {
                if existing.Code == xwalk.ToCode {
                    // Boost confidence if found via crosswalk
                    codes.SIC[i].Confidence = math.Min(existing.Confidence + 0.15, 0.98)
                    found = true
                    break
                }
            }
            
            if !found && len(codes.SIC) < 3 {
                // Add new code from crosswalk
                codes.SIC = append(codes.SIC, CodeResult{
                    Code:        xwalk.ToCode,
                    Description: xwalk.ToDescription,
                    Confidence:  codes.MCC[0].Confidence * 0.85, // Slightly lower than source
                    Source:      "crosswalk_from_mcc",
                })
            }
        }
        
        // Get crosswalks from MCC to NAICS
        naicsCrosswalks := g.repo.GetCrosswalks("MCC", codes.MCC[0].Code, "NAICS")
        for _, xwalk := range naicsCrosswalks {
            found := false
            for i, existing := range codes.NAICS {
                if existing.Code == xwalk.ToCode {
                    codes.NAICS[i].Confidence = math.Min(existing.Confidence + 0.15, 0.98)
                    found = true
                    break
                }
            }
            
            if !found && len(codes.NAICS) < 3 {
                codes.NAICS = append(codes.NAICS, CodeResult{
                    Code:        xwalk.ToCode,
                    Description: xwalk.ToDescription,
                    Confidence:  codes.MCC[0].Confidence * 0.85,
                    Source:      "crosswalk_from_mcc",
                })
            }
        }
    }
    
    // Re-sort after enrichment
    sort.Slice(codes.SIC, func(i, j int) bool {
        return codes.SIC[i].Confidence > codes.SIC[j].Confidence
    })
    sort.Slice(codes.NAICS, func(i, j int) bool {
        return codes.NAICS[i].Confidence > codes.NAICS[j].Confidence
    })
    
    return codes
}

func (g *ClassificationCodeGenerator) fillGapsWithCrosswalks(codes *ClassificationCodes) *ClassificationCodes {
    // If MCC has codes but SIC doesn't, use crosswalks
    if len(codes.MCC) > 0 && len(codes.SIC) < 3 {
        for _, mcc := range codes.MCC {
            if len(codes.SIC) >= 3 {
                break
            }
            
            xwalks := g.repo.GetCrosswalks("MCC", mcc.Code, "SIC")
            for _, xw := range xwalks {
                // Check if already present
                found := false
                for _, existing := range codes.SIC {
                    if existing.Code == xw.ToCode {
                        found = true
                        break
                    }
                }
                
                if !found {
                    codes.SIC = append(codes.SIC, CodeResult{
                        Code:        xw.ToCode,
                        Description: xw.ToDescription,
                        Confidence:  mcc.Confidence * 0.80,
                        Source:      "crosswalk_gap_fill",
                    })
                    
                    if len(codes.SIC) >= 3 {
                        break
                    }
                }
            }
        }
    }
    
    // Similar logic for NAICS
    if len(codes.MCC) > 0 && len(codes.NAICS) < 3 {
        for _, mcc := range codes.MCC {
            if len(codes.NAICS) >= 3 {
                break
            }
            
            xwalks := g.repo.GetCrosswalks("MCC", mcc.Code, "NAICS")
            for _, xw := range xwalks {
                found := false
                for _, existing := range codes.NAICS {
                    if existing.Code == xw.ToCode {
                        found = true
                        break
                    }
                }
                
                if !found {
                    codes.NAICS = append(codes.NAICS, CodeResult{
                        Code:        xw.ToCode,
                        Description: xw.ToDescription,
                        Confidence:  mcc.Confidence * 0.80,
                        Source:      "crosswalk_gap_fill",
                    })
                    
                    if len(codes.NAICS) >= 3 {
                        break
                    }
                }
            }
        }
    }
    
    return codes
}
```

#### Step 3: Update Repository Methods

**File:** `internal/classification/repository/supabase_repository.go`

```go
// Add new repository methods if they don't exist

func (r *SupabaseKeywordRepository) GetCodesByKeywords(
    ctx context.Context,
    codeType string,
    keywords []string,
) []struct {
    Code        string
    Description string
    Weight      float64
} {
    
    if len(keywords) == 0 {
        return []struct {
            Code        string
            Description string
            Weight      float64
        }{}
    }
    
    // Query code_keywords table
    query := `
        SELECT DISTINCT
            ck.code,
            cc.description,
            MAX(ck.weight) as max_weight
        FROM code_keywords ck
        JOIN classification_codes cc ON cc.code = ck.code AND cc.code_type = ck.code_type
        WHERE ck.code_type = $1
            AND ck.keyword = ANY($2)
        GROUP BY ck.code, cc.description
        ORDER BY max_weight DESC
        LIMIT 10
    `
    
    var results []struct {
        Code        string  `db:"code"`
        Description string  `db:"description"`
        Weight      float64 `db:"max_weight"`
    }
    
    err := r.db.SelectContext(ctx, &results, query, codeType, pq.Array(keywords))
    if err != nil {
        slog.Error("Failed to get codes by keywords", "error", err)
        return []struct {
            Code        string
            Description string
            Weight      float64
        }{}
    }
    
    return results
}

func (r *SupabaseKeywordRepository) GetCodesByTrigramSimilarity(
    ctx context.Context,
    codeType string,
    industryName string,
    threshold float64,
    limit int,
) []struct {
    Code        string
    Description string
    Similarity  float64
} {
    
    // Use existing trigram function
    query := `
        SELECT 
            code,
            description,
            similarity(description, $2) as similarity
        FROM classification_codes
        WHERE code_type = $1
            AND similarity(description, $2) > $3
        ORDER BY similarity DESC
        LIMIT $4
    `
    
    var results []struct {
        Code        string  `db:"code"`
        Description string  `db:"description"`
        Similarity  float64 `db:"similarity"`
    }
    
    err := r.db.SelectContext(ctx, &results, query, codeType, industryName, threshold, limit)
    if err != nil {
        slog.Error("Failed to get codes by trigram", "error", err)
        return []struct {
            Code        string
            Description string
            Similarity  float64
        }{}
    }
    
    return results
}

func (r *SupabaseKeywordRepository) GetCrosswalks(
    fromCodeType string,
    fromCode string,
    toCodeType string,
) []struct {
    ToCode        string
    ToDescription string
} {
    
    query := `
        SELECT 
            xw.to_code,
            cc.description as to_description
        FROM industry_code_crosswalks xw
        JOIN classification_codes cc ON cc.code = xw.to_code AND cc.code_type = xw.to_code_type
        WHERE xw.from_code_type = $1
            AND xw.from_code = $2
            AND xw.to_code_type = $3
        LIMIT 5
    `
    
    var results []struct {
        ToCode        string `db:"to_code"`
        ToDescription string `db:"to_description"`
    }
    
    err := r.db.SelectContext(context.Background(), &results, query, fromCodeType, fromCode, toCodeType)
    if err != nil {
        slog.Error("Failed to get crosswalks", "error", err)
        return []struct {
            ToCode        string
            ToDescription string
        }{}
    }
    
    return results
}
```

#### Step 4: Update API Response

**File:** `services/classification-service/internal/handlers/classification.go`

```go
// Update response structure
type ClassificationResponse struct {
    RequestID      string                    `json:"request_id"`
    Classification ClassificationData        `json:"classification"`
    Codes          ClassificationCodesOutput `json:"codes"`
    ProcessingTime int64                     `json:"processing_time_ms"`
}

type ClassificationCodesOutput struct {
    MCC   []CodeResultOutput `json:"mcc"`
    SIC   []CodeResultOutput `json:"sic"`
    NAICS []CodeResultOutput `json:"naics"`
}

type CodeResultOutput struct {
    Code        string  `json:"code"`
    Description string  `json:"description"`
    Confidence  float64 `json:"confidence"`
}

// Update handler to return new structure
func (h *ClassificationHandler) HandleClassify(w http.ResponseWriter, r *http.Request) {
    // ... existing request parsing ...
    
    result, err := h.service.DetectIndustry(ctx, req.BusinessName, req.Description, req.WebsiteURL)
    if err != nil {
        // ... error handling ...
        return
    }
    
    response := ClassificationResponse{
        RequestID: req.RequestID,
        Classification: ClassificationData{
            PrimaryIndustry: result.Classification.PrimaryIndustry,
            Confidence:      result.Classification.Confidence,
            Method:          result.Classification.Method,
        },
        Codes: ClassificationCodesOutput{
            MCC:   convertCodesToOutput(result.Codes.MCC),
            SIC:   convertCodesToOutput(result.Codes.SIC),
            NAICS: convertCodesToOutput(result.Codes.NAICS),
        },
        ProcessingTime: result.ProcessingTimeMs,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func convertCodesToOutput(codes []CodeResult) []CodeResultOutput {
    output := make([]CodeResultOutput, len(codes))
    for i, code := range codes {
        output[i] = CodeResultOutput{
            Code:        code.Code,
            Description: code.Description,
            Confidence:  math.Round(code.Confidence*100) / 100, // Round to 2 decimals
        }
    }
    return output
}
```

**Test After Task 1:**
```bash
# Test classification endpoint
curl -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Restaurant",
    "website_url": "https://example-restaurant.com"
  }'

# Expected response:
# {
#   "codes": {
#     "mcc": [
#       {"code": "5812", "description": "Eating Places", "confidence": 0.92},
#       {"code": "5814", "description": "Fast Food Restaurants", "confidence": 0.85},
#       {"code": "5813", "description": "Drinking Places", "confidence": 0.67}
#     ],
#     "sic": [...],  // 3 codes
#     "naics": [...] // 3 codes
#   }
# }
```

---

### Task 2: Improve Confidence Calibration (Day 3-4)

**Current Issue:** Confidence scores are too low (50-70%) even when classifications are correct.

**File:** `internal/classification/confidence_calibrator.go`

```go
package classification

import (
    "math"
)

type ConfidenceCalibrator struct {
    // Historical accuracy data (could be loaded from database)
    historicalAccuracy map[string]float64
}

func NewConfidenceCalibrator() *ConfidenceCalibrator {
    return &ConfidenceCalibrator{
        historicalAccuracy: map[string]float64{
            "keyword":       0.88,  // Keyword strategy 88% accurate historically
            "entity":        0.82,
            "topic":         0.78,
            "co_occurrence": 0.75,
        },
    }
}

func (c *ConfidenceCalibrator) CalibrateConfidence(
    strategyResults map[string]float64,
    contentQuality float64,
    codeAgreement float64,
    methodUsed string,
) float64 {
    
    // Start with weighted average of strategy results
    baseConfidence := c.calculateWeightedAverage(strategyResults)
    
    // Apply calibration factors
    calibratedConfidence := baseConfidence
    
    // Factor 1: Content quality boost
    if contentQuality > 0.8 {
        calibratedConfidence *= 1.10 // +10% boost for high-quality content
    } else if contentQuality < 0.5 {
        calibratedConfidence *= 0.90 // -10% penalty for low-quality content
    }
    
    // Factor 2: Strategy agreement (low variance = high agreement)
    strategyVariance := c.calculateVariance(strategyResults)
    if strategyVariance < 0.05 {
        // Strategies strongly agree
        calibratedConfidence *= 1.15
    } else if strategyVariance < 0.10 {
        // Moderate agreement
        calibratedConfidence *= 1.08
    } else if strategyVariance > 0.25 {
        // High disagreement
        calibratedConfidence *= 0.85
    }
    
    // Factor 3: Code agreement (MCC/SIC/NAICS align via crosswalks)
    if codeAgreement > 0.85 {
        // Strong crosswalk validation
        calibratedConfidence *= 1.20
    } else if codeAgreement > 0.70 {
        // Moderate validation
        calibratedConfidence *= 1.10
    } else if codeAgreement < 0.40 {
        // Codes don't align - potential issue
        calibratedConfidence *= 0.85
    }
    
    // Factor 4: Method-specific calibration
    switch methodUsed {
    case "multi_strategy":
        // Multi-strategy tends to be reliable
        calibratedConfidence *= 1.05
    case "keyword_dominant":
        // Keyword-heavy classifications are very reliable
        if strategyResults["keyword"] > 0.85 {
            calibratedConfidence *= 1.15
        }
    case "ml_validated":
        // ML validation adds confidence
        calibratedConfidence *= 1.10
    }
    
    // Factor 5: Historical accuracy adjustment
    if historicalAccuracy, exists := c.historicalAccuracy[methodUsed]; exists {
        // Adjust based on historical performance
        calibratedConfidence *= historicalAccuracy
    }
    
    // Cap confidence at 0.95 (never claim 100% certainty)
    calibratedConfidence = math.Min(calibratedConfidence, 0.95)
    
    // Floor at 0.30 (anything lower suggests we shouldn't classify)
    calibratedConfidence = math.Max(calibratedConfidence, 0.30)
    
    return calibratedConfidence
}

func (c *ConfidenceCalibrator) calculateWeightedAverage(strategyResults map[string]float64) float64 {
    weights := map[string]float64{
        "keyword":       0.40,
        "entity":        0.25,
        "topic":         0.20,
        "co_occurrence": 0.15,
    }
    
    weightedSum := 0.0
    totalWeight := 0.0
    
    for strategy, confidence := range strategyResults {
        if weight, exists := weights[strategy]; exists {
            weightedSum += confidence * weight
            totalWeight += weight
        }
    }
    
    if totalWeight == 0 {
        return 0.5 // Default
    }
    
    return weightedSum / totalWeight
}

func (c *ConfidenceCalibrator) calculateVariance(strategyResults map[string]float64) float64 {
    if len(strategyResults) == 0 {
        return 0
    }
    
    // Calculate mean
    sum := 0.0
    for _, confidence := range strategyResults {
        sum += confidence
    }
    mean := sum / float64(len(strategyResults))
    
    // Calculate variance
    varianceSum := 0.0
    for _, confidence := range strategyResults {
        diff := confidence - mean
        varianceSum += diff * diff
    }
    
    return varianceSum / float64(len(strategyResults))
}

func (c *ConfidenceCalibrator) CalculateCodeAgreement(codes *ClassificationCodes) float64 {
    // Measure how well MCC/SIC/NAICS codes align
    // Higher score = codes validate each other through crosswalks
    
    agreementScore := 0.0
    checks := 0
    
    // Check if top MCC aligns with top SIC via crosswalks
    if len(codes.MCC) > 0 && len(codes.SIC) > 0 {
        // In real implementation, query crosswalks table
        // For now, use confidence as proxy
        mccConfidence := codes.MCC[0].Confidence
        sicConfidence := codes.SIC[0].Confidence
        
        // If both are high confidence, assume alignment
        if mccConfidence > 0.8 && sicConfidence > 0.8 {
            agreementScore += 1.0
        } else {
            agreementScore += (mccConfidence + sicConfidence) / 2.0
        }
        checks++
    }
    
    // Check if top MCC aligns with top NAICS
    if len(codes.MCC) > 0 && len(codes.NAICS) > 0 {
        mccConfidence := codes.MCC[0].Confidence
        naicsConfidence := codes.NAICS[0].Confidence
        
        if mccConfidence > 0.8 && naicsConfidence > 0.8 {
            agreementScore += 1.0
        } else {
            agreementScore += (mccConfidence + naicsConfidence) / 2.0
        }
        checks++
    }
    
    if checks == 0 {
        return 0.5 // Default
    }
    
    return agreementScore / float64(checks)
}
```

**Integrate into Service:**

**File:** `internal/classification/service.go`

```go
type IndustryDetectionService struct {
    // ... existing fields ...
    confidenceCalibrator *ConfidenceCalibrator // NEW
}

func NewIndustryDetectionService(...) *IndustryDetectionService {
    return &IndustryDetectionService{
        // ... existing initialization ...
        confidenceCalibrator: NewConfidenceCalibrator(),
    }
}

func (s *IndustryDetectionService) DetectIndustry(
    ctx context.Context,
    businessName, description, websiteURL string,
) (*IndustryDetectionResult, error) {
    
    // ... existing scraping and classification logic ...
    
    // Generate codes
    codes, err := s.codeGenerator.GenerateCodes(
        ctx,
        multiResult.Industry,
        multiResult.Keywords,
        multiResult.Confidence,
    )
    
    // Calculate code agreement
    codeAgreement := s.confidenceCalibrator.CalculateCodeAgreement(codes)
    
    // Calibrate confidence
    calibratedConfidence := s.confidenceCalibrator.CalibrateConfidence(
        multiResult.StrategyScores,
        content.QualityScore,
        codeAgreement,
        multiResult.Method,
    )
    
    // Update result with calibrated confidence
    multiResult.Confidence = calibratedConfidence
    
    // ... rest of logic ...
    
    return result, nil
}
```

**Test After Task 2:**
```bash
# Test on diverse businesses
# Expect confidence scores:
# - High-quality, clear cases: 85-95%
# - Medium-quality, ambiguous: 70-85%
# - Low-quality content: 60-75%

# Test restaurants (should be high confidence)
curl -X POST http://localhost:8080/v1/classify \
  -d '{"website_url": "https://mcdonalds.com"}'
# Expected confidence: 0.90-0.95

# Test ambiguous business (should be medium confidence)
curl -X POST http://localhost:8080/v1/classify \
  -d '{"website_url": "https://some-consulting-firm.com"}'
# Expected confidence: 0.75-0.85
```

---

### Task 3: Optimize Database Queries (Day 5)

**Goal:** Improve query performance for keyword and trigram lookups.

**File:** `supabase-migrations/040_optimize_classification_queries.sql`

```sql
-- Migration: Optimize classification queries for Phase 2

-- 1. Add composite indexes for common query patterns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_code_keywords_composite
ON code_keywords (code_type, keyword, weight DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_codes_type_description
ON classification_codes (code_type, description);

-- 2. Add index for crosswalk queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_crosswalks_lookup
ON industry_code_crosswalks (from_code_type, from_code, to_code_type);

-- 3. Optimize trigram queries with GIN index (if not already exists)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_codes_description_trgm
ON classification_codes USING gin (description gin_trgm_ops);

-- 4. Add covering index for keyword queries (includes all needed columns)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_code_keywords_covering
ON code_keywords (code_type, keyword) INCLUDE (code, weight);

-- 5. Refresh trigram indexes for better performance
ANALYZE code_keywords;
ANALYZE classification_codes;
ANALYZE industry_code_crosswalks;

-- 6. Create materialized view for frequently accessed code metadata
CREATE MATERIALIZED VIEW IF NOT EXISTS code_search_cache AS
SELECT 
    cc.code_type,
    cc.code,
    cc.description,
    cc.risk_level,
    array_agg(DISTINCT ck.keyword) as keywords,
    MAX(ck.weight) as max_keyword_weight
FROM classification_codes cc
LEFT JOIN code_keywords ck ON ck.code = cc.code AND ck.code_type = cc.code_type
GROUP BY cc.code_type, cc.code, cc.description, cc.risk_level;

CREATE UNIQUE INDEX ON code_search_cache (code_type, code);
CREATE INDEX ON code_search_cache USING gin (keywords);

-- Refresh materialized view (run periodically or on code updates)
REFRESH MATERIALIZED VIEW CONCURRENTLY code_search_cache;
```

**Run Migration:**
```bash
# Apply migration to Supabase
psql $SUPABASE_DB_URL -f supabase-migrations/040_optimize_classification_queries.sql

# Or via Supabase CLI
supabase db push
```

**Verify Performance Improvement:**
```sql
-- Test query performance before and after
EXPLAIN ANALYZE
SELECT DISTINCT
    ck.code,
    cc.description,
    MAX(ck.weight) as max_weight
FROM code_keywords ck
JOIN classification_codes cc ON cc.code = ck.code AND cc.code_type = ck.code_type
WHERE ck.code_type = 'MCC'
    AND ck.keyword = ANY(ARRAY['restaurant', 'food', 'dining'])
GROUP BY ck.code, cc.description
ORDER BY max_weight DESC
LIMIT 10;

-- Should show index usage and <50ms execution time
```

---

## Week 4: Performance & Explanations

### Task 4: Implement Fast Path (Day 1-2)

**Goal:** Handle obvious cases in <100ms without full multi-strategy processing.

**File:** `internal/classification/multi_strategy_classifier.go`

```go
// Add fast path logic
func (c *MultiStrategyClassifier) ClassifyWithMultiStrategy(
    ctx context.Context,
    businessName, description, websiteURL string,
) (*ClassificationResult, error) {
    
    startTime := time.Now()
    
    // Get scraped content
    content, err := c.scraper.Scrape(websiteURL)
    if err != nil {
        return nil, err
    }
    
    // Try fast path first
    if fastResult, isFastPath := c.tryFastPath(ctx, content); isFastPath {
        fastResult.ProcessingTimeMs = time.Since(startTime).Milliseconds()
        slog.Info("Fast path succeeded",
            "industry", fastResult.Industry,
            "confidence", fastResult.Confidence,
            "duration_ms", fastResult.ProcessingTimeMs)
        return fastResult, nil
    }
    
    // Fall back to full multi-strategy
    fullResult := c.classifyFullStrategy(ctx, content, businessName, description)
    fullResult.ProcessingTimeMs = time.Since(startTime).Milliseconds()
    
    return fullResult, nil
}

func (c *MultiStrategyClassifier) tryFastPath(
    ctx context.Context,
    content *ScrapedContent,
) (*ClassificationResult, bool) {
    
    // Extract obvious keywords from high-signal content
    obviousKeywords := c.extractObviousKeywords(content)
    
    if len(obviousKeywords) == 0 {
        return nil, false // No obvious keywords
    }
    
    // Check each obvious keyword for direct industry match
    for _, keyword := range obviousKeywords {
        // Query for high-confidence keyword matches
        matches := c.repo.GetIndustriesByKeyword(ctx, keyword, 0.90) // 90%+ weight
        
        if len(matches) > 0 {
            // Found high-confidence match via obvious keyword
            industry := matches[0]
            
            result := &ClassificationResult{
                Industry:   industry.Name,
                Confidence: 0.92, // High confidence for fast path
                Method:     "fast_path_keyword",
                Keywords:   []string{keyword},
                StrategyScores: map[string]float64{
                    "keyword": 0.95,
                },
            }
            
            return result, true
        }
    }
    
    // No fast path match found
    return nil, false
}

func (c *MultiStrategyClassifier) extractObviousKeywords(content *ScrapedContent) []string {
    // Look for high-signal keywords in title, meta description, and headings
    
    obviousKeywords := []string{}
    
    // Common obvious industry keywords
    obviouskeywordMap := map[string]bool{
        // Food & Beverage
        "restaurant": true, "cafe": true, "coffee": true, "bakery": true,
        "bar": true, "pub": true, "brewery": true, "winery": true,
        "pizzeria": true, "diner": true, "bistro": true,
        
        // Retail
        "shop": true, "store": true, "boutique": true, "market": true,
        "mall": true, "retail": true,
        
        // Professional Services
        "law firm": true, "attorney": true, "lawyer": true,
        "dentist": true, "dental": true, "orthodontist": true,
        "doctor": true, "physician": true, "medical": true,
        "accountant": true, "accounting": true, "cpa": true,
        
        // Home Services
        "plumber": true, "plumbing": true, "electrician": true,
        "contractor": true, "construction": true, "roofing": true,
        "hvac": true, "landscaping": true,
        
        // Automotive
        "auto repair": true, "mechanic": true, "car wash": true,
        "dealership": true, "automotive": true,
        
        // Personal Services
        "salon": true, "barber": true, "spa": true, "gym": true,
        "fitness": true, "yoga": true,
        
        // Hospitality
        "hotel": true, "motel": true, "inn": true, "resort": true,
        
        // Education
        "school": true, "university": true, "college": true,
        "tutoring": true, "academy": true,
    }
    
    // Check title (highest signal)
    titleWords := strings.Fields(strings.ToLower(content.Title))
    for _, word := range titleWords {
        if obviouskeywordMap[word] {
            obviousKeywords = append(obviousKeywords, word)
        }
    }
    
    // Check for phrase matches in title
    titleLower := strings.ToLower(content.Title)
    for keyword := range obviouskeywordMap {
        if strings.Contains(titleLower, keyword) {
            if !contains(obviousKeywords, keyword) {
                obviousKeywords = append(obviousKeywords, keyword)
            }
        }
    }
    
    // Check meta description if title didn't yield results
    if len(obviousKeywords) == 0 {
        descLower := strings.ToLower(content.MetaDesc)
        for keyword := range obviouskeywordMap {
            if strings.Contains(descLower, keyword) {
                obviousKeywords = append(obviousKeywords, keyword)
                break // Only need one from description
            }
        }
    }
    
    // Check first heading if still no matches
    if len(obviousKeywords) == 0 && len(content.Headings) > 0 {
        firstHeading := strings.ToLower(content.Headings[0])
        for keyword := range obviouskeywordMap {
            if strings.Contains(firstHeading, keyword) {
                obviousKeywords = append(obviousKeywords, keyword)
                break
            }
        }
    }
    
    return obviousKeywords
}

func (c *MultiStrategyClassifier) classifyFullStrategy(
    ctx context.Context,
    content *ScrapedContent,
    businessName, description string,
) *ClassificationResult {
    // Existing full multi-strategy logic
    // (your current ClassifyWithMultiStrategy implementation)
    // ... 
}

// Helper function
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

**Add Repository Method:**

**File:** `internal/classification/repository/supabase_repository.go`

```go
func (r *SupabaseKeywordRepository) GetIndustriesByKeyword(
    ctx context.Context,
    keyword string,
    minWeight float64,
) []struct {
    Name string
    Weight float64
} {
    
    query := `
        SELECT DISTINCT
            i.name,
            kw.weight
        FROM keyword_weights kw
        JOIN industries i ON i.id = kw.industry_id
        WHERE LOWER(kw.keyword) = LOWER($1)
            AND kw.weight >= $2
        ORDER BY kw.weight DESC
        LIMIT 5
    `
    
    var results []struct {
        Name   string  `db:"name"`
        Weight float64 `db:"weight"`
    }
    
    err := r.db.SelectContext(ctx, &results, query, keyword, minWeight)
    if err != nil {
        slog.Error("Failed to get industries by keyword", "error", err)
        return nil
    }
    
    return results
}
```

**Test Fast Path:**
```bash
# Test with obvious keywords
curl -X POST http://localhost:8080/v1/classify \
  -d '{"business_name": "Joe'\''s Pizza Restaurant", "website_url": "https://joespizza.com"}'

# Check logs for:
# INFO: Fast path succeeded industry=Restaurant confidence=0.92 duration_ms=45

# Expected:
# - Processing time: <100ms
# - Method: "fast_path_keyword"
# - High confidence: 0.90-0.95
```

---

### Task 5: Generate Structured Explanations (Day 3-4)

**Goal:** Create human-readable explanations that can be displayed in UI.

**File:** `internal/classification/explanation_generator.go`

```go
package classification

import (
    "fmt"
    "strings"
)

type ClassificationExplanation struct {
    PrimaryReason     string              `json:"primary_reason"`
    SupportingFactors []string            `json:"supporting_factors"`
    KeyTermsFound     []string            `json:"key_terms_found"`
    ConfidenceFactors map[string]float64  `json:"confidence_factors"`
    MethodUsed        string              `json:"method_used"`
    ProcessingPath    string              `json:"processing_path"` // "fast_path", "full_strategy", "ml_validated"
}

type ExplanationGenerator struct{}

func NewExplanationGenerator() *ExplanationGenerator {
    return &ExplanationGenerator{}
}

func (g *ExplanationGenerator) GenerateExplanation(
    result *ClassificationResult,
    codes *ClassificationCodes,
    content *ScrapedContent,
) *ClassificationExplanation {
    
    exp := &ClassificationExplanation{
        MethodUsed:        result.Method,
        KeyTermsFound:     result.Keywords,
        ConfidenceFactors: result.StrategyScores,
        ProcessingPath:    g.determineProcessingPath(result),
    }
    
    // Generate primary reason based on method
    exp.PrimaryReason = g.generatePrimaryReason(result, content)
    
    // Generate supporting factors
    exp.SupportingFactors = g.generateSupportingFactors(result, codes, content)
    
    return exp
}

func (g *ExplanationGenerator) determineProcessingPath(result *ClassificationResult) string {
    switch {
    case strings.Contains(result.Method, "fast_path"):
        return "fast_path"
    case strings.Contains(result.Method, "ml"):
        return "ml_validated"
    default:
        return "full_strategy"
    }
}

func (g *ExplanationGenerator) generatePrimaryReason(
    result *ClassificationResult,
    content *ScrapedContent,
) string {
    
    switch {
    case strings.Contains(result.Method, "fast_path"):
        // Fast path - obvious keyword match
        if len(result.Keywords) > 0 {
            return fmt.Sprintf(
                "Strong match based on clear industry indicator '%s' found in website title",
                result.Keywords[0],
            )
        }
        return "Clear industry classification from website title"
        
    case result.StrategyScores["keyword"] > 0.85:
        // Keyword-dominant classification
        keywordStr := strings.Join(result.Keywords[:min(3, len(result.Keywords))], ", ")
        return fmt.Sprintf(
            "Classified as '%s' based on strong keyword matches: %s",
            result.Industry,
            keywordStr,
        )
        
    case result.StrategyScores["entity"] > 0.80:
        // Entity-based classification
        return fmt.Sprintf(
            "Business entities and services indicate '%s' industry",
            result.Industry,
        )
        
    case result.StrategyScores["topic"] > 0.75:
        // Topic-based classification
        return fmt.Sprintf(
            "Website content and topic analysis indicates '%s' sector",
            result.Industry,
        )
        
    default:
        // Multi-strategy ensemble
        return fmt.Sprintf(
            "Classification as '%s' based on multiple indicators from website content",
            result.Industry,
        )
    }
}

func (g *ExplanationGenerator) generateSupportingFactors(
    result *ClassificationResult,
    codes *ClassificationCodes,
    content *ScrapedContent,
) []string {
    
    factors := []string{}
    
    // Factor: Website content quality
    if content.QualityScore > 0.8 {
        factors = append(factors,
            "High-quality website with comprehensive business information")
    } else if content.QualityScore > 0.6 {
        factors = append(factors,
            "Website provides good business context")
    }
    
    // Factor: Specific content elements found
    if content.AboutText != "" && len(content.AboutText) > 100 {
        factors = append(factors,
            "Detailed 'About' section provides clear business description")
    }
    
    if len(content.ProductList) > 0 {
        factors = append(factors,
            fmt.Sprintf("Product/service offerings identified (%d items)",
                min(len(content.ProductList), 10)))
    }
    
    if len(content.NavMenu) > 0 {
        factors = append(factors,
            "Website navigation structure indicates business operations")
    }
    
    // Factor: Multiple strategy agreement
    agreementCount := 0
    for _, score := range result.StrategyScores {
        if score > 0.70 {
            agreementCount++
        }
    }
    
    if agreementCount >= 3 {
        factors = append(factors,
            fmt.Sprintf("Multiple classification strategies agree (%d/4 with high confidence)",
                agreementCount))
    }
    
    // Factor: Code validation
    if len(codes.MCC) > 0 && len(codes.SIC) > 0 && len(codes.NAICS) > 0 {
        // Check if top codes have high confidence
        if codes.MCC[0].Confidence > 0.85 &&
            codes.SIC[0].Confidence > 0.85 &&
            codes.NAICS[0].Confidence > 0.85 {
            factors = append(factors,
                "Industry codes (MCC/SIC/NAICS) strongly align and validate classification")
        } else {
            factors = append(factors,
                "Industry codes identified across all classification systems")
        }
    }
    
    // Factor: Keywords found
    if len(result.Keywords) > 0 {
        if len(result.Keywords) <= 3 {
            factors = append(factors,
                fmt.Sprintf("Industry-specific terms found: %s",
                    strings.Join(result.Keywords, ", ")))
        } else {
            factors = append(factors,
                fmt.Sprintf("Multiple industry-specific terms found (%d total)",
                    len(result.Keywords)))
        }
    }
    
    // Limit to top 5 factors
    if len(factors) > 5 {
        factors = factors[:5]
    }
    
    return factors
}

// Helper function
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

**Integrate into Service:**

**File:** `internal/classification/service.go`

```go
type IndustryDetectionService struct {
    // ... existing fields ...
    explanationGenerator *ExplanationGenerator // NEW
}

func NewIndustryDetectionService(...) *IndustryDetectionService {
    return &IndustryDetectionService{
        // ... existing initialization ...
        explanationGenerator: NewExplanationGenerator(),
    }
}

func (s *IndustryDetectionService) DetectIndustry(
    ctx context.Context,
    businessName, description, websiteURL string,
) (*IndustryDetectionResult, error) {
    
    // ... existing classification logic ...
    
    // Generate explanation
    explanation := s.explanationGenerator.GenerateExplanation(
        multiResult,
        codes,
        content,
    )
    
    result := &IndustryDetectionResult{
        Classification: multiResult,
        Codes:          codes,
        Explanation:    explanation, // NEW
        ProcessingTimeMs: time.Since(startTime).Milliseconds(),
    }
    
    return result, nil
}
```

**Update API Response:**

**File:** `services/classification-service/internal/handlers/classification.go`

```go
type ClassificationResponse struct {
    RequestID      string                        `json:"request_id"`
    Classification ClassificationData            `json:"classification"`
    Codes          ClassificationCodesOutput     `json:"codes"`
    Explanation    ClassificationExplanation     `json:"explanation"` // NEW
    ProcessingTime int64                         `json:"processing_time_ms"`
}

// Update handler to include explanation
func (h *ClassificationHandler) HandleClassify(w http.ResponseWriter, r *http.Request) {
    // ... existing code ...
    
    response := ClassificationResponse{
        RequestID: req.RequestID,
        Classification: ClassificationData{
            PrimaryIndustry: result.Classification.PrimaryIndustry,
            Confidence:      result.Classification.Confidence,
            Method:          result.Classification.Method,
        },
        Codes: ClassificationCodesOutput{
            MCC:   convertCodesToOutput(result.Codes.MCC),
            SIC:   convertCodesToOutput(result.Codes.SIC),
            NAICS: convertCodesToOutput(result.Codes.NAICS),
        },
        Explanation: result.Explanation, // NEW
        ProcessingTime: result.ProcessingTimeMs,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

**Test Explanation Generation:**
```bash
curl -X POST http://localhost:8080/v1/classify \
  -d '{"website_url": "https://example-restaurant.com"}' | jq '.explanation'

# Expected output:
# {
#   "primary_reason": "Classified as 'Restaurant' based on strong keyword matches: restaurant, dining, menu",
#   "supporting_factors": [
#     "High-quality website with comprehensive business information",
#     "Detailed 'About' section provides clear business description",
#     "Product/service offerings identified (10 items)",
#     "Multiple classification strategies agree (3/4 with high confidence)",
#     "Industry codes (MCC/SIC/NAICS) strongly align and validate classification"
#   ],
#   "key_terms_found": ["restaurant", "dining", "menu", "reservations"],
#   "method_used": "multi_strategy",
#   "processing_path": "full_strategy"
# }
```

---

### Task 6: Fix "General Business" Fallback (Day 4-5)

**Goal:** Be more specific or admit uncertainty instead of defaulting to generic classifications.

**File:** `internal/classification/service.go`

```go
func (s *IndustryDetectionService) selectBestIndustry(
    candidates []IndustryCandidate,
    minConfidence float64,
) (*IndustryCandidate, error) {
    
    if len(candidates) == 0 {
        return nil, fmt.Errorf("no industry candidates")
    }
    
    // Sort by confidence
    sort.Slice(candidates, func(i, j int) bool {
        return candidates[i].Confidence > candidates[j].Confidence
    })
    
    topCandidate := candidates[0]
    
    // Define generic/vague industries to avoid
    genericIndustries := map[string]bool{
        "General Business":      true,
        "Other Services":        true,
        "Miscellaneous":         true,
        "Business Services":     true,
        "Professional Services": true,
        "General Merchandise":   true,
        "Other":                 true,
    }
    
    // If top candidate is generic and confidence is low, try next
    if genericIndustries[topCandidate.Name] && topCandidate.Confidence < 0.75 {
        slog.Warn("Top candidate is generic with low confidence, checking alternatives",
            "industry", topCandidate.Name,
            "confidence", topCandidate.Confidence)
        
        // Look for more specific alternative
        for i := 1; i < len(candidates) && i < 5; i++ {
            candidate := candidates[i]
            
            // If we find a specific industry within 0.15 of generic confidence, prefer it
            if !genericIndustries[candidate.Name] &&
                (topCandidate.Confidence - candidate.Confidence) < 0.15 {
                
                slog.Info("Preferring specific industry over generic",
                    "chosen", candidate.Name,
                    "confidence", candidate.Confidence,
                    "rejected_generic", topCandidate.Name)
                
                return &candidate, nil
            }
        }
    }
    
    // Require minimum confidence threshold
    if topCandidate.Confidence < minConfidence {
        return nil, fmt.Errorf("insufficient confidence: %.2f < %.2f",
            topCandidate.Confidence, minConfidence)
    }
    
    // For generic industries, require higher confidence
    if genericIndustries[topCandidate.Name] {
        if topCandidate.Confidence < 0.70 {
            return nil, fmt.Errorf(
                "generic industry '%s' requires ≥0.70 confidence, got %.2f",
                topCandidate.Name,
                topCandidate.Confidence)
        }
    }
    
    return &topCandidate, nil
}
```

**Add Specific Industry Boosting:**

```go
func (s *IndustryDetectionService) boostSpecificIndustries(
    candidates []IndustryCandidate,
) []IndustryCandidate {
    
    // Boost confidence of specific industries slightly
    specificBoost := 0.05
    
    genericIndustries := map[string]bool{
        "General Business":      true,
        "Other Services":        true,
        "Miscellaneous":         true,
        "Business Services":     true,
        "Professional Services": true,
    }
    
    for i := range candidates {
        if !genericIndustries[candidates[i].Name] {
            candidates[i].Confidence = math.Min(
                candidates[i].Confidence + specificBoost,
                0.95,
            )
        }
    }
    
    return candidates
}
```

**Test Generic Fallback:**
```bash
# Test with ambiguous business
curl -X POST http://localhost:8080/v1/classify \
  -d '{"business_name": "Acme Consulting", "website_url": "https://acme-consulting.com"}'

# Should return specific industry like:
# - "Management Consulting" (not "Business Services")
# - "Technology Consulting" (not "General Business")
# - Or lower confidence if truly ambiguous

# Should NOT return "General Business" unless confidence > 0.70
```

---

## Phase 2 Testing & Validation (Day 5)

### Full Test Suite

**Run comprehensive tests on your test set:**

```bash
# Test script (pseudo-code)
for test_case in test_set:
    result = classify(test_case.url)
    
    # Validate:
    # 1. Returns 3 codes per type
    assert len(result.codes.mcc) == 3
    assert len(result.codes.sic) == 3
    assert len(result.codes.naics) == 3
    
    # 2. Confidence is calibrated (not too low)
    assert result.confidence >= 0.60
    
    # 3. Has explanation
    assert result.explanation.primary_reason != ""
    assert len(result.explanation.supporting_factors) > 0
    
    # 4. Not generic (unless truly generic business)
    if not test_case.is_generic:
        assert result.primary_industry not in ["General Business", "Other Services"]
    
    # 5. Fast path used when appropriate
    if test_case.has_obvious_keyword:
        assert result.explanation.processing_path == "fast_path"
        assert result.processing_time_ms < 100
```

### Performance Benchmarks

**Expected performance distribution:**

```
Fast Path (60-70% of requests):
- Processing time: <100ms
- Confidence: 0.90-0.95
- Example: Restaurants, retail stores, obvious cases

Full Strategy (30-40% of requests):
- Processing time: 200-500ms
- Confidence: 0.70-0.90
- Example: Consulting firms, mixed businesses

Performance Summary:
- p50: <150ms
- p90: <400ms
- p95: <500ms
```

**Measure with your test set:**
```bash
# Calculate percentiles
cat test_results.json | jq '[.[] | .processing_time_ms] | sort | .[length/2], .[length*0.9|floor], .[length*0.95|floor]'
```

---

## Phase 2 Success Criteria

Before moving to Phase 3, verify:

### Functionality
- [ ] ✅ Returns top 3 codes per type (MCC/SIC/NAICS)
- [ ] ✅ Each code has confidence score
- [ ] ✅ Crosswalk validation enriches codes
- [ ] ✅ Codes align (high code agreement score)

### Confidence
- [ ] ✅ Confidence scores range 60-95% (not all 50%)
- [ ] ✅ High-quality cases show 85-95% confidence
- [ ] ✅ Ambiguous cases show 65-80% confidence
- [ ] ✅ Confidence correlates with actual accuracy

### Performance
- [ ] ✅ Fast path handles 60-70% of requests
- [ ] ✅ Fast path latency <100ms (p95)
- [ ] ✅ Full strategy latency <500ms (p95)
- [ ] ✅ Database queries optimized (<50ms each)

### Explanations
- [ ] ✅ Every classification has primary reason
- [ ] ✅ Supporting factors listed (3-5 items)
- [ ] ✅ Key terms identified
- [ ] ✅ Method and processing path indicated

### Quality
- [ ] ✅ "General Business" only used appropriately
- [ ] ✅ Specific industries preferred over generic
- [ ] ✅ Accuracy improved to 80-85% on test set

---

## Expected Improvement Summary

**Before Phase 2:**
- Codes: Only 1 per type
- Confidence: ~50-70%
- Speed: 200-500ms for all
- Explanations: Not structured
- Generic classifications: Common
- Accuracy: 50-60%

**After Phase 2:**
- Codes: ✅ Top 3 per type with confidence
- Confidence: ✅ 70-95% (calibrated)
- Speed: ✅ <100ms for 70% of requests
- Explanations: ✅ Structured and informative
- Generic classifications: ✅ Minimized
- Accuracy: ✅ 80-85%

---

## Next Steps: Phase 3

Once Phase 2 is complete and validated:
- **Phase 3 (Weeks 5-6):** Add embedding-based similarity search
- **Expected accuracy:** 85-90%
- **What it adds:** Better handling of edge cases and novel business types

**Key Phase 3 Components:**
- Enable pgvector in Supabase
- Pre-compute code embeddings
- Deploy embedding service (Python)
- Add Layer 2 routing

**Phase 3 Guide will be provided once Phase 2 is complete.**

---

## Troubleshooting

**Issue: Still only returning 1 code per type**
- Check `selectTopCodes` function limits
- Verify API response structure updated
- Check frontend is handling array instead of single value

**Issue: Confidence still low**
- Review calibration factors
- Check strategy scores in logs
- Verify code agreement calculation
- May need to adjust calibration multipliers

**Issue: Fast path not triggering**
- Check `extractObviousKeywords` logic
- Verify keyword database has high-weight entries
- Review logs for "Fast path succeeded" messages
- May need to expand obvious keyword list

**Issue: Generic industries still appearing**
- Check `selectBestIndustry` logic
- Verify generic industry list
- Review confidence thresholds
- May need to boost specific industries more

**Issue: Database queries slow**
- Run `EXPLAIN ANALYZE` on slow queries
- Verify indexes created properly
- Check `pg_stat_user_indexes` for index usage
- May need to refresh statistics: `ANALYZE table_name`

---

## Questions During Phase 2?

If you encounter issues:
1. Check logs for detailed error messages
2. Test individual components (fast path, confidence, explanation)
3. Validate database query performance
4. Review test results for patterns

You've got the scraping foundation solid from Phase 1. Now let's make that classifier shine! 🚀

**Ready to start? Begin with Task 1: Return Top 3 Codes!**
