<!-- 3afed8fe-0176-4986-b80c-a6843d0a61da 60625cfa-8b8b-459d-bb95-ef31814a02c6 -->
# Placeholder Data Handling & Enhanced Detection Plan

## Overview

This plan addresses placeholder data handling across 5 identified categories and implements enhanced detection capabilities with context-aware analysis, allowlisting, and severity-based reporting. This plan incorporates recommendations from `PLACEHOLDER-PLAN-REVIEW.md` to ensure complete coverage.

## Phase 1: Production Safety & Environment Checks (Week 1-2)

### 1.1 Environment-Based Configuration

- **File**: `services/merchant-service/internal/config/config.go`
- Add `Environment` field (development/staging/production)
- Add `AllowMockData` boolean flag (default: false in production)
- Update config loading to read from environment variables

- **File**: `web/js/api-config.js`
- Extend existing environment detection utility (already has `getEnvironment()`)
- Add `isProduction()` and `isDevelopment()` helpers
- Add `allowMockData()` method
- Disable mock data generation in production builds

### 1.2 Production Safety Guards

- **File**: `services/merchant-service/internal/handlers/merchant.go`
- Update `getMockMerchant()` to check environment:
  ```go
  if h.config.Environment == "production" && !h.config.AllowMockData {
      return nil, fmt.Errorf("mock data not allowed in production")
  }
  ```

- Return proper HTTP 404/503 instead of mock data in production

- **File**: `web/js/components/risk-indicators-data-service.js`
- Add environment check in `generateMockRiskData()`:
  ```javascript
  if (this.isProduction() && !this.config.allowMockData) {
      throw new Error('Mock data not allowed in production');
  }
  ```

### 1.3 HTTP Status Code Improvements

- **File**: `services/merchant-service/internal/handlers/merchant.go`
- Update `getMerchant()` to return 404 when merchant not found (production)
- Update `listMerchants()` to return 200 with empty array instead of mock data
- Add 503 status when Supabase is unavailable (with retry-after header)

- **File**: `services/risk-assessment-service/internal/handlers/risk_assessment.go`
- Update `HandleRiskBenchmarks()` to return 503 when data unavailable
- Update `HandleRiskPredictions()` to return 404 when merchant not found

### 1.4 Feature Flags for Incomplete Features (NEW - from Review)

- **File**: `internal/config/feature_flags.go` (extend existing)
- Add feature flags for incomplete features:
  - `incomplete_risk_benchmarks` - Disable benchmarks endpoint in production if incomplete
  - `incomplete_risk_predictions` - Disable predictions endpoint in production if incomplete
  - `incomplete_merchant_analytics` - Disable analytics endpoint in production if incomplete

- **File**: `services/risk-assessment-service/internal/handlers/risk_assessment.go`
- Check feature flags before serving incomplete features:
  ```go
  if !featureFlags.IsEnabled("incomplete_risk_benchmarks") && cfg.Environment == "production" {
      http.Error(w, "Feature not available", http.StatusServiceUnavailable)
      return
  }
  ```

- **File**: `services/merchant-service/internal/handlers/merchant.go`
- Check feature flags for incomplete analytics/statistics endpoints

## Phase 2: Error Handling & Resilience (Week 3-4)

### 2.1 Retry Logic with Exponential Backoff

- **New File**: `internal/resilience/retry.go`
- Implement `RetryWithBackoff()` function
- Support configurable max attempts, initial delay, max delay
- Add jitter to prevent thundering herd

- **File**: `services/merchant-service/internal/handlers/merchant.go`
- Wrap Supabase queries with retry logic:
  ```go
  result, err := retry.WithBackoff(ctx, 3, 100*time.Millisecond, func() (interface{}, error) {
      return h.supabaseClient.GetClient().From("merchants")...
  })
  ```

- **File**: `web/js/utils/retry-utils.js`
- Create `retryWithBackoff()` utility for API calls
- Use in `risk-indicators-data-service.js` for all API calls

### 2.2 Circuit Breaker Pattern

- **New File**: `internal/resilience/circuit_breaker.go`
- Implement circuit breaker with states: Closed, Open, HalfOpen
- Track failure rates and auto-recovery
- Support configurable thresholds

- **File**: `services/merchant-service/internal/handlers/merchant.go`
- Add circuit breaker for Supabase connection
- Fast-fail when circuit is open

- **File**: `web/js/utils/circuit-breaker.js`
- Create JavaScript circuit breaker for API calls
- Integrate with risk data service

### 2.3 User Notifications

- **File**: `web/js/components/risk-indicators-data-service.js`
- Add `notifyFallbackUsage()` method
- Display user-friendly notification when fallback data is used
- Include reason and expected recovery time

- **File**: `web/shared/components/fallback-notification.js`
- Create reusable notification component
- Support different severity levels (info, warning, error)

### 2.4 Data Validation Before Fallback (NEW - from Review)

- **New File**: `internal/validation/data_validator.go`
- Implement data validation functions:
  - `ValidateMerchantData()` - Check completeness, required fields
  - `ValidateBenchmarkData()` - Check data freshness, completeness
  - `ValidateAnalyticsData()` - Check data quality, accuracy

