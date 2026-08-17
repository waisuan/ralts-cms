import { apiClient, ApiResponse } from '../utils/api';
import { getAccessToken, getRefreshToken, refreshAccessToken } from '../utils/tokens';
import { Machine } from '../types/machine';

function apiBase(): string {
  return typeof window !== 'undefined'
    ? ''
    : process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
}

async function fetchWithTokenRefresh(path: string, init: RequestInit): Promise<Response> {
  const run = (access: string | null) => {
    const headers = new Headers(init.headers);
    if (access) {
      headers.set('Authorization', `Bearer ${access}`);
    } else {
      headers.delete('Authorization');
    }
    return fetch(path, { ...init, headers });
  };
  let res = await run(getAccessToken());
  if (res.status === 401 && getRefreshToken()) {
    const next = await refreshAccessToken(apiBase);
    if (next) {
      res = await run(next);
    }
  }
  return res;
}

export interface MachineFilters {
  limit?: number;
  offset?: number;
  sort?: string;
  ppm_status_filter?: string;
  q?: string; // Search query parameter
  // Date range filters (format: YYYY-MM-DD)
  ppm_date_from?: string;
  ppm_date_to?: string;
  tnc_date_from?: string;
  tnc_date_to?: string;
}

export interface MachineListResponse {
  machines: Machine[];
  overdue_count: number;
  due_count: number;
  almost_due_count: number;
  count: number;
  offset: number;
  limit: number;
  sort: string;
}

export interface CreateMachineRequest {
  serial_number: string;
  customer: string;
  state: string;
  account_type: string;
  model: string;
  status: string;
  brand: string;
  district: string;
  // An assignee is either a registered user (send assigned_user_id, and the
  // server derives person_in_charge from their username) or a free-text name
  // (send person_in_charge only; such an assignee gets no notifications).
  assigned_user_id?: number | null;
  person_in_charge?: string;
  reported_by: string;
  additional_notes?: string;
  attachment?: string;
  ppm_status: string;
  tnc_date: string;
  ppm_date: string;
}

export interface UpdateMachineRequest extends Partial<CreateMachineRequest> {
  serial_number: string;
}

export class MachineService {
  private static readonly BASE_PATH = '/api/v1/machines';

  /**
   * Fetch a list of machines with optional filtering and pagination
   */
  static async getMachines(
    page: number = 1,
    limit: number = 50,
    filters?: MachineFilters
  ): Promise<ApiResponse<MachineListResponse>> {
    const calculatedOffset = (page - 1) * limit;
    const params: Record<string, string> = {
      limit: limit.toString(),
      offset: calculatedOffset.toString(),
    };



    // Add filters to query parameters
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          if (key === 'ppm_status_filter' && typeof value === 'string') {
            params['ppm_status_filter'] = value;
          } else if (key === 'sort' && typeof value === 'string') {
            params[key] = value;
          } else if (key === 'limit' && typeof value === 'number') {
            params[key] = value.toString();
          } else if (key === 'offset' && typeof value === 'number') {
            params[key] = value.toString();
          } else if (key === 'q' && typeof value === 'string') {
            params['q'] = value;
          } else if (key === 'ppm_date_from' && typeof value === 'string') {
            params['ppm_date_from'] = value;
          } else if (key === 'ppm_date_to' && typeof value === 'string') {
            params['ppm_date_to'] = value;
          } else if (key === 'tnc_date_from' && typeof value === 'string') {
            params['tnc_date_from'] = value;
          } else if (key === 'tnc_date_to' && typeof value === 'string') {
            params['tnc_date_to'] = value;
          }
        }
      });
    }

    console.log('🔧 API: GET /api/v1/machines', { params });
    const response = await apiClient.get<MachineListResponse>(this.BASE_PATH, params);
    console.log('🔧 API: GET /api/v1/machines response', { 
      machinesCount: response.data?.machines?.length || 0,
      totalCount: response.data?.count || 0
    });



    return response;
  }

  /**
   * Fetch a single machine by serial number
   */
  static async getMachine(serialNumber: string): Promise<ApiResponse<Machine>> {
    const encodedSerialNumber = encodeURIComponent(serialNumber);
    return apiClient.get<Machine>(`${this.BASE_PATH}/${encodedSerialNumber}`);
  }

  /**
   * Create a new machine
   */
  static async createMachine(data: CreateMachineRequest): Promise<ApiResponse<Machine>> {
    return apiClient.post<Machine>(this.BASE_PATH, data);
  }

  /**
   * Update an existing machine
   */
  static async updateMachine(
    serialNumber: string,
    data: UpdateMachineRequest
  ): Promise<ApiResponse<Machine>> {
    const encodedSerialNumber = encodeURIComponent(serialNumber);
    return apiClient.put<Machine>(`${this.BASE_PATH}/${encodedSerialNumber}`, data);
  }

  /**
   * Delete a machine
   */
  static async deleteMachine(serialNumber: string): Promise<ApiResponse<void>> {
    const encodedSerialNumber = encodeURIComponent(serialNumber);
    return apiClient.delete<void>(`${this.BASE_PATH}/${encodedSerialNumber}`);
  }

  /**
   * Get machine statistics (if available)
   */
  static async getMachineStats(): Promise<ApiResponse<unknown>> {
    return apiClient.get<unknown>(`${this.BASE_PATH}/stats`);
  }

  /**
   * Download machines as CSV
   */
  static async exportMachinesCSV(filters?: MachineFilters): Promise<void> {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          params.set(key, String(value));
        }
      });
    }
    const qs = params.toString();
    const url = `${this.BASE_PATH}/export/csv${qs ? `?${qs}` : ''}`;
    await downloadCSV(url, 'machines.csv');
  }

  /**
   * Download maintenance records for a machine as CSV
   */
  static async exportMaintenanceCSV(serialNumber: string, query?: string): Promise<void> {
    const encoded = encodeURIComponent(serialNumber);
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    const qs = params.toString();
    const url = `${this.BASE_PATH}/${encoded}/maintenance/export/csv${qs ? `?${qs}` : ''}`;
    await downloadCSV(url, `maintenance_${serialNumber}.csv`);
  }
}

async function downloadCSV(path: string, fallbackFilename: string): Promise<void> {
  const resp = await fetchWithTokenRefresh(path, {});

  if (!resp.ok) {
    throw new Error(`CSV export failed: ${resp.status}`);
  }

  const blob = await resp.blob();
  const disposition = resp.headers.get('Content-Disposition');
  let filename = fallbackFilename;
  if (disposition) {
    const match = disposition.match(/filename="?([^"]+)"?/);
    if (match) filename = match[1];
  }

  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
} 