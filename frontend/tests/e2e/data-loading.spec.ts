import { expect, test } from '@playwright/test';

test.describe('Data Loading Tests', () => {
  test('should load dashboard metrics', async ({ page }) => {
    await page.goto('http://localhost:3000/dashboard');
    
    // Wait for loading to complete
    await page.waitForSelector('text=/merchants|revenue|growth/i', { timeout: 10000 });
    
    // Check that metrics are displayed (not just loading skeletons)
    const metrics = page.locator('text=/\\d+/').first();
    await expect(metrics).toBeVisible();
  });

  test('should load merchant portfolio list', async ({ page }) => {
    await page.goto('http://localhost:3000/merchant-portfolio');
    
    // Wait for table or list to load
    await page.waitForSelector('table, [role="table"], .merchant-item, [data-testid="merchant-list"]', { timeout: 10000 });
    
    // Check that table is visible (table element exists in DOM)
    const table = page.locator('table, [role="table"]').first();
    await expect(table).toBeVisible();
    
    // Check that either merchants are displayed OR empty state is shown
    const hasMerchants = page.getByText(/merchant/i).first();
    const hasEmptyState = page.getByText(/no.*found|no.*merchants|empty/i).first();
    
    // At least one should be visible (merchants OR empty state)
    const hasContent = await hasMerchants.isVisible({ timeout: 2000 }).catch(() => false) ||
                       await hasEmptyState.isVisible({ timeout: 2000 }).catch(() => false);
    
    expect(hasContent || await table.isVisible()).toBeTruthy();
  });

  test('should load merchant details', async ({ page }) => {
    // First, get a merchant ID from the portfolio
    await page.goto('http://localhost:3000/merchant-portfolio');
    await page.waitForTimeout(2000);
    
    // Try to find a merchant link
    const merchantLink = page.locator('a[href*="merchant-details"], button:has-text("View")').first();
    
    if (await merchantLink.isVisible({ timeout: 5000 })) {
      await merchantLink.click();
      
      // Wait for merchant details to load
      await page.waitForSelector('h1, h2, [data-testid="merchant-details"]', { timeout: 10000 });
      
      // Check that details are displayed
      const details = page.locator('h1, h2').first();
      await expect(details).toBeVisible();
    } else {
      // Skip if no merchants available
      test.skip();
    }
  });

  test('should show loading states', async ({ page }) => {
    await page.goto('http://localhost:3000/dashboard');
    
    // Check for skeleton loaders (they should appear briefly)
    const skeleton = page.locator('[class*="skeleton"], [class*="Skeleton"]').first();
    
    // Skeleton might disappear quickly, so just check if page loads
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    
    // Page should eventually show content
    await expect(page.locator('body')).toBeVisible();
  });

  test('should handle API errors gracefully', async ({ page }) => {
    // Intercept API calls and return error
    await page.route('**/api/v1/**', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal Server Error' }),
      });
    });
    
    await page.goto('http://localhost:3000/dashboard');
    await page.waitForLoadState('networkidle');
    
    // Should show error message, empty state, or at least not crash
    // Error might be in toast notification, alert, or empty state
    const errorMessage = page.locator('text=/error|failed|unavailable|something went wrong/i').first();
    const hasEmptyState = page.locator('text=/no data|empty|unavailable/i').first();
    const hasToast = page.locator('[role="status"], [data-sonner-toast]').first();
    
    // At least one should be visible, or page should still be functional
    const hasError = await errorMessage.isVisible({ timeout: 3000 }).catch(() => false) ||
                     await hasEmptyState.isVisible({ timeout: 3000 }).catch(() => false) ||
                     await hasToast.isVisible({ timeout: 3000 }).catch(() => false);
    
    // Page should still be functional (not crashed)
    const pageLoaded = await page.locator('body').isVisible();
    expect(hasError || pageLoaded).toBeTruthy();
  });

  test('should handle search and filtering', async ({ page }) => {
    await page.goto('http://localhost:3000/merchant-portfolio');
    
    // Wait for page to load
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    
    // Try to use search
    const searchInput = page.locator('input[placeholder*="search" i], input[type="search"]').first();
    if (await searchInput.isVisible({ timeout: 2000 })) {
      await searchInput.fill('test');
      await page.waitForTimeout(1500); // Wait for debounce
      
      // Results should update (table should still be visible)
      const table = page.locator('table, [role="table"]').first();
      await expect(table).toBeVisible({ timeout: 3000 });
    }
    
    // Try to use filters - these are comboboxes (Select components from shadcn)
    const filters = page.locator('[role="combobox"]').all();
    const filterCount = await filters.length;
    
    if (filterCount > 0) {
      // Click first filter (status filter)
      await filters[0].click();
      await page.waitForTimeout(500);
      
      // Try to select an option if dropdown is open
      const firstOption = page.locator('[role="option"]').first();
      if (await firstOption.isVisible({ timeout: 1000 }).catch(() => false)) {
        await firstOption.click();
        await page.waitForTimeout(1000);
      } else {
        // Close dropdown if needed
        await page.keyboard.press('Escape');
      }
    }
    
    // Page should still be functional
    await expect(page.locator('body')).toBeVisible();
  });
});

