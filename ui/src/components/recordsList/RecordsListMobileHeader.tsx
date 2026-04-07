import type { Dispatch, SetStateAction } from 'react';
import type { MachineListFilterType, MachineListSortType as SortType } from '@/utils/machineListFilters';
import { SORT_OPTIONS } from './recordsListConstants';

interface RecordsListMobileHeaderProps {
  filterType: MachineListFilterType;
  filterStatusSuffix: string;
  machinesLength: number;
  total: number;
  overdueCount: number;
  dueCount: number;
  loading: boolean;
  sortBy: SortType;
  showMobileSortMenu: boolean;
  setShowMobileSortMenu: Dispatch<SetStateAction<boolean>>;
  onShowAll?: () => void;
  onSortChange?: (sortBy: SortType) => void;
  onOpenAddModal: () => void;
}

export default function RecordsListMobileHeader({
  filterType,
  filterStatusSuffix,
  machinesLength,
  total,
  overdueCount,
  dueCount,
  loading,
  sortBy,
  showMobileSortMenu,
  setShowMobileSortMenu,
  onShowAll,
  onSortChange,
  onOpenAddModal,
}: RecordsListMobileHeaderProps) {
  return (
    <div className="md:hidden">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-gray-900">Machines</h2>
        <div className="flex items-center gap-2">
          {(filterType === 'overdue' || filterType === 'due') && onShowAll && (
            <button
              type="button"
              onClick={onShowAll}
              className="p-2 border border-gray-300 rounded-lg bg-white text-gray-700 hover:bg-gray-50 transition-colors"
              aria-label="Show all machines"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
          )}
          {onSortChange && (
            <div className="relative">
              <button
                type="button"
                onClick={() => setShowMobileSortMenu((prev) => !prev)}
                className={`p-2 border rounded-lg transition-colors ${
                  showMobileSortMenu
                    ? 'bg-blue-50 border-blue-300 text-blue-700'
                    : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
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
                        type="button"
                        onClick={() => {
                          onSortChange(opt.value);
                          setShowMobileSortMenu(false);
                        }}
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
          <button
            type="button"
            onClick={onOpenAddModal}
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
        {machinesLength}/{total}
        {overdueCount > 0 && <span className="text-red-600"> · {overdueCount} overdue</span>}
        {dueCount > 0 && <span className="text-orange-600"> · {dueCount} due</span>}
        {filterStatusSuffix}
        {loading && ' · updating...'}
      </p>
    </div>
  );
}
