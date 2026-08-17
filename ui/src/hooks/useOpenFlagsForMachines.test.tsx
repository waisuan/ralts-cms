import { renderHook, waitFor } from '@testing-library/react';
import { useOpenFlagsForMachines } from './useOpenFlagsForMachines';
import { FlagService } from '../services/flagService';
import { useOptionalAuth } from '../contexts/AuthContext';

jest.mock('../services/flagService', () => ({
  FlagService: {
    listOpenByMachine: jest.fn(),
  },
}));

jest.mock('../contexts/AuthContext', () => ({
  useOptionalAuth: jest.fn(),
}));

const mockAuthRole = (role: string | null) => {
  (useOptionalAuth as jest.Mock).mockReturnValue(
    role === null ? null : { user: { id: 1, username: 'u', email: 'e', role, approved: true } },
  );
};

const openFlag = {
  id: 'f-1',
  machine_serial_number: 'SN-1',
  reason: 'missing_values' as const,
  status: 'open' as const,
  created_at: '2024-01-01T00:00:00Z',
};

describe('useOpenFlagsForMachines', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (FlagService.listOpenByMachine as jest.Mock).mockResolvedValue({
      data: { flags: { 'SN-1': [openFlag] } },
    });
  });

  it('fetches open flags for admins', async () => {
    mockAuthRole('ADMIN');

    const { result } = renderHook(() => useOpenFlagsForMachines(['SN-1']));

    await waitFor(() => {
      expect(result.current.flagsBySerial['SN-1']).toHaveLength(1);
    });
    expect(FlagService.listOpenByMachine).toHaveBeenCalledWith(['SN-1']);
  });

  it('fetches open flags for non-admins too', async () => {
    mockAuthRole('USER');

    const { result } = renderHook(() => useOpenFlagsForMachines(['SN-1']));

    await waitFor(() => {
      expect(result.current.flagsBySerial['SN-1']).toHaveLength(1);
    });
    expect(FlagService.listOpenByMachine).toHaveBeenCalledWith(['SN-1']);
  });

  it('returns nothing when there is no signed-in user', async () => {
    mockAuthRole(null);

    const { result } = renderHook(() => useOpenFlagsForMachines(['SN-1']));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });
    expect(result.current.flagsBySerial).toEqual({});
    expect(FlagService.listOpenByMachine).not.toHaveBeenCalled();
  });
});
