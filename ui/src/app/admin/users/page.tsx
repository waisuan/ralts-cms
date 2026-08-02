'use client';

import { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { AdminUserService, USER_ROLE, USER_STATUS, UserStatus, isDestructiveStatus } from '@/services/adminUserService';
import { User } from '@/services/userService';
import LoadingSpinner from '@/components/LoadingSpinner';
import UserList from '@/components/admin/UserList';

export default function AdminUsersPage() {
  const { user, isLoading: authLoading } = useAuth();
  const router = useRouter();
  const [isPageLoading, setIsPageLoading] = useState(true);
  
  // User management state
  const [users, setUsers] = useState<User[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(25);
  const [isLoadingUsers, setIsLoadingUsers] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load users data
  const loadUsers = async (page: number = currentPage) => {
    setIsLoadingUsers(true);
    setError(null);
    
    try {
      const offset = (page - 1) * pageSize;
      const response = await AdminUserService.getUsers(pageSize, offset);
      
      if (response.data) {
        setUsers(response.data.users);
        setTotalCount(response.data.total_count);
        setCurrentPage(page);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load users');
      console.error('Failed to load users:', err);
    } finally {
      setIsLoadingUsers(false);
    }
  };

  // Handle bulk status update
  const handleBulkStatusUpdate = async (userIds: number[], status: UserStatus) => {
    try {
      await AdminUserService.updateMultipleUserStatuses(userIds, status);

      if (isDestructiveStatus(status)) {
        // Rejected users are permanently deleted server-side; remove them locally too
        setUsers(prevUsers => prevUsers.filter(u => !userIds.includes(u.id)));
        setTotalCount(prevCount => Math.max(0, prevCount - userIds.length));
        return;
      }

      // Update local state optimistically
      setUsers(prevUsers => 
        prevUsers.map(u => 
          userIds.includes(u.id) ? { ...u, status, approved: status === USER_STATUS.APPROVED } : u
        )
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update user statuses');
      throw err; // Re-throw to let component handle loading states
    }
  };

  // Handle page change
  const handlePageChange = (page: number) => {
    loadUsers(page);
  };

  useEffect(() => {
    // Wait for auth to load
    if (authLoading) return;

    // Redirect non-authenticated users
    if (!user) {
      router.push('/');
      return;
    }

    // Redirect non-admin users
    if (user.role !== USER_ROLE.ADMIN) {
      router.push('/');
      return;
    }

    // User is authenticated and is admin
    setIsPageLoading(false);
    
    // Load initial users data
    loadUsers(1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user, authLoading, router]);

  // Reload data when page loads
  useEffect(() => {
    if (!isPageLoading && user?.role === USER_ROLE.ADMIN) {
      loadUsers(currentPage);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isPageLoading]);

  // Show loading spinner while checking authentication and authorization
  if (authLoading || isPageLoading) {
    return <LoadingSpinner />;
  }

  // This should not render if user is not admin due to redirect above
  if (!user || user.role !== USER_ROLE.ADMIN) {
    return null;
  }

  return (
    <div className="container mx-auto px-4 py-6">
      {/* Page Header with Breadcrumbs */}
      <div className="mb-8">
        <nav className="flex mb-4" aria-label="Breadcrumb">
          <ol className="inline-flex items-center space-x-1 md:space-x-3">
            <li className="inline-flex items-center">
              <Link
                href="/"
                className="inline-flex items-center text-sm font-medium text-gray-700 hover:text-blue-600"
              >
                <svg
                  className="w-3 h-3 mr-2.5"
                  aria-hidden="true"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path d="m19.707 9.293-2-2-7-7a1 1 0 0 0-1.414 0l-7 7-2 2a1 1 0 0 0 1.414 1.414L2 10.414V18a2 2 0 0 0 2 2h3a1 1 0 0 0 1-1v-4a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v4a1 1 0 0 0 1 1h3a2 2 0 0 0 2-2v-7.586l.293.293a1 1 0 0 0 1.414-1.414Z"/>
                </svg>
                Home
              </Link>
            </li>
            <li>
              <div className="flex items-center">
                <svg
                  className="w-3 h-3 text-gray-400 mx-1"
                  aria-hidden="true"
                  fill="none"
                  viewBox="0 0 6 10"
                >
                  <path
                    stroke="currentColor"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="m1 9 4-4-4-4"
                  />
                </svg>
                <span className="ml-1 text-sm font-medium text-gray-500 md:ml-2">
                  Admin
                </span>
              </div>
            </li>
            <li aria-current="page">
              <div className="flex items-center">
                <svg
                  className="w-3 h-3 text-gray-400 mx-1"
                  aria-hidden="true"
                  fill="none"
                  viewBox="0 0 6 10"
                >
                  <path
                    stroke="currentColor"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="m1 9 4-4-4-4"
                  />
                </svg>
                <span className="ml-1 text-sm font-medium text-gray-500 md:ml-2">
                  User Management
                </span>
              </div>
            </li>
          </ol>
        </nav>

        {/* Page Title and Description */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">User Management</h1>
            <p className="mt-2 text-sm text-gray-600">
              Manage user accounts, statuses, and permissions across the system.
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <div className="flex items-center px-3 py-1 bg-blue-100 text-blue-800 text-xs font-medium rounded-full">
              <svg
                className="w-3 h-3 mr-1"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path
                  fillRule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-6-3a2 2 0 11-4 0 2 2 0 014 0zm-2 4a5 5 0 00-4.546 2.916A5.986 5.986 0 0010 16a5.986 5.986 0 004.546-2.084A5 5 0 0010 11z"
                  clipRule="evenodd"
                />
              </svg>
              Administrator
            </div>
          </div>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div className="mb-6 bg-red-50 border border-red-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error</h3>
              <div className="mt-2 text-sm text-red-700">{error}</div>
            </div>
            <div className="ml-auto pl-3">
              <div className="-mx-1.5 -my-1.5">
                <button
                  type="button"
                  onClick={() => setError(null)}
                  className="inline-flex bg-red-50 rounded-md p-1.5 text-red-500 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-red-50 focus:ring-red-600"
                >
                  <span className="sr-only">Dismiss</span>
                  <svg className="h-3 w-3" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* User Management Interface */}
      <UserList
        users={users}
        totalCount={totalCount}
        currentPage={currentPage}
        pageSize={pageSize}
        onPageChange={handlePageChange}
        onBulkStatusUpdate={handleBulkStatusUpdate}
        isLoading={isLoadingUsers}
        error={error}
      />
    </div>
  );
}
