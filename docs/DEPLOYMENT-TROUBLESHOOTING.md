# Deployment Troubleshooting Guide

**Issue**: Services returning 404/502 errors after deployment

---

## Quick Diagnosis

### Check Railway Dashboard

1. **Verify Deployment Status**:
   - Go to Railway dashboard
   - Check if services show "Deployed" status
   - Look for any build/deployment errors

2. **Check Service Logs**:
   - Open service logs in Railway
   - Look for startup errors
   - Check for missing environment variables

3. **Verify Service URLs**:
   - Confirm actual Railway URLs
   - Check if URLs match configuration
   - Note any URL changes

---

## Common Issues & Solutions

### Issue 1: 404 "Application not found"

**Possible Causes**:
- Service not deployed
- Wrong URL
- Service crashed on startup

**Solutions**:
1. Check Railway dashboard for deployment status
2. Verify service is actually running
3. Check service logs for errors
4. Confirm URL is correct

### Issue 2: 502 "Application failed to respond"

**Possible Causes**:
- Service starting up
- Service crashed
- Port configuration issue
- Missing dependencies

**Solutions**:
1. Wait 5-10 minutes for service to start
2. Check service logs for crash errors
3. Verify PORT environment variable
4. Check for missing Go dependencies

### Issue 3: Routes Not Working

**Possible Causes**:
- Routes not registered
- API Gateway not proxying correctly
- Path mismatch

**Solutions**:
1. Verify routes are in `cmd/main.go`
2. Check API Gateway proxy configuration
3. Verify path patterns match

---

## Verification Steps

### Step 1: Check Service Health

```bash
# Try root endpoint
curl "https://kyb-api-gateway-production.up.railway.app/"

# Try health endpoint
curl "https://kyb-api-gateway-production.up.railway.app/health"
```

### Step 2: Check Service Logs

In Railway dashboard:
- Open service logs
- Look for "server starting" messages
- Check for error messages
- Verify routes are registered

### Step 3: Verify Environment Variables

Check Railway environment variables:
- `PORT` - Should be set
- `SUPABASE_URL` - Required
- `SUPABASE_ANON_KEY` - Required
- Service URLs - Should match actual Railway URLs

### Step 4: Verify Code Deployment

1. Check if latest commit is deployed
2. Verify all files are present
3. Check build succeeded
4. Verify no build errors

---

## When to Retry Tests

**Wait for**:
- ✅ Services show "Deployed" in Railway
- ✅ Health endpoints return 200
- ✅ No errors in service logs
- ✅ Services have been running for 2+ minutes

**Then retry**:
```bash
./scripts/test-risk-endpoints.sh
```

---

## Getting Help

If issues persist:
1. Check Railway documentation
2. Review service logs thoroughly
3. Verify all environment variables
4. Check for recent Railway status updates

