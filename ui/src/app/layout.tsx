import type { Metadata } from 'next';
import { Geist, Geist_Mono } from 'next/font/google';
import { NuqsAdapter } from 'nuqs/adapters/next/app';
import './globals.css';
import { AuthProvider } from '@/contexts/AuthContext';
import AppContent from '@/components/AppContent';
import PageTransitionLoader from '@/components/PageTransitionLoader';

const geistSans = Geist({
  subsets: ['latin'],
  variable: '--font-geist-sans',
});

const geistMono = Geist_Mono({
  subsets: ['latin'],
  variable: '--font-geist-mono',
});

export const metadata: Metadata = {
  title: 'Ralts CMS',
  description: 'Content Management System',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        <NuqsAdapter>
          <AuthProvider>
            <PageTransitionLoader />
            <AppContent>{children}</AppContent>
          </AuthProvider>
        </NuqsAdapter>
      </body>
    </html>
  );
}
