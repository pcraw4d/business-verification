import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { RequestDeduplicator } from '@/lib/request-deduplicator';

describe('RequestDeduplicator', () => {
  let deduplicator: RequestDeduplicator;

  beforeEach(() => {
    deduplicator = new RequestDeduplicator();
  });

  it('should execute request function', async () => {
    const mockFn = jest.fn().mockResolvedValue('result');
    const result = await deduplicator.deduplicate('key1', mockFn);

    expect(result).toBe('result');
    expect(mockFn).toHaveBeenCalledTimes(1);
  });

  it('should deduplicate concurrent requests with same key', async () => {
    const mockFn = jest.fn().mockImplementation(
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
    const mockFn = jest.fn().mockResolvedValue('result');

    await Promise.all([
      deduplicator.deduplicate('key1', mockFn),
      deduplicator.deduplicate('key2', mockFn),
      deduplicator.deduplicate('key3', mockFn),
    ]);

    expect(mockFn).toHaveBeenCalledTimes(3);
  });

  it('should handle errors and remove from pending', async () => {
    const mockFn = jest.fn().mockRejectedValue(new Error('Test error'));

    await expect(deduplicator.deduplicate('key1', mockFn)).rejects.toThrow('Test error');

    // Should be able to retry after error
    mockFn.mockResolvedValue('result');
    const result = await deduplicator.deduplicate('key1', mockFn);
    expect(result).toBe('result');
  });

  it('should clear all pending requests', async () => {
    const mockFn = jest.fn().mockImplementation(
      () => new Promise((resolve) => setTimeout(() => resolve('result'), 100))
    );

    deduplicator.deduplicate('key1', mockFn);
    deduplicator.clear();

    // After clear, new request should execute
    const result = await deduplicator.deduplicate('key1', mockFn);
    expect(result).toBe('result');
    expect(mockFn).toHaveBeenCalledTimes(2);
  });
});

