# API Gateway Deployment Instructions

**Date**: 2025-01-27  
**Status**: Code ready, deployment pending

---

## Current Status

✅ **Code Changes**: Committed and pushed to GitHub  
⏳ **Deployment**: Pending (Railway auto-deploy or manual trigger)

---

## Deployment Options

### Option 1: Wait for Auto-Deployment (Recommended)

If Railway is connected to GitHub:
1. Code is already pushed to `main` branch
2. Railway should auto-deploy within 5-10 minutes
3. Monitor deployment in Railway dashboard

**Verification**:
```bash
./scripts/verify-deployment.sh
```

### Option 2: Manual Railway CLI Deployment

If auto-deployment is not enabled:

1. **Login to Railway**:
   ```bash
   railway login
   ```

2. **Link to Project**:
   ```bash
   cd services/api-gateway
   railway link
   # Select the API Gateway service project
   ```

3. **Deploy**:
   ```bash
   railway up --detach
   ```

4. **Monitor Deployment**:
   ```bash
   railway logs --service api-gateway-service
   ```

5. **Verify**:
   ```bash
   ./scripts/verify-deployment.sh
   ```

### Option 3: Railway Dashboard Deployment

1. Go to [Railway Dashboard](https://railway.app)
2. Select your project
3. Find "API Gateway Service"
4. Click "Deploy" or "Redeploy"
5. Wait for deployment to complete
6. Verify with: `./scripts/verify-deployment.sh`

---

## Verification

### Quick Verification

```bash
# Test validation fix (should return 400)
curl -X POST https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"description":"Test"}' \
  -w "\nHTTP Status: %{http_code}\n"
```

**Expected Result**: HTTP Status: 400

### Full Verification

```bash
./scripts/verify-deployment.sh
```

This script tests:
1. Health check (should return 200)
2. Missing field validation (should return 400)
3. Valid request (should return 200)

---

## What Changed

### Code Changes
- Added validation for `business_name` field
- Returns 400 instead of 503 for missing required fields
- Added missing `context` import

### Expected Behavior After Deployment

**Before**:
```bash
# Missing business_name
curl -X POST .../api/v1/classify -d '{"description":"Test"}'
# Returns: 503 Service Unavailable
```

**After**:
```bash
# Missing business_name
curl -X POST .../api/v1/classify -d '{"description":"Test"}'
# Returns: 400 Bad Request
# Message: "business_name is required"
```

---

## Troubleshooting

### Deployment Not Starting

1. **Check Railway Connection**:
   - Verify Railway is connected to GitHub
   - Check Railway dashboard for deployment status

2. **Check Build Logs**:
   - Go to Railway dashboard
   - Select API Gateway service
   - Check "Deployments" tab
   - Review build logs for errors

3. **Manual Trigger**:
   - Use Railway CLI or dashboard to trigger deployment

### Deployment Fails

1. **Check Build Errors**:
   - Review Railway build logs
   - Verify Dockerfile is correct
   - Check environment variables

2. **Check Service Health**:
   ```bash
   curl https://api-gateway-service-production-21fd.up.railway.app/health
   ```

3. **Rollback if Needed**:
   - Go to Railway dashboard
   - Select previous successful deployment
   - Click "Redeploy"

### Validation Still Not Working

1. **Verify Deployment**:
   - Check deployment timestamp
   - Verify latest code is deployed

2. **Check Service Logs**:
   ```bash
   railway logs --service api-gateway-service
   ```

3. **Test Directly**:
   - Use verification script
   - Check response codes

---

## Post-Deployment Testing

After deployment, run full test suite:

```bash
# Run API endpoint tests
./scripts/test-api-endpoints.sh

# Run integration tests
./scripts/test-integration.sh

# Verify deployment
./scripts/verify-deployment.sh
```

---

## Monitoring

### Check Deployment Status
- Railway Dashboard → API Gateway Service → Deployments

### Monitor Logs
```bash
railway logs --service api-gateway-service --follow
```

### Monitor Metrics
- Check error rates (should decrease)
- Check response times (should improve for invalid requests)
- Check backend service load (should decrease)

---

## Success Criteria

Deployment is successful when:
- ✅ Health check returns 200
- ✅ Missing `business_name` returns 400 (not 503)
- ✅ Valid requests still return 200
- ✅ Error message is clear: "business_name is required"
- ✅ No increase in error rates
- ✅ Service logs show validation working

---

**Next Steps**: 
1. Wait for auto-deployment or trigger manually
2. Run verification script
3. Re-run full test suite
4. Monitor for any issues

