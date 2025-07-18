// API client utilities for making HTTP requests to the backend

export interface ApiResponse<T> {
  data?: T;
  message?: string;
  error?: string;
}

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public details?: any
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
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new ApiError(
          errorData.message || `HTTP error! status: ${response.status}`,
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

  async post<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
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
    return error;
  }
  
  return new ApiError(
    error instanceof Error ? error.message : 'An unexpected error occurred',
    0,
    error
  );
} 