'use client';

import { useCallback, useState } from 'react';
import RecordsList, { SortType } from '@/components/RecordsList';
import SearchBar, { SearchOptions } from '@/components/SearchBar';
import OverdueAlert from '@/components/OverdueAlert';
import { DEFAULT_SEARCH_PROPERTY } from '@/utils/constants';
import { nextFilterTypeAfterSearchChange } from '@/utils/bannerSearchLock';
import type { MachineListFilterType as FilterType } from '@/utils/machineListFilters';

export default function Home() {
  const [searchOptions, setSearchOptions] = useState<SearchOptions>({
    query: '',
    property: DEFAULT_SEARCH_PROPERTY,
  });
  const [filterType, setFilterType] = useState<FilterType>('all');
  const [sortBy, setSortBy] = useState<SortType>('newest');
  const [isOverdueDismissed, setIsOverdueDismissed] = useState(false);
  const [isDueDismissed, setIsDueDismissed] = useState(false);
  const [overdueCount, setOverdueCount] = useState(0);
  const [dueCount, setDueCount] = useState(0);
  const [isSearching, setIsSearching] = useState(false);
  const [totalResults, setTotalResults] = useState(0);

  const handleShowOverdue = () => {
    setSearchOptions({
      query: 'overdue',
      property: 'ppm_status',
    });
    setFilterType('overdue');
  };

  const handleShowDue = () => {
    setSearchOptions({
      query: 'due',
      property: 'ppm_status',
    });
    setFilterType('due');
  };

  const handleShowAll = () => {
    setFilterType('all');
    setSearchOptions({
      query: '',
      property: DEFAULT_SEARCH_PROPERTY,
    });
  };

  /** Leaving banner lock (different property/status/clear) switches to unfiltered "all" list semantics. */
  const handleSearchOptions = useCallback((opts: SearchOptions) => {
    setFilterType((prev) => nextFilterTypeAfterSearchChange(prev, opts));
    setSearchOptions(opts);
  }, []);

  const handleSortChange = (newSortBy: SortType) => {
    setSortBy(newSortBy);
  };

  const handleDismissOverdue = () => {
    setIsOverdueDismissed(true);
  };

  const handleDismissDue = () => {
    setIsDueDismissed(true);
  };

  const handleCountsUpdate = (overdue: number, due: number) => {
    setOverdueCount(overdue);
    setDueCount(due);
  };

  const handleSearchLoadingChange = (loading: boolean) => {
    setIsSearching(loading);
  };

  return (
    <div className="container mx-auto px-4">
      <div className="mb-8">
        {/* Overdue Alert Banner */}
        <OverdueAlert
          overdueCount={overdueCount}
          dueCount={dueCount}
          onShowOverdue={handleShowOverdue}
          onShowDue={handleShowDue}
          onDismissOverdue={handleDismissOverdue}
          onDismissDue={handleDismissDue}
          isOverdueDismissed={isOverdueDismissed}
          isDueDismissed={isDueDismissed}
        />

        {/* Enhanced Search Bar with Dropdown */}
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
        onShowAll={handleShowAll}
        onSortChange={handleSortChange}
        onCountsUpdate={handleCountsUpdate}
        onSearchLoadingChange={handleSearchLoadingChange}
        onTotalChange={setTotalResults}
      />
    </div>
  );
}
