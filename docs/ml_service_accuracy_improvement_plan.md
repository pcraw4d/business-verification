# ML Service Accuracy Improvement Plan

**Date**: 2025-11-30  
**Test Run**: Railway Production Test (`accuracy_report_railway_production_20251130_235426.json`)  
**Status**: 🔴 **Critical Issues Identified**

---

## Executive Summary

The comprehensive accuracy test against Railway production revealed **critical issues** preventing the ML service from improving classification accuracy:

1. **🔴 ML Service Not Being Used**: Circuit breaker is OPEN, blocking all ML service requests
2. **🔴 Zero Industry Accuracy**: 0.00% (target: 95%) - all classifications defaulting to "General Business"
3. **🔴 Low Code Accuracy**: 4.11% (target: 90%) - MCC: 9.33%, NAICS: 0.72%, SIC: 2.26%
4. **⚠️ High Processing Time**: 8.35s average per classification
5. **⚠️ Circuit Breaker Configuration**: Opens after 5 failures, needs 2 successes to close, stays open 30s

---

## Test Results Analysis

### Overall Metrics

| Metric              | Actual | Target | Status        |
| ------------------- | ------ | ------ | ------------- |
| Overall Accuracy    | 2.46%  | N/A    | 🔴 Critical   |
| Industry Accuracy   | 0.00%  | 95%    | 🔴 Critical   |
| Code Accuracy       | 4.11%  | 90%    | 🔴 Critical   |
| MCC Accuracy        | 9.33%  | N/A    | 🔴 Poor       |
| NAICS Accuracy      | 0.72%  | N/A    | 🔴 Critical   |
| SIC Accuracy        | 2.26%  | N/A    | 🔴 Critical   |
| Avg Processing Time | 8.35s  | <3s    | ⚠️ High       |
| Total Test Cases    | 184    | -      | ✅ All Passed |

### Category Breakdown

| Category              | Accuracy | Test Cases | Status      |
| --------------------- | -------- | ---------- | ----------- |
| Technology            | 4.84%    | 42         | 🔴 Poor     |
| Healthcare            | 3.19%    | 47         | 🔴 Poor     |
| Transportation        | 3.33%    | 6          | 🔴 Poor     |
| Retail                | 1.67%    | 24         | 🔴 Critical |
| Manufacturing         | 1.11%    | 9          | 🔴 Critical |
| Financial Services    | 0.97%    | 31         | 🔴 Critical |
| Professional Services | 0.00%    | 10         | 🔴 Critical |
| Construction          | 0.00%    | 5          | 🔴 Critical |
| Edge Cases            | 0.00%    | 10         | 🔴 Critical |

---

## Root Cause Analysis

### Issue 1: ML Service Circuit Breaker OPEN 🔴

**Problem**: The Python ML service circuit breaker is in OPEN state, preventing all ML classification requests.

**Evidence from Test Logs**:

```
⚠️ [CircuitBreaker] Circuit is OPEN - failing fast for Python ML service
⚠️ Python ML service classification failed: circuit breaker is open: Python ML service unavailable, falling back to Go ML classifier
```

**Root Causes**:

1. **Initialization Failures**: During initialization, the service likely failed 5+ times, opening the circuit
2. **Service Unavailability**: The Railway Python ML service may be experiencing issues or timeouts
3. **Circuit Breaker Configuration**:
   - Opens after 5 consecutive failures
   - Stays open for 30 seconds
   - Requires 2 successes to close
   - If service continues to fail, circuit stays open indefinitely

**Impact**:

- **100% of ML classifications are being blocked**
- System falls back to Go ML classifier, which has 0% industry accuracy
- No Phase 1 enhancements (keyword extraction, ML input enhancement) are being used

**Evidence**: Test logs show repeated circuit breaker warnings, and no successful ML classifications.

### Issue 2: Zero Industry Accuracy 🔴

**Problem**: All classifications are defaulting to "General Business" with low confidence scores.

**Evidence from Test Results**:

- Industry Accuracy: 0.00%
- Most actual industries: "General Business" (confidence: 28-42%)
- Expected industries: Construction, Healthcare, Technology, etc.

**Root Causes**:

1. **ML Service Not Available**: Circuit breaker blocking ML service
2. **Go ML Classifier Ineffective**: Fallback classifier not working properly
3. **Keyword Matching Issues**: Keyword-based classification not matching industries correctly
4. **Low Confidence Thresholds**: System accepting low-confidence "General Business" classifications

**Impact**:

- Cannot distinguish between different industries
- All businesses classified as "General Business"
- Industry-specific code generation fails

### Issue 3: Low Code Accuracy 🔴

**Problem**: Classification codes (MCC, NAICS, SIC) have very low accuracy rates.

