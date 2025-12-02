# Railway Logs Fixes - Implementation Plan

**Date**: December 2, 2025  
**Priority**: High - Critical performance issue identified  
**Estimated Time**: 2-3 hours

---

## Executive Summary

This plan addresses issues identified in Railway logs analysis:
1. **CRITICAL**: ML Service Timeout (3s → 5-8s)
2. **MEDIUM**: Content Quality Thresholds (review and adjust)
3. **MEDIUM**: Relevance Thresholds (review and adjust)
4. **LOW**: Database Table Missing (optional)

---

## Phase 1: Critical Fix - ML Service Timeout

### Issue
Lightweight ML model requests timing out with 3-second timeout, causing fallback to slower full model.

### Impact
- Defeats fast-path optimization
- Causes request delays
- Reduces overall performance

### Implementation

#### Step 1.1: Update Timeout in PythonMLService

**File**: `internal/machine_learning/infrastructure/python_ml_service.go`  
**Line**: 442

**Current Code**:
```go
// Execute request with shorter timeout for fast path
fastCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
```

**New Code**:
```go
// Execute request with timeout for fast path
// Increased from 3s to 5s to allow ML inference time while maintaining fast-path performance
fastCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
```

**Rationale**:
- 5s provides enough time for ML inference
- Still maintains fast-path performance (<10s total)
- Balances speed and reliability

#### Step 1.2: Add Configuration Option (Optional Enhancement)

**File**: `internal/machine_learning/infrastructure/python_ml_service.go`

**Add to `PythonMLServiceConfig` struct** (around line 50-77):
```go
// Performance configuration
MaxBatchSize        int           `json:"max_batch_size"`
InferenceTimeout    time.Duration `json:"inference_timeout"`
ModelLoadingTimeout time.Duration `json:"model_loading_timeout"`
LightweightModelTimeout time.Duration `json:"lightweight_model_timeout"` // NEW
```

**Update `NewPythonMLService`** (around line 117):
```go
config: PythonMLServiceConfig{
    // ... existing config ...
    LightweightModelTimeout: 5 * time.Second, // Default 5s
},
```

**Update `ClassifyFast` method** (around line 442):
```go
// Use configurable timeout with default
lightweightTimeout := pms.config.LightweightModelTimeout
if lightweightTimeout == 0 {
    lightweightTimeout = 5 * time.Second // Default fallback
}
fastCtx, cancel := context.WithTimeout(ctx, lightweightTimeout)
defer cancel()
```

**Note**: This is optional but recommended for future flexibility.

### Testing

#### Unit Test
```go
func TestClassifyFast_Timeout(t *testing.T) {
    // Test that timeout is correctly set to 5s
    // Verify context deadline is ~5s from now
}
```

#### Integration Test
```go
func TestClassifyFast_WithMLService(t *testing.T) {
    // Test actual ML service call
    // Verify it completes within 5s
    // Verify no timeout errors
}
```

### Verification Steps

1. **Deploy to staging**
2. **Monitor logs for**:
   - No "context deadline exceeded" errors
   - Lightweight model success messages
   - Reduced fallback to full model
3. **Performance check**:
   - Request completion times should improve
   - Fast-path should work reliably

### Rollback Plan

If issues occur:
- Revert timeout to 3s
- Investigate ML service performance
- Check network latency

---

## Phase 2: Medium Priority - Content Quality Thresholds

### Issue
Content quality checks are too strict for fast-path mode, causing many pages to fail quality checks.

### Current Thresholds
- Minimum keywords: 10 unique
- Relevance threshold: 0.7

### Implementation

#### Step 2.1: Locate Threshold Configuration

**Files identified**:
- `internal/classification/smart_website_crawler.go` (lines 1383-1394)
  - Keyword threshold: 10 (line 1384)
  - Relevance threshold: 0.7 (line 1391)
- `internal/classification/website_content_service.go` (lines 322-323)
  - Keyword threshold: 10 (line 323)
  - Content length: 500 (line 322)

