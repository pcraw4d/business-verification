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
    // Mock no existing assessment - API uses /api/v1/merchants/:merchantId/risk-score
    await page.route('**/api/v1/merchants/*/risk-score', async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({}),
      });
    });

    // Mock start assessment - API uses POST /api/v1/risk/assess
    await page.route('**/api/v1/risk/assess', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            assessmentId: 'assessment-123',
            status: 'pending',
          }),
        });
      } else {
        await route.continue();
      }
    });

    await page.goto('/merchant-details/merchant-123');
    
    // Wait for page to load first
    await expect(page.getByRole('heading', { name: 'Test Business' })).toBeVisible({ timeout: 10000 });
    
    // Navigate to Risk Assessment tab
    await page.getByRole('tab', { name: 'Risk Assessment' }).click();
    
    // Click start assessment button
    await page.getByRole('button', { name: /start assessment/i }).click();
    
    // Should show processing state
    await expect(page.getByText(/processing|pending/i)).toBeVisible();
  });

  test('should display completed risk assessment', async ({ page }) => {
    // Mock completed assessment - API uses /api/v1/merchants/:merchantId/risk-score
    await page.route('**/api/v1/merchants/*/risk-score', async (route) => {
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

    await page.goto('/merchant-details/merchant-123');
    
    // Wait for page to load first
    await expect(page.getByRole('heading', { name: 'Test Business' })).toBeVisible({ timeout: 10000 });
    
    await page.getByRole('tab', { name: 'Risk Assessment' }).click();
    
    // Should show completed assessment
    await expect(page.getByText(/completed/i)).toBeVisible();
    await expect(page.getByText(/medium/i)).toBeVisible();
  });

  test('should poll for assessment status', async ({ page }) => {
    let pollCount = 0;
    
    // Mock no existing assessment first
    await page.route('**/api/v1/merchants/*/risk-score', async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({}),
      });
    });
    
    // Mock start assessment
    await page.route('**/api/v1/risk/assess', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            assessmentId: 'assessment-123',
            status: 'pending',
          }),
        });
      } else {
        await route.continue();
      }
    });
    
    // Mock status polling - API uses GET /api/v1/risk/assess/:assessmentId
    await page.route('**/api/v1/risk/assess/assessment-123', async (route) => {
      if (route.request().method() === 'GET') {
        pollCount++;
        if (pollCount === 1) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              assessmentId: 'assessment-123',
              status: 'processing',
              progress: 50,
            }),
          });
        } else {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              assessmentId: 'assessment-123',
              status: 'completed',
              progress: 100,
            }),
          });
        }
      } else {
        await route.continue();
      }
    });

    await page.goto('/merchant-details/merchant-123');
    
    // Wait for page to load first
    await expect(page.getByRole('heading', { name: 'Test Business' })).toBeVisible({ timeout: 10000 });
    
    await page.getByRole('tab', { name: 'Risk Assessment' }).click();
    
    // Click start assessment button to trigger polling
    await page.getByRole('button', { name: /start.*assessment/i }).click();
    
    // Wait for polling to complete - component will reload assessment after status becomes 'completed'
    await expect(page.getByText(/completed/i)).toBeVisible({ timeout: 15000 });
  });
});