**Metrics**:

- MCC: 9.33% (best performing)
- NAICS: 0.72% (critical)
- SIC: 2.26% (critical)

**Root Causes**:

1. **Industry Detection Failure**: Without correct industry, code generation fails
2. **Keyword-to-Code Mapping Issues**: Keywords not matching to correct codes
3. **Crosswalk Validation Issues**: Crosswalk consistency score: 0.00% (from logs)
4. **Low Confidence Thresholds**: Accepting codes with very low relevance scores

**Impact**:

- Generated codes don't match expected codes
- Business verification fails
- Compliance issues

### Issue 4: High Processing Time ⚠️

**Problem**: Average processing time is 8.35 seconds per classification.

**Root Causes**:

1. **Website Scraping**: Enhanced website scraper taking 9+ seconds (seen in logs)
2. **Database Queries**: Multiple queries for keywords, industries, codes
3. **Fallback Processing**: Go ML classifier processing time
4. **No Caching**: Repeated queries for same data

**Impact**:

- Poor user experience
- High resource usage
- Scalability issues

### Issue 5: Circuit Breaker Configuration ⚠️

**Problem**: Circuit breaker configuration may be too aggressive for production environment.

**Current Configuration**:

- Failure Threshold: 5 failures
- Timeout: 30 seconds
- Success Threshold: 2 successes

**Issues**:

- If service is intermittently available, circuit may never close
- 30-second timeout may be too short for recovery
- No automatic retry mechanism after circuit opens

---

## Action Plan

### Phase 1: Fix ML Service Integration (Priority: 🔴 Critical)

#### Task 1.1: Diagnose Python ML Service Issues

**Owner**: DevOps/Backend  
**Timeline**: 1 day  
**Steps**:

1. Check Railway Python ML service health endpoint
2. Verify service is responding to `/ping` and `/health`
3. Test `/classify-enhanced` endpoint directly
4. Check service logs for errors
5. Verify service has sufficient resources (CPU, memory)

**Success Criteria**:

- Service responds to all endpoints
- No errors in service logs
- Service has adequate resources

#### Task 1.2: Fix Circuit Breaker Configuration

**Owner**: Backend  
**Timeline**: 2 hours  
**Steps**:

1. Review circuit breaker state during initialization
2. Add circuit breaker reset mechanism for initialization
3. Adjust circuit breaker thresholds for production:
   - Increase failure threshold to 10 (from 5)
   - Increase timeout to 60 seconds (from 30)
   - Add exponential backoff for retries
4. Add circuit breaker state logging
5. Implement circuit breaker health check endpoint

**Code Changes**:

```go
// internal/machine_learning/infrastructure/python_ml_service.go
// Update circuit breaker config:
circuitBreakerConfig.FailureThreshold = 10  // Increase from 5
circuitBreakerConfig.Timeout = 60 * time.Second // Increase from 30s
circuitBreakerConfig.SuccessThreshold = 2 // Keep at 2

// Add reset method for initialization
func (pms *PythonMLService) ResetCircuitBreaker() {
    pms.circuitBreaker.Reset()
}
```

**Success Criteria**:

- Circuit breaker resets during initialization
- Circuit breaker allows requests when service is available
- Proper logging of circuit breaker state changes

#### Task 1.3: Improve Initialization Resilience

**Owner**: Backend  
**Timeline**: 4 hours  
**Steps**:

1. Make model loading non-blocking (already done, verify)
2. Add retry logic for initialization
3. Add health check before marking service as ready
4. Implement graceful degradation: continue if ML service unavailable
5. Add initialization status endpoint

**Code Changes**:

```go
// Add retry logic with exponential backoff
func (pms *PythonMLService) InitializeWithRetry(ctx context.Context, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := pms.Initialize(ctx)
        if err == nil {
            return nil
        }
        waitTime := time.Duration(i+1) * 2 * time.Second
        time.Sleep(waitTime)
    }
    return fmt.Errorf("initialization failed after %d retries", maxRetries)
}
```

**Success Criteria**:

- Initialization succeeds even if service is temporarily unavailable
- Service can recover from transient failures
- Clear logging of initialization status

#### Task 1.4: Add ML Service Monitoring

**Owner**: Backend/DevOps  
**Timeline**: 4 hours  
**Steps**:

1. Add metrics for circuit breaker state
2. Track ML service request success/failure rates
3. Monitor ML service response times
4. Alert on circuit breaker opening
5. Dashboard for ML service health

**Success Criteria**:

- Real-time visibility into ML service health
- Alerts when circuit breaker opens
- Metrics for service availability

---

### Phase 2: Improve Industry Detection (Priority: 🔴 Critical)

