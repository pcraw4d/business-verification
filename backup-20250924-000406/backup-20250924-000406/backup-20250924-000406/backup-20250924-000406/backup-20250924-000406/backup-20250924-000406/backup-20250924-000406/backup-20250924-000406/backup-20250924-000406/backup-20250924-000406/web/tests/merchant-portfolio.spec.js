// @ts-check
const { test, expect } = require('@playwright/test');

test.describe('Merchant Portfolio', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to merchant portfolio page
    await page.goto('/merchant-portfolio.html');
    
    // Wait for page to load completely
    await page.waitForLoadState('networkidle');
    
    // Wait for merchant data to load
    await page.waitForSelector('[data-testid="merchant-list"]', { timeout: 10000 });
  });

  test('should display merchant portfolio page with all required elements', async ({ page }) => {
    // Check page title
    await expect(page).toHaveTitle(/Merchant Portfolio/);
    
    // Check main heading
    await expect(page.locator('h1')).toContainText('Merchant Portfolio');
    
    // Check search functionality
    await expect(page.locator('[data-testid="merchant-search"]')).toBeVisible();
    
    // Check portfolio type filter
    await expect(page.locator('[data-testid="portfolio-type-filter"]')).toBeVisible();
    
    // Check risk level filter
    await expect(page.locator('[data-testid="risk-level-filter"]')).toBeVisible();
    
    // Check merchant list
    await expect(page.locator('[data-testid="merchant-list"]')).toBeVisible();
    
    // Check pagination controls
    await expect(page.locator('[data-testid="pagination"]')).toBeVisible();
    
    // Check bulk operations section
    await expect(page.locator('[data-testid="bulk-operations"]')).toBeVisible();
  });

  test('should display mock data warning', async ({ page }) => {
    // Check for mock data warning
    await expect(page.locator('[data-testid="mock-data-warning"]')).toBeVisible();
    await expect(page.locator('[data-testid="mock-data-warning"]')).toContainText('Mock Data');
  });

  test('should load and display merchant list', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Check that merchants are displayed
    const merchantItems = page.locator('[data-testid="merchant-item"]');
    await expect(merchantItems).toHaveCount.greaterThan(0);
    
    // Check first merchant has required fields
    const firstMerchant = merchantItems.first();
    await expect(firstMerchant.locator('[data-testid="merchant-name"]')).toBeVisible();
    await expect(firstMerchant.locator('[data-testid="merchant-industry"]')).toBeVisible();
    await expect(firstMerchant.locator('[data-testid="merchant-portfolio-type"]')).toBeVisible();
    await expect(firstMerchant.locator('[data-testid="merchant-risk-level"]')).toBeVisible();
  });

  test('should filter merchants by portfolio type', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Get initial merchant count
    const initialCount = await page.locator('[data-testid="merchant-item"]').count();
    
    // Select "Onboarded" portfolio type
    await page.locator('[data-testid="portfolio-type-onboarded"]').click();
    
    // Wait for filter to apply
    await page.waitForTimeout(1000);
    
    // Check that filtered results are displayed
    const filteredCount = await page.locator('[data-testid="merchant-item"]').count();
    expect(filteredCount).toBeLessThanOrEqual(initialCount);
    
    // Verify all displayed merchants have "Onboarded" portfolio type
    const displayedMerchants = page.locator('[data-testid="merchant-item"]');
    const count = await displayedMerchants.count();
    
    for (let i = 0; i < count; i++) {
      const merchant = displayedMerchants.nth(i);
      await expect(merchant.locator('[data-testid="merchant-portfolio-type"]')).toContainText('Onboarded');
    }
  });

  test('should filter merchants by risk level', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Get initial merchant count
    const initialCount = await page.locator('[data-testid="merchant-item"]').count();
    
    // Select "High" risk level
    await page.locator('[data-testid="risk-level-high"]').click();
    
    // Wait for filter to apply
    await page.waitForTimeout(1000);
    
    // Check that filtered results are displayed
    const filteredCount = await page.locator('[data-testid="merchant-item"]').count();
    expect(filteredCount).toBeLessThanOrEqual(initialCount);
    
    // Verify all displayed merchants have "High" risk level
    const displayedMerchants = page.locator('[data-testid="merchant-item"]');
    const count = await displayedMerchants.count();
    
    for (let i = 0; i < count; i++) {
      const merchant = displayedMerchants.nth(i);
      await expect(merchant.locator('[data-testid="merchant-risk-level"]')).toContainText('High');
    }
  });

  test('should search merchants by name', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Get first merchant name for search
    const firstMerchantName = await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-name"]').textContent();
    
    // Type in search box
    await page.locator('[data-testid="merchant-search-input"]').fill(firstMerchantName);
    
    // Wait for search to apply
    await page.waitForTimeout(1000);
    
    // Check that search results are displayed
    const searchResults = page.locator('[data-testid="merchant-item"]');
    const count = await searchResults.count();
    
    // Verify all results contain the search term
    for (let i = 0; i < count; i++) {
      const merchant = searchResults.nth(i);
      const merchantName = await merchant.locator('[data-testid="merchant-name"]').textContent();
      expect(merchantName.toLowerCase()).toContain(firstMerchantName.toLowerCase());
    }
  });

  test('should clear search and show all merchants', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Get initial count
    const initialCount = await page.locator('[data-testid="merchant-item"]').count();
    
    // Search for something
    await page.locator('[data-testid="merchant-search-input"]').fill('test');
    await page.waitForTimeout(1000);
    
    // Clear search
    await page.locator('[data-testid="merchant-search-clear"]').click();
    await page.waitForTimeout(1000);
    
    // Check that all merchants are displayed again
    const finalCount = await page.locator('[data-testid="merchant-item"]').count();
    expect(finalCount).toBe(initialCount);
  });

  test('should navigate to merchant detail page when clicking on merchant', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Click on first merchant
    await page.locator('[data-testid="merchant-item"]').first().click();
    
    // Wait for navigation
    await page.waitForLoadState('networkidle');
    
    // Check that we're on merchant detail page
    await expect(page).toHaveURL(/merchant-detail\.html/);
    await expect(page.locator('h1')).toContainText('Merchant Details');
  });

  test('should display pagination controls', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Check pagination controls are visible
    await expect(page.locator('[data-testid="pagination"]')).toBeVisible();
    await expect(page.locator('[data-testid="pagination-prev"]')).toBeVisible();
    await expect(page.locator('[data-testid="pagination-next"]')).toBeVisible();
    await expect(page.locator('[data-testid="pagination-info"]')).toBeVisible();
  });

  test('should navigate between pages using pagination', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Check if next button is available
    const nextButton = page.locator('[data-testid="pagination-next"]');
    const isNextEnabled = await nextButton.isEnabled();
    
    if (isNextEnabled) {
      // Get first merchant name on current page
      const firstMerchantName = await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-name"]').textContent();
      
      // Click next page
      await nextButton.click();
      await page.waitForTimeout(1000);
      
      // Check that we're on a different page
      const newFirstMerchantName = await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-name"]').textContent();
      expect(newFirstMerchantName).not.toBe(firstMerchantName);
      
      // Click previous page
      await page.locator('[data-testid="pagination-prev"]').click();
      await page.waitForTimeout(1000);
      
      // Check that we're back to original page
      const backToFirstMerchantName = await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-name"]').textContent();
      expect(backToFirstMerchantName).toBe(firstMerchantName);
    }
  });

  test('should display bulk operations section', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Check bulk operations section
    await expect(page.locator('[data-testid="bulk-operations"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-select-all"]')).toBeVisible();
    await expect(page.locator('[data-testid="bulk-actions"]')).toBeVisible();
  });

  test('should select and deselect merchants for bulk operations', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Select first merchant
    await page.locator('[data-testid="merchant-item"]').first().locator('[data-testid="merchant-checkbox"]').check();
    
    // Check that bulk actions are enabled
    await expect(page.locator('[data-testid="bulk-actions"]')).toBeVisible();
    
    // Check selected count
    await expect(page.locator('[data-testid="selected-count"]')).toContainText('1');
    
    // Select all merchants
    await page.locator('[data-testid="bulk-select-all"]').check();
    
    // Check selected count updated
    const totalMerchants = await page.locator('[data-testid="merchant-item"]').count();
    await expect(page.locator('[data-testid="selected-count"]')).toContainText(totalMerchants.toString());
    
    // Deselect all
    await page.locator('[data-testid="bulk-select-all"]').uncheck();
    
    // Check that no merchants are selected
    await expect(page.locator('[data-testid="selected-count"]')).toContainText('0');
  });

  test('should export merchant data', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Set up download promise
    const downloadPromise = page.waitForEvent('download');
    
    // Click export button
    await page.locator('[data-testid="export-merchants"]').click();
    
    // Wait for download
    const download = await downloadPromise;
    
    // Check download filename
    expect(download.suggestedFilename()).toMatch(/merchants.*\.csv/);
  });

  test('should be responsive on mobile devices', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Check that page is responsive
    await expect(page.locator('[data-testid="merchant-list"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-search"]')).toBeVisible();
    
    // Check that filters are accessible on mobile
    await expect(page.locator('[data-testid="portfolio-type-filter"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-level-filter"]')).toBeVisible();
  });

  test('should handle empty search results gracefully', async ({ page }) => {
    // Wait for merchants to load
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Search for non-existent merchant
    await page.locator('[data-testid="merchant-search-input"]').fill('NonExistentMerchant12345');
    await page.waitForTimeout(1000);
    
    // Check that empty state is displayed
    await expect(page.locator('[data-testid="empty-state"]')).toBeVisible();
    await expect(page.locator('[data-testid="empty-state"]')).toContainText('No merchants found');
  });

  test('should display loading state while fetching merchants', async ({ page }) => {
    // Navigate to page and check for loading state
    await page.goto('/merchant-portfolio.html');
    
    // Check that loading indicator is shown initially
    await expect(page.locator('[data-testid="loading-indicator"]')).toBeVisible();
    
    // Wait for loading to complete
    await page.waitForSelector('[data-testid="merchant-list"]', { timeout: 10000 });
    
    // Check that loading indicator is hidden
    await expect(page.locator('[data-testid="loading-indicator"]')).not.toBeVisible();
  });
});
