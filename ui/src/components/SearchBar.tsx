'use client';

import { useState } from 'react';
import { SEARCH_PROPERTIES } from '@/utils/constants';

export interface SearchOptions {
  query: string;
  property: string;
}

interface SearchBarProps {
  onSearch: (options: SearchOptions) => void;
  searchOptions: SearchOptions;
}

export default function SearchBar({ onSearch, searchOptions }: SearchBarProps) {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);

  const handleQueryChange = (query: string) => {
    onSearch({ ...searchOptions, query });
  };

  const handlePropertyChange = (property: string) => {
    onSearch({ ...searchOptions, property });
    setIsDropdownOpen(false);
  };

  const currentProperty = SEARCH_PROPERTIES.find((prop) => prop.value === searchOptions.property);

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
            <input
              type="text"
              placeholder={`Search by ${currentProperty?.label.toLowerCase()}...`}
              value={searchOptions.query}
              onChange={(e) => handleQueryChange(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border-0 rounded-r-lg focus:outline-none placeholder-gray-400 text-gray-900"
            />
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
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
            </div>
          </div>
        </div>

        {/* Search Info */}
        {searchOptions.query && (
          <div className="mt-2 text-sm text-gray-500 text-center">
            Searching for &ldquo;{searchOptions.query}&rdquo; in{' '}
            <span className="font-medium">{currentProperty?.label}</span>
          </div>
        )}
      </div>
    </div>
  );
}
