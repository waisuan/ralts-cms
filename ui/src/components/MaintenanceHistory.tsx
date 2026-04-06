'use client';

import { useState } from 'react';
import { Machine } from '../types/machine';
import { Maintenance } from '../types/maintenance';
import { MachineService } from '../services/machineService';
import { useMaintenance } from '../hooks/useMaintenance';
import { useIsMobile } from '../hooks/useMediaQuery';
import MachineInfoCard from './MachineInfoCard';
import MaintenanceTable from './MaintenanceTable';
import MaintenanceCardList from './MaintenanceCardList';
import AddMaintenanceModal from './AddMaintenanceModal';
import EditMaintenanceModal from './EditMaintenanceModal';
import DeleteMaintenanceConfirm from './DeleteMaintenanceConfirm';
import FullPageLoader from './FullPageLoader';

interface MaintenanceHistoryProps {
  machine: Machine;
  onBack: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
}

const MAINTENANCE_SORT_OPTIONS = [
  { value: 'updated_at_desc', label: 'Newest First' },
  { value: 'updated_at_asc', label: 'Oldest First' },
  { value: 'work_order_date_desc', label: 'Work Order Date (Latest)' },
  { value: 'work_order_date_asc', label: 'Work Order Date (Earliest)' },
] as const;

export default function MaintenanceHistory({
  machine,
  onBack,
  onEdit,
  onDelete,
}: MaintenanceHistoryProps) {
  const maint = useMaintenance({ serialNumber: machine.serial_number });
  const isMobile = useIsMobile();

  // CSV export
  const [csvExporting, setCsvExporting] = useState(false);
  const [showMobileSortMenu, setShowMobileSortMenu] = useState(false);

  // Modal coordination
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<Maintenance | null>(null);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [deletingRecord, setDeletingRecord] = useState<Maintenance | null>(null);

  // Navigation
  const [isNavigatingBack, setIsNavigatingBack] = useState(false);

  const handleBackNavigation = () => {
    setIsNavigatingBack(true);
    onBack();
  };

  const handleExportMaintenanceCSV = async () => {
    setCsvExporting(true);
    try {
      await MachineService.exportMaintenanceCSV(machine.serial_number, maint.debouncedSearchQuery || undefined);
    } catch (err) {
      console.error('Maintenance CSV export failed:', err);
      alert('Failed to export maintenance CSV. Please try again.');
    } finally {
      setCsvExporting(false);
    }
  };

  const handleOpenEdit = (record: Maintenance) => {
    setEditingRecord(record);
    setIsEditModalOpen(true);
  };

  const handleCloseEdit = () => {
    setIsEditModalOpen(false);
    setEditingRecord(null);
  };

  const handleOpenDelete = (record: Maintenance) => {
    setDeletingRecord(record);
    setIsDeleteModalOpen(true);
  };

  const handleCloseDelete = () => {
    setIsDeleteModalOpen(false);
    setDeletingRecord(null);
  };

  const existingWorkOrderNumbers = maint.records.map((r) => r.work_order_number);

  // Loading state
  if (maint.isInitialLoading && maint.records.length === 0) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="container mx-auto px-4">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading maintenance records...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Error state (no records loaded)
  if (maint.error && maint.records.length === 0) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="container mx-auto px-4">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="text-red-600 text-6xl mb-4">⚠️</div>
              <h3 className="text-lg font-medium text-gray-900 mb-2">Error Loading Maintenance Records</h3>
              <p className="text-gray-500 mb-4">{maint.error}</p>
              <button
                onClick={() => maint.refetch()}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                Try Again
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-4 md:py-8">
      <div className="container mx-auto px-4">
        {/* Header */}
        <div className="mb-4 md:mb-8">
          <div className="flex items-center gap-4 mb-4">
            <button
              onClick={handleBackNavigation}
              className="flex items-center gap-2 text-gray-600 hover:text-gray-900 transition-colors"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
              Back to Machines
            </button>
          </div>

          <div className="hidden md:block">
            <MachineInfoCard machine={machine} onEdit={onEdit} onDelete={onDelete} />
          </div>

          {/* Summary Statistics (hidden on mobile) */}
          <div className="hidden md:grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-4 mb-6">
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-gray-900">{maint.total}</div>
              <div className="text-sm text-gray-600">Total Records</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-green-600">{maint.preventativeCount}</div>
              <div className="text-sm text-gray-600">Preventive</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-red-600">{maint.emergencyCount}</div>
              <div className="text-sm text-gray-600">Emergency</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-blue-600">{maint.correctiveCount}</div>
              <div className="text-sm text-gray-600">Corrective</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-purple-600">{maint.inspectionCount}</div>
              <div className="text-sm text-gray-600">Inspection</div>
            </div>
            <div className="bg-white rounded-lg shadow-sm border p-4">
              <div className="text-2xl font-bold text-gray-600">{maint.otherCount}</div>
              <div className="text-sm text-gray-600">Other</div>
            </div>
          </div>

          {/* Mobile: Title + Sort + Add */}
          <div className="md:hidden mb-4">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold text-gray-900">Maintenance</h2>
              <div className="flex items-center gap-2">
                <div className="relative">
                  <button
                    onClick={() => setShowMobileSortMenu((prev) => !prev)}
                    className={`p-2 border rounded-lg transition-colors ${
                      showMobileSortMenu ? 'bg-blue-50 border-blue-300 text-blue-700' : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
                    }`}
                    aria-label="Sort records"
                  >
                    <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4h13M3 8h9m-9 4h6m4 0l4-4m0 0l4 4m-4-4v12" />
                    </svg>
                  </button>
                  {showMobileSortMenu && (
                    <>
                      <div className="fixed inset-0 z-10" onClick={() => setShowMobileSortMenu(false)} />
                      <div className="absolute right-0 top-full mt-1 w-56 bg-white border border-gray-300 rounded-lg shadow-lg z-20">
                        {MAINTENANCE_SORT_OPTIONS.map((opt) => (
                          <button
                            key={opt.value}
                            onClick={() => { maint.setSort(opt.value); setShowMobileSortMenu(false); }}
                            className={`w-full text-left px-4 py-2 text-sm transition-colors ${
                              maint.sort === opt.value ? 'bg-blue-50 text-blue-600 font-medium' : 'text-gray-700 hover:bg-gray-50'
                            }`}
                          >
                            {opt.label}
                          </button>
                        ))}
                      </div>
                    </>
                  )}
                </div>
                <button
                  onClick={() => setIsAddModalOpen(true)}
                  className="bg-blue-600 hover:bg-blue-700 text-white p-2 rounded-lg transition-colors"
                  aria-label="Add New Record"
                >
                  <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          {/* Desktop: Section Title */}
          <div className="hidden md:block mb-6">
            <h2 className="text-xl font-semibold text-gray-900">Maintenance History</h2>
          </div>
        </div>

        {/* Search Bar — centered like machines page */}
        <div className="flex justify-center mb-6">
          <div className="w-full max-w-2xl relative">
            <div className="flex items-center bg-white border border-gray-300 rounded-lg shadow-sm focus-within:ring-2 focus-within:ring-blue-500 focus-within:border-transparent">
              <div className="flex-1 relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  {maint.isSearchLoading ? (
                    <svg className="h-5 w-5 text-gray-400 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                  ) : (
                    <svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                  )}
                </div>
                <input
                  type="text"
                  placeholder="Search maintenance records..."
                  value={maint.searchQuery}
                  onChange={(e) => maint.setSearchQuery(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 border-0 rounded-lg focus:outline-none placeholder-gray-400 text-gray-900"
                />
                {maint.hasActiveSearch && !maint.isSearchLoading && (
                  <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
                    <button
                      type="button"
                      onClick={maint.clearSearch}
                      className="text-gray-400 hover:text-gray-600 focus:outline-none focus:text-gray-600 transition-colors"
                      title="Clear search"
                    >
                      <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                )}
              </div>
            </div>
            {maint.hasActiveSearch && (
              <div className="mt-2 text-sm text-gray-500 text-center">
                Searching for &ldquo;{maint.searchQuery}&rdquo;
                {' · '}{maint.total} result{maint.total !== 1 ? 's' : ''} found
              </div>
            )}
          </div>
        </div>

        {/* Desktop toolbar: CSV + Add Record (sorting via table column headers) */}
        <div className="hidden md:flex justify-end items-center gap-3 mb-3">
          <button
            onClick={handleExportMaintenanceCSV}
            disabled={csvExporting || maint.records.length === 0}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm border border-gray-300 rounded-lg bg-white hover:bg-gray-50 text-gray-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            title="Export maintenance to CSV"
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
          <button
            onClick={() => setIsAddModalOpen(true)}
            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2"
            aria-label="Add New Record"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Add New Record
          </button>
        </div>

        {/* Inline error banner */}
        {maint.error && maint.records.length > 0 && (
          <div className="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg flex items-center justify-between">
            <div className="flex items-center gap-2">
              <svg className="h-5 w-5 text-red-400 flex-shrink-0" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
              <span className="text-sm font-medium">{maint.error}</span>
            </div>
            <button onClick={() => maint.refetch()} className="text-sm text-red-600 hover:text-red-800 font-medium underline">Retry</button>
          </div>
        )}

        {/* Maintenance Table / Cards */}
        {isMobile ? (
          <MaintenanceCardList
            records={maint.records}
            total={maint.total}
            loading={maint.isInitialLoading || maint.isSearchLoading || maint.isPaginationLoading}
            machineSerialNumber={machine.serial_number}
            searchQuery={maint.debouncedSearchQuery}
            onLoadMore={maint.loadMore}
            onEdit={handleOpenEdit}
            onDelete={handleOpenDelete}
          />
        ) : (
          <MaintenanceTable
            machineSerialNumber={machine.serial_number}
            records={maint.records}
            total={maint.total}
            currentPage={maint.currentPage}
            limit={maint.limit}
            loading={maint.isInitialLoading || maint.isSearchLoading || maint.isPaginationLoading}
            sortBy={maint.sort}
            searchQuery={maint.debouncedSearchQuery}
            onSortChange={maint.setSort}
            onPageChange={maint.goToPage}
            onPageSizeChange={maint.setLimit}
            onEdit={handleOpenEdit}
            onDelete={handleOpenDelete}
          />
        )}
      </div>

      {/* Modals */}
      <AddMaintenanceModal
        machineSerialNumber={machine.serial_number}
        existingWorkOrderNumbers={existingWorkOrderNumbers}
        isOpen={isAddModalOpen}
        onClose={() => setIsAddModalOpen(false)}
        onSuccess={() => maint.refetch()}
      />

      <EditMaintenanceModal
        machineSerialNumber={machine.serial_number}
        record={editingRecord}
        existingWorkOrderNumbers={existingWorkOrderNumbers}
        isOpen={isEditModalOpen}
        onClose={handleCloseEdit}
        onSuccess={() => maint.refetch()}
      />

      <DeleteMaintenanceConfirm
        machineSerialNumber={machine.serial_number}
        record={deletingRecord}
        isOpen={isDeleteModalOpen}
        onClose={handleCloseDelete}
        onSuccess={() => maint.refetch()}
      />

      <FullPageLoader isVisible={isNavigatingBack} message="Loading machines list..." />
    </div>
  );
}
