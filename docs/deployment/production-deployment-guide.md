# KYB Platform - Production Deployment Guide

**Document Version**: 1.0  
**Created**: January 2025  
**Status**: Production Ready  
**Target**: Production Environment Deployment

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Infrastructure Requirements](#infrastructure-requirements)
4. [Security Configuration](#security-configuration)
5. [Environment Setup](#environment-setup)
6. [Deployment Process](#deployment-process)
7. [Monitoring and Alerting](#monitoring-and-alerting)
8. [Backup and Recovery](#backup-and-recovery)
9. [Maintenance Procedures](#maintenance-procedures)
10. [Troubleshooting](#troubleshooting)
11. [Rollback Procedures](#rollback-procedures)

---

## Overview

This guide provides comprehensive instructions for deploying the KYB Platform to a production environment. The deployment includes:

- **High Availability**: Multi-instance deployment with load balancing
- **Security**: Comprehensive security configurations and monitoring
- **Monitoring**: Full observability stack with Prometheus, Grafana, and AlertManager
- **Backup**: Automated backup and disaster recovery procedures
- **Compliance**: SOC 2, PCI DSS, GDPR, and ISO 27001 compliance features

### Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Load Balancer │    │   Load Balancer │
│     (Nginx)     │    │     (Nginx)     │    │     (Nginx)     │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │                           │
            ┌───────▼────────┐        ┌────────▼────────┐
            │  KYB Platform  │        │  KYB Platform   │
            │   Instance 1   │        │   Instance 2    │
            └───────┬────────┘        └────────┬────────┘
                    │                           │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │      Shared Services      │
                    │  ┌─────────────────────┐  │
                    │  │      Supabase       │  │
                    │  │     (Database)      │  │
                    │  └─────────────────────┘  │
                    │  ┌─────────────────────┐  │
                    │  │       Redis         │  │
                    │  │      (Cache)        │  │
                    │  └─────────────────────┘  │
                    │  ┌─────────────────────┐  │
                    │  │    Prometheus       │  │
                    │  │   (Monitoring)      │  │
                    │  └─────────────────────┘  │
                    │  ┌─────────────────────┐  │
                    │  │      Grafana        │  │
                    │  │  (Visualization)    │  │
                    │  └─────────────────────┘  │
                    └───────────────────────────┘
```

---

## Prerequisites

### System Requirements

- **Operating System**: Linux (Ubuntu 20.04+ or CentOS 8+)
- **CPU**: 8+ cores per instance
- **Memory**: 16GB+ RAM per instance
- **Storage**: 100GB+ SSD storage per instance
- **Network**: 1Gbps+ network connection

### Software Requirements

- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **Git**: 2.30+
- **OpenSSL**: 1.1.1+
- **Certbot**: 1.20+ (for SSL certificates)

### External Services

- **Supabase**: Database and authentication service
- **Redis**: Caching and session storage
- **Monitoring**: Prometheus, Grafana, AlertManager
- **Logging**: Elasticsearch, Fluentd (optional)
- **Backup**: S3-compatible storage

---

## Infrastructure Requirements

### Network Configuration

```yaml
# Network topology
networks:
  public:
    subnet: "10.0.1.0/24"
    gateway: "10.0.1.1"
  private:
    subnet: "10.0.2.0/24"
    gateway: "10.0.2.1"
  monitoring:
    subnet: "10.0.3.0/24"
    gateway: "10.0.3.1"

# Firewall rules
firewall:
  inbound:
    - port: 80
      protocol: tcp
      source: 0.0.0.0/0
      description: "HTTP"
    - port: 443
      protocol: tcp
      source: 0.0.0.0/0
      description: "HTTPS"
    - port: 22
      protocol: tcp
      source: "10.0.0.0/8"
      description: "SSH (restricted)"
  outbound:
    - port: 443
      protocol: tcp
      destination: 0.0.0.0/0
      description: "HTTPS outbound"
    - port: 80
      protocol: tcp
      destination: 0.0.0.0/0
      description: "HTTP outbound"
```

### Load Balancer Configuration

```nginx
# /etc/nginx/nginx.conf
upstream kyb_backend {
    least_conn;
    server 10.0.2.10:8080 max_fails=3 fail_timeout=30s;
    server 10.0.2.11:8080 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

server {
    listen 80;
    server_name kyb-platform.com www.kyb-platform.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name kyb-platform.com www.kyb-platform.com;
    
    # SSL Configuration
    ssl_certificate /etc/ssl/certs/kyb-platform.crt;
    ssl_certificate_key /etc/ssl/private/kyb-platform.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    
    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    
    # Rate Limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;
    
    location / {
        proxy_pass http://kyb_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    location /health {
        proxy_pass http://kyb_backend/health;
        access_log off;
    }
    
    location /metrics {
        proxy_pass http://kyb_backend/metrics;
        allow 10.0.3.0/24;
        deny all;
    }
}
```

---

## Security Configuration

### SSL/TLS Setup

```bash
# Generate SSL certificate using Let's Encrypt
sudo certbot certonly --standalone -d kyb-platform.com -d www.kyb-platform.com

# Or use existing certificate
sudo cp your-certificate.crt /etc/ssl/certs/kyb-platform.crt
sudo cp your-private-key.key /etc/ssl/private/kyb-platform.key
sudo chmod 600 /etc/ssl/private/kyb-platform.key
```

### Environment Variables

Create a secure environment file:

```bash
# /etc/kyb-platform/production.env
# Copy from configs/production/production.env and set actual values

# Database
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-supabase-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-supabase-service-role-key

# Security
JWT_SECRET=your-super-secure-jwt-secret-key-here
API_KEYS=prod-api-key-1:user-1,prod-api-key-2:user-2

# Monitoring
GRAFANA_ADMIN_PASSWORD=your-secure-grafana-password
PROMETHEUS_USERNAME=prometheus
PROMETHEUS_PASSWORD=your-secure-prometheus-password

# External Services
REDIS_PASSWORD=your-secure-redis-password
SMTP_HOST=your-smtp-host
SMTP_USERNAME=your-smtp-username
SMTP_PASSWORD=your-smtp-password
SLACK_WEBHOOK_URL=your-slack-webhook-url
PAGERDUTY_INTEGRATION_KEY=your-pagerduty-key

# Backup
BACKUP_S3_BUCKET=your-backup-bucket
BACKUP_S3_REGION=us-east-1
BACKUP_S3_ACCESS_KEY=your-s3-access-key
BACKUP_S3_SECRET_KEY=your-s3-secret-key
```

### Security Hardening

```bash
# System hardening
sudo ufw enable
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Docker security
sudo groupadd docker
sudo usermod -aG docker $USER
sudo systemctl enable docker
sudo systemctl start docker

# File permissions
sudo chmod 600 /etc/kyb-platform/production.env
sudo chown root:root /etc/kyb-platform/production.env
```

---

## Environment Setup

### 1. Clone Repository

```bash
git clone https://github.com/your-org/kyb-platform.git
cd kyb-platform
git checkout production
```

### 2. Create Production Directory Structure

```bash
sudo mkdir -p /opt/kyb-platform/{configs,logs,backups,data}
sudo chown -R $USER:$USER /opt/kyb-platform
```

### 3. Copy Configuration Files

```bash
# Copy production configurations
cp configs/production/* /opt/kyb-platform/configs/
cp docker-compose.production.yml /opt/kyb-platform/
cp Dockerfile.production /opt/kyb-platform/

# Set proper permissions
chmod 600 /opt/kyb-platform/configs/production.env
chmod 644 /opt/kyb-platform/configs/*.yml
```

### 4. Create Monitoring Directories

```bash
mkdir -p /opt/kyb-platform/monitoring/{prometheus,grafana,alertmanager}
mkdir -p /opt/kyb-platform/nginx/{ssl,conf.d}
```

---

## Deployment Process

### 1. Pre-deployment Checks

```bash
# Verify system requirements
./scripts/check-requirements.sh

# Run security scan
./scripts/security-scan.sh

# Validate configuration
./scripts/validate-config.sh
```

### 2. Build Production Image

```bash
# Build the production Docker image
docker build -f Dockerfile.production -t kyb-platform:production .

# Tag for registry
docker tag kyb-platform:production your-registry.com/kyb-platform:production

# Push to registry
docker push your-registry.com/kyb-platform:production
```

### 3. Deploy Services

```bash
# Start services in order
cd /opt/kyb-platform

# Start infrastructure services first
docker-compose -f docker-compose.production.yml up -d redis prometheus grafana alertmanager

# Wait for services to be ready
sleep 30

# Start application services
docker-compose -f docker-compose.production.yml up -d kyb-api

# Verify deployment
docker-compose -f docker-compose.production.yml ps
```

### 4. Health Checks

```bash
# Check application health
curl -f http://localhost:8080/health

# Check metrics endpoint
curl -f http://localhost:9090/metrics

# Check Grafana
curl -f http://localhost:3000/api/health

# Check logs
docker-compose -f docker-compose.production.yml logs kyb-api
```

### 5. Configure Load Balancer

```bash
# Copy nginx configuration
sudo cp nginx/nginx.conf /etc/nginx/nginx.conf

# Test configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

---

## Monitoring and Alerting

### 1. Prometheus Configuration

```yaml
# /opt/kyb-platform/monitoring/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'kyb-platform'
    static_configs:
      - targets: ['kyb-api:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

  - job_name: 'node'
    static_configs:
      - targets: ['node-exporter:9100']
```

### 2. Grafana Dashboards

```bash
# Import dashboards
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $GRAFANA_API_KEY" \
  -d @monitoring/grafana/dashboards/kyb-platform.json \
  http://localhost:3000/api/dashboards/db
```

### 3. Alert Rules

```yaml
# /opt/kyb-platform/monitoring/prometheus/alert_rules.yml
groups:
  - name: kyb-platform
    rules:
      - alert: HighErrorRate
        expr: rate(kyb_prod_http_requests_total{status=~"5.."}[5m]) / rate(kyb_prod_http_requests_total[5m]) > 0.01
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(kyb_prod_http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }}s"

      - alert: ServiceDown
        expr: up{job="kyb-platform"} == 0
        for: 30s
        labels:
          severity: critical
        annotations:
          summary: "KYB Platform service is down"
          description: "Service has been down for more than 30 seconds"
```

---

## Backup and Recovery

### 1. Automated Backup Script

```bash
#!/bin/bash
# /opt/kyb-platform/scripts/backup.sh

BACKUP_DIR="/opt/kyb-platform/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="kyb-platform-backup-$DATE.tar.gz"

# Create backup
docker-compose -f docker-compose.production.yml exec -T kyb-api \
  pg_dump $DATABASE_URL > "$BACKUP_DIR/database-$DATE.sql"

# Compress and encrypt
tar -czf "$BACKUP_DIR/$BACKUP_FILE" \
  -C /opt/kyb-platform \
  configs/ \
  logs/ \
  data/ \
  "$BACKUP_DIR/database-$DATE.sql"

# Upload to S3
aws s3 cp "$BACKUP_DIR/$BACKUP_FILE" \
  "s3://$BACKUP_S3_BUCKET/backups/$BACKUP_FILE"

# Cleanup old backups
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete
```

### 2. Recovery Procedure

```bash
#!/bin/bash
# /opt/kyb-platform/scripts/restore.sh

BACKUP_FILE=$1
BACKUP_DIR="/opt/kyb-platform/backups"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file>"
    exit 1
fi

# Download from S3
aws s3 cp "s3://$BACKUP_S3_BUCKET/backups/$BACKUP_FILE" "$BACKUP_DIR/"

# Extract backup
tar -xzf "$BACKUP_DIR/$BACKUP_FILE" -C /opt/kyb-platform

# Restore database
docker-compose -f docker-compose.production.yml exec -T kyb-api \
  psql $DATABASE_URL < "$BACKUP_DIR/database-*.sql"

# Restart services
docker-compose -f docker-compose.production.yml restart
```

---

## Maintenance Procedures

### 1. Regular Maintenance Tasks

```bash
# Daily tasks
./scripts/daily-maintenance.sh

# Weekly tasks
./scripts/weekly-maintenance.sh

# Monthly tasks
./scripts/monthly-maintenance.sh
```

### 2. Update Procedure

```bash
# 1. Create backup
./scripts/backup.sh

# 2. Pull latest code
git pull origin production

# 3. Build new image
docker build -f Dockerfile.production -t kyb-platform:production .

# 4. Rolling update
docker-compose -f docker-compose.production.yml up -d --no-deps kyb-api

# 5. Verify deployment
./scripts/health-check.sh

# 6. Cleanup old images
docker image prune -f
```

### 3. Log Rotation

```bash
# Configure logrotate
sudo tee /etc/logrotate.d/kyb-platform << EOF
/opt/kyb-platform/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 kyb kyb
    postrotate
        docker-compose -f /opt/kyb-platform/docker-compose.production.yml restart kyb-api
    endscript
}
EOF
```

---

## Troubleshooting

### Common Issues

#### 1. Service Won't Start

```bash
# Check logs
docker-compose -f docker-compose.production.yml logs kyb-api

# Check configuration
docker-compose -f docker-compose.production.yml config

# Check resources
docker stats
```

#### 2. High Memory Usage

```bash
# Check memory usage
docker stats kyb-api

# Check for memory leaks
docker exec kyb-api go tool pprof http://localhost:6060/debug/pprof/heap

# Restart service
docker-compose -f docker-compose.production.yml restart kyb-api
```

#### 3. Database Connection Issues

```bash
# Test database connection
docker exec kyb-api pg_isready -h $SUPABASE_HOST -p 5432

# Check connection pool
curl http://localhost:8080/metrics | grep database_connections
```

#### 4. SSL Certificate Issues

```bash
# Check certificate validity
openssl x509 -in /etc/ssl/certs/kyb-platform.crt -text -noout

# Renew certificate
sudo certbot renew --dry-run
```

### Performance Tuning

#### 1. Database Optimization

```sql
-- Check slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- Add indexes
CREATE INDEX CONCURRENTLY idx_merchants_created_at ON merchants(created_at);
CREATE INDEX CONCURRENTLY idx_audit_logs_timestamp ON audit_logs(timestamp);
```

#### 2. Application Optimization

```bash
# Enable profiling
export PROFILING_ENABLED=true
docker-compose -f docker-compose.production.yml restart kyb-api

# Check performance metrics
curl http://localhost:6060/debug/pprof/profile?seconds=30 > profile.out
go tool pprof profile.out
```

---

## Rollback Procedures

### 1. Quick Rollback

```bash
# Rollback to previous version
docker-compose -f docker-compose.production.yml down
docker tag kyb-platform:production kyb-platform:rollback
docker tag kyb-platform:previous kyb-platform:production
docker-compose -f docker-compose.production.yml up -d
```

### 2. Database Rollback

```bash
# Restore from backup
./scripts/restore.sh kyb-platform-backup-YYYYMMDD_HHMMSS.tar.gz
```

### 3. Configuration Rollback

```bash
# Restore configuration
cp /opt/kyb-platform/configs/production.env.backup /opt/kyb-platform/configs/production.env
docker-compose -f docker-compose.production.yml restart
```

---

## Security Checklist

### Pre-deployment Security

- [ ] SSL certificates installed and valid
- [ ] Environment variables secured
- [ ] Firewall rules configured
- [ ] Security headers enabled
- [ ] Rate limiting configured
- [ ] Authentication enabled
- [ ] Authorization configured
- [ ] Audit logging enabled

### Post-deployment Security

- [ ] Security scan completed
- [ ] Vulnerability assessment done
- [ ] Penetration testing scheduled
- [ ] Monitoring alerts configured
- [ ] Backup procedures tested
- [ ] Incident response plan ready
- [ ] Security documentation updated

---

## Compliance Checklist

### SOC 2 Compliance

- [ ] Access controls implemented
- [ ] Change management process
- [ ] Risk assessment completed
- [ ] Security monitoring enabled
- [ ] Incident response procedures
- [ ] Business continuity plan
- [ ] Data protection measures
- [ ] System operations documented

### PCI DSS Compliance

- [ ] Secure network configuration
- [ ] Cardholder data protection
- [ ] Vulnerability management
- [ ] Access control measures
- [ ] Network monitoring
- [ ] Security policy documented

### GDPR Compliance

- [ ] Data protection by design
- [ ] Data minimization implemented
- [ ] Purpose limitation enforced
- [ ] Storage limitation configured
- [ ] Data accuracy measures
- [ ] Integrity and confidentiality
- [ ] Accountability measures

---

## Support and Contacts

### Emergency Contacts

- **On-call Engineer**: +1-XXX-XXX-XXXX
- **Security Team**: security@kyb-platform.com
- **DevOps Team**: devops@kyb-platform.com
- **Management**: management@kyb-platform.com

### Documentation

- **API Documentation**: https://docs.kyb-platform.com/api
- **User Guide**: https://docs.kyb-platform.com/user
- **Developer Guide**: https://docs.kyb-platform.com/developer
- **Architecture**: https://docs.kyb-platform.com/architecture

### Monitoring Dashboards

- **Grafana**: https://grafana.kyb-platform.com
- **Prometheus**: https://prometheus.kyb-platform.com
- **AlertManager**: https://alerts.kyb-platform.com

---

**Document Version**: 1.0  
**Last Updated**: January 2025  
**Next Review**: April 2025
