'use client';

import { useEffect, useState } from 'react';
import { useOptionalAuth } from '../contexts/AuthContext';
import { Flag, FlagService } from '../services/flagService';

/**
 * Batch-fetches open flags for the given list of machine serial numbers.
 * Returns a map keyed by serial number — empty/missing entries mean no open
 * flags for that machine. Refetches whenever the joined key of serial numbers
 * changes so pagination and search stay in sync.
 *
 * Every signed-in user sees badges, regardless of who a machine is assigned to:
 * knowing a machine is flagged is useful to anyone looking at it, even though
 * only admins can raise a flag. Signed-out visitors get an empty map and no
 * request, since the endpoint needs a token. Gating here rather than in each
 * consumer keeps the rule in one place.
 *
 * Bumping `reloadToken` forces a refetch, which callers use after raising a new
 * flag so the badge shows up straight away.
 */
export function useOpenFlagsForMachines(
  serialNumbers: string[],
  reloadToken: number = 0
): {
  flagsBySerial: Record<string, Flag[]>;
  isLoading: boolean;
} {
  const auth = useOptionalAuth();
  const isSignedIn = !!auth?.user;
  const [flagsBySerial, setFlagsBySerial] = useState<Record<string, Flag[]>>({});
  const [isLoading, setIsLoading] = useState(false);
  const key = serialNumbers.slice().sort().join('|');

  useEffect(() => {
    if (!isSignedIn || serialNumbers.length === 0) {
      setFlagsBySerial({});
      return;
    }
    let cancelled = false;
    setIsLoading(true);
    FlagService.listOpenByMachine(serialNumbers)
      .then((res) => {
        if (cancelled) return;
        setFlagsBySerial(res.data?.flags ?? {});
      })
      .catch((err) => {
        if (cancelled) return;
        console.debug('Failed to load open flags:', err);
        setFlagsBySerial({});
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
    // We intentionally depend on the joined key (a stable string) rather than
    // the array reference so parents don't need to memoize the array.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key, isSignedIn, reloadToken]);

  return { flagsBySerial, isLoading };
}
