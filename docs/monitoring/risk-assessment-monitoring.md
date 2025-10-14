# Risk Assessment Service Monitoring Guide
## Comprehensive Monitoring and Observability

### Overview

This document provides comprehensive guidance on monitoring the Risk Assessment Service in the KYB Platform. It covers metrics collection, alerting, dashboards, log analysis, and performance monitoring to ensure optimal service health and performance.

---

## Table of Contents

1. [Monitoring Architecture](#monitoring-architecture)
2. [Metrics Collection](#metrics-collection)
3. [Prometheus Integration](#prometheus-integration)
4. [Grafana Dashboards](#grafana-dashboards)
5. [Alerting Configuration](#alerting-configuration)
6. [Log Monitoring](#log-monitoring)
7. [Performance Monitoring](#performance-monitoring)
8. [Business Metrics](#business-metrics)
9. [Infrastructure Monitoring](#infrastructure-monitoring)
10. [Security Monitoring](#security-monitoring)
11. [Troubleshooting Monitoring Issues](#troubleshooting-monitoring-issues)
12. [Best Practices](#best-practices)

---

## Monitoring Architecture

### Monitoring Stack

```
┌─────────────────────────────────────────────────────────────┐
│                    Monitoring Stack                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Grafana    │  │ Prometheus  │  │   Fluentd   │        │
│  │ Dashboards  │  │   Metrics   │  │ Log Aggreg. │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Alert     │  │   Jaeger    │  │   ELK       │        │
│  │  Manager    │  │  Tracing    │  │   Stack     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ Risk Assess │  │ PostgreSQL  │  │    Redis    │        │
│  │   Service   │  │  Database   │  │    Cache    │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Metrics**: Risk Assessment Service → Prometheus → Grafana
2. **Logs**: Risk Assessment Service → Fluentd → ELK Stack
3. **Traces**: Risk Assessment Service → Jaeger
4. **Alerts**: Prometheus → Alert Manager → Notification Channels

---

## Metrics Collection

### Service Metrics

#### HTTP Metrics

```go
// HTTP request metrics
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status_code", "service"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path", "service"},
    )
    
    httpActiveConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "http_active_connections",
            Help: "Number of active HTTP connections",
        },
    )
)
```

#### Business Metrics

```go
// Risk assessment business metrics
var (
    riskAssessmentsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "risk_assessments_total",
            Help: "Total number of risk assessments",
        },
        []string{"tenant_id", "country", "industry", "status"},
    )
    
    riskAssessmentDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "risk_assessment_duration_seconds",
            Help: "Risk assessment duration in seconds",
            Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0, 60.0},
        },
        []string{"tenant_id", "country", "industry"},
    )
    
    riskScores = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "risk_scores",
            Help: "Distribution of risk scores",
            Buckets: []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
        },
        []string{"tenant_id", "country", "industry"},
    )
)
```

#### System Metrics

```go
// System resource metrics
var (
    processCPU = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "process_cpu_seconds_total",
            Help: "Total user and system CPU time spent in seconds",
        },
    )
    
    processMemory = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "process_resident_memory_bytes",
            Help: "Resident memory size in bytes",
        },
    )
    
    goGoroutines = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "go_goroutines",
            Help: "Number of goroutines that currently exist",
        },
    )
    
    goGCPause = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "go_gc_duration_seconds",
            Help: "A summary of the pause duration of garbage collection cycles",
        },
    )
)
```

### Database Metrics

```go
// Database connection metrics
var (
    dbConnectionsActive = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_active",
            Help: "Number of active database connections",
        },
    )
    
    dbConnectionsIdle = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_idle",
            Help: "Number of idle database connections",
        },
    )
    
    dbQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "db_query_duration_seconds",
            Help: "Database query duration in seconds",
        },
        []string{"query_type", "table"},
    )
    
    dbQueryErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_query_errors_total",
            Help: "Total number of database query errors",
        },
        []string{"query_type", "error_type"},
    )
)
```

### Redis Metrics

```go
// Redis connection metrics
var (
    redisConnectionsActive = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "redis_connections_active",
            Help: "Number of active Redis connections",
        },
    )
    
    redisMemoryUsage = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "redis_memory_used_bytes",
            Help: "Redis memory usage in bytes",
        },
    )
    
    redisOperations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "redis_operations_total",
            Help: "Total number of Redis operations",
        },
        []string{"operation", "status"},
    )
    
    redisOperationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "redis_operation_duration_seconds",
            Help: "Redis operation duration in seconds",
        },
        []string{"operation"},
    )
)
```

---

## Prometheus Integration

### Scrape Configuration

```yaml
# configs/monitoring/prometheus-scrape-config.yaml
- job_name: 'risk-assessment-service'
  scrape_interval: 15s
  scrape_timeout: 10s
  metrics_path: /metrics
  static_configs:
    - targets: ['risk-assessment-service:8080']
      labels:
        service: 'risk-assessment-service'
        environment: 'production'
        tier: 'application'
  
  # Relabeling rules
  relabel_configs:
    - source_labels: [__address__]
      target_label: instance
      regex: '([^:]+):.*'
      replacement: '${1}'
  
  # Metric relabeling
  metric_relabel_configs:
    - source_labels: [__name__]
      regex: 'go_.*'
      target_label: component
      replacement: 'golang'
    - source_labels: [__name__]
      regex: 'http_.*'
      target_label: component
      replacement: 'http'
    - source_labels: [__name__]
      regex: 'db_.*'
      target_label: component
      replacement: 'database'
    - source_labels: [__name__]
      regex: 'redis_.*'
      target_label: component
      replacement: 'redis'
