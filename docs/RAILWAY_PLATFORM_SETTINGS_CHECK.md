# Railway Platform Settings Check - 502 Error Investigation

**Date**: December 22, 2025  
**Status**: ✅ **TIMEOUT CONFIG VERIFIED** | 🔍 **PLATFORM SETTINGS CHECK REQUIRED**  
**Priority**: HIGH

---

## Executive Summary

All Railway API Gateway timeout variables are correctly set to 120s. However, **Railway platform-level settings** may have additional timeouts or configurations that could cause 502 errors. This document provides a checklist for verifying and configuring Railway platform settings.

---

## Railway Platform Settings to Check

### 1. Service Health Check Configuration ⚠️ **CRITICAL**

**Location**: Railway Dashboard → Service → Settings → Health Check

**Settings to Verify**:
- **Health Check Path**: Should be `/health` (or `/health?detailed=true` for comprehensive checks)
- **Health Check Interval**: Should be ≥30s (default is usually 30s)
- **Health Check Timeout**: Should be ≥10s (default is usually 5s)
- **Start Period**: Should be ≥60s for cold start (default is usually 30s)
- **Retries**: Should be ≥3 (default is usually 3)

**Recommended Configuration**:
```yaml
health_check:
  path: /health
  interval: 30s
  timeout: 10s
  start_period: 60s  # Allow 60s for cold start
  retries: 3
```

**How to Check**:
1. Go to Railway Dashboard
2. Select `classification-service` (or `api-gateway-service`)
3. Go to **Settings** tab
4. Scroll to **Health Check** section
5. Verify all settings match recommendations above

**Expected Impact**: Prevents premature health check failures during cold start

---

### 2. Service "Always On" Setting ⚠️ **HIGH PRIORITY**

**Location**: Railway Dashboard → Service → Settings → Always On

**What It Does**:
- Prevents service from going to sleep after inactivity
- Reduces cold start frequency
- May incur additional costs (check Railway pricing)

**Recommended Setting**:
- ✅ **Enable "Always On"** for production services
- ⚠️ **Consider cost implications** (may increase monthly bill)

**How to Check/Enable**:
1. Go to Railway Dashboard
2. Select `classification-service`
3. Go to **Settings** tab
4. Look for **"Always On"** or **"Keep Alive"** toggle
5. Enable if available and cost is acceptable

**Expected Impact**: Reduces cold start frequency by 80-90%

---

### 3. Service Resource Limits ⚠️ **MEDIUM PRIORITY**

**Location**: Railway Dashboard → Service → Settings → Resources

**Settings to Verify**:
- **Memory Limit**: Should be ≥512MB (recommended: 1GB for classification service)
- **CPU Limit**: Should be ≥0.5 vCPU (recommended: 1 vCPU for better performance)
- **Disk Space**: Should be ≥1GB (default is usually sufficient)

**How to Check**:
1. Go to Railway Dashboard
2. Select `classification-service`
3. Go to **Settings** tab
4. Scroll to **Resources** section
5. Verify memory and CPU limits

**Expected Impact**: Prevents OOM kills and improves cold start performance

---

### 4. Service Scaling Configuration ⚠️ **MEDIUM PRIORITY**

**Location**: Railway Dashboard → Service → Settings → Scaling

**Settings to Verify**:
- **Min Instances**: Should be ≥1 (keeps at least one instance warm)
- **Max Instances**: Should be ≥2 (allows horizontal scaling)
- **Auto-scaling**: Should be enabled for production

**How to Check**:
1. Go to Railway Dashboard
2. Select `classification-service`
3. Go to **Settings** tab
4. Scroll to **Scaling** section
5. Verify min/max instances

**Expected Impact**: Reduces cold start frequency by keeping instances warm

---

### 5. Railway Platform-Level Timeout ⚠️ **CRITICAL**

**Location**: Railway Dashboard → Project → Settings → Platform Settings

