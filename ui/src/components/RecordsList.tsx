'use client';

import { useState, useMemo, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Machine } from '../types/machine';
import { MachineFilters, MachineService } from '../services/machineService';
import { useMachines } from '../hooks/useMachines';
import { useMachine } from '../hooks/useMachine';
import { SearchOptions } from './SearchBar';
import { isDateProperty } from '../utils/constants';
import { useIsMobile } from '../hooks/useMediaQuery';
import RecordCard from './RecordCard';
import RecordsTable from './RecordsTable';
import MachineModal from './MachineModal';
import FullPageLoader from './FullPageLoader';
import LoadingOverlay from './LoadingOverlay';
import DateRangePicker, { DateRangeValue } from './DateRangePicker';

type FilterType = 'all' | 'overdue' | 'due';
export type SortType = 'newest' | 'oldest' | 'ppm_date_asc' | 'ppm_date_desc' | 'tnc_date_asc' | 'tnc_date_desc';

interface RecordsListProps {
  searchOptions: SearchOptions;
  filterType?: FilterType;
  sortBy?: SortType;
  onShowAll?: () => void;
  onSortChange?: (sortBy: SortType) => void;
  onCountsUpdate?: (overdue: number, due: number) => void;
  onSearchLoadingChange?: (loading: boolean) => void;
}

type ViewMode = 'table' | 'cards';
const VIEW_MODE_STORAGE_KEY = 'ralts-view-mode';
const ITEMS_PER_PAGE_TABLE = 50;

function loadViewMode(): ViewMode {
  if (typeof window === 'undefined') return 'table';
  try {
    const stored = localStorage.getItem(VIEW_MODE_STORAGE_KEY);
    if (stored === 'table' || stored === 'cards') return stored;
  } catch { /* use default */ }
  return 'table';
}

const SORT_TYPE_TO_API: Record<SortType, string> = {
  newest: 'updated_at_desc',
  oldest: 'updated_at_asc',
  ppm_date_asc: 'ppm_date_asc',
  ppm_date_desc: 'ppm_date_desc',
  tnc_date_asc: 'tnc_date_asc',
  tnc_date_desc: 'tnc_date_desc',
};

const API_TO_SORT_TYPE: Record<string, SortType> = {
  updated_at_desc: 'newest',
  updated_at_asc: 'oldest',
  ppm_date_asc: 'ppm_date_asc',
  ppm_date_desc: 'ppm_date_desc',
  tnc_date_asc: 'tnc_date_asc',
  tnc_date_desc: 'tnc_date_desc',
};

