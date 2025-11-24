# Railway Dashboard Checklist - bi-service Configuration

**Date:** November 24, 2025  
**Service:** `bi-service`  
**URL:** `https://bi-service-production.up.railway.app`

---

## Configuration Checklist

### 1. Service Status
- [ ] Service is **Running** (not stopped/crashed)
- [ ] Latest deployment completed successfully
- [ ] No build errors in deployment logs

### 2. Environment Variables
- [ ] `PORT` environment variable is set (Railway sets this automatically)
- [ ] `SERVICE_NAME` environment variable (optional, defaults to "bi-service")
- [ ] Verify no conflicting or incorrect environment variables

### 3. Service Settings
- [ ] **Root Directory:** `cmd/business-intelligence-gateway`
- [ ] **Builder Type:** Dockerfile (NOT Railpack)
- [ ] **Dockerfile Path:** `Dockerfile`
- [ ] **Start Command:** `./kyb-business-intelligence-gateway` (or auto-detected)

### 4. Health Check Configuration
- [ ] **Health Check Path:** `/health`
- [ ] **Health Check Timeout:** 300 seconds (or appropriate value)
- [ ] **Health Check Interval:** 30 seconds (or appropriate value)
- [ ] Health check is enabled

### 5. Port Configuration
- [ ] Railway automatically sets `PORT` environment variable
- [ ] Service should listen on `0.0.0.0:${PORT}`
- [ ] Verify port is not hardcoded in code

### 6. Network Configuration
- [ ] Service is publicly accessible
- [ ] No network restrictions blocking access
- [ ] Service is properly linked in Railway project

### 7. Logs Review
Check recent logs for:
- [ ] Service startup messages
- [ ] "Starting on 0.0.0.0:PORT" message
- [ ] "ready and listening" message
- [ ] Any error or panic messages
- [ ] Port binding confirmation

---

## Common Issues and Solutions

### Issue: Service Returns 502
**Possible Causes:**
1. Service not starting (check logs for errors)
2. Port mismatch (verify PORT env var)
3. Health check failing (check health check path)
4. Service binding to wrong interface (should be 0.0.0.0)

### Issue: Service Not Accessible
**Possible Causes:**
1. Service not running (check status)
2. Network configuration issue
3. Railway proxy not routing correctly
4. Service not properly linked

### Issue: Health Check Failing
**Possible Causes:**
1. Health check path incorrect
2. Service not responding on health endpoint
3. Health check timeout too short
4. Service taking too long to start

---

## Verification Steps

1. **Check Service Logs:**
   - Look for startup messages
   - Verify service is listening on correct port
   - Check for any errors

2. **Test Health Endpoint:**
   ```bash
   curl https://bi-service-production.up.railway.app/health
   ```

3. **Check Environment Variables:**
   - Verify PORT is set
   - Check for any conflicting variables

4. **Verify Service Configuration:**
   - Root directory is correct
   - Builder type is Dockerfile
   - Health check is configured

---

**Last Updated:** November 24, 2025  
**Status:** 📋 **CHECKLIST READY** - Use this to verify Railway configuration

