import { apiClient, ApiResponse } from '../utils/api';
import { User } from './userService';

export interface ListUsersResponse {
  users: User[];
  total_count: number;
  limit: number;
  offset: number;
}

export interface UpdateUserStatusRequest {
  status: string;
}

export interface UpdateUserStatusResponse {
  message: string;
  user_id: number;
  status: string;
}

export interface BulkUpdateStatusRequest {
  user_ids: number[];
  status: string;
}

export interface BulkUpdateStatusResponse {
  message: string;
  updated_count: number;
  status: string;
  affected_users: number[];
}

// User status constants - must match backend constants
export const USER_STATUS = {
  PENDING_APPROVAL: 'pending_approval',
  APPROVED: 'approved',
  SUSPENDED: 'suspended',
  INACTIVE: 'inactive',
} as const;

export type UserStatus = typeof USER_STATUS[keyof typeof USER_STATUS];

// User role constants - must match backend constants
export const USER_ROLE = {
  ADMIN: 'ADMIN',
  NON_ADMIN: 'NON_ADMIN',
} as const;

export type UserRole = typeof USER_ROLE[keyof typeof USER_ROLE];

export class AdminUserService {
  private static readonly BASE_PATH = '/api/v1/admin/users';

  /**
   * Get a paginated list of all users
   * @param limit - Number of users to return (1-100, default: 50)
   * @param offset - Number of users to skip (default: 0)
   */
  static async getUsers(
    limit: number = 50, 
    offset: number = 0
  ): Promise<ApiResponse<ListUsersResponse>> {
    console.log('👥 API: GET /api/v1/admin/users', { limit, offset });
    
    const queryParams = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });
    
    const response = await apiClient.get<ListUsersResponse>(
      `${this.BASE_PATH}?${queryParams.toString()}`
    );
    
    console.log('👥 API: GET /api/v1/admin/users response', {
      userCount: response.data?.users?.length,
      totalCount: response.data?.total_count,
      limit: response.data?.limit,
      offset: response.data?.offset,
    });

    return response;
  }

  /**
   * Update the status of a single user
   * @param userId - ID of the user to update
   * @param status - New status for the user
   */
  static async updateUserStatus(
    userId: number, 
    status: UserStatus
  ): Promise<ApiResponse<UpdateUserStatusResponse>> {
    console.log('👤 API: PUT /api/v1/admin/users/{id}/status', { userId, status });
    
    const requestData: UpdateUserStatusRequest = { status };
    
    const response = await apiClient.put<UpdateUserStatusResponse>(
      `${this.BASE_PATH}/${userId}/status`,
      requestData
    );
    
    console.log('👤 API: PUT /api/v1/admin/users/{id}/status response', {
      userId: response.data?.user_id,
      status: response.data?.status,
      message: response.data?.message,
    });

    return response;
  }

  /**
   * Update the status of multiple users in bulk
   * @param userIds - Array of user IDs to update (max 100)
   * @param status - New status for all users
   */
  static async updateMultipleUserStatuses(
    userIds: number[], 
    status: UserStatus
  ): Promise<ApiResponse<BulkUpdateStatusResponse>> {
    console.log('👥 API: PUT /api/v1/admin/users/bulk-status', { 
      userCount: userIds.length, 
      status,
      userIds: userIds.slice(0, 5) // Log first 5 IDs for debugging
    });
    
    if (userIds.length === 0) {
      throw new Error('User IDs list cannot be empty');
    }
    
    if (userIds.length > 100) {
      throw new Error('Cannot update more than 100 users at once');
    }
    
    const requestData: BulkUpdateStatusRequest = {
      user_ids: userIds,
      status,
    };
    
    const response = await apiClient.put<BulkUpdateStatusResponse>(
      `${this.BASE_PATH}/bulk-status`,
      requestData
    );
    
    console.log('👥 API: PUT /api/v1/admin/users/bulk-status response', {
      updatedCount: response.data?.updated_count,
      status: response.data?.status,
      affectedUsers: response.data?.affected_users?.length,
      message: response.data?.message,
    });

    return response;
  }

  /**
   * Validate if a status value is valid
   * @param status - Status to validate
   */
  static isValidStatus(status: string): status is UserStatus {
    return Object.values(USER_STATUS).includes(status as UserStatus);
  }

  /**
   * Get a human-readable label for a user status
   * @param status - User status
   */
  static getStatusLabel(status: UserStatus): string {
    switch (status) {
      case USER_STATUS.PENDING_APPROVAL:
        return 'Pending Approval';
      case USER_STATUS.APPROVED:
        return 'Approved';
      case USER_STATUS.SUSPENDED:
        return 'Suspended';
      case USER_STATUS.INACTIVE:
        return 'Inactive';
      default:
        return 'Unknown';
    }
  }

  /**
   * Get a CSS class name for styling user status badges
   * @param status - User status
   */
  static getStatusClassName(status: UserStatus): string {
    switch (status) {
      case USER_STATUS.PENDING_APPROVAL:
        return 'status-pending';
      case USER_STATUS.APPROVED:
        return 'status-approved';
      case USER_STATUS.SUSPENDED:
        return 'status-suspended';
      case USER_STATUS.INACTIVE:
        return 'status-inactive';
      default:
        return 'status-unknown';
    }
  }

  /**
   * Get a human-readable label for a user role
   * @param role - User role
   */
  static getRoleLabel(role: UserRole): string {
    switch (role) {
      case USER_ROLE.ADMIN:
        return 'Administrator';
      case USER_ROLE.NON_ADMIN:
        return 'User';
      default:
        return 'Unknown';
    }
  }
}
