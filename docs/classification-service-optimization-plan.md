# Classification Service Optimization Plan

## Overview

This plan addresses 20 critical optimization opportunities identified through flow analysis to improve classification service accuracy, efficiency, and speed. Expected improvements: 60-70% speed reduction, 10-15% accuracy increase, 40-60% efficiency improvement.

## Phase 1: Quick Wins (Week 1) - High Impact, Low Effort

### 1. Fix Keyword Extraction Accuracy (CRITICAL)

**Problem**: `isValidEnglishWord` function allows gibberish like "ivdi", "fays", "yilp", "dioy", "ukxa" to pass validation.

**Files to modify**:

- `internal/classification/smart_website_crawler.go:1577-1635` - Enhance `isValidEnglishWord`
- `internal/classification/smart_website_crawler.go:1527-1575` - Improve `extractWordsFromText`
- `internal/classification/repository/supabase_repository.go:3895-4030` - Add post-processing filter

**Implementation**:

1. Add common English word dictionary (10k most common words) as lookup map
2. Enhance `isValidEnglishWord` with:
   - Dictionary lookup check
   - Suspicious pattern detection (repeated letters, unusual sequences)
   - N-gram frequency validation (check bigram/trigram patterns)
   - Common letter combination validation
3. Add post-processing filter to remove isolated gibberish words
4. Improve HTML cleaning to prevent encoded character issues

**Expected Impact**: 60-80% reduction in gibberish keywords, +20-30% accuracy improvement

### 2. Request Deduplication with In-Flight Tracking

**Problem**: Identical concurrent requests processed multiple times.

**Files to modify**:

- `services/classification-service/internal/handlers/classification.go:234-292` - Add deduplication logic

**Implementation**:

1. Add in-flight request tracking map with mutex
2. Check if identical request (same cache key) is already processing
3. If in-flight: wait for completion and return same result
4. Clean up completed requests from tracking map

**Expected Impact**: 50-80% faster for duplicate requests

### 3. Content Quality Validation Before ML

**Problem**: Python ML service called with < 20 characters of content.

**Files to modify**:

- `services/classification-service/internal/handlers/classification.go:1003-1066` - Add content validation
- `internal/classification/methods/ml_method.go:220-322` - Skip ML if content insufficient

**Implementation**:

1. Validate content length before calling Python ML service
2. Skip ML if content < threshold (e.g., 50 chars)
3. Use description/business_name only if website content insufficient
4. Log content quality metrics

**Expected Impact**: 15-20% accuracy improvement, 30-40% faster for low-content sites

### 4. Enhanced Connection Pooling

**Problem**: MaxIdleConns: 10, not fully utilized, no HTTP/2.

**Files to modify**:

- `internal/classification/smart_website_crawler.go:257-278` - Improve HTTP client config

**Implementation**:

1. Increase MaxIdleConns to 100
2. Enable HTTP/2 support
3. Increase keep-alive timeout
4. Add connection pool metrics

**Expected Impact**: 20-30% faster HTTP requests

### 5. DNS Resolution Caching

**Problem**: DNS resolved for every page, even same domain.

**Files to modify**:

- `internal/classification/smart_website_crawler.go:197-232` - Add DNS cache

**Implementation**:

1. Add DNS result cache with TTL (5 minutes)
2. Cache by domain name
3. Invalidate on DNS errors
4. Thread-safe cache with mutex

**Expected Impact**: 15-25% faster page analysis

### 6. Early Termination for Low Confidence

**Problem**: Full processing continues even when confidence < 0.3.

**Files to modify**:

- `services/classification-service/internal/handlers/classification.go:1068-1101` - Add early termination
- `internal/classification/smart_website_crawler.go:675-704` - Stop crawling when confidence sufficient

**Implementation**:

1. Check confidence after initial steps
2. Terminate early if confidence < threshold and keywords < 2
3. Stop crawling when confidence >= 0.85 after 3+ pages
4. Return partial results with low confidence flag

**Expected Impact**: 50-70% faster for low-confidence cases, +5-10% accuracy

