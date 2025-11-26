# CI/CD Setup for Railway Auto-Deployment

**Date:** November 23, 2025  
**Status:** ✅ **WORKFLOW CREATED**

## Overview

Created a new GitHub Actions workflow (`.github/workflows/railway-deploy.yml`) that automatically deploys services to Railway when code is pushed to the `main` branch.

## How It Works

### 1. Change Detection
The workflow automatically detects which services have changed:
- **API Gateway**: Changes in `services/api-gateway/` or `internal/api/`
- **Merchant Service**: Changes in `services/merchant-service/`
- **Risk Assessment Service**: Changes in `services/risk-assessment-service/`
- **Frontend Service**: Changes in `frontend/` or `services/frontend-service/`
- **Python ML Service**: Changes in `python_ml_service/`

### 2. Selective Deployment
Only services with changes are deployed, saving time and resources.

### 3. Health Check Verification
After each deployment, the workflow verifies the service is healthy before marking the job as successful.

## Required GitHub Secrets

The workflow requires the following secrets to be configured in GitHub:

1. **RAILWAY_TOKEN**
   - Your Railway API token
   - Get it from: Railway Dashboard > Settings > Tokens
   - Or: `railway tokens` command

2. **RAILWAY_PROJECT_ID** (optional, for reference)
   - Your Railway project ID
   - Found in Railway Dashboard URL or project settings

### Setting Up Secrets

1. Go to: `https://github.com/pcraw4d/business-verification/settings/secrets/actions`
2. Click "New repository secret"
3. Add:
   - Name: `RAILWAY_TOKEN`
   - Value: Your Railway token

## Workflow Triggers

The workflow runs automatically when:
- Code is pushed to `main` branch
- Changes are detected in service directories
- Manually triggered via `workflow_dispatch`

## Deployment Process

1. **Change Detection** - Determines which services changed
2. **Service Deployment** - Deploys each changed service in parallel
3. **Health Verification** - Verifies each service is healthy
4. **Summary Generation** - Creates deployment summary

## Service URLs

After deployment, services are available at:
- **API Gateway**: https://api-gateway-service-production-21fd.up.railway.app
- **Merchant Service**: https://merchant-service-production.up.railway.app
- **Risk Assessment Service**: https://risk-assessment-service-production.up.railway.app
- **Frontend Service**: https://frontend-service-production-b225.up.railway.app
- **Python ML Service**: Check Railway dashboard for URL (or set `PYTHON_ML_SERVICE_URL` secret)

## Monitoring

### GitHub Actions
- View workflow runs: `https://github.com/pcraw4d/business-verification/actions`
- Check deployment status for each service
- View logs for troubleshooting

### Railway Dashboard
- Monitor build progress
- View deployment logs
- Check service health

## Troubleshooting

### Workflow Not Running
- Check if `RAILWAY_TOKEN` secret is configured
- Verify workflow file is in `.github/workflows/` directory
- Check if changes are in monitored paths

### Deployment Failing
- Check Railway build logs
- Verify service configuration
- Check for compilation errors
- Verify Railway token has correct permissions

### Service Not Healthy
- Check Railway service logs
- Verify environment variables are set
- Check database connections
- Verify service dependencies

## Next Steps

1. **Configure Secrets** (if not already done)
   - Add `RAILWAY_TOKEN` to GitHub secrets

2. **Test Workflow**
   - Push a small change to trigger the workflow
   - Monitor GitHub Actions for execution
   - Verify services deploy successfully

3. **Monitor Deployments**
   - Check GitHub Actions dashboard regularly
   - Set up notifications if needed
   - Review deployment summaries

---

**Last Updated:** November 23, 2025

