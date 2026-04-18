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

const STEP_BUTTON_CLASS =
  'min-w-[44px] min-h-[44px] flex items-center justify-center text-sm text-gray-900 border border-gray-400 rounded hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed bg-white';

function PageSizeSelect({
  limit,
  options,
  disabled,
  onChange,
  compact,
}: {
  limit: number;
  options: number[];
  disabled: boolean;
  onChange: (size: number) => void;
  compact?: boolean;
}) {
  return (
    <select
      value={limit}
      onChange={(e) => onChange(Number(e.target.value))}
      disabled={disabled}
      aria-label="Rows per page"
      className={`border border-gray-300 rounded px-2 py-1 bg-white disabled:opacity-50 ${
        compact ? 'text-xs' : 'text-sm'
      }`}
    >
      {options.map((size) => (
        <option key={size} value={size}>
          {size} / page
        </option>
      ))}
    </select>
  );
}

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

  const rangeSummary = `${rangeStart}-${rangeEnd} of ${total}`;
  const pageIndicator = `Page ${pageIndex + 1} of ${pageCount || 1}`;

  const compactButtonClass =
    'h-9 px-3 flex items-center justify-center gap-1 text-xs font-medium text-gray-900 border border-gray-400 rounded hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed bg-white';

  return (
    <div className="px-3 py-2 border-t border-gray-200 bg-gray-50 sm:px-4 sm:py-3">
      <div className="flex items-center justify-between gap-2 sm:hidden">
        <button
          onClick={() => onPageChange(pageIndex - 1)}
          disabled={!canPreviousPage || loading}
          className={compactButtonClass}
          aria-label="Previous page"
        >
          <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          Prev
        </button>
        <div className="flex flex-col items-center text-center leading-tight min-w-0">
          <span className="text-xs font-medium text-gray-900 tabular-nums whitespace-nowrap">
            {pageIndicator}
          </span>
          <span className="text-[11px] text-gray-500 tabular-nums whitespace-nowrap">
            {rangeSummary}
          </span>
        </div>
        <button
          onClick={() => onPageChange(pageIndex + 1)}
          disabled={!canNextPage || loading}
          className={compactButtonClass}
          aria-label="Next page"
        >
          Next
          <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>

      <div className="hidden sm:flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm text-gray-700">
          <span>{rangeSummary}</span>
          <PageSizeSelect
            limit={limit}
            options={effectiveOptions}
            disabled={loading}
            onChange={onPageSizeChange}
          />
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={() => onPageChange(0)}
            disabled={!canPreviousPage || loading}
            className={STEP_BUTTON_CLASS}
            title="First page"
            aria-label="First page"
          >
            ««
          </button>
          <button
            onClick={() => onPageChange(pageIndex - 1)}
            disabled={!canPreviousPage || loading}
            className={STEP_BUTTON_CLASS}
            title="Previous page"
            aria-label="Previous page"
          >
            «
          </button>
          <span className="px-3 py-1 text-sm text-gray-900">{pageIndicator}</span>
          <button
            onClick={() => onPageChange(pageIndex + 1)}
            disabled={!canNextPage || loading}
            className={STEP_BUTTON_CLASS}
            title="Next page"
            aria-label="Next page"
          >
            »
          </button>
          <button
            onClick={() => onPageChange(pageCount - 1)}
            disabled={!canNextPage || loading}
            className={STEP_BUTTON_CLASS}
            title="Last page"
            aria-label="Last page"
          >
            »»
          </button>
        </div>
      </div>
    </div>
  );
}
