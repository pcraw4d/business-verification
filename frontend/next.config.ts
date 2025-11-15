import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // API proxy for development
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: process.env.NEXT_PUBLIC_API_BASE_URL 
          ? `${process.env.NEXT_PUBLIC_API_BASE_URL}/api/:path*`
          : 'http://localhost:8080/api/:path*',
      },
    ];
  },
  // Output configuration for static export (if needed)
  // output: 'export', // Uncomment if serving static files from Go service
};

export default nextConfig;
