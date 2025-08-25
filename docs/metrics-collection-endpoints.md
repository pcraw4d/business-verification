# Metrics Collection Endpoints

## Overview

The Enhanced Business Intelligence System provides comprehensive metrics collection endpoints that enable real-time monitoring, performance tracking, and business intelligence. These endpoints collect and expose system metrics, API performance data, business-specific metrics, and resource utilization information.

## API Reference

### Base URL
```
https://api.kyb-platform.com/v3
```

### Authentication
All metrics endpoints require authentication using either:
- API Key: `Authorization: Bearer <api-key>`
- JWT Token: `Authorization: Bearer <jwt-token>`

### Rate Limiting
- **Comprehensive Metrics**: 60 requests per minute
- **Prometheus Metrics**: 120 requests per minute
- **System Metrics**: 60 requests per minute
- **API Metrics**: 60 requests per minute
- **Business Metrics**: 60 requests per minute

## Endpoints

### 1. Comprehensive Metrics

**Endpoint**: `GET /metrics/comprehensive`

**Description**: Returns comprehensive system metrics including system, API, business, performance, resource, and error metrics.

**Response Format**: JSON

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/metrics/comprehensive" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json"
```

**Example Response**:
```json
{
  "timestamp": "2024-12-19T10:30:00Z",
  "version": "1.0.0",
  "environment": "production",
  "system_metrics": {
    "uptime": "2h30m15s",
    "start_time": "2024-12-19T08:00:00Z",
    "process_id": 12345,
    "go_version": "go1.22.0",
    "num_cpu": 8,
    "num_goroutines": 45,
    "num_cgo_call": 0,
    "memory_stats": {
      "alloc_bytes": 52428800,
      "total_alloc_bytes": 104857600,
      "sys_bytes": 67108864,
      "heap_alloc_bytes": 52428800,
      "heap_sys_bytes": 67108864,
      "heap_idle_bytes": 14680064,
      "heap_inuse_bytes": 52428800,
      "heap_released_bytes": 0,
      "heap_objects": 1000000,
      "stack_inuse_bytes": 2097152,
      "stack_sys_bytes": 2097152,
      "mspan_inuse_bytes": 8192,
      "mspan_sys_bytes": 16384,
      "mcache_inuse_bytes": 6912,
      "mcache_sys_bytes": 16384,
      "buck_hash_sys_bytes": 1441792,
      "gc_sys_bytes": 4194304,
      "other_sys_bytes": 1048576,
      "next_gc_bytes": 67108864,
      "last_gc_nanoseconds": 1702989000000000000,
      "pause_total_nanoseconds": 5000000,
      "pause_nanoseconds": [1000000, 2000000, 2000000],
      "pause_end_nanoseconds": [1702989000000000000, 1702989000000000000, 1702989000000000000],
      "num_gc": 3,
      "num_forced_gc": 0,
      "gc_cpu_fraction": 0.001,
      "enable_gc": true,
      "debug_gc": false
    },
    "gc_stats": {
      "num_gc": 3,
      "pause_total_nanoseconds": 5000000,
      "pause_nanoseconds": [1000000, 2000000, 2000000],
      "pause_end_nanoseconds": [1702989000000000000, 1702989000000000000, 1702989000000000000],
      "gc_cpu_fraction": 0.001,
      "last_gc_nanoseconds": 1702989000000000000,
      "next_gc_bytes": 67108864
    }
  },
  "api_metrics": {
    "total_requests": 1000,
    "requests_per_second": 50.5,
    "average_response_time": "150ms",
    "response_time_p95": "300ms",
    "response_time_p99": "500ms",
    "active_requests": 5,
    "requests_by_method": {
      "GET": 600,
      "POST": 300,
      "PUT": 50,
      "DELETE": 50
    },
    "requests_by_endpoint": {
      "/health": 200,
      "/metrics": 100,
      "/verify": 400,
      "/classify": 200,
      "/risk": 100
    },
    "requests_by_status": {
      "200": 950,
      "400": 30,
      "401": 10,
      "500": 10
    },
    "errors_by_type": {
      "validation": 25,
      "authentication": 10,
      "authorization": 5,
      "internal": 10
    },
    "rate_limit_hits": 5,
    "authentication_failures": 10,
    "authorization_failures": 5
  },
  "business_metrics": {
    "total_verifications": 5000,
    "verifications_per_second": 25.3,
    "average_verification_time": "2s",
    "verifications_by_status": {
      "pending": 100,
      "completed": 4800,
      "failed": 100
    },
    "verifications_by_type": {
      "basic": 3000,
      "advanced": 1500,
      "premium": 500
    },
    "verifications_by_industry": {
      "technology": 1500,
      "finance": 1200,
      "healthcare": 800,
      "retail": 1000,
      "manufacturing": 500
    },
    "success_rate": 96.0,
    "error_rate": 4.0,
    "average_confidence_score": 0.85,
    "cache_hit_rate": 75.0,
    "cache_miss_rate": 25.0,
    "external_api_calls": 15000,
    "external_api_latency": "500ms",
    "ml_model_predictions": 8000,
    "ml_model_accuracy": 92.5
  },
  "performance_metrics": {
    "cpu_usage_percent": 45.2,
    "memory_usage_percent": 62.8,
    "disk_usage_percent": 35.5,
    "network_latency": "25ms",
    "network_throughput_mbps": 850.5,
    "database_connections": 15,
    "database_latency": "50ms",
    "cache_connections": 8,
    "cache_latency": "5ms",
    "queue_depth": 12,
    "worker_utilization_percent": 78.5
  },
  "resource_metrics": {
    "open_files": 125,
    "max_files": 1024,
    "file_descriptors": 125,
    "threads": 45,
    "max_threads": 1000,
    "load_average": [1.2, 1.5, 1.8],
    "disk_io_read_bytes": 104857600,
    "disk_io_write_bytes": 52428800,
    "disk_io_read_ops": 1000,
    "disk_io_write_ops": 500,
    "network_bytes_received": 209715200,
    "network_bytes_sent": 157286400,
    "network_packets_received": 50000,
    "network_packets_sent": 40000
  },
  "error_metrics": {
    "total_errors": 50,
    "errors_per_second": 2.5,
    "errors_by_type": {
      "validation": 20,
      "authentication": 10,
      "authorization": 5,
      "internal": 10,
      "external": 5
    },
    "errors_by_endpoint": {
      "/verify": 25,
      "/classify": 15,
      "/risk": 10
    },
    "errors_by_status": {
      "400": 30,
      "401": 10,
      "403": 5,
      "500": 5
    },
    "last_error_time": "2024-12-19T10:25:00Z",
    "last_error_message": "Database connection timeout",
    "error_rate": 5.0,
    "critical_errors": 5,
    "warning_errors": 20,
    "info_errors": 25
  },
  "custom_metrics": {
    "custom_collector": {
      "custom_metric_1": 42,
      "custom_metric_2": "test-value"
    }
  }
}
```

### 2. Prometheus Metrics

**Endpoint**: `GET /metrics/prometheus`

**Description**: Returns metrics in Prometheus format for integration with Prometheus monitoring systems.

**Response Format**: Text/Plain (Prometheus format)

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/metrics/prometheus" \
  -H "Authorization: Bearer your-api-key"
```

