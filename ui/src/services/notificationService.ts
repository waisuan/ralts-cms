import { apiClient, ApiResponse } from '../utils/api';
import { FlagStatus } from './flagService';

export type NotificationType = 'assigned' | 'flagged' | 'flag_resolved';

export interface Notification {
  id: string;
  user_id: number;
  type: NotificationType;
  machine_serial_number?: string | null;
  flag_id?: string | null;
  /** Current status of the referenced flag; absent for non-flag notifications. */
  flag_status?: FlagStatus | null;
  title: string;
  body?: string;
  actor_user_id?: number | null;
  actor_username?: string | null;
  read_at?: string | null;
  created_at: string;
}

export interface ListNotificationsResponse {
  notifications: Notification[];
  count: number;
  unread_count: number;
  limit: number;
  offset: number;
}

export interface UnreadCountResponse {
  unread_count: number;
}

export interface ListNotificationsParams {
  limit?: number;
  offset?: number;
  unread_only?: boolean;
}

export class NotificationService {
  private static readonly BASE_PATH = '/api/v1/notifications';

  static async list(
    params: ListNotificationsParams = {}
  ): Promise<ApiResponse<ListNotificationsResponse>> {
    const query: Record<string, string> = {};
    if (params.limit != null) query.limit = String(params.limit);
    if (params.offset != null) query.offset = String(params.offset);
    if (params.unread_only) query.unread_only = 'true';
    return apiClient.get<ListNotificationsResponse>(this.BASE_PATH, query);
  }

  static async unreadCount(): Promise<ApiResponse<UnreadCountResponse>> {
    return apiClient.get<UnreadCountResponse>(`${this.BASE_PATH}/unread-count`);
  }

  static async markRead(id: string): Promise<ApiResponse<void>> {
    return apiClient.post<void>(`${this.BASE_PATH}/${encodeURIComponent(id)}/read`, {});
  }

  static async markAllRead(): Promise<ApiResponse<{ updated: number }>> {
    return apiClient.post<{ updated: number }>(`${this.BASE_PATH}/read-all`, {});
  }
}
