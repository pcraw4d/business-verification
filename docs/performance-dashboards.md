# KYB Platform - Performance Dashboards

## Overview

The KYB Platform implements comprehensive performance dashboards using Grafana, Prometheus, and custom dashboard systems. These dashboards provide real-time visibility into system performance, business metrics, security events, and infrastructure health across all application components.

## Dashboard Architecture

### Dashboard Stack

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   KYB Platform  │    │   Prometheus    │    │     Grafana     │
│      API        │───▶│   (Metrics)     │───▶│   (Dashboards)  │
│                 │    │                 │    │                 │
│ • Custom        │    │ • Time Series   │    │ • Visualizations│
│   Dashboards    │    │ • Aggregations  │    │ • Alerts        │
│ • Real-time     │    │ • Recording     │    │ • Annotations   │
│   Metrics       │    │   Rules         │    │ • Templating    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Dashboard     │    │   AlertManager  │    │   Dashboard     │
│   System        │    │                 │    │   Export        │
│                 │    │ • Alert Rules   │    │                 │
│ • Widgets       │    │ • Notifications │    │ • JSON/CSV      │
│ • Layouts       │    │ • Escalations   │    │ • Prometheus    │
│ • Configs       │    │ • Silencing     │    │ • Integration   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Dashboard Flow

1. **Metric Collection**: KYB Platform generates metrics via Prometheus
2. **Metric Storage**: Prometheus stores time-series metrics
3. **Dashboard Queries**: Grafana queries Prometheus for metrics
4. **Visualization**: Grafana renders dashboards with charts and graphs
5. **Alerting**: AlertManager processes alert rules and notifications
6. **Export**: Dashboards can be exported in various formats

## Dashboard Components

### 1. Dashboard System

**Location**: `internal/observability/dashboards.go`

**Features**:
- Comprehensive dashboard metrics collection
- Real-time widget generation
- Alert generation and management
- Multiple dashboard categories
- Export capabilities (JSON, CSV, Prometheus)
- Performance threshold monitoring
- Business metrics tracking
- Security event monitoring
- Infrastructure health monitoring

**Key Components**:
```go
// Dashboard System
type DashboardSystem struct {
    logger         *zap.Logger
    monitoring     *MonitoringSystem
    logAggregation *LogAggregationSystem
    config         *DashboardConfig
}

// Dashboard Metrics
type DashboardMetrics struct {
    Timestamp      time.Time `json:"timestamp"`
    System         struct { /* System metrics */ } `json:"system"`
    Performance    struct { /* Performance metrics */ } `json:"performance"`
    Business       struct { /* Business metrics */ } `json:"business"`
    Security       struct { /* Security metrics */ } `json:"security"`
    Infrastructure struct { /* Infrastructure metrics */ } `json:"infrastructure"`
    Alerts         []DashboardAlert `json:"alerts"`
}

// Dashboard Widget
type DashboardWidget struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    Position    WidgetPosition         `json:"position"`
    Size        WidgetSize             `json:"size"`
    Config      map[string]interface{} `json:"config"`
    Data        interface{}            `json:"data"`
}
```

### 2. Dashboard Categories

**System Dashboard**:
- CPU usage monitoring
- Memory usage tracking
- Goroutine count
- Heap memory allocation
- System uptime

**Performance Dashboard**:
- Request rate monitoring
- Response time percentiles (P50, P95, P99)
- Error rate tracking
- Throughput measurement
- Active connections

**Business Dashboard**:
- Classification requests and accuracy
- Risk assessment metrics
- Compliance check performance
- Active user tracking
- API key usage patterns

**Security Dashboard**:
- Authentication attempts and failures
- Rate limit violations
- Security incidents
- API key abuse detection
- Access control events

**Infrastructure Dashboard**:
- Database connections and performance
- External API calls and latency
- Redis connections and memory
- Network traffic monitoring
- Service health checks

### 3. Grafana Dashboard Configurations

**Performance Dashboard**: `deployments/grafana/dashboards/kyb-performance-dashboard.json`
- 17 comprehensive panels
- Real-time metrics visualization
- Performance threshold alerts
- Business metrics integration
- Infrastructure monitoring

