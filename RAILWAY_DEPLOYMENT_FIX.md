# Railway Deployment Fix

## Issues Identified

### 1. Frontend Service Dockerfile
- ❌ Missing `wget` package (needed for health check)
- ❌ No verification that static directory exists
- ❌ No verification that binary was created
- ✅ Fixed: Added wget, directory check, and binary verification

### 2. Potential Issues
- Railway might auto-detect Go and use Railpack instead of Dockerfile
- Need to ensure root directory is set correctly in Railway dashboard

## Fixes Applied

### Frontend Service Dockerfile
```dockerfile
# Added wget for health checks
RUN apk add --no-cache git ca-certificates wget

# Added go mod tidy to ensure dependencies are correct
RUN go mod tidy && go mod download

# Added verification steps
RUN test -d static || (echo "ERROR: static directory not found" && exit 1)
RUN test -f frontend-service || (echo "ERROR: frontend-service binary not created" && exit 1)

# Fixed health check to use PORT environment variable
CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8086}/health || exit 1
```

## Railway Dashboard Configuration

### For Each Service:

1. **Go to Railway Dashboard** → Your Project → Service
2. **Settings** → **Service Settings**
3. **Root Directory**: Set to service directory
   - Frontend: `cmd/frontend-service`
   - API Gateway: `services/api-gateway`
   - Classification: `services/classification-service`
   - Merchant: `services/merchant-service`
   - Risk Assessment: `services/risk-assessment-service`

4. **Settings** → **Build & Deploy**
5. **Builder**: Select "Dockerfile" (not Railpack)
6. **Dockerfile Path**: `Dockerfile`
7. **Save and Redeploy**

## Verification Steps

After fixing, verify:
1. Build completes successfully
2. Health check passes
3. Service starts correctly
4. Logs show no errors

