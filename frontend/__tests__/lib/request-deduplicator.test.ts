import { RequestDeduplicator } from '@/lib/request-deduplicator';
import { beforeEach, describe, expect, it, vi } from 'vitest';

describe('RequestDeduplicator', () => {
  let deduplicator: RequestDeduplicator;

  beforeEach(() => {
    deduplicator = new RequestDeduplicator();
  });

  it('should execute request function', async () => {
    const mockFn = vi.fn().mockResolvedValue('result');
    const result = await deduplicator.deduplicate('key1', mockFn);

    expect(result).toBe('result');
    expect(mockFn).toHaveBeenCalledTimes(1);
  });

  it('should deduplicate concurrent requests with same key', async () => {
    const mockFn = vi.fn().mockImplementation(
      () => new Promise((resolve) => setTimeout(() => resolve('result'), 100))
    );

    const promise1 = deduplicator.deduplicate('key1', mockFn);
    const promise2 = deduplicator.deduplicate('key1', mockFn);
    const promise3 = deduplicator.deduplicate('key1', mockFn);

    const [result1, result2, result3] = await Promise.all([promise1, promise2, promise3]);

    expect(result1).toBe('result');
    expect(result2).toBe('result');
    expect(result3).toBe('result');
    expect(mockFn).toHaveBeenCalledTimes(1); // Only called once
  });

  it('should not deduplicate requests with different keys', async () => {
    const mockFn = vi.fn().mockResolvedValue('result');

    await Promise.all([
      deduplicator.deduplicate('key1', mockFn),
      deduplicator.deduplicate('key2', mockFn),
      deduplicator.deduplicate('key3', mockFn),
    ]);

    expect(mockFn).toHaveBeenCalledTimes(3);
  });

  it('should handle errors and remove from pending', async () => {
    const mockFn = vi.fn().mockRejectedValue(new Error('Test error'));

    await expect(deduplicator.deduplicate('key1', mockFn)).rejects.toThrow('Test error');

    // Should be able to retry after error
    mockFn.mockResolvedValue('result');
    const result = await deduplicator.deduplicate('key1', mockFn);
    expect(result).toBe('result');
  });

  it('should clear all pending requests', async () => {
    const mockFn = vi.fn().mockImplementation(
      () => new Promise((resolve) => setTimeout(() => resolve('result'), 100))
    );

    deduplicator.deduplicate('key1', mockFn);
    deduplicator.clear();

    // After clear, new request should execute
    const result = await deduplicator.deduplicate('key1', mockFn);
    expect(result).toBe('result');
    expect(mockFn).toHaveBeenCalledTimes(2);
  });

  it('should handle multiple concurrent requests with same key', async () => {
    const mockFn = vi.fn().mockImplementation(
      () => new Promise((resolve) => setTimeout(() => resolve('result'), 50))
    );

    // Start 10 concurrent requests with same key
    const promises = Array.from({ length: 10 }, () =>
      deduplicator.deduplicate('key1', mockFn)
    );

    const results = await Promise.all(promises);

    // All should return same result
    results.forEach((result) => {
      expect(result).toBe('result');
    });

    // Function should only be called once
    expect(mockFn).toHaveBeenCalledTimes(1);
  });

  it('should handle cleanup of completed requests', async () => {
    const mockFn = vi.fn().mockResolvedValue('result');

    // Execute request
    await deduplicator.deduplicate('key1', mockFn);

    // After completion, new request with same key should execute again
    const result = await deduplicator.deduplicate('key1', mockFn);
    expect(result).toBe('result');
    expect(mockFn).toHaveBeenCalledTimes(2);
  });

  it('should generate unique keys for different requests', async () => {
    const mockFn = vi.fn().mockResolvedValue('result');

    await Promise.all([
      deduplicator.deduplicate('key1', mockFn),
      deduplicator.deduplicate('key2', mockFn),
      deduplicator.deduplicate('key3', mockFn),
    ]);

    // Each unique key should trigger a separate execution
    expect(mockFn).toHaveBeenCalledTimes(3);
  });

  it('should handle rapid sequential requests', async () => {
    const mockFn = vi.fn().mockResolvedValue('result');

    // Execute requests sequentially (not concurrently)
    await deduplicator.deduplicate('key1', mockFn);
    await deduplicator.deduplicate('key1', mockFn);
    await deduplicator.deduplicate('key1', mockFn);

    // Each sequential request should execute separately
    expect(mockFn).toHaveBeenCalledTimes(3);
  });
});

