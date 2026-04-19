'use client';

import { ReactNode } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { usePathname } from 'next/navigation';
import Header from './Header';
import AnnouncementBanner from './AnnouncementBanner';
import LoginPage from './LoginPage';
import RegisterPage from './RegisterPage';
import LoadingSpinner from './LoadingSpinner';
import AppFooter from './AppFooter';

export default function AppContent({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const pathname = usePathname();

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (pathname === '/register') {
    return <RegisterPage />;
  }

  return (
    <div className="flex min-h-screen flex-col bg-gray-50">
      {isAuthenticated ? (
        <>
          <AnnouncementBanner />
          <Header />
          <main className="flex-1 py-8">{children}</main>
          <AppFooter />
        </>
      ) : (
        <LoginPage />
      )}
    </div>
  );
}
