/**
 * Setup verification test for KYB Platform visual regression testing
 * This test verifies that the Playwright setup is working correctly
 */

const { test, expect } = require('@playwright/test');
const { navigateToDashboard, waitForPageStable, setViewportSize } = require('../utils/test-helpers');

test.describe('Setup Verification', () => {
  
  test('should load risk dashboard page', async ({ page }) => {
    await navigateToDashboard(page, 'risk-dashboard');
    
    // Verify page title
    await expect(page).toHaveTitle(/KYB Platform.*Enhanced Business Intelligence/);
    
    // Verify main content is loaded
    await expect(page.locator('body')).toBeVisible();
    
    // Verify Tailwind CSS is loaded
    const tailwindLoaded = await page.evaluate(() => {
      return document.querySelector('link[href*="tailwindcss"]') !== null;
    });
    expect(tailwindLoaded).toBe(true);
    
    // Verify Chart.js is loaded
    const chartJsLoaded = await page.evaluate(() => {
      return typeof window.Chart !== 'undefined';
    });
    expect(chartJsLoaded).toBe(true);
  });

  test('should load enhanced risk indicators page', async ({ page }) => {
    await navigateToDashboard(page, 'enhanced-risk-indicators');
    
    // Verify page title
    await expect(page).toHaveTitle(/Enhanced Risk Level Indicators/);
    
    // Verify main content is loaded
    await expect(page.locator('body')).toBeVisible();
    
    // Verify risk indicators are present
    const riskIndicators = page.locator('.risk-indicator, .risk-badge');
    await expect(riskIndicators.first()).toBeVisible();
  });

  test('should handle different viewport sizes', async ({ page }) => {
    await navigateToDashboard(page, 'risk-dashboard');
    
    // Test mobile viewport
    await setViewportSize(page, 'mobile');
    await waitForPageStable(page);
    await expect(page.locator('body')).toBeVisible();
    
    // Test tablet viewport
    await setViewportSize(page, 'tablet');
    await waitForPageStable(page);
    await expect(page.locator('body')).toBeVisible();
    
    // Test desktop viewport
    await setViewportSize(page, 'desktop');
    await waitForPageStable(page);
    await expect(page.locator('body')).toBeVisible();
  });

  test('should handle page navigation with query parameters', async ({ page }) => {
    await navigateToDashboard(page, 'risk-dashboard', { 
      risk: 'high',
      test: 'true'
    });
    
    // Verify URL contains query parameters
    const url = page.url();
    expect(url).toContain('risk=high');
    expect(url).toContain('test=true');
    
    // Verify page still loads correctly
    await expect(page.locator('body')).toBeVisible();
  });

  test('should capture screenshots', async ({ page }) => {
    await navigateToDashboard(page, 'risk-dashboard');
    
    // Take a full page screenshot
    await page.screenshot({ 
      path: 'test-results/artifacts/setup-verification-full.png',
      fullPage: true 
    });
    
    // Take a viewport screenshot
    await page.screenshot({ 
      path: 'test-results/artifacts/setup-verification-viewport.png',
      fullPage: false 
    });
    
    // Verify screenshots were created
    const fs = require('fs');
    expect(fs.existsSync('test-results/artifacts/setup-verification-full.png')).toBe(true);
    expect(fs.existsSync('test-results/artifacts/setup-verification-viewport.png')).toBe(true);
  });

});
