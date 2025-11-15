import { describe, it, expect, jest } from '@jest/globals';
import { deferNonCriticalDataLoad } from '@/lib/lazy-loader';

describe('Lazy Loading Performance', () => {
  it('should defer loading until element is visible', async () => {
    const loadFn = jest.fn();
    const mockElement = document.createElement('div');
    
    // Mock IntersectionObserver
    const mockObserver = {
      observe: jest.fn(),
      unobserve: jest.fn(),
      disconnect: jest.fn(),
    };
    
    global.IntersectionObserver = jest.fn(() => mockObserver) as any;

    deferNonCriticalDataLoad(loadFn, mockElement);

    // Function should not be called immediately
    expect(loadFn).not.toHaveBeenCalled();
    
    // Observer should be set up
    expect(mockObserver.observe).toHaveBeenCalledWith(mockElement);
  });

  it('should load data when element becomes visible', () => {
    const loadFn = jest.fn();
    const mockElement = document.createElement('div');
    
    let intersectionCallback: IntersectionObserverCallback | null = null;
    
    global.IntersectionObserver = jest.fn((callback) => {
      intersectionCallback = callback;
      return {
        observe: jest.fn(),
        unobserve: jest.fn(),
        disconnect: jest.fn(),
      } as any;
    }) as any;

    deferNonCriticalDataLoad(loadFn, mockElement);

    // Simulate intersection
    if (intersectionCallback) {
      intersectionCallback([{
        target: mockElement,
        isIntersecting: true,
        intersectionRatio: 1,
        boundingClientRect: {} as DOMRectReadOnly,
        rootBounds: null,
        time: Date.now(),
      }], {} as IntersectionObserver);
    }

    expect(loadFn).toHaveBeenCalled();
  });

  it('should not load data multiple times', () => {
    const loadFn = jest.fn();
    const mockElement = document.createElement('div');
    
    let intersectionCallback: IntersectionObserverCallback | null = null;
    
    global.IntersectionObserver = jest.fn((callback) => {
      intersectionCallback = callback;
      return {
        observe: jest.fn(),
        unobserve: jest.fn(),
        disconnect: jest.fn(),
      } as any;
    }) as any;

    deferNonCriticalDataLoad(loadFn, mockElement);

    // Simulate multiple intersections
    if (intersectionCallback) {
      const entry = {
        target: mockElement,
        isIntersecting: true,
        intersectionRatio: 1,
        boundingClientRect: {} as DOMRectReadOnly,
        rootBounds: null,
        time: Date.now(),
      };
      
      intersectionCallback([entry], {} as IntersectionObserver);
      intersectionCallback([entry], {} as IntersectionObserver);
      intersectionCallback([entry], {} as IntersectionObserver);
    }

    // Should only be called once
    expect(loadFn).toHaveBeenCalledTimes(1);
  });
});