```

### Recording Rules

```yaml
# configs/monitoring/prometheus-recording-rules.yaml
groups:
  - name: risk-assessment-service-recording-rules
    rules:
      # Request rate
      - record: risk_assessment_service:http_requests_per_second
        expr: rate(http_requests_total{service="risk-assessment-service"}[5m])
      
      # Error rate
      - record: risk_assessment_service:http_error_rate
        expr: rate(http_requests_total{service="risk-assessment-service", status_code=~"5.."}[5m]) / rate(http_requests_total{service="risk-assessment-service"}[5m])
      
      # Response time percentiles
      - record: risk_assessment_service:http_request_duration_p50
        expr: histogram_quantile(0.50, rate(http_request_duration_seconds_bucket{service="risk-assessment-service"}[5m]))
      
      - record: risk_assessment_service:http_request_duration_p95
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{service="risk-assessment-service"}[5m]))
      
      - record: risk_assessment_service:http_request_duration_p99
        expr: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{service="risk-assessment-service"}[5m]))
      
      # Business metrics
      - record: risk_assessment_service:assessments_per_minute
        expr: rate(risk_assessments_total[1m]) * 60
      
      - record: risk_assessment_service:average_risk_score
        expr: avg(risk_scores)
      
      - record: risk_assessment_service:high_risk_assessments_rate
        expr: rate(risk_assessments_total{status="high_risk"}[5m]) / rate(risk_assessments_total[5m])
