import { apiClient, ApiResponse } from '../utils/api';
import { Maintenance } from '../types/maintenance';
import { htmlDateToBackendDate } from '../utils/dateUtils';

export interface MaintenanceFilters {
  limit?: number;
  offset?: number;
  sort?: string;
  q?: string; // General search query parameter
  work_order_q?: string; // Work order specific search
  reported_by_q?: string; // Reported by specific search
  work_order_type_q?: string; // Work order type specific search
}

export interface MaintenanceListResponse {
  maintenance: Maintenance[];
  preventative_count: number;
  corrective_count: number;
  emergency_count: number;
  inspection_count: number;
  other_count: number;
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
  work_order_type: string;
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
    limit: number = 50,
    filters?: MaintenanceFilters
  ): Promise<ApiResponse<MaintenanceListResponse>> {
    const encodedSerialNumber = encodeURIComponent(machineSerialNumber);
    const calculatedOffset = (page - 1) * limit;
    const params: Record<string, string> = {
      limit: limit.toString(),
      offset: calculatedOffset.toString(),
    };

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
          } else if (typeof value === 'string' && ['q', 'work_order_q', 'reported_by_q', 'work_order_type_q'].includes(key)) {
            params[key] = value;
          }
        }
      });
    }

    const response = await apiClient.get<MaintenanceListResponse>(
      `${this.BASE_PATH}/${encodedSerialNumber}/maintenance`,
      params
    );

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
    
    // Convert HTML date to backend ISO format
    const backendData = {
      ...data,
      work_order_date: htmlDateToBackendDate(data.work_order_date)
    };
    
    return apiClient.post<Maintenance>(
      `${this.BASE_PATH}/${encodedSerialNumber}/maintenance`,
      backendData
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
    
    // Convert HTML date to backend ISO format if work_order_date is provided
    const backendData = {
      ...data,
      ...(data.work_order_date && { work_order_date: htmlDateToBackendDate(data.work_order_date) })
    };
    
    return apiClient.put<Maintenance>(
      `${this.BASE_PATH}/${encodedSerialNumber}/maintenance`,
      backendData
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