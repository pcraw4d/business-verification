// Vitest globals are available via globals: true in vitest.config.ts
import { APICache, cachedFetch } from '@/lib/api-cache';

describe('APICache', () => {
  let cache: APICache;

  beforeEach(() => {
    cache = new APICache(5 * 60 * 1000); // 5 minutes
  });

  describe('get and set', () => {
    it('should store and retrieve data', () => {
      const testData = { id: '123', name: 'Test' };
      cache.set('test-key', testData);

      const result = cache.get<typeof testData>('test-key');
      expect(result).toEqual(testData);
    });

    it('should return null for non-existent keys', () => {
      const result = cache.get('non-existent');
      expect(result).toBeNull();
    });

    it('should expire data after TTL', () => {
      const cache = new APICache(100); // 100ms TTL
      cache.set('test-key', { data: 'test' });

      // Wait for expiration
      return new Promise<void>((resolve) => {
        setTimeout(() => {
          const result = cache.get('test-key');
          expect(result).toBeNull();
          resolve();
        }, 150);
      });
    });

    it('should use custom TTL when provided', () => {
      const cache = new APICache(1000); // 1 second default
      cache.set('test-key', { data: 'test' }, 50); // 50ms custom TTL

      return new Promise<void>((resolve) => {
        setTimeout(() => {
          const result = cache.get('test-key');
          expect(result).toBeNull();
          resolve();
        }, 100);
      });
    });
  });

  describe('clear and delete', () => {
    it('should clear all cache entries', () => {
      cache.set('key1', { data: '1' });
      cache.set('key2', { data: '2' });

      cache.clear();

      expect(cache.get('key1')).toBeNull();
      expect(cache.get('key2')).toBeNull();
    });

    it('should delete specific cache entry', () => {
      cache.set('key1', { data: '1' });
      cache.set('key2', { data: '2' });

      cache.delete('key1');

      expect(cache.get('key1')).toBeNull();
      expect(cache.get('key2')).not.toBeNull();
    });
  });

  describe('generateKey', () => {
    it('should generate consistent cache keys', () => {
      const key1 = APICache.generateKey('/api/test', { method: 'GET' });
      const key2 = APICache.generateKey('/api/test', { method: 'GET' });

      expect(key1).toBe(key2);
    });

    it('should generate different keys for different methods', () => {
      const key1 = APICache.generateKey('/api/test', { method: 'GET' });
      const key2 = APICache.generateKey('/api/test', { method: 'POST' });

      expect(key1).not.toBe(key2);
    });

    it('should include body in cache key for POST requests', () => {
      const key1 = APICache.generateKey('/api/test', {
        method: 'POST',
        body: JSON.stringify({ a: 1 }),
      });
      const key2 = APICache.generateKey('/api/test', {
        method: 'POST',
        body: JSON.stringify({ a: 2 }),
      });

      expect(key1).not.toBe(key2);
    });
  });
});

describe('cachedFetch', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock fetch globally for these tests
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should cache GET requests', async () => {
    const cache = new APICache();
    const mockData = { id: '123' };

    vi.mocked(global.fetch).mockResolvedValueOnce({
      ok: true,
      json: async () => mockData,
    } as Response);

    // First call - should fetch
    const result1 = await cachedFetch('/api/test', { method: 'GET' }, cache);
    expect(result1).toEqual(mockData);
    expect(global.fetch).toHaveBeenCalledTimes(1);

    // Second call - should use cache
    const result2 = await cachedFetch('/api/test', { method: 'GET' }, cache);
    expect(result2).toEqual(mockData);
    expect(global.fetch).toHaveBeenCalledTimes(1); // Still 1, not 2
  });

  it('should not cache POST requests', async () => {
    const cache = new APICache();
    const mockData = { id: '123' };

    vi.mocked(global.fetch).mockResolvedValue({
      ok: true,
      json: async () => mockData,
    } as Response);

    await cachedFetch('/api/test', { method: 'POST' }, cache);
    await cachedFetch('/api/test', { method: 'POST' }, cache);

    expect(global.fetch).toHaveBeenCalledTimes(2);
  });

  it('should handle errors', async () => {
    const cache = new APICache();

    vi.mocked(global.fetch).mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
    } as Response);

    await expect(
      cachedFetch('/api/test', { method: 'GET' }, cache)
    ).rejects.toThrow();
  });
});

