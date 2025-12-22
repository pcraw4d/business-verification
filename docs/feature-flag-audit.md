# Feature Flag Configuration Audit - Track 7.1

## Executive Summary

Investigation of feature flag configuration reveals **most critical flags are enabled by default**, but **verification of actual flag values in production is needed**. Feature flags control critical functionality including ML service, keyword matching, ensemble voting, and multi-page analysis.

**Status**: ⚠️ **MEDIUM** - Defaults look correct, but production values need verification

## Feature Flag Configuration

### Configuration Location

**Primary Config**: `services/classification-service/internal/config/config.go:102-146`

**Environment Variable Files**:
- `configs/feature-flags.env` - Modular architecture flags
- `configs/granular-feature-flags.env` - Granular model/service flags

**Railway Configuration**: `railway.json` (environment-specific variables)

### Critical Feature Flags

#### 1. ML Service Flags

**Flag**: `ML_ENABLED`
- **Location**: `config.go:112`
- **Default**: `true`
- **Environment Variable**: `ML_ENABLED`
- **Usage**: Controls whether ML service is used for classification
- **Impact**: ⚠️ **CRITICAL** - Disabling blocks ML-based classification
- **Status**: ✅ Default enabled

**Flag**: `ENSEMBLE_ENABLED`
- **Location**: `config.go:114`
- **Default**: `true`
- **Environment Variable**: `ENSEMBLE_ENABLED`
- **Usage**: Controls ensemble voting (Python ML + Go classification)
- **Impact**: ⚠️ **CRITICAL** - Disabling prevents ensemble voting
- **Status**: ✅ Default enabled

**Flag**: `PYTHON_ML_WEIGHT`
- **Location**: `config.go:140`
- **Default**: `0.60` (60%)
- **Environment Variable**: `PYTHON_ML_WEIGHT`
- **Usage**: Weight for Python ML service in ensemble voting
- **Impact**: ⚠️ **HIGH** - Affects ensemble voting balance
- **Status**: ✅ Default set

#### 2. Classification Method Flags

**Flag**: `KEYWORD_METHOD_ENABLED`
- **Location**: `config.go:113`
- **Default**: `true`
- **Environment Variable**: `KEYWORD_METHOD_ENABLED`
- **Usage**: Controls keyword-based classification
- **Impact**: ⚠️ **CRITICAL** - Disabling removes fallback classification method
- **Status**: ✅ Default enabled

**Flag**: `ENABLE_EARLY_TERMINATION`
- **Location**: `config.go:135`
- **Default**: `true`
- **Environment Variable**: `ENABLE_EARLY_TERMINATION`
- **Usage**: Controls early termination when confidence threshold is met
- **Impact**: ⚠️ **HIGH** - Affects classification flow and performance
- **Status**: ✅ Default enabled

**Flag**: `EARLY_TERMINATION_CONFIDENCE_THRESHOLD`
- **Location**: `config.go:136`
- **Default**: `0.85` (85%)
- **Environment Variable**: `EARLY_TERMINATION_CONFIDENCE_THRESHOLD`
- **Usage**: Confidence threshold for early termination
- **Impact**: ⚠️ **HIGH** - Too high may prevent ML service usage
- **Status**: ⚠️ **POTENTIAL ISSUE** - Default 0.85 may be too high (from Track 3.2)

#### 3. Scraping Flags

**Flag**: `ENABLE_MULTI_PAGE_ANALYSIS`
- **Location**: `config.go:123`
- **Default**: `true`
- **Environment Variable**: `ENABLE_MULTI_PAGE_ANALYSIS`
- **Usage**: Controls multi-page website analysis
- **Impact**: ⚠️ **HIGH** - Disabling reduces scraping success rate
- **Status**: ✅ Default enabled

**Flag**: `ENABLE_FAST_PATH_SCRAPING`
- **Location**: `config.go:126`
- **Default**: `true`
- **Environment Variable**: `ENABLE_FAST_PATH_SCRAPING`
- **Usage**: Controls fast-path scraping optimization
- **Impact**: ⚠️ **MEDIUM** - Affects scraping performance
- **Status**: ✅ Default enabled

**Flag**: `ENABLE_STRUCTURED_DATA_EXTRACTION`
- **Location**: `config.go:124`
- **Default**: `true`
- **Environment Variable**: `ENABLE_STRUCTURED_DATA_EXTRACTION`
- **Usage**: Controls structured data extraction from websites
- **Impact**: ⚠️ **MEDIUM** - Affects data quality
- **Status**: ✅ Default enabled

