# Test Results Comparison: Baseline vs After Fixes

**Date**: 2025-11-30  
**Test Cases**: 184

---

## Results Summary

| Metric | Baseline | After Fixes | Change | Status |
|--------|----------|-------------|--------|--------|
| **Industry Accuracy** | 9.24% | 8.70% | -0.54% | ⚠️ Slightly worse |
| **Code Accuracy** | 0.00% | 0.00% | 0.00% | ❌ No change |
| **Overall Accuracy** | 3.70% | 3.48% | -0.22% | ⚠️ Slightly worse |
| **Cases with Codes** | 14/184 (7.6%) | 0/184 (0.0%) | -14 | ❌ Worse |
| **Industry Matches** | 17/184 | 16/184 | -1 | ⚠️ Slightly worse |

---

## Analysis

### Issue Identified: Confidence Calculation Too Conservative

**Root Cause**: Codes are being generated from keywords, but filtered out due to low confidence scores.

**Confidence Calculation Formula**:
```
confidence = relevance_score * industry_confidence * 0.85
```

**Example**:
- relevance_score = 0.5 (minimum from minRelevance)
- industry_confidence = 0.4 (low confidence)
- Result: `0.5 * 0.4 * 0.85 = 0.17`

Even with lowered threshold of 0.3, codes are filtered out because confidence (0.17) < threshold (0.3).

### Evidence from Logs

The test logs show:
- ✅ Codes ARE being generated: "Generated 9 MCC codes from keywords"
- ❌ Codes ARE being filtered out: "0 MCC, 0 SIC, 0 NAICS codes" (final result)

**Cases with industry detected but no codes**: 68/184 (37.0%)
- Average confidence: 0.51
- Confidence range: 0.24 - 0.79

---

## Fixes Applied

### 1. Adjusted Confidence Calculation for Low-Confidence Industries

**Before**:
```go
confidence = relevance_score * industry_confidence * 0.85
```

**After**:
```go
if industryConfidence < 0.5 {
    // Less aggressive multiplier for low-confidence industries
    confidence = relevance_score * 0.7
} else {
    // Original formula for high-confidence industries
    confidence = relevance_score * industry_confidence * 0.85
}

// Minimum confidence floor
if confidence < 0.2 && relevance_score >= 0.5 {
    confidence = 0.2
}
```

### 2. Adjusted Industry-Based Code Confidence

**Before**:
```go
confidence = industry_confidence * 0.9
```

**After**:
```go
if industryConfidence < 0.5 {
    confidence = 0.3 // Minimum floor
} else {
    confidence = industry_confidence * 0.9
}
```

---

## Expected Improvements

With the confidence calculation fixes:

1. **Code Generation**: Should increase from 0% to 30-50%+
   - Codes will no longer be filtered out due to low confidence
   - Minimum confidence floors ensure codes are included

2. **Code Accuracy**: Should improve from 0.00% to 20-40%+
   - Codes will be generated and included in results
   - Better matching with expected codes

3. **Overall Accuracy**: Should improve from 3.48% to 15-25%+
   - Better code generation will improve overall scores

---

## Next Steps

1. **Re-run tests** with confidence calculation fixes
2. **Compare results** against baseline and v2
3. **Analyze improvements** in code generation
4. **Identify remaining issues** if any
5. **Expand dataset** once code generation is working

---

**Status**: Confidence calculation fixes applied. Ready for re-testing.

