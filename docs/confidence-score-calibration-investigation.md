# Confidence Score Calibration Investigation - Track 3.2

## Executive Summary

Investigation of confidence score calculation reveals **critical issues** contributing to the **24.65% average confidence** (target: >70%). The primary issues are:

1. **Confidence Floor Too Low** - Floor of 0.30 (30%) allows low-confidence classifications
2. **Early Termination Threshold Too High** - 0.85 (85%) prevents most requests from using ML
3. **Calibration Factors May Not Be Applied** - Confidence calibration may not be working correctly
4. **Base Confidence Too Low** - Initial confidence calculations producing low scores

## Current Confidence Calculation

### Confidence Calibrator

**Location**: `internal/classification/confidence_calibrator.go:344-419`

**Calibration Factors**:
1. **Content Quality Boost** (Factor 1):
   - High quality (>0.8): +10% boost
   - Low quality (<0.5): -10% penalty

2. **Strategy Agreement** (Factor 2):
   - Strong agreement (variance <0.05): +15% boost
   - Moderate agreement (variance <0.10): +8% boost
   - High disagreement (variance >0.25): -15% penalty

3. **Code Agreement** (Factor 3):
   - Strong alignment (>0.85): +20% boost
   - Moderate alignment (>0.70): +10% boost
   - Poor alignment (<0.40): -15% penalty

4. **Method-Specific Calibration** (Factor 4):
   - Multi-strategy: +5% boost
   - Keyword-dominant (>0.85): +15% boost
   - ML-validated: +10% boost
   - Fast-path keyword: +12% boost

5. **Historical Accuracy Adjustment** (Factor 5):
   - Uses calibration bins to adjust based on historical accuracy

**Confidence Bounds**:
- **Floor**: 0.30 (30%) - Minimum confidence
- **Ceiling**: 0.95 (95%) - Maximum confidence

**Issues Identified**:
1. **Floor too low** - 0.30 allows very low-confidence classifications
2. **Calibration may not be applied** - If base confidence is very low, calibration may not help enough
3. **Code agreement factor** - May not be available when codes aren't generated

### Base Confidence Calculation

**Location**: `internal/classification/confidence_calibrator.go:421-445`

**Weighted Average Calculation**:
- Keyword: 40% weight
- Entity: 25% weight
- Topic: 20% weight
- Co-occurrence: 15% weight

**Issue**: If strategy results are low, weighted average will also be low.

### Confidence Thresholds

**Location**: `services/classification-service/internal/config/config.go:136`

**Current Thresholds**:
- **Early Termination**: 0.85 (85%) - Default
- **Code Generation**: 0.15 (15%) - Reduced from 0.5
- **Layer 2 Threshold**: 0.80 (80%) - In service.go

**Issues Identified**:
1. **Early termination too high** - 0.85 means most requests won't skip ML
2. **Layer 2 threshold too high** - 0.80 means most requests won't use embeddings

## Test Results Analysis

### Current Performance

- **Average Confidence**: 24.65% (target: >70%)
- **Overall Accuracy**: 10.7% (target: ≥80%)
- **High Confidence Results**: Very few (most below 30%)

### Confidence Distribution (Estimated)

Based on average confidence of 24.65%:
- **High Confidence (≥70%)**: <5% of results
- **Medium Confidence (30-69%)**: ~20-30% of results
- **Low Confidence (<30%)**: ~65-75% of results

## Root Cause Analysis

### Primary Issues

1. **Confidence Floor Too Low** ⚠️ **HIGH**
   - Floor of 0.30 allows low-confidence classifications
   - **Impact**: Many results have confidence <30%
   - **Fix**: Increase floor to 0.50 or adjust calibration to boost low scores

2. **Base Confidence Too Low** ⚠️ **HIGH**
   - Initial confidence calculations producing low scores
   - **Impact**: Even with calibration, confidence remains low
   - **Fix**: Improve base confidence calculation or boost factors

3. **Calibration Factors Not Effective** ⚠️ **MEDIUM**
   - Calibration may not be applied correctly
   - Code agreement factor may not be available
   - **Impact**: Confidence not being boosted enough
   - **Fix**: Review calibration application and factor weights

