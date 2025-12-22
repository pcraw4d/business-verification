# Classification Algorithm Investigation - Track 3.1

## Executive Summary

Investigation of the classification algorithm reveals **critical issues** contributing to the **10.7% classification accuracy** (target: ≥80%). The primary issues are:

1. **Python ML Service Unavailable** - Circuit breaker is OPEN, forcing fallback to keyword-based classification
2. **Low Confidence Scores** - Average confidence of 24.65% (target: >70%)
3. **Defaulting to "General Business"** - Most classifications defaulting to generic industry
4. **Industry Detection Logic Issues** - Many industries showing 0% accuracy

## Current Classification Flow

### Flow Overview

1. **Entry Point**: `HandleClassification` → `generateEnhancedClassification`
2. **Classification Routing**:
   - **Fast Path**: If context deadline < 5s, use lightweight model
   - **Early Termination**: If Go classification confidence ≥ 0.85, skip ML
   - **Ensemble Voting**: If ML available and content sufficient, combine Go + ML
   - **Fallback**: Use Go classification only if ML unavailable

3. **Go Classification** (`runGoClassification`):
   - Uses `IndustryDetectionService.performClassification`
   - Multi-strategy classifier (keyword, description, name-based)
   - Confidence calibration
   - Layer routing (Layer 1, 2, 3)

4. **Python ML Classification** (`runPythonMLClassification`):
   - Model selection (lightweight vs full)
   - Enhanced classification with website scraping
   - **Issue**: Circuit breaker OPEN, so this is not being used

### Classification Routing Logic

**Location**: `services/classification-service/internal/handlers/classification.go:3497-3590`

```go
// Early termination threshold: 0.85 (very high)
if goResult.ConfidenceScore >= 0.85 {
    skipML = true  // Skip ML if Go has high confidence
}

// Ensemble voting enabled if:
// - ML service available
// - Website URL provided
// - Content length >= 50 characters
// - Not skipped due to time constraints
```

**Issues Identified**:
1. **Early termination threshold too high** (0.85) - Most requests won't meet this
2. **ML service unavailable** - Circuit breaker OPEN prevents ensemble voting
3. **Content quality validation** - May be too restrictive (50 chars minimum)

## Industry Detection Logic

### Multi-Strategy Classifier

**Location**: `internal/classification/service.go:341-487`

**Flow**:
1. Check cache
2. Run multi-strategy classifier (keyword, description, name-based)
3. Apply confidence calibration
4. Layer routing:
   - **Layer 1** (≥0.80 confidence): Use multi-strategy result
   - **Layer 2** (<0.80 confidence): Try embeddings if available
   - **Layer 3** (ambiguous cases): Try LLM if available

**Confidence Thresholds**:
- `layer2Threshold = 0.80` - Try Layer 2 if below this
- `highConfidenceThreshold = 0.95` - Use Layer 1 if above this
- `EarlyTerminationConfidenceThreshold = 0.85` (default)

**Issues Identified**:
1. **Confidence thresholds may be too high** - Most requests have low confidence (24.65% avg)
2. **Fallback to "General Business"** - When classification fails or confidence too low
3. **Keyword matching may be insufficient** - Many industries have 0% accuracy

### Fallback Logic

**Location**: `internal/classification/service.go:1204-1209`

```go
if classification == nil {
    return &IndustryDetectionResult{
        IndustryName: "General Business",  // Default fallback
        Confidence:   0.30,                 // Low confidence
        Keywords:     keywords,
        Reasoning:    "No matching industry found in database",
    }, nil
}
```

**Issue**: Defaulting to "General Business" with 0.30 confidence when no match found.

## Test Results Analysis

### Industry Accuracy Breakdown

| Industry | Accuracy | Status | Issue |
|----------|----------|--------|-------|
| **banking** | 100.0% | ✅ | Working correctly |
| **technology** | 20.7% | ❌ | Low accuracy |
| **healthcare** | 17.9% | ❌ | Low accuracy |
| **manufacturing** | 15.2% | ❌ | Low accuracy |
| **retail** | 16.3% | ❌ | Low accuracy |
| **financial services** | 10.5% | ❌ | Very low accuracy |
| **real estate** | 5.6% | ❌ | Very low accuracy |
| **arts & entertainment** | 0.0% | ❌ | Complete failure |
| **construction** | 0.0% | ❌ | Complete failure |
| **energy** | 0.0% | ❌ | Complete failure |
| **food & beverage** | 0.0% | ❌ | Complete failure |
| **professional services** | 0.0% | ❌ | Complete failure |
| **transportation** | 0.0% | ❌ | Complete failure |

