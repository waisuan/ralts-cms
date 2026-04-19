import { PPM_STATUSES, PPM_STATUS_COLORS, type PPMStatus } from './constants';
import { isMachineDateUnset } from './dateUtils';

export interface PPMStatusInfo {
  label: PPMStatus;
  color: string;
}

/**
 * Calculate PPM status based on PPM date
 */
export function getPPMStatus(ppm_date: string): PPMStatusInfo | null {
  if (!ppm_date || isMachineDateUnset(ppm_date)) return null;

  const today = new Date();
  const ppm = new Date(ppm_date);

  // Remove time for accurate day comparison
  today.setHours(0, 0, 0, 0);
  ppm.setHours(0, 0, 0, 0);

  const diffDays = Math.ceil((ppm.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));

  if (diffDays < 0) {
    return {
      label: PPM_STATUSES.OVERDUE,
      color: PPM_STATUS_COLORS[PPM_STATUSES.OVERDUE],
    };
  } else if (diffDays === 0) {
    return {
      label: PPM_STATUSES.DUE,
      color: PPM_STATUS_COLORS[PPM_STATUSES.DUE],
    };
  } else if (diffDays > 0 && diffDays <= 14) { // 2 weeks = 14 days
    return {
      label: PPM_STATUSES.ALMOST_DUE,
      color: PPM_STATUS_COLORS[PPM_STATUSES.ALMOST_DUE],
    };
  } else {
    // More than 2 weeks in future - no status
    return null;
  }
}

/**
 * Get PPM status label only (useful for filtering)
 */
export function getPPMStatusLabel(ppm_date: string): PPMStatus | null {
  const status = getPPMStatus(ppm_date);
  return status ? status.label : null;
}

export interface PPMStatusDisplay {
  label: string;
  color: string;
}

/**
 * Map the server-provided `machine.ppm_status` string to a pill-ready
 * { label, color } for rendering. Returns null for unknown/blank values so
 * callers can skip rendering.
 */
export function getPPMStatusDisplay(ppmStatus: string): PPMStatusDisplay | null {
  if (!ppmStatus) return null;
  switch (ppmStatus) {
    case PPM_STATUSES.OVERDUE:
      return { label: 'Overdue', color: PPM_STATUS_COLORS[PPM_STATUSES.OVERDUE] };
    case PPM_STATUSES.DUE:
      return { label: 'Due', color: PPM_STATUS_COLORS[PPM_STATUSES.DUE] };
    case PPM_STATUSES.ALMOST_DUE:
      return { label: 'Upcoming', color: PPM_STATUS_COLORS[PPM_STATUSES.ALMOST_DUE] };
    default:
      return null;
  }
}
