# Railway Environment Variables Configuration

## Required Environment Variables for All Services

### Supabase Configuration
These variables must be set for **ALL** services that use Supabase:

```bash
SUPABASE_URL=https://qpqhuqqmkjxsltzshfam.supabase.co
SUPABASE_ANON_KEY=your_supabase_anon_key_here
SUPABASE_SERVICE_ROLE_KEY=your_supabase_service_role_key_here
SUPABASE_JWT_SECRET=your_supabase_jwt_secret_here
```

### Service-Specific Ports
Each service should have its PORT variable set:

```bash
# API Gateway
PORT=8080

# Classification Service  
PORT=8081

# Merchant Service
PORT=8082

# Monitoring Service
PORT=8084

# Pipeline Service
PORT=8085

# Frontend Service
PORT=8086

# Service Discovery
PORT=8086

# Business Intelligence Service
PORT=8087
```

### Common Configuration
These can be set as shared variables:

```bash
# Environment
ENV=production

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# CORS (for API Gateway)
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=*
CORS_ALLOW_CREDENTIALS=true

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER=100
RATE_LIMIT_WINDOW_SIZE=60
```

## How to Set Environment Variables in Railway

### Method 1: Individual Service Variables
1. Go to Railway dashboard
2. Click on each service
3. Go to "Variables" tab
4. Add the required variables

### Method 2: Shared Variables (Recommended)
1. Go to Railway dashboard
2. Click on your project
3. Go to "Variables" tab
4. Add shared variables that all services can use

### Method 3: Using Railway CLI
```bash
# Set shared variables
railway variables set SUPABASE_URL=https://qpqhuqqmkjxsltzshfam.supabase.co
railway variables set SUPABASE_ANON_KEY=your_anon_key
railway variables set SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
railway variables set SUPABASE_JWT_SECRET=your_jwt_secret

# Set service-specific variables
railway variables set PORT=8080 --service api-gateway-service
railway variables set PORT=8081 --service classification-service
railway variables set PORT=8082 --service merchant-service
railway variables set PORT=8084 --service monitoring-service
railway variables set PORT=8085 --service pipeline-service
railway variables set PORT=8086 --service frontend-service
railway variables set PORT=8086 --service service-discovery
railway variables set PORT=8087 --service bi-service
```

## Critical Issue Found

**The main issue is the environment variable name mismatch:**

- ❌ **Development.env uses**: `SUPABASE_API_KEY`
- ✅ **Services expect**: `SUPABASE_ANON_KEY`

**Solution**: In Railway, make sure to set `SUPABASE_ANON_KEY` (not `SUPABASE_API_KEY`)

## Verification Commands

After setting the variables, you can verify they're working:

```bash
# Check if services can connect to Supabase
curl https://your-service-url/health

# Look for "supabase_status": {"connected": true} in the response
```

## Troubleshooting

### If Supabase connection still shows false:
1. Verify the Supabase URL is correct
2. Check that the anon key is valid
3. Ensure the service role key has proper permissions
4. Check Railway logs for connection errors

### If services are still unhealthy:
1. Check Railway service logs
2. Verify all required environment variables are set
3. Ensure Supabase project is active
4. Check network connectivity
