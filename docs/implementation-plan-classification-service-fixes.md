# Implementation Plan - Classification Service Performance Fixes

**Date**: December 22, 2025  
**Status**: Ready for Implementation  
**Objective**: Systematically implement all fixes identified in the root cause investigation

---

## Executive Summary

This plan consolidates all required fixes from the unified investigation summary into a phased, actionable implementation plan. The plan is organized by priority, dependencies, and expected impact.

**Current Performance**:

- Error Rate: 67.1% (Target: <5%)
- Average Latency: 43.7s (Target: <10s)
- Classification Accuracy: 9.5% (Target: ≥80%)
- Code Generation Rate: 23.1% (Target: ≥90%)
- Scraping Success Rate: 10.4% (Target: ≥70%)

**Expected Outcomes After All Fixes**:

- Error Rate: 67.1% → <5%
- Average Latency: 43.7s → <10s
- Classification Accuracy: 9.5% → ≥80%
- Code Generation Rate: 23.1% → ≥90%
- Scraping Success Rate: 10.4% → ≥70%

---

## Implementation Phases

### Phase 1: Critical Infrastructure Fixes (Week 1)

**Priority**: CRITICAL  
**Expected Duration**: 3-5 days  
**Dependencies**: None

### Phase 2: High Priority Algorithm & Configuration Fixes (Week 2)

**Priority**: HIGH  
**Expected Duration**: 4-6 days  
**Dependencies**: Phase 1 (ML Service)

### Phase 3: Medium Priority Optimizations (Week 3-4)

**Priority**: MEDIUM  
**Expected Duration**: 5-7 days  
**Dependencies**: Phase 1, Phase 2

### Phase 4: Low Priority Enhancements (Week 5+)

**Priority**: LOW  
**Expected Duration**: 3-5 days  
**Dependencies**: Phase 1-3

---

## Phase 1: Critical Infrastructure Fixes

### Fix 1.1: Reset Python ML Service Circuit Breaker

**Priority**: CRITICAL  
**Effort**: 1 hour  
**Impact**: Enables ML classification, improves accuracy 10.7% → 50-70%

#### Problem

- Circuit breaker is OPEN, blocking all ML classification requests
- ML service is healthy but circuit breaker hasn't recovered
- System falling back to Go keyword-based classification only

#### Solution

1. **Check Circuit Breaker State** (5 min)

   - Use health endpoint: `GET /health`
   - Check `ml_service_status.circuit_breaker_state`
   - Review circuit breaker metrics

2. **Verify Service Health** (10 min)

   - Test service directly: `https://python-ml-service-production.up.railway.app/health`
   - Run test script: `scripts/test_python_ml_service.go`
   - Verify service responds correctly

3. **Reset Circuit Breaker** (5 min)

   - **Option A**: Manual reset via endpoint
     ```bash
     curl -X POST https://classification-service-production.up.railway.app/admin/circuit-breaker/reset \
       -H "X-Admin-Key: <admin-key>"
     ```
   - **Option B**: Wait for automatic recovery (60s timeout)
   - **Option C**: Redeploy service (circuit breaker resets on restart)

4. **Monitor Recovery** (30 min)
   - Watch health endpoint for state change
   - Verify ML service usage increases
   - Check ensemble voting is working

#### Code Locations

- Health Endpoint: `services/classification-service/internal/handlers/classification.go:5075-5175`
- Reset Endpoint: `services/classification-service/internal/handlers/classification.go:4971-5031`
- Circuit Breaker: `internal/machine_learning/infrastructure/python_ml_service.go:100-107`

#### Validation

- [ ] Circuit breaker state is CLOSED
- [ ] ML service usage >80% (check logs)
- [ ] Ensemble voting enabled (check classification flow)
- [ ] Classification accuracy improved (run 50-sample test)

#### Expected Impact

- ML Service Usage: 0% → >80%
- Classification Accuracy: 10.7% → 50-70%
- Confidence Scores: 24.65% → 50-60%
- Ensemble Voting: Enabled

---

### Fix 1.2: Verify and Fix DNS Resolution

**Priority**: CRITICAL  
**Effort**: 2-3 days  
**Impact**: Reduces 63.5% of errors (DNS failures)