**Example Response**:
```
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines{version="1.0.0",environment="production"} 45

# HELP go_threads Number of OS threads created.
# TYPE go_threads gauge
go_threads{version="1.0.0",environment="production"} 8

# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes{version="1.0.0",environment="production"} 52428800

# HELP go_memstats_sys_bytes Number of bytes obtained from system.
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes{version="1.0.0",environment="production"} 67108864

# HELP api_requests_total Total number of API requests.
# TYPE api_requests_total counter
api_requests_total{version="1.0.0",environment="production"} 1000

# HELP api_requests_per_second Number of API requests per second.
# TYPE api_requests_per_second gauge
api_requests_per_second{version="1.0.0",environment="production"} 50.5

# HELP business_verifications_total Total number of business verifications.
# TYPE business_verifications_total counter
business_verifications_total{version="1.0.0",environment="production"} 5000

# HELP business_success_rate Success rate of business verifications.
# TYPE business_success_rate gauge
business_success_rate{version="1.0.0",environment="production"} 96.0

# HELP performance_cpu_usage_percent CPU usage percentage.
# TYPE performance_cpu_usage_percent gauge
performance_cpu_usage_percent{version="1.0.0",environment="production"} 45.2

# HELP performance_memory_usage_percent Memory usage percentage.
# TYPE performance_memory_usage_percent gauge
performance_memory_usage_percent{version="1.0.0",environment="production"} 62.8

# HELP error_total Total number of errors.
# TYPE error_total counter
error_total{version="1.0.0",environment="production"} 50

# HELP error_rate Error rate percentage.
# TYPE error_rate gauge
error_rate{version="1.0.0",environment="production"} 5.0
```

