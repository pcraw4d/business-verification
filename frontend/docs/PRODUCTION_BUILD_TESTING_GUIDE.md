# Production Build & Hydration Testing Guide

This guide covers testing the production build for hydration errors (React Error #418) and cross-browser compatibility.

## Overview

After implementing Phase 3 fixes for hydration mismatches, we need to verify:
1. Production build completes successfully
2. No hydration errors in production mode
3. Cross-browser compatibility (Chrome, Firefox, Safari)
4. All date/number formatting works correctly

## Prerequisites

```bash
# Install dependencies
cd frontend
npm install

# Install Playwright browsers
npx playwright install
```

## Quick Start

### Automated Testing

Run the automated production build and hydration test suite:

```bash
cd frontend
node scripts/test-hydration-production.js
```

This script will:
1. Clean previous builds
2. Build production bundle
3. Start production server
4. Run Playwright tests across Chrome, Firefox, and Safari
5. Report results

### Manual Testing

#### 1. Build Production Bundle

```bash
cd frontend
npm run build
```

Expected output:
- Build completes without errors
- `.next` directory created with production build
- No TypeScript errors
- No build warnings

#### 2. Start Production Server

```bash
npm run start
```

Server should start on `http://localhost:3000`

#### 3. Test in Browsers

Open the following URLs in each browser and check the console:

**Chrome:**
1. Open `http://localhost:3000/merchant-details/[merchant-id]`
2. Open DevTools (F12)
3. Check Console tab for errors
4. Look for any "hydration" or "Text content does not match" errors
5. Verify dates and numbers display correctly

**Firefox:**
1. Same steps as Chrome
2. Use Firefox DevTools (F12)
3. Check Console for hydration errors

**Safari:**
1. Enable Develop menu: Preferences → Advanced → Show Develop menu
2. Open `http://localhost:3000/merchant-details/[merchant-id]`
3. Develop → Show Web Inspector
4. Check Console for errors

#### 4. Test Specific Components

Navigate to merchant details page and test:

**Merchant Overview Tab:**
- [ ] Created date displays correctly
- [ ] Updated date displays correctly
- [ ] Founded date displays correctly
- [ ] Employee count displays with commas
- [ ] Annual revenue displays as currency
- [ ] No "Loading..." text after page loads

**Business Analytics Tab:**
- [ ] Employee count from analytics displays correctly
- [ ] Annual revenue from analytics displays correctly
- [ ] Comparison notes display correctly
- [ ] No hydration warnings in console

**Risk Assessment Tab:**
- [ ] Assessment dates display correctly
- [ ] Risk history dates display correctly
- [ ] Chart data renders without errors
- [ ] No hydration warnings

**Risk Indicators Tab:**
- [ ] Alert dates display correctly
- [ ] Indicator dates display correctly
- [ ] No hydration warnings

**Portfolio Comparison Card:**
- [ ] Portfolio size displays with commas
- [ ] Risk scores display correctly
- [ ] No hydration warnings

## What to Look For

### ✅ Success Indicators

- No console errors or warnings
- All dates display correctly (not "Loading...")
- All numbers formatted correctly (commas, currency)
- Page loads without flickering
- No React hydration warnings
- Server-rendered HTML matches client-rendered HTML

### ❌ Failure Indicators

- Console errors mentioning "hydration"
- "Text content does not match" errors
- Dates showing "Loading..." permanently
- Numbers not formatted (raw numbers)
- Page content flickering on load
- React DevTools showing hydration warnings

## Browser-Specific Testing

### Chrome

```bash
# Run Chrome-specific tests
npx playwright test tests/e2e/hydration.spec.ts --project=chromium
```

### Firefox

```bash
# Run Firefox-specific tests
npx playwright test tests/e2e/hydration.spec.ts --project=firefox
```

### Safari (WebKit)

```bash
# Run Safari-specific tests
npx playwright test tests/e2e/hydration.spec.ts --project=webkit
```

## Debugging Hydration Issues

If you encounter hydration errors:

1. **Check Console Logs**
   - Look for specific error messages
   - Note which component is causing the issue

2. **Inspect HTML**
   - Compare server-rendered HTML (view source)
   - Compare client-rendered HTML (DevTools Elements)
   - Look for differences in date/number formatting

3. **Check Component State**
   - Verify `mounted` state is set correctly
   - Check that `useEffect` hooks run properly
   - Ensure `suppressHydrationWarning` is on correct elements

4. **Review Fixes**
   - Check that all date formatting uses `useState` + `useEffect`
   - Verify `suppressHydrationWarning` is added
   - Ensure no direct date/number formatting in render

## Test Results

After running tests, check:

```bash
# View Playwright report
npx playwright show-report
```

The report will show:
- Test results for each browser
- Screenshots of failures
- Console logs
- Network requests

## Continuous Integration

For CI/CD pipelines, add:

```yaml
- name: Build production
  run: cd frontend && npm run build

- name: Test hydration
  run: cd frontend && node scripts/test-hydration-production.js
```

## Troubleshooting

### Build Fails

- Check TypeScript errors: `npm run type-check`
- Verify environment variables are set
- Check Next.js config

### Server Won't Start

- Check if port 3000 is available
- Verify build completed successfully
- Check server logs

### Tests Fail

- Verify server is running
- Check Playwright browser installation
- Review test logs in `playwright-report/`

## Next Steps

After successful testing:

1. ✅ Production build works
2. ✅ No hydration errors
3. ✅ Cross-browser compatible
4. → Ready for Phase 4: Add Missing API Integrations

## References

- [Next.js Production Deployment](https://nextjs.org/docs/deployment)
- [React Hydration Errors](https://react.dev/reference/react-dom/client/hydrateRoot#handling-different-client-and-server-content)
- [Playwright Testing](https://playwright.dev/)

