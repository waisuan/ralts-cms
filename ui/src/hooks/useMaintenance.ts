import { useState, useEffect, useCallback, useRef } from 'react';
import { Maintenance } from '../types/maintenance';
import { MaintenanceService, MaintenanceFilters } from '../services/maintenanceService';
import { handleApiError } from '../utils/api';
import { isAuthError } from '../utils/auth';

export interface UseMaintenanceOptions {
  serialNumber: string;
  page?: number; // 1-indexed
  limit?: number;
  filters?: MaintenanceFilters;
  autoFetch?: boolean;
}

export interface UseMaintenanceReturn {
  records: Maintenance[];
  total: number;
  totalPages: number;
  loading: boolean;
  error: string | null;
  preventativeCount: number;
  correctiveCount: number;
  emergencyCount: number;
  inspectionCount: number;
  otherCount: number;
  refetch: () => Promise<void>;
}

const DEFAULT_LIMIT = 50;

export function useMaintenance(options: UseMaintenanceOptions): UseMaintenanceReturn {
  const {
    serialNumber,
    page = 1,
    limit = DEFAULT_LIMIT,
    filters = {},
    autoFetch = true,
  } = options;

  const [records, setRecords] = useState<Maintenance[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [preventativeCount, setPreventativeCount] = useState(0);
  const [correctiveCount, setCorrectiveCount] = useState(0);
  const [emergencyCount, setEmergencyCount] = useState(0);
  const [inspectionCount, setInspectionCount] = useState(0);
  const [otherCount, setOtherCount] = useState(0);

  // Keep latest props in refs so fetchMaintenance stays stable across renders,
  // matching the useMachines pattern.
  const pageRef = useRef(page);
  const limitRef = useRef(limit);
  const filtersRef = useRef(filters);
  const serialRef = useRef(serialNumber);
  const requestIdRef = useRef(0);

  useEffect(() => {
    pageRef.current = page;
    limitRef.current = limit;
    filtersRef.current = filters;
    serialRef.current = serialNumber;
  }, [page, limit, filters, serialNumber]);

  const fetchMaintenance = useCallback(async () => {
    const currentSerial = serialRef.current;
    if (!currentSerial) return;

    const myRequestId = ++requestIdRef.current;
    setLoading(true);
    setError(null);

    try {
      const response = await MaintenanceService.getMaintenanceList(
        currentSerial,
        pageRef.current,
        limitRef.current,
        filtersRef.current,
      );
      if (myRequestId !== requestIdRef.current) return;

      const data = response.data;
      if (!data) {
        setRecords([]);
        setTotal(0);
        setPreventativeCount(0);
        setCorrectiveCount(0);
        setEmergencyCount(0);
        setInspectionCount(0);
        setOtherCount(0);
        return;
      }

      setRecords(data.maintenance || []);
      setTotal(data.count || 0);
      setPreventativeCount(data.preventative_count || 0);
      setCorrectiveCount(data.corrective_count || 0);
      setEmergencyCount(data.emergency_count || 0);
      setInspectionCount(data.inspection_count || 0);
      setOtherCount(data.other_count || 0);
    } catch (err) {
      if (myRequestId !== requestIdRef.current) return;
      if (isAuthError(err)) return;
      setError(handleApiError(err).message);
    } finally {
      if (myRequestId === requestIdRef.current) {
        setLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    if (!autoFetch) return;
    if (!serialNumber) return;
    fetchMaintenance();
  }, [fetchMaintenance, autoFetch, serialNumber, page, limit, filters]);

  const refetch = useCallback(async () => {
    await fetchMaintenance();
  }, [fetchMaintenance]);

  return {
    records,
    total,
    totalPages: Math.ceil(total / limit),
    loading,
    error,
    preventativeCount,
    correctiveCount,
    emergencyCount,
    inspectionCount,
    otherCount,
    refetch,
  };
}
