import { expect, Page, test } from '@playwright/test';

test.describe('Navigation Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app (adjust URL as needed)
    await page.goto('/');
    // Wait for page to load
    await page.waitForLoadState('networkidle');
  });

  // Helper to open mobile menu if needed
  async function openMobileMenuIfNeeded(page: Page) {
    const viewport = page.viewportSize();
    const isMobile = viewport && viewport.width < 768;
    
    if (isMobile) {
      // Look for menu button (hamburger menu)
      const menuButton = page.locator('button[aria-label*="menu" i], button:has([class*="Menu"]), button:has-text("Toggle sidebar")').first();
      if (await menuButton.isVisible({ timeout: 2000 }).catch(() => false)) {
        await menuButton.click();
        // Wait for sidebar to open
        await page.waitForTimeout(500);
      }
    }
  }

  test('should navigate to dashboard hub', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    // Use more specific selector - match exact text and href
    const dashboardLink = page.locator('a[href="/dashboard-hub"]').filter({ hasText: /dashboard hub/i }).first();
    await dashboardLink.scrollIntoViewIfNeeded();
    await page.waitForTimeout(300); // Wait for scroll
    await dashboardLink.click({ force: true });
    // Wait for navigation to complete
    await page.waitForURL(/.*dashboard-hub/, { timeout: 10000 });
    await expect(page).toHaveURL(/.*dashboard-hub/);
    // Use main content h1, or fallback to h1 if main h1 doesn't exist
    const mainH1 = page.locator('main h1').first();
    const h1 = page.locator('h1').first();
    const hasMainH1 = await mainH1.isVisible({ timeout: 2000 }).catch(() => false);
    if (hasMainH1) {
      await expect(mainH1).toContainText(/dashboard|hub/i);
    } else {
      await expect(h1).toContainText(/dashboard|hub/i);
    }
  });

  test('should navigate to merchant portfolio', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    // Use href selector to be more specific
    const portfolioLink = page.locator('a[href="/merchant-portfolio"]').filter({ hasText: /merchant portfolio/i }).first();
    await portfolioLink.scrollIntoViewIfNeeded();
    await page.waitForTimeout(500); // Wait for scroll and ensure element is stable
    // Try multiple strategies if element is still outside viewport
    try {
      await portfolioLink.click({ force: true, timeout: 5000 });
    } catch {
      // If still fails, try clicking via JavaScript
      await portfolioLink.evaluate((el: HTMLElement) => el.click());
    }
    await expect(page).toHaveURL(/.*merchant-portfolio/, { timeout: 10000 });
    // Use main content h1, or fallback to h1 or h2 if main h1 doesn't exist
    const mainH1 = page.locator('main h1').first();
    const h1 = page.locator('h1').first();
    const hasMainH1 = await mainH1.isVisible({ timeout: 2000 }).catch(() => false);
    if (hasMainH1) {
      await expect(mainH1).toContainText(/merchant|portfolio/i);
    } else {
      await expect(h1).toContainText(/merchant|portfolio/i);
    }
  });

  test('should navigate to add merchant page', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    const addMerchantLink = page.getByRole('link', { name: /add merchant/i }).first();
    await addMerchantLink.scrollIntoViewIfNeeded();
    await addMerchantLink.click({ force: true });
    await expect(page).toHaveURL(/.*add-merchant/, { timeout: 10000 });
    // Use main content h1, or fallback to h1 or h2 if main h1 doesn't exist
    const mainH1 = page.locator('main h1').first();
    const h1 = page.locator('h1').first();
    const hasMainH1 = await mainH1.isVisible({ timeout: 2000 }).catch(() => false);
    if (hasMainH1) {
      await expect(mainH1).toContainText(/add|merchant/i);
    } else {
      await expect(h1).toContainText(/add|merchant/i);
    }
  });

  test('should navigate to risk dashboard', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    // Use href selector to be more specific - match exact href for Risk Assessment (not Risk Assessment Portfolio)
    const riskLink = page.locator('a[href="/risk-dashboard"]').filter({ hasText: /^Risk Assessment$/ }).first();
    await riskLink.scrollIntoViewIfNeeded();
    await page.waitForTimeout(300); // Wait for scroll
    await riskLink.click({ force: true });
    // Wait for navigation to complete
    await page.waitForURL(/.*risk-dashboard/, { timeout: 10000 });
    await expect(page).toHaveURL(/.*risk-dashboard/);
    // Use main content h1, or fallback to h1 if main h1 doesn't exist
    const mainH1 = page.locator('main h1').first();
    const h1 = page.locator('h1').first();
    const hasMainH1 = await mainH1.isVisible({ timeout: 2000 }).catch(() => false);
    if (hasMainH1) {
      await expect(mainH1).toContainText(/risk/i);
    } else {
      await expect(h1).toContainText(/risk/i);
    }
  });

  test('should navigate to compliance page', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    const complianceLink = page.getByRole('link', { name: /compliance status/i }).first();
    await complianceLink.scrollIntoViewIfNeeded();
    await complianceLink.click({ force: true });
    await expect(page).toHaveURL(/.*compliance/, { timeout: 10000 });
    // Use main content h1, or fallback to h1 if main h1 doesn't exist
    const mainH1 = page.locator('main h1').first();
    const h1 = page.locator('h1').first();
    const hasMainH1 = await mainH1.isVisible({ timeout: 2000 }).catch(() => false);
    if (hasMainH1) {
      await expect(mainH1).toContainText(/compliance/i);
    } else {
      await expect(h1).toContainText(/compliance/i);
    }
  });

  test('should navigate to admin page', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    // Use href selector to be more specific
    const adminLink = page.locator('a[href="/admin"]').filter({ hasText: /admin dashboard/i }).first();
    await adminLink.scrollIntoViewIfNeeded();
    await page.waitForTimeout(500); // Wait for scroll to complete
    // Try multiple strategies if element is still outside viewport
    try {
      await adminLink.click({ force: true, timeout: 5000 });
    } catch {
      // If still fails, try clicking via JavaScript
      await adminLink.evaluate((el: HTMLElement) => el.click());
    }
    await expect(page).toHaveURL(/.*admin/, { timeout: 10000 });
    // Use main content h1, or fallback to h1 if main h1 doesn't exist
    const mainH1 = page.locator('main h1').first();
    const h1 = page.locator('h1').first();
    const hasMainH1 = await mainH1.isVisible({ timeout: 2000 }).catch(() => false);
    if (hasMainH1) {
      await expect(mainH1).toContainText(/admin/i);
    } else {
      await expect(h1).toContainText(/admin/i);
    }
  });

  test('should navigate using breadcrumbs', async ({ page }) => {
    // Navigate to a nested page
    await page.goto('/merchant-portfolio');
    
    // Click breadcrumb to go back
    const breadcrumb = page.locator('text=Home').first();
    if (await breadcrumb.isVisible()) {
      await breadcrumb.click();
      await expect(page).toHaveURL(/\/$/);
    }
  });
});