**What to Check**:
- Railway may have a **platform-level timeout** (separate from service timeouts)
- This could be a **request timeout** or **connection timeout**
- Default is usually **120s** or **300s**, but may vary

**How to Check**:
1. Go to Railway Dashboard
2. Select your **Project**
3. Go to **Settings** tab
4. Look for **"Platform Settings"** or **"Request Timeout"**
5. Verify timeout is ≥120s (recommended: 300s for long-running requests)

**Expected Impact**: Prevents Railway platform from timing out before service responds

---

### 6. Railway API Gateway Timeout (Already Verified) ✅

**Status**: ✅ **VERIFIED** - All timeout variables set to 120s:
- `READ_TIMEOUT`: 120s
- `WRITE_TIMEOUT`: 120s
- `HTTP_CLIENT_TIMEOUT`: 120s

**Location**: Railway Dashboard → `api-gateway-service` → Variables

---

## Verification Checklist

Use this checklist to verify all settings:

- [ ] **Health Check Path**: `/health` configured
- [ ] **Health Check Timeout**: ≥10s
- [ ] **Health Check Start Period**: ≥60s
- [ ] **Always On**: Enabled (if available and cost acceptable)
- [ ] **Memory Limit**: ≥512MB (recommended: 1GB)
- [ ] **CPU Limit**: ≥0.5 vCPU (recommended: 1 vCPU)
- [ ] **Min Instances**: ≥1
- [ ] **Max Instances**: ≥2
- [ ] **Platform Timeout**: ≥120s (recommended: 300s)
- [ ] **API Gateway Timeouts**: All set to 120s ✅

---

## How to Access Railway Dashboard

### Option 1: Web Dashboard
1. Go to https://railway.app
2. Log in with your account
3. Select your project
4. Navigate to service settings

### Option 2: Railway CLI
```bash
# Install Railway CLI (if not installed)
npm i -g @railway/cli

# Login
railway login

# List services
railway status

# View service settings
railway service

# View variables
railway variables
```

---

## Expected Impact After Configuration

### Before (Current State)
- **Cold Start Frequency**: High (service sleeps after inactivity)
- **Cold Start Duration**: 30-40s
- **502 Error Rate**: 4% (2 out of 50 requests)
- **First Request Latency**: 30-95s (cold start + processing)

### After (With Optimizations)
- **Cold Start Frequency**: Low (Always On enabled)
- **Cold Start Duration**: 10-15s (optimized initialization)
- **502 Error Rate**: <1% (retry logic handles remaining cases)
- **First Request Latency**: 1-5s (warm service)

---

## Cost Considerations

### "Always On" Cost Impact
- **Free Tier**: May not be available
- **Pro Tier**: Usually included or minimal cost
- **Enterprise**: Usually included

**Recommendation**: Enable "Always On" for production services if cost is acceptable (<$10/month per service).

---

## Next Steps

1. **Immediate** (This Week):
   - [ ] Verify health check configuration
   - [ ] Check platform-level timeout
   - [ ] Enable "Always On" if available and cost acceptable

2. **Short Term** (Next Week):
   - [ ] Optimize resource limits (memory/CPU)
   - [ ] Configure service scaling (min instances)
   - [ ] Monitor cold start frequency

3. **Long Term** (Next Month):
   - [ ] Review cost impact of "Always On"
   - [ ] Optimize cold start further (if needed)
   - [ ] Consider Railway Pro/Enterprise for better features

---

## Related Documents

- `docs/502_ERROR_INVESTIGATION_20251222.md` - Root cause analysis
- `docs/502_ERROR_ROOT_CAUSE_UPDATE.md` - Updated root cause analysis
- `docs/API_GATEWAY_TIMEOUT_CONFIGURATION_CHECK.md` - API Gateway timeout verification

---

**Last Updated**: December 22, 2025  
**Next Review**: December 29, 2025

