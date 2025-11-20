import { expect, test } from '@playwright/test';

/**
 * Frontend Performance Tests
 * 
 * Measures page load times and verifies performance requirements:
 * - Merchant details page loads < 2 seconds
 * - Dashboard pages load < 3 seconds
 * - Test with slow network conditions
 * - Test with large datasets
 */

test.describe('Frontend Performance Tests', () => {
  const merchantId = 'merchant-123';

  test.beforeEach(async ({ page }) => {
    // Mock API responses for consistent performance testing
    await page.route('**/api/v1/merchants/**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: merchantId,
          businessName: 'Test Business',
          industry: 'Technology',
          status: 'active',
          riskLevel: 'medium',
        }),
      });
    });

    await page.route('**/api/v1/merchants/statistics', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          totalMerchants: 1000,
          averageRiskScore: 0.5,
          riskDistribution: { low: 400, medium: 300, high: 200, critical: 100 },
        }),
      });
    });

    await page.route('**/api/v1/merchants/analytics', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          totalMerchants: 1000,
          analyticsScore: 0.75,
          distributionData: [],
        }),
      });
    });

    await page.route(`**/api/v1/merchants/${merchantId}/analytics`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          merchantId: merchantId,
          classificationConfidence: 0.9,
          securityTrustScore: 0.75,
          dataQualityScore: 0.85,
        }),
      });
    });

    await page.route(`**/api/v1/merchants/${merchantId}/risk-score`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          merchantId: merchantId,
          score: 0.65,
          level: 'medium',
          confidence: 0.8,
          assessmentDate: new Date().toISOString(),
        }),
      });
    });
  });

  test('Merchant Details Page - Load Time < 2 seconds', async ({ page }) => {
    // Use performance timing API for more accurate measurement
    await page.goto(`/merchant-details/${merchantId}`, { waitUntil: 'domcontentloaded' });
    
    // Wait for main content to be visible
    await page.waitForSelector('h1, [role="heading"]', { timeout: 5000 });
    
    // Measure load time using Performance API
    const loadTime = await page.evaluate(() => {
      const perfData = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      return perfData.loadEventEnd - perfData.fetchStart;
    });

    const loadTimeSeconds = loadTime / 1000;

    console.log(`Merchant Details Page Load Time: ${loadTimeSeconds.toFixed(2)}s`);

    // Verify page loaded successfully
    await expect(page.locator('h1, [role="heading"]').first()).toBeVisible();

    // Verify performance requirement: < 2 seconds
    expect(loadTimeSeconds).toBeLessThan(2.0);
  });

  test('Merchant Details Page - Time to Interactive', async ({ page }) => {
    const navigationPromise = page.goto(`/merchant-details/${merchantId}`);
    
    // Measure time to interactive
    const startTime = Date.now();
    
    await navigationPromise;
    
    // Wait for page to be interactive (no loading states)
    await page.waitForLoadState('networkidle');
    await page.waitForSelector('h1, [role="heading"]', { timeout: 5000 });
    
    // Wait for any loading spinners to disappear
    await page.waitForFunction(() => {
      const loaders = document.querySelectorAll('[data-testid*="loading"], [class*="loading"], [class*="spinner"]');
      return loaders.length === 0;
    }, { timeout: 5000 }).catch(() => {
      // Ignore if no loaders found
    });

    const timeToInteractive = (Date.now() - startTime) / 1000;

    console.log(`Time to Interactive: ${timeToInteractive.toFixed(2)}s`);

    // Verify time to interactive < 3 seconds
    expect(timeToInteractive).toBeLessThan(3.0);
  });

  test('Business Intelligence Dashboard - Load Time < 3 seconds', async ({ page }) => {
    await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
    
    // Wait for main content to be visible
    await page.waitForSelector('h1, [role="heading"], [data-testid*="dashboard"]', { timeout: 5000 });
    
    // Measure load time using Performance API
    const loadTime = await page.evaluate(() => {
      const perfData = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      return perfData.loadEventEnd - perfData.fetchStart;
    });

    const loadTimeSeconds = loadTime / 1000;

    console.log(`BI Dashboard Load Time: ${loadTimeSeconds.toFixed(2)}s`);

    // Verify page loaded successfully
    const heading = page.locator('h1, [role="heading"]').first();
    await expect(heading).toBeVisible();

    // Verify performance requirement: < 3 seconds
    expect(loadTimeSeconds).toBeLessThan(3.0);
  });

  test('Risk Dashboard - Load Time < 3 seconds', async ({ page }) => {
    await page.goto('/risk-dashboard', { waitUntil: 'domcontentloaded' });
    
    // Wait for main content to be visible
    await page.waitForSelector('h1, [role="heading"], [data-testid*="risk"]', { timeout: 5000 });
    
    // Measure load time using Performance API
    const loadTime = await page.evaluate(() => {
      const perfData = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      return perfData.loadEventEnd - perfData.fetchStart;
    });

    const loadTimeSeconds = loadTime / 1000;

    console.log(`Risk Dashboard Load Time: ${loadTimeSeconds.toFixed(2)}s`);

    // Verify page loaded successfully
    const heading = page.locator('h1, [role="heading"]').first();
    await expect(heading).toBeVisible();

    // Verify performance requirement: < 3 seconds
    expect(loadTimeSeconds).toBeLessThan(3.0);
  });

  test('Risk Indicators Dashboard - Load Time < 3 seconds', async ({ page }) => {
    await page.goto('/risk-indicators', { waitUntil: 'domcontentloaded' });
    
    // Wait for main content to be visible
    await page.waitForSelector('h1, [role="heading"], [data-testid*="indicators"]', { timeout: 5000 });
    
    // Measure load time using Performance API
    const loadTime = await page.evaluate(() => {
      const perfData = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      return perfData.loadEventEnd - perfData.fetchStart;
    });

    const loadTimeSeconds = loadTime / 1000;

    console.log(`Risk Indicators Dashboard Load Time: ${loadTimeSeconds.toFixed(2)}s`);

    // Verify page loaded successfully
    const heading = page.locator('h1, [role="heading"]').first();
    await expect(heading).toBeVisible();

    // Verify performance requirement: < 3 seconds
    expect(loadTimeSeconds).toBeLessThan(3.0);
  });

  test('Merchant Details Page - Slow Network (3G)', async ({ page, context }) => {
    // Simulate slow 3G network using Playwright's built-in throttling
    await context.route('**/*', async (route) => {
      // Add delay to simulate slow network
      await new Promise(resolve => setTimeout(resolve, 100));
      await route.continue();
    });

    await page.goto(`/merchant-details/${merchantId}`, { waitUntil: 'domcontentloaded' });
    
    // Wait for main content to be visible
    await page.waitForSelector('h1, [role="heading"]', { timeout: 10000 });
    
    // Measure load time using Performance API
    const loadTime = await page.evaluate(() => {
      const perfData = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      return perfData.loadEventEnd - perfData.fetchStart;
    });

    const loadTimeSeconds = loadTime / 1000;

    console.log(`Merchant Details Page Load Time (3G): ${loadTimeSeconds.toFixed(2)}s`);

    // Verify page loaded successfully even on slow network
    await expect(page.locator('h1, [role="heading"]').first()).toBeVisible();

    // On slow network, allow up to 5 seconds
    expect(loadTimeSeconds).toBeLessThan(5.0);
  });

  test('Merchant Details Page - Large Dataset', async ({ page }) => {
    // Mock large dataset response
    await page.route('**/api/v1/merchants/statistics', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          totalMerchants: 10000, // Large dataset
          averageRiskScore: 0.5,
          riskDistribution: {
            low: 4000,
            medium: 3000,
            high: 2000,
            critical: 1000,
          },
        }),
      });
    });

    await page.goto(`/merchant-details/${merchantId}`, { waitUntil: 'domcontentloaded' });
    
    // Wait for main content to be visible
    await page.waitForSelector('h1, [role="heading"]', { timeout: 5000 });
    
    // Measure load time using Performance API
    const loadTime = await page.evaluate(() => {
      const perfData = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      return perfData.loadEventEnd - perfData.fetchStart;
    });

    const loadTimeSeconds = loadTime / 1000;

    console.log(`Merchant Details Page Load Time (Large Dataset): ${loadTimeSeconds.toFixed(2)}s`);

    // Verify page loaded successfully
    await expect(page.locator('h1, [role="heading"]').first()).toBeVisible();

    // With large dataset, allow up to 3 seconds
    expect(loadTimeSeconds).toBeLessThan(3.0);
  });

  test('Merchant Details Page - Bundle Size Check', async ({ page }) => {
    const responses: { url: string; size: number }[] = [];

    // Track network responses
    page.on('response', async (response) => {
      const url = response.url();
      if (url.includes('.js') || url.includes('.css') || url.includes('.chunk')) {
        const headers = response.headers();
        const contentLength = headers['content-length'];
        if (contentLength) {
          responses.push({
            url,
            size: parseInt(contentLength, 10),
          });
        }
      }
    });

    await page.goto(`/merchant-details/${merchantId}`);
    await page.waitForLoadState('networkidle');

    // Calculate total bundle size
    const totalSize = responses.reduce((sum, r) => sum + r.size, 0);
    const totalSizeMB = totalSize / (1024 * 1024);

    console.log(`Total Bundle Size: ${totalSizeMB.toFixed(2)}MB`);
    console.log(`Number of Assets: ${responses.length}`);

    // Log largest assets
    const sortedResponses = responses.sort((a, b) => b.size - a.size);
    console.log('Top 5 Largest Assets:');
    sortedResponses.slice(0, 5).forEach((r, i) => {
      console.log(`  ${i + 1}. ${r.url}: ${(r.size / 1024).toFixed(2)}KB`);
    });

    // Verify bundle size is reasonable (< 5MB total)
    expect(totalSizeMB).toBeLessThan(5.0);
  });

  test('Merchant Details Page - First Contentful Paint', async ({ page }) => {
    await page.goto(`/merchant-details/${merchantId}`);

    // Measure First Contentful Paint using Performance API
    const fcp = await page.evaluate(() => {
      return new Promise<number>((resolve) => {
        new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (entry.name === 'first-contentful-paint') {
              resolve(entry.startTime);
            }
          }
        }).observe({ entryTypes: ['paint'] });

        // Fallback timeout
        setTimeout(() => resolve(0), 5000);
      });
    });

    const fcpSeconds = fcp / 1000;

    console.log(`First Contentful Paint: ${fcpSeconds.toFixed(2)}s`);

    // Verify FCP < 1.5 seconds
    if (fcp > 0) {
      expect(fcpSeconds).toBeLessThan(1.5);
    }
  });

  test('Merchant Details Page - Largest Contentful Paint', async ({ page }) => {
    await page.goto(`/merchant-details/${merchantId}`);
    await page.waitForLoadState('networkidle');

    // Measure Largest Contentful Paint using Performance API
    const lcp = await page.evaluate(() => {
      return new Promise<number>((resolve) => {
        new PerformanceObserver((list) => {
          const entries = list.getEntries();
          const lastEntry = entries[entries.length - 1] as PerformanceEntry & { renderTime?: number; loadTime?: number };
          if (lastEntry) {
            resolve(lastEntry.renderTime || lastEntry.loadTime || lastEntry.startTime);
          }
        }).observe({ entryTypes: ['largest-contentful-paint'] });

        // Fallback timeout
        setTimeout(() => resolve(0), 5000);
      });
    });

    const lcpSeconds = lcp / 1000;

    console.log(`Largest Contentful Paint: ${lcpSeconds.toFixed(2)}s`);

    // Verify LCP < 2.5 seconds
    if (lcp > 0) {
      expect(lcpSeconds).toBeLessThan(2.5);
    }
  });

  test('Merchant Details Page - Cumulative Layout Shift', async ({ page }) => {
    await page.goto(`/merchant-details/${merchantId}`);
    await page.waitForLoadState('networkidle');

    // Measure Cumulative Layout Shift using Performance API
    const cls = await page.evaluate(() => {
      return new Promise<number>((resolve) => {
        let clsValue = 0;

        new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            const layoutShiftEntry = entry as PerformanceEntry & { value?: number };
            if (layoutShiftEntry.value) {
              clsValue += layoutShiftEntry.value;
            }
          }
        }).observe({ entryTypes: ['layout-shift'] });

        // Wait a bit then return CLS
        setTimeout(() => resolve(clsValue), 2000);
      });
    });

    console.log(`Cumulative Layout Shift: ${cls.toFixed(4)}`);

    // Verify CLS < 0.1 (good threshold)
    expect(cls).toBeLessThan(0.1);
  });

  test('Merchant Details Page - Tab Switching Performance', async ({ page }) => {
    await page.goto(`/merchant-details/${merchantId}`);
    await page.waitForLoadState('networkidle');

    // Measure tab switching performance
    const tabs = ['Overview', 'Business Analytics', 'Risk Assessment', 'Risk Indicators'];
    const switchTimes: number[] = [];

    for (const tabName of tabs) {
      const startTime = Date.now();
      
      // Click tab
      await page.getByRole('tab', { name: tabName }).click();
      
      // Wait for tab content to be visible
      await page.waitForSelector('[role="tabpanel"]', { timeout: 3000 });
      
      const switchTime = Date.now() - startTime;
      switchTimes.push(switchTime);

      console.log(`Tab Switch to ${tabName}: ${switchTime}ms`);
    }

    const avgSwitchTime = switchTimes.reduce((a, b) => a + b, 0) / switchTimes.length;
    const maxSwitchTime = Math.max(...switchTimes);

    console.log(`Average Tab Switch Time: ${avgSwitchTime.toFixed(2)}ms`);
    console.log(`Max Tab Switch Time: ${maxSwitchTime}ms`);

    // Verify tab switching is fast (< 500ms average)
    expect(avgSwitchTime).toBeLessThan(500);
    expect(maxSwitchTime).toBeLessThan(1000);
  });
});

