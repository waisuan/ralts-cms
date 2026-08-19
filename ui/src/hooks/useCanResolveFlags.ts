'use client';

import { useCallback } from 'react';
import { useOptionalAuth } from '../contexts/AuthContext';
import { USER_ROLE } from '../services/adminUserService';
import { Machine } from '../types/machine';

/**
 * Returns a predicate for whether the signed-in user may resolve flags on a
 * machine, mirroring what the API allows: an admin anywhere, anyone else only on
 * machines currently assigned to them. A free-text assignee has no user id, so
 * only an admin can clear flags on those machines.
 *
 * Keeping the rule here means the records table, the cards and anything added
 * later agree with each other, and with the flagged records page, about which
 * flags come with a resolve action.
 */
export function useCanResolveFlags(): (machine: Machine) => boolean {
  const user = useOptionalAuth()?.user ?? null;

  return useCallback(
    (machine: Machine) => {
      if (!user) return false;
      if (user.role === USER_ROLE.ADMIN) return true;
      return machine.assigned_user_id != null && machine.assigned_user_id === user.id;
    },
    [user]
  );
}
