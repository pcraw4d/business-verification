# Railway Service Rebuild Instructions

## Issue Summary

BI-service and service-discovery are using incorrect Dockerfiles (building risk-assessment-service instead of their own services). The Dockerfiles are correct in the codebase, but Railway appears to be using cached/stale build definitions.

## Verification Completed ✅

### Dockerfile Configurations Verified:
1. **BI-service** (`cmd/business-intelligence-gateway/`):
   - ✅ Dockerfile exists and builds `kyb-business-intelligence-gateway` from `main.go`
   - ✅ railway.json points to `Dockerfile` (correct)
   - ✅ Start command: `./kyb-business-intelligence-gateway` (correct)

2. **service-discovery** (`cmd/service-discovery/`):
   - ✅ Dockerfile exists and builds `kyb-service-discovery` from `main.go`
   - ✅ railway.json points to `Dockerfile` (correct)
   - ✅ Start command: `./kyb-service-discovery` (correct)

## Solutions

### Option 1: Trigger Redeploy via Railway CLI

```bash
# Navigate to service directory
cd cmd/business-intelligence-gateway
railway service link bi-service  # Link to the service
railway redeploy

cd ../service-discovery
railway service link service-discovery  # Link to the service
railway redeploy
```

### Option 2: Force Rebuild via Railway Dashboard

1. Go to Railway Dashboard
2. Navigate to **bi-service**
3. Go to **Settings** → **Builds**
4. Click **Clear Build Cache** (if available) or **Redeploy**
5. Repeat for **service-discovery**

### Option 3: Force Rebuild by Making a Trivial Change

Make a small change to force Railway to rebuild:

```bash
# Add a comment to Dockerfiles to force rebuild
echo "# Rebuild $(date)" >> cmd/business-intelligence-gateway/Dockerfile
echo "# Rebuild $(date)" >> cmd/service-discovery/Dockerfile
git add cmd/business-intelligence-gateway/Dockerfile cmd/service-discovery/Dockerfile
git commit -m "Force Railway rebuild for BI-service and service-discovery"
git push origin main
```

### Option 4: Verify Service Root Directories

Ensure Railway service configurations have correct root directories:
- **bi-service**: Root should be `cmd/business-intelligence-gateway`
- **service-discovery**: Root should be `cmd/service-discovery`

Check in Railway Dashboard:
1. Service → Settings → Source
2. Verify **Root Directory** is set correctly

## Expected Results

After rebuild:
- ✅ BI-service should build `kyb-business-intelligence-gateway`
- ✅ service-discovery should build `kyb-service-discovery`
- ✅ Build logs should show correct build commands
- ✅ Services should start successfully

## Monitoring

After triggering rebuilds, monitor:
1. Build logs in Railway Dashboard
2. Verify build commands match expected output
3. Check service health endpoints
4. Confirm services are running correctly