### 3. System Metrics

**Endpoint**: `GET /metrics/system`

**Description**: Returns system-level metrics including memory usage, garbage collection statistics, and runtime information.

**Response Format**: JSON

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/metrics/system" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json"
```

**Example Response**:
```json
{
  "timestamp": "2024-12-19T10:30:00Z",
  "version": "1.0.0",
  "environment": "production",
  "system_metrics": {
    "uptime": "2h30m15s",
    "start_time": "2024-12-19T08:00:00Z",
    "process_id": 12345,
    "go_version": "go1.22.0",
    "num_cpu": 8,
    "num_goroutines": 45,
    "num_cgo_call": 0,
    "memory_stats": {
      "alloc_bytes": 52428800,
      "total_alloc_bytes": 104857600,
      "sys_bytes": 67108864,
      "heap_alloc_bytes": 52428800,
      "heap_sys_bytes": 67108864,
      "heap_idle_bytes": 14680064,
      "heap_inuse_bytes": 52428800,
      "heap_released_bytes": 0,
      "heap_objects": 1000000,
      "stack_inuse_bytes": 2097152,
      "stack_sys_bytes": 2097152,
      "mspan_inuse_bytes": 8192,
      "mspan_sys_bytes": 16384,
      "mcache_inuse_bytes": 6912,
      "mcache_sys_bytes": 16384,
      "buck_hash_sys_bytes": 1441792,
      "gc_sys_bytes": 4194304,
      "other_sys_bytes": 1048576,
      "next_gc_bytes": 67108864,
      "last_gc_nanoseconds": 1702989000000000000,
      "pause_total_nanoseconds": 5000000,
      "pause_nanoseconds": [1000000, 2000000, 2000000],
      "pause_end_nanoseconds": [1702989000000000000, 1702989000000000000, 1702989000000000000],
      "num_gc": 3,
      "num_forced_gc": 0,
      "gc_cpu_fraction": 0.001,
      "enable_gc": true,
      "debug_gc": false
    },
    "gc_stats": {
      "num_gc": 3,
      "pause_total_nanoseconds": 5000000,
      "pause_nanoseconds": [1000000, 2000000, 2000000],
      "pause_end_nanoseconds": [1702989000000000000, 1702989000000000000, 1702989000000000000],
      "gc_cpu_fraction": 0.001,
      "last_gc_nanoseconds": 1702989000000000000,
      "next_gc_bytes": 67108864
    }
  }
}
```

### 4. API Metrics

**Endpoint**: `GET /metrics/api`

**Description**: Returns API-level metrics including request counts, response times, and error rates.

**Response Format**: JSON

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/metrics/api" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json"
```

