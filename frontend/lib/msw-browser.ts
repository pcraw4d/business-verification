/**
 * MSW Browser Integration
 * 
 * This module initializes Mock Service Worker in the browser for development/testing.
 * MSW intercepts fetch requests at the network level, allowing us to mock API responses
 * without modifying the application code.
 * 
 * Usage:
 * 1. Enable MSW in browser: Set NEXT_PUBLIC_MSW_ENABLED=true or localStorage.setItem('msw-enabled', 'true')
 * 2. Import this module in your app: import '@/lib/msw-browser'
 * 3. MSW will intercept all API calls matching the handlers
 * 
 * Note: MSW only works in browser (not SSR). For SSR testing, use the Node.js server setup.
 */

async function initMSW() {
  // Check if MSW should be enabled
  const isEnabled = 
    process.env.NEXT_PUBLIC_MSW_ENABLED === 'true' ||
    (typeof window !== 'undefined' && localStorage.getItem('msw-enabled') === 'true');

  if (!isEnabled) {
    if (process.env.NODE_ENV === 'development') {
      console.log('[MSW] Mock Service Worker is disabled. Enable with NEXT_PUBLIC_MSW_ENABLED=true or localStorage.setItem("msw-enabled", "true")');
    }
    return;
  }

  // Only run in browser (not SSR)
  if (typeof window === 'undefined') {
    return;
  }

  // Only run in development/test environment
  if (process.env.NODE_ENV === 'production') {
    console.warn('[MSW] Mock Service Worker should not be enabled in production');
    return;
  }

  try {
    // Dynamically import MSW browser setup
    const { setupWorker } = await import('msw/browser');
    const { handlers } = await import('@/__tests__/mocks/handlers');
    const { errorHandlers } = await import('@/__tests__/mocks/handlers-error-scenarios');
    const { testMerchantHandlers } = await import('@/__tests__/mocks/handlers-error-scenarios');

    // Combine all handlers
    const allHandlers = [
      ...handlers,
      ...errorHandlers,
      ...testMerchantHandlers,
    ];

    // Create worker
    const worker = setupWorker(...allHandlers);

    // Start worker
    await worker.start({
      onUnhandledRequest: (request, print) => {
        // Only warn about unhandled requests in development
        if (process.env.NODE_ENV === 'development') {
          // Skip Next.js internal requests
          if (request.url.includes('/_next/') || request.url.includes('/__webpack')) {
            return;
          }
          console.warn('[MSW] Unhandled request:', request.method, request.url);
        }
      },
    });

    console.log('[MSW] ✅ Mock Service Worker started in browser');
    console.log('[MSW] Handlers loaded:', allHandlers.length);
    console.log('[MSW] To disable: localStorage.setItem("msw-enabled", "false") or remove NEXT_PUBLIC_MSW_ENABLED');

    // Expose worker globally for debugging
    if (typeof window !== 'undefined') {
      (window as any).__MSW_WORKER__ = worker;
    }
  } catch (error) {
    console.error('[MSW] Failed to initialize Mock Service Worker:', error);
    console.error('[MSW] Make sure you have run: npx msw init public/');
  }
}

// Auto-initialize if in browser
if (typeof window !== 'undefined') {
  // Wait for DOM to be ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initMSW);
  } else {
    initMSW();
  }
}

export { initMSW };

