# Railway Setup Guide

## Quick Setup Instructions

### Step 1: Choose Your Environment File

You have two options:

1. **`railway-essential.env`** - Contains only the essential variables (recommended for quick setup)
2. **`railway.env`** - Contains all possible variables (comprehensive setup)

### Step 2: Set Environment Variables in Railway

#### Option A: Using Railway Dashboard (Recommended)

1. Go to [Railway Dashboard](https://railway.app/dashboard)
2. Click on your project
3. Go to "Variables" tab
4. Add variables from your chosen .env file
5. **Replace placeholder values** with your actual values

#### Option B: Using Railway CLI

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login to Railway
railway login

# Set variables from .env file
railway variables set --file railway-essential.env

# Or set individual variables
railway variables set SUPABASE_URL=https://qpqhuqqmkjxsltzshfam.supabase.co
railway variables set SUPABASE_ANON_KEY=your_actual_anon_key
railway variables set SUPABASE_SERVICE_ROLE_KEY=your_actual_service_role_key
railway variables set SUPABASE_JWT_SECRET=your_actual_jwt_secret
```

### Step 3: Set Service-Specific Ports

For each service, set the appropriate PORT variable:

```bash
# API Gateway
railway variables set PORT=8080 --service api-gateway-service

# Classification Service
railway variables set PORT=8081 --service classification-service

# Merchant Service
railway variables set PORT=8082 --service merchant-service

# Monitoring Service
railway variables set PORT=8084 --service monitoring-service

# Pipeline Service
railway variables set PORT=8085 --service pipeline-service

# Frontend Service
railway variables set PORT=8086 --service frontend-service

# Service Discovery
railway variables set PORT=8086 --service service-discovery

# Business Intelligence Service
railway variables set PORT=8087 --service bi-service
```

### Step 4: Critical Variables to Set First

These are the most important variables that must be set correctly:

```bash
# Supabase Configuration (CRITICAL)
SUPABASE_URL=https://qpqhuqqmkjxsltzshfam.supabase.co
SUPABASE_ANON_KEY=your_supabase_anon_key_here
SUPABASE_SERVICE_ROLE_KEY=your_supabase_service_role_key_here
SUPABASE_JWT_SECRET=your_supabase_jwt_secret_here

# Environment
ENV=production
ENVIRONMENT=production

# API Gateway Service URLs
CLASSIFICATION_SERVICE_URL=https://classification-service-production.up.railway.app
MERCHANT_SERVICE_URL=https://merchant-service-production.up.railway.app
FRONTEND_URL=https://frontend-service-production-b225.up.railway.app
```

### Step 5: Verify Setup

After setting the variables, test your deployment:

```bash
# Run the test script
./test_fixes.sh

# Or test manually
curl https://your-api-gateway-url/health
curl https://your-merchant-service-url/health
curl https://your-classification-service-url/health
```

### Step 6: Check for Success

Look for these indicators of success:

1. **Health endpoints return 200 OK**
2. **Supabase connection shows `"connected": true`**
3. **API Gateway can proxy to other services**
4. **Classification API returns business codes**
5. **No environment variable errors in logs**

## Troubleshooting

### Common Issues

1. **Supabase Connection Failed**
   - Check `SUPABASE_ANON_KEY` is set (not `SUPABASE_API_KEY`)
   - Verify Supabase project is active
   - Check network connectivity

2. **API Gateway 404 Errors**
   - Verify service URLs are correct
   - Check that services are running
   - Ensure PORT variables are set correctly

3. **Services Not Starting**
   - Check Railway logs for errors
   - Verify all required environment variables are set
   - Ensure Supabase credentials are valid

### Getting Supabase Credentials

1. Go to [Supabase Dashboard](https://supabase.com/dashboard)
2. Select your project
3. Go to Settings > API
4. Copy the following:
   - **Project URL** → `SUPABASE_URL`
   - **anon public** key → `SUPABASE_ANON_KEY`
   - **service_role** key → `SUPABASE_SERVICE_ROLE_KEY`
   - **JWT Secret** → `SUPABASE_JWT_SECRET`

## File Descriptions

- **`railway-essential.env`** - Essential variables only (recommended)
- **`railway.env`** - Complete variable set (comprehensive)
- **`test_fixes.sh`** - Testing script to verify deployment
- **`RAILWAY_ENVIRONMENT_VARIABLES.md`** - Detailed documentation

## Next Steps

After successful setup:

1. Test all API endpoints
2. Verify inter-service communication
3. Test with real business data
4. Set up monitoring and alerting
5. Configure custom domains (optional)

## Support

If you encounter issues:

1. Check Railway service logs
2. Verify environment variables
3. Test individual services
4. Check Supabase connectivity
5. Review this setup guide
