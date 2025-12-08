'use client';

import { useState, useRef, useEffect } from 'react';
import { DayPicker, DateRange } from 'react-day-picker';
import { format } from 'date-fns';
import 'react-day-picker/dist/style.css';

export interface DateRangeValue {
  from: string | undefined; // YYYY-MM-DD format
  to: string | undefined;   // YYYY-MM-DD format
}

interface DateRangePickerProps {
  label: string;
  value: DateRangeValue;
  onChange: (value: DateRangeValue) => void;
  placeholder?: string;
}

export default function DateRangePicker({
  label,
  value,
  onChange,
  placeholder = 'Select date range...',
}: DateRangePickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Convert string dates to Date objects for the picker
  const selectedRange: DateRange | undefined = {
    from: value.from ? new Date(value.from + 'T00:00:00') : undefined,
    to: value.to ? new Date(value.to + 'T00:00:00') : undefined,
  };

  // Handle date selection
  const handleSelect = (range: DateRange | undefined) => {
    onChange({
      from: range?.from ? format(range.from, 'yyyy-MM-dd') : undefined,
      to: range?.to ? format(range.to, 'yyyy-MM-dd') : undefined,
    });
  };

  // Close picker when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Format display text
  const getDisplayText = () => {
    if (value.from && value.to) {
      if (value.from === value.to) {
        return format(new Date(value.from + 'T00:00:00'), 'MMM d, yyyy');
      }
      return `${format(new Date(value.from + 'T00:00:00'), 'MMM d, yyyy')} - ${format(new Date(value.to + 'T00:00:00'), 'MMM d, yyyy')}`;
    }
    if (value.from) {
      return `From ${format(new Date(value.from + 'T00:00:00'), 'MMM d, yyyy')}`;
    }
    return '';
  };

  const hasValue = value.from || value.to;
  const displayText = getDisplayText();

  // Clear the selection
  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation();
    onChange({ from: undefined, to: undefined });
  };

  return (
    <div className="relative" ref={containerRef}>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={`w-full flex items-center justify-between px-3 py-2 border rounded-lg text-sm transition-colors ${
          isOpen
            ? 'border-blue-500 ring-2 ring-blue-500'
            : 'border-gray-300 hover:border-gray-400'
        } bg-white`}
      >
        <div className="flex items-center gap-2">
          {/* Calendar Icon */}
          <svg
            className="h-5 w-5 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
            />
          </svg>
          <span className={hasValue ? 'text-gray-900' : 'text-gray-400'}>
            {hasValue ? displayText : placeholder}
          </span>
        </div>
        <div className="flex items-center gap-1">
          {hasValue && (
            <button
              type="button"
              onClick={handleClear}
              className="p-1 hover:bg-gray-100 rounded-full transition-colors"
              title="Clear selection"
            >
              <svg
                className="h-4 w-4 text-gray-400 hover:text-gray-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          )}
          <svg
            className={`h-4 w-4 text-gray-400 transition-transform ${isOpen ? 'rotate-180' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </button>

      {/* Dropdown Calendar */}
      {isOpen && (
        <div className="absolute top-full left-0 mt-1 z-50 bg-white border border-gray-200 rounded-lg shadow-lg">
          <DayPicker
            mode="range"
            selected={selectedRange}
            onSelect={handleSelect}
            numberOfMonths={1}
            showOutsideDays
            classNames={{
              root: 'p-3',
              months: 'flex flex-col',
              month: 'space-y-4',
              caption: 'flex justify-center pt-1 relative items-center',
              caption_label: 'text-sm font-medium text-gray-900',
              nav: 'space-x-1 flex items-center',
              nav_button: 'h-7 w-7 bg-transparent p-0 opacity-50 hover:opacity-100 inline-flex items-center justify-center rounded-md border border-gray-200 hover:bg-gray-100',
              nav_button_previous: 'absolute left-1',
              nav_button_next: 'absolute right-1',
              table: 'w-full border-collapse space-y-1',
              head_row: 'flex',
              head_cell: 'text-gray-500 rounded-md w-9 font-normal text-[0.8rem]',
              row: 'flex w-full mt-2',
              cell: 'text-center text-sm p-0 relative [&:has([aria-selected])]:bg-blue-50 first:[&:has([aria-selected])]:rounded-l-md last:[&:has([aria-selected])]:rounded-r-md focus-within:relative focus-within:z-20',
              day: 'h-9 w-9 p-0 font-normal aria-selected:opacity-100 hover:bg-gray-100 rounded-md inline-flex items-center justify-center',
              day_selected: 'bg-blue-600 text-white hover:bg-blue-600 hover:text-white focus:bg-blue-600 focus:text-white',
              day_today: 'bg-gray-100 text-gray-900',
              day_outside: 'text-gray-400 opacity-50',
              day_disabled: 'text-gray-400 opacity-50',
              day_range_middle: 'aria-selected:bg-blue-50 aria-selected:text-blue-900',
              day_hidden: 'invisible',
            }}
          />
          <div className="border-t border-gray-200 p-2 flex justify-end gap-2">
            <button
              type="button"
              onClick={() => onChange({ from: undefined, to: undefined })}
              className="px-3 py-1.5 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-md transition-colors"
            >
              Clear
            </button>
            <button
              type="button"
              onClick={() => setIsOpen(false)}
              className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
            >
              Done
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

