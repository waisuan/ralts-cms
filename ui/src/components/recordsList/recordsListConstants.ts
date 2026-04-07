import type { MachineListSortType as SortType } from '@/utils/machineListFilters';

export type ViewMode = 'table' | 'cards';

export const VIEW_MODE_STORAGE_KEY = 'ralts-view-mode';
export const ITEMS_PER_PAGE_TABLE = 50;

export function loadViewMode(): ViewMode {
  if (typeof window === 'undefined') return 'table';
  try {
    const stored = localStorage.getItem(VIEW_MODE_STORAGE_KEY);
    if (stored === 'table' || stored === 'cards') return stored;
  } catch {
    /* use default */
  }
  return 'table';
}

export const SORT_TYPE_TO_API: Record<SortType, string> = {
  newest: 'updated_at_desc',
  oldest: 'updated_at_asc',
  ppm_date_asc: 'ppm_date_asc',
  ppm_date_desc: 'ppm_date_desc',
  tnc_date_asc: 'tnc_date_asc',
  tnc_date_desc: 'tnc_date_desc',
};

export const API_TO_SORT_TYPE: Record<string, SortType> = {
  updated_at_desc: 'newest',
  updated_at_asc: 'oldest',
  ppm_date_asc: 'ppm_date_asc',
  ppm_date_desc: 'ppm_date_desc',
  tnc_date_asc: 'tnc_date_asc',
  tnc_date_desc: 'tnc_date_desc',
};

export const SORT_OPTIONS: ReadonlyArray<{ value: SortType; label: string }> = [
  { value: 'newest', label: 'Newest First' },
  { value: 'oldest', label: 'Oldest First' },
  { value: 'ppm_date_asc', label: 'PPM Date (Earliest)' },
  { value: 'ppm_date_desc', label: 'PPM Date (Latest)' },
  { value: 'tnc_date_asc', label: 'TNC Date (Earliest)' },
  { value: 'tnc_date_desc', label: 'TNC Date (Latest)' },
];
