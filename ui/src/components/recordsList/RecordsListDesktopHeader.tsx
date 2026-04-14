import type { MachineListFilterType, MachineListSortType as SortType } from '@/utils/machineListFilters';
import type { ViewMode } from './recordsListConstants';
import { SORT_OPTIONS } from './recordsListConstants';

interface RecordsListDesktopHeaderProps {
  filterType: MachineListFilterType;
  filterStatusSuffix: string;
  effectiveViewMode: ViewMode;
  total: number;
  overdueCount: number;
  dueCount: number;
  loading: boolean;
  sortBy: SortType;
  showDateFilters: boolean;
  setShowDateFilters: (open: boolean) => void;
  ppmDateRangeFrom?: string;
  tncDateRangeFrom?: string;
  csvExporting: boolean;
  onExportCsv: () => void;
  onViewModeChange: (mode: ViewMode) => void;
  onShowAll?: () => void;
  onSortChange?: (sortBy: SortType) => void;
  onOpenAddModal: () => void;
}

export default function RecordsListDesktopHeader({
  filterType,
  filterStatusSuffix,
  effectiveViewMode,
  total,
  overdueCount,
  dueCount,
  loading,
  sortBy,
  showDateFilters,
  setShowDateFilters,
  ppmDateRangeFrom,
  tncDateRangeFrom,
  csvExporting,
  onExportCsv,
  onViewModeChange,
  onShowAll,
  onSortChange,
  onOpenAddModal,
}: RecordsListDesktopHeaderProps) {
  return (
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
        {effectiveViewMode === 'table' && (
          <p className="text-sm text-gray-500 mt-1">
            {total} machine{total !== 1 ? 's' : ''}
            {filterStatusSuffix}
            {loading && ' (updating...)'}
          </p>
        )}
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <button
          type="button"
          onClick={() => setShowDateFilters(!showDateFilters)}
          className={`flex items-center gap-2 px-3 py-2 rounded-lg font-medium transition-colors text-sm border ${
            showDateFilters || ppmDateRangeFrom || tncDateRangeFrom
              ? 'bg-blue-50 border-blue-300 text-blue-700'
              : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
          }`}
        >
          <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
          </svg>
          Date Filters
          {(ppmDateRangeFrom || tncDateRangeFrom) && (
            <span className="bg-blue-600 text-white text-xs px-1.5 py-0.5 rounded-full">
              {(ppmDateRangeFrom ? 1 : 0) + (tncDateRangeFrom ? 1 : 0)}
            </span>
          )}
        </button>

        {onSortChange && (
          <div className="flex items-center gap-2">
            <label htmlFor="records-list-sort" className="text-sm font-medium text-gray-700">
              Sort By:
            </label>
            <select
              id="records-list-sort"
              value={sortBy}
              onChange={(e) => onSortChange(e.target.value as SortType)}
              className="border border-gray-300 rounded-lg px-3 py-2 text-sm text-black focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white"
            >
              {SORT_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        )}

        <button
          type="button"
          onClick={onExportCsv}
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
            type="button"
            onClick={() => onViewModeChange('table')}
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
            type="button"
            onClick={() => onViewModeChange('cards')}
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
            type="button"
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
          type="button"
          onClick={onOpenAddModal}
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
  );
}
