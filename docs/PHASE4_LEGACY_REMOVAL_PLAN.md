# Phase 4: Legacy UI Removal Plan

**Date**: 2025-01-17  
**Status**: ✅ **COMPLETE**  
**Phase**: 4 of 4 (Deprecation Timeline)

## Overview

This document outlines the plan for completely removing legacy UI files after the deprecation period. This is the final phase of the UI migration.

## Prerequisites

Before proceeding with Phase 4, ensure:

- ✅ **Phase 3 Complete**: All deprecation banners added
- ✅ **New UI Verified**: All pages working in production
- ✅ **Usage Monitoring**: No active users on legacy UI
- ✅ **Archive Created**: Legacy files backed up
- ✅ **Documentation Updated**: Migration guides complete

## Removal Strategy

### Step 1: Archive Legacy Files (Week 1)

**Action**: Create backup archive of all legacy files

```bash
# Dry run first
./scripts/archive-legacy-ui.sh --dry-run

# Create archive
./scripts/archive-legacy-ui.sh
```

**Archive Location**: `archive/legacy-ui/{timestamp}/`

**Contents**:
- All HTML files
- JavaScript files (`js/`)
- CSS files (`css/`)
- Component files (`components/`)

**Verification**:
- Check archive contents
- Verify file counts match
- Test archive integrity

### Step 2: Update Routing (Week 1)

**Action**: Remove legacy UI fallback from routing

**Files to Modify**:
- `cmd/frontend-service/routing.go`
- Remove legacy path fallback logic
- Default to new UI only

**Changes**:
```go
// Remove this fallback:
// Fall back to legacy UI
legacyPath := rc.getLegacyPath(route)
if _, err := os.Stat(legacyPath); err == nil {
    http.ServeFile(w, r, legacyPath)
    return
}
```

### Step 3: Remove Legacy Files (Week 2)

**Action**: Delete legacy files from deployment directories

**Directories to Clean**:
- `cmd/frontend-service/static/*.html` (except `.next/`)
- `cmd/frontend-service/static/js/` (if not used by new UI)
- `cmd/frontend-service/static/css/` (if not used by new UI)
- `cmd/frontend-service/static/components/` (if not used by new UI)
- `services/frontend/public/*.html` (if not needed)

**Script**:
```bash
# Review what will be removed
./scripts/remove-legacy-files.sh --dry-run

# Remove files
./scripts/remove-legacy-files.sh
```

### Step 4: Update Documentation (Week 2)

**Action**: Update all documentation to reflect new UI only

**Files to Update**:
- `README.md` - Remove legacy UI references
- `docs/LEGACY_UI_DEPRECATION.md` - Mark Phase 4 complete
- `docs/FRONTEND_UI_AUDIT_REPORT.md` - Update status
- API documentation - Remove legacy endpoints
- Deployment guides - Remove legacy configuration

### Step 5: Final Verification (Week 2)

**Action**: Comprehensive testing and verification

**Checklist**:
- [ ] All pages load correctly
- [ ] No 404 errors for migrated routes
- [ ] API integration works
- [ ] Forms submit correctly
- [ ] Charts render properly
- [ ] Navigation works
- [ ] Mobile responsive
- [ ] Performance metrics acceptable
- [ ] No console errors
- [ ] E2E tests pass

## Rollback Plan

If issues are discovered:

1. **Immediate**: Restore files from archive
   ```bash
   cp -r archive/legacy-ui/{timestamp}/* cmd/frontend-service/static/
   ```

2. **Revert Routing**: Restore legacy fallback in `routing.go`

3. **Redeploy**: Deploy previous version

4. **Investigate**: Identify and fix issues

5. **Retry**: Attempt removal again after fixes

## Timeline

### Week 1: Archive & Preparation
- Day 1-2: Create archive, verify contents
- Day 3-4: Update routing, test thoroughly
- Day 5: Final review before removal

### Week 2: Removal & Verification
- Day 1-2: Remove legacy files
- Day 3-4: Update documentation
- Day 5: Final verification and testing

## Success Criteria

- ✅ All legacy files archived
- ✅ No legacy UI code in deployment
- ✅ All pages working with new UI
- ✅ Documentation updated
- ✅ No broken links or 404s
- ✅ Performance maintained or improved
- ✅ All tests passing

## Risks & Mitigation

### Risk 1: Missing Files
**Mitigation**: Comprehensive archive before removal

### Risk 2: Broken Routes
**Mitigation**: Thorough testing before and after removal

### Risk 3: Performance Regression
**Mitigation**: Monitor metrics, have rollback plan

### Risk 4: User Impact
**Mitigation**: Ensure new UI is fully functional, monitor errors

## Post-Removal Tasks

1. **Monitor**: Watch for errors or issues
2. **Optimize**: Remove unused dependencies
3. **Clean**: Remove archive after 30 days (if no issues)
4. **Document**: Update final migration status

## Scripts

- `scripts/archive-legacy-ui.sh` - Create archive
- `scripts/remove-legacy-files.sh` - Remove files (to be created)
- `scripts/verify-new-ui.sh` - Verify new UI (to be created)

## Notes

- Keep archive for at least 30 days
- Monitor error logs closely after removal
- Have rollback plan ready
- Communicate changes to team

---

**Status**: Ready for execution after Phase 3 completion and new UI verification

