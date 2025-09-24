// web/tests/utils/interactive-helpers.js
/**
 * Interactive element testing utilities for visual regression testing
 * Provides helper functions for testing user interactions and animations
 */

/**
 * Simulate hover state on element
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} selector - CSS selector for element
 * @param {Object} options - Hover options
 */
async function simulateHover(page, selector, options = {}) {
  const defaultOptions = {
    waitForAnimation: true,
    animationDelay: 300,
    ...options
  };

  const element = page.locator(selector);
  if (await element.count() === 0) {
    console.warn(`Element not found: ${selector}`);
    return false;
  }

  await element.hover();
  
  if (defaultOptions.waitForAnimation) {
    await page.waitForTimeout(defaultOptions.animationDelay);
  }

  console.log(`🖱️ Simulated hover on: ${selector}`);
  return true;
}

/**
 * Simulate focus state on element
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} selector - CSS selector for element
 * @param {Object} options - Focus options
 */
async function simulateFocus(page, selector, options = {}) {
  const defaultOptions = {
    waitForFocus: true,
    focusDelay: 300,
    ...options
  };

  const element = page.locator(selector);
  if (await element.count() === 0) {
    console.warn(`Element not found: ${selector}`);
    return false;
  }

  await element.focus();
  
  if (defaultOptions.waitForFocus) {
    await page.waitForTimeout(defaultOptions.focusDelay);
  }

  console.log(`⌨️ Simulated focus on: ${selector}`);
  return true;
}

/**
 * Simulate click interaction on element
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} selector - CSS selector for element
 * @param {Object} options - Click options
 */
async function simulateClick(page, selector, options = {}) {
  const defaultOptions = {
    waitForAnimation: true,
    animationDelay: 300,
    ...options
  };

  const element = page.locator(selector);
  if (await element.count() === 0) {
    console.warn(`Element not found: ${selector}`);
    return false;
  }

  await element.click();
  
  if (defaultOptions.waitForAnimation) {
    await page.waitForTimeout(defaultOptions.animationDelay);
  }

  console.log(`👆 Simulated click on: ${selector}`);
  return true;
}

/**
 * Simulate touch interaction on element (mobile)
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} selector - CSS selector for element
 * @param {Object} options - Touch options
 */
async function simulateTouch(page, selector, options = {}) {
  const defaultOptions = {
    waitForAnimation: true,
    animationDelay: 300,
    ...options
  };

  const element = page.locator(selector);
  if (await element.count() === 0) {
    console.warn(`Element not found: ${selector}`);
    return false;
  }

  await element.tap();
  
  if (defaultOptions.waitForAnimation) {
    await page.waitForTimeout(defaultOptions.animationDelay);
  }

  console.log(`👆 Simulated touch on: ${selector}`);
  return true;
}

/**
 * Simulate keyboard navigation
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {Array} keys - Array of keys to press
 * @param {Object} options - Navigation options
 */
async function simulateKeyboardNavigation(page, keys, options = {}) {
  const defaultOptions = {
    keyDelay: 300,
    ...options
  };

  for (const key of keys) {
    await page.keyboard.press(key);
    await page.waitForTimeout(defaultOptions.keyDelay);
  }

  console.log(`⌨️ Simulated keyboard navigation: ${keys.join(', ')}`);
}

/**
 * Wait for animation to complete
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} selector - CSS selector for animated element
 * @param {Object} options - Animation options
 */
async function waitForAnimation(page, selector, options = {}) {
  const defaultOptions = {
    timeout: 5000,
    ...options
  };

  try {
    await page.waitForFunction((sel) => {
      const element = document.querySelector(sel);
      if (!element) return true;
      
      const animations = element.getAnimations();
      return animations.every(animation => 
        animation.playState === 'finished' || animation.playState === 'idle'
      );
    }, selector, { timeout: defaultOptions.timeout });
    
    console.log(`🎬 Animation completed for: ${selector}`);
    return true;
  } catch (error) {
    console.warn(`Animation timeout for: ${selector}`);
    return false;
  }
}

/**
 * Capture tooltip state
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} triggerSelector - CSS selector for tooltip trigger
 * @param {Object} options - Tooltip options
 */
async function captureTooltip(page, triggerSelector, options = {}) {
  const defaultOptions = {
    tooltipSelector: '.tooltip, [role="tooltip"], .tooltip-content',
    hoverDelay: 500,
    ...options
  };

  const trigger = page.locator(triggerSelector);
  if (await trigger.count() === 0) {
    console.warn(`Tooltip trigger not found: ${triggerSelector}`);
    return null;
  }

  await trigger.hover();
  await page.waitForTimeout(defaultOptions.hoverDelay);

  const tooltip = page.locator(defaultOptions.tooltipSelector);
  if (await tooltip.count() > 0) {
    console.log(`💬 Captured tooltip for: ${triggerSelector}`);
    return tooltip.first();
  }

  console.warn(`Tooltip not found for: ${triggerSelector}`);
  return null;
}

