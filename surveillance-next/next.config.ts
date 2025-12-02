import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.API_URL || 'http://api-server:8080'}/api/:path*`
      }
    ]
  }
}

export default nextConfig