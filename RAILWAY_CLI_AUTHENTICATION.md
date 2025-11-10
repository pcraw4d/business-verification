# Railway CLI Authentication and Log Access

## Browserless Authentication

Since Railway CLI requires interactive authentication, you'll need to run these commands in your terminal:

### Step 1: Authenticate with Browserless Mode

```bash
railway login --browserless
```

This will:
1. Display a pairing code
2. Provide a URL to visit
3. You'll enter the pairing code on the website
4. The CLI will be authenticated

### Step 2: Link to Your Project

```bash
cd "/Users/petercrawford/New tool"
railway link
```

Select your project when prompted.

### Step 3: View Logs for Failing Services

```bash
# View logs for risk-assessment-service
railway logs --service risk-assessment-service --tail 100

# View logs for classification-service
railway logs --service classification-service --tail 100

# View all logs
railway logs --tail 100
```

### Step 4: Check Deployment Status

```bash
# Check status of all services
railway status

# Check specific service
railway status --service risk-assessment-service
```

### Alternative: Use Railway Dashboard

If CLI authentication is not possible, you can:
1. Go to https://railway.app
2. Navigate to your project
3. Click on each service (risk-assessment-service, classification-service)
4. View the "Deployments" tab for build logs
5. View the "Logs" tab for runtime logs

## Common Issues to Check in Logs

### For risk-assessment-service:
- Go version mismatch (should be 1.24.0)
- Missing startup_debug.sh file
- LD_LIBRARY_PATH undefined variable
- ONNX Runtime library loading issues
- Model file not found

### For classification-service:
- Module path issues (kyb-platform/internal/* imports)
- Build context issues (root directory vs service directory)
- Go module resolution errors
- Missing internal packages

## Quick Fix Commands

After reviewing logs, you can:

```bash
# Redeploy a specific service
railway up --service risk-assessment-service

# View recent build logs
railway logs --service risk-assessment-service --deployment

# Check environment variables
railway variables --service risk-assessment-service
```

