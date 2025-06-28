'use client';

import { useState } from 'react';
import RecordsList from '@/components/RecordsList';

export default function Home() {
  const [searchQuery, setSearchQuery] = useState('');

  return (
    <main className="min-h-screen bg-gray-50 py-8">
      <div className="container mx-auto px-4">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Ralts CMS</h1>
          <p className="text-gray-600 mb-6">Content Management System</p>
          
          {/* Centered Search Bar */}
          <div className="flex justify-center mb-6">
            <div className="w-full max-w-md relative">
              <input
                type="text"
                placeholder="Search records..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent placeholder-gray-400"
              />
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </div>
            </div>
          </div>
        </div>
        
        <RecordsList searchQuery={searchQuery} />
      </div>
    </main>
  );
}
