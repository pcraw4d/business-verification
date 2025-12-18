# Railway Snapshot Timeout Fix

## Problem

Railway deployments for `embedding-service` and `pipeline-service` were failing with:
```
Repository snapshot operation timed out. This may be due to a large repository size or network issues.
```

## Root Cause

Railway creates a snapshot of the **entire git repository** (including all history) before applying `.railwayignore`. The repository was 1.11 GiB packed size, with large directories tracked in git:
- `archive/` directory: 119MB, 179 files
- `reports/` directory: 23MB
- `test/results/`: 228KB
- `docs/railway log/`: 1.1MB

The `.railwayignore` file only works **after** Railway creates the snapshot, so it doesn't help with the snapshot timeout issue.

## Solution

### 1. Remove Large Directories from Git Tracking

Removed the following directories from git tracking (but kept them locally):
- `archive/` - Legacy UI files (119MB)
- `reports/` - Build reports (23MB)
- `test/results/` - Test output files
- `load_test_results/` - Load test results
- `docs/railway log/` - Railway log files

**Command used:**
```bash
git rm -r --cached archive/ reports/ test/results/ load_test_results/ "docs/railway log/"
```

### 2. Update .gitignore

Added these directories to `.gitignore` to prevent them from being tracked again:
```
archive/
reports/
test/results/
load_test_results/
docs/railway log/
```

### 3. Update .railwayignore

Enhanced `.railwayignore` to exclude additional large directories and files that shouldn't be deployed.

## Impact

- **199 files removed** from git tracking
- **195,276 lines deleted** from repository
- Repository snapshot size should be significantly reduced
- Files still exist locally for reference

## Next Steps

1. **Wait for Railway to retry**: Railway will automatically retry deployments after detecting the new commit
2. **Monitor deployments**: Check if `embedding-service` and `pipeline-service` deploy successfully
3. **If still timing out**: Consider setting service-specific root directories in Railway dashboard:
   - Go to Railway Dashboard → Service Settings
   - Set **Root Directory** to the service directory (e.g., `services/embedding-service` or `cmd/pipeline-service`)
   - This limits Railway to only snapshot the service directory

## Service-Specific Root Directory Configuration

### embedding-service
- **Root Directory**: `services/embedding-service`
- **Dockerfile Path**: `Dockerfile`
- **Docker Context**: `.` (service directory)

### pipeline-service
- **Root Directory**: `cmd/pipeline-service`
- **Dockerfile Path**: `Dockerfile`
- **Docker Context**: `.` (service directory)

Setting root directories will further reduce snapshot size by only including service-specific files.

## Verification

After the fix:
- Railway should create snapshots faster
- Deployments should complete without timeout errors
- Repository size in git should be reduced (though old history still exists in git packs)

## Notes

- Files removed from git tracking still exist locally
- They are now ignored by git and won't be committed again
- Railway will only clone the latest commit, which doesn't include these files
- If Railway clones with full history, it may still be slow, but the current commit is much smaller

