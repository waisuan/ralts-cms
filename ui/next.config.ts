import type { NextConfig } from 'next';

const apiTarget =
  process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

const nextConfig: NextConfig = {
  transpilePackages: ['nuqs'],
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'images.unsplash.com',
        port: '',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 'ui-avatars.com',
        port: '',
        pathname: '/**',
      },
    ],
  },
  poweredByHeader: false,
  // The browser reaches this dev server through the WSL localhost relay, so dev
  // resource requests arrive as localhost or 127.0.0.1 and Next 16 blocks them
  // (including the HMR endpoint) unless the origin is listed here.
  allowedDevOrigins: ['localhost', '127.0.0.1'],
  async rewrites() {
    return [
      { source: '/api/:path*', destination: `${apiTarget}/api/:path*` },
    ];
  },
};

export default nextConfig;
