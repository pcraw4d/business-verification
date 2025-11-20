import { expect, test } from '@playwright/test';

/**
 * Console Error Detection Tests
 * 
 * Specifically tests for the "Cannot read properties of undefined (reading 'toFixed')" error
 * and other runtime errors that could crash the application.
 */

const TEST_MERCHANT_ID = 'merchant-123';

test.describe('Console Error Detection', () => {
  test('should not have toFixed errors when loading merchant details with partial data', async ({ page }) => {
    const consoleErrors: string[] = [];
    const pageErrors: Error[] = [];

    // Capture console errors
    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        const text = msg.text();
        consoleErrors.push(text);
        // Fail test if we see the specific toFixed error
        if (text.includes("Cannot read properties of undefined (reading 'toFixed')")) {
          throw new Error(`Found toFixed error: ${text}`);
        }
      }
    });

    // Capture page errors
    page.on('pageerror', (error) => {
      pageErrors.push(error);
      if (error.message.includes("Cannot read properties of undefined (reading 'toFixed')")) {
        throw new Error(`Found toFixed page error: ${error.message}`);
      }
    });

    // Mock merchant data with partial/missing numeric fields
    await page.route('**/api/v1/merchants/merchant-123**', async (route) => {
      const url = route.request().url();
      if (!url.includes('/analytics') && !url.includes('/risk') && !url.includes('/website')) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            id: TEST_MERCHANT_ID,
            businessName: 'Test Business',
            industry: 'Technology',
            status: 'active',
          }),
        });
      } else {
        await route.continue();
      }
    });

    // Mock risk score with missing fields
    await page.route('**/api/v1/merchants/*/risk-score**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          risk_score: undefined, // Missing risk_score
          confidence_score: null, // Null confidence_score
          factors: [
            {
              name: 'Factor 1',
              score: undefined, // Missing score
            },
          ],
        }),
      });
    });

    // Mock risk benchmarks with partial data
    await page.route('**/api/v1/risk/benchmarks**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          average_risk_score: undefined, // Missing average
          median_risk_score: null, // Null median
          percentile_25: undefined,
          percentile_75: null,
          percentile_90: undefined,
        }),
      });
    });

    // Mock portfolio statistics with missing fields
    await page.route('**/api/v1/merchants/statistics**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          averageRiskScore: undefined, // Missing average
          totalMerchants: 100,
        }),
      });
    });

    // Navigate to merchant details page
    await page.goto(`/merchant-details/${TEST_MERCHANT_ID}`);
    await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
    await page.waitForTimeout(3000);

    // Navigate through all tabs to trigger all components
    const tabs = [
      'Overview',
      'Business Analytics',
      'Risk Assessment',
      'Risk Indicators',
    ];

    for (const tabName of tabs) {
      try {
        const tab = page.getByRole('tab', { name: tabName });
        const isVisible = await tab.isVisible({ timeout: 5000 }).catch(() => false);
        if (isVisible) {
          await tab.click();
          await page.waitForTimeout(2000); // Wait for components to render
        }
      } catch (e) {
        // Tab might not exist, continue
      }
    }

    // Wait a bit more for any async operations
    await page.waitForTimeout(2000);

    // Check that no toFixed errors occurred
    const toFixedErrors = consoleErrors.filter((err) =>
      err.includes("Cannot read properties of undefined (reading 'toFixed')")
    );
    const toFixedPageErrors = pageErrors.filter((err) =>
      err.message.includes("Cannot read properties of undefined (reading 'toFixed')")
    );

    expect(toFixedErrors.length).toBe(0);
    expect(toFixedPageErrors.length).toBe(0);

    // Page should still be functional
    const pageContent = page.locator('main, [role="main"], body');
    await expect(pageContent.first()).toBeVisible();
  });

  test('should not have toFixed errors when loading dashboards with partial data', async ({ page }) => {
    const consoleErrors: string[] = [];
    const pageErrors: Error[] = [];

    // Capture console errors
    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        const text = msg.text();
        consoleErrors.push(text);
        if (text.includes("Cannot read properties of undefined (reading 'toFixed')")) {
          throw new Error(`Found toFixed error: ${text}`);
        }
      }
    });

    // Capture page errors
    page.on('pageerror', (error) => {
      pageErrors.push(error);
      if (error.message.includes("Cannot read properties of undefined (reading 'toFixed')")) {
        throw new Error(`Found toFixed page error: ${error.message}`);
      }
    });

    // Mock dashboard APIs with partial data
    await page.route('**/api/v1/merchants/statistics**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          totalMerchants: 100,
          averageRiskScore: undefined, // Missing
          growthRate: null, // Null
        }),
      });
    });

    await page.route('**/api/v1/risk/metrics**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          overallRiskScore: undefined, // Missing
          riskTrend: null, // Null
        }),
      });
    });

    // Test main dashboard
    await page.goto('/dashboard');
    await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
    await page.waitForTimeout(3000);

    // Test risk dashboard
    await page.goto('/risk-dashboard');
    await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
    await page.waitForTimeout(3000);

    // Test business intelligence dashboard
    await page.goto('/business-intelligence');
    await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
    await page.waitForTimeout(3000);

    // Check that no toFixed errors occurred
    const toFixedErrors = consoleErrors.filter((err) =>
      err.includes("Cannot read properties of undefined (reading 'toFixed')")
    );
    const toFixedPageErrors = pageErrors.filter((err) =>
      err.message.includes("Cannot read properties of undefined (reading 'toFixed')")
    );

    expect(toFixedErrors.length).toBe(0);
    expect(toFixedPageErrors.length).toBe(0);
  });
});

