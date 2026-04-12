import DateRangePicker, { type DateRangeValue } from '@/components/DateRangePicker';
import { formatLocalDate } from '@/utils/formatters';

interface RecordsListDateFiltersPanelProps {
  ppmDateRange: DateRangeValue;
  tncDateRange: DateRangeValue;
  onPpmChange: (v: DateRangeValue) => void;
  onTncChange: (v: DateRangeValue) => void;
  onClearAll: () => void;
}

export default function RecordsListDateFiltersPanel({
  ppmDateRange,
  tncDateRange,
  onPpmChange,
  onTncChange,
  onClearAll,
}: RecordsListDateFiltersPanelProps) {
  return (
    <div className="mb-6 p-4 bg-gray-50 border border-gray-200 rounded-lg">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-gray-700">Filter by Date Range</h3>
        {(ppmDateRange.from || tncDateRange.from) && (
          <button type="button" onClick={onClearAll} className="text-sm text-blue-600 hover:text-blue-800 font-medium">
            Clear All Filters
          </button>
        )}
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <DateRangePicker
          label="PPM Date Range"
          value={ppmDateRange}
          onChange={onPpmChange}
          placeholder="Filter by PPM date..."
        />
        <DateRangePicker
          label="TNC Date Range"
          value={tncDateRange}
          onChange={onTncChange}
          placeholder="Filter by TNC date..."
        />
      </div>
      {(ppmDateRange.from || tncDateRange.from) && (
        <div className="mt-3 flex flex-wrap gap-2">
          {ppmDateRange.from && (
            <span className="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
              PPM:{' '}
              {ppmDateRange.from === ppmDateRange.to || !ppmDateRange.to
                ? formatLocalDate(ppmDateRange.from)
                : `${formatLocalDate(ppmDateRange.from)} to ${formatLocalDate(ppmDateRange.to!)}`}
              <button
                type="button"
                onClick={() => onPpmChange({ from: undefined, to: undefined })}
                className="hover:bg-blue-200 rounded-full p-0.5"
              >
                <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </span>
          )}
          {tncDateRange.from && (
            <span className="inline-flex items-center gap-1 px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full">
              TNC:{' '}
              {tncDateRange.from === tncDateRange.to || !tncDateRange.to
                ? formatLocalDate(tncDateRange.from)
                : `${formatLocalDate(tncDateRange.from)} to ${formatLocalDate(tncDateRange.to!)}`}
              <button
                type="button"
                onClick={() => onTncChange({ from: undefined, to: undefined })}
                className="hover:bg-green-200 rounded-full p-0.5"
              >
                <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </span>
          )}
        </div>
      )}
    </div>
  );
}
