# Railway Frontend Rebuild Required - API URL Fix

## Critical Issue

The frontend is hardcoded to use `http://localhost:8080` because the `NEXT_PUBLIC_API_BASE_URL` environment variable was not set during the build process.

**Important**: Next.js embeds `NEXT_PUBLIC_*` environment variables at **build time**, not runtime. Simply setting the variable in Railway will NOT fix the issue - you must rebuild the application.

## Root Cause

In `frontend/lib/api.ts` line 26:
```typescript
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
```

When Next.js builds the application, it replaces `process.env.NEXT_PUBLIC_API_BASE_URL` with the actual value at build time. If the variable wasn't set, it defaults to `localhost:8080` and that value is baked into the JavaScript bundle.

## Solution Steps

### Step 1: Set Environment Variable in Railway

1. Go to Railway Dashboard
2. Navigate to your project
3. Select the **frontend-service** (or the service hosting Next.js)
4. Go to **Variables** tab
5. Add/Update these variables:

```bash
NEXT_PUBLIC_API_BASE_URL=https://api-gateway-service-production-21fd.up.railway.app
USE_NEW_UI=true
NEXT_PUBLIC_USE_NEW_UI=true
NODE_ENV=production
PORT=8086
```

### Step 2: Trigger a Rebuild

After setting the environment variables, Railway should automatically trigger a rebuild. If not:

1. Go to the **Deployments** tab in Railway
2. Click **Redeploy** or **Deploy Latest**
3. Wait for the build to complete

### Step 3: Verify the Build

Check the Railway build logs to ensure:
- The environment variable is visible during build
- The build completes successfully
- No errors related to missing environment variables

### Step 4: Test the Fix

1. Visit: `https://frontend-service-production-b225.up.railway.app/add-merchant`
2. Fill out the merchant form
3. Submit the form
4. Check browser console - should see API calls to:
   - ✅ `https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants`
   - ❌ NOT `http://localhost:8080/api/v1/merchants`

## Verification Commands

You can verify the environment variable is set correctly by checking Railway logs:

```bash
# In Railway dashboard, check the build logs for:
echo $NEXT_PUBLIC_API_BASE_URL
# Should output: https://api-gateway-service-production-21fd.up.railway.app
```

## Alternative: Runtime Configuration (Advanced)

If you need to change the API URL without rebuilding, you would need to:

1. Use a runtime configuration approach (not recommended for production)
2. Create a config endpoint that serves the API URL
3. Fetch the config on app initialization

However, the recommended approach is to set the environment variable and rebuild.

## Why This Happens

Next.js optimizes performance by:
1. Embedding `NEXT_PUBLIC_*` variables at build time
2. Replacing them in the JavaScript bundle
3. This means the variable value is static in the built code

This is by design for performance, but requires rebuilds when changing these values.

## Prevention

To prevent this issue in the future:

1. **Always set `NEXT_PUBLIC_*` variables before the first build**
2. **Document required environment variables** in deployment guides
3. **Use Railway's environment variable templates** to ensure consistency
4. **Test builds locally** with the same environment variables

## Quick Reference

**Required Variable:**
```
NEXT_PUBLIC_API_BASE_URL=https://api-gateway-service-production-21fd.up.railway.app
```

**Action Required:**
1. Set variable in Railway ✅
2. Trigger rebuild ✅
3. Wait for deployment ✅
4. Test the application ✅

## Related Files

- `frontend/lib/api.ts` - API client configuration
- `frontend/next.config.ts` - Next.js configuration
- `railway-environment-variables.txt` - Environment variable reference
- `docs/RAILWAY_FRONTEND_DEPLOYMENT.md` - Deployment guide

