import { apiClient, ApiResponse, ApiError } from '../utils/api';

export interface User {
  id: number;
  username: string;
  email: string;
  role: string;
  approved: boolean;
  status?: string;
  avatar?: string;
  created_at: string;
  updated_at?: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  role?: string;
  approved?: boolean;
  status?: string;
  avatar?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
}

export interface UpdatePasswordRequest {
  current_password: string;
  new_password: string;
}

export interface UpdatePasswordResponse {
  message: string;
}

export class UserService {
  private static readonly BASE_PATH = '/api/v1/users';

  /**
   * Register a new user
   */
  static async registerUser(data: CreateUserRequest): Promise<ApiResponse<User>> {
    console.log('👤 API: POST /api/v1/users', { 
      username: data.username, 
      email: data.email,
      hasPassword: !!data.password 
    });
    
    try {
      const response = await apiClient.post<User>(this.BASE_PATH, data);
      
      console.log('👤 API: POST /api/v1/users response', { 
        userId: response.data?.id,
        username: response.data?.username,
        userEmail: response.data?.email
      });

      return response;
    } catch (error) {
      // Handle user-specific 409 conflict errors
      if (error instanceof ApiError && error.status === 409) {
        throw new ApiError(
          'This user already exists. Please try different credentials.',
          error.status,
          error.details
        );
      }
      // Re-throw other errors unchanged
      throw error;
    }
  }

  /**
   * Login user
   */
  static async loginUser(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
    console.log('👤 API: POST /api/v1/users/login', { 
      username: data.username,
      hasPassword: !!data.password 
    });
    
    try {
      const response = await apiClient.post<LoginResponse>(`${this.BASE_PATH}/login`, data);
      
      console.log('👤 API: POST /api/v1/users/login response', { 
        userId: response.data?.user?.id,
        username: response.data?.user?.username,
        hasToken: !!response.data?.token
      });

      return response;
    } catch (error) {
      // Handle user-specific login errors
      if (error instanceof ApiError) {
        if (error.status === 401) {
          throw new ApiError(
            'Invalid username or password.',
            error.status,
            error.details
          );
        }
        if (error.status === 403) {
          throw new ApiError(
            'Your account is not currently approved for the system. Please contact an administrator to activate your account.',
            error.status,
            error.details
          );
        }
      }
      // Re-throw other errors unchanged
      throw error;
    }
  }

  /**
   * Update user password
   */
  static async updatePassword(data: UpdatePasswordRequest): Promise<ApiResponse<UpdatePasswordResponse>> {
    console.log('🔐 API: PUT /api/v1/users/password');
    
    try {
      const response = await apiClient.put<UpdatePasswordResponse>(`${this.BASE_PATH}/password`, data);
      
      console.log('🔐 API: PUT /api/v1/users/password response', { 
        message: response.data?.message
      });

      return response;
    } catch (error) {
      // Handle password update specific errors
      if (error instanceof ApiError) {
        if (error.status === 401) {
          throw new ApiError(
            'Current password is incorrect.',
            error.status,
            error.details
          );
        }
        if (error.status === 400) {
          throw new ApiError(
            error.message || 'Invalid password format. Password must be at least 6 characters.',
            error.status,
            error.details
          );
        }
      }
      // Re-throw other errors unchanged
      throw error;
    }
  }
}