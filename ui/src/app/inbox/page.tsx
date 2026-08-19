'use client';

import Link from 'next/link';
import { useEffect, useState, useCallback, useRef } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useRouter } from 'next/navigation';
import {
  NotificationService,
  Notification,
  NotificationType,
} from '@/services/notificationService';
import { useNotificationReadSync } from '@/hooks/useNotificationReadSync';
import { machineDetailHref } from '@/utils/machineRoutes';
import LoadingSpinner from '@/components/LoadingSpinner';

const PAGE_SIZE = 25;

/** Labels for the type chip, since raw values like flag_resolved read badly. */
const TYPE_LABELS: Record<NotificationType, string> = {
  assigned: 'Assigned',
  flagged: 'Flagged',
  flag_resolved: 'Flag resolved',
};

function formatWhen(iso: string): string {
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

/**
 * Inbox page — a read-only list of notifications with pagination and a filter.
 * Flags are resolved on the flagged records page, which each notification about
 * an open flag links to.
 *
 * Marking anything read here also updates the header bell, and vice versa, since
 * both are on screen at once.
 */
export default function InboxPage() {
  const { user, isLoading: authLoading } = useAuth();
  const router = useRouter();

  const [items, setItems] = useState<Notification[]>([]);
  const [count, setCount] = useState(0);
  const [unreadCount, setUnreadCount] = useState(0);
  const [page, setPage] = useState(1);
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  // Identifies the most recent list request, so a slower one it overtook cannot
  // put its stale page back on screen.
  const latestRequest = useRef(0);

  useEffect(() => {
    if (!authLoading && !user) {
      router.replace('/');
    }
  }, [authLoading, user, router]);

  const load = useCallback(async (nextPage: number, filterUnreadOnly: boolean, quiet = false) => {
    // A quiet load is a refresh of something the user did elsewhere on the page,
    // so it leaves the rows they are reading alone: no spinner in their place,
    // and no error banner if the refresh fails, since they asked for nothing.
    if (!quiet) {
      setIsLoading(true);
      setError(null);
    }
    const request = ++latestRequest.current;
    try {
      const offset = (nextPage - 1) * PAGE_SIZE;
      const res = await NotificationService.list({
        limit: PAGE_SIZE,
        offset,
        unread_only: filterUnreadOnly,
      });
      // A newer load has since been asked for; its answer is the current one.
      if (request !== latestRequest.current) return;
      setItems(res.data?.notifications ?? []);
      setCount(res.data?.count ?? 0);
      setUnreadCount(res.data?.unread_count ?? 0);
    } catch (err) {
      console.error('Failed to load notifications:', err);
      if (!quiet) setError('Failed to load notifications');
    } finally {
      if (!quiet) setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!user) return;
    load(page, unreadOnly);
  }, [user, page, unreadOnly, load]);

  // The header bell can mark notifications read while this page is open, so
  // re-read the list when it does rather than leaving these rows looking unread.
  const announceRead = useNotificationReadSync(() => {
    void load(page, unreadOnly, true);
  });

  const totalPages = Math.max(1, Math.ceil(count / PAGE_SIZE));

  const handleMarkRead = async (id: string) => {
    try {
      await NotificationService.markRead(id);
      setItems((prev) =>
        prev.map((n) => (n.id === id ? { ...n, read_at: new Date().toISOString() } : n))
      );
      setUnreadCount((u) => Math.max(0, u - 1));
      announceRead();
    } catch (err) {
      console.error('markRead failed', err);
    }
  };

  const handleMarkAllRead = async () => {
    try {
      await NotificationService.markAllRead();
      const now = new Date().toISOString();
      setItems((prev) => prev.map((n) => ({ ...n, read_at: n.read_at ?? now })));
      setUnreadCount(0);
      announceRead();
    } catch (err) {
      console.error('markAllRead failed', err);
    }
  };

  if (authLoading || !user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="container mx-auto max-w-4xl px-4 py-8">
      <div className="flex flex-wrap items-center justify-between gap-3 mb-6">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">Inbox</h1>
          {/* An empty list says "all caught up" itself, so don't repeat it here. */}
          {(unreadCount > 0 || items.length > 0) && (
            <p className="mt-1 text-sm text-gray-600">
              {unreadCount > 0 ? `${unreadCount} unread` : 'All caught up'}
            </p>
          )}
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-sm text-gray-700">
            <input
              type="checkbox"
              checked={unreadOnly}
              onChange={(e) => {
                setUnreadOnly(e.target.checked);
                setPage(1);
              }}
              className="rounded border-gray-300"
            />
            Unread only
          </label>
          {unreadCount > 0 && (
            <button
              type="button"
              onClick={handleMarkAllRead}
              className="text-sm px-3 py-1.5 rounded border border-gray-300 text-gray-900 hover:bg-gray-50"
            >
              Mark all read
            </button>
          )}
        </div>
      </div>

      {isLoading ? (
        <LoadingSpinner />
      ) : error ? (
        <div className="rounded border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      ) : items.length === 0 ? (
        <div className="rounded border border-gray-200 bg-white px-4 py-16 text-center text-gray-500">
          All caught up 🎉
        </div>
      ) : (
        <ul className="divide-y divide-gray-200 rounded border border-gray-200 bg-white">
          {items.map((n) => {
            const isUnread = !n.read_at;
            const href = n.machine_serial_number
              ? machineDetailHref(n.machine_serial_number)
              : undefined;
            const hasOpenFlag = n.flag_status === 'open' && !!n.flag_id;

            const titleNode = <span className="font-medium text-gray-900">{n.title}</span>;

            return (
              <li
                key={n.id}
                className={`px-4 py-4 flex items-start gap-3 ${isUnread ? 'bg-blue-50/40' : ''}`}
              >
                {isUnread ? (
                  <span
                    className="mt-1.5 inline-block h-2 w-2 rounded-full bg-blue-600 shrink-0"
                    aria-label="Unread"
                  />
                ) : (
                  <span className="mt-1.5 inline-block h-2 w-2 shrink-0" aria-hidden="true" />
                )}
                <div className="min-w-0 flex-1">
                  <div className="flex items-start justify-between gap-2">
                    {href ? (
                      <Link href={href} className="hover:underline">
                        {titleNode}
                      </Link>
                    ) : (
                      titleNode
                    )}
                    <span className="text-xs text-gray-500 whitespace-nowrap">
                      {formatWhen(n.created_at)}
                    </span>
                  </div>
                  {n.body && <p className="mt-1 text-sm text-gray-600 break-words">{n.body}</p>}
                  <div className="mt-2 flex flex-wrap items-center gap-3 text-xs text-gray-500">
                    <span className="uppercase tracking-wide">
                      {TYPE_LABELS[n.type] ?? n.type}
                    </span>
                    {n.actor_username && <span>By {n.actor_username}</span>}
                    {isUnread && (
                      <button
                        type="button"
                        onClick={() => handleMarkRead(n.id)}
                        className="text-blue-600 hover:text-blue-800"
                      >
                        Mark as read
                      </button>
                    )}
                    {hasOpenFlag && (
                      // The inbox itself is read-only; flags are worked off on
                      // the flagged records page.
                      <Link href="/flags" className="font-medium text-amber-800 hover:underline">
                        Resolve in flagged records
                      </Link>
                    )}
                    {/* Redundant on a resolution notification, whose chip says as much. */}
                    {n.type === 'flagged' && n.flag_status === 'resolved' && (
                      <span className="font-medium text-green-700">Flag resolved</span>
                    )}
                  </div>
                </div>
              </li>
            );
          })}
        </ul>
      )}

      {count > PAGE_SIZE && (
        <div className="mt-6 flex items-center justify-between">
          <div className="text-sm text-gray-600">
            Page {page} of {totalPages}
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              disabled={page <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              className="text-sm px-3 py-1.5 rounded border border-gray-300 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <button
              type="button"
              disabled={page >= totalPages}
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              className="text-sm px-3 py-1.5 rounded border border-gray-300 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