/**
 * Test focus trap in modal
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} modalSelector - CSS selector for modal
 * @param {Object} options - Focus trap options
 */
async function testFocusTrap(page, modalSelector, options = {}) {
  const defaultOptions = {
    focusableSelector: 'button, input, select, textarea, a, [tabindex]:not([tabindex="-1"])',
    ...options
  };

  const modal = page.locator(modalSelector);
  if (await modal.count() === 0) {
    console.warn(`Modal not found: ${modalSelector}`);
    return false;
  }

  const focusableElements = modal.locator(defaultOptions.focusableSelector);
  const count = await focusableElements.count();
  
  if (count === 0) {
    console.warn(`No focusable elements found in modal: ${modalSelector}`);
    return false;
  }

  // Test focus on first element
  await focusableElements.first().focus();
  await page.waitForTimeout(300);

  // Test Tab navigation
  for (let i = 0; i < count; i++) {
    await page.keyboard.press('Tab');
    await page.waitForTimeout(300);
  }

  console.log(`🎯 Tested focus trap in modal: ${modalSelector}`);
  return true;
}

/**
 * Simulate form validation states
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} formSelector - CSS selector for form
 * @param {Array} validationErrors - Array of validation errors
 */
async function simulateFormValidation(page, formSelector, validationErrors = []) {
  const form = page.locator(formSelector);
  if (await form.count() === 0) {
    console.warn(`Form not found: ${formSelector}`);
    return false;
  }

  await page.evaluate(({ formSelector, validationErrors }) => {
    const form = document.querySelector(formSelector);
    if (!form) return;

    // Clear existing validation states
    const existingErrors = form.querySelectorAll('.error, .error-message, .form-error-summary');
    existingErrors.forEach(el => el.remove());

    // Add validation errors
    validationErrors.forEach(error => {
      const field = form.querySelector(`[name="${error.field}"]`);
      if (field) {
        field.classList.add('error');
        field.setAttribute('aria-invalid', 'true');
        
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.textContent = error.message;
        field.parentNode.appendChild(errorDiv);
      }
    });

    // Add form error summary
    if (validationErrors.length > 0) {
      const errorSummary = document.createElement('div');
      errorSummary.className = 'form-error-summary';
      errorSummary.innerHTML = `
        <h4>Please correct the following errors:</h4>
        <ul>
          ${validationErrors.map(error => `<li>${error.message}</li>`).join('')}
        </ul>
      `;
      form.insertBefore(errorSummary, form.firstChild);
    }

    // Add validation styles
    const style = document.createElement('style');
    style.textContent = `
      .error {
        border-color: #e74c3c !important;
        box-shadow: 0 0 0 2px rgba(231, 76, 60, 0.2) !important;
        animation: errorPulse 0.5s ease-in-out;
      }
      .error-message {
        color: #e74c3c;
        font-size: 12px;
        margin-top: 5px;
        display: block;
        animation: shake 0.5s ease-in-out;
      }
      .form-error-summary {
        background: #fff5f5;
        border: 1px solid #e74c3c;
        border-radius: 4px;
        padding: 15px;
        margin-bottom: 20px;
      }
      .form-error-summary h4 {
        color: #e74c3c;
        margin-bottom: 10px;
      }
      .form-error-summary ul {
        margin: 0;
        padding-left: 20px;
      }
      .form-error-summary li {
        color: #666;
        margin-bottom: 5px;
      }
      @keyframes errorPulse {
        0% { box-shadow: 0 0 0 0 rgba(231, 76, 60, 0.4); }
        100% { box-shadow: 0 0 0 2px rgba(231, 76, 60, 0.2); }
      }
      @keyframes shake {
        0%, 100% { transform: translateX(0); }
        25% { transform: translateX(-5px); }
        75% { transform: translateX(5px); }
      }
    `;
    document.head.appendChild(style);
  }, { formSelector, validationErrors });

  await page.waitForTimeout(500); // Wait for animations
  console.log(`📝 Simulated form validation with ${validationErrors.length} errors`);
  return true;
}

/**
 * Test responsive interactions
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} viewport - Viewport size (mobile, tablet, desktop)
 * @param {string} selector - CSS selector for element to test
 */
