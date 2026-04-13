'use client';

import Link from 'next/link';
import UserMenu from './UserMenu';

export default function Header() {
  return (
    <header className="sticky top-0 z-50 bg-white shadow-sm border-b border-gray-200">
      <div className="container mx-auto px-4">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center">
            <Link
              href="/"
              className="text-xl font-semibold text-gray-900 hover:text-blue-600 transition-colors"
            >
              Ralts CMS
            </Link>
            <span className="ml-2 text-sm text-gray-500 hidden sm:inline">Content Management System</span>
          </div>

          <div className="flex items-center space-x-4">
            <UserMenu />
          </div>
        </div>
      </div>
    </header>
  );
}
