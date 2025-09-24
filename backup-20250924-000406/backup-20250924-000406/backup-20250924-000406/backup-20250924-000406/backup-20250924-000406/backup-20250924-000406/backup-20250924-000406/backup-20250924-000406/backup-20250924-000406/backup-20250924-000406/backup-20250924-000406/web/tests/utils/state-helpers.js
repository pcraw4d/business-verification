// web/tests/utils/state-helpers.js
/**
 * State management utilities for visual regression testing
 * Provides helper functions for testing different application states
 */

/**
 * Set loading state for the page
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} type - Loading type (overlay, skeleton, spinner)
 * @param {Object} options - Loading options
 */
async function setLoadingState(page, type = 'overlay', options = {}) {
  const defaultOptions = {
    message: 'Loading...',
    showSpinner: true,
    overlay: true,
    ...options
  };

  await page.evaluate(({ type, options }) => {
    // Remove existing loading states
    const existingLoading = document.querySelectorAll('.loading-overlay, .loading-skeleton, .loading-spinner');
    existingLoading.forEach(el => el.remove());

    let loadingElement;

    switch (type) {
      case 'overlay':
        loadingElement = createLoadingOverlay(options);
        break;
      case 'skeleton':
        loadingElement = createLoadingSkeleton(options);
        break;
      case 'spinner':
        loadingElement = createLoadingSpinner(options);
        break;
      default:
        loadingElement = createLoadingOverlay(options);
    }

    document.body.appendChild(loadingElement);
    addLoadingStyles();
  }, { type, options: defaultOptions });

  console.log(`🔄 Set loading state: ${type}`);
}

/**
 * Set error state for the page
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} type - Error type (overlay, card, form)
 * @param {Object} options - Error options
 */
async function setErrorState(page, type = 'overlay', options = {}) {
  const defaultOptions = {
    title: 'Error',
    message: 'An error occurred',
    showRetry: true,
    ...options
  };

  await page.evaluate(({ type, options }) => {
    // Remove existing error states
    const existingErrors = document.querySelectorAll('.error-overlay, .error-card, .error-message');
    existingErrors.forEach(el => el.remove());

    let errorElement;

    switch (type) {
      case 'overlay':
        errorElement = createErrorOverlay(options);
        break;
      case 'card':
        errorElement = createErrorCard(options);
        break;
      case 'form':
        errorElement = createFormErrors(options);
        break;
      default:
        errorElement = createErrorOverlay(options);
    }

    if (type === 'form') {
      const form = document.querySelector('#riskAssessmentForm');
      if (form) {
        form.insertBefore(errorElement, form.firstChild);
      }
    } else {
      document.body.appendChild(errorElement);
    }

    addErrorStyles();
  }, { type, options: defaultOptions });

  console.log(`❌ Set error state: ${type}`);
}

/**
 * Set empty data state for the page
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {string} type - Empty state type (dashboard, cards, charts)
 * @param {Object} options - Empty state options
 */
async function setEmptyState(page, type = 'dashboard', options = {}) {
  const defaultOptions = {
    title: 'No Data Available',
    message: 'No data to display at this time',
    showCTA: true,
    ctaText: 'Get Started',
    ...options
  };

  await page.evaluate(({ type, options }) => {
    // Remove existing empty states
    const existingEmpty = document.querySelectorAll('.empty-state, .empty-card, .empty-chart');
    existingEmpty.forEach(el => el.remove());

    let emptyElement;

    switch (type) {
      case 'dashboard':
        emptyElement = createEmptyDashboard(options);
        break;
      case 'cards':
        emptyElement = createEmptyCards(options);
        break;
      case 'charts':
        emptyElement = createEmptyCharts(options);
        break;
      default:
        emptyElement = createEmptyDashboard(options);
    }

    if (type === 'dashboard') {
      const mainContent = document.querySelector('#main-content');
      if (mainContent) {
        mainContent.innerHTML = '';
        mainContent.appendChild(emptyElement);
      }
    } else if (type === 'cards') {
      const cards = document.querySelectorAll('.risk-card');
      cards.forEach(card => {
        card.innerHTML = '';
        card.appendChild(emptyElement.cloneNode(true));
      });
    } else if (type === 'charts') {
      const chartContainers = document.querySelectorAll('.chart-container');
      chartContainers.forEach(container => {
        container.innerHTML = '';
        container.appendChild(emptyElement.cloneNode(true));
      });
    }

    addEmptyStateStyles();
  }, { type, options: defaultOptions });

  console.log(`📭 Set empty state: ${type}`);
}

/**
 * Set form validation state
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {Array} errors - Array of field errors
 */
