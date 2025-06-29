export interface Machine {
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
  additional_notes: string;
  attachment: string;
  ppm_status: string;
  tnc_date: string; // ISO 8601
  ppm_date: string; // ISO 8601
  created_at: string; // ISO 8601
  updated_at: string; // ISO 8601
}
