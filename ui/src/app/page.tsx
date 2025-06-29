'use client';

import { useState } from 'react';
import RecordsList from '@/components/RecordsList';
import SearchBar, { SearchOptions } from '@/components/SearchBar';
import { DEFAULT_SEARCH_PROPERTY } from '@/utils/constants';

export default function Home() {
  const [searchOptions, setSearchOptions] = useState<SearchOptions>({
    query: '',
    property: DEFAULT_SEARCH_PROPERTY,
  });

  return (
    <main className="min-h-screen bg-gray-50 py-8">
      <div className="container mx-auto px-4">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Ralts CMS</h1>
          <p className="text-gray-600 mb-6">Content Management System</p>

          {/* Enhanced Search Bar with Dropdown */}
          <SearchBar searchOptions={searchOptions} onSearch={setSearchOptions} />
        </div>

        <RecordsList searchOptions={searchOptions} />
      </div>
    </main>
  );
}
