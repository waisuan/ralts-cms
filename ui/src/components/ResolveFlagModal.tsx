'use client';

import React, { useState } from 'react';
import { Flag, FlagService, MAX_FLAG_NOTE_LENGTH } from '../services/flagService';
import { flaggedAtLabel } from '../utils/flagTimestamps';
import { useLockBodyScroll } from '../hooks/useLockBodyScroll';

interface ResolveFlagModalProps {
  /** The flag being resolved. Render the modal only when there is one. */
  flag: Flag;
  onClose: () => void;
  /** Called after the flag is resolved, so callers can refresh their list. */
  onResolved?: () => void;
}

/**
 * ResolveFlagModal closes a flag, with an optional note saying what was done.
 * The note is kept on the flag as part of its audit trail; it is not sent as a
 * notification.
 *
 * Callers mount this only while a flag is selected, so the form starts empty
 * every time without a reset effect.
 */
export default function ResolveFlagModal({ flag, onClose, onResolved }: ResolveFlagModalProps) {
  useLockBodyScroll(true);

  const [note, setNote] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    setIsSubmitting(true);
    setError(null);
    try {
      await FlagService.resolve(flag.machine_serial_number, flag.id, note);
      onResolved?.();
      onClose();
    } catch (err) {
      console.error('Failed to resolve flag:', err);
      setError(err instanceof Error ? err.message : 'Failed to resolve flag');
      setIsSubmitting(false);
    }
  };

  const remaining = MAX_FLAG_NOTE_LENGTH - note.length;

  return (
    <div
      className="fixed inset-0 z-50 overflow-y-auto bg-black bg-opacity-50"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-labelledby="resolve-flag-title"
    >
      <div className="flex min-h-full items-center justify-center p-4">
        <div
          className="relative w-full max-w-md bg-white rounded-lg shadow-xl"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
            <div>
              <h2 id="resolve-flag-title" className="text-xl font-semibold text-gray-900">
                Resolve Flag
              </h2>
              <p className="mt-0.5 text-sm text-gray-500">
                {flag.machine_serial_number} · {FlagService.reasonLabel(flag.reason)}
              </p>
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
            <div className="mb-4 rounded-md border border-gray-200 bg-gray-50 px-3 py-2">
              <p className="text-xs font-medium text-gray-500">{flaggedAtLabel(flag)}</p>
              {flag.note && (
                <p className="mt-1 text-sm text-gray-800 whitespace-pre-line">{flag.note}</p>
              )}
            </div>

            <div className="mb-4">
              <label
                htmlFor="resolution-note"
                className="block text-sm font-medium text-gray-700 mb-2"
              >
                Resolution note <span className="text-gray-400">(optional)</span>
              </label>
              <textarea
                id="resolution-note"
                value={note}
                onChange={(e) => {
                  setNote(e.target.value);
                  setError(null);
                }}
                rows={3}
                maxLength={MAX_FLAG_NOTE_LENGTH}
                disabled={isSubmitting}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg placeholder-gray-400 text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="What did you do about it?"
              />
              <div className="mt-1 flex items-start justify-between gap-3 text-xs text-gray-500">
                <p>Kept on the record so others can see how this was closed.</p>
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
                disabled={isSubmitting}
                className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isSubmitting ? 'Resolving...' : 'Resolve Flag'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
