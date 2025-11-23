// @ts-check
const { test, expect } = require('@playwright/test');
const { setupAPIMocks } = require('./utils/api-mock-helpers');

test.describe('Merchant Bulk Operations', () => {
  test.beforeEach(async ({ page }) => {
    // Setup API mocks before navigation
    await setupAPIMocks(page);
    
    // Navigate to merchant bulk operations page
    await page.goto('/merchant-bulk-operations.html');
    await page.waitForLoadState('networkidle');
    
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-list"]', { timeout: 10000 });
  });

  test('should display bulk operations page with all required elements', async ({ page }) => {
    // Check page title
    await expect(page).toHaveTitle(/Bulk Operations/);
    
    // Check main heading
    await expect(page.locator('h1')).toContainText('Bulk Operations');
    
    // Check merchant list
    await expect(page.locator('[data-testid="merchant-list"]')).toBeVisible();
    
    // Check bulk selection controls
    await expect(page.locator('[data-testid="bulk-select-all"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-deselect-all"]')).toBeVisible();
    
    // Check bulk action buttons
    await expect(page.locator('[data-testid="bulk-actions"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-update-portfolio-type"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-update-risk-level"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-export"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-deactivate"]')).toBeVisible();
    
    // Check progress tracking
    await expect(page.locator('[data-testid="bulk-progress"]')).toBeVisible();
    await expect(page.locator('[data-testid="selected-count"]')).toBeVisible();
    
    // Check operation history
    await expect(page.locator('[data-testid="operation-history"]')).toBeVisible();
  });

  test('should display mock data warning', async ({ page }) => {
    // Check for mock data warning
    await expect(page.locator('[data-testid="mock-data-warning"]')).toBeVisible();
    await expect(page.locator('[data-testid="mock-data-warning"]')).toContainText('Mock Data');
  });

  test('should select and deselect individual merchants', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Check initial selected count
    await expect(page.locator('[data-testid="selected-count"]')).toContainText('0');
    
    // Select first merchant
    await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-checkbox"]').check();
    
    // Check selected count updated
    await expect(page.locator('[data-testid="selected-count"]')).toContainText('1');
    
    // Select second merchant
    await page.locator('[data-testid="merchant-item"]').nth(1).locator('[data-testid="merchant-checkbox"]').check();
    
    // Check selected count updated
    await expect(page.locator('[data-testid="selected-count"]')).toContainText('2');
    
    // Deselect first merchant
    await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-checkbox"]').uncheck();
    
    // Check selected count updated
    await expect(page.locator('[data-testid="selected-count"]')).toContainText('1');
  });

  test('should select and deselect all merchants', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Get total merchant count
    const totalMerchants = await page.locator('[data-testid="merchant-item"]').count();
    
    // Check initial selected count
    await expect(page.locator('[data-testid="selected-count"]')).toContainText('0');
    
    // Select all merchants
    await page.locator('[data-testid="bulk-select-all"]').click();
    
    // Check all merchants are selected
    await expect(page.locator('[data-testid="selected-count"]')).toContainText(totalMerchants.toString());
    
    // Verify all checkboxes are checked
    for (let i = 0; i < totalMerchants; i++) {
      const checkbox = page.locator('[data-testid="merchant-item"]').nth(i).locator('[data-testid="merchant-checkbox"]');
      await expect(checkbox).toBeChecked();
    }
    
    // Deselect all merchants
    await page.locator('[data-testid="bulk-deselect-all"]').click();
    
    // Check no merchants are selected
    await expect(page.locator('[data-testid="selected-count"]')).toContainText('0');
    
    // Verify all checkboxes are unchecked
    for (let i = 0; i < totalMerchants; i++) {
      const checkbox = page.locator('[data-testid="merchant-item"]').nth(i).locator('[data-testid="merchant-checkbox"]');
      await expect(checkbox).not.toBeChecked();
    }
  });

  test('should update portfolio type for selected merchants', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Select first two merchants
    await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-checkbox"]').check();
    await page.locator('[data-testid="merchant-item"]').nth(1).locator('[data-testid="merchant-checkbox"]').check();
    
    // Click bulk update portfolio type
    await page.locator('[data-testid="bulk-update-portfolio-type"]').click();
    
    // Wait for modal to appear
    await page.waitForSelector('[data-testid="portfolio-type-modal"]');
    
    // Select new portfolio type
    await page.locator('[data-testid="portfolio-type-pending"]').click();
    
    // Confirm update
    await page.locator('[data-testid="confirm-bulk-update"]').click();
    
    // Wait for operation to complete
    await page.waitForSelector('[data-testid="bulk-progress"]', { state: 'hidden' });
    
    // Check that selected merchants have updated portfolio type
    await expect(page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-portfolio-type"]')).toContainText('Pending');
    await expect(page.locator('[data-testid="merchant-item"]').nth(1).locator('[data-testid="merchant-portfolio-type"]')).toContainText('Pending');
  });

  test('should update risk level for selected merchants', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Select first merchant
    await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-checkbox"]').check();
    
    // Click bulk update risk level
    await page.locator('[data-testid="bulk-update-risk-level"]').click();
    
    // Wait for modal to appear
    await page.waitForSelector('[data-testid="risk-level-modal"]');
    
    // Select new risk level
    await page.locator('[data-testid="risk-level-high"]').click();
    
    // Confirm update
    await page.locator('[data-testid="confirm-bulk-update"]').click();
    
    // Wait for operation to complete
    await page.waitForSelector('[data-testid="bulk-progress"]', { state: 'hidden' });
    
    // Check that selected merchant has updated risk level
    await expect(page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-risk-level"]')).toContainText('High');
  });

  test('should export selected merchants', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Select first two merchants
    await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-checkbox"]').check();
    await page.locator('[data-testid="merchant-item"]').nth(1).locator('[data-testid="merchant-checkbox"]').check();
    
    // Set up download promise
    const downloadPromise = page.waitForEvent('download');
    
    // Click bulk export
    await page.locator('[data-testid="bulk-export"]').click();
    
    // Wait for download
    const download = await downloadPromise;
    
    // Check download filename
    expect(download.suggestedFilename()).toMatch(/bulk-merchants.*\.csv/);
  });

  test('should deactivate selected merchants', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Select first merchant
    await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-checkbox"]').check();
    
    // Click bulk deactivate
    await page.locator('[data-testid="bulk-deactivate"]').click();
    
    // Wait for confirmation modal
    await page.waitForSelector('[data-testid="deactivation-confirmation"]');
    
    // Confirm deactivation
    await page.locator('[data-testid="confirm-deactivation"]').click();
    
    // Wait for operation to complete
    await page.waitForSelector('[data-testid="bulk-progress"]', { state: 'hidden' });
    
    // Check that selected merchant is deactivated
    await expect(page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-portfolio-type"]')).toContainText('Deactivated');
  });

  test('should display progress tracking during bulk operations', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Select multiple merchants
    await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-checkbox"]').check();
    await page.locator('[data-testid="merchant-item"]').nth(1).locator('[data-testid="merchant-checkbox"]').check();
    await page.locator('[data-testid="merchant-item"]').nth(2).locator('[data-testid="merchant-checkbox"]').check();
    
    // Click bulk update portfolio type
    await page.locator('[data-testid="bulk-update-portfolio-type"]').click();
    
    // Wait for modal
    await page.waitForSelector('[data-testid="portfolio-type-modal"]');
    
    // Select portfolio type
    await page.locator('[data-testid="portfolio-type-pending"]').click();
    
    // Confirm update
    await page.locator('[data-testid="confirm-bulk-update"]').click();
    
    // Check that progress tracking is displayed
    await expect(page.locator('[data-testid="bulk-progress"]')).toBeVisible();
    await expect(page.locator('[data-testid="progress-bar"]')).toBeVisible();
    await expect(page.locator('[data-testid="progress-text"]')).toBeVisible();
    
    // Wait for operation to complete
    await page.waitForSelector('[data-testid="bulk-progress"]', { state: 'hidden' });
  });

  test('should allow pausing and resuming bulk operations', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Select multiple merchants
    await page.locator('[data-testid="bulk-select-all"]').click();
    
    // Click bulk update portfolio type
    await page.locator('[data-testid="bulk-update-portfolio-type"]').click();
    
    // Wait for modal
    await page.waitForSelector('[data-testid="portfolio-type-modal"]');
    
    // Select portfolio type
    await page.locator('[data-testid="portfolio-type-pending"]').click();
    
    // Confirm update
    await page.locator('[data-testid="confirm-bulk-update"]').click();
    
    // Wait for progress to start
    await page.waitForSelector('[data-testid="bulk-progress"]');
    
    // Click pause button
    await page.locator('[data-testid="pause-operation"]').click();
    
    // Check that operation is paused
    await expect(page.locator('[data-testid="operation-status"]')).toContainText('Paused');
    
    // Click resume button
    await page.locator('[data-testid="resume-operation"]').click();
    
    // Check that operation is resumed
    await expect(page.locator('[data-testid="operation-status"]')).toContainText('In Progress');
    
    // Wait for operation to complete
    await page.waitForSelector('[data-testid="bulk-progress"]', { state: 'hidden' });
  });

  test('should display operation history', async ({ page }) => {
    // Check operation history section
    await expect(page.locator('[data-testid="operation-history"]')).toBeVisible();
    
    // Check operation history table
    const historyTable = page.locator('[data-testid="operation-history-table"]');
    await expect(historyTable).toBeVisible();
    
    // Check table headers
    await expect(page.locator('[data-testid="operation-timestamp-header"]')).toBeVisible();
    await expect(page.locator('[data-testid="operation-type-header"]')).toBeVisible();
    await expect(page.locator('[data-testid="operation-count-header"]')).toBeVisible();
    await expect(page.locator('[data-testid="operation-status-header"]')).toBeVisible();
    
    // Check that there are operation history entries
    const historyRows = page.locator('[data-testid="operation-history-row"]');
    const count = await historyRows.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should handle bulk operation errors gracefully', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Select first merchant
    await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-checkbox"]').check();
    
    // Mock network error for bulk operation
    await page.route('**/api/merchants/bulk-update', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal server error' })
      });
    });
    
    // Click bulk update portfolio type
    await page.locator('[data-testid="bulk-update-portfolio-type"]').click();
    
    // Wait for modal
    await page.waitForSelector('[data-testid="portfolio-type-modal"]');
    
    // Select portfolio type
    await page.locator('[data-testid="portfolio-type-pending"]').click();
    
    // Confirm update
    await page.locator('[data-testid="confirm-bulk-update"]').click();
    
    // Check that error message is displayed
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-message"]')).toContainText('Operation failed');
  });

  test('should validate bulk operation requirements', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Try to perform bulk operation without selecting any merchants
    await page.locator('[data-testid="bulk-update-portfolio-type"]').click();
    
    // Check that validation message is displayed
    await expect(page.locator('[data-testid="validation-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="validation-message"]')).toContainText('Please select at least one merchant');
  });

  test('should be responsive on mobile devices', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Check that all elements are still visible and accessible
    await expect(page.locator('[data-testid="merchant-list"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-select-all"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-actions"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-progress"]')).toBeVisible();
    
    // Check that bulk actions are accessible on mobile
    await expect(page.locator('[data-testid="bulk-update-portfolio-type"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-update-risk-level"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-export"]')).toBeVisible();
  });

  test('should handle large number of merchants efficiently', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Select all merchants
    await page.locator('[data-testid="bulk-select-all"]').click();
    
    // Check that all merchants are selected
    const totalMerchants = await page.locator('[data-testid="merchant-item"]').count();
    await expect(page.locator('[data-testid="selected-count"]')).toContainText(totalMerchants.toString());
    
    // Perform bulk operation
    await page.locator('[data-testid="bulk-update-portfolio-type"]').click();
    await page.waitForSelector('[data-testid="portfolio-type-modal"]');
    await page.locator('[data-testid="portfolio-type-pending"]').click();
    await page.locator('[data-testid="confirm-bulk-update"]').click();
    
    // Wait for operation to complete
    await page.waitForSelector('[data-testid="bulk-progress"]', { state: 'hidden' });
    
    // Check that operation completed successfully
    await expect(page.locator('[data-testid="success-message"]')).toBeVisible();
  });
});
