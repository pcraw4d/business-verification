# Python ML Service Railway Setup

**Date**: November 26, 2025  
**Status**: ✅ **Auto-Deployment Configured**

---

## Overview

The Python ML Service is now included in the Railway auto-deployment workflow. When you push changes to the `python_ml_service/` directory, it will automatically deploy to Railway.

---

## Railway Service Setup

### 1. Create Railway Service (First Time Only)

If the Python ML service doesn't exist in Railway yet:

1. Go to Railway Dashboard
2. Select your project
3. Click "New Service"
4. Choose "GitHub Repo" or "Empty Service"
5. Name it: `python-ml-service`
6. Connect it to your repository

### 2. Configure Service Settings

The service will use the `python_ml_service/railway.json` configuration which includes:
- Dockerfile build
- Health check endpoint: `/health`
- Quantization enabled by default
- Production environment variables

### 3. Set Environment Variables (Optional)

If you need to override defaults, set these in Railway:
- `USE_QUANTIZATION=true` (default)
- `QUANTIZATION_DTYPE=qint8` (default)
- `MODEL_SAVE_PATH=/app/models/distilbart` (default)
- `QUANTIZED_MODELS_PATH=/app/models/quantized` (default)

---

## Auto-Deployment

### How It Works

1. **Push to GitHub**: When you push changes to `python_ml_service/` directory
2. **Workflow Triggers**: GitHub Actions detects the changes
3. **Service Deploys**: Railway CLI deploys the service
4. **Health Check**: Workflow verifies the service is healthy

### Monitoring Deployment

1. **GitHub Actions**: 
   - Go to: `https://github.com/pcraw4d/business-verification/actions`
   - Look for "Railway Auto-Deploy" workflow
   - Check "Deploy Python ML Service" job

2. **Railway Dashboard**:
   - Go to your Railway project
   - Click on `python-ml-service`
   - View deployment logs

### Verify Quantization

After deployment, check the logs for:
```
✅ DistilBART classifier initialized with quantization: True
```

Or via API:
```bash
curl https://your-python-ml-service-url.up.railway.app/model-info | jq '.quantization_enabled'
```

---

## Service URL

After first deployment:

1. **Get URL from Railway Dashboard**:
   - Go to Railway project
   - Click on `python-ml-service`
   - Copy the public URL

2. **Set GitHub Secret (Optional)**:
   - Go to: `https://github.com/pcraw4d/business-verification/settings/secrets/actions`
   - Add secret: `PYTHON_ML_SERVICE_URL`
   - Value: Your Railway service URL (e.g., `https://python-ml-service-production.up.railway.app`)

This enables health check verification in the deployment workflow.

---

## Troubleshooting

### Service Not Deploying

**Check**:
- Is `python-ml-service` created in Railway?
- Is `RAILWAY_TOKEN` secret set in GitHub?
- Are changes in `python_ml_service/` directory?

**Fix**:
- Create the service in Railway if it doesn't exist
- Verify Railway token has correct permissions
- Check GitHub Actions logs for errors

### Health Check Failing

**Check**:
- Service is running in Railway
- Health endpoint is accessible: `/health`
- Service URL is correct

**Fix**:
- Check Railway service logs
- Verify service started successfully
- Ensure port 8000 is exposed

### Quantization Not Enabled

**Check**:
- Environment variables in Railway
- Service logs for quantization status
- Model loading errors

**Fix**:
- Verify `USE_QUANTIZATION=true` is set
- Check logs for model loading errors
- Ensure PyTorch >= 2.6.0 is installed

---

## Next Steps

1. ✅ **Workflow Updated**: Python ML service added to auto-deployment
2. ⏳ **Create Railway Service**: Create `python-ml-service` in Railway (if not exists)
3. ⏳ **First Deployment**: Push a change to trigger deployment
4. ⏳ **Verify Quantization**: Check logs for quantization confirmation
5. ⏳ **Set Service URL**: Add `PYTHON_ML_SERVICE_URL` secret (optional)

---

**Status**: ✅ **Ready for Auto-Deployment**

