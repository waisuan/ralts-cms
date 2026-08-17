'use client';

import React, { useEffect, useState } from 'react';
import { Flag, FlagService, MAX_FLAG_NOTE_LENGTH } from '../services/flagService';
import { formatDateTime } from '../utils/formatters';
import { useLockBodyScroll } from '../hooks/useLockBodyScroll';

interface FlagMachineModalProps {
  isOpen: boolean;
  /** Serial number of the machine being flagged. */
  serialNumber: string;
  onClose: () => void;
  /** Called after the flag is persisted, so callers can refresh badges. */
  onCreated?: (flag: Flag) => void;
}

type ManualReason = 'missing_values' | 'other';

/**
 * FlagMachineModal raises a manual flag on a machine. Admin-only on the API;
 * callers are expected to only offer the action to admins. The machine's
 * assignee is notified when the flag is created.
 *
 * A machine carries one open flag at a time. If it already has one, this form
 * opens on that flag's reason and note and saving replaces it, rather than
 * adding a second flag to the same machine.
 */
export default function FlagMachineModal({
  isOpen,
  serialNumber,
  onClose,
  onCreated,
}: FlagMachineModalProps) {
  useLockBodyScroll(isOpen);

  const [reason, setReason] = useState<ManualReason>('missing_values');
  const [note, setNote] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [existing, setExisting] = useState<Flag | null>(null);
  const [isLoadingExisting, setIsLoadingExisting] = useState(false);

  useEffect(() => {
    if (!isOpen) return;

    setReason('missing_values');
    setNote('');
    setIsSubmitting(false);
    setError(null);
    setExisting(null);
    setIsLoadingExisting(true);

    let cancelled = false;
    FlagService.listForMachine(serialNumber)
      .then((res) => {
        if (cancelled) return;
        // At most one open flag exists, so the first is the one being replaced.
        const open = res.data?.flags?.[0];
        if (!open) return;
        setExisting(open);
        if (open.reason === 'missing_values' || open.reason === 'other') {
          setReason(open.reason);
        }
        setNote(open.note ?? '');
      })
      .catch((err) => {
        // Only the prefill is lost: submitting still replaces the open flag.
        console.debug('Could not load the existing flag for', serialNumber, err);
      })
      .finally(() => {
        if (!cancelled) setIsLoadingExisting(false);
      });

    return () => {
      cancelled = true;
    };
  }, [isOpen, serialNumber]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (reason === 'other' && !note.trim()) {
      setError('A note is required when the reason is "Other".');
      return;
    }

    setIsSubmitting(true);
    setError(null);
    try {
      const res = await FlagService.create(serialNumber, { reason, note: note.trim() });
      if (res.data) onCreated?.(res.data);
      onClose();
    } catch (err) {
      console.error('Failed to create flag:', err);
      setError(err instanceof Error ? err.message : 'Failed to create flag');
    } finally {
      setIsSubmitting(false);
    }
  };

  const remaining = MAX_FLAG_NOTE_LENGTH - note.length;

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 overflow-y-auto bg-black bg-opacity-50"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-labelledby="flag-machine-title"
    >
      <div className="flex min-h-full items-center justify-center p-4">
        <div
          className="relative w-full max-w-md bg-white rounded-lg shadow-xl"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
            <div>
              <h2 id="flag-machine-title" className="text-xl font-semibold text-gray-900">
                {existing ? 'Update Flag' : 'Flag Machine'}
              </h2>
              <p className="mt-0.5 text-sm text-gray-500">{serialNumber}</p>
            </div>
            <button
              type="button"
              onClick={onClose}
              aria-label="Close"
              className="text-gray-400 hover:text-gray-600 focus:outline-none focus:text-gray-600 transition-colors"
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

          <form onSubmit={handleSubmit} className="px-6 py-6">
            {existing && (
              <div className="mb-4 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
                This machine was already flagged {formatDateTime(existing.created_at)}
                {existing.created_by_username ? ` by ${existing.created_by_username}` : ''}. Saving
                replaces that flag rather than adding another, and resets when it was flagged.
              </div>
            )}

            <div className="mb-4">
              <label htmlFor="flag-reason" className="block text-sm font-medium text-gray-700 mb-2">
                Reason <span className="text-red-500">*</span>
              </label>
              <select
                id="flag-reason"
                value={reason}
                onChange={(e) => {
                  setReason(e.target.value as ManualReason);
                  setError(null);
                }}
                disabled={isSubmitting}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-white text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="missing_values">Missing values</option>
                <option value="other">Other</option>
              </select>
            </div>

            <div className="mb-4">
              <label htmlFor="flag-note" className="block text-sm font-medium text-gray-700 mb-2">
                Note {reason === 'other' && <span className="text-red-500">*</span>}
              </label>
              <textarea
                id="flag-note"
                value={note}
                onChange={(e) => {
                  setNote(e.target.value);
                  setError(null);
                }}
                rows={3}
                maxLength={MAX_FLAG_NOTE_LENGTH}
                disabled={isSubmitting}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg placeholder-gray-400 text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder={
                  reason === 'missing_values'
                    ? 'Optional: which fields are missing?'
                    : 'Explain why this machine is being flagged'
                }
              />
              <div className="mt-1 flex items-start justify-between gap-3 text-xs text-gray-500">
                <p>The assignee of this machine is notified with this note.</p>
                <p
                  aria-live="polite"
                  className={`shrink-0 tabular-nums ${remaining <= 50 ? 'text-amber-700' : ''}`}
                >
                  {note.length}/{MAX_FLAG_NOTE_LENGTH}
                </p>
              </div>
            </div>

            {error && (
              <div className="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md text-sm">
                {error}
              </div>
            )}

            <div className="flex justify-end gap-3">
              <button
                type="button"
                onClick={onClose}
                disabled={isSubmitting}
                className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                type="submit"
                // Held back until the lookup lands, so a save cannot overwrite an
                // existing note the admin has not seen yet.
                disabled={isSubmitting || isLoadingExisting}
                className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isSubmitting
                  ? existing
                    ? 'Updating...'
                    : 'Flagging...'
                  : existing
                    ? 'Update Flag'
                    : 'Flag Machine'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
