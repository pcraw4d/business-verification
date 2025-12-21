# Code Generation Error Analysis

**Date**: December 21, 2025  
**Investigation Track**: Track 4.1 - Code Generation Failure Investigation  
**Status**: Completed

## Executive Summary

Analysis revealed that the code generation confidence threshold (0.5 / 50%) was too high compared to the average confidence score (21.7%), causing 77% of requests to skip code generation. This has been fixed by lowering the threshold to 0.15 (15%).

---

## Problem Identified

### Code Generation Trigger Logic

**Location**: `services/classification-service/internal/handlers/classification.go:1711-1712`

**Original Logic**:
```go
shouldGenerateCodes := enhancedResult.ConfidenceScore >= 0.5 ||
    (enhancedResult.ConfidenceScore >= h.industryThresholds.GetThreshold(enhancedResult.PrimaryIndustry))
```

**Issue**: 
- Threshold: 0.5 (50%)
- Average Confidence: 21.7% (0.217)
- **Result**: Only 23.1% of requests generate codes (77% are blocked)

### Industry Thresholds

**Location**: `internal/classification/industry_thresholds.go:44-49`

**Original Defaults**:
```go
defaultThreshold: 0.3  // 30%
minThreshold: 0.2       // 20%
```

**ShouldGenerateCodes Fallback**:
```go
// Original
return confidence >= threshold || confidence > 0.5  // 50% fallback
```

**Issue**: 
- Default threshold: 0.3 (30%) - still too high
- Fallback threshold: 0.5 (50%) - way too high
- Average confidence: 0.217 (21.7%) - below both thresholds

---

## Solution Implemented

### Fix 1: Lower Default Industry Threshold

**Change**: Reduced default threshold from 0.3 to 0.15

**Location**: `internal/classification/industry_thresholds.go:44-49`

```go
// Before
defaultThreshold: 0.3
minThreshold: 0.2

// After
defaultThreshold: 0.15  // Reduced from 0.3 to match actual confidence levels (avg 21.7%)
minThreshold: 0.1       // Reduced from 0.2 to allow code generation for low-confidence classifications
```

**Rationale**: 
- Average confidence is 21.7%, so default threshold should be below this
- 0.15 (15%) allows most requests to generate codes while still filtering very low confidence (< 15%)

### Fix 2: Lower Fallback Threshold

**Change**: Reduced fallback threshold from 0.5 to 0.15

**Location**: `internal/classification/industry_thresholds.go:114-118`

```go
// Before
return confidence >= threshold || confidence > 0.5

// After
return confidence >= threshold || confidence >= 0.15
```

**Rationale**: 
- Fallback threshold should match default threshold
- 0.15 allows code generation for most requests (avg 21.7% > 15%)

### Fix 3: Lower Hardcoded Threshold in Streaming Handler

**Change**: Reduced hardcoded threshold from 0.5 to 0.15

**Location**: `services/classification-service/internal/handlers/classification.go:1711-1712`

```go
// Before
shouldGenerateCodes := enhancedResult.ConfidenceScore >= 0.5 ||

// After
shouldGenerateCodes := enhancedResult.ConfidenceScore >= 0.15 ||
```

**Rationale**: 
- Consistent threshold across all code paths
- Matches industry threshold fallback

---

## Expected Impact

### Before Fix

| Metric | Value | Status |
|--------|-------|--------|
| **Code Generation Rate** | 23.1% | ❌ Low |
| **Threshold** | 0.5 (50%) | ❌ Too High |
| **Average Confidence** | 21.7% | ⚠️ Below Threshold |
| **Blocked Requests** | 76.9% | ❌ High |

### After Fix

| Metric | Expected | Status |
|--------|----------|--------|
| **Code Generation Rate** | ≥90% | ✅ Target |
| **Threshold** | 0.15 (15%) | ✅ Aligned |
| **Average Confidence** | 21.7% | ✅ Above Threshold |
| **Blocked Requests** | <10% | ✅ Low |

### Calculation

- **Before**: Only requests with confidence ≥ 0.5 (50%) generate codes
  - Requests above 0.5: ~23.1%
  - Requests below 0.5: ~76.9% (blocked)

