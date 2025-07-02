'use client';

import { ReactNode } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { usePathname } from 'next/navigation';
import Header from './Header';
import LoginPage from './LoginPage';
import RegisterPage from './RegisterPage';
import LoadingSpinner from './LoadingSpinner';

export default function AppContent({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const pathname = usePathname();

  // Show loading spinner while checking authentication
  if (isLoading) {
    return <LoadingSpinner />;
  }

  // Show registration page if on /register route
  if (pathname === '/register') {
    return <RegisterPage />;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {isAuthenticated ? (
        <>
          <Header />
          <main className="py-8">{children}</main>
        </>
      ) : (
        <LoginPage />
      )}
    </div>
  );
}
