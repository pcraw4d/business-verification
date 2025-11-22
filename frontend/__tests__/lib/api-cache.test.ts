// Vitest globals are available via globals: true in vitest.config.ts
import { APICache, cachedFetch } from '@/lib/api-cache';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

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

  it('should handle cache hit scenario', async () => {
    const cache = new APICache();
    const mockData = { id: '123' };

    vi.mocked(global.fetch).mockResolvedValueOnce({
      ok: true,
      json: async () => mockData,
    } as Response);

    // First call - cache miss
    const result1 = await cachedFetch('/api/test', { method: 'GET' }, cache);
    expect(result1).toEqual(mockData);
    expect(global.fetch).toHaveBeenCalledTimes(1);

    // Second call - cache hit
    const result2 = await cachedFetch('/api/test', { method: 'GET' }, cache);
    expect(result2).toEqual(mockData);
    expect(global.fetch).toHaveBeenCalledTimes(1); // Still 1, cache hit
  });

  it('should handle cache miss scenario', async () => {
    const cache = new APICache();
    const mockData1 = { id: '123' };
    const mockData2 = { id: '456' };

    vi.mocked(global.fetch)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockData1,
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockData2,
      } as Response);

    // First call - cache miss
    const result1 = await cachedFetch('/api/test1', { method: 'GET' }, cache);
    expect(result1).toEqual(mockData1);
    expect(global.fetch).toHaveBeenCalledTimes(1);

    // Second call with different URL - cache miss
    const result2 = await cachedFetch('/api/test2', { method: 'GET' }, cache);
    expect(result2).toEqual(mockData2);
    expect(global.fetch).toHaveBeenCalledTimes(2); // Cache miss, new fetch
  });

  it('should invalidate cache entry', async () => {
    const cache = new APICache();
    const mockData = { id: '123' };

    vi.mocked(global.fetch).mockResolvedValue({
      ok: true,
      json: async () => mockData,
    } as Response);

    // First call - cache miss
    await cachedFetch('/api/test', { method: 'GET' }, cache);
    expect(global.fetch).toHaveBeenCalledTimes(1);

    // Second call - cache hit
    await cachedFetch('/api/test', { method: 'GET' }, cache);
    expect(global.fetch).toHaveBeenCalledTimes(1);

    // Invalidate cache
    cache.delete(APICache.generateKey('/api/test', { method: 'GET' }));

    // Third call - cache miss after invalidation
    await cachedFetch('/api/test', { method: 'GET' }, cache);
    expect(global.fetch).toHaveBeenCalledTimes(2);
  });

  it('should handle multiple cache entries', () => {
    const cache = new APICache(1000);
    const data1 = { id: '1' };
    const data2 = { id: '2' };
    const data3 = { id: '3' };

    cache.set('key1', data1);
    cache.set('key2', data2);
    cache.set('key3', data3);
    
    expect(cache.get('key1')).toEqual(data1);
    expect(cache.get('key2')).toEqual(data2);
    expect(cache.get('key3')).toEqual(data3);
  });

  it('should handle cache expiration cleanup', () => {
    const cache = new APICache(50); // 50ms TTL
    const data = { id: '1' };

    cache.set('key1', data);
    expect(cache.get('key1')).toEqual(data);

    return new Promise<void>((resolve) => {
      setTimeout(() => {
        // After expiration, entry should be cleaned up
        const result = cache.get('key1');
        expect(result).toBeNull();
        resolve();
      }, 100);
    });
  });

  describe('persist and restore', () => {
    beforeEach(() => {
      // Mock sessionStorage
      const sessionStorageMock = {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
      };
      Object.defineProperty(window, 'sessionStorage', {
        value: sessionStorageMock,
        writable: true,
      });
    });

    it('should persist cache entry to sessionStorage', () => {
      const cache = new APICache();
      const data = { id: '123' };
      cache.set('test-key', data);
      
      cache.persist('test-key');
      
      expect(window.sessionStorage.setItem).toHaveBeenCalledWith(
        'cache:test-key',
        expect.stringContaining('"data":{"id":"123"}')
      );
    });

    it('should not persist if entry does not exist', () => {
      const cache = new APICache();
      
      cache.persist('non-existent');
      
      expect(window.sessionStorage.setItem).not.toHaveBeenCalled();
    });

    it('should not persist if window is undefined (SSR)', () => {
      const originalWindow = global.window;
      // @ts-ignore
      delete global.window;
      
      const cache = new APICache();
      const data = { id: '123' };
      cache.set('test-key', data);
      
      // Should not throw
      expect(() => cache.persist('test-key')).not.toThrow();
      
      global.window = originalWindow;
    });

    it('should handle persist errors gracefully', () => {
      const cache = new APICache();
      const data = { id: '123' };
      cache.set('test-key', data);
      
      // Mock setItem to throw
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      vi.mocked(window.sessionStorage.setItem).mockImplementation(() => {
        throw new Error('Storage quota exceeded');
      });
      
      expect(() => cache.persist('test-key')).not.toThrow();
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        'Failed to persist cache entry:',
        expect.any(Error)
      );
      
      consoleWarnSpy.mockRestore();
    });

    it('should restore cache entry from sessionStorage', () => {
      const cache = new APICache();
      const data = { id: '123' };
      const entry = {
        data,
        expiresAt: Date.now() + 5000, // 5 seconds from now
      };
      
      vi.mocked(window.sessionStorage.getItem).mockReturnValue(
        JSON.stringify(entry)
      );
      
      const result = cache.restore<typeof data>('test-key');
      
      expect(result).toEqual(data);
      expect(cache.get('test-key')).toEqual(data);
    });

    it('should return null if entry not in sessionStorage', () => {
      const cache = new APICache();
      
      vi.mocked(window.sessionStorage.getItem).mockReturnValue(null);
      
      const result = cache.restore('test-key');
      
      expect(result).toBeNull();
    });

    it('should return null if entry is expired', () => {
      const cache = new APICache();
      const data = { id: '123' };
      const entry = {
        data,
        expiresAt: Date.now() - 1000, // Expired 1 second ago
      };
      
      vi.mocked(window.sessionStorage.getItem).mockReturnValue(
        JSON.stringify(entry)
      );
      
      const result = cache.restore<typeof data>('test-key');
      
      expect(result).toBeNull();
      expect(window.sessionStorage.removeItem).toHaveBeenCalledWith('cache:test-key');
    });

    it('should return null if window is undefined (SSR)', () => {
      const originalWindow = global.window;
      // @ts-ignore
      delete global.window;
      
      const cache = new APICache();
      
      const result = cache.restore('test-key');
      
      expect(result).toBeNull();
      
      global.window = originalWindow;
    });

    it('should handle restore errors gracefully', () => {
      const cache = new APICache();
      
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      vi.mocked(window.sessionStorage.getItem).mockReturnValue('invalid json');
      
      const result = cache.restore('test-key');
      
      expect(result).toBeNull();
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        'Failed to restore cache entry:',
        expect.any(Error)
      );
      
      consoleWarnSpy.mockRestore();
    });

    it('should handle restore when JSON.parse throws', () => {
      const cache = new APICache();
      
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      vi.mocked(window.sessionStorage.getItem).mockReturnValue('{invalid json}');
      
      const result = cache.restore('test-key');
      
      expect(result).toBeNull();
      expect(consoleWarnSpy).toHaveBeenCalled();
      
      consoleWarnSpy.mockRestore();
    });
  });

  describe('cachedFetch edge cases', () => {
    it('should handle fetch without cache', async () => {
      const mockData = { id: '123' };

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData,
      } as Response);

      const result = await cachedFetch('/api/test', { method: 'GET' });

      expect(result).toEqual(mockData);
      expect(global.fetch).toHaveBeenCalledTimes(1);
    });

    it('should handle non-GET requests without cache', async () => {
      const mockData = { id: '123' };

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData,
      } as Response);

      const result = await cachedFetch('/api/test', { method: 'PUT' });

      expect(result).toEqual(mockData);
      expect(global.fetch).toHaveBeenCalledTimes(1);
    });

    it('should handle cache miss with null return', async () => {
      const cache = new APICache();
      const mockData = { id: '123' };

      // Mock cache.get to return null (cache miss)
      vi.spyOn(cache, 'get').mockReturnValue(null);

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockData,
      } as Response);

      const result = await cachedFetch('/api/test', { method: 'GET' }, cache);

      expect(result).toEqual(mockData);
      expect(global.fetch).toHaveBeenCalledTimes(1);
    });
  });
});

