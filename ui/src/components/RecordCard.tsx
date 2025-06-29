import { useState } from 'react';
import { Machine } from '../types/machine';
import { getPPMStatus } from '../utils/ppmUtils';

interface RecordCardProps {
  machine: Machine;
  onView: (serial_number: string) => void;
  onEdit: (serial_number: string) => void;
  onDelete: (serial_number: string) => void;
}

export default function RecordCard({ machine, onView, onEdit, onDelete }: RecordCardProps) {
  const [showNotesModal, setShowNotesModal] = useState(false);
  const status = getPPMStatus(machine.ppm_date);

  const formatDate = (dateString: string) => {
    if (!dateString) return '-';
    return new Date(dateString).toLocaleDateString();
  };

  const handleDownloadAttachment = () => {
    if (machine.attachment) {
      // In a real app, this would trigger the actual download
      console.log('Downloading attachment:', machine.attachment);
      // For now, just show an alert
      alert(`Downloading attachment: ${machine.attachment}`);
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
              <h3 className="text-lg font-semibold text-gray-900 truncate">
                {machine.serial_number}{' '}
                <span className="text-xs text-gray-500">({machine.model})</span>
              </h3>
              <div className="text-xs text-gray-500 mt-1">
                {machine.brand} &middot; {machine.district}, {machine.state}
              </div>
            </div>
            <div className="flex items-center gap-2">
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

          <div className="text-xs text-gray-500 mb-4 space-y-1">
            <div>TNC Date: {formatDate(machine.tnc_date)}</div>
            <div>PPM Date: {formatDate(machine.ppm_date)}</div>
            <div>Reported By: {machine.reported_by || 'Not specified'}</div>
            <div>Created: {formatDate(machine.created_at)}</div>
            <div>Updated: {formatDate(machine.updated_at)}</div>
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
              <strong>Machine:</strong> {machine.serial_number} ({machine.model})
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
