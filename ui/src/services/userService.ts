import { apiClient, ApiResponse } from '../utils/api';

export interface User {
  id: number;
  name: string;
  email: string;
  role: string;
  status: string;
  avatar?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  name: string;
  email: string;
  password: string;
  role?: string;
  status?: string;
  avatar?: string;
}

export interface LoginRequest {
  email: string;
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
      name: data.name, 
      email: data.email,
      hasPassword: !!data.password 
    });
    
    const response = await apiClient.post<User>(this.BASE_PATH, data);
    
    console.log('👤 API: POST /api/v1/users response', { 
      userId: response.data?.id,
      userName: response.data?.name,
      userEmail: response.data?.email
    });

    return response;
  }

  /**
   * Login user
   */
  static async loginUser(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
    console.log('👤 API: POST /api/v1/users/login', { 
      email: data.email,
      hasPassword: !!data.password 
    });
    
    const response = await apiClient.post<LoginResponse>(`${this.BASE_PATH}/login`, data);
    
    console.log('👤 API: POST /api/v1/users/login response', { 
      userId: response.data?.user?.id,
      userName: response.data?.user?.name,
      hasToken: !!response.data?.token
    });

    return response;
  }
} 