## Phase 2: Strategic Improvements (Weeks 2-3) - High Impact, Medium Effort

### 7. Parallel Processing of Independent Steps

**Problem**: Industry detection, code generation, risk assessment run sequentially.

**Files to modify**:

- `services/classification-service/internal/handlers/classification.go:372-477` - Parallelize independent steps

**Implementation**:

1. Use `sync.WaitGroup` to parallelize:
   - Industry detection
   - Code generation
   - Risk assessment
   - Website analysis (if applicable)
2. Collect results and combine
3. Add timeout per parallel operation

**Expected Impact**: 40-60% faster overall processing

### 8. Ensemble Voting (Combine Python ML + Go Results)

**Problem**: Python ML and Go classification paths are mutually exclusive.

**Files to modify**:

- `services/classification-service/internal/handlers/classification.go:1003-1066` - Run both in parallel
- `services/classification-service/internal/handlers/classification.go:1382-1456` - Combine results

**Implementation**:

1. Run Python ML and Go classification in parallel
2. Combine results with weighted voting:
   - Python ML: 60% weight
   - Go classification: 40% weight
3. Use consensus for confidence boost
4. Merge keywords and codes

**Expected Impact**: +10-15% accuracy improvement through consensus

### 9. Distributed Caching (Redis)

**Problem**: In-memory cache only works for single instance.

**Files to modify**:

- `services/classification-service/internal/handlers/classification.go:68-127` - Add Redis cache
- Create new file: `services/classification-service/internal/cache/redis_cache.go`

**Implementation**:

1. Add Redis client dependency
2. Implement Redis cache adapter
3. Fallback to in-memory if Redis unavailable
4. Add cache metrics and monitoring

**Expected Impact**: 60-80% better cache hit rate in multi-instance deployments

### 10. Circuit Breaker for External Services

**Problem**: Python ML failures cause 30s timeout on every request.

**Files to modify**:

- `internal/machine_learning/infrastructure/python_ml_service.go:328-416` - Add circuit breaker
- Create new file: `internal/machine_learning/infrastructure/circuit_breaker.go`

**Implementation**:

1. Implement circuit breaker pattern:
   - Open after N consecutive failures
   - Half-open after timeout
   - Close on success
2. Fail fast when circuit open
3. Track failure rates and latency

**Expected Impact**: 95% reduction in timeout overhead when service is down

### 11. Adaptive Retry Strategy

**Problem**: Fixed 3 retries for all errors.

**Files to modify**:

- `internal/external/website_scraper.go:138-169` - Improve retry logic
- `internal/classification/smart_website_crawler.go:793-822` - Adaptive retries

**Implementation**:

1. Don't retry permanent errors (400, 403, 404)
2. Check error history success rates
3. Adjust retry count based on error type
4. Exponential backoff with jitter

**Expected Impact**: 30-40% fewer wasted retry attempts

### 12. Content Extraction Caching Per Request

**Problem**: Website scraped multiple times in same request (ML method + crawler).

**Files to modify**:

- `internal/classification/methods/ml_method.go:483-534` - Add request-scoped cache
- `internal/classification/enhanced_website_analyzer.go:66-113` - Share cache

**Implementation**:

1. Add request-scoped content cache (map in context)
2. Check cache before scraping
3. Store scraped content in cache
4. Share between ML method and crawler

**Expected Impact**: 30-50% faster for multi-path requests

### 19. Robots.txt Crawl Delay Enforcement

**Problem**: Crawl delay extracted from robots.txt but not consistently enforced during page analysis.

**Files to modify**:

- `internal/classification/smart_website_crawler.go:675-704` - Enforce crawl delay in analyzePages
- `internal/classification/smart_website_crawler.go:706-909` - Pass crawl delay to analyzePage
- `internal/classification/smart_website_crawler.go:937-1028` - Store crawl delay from checkRobotsTxt

**Implementation**:

