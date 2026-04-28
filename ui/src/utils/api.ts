// API client utilities for making HTTP requests to the backend
import { redirectToLogin, isAuthError } from './auth';
import {
  getAccessToken,
  getRefreshToken,
  isPublicAuthPath,
  refreshAccessToken,
  shouldSuppressAuthRedirectOn401,
} from './tokens';

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

/** Parse JSON error body, else response text, for a failed fetch Response. */
async function readHttpErrorPayload(response: Response): Promise<{
  errorData: Record<string, unknown>;
  fromBody: string;
}> {
  let errorData: Record<string, unknown> = {};
  let fromBody = '';
  try {
    errorData = await response.json();
    fromBody = (errorData.message as string) || '';
  } catch {
    try {
      fromBody = await response.text();
    } catch {
      fromBody = '';
    }
  }
  return { errorData, fromBody };
}

function userFacingHttpMessage(status: number, fromBody: string): string {
  if (fromBody) {
    return fromBody;
  }
  switch (status) {
    case 501:
      return 'This feature is not yet implemented on the server.';
    case 404:
      return 'The requested resource was not found.';
    case 400:
      return 'Invalid request. Please check your input.';
    case 401:
      return 'Authentication required. Please log in again.';
    case 403:
      return 'You do not have permission to perform this action.';
    case 409:
      return 'The resource already exists or there is a conflict.';
    case 500:
      return 'Server error. Please try again later.';
    default:
      return `HTTP error! status: ${status}`;
  }
}

export class ApiClient {
  private baseURL: string;

  constructor(baseURL?: string) {
    this.baseURL = baseURL || (typeof window !== 'undefined'
      ? ''
      : process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080');
  }

  private async fetchWithAuth(
    fullUrl: string,
    init: RequestInit,
    endpoint: string
  ): Promise<Response> {
    const run = (bearer: string | null) => {
      const headers = new Headers();
      const src = init.headers;
      if (src instanceof Headers) {
        src.forEach((v, k) => {
          headers.set(k, v);
        });
      } else if (src && typeof src === 'object' && !Array.isArray(src)) {
        for (const [k, v] of Object.entries(src as Record<string, string>)) {
          if (typeof v === 'string') {
            headers.set(k, v);
          }
        }
      } else if (Array.isArray(src)) {
        for (const [k, v] of src) {
          headers.set(k, v);
        }
      }
      if (bearer) {
        headers.set('Authorization', `Bearer ${bearer}`);
      }
      return fetch(fullUrl, { ...init, headers });
    };
    let res = await run(getAccessToken());
    if (
      res.status === 401 &&
      !isPublicAuthPath(endpoint) &&
      getRefreshToken()
    ) {
      const next = await refreshAccessToken(() => this.baseURL);
      if (next) {
        res = await run(next);
      }
    }
    return res;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;

    const access = getAccessToken();
    const init: RequestInit = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(access && { 'Authorization': `Bearer ${access}` }),
        ...options.headers,
      },
    };

    try {
      const response = await this.fetchWithAuth(url, init, endpoint);
      
      if (!response.ok) {
        const { errorData, fromBody } = await readHttpErrorPayload(response);
        throw new ApiError(
          userFacingHttpMessage(response.status, fromBody),
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
          if (!shouldSuppressAuthRedirectOn401(endpoint)) {
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
    let path = endpoint;
    if (params && Object.keys(params).length > 0) {
      const qs = new URLSearchParams(params).toString();
      path += `${endpoint.includes('?') ? '&' : '?'}${qs}`;
    }
    return this.request<T>(path);
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

  private async formDataRequest<T>(
    method: 'POST' | 'PUT',
    endpoint: string,
    formData: FormData
  ): Promise<ApiResponse<T>> {
    const access = getAccessToken();
    const init: RequestInit = {
      method,
      // Omit Content-Type so the browser sets multipart boundary.
      headers: {
        ...(access && { 'Authorization': `Bearer ${access}` }),
      },
      body: formData,
    };
    const url = `${this.baseURL}${endpoint}`;

    try {
      const response = await this.fetchWithAuth(url, init, endpoint);

      if (!response.ok) {
        if (response.status === 401 && !shouldSuppressAuthRedirectOn401(endpoint)) {
          console.log('🔐 Authentication error detected, redirecting to login');
          redirectToLogin();
          throw new ApiError('Redirecting to login...', 401);
        }
        const { errorData, fromBody } = await readHttpErrorPayload(response);
        throw new ApiError(
          userFacingHttpMessage(response.status, fromBody),
          response.status,
          errorData
        );
      }

      if (response.status === 204 || response.headers.get('content-length') === '0') {
        return { data: undefined as T };
      }
      if (response.status >= 200 && response.status < 300) {
        const contentType = response.headers.get('content-type');
        if (!contentType || !contentType.includes('application/json')) {
          return { data: undefined as T };
        }
      }

      const responseData = await response.json().catch(() => null);
      if (responseData === null) {
        return { data: undefined as T };
      }
      if (responseData.data !== undefined) {
        return responseData;
      }
      return { data: responseData };
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError(
        error instanceof Error ? error.message : 'Network error occurred',
        0,
        error
      );
    }
  }

  async postFormData<T>(endpoint: string, formData: FormData): Promise<ApiResponse<T>> {
    return this.formDataRequest<T>('POST', endpoint, formData);
  }

  async putFormData<T>(endpoint: string, formData: FormData): Promise<ApiResponse<T>> {
    return this.formDataRequest<T>('PUT', endpoint, formData);
  }

  async getBlob(endpoint: string, params?: Record<string, string>): Promise<Blob> {
    let path = endpoint;
    if (params && Object.keys(params).length > 0) {
      const qs = new URLSearchParams(params).toString();
      path += `${endpoint.includes('?') ? '&' : '?'}${qs}`;
    }

    const url = `${this.baseURL}${path}`;

    const access = getAccessToken();
    const init: RequestInit = {
      headers: {
        ...(access && { 'Authorization': `Bearer ${access}` }),
      },
    };

    try {
      const response = await this.fetchWithAuth(url, init, path);

      if (!response.ok) {
        if (response.status === 401 && !shouldSuppressAuthRedirectOn401(path)) {
          redirectToLogin();
          throw new ApiError('Redirecting to login...', 401);
        }
        const errorText = await response.text();
        throw new ApiError(
          `Request failed: ${errorText || response.statusText}`,
          response.status,
          errorText
        );
      }

      return await response.blob();
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }

      throw new ApiError(
        error instanceof Error ? error.message : 'Network error occurred',
        0,
        error
      );
    }
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