/**
 * Test utilities and helper functions for KYB Platform visual regression tests
 */

/**
 * Wait for page to be fully loaded and stable
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {number} timeout - Timeout in milliseconds (default: 5000)
 */
async function waitForPageStable(page, timeout = 5000) {
  // Wait for network to be idle
  await page.waitForLoadState('networkidle', { timeout });
  
  // Wait for all images to load
  await page.waitForFunction(() => {
    const images = Array.from(document.images);
    return images.every(img => img.complete);
  }, { timeout });
  
  // Wait for any animations to complete
  await page.waitForTimeout(1000);
}

/**
 * Set viewport size for responsive testing
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} device - Device name (mobile, tablet, desktop, large)
 */
async function setViewportSize(page, device) {
  const viewports = {
    mobile: { width: 375, height: 667 },      // iPhone
    tablet: { width: 768, height: 1024 },     // iPad
    desktop: { width: 1920, height: 1080 },   // Desktop
    large: { width: 2560, height: 1440 }      // Large screen
  };
  
  const viewport = viewports[device] || viewports.desktop;
  await page.setViewportSize(viewport);
  console.log(`📱 Set viewport to ${device}: ${viewport.width}x${viewport.height}`);
}

/**
 * Navigate to a dashboard page with optional query parameters
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} pageName - Page name (risk-dashboard, enhanced-risk-indicators, etc.)
 * @param {Object} params - Query parameters to add to URL
 */
async function navigateToDashboard(page, pageName, params = {}) {
  const baseUrl = process.env.BASE_URL || 'https://shimmering-comfort-production.up.railway.app';
  let url = `${baseUrl}/${pageName}.html`;
  
  // Add query parameters if provided
  if (Object.keys(params).length > 0) {
    const searchParams = new URLSearchParams(params);
    url += `?${searchParams.toString()}`;
  }
  
  await page.goto(url);
  await waitForPageStable(page);
  console.log(`🌐 Navigated to: ${url}`);
}

/**
 * Take a screenshot with consistent naming
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} name - Screenshot name
 * @param {Object} options - Screenshot options
 */
async function takeScreenshot(page, name, options = {}) {
  const defaultOptions = {
    fullPage: true,
    animations: 'disabled',
    ...options
  };
  
  await page.screenshot({ 
    path: `test-results/artifacts/${name}.png`,
    ...defaultOptions 
  });
  console.log(`📸 Screenshot saved: ${name}.png`);
}

/**
 * Wait for Chart.js charts to be rendered
 * @param {import('@playwright/test').Page} page - Playwright page object
 */
async function waitForCharts(page) {
  // Wait for Chart.js to be loaded
  await page.waitForFunction(() => {
    return typeof window.Chart !== 'undefined';
  }, { timeout: 10000 });
  
  // Wait for all canvas elements to be rendered
  await page.waitForFunction(() => {
    const canvases = document.querySelectorAll('canvas');
    return Array.from(canvases).every(canvas => {
      const ctx = canvas.getContext('2d');
      return ctx && canvas.width > 0 && canvas.height > 0;
    });
  }, { timeout: 10000 });
  
  console.log('📊 Charts rendered successfully');
}

/**
 * Simulate different risk states for testing
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} riskLevel - Risk level (low, medium, high, critical)
 */
async function setRiskState(page, riskLevel) {
  // Inject test data into the page
  await page.evaluate((level) => {
    // Set global risk state for testing
    window.testRiskState = level;
    
    // Update any existing risk indicators
    const riskElements = document.querySelectorAll('[data-risk-level]');
    riskElements.forEach(element => {
      element.setAttribute('data-risk-level', level);
      element.className = element.className.replace(/risk-\w+/g, `risk-${level}`);
    });
    
    // Update risk scores if they exist
    const scoreElements = document.querySelectorAll('[data-risk-score]');
    const scores = {
      low: 25,
      medium: 50,
      high: 75,
      critical: 95
    };
    
    scoreElements.forEach(element => {
      element.textContent = scores[level] || 50;
      element.setAttribute('data-risk-score', scores[level] || 50);
    });
  }, riskLevel);
  
  console.log(`🎯 Set risk state to: ${riskLevel}`);
}

/**
 * Check if element is visible and stable
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} selector - CSS selector
 * @param {number} timeout - Timeout in milliseconds
 */
async function waitForElementStable(page, selector, timeout = 5000) {
  await page.waitForSelector(selector, { state: 'visible', timeout });
  
  // Wait for element to be stable (no animations)
  await page.waitForFunction((sel) => {
    const element = document.querySelector(sel);
    if (!element) return false;
    
    const rect = element.getBoundingClientRect();
    return rect.width > 0 && rect.height > 0;
  }, selector, { timeout });
  
  console.log(`✅ Element stable: ${selector}`);
}

module.exports = {
  waitForPageStable,
  setViewportSize,
  navigateToDashboard,
  takeScreenshot,
  waitForCharts,
  setRiskState,
  waitForElementStable
};

// Export state management helpers
const stateHelpers = require('./state-helpers');
const interactiveHelpers = require('./interactive-helpers');
module.exports = {
  ...module.exports,
  ...stateHelpers,
  ...interactiveHelpers
};
