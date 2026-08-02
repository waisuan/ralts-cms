'use client';

import { useState } from 'react';
import { Maintenance } from '../types/maintenance';
import { AttachmentService } from '../services/attachmentService';
import {
  formatDate,
  formatDateTime,
  getTypeColor,
  isCustomWorkOrderType,
  workOrderTypePillLabel,
} from '../utils/formatters';

interface MaintenanceCardProps {
  record: Maintenance;
  machineSerialNumber: string;
  onEdit: (record: Maintenance) => void;
  onDelete: (record: Maintenance) => void;
}

export default function MaintenanceCard({
  record,
  machineSerialNumber,
  onEdit,
  onDelete,
}: MaintenanceCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [downloading, setDownloading] = useState(false);

  const handleDownloadAttachment = async () => {
    if (!record.attachment || downloading) return;
    setDownloading(true);
    try {
      const blob = await AttachmentService.downloadMaintenanceAttachment(
        machineSerialNumber,
        record.work_order_number,
        record.attachment
      );
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = record.attachment;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to download attachment:', error);
      alert(`Failed to download attachment: ${record.attachment}`);
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
      {/* Clickable header area */}
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
              <span className="font-semibold text-gray-900 text-sm truncate">{record.work_order_number}</span>
              <span
                className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${getTypeColor(workOrderTypePillLabel(record))}`}
              >
                {workOrderTypePillLabel(record)}
              </span>
            </div>
            <p className="text-xs text-gray-500 mt-1">{formatDate(record.work_order_date)}</p>
            {!isExpanded && (
              <p className="text-sm text-gray-700 mt-2 line-clamp-2">{record.action_taken}</p>
            )}
          </div>
          <svg
            className={`h-5 w-5 text-gray-400 flex-shrink-0 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </div>
        {!isExpanded && (
          <p className="text-xs text-gray-500 mt-1">Reported by: {record.reported_by}</p>
        )}
      </div>

      {/* Expanded details */}
      {isExpanded && (
        <div className="px-4 pb-4 border-t border-gray-100 pt-3 space-y-3">
          {isCustomWorkOrderType(record) && (
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Maintenance type</span>
              <p className="text-sm text-gray-900 mt-1">{record.work_order_type}</p>
            </div>
          )}
          <div>
            <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Action Taken</span>
            <p className="text-sm text-gray-700 mt-1 whitespace-pre-wrap">{record.action_taken}</p>
          </div>
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Reported By</span>
              <p className="text-gray-900 mt-0.5">{record.reported_by}</p>
            </div>
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Work Order Date</span>
              <p className="text-gray-900 mt-0.5">{formatDate(record.work_order_date)}</p>
            </div>
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Created</span>
              <p className="text-gray-900 mt-0.5">{formatDateTime(record.created_at)}</p>
            </div>
            <div>
              <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Updated</span>
              <p className="text-gray-900 mt-0.5">{formatDateTime(record.updated_at)}</p>
            </div>
            {record.updated_by && (
              <div>
                <span className="text-xs font-medium text-gray-500 uppercase tracking-wide">Updated By</span>
                <p className="text-gray-900 mt-0.5">{record.updated_by}</p>
              </div>
            )}
          </div>

          {record.attachment && (
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
                {downloading ? 'Downloading...' : record.attachment}
              </button>
            </div>
          )}

          {/* Action buttons */}
          <div className="flex items-center gap-2 pt-2 border-t border-gray-100">
            <button
              onClick={(e) => { e.stopPropagation(); onEdit(record); }}
              className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium text-yellow-700 bg-yellow-50 hover:bg-yellow-100 rounded-lg transition-colors"
              aria-label={`Edit record ${record.work_order_number}`}
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              Edit
            </button>
            <button
              onClick={(e) => { e.stopPropagation(); onDelete(record); }}
              className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium text-red-700 bg-red-50 hover:bg-red-100 rounded-lg transition-colors"
              aria-label={`Delete record ${record.work_order_number}`}
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              Delete
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
