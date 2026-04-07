import type { MachineFilters } from '@/services/machineService';

export type MachineListFilterType = 'all' | 'overdue' | 'due';

export type MachineListSortType =
  | 'newest'
  | 'oldest'
  | 'ppm_date_asc'
  | 'ppm_date_desc'
  | 'tnc_date_asc'
  | 'tnc_date_desc';

export interface MachineListSearchOptions {
  query: string;
  property: string;
}

export interface MachineListDateRange {
  from: string | undefined;
  to: string | undefined;
}

/** Maps home / list UI state to API query filters (banner lock, search, sort, date ranges). */
export function buildMachineListFilters(params: {
  filterType: MachineListFilterType;
  sortBy: MachineListSortType;
  searchOptions: MachineListSearchOptions;
  ppmDateRange: MachineListDateRange;
  tncDateRange: MachineListDateRange;
}): MachineFilters {
  const { filterType, sortBy, searchOptions, ppmDateRange, tncDateRange } = params;
  const filters: MachineFilters = {};

  if (filterType === 'due') {
    filters.ppm_status_filter = 'due';
  } else if (filterType === 'overdue') {
    filters.ppm_status_filter = 'overdue';
  }

  if (sortBy === 'newest') {
    filters.sort = 'updated_at_desc';
  } else if (sortBy === 'oldest') {
    filters.sort = 'updated_at_asc';
  } else {
    filters.sort = sortBy;
  }

  if (filterType === 'all') {
    const q = searchOptions.query.trim();
    if (q) {
      if (searchOptions.property === 'any') {
        filters.q = q;
      } else if (searchOptions.property === 'ppm_status') {
        filters.ppm_status_filter = searchOptions.query;
      }
    }
  }

  if (ppmDateRange.from) {
    filters.ppm_date_from = ppmDateRange.from;
  }
  if (ppmDateRange.to) {
    filters.ppm_date_to = ppmDateRange.to;
  }
  if (tncDateRange.from) {
    filters.tnc_date_from = tncDateRange.from;
  }
  if (tncDateRange.to) {
    filters.tnc_date_to = tncDateRange.to;
  }

  return filters;
}
