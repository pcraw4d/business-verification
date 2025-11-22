import { render } from '@testing-library/react';
import { PerformanceOptimizer } from '@/components/performance/PerformanceOptimizer';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import * as preloadModule from '@/lib/preload';

// Mock the preload module
vi.mock('@/lib/preload', () => ({
  initPerformanceOptimizations: vi.fn(),
  dnsPrefetch: vi.fn(),
}));

describe('PerformanceOptimizer', () => {
  const mockDnsPrefetch = vi.mocked(preloadModule.dnsPrefetch);
  const mockInitPerformanceOptimizations = vi.mocked(preloadModule.initPerformanceOptimizations);

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize performance optimizations on mount', () => {
    render(<PerformanceOptimizer />);
    
    expect(mockDnsPrefetch).toHaveBeenCalled();
    expect(mockInitPerformanceOptimizations).toHaveBeenCalled();
  });

  it('should use NEXT_PUBLIC_API_BASE_URL for DNS prefetch', () => {
    const originalEnv = process.env.NEXT_PUBLIC_API_BASE_URL;
    
    // Use vi.stubEnv to properly mock environment variables in Vitest
    vi.stubEnv('NEXT_PUBLIC_API_BASE_URL', 'https://api.example.com');
    
    render(<PerformanceOptimizer />);
    
    expect(mockDnsPrefetch).toHaveBeenCalledWith('https://api.example.com');
    
    // Restore original env
    vi.unstubAllEnvs();
    if (originalEnv) {
      process.env.NEXT_PUBLIC_API_BASE_URL = originalEnv;
    }
  });

  it('should use localhost as fallback when NEXT_PUBLIC_API_BASE_URL is not set', () => {
    const originalEnv = process.env.NEXT_PUBLIC_API_BASE_URL;
    
    // Unset the env var
    vi.stubEnv('NEXT_PUBLIC_API_BASE_URL', undefined);
    
    render(<PerformanceOptimizer />);
    
    expect(mockDnsPrefetch).toHaveBeenCalledWith('http://localhost:8080');
    
    // Restore original env
    vi.unstubAllEnvs();
    if (originalEnv) {
      process.env.NEXT_PUBLIC_API_BASE_URL = originalEnv;
    }
  });

  it('should render nothing (null)', () => {
    const { container } = render(<PerformanceOptimizer />);
    expect(container.firstChild).toBeNull();
  });

  it('should only initialize once on mount', () => {
    const { rerender } = render(<PerformanceOptimizer />);
    
    expect(mockDnsPrefetch).toHaveBeenCalledTimes(1);
    expect(mockInitPerformanceOptimizations).toHaveBeenCalledTimes(1);
    
    rerender(<PerformanceOptimizer />);
    
    // Should still be called only once (useEffect with empty deps)
    expect(mockDnsPrefetch).toHaveBeenCalledTimes(1);
    expect(mockInitPerformanceOptimizations).toHaveBeenCalledTimes(1);
  });
});

