'use client';

import { Flag } from '../services/flagService';
import { formatDateTime } from '../utils/formatters';

interface FlagBadgeProps {
  flags: Flag[];
  /** When compact, only shows a red icon; otherwise renders count and reason. */
  compact?: boolean;
}

/**
 * FlagBadge renders a small pill indicating a machine has open flags. When
 * multiple flags are open, hover reveals the individual reasons.
 */
export default function FlagBadge({ flags, compact = false }: FlagBadgeProps) {
  if (!flags || flags.length === 0) return null;

  const title = flags
    .map((f) => {
      const raised = `${f.reason} · flagged ${formatDateTime(f.created_at)}`;
      return f.note ? `${raised}\n${f.note}` : raised;
    })
    .join('\n');

  if (compact) {
    return (
      <span
        title={title}
        className="inline-flex items-center gap-0.5 rounded bg-orange-100 text-orange-700 border border-orange-200 px-1.5 py-0.5 text-xs font-medium"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-3 w-3"
          fill="currentColor"
          viewBox="0 0 20 20"
          aria-hidden="true"
        >
          <path d="M3 3a1 1 0 011-1h11.586A1 1 0 0116.293 3.707L14 6l2.293 2.293A1 1 0 0115.586 10H5v7a1 1 0 11-2 0V3z" />
        </svg>
        {flags.length}
      </span>
    );
  }

  return (
    <span
      title={title}
      className="inline-flex items-center gap-1 rounded bg-orange-100 text-orange-800 border border-orange-200 px-2 py-0.5 text-xs font-medium"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        className="h-3.5 w-3.5"
        fill="currentColor"
        viewBox="0 0 20 20"
        aria-hidden="true"
      >
        <path d="M3 3a1 1 0 011-1h11.586A1 1 0 0116.293 3.707L14 6l2.293 2.293A1 1 0 0115.586 10H5v7a1 1 0 11-2 0V3z" />
      </svg>
      {flags.length} {flags.length === 1 ? 'flag' : 'flags'}
    </span>
  );
}