#### Task 2.1: Fix Go ML Classifier Fallback

**Owner**: Backend  
**Timeline**: 1 day  
**Steps**:

1. Investigate why Go ML classifier returns "General Business" for all inputs
2. Check model loading and initialization
3. Verify model is being used correctly
4. Test with sample inputs
5. Fix or replace Go ML classifier if broken

**Success Criteria**:

- Go ML classifier returns correct industries
- Industry accuracy > 50% with Go ML classifier alone
- Proper confidence scores

#### Task 2.2: Improve Keyword-Based Industry Detection

**Owner**: Backend  
**Timeline**: 2 days  
**Steps**:

1. Review keyword-to-industry mappings in database
2. Add more industry-specific keywords
3. Improve keyword extraction algorithm
4. Increase keyword matching relevance thresholds
5. Add industry-specific keyword patterns

**Success Criteria**:

- Keyword-based industry accuracy > 60%
- Better industry detection for all categories
- Reduced "General Business" classifications

#### Task 2.3: Enhance Multi-Method Classification

**Owner**: Backend  
**Timeline**: 2 days  
**Steps**:

1. Review method weighting algorithm
2. Improve ensemble confidence calculation
3. Add industry-specific method preferences
4. Improve crosswalk consistency scoring
5. Add method result validation

**Success Criteria**:

- Multi-method accuracy > 70%
- Better combination of keyword + ML + description methods
- Higher confidence scores

---

### Phase 3: Improve Code Generation (Priority: 🔴 Critical)

#### Task 3.1: Fix Code Generation Logic

**Owner**: Backend  
**Timeline**: 2 days  
**Steps**:

1. Review code generation for each industry
2. Fix industry-to-code mappings
3. Improve keyword-to-code matching
4. Increase minimum relevance thresholds
5. Add code validation logic

**Success Criteria**:

- MCC accuracy > 50%
- NAICS accuracy > 40%
- SIC accuracy > 40%

#### Task 3.2: Improve Crosswalk Validation

**Owner**: Backend  
**Timeline**: 1 day  
**Steps**:

1. Review crosswalk consistency calculation (currently 0.00%)
2. Fix crosswalk data in database
3. Improve crosswalk validation algorithm
4. Add crosswalk-based code suggestions
5. Use crosswalk to improve code accuracy

**Success Criteria**:

- Crosswalk consistency score > 0.50
- Crosswalk improves code accuracy
- Better code alignment across MCC/NAICS/SIC

#### Task 3.3: Enhance Code Confidence Scoring

**Owner**: Backend  
**Timeline**: 1 day  
**Steps**:

1. Review confidence threshold logic
2. Increase minimum confidence for code acceptance
3. Add code relevance scoring
4. Improve code ranking algorithm
5. Filter out low-confidence codes

**Success Criteria**:

- Only high-confidence codes are returned
- Code accuracy improves
- Fewer incorrect codes in results

---

### Phase 4: Performance Optimization (Priority: ⚠️ High)

#### Task 4.1: Optimize Website Scraping

**Owner**: Backend  
**Timeline**: 1 day  
**Steps**:

1. Add timeout to website scraping (currently taking 9+ seconds)
2. Implement caching for scraped content
3. Add parallel scraping for multiple URLs
4. Optimize content extraction
5. Add scraping retry logic with backoff

**Code Changes**:

```go
// Add timeout to website scraping
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

// Add caching
cacheKey := fmt.Sprintf("website:%s", websiteURL)
if cached, err := cache.Get(cacheKey); err == nil {
    return cached
}
```

**Success Criteria**:

- Website scraping < 3 seconds
- Caching reduces redundant requests
- Better error handling

#### Task 4.2: Optimize Database Queries

**Owner**: Backend  
**Timeline**: 1 day  
**Steps**:

1. Add query result caching
2. Batch database queries
3. Optimize keyword lookup queries
4. Add database connection pooling
5. Use prepared statements

**Success Criteria**:

- Database query time reduced by 50%
- Fewer database round trips
- Better connection management

#### Task 4.3: Add Response Caching

**Owner**: Backend  
**Timeline**: 4 hours  
**Steps**:

1. Cache classification results by business name
2. Cache industry detection results
3. Cache code generation results
4. Implement cache invalidation strategy
5. Add cache hit/miss metrics

**Success Criteria**:

- Cache hit rate > 30%
- Reduced processing time for cached requests
- Proper cache invalidation

---

### Phase 5: Testing and Validation (Priority: ⚠️ High)

#### Task 5.1: Create ML Service Integration Tests

**Owner**: Backend  
**Timeline**: 1 day  
**Steps**:

