
import { RequestDeduplicator } from '@/lib/request-deduplicator';

describe('Request Deduplication Performance', () => {
  it('should deduplicate concurrent requests', async () => {
    const deduplicator = new RequestDeduplicator();
    const fetchFn = vi.fn().mockResolvedValue({ data: 'test' });
    
    const key = 'test-key';
    
    // Make concurrent requests
    const promises = [
      deduplicator.deduplicate(key, fetchFn),
      deduplicator.deduplicate(key, fetchFn),
      deduplicator.deduplicate(key, fetchFn),
    ];
    
    await Promise.all(promises);
    
    // Function should only be called once
    expect(fetchFn).toHaveBeenCalledTimes(1);
  });

  it('should handle many concurrent requests efficiently', async () => {
    const deduplicator = new RequestDeduplicator();
    const fetchFn = vi.fn().mockResolvedValue({ data: 'test' });
    
    const concurrentRequests = 100;
    const promises = [];
    
    const start = performance.now();
    
    for (let i = 0; i < concurrentRequests; i++) {
      promises.push(deduplicator.deduplicate('same-key', fetchFn));
    }
    
    await Promise.all(promises);
    
    const duration = performance.now() - start;
    
    // Should complete quickly
    expect(duration).toBeLessThan(100); // < 100ms
    expect(fetchFn).toHaveBeenCalledTimes(1);
  });

  it('should handle different keys independently', async () => {
    const deduplicator = new RequestDeduplicator();
    const fetchFn = vi.fn().mockResolvedValue({ data: 'test' });
    
    await Promise.all([
      deduplicator.deduplicate('key-1', fetchFn),
      deduplicator.deduplicate('key-2', fetchFn),
      deduplicator.deduplicate('key-3', fetchFn),
    ]);
    
    // Each key should trigger a separate call
    expect(fetchFn).toHaveBeenCalledTimes(3);
  });

  it('should handle errors in deduplicated requests', async () => {
    const deduplicator = new RequestDeduplicator();
    const error = new Error('Test error');
    const fetchFn = vi.fn().mockRejectedValue(error);
    
    const promises = [
      deduplicator.deduplicate('error-key', fetchFn),
      deduplicator.deduplicate('error-key', fetchFn),
    ];
    
    await expect(Promise.all(promises)).rejects.toThrow('Test error');
    expect(fetchFn).toHaveBeenCalledTimes(1);
  });
});

