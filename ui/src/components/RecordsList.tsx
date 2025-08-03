'use client';

import { useState, useMemo, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Machine } from '../types/machine';
import { MachineFilters } from '../services/machineService';
import { useMachines } from '../hooks/useMachines';
import { useMachine } from '../hooks/useMachine';
import { SearchOptions } from './SearchBar';
import { isDateProperty } from '../utils/constants';
import RecordCard from './RecordCard';
import MachineModal from './MachineModal';
import FullPageLoader from './FullPageLoader';
import LoadingOverlay from './LoadingOverlay';

type FilterType = 'all' | 'overdue' | 'due';
export type SortType = 'newest' | 'oldest';

interface RecordsListProps {
  searchOptions: SearchOptions;
  filterType?: FilterType;
  sortBy?: SortType;
  onShowAll?: () => void;
  onSortChange?: (sortBy: SortType) => void;
  onCountsUpdate?: (overdue: number, due: number) => void;
  onSearchLoadingChange?: (loading: boolean) => void;
}

const ITEMS_PER_PAGE = 12; // Show 12 machines per page

export default function RecordsList({
  searchOptions,
  filterType = 'all',
  sortBy = 'newest',
  onShowAll,
  onSortChange,
  onCountsUpdate,
  onSearchLoadingChange,
}: RecordsListProps) {
  const router = useRouter();
  const [isMachineModalOpen, setIsMachineModalOpen] = useState(false);
  const [modalMode, setModalMode] = useState<'add' | 'edit'>('add');
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [machineToDelete, setMachineToDelete] = useState<Machine | null>(null);
  const [machineToEdit, setMachineToEdit] = useState<Machine | null>(null);
  const [isNavigating, setIsNavigating] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  // CRUD operation loading states
  const [isDeletingMachine, setIsDeletingMachine] = useState(false);
  const [isCreatingMachine, setIsCreatingMachine] = useState(false);
  const [isUpdatingMachine, setIsUpdatingMachine] = useState(false);

  // Convert search options to API filters
  const apiFilters: MachineFilters = useMemo(() => {
    const filters: MachineFilters = {};
    
    // Handle ppm_status_filter based on filterType
    if (filterType === 'due') {
      filters.ppm_status_filter = 'due';
    } else if (filterType === 'overdue') {
      filters.ppm_status_filter = 'overdue';
    }
    
    // Handle sorting
    if (sortBy === 'newest') {
      filters.sort = 'created_at_desc';
    } else if (sortBy === 'oldest') {
      filters.sort = 'created_at_asc';
    }
    
    // Handle search query - only send to API for 'any' property or specific supported properties
    if (searchOptions.query.trim()) {
      if (searchOptions.property === 'any') {
        // For 'any' search, send the query to backend
        filters.q = searchOptions.query.trim();
      } else if (searchOptions.property === 'ppm_status') {
        // For PPM status, use the existing ppm_status_filter
        filters.ppm_status_filter = searchOptions.query;
      } else if (isDateProperty(searchOptions.property)) {
        // For date properties, we'll handle client-side for now
        // Could be enhanced to use backend date filtering in the future
      }
    }
    
    return filters;
  }, [filterType, sortBy, searchOptions]);

  // Use the machines API hook with server-side pagination and debounced search
  const {
    machines,
    total,
    offset,
    limit,
    loading,
    error,
    refetch,
    loadMore,
    overdueCount,
    dueCount,
  } = useMachines({
    page: 1,
    limit: ITEMS_PER_PAGE,
    filters: apiFilters,
    autoFetch: true,
  });

  // Use individual machine hook for CRUD operations
  const {
    createMachine,
    updateMachine,
    deleteMachine,
  } = useMachine();

  // Update parent component with counts when they change
  useEffect(() => {
    if (onCountsUpdate) {
      onCountsUpdate(overdueCount, dueCount);
    }
  }, [overdueCount, dueCount, onCountsUpdate]);

  // Update parent component with search loading state
  useEffect(() => {
    if (onSearchLoadingChange) {
      onSearchLoadingChange(loading);
    }
  }, [loading, onSearchLoadingChange]);



  // Use machines directly from API (server-side search is now handled by the backend)
  const filteredMachines = machines;



  const handleLoadMore = async () => {
    await loadMore();
  };

  const handleView = (serial_number: string) => {
    setIsNavigating(true);
    router.push(`/machines/${encodeURIComponent(serial_number)}`);
  };

  const handleEdit = (serial_number: string) => {
    const machine = machines.find((m) => m.serial_number === serial_number);
    if (machine) {
      setMachineToEdit(machine);
      setModalMode('edit');
      setIsMachineModalOpen(true);
    }
  };

  const handleDelete = (serial_number: string) => {
    const machine = machines.find((m) => m.serial_number === serial_number);
    if (machine) {
      setMachineToDelete(machine);
      setDeleteError(null);
      setShowDeleteConfirm(true);
    }
  };

  const handleConfirmDelete = async () => {
    if (machineToDelete) {
      try {
        setIsDeletingMachine(true);
        setDeleteError(null);
        const success = await deleteMachine(machineToDelete.serial_number);
        if (success) {
          setMachineToDelete(null);
          setShowDeleteConfirm(false);
          // Refresh the machines list
          refetch();
        } else {
          // Show error message when delete fails
          setDeleteError('Failed to delete machine. Please try again.');
        }
      } catch (error) {
        console.error('Failed to delete machine:', error);
        setDeleteError('Failed to delete machine. Please try again.');
      } finally {
        setIsDeletingMachine(false);
      }
    }
  };

  const handleCancelDelete = () => {
    setMachineToDelete(null);
    setDeleteError(null);
    setShowDeleteConfirm(false);
  };

  const handleAddMachine = async (newMachine: Omit<Machine, 'created_at' | 'updated_at'>) => {
    try {
      setIsCreatingMachine(true);
      const createdMachine = await createMachine(newMachine);
      if (createdMachine) {
        // Refresh the machines list
        refetch();
        setIsCreatingMachine(false);
      } else {
        // Create failed, don't close modal - let the modal handle the error
        setIsCreatingMachine(false);
        throw new Error('Failed to create machine');
      }
    } catch (error) {
      console.error('Failed to create machine:', error);
      setIsCreatingMachine(false);
      // Re-throw the error so the modal can handle it
      throw error;
    }
  };

  const handleEditMachine = async (updatedMachine: Machine) => {
    try {
      setIsUpdatingMachine(true);
      const updated = await updateMachine(updatedMachine.serial_number, updatedMachine);
      if (updated) {
        setMachineToEdit(null);
        setIsMachineModalOpen(false);
        // Refresh the machines list
        refetch();
        setIsUpdatingMachine(false);
      } else {
        // Update failed, don't close modal - let the modal handle the error
        setIsUpdatingMachine(false);
        throw new Error('Failed to update machine');
      }
    } catch (error) {
      console.error('Failed to update machine:', error);
      setIsUpdatingMachine(false);
      // Re-throw the error so the modal can handle it
      throw error;
    }
  };

  const handleOpenAddModal = () => {
    setModalMode('add');
    setMachineToEdit(null);
    setIsMachineModalOpen(true);
  };

  const handleCloseMachineModal = () => {
    setMachineToEdit(null);
    setIsMachineModalOpen(false);
  };

  const handleMachineSubmit = async (machine: Machine | Omit<Machine, 'created_at' | 'updated_at'>) => {
    if (modalMode === 'add') {
      await handleAddMachine(machine as Omit<Machine, 'created_at' | 'updated_at'>);
    } else {
      await handleEditMachine(machine as Machine);
    }
  };

  const getFilterStatusText = () => {
    if (filterType === 'overdue') return ' (overdue only)';
    if (filterType === 'due') return ' (due today only)';
    if (searchOptions.query.trim() && filteredMachines.length !== machines.length) {
      return ` (filtered from ${machines.length} loaded)`;
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

  // Show loading state
  if (loading && machines.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex justify-center items-center py-12">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Loading machines...</p>
          </div>
        </div>
      </div>
    );
  }

  // Show error state
  if (error) {
    return (
      <div className="space-y-6">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">
                Error loading machines
              </h3>
              <div className="mt-2 text-sm text-red-700">
                {error.message}
              </div>
              <div className="mt-4">
                <button
                  onClick={refetch}
                  className="bg-red-100 text-red-800 px-3 py-2 rounded-md text-sm font-medium hover:bg-red-200"
                >
                  Try again
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Loading Overlay for CRUD operations */}
      <LoadingOverlay 
        isVisible={isCreatingMachine || isUpdatingMachine || isDeletingMachine} 
        message={
          isCreatingMachine ? "Creating machine..." :
          isUpdatingMachine ? "Updating machine..." :
          isDeletingMachine ? "Deleting machine..." :
          "Loading..."
        }
      />
      
      <div className="flex justify-between items-center">
        <div className="flex-1">
          <div className="flex items-center gap-4">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">All Machines</h2>
              <p className="text-sm text-gray-500 mt-1">
                {searchOptions.query.trim() ? (
                  <>
                    Showing {machines.length} of {total} machine
                    {total !== 1 ? 's' : ''}
                    {getFilterStatusText()}
                  </>
                ) : (
                  <>
                    Showing {machines.length} of {total} machine
                    {total !== 1 ? 's' : ''}
                    {getFilterStatusText()}
                  </>
                )}
                {loading && ' (updating...)'}
              </p>
            </div>

            {/* Overdue Statistics Badge */}
            {overdueCount > 0 && (
              <div className="flex items-center gap-2">
                {overdueCount > 0 && (
                  <div className="bg-red-100 text-red-800 px-3 py-1 rounded-full text-sm font-medium flex items-center gap-1">
                    <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"
                      />
                    </svg>
                    {overdueCount} Overdue
                  </div>
                )}
                {dueCount > 0 && (
                  <div className="bg-orange-100 text-orange-800 px-3 py-1 rounded-full text-sm font-medium flex items-center gap-1">
                    <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    {dueCount} Due Today
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
            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
            </svg>
            Add New Machine
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredMachines.map((machine) => (
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
      {offset + limit < total && (
        <div className="flex justify-center pt-4">
          <button
            onClick={handleLoadMore}
            disabled={loading}
            className="bg-gray-100 hover:bg-gray-200 disabled:bg-gray-50 disabled:text-gray-400 text-gray-700 px-6 py-3 rounded-lg font-medium transition-colors flex items-center gap-2"
          >
            {loading ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600"></div>
                Loading...
              </>
            ) : (
              `Load More (${total - (offset + limit)} remaining)`
            )}
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

      {/* Machine Modal (Add/Edit) */}
      <MachineModal
        isOpen={isMachineModalOpen}
        mode={modalMode}
        machine={machineToEdit}
        onClose={handleCloseMachineModal}
        onSubmit={handleMachineSubmit}
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
                <div>
                  <h3 className="text-lg font-medium text-gray-900 mb-2 text-center">
                    Delete Machine
                  </h3>
                  <p className="text-sm text-gray-500 mb-2 text-center">
                    Are you sure you want to delete machine{' '}
                    <strong>{machineToDelete.serial_number}</strong>?
                  </p>
                  <p className="text-sm text-gray-500 mb-2 text-center">
                    This action cannot be undone. All data associated with this machine will be
                    permanently removed.
                  </p>
                  {/* Error Display */}
                  {deleteError && (
                    <div className="mt-4 mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md">
                      <div className="flex">
                        <div className="flex-shrink-0">
                          <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                          </svg>
                        </div>
                        <div className="ml-3">
                          <p className="text-sm font-medium">{deleteError}</p>
                        </div>
                      </div>
                    </div>
                  )}
                  {/* Maintenance count warning removed for now - will be implemented when maintenance API is ready */}
                </div>
                <div className="flex space-x-3 mt-6">
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
                    disabled={isDeletingMachine}
                    className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 transition-colors disabled:bg-red-400 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                  >
                    {isDeletingMachine ? (
                      <>
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                        Deleting...
                      </>
                    ) : (
                      'Delete Machine'
                    )}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Full Page Loading Overlay */}
      <FullPageLoader 
        isVisible={isNavigating} 
        message="Loading machine details..." 
      />
    </div>
  );
}
