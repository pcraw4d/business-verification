// web/tests/utils/browser-compatibility.js
const { expect } = require('@playwright/test');

/**
 * Browser compatibility utilities for cross-browser testing
 * Provides helper functions for browser-specific testing scenarios
 */

/**
 * Get browser-specific filename prefix
 * @param {string} browserName - The browser name from Playwright
 * @returns {string} Browser prefix for filenames
 */
function getBrowserPrefix(browserName) {
  const browserMap = {
    'chromium': 'chrome',
    'firefox': 'firefox',
    'webkit': 'safari',
    'msedge': 'edge'
  };
  return browserMap[browserName] || browserName;
}

/**
 * Check if current browser supports specific CSS features
 * @param {Object} page - Playwright page object
 * @param {string} feature - CSS feature to check
 * @returns {Promise<boolean>} Whether the feature is supported
 */
async function supportsCSSFeature(page, feature) {
  const supportMap = {
    'css-grid': 'CSS.supports("display", "grid")',
    'flexbox': 'CSS.supports("display", "flex")',
    'backdrop-filter': 'CSS.supports("backdrop-filter", "blur(10px)")',
    'css-variables': 'CSS.supports("--custom-property", "value")',
    'css-animations': 'CSS.supports("animation", "fadeIn 1s")',
    'css-transforms': 'CSS.supports("transform", "translateX(10px)")'
  };

  const cssCheck = supportMap[feature];
  if (!cssCheck) {
    throw new Error(`Unknown CSS feature: ${feature}`);
  }

  try {
    const result = await page.evaluate((check) => {
      return eval(check);
    }, cssCheck);
    return result;
  } catch (error) {
    console.warn(`Error checking CSS feature ${feature}:`, error);
    return false;
  }
}

/**
 * Wait for browser-specific rendering to complete
 * @param {Object} page - Playwright page object
 * @param {string} browserName - Browser name
 * @returns {Promise<void>}
 */
async function waitForBrowserRendering(page, browserName) {
  // Browser-specific rendering delays
  const renderingDelays = {
    'chromium': 100,
    'firefox': 150,
    'webkit': 200,
    'msedge': 100
  };

  const delay = renderingDelays[browserName] || 100;
  await page.waitForTimeout(delay);

  // Wait for fonts to load (especially important for Safari)
  if (browserName === 'webkit') {
    await page.evaluate(() => {
      return document.fonts.ready;
    });
  }

  // Wait for any pending animations
  await page.waitForFunction(() => {
    return document.getAnimations().every(animation => animation.playState === 'finished' || animation.playState === 'idle');
  }, { timeout: 5000 }).catch(() => {
    // Ignore timeout - animations might not be present
  });
}

/**
 * Take browser-specific screenshot with proper naming
 * @param {Object} locator - Playwright locator
 * @param {string} browserName - Browser name
 * @param {string} testName - Test name for filename
 * @param {Object} options - Screenshot options
 * @returns {Promise<void>}
 */
async function takeBrowserScreenshot(locator, browserName, testName, options = {}) {
  const browserPrefix = getBrowserPrefix(browserName);
  const filename = `${browserPrefix}-${testName}.png`;
  
  await locator.screenshot({
    path: `test-results/cross-browser-artifacts/${filename}`,
    ...options
  });
}

/**
 * Test browser-specific CSS compatibility
 * @param {Object} page - Playwright page object
 * @param {string} browserName - Browser name
 * @param {string} selector - CSS selector to test
 * @param {string} testName - Test name for filename
 * @returns {Promise<void>}
 */
async function testCSSCrossBrowser(page, browserName, selector, testName) {
  const element = page.locator(selector);
  
  if (await element.count() === 0) {
    console.warn(`Element not found: ${selector}`);
    return;
  }

  await waitForBrowserRendering(page, browserName);
  await takeBrowserScreenshot(element, browserName, testName);
}

/**
 * Test responsive design across browsers
 * @param {Object} page - Playwright page object
 * @param {string} browserName - Browser name
 * @param {string} viewport - Viewport size ('mobile', 'tablet', 'desktop')
 * @param {string} testName - Test name for filename
 * @returns {Promise<void>}
 */
