/**
 * Baseline screenshot generation for KYB Platform visual regression testing
 * This test generates baseline screenshots for all dashboard pages and risk states
 */

const { test, expect } = require('@playwright/test');
const { 
  navigateToDashboard, 
  waitForPageStable, 
  setViewportSize, 
  setRiskState,
  waitForCharts 
} = require('../utils/test-helpers');

test.describe('Baseline Screenshot Generation', () => {
  
  // Test different viewport sizes
  const viewports = ['mobile', 'tablet', 'desktop', 'large'];
  
  // Test different risk states
  const riskStates = ['low', 'medium', 'high', 'critical'];
  
  // Test different dashboard pages
  const dashboardPages = [
    'risk-dashboard',
    'enhanced-risk-indicators', 
    'dashboard',
    'index'
  ];

  // Generate baseline screenshots for risk dashboard
  test('generate risk dashboard baselines', async ({ page }) => {
    for (const viewport of viewports) {
      await setViewportSize(page, viewport);
      
      // Default state
      await navigateToDashboard(page, 'risk-dashboard');
      await waitForCharts(page);
      await page.screenshot({ 
        path: `test-results/artifacts/baseline-risk-dashboard-${viewport}.png`,
        fullPage: true,
        animations: 'disabled'
      });
      
      // Test different risk states
      for (const riskState of riskStates) {
        await setRiskState(page, riskState);
        await waitForPageStable(page);
        await page.screenshot({ 
          path: `test-results/artifacts/baseline-risk-dashboard-${riskState}-${viewport}.png`,
          fullPage: true,
          animations: 'disabled'
        });
      }
    }
  });

  // Generate baseline screenshots for enhanced risk indicators
  test('generate enhanced risk indicators baselines', async ({ page }) => {
    // Increase timeout for this test
    test.setTimeout(120000); // 2 minutes
    
    for (const viewport of viewports) {
      await setViewportSize(page, viewport);
      
      // Default state
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      await waitForPageStable(page);
      
      // Wait for animations to complete
      await page.waitForTimeout(2000);
      
      await page.screenshot({ 
        path: `test-results/artifacts/baseline-enhanced-indicators-${viewport}.png`,
        fullPage: true,
        animations: 'disabled',
        timeout: 30000
      });
      
      // Test different risk states
      for (const riskState of riskStates) {
        await setRiskState(page, riskState);
        await waitForPageStable(page);
        
        // Wait for state changes to render
        await page.waitForTimeout(1000);
        
        await page.screenshot({ 
          path: `test-results/artifacts/baseline-enhanced-indicators-${riskState}-${viewport}.png`,
          fullPage: true,
          animations: 'disabled',
          timeout: 30000
        });
      }
    }
  });

  // Generate baseline screenshots for main dashboard
  test('generate main dashboard baselines', async ({ page }) => {
    for (const viewport of viewports) {
      await setViewportSize(page, viewport);
      
      await navigateToDashboard(page, 'dashboard');
      await waitForPageStable(page);
      await page.screenshot({ 
        path: `test-results/artifacts/baseline-dashboard-${viewport}.png`,
        fullPage: true,
        animations: 'disabled'
      });
    }
  });

  // Generate baseline screenshots for index page
  test('generate index page baselines', async ({ page }) => {
    for (const viewport of viewports) {
      await setViewportSize(page, viewport);
      
      await navigateToDashboard(page, 'index');
      await waitForPageStable(page);
      await page.screenshot({ 
        path: `test-results/artifacts/baseline-index-${viewport}.png`,
        fullPage: true,
        animations: 'disabled'
      });
    }
  });

  // Generate component-specific screenshots
  test('generate component baselines', async ({ page }) => {
    await setViewportSize(page, 'desktop');
    await navigateToDashboard(page, 'risk-dashboard');
    await waitForCharts(page);
    
    // Risk gauge component
    const riskGauge = page.locator('.risk-gauge, [data-testid="risk-gauge"]').first();
    if (await riskGauge.isVisible()) {
      await riskGauge.screenshot({ 
        path: 'test-results/artifacts/baseline-risk-gauge.png'
      });
    }
    
    // Risk cards
    const riskCards = page.locator('.risk-card, [data-testid="risk-card"]').first();
    if (await riskCards.isVisible()) {
      await riskCards.screenshot({ 
        path: 'test-results/artifacts/baseline-risk-cards.png'
      });
    }
    
    // Risk level indicators
    const riskIndicators = page.locator('.risk-indicator, .risk-badge').first();
    if (await riskIndicators.isVisible()) {
      await riskIndicators.screenshot({ 
        path: 'test-results/artifacts/baseline-risk-indicators.png'
      });
    }
    
    // Charts
    const charts = page.locator('canvas').first();
    if (await charts.isVisible()) {
      await charts.screenshot({ 
        path: 'test-results/artifacts/baseline-charts.png'
      });
    }
  });

  // Generate interactive state screenshots
  test('generate interactive state baselines', async ({ page }) => {
    await setViewportSize(page, 'desktop');
    await navigateToDashboard(page, 'risk-dashboard');
    await waitForCharts(page);
    
    // Hover states
    const hoverElements = page.locator('.card-hover, .hover-target, .risk-card');
    if (await hoverElements.first().isVisible()) {
      await hoverElements.first().hover();
      await page.waitForTimeout(500); // Wait for hover animation
      await hoverElements.first().screenshot({ 
        path: 'test-results/artifacts/baseline-hover-state.png'
      });
    }
    
    // Focus states
    const focusableElements = page.locator('button, [tabindex], .clickable');
    if (await focusableElements.first().isVisible()) {
      await focusableElements.first().focus();
      await page.waitForTimeout(500); // Wait for focus animation
      await focusableElements.first().screenshot({ 
        path: 'test-results/artifacts/baseline-focus-state.png'
      });
    }
  });

  // Generate loading state screenshots
  test('generate loading state baselines', async ({ page }) => {
    await setViewportSize(page, 'desktop');
    
    // Simulate loading state by navigating with slow network
    await page.route('**/*', route => {
      // Add delay to simulate loading
      setTimeout(() => route.continue(), 100);
    });
    
    await navigateToDashboard(page, 'risk-dashboard');
    await page.screenshot({ 
      path: 'test-results/artifacts/baseline-loading-state.png',
      fullPage: true
    });
  });

  // Generate error state screenshots
  test('generate error state baselines', async ({ page }) => {
    await setViewportSize(page, 'desktop');
    
    // Simulate error state by blocking resources
    await page.route('**/*.js', route => route.abort());
    await page.route('**/*.css', route => route.abort());
    
    await navigateToDashboard(page, 'risk-dashboard');
    await page.screenshot({ 
      path: 'test-results/artifacts/baseline-error-state.png',
      fullPage: true
    });
  });

});
