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

// Mock localStorage with actual storage implementation for tests
// This allows EnrichmentContext and other components to use localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value.toString();
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
    get length() {
      return Object.keys(store).length;
    },
    key: (index: number) => {
      const keys = Object.keys(store);
      return keys[index] || null;
    },
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true,
  configurable: true,
});
global.localStorage = localStorageMock as unknown as Storage;

// Clear localStorage before each test (handled in beforeEach in test files)

// Fix Radix UI hasPointerCapture issue in jsdom
// Radix UI components use hasPointerCapture which isn't available in jsdom
if (typeof Element !== 'undefined' && !Element.prototype.hasPointerCapture) {
  Object.defineProperty(Element.prototype, 'hasPointerCapture', {
    value: vi.fn().mockReturnValue(false),
    writable: true,
    configurable: true,
  });
  
  Object.defineProperty(Element.prototype, 'setPointerCapture', {
    value: vi.fn(),
    writable: true,
    configurable: true,
  });
  
  Object.defineProperty(Element.prototype, 'releasePointerCapture', {
    value: vi.fn(),
    writable: true,
    configurable: true,
  });
}

// Fix SVG transform issues for D3 in jsdom
// D3 transitions require SVG transform parsing which jsdom doesn't fully support
if (typeof SVGElement !== 'undefined') {
  // Mock SVGElement.transform for D3 compatibility
  const originalGetAttribute = SVGElement.prototype.getAttribute;
  SVGElement.prototype.getAttribute = function(name: string) {
    // Check if transform attribute exists before calling getAttribute to avoid recursion
    const hasTransform = originalGetAttribute.call(this, 'transform');
    if (name === 'transform' && !hasTransform) {
      // Return empty transform if not set to avoid D3 errors
      return '';
    }
    return originalGetAttribute.call(this, name);
  };
}

// Mock ResizeObserver for components that use it
// Must be a proper constructor class
class ResizeObserverMock {
  observe = vi.fn();
  unobserve = vi.fn();
  disconnect = vi.fn();
}
global.ResizeObserver = ResizeObserverMock as unknown as typeof ResizeObserver;

// Mock IntersectionObserver for components that use it
// Must be a proper constructor class
class IntersectionObserverMock {
  observe = vi.fn();
  unobserve = vi.fn();
  disconnect = vi.fn();
  root = null;
  rootMargin = '';
  thresholds: number[] = [];
}
global.IntersectionObserver = IntersectionObserverMock as unknown as typeof IntersectionObserver;

// Mock D3 transitions to prevent errors in test environment
// D3 transitions use browser APIs that jsdom doesn't fully support
vi.mock('d3-transition', () => {
  const mockTransition = () => ({
    attr: vi.fn().mockReturnThis(),
    attrTween: vi.fn().mockReturnThis(),
    style: vi.fn().mockReturnThis(),
    styleTween: vi.fn().mockReturnThis(),
    text: vi.fn().mockReturnThis(),
    textTween: vi.fn().mockReturnThis(),
    remove: vi.fn().mockReturnThis(),
    duration: vi.fn().mockReturnThis(),
    delay: vi.fn().mockReturnThis(),
    ease: vi.fn().mockReturnThis(),
    on: vi.fn().mockReturnThis(),
    end: vi.fn().mockResolvedValue(undefined),
    selection: vi.fn().mockReturnThis(),
  });

  return {
    transition: vi.fn().mockImplementation(mockTransition),
    active: vi.fn().mockReturnValue(null),
    interrupt: vi.fn(),
  };
});

// Mock d3-interpolate to prevent SVG transform parsing errors
vi.mock('d3-interpolate', () => {
  const originalModule = vi.importActual('d3-interpolate');
  return {
    ...originalModule,
    interpolateTransformSvg: vi.fn().mockReturnValue(() => 'translate(0,0)'),
  };
});

// Suppress SelectContentImpl errors from Radix UI Select
// These occur because SelectContent renders in a portal which may not be fully set up in tests
const originalError = console.error;
console.error = (...args: any[]) => {
  const message = args[0]?.toString() || '';
  // Suppress SelectContentImpl errors
  if (
    message.includes('SelectContentImpl') ||
    message.includes('An error occurred in the <SelectContentImpl> component')
  ) {
    return;
  }
  originalError(...args);
};

