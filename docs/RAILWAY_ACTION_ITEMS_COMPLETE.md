# Railway Deployment Action Items - Completion Guide

**Date**: November 13, 2025  
**Status**: In Progress

---

## ✅ Action Item 1: Review Railway Dashboard Logs

### Service Discovery Logs Analysis

From the service discovery logs:
- ✅ **8/10 services healthy** - Good overall health
- ⚠️ **legacy-frontend-service** - Returning 404 (expected - legacy service)
- ⚠️ **legacy-api-service** - Returning 404 (expected - legacy service)

**Action**: These are legacy services that can be ignored or removed from service discovery.

### Log Review Checklist

**For each service, check Railway dashboard logs for:**

#### API Gateway
- [ ] "🚀 Starting KYB API Gateway Service" - Service started
- [ ] "✅ Supabase client initialized successfully" - Supabase connected
- [ ] "🌐 API Gateway server starting" - Server running
- [ ] Route registration messages
- [ ] Any error messages

#### Classification Service
- [ ] "Server starting" message
- [ ] "Connected to database" or Supabase connection
- [ ] "Redis connection established" (if using Redis)
- [ ] Route registration messages

#### Merchant Service
- [ ] "Server starting" message
- [ ] Database connection messages
- [ ] Redis connection messages
- [ ] Route registration messages

#### Risk Assessment Service
- [ ] "Server starting" message
- [ ] Database connection messages
- [ ] Redis connection messages
- [ ] Model loading messages

#### Frontend Service
- [ ] "Server starting" message
- [ ] Static file serving messages
- [ ] API Gateway connection messages

#### Redis Cache
- [ ] "Redis server started" messages
- [ ] Configuration loaded messages
- [ ] No error messages

**How to Check:**
1. Go to Railway Dashboard
2. Select each service
3. Click "Logs" tab
4. Look for the messages above
5. Note any warnings or errors

---

## ✅ Action Item 2: Verify Environment Variables

### Required Environment Variables Checklist

#### All Services (Shared Variables)

**Supabase Configuration:**
- [ ] `SUPABASE_URL` - Should be: `https://qpqhuqqmkjxsltzshfam.supabase.co`
- [ ] `SUPABASE_ANON_KEY` - Should be set (check Railway dashboard)
- [ ] `SUPABASE_SERVICE_ROLE_KEY` - Should be set (check Railway dashboard)
- [ ] `SUPABASE_JWT_SECRET` - Should be set (check Railway dashboard)

**Redis Configuration:**
- [ ] `REDIS_URL` - Should be: `redis://redis-cache:6379` (for internal communication)

**Environment:**
- [ ] `ENVIRONMENT` - Should be: `production`
- [ ] `ENV` - Should be: `production`

**Logging:**
- [ ] `LOG_LEVEL` - Should be: `info` or `debug`
- [ ] `LOG_FORMAT` - Should be: `json`

#### API Gateway Specific

**Backend Service URLs:**
- [ ] `CLASSIFICATION_SERVICE_URL` - Should be: `https://classification-service-production.up.railway.app`
- [ ] `MERCHANT_SERVICE_URL` - Should be: `https://merchant-service-production.up.railway.app`
- [ ] `RISK_ASSESSMENT_SERVICE_URL` - Should be: `https://risk-assessment-service-production.up.railway.app`
- [ ] `FRONTEND_URL` - Should be: `https://frontend-service-production-b225.up.railway.app`

**CORS Configuration:**
- [ ] `CORS_ALLOWED_ORIGINS` - Should be set (e.g., `*` or specific domains)
- [ ] `CORS_ALLOWED_METHODS` - Should include: `GET,POST,PUT,DELETE,OPTIONS`
- [ ] `CORS_ALLOWED_HEADERS` - Should be set (e.g., `*`)

**Rate Limiting:**
- [ ] `RATE_LIMIT_ENABLED` - Should be: `true`
- [ ] `RATE_LIMIT_REQUESTS_PER` - Should be set (e.g., `1000`)
- [ ] `RATE_LIMIT_WINDOW_SIZE` - Should be set (e.g., `3600`)

**How to Verify:**
1. Go to Railway Dashboard
2. Select each service
3. Click "Variables" tab
4. Verify all variables listed above are set
5. For shared variables, check if they're set at project level

---

## ✅ Action Item 3: Configure Missing Routes

### Route Configuration Analysis

**Routes ARE registered in code:**
- ✅ `/api/v1/classify` - Registered as POST (line 101 in main.go)
- ✅ `/api/v1/risk/assess` - Registered as POST (line 118 in main.go)

**Why routes might return 404:**

1. **Authentication Middleware Blocking**
   - Routes are marked as public in `auth.go` (lines 97, 104-105)
   - But middleware might still be blocking

2. **Handler Implementation Issues**
   - `ProxyToClassification` might be failing
   - `ProxyToRiskAssessment` might be failing

3. **Service URL Configuration**
   - Backend service URLs might not be configured
   - Services might not be reachable

