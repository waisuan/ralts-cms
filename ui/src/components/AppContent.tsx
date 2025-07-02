'use client';

import { ReactNode } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Header from './Header';
import LoginPage from './LoginPage';
import LoadingSpinner from './LoadingSpinner';

export default function AppContent({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();

  // Show loading spinner while checking authentication
  if (isLoading) {
    return <LoadingSpinner />;
  }

  // Show login page if not authenticated
  if (!isAuthenticated) {
    return <LoginPage />;
  }

  // Show main app with header if authenticated
  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <main className="py-8">
        {children}
      </main>
    </div>
  );
} 