**Example Response**:
```json
{
  "timestamp": "2024-12-19T10:30:00Z",
  "version": "1.0.0",
  "environment": "production",
  "api_metrics": {
    "total_requests": 1000,
    "requests_per_second": 50.5,
    "average_response_time": "150ms",
    "response_time_p95": "300ms",
    "response_time_p99": "500ms",
    "active_requests": 5,
    "requests_by_method": {
      "GET": 600,
      "POST": 300,
      "PUT": 50,
      "DELETE": 50
    },
    "requests_by_endpoint": {
      "/health": 200,
      "/metrics": 100,
      "/verify": 400,
      "/classify": 200,
      "/risk": 100
    },
    "requests_by_status": {
      "200": 950,
      "400": 30,
      "401": 10,
      "500": 10
    },
    "errors_by_type": {
      "validation": 25,
      "authentication": 10,
      "authorization": 5,
      "internal": 10
    },
    "rate_limit_hits": 5,
    "authentication_failures": 10,
    "authorization_failures": 5
  }
}
```

### 5. Business Metrics

**Endpoint**: `GET /metrics/business`

**Description**: Returns business-specific metrics including verification counts, success rates, and industry breakdowns.

**Response Format**: JSON

**Example Request**:
```bash
curl -X GET "https://api.kyb-platform.com/v3/metrics/business" \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json"
```

**Example Response**:
```json
{
  "timestamp": "2024-12-19T10:30:00Z",
  "version": "1.0.0",
  "environment": "production",
  "business_metrics": {
    "total_verifications": 5000,
    "verifications_per_second": 25.3,
    "average_verification_time": "2s",
    "verifications_by_status": {
      "pending": 100,
      "completed": 4800,
      "failed": 100
    },
    "verifications_by_type": {
      "basic": 3000,
      "advanced": 1500,
      "premium": 500
    },
    "verifications_by_industry": {
      "technology": 1500,
      "finance": 1200,
      "healthcare": 800,
      "retail": 1000,
      "manufacturing": 500
    },
    "success_rate": 96.0,
    "error_rate": 4.0,
    "average_confidence_score": 0.85,
    "cache_hit_rate": 75.0,
    "cache_miss_rate": 25.0,
    "external_api_calls": 15000,
    "external_api_latency": "500ms",
    "ml_model_predictions": 8000,
    "ml_model_accuracy": 92.5
  }
}
```

## Error Responses