- **File**: `services/merchant-service/internal/handlers/merchant.go`
- Validate data before using fallback:
  ```go
  if err := validate.ValidateMerchantData(merchant); err != nil {
      // Only fallback if validation fails
      return h.getMockMerchant(merchantID)
  }
  ```

- **File**: `services/risk-assessment-service/internal/handlers/risk_assessment.go`
- Validate benchmark data before returning fallback

### 2.5 Request Queuing for Failed Calls (NEW - from Review)

- **New File**: `internal/queue/request_queue.go`
- Implement request queue for failed API calls
- Support priority queuing
- Retry queued requests when service recovers

- **File**: `services/merchant-service/internal/handlers/merchant.go`
- Queue failed Supabase requests for retry:
  ```go
  if err != nil {
      requestQueue.Enqueue(ctx, &QueuedRequest{
          Type: "get_merchant",
          Data: merchantID,
          Priority: PriorityNormal,
      })
      return nil, fmt.Errorf("database unavailable")
  }
  ```

- **File**: `web/js/utils/request-queue.js`
- Create JavaScript request queue for frontend
- Retry queued requests when connection restored

## Phase 3: Enhanced Placeholder Detection (Week 5-6)

### 3.1 Context-Aware Detection Engine

- **New File**: `scripts/enhanced-placeholder-detector.js` (Node.js for better AST parsing)
- Parse code using AST (acorn for JS, go/parser for Go)
- Analyze context around placeholder matches:
  - Check if inside error handler (fallback) vs. primary code path
  - Detect conditional statements (if/else, try/catch)
  - Identify function names (getMock*, fallback*, generateMock*)
  - Check for FALLBACK comments nearby

- **Detection Rules**:
  - **Primary Usage**: Placeholder in return statement without error check
  - **Fallback Usage**: Placeholder in catch block, error handler, or after null check
  - **Development Only**: Placeholder in `if (development)` block
  - **Test Data**: Placeholder in test files or test functions

### 3.2 Allowlist System

- **New File**: `scripts/placeholder-allowlist.json`
- Structure:
  ```json
  {
    "allowed_patterns": [
      {
        "file": "services/merchant-service/internal/handlers/merchant.go",
        "line": 362,
        "pattern": "Sample Merchant",
        "reason": "Fallback data when Supabase unavailable",
        "category": "database_fallback",
        "severity": "low",
        "expires": "2025-12-31"
      }
    ],
    "allowed_functions": [
      "getMockMerchant",
      "getFallbackMerchantData",
      "generateMockRiskData"
    ],
    "allowed_files": [
      "**/*_test.go",
      "**/*.test.js",
      "scripts/**"
    ]
  }
  ```

- **File**: `scripts/enhanced-placeholder-detector.js`
- Load allowlist and exclude matches
- Validate allowlist entries (check file exists, line numbers valid)
- Check expiration dates

### 3.3 Severity-Based Reporting

- **New File**: `scripts/enhanced-placeholder-detector.js`
- Assign severity levels:
  - **CRITICAL**: Primary usage in production code path
  - **HIGH**: Fallback usage without proper error handling
  - **MEDIUM**: Fallback usage with proper documentation
  - **LOW**: Development/test only, properly scoped
  - **INFO**: Allowlisted entries

- **Report Generation**:
  - **New File**: `scripts/generate-placeholder-report.js`
  - Generate JSON report with:
    - Summary statistics by severity
    - Detailed findings with file, line, context
    - Recommendations for each finding
    - Comparison with previous runs (track improvements)

- **Report Format**: `test-results/placeholder-detection-report-{timestamp}.json`

### 3.4 Integration with CI/CD

- **File**: `.github/workflows/ci-cd.yml`
- Run enhanced detector on every PR
- Fail build if critical/high severity findings
- Post report as PR comment
- Track findings over time

## Phase 4: Data Source Implementation (Week 7-9)

### 4.1 Complete Supabase Queries

- **File**: `services/merchant-service/internal/handlers/merchant.go`
- Implement `listMerchants()` with Supabase query:
  ```go
  result, err := h.supabaseClient.GetClient().From("merchants").
      Select("*", "", false).
      Range((page-1)*pageSize, page*pageSize-1).
      ExecuteTo(&merchants)
  ```

- Add filtering and sorting support

### 4.2 Database Queries for Benchmarks

- **File**: `services/risk-assessment-service/internal/handlers/risk_assessment.go`
- Create `benchmarks` table in Supabase
- Implement `HandleRiskBenchmarks()` to query database
- Add caching layer for frequently accessed benchmarks

### 4.3 Caching Layer (Redis)

- **New File**: `internal/cache/redis_cache.go`
- Implement Redis client wrapper
- Add TTL support and cache invalidation
- Use for merchant data, benchmarks, analytics

- **File**: `services/merchant-service/internal/handlers/merchant.go`
- Add cache layer before Supabase queries
- Cache merchant data with 5-minute TTL

