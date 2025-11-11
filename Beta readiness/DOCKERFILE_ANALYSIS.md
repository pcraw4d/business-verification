# Comprehensive Dockerfile and Build Analysis

**Date**: 2025-01-27  
**Service**: Merchant Service  
**Issue**: Build failing with module resolution errors

---

## Problem Analysis

### Error Messages
```
services/merchant-service/cmd/main.go:16:2: package kyb-platform/services/merchant-service/internal/config is not in std (/usr/local/go/src/kyb-platform/services/merchant-service/internal/config)
```

**Root Cause**: Go is looking for packages in the standard library path (`/usr/local/go/src/`) instead of finding them in the module. This indicates:
1. Go cannot find the `go.mod` file
2. The build is happening from the wrong directory
3. Module resolution is failing

---

## Current Dockerfile Structure

### Merchant Service Dockerfile (Current - Fixed)

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Disable workspace mode
ENV GOWORK=off

# Copy entire repository (Railway builds from root)
COPY . .

# Change to service directory where go.mod is located
WORKDIR /app/services/merchant-service

# Download dependencies
RUN go mod download

# Build from service directory
RUN CGO_ENABLED=0 GOOS=linux GOWORK=off go build -a -installsuffix cgo -o merchant-service ./cmd/main.go
```

**Key Points**:
- ✅ Builds from repository root context
- ✅ Changes to service directory before building
- ✅ Disables workspace mode
- ✅ Builds from correct directory where go.mod exists

---

## Module Structure Analysis

### Merchant Service Module
```
services/merchant-service/
├── go.mod (module: kyb-platform/services/merchant-service)
├── cmd/main.go (imports: kyb-platform/services/merchant-service/internal/...)
└── internal/...
```

### Root Module
```
/
├── go.mod (module: kyb-platform)
└── pkg/errors/ (kyb-platform/pkg/errors)
```

### Import Dependencies
- `main.go` imports: `kyb-platform/services/merchant-service/internal/*`
- `handlers/merchant.go` imports: `kyb-platform/pkg/errors`
- `go.mod` has: `replace kyb-platform => ../..`

---

## Railway Configuration

### railway.json
```json
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile",
    "dockerContext": "../.."
  }
}
```

**Analysis**:
- ✅ `dockerContext: "../.."` means build from repository root
- ✅ `dockerfilePath: "Dockerfile"` means use service's Dockerfile
- ✅ This is correct for accessing root `go.mod` and `pkg/`

### Railway Dashboard Settings Required

1. **Root Directory**: Repository root (`.`)
   - This allows Dockerfile to access root `go.mod` and `pkg/`

2. **Dockerfile Path**: `services/merchant-service/Dockerfile`
   - Service-specific Dockerfile

3. **Build Context**: Repository root
   - Set via `dockerContext` in railway.json

---

## Build Process Flow

### Correct Build Flow

1. **Railway starts build from repository root**
2. **Dockerfile copies entire repository** (`COPY . .`)
3. **Changes to service directory** (`WORKDIR /app/services/merchant-service`)
4. **Go finds go.mod** in current directory
5. **Go resolves imports**:
   - `kyb-platform/services/merchant-service/internal/*` → Found in current directory
   - `kyb-platform/pkg/errors` → Found via replace directive pointing to `../../pkg/errors`
6. **Build succeeds**

### Why Previous Build Failed

1. **Build was from `/app`** but `go.mod` is in `/app/services/merchant-service`
2. **Go couldn't find go.mod** so it looked in standard library
3. **Module resolution failed**

---

## Comparison with Classification Service

### Classification Service (Working)

**Module**: `kyb-platform` (root module)  
**Dockerfile**: Builds from root, copies everything  
**Build Command**: `go build ./services/classification-service/cmd/main.go`  
**Working Directory**: `/app` (root)

### Merchant Service (Fixed)

**Module**: `kyb-platform/services/merchant-service` (separate module)  
**Dockerfile**: Builds from root, changes to service directory  
**Build Command**: `go build ./cmd/main.go`  
**Working Directory**: `/app/services/merchant-service` (service directory)

**Key Difference**: Merchant service has its own module, so it must build from its directory.

---

## Dockerfile Verification Checklist

### ✅ Current Dockerfile (Fixed)

- [x] Disables workspace mode (`GOWORK=off`)
- [x] Copies entire repository structure
- [x] Changes to service directory before building
- [x] Downloads dependencies from service go.mod
- [x] Builds from service directory
- [x] Copies binary from correct path
- [x] Uses correct binary name

### Build Command Analysis

**Current**:
```dockerfile
WORKDIR /app/services/merchant-service
RUN go build -o merchant-service ./cmd/main.go
```

**Why This Works**:
- Go finds `go.mod` in current directory (`/app/services/merchant-service`)
- Go resolves `kyb-platform/services/merchant-service` imports from current directory
- Go resolves `kyb-platform/pkg/errors` via replace directive to `../../pkg/errors`

---

## Module Resolution Path

### Import Resolution

1. **`kyb-platform/services/merchant-service/internal/config`**
   - Go looks in current directory (`/app/services/merchant-service`)
   - Finds `internal/config/` ✅

2. **`kyb-platform/pkg/errors`**
   - Go checks `go.mod` replace directive
   - Finds `replace kyb-platform => ../..`
   - Resolves to `/app/pkg/errors` ✅

---

## Potential Issues and Solutions

### Issue 1: Railway Build Context

**Problem**: Railway might not be building from repository root

**Solution**: Verify Railway dashboard settings:
- Root Directory: `.` (repository root)
- Dockerfile Path: `services/merchant-service/Dockerfile`

### Issue 2: Cached Docker Layers

**Problem**: Old Docker layers might be cached

**Solution**: 
- Clear Railway build cache
- Force rebuild

### Issue 3: go.mod Replace Directive

**Problem**: Replace directive might not work in Docker

**Solution**: 
- Ensure `../../pkg/errors` exists in Docker context
- Verify `COPY . .` includes `pkg/` directory

---

## Verification Steps

### 1. Verify Dockerfile Structure
```bash
# Check Dockerfile syntax
docker build --dry-run -f services/merchant-service/Dockerfile .
```

### 2. Verify Module Resolution
```bash
# In service directory
cd services/merchant-service
go mod verify
go list -m all
```

### 3. Verify Replace Directive
```bash
# Check if replace works
cd services/merchant-service
go list -m kyb-platform
# Should show: kyb-platform => ../..
```

### 4. Test Local Build
```bash
# Build locally (if Docker available)
docker build -f services/merchant-service/Dockerfile -t merchant-test .
```

---

## Recommended Dockerfile (Final Version)

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Disable Go workspace mode
ENV GOWORK=off

# Copy entire repository (Railway builds from root)
COPY . .

# Change to service directory where go.mod is located
WORKDIR /app/services/merchant-service

# Download dependencies for the merchant service module
RUN go mod download

# Build the application from the service directory
RUN CGO_ENABLED=0 GOOS=linux GOWORK=off go build -a -installsuffix cgo -o merchant-service ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

RUN adduser -D -s /bin/sh appuser

WORKDIR /app

# Copy binary from builder stage (correct path)
COPY --from=builder /app/services/merchant-service/merchant-service .

RUN chown appuser:appuser /app/merchant-service

USER appuser

EXPOSE 8082

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

CMD ["./merchant-service"]
```

---

## Railway Configuration Checklist

### Service Settings
- [ ] Root Directory: `.` (repository root)
- [ ] Builder: `DOCKERFILE` (not Railpack)

### Build & Deploy Settings
- [ ] Dockerfile Path: `services/merchant-service/Dockerfile`
- [ ] Build Context: Repository root (via railway.json)

### Environment Variables
- [ ] All required variables set
- [ ] Supabase credentials configured

---

## Troubleshooting Guide

### If Build Still Fails

1. **Check Railway Logs**:
   - Look for "package ... is not in std" errors
   - Verify build context is repository root

2. **Verify go.mod**:
   - Check replace directive is present
   - Verify module name matches imports

3. **Check File Structure**:
   - Ensure `pkg/errors` exists in repository
   - Verify service directory structure is correct

4. **Clear Cache**:
   - Clear Railway build cache
   - Force clean rebuild

---

## Summary

### Root Cause
Build was happening from wrong directory - Go couldn't find `go.mod` file.

### Solution
1. Change WORKDIR to service directory before building
2. Build from service directory where go.mod exists
3. Fix binary copy path

### Status
✅ Dockerfile fixed and ready for deployment

---

**Last Updated**: 2025-01-27  
**Status**: ✅ **DOCKERFILE FIXED**

