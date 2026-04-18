'use client';

import { useState } from 'react';
import { Machine } from '../types/machine';
import { AttachmentService } from '../services/attachmentService';
import { formatDate, formatDateTime } from '../utils/formatters';
import { getPPMStatusDisplay } from '../utils/ppmUtils';

interface MachineInfoCardProps {
  machine: Machine;
  onEdit?: () => void;
  onDelete?: () => void;
}

export default function MachineInfoCard({ machine, onEdit, onDelete }: MachineInfoCardProps) {
  const [downloading, setDownloading] = useState(false);
  const ppmStatus = getPPMStatusDisplay(machine.ppm_status);

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
      console.error('Failed to download machine attachment:', error);
      alert(`Failed to download attachment: ${machine.attachment}`);
    } finally {
      setDownloading(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border p-6 mb-6">
      <h1 className="text-2xl font-bold text-gray-900 mb-4">Machine Information</h1>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div className="space-y-2">
          <div>
            <span className="text-sm text-gray-500">Serial Number:</span>
            <div className="font-medium text-gray-900">{machine.serial_number}</div>
          </div>
          <div>
            <span className="text-sm text-gray-500">Model:</span>
            <div className="font-medium text-gray-900">{machine.model}</div>
          </div>
          <div>
            <span className="text-sm text-gray-500">Brand:</span>
            <div className="font-medium text-gray-900">{machine.brand}</div>
          </div>
          <div>
            <span className="text-sm text-gray-500">Status:</span>
            <div className="font-medium text-gray-900">{machine.status || 'Not specified'}</div>
          </div>
        </div>
        <div className="space-y-2">
          <div>
            <span className="text-sm text-gray-500">Customer:</span>
            <div className="font-medium text-gray-900">{machine.customer}</div>
          </div>
          <div>
            <span className="text-sm text-gray-500">Account Type:</span>
            <div className="font-medium text-gray-900">{machine.account_type || 'Not specified'}</div>
          </div>
          <div>
            <span className="text-sm text-gray-500">Location:</span>
            <div className="font-medium text-gray-900">{machine.district}, {machine.state}</div>
          </div>
          <div>
            <span className="text-sm text-gray-500">Person in Charge:</span>
            <div className="font-medium text-gray-900">{machine.person_in_charge}</div>
          </div>
        </div>
        <div className="space-y-2">
          <div>
            <span className="text-sm text-gray-500">TNC Date:</span>
            <div className="font-medium text-gray-900">{machine.tnc_date ? formatDate(machine.tnc_date) : 'Not set'}</div>
          </div>
          <div>
            <span className="text-sm text-gray-500">PPM Date:</span>
            <div className="font-medium text-gray-900 flex items-center gap-2 flex-wrap">
              <span>{machine.ppm_date ? formatDate(machine.ppm_date) : 'Not set'}</span>
              {ppmStatus && (
                <span
                  className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${ppmStatus.color}`}
                >
                  {ppmStatus.label}
                </span>
              )}
            </div>
          </div>
          <div>
            <span className="text-sm text-gray-500">Reported By:</span>
            <div className="font-medium text-gray-900">{machine.reported_by || 'Not specified'}</div>
          </div>
        </div>
      </div>

      {/* Timestamps */}
      <div className="mt-6 pt-4 border-t border-gray-200">
        <div className="flex flex-wrap items-center justify-center gap-8 text-sm">
          <div className="flex items-center gap-2">
            <svg className="h-4 w-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            <span className="text-gray-500">Created:</span>
            <span className="font-medium text-gray-900">{formatDateTime(machine.created_at)}</span>
          </div>
          <div className="hidden sm:block text-gray-300">&bull;</div>
          <div className="flex items-center gap-2">
            <svg className="h-4 w-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
            <span className="text-gray-500">Last Updated:</span>
            <span className="font-medium text-gray-900">{formatDateTime(machine.updated_at)}</span>
          </div>
        </div>
      </div>

      {/* Notes & Attachment */}
      {(machine.additional_notes || machine.attachment) && (
        <div className="mt-6 pt-6 border-t border-gray-200">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {machine.additional_notes && (
              <div>
                <h4 className="text-sm font-semibold text-gray-700 mb-2">Additional Notes</h4>
                <div className="text-sm text-gray-700 bg-gray-100 rounded-lg p-3">{machine.additional_notes}</div>
              </div>
            )}
            {machine.attachment && (
              <div>
                <h4 className="text-sm font-semibold text-gray-700 mb-2">Attachment</h4>
                <button
                  onClick={handleDownloadAttachment}
                  disabled={downloading}
                  className="flex items-center gap-2 text-sm text-blue-600 hover:text-blue-800 transition-colors disabled:opacity-50"
                  title="Download machine attachment"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  <span>{machine.attachment}</span>
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Machine Actions */}
      {(onEdit || onDelete) && (
        <div className="mt-6 pt-4 border-t border-gray-200 flex items-center gap-3">
          {onEdit && (
            <button
              onClick={onEdit}
              className="bg-yellow-600 hover:bg-yellow-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2 text-sm"
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              Edit Machine
            </button>
          )}
          {onDelete && (
            <button
              onClick={onDelete}
              className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2 text-sm"
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              Delete Machine
            </button>
          )}
        </div>
      )}
    </div>
  );
}
