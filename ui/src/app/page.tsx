'use client';

import { useState } from 'react';
import RecordsList, { SortType } from '@/components/RecordsList';
import SearchBar, { SearchOptions } from '@/components/SearchBar';
import OverdueAlert from '@/components/OverdueAlert';
import { DEFAULT_SEARCH_PROPERTY } from '@/utils/constants';

type FilterType = 'all' | 'overdue' | 'due';

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

  const handleShowOverdue = () => {
    // Clear any existing search and show only overdue machines
    setSearchOptions({
      query: '',
      property: DEFAULT_SEARCH_PROPERTY,
    });
    setFilterType('overdue');
  };

  const handleShowDue = () => {
    // Clear any existing search and show only due machines
    setSearchOptions({
      query: '',
      property: DEFAULT_SEARCH_PROPERTY,
    });
    setFilterType('due');
  };

  const handleShowAll = () => {
    setFilterType('all');
  };

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
        <SearchBar searchOptions={searchOptions} onSearch={setSearchOptions} />
      </div>

      <RecordsList
        searchOptions={searchOptions}
        filterType={filterType}
        sortBy={sortBy}
        onShowAll={handleShowAll}
        onSortChange={handleSortChange}
        onCountsUpdate={handleCountsUpdate}
      />
    </div>
  );
}
