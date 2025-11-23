/**
 * API Mock Helpers for Playwright E2E Tests
 * Provides route interception to mock API responses without requiring a backend
 */

const testData = require('../fixtures/test-data.json');

/**
 * Generate mock merchant data
 */
function generateMockMerchants(count = 20) {
  const merchants = [];
  const industries = ['Technology', 'Retail', 'Manufacturing', 'Services', 'Healthcare'];
  const portfolioTypes = ['onboarded', 'prospective', 'pending', 'deactivated'];
  const riskLevels = ['low', 'medium', 'high', 'critical'];
  
  for (let i = 1; i <= count; i++) {
    merchants.push({
      id: `merchant-${String(i).padStart(3, '0')}`,
      name: `Test Merchant ${i}`,
      industry: industries[i % industries.length],
      portfolio_type: portfolioTypes[i % portfolioTypes.length],
      risk_level: riskLevels[i % riskLevels.length],
      risk_score: [25, 50, 75, 95][i % 4],
      address: `${i} Test Street, Test City, ST 12345`,
      phone: `+1-555-${String(i).padStart(3, '0')}-${String(i).padStart(4, '0')}`,
      email: `merchant${i}@test.com`,
      website: `https://merchant${i}.test.com`,
      onboarding_date: new Date(Date.now() - i * 86400000).toISOString(),
      created_at: new Date(Date.now() - i * 86400000).toISOString(),
      updated_at: new Date(Date.now() - i * 3600000).toISOString(),
    });
  }
  
  return merchants;
}

/**
 * Setup API route mocking for all merchant-related endpoints
 * @param {import('@playwright/test').Page} page - Playwright page object
 */
async function setupAPIMocks(page) {
  const mockMerchants = generateMockMerchants(20);
  
  // Mock GET /api/v1/merchants - List merchants
  await page.route('**/api/v1/merchants**', async (route) => {
    const url = new URL(route.request().url());
    const pageNum = parseInt(url.searchParams.get('page') || '1');
    const pageSize = parseInt(url.searchParams.get('page_size') || '20');
    const start = (pageNum - 1) * pageSize;
    const end = start + pageSize;
    
    const paginatedMerchants = mockMerchants.slice(start, end);
    
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        merchants: paginatedMerchants,
        pagination: {
          page: pageNum,
          page_size: pageSize,
          total: mockMerchants.length,
          total_pages: Math.ceil(mockMerchants.length / pageSize)
        }
      })
    });
  });
  
  // Mock GET /api/v1/merchants/:id - Get single merchant
  await page.route('**/api/v1/merchants/*', async (route) => {
    const url = route.request().url();
    const merchantId = url.match(/\/api\/v1\/merchants\/([^/?]+)/)?.[1];
    
    if (merchantId && merchantId !== 'search' && merchantId !== 'analytics' && 
        merchantId !== 'statistics' && merchantId !== 'portfolio-types' && 
        merchantId !== 'risk-levels' && merchantId !== 'counts') {
      const merchant = mockMerchants.find(m => m.id === merchantId) || mockMerchants[0];
      
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(merchant)
      });
    } else {
      await route.continue();
    }
  });
  
  // Mock GET /api/v1/merchants/search
  await page.route('**/api/v1/merchants/search**', async (route) => {
    const url = new URL(route.request().url());
    const query = url.searchParams.get('q') || '';
    
    const filtered = mockMerchants.filter(m => 
      m.name.toLowerCase().includes(query.toLowerCase()) ||
      m.industry.toLowerCase().includes(query.toLowerCase())
    );
    
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        merchants: filtered,
        total: filtered.length
      })
    });
  });
  
  // Mock GET /api/v1/merchants/analytics
  await page.route('**/api/v1/merchants/analytics**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_merchants: mockMerchants.length,
        by_portfolio_type: {
          onboarded: mockMerchants.filter(m => m.portfolio_type === 'onboarded').length,
          prospective: mockMerchants.filter(m => m.portfolio_type === 'prospective').length,
          pending: mockMerchants.filter(m => m.portfolio_type === 'pending').length,
          deactivated: mockMerchants.filter(m => m.portfolio_type === 'deactivated').length
        },
        by_risk_level: {
          low: mockMerchants.filter(m => m.risk_level === 'low').length,
          medium: mockMerchants.filter(m => m.risk_level === 'medium').length,
          high: mockMerchants.filter(m => m.risk_level === 'high').length,
          critical: mockMerchants.filter(m => m.risk_level === 'critical').length
        },
        average_risk_score: mockMerchants.reduce((sum, m) => sum + m.risk_score, 0) / mockMerchants.length
      })
    });
  });
  
  // Mock GET /api/v1/merchants/statistics
  await page.route('**/api/v1/merchants/statistics**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total: mockMerchants.length,
        active: mockMerchants.filter(m => m.portfolio_type !== 'deactivated').length,
        inactive: mockMerchants.filter(m => m.portfolio_type === 'deactivated').length
      })
    });
  });
  
  // Mock GET /api/v1/merchants/portfolio-types
  await page.route('**/api/v1/merchants/portfolio-types**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        { type: 'onboarded', count: mockMerchants.filter(m => m.portfolio_type === 'onboarded').length },
        { type: 'prospective', count: mockMerchants.filter(m => m.portfolio_type === 'prospective').length },
        { type: 'pending', count: mockMerchants.filter(m => m.portfolio_type === 'pending').length },
        { type: 'deactivated', count: mockMerchants.filter(m => m.portfolio_type === 'deactivated').length }
      ])
    });
  });
  
  // Mock GET /api/v1/merchants/risk-levels
  await page.route('**/api/v1/merchants/risk-levels**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        { level: 'low', count: mockMerchants.filter(m => m.risk_level === 'low').length },
        { level: 'medium', count: mockMerchants.filter(m => m.risk_level === 'medium').length },
        { level: 'high', count: mockMerchants.filter(m => m.risk_level === 'high').length },
        { level: 'critical', count: mockMerchants.filter(m => m.risk_level === 'critical').length }
      ])
    });
  });
  
  // Mock GET /api/v1/merchants/counts
  await page.route('**/api/v1/merchants/counts**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total: mockMerchants.length,
        by_portfolio_type: {
          onboarded: mockMerchants.filter(m => m.portfolio_type === 'onboarded').length,
          prospective: mockMerchants.filter(m => m.portfolio_type === 'prospective').length,
          pending: mockMerchants.filter(m => m.portfolio_type === 'pending').length,
          deactivated: mockMerchants.filter(m => m.portfolio_type === 'deactivated').length
        }
      })
    });
  });
  
  // Mock merchant detail endpoints
  await page.route('**/api/v1/merchants/*/risk-indicators**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        merchant_id: route.request().url().match(/\/merchants\/([^/]+)/)?.[1] || 'merchant-001',
        risk_score: 50,
        risk_level: 'medium',
        indicators: [
          { type: 'financial', score: 45, status: 'warning' },
          { type: 'operational', score: 35, status: 'ok' },
          { type: 'regulatory', score: 55, status: 'warning' }
        ]
      })
    });
  });
  
  await page.route('**/api/v1/merchants/*/website-analysis**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        merchant_id: route.request().url().match(/\/merchants\/([^/]+)/)?.[1] || 'merchant-001',
        analysis_date: new Date().toISOString(),
        status: 'completed',
        findings: []
      })
    });
  });
  
  console.log('✅ API mocks configured');
}

/**
 * Setup API mocks in beforeEach hook
 * Use this in test files: test.beforeEach(async ({ page }) => { await setupAPIMocks(page); });
 */
module.exports = {
  setupAPIMocks,
  generateMockMerchants
};