#### Problem

- DNS failures account for 63.5% of all errors
- IPv6 resolution issues (`[fd12::10]:53`)
- Malformed URLs reaching DNS lookup

#### Solution

1. **Verify DNS Implementation** (2 hours)

   - Review DNS fallback logic: `internal/classification/smart_website_crawler.go:163-239`
   - Verify fallback DNS servers: 8.8.8.8, 1.1.1.1, 8.8.4.4
   - Check IPv4 forcing logic

2. **Verify URL Validation** (2 hours)

   - Review hostname validation: `internal/external/website_scraper.go:1815-1840`
   - Test with malformed URLs
   - Ensure validation happens before DNS lookup

3. **Test DNS Resolution** (4 hours)

   - Run DNS resolution tests
   - Test with various URL formats
   - Verify fallback DNS servers work

4. **Monitor DNS Errors** (ongoing)
   - Track DNS error rate in logs
   - Verify errors reduced after fixes

#### Code Locations

- DNS Resolution: `internal/classification/smart_website_crawler.go:163-239`
- URL Validation: `internal/external/website_scraper.go:1815-1840`
- DNS Retry Logic: `internal/classification/smart_website_crawler.go:239-281`

#### Validation

- [ ] DNS error rate <5% (from 63.5%)
- [ ] Malformed URLs rejected before DNS lookup
- [ ] Fallback DNS servers working
- [ ] IPv4 resolution forced correctly

#### Expected Impact

- Error Rate: 67.1% → 30-40% (DNS fixes alone)
- DNS Failures: 63.5% → <5%
- Scraping Success: Improved with valid URLs

---

### Fix 1.3: Align Timeout Configurations

**Priority**: HIGH  
**Effort**: 1-2 days  
**Impact**: Reduces 9.9% of errors (timeouts), improves latency

#### Problem

- Timeout mismatches causing premature failures
- Classification service timeout: 120s
- ML service HTTP client timeout: 30s
- Website scraping timeout: 15s
- Context deadlines may be too short

#### Solution

1. **Review All Timeout Configurations** (4 hours)

   - Classification service: `services/classification-service/internal/config/config.go`
   - ML service client: `internal/machine_learning/infrastructure/python_ml_service.go`
   - Website scraper: `internal/external/website_scraper.go`
   - Database queries: `internal/classification/repository/supabase_repository.go`

2. **Create Timeout Budget** (4 hours)

   - Define timeout allocation:
     - Website scraping: 20s
     - ML service call: 30s
     - Database queries: 10s
     - Classification processing: 10s
     - Buffer: 10s
     - Total: 80s (within 120s limit)

3. **Update Timeout Configurations** (4 hours)

   - Align all timeouts with budget
   - Ensure context propagation
   - Add timeout logging

4. **Test Timeout Handling** (4 hours)
   - Test with slow services
   - Verify graceful degradation
   - Check timeout error handling

#### Code Locations

- Config: `services/classification-service/internal/config/config.go`
- ML Service: `internal/machine_learning/infrastructure/python_ml_service.go`
- Website Scraper: `internal/external/website_scraper.go`
- Database: `internal/classification/repository/supabase_repository.go`

#### Validation

- [ ] Timeout errors <5% (from 9.9%)
- [ ] All timeouts aligned with budget
- [ ] Context propagation working
- [ ] Graceful degradation on timeout

#### Expected Impact

- Error Rate: Additional 5-10% reduction
- Latency: Improved with proper timeouts
- Timeout Errors: 9.9% → <5%

---

## Phase 2: High Priority Algorithm & Configuration Fixes

### Fix 2.1: Adjust Confidence Score Thresholds and Calibration

**Priority**: HIGH  
**Effort**: 1 day  
**Impact**: Improves confidence scores 24.65% → 50-60%

#### Problem

- Confidence floor too low (0.30)
- Early termination threshold too high (0.85)
- Layer 2 threshold too high (0.80)
- Calibration factors not effective enough

#### Solution

1. **Increase Confidence Floor** (30 min)

   - **File**: `internal/classification/confidence_calibrator.go:416`
   - **Change**: `0.30` → `0.50`

   ```go
   calibratedConfidence = math.Max(calibratedConfidence, 0.50)
   ```

