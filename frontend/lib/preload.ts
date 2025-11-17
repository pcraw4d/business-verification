/**
 * Resource preloading utilities for performance optimization
 */

/**
 * Preload a resource (script, style, font, etc.)
 */
export function preloadResource(href: string, as: string, crossorigin?: string) {
  if (typeof window === 'undefined') return;

  const link = document.createElement('link');
  link.rel = 'preload';
  link.href = href;
  link.as = as;
  if (crossorigin) {
    link.crossOrigin = crossorigin;
  }

  document.head.appendChild(link);
}

/**
 * Prefetch a resource (for likely future navigation)
 */
export function prefetchResource(href: string, as: string) {
  if (typeof window === 'undefined') return;

  const link = document.createElement('link');
  link.rel = 'prefetch';
  link.href = href;
  link.as = as;

  document.head.appendChild(link);
}

/**
 * Preconnect to an origin (DNS lookup, TCP handshake, TLS negotiation)
 */
export function preconnectToOrigin(origin: string, crossorigin?: string) {
  if (typeof window === 'undefined') return;

  const link = document.createElement('link');
  link.rel = 'preconnect';
  link.href = origin;
  if (crossorigin) {
    link.crossOrigin = crossorigin;
  }

  document.head.appendChild(link);
}

/**
 * DNS prefetch (DNS lookup only)
 */
export function dnsPrefetch(origin: string) {
  if (typeof window === 'undefined') return;

  const link = document.createElement('link');
  link.rel = 'dns-prefetch';
  link.href = origin;

  document.head.appendChild(link);
}

/**
 * Initialize performance optimizations
 */
export function initPerformanceOptimizations() {
  if (typeof window === 'undefined') return;

  const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
  const apiOrigin = new URL(apiBaseUrl).origin;

  // Preconnect to API origin
  preconnectToOrigin(apiOrigin);

  // Prefetch likely navigation targets
  const likelyRoutes = [
    '/dashboard',
    '/merchant-portfolio',
    '/risk-dashboard',
    '/compliance',
  ];

  // Prefetch routes on idle
  if ('requestIdleCallback' in window) {
    requestIdleCallback(() => {
      likelyRoutes.forEach((route) => {
        prefetchResource(route, 'document');
      });
    });
  } else {
    // Fallback for browsers without requestIdleCallback
    setTimeout(() => {
      likelyRoutes.forEach((route) => {
        prefetchResource(route, 'document');
      });
    }, 2000);
  }
}

