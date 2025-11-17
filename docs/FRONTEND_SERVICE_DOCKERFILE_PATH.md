# Frontend Service Dockerfile Path Configuration

## Issue
Frontend service deployment failed - need to verify correct Dockerfile path.

## Correct Configuration

The Dockerfile path depends on the **Root Directory** setting in Railway:

### Option 1: Root Directory = `cmd/frontend-service` (Recommended)

**Railway Dashboard Settings:**
- **Root Directory**: `cmd/frontend-service`
- **Dockerfile Path**: `Dockerfile` (just the filename)
- **Builder**: Dockerfile

**Why this works:**
- Railway changes to `cmd/frontend-service/` directory first
- Then looks for `Dockerfile` in that directory
- Matches the file structure: `cmd/frontend-service/Dockerfile`

### Option 2: Root Directory = `.` (Repository Root)

**Railway Dashboard Settings:**
- **Root Directory**: `.` (or leave empty/default)
- **Dockerfile Path**: `cmd/frontend-service/Dockerfile` (full path from root)
- **Builder**: Dockerfile

**Why this works:**
- Railway stays at repository root
- Looks for Dockerfile at `cmd/frontend-service/Dockerfile` from root
- Matches the `railway.json` configuration

## Current Configuration

Based on `cmd/frontend-service/railway.json`:
```json
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "cmd/frontend-service/Dockerfile",
    "dockerContext": "."
  }
}
```

This indicates **Option 2** is configured:
- Root Directory should be `.` (root)
- Dockerfile Path should be `cmd/frontend-service/Dockerfile`

## How to Fix in Railway Dashboard

1. Go to Railway Dashboard → Your Project → `frontend-service-production-b225`
2. Go to **Settings** → **Service Settings**
3. Check **Root Directory**:
   - If it's `cmd/frontend-service`, set Dockerfile Path to: `Dockerfile`
   - If it's `.` or empty, set Dockerfile Path to: `cmd/frontend-service/Dockerfile`
4. Go to **Settings** → **Build & Deploy**
5. Ensure **Builder** is set to "Dockerfile" (not Railpack)
6. Set **Dockerfile Path** as determined above
7. **Save and Redeploy**

## Recommended Configuration

**Recommended: Option 1** (Root Directory = `cmd/frontend-service`)

**Advantages:**
- Simpler Dockerfile path (`Dockerfile` vs `cmd/frontend-service/Dockerfile`)
- Matches service isolation pattern
- Easier to understand

**Railway Settings:**
- Root Directory: `cmd/frontend-service`
- Dockerfile Path: `Dockerfile`
- Builder: Dockerfile

## Verification

After updating, verify:
1. Build starts successfully
2. No "Dockerfile not found" errors
3. Build completes
4. Service deploys

## Troubleshooting

If deployment still fails:

1. **Check Railway logs** for exact error message
2. **Verify Dockerfile exists** at `cmd/frontend-service/Dockerfile`
3. **Check build context** - ensure `dockerContext` in `railway.json` matches root directory
4. **Verify paths in Dockerfile** - ensure COPY commands use correct paths relative to build context

## Dockerfile Location

The Dockerfile is located at:
```
cmd/frontend-service/Dockerfile
```

This file builds:
1. Next.js frontend from `frontend/` directory
2. Go binary from `cmd/frontend-service/main.go`
3. Copies static files and Next.js build to container

---

**Last Updated**: 2025-01-17

