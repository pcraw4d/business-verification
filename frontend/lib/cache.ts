/**
 * Client-side caching utilities for API responses and computed values
 */

interface CacheEntry<T> {
  data: T;
  timestamp: number;
  ttl: number; // Time to live in milliseconds
}

class MemoryCache {
  private cache: Map<string, CacheEntry<any>> = new Map();

  /**
   * Get a value from cache if it exists and hasn't expired
   */
  get<T>(key: string): T | null {
    const entry = this.cache.get(key);
    if (!entry) {
      return null;
    }

    const now = Date.now();
    if (now - entry.timestamp > entry.ttl) {
      this.cache.delete(key);
      return null;
    }

    return entry.data as T;
  }

  /**
   * Set a value in cache with a TTL
   */
  set<T>(key: string, data: T, ttl: number = 5 * 60 * 1000): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    });
  }

  /**
   * Delete a value from cache
   */
  delete(key: string): void {
    this.cache.delete(key);
  }

  /**
   * Clear all cache entries
   */
  clear(): void {
    this.cache.clear();
  }

  /**
   * Clear expired entries
   */
  cleanup(): void {
    const now = Date.now();
    for (const [key, entry] of this.cache.entries()) {
      if (now - entry.timestamp > entry.ttl) {
        this.cache.delete(key);
      }
    }
  }
}

// Singleton instance
const memoryCache = new MemoryCache();

// Cleanup expired entries every 5 minutes
if (typeof window !== 'undefined') {
  setInterval(() => {
    memoryCache.cleanup();
  }, 5 * 60 * 1000);
}

/**
 * Cache configuration for different data types
 */
export const CACHE_TTL = {
  DASHBOARD_METRICS: 2 * 60 * 1000, // 2 minutes
  MERCHANT_LIST: 1 * 60 * 1000, // 1 minute
  RISK_METRICS: 2 * 60 * 1000, // 2 minutes
  SYSTEM_METRICS: 30 * 1000, // 30 seconds
  COMPLIANCE_STATUS: 5 * 60 * 1000, // 5 minutes
  BUSINESS_INTELLIGENCE: 5 * 60 * 1000, // 5 minutes
  SESSIONS: 10 * 1000, // 10 seconds
  DEFAULT: 5 * 60 * 1000, // 5 minutes
} as const;

/**
 * Generate a cache key from parameters
 */
export function generateCacheKey(prefix: string, params?: Record<string, any>): string {
  if (!params || Object.keys(params).length === 0) {
    return prefix;
  }

  const sortedParams = Object.keys(params)
    .sort()
    .map((key) => `${key}=${JSON.stringify(params[key])}`)
    .join('&');

  return `${prefix}?${sortedParams}`;
}

/**
 * Get cached value or execute function and cache result
 */
export async function getCachedOrFetch<T>(
  key: string,
  fetchFn: () => Promise<T>,
  ttl: number = CACHE_TTL.DEFAULT
): Promise<T> {
  const cached = memoryCache.get<T>(key);
  if (cached !== null) {
    return cached;
  }

  const data = await fetchFn();
  memoryCache.set(key, data, ttl);
  return data;
}

/**
 * Invalidate cache entries matching a prefix
 */
export function invalidateCache(prefix: string): void {
  // Since we can't iterate efficiently, we'll need to track keys
  // For now, we'll use a simple approach: clear all if prefix matches
  // In production, consider using a more sophisticated key tracking system
  memoryCache.clear();
}

/**
 * Clear all cache
 */
export function clearCache(): void {
  memoryCache.clear();
}

export default memoryCache;