2. **Boost Calibration Factors** (1 hour)

   - **File**: `internal/classification/confidence_calibrator.go:359-388`
   - **Changes**:
     - Content quality boost: `1.10` → `1.20` (+10% → +20%)
     - Strategy agreement boost: `1.15` → `1.25` (+15% → +25%)

   ```go
   // Factor 1: Content quality boost
   if contentQuality > 0.8 {
       calibratedConfidence *= 1.20 // Increased from 1.10
   }

   // Factor 2: Strategy agreement
   if strategyVariance < 0.05 {
       calibratedConfidence *= 1.25 // Increased from 1.15
   }
   ```

3. **Reduce Early Termination Threshold** (30 min)

   - **File**: `services/classification-service/internal/config/config.go:136`
   - **Change**: `0.85` → `0.70`

   ```go
   EarlyTerminationConfidenceThreshold: getEnvAsFloat("EARLY_TERMINATION_CONFIDENCE_THRESHOLD", 0.70),
   ```

4. **Reduce Layer 2 Threshold** (30 min)

   - **File**: `internal/classification/service.go:422`
   - **Change**: `0.80` → `0.60`

   ```go
   const layer2Threshold = 0.60
   ```

5. **Update Railway Configuration** (30 min)
   - Set `EARLY_TERMINATION_CONFIDENCE_THRESHOLD=0.70` in Railway
   - Redeploy service

#### Code Locations

- Confidence Calibrator: `internal/classification/confidence_calibrator.go:344-419`
- Config: `services/classification-service/internal/config/config.go:136`
- Service: `internal/classification/service.go:422`

#### Validation

- [ ] Average confidence >50% (from 24.65%)
- [ ] High confidence results >30% (from <5%)
- [ ] Early termination working with lower threshold
- [ ] Layer 2 usage increased

#### Expected Impact

- Average Confidence: 24.65% → 50-60%
- High Confidence Results: <5% → 30-40%
- Low Confidence Results: ~70% → 30-40%

---

### Fix 2.2: Improve Classification Algorithm

**Priority**: HIGH  
**Effort**: 3-4 days  
**Impact**: Improves accuracy 10.7% → 50-70%

#### Problem

- Many industries have 0% accuracy
- Keyword patterns may not match test data
- Defaulting to "General Business" too often
- ML service unavailable (from Phase 1)

#### Solution

1. **Fix ML Service** (from Fix 1.1)

   - Ensure circuit breaker is closed
   - Verify ensemble voting enabled

2. **Review Keyword Patterns** (1 day)

   - **File**: `internal/classification/keyword_matcher.go`
   - Review industries with 0% accuracy:
     - arts & entertainment
     - construction
     - energy
     - food & beverage
     - professional services
     - transportation
   - Add missing keywords
   - Test against test data

3. **Improve Fallback Logic** (1 day)

   - **File**: `internal/classification/service.go:1204-1209`
   - Instead of defaulting to "General Business":
     - Try fuzzy matching
     - Use confidence-based selection
     - Return top N industries instead of single default

   ```go
   // Improved fallback logic
   if classification == nil {
       // Try fuzzy matching instead of default
       fuzzyMatch := tryFuzzyMatching(businessData)
       if fuzzyMatch != nil && fuzzyMatch.Confidence > 0.40 {
           return fuzzyMatch, nil
       }
       // Only default if no match found
       return &IndustryDetectionResult{
           IndustryName: "General Business",
           Confidence:   0.30,
           Keywords:     keywords,
           Reasoning:    "No matching industry found after fuzzy matching",
       }, nil
   }
   ```

4. **Review Content Quality Requirements** (4 hours)

   - **File**: `services/classification-service/internal/config/config.go:137`
   - Consider reducing `MIN_CONTENT_LENGTH_FOR_ML` from 50 to 30
   - Test impact on ML usage

5. **Test Classification Accuracy** (1 day)
   - Run 100-sample E2E test
   - Analyze accuracy by industry
   - Identify remaining issues

#### Code Locations

- Keyword Matcher: `internal/classification/keyword_matcher.go`
- Service: `internal/classification/service.go:341-487`
- Fallback Logic: `internal/classification/service.go:1204-1209`
- Config: `services/classification-service/internal/config/config.go:137`

