# Railway Build Fix for Classification Service

## Problems Identified

1. **Wrong railway.json**: Railway is using the root `railway.json` which points to `risk-assessment-service` instead of the service-specific `services/classification-service/railway.json`
2. **Wrong Dockerfile**: Railway was building `risk-assessment-service` instead of `classification-service` because of the wrong railway.json
3. **Go Workspace Conflict**: Error `go: module kyb-platform appears multiple times in workspace` occurs because:
   - Root `go.mod` and `services/classification-service/go.mod` both use module name `kyb-platform`
   - Go workspace mode (`go.work`) includes both, causing a conflict during Docker builds

## Solutions Applied

### 1. Dockerfile Fixes
- Added `ENV GOWORK=off` to disable workspace mode during build
- Added `GOWORK=off` to the build command
- Updated `.dockerignore` to exclude `go.work` and `go.work.sum` files

### 2. Railway Configuration Files
- Updated `services/classification-service/railway.json` with correct paths
- This file should be used by Railway when the service is configured correctly

### 3. Railway Dashboard Configuration
Configure Railway to build from the **repository root** and use the correct railway.json.

## Railway Dashboard Configuration

**CRITICAL**: Ensure Railway is configured to use the correct service and railway.json!

The root `railway.json` points to `risk-assessment-service`, so Railway dashboard settings must override it.

1. Go to Railway Dashboard → Your Project → **Classification Service** (not Risk Assessment Service)
2. Go to **Settings** → **Service Settings**
3. Set **Root Directory** to: `.` (repository root) - this is needed for Dockerfile to access root `internal/` and `pkg/`
4. Go to **Settings** → **Build & Deploy**
5. **CRITICAL**: Set **Dockerfile Path** to: `services/classification-service/Dockerfile`
   - The root `railway.json` has `services/risk-assessment-service/Dockerfile` - you MUST override this in dashboard!
6. **CRITICAL**: Set **Start Command** to: `./classification-service`
   - The root `railway.json` has `cd services/risk-assessment-service && ./risk-assessment-service` - override this!
7. Ensure **Builder** is set to: `DOCKERFILE`
8. Save and redeploy

**Why this is needed**: When building from repo root, Railway may use the root `railway.json` which has wrong paths. Dashboard settings override the railway.json file.

## How It Works

With the root directory set to the repository root:
- Railway's build context becomes the entire repository
- The Dockerfile can access both:
  - Root `internal/` and `pkg/` directories
  - Service-specific `services/classification-service/` directory
- The Dockerfile copies:
  - Root `go.mod` and `go.sum` for module resolution
  - Root `internal/` for shared packages
  - Service `cmd/` and `internal/` for service code
- Build command: `go build ./services/classification-service/cmd/main.go`

## Alternative: If Root Directory Must Be Service Directory

If Railway must build from `services/classification-service`, you would need to:
1. Copy root `internal/` into the service directory before building (not recommended)
2. Or restructure the code to not use root internal packages (major refactor)

The recommended approach is to configure Railway to build from the repository root.

## Verification

After deploying, verify the build:
1. Check Railway build logs - should show `services/classification-service/Dockerfile` being used
2. Build should complete without "module appears multiple times" error
3. Service should start successfully and respond to health checks

## Troubleshooting

If you still see the "module appears multiple times" error:
- Verify `GOWORK=off` is set in the Dockerfile (it is)
- Check that `.dockerignore` excludes `go.work` (it does)
- Ensure Railway is building from repository root, not service directory

If Railway is using the wrong Dockerfile or railway.json:
- **Root cause**: Root `railway.json` points to `risk-assessment-service`
- **Solution**: Override in Railway dashboard settings (see above)
- Check Railway dashboard settings for the Classification Service
- Verify dashboard has `dockerfilePath: services/classification-service/Dockerfile` (not risk-assessment-service)
- Verify dashboard has `startCommand: ./classification-service` (not risk-assessment-service command)
- Ensure you're editing the correct service in Railway dashboard
- The service-specific `services/classification-service/railway.json` is correct, but Railway may use root one when building from repo root

