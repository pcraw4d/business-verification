# Frontend Service 502 Error Investigation

**Date**: 2025-11-19  
**Status**: 🔴 **INVESTIGATING**  
**Issue**: 502 Bad Gateway error when accessing frontend service

---

## Problem

**Error**: 502 Bad Gateway  
**URL**: `https://frontend-service-production-b225.up.railway.app`  
**Health Check**: Also returns 502

---

## Investigation

### Service Status

**Logs Show**:
- Service starts successfully
- Listening on port 8080
- Health endpoint registered at `/health`
- Service reports as "ready and listening"

**But**:
- Health check returns 502
- Main URL returns 502
- Railway proxy cannot reach the service

---

## Root Cause Analysis

### Potential Issues

1. **Port Mismatch**
   - Service code reads `PORT` env var (defaults to 8086)
   - Logs show service listening on 8080
   - Railway may be expecting different port

2. **Health Check Timeout**
   - Health check configured at `/health`
   - Service may be slow to respond
   - Railway timeout may be too short

3. **Service Crash After Startup**
   - Service starts but crashes
   - No error logs visible
   - Railway proxy can't connect

4. **Network/Proxy Issue**
   - Railway proxy configuration issue
   - Service not accessible from proxy
   - Port binding issue

---

## Root Cause Identified

Based on [Railway's documentation](https://docs.railway.com/reference/errors/application-failed-to-respond), the most common cause of 502 errors is:

### **Target Port Mismatch** ⚠️ **LIKELY ISSUE**

The Railway domain's **target port** must match the port the application is listening on.

**Current Situation**:
- Service logs show listening on port **8080**
- Documentation indicates frontend service should use **PORT=8086**
- Railway may have automatically set PORT=8080, or the domain's target port is set incorrectly

**Railway Documentation States**:
> "If your domain is using a target port, ensure that the target port for your public domain matches the port your application is listening on."

### Application Binding ✅ **CORRECT**

The service correctly:
- Binds to `0.0.0.0` (Go's `ListenAndServe(":port", nil)` binds to all interfaces by default)
- Uses the `PORT` environment variable (reads from `os.Getenv("PORT")`, defaults to 8086)

---

## Solutions

### Solution 1: Check Domain Target Port (Most Likely Fix)

1. Go to Railway Dashboard
2. Navigate to the frontend service
3. Go to **Settings** → **Networking** or **Domains**
4. Check the **target port** for `frontend-service-production-b225.up.railway.app`
5. **Ensure it matches the port the service is listening on** (currently 8080 based on logs)

**If target port is wrong**:
- Update it to match the PORT env var (should be 8086 per documentation, or 8080 if Railway set it)
- Or update PORT env var to match the target port

### Solution 2: Verify PORT Environment Variable

**Current Status**:
- Service logs confirm listening on port **8080**
- Railway automatically sets PORT (may not show in variables list)
- Documentation suggests PORT=8086, but Railway is using 8080

**Options**:

**Option A: Use PORT=8080 (Match Current Railway Setting)**
```bash
# Explicitly set PORT to match what Railway is using
railway variables --service frontend-service --set PORT=8080
```

**Option B: Use PORT=8086 (Match Documentation)**
```bash
# Set PORT to match documentation
railway variables --service frontend-service --set PORT=8086
# Then ensure domain target port is also 8086
```

**Important**: Whichever port you choose, **the domain target port must match**.

### Solution 3: Ensure Service Binds to 0.0.0.0 ✅

The code already does this correctly:
```go
http.ListenAndServe(":"+service.port, nil)  // Binds to 0.0.0.0 by default
```

No code changes needed.

---

## Recommended Fix Steps

### Step 1: Check Domain Target Port (Railway Dashboard)

1. Go to [Railway Dashboard](https://railway.app)
2. Navigate to your project
3. Click on the **frontend-service** service
4. Go to **Settings** tab
5. Scroll to **Networking** or **Domains** section
6. Find the domain: `frontend-service-production-b225.up.railway.app`
7. Check the **Target Port** setting

**Expected**: Target port should be **8080** (matching current service port)

**If target port is different or not set**:
- Update it to **8080** to match the service
- Or update PORT env var to match the target port

### Step 2: Verify PORT Environment Variable

Since Railway automatically sets PORT, you may need to explicitly set it:

```bash
# Option 1: Match current Railway setting (8080)
railway variables --service frontend-service --set PORT=8080

# Option 2: Use documentation standard (8086)
railway variables --service frontend-service --set PORT=8086
```

**Important**: After setting PORT, ensure the domain target port matches!

### Step 3: Redeploy

After fixing the port mismatch:
1. Railway will automatically redeploy when variables change
2. Or manually trigger a redeploy from the dashboard
3. Wait for deployment to complete
4. Test the service: `https://frontend-service-production-b225.up.railway.app`

---

## Verification Checklist

- [ ] Domain target port matches service listening port
- [ ] PORT environment variable is explicitly set
- [ ] Service redeployed after port changes
- [ ] Health check endpoint responds: `/health`
- [ ] Main URL responds without 502 error

---

---

## Resolution ✅

**Status**: ✅ **RESOLVED**  
**Resolution Date**: 2025-11-19  
**Fix Applied**: Set `PORT=8080` environment variable to match Railway's automatic port assignment

### Fix Applied

```bash
railway variables --service frontend-service --set PORT=8080
```

**Result**: 
- Service now explicitly uses PORT=8080
- Matches Railway's automatic port assignment
- Domain target port aligns with service listening port
- 502 error resolved after redeployment

### Verification

- ✅ PORT environment variable set to 8080
- ✅ Service listening on port 8080 (confirmed in logs)
- ✅ Domain target port matches service port
- ✅ 502 error resolved
- ✅ Service accessible at `https://frontend-service-production-b225.up.railway.app`

---

**Last Updated**: 2025-11-19  
**Status**: ✅ **RESOLVED**  
**Reference**: [Railway Error Documentation](https://docs.railway.com/reference/errors/application-failed-to-respond)


