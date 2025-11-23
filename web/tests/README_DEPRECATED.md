# ⚠️ DEPRECATED: Legacy Test Directory

**Status**: DEPRECATED - DO NOT USE  
**Date Deprecated**: 2025-01-19  
**Reason**: Legacy UI has been removed from production

## Overview

This directory contains test files for the **legacy HTML/CSS/JS UI** that has been archived and removed from production. These tests are **no longer relevant** and should **not be run** as part of normal testing workflows.

## Migration Status

- ✅ **Legacy UI**: Removed from production (Phase 4 complete)
- ✅ **New UI**: Next.js with shadcn UI is now the default
- ✅ **New Tests**: Located in `frontend/tests/e2e/`

## What Changed

### Legacy UI (Deprecated)
- **Location**: `web/` directory
- **Technology**: Static HTML files with vanilla JavaScript
- **Tests**: This directory (`web/tests/`)
- **Status**: Archived, not deployed

### New UI (Active)
- **Location**: `frontend/` directory
- **Technology**: Next.js with shadcn UI components
- **Tests**: `frontend/tests/e2e/`
- **Status**: Production default

## Test Scripts

### ❌ Deprecated Scripts (DO NOT USE)

All scripts prefixed with `test:legacy:` in the root `package.json` are deprecated:

```bash
# ❌ DO NOT USE - These test the legacy UI
npm run test:legacy:cross-browser
npm run test:legacy:state-based
npm run test:legacy:interactive
```

### ✅ Active Scripts (USE THESE)

Use these scripts to test the new frontend:

```bash
# ✅ Use these for new frontend tests
npm test                    # Run all E2E tests
npm run test:ui            # Run with UI mode
npm run test:headed        # Run in headed mode
npm run test:debug         # Run with debug
```

## Why These Tests Are Deprecated

1. **Legacy UI Removed**: The HTML/CSS/JS UI has been completely removed from production
2. **Wrong Implementation**: These tests target the old UI, not the current shadcn UI
3. **Outdated Selectors**: Test selectors reference legacy HTML structure
4. **Wrong URLs**: Tests use `.html` file extensions (legacy routing)
5. **Hardcoded Legacy URLs**: Many tests point to deprecated Railway deployments

## File Structure

### Legacy Test Files (This Directory)
- `merchant-*.spec.js` - Legacy merchant tests
- `visual/` - Visual regression tests for legacy UI
- `config/` - Legacy test configurations
- `scripts/` - Legacy test runner scripts
- `utils/` - Legacy test helpers

### New Test Files (Active)
- `frontend/tests/e2e/` - E2E tests for new UI
- `frontend/tests/visual/` - Visual tests for shadcn UI
- `frontend/playwright.config.ts` - Active test configuration

## Migration Guide

If you need to migrate test logic from here:

1. **Update Routes**: Change `.html` paths to Next.js routes
   - ❌ `/merchant-portfolio.html` 
   - ✅ `/merchant-portfolio`

2. **Update Selectors**: Use semantic selectors for shadcn UI
   - ❌ `document.querySelector('.legacy-class')`
   - ✅ `page.getByRole('button', { name: 'Submit' })`

3. **Update Base URL**: Use Next.js dev server
   - ❌ `http://localhost:8080`
   - ✅ `http://localhost:3000`

4. **Update Test Helpers**: Use Playwright's built-in helpers
   - ❌ Custom `navigateToDashboard()` with `.html` extension
   - ✅ `page.goto('/dashboard')` with Next.js routing

## When to Use Legacy Tests

**Never** - These tests are kept for reference only. If you need to:
- Review old test patterns → Check this directory
- Understand legacy UI structure → Check `archive/legacy-ui/`
- Run actual tests → Use `frontend/tests/e2e/`

## Cleanup Plan

These files will be removed in a future cleanup phase after:
- ✅ All new tests are verified working
- ✅ No references to legacy tests in CI/CD
- ✅ Team confirms no need for legacy test reference

## Questions?

- **New Test Issues**: Check `frontend/tests/e2e/`
- **Test Configuration**: See root `playwright.config.js`
- **Legacy UI Reference**: See `archive/legacy-ui/20251117_011146/`

---

**⚠️ Remember**: These tests are deprecated. Always use `npm test` which runs the new frontend tests.

