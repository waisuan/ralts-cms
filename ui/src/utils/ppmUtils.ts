import { PPM_STATUSES, PPM_STATUS_COLORS, type PPMStatus } from './constants';

export interface PPMStatusInfo {
  label: PPMStatus;
  color: string;
}

/**
 * Calculate PPM status based on PPM date
 */
export function getPPMStatus(ppm_date: string): PPMStatusInfo | null {
  if (!ppm_date) return null;

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
