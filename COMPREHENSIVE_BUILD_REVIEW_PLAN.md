# Comprehensive Build Review Plan

## Overview
This document outlines a systematic approach to identify and fix all build, Docker, and codebase issues preventing successful Railway deployment.

## Current Error (FIXED)
**Type Error**: `response.json<T>()` - Fetch API's `json()` method doesn't accept type arguments in TypeScript.
**Status**: ✅ FIXED - Changed to `await response.json() as T`

## Issues Found and Fixed

### ✅ Fixed Issues
1. **response.json<T>()** - Changed to type assertion: `await response.json() as T`
2. **typeof result** - Replaced with explicit types in `getRiskHistory` and `getEnrichmentSources`
3. **useMerchantContext.ts** - Renamed to `.tsx` (JSX requires .tsx extension)
4. **Missing EnrichmentSource type** - Added to `frontend/types/merchant.ts`
5. **Missing RiskIndicatorsData type** - Added to `frontend/types/merchant.ts`
6. **Missing dialog component** - Created `frontend/components/ui/dialog.tsx`
7. **package-lock.json out of sync** - Updated with `npm install`
8. **Go version mismatch** - Updated Dockerfile to Go 1.24
9. **Static directory path** - Fixed path check in Dockerfile
10. **vitest.setup.ts mock** - Added type assertion for sessionStorage mock

---

## Review Categories

### 1. TypeScript Configuration & Type Errors

#### Checklist
- [ ] **File Extensions**: All files with JSX must use `.tsx` extension
  - Search: `find frontend -name "*.ts" -exec grep -l "JSX\|<.*>" {} \;`
  - Fix: Rename `.ts` files containing JSX to `.tsx`

- [ ] **Type Definitions**: All imported types must exist
  - Search: `grep -r "import.*from.*@/types" frontend/`
  - Verify: All types exist in `frontend/types/merchant.ts`
  - Fix: Add missing type definitions

- [ ] **Fetch API Usage**: `response.json()` doesn't accept type arguments
  - Search: `grep -r "\.json<" frontend/`
  - Fix: Use type assertion: `await response.json() as T`

- [ ] **React Context Usage**: Context must be properly typed and exported
  - Search: `grep -r "createContext" frontend/`
  - Verify: All contexts are exported and properly typed

- [ ] **Missing Imports**: All imports must resolve
  - Run: `npx tsc --noEmit` in frontend directory
  - Fix: Add missing imports or type definitions

#### Tools
```bash
# Find TypeScript errors
cd frontend && npx tsc --noEmit

# Find files with JSX but .ts extension
find frontend -name "*.ts" -exec grep -l "<.*>" {} \;

# Find all type imports
grep -r "import.*type" frontend/ | grep -v node_modules
```

---

### 2. Docker & Build Configuration

#### Checklist
- [ ] **Dockerfile Paths**: All COPY commands use correct paths
  - Review: `cmd/frontend-service/Dockerfile`
  - Verify: Static directory path is correct
  - Verify: Go version matches go.mod

- [ ] **Multi-stage Build**: All stages complete successfully
  - Stage 1: Next.js build (frontend-builder)
  - Stage 2: Go build (go-builder)
  - Stage 3: Final image (stage-2)

- [ ] **Dependencies**: All required files are copied
  - Verify: `package.json` and `package-lock.json` are in sync
  - Verify: `go.mod` and `go.sum` are in sync
  - Verify: Static files are copied correctly

- [ ] **Build Context**: Docker build context includes all necessary files
  - Review: `.dockerignore` doesn't exclude needed files
  - Verify: `dockerContext` in railway.json is correct

#### Tools
```bash
# Test Docker build locally
docker build -f cmd/frontend-service/Dockerfile -t frontend-test .

# Check Dockerfile syntax
docker build --dry-run -f cmd/frontend-service/Dockerfile .
```

---

### 3. Package Dependencies

#### Checklist
- [ ] **package.json vs package-lock.json**: Must be in sync
  - Run: `cd frontend && npm ci` (should not fail)
  - Fix: Run `npm install` to update lock file

- [ ] **Missing Dependencies**: All imported packages must be in package.json
  - Search: `grep -r "from ['\"]" frontend/ | grep -v node_modules`
  - Verify: All packages are in `package.json`

