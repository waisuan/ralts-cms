'use client';

import { useEffect, useRef, useState } from 'react';

interface DebouncedSearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  isLoading?: boolean;
  debounceMs?: number;
  showClear?: boolean;
  infoText?: React.ReactNode;
  ariaLabel?: string;
}

const DEFAULT_DEBOUNCE_MS = 300;

export default function DebouncedSearchInput({
  value,
  onChange,
  placeholder = 'Search...',
  isLoading = false,
  debounceMs = DEFAULT_DEBOUNCE_MS,
  showClear = true,
  infoText,
  ariaLabel,
}: DebouncedSearchInputProps) {
  const [localQuery, setLocalQuery] = useState(value);
  const committedRef = useRef(value);
  const onChangeRef = useRef(onChange);

  useEffect(() => {
    onChangeRef.current = onChange;
  }, [onChange]);

  // If the parent overrides `value` out-of-band (URL hydration, clear,
  // programmatic reset), drop the pending debounced commit by resyncing
  // localQuery *and* committedRef to the new value.
  useEffect(() => {
    if (value !== committedRef.current) {
      committedRef.current = value;
      setLocalQuery(value);
    }
  }, [value]);

  useEffect(() => {
    if (localQuery === committedRef.current) return;
    const timer = setTimeout(() => {
      committedRef.current = localQuery;
      onChangeRef.current(localQuery);
    }, debounceMs);
    return () => clearTimeout(timer);
  }, [localQuery, debounceMs]);

  const commitImmediately = (next: string) => {
    committedRef.current = next;
    setLocalQuery(next);
    onChangeRef.current(next);
  };

  return (
    <div className="w-full relative">
      <div className="flex items-center bg-white border border-gray-300 rounded-lg shadow-sm focus-within:ring-2 focus-within:ring-blue-500 focus-within:border-transparent">
        <div className="flex-1 relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            {isLoading ? (
              <svg
                className="h-5 w-5 text-gray-400 animate-spin"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
            ) : (
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
                  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                />
              </svg>
            )}
          </div>
          <input
            type="text"
            aria-label={ariaLabel ?? placeholder}
            placeholder={placeholder}
            value={localQuery}
            onChange={(e) => setLocalQuery(e.target.value)}
            className="w-full pl-10 pr-10 py-2 border-0 rounded-lg focus:outline-none placeholder-gray-400 text-gray-900"
          />
          {showClear && localQuery && !isLoading && (
            <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
              <button
                type="button"
                onClick={() => commitImmediately('')}
                className="text-gray-400 hover:text-gray-600 focus:outline-none focus:text-gray-600 transition-colors"
                title="Clear search"
                aria-label="Clear search"
              >
                <svg
                  className="h-5 w-5"
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
            </div>
          )}
        </div>
      </div>
      {infoText && (
        <div className="mt-2 text-sm text-gray-500 text-center">{infoText}</div>
      )}
    </div>
  );
}
