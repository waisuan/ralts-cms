'use client';

import { Maintenance } from '../types/maintenance';
import MaintenanceCard from './MaintenanceCard';

interface MaintenanceCardListProps {
  records: Maintenance[];
  total: number;
  loading: boolean;
  machineSerialNumber: string;
  searchQuery?: string;
  onLoadMore: () => void;
  onEdit: (record: Maintenance) => void;
  onDelete: (record: Maintenance) => void;
}

export default function MaintenanceCardList({
  records,
  total,
  loading,
  machineSerialNumber,
  searchQuery,
  onLoadMore,
  onEdit,
  onDelete,
}: MaintenanceCardListProps) {
  const remaining = total - records.length;

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

  return (
    <div>
      <p className="text-xs text-gray-500 mb-3">
        Showing {records.length} of {total} record{total !== 1 ? 's' : ''}
        {loading && ' · updating...'}
      </p>

      {loading && records.length === 0 && (
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

      {remaining > 0 && (
        <div className="flex justify-center pt-4">
          <button
            onClick={onLoadMore}
            disabled={loading}
            className="bg-gray-100 hover:bg-gray-200 disabled:bg-gray-50 disabled:text-gray-400 text-gray-700 px-6 py-3 rounded-lg font-medium transition-colors flex items-center gap-2"
          >
            {loading ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600"></div>
                Loading...
              </>
            ) : (
              `Load More (${remaining} remaining)`
            )}
          </button>
        </div>
      )}
    </div>
  );
}