#### Step 2.2: Make Thresholds Configurable

**Option A: Add Fast-Path Specific Thresholds**

```go
// Fast-path mode thresholds (more lenient)
const (
    FastPathMinKeywords = 5  // Lowered from 10
    FastPathMinRelevance = 0.5 // Lowered from 0.7
    
    // Regular mode thresholds (stricter)
    RegularMinKeywords = 10
    RegularMinRelevance = 0.7
)
```

**Option B: Add to Config**

```go
type ClassificationConfig struct {
    // ... existing config ...
    
    // Content quality thresholds
    MinKeywordsForFastPath int     `json:"min_keywords_for_fast_path"` // Default: 5
    MinRelevanceForFastPath float64 `json:"min_relevance_for_fast_path"` // Default: 0.5
    MinKeywordsForRegular  int     `json:"min_keywords_for_regular"`   // Default: 10
    MinRelevanceForRegular float64 `json:"min_relevance_for_regular"`  // Default: 0.7
}
```

#### Step 2.3: Update Content Quality Check Logic

**File 1**: `internal/classification/smart_website_crawler.go` (around line 1383)

**Current Code** (lines 1383-1394):
```go
// 3. At least 10 unique keywords
if totalKeywords < 10 {
    c.logger.Printf("📊 [SmartCrawler] [ContentCheck] Insufficient keywords: %d < 10 unique", totalKeywords)
    return false
}

// 4. Average relevance score >= 0.7
avgRelevance := totalRelevance / float64(successfulPages)
if avgRelevance < 0.7 {
    c.logger.Printf("📊 [SmartCrawler] [ContentCheck] Insufficient relevance: %.2f < 0.7", avgRelevance)
    return false
}
```

**New Code** (with fast-path support):
```go
// Determine thresholds based on fast-path mode
minKeywords := 10
minRelevance := 0.7
if c.useFastPath { // Add useFastPath flag to SmartCrawler struct
    minKeywords = 5  // Lower threshold for fast-path
    minRelevance = 0.5 // Lower threshold for fast-path
}

// 3. At least N unique keywords
if totalKeywords < minKeywords {
    c.logger.Printf("📊 [SmartCrawler] [ContentCheck] Insufficient keywords: %d < %d unique", totalKeywords, minKeywords)
    return false
}

// 4. Average relevance score >= threshold
avgRelevance := totalRelevance / float64(successfulPages)
if avgRelevance < minRelevance {
    c.logger.Printf("📊 [SmartCrawler] [ContentCheck] Insufficient relevance: %.2f < %.2f", avgRelevance, minRelevance)
    return false
}
```

**File 2**: `internal/classification/website_content_service.go` (around line 322)

**Current Code** (lines 322-323):
```go
const minContentLength = 500
const minKeywordCount = 10
```

**New Code** (with fast-path support):
```go
// Make thresholds configurable based on fast-path mode
func (wcs *WebsiteContentService) isContentSufficient(textContent string, keywords []string, useFastPath bool) bool {
    minContentLength := 500
    minKeywordCount := 10
    if useFastPath {
        minContentLength = 300  // Lower for fast-path
        minKeywordCount = 5     // Lower for fast-path
    }
    // ... rest of function ...
}
```

### Testing

#### Unit Test
```go
func TestContentQualityCheck_FastPath(t *testing.T) {
    // Test with fast-path thresholds
    // Verify pages with 5+ keywords pass
    // Verify pages with 0.5+ relevance pass
}

func TestContentQualityCheck_Regular(t *testing.T) {
    // Test with regular thresholds
    // Verify stricter requirements
}
```

### Verification Steps

1. **Deploy to staging**
2. **Monitor logs for**:
   - More pages passing quality checks
   - Reduced "Insufficient keywords" warnings
   - Reduced "Insufficient relevance" warnings
3. **Accuracy check**:
   - Monitor classification accuracy
   - Ensure quality doesn't degrade
   - Track keyword extraction success rate

### Rollback Plan

