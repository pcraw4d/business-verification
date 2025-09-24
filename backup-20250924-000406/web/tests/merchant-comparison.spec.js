// @ts-check
const { test, expect } = require('@playwright/test');

test.describe('Merchant Comparison', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to merchant comparison page
    await page.goto('/merchant-comparison.html');
    await page.waitForLoadState('networkidle');
    
    // Wait for page to load completely
    await page.waitForSelector('[data-testid="comparison-container"]', { timeout: 10000 });
  });

  test('should display merchant comparison page with all required elements', async ({ page }) => {
    // Check page title
    await expect(page).toHaveTitle(/Merchant Comparison/);
    
    // Check main heading
    await expect(page.locator('h1')).toContainText('Merchant Comparison');
    
    // Check merchant selection section
    await expect(page.locator('[data-testid="merchant-selection"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-1-selector"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-selector"]')).toBeVisible();
    
    // Check comparison container
    await expect(page.locator('[data-testid="comparison-container"]')).toBeVisible();
    
    // Check comparison sections
    await expect(page.locator('[data-testid="basic-info-comparison"]')).toBeVisible();
    await expect(page.locator('[data-testid="portfolio-comparison"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-comparison"]')).toBeVisible();
    await expect(page.locator('[data-testid="compliance-comparison"]')).toBeVisible();
    
    // Check action buttons
    await expect(page.locator('[data-testid="export-comparison"]')).toBeVisible();
    await expect(page.locator('[data-testid="clear-comparison"]')).toBeVisible();
  });

  test('should display mock data warning', async ({ page }) => {
    // Check for mock data warning
    await expect(page.locator('[data-testid="mock-data-warning"]')).toBeVisible();
    await expect(page.locator('[data-testid="mock-data-warning"]')).toContainText('Mock Data');
  });

  test('should select merchants for comparison', async ({ page }) => {
    // Wait for merchant selectors to load
    await page.waitForSelector('[data-testid="merchant-1-selector"]', { timeout: 10000 });
    
    // Select first merchant
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    // Select second merchant
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="merchant-1-details"]');
    await page.waitForSelector('[data-testid="merchant-2-details"]');
    
    // Check that both merchants are displayed
    await expect(page.locator('[data-testid="merchant-1-details"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-details"]')).toBeVisible();
  });

  test('should display basic information comparison', async ({ page }) => {
    // Select merchants for comparison
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="basic-info-comparison"]');
    
    // Check basic information comparison
    await expect(page.locator('[data-testid="basic-info-comparison"]')).toBeVisible();
    
    // Check merchant names
    await expect(page.locator('[data-testid="merchant-1-name"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-name"]')).toBeVisible();
    
    // Check industries
    await expect(page.locator('[data-testid="merchant-1-industry"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-industry"]')).toBeVisible();
    
    // Check addresses
    await expect(page.locator('[data-testid="merchant-1-address"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-address"]')).toBeVisible();
    
    // Check contact information
    await expect(page.locator('[data-testid="merchant-1-phone"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-phone"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-1-email"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-email"]')).toBeVisible();
  });

  test('should display portfolio information comparison', async ({ page }) => {
    // Select merchants for comparison
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="portfolio-comparison"]');
    
    // Check portfolio comparison
    await expect(page.locator('[data-testid="portfolio-comparison"]')).toBeVisible();
    
    // Check portfolio types
    await expect(page.locator('[data-testid="merchant-1-portfolio-type"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-portfolio-type"]')).toBeVisible();
    
    // Check onboarding dates
    await expect(page.locator('[data-testid="merchant-1-onboarding-date"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-onboarding-date"]')).toBeVisible();
    
    // Check last updated dates
    await expect(page.locator('[data-testid="merchant-1-last-updated"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-last-updated"]')).toBeVisible();
  });

  test('should display risk assessment comparison', async ({ page }) => {
    // Select merchants for comparison
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="risk-comparison"]');
    
    // Check risk comparison
    await expect(page.locator('[data-testid="risk-comparison"]')).toBeVisible();
    
    // Check risk levels
    await expect(page.locator('[data-testid="merchant-1-risk-level"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-risk-level"]')).toBeVisible();
    
    // Check risk scores
    await expect(page.locator('[data-testid="merchant-1-risk-score"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-risk-score"]')).toBeVisible();
    
    // Check risk factors
    await expect(page.locator('[data-testid="merchant-1-risk-factors"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-risk-factors"]')).toBeVisible();
  });

  test('should display compliance comparison', async ({ page }) => {
    // Select merchants for comparison
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="compliance-comparison"]');
    
    // Check compliance comparison
    await expect(page.locator('[data-testid="compliance-comparison"]')).toBeVisible();
    
    // Check compliance statuses
    await expect(page.locator('[data-testid="merchant-1-compliance-status"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-compliance-status"]')).toBeVisible();
    
    // Check compliance scores
    await expect(page.locator('[data-testid="merchant-1-compliance-score"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-compliance-score"]')).toBeVisible();
    
    // Check compliance details
    await expect(page.locator('[data-testid="merchant-1-compliance-details"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-compliance-details"]')).toBeVisible();
  });

  test('should highlight differences between merchants', async ({ page }) => {
    // Select merchants for comparison
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="comparison-container"]');
    
    // Check that differences are highlighted
    const differences = page.locator('[data-testid="difference-highlight"]');
    const count = await differences.count();
    expect(count).toBeGreaterThan(0);
    
    // Check that differences have appropriate styling
    for (let i = 0; i < count; i++) {
      const difference = differences.nth(i);
      await expect(difference).toHaveClass(/highlight-difference/);
    }
  });

  test('should export comparison report', async ({ page }) => {
    // Select merchants for comparison
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="comparison-container"]');
    
    // Set up download promise
    const downloadPromise = page.waitForEvent('download');
    
    // Click export comparison button
    await page.locator('[data-testid="export-comparison"]').click();
    
    // Wait for download
    const download = await downloadPromise;
    
    // Check download filename
    expect(download.suggestedFilename()).toMatch(/merchant-comparison.*\.pdf/);
  });

  test('should clear comparison and reset form', async ({ page }) => {
    // Select merchants for comparison
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="comparison-container"]');
    
    // Click clear comparison button
    await page.locator('[data-testid="clear-comparison"]').click();
    
    // Check that comparison is cleared
    await expect(page.locator('[data-testid="comparison-container"]')).not.toBeVisible();
    
    // Check that selectors are reset
    await expect(page.locator('[data-testid="merchant-1-selector"]')).toHaveValue('');
    await expect(page.locator('[data-testid="merchant-2-selector"]')).toHaveValue('');
  });

  test('should prevent selecting the same merchant twice', async ({ page }) => {
    // Wait for merchant selectors to load
    await page.waitForSelector('[data-testid="merchant-1-selector"]', { timeout: 10000 });
    
    // Select first merchant
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    // Try to select the same merchant for second selector
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    
    // Check that the same merchant option is disabled
    const firstOption = page.locator('[data-testid="merchant-option"]').first();
    await expect(firstOption).toHaveClass(/disabled/);
    
    // Select a different merchant
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="comparison-container"]');
    
    // Check that comparison is displayed
    await expect(page.locator('[data-testid="comparison-container"]')).toBeVisible();
  });

  test('should display comparison summary', async ({ page }) => {
    // Select merchants for comparison
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="comparison-summary"]');
    
    // Check comparison summary
    await expect(page.locator('[data-testid="comparison-summary"]')).toBeVisible();
    
    // Check summary statistics
    await expect(page.locator('[data-testid="total-differences"]')).toBeVisible();
    await expect(page.locator('[data-testid="similarity-score"]')).toBeVisible();
    await expect(page.locator('[data-testid="recommendation"]')).toBeVisible();
  });

  test('should handle merchant selection errors gracefully', async ({ page }) => {
    // Mock network error for merchant selection
    await page.route('**/api/merchants', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal server error' })
      });
    });
    
    // Try to select merchant
    await page.locator('[data-testid="merchant-1-selector"]').click();
    
    // Check that error message is displayed
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-message"]')).toContainText('Failed to load merchants');
  });

  test('should be responsive on mobile devices', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    // Wait for page to load
    await page.waitForSelector('[data-testid="comparison-container"]', { timeout: 10000 });
    
    // Check that all elements are still visible and accessible
    await expect(page.locator('[data-testid="merchant-selection"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-1-selector"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-2-selector"]')).toBeVisible();
    
    // Check that action buttons are accessible
    await expect(page.locator('[data-testid="export-comparison"]')).toBeVisible();
    await expect(page.locator('[data-testid="clear-comparison"]')).toBeVisible();
  });

  test('should handle empty merchant selection gracefully', async ({ page }) => {
    // Wait for page to load
    await page.waitForSelector('[data-testid="comparison-container"]', { timeout: 10000 });
    
    // Check that comparison container is not visible initially
    await expect(page.locator('[data-testid="comparison-container"]')).not.toBeVisible();
    
    // Check that empty state message is displayed
    await expect(page.locator('[data-testid="empty-state"]')).toBeVisible();
    await expect(page.locator('[data-testid="empty-state"]')).toContainText('Select two merchants to compare');
  });

  test('should display loading state while fetching merchant data', async ({ page }) => {
    // Select merchants for comparison
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Check that loading indicator is shown
    await expect(page.locator('[data-testid="loading-indicator"]')).toBeVisible();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="comparison-container"]');
    
    // Check that loading indicator is hidden
    await expect(page.locator('[data-testid="loading-indicator"]')).not.toBeVisible();
  });

  test('should allow switching between different merchants', async ({ page }) => {
    // Select first set of merchants
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').first().click();
    
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for comparison to load
    await page.waitForSelector('[data-testid="comparison-container"]');
    
    // Get first merchant names
    const merchant1Name = await page.locator('[data-testid="merchant-1-name"]').textContent();
    const merchant2Name = await page.locator('[data-testid="merchant-2-name"]').textContent();
    
    // Switch to different merchants
    await page.locator('[data-testid="merchant-1-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(2).click();
    
    await page.locator('[data-testid="merchant-2-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(3).click();
    
    // Wait for new comparison to load
    await page.waitForSelector('[data-testid="comparison-container"]');
    
    // Check that merchant names have changed
    const newMerchant1Name = await page.locator('[data-testid="merchant-1-name"]').textContent();
    const newMerchant2Name = await page.locator('[data-testid="merchant-2-name"]').textContent();
    
    expect(newMerchant1Name).not.toBe(merchant1Name);
    expect(newMerchant2Name).not.toBe(merchant2Name);
  });
});
