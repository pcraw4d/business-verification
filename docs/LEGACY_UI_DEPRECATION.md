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

### Phase 3: Deprecation Warnings (Current)
- Deprecation banners added to legacy HTML pages
- Documentation updated
- Migration guides provided

### Phase 4: Legacy UI Removal (Future)
- Legacy files archived to `archive/legacy-ui/`
- Legacy files removed from deployment
- Documentation updated to reflect new UI only

## How to Enable New UI

Set the environment variable:
```bash
export USE_NEW_UI=true
# or
export NEXT_PUBLIC_USE_NEW_UI=true
```

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

### ⚠️ Legacy Pages (Deprecated)
All legacy HTML pages in `cmd/frontend-service/static/` are now deprecated and will be removed in Phase 4.

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

## Deprecation Notice Template

All legacy HTML pages now include a deprecation banner directing users to the new UI.

