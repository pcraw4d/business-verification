import { test, expect } from '@playwright/test';

/**
 * Visual regression tests
 * 
 * These tests capture screenshots of pages and compare them against baseline images.
 * 
 * To update baselines:
 * 1. Run: npx playwright test --update-snapshots
 * 2. Review changes in tests/visual/snapshots/
 * 3. Commit updated snapshots
 */

test.describe('Visual Regression Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Set consistent viewport for all tests
    await page.setViewportSize({ width: 1280, height: 720 });
  });

  test('home page visual regression', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Wait for any animations to complete
    await page.waitForTimeout(1000);
    
    await expect(page).toHaveScreenshot('home-page.png', {
      fullPage: true,
      maxDiffPixels: 100, // Allow small differences
    });
  });

  test('dashboard hub visual regression', async ({ page }) => {
    await page.goto('/dashboard-hub');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    await expect(page).toHaveScreenshot('dashboard-hub.png', {
      fullPage: true,
      maxDiffPixels: 100,
    });
  });

  test('merchant portfolio visual regression', async ({ page }) => {
    await page.goto('/merchant-portfolio');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    await expect(page).toHaveScreenshot('merchant-portfolio.png', {
      fullPage: true,
      maxDiffPixels: 100,
    });
  });

  test('risk dashboard visual regression', async ({ page }) => {
    await page.goto('/risk-dashboard');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    await expect(page).toHaveScreenshot('risk-dashboard.png', {
      fullPage: true,
      maxDiffPixels: 100,
    });
  });

  test('compliance page visual regression', async ({ page }) => {
    await page.goto('/compliance');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    await expect(page).toHaveScreenshot('compliance.png', {
      fullPage: true,
      maxDiffPixels: 100,
    });
  });

  test('add merchant page visual regression', async ({ page }) => {
    await page.goto('/add-merchant');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    await expect(page).toHaveScreenshot('add-merchant.png', {
      fullPage: true,
      maxDiffPixels: 100,
    });
  });

  test('admin page visual regression', async ({ page }) => {
    await page.goto('/admin');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    await expect(page).toHaveScreenshot('admin.png', {
      fullPage: true,
      maxDiffPixels: 100,
    });
  });

  test('mobile viewport visual regression', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE size
    
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    await expect(page).toHaveScreenshot('home-mobile.png', {
      fullPage: true,
      maxDiffPixels: 100,
    });
  });

  test('tablet viewport visual regression', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 }); // iPad size
    
    await page.goto('/merchant-portfolio');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    await expect(page).toHaveScreenshot('merchant-portfolio-tablet.png', {
      fullPage: true,
      maxDiffPixels: 100,
    });
  });
});

