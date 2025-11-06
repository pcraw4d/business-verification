# Railway Deployment Setup for Merchant Service

## Issue
Railway is auto-detecting Go and using Railpack instead of the Dockerfile, causing "no Go files in /app" errors.

## Solution

**You need to configure the service root directory in Railway Dashboard:**

1. Go to your Railway project
2. Find the **merchant-service** service
3. Go to **Settings** → **Service Settings**
4. Set **Root Directory** to: `services/merchant-service`
5. **IMPORTANT**: Also set **Builder** to: `DOCKERFILE` in the Build & Deploy settings
6. Save and redeploy

This will ensure Railway:
- Reads the `railway.json` from the service directory
- Uses the Dockerfile builder instead of Railpack
- Builds from the correct directory

## Files

- `railway.json` - Railway build configuration (forces Dockerfile)
- `railway.toml` - Alternative configuration format
- `Dockerfile` - Builds from service directory, uses Go 1.22 to match go.mod

## Note

Even with `railway.json` configured, Railway may still auto-detect Go and use Railpack. Setting the builder type explicitly in the Railway dashboard is the most reliable solution.

