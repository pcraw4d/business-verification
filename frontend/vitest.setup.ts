// Vitest setup file
// Per MSW FAQ: Vitest has none of the Node.js globals issues and provides native ESM support

import '@testing-library/jest-dom';
import { beforeAll, beforeEach, afterEach, afterAll, vi } from 'vitest';
import { server } from './__tests__/mocks/server';

// Setup MSW (Mock Service Worker) for API mocking
// MSW v2 setup per documentation: https://mswjs.io/docs/integrations/node
beforeAll(() => {
  server.listen({
    onUnhandledRequest: (request) => {
      console.warn('[MSW] Unhandled request:', request.method, request.url);
    },
  });
});

// Reset handlers and clear cache BEFORE each test to ensure isolation
beforeEach(async () => {
  // Reset MSW handlers to default state before each test
  // This ensures that test-specific handlers (like error responses) don't persist
  server.resetHandlers();
  
  // Clear API cache and request deduplicator BEFORE each test
  // This prevents test isolation issues where cached responses or pending requests
  // from one test affect another test
  // CRITICAL: Clear these synchronously if possible, or ensure they're cleared before any test runs
  try {
    // Use dynamic import for ES modules compatibility
    const apiModule = await import('@/lib/api');
    const apiCache = apiModule.getAPICache();
    const requestDeduplicator = apiModule.getRequestDeduplicator();
    
    if (apiCache && typeof apiCache.clear === 'function') {
      apiCache.clear();
      // Also clear any sessionStorage cache entries
      if (typeof window !== 'undefined' && window.sessionStorage) {
        const keys = Object.keys(window.sessionStorage);
        keys.forEach(key => {
          if (key.startsWith('cache:')) {
            window.sessionStorage.removeItem(key);
          }
        });
      }
    }
    if (requestDeduplicator && typeof requestDeduplicator.clear === 'function') {
      requestDeduplicator.clear();
    }
  } catch (error) {
    // Ignore errors if module not loaded yet or during cleanup
    // This is expected in some test scenarios
  }
});

// Also reset handlers after each test as a safety measure
afterEach(() => {
  server.resetHandlers();
});

// Clean up after the tests are finished
afterAll(() => server.close());

// Mock Next.js router
vi.mock('next/navigation', () => ({
  useRouter() {
    return {
      push: vi.fn(),
      replace: vi.fn(),
      prefetch: vi.fn(),
      back: vi.fn(),
      pathname: '/',
      query: {},
      asPath: '/',
    };
  },
  usePathname() {
    return '/';
  },
  useSearchParams() {
    return new URLSearchParams();
  },
}));

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock sessionStorage
const sessionStorageMock = {
  getItem: vi.fn().mockReturnValue(null),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
Object.defineProperty(window, 'sessionStorage', {
  value: sessionStorageMock,
  writable: true,
});
global.sessionStorage = sessionStorageMock as unknown as Storage;

