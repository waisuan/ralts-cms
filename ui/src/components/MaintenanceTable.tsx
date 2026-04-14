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
  ExpandedState,
  PaginationState,
} from '@tanstack/react-table';
import { Maintenance } from '../types/maintenance';
import { AttachmentService } from '../services/attachmentService';
import { isAuthError } from '../utils/auth';
import { handleApiError } from '../utils/api';
import {
  formatDate,
  formatDateTime,
  getTypeColor,
  isCustomWorkOrderType,
  workOrderTypePillLabel,
} from '../utils/formatters';
import PaginationControls from './PaginationControls';

interface MaintenanceTableProps {
  machineSerialNumber: string;
  records: Maintenance[];
  total: number;
  currentPage: number;
  limit: number;
  loading: boolean;
  sortBy: string;
  searchQuery?: string;
  onSortChange: (sort: string) => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  onEdit: (record: Maintenance) => void;
  onDelete: (record: Maintenance) => void;
}

const columnHelper = createColumnHelper<Maintenance>();

const SORT_MAP: Record<string, { id: string; desc: boolean }> = {
  updated_at_desc: { id: 'updated_at', desc: true },
  updated_at_asc: { id: 'updated_at', desc: false },
  work_order_date_desc: { id: 'work_order_date', desc: true },
  work_order_date_asc: { id: 'work_order_date', desc: false },
};

const REVERSE_SORT_MAP: Record<string, Record<string, string>> = {
  updated_at: { true: 'updated_at_desc', false: 'updated_at_asc' },
  work_order_date: { true: 'work_order_date_desc', false: 'work_order_date_asc' },
};


function useMaintenanceDownload(machineSerialNumber: string, record: Maintenance) {
  const [downloading, setDownloading] = useState(false);

  const handleDownload = async () => {
    if (!record.attachment || downloading) return;
    setDownloading(true);
    try {
      const blob = await AttachmentService.downloadMaintenanceAttachment(
        machineSerialNumber,
        record.work_order_number,
        record.attachment
      );
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = record.attachment;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to download maintenance attachment:', error);
      if (isAuthError(error)) {
        alert('Your session has expired. Please log in again.');
        window.location.href = '/login';
        return;
      }
      const errorMessage = handleApiError(error);
      alert(`Failed to download attachment: ${errorMessage}`);
    } finally {
      setDownloading(false);
    }
  };

  return { downloading, handleDownload };
}

function MaintenanceAttachmentLink({
  machineSerialNumber,
  record,
}: {
  machineSerialNumber: string;
  record: Maintenance;
}) {
  const { downloading, handleDownload } = useMaintenanceDownload(machineSerialNumber, record);

  return (
    <button
      type="button"
      onClick={(e) => { e.stopPropagation(); handleDownload(); }}
      disabled={downloading}
      className="flex items-center gap-1.5 text-blue-600 hover:text-blue-800 hover:underline transition-colors disabled:opacity-50"
      title={`Download ${record.attachment}`}
    >
      <svg className="h-4 w-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
        />
      </svg>
      <span className="text-xs">{record.attachment}</span>
    </button>
  );
}


function MaintenanceActionMenu({
  record,
  onEdit,
  onDelete,
}: {
  record: Maintenance;
  onEdit: (r: Maintenance) => void;
  onDelete: (r: Maintenance) => void;
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
              onClick={(e) => { e.stopPropagation(); setIsOpen(false); onEdit(record); }}
              className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center gap-2"
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              Edit
            </button>
            <button
              onClick={(e) => { e.stopPropagation(); setIsOpen(false); onDelete(record); }}
              className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              Delete
            </button>
          </div>
        </>
      )}
    </div>
  );
}

function MaintenanceExpandedRow({
  machineSerialNumber,
  record,
}: {
  machineSerialNumber: string;
  record: Maintenance;
}) {
  return (
    <div className="p-4 bg-gray-50 text-sm">
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        <div>
          <span className="font-medium text-gray-500">Attachment</span>
          {record.attachment ? (
            <MaintenanceAttachmentLink machineSerialNumber={machineSerialNumber} record={record} />
          ) : (
            <p className="text-gray-400">None</p>
          )}
        </div>
        <div>
          <span className="font-medium text-gray-500">Created</span>
          <p className="text-gray-900">{formatDateTime(record.created_at)}</p>
        </div>
        <div>
          <span className="font-medium text-gray-500">Last Updated</span>
          <p className="text-gray-900">{formatDateTime(record.updated_at)}</p>
        </div>
      </div>
      {isCustomWorkOrderType(record) && (
        <div className="mt-3 pt-3 border-t border-gray-200">
          <span className="font-medium text-gray-500">Maintenance type</span>
          <p className="text-gray-900 mt-1">{record.work_order_type}</p>
        </div>
      )}
      {record.action_taken && (
        <div className="mt-3 pt-3 border-t border-gray-200">
          <span className="font-medium text-gray-500">Action Summary</span>
          <p className="text-gray-900 whitespace-pre-wrap mt-1">{record.action_taken}</p>
        </div>
      )}
    </div>
  );
}

