# Code Accuracy Regression Analysis - Track 4.2

**Date**: December 21, 2025  
**Investigation Track**: Track 4.2 - Fix NAICS/SIC Code Generation & Code Accuracy Regression  
**Status**: In Progress

## Executive Summary

The 50-sample validation test shows **code accuracy regression**:
- Overall code accuracy: 27.5% → 10.8% (regressed by 16.7%)
- MCC Top 1 accuracy: 10.0% → 0.0% (regressed by 10.0%)
- MCC Top 3 accuracy: 31.2% → 12.5% (regressed by 18.7%)
- NAICS accuracy: 0% → 0% (no change - still critical)
- SIC accuracy: 0% → 0% (no change - still critical)

**Root Cause Hypothesis**: Lowering the code generation threshold (0.5 → 0.15) increased code generation rate (23.1% → 48.0%) but may be generating lower-quality codes that don't match expected values.

---

## Problem Analysis

### Code Generation Metrics

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| **Code Generation Rate** | 23.1% | 48.0% | **+24.9%** ✅ |
| **Code Confidence Avg** | 87.3% | 88.4% | **+1.1%** ✅ |
| **Overall Code Accuracy** | 27.5% | 10.8% | **-16.7%** ❌ |
| **MCC Top 1 Accuracy** | 10.0% | 0.0% | **-10.0%** ❌ |
| **MCC Top 3 Accuracy** | 31.2% | 12.5% | **-18.7%** ❌ |
| **NAICS Accuracy** | 0% | 0% | **0%** ❌ |
| **SIC Accuracy** | 0% | 0% | **0%** ❌ |

### Key Findings

1. **More Codes Generated, But Less Accurate**
   - Code generation rate doubled (23.1% → 48.0%)
   - But accuracy decreased (27.5% → 10.8%)
   - Suggests lower threshold is generating codes that don't match expected values

2. **MCC Top 1 Accuracy Dropped to 0%**
   - Previously 10% of top codes were correct
   - Now 0% of top codes are correct
   - Codes are being generated but ranked incorrectly

3. **MCC Top 3 Accuracy Still Low**
   - 12.5% of codes are in top 3 (down from 31.2%)
   - Suggests codes are being generated but not ranked correctly
   - Or codes being generated don't match expected values

4. **NAICS/SIC Still 0% Accuracy**
   - No improvement despite threshold fix
   - Suggests separate issue (database function, data, or logic)

---

## Root Cause Analysis

### Issue 1: Code Ranking Algorithm

**Location**: `internal/classification/classifier.go:339-375`

**Problem**: Codes may be generated but ranked incorrectly:
- Top code is never correct (0% top 1 accuracy)
- But codes are sometimes in top 3 (12.5% top 3 accuracy)
- Suggests ranking algorithm is not prioritizing correct codes

**Investigation Needed**:
1. Review code ranking logic
2. Check confidence score calculation
3. Verify code selection criteria

---

### Issue 2: Code Matching Algorithm

**Location**: `internal/classification/classifier.go:256-320`

**Problem**: Generated codes may not match expected codes:
- Codes are generated with high confidence (88.4%)
- But accuracy is low (10.8%)
- Suggests codes don't match expected values

**Investigation Needed**:
1. Compare generated codes vs expected codes
2. Check code matching logic
3. Verify code generation criteria

---

### Issue 3: NAICS/SIC Code Generation

**Location**: `internal/classification/repository/supabase_repository.go:1995-2120`

**Problem**: NAICS/SIC codes are not being generated correctly:
- 0% accuracy for both NAICS and SIC
- Codes may not be generated at all
- Or codes are generated but incorrect

**Possible Causes**:
1. Database function `get_codes_by_trigram_similarity` may not exist
2. NAICS/SIC code data may be missing from database
3. Code generation logic may not be working for NAICS/SIC

**Investigation Needed**:
1. Verify database function exists
2. Check database data completeness
3. Test code generation manually

---

## Investigation Steps

### Step 1: Review Code Generation Threshold Impact

**Action**: Analyze if lower threshold (0.15) is generating incorrect codes

**Hypothesis**: Lower threshold may be generating codes for low-confidence classifications that don't match expected values

**Test**: Compare code accuracy for high-confidence vs low-confidence classifications

---

### Step 2: Review Code Ranking Algorithm

**Action**: Review code ranking and selection logic

**Questions**:
1. How are codes ranked?
2. Is confidence score used for ranking?
3. Are industry-specific codes prioritized?

---

### Step 3: Verify NAICS/SIC Database Function

**Action**: Check if `get_codes_by_trigram_similarity` function exists in Supabase

**Location**: `internal/classification/repository/supabase_repository.go:2033`

**Code**:
```go
url := fmt.Sprintf("%s/rest/v1/rpc/get_codes_by_trigram_similarity", r.client.GetURL())
```

**Test**: Verify function exists and is callable

---

### Step 4: Check NAICS/SIC Code Data

**Action**: Verify NAICS/SIC code data exists in database

**Tables to Check**:
- `classification_codes` (where code_type = 'NAICS' or 'SIC')
- `code_keywords` (keywords for NAICS/SIC codes)

**Test**: Query database for code counts

---

## Recommended Fixes

### Fix 1: Review Code Generation Threshold

**File**: `internal/classification/industry_thresholds.go`

**Current**: Threshold lowered to 0.15

**Consideration**: May need to balance between:
- Generating more codes (higher rate)
- Generating accurate codes (higher accuracy)

**Option**: Use adaptive threshold based on confidence score

---

### Fix 2: Improve Code Ranking Algorithm

**File**: `internal/classification/classifier.go`

**Action**: Review and improve code ranking logic:
- Prioritize codes with higher confidence
- Consider industry-specific codes
- Improve code selection criteria

---

### Fix 3: Fix NAICS/SIC Code Generation

**File**: `internal/classification/repository/supabase_repository.go`

**Action**: 
1. Verify database function exists
2. Check database data completeness
3. Fix code generation logic if needed

---

## Next Steps

1. [ ] Review code ranking algorithm (Fix 2)
2. [ ] Verify NAICS/SIC database function (Fix 3)
3. [ ] Check NAICS/SIC code data (Fix 3)
4. [ ] Test code generation manually
5. [ ] Implement fixes
6. [ ] Re-test with 50-sample validation

---

**Document Status**: Analysis Complete  
**Next Action**: Review code ranking algorithm and verify NAICS/SIC database function

