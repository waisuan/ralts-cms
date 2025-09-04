import { apiClient, ApiResponse } from '../utils/api';

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
    
    const response = await apiClient.post<User>(this.BASE_PATH, data);
    
    console.log('👤 API: POST /api/v1/users response', { 
      userId: response.data?.id,
      username: response.data?.username,
      userEmail: response.data?.email
    });

    return response;
  }

  /**
   * Login user
   */
  static async loginUser(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
    console.log('👤 API: POST /api/v1/users/login', { 
      username: data.username,
      hasPassword: !!data.password 
    });
    
    const response = await apiClient.post<LoginResponse>(`${this.BASE_PATH}/login`, data);
    
    console.log('👤 API: POST /api/v1/users/login response', { 
      userId: response.data?.user?.id,
      username: response.data?.user?.username,
      hasToken: !!response.data?.token
    });

    return response;
  }
} 