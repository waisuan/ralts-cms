import { useState, useEffect, useCallback, useRef } from 'react';
import { MachineService, MachineFilters, MachineListResponse } from '../services/machineService';
import { Machine } from '../types/machine';
import { ApiError, handleApiError } from '../utils/api';
import { isAuthError } from '../utils/auth';
import { useDebounce } from './useDebounce';

export interface UseMachinesOptions {
  page?: number;
  limit?: number;
  filters?: MachineFilters;
  autoFetch?: boolean;
}

export interface UseMachinesReturn {
  machines: Machine[];
  total: number;
  offset: number;
  limit: number;
  totalPages: number;
  loading: boolean;
  error: ApiError | null;
  overdueCount: number;
  dueCount: number;
  almostDueCount: number;
  refetch: () => Promise<void>;
  setLimit: (limit: number) => void;
  setFilters: (filters: MachineFilters) => void;
  goToPage: (page: number) => Promise<void>;
  reset: () => void;
}

export function useMachines(options: UseMachinesOptions = {}): UseMachinesReturn {
  const {
    limit: initialLimit = 50,
    filters: initialFilters = {},
    autoFetch = true,
  } = options;

  const [machines, setMachines] = useState<Machine[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [limit, setLimit] = useState(initialLimit);
  const [filters, setFilters] = useState<MachineFilters>(initialFilters);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<ApiError | null>(null);
  const [overdueCount, setOverdueCount] = useState(0);
  const [dueCount, setDueCount] = useState(0);
  const [almostDueCount, setAlmostDueCount] = useState(0);

  // Debounce the search query to avoid excessive API calls
  const debouncedFilters = useDebounce(filters, 300); // 300ms debounce

  // Use refs to avoid stale closures
  const limitRef = useRef(limit);
  const filtersRef = useRef(filters);

  useEffect(() => {
    limitRef.current = limit;
  }, [limit]);

  useEffect(() => {
    filtersRef.current = filters;
  }, [filters]);

  // Update filters when options change
  useEffect(() => {
    if (JSON.stringify(initialFilters) !== JSON.stringify(filters)) {
      setFilters(initialFilters);
    }
  }, [initialFilters, filters]);

  const fetchMachines = useCallback(async (targetOffset?: number) => {
    setLoading(true);
    setError(null);

    const currentOffset = targetOffset !== undefined ? targetOffset : 0;
    const currentLimit = limitRef.current;
    const currentFilters = filtersRef.current;

    try {
      const page = Math.floor(currentOffset / currentLimit) + 1;
      const response = await MachineService.getMachines(page, currentLimit, currentFilters);
      
      const data = response.data as MachineListResponse;
      
      if (!data || !data.machines) {
        console.warn('API returned unexpected format, using empty machines array');
        setMachines([]);
        setTotal(0);
        setOffset(0);
        setLimit(data.limit || currentLimit);
        setOverdueCount(0);
        setDueCount(0);
        setAlmostDueCount(0);
        return;
      }
      
      setMachines(data.machines);
      
      // API count matches the filtered list (including date ranges combined with PPM status).
      setTotal(data.count ?? 0);
      setOffset(data.offset);
      setLimit(data.limit);
      setOverdueCount(data.overdue_count || 0);
      setDueCount(data.due_count || 0);
      setAlmostDueCount(data.almost_due_count || 0);
    } catch (err) {
      const apiError = handleApiError(err);
      // Don't set error if we're redirecting due to auth error
      if (!isAuthError(err)) {
        setError(apiError);
        console.error('Failed to fetch machines:', apiError);
      }
    } finally {
      setLoading(false);
    }
  }, []); // Remove dependencies to avoid stale closures

  // Auto-fetch when dependencies change
  useEffect(() => {
    if (autoFetch) {
      // When filters change, always start fresh from offset 0
      fetchMachines(0);
    }
  }, [fetchMachines, autoFetch, debouncedFilters]);

  const refetch = useCallback(async () => {
    await fetchMachines(0);
  }, [fetchMachines]);

  const goToPage = useCallback(async (page: number) => {
    const targetOffset = page * limitRef.current;
    await fetchMachines(targetOffset);
  }, [fetchMachines]);

  const reset = useCallback(() => {
    setMachines([]);
    setTotal(0);
    setOffset(0);
    setLimit(initialLimit);
    setFilters(initialFilters);
    setLoading(false);
    setError(null);
    setOverdueCount(0);
    setDueCount(0);
    setAlmostDueCount(0);
  }, [initialLimit, initialFilters]);

  const handleSetLimit = useCallback((newLimit: number) => {
    setLimit(newLimit);
    setOffset(0);
    limitRef.current = newLimit;
    fetchMachines(0);
  }, [fetchMachines]);

  const handleSetFilters = useCallback((newFilters: MachineFilters) => {
    setFilters(newFilters);
    // Don't manually clear machines or reset offset - let fetchMachines handle it
  }, []);

  return {
    machines,
    total,
    offset,
    limit,
    totalPages: Math.ceil(total / limit),
    loading,
    error,
    overdueCount,
    dueCount,
    almostDueCount,
    refetch,
    setLimit: handleSetLimit,
    setFilters: handleSetFilters,
    goToPage,
    reset,
  };
} 