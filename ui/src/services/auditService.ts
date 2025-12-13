import { apiClient, ApiResponse } from '../utils/api';

// Audit event interface matching the backend Event struct
export interface AuditEvent {
  id: string;
  user_id?: string;
  username?: string;
  action: string;
  resource_type: string;
  resource_id: string;
  details?: Record<string, unknown>;
  created_at: string;
}

// Response interface for listing audit events
export interface ListAuditEventsResponse {
  events: AuditEvent[];
  total_count: number;
  limit: number;
  offset: number;
}

// Filter parameters for querying audit events
export interface AuditFilters {
  from_date?: string;
  to_date?: string;
  resource_type?: string;
  action?: string;
}

// Audit action constants - must match backend constants
export const AUDIT_ACTION = {
  CREATED: 'created',
  UPDATED: 'updated',
  DELETED: 'deleted',
  VIEWED: 'viewed',
  LISTED: 'listed',
  LOGIN: 'login',
  LOGOUT: 'logout',
  PASSWORD_CHANGED: 'password_changed',
} as const;

export type AuditAction = typeof AUDIT_ACTION[keyof typeof AUDIT_ACTION];

// Resource type constants - must match backend constants
export const RESOURCE_TYPE = {
  MACHINE: 'machine',
  MAINTENANCE: 'maintenance',
  USER: 'user',
  SESSION: 'session',
  ATTACHMENT: 'attachment',
  AUDIT_EVENT: 'audit_event',
} as const;

export type ResourceType = typeof RESOURCE_TYPE[keyof typeof RESOURCE_TYPE];

export class AuditService {
  private static readonly BASE_PATH = '/api/v1/admin/audit';

  /**
   * Get a paginated list of audit events with optional filters
   * @param limit - Number of events to return (1-100, default: 50)
   * @param offset - Number of events to skip (default: 0)
   * @param filters - Optional filters for narrowing results
   */
  static async getEvents(
    limit: number = 50,
    offset: number = 0,
    filters?: AuditFilters
  ): Promise<ApiResponse<ListAuditEventsResponse>> {
    console.log('📋 API: GET /api/v1/admin/audit/events', { limit, offset, filters });

    const queryParams = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });

    // Add optional filters
    if (filters?.from_date) {
      queryParams.set('from_date', filters.from_date);
    }
    if (filters?.to_date) {
      queryParams.set('to_date', filters.to_date);
    }
    if (filters?.resource_type) {
      queryParams.set('resource_type', filters.resource_type);
    }
    if (filters?.action) {
      queryParams.set('action', filters.action);
    }

    const response = await apiClient.get<ListAuditEventsResponse>(
      `${this.BASE_PATH}/events?${queryParams.toString()}`
    );

    console.log('📋 API: GET /api/v1/admin/audit/events response', {
      eventCount: response.data?.events?.length,
      totalCount: response.data?.total_count,
      limit: response.data?.limit,
      offset: response.data?.offset,
    });

    return response;
  }

  /**
   * Get a human-readable label for an audit action
   * @param action - Audit action
   */
  static getActionLabel(action: string): string {
    switch (action) {
      case AUDIT_ACTION.CREATED:
        return 'Created';
      case AUDIT_ACTION.UPDATED:
        return 'Updated';
      case AUDIT_ACTION.DELETED:
        return 'Deleted';
      case AUDIT_ACTION.VIEWED:
        return 'Viewed';
      case AUDIT_ACTION.LISTED:
        return 'Listed';
      case AUDIT_ACTION.LOGIN:
        return 'Login';
      case AUDIT_ACTION.LOGOUT:
        return 'Logout';
      case AUDIT_ACTION.PASSWORD_CHANGED:
        return 'Password Changed';
      default:
        return action;
    }
  }

  /**
   * Get a CSS class for styling action badges
   * @param action - Audit action
   */
  static getActionClassName(action: string): string {
    switch (action) {
      case AUDIT_ACTION.CREATED:
        return 'bg-green-100 text-green-800';
      case AUDIT_ACTION.UPDATED:
        return 'bg-blue-100 text-blue-800';
      case AUDIT_ACTION.DELETED:
        return 'bg-red-100 text-red-800';
      case AUDIT_ACTION.VIEWED:
        return 'bg-gray-100 text-gray-800';
      case AUDIT_ACTION.LISTED:
        return 'bg-gray-100 text-gray-600';
      case AUDIT_ACTION.LOGIN:
        return 'bg-purple-100 text-purple-800';
      case AUDIT_ACTION.LOGOUT:
        return 'bg-orange-100 text-orange-800';
      case AUDIT_ACTION.PASSWORD_CHANGED:
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  }

  /**
   * Get a human-readable label for a resource type
   * @param resourceType - Resource type
   */
  static getResourceTypeLabel(resourceType: string): string {
    switch (resourceType) {
      case RESOURCE_TYPE.MACHINE:
        return 'Machine';
      case RESOURCE_TYPE.MAINTENANCE:
        return 'Maintenance';
      case RESOURCE_TYPE.USER:
        return 'User';
      case RESOURCE_TYPE.SESSION:
        return 'Session';
      case RESOURCE_TYPE.ATTACHMENT:
        return 'Attachment';
      case RESOURCE_TYPE.AUDIT_EVENT:
        return 'Audit Event';
      default:
        return resourceType;
    }
  }

  /**
   * Get a CSS class for styling resource type badges
   * @param resourceType - Resource type
   */
  static getResourceTypeClassName(resourceType: string): string {
    switch (resourceType) {
      case RESOURCE_TYPE.MACHINE:
        return 'bg-indigo-100 text-indigo-800';
      case RESOURCE_TYPE.MAINTENANCE:
        return 'bg-teal-100 text-teal-800';
      case RESOURCE_TYPE.USER:
        return 'bg-pink-100 text-pink-800';
      case RESOURCE_TYPE.SESSION:
        return 'bg-cyan-100 text-cyan-800';
      case RESOURCE_TYPE.ATTACHMENT:
        return 'bg-amber-100 text-amber-800';
      case RESOURCE_TYPE.AUDIT_EVENT:
        return 'bg-slate-100 text-slate-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  }

  /**
   * Format event timestamp for display
   * @param timestamp - ISO 8601 timestamp
   */
  static formatTimestamp(timestamp: string): string {
    const date = new Date(timestamp);
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  }

  /**
   * Format event timestamp as relative time (e.g., "5 minutes ago")
   * @param timestamp - ISO 8601 timestamp
   */
  static formatRelativeTime(timestamp: string): string {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSeconds = Math.floor(diffMs / 1000);
    const diffMinutes = Math.floor(diffSeconds / 60);
    const diffHours = Math.floor(diffMinutes / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffSeconds < 60) {
      return 'just now';
    } else if (diffMinutes < 60) {
      return `${diffMinutes} minute${diffMinutes === 1 ? '' : 's'} ago`;
    } else if (diffHours < 24) {
      return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
    } else if (diffDays < 7) {
      return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
    } else {
      return this.formatTimestamp(timestamp);
    }
  }
}

