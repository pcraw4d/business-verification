# KYB Platform - Application Monitoring

## Overview

The KYB Platform implements comprehensive application monitoring using Prometheus, Grafana, and custom metrics collection. This system provides real-time visibility into application performance, business metrics, and system health.

## Monitoring Architecture

### Monitoring Stack

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   KYB Platform  │    │    Prometheus   │    │     Grafana     │
│      API        │───▶│   (Metrics DB)  │───▶│   (Dashboards)  │
│                 │    │                 │    │                 │
│ • Custom Metrics│    │ • Time Series   │    │ • Visualization │
│ • Health Checks │    │ • Alerting      │    │ • Alerting      │
│ • Business Data │    │ • Recording     │    │ • Reporting     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AlertManager  │    │   Node Exporter │    │  Postgres Exporter│
│                 │    │                 │    │                 │
│ • Alert Routing │    │ • System Metrics│    │ • DB Metrics    │
│ • Notification  │    │ • Hardware      │    │ • Performance   │
│ • Deduplication │    │ • Resources     │    │ • Connections   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Metrics Flow

1. **Application Metrics**: KYB Platform exposes metrics via `/metrics` endpoint
2. **System Metrics**: Node exporter collects system-level metrics
3. **Database Metrics**: Postgres exporter monitors database performance
4. **Prometheus Scraping**: Prometheus collects all metrics at regular intervals
5. **Alerting**: Prometheus evaluates alert rules and sends to AlertManager
6. **Visualization**: Grafana queries Prometheus for dashboard displays

## Monitoring Components

### 1. Application Monitoring System

**Location**: `internal/observability/monitoring.go`

**Features**:
- HTTP request metrics (count, duration, status codes)
- Business metrics (classification, risk assessment, compliance)
- System metrics (memory, CPU, goroutines)
- Database metrics (connections, query duration, errors)
- External API metrics (calls, duration, errors)
- Health check metrics (status, duration)
- Custom business metrics (users, API keys, rate limits)

**Key Metrics**:
```go
// HTTP Metrics
kyb_http_requests_total{method, endpoint, status_code, environment}
kyb_http_request_duration_seconds{method, endpoint, environment}
kyb_http_requests_in_flight{method, endpoint, environment}

// Business Metrics
kyb_classification_requests_total{method, confidence_level, environment}
kyb_classification_accuracy{method, environment}
kyb_risk_assessment_requests_total{risk_level, environment}
kyb_compliance_check_requests_total{framework, status, environment}

// System Metrics
kyb_system_memory_usage_bytes{type, environment}
kyb_system_goroutines{environment}
kyb_system_heap_alloc_bytes{environment}

// Database Metrics
kyb_database_connections{database, environment}
kyb_database_query_duration_seconds{database, query_type, environment}
kyb_database_errors_total{database, error_type, environment}

// External API Metrics
kyb_external_api_calls_total{provider, endpoint, status, environment}
kyb_external_api_duration_seconds{provider, endpoint, environment}
kyb_external_api_errors_total{provider, endpoint, error_type, environment}

// Health Check Metrics
kyb_health_check_status{component, environment}
kyb_health_check_duration_seconds{component, environment}

// Custom Business Metrics
kyb_active_users{user_type, environment}
kyb_api_key_usage_total{api_key_id, endpoint, environment}
kyb_rate_limit_hits_total{api_key_id, endpoint, environment}
kyb_authentication_events_total{event_type, status, environment}
```

### 2. Monitoring Middleware

**Location**: `internal/api/middleware/monitoring.go`

**Features**:
- Automatic HTTP request monitoring
- Business metrics recording
- API key usage tracking
- Rate limit monitoring
- Response time measurement

**Integration**:
```go
// Initialize monitoring middleware
monitoringMiddleware := middleware.NewMonitoringMiddleware(
    monitoringSystem,
    logger,
    environment,
)

// Apply to HTTP handlers
router.Use(monitoringMiddleware.MonitorHTTPRequests)
router.Use(monitoringMiddleware.MonitorAPIKeyUsage)
```

### 3. Prometheus Configuration

**Location**: `deployments/prometheus/`

**Files**:
- `prometheus.yml` - Main Prometheus configuration
- `alerts.yml` - Alerting rules
- `recording_rules.yml` - Pre-computed metrics

**Key Features**:
- Multi-target scraping (API, database, system)
- Kubernetes service discovery
- Metric relabeling and filtering
- Recording rules for performance
- Comprehensive alerting

## Metrics Categories

### 1. HTTP Request Metrics