### Key Findings

1. **Only banking has 100% accuracy** - All other industries are failing
2. **6 industries have 0% accuracy** - Complete classification failure
3. **Average confidence is 24.65%** - Very low, indicating uncertainty
4. **Most classifications defaulting to "General Business"** - Per test results

## Root Cause Analysis

### Primary Issues

1. **Python ML Service Unavailable** ⚠️ **CRITICAL**
   - Circuit breaker is OPEN
   - Error: "Circuit breaker is OPEN - request rejected"
   - **Impact**: Forcing fallback to keyword-based classification only
   - **Fix**: Track 6.1 - Verify Python ML service connectivity

2. **Confidence Thresholds Too High** ⚠️ **HIGH**
   - Early termination: 0.85 (most requests won't meet this)
   - Layer 2 threshold: 0.80 (most requests below this)
   - **Impact**: Most requests using low-confidence keyword classification
   - **Fix**: Adjust thresholds based on actual confidence distribution

3. **Keyword Matching Insufficient** ⚠️ **HIGH**
   - Many industries have 0% accuracy
   - Keyword patterns may not match test data
   - **Impact**: Incorrect industry classification
   - **Fix**: Review and improve keyword patterns

4. **Defaulting to "General Business"** ⚠️ **MEDIUM**
   - When no match found, defaults to "General Business" with 0.30 confidence
   - **Impact**: Low accuracy scores
   - **Fix**: Improve fallback logic or industry matching

5. **Content Quality Validation** ⚠️ **MEDIUM**
   - Minimum 50 characters for ML service
   - May be too restrictive
   - **Impact**: Some requests may not use ML even when available
   - **Fix**: Review content quality requirements

## Recommendations

### Immediate Actions (High Priority)

1. **Fix Python ML Service** (Track 6.1):
   - Check circuit breaker status
   - Verify service availability
   - Test service manually
   - **Expected Impact**: Enable ensemble voting, improve accuracy

2. **Adjust Confidence Thresholds**:
   - Reduce early termination threshold from 0.85 to 0.70
   - Reduce Layer 2 threshold from 0.80 to 0.60
   - **Expected Impact**: More requests using higher-quality classification methods

3. **Improve Keyword Patterns**:
   - Review industries with 0% accuracy
   - Add missing keywords
   - Test keyword matching against test data
   - **Expected Impact**: Improve accuracy for failing industries

### Medium Priority Actions

4. **Improve Fallback Logic**:
   - Instead of defaulting to "General Business", try fuzzy matching
   - Use confidence-based industry selection
   - **Expected Impact**: Reduce "General Business" defaults

5. **Review Content Quality Validation**:
   - Reduce minimum content length if too restrictive
   - Improve content quality scoring
   - **Expected Impact**: More requests using ML service

6. **Enhance Confidence Calibration**:
   - Review confidence calibration algorithm
   - Adjust calibration factors
   - **Expected Impact**: More accurate confidence scores

## Next Steps

1. ✅ **Complete Track 3.1 Investigation** - This document
2. **Track 6.1**: Verify Python ML service connectivity (CRITICAL)
3. **Track 3.2**: Fix confidence score calculation
4. **Implement Threshold Adjustments**: Based on findings
5. **Improve Keyword Patterns**: For industries with 0% accuracy
6. **Validate Fixes**: Run 50-sample E2E test

## Code Locations

- **Classification Flow**: `services/classification-service/internal/handlers/classification.go:3497-3590`
- **Industry Detection**: `internal/classification/service.go:341-487`
- **Multi-Strategy Classifier**: `internal/classification/multi_strategy_classifier.go`
- **Confidence Calibration**: `internal/classification/confidence_calibrator.go`
- **Fallback Logic**: `internal/classification/service.go:1204-1209`
- **Config Thresholds**: `services/classification-service/internal/config/config.go`

