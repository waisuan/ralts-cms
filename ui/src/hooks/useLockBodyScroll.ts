import { useEffect } from 'react';

// Module-scoped counter so stacked modals (e.g. a delete confirm opened on top
// of an edit modal) restore scroll only after the last one unmounts.
let lockCount = 0;
let previousOverflow = '';
let previousPaddingRight = '';

/**
 * Locks `document.body` scroll while `locked` is true.
 *
 * When the body has a visible scrollbar we compensate with right padding so
 * the page behind the overlay doesn't shift when the scrollbar disappears.
 * Safe to call from multiple components concurrently.
 */
export function useLockBodyScroll(locked: boolean): void {
  useEffect(() => {
    if (!locked || typeof document === 'undefined') return;

    if (lockCount === 0) {
      previousOverflow = document.body.style.overflow;
      previousPaddingRight = document.body.style.paddingRight;
      const scrollbarWidth =
        window.innerWidth - document.documentElement.clientWidth;
      document.body.style.overflow = 'hidden';
      if (scrollbarWidth > 0) {
        document.body.style.paddingRight = `${scrollbarWidth}px`;
      }
    }
    lockCount += 1;

    return () => {
      lockCount = Math.max(0, lockCount - 1);
      if (lockCount === 0) {
        document.body.style.overflow = previousOverflow;
        document.body.style.paddingRight = previousPaddingRight;
      }
    };
  }, [locked]);
}
