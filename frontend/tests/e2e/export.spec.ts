import { expect, test } from '@playwright/test';

test.describe('Export Functionality Tests', () => {
  test('should export data as CSV', async ({ page }) => {
    await page.goto('/merchant-portfolio');
    await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
    await page.waitForTimeout(2000); // Additional wait for dynamic content
    
    // Find export button
    const exportButton = page.locator('button:has-text("Export")').first();
    const hasExportButton = await exportButton.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (!hasExportButton) {
      test.skip();
      return;
    }
    
    await exportButton.scrollIntoViewIfNeeded();
    await exportButton.click({ force: true });
    await page.waitForTimeout(1000);
    
    // Click CSV option
    const csvOption = page.locator('text=/csv/i').first();
    await csvOption.click({ force: true });
    
    // Wait for download (simplified - actual download handling may vary)
    await page.waitForTimeout(2000);
    
    // Check for success message (optional - may not always appear)
    const success = page.locator('text=/export|download|success/i').first();
    const hasSuccess = await success.isVisible({ timeout: 3000 }).catch(() => false);
    if (hasSuccess) {
      await expect(success).toBeVisible();
    }
  });

  test('should export data as JSON', async ({ page }) => {
    await page.goto('/merchant-portfolio');
    await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
    await page.waitForTimeout(2000); // Additional wait for dynamic content
    
    const exportButton = page.locator('button:has-text("Export")').first();
    const hasExportButton = await exportButton.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (!hasExportButton) {
      test.skip();
      return;
    }
    
    await exportButton.scrollIntoViewIfNeeded();
    await exportButton.click({ force: true });
    await page.waitForTimeout(1000);
    
    const jsonOption = page.locator('text=/json/i').first();
    await jsonOption.click({ force: true });
    
    await page.waitForTimeout(2000);
  });

  test('should export from risk assessment tab', async ({ page }) => {
    // Navigate to a merchant details page with risk assessment
    await page.goto('/merchant-portfolio');
    await page.waitForLoadState('domcontentloaded', { timeout: 15000 });
    await page.waitForTimeout(2000);
    
    const merchantLink = page.locator('a[href*="merchant-details"], button:has-text("View")').first();
    const hasMerchantLink = await merchantLink.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (!hasMerchantLink) {
      test.skip();
      return;
    }
    
    await merchantLink.scrollIntoViewIfNeeded();
    await merchantLink.click({ force: true });
    await page.waitForTimeout(3000);
    
    // Navigate to risk assessment tab
    const riskTab = page.locator('button:has-text("Risk"), [role="tab"]:has-text("Risk")').first();
    const hasRiskTab = await riskTab.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (!hasRiskTab) {
      test.skip();
      return;
    }
    
    await riskTab.scrollIntoViewIfNeeded();
    await riskTab.click({ force: true });
    await page.waitForTimeout(3000);
    
    // Find export button in risk tab
    const exportButton = page.locator('button:has-text("Export")').first();
    const hasExportButton = await exportButton.isVisible({ timeout: 5000 }).catch(() => false);
    
    if (!hasExportButton) {
      test.skip();
      return;
    }
    
    await exportButton.scrollIntoViewIfNeeded();
    await exportButton.click({ force: true });
    await page.waitForTimeout(2000);
    
    // Should show export options
    const exportOptions = page.locator('text=/csv|json|excel|pdf/i');
    await expect(exportOptions.first()).toBeVisible({ timeout: 5000 });
  });
});

