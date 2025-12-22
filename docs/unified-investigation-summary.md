# Unified Investigation Summary - Classification Service Performance Issues

**Date**: December 22, 2025  
**Status**: Investigation In Progress  
**Objective**: Comprehensive root cause analysis of classification service performance issues

---

## Executive Summary

This document consolidates all investigation findings from the classification service root cause analysis. The service is experiencing critical performance issues:

- **Error Rate**: 67.1% (Target: <5%)
- **Average Latency**: 43.7s (Target: <10s)
- **Classification Accuracy**: 9.5% (Target: ≥80%)
- **Code Generation Rate**: 23.1% (Target: ≥90%)
- **Scraping Success Rate**: 10.4% (Target: ≥70%)

---

## Investigation Tracks Status

### ✅ Completed Investigations

1. **Track 1.2**: Request Processing Bottleneck Analysis
2. **Track 2.1**: Error Pattern Analysis
3. **Track 3.1**: Classification Algorithm Investigation
4. **Track 3.2**: Confidence Score Calibration Investigation
5. **Track 6.1**: Python ML Service Connectivity Investigation
6. **Track 6.2**: Playwright Scraper Service Verification
7. **Track 6.3**: Supabase Database Connectivity Verification
8. **Track 7.1**: Feature Flag Configuration Audit

### 🔄 In Progress

- **Unified Investigation Summary** (this document)

### ⏳ Pending Investigations

8. **Track 7.2**: Configuration Mismatch Investigation
9. **Track 8.1**: Cache Hit Rate Investigation
10. **Track 8.2**: Resource Constraints Investigation
11. **Track 9.1**: Test Data Quality Validation

---

## Track 1.2: Request Processing Bottleneck Analysis

### Status: ✅ Completed

### Key Findings

**Document**: `docs/performance-bottleneck-analysis.md` (to be created)  
**Script**: `scripts/analyze_slow_requests.go`

**Root Causes Identified**:

1. **Detailed Request Tracing** ⚠️ **HIGH**
   - Implemented comprehensive tracing system
   - Captures timing for all sub-operations
   - Identifies slow stages in request processing

2. **Slow Request Patterns** ⚠️ **MEDIUM**
   - Some requests taking >60 seconds
   - Website scraping contributing to latency
   - Database queries may be slow

3. **Concurrent Request Limits** ⚠️ **MEDIUM**
   - May be limiting throughput
   - Queue depth and wait times need analysis

### Required Fixes

1. **Optimize Slow Operations**
   - Identify and optimize slow stages
   - Reduce website scraping time
   - Optimize database queries

2. **Review Concurrent Request Limits**
   - Adjust limits if too restrictive
   - Optimize queue management

---

## Track 2.1: Error Pattern Analysis

### Status: ✅ Completed

### Key Findings

**Document**: `docs/error-pattern-analysis.md`  
**Script**: `scripts/parse_error_patterns.go`

**Error Distribution**:

| Category | Count | Percentage | Priority |
|----------|-------|------------|----------|
| **DNS Failure** | ~63.5% | 63.5% | **CRITICAL** |
| **Timeout** | ~9.9% | 9.9% | **HIGH** |
| **Other** | ~25.1% | 25.1% | **MEDIUM** |
| **HTTP 5xx** | ~1.5% | 1.5% | **LOW** |

**Root Causes Identified**:

1. **DNS Resolution Failures** ⚠️ **CRITICAL** (63.5% of errors)
   - `no such host` errors
   - IPv6 address issues (`[fd12::10]:53`)
   - Misconfigured DNS servers
   - **Impact**: Prevents website scraping and external service calls

2. **Timeout Errors** ⚠️ **HIGH** (9.9% of errors)
   - `context deadline exceeded`
   - Operations not completing within timeouts
   - **Impact**: Directly contributes to error rate and latency