async function testResponsiveCrossBrowser(page, browserName, viewport, testName) {
  const viewportSizes = {
    'mobile': { width: 375, height: 667 },
    'tablet': { width: 768, height: 1024 },
    'desktop': { width: 1920, height: 1080 }
  };

  const size = viewportSizes[viewport];
  if (!size) {
    throw new Error(`Unknown viewport: ${viewport}`);
  }

  await page.setViewportSize(size);
  await waitForBrowserRendering(page, browserName);
  
  const browserPrefix = getBrowserPrefix(browserName);
  const filename = `${browserPrefix}-${viewport}-${testName}.png`;
  
  await page.screenshot({
    path: `test-results/cross-browser-artifacts/${filename}`,
    fullPage: true
  });
}

/**
 * Test browser-specific JavaScript features
 * @param {Object} page - Playwright page object
 * @param {string} browserName - Browser name
 * @param {string} feature - JavaScript feature to test
 * @returns {Promise<boolean>} Whether the feature is supported
 */
async function testJavaScriptFeature(page, browserName, feature) {
  const featureTests = {
    'es6-arrow-functions': () => {
      const test = () => true;
      return test();
    },
    'es6-promises': () => {
      return Promise.resolve(true);
    },
    'es6-async-await': async () => {
      return await Promise.resolve(true);
    },
    'es6-template-literals': () => {
      const name = 'test';
      return `Hello ${name}`;
    },
    'es6-destructuring': () => {
      const obj = { a: 1, b: 2 };
      const { a, b } = obj;
      return a + b;
    }
  };

  const testFunction = featureTests[feature];
  if (!testFunction) {
    throw new Error(`Unknown JavaScript feature: ${feature}`);
  }

  try {
    const result = await page.evaluate(testFunction);
    return result;
  } catch (error) {
    console.warn(`JavaScript feature ${feature} not supported:`, error);
    return false;
  }
}

/**
 * Get browser-specific test configuration
 * @param {string} browserName - Browser name
 * @returns {Object} Browser-specific configuration
 */
function getBrowserConfig(browserName) {
  const configs = {
    'chromium': {
      supportsCSSGrid: true,
      supportsFlexbox: true,
      supportsBackdropFilter: true,
      supportsCSSVariables: true,
      supportsAnimations: true,
      supportsTransforms: true,
      renderingDelay: 100,
      fontLoadingDelay: 0
    },
    'firefox': {
      supportsCSSGrid: true,
      supportsFlexbox: true,
      supportsBackdropFilter: false, // Limited support
      supportsCSSVariables: true,
      supportsAnimations: true,
      supportsTransforms: true,
      renderingDelay: 150,
      fontLoadingDelay: 50
    },
    'webkit': {
      supportsCSSGrid: true,
      supportsFlexbox: true,
      supportsBackdropFilter: true,
      supportsCSSVariables: true,
      supportsAnimations: true,
      supportsTransforms: true,
      renderingDelay: 200,
      fontLoadingDelay: 100
    },
    'msedge': {
      supportsCSSGrid: true,
      supportsFlexbox: true,
      supportsBackdropFilter: true,
      supportsCSSVariables: true,
      supportsAnimations: true,
      supportsTransforms: true,
      renderingDelay: 100,
      fontLoadingDelay: 0
    }
  };

  return configs[browserName] || configs['chromium'];
}

/**
 * Validate cross-browser compatibility
 * @param {Object} page - Playwright page object
 * @param {string} browserName - Browser name
 * @param {Array<string>} features - Features to validate
 * @returns {Promise<Object>} Validation results
 */
async function validateCrossBrowserCompatibility(page, browserName, features = []) {
  const results = {
    browser: browserName,
    features: {},
    overall: true
  };

  for (const feature of features) {
    try {
      const supported = await supportsCSSFeature(page, feature);
      results.features[feature] = supported;
      if (!supported) {
        results.overall = false;
      }
    } catch (error) {
      results.features[feature] = false;
      results.overall = false;
      console.error(`Error validating feature ${feature}:`, error);
    }
  }

  return results;
}

module.exports = {
  getBrowserPrefix,
  supportsCSSFeature,
  waitForBrowserRendering,
  takeBrowserScreenshot,
  testCSSCrossBrowser,
  testResponsiveCrossBrowser,
  testJavaScriptFeature,
  getBrowserConfig,
  validateCrossBrowserCompatibility
};