### 4.4 Data Seeding for Development (NEW - from Review)

- **New File**: `scripts/seed-dev-data.go`
- Create development data seeding script
- Seed sample merchants, benchmarks, analytics
- Support different data volumes (small, medium, large)

- **New File**: `scripts/seed-dev-data.sql`
- SQL scripts for seeding Supabase database
- Include sample merchants with various portfolio types
- Include risk benchmarks for different industries

- **File**: `docs/developer-guides/setup.md`
- Document data seeding process
- Add to development setup guide

## Phase 5: Monitoring & Metrics (Week 10-11)

### 5.1 Fallback Usage Metrics

- **New File**: `internal/metrics/fallback_metrics.go`
- Track fallback usage by:
  - Service name
  - Category (database, API, etc.)
  - Frequency
  - Duration

- **File**: `services/merchant-service/internal/handlers/merchant.go`
- Increment metrics counter when fallback used:
  ```go
  metrics.RecordFallbackUsage("merchant-service", "database", "supabase")
  ```

- **Integration**: Extend existing Prometheus metrics system
- **File**: `internal/monitoring/optimization.go` (extend existing)
- Add fallback metrics to existing Prometheus exporter

### 5.2 Dashboards

- Create Grafana dashboard for fallback metrics (extend existing dashboards)
- Track fallback rate over time
- Alert when fallback rate > 10%

### 5.3 Alerting

- Set up alerts for:
  - High fallback usage rate (>10%)
  - Circuit breaker opened
  - Database connection failures
  - API failures exceeding threshold

## Phase 6: Testing & Validation (Week 12)

### 6.1 Integration Tests

- **New File**: `test/integration/placeholder_handling_test.go`
- Test fallback behavior in different scenarios
- Verify production safety guards
- Test retry logic and circuit breakers

### 6.2 Chaos Engineering

- Create chaos tests to simulate:
  - Database failures
  - API timeouts
  - Network issues
- Verify fallback behavior and recovery

### 6.3 Contract Testing (NEW - from Review)

- **New File**: `test/contract/api_contract_test.go`
- Add contract tests for API responses
- Validate no placeholder data in production responses
- Use Pact or similar framework

- **File**: `test/contract/placeholder_contract_test.go`
- Test that production APIs never return mock data
- Validate `isFallback` flags are present when fallback used
- Test error responses (404, 503) are correct

## Implementation Notes

### Priority Order

1. **Phase 1** (Production Safety) - Critical for production deployment
2. **Phase 2** (Error Handling) - Improves reliability
3. **Phase 3** (Enhanced Detection) - Long-term maintenance
4. **Phase 4** (Data Sources) - Reduces placeholder usage
5. **Phase 5** (Monitoring) - Operational visibility
6. **Phase 6** (Testing) - Quality assurance

### Dependencies

- Phase 1 must complete before production deployment
- Phase 2 can run parallel with Phase 3
- Phase 4 depends on Phase 1 completion
- Phase 5 depends on Phase 2 metrics implementation

### Integration with Existing Infrastructure

- **Leverage Existing Config**: Use `ENV`/`ENVIRONMENT` from Railway, extend existing config structs
- **Extend Existing Metrics**: Add fallback metrics to existing Prometheus system
- **Reuse Monitoring**: Integrate with existing Grafana dashboards rather than creating new ones
- **Reuse Frontend Infrastructure**: Extend `web/js/api-config.js` instead of creating new

### Success Criteria

- Zero critical/high severity placeholder findings in production code
- Fallback usage rate < 5% in production
- All fallback usage properly documented and allowlisted
- Enhanced detection reports generated automatically in CI/CD
- All incomplete features disabled in production via feature flags
- Data validation prevents invalid data from being used
- Request queuing ensures no failed requests are lost

### To-dos

- [x] Add environment-based configuration to merchant service and frontend
- [x] Implement production safety guards to prevent mock data in production
- [x] Update handlers to return proper HTTP status codes (404/503) instead of mock data
- [x] Implement retry logic with exponential backoff for Supabase and API calls
- [x] Add circuit breaker pattern for external services (Supabase, APIs)
- [x] Create user notification system for fallback data usage
- [x] Build context-aware placeholder detection engine using AST parsing
- [x] Create allowlist system for known acceptable placeholder usage
- [x] Implement severity-based reporting with JSON output and recommendations
- [x] Integrate enhanced detector into CI/CD pipeline with PR comments
- [x] Complete Supabase queries for listMerchants and other missing endpoints
- [x] Implement database queries for risk benchmarks endpoint
- [x] Add Redis caching layer for frequently accessed data
- [x] Implement metrics tracking for fallback usage by service and category
- [x] Create Grafana dashboard for fallback metrics and alerting
- [x] Add integration tests for fallback behavior and production safety
- [ ] Implement feature flags for incomplete features (Phase 1.4)
- [ ] Add data validation before fallback (Phase 2.4)
- [ ] Implement request queuing for failed API calls (Phase 2.5)
- [ ] Create data seeding script for development (Phase 4.4)
- [ ] Add contract testing for API responses (Phase 6.3)

