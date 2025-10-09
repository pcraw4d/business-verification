# Railway Deployment Guide

This document provides comprehensive guidance for deploying the Risk Assessment Service to Railway with proper configuration and monitoring.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Railway Setup](#railway-setup)
3. [Configuration](#configuration)
4. [Deployment Process](#deployment-process)
5. [Monitoring and Observability](#monitoring-and-observability)
6. [Environment Management](#environment-management)
7. [Troubleshooting](#troubleshooting)
8. [Best Practices](#best-practices)

## Prerequisites

### Required Tools

1. **Railway CLI**
   ```bash
   # Install Railway CLI
   npm install -g @railway/cli
   
   # Or using curl
   curl -fsSL https://railway.app/install.sh | sh
   ```

2. **Docker** (for local testing)
   ```bash
   # Install Docker Desktop
   # https://www.docker.com/products/docker-desktop
   ```

3. **Go 1.22+** (for local development)
   ```bash
   # Install Go
   # https://golang.org/doc/install
   ```

### Required Accounts

1. **Railway Account**: Sign up at [railway.app](https://railway.app)
2. **Supabase Account**: For database services
3. **External API Keys**: NewsAPI, OpenCorporates (optional)

## Railway Setup

### 1. Authentication

```bash
# Login to Railway
railway login

# Verify authentication
railway whoami
```

### 2. Project Creation

```bash
# Create new project
railway init kyb-platform

# Or link to existing project
railway link <project-id>
```

### 3. Service Configuration

The service is configured with the following files:

- `railway.json` - Railway deployment configuration
- `Dockerfile` - Container build instructions
- `railway.env` - Environment variables template
- `.railway/config.toml` - Railway-specific settings

## Configuration

### Environment Variables

#### Required Variables

```bash
# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key

# Service Configuration
SERVICE_NAME=risk-assessment-service
ENV=production
LOG_LEVEL=info
```

#### Performance Monitoring

```bash
# Performance Targets
PERFORMANCE_TARGET_RPS=16.67
PERFORMANCE_TARGET_LATENCY=1s
PERFORMANCE_TARGET_ERROR_RATE=0.01
PERFORMANCE_TARGET_THROUGHPUT=1000

# Monitoring
PERFORMANCE_MONITORING_ENABLED=true
METRICS_ENABLED=true
```

#### External APIs (Optional)

```bash
# NewsAPI
NEWS_API_KEY=your_news_api_key
NEWS_API_ENABLED=true

# OpenCorporates
OPEN_CORPORATES_API_KEY=your_opencorporates_key
OPEN_CORPORATES_ENABLED=true

# Government APIs
GOVERNMENT_API_KEY=your_government_api_key
GOVERNMENT_API_ENABLED=true
```

### Railway Configuration

#### `railway.json`

```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile"
  },
  "deploy": {
    "startCommand": "./risk-assessment-service",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 30,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 5
  }
}
```

#### `Dockerfile`

```dockerfile
# Multi-stage build for optimized production image
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o risk-assessment-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D appuser
WORKDIR /app
COPY --from=builder /app/risk-assessment-service .
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1
CMD ["./risk-assessment-service"]
```

## Deployment Process

### Automated Deployment

Use the provided deployment script:

```bash
# Full deployment process
./scripts/deploy_railway.sh

# Verify existing deployment
./scripts/deploy_railway.sh verify

# Check deployment status
./scripts/deploy_railway.sh status

# View logs
./scripts/deploy_railway.sh logs
```

### Manual Deployment

#### 1. Set Environment Variables

```bash
# Set required variables
railway variables set SUPABASE_URL=https://your-project.supabase.co
railway variables set SUPABASE_ANON_KEY=your_anon_key
railway variables set SUPABASE_SERVICE_ROLE_KEY=your_service_role_key

# Set performance targets
railway variables set PERFORMANCE_TARGET_RPS=16.67
railway variables set PERFORMANCE_TARGET_THROUGHPUT=1000

# Set optional external API keys
railway variables set NEWS_API_KEY=your_news_api_key
railway variables set OPEN_CORPORATES_API_KEY=your_opencorporates_key
```

#### 2. Deploy Service

```bash
# Deploy to Railway
railway up

# Deploy with specific environment
railway up --environment production
```

#### 3. Verify Deployment

```bash
# Check deployment status
railway status

# View logs
railway logs

# Get service URL
railway domain
```

### Deployment Verification

#### Health Checks

```bash
# Basic health check
curl https://your-service.railway.app/health

# Performance health check
curl https://your-service.railway.app/api/v1/performance/health

# Performance statistics
curl https://your-service.railway.app/api/v1/performance/stats
```

#### Load Testing

```bash
# Run load test against deployed service
go run ./cmd/load_test.go \
  -url=https://your-service.railway.app \
  -duration=5m \
  -users=20 \
  -rps=16.67 \
  -type=load
```

## Monitoring and Observability

### Built-in Monitoring

The service includes comprehensive monitoring capabilities:

#### Performance Metrics

- **Request Count**: Total requests processed
- **Response Times**: Average, min, max response times
- **Error Rates**: Success/failure ratios
- **Throughput**: Requests per second/minute
- **Resource Usage**: Memory, CPU, goroutine count

#### Monitoring Endpoints

```bash
# Performance statistics
GET /api/v1/performance/stats

# Performance alerts
GET /api/v1/performance/alerts

# Service health
GET /api/v1/performance/health

# General metrics
GET /api/v1/metrics

# Health check
GET /health
```

#### Alert System

The service automatically generates alerts for:

- **High Latency**: Response time > 1 second
- **High Error Rate**: Error rate > 5%
- **Low Throughput**: Throughput < 800 req/min
- **Resource Issues**: High memory, CPU, or goroutine usage

### Railway Monitoring

#### Railway Dashboard

1. **Service Overview**: CPU, memory, network usage
2. **Deployment History**: Build and deployment logs
3. **Environment Variables**: Configuration management
4. **Logs**: Real-time application logs

#### Custom Monitoring

```bash
# View Railway logs
railway logs --tail 100

# Monitor resource usage
railway status

# Check deployment health
railway health
```

## Environment Management

### Environment Types

#### Production Environment

```bash
# Production configuration
ENV=production
LOG_LEVEL=info
PERFORMANCE_TARGET_RPS=16.67
RATE_LIMIT_REQUESTS_PER=1000
```

#### Staging Environment

```bash
# Staging configuration
ENV=staging
LOG_LEVEL=debug
PERFORMANCE_TARGET_RPS=8.33
RATE_LIMIT_REQUESTS_PER=500
```

### Environment Variables Management

#### Setting Variables

```bash
# Set individual variables
railway variables set KEY=value

# Set multiple variables from file
railway variables set --file railway.env

# Set variables for specific environment
railway variables set KEY=value --environment production
```

#### Viewing Variables

```bash
# List all variables
railway variables

# View specific variable
railway variables get KEY

# Export variables
railway variables export > .env
```

## Troubleshooting

### Common Issues

#### 1. Build Failures

**Problem**: Service fails to build
**Solution**:
```bash
# Check build logs
railway logs --build

# Verify Go module
go mod verify

# Test local build
go build ./cmd/main.go
```

#### 2. Health Check Failures

**Problem**: Health checks failing
**Solution**:
```bash
# Check service logs
railway logs

# Verify health endpoint
curl https://your-service.railway.app/health

# Check environment variables
railway variables
```

#### 3. Performance Issues

**Problem**: Service not meeting performance targets
**Solution**:
```bash
# Check performance metrics
curl https://your-service.railway.app/api/v1/performance/stats

# Review performance alerts
curl https://your-service.railway.app/api/v1/performance/alerts

# Run load tests
go run ./cmd/load_test.go -url=https://your-service.railway.app
```

#### 4. Database Connection Issues

**Problem**: Supabase connection failures
**Solution**:
```bash
# Verify Supabase variables
railway variables get SUPABASE_URL
railway variables get SUPABASE_ANON_KEY

# Check database connectivity
curl https://your-service.railway.app/health
```

### Debugging Commands

```bash
# View recent logs
railway logs --tail 50

# View build logs
railway logs --build

# Check service status
railway status

# View environment variables
railway variables

# Access service shell (if available)
railway shell
```

## Best Practices

### Security

1. **Environment Variables**: Never commit sensitive data to version control
2. **API Keys**: Use Railway's secure variable storage
3. **HTTPS**: Always use HTTPS in production
4. **Rate Limiting**: Enable rate limiting to prevent abuse

### Performance

1. **Resource Limits**: Set appropriate CPU and memory limits
2. **Health Checks**: Configure proper health check intervals
3. **Monitoring**: Enable comprehensive performance monitoring
4. **Caching**: Use caching for frequently accessed data

### Deployment

1. **Staging First**: Always test in staging before production
2. **Rolling Deployments**: Use Railway's rolling deployment feature
3. **Backup Strategy**: Implement proper backup procedures
4. **Monitoring**: Set up alerts for critical metrics

### Development

1. **Local Testing**: Test locally before deploying
2. **Load Testing**: Run load tests after deployment
3. **Documentation**: Keep deployment documentation updated
4. **Version Control**: Use proper version control practices

## Monitoring Dashboard

### Key Metrics to Monitor

1. **Response Time**: Should be < 1 second
2. **Throughput**: Should be ≥ 1000 req/min
3. **Error Rate**: Should be < 1%
4. **Memory Usage**: Should be < 512MB
5. **CPU Usage**: Should be < 80%

### Alert Thresholds

- **Critical**: Error rate > 5%, Response time > 2s
- **Warning**: Error rate > 1%, Response time > 1s
- **Info**: Performance degradation trends

## Conclusion

This deployment guide provides comprehensive instructions for deploying the Risk Assessment Service to Railway with proper configuration, monitoring, and observability. The service is designed to handle 1000 requests per minute reliably while maintaining high performance standards.

For additional support or questions, please refer to the Railway documentation or contact the development team.
