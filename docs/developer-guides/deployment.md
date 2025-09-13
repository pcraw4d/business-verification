# KYB Platform Deployment Guide

## Overview

This guide covers deployment procedures for the KYB Platform across different environments. The platform supports containerized deployment with Docker and can be deployed to various cloud platforms.

## Prerequisites

### System Requirements

**Minimum Requirements**:
- CPU: 2 cores
- RAM: 4GB
- Storage: 20GB SSD
- OS: Linux (Ubuntu 20.04+), macOS, or Windows with WSL2

**Recommended Requirements**:
- CPU: 4+ cores
- RAM: 8GB+
- Storage: 50GB+ SSD
- OS: Linux (Ubuntu 22.04+)

### Software Dependencies

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.22+ (for development)
- Node.js 18+ (for frontend development)
- PostgreSQL 15+ (for production)
- Redis 7+ (for caching)

## Environment Configuration

### 1. Development Environment

**Purpose**: Local development and testing

**Configuration**:
```bash
# Environment variables
export ENV=development
export DEBUG=true
export LOG_LEVEL=debug
export DATABASE_URL=postgres://kyb_dev:password@localhost:5432/kyb_dev
export REDIS_URL=redis://localhost:6379/0
export JWT_SECRET=dev-secret-key
export API_PORT=8080
```

**Setup**:
```bash
# Clone repository
git clone <repository-url>
cd kyb-platform

# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# Run database migrations
go run cmd/migrate/main.go up

# Seed development data
go run cmd/seed/main.go

# Start development server
go run cmd/server/main.go
```

### 2. Staging Environment

**Purpose**: Pre-production testing and validation

**Configuration**:
```bash
# Environment variables
export ENV=staging
export DEBUG=false
export LOG_LEVEL=info
export DATABASE_URL=postgres://kyb_staging:password@staging-db:5432/kyb_staging
export REDIS_URL=redis://staging-redis:6379/0
export JWT_SECRET=staging-secret-key
export API_PORT=8080
```

**Setup**:
```bash
# Deploy to staging
docker-compose -f docker-compose.staging.yml up -d

# Run health checks
./scripts/health-check.sh staging

# Run integration tests
go test ./test/integration/... -env=staging
```

### 3. Production Environment

**Purpose**: Live production deployment

**Configuration**:
```bash
# Environment variables
export ENV=production
export DEBUG=false
export LOG_LEVEL=warn
export DATABASE_URL=postgres://kyb_prod:secure-password@prod-db:5432/kyb_prod
export REDIS_URL=redis://prod-redis:6379/0
export JWT_SECRET=production-secret-key
export API_PORT=8080
export SSL_CERT_PATH=/etc/ssl/certs/kyb.crt
export SSL_KEY_PATH=/etc/ssl/private/kyb.key
```

## Docker Deployment

### 1. Building Images

**Backend Image**:
```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/web ./web
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./main"]
```

**Build Commands**:
```bash
# Build backend image
docker build -t kyb-platform:latest .

# Build with specific tag
docker build -t kyb-platform:v1.0.0 .

# Build for specific environment
docker build -f Dockerfile.production -t kyb-platform:prod .
```

### 2. Docker Compose Configuration

**Development** (`docker-compose.dev.yml`):
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=development
      - DATABASE_URL=postgres://kyb_dev:password@db:5432/kyb_dev
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      - db
      - redis
    volumes:
      - .:/app
      - /app/node_modules

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: kyb_dev
      POSTGRES_USER: kyb_dev
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

**Production** (`docker-compose.production.yml`):
```yaml
version: '3.8'

services:
  app:
    image: kyb-platform:latest
    ports:
      - "8080:8080"
      - "8443:8443"
    environment:
      - ENV=production
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - db
      - redis
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres_data:
  redis_data:
```

### 3. Deployment Commands

**Development**:
```bash
# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f

# Stop development environment
docker-compose -f docker-compose.dev.yml down
```

**Production**:
```bash
# Deploy to production
docker-compose -f docker-compose.production.yml up -d

# Update production deployment
docker-compose -f docker-compose.production.yml pull
docker-compose -f docker-compose.production.yml up -d

# Rollback production deployment
docker-compose -f docker-compose.production.yml down
docker-compose -f docker-compose.production.yml up -d
```

## Cloud Deployment

### 1. Railway Deployment

**Configuration** (`railway.json`):
```json
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile"
  },
  "deploy": {
    "startCommand": "./main",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 100,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

**Deployment Steps**:
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login to Railway
railway login

# Link project
railway link

# Deploy
railway up

# Set environment variables
railway variables set DATABASE_URL=$DATABASE_URL
railway variables set REDIS_URL=$REDIS_URL
railway variables set JWT_SECRET=$JWT_SECRET
```

### 2. AWS Deployment

**ECS Task Definition**:
```json
{
  "family": "kyb-platform",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::account:role/ecsTaskExecutionRole",
  "containerDefinitions": [
    {
      "name": "kyb-platform",
      "image": "kyb-platform:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "ENV",
          "value": "production"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:kyb/database-url"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/kyb-platform",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

### 3. Google Cloud Deployment

**Cloud Run Configuration**:
```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: kyb-platform
  annotations:
    run.googleapis.com/ingress: all
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/maxScale: "100"
        run.googleapis.com/cpu-throttling: "false"
    spec:
      containerConcurrency: 80
      containers:
      - image: gcr.io/project-id/kyb-platform:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: kyb-secrets
              key: database-url
        resources:
          limits:
            cpu: "2"
            memory: "2Gi"
