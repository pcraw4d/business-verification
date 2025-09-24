// web/tests/visual/dashboard-layout.spec.js
const { test, expect } = require('@playwright/test');
const { navigateToDashboard, setViewportSize, takeScreenshot, waitForCharts, setRiskState, waitForElementStable } = require('../utils/test-helpers');
const testData = require('../fixtures/test-data.json');

test.describe('Dashboard Layout Visual Regression Tests', () => {
  
  // Full-page layout regression tests
  test.describe('Full Page Layout Tests', () => {
    test('risk dashboard full page layout - desktop', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('baseline-risk-dashboard-desktop.png');
    });

    test('risk dashboard full page layout - mobile', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('baseline-risk-dashboard-mobile.png');
    });

    test('risk dashboard full page layout - tablet', async ({ page }) => {
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('baseline-risk-dashboard-tablet.png');
    });

    test('enhanced risk indicators full page layout - desktop', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await expect(page).toHaveScreenshot('baseline-enhanced-indicators-desktop.png');
    });

    test('enhanced risk indicators full page layout - mobile', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await expect(page).toHaveScreenshot('baseline-enhanced-indicators-mobile.png');
    });

    test('enhanced risk indicators full page layout - tablet', async ({ page }) => {
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await expect(page).toHaveScreenshot('baseline-enhanced-indicators-tablet.png');
    });
  });

  // Component-level visual tests using actual HTML elements
  test.describe('Component Level Visual Tests', () => {
    test('main content area visual test', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Wait for main content to be visible
      await waitForElementStable(page, '#main-content');
      
      // Capture screenshot of the main content area
      const mainContent = page.locator('#main-content');
      await expect(mainContent).toHaveScreenshot('main-content-area.png');
    });

    test('navigation bar visual test', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Wait for navigation to be visible
      await waitForElementStable(page, 'nav');
      
      // Capture screenshot of the navigation bar
      const navigation = page.locator('nav');
      await expect(navigation).toHaveScreenshot('navigation-bar.png');
    });

    test('page title visual test', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Wait for page title to be visible
      await waitForElementStable(page, '#page-title');
      
      // Capture screenshot of the page title
      const pageTitle = page.locator('#page-title');
      await expect(pageTitle).toHaveScreenshot('page-title.png');
    });

    test('form container visual test', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Wait for form to be visible
      await waitForElementStable(page, '#riskAssessmentForm');
      
      // Capture screenshot of the form container
      const formContainer = page.locator('#riskAssessmentForm');
      await expect(formContainer).toHaveScreenshot('form-container.png');
    });
  });

  // Layout consistency tests
  test.describe('Layout Consistency Tests', () => {
    test('header layout consistency across pages', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      
      // Test header on risk dashboard
      await navigateToDashboard(page, 'risk-dashboard');
      const header1 = page.locator('nav');
      await expect(header1).toHaveScreenshot('header-risk-dashboard.png');
      
      // Test header on enhanced risk indicators
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      const header2 = page.locator('nav');
      await expect(header2).toHaveScreenshot('header-enhanced-indicators.png');
    });

    test('main content layout consistency', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      
      // Test main content on risk dashboard
      await navigateToDashboard(page, 'risk-dashboard');
      const mainContent1 = page.locator('#main-content');
      await expect(mainContent1).toHaveScreenshot('main-content-risk-dashboard.png');
      
      // Test main content on enhanced risk indicators
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      const mainContent2 = page.locator('#main-content');
      await expect(mainContent2).toHaveScreenshot('main-content-enhanced-indicators.png');
    });
  });

  // Cross-page layout comparison tests
  test.describe('Cross-Page Layout Comparison Tests', () => {
    test('risk dashboard vs enhanced indicators layout comparison', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      
      // Capture risk dashboard layout
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('layout-comparison-risk-dashboard.png');
      
      // Capture enhanced indicators layout
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await expect(page).toHaveScreenshot('layout-comparison-enhanced-indicators.png');
    });

    test('main dashboard vs risk dashboard layout comparison', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      
      // Capture main dashboard layout
      await navigateToDashboard(page, 'dashboard');
      await expect(page).toHaveScreenshot('layout-comparison-main-dashboard.png');
      
      // Capture risk dashboard layout
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('layout-comparison-risk-dashboard.png');
    });
  });

  // Responsive layout tests
  test.describe('Responsive Layout Tests', () => {
    test('responsive layout - mobile to desktop', async ({ page }) => {
      // Start with mobile
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-mobile-layout.png');
      
      // Switch to desktop
      await setViewportSize(page, 'desktop');
      await expect(page).toHaveScreenshot('responsive-desktop-layout.png');
    });

    test('responsive layout - tablet to large screen', async ({ page }) => {
      // Start with tablet
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'risk-dashboard');
      await expect(page).toHaveScreenshot('responsive-tablet-layout.png');
      
      // Switch to large screen
      await setViewportSize(page, 'large');
      await expect(page).toHaveScreenshot('responsive-large-layout.png');
    });
  });
});