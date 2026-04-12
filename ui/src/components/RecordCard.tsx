'use client';

import { useState, useEffect } from 'react';
import { Machine } from '../types/machine';
import { PPM_STATUSES, PPM_STATUS_COLORS } from '../utils/constants';
import { AttachmentService } from '../services/attachmentService';
import { formatDate, formatDateTime } from '../utils/formatters';
import { useIsMobile } from '../hooks/useMediaQuery';

interface RecordCardProps {
  machine: Machine;
  onView: (serial_number: string) => void;
  onEdit: (serial_number: string) => void;
  onDelete: (serial_number: string) => void;
}

function getPPMStatusDisplay(ppmStatus: string) {
  if (!ppmStatus) return null;
  switch (ppmStatus) {
    case 'overdue':
      return { label: 'Overdue', color: PPM_STATUS_COLORS[PPM_STATUSES.OVERDUE] };
    case 'due':
      return { label: 'Due', color: PPM_STATUS_COLORS[PPM_STATUSES.DUE] };
    case 'almost_due':
      return { label: 'Upcoming', color: PPM_STATUS_COLORS[PPM_STATUSES.ALMOST_DUE] };
    default:
      return null;
  }
}

function DetailField({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">{label}</span>
      <p className="text-gray-900 mt-0.5">{value || '-'}</p>
    </div>
  );
}

