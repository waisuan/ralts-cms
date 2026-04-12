'use client';

import { useState, useEffect, useMemo, useCallback, Fragment } from 'react';
import {
  useReactTable,
  getCoreRowModel,
  getExpandedRowModel,
  flexRender,
  createColumnHelper,
  ColumnDef,
  SortingState,
  VisibilityState,
  ExpandedState,
  PaginationState,
} from '@tanstack/react-table';
import { Machine } from '../types/machine';
import { PPM_STATUSES, PPM_STATUS_COLORS } from '../utils/constants';
import { AttachmentService } from '../services/attachmentService';
import { formatDate, formatDateTime } from '../utils/formatters';

interface RecordsTableProps {
  machines: Machine[];
  total: number;
  offset: number;
  limit: number;
  loading: boolean;
  sortBy: string;
  searchQuery?: string;
  onSortChange: (sortBy: string) => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  onView: (serialNumber: string) => void;
  onEdit: (serialNumber: string) => void;
  onDelete: (serialNumber: string) => void;
}

const columnHelper = createColumnHelper<Machine>();

function getPPMStatusDisplay(ppmStatus: string) {
  if (!ppmStatus) return null;
  switch (ppmStatus) {
    case 'overdue':
      return { label: 'Overdue', color: PPM_STATUS_COLORS[PPM_STATUSES.OVERDUE] };
    case 'due':
      return { label: 'Due', color: PPM_STATUS_COLORS[PPM_STATUSES.DUE] };
    case 'almost_due':
      return { label: 'Upcoming', color: PPM_STATUS_COLORS[PPM_STATUSES.ALMOST_DUE] };
    default:
      return null;
  }
}

function useMachineDownload(machine: Machine) {
  const [downloading, setDownloading] = useState(false);

  const handleDownload = async () => {
    if (!machine.attachment || downloading) return;
    setDownloading(true);
    try {
      const blob = await AttachmentService.downloadMachineAttachment(
        machine.serial_number,
        machine.attachment
      );
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = machine.attachment;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to download attachment:', error);
      alert(`Failed to download attachment: ${machine.attachment}`);
    } finally {
      setDownloading(false);
    }
  };

  return { downloading, handleDownload };
}

function AttachmentLink({ machine }: { machine: Machine }) {
  const { downloading, handleDownload } = useMachineDownload(machine);

  return (
    <button
      type="button"
      onClick={(e) => { e.stopPropagation(); handleDownload(); }}
      disabled={downloading}
      className="flex items-center gap-1.5 text-blue-600 hover:text-blue-800 hover:underline transition-colors disabled:opacity-50"
      title={`Download ${machine.attachment}`}
    >
      <svg className="h-4 w-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
        />
      </svg>
      <span>{machine.attachment}</span>
    </button>
  );
}

function MachineAttachmentIcon({ machine }: { machine: Machine }) {
  const { downloading, handleDownload } = useMachineDownload(machine);

  return (
    <button
      type="button"
      onClick={(e) => { e.stopPropagation(); handleDownload(); }}
      disabled={downloading}
      className="text-blue-500 hover:text-blue-700 transition-colors disabled:opacity-50 mx-auto block"
      title={`Download ${machine.attachment}`}
    >
      <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
      </svg>
    </button>
  );
}

function ActionMenu({
  serial,
  onView,
  onEdit,
  onDelete,
}: {
  serial: string;
  onView: (s: string) => void;
  onEdit: (s: string) => void;
  onDelete: (s: string) => void;
}) {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    if (!isOpen) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') { setIsOpen(false); }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen]);

  return (
    <div className="relative">
      <button
        onClick={(e) => {
          e.stopPropagation();
          setIsOpen(!isOpen);
        }}
        aria-haspopup="true"
        aria-expanded={isOpen}
        className="p-1 text-gray-400 hover:text-gray-700 rounded transition-colors"
        title="Actions"
      >
        <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
          <path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
        </svg>
      </button>
      {isOpen && (
        <>
          <div className="fixed inset-0 z-10" onClick={(e) => { e.stopPropagation(); setIsOpen(false); }} />
          <div className="absolute right-0 z-20 mt-1 w-36 bg-white border border-gray-200 rounded-lg shadow-lg py-1">
            <button
              onClick={(e) => { e.stopPropagation(); setIsOpen(false); onView(serial); }}
              className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
            >
              View Details
            </button>
            <button
              onClick={(e) => { e.stopPropagation(); setIsOpen(false); onEdit(serial); }}
              className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
            >
              Edit
            </button>
            <button
              onClick={(e) => { e.stopPropagation(); setIsOpen(false); onDelete(serial); }}
              className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50"
            >
              Delete
            </button>
          </div>
        </>
      )}
    </div>
  );
}

