# Python ML Service Not Initializing - Fix Guide

**Date**: November 28, 2025  
**Issue**: Python ML service is working, but classification service doesn't initialize it  
**Status**: 🔧 **Environment Variable Not Set or Service Not Redeployed**

---

## Issue Summary

**Python ML Service Status:** ✅ **Working**
- `/ping` endpoint returns: `{"status":"ok","message":"Python ML Service is running"}`
- `/health` endpoint returns: `{"status":"healthy",...}`

**Classification Service Status:** ❌ **Not Initializing Python ML Service**
- No "🐍 Initializing Python ML Service" logs
- No "✅ Python ML Service initialized successfully" logs
- Service is using standard keyword-based classification

---

## Root Cause

The `PYTHON_ML_SERVICE_URL` environment variable is **not set** in Railway for the classification service, OR the classification service **hasn't been redeployed** since the variable was set.

**Evidence:**
- Classification service logs show no Python ML initialization attempts
- Service logs show: "Starting industry detection" (standard classification)
- No errors about Python ML service (because it's not being attempted)

---

## Solution

### Step 1: Verify Python ML Service URL

The Python ML service URL should be:
```
https://python-ml-service-production-a6b8.up.railway.app
```

**Test it:**
```bash
curl https://python-ml-service-production-a6b8.up.railway.app/ping
# Expected: {"status":"ok","message":"Python ML Service is running"}
```

### Step 2: Set Environment Variable in Railway

**Option 1: Using Automated Script (Recommended)**
```bash
./scripts/verify-python-ml-integration.sh
```

This script will:
- Test Python ML service endpoints
- Check if `PYTHON_ML_SERVICE_URL` is set in classification service
- Set it if missing
- Verify the configuration

**Option 2: Using Railway CLI**
```bash
# Link to classification service
railway link --service classification-service

# Set the environment variable
railway variables set PYTHON_ML_SERVICE_URL="https://python-ml-service-production-a6b8.up.railway.app" --service classification-service
```

**Option 3: Railway Dashboard**
1. Go to Railway Dashboard → Your Project
2. Click on `classification-service`
3. Go to "Variables" tab
4. Click "New Variable"
5. Add:
   - **Name**: `PYTHON_ML_SERVICE_URL`
   - **Value**: `https://python-ml-service-production-a6b8.up.railway.app`
6. **Important:** Remove any trailing slash
7. Save

### Step 3: Wait for Service Redeploy

After setting the environment variable:
- Railway will automatically redeploy the classification service
- This takes 1-2 minutes
- The service will restart and read the new environment variable

### Step 4: Verify Initialization

**Check Classification Service Logs:**

After redeploy, you should see:
```
🐍 Initializing Python ML Service
  url: https://python-ml-service-production-a6b8.up.railway.app
🐍 Initializing Python ML Service at https://python-ml-service-production-a6b8.up.railway.app
✅ Python ML Service initialized successfully
✅ Classification services initialized
  python_ml_service: true
```

**If you see:**
```
⚠️ Failed to initialize Python ML Service, continuing without enhanced classification
  error: [error details]
```

Then check:
- Is Python ML service accessible? (test `/ping` endpoint)
- Is the URL correct? (no trailing slash, correct domain)
- Are there network/firewall issues?

---

## Verification Checklist

- [ ] Python ML service is running and healthy (`/ping` and `/health` work)
- [ ] `PYTHON_ML_SERVICE_URL` is set in Railway for classification-service
- [ ] URL has no trailing slash
- [ ] Classification service has been redeployed after setting variable
- [ ] Classification service logs show Python ML initialization
- [ ] `python_ml_service: true` in "Classification services initialized" log

---

## Troubleshooting

### Issue: Environment Variable Not Taking Effect

**Symptoms:**
- Variable is set in Railway
- But service still doesn't initialize Python ML service

**Solution:**
1. **Force redeploy:** Go to Railway Dashboard → classification-service → Deployments → Redeploy
2. **Check variable value:** Verify it's set correctly (no trailing slash)
3. **Check service logs:** Look for "Python ML Service URL not configured" message

### Issue: Python ML Service Initialization Fails

**Symptoms:**
- Service tries to initialize but fails
- Logs show: "⚠️ Failed to initialize Python ML Service"

**Check:**
1. Is Python ML service accessible?
   ```bash
   curl https://python-ml-service-production-a6b8.up.railway.app/ping
   ```

2. Is the URL correct?
   - No trailing slash
   - Correct domain
   - HTTPS (not HTTP)

3. Check Python ML service logs for errors

### Issue: Service Initializes But Enhanced Classification Not Used

**Symptoms:**
- Python ML service initializes successfully
- But requests still use standard classification

**Check:**
1. Does the request include `website_url`?
   - Enhanced classification only works when `website_url` is provided

2. Check logs for:
   ```
   Using Python ML service for enhanced classification
   ```

---

## Expected Behavior After Fix

### Classification Service Logs:
```
🐍 Initializing Python ML Service
  url: https://python-ml-service-production-a6b8.up.railway.app
✅ Python ML Service initialized successfully
✅ Classification services initialized
  python_ml_service: true
```

### When Enhanced Classification is Used:
```
Using Python ML service for enhanced classification
  request_id: req_xxx
  website_url: https://example.com
Python ML service enhanced classification successful
  industry: Technology
  confidence: 0.92
  quantization_enabled: true
```

---

## Quick Fix Command

```bash
# Run the verification script
./scripts/verify-python-ml-integration.sh
```

This will:
1. Test Python ML service
2. Check if variable is set
3. Set it if missing
4. Provide next steps

---

## Summary

**The issue:** `PYTHON_ML_SERVICE_URL` is not set in Railway for classification-service.

**The fix:** Set the environment variable and wait for service redeploy.

**Verification:** Check logs for Python ML service initialization messages.

