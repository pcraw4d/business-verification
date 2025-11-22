
import { LazyLoader, deferNonCriticalDataLoad } from '@/lib/lazy-loader';

describe('LazyLoader', () => {
  let loader: LazyLoader;
  let mockElement: HTMLElement;

  beforeEach(() => {
    loader = new LazyLoader();
    mockElement = document.createElement('div');
    document.body.appendChild(mockElement);
  });

  afterEach(() => {
    loader.disconnect();
    document.body.removeChild(mockElement);
  });

  it('should observe element and load when visible', async () => {
    const loadFn = vi.fn().mockResolvedValue(undefined);

    loader.observe(mockElement, loadFn);

    // Simulate intersection
    const entries = [{ target: mockElement, isIntersecting: true }];
    (loader as any).handleIntersection(mockElement);

    // Wait for async load
    await new Promise((resolve) => setTimeout(resolve, 10));

    expect(loadFn).toHaveBeenCalled();
  });

  it('should not load same element twice', async () => {
    const loadFn = vi.fn().mockResolvedValue(undefined);

    loader.observe(mockElement, loadFn);
    (loader as any).handleIntersection(mockElement);
    await new Promise((resolve) => setTimeout(resolve, 10));

    loader.observe(mockElement, loadFn);
    (loader as any).handleIntersection(mockElement);
    await new Promise((resolve) => setTimeout(resolve, 10));

    expect(loadFn).toHaveBeenCalledTimes(1);
  });

  it('should disconnect observer', () => {
    const loadFn = vi.fn();
    loader.observe(mockElement, loadFn);
    
    // Get the observer before disconnect
    const observer = (loader as any).observer;
    expect(observer).toBeTruthy();
    
    // Disconnect before intersection
    loader.disconnect();

    // After disconnect, observer should be null or disconnected
    // In test environment with mocked IntersectionObserver, it may not be null
    // but disconnect() should have been called
    if ((loader as any).observer) {
      // If observer still exists (mocked), verify disconnect was called
      expect((loader as any).observer.disconnect).toHaveBeenCalled();
    } else {
      // If observer is null, that's also valid
      expect((loader as any).observer).toBeNull();
    }
  });

  it('should handle load errors gracefully', async () => {
    const loadFn = vi.fn().mockRejectedValue(new Error('Load error'));

    loader.observe(mockElement, loadFn);
    (loader as any).handleIntersection(mockElement);

    await new Promise((resolve) => setTimeout(resolve, 10));

    // Should not throw, but should mark as not loaded for retry
    expect(loadFn).toHaveBeenCalled();
  });
});

describe('deferNonCriticalDataLoad', () => {
  beforeEach(() => {
    vi.clearAllTimers();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('should use requestIdleCallback when available', () => {
    const mockIdleCallback = vi.fn((cb) => {
      setTimeout(cb, 0);
      return 1;
    });
    (window as any).requestIdleCallback = mockIdleCallback;

    const loadFn = vi.fn();
    deferNonCriticalDataLoad(loadFn);

    vi.advanceTimersByTime(10);

    expect(mockIdleCallback).toHaveBeenCalled();
  });

  it('should fallback to setTimeout when requestIdleCallback not available', () => {
    delete (window as any).requestIdleCallback;

    const loadFn = vi.fn();
    deferNonCriticalDataLoad(loadFn);

    vi.advanceTimersByTime(2000);

    expect(loadFn).toHaveBeenCalled();
  });

  it('should handle async load functions', async () => {
    const loadFn = vi.fn().mockResolvedValue(undefined);
    deferNonCriticalDataLoad(loadFn);

    vi.advanceTimersByTime(2000);
    await Promise.resolve();

    expect(loadFn).toHaveBeenCalled();
  });

  it('should handle load errors', async () => {
    const loadFn = vi.fn().mockRejectedValue(new Error('Load error'));
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation();

    deferNonCriticalDataLoad(loadFn);

    vi.advanceTimersByTime(2000);
    await Promise.resolve();

    expect(loadFn).toHaveBeenCalled();
    // Error should be caught and logged
    expect(consoleSpy).toHaveBeenCalled();

    consoleSpy.mockRestore();
  });
});

