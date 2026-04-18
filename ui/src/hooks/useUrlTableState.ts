'use client';

import { useCallback, useMemo } from 'react';
import {
  useQueryStates,
  parseAsString,
  parseAsStringLiteral,
  parseAsInteger,
  parseAsNumberLiteral,
  createParser,
} from 'nuqs';
import type { SearchOptions } from '@/components/SearchBar';
import type { MachineListFilterType, MachineListSortType } from '@/utils/machineListFilters';
import type { DateRangeValue } from '@/components/DateRangePicker';
import {
  DEFAULT_SEARCH_PROPERTY,
  SEARCH_PROPERTIES,
  type SearchPropertyValue,
} from '@/utils/constants';
import {
  ITEMS_PER_PAGE_TABLE,
  SORT_OPTIONS,
} from '@/components/recordsList/recordsListConstants';
import { LIMIT_VALUES, clampPage, coerceLimit } from '@/hooks/urlStateHelpers';

const SEARCH_PROPERTY_VALUES = SEARCH_PROPERTIES.map(
  (p) => p.value,
) as readonly SearchPropertyValue[];
const FILTER_VALUES = ['all', 'overdue', 'due'] as const satisfies readonly MachineListFilterType[];
const SORT_VALUES: readonly MachineListSortType[] = SORT_OPTIONS.map((o) => o.value);

function coerceSearchProperty(value: string): SearchPropertyValue {
  return (SEARCH_PROPERTY_VALUES as readonly string[]).includes(value)
    ? (value as SearchPropertyValue)
    : DEFAULT_SEARCH_PROPERTY;
}

const DATE_RE = /^\d{4}-\d{2}-\d{2}$/;
const parseAsDateString = createParser<string>({
  parse: (v) => (DATE_RE.test(v) ? v : null),
  serialize: (v) => v,
});

const parsers = {
  q: parseAsString.withDefault(''),
  search_by: parseAsStringLiteral(SEARCH_PROPERTY_VALUES).withDefault(DEFAULT_SEARCH_PROPERTY),
  filter: parseAsStringLiteral(FILTER_VALUES).withDefault('all'),
  sort: parseAsStringLiteral(SORT_VALUES).withDefault('newest'),
  page: parseAsInteger.withDefault(1),
  limit: parseAsNumberLiteral(LIMIT_VALUES).withDefault(ITEMS_PER_PAGE_TABLE),
  ppm_from: parseAsDateString,
  ppm_to: parseAsDateString,
  tnc_from: parseAsDateString,
  tnc_to: parseAsDateString,
};

export interface UrlTableState {
  searchOptions: SearchOptions;
  filterType: MachineListFilterType;
  sortBy: MachineListSortType;
  page: number;
  limit: number;
  ppmDateRange: DateRangeValue;
  tncDateRange: DateRangeValue;
}

export function useUrlTableState() {
  const [raw, setRaw] = useQueryStates(parsers);

  const state: UrlTableState = useMemo(
    () => ({
      searchOptions: {
        query: raw.q,
        property: raw.search_by,
      },
      filterType: raw.filter,
      sortBy: raw.sort,
      page: clampPage(raw.page),
      limit: raw.limit,
      ppmDateRange: {
        from: raw.ppm_from ?? undefined,
        to: raw.ppm_to ?? undefined,
      },
      tncDateRange: {
        from: raw.tnc_from ?? undefined,
        to: raw.tnc_to ?? undefined,
      },
    }),
    [raw],
  );

  const setSearchOptions = useCallback(
    (searchOptions: SearchOptions) => {
      setRaw({
        q: searchOptions.query,
        search_by: coerceSearchProperty(searchOptions.property),
        page: 1,
      });
    },
    [setRaw],
  );

  const setFilterType = useCallback(
    (filterType: MachineListFilterType) => {
      setRaw({ filter: filterType, page: 1 });
    },
    [setRaw],
  );

  const setSortBy = useCallback(
    (sortBy: MachineListSortType) => {
      setRaw({ sort: sortBy, page: 1 });
    },
    [setRaw],
  );

  const setPage = useCallback(
    (page: number) => {
      setRaw({ page: clampPage(page) });
    },
    [setRaw],
  );

  const setLimit = useCallback(
    (limit: number) => {
      setRaw({ limit: coerceLimit(limit), page: 1 });
    },
    [setRaw],
  );

  const setPpmDateRange = useCallback(
    (range: DateRangeValue) => {
      setRaw({
        ppm_from: range.from ?? null,
        ppm_to: range.to ?? null,
        page: 1,
      });
    },
    [setRaw],
  );

  const setTncDateRange = useCallback(
    (range: DateRangeValue) => {
      setRaw({
        tnc_from: range.from ?? null,
        tnc_to: range.to ?? null,
        page: 1,
      });
    },
    [setRaw],
  );

  const clearDateRanges = useCallback(() => {
    setRaw({
      ppm_from: null,
      ppm_to: null,
      tnc_from: null,
      tnc_to: null,
      page: 1,
    });
  }, [setRaw]);

  const setMultiple = useCallback(
    (overrides: Partial<UrlTableState>) => {
      const next: Parameters<typeof setRaw>[0] = {};
      if (overrides.searchOptions) {
        next.q = overrides.searchOptions.query;
        next.search_by = coerceSearchProperty(overrides.searchOptions.property);
      }
      if (overrides.filterType !== undefined) next.filter = overrides.filterType;
      if (overrides.sortBy !== undefined) next.sort = overrides.sortBy;
      if (overrides.limit !== undefined) next.limit = coerceLimit(overrides.limit);
      if (overrides.ppmDateRange !== undefined) {
        next.ppm_from = overrides.ppmDateRange.from ?? null;
        next.ppm_to = overrides.ppmDateRange.to ?? null;
      }
      if (overrides.tncDateRange !== undefined) {
        next.tnc_from = overrides.tncDateRange.from ?? null;
        next.tnc_to = overrides.tncDateRange.to ?? null;
      }
      next.page = overrides.page !== undefined ? clampPage(overrides.page) : 1;
      setRaw(next);
    },
    [setRaw],
  );

  return {
    ...state,
    setSearchOptions,
    setFilterType,
    setSortBy,
    setPage,
    setLimit,
    setPpmDateRange,
    setTncDateRange,
    clearDateRanges,
    setMultiple,
  };
}