1. Store crawl delay from robots.txt check (already extracted, needs storage)
2. Enforce delay between page requests in analyzePages
3. Use maximum of configured delay and robots.txt delay
4. Log when robots.txt delay is being respected
5. Apply delay after each page analysis, before next request

**Expected Impact**: Better compliance, reduced 403 errors, improved reputation with websites

### 20. Adaptive Delays Based on Response Codes

**Problem**: No delays between requests increases bot detection risk and 403/429 errors.

**Files to modify**:

- `internal/classification/smart_website_crawler.go:675-704` - Add adaptive delays
- `internal/classification/smart_website_crawler.go:706-909` - Track response codes per domain

**Implementation**:

1. Implement adaptive delay strategy:
   - 200 OK: Minimal delay (1-2s) or robots.txt delay if greater
   - 429 Rate Limited: Exponential backoff (5s, 10s, 20s)
   - 503 Service Unavailable: Moderate delay (3-5s)
   - Normal: Respect robots.txt delay or default 2s minimum
2. Track response code patterns per domain
3. Adjust delays based on domain behavior history
4. Reset delay strategy after successful requests

**Expected Impact**: 30-50% reduction in 403/429 errors, better bot evasion, improved success rate

## Phase 3: Advanced Optimizations (Weeks 4-6) - Medium Impact, High Effort

### 13. Keyword Extraction Consolidation

**Problem**: Keywords extracted 4+ times (crawler, relevance, structured data, code gen).

**Files to modify**:

- Create new file: `internal/classification/context.go` - Shared classification context
- `services/classification-service/internal/handlers/classification.go:372-477` - Use shared context
- Multiple files - Pass context instead of re-extracting

**Implementation**:

1. Create `ClassificationContext` struct with extracted keywords
2. Extract keywords once at start
3. Pass context to all steps
4. Reuse keywords throughout pipeline

**Expected Impact**: 40-60% less CPU for keyword processing

### 14. Lazy Loading of Code Generation

**Problem**: Codes generated even when not needed (low confidence).

**Files to modify**:

- `services/classification-service/internal/handlers/classification.go:1109-1131` - Add conditional generation

**Implementation**:

1. Only generate codes if confidence > 0.5 or explicitly requested
2. Skip code generation for low-confidence results
3. Return empty codes with flag indicating skipped

**Expected Impact**: 20-30% faster for low-confidence requests

### 15. Structured Data Priority Weighting

**Problem**: Structured data not weighted higher than scraped text.

**Files to modify**:

- `internal/classification/smart_website_crawler.go:2050-2177` - Increase structured data weights
- `internal/classification/smart_website_crawler.go:1749-1817` - Adjust scoring

**Implementation**:

1. Weight JSON-LD/microdata keywords 2x higher
2. Prioritize structured data in keyword ranking
3. Boost structured data keywords in final scores

**Expected Impact**: +8-12% accuracy when structured data present

### 16. Industry-Specific Confidence Thresholds

**Problem**: Single threshold (0.3) for all industries.

**Files to modify**:

- `services/classification-service/internal/handlers/classification.go:1068-1101` - Add industry-specific thresholds
- Create config file: `configs/industry_thresholds.yaml`

**Implementation**:

1. Define industry-specific thresholds:
   - Financial: 0.7
   - Healthcare: 0.65
   - Legal: 0.6
   - Default: 0.3
2. Apply thresholds in classification logic
3. Adjust confidence scores accordingly

**Expected Impact**: +12-18% accuracy for high-risk industries

### 17. Streaming Responses for Long Operations

**Problem**: Full response built before sending (10-30s wait).

**Files to modify**:

- `services/classification-service/internal/handlers/classification.go:234-363` - Add streaming support

**Implementation**:

1. Use NDJSON (newline-delimited JSON) for streaming
2. Send partial results as steps complete:
   - Industry detected
   - Codes generated
   - Risk assessed
3. Final message indicates completion

**Expected Impact**: 50-70% better perceived latency

### 18. Adaptive Page Limits in Smart Crawling

**Problem**: Always crawls 20 pages even when first 3-5 provide sufficient confidence.

**Files to modify**:

