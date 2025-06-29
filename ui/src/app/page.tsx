'use client';

import { useState } from 'react';
import RecordsList from '@/components/RecordsList';
import SearchBar, { SearchOptions } from '@/components/SearchBar';
import OverdueAlert from '@/components/OverdueAlert';
import { DEFAULT_SEARCH_PROPERTY } from '@/utils/constants';
import { mockMachines } from '@/data/mockMachines';
import { useOverdueStats } from '@/hooks/useOverdueStats';

type FilterType = 'all' | 'overdue' | 'due';

export default function Home() {
  const [searchOptions, setSearchOptions] = useState<SearchOptions>({
    query: '',
    property: DEFAULT_SEARCH_PROPERTY,
  });
  const [filterType, setFilterType] = useState<FilterType>('all');

  // Calculate overdue statistics for the alert banner
  const overdueStats = useOverdueStats(mockMachines);

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

  const handleToggleOverdueFilter = () => {
    setFilterType(filterType === 'overdue' ? 'all' : 'overdue');
  };

  return (
    <main className="min-h-screen bg-gray-50 py-8">
      <div className="container mx-auto px-4">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Ralts CMS</h1>
          <p className="text-gray-600 mb-6">Content Management System</p>

          {/* Overdue Alert Banner */}
          <OverdueAlert
            stats={overdueStats}
            onShowOverdue={handleShowOverdue}
            onShowDue={handleShowDue}
          />

          {/* Enhanced Search Bar with Dropdown */}
          <SearchBar searchOptions={searchOptions} onSearch={setSearchOptions} />
        </div>

        <RecordsList
          searchOptions={searchOptions}
          filterType={filterType}
          onToggleOverdueFilter={handleToggleOverdueFilter}
        />
      </div>
    </main>
  );
}