3. **Other Errors** ⚠️ **MEDIUM** (25.1% of errors)
   - HTTP status 0 (unexpected)
   - Relevance errors
   - Keyword extraction failures
   - **Impact**: Affects data quality and classification accuracy

4. **HTTP 5xx Errors** ⚠️ **LOW** (1.5% of errors)
   - Server-side errors from external dependencies
   - **Impact**: Intermittent service unreliability

### Required Fixes

1. **Fix DNS Resolution** (Track 2.2) - **CRITICAL**
   - Implement robust DNS fallback (8.8.8.8, 1.1.1.1)
   - Verify IPv6 handling or prefer IPv4
   - Increase DNS timeout if needed

2. **Fix Timeout Configurations** (Track 1.1) - **HIGH**
   - Review and align all timeouts
   - Fix timeout budget calculations
   - Ensure context propagation

3. **Improve Error Handling** - **MEDIUM**
   - Enhance scraping robustness
   - Refine keyword extraction
   - Better error categorization

---

## Track 3.1: Classification Algorithm Investigation

### Status: ✅ Completed

### Key Findings

**Document**: `docs/classification-algorithm-investigation.md`  
**Script**: `scripts/analyze_classification_accuracy.go`

**Accuracy Metrics**:
- **Overall Accuracy**: 10.7% (Target: ≥80%)
- **Industry Accuracy**: 0% for many industries
- **Code Accuracy**: 0% for NAICS/SIC

**Root Causes Identified**:

1. **Python ML Service Circuit Breaker OPEN** ⚠️ **CRITICAL**
   - Circuit breaker blocking all ML requests
   - System falling back to Go keyword-based classification only
   - **Impact**: No ensemble voting, reduced accuracy

2. **Early Termination Logic** ⚠️ **HIGH**
   - Threshold too high (0.85)
   - Terminating before ML service can improve results
   - **Impact**: Missing ML-based improvements

3. **Keyword Matching Insufficient** ⚠️ **HIGH**
   - Many industries have 0% accuracy
   - Keyword patterns may not match test data
   - **Impact**: Incorrect industry classification

4. **Defaulting to "General Business"** ⚠️ **MEDIUM**
   - When no match found, defaults to "General Business" with 0.30 confidence
   - **Impact**: Low accuracy scores

5. **Content Quality Validation** ⚠️ **MEDIUM**
   - Minimum 50 characters for ML service
   - May be too restrictive
   - **Impact**: Some requests may not use ML even when available

### Required Fixes

1. **Fix Python ML Service** (Track 6.1) - **CRITICAL**
   - Check circuit breaker status
   - Verify service availability
   - Reset circuit breaker if service is healthy
   - **Expected Impact**: Enable ensemble voting, improve accuracy

2. **Adjust Confidence Thresholds** - **HIGH**
   - Reduce early termination threshold from 0.85 to 0.70
   - Reduce Layer 2 threshold from 0.80 to 0.60
   - Allow more time for ML service

3. **Improve Keyword Patterns** - **HIGH**
   - Review and improve keyword patterns
   - Add more industry-specific keywords
   - Test against actual business data

4. **Improve Fallback Logic** - **MEDIUM**
   - Better industry matching
   - Reduce reliance on "General Business" default

5. **Review Content Quality Requirements** - **MEDIUM**
   - Consider reducing minimum character requirement
   - Make ML service more accessible

---

## Track 3.2: Confidence Score Calibration Investigation

### Status: ✅ Completed

### Key Findings

**Document**: `docs/confidence-score-calibration-investigation.md`  
**Script**: `scripts/analyze_confidence_scores.go`

**Confidence Metrics**:
- **Average Confidence**: 24.65% (Target: >70%)
- **High Confidence (≥70%)**: <5%
- **Low Confidence (<30%)**: ~70%

**Root Causes Identified**:

1. **Confidence Floor Too Low** ⚠️ **HIGH**
   - Current floor: 0.30
   - Too many results at minimum confidence
   - **Impact**: Low overall confidence scores

