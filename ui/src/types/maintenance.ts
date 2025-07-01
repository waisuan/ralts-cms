export interface Maintenance {
  machine_serial_number: string;
  work_order_number: string;
  work_order_date: string; // ISO 8601 format
  action_taken: string;
  reported_by: string;
  worker_order_type: string;
  attachment: string;
  created_at: string; // ISO 8601 format
  updated_at: string; // ISO 8601 format
}

export type MaintenanceOrderType = 'Preventive' | 'Corrective' | 'Emergency' | 'Inspection';
