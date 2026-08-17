'use client';

import Link from 'next/link';
import { useCallback, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import { USER_ROLE } from '@/services/adminUserService';
import { Flag, FlagReason, FlagScope, FlagService, FlagStatusFilter } from '@/services/flagService';
import { machineDetailHref } from '@/utils/machineRoutes';
import { flaggedAtLabel, resolvedAtLabel } from '@/utils/flagTimestamps';
import LoadingSpinner from '@/components/LoadingSpinner';
import ResolveFlagModal from '@/components/ResolveFlagModal';

const PAGE_SIZE = 50;

const STATUS_TABS: { value: FlagStatusFilter; label: string }[] = [
  { value: 'open', label: 'Open' },
  { value: 'resolved', label: 'Resolved' },
  { value: 'all', label: 'All' },
];

function reasonBadgeClass(reason: FlagReason): string {
  switch (reason) {
    case 'requires_attention':
      return 'bg-orange-100 text-orange-800 border-orange-200';
    case 'missing_values':
      return 'bg-amber-100 text-amber-800 border-amber-200';
    case 'other':
      return 'bg-purple-100 text-purple-800 border-purple-200';
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200';
  }
}

/**
 * Flagged records — where flags are worked off. Admins see every flag; everyone
 * else sees flags on the machines assigned to them, which is also the set they
 * are allowed to resolve. The API applies the scoping, so this page only has to
 * word it. The inbox is read-only and points here.
 */
export default function FlagsPage() {
  const { user, isLoading: authLoading } = useAuth();
  const router = useRouter();

  const [flags, setFlags] = useState<Flag[]>([]);
  const [count, setCount] = useState(0);
  const [scope, setScope] = useState<FlagScope | null>(null);
  const [status, setStatus] = useState<FlagStatusFilter>('open');
  const [page, setPage] = useState(1);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [flagToResolve, setFlagToResolve] = useState<Flag | null>(null);

  const isAdmin = user?.role === USER_ROLE.ADMIN;

  useEffect(() => {
    if (!authLoading && !user) {
      router.replace('/');
    }
  }, [authLoading, user, router]);

  const load = useCallback(async (nextStatus: FlagStatusFilter, nextPage: number) => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await FlagService.listAll({
        status: nextStatus,
        limit: PAGE_SIZE,
        offset: (nextPage - 1) * PAGE_SIZE,
      });
      setFlags(res.data?.flags ?? []);
      setCount(res.data?.count ?? 0);
      setScope(res.data?.scope ?? null);
    } catch (err) {
      console.error('Failed to load flags:', err);
      setError('Failed to load flagged records');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!user) return;
    load(status, page);
  }, [user, status, page, load]);

  const totalPages = Math.max(1, Math.ceil(count / PAGE_SIZE));
  const showsEveryMachine = (scope ?? (isAdmin ? 'all' : 'assigned')) === 'all';

  if (authLoading || !user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="container mx-auto max-w-5xl px-4 py-8">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">
          {showsEveryMachine ? 'Flagged records' : 'My flagged records'}
        </h1>
        <p className="mt-1 text-sm text-gray-600">
          {showsEveryMachine
            ? 'Every machine flagged for follow-up. Resolve a flag once the work is done.'
            : 'Flags raised on machines assigned to you. Resolve a flag once the work is done.'}
        </p>
      </div>

      <div className="mb-4 flex flex-wrap items-center gap-2">
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value}
            type="button"
            aria-pressed={status === tab.value}
            onClick={() => {
              setStatus(tab.value);
              setPage(1);
            }}
            className={`text-sm px-3 py-1.5 rounded border ${
              status === tab.value
                ? 'border-blue-600 bg-blue-50 text-blue-700 font-medium'
                : 'border-gray-300 text-gray-700 hover:bg-gray-50'
            }`}
          >
            {tab.label}
          </button>
        ))}
        <span className="ml-auto text-sm text-gray-600">
          {count} {count === 1 ? 'flag' : 'flags'}
        </span>
      </div>

      {/* Explains the gap when a machine was flagged before it was assigned here. */}
      {!showsEveryMachine && status !== 'open' && (
        <p className="mb-4 text-sm text-gray-500">
          Resolved flags are limited to ones you closed yourself.
        </p>
      )}

      {error && (
        <div className="mb-4 rounded border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {isLoading ? (
        <LoadingSpinner />
      ) : flags.length === 0 ? (
        <div className="rounded border border-gray-200 bg-white px-4 py-16 text-center text-gray-500">
          {status === 'open' ? 'No open flags. Nothing needs attention.' : 'No flags to display.'}
        </div>
      ) : (
        <ul className="divide-y divide-gray-200 rounded border border-gray-200 bg-white">
          {flags.map((flag) => {
            const isOpen = flag.status === 'open';
            const resolvedAt = resolvedAtLabel(flag);
            return (
              <li key={flag.id} className="px-4 py-4">
                <div className="flex flex-wrap items-center gap-2">
                  <Link
                    href={machineDetailHref(flag.machine_serial_number)}
                    className="font-medium text-blue-600 hover:text-blue-800"
                  >
                    {flag.machine_serial_number}
                  </Link>
                  <span
                    className={`inline-flex items-center rounded border px-2 py-0.5 text-xs font-medium ${reasonBadgeClass(
                      flag.reason
                    )}`}
                  >
                    {FlagService.reasonLabel(flag.reason)}
                    {flag.ppm_status ? ` · ${flag.ppm_status}` : ''}
                  </span>
                  {!isOpen && (
                    <span className="inline-flex items-center rounded border border-gray-200 bg-gray-50 px-2 py-0.5 text-xs text-gray-600">
                      resolved
                    </span>
                  )}
                </div>
                {flag.note && (
                  <p className="mt-2 text-sm text-gray-800 whitespace-pre-line">{flag.note}</p>
                )}
                {flag.resolution_note && (
                  <p className="mt-2 border-l-2 border-gray-200 pl-3 text-sm text-gray-600 whitespace-pre-line">
                    <span className="font-medium text-gray-500">Resolution: </span>
                    {flag.resolution_note}
                  </p>
                )}
                <div className="mt-2 flex flex-wrap items-start justify-between gap-2 text-xs text-gray-500">
                  {/* Each event keeps its own time, so a resolved flag shows both. */}
                  <div className="space-y-0.5">
                    <p>{flaggedAtLabel(flag)}</p>
                    {resolvedAt && <p>{resolvedAt}</p>}
                  </div>
                  {isOpen && (
                    <button
                      type="button"
                      onClick={() => setFlagToResolve(flag)}
                      className="rounded border border-gray-300 px-2 py-1 font-medium text-blue-600 hover:bg-gray-50"
                    >
                      Mark resolved
                    </button>
                  )}
                </div>
              </li>
            );
          })}
        </ul>
      )}

      {flagToResolve && (
        <ResolveFlagModal
          flag={flagToResolve}
          onClose={() => setFlagToResolve(null)}
          onResolved={() => {
            void load(status, page);
          }}
        />
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