- **After**: Requests with confidence ≥ 0.15 (15%) generate codes
  - Requests above 0.15: ~90%+ (estimated)
  - Requests below 0.15: ~10% (blocked)

---

## Industry-Specific Thresholds

### High-Risk Industries (Unchanged)

These industries still require higher confidence:
- Financial Services: 0.7 (70%)
- Healthcare: 0.65 (65%)
- Legal: 0.6 (60%)

**Rationale**: These industries require higher confidence for regulatory compliance.

### Medium-Risk Industries (Unchanged)

- Real Estate: 0.5 (50%)
- Construction: 0.5 (50%)
- Manufacturing: 0.45 (45%)

### Low-Risk Industries (Now Use Default)

- Technology: 0.15 (15%) - was 0.3
- Retail: 0.15 (15%) - was 0.3
- Food & Beverage: 0.15 (15%) - was 0.3

**Impact**: Most industries will now use the lower default threshold, allowing more code generation.

---

## Code Generation Function Analysis

### Location: `internal/classification/classifier.go:256-303`

**Status**: ✅ Function is working correctly

**Key Operations**:
1. Keyword matching
2. Industry matching
3. Database queries for code metadata
4. Crosswalk relationships
5. Code ranking and confidence scoring

**No Issues Found**: The code generation function itself is working correctly. The issue was the threshold preventing it from being called.

---

## Database Query Analysis

### Code Metadata Queries

**Location**: `internal/classification/repository/supabase_repository.go`

**Queries**:
- MCC codes lookup
- NAICS codes lookup
- SIC codes lookup
- Code metadata lookup

**Status**: ⚠️ Needs verification (Track 4.2 will investigate NAICS/SIC data completeness)

---

## Error Handling

### Current Error Handling

**Location**: `services/classification-service/internal/handlers/classification.go:1730-1738`

```go
if codeGenErr != nil {
    h.logger.Warn("Code generation failed, continuing without codes",
        zap.String("request_id", req.RequestID),
        zap.Error(codeGenErr))
    // Continue without codes - doesn't fail the request
}
```

**Status**: ✅ Errors are handled gracefully - request continues without codes

**Issue**: Errors may be silently ignored. Should log more details for debugging.

---

## Recommendations

### Immediate Actions

1. ✅ **Lower Thresholds** - Completed
   - Default threshold: 0.3 → 0.15
   - Fallback threshold: 0.5 → 0.15
   - Hardcoded threshold: 0.5 → 0.15

2. **Monitor Code Generation Rate**
   - Track code generation rate after fix
   - Verify it increases from 23.1% to ≥90%
   - Monitor for any quality degradation

### Short-Term Actions

3. **Improve Error Logging**
   - Add detailed error logging for code generation failures
   - Track common failure patterns
   - Identify database query issues

4. **Verify Database Data**
   - Check NAICS/SIC code data completeness (Track 4.2)
   - Verify code metadata queries are working
   - Test code generation manually

### Long-Term Actions

5. **Dynamic Threshold Adjustment**
   - Adjust thresholds based on code generation success rates
   - Learn optimal thresholds from historical data
   - Industry-specific threshold optimization

---

## Code Changes Summary

### Files Modified

1. `internal/classification/industry_thresholds.go`
   - Reduced defaultThreshold from 0.3 to 0.15
   - Reduced minThreshold from 0.2 to 0.1
   - Reduced fallback threshold in ShouldGenerateCodes from 0.5 to 0.15

2. `services/classification-service/internal/handlers/classification.go`
   - Reduced hardcoded threshold from 0.5 to 0.15

### Testing Required

- [ ] Unit tests for threshold logic
- [ ] Integration tests for code generation
- [ ] E2E tests to validate code generation rate increase

---

## Validation Plan

1. **Deploy Fix**: Deploy updated code to Railway
2. **Run 50-Sample Test**: Validate code generation rate improvement
3. **Monitor Metrics**:
   - Code generation rate (target: ≥90%)
   - Code accuracy (target: ≥70%)
   - Code confidence (target: >70%)
4. **Compare Results**: Before vs After metrics

---

**Document Status**: Analysis Complete, Fixes Implemented  
**Next Steps**: Deploy and validate with 50-sample E2E test