**Business Dashboard**: `deployments/grafana/dashboards/kyb-business-dashboard.json`
- 16 business-focused panels
- Classification accuracy tracking
- Risk assessment monitoring
- Compliance check performance
- User activity analysis

## Dashboard Metrics

### System Metrics

**CPU Usage**:
```json
{
  "metric": "kyb_system_cpu_usage",
  "description": "Current CPU usage percentage",
  "thresholds": {
    "warning": 70,
    "critical": 85
  }
}
```

**Memory Usage**:
```json
{
  "metric": "kyb_system_memory_usage",
  "description": "Current memory usage percentage",
  "thresholds": {
    "warning": 70,
    "critical": 85
  }
}
```

**Goroutines**:
```json
{
  "metric": "kyb_system_goroutines",
  "description": "Number of active goroutines",
  "thresholds": {
    "warning": 500,
    "critical": 1000
  }
}
```

**Heap Memory**:
```json
{
  "metric": "kyb_system_heap_alloc",
  "description": "Heap memory allocation in bytes",
  "thresholds": {
    "warning": "500MB",
    "critical": "800MB"
  }
}
```

### Performance Metrics

**Request Rate**:
```json
{
  "metric": "rate(kyb_http_requests_total[5m])",
  "description": "HTTP requests per second",
  "thresholds": {
    "warning": 500,
    "critical": 1000
  }
}
```

**Response Time**:
```json
{
  "metric": "histogram_quantile(0.95, rate(kyb_http_request_duration_seconds_bucket[5m]))",
  "description": "95th percentile response time",
  "thresholds": {
    "warning": "0.5s",
    "critical": "1.0s"
  }
}
```

**Error Rate**:
```json
{
  "metric": "rate(kyb_http_requests_total{status_code=~\"5..\"}[5m]) / rate(kyb_http_requests_total[5m]) * 100",
  "description": "HTTP error rate percentage",
  "thresholds": {
    "warning": 1,
    "critical": 5
  }
}
```

**Throughput**:
```json
{
  "metric": "rate(kyb_http_requests_total[5m]) * 60",
  "description": "Requests per minute",
  "thresholds": {
    "warning": 30000,
    "critical": 60000
  }
}
```

### Business Metrics

**Classification Requests**:
```json
{
  "metric": "rate(kyb_classification_requests_total[5m])",
  "description": "Business classification requests per second",
  "labels": ["method", "confidence_level"]
}
```

**Classification Accuracy**:
```json
{
  "metric": "kyb_classification_accuracy",
  "description": "Classification accuracy percentage",
  "thresholds": {
    "warning": 0.9,
    "critical": 0.85
  }
}
```

**Risk Assessment Requests**:
```json
{
  "metric": "rate(kyb_risk_assessment_requests_total[5m])",
  "description": "Risk assessment requests per second",
  "labels": ["risk_level"]
}
```

**Compliance Check Requests**:
```json
{
  "metric": "rate(kyb_compliance_check_requests_total[5m])",
  "description": "Compliance check requests per second",
  "labels": ["framework", "status"]
}
```

**Active Users**:
```json
{
  "metric": "kyb_active_users",
  "description": "Number of active users",
  "labels": ["user_type"]
}
```

### Security Metrics

**Authentication Attempts**:
```json
{
  "metric": "rate(kyb_authentication_events_total[5m])",
  "description": "Authentication attempts per second",
  "labels": ["event_type", "success"]
}
```

**Rate Limit Hits**:
```json
{
  "metric": "rate(kyb_rate_limit_hits_total[5m])",
  "description": "Rate limit violations per second",
  "labels": ["api_key_id", "endpoint"]
}
```

**Security Incidents**:
```json
{
  "metric": "kyb_security_incidents_total",
  "description": "Total security incidents",
  "labels": ["incident_type", "severity"]
}
```

### Infrastructure Metrics

**Database Connections**:
```json
{
  "metric": "kyb_database_connections",
  "description": "Active database connections",
  "labels": ["database", "state"]
}
```

