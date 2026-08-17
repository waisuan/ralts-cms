export interface AssignedUser {
  id: number;
  username: string;
  email: string;
}

export interface Machine {
  serial_number: string;
  customer: string;
  state: string;
  account_type: string;
  model: string;
  status: string;
  brand: string;
  district: string;
  // person_in_charge always holds the assignee's display name. When
  // assigned_user_id is set the server derives it from that user's username;
  // otherwise it is the free text entered for an assignee who has no account.
  person_in_charge: string;
  // assigned_user_id links the machine to a registered user, which is what makes
  // notifications possible. Null for free-text assignees.
  assigned_user_id?: number | null;
  assigned_user?: AssignedUser | null;
  reported_by: string;
  additional_notes: string;
  attachment: string;
  ppm_status: string;
  tnc_date: string; // ISO 8601
  ppm_date: string; // ISO 8601
  created_at: string; // ISO 8601
  updated_at: string; // ISO 8601
  updated_by: string; // username of the last authenticated user to save this record
  maintenance_count?: number; // Server-driven maintenance record count
}
