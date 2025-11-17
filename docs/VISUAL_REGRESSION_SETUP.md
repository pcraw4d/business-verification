# Visual Regression Testing Setup

## Overview

Visual regression testing captures screenshots of pages and compares them against baseline images to detect unintended visual changes.

## Setup

Visual regression tests are configured using Playwright's built-in screenshot comparison feature.

### Test Location
- Tests: `frontend/tests/visual/visual-regression.spec.ts`
- Snapshots: `frontend/tests/visual/snapshots/` (auto-generated)

## Running Tests

### Run Visual Tests
```bash
npm run test:visual
```

### Update Baselines
When intentional visual changes are made, update the baseline snapshots:
```bash
npm run test:visual:update
```

This will:
1. Capture new screenshots
2. Save them as baseline images
3. Future test runs will compare against these new baselines

## Test Coverage

Current visual tests cover:
- ✅ Home page
- ✅ Dashboard hub
- ✅ Merchant portfolio
- ✅ Risk dashboard
- ✅ Compliance page
- ✅ Add merchant page
- ✅ Admin page
- ✅ Mobile viewport (375x667)
- ✅ Tablet viewport (768x1024)

## Configuration

### Viewport Sizes
- **Desktop**: 1280x720 (default)
- **Mobile**: 375x667 (iPhone SE)
- **Tablet**: 768x1024 (iPad)

### Screenshot Settings
- **Full page**: Yes (captures entire scrollable page)
- **Max diff pixels**: 100 (allows small differences)
- **Wait time**: 1 second after load (for animations)

## CI/CD Integration

### GitHub Actions Example
```yaml
- name: Run Visual Regression Tests
  run: |
    npm run build
    npm run start &
    sleep 10
    npm run test:visual
```

### Handling Failures
When visual tests fail:
1. Review the diff images in `tests/visual/snapshots/`
2. Determine if changes are intentional or bugs
3. If intentional: Update baselines with `npm run test:visual:update`
4. If bug: Fix the issue and re-run tests

## Best Practices

1. **Update baselines after intentional changes**
   - UI updates
   - Design system changes
   - Component library updates

2. **Review diffs carefully**
   - Small pixel differences may be acceptable
   - Large differences indicate real issues

3. **Test multiple viewports**
   - Desktop, tablet, mobile
   - Different screen sizes catch responsive issues

4. **Keep baselines in version control**
   - Commit updated snapshots
   - Review snapshot changes in PRs

## Troubleshooting

### Tests fail with "screenshot mismatch"
- Review the diff image
- If intentional: Update baseline
- If bug: Fix the issue

### Screenshots look different on CI
- Ensure consistent viewport sizes
- Check for font rendering differences
- Verify image assets are loaded

### Flaky tests
- Increase wait times
- Add explicit waits for dynamic content
- Check for animations that complete at different times

## Next Steps

1. ✅ Visual regression tests created
2. ⚠️ Run initial baseline capture
3. ⚠️ Integrate into CI/CD pipeline
4. ⚠️ Add more page coverage
5. ⚠️ Test dark mode variations

---

**Last Updated**: 2025-01-XX