4. **Thresholds Too High** ⚠️ **MEDIUM**
   - Early termination: 0.85 (too high)
   - Layer 2: 0.80 (too high)
   - **Impact**: Most requests using low-confidence keyword classification
   - **Fix**: Reduce thresholds to match actual confidence distribution

5. **ML Service Unavailable** ⚠️ **CRITICAL** (from Track 3.1)
   - Circuit breaker OPEN prevents ensemble voting
   - **Impact**: No ML boost to confidence scores
   - **Fix**: Track 6.1 - Fix Python ML service

## Recommendations

### Immediate Actions (High Priority)

1. **Increase Confidence Floor**:
   - Change floor from 0.30 to 0.50
   - Or adjust calibration to boost low scores more aggressively
   - **Expected Impact**: Increase average confidence from 24.65% to 40-50%

2. **Boost Calibration Factors**:
   - Increase content quality boost from +10% to +20%
   - Increase strategy agreement boost from +15% to +25%
   - **Expected Impact**: Additional 10-15% confidence boost

3. **Reduce Early Termination Threshold**:
   - Reduce from 0.85 to 0.70
   - **Expected Impact**: More requests can skip ML when Go classification is confident

4. **Reduce Layer 2 Threshold**:
   - Reduce from 0.80 to 0.60
   - **Expected Impact**: More requests use embeddings for better accuracy

### Medium Priority Actions

5. **Improve Base Confidence Calculation**:
   - Review strategy weights
   - Add more strategies if needed
   - **Expected Impact**: Higher base confidence before calibration

6. **Fix ML Service** (Track 6.1):
   - Fix circuit breaker
   - Enable ensemble voting
   - **Expected Impact**: ML boost adds 10-20% to confidence

7. **Review Calibration Application**:
   - Ensure calibration is applied correctly
   - Verify all factors are being used
   - **Expected Impact**: Better confidence calibration

## Code Changes Required

### 1. Increase Confidence Floor

**File**: `internal/classification/confidence_calibrator.go:416`

```go
// Current:
calibratedConfidence = math.Max(calibratedConfidence, 0.30)

// Recommended:
calibratedConfidence = math.Max(calibratedConfidence, 0.50)
```

### 2. Boost Calibration Factors

**File**: `internal/classification/confidence_calibrator.go:359-388`

```go
// Factor 1: Content quality boost
if contentQuality > 0.8 {
    calibratedConfidence *= 1.20 // Increased from 1.10
} else if contentQuality < 0.5 {
    calibratedConfidence *= 0.90
}

// Factor 2: Strategy agreement
if strategyVariance < 0.05 {
    calibratedConfidence *= 1.25 // Increased from 1.15
} else if strategyVariance < 0.10 {
    calibratedConfidence *= 1.12 // Increased from 1.08
}
```

### 3. Reduce Thresholds

**File**: `services/classification-service/internal/config/config.go:136`

```go
// Current:
EarlyTerminationConfidenceThreshold: getEnvAsFloat("EARLY_TERMINATION_CONFIDENCE_THRESHOLD", 0.85),

// Recommended:
EarlyTerminationConfidenceThreshold: getEnvAsFloat("EARLY_TERMINATION_CONFIDENCE_THRESHOLD", 0.70),
```

**File**: `internal/classification/service.go:422`

```go
// Current:
const layer2Threshold = 0.80

// Recommended:
const layer2Threshold = 0.60
```

## Expected Impact

After implementing these fixes:

1. **Average Confidence**: 24.65% → 50-60% (target: >70%)
2. **High Confidence Results**: <5% → 30-40%
3. **Low Confidence Results**: ~70% → 30-40%

## Next Steps

1. ✅ **Complete Track 3.2 Investigation** - This document
2. **Implement Confidence Floor Increase** - Immediate fix
3. **Boost Calibration Factors** - Immediate fix
4. **Reduce Thresholds** - Immediate fix
5. **Track 6.1**: Fix Python ML service (CRITICAL for ensemble voting)
6. **Validate Fixes**: Run 50-sample E2E test

## Code Locations

- **Confidence Calibrator**: `internal/classification/confidence_calibrator.go:344-419`
- **Base Confidence**: `internal/classification/confidence_calibrator.go:421-445`
- **Thresholds**: `services/classification-service/internal/config/config.go:136`
- **Layer 2 Threshold**: `internal/classification/service.go:422`

