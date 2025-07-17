import { useState, useEffect, useCallback, useRef } from 'react';
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
  offset: number;
  limit: number;
  totalPages: number;
  loading: boolean;
  error: ApiError | null;
  refetch: () => Promise<void>;
  setLimit: (limit: number) => void;
  setFilters: (filters: MachineFilters) => void;
  loadMore: () => Promise<void>; // Load next page and append to existing machines
  reset: () => void; // Reset to initial state
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
  const [offset, setOffset] = useState(0);
  const [limit, setLimit] = useState(initialLimit);
  const [filters, setFilters] = useState<MachineFilters>(initialFilters);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<ApiError | null>(null);

  // Use refs to avoid stale closures
  const offsetRef = useRef(offset);
  const limitRef = useRef(limit);
  const filtersRef = useRef(filters);

  // Update refs when state changes
  useEffect(() => {
    offsetRef.current = offset;
  }, [offset]);

  useEffect(() => {
    limitRef.current = limit;
  }, [limit]);

  useEffect(() => {
    filtersRef.current = filters;
  }, [filters]);

  const fetchMachines = useCallback(async (targetOffset?: number, shouldAppend = false) => {
    setLoading(true);
    setError(null);

    const currentOffset = targetOffset !== undefined ? targetOffset : offsetRef.current;
    const currentLimit = limitRef.current;
    const currentFilters = filtersRef.current;

    console.log('fetchMachines called:', { 
      targetOffset, 
      currentOffset, 
      currentLimit, 
      shouldAppend,
      filters: currentFilters 
    });

    try {
      // Calculate page from offset for the API call
      const page = Math.floor(currentOffset / currentLimit) + 1;
      console.log('Making API call with:', { page, currentLimit, currentFilters });
      const response = await MachineService.getMachines(page, currentLimit, currentFilters);
      logApiResponse('/api/v1/machines', response);
      
      const data = response.data as MachineListResponse;
      
      console.log('API response:', { 
        machinesCount: data.machines?.length, 
        total: data.count, 
        offset: data.offset, 
        limit: data.limit 
      });
      
      if (!data || !data.machines) {
        console.warn('API returned unexpected format, using empty machines array');
        if (!shouldAppend) {
          setMachines([]);
        }
        setTotal(0);
        setOffset(0);
        setLimit(data.limit || currentLimit);
        return;
      }
      
      if (shouldAppend) {
        // Append new machines to existing ones
        setMachines(prev => [...prev, ...data.machines]);
      } else {
        // Replace machines array
        setMachines(data.machines);
      }
      
      setTotal(data.count);
      setOffset(data.offset);
      setLimit(data.limit);
    } catch (err) {
      const apiError = handleApiError(err);
      logApiError('/api/v1/machines', apiError);
      setError(apiError);
      console.error('Failed to fetch machines:', apiError);
    } finally {
      setLoading(false);
    }
  }, []); // Remove dependencies to avoid stale closures

  // Auto-fetch when dependencies change
  useEffect(() => {
    if (autoFetch) {
      fetchMachines();
    }
  }, [fetchMachines, autoFetch]);

  const refetch = useCallback(async () => {
    await fetchMachines(0, false); // Reset to offset 0 and replace machines
  }, [fetchMachines]);

  const loadMore = useCallback(async () => {
    const currentOffset = offsetRef.current;
    const currentLimit = limitRef.current;
    const currentTotal = total;
    
    const nextOffset = currentOffset + currentLimit;
    console.log('LoadMore called:', { currentOffset, currentLimit, nextOffset, currentTotal });
    if (nextOffset < currentTotal) {
      await fetchMachines(nextOffset, true); // Load next page and append
    }
  }, [fetchMachines, total]);

  const reset = useCallback(() => {
    setMachines([]);
    setTotal(0);
    setOffset(0);
    setLimit(initialLimit);
    setFilters(initialFilters);
    setLoading(false);
    setError(null);
  }, [initialLimit, initialFilters]);

  const handleSetLimit = useCallback((newLimit: number) => {
    setLimit(newLimit);
    setOffset(0); // Reset to first page when changing limit
    setMachines([]); // Clear machines when changing limit
  }, []);

  const handleSetFilters = useCallback((newFilters: MachineFilters) => {
    setFilters(newFilters);
    setOffset(0); // Reset to first page when changing filters
    setMachines([]); // Clear machines when changing filters
  }, []);

  return {
    machines,
    total,
    offset,
    limit,
    totalPages: Math.ceil(total / limit),
    loading,
    error,
    refetch,
    setLimit: handleSetLimit,
    setFilters: handleSetFilters,
    loadMore,
    reset,
  };
} 