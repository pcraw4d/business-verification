# Railway Deployment Checklist

## ✅ Pre-Deployment Verification

Run this before deploying:
```bash
./scripts/verify-railway-config.sh
```

## 🔧 Railway Dashboard Configuration

### For Each Service, Verify:

1. **Root Directory** (Settings → Service Settings)
   - Frontend: `cmd/frontend-service`
   - API Gateway: `services/api-gateway`
   - Classification: `services/classification-service`
   - Merchant: `services/merchant-service`
   - Risk Assessment: `services/risk-assessment-service`

2. **Builder Type** (Settings → Build & Deploy)
   - Must be set to: **Dockerfile** (NOT Railpack)
   - Dockerfile Path: `Dockerfile`

3. **Environment Variables**
   - `PORT` - Set automatically by Railway
   - Service-specific variables as needed

## 🐳 Dockerfile Requirements

All Dockerfiles should have:
- ✅ `wget` package (for health checks)
- ✅ Health check configured
- ✅ Proper port exposure
- ✅ Binary verification (if applicable)

## 📋 Services Status

| Service | Dockerfile | railway.json | Health Check | Status |
|---------|------------|--------------|-------------|--------|
| Frontend | ✅ | ✅ | ✅ | Ready |
| API Gateway | ✅ | ✅ | ✅ | Ready |
| Classification | ✅ | ✅ | ✅ | Fixed |
| Merchant | ✅ | ✅ | ✅ | Ready |
| Risk Assessment | ✅ | ✅ | ✅ | Ready |

## 🚀 Deployment Steps

1. **Verify Configuration**
   ```bash
   ./scripts/verify-railway-config.sh
   ```

2. **Check Railway Dashboard**
   - Verify root directories
   - Verify builder types
   - Check environment variables

3. **Trigger Deployment**
   - Push to main branch (auto-deploy)
   - Or manually trigger in Railway dashboard

4. **Monitor Deployment**
   - Check build logs
   - Verify health checks pass
   - Check runtime logs

## 🔍 Troubleshooting

### Build Fails
- Check Railway dashboard for build logs
- Verify root directory is correct
- Verify builder is set to Dockerfile
- Check Dockerfile syntax

### Health Check Fails
- Verify `wget` is installed in Dockerfile
- Check health check endpoint exists
- Verify PORT environment variable

### Service Not Starting
- Check runtime logs in Railway
- Verify binary was created
- Check for missing dependencies

