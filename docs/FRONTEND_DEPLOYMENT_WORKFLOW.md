# Frontend Deployment Workflow

## 🚨 Important: Directory Structure

The frontend service has **two directories** that must be kept in sync:

1. **Source Directory:** `services/frontend/public/` (where you edit files)
2. **Deployment Directory:** `cmd/frontend-service/static/` (what Railway deploys)

## ⚠️ Critical Rule

**Railway deploys from `cmd/frontend-service/static/`**, NOT from `services/frontend/public/`.

If you edit files in `services/frontend/public/`, you **MUST** sync them to `cmd/frontend-service/static/` before committing.

## 🔄 Workflow Options

### Option 1: Edit in Deployment Directory (Recommended for Quick Fixes)

```bash
# Edit files directly in deployment directory
vim cmd/frontend-service/static/add-merchant.html

# Commit and push
git add cmd/frontend-service/static/
git commit -m "Fix merchant form"
git push
```

**Pros:** No sync needed  
**Cons:** Files can get out of sync with source

### Option 2: Edit in Source + Sync (Recommended for Development)

```bash
# 1. Edit files in source directory
vim services/frontend/public/add-merchant.html

# 2. Sync to deployment directory
./scripts/sync-frontend-files.sh

# 3. Verify sync
./scripts/verify-deployment-sync.sh

# 4. Commit both directories
git add services/frontend/public/ cmd/frontend-service/static/
git commit -m "Update merchant form"
git push
```

**Pros:** Keeps source and deployment in sync  
**Cons:** Requires running sync script

## 🛡️ Automatic Safeguards

### Pre-commit Hook

A pre-commit hook automatically checks if frontend files need syncing:

- **Location:** `.git/hooks/pre-commit`
- **Function:** Detects changes in `services/frontend/public/` and auto-syncs critical files
- **Action:** Automatically runs sync script and stages synced files

### CI/CD Check

GitHub Actions automatically verifies file sync on pull requests:

- **Workflow:** `.github/workflows/frontend-sync-check.yml`
- **Function:** Checks if critical files are synced
- **Action:** Fails PR if files are out of sync

## 📋 Verification Scripts

### Sync All Files

```bash
./scripts/sync-frontend-files.sh
```

Syncs all HTML, JS, CSS, and component files from source to deployment directory.

### Verify Critical Files

```bash
./scripts/verify-deployment-sync.sh
```

Checks if critical files (`add-merchant.html`, `merchant-details.html`) are synced.

## 🚀 Deployment Process

1. **Make changes** in `services/frontend/public/` or `cmd/frontend-service/static/`
2. **Run sync script** if editing in source directory
3. **Verify sync** using verification script
4. **Commit changes** (pre-commit hook will auto-sync if needed)
5. **Push to main** (Railway auto-deploys)
6. **Verify deployment** in Railway dashboard

## 📝 Best Practices

1. **Always verify** files are synced before committing
2. **Use sync script** when editing multiple files
3. **Check CI/CD** status after opening PRs
4. **Test in production** after Railway deployment completes

## 🔍 Troubleshooting

### Files Not Deploying

1. Check if files are in `cmd/frontend-service/static/`
2. Verify Railway deployment completed
3. Check Railway logs for build errors
4. Clear browser cache (Ctrl+Shift+R)

### Sync Script Fails

1. Check if source directory exists: `services/frontend/public/`
2. Check if target directory exists: `cmd/frontend-service/static/`
3. Verify file permissions
4. Check disk space

### Pre-commit Hook Not Working

1. Verify hook is executable: `chmod +x .git/hooks/pre-commit`
2. Check hook exists: `ls -la .git/hooks/pre-commit`
3. Test manually: `.git/hooks/pre-commit`

## 📚 Related Documentation

- `SERVICE_DEPLOYMENT_AUDIT.md` - Full audit of deployment structure
- `BETA_TESTING_VERIFICATION.md` - Testing guide
- `DEPLOYMENT_FIX_VERIFICATION.md` - Deployment verification steps

