'use client';

import { Maintenance } from '../types/maintenance';
import MaintenanceCard from './MaintenanceCard';

interface MaintenanceCardListProps {
  records: Maintenance[];
  total: number;
  currentPage: number;
  limit: number;
  loading: boolean;
  machineSerialNumber: string;
  onPageChange: (page: number) => void;
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
  onPageChange,
  onEdit,
  onDelete,
}: MaintenanceCardListProps) {
  const totalPages = Math.ceil(total / limit);

  if (records.length === 0 && !loading) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-400 text-5xl mb-4">📋</div>
        <h3 className="text-lg font-medium text-gray-900 mb-2">No Maintenance Records</h3>
        <p className="text-gray-500">No records found. Add a new maintenance record to get started.</p>
      </div>
    );
  }

  return (
    <div>
      {loading && (
        <div className="flex justify-center py-4">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      )}

      <div className="space-y-3">
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

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between mt-4 pt-4 border-t border-gray-200">
          <button
            onClick={() => onPageChange(currentPage - 1)}
            disabled={currentPage <= 1 || loading}
            className="flex items-center gap-1 px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Prev
          </button>
          <span className="text-sm text-gray-600">
            Page {currentPage} of {totalPages}
          </span>
          <button
            onClick={() => onPageChange(currentPage + 1)}
            disabled={currentPage >= totalPages || loading}
            className="flex items-center gap-1 px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Next
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </div>
      )}
    </div>
  );
}