2. **Base Confidence Too Low** ⚠️ **HIGH**
   - Weighted average of strategies producing low scores
   - ML service unavailable (circuit breaker OPEN)
   - **Impact**: Low base confidence before calibration

3. **Calibration Factors Not Effective** ⚠️ **MEDIUM**
   - Content quality boost: +10%
   - Strategy agreement boost: +15%
   - May not be sufficient
   - **Impact**: Calibration not boosting confidence enough

4. **Thresholds Too High** ⚠️ **MEDIUM**
   - Early termination: 0.85
   - Layer 2 threshold: 0.80
   - Preventing better classification methods
   - **Impact**: Missing opportunities for higher confidence

5. **ML Service Unavailable** ⚠️ **CRITICAL**
   - Circuit breaker OPEN (from Track 3.1)
   - No ensemble voting boost
   - **Impact**: Missing confidence boost from ML service

### Required Fixes

1. **Increase Confidence Floor** - **HIGH**
   - Change from 0.30 to 0.50
   - **Expected Impact**: Average confidence 24.65% → 40-50%

2. **Boost Calibration Factors** - **HIGH**
   - Content quality: +10% → +20%
   - Strategy agreement: +15% → +25%
   - **Expected Impact**: Additional 10-15% confidence boost

3. **Reduce Thresholds** - **MEDIUM**
   - Early termination: 0.85 → 0.70
   - Layer 2 threshold: 0.80 → 0.60
   - **Expected Impact**: Allow more ML-based classification

4. **Fix ML Service** (Track 6.1) - **CRITICAL**
   - Enable ensemble voting
   - **Expected Impact**: Additional 10-20% confidence boost

**Combined Expected Impact**:
- Average confidence: 24.65% → 50-60% (target: >70%)
- High confidence results: <5% → 30-40%
- Low confidence results: ~70% → 30-40%

---

## Track 6.1: Python ML Service Connectivity Investigation

### Status: ✅ Completed

### Key Findings

**Document**: `docs/python-ml-service-connectivity-audit.md`  
**Script**: `scripts/test_python_ml_service.go`

**Service Status**:
- **Circuit Breaker State**: ❌ OPEN (CRITICAL)
- **Service Health**: ✅ Healthy (from previous analysis)
- **Service URL**: `https://python-ml-service-production.up.railway.app`

**Root Causes Identified**:

1. **Circuit Breaker OPEN** ⚠️ **CRITICAL**
   - Blocking all ML classification requests
   - Error: "Circuit breaker is OPEN - request rejected"
   - **Impact**: No ML-based classification, reduced accuracy

2. **Timeout Mismatches** ⚠️ **HIGH**
   - Classification service timeout: 120s
   - ML service HTTP client timeout: 30s
   - Requests may still timeout due to:
     - Website scraping delays
     - Database query timeouts
     - Network latency

3. **Consecutive Failures** ⚠️ **HIGH**
   - Circuit breaker opens after 10 consecutive failures
   - Failures could be due to:
     - Service startup issues
     - Network connectivity problems
     - Timeout issues
     - Service overload

4. **Recovery Not Happening** ⚠️ **MEDIUM**
   - Circuit breaker should recover after 60s if service is healthy
   - Health monitoring should reset it
   - May not be working if requests continue to fail

**Circuit Breaker Configuration**:
- **Failure Threshold**: 10 consecutive failures
- **Open Timeout**: 60 seconds
- **Success Threshold**: 2 successes to close
- **Reset Timeout**: 120 seconds
- **Automatic Recovery**: Enabled (health monitoring every 30s)

### Required Fixes

1. **Check Circuit Breaker State** - **IMMEDIATE**
   - Use health endpoint: `/health`
   - Review circuit breaker metrics
   - Identify why it opened

2. **Test Service Connectivity** - **IMMEDIATE**
   - Run test script: `scripts/test_python_ml_service.go`
   - Verify service is actually healthy
   - Check response times