function ExpandedRowDetail({
  machine,
  onView,
}: {
  machine: Machine;
  onView: (serialNumber: string) => void;
}) {
  return (
    <div className="p-4 bg-gray-50 text-sm">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div>
          <span className="font-medium text-gray-500">Model</span>
          <p className="text-gray-900">{machine.model || '-'}</p>
        </div>
        <div>
          <span className="font-medium text-gray-500">Brand</span>
          <p className="text-gray-900">{machine.brand || '-'}</p>
        </div>
        <div>
          <span className="font-medium text-gray-500">Account Type</span>
          <p className="text-gray-900">{machine.account_type || '-'}</p>
        </div>
        <div>
          <span className="font-medium text-gray-500">Reported By</span>
          <p className="text-gray-900">{machine.reported_by || '-'}</p>
        </div>
        <div>
          <span className="font-medium text-gray-500">Attachment</span>
          {machine.attachment ? (
            <AttachmentLink machine={machine} />
          ) : (
            <p className="text-gray-900">None</p>
          )}
        </div>
        <div>
          <span className="font-medium text-gray-500">Created</span>
          <p className="text-gray-900">{formatDateTime(machine.created_at)}</p>
        </div>
        <div>
          <span className="font-medium text-gray-500">Last Updated</span>
          <p className="text-gray-900">{formatDateTime(machine.updated_at)}</p>
        </div>
        <div>
          <span className="font-medium text-gray-500">Maintenance Records</span>
          <p className="text-gray-900">{machine.maintenance_count ?? 0}</p>
        </div>
        {machine.additional_notes && (
          <div className="col-span-2 md:col-span-4">
            <span className="font-medium text-gray-500">Additional Notes</span>
            <p className="text-gray-900 whitespace-pre-wrap">{machine.additional_notes}</p>
          </div>
        )}
      </div>
      <div className="mt-3 pt-3 border-t border-gray-200 flex items-center gap-3">
        <button
          type="button"
          onClick={(e) => { e.stopPropagation(); onView(machine.serial_number); }}
          className="inline-flex items-center gap-1.5 text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
        >
          <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
          </svg>
          View Details & Maintenance History
        </button>
      </div>
    </div>
  );
}

const SORT_MAP: Record<string, { id: string; desc: boolean }> = {
  updated_at_desc: { id: 'updated_at', desc: true },
  updated_at_asc: { id: 'updated_at', desc: false },
  ppm_date_asc: { id: 'ppm_date', desc: false },
  ppm_date_desc: { id: 'ppm_date', desc: true },
  tnc_date_asc: { id: 'tnc_date', desc: false },
  tnc_date_desc: { id: 'tnc_date', desc: true },
};

const REVERSE_SORT_MAP: Record<string, Record<string, string>> = {
  updated_at: { true: 'updated_at_desc', false: 'updated_at_asc' },
  ppm_date: { true: 'ppm_date_desc', false: 'ppm_date_asc' },
  tnc_date: { true: 'tnc_date_desc', false: 'tnc_date_asc' },
};

const DEFAULT_VISIBLE_COLUMNS: VisibilityState = {
  serial_number: true,
  customer: true,
  state: true,
  district: true,
  status: true,
  ppm_date: true,
  tnc_date: true,
  person_in_charge: true,
  updated_at: true,
  model: false,
  brand: false,
  account_type: false,
  reported_by: false,
  attachment: false,
};

const COLUMN_VISIBILITY_STORAGE_KEY = 'ralts-table-columns';

function loadColumnVisibility(): VisibilityState {
  if (typeof window === 'undefined') return DEFAULT_VISIBLE_COLUMNS;
  try {
    const stored = localStorage.getItem(COLUMN_VISIBILITY_STORAGE_KEY);
    if (stored) return JSON.parse(stored) as VisibilityState;
  } catch { /* use default */ }
  return DEFAULT_VISIBLE_COLUMNS;
}

function saveColumnVisibility(visibility: VisibilityState) {
  try {
    localStorage.setItem(COLUMN_VISIBILITY_STORAGE_KEY, JSON.stringify(visibility));
  } catch { /* ignore */ }
}

const COLUMN_LABELS: Record<string, string> = {
  serial_number: 'Serial No',
  customer: 'Customer',
  state: 'State',
  district: 'District',
  status: 'Status',
  ppm_date: 'PPM Date',
  tnc_date: 'TNC Date',
  person_in_charge: 'Assignee',
  updated_at: 'Updated',
  model: 'Model',
  brand: 'Brand',
  account_type: 'Type',
  reported_by: 'Reported By',
  attachment: 'Attachment',
};

