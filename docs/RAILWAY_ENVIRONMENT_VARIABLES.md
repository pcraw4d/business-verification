# Railway Environment Variables Documentation

**Last Updated**: 2025-11-18  
**Status**: Production Configuration Reference

## Overview

This document provides a comprehensive reference for all required and optional environment variables for the KYB Platform services deployed on Railway. All variables must be set in the Railway dashboard before services can function correctly.

---

## Critical Variables (Must Be Set)

### Frontend Service - CRITICAL

**⚠️ MUST BE SET BEFORE BUILD:**

- `NEXT_PUBLIC_API_BASE_URL` - **CRITICAL**
  - **Value**: `https://api-gateway-service-production-21fd.up.railway.app`
  - **Purpose**: Frontend API base URL for all API calls
  - **Default**: `http://localhost:8080` (will fail in production)
  - **Impact**: If not set, all API calls will fail in production
  - **Verification**: Check browser console for API errors

---

## Shared Variables (All Services)

### Supabase Configuration

These variables are required for all services that interact with Supabase:

- `SUPABASE_URL` - **Required**
  - **Example**: `https://qpqhuqqmkjxsltzshfam.supabase.co`
  - **Purpose**: Supabase project URL
  - **Verification**: Check service health endpoint for `supabase_status: {connected: true}`

- `SUPABASE_ANON_KEY` - **Required**
  - **Purpose**: Supabase anonymous key for client-side operations
  - **Security**: Public key, safe for client-side use

- `SUPABASE_SERVICE_ROLE_KEY` - **Optional**
  - **Purpose**: Supabase service role key for admin operations
  - **Security**: Keep secret, server-side only

- `SUPABASE_JWT_SECRET` - **Optional**
  - **Purpose**: JWT secret for token validation
  - **Security**: Keep secret

### Environment Configuration

- `ENV` - **Optional**
  - **Value**: `production`
  - **Purpose**: Environment identifier

- `ENVIRONMENT` - **Optional**
  - **Value**: `production`
  - **Purpose**: Environment identifier (alternative)

- `NODE_ENV` - **Frontend Service Only**
  - **Value**: `production`
  - **Purpose**: Node.js environment

---

## Service-Specific Variables

### API Gateway Service

**Required:**
- `PORT` - Default: `8080`
- `SUPABASE_URL` - Required
- `SUPABASE_ANON_KEY` - Required

**Service URLs (with defaults):**
- `CLASSIFICATION_SERVICE_URL` - Default: `https://classification-service-production.up.railway.app`
- `MERCHANT_SERVICE_URL` - Default: `https://merchant-service-production.up.railway.app`
- `FRONTEND_URL` - Default: `https://frontend-service-production-b225.up.railway.app`
- `BI_SERVICE_URL` - Default: `https://bi-service-production.up.railway.app`
- `RISK_ASSESSMENT_SERVICE_URL` - Default: `https://risk-assessment-service-production.up.railway.app`

**CORS Configuration:**
- `CORS_ALLOWED_ORIGINS` - Default: `*` (should be specific origin in production)
- `CORS_ALLOW_CREDENTIALS` - Default: `true`
- `CORS_ALLOWED_METHODS` - Default: `GET,POST,PUT,DELETE,OPTIONS`
- `CORS_ALLOWED_HEADERS` - Default: `*`
- `CORS_MAX_AGE` - Default: `86400`

**Rate Limiting:**
- `RATE_LIMIT_ENABLED` - Default: `true`
- `RATE_LIMIT_REQUESTS_PER` - Default: `100`
- `RATE_LIMIT_WINDOW_SIZE` - Default: `60` (seconds)
- `RATE_LIMIT_BURST_SIZE` - Default: `200`

### Frontend Service

**CRITICAL:**
- `NEXT_PUBLIC_API_BASE_URL` - **MUST BE SET** - `https://api-gateway-service-production-21fd.up.railway.app`

**Optional:**
- `PORT` - Default: `8086`
- `USE_NEW_UI` - Optional
- `NEXT_PUBLIC_USE_NEW_UI` - Optional
- `NODE_ENV` - Should be `production`

### Classification Service

**Required:**
- `PORT` - Default: `8080`
- `SUPABASE_URL` - Required
- `SUPABASE_ANON_KEY` - Required

### Merchant Service

**Required:**
- `PORT` - Default: `8080` (Dockerfile now uses 8080)
- `SUPABASE_URL` - Required
- `SUPABASE_ANON_KEY` - Required

**Optional:**
- `REDIS_URL` - Optional (for caching)

### Risk Assessment Service

**Required:**
- `PORT` - Default: `8080`
- `SUPABASE_URL` - Required
- `SUPABASE_ANON_KEY` - Required
- `DATABASE_URL` - Required (Supabase Transaction Pooler)

**Optional:**
- `REDIS_URL` - Optional (for caching)
- Multiple ML model configuration variables (see `railway.json`)

### Pipeline Service

**Required:**
- `PORT` - Default: `8085`

**Optional:**
- `SERVICE_NAME` - Default: `kyb-pipeline-service`

### Service Discovery

**Required:**
- `PORT` - Default: `8080` (aligned with Dockerfile)

### BI Service

**Required:**
- `PORT` - Default: `8087`

### Monitoring Service

**Required:**
- `PORT` - Default: `8084`