- [ ] **Peer Dependencies**: All peer dependencies are satisfied
  - Run: `npm ls` in frontend directory
  - Fix: Install missing peer dependencies

- [ ] **Version Conflicts**: No conflicting package versions
  - Run: `npm outdated` in frontend directory
  - Fix: Update conflicting packages

#### Tools
```bash
# Check for missing dependencies
cd frontend && npm ci

# List all dependencies
npm ls --depth=0

# Check for outdated packages
npm outdated
```

---

### 4. Next.js Configuration

#### Checklist
- [ ] **TypeScript Config**: `tsconfig.json` is properly configured
  - Verify: `jsx` is set to `"react-jsx"` or `"react"`
  - Verify: `paths` alias `@/*` is correct
  - Verify: `include` includes all necessary files

- [ ] **Next.js Config**: `next.config.ts` is valid
  - Verify: No syntax errors
  - Verify: All rewrites/proxies are correct

- [ ] **Build Output**: Next.js build completes successfully
  - Run: `cd frontend && npm run build`
  - Fix: Address any build errors

#### Tools
```bash
# Test Next.js build locally
cd frontend && npm run build

# Check Next.js config
node -e "console.log(require('./frontend/next.config.ts'))"
```

---

### 5. Go Build Configuration

#### Checklist
- [ ] **Go Version**: Matches go.mod requirement
  - Verify: Dockerfile uses Go 1.24 (matches go.mod)
  - Fix: Update Dockerfile if mismatch

- [ ] **Module Dependencies**: All Go modules are available
  - Run: `go mod tidy` in project root
  - Verify: `go.sum` is up to date

- [ ] **Build Commands**: All build commands are correct
  - Verify: `go build` command in Dockerfile
  - Verify: Output binary name matches start command

#### Tools
```bash
# Test Go build locally
go build -o frontend-service cmd/frontend-service/main.go

# Check Go modules
go mod verify
go mod tidy
```

---

### 6. Static Files & Assets

#### Checklist
- [ ] **Static Directory**: All static files are in correct location
  - Verify: `cmd/frontend-service/static/` contains all HTML files
  - Verify: All JS/CSS files are present

- [ ] **File Permissions**: All files are readable
  - Verify: No permission issues in Docker build

- [ ] **Asset Paths**: All asset paths are correct
  - Verify: Relative paths work in production
  - Verify: Absolute paths use correct base URL

#### Tools
```bash
# List static files
ls -la cmd/frontend-service/static/

# Check for missing files
find cmd/frontend-service/static -type f
```

---

## Systematic Review Process

### Phase 1: Immediate Fixes (Current Error)
1. Fix `response.json<T>()` usage in `api-cache.ts`
2. Search for all similar patterns
3. Test TypeScript compilation

### Phase 2: TypeScript Audit
1. Run `npx tsc --noEmit` in frontend
2. Fix all TypeScript errors
3. Verify all type definitions exist

### Phase 3: File Extension Audit
1. Find all `.ts` files with JSX
2. Rename to `.tsx`
3. Update all imports

### Phase 4: Dependency Audit
1. Verify package.json vs package-lock.json
2. Check for missing dependencies
3. Update lock file if needed

### Phase 5: Docker Build Test
1. Test Docker build locally
2. Fix any Docker-specific issues
3. Verify all paths are correct

### Phase 6: Integration Test
1. Run full build locally
2. Test Next.js build
3. Test Go build
4. Verify final Docker image

---

## Execution Order

1. **Fix Current Error** (response.json<T>)
2. **Run TypeScript Check** (find all errors)
3. **Fix File Extensions** (JSX files)
4. **Fix Type Definitions** (missing types)
5. **Verify Dependencies** (package.json sync)
6. **Test Docker Build** (local verification)
7. **Commit & Deploy** (after all fixes)

---

## Success Criteria

- [ ] `npx tsc --noEmit` passes with no errors
- [ ] `npm run build` completes successfully
- [ ] `go build` completes successfully
- [ ] Docker build completes successfully
- [ ] Railway deployment succeeds

---

## Notes

- All fixes should be committed incrementally
- Test each fix before moving to the next
- Keep a log of all issues found and fixed
- Update this document as issues are discovered

