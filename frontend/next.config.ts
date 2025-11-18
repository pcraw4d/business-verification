import type { NextConfig } from "next";

// Get API base URL for rewrites (server-side only, so we can use process.env directly)
const getApiBaseUrlForRewrites = () => {
  return process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
};

const nextConfig: NextConfig = {
  // API proxy for development
  async rewrites() {
    const apiBaseUrl = getApiBaseUrlForRewrites();
    return [
      {
        source: '/api/:path*',
        destination: `${apiBaseUrl}/api/:path*`,
      },
    ];
  },
  
  // Production optimizations
  reactStrictMode: true,
  
  // Image optimization
  images: {
    unoptimized: process.env.NEXT_PUBLIC_UNOPTIMIZED_IMAGES === 'true',
    formats: ['image/avif', 'image/webp'],
  },
  
  // Compiler optimizations
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production' ? {
      exclude: ['error', 'warn'],
    } : false,
  },
  
  // Output configuration
  // Use 'standalone' for server-side rendering with Go service
  // Use 'export' for static export (if serving static files from Go service)
  // Note: Dynamic routes require the route template to be generated
  // The Go service will serve the template and Next.js will handle client-side routing
  output: process.env.NEXT_PUBLIC_STATIC_EXPORT === 'true' ? 'export' : undefined,
  
  // Ensure dynamic routes are generated as templates
  // This allows the Go service to serve the template for client-side routing
  generateBuildId: async () => {
    // Use a consistent build ID for better caching
    return process.env.BUILD_ID || 'build-' + Date.now();
  },
  
  // Experimental features for better performance
  experimental: {
    optimizePackageImports: ['lucide-react', 'recharts', 'd3'],
  },
  
  // Turbopack configuration (Next.js 16+)
  turbopack: {},
  
  // Webpack optimizations (for non-Turbopack builds)
  webpack: (config, { isServer }) => {
    // Optimize bundle size
    if (!isServer) {
      config.resolve.fallback = {
        ...config.resolve.fallback,
        fs: false,
      };
      
      // Code splitting optimizations
      config.optimization = {
        ...config.optimization,
        splitChunks: {
          chunks: 'all',
          maxInitialRequests: 25,
          minSize: 20000,
          cacheGroups: {
            default: false,
            vendors: false,
            // Framework chunk (React, Next.js) - Priority 40
            framework: {
              name: 'framework',
              test: /[\\/]node_modules[\\/](react|react-dom|next|scheduler)[\\/]/,
              chunks: 'all',
              priority: 40,
              enforce: true,
            },
            // Separate chunk for chart libraries - Priority 30
            charts: {
              name: 'charts',
              test: /[\\/]node_modules[\\/](recharts|d3|d3-)[\\/]/,
              chunks: 'all',
              priority: 30,
              enforce: true,
            },
            // Export libraries - Priority 30
            exportLibs: {
              name: 'export-libs',
              test: /[\\/]node_modules[\\/](xlsx|jspdf|html2canvas)[\\/]/,
              chunks: 'all',
              priority: 30,
              enforce: true,
            },
            // Radix UI components - Priority 25
            radix: {
              name: 'radix',
              test: /[\\/]node_modules[\\/]@radix-ui[\\/]/,
              chunks: 'all',
              priority: 25,
            },
            // Vendor chunk - Priority 20 (minSize: 20KB)
            vendor: {
              name: 'vendor',
              chunks: 'all',
              test: /[\\/]node_modules[\\/]/,
              priority: 20,
              minSize: 20000,
            },
            // Common chunk - Priority 10 (minSize: 20KB, minChunks: 2)
            common: {
              name: 'common',
              minChunks: 2,
              chunks: 'all',
              priority: 10,
              reuseExistingChunk: true,
              minSize: 20000,
            },
          },
        },
      };
    }
    
    return config;
  },
  
  // Headers for caching and performance
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'X-DNS-Prefetch-Control',
            value: 'on',
          },
          {
            key: 'X-Frame-Options',
            value: 'SAMEORIGIN',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
        ],
      },
      {
        source: '/_next/static/:path*',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable',
          },
        ],
      },
      {
        source: '/_next/image',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable',
          },
        ],
      },
    ];
  },
};

export default nextConfig;
