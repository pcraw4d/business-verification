# KYB Platform - Deployment Documentation

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Development Environment](#development-environment)
4. [Staging Environment](#staging-environment)
5. [Production Environment](#production-environment)
6. [Docker Deployment](#docker-deployment)
7. [Kubernetes Deployment](#kubernetes-deployment)
8. [Cloud Deployment](#cloud-deployment)
9. [Database Deployment](#database-deployment)
10. [Monitoring Setup](#monitoring-setup)
11. [Security Configuration](#security-configuration)
12. [Backup & Recovery](#backup--recovery)
13. [Troubleshooting](#troubleshooting)
14. [Rollback Procedures](#rollback-procedures)

## Overview

This document provides comprehensive deployment procedures for the KYB Platform across different environments. The platform supports multiple deployment strategies including Docker, Kubernetes, and cloud-native deployments.

### Deployment Environments

- **Development**: Local development with hot reload
- **Staging**: Production-like environment for testing
- **Production**: High-availability production deployment

### Deployment Options

- **Docker Compose**: Simple containerized deployment
- **Kubernetes**: Orchestrated container deployment
- **Cloud Platforms**: AWS, Azure, GCP deployment
- **Bare Metal**: Direct server deployment

## Prerequisites

### System Requirements

**Minimum Requirements**
- **CPU**: 2 cores
- **Memory**: 4GB RAM
- **Storage**: 20GB available space
- **Network**: Internet connectivity for external APIs

**Recommended Requirements**
- **CPU**: 4+ cores
- **Memory**: 8GB+ RAM
- **Storage**: 100GB+ SSD
- **Network**: High-speed internet connection

### Software Dependencies

**Required Software**
- **Go**: 1.22 or higher
- **Docker**: 20.10 or higher
- **Docker Compose**: 2.0 or higher
- **PostgreSQL**: 14 or higher
- **Redis**: 6.0 or higher

**Optional Software**
- **Kubernetes**: 1.24 or higher
- **Helm**: 3.8 or higher
- **kubectl**: Latest version
- **Make**: 4.0 or higher

### Environment Variables

**Core Configuration**
```bash
# Application
KYB_ENV=development|staging|production
KYB_PORT=8080
KYB_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=kyb_platform
DB_USER=kyb_user
DB_PASSWORD=kyb_password
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Security
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRY=24h
BCRYPT_COST=12

# External APIs
EXTERNAL_API_TIMEOUT=30s
EXTERNAL_API_RETRIES=3
```

## Development Environment

### Local Development Setup

**Step 1: Clone Repository**
```bash
git clone https://github.com/your-org/kyb-platform.git
cd kyb-platform
```

**Step 2: Install Dependencies**
```bash
# Install Go dependencies
go mod download

# Install development tools
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Step 3: Setup Environment**
```bash
# Copy environment template
cp env.example .env

# Edit environment variables
nano .env
```

**Step 4: Start Dependencies**
```bash
# Start PostgreSQL and Redis with Docker
docker-compose up -d postgres redis

# Or start manually
docker run -d --name postgres \
  -e POSTGRES_DB=kyb_platform \
  -e POSTGRES_USER=kyb_user \
  -e POSTGRES_PASSWORD=kyb_password \
  -p 5432:5432 \
  postgres:14

docker run -d --name redis \
  -p 6379:6379 \
  redis:6-alpine
```

**Step 5: Run Database Migrations**
```bash
# Run migrations
make migrate

# Or manually
go run cmd/migrate/main.go
```

**Step 6: Start Application**
```bash
# Development mode with hot reload
make dev

# Or manually
air

# Or direct run
go run cmd/api/main.go
```

### Development Tools

**Hot Reload Configuration**
```toml
# .air.toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/api"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
```

**Code Quality Tools**
```bash
# Run linter
make lint

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Check for security issues
make security-check
```

## Staging Environment

### Staging Deployment

**Step 1: Prepare Staging Environment**
```bash
# Create staging directory
mkdir -p staging
cd staging

# Copy configuration
cp ../configs/staging.env .env
```

**Step 2: Build Application**
```bash
# Build for staging
make build-staging

# Or manually
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kyb-platform ./cmd/api
```

**Step 3: Deploy with Docker Compose**
```yaml
# docker-compose.staging.yml
version: '3.8'
services:
  api:
    build:
      context: ..
      dockerfile: Dockerfile
      target: staging
    ports:
      - "8080:8080"
    environment:
      - KYB_ENV=staging
    env_file:
      - .env
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ../internal/database/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    restart: unless-stopped

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
    restart: unless-stopped

volumes:
  postgres_data:
```

**Step 4: Deploy**
```bash
# Deploy staging environment
docker-compose -f docker-compose.staging.yml up -d

# Check status
docker-compose -f docker-compose.staging.yml ps

# View logs
docker-compose -f docker-compose.staging.yml logs -f api
```

### Staging Configuration

**Environment Variables**
```bash
# configs/staging.env
KYB_ENV=staging
KYB_PORT=8080
KYB_HOST=0.0.0.0

# Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=kyb_platform_staging
DB_USER=kyb_user
DB_PASSWORD=staging_password
DB_SSL_MODE=disable

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Security
JWT_SECRET=staging-jwt-secret-key
JWT_EXPIRY=24h
BCRYPT_COST=12

# Monitoring
PROMETHEUS_ENABLED=true
GRAFANA_ENABLED=true
LOG_LEVEL=info
```

## Production Environment

### Production Deployment Strategy

**Blue-Green Deployment**
```bash
# Deploy new version (Green)
kubectl apply -f k8s/production-green/

# Switch traffic
kubectl patch service kyb-platform -p '{"spec":{"selector":{"version":"green"}}}'

# Verify deployment
kubectl get pods -l app=kyb-platform,version=green

# Rollback if needed
kubectl patch service kyb-platform -p '{"spec":{"selector":{"version":"blue"}}}'
```

**Canary Deployment**
```bash
# Deploy canary version
kubectl apply -f k8s/production-canary/

# Gradually increase traffic
kubectl patch service kyb-platform -p '{"spec":{"selector":{"version":"canary"}}}'

# Monitor metrics
kubectl port-forward svc/prometheus 9090:9090

# Promote to production
kubectl apply -f k8s/production/
```

### Production Configuration

**Kubernetes Configuration**
```yaml
# k8s/production/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kyb-platform
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: kyb-platform
      version: v1.0.0
  template:
    metadata:
      labels:
        app: kyb-platform
        version: v1.0.0
    spec:
      containers:
      - name: kyb-platform
        image: kyb-platform:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: KYB_ENV
          value: "production"
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: kyb-secrets
              key: db-host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: kyb-secrets
              key: db-password
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
```

**Service Configuration**
```yaml
# k8s/production/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: kyb-platform
  namespace: production
spec:
  selector:
    app: kyb-platform
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

**Ingress Configuration**
```yaml
# k8s/production/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kyb-platform
  namespace: production
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
  - hosts:
    - api.kybplatform.com
    secretName: kyb-platform-tls
  rules:
  - host: api.kybplatform.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: kyb-platform
            port:
              number: 80
```

## Docker Deployment

### Dockerfile

**Multi-stage Build**
```dockerfile
# Dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 kyb && \
    adduser -D -s /bin/sh -u 1000 -G kyb kyb

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Change ownership
RUN chown -R kyb:kyb /app

# Switch to non-root user
USER kyb

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Run application
CMD ["./main"]
```

### Docker Compose

**Development Setup**
```yaml
# docker-compose.yml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - KYB_ENV=development
    env_file:
      - .env
    depends_on:
      - postgres
      - redis
    volumes:
      - .:/app
    command: air

  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: kyb_platform
      POSTGRES_USER: kyb_user
      POSTGRES_PASSWORD: kyb_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./internal/database/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./deployments/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana

volumes:
  postgres_data:
  prometheus_data:
  grafana_data:
```

## Kubernetes Deployment

### Helm Chart

**Chart Structure**
```
kyb-platform/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── hpa.yaml
│   └── pdb.yaml
└── charts/
```

**Chart.yaml**
```yaml
apiVersion: v2
name: kyb-platform
description: KYB Platform - Know Your Business Solution
type: application
version: 1.0.0
appVersion: "1.0.0"
keywords:
  - kyb
  - business
  - classification
  - risk
  - compliance
home: https://github.com/your-org/kyb-platform
sources:
  - https://github.com/your-org/kyb-platform
maintainers:
  - name: KYB Team
    email: team@kybplatform.com
```

**values.yaml**
```yaml
# Default values for kyb-platform
replicaCount: 3

image:
  repository: kyb-platform
  tag: "1.0.0"
  pullPolicy: IfNotPresent

imagePullSecrets: []

nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations: {}

podSecurityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000

securityContext:
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1000

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: nginx
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: api.kybplatform.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: kyb-platform-tls
      hosts:
        - api.kybplatform.com

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80

nodeSelector: {}

tolerations: []

affinity: {}

database:
  host: postgres
  port: 5432
  name: kyb_platform
  user: kyb_user
  password: ""

redis:
  host: redis
  port: 6379
  password: ""
  db: 0

monitoring:
  enabled: true
  prometheus:
    enabled: true
  grafana:
    enabled: true
```

### Deployment Commands

**Install Chart**
```bash
# Add repository
helm repo add kyb-platform https://charts.kybplatform.com
helm repo update

# Install chart
helm install kyb-platform kyb-platform/kyb-platform \
  --namespace production \
  --create-namespace \
  --values values-production.yaml

# Upgrade chart
helm upgrade kyb-platform kyb-platform/kyb-platform \
  --namespace production \
  --values values-production.yaml

# Rollback
helm rollback kyb-platform 1 --namespace production
```

## Cloud Deployment

### AWS Deployment

**ECS Fargate**
```yaml
# aws/ecs-task-definition.json
{
  "family": "kyb-platform",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "executionRoleArn": "arn:aws:iam::123456789012:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::123456789012:role/kyb-platform-task-role",
  "containerDefinitions": [
    {
      "name": "kyb-platform",
      "image": "123456789012.dkr.ecr.us-east-1.amazonaws.com/kyb-platform:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "KYB_ENV",
          "value": "production"
        }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:123456789012:secret:kyb/db-password"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/kyb-platform",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

**Application Load Balancer**
```yaml
# aws/alb.yaml
Resources:
  ApplicationLoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    Properties:
      Name: kyb-platform-alb
      Scheme: internet-facing
      Type: application
      SecurityGroups:
        - !Ref ALBSecurityGroup
      Subnets:
        - !Ref PublicSubnet1
        - !Ref PublicSubnet2

  ALBTargetGroup:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      Name: kyb-platform-tg
      Port: 8080
      Protocol: HTTP
      VpcId: !Ref VPC
      TargetType: ip
      HealthCheckPath: /health
      HealthCheckIntervalSeconds: 30
      HealthCheckTimeoutSeconds: 5
      HealthyThresholdCount: 2
      UnhealthyThresholdCount: 3

  ALBListener:
    Type: AWS::ElasticLoadBalancingV2::Listener
    Properties:
      LoadBalancerArn: !Ref ApplicationLoadBalancer
      Port: 443
      Protocol: HTTPS
      Certificates:
        - CertificateArn: !Ref SSLCertificate
      DefaultActions:
        - Type: forward
          TargetGroupArn: !Ref ALBTargetGroup
```

### Azure Deployment

**Azure Container Instances**
```yaml
# azure/aci-deployment.yaml
apiVersion: 2019-12-01
location: eastus
properties:
  containers:
  - name: kyb-platform
    properties:
      image: kybplatform.azurecr.io/kyb-platform:latest
      ports:
      - port: 8080
        protocol: TCP
      environmentVariables:
      - name: KYB_ENV
        value: production
      - name: DB_HOST
        value: kyb-postgres.postgres.database.azure.com
      resources:
        requests:
          memoryInGB: 2
          cpu: 1
        limits:
          memoryInGB: 4
          cpu: 2
      volumeMounts:
      - name: logs
        mountPath: /app/logs
  osType: Linux
  restartPolicy: Always
  volumes:
  - name: logs
    azureFile:
      shareName: kyb-logs
      storageAccountName: kybstorage
      storageAccountKey: <storage-account-key>
  ipAddress:
    type: Public
    ports:
    - protocol: TCP
      port: 8080
  tags:
    environment: production
    application: kyb-platform
```

## Database Deployment

### PostgreSQL Setup

**Production Database**
```sql
-- Create database
CREATE DATABASE kyb_platform_production;

-- Create user
CREATE USER kyb_user WITH PASSWORD 'secure_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE kyb_platform_production TO kyb_user;

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";
```

**Database Configuration**
```ini
# postgresql.conf
# Memory settings
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

# Connection settings
max_connections = 200
superuser_reserved_connections = 3

# Logging
log_destination = 'stderr'
logging_collector = on
log_directory = 'log'
log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
log_rotation_age = 1d
log_rotation_size = 100MB
log_min_duration_statement = 1000
log_checkpoints = on
log_connections = on
log_disconnections = on
log_lock_waits = on
log_temp_files = 0

# Performance
random_page_cost = 1.1
effective_io_concurrency = 200
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
```

### Redis Setup

**Redis Configuration**
```conf
# redis.conf
# Network
bind 0.0.0.0
port 6379
timeout 0
tcp-keepalive 300

# General
daemonize no
supervised no
pidfile /var/run/redis_6379.pid
loglevel notice
logfile ""
databases 16

# Snapshotting
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir ./

# Replication
replica-serve-stale-data yes
replica-read-only yes

# Security
requirepass "your_redis_password"

# Memory management
maxmemory 2gb
maxmemory-policy allkeys-lru

# Append only file
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
```

## Monitoring Setup

### Prometheus Configuration

**Prometheus Config**
```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "kyb-platform-rules.yml"

scrape_configs:
  - job_name: 'kyb-platform'
    static_configs:
      - targets: ['kyb-platform:8080']
    metrics_path: /metrics
    scrape_interval: 10s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

**Alert Rules**
```yaml
# kyb-platform-rules.yml
groups:
  - name: kyb-platform
    rules:
      - alert: HighErrorRate
        expr: rate(kyb_http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors per second"

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(kyb_http_request_duration_seconds_bucket[5m])) > 0.5
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }} seconds"

      - alert: LowCacheHitRatio
        expr: rate(kyb_cache_hits_total[5m]) / (rate(kyb_cache_hits_total[5m]) + rate(kyb_cache_misses_total[5m])) < 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low cache hit ratio"
          description: "Cache hit ratio is {{ $value | humanizePercentage }}"
```

### Grafana Dashboards

**Main Dashboard**
```json
{
  "dashboard": {
    "title": "KYB Platform - Overview",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(kyb_http_requests_total[5m])",
            "legendFormat": "{{method}} {{path}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(kyb_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(kyb_http_requests_total{status=~\"5..\"}[5m])",
            "legendFormat": "5xx errors"
          }
        ]
      }
    ]
  }
}
```

## Security Configuration

### SSL/TLS Setup

**Certificate Management**
```bash
# Generate self-signed certificate (development)
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Let's Encrypt certificate (production)
certbot certonly --standalone -d api.kybplatform.com

# Install certificate
sudo cp /etc/letsencrypt/live/api.kybplatform.com/fullchain.pem /etc/ssl/certs/
sudo cp /etc/letsencrypt/live/api.kybplatform.com/privkey.pem /etc/ssl/private/
```

**Nginx SSL Configuration**
```nginx
# nginx.conf
server {
    listen 443 ssl http2;
    server_name api.kybplatform.com;

    ssl_certificate /etc/ssl/certs/fullchain.pem;
    ssl_certificate_key /etc/ssl/private/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Network Security

**Firewall Configuration**
```bash
# UFW configuration
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8080/tcp
sudo ufw enable
```

**Security Groups (AWS)**
```yaml
# security-groups.yaml
Resources:
  ALBSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Security group for ALB
      VpcId: !Ref VPC
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 80
          ToPort: 80
          CidrIp: 0.0.0.0/0
        - IpProtocol: tcp
          FromPort: 443
          ToPort: 443
          CidrIp: 0.0.0.0/0

  AppSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Security group for application
      VpcId: !Ref VPC
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 8080
          ToPort: 8080
          SourceSecurityGroupId: !Ref ALBSecurityGroup
```

## Backup & Recovery

### Database Backup

**Automated Backup Script**
```bash
#!/bin/bash
# backup.sh

# Configuration
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="kyb_platform"
DB_USER="kyb_user"
BACKUP_DIR="/backups"
RETENTION_DAYS=30

# Create backup directory
mkdir -p $BACKUP_DIR

# Generate backup filename
BACKUP_FILE="$BACKUP_DIR/kyb_platform_$(date +%Y%m%d_%H%M%S).sql"

# Create backup
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME > $BACKUP_FILE

# Compress backup
gzip $BACKUP_FILE

# Remove old backups
find $BACKUP_DIR -name "*.sql.gz" -mtime +$RETENTION_DAYS -delete

# Upload to S3 (if using AWS)
aws s3 cp $BACKUP_FILE.gz s3://kyb-backups/database/

echo "Backup completed: $BACKUP_FILE.gz"
```

**Cron Job**
```bash
# Add to crontab
0 2 * * * /opt/kyb-platform/scripts/backup.sh >> /var/log/kyb-backup.log 2>&1
```

### Recovery Procedures

**Database Recovery**
```bash
#!/bin/bash
# restore.sh

# Configuration
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="kyb_platform"
DB_USER="kyb_user"
BACKUP_FILE="$1"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

# Stop application
systemctl stop kyb-platform

# Drop and recreate database
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "DROP DATABASE IF EXISTS $DB_NAME;"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $DB_NAME;"

# Restore from backup
gunzip -c $BACKUP_FILE | psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME

# Start application
systemctl start kyb-platform

echo "Recovery completed from: $BACKUP_FILE"
```

## Troubleshooting

### Common Issues

**Application Won't Start**
```bash
# Check logs
journalctl -u kyb-platform -f

# Check configuration
./kyb-platform --config-check

# Check dependencies
systemctl status postgresql
systemctl status redis

# Check ports
netstat -tlnp | grep :8080
```

**Database Connection Issues**
```bash
# Test database connection
psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "SELECT 1;"

# Check database status
systemctl status postgresql

# Check database logs
tail -f /var/log/postgresql/postgresql-*.log

# Check connection pool
psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "SELECT count(*) FROM pg_stat_activity;"
```

**High Memory Usage**
```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head -10

# Check Go garbage collection
curl -s http://localhost:8080/debug/vars | jq '.memstats'

# Check for memory leaks
go tool pprof http://localhost:8080/debug/pprof/heap
```

**Slow Response Times**
```bash
# Check database performance
psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
"

# Check slow queries
tail -f /var/log/postgresql/postgresql-*.log | grep "duration:"

# Check application metrics
curl -s http://localhost:8080/metrics | grep kyb_http_request_duration
```

### Performance Tuning

**Database Optimization**
```sql
-- Analyze table statistics
ANALYZE;

-- Update table statistics
VACUUM ANALYZE;

-- Check index usage
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Create missing indexes
CREATE INDEX CONCURRENTLY idx_businesses_name ON businesses USING gin(to_tsvector('english', name));
CREATE INDEX CONCURRENTLY idx_classifications_business_id ON classifications(business_id);
CREATE INDEX CONCURRENTLY idx_risk_assessments_business_id ON risk_assessments(business_id);
```

**Application Optimization**
```go
// Enable profiling
import _ "net/http/pprof"

// Add to main.go
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Memory profiling
import "runtime/pprof"

func profileMemory() {
    f, err := os.Create("memory.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    pprof.WriteHeapProfile(f)
}
```

## Rollback Procedures

### Application Rollback

**Docker Rollback**
```bash
# List available images
docker images kyb-platform

# Rollback to previous version
docker-compose down
docker-compose up -d --force-recreate

# Or specific version
docker-compose down
docker tag kyb-platform:previous kyb-platform:latest
docker-compose up -d
```

**Kubernetes Rollback**
```bash
# List deployments
kubectl get deployments -n production

# Rollback to previous version
kubectl rollout undo deployment/kyb-platform -n production

# Rollback to specific version
kubectl rollout undo deployment/kyb-platform -n production --to-revision=2

# Check rollback status
kubectl rollout status deployment/kyb-platform -n production
```

**Helm Rollback**
```bash
# List releases
helm list -n production

# Rollback to previous version
helm rollback kyb-platform 1 -n production

# Rollback to specific version
helm rollback kyb-platform 2 -n production
```

### Database Rollback

**Schema Rollback**
```bash
# List migrations
ls -la internal/database/migrations/

# Rollback last migration
go run cmd/migrate/main.go down 1

# Rollback to specific version
go run cmd/migrate/main.go down 5
```

**Data Rollback**
```bash
# Restore from backup
./scripts/restore.sh /backups/kyb_platform_20240115_020000.sql.gz

# Point-in-time recovery (if using WAL archiving)
pg_restore -h localhost -p 5432 -U kyb_user -d kyb_platform \
  --clean --if-exists \
  /backups/kyb_platform_20240115_020000.dump
```

---

## Conclusion

This deployment documentation provides comprehensive procedures for deploying the KYB Platform across different environments and platforms. Key points to remember:

- **Always test deployments in staging first**
- **Use blue-green or canary deployments for production**
- **Monitor deployments closely with proper observability**
- **Have rollback procedures ready**
- **Maintain security best practices**
- **Regular backup and recovery testing**

For additional support or questions about deployment procedures, please refer to the troubleshooting section or contact the development team.
