import type { MachineListFilterType } from './machineListFilters';

/**
 * When the list is locked to overdue/due from the notice banner, changing search away from
 * that PPM status (or off PPM property) clears the lock so list semantics match "all".
 */
export function nextFilterTypeAfterSearchChange(
  prev: MachineListFilterType,
  nextSearch: { property: string; query: string }
): MachineListFilterType {
  if (prev !== 'overdue' && prev !== 'due') return prev;
  const lockedQuery = prev === 'overdue' ? 'overdue' : 'due';
  const leavingBannerLock =
    nextSearch.property !== 'ppm_status' ||
    (nextSearch.property === 'ppm_status' && nextSearch.query !== lockedQuery);
  return leavingBannerLock ? 'all' : prev;
}