**Database Query Duration**:
```json
{
  "metric": "histogram_quantile(0.95, rate(kyb_database_query_duration_seconds_bucket[5m]))",
  "description": "95th percentile database query duration",
  "labels": ["database", "query_type"]
}
```

**External API Calls**:
```json
{
  "metric": "rate(kyb_external_api_calls_total[5m])",
  "description": "External API calls per second",
  "labels": ["provider", "endpoint", "status"]
}
```

**External API Latency**:
```json
{
  "metric": "histogram_quantile(0.95, rate(kyb_external_api_duration_seconds_bucket[5m]))",
  "description": "95th percentile external API latency",
  "labels": ["provider", "endpoint"]
}
```

## Dashboard Configuration

### Dashboard System Configuration

```go
type DashboardConfig struct {
    // Dashboard settings
    RefreshInterval time.Duration
    RetentionPeriod time.Duration
    MaxDataPoints   int
    
    // Performance thresholds
    ResponseTimeThresholds struct {
        Warning  time.Duration
        Critical time.Duration
    }
    ErrorRateThresholds struct {
        Warning  float64
        Critical float64
    }
    ThroughputThresholds struct {
        Warning  int
        Critical int
    }
    
    // Dashboard categories
    EnableSystemDashboard     bool
    EnablePerformanceDashboard bool
    EnableBusinessDashboard   bool
    EnableSecurityDashboard   bool
    EnableInfrastructureDashboard bool
    
    // Export settings
    EnablePrometheusExport bool
    EnableJSONExport       bool
    EnableCSVExport        bool
}
```

### Environment Configuration

**Development Environment**:
```yaml
dashboard:
  refresh_interval: "30s"
  retention_period: "7d"
  max_data_points: 1000
  response_time_thresholds:
    warning: "500ms"
    critical: "1s"
  error_rate_thresholds:
    warning: 1.0
    critical: 5.0
  throughput_thresholds:
    warning: 1000
    critical: 5000
  enable_system_dashboard: true
  enable_performance_dashboard: true
  enable_business_dashboard: true
  enable_security_dashboard: true
  enable_infrastructure_dashboard: true
  enable_prometheus_export: true
  enable_json_export: true
  enable_csv_export: true
```

**Production Environment**:
```yaml
dashboard:
  refresh_interval: "15s"
  retention_period: "30d"
  max_data_points: 5000
  response_time_thresholds:
    warning: "300ms"
    critical: "500ms"
  error_rate_thresholds:
    warning: 0.5
    critical: 2.0
  throughput_thresholds:
    warning: 5000
    critical: 10000
  enable_system_dashboard: true
  enable_performance_dashboard: true
  enable_business_dashboard: true
  enable_security_dashboard: true
  enable_infrastructure_dashboard: true
  enable_prometheus_export: true
  enable_json_export: false
  enable_csv_export: false
```

## Dashboard Widgets

### Widget Types

**Gauge Widgets**:
- CPU usage gauge
- Memory usage gauge
- Classification accuracy gauge
- Success rate gauge
- Active connections gauge

**Line Graph Widgets**:
- Request rate over time
- Response time percentiles
- Error rate trends
- Business metrics trends
- Infrastructure performance

**Stat Widgets**:
- Active users count
- API key usage rate
- Total requests today
- Risk assessment count
- Compliance check count

**Table Widgets**:
- Top endpoints by usage
- Error breakdown by type
- User activity summary
- Security incident list
- Performance bottlenecks

**Pie Chart Widgets**:
- Business events by type
- Error distribution
- User activity breakdown
- API usage by endpoint
- Security event types

**Heatmap Widgets**:
- User activity heatmap
- Error rate heatmap
- Performance heatmap
- Security incident heatmap
- Infrastructure health heatmap

### Widget Configuration

**Gauge Widget Example**:
```json
{
  "id": "cpu-usage",
  "type": "gauge",
  "title": "CPU Usage",
  "description": "Current CPU usage percentage",
  "position": {"x": 0, "y": 0},
  "size": {"width": 4, "height": 3},
  "config": {
    "min": 0,
    "max": 100,
    "unit": "%"
  },
  "data": 45.2
}
```