If accuracy degrades:
- Revert to original thresholds
- Investigate why thresholds were set high
- Consider gradual adjustment

---

## Phase 3: Medium Priority - Relevance Thresholds

### Issue
Relevance threshold of 0.7 may be too high for fast-path mode.

### Implementation

**Note**: This is closely related to Phase 2. If implementing Phase 2, this is already covered.

**If implementing separately**:

#### Step 3.1: Update Relevance Calculation

**Find relevance calculation function** and update threshold:

```go
// Use lower threshold for fast-path
relevanceThreshold := 0.7
if useFastPath {
    relevanceThreshold = 0.5 // More lenient for fast-path
}
```

### Testing

Same as Phase 2 testing.

---

## Phase 4: Low Priority - Database Table

### Issue
Classification accuracy tracking table missing.

### Implementation

#### Option A: Create Table (if accuracy tracking needed)

**SQL Migration**:
```sql
CREATE TABLE IF NOT EXISTS public.classification_accuracy_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id VARCHAR(255) NOT NULL,
    business_name VARCHAR(255),
    predicted_industry VARCHAR(255),
    actual_industry VARCHAR(255),
    confidence_score DECIMAL(5,4),
    accuracy_score DECIMAL(5,4),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_classification_accuracy_request_id ON public.classification_accuracy_tracking(request_id);
CREATE INDEX idx_classification_accuracy_created_at ON public.classification_accuracy_tracking(created_at);
```

#### Option B: Disable Feature (if not needed)

**Find accuracy tracking code** and add feature flag:

```go
if config.EnableAccuracyTracking {
    // Save accuracy tracking
} else {
    // Skip (log warning if needed)
}
```

### Recommendation

**Option B** (disable) is recommended for MVP unless accuracy tracking is critical.

---

## Implementation Order

### Priority 1: Critical (Do First)
1. ✅ **Phase 1: ML Service Timeout** - Fix immediately
   - Time: 30-60 minutes
   - Risk: Low
   - Impact: High

### Priority 2: Medium (Do After Critical)
2. ⚠️ **Phase 2: Content Quality Thresholds** - Review and adjust
   - Time: 1-2 hours
   - Risk: Medium (may affect accuracy)
   - Impact: Medium

3. ⚠️ **Phase 3: Relevance Thresholds** - Review and adjust
   - Time: Included in Phase 2
   - Risk: Medium
   - Impact: Medium

### Priority 3: Low (Optional)
4. ℹ️ **Phase 4: Database Table** - Optional
   - Time: 30 minutes
   - Risk: Low
   - Impact: Low

---

## Testing Strategy

### Unit Tests
- Test timeout configuration
- Test content quality checks with new thresholds
- Test relevance calculations

### Integration Tests
- Test ML service calls with new timeout
- Test fast-path mode with adjusted thresholds
- Test end-to-end classification flow

### Performance Tests
- Benchmark request completion times
- Monitor ML service response times
- Track fast-path success rate

### Production Monitoring
- Monitor logs for timeout errors
- Track content quality check pass rates
- Monitor classification accuracy

---

## Deployment Plan

### Step 1: Phase 1 (Critical Fix)
1. Implement ML service timeout fix
2. Run unit tests
3. Run integration tests
4. Deploy to staging
5. Monitor for 24 hours
6. Deploy to production
7. Monitor logs for improvements

### Step 2: Phase 2 & 3 (Medium Priority)
1. Implement threshold adjustments
2. Run unit tests
3. Run integration tests
4. Deploy to staging
5. Monitor accuracy for 48 hours
6. If accuracy maintained, deploy to production
7. Continue monitoring

### Step 3: Phase 4 (Optional)
1. Decide on approach (create table vs disable)
2. Implement chosen approach
3. Deploy when convenient

---

## Success Criteria

### Phase 1 Success
- ✅ No "context deadline exceeded" errors for lightweight model
- ✅ Lightweight model success rate > 90%
- ✅ Reduced fallback to full model
- ✅ Request completion times improved

