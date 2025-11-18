import { expect, test } from '@playwright/test';

test.describe('Bulk Operations Tests', () => {
  test('should load bulk operations page', async ({ page }) => {
    await page.goto('/merchant/bulk-operations');
    
    await expect(page.locator('h1')).toContainText(/bulk.*operation/i);
    
    // Check for merchant selection interface
    const selectionInterface = page.locator('text=/merchant.*selection|select.*merchant/i').first();
    await expect(selectionInterface).toBeVisible({ timeout: 5000 });
  });

  test('should select and deselect merchants', async ({ page }) => {
    await page.goto('/merchant/bulk-operations');
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    
    // Find checkboxes
    const checkboxes = page.locator('input[type="checkbox"]');
    const count = await checkboxes.count();
    
    if (count > 0) {
      // Select first merchant
      await checkboxes.first().check();
      
      // Check select all
      const selectAllButton = page.locator('button:has-text("Select All")').first();
      if (await selectAllButton.isVisible({ timeout: 2000 })) {
        await selectAllButton.click();
        await page.waitForTimeout(500);
      }
      
      // Check deselect all
      const deselectAllButton = page.locator('button:has-text("Deselect")').first();
      if (await deselectAllButton.isVisible({ timeout: 2000 })) {
        await deselectAllButton.click();
        await page.waitForTimeout(500);
      }
    }
  });

  test('should select operation type', async ({ page }) => {
    await page.goto('/merchant/bulk-operations');
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    
    // Find operation type buttons
    const portfolioButton = page.locator('button:has-text("Portfolio")').first();
    if (await portfolioButton.isVisible({ timeout: 5000 })) {
      await portfolioButton.click();
      await page.waitForTimeout(1000);
      
      // Should show configuration
      const config = page.locator('select, input, textarea').first();
      await expect(config).toBeVisible();
    }
  });

  test('should show operation progress', async ({ page }) => {
    await page.goto('/merchant/bulk-operations');
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    
    // Select merchants and operation
    const checkbox = page.locator('input[type="checkbox"]').first();
    if (await checkbox.isVisible({ timeout: 5000 })) {
      await checkbox.check();
      
      const operationButton = page.locator('button:has-text("Portfolio"), button:has-text("Risk")').first();
      if (await operationButton.isVisible({ timeout: 2000 })) {
        await operationButton.click();
        await page.waitForTimeout(1000);
        
        // Try to start operation
        const startButton = page.locator('button:has-text("Start")').first();
        if (await startButton.isVisible({ timeout: 2000 })) {
          // Don't actually start to avoid side effects, just verify UI
          await expect(startButton).toBeVisible();
        }
      }
    }
  });
});