```

---

## Grafana Dashboards

### Service Overview Dashboard

```json
{
  "dashboard": {
    "title": "Risk Assessment Service Overview",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{service=\"risk-assessment-service\"}[5m])) by (method, path)",
            "legendFormat": "{{method}} {{path}}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{service=\"risk-assessment-service\", status_code=~\"5..\"}[5m])) by (method, path) / sum(rate(http_requests_total{service=\"risk-assessment-service\"}[5m])) by (method, path) * 100",
            "legendFormat": "{{method}} {{path}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{service=\"risk-assessment-service\"}[5m])) by (le, method, path))",
            "legendFormat": "{{method}} {{path}} P99"
          },
          {
            "expr": "histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket{service=\"risk-assessment-service\"}[5m])) by (le, method, path))",
            "legendFormat": "{{method}} {{path}} P50"
          }
        ]
      },
      {
        "title": "CPU Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(process_cpu_seconds_total{service=\"risk-assessment-service\"}[5m])) by (instance)",
            "legendFormat": "CPU Usage"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "process_resident_memory_bytes{service=\"risk-assessment-service\"}",
            "legendFormat": "Memory Usage"
          }
        ]
      }
    ]
  }
}
```

### Business Metrics Dashboard

```json
{
  "dashboard": {
    "title": "Risk Assessment Business Metrics",
    "panels": [
      {
        "title": "Assessments Per Minute",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(risk_assessments_total[1m]) * 60) by (tenant_id)",
            "legendFormat": "{{tenant_id}}"
          }
        ]
      },
      {
        "title": "Risk Score Distribution",
        "type": "histogram",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, sum(rate(risk_scores_bucket[5m])) by (le))",
            "legendFormat": "P50"
          },
          {
            "expr": "histogram_quantile(0.95, sum(rate(risk_scores_bucket[5m])) by (le))",
            "legendFormat": "P95"
          },
          {
            "expr": "histogram_quantile(0.99, sum(rate(risk_scores_bucket[5m])) by (le))",
            "legendFormat": "P99"
          }
        ]
      },
      {
        "title": "High Risk Assessments Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(risk_assessments_total{status=\"high_risk\"}[5m]) / rate(risk_assessments_total[5m]) * 100",
            "legendFormat": "High Risk Rate %"
          }
        ]
      },
      {
        "title": "Assessment Duration",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, sum(rate(risk_assessment_duration_seconds_bucket[5m])) by (le))",
            "legendFormat": "P99 Duration"
          },
          {
            "expr": "histogram_quantile(0.50, sum(rate(risk_assessment_duration_seconds_bucket[5m])) by (le))",
            "legendFormat": "P50 Duration"
          }
        ]
      }
    ]
  }
}
```

### Infrastructure Dashboard

```json
{
  "dashboard": {
    "title": "Risk Assessment Infrastructure",
    "panels": [
      {
        "title": "Database Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "db_connections_active{service=\"risk-assessment-service\"}",
            "legendFormat": "Active Connections"
          },
          {
            "expr": "db_connections_idle{service=\"risk-assessment-service\"}",
            "legendFormat": "Idle Connections"
          }
        ]
      },
      {
        "title": "Database Query Duration",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, sum(rate(db_query_duration_seconds_bucket{service=\"risk-assessment-service\"}[5m])) by (le, query_type))",
            "legendFormat": "{{query_type}} P99"
          }
        ]
      },
      {
        "title": "Redis Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "redis_memory_used_bytes{service=\"risk-assessment-service\"}",
            "legendFormat": "Memory Usage"
          }
        ]
      },
      {
        "title": "Redis Operations",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(redis_operations_total{service=\"risk-assessment-service\"}[5m])) by (operation)",
            "legendFormat": "{{operation}}"
          }
        ]
      }
    ]
  }
}
```

---

## Alerting Configuration

### Critical Alerts

```yaml
# configs/monitoring/alert-rules.yaml
groups:
  - name: risk-assessment-service-critical
    rules:
      # Service down
      - alert: RiskAssessmentServiceDown
        expr: up{service="risk-assessment-service"} == 0
        for: 1m
        labels:
          severity: critical
          tier: service
        annotations:
          summary: "Risk Assessment Service is down"
          description: "The Risk Assessment Service has been down for more than 1 minute."
          runbook_url: "https://docs.kyb-platform.com/runbooks/risk-assessment-service-down"
      
      # High error rate
      - alert: RiskAssessmentServiceHighErrorRate
        expr: sum(rate(http_requests_total{service="risk-assessment-service", status_code=~"5.."}[5m])) / sum(rate(http_requests_total{service="risk-assessment-service"}[5m])) * 100 > 5
        for: 2m
        labels:
          severity: critical
          tier: service
        annotations:
          summary: "High error rate for Risk Assessment Service"
          description: "The Risk Assessment Service is experiencing a high error rate ({{ $value | printf \"%.2f\" }}% > 5%) for more than 2 minutes."
          runbook_url: "https://docs.kyb-platform.com/runbooks/risk-assessment-service-high-error-rate"
      
      # Database connection failures
      - alert: RiskAssessmentServiceDatabaseConnectionFailure
        expr: increase(db_query_errors_total{service="risk-assessment-service", error_type="connection_failure"}[5m]) > 0
        for: 1m
        labels:
          severity: critical
          tier: database
        annotations:
          summary: "Database connection failures for Risk Assessment Service"
          description: "The Risk Assessment Service is experiencing database connection failures."
          runbook_url: "https://docs.kyb-platform.com/runbooks/risk-assessment-service-database-connection-failure"
      
      # Redis connection failures
      - alert: RiskAssessmentServiceRedisConnectionFailure
        expr: increase(redis_operations_total{service="risk-assessment-service", status="error"}[5m]) > 10
        for: 1m
        labels:
          severity: critical
          tier: cache
        annotations:
          summary: "Redis connection failures for Risk Assessment Service"
          description: "The Risk Assessment Service is experiencing Redis connection failures."
          runbook_url: "https://docs.kyb-platform.com/runbooks/risk-assessment-service-redis-connection-failure"
