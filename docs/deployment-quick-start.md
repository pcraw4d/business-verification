# Enhanced Business Intelligence System - Deployment Quick Start Guide

## Overview

This quick-start guide provides step-by-step instructions for deploying the Enhanced Business Intelligence System in common scenarios. For comprehensive documentation, refer to `deployment-documentation.md`.

## Quick Deployment Options

### 1. Local Development (Docker Compose)

**Prerequisites**: Docker and Docker Compose

```bash
# Clone the repository
git clone <repository-url>
cd kyb-platform

# Copy environment configuration
cp configs/development.env.example configs/development.env

# Start all services
docker-compose up -d

# Verify deployment
curl http://localhost:8080/health

# View logs
docker-compose logs -f kyb-platform
```

**Access Points**:
- API: http://localhost:8080
- Health Check: http://localhost:8080/health
- Metrics: http://localhost:8080/metrics

### 2. Railway Deployment (Recommended for Beta/Staging)

**Prerequisites**: Railway account and CLI

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login to Railway
railway login

# Link project
railway link

# Set environment variables
railway variables set ENVIRONMENT=production
railway variables set DB_HOST=your-db-host
railway variables set REDIS_HOST=your-redis-host

# Deploy
railway up

# Open deployment
railway open
```

### 3. Kubernetes Deployment (Production)

**Prerequisites**: kubectl configured with cluster access

```bash
# Create namespace
kubectl create namespace kyb-platform

# Apply all resources
kubectl apply -f deployments/kubernetes/

# Check deployment status
kubectl get pods -n kyb-platform

# Get service URL
kubectl get svc -n kyb-platform
```

### 4. AWS ECS Deployment (Enterprise)

**Prerequisites**: AWS CLI configured

```bash
# Register task definition
aws ecs register-task-definition --cli-input-json file://deployments/ecs-task-definition.json

# Create service
aws ecs create-service \
  --cluster kyb-platform-cluster \
  --service-name kyb-platform-api \
  --task-definition kyb-platform-api:1 \
  --desired-count 3 \
  --launch-type FARGATE
```

### 5. Supabase Deployment (Recommended for MVP)

**Prerequisites**: Supabase account and project

```bash
# Create Supabase project at https://supabase.com
# Note your project URL and API keys

# Install Supabase CLI
npm install -g supabase

# Initialize Supabase project
supabase init

# Link to your Supabase project
supabase link --project-ref your-project-ref

# Set environment variables
export SUPABASE_URL=https://your-project.supabase.co
export SUPABASE_API_KEY=your_anon_key
export SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
export SUPABASE_JWT_SECRET=your_jwt_secret

# Start with Supabase
docker-compose -f docker-compose.supabase.yml up -d

# Run database migrations
supabase db push

# Verify deployment
curl http://localhost:8081/health
```

## Environment Configuration

### Required Environment Variables

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

# Redis Configuration
REDIS_HOST=your-redis-host
REDIS_PORT=6379

# Security
JWT_SECRET=your-jwt-secret
API_KEY_SECRET=your-api-key-secret

# Monitoring
ENABLE_METRICS=true
ENABLE_TRACING=true
```

### Supabase Environment Variables

```bash
# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_API_KEY=your_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
SUPABASE_JWT_SECRET=your_jwt_secret

# Database Configuration (Supabase PostgreSQL)
DB_HOST=db.your-project.supabase.co
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=your_db_password
DB_SSL_MODE=require

# Provider Configuration
PROVIDER_DATABASE=supabase
PROVIDER_AUTH=supabase
PROVIDER_CACHE=supabase
PROVIDER_STORAGE=supabase
```

### Optional Environment Variables

```bash
# Performance Tuning
MAX_CONCURRENT_REQUESTS=1000
REQUEST_TIMEOUT=30s
CACHE_TTL=5m

# External Services
EXTERNAL_API_TIMEOUT=10s
EXTERNAL_API_RETRIES=3

# Feature Flags
ENABLE_BETA_FEATURES=false
ENABLE_DEBUG_MODE=false
```

## Health Checks

### Basic Health Check

```bash
# Check application health
curl http://your-domain/health

# Expected response
{
  "status": "healthy",
  "timestamp": "2024-12-19T10:30:00Z",
  "version": "1.0.0",
  "uptime": "2h30m15s"
}
```

### Detailed Health Check

```bash
# Check detailed health status
curl http://your-domain/health/detailed

# Expected response
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected",
  "external_apis": "available",
  "modules": {
    "classification": "healthy",
    "caching": "healthy",
    "monitoring": "healthy"
  }
}
```

