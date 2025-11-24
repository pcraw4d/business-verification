# ERROR #4 Railway Investigation Guide

**Date:** November 24, 2025  
**Service:** `bi-service`  
**Issue:** Service starts successfully but returns 502 Bad Gateway

---

## Current Status

**Service Logs Show:**
- ✅ Service starts successfully
- ✅ Service reports "ready and listening on :8080"
- ✅ Service reports "Starting on 0.0.0.0:8080" (from our fix)
- ❌ External requests return 502 Bad Gateway

**Conclusion:** Service is running but Railway's proxy cannot route to it.

---

## Railway Dashboard Investigation Steps

### Step 1: Verify Service Status

**Location:** Railway Dashboard → `bi-service` → Overview

**Check:**
- [ ] Service status is **"Running"** (not "Stopped" or "Crashed")
- [ ] Latest deployment shows **"Deployed"** status
- [ ] No deployment errors or build failures

**If Service is Not Running:**
- Check deployment logs for errors
- Verify build completed successfully
- Check for startup errors

---

### Step 2: Check Service Settings

**Location:** Railway Dashboard → `bi-service` → Settings

**Verify:**
- [ ] **Root Directory:** `cmd/business-intelligence-gateway`
- [ ] **Builder Type:** `Dockerfile` (NOT "Railpack" or "Nixpacks")
- [ ] **Dockerfile Path:** `Dockerfile` (or leave blank if in root directory)
- [ ] **Start Command:** `./kyb-business-intelligence-gateway` (or auto-detected)

**Common Issues:**
- Wrong root directory → Service can't find Dockerfile
- Wrong builder type → Service won't build correctly
- Missing start command → Service won't start

---

### Step 3: Check Environment Variables

**Location:** Railway Dashboard → `bi-service` → Variables

**Verify:**
- [ ] `PORT` environment variable is set (Railway sets this automatically)
- [ ] No conflicting `PORT` values
- [ ] No incorrect environment variables

**Note:** Railway automatically sets `PORT` - don't manually set it unless needed.

---

### Step 4: Check Health Check Configuration

**Location:** Railway Dashboard → `bi-service` → Settings → Health Check

**Verify:**
- [ ] **Health Check Path:** `/health`
- [ ] **Health Check Timeout:** 300 seconds (or appropriate value)
- [ ] **Health Check Interval:** 30 seconds (or appropriate value)
- [ ] Health check is **enabled**

**If Health Check is Failing:**
- Service might not be responding on health endpoint
- Health check timeout might be too short
- Service might be taking too long to start

---

### Step 5: Review Service Logs

**Location:** Railway Dashboard → `bi-service` → Logs

**Look For:**
- [ ] "🚀 Starting bi-service v4.0.4-BI-SYNTAX-FIX-FINAL on 0.0.0.0:PORT"
- [ ] "✅ bi-service v4.0.4-BI-SYNTAX-FIX-FINAL is ready and listening"
- [ ] Any error or panic messages
- [ ] Port binding confirmation

**Expected Log Messages:**
```
🚀 Starting bi-service v4.0.4-BI-SYNTAX-FIX-FINAL on 0.0.0.0:8080
✅ bi-service v4.0.4-BI-SYNTAX-FIX-FINAL is ready and listening on :8080
🔗 Health: http://localhost:8080/health
```

**If Logs Show Errors:**
- Note the error message
- Check if service is crashing
- Verify all dependencies are available

---

### Step 6: Check Network Configuration

**Location:** Railway Dashboard → `bi-service` → Settings → Network

**Verify:**
- [ ] Service is publicly accessible
- [ ] No network restrictions
- [ ] Service is properly linked in Railway project

---

## Common Issues and Solutions

### Issue 1: Wrong Root Directory
**Symptom:** Service builds but doesn't start correctly  
**Solution:** Set root directory to `cmd/business-intelligence-gateway`

### Issue 2: Wrong Builder Type
**Symptom:** Build fails or service doesn't start  
**Solution:** Set builder type to `Dockerfile`

### Issue 3: Port Mismatch
**Symptom:** Service starts but not accessible  
**Solution:** Verify `PORT` environment variable matches what service is listening on

### Issue 4: Health Check Failing
**Symptom:** Service starts but health checks fail  
**Solution:** Verify health check path is `/health` and service responds

### Issue 5: Service Not Linked
**Symptom:** Service exists but not accessible  
**Solution:** Verify service is properly linked in Railway project

---

## Next Steps After Investigation

1. **If Service Configuration is Wrong:**
   - Fix configuration in Railway dashboard
   - Redeploy service
   - Retest endpoints

2. **If Service is Not Starting:**
   - Check logs for errors
   - Verify Dockerfile is correct
   - Check for missing dependencies

3. **If Service Starts But Not Accessible:**
   - Check network configuration
   - Verify port binding
   - Check Railway proxy settings

4. **If Issue Persists:**
   - Contact Railway support
   - Check Railway status page
   - Review Railway documentation

---

**Last Updated:** November 24, 2025  
**Status:** 📋 **INVESTIGATION GUIDE READY** - Use this to check Railway dashboard

