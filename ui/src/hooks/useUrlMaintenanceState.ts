'use client';

import { useCallback, useMemo } from 'react';
import {
  useQueryStates,
  parseAsString,
  parseAsStringLiteral,
  parseAsInteger,
  parseAsNumberLiteral,
} from 'nuqs';
import { ITEMS_PER_PAGE_TABLE } from '@/components/recordsList/recordsListConstants';
import {
  DEFAULT_MAINTENANCE_SORT,
  MAINTENANCE_SORT_VALUES,
  type MaintenanceSortValue,
} from '@/utils/maintenanceSortOptions';
import { LIMIT_VALUES, clampPage, coerceLimit } from '@/hooks/urlStateHelpers';

const parsers = {
  q: parseAsString.withDefault(''),
  sort: parseAsStringLiteral(MAINTENANCE_SORT_VALUES).withDefault(DEFAULT_MAINTENANCE_SORT),
  page: parseAsInteger.withDefault(1),
  limit: parseAsNumberLiteral(LIMIT_VALUES).withDefault(ITEMS_PER_PAGE_TABLE),
};

export interface UrlMaintenanceState {
  q: string;
  sort: MaintenanceSortValue;
  page: number;
  limit: number;
}

export function useUrlMaintenanceState() {
  const [raw, setRaw] = useQueryStates(parsers);

  const state: UrlMaintenanceState = useMemo(
    () => ({
      q: raw.q,
      sort: raw.sort,
      page: clampPage(raw.page),
      limit: raw.limit,
    }),
    [raw],
  );

  const setSearchQuery = useCallback(
    (q: string) => setRaw({ q, page: 1 }),
    [setRaw],
  );

  const setSort = useCallback(
    (sort: MaintenanceSortValue) => setRaw({ sort, page: 1 }),
    [setRaw],
  );

  const setPage = useCallback(
    (page: number) => setRaw({ page: clampPage(page) }),
    [setRaw],
  );

  const setLimit = useCallback(
    (limit: number) => setRaw({ limit: coerceLimit(limit), page: 1 }),
    [setRaw],
  );

  const setMultiple = useCallback(
    (overrides: Partial<UrlMaintenanceState>) => {
      const next: Parameters<typeof setRaw>[0] = {};
      if (overrides.q !== undefined) next.q = overrides.q;
      if (overrides.sort !== undefined) next.sort = overrides.sort;
      if (overrides.limit !== undefined) next.limit = coerceLimit(overrides.limit);
      next.page = overrides.page !== undefined ? clampPage(overrides.page) : 1;
      return setRaw(next);
    },
    [setRaw],
  );

  return {
    ...state,
    setSearchQuery,
    setSort,
    setPage,
    setLimit,
    setMultiple,
  };
}
