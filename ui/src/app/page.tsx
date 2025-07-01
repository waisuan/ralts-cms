'use client';

import { useState } from 'react';
import RecordsList, { SortType } from '@/components/RecordsList';
import SearchBar, { SearchOptions } from '@/components/SearchBar';
import OverdueAlert from '@/components/OverdueAlert';
import MaintenanceHistory from '@/components/MaintenanceHistory';
import { DEFAULT_SEARCH_PROPERTY } from '@/utils/constants';
import { mockMachines } from '@/data/mockMachines';
import { useOverdueStats } from '@/hooks/useOverdueStats';
import { Machine } from '@/types/machine';

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
  const [selectedMachineForHistory, setSelectedMachineForHistory] = useState<Machine | null>(null);

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

  const handleViewMachineHistory = (machine: Machine) => {
    setSelectedMachineForHistory(machine);
  };

  const handleBackFromHistory = () => {
    setSelectedMachineForHistory(null);
  };

  // If showing maintenance history, render that component
  if (selectedMachineForHistory) {
    return (
      <MaintenanceHistory
        machine={selectedMachineForHistory}
        onBack={handleBackFromHistory}
      />
    );
  }

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
          onViewMachineHistory={handleViewMachineHistory}
        />
      </div>
    </main>
  );
}
