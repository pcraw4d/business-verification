# KYB Platform Deployment Guide

**Version**: 1.0  
**Last Updated**: 2025-01-27  
**Platform**: Railway

---

## Table of Contents

1. [Overview](#overview)
2. [Service URLs](#service-urls)
3. [Environment Variables](#environment-variables)
4. [Database Schema](#database-schema)
5. [Monitoring Setup](#monitoring-setup)
6. [Deployment Process](#deployment-process)
7. [Troubleshooting](#troubleshooting)

---

## Overview

The KYB Platform consists of multiple microservices deployed on Railway. This guide provides comprehensive deployment information including service URLs, environment variables, database schema, and monitoring configuration.

### Architecture

- **API Gateway**: Entry point for all API requests
- **Classification Service**: Business classification and industry code detection
- **Merchant Service**: Merchant management and CRUD operations
- **Risk Assessment Service**: Risk scoring and assessment
- **Frontend Service**: Web interface
- **Business Intelligence Service**: Analytics and reporting
- **Service Discovery**: Service registry and health monitoring

---

## Service URLs

### Production URLs

| Service | URL | Status | Health Check |
|---------|-----|--------|--------------|
| **API Gateway** | `https://api-gateway-service-production-21fd.up.railway.app` | ✅ Active | `/health` |
| **Classification Service** | `https://classification-service-production.up.railway.app` | ✅ Active | `/health` |
| **Merchant Service** | `https://merchant-service-production.up.railway.app` | ✅ Active | `/health` |
| **Risk Assessment Service** | `https://risk-assessment-service-production.up.railway.app` | ✅ Active | `/health` |
| **Frontend Service** | `https://frontend-service-production-b225.up.railway.app` | ✅ Active | `/` |
| **BI Service** | `https://bi-service-production.up.railway.app` | ✅ Active | `/health` |
| **Pipeline Service** | `https://pipeline-service-production.up.railway.app` | ✅ Active | `/health` |
| **Monitoring Service** | `https://monitoring-service-production.up.railway.app` | ✅ Active | `/health` |
| **Service Discovery** | `https://service-discovery-production-d397.up.railway.app` | ✅ Active | `/health` |

### Important Notes

⚠️ **DO NOT USE OLD URLs**:
- ❌ `kyb-api-gateway-production.up.railway.app` (OLD - DO NOT USE)
- ✅ `api-gateway-service-production-21fd.up.railway.app` (CORRECT)

---

## Environment Variables

### Shared Variables (Set at Project Level)

These variables should be set at the Railway project level so all services can access them:

```bash
# =============================================================================
# CRITICAL SUPABASE VARIABLES (REQUIRED FOR ALL SERVICES)
# =============================================================================
SUPABASE_URL=https://qpqhuqqmkjxsltzshfam.supabase.co
SUPABASE_ANON_KEY=your_supabase_anon_key_here
SUPABASE_SERVICE_ROLE_KEY=your_supabase_service_role_key_here
SUPABASE_JWT_SECRET=your_supabase_jwt_secret_here

# =============================================================================
# ENVIRONMENT CONFIGURATION
# =============================================================================
ENV=production
ENVIRONMENT=production
LOG_LEVEL=info
LOG_FORMAT=json

# =============================================================================
# CORS CONFIGURATION (FOR API GATEWAY)
# =============================================================================
CORS_ALLOWED_ORIGINS=https://frontend-service-production-b225.up.railway.app
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=*
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=86400

# =============================================================================
# RATE LIMITING
# =============================================================================
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER=1000
RATE_LIMIT_WINDOW_SIZE=3600
RATE_LIMIT_BURST_SIZE=2000
```

### Service-Specific Variables

#### API Gateway Service

```bash
PORT=8080
CLASSIFICATION_SERVICE_URL=https://classification-service-production.up.railway.app
MERCHANT_SERVICE_URL=https://merchant-service-production.up.railway.app
FRONTEND_URL=https://frontend-service-production-b225.up.railway.app
RISK_ASSESSMENT_URL=https://risk-assessment-service-production.up.railway.app
BI_SERVICE_URL=https://bi-service-production.up.railway.app
```

#### Classification Service

```bash
PORT=8081

# Caching Configuration
CACHE_ENABLED=true
CACHE_TTL=5m

# Multi-Strategy Classifier Configuration (v3.0+)
# The service now uses a multi-strategy classifier combining:
# - Keyword-based classification (40% weight)
# - Entity-based classification (25% weight)
# - Topic-based classification (20% weight)
# - Co-occurrence-based classification (15% weight)
MULTI_STRATEGY_ENABLED=true
CONFIDENCE_CALIBRATION_ENABLED=true

# Legacy Feature Flags (for backward compatibility)
ML_ENABLED=true
KEYWORD_METHOD_ENABLED=true
ENSEMBLE_ENABLED=true

# Performance Configuration
MAX_CONCURRENT_REQUESTS=100
REQUEST_TIMEOUT=10s
CLASSIFICATION_OVERALL_TIMEOUT=60s

# Multi-Page Analysis Configuration
ENABLE_MULTI_PAGE_ANALYSIS=true
CLASSIFICATION_MAX_PAGES_TO_ANALYZE=15
CLASSIFICATION_PAGE_ANALYSIS_TIMEOUT=15s
CLASSIFICATION_CONCURRENT_PAGES=5

# Structured Data Extraction
ENABLE_STRUCTURED_DATA_EXTRACTION=true

# Brand Matching
CLASSIFICATION_BRAND_MATCH_ENABLED=true
CLASSIFICATION_BRAND_MATCH_MCC_RANGE=3000-3831
```

#### Merchant Service

```bash
PORT=8082
MERCHANT_SEARCH_LIMIT=100
MERCHANT_REQUEST_TIMEOUT=30s
```

#### Risk Assessment Service

```bash
PORT=8083
# Prometheus Metrics
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
PROMETHEUS_PATH=/metrics

# Grafana
GRAFANA_ENABLED=true
GRAFANA_BASE_URL=https://your-grafana-instance.com
GRAFANA_AUTO_CREATE=true
GRAFANA_DASHBOARD_UID=risk-assessment

# Feature Flags
ENABLE_INCOMPLETE_RISK_BENCHMARKS=false  # Set to true to enable in production
```

#### Frontend Service

```bash
PORT=8086
API_GATEWAY_URL=https://api-gateway-service-production-21fd.up.railway.app
```

### Setting Environment Variables in Railway

#### Method 1: Railway Dashboard

1. Go to Railway dashboard
2. Click on your project
3. Go to "Variables" tab
4. Add shared variables at project level
5. Add service-specific variables to each service

#### Method 2: Railway CLI

```bash
# Set shared variables
railway variables set SUPABASE_URL=https://qpqhuqqmkjxsltzshfam.supabase.co
railway variables set SUPABASE_ANON_KEY=your_anon_key
railway variables set SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
railway variables set SUPABASE_JWT_SECRET=your_jwt_secret

# Set service-specific variables
railway variables set PORT=8080 --service api-gateway-service
railway variables set PORT=8081 --service classification-service
railway variables set PORT=8082 --service merchant-service
```

---

## Database Schema

### Supabase Configuration

- **URL**: `https://qpqhuqqmkjxsltzshfam.supabase.co`
- **Database**: PostgreSQL (managed by Supabase)
- **Connection**: Via Supabase PostgREST API

### Core Tables

#### Merchants Table

```sql
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    registration_number VARCHAR(100),
    tax_id VARCHAR(100),
    industry VARCHAR(100),
    industry_code VARCHAR(20),
    business_type VARCHAR(50),
    founded_date DATE,
    employee_count INTEGER,
    annual_revenue DECIMAL(15,2),
    address JSONB,
    contact_info JSONB,
    portfolio_type VARCHAR(50),
    risk_level VARCHAR(50),
    compliance_status VARCHAR(50) DEFAULT 'pending',
    status VARCHAR(50) DEFAULT 'active',
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Indexes
CREATE INDEX idx_merchants_created_at ON merchants(created_at);
CREATE INDEX idx_merchants_portfolio_type ON merchants(portfolio_type);
CREATE INDEX idx_merchants_risk_level ON merchants(risk_level);
CREATE INDEX idx_merchants_status ON merchants(status);
CREATE INDEX idx_merchants_name ON merchants(name);
CREATE INDEX idx_merchants_portfolio_risk ON merchants(portfolio_type, risk_level);
```

#### Classifications Table

```sql
CREATE TABLE business_classifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_name TEXT NOT NULL,
    website_url TEXT,
    description TEXT,
    primary_industry JSONB,
    secondary_industries JSONB,
    confidence_score DECIMAL(3,2) CHECK (confidence_score >= 0 AND confidence_score <= 1),
    classification_metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_classifications_business_name ON business_classifications(business_name);
CREATE INDEX idx_classifications_created_at ON business_classifications(created_at);
```

#### Risk Assessments Table

```sql
CREATE TABLE risk_assessments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id UUID,
    risk_factors JSONB,
    risk_score DECIMAL(3,2) CHECK (risk_score >= 0 AND risk_score <= 1),
    risk_level TEXT CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    assessment_metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_risk_assessments_business_id ON risk_assessments(business_id);
CREATE INDEX idx_risk_assessments_created_at ON risk_assessments(created_at);
CREATE INDEX idx_risk_assessments_risk_score ON risk_assessments(risk_score);
CREATE INDEX idx_risk_assessments_business_created ON risk_assessments(business_id, created_at);
```

### Database Connection

All services connect to Supabase via the PostgREST API using:
- **URL**: `SUPABASE_URL`
- **API Key**: `SUPABASE_ANON_KEY` (for read operations)
- **Service Role Key**: `SUPABASE_SERVICE_ROLE_KEY` (for write operations)

---

## Monitoring Setup

### Prometheus Metrics

#### Risk Assessment Service

- **Metrics Endpoint**: `http://risk-assessment-service:9090/metrics`
- **Enabled**: `PROMETHEUS_ENABLED=true`
- **Port**: `9090` (configurable via `PROMETHEUS_PORT`)

#### Other Services

- **Metrics Endpoint**: `/metrics` on main server port
- All services expose Prometheus metrics

### Grafana Dashboards

#### Risk Assessment Service

- **Auto-Creation**: Enabled via `GRAFANA_AUTO_CREATE=true`
- **Base URL**: Set via `GRAFANA_BASE_URL`
- **Dashboard UID**: `risk-assessment` (configurable)

#### Available Metrics

- **HTTP Metrics**:
  - `http_requests_total` - Total HTTP requests
  - `http_request_duration_seconds` - Request duration
  - `http_errors_total` - Error count

- **Database Metrics**:
  - `db_query_duration_seconds` - Query duration
  - `db_query_total` - Total queries
  - `db_slow_queries_total` - Slow query count

- **Service Metrics**:
  - `service_health` - Service health status
  - `service_uptime_seconds` - Service uptime

### Health Checks

All services expose health check endpoints:

- **API Gateway**: `GET /health`
- **Classification Service**: `GET /health`
- **Merchant Service**: `GET /health`
- **Risk Assessment Service**: `GET /health`

**Health Check Response**:
```json
{
  "status": "healthy",
  "service": "service-name",
  "version": "1.0.0",
  "timestamp": "2025-01-27T12:00:00Z"
}
```

### Logging

- **Format**: JSON (production), Console (development)
- **Level**: `info` (production), `debug` (development)
- **Location**: Railway logs dashboard

---

## Deployment Process

### Prerequisites

1. Railway account with project access
2. Supabase project with database configured
3. Environment variables configured
4. Service URLs verified

### Deployment Steps

#### 1. Configure Environment Variables

1. Set shared variables at project level
2. Set service-specific variables for each service
3. Verify all required variables are set

#### 2. Deploy Services

Services are automatically deployed via Railway when code is pushed to the repository.

**Manual Deployment**:
```bash
# Using Railway CLI
railway up

# Or via Railway dashboard
# 1. Go to service
# 2. Click "Deploy"
# 3. Select branch/commit
```

#### 3. Verify Deployment

```bash
# Check API Gateway
curl https://api-gateway-service-production-21fd.up.railway.app/health

# Check Classification Service
curl https://classification-service-production.up.railway.app/health

# Check Merchant Service
curl https://merchant-service-production.up.railway.app/health

# Check Risk Assessment Service
curl https://risk-assessment-service-production.up.railway.app/health
```

#### 4. Verify Service Discovery

```bash
# Check service registry
curl https://service-discovery-production-d397.up.railway.app/api/v1/services
```

---

## Troubleshooting

### Common Issues

#### 1. Service Not Starting

**Symptoms**: Service fails to start or crashes immediately

**Solutions**:
- Check environment variables are set correctly
- Verify Supabase credentials are valid
- Check logs for error messages
- Verify PORT variable is set

#### 2. Database Connection Errors

**Symptoms**: `Failed to initialize Supabase client` or connection timeouts

**Solutions**:
- Verify `SUPABASE_URL` is correct
- Check `SUPABASE_ANON_KEY` and `SUPABASE_SERVICE_ROLE_KEY` are valid
- Verify Supabase project is active
- Check network connectivity

#### 3. CORS Errors

**Symptoms**: CORS errors in browser console

**Solutions**:
- Verify `CORS_ALLOWED_ORIGINS` includes frontend URL
- Check `CORS_ALLOW_CREDENTIALS` is set correctly
- Verify CORS middleware is applied first

#### 4. Rate Limiting Issues

**Symptoms**: `429 Too Many Requests` errors

**Solutions**:
- Check rate limit configuration
- Verify `RATE_LIMIT_REQUESTS_PER` and `RATE_LIMIT_WINDOW_SIZE`
- Wait for rate limit window to reset
- Consider increasing limits for beta testing

#### 5. Health Check Failures

**Symptoms**: Health check returns non-200 status

**Solutions**:
- Check service logs
- Verify database connectivity
- Check service dependencies
- Review health check implementation

### Debugging

#### View Logs

```bash
# Using Railway CLI
railway logs --service api-gateway-service

# Or via Railway dashboard
# 1. Go to service
# 2. Click "Logs" tab
```

#### Test Endpoints

```bash
# Health check
curl https://api-gateway-service-production-21fd.up.railway.app/health

# With detailed info
curl "https://api-gateway-service-production-21fd.up.railway.app/health?detailed=true"

# Metrics
curl https://api-gateway-service-production-21fd.up.railway.app/metrics
```

#### Check Service Status

```bash
# Service discovery
curl https://service-discovery-production-d397.up.railway.app/api/v1/services

# Individual service health
curl https://api-gateway-service-production-21fd.up.railway.app/health
curl https://classification-service-production.up.railway.app/health
curl https://merchant-service-production.up.railway.app/health
curl https://risk-assessment-service-production.up.railway.app/health
```

---

## Service Dependencies

### Dependency Graph

```
Frontend Service
    └──> API Gateway
            ├──> Classification Service
            ├──> Merchant Service
            ├──> Risk Assessment Service
            └──> BI Service

All Services
    └──> Supabase (Database)
```

### Startup Order

1. **Supabase** (external, must be available)
2. **Service Discovery** (optional, for service registry)
3. **Backend Services** (Classification, Merchant, Risk Assessment)
4. **API Gateway** (depends on backend services)
5. **Frontend Service** (depends on API Gateway)

---

## Security Considerations

### Environment Variables

- ✅ Never commit secrets to repository
- ✅ Use Railway secrets management
- ✅ Rotate keys regularly
- ✅ Use different keys for different environments

### Network Security

- ✅ All services use HTTPS
- ✅ CORS configured for allowed origins
- ✅ Rate limiting enabled
- ✅ Security headers implemented

### Database Security

- ✅ Use Supabase Row Level Security (RLS)
- ✅ Use service role key only for backend services
- ✅ Use anon key for public endpoints
- ✅ Validate all inputs before database operations

---

## Backup and Recovery

### Database Backups

Supabase provides automatic backups. Manual backup:

```bash
# Using Supabase CLI
supabase db dump -f backup.sql
```

### Service Recovery

1. **Rollback Deployment**:
   - Go to Railway dashboard
   - Select service
   - Click "Deployments"
   - Select previous working deployment
   - Click "Redeploy"

2. **Restore Environment Variables**:
   - Verify all environment variables are set
   - Check variable values are correct

3. **Restart Services**:
   - Services restart automatically on Railway
   - Or manually restart via dashboard

---

## Performance Tuning

### Recommended Settings

- **Connection Pooling**: Handled by Supabase PostgREST
- **Caching**: Enabled for Classification Service (5-minute TTL)
- **Rate Limiting**: 1000 requests/hour per IP
- **Request Timeout**: 30 seconds (configurable)

### Classification Service Performance (v3.0+)

The Classification Service now uses a **multi-strategy classifier** with the following performance characteristics:

- **Average Response Time**: ~1.4 seconds (72% faster than 5s target)
- **Max Response Time**: ~4.8 seconds (under 5s target)
- **Simple Classification**: ~100µs (name only)
- **Complex Classification**: 1.2-1.6s (with website scraping)

**Performance Breakdown**:
- Website scraping: ~500ms-1.1s (60-70% of total time)
- Multi-strategy classification: ~200-300ms
- Database queries: ~50-100ms
- NLP processing: ~50-100ms

For detailed performance metrics and optimization recommendations, see [Classification Service Deployment Guide](./classification-service-deployment-guide.md).

### Monitoring

- Monitor response times via Prometheus metrics
- Set up alerts for slow queries (> 1 second)
- Monitor error rates
- Track cache hit rates
- **Classification Service**: Monitor response times (< 5s target), success rate (> 99%), error rate (< 1%)

---

## Support

For deployment issues:
- **Documentation**: See this guide and API documentation
- **Logs**: Check Railway logs dashboard
- **Health Checks**: Use `/health` endpoints
- **Metrics**: Check Prometheus metrics

---

---

## Additional Documentation

- **[Classification Service Deployment Guide](./classification-service-deployment-guide.md)** - Detailed guide for multi-strategy classifier (v3.0+)
- **[Performance Benchmark Results](../.cursor/plans/performance-benchmark-results.md)** - Performance test results and metrics
- **[Frontend Integration Test Results](../.cursor/plans/frontend-integration-test-results.md)** - Frontend compatibility verification

---

**Last Updated**: 2025-11-26  
**Version**: 1.1  
**Status**: ✅ Production Ready

