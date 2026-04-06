'use client';

import { useState } from 'react';
import { Machine } from '../types/machine';
import { PPM_STATUSES, PPM_STATUS_COLORS } from '../utils/constants';
import { AttachmentService } from '../services/attachmentService';
import { formatDate, formatDateTime } from '../utils/formatters';

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

export default function RecordCard({ machine, onView, onEdit, onDelete }: RecordCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const status = getPPMStatusDisplay(machine.ppm_status);

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

  return (
    <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
      {/* Collapsed header — always visible */}
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
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </div>

      {/* Expanded details */}
      {isExpanded && (
        <div className="px-4 pb-4 border-t border-gray-100 pt-3 space-y-3">
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">TNC Date</span>
              <p className="text-gray-900 mt-0.5">{machine.tnc_date ? formatDate(machine.tnc_date) : '-'}</p>
            </div>
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">PPM Date</span>
              <p className="text-gray-900 mt-0.5">{machine.ppm_date ? formatDate(machine.ppm_date) : '-'}</p>
            </div>
            {machine.model && (
              <div>
                <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Model</span>
                <p className="text-gray-900 mt-0.5">{machine.model}</p>
              </div>
            )}
            {machine.brand && (
              <div>
                <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Brand</span>
                <p className="text-gray-900 mt-0.5">{machine.brand}</p>
              </div>
            )}
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">District</span>
              <p className="text-gray-900 mt-0.5">{machine.district || '-'}</p>
            </div>
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Status</span>
              <p className="text-gray-900 mt-0.5">{machine.status || '-'}</p>
            </div>
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Account Type</span>
              <p className="text-gray-900 mt-0.5">{machine.account_type || '-'}</p>
            </div>
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Person In Charge</span>
              <p className="text-gray-900 mt-0.5">{machine.person_in_charge || '-'}</p>
            </div>
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Reported By</span>
              <p className="text-gray-900 mt-0.5">{machine.reported_by || '-'}</p>
            </div>
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Created</span>
              <p className="text-gray-900 mt-0.5">{formatDateTime(machine.created_at)}</p>
            </div>
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Updated</span>
              <p className="text-gray-900 mt-0.5">{formatDateTime(machine.updated_at)}</p>
            </div>
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

          {/* Action buttons */}
          <div className="flex items-center gap-2 pt-2 border-t border-gray-100">
            <button
              onClick={(e) => { e.stopPropagation(); onView(machine.serial_number); }}
              className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-lg transition-colors"
            >
              View
            </button>
            <button
              onClick={(e) => { e.stopPropagation(); onEdit(machine.serial_number); }}
              className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium text-yellow-700 bg-yellow-50 hover:bg-yellow-100 rounded-lg transition-colors"
            >
              Edit
            </button>
            <button
              onClick={(e) => { e.stopPropagation(); onDelete(machine.serial_number); }}
              className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium text-red-700 bg-red-50 hover:bg-red-100 rounded-lg transition-colors"
            >
              Delete
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
