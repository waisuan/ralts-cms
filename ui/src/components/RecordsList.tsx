'use client';

import { useState, useMemo, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Machine } from '../types/machine';
import { MachineFilters, MachineService } from '../services/machineService';
import { useMachines } from '../hooks/useMachines';
import { useMachine } from '../hooks/useMachine';
import { SearchOptions } from './SearchBar';
import { useIsMobile } from '../hooks/useMediaQuery';
import RecordCard from './RecordCard';
import RecordsTable from './RecordsTable';
import PaginationControls from './PaginationControls';
import MachineModal from './MachineModal';
import FullPageLoader from './FullPageLoader';
import LoadingOverlay from './LoadingOverlay';
import { DateRangeValue } from './DateRangePicker';
import {
  buildMachineListFilters,
  type MachineListFilterType as FilterType,
  type MachineListSortType,
} from '@/utils/machineListFilters';
import {
  ITEMS_PER_PAGE_TABLE,
  VIEW_MODE_STORAGE_KEY,
  loadViewMode,
  SORT_TYPE_TO_API,
  API_TO_SORT_TYPE,
  type ViewMode,
} from './recordsList/recordsListConstants';
import { getRecordsListFilterStatusSuffix, getRecordsListEmptyState } from './recordsList/recordsListCopy';
import RecordsListLoadingState from './recordsList/RecordsListLoadingState';
import RecordsListErrorState from './recordsList/RecordsListErrorState';
import RecordsListMobileHeader from './recordsList/RecordsListMobileHeader';
import RecordsListDesktopHeader from './recordsList/RecordsListDesktopHeader';
import RecordsListDateFiltersPanel from './recordsList/RecordsListDateFiltersPanel';
import RecordsListDeleteModal from './recordsList/RecordsListDeleteModal';

export type SortType = MachineListSortType;

interface RecordsListProps {
  searchOptions: SearchOptions;
  filterType?: FilterType;
  sortBy?: SortType;
  onShowAll?: () => void;
  onSortChange?: (sortBy: SortType) => void;
  onCountsUpdate?: (overdue: number, due: number) => void;
  onSearchLoadingChange?: (loading: boolean) => void;
  onTotalChange?: (total: number) => void;
}

