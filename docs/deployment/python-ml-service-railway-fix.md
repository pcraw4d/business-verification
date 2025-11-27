# Python ML Service Railway Build Fix

**Issue**: Railway is auto-detecting the wrong Dockerfile (Go service) instead of using the Python ML service Dockerfile.

**Error**:

```
ERROR: failed to build: failed to solve: failed to compute cache key: failed to calculate checksum of ref 8qf4slnal40gxtfuku357sdv3::jku0s9w8v8o9uac5qn84veerw: "/go.sum": not found
```

## Root Cause

Railway is auto-detecting a Dockerfile from the repository root instead of using the `railway.json` configuration. This happens when:

1. Railway service root directory is not set to `python_ml_service/`
2. Railway auto-detects a Dockerfile before reading `railway.json`
3. The build context is incorrect

## Solution

### Option 1: Set Service Root Directory in Railway Dashboard (Recommended)

1. Go to Railway Dashboard
2. Click on `python-ml-service`
3. Go to **Settings** → **Service Settings**
4. Set **Root Directory** to: `python_ml_service`
5. Set **Dockerfile Path** to: `Dockerfile` (relative to root directory)
6. Redeploy

This ensures Railway:

- Builds from the `python_ml_service/` directory
- Uses the correct Dockerfile
- Has the correct build context

### Option 2: Update Railway Service Configuration

If Option 1 doesn't work, configure the service manually:

1. In Railway Dashboard → `python-ml-service` → Settings
2. **Build Command**: Leave empty (uses Dockerfile)
3. **Root Directory**: `python_ml_service`
4. **Dockerfile Path**: `Dockerfile`
5. **Build Context**: `.` (current directory)

### Option 3: Use Railway CLI (Automated)

Run the configuration script:

```bash
./scripts/configure-python-ml-railway.sh
```

Or manually:

```bash
# Link to the service
cd python_ml_service
railway link --service python-ml-service

# Note: Root directory must be set in Railway dashboard
# Railway CLI doesn't support setting root directory directly
# After linking, go to dashboard and set:
# - Root Directory: python_ml_service
# - Dockerfile Path: Dockerfile

# Deploy
railway up
```

## Verification

After fixing, the build should:

1. ✅ Use Python base image (`python:3.11-slim`)
2. ✅ Install Python dependencies from `requirements.txt`
3. ✅ Build the production stage
4. ✅ Not look for `go.mod` or `go.sum`

## Current Configuration

The `python_ml_service/railway.json` is correctly configured:

```json
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "python_ml_service/Dockerfile",
    "dockerContext": "python_ml_service"
  }
}
```

However, Railway may ignore this if the service root directory isn't set correctly in the dashboard.

## Next Steps

1. ✅ Set service root directory in Railway dashboard
2. ✅ Verify Dockerfile path
3. ✅ Trigger a new deployment
4. ✅ Check build logs for Python image usage

---

**Status**: ⚠️ **Requires Railway Dashboard Configuration**
