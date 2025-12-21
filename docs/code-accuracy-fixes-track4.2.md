# Code Accuracy Fixes - Track 4.2

**Date**: December 21, 2025  
**Investigation Track**: Track 4.2 - Fix NAICS/SIC Code Generation & Code Accuracy Regression  
**Status**: In Progress

## Executive Summary

Implemented fixes to address code accuracy regression and NAICS/SIC code generation issues:

1. ✅ **Lowered confidence threshold** in `mergeCodeResults` (0.6 → 0.4, 0.3 → 0.2)
2. ✅ **Boosted industry-based codes** to prioritize them over keyword-based codes
3. ✅ **Improved code ranking** to prioritize industry_match over keyword_match
4. ✅ **Created database function** `get_codes_by_trigram_similarity` for NAICS/SIC

---

## Problem Analysis

### Code Accuracy Regression

**Before**: 27.5% overall accuracy, 10.0% MCC Top 1  
**After**: 10.8% overall accuracy, 0.0% MCC Top 1

**Root Cause**: 
- Lower threshold (0.5 → 0.15) increased code generation rate (23.1% → 48.0%)
- But generated codes don't match expected codes
- Industry-based codes not being prioritized over keyword-based codes
- Code ranking algorithm doesn't prioritize industry_match

### NAICS/SIC Code Generation

**Status**: 0% accuracy for both NAICS and SIC

**Root Cause**:
- Database function `get_codes_by_trigram_similarity` may not exist
- Function returns 404 when called
- Codes may not be generated at all

---

## Fixes Implemented

### Fix 1: Lower Confidence Threshold in mergeCodeResults

**File**: `internal/classification/classifier.go:414-421`

**Change**:
- Default threshold: 0.6 → 0.4
- Low confidence threshold: 0.3 → 0.2

**Rationale**: 
- Lower threshold allows more codes to pass filtering
- But we also boost industry-based codes to maintain accuracy

**Code**:
```go
confidenceThreshold := 0.4  // Lowered from 0.6 to 0.4
if industryConfidence < 0.5 {
    confidenceThreshold = 0.2  // Lowered from 0.3 to 0.2
}
```

---

### Fix 2: Boost Industry-Based Codes

**File**: `internal/classification/classifier.go:437-459`

**Changes**:
1. Increased confidence for low-confidence industries: 0.3 → 0.4
2. Increased confidence for high-confidence industries: 0.9 → 0.95
3. Added boost for primary codes (confidence > 0.8): +10%

**Rationale**:
- Industry-based codes are more reliable than keyword-based codes
- Boosting their confidence ensures they rank higher
- Primary codes should be prioritized

**Code**:
```go
if industryConfidence < 0.5 {
    confidence = 0.4 // Increased from 0.3
} else {
    confidence = industryConfidence * 0.95 // Increased from 0.9
}

// Boost primary codes
if code.Confidence > 0.8 {
    confidence = math.Min(confidence * 1.1, 0.98) // +10% boost
}
```

---

### Fix 3: Improve Code Ranking in mergeCodeResults

**File**: `internal/classification/classifier.go:515-540`

**Change**: Added secondary sort to prioritize industry-based codes

**Rationale**:
- Industry-based codes are more accurate than keyword-based codes
- When confidence is equal, industry-based codes should rank higher

**Code**:
```go
// Secondary sort: prioritize industry-based codes over keyword-based
iHasIndustry := false
jHasIndustry := false
for _, source := range results[i].Sources {
    if source == "industry" {
        iHasIndustry = true
        break
    }
}
// ... (similar for j)
if iHasIndustry != jHasIndustry {
    return iHasIndustry // Industry-based codes come first
}
```

---

### Fix 4: Improve Code Ranking in selectTopCodes

**File**: `internal/classification/classifier.go:920-937`

**Change**: Added secondary sort to prioritize industry_match over keyword_match

**Rationale**:
- Industry_match codes are more reliable for accuracy
- When confidence is equal, industry_match should rank higher

**Code**:
```go
// Secondary sort: prioritize industry_match over keyword_match
if candidates[i].Source == "industry_match" && candidates[j].Source != "industry_match" {
    return true
}
if candidates[i].Source != "industry_match" && candidates[j].Source == "industry_match" {
    return false
}
```

---

### Fix 5: Create Database Function for NAICS/SIC

**File**: `supabase-migrations/035_create_get_codes_by_trigram_similarity.sql`

**Change**: Created `get_codes_by_trigram_similarity` function

**Rationale**:
- Function is called by `GetCodesByTrigramSimilarity` but may not exist
- Function is needed for NAICS/SIC code generation via trigram similarity

**Function**:
```sql
CREATE OR REPLACE FUNCTION get_codes_by_trigram_similarity(
    p_code_type text,
    p_industry_name text,
    p_threshold float DEFAULT 0.3,
    p_limit int DEFAULT 3
)
RETURNS TABLE (
    code text,
    description text,
    similarity float
)
```

**Features**:
- Uses trigram similarity to match industry name to code descriptions
- Filters by code_type (MCC, SIC, NAICS)
- Returns top N codes by similarity
- Includes trigram index for performance

---

## Expected Impact

### Code Accuracy

| Metric | Before | Expected After | Change |
|--------|--------|----------------|--------|
| **Overall Code Accuracy** | 10.8% | 25-35% | **+15-25%** |
| **MCC Top 1 Accuracy** | 0.0% | 10-20% | **+10-20%** |
| **MCC Top 3 Accuracy** | 12.5% | 25-35% | **+12.5-22.5%** |

**Rationale**:
- Industry-based codes are now prioritized, which should improve accuracy
- Lower threshold allows more codes, but ranking ensures best codes are first
- Primary codes are boosted, improving top 1 accuracy

### NAICS/SIC Code Generation

| Metric | Before | Expected After | Change |
|--------|--------|----------------|--------|
| **NAICS Accuracy** | 0% | 20-40% | **+20-40%** |
| **SIC Accuracy** | 0% | 20-40% | **+20-40%** |

**Rationale**:
- Database function now exists, enabling trigram similarity matching
- Codes should be generated for NAICS/SIC types
- Accuracy should improve as codes are generated correctly

---

## Testing Required

1. **Deploy Database Migration**
   - Run `035_create_get_codes_by_trigram_similarity.sql` in Supabase
   - Verify function exists and is callable

2. **Deploy Code Changes**
   - Deploy updated `classifier.go` with ranking improvements
   - Deploy updated `website_scraper.go` with validation fixes

3. **Run 50-Sample Validation Test**
   - Measure code accuracy improvements
   - Verify NAICS/SIC codes are generated
   - Check if top 1 accuracy improves

---

## Next Steps

1. [ ] Deploy database migration (035_create_get_codes_by_trigram_similarity.sql)
2. [ ] Deploy code changes to Railway
3. [ ] Run 50-sample validation test
4. [ ] Analyze results and iterate if needed

---

**Document Status**: Fixes Implemented  
**Next Action**: Deploy fixes and test