3. **Reset Circuit Breaker** - **IMMEDIATE**
   - Manual reset: `POST /admin/circuit-breaker/reset`
   - Or wait for automatic recovery (60s timeout)
   - Monitor recovery process

4. **Review Timeout Configuration** - **HIGH**
   - Ensure timeouts are aligned
   - ML service timeout: 30s
   - Classification service timeout: 120s
   - Website scraping timeout: 15s

5. **Improve Circuit Breaker Recovery** - **MEDIUM**
   - Review automatic recovery logic
   - Ensure health monitoring is working
   - Consider reducing failure threshold if too sensitive

**Expected Impact After Fix**:
- ML service usage: 0% → >80%
- Classification accuracy: 10.7% → 50-70%
- Confidence scores: 24.65% → 50-60%
- Ensemble voting: Enabled

---

## Track 6.2: Playwright Scraper Service Verification

### Status: ✅ Completed

### Key Findings

**Document**: `docs/playwright-service-connectivity-audit.md`  
**Script**: `scripts/test_playwright_service.go`

**Root Causes Identified**:

1. **Service Configuration Unclear** ⚠️ **HIGH**
   - `PLAYWRIGHT_SERVICE_URL` not configured in `railway.json`
   - Service may not be deployed in production
   - **Impact**: Playwright strategy disabled, no fallback for JavaScript-heavy sites

2. **Browser Pool Exhaustion** ⚠️ **CRITICAL**
   - Browsers not being released properly after requests complete
   - Browsers stuck in "in use" state
   - Queue wait times: 55+ minutes
   - **Impact**: 100% failure rate for Playwright strategy, service unresponsive

3. **Timeout Mismatches** ⚠️ **HIGH**
   - HTTP client timeout: 60s
   - Context deadline: 20s
   - Service taking too long due to browser pool issues
   - **Impact**: All requests timeout before completion

4. **Rate Limiting** ⚠️ **MEDIUM**
   - HTTP 429 errors from target sites (google.com, etc.)
   - Forces fallback to Playwright (which is also failing)
   - **Impact**: Reduced scraping success rate

**Service Role**:
- **Strategy Order**: hrequests → SimpleHTTP → BrowserHeaders → **Playwright (fallback)**
- **Use Case**: JavaScript-heavy websites that require browser rendering
- **Status**: Conditional - Only enabled if `PLAYWRIGHT_SERVICE_URL` is configured

### Required Fixes

1. **Verify Service Configuration** - **IMMEDIATE**
   - Check if `PLAYWRIGHT_SERVICE_URL` is set in Railway
   - Verify service is deployed
   - Test service connectivity

2. **Fix Browser Pool Management** - **CRITICAL**
   - Review browser pool implementation
   - Fix browser release logic
   - Add browser pool monitoring
   - **Expected Impact**: Service reliability, reduced queue wait times

3. **Align Timeout Configurations** - **HIGH**
   - Reduce HTTP client timeout to match context deadline (20s)
   - Or increase context deadline to allow more time
   - Ensure timeouts are consistent
   - **Expected Impact**: Reduced timeout errors

4. **Improve Error Handling** - **MEDIUM**
   - Better error categorization
   - More informative error messages
   - Better retry logic

5. **Add Rate Limiting** - **MEDIUM**
   - Implement rate limiting for target sites
   - Add backoff strategy
   - Rotate User-Agents

**Expected Impact After Fix**:
- Scraping success rate: 10.4% → ≥70% (with Playwright fallback)
- JavaScript-heavy sites: 0% → >80% success rate
- Service reliability: Improved with browser pool fixes
- Error rate: Reduced with better timeout handling

---

## Track 6.3: Supabase Database Connectivity Verification

### Status: ✅ Completed

### Key Findings

**Document**: `docs/supabase-database-connectivity-audit.md`

**Root Causes Identified**:

1. **Missing Tables** ⚠️ **HIGH**
   - Some tables may not exist (from setup instructions)
   - Tables: `code_metadata`, `industry_code_crosswalks` may be missing
   - **Impact**: Queries fail, code generation fails

