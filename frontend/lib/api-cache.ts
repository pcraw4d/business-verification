// API Response Caching Utility

interface CacheEntry<T> {
  data: T;
  expiresAt: number;
}

export class APICache {
  private cache: Map<string, CacheEntry<unknown>>;
  private defaultTTL: number;

  constructor(defaultTTL: number = 5 * 60 * 1000) {
    // 5 minutes default
    this.cache = new Map();
    this.defaultTTL = defaultTTL;
  }

  /**
   * Gets cached data if available and not expired
   */
  get<T>(key: string): T | null {
    const entry = this.cache.get(key);
    if (!entry) {
      return null;
    }

    // Check if expired
    if (Date.now() > entry.expiresAt) {
      this.cache.delete(key);
      return null;
    }

    return entry.data as T;
  }

  /**
   * Sets data in cache with TTL
   */
  set<T>(key: string, data: T, ttl?: number): void {
    const expiresAt = Date.now() + (ttl || this.defaultTTL);
    this.cache.set(key, {
      data,
      expiresAt,
    });
  }

  /**
   * Clears all cache entries
   */
  clear(): void {
    this.cache.clear();
  }

  /**
   * Removes a specific cache entry
   */
  delete(key: string): void {
    this.cache.delete(key);
  }

  /**
   * Persists cache entry to sessionStorage
   */
  persist(key: string): void {
    const entry = this.cache.get(key);
    if (entry && typeof window !== 'undefined') {
      try {
        sessionStorage.setItem(`cache:${key}`, JSON.stringify(entry));
      } catch (error) {
        console.warn('Failed to persist cache entry:', error);
      }
    }
  }

  /**
   * Restores cache entry from sessionStorage
   */
  restore<T>(key: string): T | null {
    if (typeof window === 'undefined') {
      return null;
    }

    try {
      const stored = sessionStorage.getItem(`cache:${key}`);
      if (stored) {
        const entry = JSON.parse(stored) as CacheEntry<T>;
        // Check if expired
        if (Date.now() > entry.expiresAt) {
          sessionStorage.removeItem(`cache:${key}`);
          return null;
        }
        // Restore to memory cache
        this.cache.set(key, entry as CacheEntry<unknown>);
        return entry.data;
      }
    } catch (error) {
      console.warn('Failed to restore cache entry:', error);
    }

    return null;
  }

  /**
   * Generates cache key from URL and options
   */
  static generateKey(url: string, options?: RequestInit): string {
    const method = options?.method || 'GET';
    const body = options?.body ? JSON.stringify(options.body) : '';
    return `${method}:${url}:${body}`;
  }
}

/**
 * Cached fetch wrapper
 */
export async function cachedFetch<T>(
  url: string,
  options?: RequestInit,
  cache?: APICache,
  ttl?: number
): Promise<T> {
  // Only cache GET requests
  if (options?.method && options.method !== 'GET') {
    const response = await fetch(url, options);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    return response.json();
  }

  if (!cache) {
    const response = await fetch(url, options);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    return response.json();
  }

  // Generate cache key
  const cacheKey = APICache.generateKey(url, options);

  // Check cache first
  const cached = cache.get<T>(cacheKey);
  if (cached !== null) {
    return cached;
  }

  // Make request
  const response = await fetch(url, options);
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }

  const data = await response.json<T>();

  // Cache response
  cache.set(cacheKey, data, ttl);

  return data;
}

