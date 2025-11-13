# Railway API Gateway Configuration Fix

## Problem

The API Gateway service is failing to build with error:
```
failed to read dockerfile: open services/api-gateway/Dockerfile: no such file or directory
```

## Root Cause

The Railway dashboard has **Root Directory** set to `/services/api-gateway`, but it should be `.` (repository root) to match the `dockerContext: "../.."` configuration.

When Root Directory is set to the service directory:
- Railway can't access the full repository structure
- The Dockerfile path resolution fails
- The build context doesn't include the entire repository

## Solution

### Railway Dashboard Settings

1. **Root Directory**: Change from `/services/api-gateway` to `.` (repository root)
   - This allows Railway to access the entire repository
   - Matches the `dockerContext: "../.."` configuration

2. **Dockerfile Path**: Should be `services/api-gateway/Dockerfile` (absolute path from repository root)
   - OR `Dockerfile` if Root Directory is set to `services/api-gateway` (but this won't work with dockerContext)

### Recommended Configuration

**Root Directory**: `.` (repository root)
**Dockerfile Path**: `services/api-gateway/Dockerfile`

This matches the pattern used by:
- `risk-assessment-service` (uses `source: "."` and `dockerfilePath: "services/risk-assessment-service/Dockerfile.go123"`)

## Alternative: If Root Directory Must Be Service Directory

If you must keep Root Directory as `/services/api-gateway`, then:
1. Remove `dockerContext: "../.."` from railway.json
2. Update Dockerfile to not rely on repository root structure
3. This would require significant Dockerfile changes

**Not recommended** - The current Dockerfile expects access to the full repository.

## Verification

After updating Root Directory to `.`:
1. Railway should find the Dockerfile at `services/api-gateway/Dockerfile`
2. Build context will be repository root (via `dockerContext: "../.."`)
3. Dockerfile can access full repository structure
4. Build should succeed

