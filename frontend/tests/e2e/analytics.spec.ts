import { test, expect } from '@playwright/test';

test.describe('Analytics Data Loading', () => {
  test.beforeEach(async ({ page }) => {
    // Mock merchant API
    await page.route('**/api/v1/merchants/*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'merchant-123',
          businessName: 'Test Business',
          status: 'active',
        }),
      });
    });
  });

  test('should load analytics data', async ({ page }) => {
    // Mock analytics API
    await page.route('**/api/v1/merchants/*/analytics', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          merchantId: 'merchant-123',
          classification: {
            primaryIndustry: 'Technology',
            confidenceScore: 0.95,
          },
          security: {
            trustScore: 0.8,
            sslValid: true,
          },
          quality: {
            completenessScore: 0.9,
            dataPoints: 100,
          },
        }),
      });
    });

    await page.goto('/merchants/merchant-123');
    
    // Navigate to Business Analytics tab
    await page.getByRole('tab', { name: 'Business Analytics' }).click();
    
    // Should display analytics data
    await expect(page.getByText('Technology')).toBeVisible();
    await expect(page.getByText(/95%/)).toBeVisible();
  });

  test('should lazy load website analysis', async ({ page }) => {
    let websiteAnalysisCalled = false;
    
    // Mock analytics API
    await page.route('**/api/v1/merchants/*/analytics', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          merchantId: 'merchant-123',
          classification: { primaryIndustry: 'Technology' },
        }),
      });
    });

    // Mock website analysis API
    await page.route('**/api/v1/merchants/*/website-analysis', async (route) => {
      websiteAnalysisCalled = true;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          merchantId: 'merchant-123',
          websiteUrl: 'https://test.com',
          performance: { score: 85 },
        }),
      });
    });

    await page.goto('/merchants/merchant-123');
    await page.getByRole('tab', { name: 'Business Analytics' }).click();
    
    // Scroll to trigger lazy loading
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    
    // Wait a bit for lazy loading
    await page.waitForTimeout(1000);
    
    // Website analysis should be called
    expect(websiteAnalysisCalled).toBeTruthy();
  });

  test('should show empty state when no analytics data', async ({ page }) => {
    // Mock empty analytics
    await page.route('**/api/v1/merchants/*/analytics', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(null),
      });
    });

    await page.goto('/merchants/merchant-123');
    await page.getByRole('tab', { name: 'Business Analytics' }).click();
    
    // Should show empty state
    await expect(page.getByText(/no analytics data/i)).toBeVisible();
  });
});

