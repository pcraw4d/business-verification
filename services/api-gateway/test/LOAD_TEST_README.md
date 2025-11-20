# API Gateway Load Testing

**Date:** 2025-01-27

## Overview

Comprehensive load testing suite for the API Gateway that tests:
- Multiple concurrent users
- API Gateway under various load conditions
- Database queries under load
- Bottleneck identification

## Load Testing Requirements

According to the implementation plan:
- Test with multiple concurrent users
- Test API Gateway under load
- Test database queries under load
- Identify bottlenecks

## Test Structure

### Test Files

1. **`load_test.go`** - Go load testing suite
   - `TestLoadAPIGatewayUnderLoad` - Tests API Gateway under various load scenarios
   - `TestLoadDatabaseQueriesUnderLoad` - Tests database-heavy endpoints under load
   - `TestLoadIdentifyBottlenecks` - Identifies bottlenecks by testing with increasing concurrency

2. **`run-load-tests.sh`** - Bash script for live load testing
   - Tests against a running API Gateway
   - Multiple load scenarios (light, medium, heavy, stress)
   - Database-heavy endpoint testing

## Running Tests

### Go Load Tests (Unit/Integration Mode)

```bash
cd services/api-gateway
go test ./test -v -run TestLoadAPIGatewayUnderLoad
go test ./test -v -run TestLoadDatabaseQueriesUnderLoad
go test ./test -v -run TestLoadIdentifyBottlenecks
```

### Live Load Testing Script

```bash
# Start API Gateway and backend services first
cd services/api-gateway/test
./run-load-tests.sh

# Or with custom configuration
API_GATEWAY_URL=http://localhost:8080 \
CONCURRENT_USERS=100 \
REQUESTS_PER_USER=50 \
./run-load-tests.sh
```

## Load Test Scenarios

### 1. Light Load - 10 Concurrent Users
- **Concurrent Users:** 10
- **Requests per User:** 10
- **Total Requests:** 100
- **Endpoint:** `/health`
- **Purpose:** Baseline performance measurement

### 2. Medium Load - 50 Concurrent Users
- **Concurrent Users:** 50
- **Requests per User:** 20
- **Total Requests:** 1,000
- **Endpoint:** `/api/v1/merchants`
- **Purpose:** Normal operating conditions

### 3. Heavy Load - 100 Concurrent Users
- **Concurrent Users:** 100
- **Requests per User:** 50
- **Total Requests:** 5,000
- **Endpoint:** `/api/v1/merchants`
- **Purpose:** Peak load testing

### 4. Stress Test - 200 Concurrent Users
- **Concurrent Users:** 200
- **Requests per User:** 100
- **Total Requests:** 20,000
- **Endpoint:** `/api/v1/merchants`
- **Purpose:** Breaking point identification

### 5. Database Load Tests
- **Concurrent Users:** 50
- **Requests per User:** 20
- **Endpoints:**
  - `/api/v1/merchants/statistics`
  - `/api/v1/merchants/analytics`
  - `/api/v1/analytics/trends?timeframe=30d`
- **Purpose:** Database query performance under load

## Performance Metrics

### Metrics Collected

1. **Total Requests** - Total number of requests made
2. **Successful Requests** - Number of successful (2xx) responses
3. **Failed Requests** - Number of failed (4xx/5xx) responses
4. **Error Rate** - Percentage of failed requests
5. **Throughput** - Requests per second
6. **Response Times:**
   - Min
   - Max
   - Mean
   - P50 (Median)
   - P95
   - P99

### Performance Thresholds

- **Error Rate:** < 1% (acceptable), < 5% (warning)
- **Throughput:** > 10 req/s (acceptable)
- **P95 Response Time:** < 2s under load (acceptable)
- **Success Rate:** > 95% (acceptable)

## Bottleneck Identification

The `TestLoadIdentifyBottlenecks` test identifies bottlenecks by:
1. Testing with increasing concurrency levels (10, 25, 50, 100, 200)
2. Monitoring throughput degradation
3. Monitoring response time degradation
4. Identifying the concurrency level where performance degrades

### Bottleneck Indicators

- **Throughput Decrease:** > 20% decrease in throughput
- **Response Time Increase:** > 50% increase in P95 response time
- **Error Rate Spike:** > 5% error rate

## Test Results Interpretation

### ✅ Good Performance
- Error rate < 1%
- Throughput stable across concurrency levels
- P95 response time < 2s
- No significant degradation

### ⚠️ Needs Optimization
- Error rate 1-5%
- Throughput decreases > 20% at higher concurrency
- P95 response time > 2s
- Some degradation observed

### 🔥 Critical Issues
- Error rate > 5%
- Throughput decreases > 50% at higher concurrency
- P95 response time > 5s
- Significant degradation or failures

## Optimization Recommendations

### If Error Rate > 1%

1. **Check Backend Services**
   - Verify all backend services are running
   - Check service health endpoints
   - Monitor service resource usage

2. **Check Database**
   - Verify database connection pool settings
   - Check for slow queries
   - Monitor database resource usage

3. **Check Rate Limiting**
   - Verify rate limiting configuration
   - Adjust rate limits if too restrictive
   - Check for rate limit errors

### If Throughput Degrades

1. **Connection Pooling**
   - Increase connection pool sizes
   - Optimize connection reuse
   - Check for connection leaks

2. **Caching**
   - Verify caching is enabled
   - Check cache hit rates
   - Optimize cache TTLs

3. **Database Optimization**
   - Add database indexes
   - Optimize slow queries
   - Consider read replicas

### If Response Time Increases

1. **Backend Service Performance**
   - Profile backend services
   - Identify slow endpoints
   - Optimize database queries

2. **Network Latency**
   - Check network connectivity
   - Verify service proximity
   - Consider CDN for static assets

3. **Resource Constraints**
   - Check CPU usage
   - Check memory usage
   - Scale services if needed

## Continuous Monitoring

### Recommended Metrics to Track

1. **Request Rates**
   - Requests per second by endpoint
   - Peak request rates
   - Request rate trends

2. **Error Rates**
   - Error rate by endpoint
   - Error rate by status code
   - Error rate trends

3. **Response Times**
   - P50, P95, P99 by endpoint
   - Response time trends
   - Slow query identification

4. **Resource Usage**
   - CPU usage
   - Memory usage
   - Database connection pool usage

## Next Steps

1. ✅ Load tests created
2. Run tests against live API Gateway
3. Document baseline performance metrics
4. Identify and optimize bottlenecks
5. Set up continuous load testing

## Files Created

1. **`load_test.go`** - Go load testing suite (490+ lines)
2. **`run-load-tests.sh`** - Bash script for live load testing
3. **`LOAD_TEST_README.md`** - This documentation

## Conclusion

**Load Testing: ✅ COMPLETE**

Comprehensive load testing suite created covering:
- Multiple concurrent users (10, 50, 100, 200)
- API Gateway under various load conditions
- Database queries under load
- Bottleneck identification

All tests are ready to run when the API Gateway and backend services are available.

