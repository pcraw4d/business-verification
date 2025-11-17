import { render, RenderOptions } from '@testing-library/react';
import { ReactElement } from 'react';
import { vi } from 'vitest';

/**
 * Custom render function with providers
 * Use this instead of the default render from @testing-library/react
 * when you need to wrap components with context providers
 */
export function renderWithProviders(
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) {
  return render(ui, {
    ...options,
  });
}

/**
 * Wait for async operations to complete
 * Useful for waiting for API calls, state updates, etc.
 */
export async function waitForAsync() {
  await new Promise((resolve) => setTimeout(resolve, 0));
}

/**
 * Mock Next.js router
 */
export function createMockRouter(overrides = {}) {
  return {
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
    back: vi.fn(),
    pathname: '/',
    query: {},
    asPath: '/',
    ...overrides,
  };
}

/**
 * Mock Next.js pathname
 */
export function createMockPathname(pathname = '/') {
  return pathname;
}

/**
 * Mock Next.js search params
 */
export function createMockSearchParams(params: Record<string, string> = {}) {
  return new URLSearchParams(params);
}

/**
 * Create mock merchant data
 */
export function createMockMerchant(overrides = {}) {
  return {
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
    ...overrides,
  };
}

/**
 * Create mock API response
 */
export function createMockResponse<T>(data: T, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? 'OK' : 'Error',
    json: async () => data,
    text: async () => JSON.stringify(data),
    headers: new Headers(),
  } as Response;
}

/**
 * Create mock fetch function
 */
export function createMockFetch(responses: Record<string, any> = {}) {
  return vi.fn((url: string | Request | URL) => {
    const urlString = typeof url === 'string' ? url : url instanceof URL ? url.toString() : url.url;
    const response = responses[urlString] || responses['*'] || { data: {}, status: 200 };
    return Promise.resolve(createMockResponse(response.data, response.status));
  });
}

/**
 * Wait for element to appear
 */
export async function waitForElement(selector: string, timeout = 5000) {
  const startTime = Date.now();
  while (Date.now() - startTime < timeout) {
    const element = document.querySelector(selector);
    if (element) return element;
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  throw new Error(`Element ${selector} not found within ${timeout}ms`);
}

/**
 * Mock toast notifications
 */
export function createMockToast() {
  return {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
    warning: vi.fn(),
    promise: vi.fn(),
  };
}

/**
 * Create mock form data
 */
export function createMockFormData(overrides = {}) {
  return {
    businessName: 'Test Business',
    websiteUrl: 'https://test.com',
    streetAddress: '123 Main St',
    city: 'San Francisco',
    state: 'CA',
    postalCode: '94102',
    country: 'US',
    phoneNumber: '+1-555-123-4567',
    email: 'test@example.com',
    registrationNumber: 'REG-123',
    analysisType: 'comprehensive',
    assessmentType: 'standard',
    ...overrides,
  };
}

/**
 * Create mock chart data
 */
export function createMockChartData(overrides = {}) {
  return {
    labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
    datasets: [
      {
        label: 'Dataset 1',
        data: [10, 20, 30, 40, 50, 60],
        backgroundColor: 'rgba(99, 102, 241, 0.5)',
        borderColor: 'rgba(99, 102, 241, 1)',
      },
    ],
    ...overrides,
  };
}

/**
 * Create mock risk data
 */
export function createMockRiskData(overrides = {}) {
  return {
    riskScore: 0.3,
    riskLevel: 'low',
    factors: [
      { name: 'Factor 1', score: 0.2, description: 'Description 1' },
      { name: 'Factor 2', score: 0.1, description: 'Description 2' },
    ],
    ...overrides,
  };
}

/**
 * Create mock dashboard metrics
 */
export function createMockDashboardMetrics(overrides = {}) {
  return {
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
    ...overrides,
  };
}

/**
 * Create mock compliance status
 */
export function createMockComplianceStatus(overrides = {}) {
  return {
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
    ...overrides,
  };
}

/**
 * Create mock session data
 */
export function createMockSession(overrides = {}) {
  return {
    id: 'session-123',
    userId: 'user-123',
    ipAddress: '192.168.1.1',
    userAgent: 'Mozilla/5.0',
    createdAt: new Date().toISOString(),
    lastActivity: new Date().toISOString(),
    isActive: true,
    requestCount: 100,
    ...overrides,
  };
}

/**
 * Helper to test form validation
 */
export function createFormValidationTest(
  formData: Record<string, any>,
  expectedErrors: Record<string, string>
) {
  return {
    formData,
    expectedErrors,
    validate: (errors: Record<string, string>) => {
      Object.keys(expectedErrors).forEach((field) => {
        expect(errors[field]).toBe(expectedErrors[field]);
      });
    },
  };
}

/**
 * Helper to test async operations
 */
export async function testAsyncOperation<T>(
  operation: () => Promise<T>,
  expectedResult?: T,
  timeout = 5000
): Promise<T> {
  const startTime = Date.now();
  const result = await Promise.race([
    operation(),
    new Promise<T>((_, reject) =>
      setTimeout(() => reject(new Error('Operation timeout')), timeout)
    ),
  ]);

  if (expectedResult !== undefined) {
    expect(result).toEqual(expectedResult);
  }

  return result;
}

/**
 * Mock window.matchMedia
 */
export function mockMatchMedia(matches = false) {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  });
}

/**
 * Mock IntersectionObserver
 */
export function mockIntersectionObserver() {
  global.IntersectionObserver = class IntersectionObserver {
    constructor() {}
    disconnect() {}
    observe() {}
    takeRecords() {
      return [];
    }
    unobserve() {}
  } as any;
}

/**
 * Helper to create test user event
 */
export function createTestUserEvent() {
  return {
    click: vi.fn(),
    type: vi.fn(),
    clear: vi.fn(),
    selectOptions: vi.fn(),
    upload: vi.fn(),
  };
}

