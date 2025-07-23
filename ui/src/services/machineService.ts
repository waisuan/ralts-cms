import { apiClient, ApiResponse } from '../utils/api';
import { Machine } from '../types/machine';

export interface MachineFilters {
  limit?: number;
  offset?: number;
  sort?: string;
  ppm_status_filter?: string;
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
  person_in_charge: string;
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
    limit: number = 10,
    filters?: MachineFilters
  ): Promise<ApiResponse<MachineListResponse>> {
    const calculatedOffset = (page - 1) * limit;
    const params: Record<string, string> = {
      limit: limit.toString(),
      offset: calculatedOffset.toString(),
    };

    console.log('🔧 MachineService: Making API call:', {
      url: this.BASE_PATH,
      params,
      page,
      limit,
      calculatedOffset,
      filters
    });

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
          }
        }
      });
    }

    const response = await apiClient.get<MachineListResponse>(this.BASE_PATH, params);

    console.log('🔧 MachineService: API response received:', {
      data: response.data ? {
        machinesCount: response.data.machines?.length || 0,
        totalCount: response.data.count,
        limit: response.data.limit,
        offset: response.data.offset,
        sort: response.data.sort
      } : null
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
} 