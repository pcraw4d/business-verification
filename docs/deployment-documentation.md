# Enhanced Business Intelligence System - Deployment Documentation

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Environment Setup](#environment-setup)
4. [Deployment Methods](#deployment-methods)
5. [Configuration Management](#configuration-management)
6. [Monitoring and Health Checks](#monitoring-and-health-checks)
7. [Security and Compliance](#security-and-compliance)
8. [Scaling and Performance](#scaling-and-performance)
9. [Backup and Disaster Recovery](#backup-and-disaster-recovery)
10. [Troubleshooting](#troubleshooting)
11. [Maintenance and Updates](#maintenance-and-updates)

## Overview

The Enhanced Business Intelligence System is designed for deployment across multiple environments with support for various deployment methods including Docker containers, Kubernetes, AWS ECS, cloud platforms like Railway, and database-as-a-service platforms like Supabase.

### System Requirements

- **Go Runtime**: 1.22 or later
- **Database**: PostgreSQL 13+ or Supabase (recommended for MVP)
- **Cache**: Redis 6+ or Supabase cache
- **Memory**: Minimum 512MB, Recommended 2GB+
- **CPU**: Minimum 1 core, Recommended 2+ cores
- **Storage**: Minimum 10GB for logs and data

## Prerequisites

### Development Environment

```bash
# Install Go 1.22+
wget https://golang.org/dl/go1.22.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install kubectl (for Kubernetes deployment)
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```

### Production Environment

```bash
# Install system dependencies
sudo apt-get update
sudo apt-get install -y curl wget git build-essential

# Install monitoring tools
sudo apt-get install -y prometheus-node-exporter
sudo systemctl enable prometheus-node-exporter
```

## Environment Setup

### Environment Variables

Create environment-specific configuration files:

```bash
# Development
cp configs/development.env.example configs/development.env

# Production
cp configs/production.env.example configs/production.env

# Staging
cp configs/staging.env.example configs/staging.env
```

### Database Setup

```sql
-- Initialize database schema
\i scripts/init-db.sql

-- Run migrations
\i scripts/run_migrations.sh
```

### Redis Configuration

```bash
# Install Redis
sudo apt-get install redis-server

# Configure Redis for production
sudo cp configs/redis.conf /etc/redis/redis.conf
sudo systemctl restart redis
```

## Deployment Methods

### 1. Docker Deployment

#### Build Docker Image

```bash
# Build production image
docker build -t kyb-platform:latest .

# Build with specific version
docker build -t kyb-platform:v1.0.0 .
```

#### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  kyb-platform:
    image: kyb-platform:latest
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=production
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: kyb_platform
      POSTGRES_USER: kyb_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

#### Run with Docker Compose

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f kyb-platform

# Stop services
docker-compose down
```

### 2. Kubernetes Deployment

#### Namespace Setup

```bash
# Create namespace
kubectl create namespace kyb-platform

# Apply namespace configuration
kubectl apply -f deployments/kubernetes/namespace.yaml
```

#### Deploy Application

```bash
# Apply all Kubernetes resources
kubectl apply -f deployments/kubernetes/

# Check deployment status
kubectl get pods -n kyb-platform

# View logs
kubectl logs -f deployment/kyb-platform-api -n kyb-platform
```

#### Kubernetes Configuration

```yaml
# deployments/kubernetes/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kyb-platform-config
  namespace: kyb-platform
data:
  db_host: "kyb-platform-db"
  db_port: "5432"
  db_name: "kyb_platform"
  redis_host: "kyb-platform-redis"
  redis_port: "6379"
  log_level: "info"
  environment: "production"
```

```yaml
# deployments/kubernetes/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: kyb-platform-secrets
  namespace: kyb-platform
type: Opaque
data:
  db_user: <base64-encoded-username>
  db_password: <base64-encoded-password>
  jwt_secret: <base64-encoded-jwt-secret>
  api_key_secret: <base64-encoded-api-key-secret>
```

### 3. AWS ECS Deployment

#### ECS Task Definition

```json
{
  "family": "kyb-platform-api",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "executionRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT_ID:role/kyb-platform-task-role",
  "containerDefinitions": [
    {
      "name": "kyb-platform-api",
      "image": "ghcr.io/REPOSITORY/kyb-platform:latest",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "ENVIRONMENT",
          "value": "production"
        }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:ACCOUNT_ID:secret:kyb-platform/db-password"
        }
      ],
      "healthCheck": {
        "command": [
          "CMD-SHELL",
          "curl -f http://localhost:8080/health || exit 1"
        ],
        "interval": 30,
        "timeout": 5,
        "retries": 3
      }
    }
  ]
}
```

#### Deploy to ECS

```bash
# Register task definition
aws ecs register-task-definition --cli-input-json file://deployments/ecs-task-definition.json

# Create service
aws ecs create-service \
  --cluster kyb-platform-cluster \
  --service-name kyb-platform-api \
  --task-definition kyb-platform-api:1 \
  --desired-count 3 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-12345],securityGroups=[sg-12345],assignPublicIp=ENABLED}"
```

### 4. Railway Deployment

#### Railway Configuration

```yaml
# railway.toml
[build]
builder = "nixpacks"

[deploy]
startCommand = "./kyb-platform-api"
healthcheckPath = "/health"
healthcheckTimeout = 300
restartPolicyType = "on_failure"

[[services]]
name = "kyb-platform-api"
```

#### Deploy to Railway

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login to Railway
railway login

# Link project
railway link

# Deploy
railway up
```

### 5. Supabase Deployment (Recommended for MVP)

#### Supabase Project Setup

```bash
# Create Supabase project
# 1. Go to https://supabase.com
# 2. Create new project
# 3. Note project URL and API keys

# Install Supabase CLI
npm install -g supabase

# Initialize Supabase project
supabase init

# Link to your Supabase project
supabase link --project-ref your-project-ref
```

#### Supabase Configuration

```bash
# Environment variables for Supabase
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_API_KEY=your_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
SUPABASE_JWT_SECRET=your_jwt_secret

# Database configuration (Supabase PostgreSQL)
DB_HOST=db.your-project.supabase.co
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=your_db_password
DB_SSL_MODE=require

# Provider configuration
PROVIDER_DATABASE=supabase
PROVIDER_AUTH=supabase
PROVIDER_CACHE=supabase
PROVIDER_STORAGE=supabase
```

#### Docker Compose with Supabase

```yaml
# docker-compose.supabase.yml
version: "3.8"

services:
  kyb-platform-supabase:
    build:
      context: .
      dockerfile: Dockerfile
      target: production
    ports:
      - "8081:8080"
    environment:
      - ENV=development
      - PROVIDER_DATABASE=supabase
      - PROVIDER_AUTH=supabase
      - PROVIDER_CACHE=supabase
      - PROVIDER_STORAGE=supabase
      - SUPABASE_URL=${SUPABASE_URL}
      - SUPABASE_API_KEY=${SUPABASE_API_KEY}
      - SUPABASE_SERVICE_ROLE_KEY=${SUPABASE_SERVICE_ROLE_KEY}
      - SUPABASE_JWT_SECRET=${SUPABASE_JWT_SECRET}
      - DB_HOST=${DB_HOST}
      - DB_PORT=${DB_PORT}
      - DB_USERNAME=${DB_USERNAME}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_DATABASE=${DB_DATABASE}
      - DB_SSL_MODE=${DB_SSL_MODE}
      - JWT_SECRET=${JWT_SECRET}
      - LOG_LEVEL=debug
      - METRICS_ENABLED=true
      - TRACING_ENABLED=true
    volumes:
      - ./configs:/app/configs
      - ./Codes:/app/Codes
      - ./.env:/app/.env
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Monitoring stack for Supabase environment
  prometheus-supabase:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./deployments/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./deployments/prometheus/alerts.yml:/etc/prometheus/alerts.yml
      - prometheus_supabase_data:/prometheus
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"
      - "--storage.tsdb.path=/prometheus"
      - "--web.console.libraries=/etc/prometheus/console_libraries"
      - "--web.console.templates=/etc/prometheus/consoles"
      - "--storage.tsdb.retention.time=200h"
      - "--web.enable-lifecycle"
    restart: unless-stopped

  grafana-supabase:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_supabase_data:/var/lib/grafana
      - ./deployments/grafana/dashboards:/etc/grafana/provisioning/dashboards
    restart: unless-stopped
    depends_on:
      - prometheus-supabase

volumes:
  prometheus_supabase_data:
  grafana_supabase_data:
```

#### Deploy with Supabase

```bash
# Start Supabase services
docker-compose -f docker-compose.supabase.yml up -d

# Run database migrations
supabase db push

# Verify deployment
curl http://localhost:8081/health

# View logs
docker-compose -f docker-compose.supabase.yml logs -f kyb-platform-supabase
```

#### Supabase Database Schema

```sql
-- Users table (extends Supabase auth.users)
CREATE TABLE public.profiles (
    id UUID REFERENCES auth.users(id) PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    full_name TEXT,
    role TEXT CHECK (role IN ('compliance_officer', 'risk_manager', 'business_analyst', 'developer', 'other')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Business classifications table
CREATE TABLE public.business_classifications (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.profiles(id),
    business_name TEXT NOT NULL,
    website_url TEXT,
    description TEXT,
    primary_industry JSONB,
    secondary_industries JSONB,
    confidence_score DECIMAL(3,2),
    classification_metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Risk assessments table
CREATE TABLE public.risk_assessments (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES public.profiles(id),
    business_id UUID REFERENCES public.business_classifications(id),
    risk_factors JSONB,
    risk_score DECIMAL(3,2),
    risk_level TEXT CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    assessment_metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Row Level Security (RLS) policies
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.business_classifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.risk_assessments ENABLE ROW LEVEL SECURITY;

-- RLS policies for profiles
CREATE POLICY "Users can view own profile" ON public.profiles
    FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can update own profile" ON public.profiles
    FOR UPDATE USING (auth.uid() = id);

-- RLS policies for business classifications
CREATE POLICY "Users can view own classifications" ON public.business_classifications
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own classifications" ON public.business_classifications
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own classifications" ON public.business_classifications
    FOR UPDATE USING (auth.uid() = user_id);

-- RLS policies for risk assessments
CREATE POLICY "Users can view own assessments" ON public.risk_assessments
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own assessments" ON public.risk_assessments
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own assessments" ON public.risk_assessments
    FOR UPDATE USING (auth.uid() = user_id);
```

## Configuration Management

### Environment-Specific Configurations

#### Development Configuration

```bash
# configs/development.env
ENVIRONMENT=development
LOG_LEVEL=debug
DB_HOST=localhost
DB_PORT=5432
DB_NAME=kyb_platform_dev
REDIS_HOST=localhost
REDIS_PORT=6379
API_PORT=8080
ENABLE_METRICS=true
ENABLE_TRACING=true
```

#### Production Configuration

```bash
# configs/production.env
ENVIRONMENT=production
LOG_LEVEL=info
DB_HOST=kyb-platform-db.production.com
DB_PORT=5432
DB_NAME=kyb_platform
REDIS_HOST=kyb-platform-redis.production.com
REDIS_PORT=6379
API_PORT=8080
ENABLE_METRICS=true
ENABLE_TRACING=true
JWT_SECRET=your-jwt-secret
API_KEY_SECRET=your-api-key-secret
```

#### Supabase Configuration

```bash
# configs/supabase.env
ENVIRONMENT=production
LOG_LEVEL=info
API_PORT=8080

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

# Monitoring
ENABLE_METRICS=true
ENABLE_TRACING=true
```

### Configuration Validation

```bash
# Validate configuration
./scripts/validate-config.sh

# Test configuration
./scripts/test-config.sh
```

## Monitoring and Health Checks

### Health Check Endpoints

```go
// Health check endpoint
GET /health
Response: {
  "status": "healthy",
  "timestamp": "2024-12-19T10:30:00Z",
  "version": "1.0.0",
  "uptime": "2h30m15s"
}

// Detailed health check
GET /health/detailed
Response: {
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

### Prometheus Metrics

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'kyb-platform'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

### Grafana Dashboards

```json
{
  "dashboard": {
    "title": "KYB Platform Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

## Security and Compliance

### SSL/TLS Configuration

```nginx
# nginx.conf
server {
    listen 443 ssl http2;
    server_name api.kyb-platform.com;
    
    ssl_certificate /etc/ssl/certs/kyb-platform.crt;
    ssl_certificate_key /etc/ssl/private/kyb-platform.key;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Security Headers

```go
// Security middleware
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        next.ServeHTTP(w, r)
    })
}
```

### Access Control

```yaml
# Kubernetes Network Policies
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kyb-platform-network-policy
  namespace: kyb-platform
spec:
  podSelector:
    matchLabels:
      app: kyb-platform-api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: database
    ports:
    - protocol: TCP
      port: 5432
```

## Scaling and Performance

### Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: kyb-platform-hpa
  namespace: kyb-platform
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: kyb-platform-api
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Load Balancer Configuration

```yaml
apiVersion: v1
kind: Service
metadata:
  name: kyb-platform-service
  namespace: kyb-platform
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
    service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled: "true"
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
  selector:
    app: kyb-platform-api
```

### Performance Tuning

```go
// Performance configuration
type PerformanceConfig struct {
    MaxConcurrentRequests int           `json:"max_concurrent_requests"`
    RequestTimeout        time.Duration `json:"request_timeout"`
    CacheTTL             time.Duration `json:"cache_ttl"`
    DatabasePoolSize     int           `json:"database_pool_size"`
    RedisPoolSize        int           `json:"redis_pool_size"`
}

// Default performance settings
var DefaultPerformanceConfig = PerformanceConfig{
    MaxConcurrentRequests: 1000,
    RequestTimeout:        30 * time.Second,
    CacheTTL:             5 * time.Minute,
    DatabasePoolSize:     20,
    RedisPoolSize:        10,
}
```

## Backup and Disaster Recovery

### Database Backup

```bash
#!/bin/bash
# scripts/backup-database.sh

BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="kyb_platform"

# Create backup
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME > $BACKUP_DIR/backup_$DATE.sql

# Compress backup
gzip $BACKUP_DIR/backup_$DATE.sql

# Upload to S3
aws s3 cp $BACKUP_DIR/backup_$DATE.sql.gz s3://kyb-platform-backups/

# Clean old backups (keep last 30 days)
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete
```

### Disaster Recovery Plan

```yaml
# disaster-recovery-plan.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: disaster-recovery-plan
  namespace: kyb-platform
data:
  recovery-steps: |
    1. Assess the scope of the disaster
    2. Activate backup systems
    3. Restore database from latest backup
    4. Verify data integrity
    5. Restart application services
    6. Run health checks
    7. Monitor system performance
    8. Document incident and lessons learned
```

## Troubleshooting

### Common Issues

#### High Memory Usage

```bash
# Check memory usage
kubectl top pods -n kyb-platform

# Analyze memory usage
kubectl exec -it deployment/kyb-platform-api -n kyb-platform -- go tool pprof http://localhost:8080/debug/pprof/heap
```

#### Database Connection Issues

```bash
# Test database connectivity
kubectl exec -it deployment/kyb-platform-api -n kyb-platform -- nc -zv $DB_HOST $DB_PORT

# Check database logs
kubectl logs -f deployment/postgres -n kyb-platform
```

#### Redis Connection Issues

```bash
# Test Redis connectivity
kubectl exec -it deployment/kyb-platform-api -n kyb-platform -- redis-cli -h $REDIS_HOST ping

# Check Redis logs
kubectl logs -f deployment/redis -n kyb-platform
```

### Log Analysis

```bash
# View application logs
kubectl logs -f deployment/kyb-platform-api -n kyb-platform

# Search for errors
kubectl logs deployment/kyb-platform-api -n kyb-platform | grep ERROR

# Analyze log patterns
kubectl logs deployment/kyb-platform-api -n kyb-platform | jq '.level' | sort | uniq -c
```

### Performance Diagnostics

```bash
# Run performance tests
./scripts/performance-testing.sh

# Generate performance report
./scripts/generate-performance-report.sh

# Analyze bottlenecks
go tool pprof http://localhost:8080/debug/pprof/profile
```

## Maintenance and Updates

### Rolling Updates

```bash
# Update application
kubectl set image deployment/kyb-platform-api kyb-platform-api=kyb-platform:v1.1.0 -n kyb-platform

# Monitor update progress
kubectl rollout status deployment/kyb-platform-api -n kyb-platform

# Rollback if needed
kubectl rollout undo deployment/kyb-platform-api -n kyb-platform
```

### Database Migrations

```bash
# Run database migrations
./scripts/run_migrations.sh

# Verify migration status
./scripts/verify-migrations.sh

# Rollback migrations if needed
./scripts/rollback-migrations.sh
```

### Monitoring Updates

```bash
# Update monitoring configuration
kubectl apply -f monitoring/prometheus.yml

# Reload Prometheus configuration
kubectl exec -it deployment/prometheus -n monitoring -- wget --post-data='' http://localhost:9090/-/reload

# Update Grafana dashboards
kubectl apply -f monitoring/grafana-dashboards/
```

### Security Updates

```bash
# Update security policies
kubectl apply -f security/network-policies.yaml

# Rotate secrets
./scripts/rotate-secrets.sh

# Update SSL certificates
./scripts/update-ssl-certificates.sh
```

---

**Document Version**: 1.0.0  
**Last Updated**: December 19, 2024  
**Next Review**: March 19, 2025
