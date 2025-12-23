# API Gateway Timeout Configuration Check

**Date**: December 22, 2025  
**Status**: ✅ **CODE ANALYSIS COMPLETE** | ⚠️ **RAILWAY VERIFICATION REQUIRED**

---

## Executive Summary

The API Gateway service code is **correctly configured** with 120-second timeouts by default. However, **Railway environment variables may be overriding** these defaults with lower values, causing the 502 errors.

---

## Code Configuration Analysis

### ✅ Code Defaults (Correct)

From `services/api-gateway/internal/config/config.go` (lines 72-75):

```go
ReadTimeout:       getEnvAsDuration("READ_TIMEOUT", 120*time.Second),      // Default: 120s ✅
WriteTimeout:      getEnvAsDuration("WRITE_TIMEOUT", 120*time.Second),     // Default: 120s ✅
HTTPClientTimeout: getEnvAsDuration("HTTP_CLIENT_TIMEOUT", 120*time.Second), // Default: 120s ✅
```

**Status**: ✅ **Code defaults are correct** (120s matches classification service timeout)

### ⚠️ Environment Variable Override Risk

The code uses `getEnvAsDuration()` which **reads from Railway environment variables**. If Railway has these variables set to lower values (e.g., 30s), they will override the code defaults.

**Environment Variables to Check**:
- `READ_TIMEOUT` (should be ≥120s)
- `WRITE_TIMEOUT` (should be ≥120s)
- `HTTP_CLIENT_TIMEOUT` (should be ≥120s)

---

## How to Check Railway Configuration

### Method 1: Railway CLI (Requires Authentication)

```bash
# Login to Railway
railway login

# Check API Gateway service variables
railway variables --service api-gateway-service --json | \
  jq -r '.["READ_TIMEOUT"], .["WRITE_TIMEOUT"], .["HTTP_CLIENT_TIMEOUT"]'

# Expected output (if correctly configured):
# 120s
# 120s
# 120s

# If output shows 30s or other values, they need to be updated
```

### Method 2: Railway Dashboard

1. Go to Railway Dashboard
2. Select your project
3. Click on **"api-gateway-service"** (or the API Gateway service name)
4. Go to **"Variables"** tab
5. Check for these variables:
   - `READ_TIMEOUT`
   - `WRITE_TIMEOUT`
   - `HTTP_CLIENT_TIMEOUT`

### Method 3: Check Service Logs

The API Gateway logs its timeout configuration on startup. Check Railway logs for:

```
🔧 Configuration loaded
  read_timeout: 120s (or 30s if misconfigured)
  write_timeout: 120s (or 30s if misconfigured)
  http_client_timeout: 120s (or 30s if misconfigured)
```

**Location**: Railway Dashboard → api-gateway-service → Logs → Look for startup messages

---

## Historical Context

### Previous Issue (December 9, 2025)

According to `docs/railway-routing-issue-resolved.md`:

**Before Fix**:
```json
"READ_TIMEOUT": "30s"        ❌ Too short
"WRITE_TIMEOUT": "30s"       ❌ Too short  
"HTTP_CLIENT_TIMEOUT": null  ⚠️  Using default (120s)
```

**After Fix**:
```json
"READ_TIMEOUT": "120s"       ✅ Fixed
"WRITE_TIMEOUT": "120s"      ✅ Fixed
"HTTP_CLIENT_TIMEOUT": "120s" ✅ Explicitly set
```

**Status**: Fix was applied, but **may have been reverted** or **not persisted** after service restarts.

---

## Current 502 Error Analysis

### Amazon Failure (32.9 seconds)
- **Close to 30s timeout**: Suggests `READ_TIMEOUT=30s` may still be set
- **Root Cause**: API Gateway closes connection after 30s, returns 502

### Tesla Failure (95.7 seconds)
- **Close to 90s**: May indicate a different timeout or processing limit
- **Root Cause**: Service processing exceeded some timeout threshold

---

## Recommended Actions

### ✅ Immediate: Verify Railway Configuration

**Step 1**: Check current Railway environment variables

```bash
# If Railway CLI is authenticated
railway variables --service api-gateway-service --json | \
  jq '.["READ_TIMEOUT"], .["WRITE_TIMEOUT"], .["HTTP_CLIENT_TIMEOUT"]'

# Or check via Railway Dashboard
# Railway Dashboard → api-gateway-service → Variables
```

**Step 2**: Check API Gateway startup logs

Look for these log messages in Railway:
```
🔧 Configuration loaded
  read_timeout: <value>
  write_timeout: <value>
  http_client_timeout: <value>
```

### ✅ If Timeouts Are Incorrect: Update Railway Variables

```bash
# Update to 120s (matching code defaults)
railway variables --set "READ_TIMEOUT=120s" \
  --set "WRITE_TIMEOUT=120s" \
  --set "HTTP_CLIENT_TIMEOUT=120s" \
  --service api-gateway-service

# Restart API Gateway service to apply changes
railway restart --service api-gateway-service
```

### ✅ Verify Fix

**Step 1**: Check logs after restart
- Verify timeouts are now 120s in startup logs

**Step 2**: Test Amazon and Tesla URLs
```bash
# Test Amazon
curl -X POST https://api-gateway-production.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name":"Amazon","website_url":"https://amazon.com"}' \
  -w "\nTime: %{time_total}s\n"

# Test Tesla
curl -X POST https://api-gateway-production.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name":"Tesla Inc","website_url":"https://tesla.com"}' \
  -w "\nTime: %{time_total}s\n"
```

**Step 3**: Re-run E2E test
```bash
python3 test/scripts/run_e2e_metrics.py
```

**Success Criteria**:
- No 502 errors for Amazon/Tesla
- Error rate <2% (down from 4%)

---

## Configuration Matrix

| Variable | Code Default | Railway (Expected) | Railway (If Wrong) | Impact |
|----------|--------------|-------------------|-------------------|--------|
| `READ_TIMEOUT` | 120s | 120s | 30s | ❌ 502 errors at 30s |
| `WRITE_TIMEOUT` | 120s | 120s | 30s | ❌ 502 errors at 30s |
| `HTTP_CLIENT_TIMEOUT` | 120s | 120s | 30s | ❌ Backend requests timeout |

---

## Code References

### Configuration Loading
- **File**: `services/api-gateway/internal/config/config.go`
- **Lines**: 72-75
- **Function**: `Load()`

### Server Configuration
- **File**: `services/api-gateway/cmd/main.go`
- **Lines**: 219-221
- **Function**: HTTP server initialization

### HTTP Client Configuration
- **File**: `services/api-gateway/internal/handlers/gateway.go`
- **Lines**: 75-85
- **Function**: HTTP client creation with timeout

---

## Conclusion

**Code Status**: ✅ **Correctly configured** (120s defaults)

**Railway Status**: ⚠️ **Needs verification** (may have 30s overrides)

**Action Required**: 
1. Verify Railway environment variables
2. Update to 120s if incorrect
3. Restart API Gateway service
4. Test and validate fix

**Expected Outcome**: 502 error rate reduction from 4% to <1%

---

**Next Step**: Check Railway Dashboard or use Railway CLI to verify current timeout values.

