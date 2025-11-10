# Railway Build Fix for Classification Service

## Problem
The classification-service imports `kyb-platform/internal/*` packages, but Railway builds from `services/classification-service`, making it impossible to access the root `internal/` directory using `COPY ../internal/` (Docker doesn't allow copying outside the build context).

## Solution
Configure Railway to build from the **repository root** instead of the service directory.

## Railway Dashboard Configuration

1. Go to Railway Dashboard → Your Project → Classification Service
2. Go to **Settings** → **Service Settings**
3. Set **Root Directory** to: `.` (repository root) or leave it empty
4. Go to **Settings** → **Build & Deploy**
5. Set **Dockerfile Path** to: `services/classification-service/Dockerfile`
6. Ensure **Builder** is set to: `DOCKERFILE`
7. Save and redeploy

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

