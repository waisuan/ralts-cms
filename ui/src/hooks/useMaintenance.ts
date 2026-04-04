import { useState, useEffect, useCallback, useRef } from 'react';
import { Maintenance } from '../types/maintenance';
import { MaintenanceService, MaintenanceFilters } from '../services/maintenanceService';
import { handleApiError, ApiError } from '../utils/api';
import { isAuthError } from '../utils/auth';
import { useDebounce } from './useDebounce';

export interface UseMaintenanceOptions {
  serialNumber: string;
  limit?: number;
  autoFetch?: boolean;
}

export interface UseMaintenanceReturn {
  records: Maintenance[];
  total: number;
  currentPage: number;
  limit: number;
  totalPages: number;
  sort: string;
  searchQuery: string;
  debouncedSearchQuery: string;

  isInitialLoading: boolean;
  isSearchLoading: boolean;
  isPaginationLoading: boolean;
  error: string | null;

  preventativeCount: number;
  correctiveCount: number;
  emergencyCount: number;
  inspectionCount: number;
  otherCount: number;

  setSearchQuery: (query: string) => void;
  clearSearch: () => void;
  hasActiveSearch: boolean;
  setSort: (sort: string) => void;
  goToPage: (page: number) => void;
  setLimit: (limit: number) => void;
  refetch: () => void;
}

const DEFAULT_LIMIT = 10;

