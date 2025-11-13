# Railway Deployment Fixes for redis-cache and api-gateway

## Issues Identified

### 1. Redis Cache Service
**Problem**: Missing `dockerContext` in `railway.json` causing build context issues
**Problem**: `healthcheckPath: "/ping"` is invalid - Railway expects HTTP endpoints, but Redis uses `redis-cli ping`

**Fix Applied**:
- Added `"dockerContext": "."` to build from the service directory
- Removed `healthcheckPath` (Railway will use container health checks instead)

### 2. API Gateway Service  
**Problem**: Missing `dockerContext` in `railway.json` causing build context issues

**Fix Applied**:
- Added `"dockerContext": "../.."` to build from repository root (needed for Go module resolution)

## Changes Made

### railway.json Updates

#### Redis Cache Service
```json
{
  "name": "redis-cache",
  "source": "services/redis-cache",
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile",
    "dockerContext": "."  // ← Added: Build from service directory
  },
  "deploy": {
    "startCommand": "redis-server /usr/local/etc/redis/redis.conf",
    // Removed healthcheckPath - Railway doesn't support HTTP health checks for Redis
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 3
  }
}
```

#### API Gateway Service
```json
{
  "name": "api-gateway",
  "source": "services/api-gateway",
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile",
    "dockerContext": "../.."  // ← Added: Build from repository root
  },
  "deploy": {
    "startCommand": "./api-gateway",
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10,
    "numReplicas": 1,
    "healthcheckPath": "/health",
    "healthcheckTimeout": 30,
    "healthcheckInterval": 60
  }
}
```

## Why These Fixes Work

### Redis Cache
- **dockerContext: "."**: Since Redis only needs `redis.conf` from its own directory, building from the service directory is sufficient
- **Removed healthcheckPath**: Railway's HTTP health check doesn't work for Redis. Railway will use Docker's built-in health checks or monitor the container status instead

### API Gateway
- **dockerContext: "../.."**: The API Gateway Dockerfile needs access to the full repository structure because:
  1. It copies the entire repository (`COPY . .`)
  2. Then changes to the service directory (`WORKDIR /app/services/api-gateway`)
  3. This allows Go to resolve both service-specific and shared packages

## Verification Steps

After deploying, verify:

1. **Check Railway Dashboard**:
   - Both services should show "Deployed" status
   - Check build logs for any errors
   - Verify services are running

2. **Test Redis**:
   ```bash
   # From another service or Railway CLI
   redis-cli -h redis-cache ping
   # Should return: PONG
   ```

3. **Test API Gateway**:
   ```bash
   curl https://api-gateway-production.up.railway.app/health
   # Should return: {"status":"healthy",...}
   ```

4. **Check Service Logs**:
   ```bash
   railway logs --service redis-cache
   railway logs --service api-gateway
   ```

## Additional Notes

- Railway will automatically retry failed deployments based on `restartPolicyType` and `restartPolicyMaxRetries`
- If services still fail after these fixes, check Railway logs for specific error messages
- Ensure environment variables are properly set in Railway dashboard

