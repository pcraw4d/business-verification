# Railway Frontend API Configuration Fix

## Issue Summary

The frontend deployed on Railway is attempting to connect to `http://localhost:8080` instead of the Railway API Gateway service, causing CORS errors and failed API requests.

**Error Message:**
```
Access to fetch at 'http://localhost:8080/api/v1/merchants' from origin 'https://frontend-service-production-b225.up.railway.app' has been blocked by CORS policy
```

## Root Cause

The frontend code uses `process.env.NEXT_PUBLIC_API_BASE_URL` which defaults to `http://localhost:8080` when not set. The environment variable is missing from the Railway frontend service configuration.

## Solution

Add the following environment variables to the **Frontend Service** in Railway:

### Required Environment Variables

1. Go to Railway Dashboard
2. Navigate to your project
3. Select the **frontend-service** (or the service hosting the Next.js frontend)
4. Go to the **Variables** tab
5. Add these variables:

```bash
NEXT_PUBLIC_API_BASE_URL=https://api-gateway-service-production-21fd.up.railway.app
USE_NEW_UI=true
NEXT_PUBLIC_USE_NEW_UI=true
NODE_ENV=production
PORT=8086
```

### Important Notes

- **`NEXT_PUBLIC_API_BASE_URL`** is the critical variable that fixes the issue
- The `NEXT_PUBLIC_` prefix is required for Next.js to expose the variable to the browser
- After adding variables, Railway will automatically redeploy the service
- The API Gateway URL should match your actual Railway API Gateway service URL

## Verification Steps

After setting the environment variables:

1. **Wait for redeployment** - Railway will automatically redeploy when variables change
2. **Check the frontend** - Visit `https://frontend-service-production-b225.up.railway.app`
3. **Test the add merchant flow** - Try adding a merchant and verify API calls succeed
4. **Check browser console** - Should see API calls going to the Railway API Gateway URL, not localhost

## Expected Behavior After Fix

- API calls should go to: `https://api-gateway-service-production-21fd.up.railway.app/api/v1/merchants`
- No more CORS errors in the browser console
- Merchant form submissions should work correctly
- All API endpoints should be accessible

## Related Files

The following files reference the API base URL configuration:

- `frontend/lib/api.ts` - Main API client (line 26)
- `frontend/next.config.ts` - Next.js rewrite configuration (line 11)
- `frontend/app/layout.tsx` - DNS prefetch configuration (lines 37-38)
- `frontend/lib/websocket.ts` - WebSocket connection (line 213)
- `frontend/lib/preload.ts` - Performance optimizations (line 71)

## Additional Configuration

If you need to update the API Gateway URL in the future:

1. Update the `NEXT_PUBLIC_API_BASE_URL` variable in Railway
2. The service will automatically redeploy
3. All API calls will use the new URL

## Troubleshooting

### Still seeing localhost errors?

1. **Clear browser cache** - Hard refresh (Ctrl+Shift+R or Cmd+Shift+R)
2. **Check Railway logs** - Verify the environment variable is set correctly
3. **Verify build** - Check that Next.js build includes the environment variable
4. **Check service URL** - Ensure the API Gateway URL is correct

### API Gateway not responding?

1. Verify the API Gateway service is running in Railway
2. Check the API Gateway health endpoint: `https://api-gateway-service-production-21fd.up.railway.app/health`
3. Verify CORS is configured on the API Gateway to allow the frontend origin

### Environment variable not taking effect?

1. **Next.js requires rebuild** - Environment variables are baked into the build
2. **Check variable name** - Must start with `NEXT_PUBLIC_` for browser access
3. **Verify deployment** - Check Railway deployment logs for build output

## Quick Reference

**Frontend Service URL:**
```
https://frontend-service-production-b225.up.railway.app
```

**API Gateway URL:**
```
https://api-gateway-service-production-21fd.up.railway.app
```

**Required Variable:**
```
NEXT_PUBLIC_API_BASE_URL=https://api-gateway-service-production-21fd.up.railway.app
```