```

### Warning Alerts

```yaml
  - name: risk-assessment-service-warning
    rules:
      # High latency
      - alert: RiskAssessmentServiceHighLatency
        expr: histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{service="risk-assessment-service"}[5m])) by (le)) > 0.5
        for: 5m
        labels:
          severity: warning
          tier: service
        annotations:
          summary: "High latency for Risk Assessment Service"
          description: "The P99 request latency for the Risk Assessment Service is above 500ms ({{ $value | printf \"%.2f\" }}s) for more than 5 minutes."
          runbook_url: "https://docs.kyb-platform.com/runbooks/risk-assessment-service-high-latency"
      
      # High CPU usage
      - alert: RiskAssessmentServiceHighCPUUsage
        expr: sum(rate(process_cpu_seconds_total{service="risk-assessment-service"}[5m])) by (service) > 0.8
        for: 5m
        labels:
          severity: warning
          tier: service
        annotations:
          summary: "High CPU usage for Risk Assessment Service"
          description: "The Risk Assessment Service is experiencing high CPU usage ({{ $value | printf \"%.2f\" }} cores) for more than 5 minutes."
          runbook_url: "https://docs.kyb-platform.com/runbooks/risk-assessment-service-high-cpu-usage"
      
      # High memory usage
      - alert: RiskAssessmentServiceHighMemoryUsage
        expr: process_resident_memory_bytes{service="risk-assessment-service"} > 500 * 1024 * 1024
        for: 5m
        labels:
          severity: warning
          tier: service
        annotations:
          summary: "High memory usage for Risk Assessment Service"
          description: "The Risk Assessment Service is using more than 500MB of memory ({{ $value | humanizeBytes }}) for more than 5 minutes."
          runbook_url: "https://docs.kyb-platform.com/runbooks/risk-assessment-service-high-memory-usage"
      
      # High goroutine count
      - alert: RiskAssessmentServiceHighGoroutineCount
        expr: go_goroutines{service="risk-assessment-service"} > 1000
        for: 5m
        labels:
          severity: warning
          tier: service
        annotations:
          summary: "High goroutine count for Risk Assessment Service"
          description: "The Risk Assessment Service has more than 1000 goroutines ({{ $value }}) for more than 5 minutes."
          runbook_url: "https://docs.kyb-platform.com/runbooks/risk-assessment-service-high-goroutine-count"
