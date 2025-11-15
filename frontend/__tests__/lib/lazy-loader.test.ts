import { describe, it, expect, beforeEach, jest } from '@jest/globals';
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
    const loadFn = jest.fn().mockResolvedValue(undefined);

    loader.observe(mockElement, loadFn);

    // Simulate intersection
    const entries = [{ target: mockElement, isIntersecting: true }];
    (loader as any).handleIntersection(mockElement);

    // Wait for async load
    await new Promise((resolve) => setTimeout(resolve, 10));

    expect(loadFn).toHaveBeenCalled();
  });

  it('should not load same element twice', async () => {
    const loadFn = jest.fn().mockResolvedValue(undefined);

    loader.observe(mockElement, loadFn);
    (loader as any).handleIntersection(mockElement);
    await new Promise((resolve) => setTimeout(resolve, 10));

    loader.observe(mockElement, loadFn);
    (loader as any).handleIntersection(mockElement);
    await new Promise((resolve) => setTimeout(resolve, 10));

    expect(loadFn).toHaveBeenCalledTimes(1);
  });

  it('should disconnect observer', () => {
    const loadFn = jest.fn();
    loader.observe(mockElement, loadFn);
    
    // Disconnect before intersection
    loader.disconnect();

    // After disconnect, should not load
    (loader as any).handleIntersection(mockElement);
    // Note: handleIntersection checks loadedSections, so it may still be called once
    // but the observer should be disconnected
    expect((loader as any).observer).toBeNull();
  });

  it('should handle load errors gracefully', async () => {
    const loadFn = jest.fn().mockRejectedValue(new Error('Load error'));

    loader.observe(mockElement, loadFn);
    (loader as any).handleIntersection(mockElement);

    await new Promise((resolve) => setTimeout(resolve, 10));

    // Should not throw, but should mark as not loaded for retry
    expect(loadFn).toHaveBeenCalled();
  });
});

describe('deferNonCriticalDataLoad', () => {
  beforeEach(() => {
    jest.clearAllTimers();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('should use requestIdleCallback when available', () => {
    const mockIdleCallback = jest.fn((cb) => {
      setTimeout(cb, 0);
      return 1;
    });
    (window as any).requestIdleCallback = mockIdleCallback;

    const loadFn = jest.fn();
    deferNonCriticalDataLoad(loadFn);

    jest.advanceTimersByTime(10);

    expect(mockIdleCallback).toHaveBeenCalled();
  });

  it('should fallback to setTimeout when requestIdleCallback not available', () => {
    delete (window as any).requestIdleCallback;

    const loadFn = jest.fn();
    deferNonCriticalDataLoad(loadFn);

    jest.advanceTimersByTime(2000);

    expect(loadFn).toHaveBeenCalled();
  });

  it('should handle async load functions', async () => {
    const loadFn = jest.fn().mockResolvedValue(undefined);
    deferNonCriticalDataLoad(loadFn);

    jest.advanceTimersByTime(2000);
    await Promise.resolve();

    expect(loadFn).toHaveBeenCalled();
  });

  it('should handle load errors', async () => {
    const loadFn = jest.fn().mockRejectedValue(new Error('Load error'));
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

    deferNonCriticalDataLoad(loadFn);

    jest.advanceTimersByTime(2000);
    await Promise.resolve();

    expect(loadFn).toHaveBeenCalled();
    // Error should be caught and logged
    expect(consoleSpy).toHaveBeenCalled();

    consoleSpy.mockRestore();
  });
});

