import { expect, test } from '@playwright/test';

test.describe('Navigation Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app (adjust URL as needed)
    await page.goto('http://localhost:3000');
    // Wait for page to load
    await page.waitForLoadState('networkidle');
  });

  // Helper to open mobile menu if needed
  async function openMobileMenuIfNeeded(page: any) {
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
    // Use getByRole for better reliability
    await page.getByRole('link', { name: /dashboard hub/i }).click();
    await expect(page).toHaveURL(/.*dashboard-hub/);
    await expect(page.locator('h1')).toContainText(/dashboard|hub/i);
  });

  test('should navigate to merchant portfolio', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    await page.getByRole('link', { name: /merchant portfolio/i }).click();
    await expect(page).toHaveURL(/.*merchant-portfolio/);
    await expect(page.locator('h1')).toContainText(/merchant|portfolio/i);
  });

  test('should navigate to add merchant page', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    await page.getByRole('link', { name: /add merchant/i }).click();
    await expect(page).toHaveURL(/.*add-merchant/);
    await expect(page.locator('h1, h2')).toContainText(/add|merchant/i);
  });

  test('should navigate to risk dashboard', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    await page.getByRole('link', { name: /risk assessment/i }).click();
    await expect(page).toHaveURL(/.*risk-dashboard/);
    await expect(page.locator('h1')).toContainText(/risk/i);
  });

  test('should navigate to compliance page', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    await page.getByRole('link', { name: /compliance status/i }).click();
    await expect(page).toHaveURL(/.*compliance/);
    await expect(page.locator('h1')).toContainText(/compliance/i);
  });

  test('should navigate to admin page', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    await page.getByRole('link', { name: /admin dashboard/i }).click();
    await expect(page).toHaveURL(/.*admin/);
    await expect(page.locator('h1')).toContainText(/admin/i);
  });

  test('should navigate using breadcrumbs', async ({ page }) => {
    // Navigate to a nested page
    await page.goto('http://localhost:3000/merchant-portfolio');
    
    // Click breadcrumb to go back
    const breadcrumb = page.locator('text=Home').first();
    if (await breadcrumb.isVisible()) {
      await breadcrumb.click();
      await expect(page).toHaveURL(/\/$/);
    }
  });
});