**Purpose**: Monitor API performance and usage

**Key Metrics**:
- Request count by endpoint, method, status code
- Response time percentiles (p50, p95, p99)
- Request rate (requests per second)
- Error rate (4xx, 5xx responses)
- In-flight requests

**Example Queries**:
```promql
# Request rate by endpoint
rate(kyb_http_requests_total[5m]) by (endpoint)

# 95th percentile response time
histogram_quantile(0.95, rate(kyb_http_request_duration_seconds_bucket[5m]))

# Error rate
rate(kyb_http_requests_total{status_code=~"5.."}[5m]) / rate(kyb_http_requests_total[5m])
```

### 2. Business Metrics

**Purpose**: Monitor core business functionality

**Key Metrics**:
- Classification request volume and accuracy
- Risk assessment requests and performance
- Compliance check requests and status
- Authentication events and success rates
- API key usage patterns

**Example Queries**:
```promql
# Classification accuracy
kyb_classification_accuracy

# Risk assessment rate
rate(kyb_risk_assessment_requests_total[5m])

# Authentication success rate
rate(kyb_authentication_events_total{status="success"}[5m]) / rate(kyb_authentication_events_total[5m])
```

### 3. System Metrics

**Purpose**: Monitor application health and resources

**Key Metrics**:
- Memory usage (allocated, system, heap)
- CPU usage percentage
- Goroutine count
- Heap allocation and system memory
- Runtime statistics

**Example Queries**:
```promql
# Memory usage percentage
kyb_system_memory_usage_bytes{type="alloc"} / kyb_system_memory_usage_bytes{type="sys"}

# Goroutine count
kyb_system_goroutines

# Heap allocation
kyb_system_heap_alloc_bytes
```

### 4. Database Metrics

**Purpose**: Monitor database performance and health

**Key Metrics**:
- Active connection count
- Query duration by type
- Error count by type
- Connection pool utilization
- Query performance trends

**Example Queries**:
```promql
# Database connections
kyb_database_connections

# Query duration 95th percentile
histogram_quantile(0.95, rate(kyb_database_query_duration_seconds_bucket[5m]))

# Database error rate
rate(kyb_database_errors_total[5m])
```

### 5. External API Metrics

**Purpose**: Monitor external service dependencies

**Key Metrics**:
- API call count by provider and endpoint
- Response time by provider
- Error count by type
- Success rate by provider
- Dependency health

**Example Queries**:
```promql
# External API success rate
rate(kyb_external_api_calls_total{status="success"}[5m]) / rate(kyb_external_api_calls_total[5m])

# External API error rate
rate(kyb_external_api_errors_total[5m])

# API response time
histogram_quantile(0.95, rate(kyb_external_api_duration_seconds_bucket[5m]))
```

## Alerting Rules

### Alert Categories

1. **Critical Alerts** (Immediate action required)
   - Service unavailable
   - High error rates (>5%)
   - Critical response times (>5s)
   - Database connection issues
   - Health check failures

2. **Warning Alerts** (Monitor and investigate)
   - High response times (>1s)
   - High memory usage (>80%)
   - High goroutine count (>1000)
   - Authentication failures
   - Rate limit hits

3. **Business Alerts** (Business impact)
   - Low classification accuracy (<90%)
   - No active users
   - High request volume
   - External API issues

### Alert Configuration

**Alert Severity Levels**:
- `critical` - Immediate action required
- `warning` - Monitor and investigate
- `info` - Informational alerts

**Alert Routing**:
- Critical alerts → PagerDuty/Slack + Email
- Warning alerts → Slack + Email
- Info alerts → Slack only

**Alert Deduplication**:
- Group similar alerts
- Suppress repeated alerts
- Escalate unresolved alerts

## Monitoring Dashboards

### Dashboard Categories

1. **Overview Dashboard**
   - System health summary
   - Key performance indicators
   - Recent alerts
   - Service status

2. **Performance Dashboard**
   - Response time trends
   - Throughput metrics
   - Error rates
   - Resource utilization

3. **Business Dashboard**
   - Classification metrics
   - Risk assessment data
   - Compliance status
   - User activity

4. **Infrastructure Dashboard**
   - System resources
   - Database performance
   - External API health
   - Network metrics

5. **Security Dashboard**
   - Authentication events
   - API key usage
   - Rate limit hits
   - Security incidents

### Dashboard Metrics

**Key Performance Indicators**:
- Request rate (RPS)
- Response time (p95, p99)
- Error rate (4xx, 5xx)
- Availability (uptime)
- Throughput (requests/minute)

