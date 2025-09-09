import { useState, useEffect, useRef } from 'react';
import { UserStatus, USER_STATUS } from '@/services/adminUserService';
import StatusBadge from './StatusBadge';

interface BulkActionsProps {
  selectedUserIds: number[];
  onBulkStatusUpdate: (userIds: number[], status: UserStatus) => Promise<void>;
  onClearSelection: () => void;
  isUpdating: boolean;
}

export default function BulkActions({
  selectedUserIds,
  onBulkStatusUpdate,
  onClearSelection,
  isUpdating
}: BulkActionsProps) {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const [selectedStatus, setSelectedStatus] = useState<UserStatus | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Handle click outside to close dropdown
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    };

    if (isDropdownOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => {
        document.removeEventListener('mousedown', handleClickOutside);
      };
    }
  }, [isDropdownOpen]);

  const handleStatusSelect = (status: UserStatus) => {
    setSelectedStatus(status);
    setIsDropdownOpen(false);
    setShowConfirmDialog(true);
  };

  const handleConfirmUpdate = async () => {
    if (!selectedStatus) return;

    try {
      await onBulkStatusUpdate(selectedUserIds, selectedStatus);
      setShowConfirmDialog(false);
      setSelectedStatus(null);
      onClearSelection();
    } catch (error) {
      console.error('Bulk update failed:', error);
      setShowConfirmDialog(false);
      setSelectedStatus(null);
    }
  };

  const handleCancelUpdate = () => {
    setShowConfirmDialog(false);
    setSelectedStatus(null);
  };

  if (selectedUserIds.length === 0) {
    return null;
  }

  return (
    <>
      {/* Bulk Actions Bar */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center">
            <svg className="h-5 w-5 text-blue-600 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span className="text-sm font-medium text-blue-900">
              {selectedUserIds.length} user{selectedUserIds.length !== 1 ? 's' : ''} selected
            </span>
          </div>
          
          <div className="flex items-center space-x-3">
            {/* Bulk Status Update Dropdown */}
            <div className="relative" ref={dropdownRef}>
              <button
                type="button"
                onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                disabled={isUpdating}
                className="inline-flex items-center px-3 py-2 border border-blue-300 shadow-sm text-sm font-medium rounded-md text-blue-700 bg-white hover:bg-blue-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
              >
                {isUpdating ? (
                  <>
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-blue-600" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Updating...
                  </>
                ) : (
                  <>
                    <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                    Change Status
                    <svg className="ml-1 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                    </svg>
                  </>
                )}
              </button>

              {isDropdownOpen && !isUpdating && (
                <div className="absolute right-0 mt-2 w-56 bg-white rounded-md shadow-lg z-50 border border-gray-200">
                  <div className="py-1">
                    <div className="px-4 py-2 text-xs font-medium text-gray-500 uppercase tracking-wide border-b border-gray-100">
                      Update {selectedUserIds.length} user{selectedUserIds.length !== 1 ? 's' : ''} to:
                    </div>
                    {Object.values(USER_STATUS).map((status) => (
                      <button
                        key={status}
                        onClick={() => handleStatusSelect(status)}
                        className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 focus:outline-none focus:bg-gray-100"
                      >
                        <StatusBadge status={status} />
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {/* Clear Selection Button */}
            <button
              type="button"
              onClick={onClearSelection}
              disabled={isUpdating}
              className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
            >
              Clear Selection
            </button>
          </div>
        </div>
      </div>

      {/* Confirmation Dialog */}
      {showConfirmDialog && selectedStatus && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
            <div className="mt-3">
              <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-yellow-100">
                <svg className="h-6 w-6 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
                </svg>
              </div>
              <div className="mt-2 text-center">
                <h3 className="text-lg leading-6 font-medium text-gray-900">Confirm Bulk Status Update</h3>
                <div className="mt-2">
                  <p className="text-sm text-gray-500">
                    Are you sure you want to update <span className="font-medium">{selectedUserIds.length}</span> user
                    {selectedUserIds.length !== 1 ? 's' : ''} to{' '}
                    <StatusBadge status={selectedStatus} className="mx-1" />?
                  </p>
                  <p className="text-xs text-gray-400 mt-2">
                    This action cannot be undone.
                  </p>
                </div>
              </div>
            </div>
            <div className="items-center px-4 py-3">
              <div className="flex space-x-3">
                <button
                  onClick={handleCancelUpdate}
                  className="flex-1 px-4 py-2 bg-gray-500 text-white text-base font-medium rounded-md shadow-sm hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-300"
                >
                  Cancel
                </button>
                <button
                  onClick={handleConfirmUpdate}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white text-base font-medium rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-300"
                >
                  Confirm Update
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
