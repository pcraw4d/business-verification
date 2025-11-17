import { http, HttpResponse } from 'msw';
import { vi } from 'vitest';

/**
 * Common mock data factories
 */
export const mockData = {
  merchant: {
    id: 'merchant-123',
    businessName: 'Test Business',
    industry: 'Technology',
    status: 'active',
    email: 'test@example.com',
    phone: '+1-555-123-4567',
    website: 'https://test.com',
    address: {
      street: '123 Main St',
      city: 'San Francisco',
      state: 'CA',
      postalCode: '94102',
      country: 'USA',
    },
    registrationNumber: 'REG-123',
    taxId: 'TAX-456',
    foundedYear: 2020,
    employeeCount: 50,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  
  merchants: [
    {
      id: 'merchant-1',
      businessName: 'Business 1',
      industry: 'Technology',
      status: 'active',
    },
    {
      id: 'merchant-2',
      businessName: 'Business 2',
      industry: 'Finance',
      status: 'pending',
    },
  ],
  
  riskAssessment: {
    merchantId: 'merchant-123',
    riskScore: 0.3,
    riskLevel: 'low',
    factors: [
      { name: 'Factor 1', score: 0.2, description: 'Description 1' },
      { name: 'Factor 2', score: 0.1, description: 'Description 2' },
    ],
  },
  
  analytics: {
    merchantId: 'merchant-123',
    classification: {
      primaryIndustry: 'Technology',
      confidenceScore: 0.95,
      riskLevel: 'low',
    },
    security: {
      trustScore: 0.8,
      sslValid: true,
    },
  },
  
  dashboardMetrics: {
    overview: {
      totalMerchants: 100,
      activeMerchants: 80,
      pendingVerifications: 10,
      riskAlerts: 5,
    },
    performance: {
      averageRiskScore: 0.35,
      complianceRate: 0.95,
      verificationSuccessRate: 0.88,
    },
    business: {
      totalRevenue: 1000000,
      monthlyGrowth: 0.05,
      topIndustries: ['Technology', 'Finance', 'Retail'],
    },
  },
  
  complianceStatus: {
    overallScore: 0.95,
    status: 'compliant',
    frameworks: [
      { name: 'PCI DSS', status: 'compliant', score: 0.98 },
      { name: 'GDPR', status: 'compliant', score: 0.95 },
    ],
    requirements: [
      { id: 'req-1', name: 'Requirement 1', status: 'met', lastChecked: new Date().toISOString() },
    ],
    alerts: [],
  },
  
  sessions: [
    {
      id: 'session-1',
      userId: 'user-1',
      ipAddress: '192.168.1.1',
      userAgent: 'Mozilla/5.0',
      createdAt: new Date().toISOString(),
      lastActivity: new Date().toISOString(),
      isActive: true,
      requestCount: 100,
    },
  ],
  
  enrichmentSources: [
    { id: 'source-1', name: 'Source 1', description: 'Description 1' },
    { id: 'source-2', name: 'Source 2', description: 'Description 2' },
  ],
};

/**
 * MSW handlers for common API endpoints
 */
export const commonHandlers = [
  // Merchants
  http.get('*/api/v1/merchants/:merchantId', ({ params }) => {
    return HttpResponse.json(mockData.merchant);
  }),
  
  http.get('*/api/v1/merchants', () => {
    return HttpResponse.json({ merchants: mockData.merchants });
  }),
  
  http.post('*/api/v1/merchants', () => {
    return HttpResponse.json({ id: 'merchant-new', ...mockData.merchant }, { status: 201 });
  }),
  
  // Risk Assessment
  http.get('*/api/v1/merchants/:merchantId/risk-score', () => {
    return HttpResponse.json(mockData.riskAssessment);
  }),
  
  http.get('*/api/v1/risk/indicators/:merchantId', () => {
    return HttpResponse.json({ merchantId: 'merchant-123', indicators: [] });
  }),
  
  // Analytics
  http.get('*/api/v1/merchants/:merchantId/analytics', () => {
    return HttpResponse.json(mockData.analytics);
  }),
  
  // Dashboard Metrics
  http.get('*/api/v3/dashboard/metrics', () => {
    return HttpResponse.json(mockData.dashboardMetrics);
  }),
  
  // Compliance
  http.get('*/api/v1/compliance/status', () => {
    return HttpResponse.json(mockData.complianceStatus);
  }),
  
  // Sessions
  http.get('*/api/v1/sessions', () => {
    return HttpResponse.json(mockData.sessions);
  }),
  
  // Enrichment
  http.get('*/api/v1/merchants/:merchantId/enrichment/sources', () => {
    return HttpResponse.json({ sources: mockData.enrichmentSources });
  }),
  
  http.post('*/api/v1/merchants/:merchantId/enrichment/trigger', () => {
    return HttpResponse.json({ jobId: 'job-123', status: 'pending' });
  }),
];

/**
 * Mock functions for common operations
 */
export const mockFunctions = {
  router: {
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
    back: vi.fn(),
    pathname: '/',
    query: {},
    asPath: '/',
  },
  
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
    warning: vi.fn(),
    promise: vi.fn(),
  },
  
  fetch: vi.fn(),
  
  localStorage: {
    getItem: vi.fn(),
    setItem: vi.fn(),
    removeItem: vi.fn(),
    clear: vi.fn(),
  },
  
  sessionStorage: {
    getItem: vi.fn(),
    setItem: vi.fn(),
    removeItem: vi.fn(),
    clear: vi.fn(),
  },
};

/**
 * Reset all mocks
 */
export function resetAllMocks() {
  vi.clearAllMocks();
  mockFunctions.router.push.mockClear();
  mockFunctions.router.replace.mockClear();
  mockFunctions.toast.success.mockClear();
  mockFunctions.toast.error.mockClear();
  mockFunctions.toast.info.mockClear();
  mockFunctions.toast.warning.mockClear();
  mockFunctions.fetch.mockClear();
}

/**
 * Create error response handler
 */
export function createErrorHandler(status: number, message: string) {
  return http.all('*', () => {
    return HttpResponse.json({ error: message }, { status });
  });
}

/**
 * Create delay handler (for testing loading states)
 */
export function createDelayHandler(delay: number) {
  return http.all('*', async () => {
    await new Promise((resolve) => setTimeout(resolve, delay));
    return HttpResponse.json({});
  });
}

