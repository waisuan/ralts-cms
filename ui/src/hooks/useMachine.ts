import { useState, useCallback } from 'react';
import { MachineService, CreateMachineRequest, UpdateMachineRequest } from '../services/machineService';
import { Machine } from '../types/machine';
import { ApiError, handleApiError } from '../utils/api';
import { isAuthError } from '../utils/auth';

export interface UseMachineReturn {
  machine: Machine | null;
  loading: boolean;
  error: ApiError | null;
  fetchMachine: (serialNumber: string) => Promise<void>;
  createMachine: (data: CreateMachineRequest) => Promise<Machine | null>;
  updateMachine: (serialNumber: string, data: UpdateMachineRequest) => Promise<Machine | null>;
  deleteMachine: (serialNumber: string) => Promise<boolean>;
  clearError: () => void;
  reset: () => void;
}

export function useMachine(): UseMachineReturn {
  const [machine, setMachine] = useState<Machine | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<ApiError | null>(null);

  const fetchMachine = useCallback(async (serialNumber: string) => {
    setLoading(true);
    setError(null);

    try {
      const response = await MachineService.getMachine(serialNumber);
      setMachine(response.data as Machine);
    } catch (err) {
      const apiError = handleApiError(err);
      // Don't set error if we're redirecting due to auth error
      if (!isAuthError(err)) {
        setError(apiError);
        console.error('Failed to fetch machine:', apiError);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  const createMachine = useCallback(async (data: CreateMachineRequest): Promise<Machine | null> => {
    setLoading(true);
    setError(null);

    try {
      const response = await MachineService.createMachine(data);
      const newMachine = response.data as Machine;
      setMachine(newMachine);
      return newMachine;
    } catch (err) {
      const apiError = handleApiError(err);
      // Don't set error if we're redirecting due to auth error
      if (!isAuthError(err)) {
        setError(apiError);
        console.error('Failed to create machine:', apiError);
      }
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  const updateMachine = useCallback(async (
    serialNumber: string,
    data: UpdateMachineRequest
  ): Promise<Machine | null> => {
    setLoading(true);
    setError(null);

    try {
      const response = await MachineService.updateMachine(serialNumber, data);
      const updatedMachine = response.data as Machine;
      setMachine(updatedMachine);
      return updatedMachine;
    } catch (err) {
      const apiError = handleApiError(err);
      // Don't set error if we're redirecting due to auth error
      if (!isAuthError(err)) {
        setError(apiError);
        console.error('Failed to update machine:', apiError);
      }
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  const deleteMachine = useCallback(async (serialNumber: string): Promise<boolean> => {
    setLoading(true);
    setError(null);

    try {
      await MachineService.deleteMachine(serialNumber);
      setMachine(null);
      return true;
    } catch (err) {
      const apiError = handleApiError(err);
      // Don't set error if we're redirecting due to auth error
      if (!isAuthError(err)) {
        setError(apiError);
        console.error('Failed to delete machine:', apiError);
      }
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const reset = useCallback(() => {
    setMachine(null);
    setLoading(false);
    setError(null);
  }, []);

  return {
    machine,
    loading,
    error,
    fetchMachine,
    createMachine,
    updateMachine,
    deleteMachine,
    clearError,
    reset,
  };
} 