# GitHub Actions Workflows

**Date**: 2025-01-XX  
**Status**: Complete

## Summary

Created comprehensive GitHub Actions workflows for frontend testing, quality assurance, and CI/CD.

## Workflows Created

### 1. Frontend E2E Tests (`frontend-e2e.yml`)

**Purpose**: Run Playwright E2E tests on frontend changes.

**Triggers**:
- Push to `main` or `develop` branches (when `frontend/**` changes)
- Pull requests (when `frontend/**` changes)
- Manual dispatch

**Features**:
- Installs dependencies and Playwright browsers
- Builds production Next.js app
- Runs E2E tests
- Uploads test results and HTML reports
- Comments PR with test results

**Artifacts**:
- `playwright-report/` - HTML test report
- `test-results/` - Test execution results

### 2. Frontend Visual Regression Tests (`frontend-visual-regression.yml`)

**Purpose**: Run visual regression tests and update baselines.

**Triggers**:
- Push to `main` or `develop` branches (when `frontend/**` changes)
- Pull requests (when `frontend/**` changes)
- Manual dispatch

**Features**:
- Runs visual regression tests
- Updates baselines on main branch
- Uploads snapshots and test results
- Comments PR when visual differences detected

**Artifacts**:
- `visual-test-results/` - Test execution results
- `visual-snapshots/` - Screenshot snapshots

### 3. Lighthouse CI (`lighthouse-ci.yml`)

**Purpose**: Run Lighthouse audits and enforce performance budgets.

**Triggers**:
- Push to `main` branch (when `frontend/**` changes)
- Pull requests (when `frontend/**` changes)
- Manual dispatch

**Features**:
- Builds production app
- Runs Lighthouse CI with 3 runs for stability
- Enforces score thresholds:
  - Performance: ≥ 80
  - Accessibility: ≥ 90
  - Best Practices: ≥ 80
  - SEO: ≥ 80
- Comments PR with scores and key metrics
- Uploads Lighthouse reports

**Artifacts**:
- `lighthouse-reports/` - HTML and JSON reports

### 4. Bundle Size Analysis (`bundle-analysis.yml`)

**Purpose**: Analyze bundle sizes and monitor for regressions.

**Triggers**:
- Push to `main` branch (when `frontend/**` or `package*.json` changes)
- Pull requests (when `frontend/**` or `package*.json` changes)
- Manual dispatch

**Features**:
- Builds production app
- Runs bundle analysis
- Comments PR with bundle size information
- Uploads analysis artifacts

**Artifacts**:
- `bundle-analysis/` - Bundle analysis reports

### 5. Frontend CI/CD (`frontend-ci.yml`)

**Purpose**: Lint, test, and build frontend application.

**Triggers**:
- Push (when `frontend/**` changes)
- Pull requests (when `frontend/**` changes)

**Features**:
- Runs ESLint
- Runs unit tests with Vitest
- Uploads test coverage to Codecov
- Builds production app
- Uploads build artifacts

**Artifacts**:
- `frontend-build/` - Production build output

## Workflow Dependencies

```
frontend-ci.yml (lint, test, build)
    ↓
frontend-e2e.yml (E2E tests)
    ↓
frontend-visual-regression.yml (Visual tests)
    ↓
lighthouse-ci.yml (Performance audit)
    ↓
bundle-analysis.yml (Bundle size check)
```

## Score Thresholds

### Lighthouse
- **Performance**: ≥ 80 (error if below)
- **Accessibility**: ≥ 90 (error if below)
- **Best Practices**: ≥ 80 (error if below)
- **SEO**: ≥ 80 (error if below)

### Performance Metrics (Warnings)
- **First Contentful Paint**: ≤ 2000ms
- **Largest Contentful Paint**: ≤ 2500ms
- **Cumulative Layout Shift**: ≤ 0.1
- **Total Blocking Time**: ≤ 300ms

## PR Comments

All workflows that run on pull requests will automatically comment with:
- Test results summary
- Pass/fail status
- Links to detailed reports
- Recommendations for fixes

## Secrets Required

### Optional
- `LHCI_GITHUB_APP_TOKEN` - For Lighthouse CI server integration (optional)

## Usage

### Manual Trigger

All workflows can be manually triggered from the GitHub Actions tab:
1. Go to "Actions" in GitHub
2. Select the workflow
3. Click "Run workflow"
4. Choose branch and options

### Local Testing

Before pushing, test workflows locally:

```bash
# E2E tests
cd frontend && npm run test:e2e

# Visual regression
cd frontend && npm run test:visual

# Lighthouse
cd frontend && npm run lighthouse:ci

# Bundle analysis
cd frontend && npm run analyze-bundle
```

## Next Steps

1. ✅ Workflows created
2. ⏳ Test workflows on next PR
3. ⏳ Monitor workflow performance
4. ⏳ Adjust thresholds based on results
5. ⏳ Set up Lighthouse CI server (optional)

---

**Last Updated**: 2025-01-XX

