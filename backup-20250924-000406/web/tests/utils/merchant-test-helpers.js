/**
 * Merchant Test Helpers
 * 
 * Common utilities and helper functions for merchant-related Playwright tests
 */

class MerchantTestHelpers {
  constructor(page) {
    this.page = page;
  }

  /**
   * Navigate to merchant portfolio page and wait for it to load
   */
  async navigateToMerchantPortfolio() {
    await this.page.goto('/merchant-portfolio.html');
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForSelector('[data-testid="merchant-list"]', { timeout: 10000 });
  }

  /**
   * Navigate to merchant detail page for a specific merchant
   */
  async navigateToMerchantDetail(merchantId) {
    await this.page.goto(`/merchant-detail.html?id=${merchantId}`);
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForSelector('[data-testid="merchant-info"]', { timeout: 10000 });
  }

  /**
   * Navigate to merchant bulk operations page
   */
  async navigateToBulkOperations() {
    await this.page.goto('/merchant-bulk-operations.html');
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForSelector('[data-testid="merchant-list"]', { timeout: 10000 });
  }

  /**
   * Navigate to merchant comparison page
   */
  async navigateToMerchantComparison() {
    await this.page.goto('/merchant-comparison.html');
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForSelector('[data-testid="comparison-container"]', { timeout: 10000 });
  }

  /**
   * Navigate to merchant hub integration page
   */
  async navigateToHubIntegration() {
    await this.page.goto('/merchant-hub-integration.html');
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForSelector('[data-testid="hub-integration"]', { timeout: 10000 });
  }

  /**
   * Select a merchant from the merchant selector dropdown
   */
  async selectMerchant(selectorId, optionIndex = 0) {
    await this.page.locator(`[data-testid="${selectorId}"]`).click();
    await this.page.waitForSelector('[data-testid="merchant-dropdown"]');
    await this.page.locator('[data-testid="merchant-option"]').nth(optionIndex).click();
    await this.page.waitForTimeout(1000); // Wait for selection to process
  }

  /**
   * Search for a merchant by name
   */
  async searchMerchant(searchTerm) {
    await this.page.locator('[data-testid="merchant-search-input"]').fill(searchTerm);
    await this.page.waitForTimeout(1000); // Wait for search to process
  }

  /**
   * Clear merchant search
   */
  async clearMerchantSearch() {
    await this.page.locator('[data-testid="merchant-search-clear"]').click();
    await this.page.waitForTimeout(1000); // Wait for search to clear
  }

  /**
   * Filter merchants by portfolio type
   */
  async filterByPortfolioType(portfolioType) {
    await this.page.locator(`[data-testid="portfolio-type-${portfolioType.toLowerCase()}"]`).click();
    await this.page.waitForTimeout(1000); // Wait for filter to apply
  }

  /**
   * Filter merchants by risk level
   */
  async filterByRiskLevel(riskLevel) {
    await this.page.locator(`[data-testid="risk-level-${riskLevel.toLowerCase()}"]`).click();
    await this.page.waitForTimeout(1000); // Wait for filter to apply
  }

  /**
   * Select merchants for bulk operations
   */
  async selectMerchantsForBulkOperation(count = 2) {
    for (let i = 0; i < count; i++) {
      await this.page.locator('[data-testid="merchant-item"]').nth(i).locator('[data-testid="merchant-checkbox"]').check();
    }
  }

  /**
   * Select all merchants for bulk operations
   */
  async selectAllMerchants() {
    await this.page.locator('[data-testid="bulk-select-all"]').click();
  }

  /**
   * Deselect all merchants
   */
  async deselectAllMerchants() {
    await this.page.locator('[data-testid="bulk-deselect-all"]').click();
  }

  /**
   * Perform bulk portfolio type update
   */
  async performBulkPortfolioTypeUpdate(portfolioType) {
    await this.page.locator('[data-testid="bulk-update-portfolio-type"]').click();
    await this.page.waitForSelector('[data-testid="portfolio-type-modal"]');
    await this.page.locator(`[data-testid="portfolio-type-${portfolioType.toLowerCase()}"]`).click();
    await this.page.locator('[data-testid="confirm-bulk-update"]').click();
    await this.page.waitForSelector('[data-testid="bulk-progress"]', { state: 'hidden' });
  }

