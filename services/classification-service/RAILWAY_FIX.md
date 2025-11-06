# Railway Build Issue - Classification Service

## Problem
Railway is using Railpack (auto Go builder) instead of the Dockerfile, causing:
- `"using build driver railpack-v0.9.2"` 
- `"Detected Golang"`
- `"no Go files in /app"` error

## Root Cause
Railway detects Go files **before** reading `railway.json` or `railway.toml`, so it auto-selects Railpack instead of using the Dockerfile.

## Solution - Configure Builder in Railway Dashboard

**You MUST set the builder type in Railway Dashboard, not just in files:**

1. Go to Railway Dashboard → Your Project → Classification Service
2. Click **Settings** → **Build & Deploy**
3. Under **Build Settings**, find **Builder** or **Build Command**
4. **Manually set**: `DOCKERFILE` or select "Dockerfile" from dropdown
5. **Dockerfile Path**: `Dockerfile` (relative to root directory)
6. **Root Directory**: Ensure this is set to `services/classification-service`
7. Save and redeploy

## Alternative: Use Railway CLI

```bash
railway service
railway variables set RAILWAY_BUILDER=DOCKERFILE
railway variables set RAILWAY_DOCKERFILE_PATH=Dockerfile
```

## Files Status
✅ `railway.json` - Correctly configured  
✅ `railway.toml` - Correctly configured  
✅ `Dockerfile` - Correctly configured  
❌ Railway Dashboard - **Builder type must be set manually**

## Why This Happens
Railway's auto-detection runs before config file parsing, so detecting Go files triggers Railpack immediately, ignoring the config files that specify Dockerfile.

