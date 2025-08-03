'use client';

import { useState } from 'react';
import {
  SEARCH_PROPERTIES,
  SEARCHABLE_PPM_STATUSES,
  isDateProperty,
  isPPMStatusProperty,
} from '@/utils/constants';

export interface SearchOptions {
  query: string;
  property: string;
}

interface SearchBarProps {
  onSearch: (options: SearchOptions) => void;
  searchOptions: SearchOptions;
  isSearching?: boolean;
}

// Helper function to convert internal PPM status to consumer-facing text
function getPPMStatusDisplayText(internalValue: string): string {
  const statusOption = SEARCHABLE_PPM_STATUSES.find(status => status.value === internalValue);
  return statusOption ? statusOption.label : internalValue;
}

export default function SearchBar({ onSearch, searchOptions, isSearching = false }: SearchBarProps) {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);

  const handleQueryChange = (query: string) => {
    onSearch({ ...searchOptions, query });
  };

  const handlePropertyChange = (property: string) => {
    // Clear the query when switching to a different property type
    onSearch({ query: '', property });
    setIsDropdownOpen(false);
  };

  const currentProperty = SEARCH_PROPERTIES.find((prop) => prop.value === searchOptions.property);
  const isDatePropertyValue = isDateProperty(searchOptions.property);
  const isPPMStatusPropertyValue = isPPMStatusProperty(searchOptions.property);

  return (
    <div className="flex justify-center mb-6">
      <div className="w-full max-w-2xl relative">
        <div className="flex items-center bg-white border border-gray-300 rounded-lg shadow-sm focus-within:ring-2 focus-within:ring-blue-500 focus-within:border-transparent">
          {/* Search Property Dropdown */}
          <div className="relative">
            <button
              type="button"
              onClick={() => setIsDropdownOpen(!isDropdownOpen)}
              className="flex items-center px-4 py-2 bg-gray-50 border-r border-gray-300 rounded-l-lg hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
            >
              <span className="text-sm font-medium text-gray-700 min-w-max">
                {currentProperty?.label}
              </span>
              <svg
                className="ml-2 h-4 w-4 text-gray-500"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M19 9l-7 7-7-7"
                />
              </svg>
            </button>

            {/* Dropdown Menu */}
            {isDropdownOpen && (
              <>
                {/* Backdrop */}
                <div className="fixed inset-0 z-10" onClick={() => setIsDropdownOpen(false)} />

                {/* Dropdown Content */}
                <div className="absolute top-full left-0 mt-1 w-48 bg-white border border-gray-300 rounded-md shadow-lg z-20 max-h-60 overflow-y-auto">
                  {SEARCH_PROPERTIES.map((property) => (
                    <button
                      key={property.value}
                      onClick={() => handlePropertyChange(property.value)}
                      className={`w-full text-left px-4 py-2 text-sm hover:bg-gray-50 focus:bg-gray-50 focus:outline-none transition-colors ${
                        searchOptions.property === property.value
                          ? 'bg-blue-50 text-blue-600 font-medium'
                          : 'text-gray-700'
                      }`}
                    >
                      {property.label}
                    </button>
                  ))}
                </div>
              </>
            )}
          </div>

          {/* Search Input */}
          <div className="flex-1 relative">
            {isDatePropertyValue ? (
              /* Date Picker Input */
              <input
                type="date"
                value={searchOptions.query}
                onChange={(e) => handleQueryChange(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border-0 rounded-r-lg focus:outline-none text-gray-900"
              />
            ) : isPPMStatusPropertyValue ? (
              /* PPM Status Dropdown */
              <select
                value={searchOptions.query}
                onChange={(e) => handleQueryChange(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border-0 rounded-r-lg focus:outline-none text-gray-900 bg-white appearance-none"
              >
                <option value="">Select PPM Status...</option>
                {SEARCHABLE_PPM_STATUSES.map((status) => (
                  <option key={status.value} value={status.value}>
                    {status.label}
                  </option>
                ))}
              </select>
            ) : (
              /* Text Input */
              <input
                type="text"
                placeholder={`Search by ${currentProperty?.label.toLowerCase()}...`}
                value={searchOptions.query}
                onChange={(e) => handleQueryChange(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border-0 rounded-r-lg focus:outline-none placeholder-gray-400 text-gray-900"
              />
            )}
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              {isSearching ? (
                /* Loading Spinner */
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
              ) : isDatePropertyValue ? (
                /* Calendar Icon for Date Properties */
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
              ) : isPPMStatusPropertyValue ? (
                /* List Icon for PPM Status Property */
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
                    d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"
                  />
                </svg>
              ) : (
                /* Search Icon for Text Properties */
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
            
            {/* Clear Button - Only show when there's a query and not loading */}
            {searchOptions.query && !isSearching && (
              <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
                <button
                  type="button"
                  onClick={() => handleQueryChange('')}
                  className="text-gray-400 hover:text-gray-600 focus:outline-none focus:text-gray-600 transition-colors"
                  title="Clear search"
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

        {/* Search Info */}
        {searchOptions.query && (
          <div className="mt-2 text-sm text-gray-500 text-center">
            {isDatePropertyValue ? (
              <>
                Filtering by <span className="font-medium">{currentProperty?.label}</span> on{' '}
                <span className="font-medium">{searchOptions.query}</span>
              </>
            ) : isPPMStatusPropertyValue ? (
              <>
                Filtering by <span className="font-medium">{currentProperty?.label}</span>:{' '}
                <span className="font-medium">{getPPMStatusDisplayText(searchOptions.query)}</span>
              </>
            ) : (
              <>
                Searching for &ldquo;{searchOptions.query}&rdquo; in{' '}
                <span className="font-medium">{currentProperty?.label}</span>
              </>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
