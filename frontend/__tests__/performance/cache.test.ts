
import { APICache } from '@/lib/api-cache';

describe('Cache Performance', () => {
  it('should retrieve cached data quickly', () => {
    const cache = new APICache(5 * 60 * 1000); // 5 minutes
    const testData = { id: 'test-123', name: 'Test' };
    
    cache.set('test-key', testData);
    
    const start = performance.now();
    const result = cache.get('test-key');
    const duration = performance.now() - start;
    
    expect(result).toEqual(testData);
    expect(duration).toBeLessThan(1); // Should be < 1ms
  });

  it('should handle cache misses efficiently', () => {
    const cache = new APICache(5 * 60 * 1000);
    
    const start = performance.now();
    const result = cache.get('non-existent-key');
    const duration = performance.now() - start;
    
    expect(result).toBeNull();
    expect(duration).toBeLessThan(1); // Should be < 1ms
  });

  it('should handle many cache operations efficiently', () => {
    const cache = new APICache(5 * 60 * 1000);
    const iterations = 1000;
    
    // Set many items
    const setStart = performance.now();
    for (let i = 0; i < iterations; i++) {
      cache.set(`key-${i}`, { id: i });
    }
    const setDuration = performance.now() - setStart;
    
    // Get many items
    const getStart = performance.now();
    for (let i = 0; i < iterations; i++) {
      cache.get(`key-${i}`);
    }
    const getDuration = performance.now() - getStart;
    
    const avgSetTime = setDuration / iterations;
    const avgGetTime = getDuration / iterations;
    
    expect(avgSetTime).toBeLessThan(0.1); // < 0.1ms per set
    expect(avgGetTime).toBeLessThan(0.1); // < 0.1ms per get
  });

  it('should expire cached data correctly', () => {
    const cache = new APICache(100); // 100ms TTL
    
    cache.set('expiring-key', { data: 'test' });
    
    // Should be available immediately
    expect(cache.get('expiring-key')).not.toBeNull();
    
    // Wait for expiration
    return new Promise<void>((resolve) => {
      setTimeout(() => {
        expect(cache.get('expiring-key')).toBeNull();
        resolve();
      }, 150);
    });
  });
});

