# Monitoring Procedures Documentation

## Overview

This document provides comprehensive monitoring procedures for the KYB Platform, covering system health monitoring, performance monitoring, business metrics monitoring, and alerting procedures. These procedures ensure optimal system performance and early detection of issues.

## Table of Contents

1. [Monitoring Architecture](#monitoring-architecture)
2. [System Health Monitoring](#system-health-monitoring)
3. [Performance Monitoring](#performance-monitoring)
4. [Business Metrics Monitoring](#business-metrics-monitoring)
5. [Database Monitoring](#database-monitoring)
6. [External Service Monitoring](#external-service-monitoring)
7. [Alerting Procedures](#alerting-procedures)
8. [Monitoring Dashboards](#monitoring-dashboards)
9. [Incident Response](#incident-response)
10. [Maintenance and Optimization](#maintenance-and-optimization)

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

### Key Components

#### 1. Metrics Collection
- **Application Metrics**: Custom business and performance metrics
- **System Metrics**: CPU, memory, disk, network utilization
- **Database Metrics**: Connection pools, query performance, locks
- **External API Metrics**: Response times, error rates, rate limits

#### 2. Metrics Storage
- **Prometheus**: Time-series database for metrics storage
- **Retention**: 30 days for detailed metrics, 1 year for aggregated data
- **Compression**: Automatic data compression for long-term storage

#### 3. Visualization
- **Grafana**: Dashboard creation and visualization
- **Custom Dashboards**: Business-specific and technical dashboards
- **Real-time Monitoring**: Live system status and performance

#### 4. Alerting
- **AlertManager**: Alert routing and notification management
- **Multiple Channels**: Email, Slack, PagerDuty integration
- **Escalation Policies**: Automatic escalation for critical alerts

## System Health Monitoring

### Health Check Endpoints

#### Application Health Checks

```bash
# Basic health check
curl -f http://localhost:8080/health

# Detailed health check
curl -f http://localhost:8080/health/detailed

# Kubernetes health checks
curl -f http://localhost:8080/health/live    # Liveness probe
curl -f http://localhost:8080/health/ready   # Readiness probe
curl -f http://localhost:8080/health/startup # Startup probe
```

#### Health Check Script

```bash
#!/bin/bash
# scripts/health-check.sh

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
LOG_FILE="/var/log/kyb-platform/health-check.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_endpoint() {
    local endpoint="$1"
    local expected_status="${2:-200}"
    
    local response=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}${endpoint}")
    
    if [[ "$response" == "$expected_status" ]]; then
        log "✓ $endpoint: OK ($response)"
        return 0
    else
        log "✗ $endpoint: FAILED ($response, expected $expected_status)"
        return 1
    fi
}

check_database() {
    log "Checking database connectivity..."
    
    if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
        log "✓ Database: OK"
        return 0
    else
        log "✗ Database: FAILED"
        return 1
    fi
}

check_redis() {
    log "Checking Redis connectivity..."
    
    if redis-cli ping >/dev/null 2>&1; then
        log "✓ Redis: OK"
        return 0
    else
        log "✗ Redis: FAILED"
        return 1
    fi
}

check_disk_space() {
    log "Checking disk space..."
    
    local usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [[ $usage -lt 80 ]]; then
        log "✓ Disk space: OK ($usage% used)"
        return 0
    else
        log "✗ Disk space: WARNING ($usage% used)"
        return 1
    fi
}

check_memory() {
    log "Checking memory usage..."
    
    local usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    
    if [[ $usage -lt 85 ]]; then
        log "✓ Memory: OK ($usage% used)"
        return 0
    else
        log "✗ Memory: WARNING ($usage% used)"
        return 1
    fi
}

main() {
    log "Starting health check..."
    
    local failed_checks=0
    
    # Application health checks
    check_endpoint "/health" || ((failed_checks++))
    check_endpoint "/health/detailed" || ((failed_checks++))
    
    # Infrastructure checks
    check_database || ((failed_checks++))
    check_redis || ((failed_checks++))
    check_disk_space || ((failed_checks++))
    check_memory || ((failed_checks++))
    
    if [[ $failed_checks -eq 0 ]]; then
        log "All health checks passed"
        exit 0
    else
        log "Health check failed: $failed_checks issues found"
        exit 1
    fi
}

main "$@"
```

### System Resource Monitoring

#### Resource Monitoring Script

```bash
#!/bin/bash
# scripts/monitor-resources.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/resource-monitoring.log"
ALERT_THRESHOLD_CPU=80
ALERT_THRESHOLD_MEMORY=85
ALERT_THRESHOLD_DISK=90

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_cpu() {
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
    local cpu_int=${cpu_usage%.*}
    
    if [[ $cpu_int -gt $ALERT_THRESHOLD_CPU ]]; then
        log "ALERT: High CPU usage: ${cpu_usage}%"
        return 1
    else
        log "CPU usage: ${cpu_usage}%"
        return 0
    fi
}

check_memory() {
    local memory_usage=$(free | awk 'NR==2{printf "%.1f", $3*100/$2}')
    local memory_int=${memory_usage%.*}
    
    if [[ $memory_int -gt $ALERT_THRESHOLD_MEMORY ]]; then
        log "ALERT: High memory usage: ${memory_usage}%"
        return 1
    else
        log "Memory usage: ${memory_usage}%"
        return 0
    fi
}

check_disk() {
    local disk_usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [[ $disk_usage -gt $ALERT_THRESHOLD_DISK ]]; then
        log "ALERT: High disk usage: ${disk_usage}%"
        return 1
    else
        log "Disk usage: ${disk_usage}%"
        return 0
    fi
}

check_network() {
    local network_connections=$(netstat -an | grep :8080 | wc -l)
    
    if [[ $network_connections -gt 1000 ]]; then
        log "ALERT: High number of network connections: $network_connections"
        return 1
    else
        log "Network connections: $network_connections"
        return 0
    fi
}

main() {
    log "Starting resource monitoring..."
    
    local alerts=0
    
    check_cpu || ((alerts++))
    check_memory || ((alerts++))
    check_disk || ((alerts++))
    check_network || ((alerts++))
    
    if [[ $alerts -gt 0 ]]; then
        log "Resource monitoring completed with $alerts alerts"
        exit 1
    else
        log "Resource monitoring completed - all systems normal"
        exit 0
    fi
}

main "$@"
```

## Performance Monitoring

### Application Performance Metrics

#### Key Performance Indicators (KPIs)

```go
// Performance metrics collection
type PerformanceMetrics struct {
    // HTTP Metrics
    RequestRate        float64 `json:"request_rate"`         // requests/second
    ResponseTime       float64 `json:"response_time"`        // average response time
    ResponseTimeP95    float64 `json:"response_time_p95"`    // 95th percentile
    ResponseTimeP99    float64 `json:"response_time_p99"`    // 99th percentile
    ErrorRate          float64 `json:"error_rate"`           // error percentage
    
    // Business Metrics
    ClassificationRate float64 `json:"classification_rate"`  // classifications/second
    RiskAssessmentRate float64 `json:"risk_assessment_rate"` // risk assessments/second
    MLModelLatency     float64 `json:"ml_model_latency"`     // ML model response time
    
    // System Metrics
    MemoryUsage        float64 `json:"memory_usage"`         // memory usage percentage
    CPUUsage           float64 `json:"cpu_usage"`            // CPU usage percentage
    GoroutineCount     int     `json:"goroutine_count"`      // number of goroutines
    GCCollectionTime   float64 `json:"gc_collection_time"`   // garbage collection time
}

// Performance monitoring endpoint
func (h *MonitoringHandler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
    metrics := h.performanceMonitor.GetMetrics()
    
    response := map[string]interface{}{
        "timestamp": time.Now().UTC(),
        "metrics":   metrics,
        "status":    "healthy",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

#### Performance Monitoring Script

```bash
#!/bin/bash
# scripts/monitor-performance.sh

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
LOG_FILE="/var/log/kyb-platform/performance-monitoring.log"
ALERT_THRESHOLD_RESPONSE_TIME=2.0
ALERT_THRESHOLD_ERROR_RATE=5.0

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_response_time() {
    local response_time=$(curl -s -o /dev/null -w "%{time_total}" "${BASE_URL}/health")
    
    if (( $(echo "$response_time > $ALERT_THRESHOLD_RESPONSE_TIME" | bc -l) )); then
        log "ALERT: High response time: ${response_time}s"
        return 1
    else
        log "Response time: ${response_time}s"
        return 0
    fi
}

check_error_rate() {
    local error_count=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_http_requests_total.*5[0-9][0-9]' | awk '{sum+=$2} END {print sum}')
    local total_requests=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_http_requests_total' | awk '{sum+=$2} END {print sum}')
    
    if [[ $total_requests -gt 0 ]]; then
        local error_rate=$(echo "scale=2; $error_count * 100 / $total_requests" | bc)
        
        if (( $(echo "$error_rate > $ALERT_THRESHOLD_ERROR_RATE" | bc -l) )); then
            log "ALERT: High error rate: ${error_rate}%"
            return 1
        else
            log "Error rate: ${error_rate}%"
            return 0
        fi
    else
        log "No requests recorded"
        return 0
    fi
}

check_throughput() {
    local classification_rate=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_classification_requests_total' | awk '{sum+=$2} END {print sum}')
    local risk_assessment_rate=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_risk_assessment_requests_total' | awk '{sum+=$2} END {print sum}')
    
    log "Classification rate: $classification_rate requests"
    log "Risk assessment rate: $risk_assessment_rate requests"
}

main() {
    log "Starting performance monitoring..."
    
    local alerts=0
    
    check_response_time || ((alerts++))
    check_error_rate || ((alerts++))
    check_throughput
    
    if [[ $alerts -gt 0 ]]; then
        log "Performance monitoring completed with $alerts alerts"
        exit 1
    else
        log "Performance monitoring completed - performance normal"
        exit 0
    fi
}

main "$@"
```

## Business Metrics Monitoring

### Business Intelligence Metrics

#### Classification Metrics

```bash
#!/bin/bash
# scripts/monitor-business-metrics.sh

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
LOG_FILE="/var/log/kyb-platform/business-metrics.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

get_classification_metrics() {
    log "Collecting classification metrics..."
    
    # Get classification success rate
    local total_classifications=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_classification_requests_total' | awk '{sum+=$2} END {print sum}')
    local successful_classifications=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_classification_requests_total.*status="success"' | awk '{sum+=$2} END {print sum}')
    
    if [[ $total_classifications -gt 0 ]]; then
        local success_rate=$(echo "scale=2; $successful_classifications * 100 / $total_classifications" | bc)
        log "Classification success rate: ${success_rate}%"
    fi
    
    # Get average confidence score
    local avg_confidence=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_classification_confidence_avg' | awk '{sum+=$2; count++} END {if(count>0) print sum/count; else print 0}')
    log "Average classification confidence: ${avg_confidence}"
    
    # Get processing time
    local avg_processing_time=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_classification_processing_time_seconds' | awk '{sum+=$2; count++} END {if(count>0) print sum/count; else print 0}')
    log "Average classification processing time: ${avg_processing_time}s"
}

get_risk_assessment_metrics() {
    log "Collecting risk assessment metrics..."
    
    # Get risk level distribution
    local low_risk=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_risk_assessment_requests_total.*risk_level="low"' | awk '{sum+=$2} END {print sum}')
    local medium_risk=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_risk_assessment_requests_total.*risk_level="medium"' | awk '{sum+=$2} END {print sum}')
    local high_risk=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_risk_assessment_requests_total.*risk_level="high"' | awk '{sum+=$2} END {print sum}')
    
    log "Risk level distribution - Low: $low_risk, Medium: $medium_risk, High: $high_risk"
    
    # Get average risk score
    local avg_risk_score=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_risk_score_avg' | awk '{sum+=$2; count++} END {if(count>0) print sum/count; else print 0}')
    log "Average risk score: ${avg_risk_score}"
}

get_ml_model_metrics() {
    log "Collecting ML model metrics..."
    
    # Get model usage statistics
    local bert_usage=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_ml_model_requests_total.*model="bert"' | awk '{sum+=$2} END {print sum}')
    local distilbert_usage=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_ml_model_requests_total.*model="distilbert"' | awk '{sum+=$2} END {print sum}')
    local custom_usage=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_ml_model_requests_total.*model="custom"' | awk '{sum+=$2} END {print sum}')
    
    log "ML model usage - BERT: $bert_usage, DistilBERT: $distilbert_usage, Custom: $custom_usage"
    
    # Get model accuracy
    local bert_accuracy=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_ml_model_accuracy.*model="bert"' | awk '{print $2}')
    local distilbert_accuracy=$(curl -s "${BASE_URL}/metrics" | grep 'kyb_ml_model_accuracy.*model="distilbert"' | awk '{print $2}')
    
    log "ML model accuracy - BERT: ${bert_accuracy}, DistilBERT: ${distilbert_accuracy}"
}

main() {
    log "Starting business metrics monitoring..."
    
    get_classification_metrics
    get_risk_assessment_metrics
    get_ml_model_metrics
    
    log "Business metrics monitoring completed"
}

main "$@"
```

## Database Monitoring

### Database Performance Monitoring

#### Database Monitoring Script

```bash
#!/bin/bash
# scripts/monitor-database.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/database-monitoring.log"
ALERT_THRESHOLD_CONNECTIONS=80
ALERT_THRESHOLD_QUERY_TIME=5.0

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_connection_pool() {
    log "Checking database connection pool..."
    
    local active_connections=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';")
    local max_connections=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SHOW max_connections;" | tr -d ' ')
    
    local connection_usage=$(echo "scale=2; $active_connections * 100 / $max_connections" | bc)
    
    if (( $(echo "$connection_usage > $ALERT_THRESHOLD_CONNECTIONS" | bc -l) )); then
        log "ALERT: High connection usage: ${connection_usage}% ($active_connections/$max_connections)"
        return 1
    else
        log "Connection usage: ${connection_usage}% ($active_connections/$max_connections)"
        return 0
    fi
}

check_slow_queries() {
    log "Checking for slow queries..."
    
    local slow_queries=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT count(*) 
        FROM pg_stat_activity 
        WHERE state = 'active' 
        AND query_start < now() - interval '${ALERT_THRESHOLD_QUERY_TIME} seconds';")
    
    if [[ $slow_queries -gt 0 ]]; then
        log "ALERT: $slow_queries slow queries detected"
        return 1
    else
        log "No slow queries detected"
        return 0
    fi
}

check_database_size() {
    log "Checking database size..."
    
    local db_size=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT pg_size_pretty(pg_database_size('$DB_NAME'));")
    
    log "Database size: $db_size"
}

check_table_statistics() {
    log "Checking table statistics..."
    
    # Get table sizes
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT 
            schemaname,
            tablename,
            pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
        FROM pg_tables 
        WHERE schemaname = 'public'
        ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
        LIMIT 10;"
}

main() {
    log "Starting database monitoring..."
    
    local alerts=0
    
    check_connection_pool || ((alerts++))
    check_slow_queries || ((alerts++))
    check_database_size
    check_table_statistics
    
    if [[ $alerts -gt 0 ]]; then
        log "Database monitoring completed with $alerts alerts"
        exit 1
    else
        log "Database monitoring completed - database healthy"
        exit 0
    fi
}

main "$@"
```

## External Service Monitoring

### External API Monitoring

#### External Service Health Check

```bash
#!/bin/bash
# scripts/monitor-external-services.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/external-services-monitoring.log"
TIMEOUT=10

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_external_api() {
    local service_name="$1"
    local url="$2"
    local expected_status="${3:-200}"
    
    log "Checking $service_name..."
    
    local response=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT "$url")
    
    if [[ "$response" == "$expected_status" ]]; then
        log "✓ $service_name: OK ($response)"
        return 0
    else
        log "✗ $service_name: FAILED ($response, expected $expected_status)"
        return 1
    fi
}

check_ml_service() {
    log "Checking ML service..."
    
    local response=$(curl -s --max-time $TIMEOUT "http://ml-service:8080/health")
    
    if [[ $? -eq 0 ]] && [[ "$response" == *"healthy"* ]]; then
        log "✓ ML Service: OK"
        return 0
    else
        log "✗ ML Service: FAILED"
        return 1
    fi
}

check_redis_service() {
    log "Checking Redis service..."
    
    if redis-cli ping >/dev/null 2>&1; then
        log "✓ Redis: OK"
        return 0
    else
        log "✗ Redis: FAILED"
        return 1
    fi
}

check_supabase_service() {
    log "Checking Supabase service..."
    
    local response=$(curl -s --max-time $TIMEOUT "https://api.supabase.com/health")
    
    if [[ $? -eq 0 ]]; then
        log "✓ Supabase: OK"
        return 0
    else
        log "✗ Supabase: FAILED"
        return 1
    fi
}

main() {
    log "Starting external services monitoring..."
    
    local alerts=0
    
    # Check internal services
    check_ml_service || ((alerts++))
    check_redis_service || ((alerts++))
    
    # Check external APIs
    check_supabase_service || ((alerts++))
    
    if [[ $alerts -gt 0 ]]; then
        log "External services monitoring completed with $alerts alerts"
        exit 1
    else
        log "External services monitoring completed - all services healthy"
        exit 0
    fi
}

main "$@"
```

## Alerting Procedures

### Alert Configuration

#### Prometheus Alert Rules

```yaml
# configs/prometheus/alerts.yml
groups:
  - name: kyb-platform-alerts
    rules:
      # High error rate
      - alert: HighErrorRate
        expr: rate(kyb_http_requests_total{status_code=~"5.."}[5m]) / rate(kyb_http_requests_total[5m]) * 100 > 5
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }}% for the last 5 minutes"

      # High response time
      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(kyb_http_request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }}s"

      # High memory usage
      - alert: HighMemoryUsage
        expr: (kyb_system_memory_usage_bytes / kyb_system_memory_total_bytes) * 100 > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage detected"
          description: "Memory usage is {{ $value }}%"

      # High CPU usage
      - alert: HighCPUUsage
        expr: 100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"
          description: "CPU usage is {{ $value }}%"

      # Database connection issues
      - alert: DatabaseConnectionHigh
        expr: kyb_database_connections_active / kyb_database_connections_max * 100 > 80
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High database connection usage"
          description: "Database connection usage is {{ $value }}%"

      # Service down
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service is down"
          description: "{{ $labels.instance }} is down"

      # ML model accuracy degradation
      - alert: MLModelAccuracyLow
        expr: kyb_ml_model_accuracy < 0.85
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "ML model accuracy is low"
          description: "ML model {{ $labels.model }} accuracy is {{ $value }}"
```

#### AlertManager Configuration

```yaml
# configs/alertmanager/alertmanager.yml
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@kyb-platform.com'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
    - match:
        severity: warning
      receiver: 'warning-alerts'

receivers:
  - name: 'web.hook'
    webhook_configs:
      - url: 'http://localhost:5001/'

  - name: 'critical-alerts'
    email_configs:
      - to: 'admin@kyb-platform.com'
        subject: 'CRITICAL: {{ .GroupLabels.alertname }}'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          {{ end }}
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
        channel: '#alerts-critical'
        title: 'Critical Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'

  - name: 'warning-alerts'
    email_configs:
      - to: 'team@kyb-platform.com'
        subject: 'WARNING: {{ .GroupLabels.alertname }}'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          {{ end }}
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
        channel: '#alerts-warning'
        title: 'Warning Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
```

### Alert Testing

#### Alert Testing Script

```bash
#!/bin/bash
# scripts/test-alerts.sh

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
ALERTMANAGER_URL="${ALERTMANAGER_URL:-http://localhost:9093}"

test_alert() {
    local alert_name="$1"
    local test_endpoint="$2"
    
    echo "Testing alert: $alert_name"
    
    # Trigger the condition that should cause the alert
    curl -s "$BASE_URL$test_endpoint" >/dev/null
    
    # Wait for alert to be triggered
    sleep 30
    
    # Check if alert is active
    local alert_status=$(curl -s "$ALERTMANAGER_URL/api/v1/alerts" | jq -r ".data[] | select(.labels.alertname == \"$alert_name\") | .status.state")
    
    if [[ "$alert_status" == "active" ]]; then
        echo "✓ Alert $alert_name triggered successfully"
        return 0
    else
        echo "✗ Alert $alert_name not triggered"
        return 1
    fi
}

main() {
    echo "Starting alert testing..."
    
    local failed_tests=0
    
    # Test high error rate alert
    test_alert "HighErrorRate" "/test/error" || ((failed_tests++))
    
    # Test high response time alert
    test_alert "HighResponseTime" "/test/slow" || ((failed_tests++))
    
    if [[ $failed_tests -eq 0 ]]; then
        echo "All alert tests passed"
        exit 0
    else
        echo "$failed_tests alert tests failed"
        exit 1
    fi
}

main "$@"
```

## Monitoring Dashboards

### Grafana Dashboard Configuration

#### Main Dashboard

```json
{
  "dashboard": {
    "title": "KYB Platform - Main Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(kyb_http_requests_total[5m])",
            "legendFormat": "Requests/sec"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(kyb_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          },
          {
            "expr": "histogram_quantile(0.95, rate(kyb_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.99, rate(kyb_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "99th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(kyb_http_requests_total{status_code=~\"5..\"}[5m]) / rate(kyb_http_requests_total[5m]) * 100",
            "legendFormat": "Error Rate %"
          }
        ]
      },
      {
        "title": "Classification Metrics",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(kyb_classification_requests_total[5m])",
            "legendFormat": "Classifications/sec"
          },
          {
            "expr": "kyb_classification_confidence_avg",
            "legendFormat": "Avg Confidence"
          }
        ]
      },
      {
        "title": "Risk Assessment Metrics",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(kyb_risk_assessment_requests_total[5m])",
            "legendFormat": "Risk Assessments/sec"
          },
          {
            "expr": "kyb_risk_score_avg",
            "legendFormat": "Avg Risk Score"
          }
        ]
      },
      {
        "title": "ML Model Performance",
        "type": "graph",
        "targets": [
          {
            "expr": "kyb_ml_model_accuracy",
            "legendFormat": "{{model}} Accuracy"
          },
          {
            "expr": "kyb_ml_model_latency_seconds",
            "legendFormat": "{{model}} Latency"
          }
        ]
      }
    ]
  }
}
```

## Incident Response

### Incident Response Procedures

#### Incident Classification

```bash
#!/bin/bash
# scripts/incident-response.sh

set -euo pipefail

INCIDENT_LOG="/var/log/kyb-platform/incidents.log"

log_incident() {
    local severity="$1"
    local description="$2"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    echo "[$timestamp] SEVERITY: $severity - $description" >> "$INCIDENT_LOG"
}

escalate_incident() {
    local severity="$1"
    local description="$2"
    
    log_incident "$severity" "$description"
    
    case "$severity" in
        "critical")
            # Send to on-call engineer
            curl -X POST "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
                 -H "Content-Type: application/json" \
                 -d "{\"text\":\"🚨 CRITICAL INCIDENT: $description\"}"
            
            # Send email to admin
            echo "CRITICAL INCIDENT: $description" | mail -s "CRITICAL: KYB Platform Incident" admin@kyb-platform.com
            ;;
        "warning")
            # Send to team channel
            curl -X POST "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
                 -H "Content-Type: application/json" \
                 -d "{\"text\":\"⚠️ WARNING: $description\"}"
            ;;
    esac
}

