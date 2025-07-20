import { useMemo } from 'react';
import { Machine } from '../types/machine';
import { getPPMStatus } from '../utils/ppmUtils';
import { PPM_STATUSES } from '../utils/constants';

export interface OverdueStats {
  overdueCount: number;
  dueCount: number;
  almostDueCount: number;
  totalCriticalCount: number; // overdue + due
  overdueMachines: Machine[];
  dueMachines: Machine[];
}

export function useOverdueStats(machines: Machine[]): OverdueStats {
  return useMemo(() => {
    let overdueCount = 0;
    let dueCount = 0;
    let almostDueCount = 0;
    const overdueMachines: Machine[] = [];
    const dueMachines: Machine[] = [];

    machines.forEach((machine) => {
      const status = getPPMStatus(machine.ppm_date);
      if (status) {
        switch (status.label) {
          case PPM_STATUSES.OVERDUE:
            overdueCount++;
            overdueMachines.push(machine);
            break;
          case PPM_STATUSES.DUE:
            dueCount++;
            dueMachines.push(machine);
            break;
          case PPM_STATUSES.ALMOST_DUE:
            almostDueCount++;
            break;
        }
      }
    });

    return {
      overdueCount,
      dueCount,
      almostDueCount,
      totalCriticalCount: overdueCount + dueCount,
      overdueMachines,
      dueMachines,
    };
  }, [machines]);
}