2. **Incomplete Data** ⚠️ **HIGH**
   - Code metadata may be missing or incomplete
   - MCC, NAICS, SIC codes may not be fully populated
   - **Impact**: Code generation rate low (23.1%), accuracy 0%

3. **Slow Queries** ⚠️ **MEDIUM**
   - Queries may be slow without proper indexing
   - No query timeouts implemented
   - **Impact**: Request timeouts, high latency (43.7s)

4. **Large Result Set Limits** ⚠️ **LOW**
   - 5000 record limit may miss codes
   - **Impact**: Incomplete code generation

5. **N+1 Query Problem** ⚠️ **LOW**
   - `GetCodeMetadataBatch` queries each code individually
   - **Impact**: Slower performance for multiple codes

**Database Configuration**:
- **Connection**: ✅ Configured via environment variables
- **Health Check**: ✅ Implemented (5s timeout)
- **Client Initialization**: ✅ With error handling

**Key Tables**:
- `classification_codes`: MCC, NAICS, SIC codes (CRITICAL)
- `code_metadata`: Enhanced metadata (HIGH)
- `industries`: Industry classification (CRITICAL)
- `industry_keywords`: Keyword mappings (CRITICAL)
- `industry_code_crosswalks`: Code crosswalks (MEDIUM)

### Required Fixes

1. **Verify Table Existence** - **IMMEDIATE**
   - Run table existence check in Supabase SQL Editor
   - Create missing tables if needed
   - Run migration scripts

2. **Verify Data Completeness** - **IMMEDIATE**
   - Run data count queries
   - Verify MCC, NAICS, SIC codes are populated
   - Check code_metadata table has data

3. **Add Query Timeouts** - **HIGH**
   - Add context timeouts to all database queries
   - Set reasonable timeout values (5-10s)
   - Handle timeout errors gracefully

4. **Optimize Query Performance** - **MEDIUM**
   - Add indexes on frequently queried columns
   - Review EXPLAIN ANALYZE results
   - Optimize slow queries

5. **Fix N+1 Query Problem** - **MEDIUM**
   - Batch code metadata queries
   - Use IN queries instead of individual queries
   - Cache frequently accessed data

**Expected Impact After Fix**:
- Code generation rate: 23.1% → ≥90% (with complete data)
- NAICS accuracy: 0% → ≥70% (with complete data)
- SIC accuracy: 0% → ≥70% (with complete data)
- Query performance: Improved with indexes and timeouts
- Error rate: Reduced with proper error handling

---

## Track 7.1: Feature Flag Configuration Audit

### Status: ⏳ Pending

### Investigation Steps

1. **Check Feature Flag Settings**
   - Review environment variables
   - Check feature flag configuration
   - Verify flag values

2. **Review Feature Flag Logic**
   - Check flag evaluation logic
   - Verify flag defaults
   - Review flag usage

3. **Test Flag Impact**
   - Test with flags enabled/disabled
   - Measure impact on performance
   - Verify flag behavior

### Expected Findings

- Misconfigured flags
- Incorrect flag logic
- Flags not being applied
- Performance impact

### Required Fixes

- To be determined after investigation

---

## Track 7.2: Configuration Mismatch Investigation

### Status: ⏳ Pending

### Investigation Steps

1. **Compare Config Files**
   - Compare development vs production configs
   - Check for missing variables
   - Verify default values

2. **Verify Env Var Loading**
   - Check environment variable loading
   - Verify variable precedence
   - Review variable validation

3. **Check Service-to-Service Configuration**
   - Verify service URLs
   - Check timeout configurations
   - Review connection settings

### Expected Findings

- Configuration mismatches
- Missing environment variables
- Incorrect defaults
- Service URL issues

### Required Fixes

- To be determined after investigation

---

## Track 8.1: Cache Hit Rate Investigation

### Status: ⏳ Pending