### Debugging Steps

#### Step 1: Test Routes Directly

```bash
# Test classify endpoint (POST required)
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Business"}'

# Test risk assess endpoint (POST required)
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/assess \
  -H "Content-Type: application/json" \
  -d '{"business_id": "test-123"}'
```

#### Step 2: Check API Gateway Logs

Look for:
- Route matching messages
- Handler execution messages
- Proxy request messages
- Error messages from handlers

#### Step 3: Verify Service URLs

Check that API Gateway has:
- `CLASSIFICATION_SERVICE_URL` set correctly
- `RISK_ASSESSMENT_SERVICE_URL` set correctly

#### Step 4: Test Backend Services Directly

```bash
# Test classification service directly
curl -X POST https://classification-service-production.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Business"}'

# Test risk assessment service directly
curl -X POST https://risk-assessment-service-production.up.railway.app/api/v1/risk/assess \
  -H "Content-Type: application/json" \
  -d '{"business_id": "test-123"}'
```

### Potential Fixes

#### Fix 1: Ensure Routes Accept GET for Testing

If routes only accept POST, test with POST method:

```bash
# Use POST method, not GET
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Business"}'
```

#### Fix 2: Check Handler Implementation

Review `services/api-gateway/internal/handlers/gateway.go`:
- `ProxyToClassification` function
- `ProxyToRiskAssessment` function
- Ensure they're handling requests correctly

#### Fix 3: Verify Service URLs in Environment

Ensure API Gateway has correct backend service URLs configured.

---

## ✅ Action Item 4: Set Up Monitoring Alerts

### Railway Built-in Alerts

#### Step 1: Configure Service Down Alerts

1. Go to Railway Dashboard
2. Select your project
3. Go to "Settings" → "Alerts" (or "Notifications")
4. Create alert:
   - **Name**: "Service Down"
   - **Trigger**: Service deployment fails or becomes unresponsive
   - **Action**: Email notification
   - **Threshold**: Service unavailable for > 2 minutes

#### Step 2: Configure High Error Rate Alerts

1. Create alert:
   - **Name**: "High Error Rate"
   - **Trigger**: Error rate > 5% for 5 minutes
   - **Action**: Email/Slack notification
   - **Threshold**: 5% error rate sustained for 5 minutes

#### Step 3: Configure Resource Usage Alerts

1. **CPU Alert**:
   - **Name**: "High CPU Usage"
   - **Trigger**: CPU usage > 80% for 10 minutes
   - **Action**: Email notification

2. **Memory Alert**:
   - **Name**: "High Memory Usage"
   - **Trigger**: Memory usage > 90% for 5 minutes
   - **Action**: Email notification

### External Monitoring (Optional)

#### Option 1: Uptime Robot (Free)

1. Sign up at https://uptimerobot.com
2. Add monitors for each service health endpoint:
   - API Gateway: `https://api-gateway-service-production-21fd.up.railway.app/health`
   - Classification: `https://classification-service-production.up.railway.app/health`
   - Merchant: `https://merchant-service-production.up.railway.app/health`
   - Risk Assessment: `https://risk-assessment-service-production.up.railway.app/health`
   - Frontend: `https://frontend-service-production-b225.up.railway.app/health`

3. Configure alerting:
   - Email notifications
   - Check interval: 5 minutes
   - Alert on: Service down for 2 minutes

#### Option 2: StatusCake (Free Tier)

1. Sign up at https://www.statuscake.com
2. Create uptime tests for all health endpoints
3. Configure alerting channels

### Monitoring Dashboard

Create a simple monitoring dashboard showing:
- Service health status
- Uptime percentage
- Response times
- Error rates

---

## 📋 Verification Checklist

### Completed ✅
- [x] All services deployed successfully
- [x] Health endpoints tested and working
- [x] Production URLs documented
- [x] Verification script created
- [x] Monitoring guide created

### In Progress ⚠️
- [ ] Railway dashboard logs reviewed
- [ ] Environment variables verified
- [ ] Route configuration debugged
- [ ] Monitoring alerts configured

### Next Steps 🎯
1. Review Railway dashboard logs for each service
2. Verify all environment variables are set
3. Debug route 404 issues (test with POST method)
4. Configure Railway alerts
5. Set up external monitoring (optional)

---

## 🔍 Route Debugging Commands

```bash
# Test classify with POST (correct method)
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Business", "website": "https://example.com"}'

# Test risk assess with POST (correct method)
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/risk/assess \
  -H "Content-Type: application/json" \
  -d '{"business_id": "test-123", "assessment_type": "standard"}'

# Check API Gateway root endpoint (shows available routes)
curl https://api-gateway-service-production-21fd.up.railway.app/
```

---

## 📝 Notes

- Routes are registered in code but may need POST method (not GET)
- Legacy services (legacy-frontend-service, legacy-api-service) can be ignored
- Service discovery shows 8/10 healthy (2 are legacy services)
- All core services are healthy and responding

