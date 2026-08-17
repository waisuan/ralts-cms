import { apiClient, ApiResponse } from '../utils/api';

export interface DirectoryEntry {
  id: number;
  username: string;
  email: string;
}

export interface DirectoryResponse {
  users: DirectoryEntry[];
}

/**
 * Slim, non-admin-accessible listing of approved+active users used to power
 * assignee pickers. Backed by GET /api/v1/users/directory.
 */
export class UserDirectoryService {
  private static readonly BASE_PATH = '/api/v1/users/directory';

  static async list(): Promise<ApiResponse<DirectoryResponse>> {
    return apiClient.get<DirectoryResponse>(this.BASE_PATH);
  }
}
