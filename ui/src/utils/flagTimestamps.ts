import { Flag } from '@/services/flagService';
import { formatDateTime } from './formatters';

/**
 * When the flag was raised and by whom, e.g. "Flagged Jun 1, 2024, 09:00 AM by
 * admin". Re-flagging a machine rewrites its open flag, so this is the time of
 * the latest request rather than of the first one.
 */
export function flaggedAtLabel(flag: Flag): string {
  return `Flagged ${formatDateTime(flag.created_at)} by ${flag.created_by_username ?? 'an unknown user'}`;
}

/**
 * The same for the resolution, or null while the flag is still open. The time is
 * the flag's own `resolved_at`, not its raise time, which is why a resolved row
 * carries both lines.
 */
export function resolvedAtLabel(flag: Flag): string | null {
  if (flag.status !== 'resolved') return null;

  const who = flag.resolved_by_username ?? 'an unknown user';
  return flag.resolved_at
    ? `Resolved ${formatDateTime(flag.resolved_at)} by ${who}`
    : `Resolved by ${who}`;
}
