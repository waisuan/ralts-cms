export const MAINTENANCE_SORT_VALUES = [
  'updated_at_desc',
  'updated_at_asc',
  'work_order_date_desc',
  'work_order_date_asc',
] as const;

export type MaintenanceSortValue = (typeof MAINTENANCE_SORT_VALUES)[number];

export const DEFAULT_MAINTENANCE_SORT: MaintenanceSortValue = 'updated_at_desc';

const MAINTENANCE_SORT_LABELS: Record<MaintenanceSortValue, string> = {
  updated_at_desc: 'Newest First',
  updated_at_asc: 'Oldest First',
  work_order_date_desc: 'Work Order Date (Latest)',
  work_order_date_asc: 'Work Order Date (Earliest)',
};

export const MAINTENANCE_SORT_OPTIONS: ReadonlyArray<{
  value: MaintenanceSortValue;
  label: string;
}> = MAINTENANCE_SORT_VALUES.map((value) => ({
  value,
  label: MAINTENANCE_SORT_LABELS[value],
}));