**Flag**: `ENABLE_WEBSITE_CONTENT_CACHE`
- **Location**: `config.go:133`
- **Default**: `true`
- **Environment Variable**: `ENABLE_WEBSITE_CONTENT_CACHE`
- **Usage**: Controls website content caching
- **Impact**: ⚠️ **MEDIUM** - Affects performance and cache hit rate
- **Status**: ✅ Default enabled

#### 4. Content Quality Flags

**Flag**: `MIN_CONTENT_LENGTH_FOR_ML`
- **Location**: `config.go:137`
- **Default**: `50` characters
- **Environment Variable**: `MIN_CONTENT_LENGTH_FOR_ML`
- **Usage**: Minimum content length required for ML service
- **Impact**: ⚠️ **MEDIUM** - Too high may prevent ML usage
- **Status**: ✅ Default set (may need review from Track 3.1)

**Flag**: `SKIP_FULL_CRAWL_IF_CONTENT_SUFFICIENT`
- **Location**: `config.go:138`
- **Default**: `true`
- **Environment Variable**: `SKIP_FULL_CRAWL_IF_CONTENT_SUFFICIENT`
- **Usage**: Skip full crawl if initial content is sufficient
- **Impact**: ⚠️ **MEDIUM** - Affects scraping behavior
- **Status**: ✅ Default enabled

### Granular Feature Flags

**Location**: `internal/config/granular_feature_flags.go`

**Service Flags**:
- `ENABLE_PYTHON_ML_SERVICE` (default: `true`)
- `ENABLE_GO_RULE_ENGINE` (default: `true`)
- `ENABLE_API_GATEWAY` (default: `true`)

**Model Flags**:
- `ENABLE_BERT_CLASSIFICATION` (default: `true`)
- `ENABLE_DISTILBERT_CLASSIFICATION` (default: `true`)
- `ENABLE_CUSTOM_NEURAL_NET` (default: `false`)
- `ENABLE_KEYWORD_MATCHING` (default: `true`)
- `ENABLE_MCC_CODE_LOOKUP` (default: `true`)

**Status**: ✅ Defaults look correct (most enabled)

## Feature Flag Usage

### ML Service Usage

**Location**: `services/classification-service/internal/handlers/classification.go:3520-3575`

**Logic**:
```go
// Check if we should use ensemble voting (Python ML available, sufficient content, and not skipped)
if !skipML && h.pythonMLService != nil && req.WebsiteURL != "" {
    // Ensemble voting enabled
}
```

**Conditions**:
1. `skipML` flag is false (based on adaptive timeout logic)
2. `pythonMLService` is not nil (service initialized)
3. `req.WebsiteURL` is not empty
4. `cfg.Classification.EnsembleEnabled` is true (default: true)
5. `cfg.Classification.MLEnabled` is true (default: true)

**Status**: ✅ Logic looks correct, but depends on service availability

### Keyword Method Usage

**Location**: `internal/classification/service.go` (keyword-based classification)

**Logic**:
- Uses `cfg.Classification.KeywordMethodEnabled` (default: true)
- Fallback when ML service unavailable or fails

**Status**: ✅ Should be enabled as fallback

### Multi-Page Analysis Usage

**Location**: `internal/classification/enhanced_website_scraper.go`

**Logic**:
- Uses `cfg.Classification.MultiPageAnalysisEnabled` (default: true)
- Controls whether multiple pages are crawled

**Status**: ✅ Default enabled, but may not be working (from Track 5.2)

## Investigation Steps

### Step 1: Check Railway Environment Variables

**Check Production Flags**:
```bash
# Via Railway dashboard or API
# Check these critical flags:
- ML_ENABLED
- ENSEMBLE_ENABLED
- KEYWORD_METHOD_ENABLED
- ENABLE_MULTI_PAGE_ANALYSIS
- ENABLE_FAST_PATH_SCRAPING
- ENABLE_EARLY_TERMINATION
- EARLY_TERMINATION_CONFIDENCE_THRESHOLD
```

**Status**: ⏳ **PENDING** - Need to verify in Railway dashboard

### Step 2: Review Feature Flag Logic