```

## Database Deployment

### 1. Migration Strategy

**Migration Commands**:
```bash
# Run all pending migrations
go run cmd/migrate/main.go up

# Run specific migration
go run cmd/migrate/main.go up 1

# Rollback last migration
go run cmd/migrate/main.go down 1

# Check migration status
go run cmd/migrate/main.go status
```

**Migration Files**:
```sql
-- 001_initial_schema.sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT NOW()
);

-- 002_merchants_table.sql
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    business_type VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### 2. Data Seeding

**Seed Commands**:
```bash
# Seed development data
go run cmd/seed/main.go

# Seed specific data
go run cmd/seed/main.go --type=merchants

# Clear and reseed
go run cmd/seed/main.go --clear
```

## Monitoring and Health Checks

### 1. Health Check Endpoints

**Application Health**:
```go
// Health check endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now().UTC(),
        "version":   version,
        "services": map[string]string{
            "database": checkDatabase(),
            "redis":    checkRedis(),
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}
```

**Health Check Script**:
```bash
#!/bin/bash
# scripts/health-check.sh

ENDPOINT=${1:-"http://localhost:8080"}
TIMEOUT=${2:-30}

echo "Checking health of KYB Platform at $ENDPOINT"

# Check if service is responding
if curl -f -s --max-time $TIMEOUT "$ENDPOINT/health" > /dev/null; then
    echo "✅ Service is healthy"
    exit 0
else
    echo "❌ Service is unhealthy"
    exit 1
fi
```

### 2. Monitoring Setup

**Prometheus Metrics**:
```go
// Metrics collection
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)
```

## Security Considerations

### 1. Environment Variables

**Secure Configuration**:
```bash
# Use secrets management
export DATABASE_URL=$(aws secretsmanager get-secret-value --secret-id kyb/database-url --query SecretString --output text)
export JWT_SECRET=$(aws secretsmanager get-secret-value --secret-id kyb/jwt-secret --query SecretString --output text)
```

### 2. SSL/TLS Configuration

**Nginx Configuration**:
```nginx
server {
    listen 443 ssl http2;
    server_name kyb-platform.com;
    
    ssl_certificate /etc/ssl/certs/kyb.crt;
    ssl_certificate_key /etc/ssl/private/kyb.key;
    
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

## Rollback Procedures

### 1. Application Rollback

**Docker Rollback**:
```bash
# Rollback to previous version
docker-compose -f docker-compose.production.yml down
docker tag kyb-platform:latest kyb-platform:rollback
docker tag kyb-platform:previous kyb-platform:latest
docker-compose -f docker-compose.production.yml up -d
```

**Database Rollback**:
```bash
# Rollback database migrations
go run cmd/migrate/main.go down 1

# Restore from backup
pg_restore -h localhost -U kyb_prod -d kyb_prod backup_$(date -d '1 day ago' +%Y%m%d).sql
```

### 2. Emergency Procedures

**Emergency Rollback Script**:
```bash
#!/bin/bash
# scripts/rollback/database-rollback.sh --force --target previous-stable-version full

echo "🚨 Emergency rollback initiated"

# Stop current deployment
docker-compose -f docker-compose.production.yml down

# Restore previous version
docker tag kyb-platform:rollback kyb-platform:latest

# Start with previous version
docker-compose -f docker-compose.production.yml up -d

# Verify health
./scripts/health-check.sh

echo "✅ Emergency rollback completed"
```

## Troubleshooting

### Common Issues

**1. Database Connection Issues**:
```bash
# Check database connectivity
pg_isready -h localhost -p 5432 -U kyb_prod

# Check database logs
docker logs kyb-platform-db-1
```

**2. Redis Connection Issues**:
```bash
# Check Redis connectivity
redis-cli -h localhost -p 6379 ping

# Check Redis logs
docker logs kyb-platform-redis-1
```

**3. Application Issues**:
```bash
# Check application logs
docker logs kyb-platform-app-1

# Check application health
curl -f http://localhost:8080/health
```

### Performance Issues

**Database Performance**:
```sql
-- Check slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- Check table sizes
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

**Application Performance**:
```bash
# Check resource usage
docker stats kyb-platform-app-1

# Check application metrics
curl http://localhost:8080/metrics
```

## Best Practices

### 1. Deployment Best Practices

- Always test deployments in staging first
- Use blue-green deployments for zero downtime
- Implement proper health checks
- Monitor deployment metrics
- Have rollback procedures ready

### 2. Security Best Practices

- Use secrets management for sensitive data
- Enable SSL/TLS encryption
- Implement proper authentication
- Regular security updates
- Monitor security events

### 3. Monitoring Best Practices

- Set up comprehensive monitoring
- Implement alerting for critical issues
- Regular performance reviews
- Log aggregation and analysis
- Capacity planning

## Conclusion

This deployment guide provides comprehensive instructions for deploying the KYB Platform across different environments. Follow the procedures carefully and always test in staging before production deployment.

For additional support, refer to the troubleshooting section or contact the development team.
