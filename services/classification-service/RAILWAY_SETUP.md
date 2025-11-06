# Railway Deployment Setup for Classification Service

## Issue
Railway is auto-detecting Go and using Railpack instead of the Dockerfile, causing "no Go files in /app" errors.

## Solution

**You need to configure the service root directory in Railway Dashboard:**

1. Go to your Railway project
2. Find the **classification-service** service
3. Go to **Settings** → **Service Settings**
4. Set **Root Directory** to: `services/classification-service`
5. Save and redeploy

This will ensure Railway:
- Reads the `railway.json` from the service directory
- Uses the Dockerfile builder instead of Railpack
- Builds from the correct directory

## Alternative: Manual Configuration

If you can't set the root directory, Railway should use the Dockerfile when:
- `railway.json` specifies `"builder": "DOCKERFILE"`
- The Dockerfile path is correct

The current `railway.json` is configured correctly, but Railway may still auto-detect Go if building from repository root.

## Files

- `railway.json` - Railway build configuration (forces Dockerfile)
- `Dockerfile` - Builds from repository root, copies service files correctly

