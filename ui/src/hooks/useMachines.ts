import { useState, useEffect, useCallback } from 'react';
import { MachineService, MachineFilters, MachineListResponse } from '../services/machineService';
import { Machine } from '../types/machine';
import { ApiError, handleApiError } from '../utils/api';
import { logApiResponse, logApiError } from '../utils/debug';

export interface UseMachinesOptions {
  page?: number;
  limit?: number;
  filters?: MachineFilters;
  autoFetch?: boolean;
}

export interface UseMachinesReturn {
  machines: Machine[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
  loading: boolean;
  error: ApiError | null;
  refetch: () => Promise<void>;
  setPage: (page: number) => void;
  setLimit: (limit: number) => void;
  setFilters: (filters: MachineFilters) => void;
}

export function useMachines(options: UseMachinesOptions = {}): UseMachinesReturn {
  const {
    page: initialPage = 1,
    limit: initialLimit = 10,
    filters: initialFilters = {},
    autoFetch = true,
  } = options;

  const [machines, setMachines] = useState<Machine[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(initialPage);
  const [limit, setLimit] = useState(initialLimit);
  const [filters, setFilters] = useState<MachineFilters>(initialFilters);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<ApiError | null>(null);

  const fetchMachines = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await MachineService.getMachines(page, limit, filters);
      logApiResponse('/api/v1/machines', response);
      
      const data = response.data as MachineListResponse;
      
      if (!data || !data.machines) {
        console.warn('API returned unexpected format, using empty machines array');
        setMachines([]);
        setTotal(0);
        setPage(1);
        setLimit(limit);
        return;
      }
      
      setMachines(data.machines);
      setTotal(data.count);
      // Calculate current page from offset and limit
      const currentPage = Math.floor((data.offset || 0) / limit) + 1;
      setPage(currentPage);
      setLimit(data.limit);
    } catch (err) {
      const apiError = handleApiError(err);
      logApiError('/api/v1/machines', apiError);
      setError(apiError);
      console.error('Failed to fetch machines:', apiError);
    } finally {
      setLoading(false);
    }
  }, [page, limit, filters]);

  // Auto-fetch when dependencies change
  useEffect(() => {
    if (autoFetch) {
      fetchMachines();
    }
  }, [fetchMachines, autoFetch]);

  const refetch = useCallback(async () => {
    await fetchMachines();
  }, [fetchMachines]);

  const handleSetPage = useCallback((newPage: number) => {
    setPage(newPage);
  }, []);

  const handleSetLimit = useCallback((newLimit: number) => {
    setLimit(newLimit);
    setPage(1); // Reset to first page when changing limit
  }, []);

  const handleSetFilters = useCallback((newFilters: MachineFilters) => {
    setFilters(newFilters);
    setPage(1); // Reset to first page when changing filters
  }, []);

  return {
    machines,
    total,
    page,
    limit,
    totalPages: Math.ceil(total / limit),
    loading,
    error,
    refetch,
    setPage: handleSetPage,
    setLimit: handleSetLimit,
    setFilters: handleSetFilters,
  };
} 