import { test, expect } from '@playwright/test';

test.describe('Risk Assessment Flow', () => {
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

  test('should start risk assessment', async ({ page }) => {
    // Mock no existing assessment
    await page.route('**/api/v1/risk/assessments/*', async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({}),
      });
    });

    // Mock start assessment
    await page.route('**/api/v1/risk/assess', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          assessmentId: 'assessment-123',
          status: 'pending',
        }),
      });
    });

    await page.goto('/merchants/merchant-123');
    
    // Navigate to Risk Assessment tab
    await page.getByRole('tab', { name: 'Risk Assessment' }).click();
    
    // Click start assessment button
    await page.getByRole('button', { name: /start assessment/i }).click();
    
    // Should show processing state
    await expect(page.getByText(/processing|pending/i)).toBeVisible();
  });

  test('should display completed risk assessment', async ({ page }) => {
    // Mock completed assessment
    await page.route('**/api/v1/risk/assessments/*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'assessment-123',
          merchantId: 'merchant-123',
          status: 'completed',
          progress: 100,
          result: {
            overallScore: 0.7,
            riskLevel: 'medium',
            factors: [],
          },
        }),
      });
    });

    await page.goto('/merchants/merchant-123');
    await page.getByRole('tab', { name: 'Risk Assessment' }).click();
    
    // Should show completed assessment
    await expect(page.getByText(/completed/i)).toBeVisible();
    await expect(page.getByText(/medium/i)).toBeVisible();
  });

  test('should poll for assessment status', async ({ page }) => {
    let pollCount = 0;
    
    // Mock status polling
    await page.route('**/api/v1/risk/assessments/*/status', async (route) => {
      pollCount++;
      if (pollCount === 1) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            status: 'processing',
            progress: 50,
          }),
        });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            status: 'completed',
            progress: 100,
          }),
        });
      }
    });

    await page.goto('/merchants/merchant-123');
    await page.getByRole('tab', { name: 'Risk Assessment' }).click();
    
    // Wait for polling to complete
    await expect(page.getByText(/completed/i)).toBeVisible({ timeout: 10000 });
  });
});