export default function RecordsTable({
  machines,
  total,
  offset,
  limit,
  loading,
  sortBy,
  searchQuery,
  onSortChange,
  onPageChange,
  onPageSizeChange,
  onView,
  onEdit,
  onDelete,
}: RecordsTableProps) {
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>(loadColumnVisibility);
  const [expanded, setExpanded] = useState<ExpandedState>({});
  const [showColumnMenu, setShowColumnMenu] = useState(false);

  const handleColumnVisibilityChange = useCallback((updater: VisibilityState | ((old: VisibilityState) => VisibilityState)) => {
    setColumnVisibility((prev) => {
      const next = typeof updater === 'function' ? updater(prev) : updater;
      saveColumnVisibility(next);
      return next;
    });
  }, []);

  const sorting: SortingState = useMemo(() => {
    const mapped = SORT_MAP[sortBy];
    if (mapped) return [{ id: mapped.id, desc: mapped.desc }];
    return [];
  }, [sortBy]);

  const handleSortingChange = useCallback(
    (updater: SortingState | ((old: SortingState) => SortingState)) => {
      const newSorting = typeof updater === 'function' ? updater(sorting) : updater;
      if (newSorting.length > 0) {
        const { id, desc } = newSorting[0];
        const mapped = REVERSE_SORT_MAP[id]?.[String(desc)];
        if (mapped) onSortChange(mapped);
      }
    },
    [sorting, onSortChange]
  );

  const pageIndex = Math.floor(offset / limit);
  const pageCount = Math.ceil(total / limit);

  const pagination: PaginationState = useMemo(
    () => ({ pageIndex, pageSize: limit }),
    [pageIndex, limit]
  );

  const handlePaginationChange = useCallback(
    (updater: PaginationState | ((old: PaginationState) => PaginationState)) => {
      const next = typeof updater === 'function' ? updater(pagination) : updater;
      if (next.pageSize !== limit) {
        onPageSizeChange(next.pageSize);
      }
      if (next.pageIndex !== pageIndex) {
        onPageChange(next.pageIndex);
      }
    },
    [pagination, limit, pageIndex, onPageChange, onPageSizeChange]
  );

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const columns = useMemo<ColumnDef<Machine, any>[]>(
    () => [
      {
        id: 'expander',
        header: () => null,
        cell: ({ row }) => (
          <button
            onClick={(e) => {
              e.stopPropagation();
              row.toggleExpanded();
            }}
            aria-expanded={row.getIsExpanded()}
            aria-label="Toggle row details"
            className="p-1 text-gray-400 hover:text-gray-700 transition-colors"
            title={row.getIsExpanded() ? 'Collapse details' : 'Expand details'}
          >
            <svg
              className={`h-4 w-4 transition-transform ${row.getIsExpanded() ? 'rotate-90' : ''}`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </button>
        ),
        size: 36,
        enableSorting: false,
      },
      columnHelper.accessor('serial_number', {
        header: 'Serial No',
        size: 130,
        cell: (info) => (
          <span className="font-medium text-gray-900 truncate block" title={info.getValue()}>
            {info.getValue()}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('customer', {
        header: 'Customer',
        size: 170,
        cell: (info) => (
          <span className="text-gray-700 truncate block" title={info.getValue() || undefined}>
            {info.getValue() || '-'}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('state', {
        header: 'State',
        size: 100,
        cell: (info) => (
          <span className="text-gray-700 truncate block" title={info.getValue() || undefined}>
            {info.getValue() || '-'}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('district', {
        header: 'District',
        size: 110,
        cell: (info) => (
          <span className="text-gray-700 truncate block" title={info.getValue() || undefined}>
            {info.getValue() || '-'}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('model', {
        header: 'Model',
        size: 100,
        cell: (info) => (
          <span className="text-gray-700 truncate block" title={info.getValue() || undefined}>
            {info.getValue() || '-'}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('brand', {
        header: 'Brand',
        size: 100,
        cell: (info) => (
          <span className="text-gray-700 truncate block" title={info.getValue() || undefined}>
            {info.getValue() || '-'}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('account_type', {
        header: 'Type',
        size: 100,
        cell: (info) => (
          <span className="text-gray-700 truncate block" title={info.getValue() || undefined}>
            {info.getValue() || '-'}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('status', {
        header: 'Status',
        size: 130,
        cell: (info) => (
          <span className="text-gray-700 truncate block" title={info.getValue() || undefined}>
            {info.getValue() || '-'}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('ppm_date', {
        header: 'PPM Date',
        size: 145,
        cell: ({ row }) => {
          const status = getPPMStatusDisplay(row.original.ppm_status);
          return (
            <div className="flex items-center gap-1.5 overflow-hidden whitespace-nowrap">
              <span className="text-gray-700">{formatDate(row.original.ppm_date)}</span>
              {status && (
                <span className={`px-1.5 py-0.5 text-xs font-medium rounded-full shrink-0 ${status.color}`}>
                  {status.label}
                </span>
              )}
            </div>
          );
        },
        enableSorting: true,
      }),
      columnHelper.accessor('tnc_date', {
        header: 'TNC Date',
        size: 95,
        cell: (info) => <span className="text-gray-700 whitespace-nowrap">{formatDate(info.getValue())}</span>,
        enableSorting: true,
      }),
      columnHelper.accessor('person_in_charge', {
        header: 'Assignee',
        size: 130,
        cell: (info) => (
          <span className="text-gray-700 truncate block" title={info.getValue() || undefined}>
            {info.getValue() || '-'}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('reported_by', {
        header: 'Reported By',
        size: 120,
        cell: (info) => (
          <span className="text-gray-700 truncate block" title={info.getValue() || undefined}>
            {info.getValue() || '-'}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('attachment', {
        header: () => (
          <svg className="h-4 w-4 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-label="Attachment">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
          </svg>
        ),
        size: 40,
        cell: ({ row }) => {
          const machine = row.original;
          if (!machine.attachment) return null;
          return <MachineAttachmentIcon machine={machine} />;
        },
        enableSorting: false,
      }),
      columnHelper.accessor('updated_at', {
        header: 'Updated',
        size: 95,
        cell: (info) => (
          <span className="text-gray-500 whitespace-nowrap text-xs" title={formatDateTime(info.getValue())}>
            {formatDate(info.getValue())}
          </span>
        ),
        enableSorting: true,
      }),
      {
        id: 'actions',
        header: () => null,
        cell: ({ row }) => (
          <ActionMenu
            serial={row.original.serial_number}
            onView={onView}
            onEdit={onEdit}
            onDelete={onDelete}
          />
        ),
        size: 44,
        enableSorting: false,
      },
    ],
    [onView, onEdit, onDelete]
  );

  const table = useReactTable({
    data: machines,
    columns,
    state: {
      sorting,
      columnVisibility,
      expanded,
      pagination,
    },
    onSortingChange: handleSortingChange,
    onColumnVisibilityChange: handleColumnVisibilityChange,
    onExpandedChange: setExpanded,
    onPaginationChange: handlePaginationChange,
    getCoreRowModel: getCoreRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    manualSorting: true,
    enableSortingRemoval: false,
    manualPagination: true,
    pageCount,
    rowCount: total,
    getRowId: (row) => row.serial_number,
    getRowCanExpand: () => true,
  });

  const toggleableColumns = table.getAllLeafColumns().filter(
    (col) => col.id !== 'expander' && col.id !== 'actions'
  );

  return (
    <div>
      {/* Column visibility toggle */}
      <div className="flex justify-end mb-2 relative">
        <button
          onClick={() => setShowColumnMenu(!showColumnMenu)}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm border border-gray-300 rounded-lg bg-white hover:bg-gray-50 text-gray-700 transition-colors"
        >
          <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4h18M3 8h18M3 12h18M3 16h18M3 20h18" />
          </svg>
          Columns
        </button>
        {showColumnMenu && (
          <>
            <div className="fixed inset-0 z-10" onClick={() => setShowColumnMenu(false)} />
            <div className="absolute right-0 top-full mt-1 z-20 w-52 bg-white border border-gray-200 rounded-lg shadow-lg py-2 max-h-80 overflow-y-auto">
              <div className="flex items-center justify-between px-3 py-1.5">
                <span className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Toggle Columns
                </span>
                <button
                  onClick={() => handleColumnVisibilityChange(DEFAULT_VISIBLE_COLUMNS)}
                  className="text-xs text-blue-600 hover:text-blue-800"
                >
                  Reset
                </button>
              </div>
              {toggleableColumns.map((col) => (
                <label
                  key={col.id}
                  className="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50 cursor-pointer"
                >
                  <input
                    type="checkbox"
                    checked={col.getIsVisible()}
                    onChange={col.getToggleVisibilityHandler()}
                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  {COLUMN_LABELS[col.id] || col.id}
                </label>
              ))}
            </div>
          </>
        )}
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full table-fixed" aria-label="Machines list">
            <thead className="bg-gray-50 border-b border-gray-200">
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id}>
                  {headerGroup.headers.map((header) => {
                    const canSort = header.column.getCanSort();
                    const sorted = header.column.getIsSorted();
                    return (
                      <th
                        key={header.id}
                        scope="col"
                        aria-sort={
                          canSort
                            ? sorted === 'asc'
                              ? 'ascending'
                              : sorted === 'desc'
                                ? 'descending'
                                : 'none'
                            : undefined
                        }
                        className={`px-3 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider ${
                          canSort ? 'cursor-pointer select-none hover:text-gray-700' : ''
                        }`}
                        style={{ width: header.getSize() }}
                        onClick={canSort ? header.column.getToggleSortingHandler() : undefined}
                      >
                        <div className="flex items-center gap-1">
                          {header.isPlaceholder
                            ? null
                            : flexRender(header.column.columnDef.header, header.getContext())}
                          {canSort && (
                            <span className="text-gray-400">
                              {sorted === 'asc' ? '▲' : sorted === 'desc' ? '▼' : '⇅'}
                            </span>
                          )}
                        </div>
                      </th>
                    );
                  })}
                </tr>
              ))}
            </thead>
            <tbody className="divide-y divide-gray-200">
              {loading && machines.length === 0 ? (
                <tr>
                  <td colSpan={table.getVisibleLeafColumns().length} className="px-6 py-12 text-center">
                    <div className="flex items-center justify-center gap-2">
                      <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600" />
                      <span className="text-gray-600">Loading machines...</span>
                    </div>
                  </td>
                </tr>
              ) : machines.length === 0 ? (
                <tr>
                  <td colSpan={table.getVisibleLeafColumns().length} className="px-6 py-12 text-center text-gray-500">
                    {searchQuery
                      ? `No machines match "${searchQuery}". Try a different search term.`
                      : 'No machines found.'}
                  </td>
                </tr>
              ) : (
                table.getRowModel().rows.map((row) => (
                  <Fragment key={row.id}>
                    <tr
                      className="hover:bg-gray-50 cursor-pointer transition-colors"
                      onClick={() => row.toggleExpanded()}
                    >
                      {row.getVisibleCells().map((cell) => (
                        <td key={cell.id} className="px-3 py-2.5 text-sm min-w-0">
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </td>
                      ))}
                    </tr>
                    {row.getIsExpanded() && (
                      <tr>
                        <td colSpan={row.getVisibleCells().length}>
                          <ExpandedRowDetail machine={row.original} onView={onView} />
                        </td>
                      </tr>
                    )}
                  </Fragment>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {total > 0 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-gray-200 bg-gray-50">
            <div className="flex items-center gap-2 text-sm text-gray-700">
              <span>
                {offset + 1}-{Math.min(offset + limit, total)} of {total}
              </span>
              <select
                value={limit}
                onChange={(e) => onPageSizeChange(Number(e.target.value))}
                aria-label="Rows per page"
                className="border border-gray-300 rounded px-2 py-1 text-sm bg-white"
              >
                {[50, 100].map((size) => (
                  <option key={size} value={size}>
                    {size} / page
                  </option>
                ))}
              </select>
            </div>
            <div className="flex items-center gap-1">
              <button
                onClick={() => table.firstPage()}
                disabled={!table.getCanPreviousPage()}
                className="px-2 py-1 text-sm text-gray-900 border border-gray-400 rounded hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed bg-white"
                title="First page"
              >
                ««
              </button>
              <button
                onClick={() => table.previousPage()}
                disabled={!table.getCanPreviousPage()}
                className="px-2 py-1 text-sm text-gray-900 border border-gray-400 rounded hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed bg-white"
                title="Previous page"
              >
                «
              </button>
              <span className="px-3 py-1 text-sm text-gray-900">
                Page {pageIndex + 1} of {pageCount || 1}
              </span>
              <button
                onClick={() => table.nextPage()}
                disabled={!table.getCanNextPage()}
                className="px-2 py-1 text-sm text-gray-900 border border-gray-400 rounded hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed bg-white"
                title="Next page"
              >
                »
              </button>
              <button
                onClick={() => table.lastPage()}
                disabled={!table.getCanNextPage()}
                className="px-2 py-1 text-sm text-gray-900 border border-gray-400 rounded hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed bg-white"
                title="Last page"
              >
                »»
              </button>
            </div>
          </div>
        )}
      </div>

    </div>
  );
}
