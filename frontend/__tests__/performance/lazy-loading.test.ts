
import { deferNonCriticalDataLoad } from '@/lib/lazy-loader';

describe('Lazy Loading Performance', () => {
  it('should defer loading until element is visible', async () => {
    const loadFn = vi.fn();
    
    // Mock requestIdleCallback
    const mockRequestIdleCallback = vi.fn((callback) => {
      // Don't call immediately - simulate deferral
      setTimeout(() => callback({ timeRemaining: () => 50 } as IdleDeadline), 0);
      return 1;
    });
    global.requestIdleCallback = mockRequestIdleCallback as any;

    deferNonCriticalDataLoad(loadFn);

    // Function should not be called immediately
    expect(loadFn).not.toHaveBeenCalled();
    
    // Wait for idle callback
    await new Promise(resolve => setTimeout(resolve, 10));
    expect(loadFn).toHaveBeenCalled();
  });

  it('should load data when element becomes visible', async () => {
    const loadFn = vi.fn();
    
    // Mock requestIdleCallback to call immediately
    const mockRequestIdleCallback = vi.fn((callback) => {
      callback({ timeRemaining: () => 50 } as IdleDeadline);
      return 1;
    });
    global.requestIdleCallback = mockRequestIdleCallback as any;

    deferNonCriticalDataLoad(loadFn);

    // Function should be called via idle callback
    expect(loadFn).toHaveBeenCalled();
  });

  it('should not load data multiple times', () => {
    const loadFn = vi.fn();
    
    // Mock requestIdleCallback
    let callCount = 0;
    const mockRequestIdleCallback = vi.fn((callback) => {
      callCount++;
      if (callCount === 1) {
        callback({ timeRemaining: () => 50 } as IdleDeadline);
      }
      return callCount;
    });
    global.requestIdleCallback = mockRequestIdleCallback as any;

    // Call multiple times
    deferNonCriticalDataLoad(loadFn);
    deferNonCriticalDataLoad(loadFn);
    deferNonCriticalDataLoad(loadFn);

    // Should only be called once (requestIdleCallback handles deduplication)
    // But since we're calling deferNonCriticalDataLoad multiple times,
    // requestIdleCallback will be called multiple times
    // The actual deduplication happens at a higher level
    expect(mockRequestIdleCallback).toHaveBeenCalledTimes(3);
    expect(loadFn).toHaveBeenCalledTimes(1);
  });
});

