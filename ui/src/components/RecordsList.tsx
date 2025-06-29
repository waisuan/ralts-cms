'use client';

import { useState, useMemo, useEffect } from 'react';
import RecordCard from './RecordCard';
import AddMachineModal from './AddMachineModal';
import { mockMachines } from '../data/mockMachines';
import { SearchOptions } from './SearchBar';
import { isDateProperty } from '@/utils/constants';
import { getPPMStatusLabel } from '@/utils/ppmUtils';
import { useOverdueStats } from '../hooks/useOverdueStats';
import { Machine } from '../types/machine';

type FilterType = 'all' | 'overdue' | 'due';
export type SortType = 'newest' | 'oldest';

interface RecordsListProps {
  searchOptions: SearchOptions;
  filterType?: FilterType;
  sortBy?: SortType;
  onShowAll?: () => void;
  onSortChange?: (sortBy: SortType) => void;
}

const ITEMS_PER_PAGE = 3; // Show 3 machines initially, then load more

export default function RecordsList({
  searchOptions,
  filterType = 'all',
  sortBy = 'newest',
  onShowAll,
  onSortChange,
}: RecordsListProps) {
  const [machines, setMachines] = useState(mockMachines);
  const [displayedCount, setDisplayedCount] = useState(ITEMS_PER_PAGE);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [machineToDelete, setMachineToDelete] = useState<Machine | null>(null);

  // Calculate overdue statistics
  const overdueStats = useOverdueStats(machines);

  // Filter and sort machines based on search options, filter type, and sort order
  const filteredMachines = useMemo(() => {
    let filtered = machines;

    // Apply filter type first if enabled
    if (filterType === 'overdue') {
      filtered = overdueStats.overdueMachines;
    } else if (filterType === 'due') {
      filtered = overdueStats.dueMachines;
    }

    // Then apply search filter
    if (searchOptions.query.trim()) {
      const query = searchOptions.query.toLowerCase();
      const { property } = searchOptions;

      filtered = filtered.filter((machine) => {
        // Handle date properties differently
        if (isDateProperty(property)) {
          const fieldValue = machine[property as keyof typeof machine];
          if (!fieldValue) return false;

          // Convert both dates to YYYY-MM-DD format for comparison
          const machineDate = fieldValue.toString().split('T')[0]; // Extract date part from ISO string
          const searchDate = searchOptions.query; // Already in YYYY-MM-DD format from date input

          return machineDate === searchDate;
        } else if (property === 'ppm_status') {
          // Handle PPM status search by calculating status from ppm_date
          const calculatedStatus = getPPMStatusLabel(machine.ppm_date);
          if (!calculatedStatus) return false;

          // Use exact matching for PPM status since user selects from dropdown
          return calculatedStatus === searchOptions.query;
        } else {
          // Search in specific text property
          const fieldValue = machine[property as keyof typeof machine];
          return fieldValue && fieldValue.toString().toLowerCase().includes(query);
        }
      });
    }

    // Apply sorting
    const sorted = [...filtered].sort((a, b) => {
      const dateA = new Date(a.created_at).getTime();
      const dateB = new Date(b.created_at).getTime();

      return sortBy === 'newest' ? dateB - dateA : dateA - dateB;
    });

    return sorted;
  }, [
    machines,
    searchOptions,
    filterType,
    sortBy,
    overdueStats.overdueMachines,
    overdueStats.dueMachines,
  ]);

  // Get machines to display (limited by displayedCount)
  const displayedMachines = filteredMachines.slice(0, displayedCount);
  const hasMoreMachines = displayedCount < filteredMachines.length;

  // Reset displayed count when search changes
  useEffect(() => {
    setDisplayedCount(ITEMS_PER_PAGE);
  }, [searchOptions, filterType]);

  const handleLoadMore = () => {
    setDisplayedCount((prev) => Math.min(prev + ITEMS_PER_PAGE, filteredMachines.length));
  };

  const handleView = (serial_number: string) => {
    console.log('View machine:', serial_number);
  };

  const handleEdit = (serial_number: string) => {
    console.log('Edit machine:', serial_number);
  };

  const handleDelete = (serial_number: string) => {
    const machine = machines.find((m) => m.serial_number === serial_number);
    if (machine) {
      setMachineToDelete(machine);
      setShowDeleteConfirm(true);
    }
  };

  const handleConfirmDelete = () => {
    if (machineToDelete) {
      setMachines(
        machines.filter((machine) => machine.serial_number !== machineToDelete.serial_number)
      );
      setMachineToDelete(null);
    }
    setShowDeleteConfirm(false);
  };

  const handleCancelDelete = () => {
    setMachineToDelete(null);
    setShowDeleteConfirm(false);
  };

  const handleAddMachine = (newMachine: Omit<Machine, 'created_at' | 'updated_at'>) => {
    const machine: Machine = {
      ...newMachine,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    setMachines([machine, ...machines]);
  };

  const handleOpenAddModal = () => {
    setIsAddModalOpen(true);
  };

  const handleCloseAddModal = () => {
    setIsAddModalOpen(false);
  };

  const getFilterStatusText = () => {
    if (filterType === 'overdue') return ' (overdue only)';
    if (filterType === 'due') return ' (due today only)';
    if (searchOptions.query && filteredMachines.length !== machines.length) {
      return ` (filtered from ${machines.length} total)`;
    }
    return '';
  };

  const getEmptyStateMessage = () => {
    if (filterType === 'overdue') {
      return {
        title: 'No overdue machines found',
        subtitle: 'Great! All machines are up to date with their PPM maintenance.',
      };
    }
    if (filterType === 'due') {
      return {
        title: 'No machines due today',
        subtitle: 'No machines require PPM maintenance today.',
      };
    }
    if (searchOptions.query) {
      return {
        title: 'No machines found',
        subtitle: `No machines match "${searchOptions.query}". Try a different search term.`,
      };
    }
    return {
      title: 'No machines found',
      subtitle: 'Get started by creating your first machine.',
    };
  };

  const emptyState = getEmptyStateMessage();

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div className="flex-1">
          <div className="flex items-center gap-4">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">All Machines</h2>
              <p className="text-sm text-gray-500 mt-1">
                Showing {displayedMachines.length} of {filteredMachines.length} machine
                {filteredMachines.length !== 1 ? 's' : ''}
                {getFilterStatusText()}
              </p>
            </div>

            {/* Overdue Statistics Badge */}
            {overdueStats.totalCriticalCount > 0 && (
              <div className="flex items-center gap-2">
                {overdueStats.overdueCount > 0 && (
                  <div className="bg-red-100 text-red-800 px-3 py-1 rounded-full text-sm font-medium flex items-center gap-1">
                    <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"
                      />
                    </svg>
                    {overdueStats.overdueCount} Overdue
                  </div>
                )}
                {overdueStats.dueCount > 0 && (
                  <div className="bg-orange-100 text-orange-800 px-3 py-1 rounded-full text-sm font-medium flex items-center gap-1">
                    <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    {overdueStats.dueCount} Due Today
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        <div className="flex items-center gap-3">
          {/* Sort By Dropdown */}
          {onSortChange && (
            <div className="flex items-center gap-2">
              <label className="text-sm font-medium text-gray-700">Sort By:</label>
              <select
                value={sortBy}
                onChange={(e) => onSortChange(e.target.value as SortType)}
                className="border border-gray-300 rounded-lg px-3 py-2 text-sm text-black focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white"
              >
                <option value="newest">Newest First</option>
                <option value="oldest">Oldest First</option>
              </select>
            </div>
          )}

          {/* Show All button when filtering */}
          {(filterType === 'overdue' || filterType === 'due') && onShowAll && (
            <button
              onClick={onShowAll}
              className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2"
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>
              Show All
            </button>
          )}

          <button
            onClick={handleOpenAddModal}
            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors"
          >
            Add New Machine
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {displayedMachines.map((machine) => (
          <RecordCard
            key={machine.serial_number}
            machine={machine}
            onView={handleView}
            onEdit={handleEdit}
            onDelete={handleDelete}
          />
        ))}
      </div>

      {/* Load More Button */}
      {hasMoreMachines && (
        <div className="flex justify-center pt-4">
          <button
            onClick={handleLoadMore}
            className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-6 py-3 rounded-lg font-medium transition-colors"
          >
            Load More ({filteredMachines.length - displayedCount} remaining)
          </button>
        </div>
      )}

      {filteredMachines.length === 0 && (
        <div className="text-center py-12">
          <div className="text-gray-400 text-6xl mb-4">📄</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">{emptyState.title}</h3>
          <p className="text-gray-500">{emptyState.subtitle}</p>
        </div>
      )}

      {/* Add Machine Modal */}
      <AddMachineModal
        isOpen={isAddModalOpen}
        onClose={handleCloseAddModal}
        onAdd={handleAddMachine}
      />

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && machineToDelete && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="fixed inset-0 bg-black bg-opacity-50" onClick={handleCancelDelete} />
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="relative bg-white rounded-lg shadow-xl max-w-md w-full">
              <div className="p-6">
                <div className="flex items-center mb-4">
                  <div className="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-100">
                    <svg
                      className="h-6 w-6 text-red-600"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      />
                    </svg>
                  </div>
                </div>
                <div className="text-center">
                  <h3 className="text-lg font-medium text-gray-900 mb-2">Delete Machine</h3>
                  <p className="text-sm text-gray-500 mb-2">
                    Are you sure you want to delete machine{' '}
                    <strong>{machineToDelete.serial_number}</strong>?
                  </p>
                  <p className="text-sm text-gray-500 mb-6">
                    This action cannot be undone. All data associated with this machine will be
                    permanently removed.
                  </p>
                </div>
                <div className="flex space-x-3">
                  <button
                    type="button"
                    onClick={handleCancelDelete}
                    className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={handleConfirmDelete}
                    className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors"
                  >
                    Delete Machine
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
