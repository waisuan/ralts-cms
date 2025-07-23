import { apiClient, ApiResponse } from '../utils/api';
import { Maintenance } from '../types/maintenance';

export interface MaintenanceFilters {
  limit?: number;
  offset?: number;
  sort?: string;
}

export interface MaintenanceListResponse {
  maintenance: Maintenance[];
  count: number;
  limit: number;
  offset: number;
  sort: string;
}

export interface CreateMaintenanceRequest {
  work_order_number: string;
  work_order_date: string;
  action_taken: string;
  reported_by: string;
  worker_order_type: string;
  attachment?: string;
}

export interface UpdateMaintenanceRequest extends Partial<CreateMaintenanceRequest> {
  work_order_number: string;
}

export class MaintenanceService {
  private static readonly BASE_PATH = '/api/v1/machines';

  /**
   * Fetch a list of maintenance records for a specific machine
   */
  static async getMaintenanceList(
    machineSerialNumber: string,
    page: number = 1,
    limit: number = 10,
    filters?: MaintenanceFilters
  ): Promise<ApiResponse<MaintenanceListResponse>> {
    const encodedSerialNumber = encodeURIComponent(machineSerialNumber);
    const calculatedOffset = (page - 1) * limit;
    const params: Record<string, string> = {
      limit: limit.toString(),
      offset: calculatedOffset.toString(),
    };

    console.log('🔧 MaintenanceService: Making API call:', {
      url: `${this.BASE_PATH}/${encodedSerialNumber}/maintenance`,
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
          if (key === 'sort' && typeof value === 'string') {
            params[key] = value;
          } else if (key === 'limit' && typeof value === 'number') {
            params[key] = value.toString();
          } else if (key === 'offset' && typeof value === 'number') {
            params[key] = value.toString();
          }
        }
      });
    }

    const response = await apiClient.get<MaintenanceListResponse>(
      `${this.BASE_PATH}/${encodedSerialNumber}/maintenance`,
      params
    );

    console.log('🔧 MaintenanceService: API response received:', {
      data: response.data ? {
        maintenanceCount: response.data.maintenance?.length || 0,
        totalCount: response.data.count,
        limit: response.data.limit,
        offset: response.data.offset,
        sort: response.data.sort
      } : null
    });

    return response;
  }

  /**
   * Fetch a single maintenance record by work order number
   */
  static async getMaintenance(
    machineSerialNumber: string,
    workOrderNumber: string
  ): Promise<ApiResponse<Maintenance>> {
    const encodedSerialNumber = encodeURIComponent(machineSerialNumber);
    const encodedWorkOrderNumber = encodeURIComponent(workOrderNumber);
    return apiClient.get<Maintenance>(
      `${this.BASE_PATH}/${encodedSerialNumber}/maintenance/${encodedWorkOrderNumber}`
    );
  }

  /**
   * Create a new maintenance record
   */
  static async createMaintenance(
    machineSerialNumber: string,
    data: CreateMaintenanceRequest
  ): Promise<ApiResponse<Maintenance>> {
    const encodedSerialNumber = encodeURIComponent(machineSerialNumber);
    return apiClient.post<Maintenance>(
      `${this.BASE_PATH}/${encodedSerialNumber}/maintenance`,
      data
    );
  }

  /**
   * Update an existing maintenance record
   */
  static async updateMaintenance(
    machineSerialNumber: string,
    workOrderNumber: string,
    data: UpdateMaintenanceRequest
  ): Promise<ApiResponse<Maintenance>> {
    const encodedSerialNumber = encodeURIComponent(machineSerialNumber);
    return apiClient.put<Maintenance>(
      `${this.BASE_PATH}/${encodedSerialNumber}/maintenance`,
      data
    );
  }

  /**
   * Delete a maintenance record
   */
  static async deleteMaintenance(
    machineSerialNumber: string,
    workOrderNumber: string
  ): Promise<ApiResponse<void>> {
    const encodedSerialNumber = encodeURIComponent(machineSerialNumber);
    const encodedWorkOrderNumber = encodeURIComponent(workOrderNumber);
    return apiClient.delete<void>(
      `${this.BASE_PATH}/${encodedSerialNumber}/maintenance/${encodedWorkOrderNumber}`
    );
  }
} 