**Business Metrics**:
- Classification accuracy
- Risk assessment volume
- Compliance check status
- Active users
- API usage patterns

**System Metrics**:
- CPU utilization
- Memory usage
- Disk I/O
- Network traffic
- Goroutine count

## Health Checks

### Health Check Components

1. **Application Health**
   - HTTP server status
   - Database connectivity
   - External API connectivity
   - System resources

2. **Business Health**
   - Classification service
   - Risk assessment service
   - Compliance service
   - Authentication service

3. **Infrastructure Health**
   - Database connections
   - Redis connectivity
   - Load balancer status
   - Network connectivity

### Health Check Endpoints

```http
GET /health
GET /health/detailed
GET /ready
GET /live
```

**Response Format**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "components": {
    "database": "healthy",
    "external_api": "healthy",
    "system": "healthy"
  },
  "metrics": {
    "uptime": "24h30m15s",
    "memory_usage": "45%",
    "goroutines": 150
  }
}
```

## Monitoring Best Practices

### 1. Metric Design

**Naming Conventions**:
- Use descriptive names
- Include units in metric names
- Use consistent naming patterns
- Avoid abbreviations

**Labeling Strategy**:
- Use relevant labels for filtering
- Avoid high cardinality labels
- Include environment labels
- Use consistent label names

**Metric Types**:
- Counters for cumulative values
- Gauges for current values
- Histograms for distributions
- Summaries for quantiles

### 2. Alerting Strategy

**Alert Thresholds**:
- Set realistic thresholds
- Use different thresholds for different environments
- Consider business impact
- Review and adjust regularly

**Alert Grouping**:
- Group related alerts
- Use consistent severity levels
- Include runbook links
- Provide actionable descriptions

**Alert Lifecycle**:
- Acknowledge alerts promptly
- Escalate unresolved alerts
- Document resolutions
- Review alert effectiveness

### 3. Performance Optimization

**Scraping Configuration**:
- Optimize scrape intervals
- Use metric filtering
- Implement metric relabeling
- Monitor scrape performance

**Storage Optimization**:
- Configure retention policies
- Use recording rules
- Implement metric aggregation
- Monitor storage usage

**Query Optimization**:
- Use recording rules for common queries
- Optimize dashboard queries
- Implement query caching
- Monitor query performance

### 4. Security Considerations

**Access Control**:
- Implement authentication
- Use role-based access
- Monitor access patterns
- Audit configuration changes

**Data Protection**:
- Encrypt sensitive metrics
- Implement data retention
- Monitor data access
- Comply with regulations

**Network Security**:
- Use TLS for communication
- Implement network policies
- Monitor network traffic
- Secure API endpoints

## Troubleshooting

### Common Issues

1. **High Memory Usage**
   - Check for memory leaks
   - Review goroutine count
   - Monitor heap allocation
   - Investigate large objects

2. **High Response Times**
   - Check database performance
   - Monitor external API calls
   - Review application logic
   - Investigate resource contention

3. **High Error Rates**
   - Check application logs
   - Monitor external dependencies
   - Review error patterns
   - Investigate recent changes

4. **Metric Collection Issues**
   - Check Prometheus configuration
   - Verify target accessibility
   - Review scrape intervals
   - Monitor Prometheus performance

### Debugging Commands

```bash
# Check Prometheus targets
curl -s http://prometheus:9090/api/v1/targets | jq

# Query metrics directly
curl -s "http://prometheus:9090/api/v1/query?query=up" | jq

# Check alert rules
curl -s http://prometheus:9090/api/v1/rules | jq

# Check metric metadata
curl -s "http://prometheus:9090/api/v1/metadata?metric=kyb_http_requests_total" | jq

# Check target health
curl -s http://kyb-platform-api:8080/metrics | head -20
```

## Future Enhancements

### Planned Improvements

1. **Advanced Analytics**
   - Machine learning-based anomaly detection
   - Predictive performance modeling
   - Business intelligence integration
   - Custom metric aggregation

2. **Enhanced Alerting**
   - Dynamic threshold adjustment
   - Context-aware alerting
   - Automated remediation
   - Alert correlation

3. **Performance Optimization**
   - Metric compression
   - Query optimization
   - Storage optimization
   - Caching strategies

4. **Integration Enhancements**
   - Additional data sources
   - Third-party integrations
   - Custom dashboards
   - API enhancements

---

This documentation provides a comprehensive overview of the KYB Platform's application monitoring system. For specific implementation details, refer to the monitoring code and configuration files referenced throughout this document.
