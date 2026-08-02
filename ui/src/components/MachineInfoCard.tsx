'use client';

import { useEffect, useState } from 'react';
import { Machine } from '../types/machine';
import { AttachmentService } from '../services/attachmentService';
import {
  formatDateTime,
  formatMachineDateDisplay,
  machineDateUnsetClassName,
} from '../utils/formatters';
import { isMachineDateUnset } from '../utils/dateUtils';
import { getPPMStatusDisplay } from '../utils/ppmUtils';

interface MachineInfoCardProps {
  machine: Machine;
  onEdit?: () => void;
  onDelete?: () => void;
}

function formatMachineLocation(district: string, state: string): string {
  const parts = [district?.trim(), state?.trim()].filter(Boolean);
  return parts.length > 0 ? parts.join(', ') : '—';
}

export function machineInfoExpandedSessionKey(serialNumber: string): string {
  return `ralts_machine_info_card_expanded:${encodeURIComponent(serialNumber)}`;
}

function getSessionExpanded(serialNumber: string): boolean {
  try {
    return sessionStorage.getItem(machineInfoExpandedSessionKey(serialNumber)) === '1';
  } catch {
    return false;
  }
}

function setSessionExpanded(serialNumber: string, expanded: boolean): void {
  try {
    sessionStorage.setItem(machineInfoExpandedSessionKey(serialNumber), expanded ? '1' : '0');
  } catch {
    /* ignore */
  }
}

