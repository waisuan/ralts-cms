import { useState, useEffect, useCallback, useRef } from 'react';
import { MachineService, MachineFilters, MachineListResponse } from '../services/machineService';
import { Machine } from '../types/machine';
import { ApiError, handleApiError } from '../utils/api';
import { isAuthError } from '../utils/auth';

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
  reset: () => void;
}

export function useMachines(options: UseMachinesOptions = {}): UseMachinesReturn {
  const {
    page = 0,
    limit: propLimit = 50,
    filters: propFilters = {},
    autoFetch = true,
  } = options;

  const [machines, setMachines] = useState<Machine[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [limit, setLimit] = useState(propLimit);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<ApiError | null>(null);
  const [overdueCount, setOverdueCount] = useState(0);
  const [dueCount, setDueCount] = useState(0);
  const [almostDueCount, setAlmostDueCount] = useState(0);

  const limitRef = useRef(propLimit);
  const filtersRef = useRef(propFilters);

  useEffect(() => {
    limitRef.current = propLimit;
    setLimit(propLimit);
  }, [propLimit]);

  useEffect(() => {
    filtersRef.current = propFilters;
  }, [propFilters]);

  const fetchMachines = useCallback(async (targetOffset: number) => {
    setLoading(true);
    setError(null);

    const currentLimit = limitRef.current;
    const currentFilters = filtersRef.current;

    try {
      const apiPage = Math.floor(targetOffset / currentLimit) + 1;
      const response = await MachineService.getMachines(apiPage, currentLimit, currentFilters);
      
      const data = response.data as MachineListResponse;
      
      if (!data || !data.machines) {
        console.warn('API returned unexpected format, using empty machines array');
        setMachines([]);
        setTotal(0);
        setOffset(0);
        setOverdueCount(0);
        setDueCount(0);
        setAlmostDueCount(0);
        return;
      }
      
      setMachines(data.machines);
      setTotal(data.count ?? 0);
      setOffset(data.offset);
      setOverdueCount(data.overdue_count || 0);
      setDueCount(data.due_count || 0);
      setAlmostDueCount(data.almost_due_count || 0);
    } catch (err) {
      const apiError = handleApiError(err);
      if (!isAuthError(err)) {
        setError(apiError);
        console.error('Failed to fetch machines:', apiError);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!autoFetch) return;
    fetchMachines(page * limitRef.current);
  }, [fetchMachines, autoFetch, propFilters, page, propLimit]);

  const refetch = useCallback(async () => {
    await fetchMachines(page * limitRef.current);
  }, [fetchMachines, page]);

  const reset = useCallback(() => {
    setMachines([]);
    setTotal(0);
    setOffset(0);
    setLimit(propLimit);
    setLoading(false);
    setError(null);
    setOverdueCount(0);
    setDueCount(0);
    setAlmostDueCount(0);
  }, [propLimit]);

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
    reset,
  };
}