**Check Flag Evaluation**:
- Verify flags are being read correctly
- Check for any flag evaluation bugs
- Verify defaults are applied when flags not set

**Status**: ✅ **REVIEWED** - Logic looks correct

### Step 3: Test Feature Flag Impact

**Test Scenarios**:
1. ML service disabled → Should fallback to keyword method
2. Ensemble disabled → Should use single method
3. Multi-page analysis disabled → Should only analyze single page
4. Early termination disabled → Should always complete full flow

**Status**: ⏳ **PENDING** - Need to test

### Step 4: Check Flag Conflicts

**Potential Conflicts**:
- Early termination threshold too high → May prevent ML usage
- ML enabled but circuit breaker open → ML won't be used
- Multi-page enabled but crawling not working → No effect

**Status**: ⚠️ **IDENTIFIED** - Some conflicts exist

## Root Cause Analysis

### Potential Issues

1. **Early Termination Threshold Too High** ⚠️ **HIGH**
   - Default: 0.85 (85%)
   - From Track 3.2: Average confidence is 24.65%
   - **Impact**: Early termination may never trigger, or prevents ML usage
   - **Evidence**: Track 3.2 findings

2. **Flag Values Not Set in Production** ⚠️ **MEDIUM**
   - Defaults may not be applied if flags explicitly set to false
   - **Impact**: Critical functionality may be disabled
   - **Evidence**: Need to verify in Railway

3. **Flag Conflicts** ⚠️ **MEDIUM**
   - ML enabled but circuit breaker open → ML won't be used
   - Multi-page enabled but crawling not working → No effect
   - **Impact**: Flags appear enabled but functionality doesn't work
   - **Evidence**: Track 6.1 (circuit breaker OPEN)

4. **Content Quality Requirements** ⚠️ **LOW**
   - `MIN_CONTENT_LENGTH_FOR_ML` = 50 characters
   - May be too restrictive
   - **Impact**: Some requests may not use ML
   - **Evidence**: Track 3.1 findings

## Recommendations

### Immediate Actions (High Priority)

1. **Verify Production Flag Values**:
   - Check Railway dashboard for all critical flags
   - Ensure flags are set correctly
   - Document actual values

2. **Fix Early Termination Threshold**:
   - Reduce from 0.85 to 0.70 (from Track 3.2)
   - Update default in code
   - Set in Railway if needed

3. **Resolve Flag Conflicts**:
   - Fix ML service circuit breaker (Track 6.1)
   - Fix multi-page crawling (Track 5.2)
   - Ensure flags match actual functionality

### Medium Priority Actions

4. **Review Content Quality Requirements**:
   - Consider reducing `MIN_CONTENT_LENGTH_FOR_ML`
   - Test impact on ML usage

5. **Add Flag Monitoring**:
   - Log flag values on startup
   - Include flags in health endpoint
   - Alert on critical flags being disabled

6. **Document Flag Dependencies**:
   - Document which flags depend on others
   - Create flag dependency matrix
   - Add validation for conflicting flags

## Code Locations

- **Config Definition**: `services/classification-service/internal/config/config.go:102-146`
- **Flag Usage**: `services/classification-service/internal/handlers/classification.go:3520-3575`
- **Granular Flags**: `internal/config/granular_feature_flags.go`
- **Config Files**: `configs/feature-flags.env`, `configs/granular-feature-flags.env`

## Next Steps

1. ✅ **Complete Track 7.1 Investigation** - This document
2. **Verify Production Flag Values** - Check Railway dashboard
3. **Fix Early Termination Threshold** - Reduce to 0.70
4. **Resolve Flag Conflicts** - Fix underlying issues
5. **Add Flag Monitoring** - Include in health endpoint
6. **Document Flag Dependencies** - Create dependency matrix

## Expected Impact

After fixing issues:

1. **ML Service Usage**: Improved with correct flags and circuit breaker fix
2. **Classification Accuracy**: Improved with proper flag configuration
3. **Early Termination**: More effective with lower threshold
4. **Multi-Page Analysis**: Working with flag and crawling fixes

## References

- Config Implementation: `services/classification-service/internal/config/config.go`
- Handler Usage: `services/classification-service/internal/handlers/classification.go`
- Granular Flags: `internal/config/granular_feature_flags.go`
- Track 3.2: `docs/confidence-score-calibration-investigation.md`
- Track 6.1: `docs/python-ml-service-connectivity-audit.md`