### Phase 2 & 3 Success
- ✅ More pages passing quality checks (target: 50%+ improvement)
- ✅ Classification accuracy maintained or improved
- ✅ Reduced "Insufficient" warnings in logs

### Overall Success
- ✅ Fast-path mode working reliably
- ✅ ML service calls succeeding
- ✅ Performance improved
- ✅ No accuracy degradation

---

## Risk Assessment

### Phase 1: Low Risk
- Simple timeout change
- Easy to rollback
- Low chance of breaking changes

### Phase 2 & 3: Medium Risk
- May affect classification accuracy
- Need careful monitoring
- May need adjustment based on results

### Phase 4: Low Risk
- Optional feature
- Can be disabled if issues

---

## Rollback Procedures

### Phase 1 Rollback
1. Revert timeout to 3s
2. Investigate ML service performance
3. Check network latency

### Phase 2 & 3 Rollback
1. Revert thresholds to original values
2. Monitor accuracy recovery
3. Investigate why thresholds were high

---

## Monitoring and Metrics

### Key Metrics to Track

1. **ML Service Timeout Fix**:
   - Lightweight model success rate
   - Timeout error count
   - Request completion times
   - Fallback to full model rate

2. **Content Quality Thresholds**:
   - Pages passing quality checks
   - "Insufficient" warning count
   - Classification accuracy
   - Keyword extraction success rate

3. **Overall Performance**:
   - Fast-path success rate
   - Average request completion time
   - Error rate
   - User satisfaction metrics

### Dashboard Queries

```sql
-- ML Service Success Rate
SELECT 
    COUNT(*) FILTER (WHERE success = true) * 100.0 / COUNT(*) as success_rate
FROM ml_service_logs
WHERE timestamp > NOW() - INTERVAL '1 hour';

-- Content Quality Check Pass Rate
SELECT 
    COUNT(*) FILTER (WHERE quality_check_passed = true) * 100.0 / COUNT(*) as pass_rate
FROM website_scraping_logs
WHERE timestamp > NOW() - INTERVAL '1 hour';
```

---

## Timeline

### Week 1
- **Day 1**: Implement Phase 1 (Critical Fix)
- **Day 2**: Test and deploy Phase 1
- **Day 3-4**: Monitor Phase 1 in production

### Week 2
- **Day 1-2**: Implement Phase 2 & 3
- **Day 3**: Test Phase 2 & 3
- **Day 4-5**: Deploy and monitor Phase 2 & 3

### Week 3 (Optional)
- **Day 1**: Implement Phase 4 if needed
- **Day 2**: Deploy Phase 4

---

## Files to Modify

### Phase 1
- `internal/machine_learning/infrastructure/python_ml_service.go` (line 442)

### Phase 2 & 3
- `internal/classification/website_content_service.go` (or similar)
- `services/classification-service/internal/config/config.go` (add thresholds)
- Content quality check functions

### Phase 4
- Database migration file (if creating table)
- Accuracy tracking code (if disabling)

---

## Dependencies

- No external dependencies
- All changes are internal
- No breaking API changes

---

## Notes

1. **ML Service Timeout**: Start with 5s, can increase to 8s if needed
2. **Content Quality**: Start conservative, adjust based on results
3. **Monitoring**: Critical for Phase 2 & 3 to ensure accuracy maintained
4. **Gradual Rollout**: Consider feature flags for Phase 2 & 3

---

## References

- **Analysis Documents**:
  - `docs/railway-logs-comprehensive-analysis.md`
  - `docs/railway-logs-issues-and-recommendations.md`
- **Code Locations**:
  - `internal/machine_learning/infrastructure/python_ml_service.go:442`
  - Content quality check functions (to be located)

---

## Approval

- [ ] Code review completed
- [ ] Tests passing
- [ ] Staging deployment successful
- [ ] Production deployment approved

---

**Plan Created**: December 2, 2025  
**Last Updated**: December 2, 2025  
**Status**: Ready for Implementation

