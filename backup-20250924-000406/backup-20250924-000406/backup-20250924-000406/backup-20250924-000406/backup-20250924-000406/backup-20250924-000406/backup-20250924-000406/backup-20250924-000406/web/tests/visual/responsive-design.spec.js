// web/tests/visual/responsive-design.spec.js
const { test, expect } = require('@playwright/test');
const { navigateToDashboard, setViewportSize, takeScreenshot, waitForCharts, setRiskState, waitForElementStable } = require('../utils/test-helpers');
const testData = require('../fixtures/test-data.json');

test.describe('Responsive Design Visual Regression Tests', () => {
  
  // Mobile viewport tests (375x667 - iPhone)
  test.describe('Mobile Viewport Tests (375x667)', () => {
    test('risk dashboard mobile layout', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-mobile-risk-dashboard.png');
    });

    test('enhanced risk indicators mobile layout', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await expect(page).toHaveScreenshot('responsive-mobile-enhanced-indicators.png');
    });

    test('main dashboard mobile layout', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'dashboard');
      await expect(page).toHaveScreenshot('responsive-mobile-main-dashboard.png');
    });

    test('index page mobile layout', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'index');
      await expect(page).toHaveScreenshot('responsive-mobile-index.png');
    });

    test('mobile navigation behavior', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test mobile navigation
      const navigation = page.locator('nav');
      await expect(navigation).toHaveScreenshot('responsive-mobile-navigation.png');
    });

    test('mobile form layout', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test mobile form layout
      const form = page.locator('#riskAssessmentForm');
      await expect(form).toHaveScreenshot('responsive-mobile-form.png');
    });
  });

  // Tablet viewport tests (768x1024 - iPad)
  test.describe('Tablet Viewport Tests (768x1024)', () => {
    test('risk dashboard tablet layout', async ({ page }) => {
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-tablet-risk-dashboard.png');
    });

    test('enhanced risk indicators tablet layout', async ({ page }) => {
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await expect(page).toHaveScreenshot('responsive-tablet-enhanced-indicators.png');
    });

    test('main dashboard tablet layout', async ({ page }) => {
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'dashboard');
      await expect(page).toHaveScreenshot('responsive-tablet-main-dashboard.png');
    });

    test('index page tablet layout', async ({ page }) => {
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'index');
      await expect(page).toHaveScreenshot('responsive-tablet-index.png');
    });

    test('tablet navigation behavior', async ({ page }) => {
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test tablet navigation
      const navigation = page.locator('nav');
      await expect(navigation).toHaveScreenshot('responsive-tablet-navigation.png');
    });

    test('tablet form layout', async ({ page }) => {
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test tablet form layout
      const form = page.locator('#riskAssessmentForm');
      await expect(form).toHaveScreenshot('responsive-tablet-form.png');
    });
  });

  // Desktop viewport tests (1920x1080)
  test.describe('Desktop Viewport Tests (1920x1080)', () => {
    test('risk dashboard desktop layout', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-desktop-risk-dashboard.png');
    });

    test('enhanced risk indicators desktop layout', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await expect(page).toHaveScreenshot('responsive-desktop-enhanced-indicators.png');
    });

    test('main dashboard desktop layout', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'dashboard');
      await expect(page).toHaveScreenshot('responsive-desktop-main-dashboard.png');
    });

    test('index page desktop layout', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'index');
      await expect(page).toHaveScreenshot('responsive-desktop-index.png');
    });

    test('desktop navigation behavior', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test desktop navigation
      const navigation = page.locator('nav');
      await expect(navigation).toHaveScreenshot('responsive-desktop-navigation.png');
    });

    test('desktop form layout', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test desktop form layout
      const form = page.locator('#riskAssessmentForm');
      await expect(form).toHaveScreenshot('responsive-desktop-form.png');
    });
  });

  // Large screen tests (2560x1440)
  test.describe('Large Screen Tests (2560x1440)', () => {
    test('risk dashboard large screen layout', async ({ page }) => {
      await setViewportSize(page, 'large');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-large-risk-dashboard.png');
    });

    test('enhanced risk indicators large screen layout', async ({ page }) => {
      await setViewportSize(page, 'large');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await expect(page).toHaveScreenshot('responsive-large-enhanced-indicators.png');
    });

    test('main dashboard large screen layout', async ({ page }) => {
      await setViewportSize(page, 'large');
      await navigateToDashboard(page, 'dashboard');
      await expect(page).toHaveScreenshot('responsive-large-main-dashboard.png');
    });

    test('index page large screen layout', async ({ page }) => {
      await setViewportSize(page, 'large');
      await navigateToDashboard(page, 'index');
      await expect(page).toHaveScreenshot('responsive-large-index.png');
    });

    test('large screen navigation behavior', async ({ page }) => {
      await setViewportSize(page, 'large');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test large screen navigation
      const navigation = page.locator('nav');
      await expect(navigation).toHaveScreenshot('responsive-large-navigation.png');
    });

    test('large screen form layout', async ({ page }) => {
      await setViewportSize(page, 'large');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test large screen form layout
      const form = page.locator('#riskAssessmentForm');
      await expect(form).toHaveScreenshot('responsive-large-form.png');
    });
  });

  // Cross-viewport responsive behavior tests
  test.describe('Cross-Viewport Responsive Behavior Tests', () => {
    test('responsive breakpoint transitions - mobile to tablet', async ({ page }) => {
      // Start with mobile
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-transition-mobile.png');
      
      // Transition to tablet
      await setViewportSize(page, 'tablet');
      await expect(page).toHaveScreenshot('responsive-transition-tablet.png');
    });

    test('responsive breakpoint transitions - tablet to desktop', async ({ page }) => {
      // Start with tablet
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-transition-tablet-start.png');
      
      // Transition to desktop
      await setViewportSize(page, 'desktop');
      await expect(page).toHaveScreenshot('responsive-transition-desktop.png');
    });

    test('responsive breakpoint transitions - desktop to large', async ({ page }) => {
      // Start with desktop
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-transition-desktop-start.png');
      
      // Transition to large screen
      await setViewportSize(page, 'large');
      await expect(page).toHaveScreenshot('responsive-transition-large.png');
    });

    test('responsive breakpoint transitions - large to mobile', async ({ page }) => {
      // Start with large screen
      await setViewportSize(page, 'large');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-transition-large-start.png');
      
      // Transition to mobile
      await setViewportSize(page, 'mobile');
      await expect(page).toHaveScreenshot('responsive-transition-mobile-end.png');
    });
  });

  // Component-specific responsive tests
  test.describe('Component-Specific Responsive Tests', () => {
    test('main content responsive behavior', async ({ page }) => {
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test main content across all viewports
      await setViewportSize(page, 'mobile');
      const mainContentMobile = page.locator('#main-content');
      await expect(mainContentMobile).toHaveScreenshot('responsive-main-content-mobile.png');
      
      await setViewportSize(page, 'tablet');
      const mainContentTablet = page.locator('#main-content');
      await expect(mainContentTablet).toHaveScreenshot('responsive-main-content-tablet.png');
      
      await setViewportSize(page, 'desktop');
      const mainContentDesktop = page.locator('#main-content');
      await expect(mainContentDesktop).toHaveScreenshot('responsive-main-content-desktop.png');
      
      await setViewportSize(page, 'large');
      const mainContentLarge = page.locator('#main-content');
      await expect(mainContentLarge).toHaveScreenshot('responsive-main-content-large.png');
    });

    test('page title responsive behavior', async ({ page }) => {
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test page title across all viewports
      await setViewportSize(page, 'mobile');
      const pageTitleMobile = page.locator('#page-title');
      await expect(pageTitleMobile).toHaveScreenshot('responsive-page-title-mobile.png');
      
      await setViewportSize(page, 'tablet');
      const pageTitleTablet = page.locator('#page-title');
      await expect(pageTitleTablet).toHaveScreenshot('responsive-page-title-tablet.png');
      
      await setViewportSize(page, 'desktop');
      const pageTitleDesktop = page.locator('#page-title');
      await expect(pageTitleDesktop).toHaveScreenshot('responsive-page-title-desktop.png');
      
      await setViewportSize(page, 'large');
      const pageTitleLarge = page.locator('#page-title');
      await expect(pageTitleLarge).toHaveScreenshot('responsive-page-title-large.png');
    });

    test('form responsive behavior', async ({ page }) => {
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test form across all viewports
      await setViewportSize(page, 'mobile');
      const formMobile = page.locator('#riskAssessmentForm');
      await expect(formMobile).toHaveScreenshot('responsive-form-mobile.png');
      
      await setViewportSize(page, 'tablet');
      const formTablet = page.locator('#riskAssessmentForm');
      await expect(formTablet).toHaveScreenshot('responsive-form-tablet.png');
      
      await setViewportSize(page, 'desktop');
      const formDesktop = page.locator('#riskAssessmentForm');
      await expect(formDesktop).toHaveScreenshot('responsive-form-desktop.png');
      
      await setViewportSize(page, 'large');
      const formLarge = page.locator('#riskAssessmentForm');
      await expect(formLarge).toHaveScreenshot('responsive-form-large.png');
    });
  });

  // Edge case responsive tests
  test.describe('Edge Case Responsive Tests', () => {
    test('very small mobile viewport (320x568)', async ({ page }) => {
      await page.setViewportSize({ width: 320, height: 568 });
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-edge-small-mobile.png');
    });

    test('large tablet viewport (1024x768)', async ({ page }) => {
      await page.setViewportSize({ width: 1024, height: 768 });
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-edge-large-tablet.png');
    });

    test('ultra-wide desktop (3440x1440)', async ({ page }) => {
      await page.setViewportSize({ width: 3440, height: 1440 });
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-edge-ultra-wide.png');
    });

    test('portrait tablet (768x1024)', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 });
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-edge-portrait-tablet.png');
    });

    test('landscape mobile (667x375)', async ({ page }) => {
      await page.setViewportSize({ width: 667, height: 375 });
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-edge-landscape-mobile.png');
    });
  });

  // Responsive navigation tests
  test.describe('Responsive Navigation Tests', () => {
    test('mobile navigation menu behavior', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test mobile navigation
      const navigation = page.locator('nav');
      await expect(navigation).toHaveScreenshot('responsive-nav-mobile.png');
      
      // Test if mobile menu exists and is accessible
      const mobileMenuButton = page.locator('[aria-label*="menu"], [aria-label*="Menu"], .mobile-menu-button, .hamburger');
      if (await mobileMenuButton.count() > 0) {
        await expect(mobileMenuButton).toHaveScreenshot('responsive-nav-mobile-button.png');
      }
    });

    test('tablet navigation behavior', async ({ page }) => {
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test tablet navigation
      const navigation = page.locator('nav');
      await expect(navigation).toHaveScreenshot('responsive-nav-tablet.png');
    });

    test('desktop navigation behavior', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test desktop navigation
      const navigation = page.locator('nav');
      await expect(navigation).toHaveScreenshot('responsive-nav-desktop.png');
    });

    test('large screen navigation behavior', async ({ page }) => {
      await setViewportSize(page, 'large');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test large screen navigation
      const navigation = page.locator('nav');
      await expect(navigation).toHaveScreenshot('responsive-nav-large.png');
    });
  });
});