#### Validation

- [ ] Classification accuracy >50% (from 10.7%)
- [ ] Industries with 0% accuracy improved
- [ ] "General Business" defaults <10% (from high %)
- [ ] ML service usage >80%

#### Expected Impact

- Classification Accuracy: 10.7% → 50-70%
- Industry Accuracy: Improved across all industries
- ML Service Usage: >80% (from Phase 1)

---

### Fix 2.3: Verify and Fix Database Connectivity

**Priority**: HIGH  
**Effort**: 1-2 days  
**Impact**: Improves code generation 23.1% → ≥90%

#### Problem

- Missing tables may exist
- Incomplete data (MCC, NAICS, SIC codes)
- No query timeouts
- Slow queries

#### Solution

1. **Verify Table Existence** (2 hours)

   - Run in Supabase SQL Editor:

   ```sql
   SELECT table_name
   FROM information_schema.tables
   WHERE table_schema = 'public'
   AND table_name IN (
     'classification_codes',
     'code_metadata',
     'industries',
     'industry_keywords',
     'industry_code_crosswalks'
   );
   ```

   - Create missing tables if needed
   - Run migration scripts

2. **Verify Data Completeness** (2 hours)

   - Run count queries:

   ```sql
   SELECT code_type, COUNT(*) as count
   FROM classification_codes
   WHERE is_active = true
   GROUP BY code_type;
   ```

   - Verify MCC, NAICS, SIC codes are populated
   - Check code_metadata table has data

3. **Add Query Timeouts** (4 hours)

   - **File**: `internal/classification/repository/supabase_repository.go`
   - Add context timeouts to all database queries
   - Set reasonable timeout values (5-10s)

   ```go
   ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
   defer cancel()
   // Use ctx in query
   ```

4. **Optimize Query Performance** (4 hours)

   - Add indexes on frequently queried columns:
     - `classification_codes(industry_id, is_active)`
     - `classification_codes(code_type, is_active)`
     - `code_metadata(code_type, code, is_active)`
   - Review EXPLAIN ANALYZE results
   - Optimize slow queries

5. **Fix N+1 Query Problem** (4 hours)
   - **File**: `internal/classification/repository/code_metadata_repository.go`
   - Batch code metadata queries
   - Use IN queries instead of individual queries

#### Code Locations

- Repository: `internal/classification/repository/supabase_repository.go`
- Code Metadata: `internal/classification/repository/code_metadata_repository.go`
- Client: `services/classification-service/internal/supabase/client.go`

#### Validation

- [ ] All required tables exist
- [ ] Data counts are reasonable (MCC: ~500-1000, NAICS: ~1000-2000, SIC: ~1000)
- [ ] Query timeouts implemented
- [ ] Query performance improved (EXPLAIN ANALYZE)

#### Expected Impact

- Code Generation Rate: 23.1% → ≥90%
- NAICS Accuracy: 0% → ≥70%
- SIC Accuracy: 0% → ≥70%
- Query Performance: Improved with indexes and timeouts

---

## Phase 3: Medium Priority Optimizations

### Fix 3.1: Fix Web Scraping Infrastructure

**Priority**: MEDIUM  
**Effort**: 4-5 days  
**Impact**: Improves scraping success 10.4% → ≥70%

#### Problem

- DNS failures (from Phase 1)
- Playwright service browser pool exhaustion
- Timeout mismatches
- Rate limiting (HTTP 429)

#### Solution

1. **Fix DNS** (from Fix 1.2)

   - DNS resolution fixes from Phase 1

2. **Verify Playwright Service Configuration** (2 hours)

   - Check if `PLAYWRIGHT_SERVICE_URL` is set in Railway
   - Verify service is deployed
   - Test service connectivity

3. **Fix Browser Pool Management** (2 days)

   - **File**: `services/playwright-scraper/index.js`
   - Review browser pool implementation
   - Fix browser release logic
   - Add browser pool monitoring
   - Ensure browsers are released after requests

