// Search-related constants
export const DEFAULT_SEARCH_PROPERTY = 'serial_number' as const;

// Other application constants can be added here as needed
export const SEARCH_PROPERTIES = [
  { value: 'serial_number', label: 'Serial Number' },
  { value: 'customer', label: 'Customer' },
  { value: 'state', label: 'State' },
  { value: 'account_type', label: 'Account Type' },
  { value: 'model', label: 'Model' },
  { value: 'brand', label: 'Brand' },
  { value: 'district', label: 'District' },
  { value: 'person_in_charge', label: 'Person in Charge' },
  { value: 'reported_by', label: 'Reported By' },
  { value: 'status', label: 'Status' },
  { value: 'ppm_status', label: 'PPM Status' },
  { value: 'tnc_date', label: 'TNC Date' },
  { value: 'ppm_date', label: 'PPM Date' },
] as const;

// Properties that should use date picker instead of text input
export const DATE_PROPERTIES = ['tnc_date', 'ppm_date'] as const;

// PPM Status constants
export const PPM_STATUSES = {
  OVERDUE: 'Overdue',
  DUE: 'Due',
  DUE_SOON: 'Due Soon',
  UPCOMING: 'Upcoming',
} as const;

export const PPM_STATUS_COLORS = {
  [PPM_STATUSES.OVERDUE]: 'bg-red-100 text-red-800',
  [PPM_STATUSES.DUE]: 'bg-orange-100 text-orange-800',
  [PPM_STATUSES.DUE_SOON]: 'bg-yellow-100 text-yellow-800',
  [PPM_STATUSES.UPCOMING]: 'bg-green-100 text-green-800',
} as const;

// Type for search property values
export type SearchPropertyValue = (typeof SEARCH_PROPERTIES)[number]['value'];
export type DateProperty = (typeof DATE_PROPERTIES)[number];
export type PPMStatus = (typeof PPM_STATUSES)[keyof typeof PPM_STATUSES];

// Type guard to check if a property is a date property
export function isDateProperty(property: string): property is DateProperty {
  return (DATE_PROPERTIES as readonly string[]).includes(property);
}
