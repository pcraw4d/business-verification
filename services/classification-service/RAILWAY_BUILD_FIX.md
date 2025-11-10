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
- Updated root `railway.json` to point to `classification-service` instead of `risk-assessment-service`
- Updated `services/classification-service/railway.json` with correct paths
- Both files now have consistent configuration for classification-service

### 3. Railway Dashboard Configuration
Configure Railway to build from the **repository root** and use the correct railway.json.

## Railway Dashboard Configuration

**Root railway.json has been fixed!** The root `railway.json` now correctly points to `classification-service`.

1. Go to Railway Dashboard → Your Project → **Classification Service**
2. Go to **Settings** → **Service Settings**
3. Set **Root Directory** to: `.` (repository root) - this is needed for Dockerfile to access root `internal/` and `pkg/`
4. Go to **Settings** → **Build & Deploy**
5. Verify **Dockerfile Path** is: `services/classification-service/Dockerfile` (should match root railway.json)
6. Verify **Start Command** is: `./classification-service` (should match root railway.json)
7. Ensure **Builder** is set to: `DOCKERFILE`
8. Save and redeploy

**Note**: The root `railway.json` is now correctly configured, so Railway should automatically use the right paths. You can still override settings in the dashboard if needed.

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
- **Root cause**: Root `railway.json` was pointing to `risk-assessment-service` (now fixed)
- **Solution**: Root `railway.json` has been updated to point to `classification-service`
- Check Railway dashboard settings for the Classification Service
- Verify root `railway.json` has `dockerfilePath: services/classification-service/Dockerfile`
- Verify root `railway.json` has `startCommand: ./classification-service`
- Ensure you're editing the correct service in Railway dashboard
- If issues persist, you can override settings in Railway dashboard

