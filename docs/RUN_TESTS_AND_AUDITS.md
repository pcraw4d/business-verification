# How to Run Tests and Audits

**Date**: 2025-01-XX

## Quick Start

### 1. E2E Tests ✅

**Auto-starts dev server** - No manual setup needed!

```bash
npm run test:e2e
```

This will:
- Auto-start the dev server if not running
- Run 165 tests across multiple browsers
- Generate HTML report
- Show results in terminal

**View Report**:
```bash
npx playwright show-report
```

**Run Specific Test Suite**:
```bash
npm run test:e2e:navigation
npm run test:e2e:forms
npm run test:e2e:data
npm run test:e2e:export
npm run test:e2e:bulk
```

---

### 2. Visual Regression Tests ⚠️

**Requires dev server running**

**Step 1: Start Dev Server**
```bash
npm run dev
```

**Step 2: Capture Baselines** (First time only)
```bash
# In another terminal
npm run test:visual:update
```

This will:
- Capture screenshots of all pages
- Save them as baseline images in `tests/visual/snapshots/`
- These become the reference for future tests

**Step 3: Run Visual Tests** (After baselines captured)
```bash
npm run test:visual
```

This will:
- Compare current screenshots with baselines
- Report any differences
- Fail if differences exceed threshold

**Update Baselines** (When intentional changes made):
```bash
npm run test:visual:update
```

---

### 3. Lighthouse Audit ⚠️

**Requires production server running**

**Step 1: Build Production Bundle**
```bash
npm run build
```

**Step 2: Start Production Server**
```bash
npm run start
```

**Step 3: Run Lighthouse** (In another terminal)
```bash
npm run lighthouse
```

This will:
- Open browser
- Run Lighthouse audit
- Generate HTML report
- Open report automatically

**Alternative: Lighthouse CI**
```bash
lhci autorun
```

---

### 4. Bundle Analysis ✅

**No server needed**

```bash
# Build first
npm run build

# Analyze bundle
npm run analyze-bundle
```

**Output**:
- Total bundle sizes
- Largest chunks
- Chunk analysis
- Recommendations

---

### 5. Performance Audit ✅

**No server needed**

```bash
# Build first
npm run build

# Run audit
npm run performance-audit
```

**Output**:
- Build manifest analysis
- Route analysis
- Optimization checks

---

### 6. Accessibility Audit ✅

**No server needed**

```bash
npm run accessibility-audit
```

**Output**:
- List of files with potential issues
- WCAG rule references
- Recommendations

---

### 7. Run All Audits ✅

**No server needed**

```bash
# Build first
npm run build

# Run all audits
npm run audit:all
```

This runs:
- Bundle analysis
- Performance audit
- Accessibility audit

---

## Test Results Summary

### E2E Tests
- **Status**: ✅ Ready
- **Total**: 165 tests
- **Browsers**: Chrome, Firefox, Safari, Mobile Chrome, Mobile Safari
- **Auto-start**: Dev server auto-starts

### Visual Regression
- **Status**: ⚠️ Ready (needs baseline capture)
- **Coverage**: 9 scenarios
- **Viewports**: Desktop, Mobile, Tablet

### Bundle Analysis
- **Status**: ✅ Complete
- **Total Size**: 3.82 MB (3.75 MB JS + 65.86 KB CSS)
- **Largest Chunk**: 407.75 KB
- **Optimizations**: Lazy loading implemented

### Performance Audit
- **Status**: ✅ Complete
- **Routes**: 153 routes detected
- **Optimizations**: Code splitting, image optimization configured

### Accessibility Audit
- **Status**: ✅ Complete
- **Issues Found**: 16 files (mostly warnings)
- **Critical**: All critical issues addressed

---

## Troubleshooting

### E2E Tests Fail
- Check if dev server is accessible at `http://localhost:3000`
- Check browser installation (Playwright installs browsers automatically)
- Review test output for specific errors

### Visual Tests Fail
- Ensure baselines are captured first
- Check viewport sizes match
- Review diff images in `tests/visual/snapshots/`

### Lighthouse Fails
- Ensure production server is running
- Check server is accessible at configured URL
- Verify build completed successfully

### Bundle Analysis Shows 0 Bytes for Chunks
- This is normal for Turbopack builds
- Chunks may be named differently
- Check `.next/static/chunks/` directory manually

---

## Next Steps

1. **Run E2E Tests**: `npm run test:e2e`
2. **Capture Visual Baselines**: `npm run dev` then `npm run test:visual:update`
3. **Run Lighthouse**: `npm run build && npm run start` then `npm run lighthouse`
4. **Review Results**: Check all reports and fix any issues

---

**Last Updated**: 2025-01-XX

