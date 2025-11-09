# Railway Deployment Guide - Frontend Service

## Pre-Deployment Checklist ✅

All components have been verified and are ready for deployment:

- ✅ **52 new UI component files** synced to `cmd/frontend-service/static/`
- ✅ **Dockerfile** configured for Go 1.22
- ✅ **railway.json** configuration file created
- ✅ **All routes** registered in `main.go`
- ✅ **All changes** committed and pushed to GitHub

## New Features Deployed

1. **Admin Dashboard** (`/admin`) - Memory monitoring, system metrics
2. **User Registration** (`/register`) - User registration UI
3. **Session Management** (`/sessions`) - Active session management
4. **ML Model Management** (`/admin/models`) - Model performance tracking
5. **Analytics Insights** (`/analytics-insights`) - Business insights dashboard
6. **Queue Management** (`/admin/queue`) - Task queue management
7. **Export Functionality** - CSV, PDF, JSON export buttons
8. **Data Enrichment** - External data source integration
9. **Risk UI Components** - Tooltips, score panels, drag-drop

## Deployment Options

### Option 1: Railway Dashboard (Recommended)

1. **Navigate to Railway Dashboard**
   - Go to https://railway.app
   - Log in to your account
   - Select your project (or create a new one)

2. **Add New Service**
   - Click "New" → "GitHub Repo"
   - Select your repository
   - Railway will detect the `railway.json` in `cmd/frontend-service/`

3. **Configure Service**
   - **Root Directory**: Set to `cmd/frontend-service`
   - **Build Command**: Railway will use the Dockerfile automatically
   - **Start Command**: `./frontend-service` (from railway.json)
   - **Port**: Railway will auto-detect from EXPOSE in Dockerfile (8086)

4. **Environment Variables** (if needed)
   - `PORT`: Railway will set this automatically
   - `SERVICE_NAME`: `frontend-service` (optional)
   - Any API gateway URLs or service endpoints

5. **Deploy**
   - Railway will automatically build and deploy
   - Monitor the deployment logs
   - Once deployed, Railway will provide a public URL

### Option 2: Railway CLI

```bash
# 1. Login to Railway
railway login

# 2. Navigate to frontend service directory
cd cmd/frontend-service

# 3. Link to Railway project (if not already linked)
railway link

# 4. Deploy
railway up
```

### Option 3: GitHub Integration (Automatic)

If your Railway project is connected to GitHub:

1. **Verify GitHub Connection**
   - Go to Railway Dashboard → Your Project → Settings
   - Ensure GitHub repository is connected
   - Enable "Auto Deploy" for main branch

2. **Automatic Deployment**
   - Every push to `main` will trigger a deployment
   - Railway will detect changes and rebuild automatically

## Verification Steps

After deployment, verify all features:

1. **Health Check**
   ```
   https://your-railway-url.up.railway.app/health
   ```

2. **Admin Dashboard**
   ```
   https://your-railway-url.up.railway.app/admin
   ```

3. **User Registration**
   ```
   https://your-railway-url.up.railway.app/register
   ```

4. **Session Management**
   ```
   https://your-railway-url.up.railway.app/sessions
   ```

5. **ML Models**
   ```
   https://your-railway-url.up.railway.app/admin/models
   ```

6. **Analytics Insights**
   ```
   https://your-railway-url.up.railway.app/analytics-insights
   ```

7. **Queue Management**
   ```
   https://your-railway-url.up.railway.app/admin/queue
   ```

## Configuration Files

### `cmd/frontend-service/railway.json`
```json
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile"
  },
  "deploy": {
    "startCommand": "./frontend-service",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 300,
    "healthcheckInterval": 30,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

### `cmd/frontend-service/Dockerfile`
- Uses Go 1.22 Alpine
- Builds static binary
- Exposes port 8086
- Includes health check

## Troubleshooting

### Build Failures
- Check Railway build logs
- Verify `go.mod` and dependencies
- Ensure Dockerfile syntax is correct

### Runtime Errors
- Check Railway service logs
- Verify environment variables
- Ensure port configuration matches

### Health Check Failures
- Verify `/health` endpoint is accessible
- Check service is listening on correct port
- Review health check timeout settings

## Post-Deployment

1. **Update API Gateway URLs**
   - Update frontend API configuration to point to Railway URLs
   - Update CORS settings if needed

2. **Test All Features**
   - Navigate through all new pages
   - Test export functionality
   - Verify admin features (with admin role)

3. **Monitor Performance**
   - Check Railway metrics dashboard
   - Monitor response times
   - Review error logs

## Support

For Railway-specific issues:
- Railway Docs: https://docs.railway.app
- Railway Discord: https://discord.gg/railway

For application issues:
- Check service logs in Railway dashboard
- Review application logs in `cmd/frontend-service`