export default function RecordCard({ machine, onView, onEdit, onDelete }: RecordCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [showNotesModal, setShowNotesModal] = useState(false);
  const isMobile = useIsMobile();
  const status = getPPMStatusDisplay(machine.ppm_status);
  const maintenanceCount = machine.maintenance_count ?? 0;

  const handleDownloadAttachment = async () => {
    if (!machine.attachment || downloading) return;
    setDownloading(true);
    try {
      const blob = await AttachmentService.downloadMachineAttachment(
        machine.serial_number,
        machine.attachment
      );
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = machine.attachment;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to download attachment:', error);
      alert(`Failed to download attachment: ${machine.attachment}`);
    } finally {
      setDownloading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      setIsExpanded(!isExpanded);
    }
  };

  const handleBackdropClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) setShowNotesModal(false);
  };

  useEffect(() => {
    if (!showNotesModal) return;
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setShowNotesModal(false);
    };
    document.addEventListener('keydown', handleEsc);
    return () => document.removeEventListener('keydown', handleEsc);
  }, [showNotesModal]);

  if (isMobile) {
    return (
      <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
        <div
          role="button"
          tabIndex={0}
          aria-expanded={isExpanded}
          onClick={() => setIsExpanded(!isExpanded)}
          onKeyDown={handleKeyDown}
          className="p-4 cursor-pointer hover:bg-gray-50 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset"
        >
          <div className="flex items-start justify-between gap-2">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 flex-wrap">
                <span className="font-semibold text-gray-900 text-sm truncate">
                  {machine.serial_number}
                  {machine.model && <span className="text-gray-500 font-normal"> ({machine.model})</span>}
                </span>
                {status && (
                  <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${status.color}`}>
                    {status.label}
                  </span>
                )}
              </div>
              <p className="text-xs text-gray-500 mt-1">
                {machine.customer} &middot; {machine.state}
              </p>
              {!isExpanded && (
                <p className="text-xs text-gray-500 mt-1">
                  TNC: {machine.tnc_date ? formatDate(machine.tnc_date) : '-'}
                  {' · '}
                  PPM: {machine.ppm_date ? formatDate(machine.ppm_date) : '-'}
                </p>
              )}
            </div>
            <svg
              className={`h-5 w-5 text-gray-400 flex-shrink-0 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
              fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </div>
        </div>

        {isExpanded && (
          <div className="px-4 pb-4 border-t border-gray-100 pt-3 space-y-3">
            <div className="grid grid-cols-2 gap-3 text-sm">
              <DetailField label="TNC Date" value={machine.tnc_date ? formatDate(machine.tnc_date) : '-'} />
              <DetailField label="PPM Date" value={machine.ppm_date ? formatDate(machine.ppm_date) : '-'} />
              {machine.model && <DetailField label="Model" value={machine.model} />}
              {machine.brand && <DetailField label="Brand" value={machine.brand} />}
              <DetailField label="District" value={machine.district || '-'} />
              <DetailField label="Status" value={machine.status || '-'} />
              <DetailField label="Account Type" value={machine.account_type || '-'} />
              <DetailField label="Person In Charge" value={machine.person_in_charge || '-'} />
              <DetailField label="Reported By" value={machine.reported_by || '-'} />
              <DetailField label="Created" value={formatDateTime(machine.created_at)} />
              <DetailField label="Updated" value={formatDateTime(machine.updated_at)} />
            </div>

            {machine.attachment && (
              <div>
                <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Attachment</span>
                <button
                  onClick={(e) => { e.stopPropagation(); handleDownloadAttachment(); }}
                  disabled={downloading}
                  className="mt-1 flex items-center gap-1.5 text-sm text-blue-600 hover:text-blue-800 transition-colors disabled:opacity-50"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  {downloading ? 'Downloading...' : machine.attachment}
                </button>
              </div>
            )}

            {machine.additional_notes && (
              <div>
                <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Notes</span>
                <p className="text-sm text-gray-700 mt-1 whitespace-pre-wrap">{machine.additional_notes}</p>
              </div>
            )}

            <div className="flex items-center gap-2 pt-2 border-t border-gray-100">
              <button onClick={(e) => { e.stopPropagation(); onView(machine.serial_number); }} className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-lg transition-colors">View</button>
              <button onClick={(e) => { e.stopPropagation(); onEdit(machine.serial_number); }} className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium text-yellow-700 bg-yellow-50 hover:bg-yellow-100 rounded-lg transition-colors">Edit</button>
              <button onClick={(e) => { e.stopPropagation(); onDelete(machine.serial_number); }} className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium text-red-700 bg-red-50 hover:bg-red-100 rounded-lg transition-colors">Delete</button>
            </div>
          </div>
        )}
      </div>
    );
  }

  // Desktop card — original layout with top-right quick action icons
  return (
    <>
      <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-200 overflow-hidden">
        <div className="p-6">
          <div className="flex justify-between items-start mb-4">
            <div className="flex-1">
              <h3 className="text-lg font-semibold text-gray-900 truncate" title={machine.serial_number}>
                {machine.serial_number}{' '}
                {machine.model && <span className="text-xs text-gray-500">({machine.model})</span>}
              </h3>
              <div className="text-xs text-gray-500 mt-1">
                {machine.brand ? `${machine.brand} · ` : ''}{machine.district ? `${machine.district}, ` : ''}{machine.state}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button
                type="button"
                className="flex items-center gap-1 bg-indigo-100 text-indigo-800 px-2 py-1 rounded-full text-xs font-medium hover:bg-indigo-200 transition-colors"
                title={`${maintenanceCount} maintenance record${maintenanceCount !== 1 ? 's' : ''}`}
                onClick={() => onView(machine.serial_number)}
              >
                <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
                </svg>
                <span>{maintenanceCount}</span>
              </button>
              {machine.attachment && (
                <button
                  type="button"
                  onClick={handleDownloadAttachment}
                  disabled={downloading}
                  className="p-1 text-gray-400 hover:text-blue-600 transition-colors disabled:opacity-50"
                  title={`Download ${machine.attachment}`}
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </button>
              )}
              {machine.additional_notes && (
                <button
                  type="button"
                  onClick={() => setShowNotesModal(true)}
                  className="p-1 text-gray-400 hover:text-green-600 transition-colors"
                  title="View additional notes"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </button>
              )}
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
              Account Type: <span className="font-normal">{machine.account_type || 'Not specified'}</span>
            </div>
            <div className="text-sm text-gray-700 font-medium">
              Person In Charge: <span className="font-normal">{machine.person_in_charge}</span>
            </div>
          </div>

          <div className="text-xs mb-4 space-y-1">
            <div className="text-gray-500">
              TNC Date: <span className="text-gray-700">{machine.tnc_date ? formatDate(machine.tnc_date) : '-'}</span>
            </div>
            <div className="text-gray-500">
              PPM Date: <span className="text-gray-700">{machine.ppm_date ? formatDate(machine.ppm_date) : '-'}</span>
            </div>
            <div className="text-gray-500">
              Reported By: <span className="text-gray-700">{machine.reported_by || 'Not specified'}</span>
            </div>
          </div>

          <div className="flex gap-2">
            <button onClick={() => onView(machine.serial_number)} className="flex-1 bg-blue-50 hover:bg-blue-100 text-blue-700 px-3 py-2 rounded-md text-sm font-medium transition-colors">View</button>
            <button onClick={() => onEdit(machine.serial_number)} className="flex-1 bg-yellow-50 hover:bg-yellow-100 text-yellow-700 px-3 py-2 rounded-md text-sm font-medium transition-colors">Edit</button>
            <button onClick={() => onDelete(machine.serial_number)} className="flex-1 bg-red-50 hover:bg-red-100 text-red-700 px-3 py-2 rounded-md text-sm font-medium transition-colors">Delete</button>
          </div>

          <div className="text-center text-xs mt-3 pt-3 border-t border-gray-100 flex items-center justify-center gap-3">
            <span className="flex items-center gap-1 text-gray-500">
              <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
              <span>Created:</span>
              <span className="text-gray-700">{formatDate(machine.created_at)}</span>
            </span>
            <span className="text-gray-400">&bull;</span>
            <span className="flex items-center gap-1 text-gray-500">
              <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              <span>Updated:</span>
              <span className="text-gray-700">{formatDate(machine.updated_at)}</span>
            </span>
          </div>
        </div>
      </div>

      {showNotesModal && (
        <div className="fixed inset-0 backdrop-blur-md flex items-center justify-center z-50" onClick={handleBackdropClick}>
          <div className="bg-white rounded-lg border-2 border-gray-800 p-6 max-w-md w-full mx-4 max-h-96 overflow-y-auto">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold text-gray-900">Additional Notes</h3>
              <button onClick={() => setShowNotesModal(false)} className="text-gray-400 hover:text-gray-600 transition-colors">
                <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div className="text-sm text-gray-700 mb-4">
              <strong>Machine:</strong>{' '}
              {machine.serial_number}
              {machine.model && ` (${machine.model})`}
            </div>
            <div className="text-sm text-gray-700 whitespace-pre-wrap">
              {machine.additional_notes || 'No additional notes available.'}
            </div>
            <div className="mt-6 flex justify-end">
              <button onClick={() => setShowNotesModal(false)} className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-4 py-2 rounded-md text-sm font-medium transition-colors">Close</button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
