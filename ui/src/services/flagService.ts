import { apiClient, ApiResponse } from '../utils/api';

export type FlagReason = 'requires_attention' | 'missing_values' | 'other';
export type FlagStatus = 'open' | 'resolved';

export interface Flag {
  id: string;
  machine_serial_number: string;
  reason: FlagReason;
  ppm_status?: string;
  note?: string;
  status: FlagStatus;
  created_by?: number | null;
  created_by_username?: string | null;
  created_at: string;
  resolved_by?: number | null;
  resolved_by_username?: string | null;
  resolved_at?: string | null;
  /** What the resolver said they did about the flag. Optional. */
  resolution_note?: string;
}

export interface ListFlagsResponse {
  flags: Flag[];
}

/** Status filter accepted by the flagged-records listing. */
export type FlagStatusFilter = 'open' | 'resolved' | 'all';

/**
 * Which flags the listing covered: every machine (admins) or only those
 * assigned to the caller (everyone else). Decided by the API.
 */
export type FlagScope = 'all' | 'assigned';

export interface ListAllFlagsResponse {
  flags: Flag[];
  count: number;
  limit: number;
  offset: number;
  scope: FlagScope;
}

export interface CreateFlagRequest {
  reason: 'missing_values' | 'other';
  note?: string;
}

/** Mirrors flags.MaxNoteLength on the API, which rejects anything longer. */
export const MAX_FLAG_NOTE_LENGTH = 500;

export interface OpenFlagsBatchResponse {
  flags: Record<string, Flag[]>;
}

export class FlagService {
  /**
   * Batch-fetch open flags for a list of machine serial numbers. Used to
   * render flag badges on the machines table without issuing one request per
   * row.
   */
  static async listOpenByMachine(
    serialNumbers: string[]
  ): Promise<ApiResponse<OpenFlagsBatchResponse>> {
    if (serialNumbers.length === 0) {
      return { data: { flags: {} } };
    }
    return apiClient.get<OpenFlagsBatchResponse>(`/api/v1/machines/flags/open-by-machine`, {
      serials: serialNumbers.join(','),
    });
  }

  /**
   * List flags for the flagged-records page. The API returns every machine's
   * flags for admins and only the caller's assigned machines otherwise, and
   * reports which of the two in `scope`.
   */
  static async listAll(
    options: { status?: FlagStatusFilter; limit?: number; offset?: number } = {}
  ): Promise<ApiResponse<ListAllFlagsResponse>> {
    const params: Record<string, string> = {};
    if (options.status) params.status = options.status;
    if (options.limit != null) params.limit = String(options.limit);
    if (options.offset != null) params.offset = String(options.offset);
    return apiClient.get<ListAllFlagsResponse>('/api/v1/flags', params);
  }

  static async listForMachine(
    serialNumber: string,
    includeResolved: boolean = false
  ): Promise<ApiResponse<ListFlagsResponse>> {
    const encoded = encodeURIComponent(serialNumber);
    const params: Record<string, string> = {};
    if (includeResolved) params.include_resolved = 'true';
    return apiClient.get<ListFlagsResponse>(`/api/v1/machines/${encoded}/flags`, params);
  }

  static async create(serialNumber: string, data: CreateFlagRequest): Promise<ApiResponse<Flag>> {
    const encoded = encodeURIComponent(serialNumber);
    return apiClient.post<Flag>(`/api/v1/machines/${encoded}/flags`, data);
  }

  /**
   * Resolve a flag, optionally recording what was done about it. The note is
   * bounded by MAX_FLAG_NOTE_LENGTH, as on creation.
   */
  static async resolve(
    serialNumber: string,
    flagId: string,
    note?: string
  ): Promise<ApiResponse<void>> {
    const encoded = encodeURIComponent(serialNumber);
    return apiClient.post<void>(
      `/api/v1/machines/${encoded}/flags/${encodeURIComponent(flagId)}/resolve`,
      { note: note?.trim() ?? '' }
    );
  }

  static reasonLabel(reason: FlagReason): string {
    switch (reason) {
      case 'requires_attention':
        return 'Requires attention';
      case 'missing_values':
        return 'Missing values';
      case 'other':
        return 'Other';
      default:
        return reason;
    }
  }
}
