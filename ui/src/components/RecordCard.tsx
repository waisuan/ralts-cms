'use client';

import { useState } from 'react';
import { Machine } from '../types/machine';
import { PPM_STATUSES, PPM_STATUS_COLORS } from '../utils/constants';
import { AttachmentService } from '../services/attachmentService';

interface RecordCardProps {
  machine: Machine;
  onView: (serial_number: string) => void;
  onEdit: (serial_number: string) => void;
  onDelete: (serial_number: string) => void;
}

// Helper function to get PPM status styling based on backend ppm_status field
function getPPMStatusDisplay(ppmStatus: string) {
  if (!ppmStatus) return null;
  
  switch (ppmStatus) {
    case 'overdue':
      return {
        label: 'Overdue',
        color: PPM_STATUS_COLORS[PPM_STATUSES.OVERDUE],
      };
    case 'due':
      return {
        label: 'Due',
        color: PPM_STATUS_COLORS[PPM_STATUSES.DUE],
      };
    case 'almost_due':
      return {
        label: 'Upcoming',
        color: PPM_STATUS_COLORS[PPM_STATUSES.ALMOST_DUE],
      };
    default:
      return null;
  }
}

export default function RecordCard({ machine, onView, onEdit, onDelete }: RecordCardProps) {
  const [showNotesModal, setShowNotesModal] = useState(false);
  const status = getPPMStatusDisplay(machine.ppm_status);

  // Use server-driven maintenance count, fallback to 0 if not available
  const maintenanceCount = machine.maintenance_count ?? 0;

  // Helper function to truncate serial number
  const getTruncatedSerialNumber = (serialNumber: string, maxLength: number = 10) => {
    if (serialNumber.length <= maxLength) {
      return serialNumber;
    }
    return serialNumber.slice(0, maxLength) + '...';
  };

  // Check if serial number is truncated
  const isSerialNumberTruncated = machine.serial_number.length > 10;

  const formatDate = (dateString: string) => {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleDateString();
  };

  const handleDownloadAttachment = async () => {
    if (machine.attachment) {
      try {
        console.log('🔧 Downloading attachment:', machine.attachment);
        const blob = await AttachmentService.downloadMachineAttachment(
          machine.serial_number,
          machine.attachment
        );
        
        // Create download link
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = machine.attachment;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);
        
        console.log('✅ Attachment downloaded successfully:', machine.attachment);
      } catch (error) {
        console.error('❌ Failed to download attachment:', error);
        alert(`Failed to download attachment: ${machine.attachment}`);
      }
    }
  };

  const handleShowNotes = () => {
    setShowNotesModal(true);
  };

  const handleCloseNotes = () => {
    setShowNotesModal(false);
  };

  const handleBackdropClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) {
      setShowNotesModal(false);
    }
  };

  return (
    <>
      <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-200 overflow-hidden">
        <div className="p-6">
          <div className="flex justify-between items-start mb-4">
            <div className="flex-1">
              <h3 
                className="text-lg font-semibold text-gray-900 truncate"
                title={isSerialNumberTruncated ? machine.serial_number : undefined}
              >
                {getTruncatedSerialNumber(machine.serial_number)}{' '}
                <span className="text-xs text-gray-500">({machine.model})</span>
              </h3>
              <div className="text-xs text-gray-500 mt-1">
                {machine.brand} &middot; {machine.district}, {machine.state}
              </div>
            </div>
            <div className="flex items-center gap-2">
              {/* Maintenance History Count Badge */}
              <div
                className="flex items-center gap-1 bg-indigo-100 text-indigo-800 px-2 py-1 rounded-full text-xs font-medium cursor-pointer hover:bg-indigo-200 transition-colors"
                title={`${maintenanceCount} maintenance record${maintenanceCount !== 1 ? 's' : ''} available`}
                onClick={() => onView(machine.serial_number)}
              >
                <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"
                  />
                </svg>
                <span>{maintenanceCount}</span>
              </div>

              {/* Attachment Download Icon */}
              {machine.attachment && (
                <button
                  onClick={handleDownloadAttachment}
                  className="p-1 text-gray-400 hover:text-blue-600 transition-colors"
                  title="Download attachment"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                </button>
              )}

              {/* Notes Modal Button */}
              {machine.additional_notes && (
                <button
                  onClick={handleShowNotes}
                  className="p-1 text-gray-400 hover:text-green-600 transition-colors"
                  title="View additional notes"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                </button>
              )}

              {/* PPM Status Badge */}
              {status && (
                <span className={`px-2 py-1 text-xs font-medium rounded-full ${status.color}`}>
                  {status.label}
                </span>
              )}
            </div>
          </div>

          <div className="mb-4 space-y-2">
            <div className="text-sm text-gray-700 font-medium">
              Customer: <span className="font-normal">{machine.customer}</span>
            </div>
            <div className="text-sm text-gray-700 font-medium">
              Status: <span className="font-normal">{machine.status || 'Not specified'}</span>
            </div>
            <div className="text-sm text-gray-700 font-medium">
              Account Type:{' '}
              <span className="font-normal">{machine.account_type || 'Not specified'}</span>
            </div>
            <div className="text-sm text-gray-700 font-medium">
              Person In Charge: <span className="font-normal">{machine.person_in_charge}</span>
            </div>
          </div>

          <div className="text-xs mb-4 space-y-1">
            <div className="text-gray-500">
              TNC Date: <span className="text-gray-700">{formatDate(machine.tnc_date)}</span>
            </div>
            <div className="text-gray-500">
              PPM Date: <span className="text-gray-700">{formatDate(machine.ppm_date)}</span>
            </div>
            <div className="text-gray-500">
              Reported By:{' '}
              <span className="text-gray-700">{machine.reported_by || 'Not specified'}</span>
            </div>
          </div>

          <div className="flex gap-2">
            <button
              onClick={() => onView(machine.serial_number)}
              className="flex-1 bg-blue-50 hover:bg-blue-100 text-blue-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              View
            </button>
            <button
              onClick={() => onEdit(machine.serial_number)}
              className="flex-1 bg-yellow-50 hover:bg-yellow-100 text-yellow-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              Edit
            </button>
            <button
              onClick={() => onDelete(machine.serial_number)}
              className="flex-1 bg-red-50 hover:bg-red-100 text-red-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              Delete
            </button>
          </div>

          {/* Timestamps Footer */}
          <div className="text-center text-xs mt-3 pt-3 border-t border-gray-100 flex items-center justify-center gap-3">
            <span className="flex items-center gap-1 text-gray-500">
              <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                />
              </svg>
              <span>Created:</span>
              <span className="text-gray-700">{formatDate(machine.created_at)}</span>
            </span>
            <span className="text-gray-400">•</span>
            <span className="flex items-center gap-1 text-gray-500">
              <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                />
              </svg>
              <span>Updated:</span>
              <span className="text-gray-700">{formatDate(machine.updated_at)}</span>
            </span>
          </div>
        </div>
      </div>

      {/* Notes Modal */}
      {showNotesModal && (
        <div
          className="fixed inset-0 backdrop-blur-md flex items-center justify-center z-50"
          onClick={handleBackdropClick}
        >
          <div className="bg-white rounded-lg border-2 border-gray-800 p-6 max-w-md w-full mx-4 max-h-96 overflow-y-auto">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold text-gray-900">Additional Notes</h3>
              <button
                onClick={handleCloseNotes}
                className="text-gray-400 hover:text-gray-600 transition-colors"
              >
                <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
            <div className="text-sm text-gray-700 mb-4">
              <strong>Machine:</strong> 
              <span title={isSerialNumberTruncated ? machine.serial_number : undefined}>
                {getTruncatedSerialNumber(machine.serial_number)}
              </span> ({machine.model})
            </div>
            <div className="text-sm text-gray-700 whitespace-pre-wrap">
              {machine.additional_notes || 'No additional notes available.'}
            </div>
            <div className="mt-6 flex justify-end">
              <button
                onClick={handleCloseNotes}
                className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-4 py-2 rounded-md text-sm font-medium transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
