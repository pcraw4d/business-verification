// Learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom';

// CRITICAL: Load polyfills FIRST before any MSW imports
// This ensures Response, Request, etc. are available when MSW modules are loaded
require('./__tests__/mocks/polyfills.js')

// Setup MSW (Mock Service Worker) for API mocking
// Import AFTER polyfills are loaded
import { server } from './__tests__/mocks/server';

// Establish API mocking before all tests
// MSW v2 setup per documentation: https://mswjs.io/docs/integrations/node
beforeAll(() => {
  server.listen({
    onUnhandledRequest: (request) => {
      console.warn('[MSW] Unhandled request:', request.method, request.url);
    },
  });
})

// Reset any request handlers that are declared as a part of our tests
// (i.e. for testing one-time error scenarios)
afterEach(() => server.resetHandlers())

// Clean up after the tests are finished
afterAll(() => server.close())

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter() {
    return {
      push: jest.fn(),
      replace: jest.fn(),
      prefetch: jest.fn(),
      back: jest.fn(),
      pathname: '/',
      query: {},
      asPath: '/',
    }
  },
  usePathname() {
    return '/'
  },
  useSearchParams() {
    return new URLSearchParams()
  },
}))

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(), // deprecated
    removeListener: jest.fn(), // deprecated
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
})

// Mock sessionStorage
const sessionStorageMock = {
  getItem: jest.fn().mockReturnValue(null),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
}
Object.defineProperty(window, 'sessionStorage', {
  value: sessionStorageMock,
  writable: true,
})
global.sessionStorage = sessionStorageMock