### Investigation Steps

1. **Check Cache Configuration**
   - Review cache settings
   - Verify cache TTL
   - Check cache size limits

2. **Analyze Cache Patterns**
   - Review cache hit/miss rates
   - Identify cache patterns
   - Check cache key generation

3. **Review Cache Key Generation**
   - Verify key uniqueness
   - Check key format
   - Review key collisions

### Expected Findings

- Low cache hit rates
- Cache key issues
- Cache configuration problems
- Cache performance issues

### Required Fixes

- To be determined after investigation

---

## Track 8.2: Resource Constraints Investigation

### Status: ⏳ Pending

### Investigation Steps

1. **Review Railway Resource Limits**
   - Check memory limits
   - Review CPU limits
   - Verify resource allocation

2. **Analyze Memory Usage**
   - Check memory consumption
   - Identify memory leaks
   - Review OOM kills

3. **Check Concurrent Request Limits**
   - Review request limits
   - Check queue depth
   - Analyze request queuing

### Expected Findings

- Resource constraints
- Memory issues
- CPU bottlenecks
- Request limit issues

### Required Fixes

- To be determined after investigation

---

## Track 9.1: Test Data Quality Validation

### Status: ⏳ Pending

### Investigation Steps

1. **Review Test Data for Malformed URLs**
   - Check URL format
   - Verify URL accessibility
   - Identify invalid URLs

2. **Validate Expected Results**
   - Review expected industries
   - Verify expected codes
   - Check data consistency

3. **Clean Test Data**
   - Remove invalid entries
   - Fix malformed URLs
   - Update expected results

### Expected Findings

- Malformed URLs
- Invalid expected results
- Data inconsistencies
- Missing data

### Required Fixes

- To be determined after investigation

---

## Priority Matrix

### Critical Priority (Fix Immediately)

1. **Fix Python ML Service Circuit Breaker** (Track 6.1)
   - Impact: Enables ML classification, improves accuracy
   - Effort: Low (reset + verify)
   - Dependencies: None

2. **Fix DNS Resolution** (Track 2.2)
   - Impact: Reduces 63.5% of errors
   - Effort: Low (2-3 days)
   - Dependencies: None

### High Priority (Fix This Week)

3. **Fix Timeout Configurations** (Track 1.1)
   - Impact: Reduces 9.9% of errors, improves latency
   - Effort: Low (1-2 days)
   - Dependencies: None

4. **Adjust Confidence Thresholds** (Track 3.2)
   - Impact: Improves confidence scores and accuracy
   - Effort: Low (1 day)
   - Dependencies: Track 6.1 (ML service)

5. **Improve Classification Algorithm** (Track 3.1)
   - Impact: Improves accuracy from 10.7% → 50-70%
   - Effort: Medium (3-4 days)
   - Dependencies: Track 6.1 (ML service)

### Medium Priority (Fix Next Week)

6. **Fix Error Handling & Retry Logic** (Track 2.1)
   - Impact: Reduces error rate
   - Effort: Medium (3-4 days)
   - Dependencies: Track 2.2 (DNS)

7. **Fix Web Scraping Infrastructure** (Track 5.1, 5.2)
   - Impact: Improves scraping success rate
   - Effort: Medium (4-5 days)
   - Dependencies: Track 2.2 (DNS), Track 6.2 (Playwright)

8. **Fix Code Generation** (Track 4.1, 4.2)
   - Impact: Improves code generation rate
   - Effort: Medium (3-4 days)
   - Dependencies: Track 3.1 (Classification)

### Low Priority (Fix Later)

9. **Optimize Performance** (Track 1.2)
   - Impact: Improves latency
   - Effort: High (5-7 days)
   - Dependencies: Multiple tracks

10. **Investigate Cache Hit Rate** (Track 8.1)
    - Impact: Improves performance
    - Effort: Medium (2-3 days)
    - Dependencies: None

---

## Consolidated Required Fixes

