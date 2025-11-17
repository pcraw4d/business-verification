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
    // Use main content h1, not header h1
    await expect(page.locator('main h1')).toContainText(/dashboard|hub/i);
  });

  test('should navigate to merchant portfolio', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    // Use first() to select the sidebar link, not the button on the page
    await page.getByRole('link', { name: /merchant portfolio/i }).first().click();
    await expect(page).toHaveURL(/.*merchant-portfolio/);
    // Use main content h1, not header h1
    await expect(page.locator('main h1')).toContainText(/merchant|portfolio/i);
  });

  test('should navigate to add merchant page', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    await page.getByRole('link', { name: /add merchant/i }).click();
    await expect(page).toHaveURL(/.*add-merchant/);
    // Use main content h1, exclude sr-only h2
    await expect(page.locator('main h1')).toContainText(/add|merchant/i);
  });

  test('should navigate to risk dashboard', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    // Be more specific - use exact match for "Risk Assessment" (not "Risk Assessment Portfolio")
    await page.getByRole('link', { name: 'Risk Assessment', exact: true }).click();
    await expect(page).toHaveURL(/.*risk-dashboard/);
    // Use main content h1, not header h1
    await expect(page.locator('main h1')).toContainText(/risk/i);
  });

  test('should navigate to compliance page', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    await page.getByRole('link', { name: /compliance status/i }).click();
    await expect(page).toHaveURL(/.*compliance/);
    // Use main content h1, not header h1
    await expect(page.locator('main h1')).toContainText(/compliance/i);
  });

  test('should navigate to admin page', async ({ page }) => {
    await openMobileMenuIfNeeded(page);
    // Scroll to ensure element is in viewport
    const adminLink = page.getByRole('link', { name: /admin dashboard/i });
    await adminLink.scrollIntoViewIfNeeded();
    await adminLink.click();
    await expect(page).toHaveURL(/.*admin/);
    // Use main content h1, not header h1
    await expect(page.locator('main h1')).toContainText(/admin/i);
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

