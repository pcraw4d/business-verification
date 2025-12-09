# Railway Service Health Check Fixes

## Summary

Five services are showing warning indicators in Railway dashboard:

1. **service-discovery** - ▲ 10 warnings (health check failures for other services)
2. **pipeline-service** - ▲ 10 warnings (502 errors from service-discovery)
3. **bi-service** - ▲ 10 warnings
4. **monitoring-service** - ▲ 11 warnings (502 errors from service-discovery)
5. **risk-assessment-service** - ▲ 9 warnings

## Root Cause Analysis

### Services Status

- ✅ All services **build successfully**
- ✅ All services **start correctly** and bind to ports
- ✅ Services use `PORT` environment variable correctly
- ✅ Health endpoints exist at `/health` for all services
- ❌ **service-discovery** is getting **502 Bad Gateway** errors when checking:
  - `pipeline-service`
  - `monitoring-service`

### Findings

1. **risk-assessment-service** is actually healthy - logs show 200 responses
2. **pipeline-service** and **monitoring-service** started on Dec 2, but service-discovery continues to report 502s
3. No recent error logs found for failing services
4. Services are configured correctly with health check paths

## Likely Issues

### 1. Service Restart Loop

Services may be crashing and restarting, causing intermittent 502s during restarts.

### 2. Network Connectivity

service-discovery may not be able to reach other services due to:

- Network policies
- Service discovery configuration
- Internal service URLs

### 3. Health Check Timeout

Health check timeouts may be too short for services that take time to start.

## Recommended Fixes

### Fix 1: Redeploy Services

Redeploy the failing services to ensure they're running the latest code:

```bash
# Trigger redeploy via Railway CLI or dashboard
railway redeploy --service pipeline-service
railway redeploy --service monitoring-service
railway redeploy --service bi-service
```

### Fix 2: Verify Service URLs in service-discovery

Check that service-discovery is using correct internal URLs for health checks.

### Fix 3: Increase Health Check Timeout

Current timeouts:

- pipeline-service: 300s (5 minutes) ✅
- monitoring-service: Not specified in railway.json (may default to 30s) ⚠️
- bi-service: 300s ✅

**Action**: Add explicit health check timeout to monitoring-service railway.json:

```json
{
  "deploy": {
    "healthcheckTimeout": 300,
    "healthcheckInterval": 30
  }
}
```

### Fix 4: Check Service Discovery Configuration

Verify service-discovery is correctly configured to find and check other services.

## Immediate Actions

1. **Redeploy failing services** to refresh their state
2. **Check Railway dashboard** for service restart counts
3. **Review service-discovery logs** for detailed error messages
4. **Verify internal networking** between services

## Files to Review

- `cmd/service-discovery/main.go` - Service discovery logic
- `cmd/pipeline-service/railway.json` - Pipeline service config
- `cmd/monitoring-service/` - Monitoring service (no railway.json found - needs creation)
- `cmd/business-intelligence-gateway/railway.json` - BI service config

## Next Steps

1. Create `railway.json` for monitoring-service with proper health check config
2. Redeploy all failing services
3. Monitor logs after redeploy
4. Verify health check endpoints are accessible
