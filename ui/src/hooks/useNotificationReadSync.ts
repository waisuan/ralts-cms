'use client';

import { useCallback, useEffect, useRef } from 'react';

/** Dispatched on window whenever notifications are marked read. */
export const NOTIFICATIONS_READ_EVENT = 'ralts:notifications-read';

/**
 * Keeps the two notification views that can be on screen together — the header
 * bell and the inbox page — agreeing about what has been read. Each holds its
 * own copy of that state, so without this a count cleared in one would sit stale
 * in the other until the next navigation.
 *
 * Pass what to do when read state changes elsewhere; call the returned function
 * after marking anything read. A view never hears its own announcement, having
 * already applied that change itself.
 */
export function useNotificationReadSync(onReadElsewhere: () => void): () => void {
  const handler = useRef(onReadElsewhere);
  // dispatchEvent is synchronous, so a flag held across the call is enough to
  // recognise our own announcement while it is being delivered.
  const announcing = useRef(false);

  useEffect(() => {
    handler.current = onReadElsewhere;
  }, [onReadElsewhere]);

  // Subscribing once, and reaching the callback through a ref, means callers do
  // not have to memoize a handler that closes over their current state.
  useEffect(() => {
    const listener = () => {
      if (!announcing.current) handler.current();
    };
    window.addEventListener(NOTIFICATIONS_READ_EVENT, listener);
    return () => window.removeEventListener(NOTIFICATIONS_READ_EVENT, listener);
  }, []);

  return useCallback(() => {
    announcing.current = true;
    try {
      window.dispatchEvent(new Event(NOTIFICATIONS_READ_EVENT));
    } finally {
      announcing.current = false;
    }
  }, []);
}