4. **Align Timeout Configurations** (4 hours)

   - **File**: `internal/external/website_scraper.go:143`
   - Reduce HTTP client timeout to match context deadline (20s)
   - Or increase context deadline to allow more time
   - Ensure timeouts are consistent

5. **Improve Error Handling** (4 hours)

   - Better error categorization
   - More informative error messages
   - Better retry logic

6. **Add Rate Limiting** (1 day)
   - Implement rate limiting for target sites
   - Add backoff strategy
   - Rotate User-Agents

#### Code Locations

- Website Scraper: `internal/external/website_scraper.go`
- Playwright Service: `services/playwright-scraper/index.js`

#### Validation

- [ ] Scraping success rate >70% (from 10.4%)
- [ ] JavaScript-heavy sites >80% success
- [ ] Browser pool not exhausted
- [ ] Timeout errors reduced

#### Expected Impact

- Scraping Success Rate: 10.4% → ≥70%
- JavaScript-Heavy Sites: 0% → >80% success rate
- Service Reliability: Improved with browser pool fixes

---

### Fix 3.2: Optimize Cache Hit Rate

**Priority**: MEDIUM  
**Effort**: 2-3 days  
**Impact**: Improves performance for cached requests

#### Problem

- Redis may not be enabled
- Cache key too specific
- URL normalization missing
- Cache TTL may be too short

#### Solution

1. **Verify Cache Configuration** (1 hour)

   - Check `CACHE_ENABLED` is `true` in Railway
   - Check `CACHE_TTL` is set appropriately
   - Verify Redis is enabled if available

2. **Analyze Cache Hit Rate** (2 hours)

   - Parse Railway logs for cache hit/miss patterns
   - Calculate actual hit rate
   - Identify patterns in cache misses

3. **Improve URL Normalization** (4 hours)

   - **File**: `services/classification-service/internal/handlers/classification.go:575-600`
   - Normalize URLs in cache key generation:
     - Remove `www.`
     - Remove trailing slashes
     - Remove query parameters

   ```go
   // Normalize URL
   websiteURL = normalizeURL(websiteURL)

   func normalizeURL(url string) string {
       u, err := url.Parse(url)
       if err != nil {
           return url
       }
       // Remove www
       if strings.HasPrefix(u.Host, "www.") {
           u.Host = u.Host[4:]
       }
       // Remove trailing slash
       u.Path = strings.TrimSuffix(u.Path, "/")
       // Remove query and fragment
       u.RawQuery = ""
       u.Fragment = ""
       return u.String()
   }
   ```

4. **Enable Redis Cache** (2 hours)

   - Set `REDIS_ENABLED=true` if Redis is available
   - Set `REDIS_URL` correctly
   - Verify Redis is being used

5. **Optimize Cache Key Generation** (4 hours)
   - Consider fuzzy matching for descriptions
   - Normalize business names (remove common suffixes)
   - Test cache hit rate improvement

#### Code Locations

- Cache Implementation: `services/classification-service/internal/handlers/classification.go:740-818`
- Cache Key Generation: `services/classification-service/internal/handlers/classification.go:575-600`
- Redis Cache: `services/classification-service/internal/cache/redis_cache.go`

#### Validation

- [ ] Cache hit rate >60% (target)
- [ ] Redis enabled and working
- [ ] URL normalization working
- [ ] Response time improved for cached requests

#### Expected Impact

- Cache Hit Rate: Current → >60% (target)
- Response Time: Improved for cached requests
- Database Load: Reduced with higher cache hit rate

---

### Fix 3.3: Optimize Performance

**Priority**: MEDIUM  
**Effort**: 5-7 days  
**Impact**: Improves latency 43.7s → <10s

#### Problem

- Slow operations identified
- Concurrent request limits may be too restrictive
- Database queries may be slow
- Website scraping contributing to latency

#### Solution

1. **Identify Slow Operations** (1 day)

   - Review detailed request tracing
   - Identify slow stages in request processing
   - Profile code to find bottlenecks

2. **Optimize Slow Stages** (2 days)

   - Optimize website scraping (from Fix 3.1)
   - Optimize database queries (from Fix 2.3)
   - Optimize classification processing
   - Add parallel processing where possible

3. **Review Concurrent Request Limits** (1 day)

   - Test with different limits
   - Monitor queue depth
   - Adjust if needed (currently 20)

