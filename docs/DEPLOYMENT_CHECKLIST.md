# Frontend Deployment Checklist

**Service**: Frontend (Next.js)  
**Platform**: Railway  
**Last Updated**: 2025-01-XX

## Pre-Deployment Checklist

### Code Quality ✅
- [ ] All TypeScript compilation passes (`npm run build`)
- [ ] All linting passes (`npm run lint`)
- [ ] All unit tests pass (`npm test`)
- [ ] All E2E tests pass (`npm run test:e2e`)
- [ ] No console errors in development mode
- [ ] Build verification script passes (`npm run verify-env`)

### Environment Variables ✅
- [ ] `NEXT_PUBLIC_API_BASE_URL` is set
  - **Value**: `https://api-gateway-service-production-21fd.up.railway.app`
  - **Critical**: Must be set BEFORE building
- [ ] `NODE_ENV` is set to `production`
- [ ] `USE_NEW_UI` is set (if applicable)
- [ ] `NEXT_PUBLIC_USE_NEW_UI` is set (if applicable)

### Testing ✅
- [ ] Critical paths tested manually
- [ ] All pages load without errors
- [ ] API calls verified (not localhost)
- [ ] Form submissions work
- [ ] Navigation works correctly
- [ ] No 404 errors on RSC requests

## Railway Deployment Steps

### Step 1: Configure Environment Variables

1. Go to Railway Dashboard
2. Select the **frontend service**
3. Navigate to **Variables** tab
4. Add/Update the following variables:

```
NEXT_PUBLIC_API_BASE_URL=https://api-gateway-service-production-21fd.up.railway.app
NODE_ENV=production
USE_NEW_UI=true
NEXT_PUBLIC_USE_NEW_UI=true
```

**⚠️ CRITICAL**: Set these variables BEFORE triggering a rebuild!

### Step 2: Trigger Rebuild

**Option A: Automatic Rebuild**
- Railway automatically rebuilds when environment variables change
- Monitor the build logs

**Option B: Manual Rebuild**
1. Go to Railway Dashboard
2. Select the frontend service
3. Click **Deploy** or **Redeploy**
4. Monitor build logs

### Step 3: Verify Build

**Check Build Logs For**:
- ✅ Build verification script output
- ✅ Environment variable detection
- ✅ Next.js build completion
- ✅ No build errors

**Expected Build Log Output**:
```
> frontend@0.1.0 verify-env
> node scripts/verify-build-env.js

🔍 Checking required environment variables...

✅ NEXT_PUBLIC_API_BASE_URL is set: https://api-gateway-service-production-21fd.up.railway.app

✅ All required checks passed!
```

### Step 4: Post-Deployment Verification

#### Immediate Checks (Within 5 minutes)

- [ ] Frontend service is accessible
  - URL: `https://frontend-service-production-b225.up.railway.app`
- [ ] Home page loads
- [ ] No console errors in browser
- [ ] API calls go to Railway (check Network tab)
- [ ] No CORS errors

#### Automated Testing (Within 15 minutes)

```bash
npm run test:pages -- \
  --base-url https://frontend-service-production-b225.up.railway.app \
  --api-url https://api-gateway-service-production-21fd.up.railway.app
```

**Expected**: All pages return 200 status

#### Manual Testing (Within 30 minutes)

**Critical Paths**:
- [ ] Add Merchant flow works
- [ ] Merchant Portfolio loads
- [ ] Merchant Details page loads
- [ ] Risk Dashboard loads
- [ ] Compliance pages load (no 404s)

**Browser Console Checks**:
- [ ] No errors
- [ ] No warnings about localhost
- [ ] All API calls show Railway URLs
- [ ] No CORS errors

#### Network Tab Verification

1. Open browser DevTools
2. Go to Network tab
3. Navigate through the application
4. Verify:
   - ✅ All API calls go to `api-gateway-service-production-21fd.up.railway.app`
   - ✅ No calls to `localhost:8080`
   - ✅ All requests return appropriate status codes
   - ✅ No CORS errors

## Post-Deployment Monitoring

### First Hour
- [ ] Monitor error logs
- [ ] Check API error rates
- [ ] Verify no localhost API calls (check logs)
- [ ] Monitor user reports

### First 24 Hours
- [ ] Daily automated test run
- [ ] Check error logs daily
- [ ] Monitor API error rates
- [ ] Verify performance metrics
- [ ] Check for any user-reported issues

### Ongoing
- [ ] Weekly full testing cycle
- [ ] Monthly performance review
- [ ] Quarterly security audit

## Rollback Procedure

If deployment fails or issues are discovered:

1. **Immediate Rollback**:
   - Go to Railway Dashboard
   - Select frontend service
   - Go to **Deployments** tab
   - Click **Redeploy** on previous working deployment

2. **Investigate Issues**:
   - Check build logs
   - Check error logs
   - Review environment variables
   - Test locally with production env vars

3. **Fix and Redeploy**:
   - Fix identified issues
   - Re-run pre-deployment checklist
   - Redeploy following deployment steps

## Troubleshooting

### Build Fails

**Check**:
- Environment variables are set
- Build verification script output
- TypeScript compilation errors
- Next.js build errors

**Common Issues**:
- Missing `NEXT_PUBLIC_API_BASE_URL`
- Invalid URL format
- TypeScript errors
- Missing dependencies

### Pages Return 404

**Check**:
- Next.js routing configuration
- Page file structure
- Metadata exports on server components
- Build output includes pages

### API Calls Go to Localhost

**Check**:
- `NEXT_PUBLIC_API_BASE_URL` is set correctly
- Service was rebuilt AFTER setting env var
- Build logs show env var was detected
- Browser cache cleared

**Solution**:
1. Verify env var in Railway
2. Trigger rebuild
3. Clear browser cache
4. Test in incognito mode

### CORS Errors

**Check**:
- API Gateway CORS configuration
- Frontend origin matches allowed origins
- CORS headers in API responses

**Solution**:
1. Verify `CORS_ALLOWED_ORIGINS` in API Gateway
2. Check API Gateway logs
3. Verify frontend URL is in allowed list

## Success Criteria

### Critical (Must Pass)
- ✅ Zero 404 errors on page loads
- ✅ Zero localhost API calls in production
- ✅ Zero CORS errors
- ✅ All forms submit successfully
- ✅ All critical pages load

### High Priority (Should Pass)
- ✅ All pages return 200 status
- ✅ API endpoints respond in <500ms
- ✅ No console errors
- ✅ All data loads correctly
- ✅ Enhanced features work (caching, deduplication)

### Quality Metrics
- ✅ Lighthouse scores maintain or improve
- ✅ Error rate <0.1%
- ✅ Bundle size not increased >5%
- ✅ Build completes successfully

## Deployment Sign-Off

**Deployed By**: _______________  
**Date**: _______________  
**Time**: _______________  
**Version**: _______________  

**Verification**:
- [ ] All pre-deployment checks passed
- [ ] Environment variables configured
- [ ] Build completed successfully
- [ ] Post-deployment verification passed
- [ ] No critical issues found

**Notes**:
_______________________________________________
_______________________________________________
_______________________________________________

---

**Next Deployment**: _______________  
**Last Reviewed**: _______________
