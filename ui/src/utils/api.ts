// API client utilities for making HTTP requests to the backend
import { redirectToLogin, isAuthError } from './auth';

export interface ApiResponse<T> {
  data?: T;
  message?: string;
  error?: string;
}

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public details?: unknown
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export class ApiClient {
  private baseURL: string;

  constructor(baseURL?: string) {
    this.baseURL = baseURL || process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;
    
    // Get JWT token from localStorage if available
    const token = typeof window !== 'undefined' ? localStorage.getItem('ralts_token') : null;
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` }),
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        
        // Provide user-friendly messages for specific status codes
        let errorMessage = errorData.message;
        if (!errorMessage) {
          switch (response.status) {
            case 501:
              errorMessage = 'This feature is not yet implemented on the server.';
              break;
            case 404:
              errorMessage = 'The requested resource was not found.';
              break;
            case 400:
              errorMessage = 'Invalid request. Please check your input.';
              break;
            case 401:
              errorMessage = 'Authentication required. Please log in again.';
              break;
            case 403:
              errorMessage = 'You do not have permission to perform this action.';
              break;
            case 500:
              errorMessage = 'Server error. Please try again later.';
              break;
            default:
              errorMessage = `HTTP error! status: ${response.status}`;
          }
        }
        
        throw new ApiError(
          errorMessage,
          response.status,
          errorData
        );
      }

      // Handle responses with no body (like 204 No Content)
      if (response.status === 204 || response.headers.get('content-length') === '0') {
        return { data: undefined as T };
      }

      // For DELETE operations, don't try to parse JSON if status is 2xx
      if (options.method === 'DELETE' && response.status >= 200 && response.status < 300) {
        return { data: undefined as T };
      }

      // Try to parse JSON response
      const responseData = await response.json().catch((jsonError) => {
        console.warn('Failed to parse JSON response:', jsonError);
        return null;
      });
      
      if (responseData === null) {
        // No JSON body, return empty response
        return { data: undefined as T };
      }
      
      // Handle both wrapped and unwrapped responses
      if (responseData.data !== undefined) {
        // Response is wrapped in a data property
        return responseData;
      } else {
        // Response is not wrapped, return it directly
        return { data: responseData };
      }
    } catch (error) {
      if (error instanceof ApiError) {
        // Check if this is a 401 authentication error
        if (isAuthError(error)) {
          // Don't redirect for login-related endpoints since user is already on login page
          const isLoginEndpoint = endpoint.includes('/login') || endpoint.includes('/users');
          if (!isLoginEndpoint) {
            console.log('🔐 Authentication error detected, redirecting to login');
            redirectToLogin();
            // Don't throw the error since we're redirecting
            throw new ApiError('Redirecting to login...', 401);
          }
        }
        throw error;
      }
      
      // Network or other errors
      throw new ApiError(
        error instanceof Error ? error.message : 'Network error occurred',
        0,
        error
      );
    }
  }

  async get<T>(endpoint: string, params?: Record<string, string>): Promise<ApiResponse<T>> {
    const url = new URL(`${this.baseURL}${endpoint}`);
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        url.searchParams.append(key, value);
      });
    }
    
    return this.request<T>(url.pathname + url.search);
  }

  async post<T>(endpoint: string, data?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'DELETE',
    });
  }
}

// Global API client instance
export const apiClient = new ApiClient();

// Helper function to handle API errors
export function handleApiError(error: unknown): ApiError {
  if (error instanceof ApiError) {
    // Check if this is a 401 authentication error and redirect
    if (isAuthError(error)) {
      // For handleApiError, we don't have endpoint context, so we'll be more conservative
      // Only redirect if we're not on a login-related page
      const currentPath = typeof window !== 'undefined' ? window.location.pathname : '';
      const isLoginPage = currentPath === '/' || currentPath === '/login' || currentPath === '/register';
      
      if (!isLoginPage) {
        console.log('🔐 Authentication error detected in handleApiError, redirecting to login');
        redirectToLogin();
        // Return a generic error since we're redirecting
        return new ApiError('Redirecting to login...', 401);
      }
    }
    return error;
  }
  
  return new ApiError(
    error instanceof Error ? error.message : 'An unexpected error occurred',
    0,
    error
  );
} 