**Line Graph Widget Example**:
```json
{
  "id": "request-rate",
  "type": "line",
  "title": "Request Rate",
  "description": "Requests per second over time",
  "position": {"x": 0, "y": 0},
  "size": {"width": 12, "height": 8},
  "config": {
    "timeRange": "1h",
    "unit": "req/s"
  },
  "data": [
    {"timestamp": "2024-01-15T10:00:00Z", "value": 150.5},
    {"timestamp": "2024-01-15T10:01:00Z", "value": 165.2}
  ]
}
```

## Dashboard Alerts

### Alert Generation

**System Alerts**:
- High CPU usage (>80%)
- High memory usage (>85%)
- High goroutine count (>1000)
- High heap memory usage (>800MB)

**Performance Alerts**:
- High response time (P95 > 1s)
- High error rate (>5%)
- Low throughput (<1000 req/min)
- High active connections (>100)

**Business Alerts**:
- Low classification accuracy (<90%)
- High risk assessment failures
- Low compliance check success rate
- Unusual user activity patterns

**Security Alerts**:
- High authentication failures (>10/min)
- High rate limit hits (>50/min)
- Security incidents detected
- API key abuse detected

**Infrastructure Alerts**:
- High database error rate (>5/min)
- High external API error rate (>10/min)
- Database connection pool exhaustion
- Redis memory usage high (>80%)

### Alert Configuration

**Alert Rule Example**:
```yaml
groups:
  - name: kyb-performance-alerts
    rules:
      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(kyb_http_request_duration_seconds_bucket[5m])) > 1.0
        for: 2m
        labels:
          severity: warning
          team: platform
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }}s"
          runbook_url: "https://runbook.kybplatform.com/high-response-time"
```

## Dashboard Integration

### API Integration

**Dashboard Endpoints**:
```go
// Get dashboard metrics
GET /v1/dashboard/metrics

// Get system dashboard
GET /v1/dashboard/system

// Get performance dashboard
GET /v1/dashboard/performance

// Get business dashboard
GET /v1/dashboard/business

// Get security dashboard
GET /v1/dashboard/security

// Get infrastructure dashboard
GET /v1/dashboard/infrastructure
```

**Dashboard Handler**:
```go
func (ds *DashboardSystem) DashboardHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        dashboardType := r.URL.Query().Get("type")
        
        var data interface{}
        var err error
        
        switch dashboardType {
        case "metrics":
            data, err = ds.GetDashboardMetrics(ctx)
        case "system":
            data, err = ds.GetSystemDashboard(ctx)
        case "performance":
            data, err = ds.GetPerformanceDashboard(ctx)
        case "business":
            data, err = ds.GetBusinessDashboard(ctx)
        case "security":
            data, err = ds.GetSecurityDashboard(ctx)
        case "infrastructure":
            data, err = ds.GetInfrastructureDashboard(ctx)
        }
        
        // Return JSON response
        json.NewEncoder(w).Encode(data)
    }
}
```

### Grafana Integration

**Dashboard Import**:
```bash
# Import performance dashboard
curl -X POST http://grafana:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $GRAFANA_API_KEY" \
  -d @deployments/grafana/dashboards/kyb-performance-dashboard.json

# Import business dashboard
curl -X POST http://grafana:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $GRAFANA_API_KEY" \
  -d @deployments/grafana/dashboards/kyb-business-dashboard.json
```

**Dashboard Configuration**:
```yaml
grafana:
  dashboards:
    default:
      timezone: browser
      refresh: 30s
      schemaVersion: 16
      version: 1
    providers:
      - name: 'default'
        orgId: 1
        folder: ''
        type: file
        disableDeletion: false
        updateIntervalSeconds: 10
        allowUiUpdates: true
        options:
          path: /var/lib/grafana/dashboards
```

## Dashboard Export

### Export Formats