**Optional:**
- `SERVICE_NAME` - Default: `kyb-monitoring`

---

## Verification Checklist

### Step 1: Verify in Railway Dashboard

1. Go to Railway Dashboard → Your Project
2. Click on each service
3. Go to "Variables" tab
4. Verify all required variables are set
5. Check that `NEXT_PUBLIC_API_BASE_URL` is set for Frontend Service

### Step 2: Verify Service Health

Test each service health endpoint:

```bash
# API Gateway
curl https://api-gateway-service-production-21fd.up.railway.app/health

# Frontend
curl https://frontend-service-production-b225.up.railway.app/health

# Classification
curl https://classification-service-production.up.railway.app/health

# Merchant
curl https://merchant-service-production.up.railway.app/health

# Risk Assessment
curl https://risk-assessment-service-production.up.railway.app/health
```

**Expected Response:**
- Status code: `200 OK`
- JSON with `status: "healthy"`
- `supabase_status: {connected: true}` (for services using Supabase)

### Step 3: Verify Frontend API Calls

1. Open frontend in browser: `https://frontend-service-production-b225.up.railway.app`
2. Open browser DevTools → Network tab
3. Check that API calls are going to correct URL
4. Verify no CORS errors
5. Check console for API configuration warnings

### Step 4: Check Railway Logs

1. Go to Railway Dashboard → Service → Logs
2. Look for environment variable errors
3. Check for "missing environment variable" messages
4. Verify Supabase connection status

---

## Common Issues and Solutions

### Issue: Frontend API calls failing

**Symptoms:**
- API calls return 404 or connection errors
- Browser console shows "API base URL is set to localhost in production!"

**Solution:**
1. Verify `NEXT_PUBLIC_API_BASE_URL` is set in Railway
2. Value should be: `https://api-gateway-service-production-21fd.up.railway.app`
3. Redeploy frontend service after setting variable

### Issue: Supabase connection failures

**Symptoms:**
- Health check shows `supabase_status: {connected: false}`
- Database operation errors

**Solution:**
1. Verify `SUPABASE_URL` is correct
2. Verify `SUPABASE_ANON_KEY` is set and valid
3. Check Supabase project is active
4. Verify network connectivity

### Issue: CORS errors in browser

**Symptoms:**
- Browser console shows CORS errors
- API calls blocked by browser

**Solution:**
1. Verify `CORS_ALLOWED_ORIGINS` is set correctly
2. Should be specific origin (not `*`) if `CORS_ALLOW_CREDENTIALS=true`
3. Recommended: `https://frontend-service-production-b225.up.railway.app`
4. Clear browser cache and retry

### Issue: Service port conflicts

**Symptoms:**
- Service fails to start
- Port already in use errors

**Solution:**
1. Verify `PORT` environment variable matches Dockerfile
2. Railway automatically sets `PORT` - don't override unless necessary
3. Check service-specific port requirements

---

## Environment Variable Priority

1. **Railway Dashboard Variables** (highest priority)
2. **Service-specific variables** (override shared)
3. **Default values in code** (lowest priority)

---

## Setting Variables in Railway

### Method 1: Project-Level Variables (Shared)

1. Go to Railway Dashboard → Project
2. Click "Variables" tab
3. Add shared variables (SUPABASE_*, CORS_*, etc.)
4. These apply to all services

### Method 2: Service-Level Variables (Specific)

1. Go to Railway Dashboard → Service
2. Click "Variables" tab
3. Add service-specific variables
4. These override project-level variables

### Method 3: Using Railway CLI

```bash
railway variables set NEXT_PUBLIC_API_BASE_URL=https://api-gateway-service-production-21fd.up.railway.app --service frontend-service
```

---

## Security Best Practices

1. **Never commit secrets to git**
   - Use Railway variables for all secrets
   - Use `.env.example` files for documentation

2. **Use service role keys only server-side**
   - Never expose `SUPABASE_SERVICE_ROLE_KEY` to frontend

3. **Rotate keys regularly**
   - Update Supabase keys periodically
   - Update Railway variables when keys change

4. **Use specific CORS origins**
   - Avoid wildcard (`*`) with credentials
   - Use specific frontend URL

5. **Monitor variable usage**
   - Check Railway logs for missing variables
   - Set up alerts for configuration errors

---

## Quick Reference Table

| Variable | Service | Required | Default | Notes |
|----------|---------|----------|---------|-------|
| `NEXT_PUBLIC_API_BASE_URL` | Frontend | ✅ Critical | `http://localhost:8080` | Must be set in production |
| `SUPABASE_URL` | All | ✅ Required | None | Supabase project URL |
| `SUPABASE_ANON_KEY` | All | ✅ Required | None | Supabase anonymous key |
| `PORT` | All | Optional | Service-specific | Railway sets automatically |
| `CORS_ALLOWED_ORIGINS` | API Gateway | Optional | `*` | Should be specific in production |
| `CORS_ALLOW_CREDENTIALS` | API Gateway | Optional | `true` | Cannot use with wildcard |

---

## Additional Resources

- Railway Documentation: https://docs.railway.app/
- Supabase Documentation: https://supabase.com/docs
- Environment Variable Best Practices: See project README

---

**Document Version**: 1.0.0  
**Last Updated**: 2025-11-18  
**Next Review**: 2025-12-18

