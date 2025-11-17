# Railway New UI Deployment Status

**Date**: 2025-01-17  
**Frontend Service URL**: `https://frontend-service-production-b225.up.railway.app`

## Current Status

### ✅ Code Deployed
- Next.js frontend code is in the repository
- All pages migrated to shadcn UI
- Build configuration ready

### ⚠️ Feature Flag Status: UNKNOWN
The new UI requires environment variables to be set in Railway:
- `USE_NEW_UI=true`
- `NEXT_PUBLIC_USE_NEW_UI=true`

**Action Required**: Check Railway dashboard to verify if these are set.

## How to Enable New UI

### Step 1: Access Railway Dashboard
1. Go to https://railway.app
2. Navigate to your project
3. Select the **frontend-service** service

### Step 2: Set Environment Variables
In the "Variables" tab, add:

```bash
USE_NEW_UI=true
NEXT_PUBLIC_USE_NEW_UI=true
NEXT_PUBLIC_API_BASE_URL=https://api-gateway-service-production-21fd.up.railway.app
NODE_ENV=production
```

### Step 3: Redeploy
- Railway will automatically redeploy when variables change
- Or manually trigger a redeploy

### Step 4: Verify
Visit: `https://frontend-service-production-b225.up.railway.app`

**Expected**: Should see shadcn UI components (new UI)  
**If legacy UI**: Environment variables not set correctly

## Routing Logic

The Go frontend service (`cmd/frontend-service/routing.go`) checks:
```go
useNewUI := os.Getenv("NEXT_PUBLIC_USE_NEW_UI") == "true" || os.Getenv("USE_NEW_UI") == "true"
```

- If `true`: Routes to Next.js application (new UI)
- If `false` or unset: Routes to legacy HTML files (old UI)

## Verification Checklist

- [ ] Check Railway dashboard for `USE_NEW_UI` variable
- [ ] Check Railway dashboard for `NEXT_PUBLIC_USE_NEW_UI` variable
- [ ] Verify Next.js build exists in deployment
- [ ] Test frontend service URL
- [ ] Verify shadcn UI components render
- [ ] Check API integration works

## Troubleshooting

### Legacy UI Still Showing
- Verify environment variables are set to `"true"` (not `true` without quotes)
- Check service logs for routing decisions
- Ensure Next.js build completed successfully

### Next.js Build Missing
- Check if build step runs in Railway
- Verify `frontend/.next/` directory exists in deployment
- Check build logs for errors

### API Calls Failing
- Verify `NEXT_PUBLIC_API_BASE_URL` is set correctly
- Check API gateway is accessible
- Review network requests in browser console

