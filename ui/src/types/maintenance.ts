export interface Maintenance {
  machine_serial_number: string;
  work_order_number: string;
  work_order_date: string; // ISO 8601 format
  action_taken: string;
  reported_by: string;
  work_order_type: string;
  /** When present, from API: true if type is one of Preventive/Corrective/Emergency/Inspection. */
  work_order_type_is_standard?: boolean;
  attachment: string | null;
  created_at: string; // ISO 8601 format
  updated_at: string; // ISO 8601 format
  updated_by: string; // username of the last authenticated user to save this record
}

export type MaintenanceOrderType = 'Preventive' | 'Corrective' | 'Emergency' | 'Inspection' | 'Other';
