# Risk Assessment Service Operations Guide
## Production Operations and Maintenance

### Overview

This document provides comprehensive operational guidance for the Risk Assessment Service in the KYB Platform. It covers deployment, monitoring, troubleshooting, maintenance, and emergency procedures for production operations.

---

## Table of Contents

1. [Service Overview](#service-overview)
2. [Deployment Procedures](#deployment-procedures)
3. [Monitoring and Alerting](#monitoring-and-alerting)
4. [Health Checks and Diagnostics](#health-checks-and-diagnostics)
5. [Performance Monitoring](#performance-monitoring)
6. [Database Operations](#database-operations)
7. [Redis Operations](#redis-operations)
8. [Log Management](#log-management)
9. [Security Operations](#security-operations)
10. [Troubleshooting Guide](#troubleshooting-guide)
11. [Maintenance Procedures](#maintenance-procedures)
12. [Emergency Procedures](#emergency-procedures)
13. [Backup and Recovery](#backup-and-recovery)
14. [Capacity Planning](#capacity-planning)
15. [Change Management](#change-management)

---

## Service Overview

### Service Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                Risk Assessment Service                      │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   HTTP      │  │   gRPC      │  │   WebSocket │        │
│  │   API       │  │   API       │  │   API       │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Business   │  │   ML        │  │  External   │        │
│  │   Logic     │  │  Engine     │  │   APIs      │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ PostgreSQL  │  │    Redis    │  │   File      │        │
│  │  Database   │  │    Cache    │  │  Storage    │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### Service Components

| Component | Description | Port | Health Check |
|-----------|-------------|------|--------------|
| HTTP API | REST API endpoints | 8080 | `/health` |
| gRPC API | Internal service communication | 8081 | `/grpc.health.v1.Health/Check` |
| WebSocket API | Real-time updates | 8082 | `/ws/health` |
| Business Logic | Risk assessment algorithms | - | Internal |
| ML Engine | Machine learning models | - | Internal |
| External APIs | Third-party integrations | - | Internal |

### Service Dependencies

| Dependency | Type | Critical | Health Check |
|------------|------|----------|--------------|
| PostgreSQL | Database | Yes | Connection test |
| Redis | Cache | Yes | `PING` command |
| External APIs | External | No | HTTP health check |
| File Storage | Storage | No | File system check |

---

## Deployment Procedures

### Pre-Deployment Checklist

- [ ] Code review completed
- [ ] Tests passing (unit, integration, performance)
- [ ] Security scan completed
- [ ] Database migrations tested
- [ ] Configuration validated
- [ ] Monitoring configured
- [ ] Rollback plan prepared
- [ ] Team notified

### Deployment Steps

#### 1. Staging Deployment

```bash
# Deploy to staging environment
cd services/risk-assessment-service
./scripts/deploy_railway.sh --environment=staging

# Verify staging deployment
curl -f https://risk-assessment-service-staging.up.railway.app/health

# Run staging tests
go run ./cmd/load_test.go \
  -url="https://risk-assessment-service-staging.up.railway.app" \
  -duration=5m \
  -users=10 \
  -type=smoke
```

#### 2. Production Deployment

```bash
# Deploy to production
./scripts/deploy_railway.sh --environment=production

# Verify production deployment
curl -f https://risk-assessment-service-production.up.railway.app/health

# Run production smoke tests
go run ./cmd/load_test.go \
  -url="https://risk-assessment-service-production.up.railway.app" \
  -duration=2m \
  -users=5 \
  -type=smoke
```

#### 3. Post-Deployment Verification

```bash
# Check service health
curl -f https://risk-assessment-service-production.up.railway.app/health

# Check metrics endpoint
curl -f https://risk-assessment-service-production.up.railway.app/metrics

# Verify API Gateway integration
curl -f https://api-gateway.kyb-platform.com/api/v1/risk/health

# Test risk assessment endpoint
curl -X POST https://api-gateway.kyb-platform.com/api/v1/risk/assess \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"business_name": "Test Company", "business_address": "123 Test St"}'
```

### Rollback Procedures

#### Automatic Rollback

```bash
# Railway automatic rollback on failure
railway rollback --service risk-assessment-service

# Verify rollback
curl -f https://risk-assessment-service-production.up.railway.app/health
```

#### Manual Rollback

```bash
# Deploy previous version
railway deploy --service risk-assessment-service --version <previous-version>

# Verify rollback
curl -f https://risk-assessment-service-production.up.railway.app/health

# Check logs
railway logs --service risk-assessment-service --tail 100
```

---

## Monitoring and Alerting

### Key Metrics

#### Service Health Metrics

| Metric | Description | Threshold | Alert Level |
|--------|-------------|-----------|-------------|
| `up` | Service availability | 0 | Critical |
| `http_requests_total` | Total HTTP requests | - | Info |
| `http_request_duration_seconds` | Request latency | P99 > 500ms | Warning |
| `http_requests_total{status_code=~"5.."}` | Error rate | > 5% | Critical |

#### Business Metrics

| Metric | Description | Threshold | Alert Level |
|--------|-------------|-----------|-------------|
| `risk_assessments_total` | Total risk assessments | - | Info |
| `risk_assessments_duration_seconds` | Assessment duration | P99 > 10s | Warning |
| `risk_assessments_errors_total` | Assessment errors | > 1% | Critical |
| `ml_predictions_total` | ML predictions | - | Info |

#### System Metrics

| Metric | Description | Threshold | Alert Level |
|--------|-------------|-----------|-------------|
| `process_cpu_seconds_total` | CPU usage | > 80% | Warning |
| `process_resident_memory_bytes` | Memory usage | > 500MB | Warning |
| `go_goroutines` | Goroutine count | > 1000 | Warning |
| `go_gc_duration_seconds` | GC duration | P99 > 100ms | Warning |

### Alert Rules

#### Critical Alerts

```yaml
# Service down
- alert: RiskAssessmentServiceDown
  expr: up{service="risk-assessment-service"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Risk Assessment Service is down"

# High error rate
- alert: RiskAssessmentServiceHighErrorRate
  expr: sum(rate(http_requests_total{service="risk-assessment-service", status_code=~"5.."}[5m])) / sum(rate(http_requests_total{service="risk-assessment-service"}[5m])) * 100 > 5
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "High error rate for Risk Assessment Service"
```

#### Warning Alerts

```yaml
# High latency
- alert: RiskAssessmentServiceHighLatency
  expr: histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{service="risk-assessment-service"}[5m])) by (le)) > 0.5
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High latency for Risk Assessment Service"

# High CPU usage
- alert: RiskAssessmentServiceHighCPUUsage
  expr: sum(rate(process_cpu_seconds_total{service="risk-assessment-service"}[5m])) > 0.8
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High CPU usage for Risk Assessment Service"
```

### Alert Channels

#### Email Alerts

```yaml
# Email configuration
email_configs:
  - to: 'ops-team@kyb-platform.com'
    from: 'alerts@kyb-platform.com'
    smarthost: 'smtp.gmail.com:587'
    auth_username: 'alerts@kyb-platform.com'
    auth_password: '${SMTP_PASSWORD}'
    headers:
      Subject: 'KYB Platform Alert: {{ .GroupLabels.alertname }}'
```

#### Webhook Alerts

```yaml
# Slack webhook
webhook_configs:
  - url: 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX'
    send_resolved: true
    title: 'KYB Platform Alert'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
```

#### PagerDuty Integration

```yaml
# PagerDuty configuration
pagerduty_configs:
  - service_key: '${PAGERDUTY_SERVICE_KEY}'
    description: '{{ .GroupLabels.alertname }}'
    details:
      service: '{{ .GroupLabels.service }}'
      environment: '{{ .GroupLabels.environment }}'
```

---

## Health Checks and Diagnostics

### Health Check Endpoints

#### Basic Health Check

```bash
# Check service health
curl -f https://risk-assessment-service-production.up.railway.app/health

# Expected response
{
  "service": "risk-assessment-service",
  "version": "1.0.0",
  "status": "healthy",
  "timestamp": "2024-12-01T10:30:00Z",
  "uptime": "24h30m15s",
  "dependencies": {
    "database": "healthy",
    "redis": "healthy",
    "external_apis": "healthy"
  }
}
```

#### Detailed Health Check

```bash
# Check detailed health
curl -f https://risk-assessment-service-production.up.railway.app/health/detailed

# Expected response
{
  "service": "risk-assessment-service",
  "version": "1.0.0",
  "status": "healthy",
  "timestamp": "2024-12-01T10:30:00Z",
  "uptime": "24h30m15s",
  "dependencies": {
    "database": {
      "status": "healthy",
      "response_time": "5ms",
      "connection_pool": {
        "active": 5,
        "idle": 10,
        "max": 25
      }
    },
    "redis": {
      "status": "healthy",
      "response_time": "2ms",
      "memory_usage": "45MB",
      "connected_clients": 3
    },
    "external_apis": {
      "status": "healthy",
      "apis": {
        "government_data": "healthy",
        "credit_bureau": "healthy",
        "sanctions_list": "healthy"
      }
    }
  },
  "metrics": {
    "requests_per_second": 15.5,
    "average_response_time": "120ms",
    "error_rate": "0.1%",
    "active_connections": 8
  }
}
```

### Diagnostic Commands

#### Service Status

```bash
# Check service status
railway status --service risk-assessment-service

# Check service logs
railway logs --service risk-assessment-service --tail 100

# Check service metrics
curl https://risk-assessment-service-production.up.railway.app/metrics
```

#### Database Diagnostics

```bash
# Check database connection
psql $DATABASE_URL -c "SELECT 1;"

# Check database performance
psql $DATABASE_URL -c "
SELECT 
  schemaname,
  tablename,
  attname,
  n_distinct,
  correlation
FROM pg_stats 
WHERE schemaname = 'public' 
ORDER BY n_distinct DESC;"

# Check slow queries
psql $DATABASE_URL -c "
SELECT 
  query,
  calls,
  total_time,
  mean_time,
  rows
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;"
```

#### Redis Diagnostics

```bash
# Check Redis connection
redis-cli -u $REDIS_URL ping

# Check Redis info
redis-cli -u $REDIS_URL info

# Check Redis memory usage
redis-cli -u $REDIS_URL info memory

# Check Redis slow log
redis-cli -u $REDIS_URL slowlog get 10
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

### Performance Optimization

#### Database Optimization

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

#### Redis Optimization

```bash
# Check Redis memory usage
redis-cli -u $REDIS_URL info memory

# Check Redis key distribution
redis-cli -u $REDIS_URL --scan --pattern "*" | head -100

# Check Redis slow log
redis-cli -u $REDIS_URL slowlog get 10

# Monitor Redis commands
redis-cli -u $REDIS_URL monitor
```

---

## Database Operations

### Database Maintenance

#### Regular Maintenance Tasks

```bash
# Daily maintenance
psql $DATABASE_URL -c "VACUUM ANALYZE;"

# Weekly maintenance
psql $DATABASE_URL -c "REINDEX DATABASE postgres;"

# Monthly maintenance
psql $DATABASE_URL -c "VACUUM FULL;"
```

#### Database Monitoring

```sql
-- Check database size
SELECT 
  pg_database.datname,
  pg_size_pretty(pg_database_size(pg_database.datname)) AS size
FROM pg_database;

-- Check table sizes
SELECT 
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

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
```

### Database Migrations

#### Running Migrations

```bash
# Run all migrations
./scripts/run-supabase-migrations.sh

# Run specific migration
psql $DATABASE_URL -f supabase-migrations/risk-assessment-schema.sql

# Check migration status
psql $DATABASE_URL -c "
SELECT 
  version,
  applied_at,
  description
FROM schema_migrations 
ORDER BY applied_at DESC;"
```

#### Migration Rollback

```bash
# Rollback last migration
psql $DATABASE_URL -c "
DELETE FROM schema_migrations 
WHERE version = (
  SELECT version 
  FROM schema_migrations 
  ORDER BY applied_at DESC 
  LIMIT 1
);"

# Rollback specific migration
psql $DATABASE_URL -c "
DELETE FROM schema_migrations 
WHERE version = '20241201000000';"
```

---

## Redis Operations

### Redis Maintenance

#### Regular Maintenance Tasks

```bash
# Check Redis memory usage
redis-cli -u $REDIS_URL info memory

# Check Redis key count
redis-cli -u $REDIS_URL dbsize

# Check Redis slow log
redis-cli -u $REDIS_URL slowlog get 10

# Clear slow log
redis-cli -u $REDIS_URL slowlog reset
```

#### Redis Monitoring

```bash
# Monitor Redis commands
redis-cli -u $REDIS_URL monitor

# Check Redis info
redis-cli -u $REDIS_URL info

# Check Redis configuration
redis-cli -u $REDIS_URL config get "*"

# Check Redis clients
redis-cli -u $REDIS_URL client list
```

### Redis Optimization

#### Memory Optimization

```bash
# Check memory usage by key pattern
redis-cli -u $REDIS_URL --scan --pattern "ra:*" | \
  xargs -I {} redis-cli -u $REDIS_URL memory usage {}

# Check memory usage by key type
redis-cli -u $REDIS_URL --scan --pattern "*" | \
  xargs -I {} redis-cli -u $REDIS_URL type {} | \
  sort | uniq -c
```

#### Performance Optimization

```bash
# Check Redis performance
redis-cli -u $REDIS_URL --latency

# Check Redis latency history
redis-cli -u $REDIS_URL --latency-history

# Check Redis latency distribution
redis-cli -u $REDIS_URL --latency-dist
```

---

## Log Management

### Log Collection

#### Application Logs

```bash
# View recent logs
railway logs --service risk-assessment-service --tail 100

# View logs with timestamps
railway logs --service risk-assessment-service --tail 100 --timestamps

# View logs for specific time range
railway logs --service risk-assessment-service --since 1h

# View logs with specific level
railway logs --service risk-assessment-service --tail 100 | grep ERROR
```

#### System Logs

```bash
# View system logs
journalctl -u risk-assessment-service -f

# View system logs with timestamps
journalctl -u risk-assessment-service -f --since "1 hour ago"

# View system logs for specific level
journalctl -u risk-assessment-service -p err -f
```

### Log Analysis

#### Error Analysis

```bash
# Count errors by type
railway logs --service risk-assessment-service --tail 1000 | \
  grep ERROR | \
  awk '{print $4}' | \
  sort | uniq -c | sort -nr

# Find most common errors
railway logs --service risk-assessment-service --tail 1000 | \
  grep ERROR | \
  awk -F'"' '{print $2}' | \
  sort | uniq -c | sort -nr | head -10
```

#### Performance Analysis

```bash
# Find slow requests
railway logs --service risk-assessment-service --tail 1000 | \
  grep "duration" | \
  awk '{print $NF}' | \
  sort -nr | head -10

# Find high memory usage
railway logs --service risk-assessment-service --tail 1000 | \
  grep "memory" | \
  awk '{print $NF}' | \
  sort -nr | head -10
```

---

## Security Operations

### Security Monitoring

#### Authentication Monitoring

```bash
# Check authentication failures
railway logs --service risk-assessment-service --tail 1000 | \
  grep "authentication failed" | \
  wc -l

# Check rate limiting triggers
railway logs --service risk-assessment-service --tail 1000 | \
  grep "rate limit exceeded" | \
  wc -l

# Check suspicious activity
railway logs --service risk-assessment-service --tail 1000 | \
  grep -E "(suspicious|anomaly|attack)" | \
  wc -l
```

#### Access Monitoring

```bash
# Check API access patterns
railway logs --service risk-assessment-service --tail 1000 | \
  grep "POST /api/v1/risk/assess" | \
  awk '{print $1, $2}' | \
  sort | uniq -c | sort -nr

# Check IP addresses
railway logs --service risk-assessment-service --tail 1000 | \
  grep "remote_addr" | \
  awk '{print $NF}' | \
  sort | uniq -c | sort -nr | head -10
```

### Security Updates

#### Dependency Updates

```bash
# Check for security vulnerabilities
go list -json -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy

# Check for known vulnerabilities
gosec ./...
```

#### Configuration Updates

```bash
# Update security headers
railway variables set SECURITY_HEADERS_ENABLED=true

# Update rate limits
railway variables set RATE_LIMIT_REQUESTS_PER_MINUTE=1000

# Update authentication settings
railway variables set JWT_SECRET=<new-secret>
```

---

## Troubleshooting Guide

### Common Issues

#### 1. Service Unavailable (502 Bad Gateway)

**Symptoms**: API Gateway returns 502 errors
**Causes**: 
- Risk Assessment Service is down
- Network connectivity issues
- Service overload

**Solutions**:
```bash
# Check service health
curl -f https://risk-assessment-service-production.up.railway.app/health

# Check service logs
railway logs --service risk-assessment-service --tail 100

# Check service status
railway status --service risk-assessment-service

# Restart service if needed
railway restart --service risk-assessment-service
```

#### 2. High Response Times

**Symptoms**: Requests taking > 500ms
**Causes**:
- Database performance issues
- Redis connectivity problems
- External API delays
- High CPU/memory usage

**Solutions**:
```bash
# Check database performance
psql $DATABASE_URL -c "
SELECT 
  query,
  calls,
  total_time,
  mean_time
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;"

# Check Redis performance
redis-cli -u $REDIS_URL --latency

# Check system resources
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep -E "(cpu|memory|goroutines)"

# Check external API status
curl -f https://external-api.example.com/health
```

#### 3. Authentication Failures

**Symptoms**: 401 Unauthorized errors
**Causes**:
- Invalid JWT tokens
- Expired tokens
- Incorrect API keys
- Service token issues

**Solutions**:
```bash
# Check JWT token validity
jwt decode <token>

# Check token expiration
jwt decode <token> | jq '.exp'

# Check API key configuration
railway variables | grep API_KEY

# Check service token
railway variables | grep SERVICE_TOKEN
```

#### 4. Database Connection Issues

**Symptoms**: Database connection errors
**Causes**:
- Database server down
- Connection pool exhausted
- Network issues
- Authentication problems

**Solutions**:
```bash
# Test database connection
psql $DATABASE_URL -c "SELECT 1;"

# Check connection pool status
curl -s https://risk-assessment-service-production.up.railway.app/health/detailed | \
  jq '.dependencies.database.connection_pool'

# Check database logs
psql $DATABASE_URL -c "
SELECT 
  pid,
  usename,
  application_name,
  client_addr,
  state,
  query_start,
  query
FROM pg_stat_activity 
WHERE state = 'active';"
```

#### 5. Redis Connection Issues

**Symptoms**: Redis connection errors
**Causes**:
- Redis server down
- Network issues
- Authentication problems
- Memory issues

**Solutions**:
```bash
# Test Redis connection
redis-cli -u $REDIS_URL ping

# Check Redis status
redis-cli -u $REDIS_URL info

# Check Redis memory usage
redis-cli -u $REDIS_URL info memory

# Check Redis clients
redis-cli -u $REDIS_URL client list
```

### Debug Commands

#### Service Debug

```bash
# Check service configuration
curl -s https://risk-assessment-service-production.up.railway.app/config

# Check service metrics
curl -s https://risk-assessment-service-production.up.railway.app/metrics

# Check service health
curl -s https://risk-assessment-service-production.up.railway.app/health/detailed
```

#### Database Debug

```bash
# Check database connections
psql $DATABASE_URL -c "
SELECT 
  count(*) as total_connections,
  count(*) FILTER (WHERE state = 'active') as active_connections,
  count(*) FILTER (WHERE state = 'idle') as idle_connections
FROM pg_stat_activity;"

# Check database locks
psql $DATABASE_URL -c "
SELECT 
  pid,
  mode,
  locktype,
  relation::regclass,
  granted
FROM pg_locks 
WHERE NOT granted;"
```

#### Redis Debug

```bash
# Check Redis memory usage
redis-cli -u $REDIS_URL info memory

# Check Redis key count
redis-cli -u $REDIS_URL dbsize

# Check Redis slow log
redis-cli -u $REDIS_URL slowlog get 10
```

---

## Maintenance Procedures

### Daily Maintenance

#### Health Checks

```bash
# Check service health
curl -f https://risk-assessment-service-production.up.railway.app/health

# Check database health
psql $DATABASE_URL -c "SELECT 1;"

# Check Redis health
redis-cli -u $REDIS_URL ping

# Check external APIs
curl -f https://external-api.example.com/health
```

#### Performance Monitoring

```bash
# Check response times
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'http_request_duration_seconds'

# Check error rates
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'http_requests_total{status_code=~"5.."}'

# Check resource usage
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep -E "(cpu|memory|goroutines)"
```

### Weekly Maintenance

#### Database Maintenance

```bash
# Run VACUUM ANALYZE
psql $DATABASE_URL -c "VACUUM ANALYZE;"

# Check table sizes
psql $DATABASE_URL -c "
SELECT 
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"

# Check index usage
psql $DATABASE_URL -c "
SELECT 
  schemaname,
  tablename,
  indexname,
  idx_scan,
  idx_tup_read,
  idx_tup_fetch
FROM pg_stat_user_indexes 
ORDER BY idx_scan DESC;"
```

#### Redis Maintenance

```bash
# Check Redis memory usage
redis-cli -u $REDIS_URL info memory

# Check Redis key distribution
redis-cli -u $REDIS_URL --scan --pattern "*" | \
  head -100 | \
  xargs -I {} redis-cli -u $REDIS_URL type {}

# Check Redis slow log
redis-cli -u $REDIS_URL slowlog get 10
```

### Monthly Maintenance

#### Security Updates

```bash
# Check for security vulnerabilities
go list -json -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy

# Run security scan
gosec ./...
```

#### Performance Optimization

```bash
# Analyze slow queries
psql $DATABASE_URL -c "
SELECT 
  query,
  calls,
  total_time,
  mean_time,
  rows
FROM pg_stat_statements 
WHERE mean_time > 100
ORDER BY mean_time DESC;"

# Check index usage
psql $DATABASE_URL -c "
SELECT 
  schemaname,
  tablename,
  indexname,
  idx_scan,
  idx_tup_read,
  idx_tup_fetch
FROM pg_stat_user_indexes 
WHERE idx_scan = 0
ORDER BY pg_relation_size(indexrelid) DESC;"
```

---

## Emergency Procedures

### Service Outage

#### Immediate Response

1. **Assess the situation**
   ```bash
   # Check service status
   curl -f https://risk-assessment-service-production.up.railway.app/health
   
   # Check service logs
   railway logs --service risk-assessment-service --tail 100
   
   # Check service metrics
   curl -s https://risk-assessment-service-production.up.railway.app/metrics
   ```

2. **Notify stakeholders**
   - Send alert to ops team
   - Update status page
   - Notify business stakeholders

3. **Attempt quick fixes**
   ```bash
   # Restart service
   railway restart --service risk-assessment-service
   
   # Check if restart resolved the issue
   curl -f https://risk-assessment-service-production.up.railway.app/health
   ```

#### Escalation Procedures

1. **If restart doesn't work**
   ```bash
   # Rollback to previous version
   railway rollback --service risk-assessment-service
   
   # Verify rollback
   curl -f https://risk-assessment-service-production.up.railway.app/health
   ```

2. **If rollback doesn't work**
   - Check database connectivity
   - Check Redis connectivity
   - Check external API status
   - Review recent changes

3. **If issue persists**
   - Escalate to senior engineers
   - Consider failover procedures
   - Document incident details

### Data Corruption

#### Immediate Response

1. **Stop service**
   ```bash
   railway stop --service risk-assessment-service
   ```

2. **Assess data integrity**
   ```bash
   # Check database integrity
   psql $DATABASE_URL -c "SELECT pg_database_size('postgres');"
   
   # Check Redis integrity
   redis-cli -u $REDIS_URL info keyspace
   ```

3. **Restore from backup**
   ```bash
   # Restore database from backup
   pg_restore -d $DATABASE_URL backup.sql
   
   # Restore Redis from backup
   redis-cli -u $REDIS_URL --rdb backup.rdb
   ```

### Security Incident

#### Immediate Response

1. **Isolate service**
   ```bash
   # Stop service
   railway stop --service risk-assessment-service
   
   # Block suspicious IPs
   # (Implementation depends on infrastructure)
   ```

2. **Preserve evidence**
   ```bash
   # Export logs
   railway logs --service risk-assessment-service --since 24h > incident.log
   
   # Export metrics
   curl -s https://risk-assessment-service-production.up.railway.app/metrics > metrics.txt
   ```

3. **Notify security team**
   - Send incident report
   - Provide evidence
   - Coordinate response

---

## Backup and Recovery

### Database Backups

#### Automated Backups

```bash
# Daily backup script
#!/bin/bash
BACKUP_DIR="/backups/postgres"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/risk_assessment_$DATE.sql"

# Create backup
pg_dump $DATABASE_URL > $BACKUP_FILE

# Compress backup
gzip $BACKUP_FILE

# Upload to cloud storage
aws s3 cp "$BACKUP_FILE.gz" s3://kyb-backups/postgres/

# Clean up old backups (keep 30 days)
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete
```

#### Manual Backups

```bash
# Create manual backup
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d_%H%M%S).sql

# Create compressed backup
pg_dump $DATABASE_URL | gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz
```

### Redis Backups

#### Automated Backups

```bash
# Daily Redis backup script
#!/bin/bash
BACKUP_DIR="/backups/redis"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/redis_$DATE.rdb"

# Create backup
redis-cli -u $REDIS_URL --rdb $BACKUP_FILE

# Compress backup
gzip $BACKUP_FILE

# Upload to cloud storage
aws s3 cp "$BACKUP_FILE.gz" s3://kyb-backups/redis/

# Clean up old backups (keep 30 days)
find $BACKUP_DIR -name "*.rdb.gz" -mtime +30 -delete
```

#### Manual Backups

```bash
# Create manual Redis backup
redis-cli -u $REDIS_URL --rdb backup_$(date +%Y%m%d_%H%M%S).rdb

# Create compressed backup
redis-cli -u $REDIS_URL --rdb - | gzip > backup_$(date +%Y%m%d_%H%M%S).rdb.gz
```

### Recovery Procedures

#### Database Recovery

```bash
# Restore from backup
pg_restore -d $DATABASE_URL backup.sql

# Restore from compressed backup
gunzip -c backup.sql.gz | psql $DATABASE_URL

# Verify recovery
psql $DATABASE_URL -c "SELECT count(*) FROM risk_assessments;"
```

#### Redis Recovery

```bash
# Restore from backup
redis-cli -u $REDIS_URL --rdb backup.rdb

# Restore from compressed backup
gunzip -c backup.rdb.gz | redis-cli -u $REDIS_URL --rdb -

# Verify recovery
redis-cli -u $REDIS_URL dbsize
```

---

## Capacity Planning

### Resource Monitoring

#### CPU Usage

```bash
# Monitor CPU usage
watch -n 5 'curl -s https://risk-assessment-service-production.up.railway.app/metrics | grep process_cpu_seconds_total'

# Check CPU trends
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep process_cpu_seconds_total | \
  awk '{print $2}' | \
  tail -100 | \
  awk '{sum+=$1; count++} END {print "Average CPU:", sum/count}'
```

#### Memory Usage

```bash
# Monitor memory usage
watch -n 5 'curl -s https://risk-assessment-service-production.up.railway.app/metrics | grep process_resident_memory_bytes'

# Check memory trends
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep process_resident_memory_bytes | \
  awk '{print $2}' | \
  tail -100 | \
  awk '{sum+=$1; count++} END {print "Average Memory:", sum/count/1024/1024 "MB"}'
```

#### Database Usage

```bash
# Monitor database size
psql $DATABASE_URL -c "
SELECT 
  pg_database.datname,
  pg_size_pretty(pg_database_size(pg_database.datname)) AS size
FROM pg_database;"

# Monitor table growth
psql $DATABASE_URL -c "
SELECT 
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
```

### Scaling Procedures

#### Horizontal Scaling

```bash
# Scale service horizontally
railway scale --service risk-assessment-service --replicas 3

# Verify scaling
railway status --service risk-assessment-service

# Check load distribution
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep 'http_requests_total'
```

#### Vertical Scaling

```bash
# Scale service vertically
railway scale --service risk-assessment-service --cpu 2 --memory 2GB

# Verify scaling
railway status --service risk-assessment-service

# Check resource usage
curl -s https://risk-assessment-service-production.up.railway.app/metrics | \
  grep -E "(cpu|memory)"
```

---

## Change Management

### Change Process

#### 1. Change Request

- Document change details
- Assess impact and risk
- Get approval from stakeholders
- Schedule change window

#### 2. Change Implementation

```bash
# Create change branch
git checkout -b change/risk-assessment-update

# Implement changes
# ... make changes ...

# Test changes
go test ./...
go run ./cmd/load_test.go -url=http://localhost:8080 -duration=5m

# Commit changes
git add .
git commit -m "feat: update risk assessment service"
git push origin change/risk-assessment-update
```

#### 3. Change Deployment

```bash
# Deploy to staging
./scripts/deploy_railway.sh --environment=staging

# Verify staging deployment
curl -f https://risk-assessment-service-staging.up.railway.app/health

# Deploy to production
./scripts/deploy_railway.sh --environment=production

# Verify production deployment
curl -f https://risk-assessment-service-production.up.railway.app/health
```

#### 4. Change Verification

```bash
# Run smoke tests
go run ./cmd/load_test.go \
  -url="https://risk-assessment-service-production.up.railway.app" \
  -duration=2m \
  -users=5 \
  -type=smoke

# Check metrics
curl -s https://risk-assessment-service-production.up.railway.app/metrics

# Monitor for issues
railway logs --service risk-assessment-service --tail 100
```

### Rollback Procedures

#### Automatic Rollback

```bash
# Railway automatic rollback on failure
railway rollback --service risk-assessment-service

# Verify rollback
curl -f https://risk-assessment-service-production.up.railway.app/health
```

#### Manual Rollback

```bash
# Deploy previous version
railway deploy --service risk-assessment-service --version <previous-version>

# Verify rollback
curl -f https://risk-assessment-service-production.up.railway.app/health

# Check logs
railway logs --service risk-assessment-service --tail 100
```

---

## Contact Information

### Team Contacts

| Role | Name | Email | Phone |
|------|------|-------|-------|
| Platform Lead | John Doe | john.doe@kyb-platform.com | +1-XXX-XXX-XXXX |
| DevOps Engineer | Jane Smith | jane.smith@kyb-platform.com | +1-XXX-XXX-XXXX |
| Database Admin | Bob Johnson | bob.johnson@kyb-platform.com | +1-XXX-XXX-XXXX |
| Security Engineer | Alice Brown | alice.brown@kyb-platform.com | +1-XXX-XXX-XXXX |

### Emergency Contacts

| Role | Name | Email | Phone |
|------|------|-------|-------|
| On-Call Engineer | - | oncall@kyb-platform.com | +1-XXX-XXX-XXXX |
| Platform Manager | - | platform-manager@kyb-platform.com | +1-XXX-XXX-XXXX |
| CTO | - | cto@kyb-platform.com | +1-XXX-XXX-XXXX |

### External Contacts

| Service | Contact | Email | Phone |
|---------|---------|-------|-------|
| Railway Support | - | support@railway.app | - |
| Supabase Support | - | support@supabase.com | - |
| Redis Cloud Support | - | support@redis.com | - |

---

**Document Version**: 1.0.0  
**Last Updated**: December 2024  
**Next Review**: March 2025
