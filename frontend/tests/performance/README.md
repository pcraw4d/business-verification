# Frontend Performance Testing

**Date:** 2025-01-27

## Overview

Comprehensive performance testing suite for the frontend that measures:
- Page load times
- Time to Interactive (TTI)
- First Contentful Paint (FCP)
- Largest Contentful Paint (LCP)
- Cumulative Layout Shift (CLS)
- Bundle sizes
- Performance under slow network conditions
- Performance with large datasets

## Performance Requirements

According to the implementation plan:
- **Merchant Details Page:** < 2 seconds load time
- **Dashboard Pages:** < 3 seconds load time
- **Time to Interactive:** < 3 seconds
- **First Contentful Paint:** < 1.5 seconds
- **Largest Contentful Paint:** < 2.5 seconds
- **Cumulative Layout Shift:** < 0.1
- **Bundle Size:** < 5MB total

## Test Structure

### Test Files

1. **`performance.spec.ts`** - Playwright performance tests
   - Page load time measurements
   - Time to Interactive
   - Slow network testing
   - Large dataset testing
   - Bundle size checks
   - Web Vitals (FCP, LCP, CLS)
   - Tab switching performance

2. **`lighthouse.config.js`** - Lighthouse configuration
   - Performance audits
   - Throttling settings
   - Screen emulation

## Running Tests

### Playwright Performance Tests

```bash
cd frontend
npm run test:e2e -- tests/e2e/performance.spec.ts
```

### Run Specific Performance Test

```bash
# Test merchant details page load time
npm run test:e2e -- tests/e2e/performance.spec.ts -g "Merchant Details Page - Load Time"

# Test slow network
npm run test:e2e -- tests/e2e/performance.spec.ts -g "Slow Network"

# Test bundle size
npm run test:e2e -- tests/e2e/performance.spec.ts -g "Bundle Size"
```

### Lighthouse Audits

```bash
# Start dev server
npm run dev

# Run Lighthouse audit
npx lighthouse http://localhost:3000/merchant-details/merchant-123 \
  --config-path=./tests/performance/lighthouse.config.js \
  --output=html \
  --output-path=./tests/performance/lighthouse-report.html
```

## Test Coverage

### Pages Tested

1. ✅ **Merchant Details Page**
   - Load time < 2 seconds
   - Time to Interactive < 3 seconds
   - FCP < 1.5 seconds
   - LCP < 2.5 seconds
   - CLS < 0.1
   - Tab switching < 500ms average

2. ✅ **Business Intelligence Dashboard**
   - Load time < 3 seconds

3. ✅ **Risk Dashboard**
   - Load time < 3 seconds

4. ✅ **Risk Indicators Dashboard**
   - Load time < 3 seconds

### Performance Scenarios

1. ✅ **Normal Network**
   - Standard page load times
   - Baseline performance metrics

2. ✅ **Slow Network (3G)**
   - Simulated 3G network conditions
   - Tests performance under slow connections

3. ✅ **Large Dataset**
   - Tests with 10,000+ merchants
   - Verifies performance with large data volumes

4. ✅ **Bundle Size**
   - Measures total JavaScript/CSS bundle size
   - Identifies large assets
   - Verifies bundle size < 5MB

### Web Vitals Measured

1. ✅ **First Contentful Paint (FCP)**
   - Time until first content is rendered
   - Target: < 1.5 seconds

2. ✅ **Largest Contentful Paint (LCP)**
   - Time until largest content element is rendered
   - Target: < 2.5 seconds

3. ✅ **Cumulative Layout Shift (CLS)**
   - Measures visual stability
   - Target: < 0.1

4. ✅ **Time to Interactive (TTI)**
   - Time until page is fully interactive
   - Target: < 3 seconds

## Performance Optimization Recommendations

### If Load Time > Target

1. **Code Splitting**
   - Implement route-based code splitting
   - Lazy load components
   - Split vendor bundles

2. **Image Optimization**
   - Use Next.js Image component
   - Implement lazy loading
   - Use WebP format

3. **API Optimization**
   - Reduce API calls
   - Implement request batching
   - Use caching effectively

4. **Bundle Optimization**
   - Remove unused dependencies
   - Tree-shake unused code
   - Minimize bundle size

### If Bundle Size > 5MB

1. **Analyze Bundle**
   ```bash
   npm run build
   npx @next/bundle-analyzer
   ```

2. **Optimize Dependencies**
   - Remove large unused libraries
   - Use lighter alternatives
   - Split large dependencies

3. **Code Splitting**
   - Implement dynamic imports
   - Split routes
   - Lazy load heavy components

### If CLS > 0.1

1. **Image Dimensions**
   - Set explicit width/height
   - Use aspect ratio boxes
   - Reserve space for images

2. **Font Loading**
   - Preload critical fonts
   - Use font-display: swap
   - Avoid invisible text

3. **Dynamic Content**
   - Reserve space for dynamic content
   - Avoid inserting content above existing content
   - Use skeleton loaders

## Continuous Monitoring

### Recommended Metrics to Track

1. **Page Load Times**
   - Track per page
   - Alert if > target
   - Monitor trends

2. **Web Vitals**
   - FCP, LCP, CLS
   - Track in production
   - Set up alerts

3. **Bundle Sizes**
   - Track bundle size over time
   - Alert on size increases
   - Monitor for regressions

4. **Error Rates**
   - Track performance-related errors
   - Monitor slow page loads
   - Alert on degradation

## Test Execution Examples

### Example 1: Quick Performance Check

```bash
# Run all performance tests
npm run test:e2e -- tests/e2e/performance.spec.ts
```

### Example 2: Test Specific Page

```bash
# Test merchant details page
npm run test:e2e -- tests/e2e/performance.spec.ts -g "Merchant Details"
```

### Example 3: Lighthouse Audit

```bash
# Run Lighthouse audit
npx lighthouse http://localhost:3000/merchant-details/merchant-123 \
  --config-path=./tests/performance/lighthouse.config.js \
  --output=json \
  --output-path=./tests/performance/lighthouse-report.json
```

## Next Steps

1. ✅ Performance tests created
2. Run tests against live frontend
3. Document baseline performance metrics
4. Identify and optimize slow pages
5. Set up continuous performance monitoring

## Files Created

1. **`performance.spec.ts`** - Playwright performance tests
2. **`lighthouse.config.js`** - Lighthouse configuration
3. **`README.md`** - This documentation

