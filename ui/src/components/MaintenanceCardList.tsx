'use client';

import { Maintenance } from '../types/maintenance';
import MaintenanceCard from './MaintenanceCard';
import PaginationControls from './PaginationControls';

interface MaintenanceCardListProps {
  records: Maintenance[];
  total: number;
  currentPage: number; // 1-indexed
  limit: number;
  loading: boolean;
  machineSerialNumber: string;
  searchQuery?: string;
  onPageChange: (page: number) => void;
  onLimitChange: (limit: number) => void;
  onEdit: (record: Maintenance) => void;
  onDelete: (record: Maintenance) => void;
}

export default function MaintenanceCardList({
  records,
  total,
  currentPage,
  limit,
  loading,
  machineSerialNumber,
  searchQuery,
  onPageChange,
  onLimitChange,
  onEdit,
  onDelete,
}: MaintenanceCardListProps) {
  if (records.length === 0 && !loading) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-400 text-6xl mb-4">📋</div>
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          {searchQuery ? 'No records found' : 'No Maintenance Records'}
        </h3>
        <p className="text-gray-500">
          {searchQuery
            ? `No records match "${searchQuery}". Try a different search term.`
            : 'Add a new maintenance record to get started.'}
        </p>
      </div>
    );
  }

  const pageCount = Math.ceil(total / limit);
  const scrollTop = () => window.scrollTo({ top: 0, behavior: 'smooth' });

  return (
    <div>
      {total > 0 && (
        <div className="bg-white rounded-lg shadow-sm border overflow-hidden mb-4">
          <PaginationControls
            pageIndex={currentPage - 1}
            pageCount={pageCount}
            limit={limit}
            total={total}
            onPageChange={(idx) => {
              onPageChange(idx + 1);
              scrollTop();
            }}
            onPageSizeChange={(size) => {
              onLimitChange(size);
              scrollTop();
            }}
            loading={loading}
          />
        </div>
      )}

      {loading && records.length === 0 && (
        <div className="flex justify-center py-4">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      )}

      <div
        className={`space-y-3 transition-opacity ${
          loading ? 'opacity-50 pointer-events-none' : ''
        }`}
      >
        {records.map((record) => (
          <MaintenanceCard
            key={record.work_order_number}
            record={record}
            machineSerialNumber={machineSerialNumber}
            onEdit={onEdit}
            onDelete={onDelete}
          />
        ))}
      </div>

      {records.length > 0 && (
        <div className="flex justify-center pt-4">
          <button
            type="button"
            onClick={scrollTop}
            className="inline-flex items-center gap-1.5 px-4 py-2 text-sm text-gray-600 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            aria-label="Scroll to top"
          >
            <svg
              className="h-4 w-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 15l7-7 7 7"
              />
            </svg>
            Top
          </button>
        </div>
      )}
    </div>
  );
}
