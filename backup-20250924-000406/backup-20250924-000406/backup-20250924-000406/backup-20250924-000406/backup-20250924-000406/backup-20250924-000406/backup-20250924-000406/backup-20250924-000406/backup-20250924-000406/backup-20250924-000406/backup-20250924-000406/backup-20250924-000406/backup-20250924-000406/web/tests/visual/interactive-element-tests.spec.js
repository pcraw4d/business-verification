// web/tests/visual/interactive-element-tests.spec.js
const { test, expect } = require('@playwright/test');
const { navigateToDashboard, setViewportSize, waitForElementStable, setRiskState, waitForPageStable } = require('../utils/test-helpers');
const testData = require('../fixtures/test-data.json');

test.describe('Interactive Element Visual Regression Tests', () => {
  
  // Hover State Tests
  test.describe('Hover State Tests', () => {
    test('risk dashboard - button hover states', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test primary button hover
      const primaryButton = page.locator('button[type="submit"], .btn-primary, .primary-button');
      if (await primaryButton.count() > 0) {
        await primaryButton.first().hover();
        await page.waitForTimeout(300); // Wait for hover animation
        await expect(primaryButton.first()).toHaveScreenshot('interactive-button-hover-primary.png');
      }
      
      // Test secondary button hover
      const secondaryButton = page.locator('.btn-secondary, .secondary-button, button:not([type="submit"])');
      if (await secondaryButton.count() > 0) {
        await secondaryButton.first().hover();
        await page.waitForTimeout(300);
        await expect(secondaryButton.first()).toHaveScreenshot('interactive-button-hover-secondary.png');
      }
    });

    test('enhanced risk indicators - risk card hover states', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test risk card hover
      const riskCards = page.locator('.risk-card, .card, .indicator-card');
      if (await riskCards.count() > 0) {
        await riskCards.first().hover();
        await page.waitForTimeout(300);
        await expect(riskCards.first()).toHaveScreenshot('interactive-risk-card-hover.png');
        
        // Test multiple cards hover
        if (await riskCards.count() > 1) {
          await riskCards.nth(1).hover();
          await page.waitForTimeout(300);
          await expect(riskCards.nth(1)).toHaveScreenshot('interactive-risk-card-hover-2.png');
        }
      }
    });

    test('navigation elements - hover states', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test navigation link hover
      const navLinks = page.locator('nav a, .nav-link, .navigation a');
      if (await navLinks.count() > 0) {
        await navLinks.first().hover();
        await page.waitForTimeout(300);
        await expect(navLinks.first()).toHaveScreenshot('interactive-nav-link-hover.png');
      }
      
      // Test navigation button hover
      const navButtons = page.locator('nav button, .nav-button');
      if (await navButtons.count() > 0) {
        await navButtons.first().hover();
        await page.waitForTimeout(300);
        await expect(navButtons.first()).toHaveScreenshot('interactive-nav-button-hover.png');
      }
    });

    test('form elements - input hover states', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test input field hover
      const inputs = page.locator('input[type="text"], input[type="email"], input[type="tel"], textarea, select');
      if (await inputs.count() > 0) {
        await inputs.first().hover();
        await page.waitForTimeout(300);
        await expect(inputs.first()).toHaveScreenshot('interactive-input-hover.png');
      }
      
      // Test checkbox/radio hover
      const checkboxes = page.locator('input[type="checkbox"], input[type="radio"]');
      if (await checkboxes.count() > 0) {
        await checkboxes.first().hover();
        await page.waitForTimeout(300);
        await expect(checkboxes.first()).toHaveScreenshot('interactive-checkbox-hover.png');
      }
    });

    test('interactive elements - hover state consistency', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test multiple interactive elements hover consistency
      const interactiveElements = page.locator('button, .btn, .card, .interactive-element, [role="button"]');
      
      if (await interactiveElements.count() > 0) {
        for (let i = 0; i < Math.min(3, await interactiveElements.count()); i++) {
          const element = interactiveElements.nth(i);
          await element.hover();
          await page.waitForTimeout(300);
          await expect(element).toHaveScreenshot(`interactive-element-hover-${i + 1}.png`);
        }
      }
    });
  });

  // Tooltip Tests
  test.describe('Tooltip Tests', () => {
    test('risk indicators - tooltip display', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test tooltip on risk indicators
      const tooltipTriggers = page.locator('[data-tooltip], [title], .tooltip-trigger, .has-tooltip');
      if (await tooltipTriggers.count() > 0) {
        await tooltipTriggers.first().hover();
        await page.waitForTimeout(500); // Wait for tooltip to appear
        
        // Check if tooltip is visible
        const tooltip = page.locator('.tooltip, [role="tooltip"], .tooltip-content');
        if (await tooltip.count() > 0) {
          await expect(tooltip.first()).toHaveScreenshot('interactive-tooltip-risk-indicator.png');
        }
      }
    });

    test('form elements - tooltip display', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test tooltip on form elements
      const formTooltips = page.locator('input[title], select[title], textarea[title], .form-tooltip');
      if (await formTooltips.count() > 0) {
        await formTooltips.first().hover();
        await page.waitForTimeout(500);
        
        const tooltip = page.locator('.tooltip, [role="tooltip"], .tooltip-content');
        if (await tooltip.count() > 0) {
          await expect(tooltip.first()).toHaveScreenshot('interactive-tooltip-form.png');
        }
      }
    });

    test('navigation - tooltip display', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test tooltip on navigation elements
      const navTooltips = page.locator('nav [title], .nav-tooltip, .navigation [data-tooltip]');
      if (await navTooltips.count() > 0) {
        await navTooltips.first().hover();
        await page.waitForTimeout(500);
        
        const tooltip = page.locator('.tooltip, [role="tooltip"], .tooltip-content');
        if (await tooltip.count() > 0) {
          await expect(tooltip.first()).toHaveScreenshot('interactive-tooltip-navigation.png');
        }
      }
    });

    test('tooltip positioning and styling', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Create test tooltip if none exist
      await page.evaluate(() => {
        const testElement = document.createElement('div');
        testElement.className = 'test-tooltip-trigger';
        testElement.setAttribute('data-tooltip', 'This is a test tooltip');
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
            bottom: 100%;
            left: 50%;
            transform: translateX(-50%);
            background: #2c3e50;
            color: white;
            padding: 8px 12px;
            border-radius: 4px;
            font-size: 14px;
            white-space: nowrap;
            z-index: 1000;
            margin-bottom: 5px;
          }
          .test-tooltip-trigger:hover::before {
            content: '';
            position: absolute;
            bottom: 100%;
            left: 50%;
            transform: translateX(-50%);
            border: 5px solid transparent;
            border-top-color: #2c3e50;
            z-index: 1000;
          }
        `;
        document.head.appendChild(style);
      });
      
      const testTrigger = page.locator('.test-tooltip-trigger');
      await testTrigger.hover();
      await page.waitForTimeout(500);
      await expect(testTrigger).toHaveScreenshot('interactive-tooltip-positioning.png');
    });

    test('tooltip responsive behavior', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test tooltip behavior on mobile (should show on tap)
      const tooltipTriggers = page.locator('[data-tooltip], [title], .tooltip-trigger');
      if (await tooltipTriggers.count() > 0) {
        await tooltipTriggers.first().tap();
        await page.waitForTimeout(500);
        
        const tooltip = page.locator('.tooltip, [role="tooltip"], .tooltip-content');
        if (await tooltip.count() > 0) {
          await expect(tooltip.first()).toHaveScreenshot('interactive-tooltip-mobile.png');
        }
      }
    });
  });

  // Animation State Tests
  test.describe('Animation State Tests', () => {
    test('risk level transitions - animation states', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      const riskIndicator = page.locator('.risk-level-indicator, .risk-indicator');
      
      if (await riskIndicator.count() > 0) {
        // Test transition from low to high
        await setRiskState(page, 'low');
        await page.waitForTimeout(300);
        await expect(riskIndicator.first()).toHaveScreenshot('interactive-animation-low-state.png');
        
        // Transition to high with animation
        await setRiskState(page, 'high');
        await page.waitForTimeout(150); // Capture mid-animation
        await expect(riskIndicator.first()).toHaveScreenshot('interactive-animation-transition.png');
        
        // Wait for animation to complete
        await page.waitForTimeout(500);
        await expect(riskIndicator.first()).toHaveScreenshot('interactive-animation-high-state.png');
      }
    });

    test('loading animations - spinner states', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Create loading spinner for testing
      await page.evaluate(() => {
        const spinner = document.createElement('div');
        spinner.className = 'test-spinner';
        spinner.innerHTML = `
          <div class="spinner-container">
            <div class="spinner"></div>
            <p>Loading...</p>
          </div>
        `;
        spinner.style.cssText = `
          position: fixed;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          z-index: 9999;
        `;
        document.body.appendChild(spinner);
        
        const style = document.createElement('style');
        style.textContent = `
          .spinner-container {
            text-align: center;
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.1);
          }
          .spinner {
            width: 40px;
            height: 40px;
            border: 4px solid #f3f3f3;
            border-top: 4px solid #3498db;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin: 0 auto 15px;
          }
          @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
          }
        `;
        document.head.appendChild(style);
      });
      
      const spinner = page.locator('.test-spinner');
      await expect(spinner).toHaveScreenshot('interactive-animation-spinner.png');
    });

    test('button click animations', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      const buttons = page.locator('button, .btn, [role="button"]');
      if (await buttons.count() > 0) {
        const button = buttons.first();
        
        // Test button click animation
        await button.click();
        await page.waitForTimeout(100); // Capture click animation
        await expect(button).toHaveScreenshot('interactive-animation-button-click.png');
        
        // Wait for animation to complete
        await page.waitForTimeout(300);
        await expect(button).toHaveScreenshot('interactive-animation-button-complete.png');
      }
    });

    test('card hover animations', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      const cards = page.locator('.card, .risk-card, .indicator-card');
      if (await cards.count() > 0) {
        const card = cards.first();
        
        // Test card hover animation
        await card.hover();
        await page.waitForTimeout(150); // Capture hover animation
        await expect(card).toHaveScreenshot('interactive-animation-card-hover.png');
        
        // Test card leave animation
        await card.hover({ position: { x: -10, y: -10 } }); // Move away
        await page.waitForTimeout(150);
        await expect(card).toHaveScreenshot('interactive-animation-card-leave.png');
      }
    });

    test('form validation animations', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Create form validation animation
      await page.evaluate(() => {
        const form = document.querySelector('#riskAssessmentForm');
        if (form) {
          const input = form.querySelector('input[type="text"]');
          if (input) {
            input.classList.add('error');
            input.setAttribute('aria-invalid', 'true');
            
            const errorDiv = document.createElement('div');
            errorDiv.className = 'error-message';
            errorDiv.textContent = 'This field is required';
            errorDiv.style.cssText = `
              color: #e74c3c;
              font-size: 12px;
              margin-top: 5px;
              animation: shake 0.5s ease-in-out;
            `;
            input.parentNode.appendChild(errorDiv);
            
            const style = document.createElement('style');
            style.textContent = `
              .error {
                border-color: #e74c3c !important;
                box-shadow: 0 0 0 2px rgba(231, 76, 60, 0.2) !important;
                animation: errorPulse 0.5s ease-in-out;
              }
              @keyframes shake {
                0%, 100% { transform: translateX(0); }
                25% { transform: translateX(-5px); }
                75% { transform: translateX(5px); }
              }
              @keyframes errorPulse {
                0% { box-shadow: 0 0 0 0 rgba(231, 76, 60, 0.4); }
                100% { box-shadow: 0 0 0 2px rgba(231, 76, 60, 0.2); }
              }
            `;
            document.head.appendChild(style);
          }
        }
      });
      
      const errorInput = page.locator('input.error');
      if (await errorInput.count() > 0) {
        await expect(errorInput.first()).toHaveScreenshot('interactive-animation-form-error.png');
      }
    });
  });

  // Focus State Tests
  test.describe('Focus State Tests', () => {
    test('form inputs - focus states', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test input focus
      const inputs = page.locator('input[type="text"], input[type="email"], input[type="tel"], textarea');
      if (await inputs.count() > 0) {
        await inputs.first().focus();
        await page.waitForTimeout(300);
        await expect(inputs.first()).toHaveScreenshot('interactive-focus-input.png');
      }
      
      // Test select focus
      const selects = page.locator('select');
      if (await selects.count() > 0) {
        await selects.first().focus();
        await page.waitForTimeout(300);
        await expect(selects.first()).toHaveScreenshot('interactive-focus-select.png');
      }
    });

    test('buttons - focus states', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      const buttons = page.locator('button, .btn, [role="button"]');
      if (await buttons.count() > 0) {
        await buttons.first().focus();
        await page.waitForTimeout(300);
        await expect(buttons.first()).toHaveScreenshot('interactive-focus-button.png');
      }
    });

    test('navigation - focus states', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test navigation link focus
      const navLinks = page.locator('nav a, .nav-link');
      if (await navLinks.count() > 0) {
        await navLinks.first().focus();
        await page.waitForTimeout(300);
        await expect(navLinks.first()).toHaveScreenshot('interactive-focus-nav-link.png');
      }
      
      // Test navigation button focus
      const navButtons = page.locator('nav button, .nav-button');
      if (await navButtons.count() > 0) {
        await navButtons.first().focus();
        await page.waitForTimeout(300);
        await expect(navButtons.first()).toHaveScreenshot('interactive-focus-nav-button.png');
      }
    });

    test('interactive elements - focus ring consistency', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test focus ring consistency across interactive elements
      const interactiveElements = page.locator('button, .btn, input, select, textarea, a, [tabindex]');
      
      if (await interactiveElements.count() > 0) {
        for (let i = 0; i < Math.min(3, await interactiveElements.count()); i++) {
          const element = interactiveElements.nth(i);
          await element.focus();
          await page.waitForTimeout(300);
          await expect(element).toHaveScreenshot(`interactive-focus-element-${i + 1}.png`);
        }
      }
    });

    test('keyboard navigation - focus management', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Test Tab key navigation
      await page.keyboard.press('Tab');
      await page.waitForTimeout(300);
      await expect(page).toHaveScreenshot('interactive-focus-tab-1.png');
      
      await page.keyboard.press('Tab');
      await page.waitForTimeout(300);
      await expect(page).toHaveScreenshot('interactive-focus-tab-2.png');
      
      await page.keyboard.press('Tab');
      await page.waitForTimeout(300);
      await expect(page).toHaveScreenshot('interactive-focus-tab-3.png');
    });

    test('focus trap - modal focus management', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Create test modal for focus trap testing
      await page.evaluate(() => {
        const modal = document.createElement('div');
        modal.className = 'test-modal';
        modal.innerHTML = `
          <div class="modal-overlay">
            <div class="modal-content">
              <h3>Test Modal</h3>
              <p>This is a test modal for focus trap testing.</p>
              <button class="modal-button">Button 1</button>
              <button class="modal-button">Button 2</button>
              <button class="modal-close">Close</button>
            </div>
          </div>
        `;
        modal.style.cssText = `
          position: fixed;
          top: 0;
          left: 0;
          width: 100%;
          height: 100%;
          z-index: 10000;
        `;
        document.body.appendChild(modal);
        
        const style = document.createElement('style');
        style.textContent = `
          .modal-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.5);
            display: flex;
            justify-content: center;
            align-items: center;
          }
          .modal-content {
            background: white;
            padding: 30px;
            border-radius: 8px;
            max-width: 400px;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
          }
          .modal-button {
            margin: 5px;
            padding: 10px 20px;
            background: #3498db;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
          }
          .modal-close {
            margin: 5px;
            padding: 10px 20px;
            background: #e74c3c;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
          }
        `;
        document.head.appendChild(style);
      });
      
      const modal = page.locator('.test-modal');
      await expect(modal).toHaveScreenshot('interactive-focus-modal.png');
      
      // Test focus on first button
      const firstButton = page.locator('.modal-button').first();
      await firstButton.focus();
      await page.waitForTimeout(300);
      await expect(firstButton).toHaveScreenshot('interactive-focus-modal-button.png');
    });
  });

  // Responsive Interactive Tests
  test.describe('Responsive Interactive Tests', () => {
    test('mobile - touch interactions', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test mobile touch interactions
      const touchElements = page.locator('button, .btn, .card, .interactive-element');
      if (await touchElements.count() > 0) {
        await touchElements.first().tap();
        await page.waitForTimeout(300);
        await expect(touchElements.first()).toHaveScreenshot('interactive-mobile-touch.png');
      }
    });

    test('tablet - hover and touch interactions', async ({ page }) => {
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test tablet interactions (both hover and touch)
      const interactiveElements = page.locator('button, .btn, .card');
      if (await interactiveElements.count() > 0) {
        // Test hover on tablet
        await interactiveElements.first().hover();
        await page.waitForTimeout(300);
        await expect(interactiveElements.first()).toHaveScreenshot('interactive-tablet-hover.png');
        
        // Test touch on tablet
        await interactiveElements.first().tap();
        await page.waitForTimeout(300);
        await expect(interactiveElements.first()).toHaveScreenshot('interactive-tablet-touch.png');
      }
    });

    test('desktop - full interaction suite', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      const interactiveElement = page.locator('button, .btn, .card').first();
      if (await interactiveElement.count() > 0) {
        // Test hover
        await interactiveElement.hover();
        await page.waitForTimeout(300);
        await expect(interactiveElement).toHaveScreenshot('interactive-desktop-hover.png');
        
        // Test focus
        await interactiveElement.focus();
        await page.waitForTimeout(300);
        await expect(interactiveElement).toHaveScreenshot('interactive-desktop-focus.png');
        
        // Test click
        await interactiveElement.click();
        await page.waitForTimeout(300);
        await expect(interactiveElement).toHaveScreenshot('interactive-desktop-click.png');
      }
    });
  });

  // Accessibility Interactive Tests
  test.describe('Accessibility Interactive Tests', () => {
    test('keyboard navigation - arrow keys', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test arrow key navigation
      const focusableElements = page.locator('button, input, select, textarea, a, [tabindex]');
      if (await focusableElements.count() > 0) {
        await focusableElements.first().focus();
        await page.waitForTimeout(300);
        await expect(focusableElements.first()).toHaveScreenshot('interactive-a11y-arrow-start.png');
        
        await page.keyboard.press('ArrowRight');
        await page.waitForTimeout(300);
        await expect(page).toHaveScreenshot('interactive-a11y-arrow-right.png');
        
        await page.keyboard.press('ArrowDown');
        await page.waitForTimeout(300);
        await expect(page).toHaveScreenshot('interactive-a11y-arrow-down.png');
      }
    });

    test('screen reader - aria labels and roles', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Test elements with ARIA labels
      const ariaElements = page.locator('[aria-label], [aria-labelledby], [role]');
      if (await ariaElements.count() > 0) {
        await ariaElements.first().focus();
        await page.waitForTimeout(300);
        await expect(ariaElements.first()).toHaveScreenshot('interactive-a11y-aria-element.png');
      }
    });

    test('high contrast mode - interactive elements', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Simulate high contrast mode
      await page.evaluate(() => {
        const style = document.createElement('style');
        style.textContent = `
          * {
            background: white !important;
            color: black !important;
            border-color: black !important;
          }
          button, .btn {
            background: black !important;
            color: white !important;
            border: 2px solid black !important;
          }
          button:hover, .btn:hover {
            background: white !important;
            color: black !important;
            border: 2px solid black !important;
          }
          button:focus, .btn:focus {
            outline: 3px solid black !important;
            outline-offset: 2px !important;
          }
        `;
        document.head.appendChild(style);
      });
      
      const interactiveElement = page.locator('button, .btn').first();
      if (await interactiveElement.count() > 0) {
        await interactiveElement.focus();
        await page.waitForTimeout(300);
        await expect(interactiveElement).toHaveScreenshot('interactive-a11y-high-contrast.png');
      }
    });
  });
});
