import { test, expect } from '@playwright/test';

test.describe('Merchant Details Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API responses
    await page.route('**/api/v1/merchants/*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'merchant-123',
          businessName: 'Test Business',
          industry: 'Technology',
          status: 'active',
          email: 'test@example.com',
          phone: '+1-555-123-4567',
          website: 'https://test.com',
        }),
      });
    });
  });

  test('should load merchant details page', async ({ page }) => {
    await page.goto('/merchant-details/merchant-123');
    
    // Wait for merchant name to appear - use getByRole for heading to avoid strict mode violation
    await expect(page.getByRole('heading', { name: 'Test Business' })).toBeVisible({ timeout: 10000 });
    
    // Verify tabs are present
    await expect(page.getByRole('tab', { name: 'Overview' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Business Analytics' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Risk Assessment' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Risk Indicators' })).toBeVisible();
  });

  test('should navigate between tabs', async ({ page }) => {
    await page.goto('/merchant-details/merchant-123');
    
    // Wait for page to load - use getByRole for heading to avoid strict mode violation
    await expect(page.getByRole('heading', { name: 'Test Business' })).toBeVisible({ timeout: 10000 });
    
    // Click Business Analytics tab
    await page.getByRole('tab', { name: 'Business Analytics' }).click();
    await expect(page.getByRole('tab', { name: 'Business Analytics' })).toHaveAttribute('data-state', 'active');
    
    // Click Risk Assessment tab
    await page.getByRole('tab', { name: 'Risk Assessment' }).click();
    await expect(page.getByRole('tab', { name: 'Risk Assessment' })).toHaveAttribute('data-state', 'active');
  });

  test('should display merchant information in Overview tab', async ({ page }) => {
    await page.goto('/merchant-details/merchant-123');
    
    // Wait for page to load - use getByRole for heading to avoid strict mode violation
    await expect(page.getByRole('heading', { name: 'Test Business' })).toBeVisible({ timeout: 10000 });
    
    // Overview tab should be active by default
    // Use getByRole('tabpanel') to scope the search to the Overview tab content
    const overviewTab = page.getByRole('tabpanel', { name: 'Overview' });
    await expect(overviewTab.getByText('Test Business')).toBeVisible();
    await expect(overviewTab.getByText('Technology')).toBeVisible();
    await expect(overviewTab.getByText(/active/i)).toBeVisible();
  });

  test('should handle API errors gracefully', async ({ page }) => {
    // Mock API error - override the beforeEach mock
    await page.route('**/api/v1/merchants/*', async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 'NOT_FOUND',
          message: 'Merchant not found',
        }),
      });
    });

    await page.goto('/merchant-details/merchant-123');
    
    // Should show error message in Alert component - scope to alert role to avoid strict mode violation
    // The component renders error in an Alert with role="alert"
    const alert = page.getByRole('alert');
    await expect(alert.getByText(/API Error|Merchant not found|error|failed/i)).toBeVisible({ timeout: 10000 });
  });
});

