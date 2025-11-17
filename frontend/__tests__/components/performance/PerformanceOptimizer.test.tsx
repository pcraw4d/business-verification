import { render } from '@testing-library/react';
import { PerformanceOptimizer } from '@/components/performance/PerformanceOptimizer';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock the preload module
vi.mock('@/lib/preload', () => ({
  initPerformanceOptimizations: vi.fn(),
  dnsPrefetch: vi.fn(),
}));

describe('PerformanceOptimizer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize performance optimizations on mount', () => {
    const { initPerformanceOptimizations, dnsPrefetch } = require('@/lib/preload');
    
    render(<PerformanceOptimizer />);
    
    expect(dnsPrefetch).toHaveBeenCalled();
    expect(initPerformanceOptimizations).toHaveBeenCalled();
  });

  it('should use NEXT_PUBLIC_API_BASE_URL for DNS prefetch', () => {
    const { dnsPrefetch } = require('@/lib/preload');
    const originalEnv = process.env.NEXT_PUBLIC_API_BASE_URL;
    
    process.env.NEXT_PUBLIC_API_BASE_URL = 'https://api.example.com';
    
    render(<PerformanceOptimizer />);
    
    expect(dnsPrefetch).toHaveBeenCalledWith('https://api.example.com');
    
    // Restore original env
    if (originalEnv) {
      process.env.NEXT_PUBLIC_API_BASE_URL = originalEnv;
    } else {
      delete process.env.NEXT_PUBLIC_API_BASE_URL;
    }
  });

  it('should use localhost as fallback when NEXT_PUBLIC_API_BASE_URL is not set', () => {
    const { dnsPrefetch } = require('@/lib/preload');
    const originalEnv = process.env.NEXT_PUBLIC_API_BASE_URL;
    
    delete process.env.NEXT_PUBLIC_API_BASE_URL;
    
    render(<PerformanceOptimizer />);
    
    expect(dnsPrefetch).toHaveBeenCalledWith('http://localhost:8080');
    
    // Restore original env
    if (originalEnv) {
      process.env.NEXT_PUBLIC_API_BASE_URL = originalEnv;
    }
  });

  it('should render nothing (null)', () => {
    const { container } = render(<PerformanceOptimizer />);
    expect(container.firstChild).toBeNull();
  });

  it('should only initialize once on mount', () => {
    const { initPerformanceOptimizations, dnsPrefetch } = require('@/lib/preload');
    
    const { rerender } = render(<PerformanceOptimizer />);
    
    expect(dnsPrefetch).toHaveBeenCalledTimes(1);
    expect(initPerformanceOptimizations).toHaveBeenCalledTimes(1);
    
    rerender(<PerformanceOptimizer />);
    
    // Should still be called only once (useEffect with empty deps)
    expect(dnsPrefetch).toHaveBeenCalledTimes(1);
    expect(initPerformanceOptimizations).toHaveBeenCalledTimes(1);
  });
});

