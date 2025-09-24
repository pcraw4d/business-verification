// web/tests/visual/state-based-tests.spec.js
const { test, expect } = require('@playwright/test');
const { navigateToDashboard, setViewportSize, waitForElementStable, setRiskState, waitForPageStable } = require('../utils/test-helpers');
const testData = require('../fixtures/test-data.json');

test.describe('State-Based Visual Regression Tests', () => {
  
  // Risk Level State Tests
  test.describe('Risk Level State Tests', () => {
    test('risk dashboard - low risk state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Set low risk state
      await setRiskState(page, 'low');
      await waitForElementStable(page, '.risk-level-indicator');
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('state-risk-dashboard-low.png');
      
      // Take component-level screenshot
      const riskIndicator = page.locator('.risk-level-indicator');
      await expect(riskIndicator).toHaveScreenshot('state-risk-indicator-low.png');
    });

    test('risk dashboard - medium risk state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Set medium risk state
      await setRiskState(page, 'medium');
      await waitForElementStable(page, '.risk-level-indicator');
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('state-risk-dashboard-medium.png');
      
      // Take component-level screenshot
      const riskIndicator = page.locator('.risk-level-indicator');
      await expect(riskIndicator).toHaveScreenshot('state-risk-indicator-medium.png');
    });

    test('risk dashboard - high risk state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Set high risk state
      await setRiskState(page, 'high');
      await waitForElementStable(page, '.risk-level-indicator');
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('state-risk-dashboard-high.png');
      
      // Take component-level screenshot
      const riskIndicator = page.locator('.risk-level-indicator');
      await expect(riskIndicator).toHaveScreenshot('state-risk-indicator-high.png');
    });

    test('risk dashboard - critical risk state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Set critical risk state
      await setRiskState(page, 'critical');
      await waitForElementStable(page, '.risk-level-indicator');
      
      // Take full page screenshot
      await expect(page).toHaveScreenshot('state-risk-dashboard-critical.png');
      
      // Take component-level screenshot
      const riskIndicator = page.locator('.risk-level-indicator');
      await expect(riskIndicator).toHaveScreenshot('state-risk-indicator-critical.png');
    });

    test('enhanced risk indicators - all risk levels comparison', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      const riskLevels = ['low', 'medium', 'high', 'critical'];
      
      for (const level of riskLevels) {
        await setRiskState(page, level);
        await waitForElementStable(page, '.risk-level-indicator');
        
        // Take screenshot for each risk level
        await expect(page).toHaveScreenshot(`state-enhanced-indicators-${level}.png`);
        
        // Take component screenshot
        const riskIndicator = page.locator('.risk-level-indicator');
        await expect(riskIndicator).toHaveScreenshot(`state-enhanced-indicator-${level}.png`);
      }
    });

    test('risk level transitions - smooth state changes', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      const riskIndicator = page.locator('.risk-level-indicator');
      
      // Test transition from low to high
      await setRiskState(page, 'low');
      await waitForElementStable(page, '.risk-level-indicator');
      await expect(riskIndicator).toHaveScreenshot('state-transition-low.png');
      
      // Transition to high
      await setRiskState(page, 'high');
      await page.waitForTimeout(500); // Wait for transition animation
      await expect(riskIndicator).toHaveScreenshot('state-transition-high.png');
      
      // Transition to critical
      await setRiskState(page, 'critical');
      await page.waitForTimeout(500); // Wait for transition animation
      await expect(riskIndicator).toHaveScreenshot('state-transition-critical.png');
    });
  });

  // Loading State Tests
  test.describe('Loading State Tests', () => {
    test('risk dashboard - loading state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      
      // Navigate to dashboard and simulate loading state
      await page.goto('http://localhost:8080/risk-dashboard.html');
      
      // Inject loading state before page fully loads
      await page.evaluate(() => {
        // Add loading overlay
        const loadingOverlay = document.createElement('div');
        loadingOverlay.id = 'loading-overlay';
        loadingOverlay.className = 'loading-overlay';
        loadingOverlay.innerHTML = `
          <div class="loading-spinner">
            <div class="spinner"></div>
            <p>Loading risk assessment...</p>
          </div>
        `;
        document.body.appendChild(loadingOverlay);
        
        // Add loading styles
        const style = document.createElement('style');
        style.textContent = `
          .loading-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(255, 255, 255, 0.9);
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 9999;
          }
          .loading-spinner {
            text-align: center;
          }
          .spinner {
            width: 40px;
            height: 40px;
            border: 4px solid #f3f3f3;
            border-top: 4px solid #3498db;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin: 0 auto 20px;
          }
          @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
          }
        `;
        document.head.appendChild(style);
      });
      
      await waitForElementStable(page, '#loading-overlay');
      await expect(page).toHaveScreenshot('state-loading-overlay.png');
      
      // Test loading spinner component
      const loadingSpinner = page.locator('.loading-spinner');
      await expect(loadingSpinner).toHaveScreenshot('state-loading-spinner.png');
    });

    test('enhanced risk indicators - loading state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Simulate loading state for risk indicators
      await page.evaluate(() => {
        const riskCards = document.querySelectorAll('.risk-card');
        riskCards.forEach(card => {
          card.classList.add('loading');
          card.innerHTML = `
            <div class="loading-skeleton">
              <div class="skeleton-line"></div>
              <div class="skeleton-line short"></div>
              <div class="skeleton-circle"></div>
            </div>
          `;
        });
        
        // Add skeleton loading styles
        const style = document.createElement('style');
        style.textContent = `
          .loading-skeleton {
            padding: 20px;
          }
          .skeleton-line {
            height: 12px;
            background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
            background-size: 200% 100%;
            animation: loading 1.5s infinite;
            border-radius: 6px;
            margin-bottom: 10px;
          }
          .skeleton-line.short {
            width: 60%;
          }
          .skeleton-circle {
            width: 40px;
            height: 40px;
            border-radius: 50%;
            background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
            background-size: 200% 100%;
            animation: loading 1.5s infinite;
          }
          @keyframes loading {
            0% { background-position: 200% 0; }
            100% { background-position: -200% 0; }
          }
        `;
        document.head.appendChild(style);
      });
      
      await waitForElementStable(page, '.loading-skeleton');
      await expect(page).toHaveScreenshot('state-loading-skeleton.png');
    });

    test('charts loading state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Simulate chart loading state
      await page.evaluate(() => {
        const chartContainers = document.querySelectorAll('.chart-container');
        chartContainers.forEach(container => {
          container.innerHTML = `
            <div class="chart-loading">
              <div class="chart-skeleton">
                <div class="skeleton-bars">
                  <div class="skeleton-bar" style="height: 60%"></div>
                  <div class="skeleton-bar" style="height: 80%"></div>
                  <div class="skeleton-bar" style="height: 40%"></div>
                  <div class="skeleton-bar" style="height: 90%"></div>
                  <div class="skeleton-bar" style="height: 70%"></div>
                </div>
                <p class="loading-text">Loading chart data...</p>
              </div>
            </div>
          `;
        });
        
        // Add chart loading styles
        const style = document.createElement('style');
        style.textContent = `
          .chart-loading {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 300px;
            background: #f8f9fa;
            border-radius: 8px;
          }
          .chart-skeleton {
            text-align: center;
          }
          .skeleton-bars {
            display: flex;
            align-items: end;
            justify-content: center;
            gap: 10px;
            height: 200px;
            margin-bottom: 20px;
          }
          .skeleton-bar {
            width: 20px;
            background: linear-gradient(90deg, #e0e0e0 25%, #d0d0d0 50%, #e0e0e0 75%);
            background-size: 200% 100%;
            animation: loading 1.5s infinite;
            border-radius: 4px 4px 0 0;
          }
          .loading-text {
            color: #666;
            font-size: 14px;
            margin: 0;
          }
        `;
        document.head.appendChild(style);
      });
      
      await waitForElementStable(page, '.chart-loading');
      await expect(page).toHaveScreenshot('state-chart-loading.png');
    });
  });

  // Error State Tests
  test.describe('Error State Tests', () => {
    test('risk dashboard - API error state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Simulate API error state
      await page.evaluate(() => {
        // Add error overlay
        const errorOverlay = document.createElement('div');
        errorOverlay.id = 'error-overlay';
        errorOverlay.className = 'error-overlay';
        errorOverlay.innerHTML = `
          <div class="error-message">
            <div class="error-icon">⚠️</div>
            <h3>Unable to Load Risk Assessment</h3>
            <p>There was an error loading your risk assessment data. Please try again.</p>
            <button class="retry-button">Retry</button>
          </div>
        `;
        document.body.appendChild(errorOverlay);
        
        // Add error styles
        const style = document.createElement('style');
        style.textContent = `
          .error-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(255, 255, 255, 0.95);
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 9999;
          }
          .error-message {
            text-align: center;
            padding: 40px;
            background: white;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
            max-width: 400px;
          }
          .error-icon {
            font-size: 48px;
            margin-bottom: 20px;
          }
          .error-message h3 {
            color: #e74c3c;
            margin-bottom: 15px;
          }
          .error-message p {
            color: #666;
            margin-bottom: 25px;
            line-height: 1.5;
          }
          .retry-button {
            background: #3498db;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 16px;
          }
          .retry-button:hover {
            background: #2980b9;
          }
        `;
        document.head.appendChild(style);
      });
      
      await waitForElementStable(page, '#error-overlay');
      await expect(page).toHaveScreenshot('state-error-overlay.png');
      
      // Test error message component
      const errorMessage = page.locator('.error-message');
      await expect(errorMessage).toHaveScreenshot('state-error-message.png');
    });

    test('enhanced risk indicators - data error state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Simulate data error state
      await page.evaluate(() => {
        const riskCards = document.querySelectorAll('.risk-card');
        riskCards.forEach(card => {
          card.classList.add('error-state');
          card.innerHTML = `
            <div class="error-card">
              <div class="error-icon">❌</div>
              <h4>Data Unavailable</h4>
              <p>Unable to load risk data for this category.</p>
              <button class="refresh-button">Refresh</button>
            </div>
          `;
        });
        
        // Add error card styles
        const style = document.createElement('style');
        style.textContent = `
          .error-card {
            text-align: center;
            padding: 30px 20px;
            background: #fff5f5;
            border: 2px dashed #e74c3c;
            border-radius: 8px;
            color: #e74c3c;
          }
          .error-card .error-icon {
            font-size: 32px;
            margin-bottom: 15px;
          }
          .error-card h4 {
            margin-bottom: 10px;
            color: #e74c3c;
          }
          .error-card p {
            margin-bottom: 20px;
            color: #666;
            font-size: 14px;
          }
          .refresh-button {
            background: #e74c3c;
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
          }
          .refresh-button:hover {
            background: #c0392b;
          }
        `;
        document.head.appendChild(style);
      });
      
      await waitForElementStable(page, '.error-card');
      await expect(page).toHaveScreenshot('state-error-cards.png');
    });

    test('form validation error state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Simulate form validation errors
      await page.evaluate(() => {
        const form = document.querySelector('#riskAssessmentForm');
        if (form) {
          // Add error states to form fields
          const inputs = form.querySelectorAll('input, select, textarea');
          inputs.forEach(input => {
            input.classList.add('error');
            input.setAttribute('aria-invalid', 'true');
            
            // Add error message
            const errorDiv = document.createElement('div');
            errorDiv.className = 'error-message';
            errorDiv.textContent = 'This field is required';
            input.parentNode.appendChild(errorDiv);
          });
          
          // Add form error styles
          const style = document.createElement('style');
          style.textContent = `
            .error {
              border-color: #e74c3c !important;
              box-shadow: 0 0 0 2px rgba(231, 76, 60, 0.2) !important;
            }
            .error-message {
              color: #e74c3c;
              font-size: 12px;
              margin-top: 5px;
              display: block;
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
          `;
          document.head.appendChild(style);
          
          // Add form error summary
          const errorSummary = document.createElement('div');
          errorSummary.className = 'form-error-summary';
          errorSummary.innerHTML = `
            <h4>Please correct the following errors:</h4>
            <ul>
              <li>Business name is required</li>
              <li>Industry type is required</li>
              <li>Email address is invalid</li>
            </ul>
          `;
          form.insertBefore(errorSummary, form.firstChild);
        }
      });
      
      await waitForElementStable(page, '.form-error-summary');
      await expect(page).toHaveScreenshot('state-form-errors.png');
    });
  });

  // Empty Data State Tests
  test.describe('Empty Data State Tests', () => {
    test('risk dashboard - empty data state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Simulate empty data state
      await page.evaluate(() => {
        // Clear existing content
        const mainContent = document.querySelector('#main-content');
        if (mainContent) {
          mainContent.innerHTML = `
            <div class="empty-state">
              <div class="empty-icon">📊</div>
              <h3>No Risk Assessment Data</h3>
              <p>You haven't completed any risk assessments yet. Start by filling out the form below to get your first assessment.</p>
              <button class="cta-button">Start Risk Assessment</button>
            </div>
          `;
        }
        
        // Add empty state styles
        const style = document.createElement('style');
        style.textContent = `
          .empty-state {
            text-align: center;
            padding: 60px 20px;
            background: #f8f9fa;
            border-radius: 12px;
            margin: 40px 0;
          }
          .empty-icon {
            font-size: 64px;
            margin-bottom: 20px;
            opacity: 0.5;
          }
          .empty-state h3 {
            color: #2c3e50;
            margin-bottom: 15px;
            font-size: 24px;
          }
          .empty-state p {
            color: #666;
            margin-bottom: 30px;
            max-width: 500px;
            margin-left: auto;
            margin-right: auto;
            line-height: 1.6;
          }
          .cta-button {
            background: #3498db;
            color: white;
            border: none;
            padding: 15px 30px;
            border-radius: 8px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
          }
          .cta-button:hover {
            background: #2980b9;
          }
        `;
        document.head.appendChild(style);
      });
      
      await waitForElementStable(page, '.empty-state');
      await expect(page).toHaveScreenshot('state-empty-data.png');
    });

    test('enhanced risk indicators - empty indicators state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Simulate empty indicators state
      await page.evaluate(() => {
        const riskCards = document.querySelectorAll('.risk-card');
        riskCards.forEach(card => {
          card.classList.add('empty-state');
          card.innerHTML = `
            <div class="empty-indicator">
              <div class="empty-icon">📈</div>
              <h4>No Data Available</h4>
              <p>Risk indicators will appear here once data is available.</p>
            </div>
          `;
        });
        
        // Add empty indicator styles
        const style = document.createElement('style');
        style.textContent = `
          .empty-indicator {
            text-align: center;
            padding: 40px 20px;
            background: #f8f9fa;
            border: 2px dashed #dee2e6;
            border-radius: 8px;
            color: #6c757d;
          }
          .empty-indicator .empty-icon {
            font-size: 32px;
            margin-bottom: 15px;
            opacity: 0.5;
          }
          .empty-indicator h4 {
            margin-bottom: 10px;
            color: #495057;
            font-size: 16px;
          }
          .empty-indicator p {
            margin: 0;
            font-size: 14px;
            line-height: 1.4;
          }
        `;
        document.head.appendChild(style);
      });
      
      await waitForElementStable(page, '.empty-indicator');
      await expect(page).toHaveScreenshot('state-empty-indicators.png');
    });

    test('charts empty state', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Simulate empty charts state
      await page.evaluate(() => {
        const chartContainers = document.querySelectorAll('.chart-container');
        chartContainers.forEach(container => {
          container.innerHTML = `
            <div class="empty-chart">
              <div class="empty-chart-icon">📊</div>
              <h4>No Chart Data</h4>
              <p>Chart will display here when data is available.</p>
            </div>
          `;
        });
        
        // Add empty chart styles
        const style = document.createElement('style');
        style.textContent = `
          .empty-chart {
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            height: 300px;
            background: #f8f9fa;
            border: 2px dashed #dee2e6;
            border-radius: 8px;
            color: #6c757d;
          }
          .empty-chart-icon {
            font-size: 48px;
            margin-bottom: 15px;
            opacity: 0.5;
          }
          .empty-chart h4 {
            margin-bottom: 10px;
            color: #495057;
            font-size: 18px;
          }
          .empty-chart p {
            margin: 0;
            font-size: 14px;
            text-align: center;
          }
        `;
        document.head.appendChild(style);
      });
      
      await waitForElementStable(page, '.empty-chart');
      await expect(page).toHaveScreenshot('state-empty-charts.png');
    });
  });

  // Responsive State Tests
  test.describe('Responsive State Tests', () => {
    test('risk level states - mobile viewport', async ({ page }) => {
      await setViewportSize(page, 'mobile');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      const riskLevels = ['low', 'medium', 'high', 'critical'];
      
      for (const level of riskLevels) {
        await setRiskState(page, level);
        await waitForElementStable(page, '.risk-level-indicator');
        
        // Take mobile screenshot for each risk level
        await expect(page).toHaveScreenshot(`state-mobile-risk-${level}.png`);
      }
    });

    test('loading states - tablet viewport', async ({ page }) => {
      await setViewportSize(page, 'tablet');
      await navigateToDashboard(page, 'risk-dashboard');
      
      // Simulate loading state on tablet
      await page.evaluate(() => {
        const loadingOverlay = document.createElement('div');
        loadingOverlay.id = 'loading-overlay';
        loadingOverlay.className = 'loading-overlay';
        loadingOverlay.innerHTML = `
          <div class="loading-spinner">
            <div class="spinner"></div>
            <p>Loading risk assessment...</p>
          </div>
        `;
        document.body.appendChild(loadingOverlay);
        
        const style = document.createElement('style');
        style.textContent = `
          .loading-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(255, 255, 255, 0.9);
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 9999;
          }
          .loading-spinner {
            text-align: center;
          }
          .spinner {
            width: 40px;
            height: 40px;
            border: 4px solid #f3f3f3;
            border-top: 4px solid #3498db;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin: 0 auto 20px;
          }
          @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
          }
        `;
        document.head.appendChild(style);
      });
      
      await waitForElementStable(page, '#loading-overlay');
      await expect(page).toHaveScreenshot('state-tablet-loading.png');
    });

    test('error states - desktop viewport', async ({ page }) => {
      await setViewportSize(page, 'desktop');
      await navigateToDashboard(page, 'enhanced-risk-indicators');
      
      // Simulate error state on desktop
      await page.evaluate(() => {
        const errorOverlay = document.createElement('div');
        errorOverlay.id = 'error-overlay';
        errorOverlay.className = 'error-overlay';
        errorOverlay.innerHTML = `
          <div class="error-message">
            <div class="error-icon">⚠️</div>
            <h3>Unable to Load Risk Assessment</h3>
            <p>There was an error loading your risk assessment data. Please try again.</p>
            <button class="retry-button">Retry</button>
          </div>
        `;
        document.body.appendChild(errorOverlay);
        
        const style = document.createElement('style');
        style.textContent = `
          .error-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(255, 255, 255, 0.95);
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 9999;
          }
          .error-message {
            text-align: center;
            padding: 40px;
            background: white;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
            max-width: 400px;
          }
          .error-icon {
            font-size: 48px;
            margin-bottom: 20px;
          }
          .error-message h3 {
            color: #e74c3c;
            margin-bottom: 15px;
          }
          .error-message p {
            color: #666;
            margin-bottom: 25px;
            line-height: 1.5;
          }
          .retry-button {
            background: #3498db;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 16px;
          }
          .retry-button:hover {
            background: #2980b9;
          }
        `;
        document.head.appendChild(style);
      });
      
      await waitForElementStable(page, '#error-overlay');
      await expect(page).toHaveScreenshot('state-desktop-error.png');
    });
  });
});
