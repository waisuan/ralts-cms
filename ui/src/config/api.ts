// API configuration and environment settings

export const API_CONFIG = {
  BASE_URL: typeof window !== 'undefined'
    ? ''
    : process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
  
  // API endpoints
  ENDPOINTS: {
    MACHINES: '/api/v1/machines',
    MAINTENANCE: '/api/v1/maintenance',
    USERS: '/api/v1/users',
  },
  
  // Request timeout in milliseconds
  TIMEOUT: 10000,
  
  // Default pagination settings
  DEFAULT_PAGE_SIZE: 50,
  MAX_PAGE_SIZE: 100,
} as const;

// Environment detection
export const isDevelopment = process.env.NODE_ENV === 'development';
export const isProduction = process.env.NODE_ENV === 'production';

// API URL builder
export function buildApiUrl(endpoint: string): string {
  return `${API_CONFIG.BASE_URL}${endpoint}`;
}

// Validation helpers
export function validateApiResponse<T>(response: unknown): response is T {
  return Boolean(response && typeof response === 'object');
} 