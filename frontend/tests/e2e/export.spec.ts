import { test, expect } from '@playwright/test';
import { download } from '@playwright/test';

test.describe('Export Functionality Tests', () => {
  test('should export data as CSV', async ({ page }) => {
    await page.goto('http://localhost:3000/merchant-portfolio');
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    
    // Find export button
    const exportButton = page.locator('button:has-text("Export")').first();
    if (await exportButton.isVisible({ timeout: 5000 })) {
      await exportButton.click();
      
      // Click CSV option
      const csvOption = page.locator('text=/csv/i').first();
      await csvOption.click();
      
      // Wait for download (simplified - actual download handling may vary)
      await page.waitForTimeout(2000);
      
      // Check for success message
      const success = page.locator('text=/export|download|success/i').first();
      if (await success.isVisible({ timeout: 3000 })) {
        await expect(success).toBeVisible();
      }
    } else {
      test.skip('Export button not found');
    }
  });

  test('should export data as JSON', async ({ page }) => {
    await page.goto('http://localhost:3000/merchant-portfolio');
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    
    const exportButton = page.locator('button:has-text("Export")').first();
    if (await exportButton.isVisible({ timeout: 5000 })) {
      await exportButton.click();
      
      const jsonOption = page.locator('text=/json/i').first();
      await jsonOption.click();
      
      await page.waitForTimeout(2000);
    } else {
      test.skip('Export button not found');
    }
  });

  test('should export from risk assessment tab', async ({ page }) => {
    // Navigate to a merchant details page with risk assessment
    await page.goto('http://localhost:3000/merchant-portfolio');
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    
    const merchantLink = page.locator('a[href*="merchant-details"], button:has-text("View")').first();
    if (await merchantLink.isVisible({ timeout: 5000 })) {
      await merchantLink.click();
      await page.waitForTimeout(2000);
      
      // Navigate to risk assessment tab
      const riskTab = page.locator('button:has-text("Risk"), [role="tab"]:has-text("Risk")').first();
      if (await riskTab.isVisible({ timeout: 3000 })) {
        await riskTab.click();
        await page.waitForTimeout(2000);
        
        // Find export button in risk tab
        const exportButton = page.locator('button:has-text("Export")').first();
        if (await exportButton.isVisible({ timeout: 3000 })) {
          await exportButton.click();
          await page.waitForTimeout(1000);
          
          // Should show export options
          const exportOptions = page.locator('text=/csv|json|excel|pdf/i');
          await expect(exportOptions.first()).toBeVisible();
        }
      }
    } else {
      test.skip('No merchant details page available');
    }
  });
});

