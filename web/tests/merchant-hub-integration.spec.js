// @ts-check
const { test, expect } = require('@playwright/test');

test.describe('Merchant Hub Integration', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to merchant hub integration page
    await page.goto('/merchant-hub-integration.html');
    await page.waitForLoadState('networkidle');
    
    // Wait for page to load completely
    await page.waitForSelector('[data-testid="hub-integration"]', { timeout: 10000 });
  });

  test('should display hub integration page with all required elements', async ({ page }) => {
    // Check page title
    await expect(page).toHaveTitle(/Merchant Hub Integration/);
    
    // Check main heading
    await expect(page.locator('h1')).toContainText('Merchant Hub Integration');
    
    // Check hub integration container
    await expect(page.locator('[data-testid="hub-integration"]')).toBeVisible();
    
    // Check navigation elements
    await expect(page.locator('[data-testid="main-navigation"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-context"]')).toBeVisible();
    
    // Check merchant selection
    await expect(page.locator('[data-testid="merchant-selector"]')).toBeVisible();
    await expect(page.locator('[data-testid="current-merchant"]')).toBeVisible();
    
    // Check dashboard sections
    await expect(page.locator('[data-testid="dashboard-overview"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-portfolio"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-dashboard"]')).toBeVisible();
    await expect(page.locator('[data-testid="compliance-dashboard"]')).toBeVisible();
  });

  test('should display mock data warning', async ({ page }) => {
    // Check for mock data warning
    await expect(page.locator('[data-testid="mock-data-warning"]')).toBeVisible();
    await expect(page.locator('[data-testid="mock-data-warning"]')).toContainText('Mock Data');
  });

  test('should display main navigation with all required links', async ({ page }) => {
    // Check main navigation
    await expect(page.locator('[data-testid="main-navigation"]')).toBeVisible();
    
    // Check navigation links
    await expect(page.locator('[data-testid="nav-dashboard"]')).toBeVisible();
    await expect(page.locator('[data-testid="nav-merchants"]')).toBeVisible();
    await expect(page.locator('[data-testid="nav-risk"]')).toBeVisible();
    await expect(page.locator('[data-testid="nav-compliance"]')).toBeVisible();
    await expect(page.locator('[data-testid="nav-reports"]')).toBeVisible();
    await expect(page.locator('[data-testid="nav-settings"]')).toBeVisible();
  });

  test('should display merchant context information', async ({ page }) => {
    // Wait for merchant context to load
    await page.waitForSelector('[data-testid="merchant-context"]', { timeout: 10000 });
    
    // Check merchant context elements
    await expect(page.locator('[data-testid="current-merchant"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-name"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-status"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-risk-level"]')).toBeVisible();
    
    // Check merchant switching controls
    await expect(page.locator('[data-testid="merchant-selector"]')).toBeVisible();
    await expect(page.locator('[data-testid="switch-merchant"]')).toBeVisible();
  });

  test('should allow switching between merchants', async ({ page }) => {
    // Wait for merchant context to load
    await page.waitForSelector('[data-testid="merchant-context"]', { timeout: 10000 });
    
    // Get current merchant name
    const currentMerchantName = await page.locator('[data-testid="merchant-name"]').textContent();
    
    // Click merchant selector
    await page.locator('[data-testid="merchant-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    
    // Select different merchant
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for merchant context to update
    await page.waitForTimeout(1000);
    
    // Check that merchant name has changed
    const newMerchantName = await page.locator('[data-testid="merchant-name"]').textContent();
    expect(newMerchantName).not.toBe(currentMerchantName);
  });

  test('should update dashboard content when switching merchants', async ({ page }) => {
    // Wait for merchant context to load
    await page.waitForSelector('[data-testid="merchant-context"]', { timeout: 10000 });
    
    // Get current dashboard content
    const currentDashboardContent = await page.locator('[data-testid="dashboard-overview"]').textContent();
    
    // Switch merchant
    await page.locator('[data-testid="merchant-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Wait for dashboard to update
    await page.waitForTimeout(1000);
    
    // Check that dashboard content has updated
    const newDashboardContent = await page.locator('[data-testid="dashboard-overview"]').textContent();
    expect(newDashboardContent).not.toBe(currentDashboardContent);
  });

  test('should maintain merchant context across navigation', async ({ page }) => {
    // Wait for merchant context to load
    await page.waitForSelector('[data-testid="merchant-context"]', { timeout: 10000 });
    
    // Get current merchant name
    const merchantName = await page.locator('[data-testid="merchant-name"]').textContent();
    
    // Navigate to risk dashboard
    await page.locator('[data-testid="nav-risk"]').click();
    await page.waitForLoadState('networkidle');
    
    // Check that merchant context is maintained
    await expect(page.locator('[data-testid="merchant-name"]')).toContainText(merchantName);
    
    // Navigate to compliance dashboard
    await page.locator('[data-testid="nav-compliance"]').click();
    await page.waitForLoadState('networkidle');
    
    // Check that merchant context is still maintained
    await expect(page.locator('[data-testid="merchant-name"]')).toContainText(merchantName);
  });

  test('should display dashboard overview with merchant-specific data', async ({ page }) => {
    // Wait for dashboard to load
    await page.waitForSelector('[data-testid="dashboard-overview"]', { timeout: 10000 });
    
    // Check dashboard overview elements
    await expect(page.locator('[data-testid="overview-cards"]')).toBeVisible();
    await expect(page.locator('[data-testid="transaction-summary"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-summary"]')).toBeVisible();
    await expect(page.locator('[data-testid="compliance-summary"]')).toBeVisible();
    
    // Check that data is displayed
    await expect(page.locator('[data-testid="transaction-count"]')).toBeVisible();
    await expect(page.locator('[data-testid="transaction-volume"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-score"]')).toBeVisible();
    await expect(page.locator('[data-testid="compliance-score"]')).toBeVisible();
  });

  test('should display merchant portfolio section', async ({ page }) => {
    // Wait for merchant portfolio to load
    await page.waitForSelector('[data-testid="merchant-portfolio"]', { timeout: 10000 });
    
    // Check merchant portfolio elements
    await expect(page.locator('[data-testid="portfolio-status"]')).toBeVisible();
    await expect(page.locator('[data-testid="portfolio-type"]')).toBeVisible();
    await expect(page.locator('[data-testid="onboarding-date"]')).toBeVisible();
    await expect(page.locator('[data-testid="last-updated"]')).toBeVisible();
    
    // Check portfolio actions
    await expect(page.locator('[data-testid="view-portfolio"]')).toBeVisible();
    await expect(page.locator('[data-testid="edit-portfolio"]')).toBeVisible();
  });

  test('should display risk dashboard section', async ({ page }) => {
    // Wait for risk dashboard to load
    await page.waitForSelector('[data-testid="risk-dashboard"]', { timeout: 10000 });
    
    // Check risk dashboard elements
    await expect(page.locator('[data-testid="risk-level"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-score"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-factors"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-trend"]')).toBeVisible();
    
    // Check risk actions
    await expect(page.locator('[data-testid="view-risk-details"]')).toBeVisible();
    await expect(page.locator('[data-testid="update-risk-assessment"]')).toBeVisible();
  });

  test('should display compliance dashboard section', async ({ page }) => {
    // Wait for compliance dashboard to load
    await page.waitForSelector('[data-testid="compliance-dashboard"]', { timeout: 10000 });
    
    // Check compliance dashboard elements
    await expect(page.locator('[data-testid="compliance-status"]')).toBeVisible();
    await expect(page.locator('[data-testid="compliance-score"]')).toBeVisible();
    await expect(page.locator('[data-testid="compliance-requirements"]')).toBeVisible();
    await expect(page.locator('[data-testid="compliance-alerts"]')).toBeVisible();
    
    // Check compliance actions
    await expect(page.locator('[data-testid="view-compliance-details"]')).toBeVisible();
    await expect(page.locator('[data-testid="run-compliance-check"]')).toBeVisible();
  });

  test('should handle merchant switching with loading states', async ({ page }) => {
    // Wait for merchant context to load
    await page.waitForSelector('[data-testid="merchant-context"]', { timeout: 10000 });
    
    // Click merchant selector
    await page.locator('[data-testid="merchant-selector"]').click();
    await page.waitForSelector('[data-testid="merchant-dropdown"]');
    
    // Select different merchant
    await page.locator('[data-testid="merchant-option"]').nth(1).click();
    
    // Check that loading indicator is shown
    await expect(page.locator('[data-testid="loading-indicator"]')).toBeVisible();
    
    // Wait for loading to complete
    await page.waitForSelector('[data-testid="loading-indicator"]', { state: 'hidden' });
    
    // Check that new merchant context is displayed
    await expect(page.locator('[data-testid="merchant-name"]')).toBeVisible();
  });

  test('should maintain session state across page refreshes', async ({ page }) => {
    // Wait for merchant context to load
    await page.waitForSelector('[data-testid="merchant-context"]', { timeout: 10000 });
    
    // Get current merchant name
    const merchantName = await page.locator('[data-testid="merchant-name"]').textContent();
    
    // Refresh page
    await page.reload();
    await page.waitForLoadState('networkidle');
    
    // Wait for merchant context to load again
    await page.waitForSelector('[data-testid="merchant-context"]', { timeout: 10000 });
    
    // Check that merchant context is restored
    await expect(page.locator('[data-testid="merchant-name"]')).toContainText(merchantName);
  });

  test('should handle merchant not found errors gracefully', async ({ page }) => {
    // Mock network error for merchant data
    await page.route('**/api/merchants/*', route => {
      route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Merchant not found' })
      });
    });
    
    // Navigate to page
    await page.goto('/merchant-hub-integration.html?merchantId=non-existent');
    await page.waitForLoadState('networkidle');
    
    // Check that error message is displayed
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-message"]')).toContainText('Merchant not found');
    
    // Check that merchant selector is still available
    await expect(page.locator('[data-testid="merchant-selector"]')).toBeVisible();
  });

  test('should be responsive on mobile devices', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    // Wait for page to load
    await page.waitForSelector('[data-testid="hub-integration"]', { timeout: 10000 });
    
    // Check that all elements are still visible and accessible
    await expect(page.locator('[data-testid="main-navigation"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-context"]')).toBeVisible();
    await expect(page.locator('[data-testid="dashboard-overview"]')).toBeVisible();
    
    // Check that navigation is accessible on mobile
    await expect(page.locator('[data-testid="nav-dashboard"]')).toBeVisible();
    await expect(page.locator('[data-testid="nav-merchants"]')).toBeVisible();
    await expect(page.locator('[data-testid="nav-risk"]')).toBeVisible();
  });

  test('should handle concurrent user sessions', async ({ page }) => {
    // Wait for merchant context to load
    await page.waitForSelector('[data-testid="merchant-context"]', { timeout: 10000 });
    
    // Check that session information is displayed
    await expect(page.locator('[data-testid="session-info"]')).toBeVisible();
    await expect(page.locator('[data-testid="session-user"]')).toBeVisible();
    await expect(page.locator('[data-testid="session-timeout"]')).toBeVisible();
    
    // Check that session management controls are available
    await expect(page.locator('[data-testid="extend-session"]')).toBeVisible();
    await expect(page.locator('[data-testid="logout"]')).toBeVisible();
  });

  test('should display breadcrumb navigation', async ({ page }) => {
    // Wait for page to load
    await page.waitForSelector('[data-testid="hub-integration"]', { timeout: 10000 });
    
    // Check breadcrumb navigation
    await expect(page.locator('[data-testid="breadcrumb"]')).toBeVisible();
    
    // Check breadcrumb items
    await expect(page.locator('[data-testid="breadcrumb-home"]')).toBeVisible();
    await expect(page.locator('[data-testid="breadcrumb-merchants"]')).toBeVisible();
    await expect(page.locator('[data-testid="breadcrumb-current"]')).toBeVisible();
    
    // Check that breadcrumb items are clickable
    await expect(page.locator('[data-testid="breadcrumb-home"]')).toBeEnabled();
    await expect(page.locator('[data-testid="breadcrumb-merchants"]')).toBeEnabled();
  });

  test('should handle navigation between different dashboard sections', async ({ page }) => {
    // Wait for page to load
    await page.waitForSelector('[data-testid="hub-integration"]', { timeout: 10000 });
    
    // Navigate to risk dashboard
    await page.locator('[data-testid="nav-risk"]').click();
    await page.waitForLoadState('networkidle');
    
    // Check that risk dashboard is active
    await expect(page.locator('[data-testid="nav-risk"]')).toHaveClass(/active/);
    
    // Navigate to compliance dashboard
    await page.locator('[data-testid="nav-compliance"]').click();
    await page.waitForLoadState('networkidle');
    
    // Check that compliance dashboard is active
    await expect(page.locator('[data-testid="nav-compliance"]')).toHaveClass(/active/);
    
    // Navigate back to main dashboard
    await page.locator('[data-testid="nav-dashboard"]').click();
    await page.waitForLoadState('networkidle');
    
    // Check that main dashboard is active
    await expect(page.locator('[data-testid="nav-dashboard"]')).toHaveClass(/active/);
  });

  test('should display real-time updates for merchant data', async ({ page }) => {
    // Wait for merchant context to load
    await page.waitForSelector('[data-testid="merchant-context"]', { timeout: 10000 });
    
    // Check that real-time indicators are displayed
    await expect(page.locator('[data-testid="real-time-indicator"]')).toBeVisible();
    await expect(page.locator('[data-testid="last-updated"]')).toBeVisible();
    
    // Check that auto-refresh is enabled
    await expect(page.locator('[data-testid="auto-refresh"]')).toBeVisible();
    
    // Check that manual refresh button is available
    await expect(page.locator('[data-testid="refresh-data"]')).toBeVisible();
  });

  test('should handle network connectivity issues gracefully', async ({ page }) => {
    // Wait for page to load
    await page.waitForSelector('[data-testid="hub-integration"]', { timeout: 10000 });
    
    // Simulate network disconnection
    await page.context().setOffline(true);
    
    // Check that offline indicator is displayed
    await expect(page.locator('[data-testid="offline-indicator"]')).toBeVisible();
    
    // Check that cached data is still displayed
    await expect(page.locator('[data-testid="merchant-context"]')).toBeVisible();
    
    // Restore network connection
    await page.context().setOffline(false);
    
    // Check that online indicator is displayed
    await expect(page.locator('[data-testid="online-indicator"]')).toBeVisible();
  });
});
