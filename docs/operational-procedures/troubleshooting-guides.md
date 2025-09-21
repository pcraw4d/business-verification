# Troubleshooting Guides Documentation

## Overview

This document provides comprehensive troubleshooting guides for the KYB Platform, covering common issues, diagnostic procedures, and resolution steps. These guides ensure rapid problem identification and resolution to maintain system reliability and performance.

## Table of Contents

1. [Quick Diagnostic Checklist](#quick-diagnostic-checklist)
2. [Application Issues](#application-issues)
3. [Database Issues](#database-issues)
4. [Performance Issues](#performance-issues)
5. [ML Model Issues](#ml-model-issues)
6. [External Service Issues](#external-service-issues)
7. [Security Issues](#security-issues)
8. [Deployment Issues](#deployment-issues)
9. [Monitoring Issues](#monitoring-issues)
10. [Emergency Procedures](#emergency-procedures)

## Quick Diagnostic Checklist

### Pre-Troubleshooting Steps

1. **Check System Status**
   ```bash
   # Overall system health
   curl -f http://localhost:8080/health
   
   # Detailed health check
   curl -f http://localhost:8080/health/detailed
   
   # System resources
   free -h && df -h && top -n 1
   ```

2. **Check Logs**
   ```bash
   # Application logs
   tail -f /var/log/kyb-platform/app.log
   
   # System logs
   journalctl -u kyb-platform -f
   
   # Error logs
   grep -i error /var/log/kyb-platform/*.log
   ```

3. **Check Services**
   ```bash
   # Service status
   systemctl status kyb-platform
   systemctl status postgresql
   systemctl status redis
   
   # Port availability
   netstat -tlnp | grep :8080
   ```

### Common Quick Fixes

| Issue | Quick Fix |
|-------|-----------|
| Application won't start | Check port availability, restart service |
| Database connection failed | Verify credentials, check network connectivity |
| High memory usage | Restart application, check for memory leaks |
| Slow response times | Check database performance, restart services |
| ML model errors | Restart ML service, check model files |

## Application Issues

### Application Won't Start

#### Symptoms
- Application fails to start or crashes immediately
- Service status shows "failed" or "inactive"
- Port 8080 not listening

#### Diagnostic Steps

```bash
#!/bin/bash
# scripts/diagnose-startup-issues.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/startup-diagnosis.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_configuration() {
    log "Checking application configuration..."
    
    # Check environment variables
    if [[ -z "${DB_HOST:-}" ]]; then
        log "ERROR: DB_HOST environment variable not set"
        return 1
    fi
    
    if [[ -z "${DB_PASSWORD:-}" ]]; then
        log "ERROR: DB_PASSWORD environment variable not set"
        return 1
    fi
    
    # Check configuration files
    if [[ ! -f "/opt/kyb-platform/configs/production.yml" ]]; then
        log "ERROR: Configuration file not found"
        return 1
    fi
    
    log "Configuration check passed"
    return 0
}

check_dependencies() {
    log "Checking dependencies..."
    
    # Check database connectivity
    if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
        log "ERROR: Database connection failed"
        return 1
    fi
    
    # Check Redis connectivity
    if ! redis-cli ping >/dev/null 2>&1; then
        log "ERROR: Redis connection failed"
        return 1
    fi
    
    log "Dependencies check passed"
    return 0
}

check_ports() {
    log "Checking port availability..."
    
    if netstat -tlnp | grep :8080 >/dev/null; then
        log "ERROR: Port 8080 is already in use"
        netstat -tlnp | grep :8080
        return 1
    fi
    
    log "Port check passed"
    return 0
}

check_permissions() {
    log "Checking file permissions..."
    
    if [[ ! -r "/opt/kyb-platform/kyb-platform" ]]; then
        log "ERROR: Application binary not readable"
        return 1
    fi
    
    if [[ ! -x "/opt/kyb-platform/kyb-platform" ]]; then
        log "ERROR: Application binary not executable"
        return 1
    fi
    
    log "Permissions check passed"
    return 0
}

main() {
    log "Starting startup issue diagnosis..."
    
    local failed_checks=0
    
    check_configuration || ((failed_checks++))
    check_dependencies || ((failed_checks++))
    check_ports || ((failed_checks++))
    check_permissions || ((failed_checks++))
    
    if [[ $failed_checks -eq 0 ]]; then
        log "All startup checks passed"
        exit 0
    else
        log "Startup diagnosis failed: $failed_checks issues found"
        exit 1
    fi
}

main "$@"
```

#### Common Solutions

1. **Port Already in Use**
   ```bash
   # Find process using port 8080
   lsof -i :8080
   
   # Kill the process
   sudo kill -9 <PID>
   
   # Or change the port in configuration
   export PORT=8081
   ```

2. **Configuration Error**
   ```bash
   # Validate environment variables
   env | grep KYB_
   
   # Check configuration file syntax
   ./kyb-platform --config-check
   
   # Test configuration
   ./kyb-platform --dry-run
   ```

3. **Permission Issues**
   ```bash
   # Fix file permissions
   sudo chown -R kyb:kyb /opt/kyb-platform/
   sudo chmod +x /opt/kyb-platform/kyb-platform
   
   # Check SELinux status
   getenforce
   sudo setsebool -P httpd_can_network_connect 1
   ```

### Application Crashes

#### Symptoms
- Application starts but crashes periodically
- Memory usage increases over time
- Goroutine leaks
- Panic errors in logs

#### Diagnostic Steps

```bash
#!/bin/bash
# scripts/diagnose-crash-issues.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/crash-diagnosis.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_memory_usage() {
    log "Checking memory usage patterns..."
    
    # Get current memory usage
    local memory_usage=$(ps -o pid,vsz,rss,comm -p $(pgrep kyb-platform) | tail -n1)
    log "Current memory usage: $memory_usage"
    
    # Check for memory leaks
    local memory_trend=$(journalctl -u kyb-platform --since "1 hour ago" | grep -i "memory" | tail -5)
    if [[ -n "$memory_trend" ]]; then
        log "Memory trend: $memory_trend"
    fi
}

check_goroutine_count() {
    log "Checking goroutine count..."
    
    # Get goroutine count from metrics endpoint
    local goroutine_count=$(curl -s http://localhost:8080/metrics | grep 'go_goroutines' | awk '{print $2}')
    log "Current goroutine count: $goroutine_count"
    
    if [[ $goroutine_count -gt 1000 ]]; then
        log "WARNING: High goroutine count detected"
        return 1
    fi
    
    return 0
}

check_panic_logs() {
    log "Checking for panic logs..."
    
    local panic_count=$(journalctl -u kyb-platform --since "24 hours ago" | grep -i panic | wc -l)
    log "Panic count in last 24 hours: $panic_count"
    
    if [[ $panic_count -gt 0 ]]; then
        log "Recent panics found:"
        journalctl -u kyb-platform --since "24 hours ago" | grep -i panic | tail -5
        return 1
    fi
    
    return 0
}

check_error_logs() {
    log "Checking error logs..."
    
    local error_count=$(journalctl -u kyb-platform --since "1 hour ago" | grep -i error | wc -l)
    log "Error count in last hour: $error_count"
    
    if [[ $error_count -gt 10 ]]; then
        log "High error rate detected:"
        journalctl -u kyb-platform --since "1 hour ago" | grep -i error | tail -10
        return 1
    fi
    
    return 0
}

main() {
    log "Starting crash issue diagnosis..."
    
    local issues=0
    
    check_memory_usage
    check_goroutine_count || ((issues++))
    check_panic_logs || ((issues++))
    check_error_logs || ((issues++))
    
    if [[ $issues -eq 0 ]]; then
        log "No crash issues detected"
        exit 0
    else
        log "Crash diagnosis completed: $issues issues found"
        exit 1
    fi
}

main "$@"
```

#### Common Solutions

1. **Memory Leaks**
   ```bash
   # Restart application
   systemctl restart kyb-platform
   
   # Check for memory leaks in code
   go tool pprof http://localhost:8080/debug/pprof/heap
   
   # Monitor memory usage
   watch -n 1 'ps -o pid,vsz,rss,comm -p $(pgrep kyb-platform)'
   ```

2. **Goroutine Leaks**
   ```bash
   # Get goroutine dump
   curl http://localhost:8080/debug/pprof/goroutine?debug=1
   
   # Analyze goroutine stack
   go tool pprof http://localhost:8080/debug/pprof/goroutine
   ```

3. **Panic Recovery**
   ```bash
   # Check panic logs
   journalctl -u kyb-platform | grep -i panic
   
   # Enable panic recovery in code
   # Add defer recover() in main functions
   ```

## Database Issues

### Database Connection Issues

#### Symptoms
- "Failed to connect to database" errors
- Connection timeout errors
- Connection pool exhaustion
- Database lock timeouts

#### Diagnostic Steps

```bash
#!/bin/bash
# scripts/diagnose-database-issues.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/database-diagnosis.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_connectivity() {
    log "Checking database connectivity..."
    
    if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
        log "✓ Database connectivity: OK"
        return 0
    else
        log "✗ Database connectivity: FAILED"
        return 1
    fi
}

check_connection_pool() {
    log "Checking connection pool status..."
    
    local active_connections=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';")
    local max_connections=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SHOW max_connections;" | tr -d ' ')
    
    log "Active connections: $active_connections"
    log "Max connections: $max_connections"
    
    local usage_percent=$(echo "scale=2; $active_connections * 100 / $max_connections" | bc)
    log "Connection usage: ${usage_percent}%"
    
    if (( $(echo "$usage_percent > 80" | bc -l) )); then
        log "WARNING: High connection usage"
        return 1
    fi
    
    return 0
}

check_slow_queries() {
    log "Checking for slow queries..."
    
    local slow_queries=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT count(*) 
        FROM pg_stat_activity 
        WHERE state = 'active' 
        AND query_start < now() - interval '5 seconds';")
    
    log "Slow queries (>5s): $slow_queries"
    
    if [[ $slow_queries -gt 0 ]]; then
        log "Slow queries detected:"
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            SELECT pid, now() - pg_stat_activity.query_start AS duration, query 
            FROM pg_stat_activity 
            WHERE (now() - pg_stat_activity.query_start) > interval '5 seconds';"
        return 1
    fi
    
    return 0
}

check_database_locks() {
    log "Checking for database locks..."
    
    local lock_count=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT count(*) 
        FROM pg_locks 
        WHERE NOT granted;")
    
    log "Blocked locks: $lock_count"
    
    if [[ $lock_count -gt 0 ]]; then
        log "Database locks detected:"
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            SELECT blocked_locks.pid AS blocked_pid,
                   blocked_activity.usename AS blocked_user,
                   blocking_locks.pid AS blocking_pid,
                   blocking_activity.usename AS blocking_user,
                   blocked_activity.query AS blocked_statement,
                   blocking_activity.query AS current_statement_in_blocking_process
            FROM pg_catalog.pg_locks blocked_locks
            JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
            JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
            AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
            AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
            AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
            AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
            AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
            AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
            AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
            AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
            AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
            AND blocking_locks.pid != blocked_locks.pid
            JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
            WHERE NOT blocked_locks.granted;"
        return 1
    fi
    
    return 0
}

check_database_size() {
    log "Checking database size..."
    
    local db_size=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT pg_size_pretty(pg_database_size('$DB_NAME'));")
    
    log "Database size: $db_size"
    
    # Check for large tables
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
    log "Starting database issue diagnosis..."
    
    local issues=0
    
    check_connectivity || ((issues++))
    check_connection_pool || ((issues++))
    check_slow_queries || ((issues++))
    check_database_locks || ((issues++))
    check_database_size
    
    if [[ $issues -eq 0 ]]; then
        log "No database issues detected"
        exit 0
    else
        log "Database diagnosis completed: $issues issues found"
        exit 1
    fi
}

main "$@"
```

#### Common Solutions

1. **Connection Pool Exhaustion**
   ```bash
   # Increase connection pool size
   export DB_MAX_CONNECTIONS=100
   
   # Kill idle connections
   PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
       SELECT pg_terminate_backend(pid) 
       FROM pg_stat_activity 
       WHERE state = 'idle' 
       AND query_start < now() - interval '1 hour';"
   ```

2. **Slow Queries**
   ```bash
   # Kill long-running queries
   PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
       SELECT pg_terminate_backend(pid) 
       FROM pg_stat_activity 
       WHERE state = 'active' 
       AND query_start < now() - interval '10 minutes';"
   
   # Analyze query performance
   PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
       SELECT query, mean_time, calls 
       FROM pg_stat_statements 
       ORDER BY mean_time DESC 
       LIMIT 10;"
   ```

3. **Database Locks**
   ```bash
   # Kill blocking processes
   PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
       SELECT pg_terminate_backend(blocking_pid) 
       FROM (SELECT blocking_locks.pid AS blocking_pid
             FROM pg_catalog.pg_locks blocked_locks
             JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
             WHERE NOT blocked_locks.granted) AS blocking;"
   ```

## Performance Issues

### Slow Response Times

#### Symptoms
- API response times > 2 seconds
- High 95th percentile latency
- User complaints about slow performance
- Timeout errors

#### Diagnostic Steps

```bash
#!/bin/bash
# scripts/diagnose-performance-issues.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/performance-diagnosis.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_response_times() {
    log "Checking API response times..."
    
    local endpoints=("/health" "/api/v1/classify" "/api/v1/risk/assess")
    
    for endpoint in "${endpoints[@]}"; do
        local response_time=$(curl -s -o /dev/null -w "%{time_total}" "http://localhost:8080$endpoint")
        log "Response time for $endpoint: ${response_time}s"
        
        if (( $(echo "$response_time > 2.0" | bc -l) )); then
            log "WARNING: Slow response time for $endpoint"
        fi
    done
}

check_system_resources() {
    log "Checking system resources..."
    
    # CPU usage
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
    log "CPU usage: ${cpu_usage}%"
    
    # Memory usage
    local memory_usage=$(free | awk 'NR==2{printf "%.1f", $3*100/$2}')
    log "Memory usage: ${memory_usage}%"
    
    # Disk I/O
    local disk_usage=$(iostat -x 1 1 | grep -v "^$" | tail -n +4)
    log "Disk I/O: $disk_usage"
}

check_database_performance() {
    log "Checking database performance..."
    
    # Check for slow queries
    local slow_queries=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT count(*) 
        FROM pg_stat_activity 
        WHERE state = 'active' 
        AND query_start < now() - interval '1 second';")
    
    log "Active queries: $slow_queries"
    
    # Check database connections
    local connections=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT count(*) FROM pg_stat_activity;")
    
    log "Database connections: $connections"
}

check_application_metrics() {
    log "Checking application metrics..."
    
    # Get metrics from Prometheus endpoint
    local request_rate=$(curl -s http://localhost:8080/metrics | grep 'kyb_http_requests_total' | awk '{sum+=$2} END {print sum}')
    local error_rate=$(curl -s http://localhost:8080/metrics | grep 'kyb_http_requests_total.*5[0-9][0-9]' | awk '{sum+=$2} END {print sum}')
    
    log "Request rate: $request_rate"
    log "Error rate: $error_rate"
    
    if [[ $request_rate -gt 0 ]]; then
        local error_percentage=$(echo "scale=2; $error_rate * 100 / $request_rate" | bc)
        log "Error percentage: ${error_percentage}%"
    fi
}

main() {
    log "Starting performance issue diagnosis..."
    
    check_response_times
    check_system_resources
    check_database_performance
    check_application_metrics
    
    log "Performance diagnosis completed"
}

main "$@"
```

#### Common Solutions

1. **High CPU Usage**
   ```bash
   # Check for CPU-intensive processes
   top -o %CPU
   
   # Profile CPU usage
   go tool pprof http://localhost:8080/debug/pprof/profile
   
   # Optimize database queries
   PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
       SELECT query, mean_time, calls 
       FROM pg_stat_statements 
       ORDER BY mean_time DESC 
       LIMIT 10;"
   ```

2. **High Memory Usage**
   ```bash
   # Check memory usage by process
   ps aux --sort=-%mem | head -10
   
   # Profile memory usage
   go tool pprof http://localhost:8080/debug/pprof/heap
   
   # Check for memory leaks
   go tool pprof -alloc_space http://localhost:8080/debug/pprof/heap
   ```

3. **Database Performance**
   ```bash
   # Update table statistics
   PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "ANALYZE;"
   
   # Reindex tables
   PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "REINDEX DATABASE $DB_NAME;"
   
   # Check for missing indexes
   PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
       SELECT schemaname, tablename, attname, n_distinct, correlation 
       FROM pg_stats 
       WHERE schemaname = 'public' 
       ORDER BY n_distinct DESC;"
   ```

## ML Model Issues

### ML Model Performance Issues

#### Symptoms
- Low classification accuracy
- High ML model latency
- Model prediction errors
- ML service unavailability

#### Diagnostic Steps

```bash
#!/bin/bash
# scripts/diagnose-ml-issues.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/ml-diagnosis.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_ml_service_health() {
    log "Checking ML service health..."
    
    local response=$(curl -s --max-time 10 "http://ml-service:8080/health")
    
    if [[ $? -eq 0 ]] && [[ "$response" == *"healthy"* ]]; then
        log "✓ ML Service: OK"
        return 0
    else
        log "✗ ML Service: FAILED"
        return 1
    fi
}

check_model_accuracy() {
    log "Checking model accuracy..."
    
    # Get accuracy metrics from Prometheus
    local bert_accuracy=$(curl -s http://localhost:8080/metrics | grep 'kyb_ml_model_accuracy.*model="bert"' | awk '{print $2}')
    local distilbert_accuracy=$(curl -s http://localhost:8080/metrics | grep 'kyb_ml_model_accuracy.*model="distilbert"' | awk '{print $2}')
    
    log "BERT accuracy: ${bert_accuracy}"
    log "DistilBERT accuracy: ${distilbert_accuracy}"
    
    if (( $(echo "$bert_accuracy < 0.85" | bc -l) )); then
        log "WARNING: BERT model accuracy is low"
        return 1
    fi
    
    if (( $(echo "$distilbert_accuracy < 0.80" | bc -l) )); then
        log "WARNING: DistilBERT model accuracy is low"
        return 1
    fi
    
    return 0
}

check_model_latency() {
    log "Checking model latency..."
    
    local bert_latency=$(curl -s http://localhost:8080/metrics | grep 'kyb_ml_model_latency_seconds.*model="bert"' | awk '{print $2}')
    local distilbert_latency=$(curl -s http://localhost:8080/metrics | grep 'kyb_ml_model_latency_seconds.*model="distilbert"' | awk '{print $2}')
    
    log "BERT latency: ${bert_latency}s"
    log "DistilBERT latency: ${distilbert_latency}s"
    
    if (( $(echo "$bert_latency > 5.0" | bc -l) )); then
        log "WARNING: BERT model latency is high"
        return 1
    fi
    
    if (( $(echo "$distilbert_latency > 2.0" | bc -l) )); then
        log "WARNING: DistilBERT model latency is high"
        return 1
    fi
    
    return 0
}

check_model_files() {
    log "Checking model files..."
    
    local model_path="/opt/kyb-platform/models"
    
    if [[ ! -d "$model_path" ]]; then
        log "ERROR: Model directory not found"
        return 1
    fi
    
    local bert_model="${model_path}/bert-model"
    local distilbert_model="${model_path}/distilbert-model"
    
    if [[ ! -f "$bert_model" ]]; then
        log "ERROR: BERT model file not found"
        return 1
    fi
    
    if [[ ! -f "$distilbert_model" ]]; then
        log "ERROR: DistilBERT model file not found"
        return 1
    fi
    
    log "Model files check passed"
    return 0
}

test_model_prediction() {
    log "Testing model prediction..."
    
    local test_data='{
        "business_name": "Test Company",
        "description": "Software development company",
        "options": {
            "strategies": ["ml_bert", "ml_distilbert"]
        }
    }'
    
    local response=$(curl -s -X POST "http://localhost:8080/api/v2/classify" \
        -H "Content-Type: application/json" \
        -d "$test_data")
    
    if [[ $? -eq 0 ]] && [[ "$response" == *"success"* ]]; then
        log "✓ Model prediction test: OK"
        return 0
    else
        log "✗ Model prediction test: FAILED"
        log "Response: $response"
        return 1
    fi
}

main() {
    log "Starting ML issue diagnosis..."
    
    local issues=0
    
    check_ml_service_health || ((issues++))
    check_model_accuracy || ((issues++))
    check_model_latency || ((issues++))
    check_model_files || ((issues++))
    test_model_prediction || ((issues++))
    
    if [[ $issues -eq 0 ]]; then
        log "No ML issues detected"
        exit 0
    else
        log "ML diagnosis completed: $issues issues found"
        exit 1
    fi
}

main "$@"
```

#### Common Solutions

1. **Model Accuracy Issues**
   ```bash
   # Retrain models with new data
   curl -X POST "http://ml-service:8080/api/v1/models/retrain" \
        -H "Content-Type: application/json" \
        -d '{"model": "bert", "data_version": "latest"}'
   
   # Check model performance metrics
   curl -s "http://ml-service:8080/api/v1/models/performance"
   ```

2. **High Model Latency**
   ```bash
   # Restart ML service
   systemctl restart ml-service
   
   # Check ML service resources
   docker stats ml-service
   
   # Optimize model inference
   curl -X POST "http://ml-service:8080/api/v1/models/optimize" \
        -H "Content-Type: application/json" \
        -d '{"model": "bert", "optimization": "quantization"}'
   ```

3. **Model File Issues**
   ```bash
   # Download latest models
   curl -X POST "http://ml-service:8080/api/v1/models/download" \
        -H "Content-Type: application/json" \
        -d '{"models": ["bert", "distilbert"]}'
   
   # Verify model integrity
   curl -s "http://ml-service:8080/api/v1/models/verify"
   ```

## External Service Issues

### External API Issues

#### Symptoms
- External API timeouts
- Rate limiting errors
- Authentication failures
- Service unavailability

#### Diagnostic Steps

```bash
#!/bin/bash
# scripts/diagnose-external-service-issues.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/external-service-diagnosis.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_supabase_connectivity() {
    log "Checking Supabase connectivity..."
    
    local response=$(curl -s --max-time 10 "https://api.supabase.com/health")
    
    if [[ $? -eq 0 ]]; then
        log "✓ Supabase: OK"
        return 0
    else
        log "✗ Supabase: FAILED"
        return 1
    fi
}

check_external_api_rates() {
    log "Checking external API rate limits..."
    
    # Check rate limit headers
    local response=$(curl -s -I "https://api.external-service.com/endpoint")
    local rate_limit=$(echo "$response" | grep -i "x-ratelimit-limit" | cut -d' ' -f2)
    local rate_remaining=$(echo "$response" | grep -i "x-ratelimit-remaining" | cut -d' ' -f2)
    
    log "Rate limit: $rate_limit"
    log "Rate remaining: $rate_remaining"
    
    if [[ -n "$rate_remaining" ]] && [[ $rate_remaining -lt 10 ]]; then
        log "WARNING: Low rate limit remaining"
        return 1
    fi
    
    return 0
}

check_authentication() {
    log "Checking external service authentication..."
    
    local auth_response=$(curl -s -X POST "https://api.external-service.com/auth" \
        -H "Content-Type: application/json" \
        -d '{"api_key": "'"$EXTERNAL_API_KEY"'"}')
    
    if [[ "$auth_response" == *"success"* ]]; then
        log "✓ Authentication: OK"
        return 0
    else
        log "✗ Authentication: FAILED"
        log "Response: $auth_response"
        return 1
    fi
}

check_network_connectivity() {
    log "Checking network connectivity..."
    
    # Test DNS resolution
    if nslookup api.external-service.com >/dev/null 2>&1; then
        log "✓ DNS resolution: OK"
    else
        log "✗ DNS resolution: FAILED"
        return 1
    fi
    
    # Test port connectivity
    if nc -z api.external-service.com 443 >/dev/null 2>&1; then
        log "✓ Port connectivity: OK"
    else
        log "✗ Port connectivity: FAILED"
        return 1
    fi
    
    return 0
}

main() {
    log "Starting external service issue diagnosis..."
    
    local issues=0
    
    check_supabase_connectivity || ((issues++))
    check_external_api_rates || ((issues++))
    check_authentication || ((issues++))
    check_network_connectivity || ((issues++))
    
    if [[ $issues -eq 0 ]]; then
        log "No external service issues detected"
        exit 0
    else
        log "External service diagnosis completed: $issues issues found"
        exit 1
    fi
}

main "$@"
```

## Emergency Procedures

### Critical System Failure

#### Emergency Response Checklist

```bash
#!/bin/bash
# scripts/emergency-response.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/emergency-response.log"
ALERT_EMAIL="admin@kyb-platform.com"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

send_emergency_alert() {
    local message="$1"
    
    # Send email alert
    echo "$message" | mail -s "EMERGENCY: KYB Platform Critical Issue" "$ALERT_EMAIL"
    
    # Send Slack alert
    curl -X POST "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
         -H "Content-Type: application/json" \
         -d "{\"text\":\"🚨 EMERGENCY: $message\"}"
    
    log "Emergency alert sent: $message"
}

assess_system_status() {
    log "Assessing system status..."
    
    local critical_issues=0
    
    # Check if application is responding
    if ! curl -f http://localhost:8080/health >/dev/null 2>&1; then
        log "CRITICAL: Application not responding"
        ((critical_issues++))
    fi
    
    # Check database connectivity
    if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
        log "CRITICAL: Database not accessible"
        ((critical_issues++))
    fi
    
    # Check disk space
    local disk_usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [[ $disk_usage -gt 95 ]]; then
        log "CRITICAL: Disk space critically low: ${disk_usage}%"
        ((critical_issues++))
    fi
    
    # Check memory usage
    local memory_usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    if [[ $memory_usage -gt 95 ]]; then
        log "CRITICAL: Memory usage critically high: ${memory_usage}%"
        ((critical_issues++))
    fi
    
    return $critical_issues
}

initiate_failover() {
    log "Initiating failover procedures..."
    
    # Stop primary application
    systemctl stop kyb-platform
    
    # Start backup application
    systemctl start kyb-platform-backup
    
    # Update DNS/load balancer
    # This would depend on your infrastructure setup
    
    log "Failover initiated"
}

restore_from_backup() {
    log "Initiating backup restoration..."
    
    # Find latest backup
    local latest_backup=$(ls -t /backups/database/full | head -n1)
    
    if [[ -n "$latest_backup" ]]; then
        # Restore database
        ./scripts/restore-database.sh "$latest_backup"
        
        # Restart application
        systemctl restart kyb-platform
        
        log "Backup restoration completed"
    else
        log "ERROR: No backup found for restoration"
        return 1
    fi
}

main() {
    log "Starting emergency response procedures..."
    
    local critical_issues
    assess_system_status
    critical_issues=$?
    
    if [[ $critical_issues -gt 0 ]]; then
        send_emergency_alert "Critical system failure detected: $critical_issues issues"
        
        # Attempt automatic recovery
        if [[ $critical_issues -ge 3 ]]; then
            log "Multiple critical issues detected, initiating failover"
            initiate_failover
        else
            log "Attempting backup restoration"
            restore_from_backup
        fi
    else
        log "No critical issues detected"
    fi
    
    log "Emergency response procedures completed"
}

main "$@"
```

### Data Recovery Procedures

#### Data Recovery Checklist

```bash
#!/bin/bash
# scripts/data-recovery-procedures.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/data-recovery.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

assess_data_loss() {
    log "Assessing data loss..."
    
    # Check for missing tables
    local missing_tables=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT tablename 
        FROM pg_tables 
        WHERE schemaname = 'public' 
        AND tablename NOT IN (
            SELECT tablename 
            FROM information_schema.tables 
            WHERE table_schema = 'public'
        );")
    
    if [[ -n "$missing_tables" ]]; then
        log "Missing tables detected: $missing_tables"
        return 1
    fi
    
    # Check for corrupted data
    local corrupted_data=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT count(*) 
        FROM users 
        WHERE id IS NULL OR email IS NULL;")
    
    if [[ $corrupted_data -gt 0 ]]; then
        log "Corrupted data detected: $corrupted_data records"
        return 1
    fi
    
    log "No data loss detected"
    return 0
}

restore_critical_data() {
    log "Restoring critical data..."
    
    # Find latest backup
    local latest_backup=$(ls -t /backups/database/full | head -n1)
    
    if [[ -n "$latest_backup" ]]; then
        # Create temporary database for restoration
        local temp_db="kyb_recovery_$(date +%s)"
        
        PGPASSWORD="$DB_PASSWORD" createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$temp_db"
        
        # Restore backup to temporary database
        local backup_file="/backups/database/full/$latest_backup/database.sql.gz"
        gunzip -c "$backup_file" | PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$temp_db"
        
        # Verify restoration
        local restored_count=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$temp_db" -t -c "SELECT count(*) FROM users;")
        
        if [[ $restored_count -gt 0 ]]; then
            log "Data restoration verified: $restored_count users restored"
            
            # Clean up temporary database
            PGPASSWORD="$DB_PASSWORD" dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$temp_db"
            
            return 0
        else
            log "ERROR: Data restoration failed"
            PGPASSWORD="$DB_PASSWORD" dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$temp_db"
            return 1
        fi
    else
        log "ERROR: No backup found for restoration"
        return 1
    fi
}

main() {
    log "Starting data recovery procedures..."
    
    if ! assess_data_loss; then
        log "Data loss detected, initiating recovery..."
        
        if restore_critical_data; then
            log "Data recovery completed successfully"
            exit 0
        else
            log "Data recovery failed"
            exit 1
        fi
    else
        log "No data recovery needed"
        exit 0
    fi
}

main "$@"
```

This comprehensive troubleshooting documentation ensures the KYB Platform can quickly identify and resolve issues, maintaining system reliability and minimizing downtime.
