# Service Deployment Directory Audit

## 🔍 Issue Found

**Problem:** Frontend files were being edited in `services/frontend/public/` but Railway deploys from `cmd/frontend-service/static/`. This caused fixes to not be deployed.

## ✅ Fix Applied

1. ✅ Synced `add-merchant.html` to deployment directory
2. ✅ Synced `merchant-details.html` to deployment directory
3. ✅ Created sync script: `scripts/sync-frontend-files.sh`

## 📋 Service Directory Structure Analysis

### Frontend Services

| Service | Source Directory | Deployment Directory | Status |
|---------|-----------------|---------------------|--------|
| **Frontend (Deployed)** | `services/frontend/public/` | `cmd/frontend-service/static/` | ⚠️ **MISMATCH - FIXED** |
| Frontend (Alternative) | `services/frontend/public/` | `services/frontend/public/` | ✅ Matches (not deployed) |
| Frontend Service (Minimal) | N/A | `services/frontend-service/static/` | ✅ Minimal files only |

### Other Services Analysis

#### API Gateway Service
- **Location:** `services/api-gateway/`
- **Deployment:** `services/api-gateway/` (matches)
- **Static Files:** None (API only)
- **Status:** ✅ No issues

#### Classification Service
- **Location:** `services/classification-service/`
- **Deployment:** `services/classification-service/` (matches)
- **Static Files:** None (API only)
- **Status:** ✅ No issues

#### Merchant Service
- **Location:** `services/merchant-service/`
- **Deployment:** `services/merchant-service/` (matches)
- **Static Files:** None (API only)
- **Status:** ✅ No issues

#### Risk Assessment Service
- **Location:** `services/risk-assessment-service/`
- **Deployment:** `services/risk-assessment-service/` (matches)
- **Static Files:** `beta-testing/dashboard.html` (internal only)
- **Status:** ✅ No issues

## 🎯 Root Cause

The codebase has **multiple frontend service implementations**:

1. **`cmd/frontend-service/`** - ✅ **THIS IS DEPLOYED TO RAILWAY**
   - Uses `./static/` directory
   - Has `railway.json` configured
   - Serves all HTML pages

2. **`services/frontend/`** - ❌ **NOT DEPLOYED**
   - Uses `./public/` directory
   - Has `railway.json` but may not be active
   - This is where we were editing files

3. **`services/frontend-service/`** - ❌ **NOT DEPLOYED**
   - Minimal implementation
   - Only has `index.html` and basic JS

## 🔧 Solution

### Immediate Fix (Applied)
- ✅ Synced critical files: `add-merchant.html`, `merchant-details.html`
- ✅ Created sync script for future use

### Long-term Solution

**Option 1: Use Sync Script (Recommended)**
```bash
# Run before every deployment
./scripts/sync-frontend-files.sh
git add cmd/frontend-service/static/
git commit -m "Sync frontend files for deployment"
git push
```

**Option 2: Consolidate Services**
- Remove duplicate frontend services
- Keep only `cmd/frontend-service/` as the single source of truth
- Move all development to that directory

**Option 3: Update Development Workflow**
- Always edit files in `cmd/frontend-service/static/`
- Use `services/frontend/public/` only for reference/backup

## 📊 File Count Comparison

- **Source (`services/frontend/public/`):** 38 HTML files
- **Deployment (`cmd/frontend-service/static/`):** 37 HTML files
- **Difference:** 1 file (likely a new file not yet synced)

## ✅ Verification Checklist

- [x] `add-merchant.html` synced
- [x] `merchant-details.html` synced
- [x] All other HTML files synced (sync script completed)
- [x] JS files synced (sync script completed)
- [x] CSS files synced (sync script completed)
- [x] Components synced (sync script completed)
- [x] Pre-commit hook created (`.git/hooks/pre-commit-frontend-sync`)
- [x] CI/CD check created (`.github/workflows/frontend-sync-check.yml`)
- [x] Verification script created (`scripts/verify-deployment-sync.sh`)

## 🚨 Action Required

**Before making any frontend changes:**

1. **Edit files in:** `cmd/frontend-service/static/` (deployment directory)
2. **OR** edit in `services/frontend/public/` and run sync script
3. **Always verify** changes are in `cmd/frontend-service/static/` before committing

## 📝 Other Services Status

All other services (API Gateway, Classification, Merchant, Risk Assessment) are **NOT affected** because they:
- Don't serve static HTML files
- Have matching source and deployment directories
- Are pure API services

**Conclusion:** The directory mismatch issue **ONLY affects the frontend service**.

