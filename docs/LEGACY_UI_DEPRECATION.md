# Legacy UI Deprecation Notice

## Overview

This document outlines the deprecation plan for the legacy HTML/CSS/JS UI in favor of the new shadcn UI implementation with Next.js.

## Deprecation Timeline

### Phase 1: Parallel Operation (Completed)
- Both old and new UI run in parallel
- Feature flag `NEXT_PUBLIC_USE_NEW_UI` or `USE_NEW_UI` controls routing
- Default: Legacy UI (backward compatibility)

### Phase 2: Gradual Migration (In Progress)
- New UI pages are enabled as they are migrated
- Monitoring and feedback collection
- Bug fixes and improvements

### Phase 3: Deprecation Warnings (✅ COMPLETE)
- ✅ Deprecation banners added to all legacy HTML pages (66/66 files)
- ✅ Empty files fixed and banners added
- ✅ Documentation updated
- ✅ Migration guides provided

### Phase 4: Legacy UI Removal (✅ COMPLETE)
- ✅ Legacy files archived to `archive/legacy-ui/20251117_011146/`
- ✅ Legacy files removed from deployment (33 HTML files + directories)
- ✅ Documentation updated to reflect new UI only
- ✅ Routing updated to remove legacy fallback

## New UI Status

**New UI is now the default** - no environment variables needed.

The legacy UI has been completely removed. If you need to access archived legacy files, they are available in `archive/legacy-ui/20251117_011146/`.

## Migration Status

### ✅ Fully Migrated Pages
- Landing page (`/`)
- Add Merchant (`/add-merchant`)
- Merchant Portfolio (`/merchant-portfolio`)
- Dashboard Hub (`/dashboard-hub`)
- Register (`/register`)
- Dashboard (`/dashboard`)
- Risk Dashboard (`/risk-dashboard`)
- Risk Indicators (`/risk-indicators`)
- Compliance (`/compliance`)
- Admin Dashboard (`/admin`)
- Merchant Hub (`/merchant-hub`)
- Business Intelligence (`/business-intelligence`)
- Monitoring (`/monitoring`)
- All Compliance sub-pages
- All Merchant Management pages
- All Market Intelligence pages
- All Admin pages
- Utility pages

### ✅ Legacy Pages (Removed)
All legacy HTML pages have been removed in Phase 4. Files archived to `archive/legacy-ui/20251117_011146/`

## Migration Guide

### For Developers
1. Use the new Next.js routes in `frontend/app/`
2. Use shadcn UI components from `frontend/components/ui/`
3. Follow the new component patterns in `frontend/components/`

### For Users
- The new UI provides the same functionality with improved design
- All features are preserved
- Performance improvements expected

## Support

For questions or issues during migration, please refer to:
- `docs/FRONTEND_UI_AUDIT_REPORT.md` - Complete UI audit
- `frontend/README.md` - Next.js setup guide
- Component documentation in `frontend/components/`

## Phase 4 Complete

✅ **Legacy UI has been completely removed**
- All legacy files archived to `archive/legacy-ui/20251117_011146/`
- Legacy files removed from deployment directories
- Routing updated to serve only Next.js UI
- Documentation updated

The platform now uses only the new Next.js UI with shadcn components.

