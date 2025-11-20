# API Gateway Performance Testing

**Date:** 2025-01-27

## Overview

Comprehensive performance testing suite for the API Gateway that measures:
- API response times (p50, p95, p99)
- Concurrent request handling
- Caching effectiveness
- Slow query identification

## Performance Requirements

According to the implementation plan:
- **P95 Response Time:** < 500ms for all endpoints
- **P99 Response Time:** < 1000ms
- **Success Rate:** >= 95%
- **Concurrent Users:** Support 20+ concurrent users

## Test Structure

### Test Files

1. **`performance_test.go`** - Main performance test suite
   - `TestPerformanceAPIResponseTimes` - Measures response times for all endpoints
   - `TestPerformanceConcurrentRequests` - Tests concurrent request handling
   - `TestPerformanceCachingEffectiveness` - Tests caching effectiveness
   - `TestPerformanceSlowQueries` - Identifies slow queries
   - `BenchmarkAPIEndpoints` - Go benchmarks for API endpoints

2. **`run-performance-tests.sh`** - Bash script for live performance testing
   - Tests against running API Gateway
   - Measures response times with curl
   - Calculates percentiles and statistics

## Running Tests

### Unit Tests (No Services Required)

```bash
cd services/api-gateway
go test ./test -v -run TestPerformance
```

These tests use `httptest` and measure middleware/router performance.

### Benchmarks

```bash
cd services/api-gateway
go test ./test -bench=BenchmarkAPIEndpoints -benchmem
```

### Live Performance Tests (Against Running API Gateway)

```bash
# Set environment variables
export API_GATEWAY_URL=http://localhost:8080
export TEST_MERCHANT_ID=merchant-123
export ITERATIONS=50
export CONCURRENT=10

# Run performance tests
cd services/api-gateway/test
./run-performance-tests.sh
```

## Test Coverage

### Endpoints Tested

#### Health & Metrics
- ✅ `/health` - Health check (target: < 100ms p95)

#### Merchant Routes
- ✅ `GET /api/v1/merchants` - List merchants (target: < 500ms p95)
- ✅ `GET /api/v1/merchants/{id}` - Get merchant by ID (target: < 500ms p95)
- ✅ `GET /api/v1/merchants/analytics` - Portfolio analytics (target: < 500ms p95)
- ✅ `GET /api/v1/merchants/statistics` - Portfolio statistics (target: < 500ms p95)

#### Analytics Routes
- ✅ `GET /api/v1/analytics/trends` - Risk trends (target: < 500ms p95)
- ✅ `GET /api/v1/analytics/insights` - Risk insights (target: < 500ms p95)

#### Risk Assessment Routes
- ✅ `GET /api/v1/risk/benchmarks` - Risk benchmarks (target: < 500ms p95)
- ✅ `GET /api/v1/risk/metrics` - Risk metrics (target: < 500ms p95)

### Performance Metrics Collected

1. **Response Time Statistics**
   - Min, Max, Mean, Median
   - P50, P95, P99 percentiles
   - Success/Failure counts

2. **Concurrent Request Handling**
   - Tests with 1, 5, 10, 20, 50 concurrent requests
   - Measures performance degradation under load
   - Identifies bottlenecks

3. **Caching Effectiveness**
   - First request (cache miss) vs subsequent requests (cache hit)
   - Cache speedup calculation
   - Cache improvement percentage

4. **Slow Query Identification**
   - Identifies queries with P95 > 1s
   - Flags queries with P95 > 500ms for monitoring
   - Provides optimization recommendations

## Test Results Interpretation

### Response Time Percentiles

- **P50 (Median):** 50% of requests complete within this time
- **P95:** 95% of requests complete within this time (target: < 500ms)
- **P99:** 99% of requests complete within this time (target: < 1000ms)

### Success Rate

- **Target:** >= 95%
- **Calculation:** (Successful requests / Total requests) * 100

### Performance Thresholds

| Endpoint Type | P95 Target | P99 Target |
|--------------|------------|------------|
| Health Check | < 100ms | < 200ms |
| Simple Queries | < 500ms | < 1000ms |
| Complex Queries | < 1000ms | < 2000ms |

## Optimization Recommendations

### If P95 > 500ms

1. **Check Backend Service Performance**
   - Verify backend services are responding quickly
   - Check database query performance
   - Review service logs for slow operations

2. **Enable Caching**
   - Verify caching is enabled
   - Check cache hit rates
   - Adjust cache TTL if needed

3. **Optimize Database Queries**
   - Review slow query logs
   - Add database indexes
   - Optimize query patterns

4. **Review Middleware**
   - Check middleware overhead
   - Optimize authentication checks
   - Review logging verbosity

### If Success Rate < 95%

1. **Check Error Rates**
   - Review error logs
   - Identify failing endpoints
   - Check service health

2. **Review Timeouts**
   - Increase timeout values if needed
   - Check backend service timeouts
   - Review network connectivity

## Test Execution Examples

### Example 1: Quick Performance Check

```bash
# Test health check endpoint
curl -w "\nTime: %{time_total}s\n" http://localhost:8080/health
```

### Example 2: Measure Single Endpoint

```bash
# Measure portfolio analytics endpoint
for i in {1..10}; do
    curl -w "\nTime: %{time_total}s\n" -o /dev/null -s \
        http://localhost:8080/api/v1/merchants/analytics
done
```

### Example 3: Concurrent Load Test

```bash
# Test with 20 concurrent requests
for i in {1..20}; do
    curl -w "\nTime: %{time_total}s\n" -o /dev/null -s \
        http://localhost:8080/api/v1/merchants &
done
wait
```

## Continuous Monitoring

### Recommended Metrics to Track

1. **Response Time Percentiles**
   - Track P50, P95, P99 over time
   - Alert if P95 > 500ms
   - Alert if P99 > 1000ms

2. **Success Rate**
   - Track success rate per endpoint
   - Alert if success rate < 95%

3. **Error Rates**
   - Track 4xx and 5xx errors
   - Alert if error rate > 0.1%

4. **Throughput**
   - Track requests per second
   - Monitor for capacity issues

## Next Steps

1. ✅ Performance tests created
2. Run tests against live API Gateway
3. Document baseline performance metrics
4. Identify and optimize slow queries
5. Set up continuous performance monitoring

## Files Created

1. **`performance_test.go`** - Go performance test suite
2. **`run-performance-tests.sh`** - Bash script for live testing
3. **`PERFORMANCE_TEST_README.md`** - This documentation