```

### Business Alerts

```yaml
  - name: risk-assessment-service-business
    rules:
      # High risk assessment rate
      - alert: RiskAssessmentServiceHighRiskRate
        expr: rate(risk_assessments_total{status="high_risk"}[5m]) / rate(risk_assessments_total[5m]) * 100 > 20
        for: 10m
        labels:
          severity: warning
          tier: business
        annotations:
          summary: "High risk assessment rate"
          description: "The rate of high-risk assessments is above 20% ({{ $value | printf \"%.2f\" }}%) for more than 10 minutes."
          runbook_url: "https://docs.kyb-platform.com/runbooks/risk-assessment-service-high-risk-rate"
      
      # Low assessment throughput
      - alert: RiskAssessmentServiceLowThroughput
        expr: rate(risk_assessments_total[5m]) * 60 < 10
        for: 10m
        labels:
          severity: warning
          tier: business
        annotations:
          summary: "Low assessment throughput"
          description: "The Risk Assessment Service is processing fewer than 10 assessments per minute ({{ $value | printf \"%.2f\" }}) for more than 10 minutes."
          runbook_url: "https://docs.kyb-platform.com/runbooks/risk-assessment-service-low-throughput"
      
      # High assessment duration
      - alert: RiskAssessmentServiceHighAssessmentDuration
        expr: histogram_quantile(0.99, sum(rate(risk_assessment_duration_seconds_bucket[5m])) by (le)) > 10
        for: 5m
        labels:
          severity: warning
          tier: business
        annotations:
          summary: "High assessment duration"
          description: "The P99 assessment duration is above 10 seconds ({{ $value | printf \"%.2f\" }}s) for more than 5 minutes."
          runbook_url: "https://docs.kyb-platform.com/runbooks/risk-assessment-service-high-assessment-duration"