## Monitoring Setup

### Prometheus Metrics

```bash
# Access metrics endpoint
curl http://your-domain/metrics

# Configure Prometheus to scrape metrics
# Add to prometheus.yml:
scrape_configs:
  - job_name: 'kyb-platform'
    static_configs:
      - targets: ['your-domain:8080']
    metrics_path: '/metrics'
```

### Grafana Dashboard

Import the provided Grafana dashboard configuration:

```bash
# Dashboard configuration available in:
# monitoring/grafana-dashboards/kyb-platform-dashboard.json
```

## Common Deployment Issues

### Issue: Application Won't Start

**Symptoms**: Container exits immediately or health checks fail

**Solutions**:
```bash
# Check logs
docker-compose logs kyb-platform
# or
kubectl logs deployment/kyb-platform-api -n kyb-platform

# Verify environment variables
docker-compose exec kyb-platform env | grep DB_
# or
kubectl exec deployment/kyb-platform-api -n kyb-platform -- env | grep DB_

# Check database connectivity
docker-compose exec kyb-platform nc -zv $DB_HOST $DB_PORT
```

### Issue: High Memory Usage

**Symptoms**: Application crashes or becomes unresponsive

**Solutions**:
```bash
# Check memory usage
docker stats
# or
kubectl top pods -n kyb-platform

# Analyze memory usage
curl http://your-domain/debug/pprof/heap

# Adjust memory limits in deployment configuration
```

### Issue: Database Connection Errors

**Symptoms**: 500 errors or database connection timeouts

**Solutions**:
```bash
# Test database connectivity
docker-compose exec kyb-platform nc -zv $DB_HOST $DB_PORT

# Check database logs
docker-compose logs postgres

# Verify database credentials
docker-compose exec kyb-platform psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;"
```

## Performance Optimization

### Resource Limits

```yaml
# Kubernetes resource limits
resources:
  requests:
    memory: "512Mi"
    cpu: "250m"
  limits:
    memory: "1Gi"
    cpu: "500m"
```

### Scaling Configuration

```yaml
# Horizontal Pod Autoscaler
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: kyb-platform-hpa
spec:
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Security Checklist

### Pre-Deployment Security

- [ ] Environment variables are properly set
- [ ] Secrets are stored securely (not in code)
- [ ] SSL/TLS certificates are configured
- [ ] Network policies are applied
- [ ] Security headers are enabled
- [ ] Rate limiting is configured

### Post-Deployment Security

- [ ] Health checks are passing
- [ ] Monitoring is working
- [ ] Logs are being collected
- [ ] Backup procedures are tested
- [ ] Access controls are verified

## Troubleshooting Commands

### Docker Compose

```bash
# View all logs
docker-compose logs

# View specific service logs
docker-compose logs kyb-platform

# Restart services
docker-compose restart

# Rebuild and restart
docker-compose up --build -d

# Check service status
docker-compose ps
```

### Kubernetes

```bash
# Check pod status
kubectl get pods -n kyb-platform

# View pod logs
kubectl logs -f deployment/kyb-platform-api -n kyb-platform

# Describe pod for details
kubectl describe pod <pod-name> -n kyb-platform

# Execute commands in pod
kubectl exec -it <pod-name> -n kyb-platform -- /bin/sh

# Check service endpoints
kubectl get endpoints -n kyb-platform
```

### AWS ECS

```bash
# List services
aws ecs list-services --cluster kyb-platform-cluster

# Describe service
aws ecs describe-services --cluster kyb-platform-cluster --services kyb-platform-api

# View service logs
aws logs tail /ecs/kyb-platform-api --follow

# Update service
aws ecs update-service --cluster kyb-platform-cluster --service kyb-platform-api --force-new-deployment
```

## Support and Resources

### Documentation

- **Comprehensive Guide**: `docs/deployment-documentation.md`
- **API Reference**: `docs/code-documentation/api-reference.md`
- **Module Documentation**: `docs/code-documentation/module-documentation.md`

### Scripts and Tools

- **Deployment Scripts**: `scripts/deploy-*.sh`
- **Health Check Scripts**: `scripts/check-*.sh`
- **Performance Testing**: `scripts/performance-*.sh`

### Monitoring and Logs

- **Application Metrics**: `/metrics` endpoint
- **Health Status**: `/health` endpoint
- **Debug Information**: `/debug/pprof/*` endpoints

---

**Document Version**: 1.0.0  
**Last Updated**: December 19, 2024  
**Next Review**: March 19, 2025