### 401 Unauthorized
```json
{
  "error": "unauthorized",
  "message": "Authentication required",
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### 403 Forbidden
```json
{
  "error": "forbidden",
  "message": "Insufficient permissions",
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### 429 Too Many Requests
```json
{
  "error": "rate_limit_exceeded",
  "message": "Rate limit exceeded",
  "retry_after": 60,
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal_error",
  "message": "Internal server error",
  "timestamp": "2024-12-19T10:30:00Z"
}
```

## Integration Examples

### Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'kyb-platform'
    static_configs:
      - targets: ['api.kyb-platform.com:443']
    metrics_path: '/v3/metrics/prometheus'
    scheme: 'https'
    tls_config:
      insecure_skip_verify: false
    authorization:
      type: 'Bearer'
      credentials: 'your-api-key'
    scrape_interval: 30s
    scrape_timeout: 10s
```

### Grafana Dashboard

Create a Grafana dashboard with the following queries:

**System Metrics**:
```
# CPU Usage
rate(performance_cpu_usage_percent[5m])

# Memory Usage
rate(performance_memory_usage_percent[5m])

# Goroutines
go_goroutines

# Memory Allocation
go_memstats_alloc_bytes
```

**API Metrics**:
```
# Request Rate
rate(api_requests_total[5m])

# Error Rate
rate(error_total[5m])

# Response Time
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

**Business Metrics**:
```
# Verification Rate
rate(business_verifications_total[5m])

# Success Rate
business_success_rate

# Cache Hit Rate
business_cache_hit_rate
```

### Custom Metrics Collector

Implement a custom metrics collector:

```go
package main

import (
    "context"
    "time"
)

type CustomMetricsCollector struct {
    name string
}

func (c *CustomMetricsCollector) Collect(ctx context.Context) (map[string]interface{}, error) {
    return map[string]interface{}{
        "custom_metric_1": 42,
        "custom_metric_2": "test-value",
        "custom_timestamp": time.Now().Unix(),
    }, nil
}

func (c *CustomMetricsCollector) Name() string {
    return c.name
}

// Register the collector
handler.RegisterCollector(&CustomMetricsCollector{name: "custom-collector"})
```

## Monitoring and Alerting

### Key Metrics to Monitor

1. **System Health**:
   - CPU usage > 80%
   - Memory usage > 85%
   - Disk usage > 90%
   - Goroutine count > 1000

2. **API Performance**:
   - Response time P95 > 500ms
   - Error rate > 5%
   - Request rate drop > 50%

3. **Business Metrics**:
   - Success rate < 90%
   - Verification rate drop > 30%
   - Cache hit rate < 60%

### Alerting Rules

**Prometheus Alerting Rules**:
```yaml
groups:
  - name: kyb-platform
    rules:
      - alert: HighCPUUsage
        expr: performance_cpu_usage_percent > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"
          description: "CPU usage is {{ $value }}%"

      - alert: HighErrorRate
        expr: rate(error_total[5m]) > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors per second"

      - alert: LowSuccessRate
        expr: business_success_rate < 90
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low success rate detected"
          description: "Success rate is {{ $value }}%"
```

## Best Practices

### 1. Caching
- Metrics are cached for 30 seconds to improve performance
- Use the comprehensive metrics endpoint for dashboard data
- Use specific endpoints for targeted monitoring

### 2. Rate Limiting
- Respect rate limits to avoid service disruption
- Implement exponential backoff for retries
- Use appropriate polling intervals (30s for Prometheus)

### 3. Security
- Use API keys with appropriate permissions
- Rotate API keys regularly
- Monitor access patterns for anomalies

### 4. Performance
- Collect metrics at appropriate intervals
- Use Prometheus format for time-series databases
- Implement custom collectors for business-specific metrics

### 5. Monitoring
- Set up comprehensive alerting
- Monitor both technical and business metrics
- Use dashboards for visualization
- Implement log correlation with metrics

## Troubleshooting

### Common Issues

1. **Authentication Errors**:
   - Verify API key is valid and not expired
   - Check permissions for metrics endpoints
   - Ensure proper Authorization header format

2. **Rate Limiting**:
   - Reduce polling frequency
   - Implement proper retry logic
   - Check rate limit headers in responses

3. **Data Accuracy**:
   - Metrics are cached for 30 seconds
   - Use timestamps for data correlation
   - Verify environment and version tags

4. **Prometheus Integration**:
   - Check TLS configuration
   - Verify metrics format compliance
   - Test scrape configuration

### Debug Information

Enable debug logging to troubleshoot issues:

```bash
curl -X GET "https://api.kyb-platform.com/v3/metrics/comprehensive" \
  -H "Authorization: Bearer your-api-key" \
  -H "X-Debug: true" \
  -v
```

## Migration Guide

### From Basic Metrics

If migrating from basic metrics endpoints:

1. **Update Endpoints**:
   - Replace `/metrics` with `/metrics/comprehensive`
   - Use `/metrics/prometheus` for Prometheus integration
   - Add specific endpoints for targeted monitoring

2. **Update Authentication**:
   - Ensure API keys have metrics permissions
   - Update Authorization headers if needed

3. **Update Dashboards**:
   - Modify queries to use new metric names
   - Add business-specific metrics
   - Update alerting rules

### Version Compatibility

- **v1.0.0**: Initial release with comprehensive metrics
- **v1.1.0**: Added custom metrics collectors
- **v1.2.0**: Enhanced Prometheus format support
- **v2.0.0**: Breaking changes in metric structure (planned)

## Support

For questions or issues with metrics collection:

- **Documentation**: [https://docs.kyb-platform.com/metrics](https://docs.kyb-platform.com/metrics)
- **API Reference**: [https://api.kyb-platform.com/docs](https://api.kyb-platform.com/docs)
- **Support Email**: metrics-support@kyb-platform.com
- **GitHub Issues**: [https://github.com/kyb-platform/issues](https://github.com/kyb-platform/issues)
