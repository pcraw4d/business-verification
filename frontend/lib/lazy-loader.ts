// Lazy Loading Utility using IntersectionObserver

type LoadFunction = () => Promise<void> | void;

interface LazyLoadOptions {
  rootMargin?: string;
  threshold?: number;
}

export class LazyLoader {
  private observer: IntersectionObserver | null;
  private loadedSections: Set<Element>;
  private loadFunctions: Map<Element, LoadFunction>;

  constructor(options: LazyLoadOptions = {}) {
    this.loadedSections = new Set();
    this.loadFunctions = new Map();

    const rootMargin = options.rootMargin || '50px';
    const threshold = options.threshold || 0.1;

    if (typeof window !== 'undefined' && 'IntersectionObserver' in window) {
      this.observer = new IntersectionObserver(
        (entries) => {
          entries.forEach((entry) => {
            if (entry.isIntersecting) {
              this.handleIntersection(entry.target);
            }
          });
        },
        {
          rootMargin,
          threshold,
        }
      );
    } else {
      this.observer = null;
    }
  }

  /**
   * Observes an element and loads data when it becomes visible
   */
  observe(element: Element, loadFn: LoadFunction): void {
    if (this.loadedSections.has(element)) {
      return; // Already loaded
    }

    // Store load function
    this.loadFunctions.set(element, loadFn);

    // Observe element
    if (this.observer) {
      this.observer.observe(element);
    } else {
      // Fallback: load immediately if IntersectionObserver not supported
      this.handleIntersection(element);
    }
  }

  /**
   * Handles intersection event
   */
  private handleIntersection(element: Element): void {
    if (this.loadedSections.has(element)) {
      return; // Already loaded
    }

    // Mark as loaded
    this.loadedSections.add(element);

    // Get and execute load function
    const loadFn = this.loadFunctions.get(element);
    if (loadFn) {
      Promise.resolve(loadFn()).catch((error) => {
        console.error('Lazy load error:', error);
        // Remove from loaded set so it can be retried
        this.loadedSections.delete(element);
      });
    }

    // Unobserve element
    if (this.observer) {
      this.observer.unobserve(element);
    }

    // Clean up
    this.loadFunctions.delete(element);
  }

  /**
   * Disconnects the observer
   */
  disconnect(): void {
    if (this.observer) {
      this.observer.disconnect();
    }
    this.loadedSections.clear();
    this.loadFunctions.clear();
  }
}

/**
 * Defer non-critical API calls using requestIdleCallback or setTimeout
 */
export function deferNonCriticalDataLoad(loadFn: () => void | Promise<void>): void {
  if (typeof window === 'undefined') {
    return;
  }

  if ('requestIdleCallback' in window) {
    window.requestIdleCallback(() => {
      Promise.resolve(loadFn()).catch(console.error);
    });
  } else {
    // Fallback: use setTimeout with 2 second delay
    setTimeout(() => {
      Promise.resolve(loadFn()).catch(console.error);
    }, 2000);
  }
}