export function useMaintenance(options: UseMaintenanceOptions): UseMaintenanceReturn {
  const { serialNumber, limit: initialLimit = DEFAULT_LIMIT, autoFetch = true } = options;

  const [records, setRecords] = useState<Maintenance[]>([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [limit, setLimitState] = useState(initialLimit);
  const [sort, setSort] = useState('updated_at_desc');
  const [searchQuery, setSearchQuery] = useState('');
  const debouncedSearchQuery = useDebounce(searchQuery, 300);
  const prevDebouncedSearch = useRef(debouncedSearchQuery);

  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [isSearchLoading, setIsSearchLoading] = useState(false);
  const [isPaginationLoading, setIsPaginationLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [preventativeCount, setPreventativeCount] = useState(0);
  const [correctiveCount, setCorrectiveCount] = useState(0);
  const [emergencyCount, setEmergencyCount] = useState(0);
  const [inspectionCount, setInspectionCount] = useState(0);
  const [otherCount, setOtherCount] = useState(0);

  const [hasLoadedInitial, setHasLoadedInitial] = useState(false);
  const [reloadCounter, setReloadCounter] = useState(0);
  const isFirstPaginationRun = useRef(true);
  const skipNextPaginationEffect = useRef(false);

  const limitRef = useRef(limit);
  const sortRef = useRef(sort);
  const searchRef = useRef(debouncedSearchQuery);
  const requestIdRef = useRef(0);

  useEffect(() => { limitRef.current = limit; }, [limit]);
  useEffect(() => { sortRef.current = sort; }, [sort]);
  useEffect(() => { searchRef.current = debouncedSearchQuery; }, [debouncedSearchQuery]);

  // Reset all state when serialNumber changes
  const prevSerialRef = useRef(serialNumber);
  useEffect(() => {
    if (prevSerialRef.current === serialNumber) return;
    prevSerialRef.current = serialNumber;
    setRecords([]);
    setTotal(0);
    setCurrentPage(1);
    setSort('updated_at_desc');
    setSearchQuery('');
    setError(null);
    setPreventativeCount(0);
    setCorrectiveCount(0);
    setEmergencyCount(0);
    setInspectionCount(0);
    setOtherCount(0);
    setHasLoadedInitial(false);
    setReloadCounter(0);
    isFirstPaginationRun.current = true;
    prevDebouncedSearch.current = '';
    searchRef.current = '';
    sortRef.current = 'updated_at_desc';
    requestIdRef.current++;
  }, [serialNumber]);

  const buildFilters = useCallback((): MaintenanceFilters => {
    const filters: MaintenanceFilters = {};
    const q = searchRef.current.trim();
    if (q) filters.q = q;
    if (sortRef.current) filters.sort = sortRef.current;
    return filters;
  }, []);

  const applyResponse = useCallback((data: {
    maintenance: Maintenance[];
    count: number;
    preventative_count: number;
    corrective_count: number;
    emergency_count: number;
    inspection_count: number;
    other_count: number;
  }) => {
    setRecords(data.maintenance || []);
    setTotal(data.count || 0);
    setPreventativeCount(data.preventative_count || 0);
    setCorrectiveCount(data.corrective_count || 0);
    setEmergencyCount(data.emergency_count || 0);
    setInspectionCount(data.inspection_count || 0);
    setOtherCount(data.other_count || 0);
  }, []);

  const handleFetchError = useCallback((err: unknown): void => {
    if (isAuthError(err)) return;
    const apiError = handleApiError(err) as ApiError;
    setError(apiError.message);
  }, []);

  // Initial load
  useEffect(() => {
    if (!serialNumber || !autoFetch || hasLoadedInitial) return;
    const myRequestId = ++requestIdRef.current;
    const load = async () => {
      setIsInitialLoading(true);
      setError(null);
      try {
        const response = await MaintenanceService.getMaintenanceList(
          serialNumber, 1, limitRef.current, buildFilters()
        );
        if (myRequestId !== requestIdRef.current) return;
        if (response.data) applyResponse(response.data);
      } catch (err) {
        if (myRequestId !== requestIdRef.current) return;
        handleFetchError(err);
      } finally {
        if (myRequestId === requestIdRef.current) {
          setHasLoadedInitial(true);
          setIsInitialLoading(false);
        }
      }
    };
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [serialNumber, autoFetch, hasLoadedInitial]);

  // Search changes
  useEffect(() => {
    if (debouncedSearchQuery === prevDebouncedSearch.current) return;
    prevDebouncedSearch.current = debouncedSearchQuery;
    if (!serialNumber || !hasLoadedInitial) return;

    if (currentPage !== 1) {
      skipNextPaginationEffect.current = true;
      setCurrentPage(1);
    }

    const myRequestId = ++requestIdRef.current;
    const load = async () => {
      setIsSearchLoading(true);
      setError(null);
      try {
        const response = await MaintenanceService.getMaintenanceList(
          serialNumber, 1, limitRef.current, buildFilters()
        );
        if (myRequestId !== requestIdRef.current) return;
        if (response.data) applyResponse(response.data);
      } catch (err) {
        if (myRequestId !== requestIdRef.current) return;
        handleFetchError(err);
      } finally {
        if (myRequestId === requestIdRef.current) {
          setIsSearchLoading(false);
        }
      }
    };
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedSearchQuery]);

  // Pagination / sort / limit changes
  useEffect(() => {
    if (isFirstPaginationRun.current) {
      isFirstPaginationRun.current = false;
      return;
    }
    if (skipNextPaginationEffect.current) {
      skipNextPaginationEffect.current = false;
      return;
    }
    if (!hasLoadedInitial || !serialNumber) return;

    const myRequestId = ++requestIdRef.current;
    const load = async () => {
      setIsPaginationLoading(true);
      setError(null);
      try {
        const response = await MaintenanceService.getMaintenanceList(
          serialNumber, currentPage, limitRef.current, buildFilters()
        );
        if (myRequestId !== requestIdRef.current) return;
        if (response.data) applyResponse(response.data);
      } catch (err) {
        if (myRequestId !== requestIdRef.current) return;
        handleFetchError(err);
      } finally {
        if (myRequestId === requestIdRef.current) {
          setIsPaginationLoading(false);
        }
      }
    };
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentPage, limit, sort]);

  // Reload trigger
  useEffect(() => {
    if (reloadCounter === 0 || !serialNumber) return;
    const myRequestId = ++requestIdRef.current;
    const load = async () => {
      setIsInitialLoading(true);
      setError(null);
      try {
        const response = await MaintenanceService.getMaintenanceList(
          serialNumber, currentPage, limitRef.current, buildFilters()
        );
        if (myRequestId !== requestIdRef.current) return;
        if (response.data) applyResponse(response.data);
      } catch (err) {
        if (myRequestId !== requestIdRef.current) return;
        handleFetchError(err);
      } finally {
        if (myRequestId === requestIdRef.current) {
          setIsInitialLoading(false);
        }
      }
    };
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [reloadCounter]);

  const totalPages = Math.ceil(total / limit);

  const goToPage = useCallback((page: number) => {
    setCurrentPage(Math.max(1, Math.min(page, Math.ceil(total / limitRef.current) || 1)));
  }, [total]);

  const handleSetLimit = useCallback((newLimit: number) => {
    setLimitState(newLimit);
    limitRef.current = newLimit;
    setCurrentPage(1);
  }, []);

  const clearSearch = useCallback(() => {
    if (currentPage !== 1) {
      skipNextPaginationEffect.current = true;
    }
    setSearchQuery('');
    setCurrentPage(1);
  }, [currentPage]);

  const refetch = useCallback(() => {
    if (currentPage !== 1) {
      skipNextPaginationEffect.current = true;
      setCurrentPage(1);
    }
    setReloadCounter((c) => c + 1);
  }, [currentPage]);

  return {
    records,
    total,
    currentPage,
    limit,
    totalPages,
    sort,
    searchQuery,
    debouncedSearchQuery,
    isInitialLoading,
    isSearchLoading,
    isPaginationLoading,
    error,
    preventativeCount,
    correctiveCount,
    emergencyCount,
    inspectionCount,
    otherCount,
    setSearchQuery,
    clearSearch,
    hasActiveSearch: debouncedSearchQuery.trim() !== '',
    setSort,
    goToPage,
    setLimit: handleSetLimit,
    refetch,
  };
}