async function setFormValidationState(page, errors = []) {
  await page.evaluate((errors) => {
    const form = document.querySelector('#riskAssessmentForm');
    if (!form) return;

    // Clear existing errors
    const existingErrors = form.querySelectorAll('.error, .error-message, .form-error-summary');
    existingErrors.forEach(el => el.remove());

    // Add field errors
    errors.forEach(error => {
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
    if (errors.length > 0) {
      const errorSummary = document.createElement('div');
      errorSummary.className = 'form-error-summary';
      errorSummary.innerHTML = `
        <h4>Please correct the following errors:</h4>
        <ul>
          ${errors.map(error => `<li>${error.message}</li>`).join('')}
        </ul>
      `;
      form.insertBefore(errorSummary, form.firstChild);
    }

    addFormErrorStyles();
  }, errors);

  console.log(`📝 Set form validation state with ${errors.length} errors`);
}

/**
 * Simulate network delay
 * @param {import('@playwright/test').Page} page - Playwright page object
 * @param {number} delay - Delay in milliseconds
 */
async function simulateNetworkDelay(page, delay = 2000) {
  await page.route('**/*', async route => {
    await new Promise(resolve => setTimeout(resolve, delay));
    await route.continue();
  });
  console.log(`⏱️ Simulated network delay: ${delay}ms`);
}

/**
 * Clear all custom states
 * @param {import('@playwright/test').Page} page - Playwright page object
 */
async function clearAllStates(page) {
  await page.evaluate(() => {
    // Remove all custom states
    const customElements = document.querySelectorAll(`
      .loading-overlay, .loading-skeleton, .loading-spinner,
      .error-overlay, .error-card, .error-message, .form-error-summary,
      .empty-state, .empty-card, .empty-chart
    `);
    customElements.forEach(el => el.remove());

    // Remove custom styles
    const customStyles = document.querySelectorAll('style[data-test-state]');
    customStyles.forEach(style => style.remove());

    // Clear form errors
    const form = document.querySelector('#riskAssessmentForm');
    if (form) {
      const errorFields = form.querySelectorAll('.error');
      errorFields.forEach(field => {
        field.classList.remove('error');
        field.removeAttribute('aria-invalid');
      });
    }
  });

  console.log('🧹 Cleared all custom states');
}

// Helper functions for creating DOM elements (used in page.evaluate)
const createLoadingOverlay = (options) => {
  const overlay = document.createElement('div');
  overlay.className = 'loading-overlay';
  overlay.innerHTML = `
    <div class="loading-spinner">
      ${options.showSpinner ? '<div class="spinner"></div>' : ''}
      <p>${options.message}</p>
    </div>
  `;
  return overlay;
};

const createLoadingSkeleton = (options) => {
  const skeleton = document.createElement('div');
  skeleton.className = 'loading-skeleton';
  skeleton.innerHTML = `
    <div class="skeleton-line"></div>
    <div class="skeleton-line short"></div>
    <div class="skeleton-circle"></div>
  `;
  return skeleton;
};

const createLoadingSpinner = (options) => {
  const spinner = document.createElement('div');
  spinner.className = 'loading-spinner';
  spinner.innerHTML = `
    <div class="spinner"></div>
    <p>${options.message}</p>
  `;
  return spinner;
};

const createErrorOverlay = (options) => {
  const overlay = document.createElement('div');
  overlay.className = 'error-overlay';
  overlay.innerHTML = `
    <div class="error-message">
      <div class="error-icon">⚠️</div>
      <h3>${options.title}</h3>
      <p>${options.message}</p>
      ${options.showRetry ? '<button class="retry-button">Retry</button>' : ''}
    </div>
  `;
  return overlay;
};

const createErrorCard = (options) => {
  const card = document.createElement('div');
  card.className = 'error-card';
  card.innerHTML = `
    <div class="error-icon">❌</div>
    <h4>${options.title}</h4>
    <p>${options.message}</p>
    ${options.showRetry ? '<button class="refresh-button">Refresh</button>' : ''}
  `;
  return card;
};

const createFormErrors = (options) => {
  const summary = document.createElement('div');
  summary.className = 'form-error-summary';
  summary.innerHTML = `
    <h4>Please correct the following errors:</h4>
    <ul>
      <li>Business name is required</li>
      <li>Industry type is required</li>
      <li>Email address is invalid</li>
    </ul>
  `;
  return summary;
};

const createEmptyDashboard = (options) => {
  const empty = document.createElement('div');
  empty.className = 'empty-state';
  empty.innerHTML = `
    <div class="empty-icon">📊</div>
    <h3>${options.title}</h3>
    <p>${options.message}</p>
    ${options.showCTA ? `<button class="cta-button">${options.ctaText}</button>` : ''}
  `;
  return empty;
};

const createEmptyCards = (options) => {
  const empty = document.createElement('div');
  empty.className = 'empty-indicator';
  empty.innerHTML = `
    <div class="empty-icon">📈</div>
    <h4>${options.title}</h4>
    <p>${options.message}</p>
  `;
  return empty;
};

const createEmptyCharts = (options) => {
  const empty = document.createElement('div');
  empty.className = 'empty-chart';
  empty.innerHTML = `
    <div class="empty-chart-icon">📊</div>
    <h4>${options.title}</h4>
    <p>${options.message}</p>
  `;
  return empty;
};

// Style injection functions
const addLoadingStyles = () => {
  const style = document.createElement('style');
  style.setAttribute('data-test-state', 'loading');
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
};

const addErrorStyles = () => {
  const style = document.createElement('style');
  style.setAttribute('data-test-state', 'error');
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
    .retry-button, .refresh-button {
      background: #3498db;
      color: white;
      border: none;
      padding: 12px 24px;
      border-radius: 6px;
      cursor: pointer;
      font-size: 16px;
    }
    .retry-button:hover, .refresh-button:hover {
      background: #2980b9;
    }
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
      font-size: 14px;
      padding: 8px 16px;
    }
    .refresh-button:hover {
      background: #c0392b;
    }
  `;
  document.head.appendChild(style);
};

const addEmptyStateStyles = () => {
  const style = document.createElement('style');
  style.setAttribute('data-test-state', 'empty');
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
};

const addFormErrorStyles = () => {
  const style = document.createElement('style');
  style.setAttribute('data-test-state', 'form-error');
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
};

module.exports = {
  setLoadingState,
  setErrorState,
  setEmptyState,
  setFormValidationState,
  simulateNetworkDelay,
  clearAllStates
};
