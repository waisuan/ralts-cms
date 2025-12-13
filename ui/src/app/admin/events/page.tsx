'use client';

import { useEffect, useState, useCallback } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { 
  AuditService, 
  AuditEvent, 
  AuditFilters,
  AUDIT_ACTION,
  RESOURCE_TYPE,
} from '@/services/auditService';
import { USER_ROLE } from '@/services/adminUserService';
import LoadingSpinner from '@/components/LoadingSpinner';
import EventList from '@/components/admin/EventList';
import DateRangePicker, { DateRangeValue } from '@/components/DateRangePicker';

export default function AdminEventsPage() {
  const { user, isLoading: authLoading } = useAuth();
  const router = useRouter();
  const [isPageLoading, setIsPageLoading] = useState(true);
  
  // Event list state
  const [events, setEvents] = useState<AuditEvent[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(25);
  const [isLoadingEvents, setIsLoadingEvents] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Filter state
  const [dateRange, setDateRange] = useState<DateRangeValue>({ from: undefined, to: undefined });
  const [resourceType, setResourceType] = useState<string>('');
  const [action, setAction] = useState<string>('');
  const [showFilters, setShowFilters] = useState(false);

  // Build filters object
  const buildFilters = useCallback((): AuditFilters => {
    const filters: AuditFilters = {};
    
    if (dateRange.from) {
      // Convert YYYY-MM-DD to RFC3339 format for API
      filters.from_date = new Date(dateRange.from + 'T00:00:00').toISOString();
    }
    if (dateRange.to) {
      // End of day for the 'to' date
      filters.to_date = new Date(dateRange.to + 'T23:59:59').toISOString();
    }
    if (resourceType) {
      filters.resource_type = resourceType;
    }
    if (action) {
      filters.action = action;
    }
    
    return filters;
  }, [dateRange, resourceType, action]);

  // Load events data
  const loadEvents = useCallback(async (page: number = currentPage, overrideFilters?: AuditFilters) => {
    setIsLoadingEvents(true);
    setError(null);
    
    try {
      const offset = (page - 1) * pageSize;
      // Use override filters if provided, otherwise build from state
      const filters = overrideFilters !== undefined ? overrideFilters : buildFilters();
      const response = await AuditService.getEvents(pageSize, offset, filters);
      
      if (response.data) {
        setEvents(response.data.events || []);
        setTotalCount(response.data.total_count);
        setCurrentPage(page);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load events');
      console.error('Failed to load events:', err);
    } finally {
      setIsLoadingEvents(false);
    }
  }, [currentPage, pageSize, buildFilters]);

  // Handle page change
  const handlePageChange = (page: number) => {
    loadEvents(page);
  };

  // Handle filter change
  const handleApplyFilters = () => {
    setCurrentPage(1);
    loadEvents(1);
  };

  // Clear all filters
  const handleClearFilters = () => {
    setDateRange({ from: undefined, to: undefined });
    setResourceType('');
    setAction('');
    setCurrentPage(1);
    // Load with empty filters immediately
    loadEvents(1, {});
  };

  // Check if any filters are active
  const hasActiveFilters = dateRange.from || dateRange.to || resourceType || action;

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
    
    // Load initial events data
    loadEvents(1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user, authLoading, router]);

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
                  Event Log
                </span>
              </div>
            </li>
          </ol>
        </nav>

        {/* Page Title and Description */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Event Log</h1>
            <p className="mt-2 text-sm text-gray-600">
              View and search audit events across the system.
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <Link
              href="/admin/users"
              className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            >
              <svg className="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
              User Management
            </Link>
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

      {/* Filters Section */}
      <div className="mb-6 bg-white shadow-sm rounded-lg border border-gray-200">
        <div className="p-4">
          <div className="flex items-center justify-between">
            <button
              onClick={() => setShowFilters(!showFilters)}
              className="inline-flex items-center text-sm font-medium text-gray-700 hover:text-gray-900"
            >
              <svg className={`w-5 h-5 mr-2 transition-transform ${showFilters ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
              Filters
              {hasActiveFilters && (
                <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  Active
                </span>
              )}
            </button>
            
            {hasActiveFilters && (
              <button
                onClick={handleClearFilters}
                className="text-sm text-gray-500 hover:text-gray-700"
              >
                Clear all
              </button>
            )}
          </div>
          
          {showFilters && (
            <div className="mt-4 pt-4 border-t border-gray-200">
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                {/* Date Range */}
                <div>
                  <DateRangePicker
                    label="Date Range"
                    value={dateRange}
                    onChange={setDateRange}
                    placeholder="Select dates..."
                  />
                </div>
                
                {/* Resource Type */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Resource Type
                  </label>
                  <select
                    value={resourceType}
                    onChange={(e) => setResourceType(e.target.value)}
                    className="block w-full h-10 px-3 py-2 text-gray-900 bg-white border border-gray-300 rounded-md shadow-sm focus:border-blue-500 focus:ring-blue-500 focus:outline-none sm:text-sm"
                  >
                    <option value="">All Resources</option>
                    <option value={RESOURCE_TYPE.MACHINE}>Machine</option>
                    <option value={RESOURCE_TYPE.MAINTENANCE}>Maintenance</option>
                    <option value={RESOURCE_TYPE.USER}>User</option>
                    <option value={RESOURCE_TYPE.SESSION}>Session</option>
                    <option value={RESOURCE_TYPE.ATTACHMENT}>Attachment</option>
                  </select>
                </div>
                
                {/* Action Type */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Action
                  </label>
                  <select
                    value={action}
                    onChange={(e) => setAction(e.target.value)}
                    className="block w-full h-10 px-3 py-2 text-gray-900 bg-white border border-gray-300 rounded-md shadow-sm focus:border-blue-500 focus:ring-blue-500 focus:outline-none sm:text-sm"
                  >
                    <option value="">All Actions</option>
                    <option value={AUDIT_ACTION.CREATED}>Created</option>
                    <option value={AUDIT_ACTION.UPDATED}>Updated</option>
                    <option value={AUDIT_ACTION.DELETED}>Deleted</option>
                    <option value={AUDIT_ACTION.VIEWED}>Viewed</option>
                    <option value={AUDIT_ACTION.LISTED}>Listed</option>
                    <option value={AUDIT_ACTION.LOGIN}>Login</option>
                    <option value={AUDIT_ACTION.LOGOUT}>Logout</option>
                    <option value={AUDIT_ACTION.PASSWORD_CHANGED}>Password Changed</option>
                  </select>
                </div>
                
                {/* Apply Button */}
                <div className="flex items-end">
                  <button
                    onClick={handleApplyFilters}
                    disabled={isLoadingEvents}
                    className="w-full inline-flex justify-center items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                    Apply Filters
                  </button>
                </div>
              </div>
              
              {/* Active Filter Tags */}
              {hasActiveFilters && (
                <div className="mt-4 flex flex-wrap gap-2">
                  {dateRange.from && (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                      From: {dateRange.from}
                      <button
                        onClick={() => setDateRange(prev => ({ ...prev, from: undefined }))}
                        className="ml-1 text-gray-500 hover:text-gray-700"
                      >
                        ×
                      </button>
                    </span>
                  )}
                  {dateRange.to && (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                      To: {dateRange.to}
                      <button
                        onClick={() => setDateRange(prev => ({ ...prev, to: undefined }))}
                        className="ml-1 text-gray-500 hover:text-gray-700"
                      >
                        ×
                      </button>
                    </span>
                  )}
                  {resourceType && (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                      Resource: {AuditService.getResourceTypeLabel(resourceType)}
                      <button
                        onClick={() => setResourceType('')}
                        className="ml-1 text-gray-500 hover:text-gray-700"
                      >
                        ×
                      </button>
                    </span>
                  )}
                  {action && (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                      Action: {AuditService.getActionLabel(action)}
                      <button
                        onClick={() => setAction('')}
                        className="ml-1 text-gray-500 hover:text-gray-700"
                      >
                        ×
                      </button>
                    </span>
                  )}
                </div>
              )}
            </div>
          )}
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

      {/* Event List */}
      <EventList
        events={events}
        totalCount={totalCount}
        currentPage={currentPage}
        pageSize={pageSize}
        onPageChange={handlePageChange}
        isLoading={isLoadingEvents}
        error={error}
      />
    </div>
  );
}

