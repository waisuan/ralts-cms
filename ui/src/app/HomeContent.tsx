'use client';

import { useCallback, useState } from 'react';
import RecordsList from '@/components/RecordsList';
import SearchBar, { SearchOptions } from '@/components/SearchBar';
import OverdueAlert from '@/components/OverdueAlert';
import {
  DEFAULT_SEARCH_PROPERTY,
  PPM_STATUS_PROPERTY,
  PPM_STATUSES,
} from '@/utils/constants';
import { nextFilterTypeAfterSearchChange } from '@/utils/bannerSearchLock';
import { useUrlTableState } from '@/hooks/useUrlTableState';

export default function HomeContent() {
  const {
    searchOptions,
    filterType,
    sortBy,
    page,
    limit,
    ppmDateRange,
    tncDateRange,
    setSortBy,
    setPage,
    setLimit,
    setPpmDateRange,
    setTncDateRange,
    clearDateRanges,
    setMultiple,
  } = useUrlTableState();

  const [isOverdueDismissed, setIsOverdueDismissed] = useState(false);
  const [isDueDismissed, setIsDueDismissed] = useState(false);
  const [overdueCount, setOverdueCount] = useState(0);
  const [dueCount, setDueCount] = useState(0);
  const [isSearching, setIsSearching] = useState(false);
  const [totalResults, setTotalResults] = useState(0);

  const handleShowOverdue = useCallback(() => {
    setMultiple({
      searchOptions: { query: PPM_STATUSES.OVERDUE, property: PPM_STATUS_PROPERTY },
      filterType: 'overdue',
    });
  }, [setMultiple]);

  const handleShowDue = useCallback(() => {
    setMultiple({
      searchOptions: { query: PPM_STATUSES.DUE, property: PPM_STATUS_PROPERTY },
      filterType: 'due',
    });
  }, [setMultiple]);

  const handleShowAll = useCallback(() => {
    setMultiple({
      searchOptions: { query: '', property: DEFAULT_SEARCH_PROPERTY },
      filterType: 'all',
    });
  }, [setMultiple]);

  const handleSearchOptions = useCallback(
    (opts: SearchOptions) => {
      const nextFilter = nextFilterTypeAfterSearchChange(filterType, opts);
      setMultiple({ searchOptions: opts, filterType: nextFilter });
    },
    [filterType, setMultiple],
  );

  const handleCountsUpdate = useCallback((overdue: number, due: number) => {
    setOverdueCount(overdue);
    setDueCount(due);
  }, []);

  const handleSearchLoadingChange = useCallback((loading: boolean) => {
    setIsSearching(loading);
  }, []);

  return (
    <div className="container mx-auto px-4">
      <div className="mb-8">
        <OverdueAlert
          overdueCount={overdueCount}
          dueCount={dueCount}
          onShowOverdue={handleShowOverdue}
          onShowDue={handleShowDue}
          onDismissOverdue={() => setIsOverdueDismissed(true)}
          onDismissDue={() => setIsDueDismissed(true)}
          isOverdueDismissed={isOverdueDismissed}
          isDueDismissed={isDueDismissed}
        />

        <SearchBar
          searchOptions={searchOptions}
          onSearch={handleSearchOptions}
          isSearching={isSearching}
          resultCount={searchOptions.query ? totalResults : undefined}
        />
      </div>

      <RecordsList
        searchOptions={searchOptions}
        filterType={filterType}
        sortBy={sortBy}
        page={page}
        limit={limit}
        ppmDateRange={ppmDateRange}
        tncDateRange={tncDateRange}
        onPageChange={setPage}
        onLimitChange={setLimit}
        onPpmDateRangeChange={setPpmDateRange}
        onTncDateRangeChange={setTncDateRange}
        onClearDateRanges={clearDateRanges}
        onShowAll={handleShowAll}
        onSortChange={setSortBy}
        onCountsUpdate={handleCountsUpdate}
        onSearchLoadingChange={handleSearchLoadingChange}
        onTotalChange={setTotalResults}
      />
    </div>
  );
}