  /**
   * Perform bulk risk level update
   */
  async performBulkRiskLevelUpdate(riskLevel) {
    await this.page.locator('[data-testid="bulk-update-risk-level"]').click();
    await this.page.waitForSelector('[data-testid="risk-level-modal"]');
    await this.page.locator(`[data-testid="risk-level-${riskLevel.toLowerCase()}"]`).click();
    await this.page.locator('[data-testid="confirm-bulk-update"]').click();
    await this.page.waitForSelector('[data-testid="bulk-progress"]', { state: 'hidden' });
  }

  /**
   * Export merchant data
   */
  async exportMerchantData() {
    const downloadPromise = this.page.waitForEvent('download');
    await this.page.locator('[data-testid="export-merchants"]').click();
    return await downloadPromise;
  }

  /**
   * Navigate to next page in pagination
   */
  async navigateToNextPage() {
    const nextButton = this.page.locator('[data-testid="pagination-next"]');
    if (await nextButton.isEnabled()) {
      await nextButton.click();
      await this.page.waitForTimeout(1000);
    }
  }

  /**
   * Navigate to previous page in pagination
   */
  async navigateToPreviousPage() {
    const prevButton = this.page.locator('[data-testid="pagination-prev"]');
    if (await prevButton.isEnabled()) {
      await prevButton.click();
      await this.page.waitForTimeout(1000);
    }
  }

  /**
   * Get merchant count from the current page
   */
  async getMerchantCount() {
    return await this.page.locator('[data-testid="merchant-item"]').count();
  }

  /**
   * Get selected merchant count
   */
  async getSelectedMerchantCount() {
    const selectedCountText = await this.page.locator('[data-testid="selected-count"]').textContent();
    return parseInt(selectedCountText.match(/\d+/)[0]);
  }

  /**
   * Get merchant name by index
   */
  async getMerchantName(index = 0) {
    return await this.page.locator('[data-testid="merchant-item"]').nth(index).locator('[data-testid="merchant-name"]').textContent();
  }

  /**
   * Get merchant portfolio type by index
   */
  async getMerchantPortfolioType(index = 0) {
    return await this.page.locator('[data-testid="merchant-item"]').nth(index).locator('[data-testid="merchant-portfolio-type"]').textContent();
  }

  /**
   * Get merchant risk level by index
   */
  async getMerchantRiskLevel(index = 0) {
    return await this.page.locator('[data-testid="merchant-item"]').nth(index).locator('[data-testid="merchant-risk-level"]').textContent();
  }

  /**
   * Wait for loading indicator to disappear
   */
  async waitForLoadingToComplete() {
    await this.page.waitForSelector('[data-testid="loading-indicator"]', { state: 'hidden', timeout: 10000 });
  }

  /**
   * Check if mock data warning is displayed
   */
  async isMockDataWarningVisible() {
    return await this.page.locator('[data-testid="mock-data-warning"]').isVisible();
  }

  /**
   * Mock network error for testing error handling
   */
  async mockNetworkError(urlPattern, status = 500) {
    await this.page.route(urlPattern, route => {
      route.fulfill({
        status,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Mock network error' })
      });
    });
  }

  /**
   * Restore network (remove mocks)
   */
  async restoreNetwork() {
    await this.page.unroute('**/*');
  }

  /**
   * Set mobile viewport for responsive testing
   */
  async setMobileViewport() {
    await this.page.setViewportSize({ width: 375, height: 667 });
  }

  /**
   * Set tablet viewport for responsive testing
   */
  async setTabletViewport() {
    await this.page.setViewportSize({ width: 768, height: 1024 });
  }

  /**
   * Set desktop viewport for responsive testing
   */
  async setDesktopViewport() {
    await this.page.setViewportSize({ width: 1280, height: 720 });
  }

