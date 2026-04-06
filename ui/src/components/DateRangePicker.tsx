'use client';

import { useState, useRef, useEffect } from 'react';
import { DayPicker, DateRange } from 'react-day-picker';
import { format } from 'date-fns';
import 'react-day-picker/style.css';

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
  
  // Internal pending state - only applied when "Done" is clicked
  const [pendingRange, setPendingRange] = useState<DateRange | undefined>(undefined);

  // Convert string dates to Date objects for the picker
  const selectedRange: DateRange | undefined = {
    from: value.from ? new Date(value.from + 'T00:00:00') : undefined,
    to: value.to ? new Date(value.to + 'T00:00:00') : undefined,
  };

  // Sync pending range when dropdown opens or value changes externally
  useEffect(() => {
    if (isOpen) {
      setPendingRange(selectedRange);
    }
  }, [isOpen]); // eslint-disable-line react-hooks/exhaustive-deps

  // Handle date selection (updates pending state only)
  const handleSelect = (range: DateRange | undefined) => {
    setPendingRange(range);
  };

  // Apply selection when "Done" is clicked
  const handleDone = () => {
    onChange({
      from: pendingRange?.from ? format(pendingRange.from, 'yyyy-MM-dd') : undefined,
      to: pendingRange?.to ? format(pendingRange.to, 'yyyy-MM-dd') : undefined,
    });
    setIsOpen(false);
  };

  // Clear pending selection
  const handleClearPending = () => {
    setPendingRange(undefined);
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

  const handleClear = (e: React.MouseEvent | React.KeyboardEvent) => {
    e.stopPropagation();
    onChange({ from: undefined, to: undefined });
  };

  const handleClearKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      e.stopPropagation();
      onChange({ from: undefined, to: undefined });
    }
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
            <span
              role="button"
              tabIndex={0}
              onClick={handleClear}
              onKeyDown={handleClearKeyDown}
              className="p-1 hover:bg-gray-100 rounded-full transition-colors cursor-pointer inline-flex"
              title="Clear selection"
              aria-label="Clear selection"
            >
              <svg
                className="h-4 w-4 text-gray-400 hover:text-gray-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                aria-hidden
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </span>
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
        <div className="absolute top-full left-0 mt-1 z-50 bg-white border border-gray-200 rounded-lg shadow-lg p-3 rdp-custom">
          <style>{`
            .rdp-custom {
              --rdp-accent-color: #2563eb;
              --rdp-accent-background-color: #dbeafe;
            }
            .rdp-custom .rdp-root {
              color: #111827;
            }
          `}</style>
          <DayPicker
            mode="range"
            selected={pendingRange}
            onSelect={handleSelect}
            numberOfMonths={1}
            captionLayout="dropdown"
            startMonth={new Date(2020, 0)}
            endMonth={new Date(2030, 11)}
          />
          <div className="border-t border-gray-200 pt-2 mt-2 flex justify-end gap-2">
            <button
              type="button"
              onClick={handleClearPending}
              className="px-3 py-1.5 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-md transition-colors"
            >
              Clear
            </button>
            <button
              type="button"
              onClick={handleDone}
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
