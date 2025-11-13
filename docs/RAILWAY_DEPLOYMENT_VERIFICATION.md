# Railway Deployment Verification Guide

## ✅ Deployment Status

All services have been successfully deployed to Railway:
- ✅ **api-gateway** - Deployed and passing
- ✅ **redis-cache** - Deployed and passing
- ✅ **classification-service** - Deployed
- ✅ **merchant-service** - Deployed
- ✅ **risk-assessment-service** - Deployed
- ✅ **frontend** - Deployed

## 🔍 Next Steps: Verification & Testing

### 1. Verify Service Health

Check that all services are responding to health checks:

```bash
# API Gateway
curl https://api-gateway-production.up.railway.app/health

# Classification Service
curl https://classification-service-production.up.railway.app/health

# Merchant Service
curl https://merchant-service-production.up.railway.app/health

# Risk Assessment Service
curl https://risk-assessment-service-production.up.railway.app/health

# Frontend
curl https://frontend-service-production.up.railway.app/health
```

**Expected**: All should return `200 OK` with health status JSON.

### 2. Verify Redis Connectivity

Check that services can connect to Redis:

```bash
# From Railway dashboard, check service logs for:
# - "Redis connection established"
# - "Redis health check passed"
# - No connection errors
```

**Expected**: All services using Redis should show successful connections.

### 3. Test API Gateway Routing

Verify that the API Gateway correctly routes requests to backend services:

```bash
# Test classification endpoint
curl https://api-gateway-production.up.railway.app/api/v1/classify

# Test merchant endpoint
curl https://api-gateway-production.up.railway.app/api/v1/merchants

# Test risk assessment endpoint
curl https://api-gateway-production.up.railway.app/api/v1/risk/assess
```

**Expected**: Requests should be proxied to the correct backend services.

### 4. Verify Environment Variables

In Railway dashboard, verify that all services have required environment variables:

**Shared Variables (all services):**
- `SUPABASE_URL`
- `SUPABASE_ANON_KEY`
- `SUPABASE_SERVICE_ROLE_KEY`
- `SUPABASE_JWT_SECRET`

**Service-Specific Variables:**
- **API Gateway**: `CLASSIFICATION_SERVICE_URL`, `MERCHANT_SERVICE_URL`, `RISK_ASSESSMENT_SERVICE_URL`
- **All Services**: `REDIS_URL=redis://redis-cache:6379`
- **Each Service**: `PORT` (set by Railway automatically)

### 5. Check Service Logs

Review logs for each service in Railway dashboard:

**Look for:**
- ✅ "Server starting on port XXXX"
- ✅ "Connected to database"
- ✅ "Redis connection established"
- ✅ "Routes registered successfully"
- ❌ Any error messages or warnings

### 6. Test End-to-End Workflow

Test a complete business verification flow:

```bash
# 1. Create a merchant
curl -X POST https://api-gateway-production.up.railway.app/api/v1/merchants \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Business", "address": "123 Test St"}'

# 2. Classify the business
curl -X POST https://api-gateway-production.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Business"}'

# 3. Perform risk assessment
curl -X POST https://api-gateway-production.up.railway.app/api/v1/risk/assess \
  -H "Content-Type: application/json" \
  -d '{"business_id": "test-123"}'
```

### 7. Monitor Service Metrics

In Railway dashboard, check:
- **CPU Usage**: Should be reasonable for each service
- **Memory Usage**: Should be within limits
- **Request Count**: Should show activity
- **Error Rate**: Should be low or zero

### 8. Verify Service Discovery

Check that services can communicate with each other:

```bash
# API Gateway should be able to reach:
# - Classification Service
# - Merchant Service
# - Risk Assessment Service
# - Redis Cache
```

**Expected**: All inter-service communication should work via Railway's internal network.

## 🐛 Troubleshooting

### If Services Are Not Responding

1. **Check Railway Dashboard**:
   - Verify services show "Deployed" status
   - Check for any deployment errors
   - Review service logs

2. **Check Environment Variables**:
   - Ensure all required variables are set
   - Verify Supabase credentials are correct
   - Check Redis URL is set correctly

3. **Check Service Logs**:
   - Look for startup errors
   - Check for connection failures
   - Verify port configuration

### If Redis Connection Fails

1. **Verify Redis Service**:
   - Check Redis service is running
   - Verify Redis URL in environment variables
   - Test Redis connectivity from another service

2. **Check Network Configuration**:
   - Services should use `redis://redis-cache:6379` for internal communication
   - Verify Railway service discovery is working

### If API Gateway Routing Fails

1. **Check Service URLs**:
   - Verify `CLASSIFICATION_SERVICE_URL` is correct
   - Verify `MERCHANT_SERVICE_URL` is correct
   - Verify `RISK_ASSESSMENT_SERVICE_URL` is correct

2. **Check Backend Services**:
   - Ensure backend services are running
   - Verify backend services are responding to health checks

## 📊 Success Criteria

All of the following should be true:
- ✅ All services show "Deployed" status in Railway
- ✅ All health endpoints return `200 OK`
- ✅ API Gateway successfully routes requests
- ✅ Services can connect to Redis
- ✅ Services can connect to Supabase
- ✅ No critical errors in service logs
- ✅ End-to-end workflows complete successfully

## 🎯 Post-Deployment Tasks

1. **Set up monitoring alerts** for:
   - Service downtime
   - High error rates
   - Resource usage spikes

2. **Configure custom domains** (if needed):
   - Set up domain for API Gateway
   - Configure SSL certificates

3. **Set up CI/CD** (if not already):
   - Configure automatic deployments on git push
   - Set up staging environment

4. **Document service URLs**:
   - Update documentation with production URLs
   - Share URLs with team members

5. **Performance testing**:
   - Load test the API Gateway
   - Test Redis caching performance
   - Verify service scalability

## 📝 Summary of Fixes Applied

### API Gateway
- Fixed Railway configuration to use repository root source
- Updated Dockerfile path resolution
- Fixed service-level railway.json configuration

### Redis Cache
- Fixed Railway configuration to use repository root source
- Updated Dockerfile to use correct path for redis.conf
- Removed invalid healthcheckPath (Redis doesn't use HTTP health checks)

### All Services
- Standardized Railway configuration patterns
- Fixed build context issues
- Ensured proper Dockerfile path resolution