const SORT_OPTIONS: ReadonlyArray<{ value: SortType; label: string }> = [
  { value: 'newest', label: 'Newest First' },
  { value: 'oldest', label: 'Oldest First' },
  { value: 'ppm_date_asc', label: 'PPM Date (Earliest)' },
  { value: 'ppm_date_desc', label: 'PPM Date (Latest)' },
  { value: 'tnc_date_asc', label: 'TNC Date (Earliest)' },
  { value: 'tnc_date_desc', label: 'TNC Date (Latest)' },
];

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
  const isMobile = useIsMobile();
  const [viewMode, setViewMode] = useState<ViewMode>(loadViewMode);
  const effectiveViewMode = isMobile ? 'cards' : viewMode;
  const [isMachineModalOpen, setIsMachineModalOpen] = useState(false);
  const [modalMode, setModalMode] = useState<'add' | 'edit'>('add');
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [machineToDelete, setMachineToDelete] = useState<Machine | null>(null);
  const [machineToEdit, setMachineToEdit] = useState<Machine | null>(null);
  const [isNavigating, setIsNavigating] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  
  // Date range filter states
  const [ppmDateRange, setPpmDateRange] = useState<DateRangeValue>({ from: undefined, to: undefined });
  const [tncDateRange, setTncDateRange] = useState<DateRangeValue>({ from: undefined, to: undefined });
  const [showDateFilters, setShowDateFilters] = useState(false);
  const [showMobileSortMenu, setShowMobileSortMenu] = useState(false);

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
      filters.sort = 'updated_at_desc';
    } else if (sortBy === 'oldest') {
      filters.sort = 'updated_at_asc';
    } else {
      // Pass through directly for ppm_date_* and tnc_date_*
      filters.sort = sortBy;
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
    
    // Handle PPM date range filter
    if (ppmDateRange.from) {
      filters.ppm_date_from = ppmDateRange.from;
    }
    if (ppmDateRange.to) {
      filters.ppm_date_to = ppmDateRange.to;
    }
    
    // Handle TNC date range filter
    if (tncDateRange.from) {
      filters.tnc_date_from = tncDateRange.from;
    }
    if (tncDateRange.to) {
      filters.tnc_date_to = tncDateRange.to;
    }
    
    return filters;
  }, [filterType, sortBy, searchOptions, ppmDateRange, tncDateRange]);

  const {
    machines,
    total,
    offset,
    limit,
    loading,
    error,
    refetch,
    loadMore,
    goToPage,
    setLimit,
    overdueCount,
    dueCount,
  } = useMachines({
    page: 1,
    limit: ITEMS_PER_PAGE_TABLE,
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

  const [csvExporting, setCsvExporting] = useState(false);

  const handleExportCSV = useCallback(async () => {
    setCsvExporting(true);
    try {
      await MachineService.exportMachinesCSV(apiFilters);
    } catch (err) {
      console.error('CSV export failed:', err);
      alert('Failed to export CSV. Please try again.');
    } finally {
      setCsvExporting(false);
    }
  }, [apiFilters]);

  const handleViewModeChange = useCallback((mode: ViewMode) => {
    setViewMode(mode);
    try { localStorage.setItem(VIEW_MODE_STORAGE_KEY, mode); } catch { /* ignore */ }
  }, []);

  const handleTableSortChange = useCallback((apiSort: string) => {
    const mapped = API_TO_SORT_TYPE[apiSort];
    if (mapped && onSortChange) onSortChange(mapped);
  }, [onSortChange]);

  const handleTablePageChange = useCallback((page: number) => {
    goToPage(page);
  }, [goToPage]);

  const handleTablePageSizeChange = useCallback((size: number) => {
    setLimit(size);
  }, [setLimit]);

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
      
      {/* Mobile header */}
      <div className="md:hidden">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900">Machines</h2>
          <div className="flex items-center gap-2">
            {/* Show All button when filtering */}
            {(filterType === 'overdue' || filterType === 'due') && onShowAll && (
              <button
                onClick={onShowAll}
                className="p-2 border border-gray-300 rounded-lg bg-white text-gray-700 hover:bg-gray-50 transition-colors"
                aria-label="Show all machines"
              >
                <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                </svg>
              </button>
            )}
            {/* Sort */}
            {onSortChange && (
              <div className="relative">
                <button
                  onClick={() => setShowMobileSortMenu((prev) => !prev)}
                  className={`p-2 border rounded-lg transition-colors ${
                    showMobileSortMenu ? 'bg-blue-50 border-blue-300 text-blue-700' : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
                  }`}
                  aria-label="Sort machines"
                >
                  <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4h13M3 8h9m-9 4h6m4 0l4-4m0 0l4 4m-4-4v12" />
                  </svg>
                </button>
                {showMobileSortMenu && (
                  <>
                    <div className="fixed inset-0 z-10" onClick={() => setShowMobileSortMenu(false)} />
                    <div className="absolute right-0 top-full mt-1 w-48 bg-white border border-gray-300 rounded-lg shadow-lg z-20">
                      {SORT_OPTIONS.map((opt) => (
                        <button
                          key={opt.value}
                          onClick={() => { onSortChange(opt.value); setShowMobileSortMenu(false); }}
                          className={`w-full text-left px-4 py-2 text-sm transition-colors ${
                            sortBy === opt.value ? 'bg-blue-50 text-blue-600 font-medium' : 'text-gray-700 hover:bg-gray-50'
                          }`}
                        >
                          {opt.label}
                        </button>
                      ))}
                    </div>
                  </>
                )}
              </div>
            )}
            {/* Add */}
            <button
              onClick={handleOpenAddModal}
              className="bg-blue-600 hover:bg-blue-700 text-white p-2 rounded-lg transition-colors"
              aria-label="Add New Machine"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </div>
        </div>
        <p className="text-xs text-gray-500 mt-1">
          {machines.length}/{total}
          {overdueCount > 0 && <span className="text-red-600"> · {overdueCount} overdue</span>}
          {dueCount > 0 && <span className="text-orange-600"> · {dueCount} due</span>}
          {getFilterStatusText()}
          {loading && ' · updating...'}
        </p>
      </div>

      {/* Desktop header */}
      <div className="hidden md:flex md:justify-between md:items-center gap-4">
        <div className="flex-1">
          <div className="flex items-center gap-2 flex-wrap">
            <h2 className="text-xl font-semibold text-gray-900">Machines</h2>
            {overdueCount > 0 && (
              <span className="bg-red-100 text-red-800 px-2.5 py-0.5 rounded-full text-xs font-medium inline-flex items-center gap-1">
                <svg className="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
                </svg>
                {overdueCount} Overdue
              </span>
            )}
            {dueCount > 0 && (
              <span className="bg-orange-100 text-orange-800 px-2.5 py-0.5 rounded-full text-xs font-medium inline-flex items-center gap-1">
                <svg className="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {dueCount} Due Today
              </span>
            )}
          </div>
          <p className="text-sm text-gray-500 mt-1">
            Showing {machines.length} of {total} machine{total !== 1 ? 's' : ''}
            {getFilterStatusText()}
            {loading && ' (updating...)'}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <button
            onClick={() => setShowDateFilters(!showDateFilters)}
            className={`flex items-center gap-2 px-3 py-2 rounded-lg font-medium transition-colors text-sm border ${
              showDateFilters || ppmDateRange.from || tncDateRange.from
                ? 'bg-blue-50 border-blue-300 text-blue-700'
                : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
            }`}
          >
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
            </svg>
            Date Filters
            {(ppmDateRange.from || tncDateRange.from) && (
              <span className="bg-blue-600 text-white text-xs px-1.5 py-0.5 rounded-full">
                {(ppmDateRange.from ? 1 : 0) + (tncDateRange.from ? 1 : 0)}
              </span>
            )}
          </button>

          {onSortChange && (
            <div className="flex items-center gap-2">
              <label className="text-sm font-medium text-gray-700">Sort By:</label>
              <select
                value={sortBy}
                onChange={(e) => onSortChange(e.target.value as SortType)}
                className="border border-gray-300 rounded-lg px-3 py-2 text-sm text-black focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white"
              >
                {SORT_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            </div>
          )}

          <button
            onClick={handleExportCSV}
            disabled={csvExporting || loading || total === 0}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm border border-gray-300 rounded-lg bg-white hover:bg-gray-50 text-gray-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            title="Export machines to CSV"
          >
            {csvExporting ? (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600" />
            ) : (
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            )}
            CSV
          </button>

          <div className="flex items-center border border-gray-300 rounded-lg overflow-hidden">
            <button
              onClick={() => handleViewModeChange('table')}
              className={`p-2 transition-colors ${
                effectiveViewMode === 'table' ? 'bg-blue-50 text-blue-700' : 'bg-white text-gray-500 hover:bg-gray-50'
              }`}
              title="Table view"
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M3 14h18M3 6h18M3 18h18" />
              </svg>
            </button>
            <button
              onClick={() => handleViewModeChange('cards')}
              className={`p-2 transition-colors border-l border-gray-300 ${
                effectiveViewMode === 'cards' ? 'bg-blue-50 text-blue-700' : 'bg-white text-gray-500 hover:bg-gray-50'
              }`}
              title="Card view"
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
              </svg>
            </button>
          </div>

          {(filterType === 'overdue' || filterType === 'due') && onShowAll && (
            <button
              onClick={onShowAll}
              className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2"
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
              Show All
            </button>
          )}

          <button
            onClick={handleOpenAddModal}
            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2"
            aria-label="Add New Machine"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Add New Machine
          </button>
        </div>
      </div>

      {/* Date Range Filters Panel */}
      {showDateFilters && (
        <div className="mb-6 p-4 bg-gray-50 border border-gray-200 rounded-lg">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-gray-700">Filter by Date Range</h3>
            {(ppmDateRange.from || tncDateRange.from) && (
              <button
                onClick={() => {
                  setPpmDateRange({ from: undefined, to: undefined });
                  setTncDateRange({ from: undefined, to: undefined });
                }}
                className="text-sm text-blue-600 hover:text-blue-800 font-medium"
              >
                Clear All Filters
              </button>
            )}
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <DateRangePicker
              label="PPM Date Range"
              value={ppmDateRange}
              onChange={setPpmDateRange}
              placeholder="Filter by PPM date..."
            />
            <DateRangePicker
              label="TNC Date Range"
              value={tncDateRange}
              onChange={setTncDateRange}
              placeholder="Filter by TNC date..."
            />
          </div>
          {(ppmDateRange.from || tncDateRange.from) && (
            <div className="mt-3 flex flex-wrap gap-2">
              {ppmDateRange.from && (
                <span className="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
                  PPM: {ppmDateRange.from === ppmDateRange.to || !ppmDateRange.to
                    ? ppmDateRange.from
                    : `${ppmDateRange.from} to ${ppmDateRange.to}`}
                  <button
                    onClick={() => setPpmDateRange({ from: undefined, to: undefined })}
                    className="hover:bg-blue-200 rounded-full p-0.5"
                  >
                    <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </span>
              )}
              {tncDateRange.from && (
                <span className="inline-flex items-center gap-1 px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full">
                  TNC: {tncDateRange.from === tncDateRange.to || !tncDateRange.to
                    ? tncDateRange.from
                    : `${tncDateRange.from} to ${tncDateRange.to}`}
                  <button
                    onClick={() => setTncDateRange({ from: undefined, to: undefined })}
                    className="hover:bg-green-200 rounded-full p-0.5"
                  >
                    <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </span>
              )}
            </div>
          )}
        </div>
      )}

      {effectiveViewMode === 'table' ? (
        <RecordsTable
          machines={filteredMachines}
          total={total}
          offset={offset}
          limit={limit}
          loading={loading}
          sortBy={SORT_TYPE_TO_API[sortBy] || 'updated_at_desc'}
          onSortChange={handleTableSortChange}
          onPageChange={handleTablePageChange}
          onPageSizeChange={handleTablePageSizeChange}
          onView={handleView}
          onEdit={handleEdit}
          onDelete={handleDelete}
        />
      ) : (
        <>
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
        </>
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