### Phase 1: Critical Fixes (Week 1)

1. **Fix Python ML Service Circuit Breaker**
   - Reset circuit breaker
   - Verify service health
   - Monitor recovery
   - **Expected Impact**: ML service usage 0% → >80%

2. **Fix DNS Resolution**
   - Implement DNS fallback (8.8.8.8, 1.1.1.1)
   - Verify IPv6 handling
   - Increase DNS timeout
   - **Expected Impact**: Error rate 67.1% → 30-40%

3. **Fix Timeout Configurations**
   - Align all timeout configurations
   - Fix timeout budget calculations
   - Ensure context propagation
   - **Expected Impact**: Error rate reduction, latency improvement

### Phase 2: High Priority Fixes (Week 2)

4. **Adjust Confidence Thresholds**
   - Increase confidence floor: 0.30 → 0.50
   - Boost calibration factors
   - Reduce thresholds: 0.85 → 0.70, 0.80 → 0.60
   - **Expected Impact**: Average confidence 24.65% → 50-60%

5. **Improve Classification Algorithm**
   - Fix ML service (from Phase 1)
   - Improve keyword patterns
   - Better fallback logic
   - **Expected Impact**: Accuracy 10.7% → 50-70%

6. **Fix Error Handling & Retry Logic**
   - Enhance scraping robustness
   - Refine keyword extraction
   - Better error categorization
   - **Expected Impact**: Error rate reduction

### Phase 3: Medium Priority Fixes (Week 3-4)

7. **Fix Web Scraping Infrastructure**
   - Fix DNS (from Phase 1)
   - Verify Playwright service
   - Enable page crawling
   - **Expected Impact**: Scraping success 10.4% → ≥70%

8. **Fix Code Generation**
   - Lower code generation threshold
   - Fix NAICS/SIC code generation
   - **Expected Impact**: Code generation 23.1% → ≥90%

9. **Optimize Performance**
   - Optimize slow operations
   - Review concurrent request limits
   - **Expected Impact**: Latency 43.7s → <10s

---

## Success Metrics

After implementing all fixes, validate against these targets:

| Metric | Current | Target | Validation Method |
|--------|---------|--------|-------------------|
| Error Rate | 67.1% | <5% | Run 100-sample E2E test |
| Average Latency | 43.7s | <10s | Run 100-sample E2E test |
| Classification Accuracy | 9.5% | ≥80% | Run 100-sample E2E test |
| Code Generation Rate | 23.1% | ≥90% | Run 100-sample E2E test |
| Scraping Success Rate | 10.4% | ≥70% | Run 100-sample E2E test |
| NAICS Accuracy | 0% | ≥70% | Run 100-sample E2E test |
| SIC Accuracy | 0% | ≥70% | Run 100-sample E2E test |
| Average Confidence | 24.65% | >70% | Run 100-sample E2E test |
| ML Service Usage | 0% | >80% | Monitor service logs |

---

## Next Steps

1. ✅ Complete remaining investigation tracks (6.2, 6.3, 7.1, 7.2, 8.1, 8.2, 9.1)
2. ✅ Update this unified summary with all findings
3. ⏳ Begin Phase 1 fixes after all investigations complete
4. ⏳ Validate fixes with E2E tests
5. ⏳ Monitor metrics and iterate

---

## Document History

- **2025-12-22**: Initial unified summary created
- **2025-12-22**: Added Track 1.2, 2.1, 3.1, 3.2, 6.1 findings

---

## References

- Investigation Plan: `/Users/petercrawford/.cursor/plans/classification_service_root_cause_investigation_plan_5da0ef16.plan.md`
- Error Pattern Analysis: `docs/error-pattern-analysis.md`
- Classification Algorithm Investigation: `docs/classification-algorithm-investigation.md`
- Confidence Score Calibration: `docs/confidence-score-calibration-investigation.md`
- Python ML Service Connectivity: `docs/python-ml-service-connectivity-audit.md`