main() {
    # Check for critical issues
    if ! ./scripts/health-check.sh; then
        escalate_incident "critical" "Health check failed"
    fi
    
    if ! ./scripts/monitor-resources.sh; then
        escalate_incident "warning" "Resource monitoring alerts"
    fi
    
    if ! ./scripts/monitor-database.sh; then
        escalate_incident "critical" "Database monitoring alerts"
    fi
}

main "$@"
```

## Maintenance and Optimization

### Monitoring System Maintenance

#### Daily Maintenance Tasks

```bash
#!/bin/bash
# scripts/daily-monitoring-maintenance.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/monitoring-maintenance.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

cleanup_old_logs() {
    log "Cleaning up old monitoring logs..."
    
    # Keep logs for 30 days
    find /var/log/kyb-platform -name "*.log" -mtime +30 -delete
    
    log "Log cleanup completed"
}

verify_monitoring_services() {
    log "Verifying monitoring services..."
    
    local services=("prometheus" "grafana" "alertmanager")
    
    for service in "${services[@]}"; do
        if systemctl is-active --quiet "$service"; then
            log "✓ $service is running"
        else
            log "✗ $service is not running"
            systemctl start "$service"
        fi
    done
}

update_dashboards() {
    log "Updating monitoring dashboards..."
    
    # Backup current dashboards
    curl -s "http://localhost:3000/api/dashboards/home" > "/backups/dashboards/$(date +%Y%m%d).json"
    
    # Update dashboard configurations
    # This would typically involve updating dashboard JSON files
    
    log "Dashboard update completed"
}

main() {
    log "Starting daily monitoring maintenance..."
    
    cleanup_old_logs
    verify_monitoring_services
    update_dashboards
    
    log "Daily monitoring maintenance completed"
}

main "$@"
```

This comprehensive monitoring documentation ensures the KYB Platform maintains optimal performance and reliability through robust monitoring, alerting, and incident response procedures.
