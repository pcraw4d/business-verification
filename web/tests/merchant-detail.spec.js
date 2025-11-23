// @ts-check
const { test, expect } = require('@playwright/test');
const { setupAPIMocks } = require('./utils/api-mock-helpers');

test.describe('Merchant Detail Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Setup API mocks before navigation
    await setupAPIMocks(page);
    
    // Navigate to merchant portfolio first to get a merchant ID
    await page.goto('/merchant-portfolio.html');
    await page.waitForLoadState('networkidle');
    await page.waitForSelector('[data-testid="merchant-item"]', { timeout: 10000 });
    
    // Click on first merchant to navigate to detail page
    await page.locator('[data-testid="merchant-item"]').first().click();
    await page.waitForLoadState('networkidle');
  });

  test('should display merchant detail page with all required sections', async ({ page }) => {
    // Check page title
    await expect(page).toHaveTitle(/Merchant Details/);
    
    // Check main heading
    await expect(page.locator('h1')).toContainText('Merchant Details');
    
    // Check merchant information section
    await expect(page.locator('[data-testid="merchant-info"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-name"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-industry"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-address"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-phone"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-email"]')).toBeVisible();
    await expect(page.locator('[data-testid="merchant-website"]')).toBeVisible();
    
    // Check portfolio information
    await expect(page.locator('[data-testid="portfolio-info"]')).toBeVisible();
    await expect(page.locator('[data-testid="portfolio-type"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-level"]')).toBeVisible();
    await expect(page.locator('[data-testid="onboarding-date"]')).toBeVisible();
    await expect(page.locator('[data-testid="last-updated"]')).toBeVisible();
    
    // Check compliance section
    await expect(page.locator('[data-testid="compliance-section"]')).toBeVisible();
    await expect(page.locator('[data-testid="compliance-status"]')).toBeVisible();
    await expect(page.locator('[data-testid="compliance-score"]')).toBeVisible();
    
    // Check risk assessment section
    await expect(page.locator('[data-testid="risk-assessment"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-score"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-factors"]')).toBeVisible();
    
    // Check transaction history section
    await expect(page.locator('[data-testid="transaction-history"]')).toBeVisible();
    
    // Check audit log section
    await expect(page.locator('[data-testid="audit-log"]')).toBeVisible();
  });

  test('should display mock data warning', async ({ page }) => {
    // Check for mock data warning
    await expect(page.locator('[data-testid="mock-data-warning"]')).toBeVisible();
    await expect(page.locator('[data-testid="mock-data-warning"]')).toContainText('Mock Data');
  });

  test('should display merchant basic information correctly', async ({ page }) => {
    // Check that merchant name is displayed
    const merchantName = page.locator('[data-testid="merchant-name"]');
    await expect(merchantName).toBeVisible();
    await expect(merchantName).not.toBeEmpty();
    
    // Check that industry is displayed
    const industry = page.locator('[data-testid="merchant-industry"]');
    await expect(industry).toBeVisible();
    await expect(industry).not.toBeEmpty();
    
    // Check that address is displayed
    const address = page.locator('[data-testid="merchant-address"]');
    await expect(address).toBeVisible();
    await expect(address).not.toBeEmpty();
    
    // Check that phone is displayed (if available)
    const phone = page.locator('[data-testid="merchant-phone"]');
    if (await phone.isVisible()) {
      await expect(phone).not.toBeEmpty();
    }
    
    // Check that email is displayed (if available)
    const email = page.locator('[data-testid="merchant-email"]');
    if (await email.isVisible()) {
      await expect(email).not.toBeEmpty();
    }
    
    // Check that website is displayed (if available)
    const website = page.locator('[data-testid="merchant-website"]');
    if (await website.isVisible()) {
      await expect(website).not.toBeEmpty();
    }
  });

  test('should display portfolio information correctly', async ({ page }) => {
    // Check portfolio type
    const portfolioType = page.locator('[data-testid="portfolio-type"]');
    await expect(portfolioType).toBeVisible();
    await expect(portfolioType).not.toBeEmpty();
    
    // Check risk level
    const riskLevel = page.locator('[data-testid="risk-level"]');
    await expect(riskLevel).toBeVisible();
    await expect(riskLevel).not.toBeEmpty();
    
    // Check onboarding date
    const onboardingDate = page.locator('[data-testid="onboarding-date"]');
    await expect(onboardingDate).toBeVisible();
    await expect(onboardingDate).not.toBeEmpty();
    
    // Check last updated
    const lastUpdated = page.locator('[data-testid="last-updated"]');
    await expect(lastUpdated).toBeVisible();
    await expect(lastUpdated).not.toBeEmpty();
  });

  test('should display compliance information', async ({ page }) => {
    // Check compliance status
    const complianceStatus = page.locator('[data-testid="compliance-status"]');
    await expect(complianceStatus).toBeVisible();
    await expect(complianceStatus).not.toBeEmpty();
    
    // Check compliance score
    const complianceScore = page.locator('[data-testid="compliance-score"]');
    await expect(complianceScore).toBeVisible();
    await expect(complianceScore).not.toBeEmpty();
    
    // Check that compliance score is a number
    const scoreText = await complianceScore.textContent();
    expect(scoreText).toMatch(/\d+/);
  });

  test('should display risk assessment information', async ({ page }) => {
    // Check risk score
    const riskScore = page.locator('[data-testid="risk-score"]');
    await expect(riskScore).toBeVisible();
    await expect(riskScore).not.toBeEmpty();
    
    // Check that risk score is a number
    const scoreText = await riskScore.textContent();
    expect(scoreText).toMatch(/\d+/);
    
    // Check risk factors
    const riskFactors = page.locator('[data-testid="risk-factors"]');
    await expect(riskFactors).toBeVisible();
    
    // Check that risk factors list is not empty
    const riskFactorItems = page.locator('[data-testid="risk-factor-item"]');
    const count = await riskFactorItems.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should display transaction history', async ({ page }) => {
    // Check transaction history section
    const transactionHistory = page.locator('[data-testid="transaction-history"]');
    await expect(transactionHistory).toBeVisible();
    
    // Check transaction history table
    const transactionTable = page.locator('[data-testid="transaction-table"]');
    await expect(transactionTable).toBeVisible();
    
    // Check table headers
    await expect(page.locator('[data-testid="transaction-date-header"]')).toBeVisible();
    await expect(page.locator('[data-testid="transaction-amount-header"]')).toBeVisible();
    await expect(page.locator('[data-testid="transaction-status-header"]')).toBeVisible();
    
    // Check that there are transaction rows
    const transactionRows = page.locator('[data-testid="transaction-row"]');
    const count = await transactionRows.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should display audit log', async ({ page }) => {
    // Check audit log section
    const auditLog = page.locator('[data-testid="audit-log"]');
    await expect(auditLog).toBeVisible();
    
    // Check audit log table
    const auditTable = page.locator('[data-testid="audit-table"]');
    await expect(auditTable).toBeVisible();
    
    // Check table headers
    await expect(page.locator('[data-testid="audit-timestamp-header"]')).toBeVisible();
    await expect(page.locator('[data-testid="audit-action-header"]')).toBeVisible();
    await expect(page.locator('[data-testid="audit-user-header"]')).toBeVisible();
    await expect(page.locator('[data-testid="audit-details-header"]')).toBeVisible();
    
    // Check that there are audit log entries
    const auditRows = page.locator('[data-testid="audit-row"]');
    const count = await auditRows.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should have navigation back to portfolio', async ({ page }) => {
    // Check back to portfolio button
    const backButton = page.locator('[data-testid="back-to-portfolio"]');
    await expect(backButton).toBeVisible();
    
    // Click back button
    await backButton.click();
    
    // Check that we're back on portfolio page
    await expect(page).toHaveURL(/merchant-portfolio\.html/);
    await expect(page.locator('h1')).toContainText('Merchant Portfolio');
  });

  test('should have edit merchant functionality', async ({ page }) => {
    // Check edit button
    const editButton = page.locator('[data-testid="edit-merchant"]');
    await expect(editButton).toBeVisible();
    
    // Click edit button
    await editButton.click();
    
    // Check that edit form is displayed
    await expect(page.locator('[data-testid="edit-form"]')).toBeVisible();
    
    // Check form fields
    await expect(page.locator('[data-testid="edit-name"]')).toBeVisible();
    await expect(page.locator('[data-testid="edit-industry"]')).toBeVisible();
    await expect(page.locator('[data-testid="edit-address"]')).toBeVisible();
    await expect(page.locator('[data-testid="edit-phone"]')).toBeVisible();
    await expect(page.locator('[data-testid="edit-email"]')).toBeVisible();
    await expect(page.locator('[data-testid="edit-website"]')).toBeVisible();
    
    // Check save and cancel buttons
    await expect(page.locator('[data-testid="save-changes"]')).toBeVisible();
    await expect(page.locator('[data-testid="cancel-edit"]')).toBeVisible();
  });

  test('should save merchant changes', async ({ page }) => {
    // Click edit button
    await page.locator('[data-testid="edit-merchant"]').click();
    
    // Wait for edit form
    await page.waitForSelector('[data-testid="edit-form"]');
    
    // Update merchant name
    const nameField = page.locator('[data-testid="edit-name"]');
    await nameField.clear();
    await nameField.fill('Updated Merchant Name');
    
    // Save changes
    await page.locator('[data-testid="save-changes"]').click();
    
    // Wait for form to close
    await page.waitForSelector('[data-testid="edit-form"]', { state: 'hidden' });
    
    // Check that changes are reflected
    await expect(page.locator('[data-testid="merchant-name"]')).toContainText('Updated Merchant Name');
  });

  test('should cancel merchant changes', async ({ page }) => {
    // Get original merchant name
    const originalName = await page.locator('[data-testid="merchant-name"]').textContent();
    
    // Click edit button
    await page.locator('[data-testid="edit-merchant"]').click();
    
    // Wait for edit form
    await page.waitForSelector('[data-testid="edit-form"]');
    
    // Update merchant name
    const nameField = page.locator('[data-testid="edit-name"]');
    await nameField.clear();
    await nameField.fill('Modified Name');
    
    // Cancel changes
    await page.locator('[data-testid="cancel-edit"]').click();
    
    // Wait for form to close
    await page.waitForSelector('[data-testid="edit-form"]', { state: 'hidden' });
    
    // Check that original name is still displayed
    await expect(page.locator('[data-testid="merchant-name"]')).toContainText(originalName);
  });

  test('should display risk level with appropriate styling', async ({ page }) => {
    // Check risk level display
    const riskLevel = page.locator('[data-testid="risk-level"]');
    await expect(riskLevel).toBeVisible();
    
    // Check that risk level has appropriate CSS class
    const riskLevelText = await riskLevel.textContent();
    if (riskLevelText.toLowerCase().includes('high')) {
      await expect(riskLevel).toHaveClass(/risk-high/);
    } else if (riskLevelText.toLowerCase().includes('medium')) {
      await expect(riskLevel).toHaveClass(/risk-medium/);
    } else if (riskLevelText.toLowerCase().includes('low')) {
      await expect(riskLevel).toHaveClass(/risk-low/);
    }
  });

  test('should display portfolio type with appropriate styling', async ({ page }) => {
    // Check portfolio type display
    const portfolioType = page.locator('[data-testid="portfolio-type"]');
    await expect(portfolioType).toBeVisible();
    
    // Check that portfolio type has appropriate CSS class
    const portfolioTypeText = await portfolioType.textContent();
    if (portfolioTypeText.toLowerCase().includes('onboarded')) {
      await expect(portfolioType).toHaveClass(/portfolio-onboarded/);
    } else if (portfolioTypeText.toLowerCase().includes('pending')) {
      await expect(portfolioType).toHaveClass(/portfolio-pending/);
    } else if (portfolioTypeText.toLowerCase().includes('deactivated')) {
      await expect(portfolioType).toHaveClass(/portfolio-deactivated/);
    } else if (portfolioTypeText.toLowerCase().includes('prospective')) {
      await expect(portfolioType).toHaveClass(/portfolio-prospective/);
    }
  });

  test('should be responsive on mobile devices', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    // Check that all sections are still visible and accessible
    await expect(page.locator('[data-testid="merchant-info"]')).toBeVisible();
    await expect(page.locator('[data-testid="portfolio-info"]')).toBeVisible();
    await expect(page.locator('[data-testid="compliance-section"]')).toBeVisible();
    await expect(page.locator('[data-testid="risk-assessment"]')).toBeVisible();
    
    // Check that navigation is accessible
    await expect(page.locator('[data-testid="back-to-portfolio"]')).toBeVisible();
    await expect(page.locator('[data-testid="edit-merchant"]')).toBeVisible();
  });

  test('should handle loading states gracefully', async ({ page }) => {
    // Navigate to merchant detail page
    await page.goto('/merchant-detail.html?id=test-merchant-1');
    
    // Check that loading indicator is shown initially
    await expect(page.locator('[data-testid="loading-indicator"]')).toBeVisible();
    
    // Wait for loading to complete
    await page.waitForSelector('[data-testid="merchant-info"]', { timeout: 10000 });
    
    // Check that loading indicator is hidden
    await expect(page.locator('[data-testid="loading-indicator"]')).not.toBeVisible();
  });

  test('should handle merchant not found error', async ({ page }) => {
    // Navigate to non-existent merchant
    await page.goto('/merchant-detail.html?id=non-existent-merchant');
    
    // Wait for error state
    await page.waitForSelector('[data-testid="error-state"]', { timeout: 10000 });
    
    // Check error message
    await expect(page.locator('[data-testid="error-message"]')).toContainText('Merchant not found');
    
    // Check back to portfolio button
    await expect(page.locator('[data-testid="back-to-portfolio"]')).toBeVisible();
  });
});