- `internal/classification/smart_website_crawler.go:675-704` - Add adaptive limits

**Implementation**:

1. Check confidence after each page
2. Stop when confidence >= 0.85 after 3+ pages
3. Continue if confidence improving
4. Max pages still 20 as hard limit

**Expected Impact**: 50-70% faster for high-confidence sites

## Implementation Priority Matrix

### Critical (Do First)

1. Keyword extraction accuracy fix (#1)
2. Request deduplication (#2)
3. Content quality validation (#3)

### High Priority (Week 1-2)

4. Connection pooling (#4)
5. DNS caching (#5)
6. Early termination (#6)
7. Parallel processing (#7)
8. Robots.txt crawl delay enforcement (#19)
9. Adaptive delays based on response codes (#20)

### Medium Priority (Week 2-3)

8. Ensemble voting (#8)
9. Distributed caching (#9)
10. Circuit breaker (#10)
11. Adaptive retries (#11)
12. Content extraction caching (#12)

### Lower Priority (Week 4-6)

13. Keyword consolidation (#13)
14. Lazy code generation (#14)
15. Structured data weighting (#15)
16. Industry thresholds (#16)
17. Streaming responses (#17)
18. Adaptive page limits (#18)

## Testing Strategy

### Unit Tests

- Test keyword validation with gibberish words (comprehensive test suite with known bad words)
- Test request deduplication with concurrent requests (race conditions, cleanup)
- Test early termination logic with various confidence thresholds
- Test parallel processing coordination (WaitGroup behavior, timeout handling)
- Test robots.txt crawl delay enforcement (delay calculation, enforcement logic)
- Test adaptive delay logic with various response codes (200, 429, 503)
- Test circuit breaker state transitions (open, half-open, closed)
- Test cache hit/miss scenarios (in-memory and Redis)
- Test DNS caching with TTL expiration
- Test keyword extraction accuracy with benchmark dataset

### Integration Tests

- Test end-to-end classification with all optimizations enabled
- Test cache behavior (in-memory and Redis) with cache invalidation
- Test circuit breaker behavior with service failures
- Test fallback mechanisms (Python ML → Go classification)
- Test ensemble voting with both services (weight calculation, consensus)
- Test content extraction caching across multiple paths
- Test DNS caching with TTL expiration and invalidation
- Test robots.txt compliance with various robots.txt configurations
- Test adaptive delays with real website responses

### Performance Tests

- Benchmark before/after for each optimization individually
- Load testing with concurrent requests (100, 500, 1000 concurrent)
- Measure cache hit rates under load
- Monitor resource usage (CPU, memory, network, database connections)
- Test with various website types (fast, slow, blocked, rate-limited)
- Measure latency percentiles (p50, p95, p99)
- Test circuit breaker impact on failed services (latency reduction)
- Test parallel processing speedup vs sequential
- Test request deduplication effectiveness under load

### Accuracy Tests

- Test keyword extraction accuracy with known good/bad keywords
- Test classification accuracy with benchmark dataset (before/after optimizations)
- Test industry-specific threshold accuracy improvements
- Test ensemble voting accuracy vs single method
- Test structured data weighting impact on accuracy
- Test keyword validation filter effectiveness (gibberish removal rate)

### Security Tests

- Test input sanitization with XSS/SQL injection attempts
- Test rate limiting effectiveness (per IP, per client)
- Test request deduplication with malicious requests
- Test robots.txt compliance enforcement (legal compliance)
- Test timeout enforcement at all levels
- Test request size limits (DoS prevention)

## Success Metrics

### Speed Improvements

- Target: 60-70% reduction in average processing time
- Current: 2-10s average
- Target: 0.5-3s average

### Accuracy Improvements

- Target: 10-15% increase in classification accuracy
- Current: ~75-80% accuracy
- Target: ~85-92% accuracy

### Efficiency Improvements

- Target: 40-60% reduction in CPU usage
- Target: 50-70% fewer database queries
- Target: 60-80% better cache hit rate

## Risk Mitigation

### Backward Compatibility

- All optimizations maintain existing API contracts
- Add feature flags for gradual rollout
- Monitor error rates during deployment

### Rollback Plan

- Each optimization can be disabled via config
- Keep old code paths until validated
- Gradual rollout per optimization

### Monitoring

- Add metrics for each optimization
- Alert on performance regressions
- Track accuracy improvements
- Monitor resource usage

## Dependencies

### New Dependencies

- Redis client library: `github.com/redis/go-redis/v9` (v9.0.0+)
- Circuit breaker library: `github.com/sony/gobreaker` (v0.5.0+) or implement custom
- Word frequency database: Embed common word list (10k most common English words) or use `github.com/dwyl/english-words`
- Optional: Spell-check library for advanced validation: `github.com/kljensen/snowball` (for stemming)

### Configuration Changes

- Add Redis connection config:
  - Host, port, password, DB number
  - Connection pool size (min/max connections)
  - Timeout settings
  - TLS configuration (if using secure Redis)
- Add circuit breaker thresholds:
  - Failure threshold (N consecutive failures)
  - Timeout duration
  - Half-open max requests
  - Success threshold to close
- Add industry-specific confidence thresholds (YAML config file)
- Add feature flags for optimizations (enable/disable per optimization)
- Add adaptive delay configuration:
  - Min/max delays
  - Backoff multipliers
  - Response code delay mappings
- Add robots.txt enforcement toggle (respect robots.txt: true/false)
- Add observability config:
  - Prometheus endpoint configuration
  - OpenTelemetry exporter settings
  - Metrics collection intervals

## Best Practices & Additional Improvements

### Observability & Monitoring

**Problem**: Metrics infrastructure exists (Prometheus/OpenTelemetry) but not fully utilized for optimizations.

**Files to modify**:

- `services/classification-service/internal/handlers/classification.go` - Add comprehensive metrics
- Create: `services/classification-service/internal/observability/metrics.go` - Centralized metrics

**Implementation**:

1. Add Prometheus metrics for each optimization:
   - Request deduplication hit rate (counter)
   - Cache hit/miss rates (in-memory and Redis) (counter)
   - Circuit breaker state transitions (gauge)
   - Early termination frequency (counter)
   - Parallel processing duration (histogram)
   - Keyword extraction accuracy metrics (histogram of valid vs invalid keywords)
   - Robots.txt compliance rate (counter)
   - Adaptive delay effectiveness (histogram of delays applied)
2. Add OpenTelemetry spans for:
   - Request lifecycle (start to finish)
   - External service calls (Python ML, Redis)
   - Cache operations (hit/miss)
   - Parallel processing coordination
   - Circuit breaker state changes
3. Add structured logging with correlation IDs:
   - Request ID propagation through all layers
   - Trace ID for distributed tracing
   - Context-aware logging
4. Create Grafana dashboards for:
   - Performance metrics (latency p50/p95/p99, throughput)
   - Accuracy metrics (confidence distribution, keyword quality)
   - Error rates by type (403, 429, 500, timeout)
   - Cache effectiveness (hit rate, miss rate, eviction rate)
   - Circuit breaker status (open/closed/half-open per service)
   - Resource usage (CPU, memory, network, database connections)

**Expected Impact**: Better visibility, faster debugging, proactive issue detection, data-driven optimization

### Security Best Practices

**Implementation**:

1. **Input Validation & Sanitization**:
   - Enhance existing sanitization (XSS, SQL injection)
   - Validate URL formats strictly
   - Enforce maximum request size limits (prevent DoS)
   - Validate content length before processing
2. **Rate Limiting**:
   - Verify existing rate limiting effectiveness
   - Add per-IP rate limiting if not present
   - Add per-client rate limiting
   - Track and alert on rate limit violations
3. **Timeout Enforcement**:
   - Enforce timeouts at all levels (HTTP, database, external services)
   - Prevent resource exhaustion
   - Add timeout metrics and alerts
4. **Secure External Connections**:
   - Use TLS for Redis connections
   - Verify SSL certificates for external APIs
   - Use secure authentication for Redis
5. **Request Deduplication Security**:
   - Prevent replay attacks through deduplication
   - Add request signature validation
   - Implement request expiration
6. **Error Handling Security**:
   - Don't expose internal details in error messages
   - Log security events separately
   - Monitor for suspicious patterns

### Error Handling Best Practices

**Implementation**:

1. **Structured Error Types**:
   - Define error codes for each error type
   - Use custom error types with context
   - Implement error wrapping with `fmt.Errorf` and `%w`
2. **Error Classification**:
   - Retryable vs non-retryable error classification
   - Permanent vs transient errors
   - User errors vs system errors
3. **Error Aggregation**:
   - Aggregate similar errors
   - Track error rates by type
   - Alert on error rate spikes
4. **Graceful Degradation**:
   - Fallback at all levels (service, feature, operation)
   - Partial results when possible
   - Clear error messages to users
5. **Error Reporting**:
   - Structured error logging
   - Error correlation with request IDs
   - Error metrics and dashboards

### Code Quality Best Practices

**Implementation**:

1. **Test Coverage**:
   - Target 80%+ unit test coverage
   - Integration tests for critical paths
   - Performance benchmarks for each optimization
2. **Code Review Checklist**:
   - Error handling verification
   - Context propagation
   - Resource cleanup (defer, close)
   - Concurrency safety
   - Performance considerations
3. **Documentation**:
   - Document complex algorithms (keyword validation, ensemble voting)
   - Document optimization decisions and trade-offs
   - Document configuration options
4. **Type Safety**:
   - Avoid `interface{}` where possible
   - Use strong typing
   - Leverage Go's type system
5. **Context Propagation**:
   - Always pass context for cancellation/timeouts
   - Use context for request-scoped values
   - Respect context cancellation

## Documentation Updates

### Code Documentation

- Document new keyword validation logic (algorithm, dictionary source, validation rules)
- Document request deduplication behavior (TTL, cleanup strategy, race condition handling)
- Document parallel processing coordination (synchronization points, timeout handling)
- Document cache strategies (in-memory vs Redis, TTL, invalidation, fallback)
- Document circuit breaker behavior (state machine, thresholds, recovery)
- Document ensemble voting algorithm (weight calculation, consensus logic, fallback)
- Document robots.txt enforcement (crawl delay calculation, compliance, logging)
- Document adaptive delay strategy (response code mappings, backoff logic)

### API Documentation

- Update response times in API docs (reflect new performance targets)
- Document new streaming response format (if implemented) with examples
- Update accuracy metrics (new baseline after optimizations)
- Document new response fields (cache headers, confidence breakdown, optimization flags)
- Document error codes and retry behavior
- Document rate limiting headers and behavior

### Deployment Documentation

- Document Redis setup requirements:
  - HA (High Availability) configuration
  - Persistence settings
  - Monitoring and alerting
  - Backup and recovery
- Document new configuration options (all feature flags, thresholds, delays)
- Document monitoring and alerting setup:
  - Prometheus scraping configuration
  - Grafana dashboard setup
  - Alert rules and thresholds
  - Log aggregation setup
- Document performance tuning guidelines:
  - Connection pool sizing
  - Cache sizing recommendations
  - Timeout configuration
  - Resource limits
- Document rollback procedures for each optimization
- Document capacity planning:
  - Expected load calculations
  - Resource requirements (CPU, memory, network)
  - Scaling recommendations

### Runbook Documentation

- Troubleshooting guide for each optimization:
  - Common issues and solutions
  - Debugging steps
  - Log analysis
- Performance degradation investigation steps:
  - Metrics to check
  - Common causes
  - Resolution steps
- Cache invalidation procedures:
  - When to invalidate
  - How to invalidate
  - Impact assessment
- Circuit breaker recovery procedures:
  - Manual recovery steps
  - Automatic recovery monitoring
- Accuracy regression investigation:
  - Metrics to monitor
  - Root cause analysis
  - Fix procedures