async function testResponsiveInteraction(page, viewport, selector) {
  const viewportSizes = {
    mobile: { width: 375, height: 667 },
    tablet: { width: 768, height: 1024 },
    desktop: { width: 1920, height: 1080 }
  };

  const size = viewportSizes[viewport];
  if (!size) {
    throw new Error(`Unknown viewport: ${viewport}`);
  }

  await page.setViewportSize(size);
  await page.waitForTimeout(300);

  const element = page.locator(selector);
  if (await element.count() === 0) {
    console.warn(`Element not found: ${selector}`);
    return false;
  }

  // Test appropriate interaction for viewport
  if (viewport === 'mobile') {
    await element.tap();
  } else {
    await element.hover();
  }

  await page.waitForTimeout(300);
  console.log(`📱 Tested ${viewport} interaction on: ${selector}`);
  return true;
}

/**
 * Create test tooltip for testing
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {Object} options - Tooltip options
 */
async function createTestTooltip(page, options = {}) {
  const defaultOptions = {
    text: 'Test tooltip',
    position: 'top',
    trigger: 'hover',
    ...options
  };

  await page.evaluate((options) => {
    const testElement = document.createElement('div');
    testElement.className = 'test-tooltip-trigger';
    testElement.setAttribute('data-tooltip', options.text);
    testElement.textContent = 'Hover for tooltip';
    testElement.style.cssText = `
      position: absolute;
      top: 100px;
      left: 100px;
      padding: 10px;
      background: #3498db;
      color: white;
      border-radius: 4px;
      cursor: pointer;
    `;
    document.body.appendChild(testElement);
    
    // Add tooltip styles
    const style = document.createElement('style');
    style.textContent = `
      .test-tooltip-trigger:hover::after {
        content: attr(data-tooltip);
        position: absolute;
        ${options.position === 'top' ? 'bottom: 100%;' : 'top: 100%;'}
        left: 50%;
        transform: translateX(-50%);
        background: #2c3e50;
        color: white;
        padding: 8px 12px;
        border-radius: 4px;
        font-size: 14px;
        white-space: nowrap;
        z-index: 1000;
        margin-${options.position === 'top' ? 'bottom' : 'top'}: 5px;
      }
      .test-tooltip-trigger:hover::before {
        content: '';
        position: absolute;
        ${options.position === 'top' ? 'bottom: 100%;' : 'top: 100%;'}
        left: 50%;
        transform: translateX(-50%);
        border: 5px solid transparent;
        border-${options.position === 'top' ? 'top' : 'bottom'}-color: #2c3e50;
        z-index: 1000;
      }
    `;
    document.head.appendChild(style);
  }, defaultOptions);

  console.log(`💬 Created test tooltip: ${defaultOptions.text}`);
  return '.test-tooltip-trigger';
}

/**
 * Test animation performance
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} selector - CSS selector for animated element
 * @param {Object} options - Performance options
 */
async function testAnimationPerformance(page, selector, options = {}) {
  const defaultOptions = {
    duration: 1000,
    ...options
  };

  const element = page.locator(selector);
  if (await element.count() === 0) {
    console.warn(`Element not found: ${selector}`);
    return null;
  }

  const startTime = Date.now();
  
  // Trigger animation
  await element.hover();
  
  // Wait for animation to complete
  await waitForAnimation(page, selector, { timeout: defaultOptions.duration + 1000 });
  
  const endTime = Date.now();
  const actualDuration = endTime - startTime;
  
  console.log(`🎬 Animation performance: ${actualDuration}ms for ${selector}`);
  return {
    expectedDuration: defaultOptions.duration,
    actualDuration,
    performance: actualDuration <= defaultOptions.duration ? 'good' : 'slow'
  };
}

/**
 * Clean up test elements
 * @param {import('@playwright/test').Page} page - Playwright page object
 */
async function cleanupTestElements(page) {
  await page.evaluate(() => {
    // Remove test elements
    const testElements = document.querySelectorAll('.test-tooltip-trigger, .test-modal, .test-spinner');
    testElements.forEach(el => el.remove());
    
    // Remove test styles
    const testStyles = document.querySelectorAll('style[data-test-interactive]');
    testStyles.forEach(style => style.remove());
  });

  console.log('🧹 Cleaned up test elements');
}

module.exports = {
  simulateHover,
  simulateFocus,
  simulateClick,
  simulateTouch,
  simulateKeyboardNavigation,
  waitForAnimation,
  captureTooltip,
  testFocusTrap,
  simulateFormValidation,
  testResponsiveInteraction,
  createTestTooltip,
  testAnimationPerformance,
  cleanupTestElements
};