export default function MachineInfoCard({ machine, onEdit, onDelete }: MachineInfoCardProps) {
  const [expanded, setExpanded] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const ppmStatus = getPPMStatusDisplay(machine.ppm_status);
  const showPpmServerPill = ppmStatus && !isMachineDateUnset(machine.ppm_date);
  const tncDisplay = formatMachineDateDisplay(machine.tnc_date);
  const ppmDisplay = formatMachineDateDisplay(machine.ppm_date);
  const locationSummary = formatMachineLocation(machine.district, machine.state);

  useEffect(() => {
    setExpanded(getSessionExpanded(machine.serial_number));
  }, [machine.serial_number]);

  const toggleExpanded = () => {
    setExpanded((prev) => {
      const next = !prev;
      setSessionExpanded(machine.serial_number, next);
      return next;
    });
  };

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
    <div className="mb-6 overflow-hidden rounded-lg border bg-white shadow-sm">
      <div className="p-6 pb-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="min-w-0 flex-1">
            <h1 className="text-2xl font-bold leading-tight text-gray-900">Machine Information</h1>
            {!expanded && (
              <div className="mt-3 min-w-0">
                <p className="text-sm leading-snug text-gray-800">
                  <span className="font-medium text-gray-900">{machine.serial_number}</span>
                  {' · '}
                  <span className="break-words">{machine.customer || '—'}</span>
                  {' · '}
                  <span>{locationSummary}</span>
                </p>
                {(showPpmServerPill || machine.attachment) && (
                  <div className="mt-1.5 flex flex-wrap items-center gap-3">
                    {showPpmServerPill && (
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${ppmStatus.color}`}
                      >
                        {ppmStatus.label}
                      </span>
                    )}
                    {machine.attachment && (
                      <button
                        type="button"
                        onClick={handleDownloadAttachment}
                        disabled={downloading}
                        aria-label={`Download attachment ${machine.attachment}`}
                        className="inline-flex min-w-0 items-center gap-1.5 text-sm text-blue-600 hover:text-blue-800 underline-offset-2 hover:underline disabled:opacity-50"
                        title="Download attachment"
                      >
                        <svg className="h-4 w-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden>
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                          />
                        </svg>
                        <span className="truncate">{machine.attachment}</span>
                      </button>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
          {!expanded && (onEdit || onDelete) && (
            <div className="flex flex-shrink-0 flex-wrap items-center gap-2 sm:justify-end">
              {onEdit && (
                <button
                  type="button"
                  onClick={onEdit}
                  className="bg-yellow-600 hover:bg-yellow-700 text-white px-3 py-2 rounded-lg font-medium transition-colors flex items-center gap-2 text-sm"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                    />
                  </svg>
                  Edit Machine
                </button>
              )}
              {onDelete && (
                <button
                  type="button"
                  onClick={onDelete}
                  className="bg-red-600 hover:bg-red-700 text-white px-3 py-2 rounded-lg font-medium transition-colors flex items-center gap-2 text-sm"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                  Delete Machine
                </button>
              )}
            </div>
          )}
        </div>
      </div>

      <div id="machine-info-details" hidden={!expanded} className="px-6 pb-6 pt-0">
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
                <div className="font-medium text-gray-900">{locationSummary}</div>
              </div>
              <div>
                <span className="text-sm text-gray-500">Person in Charge:</span>
                <div className="font-medium text-gray-900">{machine.person_in_charge}</div>
              </div>
            </div>
            <div className="space-y-2">
              <div>
                <span className="text-sm text-gray-500">TNC Date:</span>
                <div
                  className={tncDisplay.isUnset ? machineDateUnsetClassName : 'font-medium text-gray-900'}
                >
                  {tncDisplay.text}
                </div>
              </div>
              <div>
                <span className="text-sm text-gray-500">PPM Date:</span>
                <div
                  className={
                    ppmDisplay.isUnset
                      ? 'flex items-center gap-2 flex-wrap'
                      : 'font-medium text-gray-900 flex items-center gap-2 flex-wrap'
                  }
                >
                  <span className={ppmDisplay.isUnset ? machineDateUnsetClassName : 'text-gray-900'}>
                    {ppmDisplay.text}
                  </span>
                  {showPpmServerPill && (
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
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                  />
                </svg>
                <span className="text-gray-500">Created:</span>
                <span className="font-medium text-gray-900">{formatDateTime(machine.created_at)}</span>
              </div>
              <div className="hidden sm:block text-gray-300">&bull;</div>
              <div className="flex items-center gap-2">
                <svg className="h-4 w-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
                <span className="text-gray-500">Last Updated:</span>
                <span className="font-medium text-gray-900">{formatDateTime(machine.updated_at)}</span>
              </div>
              {machine.updated_by && (
                <>
                  <div className="hidden sm:block text-gray-300">&bull;</div>
                  <div className="flex items-center gap-2">
                    <svg className="h-4 w-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                      />
                    </svg>
                    <span className="text-gray-500">Updated By:</span>
                    <span className="font-medium text-gray-900">{machine.updated_by}</span>
                  </div>
                </>
              )}
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
                      type="button"
                      onClick={handleDownloadAttachment}
                      disabled={downloading}
                      className="flex items-center gap-2 text-sm text-blue-600 hover:text-blue-800 transition-colors disabled:opacity-50"
                      title="Download machine attachment"
                    >
                      <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                        />
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
                  type="button"
                  onClick={onEdit}
                  className="bg-yellow-600 hover:bg-yellow-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2 text-sm"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                    />
                  </svg>
                  Edit Machine
                </button>
              )}
              {onDelete && (
                <button
                  type="button"
                  onClick={onDelete}
                  className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2 text-sm"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                  Delete Machine
                </button>
              )}
            </div>
          )}
      </div>

      <button
        type="button"
        id="machine-info-toggle"
        aria-expanded={expanded}
        aria-controls="machine-info-details"
        aria-label={expanded ? 'Hide machine details' : 'Show machine details'}
        onClick={toggleExpanded}
        title={expanded ? 'Hide details' : 'Show details'}
        className="flex w-full items-center justify-center gap-2 border-t border-gray-200 bg-gray-50/90 py-2 text-sm text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-blue-500"
      >
        <svg
          className={`h-4 w-4 shrink-0 transition-transform duration-200 ${expanded ? 'rotate-180' : ''}`}
          viewBox="0 0 20 20"
          fill="currentColor"
          aria-hidden
        >
          <path
            fillRule="evenodd"
            d="M5.22 8.22a.75.75 0 011.06 0L10 11.94l3.72-3.72a.75.75 0 111.06 1.06l-4.25 4.25a.75.75 0 01-1.06 0L5.22 9.28a.75.75 0 010-1.06z"
            clipRule="evenodd"
          />
        </svg>
        <span>{expanded ? 'Hide details' : 'Show details'}</span>
      </button>
    </div>
  );
}