  /**
   * Take screenshot with timestamp
   */
  async takeScreenshot(name) {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    await this.page.screenshot({ 
      path: `test-results/merchant-artifacts/screenshot-${name}-${timestamp}.png`,
      fullPage: true
    });
  }

  /**
   * Wait for element to be visible with custom timeout
   */
  async waitForElement(selector, timeout = 10000) {
    await this.page.waitForSelector(selector, { timeout });
  }

  /**
   * Check if element exists without waiting
   */
  async elementExists(selector) {
    try {
      await this.page.locator(selector).first().waitFor({ timeout: 1000 });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get element text content safely
   */
  async getElementText(selector) {
    try {
      return await this.page.locator(selector).textContent();
    } catch {
      return null;
    }
  }

  /**
   * Check if element is visible safely
   */
  async isElementVisible(selector) {
    try {
      return await this.page.locator(selector).isVisible();
    } catch {
      return false;
    }
  }

  /**
   * Check if element is enabled safely
   */
  async isElementEnabled(selector) {
    try {
      return await this.page.locator(selector).isEnabled();
    } catch {
      return false;
    }
  }

  /**
   * Check if element is checked safely
   */
  async isElementChecked(selector) {
    try {
      return await this.page.locator(selector).isChecked();
    } catch {
      return false;
    }
  }

  /**
   * Get element attribute safely
   */
  async getElementAttribute(selector, attribute) {
    try {
      return await this.page.locator(selector).getAttribute(attribute);
    } catch {
      return null;
    }
  }

  /**
   * Get element CSS class safely
   */
  async getElementClass(selector) {
    try {
      return await this.page.locator(selector).getAttribute('class');
    } catch {
      return null;
    }
  }

  /**
   * Wait for network to be idle
   */
  async waitForNetworkIdle() {
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Wait for DOM content to be loaded
   */
  async waitForDOMContentLoaded() {
    await this.page.waitForLoadState('domcontentloaded');
  }

  /**
   * Wait for specific amount of time
   */
  async wait(ms) {
    await this.page.waitForTimeout(ms);
  }

  /**
   * Scroll element into view
   */
  async scrollIntoView(selector) {
    await this.page.locator(selector).scrollIntoViewIfNeeded();
  }

  /**
   * Click element and wait for navigation
   */
  async clickAndWaitForNavigation(selector) {
    await Promise.all([
      this.page.waitForNavigation(),
      this.page.locator(selector).click()
    ]);
  }

  /**
   * Fill form field safely
   */
  async fillField(selector, value) {
    await this.page.locator(selector).clear();
    await this.page.locator(selector).fill(value);
  }

  /**
   * Select dropdown option safely
   */
  async selectOption(selector, value) {
    await this.page.locator(selector).selectOption(value);
  }

  /**
   * Check checkbox safely
   */
  async checkCheckbox(selector) {
    await this.page.locator(selector).check();
  }

  /**
   * Uncheck checkbox safely
   */
  async uncheckCheckbox(selector) {
    await this.page.locator(selector).uncheck();
  }

  /**
   * Hover over element safely
   */
  async hover(selector) {
    await this.page.locator(selector).hover();
  }

  /**
   * Double click element safely
   */
  async doubleClick(selector) {
    await this.page.locator(selector).dblclick();
  }

  /**
   * Right click element safely
   */
  async rightClick(selector) {
    await this.page.locator(selector).click({ button: 'right' });
  }

  /**
   * Press key safely
   */
  async pressKey(key) {
    await this.page.keyboard.press(key);
  }

  /**
   * Type text safely
   */
  async typeText(selector, text) {
    await this.page.locator(selector).type(text);
  }

  /**
   * Clear text field safely
   */
  async clearText(selector) {
    await this.page.locator(selector).clear();
  }

  /**
   * Focus element safely
   */
  async focus(selector) {
    await this.page.locator(selector).focus();
  }

  /**
   * Blur element safely
   */
  async blur(selector) {
    await this.page.locator(selector).blur();
  }
}

module.exports = MerchantTestHelpers;
