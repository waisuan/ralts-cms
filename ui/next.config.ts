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
  async rewrites() {
    return [
      { source: '/api/:path*', destination: `${apiTarget}/api/:path*` },
    ];
  },
};

export default nextConfig;