export default function RecordsList({
  searchOptions,
  filterType = 'all',
  sortBy = 'newest',
  onShowAll,
  onSortChange,
  onCountsUpdate,
  onSearchLoadingChange,
  onTotalChange,
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

  const [ppmDateRange, setPpmDateRange] = useState<DateRangeValue>({ from: undefined, to: undefined });
  const [tncDateRange, setTncDateRange] = useState<DateRangeValue>({ from: undefined, to: undefined });
  const [showDateFilters, setShowDateFilters] = useState(false);
  const [showMobileSortMenu, setShowMobileSortMenu] = useState(false);

  const [isDeletingMachine, setIsDeletingMachine] = useState(false);
  const [isCreatingMachine, setIsCreatingMachine] = useState(false);
  const [isUpdatingMachine, setIsUpdatingMachine] = useState(false);

  const apiFilters: MachineFilters = useMemo(
    () =>
      buildMachineListFilters({
        filterType,
        sortBy,
        searchOptions,
        ppmDateRange,
        tncDateRange,
      }),
    [filterType, sortBy, searchOptions, ppmDateRange, tncDateRange]
  );

  const {
    machines,
    total,
    offset,
    limit,
    loading,
    error,
    refetch,
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

  const { createMachine, updateMachine, deleteMachine } = useMachine();

  useEffect(() => {
    if (onCountsUpdate) {
      onCountsUpdate(overdueCount, dueCount);
    }
  }, [overdueCount, dueCount, onCountsUpdate]);

  useEffect(() => {
    if (onSearchLoadingChange) {
      onSearchLoadingChange(loading);
    }
  }, [loading, onSearchLoadingChange]);

  useEffect(() => {
    if (onTotalChange) {
      onTotalChange(total);
    }
  }, [total, onTotalChange]);

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
          refetch();
        } else {
          setDeleteError('Failed to delete machine. Please try again.');
        }
      } catch (e) {
        console.error('Failed to delete machine:', e);
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
        refetch();
        setIsCreatingMachine(false);
      } else {
        setIsCreatingMachine(false);
        throw new Error('Failed to create machine');
      }
    } catch (e) {
      console.error('Failed to create machine:', e);
      setIsCreatingMachine(false);
      throw e;
    }
  };

  const handleEditMachine = async (updatedMachine: Machine) => {
    try {
      setIsUpdatingMachine(true);
      const updated = await updateMachine(updatedMachine.serial_number, updatedMachine);
      if (updated) {
        setMachineToEdit(null);
        setIsMachineModalOpen(false);
        refetch();
        setIsUpdatingMachine(false);
      } else {
        setIsUpdatingMachine(false);
        throw new Error('Failed to update machine');
      }
    } catch (e) {
      console.error('Failed to update machine:', e);
      setIsUpdatingMachine(false);
      throw e;
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
    try { localStorage.setItem(VIEW_MODE_STORAGE_KEY, mode); } catch {}
  }, []);

  const handleTableSortChange = useCallback(
    (apiSort: string) => {
      const mapped = API_TO_SORT_TYPE[apiSort];
      if (mapped && onSortChange) onSortChange(mapped);
    },
    [onSortChange]
  );

  const handlePageChange = useCallback(
    (page: number) => {
      goToPage(page);
    },
    [goToPage]
  );

  const handlePageSizeChange = useCallback(
    (size: number) => {
      setLimit(size);
    },
    [setLimit]
  );

  const filterStatusSuffix = getRecordsListFilterStatusSuffix(filterType);
  const emptyState = getRecordsListEmptyState(filterType, searchOptions.query);

  const clearDateRanges = useCallback(() => {
    setPpmDateRange({ from: undefined, to: undefined });
    setTncDateRange({ from: undefined, to: undefined });
  }, []);

  if (loading && machines.length === 0) {
    return <RecordsListLoadingState />;
  }

  if (error) {
    return <RecordsListErrorState error={error} onRetry={refetch} />;
  }

  return (
    <div className="space-y-6">
      <LoadingOverlay
        isVisible={isCreatingMachine || isUpdatingMachine || isDeletingMachine}
        message={
          isCreatingMachine
            ? 'Creating machine...'
            : isUpdatingMachine
              ? 'Updating machine...'
              : isDeletingMachine
                ? 'Deleting machine...'
                : 'Loading...'
        }
      />

      <RecordsListMobileHeader
        filterType={filterType}
        filterStatusSuffix={filterStatusSuffix}
        machinesLength={machines.length}
        total={total}
        overdueCount={overdueCount}
        dueCount={dueCount}
        loading={loading}
        sortBy={sortBy}
        showMobileSortMenu={showMobileSortMenu}
        setShowMobileSortMenu={setShowMobileSortMenu}
        onShowAll={onShowAll}
        onSortChange={onSortChange}
        onOpenAddModal={handleOpenAddModal}
      />

      <RecordsListDesktopHeader
        filterType={filterType}
        filterStatusSuffix={filterStatusSuffix}
        effectiveViewMode={effectiveViewMode}
        total={total}
        overdueCount={overdueCount}
        dueCount={dueCount}
        loading={loading}
        sortBy={sortBy}
        showDateFilters={showDateFilters}
        setShowDateFilters={setShowDateFilters}
        ppmDateRangeFrom={ppmDateRange.from}
        tncDateRangeFrom={tncDateRange.from}
        csvExporting={csvExporting}
        onExportCsv={handleExportCSV}
        onViewModeChange={handleViewModeChange}
        onShowAll={onShowAll}
        onSortChange={onSortChange}
        onOpenAddModal={handleOpenAddModal}
      />

      {showDateFilters && (
        <RecordsListDateFiltersPanel
          ppmDateRange={ppmDateRange}
          tncDateRange={tncDateRange}
          onPpmChange={setPpmDateRange}
          onTncChange={setTncDateRange}
          onClearAll={clearDateRanges}
        />
      )}

      {effectiveViewMode === 'table' ? (
        <RecordsTable
          machines={machines}
          total={total}
          offset={offset}
          limit={limit}
          loading={loading}
          sortBy={SORT_TYPE_TO_API[sortBy] || 'updated_at_desc'}
          searchQuery={searchOptions.query.trim() || undefined}
          onSortChange={handleTableSortChange}
          onPageChange={handlePageChange}
          onPageSizeChange={handlePageSizeChange}
          onView={handleView}
          onEdit={handleEdit}
          onDelete={handleDelete}
        />
      ) : (
        <>
          {total > 0 && (
            <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
              <PaginationControls
                pageIndex={Math.floor(offset / limit)}
                pageCount={Math.ceil(total / limit)}
                limit={limit}
                total={total}
                onPageChange={(idx) => { handlePageChange(idx); window.scrollTo({ top: 0, behavior: 'smooth' }); }}
                onPageSizeChange={(size) => { handlePageSizeChange(size); window.scrollTo({ top: 0, behavior: 'smooth' }); }}
                loading={loading}
              />
            </div>
          )}

          <div className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 transition-opacity ${loading ? 'opacity-50 pointer-events-none' : ''}`}>
            {machines.map((machine) => (
              <RecordCard
                key={machine.serial_number}
                machine={machine}
                onView={handleView}
                onEdit={handleEdit}
                onDelete={handleDelete}
              />
            ))}
          </div>

          {machines.length > 0 && (
            <div className="flex justify-center gap-3 pt-2">
              <button
                type="button"
                onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
                className="inline-flex items-center gap-1.5 px-4 py-2 text-sm text-gray-600 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                aria-label="Scroll to top"
              >
                <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
                </svg>
                Top
              </button>
              <button
                type="button"
                onClick={() => window.scrollTo({ top: document.documentElement.scrollHeight, behavior: 'smooth' })}
                className="inline-flex items-center gap-1.5 px-4 py-2 text-sm text-gray-600 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                aria-label="Scroll to bottom"
              >
                <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
                Bottom
              </button>
            </div>
          )}

          {machines.length === 0 && (
            <div className="text-center py-12">
              <div className="text-gray-400 text-6xl mb-4">📄</div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">{emptyState.title}</h3>
              <p className="text-gray-500">{emptyState.subtitle}</p>
            </div>
          )}
        </>
      )}

      <MachineModal
        isOpen={isMachineModalOpen}
        mode={modalMode}
        machine={machineToEdit}
        onClose={handleCloseMachineModal}
        onSubmit={handleMachineSubmit}
      />

      {showDeleteConfirm && machineToDelete && (
        <RecordsListDeleteModal
          machine={machineToDelete}
          deleteError={deleteError}
          isDeleting={isDeletingMachine}
          onCancel={handleCancelDelete}
          onConfirm={handleConfirmDelete}
        />
      )}

      <FullPageLoader isVisible={isNavigating} message="Loading machine details..." />
    </div>
  );
}
