# Performance Monitoring and Load Testing Guide

This document provides comprehensive guidance on performance monitoring and load testing for the Risk Assessment Service, with a target of **1000 requests per minute**.

## Table of Contents

1. [Performance Monitoring](#performance-monitoring)
2. [Load Testing Framework](#load-testing-framework)
3. [Performance Targets](#performance-targets)
4. [Monitoring Endpoints](#monitoring-endpoints)
5. [Load Testing Commands](#load-testing-commands)
6. [Performance Analysis](#performance-analysis)
7. [Troubleshooting](#troubleshooting)

## Performance Monitoring

### Overview

The Risk Assessment Service includes comprehensive performance monitoring capabilities that track:

- **Request Metrics**: Total requests, success/failure rates, response times
- **Throughput Metrics**: Requests per second, requests per minute
- **System Metrics**: Memory usage, CPU usage, goroutine count
- **Alert System**: Automated alerts for performance issues

### Key Features

- **Real-time Monitoring**: Continuous performance tracking
- **Automated Alerts**: Configurable alerts for performance thresholds
- **Historical Data**: Performance trend analysis
- **Health Checks**: Service health monitoring
- **Target Validation**: Automatic validation against performance targets

## Load Testing Framework

### Test Types

#### 1. Standard Load Test
- **Purpose**: Validate normal operating conditions
- **Target**: 1000 requests/minute (16.67 RPS)
- **Duration**: 5 minutes
- **Users**: 20 concurrent users

#### 2. Stress Test
- **Purpose**: Find breaking points and system limits
- **Method**: Gradually increase load until failure
- **Metrics**: Maximum sustainable throughput

#### 3. Spike Test
- **Purpose**: Test system recovery from traffic spikes
- **Method**: Normal load → 10x spike → recovery
- **Duration**: 2 minutes spike, 5 minutes recovery

#### 4. High Load Test
- **Purpose**: Validate 1000 req/min target
- **Users**: 50 concurrent users
- **Requests**: 20 requests per user
- **Target**: 16.67 RPS sustained

#### 5. Sustained Load Test
- **Purpose**: Test long-term stability
- **Duration**: 10 minutes
- **Users**: 30 concurrent users
- **Focus**: Memory leaks, resource exhaustion

## Performance Targets

### Primary Targets

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| **Throughput** | 1000 req/min | < 800 req/min |
| **Response Time** | < 1 second | > 1 second |
| **Error Rate** | < 1% | > 5% |
| **Availability** | 99.9% | < 99% |

### Secondary Targets

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| **Memory Usage** | < 500MB | > 500MB |
| **CPU Usage** | < 80% | > 80% |
| **Goroutines** | < 1000 | > 1000 |

## Monitoring Endpoints

### Performance Statistics
```http
GET /api/v1/performance/stats
```

**Response:**
```json
{
  "request_count": 1500,
  "successful_requests": 1485,
  "failed_requests": 15,
  "requests_per_second": 16.67,
  "requests_per_minute": 1000.2,
  "error_rate": 0.01,
  "average_response_time": "850ms",
  "max_response_time": "1.2s",
  "min_response_time": "200ms",
  "memory_usage": 256000000,
  "cpu_usage": 45.5,
  "goroutine_count": 150,
  "timestamp": "2023-01-01T12:00:00Z"
}
```

### Performance Alerts
```http
GET /api/v1/performance/alerts
```

**Response:**
```json
{
  "alerts": [
    {
      "type": "high_latency",
      "message": "High latency detected: 1.2s (target: 1s)",
      "timestamp": "2023-01-01T12:00:00Z",
      "severity": "warning",
      "threshold": 1000000000,
      "current_value": 1200000000
    }
  ],
  "count": 1,
  "timestamp": "2023-01-01T12:00:00Z"
}
```

### Performance Health
```http
GET /api/v1/performance/health
```

**Response:**
```json
{
  "status": "healthy",
  "is_healthy": true,
  "requests_per_second": 16.67,
  "requests_per_minute": 1000.2,
  "error_rate": 0.01,
  "max_response_time": "1.2s",
  "memory_usage_mb": 256,
  "cpu_usage_percent": 45.5,
  "goroutine_count": 150,
  "alert_count": 0,
  "timestamp": "2023-01-01T12:00:00Z"
}
```

### Update Performance Targets
```http
POST /api/v1/performance/targets
```

**Request:**
```json
{
  "rps": 16.67,
  "latency": "1s",
  "error_rate": 0.01,
  "throughput": 1000
}
```

### Reset Performance Metrics
```http
POST /api/v1/performance/reset
```

### Clear Performance Alerts
```http
POST /api/v1/performance/alerts/clear
```

## Load Testing Commands

### Quick Start

```bash
# Run all load tests
make load-test

# Run custom load test
make load-test-custom

# Run stress test
make stress-test

# Run spike test
make spike-test
```

### Command Line Tool

```bash
# Basic load test
go run ./cmd/load_test.go -url=http://localhost:8080 -duration=5m -users=20

# Stress test
go run ./cmd/load_test.go -type=stress -duration=10m -users=50

# Spike test
go run ./cmd/load_test.go -type=spike -duration=5m -users=30

# High load test (1000 req/min target)
go run ./cmd/load_test.go -users=50 -requests=20 -rps=16.67 -duration=5m
```

### Load Testing Script

```bash
# Run comprehensive load test suite
./scripts/run_load_tests.sh

# With custom parameters
SERVICE_URL=http://localhost:8080 \
TEST_DURATION=10m \
CONCURRENT_USERS=30 \
TARGET_RPS=16.67 \
./scripts/run_load_tests.sh
```

## Performance Analysis

### Key Metrics to Monitor

#### 1. Throughput Analysis
- **Target**: 1000 requests/minute
- **Measurement**: Requests per minute over time
- **Alert**: If sustained throughput < 800 req/min

#### 2. Response Time Analysis
- **Target**: < 1 second average
- **Measurement**: 95th percentile response time
- **Alert**: If 95th percentile > 1 second

#### 3. Error Rate Analysis
- **Target**: < 1% error rate
- **Measurement**: Failed requests / total requests
- **Alert**: If error rate > 5%

#### 4. Resource Usage Analysis
- **Memory**: Monitor for memory leaks
- **CPU**: Ensure CPU usage < 80%
- **Goroutines**: Monitor for goroutine leaks

### Performance Bottlenecks

#### Common Bottlenecks
1. **Database Connections**: Connection pool exhaustion
2. **External API Calls**: Timeout or rate limiting
3. **Memory Allocation**: Excessive garbage collection
4. **CPU Intensive Operations**: ML model inference
5. **Network I/O**: Slow external service responses

#### Bottleneck Identification
```bash
# Monitor performance during load test
curl -s http://localhost:8080/api/v1/performance/stats | jq '.'

# Check for alerts
curl -s http://localhost:8080/api/v1/performance/alerts | jq '.'

# Monitor system resources
top -p $(pgrep risk-assessment-service)
```

## Troubleshooting

### Performance Issues

#### High Response Times
1. **Check Database Performance**
   ```bash
   # Monitor database connections
   curl -s http://localhost:8080/api/v1/performance/stats | jq '.database_connections'
   ```

2. **Check External API Latency**
   ```bash
   # Monitor external service response times
   curl -s http://localhost:8080/api/v1/performance/stats | jq '.external_api_latency'
   ```

3. **Check Memory Usage**
   ```bash
   # Monitor memory consumption
   curl -s http://localhost:8080/api/v1/performance/stats | jq '.memory_usage'
   ```

#### High Error Rates
1. **Check Service Health**
   ```bash
   curl -s http://localhost:8080/api/v1/performance/health
   ```

2. **Review Error Logs**
   ```bash
   # Check application logs
   tail -f /var/log/risk-assessment-service.log
   ```

3. **Check Resource Limits**
   ```bash
   # Monitor system resources
   htop
   ```

#### Low Throughput
1. **Check Rate Limiting**
   ```bash
   # Verify rate limit settings
   curl -s http://localhost:8080/api/v1/performance/stats | jq '.rate_limit_status'
   ```

2. **Check Concurrent User Limits**
   ```bash
   # Monitor active connections
   netstat -an | grep :8080 | wc -l
   ```

3. **Check CPU Usage**
   ```bash
   # Monitor CPU utilization
   top -p $(pgrep risk-assessment-service)
   ```

### Load Test Failures

#### Common Issues
1. **Service Not Ready**: Ensure service is running and healthy
2. **Insufficient Resources**: Check CPU, memory, and network
3. **Rate Limiting**: Adjust rate limits for load testing
4. **Database Connections**: Increase connection pool size
5. **External API Limits**: Check external service rate limits

#### Debugging Steps
1. **Check Service Health**
   ```bash
   curl -s http://localhost:8080/health
   ```

2. **Monitor Performance During Test**
   ```bash
   # In another terminal
   watch -n 1 'curl -s http://localhost:8080/api/v1/performance/stats | jq "."'
   ```

3. **Check System Resources**
   ```bash
   # Monitor system resources
   htop
   iostat -x 1
   ```

4. **Review Load Test Results**
   ```bash
   # Check load test output
   cat load_test_results/standard_load_*.json | jq '.'
   ```

## Best Practices

### Performance Monitoring
1. **Set Realistic Targets**: Base targets on business requirements
2. **Monitor Continuously**: Use automated monitoring
3. **Alert on Trends**: Set up alerts for performance degradation
4. **Regular Reviews**: Review performance metrics weekly

### Load Testing
1. **Start Small**: Begin with low load and increase gradually
2. **Test Realistic Scenarios**: Use realistic data and user behavior
3. **Monitor System Resources**: Watch CPU, memory, and network
4. **Test Edge Cases**: Include stress and spike tests
5. **Document Results**: Keep detailed records of test results

### Performance Optimization
1. **Profile First**: Identify bottlenecks before optimizing
2. **Optimize Critical Path**: Focus on high-impact improvements
3. **Test Changes**: Validate optimizations with load tests
4. **Monitor Impact**: Track performance improvements over time

## Conclusion

The Risk Assessment Service includes comprehensive performance monitoring and load testing capabilities designed to ensure it can handle 1000 requests per minute reliably. Regular load testing and performance monitoring are essential for maintaining service quality and identifying issues before they impact users.

For questions or issues with performance monitoring or load testing, please refer to the troubleshooting section or contact the development team.
