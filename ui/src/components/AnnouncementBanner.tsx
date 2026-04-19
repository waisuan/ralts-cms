'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { CHANGELOG_PATH } from '@/utils/constants';

const SESSION_KEY = 'ralts_announcement_dismissed_until';

function announcementUntilRaw(): string | null {
  const raw = process.env.NEXT_PUBLIC_ANNOUNCEMENT_UNTIL?.trim();
  return raw || null;
}

function isAnnouncementWindowOpen(raw: string): boolean {
  const untilMs = new Date(raw).getTime();
  return !Number.isNaN(untilMs) && Date.now() < untilMs;
}

/** Env `NEXT_PUBLIC_ANNOUNCEMENT_UNTIL` (ISO UTC); dismiss stored in sessionStorage per campaign string. */
export default function AnnouncementBanner() {
  const [show, setShow] = useState(false);

  useEffect(() => {
    const raw = announcementUntilRaw();
    if (!raw || !isAnnouncementWindowOpen(raw)) return;

    try {
      if (sessionStorage.getItem(SESSION_KEY) === raw) return;
    } catch {
      /* storage unavailable */
    }
    // eslint-disable-next-line react-hooks/set-state-in-effect -- client-only gate after mount
    setShow(true);
  }, []);

  const dismiss = () => {
    const raw = announcementUntilRaw();
    if (!raw) return;
    try {
      sessionStorage.setItem(SESSION_KEY, raw);
    } catch {
      /* ignore */
    }
    setShow(false);
  };

  if (!show) return null;

  return (
    <div
      role="region"
      aria-label="Announcement"
      className="border-b border-blue-200 bg-blue-50 py-2 text-sm text-blue-950 sm:py-2.5"
    >
      <div className="container mx-auto flex items-center gap-2 px-4">
        {/* Same width as dismiss column so the message stays optically centered */}
        <div className="w-10 shrink-0 sm:w-11" aria-hidden />
        <p className="min-w-0 flex-1 text-center leading-snug">
          <span aria-hidden className="mr-1.5">
            📢
          </span>
          <span className="font-medium">You&apos;re on the new Ralts CMS.</span>{' '}
          <span className="text-blue-950/90">
            See what changed from the previous app —{' '}
            <Link
              href={CHANGELOG_PATH}
              className="font-medium text-blue-800 underline underline-offset-2 hover:text-blue-950"
            >
              What&apos;s new?
            </Link>
          </span>
        </p>
        <div className="flex w-10 shrink-0 items-center justify-end sm:w-11">
          <button
            type="button"
            onClick={dismiss}
            aria-label="Dismiss announcement"
            className="rounded-md p-2 text-blue-800 transition-colors hover:bg-blue-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden>
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  );
}
