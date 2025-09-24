// web/tests/visual/cross-browser.spec.js
const { test, expect } = require('@playwright/test');
const { navigateToDashboard, setViewportSize, waitForElementStable, setRiskState } = require('../utils/test-helpers');
const testData = require('../fixtures/test-data.json');

test.describe('Cross-Browser Visual Regression Tests', () => {
  
  // Chrome browser testing
  test.describe('Chrome Browser Tests', () => {
    test('risk dashboard - Chrome desktop layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'chromium', 'Chrome-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('chrome-risk-dashboard-desktop.png');
    });

    test('enhanced risk indicators - Chrome desktop layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'chromium', 'Chrome-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('chrome-enhanced-indicators-desktop.png');
    });

    test('risk dashboard - Chrome mobile layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'chromium', 'Chrome-specific test');
      
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('chrome-risk-dashboard-mobile.png');
    });

    test('risk level indicators - Chrome cross-browser compatibility', async ({ page, browserName }) => {
      test.skip(browserName !== 'chromium', 'Chrome-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test different risk levels
      await setRiskState(page, 'low');
      await waitForElementStable(page, '.risk-level-indicator');
      await expect(page.locator('.risk-level-indicator')).toHaveScreenshot('chrome-risk-level-low.png');
      
      await setRiskState(page, 'medium');
      await waitForElementStable(page, '.risk-level-indicator');
      await expect(page.locator('.risk-level-indicator')).toHaveScreenshot('chrome-risk-level-medium.png');
      
      await setRiskState(page, 'high');
      await waitForElementStable(page, '.risk-level-indicator');
      await expect(page.locator('.risk-level-indicator')).toHaveScreenshot('chrome-risk-level-high.png');
      
      await setRiskState(page, 'critical');
      await waitForElementStable(page, '.risk-level-indicator');
      await expect(page.locator('.risk-level-indicator')).toHaveScreenshot('chrome-risk-level-critical.png');
    });
  });

  // Firefox browser testing
  test.describe('Firefox Browser Tests', () => {
    test('risk dashboard - Firefox desktop layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'firefox', 'Firefox-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('firefox-risk-dashboard-desktop.png');
    });

    test('enhanced risk indicators - Firefox desktop layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'firefox', 'Firefox-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('firefox-enhanced-indicators-desktop.png');
    });

    test('risk dashboard - Firefox mobile layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'firefox', 'Firefox-specific test');
      
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('firefox-risk-dashboard-mobile.png');
    });

    test('CSS animations and transitions - Firefox compatibility', async ({ page, browserName }) => {
      test.skip(browserName !== 'firefox', 'Firefox-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test animation states
      const riskIndicator = page.locator('.risk-level-indicator');
      await riskIndicator.hover();
      await page.waitForTimeout(500); // Wait for animation
      await expect(riskIndicator).toHaveScreenshot('firefox-risk-indicator-hover.png');
      
      // Test transition effects
      await setRiskState(page, 'low');
      await page.waitForTimeout(300); // Wait for transition
      await expect(riskIndicator).toHaveScreenshot('firefox-risk-transition-low.png');
    });

    test('form elements - Firefox styling compatibility', async ({ page, browserName }) => {
      test.skip(browserName !== 'firefox', 'Firefox-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      const form = page.locator('#riskAssessmentForm');
      await waitForElementStable(page, '#riskAssessmentForm');
      await expect(form).toHaveScreenshot('firefox-form-styling.png');
    });
  });

  // Safari/WebKit browser testing
  test.describe('Safari Browser Tests', () => {
    test('risk dashboard - Safari desktop layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'webkit', 'Safari-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('safari-risk-dashboard-desktop.png');
    });

    test('enhanced risk indicators - Safari desktop layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'webkit', 'Safari-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('safari-enhanced-indicators-desktop.png');
    });

    test('risk dashboard - Safari mobile layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'webkit', 'Safari-specific test');
      
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('safari-risk-dashboard-mobile.png');
    });

    test('WebKit-specific CSS features - Safari compatibility', async ({ page, browserName }) => {
      test.skip(browserName !== 'webkit', 'Safari-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test WebKit-specific features like backdrop-filter
      const riskCard = page.locator('.risk-card');
      await waitForElementStable(page, '.risk-card');
      await expect(riskCard).toHaveScreenshot('safari-webkit-features.png');
    });

    test('touch interactions - Safari mobile compatibility', async ({ page, browserName }) => {
      test.skip(browserName !== 'webkit', 'Safari-specific test');
      
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test touch interactions
      const interactiveElement = page.locator('.interactive-risk-element');
      if (await interactiveElement.count() > 0) {
        await interactiveElement.tap();
        await page.waitForTimeout(300);
        await expect(interactiveElement).toHaveScreenshot('safari-touch-interaction.png');
      }
    });
  });

  // Edge browser testing
  test.describe('Edge Browser Tests', () => {
    test('risk dashboard - Edge desktop layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'msedge', 'Edge-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('edge-risk-dashboard-desktop.png');
    });

    test('enhanced risk indicators - Edge desktop layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'msedge', 'Edge-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('edge-enhanced-indicators-desktop.png');
    });

    test('risk dashboard - Edge mobile layout', async ({ page, browserName }) => {
      test.skip(browserName !== 'msedge', 'Edge-specific test');
      
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      await waitForElementStable(page, '#main-content');
      await expect(page).toHaveScreenshot('edge-risk-dashboard-mobile.png');
    });

    test('Edge-specific features and compatibility', async ({ page, browserName }) => {
      test.skip(browserName !== 'msedge', 'Edge-specific test');
      
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test Edge-specific rendering
      const riskVisualization = page.locator('.risk-visualization');
      await waitForElementStable(page, '.risk-visualization');
      await expect(riskVisualization).toHaveScreenshot('edge-visualization-rendering.png');
    });
  });

  // Cross-browser comparison tests
  test.describe('Cross-Browser Comparison Tests', () => {
    test('risk level indicators - cross-browser consistency', async ({ page, browserName }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Set to medium risk for consistent comparison
      await setRiskState(page, 'medium');
      await waitForElementStable(page, '.risk-level-indicator');
      
      // Take screenshot with browser name in filename
      const browserPrefix = browserName === 'chromium' ? 'chrome' : browserName;
      await expect(page.locator('.risk-level-indicator')).toHaveScreenshot(`${browserPrefix}-risk-level-medium.png`);
    });

    test('form styling - cross-browser consistency', async ({ page, browserName }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      const form = page.locator('#riskAssessmentForm');
      await waitForElementStable(page, '#riskAssessmentForm');
      
      const browserPrefix = browserName === 'chromium' ? 'chrome' : browserName;
      await expect(form).toHaveScreenshot(`${browserPrefix}-form-styling.png`);
    });

    test('navigation bar - cross-browser consistency', async ({ page, browserName }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      const navigation = page.locator('nav');
      await waitForElementStable(page, 'nav');
      
      const browserPrefix = browserName === 'chromium' ? 'chrome' : browserName;
      await expect(navigation).toHaveScreenshot(`${browserPrefix}-navigation-bar.png`);
    });

    test('responsive layout - cross-browser mobile consistency', async ({ page, browserName }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'risk-dashboard');
      await waitForElementStable(page, '#main-content');
      
      const browserPrefix = browserName === 'chromium' ? 'chrome' : browserName;
      await expect(page).toHaveScreenshot(`${browserPrefix}-mobile-layout.png`);
    });
  });

  // Browser-specific feature tests
  test.describe('Browser-Specific Feature Tests', () => {
    test('CSS Grid and Flexbox - cross-browser compatibility', async ({ page, browserName }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test CSS Grid layout
      const gridContainer = page.locator('.risk-grid-container');
      if (await gridContainer.count() > 0) {
        await waitForElementStable(page, '.risk-grid-container');
        const browserPrefix = browserName === 'chromium' ? 'chrome' : browserName;
        await expect(gridContainer).toHaveScreenshot(`${browserPrefix}-css-grid-layout.png`);
      }
      
      // Test Flexbox layout
      const flexContainer = page.locator('.risk-flex-container');
      if (await flexContainer.count() > 0) {
        await waitForElementStable(page, '.risk-flex-container');
        const browserPrefix = browserName === 'chromium' ? 'chrome' : browserName;
        await expect(flexContainer).toHaveScreenshot(`${browserPrefix}-flexbox-layout.png`);
      }
    });

    test('CSS animations - cross-browser performance', async ({ page, browserName }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test animation performance across browsers
      const animatedElement = page.locator('.animated-risk-element');
      if (await animatedElement.count() > 0) {
        await animatedElement.hover();
        await page.waitForTimeout(500); // Wait for animation
        
        const browserPrefix = browserName === 'chromium' ? 'chrome' : browserName;
        await expect(animatedElement).toHaveScreenshot(`${browserPrefix}-animation-state.png`);
      }
    });

    test('JavaScript interactions - cross-browser functionality', async ({ page, browserName }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test JavaScript-driven interactions
      const interactiveButton = page.locator('.interactive-button');
      if (await interactiveButton.count() > 0) {
        await interactiveButton.click();
        await page.waitForTimeout(300);
        
        const browserPrefix = browserName === 'chromium' ? 'chrome' : browserName;
        await expect(interactiveButton).toHaveScreenshot(`${browserPrefix}-interaction-state.png`);
      }
    });
  });

  // Performance and rendering tests
  test.describe('Performance and Rendering Tests', () => {
    test('font rendering - cross-browser consistency', async ({ page, browserName }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      const textElement = page.locator('#page-title');
      await waitForElementStable(page, '#page-title');
      
      const browserPrefix = browserName === 'chromium' ? 'chrome' : browserName;
      await expect(textElement).toHaveScreenshot(`${browserPrefix}-font-rendering.png`);
    });

    test('color rendering - cross-browser consistency', async ({ page, browserName }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test color rendering for risk levels
      await setRiskState(page, 'high');
      await waitForElementStable(page, '.risk-level-indicator');
      
      const browserPrefix = browserName === 'chromium' ? 'chrome' : browserName;
      await expect(page.locator('.risk-level-indicator')).toHaveScreenshot(`${browserPrefix}-color-rendering.png`);
    });

    test('image and icon rendering - cross-browser consistency', async ({ page, browserName }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      const iconElement = page.locator('.risk-icon');
      if (await iconElement.count() > 0) {
        await waitForElementStable(page, '.risk-icon');
        
        const browserPrefix = browserName === 'chromium' ? 'chrome' : browserName;
        await expect(iconElement).toHaveScreenshot(`${browserPrefix}-icon-rendering.png`);
      }
    });
  });
});