**JSON Export**:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "dashboard": "performance",
  "metrics": {
    "system": { /* system metrics */ },
    "performance": { /* performance metrics */ },
    "business": { /* business metrics */ },
    "security": { /* security metrics */ },
    "infrastructure": { /* infrastructure metrics */ }
  },
  "alerts": [ /* active alerts */ ]
}
```

**CSV Export**:
```csv
timestamp,metric,value,threshold,status
2024-01-15T10:30:00Z,cpu_usage,45.2,80,warning
2024-01-15T10:30:00Z,memory_usage,65.8,85,normal
2024-01-15T10:30:00Z,response_time_p95,0.8,1.0,warning
```

**Prometheus Export**:
```prometheus
# Dashboard metrics
kyb_dashboard_cpu_usage{environment="production"} 45.2
kyb_dashboard_memory_usage{environment="production"} 65.8
kyb_dashboard_response_time_p95{environment="production"} 0.8
kyb_dashboard_error_rate{environment="production"} 0.5
```

### Export Configuration

**Export Settings**:
```yaml
dashboard:
  export:
    json:
      enabled: true
      path: "/var/log/kyb-platform/dashboards"
      retention: "7d"
    csv:
      enabled: true
      path: "/var/log/kyb-platform/dashboards"
      retention: "30d"
    prometheus:
      enabled: true
      endpoint: "/metrics"
      interval: "15s"
```

## Dashboard Monitoring

### Dashboard Health Checks

**Health Check Endpoints**:
```go
// Dashboard health check
GET /v1/dashboard/health

// Metrics availability check
GET /v1/dashboard/metrics/health

// Widget rendering check
GET /v1/dashboard/widgets/health
```

**Health Check Response**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "components": {
    "metrics_collection": "healthy",
    "widget_generation": "healthy",
    "alert_processing": "healthy",
    "export_system": "healthy"
  },
  "metrics": {
    "total_widgets": 45,
    "active_alerts": 2,
    "last_update": "2024-01-15T10:29:45Z"
  }
}
```

### Dashboard Performance

**Performance Metrics**:
- Dashboard rendering time
- Widget update frequency
- Metrics collection latency
- Alert processing time
- Export generation time

**Performance Optimization**:
- Caching dashboard data
- Batch widget updates
- Optimized metric queries
- Efficient alert processing
- Compressed export formats

## Dashboard Security

### Access Control

**Authentication**:
- Grafana authentication required
- API key authentication
- Role-based access control
- Session management

**Authorization**:
- Dashboard-level permissions
- Widget-level permissions
- Metric-level permissions
- Export permissions

**Audit Logging**:
- Dashboard access events
- Widget interaction events
- Export events
- Configuration changes

### Data Protection

**Data Encryption**:
- Dashboard data encryption
- Export file encryption
- API communication encryption
- Configuration encryption

**Data Privacy**:
- PII data masking
- Sensitive metric filtering
- Access log anonymization
- Export data sanitization

## Dashboard Best Practices

### Design Principles

**Dashboard Design**:
- Keep dashboards focused and relevant
- Use consistent color schemes
- Implement proper thresholds
- Provide clear descriptions
- Include drill-down capabilities

**Widget Layout**:
- Group related widgets together
- Use appropriate widget sizes
- Maintain visual hierarchy
- Ensure responsive design
- Optimize for readability

**Performance Optimization**:
- Limit dashboard complexity
- Use efficient queries
- Implement proper caching
- Optimize refresh intervals
- Monitor resource usage

### Operational Practices

**Dashboard Management**:
- Regular dashboard reviews
- Performance monitoring
- Alert tuning
- Configuration backups
- Documentation updates

**Dashboard Maintenance**:
- Metric validation
- Threshold adjustment
- Widget optimization
- Alert refinement
- Export cleanup

**Dashboard Monitoring**:
- Dashboard availability
- Widget performance
- Alert effectiveness
- Export success rate
- User satisfaction

---

This documentation provides a comprehensive overview of the KYB Platform's performance dashboard system. For specific implementation details, refer to the dashboard code and configuration files referenced throughout this document.