export default function MaintenanceTable({
  machineSerialNumber,
  records,
  total,
  currentPage,
  limit,
  loading,
  sortBy,
  searchQuery,
  onSortChange,
  onPageChange,
  onPageSizeChange,
  onEdit,
  onDelete,
}: MaintenanceTableProps) {
  const [expanded, setExpanded] = useState<ExpandedState>({});

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

  const pageIndex = currentPage - 1;
  const pageCount = Math.ceil(total / limit);

  const pagination: PaginationState = useMemo(
    () => ({ pageIndex, pageSize: limit }),
    [pageIndex, limit]
  );

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const columns = useMemo<ColumnDef<Maintenance, any>[]>(
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
      columnHelper.accessor('work_order_number', {
        header: 'Work Order',
        cell: ({ row }) => (
          <span className="font-medium text-gray-900 whitespace-nowrap overflow-hidden text-ellipsis flex items-center gap-1" title={row.original.work_order_number}>
            {row.original.work_order_number}
            {row.original.attachment && (
              <svg className="h-3 w-3 text-blue-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-label="Has attachment">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
              </svg>
            )}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('work_order_date', {
        header: 'Date',
        cell: (info) => (
          <span className="text-gray-700 whitespace-nowrap">{formatDate(info.getValue())}</span>
        ),
        enableSorting: true,
      }),
      columnHelper.accessor('work_order_type', {
        header: 'Type',
        cell: (info) => {
          const record = info.row.original;
          const label = workOrderTypePillLabel(record);
          return (
            <span
              className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getTypeColor(label)}`}
            >
              {label}
            </span>
          );
        },
        enableSorting: false,
      }),
      columnHelper.accessor('action_taken', {
        header: 'Action Summary',
        cell: (info) => {
          const text = info.getValue() || '';
          return (
            <span className="text-gray-700 truncate block" title={text}>
              {text || '-'}
            </span>
          );
        },
        enableSorting: false,
      }),
      columnHelper.accessor('reported_by', {
        header: 'Reported By',
        cell: (info) => (
          <span className="text-gray-700 truncate block" title={info.getValue() || ''}>
            {info.getValue() || '-'}
          </span>
        ),
        enableSorting: false,
      }),
      columnHelper.accessor('updated_at', {
        header: 'Updated',
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
          <MaintenanceActionMenu
            record={row.original}
            onEdit={onEdit}
            onDelete={onDelete}
          />
        ),
        size: 44,
        enableSorting: false,
      },
    ],
    [onEdit, onDelete]
  );

  const table = useReactTable({
    data: records,
    columns,
    state: {
      sorting,
      expanded,
      pagination,
    },
    onSortingChange: handleSortingChange,
    onExpandedChange: setExpanded,
    getCoreRowModel: getCoreRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    manualSorting: true,
    enableSortingRemoval: false,
    manualPagination: true,
    pageCount,
    rowCount: total,
    getRowId: (row) => row.work_order_number,
    getRowCanExpand: () => true,
  });

  const isLoadingWithData = loading && records.length > 0;

  return (
    <div className="bg-white rounded-lg shadow-sm border overflow-hidden relative" aria-busy={isLoadingWithData}>
      {isLoadingWithData && (
        <div className="absolute inset-0 bg-white bg-opacity-60 z-10 flex items-center justify-center">
          <div className="flex items-center gap-2">
            <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600" />
            <span className="text-sm text-gray-600">Loading...</span>
          </div>
        </div>
      )}
      <div className="overflow-x-auto">
        <table className="w-full table-fixed" aria-label="Maintenance records">
          <colgroup>
            <col className="w-8" />
            <col className="w-[17%]" />
            <col className="w-[11%]" />
            <col className="w-[10%]" />
            <col className="w-[18%]" />
            <col className="w-[15%]" />
            <col className="w-[17%]" />
            <col className="w-10" />
          </colgroup>
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
            {loading && records.length === 0 ? (
              <tr>
                <td colSpan={table.getVisibleLeafColumns().length} className="px-6 py-12 text-center">
                  <div className="flex items-center justify-center gap-2">
                    <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600" />
                    <span className="text-gray-600">Loading maintenance records...</span>
                  </div>
                </td>
              </tr>
            ) : records.length === 0 ? (
              <tr>
                <td colSpan={table.getVisibleLeafColumns().length} className="px-6 py-12 text-center text-gray-500">
                  {searchQuery
                    ? `No records match "${searchQuery}". Try a different search term.`
                    : 'No maintenance records found.'}
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
                        <MaintenanceExpandedRow
                          machineSerialNumber={machineSerialNumber}
                          record={row.original}
                        />
                      </td>
                    </tr>
                  )}
                </Fragment>
              ))
            )}
          </tbody>
        </table>
      </div>

      {total > 0 && (
        <PaginationControls
          pageIndex={currentPage - 1}
          pageCount={pageCount}
          limit={limit}
          total={total}
          onPageChange={(idx) => onPageChange(idx + 1)}
          onPageSizeChange={onPageSizeChange}
          loading={loading}
        />
      )}
    </div>
  );
}
