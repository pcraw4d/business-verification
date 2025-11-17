# Performance Testing Guide

This guide provides comprehensive information about performance testing for the KYB Platform.

## Overview

Performance testing ensures the platform meets performance requirements and identifies bottlenecks before they impact users.

## Performance Testing Tools

### Recommended Tool: k6

k6 is recommended for load testing due to:
- JavaScript-based test scripts
- Excellent performance metrics
- Easy integration with CI/CD
- Free and open-source

### Alternative Tools

- **Artillery**: Node.js-based, good for API testing
- **Locust**: Python-based, good for complex scenarios
- **JMeter**: Java-based, comprehensive but heavier

## Setup

### Installing k6

```bash
# macOS
brew install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Windows
choco install k6
```

## Test Scenarios

### 1. API Endpoint Load Testing

Test individual API endpoints under load:

```javascript
// tests/performance/api-load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 },  // Ramp up to 20 users
    { duration: '1m', target: 20 },     // Stay at 20 users
    { duration: '30s', target: 50 },    // Ramp up to 50 users
    { duration: '1m', target: 50 },    // Stay at 50 users
    { duration: '30s', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95% of requests < 500ms
    http_req_failed: ['rate<0.01'],    // Error rate < 1%
  },
};

export default function () {
  const BASE_URL = __ENV.API_BASE_URL || 'https://api.example.com';
  
  // Test dashboard metrics endpoint
  const response = http.get(`${BASE_URL}/api/v3/dashboard/metrics`, {
    headers: {
      'Authorization': `Bearer ${__ENV.API_TOKEN || ''}`,
    },
  });
  
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  sleep(1);
}
```

### 2. Concurrent User Testing

Test system behavior under concurrent user load:

```javascript
// tests/performance/concurrent-users.js
import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: 100,        // 100 virtual users
  duration: '5m',  // Run for 5 minutes
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed: ['rate<0.05'],
  },
};

export default function () {
  const BASE_URL = __ENV.API_BASE_URL || 'https://api.example.com';
  
  // Simulate user journey
  const endpoints = [
    '/api/v3/dashboard/metrics',
    '/api/v1/merchants',
    '/api/v1/compliance/status',
  ];
  
  endpoints.forEach((endpoint) => {
    const response = http.get(`${BASE_URL}${endpoint}`);
    check(response, {
      'status is 200': (r) => r.status === 200,
    });
  });
}
```

### 3. Stress Testing

Test system limits:

```javascript
// tests/performance/stress-test.js
import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp up to 100 users
    { duration: '5m', target: 100 },     // Stay at 100 users
    { duration: '2m', target: 200 },   // Ramp up to 200 users
    { duration: '5m', target: 200 },   // Stay at 200 users
    { duration: '2m', target: 300 },   // Ramp up to 300 users
    { duration: '5m', target: 300 },   // Stay at 300 users
    { duration: '10m', target: 0 },    // Ramp down
  ],
};

export default function () {
  const BASE_URL = __ENV.API_BASE_URL || 'https://api.example.com';
  const response = http.get(`${BASE_URL}/api/v3/dashboard/metrics`);
  check(response, {
    'status is 200': (r) => r.status === 200,
  });
}
```

## Performance Baselines

### API Response Time Targets

| Endpoint | P50 Target | P95 Target | P99 Target |
|----------|------------|------------|------------|
| `/api/v3/dashboard/metrics` | < 200ms | < 500ms | < 1000ms |
| `/api/v1/merchants` | < 300ms | < 600ms | < 1200ms |
| `/api/v1/compliance/status` | < 250ms | < 500ms | < 1000ms |
| `/api/v1/sessions` | < 150ms | < 300ms | < 600ms |

### Page Load Time Targets

- **First Contentful Paint (FCP)**: < 1.5s
- **Largest Contentful Paint (LCP)**: < 2.5s
- **Time to Interactive (TTI)**: < 3.5s
- **Total Blocking Time (TBT)**: < 300ms

### Concurrent User Capacity

- **Target**: Support 100 concurrent users
- **Peak**: Handle 200 concurrent users with degraded performance
- **Maximum**: System should not crash under 300 concurrent users

## Running Performance Tests

### Local Testing

```bash
# Run API load test
k6 run tests/performance/api-load-test.js

# Run with environment variables
API_BASE_URL=https://api.example.com k6 run tests/performance/api-load-test.js

# Run with custom options
k6 run --vus 50 --duration 5m tests/performance/api-load-test.js
```

### CI/CD Integration

Add to GitHub Actions workflow:

