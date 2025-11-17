# Railway Docker Build Arguments for Frontend Service

## Issue

Next.js `NEXT_PUBLIC_*` environment variables are embedded at **build time**, not runtime. When building in Docker, these variables must be passed as **build arguments** to the Dockerfile.

## Solution

The Dockerfile has been updated to accept build arguments for Next.js build-time variables. Railway automatically passes environment variables as build arguments, but the Dockerfile must explicitly accept them using `ARG` declarations.

## Required Build Arguments

The following environment variables must be set in Railway and will be automatically passed as build arguments:

- `NEXT_PUBLIC_API_BASE_URL` - **REQUIRED** - API Gateway URL
- `NEXT_PUBLIC_USE_NEW_UI` - Optional - Enable new UI
- `USE_NEW_UI` - Optional - Enable new UI (alternative)
- `NODE_ENV` - Optional - Node environment (defaults to production)

## How It Works

1. **Railway Environment Variables**: Set in Railway dashboard under the service's "Variables" tab
2. **Automatic Build Args**: Railway automatically passes environment variables as build arguments
3. **Dockerfile ARG**: Dockerfile accepts them with `ARG` declarations
4. **ENV Variables**: Converted to `ENV` so they're available during `npm run build`
5. **Next.js Build**: Next.js embeds `NEXT_PUBLIC_*` variables into the client bundle

## Verification

The build verification script (`frontend/scripts/verify-build-env.js`) will check that:
- `NEXT_PUBLIC_API_BASE_URL` is set
- It's not set to localhost in production
- URL format is valid

## Troubleshooting

### Build fails with "NEXT_PUBLIC_API_BASE_URL is not set"

**Cause**: Environment variable not set in Railway or not passed to build

**Solution**:
1. Verify variable is set in Railway dashboard
2. Check variable name is exactly `NEXT_PUBLIC_API_BASE_URL` (case-sensitive)
3. Trigger a rebuild after setting the variable
4. Check build logs for verification script output

### Variables set but still not available during build

**Cause**: Railway may not be passing variables as build args

**Solution**:
1. Verify Dockerfile has `ARG` declarations for the variables
2. Check Railway build logs for build argument passing
3. Ensure variables are set at the service level, not project level (if applicable)

### Build succeeds but runtime uses localhost

**Cause**: Variables were set AFTER the build, or build didn't use the variables

**Solution**:
1. Variables must be set BEFORE building
2. Rebuild the service after setting variables
3. Check build logs to verify variables were detected by verification script

## Railway Configuration

### Setting Variables in Railway

1. Go to Railway Dashboard
2. Select the **frontend-service** (or your frontend service name)
3. Navigate to **Variables** tab
4. Add the following variables:

```
NEXT_PUBLIC_API_BASE_URL=https://api-gateway-service-production-21fd.up.railway.app
NODE_ENV=production
USE_NEW_UI=true
NEXT_PUBLIC_USE_NEW_UI=true
```

5. **Save** - Railway will automatically trigger a rebuild

### Verifying Build Arguments

Check the build logs for:
- Build verification script output showing variables are set
- No errors about missing `NEXT_PUBLIC_API_BASE_URL`
- Successful Next.js build completion

## Dockerfile Structure

```dockerfile
# Accept build arguments
ARG NEXT_PUBLIC_API_BASE_URL
ARG NEXT_PUBLIC_USE_NEW_UI
ARG USE_NEW_UI
ARG NODE_ENV

# Set as environment variables
ENV NEXT_PUBLIC_API_BASE_URL=${NEXT_PUBLIC_API_BASE_URL}
ENV NEXT_PUBLIC_USE_NEW_UI=${NEXT_PUBLIC_USE_NEW_UI}
ENV USE_NEW_UI=${USE_NEW_UI}
ENV NODE_ENV=${NODE_ENV:-production}
```

## Best Practices

1. **Set variables before first build** - Prevents failed builds
2. **Use build verification script** - Catches issues early
3. **Monitor build logs** - Verify variables are detected
4. **Test locally first** - Use Docker build with build args to test
5. **Document all variables** - Keep a list of required build-time variables

## Local Testing

To test the Dockerfile locally with build arguments:

```bash
docker build \
  --build-arg NEXT_PUBLIC_API_BASE_URL=https://api-gateway-service-production-21fd.up.railway.app \
  --build-arg NODE_ENV=production \
  -f cmd/frontend-service/Dockerfile \
  -t frontend-service:test \
  .
```

---

**Last Updated**: 2025-01-XX  
**Related**: `docs/RAILWAY_FRONTEND_DEPLOYMENT.md`, `docs/DEPLOYMENT_CHECKLIST.md`

