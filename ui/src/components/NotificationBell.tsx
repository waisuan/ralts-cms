'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { NotificationService, Notification } from '../services/notificationService';
import { machineDetailHref } from '../utils/machineRoutes';

function formatRelativeTime(iso: string): string {
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return '';
  const diffMs = Date.now() - then;
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) return 'just now';
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  return new Date(iso).toLocaleDateString();
}

/**
 * NotificationBell renders a small bell button with unread badge in the header.
 *
 * There is no background timer: the unread count is read once per page load and
 * again on each navigation, so the badge always reflects the state as of the
 * page the user is looking at. The dropdown fetches the latest few
 * notifications when it is opened.
 */
export default function NotificationBell() {
  const { isAuthenticated } = useAuth();
  const pathname = usePathname();
  const [unread, setUnread] = useState<number>(0);
  const [isOpen, setIsOpen] = useState(false);
  const [items, setItems] = useState<Notification[]>([]);
  const [loadingItems, setLoadingItems] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const refreshUnread = useCallback(async () => {
    try {
      const res = await NotificationService.unreadCount();
      setUnread(res.data?.unread_count ?? 0);
    } catch (err) {
      // Non-fatal: the badge is supplementary and the next load retries.
      console.debug('unread count fetch failed', err);
    }
  }, []);

  useEffect(() => {
    if (!isAuthenticated) {
      setUnread(0);
      return;
    }
    refreshUnread();
  }, [isAuthenticated, pathname, refreshUnread]);

  // Close dropdown on outside click.
  useEffect(() => {
    if (!isOpen) return;
    const handler = (e: MouseEvent) => {
      if (!containerRef.current) return;
      if (!containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [isOpen]);

  const loadRecent = useCallback(async () => {
    setLoadingItems(true);
    setError(null);
    try {
      const res = await NotificationService.list({ limit: 10 });
      setItems(res.data?.notifications ?? []);
      setUnread(res.data?.unread_count ?? 0);
    } catch (err) {
      console.error('Failed to load notifications:', err);
      setError('Could not load notifications');
    } finally {
      setLoadingItems(false);
    }
  }, []);

  const handleToggle = () => {
    setIsOpen((prev) => {
      const next = !prev;
      if (next) loadRecent();
      return next;
    });
  };

  const handleItemClick = async (n: Notification) => {
    if (n.read_at) return;
    // Optimistically mark as read; refresh unread count after.
    try {
      await NotificationService.markRead(n.id);
    } catch (err) {
      console.debug('markRead failed', err);
    }
    setItems((prev) =>
      prev.map((x) => (x.id === n.id ? { ...x, read_at: new Date().toISOString() } : x)),
    );
    setUnread((u) => Math.max(0, u - 1));
  };

  const handleMarkAll = async () => {
    try {
      await NotificationService.markAllRead();
      setUnread(0);
      setItems((prev) => prev.map((x) => ({ ...x, read_at: x.read_at ?? new Date().toISOString() })));
    } catch (err) {
      console.error('markAllRead failed', err);
    }
  };

  if (!isAuthenticated) return null;

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        aria-label="Notifications"
        onClick={handleToggle}
        className="relative rounded-full p-2 text-gray-600 hover:bg-gray-100 hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={1.75}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0"
          />
        </svg>
        {unread > 0 && (
          <span className="absolute -top-0.5 -right-0.5 inline-flex h-5 min-w-[1.25rem] items-center justify-center rounded-full bg-red-600 px-1 text-xs font-semibold text-white">
            {unread > 99 ? '99+' : unread}
          </span>
        )}
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-80 rounded-lg border border-gray-200 bg-white shadow-lg z-50">
          <div className="flex items-center justify-between px-4 py-3 border-b border-gray-100">
            <span className="text-sm font-semibold text-gray-900">Notifications</span>
            <button
              type="button"
              onClick={handleMarkAll}
              disabled={unread === 0}
              className="text-xs text-blue-600 hover:text-blue-800 disabled:text-gray-400 disabled:cursor-not-allowed"
            >
              Mark all read
            </button>
          </div>

          <div className="max-h-96 overflow-y-auto">
            {loadingItems && (
              <div className="p-4 text-sm text-gray-500">Loading...</div>
            )}
            {error && <div className="p-4 text-sm text-red-600">{error}</div>}
            {!loadingItems && !error && items.length === 0 && (
              <div className="p-4 text-sm text-gray-500">No notifications</div>
            )}
            {items.map((n) => {
              const isUnread = !n.read_at;
              const href = n.machine_serial_number
                ? machineDetailHref(n.machine_serial_number)
                : '/inbox';
              return (
                <Link
                  key={n.id}
                  href={href}
                  onClick={() => {
                    handleItemClick(n);
                    setIsOpen(false);
                  }}
                  className={`block px-4 py-3 text-sm border-b border-gray-100 hover:bg-gray-50 ${
                    isUnread ? 'bg-blue-50/50' : ''
                  }`}
                >
                  <div className="flex items-start gap-2">
                    {isUnread && (
                      <span className="mt-1 inline-block h-2 w-2 rounded-full bg-blue-600 shrink-0" />
                    )}
                    <div className="min-w-0 flex-1">
                      <div className="font-medium text-gray-900 truncate">{n.title}</div>
                      {n.body && (
                        <div className="text-xs text-gray-600 mt-0.5 truncate">{n.body}</div>
                      )}
                      <div className="text-xs text-gray-400 mt-1">
                        {formatRelativeTime(n.created_at)}
                      </div>
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>

          <div className="px-4 py-2 border-t border-gray-100 text-center">
            <Link
              href="/inbox"
              onClick={() => setIsOpen(false)}
              className="text-sm text-blue-600 hover:text-blue-800"
            >
              View all
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}