```yaml
- name: Run performance tests
  run: |
    k6 run tests/performance/api-load-test.js
  env:
    API_BASE_URL: ${{ secrets.API_BASE_URL }}
    API_TOKEN: ${{ secrets.API_TOKEN }}
```

## Performance Metrics

### Key Metrics to Monitor

1. **Response Time**
   - Average response time
   - P50, P95, P99 percentiles
   - Min/Max response times

2. **Throughput**
   - Requests per second (RPS)
   - Successful requests per second
   - Failed requests per second

3. **Error Rate**
   - HTTP error rate (4xx, 5xx)
   - Timeout rate
   - Connection errors

4. **Resource Utilization**
   - CPU usage
   - Memory usage
   - Network I/O

## Performance Regression Testing

### Automated Regression Tests

Create performance regression test suite:

```javascript
// tests/performance/regression.js
import http from 'k6/http';
import { check, group } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const responseTime = new Trend('response_time');
const errorRate = new Rate('errors');

export const options = {
  vus: 50,
  duration: '2m',
  thresholds: {
    response_time: ['p(95)<500'],
    errors: ['rate<0.01'],
  },
};

export default function () {
  group('API Performance', () => {
    const BASE_URL = __ENV.API_BASE_URL || 'https://api.example.com';
    const response = http.get(`${BASE_URL}/api/v3/dashboard/metrics`);
    
    const success = check(response, {
      'status is 200': (r) => r.status === 200,
      'response time < 500ms': (r) => r.timings.duration < 500,
    });
    
    responseTime.add(response.timings.duration);
    errorRate.add(!success);
  });
}
```

### Baseline Comparison

Compare current performance against baselines:

```bash
# Run baseline test
k6 run --out json=baseline.json tests/performance/regression.js

# Run current test
k6 run --out json=current.json tests/performance/regression.js

# Compare results
k6 compare baseline.json current.json
```

## Performance SLAs

### Service Level Indicators (SLIs)

1. **Availability**: 99.9% uptime
2. **Latency**: P95 < 500ms for 95% of requests
3. **Error Rate**: < 1% error rate
4. **Throughput**: Support 100 RPS per service

### Service Level Objectives (SLOs)

1. **API Availability**: 99.9% (allows ~43 minutes downtime/month)
2. **Response Time**: P95 < 500ms for 99% of requests
3. **Error Rate**: < 0.5% for 99% of requests

## Performance Optimization

### Common Bottlenecks

1. **Database Queries**
   - Slow queries
   - Missing indexes
   - N+1 query problems

2. **External API Calls**
   - Slow third-party APIs
   - No timeout configuration
   - No retry logic

3. **Caching**
   - Missing cache
   - Inefficient cache keys
   - Cache invalidation issues

4. **Frontend Performance**
   - Large bundle sizes
   - Unoptimized images
   - Excessive re-renders

### Optimization Strategies

1. **Database Optimization**
   - Add indexes for frequently queried fields
   - Optimize slow queries
   - Use connection pooling

2. **Caching Strategy**
   - Implement API response caching
   - Use CDN for static assets
   - Cache database queries

3. **Code Optimization**
   - Optimize algorithms
   - Reduce unnecessary computations
   - Use lazy loading

4. **Infrastructure Scaling**
   - Horizontal scaling
   - Load balancing
   - Auto-scaling based on metrics

## Performance Testing Schedule

### Recommended Schedule

- **Daily**: Automated regression tests in CI/CD
- **Weekly**: Full performance test suite
- **Monthly**: Stress testing and capacity planning
- **Before Releases**: Comprehensive performance validation

## Reporting

### Performance Test Reports

Generate and review performance test reports:

```bash
# Generate HTML report
k6 run --out json=results.json tests/performance/api-load-test.js
k6 report results.json --output report.html

# View in browser
open report.html
```

### Metrics Dashboard

Set up metrics dashboard to track:
- Response time trends
- Error rate trends
- Throughput trends
- Resource utilization

## Troubleshooting

### High Response Times

1. Check database query performance
2. Review external API latency
3. Check network connectivity
4. Review server resource utilization

### High Error Rates

1. Check service logs
2. Review error patterns
3. Check external dependencies
4. Verify service health

### Resource Exhaustion

1. Check CPU and memory usage
2. Review connection pool sizes
3. Check for memory leaks
4. Consider scaling up resources

## Additional Resources

- [k6 Documentation](https://k6.io/docs/)
- [k6 Examples](https://github.com/grafana/k6/tree/master/examples)
- [Performance Testing Best Practices](https://k6.io/docs/test-types/)