4. **Optimize Request Processing** (1 day)

   - Reduce redundant operations
   - Cache frequently accessed data
   - Stream large responses

5. **Add Performance Monitoring** (1 day)
   - Track request latency by stage
   - Monitor slow requests
   - Alert on performance degradation

#### Code Locations

- Request Tracing: `services/classification-service/internal/handlers/classification.go`
- Performance Monitoring: Add new monitoring code

#### Validation

- [ ] Average latency <10s (from 43.7s)
- [ ] P95 latency <15s
- [ ] Slow operations optimized
- [ ] Concurrent request limits appropriate

#### Expected Impact

- Average Latency: 43.7s → <10s
- P95 Latency: Improved
- Request Throughput: Improved

---

## Phase 4: Low Priority Enhancements

### Fix 4.1: Fix Test Data Quality

**Priority**: LOW  
**Effort**: 1-2 days  
**Impact**: Improves test reliability

#### Problem

- Malformed URLs in test data
- Missing expected results
- Invalid code formats

#### Solution

1. **Run Validation Script** (1 hour)

   - Run `scripts/validate_test_data_quality.go`
   - Review generated report

2. **Fix Malformed URLs** (4 hours)

   - Clean URLs with invalid characters
   - Fix missing schemes
   - Normalize URLs

3. **Add Missing Expected Results** (4 hours)

   - Add expected industries where missing
   - Add expected codes where missing

4. **Fix Code Formats** (2 hours)
   - Validate and fix MCC codes (4 digits)
   - Validate and fix NAICS codes (5-6 digits)
   - Validate and fix SIC codes (4 digits)

#### Code Locations

- Test Data: `test/data/comprehensive_test_samples.json`
- Validation Script: `scripts/validate_test_data_quality.go`

#### Validation

- [ ] All URLs valid
- [ ] All expected results present
- [ ] All code formats valid
- [ ] Test accuracy improved

---

### Fix 4.2: Resource Constraints Optimization

**Priority**: LOW  
**Effort**: 2-3 days  
**Impact**: Prevents OOM kills, improves stability

#### Solution

1. **Set Memory Limit** (1 hour)

   - Set `GOMEMLIMIT` in Railway
   - Use appropriate value (e.g., 512MB, 1GB)

2. **Monitor Memory Usage** (1 day)

   - Add memory usage metrics
   - Track memory over time
   - Alert on high memory usage

3. **Check Railway Logs for OOM** (2 hours)

   - Review logs for OOM kills
   - Identify patterns
   - Document findings

4. **Fix Memory Leaks** (1 day)
   - Review goroutine usage
   - Ensure connections are closed
   - Fix any leaks found

#### Code Locations

- Memory Limit: `services/classification-service/cmd/main.go:73`
- Concurrent Requests: `services/classification-service/internal/config/config.go:105`

#### Validation

- [ ] Memory limit set
- [ ] No OOM kills
- [ ] Memory usage stable
- [ ] No memory leaks

---

## Implementation Checklist

### Phase 1: Critical Infrastructure (Week 1)

- [ ] Fix 1.1: Reset Python ML Service Circuit Breaker
- [ ] Fix 1.2: Verify and Fix DNS Resolution
- [ ] Fix 1.3: Align Timeout Configurations
- [ ] Phase 1 Validation: Run 50-sample E2E test

### Phase 2: High Priority (Week 2)

- [ ] Fix 2.1: Adjust Confidence Score Thresholds
- [ ] Fix 2.2: Improve Classification Algorithm
- [ ] Fix 2.3: Verify and Fix Database Connectivity
- [ ] Phase 2 Validation: Run 100-sample E2E test

### Phase 3: Medium Priority (Week 3-4)

- [ ] Fix 3.1: Fix Web Scraping Infrastructure
- [ ] Fix 3.2: Optimize Cache Hit Rate
- [ ] Fix 3.3: Optimize Performance
- [ ] Phase 3 Validation: Run 100-sample E2E test

### Phase 4: Low Priority (Week 5+)

- [ ] Fix 4.1: Fix Test Data Quality
- [ ] Fix 4.2: Resource Constraints Optimization
- [ ] Phase 4 Validation: Run 100-sample E2E test

