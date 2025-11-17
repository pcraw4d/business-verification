# Railway Frontend Deployment Configuration

## Overview

This document describes the configuration needed to deploy the Next.js frontend to Railway with proper environment variables and feature flags.

## Environment Variables

### Required Variables

Set these in Railway for the frontend service:

```bash
# Frontend Service Port
PORT=8086

# API Base URL (for Next.js API calls)
NEXT_PUBLIC_API_BASE_URL=https://api-gateway-service-production-21fd.up.railway.app

# Feature Flag: Enable New UI
USE_NEW_UI=true
NEXT_PUBLIC_USE_NEW_UI=true

# Next.js Build Configuration
NEXT_PUBLIC_STATIC_EXPORT=false  # Set to true if using static export
NEXT_PUBLIC_UNOPTIMIZED_IMAGES=false  # Set to true if image optimization issues occur

# Environment
NODE_ENV=production
```

### Optional Variables

```bash
# Service Configuration
SERVICE_NAME=frontend-service

# Build Optimization
NEXT_PUBLIC_BUILD_ID=production
```

## Deployment Steps

### 1. Build Frontend

Before deploying, build the Next.js application:

```bash
cd frontend
npm install
npm run build
```

Or use the build script:

```bash
./scripts/build-frontend.sh
```

### 2. Configure Railway Service

**⚠️ CRITICAL: Set environment variables BEFORE building!**

Next.js embeds `NEXT_PUBLIC_*` variables at build time. If you set them after building, you must rebuild.

1. Go to Railway Dashboard
2. Select the frontend service
3. Go to "Variables" tab
4. Add all required environment variables listed above
5. **Save the variables** - Railway will automatically trigger a rebuild

### 3. Build Configuration

The frontend service should be configured to:

1. **Build Command**: `cd frontend && npm install && npm run build`
2. **Start Command**: 
   - For standalone mode: `cd frontend && npm start`
   - For static export: Serve from `cmd/frontend-service/static/.next/`

### 4. Routing Configuration

The Go frontend service (`cmd/frontend-service`) handles routing:

- If `USE_NEW_UI=true`: Routes to Next.js application
- If `USE_NEW_UI=false`: Routes to legacy HTML files

The routing logic is in `cmd/frontend-service/routing.go`.

## Feature Flag Usage

The `USE_NEW_UI` environment variable controls which UI is served:

- **`USE_NEW_UI=true`**: New Next.js UI with shadcn components
- **`USE_NEW_UI=false`**: Legacy HTML/CSS/JS UI

This allows gradual migration and rollback if needed.

## Build Output

### Standalone Mode (Default)

When using standalone mode:
- Build output: `frontend/.next/standalone/`
- Copy to: `cmd/frontend-service/static/.next/`

### Static Export Mode

When using static export:
- Set `NEXT_PUBLIC_STATIC_EXPORT=true`
- Build output: `frontend/out/`
- Copy to: `cmd/frontend-service/static/`

## Verification

After deployment, verify:

1. **Health Check**: `https://your-service.railway.app/health`
2. **New UI**: Navigate to any page and verify shadcn UI components render
3. **API Integration**: Verify API calls work correctly
4. **Charts**: Verify charts render with data

## Troubleshooting

### Issue: Next.js build fails

**Solution**: Check Node.js version (should be 18+). Update Railway build settings.

### Issue: API calls fail

**Solution**: 
1. Verify `NEXT_PUBLIC_API_BASE_URL` is set correctly in Railway
2. **IMPORTANT**: If you just set the variable, you must rebuild the service
3. Check Railway build logs to ensure the variable was available during build
4. See `docs/RAILWAY_FRONTEND_REBUILD_REQUIRED.md` for detailed instructions

### Issue: Legacy UI still showing

**Solution**: Verify `USE_NEW_UI=true` is set in Railway environment variables.

### Issue: Images not loading

**Solution**: Set `NEXT_PUBLIC_UNOPTIMIZED_IMAGES=true` if using static export or having image optimization issues.

## Rollback Procedure

If issues occur with the new UI:

1. Set `USE_NEW_UI=false` in Railway
2. Redeploy the service
3. Legacy UI will be served

## Next Steps

After successful deployment:

1. Monitor error rates
2. Check performance metrics
3. Verify all pages load correctly
4. Test form submissions
5. Verify chart rendering

