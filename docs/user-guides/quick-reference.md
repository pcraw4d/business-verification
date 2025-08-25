# Enhanced Business Intelligence System - Quick Reference Guide

## Table of Contents

1. [Quick Start Commands](#quick-start-commands)
2. [API Reference](#api-reference)
3. [Configuration Quick Reference](#configuration-quick-reference)
4. [Troubleshooting Quick Reference](#troubleshooting-quick-reference)
5. [Development Quick Reference](#development-quick-reference)
6. [Deployment Quick Reference](#deployment-quick-reference)
7. [Monitoring Quick Reference](#monitoring-quick-reference)

## Quick Start Commands

### System Setup

```bash
# Clone repository
git clone https://github.com/your-org/kyb-platform.git
cd kyb-platform

# Setup development environment
cp configs/development.env.example configs/development.env
make dev-start

# Build and run
make build-dev
make run-dev

# Verify setup
curl http://localhost:8080/health
```

### Database Operations

```bash
# Run migrations
make migrate-dev

# Seed development data
make seed-dev

# Reset database
make db-reset

# Backup database
make db-backup

# Restore database
make db-restore backup_file.sql
```

### Testing

```bash
# Run all tests
make test

# Run specific test
go test -run TestService_ClassifyBusiness ./internal/business/classification/

# Run with coverage
make test-coverage

# Run integration tests
make test-integration

# Run benchmarks
make test-benchmark
```

### Code Quality

```bash
# Format code
make fmt

# Run linting
make lint

# Run security checks
make security-check

# Run all quality checks
make quality-check
```

## API Reference

### Authentication

```bash
# Get JWT token
curl -X POST "http://localhost:8080/api/v3/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "user@company.com", "password": "password"}'

# Use API key
curl -H "Authorization: Bearer YOUR_API_KEY" \
  "http://localhost:8080/api/v3/classify"
```

### Business Classification

```bash
# Basic classification
curl -X POST "http://localhost:8080/api/v3/classify" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "business_name": "Acme Corporation",
    "website_url": "https://www.acme.com"
  }'

# Advanced classification
curl -X POST "http://localhost:8080/api/v3/classify/advanced" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "business_name": "Acme Corporation",
    "website_url": "https://www.acme.com",
    "description": "Technology solutions provider",
    "industry_hints": ["technology", "software"],
    "confidence_threshold": 0.8
  }'
```

### Risk Assessment

```bash
# Quick risk assessment
curl -X POST "http://localhost:8080/api/v3/risk/assess" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "business_id": "biz_12345"
  }'

# Comprehensive risk assessment
curl -X POST "http://localhost:8080/api/v3/risk/assess/comprehensive" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "business_id": "biz_12345",
    "assessment_type": "comprehensive",
    "risk_factors": ["industry", "geographic", "size", "compliance"]
  }'
```

### Data Discovery

```bash
# Start data discovery
curl -X POST "http://localhost:8080/api/v3/discovery/start" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "business_id": "biz_12345",
    "discovery_types": ["business_info", "compliance_data"],
    "data_sources": ["business_registry", "news_articles"]
  }'

# Get discovery status
curl -X GET "http://localhost:8080/api/v3/discovery/status/discovery_id" \
  -H "Authorization: Bearer YOUR_API_KEY"

# Get discovery results
curl -X GET "http://localhost:8080/api/v3/discovery/results/discovery_id" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Reports

```bash
# Generate business classification report
curl -X POST "http://localhost:8080/api/v3/reports/classification" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "business_ids": ["biz_12345", "biz_67890"],
    "format": "pdf",
    "include_details": true
  }'

# Generate risk assessment report
curl -X POST "http://localhost:8080/api/v3/reports/risk" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "business_ids": ["biz_12345"],
    "format": "excel",
    "include_recommendations": true
  }'
```

### Health and Monitoring

```bash
# System health
curl http://localhost:8080/health

# Detailed health
curl http://localhost:8080/health/detailed

# Metrics
curl http://localhost:8080/metrics

# Component health
curl http://localhost:8080/health/components
```

## Configuration Quick Reference

### Environment Variables

```bash
# Core Configuration
ENVIRONMENT=production
LOG_LEVEL=info
API_PORT=8080

# Database Configuration
DB_HOST=your-database-host
DB_PORT=5432
DB_NAME=kyb_platform
DB_USER=your-db-user
DB_PASSWORD=your-db-password
DB_SSL_MODE=require

# Redis Configuration
REDIS_HOST=your-redis-host
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password

# Security Configuration
JWT_SECRET=your-jwt-secret
API_KEY_SECRET=your-api-key-secret
ENCRYPTION_KEY=your-encryption-key

# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_API_KEY=your_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
SUPABASE_JWT_SECRET=your_jwt_secret

# Provider Configuration
PROVIDER_DATABASE=supabase
PROVIDER_AUTH=supabase
PROVIDER_CACHE=supabase
PROVIDER_STORAGE=supabase

# Performance Configuration
MAX_CONCURRENT_REQUESTS=1000
REQUEST_TIMEOUT=30s
CACHE_TTL=5m

# Monitoring Configuration
ENABLE_METRICS=true
ENABLE_TRACING=true
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
```

### Docker Configuration

```yaml
# docker-compose.yml
version: "3.8"
services:
  kyb-platform:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=production
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: kyb_platform
      POSTGRES_USER: kyb_user
      POSTGRES_PASSWORD: kyb_password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:6
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### Kubernetes Configuration

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kyb-platform
spec:
  replicas: 3
  selector:
    matchLabels:
      app: kyb-platform
  template:
    metadata:
      labels:
        app: kyb-platform
    spec:
      containers:
      - name: kyb-platform
        image: kyb-platform:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: kyb-platform-config
              key: db_host
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
```

## Troubleshooting Quick Reference

### Common Issues

#### Database Connection Issues

```bash
# Check database status
sudo systemctl status postgresql

# Test database connection
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;"

# Check database logs
sudo tail -f /var/log/postgresql/postgresql-*.log

# Reset database connection
make db-reset
```

#### Redis Connection Issues

```bash
# Check Redis status
sudo systemctl status redis

# Test Redis connection
redis-cli -h $REDIS_HOST -p $REDIS_PORT ping

# Check Redis logs
sudo tail -f /var/log/redis/redis-server.log

# Clear Redis cache
redis-cli -h $REDIS_HOST -p $REDIS_PORT FLUSHALL
```

#### Application Issues

```bash
# Check application logs
docker-compose logs kyb-platform

# Check application status
docker-compose ps

# Restart application
docker-compose restart kyb-platform

# Check resource usage
docker stats
```

#### Performance Issues

```bash
# Check system resources
free -h
df -h
top

# Check application performance
curl http://localhost:8080/metrics

# Profile application
go tool pprof http://localhost:8080/debug/pprof/profile

# Check slow queries
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

### Error Codes

| Error Code | Description | Solution |
|------------|-------------|----------|
| `VALIDATION_ERROR` | Input validation failed | Check input data format and required fields |
| `AUTHENTICATION_ERROR` | Invalid or missing authentication | Verify API key or JWT token |
| `AUTHORIZATION_ERROR` | Insufficient permissions | Check user role and permissions |
| `NOT_FOUND` | Resource not found | Verify resource ID and existence |
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded | Wait and retry, or increase rate limits |
| `INTERNAL_ERROR` | Internal server error | Check application logs and restart if needed |

### Health Check Commands

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health check
curl http://localhost:8080/health/detailed

# Component health check
curl http://localhost:8080/health/components

# Database health check
curl http://localhost:8080/health/database

# Cache health check
curl http://localhost:8080/health/cache
```

## Development Quick Reference

### Development Setup

```bash
# Setup development environment
make dev-setup

# Start development services
make dev-start

# Build development version
make build-dev

# Run development server
make run-dev

# Stop development services
make dev-stop
```

### Code Quality Commands

```bash
# Format code
make fmt

# Run linting
make lint

# Run security checks
make security-check

# Run all quality checks
make quality-check

# Install development tools
make install-tools
```

### Testing Commands

```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run performance tests
make test-performance

# Generate test coverage
make test-coverage

# Run all tests
make test-all

# Run specific test
go test -run TestService_ClassifyBusiness ./internal/business/classification/
```

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Stage changes
git add .

# Commit changes
git commit -m "feat: add new feature"

# Push changes
git push origin feature/your-feature-name

# Create pull request
# (Use GitHub/GitLab web interface)
```

### Debugging Commands

```bash
# Enable debug mode
export DEBUG_MODE=true

# Run with debug logging
LOG_LEVEL=debug make run-dev

# Profile CPU usage
go tool pprof http://localhost:8080/debug/pprof/profile

# Profile memory usage
go tool pprof http://localhost:8080/debug/pprof/heap

# Check goroutines
curl http://localhost:8080/debug/pprof/goroutine
```

## Deployment Quick Reference

### Docker Deployment

```bash
# Build production image
make build-prod

# Run with docker-compose
docker-compose -f docker-compose.prod.yml up -d

# Check deployment status
docker-compose ps

# View logs
docker-compose logs -f kyb-platform

# Stop deployment
docker-compose down
```

### Kubernetes Deployment

```bash
# Apply Kubernetes manifests
kubectl apply -f deployments/kubernetes/

# Check deployment status
kubectl get pods -l app=kyb-platform

# View logs
kubectl logs -f deployment/kyb-platform

# Scale deployment
kubectl scale deployment kyb-platform --replicas=5

# Update deployment
kubectl set image deployment/kyb-platform kyb-platform=kyb-platform:latest
```

### AWS ECS Deployment

```bash
# Register task definition
aws ecs register-task-definition --cli-input-json file://deployments/ecs-task-definition.json

# Update service
aws ecs update-service --cluster kyb-platform --service kyb-platform-api --force-new-deployment

# Check service status
aws ecs describe-services --cluster kyb-platform --services kyb-platform-api

# View logs
aws logs tail /ecs/kyb-platform-api --follow
```

### Railway Deployment

```bash
# Login to Railway
railway login

# Link project
railway link

# Deploy
railway up

# Check deployment status
railway status

# View logs
railway logs

# Open deployment
railway open
```

### Supabase Deployment

```bash
# Install Supabase CLI
npm install -g supabase

# Initialize Supabase project
supabase init

# Link to project
supabase link --project-ref your-project-ref

# Deploy with Supabase
docker-compose -f docker-compose.supabase.yml up -d

# Run database migrations
supabase db push

# Verify deployment
curl http://localhost:8081/health
```

## Monitoring Quick Reference

### Metrics Collection

```bash
# View Prometheus metrics
curl http://localhost:8080/metrics

# Access Prometheus UI
open http://localhost:9090

# Access Grafana UI
open http://localhost:3000

# Check alert manager
open http://localhost:9093
```

### Log Management

```bash
# View application logs
docker-compose logs -f kyb-platform

# Search logs for errors
docker-compose logs kyb-platform | grep ERROR

# View recent logs
docker-compose logs --tail=100 kyb-platform

# Export logs
docker-compose logs kyb-platform > app.log
```

### Performance Monitoring

```bash
# Check system resources
htop
iotop
nethogs

# Monitor application performance
curl http://localhost:8080/metrics | grep -E "(request_duration|error_rate)"

# Check database performance
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

# Check cache performance
redis-cli -h $REDIS_HOST -p $REDIS_PORT info memory
redis-cli -h $REDIS_HOST -p $REDIS_PORT info stats
```

### Alerting

```bash
# Check alert status
curl http://localhost:9090/api/v1/alerts

# View alert rules
cat deployments/prometheus/alerts.yml

# Test alert
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{"labels":{"alertname":"TestAlert"}}]'
```

### Backup and Recovery

```bash
# Create database backup
make db-backup

# Create configuration backup
make config-backup

# Create full system backup
make system-backup

# Restore database
make db-restore backup_file.sql

# Restore configuration
make config-restore backup_file.tar.gz

# Restore full system
make system-restore backup_date
```

---

**Document Version**: 1.0.0  
**Last Updated**: December 19, 2024  
**Next Review**: March 19, 2025