---

## Success Metrics

After implementing all fixes, validate against these targets:

| Metric                  | Current | Target | Validation Method       |
| ----------------------- | ------- | ------ | ----------------------- |
| Error Rate              | 67.1%   | <5%    | Run 100-sample E2E test |
| Average Latency         | 43.7s   | <10s   | Run 100-sample E2E test |
| Classification Accuracy | 9.5%    | ≥80%   | Run 100-sample E2E test |
| Code Generation Rate    | 23.1%   | ≥90%   | Run 100-sample E2E test |
| Scraping Success Rate   | 10.4%   | ≥70%   | Run 100-sample E2E test |
| NAICS Accuracy          | 0%      | ≥70%   | Run 100-sample E2E test |
| SIC Accuracy            | 0%      | ≥70%   | Run 100-sample E2E test |
| Average Confidence      | 24.65%  | >70%   | Run 100-sample E2E test |
| ML Service Usage        | 0%      | >80%   | Monitor service logs    |

---

## Risk Mitigation

### Risks

1. **Breaking Changes**: Fixes may introduce regressions
2. **Service Downtime**: Some fixes require redeployment
3. **Data Loss**: Database fixes may affect existing data
4. **Performance Regression**: Optimizations may have unintended effects

### Mitigation Strategies

1. **Test Thoroughly**: Run E2E tests after each phase
2. **Deploy Incrementally**: Deploy fixes in phases, not all at once
3. **Monitor Closely**: Watch metrics and logs after each deployment
4. **Rollback Plan**: Have rollback procedures ready
5. **Backup Data**: Backup database before schema changes

---

## Dependencies

### External Dependencies

- Python ML Service: Must be healthy and accessible
- Supabase Database: Must be accessible and have complete data
- Playwright Service: Must be deployed and configured (optional)
- Redis: Must be available for distributed caching (optional)

### Internal Dependencies

- Phase 2 depends on Phase 1 (ML service must be working)
- Phase 3 depends on Phase 1 (DNS fixes)
- Phase 3 depends on Phase 2 (database fixes)

---

## Timeline

**Week 1**: Phase 1 (Critical Infrastructure)

- Days 1-2: Fix 1.1, Fix 1.2
- Days 3-5: Fix 1.3, Validation

**Week 2**: Phase 2 (High Priority)

- Days 1-2: Fix 2.1, Fix 2.2
- Days 3-4: Fix 2.3
- Day 5: Validation

**Week 3-4**: Phase 3 (Medium Priority)

- Week 3: Fix 3.1, Fix 3.2
- Week 4: Fix 3.3, Validation

**Week 5+**: Phase 4 (Low Priority)

- Fix 4.1, Fix 4.2, Final Validation

---

## Next Steps

1. **Review and Approve Plan**: Review this plan with team
2. **Set Up Tracking**: Create tickets/tasks for each fix
3. **Begin Phase 1**: Start with Fix 1.1 (Circuit Breaker Reset)
4. **Monitor Progress**: Track implementation progress
5. **Validate After Each Phase**: Run E2E tests after each phase
6. **Iterate**: Adjust plan based on findings

---

## References

- Unified Investigation Summary: `docs/unified-investigation-summary.md`
- Error Pattern Analysis: `docs/error-pattern-analysis.md`
- Classification Algorithm Investigation: `docs/classification-algorithm-investigation.md`
- Confidence Score Calibration: `docs/confidence-score-calibration-investigation.md`
- Python ML Service Connectivity: `docs/python-ml-service-connectivity-audit.md`
- Playwright Service Connectivity: `docs/playwright-service-connectivity-audit.md`
- Supabase Database Connectivity: `docs/supabase-database-connectivity-audit.md`
- Feature Flag Audit: `docs/feature-flag-audit.md`
- Cache Hit Rate Investigation: `docs/cache-hit-rate-investigation.md`
- Resource Constraints Investigation: `docs/resource-constraints-investigation.md`
- Test Data Quality Investigation: `docs/test-data-quality-investigation.md`

---

**Document Version**: 1.0.0  
**Last Updated**: December 22, 2025  
**Next Review**: After Phase 1 completion
