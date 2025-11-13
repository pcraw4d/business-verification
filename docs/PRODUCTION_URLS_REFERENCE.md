# Production URLs Reference - KYB Platform

**Last Updated**: November 13, 2025  
**Status**: ✅ All Services Deployed

## 🌐 Core Service URLs

### API Gateway (Main Entry Point)
- **URL**: `https://api-gateway-service-production-21fd.up.railway.app`
- **Health**: `https://api-gateway-service-production-21fd.up.railway.app/health`
- **Status**: ✅ Active
- **Purpose**: Main API entry point, routes requests to backend services

### Classification Service
- **URL**: `https://classification-service-production.up.railway.app`
- **Health**: `https://classification-service-production.up.railway.app/health`
- **Status**: ✅ Active
- **Purpose**: Business classification and industry code detection

### Merchant Service
- **URL**: `https://merchant-service-production.up.railway.app`
- **Health**: `https://merchant-service-production.up.railway.app/health`
- **Status**: ✅ Active
- **Purpose**: Merchant management and CRUD operations

### Risk Assessment Service
- **URL**: `https://risk-assessment-service-production.up.railway.app`
- **Health**: `https://risk-assessment-service-production.up.railway.app/health`
- **Status**: ✅ Active
- **Purpose**: Risk scoring and assessment

### Frontend Service
- **URL**: `https://frontend-service-production-b225.up.railway.app`
- **Health**: `https://frontend-service-production-b225.up.railway.app/health`
- **Status**: ✅ Active
- **Purpose**: Web interface and user-facing application

### Redis Cache
- **Internal URL**: `redis://redis-cache:6379` (within Railway network)
- **Status**: ✅ Active
- **Purpose**: Caching layer for services

## 🔗 API Endpoints

### Via API Gateway

#### Health Check
```bash
GET https://api-gateway-service-production-21fd.up.railway.app/health
```

#### Merchant Endpoints
```bash
# List merchants
GET https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants

# Get merchant by ID
GET https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants/{id}

# Create merchant
POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants
```

#### Risk Assessment Endpoints
```bash
# Assess risk
POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/assess

# Get benchmarks
GET https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/benchmarks?mcc=5411

# Get predictions
GET https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/predictions/{merchant_id}
```

#### Classification Endpoints
```bash
# Classify business
POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify
```

### Direct Service Access

#### Classification Service
```bash
# Health
GET https://classification-service-production.up.railway.app/health

# Classify
POST https://classification-service-production.up.railway.app/api/v1/classify
```

#### Merchant Service
```bash
# Health
GET https://merchant-service-production.up.railway.app/health

# Merchants
GET https://merchant-service-production.up.railway.app/api/v1/merchants
```

#### Risk Assessment Service
```bash
# Health
GET https://risk-assessment-service-production.up.railway.app/health

# Benchmarks
GET https://risk-assessment-service-production.up.railway.app/api/v1/risk/benchmarks?mcc=5411
```

## 🔐 Internal Service Communication

Services communicate via Railway's internal network using service names:

- **Redis**: `redis://redis-cache:6379`
- **API Gateway → Classification**: `http://classification-service:8081`
- **API Gateway → Merchant**: `http://merchant-service:8080`
- **API Gateway → Risk Assessment**: `http://risk-assessment-service:8080`

## 📊 Service Status

| Service | Health Status | Last Verified |
|---------|--------------|---------------|
| API Gateway | ✅ Healthy | 2025-11-13 |
| Classification | ✅ Healthy | 2025-11-13 |
| Merchant | ✅ Healthy | 2025-11-13 |
| Risk Assessment | ✅ Healthy | 2025-11-13 |
| Frontend | ✅ Healthy | 2025-11-13 |
| Redis Cache | ✅ Active | 2025-11-13 |

## 🧪 Quick Test Commands

```bash
# Test all health endpoints
curl https://api-gateway-service-production-21fd.up.railway.app/health
curl https://classification-service-production.up.railway.app/health
curl https://merchant-service-production.up.railway.app/health
curl https://risk-assessment-service-production.up.railway.app/health
curl https://frontend-service-production-b225.up.railway.app/health

# Test API Gateway routing
curl https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants
curl https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/assess
```

## ⚠️ Important Notes

1. **API Gateway is the primary entry point** - Use it for all client requests
2. **Direct service access** - Only for debugging or internal use
3. **Redis is internal only** - Not accessible from outside Railway network
4. **HTTPS is required** - All production URLs use HTTPS
5. **Service discovery** - Services use Railway's internal DNS for communication

## 🔄 URL Updates

If service URLs change:
1. Update this document
2. Update `docs/RAILWAY-SERVICE-URLS.md`
3. Update environment variables in Railway dashboard
4. Update frontend API configuration if needed

## 📝 Environment Variables Reference

Services should have these URLs configured:

**API Gateway:**
- `CLASSIFICATION_SERVICE_URL=https://classification-service-production.up.railway.app`
- `MERCHANT_SERVICE_URL=https://merchant-service-production.up.railway.app`
- `RISK_ASSESSMENT_SERVICE_URL=https://risk-assessment-service-production.up.railway.app`
- `FRONTEND_URL=https://frontend-service-production-b225.up.railway.app`

**All Services:**
- `REDIS_URL=redis://redis-cache:6379` (internal)

