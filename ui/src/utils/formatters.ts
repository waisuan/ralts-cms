import type { Maintenance } from '../types/maintenance';

export function formatDate(dateString: string): string {
  if (!dateString) return '-';
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

/**
 * Format a bare YYYY-MM-DD string for display without timezone shift.
 * Unlike formatDate, this parses the string as local time so "2026-04-12"
 * always displays as Apr 12, 2026 regardless of the browser's timezone.
 */
export function formatLocalDate(dateString: string): string {
  if (!dateString) return '-';
  const [year, month, day] = dateString.split('-').map(Number);
  return new Date(year, month - 1, day).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

export function formatDateTime(dateString: string): string {
  if (!dateString) return '-';
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

const STANDARD_WORK_ORDER_TYPES = ['Preventive', 'Corrective', 'Emergency', 'Inspection'] as const;

function isCustomWorkOrderTypeFromString(type: string): boolean {
  return !(STANDARD_WORK_ORDER_TYPES as readonly string[]).includes(type);
}

/** Prefer API `work_order_type_is_standard`; fallback matches backend SQL bucketing. */
export function isCustomWorkOrderType(recordOrType: Maintenance | string): boolean {
  if (typeof recordOrType === 'string') {
    return isCustomWorkOrderTypeFromString(recordOrType);
  }
  const r = recordOrType;
  if (r.work_order_type_is_standard !== undefined) {
    return !r.work_order_type_is_standard;
  }
  return isCustomWorkOrderTypeFromString(r.work_order_type);
}

/** Pill label: standard types as-is; custom strings show as "Other". */
export function workOrderTypePillLabel(recordOrType: Maintenance | string): string {
  const raw = typeof recordOrType === 'string' ? recordOrType : recordOrType.work_order_type;
  return isCustomWorkOrderType(recordOrType) ? 'Other' : raw;
}

export function getTypeColor(type: string): string {
  switch (type) {
    case 'Preventive':
      return 'bg-green-100 text-green-800';
    case 'Corrective':
      return 'bg-blue-100 text-blue-800';
    case 'Emergency':
      return 'bg-red-100 text-red-800';
    case 'Inspection':
      return 'bg-purple-100 text-purple-800';
    case 'Other':
      return 'bg-gray-100 text-gray-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
}
