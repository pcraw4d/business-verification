'use client';

import { useEffect } from 'react';
import { initPerformanceOptimizations, dnsPrefetch } from '@/lib/preload';

/**
 * Client-side performance optimization component
 * Initializes resource preloading and other performance optimizations
 */
export function PerformanceOptimizer() {
  useEffect(() => {
    // DNS prefetch for API
    const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
    dnsPrefetch(apiBaseUrl);
    
    // Initialize performance optimizations after component mounts
    initPerformanceOptimizations();
  }, []);

  return null;
}

