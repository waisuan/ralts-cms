interface PaginationControlsProps {
  pageIndex: number;
  pageCount: number;
  limit: number;
  total: number;
  onPageChange: (pageIndex: number) => void;
  onPageSizeChange: (size: number) => void;
  pageSizeOptions?: number[];
  loading?: boolean;
}

const DEFAULT_PAGE_SIZE_OPTIONS = [50, 100];

export default function PaginationControls({
  pageIndex,
  pageCount,
  limit,
  total,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = DEFAULT_PAGE_SIZE_OPTIONS,
  loading = false,
}: PaginationControlsProps) {
  const rangeStart = total === 0 ? 0 : pageIndex * limit + 1;
  const rangeEnd = Math.min((pageIndex + 1) * limit, total);
  const canPreviousPage = pageIndex > 0;
  const canNextPage = pageIndex < pageCount - 1;

  const effectiveOptions = pageSizeOptions.includes(limit)
    ? pageSizeOptions
    : [...pageSizeOptions, limit].sort((a, b) => a - b);

  return (
    <div className="flex items-center justify-between px-4 py-3 border-t border-gray-200 bg-gray-50">
      <div className="flex items-center gap-2 text-sm text-gray-700">
        <span>
          {rangeStart}-{rangeEnd} of {total}
        </span>
        <select
          value={limit}
          onChange={(e) => onPageSizeChange(Number(e.target.value))}
          disabled={loading}
          aria-label="Rows per page"
          className="border border-gray-300 rounded px-2 py-1 text-sm bg-white disabled:opacity-50"
        >
          {effectiveOptions.map((size) => (
            <option key={size} value={size}>
              {size} / page
            </option>
          ))}
        </select>
      </div>
      <div className="flex items-center gap-1">
        <button
          onClick={() => onPageChange(0)}
          disabled={!canPreviousPage || loading}
          className="min-w-[44px] min-h-[44px] flex items-center justify-center text-sm text-gray-900 border border-gray-400 rounded hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed bg-white"
          title="First page"
          aria-label="First page"
        >
          ««
        </button>
        <button
          onClick={() => onPageChange(pageIndex - 1)}
          disabled={!canPreviousPage || loading}
          className="min-w-[44px] min-h-[44px] flex items-center justify-center text-sm text-gray-900 border border-gray-400 rounded hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed bg-white"
          title="Previous page"
          aria-label="Previous page"
        >
          «
        </button>
        <span className="px-3 py-1 text-sm text-gray-900">
          Page {pageIndex + 1} of {pageCount || 1}
        </span>
        <button
          onClick={() => onPageChange(pageIndex + 1)}
          disabled={!canNextPage || loading}
          className="min-w-[44px] min-h-[44px] flex items-center justify-center text-sm text-gray-900 border border-gray-400 rounded hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed bg-white"
          title="Next page"
          aria-label="Next page"
        >
          »
        </button>
        <button
          onClick={() => onPageChange(pageCount - 1)}
          disabled={!canNextPage || loading}
          className="min-w-[44px] min-h-[44px] flex items-center justify-center text-sm text-gray-900 border border-gray-400 rounded hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed bg-white"
          title="Last page"
          aria-label="Last page"
        >
          »»
        </button>
      </div>
    </div>
  );
}