```

---

## Log Monitoring

### Log Aggregation

#### Fluentd Configuration

```yaml
# configs/logging/fluentd-config.yaml
<source>
  @type tail
  @id input_tail_risk_assessment_app
  path /var/log/risk-assessment-service/*.log
  pos_file /var/log/td-agent/risk-assessment-service.log.pos
  tag app.risk-assessment-service
  <parse>
    @type json
    time_key time
    time_format %Y-%m-%dT%H:%M:%S.%fZ
    keep_time_key true
  </parse>
</source>

<filter app.risk-assessment-service>
  @type record_transformer
  @id filter_risk_assessment_service_metadata
  <record>
    kubernetes_container_name "risk-assessment-service"
    kubernetes_pod_name "risk-assessment-service-pod"
    kubernetes_namespace_name "default"
    service_name "risk-assessment-service"
    log_source "application"
  </record>
</filter>

<match app.risk-assessment-service>
  @type elasticsearch
  @id output_elasticsearch_risk_assessment
  host elasticsearch
  port 9200
  logstash_format true
  logstash_prefix fluentd-risk-assessment
  include_tag_key true
  tag_key @log_name
  <buffer>
    @type file
    path /var/log/td-agent/buffer/risk-assessment.buffer
    flush_interval 5s
    chunk_limit_size 2MB
    queue_limit_length 8
    retry_max_interval 30s
    retry_forever true
  </buffer>
</match>
```

### Log Analysis

#### Error Analysis

```bash
# Count errors by type
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"term": {"level": "error"}}
      ]
    }
  },
  "aggs": {
    "error_types": {
      "terms": {
        "field": "error_type.keyword",
        "size": 10
      }
    }
  }
}'

# Find most common errors
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"term": {"level": "error"}}
      ]
    }
  },
  "aggs": {
    "error_messages": {
      "terms": {
        "field": "message.keyword",
        "size": 10
      }
    }
  }
}'
```

#### Performance Analysis

```bash
# Find slow requests
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"range": {"duration": {"gte": 1000}}}
      ]
    }
  },
  "sort": [
    {"duration": {"order": "desc"}}
  ],
  "size": 10
}'

# Find high memory usage
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"range": {"memory_usage": {"gte": 100000000}}}
      ]
    }
  },
  "sort": [
    {"memory_usage": {"order": "desc"}}
  ],
  "size": 10
}'
```

---

## Performance Monitoring

### Key Performance Indicators (KPIs)

| KPI | Target | Current | Status |
|-----|--------|---------|--------|
| Response Time (P50) | < 100ms | 85ms | ✅ |
| Response Time (P99) | < 500ms | 420ms | ✅ |
| Throughput | > 1000 req/min | 1200 req/min | ✅ |
| Error Rate | < 1% | 0.2% | ✅ |
| Availability | > 99.9% | 99.95% | ✅ |
| Assessment Duration (P99) | < 10s | 8.5s | ✅ |
| High Risk Rate | < 20% | 15% | ✅ |

### Performance Monitoring Commands

#### Response Time Analysis

```bash
# Check response time percentiles
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'http_request_duration_seconds_bucket'

# Monitor response times in real-time
watch -n 5 'curl -s https://risk-assessment-service-production.up.railway.app/metrics | grep http_request_duration_seconds'
```

#### Throughput Analysis

```bash
# Check request rate
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'http_requests_total'

# Monitor throughput in real-time
watch -n 5 'curl -s https://risk-assessment-service-production.up.railway.app/metrics | grep rate'
```

#### Resource Usage

```bash
# Check CPU usage
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'process_cpu_seconds_total'

# Check memory usage
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'process_resident_memory_bytes'

# Check goroutine count
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'go_goroutines'
```

---

## Business Metrics

### Risk Assessment Metrics

#### Assessment Volume

```bash
# Total assessments per day
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"term": {"event_type": "risk_assessment_completed"}},
        {"range": {"@timestamp": {"gte": "now-1d"}}}
      ]
    }
  },
  "aggs": {
    "assessments_per_hour": {
      "date_histogram": {
        "field": "@timestamp",
        "calendar_interval": "hour"
      }
    }
  }
}'
```

#### Risk Score Distribution

```bash
# Risk score distribution
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"term": {"event_type": "risk_assessment_completed"}},
        {"range": {"@timestamp": {"gte": "now-1d"}}}
      ]
    }
  },
  "aggs": {
    "risk_score_distribution": {
      "histogram": {
        "field": "risk_score",
        "interval": 0.1,
        "min_doc_count": 1
      }
    }
  }
}'
```

#### Country-wise Analysis

```bash
# Assessments by country
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"term": {"event_type": "risk_assessment_completed"}},
        {"range": {"@timestamp": {"gte": "now-1d"}}}
      ]
    }
  },
  "aggs": {
    "assessments_by_country": {
      "terms": {
        "field": "country.keyword",
        "size": 20
      }
    }
  }
}'
```

### ML Model Performance

#### Model Accuracy

```bash
# Model accuracy over time
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"term": {"event_type": "ml_prediction_completed"}},
        {"range": {"@timestamp": {"gte": "now-7d"}}}
      ]
    }
  },
  "aggs": {
    "accuracy_over_time": {
      "date_histogram": {
        "field": "@timestamp",
        "calendar_interval": "day"
      },
      "aggs": {
        "average_accuracy": {
          "avg": {
            "field": "model_accuracy"
          }
        }
      }
    }
  }
}'
```

#### Model Performance by Type

```bash
# Model performance by type
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"term": {"event_type": "ml_prediction_completed"}},
        {"range": {"@timestamp": {"gte": "now-7d"}}}
      ]
    }
  },
  "aggs": {
    "performance_by_model": {
      "terms": {
        "field": "model_type.keyword",
        "size": 10
      },
      "aggs": {
        "average_accuracy": {
          "avg": {
            "field": "model_accuracy"
          }
        },
        "average_duration": {
          "avg": {
            "field": "prediction_duration"
          }
        }
      }
    }
  }
}'
```

---

## Infrastructure Monitoring

### Database Monitoring

#### Connection Pool Monitoring

```bash
# Check database connection pool status
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'db_connections'

# Monitor database query performance
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'db_query_duration_seconds'
```

#### Database Performance

```sql
-- Check slow queries
SELECT 
  query,
  calls,
  total_time,
  mean_time,
  rows
FROM pg_stat_statements 
WHERE mean_time > 100  -- queries taking more than 100ms
ORDER BY mean_time DESC;

-- Check index usage
SELECT 
  schemaname,
  tablename,
  indexname,
  idx_scan,
  idx_tup_read,
  idx_tup_fetch
FROM pg_stat_user_indexes 
ORDER BY idx_scan DESC;

-- Check table sizes
SELECT 
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Redis Monitoring

#### Redis Performance

```bash
# Check Redis memory usage
redis-cli -u $REDIS_URL info memory

# Check Redis key count
redis-cli -u $REDIS_URL dbsize

# Check Redis slow log
redis-cli -u $REDIS_URL slowlog get 10

# Monitor Redis commands
redis-cli -u $REDIS_URL monitor
```

#### Redis Metrics

```bash
# Check Redis metrics from service
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'redis_'

# Monitor Redis operations
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'redis_operations_total'
```

---

## Security Monitoring

### Authentication Monitoring

#### Failed Authentication Attempts

```bash
# Count authentication failures
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"term": {"event_type": "authentication_failed"}},
        {"range": {"@timestamp": {"gte": "now-1h"}}}
      ]
    }
  },
  "aggs": {
    "failures_by_ip": {
      "terms": {
        "field": "remote_addr.keyword",
        "size": 10
      }
    }
  }
}'
```

#### Rate Limiting Triggers

```bash
# Count rate limiting triggers
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"term": {"event_type": "rate_limit_exceeded"}},
        {"range": {"@timestamp": {"gte": "now-1h"}}}
      ]
    }
  },
  "aggs": {
    "rate_limits_by_ip": {
      "terms": {
        "field": "remote_addr.keyword",
        "size": 10
      }
    }
  }
}'
```

### Suspicious Activity Detection

#### Unusual Request Patterns

```bash
# Find unusual request patterns
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"term": {"event_type": "http_request"}},
        {"range": {"@timestamp": {"gte": "now-1h"}}}
      ]
    }
  },
  "aggs": {
    "requests_by_ip": {
      "terms": {
        "field": "remote_addr.keyword",
        "size": 20
      },
      "aggs": {
        "request_count": {
          "value_count": {
            "field": "request_id"
          }
        }
      }
    }
  }
}'
```

#### High-Risk Assessment Patterns

```bash
# Find high-risk assessment patterns
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"service_name": "risk-assessment-service"}},
        {"term": {"event_type": "risk_assessment_completed"}},
        {"range": {"risk_score": {"gte": 0.8}}},
        {"range": {"@timestamp": {"gte": "now-1h"}}}
      ]
    }
  },
  "aggs": {
    "high_risk_by_country": {
      "terms": {
        "field": "country.keyword",
        "size": 10
      }
    }
  }
}'
```

---

## Troubleshooting Monitoring Issues

### Common Monitoring Issues

#### 1. Metrics Not Appearing

**Symptoms**: Metrics not showing up in Prometheus/Grafana
**Causes**:
- Service not exposing metrics endpoint
- Prometheus scrape configuration issues
- Network connectivity problems

**Solutions**:
```bash
# Check if service is exposing metrics
curl -f https://risk-assessment-service-production.up.railway.app/metrics

# Check Prometheus scrape configuration
curl -X GET "prometheus:9090/api/v1/targets"

# Check Prometheus logs
docker logs prometheus

# Verify network connectivity
telnet risk-assessment-service 8080
```

#### 2. Alerts Not Firing

**Symptoms**: Alerts not triggering when conditions are met
**Causes**:
- Alert rules misconfigured
- Alert manager not running
- Notification channels not configured

**Solutions**:
```bash
# Check alert rules
curl -X GET "prometheus:9090/api/v1/rules"

# Check alert manager status
curl -X GET "alertmanager:9093/api/v1/status"

# Check alert manager configuration
curl -X GET "alertmanager:9093/api/v1/alerts"

# Test alert rules
curl -X POST "prometheus:9090/api/v1/query" -d 'query=up{service="risk-assessment-service"}'
```

#### 3. Dashboard Not Loading

**Symptoms**: Grafana dashboards not loading or showing errors
**Causes**:
- Data source configuration issues
- Query syntax errors
- Network connectivity problems

**Solutions**:
```bash
# Check Grafana data sources
curl -X GET "grafana:3000/api/datasources"

# Check Grafana logs
docker logs grafana

# Test data source connectivity
curl -X GET "grafana:3000/api/datasources/proxy/1/api/v1/query?query=up"

# Verify Prometheus connectivity
curl -X GET "prometheus:9090/api/v1/query?query=up"
```

#### 4. Logs Not Appearing

**Symptoms**: Logs not showing up in ELK stack
**Causes**:
- Fluentd configuration issues
- Elasticsearch connectivity problems
- Log format issues

**Solutions**:
```bash
# Check Fluentd status
docker logs fluentd

# Check Elasticsearch status
curl -X GET "elasticsearch:9200/_cluster/health"

# Check log format
tail -f /var/log/risk-assessment-service/app.log

# Test Elasticsearch connectivity
curl -X GET "elasticsearch:9200/_cat/indices"
```

### Debug Commands

#### Prometheus Debug

```bash
# Check Prometheus targets
curl -X GET "prometheus:9090/api/v1/targets"

# Check Prometheus rules
curl -X GET "prometheus:9090/api/v1/rules"

# Check Prometheus configuration
curl -X GET "prometheus:9090/api/v1/status/config"

# Test query
curl -X POST "prometheus:9090/api/v1/query" -d 'query=up{service="risk-assessment-service"}'
```

#### Grafana Debug

```bash
# Check Grafana data sources
curl -X GET "grafana:3000/api/datasources"

# Check Grafana dashboards
curl -X GET "grafana:3000/api/search"

# Check Grafana users
curl -X GET "grafana:3000/api/users"

# Test data source
curl -X GET "grafana:3000/api/datasources/proxy/1/api/v1/query?query=up"
```

#### Elasticsearch Debug

```bash
# Check cluster health
curl -X GET "elasticsearch:9200/_cluster/health"

# Check indices
curl -X GET "elasticsearch:9200/_cat/indices"

# Check mappings
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_mapping"

# Test search
curl -X GET "elasticsearch:9200/fluentd-risk-assessment-*/_search?size=1"
```

---

## Best Practices

### 1. Metrics Design

- Use consistent naming conventions
- Include relevant labels for filtering and aggregation
- Avoid high-cardinality labels
- Use appropriate metric types (counter, gauge, histogram, summary)

### 2. Alert Design

- Set appropriate thresholds based on historical data
- Use different severity levels (critical, warning, info)
- Include runbook URLs in alert annotations
- Test alerts regularly

### 3. Dashboard Design

- Keep dashboards focused and relevant
- Use appropriate visualization types
- Include time ranges and refresh intervals
- Organize panels logically

### 4. Log Management

- Use structured logging (JSON format)
- Include correlation IDs for tracing
- Set appropriate log levels
- Implement log rotation and retention

### 5. Performance Monitoring

- Monitor key performance indicators (KPIs)
- Set up performance baselines
- Track trends over time
- Alert on performance degradation

### 6. Security Monitoring

- Monitor authentication and authorization events
- Track suspicious activity patterns
- Implement rate limiting and DDoS protection
- Regular security audits

### 7. Capacity Planning

- Monitor resource usage trends
- Plan for growth and scaling
- Set up capacity alerts
- Regular capacity reviews

---

## Contact Information

### Monitoring Team

| Role | Name | Email | Phone |
|------|------|-------|-------|
| Monitoring Lead | John Doe | john.doe@kyb-platform.com | +1-XXX-XXX-XXXX |
| DevOps Engineer | Jane Smith | jane.smith@kyb-platform.com | +1-XXX-XXX-XXXX |
| Data Engineer | Bob Johnson | bob.johnson@kyb-platform.com | +1-XXX-XXX-XXXX |

### Emergency Contacts

| Role | Name | Email | Phone |
|------|------|-------|-------|
| On-Call Engineer | - | oncall@kyb-platform.com | +1-XXX-XXX-XXXX |
| Platform Manager | - | platform-manager@kyb-platform.com | +1-XXX-XXX-XXXX |

---

**Document Version**: 1.0.0  
**Last Updated**: December 2024  
**Next Review**: March 2025