1. Create unit tests for circuit breaker
2. Create integration tests for ML service
3. Test fallback mechanisms
4. Test initialization retry logic
5. Test circuit breaker recovery

**Success Criteria**:

- All tests pass
- Coverage > 80%
- Tests validate all scenarios

#### Task 5.2: Re-run Accuracy Tests

**Owner**: QA/Backend  
**Timeline**: 1 day  
**Steps**:

1. Fix ML service issues (Phase 1)
2. Re-run comprehensive accuracy tests
3. Compare results with baseline
4. Verify improvements
5. Document results

**Success Criteria**:

- Industry accuracy > 50%
- Code accuracy > 40%
- ML service being used successfully
- Processing time < 5 seconds

#### Task 5.3: Performance Testing

**Owner**: Backend  
**Timeline**: 1 day  
**Steps**:

1. Load test classification service
2. Test with ML service enabled
3. Test with ML service disabled (fallback)
4. Measure response times
5. Identify bottlenecks

**Success Criteria**:

- Service handles expected load
- Response times meet targets
- No performance regressions

---

## Success Metrics

### Immediate Goals (Week 1)

- ✅ ML service circuit breaker functioning correctly
- ✅ ML service being used for classifications
- ✅ Industry accuracy > 30%
- ✅ Code accuracy > 20%
- ✅ Processing time < 6 seconds

### Short-term Goals (Month 1)

- ✅ Industry accuracy > 60%
- ✅ Code accuracy > 50%
- ✅ MCC accuracy > 60%
- ✅ NAICS accuracy > 50%
- ✅ SIC accuracy > 50%
- ✅ Processing time < 4 seconds

### Long-term Goals (Month 3)

- ✅ Industry accuracy > 85% (target: 95%)
- ✅ Code accuracy > 80% (target: 90%)
- ✅ All code types > 80% accuracy
- ✅ Processing time < 3 seconds
- ✅ ML service availability > 99%

---

## Risk Assessment

### High Risk Issues

1. **ML Service Unavailable**: If Railway service is down, all ML classifications fail
   - **Mitigation**: Improve fallback mechanisms, add service redundancy
2. **Circuit Breaker Stuck Open**: If service has persistent issues, circuit never closes
   - **Mitigation**: Add manual reset, improve recovery logic
3. **Data Quality Issues**: If keyword/industry mappings are wrong, accuracy won't improve
   - **Mitigation**: Data quality audit, improve mappings

### Medium Risk Issues

1. **Performance Degradation**: High processing times may impact user experience
   - **Mitigation**: Caching, query optimization, parallel processing
2. **Code Generation Accuracy**: Low code accuracy may cause compliance issues
   - **Mitigation**: Improve code mappings, add validation

---

## Dependencies

1. **Railway Python ML Service**: Must be available and functioning
2. **Database**: Keyword, industry, and code mappings must be accurate
3. **Infrastructure**: Sufficient resources for ML service and classification service

---

## Timeline

| Phase                               | Duration    | Start Date | End Date |
| ----------------------------------- | ----------- | ---------- | -------- |
| Phase 1: Fix ML Service Integration | 2 days      | Day 1      | Day 2    |
| Phase 2: Improve Industry Detection | 5 days      | Day 3      | Day 7    |
| Phase 3: Improve Code Generation    | 4 days      | Day 8      | Day 11   |
| Phase 4: Performance Optimization   | 3 days      | Day 12     | Day 14   |
| Phase 5: Testing and Validation     | 3 days      | Day 15     | Day 17   |
| **Total**                           | **17 days** |            |          |

---

## Next Steps

1. **Immediate** (Today):

   - [ ] Diagnose Python ML service health
   - [ ] Check circuit breaker state
   - [ ] Review initialization logs

2. **This Week**:

   - [ ] Fix circuit breaker configuration
   - [ ] Improve initialization resilience
   - [ ] Fix Go ML classifier fallback

3. **Next Week**:
   - [ ] Improve industry detection
   - [ ] Improve code generation
   - [ ] Re-run accuracy tests

---

## References

- **Test Results**: `accuracy_report_railway_production_20251130_235426.json`
- **ML Service Investigation**: `docs/ml_service_investigation_summary.md`
- **Integration Phases Analysis**: `docs/integration_phases_test_results_analysis.md`
- **ML Testing Guide**: `docs/ml_accuracy_testing_guide.md`

---

## Conclusion

The test results reveal that **the ML service is not being used** due to circuit breaker issues, resulting in 0% industry accuracy and very low code accuracy. The action plan addresses these issues systematically, starting with fixing the ML service integration, then improving industry detection and code generation, followed by performance optimization and testing.

**Priority**: Fix ML service integration (Phase 1) is critical and should be addressed immediately.








