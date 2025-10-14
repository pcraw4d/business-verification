# Disaster Recovery Runbook
## Risk Assessment Service - KYB Platform

### Document Information
- **Version**: 1.0.0
- **Last Updated**: December 2024
- **Next Review**: March 2025
- **Owner**: DevOps Team
- **Approved By**: CTO

---

## Table of Contents

1. [Overview](#overview)
2. [Recovery Objectives](#recovery-objectives)
3. [Pre-Incident Preparation](#pre-incident-preparation)
4. [Incident Response Procedures](#incident-response-procedures)
5. [Database Recovery](#database-recovery)
6. [Redis Recovery](#redis-recovery)
7. [Service Recovery](#service-recovery)
8. [Data Validation](#data-validation)
9. [Communication Procedures](#communication-procedures)
10. [Post-Incident Review](#post-incident-review)
11. [Appendices](#appendices)

---

## Overview

This runbook provides step-by-step procedures for disaster recovery of the Risk Assessment Service within the KYB Platform. It covers recovery from various failure scenarios including database corruption, Redis failures, service outages, and complete infrastructure loss.

### Scope
- Risk Assessment Service
- Supabase Database
- Redis Cache
- Railway Infrastructure
- External API Dependencies

### Assumptions
- Backup systems are operational
- Recovery team has necessary access credentials
- Communication channels are available
- Monitoring systems are functional

---

## Recovery Objectives

### Recovery Time Objectives (RTO)
| Environment | RTO | Description |
|-------------|-----|-------------|
| Production | 1 hour | Critical business operations |
| Staging | 2 hours | Development and testing |
| Development | 4 hours | Development work |

### Recovery Point Objectives (RPO)
| Environment | RPO | Description |
|-------------|-----|-------------|
| Production | 1 minute | Minimal data loss |
| Staging | 4 hours | Acceptable for testing |
| Development | 24 hours | Acceptable for development |

### Service Level Agreements (SLA)
- **Availability**: 99.9% uptime
- **Response Time**: < 1 second (95th percentile)
- **Data Integrity**: 100% consistency
- **Security**: No data breaches during recovery

---

## Pre-Incident Preparation

### 1. Access and Credentials
Ensure the following are available and tested:

```bash
# Railway CLI access
railway login --token $RAILWAY_TOKEN

# Supabase access
supabase login
supabase projects list

# Redis access
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD

# Database access
psql $DATABASE_URL
```

### 2. Backup Verification
Regularly verify backup integrity:

```bash
# Check Supabase backups
supabase projects list
supabase projects get-backups <project-id>

# Check Redis backups
ls -la /backups/redis/
redis-cli --rdb /backups/redis/dump-$(date +%Y%m%d).rdb

# Verify backup restoration
./scripts/verify-backup-restore.sh
```

### 3. Monitoring Setup
Ensure monitoring is configured for:
- Service health checks
- Database connectivity
- Redis connectivity
- External API availability
- Backup status

### 4. Communication Channels
- **Slack**: #incidents-kyb-platform
- **Email**: incidents@kyb-platform.com
- **PagerDuty**: Risk Assessment Service
- **Status Page**: status.kyb-platform.com

---

## Incident Response Procedures

### 1. Initial Assessment (0-5 minutes)

#### 1.1 Identify the Incident
```bash
# Check service status
curl -f https://risk-assessment-service-production.up.railway.app/health

# Check database connectivity
psql $DATABASE_URL -c "SELECT 1;"

# Check Redis connectivity
redis-cli -h $REDIS_HOST ping

# Check Railway service status
railway status
```

#### 1.2 Classify the Incident
| Severity | Description | Response Time |
|----------|-------------|---------------|
| P1 - Critical | Complete service outage | 15 minutes |
| P2 - High | Partial service degradation | 30 minutes |
| P3 - Medium | Performance issues | 2 hours |
| P4 - Low | Minor issues | 24 hours |

#### 1.3 Activate Response Team
- **P1/P2**: Full team activation
- **P3/P4**: On-call engineer

### 2. Incident Communication (0-10 minutes)

#### 2.1 Internal Communication
```bash
# Send Slack notification
curl -X POST -H 'Content-type: application/json' \
  --data '{"text":"🚨 INCIDENT: Risk Assessment Service - Severity P1"}' \
  $SLACK_WEBHOOK_URL

# Update status page
curl -X POST -H 'Content-type: application/json' \
  --data '{"status":"investigating","message":"Service degradation detected"}' \
  $STATUS_PAGE_API
```

#### 2.2 Customer Communication
- Update status page
- Send email notifications to affected customers
- Post updates on social media if necessary

### 3. Investigation and Diagnosis (5-30 minutes)

#### 3.1 Check Service Logs
```bash
# Railway logs
railway logs --service risk-assessment-service --tail 100

# Application logs
kubectl logs -l app=risk-assessment-service --tail=100

# Database logs
supabase projects logs <project-id>
```

#### 3.2 Check Metrics and Alerts
- Prometheus metrics
- Grafana dashboards
- Alert manager notifications
- External monitoring services

#### 3.3 Identify Root Cause
Common failure scenarios:
- Database connection issues
- Redis connectivity problems
- External API failures
- Service deployment issues
- Infrastructure problems

---

## Database Recovery

### 1. Supabase Database Recovery

#### 1.1 Point-in-Time Recovery (PITR)
```bash
# List available backups
supabase projects get-backups <project-id>

# Restore to specific point in time
supabase projects restore-backup <project-id> \
  --backup-id <backup-id> \
  --target-time "2024-12-01T10:30:00Z"

# Verify restoration
psql $DATABASE_URL -c "SELECT COUNT(*) FROM risk_assessments;"
```

#### 1.2 Full Database Restore
```bash
# Stop the service
railway service stop risk-assessment-service

# Restore from latest backup
supabase projects restore-backup <project-id> \
  --backup-id <latest-backup-id>

# Run database migrations
./scripts/run-supabase-migrations.sh --environment production

# Verify data integrity
./scripts/verify-database-integrity.sh

# Restart the service
railway service start risk-assessment-service
```

#### 1.3 Table-Level Recovery
```bash
# Restore specific tables
psql $DATABASE_URL -c "
  DROP TABLE IF EXISTS risk_assessments CASCADE;
  CREATE TABLE risk_assessments AS 
  SELECT * FROM risk_assessments_backup;
"

# Recreate indexes
psql $DATABASE_URL -f supabase-migrations/risk-assessment-indexes.sql

# Verify table restoration
psql $DATABASE_URL -c "SELECT COUNT(*) FROM risk_assessments;"
```

### 2. Data Validation

#### 2.1 Row Count Verification
```bash
#!/bin/bash
# verify-row-counts.sh

TABLES=(
  "risk_assessments"
  "risk_predictions"
  "risk_factors"
  "custom_risk_models"
  "batch_risk_jobs"
  "webhooks"
  "webhook_deliveries"
  "dashboards"
  "reports"
  "report_templates"
  "scheduled_reports"
)

for table in "${TABLES[@]}"; do
  count=$(psql $DATABASE_URL -t -c "SELECT COUNT(*) FROM $table;")
  echo "$table: $count rows"
done
```

#### 2.2 Data Integrity Checks
```bash
#!/bin/bash
# verify-data-integrity.sh

# Check foreign key constraints
psql $DATABASE_URL -c "
  SELECT conname, conrelid::regclass, confrelid::regclass
  FROM pg_constraint
  WHERE contype = 'f' AND NOT convalidated;
"

# Check for orphaned records
psql $DATABASE_URL -c "
  SELECT COUNT(*) as orphaned_predictions
  FROM risk_predictions rp
  LEFT JOIN risk_assessments ra ON rp.risk_assessment_id = ra.id
  WHERE ra.id IS NULL;
"
```

---

## Redis Recovery

### 1. Redis Data Recovery

#### 1.1 RDB File Recovery
```bash
# Stop Redis service
railway service stop redis

# Backup current data directory
cp -r /data /data-backup-$(date +%Y%m%d-%H%M%S)

# Restore from RDB file
cp /backups/redis/dump-20241201.rdb /data/dump.rdb

# Start Redis service
railway service start redis

# Verify data
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD \
  DBSIZE
```

#### 1.2 AOF File Recovery
```bash
# Stop Redis service
railway service stop redis

# Restore AOF file
cp /backups/redis/appendonly-20241201.aof /data/appendonly.aof

# Start Redis service
railway service start redis

# Verify data
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD \
  INFO persistence
```

#### 1.3 Redis Cluster Recovery
```bash
# Check cluster status
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD \
  CLUSTER NODES

# Restore cluster from backup
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD \
  CLUSTER MEET <node-ip> <node-port>

# Verify cluster health
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD \
  CLUSTER INFO
```

### 2. Redis Cache Warming

#### 2.1 Preload Critical Data
```bash
#!/bin/bash
# warm-redis-cache.sh

# Preload rate limit data
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD \
  SET "rate_limit:free:default" "0"

# Preload ML model cache
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD \
  SET "ml_model:risk_assessment:v1" "loaded"

# Preload configuration cache
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD \
  SET "config:risk_assessment" "$(cat config.json)"
```

---

## Service Recovery

### 1. Railway Service Recovery

#### 1.1 Service Restart
```bash
# Check service status
railway status

# Restart service
railway service restart risk-assessment-service

# Check logs
railway logs --service risk-assessment-service --tail 50
```

#### 1.2 Service Rollback
```bash
# List available deployments
railway deployments list

# Rollback to previous version
railway rollback --service risk-assessment-service

# Verify rollback
curl -f https://risk-assessment-service-production.up.railway.app/health
```

#### 1.3 Service Redeployment
```bash
# Deploy from specific commit
railway deploy --service risk-assessment-service \
  --commit <commit-hash>

# Deploy with environment variables
railway deploy --service risk-assessment-service \
  --env-file .env.production
```

### 2. External API Recovery

#### 2.1 API Endpoint Testing
```bash
#!/bin/bash
# test-external-apis.sh

APIS=(
  "https://api.newsapi.org/v2/everything"
  "https://api.opencorporates.com/v0.4/companies/search"
  "https://api.thomsonreuters.com/risk"
)

for api in "${APIS[@]}"; do
  echo "Testing $api"
  curl -f -w "%{http_code}\n" -o /dev/null -s "$api"
done
```

#### 2.2 Circuit Breaker Reset
```bash
# Reset circuit breakers
curl -X POST https://risk-assessment-service-production.up.railway.app/api/v1/admin/circuit-breaker/reset
```

---

## Data Validation

### 1. Automated Validation Scripts

#### 1.1 Complete System Validation
```bash
#!/bin/bash
# validate-system-recovery.sh

echo "Starting system validation..."

# Test database connectivity
psql $DATABASE_URL -c "SELECT 1;" || exit 1

# Test Redis connectivity
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping || exit 1

# Test service health
curl -f https://risk-assessment-service-production.up.railway.app/health || exit 1

# Test API endpoints
curl -f -X POST https://risk-assessment-service-production.up.railway.app/api/v1/assess \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TEST_API_KEY" \
  -d '{"business_name": "Test Company", "business_address": "123 Test St"}' || exit 1

echo "System validation completed successfully"
```

#### 1.2 Data Consistency Validation
```bash
#!/bin/bash
# validate-data-consistency.sh

# Check data freshness
latest_assessment=$(psql $DATABASE_URL -t -c "
  SELECT MAX(created_at) FROM risk_assessments;
")

echo "Latest risk assessment: $latest_assessment"

# Check data completeness
missing_data=$(psql $DATABASE_URL -t -c "
  SELECT COUNT(*) FROM risk_assessments 
  WHERE business_name IS NULL OR business_address IS NULL;
")

if [ "$missing_data" -gt 0 ]; then
  echo "WARNING: $missing_data assessments with missing data"
fi
```

### 2. Performance Validation

#### 2.1 Load Testing
```bash
# Run load test
go run ./cmd/load_test.go \
  -url="https://risk-assessment-service-production.up.railway.app" \
  -duration=5m \
  -users=100 \
  -type=load
```

#### 2.2 Response Time Validation
```bash
# Test response times
for i in {1..10}; do
  time curl -f https://risk-assessment-service-production.up.railway.app/health
done
```

---

## Communication Procedures

### 1. Incident Status Updates

#### 1.1 Update Template
```
🚨 INCIDENT UPDATE - Risk Assessment Service

Status: [INVESTIGATING/IDENTIFIED/MONITORING/RESOLVED]
Severity: P[1-4]
Duration: [X] minutes
Impact: [Description of impact]

Current Status:
- [What we know]
- [What we're doing]
- [Next steps]

ETA for Resolution: [Time]

Updates will be provided every [X] minutes.
```

#### 1.2 Communication Schedule
- **P1**: Every 15 minutes
- **P2**: Every 30 minutes
- **P3**: Every 2 hours
- **P4**: Every 24 hours

### 2. Customer Communication

#### 2.1 Status Page Updates
```json
{
  "status": "investigating",
  "message": "We are currently investigating an issue with our Risk Assessment Service. We will provide updates as we learn more.",
  "incident_id": "INC-2024-001"
}
```

#### 2.2 Email Notifications
```
Subject: Service Incident - Risk Assessment Service

Dear Customer,

We are currently experiencing an issue with our Risk Assessment Service that may be affecting your ability to perform risk assessments.

We are actively working to resolve this issue and will provide updates as they become available.

We apologize for any inconvenience this may cause.

Best regards,
KYB Platform Team
```

---

## Post-Incident Review

### 1. Incident Analysis

#### 1.1 Timeline Documentation
- Incident start time
- Detection time
- Response time
- Resolution time
- Total downtime

#### 1.2 Root Cause Analysis
- What happened?
- Why did it happen?
- What could have prevented it?
- What will we do differently?

### 2. Improvement Actions

#### 2.1 Immediate Actions
- Fix the root cause
- Update monitoring
- Improve alerting
- Update documentation

#### 2.2 Long-term Actions
- Process improvements
- Technology upgrades
- Training requirements
- Policy updates

### 3. Lessons Learned

#### 3.1 What Went Well
- Effective communication
- Quick response time
- Good teamwork
- Successful recovery

#### 3.2 What Could Be Improved
- Faster detection
- Better monitoring
- More automation
- Clearer procedures

---

## Appendices

### Appendix A: Contact Information

#### Emergency Contacts
- **On-Call Engineer**: +1-XXX-XXX-XXXX
- **DevOps Lead**: +1-XXX-XXX-XXXX
- **CTO**: +1-XXX-XXX-XXXX
- **Customer Support**: support@kyb-platform.com

#### External Vendors
- **Railway Support**: support@railway.app
- **Supabase Support**: support@supabase.com
- **Redis Support**: support@redis.com

### Appendix B: Useful Commands

#### Database Commands
```bash
# Connect to database
psql $DATABASE_URL

# Check database size
psql $DATABASE_URL -c "SELECT pg_size_pretty(pg_database_size(current_database()));"

# Check table sizes
psql $DATABASE_URL -c "
  SELECT schemaname,tablename,pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
  FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
"
```

#### Redis Commands
```bash
# Connect to Redis
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD

# Check memory usage
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD INFO memory

# Check key count
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD DBSIZE

# Monitor commands
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD MONITOR
```

#### Railway Commands
```bash
# Check service status
railway status

# View logs
railway logs --service risk-assessment-service

# Check deployments
railway deployments list

# Restart service
railway service restart risk-assessment-service
```

### Appendix C: Monitoring Dashboards

#### Key Dashboards
- **Service Health**: https://grafana.kyb-platform.com/d/service-health
- **Database Performance**: https://grafana.kyb-platform.com/d/database-performance
- **Redis Metrics**: https://grafana.kyb-platform.com/d/redis-metrics
- **Business Metrics**: https://grafana.kyb-platform.com/d/business-metrics

#### Alert Channels
- **Slack**: #alerts-kyb-platform
- **PagerDuty**: Risk Assessment Service
- **Email**: alerts@kyb-platform.com

### Appendix D: Backup Locations

#### Database Backups
- **Supabase Managed**: Automatic daily backups
- **External S3**: s3://kyb-platform-backups/database/
- **Local**: /backups/database/

#### Redis Backups
- **RDB Files**: /backups/redis/rdb/
- **AOF Files**: /backups/redis/aof/
- **External S3**: s3://kyb-platform-backups/redis/

#### Configuration Backups
- **Git Repository**: https://github.com/kyb-platform/configs
- **Local**: /backups/configs/
- **External S3**: s3://kyb-platform-backups/configs/

---

**Document End**

*This runbook should be reviewed and updated quarterly or after any significant infrastructure changes.*
